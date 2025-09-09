resource "privx_workflow" "qrt-test" {
  name        = "qrt-test1"
  grant_types = ["PERMANENT"]
  target_roles = [
    {
      "id" : "920bb54e-dec9-575a-7309-86d12737d016",
      "name" : "qtr-test"
    }
  ]
  requester_roles = [ // Optional
    {
      "id" : "e2dcbeb4-6b07-50ad-788a-5af830da74ca",
      "name" : "Linux-admin"
    },
    {
      "id" : data.privx_role.privx-admin.id,
      "name" : data.privx_role.privx-admin.name
    }
  ]
  action = "GRANT"
  steps = [
    {
      "name" : "First Approval",
      "match" : "ANY",
      "approvers" : [
        {
          "role" : {
            "id" : data.privx_role.privx-admin.id,
            "name" : data.privx_role.privx-admin.name
          }
        }
      ]
    }
  ]
}
