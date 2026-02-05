package provider

import (
	"context"
	"fmt"
	"sort"
	"terraform-provider-privx/internal/utils"

	"github.com/SSHcom/privx-sdk-go/v2/api/hoststore"
	"github.com/SSHcom/privx-sdk-go/v2/restapi"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &HostResource{}
var _ resource.ResourceWithImportState = &HostResource{}

func NewHostResource() resource.Resource {
	return &HostResource{}
}

// HostResource defines the resource implementation.
type HostResource struct {
	client *hoststore.HostStore
}

// HostResourceModel contains PrivX host information.
type HostResourceModel struct {
	ID                      types.String     `tfsdk:"id"`
	CommonName              types.String     `tfsdk:"common_name"`
	Addresses               types.List       `tfsdk:"addresses"`
	ExternalID              types.String     `tfsdk:"external_id"`
	InstanceID              types.String     `tfsdk:"instance_id"`
	SourceID                types.String     `tfsdk:"source_id"`
	AccessGroupID           types.String     `tfsdk:"access_group_id"`
	CloudProvider           types.String     `tfsdk:"cloud_provider"`
	CloudProviderRegion     types.String     `tfsdk:"cloud_provider_region"`
	DistinguishedName       types.String     `tfsdk:"distinguished_name"`
	Organization            types.String     `tfsdk:"organization"`
	OrganizationalUnit      types.String     `tfsdk:"organizational_unit"`
	Zone                    types.String     `tfsdk:"zone"`
	HostType                types.String     `tfsdk:"host_type"`
	HostClassification      types.String     `tfsdk:"host_classification"`
	Comment                 types.String     `tfsdk:"comment"`
	UserMessage             types.String     `tfsdk:"user_message"`
	Disabled                types.String     `tfsdk:"disabled"`
	Deployable              types.Bool       `tfsdk:"deployable"`
	Tofu                    types.Bool       `tfsdk:"tofu"`
	Toch                    types.Bool       `tfsdk:"toch"`
	AuditEnabled            types.Bool       `tfsdk:"audit_enabled"`
	PasswordRotationEnabled types.Bool       `tfsdk:"password_rotation_enabled"`
	ContactAddress          types.String     `tfsdk:"contact_address"`
	Tags                    types.List       `tfsdk:"tags"`
	Services                []ServiceModel   `tfsdk:"services"`
	Principals              []PrincipalModel `tfsdk:"principals"`
	SSHHostPublicKeys       types.List       `tfsdk:"ssh_host_public_keys"`
	SessionRecordingOptions types.Object     `tfsdk:"session_recording_options"`
	Created                 types.String     `tfsdk:"created"`
	Updated                 types.String     `tfsdk:"updated"`
	UpdatedBy               types.String     `tfsdk:"updated_by"`
	PasswordRotation        types.Object     `tfsdk:"password_rotation"`
	StandAloneHost          types.Bool       `tfsdk:"stand_alone_host"`
}

type PasswordRotationModel struct {
	AccessGroupID                    types.String `tfsdk:"access_group_id"`
	UseMainAccount                   types.Bool   `tfsdk:"use_main_account"`
	OperatingSystem                  types.String `tfsdk:"operating_system"`
	Protocol                         types.String `tfsdk:"protocol"`
	PasswordPolicyID                 types.String `tfsdk:"password_policy_id"`
	ScriptTemplateID                 types.String `tfsdk:"script_template_id"`
	CertificateValidationOptions     types.String `tfsdk:"certificate_validation_options"`
	WinrmHostCertificateTrustAnchors types.String `tfsdk:"winrm_host_certificate_trust_anchors"`
	WinrmAddress                     types.String `tfsdk:"winrm_address"`
	WinrmPort                        types.Int64  `tfsdk:"winrm_port"`
}

// ServiceModel represents a host service.
type ServiceModel struct {
	Service                types.String `tfsdk:"service"`
	Address                types.String `tfsdk:"address"`
	Port                   types.Int64  `tfsdk:"port"`
	UseForPasswordRotation types.Bool   `tfsdk:"use_for_password_rotation"`
	SSHTunnelPort          types.Int64  `tfsdk:"ssh_tunnel_port"`
	UsePlaintextVNC        types.Bool   `tfsdk:"use_plaintext_vnc"`
	Source                 types.String `tfsdk:"source"`
}

// PrincipalModel represents a host principal.
type PrincipalModel struct {
	Principal              types.String `tfsdk:"principal"`
	Passphrase             types.String `tfsdk:"passphrase"`
	Rotate                 types.Bool   `tfsdk:"rotate"`
	UseForPasswordRotation types.Bool   `tfsdk:"use_for_password_rotation"`
	UsernameAttribute      types.String `tfsdk:"username_attribute"`
	UseUserAccount         types.Bool   `tfsdk:"use_user_account"`
	Source                 types.String `tfsdk:"source"`
	Roles                  types.List   `tfsdk:"roles"`
	Applications           types.List   `tfsdk:"applications"`
	ServiceOptions         types.Object `tfsdk:"service_options"`
	CommandRestrictions    types.Object `tfsdk:"command_restrictions"`
}

// RoleModel represents a role reference.
type RoleModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

// SSHHostPublicKeyModel represents an SSH host public key.
type SSHHostPublicKeyModel struct {
	Key         types.String `tfsdk:"key"`
	Fingerprint types.String `tfsdk:"fingerprint"`
}

// ServiceOptionsModel represents service options for a principal.
type ServiceOptionsModel struct {
	SSH types.Object `tfsdk:"ssh"`
	RDP types.Object `tfsdk:"rdp"`
	Web types.Object `tfsdk:"web"`
	VNC types.Object `tfsdk:"vnc"`
	DB  types.Object `tfsdk:"db"`
}

// SSHOptionsModel represents SSH service options.
type SSHOptionsModel struct {
	Shell        types.Bool `tfsdk:"shell"`
	FileTransfer types.Bool `tfsdk:"file_transfer"`
	Exec         types.Bool `tfsdk:"exec"`
	Tunnels      types.Bool `tfsdk:"tunnels"`
	X11          types.Bool `tfsdk:"x11"`
	Other        types.Bool `tfsdk:"other"`
}

// RDPOptionsModel represents RDP service options.
type RDPOptionsModel struct {
	FileTransfer types.Bool `tfsdk:"file_transfer"`
	Audio        types.Bool `tfsdk:"audio"`
	Clipboard    types.Bool `tfsdk:"clipboard"`
}

// WebOptionsModel represents Web service options.
type WebOptionsModel struct {
	FileTransfer types.Bool `tfsdk:"file_transfer"`
	Audio        types.Bool `tfsdk:"audio"`
	Clipboard    types.Bool `tfsdk:"clipboard"`
}

// VNCOptionsModel represents VNC service options.
type VNCOptionsModel struct {
	FileTransfer types.Bool `tfsdk:"file_transfer"`
	Clipboard    types.Bool `tfsdk:"clipboard"`
}

// DBOptionsModel represents DB service options.
type DBOptionsModel struct {
	MaxBytesUpload   types.Int64 `tfsdk:"max_bytes_upload"`
	MaxBytesDownload types.Int64 `tfsdk:"max_bytes_download"`
}

// CommandRestrictionsModel represents command restrictions.
type CommandRestrictionsModel struct {
	Enabled          types.Bool   `tfsdk:"enabled"`
	RShellVariant    types.String `tfsdk:"rshell_variant"`
	DefaultWhitelist types.Object `tfsdk:"default_whitelist"`
	Whitelists       types.List   `tfsdk:"whitelists"`
	AllowNoMatch     types.Bool   `tfsdk:"allow_no_match"`
	AuditMatch       types.Bool   `tfsdk:"audit_match"`
	AuditNoMatch     types.Bool   `tfsdk:"audit_no_match"`
	Banner           types.String `tfsdk:"banner"`
}

// WhitelistGrantModel represents a whitelist grant.
type WhitelistGrantModel struct {
	Whitelist types.Object `tfsdk:"whitelist"`
	Roles     types.List   `tfsdk:"roles"`
}

// WhitelistHandleModel represents a whitelist handle.
type WhitelistHandleModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

// SessionRecordingOptionsModel represents session recording options.
type SessionRecordingOptionsModel struct {
	DisableClipboardRecording    types.Bool `tfsdk:"disable_clipboard_recording"`
	DisableFileTransferRecording types.Bool `tfsdk:"disable_file_transfer_recording"`
}

func (r *HostResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_host"
}

func (r *HostResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Host resource for PrivX",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Host ID",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"common_name": schema.StringAttribute{
				MarkdownDescription: "Host common name",
				Required:            true,
			},
			"addresses": schema.ListAttribute{
				MarkdownDescription: "List of host addresses",
				ElementType:         types.StringType,
				Required:            true,
			},
			"external_id": schema.StringAttribute{
				MarkdownDescription: "External ID for the host",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"instance_id": schema.StringAttribute{
				MarkdownDescription: "Instance ID for the host",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"source_id": schema.StringAttribute{
				MarkdownDescription: "Source ID for the host",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"access_group_id": schema.StringAttribute{
				MarkdownDescription: "Access Group ID for the host",
				Required:            true,
			},
			"cloud_provider": schema.StringAttribute{
				MarkdownDescription: "Cloud provider for the host",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"cloud_provider_region": schema.StringAttribute{
				MarkdownDescription: "Cloud provider region for the host",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"distinguished_name": schema.StringAttribute{
				MarkdownDescription: "Distinguished name for the host",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"organization": schema.StringAttribute{
				MarkdownDescription: "Organization for the host",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"organizational_unit": schema.StringAttribute{
				MarkdownDescription: "Organizational unit for the host",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"zone": schema.StringAttribute{
				MarkdownDescription: "Zone for the host",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"host_type": schema.StringAttribute{
				MarkdownDescription: "Host type",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"host_classification": schema.StringAttribute{
				MarkdownDescription: "Host classification",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"comment": schema.StringAttribute{
				MarkdownDescription: "Comment for the host",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"user_message": schema.StringAttribute{
				MarkdownDescription: "User message for the host",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"disabled": schema.StringAttribute{
				MarkdownDescription: "Whether the host is disabled",
				Optional:            true,
				Computed:            true,
			},
			"deployable": schema.BoolAttribute{
				MarkdownDescription: "Whether the host is deployable",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"tofu": schema.BoolAttribute{
				MarkdownDescription: "TOFU (Trust On First Use) setting",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"toch": schema.BoolAttribute{
				MarkdownDescription: "TOCH setting",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"audit_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether audit is enabled for the host",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"password_rotation_enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},

			"password_rotation": schema.SingleNestedAttribute{
				MarkdownDescription: "Password rotation configuration",
				Optional:            true,
				Computed:            true, // important so state can exist even if API doesn't return it
				Attributes: map[string]schema.Attribute{
					"access_group_id": schema.StringAttribute{
						Optional: true,
						Computed: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"use_main_account": schema.BoolAttribute{
						Optional: true,
						Computed: true,
						PlanModifiers: []planmodifier.Bool{
							// add boolplanmodifier import
							boolplanmodifier.UseStateForUnknown(),
						},
					},
					"operating_system": schema.StringAttribute{
						Optional: true,
						Computed: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"protocol": schema.StringAttribute{
						Optional: true,
						Computed: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"password_policy_id": schema.StringAttribute{
						Optional: true,
						Computed: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"script_template_id": schema.StringAttribute{
						Optional: true,
						Computed: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},

					// keep these if you want them in config
					"certificate_validation_options": schema.StringAttribute{
						Optional: true,
						Computed: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"winrm_host_certificate_trust_anchors": schema.StringAttribute{
						Optional: true,
						Computed: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"winrm_address": schema.StringAttribute{
						Optional: true,
						Computed: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"winrm_port": schema.Int64Attribute{
						Optional: true,
						Computed: true,
						Default:  int64default.StaticInt64(0),
					},
				},
			},

			"contact_address": schema.StringAttribute{
				MarkdownDescription: "Contact address for the host",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"stand_alone_host": schema.BoolAttribute{
				MarkdownDescription: "Whether this is a standalone host (not managed by a discovery source).",
				Computed:            true,
			},
			"tags": schema.ListAttribute{
				MarkdownDescription: "List of tags for the host (order may change due to API sorting)",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
			},
			"services": schema.ListNestedAttribute{
				MarkdownDescription: "List of services for the host",
				Optional:            true,
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"service": schema.StringAttribute{
							MarkdownDescription: "Service type (e.g., SSH, RDP, HTTP)",
							Required:            true,
						},
						"address": schema.StringAttribute{
							MarkdownDescription: "Service address",
							Optional:            true,
							Computed:            true,
						},
						"port": schema.Int64Attribute{
							MarkdownDescription: "Service port",
							Optional:            true,
							Computed:            true,
							Default:             int64default.StaticInt64(22),
						},
						"use_for_password_rotation": schema.BoolAttribute{
							MarkdownDescription: "Use this service for password rotation",
							Optional:            true,
							Computed:            true,
							Default:             booldefault.StaticBool(false),
						},
						"ssh_tunnel_port": schema.Int64Attribute{
							MarkdownDescription: "SSH tunnel port",
							Optional:            true,
							Computed:            true,
							Default:             int64default.StaticInt64(0),
						},
						"use_plaintext_vnc": schema.BoolAttribute{
							MarkdownDescription: "Use plaintext VNC",
							Optional:            true,
							Computed:            true,
							Default:             booldefault.StaticBool(false),
						},
						"source": schema.StringAttribute{
							MarkdownDescription: "Service source",
							Optional:            true,
							Computed:            true,
							Default:             stringdefault.StaticString("UI"),
						},
					},
				},
			},
			"principals": schema.ListNestedAttribute{
				MarkdownDescription: "List of principals for the host",
				Optional:            true,
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"principal": schema.StringAttribute{
							MarkdownDescription: "Principal name",
							Required:            true,
						},
						"passphrase": schema.StringAttribute{
							MarkdownDescription: "Principal passphrase (write-only, API returns masked value)",
							Optional:            true,
							Sensitive:           true,
							// no Default
							// no Computed (unless you really want to store something when omitted)
						},
						"rotate": schema.BoolAttribute{
							MarkdownDescription: "Whether to rotate the principal",
							Optional:            true,
							Computed:            true,
							Default:             booldefault.StaticBool(false),
						},
						"use_for_password_rotation": schema.BoolAttribute{
							MarkdownDescription: "Use this principal for password rotation",
							Optional:            true,
							Computed:            true,
							Default:             booldefault.StaticBool(false),
						},
						"username_attribute": schema.StringAttribute{
							MarkdownDescription: "Username attribute",
							Optional:            true,
							Computed:            true,
							Default:             stringdefault.StaticString(""),
						},
						"use_user_account": schema.BoolAttribute{
							MarkdownDescription: "Use user account",
							Optional:            true,
							Computed:            true,
							Default:             booldefault.StaticBool(false),
						},
						"source": schema.StringAttribute{
							MarkdownDescription: "Principal source",
							Optional:            true,
							Computed:            true,
							Default:             stringdefault.StaticString("UI"),
						},
						"roles": schema.ListNestedAttribute{
							MarkdownDescription: "List of roles for the principal",
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
						"applications": schema.ListAttribute{
							MarkdownDescription: "List of applications for the principal",
							ElementType:         types.StringType,
							Optional:            true,
							Computed:            true,
						},
						"service_options": schema.SingleNestedAttribute{
							MarkdownDescription: "Service options for the principal",
							Optional:            true,
							Computed:            true,
							Attributes: map[string]schema.Attribute{
								"ssh": schema.SingleNestedAttribute{
									MarkdownDescription: "SSH service options",
									Optional:            true,
									Computed:            true,
									Attributes: map[string]schema.Attribute{
										"shell": schema.BoolAttribute{
											MarkdownDescription: "Allow shell access",
											Optional:            true,
											Computed:            true,
											Default:             booldefault.StaticBool(true),
										},
										"file_transfer": schema.BoolAttribute{
											MarkdownDescription: "Allow file transfer",
											Optional:            true,
											Computed:            true,
											Default:             booldefault.StaticBool(true),
										},
										"exec": schema.BoolAttribute{
											MarkdownDescription: "Allow exec commands",
											Optional:            true,
											Computed:            true,
											Default:             booldefault.StaticBool(true),
										},
										"tunnels": schema.BoolAttribute{
											MarkdownDescription: "Allow tunnels",
											Optional:            true,
											Computed:            true,
											Default:             booldefault.StaticBool(true),
										},
										"x11": schema.BoolAttribute{
											MarkdownDescription: "Allow X11 forwarding",
											Optional:            true,
											Computed:            true,
											Default:             booldefault.StaticBool(true),
										},
										"other": schema.BoolAttribute{
											MarkdownDescription: "Allow other SSH features",
											Optional:            true,
											Computed:            true,
											Default:             booldefault.StaticBool(true),
										},
									},
								},
								"rdp": schema.SingleNestedAttribute{
									MarkdownDescription: "RDP service options",
									Optional:            true,
									Computed:            true,
									Attributes: map[string]schema.Attribute{
										"file_transfer": schema.BoolAttribute{
											MarkdownDescription: "Allow file transfer",
											Optional:            true,
											Computed:            true,
											Default:             booldefault.StaticBool(false),
										},
										"audio": schema.BoolAttribute{
											MarkdownDescription: "Allow audio",
											Optional:            true,
											Computed:            true,
											Default:             booldefault.StaticBool(false),
										},
										"clipboard": schema.BoolAttribute{
											MarkdownDescription: "Allow clipboard",
											Optional:            true,
											Computed:            true,
											Default:             booldefault.StaticBool(false),
										},
									},
								},
								"web": schema.SingleNestedAttribute{
									MarkdownDescription: "Web service options",
									Optional:            true,
									Computed:            true,
									Attributes: map[string]schema.Attribute{
										"file_transfer": schema.BoolAttribute{
											MarkdownDescription: "Allow file transfer",
											Optional:            true,
											Computed:            true,
											Default:             booldefault.StaticBool(false),
										},
										"audio": schema.BoolAttribute{
											MarkdownDescription: "Allow audio",
											Optional:            true,
											Computed:            true,
											Default:             booldefault.StaticBool(false),
										},
										"clipboard": schema.BoolAttribute{
											MarkdownDescription: "Allow clipboard",
											Optional:            true,
											Computed:            true,
											Default:             booldefault.StaticBool(false),
										},
									},
								},
								"vnc": schema.SingleNestedAttribute{
									MarkdownDescription: "VNC service options",
									Optional:            true,
									Computed:            true,
									Attributes: map[string]schema.Attribute{
										"file_transfer": schema.BoolAttribute{
											MarkdownDescription: "Allow file transfer",
											Optional:            true,
											Computed:            true,
											Default:             booldefault.StaticBool(false),
										},
										"clipboard": schema.BoolAttribute{
											MarkdownDescription: "Allow clipboard",
											Optional:            true,
											Computed:            true,
											Default:             booldefault.StaticBool(false),
										},
									},
								},
								"db": schema.SingleNestedAttribute{
									MarkdownDescription: "Database service options",
									Optional:            true,
									Computed:            true,
									Attributes: map[string]schema.Attribute{
										"max_bytes_upload": schema.Int64Attribute{
											MarkdownDescription: "Maximum bytes for upload",
											Optional:            true,
											Computed:            true,
											Default:             int64default.StaticInt64(0),
										},
										"max_bytes_download": schema.Int64Attribute{
											MarkdownDescription: "Maximum bytes for download",
											Optional:            true,
											Computed:            true,
											Default:             int64default.StaticInt64(0),
										},
									},
								},
							},
						},
						"command_restrictions": schema.SingleNestedAttribute{
							MarkdownDescription: "Command restrictions for the principal",
							Optional:            true,
							Computed:            true,
							Attributes: map[string]schema.Attribute{
								"enabled": schema.BoolAttribute{
									MarkdownDescription: "Enable command restrictions",
									Optional:            true,
									Computed:            true,
									Default:             booldefault.StaticBool(false),
								},
								"rshell_variant": schema.StringAttribute{
									MarkdownDescription: "Shell variant (e.g., bash, sh)",
									Optional:            true,
									Computed:            true,
									Default:             stringdefault.StaticString(""),
								},
								"default_whitelist": schema.SingleNestedAttribute{
									MarkdownDescription: "Default whitelist",
									Optional:            true,
									Computed:            true,
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											MarkdownDescription: "Whitelist ID",
											Optional:            true,
											Computed:            true,
											Default:             stringdefault.StaticString(""),
										},
										"name": schema.StringAttribute{
											MarkdownDescription: "Whitelist name",
											Optional:            true,
											Computed:            true,
											Default:             stringdefault.StaticString(""),
										},
									},
								},
								"whitelists": schema.ListNestedAttribute{
									MarkdownDescription: "List of whitelists with roles",
									Optional:            true,
									Computed:            true,
									NestedObject: schema.NestedAttributeObject{
										Attributes: map[string]schema.Attribute{
											"whitelist": schema.SingleNestedAttribute{
												MarkdownDescription: "Whitelist reference",
												Required:            true,
												Attributes: map[string]schema.Attribute{
													"id": schema.StringAttribute{
														MarkdownDescription: "Whitelist ID",
														Optional:            true,
														Computed:            true,
														Default:             stringdefault.StaticString(""),
													},
													"name": schema.StringAttribute{
														MarkdownDescription: "Whitelist name",
														Optional:            true,
														Computed:            true,
														Default:             stringdefault.StaticString(""),
													},
												},
											},
											"roles": schema.ListNestedAttribute{
												MarkdownDescription: "Roles for this whitelist",
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
										},
									},
								},
								"allow_no_match": schema.BoolAttribute{
									MarkdownDescription: "Allow commands that don't match any whitelist",
									Optional:            true,
									Computed:            true,
									Default:             booldefault.StaticBool(false),
								},
								"audit_match": schema.BoolAttribute{
									MarkdownDescription: "Audit commands that match whitelists",
									Optional:            true,
									Computed:            true,
									Default:             booldefault.StaticBool(false),
								},
								"audit_no_match": schema.BoolAttribute{
									MarkdownDescription: "Audit commands that don't match whitelists",
									Optional:            true,
									Computed:            true,
									Default:             booldefault.StaticBool(false),
								},
								"banner": schema.StringAttribute{
									MarkdownDescription: "Banner message to display",
									Optional:            true,
									Computed:            true,
									Default:             stringdefault.StaticString(""),
								},
							},
						},
					},
				},
			},
			"ssh_host_public_keys": schema.ListNestedAttribute{
				MarkdownDescription: "List of SSH host public keys",
				Optional:            true,
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"key": schema.StringAttribute{
							MarkdownDescription: "SSH public key",
							Required:            true,
						},
						"fingerprint": schema.StringAttribute{
							MarkdownDescription: "SSH key fingerprint",
							Optional:            true,
							Computed:            true,
						},
					},
				},
			},
			"session_recording_options": schema.SingleNestedAttribute{
				MarkdownDescription: "Session recording options",
				Optional:            true,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"disable_clipboard_recording": schema.BoolAttribute{
						MarkdownDescription: "Disable clipboard recording",
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(false),
					},
					"disable_file_transfer_recording": schema.BoolAttribute{
						MarkdownDescription: "Disable file transfer recording",
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(false),
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

func (r *HostResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	tflog.Debug(ctx, "Creating hoststore client", map[string]interface{}{
		"connector": fmt.Sprintf("%+v", *connector),
	})

	r.client = hoststore.New(*connector)
}

func (r *HostResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *HostResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	/*tflog.Debug(ctx, "Loaded host data", map[string]interface{}{
		"data": fmt.Sprintf("%+v", data),
	})*/

	var pr PasswordRotationModel
	if !data.PasswordRotation.IsNull() && !data.PasswordRotation.IsUnknown() {
		diags := data.PasswordRotation.As(ctx, &pr, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Convert addresses list to string slice
	var addresses []string
	if !data.Addresses.IsNull() && !data.Addresses.IsUnknown() {
		data.Addresses.ElementsAs(ctx, &addresses, false)
	}

	// Convert tags list to string slice and sort for consistency
	var tags []string
	if !data.Tags.IsNull() && !data.Tags.IsUnknown() {
		data.Tags.ElementsAs(ctx, &tags, false)
		sort.Strings(tags) // Sort to ensure consistent ordering
	}

	// Convert services
	var services []hoststore.HostService
	for _, sm := range data.Services {
		service := hoststore.HostService{
			Service:                sm.Service.ValueString(),
			Address:                sm.Address.ValueString(),
			Port:                   int(sm.Port.ValueInt64()),
			UseForPasswordRotation: sm.UseForPasswordRotation.ValueBool(),
			TunnelPort:             int(sm.SSHTunnelPort.ValueInt64()),
			UsePlainTextVNC:        sm.UsePlaintextVNC.ValueBool(),
			Source:                 sm.Source.ValueString(),
		}
		services = append(services, service)
	}

	// Convert principals
	var principals []hoststore.HostPrincipals
	for _, pm := range data.Principals {
		principal := hoststore.HostPrincipals{
			Principal:              pm.Principal.ValueString(),
			Rotate:                 pm.Rotate.ValueBool(),
			UseForPasswordRotation: pm.UseForPasswordRotation.ValueBool(),
			UsernameAttribute:      pm.UsernameAttribute.ValueString(),
			UseUserAccount:         pm.UseUserAccount.ValueBool(),
			Source:                 pm.Source.ValueString(),
		}

		if !isUnsetString(pm.Passphrase) {
			principal.Passphrase = pm.Passphrase.ValueString()
		}

		// Convert roles
		if !pm.Roles.IsNull() && !pm.Roles.IsUnknown() {
			var roleModels []RoleModel
			pm.Roles.ElementsAs(ctx, &roleModels, false)

			for _, rm := range roleModels {
				role := hoststore.HostRole{
					ID:   rm.ID.ValueString(),
					Name: rm.Name.ValueString(),
				}
				principal.Roles = append(principal.Roles, role)
			}
		}

		// Convert applications
		if !pm.Applications.IsNull() && !pm.Applications.IsUnknown() {
			var apps []string
			pm.Applications.ElementsAs(ctx, &apps, false)
			for _, app := range apps {
				principal.Applications = append(principal.Applications, hoststore.HostPrincipalApplications{
					Name: app,
				})
			}
		}

		// Convert service options
		if !pm.ServiceOptions.IsNull() && !pm.ServiceOptions.IsUnknown() {
			var serviceOptionsModel ServiceOptionsModel
			pm.ServiceOptions.As(ctx, &serviceOptionsModel, basetypes.ObjectAsOptions{})

			serviceOptions := &hoststore.HostServiceOptions{}

			// SSH options
			if !serviceOptionsModel.SSH.IsNull() && !serviceOptionsModel.SSH.IsUnknown() {
				var sshModel SSHOptionsModel
				serviceOptionsModel.SSH.As(ctx, &sshModel, basetypes.ObjectAsOptions{})
				serviceOptions.SSHServiceOptions = &hoststore.SSHServiceOptions{
					Shell:        sshModel.Shell.ValueBool(),
					FileTransfer: sshModel.FileTransfer.ValueBool(),
					Exec:         sshModel.Exec.ValueBool(),
					Tunnels:      sshModel.Tunnels.ValueBool(),
					X11:          sshModel.X11.ValueBool(),
					Other:        sshModel.Other.ValueBool(),
				}
			}

			// RDP options
			if !serviceOptionsModel.RDP.IsNull() && !serviceOptionsModel.RDP.IsUnknown() {
				var rdpModel RDPOptionsModel
				serviceOptionsModel.RDP.As(ctx, &rdpModel, basetypes.ObjectAsOptions{})
				serviceOptions.RDPServiceOptions = &hoststore.RDPServiceOptions{
					FileTransfer: rdpModel.FileTransfer.ValueBool(),
					Audio:        rdpModel.Audio.ValueBool(),
					Clipboard:    rdpModel.Clipboard.ValueBool(),
				}
			}

			// Web options
			if !serviceOptionsModel.Web.IsNull() && !serviceOptionsModel.Web.IsUnknown() {
				var webModel WebOptionsModel
				serviceOptionsModel.Web.As(ctx, &webModel, basetypes.ObjectAsOptions{})
				serviceOptions.WebServiceOptions = &hoststore.WebServiceOptions{
					FileTransfer: webModel.FileTransfer.ValueBool(),
					Audio:        webModel.Audio.ValueBool(),
					Clipboard:    webModel.Clipboard.ValueBool(),
				}
			}

			// VNC options
			if !serviceOptionsModel.VNC.IsNull() && !serviceOptionsModel.VNC.IsUnknown() {
				var vncModel VNCOptionsModel
				serviceOptionsModel.VNC.As(ctx, &vncModel, basetypes.ObjectAsOptions{})
				serviceOptions.VNCServiceOptions = &hoststore.VNCServiceOptions{
					FileTransfer: vncModel.FileTransfer.ValueBool(),
					Clipboard:    vncModel.Clipboard.ValueBool(),
				}
			}

			// DB options
			if !serviceOptionsModel.DB.IsNull() && !serviceOptionsModel.DB.IsUnknown() {
				var dbModel DBOptionsModel
				serviceOptionsModel.DB.As(ctx, &dbModel, basetypes.ObjectAsOptions{})
				serviceOptions.DBServiceOptions = &hoststore.DBServiceOptions{
					MaxBytesUpload:   dbModel.MaxBytesUpload.ValueInt64(),
					MaxBytesDownload: dbModel.MaxBytesDownload.ValueInt64(),
				}
			}

			principal.ServiceOptions = serviceOptions
		}

		// Convert command restrictions
		if !pm.CommandRestrictions.IsNull() && !pm.CommandRestrictions.IsUnknown() {
			var commandRestrictionsModel CommandRestrictionsModel
			pm.CommandRestrictions.As(ctx, &commandRestrictionsModel, basetypes.ObjectAsOptions{})

			commandRestrictions := hoststore.HostCommandRestrictions{
				Enabled:       commandRestrictionsModel.Enabled.ValueBool(),
				RShellVariant: commandRestrictionsModel.RShellVariant.ValueString(),
				AllowNoMatch:  commandRestrictionsModel.AllowNoMatch.ValueBool(),
				AuditMatch:    commandRestrictionsModel.AuditMatch.ValueBool(),
				AuditNoMatch:  commandRestrictionsModel.AuditNoMatch.ValueBool(),
				Banner:        commandRestrictionsModel.Banner.ValueString(),
			}

			// Convert default whitelist
			if !commandRestrictionsModel.DefaultWhitelist.IsNull() && !commandRestrictionsModel.DefaultWhitelist.IsUnknown() {
				var defaultWhitelistModel WhitelistHandleModel
				commandRestrictionsModel.DefaultWhitelist.As(ctx, &defaultWhitelistModel, basetypes.ObjectAsOptions{})
				commandRestrictions.DefaultWhiteList = hoststore.WhiteListHandle{
					ID:   defaultWhitelistModel.ID.ValueString(),
					Name: defaultWhitelistModel.Name.ValueString(),
				}
			}

			// Convert whitelists
			if !commandRestrictionsModel.Whitelists.IsNull() && !commandRestrictionsModel.Whitelists.IsUnknown() {
				var whitelistGrantModels []WhitelistGrantModel
				commandRestrictionsModel.Whitelists.ElementsAs(ctx, &whitelistGrantModels, false)

				for _, wgm := range whitelistGrantModels {
					whitelistGrant := hoststore.WhiteListGrant{}

					// Convert whitelist handle
					if !wgm.Whitelist.IsNull() && !wgm.Whitelist.IsUnknown() {
						var whitelistHandleModel WhitelistHandleModel
						wgm.Whitelist.As(ctx, &whitelistHandleModel, basetypes.ObjectAsOptions{})
						whitelistGrant.WhiteList = hoststore.WhiteListHandle{
							ID:   whitelistHandleModel.ID.ValueString(),
							Name: whitelistHandleModel.Name.ValueString(),
						}
					}

					// Convert roles
					if !wgm.Roles.IsNull() && !wgm.Roles.IsUnknown() {
						var roleModels []RoleModel
						wgm.Roles.ElementsAs(ctx, &roleModels, false)

						for _, rm := range roleModels {
							role := hoststore.HostRole{
								ID:   rm.ID.ValueString(),
								Name: rm.Name.ValueString(),
							}
							whitelistGrant.Roles = append(whitelistGrant.Roles, role)
						}
					}

					commandRestrictions.WhiteLists = append(commandRestrictions.WhiteLists, whitelistGrant)
				}
			}

			principal.CommandRestrictions = commandRestrictions
		}

		principals = append(principals, principal)
	}

	// Convert SSH host public keys
	var sshHostPublicKeys []hoststore.HostSSHPubKeys
	if !data.SSHHostPublicKeys.IsNull() && !data.SSHHostPublicKeys.IsUnknown() {
		var keyModels []SSHHostPublicKeyModel
		data.SSHHostPublicKeys.ElementsAs(ctx, &keyModels, false)

		for _, km := range keyModels {
			key := hoststore.HostSSHPubKeys{
				Key:         km.Key.ValueString(),
				FingerPrint: km.Fingerprint.ValueString(),
			}
			sshHostPublicKeys = append(sshHostPublicKeys, key)
		}
	}

	// Convert session recording options
	var sessionRecordingOptions *hoststore.SessionRecordingOptions
	if !data.SessionRecordingOptions.IsNull() && !data.SessionRecordingOptions.IsUnknown() {
		var sroModel SessionRecordingOptionsModel
		diags := data.SessionRecordingOptions.As(ctx, &sroModel, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}

		sessionRecordingOptions = &hoststore.SessionRecordingOptions{
			DisableClipboardRecording:    sroModel.DisableClipboardRecording.ValueBool(),
			DisableFileTransferRecording: sroModel.DisableFileTransferRecording.ValueBool(),
		}
	}

	deployable := data.Deployable.ValueBool()
	tofu := data.Tofu.ValueBool()
	toch := data.Toch.ValueBool()
	auditEnabled := data.AuditEnabled.ValueBool()

	contactAddress := ""
	if !data.ContactAddress.IsNull() && !data.ContactAddress.IsUnknown() {
		contactAddress = data.ContactAddress.ValueString()
	}
	if contactAddress == "" && len(addresses) > 0 {
		contactAddress = addresses[0]
	}

	host := hoststore.Host{
		CommonName:    data.CommonName.ValueString(),
		Addresses:     addresses,
		AccessGroupID: data.AccessGroupID.ValueString(),

		// leave optional strings out here; set them conditionally below

		Deployable:              &deployable,
		Tofu:                    &tofu,
		Toch:                    &toch,
		AuditEnabled:            &auditEnabled,
		ContactAddress:          contactAddress,
		Tags:                    tags,
		Services:                services,
		Principals:              principals,
		SSHHostPubKeys:          sshHostPublicKeys,
		SessionRecordingOptions: sessionRecordingOptions,
	}

	enabled := boolOrFalse(data.PasswordRotationEnabled)
	host.PasswordRotationEnabled = enabled

	if enabled {
		rm := expandPasswordRotation(pr)
		if rm == nil {
			resp.Diagnostics.AddError(
				"Invalid password rotation configuration",
				"password_rotation_enabled=true requires password_rotation block with required fields (use_main_account, operating_system, protocol, password_policy_id, script_template_id).",
			)
			return
		}
		host.PasswordRotation = rm
	} else {
		host.PasswordRotation = nil
	}

	// only set if user provided
	if !isUnsetString(data.SourceID) {
		host.SourceID = data.SourceID.ValueString()
	}
	if !isUnsetString(data.ExternalID) {
		host.ExternalID = data.ExternalID.ValueString()
	}
	if !isUnsetString(data.InstanceID) {
		host.InstanceID = data.InstanceID.ValueString()
	}
	if !isUnsetString(data.CloudProvider) {
		host.CloudProvider = data.CloudProvider.ValueString()
	}
	if !isUnsetString(data.CloudProviderRegion) {
		host.CloudProviderRegion = data.CloudProviderRegion.ValueString()
	}
	if !isUnsetString(data.DistinguishedName) {
		host.DistinguishedName = data.DistinguishedName.ValueString()
	}
	if !isUnsetString(data.Organization) {
		host.Organization = data.Organization.ValueString()
	}
	if !isUnsetString(data.OrganizationalUnit) {
		host.OrganizationalUnit = data.OrganizationalUnit.ValueString()
	}
	if !isUnsetString(data.Zone) {
		host.Zone = data.Zone.ValueString()
	}
	if !isUnsetString(data.HostType) {
		host.HostType = data.HostType.ValueString()
	}
	if !isUnsetString(data.HostClassification) {
		host.HostClassification = data.HostClassification.ValueString()
	}
	if !isUnsetString(data.Comment) {
		host.Comment = data.Comment.ValueString()
	}
	if !isUnsetString(data.UserMessage) {
		host.UserMessage = data.UserMessage.ValueString()
	}
	if !isUnsetString(data.Disabled) {
		host.Disabled = data.Disabled.ValueString()
	}

	if data.PasswordRotationEnabled.ValueBool() && host.PasswordRotation == nil {
		resp.Diagnostics.AddError(
			"Invalid password rotation configuration",
			"password_rotation_enabled=true requires password_rotation block with required fields (use_main_account, operating_system, protocol, password_policy_id, script_template_id).",
		)
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("hoststore.Host model used: %+v", host))
	createdHost, err := r.client.CreateHost(&host)
	if err != nil {
		tflog.Error(ctx, "CreateHost failed", map[string]any{
			"err_type": fmt.Sprintf("%T", err),
			"err":      fmt.Sprintf("%+v", err),
			"err_str":  err.Error(),
			// also dump key fields
			"source_id":       host.SourceID,
			"access_group_id": host.AccessGroupID,
			"services_len":    len(host.Services),
			"principals_len":  len(host.Principals),
		})

		resp.Diagnostics.AddError(
			"Unable to Create Host",
			fmt.Sprintf("CreateHost failed: %T: %+v", err, err),
		)
		return
	}

	data.ID = types.StringValue(createdHost.ID)

	// Read back the created resource to populate all computed fields
	hostRead, err := r.client.GetHost(createdHost.ID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read created host, got error: %s", err))
		return
	}

	// Populate all the computed fields from the API response
	r.populateHostModel(ctx, data, hostRead)

	tflog.Debug(ctx, "created host resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *HostResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *HostResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	host, err := r.client.GetHost(data.ID.ValueString())
	if err != nil {
		tflog.Debug(ctx, "Error reading host", map[string]interface{}{
			"id":    data.ID.ValueString(),
			"error": err.Error(),
		})

		if utils.IsPrivxNotFound(err) {
			tflog.Info(ctx, "Host resource appears to be deleted, removing from state", map[string]interface{}{
				"id": data.ID.ValueString(),
			})
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read host, got error: %s", err))
		return
	}

	r.populateHostModel(ctx, data, host)

	tflog.Debug(ctx, "Storing host into the state", map[string]interface{}{
		"state": fmt.Sprintf("%+v", data),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *HostResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *HostResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var pr PasswordRotationModel
	if !data.PasswordRotation.IsNull() && !data.PasswordRotation.IsUnknown() {
		diags := data.PasswordRotation.As(ctx, &pr, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	currentHost, err := r.client.GetHost(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read current host, got error: %s", err))
		return
	}

	// Update the host with new values
	currentHost.CommonName = data.CommonName.ValueString()
	currentHost.CommonName = data.CommonName.ValueString()

	// Only overwrite if user explicitly set it in config/plan
	if !isUnsetString(data.ExternalID) {
		currentHost.ExternalID = data.ExternalID.ValueString()
	}
	if !isUnsetString(data.InstanceID) {
		currentHost.InstanceID = data.InstanceID.ValueString()
	}

	currentHost.AccessGroupID = data.AccessGroupID.ValueString()
	currentHost.CloudProvider = data.CloudProvider.ValueString()
	currentHost.CloudProviderRegion = data.CloudProviderRegion.ValueString()
	currentHost.DistinguishedName = data.DistinguishedName.ValueString()
	currentHost.Organization = data.Organization.ValueString()
	currentHost.OrganizationalUnit = data.OrganizationalUnit.ValueString()
	currentHost.Zone = data.Zone.ValueString()
	currentHost.HostType = data.HostType.ValueString()
	currentHost.HostClassification = data.HostClassification.ValueString()
	if !isUnsetString(data.Comment) {
		currentHost.Comment = data.Comment.ValueString()
	}
	currentHost.UserMessage = data.UserMessage.ValueString()
	currentHost.Disabled = data.Disabled.ValueString()
	deployable := data.Deployable.ValueBool()
	tofu := data.Tofu.ValueBool()
	toch := data.Toch.ValueBool()
	auditEnabled := data.AuditEnabled.ValueBool()
	currentHost.Deployable = &deployable
	currentHost.Tofu = &tofu
	currentHost.Toch = &toch
	currentHost.AuditEnabled = &auditEnabled
	// --- Password rotation handling ---
	// Only change enabled flag if it was explicitly set in config/plan.
	if boolIsKnown(data.PasswordRotationEnabled) {
		enabled := data.PasswordRotationEnabled.ValueBool()
		currentHost.PasswordRotationEnabled = enabled

		if enabled {
			rm := expandPasswordRotation(pr)
			if rm == nil {
				resp.Diagnostics.AddError(
					"Invalid password rotation configuration",
					"password_rotation_enabled=true requires password_rotation block with required fields (use_main_account, operating_system, protocol, password_policy_id, script_template_id).",
				)
				return
			}
			currentHost.PasswordRotation = rm
		} else {
			// Disable rotation -> clear metadata to avoid server-side validation / stale config
			currentHost.PasswordRotation = nil
		}
	}

	// else keep currentHost.PasswordRotation as-is.

	if !isUnsetString(data.SourceID) {
		currentHost.SourceID = data.SourceID.ValueString()
	}

	currentHost.ContactAddress = data.ContactAddress.ValueString()

	// Convert addresses list to string slice
	var addresses []string
	if !data.Addresses.IsNull() && !data.Addresses.IsUnknown() {
		data.Addresses.ElementsAs(ctx, &addresses, false)
	}
	contactAddress := ""
	if !data.ContactAddress.IsNull() && !data.ContactAddress.IsUnknown() {
		contactAddress = data.ContactAddress.ValueString()
	}
	if contactAddress == "" && len(addresses) > 0 {
		contactAddress = addresses[0]
	}
	currentHost.ContactAddress = contactAddress

	// Convert tags list to string slice and sort for consistency
	var tags []string
	if !data.Tags.IsNull() && !data.Tags.IsUnknown() {
		data.Tags.ElementsAs(ctx, &tags, false)
		sort.Strings(tags) // Sort to ensure consistent ordering
	}
	currentHost.Tags = tags

	// Convert services
	var services []hoststore.HostService
	for _, sm := range data.Services {
		service := hoststore.HostService{
			Service:                sm.Service.ValueString(),
			Address:                sm.Address.ValueString(),
			Port:                   int(sm.Port.ValueInt64()),
			UseForPasswordRotation: sm.UseForPasswordRotation.ValueBool(),
			TunnelPort:             int(sm.SSHTunnelPort.ValueInt64()),
			UsePlainTextVNC:        sm.UsePlaintextVNC.ValueBool(),
			Source:                 sm.Source.ValueString(),
		}
		services = append(services, service)
	}
	currentHost.Services = services

	// Convert principals
	var principals []hoststore.HostPrincipals
	for _, pm := range data.Principals {
		principal := hoststore.HostPrincipals{
			Principal:              pm.Principal.ValueString(),
			Rotate:                 pm.Rotate.ValueBool(),
			UseForPasswordRotation: pm.UseForPasswordRotation.ValueBool(),
			UsernameAttribute:      pm.UsernameAttribute.ValueString(),
			UseUserAccount:         pm.UseUserAccount.ValueBool(),
			Source:                 pm.Source.ValueString(),
		}

		if !isUnsetString(pm.Passphrase) {
			principal.Passphrase = pm.Passphrase.ValueString()
		}

		// Convert roles
		if !pm.Roles.IsNull() && !pm.Roles.IsUnknown() {
			var roleModels []RoleModel
			pm.Roles.ElementsAs(ctx, &roleModels, false)

			for _, rm := range roleModels {
				role := hoststore.HostRole{
					ID:   rm.ID.ValueString(),
					Name: rm.Name.ValueString(),
				}
				principal.Roles = append(principal.Roles, role)
			}
		}

		// Convert applications
		if !pm.Applications.IsNull() && !pm.Applications.IsUnknown() {
			var apps []string
			pm.Applications.ElementsAs(ctx, &apps, false)
			for _, app := range apps {
				principal.Applications = append(principal.Applications, hoststore.HostPrincipalApplications{
					Name: app,
				})
			}
		}

		// Convert service options
		if !pm.ServiceOptions.IsNull() && !pm.ServiceOptions.IsUnknown() {
			var serviceOptionsModel ServiceOptionsModel
			pm.ServiceOptions.As(ctx, &serviceOptionsModel, basetypes.ObjectAsOptions{})

			serviceOptions := &hoststore.HostServiceOptions{}

			// SSH options
			if !serviceOptionsModel.SSH.IsNull() && !serviceOptionsModel.SSH.IsUnknown() {
				var sshModel SSHOptionsModel
				serviceOptionsModel.SSH.As(ctx, &sshModel, basetypes.ObjectAsOptions{})
				serviceOptions.SSHServiceOptions = &hoststore.SSHServiceOptions{
					Shell:        sshModel.Shell.ValueBool(),
					FileTransfer: sshModel.FileTransfer.ValueBool(),
					Exec:         sshModel.Exec.ValueBool(),
					Tunnels:      sshModel.Tunnels.ValueBool(),
					X11:          sshModel.X11.ValueBool(),
					Other:        sshModel.Other.ValueBool(),
				}
			}

			// RDP options
			if !serviceOptionsModel.RDP.IsNull() && !serviceOptionsModel.RDP.IsUnknown() {
				var rdpModel RDPOptionsModel
				serviceOptionsModel.RDP.As(ctx, &rdpModel, basetypes.ObjectAsOptions{})
				serviceOptions.RDPServiceOptions = &hoststore.RDPServiceOptions{
					FileTransfer: rdpModel.FileTransfer.ValueBool(),
					Audio:        rdpModel.Audio.ValueBool(),
					Clipboard:    rdpModel.Clipboard.ValueBool(),
				}
			}

			// Web options
			if !serviceOptionsModel.Web.IsNull() && !serviceOptionsModel.Web.IsUnknown() {
				var webModel WebOptionsModel
				serviceOptionsModel.Web.As(ctx, &webModel, basetypes.ObjectAsOptions{})
				serviceOptions.WebServiceOptions = &hoststore.WebServiceOptions{
					FileTransfer: webModel.FileTransfer.ValueBool(),
					Audio:        webModel.Audio.ValueBool(),
					Clipboard:    webModel.Clipboard.ValueBool(),
				}
			}

			// VNC options
			if !serviceOptionsModel.VNC.IsNull() && !serviceOptionsModel.VNC.IsUnknown() {
				var vncModel VNCOptionsModel
				serviceOptionsModel.VNC.As(ctx, &vncModel, basetypes.ObjectAsOptions{})
				serviceOptions.VNCServiceOptions = &hoststore.VNCServiceOptions{
					FileTransfer: vncModel.FileTransfer.ValueBool(),
					Clipboard:    vncModel.Clipboard.ValueBool(),
				}
			}

			// DB options
			if !serviceOptionsModel.DB.IsNull() && !serviceOptionsModel.DB.IsUnknown() {
				var dbModel DBOptionsModel
				serviceOptionsModel.DB.As(ctx, &dbModel, basetypes.ObjectAsOptions{})
				serviceOptions.DBServiceOptions = &hoststore.DBServiceOptions{
					MaxBytesUpload:   dbModel.MaxBytesUpload.ValueInt64(),
					MaxBytesDownload: dbModel.MaxBytesDownload.ValueInt64(),
				}
			}

			principal.ServiceOptions = serviceOptions
		}

		// Convert command restrictions
		if !pm.CommandRestrictions.IsNull() && !pm.CommandRestrictions.IsUnknown() {
			var commandRestrictionsModel CommandRestrictionsModel
			pm.CommandRestrictions.As(ctx, &commandRestrictionsModel, basetypes.ObjectAsOptions{})

			commandRestrictions := hoststore.HostCommandRestrictions{
				Enabled:       commandRestrictionsModel.Enabled.ValueBool(),
				RShellVariant: commandRestrictionsModel.RShellVariant.ValueString(),
				AllowNoMatch:  commandRestrictionsModel.AllowNoMatch.ValueBool(),
				AuditMatch:    commandRestrictionsModel.AuditMatch.ValueBool(),
				AuditNoMatch:  commandRestrictionsModel.AuditNoMatch.ValueBool(),
				Banner:        commandRestrictionsModel.Banner.ValueString(),
			}

			// Convert default whitelist
			if !commandRestrictionsModel.DefaultWhitelist.IsNull() && !commandRestrictionsModel.DefaultWhitelist.IsUnknown() {
				var defaultWhitelistModel WhitelistHandleModel
				commandRestrictionsModel.DefaultWhitelist.As(ctx, &defaultWhitelistModel, basetypes.ObjectAsOptions{})
				commandRestrictions.DefaultWhiteList = hoststore.WhiteListHandle{
					ID:   defaultWhitelistModel.ID.ValueString(),
					Name: defaultWhitelistModel.Name.ValueString(),
				}
			}

			// Convert whitelists
			if !commandRestrictionsModel.Whitelists.IsNull() && !commandRestrictionsModel.Whitelists.IsUnknown() {
				var whitelistGrantModels []WhitelistGrantModel
				commandRestrictionsModel.Whitelists.ElementsAs(ctx, &whitelistGrantModels, false)

				for _, wgm := range whitelistGrantModels {
					whitelistGrant := hoststore.WhiteListGrant{}

					// Convert whitelist handle
					if !wgm.Whitelist.IsNull() && !wgm.Whitelist.IsUnknown() {
						var whitelistHandleModel WhitelistHandleModel
						wgm.Whitelist.As(ctx, &whitelistHandleModel, basetypes.ObjectAsOptions{})
						whitelistGrant.WhiteList = hoststore.WhiteListHandle{
							ID:   whitelistHandleModel.ID.ValueString(),
							Name: whitelistHandleModel.Name.ValueString(),
						}
					}

					// Convert roles
					if !wgm.Roles.IsNull() && !wgm.Roles.IsUnknown() {
						var roleModels []RoleModel
						wgm.Roles.ElementsAs(ctx, &roleModels, false)

						for _, rm := range roleModels {
							role := hoststore.HostRole{
								ID:   rm.ID.ValueString(),
								Name: rm.Name.ValueString(),
							}
							whitelistGrant.Roles = append(whitelistGrant.Roles, role)
						}
					}

					commandRestrictions.WhiteLists = append(commandRestrictions.WhiteLists, whitelistGrant)
				}
			}

			principal.CommandRestrictions = commandRestrictions
		}

		principals = append(principals, principal)
	}
	currentHost.Principals = principals

	// Convert SSH host public keys
	var sshHostPublicKeys []hoststore.HostSSHPubKeys
	if !data.SSHHostPublicKeys.IsNull() && !data.SSHHostPublicKeys.IsUnknown() {
		var keyModels []SSHHostPublicKeyModel
		data.SSHHostPublicKeys.ElementsAs(ctx, &keyModels, false)

		for _, km := range keyModels {
			key := hoststore.HostSSHPubKeys{
				Key:         km.Key.ValueString(),
				FingerPrint: km.Fingerprint.ValueString(),
			}
			sshHostPublicKeys = append(sshHostPublicKeys, key)
		}
	}
	currentHost.SSHHostPubKeys = sshHostPublicKeys

	// Convert session recording options
	if !data.SessionRecordingOptions.IsNull() && !data.SessionRecordingOptions.IsUnknown() {
		var sroModel SessionRecordingOptionsModel
		diags := data.SessionRecordingOptions.As(ctx, &sroModel, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}

		currentHost.SessionRecordingOptions = &hoststore.SessionRecordingOptions{
			DisableClipboardRecording:    sroModel.DisableClipboardRecording.ValueBool(),
			DisableFileTransferRecording: sroModel.DisableFileTransferRecording.ValueBool(),
		}
	}

	tflog.Debug(ctx, fmt.Sprintf("hoststore.Host model used for update: %+v", currentHost))

	tflog.Debug(ctx, "UpdateHost payload", map[string]any{
		"id":                        data.ID.ValueString(),
		"password_rotation_enabled": currentHost.PasswordRotationEnabled,
		"password_rotation_is_nil":  currentHost.PasswordRotation == nil,
		"password_rotation":         fmt.Sprintf("%+v", currentHost.PasswordRotation),
		"source_id":                 currentHost.SourceID,
		"access_group_id":           currentHost.AccessGroupID,
		"services_len":              len(currentHost.Services),
		"principals_len":            len(currentHost.Principals),
	})

	err = r.client.UpdateHost(data.ID.ValueString(), currentHost)

	if err != nil {
		tflog.Error(ctx, "UpdateHost failed", map[string]any{
			"err_type": fmt.Sprintf("%T", err),
			"err":      fmt.Sprintf("%+v", err),
			"err_str":  err.Error(),
		})

		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to update host, got error: %T: %+v", err, err),
		)
		return
	}

	// Read back the updated resource to populate all computed fields
	hostRead, err := r.client.GetHost(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read updated host, got error: %s", err))
		return
	}

	r.populateHostModel(ctx, data, hostRead)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *HostResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *HostResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteHost(data.ID.ValueString())
	if err != nil {
		if utils.IsPrivxNotFound(err) {
			return
		}
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to delete host, got error: %s", err),
		)
		return
	}

}

func (r *HostResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// populateHostModel populates the Terraform model from the API response.
func (r *HostResource) populateHostModel(ctx context.Context, data *HostResourceModel, host *hoststore.Host) {

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
	prTypes := map[string]attr.Type{
		"access_group_id":                      types.StringType,
		"use_main_account":                     types.BoolType,
		"operating_system":                     types.StringType,
		"protocol":                             types.StringType,
		"password_policy_id":                   types.StringType,
		"script_template_id":                   types.StringType,
		"certificate_validation_options":       types.StringType,
		"winrm_host_certificate_trust_anchors": types.StringType,
		"winrm_address":                        types.StringType,
		"winrm_port":                           types.Int64Type,
	}

	if host.PasswordRotation == nil {
		data.PasswordRotation = types.ObjectNull(prTypes)
	} else {
		prVals := map[string]attr.Value{
			"access_group_id":                      stringOrNull(host.PasswordRotation.AccessGroupID),
			"use_main_account":                     types.BoolValue(host.PasswordRotation.UseMainAccount),
			"operating_system":                     types.StringValue(host.PasswordRotation.OperatingSystem),
			"protocol":                             types.StringValue(host.PasswordRotation.Protocol),
			"password_policy_id":                   stringOrNull(host.PasswordRotation.PasswordPolicyId),
			"script_template_id":                   stringOrNull(host.PasswordRotation.ScriptTemplateId),
			"certificate_validation_options":       certValidationOrDefault(host.PasswordRotation.CertificateValidationOptions),
			"winrm_host_certificate_trust_anchors": stringOrNull(host.PasswordRotation.WinRMHostCertificateTrustAnchors),
			"winrm_address":                        stringOrNull(host.PasswordRotation.WinrmAddress),
			"winrm_port":                           types.Int64Value(int64(host.PasswordRotation.WinrmPort)),
		}

		data.PasswordRotation = types.ObjectValueMust(prTypes, prVals)
	}

	data.StandAloneHost = types.BoolValue(host.StandAloneHost)
	data.ContactAddress = types.StringValue(host.ContactAddress)

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

	// Convert tags slice to list - preserve original order if possible, otherwise sort
	var finalTags []string
	if !data.Tags.IsNull() && !data.Tags.IsUnknown() {
		// Try to preserve the original configuration order
		var originalTags []string
		data.Tags.ElementsAs(ctx, &originalTags, false)

		// Check if the API response contains the same tags (ignoring order)
		if tagsContainSameElements(originalTags, host.Tags) {
			// Use original order if the tags are the same
			finalTags = originalTags
		} else {
			// Different tags, use API response sorted
			finalTags = make([]string, len(host.Tags))
			copy(finalTags, host.Tags)
			sort.Strings(finalTags)
		}
	} else {
		// No original tags, use API response sorted
		finalTags = make([]string, len(host.Tags))
		copy(finalTags, host.Tags)
		sort.Strings(finalTags)
	}

	tagValues := make([]attr.Value, len(finalTags))
	for i, tag := range finalTags {
		tagValues[i] = types.StringValue(tag)
	}
	data.Tags = types.ListValueMust(types.StringType, tagValues)

	// Convert services - use API response directly to avoid consistency issues
	data.Services = make([]ServiceModel, len(host.Services))
	for i, service := range host.Services {
		data.Services[i] = ServiceModel{
			Service:                types.StringValue(service.Service),
			Address:                types.StringValue(service.Address),
			Port:                   types.Int64Value(int64(service.Port)),
			UseForPasswordRotation: types.BoolValue(service.UseForPasswordRotation),
			SSHTunnelPort:          types.Int64Value(int64(service.TunnelPort)),
			UsePlaintextVNC:        types.BoolValue(service.UsePlainTextVNC),
			Source:                 types.StringValue(service.Source),
		}
	}

	// Convert principals - preserve original passphrase values to avoid sensitive attribute issues
	originalPrincipals := data.Principals // Save original principals
	data.Principals = make([]PrincipalModel, len(host.Principals))

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

		// Find the original passphrase value by matching principal name
		// Since API returns masked value, preserve original to avoid showing changes
		passphraseValue := types.StringValue("") // Default
		for _, origPrincipal := range originalPrincipals {
			if origPrincipal.Principal.ValueString() == principal.Principal {
				// Preserve original passphrase since API returns masked value
				passphraseValue = origPrincipal.Passphrase
				break
			}
		}

		// Convert service options for this principal
		var serviceOptionsValue types.Object
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

		// Convert command restrictions for this principal
		var commandRestrictionsValue types.Object
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

		commandRestrictionsValue = types.ObjectValueMust(map[string]attr.Type{
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

		data.Principals[i] = PrincipalModel{
			Principal:              types.StringValue(principal.Principal),
			Passphrase:             passphraseValue, // Preserve original passphrase
			Rotate:                 types.BoolValue(principal.Rotate),
			UseForPasswordRotation: types.BoolValue(principal.UseForPasswordRotation),
			UsernameAttribute:      types.StringValue(principal.UsernameAttribute),
			UseUserAccount:         types.BoolValue(principal.UseUserAccount),
			Source:                 types.StringValue(principal.Source),
			Roles: types.ListValueMust(types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"id":   types.StringType,
					"name": types.StringType,
				},
			}, roleValues),
			Applications:        types.ListValueMust(types.StringType, appValues),
			ServiceOptions:      serviceOptionsValue,
			CommandRestrictions: commandRestrictionsValue,
		}
	}

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

// tagsContainSameElements checks if two tag slices contain the same elements (ignoring order).
func tagsContainSameElements(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	// Create maps to count occurrences
	countA := make(map[string]int)
	countB := make(map[string]int)

	for _, tag := range a {
		countA[tag]++
	}

	for _, tag := range b {
		countB[tag]++
	}

	// Compare the maps
	for tag, count := range countA {
		if countB[tag] != count {
			return false
		}
	}

	return true
}

func stringOrNull(s string) attr.Value {
	if s == "" {
		return types.StringNull()
	}
	return types.StringValue(s)
}

func isUnsetString(v types.String) bool {
	return v.IsNull() || v.IsUnknown() || v.ValueString() == ""
}
func isUnsetBool(v types.Bool) bool {
	return v.IsNull() || v.IsUnknown()
}

func expandPasswordRotation(m PasswordRotationModel) *hoststore.RotationMetadata {
	// If object truly absent -> nil
	if m.AccessGroupID.IsNull() && m.UseMainAccount.IsNull() &&
		m.OperatingSystem.IsNull() && m.Protocol.IsNull() &&
		m.PasswordPolicyID.IsNull() && m.ScriptTemplateID.IsNull() &&
		m.CertificateValidationOptions.IsNull() &&
		m.WinrmHostCertificateTrustAnchors.IsNull() &&
		m.WinrmAddress.IsNull() && m.WinrmPort.IsNull() {
		return nil
	}

	// If required fields not set -> return nil
	if isUnsetBool(m.UseMainAccount) ||
		isUnsetString(m.OperatingSystem) ||
		isUnsetString(m.Protocol) ||
		isUnsetString(m.PasswordPolicyID) ||
		isUnsetString(m.ScriptTemplateID) {
		return nil
	}

	cvo := "DISABLED"
	if !isUnsetString(m.CertificateValidationOptions) {
		cvo = m.CertificateValidationOptions.ValueString()
	}

	rm := &hoststore.RotationMetadata{
		UseMainAccount:                   m.UseMainAccount.ValueBool(),
		OperatingSystem:                  m.OperatingSystem.ValueString(),
		Protocol:                         m.Protocol.ValueString(),
		PasswordPolicyId:                 m.PasswordPolicyID.ValueString(),
		ScriptTemplateId:                 m.ScriptTemplateID.ValueString(),
		CertificateValidationOptions:     cvo,
		WinRMHostCertificateTrustAnchors: m.WinrmHostCertificateTrustAnchors.ValueString(),
	}

	if !isUnsetString(m.AccessGroupID) {
		rm.AccessGroupID = m.AccessGroupID.ValueString()
	}

	// Only set WinRM fields when protocol is WINRM (avoid polluting SSH configs)
	if rm.Protocol == "WINRM" {
		rm.WinrmAddress = m.WinrmAddress.ValueString()
		rm.WinrmPort = int(m.WinrmPort.ValueInt64())
	}

	return rm
}

func boolOrFalse(v types.Bool) bool {
	if v.IsNull() || v.IsUnknown() {
		return false
	}
	return v.ValueBool()
}

func boolIsKnown(v types.Bool) bool {
	return !v.IsNull() && !v.IsUnknown()
}

// helper: empty string from API should be treated as DISABLED
func certValidationOrDefault(v string) attr.Value {
	if v == "" {
		return types.StringValue("DISABLED")
	}
	return types.StringValue(v)
}
