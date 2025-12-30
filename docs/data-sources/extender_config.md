# privx_extender_config (Data Source)

Use this data source to get configuration sessions for a PrivX extender.

## Example Usage

```terraform
# Get extender configuration
data "privx_extender_config" "example" {
  id = "12345678-1234-1234-1234-123456789012"
}

# Output the configuration session ID and config content
output "extender_session_id" {
  value = data.privx_extender_config.example.session_id
}

output "extender_config_content" {
  value     = data.privx_extender_config.example.config
  sensitive = true
}
```

## Schema

### Required

- `id` (String) Extender ID

### Read-Only

- `session_id` (String) Extender configuration session ID
- `config` (String) Extender configuration content (base64 encoded TOML file)

**Note:** 
- The extender ID must be a valid UUID of an existing extender.
- The config content is returned as a base64 encoded string containing the TOML configuration file.
- This field contains sensitive configuration data and should be handled securely.