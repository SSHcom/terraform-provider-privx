package provider

import (
	"context"
	"fmt"

	"github.com/SSHcom/privx-sdk-go/v2/api/userstore"
	"github.com/SSHcom/privx-sdk-go/v2/restapi"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &ExtenderDataSource{}

func NewExtenderDataSource() datasource.DataSource {
	return &ExtenderDataSource{}
}

// ExtenderDataSource defines the data source implementation.
type ExtenderDataSource struct {
	client *userstore.UserStore
}

// ExtenderDataSourceModel describes the data source data model.
type ExtenderDataSourceModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	AccessGroupID   types.String `tfsdk:"access_group_id"`
	Secret          types.String `tfsdk:"secret"`
	Enabled         types.Bool   `tfsdk:"enabled"`
	Registered      types.Bool   `tfsdk:"registered"`
	Permissions     types.List   `tfsdk:"permissions"`
	RoutingPrefix   types.String `tfsdk:"routing_prefix"`
	ExtenderAddress types.List   `tfsdk:"extender_address"`
	Subnets         types.List   `tfsdk:"subnets"`
}

func (d *ExtenderDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_extender"
}

func (d *ExtenderDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Extender data source",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Extender UUID",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Extender name",
				Required:            true,
			},

			"access_group_id": schema.StringAttribute{
				MarkdownDescription: "Access Group ID",
				Computed:            true,
			},

			"secret": schema.StringAttribute{
				MarkdownDescription: "Client Secret",
				Computed:            true,
				Sensitive:           true,
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the Extender is enabled",
				Computed:            true,
			},

			"permissions": schema.ListAttribute{
				MarkdownDescription: "List of permissions for the Extender",
				ElementType:         types.StringType,
				Computed:            true,
			},
			"routing_prefix": schema.StringAttribute{
				MarkdownDescription: "Routing prefix for the Extender",
				Computed:            true,
			},
			"extender_address": schema.ListAttribute{
				MarkdownDescription: "List of extender addresses",
				ElementType:         types.StringType,
				Computed:            true,
			},
			"subnets": schema.ListAttribute{
				MarkdownDescription: "List of subnets for the Extender",
				ElementType:         types.StringType,
				Computed:            true,
			},

			"registered": schema.BoolAttribute{
				MarkdownDescription: "Whether the Extender is registered",
				Computed:            true,
			},
		},
	}
}

func (d *ExtenderDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	tflog.Debug(ctx, "Creating userstore client", map[string]interface{}{
		"connector": fmt.Sprintf("%+v", *connector),
	})

	d.client = userstore.New(*connector)
}

func (d *ExtenderDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ExtenderDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get by name - search through all clients
	searchResult, err := d.client.GetTrustedClients()
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to search Extenders, got error: %s", err))
		return
	}

	var trustedClient *userstore.TrustedClient
	for _, client := range searchResult.Items {
		if client.Name == data.Name.ValueString() {
			trustedClient = &client
			break
		}
	}

	if trustedClient == nil {
		resp.Diagnostics.AddError("Not Found", fmt.Sprintf("Extender with name '%s' not found", data.Name.ValueString()))
		return
	}

	// Map the API response to the data source model
	data.ID = types.StringValue(trustedClient.ID)
	data.Name = types.StringValue(trustedClient.Name)
	data.AccessGroupID = types.StringValue(trustedClient.AccessGroupID)
	data.Secret = types.StringValue(trustedClient.Secret)
	data.Enabled = types.BoolValue(trustedClient.Enabled)
	data.Registered = types.BoolValue(trustedClient.Registered)
	data.RoutingPrefix = types.StringValue(trustedClient.RoutingPrefix)

	// Convert permissions slice to list
	permissionValues := make([]attr.Value, len(trustedClient.Permissions))
	for i, perm := range trustedClient.Permissions {
		permissionValues[i] = types.StringValue(perm)
	}
	data.Permissions = types.ListValueMust(types.StringType, permissionValues)

	// Convert extender_address slice to list
	extenderValues := make([]attr.Value, len(trustedClient.ExtenderAddress))
	for i, addr := range trustedClient.ExtenderAddress {
		extenderValues[i] = types.StringValue(addr)
	}
	data.ExtenderAddress = types.ListValueMust(types.StringType, extenderValues)

	// Convert subnets slice to list
	subnetValues := make([]attr.Value, len(trustedClient.Subnets))
	for i, subnet := range trustedClient.Subnets {
		subnetValues[i] = types.StringValue(subnet)
	}
	data.Subnets = types.ListValueMust(types.StringType, subnetValues)

	tflog.Debug(ctx, "Storing Extender into the state", map[string]interface{}{
		"state": fmt.Sprintf("%+v", data),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
