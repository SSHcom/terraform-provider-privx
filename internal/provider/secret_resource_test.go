package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccSecretResource_basicUpdateRenameImport(t *testing.T) {
	secret1 := "tf-acc-secret-" + acctest.RandString(6)
	secret2 := secret1 + "-renamed"
	resourceName := "privx_secret.test"

	cfg1 := testAccSecretConfigCreate(secret1, "InitPaSS-123")
	cfg2 := testAccSecretConfigCreate(secret1, "Changed-456")
	cfg3 := testAccSecretConfigCreate(secret2, "Changed-456")

	writeAccConfig(t, fmt.Sprintf("%s_step_1.tf", t.Name()), cfg1)
	writeAccConfig(t, fmt.Sprintf("%s_step_2.tf", t.Name()), cfg2)
	writeAccConfig(t, fmt.Sprintf("%s_step_3.tf", t.Name()), cfg3)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: cfg1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", secret1),
					resource.TestCheckResourceAttr(resourceName, "read_roles.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "write_roles.#", "1"),
				),
			},
			{
				Config: cfg2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", secret1),
				),
			},
			{
				Config: cfg3,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", secret2),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     secret2, // because ImportState maps to "name"
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"data", // sensitive
				},
			},
		},
	})
}

func testAccSecretConfigCreate(secretName, password string) string {
	return fmt.Sprintf(`
terraform {
  required_providers {
    privx = {
      source = "hashicorp/privx"
    }
  }
}
provider "privx" {}

data "privx_role" "tf" {
  name = "terraform-provider"
}

resource "privx_secret" "test" {
  name = %q

  data = {
    username = "alice"
    password = %q
  }

  read_roles = [{
    id = data.privx_role.tf.id
  }]

  write_roles = [{
    id = data.privx_role.tf.id
  }]
}
`, secretName, password)
}
