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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
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

// Extender contains PrivX extender information.
type ExtenderResourceModel struct {
	ID              types.String `tfsdk:"id"`
	Enabled         types.Bool   `tfsdk:"enabled"`
	RoutingPrefix   types.String `tfsdk:"routing_prefix"`
	Name            types.String `tfsdk:"name"`
	Permissions     types.List   `tfsdk:"permissions"`
	WebProxyAddress types.String `tfsdk:"web_proxy_address"`
	WebProxyPort    types.Int64  `tfsdk:"web_proxy_port"`
	ExtenderAddress types.List   `tfsdk:"extender_address"`
	Subnets         types.List   `tfsdk:"subnets"`
	Registered      types.Bool   `tfsdk:"registered"`
	AccessGroupId   types.String `tfsdk:"access_group_id"`
}

func (r *ExtenderResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_extender"
}

func (r *ExtenderResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Extender resource",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Extender ID",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Extender enabled (read-only, managed by PrivX).",
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"routing_prefix": schema.StringAttribute{
				MarkdownDescription: "Routing Prefix",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Extender name",
				Required:            true,
			},
			"permissions": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Extender permissions",
				Computed:            true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"web_proxy_address": schema.StringAttribute{
				MarkdownDescription: "Web Proxy address",
				Optional:            true,
			},
			"web_proxy_port": schema.Int64Attribute{
				MarkdownDescription: "Web Proxy address",
				Optional:            true,
			},
			"extender_address": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Extender addresses",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"subnets": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Subnets",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"access_group_id": schema.StringAttribute{
				MarkdownDescription: "Access Group ID",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"registered": schema.BoolAttribute{
				MarkdownDescription: "Extender registered",
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *ExtenderResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ExtenderResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ExtenderResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Loaded extender type data", map[string]interface{}{
		"data": fmt.Sprintf("%+v", data),
	})

	//extenderPermissionPayload := []string{"READ", "WRITE", "CONNECT"}

	var extenderAddressPayload []string
	if len(data.ExtenderAddress.Elements()) > 0 {
		resp.Diagnostics.Append(data.ExtenderAddress.ElementsAs(ctx, &extenderAddressPayload, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	var extenderSubnetsPayload []string
	if len(data.Subnets.Elements()) > 0 {
		resp.Diagnostics.Append(data.Subnets.ElementsAs(ctx, &extenderSubnetsPayload, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	ag := ""
	if !data.AccessGroupId.IsNull() && !data.AccessGroupId.IsUnknown() {
		ag = data.AccessGroupId.ValueString()
	}

	rp := ""
	if !data.RoutingPrefix.IsNull() && !data.RoutingPrefix.IsUnknown() {
		rp = data.RoutingPrefix.ValueString()
	}

	extender := userstore.TrustedClient{
		Type:    "EXTENDER",
		Name:    data.Name.ValueString(),
		Enabled: true, // explicitly safe default
		//Permissions:     extenderPermissionPayload,
		AccessGroupID:   ag,
		ExtenderAddress: extenderAddressPayload,
		Subnets:         extenderSubnetsPayload,
		RoutingPrefix:   rp,
	}
	tflog.Debug(ctx, fmt.Sprintf("userstore.Extender model used: %+v", extender))

	extenderID, err := r.client.CreateTrustedClient(&extender)

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
	data.ID = types.StringValue(extenderID.ID)

	extenderRead, err := r.client.GetTrustedClient(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Resource",
			"An unexpected error occurred while attempting to read the resource.\n"+
				err.Error(),
		)
		return
	}
	data.Enabled = types.BoolValue(extenderRead.Enabled)
	data.Registered = types.BoolValue(extenderRead.Registered)

	if extenderRead.AccessGroupID == "" {
		data.AccessGroupId = types.StringNull()
	} else {
		data.AccessGroupId = types.StringValue(extenderRead.AccessGroupID)
	}

	permissions, diags := types.ListValueFrom(ctx, types.StringType, extenderRead.Permissions)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Permissions = permissions

	// extender_address from server
	extenderAddress, diags := types.ListValueFrom(ctx, types.StringType, extenderRead.ExtenderAddress)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	data.ExtenderAddress = extenderAddress

	// subnets from server
	subnets, diags := types.ListValueFrom(ctx, types.StringType, extenderRead.Subnets)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	data.Subnets = subnets
	if extenderRead.RoutingPrefix == "" {
		// If server returns empty when unset, reflect it as null in state
		data.RoutingPrefix = types.StringNull()
	} else {
		data.RoutingPrefix = types.StringValue(extenderRead.RoutingPrefix)
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, "created extender resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ExtenderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *ExtenderResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	extender, err := r.client.GetTrustedClient(data.ID.ValueString())
	if err != nil {
		if utils.IsPrivxNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read extender, got error: %s", err))
		return
	}

	data.Name = types.StringValue(extender.Name)
	data.Registered = types.BoolValue(extender.Registered)
	data.Enabled = types.BoolValue(extender.Enabled)
	if extender.RoutingPrefix == "" {
		// If server returns empty when unset, reflect it as null in state
		data.RoutingPrefix = types.StringNull()
	} else {
		data.RoutingPrefix = types.StringValue(extender.RoutingPrefix)
	}

	subnets, diags := types.ListValueFrom(ctx, types.StringType, extender.Subnets)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Subnets = subnets

	permissions, diags := types.ListValueFrom(ctx, types.StringType, extender.Permissions)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Permissions = permissions

	if extender.AccessGroupID == "" {
		data.AccessGroupId = types.StringNull() // or keep prior if you prefer
	} else {
		data.AccessGroupId = types.StringValue(extender.AccessGroupID)
	}

	extenderAddress, diags := types.ListValueFrom(ctx, types.StringType, extender.ExtenderAddress)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.ExtenderAddress = extenderAddress

	tflog.Debug(ctx, "Storing extender type into the state", map[string]interface{}{
		"createNewState": fmt.Sprintf("%+v", data),
	})
	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ExtenderResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ExtenderResourceModel
	var state ExtenderResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ag := ""
	if !state.AccessGroupId.IsNull() && !state.AccessGroupId.IsUnknown() {
		ag = state.AccessGroupId.ValueString()
	}

	rp := ""
	if !state.RoutingPrefix.IsNull() && !state.RoutingPrefix.IsUnknown() {
		rp = state.RoutingPrefix.ValueString()
	}
	if !plan.RoutingPrefix.IsNull() && !plan.RoutingPrefix.IsUnknown() {
		rp = plan.RoutingPrefix.ValueString()
	}

	current, err := r.client.GetTrustedClient(plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to read extender before update: %s", err),
		)
		return
	}

	extender := userstore.TrustedClient{
		Type:          "EXTENDER",
		Name:          plan.Name.ValueString(),
		AccessGroupID: ag,
		RoutingPrefix: rp,
		// 🔒 read-only, preserved
		Enabled: current.Enabled,
	}

	if plan.ExtenderAddress.IsNull() || plan.ExtenderAddress.IsUnknown() {
		_ = state.ExtenderAddress.ElementsAs(ctx, &extender.ExtenderAddress, false)
	} else {
		resp.Diagnostics.Append(plan.ExtenderAddress.ElementsAs(ctx, &extender.ExtenderAddress, false)...)
	}

	if plan.Subnets.IsNull() || plan.Subnets.IsUnknown() {
		_ = state.Subnets.ElementsAs(ctx, &extender.Subnets, false)
	} else {
		resp.Diagnostics.Append(plan.Subnets.ElementsAs(ctx, &extender.Subnets, false)...)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.UpdateTrustedClient(plan.ID.ValueString(), &extender); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update extender, got error: %s", err))
		return
	}

	updated, err := r.client.GetTrustedClient(plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read extender after update, got error: %s", err))
		return
	}

	plan.Enabled = types.BoolValue(updated.Enabled)
	plan.Registered = types.BoolValue(updated.Registered)
	if updated.AccessGroupID == "" {
		if !state.AccessGroupId.IsNull() && !state.AccessGroupId.IsUnknown() {
			plan.AccessGroupId = state.AccessGroupId
		} else {
			plan.AccessGroupId = types.StringNull()
		}
	} else {
		plan.AccessGroupId = types.StringValue(updated.AccessGroupID)
	}

	if updated.RoutingPrefix == "" {
		plan.RoutingPrefix = types.StringNull()
	} else {
		plan.RoutingPrefix = types.StringValue(updated.RoutingPrefix)
	}

	perms, diags := types.ListValueFrom(ctx, types.StringType, updated.Permissions)
	resp.Diagnostics.Append(diags...)
	addr, diags := types.ListValueFrom(ctx, types.StringType, updated.ExtenderAddress)
	resp.Diagnostics.Append(diags...)
	subs, diags := types.ListValueFrom(ctx, types.StringType, updated.Subnets)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	plan.Permissions = perms
	plan.ExtenderAddress = addr
	plan.Subnets = subs

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ExtenderResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *ExtenderResourceModel

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
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete extender, got error: %s", err))
		return
	}
}

func (r *ExtenderResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
