resource "privx_access_group" "example" {
  name    = "example-target-group"
  comment = "Access group for network target"
}

resource "privx_role" "example" {
  name            = "example-target-role"
  access_group_id = privx_access_group.example.id

  permissions  = ["users-view"]
  permit_agent = false

  source_rules = jsonencode({
    type  = "GROUP"
    match = "ANY"
    rules = []
  })
}

resource "privx_network_target" "example" {
  name     = "example-network-target"
  disabled = false

  roles {
    id = privx_role.example.id
  }

  dst {
    ip_start = "192.168.1.1"
    ip_end   = "192.168.1.50"
    protocol = "tcp"
  }
}

