# privx_whitelist (Data Source)

Whitelist data source for retrieving command restriction whitelists in PrivX.

## Example Usage

```terraform
# Lookup by name
data "privx_whitelist" "example" {
  name = "example-whitelist"
}

# Lookup by ID
data "privx_whitelist" "by_id" {
  id = "12345678-1234-1234-1234-123456789012"
}

output "whitelist_name" {
  value = data.privx_whitelist.example.name
}

output "whitelist_patterns" {
  value = data.privx_whitelist.example.whitelist_patterns
}
```

## Schema

### Optional

- `id` (String) Whitelist ID (either id or name must be specified)
- `name` (String) Whitelist name (either id or name must be specified)

### Read-Only

- `comment` (String) Whitelist comment/description
- `type` (String) Whitelist type
- `whitelist_patterns` (Set of String) List of command patterns allowed by this whitelist