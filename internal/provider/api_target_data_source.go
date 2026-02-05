package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/SSHcom/privx-sdk-go/v2/api/apiproxy"
	"github.com/SSHcom/privx-sdk-go/v2/restapi"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &APITargetDataSource{}
var _ datasource.DataSourceWithConfigure = &APITargetDataSource{}

type APITargetDataSource struct {
	client *apiproxy.ApiProxy
}

func NewAPITargetDataSource() datasource.DataSource {
	return &APITargetDataSource{}
}

func (d *APITargetDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_target"
}

type APITargetDataSourceModel struct {
	// selector
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`

	// outputs
	Comment               types.String             `tfsdk:"comment"`
	Tags                  types.Set                `tfsdk:"tags"`
	AccessGroupID         types.String             `tfsdk:"access_group_id"`
	Roles                 []RolesRefModel          `tfsdk:"roles"`
	AuthorizedEndpoints   []ApiTargetEndpointModel `tfsdk:"authorized_endpoints"`
	UnauthorizedEndpoints []ApiTargetEndpointModel `tfsdk:"unauthorized_endpoints"`
	TLSTrustAnchors       types.String             `tfsdk:"tls_trust_anchors"`
	TLSInsecureSkipVerify types.Bool               `tfsdk:"tls_insecure_skip_verify"`
	TargetCredential      *TargetCredentialModel   `tfsdk:"target_credential"`
	Disabled              types.String             `tfsdk:"disabled"`
	AuditEnabled          types.Bool               `tfsdk:"audit_enabled"`
	Created               types.String             `tfsdk:"created"`
	Author                types.String             `tfsdk:"author"`
	Updated               types.String             `tfsdk:"updated"`
	UpdatedBy             types.String             `tfsdk:"updated_by"`
}

func (d *APITargetDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lookup an existing PrivX API target (API-Proxy).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:            true,
				Computed:            true, // allow name-only lookup to populate id
				MarkdownDescription: "ID of the API target. Set exactly one of `id` or `name`.",
			},
			"name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Name of the API target. Set exactly one of `id` or `name`.",
			},

			"comment":         schema.StringAttribute{Computed: true},
			"tags":            schema.SetAttribute{Computed: true, ElementType: types.StringType},
			"access_group_id": schema.StringAttribute{Computed: true},

			// FIX: model is []RolesRefModel => use ListNestedAttribute (not SetNestedAttribute)
			"roles": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":   schema.StringAttribute{Computed: true},
						"name": schema.StringAttribute{Computed: true},
					},
				},
			},

			"authorized_endpoints": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"host":                  schema.StringAttribute{Computed: true},
						"protocols":             schema.SetAttribute{Computed: true, ElementType: types.StringType},
						"methods":               schema.SetAttribute{Computed: true, ElementType: types.StringType},
						"paths":                 schema.SetAttribute{Computed: true, ElementType: types.StringType},
						"allow_unauthenticated": schema.BoolAttribute{Computed: true},
						"nat_target_host":       schema.StringAttribute{Computed: true},
					},
				},
			},

			"unauthorized_endpoints": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"host":                  schema.StringAttribute{Computed: true},
						"protocols":             schema.SetAttribute{Computed: true, ElementType: types.StringType},
						"methods":               schema.SetAttribute{Computed: true, ElementType: types.StringType},
						"paths":                 schema.SetAttribute{Computed: true, ElementType: types.StringType},
						"allow_unauthenticated": schema.BoolAttribute{Computed: true},
						"nat_target_host":       schema.StringAttribute{Computed: true},
					},
				},
			},

			"tls_trust_anchors": schema.StringAttribute{
				Computed:  true,
				Sensitive: true,
			},
			"tls_insecure_skip_verify": schema.BoolAttribute{Computed: true},

			"target_credential": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"type":                schema.StringAttribute{Computed: true},
					"basic_auth_username": schema.StringAttribute{Computed: true},
					"basic_auth_password": schema.StringAttribute{Computed: true, Sensitive: true},
					"bearer_token":        schema.StringAttribute{Computed: true, Sensitive: true},
					"certificate":         schema.StringAttribute{Computed: true, Sensitive: true},
					"private_key":         schema.StringAttribute{Computed: true, Sensitive: true},
				},
			},

			"disabled":      schema.StringAttribute{Computed: true},
			"audit_enabled": schema.BoolAttribute{Computed: true},
			"created":       schema.StringAttribute{Computed: true},
			"author":        schema.StringAttribute{Computed: true},
			"updated":       schema.StringAttribute{Computed: true},
			"updated_by":    schema.StringAttribute{Computed: true},
		},
	}
}

func (d *APITargetDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	connector, ok := req.ProviderData.(*restapi.Connector)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected DataSource Configure Type",
			fmt.Sprintf("Expected *restapi.Connector, got: %T", req.ProviderData),
		)
		return
	}
	d.client = apiproxy.New(*connector)
}

func (d *APITargetDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var cfg APITargetDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := ""
	if !cfg.ID.IsNull() && !cfg.ID.IsUnknown() {
		id = strings.TrimSpace(cfg.ID.ValueString())
	}
	name := ""
	if !cfg.Name.IsNull() && !cfg.Name.IsUnknown() {
		name = strings.TrimSpace(cfg.Name.ValueString())
	}

	if (id == "" && name == "") || (id != "" && name != "") {
		resp.Diagnostics.AddError("Invalid configuration", "Set exactly one of `id` or `name`.")
		return
	}

	var remote *apiproxy.ApiTarget
	var err error

	if id != "" {
		remote, err = d.client.GetApiTarget(id)
	} else {
		remote, err = findApiTargetByName(d.client, name)
	}
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read api_target: %s", err))
		return
	}

	out, diags := flattenApiTargetDataSource(ctx, remote)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &out)...)
}

func flattenApiTargetDataSource(ctx context.Context, remote *apiproxy.ApiTarget) (APITargetDataSourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	var out APITargetDataSourceModel

	out.ID = types.StringValue(remote.ID)
	out.Name = types.StringValue(remote.Name)
	out.Comment = types.StringValue(remote.Comment)

	// tags
	if len(remote.Tags) == 0 {
		out.Tags = types.SetNull(types.StringType)
	} else {
		set, d := types.SetValueFrom(ctx, types.StringType, remote.Tags)
		diags.Append(d...)
		out.Tags = set
	}

	out.AccessGroupID = types.StringNull()
	if strings.TrimSpace(remote.AccessGroupID) != "" {
		out.AccessGroupID = types.StringValue(remote.AccessGroupID)
	}

	// roles
	roles := make([]RolesRefModel, 0, len(remote.Roles))
	for _, rr := range remote.Roles {
		r := RolesRefModel{
			ID: types.StringValue(rr.ID),
		}
		if strings.TrimSpace(rr.Name) == "" {
			r.Name = types.StringNull()
		} else {
			r.Name = types.StringValue(rr.Name)
		}
		roles = append(roles, r)
	}
	out.Roles = roles

	// endpoints (reuse your existing helper if available)
	out.AuthorizedEndpoints, diags = flattenEndpoints(ctx, remote.AuthorizedEndpoints, diags)

	if len(remote.UnauthorizedEndpoints) == 0 {
		out.UnauthorizedEndpoints = nil
	} else {
		out.UnauthorizedEndpoints, diags = flattenEndpoints(ctx, remote.UnauthorizedEndpoints, diags)
	}

	out.TLSInsecureSkipVerify = types.BoolValue(remote.TLSInsecureSkipVerify)

	// tls trust anchors: DS should not "preserve"; just return what API gives (or null if masked/empty)
	ta := strings.TrimSpace(remote.TLSTrustAnchors)
	if ta == "" || isMaskedSecret(ta) {
		out.TLSTrustAnchors = types.StringNull()
	} else {
		// normalize newline to reduce diffs
		if !strings.HasSuffix(ta, "\n") {
			ta += "\n"
		}
		out.TLSTrustAnchors = types.StringValue(ta)
	}

	out.Disabled = types.StringNull()
	if strings.TrimSpace(remote.Disabled) != "" {
		out.Disabled = types.StringValue(remote.Disabled)
	}

	out.AuditEnabled = types.BoolValue(remote.AuditEnabled)

	out.Created = types.StringNull()
	if strings.TrimSpace(remote.Created) != "" {
		out.Created = types.StringValue(remote.Created)
	}
	out.Author = types.StringNull()
	if strings.TrimSpace(remote.Author) != "" {
		out.Author = types.StringValue(remote.Author)
	}
	out.Updated = types.StringNull()
	if strings.TrimSpace(remote.Updated) != "" {
		out.Updated = types.StringValue(remote.Updated)
	}
	out.UpdatedBy = types.StringNull()
	if strings.TrimSpace(remote.UpdatedBy) != "" {
		out.UpdatedBy = types.StringValue(remote.UpdatedBy)
	}

	out.TargetCredential = flattenTargetCredentialDataSource(remote.TargetCredential)

	return out, diags
}

// --- unchanged: your findApiTargetByName (but ensure rs.Items exists in ResultSet) ---
// func findApiTargetByName(...)

func flattenTargetCredentialDataSource(remote apiproxy.TargetCredential) *TargetCredentialModel {
	if strings.TrimSpace(remote.Type) == "" &&
		strings.TrimSpace(remote.BasicAuthUsername) == "" &&
		strings.TrimSpace(remote.BasicAuthPassword) == "" &&
		strings.TrimSpace(remote.BearerToken) == "" &&
		strings.TrimSpace(remote.Certificate) == "" &&
		strings.TrimSpace(remote.PrivateKey) == "" {
		return nil
	}

	mk := func(v string) types.String {
		v = strings.TrimSpace(v)
		if v == "" || isMaskedSecret(v) {
			return types.StringNull()
		}
		return types.StringValue(v)
	}

	return &TargetCredentialModel{
		Type:              mk(remote.Type),
		BasicAuthUsername: mk(remote.BasicAuthUsername),
		BasicAuthPassword: mk(remote.BasicAuthPassword),
		BearerToken:       mk(remote.BearerToken),
		Certificate:       mk(remote.Certificate),
		PrivateKey:        mk(remote.PrivateKey),
	}
}

func findApiTargetByName(client *apiproxy.ApiProxy, name string) (*apiproxy.ApiTarget, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("name is empty")
	}

	// Server-side search (exact constraint supported by ApiTargetSearchRequest.Name)
	searchReq := &apiproxy.ApiTargetSearchRequest{
		Name: name,
	}

	rs, err := client.SearchApiTargets(searchReq)
	if err == nil && rs != nil {
		// exact match first (names are unique per ApiTarget comment)
		for i := range rs.Items {
			if rs.Items[i].Name == name {
				return client.GetApiTarget(rs.Items[i].ID)
			}
		}

		// if backend search is fuzzy, resolve case-insensitively but require uniqueness
		var matches []apiproxy.ApiTarget
		for _, it := range rs.Items {
			if strings.EqualFold(it.Name, name) {
				matches = append(matches, it)
			}
		}
		if len(matches) == 1 {
			return client.GetApiTarget(matches[0].ID)
		}
		if len(matches) > 1 {
			return nil, fmt.Errorf("multiple api targets matched name %q (case-insensitive); please use id instead", name)
		}
		// else fall through to list()
	}

	// Fallback: list all
	list, err2 := client.GetApiTargets()
	if err2 != nil {
		if err != nil {
			return nil, fmt.Errorf("search api targets failed: %v; list api targets failed: %w", err, err2)
		}
		return nil, err2
	}
	if list == nil {
		return nil, fmt.Errorf("api targets list is nil")
	}

	for i := range list.Items {
		if list.Items[i].Name == name {
			return client.GetApiTarget(list.Items[i].ID)
		}
	}

	var matches []apiproxy.ApiTarget
	for _, it := range list.Items {
		if strings.EqualFold(it.Name, name) {
			matches = append(matches, it)
		}
	}
	if len(matches) == 1 {
		return client.GetApiTarget(matches[0].ID)
	}
	if len(matches) > 1 {
		return nil, fmt.Errorf("multiple api targets matched name %q (case-insensitive); please use id instead", name)
	}

	return nil, fmt.Errorf("api target with name %q not found", name)
}
