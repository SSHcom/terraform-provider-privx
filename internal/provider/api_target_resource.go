package provider

import (
	"context"
	"fmt"
	"strings"
	"terraform-provider-privx/internal/utils"

	"github.com/SSHcom/privx-sdk-go/v2/api/apiproxy"
	"github.com/SSHcom/privx-sdk-go/v2/restapi"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &APITargetResource{}
var _ resource.ResourceWithImportState = &APITargetResource{}

func NewAPITargetResource() resource.Resource {
	return &APITargetResource{}
}

// APITargetResource defines the resource implementation.
type APITargetResource struct {
	client *apiproxy.ApiProxy
}

type ApiTargetEndpointModel struct {
	Host                 types.String `tfsdk:"host"`
	Protocols            types.Set    `tfsdk:"protocols"` // set(string)
	Methods              types.Set    `tfsdk:"methods"`   // set(string)
	Paths                types.Set    `tfsdk:"paths"`     // set(string)
	AllowUnauthenticated types.Bool   `tfsdk:"allow_unauthenticated"`
	NATTargetHost        types.String `tfsdk:"nat_target_host"`
}

type TargetCredentialModel struct {
	Type              types.String `tfsdk:"type"`
	BasicAuthUsername types.String `tfsdk:"basic_auth_username"`
	BasicAuthPassword types.String `tfsdk:"basic_auth_password"`
	BearerToken       types.String `tfsdk:"bearer_token"`
	Certificate       types.String `tfsdk:"certificate"`
	PrivateKey        types.String `tfsdk:"private_key"`
}

// APITargetModel describes the resource data model.
type APITargetModel struct {
	ID                    types.String             `tfsdk:"id"`
	Name                  types.String             `tfsdk:"name"`
	Comment               types.String             `tfsdk:"comment"`
	Tags                  types.Set                `tfsdk:"tags"` // set(string)
	AccessGroupID         types.String             `tfsdk:"access_group_id"`
	Roles                 []RolesRefModel          `tfsdk:"roles"`
	AuthorizedEndpoints   []ApiTargetEndpointModel `tfsdk:"authorized_endpoints"`
	UnauthorizedEndpoints []ApiTargetEndpointModel `tfsdk:"unauthorized_endpoints"`
	TLSTrustAnchors       types.String             `tfsdk:"tls_trust_anchors"`
	TLSInsecureSkipVerify types.Bool               `tfsdk:"tls_insecure_skip_verify"`
	TargetCredential      *TargetCredentialModel   `tfsdk:"target_credential"`
	Disabled              types.String             `tfsdk:"disabled"`
	AuditEnabled          types.Bool               `tfsdk:"audit_enabled"`

	Created   types.String `tfsdk:"created"`
	Author    types.String `tfsdk:"author"`
	Updated   types.String `tfsdk:"updated"`
	UpdatedBy types.String `tfsdk:"updated_by"`
}

func (r *APITargetResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_target"
}

func (r *APITargetResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "API target resource (API-Proxy). Defines backend API endpoint(s), allowed request patterns, roles, and target credentials.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "ID of the API target",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"name": schema.StringAttribute{
				MarkdownDescription: "Unique name of the API target",
				Required:            true,
			},

			"comment": schema.StringAttribute{
				MarkdownDescription: "Optional comment",
				Optional:            true,
			},

			"tags": schema.SetAttribute{
				MarkdownDescription: "Optional tags",
				Optional:            true,
				ElementType:         types.StringType,
			},

			"access_group_id": schema.StringAttribute{
				MarkdownDescription: "Optional access group ID",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"roles": schema.SetNestedAttribute{
				MarkdownDescription: "Roles which grant access to the API target",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "Role ID",
							Required:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "Role name (computed from server). If server omits, provider preserves prior/planned name when present.",
							Optional:            true,
							Computed:            true,
						},
					},
				},
			},

			"authorized_endpoints": schema.ListNestedAttribute{
				MarkdownDescription: "Authorized endpoint patterns. A request must match at least one authorized endpoint.",
				Required:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"host": schema.StringAttribute{
							MarkdownDescription: "Host (optionally with port) matched against request URL host",
							Required:            true,
						},
						"protocols": schema.SetAttribute{
							MarkdownDescription: `Protocols matched against scheme: "http", "https", or "*"`,
							Optional:            true,
							ElementType:         types.StringType,
						},
						"methods": schema.SetAttribute{
							MarkdownDescription: `HTTP methods matched: "GET", "POST", ..., or "*"`,
							Optional:            true,
							ElementType:         types.StringType,
						},
						"paths": schema.SetAttribute{
							MarkdownDescription: `Paths matched; supports "*" and "**" wildcards`,
							Optional:            true,
							ElementType:         types.StringType,
						},
						"allow_unauthenticated": schema.BoolAttribute{
							MarkdownDescription: "Allow unauthenticated requests for this endpoint pattern",
							Optional:            true,
							Computed:            true,
							Default:             booldefault.StaticBool(false),
						},
						"nat_target_host": schema.StringAttribute{
							MarkdownDescription: "Optional NAT target host when forwarding over an extender",
							Optional:            true,
						},
					},
				},
			},

			"unauthorized_endpoints": schema.ListNestedAttribute{
				MarkdownDescription: "Unauthorized endpoint patterns. An authorized request must not match any unauthorized endpoint.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"host": schema.StringAttribute{
							MarkdownDescription: "Host (optionally with port) matched against request URL host",
							Required:            true,
						},
						"protocols": schema.SetAttribute{
							MarkdownDescription: `Protocols matched against scheme: "http", "https", or "*"`,
							Optional:            true,
							ElementType:         types.StringType,
						},
						"methods": schema.SetAttribute{
							MarkdownDescription: `HTTP methods matched: "GET", "POST", ..., or "*"`,
							Optional:            true,
							ElementType:         types.StringType,
						},
						"paths": schema.SetAttribute{
							MarkdownDescription: `Paths matched; supports "*" and "**" wildcards`,
							Optional:            true,
							ElementType:         types.StringType,
						},
						"allow_unauthenticated": schema.BoolAttribute{
							MarkdownDescription: "Allow unauthenticated requests for this endpoint pattern",
							Optional:            true,
							Computed:            true,
							Default:             booldefault.StaticBool(false),
						},
						"nat_target_host": schema.StringAttribute{
							MarkdownDescription: "Optional NAT target host when forwarding over an extender",
							Optional:            true,
						},
					},
				},
			},

			"tls_trust_anchors": schema.StringAttribute{
				MarkdownDescription: "Optional PEM trust anchors used for validating target TLS certificates",
				Optional:            true,
				Sensitive:           true,
				PlanModifiers: []planmodifier.String{
					// Avoid plan churn on unknowns / server not echoing the exact PEM formatting.
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"tls_insecure_skip_verify": schema.BoolAttribute{
				MarkdownDescription: "Skip TLS server certificate verification when connecting to API target",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},

			"target_credential": schema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: `Credential type (lowercase in Terraform): "token", "basic_auth", "certificate".`,
						Validators: []validator.String{
							stringvalidator.OneOf("token", "basic_auth", "certificate"),
						},
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"basic_auth_username": schema.StringAttribute{
						Optional: true,
						Computed: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"basic_auth_password": schema.StringAttribute{
						Optional:  true,
						Computed:  true,
						Sensitive: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"bearer_token": schema.StringAttribute{
						Optional:  true,
						Computed:  true,
						Sensitive: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"certificate": schema.StringAttribute{
						Optional:  true,
						Computed:  true,
						Sensitive: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"private_key": schema.StringAttribute{
						Optional:  true,
						Computed:  true,
						Sensitive: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
				},
			},

			"disabled": schema.StringAttribute{
				MarkdownDescription: `Disabled state. Can be one of NOT_DISABLED, BY_ADMIN, BY_LICENSE. If not set, server decides.`,
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"audit_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether to session record requests to this target API",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},

			"created": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"author": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_by": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *APITargetResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	connector, ok := req.ProviderData.(*restapi.Connector)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *restapi.Connector, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	r.client = apiproxy.New(*connector)
}

func (r *APITargetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan *APITargetModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload, diags := expandApiTarget(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := r.client.CreateApiTarget(payload)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Create api_target Resource", err.Error())
		return
	}

	remote, err := r.client.GetApiTarget(id.ID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read api_target after create, got error: %s", err))
		return
	}

	state, diags := flattenApiTarget(ctx, remote, plan)
	preservePlannedCredentialSecrets(plan, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// GUARANTEED: preserve planned tls_trust_anchors into state on Create
	if !plan.TLSTrustAnchors.IsNull() && !plan.TLSTrustAnchors.IsUnknown() {
		state.TLSTrustAnchors = plan.TLSTrustAnchors
	}

	// Set ID explicitly
	state.ID = types.StringValue(remote.ID)

	ctx = tflog.SetField(ctx, "api_target_id", remote.ID)
	tflog.Debug(ctx, "Created api_target")

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *APITargetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state *APITargetModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	remote, err := r.client.GetApiTarget(state.ID.ValueString())
	if err != nil {
		if utils.IsPrivxNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read api_target, got error: %s", err))
		return
	}

	next, diags := flattenApiTarget(ctx, remote, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Keep ID
	next.ID = state.ID

	resp.Diagnostics.Append(resp.State.Set(ctx, &next)...)
}

func (r *APITargetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan *APITargetModel
	var prior *APITargetModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &prior)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload, diags := expandApiTarget(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Ensure ID is set
	payload.ID = prior.ID.ValueString()

	if err := r.client.UpdateApiTarget(prior.ID.ValueString(), payload); err != nil {
		resp.Diagnostics.AddError("Unable to Update api_target Resource", err.Error())
		return
	}

	remote, err := r.client.GetApiTarget(prior.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read api_target after update, got error: %s", err))
		return
	}

	next, diags := flattenApiTarget(ctx, remote, prior)
	preservePlannedCredentialSecrets(plan, &next)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// ✅ Preserve planned tls_trust_anchors into state (sensitive; API may not echo back)
	if !plan.TLSTrustAnchors.IsNull() && !plan.TLSTrustAnchors.IsUnknown() {
		next.TLSTrustAnchors = plan.TLSTrustAnchors
	}

	next.ID = prior.ID
	resp.Diagnostics.Append(resp.State.Set(ctx, &next)...)
}

func (r *APITargetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state *APITargetModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteApiTarget(state.ID.ValueString()); err != nil {
		if utils.IsPrivxNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete api_target, got error: %s", err))
		return
	}
}

func (r *APITargetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// ----------------------- helpers -----------------------

func expandApiTarget(ctx context.Context, m *APITargetModel) (*apiproxy.ApiTarget, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Tags
	var tags []string
	if !m.Tags.IsNull() && !m.Tags.IsUnknown() {
		diags.Append(m.Tags.ElementsAs(ctx, &tags, false)...)
	}

	// Roles
	var roles []apiproxy.RoleHandle
	for _, rr := range m.Roles {
		if rr.ID.IsNull() || rr.ID.IsUnknown() || rr.ID.ValueString() == "" {
			continue
		}
		rh := apiproxy.RoleHandle{ID: rr.ID.ValueString()}
		// Name is optional for payload; send if user provided a non-empty one.
		if !rr.Name.IsNull() && !rr.Name.IsUnknown() && rr.Name.ValueString() != "" {
			rh.Name = rr.Name.ValueString()
		}
		roles = append(roles, rh)
	}

	authorized, d1 := expandEndpoints(ctx, m.AuthorizedEndpoints)
	diags.Append(d1...)

	unauthorized, d2 := expandEndpoints(ctx, m.UnauthorizedEndpoints)
	diags.Append(d2...)

	tc := expandTargetCredential(m.TargetCredential)

	payload := &apiproxy.ApiTarget{
		Name:                  strings.TrimSpace(m.Name.ValueString()),
		Comment:               strings.TrimSpace(m.Comment.ValueString()),
		Tags:                  tags,
		AccessGroupID:         strings.TrimSpace(m.AccessGroupID.ValueString()),
		Roles:                 roles,
		AuthorizedEndpoints:   authorized,
		UnauthorizedEndpoints: unauthorized,
		TLSTrustAnchors:       normalizePEM(m.TLSTrustAnchors),
		TLSInsecureSkipVerify: !m.TLSInsecureSkipVerify.IsNull() &&
			!m.TLSInsecureSkipVerify.IsUnknown() &&
			m.TLSInsecureSkipVerify.ValueBool(),
		TargetCredential: tc,
	}

	// Disabled: only send if user explicitly set it
	if !m.Disabled.IsNull() && !m.Disabled.IsUnknown() {
		if v := strings.TrimSpace(m.Disabled.ValueString()); v != "" {
			payload.Disabled = v
		}
	}

	// AuditEnabled: only send if user explicitly set it
	if !m.AuditEnabled.IsNull() && !m.AuditEnabled.IsUnknown() {
		payload.AuditEnabled = m.AuditEnabled.ValueBool()
	}

	return payload, diags
}

func expandEndpoints(ctx context.Context, in []ApiTargetEndpointModel) ([]apiproxy.ApiTargetEndpoint, diag.Diagnostics) {
	var diags diag.Diagnostics
	var out []apiproxy.ApiTargetEndpoint

	for _, ep := range in {
		var protocols, methods, paths []string
		if !ep.Protocols.IsNull() && !ep.Protocols.IsUnknown() {
			diags.Append(ep.Protocols.ElementsAs(ctx, &protocols, false)...)
		}
		if !ep.Methods.IsNull() && !ep.Methods.IsUnknown() {
			diags.Append(ep.Methods.ElementsAs(ctx, &methods, false)...)
		}
		if !ep.Paths.IsNull() && !ep.Paths.IsUnknown() {
			diags.Append(ep.Paths.ElementsAs(ctx, &paths, false)...)
		}

		nat := ""
		if !ep.NATTargetHost.IsNull() && !ep.NATTargetHost.IsUnknown() {
			nat = strings.TrimSpace(ep.NATTargetHost.ValueString())
		}

		allowUnauth := false
		if !ep.AllowUnauthenticated.IsNull() && !ep.AllowUnauthenticated.IsUnknown() {
			allowUnauth = ep.AllowUnauthenticated.ValueBool()
		}

		out = append(out, apiproxy.ApiTargetEndpoint{
			Host:                 strings.TrimSpace(ep.Host.ValueString()),
			Protocols:            protocols,
			Methods:              methods,
			Paths:                paths,
			AllowUnauthenticated: allowUnauth,
			NATTargetHost:        nat,
		})
	}

	return out, diags
}

func expandTargetCredential(tc *TargetCredentialModel) apiproxy.TargetCredential {
	if tc == nil {
		return apiproxy.TargetCredential{}
	}

	out := apiproxy.TargetCredential{
		Type: credentialTypeToAPI(tc.Type.ValueString()),
	}

	if !tc.BasicAuthUsername.IsNull() && !tc.BasicAuthUsername.IsUnknown() {
		out.BasicAuthUsername = tc.BasicAuthUsername.ValueString()
	}
	if !tc.BasicAuthPassword.IsNull() && !tc.BasicAuthPassword.IsUnknown() {
		out.BasicAuthPassword = tc.BasicAuthPassword.ValueString()
	}
	if !tc.BearerToken.IsNull() && !tc.BearerToken.IsUnknown() {
		out.BearerToken = tc.BearerToken.ValueString()
	}
	if !tc.Certificate.IsNull() && !tc.Certificate.IsUnknown() {
		out.Certificate = tc.Certificate.ValueString()
	}
	if !tc.PrivateKey.IsNull() && !tc.PrivateKey.IsUnknown() {
		out.PrivateKey = tc.PrivateKey.ValueString()
	}

	return out
}

func credentialTypeToAPI(tf string) string {
	switch strings.ToLower(strings.TrimSpace(tf)) {
	case "token":
		return "Token"
	case "basic_auth":
		return "BasicAuth"
	case "certificate":
		return "Certificate"
	default:
		// fall back to input; validator should prevent this
		return strings.TrimSpace(tf)
	}
}

func credentialTypeFromAPI(api string) string {
	switch strings.TrimSpace(api) {
	case "Token":
		return "token"
	case "BasicAuth":
		return "basic_auth"
	case "Certificate":
		return "certificate"
	default:
		return ""
	}
}

func flattenApiTarget(ctx context.Context, remote *apiproxy.ApiTarget, prior *APITargetModel) (APITargetModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	var out APITargetModel

	// Non-secret fields
	out.ID = types.StringValue(remote.ID)
	out.Name = types.StringValue(remote.Name)
	out.Comment = types.StringValue(remote.Comment)
	out.AccessGroupID = types.StringNull()
	if strings.TrimSpace(remote.AccessGroupID) != "" {
		out.AccessGroupID = types.StringValue(remote.AccessGroupID)
	} else if prior != nil && !prior.AccessGroupID.IsNull() && !prior.AccessGroupID.IsUnknown() {
		out.AccessGroupID = prior.AccessGroupID
	}

	out.TLSInsecureSkipVerify = types.BoolValue(remote.TLSInsecureSkipVerify)

	out.Disabled = types.StringNull()
	if strings.TrimSpace(remote.Disabled) != "" {
		out.Disabled = types.StringValue(remote.Disabled)
	} else if prior != nil && !prior.Disabled.IsNull() && !prior.Disabled.IsUnknown() {
		out.Disabled = prior.Disabled
	}

	out.AuditEnabled = types.BoolValue(remote.AuditEnabled)
	// created
	out.Created = types.StringNull()
	if strings.TrimSpace(remote.Created) != "" {
		out.Created = types.StringValue(remote.Created)
	}

	// author
	out.Author = types.StringNull()
	if strings.TrimSpace(remote.Author) != "" {
		out.Author = types.StringValue(remote.Author)
	}

	// updated / updated_by:
	// Avoid null -> value change during Apply consistency check.
	// If prior was null, keep it null during apply; it will populate on next Read/refresh.
	out.Updated = types.StringNull()
	if prior != nil && !prior.Updated.IsUnknown() && prior.Updated.IsNull() {
		// keep null
	} else if strings.TrimSpace(remote.Updated) != "" {
		out.Updated = types.StringValue(remote.Updated)
	} else if prior != nil && !prior.Updated.IsNull() && !prior.Updated.IsUnknown() {
		out.Updated = prior.Updated
	}

	out.UpdatedBy = types.StringNull()
	if prior != nil && !prior.UpdatedBy.IsUnknown() && prior.UpdatedBy.IsNull() {
		// keep null
	} else if strings.TrimSpace(remote.UpdatedBy) != "" {
		out.UpdatedBy = types.StringValue(remote.UpdatedBy)
	} else if prior != nil && !prior.UpdatedBy.IsNull() && !prior.UpdatedBy.IsUnknown() {
		out.UpdatedBy = prior.UpdatedBy
	}

	// Tags set
	if len(remote.Tags) == 0 && (prior == nil || prior.Tags.IsNull()) {
		out.Tags = types.SetNull(types.StringType)
	} else {
		set, d := types.SetValueFrom(ctx, types.StringType, remote.Tags)
		diags.Append(d...)
		out.Tags = set
	}

	// Roles: prefer API name; otherwise preserve prior/planned name if present
	var roles []RolesRefModel
	for _, rrole := range remote.Roles {
		roleName := types.StringNull()
		if rrole.Name != "" {
			roleName = types.StringValue(rrole.Name)
		} else if prior != nil {
			for _, prev := range prior.Roles {
				if prev.ID.ValueString() == rrole.ID &&
					!prev.Name.IsNull() &&
					!prev.Name.IsUnknown() &&
					prev.Name.ValueString() != "" {
					roleName = prev.Name
					break
				}
			}
		}
		roles = append(roles, RolesRefModel{
			ID:   types.StringValue(rrole.ID),
			Name: roleName,
		})
	}
	out.Roles = roles

	// Endpoints
	out.AuthorizedEndpoints, diags = flattenEndpoints(ctx, remote.AuthorizedEndpoints, diags)
	// Unauthorized endpoints: keep null if API returns empty AND user didn't set it
	if len(remote.UnauthorizedEndpoints) == 0 && (prior == nil || len(prior.UnauthorizedEndpoints) == 0) {
		out.UnauthorizedEndpoints = nil // null in TF
	} else {
		out.UnauthorizedEndpoints, diags = flattenEndpoints(ctx, remote.UnauthorizedEndpoints, diags)
	}

	// TLS trust anchors (Sensitive):
	// - If API returns a real value, normalize it and store it.
	// - If API returns empty or masked, preserve prior to avoid perpetual diffs.
	remoteTA := strings.TrimSpace(remote.TLSTrustAnchors)
	if isMaskedSecret(remoteTA) || remoteTA == "" {
		if prior != nil && !prior.TLSTrustAnchors.IsNull() && !prior.TLSTrustAnchors.IsUnknown() {
			out.TLSTrustAnchors = prior.TLSTrustAnchors
		} else {
			out.TLSTrustAnchors = types.StringNull()
		}
	} else {
		// normalize to match expand/plan normalization
		if !strings.HasSuffix(remoteTA, "\n") {
			remoteTA += "\n"
		}
		out.TLSTrustAnchors = types.StringValue(remoteTA)
	}

	// Target credential: preserve secrets if remote omits them
	out.TargetCredential = flattenTargetCredential(remote.TargetCredential, prior)

	return out, diags
}

func flattenEndpoints(ctx context.Context, in []apiproxy.ApiTargetEndpoint, diags diag.Diagnostics) ([]ApiTargetEndpointModel, diag.Diagnostics) {
	out := make([]ApiTargetEndpointModel, 0, len(in))

	for _, ep := range in {
		protoSet, d1 := types.SetValueFrom(ctx, types.StringType, ep.Protocols)
		diags.Append(d1...)
		methodSet, d2 := types.SetValueFrom(ctx, types.StringType, ep.Methods)
		diags.Append(d2...)
		pathSet, d3 := types.SetValueFrom(ctx, types.StringType, ep.Paths)
		diags.Append(d3...)

		nat := types.StringNull()
		if ep.NATTargetHost != "" {
			nat = types.StringValue(ep.NATTargetHost)
		}

		out = append(out, ApiTargetEndpointModel{
			Host:                 types.StringValue(ep.Host),
			Protocols:            protoSet,
			Methods:              methodSet,
			Paths:                pathSet,
			AllowUnauthenticated: types.BoolValue(ep.AllowUnauthenticated),
			NATTargetHost:        nat,
		})
	}

	return out, diags
}

func flattenTargetCredential(remote apiproxy.TargetCredential, prior *APITargetModel) *TargetCredentialModel {
	preserveStr := func(remoteVal string, priorVal types.String) types.String {
		remoteVal = strings.TrimSpace(remoteVal)

		// Treat masked secrets as "not returned"
		if isMaskedSecret(remoteVal) {
			remoteVal = ""
		}

		// If server gave a real value, use it
		if remoteVal != "" {
			return types.StringValue(remoteVal)
		}

		// ✅ Never return Unknown in state after apply
		// If prior is unknown, collapse to null (unknown is plan-time only).
		if priorVal.IsUnknown() {
			return types.StringNull()
		}

		// Preserve prior value if it exists (known string or known empty)
		if !priorVal.IsNull() {
			return priorVal
		}

		return types.StringNull()
	}

	// Prior values (if any)
	priorType := types.StringNull()
	priorUser := types.StringNull()
	priorPass := types.StringNull()
	priorToken := types.StringNull()
	priorCert := types.StringNull()
	priorKey := types.StringNull()

	// IMPORTANT: actually load prior values so refresh can preserve secrets
	if prior != nil && prior.TargetCredential != nil {
		priorType = prior.TargetCredential.Type
		priorUser = prior.TargetCredential.BasicAuthUsername
		priorPass = prior.TargetCredential.BasicAuthPassword
		priorToken = prior.TargetCredential.BearerToken
		priorCert = prior.TargetCredential.Certificate
		priorKey = prior.TargetCredential.PrivateKey
	}

	// If API doesn’t echo anything at all, just keep what we already had.
	// (Covers backends that return an empty credential object.)
	if prior != nil && prior.TargetCredential != nil {
		if strings.TrimSpace(remote.Type) == "" &&
			strings.TrimSpace(remote.BasicAuthUsername) == "" &&
			strings.TrimSpace(remote.BasicAuthPassword) == "" &&
			strings.TrimSpace(remote.BearerToken) == "" &&
			strings.TrimSpace(remote.Certificate) == "" &&
			strings.TrimSpace(remote.PrivateKey) == "" {
			return prior.TargetCredential
		}
	}

	tfType := strings.TrimSpace(credentialTypeFromAPI(remote.Type))

	outType := types.StringNull()
	if tfType != "" {
		outType = types.StringValue(tfType)
	} else if !priorType.IsNull() && !priorType.IsUnknown() {
		outType = priorType
	}

	return &TargetCredentialModel{
		Type:              outType,
		BasicAuthUsername: preserveStr(remote.BasicAuthUsername, priorUser),

		// secrets
		BasicAuthPassword: preserveStr(remote.BasicAuthPassword, priorPass),
		BearerToken:       preserveStr(remote.BearerToken, priorToken),
		Certificate:       preserveStr(remote.Certificate, priorCert),
		PrivateKey:        preserveStr(remote.PrivateKey, priorKey),
	}
}

func preservePlannedCredentialSecrets(plan *APITargetModel, state *APITargetModel) {
	if plan == nil || state == nil || plan.TargetCredential == nil || state.TargetCredential == nil {
		return
	}

	if !plan.TargetCredential.BasicAuthPassword.IsNull() && !plan.TargetCredential.BasicAuthPassword.IsUnknown() {
		state.TargetCredential.BasicAuthPassword = plan.TargetCredential.BasicAuthPassword
	}
	if !plan.TargetCredential.BearerToken.IsNull() && !plan.TargetCredential.BearerToken.IsUnknown() {
		state.TargetCredential.BearerToken = plan.TargetCredential.BearerToken
	}
	if !plan.TargetCredential.Certificate.IsNull() && !plan.TargetCredential.Certificate.IsUnknown() {
		state.TargetCredential.Certificate = plan.TargetCredential.Certificate
	}
	if !plan.TargetCredential.PrivateKey.IsNull() && !plan.TargetCredential.PrivateKey.IsUnknown() {
		state.TargetCredential.PrivateKey = plan.TargetCredential.PrivateKey
	}
}

func normalizePEM(v types.String) string {
	if v.IsNull() || v.IsUnknown() {
		return ""
	}
	s := strings.TrimSpace(v.ValueString())
	if s == "" {
		return ""
	}
	// normalize to end with newline to reduce diffs
	if !strings.HasSuffix(s, "\n") {
		s += "\n"
	}
	return s
}

// Optional: ensure nested types compile even if unused in this file.
var _ = attr.Type(types.StringType)

func isMaskedSecret(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false // empty is empty, not "masked"
	}
	if s == "****************" || s == "********" || s == "*****" {
		return true
	}
	for _, r := range s {
		if r != '*' {
			return false
		}
	}
	return true
}
