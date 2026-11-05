package provider

import (
	"context"
	"fmt"

	"github.com/SSHcom/privx-sdk-go/v2/api/authorizer"
	"github.com/SSHcom/privx-sdk-go/v2/restapi"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &CACertificateDataSource{}

// NewCACertificateDataSource returns a new PrivX CA certificate data source.
func NewCACertificateDataSource() datasource.DataSource {
	return &CACertificateDataSource{}
}

// CACertificateDataSource exposes the PrivX authorizer root CA certificates.
// It is primarily useful as a drift guard: if a CA is rotated in PrivX, the
// exposed public keys change and dependent configuration produces a diff.
type CACertificateDataSource struct {
	client *authorizer.Authorizer
}

// CACertificateDataSourceModel describes the data source data model.
type CACertificateDataSourceModel struct {
	AccessGroupID types.String       `tfsdk:"access_group_id"`
	Certificates  []CACertificateDSM `tfsdk:"certificates"`
}

// CACertificateDSM describes a single CA certificate entry.
type CACertificateDSM struct {
	ID                types.String `tfsdk:"id"`
	GroupID           types.String `tfsdk:"group_id"`
	AccessGroupID     types.String `tfsdk:"access_group_id"`
	Type              types.String `tfsdk:"type"`
	Size              types.Int64  `tfsdk:"size"`
	PublicKey         types.String `tfsdk:"public_key"`
	PublicKeyString   types.String `tfsdk:"public_key_string"`
	X509Certificate   types.String `tfsdk:"x509_certificate"`
	Subject           types.String `tfsdk:"subject"`
	Issuer            types.String `tfsdk:"issuer"`
	SerialNumber      types.String `tfsdk:"serial"`
	NotBefore         types.String `tfsdk:"not_before"`
	NotAfter          types.String `tfsdk:"not_after"`
	FingerPrintSHA1   types.String `tfsdk:"fingerprint_sha1"`
	FingerPrintSHA256 types.String `tfsdk:"fingerprint_sha256"`
	Comment           types.String `tfsdk:"comment"`
}

func (d *CACertificateDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ca_certificate"
}

func (d *CACertificateDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "PrivX authorizer root CA certificates. Use this to expose the " +
			"PrivX CA public key(s). Changes to the returned certificates (for example a CA " +
			"rotation) produce a plan diff, acting as a drift guard.",
		Attributes: map[string]schema.Attribute{
			"access_group_id": schema.StringAttribute{
				MarkdownDescription: "Optional access group ID to filter the returned CA certificates.",
				Optional:            true,
			},
			"certificates": schema.ListNestedAttribute{
				MarkdownDescription: "List of CA certificates returned by the PrivX authorizer.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "CA UUID.",
							Computed:            true,
						},
						"group_id": schema.StringAttribute{
							MarkdownDescription: "Group ID of the CA.",
							Computed:            true,
						},
						"access_group_id": schema.StringAttribute{
							MarkdownDescription: "Access group ID of the CA.",
							Computed:            true,
						},
						"type": schema.StringAttribute{
							MarkdownDescription: "CA key type.",
							Computed:            true,
						},
						"size": schema.Int64Attribute{
							MarkdownDescription: "CA key size.",
							Computed:            true,
						},
						"public_key": schema.StringAttribute{
							MarkdownDescription: "CA public key.",
							Computed:            true,
						},
						"public_key_string": schema.StringAttribute{
							MarkdownDescription: "CA public key as a string.",
							Computed:            true,
						},
						"x509_certificate": schema.StringAttribute{
							MarkdownDescription: "X.509 certificate (PEM).",
							Computed:            true,
						},
						"subject": schema.StringAttribute{
							MarkdownDescription: "Certificate subject.",
							Computed:            true,
						},
						"issuer": schema.StringAttribute{
							MarkdownDescription: "Certificate issuer.",
							Computed:            true,
						},
						"serial": schema.StringAttribute{
							MarkdownDescription: "Certificate serial number.",
							Computed:            true,
						},
						"not_before": schema.StringAttribute{
							MarkdownDescription: "Certificate validity start.",
							Computed:            true,
						},
						"not_after": schema.StringAttribute{
							MarkdownDescription: "Certificate validity end.",
							Computed:            true,
						},
						"fingerprint_sha1": schema.StringAttribute{
							MarkdownDescription: "SHA-1 fingerprint.",
							Computed:            true,
						},
						"fingerprint_sha256": schema.StringAttribute{
							MarkdownDescription: "SHA-256 fingerprint.",
							Computed:            true,
						},
						"comment": schema.StringAttribute{
							MarkdownDescription: "Optional human readable comment.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *CACertificateDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	tflog.Debug(ctx, "Creating authorizer client for CA certificate data source")

	d.client = authorizer.New(*connector)
}

func (d *CACertificateDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data CACertificateDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := d.client.GetCACertificates()
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read CA certificates, got error: %s", err))
		return
	}

	filterAccessGroup := data.AccessGroupID.ValueString()

	data.Certificates = make([]CACertificateDSM, 0, len(result.Items))
	for _, ca := range result.Items {
		if filterAccessGroup != "" && ca.AccessGroupID != filterAccessGroup {
			continue
		}
		data.Certificates = append(data.Certificates, CACertificateDSM{
			ID:                types.StringValue(ca.ID),
			GroupID:           types.StringValue(ca.GroupID),
			AccessGroupID:     types.StringValue(ca.AccessGroupID),
			Type:              types.StringValue(ca.Type),
			Size:              types.Int64Value(int64(ca.Size)),
			PublicKey:         types.StringValue(ca.PublicKey),
			PublicKeyString:   types.StringValue(ca.PublicKeyString),
			X509Certificate:   types.StringValue(ca.X509Certificate),
			Subject:           types.StringValue(ca.Subject),
			Issuer:            types.StringValue(ca.Issuer),
			SerialNumber:      types.StringValue(ca.SerialNumber),
			NotBefore:         types.StringValue(ca.NotBefore),
			NotAfter:          types.StringValue(ca.NotAfter),
			FingerPrintSHA1:   types.StringValue(ca.FingerPrintSHA1),
			FingerPrintSHA256: types.StringValue(ca.FingerPrintSHA256),
			Comment:           types.StringValue(ca.Comment),
		})
	}

	tflog.Debug(ctx, "Storing CA certificates into state", map[string]interface{}{
		"count": len(data.Certificates),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
