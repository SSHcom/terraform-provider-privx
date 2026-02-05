package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// This test relies only on existing PrivX API Proxy configuration and does not depend on any Terraform-created resources.
// It is optional and for human visibility/debugging only.
func TestAccApiProxyConfigDataSource(t *testing.T) {
	cfg := testAccApiProxyConfigDataSourceConfig()
	writeAccConfig(t, fmt.Sprintf("%s_step_1.tf", t.Name()), cfg)

	ds := "data.privx_api_proxy_config.this"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: cfg,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(ds, "id"),

					// addresses[0] exists and looks like http(s) URL
					resource.TestCheckResourceAttrSet(ds, "addresses.0"),
					resource.TestMatchResourceAttr(
						ds,
						"addresses.0",
						regexp.MustCompile(`^https?://`),
					),

					// CA chain exists and looks like PEM
					resource.TestCheckResourceAttrSet(ds, "ca_certificate_chain"),
					resource.TestMatchResourceAttr(
						ds,
						"ca_certificate_chain",
						regexp.MustCompile(`-----BEGIN CERTIFICATE-----`),
					),

					testAccPrintApiProxyConfig(ds),
				),
			},
		},
	})
}

func testAccPrintApiProxyConfig(ds string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[ds]
		if !ok {
			return fmt.Errorf("resource not found in state: %s", ds)
		}

		fmt.Println("=== api_proxy_config data source ===")

		if v, ok := rs.Primary.Attributes["id"]; ok {
			fmt.Println("id:", v)
		}
		if v, ok := rs.Primary.Attributes["addresses.0"]; ok {
			fmt.Println("addresses[0]:", v)
		}
		if v, ok := rs.Primary.Attributes["ca_certificate_chain"]; ok {
			fmt.Println("ca_certificate_chain:\n", v)
		}

		fmt.Println("===================================")

		return nil
	}
}

func testAccApiProxyConfigDataSourceConfig() string {
	return `
terraform {
  required_providers {
    privx = {
      source = "hashicorp/privx"
    }
  }
}

provider "privx" {}

data "privx_api_proxy_config" "this" {}
`
}
