resource "privx_carrier" "example" {
  name            = "mycarrier"
  access_group_id = "33304b0b-e62e-44ef-b78f-26bb5e0edc88"
  subnets = [
    "0.0.0.0/0"
  ]
  enabled = true
  extender_address = [
    "0.0.0.0/0",
  ]
  web_proxy_address                 = "0.0.0.0"
  routing_prefix                    = "carrier"
  web_proxy_extender_route_patterns = ["route_pattern"]
}