package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccExtenderResource(t *testing.T) {
	name := "tfextender" + acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)
	resourceName := "privx_extender.extender_test"

	cfgCreate := testAccExtenderWithRoutingPrefix(name, "rp1")
	writeAccConfig(t, fmt.Sprintf("%s_step_1.tf", t.Name()), cfgCreate)

	cfgUpdate := testAccExtenderWithRoutingPrefix(name, "rp2")
	writeAccConfig(t, fmt.Sprintf("%s_step_2.tf", t.Name()), cfgUpdate)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// 1) Create
			{
				Config: cfgCreate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "routing_prefix", "rp1"),

					// server-managed fields just need to exist (or be stable)
					resource.TestCheckResourceAttrSet(resourceName, "enabled"),
					resource.TestCheckResourceAttrSet(resourceName, "registered"),
				),
			},

			// 2) Refresh-only: ensure Read is stable
			{
				RefreshState: true,
			},

			// 3) Update a user-controlled field
			{
				Config: cfgUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "routing_prefix", "rp2"),
				),
			},

			// 4) Import
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"registered", // can change
				},
			},
		},
	})
}

func testAccExtenderWithRoutingPrefix(name, rp string) string {
	return fmt.Sprintf(`
terraform {
  required_providers {
    privx = {
      source = "hashicorp/privx"
    }
  }
}

provider "privx" {}

resource "privx_extender" "extender_test" {
  name           = %q
  routing_prefix = %q
}
`, name, rp)
}
