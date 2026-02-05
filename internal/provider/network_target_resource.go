package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"terraform-provider-privx/internal/utils"

	"github.com/SSHcom/privx-sdk-go/v2/api/networkaccessmanager"
	"github.com/SSHcom/privx-sdk-go/v2/restapi"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &NetworkTargetResource{}
var _ resource.ResourceWithImportState = &NetworkTargetResource{}

// -------------------------------------------------------------------
// Resource type
// -------------------------------------------------------------------

func NewNetworkTargetResource() resource.Resource {
	return &NetworkTargetResource{}
}

type NetworkTargetResource struct {
	client *networkaccessmanager.NetworkAccessManager
}
type RolesHandleModel struct {
	ID types.String `tfsdk:"id"`
}

type DestinationModel struct {
	IPStart   types.String `tfsdk:"ip_start"`
	IPEnd     types.String `tfsdk:"ip_end"`
	Protocol  types.String `tfsdk:"protocol"`
	PortStart types.Int64  `tfsdk:"port_start"`
	PortEnd   types.Int64  `tfsdk:"port_end"`
	NATAddr   types.String `tfsdk:"nat_addr"`
	NATPort   types.Int64  `tfsdk:"nat_port"`
}

type NetworkTargetResourceModel struct {
	ID               types.String       `tfsdk:"id"`
	Name             types.String       `tfsdk:"name"`
	Comment          types.String       `tfsdk:"comment"`
	UserInstructions types.String       `tfsdk:"user_instructions"`
	StaticConfig     types.String       `tfsdk:"static_config"`
	IntegrationType  types.String       `tfsdk:"integration_type"`
	Disabled         types.Bool         `tfsdk:"disabled"`
	SrcNAT           types.Bool         `tfsdk:"src_nat"`
	ExclusiveAccess  types.Bool         `tfsdk:"exclusive_access"`
	Tags             types.List         `tfsdk:"tags"`
	Roles            []RolesHandleModel `tfsdk:"roles"`
	Dst              []DestinationModel `tfsdk:"dst"`
}

// -------------------------------------------------------------------
// Metadata
// -------------------------------------------------------------------

func (r *NetworkTargetResource) Metadata(
	ctx context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_network_target"
}

// -------------------------------------------------------------------
// Schema
// -------------------------------------------------------------------

func (r *NetworkTargetResource) Schema(
	ctx context.Context,
	req resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "PrivX Network Target",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"name": schema.StringAttribute{
				Required: true,
			},

			"comment": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(""),
			},

			"user_instructions": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(""),
			},

			"static_config": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(""),
			},

			"integration_type": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(""),
			},

			"disabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},

			"src_nat": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},

			"exclusive_access": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},

			"tags": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
			},
		},

		Blocks: map[string]schema.Block{

			"roles": schema.ListNestedBlock{
				MarkdownDescription: "PrivX roles allowed for this network target",

				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Required: true,
						},
					},
				},
			},

			"dst": schema.ListNestedBlock{
				MarkdownDescription: "Destination endpoints",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"ip_start": schema.StringAttribute{
							Required: true,
						},
						"ip_end": schema.StringAttribute{
							Required: true,
						},
						"protocol": schema.StringAttribute{
							Optional: true,
							Computed: true,
						},
						"port_start": schema.Int64Attribute{
							Optional: true,
						},
						"port_end": schema.Int64Attribute{
							Optional: true,
						},
						"nat_addr": schema.StringAttribute{
							Optional: true,
						},
						"nat_port": schema.Int64Attribute{
							Optional: true,
						},
					},
				},
			},
		},
	}
}

// -------------------------------------------------------------------
// Configure
// -------------------------------------------------------------------

func (r *NetworkTargetResource) Configure(
	ctx context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	connector, ok := req.ProviderData.(*restapi.Connector)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Configure Type",
			fmt.Sprintf("Expected *restapi.Connector, got %T", req.ProviderData),
		)
		return
	}

	r.client = networkaccessmanager.New(*connector)
}

func (r *NetworkTargetResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data NetworkTargetResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	nt := networkaccessmanager.NetworkTarget{
		Name:             data.Name.ValueString(),
		Comment:          data.Comment.ValueString(),
		UserInstructions: data.UserInstructions.ValueString(),
		StaticConfig:     data.StaticConfig.ValueString(),
		IntegrationType:  data.IntegrationType.ValueString(),
		Disabled:         strconv.FormatBool(data.Disabled.ValueBool()),
		SrcNAT:           data.SrcNAT.ValueBool(),
		ExclusiveAccess:  data.ExclusiveAccess.ValueBool(),
	}

	if !data.Tags.IsNull() && !data.Tags.IsUnknown() {
		var tags []string
		resp.Diagnostics.Append(data.Tags.ElementsAs(ctx, &tags, false)...)
		nt.Tags = tags
	}

	for _, role := range data.Roles {
		nt.Roles = append(nt.Roles, networkaccessmanager.RoleHandle{
			ID: role.ID.ValueString(),
		})
	}

	for _, d := range data.Dst {
		sel := networkaccessmanager.Selector{
			IP: networkaccessmanager.IPRange{
				Start: d.IPStart.ValueString(),
				End:   d.IPEnd.ValueString(),
			},
		}
		if !d.Protocol.IsNull() {
			sel.Protocol = d.Protocol.ValueString()
		}
		if !d.PortStart.IsNull() && !d.PortEnd.IsNull() {
			sel.Port = &networkaccessmanager.PortRange{
				Start: int(d.PortStart.ValueInt64()),
				End:   int(d.PortEnd.ValueInt64()),
			}
		}

		var nat *networkaccessmanager.NATParameters
		if !d.NATAddr.IsNull() || !d.NATPort.IsNull() {
			nat = &networkaccessmanager.NATParameters{
				Addr: d.NATAddr.ValueString(),
				Port: int(d.NATPort.ValueInt64()),
			}
		}

		nt.Dst = append(nt.Dst, networkaccessmanager.Destination{
			Sel: sel,
			NAT: nat,
		})
	}

	out, err := r.client.CreateNetworkTarget(&nt)
	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to create network target: %s", err))
		return
	}

	// Re-read created object to populate state consistently
	created, err := r.client.GetNetworkTarget(out.ID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to read created network target (%s): %s", out.ID, err))
		return
	}

	data.ID = types.StringValue(created.ID)
	data.Name = types.StringValue(created.Name)
	data.Comment = types.StringValue(created.Comment)
	data.UserInstructions = types.StringValue(created.UserInstructions)
	data.StaticConfig = types.StringValue(created.StaticConfig)
	data.IntegrationType = types.StringValue(created.IntegrationType)
	data.Disabled = types.BoolValue(strings.EqualFold(created.Disabled, "true"))
	data.SrcNAT = types.BoolValue(created.SrcNAT)
	data.ExclusiveAccess = types.BoolValue(created.ExclusiveAccess)

	// tags from server
	tags, diags := types.ListValueFrom(ctx, types.StringType, created.Tags)
	resp.Diagnostics.Append(diags...)
	data.Tags = tags

	// roles from server
	data.Roles = []RolesHandleModel{}
	for _, role := range created.Roles {
		data.Roles = append(data.Roles, RolesHandleModel{
			ID: types.StringValue(role.ID),
		})
	}

	// dst from server
	data.Dst = []DestinationModel{}
	for _, d := range created.Dst {
		dm := DestinationModel{
			IPStart:  types.StringValue(d.Sel.IP.Start),
			IPEnd:    types.StringValue(d.Sel.IP.End),
			Protocol: types.StringValue(d.Sel.Protocol),
		}
		if d.Sel.Port != nil {
			dm.PortStart = types.Int64Value(int64(d.Sel.Port.Start))
			dm.PortEnd = types.Int64Value(int64(d.Sel.Port.End))
		}
		if d.NAT != nil {
			dm.NATAddr = types.StringValue(d.NAT.Addr)
			dm.NATPort = types.Int64Value(int64(d.NAT.Port))
		}
		data.Dst = append(data.Dst, dm)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

}

func (r *NetworkTargetResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data NetworkTargetResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	nt, err := r.client.GetNetworkTarget(data.ID.ValueString())
	if err != nil {
		if utils.IsPrivxNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read network target, got error: %s", err))
		return
	}

	data.Name = types.StringValue(nt.Name)
	data.Comment = types.StringValue(nt.Comment)
	data.UserInstructions = types.StringValue(nt.UserInstructions)
	data.StaticConfig = types.StringValue(nt.StaticConfig)
	data.IntegrationType = types.StringValue(nt.IntegrationType)
	data.Disabled = types.BoolValue(strings.EqualFold(nt.Disabled, "true"))
	data.SrcNAT = types.BoolValue(nt.SrcNAT)
	data.ExclusiveAccess = types.BoolValue(nt.ExclusiveAccess)

	tags, diags := types.ListValueFrom(ctx, types.StringType, nt.Tags)
	resp.Diagnostics.Append(diags...)
	data.Tags = tags

	data.Roles = []RolesHandleModel{}
	for _, role := range nt.Roles {
		data.Roles = append(data.Roles, RolesHandleModel{
			ID: types.StringValue(role.ID),
		})
	}

	data.Dst = []DestinationModel{}
	data.Dst = []DestinationModel{}
	for _, d := range nt.Dst {
		dm := DestinationModel{
			IPStart:  types.StringValue(d.Sel.IP.Start),
			IPEnd:    types.StringValue(d.Sel.IP.End),
			Protocol: types.StringValue(d.Sel.Protocol),
		}

		if d.Sel.Port != nil {
			dm.PortStart = types.Int64Value(int64(d.Sel.Port.Start))
			dm.PortEnd = types.Int64Value(int64(d.Sel.Port.End))
		}

		if d.NAT != nil {
			dm.NATAddr = types.StringValue(d.NAT.Addr)
			dm.NATPort = types.Int64Value(int64(d.NAT.Port))
		}

		data.Dst = append(data.Dst, dm)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NetworkTargetResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data NetworkTargetResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	nt := networkaccessmanager.NetworkTarget{
		ID:               data.ID.ValueString(),
		Name:             data.Name.ValueString(),
		Comment:          data.Comment.ValueString(),
		UserInstructions: data.UserInstructions.ValueString(),
		StaticConfig:     data.StaticConfig.ValueString(),
		IntegrationType:  data.IntegrationType.ValueString(),
		Disabled:         strconv.FormatBool(data.Disabled.ValueBool()),
		SrcNAT:           data.SrcNAT.ValueBool(),
		ExclusiveAccess:  data.ExclusiveAccess.ValueBool(),
	}

	// tags
	if !data.Tags.IsNull() && !data.Tags.IsUnknown() {
		var tags []string
		resp.Diagnostics.Append(data.Tags.ElementsAs(ctx, &tags, false)...)
		nt.Tags = tags
	}

	// roles
	for _, role := range data.Roles {
		nt.Roles = append(nt.Roles, networkaccessmanager.RoleHandle{
			ID: role.ID.ValueString(),
		})
	}

	// destinations (v42 selector + NAT)
	for _, d := range data.Dst {
		sel := networkaccessmanager.Selector{
			IP: networkaccessmanager.IPRange{
				Start: d.IPStart.ValueString(),
				End:   d.IPEnd.ValueString(),
			},
		}

		// optional protocol
		if !d.Protocol.IsNull() && !d.Protocol.IsUnknown() {
			sel.Protocol = d.Protocol.ValueString()
		}

		// optional port range
		if !d.PortStart.IsNull() && !d.PortEnd.IsNull() {
			sel.Port = &networkaccessmanager.PortRange{
				Start: int(d.PortStart.ValueInt64()),
				End:   int(d.PortEnd.ValueInt64()),
			}
		}

		// optional NAT
		var nat *networkaccessmanager.NATParameters
		if (!d.NATAddr.IsNull() && !d.NATAddr.IsUnknown()) ||
			(!d.NATPort.IsNull() && !d.NATPort.IsUnknown()) {
			nat = &networkaccessmanager.NATParameters{
				Addr: d.NATAddr.ValueString(),
				Port: int(d.NATPort.ValueInt64()),
			}
		}

		nt.Dst = append(nt.Dst, networkaccessmanager.Destination{
			Sel: sel,
			NAT: nat,
		})
	}

	// apply update
	if err := r.client.UpdateNetworkTarget(data.ID.ValueString(), &nt); err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to update network target: %s", err),
		)
		return
	}

	updated, err := r.client.GetNetworkTarget(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to read updated network target (%s): %s", data.ID.ValueString(), err))
		return
	}

	// Map updated -> data exactly like in Read()
	data.Name = types.StringValue(updated.Name)
	data.Comment = types.StringValue(updated.Comment)
	data.UserInstructions = types.StringValue(updated.UserInstructions)
	data.StaticConfig = types.StringValue(updated.StaticConfig)
	data.IntegrationType = types.StringValue(updated.IntegrationType)
	data.Disabled = types.BoolValue(strings.EqualFold(updated.Disabled, "true"))
	data.SrcNAT = types.BoolValue(updated.SrcNAT)
	data.ExclusiveAccess = types.BoolValue(updated.ExclusiveAccess)

	tags, diags := types.ListValueFrom(ctx, types.StringType, updated.Tags)
	resp.Diagnostics.Append(diags...)
	data.Tags = tags

	data.Roles = []RolesHandleModel{}
	for _, role := range updated.Roles {
		data.Roles = append(data.Roles, RolesHandleModel{ID: types.StringValue(role.ID)})
	}

	data.Dst = []DestinationModel{}
	for _, d := range updated.Dst {
		dm := DestinationModel{
			IPStart:  types.StringValue(d.Sel.IP.Start),
			IPEnd:    types.StringValue(d.Sel.IP.End),
			Protocol: types.StringValue(d.Sel.Protocol),
		}
		if d.Sel.Port != nil {
			dm.PortStart = types.Int64Value(int64(d.Sel.Port.Start))
			dm.PortEnd = types.Int64Value(int64(d.Sel.Port.End))
		}
		if d.NAT != nil {
			dm.NATAddr = types.StringValue(d.NAT.Addr)
			dm.NATPort = types.Int64Value(int64(d.NAT.Port))
		}
		data.Dst = append(data.Dst, dm)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

}

func (r *NetworkTargetResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data NetworkTargetResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteNetworkTarget(data.ID.ValueString())
	if err != nil {
		if utils.IsPrivxNotFound(err) {
			// already deleted out-of-band -> treat as success
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete network target: %s", err))
		return
	}
}

// -------------------------------------------------------------------
// IMPORT
// -------------------------------------------------------------------

func (r *NetworkTargetResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
