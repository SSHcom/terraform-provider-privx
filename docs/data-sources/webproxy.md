# privx_webproxy (Data Source)

Use this data source to get information about a PrivX webproxy client.

## Example Usage

```terraform
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
    enabled                           = data.privx_webproxy.example.enabled
    registered                        = data.privx_webproxy.example.registered
    permissions                       = data.privx_webproxy.example.permissions
    routing_prefix                    = data.privx_webproxy.example.routing_prefix
    web_proxy_address                 = data.privx_webproxy.example.web_proxy_address
    web_proxy_extender_route_patterns = data.privx_webproxy.example.web_proxy_extender_route_patterns
  }
}
```

## Schema

### Required

- `name` (String) WebProxy name

### Read-Only

- `id` (String) WebProxy UUID
- `access_group_id` (String) Access Group ID
- `group_id` (String) Group ID for the webproxy
- `enabled` (Boolean) Whether the webproxy is enabled
- `registered` (Boolean) Whether the webproxy is registered
- `permissions` (List of String) List of permissions for the webproxy
- `routing_prefix` (String) Routing prefix for the webproxy
- `extender_address` (List of String) List of extender addresses
- `subnets` (List of String) List of subnets for the webproxy
- `web_proxy_address` (String) Web proxy address for the webproxy
- `web_proxy_extender_route_patterns` (List of String) List of web proxy extender route patterns
- `secret` (String, Sensitive) Client Secret

**Note:** The webproxy name must match an existing webproxy with type "ICAP" in PrivX.