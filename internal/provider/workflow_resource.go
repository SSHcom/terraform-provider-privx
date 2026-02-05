package provider

import (
	"context"
	"fmt"

	"github.com/SSHcom/privx-sdk-go/v2/api/workflow"
	"github.com/SSHcom/privx-sdk-go/v2/restapi"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &WorkflowResource{}
var _ resource.ResourceWithImportState = &WorkflowResource{}

func NewWorkflowResource() resource.Resource {
	return &WorkflowResource{}
}

// WorkflowResource defines the resource implementation.
type WorkflowResource struct {
	client *workflow.WorkflowEngine
}

type WorkflowResourceRoleModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type WorkflowResourceApproverModel struct {
	Role WorkflowResourceRoleModel `tfsdk:"role"`
}

type WorkflowResourceStepModel struct {
	Name      types.String `tfsdk:"name"`
	Match     types.String `tfsdk:"match"`
	Approvers types.Set    `tfsdk:"approvers"`
}

// WorkflowResourceModel describes the resource data model.
type WorkflowResourceModel struct {
	ID                        types.String `tfsdk:"id"`
	Name                      types.String `tfsdk:"name"`
	GrantTypes                types.List   `tfsdk:"grant_types"`
	MaxTimeRestrictedDuration types.Int64  `tfsdk:"max_time_restricted_duration"`
	MaxFloatingDuration       types.Int64  `tfsdk:"max_floating_duration"`
	MaxActiveRequests         types.Int64  `tfsdk:"max_active_requests"`
	TargetRoles               types.Set    `tfsdk:"target_roles"`
	RequesterRoles            types.Set    `tfsdk:"requester_roles"`
	Action                    types.String `tfsdk:"action"`
	CanBypassRevokeWF         types.Bool   `tfsdk:"can_bypass_revoke_workflow"`
	Comment                   types.String `tfsdk:"comment"`
	Steps                     types.Set    `tfsdk:"steps"`
	RequiresJustification     types.Bool   `tfsdk:"requires_justification"`
}

func (r *WorkflowResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workflow"
}

func (r *WorkflowResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Workflow resource",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Workflow ID",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the workflow",
				Required:            true,
			},
			"grant_types": schema.ListAttribute{
				MarkdownDescription: "List of role granting types. Is the role granted permanently, or is the grant time restricted, or a floating window.",
				Required:            true,
				ElementType:         types.StringType,
			},
			"max_time_restricted_duration": schema.Int64Attribute{
				MarkdownDescription: "Maximum time in days where duration between start-date and end-date of role request must not exceeded this duration.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(0),
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
			},
			"max_floating_duration": schema.Int64Attribute{
				MarkdownDescription: "Time in hours how long the grant should not exceed after initial connection.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(0),
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
			},
			"max_active_requests": schema.Int64Attribute{
				MarkdownDescription: "Maximum number of concurrent open requests a user can have per target role.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(1),
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"target_roles": schema.SetNestedAttribute{
				MarkdownDescription: "List of target roles for the workflow",
				Required:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "Unique identifier of the role",
							Required:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "Name of the role",
							Optional:            true,
							Computed:            true,
						},
					},
				},
			},
			"requester_roles": schema.SetNestedAttribute{
				MarkdownDescription: "List of requester roles for the workflow",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "Unique identifier of the role",
							Required:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "Name of the role",
							Optional:            true,
							Computed:            true,
						},
					},
				},
			},
			"action": schema.StringAttribute{
				MarkdownDescription: "Does the workflow GRANT or REMOVE the user from the role. Workflow engine needs to check that the requested action matches allowed actions defined in the template.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("GRANT", "REMOVE", "BOTH"),
				},
			},
			"can_bypass_revoke_workflow": schema.BoolAttribute{
				MarkdownDescription: "A flag used to determine if approvers can bypass the revoke workflow to revoke a role.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"comment": schema.StringAttribute{
				MarkdownDescription: "Optional human readable description",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"steps": schema.SetNestedAttribute{
				MarkdownDescription: "List of workflow steps",
				Required:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							MarkdownDescription: "Name of the step",
							Required:            true,
						},
						"match": schema.StringAttribute{
							MarkdownDescription: "Match condition for the step (e.g., ANY, ALL)",
							Required:            true,
							Validators: []validator.String{
								stringvalidator.OneOf("ANY", "ALL"),
							},
						},
						"approvers": schema.SetNestedAttribute{
							MarkdownDescription: "List of approvers for the step",
							Required:            true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"role": schema.SingleNestedAttribute{
										MarkdownDescription: "Role assigned to the approver",
										Required:            true,
										Attributes: map[string]schema.Attribute{
											"id": schema.StringAttribute{
												MarkdownDescription: "Unique identifier of the role",
												Required:            true,
											},
											"name": schema.StringAttribute{
												MarkdownDescription: "Name of the role",
												Optional:            true,
												Computed:            true,
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"requires_justification": schema.BoolAttribute{
				MarkdownDescription: "A flag used to determine if requesters can bypass the justification on role requests.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
		},
	}
}

func (r *WorkflowResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	tflog.Debug(ctx, "Creating workflow", map[string]interface{}{
		"connector : ": fmt.Sprintf("%+v", *connector),
	})

	r.client = workflow.New(*connector)
}

func (r *WorkflowResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data WorkflowResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Clear read-only fields to prevent them from being included in the update payload
	data.ID = types.StringNull()

	tflog.Debug(ctx, "Loaded workflow type data", map[string]interface{}{
		"data": fmt.Sprintf("%+v", data),
	})

	// Convert grant types
	var grantTypesPayload []string
	if len(data.GrantTypes.Elements()) > 0 {
		resp.Diagnostics.Append(data.GrantTypes.ElementsAs(ctx, &grantTypesPayload, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Convert target roles
	var targetRolesModels []WorkflowResourceRoleModel
	resp.Diagnostics.Append(data.TargetRoles.ElementsAs(ctx, &targetRolesModels, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var targetRolesPayload []workflow.WorkflowRole
	for _, role := range targetRolesModels {
		targetRolesPayload = append(targetRolesPayload, workflow.WorkflowRole{
			ID:   role.ID.ValueString(),
			Name: role.Name.ValueString(),
		})
	}

	// Convert requester roles
	var requesterRolesModels []WorkflowResourceRoleModel
	resp.Diagnostics.Append(data.RequesterRoles.ElementsAs(ctx, &requesterRolesModels, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var requesterRolesPayload []workflow.WorkflowRole
	for _, role := range requesterRolesModels {
		requesterRolesPayload = append(requesterRolesPayload, workflow.WorkflowRole{
			ID:   role.ID.ValueString(),
			Name: role.Name.ValueString(),
		})
	}

	// Convert steps
	var stepsModels []WorkflowResourceStepModel
	resp.Diagnostics.Append(data.Steps.ElementsAs(ctx, &stepsModels, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var stepsPayload []workflow.WorkflowStep

	for _, step := range stepsModels {
		// Decode approvers (types.Set -> []WorkflowResourceApproverModel)
		var approverModels []WorkflowResourceApproverModel
		resp.Diagnostics.Append(step.Approvers.ElementsAs(ctx, &approverModels, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		var approversPayload []workflow.WorkflowStepApprover
		for _, approver := range approverModels {
			approversPayload = append(approversPayload, workflow.WorkflowStepApprover{
				Role: workflow.WorkflowRole{
					ID:   approver.Role.ID.ValueString(),
					Name: approver.Role.Name.ValueString(),
				},
			})
		}

		stepsPayload = append(stepsPayload, workflow.WorkflowStep{
			Name:      step.Name.ValueString(),
			Match:     step.Match.ValueString(),
			Approvers: approversPayload,
		})
	}

	workflowPayload := workflow.Workflow{
		Name:                      data.Name.ValueString(),
		Comment:                   data.Comment.ValueString(),
		GrantTypes:                grantTypesPayload,
		MaxTimeRestrictedDuration: data.MaxTimeRestrictedDuration.ValueInt64(),
		MaxFloatingDuration:       data.MaxFloatingDuration.ValueInt64(),
		MaxActiveRequests:         data.MaxActiveRequests.ValueInt64(),
		TargetRoles:               targetRolesPayload,
		RequestorRoles:            requesterRolesPayload,
		Action:                    data.Action.ValueString(),
		CanBypassRevokeWF:         data.CanBypassRevokeWF.ValueBool(),
		Steps:                     stepsPayload,
		RequiresJustification:     data.RequiresJustification.ValueBool(),
	}

	tflog.Debug(ctx, fmt.Sprintf("workflow.Workflow model used: %+v", workflowPayload))

	workflowID, err := r.client.CreateWorkflow(&workflowPayload)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create the workflow, got error: %s", err))
		return
	}

	data.ID = types.StringValue(workflowID.ID)

	// Note: Timestamp and internal ID fields are not stored in Terraform state

	tflog.Debug(ctx, "created workflow resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *WorkflowResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *WorkflowResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	workflowData, err := r.client.GetWorkflow(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read workflow, got error: %s", err))
		return
	}

	/*fmt.Printf(
		"WORKFLOW READ BACK:\n"+
			"  id=%s\n"+
			"  comment=%q\n"+
			"  max_active_requests=%d\n"+
			"  steps=%+v\n\n",
		data.ID.ValueString(),
		workflowData.Comment,
		workflowData.MaxActiveRequests,
		workflowData.Steps,
	)*/

	data.Name = types.StringValue(workflowData.Name)
	data.Comment = types.StringValue(workflowData.Comment)
	data.Action = types.StringValue(workflowData.Action)
	data.CanBypassRevokeWF = types.BoolValue(workflowData.CanBypassRevokeWF)
	data.RequiresJustification = types.BoolValue(workflowData.RequiresJustification)
	data.MaxTimeRestrictedDuration = types.Int64Value(workflowData.MaxTimeRestrictedDuration)
	data.MaxFloatingDuration = types.Int64Value(workflowData.MaxFloatingDuration)
	data.MaxActiveRequests = types.Int64Value(workflowData.MaxActiveRequests)

	grantTypes, diags := types.ListValueFrom(ctx, types.StringType, workflowData.GrantTypes)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	data.GrantTypes = grantTypes

	var requesterRoleModels []WorkflowResourceRoleModel
	for _, role := range workflowData.RequestorRoles {
		requesterRoleModels = append(requesterRoleModels, WorkflowResourceRoleModel{
			ID:   types.StringValue(role.ID),
			Name: types.StringValue(role.Name),
		})
	}

	requesterRoles, diags := types.SetValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"id":   types.StringType,
			"name": types.StringType,
		},
	}, requesterRoleModels)

	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	data.RequesterRoles = requesterRoles

	var t_roleModels []WorkflowResourceRoleModel
	for _, r := range workflowData.TargetRoles {
		t_roleModels = append(t_roleModels, WorkflowResourceRoleModel{
			ID:   types.StringValue(r.ID),
			Name: types.StringValue(r.Name),
		})
	}

	targetRoles, diags := types.SetValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"id":   types.StringType,
			"name": types.StringType,
		},
	}, t_roleModels)

	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	data.TargetRoles = targetRoles

	var stepModels []WorkflowResourceStepModel

	approverObjectType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"role": types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"id":   types.StringType,
					"name": types.StringType,
				},
			},
		},
	}

	for _, s := range workflowData.Steps {
		// Build []WorkflowResourceApproverModel from API
		var approverModels []WorkflowResourceApproverModel
		for _, a := range s.Approvers {
			approverModels = append(approverModels, WorkflowResourceApproverModel{
				Role: WorkflowResourceRoleModel{
					ID:   types.StringValue(a.Role.ID),
					Name: types.StringValue(a.Role.Name),
				},
			})
		}

		// Convert []WorkflowResourceApproverModel -> types.Set
		approversSet, diags := types.SetValueFrom(ctx, approverObjectType, approverModels)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}

		// Put into step model (Approvers is now types.Set)
		stepModels = append(stepModels, WorkflowResourceStepModel{
			Name:      types.StringValue(s.Name),
			Match:     types.StringValue(s.Match),
			Approvers: approversSet,
		})
	}

	stepsSet, diags := types.SetValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"name":  types.StringType,
			"match": types.StringType,
			"approvers": types.SetType{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"role": types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"id":   types.StringType,
								"name": types.StringType,
							},
						},
					},
				},
			},
		},
	}, stepModels)

	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	data.Steps = stepsSet

	tflog.Debug(ctx, "Storing workflow type into the state", map[string]interface{}{
		"createNewState": fmt.Sprintf("%+v", data),
	})
	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *WorkflowResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *WorkflowResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var mar types.Int64
	diags := req.Plan.GetAttribute(ctx, path.Root("max_active_requests"), &mar)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	/*
		fmt.Printf("PLAN max_active_requests = %d (isNull=%v isUnknown=%v)\n",
			mar.ValueInt64(), mar.IsNull(), mar.IsUnknown())
	*/

	// Convert grant types
	var grantTypesPayload []string
	if len(data.GrantTypes.Elements()) > 0 {
		resp.Diagnostics.Append(data.GrantTypes.ElementsAs(ctx, &grantTypesPayload, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Convert target roles
	var targetRolesModels []WorkflowResourceRoleModel
	resp.Diagnostics.Append(data.TargetRoles.ElementsAs(ctx, &targetRolesModels, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var targetRolesPayload []workflow.WorkflowRole
	for _, role := range targetRolesModels {
		targetRolesPayload = append(targetRolesPayload, workflow.WorkflowRole{
			ID:   role.ID.ValueString(),
			Name: role.Name.ValueString(),
		})
	}

	// Convert requester roles
	var requesterRolesModels []WorkflowResourceRoleModel
	resp.Diagnostics.Append(data.RequesterRoles.ElementsAs(ctx, &requesterRolesModels, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var requesterRolesPayload []workflow.WorkflowRole
	for _, role := range requesterRolesModels {
		requesterRolesPayload = append(requesterRolesPayload, workflow.WorkflowRole{
			ID:   role.ID.ValueString(),
			Name: role.Name.ValueString(),
		})
	}

	// Convert steps
	var stepsModels []WorkflowResourceStepModel
	resp.Diagnostics.Append(data.Steps.ElementsAs(ctx, &stepsModels, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var stepsPayload []workflow.WorkflowStep

	for _, step := range stepsModels {
		// Decode approvers set -> []WorkflowResourceApproverModel
		var approverModels []WorkflowResourceApproverModel
		resp.Diagnostics.Append(step.Approvers.ElementsAs(ctx, &approverModels, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		var approversPayload []workflow.WorkflowStepApprover
		for _, approver := range approverModels {
			approversPayload = append(approversPayload, workflow.WorkflowStepApprover{
				Role: workflow.WorkflowRole{
					ID:   approver.Role.ID.ValueString(),
					Name: approver.Role.Name.ValueString(),
				},
			})
		}

		stepsPayload = append(stepsPayload, workflow.WorkflowStep{
			Name:      step.Name.ValueString(),
			Match:     step.Match.ValueString(),
			Approvers: approversPayload,
		})
	}

	//fmt.Printf("PLAN stepsPayload  = %+v\n", stepsPayload)

	// Get current workflow data to include read-only fields
	currentWorkflow, err := r.client.GetWorkflow(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read current workflow, got error: %s", err))
		return
	}

	// Create update payload with ALL fields including read-only ones (like the UI does)
	workflowPayload := workflow.Workflow{
		ID:                        data.ID.ValueString(),
		Author:                    currentWorkflow.Author,
		Created:                   currentWorkflow.Created,
		Updated:                   currentWorkflow.Updated,
		UpdatedBy:                 currentWorkflow.UpdatedBy,
		Name:                      data.Name.ValueString(),
		Comment:                   data.Comment.ValueString(),
		GrantTypes:                grantTypesPayload,
		MaxTimeRestrictedDuration: data.MaxTimeRestrictedDuration.ValueInt64(),
		MaxFloatingDuration:       data.MaxFloatingDuration.ValueInt64(),
		MaxActiveRequests:         data.MaxActiveRequests.ValueInt64(),
		TargetRoles:               targetRolesPayload,
		RequestorRoles:            requesterRolesPayload,
		Action:                    data.Action.ValueString(),
		CanBypassRevokeWF:         data.CanBypassRevokeWF.ValueBool(),
		Steps:                     stepsPayload,
		RequiresJustification:     data.RequiresJustification.ValueBool(),
	}

	/*fmt.Printf(
		"\nWORKFLOW UPDATE PAYLOAD:\n"+
			"  id=%s\n"+
			"  max_active_requests=%d\n"+
			"  max_floating_duration=%d\n"+
			"  max_time_restricted_duration=%d\n"+
			"  steps=%+v\n\n",
		data.ID.ValueString(),
		data.MaxActiveRequests.ValueInt64(),
		data.MaxFloatingDuration.ValueInt64(),
		data.MaxTimeRestrictedDuration.ValueInt64(),
		stepsPayload,
	)*/

	tflog.Debug(ctx, fmt.Sprintf("workflow.Workflow model used: %+v", workflowPayload))

	err = r.client.UpdateWorkflow(
		data.ID.ValueString(),
		&workflowPayload)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update workflow, got error: %s", err))
		return
	}

	// Note: Timestamp and internal ID fields are not stored in Terraform state

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *WorkflowResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *WorkflowResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteWorkflow(data.ID.ValueString())

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete workflow, got error: %s", err))
		return
	}
}

func (r *WorkflowResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
