package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccCarrierResource(t *testing.T) {
	// Avoid name validator problems: only lowercase letters + digits
	carrierName := "tfcarrier" + acctest.RandStringFromCharSet(8, "abcdefghijklmnopqrstuvwxyz0123456789")
	carrierName2 := carrierName + "b" // still lowercase+digits

	resourceName := "privx_carrier.test"

	cfg1 := testAccCarrierConfigWithAccessGroup(carrierName, "carriera",
		"0.0.0.0",
		[]string{"route_pattern_a"},
		[]string{"10.10.0.0/16"},
		[]string{"10.10.10.10"},
	)
	writeAccConfig(t, fmt.Sprintf("%s_step_1.tf", t.Name()), cfg1)

	cfg2 := testAccCarrierConfigWithAccessGroup(carrierName2, "carrierb",
		"0.0.0.0",
		[]string{"route_pattern_a"}, // keep same
		[]string{"10.11.0.0/16"},    // CHANGED
		[]string{"10.10.10.10"},
	)
	writeAccConfig(t, fmt.Sprintf("%s_step_2.tf", t.Name()), cfg2)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// 1) Create
			{
				Config: cfg1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "name", carrierName),
					resource.TestCheckResourceAttr(resourceName, "routing_prefix", "carriera"),
					resource.TestCheckResourceAttr(resourceName, "web_proxy_address", "0.0.0.0"),

					// computed
					resource.TestCheckResourceAttrSet(resourceName, "registered"),
					resource.TestCheckResourceAttr(resourceName, "type", "CARRIER"),

					// lists
					resource.TestCheckResourceAttr(resourceName, "web_proxy_extender_route_patterns.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "web_proxy_extender_route_patterns.*", "route_pattern_a"),

					resource.TestCheckResourceAttr(resourceName, "subnets.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "subnets.*", "10.10.0.0/16"),

					resource.TestCheckResourceAttr(resourceName, "extender_address.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "extender_address.*", "10.10.10.10"),
				),
			},

			// 2) Update (toggle enabled, change routing_prefix + name + subnets)
			{
				Config: cfg2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "routing_prefix", "carrierb"),
					resource.TestCheckResourceAttr(resourceName, "name", carrierName2),

					resource.TestCheckResourceAttr(resourceName, "subnets.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "subnets.*", "10.11.0.0/16"),
				),
			},

			// 3) Import
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"registered",
					"group_id",
					"permissions",
					"access_group_id",
				},
			},
		},
	})
}

func TestAccCarrierResource_subnetsDefaultedWhenOmitted(t *testing.T) {
	carrierName := "tfcarrier" + acctest.RandStringFromCharSet(8, "abcdefghijklmnopqrstuvwxyz0123456789")
	resourceName := "privx_carrier.test"

	cfg := testAccCarrierConfigOmitSubnets(
		carrierName,
		"carriera",
		"0.0.0.0",
		[]string{"route_pattern_a"},
		[]string{"10.10.10.10"},
	)
	writeAccConfig(t, fmt.Sprintf("%s_default_subnets.tf", t.Name()), cfg)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: cfg,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "name", carrierName),

					// API defaults should appear and should NOT cause "inconsistent result" anymore
					resource.TestCheckResourceAttr(resourceName, "subnets.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "subnets.*", "0.0.0.0/0"),
					resource.TestCheckTypeSetElemAttr(resourceName, "subnets.*", "::/0"),
				),
			},
		},
	})
}

// helper to render []string into Terraform HCL list: ["a","b"]
func toHCLStringList(items []string) string {
	if len(items) == 0 {
		return "[]"
	}
	out := "["
	for i, s := range items {
		if i > 0 {
			out += ","
		}
		out += fmt.Sprintf("%q", s)
	}
	out += "]"
	return out
}

func testAccCarrierConfigWithAccessGroup(carrierName string, routingPrefix, webProxyAddress string,
	routePatterns, subnets, extenderAddrs []string,
) string {
	return fmt.Sprintf(`
terraform {
  required_providers {
    privx = {
      source = "hashicorp/privx"
    }
  }
}

provider "privx" {}

resource "privx_carrier" "test" {
  name              = %q
  routing_prefix    = %q
  web_proxy_address = %q

  web_proxy_extender_route_patterns = %s
  subnets                            = %s
  extender_address                   = %s
}
`, carrierName, routingPrefix, webProxyAddress,
		toHCLStringList(routePatterns),
		toHCLStringList(subnets),
		toHCLStringList(extenderAddrs),
	)
}

func testAccCarrierConfigOmitSubnets(carrierName, routingPrefix, webProxyAddress string,
	routePatterns, extenderAddrs []string,
) string {
	return fmt.Sprintf(`
terraform {
  required_providers {
    privx = {
      source = "hashicorp/privx"
    }
  }
}

provider "privx" {}

resource "privx_carrier" "test" {
  name              = %q
  routing_prefix    = %q
  web_proxy_address = %q

  web_proxy_extender_route_patterns = %s
  extender_address                   = %s
  # NOTE: subnets intentionally omitted to be NULL in plan
}
`, carrierName, routingPrefix, webProxyAddress,
		toHCLStringList(routePatterns),
		toHCLStringList(extenderAddrs),
	)
}
