package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccLocalUserResource_withPasswordRotation(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC not set")
	}

	suffix := acctest.RandStringFromCharSet(6, acctest.CharSetAlphaNum)
	username := fmt.Sprintf("tf-acc-user-%s", suffix)
	initialPassword := "InitPaSS-123"
	rotatedPassword := "R0tatePaSS-456"

	cfg1 := testAccLocalUserInitialPassword(username, initialPassword)
	writeAccConfig(t, fmt.Sprintf("%s_step_1.tf", t.Name()), cfg1)
	cfg2 := testAccLocalUserRotatedPassword(username, rotatedPassword)
	writeAccConfig(t, fmt.Sprintf("%s_step_2.tf", t.Name()), cfg2)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			t.Helper()
			t.Log("PreCheck: validating acceptance test environment")
			testAccPreCheck(t)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Step 1: create user with initial password
				Config: cfg1,
				Check: func(_ *terraform.State) error {
					t.Logf("Step 1: created user %q with initial password (not printed)", username)
					// time.Sleep(30 * time.Second) // login manually if desired
					return nil
				},
			},
			{
				// Step 2: rotate password
				Config: cfg2,
				Check: func(_ *terraform.State) error {
					t.Logf("Step 2: rotated password for user %q (new password not printed)", username)
					// time.Sleep(30 * time.Second) // login manually if desired
					return nil
				},
			},
		},
	})
}

func testAccLocalUserInitialPassword(username, password string) string {
	return fmt.Sprintf(`

terraform {
  required_providers {
    privx = {
      source = "hashicorp/privx"
    }
  }
}
provider "privx" {}

resource "privx_local_user" "test" {
  username  = "%s"
  full_name = "Terraform provider Test User"
  job_title = "worker"
  email     = "%s@example.com"
  password  = "%s"
  password_change_required = true
  tags = ["team-a", "oncall"]
}
`, username, username, password)
}

func testAccLocalUserRotatedPassword(username, rotatedPassword string) string {
	return fmt.Sprintf(`
terraform {
  required_providers {
    privx = {
      source = "hashicorp/privx"
    }
  }
}
provider "privx" {}

resource "privx_local_user" "test" {
  username  = "%s"
  full_name = "Terraform provider Test User"
  job_title = "worker"
  email     = "%s@example.com"
  password_change_required = false
  tags = ["team-a", "oncall"]
}

resource "privx_local_user_password" "rotate" {
  user_id  = privx_local_user.test.id
  password = "%s"
}
`, username, username, rotatedPassword)
}
