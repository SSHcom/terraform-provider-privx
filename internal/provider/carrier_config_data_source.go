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
var _ datasource.DataSource = &CarrierConfigDataSource{}

func NewCarrierConfigDataSource() datasource.DataSource {
	return &CarrierConfigDataSource{}
}

// CarrierConfigDataSource defines the data source implementation.
type CarrierConfigDataSource struct {
	client *authorizer.Authorizer
}

// CarrierConfigDataSourceModel describes the data source data model.
type CarrierConfigDataSourceModel struct {
	ID        types.String `tfsdk:"id"`
	SessionID types.String `tfsdk:"session_id"`
	Config    types.String `tfsdk:"config"`
}

func (d *CarrierConfigDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_carrier_config"
}

func (d *CarrierConfigDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "PrivX Carrier Configuration data source",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Carrier ID",
				Required:            true,
			},
			"session_id": schema.StringAttribute{
				MarkdownDescription: "Carrier configuration session ID",
				Computed:            true,
			},
			"config": schema.StringAttribute{
				MarkdownDescription: "Carrier configuration content",
				Computed:            true,
			},
		},
	}
}

func (d *CarrierConfigDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *CarrierConfigDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data CarrierConfigDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	carrierID := data.ID.ValueString()

	// Get carrier configuration session
	sessionResponse, err := d.client.GetCarrierConfigSessions(carrierID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read carrier config session, got error: %s", err))
		return
	}

	sessionID := sessionResponse.SessionID
	data.SessionID = types.StringValue(sessionID)

	// Create a temporary file to download the config
	tempDir, err := os.MkdirTemp("", "privx-carrier-config-")
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

	configFileName := filepath.Join(tempDir, "carrier-config.toml")

	// Download the carrier configuration
	err = d.client.DownloadCarrierConfig(carrierID, sessionID, configFileName)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to download carrier config, got error: %s", err))
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

	tflog.Debug(ctx, "Storing carrier config into the state", map[string]interface{}{
		"carrier_id":  carrierID,
		"session_id":  sessionID,
		"config_size": len(configContent),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
