package provider

import (
	"context"
	"fmt"

	"github.com/SSHcom/privx-sdk-go/v2/api/userstore"
	"github.com/SSHcom/privx-sdk-go/v2/restapi"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &APIClientDataSource{}

func NewAPIClientDataSource() datasource.DataSource {
	return &APIClientDataSource{}
}

// APIClientDataSource defines the data source implementation.
type APIClientDataSource struct {
	client *userstore.UserStore
}

// APIClientDataSourceModel describes the data source data model.
type APIClientDataSourceModel APIClientResourceModel

func (d *APIClientDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_client"
}

func (d *APIClientDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "API Client data source",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "API Client ID",
				Optional:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the API client",
				Optional:            true,
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

func (d *APIClientDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	tflog.Debug(ctx, "Creating userstore", map[string]interface{}{
		"connector": fmt.Sprintf("%+v", *connector),
	})

	d.client = userstore.New(*connector)
}

func (d *APIClientDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data APIClientDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiClient *userstore.APIClient
	var err error

	// If ID is provided, get by ID
	if !data.ID.IsNull() && !data.ID.IsUnknown() {
		apiClient, err = d.client.GetAPIClient(data.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read API client by ID, got error: %s", err))
			return
		}
	} else if !data.Name.IsNull() && !data.Name.IsUnknown() {
		// If name is provided, search by name
		search := &userstore.APIClientSearch{
			Keywords: data.Name.ValueString(),
		}

		clients, err := d.client.SearchAPIClients(search)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to search API clients, got error: %s", err))
			return
		}

		if len(clients.Items) == 0 {
			resp.Diagnostics.AddError("Search Error", fmt.Sprintf("Could not find API client with name: %s", data.Name.ValueString()))
			return
		}

		// Find exact match by name
		var foundClient *userstore.APIClient
		for _, client := range clients.Items {
			if client.Name == data.Name.ValueString() {
				foundClient = &client
				break
			}
		}

		if foundClient == nil {
			resp.Diagnostics.AddError("Search Error", fmt.Sprintf("Could not find exact match for API client name: %s", data.Name.ValueString()))
			return
		}

		apiClient = foundClient
	} else {
		resp.Diagnostics.AddError("Configuration Error", "Either 'id' or 'name' must be specified")
		return
	}

	data.ID = types.StringValue(apiClient.ID)
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
