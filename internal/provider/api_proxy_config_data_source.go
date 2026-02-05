package provider

import (
	"context"
	"fmt"

	"github.com/SSHcom/privx-sdk-go/v2/api/apiproxy"
	"github.com/SSHcom/privx-sdk-go/v2/restapi"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &apiProxyConfigDataSource{}
	_ datasource.DataSourceWithConfigure = &apiProxyConfigDataSource{}
)

func NewApiProxyConfigDataSource() datasource.DataSource {
	return &apiProxyConfigDataSource{}
}

type apiProxyConfigDataSource struct {
	client *apiproxy.ApiProxy
}

type apiProxyConfigModel struct {
	ID                 types.String         `tfsdk:"id"`
	Addresses          types.List           `tfsdk:"addresses"`            // list(string)
	CACertificateChain types.String         `tfsdk:"ca_certificate_chain"` // PEM chain
	CACertificate      *apiProxyCACertModel `tfsdk:"ca_certificate"`       // optional nested object
}

type apiProxyCACertModel struct {
	Subject           types.String `tfsdk:"subject"`
	Issuer            types.String `tfsdk:"issuer"`
	Serial            types.String `tfsdk:"serial"`
	NotBefore         types.String `tfsdk:"not_before"`
	NotAfter          types.String `tfsdk:"not_after"`
	FingerprintSHA1   types.String `tfsdk:"fingerprint_sha1"`
	FingerprintSHA256 types.String `tfsdk:"fingerprint_sha256"`
}

func (d *apiProxyConfigDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_proxy_config"
}

func (d *apiProxyConfigDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads PrivX API Proxy configuration (proxy listener addresses and CA certificate chain).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Static ID for this singleton config.",
			},
			"addresses": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "API proxy listener addresses (HTTP Proxy Public Addresses).",
			},
			"ca_certificate_chain": schema.StringAttribute{
				Computed:    true,
				Description: "API proxy CA certificate chain (PEM).",
			},
			"ca_certificate": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "API proxy CA certificate metadata (if available).",
				Attributes: map[string]schema.Attribute{
					"subject":            schema.StringAttribute{Computed: true},
					"issuer":             schema.StringAttribute{Computed: true},
					"serial":             schema.StringAttribute{Computed: true},
					"not_before":         schema.StringAttribute{Computed: true},
					"not_after":          schema.StringAttribute{Computed: true},
					"fingerprint_sha1":   schema.StringAttribute{Computed: true},
					"fingerprint_sha256": schema.StringAttribute{Computed: true},
				},
			},
		},
	}
}

func (d *apiProxyConfigDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	// Your provider passes ProviderData as *restapi.Connector (pointer to interface)
	connPtr, ok := req.ProviderData.(*restapi.Connector)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected provider data type",
			fmt.Sprintf("Expected *restapi.Connector, got: %T", req.ProviderData),
		)
		return
	}
	if connPtr == nil || *connPtr == nil {
		resp.Diagnostics.AddError("Provider not configured", "Connector was nil")
		return
	}

	d.client = apiproxy.New(*connPtr)
}

func (d *apiProxyConfigDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	conf, err := d.client.GetApiProxyConfig()
	if err != nil {
		resp.Diagnostics.AddError("Read API proxy config failed", err.Error())
		return
	}

	addrs, diags := types.ListValueFrom(ctx, types.StringType, conf.Addresses)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state := apiProxyConfigModel{
		ID:                 types.StringValue("api-proxy-config"),
		Addresses:          addrs,
		CACertificateChain: types.StringValue(conf.Chain),
	}

	if conf.CACertificate != nil {
		state.CACertificate = &apiProxyCACertModel{
			Subject:           types.StringValue(conf.CACertificate.Subject),
			Issuer:            types.StringValue(conf.CACertificate.Issuer),
			Serial:            types.StringValue(conf.CACertificate.Serial),
			NotBefore:         types.StringValue(conf.CACertificate.NotBefore),
			NotAfter:          types.StringValue(conf.CACertificate.NotAfter),
			FingerprintSHA1:   types.StringValue(conf.CACertificate.FingerPrintSHA1),
			FingerprintSHA256: types.StringValue(conf.CACertificate.FingerPrintSHA256),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
