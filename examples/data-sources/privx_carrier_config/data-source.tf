# Get carrier configuration by carrier ID
data "privx_carrier_config" "example" {
  id = "12345678-1234-1234-1234-123456789012"
}

# You can also use the ID from a carrier resource
data "privx_carrier_config" "from_resource" {
  id = privx_carrier.example.id
}

# Output the configuration information
output "carrier_config_info" {
  value = {
    carrier_id = data.privx_carrier_config.example.id
    session_id = data.privx_carrier_config.example.session_id
  }
}

# Output the configuration content (sensitive)
output "carrier_config_content" {
  value     = data.privx_carrier_config.example.config
  sensitive = true
}

# Example: Save config to a local TOML file (base64 decode)
resource "local_file" "carrier_config" {
  content_base64 = data.privx_carrier_config.example.config
  filename       = "carrier-config.toml"
}