# Get webproxy configuration by webproxy ID
data "privx_webproxy_config" "example" {
  id = "12345678-1234-1234-1234-123456789012"
}

# You can also use the ID from a carrier resource (if it has webproxy functionality)
data "privx_webproxy_config" "from_carrier" {
  id = privx_carrier.example.id
}

# Output the configuration information
output "webproxy_config_info" {
  value = {
    webproxy_id = data.privx_webproxy_config.example.id
    session_id  = data.privx_webproxy_config.example.session_id
  }
}

# Output the configuration content (sensitive)
output "webproxy_config_content" {
  value     = data.privx_webproxy_config.example.config
  sensitive = true
}

# Example: Save config to a local TOML file (base64 decode)
resource "local_file" "webproxy_config" {
  content_base64 = data.privx_webproxy_config.example.config
  filename       = "webproxy-config.toml"
}