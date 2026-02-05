package provider

import (
	"context"
	"fmt"
	"strings"
	"terraform-provider-privx/internal/utils"

	"github.com/SSHcom/privx-sdk-go/v2/api/apiproxy"
	"github.com/SSHcom/privx-sdk-go/v2/restapi"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &apiProxyCredentialResource{}
	_ resource.ResourceWithConfigure   = &apiProxyCredentialResource{}
	_ resource.ResourceWithImportState = &apiProxyCredentialResource{}
)

func NewApiProxyCredentialResource() resource.Resource {
	return &apiProxyCredentialResource{}
}

type apiProxyCredentialResource struct {
	client *apiproxy.ApiProxy
}

// Model.
type apiProxyCredentialModel struct {
	ID            types.String `tfsdk:"id"`
	UserID        types.String `tfsdk:"user_id"`
	TargetID      types.String `tfsdk:"target_id"`
	Name          types.String `tfsdk:"name"`
	Comment       types.String `tfsdk:"comment"`
	NotBefore     types.String `tfsdk:"not_before"`
	NotAfter      types.String `tfsdk:"not_after"`
	SourceAddress types.List   `tfsdk:"source_address"` // list(string)
	Enabled       types.Bool   `tfsdk:"enabled"`
	Type          types.String `tfsdk:"type"`

	// Secret returned by PrivX for client usage.
	// We read it on create (and optionally on read if you want), but generally keep in state.
	Secret types.String `tfsdk:"secret"`

	// Read-only metadata
	LastUsed  types.String `tfsdk:"last_used"`
	Created   types.String `tfsdk:"created"`
	Author    types.String `tfsdk:"author"`
	Updated   types.String `tfsdk:"updated"`
	UpdatedBy types.String `tfsdk:"updated_by"`
}

func (r *apiProxyCredentialResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_proxy_credential"
}

func (r *apiProxyCredentialResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages PrivX API Proxy credentials (client credentials) for current user or a specified user_id.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
				Description:   "Credential ID.",
			},
			"user_id": schema.StringAttribute{
				Optional:    true,
				Computed:    false,
				Description: "If set, manage credential for this user (admin use-case). If omitted, uses current user endpoints.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"target_id": schema.StringAttribute{
				Required:    true,
				Description: "API target ID this credential is associated with.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Human readable credential name.",
			},
			"comment": schema.StringAttribute{
				Optional:    true,
				Description: "Optional comment.",
			},
			"not_before": schema.StringAttribute{
				Required:    true,
				Description: "Validity start timestamp (string; should be RFC3339 as expected by PrivX).",
			},
			"not_after": schema.StringAttribute{
				Required:    true,
				Description: "Validity end timestamp (string; should be RFC3339 as expected by PrivX).",
			},
			"source_address": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Optional list of allowed client IPs/CIDRs.",
			},
			"enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Whether the credential is enabled.",
			},
			"type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("token"),
				Description: `Credential type. One of: "token", "basicauth", "certificate".`,
			},
			"secret": schema.StringAttribute{
				Computed:    true,
				Sensitive:   true,
				Description: "Client credential secret to be used by the API client. Typically only returned on create; stored in state.",
				PlanModifiers: []planmodifier.String{
					// Don’t force a diff if API does not return it later.
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			// Metadata (computed)
			"last_used": schema.StringAttribute{Computed: true},
			"created":   schema.StringAttribute{Computed: true},
			"author":    schema.StringAttribute{Computed: true},
			"updated":   schema.StringAttribute{Computed: true},
			"updated_by": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (r *apiProxyCredentialResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	connPtr, ok := req.ProviderData.(*restapi.Connector)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected provider data type",
			fmt.Sprintf("Expected *restapi.Connector, got: %T", req.ProviderData),
		)
		return
	}
	if connPtr == nil || *connPtr == nil {
		resp.Diagnostics.AddError("Provider not configured", "Connector was nil")
		return
	}

	r.client = apiproxy.New(*connPtr)
}

func (r *apiProxyCredentialResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan apiProxyCredentialModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cred, diags := expandClientCredential(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var (
		createdID string
		secret    []byte
		err       error
	)

	if !plan.UserID.IsNull() && plan.UserID.ValueString() != "" {
		idResp, e := r.client.CreateUserClientCredential(plan.UserID.ValueString(), cred)
		if e != nil {
			resp.Diagnostics.AddError("Create API proxy credential failed", e.Error())
			return
		}
		createdID = idResp.ID

		secret, err = r.client.GetUserClientCredentialSecret(plan.UserID.ValueString(), createdID)
	} else {
		idResp, e := r.client.CreateCurrentUserClientCredential(cred)
		if e != nil {
			resp.Diagnostics.AddError("Create API proxy credential failed", e.Error())
			return
		}
		createdID = idResp.ID

		secret, err = r.client.GetCurrentUserClientCredentialSecret(createdID)
	}

	if err != nil {
		resp.Diagnostics.AddError("Reading API proxy credential secret failed", err.Error())
		return
	}

	plan.ID = types.StringValue(createdID)
	plan.Secret = types.StringValue(string(secret))

	// Refresh server view into state (includes metadata)
	state, diags := r.readIntoState(ctx, plan.UserID, createdID, plan.Secret)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *apiProxyCredentialResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state apiProxyCredentialModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Keep whatever secret we already have in state (API might not re-expose it)
	secret := state.Secret

	// Normalize user_id:
	userID := state.UserID
	if userID.IsUnknown() || userID.IsNull() || userID.ValueString() == "" {
		userID = types.StringNull()
	}

	newState, diags := r.readIntoState(ctx, userID, state.ID.ValueString(), secret)
	if diags.HasError() {
		if isNotFoundDiag(diags) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *apiProxyCredentialResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan apiProxyCredentialModel
	var state apiProxyCredentialModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cred, diags := expandClientCredential(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var err error
	if !plan.UserID.IsNull() && plan.UserID.ValueString() != "" {
		err = r.client.UpdateUserClientCredential(plan.UserID.ValueString(), state.ID.ValueString(), cred)
	} else {
		err = r.client.UpdateCurrentUserClientCredential(state.ID.ValueString(), cred)
	}
	if err != nil {
		resp.Diagnostics.AddError("Update API proxy credential failed", err.Error())
		return
	}

	// Preserve existing secret unless you explicitly want to re-fetch it.
	secret := state.Secret
	newState, diags := r.readIntoState(ctx, plan.UserID, state.ID.ValueString(), secret)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *apiProxyCredentialResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state apiProxyCredentialModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var err error
	if !state.UserID.IsNull() && !state.UserID.IsUnknown() && state.UserID.ValueString() != "" {
		err = r.client.DeleteUserClientCredential(state.UserID.ValueString(), state.ID.ValueString())
	} else {
		err = r.client.DeleteCurrentUserClientCredential(state.ID.ValueString())
	}

	if err != nil {
		if utils.IsPrivxNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Delete API proxy credential failed", err.Error())
		return
	}
}

func (r *apiProxyCredentialResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Support:
	//  - "<cred_id>" (current user)
	//  - "<user_id>/<cred_id>" (admin managing other user)
	id := strings.TrimSpace(req.ID)
	if id == "" {
		resp.Diagnostics.AddError("Invalid import id", "Expected '<cred_id>' or '<user_id>/<cred_id>'.")
		return
	}

	if strings.Contains(id, "/") {
		parts := strings.Split(id, "/")
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			resp.Diagnostics.AddError("Invalid import id", "Expected '<user_id>/<cred_id>'.")
			return
		}
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("user_id"), parts[0])...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}

// Helpers

func expandClientCredential(ctx context.Context, m apiProxyCredentialModel) (*apiproxy.ClientCredential, diag.Diagnostics) {
	var diags diag.Diagnostics

	sourceAddrs := []string{}
	if !m.SourceAddress.IsNull() && !m.SourceAddress.IsUnknown() {
		var elems []types.String
		diags.Append(m.SourceAddress.ElementsAs(ctx, &elems, false)...)
		for _, e := range elems {
			if !e.IsNull() && !e.IsUnknown() && e.ValueString() != "" {
				sourceAddrs = append(sourceAddrs, e.ValueString())
			}
		}
	}

	cred := &apiproxy.ClientCredential{
		Name:          m.Name.ValueString(),
		Comment:       m.Comment.ValueString(),
		NotBefore:     m.NotBefore.ValueString(),
		NotAfter:      m.NotAfter.ValueString(),
		SourceAddress: sourceAddrs,
		Enabled:       m.Enabled.ValueBool(),
		Type:          m.Type.ValueString(),
		Target: apiproxy.ApiTargetHandle{
			ID: m.TargetID.ValueString(),
		},
	}

	return cred, diags
}

func (r *apiProxyCredentialResource) readIntoState(ctx context.Context, userID types.String, credID string, secret types.String) (apiProxyCredentialModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	var (
		cred *apiproxy.ClientCredential
		err  error
	)

	if !userID.IsNull() && userID.ValueString() != "" {
		cred, err = r.client.GetUserClientCredential(userID.ValueString(), credID)
	} else {
		cred, err = r.client.GetCurrentUserClientCredential(credID)
	}

	if err != nil {
		if isNotFoundError(err) {
			diags.AddError("NotFound", err.Error())
			return apiProxyCredentialModel{}, diags
		}
		diags.AddError("Read API proxy credential failed", err.Error())
		return apiProxyCredentialModel{}, diags
	}

	srcList, d := types.ListValueFrom(ctx, types.StringType, cred.SourceAddress)
	diags.Append(d...)

	state := apiProxyCredentialModel{
		ID:            types.StringValue(cred.ID),
		UserID:        userID,
		TargetID:      types.StringValue(cred.Target.ID),
		Name:          types.StringValue(cred.Name),
		Comment:       types.StringValue(cred.Comment),
		NotBefore:     types.StringValue(cred.NotBefore),
		NotAfter:      types.StringValue(cred.NotAfter),
		SourceAddress: srcList,
		Enabled:       types.BoolValue(cred.Enabled),
		Type:          types.StringValue(cred.Type),
		Secret:        secret,

		LastUsed:  types.StringValue(cred.LastUsed),
		Created:   types.StringValue(cred.Created),
		Author:    types.StringValue(cred.Author),
		Updated:   types.StringValue(cred.Updated),
		UpdatedBy: types.StringValue(cred.UpdatedBy),
	}

	return state, diags
}

func isNotFoundDiag(diags diag.Diagnostics) bool {
	for _, d := range diags {
		if strings.Contains(strings.ToLower(d.Summary()), "notfound") {
			return true
		}
	}
	return false
}
