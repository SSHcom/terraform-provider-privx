package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccHostDataSource_readsCreatedHost(t *testing.T) {
	suffix := acctest.RandStringFromCharSet(6, acctest.CharSetAlphaNum)
	commonName := fmt.Sprintf("tf-acc-host-ds-%s", suffix)

	octet := acctest.RandIntRange(1, 250)
	hostIP := fmt.Sprintf("192.0.2.%d", octet)

	cfg := testAccHostDataSourceConfig(commonName, hostIP)

	writeAccConfig(t, fmt.Sprintf("%s.tf", t.Name()), cfg)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: cfg,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Lookup by ID returns the created host.
					resource.TestCheckResourceAttrPair(
						"data.privx_host.by_id", "id",
						"privx_host.test", "id",
					),
					resource.TestCheckResourceAttr(
						"data.privx_host.by_id", "common_name", commonName,
					),
					// Lookup by common_name resolves to the same host.
					resource.TestCheckResourceAttrPair(
						"data.privx_host.by_name", "id",
						"privx_host.test", "id",
					),
					resource.TestCheckResourceAttr(
						"data.privx_host.by_name", "common_name", commonName,
					),
				),
			},
		},
	})
}

func testAccHostDataSourceConfig(commonName, hostIP string) string {
	// 1 = commonName, 2 = hostIP
	return fmt.Sprintf(`
provider "privx" {}

data "privx_access_group" "ag" {
  name = "Default"
}

resource "privx_host" "test" {
  common_name     = %[1]q
  addresses       = [%[2]q]
  access_group_id = data.privx_access_group.ag.id

  comment = "created by host data source acc test"

  services = [{
    service = "SSH"
    address = %[2]q
    port    = 22
  }]
}

data "privx_host" "by_id" {
  id = privx_host.test.id
}

data "privx_host" "by_name" {
  common_name = privx_host.test.common_name
}
`, commonName, hostIP)
}
