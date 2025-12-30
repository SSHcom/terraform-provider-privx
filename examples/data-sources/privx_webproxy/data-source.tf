# Get webproxy by name
data "privx_webproxy" "example" {
  name = "example-webproxy"
}

# Output the webproxy information
output "webproxy_info" {
  value = {
    id                                = data.privx_webproxy.example.id
    name                              = data.privx_webproxy.example.name
    access_group_id                   = data.privx_webproxy.example.access_group_id
    group_id                          = data.privx_webproxy.example.group_id
    enabled                           = data.privx_webproxy.example.enabled
    registered                        = data.privx_webproxy.example.registered
    permissions                       = data.privx_webproxy.example.permissions
    routing_prefix                    = data.privx_webproxy.example.routing_prefix
    extender_address                  = data.privx_webproxy.example.extender_address
    subnets                           = data.privx_webproxy.example.subnets
    web_proxy_address                 = data.privx_webproxy.example.web_proxy_address
    web_proxy_extender_route_patterns = data.privx_webproxy.example.web_proxy_extender_route_patterns
  }
}

# Output sensitive information separately (be careful with sensitive data)
output "webproxy_secret" {
  value     = data.privx_webproxy.example.secret
  sensitive = true
}