package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccCACertificateDataSource_readsCAs(t *testing.T) {
	cfg := `
provider "privx" {}

data "privx_ca_certificate" "all" {}
`

	writeAccConfig(t, "TestAccCACertificateDataSource_readsCAs.tf", cfg)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: cfg,
				Check: resource.ComposeAggregateTestCheckFunc(
					// A PrivX instance always has at least one authorizer CA, and
					// each entry must expose a public key (the drift-guard value).
					resource.TestCheckResourceAttrSet("data.privx_ca_certificate.all", "certificates.0.id"),
					resource.TestCheckResourceAttrSet("data.privx_ca_certificate.all", "certificates.0.public_key"),
				),
			},
		},
	})
}
