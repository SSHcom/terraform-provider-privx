package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccNetworkTargetResource(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC not set")
	}

	resourceName := "privx_network_target.test"

	targetName := fmt.Sprintf("tf-acc-test-target-%s",
		acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum),
	)

	roleName := acctest.RandomWithPrefix("tf-acc-target-role")
	groupName := acctest.RandomWithPrefix("tf-acc-ag")

	cfg := testAccNetworkTargetWithRoleConfig(roleName, groupName, targetName)
	writeAccConfig(t, fmt.Sprintf("%s_step_1.tf", t.Name()), cfg)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,

		Steps: []resource.TestStep{
			{
				Config: cfg,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", targetName),
					resource.TestCheckResourceAttr(resourceName, "roles.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "disabled", "false"),

					resource.TestCheckResourceAttr(resourceName, "tags.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tags.*", "tag-a"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tags.*", "tag-b"),
				),
			},
		},
	})
}

func testAccNetworkTargetWithRoleConfig(roleName, groupName, targetName string) string {
	return fmt.Sprintf(`
terraform {
  required_providers {
    privx = {
      source = "hashicorp/privx"
    }
  }
}
provider "privx" {}

resource "privx_access_group" "grp" {
  name    = %q
  comment = "temp group for network target test"
}

resource "privx_role" "role_target" {
  name            = %q
  access_group_id = privx_access_group.grp.id

  permissions  = ["users-view"]
  permit_agent = false

  source_rules = jsonencode({
    type  = "GROUP"
    match = "ANY"
    rules = []
  })
}

resource "privx_network_target" "test" {
  name     = %q
  disabled = false

  tags = ["tag-a", "tag-b"]

  roles { id = privx_role.role_target.id }

  dst {
    ip_start = "192.168.1.1"
    ip_end   = "192.168.1.50"
    protocol = "tcp"
  }
}
`, groupName, roleName, targetName)
}
