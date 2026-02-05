package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccWhitelistResource_basic(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC not set")
	}
	resourceName := "privx_whitelist.test"
	name := fmt.Sprintf("tf-acc-whitelist-%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum))
	cfg1 := testAccWhitelistConfigCreate(name)
	writeAccConfig(t, fmt.Sprintf("%s_step_1.tf", t.Name()), cfg1)

	cfg2 := testAccWhitelistConfigUpdate(name)
	writeAccConfig(t, fmt.Sprintf("%s_step_2.tf", t.Name()), cfg2)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,

		Steps: []resource.TestStep{
			// CREATE
			{
				Config: cfg1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "type", "glob"),
					resource.TestCheckResourceAttr(resourceName, "comment", "Example whitelist for command restrictions"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),

					// Set attribute checks
					resource.TestCheckResourceAttr(resourceName, "whitelist_patterns.#", "3"),
					resource.TestCheckTypeSetElemAttr(resourceName, "whitelist_patterns.*", "ls -la"),
					resource.TestCheckTypeSetElemAttr(resourceName, "whitelist_patterns.*", "cat /etc/passwd"),
					resource.TestCheckTypeSetElemAttr(resourceName, "whitelist_patterns.*", "systemctl status *"),
				),
			},

			// UPDATE
			{
				Config: cfg2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "type", "regex"),
					resource.TestCheckResourceAttr(resourceName, "comment", "Updated whitelist comment"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),

					resource.TestCheckResourceAttr(resourceName, "whitelist_patterns.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "whitelist_patterns.*", `^git .*`),
					resource.TestCheckTypeSetElemAttr(resourceName, "whitelist_patterns.*", `^docker .*`),
				),
			},
		},
	})
}

func testAccWhitelistConfigCreate(name string) string {
	return fmt.Sprintf(`
terraform {
  required_providers {
    privx = {
      source = "hashicorp/privx"
    }
  }
}

provider "privx" {}

resource "privx_whitelist" "test" {
  name    = "%s"
  comment = "Example whitelist for command restrictions"
  type    = "glob"

  whitelist_patterns = [
    "ls -la",
    "cat /etc/passwd",
    "systemctl status *"
  ]
}
`, name)
}

func testAccWhitelistConfigUpdate(name string) string {
	return fmt.Sprintf(`
terraform {
  required_providers {
    privx = {
      source = "hashicorp/privx"
    }
  }
}

provider "privx" {}

resource "privx_whitelist" "test" {
  name    = "%s"
  comment = "Updated whitelist comment"
  type    = "regex"

  whitelist_patterns = [
    "^git .*",
    "^docker .*"
  ]
}
`, name)
}
