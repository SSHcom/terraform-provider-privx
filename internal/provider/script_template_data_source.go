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
var _ datasource.DataSource = &ScriptTemplateDataSource{}

func NewScriptTemplateDataSource() datasource.DataSource {
	return &ScriptTemplateDataSource{}
}

type ScriptTemplateDataSource struct {
	client *secretsmanager.SecretsManager
}

type scriptTemplateDataSourceModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	OperatingSystem types.String `tfsdk:"operating_system"`
	Script          types.String `tfsdk:"script"`
}

func (d *ScriptTemplateDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_script_template"
}

func (d *ScriptTemplateDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lookup a PrivX Secrets Manager script template by id or name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Script template ID. If set, lookup is done by ID.",
			},
			"name": schema.StringAttribute{
				Optional:    true,
				Description: "Script template name. If set, lookup is done by name.",
			},
			"operating_system": schema.StringAttribute{
				Computed:    true,
				Description: "Operating system for the template (e.g. LINUX).",
			},
			"script": schema.StringAttribute{
				Computed:    true,
				Sensitive:   true,
				Description: "Template script content.",
			},
		},
	}
}

func (d *ScriptTemplateDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	// Your ProviderData is *restapi.Connector
	conn, ok := req.ProviderData.(restapi.Connector)
	if !ok {
		// sometimes it's *restapi.Connector depending on how you store it:
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

func (d *ScriptTemplateDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state scriptTemplateDataSourceModel
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

	var tmpl *secretsmanager.ScriptTemplate
	var err error

	if id != "" {
		tmpl, err = d.client.GetScriptTemplate(id)
		if err != nil {
			resp.Diagnostics.AddError("Unable to read script template", err.Error())
			return
		}
	} else {
		rs, err := d.client.GetScriptTemplates()
		if err != nil {
			resp.Diagnostics.AddError("Unable to list script templates", err.Error())
			return
		}

		// NOTE: Confirm the field name on ResultSet[T] (often Items).
		// If it’s not Items, replace rs.Items with the correct slice.
		matches := make([]secretsmanager.ScriptTemplate, 0, 1)
		for _, t := range rs.Items {
			if t.Name == name {
				matches = append(matches, t)
			}
		}

		if len(matches) == 0 {
			resp.Diagnostics.AddError("Not found", fmt.Sprintf("Script template with name %q not found.", name))
			return
		}
		if len(matches) > 1 {
			resp.Diagnostics.AddError("Not unique", fmt.Sprintf("Found %d script templates named %q. Make the name unique.", len(matches), name))
			return
		}

		tmpl = &matches[0]
	}

	state.ID = types.StringValue(tmpl.ID)
	state.Name = types.StringValue(tmpl.Name)
	state.OperatingSystem = types.StringValue(tmpl.OperatingSystem)
	state.Script = types.StringValue(tmpl.Script)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
