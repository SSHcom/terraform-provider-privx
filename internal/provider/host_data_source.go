package provider

import (
	"context"
	"fmt"

	"github.com/SSHcom/privx-sdk-go/v2/api/hoststore"
	"github.com/SSHcom/privx-sdk-go/v2/restapi"
	"github.com/hashicorp/terraform-plugin-framework/attr"
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

			for _, h := range getAllResult.Items {
				tflog.Debug(ctx, "Checking host", map[string]interface{}{
					"host_id":     h.ID,
					"common_name": h.CommonName,
					"target_name": data.CommonName.ValueString(),
				})

				if h.CommonName == data.CommonName.ValueString() {
					host = &h
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

	// Populate the data source model
	d.populateHostDataSourceModel(ctx, &data, host)

	tflog.Debug(ctx, "Storing host into the state", map[string]interface{}{
		"state": fmt.Sprintf("%+v", data),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// populateHostDataSourceModel populates the Terraform data source model from the API response
func (d *HostDataSource) populateHostDataSourceModel(ctx context.Context, data *HostDataSourceModel, host *hoststore.Host) {
	data.ID = types.StringValue(host.ID)
	data.CommonName = types.StringValue(host.CommonName)
	data.ExternalID = types.StringValue(host.ExternalID)
	data.InstanceID = types.StringValue(host.InstanceID)
	data.SourceID = types.StringValue(host.SourceID)
	data.AccessGroupID = types.StringValue(host.AccessGroupID)
	data.CloudProvider = types.StringValue(host.CloudProvider)
	data.CloudProviderRegion = types.StringValue(host.CloudProviderRegion)
	data.DistinguishedName = types.StringValue(host.DistinguishedName)
	data.Organization = types.StringValue(host.Organization)
	data.OrganizationalUnit = types.StringValue(host.OrganizationalUnit)
	data.Zone = types.StringValue(host.Zone)
	data.HostType = types.StringValue(host.HostType)
	data.HostClassification = types.StringValue(host.HostClassification)
	data.Comment = types.StringValue(host.Comment)
	data.UserMessage = types.StringValue(host.UserMessage)
	data.Disabled = types.StringValue(host.Disabled)
	if host.Deployable != nil {
		data.Deployable = types.BoolValue(*host.Deployable)
	} else {
		data.Deployable = types.BoolValue(false)
	}
	if host.Tofu != nil {
		data.Tofu = types.BoolValue(*host.Tofu)
	} else {
		data.Tofu = types.BoolValue(false)
	}
	if host.Toch != nil {
		data.Toch = types.BoolValue(*host.Toch)
	} else {
		data.Toch = types.BoolValue(false)
	}
	if host.AuditEnabled != nil {
		data.AuditEnabled = types.BoolValue(*host.AuditEnabled)
	} else {
		data.AuditEnabled = types.BoolValue(false)
	}
	data.PasswordRotationEnabled = types.BoolValue(host.PasswordRotationEnabled)
	data.ContactAddress = types.StringValue(host.ContactAddress)
	data.Created = types.StringValue(host.Created)
	data.Updated = types.StringValue(host.Updated)
	data.UpdatedBy = types.StringValue(host.UpdatedBy)

	// Convert addresses slice to list
	addressValues := make([]attr.Value, len(host.Addresses))
	for i, addr := range host.Addresses {
		addressValues[i] = types.StringValue(addr)
	}
	data.Addresses = types.ListValueMust(types.StringType, addressValues)

	// Convert tags slice to list
	tagValues := make([]attr.Value, len(host.Tags))
	for i, tag := range host.Tags {
		tagValues[i] = types.StringValue(tag)
	}
	data.Tags = types.ListValueMust(types.StringType, tagValues)

	// Convert services
	serviceValues := make([]attr.Value, len(host.Services))
	for i, service := range host.Services {
		serviceAttrs := map[string]attr.Value{
			"service":                   types.StringValue(service.Service),
			"address":                   types.StringValue(service.Address),
			"port":                      types.Int64Value(int64(service.Port)),
			"use_for_password_rotation": types.BoolValue(service.UseForPasswordRotation),
			"ssh_tunnel_port":           types.Int64Value(int64(service.TunnelPort)),
			"use_plaintext_vnc":         types.BoolValue(service.UsePlainTextVNC),
			"source":                    types.StringValue(service.Source),
			"status":                    types.StringValue(service.HealthCheckStatus),
			"status_updated":            types.StringValue(service.HealthCheckStatusUpdated),
		}
		serviceValues[i] = types.ObjectValueMust(map[string]attr.Type{
			"service":                   types.StringType,
			"address":                   types.StringType,
			"port":                      types.Int64Type,
			"use_for_password_rotation": types.BoolType,
			"ssh_tunnel_port":           types.Int64Type,
			"use_plaintext_vnc":         types.BoolType,
			"source":                    types.StringType,
			"status":                    types.StringType,
			"status_updated":            types.StringType,
		}, serviceAttrs)
	}
	data.Services = types.ListValueMust(types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"service":                   types.StringType,
			"address":                   types.StringType,
			"port":                      types.Int64Type,
			"use_for_password_rotation": types.BoolType,
			"ssh_tunnel_port":           types.Int64Type,
			"use_plaintext_vnc":         types.BoolType,
			"source":                    types.StringType,
			"status":                    types.StringType,
			"status_updated":            types.StringType,
		},
	}, serviceValues)

	// Convert principals
	principalValues := make([]attr.Value, len(host.Principals))
	for i, principal := range host.Principals {
		// Convert roles for this principal
		roleValues := make([]attr.Value, len(principal.Roles))
		for j, role := range principal.Roles {
			roleAttrs := map[string]attr.Value{
				"id":   types.StringValue(role.ID),
				"name": types.StringValue(role.Name),
			}
			roleValues[j] = types.ObjectValueMust(map[string]attr.Type{
				"id":   types.StringType,
				"name": types.StringType,
			}, roleAttrs)
		}

		// Convert applications for this principal
		appValues := make([]attr.Value, len(principal.Applications))
		for j, app := range principal.Applications {
			appValues[j] = types.StringValue(app.Name)
		}

		// Convert service options for this principal
		var serviceOptionsValue attr.Value
		if principal.ServiceOptions != nil {
			serviceOptionsAttrs := map[string]attr.Value{}

			// SSH options
			if principal.ServiceOptions.SSHServiceOptions != nil {
				sshAttrs := map[string]attr.Value{
					"shell":         types.BoolValue(principal.ServiceOptions.SSHServiceOptions.Shell),
					"file_transfer": types.BoolValue(principal.ServiceOptions.SSHServiceOptions.FileTransfer),
					"exec":          types.BoolValue(principal.ServiceOptions.SSHServiceOptions.Exec),
					"tunnels":       types.BoolValue(principal.ServiceOptions.SSHServiceOptions.Tunnels),
					"x11":           types.BoolValue(principal.ServiceOptions.SSHServiceOptions.X11),
					"other":         types.BoolValue(principal.ServiceOptions.SSHServiceOptions.Other),
				}
				serviceOptionsAttrs["ssh"] = types.ObjectValueMust(map[string]attr.Type{
					"shell":         types.BoolType,
					"file_transfer": types.BoolType,
					"exec":          types.BoolType,
					"tunnels":       types.BoolType,
					"x11":           types.BoolType,
					"other":         types.BoolType,
				}, sshAttrs)
			} else {
				serviceOptionsAttrs["ssh"] = types.ObjectNull(map[string]attr.Type{
					"shell":         types.BoolType,
					"file_transfer": types.BoolType,
					"exec":          types.BoolType,
					"tunnels":       types.BoolType,
					"x11":           types.BoolType,
					"other":         types.BoolType,
				})
			}

			// RDP options
			if principal.ServiceOptions.RDPServiceOptions != nil {
				rdpAttrs := map[string]attr.Value{
					"file_transfer": types.BoolValue(principal.ServiceOptions.RDPServiceOptions.FileTransfer),
					"audio":         types.BoolValue(principal.ServiceOptions.RDPServiceOptions.Audio),
					"clipboard":     types.BoolValue(principal.ServiceOptions.RDPServiceOptions.Clipboard),
				}
				serviceOptionsAttrs["rdp"] = types.ObjectValueMust(map[string]attr.Type{
					"file_transfer": types.BoolType,
					"audio":         types.BoolType,
					"clipboard":     types.BoolType,
				}, rdpAttrs)
			} else {
				serviceOptionsAttrs["rdp"] = types.ObjectNull(map[string]attr.Type{
					"file_transfer": types.BoolType,
					"audio":         types.BoolType,
					"clipboard":     types.BoolType,
				})
			}

			// Web options
			if principal.ServiceOptions.WebServiceOptions != nil {
				webAttrs := map[string]attr.Value{
					"file_transfer": types.BoolValue(principal.ServiceOptions.WebServiceOptions.FileTransfer),
					"audio":         types.BoolValue(principal.ServiceOptions.WebServiceOptions.Audio),
					"clipboard":     types.BoolValue(principal.ServiceOptions.WebServiceOptions.Clipboard),
				}
				serviceOptionsAttrs["web"] = types.ObjectValueMust(map[string]attr.Type{
					"file_transfer": types.BoolType,
					"audio":         types.BoolType,
					"clipboard":     types.BoolType,
				}, webAttrs)
			} else {
				serviceOptionsAttrs["web"] = types.ObjectNull(map[string]attr.Type{
					"file_transfer": types.BoolType,
					"audio":         types.BoolType,
					"clipboard":     types.BoolType,
				})
			}

			// VNC options
			if principal.ServiceOptions.VNCServiceOptions != nil {
				vncAttrs := map[string]attr.Value{
					"file_transfer": types.BoolValue(principal.ServiceOptions.VNCServiceOptions.FileTransfer),
					"clipboard":     types.BoolValue(principal.ServiceOptions.VNCServiceOptions.Clipboard),
				}
				serviceOptionsAttrs["vnc"] = types.ObjectValueMust(map[string]attr.Type{
					"file_transfer": types.BoolType,
					"clipboard":     types.BoolType,
				}, vncAttrs)
			} else {
				serviceOptionsAttrs["vnc"] = types.ObjectNull(map[string]attr.Type{
					"file_transfer": types.BoolType,
					"clipboard":     types.BoolType,
				})
			}

			// DB options
			if principal.ServiceOptions.DBServiceOptions != nil {
				dbAttrs := map[string]attr.Value{
					"max_bytes_upload":   types.Int64Value(principal.ServiceOptions.DBServiceOptions.MaxBytesUpload),
					"max_bytes_download": types.Int64Value(principal.ServiceOptions.DBServiceOptions.MaxBytesDownload),
				}
				serviceOptionsAttrs["db"] = types.ObjectValueMust(map[string]attr.Type{
					"max_bytes_upload":   types.Int64Type,
					"max_bytes_download": types.Int64Type,
				}, dbAttrs)
			} else {
				serviceOptionsAttrs["db"] = types.ObjectNull(map[string]attr.Type{
					"max_bytes_upload":   types.Int64Type,
					"max_bytes_download": types.Int64Type,
				})
			}

			serviceOptionsValue = types.ObjectValueMust(map[string]attr.Type{
				"ssh": types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"shell":         types.BoolType,
						"file_transfer": types.BoolType,
						"exec":          types.BoolType,
						"tunnels":       types.BoolType,
						"x11":           types.BoolType,
						"other":         types.BoolType,
					},
				},
				"rdp": types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"file_transfer": types.BoolType,
						"audio":         types.BoolType,
						"clipboard":     types.BoolType,
					},
				},
				"web": types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"file_transfer": types.BoolType,
						"audio":         types.BoolType,
						"clipboard":     types.BoolType,
					},
				},
				"vnc": types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"file_transfer": types.BoolType,
						"clipboard":     types.BoolType,
					},
				},
				"db": types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"max_bytes_upload":   types.Int64Type,
						"max_bytes_download": types.Int64Type,
					},
				},
			}, serviceOptionsAttrs)
		} else {
			serviceOptionsValue = types.ObjectNull(map[string]attr.Type{
				"ssh": types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"shell":         types.BoolType,
						"file_transfer": types.BoolType,
						"exec":          types.BoolType,
						"tunnels":       types.BoolType,
						"x11":           types.BoolType,
						"other":         types.BoolType,
					},
				},
				"rdp": types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"file_transfer": types.BoolType,
						"audio":         types.BoolType,
						"clipboard":     types.BoolType,
					},
				},
				"web": types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"file_transfer": types.BoolType,
						"audio":         types.BoolType,
						"clipboard":     types.BoolType,
					},
				},
				"vnc": types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"file_transfer": types.BoolType,
						"clipboard":     types.BoolType,
					},
				},
				"db": types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"max_bytes_upload":   types.Int64Type,
						"max_bytes_download": types.Int64Type,
					},
				},
			})
		}

		principalAttrs := map[string]attr.Value{
			"principal":                 types.StringValue(principal.Principal),
			"passphrase":                types.StringValue(principal.Passphrase),
			"rotate":                    types.BoolValue(principal.Rotate),
			"use_for_password_rotation": types.BoolValue(principal.UseForPasswordRotation),
			"username_attribute":        types.StringValue(principal.UsernameAttribute),
			"use_user_account":          types.BoolValue(principal.UseUserAccount),
			"source":                    types.StringValue(principal.Source),
			"roles": types.ListValueMust(types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"id":   types.StringType,
					"name": types.StringType,
				},
			}, roleValues),
			"applications":    types.ListValueMust(types.StringType, appValues),
			"service_options": serviceOptionsValue,
		}

		// Convert command restrictions for this principal
		commandRestrictionsAttrs := map[string]attr.Value{
			"enabled":        types.BoolValue(principal.CommandRestrictions.Enabled),
			"rshell_variant": types.StringValue(principal.CommandRestrictions.RShellVariant),
			"allow_no_match": types.BoolValue(principal.CommandRestrictions.AllowNoMatch),
			"audit_match":    types.BoolValue(principal.CommandRestrictions.AuditMatch),
			"audit_no_match": types.BoolValue(principal.CommandRestrictions.AuditNoMatch),
			"banner":         types.StringValue(principal.CommandRestrictions.Banner),
		}

		// Convert default whitelist
		defaultWhitelistAttrs := map[string]attr.Value{
			"id":   types.StringValue(principal.CommandRestrictions.DefaultWhiteList.ID),
			"name": types.StringValue(principal.CommandRestrictions.DefaultWhiteList.Name),
		}
		commandRestrictionsAttrs["default_whitelist"] = types.ObjectValueMust(map[string]attr.Type{
			"id":   types.StringType,
			"name": types.StringType,
		}, defaultWhitelistAttrs)

		// Convert whitelists
		whitelistValues := make([]attr.Value, len(principal.CommandRestrictions.WhiteLists))
		for j, whitelistGrant := range principal.CommandRestrictions.WhiteLists {
			// Convert whitelist handle
			whitelistHandleAttrs := map[string]attr.Value{
				"id":   types.StringValue(whitelistGrant.WhiteList.ID),
				"name": types.StringValue(whitelistGrant.WhiteList.Name),
			}

			// Convert roles for this whitelist
			whitelistRoleValues := make([]attr.Value, len(whitelistGrant.Roles))
			for k, role := range whitelistGrant.Roles {
				roleAttrs := map[string]attr.Value{
					"id":   types.StringValue(role.ID),
					"name": types.StringValue(role.Name),
				}
				whitelistRoleValues[k] = types.ObjectValueMust(map[string]attr.Type{
					"id":   types.StringType,
					"name": types.StringType,
				}, roleAttrs)
			}

			whitelistGrantAttrs := map[string]attr.Value{
				"whitelist": types.ObjectValueMust(map[string]attr.Type{
					"id":   types.StringType,
					"name": types.StringType,
				}, whitelistHandleAttrs),
				"roles": types.ListValueMust(types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"id":   types.StringType,
						"name": types.StringType,
					},
				}, whitelistRoleValues),
			}

			whitelistValues[j] = types.ObjectValueMust(map[string]attr.Type{
				"whitelist": types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"id":   types.StringType,
						"name": types.StringType,
					},
				},
				"roles": types.ListType{
					ElemType: types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"id":   types.StringType,
							"name": types.StringType,
						},
					},
				},
			}, whitelistGrantAttrs)
		}

		commandRestrictionsAttrs["whitelists"] = types.ListValueMust(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"whitelist": types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"id":   types.StringType,
						"name": types.StringType,
					},
				},
				"roles": types.ListType{
					ElemType: types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"id":   types.StringType,
							"name": types.StringType,
						},
					},
				},
			},
		}, whitelistValues)

		commandRestrictionsValue := types.ObjectValueMust(map[string]attr.Type{
			"enabled":        types.BoolType,
			"rshell_variant": types.StringType,
			"allow_no_match": types.BoolType,
			"audit_match":    types.BoolType,
			"audit_no_match": types.BoolType,
			"banner":         types.StringType,
			"default_whitelist": types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"id":   types.StringType,
					"name": types.StringType,
				},
			},
			"whitelists": types.ListType{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"whitelist": types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"id":   types.StringType,
								"name": types.StringType,
							},
						},
						"roles": types.ListType{
							ElemType: types.ObjectType{
								AttrTypes: map[string]attr.Type{
									"id":   types.StringType,
									"name": types.StringType,
								},
							},
						},
					},
				},
			},
		}, commandRestrictionsAttrs)

		principalAttrs["command_restrictions"] = commandRestrictionsValue
		principalValues[i] = types.ObjectValueMust(map[string]attr.Type{
			"principal":                 types.StringType,
			"passphrase":                types.StringType,
			"rotate":                    types.BoolType,
			"use_for_password_rotation": types.BoolType,
			"username_attribute":        types.StringType,
			"use_user_account":          types.BoolType,
			"source":                    types.StringType,
			"roles": types.ListType{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"id":   types.StringType,
						"name": types.StringType,
					},
				},
			},
			"applications": types.ListType{ElemType: types.StringType},
			"service_options": types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"ssh": types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"shell":         types.BoolType,
							"file_transfer": types.BoolType,
							"exec":          types.BoolType,
							"tunnels":       types.BoolType,
							"x11":           types.BoolType,
							"other":         types.BoolType,
						},
					},
					"rdp": types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"file_transfer": types.BoolType,
							"audio":         types.BoolType,
							"clipboard":     types.BoolType,
						},
					},
					"web": types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"file_transfer": types.BoolType,
							"audio":         types.BoolType,
							"clipboard":     types.BoolType,
						},
					},
					"vnc": types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"file_transfer": types.BoolType,
							"clipboard":     types.BoolType,
						},
					},
					"db": types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"max_bytes_upload":   types.Int64Type,
							"max_bytes_download": types.Int64Type,
						},
					},
				},
			},
			"command_restrictions": types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"enabled":        types.BoolType,
					"rshell_variant": types.StringType,
					"allow_no_match": types.BoolType,
					"audit_match":    types.BoolType,
					"audit_no_match": types.BoolType,
					"banner":         types.StringType,
					"default_whitelist": types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"id":   types.StringType,
							"name": types.StringType,
						},
					},
					"whitelists": types.ListType{
						ElemType: types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"whitelist": types.ObjectType{
									AttrTypes: map[string]attr.Type{
										"id":   types.StringType,
										"name": types.StringType,
									},
								},
								"roles": types.ListType{
									ElemType: types.ObjectType{
										AttrTypes: map[string]attr.Type{
											"id":   types.StringType,
											"name": types.StringType,
										},
									},
								},
							},
						},
					},
				},
			},
		}, principalAttrs)
	}
	data.Principals = types.ListValueMust(types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"principal":                 types.StringType,
			"passphrase":                types.StringType,
			"rotate":                    types.BoolType,
			"use_for_password_rotation": types.BoolType,
			"username_attribute":        types.StringType,
			"use_user_account":          types.BoolType,
			"source":                    types.StringType,
			"roles": types.ListType{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"id":   types.StringType,
						"name": types.StringType,
					},
				},
			},
			"applications": types.ListType{ElemType: types.StringType},
			"service_options": types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"ssh": types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"shell":         types.BoolType,
							"file_transfer": types.BoolType,
							"exec":          types.BoolType,
							"tunnels":       types.BoolType,
							"x11":           types.BoolType,
							"other":         types.BoolType,
						},
					},
					"rdp": types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"file_transfer": types.BoolType,
							"audio":         types.BoolType,
							"clipboard":     types.BoolType,
						},
					},
					"web": types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"file_transfer": types.BoolType,
							"audio":         types.BoolType,
							"clipboard":     types.BoolType,
						},
					},
					"vnc": types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"file_transfer": types.BoolType,
							"clipboard":     types.BoolType,
						},
					},
					"db": types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"max_bytes_upload":   types.Int64Type,
							"max_bytes_download": types.Int64Type,
						},
					},
				},
			},
			"command_restrictions": types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"enabled":        types.BoolType,
					"rshell_variant": types.StringType,
					"allow_no_match": types.BoolType,
					"audit_match":    types.BoolType,
					"audit_no_match": types.BoolType,
					"banner":         types.StringType,
					"default_whitelist": types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"id":   types.StringType,
							"name": types.StringType,
						},
					},
					"whitelists": types.ListType{
						ElemType: types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"whitelist": types.ObjectType{
									AttrTypes: map[string]attr.Type{
										"id":   types.StringType,
										"name": types.StringType,
									},
								},
								"roles": types.ListType{
									ElemType: types.ObjectType{
										AttrTypes: map[string]attr.Type{
											"id":   types.StringType,
											"name": types.StringType,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}, principalValues)

	// Convert SSH host public keys
	keyValues := make([]attr.Value, len(host.SSHHostPubKeys))
	for i, key := range host.SSHHostPubKeys {
		keyAttrs := map[string]attr.Value{
			"key":         types.StringValue(key.Key),
			"fingerprint": types.StringValue(key.FingerPrint),
		}
		keyValues[i] = types.ObjectValueMust(map[string]attr.Type{
			"key":         types.StringType,
			"fingerprint": types.StringType,
		}, keyAttrs)
	}
	data.SSHHostPublicKeys = types.ListValueMust(types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"key":         types.StringType,
			"fingerprint": types.StringType,
		},
	}, keyValues)

	// Convert session recording options
	var sroAttrs map[string]attr.Value
	if host.SessionRecordingOptions != nil {
		sroAttrs = map[string]attr.Value{
			"disable_clipboard_recording":     types.BoolValue(host.SessionRecordingOptions.DisableClipboardRecording),
			"disable_file_transfer_recording": types.BoolValue(host.SessionRecordingOptions.DisableFileTransferRecording),
		}
	} else {
		sroAttrs = map[string]attr.Value{
			"disable_clipboard_recording":     types.BoolValue(false),
			"disable_file_transfer_recording": types.BoolValue(false),
		}
	}
	data.SessionRecordingOptions = types.ObjectValueMust(map[string]attr.Type{
		"disable_clipboard_recording":     types.BoolType,
		"disable_file_transfer_recording": types.BoolType,
	}, sroAttrs)
}
