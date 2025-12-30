# privx_carrier (Resource)

Manages a PrivX carrier. Carriers are trusted clients that provide web proxy functionality for PrivX.

## Example Usage

```terraform
resource "privx_carrier" "example" {
  name                              = "example-carrier"
  enabled                           = true
  permissions                       = ["privx-carrier"]
  routing_prefix                    = "example"
  web_proxy_address                 = "10.31.34.194"
  web_proxy_extender_route_patterns = ["rp1"]
  extender_address                  = []
  subnets                           = []
}
```

## Schema

### Required

- `name` (String) The name of the carrier

### Optional

- `access_group_id` (String) The access group ID for the carrier
- `enabled` (Boolean) Whether the carrier is enabled (default: true)
- `permissions` (List of String) List of permissions for the carrier
- `routing_prefix` (String) Routing prefix for the carrier
- `extender_address` (List of String) List of extender addresses
- `subnets` (List of String) List of subnets for the carrier
- `web_proxy_address` (String) Web proxy address for the carrier
- `web_proxy_extender_route_patterns` (List of String) List of web proxy extender route patterns

### Read-Only

- `id` (String) The ID of the carrier
- `group_id` (String) The group ID for the carrier
- `secret` (String, Sensitive) The client secret
- `registered` (Boolean) Whether the carrier is registered

## Notes

- The `type` is automatically set to `"CARRIER"` as this resource is specifically for creating carrier clients.
- When `extender_address` and `subnets` are provided as empty lists, the PrivX API will automatically set them to `["0.0.0.0/0", "::/0"]` (allowing all IPv4 and IPv6 traffic).
- The `access_group_id` can be left empty and will be handled by the API.

## Import

Import is supported using the following syntax:

```shell
terraform import privx_carrier.example <carrier_id>
```