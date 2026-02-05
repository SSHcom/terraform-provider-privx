package provider

import (
	"context"
	"fmt"
	"terraform-provider-privx/internal/utils"

	"github.com/SSHcom/privx-sdk-go/v2/api/rolestore"
	"github.com/SSHcom/privx-sdk-go/v2/api/vault"
	"github.com/SSHcom/privx-sdk-go/v2/restapi"
	"github.com/hashicorp/terraform-plugin-framework/attr"
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
var _ resource.Resource = &SecretResource{}
var _ resource.ResourceWithImportState = &SecretResource{}

func NewSecretResource() resource.Resource {
	return &SecretResource{}
}

// SecretResource defines the resource implementation.
type SecretResource struct {
	client *vault.Vault
}

// SecretResourceModel contains PrivX secret information.
type SecretResourceModel struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	ReadRoles  types.List   `tfsdk:"read_roles"`
	WriteRoles types.List   `tfsdk:"write_roles"`
	Data       types.Map    `tfsdk:"data"`
	OwnerID    types.String `tfsdk:"owner_id"`
	Created    types.String `tfsdk:"created"`
	Updated    types.String `tfsdk:"updated"`
	UpdatedBy  types.String `tfsdk:"updated_by"`
	Author     types.String `tfsdk:"author"`
	Path       types.String `tfsdk:"path"`
}

// RoleHandleModel represents a role handle.
type RoleHandleModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func (r *SecretResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_secret"
}

func (r *SecretResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "PrivX Secret resource",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Secret ID (same as name)",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Secret name (used as identifier)",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"read_roles": schema.ListNestedAttribute{
				MarkdownDescription: "List of roles that can read this secret",
				Optional:            true,
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "Role ID",
							Required:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "Role name",
							Optional:            true,
							Computed:            true,
						},
					},
				},
			},
			"write_roles": schema.ListNestedAttribute{
				MarkdownDescription: "List of roles that can write to this secret",
				Optional:            true,
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "Role ID",
							Required:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "Role name",
							Optional:            true,
							Computed:            true,
						},
					},
				},
			},
			"data": schema.MapAttribute{
				MarkdownDescription: "Secret data as key-value pairs",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Sensitive:           true,
			},
			"owner_id": schema.StringAttribute{
				MarkdownDescription: "Owner ID of the secret",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
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
				MarkdownDescription: "ID of user who last updated the secret",
				Computed:            true,
			},
			"author": schema.StringAttribute{
				MarkdownDescription: "Author of the secret",
				Computed:            true,
			},
			"path": schema.StringAttribute{
				MarkdownDescription: "Secret path",
				Computed:            true,
			},
		},
	}
}

func (r *SecretResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*restapi.Connector)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *restapi.Connector, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = vault.New(*client)
}

func (r *SecretResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *SecretResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert read roles
	var readRoles []rolestore.RoleHandle
	if !data.ReadRoles.IsNull() && !data.ReadRoles.IsUnknown() {
		var readRoleModels []RoleHandleModel
		data.ReadRoles.ElementsAs(ctx, &readRoleModels, false)

		for _, rm := range readRoleModels {
			role := rolestore.RoleHandle{
				ID:   rm.ID.ValueString(),
				Name: rm.Name.ValueString(),
			}
			readRoles = append(readRoles, role)
		}
	}

	// Convert write roles
	var writeRoles []rolestore.RoleHandle
	if !data.WriteRoles.IsNull() && !data.WriteRoles.IsUnknown() {
		var writeRoleModels []RoleHandleModel
		data.WriteRoles.ElementsAs(ctx, &writeRoleModels, false)

		for _, rm := range writeRoleModels {
			role := rolestore.RoleHandle{
				ID:   rm.ID.ValueString(),
				Name: rm.Name.ValueString(),
			}
			writeRoles = append(writeRoles, role)
		}
	}

	// Convert secret data
	var secretData *map[string]interface{}
	if !data.Data.IsNull() && !data.Data.IsUnknown() {
		dataMap := make(map[string]string)
		data.Data.ElementsAs(ctx, &dataMap, false)

		// Convert to interface{} map
		interfaceMap := make(map[string]interface{})
		for k, v := range dataMap {
			interfaceMap[k] = v
		}
		secretData = &interfaceMap
	}

	secretRequest := &vault.SecretRequest{
		Name:       data.Name.ValueString(),
		ReadRoles:  readRoles,
		WriteRoles: writeRoles,
		Data:       secretData,
		OwnerID:    data.OwnerID.ValueString(),
	}

	tflog.Debug(ctx, "Creating secret", map[string]interface{}{
		"name": data.Name.ValueString(),
	})

	created, err := r.client.CreateSecret(secretRequest)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create secret, got error: %s", err))
		return
	}

	tflog.Debug(ctx, "Created secret", map[string]interface{}{
		"name": created.Name,
	})

	// Read the created secret to get all fields
	secret, err := r.client.GetSecret(data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read created secret, got error: %s", err))
		return
	}
	r.populateSecretModel(data, secret)
	tflog.Debug(ctx, "Storing secret into the state", map[string]interface{}{
		"state": fmt.Sprintf("%+v", data),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SecretResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *SecretResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	secret, err := r.client.GetSecret(data.Name.ValueString())
	if err != nil {
		tflog.Debug(ctx, "Error reading secret", map[string]interface{}{
			"name":  data.Name.ValueString(),
			"error": err.Error(),
		})

		if utils.IsPrivxNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read secret, got error: %s", err))
		return
	}

	r.populateSecretModel(data, secret)

	tflog.Debug(ctx, "Storing secret into the state", map[string]interface{}{
		"state": fmt.Sprintf("%+v", data),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SecretResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *SecretResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert read roles
	var readRoles []rolestore.RoleHandle
	if !data.ReadRoles.IsNull() && !data.ReadRoles.IsUnknown() {
		var readRoleModels []RoleHandleModel
		data.ReadRoles.ElementsAs(ctx, &readRoleModels, false)

		for _, rm := range readRoleModels {
			role := rolestore.RoleHandle{
				ID:   rm.ID.ValueString(),
				Name: rm.Name.ValueString(),
			}
			readRoles = append(readRoles, role)
		}
	}

	// Convert write roles
	var writeRoles []rolestore.RoleHandle
	if !data.WriteRoles.IsNull() && !data.WriteRoles.IsUnknown() {
		var writeRoleModels []RoleHandleModel
		data.WriteRoles.ElementsAs(ctx, &writeRoleModels, false)

		for _, rm := range writeRoleModels {
			role := rolestore.RoleHandle{
				ID:   rm.ID.ValueString(),
				Name: rm.Name.ValueString(),
			}
			writeRoles = append(writeRoles, role)
		}
	}

	// Convert secret data
	var secretData *map[string]interface{}
	if !data.Data.IsNull() && !data.Data.IsUnknown() {
		dataMap := make(map[string]string)
		data.Data.ElementsAs(ctx, &dataMap, false)

		// Convert to interface{} map
		interfaceMap := make(map[string]interface{})
		for k, v := range dataMap {
			interfaceMap[k] = v
		}
		secretData = &interfaceMap
	} else {
		// If data is null/unknown, read the current secret to preserve existing data
		currentSecret, err := r.client.GetSecret(data.Name.ValueString())
		if err == nil && currentSecret.Data != nil {
			secretData = currentSecret.Data
		}
	}

	secretRequest := &vault.SecretRequest{
		Name:       data.Name.ValueString(),
		ReadRoles:  readRoles,
		WriteRoles: writeRoles,
		Data:       secretData,
		OwnerID:    data.OwnerID.ValueString(),
	}

	tflog.Debug(ctx, "Updating secret", map[string]interface{}{
		"name": data.Name.ValueString(),
	})

	err := r.client.UpdateSecret(data.Name.ValueString(), secretRequest)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update secret, got error: %s", err))
		return
	}

	// Read the updated secret to get all fields
	secret, err := r.client.GetSecret(data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read updated secret, got error: %s", err))
		return
	}

	r.populateSecretModel(data, secret)

	tflog.Debug(ctx, "Storing updated secret into the state", map[string]interface{}{
		"state": fmt.Sprintf("%+v", data),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SecretResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *SecretResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting secret", map[string]interface{}{
		"name": data.Name.ValueString(),
	})

	err := r.client.DeleteSecret(data.Name.ValueString())
	if err != nil {
		if utils.IsPrivxNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete secret, got error: %s", err))
		return
	}

	tflog.Debug(ctx, "Deleted secret", map[string]interface{}{
		"name": data.Name.ValueString(),
	})
}

func (r *SecretResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(
		resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...,
	)
	resp.Diagnostics.Append(
		resp.State.SetAttribute(ctx, path.Root("name"), req.ID)...,
	)
}

// populateSecretModel populates the Terraform model from the API response.
func (r *SecretResource) populateSecretModel(data *SecretResourceModel, secret *vault.Secret) {
	data.ID = types.StringValue(secret.Name)
	data.Name = types.StringValue(secret.Name)
	data.OwnerID = types.StringValue(secret.OwnerID)
	data.Created = types.StringValue(secret.Created.Format("2006-01-02T15:04:05Z"))
	data.Updated = types.StringValue(secret.Updated.Format("2006-01-02T15:04:05Z"))
	data.UpdatedBy = types.StringValue(secret.UpdatedBy)
	data.Author = types.StringValue(secret.Author)
	data.Path = types.StringValue(secret.Path)

	// Convert read roles
	readRoleValues := make([]attr.Value, len(secret.ReadRoles))
	for i, role := range secret.ReadRoles {
		roleAttrs := map[string]attr.Value{
			"id":   types.StringValue(role.ID),
			"name": types.StringValue(role.Name),
		}
		readRoleValues[i] = types.ObjectValueMust(map[string]attr.Type{
			"id":   types.StringType,
			"name": types.StringType,
		}, roleAttrs)
	}
	data.ReadRoles = types.ListValueMust(types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"id":   types.StringType,
			"name": types.StringType,
		},
	}, readRoleValues)

	// Convert write roles
	writeRoleValues := make([]attr.Value, len(secret.WriteRoles))
	for i, role := range secret.WriteRoles {
		roleAttrs := map[string]attr.Value{
			"id":   types.StringValue(role.ID),
			"name": types.StringValue(role.Name),
		}
		writeRoleValues[i] = types.ObjectValueMust(map[string]attr.Type{
			"id":   types.StringType,
			"name": types.StringType,
		}, roleAttrs)
	}
	data.WriteRoles = types.ListValueMust(types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"id":   types.StringType,
			"name": types.StringType,
		},
	}, writeRoleValues)

	// Convert secret data
	if secret.Data != nil {
		dataMap := make(map[string]attr.Value)
		for k, v := range *secret.Data {
			// Convert interface{} to string
			if str, ok := v.(string); ok {
				dataMap[k] = types.StringValue(str)
			} else {
				dataMap[k] = types.StringValue(fmt.Sprintf("%v", v))
			}
		}
		data.Data = types.MapValueMust(types.StringType, dataMap)
	} else {
		data.Data = types.MapNull(types.StringType)
	}
}

// Helper function to check if a string contains a substring (case-insensitive)
/*func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			func() bool {
				for i := 0; i <= len(s)-len(substr); i++ {
					if s[i:i+len(substr)] == substr {
						return true
					}
				}
				return false
			}())))
}*/
