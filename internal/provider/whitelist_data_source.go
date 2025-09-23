package provider

import (
	"context"
	"fmt"

	"github.com/SSHcom/privx-sdk-go/v2/api/hoststore"
	"github.com/SSHcom/privx-sdk-go/v2/restapi"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &WhitelistDataSource{}

func NewWhitelistDataSource() datasource.DataSource {
	return &WhitelistDataSource{}
}

// WhitelistDataSource defines the data source implementation.
type WhitelistDataSource struct {
	client *hoststore.HostStore
}

// WhitelistDataSourceModel describes the data source data model.
type WhitelistDataSourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Comment          types.String `tfsdk:"comment"`
	Type             types.String `tfsdk:"type"`
	WhitelistPatterns types.Set    `tfsdk:"whitelist_patterns"`
}

func (d *WhitelistDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_whitelist"
}

func (d *WhitelistDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Whitelist data source",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Whitelist ID",
				Optional:            true,
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Whitelist name",
				Optional:            true,
				Computed:            true,
			},
			"comment": schema.StringAttribute{
				MarkdownDescription: "Whitelist comment/description",
				Computed:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Whitelist type",
				Computed:            true,
			},
			"whitelist_patterns": schema.SetAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of command patterns allowed by this whitelist",
				Computed:            true,
			},
		},
	}
}

func (d *WhitelistDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	connector, ok := req.ProviderData.(*restapi.Connector)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *restapi.Connector, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	tflog.Debug(ctx, "Creating hoststore client for whitelist data source", map[string]interface{}{
		"connector": fmt.Sprintf("%+v", *connector),
	})

	d.client = hoststore.New(*connector)
}

func (d *WhitelistDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data WhitelistDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Validate that either ID or name is provided
	if data.ID.IsNull() && data.Name.IsNull() {
		resp.Diagnostics.AddError("Configuration Error", "Either 'id' or 'name' must be specified")
		return
	}

	var whitelist *hoststore.Whitelist
	var found bool

	// If ID is provided, try to get whitelist by ID first
	if !data.ID.IsNull() {
		wl, err := d.client.GetWhitelist(data.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read whitelist by ID, got error: %s", err))
			return
		}
		whitelist = wl
		found = true
	} else if !data.Name.IsNull() {
		// If only name is provided, search through all whitelists
		searchResult, err := d.client.GetWhitelists()
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read whitelists, got error: %s", err))
			return
		}

		for _, result := range searchResult.Items {
			if result.Name == data.Name.ValueString() {
				whitelist = &result
				found = true
				break
			}
		}

		if !found {
			resp.Diagnostics.AddError("Not Found", fmt.Sprintf("Whitelist with name '%s' not found", data.Name.ValueString()))
			return
		}
	}

	// Update the model with the retrieved data
	data.ID = types.StringValue(whitelist.ID)
	data.Name = types.StringValue(whitelist.Name)
	data.Comment = types.StringValue(whitelist.Comment)
	data.Type = types.StringValue(whitelist.Type)

	// Convert patterns to types.Set
	if len(whitelist.WhiteListPatterns) > 0 {
		patternsSet, diags := types.SetValueFrom(ctx, types.StringType, whitelist.WhiteListPatterns)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.WhitelistPatterns = patternsSet
	}

	tflog.Debug(ctx, "Read whitelist data source", map[string]interface{}{
		"whitelist_id": data.ID.ValueString(),
	})

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}