package provider

import (
	"context"
	"fmt"

	"github.com/SSHcom/privx-sdk-go/v2/api/rolestore"
	"github.com/SSHcom/privx-sdk-go/v2/restapi"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &SourceDataSource{}

func NewSourceDataSource() datasource.DataSource {
	return &SourceDataSource{}
}

// SourceDataSource defines the data source implementation.
type SourceDataSource struct {
	client *rolestore.RoleStore
}

type (
	OIDCConnectionDataModel struct {
		Address           types.String `tfsdk:"address"`
		Enabled           types.Bool   `tfsdk:"enabled"`
		Issuer            types.String `tfsdk:"issuer"`
		ButtonTitle       types.String `tfsdk:"button_title"`
		ClientID          types.String `tfsdk:"client_id"`
		ClientSecret      types.String `tfsdk:"client_secret"`
		TagsAttributeName types.String `tfsdk:"tags_attribute_name"`
		ScopesSecret      types.List   `tfsdk:"additional_scopes_secret"`
	}

	EUMDataModel struct {
		SourceID          types.String `tfsdk:"source_id"`
		SourceSearchField types.String `tfsdk:"source_search_field"`
	}

	// SourceDataSourceModel describes the data source data model.
	SourceDataSourceModel struct {
		ID                  types.String             `tfsdk:"id"`
		Name                types.String             `tfsdk:"name"`
		Enabled             types.Bool               `tfsdk:"enabled"`
		TTL                 types.Int64              `tfsdk:"ttl"`
		Comment             types.String             `tfsdk:"comment"`
		Tags                types.List               `tfsdk:"tags"`
		UsernamePattern     types.List               `tfsdk:"username_pattern"`
		ExternalUserMapping []*EUMDataModel          `tfsdk:"external_user_mapping"`
		OIDCConnection      *OIDCConnectionDataModel `tfsdk:"oidc_connection"`
	}
)

func (d *SourceDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_source"
}

func (d *SourceDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Source data source",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Source ID",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Source name",
				Required:            true,
			},
			"comment": schema.StringAttribute{
				MarkdownDescription: "Source comment",
				Computed:            true,
			},
			"ttl": schema.Int64Attribute{
				MarkdownDescription: "Source ttl",
				Computed:            true,
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Source enabled",
				Computed:            true,
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Source tags",
				Computed:            true,
			},
			"username_pattern": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Source external user pattern",
				Computed:            true,
			},
			"external_user_mapping": schema.ListAttribute{
				MarkdownDescription: "Source external user mapping",
				Computed:            true,
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"source_id":           types.StringType,
						"source_search_field": types.StringType,
					}},
			},
			"oidc_connection": schema.SingleNestedAttribute{
				MarkdownDescription: "OIDC connection",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"address": schema.StringAttribute{
						MarkdownDescription: "oidc connection address",
						Computed:            true,
					},
					"enabled": schema.BoolAttribute{
						MarkdownDescription: "oidc connection enabled",
						Computed:            true,
					},
					"issuer": schema.StringAttribute{
						MarkdownDescription: "oidc connection issuer",
						Computed:            true,
					},
					"button_title": schema.StringAttribute{
						MarkdownDescription: "oidc connection title",
						Computed:            true,
					},
					"client_id": schema.StringAttribute{
						MarkdownDescription: "oidc connection client ID",
						Computed:            true,
					},
					"client_secret": schema.StringAttribute{
						MarkdownDescription: "oidc connection client Secret",
						Computed:            true,
						Sensitive:           true,
					},
					"tags_attribute_name": schema.StringAttribute{
						MarkdownDescription: "oidc connection tags attribute name",
						Computed:            true,
					},
					"additional_scopes_secret": schema.ListAttribute{
						ElementType:         types.StringType,
						MarkdownDescription: "oidc additional scopes secret",
						Computed:            true,
					},
				},
			},
		},
	}
}

func (d *SourceDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	d.client = rolestore.New(*connector)
}

func (d *SourceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SourceDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get all sources from PrivX API
	sourcesResult, err := d.client.GetSources()
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read sources, got error: %s", err))
		return
	}

	// Find the source by name
	var foundSource *rolestore.Source
	sourceName := data.Name.ValueString()
	for _, source := range sourcesResult.Items {
		if source.Name == sourceName {
			foundSource = &source
			break
		}
	}

	if foundSource == nil {
		resp.Diagnostics.AddError(
			"Source Not Found",
			fmt.Sprintf("No source found with name: %s", sourceName),
		)
		return
	}

	// Convert tags from API response to Terraform types
	tags, diags := types.ListValueFrom(ctx, types.StringType, foundSource.Tags)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Convert username pattern from API response to Terraform types
	usernamePattern, diags := types.ListValueFrom(ctx, types.StringType, foundSource.UsernamePattern)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Convert external user mapping from API response to Terraform model
	var eum []*EUMDataModel
	for _, v := range foundSource.ExternalUserMapping {
		eum = append(eum, &EUMDataModel{
			SourceID:          types.StringValue(v.SourceID),
			SourceSearchField: types.StringValue(v.SourceSearchField),
		})
	}

	// Convert OIDC additional scopes from API response to Terraform types
	scopesSecret, diags := types.ListValueFrom(ctx, types.StringType, foundSource.Connection.OIDCAdditionalScopes)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Convert OIDC connection from API response to Terraform model
	connection := &OIDCConnectionDataModel{
		Address:           types.StringValue(foundSource.Connection.Address),
		Enabled:           types.BoolValue(foundSource.Connection.OIDCEnabled),
		ButtonTitle:       types.StringValue(foundSource.Connection.OIDCButtonTitle),
		Issuer:            types.StringValue(foundSource.Connection.OIDCIssuer),
		ClientID:          types.StringValue(foundSource.Connection.OIDCClientID),
		ClientSecret:      types.StringValue(foundSource.Connection.OIDCClientSecret), // Note: PrivX may return masked value
		TagsAttributeName: types.StringValue(foundSource.Connection.OIDCTagsAttributeName),
		ScopesSecret:      scopesSecret,
	}

	// Set all computed attributes
	data.ID = types.StringValue(foundSource.ID)
	data.Name = types.StringValue(foundSource.Name)
	data.Enabled = types.BoolValue(foundSource.Enabled)
	data.TTL = types.Int64Value(int64(foundSource.TTL))
	data.Comment = types.StringValue(foundSource.Comment)
	data.Tags = tags
	data.UsernamePattern = usernamePattern
	data.ExternalUserMapping = eum
	data.OIDCConnection = connection

	tflog.Debug(ctx, fmt.Sprintf("Source data read: %+v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
