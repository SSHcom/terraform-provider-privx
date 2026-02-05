package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccWorkflowResource(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC not set")
	}

	// These must exist in the PrivX environment (data sources)
	targetRoleName := "privx-user"
	requesterRoleName := "linux-admin"
	adminRoleName := "privx-admin"

	suffix := acctest.RandStringFromCharSet(8, "abcdefghijklmnopqrstuvwxyz0123456789")

	workflowName := "tf-acc-workflow-" + suffix
	workflowNameUpdated := workflowName + "-updated"
	groupName := "tf-acc-wf-group-" + suffix

	// IMPORTANT: created by this test -> must be unique
	targetRoleResourceName := "tf-acc-wf-target-role-" + suffix

	cfg := testAccWorkflowConfigByRoleNames(
		workflowName,
		groupName,
		targetRoleResourceName,
		targetRoleName,
		requesterRoleName,
		adminRoleName,
		"ANY",
		1,
	)
	writeAccConfig(t, fmt.Sprintf("%s_step_1.tf", t.Name()), cfg)

	cfgUpdate := testAccWorkflowConfigByRoleNames(
		workflowNameUpdated,
		groupName,
		targetRoleResourceName, // keep same role across update
		targetRoleName,
		requesterRoleName,
		adminRoleName,
		"ALL",
		2,
	)
	writeAccConfig(t, fmt.Sprintf("%s_step_2.tf", t.Name()), cfgUpdate)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,

		Steps: []resource.TestStep{
			{
				// Step 1: Create
				Config: cfg,
				Check: resource.ComposeTestCheckFunc(
					// data sources resolved (existing roles)
					resource.TestCheckResourceAttr("data.privx_role.existing_target", "name", targetRoleName),
					resource.TestCheckResourceAttr("data.privx_role.requester", "name", requesterRoleName),
					resource.TestCheckResourceAttr("data.privx_role.admin", "name", adminRoleName),

					// workflow exists
					resource.TestCheckResourceAttr("privx_workflow.test", "name", workflowName),
					resource.TestCheckResourceAttr("privx_workflow.test", "action", "GRANT"),
					resource.TestCheckResourceAttrSet("privx_workflow.test", "id"),

					// ensure workflow targets the created role
					resource.TestCheckTypeSetElemAttrPair(
						"privx_workflow.test",
						"target_roles.*.id",
						"privx_role.target",
						"id",
					),

					// Step 1 defaults (so update step can prove a change)
					resource.TestCheckResourceAttr("privx_workflow.test", "max_active_requests", "1"),
					resource.TestCheckResourceAttr("privx_workflow.test", "steps.0.name", "First Approval"),
					resource.TestCheckResourceAttr("privx_workflow.test", "steps.0.match", "ANY"),
				),
			},
			{
				// Step 2: Update
				Config: cfgUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("privx_workflow.test", "max_active_requests", "2"),
					resource.TestCheckResourceAttr("privx_workflow.test", "steps.0.name", "First Approval"),
					resource.TestCheckResourceAttr("privx_workflow.test", "steps.0.match", "ALL"),
				),
			},
			{
				// Step 3: Import
				ResourceName:      "privx_workflow.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccWorkflowConfigByRoleNames(
	workflowName string,
	groupName string,
	targetRoleResourceName string, // created by test, must be unique
	targetRoleName string, // existing role in env for data source check
	requesterRoleName string,
	adminRoleName string,
	stepMatch string,
	maxActiveRequests int,
) string {
	// Use indexed formatting so argument order can never "shift" again.
	return fmt.Sprintf(`
terraform {
  required_providers {
    privx = {
      source = "hashicorp/privx"
    }
  }
}

provider "privx" {}

resource "privx_access_group" "acc" {
  name    = %[1]q
  comment = "temp group for workflow test"
}

resource "privx_role" "target" {
  name            = %[2]q
  access_group_id = privx_access_group.acc.id

  permissions  = ["users-view"]
  permit_agent = false

  source_rules = jsonencode({
    type  = "GROUP"
    match = "ANY"
    rules = []
  })
}

# Existing roles looked up by name
data "privx_role" "existing_target" {
  name = %[3]q
}

data "privx_role" "requester" {
  name = %[4]q
}

data "privx_role" "admin" {
  name = %[5]q
}

resource "privx_workflow" "test" {
  name        = %[6]q
  grant_types = ["PERMANENT"]
  max_active_requests = %[7]d

  target_roles = [{
    id   = privx_role.target.id
    name = privx_role.target.name
  }]

  requester_roles = [
    {
      id   = data.privx_role.requester.id
      name = data.privx_role.requester.name
    },
    {
      id   = data.privx_role.admin.id
      name = data.privx_role.admin.name
    }
  ]

  action = "GRANT"

  steps = [
    {
      name  = "First Approval"
      match = %[8]q
      approvers = [
        {
          role = {
            id   = data.privx_role.admin.id
            name = data.privx_role.admin.name
          }
        }
      ]
    }
  ]
}
`,
		groupName,              // 1
		targetRoleResourceName, // 2
		targetRoleName,         // 3
		requesterRoleName,      // 4
		adminRoleName,          // 5
		workflowName,           // 6
		maxActiveRequests,      // 7
		stepMatch,              // 8
	)
}
