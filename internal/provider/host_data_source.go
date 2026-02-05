package provider

import (
	"context"
	"fmt"

	"github.com/SSHcom/privx-sdk-go/v2/api/hoststore"
	"github.com/SSHcom/privx-sdk-go/v2/restapi"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &HostDataSource{}

func NewHostDataSource() datasource.DataSource {
	return &HostDataSource{}
}

// HostDataSource defines the data source implementation.
type HostDataSource struct {
	client *hoststore.HostStore
}

// HostDataSourceModel describes the data source data model.
type HostDataSourceModel struct {
	ID                      types.String `tfsdk:"id"`
	CommonName              types.String `tfsdk:"common_name"`
	Addresses               types.List   `tfsdk:"addresses"`
	ExternalID              types.String `tfsdk:"external_id"`
	InstanceID              types.String `tfsdk:"instance_id"`
	SourceID                types.String `tfsdk:"source_id"`
	AccessGroupID           types.String `tfsdk:"access_group_id"`
	CloudProvider           types.String `tfsdk:"cloud_provider"`
	CloudProviderRegion     types.String `tfsdk:"cloud_provider_region"`
	DistinguishedName       types.String `tfsdk:"distinguished_name"`
	Organization            types.String `tfsdk:"organization"`
	OrganizationalUnit      types.String `tfsdk:"organizational_unit"`
	Zone                    types.String `tfsdk:"zone"`
	HostType                types.String `tfsdk:"host_type"`
	HostClassification      types.String `tfsdk:"host_classification"`
	Comment                 types.String `tfsdk:"comment"`
	UserMessage             types.String `tfsdk:"user_message"`
	Disabled                types.String `tfsdk:"disabled"`
	Deployable              types.Bool   `tfsdk:"deployable"`
	Tofu                    types.Bool   `tfsdk:"tofu"`
	Toch                    types.Bool   `tfsdk:"toch"`
	AuditEnabled            types.Bool   `tfsdk:"audit_enabled"`
	PasswordRotationEnabled types.Bool   `tfsdk:"password_rotation_enabled"`
	ContactAddress          types.String `tfsdk:"contact_address"`
	Tags                    types.List   `tfsdk:"tags"`
	Services                types.List   `tfsdk:"services"`
	Principals              types.List   `tfsdk:"principals"`
	SSHHostPublicKeys       types.List   `tfsdk:"ssh_host_public_keys"`
	SessionRecordingOptions types.Object `tfsdk:"session_recording_options"`
	Created                 types.String `tfsdk:"created"`
	Updated                 types.String `tfsdk:"updated"`
	UpdatedBy               types.String `tfsdk:"updated_by"`
}

func (d *HostDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_host"
}

func (d *HostDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "PrivX Host data source",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Host UUID. Either id or common_name must be specified.",
				Optional:            true,
				Computed:            true,
			},
			"common_name": schema.StringAttribute{
				MarkdownDescription: "Host common name. Either id or common_name must be specified.",
				Optional:            true,
				Computed:            true,
			},
			"addresses": schema.ListAttribute{
				MarkdownDescription: "List of host addresses",
				ElementType:         types.StringType,
				Computed:            true,
			},
			"external_id": schema.StringAttribute{
				MarkdownDescription: "External ID for the host",
				Computed:            true,
			},
			"instance_id": schema.StringAttribute{
				MarkdownDescription: "Instance ID for the host",
				Computed:            true,
			},
			"source_id": schema.StringAttribute{
				MarkdownDescription: "Source ID for the host",
				Computed:            true,
			},
			"access_group_id": schema.StringAttribute{
				MarkdownDescription: "Access Group ID for the host",
				Computed:            true,
			},
			"cloud_provider": schema.StringAttribute{
				MarkdownDescription: "Cloud provider for the host",
				Computed:            true,
			},
			"cloud_provider_region": schema.StringAttribute{
				MarkdownDescription: "Cloud provider region for the host",
				Computed:            true,
			},
			"distinguished_name": schema.StringAttribute{
				MarkdownDescription: "Distinguished name for the host",
				Computed:            true,
			},
			"organization": schema.StringAttribute{
				MarkdownDescription: "Organization for the host",
				Computed:            true,
			},
			"organizational_unit": schema.StringAttribute{
				MarkdownDescription: "Organizational unit for the host",
				Computed:            true,
			},
			"zone": schema.StringAttribute{
				MarkdownDescription: "Zone for the host",
				Computed:            true,
			},
			"host_type": schema.StringAttribute{
				MarkdownDescription: "Host type",
				Computed:            true,
			},
			"host_classification": schema.StringAttribute{
				MarkdownDescription: "Host classification",
				Computed:            true,
			},
			"comment": schema.StringAttribute{
				MarkdownDescription: "Comment for the host",
				Computed:            true,
			},
			"user_message": schema.StringAttribute{
				MarkdownDescription: "User message for the host",
				Computed:            true,
			},
			"disabled": schema.StringAttribute{
				MarkdownDescription: "Whether the host is disabled",
				Computed:            true,
			},
			"deployable": schema.BoolAttribute{
				MarkdownDescription: "Whether the host is deployable",
				Computed:            true,
			},
			"tofu": schema.BoolAttribute{
				MarkdownDescription: "TOFU (Trust On First Use) setting",
				Computed:            true,
			},
			"toch": schema.BoolAttribute{
				MarkdownDescription: "TOCH setting",
				Computed:            true,
			},
			"audit_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether audit is enabled for the host",
				Computed:            true,
			},
			"password_rotation_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether password rotation is enabled",
				Computed:            true,
			},
			"contact_address": schema.StringAttribute{
				MarkdownDescription: "Contact address for the host",
				Computed:            true,
			},
			"tags": schema.ListAttribute{
				MarkdownDescription: "List of tags for the host",
				ElementType:         types.StringType,
				Computed:            true,
			},
			"services": schema.ListNestedAttribute{
				MarkdownDescription: "List of services for the host",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"service": schema.StringAttribute{
							MarkdownDescription: "Service type (e.g., SSH, RDP, HTTP)",
							Computed:            true,
						},
						"address": schema.StringAttribute{
							MarkdownDescription: "Service address",
							Computed:            true,
						},
						"port": schema.Int64Attribute{
							MarkdownDescription: "Service port",
							Computed:            true,
						},
						"use_for_password_rotation": schema.BoolAttribute{
							MarkdownDescription: "Use this service for password rotation",
							Computed:            true,
						},
						"ssh_tunnel_port": schema.Int64Attribute{
							MarkdownDescription: "SSH tunnel port",
							Computed:            true,
						},
						"use_plaintext_vnc": schema.BoolAttribute{
							MarkdownDescription: "Use plaintext VNC",
							Computed:            true,
						},
						"source": schema.StringAttribute{
							MarkdownDescription: "Service source",
							Computed:            true,
						},
						"status": schema.StringAttribute{
							MarkdownDescription: "Service status",
							Computed:            true,
						},
						"status_updated": schema.StringAttribute{
							MarkdownDescription: "Service status last updated",
							Computed:            true,
						},
					},
				},
			},
			"principals": schema.ListNestedAttribute{
				MarkdownDescription: "List of principals for the host",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"principal": schema.StringAttribute{
							MarkdownDescription: "Principal name",
							Computed:            true,
						},
						"passphrase": schema.StringAttribute{
							MarkdownDescription: "Principal passphrase",
							Computed:            true,
							Sensitive:           true,
						},
						"rotate": schema.BoolAttribute{
							MarkdownDescription: "Whether to rotate the principal",
							Computed:            true,
						},
						"use_for_password_rotation": schema.BoolAttribute{
							MarkdownDescription: "Use this principal for password rotation",
							Computed:            true,
						},
						"username_attribute": schema.StringAttribute{
							MarkdownDescription: "Username attribute",
							Computed:            true,
						},
						"use_user_account": schema.BoolAttribute{
							MarkdownDescription: "Use user account",
							Computed:            true,
						},
						"source": schema.StringAttribute{
							MarkdownDescription: "Principal source",
							Computed:            true,
						},
						"roles": schema.ListNestedAttribute{
							MarkdownDescription: "List of roles for the principal",
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
						"applications": schema.ListAttribute{
							MarkdownDescription: "List of applications for the principal",
							ElementType:         types.StringType,
							Computed:            true,
						},
						"service_options": schema.SingleNestedAttribute{
							MarkdownDescription: "Service options for the principal",
							Computed:            true,
							Attributes: map[string]schema.Attribute{
								"ssh": schema.SingleNestedAttribute{
									MarkdownDescription: "SSH service options",
									Computed:            true,
									Attributes: map[string]schema.Attribute{
										"shell": schema.BoolAttribute{
											MarkdownDescription: "Allow shell access",
											Computed:            true,
										},
										"file_transfer": schema.BoolAttribute{
											MarkdownDescription: "Allow file transfer",
											Computed:            true,
										},
										"exec": schema.BoolAttribute{
											MarkdownDescription: "Allow exec commands",
											Computed:            true,
										},
										"tunnels": schema.BoolAttribute{
											MarkdownDescription: "Allow tunnels",
											Computed:            true,
										},
										"x11": schema.BoolAttribute{
											MarkdownDescription: "Allow X11 forwarding",
											Computed:            true,
										},
										"other": schema.BoolAttribute{
											MarkdownDescription: "Allow other SSH features",
											Computed:            true,
										},
									},
								},
								"rdp": schema.SingleNestedAttribute{
									MarkdownDescription: "RDP service options",
									Computed:            true,
									Attributes: map[string]schema.Attribute{
										"file_transfer": schema.BoolAttribute{
											MarkdownDescription: "Allow file transfer",
											Computed:            true,
										},
										"audio": schema.BoolAttribute{
											MarkdownDescription: "Allow audio",
											Computed:            true,
										},
										"clipboard": schema.BoolAttribute{
											MarkdownDescription: "Allow clipboard",
											Computed:            true,
										},
									},
								},
								"web": schema.SingleNestedAttribute{
									MarkdownDescription: "Web service options",
									Computed:            true,
									Attributes: map[string]schema.Attribute{
										"file_transfer": schema.BoolAttribute{
											MarkdownDescription: "Allow file transfer",
											Computed:            true,
										},
										"audio": schema.BoolAttribute{
											MarkdownDescription: "Allow audio",
											Computed:            true,
										},
										"clipboard": schema.BoolAttribute{
											MarkdownDescription: "Allow clipboard",
											Computed:            true,
										},
									},
								},
								"vnc": schema.SingleNestedAttribute{
									MarkdownDescription: "VNC service options",
									Computed:            true,
									Attributes: map[string]schema.Attribute{
										"file_transfer": schema.BoolAttribute{
											MarkdownDescription: "Allow file transfer",
											Computed:            true,
										},
										"clipboard": schema.BoolAttribute{
											MarkdownDescription: "Allow clipboard",
											Computed:            true,
										},
									},
								},
								"db": schema.SingleNestedAttribute{
									MarkdownDescription: "Database service options",
									Computed:            true,
									Attributes: map[string]schema.Attribute{
										"max_bytes_upload": schema.Int64Attribute{
											MarkdownDescription: "Maximum bytes for upload",
											Computed:            true,
										},
										"max_bytes_download": schema.Int64Attribute{
											MarkdownDescription: "Maximum bytes for download",
											Computed:            true,
										},
									},
								},
							},
						},
						"command_restrictions": schema.SingleNestedAttribute{
							MarkdownDescription: "Command restrictions for the principal",
							Computed:            true,
							Attributes: map[string]schema.Attribute{
								"enabled": schema.BoolAttribute{
									MarkdownDescription: "Enable command restrictions",
									Computed:            true,
								},
								"rshell_variant": schema.StringAttribute{
									MarkdownDescription: "Shell variant (e.g., bash, sh)",
									Computed:            true,
								},
								"default_whitelist": schema.SingleNestedAttribute{
									MarkdownDescription: "Default whitelist",
									Computed:            true,
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											MarkdownDescription: "Whitelist ID",
											Computed:            true,
										},
										"name": schema.StringAttribute{
											MarkdownDescription: "Whitelist name",
											Computed:            true,
										},
									},
								},
								"whitelists": schema.ListNestedAttribute{
									MarkdownDescription: "List of whitelists with roles",
									Computed:            true,
									NestedObject: schema.NestedAttributeObject{
										Attributes: map[string]schema.Attribute{
											"whitelist": schema.SingleNestedAttribute{
												MarkdownDescription: "Whitelist reference",
												Computed:            true,
												Attributes: map[string]schema.Attribute{
													"id": schema.StringAttribute{
														MarkdownDescription: "Whitelist ID",
														Computed:            true,
													},
													"name": schema.StringAttribute{
														MarkdownDescription: "Whitelist name",
														Computed:            true,
													},
												},
											},
											"roles": schema.ListNestedAttribute{
												MarkdownDescription: "Roles for this whitelist",
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
										},
									},
								},
								"allow_no_match": schema.BoolAttribute{
									MarkdownDescription: "Allow commands that don't match any whitelist",
									Computed:            true,
								},
								"audit_match": schema.BoolAttribute{
									MarkdownDescription: "Audit commands that match whitelists",
									Computed:            true,
								},
								"audit_no_match": schema.BoolAttribute{
									MarkdownDescription: "Audit commands that don't match whitelists",
									Computed:            true,
								},
								"banner": schema.StringAttribute{
									MarkdownDescription: "Banner message to display",
									Computed:            true,
								},
							},
						},
					},
				},
			},
			"ssh_host_public_keys": schema.ListNestedAttribute{
				MarkdownDescription: "List of SSH host public keys",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"key": schema.StringAttribute{
							MarkdownDescription: "SSH public key",
							Computed:            true,
						},
						"fingerprint": schema.StringAttribute{
							MarkdownDescription: "SSH key fingerprint",
							Computed:            true,
						},
					},
				},
			},
			"session_recording_options": schema.SingleNestedAttribute{
				MarkdownDescription: "Session recording options",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"disable_clipboard_recording": schema.BoolAttribute{
						MarkdownDescription: "Disable clipboard recording",
						Computed:            true,
					},
					"disable_file_transfer_recording": schema.BoolAttribute{
						MarkdownDescription: "Disable file transfer recording",
						Computed:            true,
					},
				},
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
				MarkdownDescription: "ID of user who last updated the host",
				Computed:            true,
			},
		},
	}
}

func (d *HostDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	tflog.Debug(ctx, "Creating hoststore client", map[string]interface{}{
		"connector": fmt.Sprintf("%+v", *connector),
	})

	d.client = hoststore.New(*connector)
}

func (d *HostDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data HostDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Host data source read started", map[string]interface{}{
		"id":          data.ID.ValueString(),
		"common_name": data.CommonName.ValueString(),
	})

	var host *hoststore.Host
	var err error

	// Check if ID is provided
	if !data.ID.IsNull() && !data.ID.IsUnknown() && data.ID.ValueString() != "" {
		// Get by ID
		host, err = d.client.GetHost(data.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read host by ID, got error: %s", err))
			return
		}
	} else if !data.CommonName.IsNull() && !data.CommonName.IsUnknown() && data.CommonName.ValueString() != "" {
		// Get by common name - try search functionality first
		searchQuery := &hoststore.HostSearch{
			CommonName: []string{data.CommonName.ValueString()},
		}

		tflog.Debug(ctx, "Searching for host by common name using SearchHosts", map[string]interface{}{
			"common_name": data.CommonName.ValueString(),
		})

		searchResult, err := d.client.SearchHosts(searchQuery)
		if err != nil {
			tflog.Debug(ctx, "SearchHosts failed, trying GetHosts as fallback", map[string]interface{}{
				"error": err.Error(),
			})

			// Fallback to GetHosts and manual filtering
			getAllResult, getAllErr := d.client.GetHosts()
			if getAllErr != nil {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to search hosts by common name (SearchHosts failed: %s, GetHosts failed: %s)", err.Error(), getAllErr.Error()))
				return
			}

			tflog.Debug(ctx, "GetHosts returned results", map[string]interface{}{
				"total_hosts": len(getAllResult.Items),
			})

			for i := range getAllResult.Items {
				if getAllResult.Items[i].CommonName == data.CommonName.ValueString() {
					host = &getAllResult.Items[i]
					break
				}
			}

			if host == nil {
				resp.Diagnostics.AddError("Not Found", fmt.Sprintf("Host with common name '%s' not found (searched %d hosts)", data.CommonName.ValueString(), len(getAllResult.Items)))
				return
			}
		} else {
			tflog.Debug(ctx, "SearchHosts returned results", map[string]interface{}{
				"total_results": len(searchResult.Items),
			})

			if len(searchResult.Items) == 0 {
				resp.Diagnostics.AddError("Not Found", fmt.Sprintf("Host with common name '%s' not found", data.CommonName.ValueString()))
				return
			}

			// Use the first match
			host = &searchResult.Items[0]
		}

		tflog.Debug(ctx, "Found host by common name", map[string]interface{}{
			"host_id":     host.ID,
			"common_name": host.CommonName,
		})
	} else {
		resp.Diagnostics.AddError("Missing Required Attribute", "Either 'id' or 'common_name' must be specified")
		return
	}
	tflog.Debug(ctx, "Storing host into the state", map[string]interface{}{
		"host_id":     host.ID,
		"common_name": host.CommonName,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
