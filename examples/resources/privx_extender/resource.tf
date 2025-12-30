resource "privx_extender" "example" {
  name           = "example-trusted-client"
  enabled        = true
  permissions    = ["privx-extender"]
  routing_prefix = "example"
  # Note: When extender_address and subnets are empty, 
  # the API will set them to ["0.0.0.0/0", "::/0"] by default
  extender_address = []
  subnets          = []
}

# Output the client credentials (be careful with sensitive data)
output "oauth_client_id" {
  value = privx_extender.example.oauth_client_id
}

output "oauth_client_secret" {
  value     = privx_extender.example.oauth_client_secret
  sensitive = true
}

output "client_secret" {
  value     = privx_extender.example.secret
  sensitive = true
}

# Output additional information
output "client_id" {
  value = privx_extender.example.id
}

output "registered" {
  value = privx_extender.example.registered
}

output "created" {
  value = privx_extender.example.created
}