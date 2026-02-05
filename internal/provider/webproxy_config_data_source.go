package provider

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"

	"github.com/SSHcom/privx-sdk-go/v2/api/authorizer"
	"github.com/SSHcom/privx-sdk-go/v2/restapi"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &WebProxyConfigDataSource{}

func NewWebProxyConfigDataSource() datasource.DataSource {
	return &WebProxyConfigDataSource{}
}

// WebProxyConfigDataSource defines the data source implementation.
type WebProxyConfigDataSource struct {
	client *authorizer.Authorizer
}

// WebProxyConfigDataSourceModel describes the data source data model.
type WebProxyConfigDataSourceModel struct {
	ID        types.String `tfsdk:"id"`
	SessionID types.String `tfsdk:"session_id"`
	Config    types.String `tfsdk:"config"`
}

func (d *WebProxyConfigDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_webproxy_config"
}

func (d *WebProxyConfigDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "PrivX WebProxy Configuration data source",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "WebProxy ID",
				Required:            true,
			},
			"session_id": schema.StringAttribute{
				MarkdownDescription: "WebProxy configuration session ID",
				Computed:            true,
			},
			"config": schema.StringAttribute{
				MarkdownDescription: "WebProxy configuration content",
				Computed:            true,
			},
		},
	}
}

func (d *WebProxyConfigDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	tflog.Debug(ctx, "Creating authorizer client", map[string]interface{}{
		"connector": fmt.Sprintf("%+v", *connector),
	})

	d.client = authorizer.New(*connector)
}

func (d *WebProxyConfigDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data WebProxyConfigDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	webProxyID := data.ID.ValueString()

	// Get webproxy configuration session
	sessionResponse, err := d.client.GetWebProxyConfigSessions(webProxyID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read webproxy config session, got error: %s", err))
		return
	}

	sessionID := sessionResponse.SessionID
	data.SessionID = types.StringValue(sessionID)

	// Create a temporary file to download the config
	tempDir, err := os.MkdirTemp("", "privx-webproxy-config-")
	if err != nil {
		resp.Diagnostics.AddError("File System Error", fmt.Sprintf("Unable to create temporary directory, got error: %s", err))
		return
	}
	defer func() {
		if removeErr := os.RemoveAll(tempDir); removeErr != nil {
			tflog.Warn(ctx, "Failed to clean up temporary directory", map[string]interface{}{
				"temp_dir": tempDir,
				"error":    removeErr.Error(),
			})
		}
	}()

	configFileName := filepath.Join(tempDir, "webproxy-config.toml")

	// Download the webproxy configuration
	err = d.client.DownloadWebProxyConfig(webProxyID, sessionID, configFileName)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to download webproxy config, got error: %s", err))
		return
	}

	// Read the downloaded config file
	// #nosec G304 -- configFileName is created in a fresh temp dir controlled by us, not user input.
	configContent, err := os.ReadFile(configFileName)
	if err != nil {
		resp.Diagnostics.AddError("File System Error", fmt.Sprintf("Unable to read downloaded config file, got error: %s", err))
		return
	}

	// Store the config content as base64 encoded string (TOML configuration file)
	configBase64 := base64.StdEncoding.EncodeToString(configContent)
	data.Config = types.StringValue(configBase64)

	tflog.Debug(ctx, "Storing webproxy config into the state", map[string]interface{}{
		"webproxy_id": webProxyID,
		"session_id":  sessionID,
		"config_size": len(configContent),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
