package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/SSHcom/privx-sdk-go/v2/api/userstore"
	"github.com/SSHcom/privx-sdk-go/v2/restapi"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &CarrierResource{}
var _ resource.ResourceWithImportState = &CarrierResource{}

func NewCarrierResource() resource.Resource {
	return &CarrierResource{}
}

// CarrierResource defines the resource implementation.
type CarrierResource struct {
	client *userstore.UserStore
}

// CarrierResourceModel contains PrivX carrier information.
type CarrierResourceModel struct {
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

func (r *CarrierResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_carrier"
}

func (r *CarrierResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Carrier resource for PrivX",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Carrier ID",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Carrier name",
				Required:            true,
			},
			"access_group_id": schema.StringAttribute{
				MarkdownDescription: "Access Group ID",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
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
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"registered": schema.BoolAttribute{
				MarkdownDescription: "Whether the carrier is registered",
				Computed:            true,
			},
			"permissions": schema.ListAttribute{
				MarkdownDescription: "List of permissions for the carrier",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
			},
			"routing_prefix": schema.StringAttribute{
				MarkdownDescription: "Routing prefix for the carrier",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"extender_address": schema.ListAttribute{
				MarkdownDescription: "List of extender addresses",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
			},
			"subnets": schema.ListAttribute{
				MarkdownDescription: "List of subnets for the carrier",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
			},
			"web_proxy_address": schema.StringAttribute{
				MarkdownDescription: "Web proxy address for the carrier",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"web_proxy_extender_route_patterns": schema.ListAttribute{
				MarkdownDescription: "List of web proxy extender route patterns",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
			},
		},
	}
}

func (r *CarrierResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	tflog.Debug(ctx, "Creating userstore client", map[string]interface{}{
		"connector": fmt.Sprintf("%+v", *connector),
	})

	r.client = userstore.New(*connector)
}

func (r *CarrierResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *CarrierResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Loaded carrier data", map[string]interface{}{
		"data": fmt.Sprintf("%+v", data),
	})

	// Convert permissions list to string slice
	var permissions []string
	if !data.Permissions.IsNull() && !data.Permissions.IsUnknown() {
		data.Permissions.ElementsAs(ctx, &permissions, false)
	}

	// Convert extender_address list to string slice
	var extenderAddress []string
	if !data.ExtenderAddress.IsNull() && !data.ExtenderAddress.IsUnknown() {
		data.ExtenderAddress.ElementsAs(ctx, &extenderAddress, false)
	}

	// Convert subnets list to string slice
	var subnets []string
	if !data.Subnets.IsNull() && !data.Subnets.IsUnknown() {
		data.Subnets.ElementsAs(ctx, &subnets, false)
	}

	// Convert web_proxy_extender_route_patterns list to string slice
	var webProxyRoutePatterns []string
	if !data.WebProxyExtenderRoutePatterns.IsNull() && !data.WebProxyExtenderRoutePatterns.IsUnknown() {
		data.WebProxyExtenderRoutePatterns.ElementsAs(ctx, &webProxyRoutePatterns, false)
	}

	trustedClient := userstore.TrustedClient{
		Name:                          data.Name.ValueString(),
		Type:                          "CARRIER",
		AccessGroupID:                 data.AccessGroupID.ValueString(),
		Enabled:                       data.Enabled.ValueBool(),
		Permissions:                   permissions,
		RoutingPrefix:                 data.RoutingPrefix.ValueString(),
		ExtenderAddress:               extenderAddress,
		Subnets:                       subnets,
		WebProxyAddress:               data.WebProxyAddress.ValueString(),
		WebProxyExtenderRoutePatterns: webProxyRoutePatterns,
	}

	tflog.Debug(ctx, fmt.Sprintf("userstore.TrustedClient model used: %+v", trustedClient))

	createdClient, err := r.client.CreateTrustedClient(&trustedClient)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Carrier",
			"An unexpected error occurred while attempting to create the carrier.\n"+
				err.Error(),
		)
		return
	}

	data.ID = types.StringValue(createdClient.ID)

	// Read back the created resource to populate all computed fields
	trustedClientRead, err := r.client.GetTrustedClient(createdClient.ID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read created carrier, got error: %s", err))
		return
	}

	// Populate all the computed fields from the API response
	data.Name = types.StringValue(trustedClientRead.Name)
	data.AccessGroupID = types.StringValue(trustedClientRead.AccessGroupID)
	data.GroupID = types.StringValue(trustedClientRead.GroupID)
	data.Secret = types.StringValue(trustedClientRead.Secret)
	data.Enabled = types.BoolValue(trustedClientRead.Enabled)
	data.Registered = types.BoolValue(trustedClientRead.Registered)
	data.RoutingPrefix = types.StringValue(trustedClientRead.RoutingPrefix)
	data.WebProxyAddress = types.StringValue(trustedClientRead.WebProxyAddress)

	// Convert permissions slice to list
	permissionValues := make([]attr.Value, len(trustedClientRead.Permissions))
	for i, perm := range trustedClientRead.Permissions {
		permissionValues[i] = types.StringValue(perm)
	}
	data.Permissions = types.ListValueMust(types.StringType, permissionValues)

	// Convert extender_address slice to list
	extenderValues := make([]attr.Value, len(trustedClientRead.ExtenderAddress))
	for i, addr := range trustedClientRead.ExtenderAddress {
		extenderValues[i] = types.StringValue(addr)
	}
	data.ExtenderAddress = types.ListValueMust(types.StringType, extenderValues)

	// Convert subnets slice to list
	subnetValues := make([]attr.Value, len(trustedClientRead.Subnets))
	for i, subnet := range trustedClientRead.Subnets {
		subnetValues[i] = types.StringValue(subnet)
	}
	data.Subnets = types.ListValueMust(types.StringType, subnetValues)

	// Convert web_proxy_extender_route_patterns slice to list
	routePatternValues := make([]attr.Value, len(trustedClientRead.WebProxyExtenderRoutePatterns))
	for i, pattern := range trustedClientRead.WebProxyExtenderRoutePatterns {
		routePatternValues[i] = types.StringValue(pattern)
	}
	data.WebProxyExtenderRoutePatterns = types.ListValueMust(types.StringType, routePatternValues)

	tflog.Debug(ctx, "created carrier resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CarrierResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *CarrierResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	trustedClient, err := r.client.GetTrustedClient(data.ID.ValueString())
	if err != nil {
		// Log the full error for debugging
		errorStr := err.Error()
		tflog.Debug(ctx, "Error reading carrier", map[string]interface{}{
			"id":    data.ID.ValueString(),
			"error": errorStr,
		})

		// Check if the error indicates the resource no longer exists
		// PrivX may return BAD_REQUEST when a trusted client ID doesn't exist
		if strings.Contains(errorStr, "NOT_FOUND") ||
			strings.Contains(errorStr, "404") ||
			strings.Contains(errorStr, "BAD_REQUEST") {
			// Resource likely no longer exists, remove from state
			tflog.Info(ctx, "Carrier resource appears to be deleted, removing from state", map[string]interface{}{
				"id":    data.ID.ValueString(),
				"error": errorStr,
			})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read carrier, got error: %s", err))
		return
	}

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

func (r *CarrierResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *CarrierResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	currentClient, err := r.client.GetTrustedClient(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read current carrier, got error: %s", err))
		return
	}

	currentClient.Name = data.Name.ValueString()
	currentClient.Type = "CARRIER"
	currentClient.AccessGroupID = data.AccessGroupID.ValueString()
	currentClient.Enabled = data.Enabled.ValueBool()
	currentClient.RoutingPrefix = data.RoutingPrefix.ValueString()
	currentClient.WebProxyAddress = data.WebProxyAddress.ValueString()

	// Convert permissions list to string slice
	var permissions []string
	if !data.Permissions.IsNull() && !data.Permissions.IsUnknown() {
		data.Permissions.ElementsAs(ctx, &permissions, false)
	}
	currentClient.Permissions = permissions

	// Convert extender_address list to string slice
	var extenderAddress []string
	if !data.ExtenderAddress.IsNull() && !data.ExtenderAddress.IsUnknown() {
		data.ExtenderAddress.ElementsAs(ctx, &extenderAddress, false)
	}
	currentClient.ExtenderAddress = extenderAddress

	// Convert subnets list to string slice
	var subnets []string
	if !data.Subnets.IsNull() && !data.Subnets.IsUnknown() {
		data.Subnets.ElementsAs(ctx, &subnets, false)
	}
	currentClient.Subnets = subnets

	// Convert web_proxy_extender_route_patterns list to string slice
	var webProxyRoutePatterns []string
	if !data.WebProxyExtenderRoutePatterns.IsNull() && !data.WebProxyExtenderRoutePatterns.IsUnknown() {
		data.WebProxyExtenderRoutePatterns.ElementsAs(ctx, &webProxyRoutePatterns, false)
	}
	currentClient.WebProxyExtenderRoutePatterns = webProxyRoutePatterns

	tflog.Debug(ctx, fmt.Sprintf("userstore.TrustedClient model used for update: %+v", currentClient))

	err = r.client.UpdateTrustedClient(data.ID.ValueString(), currentClient)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update carrier, got error: %s", err))
		return
	}

	// Read back the updated resource to populate all computed fields
	trustedClientRead, err := r.client.GetTrustedClient(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read updated carrier, got error: %s", err))
		return
	}

	// Populate all the computed fields from the API response
	data.Name = types.StringValue(trustedClientRead.Name)
	data.AccessGroupID = types.StringValue(trustedClientRead.AccessGroupID)
	data.GroupID = types.StringValue(trustedClientRead.GroupID)
	data.Secret = types.StringValue(trustedClientRead.Secret)
	data.Enabled = types.BoolValue(trustedClientRead.Enabled)
	data.Registered = types.BoolValue(trustedClientRead.Registered)
	data.RoutingPrefix = types.StringValue(trustedClientRead.RoutingPrefix)
	data.WebProxyAddress = types.StringValue(trustedClientRead.WebProxyAddress)

	// Convert permissions slice to list
	permissionValues := make([]attr.Value, len(trustedClientRead.Permissions))
	for i, perm := range trustedClientRead.Permissions {
		permissionValues[i] = types.StringValue(perm)
	}
	data.Permissions = types.ListValueMust(types.StringType, permissionValues)

	// Convert extender_address slice to list
	extenderValues := make([]attr.Value, len(trustedClientRead.ExtenderAddress))
	for i, addr := range trustedClientRead.ExtenderAddress {
		extenderValues[i] = types.StringValue(addr)
	}
	data.ExtenderAddress = types.ListValueMust(types.StringType, extenderValues)

	// Convert subnets slice to list
	subnetValues := make([]attr.Value, len(trustedClientRead.Subnets))
	for i, subnet := range trustedClientRead.Subnets {
		subnetValues[i] = types.StringValue(subnet)
	}
	data.Subnets = types.ListValueMust(types.StringType, subnetValues)

	// Convert web_proxy_extender_route_patterns slice to list
	routePatternValues := make([]attr.Value, len(trustedClientRead.WebProxyExtenderRoutePatterns))
	for i, pattern := range trustedClientRead.WebProxyExtenderRoutePatterns {
		routePatternValues[i] = types.StringValue(pattern)
	}
	data.WebProxyExtenderRoutePatterns = types.ListValueMust(types.StringType, routePatternValues)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CarrierResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *CarrierResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteTrustedClient(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete carrier, got error: %s", err))
		return
	}
}

func (r *CarrierResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
