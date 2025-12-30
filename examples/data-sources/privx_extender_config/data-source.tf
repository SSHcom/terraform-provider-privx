# Get extender configuration by extender ID
data "privx_extender_config" "example" {
  id = "12345678-1234-1234-1234-123456789012"
}

# You can also use the ID from an extender resource
data "privx_extender_config" "from_resource" {
  id = privx_extender.example.id
}

# Output the configuration information
output "extender_config_info" {
  value = {
    extender_id = data.privx_extender_config.example.id
    session_id  = data.privx_extender_config.example.session_id
  }
}

# Output the configuration content (sensitive)
output "extender_config_content" {
  value     = data.privx_extender_config.example.config
  sensitive = true
}

# Example: Save config to a local TOML file (base64 decode)
resource "local_file" "extender_config" {
  content_base64 = data.privx_extender_config.example.config
  filename       = "extender-config.toml"
}