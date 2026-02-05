package provider

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccHostResource_passwordRotation(t *testing.T) {
	resourceName := "privx_host.test"

	suffix := acctest.RandStringFromCharSet(6, "abcdefghijklmnopqrstuvwxyz0123456789")
	externalID := "tf-ext-" + suffix

	octet := acctest.RandIntRange(1, 250)
	hostIP := fmt.Sprintf("192.0.2.%d", octet)

	commonName := "tf-acc-host-rotation-" + suffix

	cfgStep1 := testAccHostPasswordRotationConfig_step1_withExternalID(externalID, hostIP, commonName)
	cfgStep2 := testAccHostPasswordRotationConfig_step2_omitExternalID_changeComment(hostIP, commonName)

	writeAccConfig(t, fmt.Sprintf("%s_step_1.tf", t.Name()), cfgStep1)
	writeAccConfig(t, fmt.Sprintf("%s_step_2.tf", t.Name()), cfgStep2)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:             cfgStep1,
				ExpectNonEmptyPlan: false,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "external_id", externalID),

					resource.TestCheckResourceAttr(resourceName, "password_rotation_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "password_rotation.use_main_account", "true"),
					resource.TestCheckResourceAttr(resourceName, "password_rotation.operating_system", "LINUX"),
					resource.TestCheckResourceAttr(resourceName, "password_rotation.protocol", "SSH"),

					resource.TestCheckResourceAttrPair(
						resourceName, "password_rotation.password_policy_id",
						"data.privx_password_policy.pp", "id",
					),
					resource.TestCheckResourceAttrPair(
						resourceName, "password_rotation.script_template_id",
						"data.privx_script_template.st", "id",
					),

					resource.TestCheckResourceAttr(resourceName, "password_rotation.certificate_validation_options", "DISABLED"),

					testAccCheckHostHasPrincipal(resourceName, "tf-acc-rotate"),
					testAccCheckHostPrincipalBool(resourceName, "tf-acc-rotate", "rotate", "true"),
					testAccCheckHostPrincipalBool(resourceName, "tf-acc-rotate", "use_for_password_rotation", "true"),

					testAccCheckHostHasService(resourceName, "SSH"),
				),
			},
			{
				// Step 2: omit external_id; change comment to force Update()
				Config:             cfgStep2,
				ExpectNonEmptyPlan: false,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "external_id", externalID),
					resource.TestCheckResourceAttr(resourceName, "comment", "step2-update"),
				),
			},
			{
				ResourceName: resourceName,
				ImportState:  true,
			},
		},
	})
}

func testAccCheckHostHasPrincipal(name, principal string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("resource not found: %s", name)
		}

		for k, v := range rs.Primary.Attributes {
			if k == "principals.#" {
				continue
			}
			if strings.HasSuffix(k, ".principal") && v == principal {
				return nil
			}
		}
		return fmt.Errorf("principal %q not found", principal)
	}
}

func testAccCheckHostPrincipalBool(name, principal, field string, expected string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("resource not found: %s", name)
		}

		for k, v := range rs.Primary.Attributes {
			if strings.HasSuffix(k, ".principal") && v == principal {
				prefix := strings.TrimSuffix(k, ".principal")
				key := prefix + "." + field
				if rs.Primary.Attributes[key] == expected {
					return nil
				}
				return fmt.Errorf("principal %q field %q expected %q, got %q",
					principal, field, expected, rs.Primary.Attributes[key])
			}
		}
		return fmt.Errorf("principal %q not found", principal)
	}
}

func testAccCheckHostHasService(name, service string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("resource not found: %s", name)
		}

		for k, v := range rs.Primary.Attributes {
			if strings.HasSuffix(k, ".service") && v == service {
				return nil
			}
		}
		return fmt.Errorf("service %q not found", service)
	}
}

func testAccHostPasswordRotationConfig_step1_withExternalID(externalID, hostIP, commonName string) string {
	// Use indexed formatting so args can't "shift"
	// 1 = hostIP, 2 = externalID, 3 = commonName
	return fmt.Sprintf(`
terraform {
  required_providers {
    privx = {
      source = "hashicorp/privx"
    }
  }
}

provider "privx" {}

data "privx_script_template" "st" {
  name = "Linux per account command template"
}

data "privx_password_policy" "pp" {
  name = "PrivX default password policy"
}

data "privx_access_group" "ag" {
  name = "Default"
}

resource "privx_host" "test" {
  common_name     = %[3]q
  addresses       = [%[1]q]
  access_group_id = data.privx_access_group.ag.id

  external_id = %[2]q
  comment     = "step1-create"

  services = [{
    service                   = "SSH"
    address                   = %[1]q
    port                      = 22
    use_for_password_rotation = true
  }]

  password_rotation_enabled = true

  password_rotation = {
    access_group_id                = data.privx_access_group.ag.id
    use_main_account               = true
    operating_system               = "LINUX"
    protocol                       = "SSH"
    certificate_validation_options = "DISABLED"
    password_policy_id             = data.privx_password_policy.pp.id
    script_template_id             = data.privx_script_template.st.id
  }

  principals = [{
    principal                 = "tf-acc-rotate"
    rotate                    = true
    use_for_password_rotation = true
    use_user_account          = false
    passphrase                = "InitPaSS-123"
  }]
}
`, hostIP, externalID, commonName)
}

func testAccHostPasswordRotationConfig_step2_omitExternalID_changeComment(hostIP, commonName string) string {
	// Same resource identity: common_name + addresses + service address remain the same.
	return fmt.Sprintf(`
terraform {
  required_providers {
    privx = {
      source = "hashicorp/privx"
    }
  }
}

provider "privx" {}

data "privx_script_template" "st" {
  name = "Linux per account command template"
}

data "privx_password_policy" "pp" {
  name = "PrivX default password policy"
}

data "privx_access_group" "ag" {
  name = "Default"
}

resource "privx_host" "test" {
  common_name     = %[2]q
  addresses       = [%[1]q]
  access_group_id = data.privx_access_group.ag.id

  # external_id intentionally NOT set here
  comment = "step2-update"

  services = [{
    service                   = "SSH"
    address                   = %[1]q
    port                      = 22
    use_for_password_rotation = true
  }]

  password_rotation_enabled = true

  password_rotation = {
    access_group_id                = data.privx_access_group.ag.id
    use_main_account               = true
    operating_system               = "LINUX"
    protocol                       = "SSH"
    certificate_validation_options = "DISABLED"
    password_policy_id             = data.privx_password_policy.pp.id
    script_template_id             = data.privx_script_template.st.id
  }

  principals = [{
    principal                 = "tf-acc-rotate"
    rotate                    = true
    use_for_password_rotation = true
    use_user_account          = false
    passphrase                = "InitPaSS-123"
  }]
}
`, hostIP, commonName)
}
