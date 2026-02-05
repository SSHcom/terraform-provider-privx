package provider

import (
	"context"
	"fmt"
	"terraform-provider-privx/internal/utils"

	"github.com/SSHcom/privx-sdk-go/v2/api/userstore"
	"github.com/SSHcom/privx-sdk-go/v2/restapi"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
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

// Carrier contains PrivX carrier information.
type CarrierResourceModel struct {
	ID                            types.String `tfsdk:"id"`
	Type                          types.String `tfsdk:"type"`
	Enabled                       types.Bool   `tfsdk:"enabled"`
	RoutingPrefix                 types.String `tfsdk:"routing_prefix"`
	Name                          types.String `tfsdk:"name"`
	Permissions                   types.List   `tfsdk:"permissions"`
	WebProxyAddress               types.String `tfsdk:"web_proxy_address"`
	WebProxyPort                  types.Int64  `tfsdk:"web_proxy_port"`
	WebProxyExtenderRoutePatterns types.Set    `tfsdk:"web_proxy_extender_route_patterns"`
	ExtenderAddress               types.Set    `tfsdk:"extender_address"`
	Subnets                       types.Set    `tfsdk:"subnets"`
	Registered                    types.Bool   `tfsdk:"registered"`
	AccessGroupId                 types.String `tfsdk:"access_group_id"`
	GroupID                       types.String `tfsdk:"group_id"`
}

func (r *CarrierResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_carrier"
}

func (r *CarrierResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Carrier resource",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Carrier ID",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"type": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("CARRIER"),
			},
			"enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},

			"routing_prefix": schema.StringAttribute{
				MarkdownDescription: "Routing Prefix",
				Optional:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Carrier name",
				Required:            true,
			},
			"permissions": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Carrier permissions",
				Computed:            true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"web_proxy_address": schema.StringAttribute{
				MarkdownDescription: "Web Proxy address",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"web_proxy_port": schema.Int64Attribute{
				MarkdownDescription: "Web Proxy address",
				Optional:            true,
			},
			"web_proxy_extender_route_patterns": schema.SetAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Web Proxy Extender Route Patterns",
				Optional:            true,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
			"extender_address": schema.SetAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Extender addresses",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
			"subnets": schema.SetAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Subnets",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
			"access_group_id": schema.StringAttribute{
				Optional: true,
				Computed: true, // safe if backend sets/returns default/empty
				Default:  stringdefault.StaticString(""),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"group_id": schema.StringAttribute{
				MarkdownDescription: "Group ID",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"registered": schema.BoolAttribute{
				MarkdownDescription: "Carrier registered",
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *CarrierResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
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
	tflog.Debug(ctx, "Creating userstore", map[string]interface{}{
		"connector : ": fmt.Sprintf("%+v", *connector),
	})

	r.client = userstore.New(*connector)
}

func (r *CarrierResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data CarrierResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Loaded Carrier type data", map[string]interface{}{
		"data": fmt.Sprintf("%+v", data),
	})

	permissionPayload := []string{"privx-carrier"}

	var extenderAddressPayload []string
	if !data.ExtenderAddress.IsNull() && !data.ExtenderAddress.IsUnknown() {
		resp.Diagnostics.Append(data.ExtenderAddress.ElementsAs(ctx, &extenderAddressPayload, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	var subnetsPayload []string
	if !data.Subnets.IsNull() && !data.Subnets.IsUnknown() {
		resp.Diagnostics.Append(data.Subnets.ElementsAs(ctx, &subnetsPayload, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	var WebProxyExtenderRoutPattersPayload []string
	if !data.WebProxyExtenderRoutePatterns.IsNull() && !data.WebProxyExtenderRoutePatterns.IsUnknown() {
		resp.Diagnostics.Append(data.WebProxyExtenderRoutePatterns.ElementsAs(ctx, &WebProxyExtenderRoutPattersPayload, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	carrier := userstore.TrustedClient{
		Type:                          data.Type.ValueString(),
		Name:                          data.Name.ValueString(),
		Enabled:                       data.Enabled.ValueBool(),
		Permissions:                   permissionPayload,
		AccessGroupID:                 data.AccessGroupId.ValueString(),
		GroupID:                       data.GroupID.ValueString(),
		ExtenderAddress:               extenderAddressPayload,
		WebProxyAddress:               data.WebProxyAddress.ValueString(),
		WebProxyExtenderRoutePatterns: WebProxyExtenderRoutPattersPayload,
		Subnets:                       subnetsPayload,
		RoutingPrefix:                 data.RoutingPrefix.ValueString(),
	}

	tflog.Debug(ctx, fmt.Sprintf("userstore.TrustedClient model used: %+v", carrier))

	carrierID, err := r.client.CreateTrustedClient(&carrier)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Resource",
			"An unexpected error occurred while attempting to create the resource.\n"+
				err.Error(),
		)
		return
	}

	// Convert from the API data model to the Terraform data model
	// and set any unknown attribute values.
	data.ID = types.StringValue(carrierID.ID)

	carrierRead, err := r.client.GetTrustedClient(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Resource",
			"An unexpected error occurred while attempting to read the resource.\n"+
				err.Error(),
		)
		return
	}
	data.Registered = types.BoolValue(carrierRead.Registered)

	/*
		// access_group_id is user-defined
		// do NOT set it from API response
		if carrierRead.AccessGroupID != "" {
			data.AccessGroupId = types.StringValue(carrierRead.AccessGroupID)
		}*/

	data.GroupID = types.StringValue(carrierRead.GroupID)
	permissions, diags := types.ListValueFrom(ctx, types.StringType, carrierRead.Permissions)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Permissions = permissions

	extenderSet, diags := types.SetValueFrom(ctx, types.StringType, carrierRead.ExtenderAddress)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.ExtenderAddress = extenderSet

	data.Type = types.StringValue(carrierRead.Type)
	data.WebProxyAddress = types.StringValue(carrierRead.WebProxyAddress)
	if carrierRead.AccessGroupID != "" {
		data.AccessGroupId = types.StringValue(carrierRead.AccessGroupID)
	}

	subnetsSet, diags := types.SetValueFrom(ctx, types.StringType, carrierRead.Subnets)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Subnets = subnetsSet

	patternsSet, diags := types.SetValueFrom(ctx, types.StringType, carrierRead.WebProxyExtenderRoutePatterns)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.WebProxyExtenderRoutePatterns = patternsSet

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, "created carrier resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CarrierResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data CarrierResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	carrier, err := r.client.GetTrustedClient(data.ID.ValueString())
	if err != nil {
		if utils.IsPrivxNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read carrier, got error: %s", err))
		return
	}

	// Scalars
	data.ID = types.StringValue(carrier.ID) // if carrier.ID exists; otherwise keep existing
	data.Name = types.StringValue(carrier.Name)
	data.Enabled = types.BoolValue(carrier.Enabled)
	data.RoutingPrefix = types.StringValue(carrier.RoutingPrefix)

	// IMPORTANT: fill these (your import diff shows they are missing)
	data.Type = types.StringValue(carrier.Type)
	data.WebProxyAddress = types.StringValue(carrier.WebProxyAddress)
	if carrier.AccessGroupID != "" {
		data.AccessGroupId = types.StringValue(carrier.AccessGroupID)
	}

	data.Registered = types.BoolValue(carrier.Registered)
	data.GroupID = types.StringValue(carrier.GroupID)

	// Lists
	subnetsSet, diags := types.SetValueFrom(ctx, types.StringType, carrier.Subnets)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Subnets = subnetsSet

	perms, diags := types.ListValueFrom(ctx, types.StringType, carrier.Permissions)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Permissions = perms

	extenderSet, diags := types.SetValueFrom(ctx, types.StringType, carrier.ExtenderAddress)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.ExtenderAddress = extenderSet

	patterns, diags := types.SetValueFrom(ctx, types.StringType, carrier.WebProxyExtenderRoutePatterns)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.WebProxyExtenderRoutePatterns = patterns

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CarrierResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan CarrierResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read current object from API first
	current, err := r.client.GetTrustedClient(plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read carrier before update, got error: %s", err))
		return
	}

	// Build list payloads from plan (only if known)
	var extenderAddressPayload []string
	if !plan.ExtenderAddress.IsNull() && !plan.ExtenderAddress.IsUnknown() {
		resp.Diagnostics.Append(plan.ExtenderAddress.ElementsAs(ctx, &extenderAddressPayload, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		current.ExtenderAddress = extenderAddressPayload
	}

	var subnetsPayload []string
	if !plan.Subnets.IsNull() && !plan.Subnets.IsUnknown() {
		resp.Diagnostics.Append(plan.Subnets.ElementsAs(ctx, &subnetsPayload, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		current.Subnets = subnetsPayload
	}

	var patternsPayload []string
	if !plan.WebProxyExtenderRoutePatterns.IsNull() && !plan.WebProxyExtenderRoutePatterns.IsUnknown() {
		resp.Diagnostics.Append(plan.WebProxyExtenderRoutePatterns.ElementsAs(ctx, &patternsPayload, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		current.WebProxyExtenderRoutePatterns = patternsPayload
	}

	// Apply only the fields you actually want to change
	current.Name = plan.Name.ValueString()
	current.Enabled = plan.Enabled.ValueBool()
	current.RoutingPrefix = plan.RoutingPrefix.ValueString()
	current.WebProxyAddress = plan.WebProxyAddress.ValueString()

	// IMPORTANT: do NOT touch current.GroupID / current.AccessGroupID / current.Type / current.Permissions here
	// Let them remain exactly as the API returned.

	// Debug print (stderr) if you want
	// fmt.Fprintln(os.Stderr, "DEBUG update sending:", "id=", plan.ID.ValueString(), "enabled=", current.Enabled, "routing_prefix=", current.RoutingPrefix)

	if err := r.client.UpdateTrustedClient(plan.ID.ValueString(), current); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update carrier, got error: %s", err))
		return
	}

	// Re-read and store state
	updated, err := r.client.GetTrustedClient(plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read updated carrier, got error: %s", err))
		return
	}

	// Populate state from API (preserve access_group_id behavior if API omits it)
	plan.Type = types.StringValue(updated.Type)
	plan.Name = types.StringValue(updated.Name)
	plan.Enabled = types.BoolValue(updated.Enabled)
	plan.RoutingPrefix = types.StringValue(updated.RoutingPrefix)
	plan.WebProxyAddress = types.StringValue(updated.WebProxyAddress)
	plan.Registered = types.BoolValue(updated.Registered)
	plan.GroupID = types.StringValue(updated.GroupID)

	if updated.AccessGroupID != "" {
		plan.AccessGroupId = types.StringValue(updated.AccessGroupID)
	}

	perms, diags := types.ListValueFrom(ctx, types.StringType, updated.Permissions)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.Permissions = perms

	extSet, diags := types.SetValueFrom(ctx, types.StringType, updated.ExtenderAddress)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ExtenderAddress = extSet

	subsSet, diags := types.SetValueFrom(ctx, types.StringType, updated.Subnets)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.Subnets = subsSet

	patsSet, diags := types.SetValueFrom(ctx, types.StringType, updated.WebProxyExtenderRoutePatterns)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.WebProxyExtenderRoutePatterns = patsSet

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *CarrierResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data CarrierResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteTrustedClient(data.ID.ValueString())
	if err != nil {
		if utils.IsPrivxNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete carrier, got error: %s", err))
		return
	}
}

func (r *CarrierResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
