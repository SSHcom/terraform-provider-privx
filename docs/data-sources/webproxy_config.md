# privx_webproxy_config (Data Source)

Use this data source to get configuration for a PrivX webproxy.

## Example Usage

```terraform
# Get webproxy configuration
data "privx_webproxy_config" "example" {
  id = "12345678-1234-1234-1234-123456789012"
}

# Output the configuration session ID and config content
output "webproxy_session_id" {
  value = data.privx_webproxy_config.example.session_id
}

output "webproxy_config_content" {
  value     = data.privx_webproxy_config.example.config
  sensitive = true
}
```

## Schema

### Required

- `id` (String) WebProxy ID

### Read-Only

- `session_id` (String) WebProxy configuration session ID
- `config` (String) WebProxy configuration content (base64 encoded TOML file)

**Note:** 
- The webproxy ID must be a valid UUID of an existing webproxy.
- The config content is returned as a base64 encoded string containing the TOML configuration file.
- This field contains sensitive configuration data and should be handled securely.