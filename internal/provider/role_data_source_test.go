package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccRoleDataSource_existing(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Skipping test: TF_ACC environment variable not set")
	}

	expectedRoles := []string{"privx-user", "privx-admin"}

	for _, role := range expectedRoles {
		t.Run(fmt.Sprintf("Verify role: %s", role), func(t *testing.T) {
			//t.Logf("Verifying existence of role: %q", role)

			resource.Test(t, resource.TestCase{
				PreCheck: func() {
					testAccPreCheck(t)
				},
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: generateRoleDataSourceConfig(role),
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttr(
								"data.privx_role.existing", "name", role,
							),
						),
					},
				},
			})
		})
	}
}

// generateRoleDataSourceConfig creates the Terraform configuration for the test.
func generateRoleDataSourceConfig(roleName string) string {
	return fmt.Sprintf(`

terraform {
  required_providers {
    privx = {
      source = "hashicorp/privx"
    }
  }
}
provider "privx" {}

data "privx_role" "existing" {
  name = "%s"
}
		`, roleName)
}
