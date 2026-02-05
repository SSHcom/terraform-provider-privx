package provider

import (
	"context"
	"fmt"

	"github.com/SSHcom/privx-sdk-go/v2/api/userstore"
	"github.com/SSHcom/privx-sdk-go/v2/restapi"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Interface checks.
var _ resource.Resource = &LocalUserPasswordResource{}

// Constructor.
func NewLocalUserPasswordResource() resource.Resource {
	return &LocalUserPasswordResource{}
}

// Resource implementation.
type LocalUserPasswordResource struct {
	client *userstore.UserStore
}

// Terraform model.
type LocalUserPasswordModel struct {
	ID       types.String `tfsdk:"id"`
	UserID   types.String `tfsdk:"user_id"`
	Password types.String `tfsdk:"password"`
}

// Metadata.
func (r *LocalUserPasswordResource) Metadata(
	ctx context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_local_user_password"
}

// Schema.
func (r *LocalUserPasswordResource) Schema(
	ctx context.Context,
	req resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Sets or resets a PrivX local user password",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"user_id": schema.StringAttribute{
				MarkdownDescription: "Local user ID",
				Required:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "New password (write-only)",
				Required:            true,
				Sensitive:           true,
			},
		},
	}
}

// Configure.
func (r *LocalUserPasswordResource) Configure(
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
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *restapi.Connector, got %T", req.ProviderData),
		)
		return
	}

	r.client = userstore.New(*connector)
}

// CREATE → PUT /local-user-store/api/v1/users/{id}/password.
func (r *LocalUserPasswordResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data LocalUserPasswordModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	password := userstore.LocalUserPassword{
		Password: data.Password.ValueString(),
	}

	err := r.client.UpdateUserPassword(
		data.UserID.ValueString(),
		password,
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to set local user password: %s", err),
		)
		return
	}

	// Synthetic ID: password::<user_id>
	data.ID = types.StringValue("password::" + data.UserID.ValueString())

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// READ → NO-OP (passwords cannot be read).
func (r *LocalUserPasswordResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	// Intentionally empty:
	// - Passwords cannot be read
	// - Resource always assumed to exist
}

// UPDATE → Reset password.
func (r *LocalUserPasswordResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data LocalUserPasswordModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	password := userstore.LocalUserPassword{
		Password: data.Password.ValueString(),
	}

	err := r.client.UpdateUserPassword(
		data.UserID.ValueString(),
		password,
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to update local user password: %s", err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// DELETE → NO-OP.
func (r *LocalUserPasswordResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	// Intentionally empty:
	// - Deleting a password is not supported
	// - User deletion handles cleanup
}
