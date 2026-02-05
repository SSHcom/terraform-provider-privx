package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccApiProxyCredentialResource(t *testing.T) {
	testAPIHost := "kubernetes.example.com:6443"
	apiTargetName := acctest.RandomWithPrefix("tf-kube")

	// local user fields
	username := acctest.RandomWithPrefix("tf-acc-user")
	password := "InitPaSS-123" // or generate if you prefer

	targetRes := "privx_api_target.kube"
	credRes := "privx_api_proxy_credential.kube"
	proxyDS := "data.privx_api_proxy_config.this"
	userRes := "privx_local_user.test"

	cfg := testAccApiProxyCredentialConfig(testAPIHost, apiTargetName, username, password)
	writeAccConfig(t, fmt.Sprintf("%s_step_1.tf", t.Name()), cfg)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: cfg,
				Check: resource.ComposeTestCheckFunc(
					// local user
					resource.TestCheckResourceAttrSet(userRes, "id"),

					// api_target
					resource.TestCheckResourceAttrSet(targetRes, "id"),
					resource.TestCheckResourceAttr(targetRes, "name", apiTargetName),

					// api_proxy_credential (for that user)
					resource.TestCheckResourceAttrSet(credRes, "id"),
					resource.TestCheckResourceAttrPair(credRes, "user_id", userRes, "id"),
					resource.TestCheckResourceAttr(credRes, "name", apiTargetName),
					resource.TestCheckResourceAttrPair(credRes, "target_id", targetRes, "id"),
					resource.TestCheckResourceAttrSet(credRes, "secret"),

					// api_proxy_config data source
					resource.TestCheckResourceAttrSet(proxyDS, "id"),
					resource.TestCheckResourceAttrSet(proxyDS, "addresses.0"),
					resource.TestMatchResourceAttr(proxyDS, "addresses.0", regexp.MustCompile(`^https?://`)),
					resource.TestCheckResourceAttrSet(proxyDS, "ca_certificate_chain"),
					resource.TestMatchResourceAttr(proxyDS, "ca_certificate_chain", regexp.MustCompile(`-----BEGIN CERTIFICATE-----`)),
				),
			},
			{
				ResourceName:      credRes,
				ImportState:       true,
				ImportStateIdFunc: importUserCredID(userRes, credRes),
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"secret",
				},
			},
		},
	})
}

func importUserCredID(userRes, credRes string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		cred, ok := s.RootModule().Resources[credRes]
		if !ok {
			return "", fmt.Errorf("not found: %s", credRes)
		}
		user, ok := s.RootModule().Resources[userRes]
		if !ok || user == nil {
			return "", fmt.Errorf("not found: %s", userRes)
		}
		return fmt.Sprintf("%s/%s", user.Primary.ID, cred.Primary.ID), nil
	}
}

func testAccApiProxyCredentialConfig(testAPIHost, apiTargetName, username, password string) string {
	return fmt.Sprintf(`
terraform {
  required_providers {
    privx = {
      source = "hashicorp/privx"
    }
  }
}

provider "privx" {}

data "privx_role" "users" {
  name = "privx-user"
}

resource "privx_local_user" "test" {
  username  = %q
  full_name = "Terraform provider AccTest User"
  job_title = "worker"
  email     = %q
  password  = %q
  password_change_required = false
  tags = ["team-a", "oncall"]
}

resource "privx_api_target" "kube" {
  name    = %q
  comment = "test"

  roles = [
    {
      id   = data.privx_role.users.id
      name = data.privx_role.users.name
    }
  ]

  authorized_endpoints = [
    {
      host      = %q
      protocols = ["https"]
      methods   = ["*"]
      paths     = ["**"]
    }
  ]

  target_credential = {
    type         = "token"
    bearer_token = "TEST_K8S_DUMMY_TOKEN"
  }

  audit_enabled = false
}

resource "privx_api_proxy_credential" "kube" {
  user_id   = privx_local_user.test.id
  name      = %q
  target_id = privx_api_target.kube.id

  // Use far-future so test doesn't start failing as time passes
  not_before = "2099-01-08T00:00:00Z"
  not_after  = "2099-01-09T00:00:00Z"

  enabled        = true
  type           = "token"
  comment        = "test"
  source_address = []
}

data "privx_api_proxy_config" "this" {}
`, username, fmt.Sprintf("%s@test.local", username), password, apiTargetName, testAPIHost, apiTargetName)
}
