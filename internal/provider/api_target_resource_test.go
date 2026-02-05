package provider

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccApiTargetResource(t *testing.T) {

	testAPIHost := "kubernetes.example.com:6443"
	testAPIHostBearerToken := "API_TARGET_TEST_DUMMY_TOKEN"
	suffix := acctest.RandStringFromCharSet(6, acctest.CharSetAlphaNum)
	apiTargetName := fmt.Sprintf("tf-api-target-%s", suffix)
	accGroupName := "Default"

	trustPEM := genSelfSignedCAPEM(t)
	trustPath := writeTempPEM(t, trustPEM)

	cfgCreate := testAccApiTargetConfig(
		"privx-user",
		accGroupName,
		apiTargetName,
		"created by acc test",
		testAPIHost,
		testAPIHostBearerToken,
		trustPath,
	)
	writeAccConfig(t, fmt.Sprintf("%s_step_1.tf", t.Name()), cfgCreate)

	cfgUpdate := testAccApiTargetConfig(
		"privx-user",
		accGroupName,
		apiTargetName,
		"updated by acc test",
		testAPIHost,
		testAPIHostBearerToken,
		trustPath,
	)

	writeAccConfig(t, fmt.Sprintf("%s_step_2.tf", t.Name()), cfgUpdate)

	resourceName := "privx_api_target.k8s"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: cfgCreate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", apiTargetName),
					resource.TestCheckResourceAttr(resourceName, "comment", "created by acc test"),

					resource.TestCheckResourceAttr(resourceName, "authorized_endpoints.0.host", testAPIHost),

					resource.TestCheckResourceAttr(resourceName, "target_credential.type", "token"),

					resource.TestCheckResourceAttrSet(resourceName, "roles.0.id"),
					resource.TestCheckResourceAttr(resourceName, "tls_insecure_skip_verify", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "access_group_id"),

					resource.TestCheckNoResourceAttr(resourceName, "target_credential.basic_auth_password"),
					resource.TestCheckNoResourceAttr(resourceName, "target_credential.certificate"),
					resource.TestCheckNoResourceAttr(resourceName, "target_credential.private_key"),
				),
			},
			{
				// ✅ Refresh-only step: no Config allowed
				RefreshState: true,
			},
			{
				// Update to exercise Update()
				Config: cfgUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "comment", "updated by acc test"),
				),
			},
			{
				// Import test
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"target_credential.bearer_token",
					"tls_trust_anchors",
					"target_credential.type",
				},
			},
			{
				Config: cfgUpdate,
				// No changes; must be a no-op apply and must not error.
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "comment", "updated by acc test"),
					resource.TestCheckResourceAttr(resourceName, "target_credential.type", "token"),
				),
			},
		},
	})
}

func genSelfSignedCAPEM(t *testing.T) string {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		t.Fatalf("serial: %v", err)
	}

	tmpl := x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			CommonName: "tf-acc-test-ca",
		},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	der, err := x509.CreateCertificate(rand.Reader, &tmpl, &tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("create cert: %v", err)
	}

	return string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}))
}

func writeTempPEM(t *testing.T, pemStr string) string {
	t.Helper()

	dir := t.TempDir()
	p := filepath.Join(dir, "trust-anchor.pem")
	if err := os.WriteFile(p, []byte(pemStr), 0o600); err != nil {
		t.Fatalf("write pem: %v", err)
	}
	return p
}

func testAccApiTargetConfig(roleName, accGroupName, apiTargetName, comment string, testAPIHost string, testAPIHostBearerToken string, trustAnchorPath string) string {
	return fmt.Sprintf(`
terraform {
  required_providers {
    privx = {
      source = "hashicorp/privx"
    }
  }
}

provider "privx" {}

data "privx_role" "existing_role" {
  name = %q
}

data "privx_access_group" "foo" {
  name = %q
}

resource "privx_api_target" "k8s" {
  name    = %q
  comment = %q

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
    type               = "token"
    bearer_token       = %q

    # Stress-test nested sensitive state handling:
	# provider must not return Unknown after apply and must keep object stable.

    basic_auth_password = null
    certificate         = null
    private_key         = null
  }

  tls_trust_anchors = file(%q)
  tls_insecure_skip_verify = false
}
`,
		roleName,
		accGroupName,
		apiTargetName,
		comment,
		testAPIHost,
		testAPIHostBearerToken,
		trustAnchorPath,
	)
}
