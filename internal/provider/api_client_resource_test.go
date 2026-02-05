package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAPIClientResource_withRole(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC not set")
	}

	suffix := acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)

	roleName := fmt.Sprintf("tf-acc-role-api-clients-manage-%s", suffix)
	apiClientName := fmt.Sprintf("tf-acc-api-client-with-role-%s", suffix)
	groupName := fmt.Sprintf("tf-acc-group-api-client-%s", suffix)

	cfg := testAccAPIClientWithRoleConfig(roleName, apiClientName, groupName)
	writeAccConfig(t, fmt.Sprintf("%s_step_1.tf", t.Name()), cfg)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,

		Steps: []resource.TestStep{
			{
				Config: cfg,

				Check: resource.ComposeTestCheckFunc(
					// Role exists + permission set
					resource.TestCheckResourceAttr("privx_role.role_target", "name", roleName),
					resource.TestCheckResourceAttr("privx_role.role_target", "permissions.#", "1"),
					resource.TestCheckTypeSetElemAttr("privx_role.role_target", "permissions.*", "api-clients-manage"),

					// API client exists
					resource.TestCheckResourceAttr("privx_api_client.test", "name", apiClientName),
					resource.TestCheckResourceAttrSet("privx_api_client.test", "id"),
					// Verify secrets are present using correct schema names
					resource.TestCheckResourceAttrSet("privx_api_client.test", "oauth_client_id"),
					resource.TestCheckResourceAttrSet("privx_api_client.test", "oauth_client_secret"),
					resource.TestCheckResourceAttr("privx_api_client.test", "roles.#", "1"),

					resource.TestCheckTypeSetElemAttrPair("privx_api_client.test", "roles.*.id", "privx_role.role_target", "id"),
				),
			},
			{
				RefreshState: true,
			},

			// 3️⃣ (Optional) Import, Update, etc.
			{
				ResourceName:      "privx_api_client.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAPIClientWithRoleConfig(roleName, apiClientName string, groupName string) string {
	return fmt.Sprintf(`
terraform {
  required_providers {
    privx = {
      source = "hashicorp/privx"
    }
  }
}

provider "privx" {}

resource "privx_access_group" "foo" {
  name    = %q
  comment = "temp group for api client acc test"
}

resource "privx_role" "role_target" {
  name            = "%s"
  access_group_id = privx_access_group.foo.id

  permissions  = ["api-clients-manage"]
  permit_agent = false

  source_rules = jsonencode({
    type  = "GROUP"
    match = "ANY"
    rules = []
  })
}

resource "privx_api_client" "test" {
  name = "%s"

  roles = [
    {
      id = privx_role.role_target.id
    }
  ]
}
`, groupName, roleName, apiClientName)
}
