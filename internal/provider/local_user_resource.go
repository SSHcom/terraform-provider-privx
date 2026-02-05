package provider

import (
	"context"
	"fmt"
	"terraform-provider-privx/internal/utils"

	"github.com/SSHcom/privx-sdk-go/v2/api/userstore"
	"github.com/SSHcom/privx-sdk-go/v2/restapi"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &LocalUserResource{}
var _ resource.ResourceWithImportState = &LocalUserResource{}

func NewLocalUserResource() resource.Resource {
	return &LocalUserResource{}
}

type LocalUserResource struct {
	client *userstore.UserStore
}

type LocalUserResourceModel struct {
	ID                     types.String `tfsdk:"id"`
	Username               types.String `tfsdk:"username"`
	FullName               types.String `tfsdk:"full_name"`
	Email                  types.String `tfsdk:"email"`
	FirstName              types.String `tfsdk:"first_name"`
	LastName               types.String `tfsdk:"last_name"`
	Password               types.String `tfsdk:"password"`
	PasswordChangeRequired types.Bool   `tfsdk:"password_change_required"`
	JobTitle               types.String `tfsdk:"job_title"`
	Department             types.String `tfsdk:"department"`
	Company                types.String `tfsdk:"company"`
	Telephone              types.String `tfsdk:"telephone"`
	Locale                 types.String `tfsdk:"locale"`
	UnixAccount            types.String `tfsdk:"unix_account"`
	WindowsAccount         types.String `tfsdk:"windows_account"`
	Tags                   types.List   `tfsdk:"tags"`
}

func (r *LocalUserResource) Metadata(
	ctx context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_local_user"
}

func (r *LocalUserResource) Schema(
	ctx context.Context,
	req resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "PrivX local user",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"full_name": schema.StringAttribute{
				MarkdownDescription: "Full name of the user",
				Required:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Initial password (write-only)",
				Optional:            true,
				Sensitive:           true,
			},
			"password_change_required": schema.BoolAttribute{
				MarkdownDescription: "Whether the user must change their password at next login",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"username": schema.StringAttribute{
				Required: true,
			},
			"email": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(""),
			},
			"first_name": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(""),
			},
			"last_name": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(""),
			},
			"job_title": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(""),
			},
			"department": schema.StringAttribute{
				MarkdownDescription: "Department of the user",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"company": schema.StringAttribute{
				MarkdownDescription: "Company of the user",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"telephone": schema.StringAttribute{
				MarkdownDescription: "Phone number of the user",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"locale": schema.StringAttribute{
				MarkdownDescription: "Locale of the user",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"unix_account": schema.StringAttribute{
				MarkdownDescription: "Unix account name",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"windows_account": schema.StringAttribute{
				MarkdownDescription: "Windows account name",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"tags": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (r *LocalUserResource) Configure(
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

func (r *LocalUserResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data LocalUserResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	user := userstore.LocalUser{
		Principal:              data.Username.ValueString(),
		FullName:               data.FullName.ValueString(),
		Email:                  data.Email.ValueString(),
		FirstName:              data.FirstName.ValueString(),
		LastName:               data.LastName.ValueString(),
		JobTitle:               data.JobTitle.ValueString(),
		Department:             data.Department.ValueString(),
		Company:                data.Company.ValueString(),
		Telephone:              data.Telephone.ValueString(),
		Locale:                 data.Locale.ValueString(),
		UnixAccount:            data.UnixAccount.ValueString(),
		WindowsAccount:         data.WindowsAccount.ValueString(),
		PasswordChangeRequired: data.PasswordChangeRequired.ValueBool(),
	}

	if !data.Tags.IsNull() && !data.Tags.IsUnknown() {
		var tags []string
		diags := data.Tags.ElementsAs(ctx, &tags, false)
		resp.Diagnostics.Append(diags...)
		user.Tags = tags
	}

	if !data.Password.IsNull() && !data.Password.IsUnknown() {
		user.Password = userstore.LocalUserPassword{
			Password: data.Password.ValueString(),
		}
	}

	tflog.Debug(ctx, "Creating local user", map[string]any{
		"principal": user.Principal,
	})

	identifier, err := r.client.CreateUser(&user)
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to create local user: %s", err),
		)
		return
	}

	data.ID = types.StringValue(identifier.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *LocalUserResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data LocalUserResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	user, err := r.client.GetUser(data.ID.ValueString())
	if err != nil {
		if utils.IsPrivxNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read user, got error: %s", err))
		return
	}

	data.Username = types.StringValue(user.Principal)
	data.FullName = types.StringValue(user.FullName)
	data.Email = types.StringValue(user.Email)
	data.FirstName = types.StringValue(user.FirstName)
	data.LastName = types.StringValue(user.LastName)
	data.JobTitle = types.StringValue(user.JobTitle)
	data.Department = types.StringValue(user.Department)
	data.Company = types.StringValue(user.Company)
	data.Telephone = types.StringValue(user.Telephone)
	data.Locale = types.StringValue(user.Locale)
	data.UnixAccount = types.StringValue(user.UnixAccount)
	data.WindowsAccount = types.StringValue(user.WindowsAccount)
	data.PasswordChangeRequired = types.BoolValue(user.PasswordChangeRequired)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *LocalUserResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data LocalUserResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	user := userstore.LocalUser{
		ID:                     data.ID.ValueString(),
		Principal:              data.Username.ValueString(),
		FullName:               data.FullName.ValueString(),
		Email:                  data.Email.ValueString(),
		FirstName:              data.FirstName.ValueString(),
		LastName:               data.LastName.ValueString(),
		JobTitle:               data.JobTitle.ValueString(),
		Department:             data.Department.ValueString(),
		Company:                data.Company.ValueString(),
		Telephone:              data.Telephone.ValueString(),
		Locale:                 data.Locale.ValueString(),
		UnixAccount:            data.UnixAccount.ValueString(),
		WindowsAccount:         data.WindowsAccount.ValueString(),
		PasswordChangeRequired: data.PasswordChangeRequired.ValueBool(),
	}

	// --- TAGS HANDLING ---
	if !data.Tags.IsNull() && !data.Tags.IsUnknown() {
		var tags []string
		diags := data.Tags.ElementsAs(ctx, &tags, false)
		resp.Diagnostics.Append(diags...)
		user.Tags = tags
	}
	// ---------------------

	err := r.client.UpdateUser(data.ID.ValueString(), &user)
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to update local user: %s", err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *LocalUserResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data LocalUserResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteUser(data.ID.ValueString())
	if err != nil {
		if utils.IsPrivxNotFound(err) {
			return
		}
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to delete local user: %s", err),
		)
		return
	}

}

func (r *LocalUserResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
