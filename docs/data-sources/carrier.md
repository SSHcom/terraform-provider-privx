# privx_carrier (Data Source)

Use this data source to get information about a PrivX carrier client.

## Example Usage

```terraform
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
    enabled                           = data.privx_carrier.example.enabled
    registered                        = data.privx_carrier.example.registered
    permissions                       = data.privx_carrier.example.permissions
    routing_prefix                    = data.privx_carrier.example.routing_prefix
    web_proxy_address                 = data.privx_carrier.example.web_proxy_address
    web_proxy_extender_route_patterns = data.privx_carrier.example.web_proxy_extender_route_patterns
  }
}
```

## Schema

### Required

- `name` (String) Carrier name

### Read-Only

- `id` (String) Carrier UUID
- `access_group_id` (String) Access Group ID
- `group_id` (String) Group ID for the carrier
- `enabled` (Boolean) Whether the carrier is enabled
- `registered` (Boolean) Whether the carrier is registered
- `permissions` (List of String) List of permissions for the carrier
- `routing_prefix` (String) Routing prefix for the carrier
- `extender_address` (List of String) List of extender addresses
- `subnets` (List of String) List of subnets for the carrier
- `web_proxy_address` (String) Web proxy address for the carrier
- `web_proxy_extender_route_patterns` (List of String) List of web proxy extender route patterns
- `secret` (String, Sensitive) Client Secret

**Note:** The carrier name must match an existing carrier in PrivX.