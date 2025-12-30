# privx_carrier_config (Data Source)

Use this data source to get configuration for a PrivX carrier.

## Example Usage

```terraform
# Get carrier configuration
data "privx_carrier_config" "example" {
  id = "12345678-1234-1234-1234-123456789012"
}

# Output the configuration session ID and config content
output "carrier_session_id" {
  value = data.privx_carrier_config.example.session_id
}

output "carrier_config_content" {
  value     = data.privx_carrier_config.example.config
  sensitive = true
}
```

## Schema

### Required

- `id` (String) Carrier ID

### Read-Only

- `session_id` (String) Carrier configuration session ID
- `config` (String) Carrier configuration content (base64 encoded TOML file)

**Note:** 
- The carrier ID must be a valid UUID of an existing carrier.
- The config content is returned as a base64 encoded string containing the TOML configuration file.
- This field contains sensitive configuration data and should be handled securely.