package provider

import (
	"context"
	"fmt"

	"github.com/SSHcom/privx-sdk-go/v2/api/rolestore"
	"github.com/SSHcom/privx-sdk-go/v2/api/userstore"
	"github.com/SSHcom/privx-sdk-go/v2/restapi"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &APIClientResource{}
var _ resource.ResourceWithImportState = &APIClientResource{}

func NewAPIClientResource() resource.Resource {
	return &APIClientResource{}
}

// APIClientResource defines the resource implementation.
type APIClientResource struct {
	client *userstore.UserStore
}

// APIClientResourceModel contains PrivX API client information.
type APIClientResourceModel struct {
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	Secret            types.String `tfsdk:"secret"`
	Created           types.String `tfsdk:"created"`
	Updated           types.String `tfsdk:"updated"`
	UpdatedBy         types.String `tfsdk:"updated_by"`
	Author            types.String `tfsdk:"author"`
	Roles             types.List   `tfsdk:"roles"`
	OAuthClientID     types.String `tfsdk:"oauth_client_id"`
	OAuthClientSecret types.String `tfsdk:"oauth_client_secret"`
}

func (r *APIClientResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_client"
}

func (r *APIClientResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "API Client resource",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "API Client ID",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the API client",
				Required:            true,
			},
			"secret": schema.StringAttribute{
				MarkdownDescription: "API Client secret",
				Computed:            true,
				Sensitive:           true,
			},
			"created": schema.StringAttribute{
				MarkdownDescription: "Creation timestamp",
				Computed:            true,
			},
			"updated": schema.StringAttribute{
				MarkdownDescription: "Last update timestamp",
				Computed:            true,
			},
			"updated_by": schema.StringAttribute{
				MarkdownDescription: "User who last updated the API client",
				Computed:            true,
			},
			"author": schema.StringAttribute{
				MarkdownDescription: "User who created the API client",
				Computed:            true,
			},
			"roles": schema.ListNestedAttribute{
				MarkdownDescription: "Roles assigned to the API client",
				Required:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "Role ID",
							Required:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "Role name",
							Optional:            true,
						},
					},
				},
			},
			"oauth_client_id": schema.StringAttribute{
				MarkdownDescription: "OAuth client ID",
				Computed:            true,
			},
			"oauth_client_secret": schema.StringAttribute{
				MarkdownDescription: "OAuth client secret",
				Computed:            true,
				Sensitive:           true,
			},
		},
	}
}

func (r *APIClientResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
		"connector": fmt.Sprintf("%+v", *connector),
	})

	r.client = userstore.New(*connector)
}

func (r *APIClientResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data APIClientResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Loaded API client data", map[string]interface{}{
		"data": fmt.Sprintf("%+v", data),
	})

	// Convert roles from Terraform model to SDK model
	var rolesPayload []rolestore.RoleHandle
	if len(data.Roles.Elements()) > 0 {
		rolesElements := make([]types.Object, 0, len(data.Roles.Elements()))
		resp.Diagnostics.Append(data.Roles.ElementsAs(ctx, &rolesElements, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		for _, roleElement := range rolesElements {
			roleAttrs := roleElement.Attributes()
			roleHandle := rolestore.RoleHandle{
				ID: roleAttrs["id"].(types.String).ValueString(),
			}
			if name, ok := roleAttrs["name"]; ok && !name.(types.String).IsNull() {
				roleHandle.Name = name.(types.String).ValueString()
			}
			rolesPayload = append(rolesPayload, roleHandle)
		}
	}

	apiClientCreate := userstore.APIClientCreate{
		Name:  data.Name.ValueString(),
		Roles: rolesPayload,
	}

	tflog.Debug(ctx, fmt.Sprintf("userstore.APIClientCreate model used: %+v", apiClientCreate))

	clientID, err := r.client.CreateAPIClient(&apiClientCreate)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create API client, got error: %s", err))
		return
	}

	data.ID = types.StringValue(clientID.ID)

	// Read the created API client to get all fields
	createdClient, err := r.client.GetAPIClient(clientID.ID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read created API client, got error: %s", err))
		return
	}

	// Update model with response data
	data.Secret = types.StringValue(createdClient.Secret)
	data.Created = types.StringValue(createdClient.Created)
	data.Updated = types.StringValue(createdClient.Updated)
	data.UpdatedBy = types.StringValue(createdClient.UpdatedBy)
	data.Author = types.StringValue(createdClient.Author)
	data.OAuthClientID = types.StringValue(createdClient.OAuthClientID)
	data.OAuthClientSecret = types.StringValue(createdClient.OAuthClientSecret)

	// Convert roles back to Terraform model
	roleObjects := make([]attr.Value, len(createdClient.Roles))
	for i, role := range createdClient.Roles {
		roleAttrs := map[string]attr.Value{
			"id":   types.StringValue(role.ID),
			"name": types.StringValue(role.Name),
		}
		roleObj, diags := types.ObjectValue(map[string]attr.Type{
			"id":   types.StringType,
			"name": types.StringType,
		}, roleAttrs)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		roleObjects[i] = roleObj
	}

	rolesList, diags := types.ListValue(types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"id":   types.StringType,
			"name": types.StringType,
		},
	}, roleObjects)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	data.Roles = rolesList

	tflog.Debug(ctx, "Created API client resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *APIClientResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *APIClientResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiClient, err := r.client.GetAPIClient(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read API client, got error: %s", err))
		return
	}

	data.Name = types.StringValue(apiClient.Name)
	data.Secret = types.StringValue(apiClient.Secret)
	data.Created = types.StringValue(apiClient.Created)
	data.Updated = types.StringValue(apiClient.Updated)
	data.UpdatedBy = types.StringValue(apiClient.UpdatedBy)
	data.Author = types.StringValue(apiClient.Author)
	data.OAuthClientID = types.StringValue(apiClient.OAuthClientID)
	data.OAuthClientSecret = types.StringValue(apiClient.OAuthClientSecret)

	// Convert roles to Terraform model
	roleObjects := make([]attr.Value, len(apiClient.Roles))
	for i, role := range apiClient.Roles {
		roleAttrs := map[string]attr.Value{
			"id":   types.StringValue(role.ID),
			"name": types.StringValue(role.Name),
		}
		roleObj, diags := types.ObjectValue(map[string]attr.Type{
			"id":   types.StringType,
			"name": types.StringType,
		}, roleAttrs)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		roleObjects[i] = roleObj
	}

	rolesList, diags := types.ListValue(types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"id":   types.StringType,
			"name": types.StringType,
		},
	}, roleObjects)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	data.Roles = rolesList

	tflog.Debug(ctx, "Storing API client into the state", map[string]interface{}{
		"readNewState": fmt.Sprintf("%+v", data),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *APIClientResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *APIClientResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert roles from Terraform model to SDK model
	var rolesPayload []rolestore.RoleHandle
	if len(data.Roles.Elements()) > 0 {
		rolesElements := make([]types.Object, 0, len(data.Roles.Elements()))
		resp.Diagnostics.Append(data.Roles.ElementsAs(ctx, &rolesElements, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		for _, roleElement := range rolesElements {
			roleAttrs := roleElement.Attributes()
			roleHandle := rolestore.RoleHandle{
				ID: roleAttrs["id"].(types.String).ValueString(),
			}
			if name, ok := roleAttrs["name"]; ok && !name.(types.String).IsNull() {
				roleHandle.Name = name.(types.String).ValueString()
			}
			rolesPayload = append(rolesPayload, roleHandle)
		}
	}

	apiClient := userstore.APIClient{
		ID:                data.ID.ValueString(),
		Name:              data.Name.ValueString(),
		Secret:            data.Secret.ValueString(),
		Created:           data.Created.ValueString(),
		Updated:           data.Updated.ValueString(),
		UpdatedBy:         data.UpdatedBy.ValueString(),
		Author:            data.Author.ValueString(),
		Roles:             rolesPayload,
		OAuthClientID:     data.OAuthClientID.ValueString(),
		OAuthClientSecret: data.OAuthClientSecret.ValueString(),
	}

	tflog.Debug(ctx, fmt.Sprintf("userstore.APIClient model used: %+v", apiClient))

	err := r.client.UpdateAPIClient(data.ID.ValueString(), &apiClient)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update API client, got error: %s", err))
		return
	}

	// Read the updated API client to get all current field values
	updatedClient, err := r.client.GetAPIClient(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read updated API client, got error: %s", err))
		return
	}

	// Update model with response data
	data.Name = types.StringValue(updatedClient.Name)
	data.Secret = types.StringValue(updatedClient.Secret)
	data.Created = types.StringValue(updatedClient.Created)
	data.Updated = types.StringValue(updatedClient.Updated)
	data.UpdatedBy = types.StringValue(updatedClient.UpdatedBy)
	data.Author = types.StringValue(updatedClient.Author)
	data.OAuthClientID = types.StringValue(updatedClient.OAuthClientID)
	data.OAuthClientSecret = types.StringValue(updatedClient.OAuthClientSecret)

	// Convert roles back to Terraform model
	roleObjects := make([]attr.Value, len(updatedClient.Roles))
	for i, role := range updatedClient.Roles {
		roleAttrs := map[string]attr.Value{
			"id":   types.StringValue(role.ID),
			"name": types.StringValue(role.Name),
		}
		roleObj, diags := types.ObjectValue(map[string]attr.Type{
			"id":   types.StringType,
			"name": types.StringType,
		}, roleAttrs)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		roleObjects[i] = roleObj
	}

	rolesList, diags := types.ListValue(types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"id":   types.StringType,
			"name": types.StringType,
		},
	}, roleObjects)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	data.Roles = rolesList

	tflog.Debug(ctx, "Updated API client resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *APIClientResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *APIClientResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteAPIClient(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete API client, got error: %s", err))
		return
	}
}

func (r *APIClientResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
