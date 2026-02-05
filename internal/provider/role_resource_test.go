package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/SSHcom/privx-sdk-go/v2/api/rolestore"
	"github.com/SSHcom/privx-sdk-go/v2/oauth"
	"github.com/SSHcom/privx-sdk-go/v2/restapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccRoleResource_recreateAfterOutOfBandDelete(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC not set")
	}

	suffix := acctest.RandStringFromCharSet(6, acctest.CharSetAlphaNum)
	roleName := fmt.Sprintf("tf-acc-test-role-with-group-%s", suffix)
	groupName := fmt.Sprintf("tf-acc-access-group-%s", suffix)

	cfg := testAccRoleWithAccessGroupConfig(groupName, roleName)
	writeAccConfig(t, fmt.Sprintf("%s_step_1.tf", t.Name()), cfg)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,

		Steps: []resource.TestStep{
			// 1) Create
			{
				Config: cfg,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("privx_access_group.foo", "name", groupName),
					resource.TestCheckResourceAttr("privx_role.test", "name", roleName),
					resource.TestCheckResourceAttrPair(
						"privx_role.test", "access_group_id",
						"privx_access_group.foo", "id",
					),
					resource.TestCheckResourceAttr("privx_role.test", "permissions.#", "1"),
					resource.TestCheckTypeSetElemAttr("privx_role.test", "permissions.*", "users-view"),
				),
			},

			// 2) Delete out-of-band. This step *will* leave drift, so allow non-empty plan.
			//    (This prevents the harness from failing the step because refresh produces +create.)
			{
				Config:             cfg,
				ExpectNonEmptyPlan: true,
				Check:              deleteRoleOutOfBand(t, "privx_role.test"),
			},

			// 3) Now explicitly verify drift via PlanOnly
			{
				Config:             cfg,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},

			// 4) Apply again should recreate
			{
				Config: cfg,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("privx_role.test", "name", roleName),
					resource.TestCheckResourceAttrSet("privx_role.test", "id"),
				),
			},
		},
	})

}

func testAccRoleWithAccessGroupConfig(groupName, roleName string) string {
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
  comment = "temp access group for role resource test"
}

resource "privx_role" "test" {
  name            = %q
  access_group_id = privx_access_group.foo.id

  permissions  = ["users-view"]
  permit_agent = false

  source_rules = jsonencode({
    type  = "GROUP"
    match = "ANY"
    rules = []
  })
}
`, groupName, roleName)
}

func deleteRoleOutOfBand(t *testing.T, addr string) resource.TestCheckFunc {
	t.Helper()

	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[addr]
		if !ok {
			return fmt.Errorf("not found in state: %s", addr)
		}
		roleID := rs.Primary.ID
		if roleID == "" {
			return fmt.Errorf("empty role id for %s", addr)
		}

		conn := testAccPrivxConnectorFromSDKEnv()
		client := rolestore.New(conn)

		// Simulate UI delete
		if err := client.DeleteRole(roleID); err != nil {
			return fmt.Errorf("DeleteRole(%s) failed: %w", roleID, err)
		}

		return nil
	}
}

// Uses PrivX SDK's built-in environment configuration.
// Requires env vars like PRIVX_API_BASE_URL, PRIVX_API_CLIENT_ID, PRIVX_API_CLIENT_SECRET, etc.
func testAccPrivxConnectorFromSDKEnv() restapi.Connector {
	// Authorizer reads auth env vars via oauth.UseEnvironment()
	authBase := restapi.New(restapi.UseEnvironment())
	authorizer := oauth.With(
		authBase,
		oauth.UseEnvironment(),
	)

	// Connector reads base URL (and optional TLS CA) via restapi.UseEnvironment()
	return restapi.New(
		restapi.Auth(authorizer),
		restapi.UseEnvironment(),
	)
}

// Optional: if you want the test to fail fast with a clearer message,
// call this from testAccPreCheck(t) or directly in the test.
/*func requirePrivxSDKEnv(t *testing.T) {
	t.Helper()
	required := []string{
		"PRIVX_API_BASE_URL",
		"PRIVX_API_CLIENT_ID",
		"PRIVX_API_CLIENT_SECRET",
	}
	for _, k := range required {
		if os.Getenv(k) == "" {
			t.Fatalf("%s must be set for TF_ACC out-of-band delete tests", k)
		}
	}
	_ = context.Background()
}*/
