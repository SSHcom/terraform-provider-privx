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
var _ resource.Resource = &ExtenderResource{}
var _ resource.ResourceWithImportState = &ExtenderResource{}

func NewExtenderResource() resource.Resource {
	return &ExtenderResource{}
}

// ExtenderResource defines the resource implementation.
type ExtenderResource struct {
	client *userstore.UserStore
}

// ExtenderResourceModel contains PrivX Extender information.
type ExtenderResourceModel struct {
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

func (r *ExtenderResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_extender"
}

func (r *ExtenderResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Extender resource for PrivX",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Extender ID",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Extender name",
				Required:            true,
			},

			"access_group_id": schema.StringAttribute{
				MarkdownDescription: "Access Group ID",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},

			"secret": schema.StringAttribute{
				MarkdownDescription: "Client Secret",
				Computed:            true,
				Sensitive:           true,
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the Extender is enabled",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"permissions": schema.ListAttribute{
				MarkdownDescription: "List of permissions for the Extender",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
			},
			"routing_prefix": schema.StringAttribute{
				MarkdownDescription: "Routing prefix for the Extender",
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
				MarkdownDescription: "List of subnets for the Extender",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
			},

			"registered": schema.BoolAttribute{
				MarkdownDescription: "Whether the Extender is registered",
				Computed:            true,
			},
		},
	}
}

func (r *ExtenderResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ExtenderResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *ExtenderResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Loaded Extender data", map[string]interface{}{
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

	trustedClient := userstore.TrustedClient{
		Name:            data.Name.ValueString(),
		Type:            "EXTENDER",
		AccessGroupID:   data.AccessGroupID.ValueString(),
		Enabled:         data.Enabled.ValueBool(),
		Permissions:     permissions,
		RoutingPrefix:   data.RoutingPrefix.ValueString(),
		ExtenderAddress: extenderAddress,
		Subnets:         subnets,
	}

	tflog.Debug(ctx, fmt.Sprintf("userstore.TrustedClient model used: %+v", trustedClient))

	createdClient, err := r.client.CreateTrustedClient(&trustedClient)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Extender",
			"An unexpected error occurred while attempting to create the Extender.\n"+
				err.Error(),
		)
		return
	}

	data.ID = types.StringValue(createdClient.ID)

	// Read back the created resource to populate all computed fields
	trustedClientRead, err := r.client.GetTrustedClient(createdClient.ID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read created Extender, got error: %s", err))
		return
	}

	// Populate all the computed fields from the API response
	data.Name = types.StringValue(trustedClientRead.Name)
	data.AccessGroupID = types.StringValue(trustedClientRead.AccessGroupID)
	data.Secret = types.StringValue(trustedClientRead.Secret)
	data.Enabled = types.BoolValue(trustedClientRead.Enabled)
	data.Registered = types.BoolValue(trustedClientRead.Registered)
	data.RoutingPrefix = types.StringValue(trustedClientRead.RoutingPrefix)

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

	tflog.Debug(ctx, "created Extender resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ExtenderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *ExtenderResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	trustedClient, err := r.client.GetTrustedClient(data.ID.ValueString())
	if err != nil {
		// Log the full error for debugging
		errorStr := err.Error()
		tflog.Debug(ctx, "Error reading Extender", map[string]interface{}{
			"id":    data.ID.ValueString(),
			"error": errorStr,
		})

		// Check if the error indicates the resource no longer exists
		// PrivX may return BAD_REQUEST when a trusted client ID doesn't exist
		if strings.Contains(errorStr, "NOT_FOUND") ||
			strings.Contains(errorStr, "404") ||
			strings.Contains(errorStr, "BAD_REQUEST") {
			// Resource likely no longer exists, remove from state
			tflog.Info(ctx, "Extender resource appears to be deleted, removing from state", map[string]interface{}{
				"id":    data.ID.ValueString(),
				"error": errorStr,
			})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read Extender, got error: %s", err))
		return
	}

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

func (r *ExtenderResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *ExtenderResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	currentClient, err := r.client.GetTrustedClient(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read current Extender, got error: %s", err))
		return
	}

	currentClient.Name = data.Name.ValueString()
	currentClient.Type = "EXTENDER"
	currentClient.AccessGroupID = data.AccessGroupID.ValueString()
	currentClient.Enabled = data.Enabled.ValueBool()
	currentClient.RoutingPrefix = data.RoutingPrefix.ValueString()

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

	tflog.Debug(ctx, fmt.Sprintf("userstore.TrustedClient model used for update: %+v", currentClient))

	err = r.client.UpdateTrustedClient(data.ID.ValueString(), currentClient)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update Extender, got error: %s", err))
		return
	}

	// Read back the updated resource to populate all computed fields
	trustedClientRead, err := r.client.GetTrustedClient(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read updated Extender, got error: %s", err))
		return
	}

	// Populate all the computed fields from the API response
	data.Name = types.StringValue(trustedClientRead.Name)
	data.AccessGroupID = types.StringValue(trustedClientRead.AccessGroupID)
	data.Secret = types.StringValue(trustedClientRead.Secret)
	data.Enabled = types.BoolValue(trustedClientRead.Enabled)
	data.Registered = types.BoolValue(trustedClientRead.Registered)
	data.RoutingPrefix = types.StringValue(trustedClientRead.RoutingPrefix)

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

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ExtenderResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *ExtenderResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteTrustedClient(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete Extender, got error: %s", err))
		return
	}
}

func (r *ExtenderResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
