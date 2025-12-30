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
var _ datasource.DataSource = &CarrierDataSource{}

func NewCarrierDataSource() datasource.DataSource {
	return &CarrierDataSource{}
}

// CarrierDataSource defines the data source implementation.
type CarrierDataSource struct {
	client *userstore.UserStore
}

// CarrierDataSourceModel describes the data source data model.
type CarrierDataSourceModel struct {
	ID                            types.String `tfsdk:"id"`
	Name                          types.String `tfsdk:"name"`
	AccessGroupID                 types.String `tfsdk:"access_group_id"`
	GroupID                       types.String `tfsdk:"group_id"`
	Secret                        types.String `tfsdk:"secret"`
	Enabled                       types.Bool   `tfsdk:"enabled"`
	Registered                    types.Bool   `tfsdk:"registered"`
	Permissions                   types.List   `tfsdk:"permissions"`
	RoutingPrefix                 types.String `tfsdk:"routing_prefix"`
	ExtenderAddress               types.List   `tfsdk:"extender_address"`
	Subnets                       types.List   `tfsdk:"subnets"`
	WebProxyAddress               types.String `tfsdk:"web_proxy_address"`
	WebProxyExtenderRoutePatterns types.List   `tfsdk:"web_proxy_extender_route_patterns"`
}

func (d *CarrierDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_carrier"
}

func (d *CarrierDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "PrivX Carrier data source",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Carrier UUID",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Carrier name",
				Required:            true,
			},
			"access_group_id": schema.StringAttribute{
				MarkdownDescription: "Access Group ID",
				Computed:            true,
			},
			"group_id": schema.StringAttribute{
				MarkdownDescription: "Group ID for the carrier",
				Computed:            true,
			},
			"secret": schema.StringAttribute{
				MarkdownDescription: "Client Secret",
				Computed:            true,
				Sensitive:           true,
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the carrier is enabled",
				Computed:            true,
			},
			"registered": schema.BoolAttribute{
				MarkdownDescription: "Whether the carrier is registered",
				Computed:            true,
			},
			"permissions": schema.ListAttribute{
				MarkdownDescription: "List of permissions for the carrier",
				ElementType:         types.StringType,
				Computed:            true,
			},
			"routing_prefix": schema.StringAttribute{
				MarkdownDescription: "Routing prefix for the carrier",
				Computed:            true,
			},
			"extender_address": schema.ListAttribute{
				MarkdownDescription: "List of extender addresses",
				ElementType:         types.StringType,
				Computed:            true,
			},
			"subnets": schema.ListAttribute{
				MarkdownDescription: "List of subnets for the carrier",
				ElementType:         types.StringType,
				Computed:            true,
			},
			"web_proxy_address": schema.StringAttribute{
				MarkdownDescription: "Web proxy address for the carrier",
				Computed:            true,
			},
			"web_proxy_extender_route_patterns": schema.ListAttribute{
				MarkdownDescription: "List of web proxy extender route patterns",
				ElementType:         types.StringType,
				Computed:            true,
			},
		},
	}
}

func (d *CarrierDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *CarrierDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data CarrierDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get by name - search through all clients
	searchResult, err := d.client.GetTrustedClients()
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to search carriers, got error: %s", err))
		return
	}

	var trustedClient *userstore.TrustedClient
	for _, client := range searchResult.Items {
		if client.Name == data.Name.ValueString() && client.Type == "CARRIER" {
			trustedClient = &client
			break
		}
	}

	if trustedClient == nil {
		resp.Diagnostics.AddError("Not Found", fmt.Sprintf("Carrier with name '%s' not found", data.Name.ValueString()))
		return
	}

	// Map the API response to the data source model
	data.ID = types.StringValue(trustedClient.ID)
	data.Name = types.StringValue(trustedClient.Name)
	data.AccessGroupID = types.StringValue(trustedClient.AccessGroupID)
	data.GroupID = types.StringValue(trustedClient.GroupID)
	data.Secret = types.StringValue(trustedClient.Secret)
	data.Enabled = types.BoolValue(trustedClient.Enabled)
	data.Registered = types.BoolValue(trustedClient.Registered)
	data.RoutingPrefix = types.StringValue(trustedClient.RoutingPrefix)
	data.WebProxyAddress = types.StringValue(trustedClient.WebProxyAddress)

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

	// Convert web_proxy_extender_route_patterns slice to list
	routePatternValues := make([]attr.Value, len(trustedClient.WebProxyExtenderRoutePatterns))
	for i, pattern := range trustedClient.WebProxyExtenderRoutePatterns {
		routePatternValues[i] = types.StringValue(pattern)
	}
	data.WebProxyExtenderRoutePatterns = types.ListValueMust(types.StringType, routePatternValues)

	tflog.Debug(ctx, "Storing carrier into the state", map[string]interface{}{
		"state": fmt.Sprintf("%+v", data),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
