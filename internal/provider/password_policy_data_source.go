package provider

import (
	"context"
	"fmt"

	"github.com/SSHcom/privx-sdk-go/v2/api/secretsmanager"
	"github.com/SSHcom/privx-sdk-go/v2/restapi"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &PasswordPolicyDataSource{}

func NewPasswordPolicyDataSource() datasource.DataSource {
	return &PasswordPolicyDataSource{}
}

type PasswordPolicyDataSource struct {
	client *secretsmanager.SecretsManager
}

type passwordPolicyDataSourceModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func (d *PasswordPolicyDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_password_policy"
}

func (d *PasswordPolicyDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lookup a PrivX password policy by id or name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Password policy ID. If set, lookup is done by ID.",
			},
			"name": schema.StringAttribute{
				Optional:    true,
				Description: "Password policy name. If set, lookup is done by name.",
			},
		},
	}
}

func (d *PasswordPolicyDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	// ProviderData is restapi.Connector (same as script template DS fix)
	conn, ok := req.ProviderData.(restapi.Connector)
	if !ok {
		if ptr, ok2 := req.ProviderData.(*restapi.Connector); ok2 && ptr != nil {
			conn = *ptr
			ok = true
		}
	}
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected provider data type",
			fmt.Sprintf("Expected restapi.Connector, got: %T", req.ProviderData),
		)
		return
	}

	d.client = secretsmanager.New(conn)
}

func (d *PasswordPolicyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state passwordPolicyDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if d.client == nil {
		resp.Diagnostics.AddError("Not configured", "SecretsManager client is not configured.")
		return
	}

	id := state.ID.ValueString()
	name := state.Name.ValueString()

	if id == "" && name == "" {
		resp.Diagnostics.AddError("Missing lookup key", "Either `id` or `name` must be set.")
		return
	}
	if id != "" && name != "" {
		resp.Diagnostics.AddError("Conflicting lookup keys", "Set only one of `id` or `name`.")
		return
	}

	var policy *secretsmanager.PasswordPolicy
	var err error

	if id != "" {
		policy, err = d.client.GetPasswordPolicy(id)
		if err != nil {
			resp.Diagnostics.AddError("Unable to read password policy", err.Error())
			return
		}
	} else {
		rs, err := d.client.GetPasswordPolicies()
		if err != nil {
			resp.Diagnostics.AddError("Unable to list password policies", err.Error())
			return
		}

		matches := make([]secretsmanager.PasswordPolicy, 0, 1)
		for _, p := range rs.Items { // confirm Items is correct
			if p.Name == name {
				matches = append(matches, p)
			}
		}

		if len(matches) == 0 {
			resp.Diagnostics.AddError("Not found", fmt.Sprintf("Password policy with name %q not found.", name))
			return
		}
		if len(matches) > 1 {
			resp.Diagnostics.AddError("Not unique", fmt.Sprintf("Found %d password policies named %q.", len(matches), name))
			return
		}

		policy = &matches[0]
	}

	state.ID = types.StringValue(policy.ID)
	state.Name = types.StringValue(policy.Name)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
