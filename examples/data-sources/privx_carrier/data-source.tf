# Get carrier by name
data "privx_carrier" "example" {
  name = "example-carrier"
}

# Output the carrier information
output "carrier_info" {
  value = {
    id                                = data.privx_carrier.example.id
    name                              = data.privx_carrier.example.name
    access_group_id                   = data.privx_carrier.example.access_group_id
    group_id                          = data.privx_carrier.example.group_id
    enabled                           = data.privx_carrier.example.enabled
    registered                        = data.privx_carrier.example.registered
    permissions                       = data.privx_carrier.example.permissions
    routing_prefix                    = data.privx_carrier.example.routing_prefix
    extender_address                  = data.privx_carrier.example.extender_address
    subnets                           = data.privx_carrier.example.subnets
    web_proxy_address                 = data.privx_carrier.example.web_proxy_address
    web_proxy_extender_route_patterns = data.privx_carrier.example.web_proxy_extender_route_patterns
  }
}

# Output sensitive information separately (be careful with sensitive data)
output "carrier_secret" {
  value     = data.privx_carrier.example.secret
  sensitive = true
}