package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccApiTargetDataSource_readsCreatedTarget(t *testing.T) {
	testAPIHost := "kubernetes.example.com:6443"
	testAPIHostBearerToken := "API_TARGET_TEST_DUMMY_TOKEN"
	suffix := acctest.RandStringFromCharSet(6, acctest.CharSetAlphaNum)
	apiTargetName := fmt.Sprintf("tf-acc-api-target-%s", suffix)
	accGroupName := "Default"

	trustPEM := genSelfSignedCAPEM(t)
	trustPath := writeTempPEM(t, trustPEM)

	cfg := fmt.Sprintf(`
provider "privx" {}

data "privx_role" "existing_role" {
  name = "privx-user"
}

data "privx_access_group" "foo" {
  name = %q
}

resource "privx_api_target" "k8s" {
  name    = %q
  comment = "created by acc test"

  roles = [{
    id   = data.privx_role.existing_role.id
    name = data.privx_role.existing_role.name
  }]

  access_group_id = data.privx_access_group.foo.id

  authorized_endpoints = [{
    host      = %q
    protocols = ["*"]
    methods   = ["*"]
    paths     = ["**"]
  }]

  target_credential = {
    type         = "token"
    bearer_token = %q
  }

  tls_trust_anchors = file(%q)
  tls_insecure_skip_verify = false
}

data "privx_api_target" "by_name" {
  name = privx_api_target.k8s.name
}
`, accGroupName, apiTargetName, testAPIHost, testAPIHostBearerToken, trustPath)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: cfg,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.privx_api_target.by_name", "id", "privx_api_target.k8s", "id"),
					resource.TestCheckResourceAttr("data.privx_api_target.by_name", "name", apiTargetName),

					// ✅ print TF state for data source + resource
					//testAccPrintApiTargetState("data.privx_api_target.by_name"),
					//testAccPrintApiTargetState("privx_api_target.k8s"),
				),
			},
		},
	})
}

/*
func testAccPrintApiTargetState(addr string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[addr]
		if !ok {
			return fmt.Errorf("resource not found in state: %s", addr)
		}

		fmt.Println("──────────── API TARGET STATE ────────────")
		for k, v := range rs.Primary.Attributes {
			fmt.Printf("%s = %s\n", k, v)
		}
		fmt.Println("──────────────────────────────────────────")

		return nil
	}
}*/
