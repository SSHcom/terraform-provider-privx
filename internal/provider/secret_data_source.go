package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/SSHcom/privx-sdk-go/v2/api/vault"
	"github.com/SSHcom/privx-sdk-go/v2/restapi"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &SecretDataSource{}

func NewSecretDataSource() datasource.DataSource {
	return &SecretDataSource{}
}

// SecretDataSource defines the data source implementation.
type SecretDataSource struct {
	client *vault.Vault
}

// SecretDataSourceModel describes the data source data model.
type SecretDataSourceModel struct {
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

func (d *SecretDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_secret"
}

func (d *SecretDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "PrivX Secret data source",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "Secret name",
				Required:            true,
			},
			"read_roles": schema.ListNestedAttribute{
				MarkdownDescription: "List of roles that can read this secret",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "Role ID",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "Role name",
							Computed:            true,
						},
					},
				},
			},
			"write_roles": schema.ListNestedAttribute{
				MarkdownDescription: "List of roles that can write to this secret",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "Role ID",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "Role name",
							Computed:            true,
						},
					},
				},
			},
			"data": schema.MapAttribute{
				MarkdownDescription: "Secret data as key-value pairs",
				ElementType:         types.StringType,
				Computed:            true,
				Sensitive:           true,
			},
			"owner_id": schema.StringAttribute{
				MarkdownDescription: "Owner ID of the secret",
				Computed:            true,
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

func (d *SecretDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*restapi.Connector)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *restapi.Connector, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = vault.New(*client)
}

func (d *SecretDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SecretDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Secret data source read started", map[string]interface{}{
		"name": data.Name.ValueString(),
	})

	secret, err := d.client.GetSecret(data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read secret, got error: %s", err))
		return
	}

	// Populate the data source model
	d.populateSecretDataSourceModel(&data, secret)

	tflog.Debug(ctx, "Storing secret into the state", map[string]interface{}{
		"name": data.Name.ValueString(),
		"path": data.Path.ValueString(),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// populateSecretDataSourceModel populates the Terraform data source model from the API response.
func (d *SecretDataSource) populateSecretDataSourceModel(data *SecretDataSourceModel, secret *vault.Secret) {
	data.Name = types.StringValue(secret.Name)
	data.OwnerID = types.StringValue(secret.OwnerID)
	data.Created = types.StringValue(secret.Created.UTC().Format(time.RFC3339))
	data.Updated = types.StringValue(secret.Updated.UTC().Format(time.RFC3339))
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
