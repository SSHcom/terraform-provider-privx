package provider

import (
	"context"
	"fmt"

	"github.com/SSHcom/privx-sdk-go/v2/api/workflow"
	"github.com/SSHcom/privx-sdk-go/v2/restapi"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &WorkflowDataSource{}

func NewWorkflowDataSource() datasource.DataSource {
	return &WorkflowDataSource{}
}

// WorkflowDataSource defines the data source implementation.
type WorkflowDataSource struct {
	client *workflow.WorkflowEngine
}

type WorkflowRoleModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type WorkflowApproverModel struct {
	ID   types.String       `tfsdk:"id"`
	Role *WorkflowRoleModel `tfsdk:"role"`
}

type WorkflowStepModel struct {
	ID        types.String            `tfsdk:"id"`
	Name      types.String            `tfsdk:"name"`
	Match     types.String            `tfsdk:"match"`
	Approvers []WorkflowApproverModel `tfsdk:"approvers"`
}

// WorkflowDataSourceModel describes the data source data model.
type WorkflowDataSourceModel struct {
	ID                        types.String `tfsdk:"id"`
	Author                    types.String `tfsdk:"author"`
	Created                   types.String `tfsdk:"created"`
	Updated                   types.String `tfsdk:"updated"`
	UpdatedBy                 types.String `tfsdk:"updated_by"`
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

func (d *WorkflowDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workflow"
}

func (d *WorkflowDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Workflow data source",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Workflow ID",
				Computed:            true,
			},
			"created": schema.StringAttribute{
				MarkdownDescription: "When the object was created",
				Computed:            true,
			},
			"updated": schema.StringAttribute{
				MarkdownDescription: "When the object was created",
				Computed:            true,
			},
			"updated_by": schema.StringAttribute{
				MarkdownDescription: "ID of the user who updated the object",
				Computed:            true,
			},
			"author": schema.StringAttribute{
				MarkdownDescription: "ID of the user who originally authored the object",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the workflow",
				Required:            true,
			},
			"grant_types": schema.ListAttribute{
				MarkdownDescription: "List of role granting types. Is the role granted permanently, or is the grant time restricted, or a floating window.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"max_time_restricted_duration": schema.Int64Attribute{
				MarkdownDescription: "Maximum time in days where duration between start-date and end-date of role request must not exceeded this duration.",
				Computed:            true,
			},
			"max_floating_duration": schema.Int64Attribute{
				MarkdownDescription: "Time in hours how long the grant should not exceed after initial connection.",
				Computed:            true,
			},
			"max_active_requests": schema.Int64Attribute{
				MarkdownDescription: "Maximum number of concurrent open requests a user can have per target role.",
				Computed:            true,
			},
			"target_roles": schema.SetNestedAttribute{
				MarkdownDescription: "List of target roles for the workflow",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "Unique identifier of the role",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "Name of the role",
							Computed:            true,
						},
					},
				},
			},
			"requester_roles": schema.SetNestedAttribute{
				MarkdownDescription: "List of requester roles for the workflow",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "Unique identifier of the role",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "Name of the role",
							Computed:            true,
						},
					},
				},
			},
			"action": schema.StringAttribute{
				MarkdownDescription: "Does the workflow GRANT or REMOVE the user from the role. Workflow engine needs to check that the requested action matches allowed actions defined in the template.",
				Computed:            true,
			},
			"can_bypass_revoke_workflow": schema.BoolAttribute{
				MarkdownDescription: "A flag used to determine if approvers can bypass the revoke workflow to revoke a role.",
				Computed:            true,
			},
			"comment": schema.StringAttribute{
				MarkdownDescription: "optional human readable description",
				Optional:            true,
			},
			"steps": schema.SetNestedAttribute{
				MarkdownDescription: "List of workflow steps",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "Unique identifier of the step",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "Name of the step",
							Computed:            true,
						},
						"match": schema.StringAttribute{
							MarkdownDescription: "Match condition for the step (e.g., ANY, ALL)",
							Computed:            true,
						},
						"approvers": schema.SetNestedAttribute{
							MarkdownDescription: "List of approvers for the step",
							Computed:            true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"id": schema.StringAttribute{
										MarkdownDescription: "Unique identifier of the approver",
										Computed:            true,
									},
									"role": schema.SingleNestedAttribute{
										MarkdownDescription: "Role assigned to the approver",
										Computed:            true,
										Attributes: map[string]schema.Attribute{
											"id": schema.StringAttribute{
												MarkdownDescription: "Unique identifier of the role",
												Computed:            true,
											},
											"name": schema.StringAttribute{
												MarkdownDescription: "Name of the role",
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
				Computed:            true,
			},
		},
	}
}

func (d *WorkflowDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
	tflog.Debug(ctx, "Creating workflow", map[string]interface{}{
		"connector : ": fmt.Sprintf("%+v", *connector),
	})

	d.client = workflow.New(*connector)
}

func (d *WorkflowDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data WorkflowDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.Name.IsNull() {
		resp.Diagnostics.AddError("Configuration Error", "Name cannot be null ")
		return
	}

	searchResult, err := d.client.GetWorkflows()
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read workflows, got error: %s", err))
		return
	}

	var workflow workflow.Workflow
	for _, result := range searchResult.Items {
		if !data.Name.IsNull() {
			if result.Name == data.Name.ValueString() {
				workflow = result
				break
			}
		}
	}

	data.ID = types.StringValue(workflow.ID)
	data.Author = types.StringValue(workflow.Author)
	data.Created = types.StringValue(workflow.Created)
	data.Updated = types.StringValue(workflow.Updated)
	data.UpdatedBy = types.StringValue(workflow.UpdatedBy)
	data.Name = types.StringValue(workflow.Name)

	grantTypes, diags := types.ListValueFrom(ctx, types.StringType, workflow.GrantTypes)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	data.GrantTypes = grantTypes

	data.MaxTimeRestrictedDuration = types.Int64Value(workflow.MaxTimeRestrictedDuration)
	data.MaxFloatingDuration = types.Int64Value(workflow.MaxFloatingDuration)
	data.MaxActiveRequests = types.Int64Value(workflow.MaxActiveRequests)

	var r_roleModels []WorkflowRoleModel
	for _, r := range workflow.RequestorRoles {
		r_roleModels = append(r_roleModels, WorkflowRoleModel{
			ID:   types.StringValue(r.ID),
			Name: types.StringValue(r.Name),
		})
	}

	requesterRoles, diags := types.SetValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"id":   types.StringType,
			"name": types.StringType,
		},
	}, r_roleModels)

	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	data.RequesterRoles = requesterRoles

	var t_roleModels []WorkflowRoleModel
	for _, r := range workflow.TargetRoles {
		t_roleModels = append(t_roleModels, WorkflowRoleModel{
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

	data.Action = types.StringValue(workflow.Action)
	data.CanBypassRevokeWF = types.BoolValue(workflow.CanBypassRevokeWF)

	var stepModels []WorkflowStepModel

	for _, s := range workflow.Steps {
		var approverModels []WorkflowApproverModel
		for _, a := range s.Approvers {
			approverModels = append(approverModels, WorkflowApproverModel{
				ID: types.StringValue(a.ID),
				Role: &WorkflowRoleModel{
					ID:   types.StringValue(a.Role.ID),
					Name: types.StringValue(a.Role.Name),
				},
			})
		}

		stepModels = append(stepModels, WorkflowStepModel{
			ID:        types.StringValue(s.ID),
			Name:      types.StringValue(s.Name),
			Match:     types.StringValue(s.Match),
			Approvers: approverModels,
		})
	}

	stepsSet, diags := types.SetValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"id":    types.StringType,
			"name":  types.StringType,
			"match": types.StringType,
			"approvers": types.SetType{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"id": types.StringType,
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

	data.RequiresJustification = types.BoolValue(workflow.RequiresJustification)

	tflog.Debug(ctx, "Storing role type into the state", map[string]interface{}{
		"createNewState": fmt.Sprintf("%+v", data),
	})
	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
