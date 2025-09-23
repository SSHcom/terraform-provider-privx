package provider

import (
	"context"
	"fmt"

	"github.com/SSHcom/privx-sdk-go/v2/api/hoststore"
	"github.com/SSHcom/privx-sdk-go/v2/restapi"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &WhitelistResource{}
var _ resource.ResourceWithImportState = &WhitelistResource{}

func NewWhitelistResource() resource.Resource {
	return &WhitelistResource{}
}

type (
	// WhitelistResource defines the resource implementation.
	WhitelistResource struct {
		client *hoststore.HostStore
	}

	// WhitelistResourceModel describes the resource data model.
	WhitelistResourceModel struct {
		ID                types.String `tfsdk:"id"`
		Name              types.String `tfsdk:"name"`
		Comment           types.String `tfsdk:"comment"`
		Type              types.String `tfsdk:"type"`
		WhitelistPatterns types.Set    `tfsdk:"whitelist_patterns"`
	}
)

func (r *WhitelistResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_whitelist"
}

func (r *WhitelistResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Whitelist resource for command restrictions",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Whitelist ID",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Whitelist name",
				Required:            true,
			},
			"comment": schema.StringAttribute{
				MarkdownDescription: "Whitelist comment/description",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Whitelist type",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"whitelist_patterns": schema.SetAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of command patterns allowed by this whitelist",
				Optional:            true,
			},
		},
	}
}

func (r *WhitelistResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	tflog.Debug(ctx, "Creating hoststore client for whitelist", map[string]interface{}{
		"connector": fmt.Sprintf("%+v", *connector),
	})

	r.client = hoststore.New(*connector)
}

func (r *WhitelistResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data WhitelistResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating whitelist", map[string]interface{}{
		"data": fmt.Sprintf("%+v", data),
	})

	// Convert whitelist_patterns from types.Set to []string
	var patternsPayload []string
	if !data.WhitelistPatterns.IsNull() && !data.WhitelistPatterns.IsUnknown() {
		resp.Diagnostics.Append(data.WhitelistPatterns.ElementsAs(ctx, &patternsPayload, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Create whitelist using hoststore API
	whitelist := hoststore.Whitelist{
		Name:              data.Name.ValueString(),
		Comment:           data.Comment.ValueString(),
		Type:              data.Type.ValueString(),
		WhiteListPatterns: patternsPayload,
	}

	// Call the SDK method to create whitelist
	identifier, err := r.client.CreateWhitelist(&whitelist)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create whitelist, got error: %s", err))
		return
	}

	// Get the created whitelist to populate all fields
	createdWhitelist, err := r.client.GetWhitelist(identifier.ID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read created whitelist, got error: %s", err))
		return
	}

	// Update the model with the created whitelist data
	data.ID = types.StringValue(createdWhitelist.ID)
	data.Name = types.StringValue(createdWhitelist.Name)
	data.Comment = types.StringValue(createdWhitelist.Comment)
	data.Type = types.StringValue(createdWhitelist.Type)

	// Convert patterns back to types.Set
	if len(createdWhitelist.WhiteListPatterns) > 0 {
		patternsSet, diags := types.SetValueFrom(ctx, types.StringType, createdWhitelist.WhiteListPatterns)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.WhitelistPatterns = patternsSet
	}

	tflog.Debug(ctx, "Created whitelist", map[string]interface{}{
		"whitelist_id": createdWhitelist.ID,
	})

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *WhitelistResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data WhitelistResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get whitelist from API
	whitelist, err := r.client.GetWhitelist(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read whitelist, got error: %s", err))
		return
	}

	// Update the model with the retrieved data
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

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *WhitelistResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data WhitelistResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Convert whitelist_patterns from types.Set to []string
	var patternsPayload []string
	if !data.WhitelistPatterns.IsNull() && !data.WhitelistPatterns.IsUnknown() {
		resp.Diagnostics.Append(data.WhitelistPatterns.ElementsAs(ctx, &patternsPayload, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Update whitelist using hoststore API
	whitelist := hoststore.Whitelist{
		ID:                data.ID.ValueString(),
		Name:              data.Name.ValueString(),
		Comment:           data.Comment.ValueString(),
		Type:              data.Type.ValueString(),
		WhiteListPatterns: patternsPayload,
	}

	// Call the SDK method to update whitelist
	err := r.client.UpdateWhitelist(whitelist.ID, whitelist)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update whitelist, got error: %s", err))
		return
	}

	// Get the updated whitelist to populate all fields
	updatedWhitelist, err := r.client.GetWhitelist(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read updated whitelist, got error: %s", err))
		return
	}

	// Update the model with the updated whitelist data
	data.Name = types.StringValue(updatedWhitelist.Name)
	data.Comment = types.StringValue(updatedWhitelist.Comment)
	data.Type = types.StringValue(updatedWhitelist.Type)

	// Convert patterns back to types.Set
	if len(updatedWhitelist.WhiteListPatterns) > 0 {
		patternsSet, diags := types.SetValueFrom(ctx, types.StringType, updatedWhitelist.WhiteListPatterns)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.WhitelistPatterns = patternsSet
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *WhitelistResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data WhitelistResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Delete whitelist using hoststore API
	err := r.client.DeleteWhitelist(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete whitelist, got error: %s", err))
		return
	}

	tflog.Debug(ctx, "Deleted whitelist", map[string]interface{}{
		"whitelist_id": data.ID.ValueString(),
	})
}

func (r *WhitelistResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
