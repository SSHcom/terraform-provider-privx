resource "privx_carrier" "example" {
  name                              = "example-carrier"
  enabled                           = true
  permissions                       = ["privx-carrier"]
  routing_prefix                    = "example"
  web_proxy_address                 = "10.31.34.194"
  web_proxy_extender_route_patterns = ["rp1"]
  # Note: When extender_address and subnets are empty, 
  # the API will set them to ["0.0.0.0/0", "::/0"] by default
  extender_address = []
  subnets          = []
}

# Output the carrier information
output "carrier_id" {
  value = privx_carrier.example.id
}

output "carrier_secret" {
  value     = privx_carrier.example.secret
  sensitive = true
}

output "carrier_registered" {
  value = privx_carrier.example.registered
}

output "carrier_group_id" {
  value = privx_carrier.example.group_id
}