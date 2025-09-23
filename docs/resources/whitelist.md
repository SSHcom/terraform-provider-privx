# privx_whitelist (Resource)

Whitelist resource for command restrictions in PrivX.

## Example Usage

```terraform
resource "privx_whitelist" "example" {
  name    = "example-whitelist"
  comment = "Example whitelist for command restrictions"
  type    = "command"
  whitelist_patterns = [
    "ls -la",
    "cat /etc/passwd",
    "grep -r 'pattern' /var/log/"
  ]
}
```

## Schema

### Required

- `name` (String) Whitelist name

### Optional

- `comment` (String) Whitelist comment/description
- `type` (String) Whitelist type
- `whitelist_patterns` (Set of String) List of command patterns allowed by this whitelist

### Read-Only

- `id` (String) Whitelist ID

## Import

Import is supported using the following syntax:

```shell
terraform import privx_whitelist.example <whitelist_id>
```