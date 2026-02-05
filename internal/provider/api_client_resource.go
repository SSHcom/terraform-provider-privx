package provider

import (
	"context"
	"fmt"
	"terraform-provider-privx/internal/utils"

	"github.com/SSHcom/privx-sdk-go/v2/api/rolestore"
	"github.com/SSHcom/privx-sdk-go/v2/api/userstore"
	"github.com/SSHcom/privx-sdk-go/v2/restapi"
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

// APIClientModel describes the resource data model.
type APIClientModel struct {
	ID                types.String    `tfsdk:"id"`
	Name              types.String    `tfsdk:"name"`
	Secret            types.String    `tfsdk:"secret"`
	OauthClientId     types.String    `tfsdk:"oauth_client_id"`
	OauthClientSecret types.String    `tfsdk:"oauth_client_secret"`
	Roles             []RolesRefModel `tfsdk:"roles"`
}

func (r *APIClientResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_client"
}

func (r *APIClientResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "API client resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "ID of the API client",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "name of the API client",
				Required:            true,
			},
			"secret": schema.StringAttribute{
				MarkdownDescription: "secret of the API client",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"oauth_client_id": schema.StringAttribute{
				MarkdownDescription: "oauth_client_id of the API client",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"oauth_client_secret": schema.StringAttribute{
				MarkdownDescription: "oauth_client_secret of the API client",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"roles": schema.SetNestedAttribute{
				MarkdownDescription: "List of roles possessed by the API client",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "Role ID",
							Required:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "Role name (computed from server).",
							Optional:            true,
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (r *APIClientResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *APIClientResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *APIClientModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var rolesPayload []rolestore.RoleHandle
	for _, roleRef := range data.Roles {
		rolesPayload = append(rolesPayload, rolestore.RoleHandle{
			ID: roleRef.ID.ValueString(),
		})
	}

	apiClientCreate := &userstore.APIClientCreate{
		Name:  data.Name.ValueString(),
		Roles: rolesPayload,
	}

	id, err := r.client.CreateAPIClient(apiClientCreate)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create API client Resource",
			"An unexpected error occurred while attempting to create the resource.\n"+
				err.Error(),
		)
		return
	}

	apiClient, err := r.client.GetAPIClient(id.ID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read api_client, got error: %s", err))
		return
	}

	// Set computed fields
	data.ID = types.StringValue(id.ID)
	data.Secret = types.StringValue(apiClient.Secret)
	data.OauthClientId = types.StringValue(apiClient.OAuthClientID)
	data.OauthClientSecret = types.StringValue(apiClient.OAuthClientSecret)

	// Build roles for state in a stable way:
	// - prefer API name if present
	// - otherwise preserve planned name if user set it
	// - otherwise keep null (NOT empty string)
	var roles []RolesRefModel
	for _, role := range apiClient.Roles {
		roleName := types.StringNull()

		if role.Name != "" {
			roleName = types.StringValue(role.Name)
		} else {
			// Preserve planned role name (if any) for same role ID
			for _, planned := range data.Roles {
				if planned.ID.ValueString() == role.ID &&
					!planned.Name.IsNull() &&
					!planned.Name.IsUnknown() &&
					planned.Name.ValueString() != "" {
					roleName = planned.Name
					break
				}
			}
		}

		roles = append(roles, RolesRefModel{
			ID:   types.StringValue(role.ID),
			Name: roleName,
		})
	}
	data.Roles = roles

	ctx = tflog.SetField(ctx, "API client name", data.Name.ValueString())
	ctx = tflog.SetField(ctx, "API client roles", data.Roles)
	tflog.Debug(ctx, "Created API client")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *APIClientResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *APIClientModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	apiClient, err := r.client.GetAPIClient(data.ID.ValueString())
	if err != nil {
		if utils.IsPrivxNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read API client, got error: %s", err))
		return
	}

	// Non-role fields
	data.Name = types.StringValue(apiClient.Name)
	data.OauthClientId = types.StringValue(apiClient.OAuthClientID)

	// Preserve secrets if API doesn't return them later
	if apiClient.Secret != "" {
		data.Secret = types.StringValue(apiClient.Secret)
	}
	if apiClient.OAuthClientSecret != "" {
		data.OauthClientSecret = types.StringValue(apiClient.OAuthClientSecret)
	}

	// Roles: prefer API name; if not available, preserve previous state's name for same ID
	var roles []RolesRefModel
	for _, role := range apiClient.Roles {
		roleName := types.StringNull()

		// 1) Prefer server-provided name
		if role.Name != "" {
			roleName = types.StringValue(role.Name)
		} else {
			// 2) Otherwise preserve whatever was in previous state for this role id (if any)
			for _, prev := range data.Roles {
				if prev.ID.ValueString() == role.ID &&
					!prev.Name.IsNull() &&
					!prev.Name.IsUnknown() &&
					prev.Name.ValueString() != "" {
					roleName = prev.Name
					break
				}
			}
		}

		roles = append(roles, RolesRefModel{
			ID:   types.StringValue(role.ID),
			Name: roleName,
		})
	}
	data.Roles = roles

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *APIClientResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *APIClientModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var rolesPayload []rolestore.RoleHandle
	for _, roleRef := range data.Roles {
		rolesPayload = append(rolesPayload, rolestore.RoleHandle{
			ID: roleRef.ID.ValueString(),
		})
	}

	apiClientPayload := userstore.APIClient{
		ID:                data.ID.ValueString(),
		Name:              data.Name.ValueString(),
		Secret:            data.Secret.ValueString(),
		OAuthClientID:     data.OauthClientId.ValueString(),
		OAuthClientSecret: data.OauthClientSecret.ValueString(),
		Roles:             rolesPayload,
	}

	if err := r.client.UpdateAPIClient(data.ID.ValueString(), &apiClientPayload); err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create API client Resource",
			"An unexpected error occurred while attempting to create the resource.\n"+
				err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *APIClientResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *APIClientModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteAPIClient(data.ID.ValueString()); err != nil {
		if utils.IsPrivxNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete API client, got error: %s", err))
		return
	}

}

func (r *APIClientResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
