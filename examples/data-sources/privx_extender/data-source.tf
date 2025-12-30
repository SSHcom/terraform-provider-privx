# Get extender by name
data "privx_extender" "example" {
  name = "example-extender"
}

# Output the extender information
output "extender_info" {
  value = {
    id              = data.privx_extender.example.id
    name            = data.privx_extender.example.name

    access_group_id = data.privx_extender.example.access_group_id
    enabled         = data.privx_extender.example.enabled
    registered      = data.privx_extender.example.registered
    permissions     = data.privx_extender.example.permissions
    routing_prefix  = data.privx_extender.example.routing_prefix
    extender_address = data.privx_extender.example.extender_address
    subnets         = data.privx_extender.example.subnets
  }
}

# Output sensitive information separately (be careful with sensitive data)
output "extender_secret" {
  value     = data.privx_extender.example.secret
  sensitive = true
}