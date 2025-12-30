# privx_secret (Data Source)

Retrieves information about a PrivX secret. This data source allows you to access secret metadata and data (if you have appropriate permissions).

## Example Usage

```terraform
# Get secret by name
data "privx_secret" "database_config" {
  name = "db-prod-password"
}

# Use secret data in other resources
resource "kubernetes_secret" "db_credentials" {
  metadata {
    name = "database-credentials"
  }

  data = {
    username = data.privx_secret.database_config.data["username"]
    password = data.privx_secret.database_config.data["password"]
    host     = data.privx_secret.database_config.data["host"]
    port     = data.privx_secret.database_config.data["port"]
  }
}

# Access secret metadata
output "secret_info" {
  value = {
    name       = data.privx_secret.database_config.name
    author     = data.privx_secret.database_config.author
    created    = data.privx_secret.database_config.created
    updated    = data.privx_secret.database_config.updated
    path       = data.privx_secret.database_config.path
  }
}

# Check secret permissions
output "secret_permissions" {
  value = {
    read_roles  = data.privx_secret.database_config.read_roles
    write_roles = data.privx_secret.database_config.write_roles
  }
}
```

## Schema

### Required

- `name` (String) Secret name

### Read-Only

- `author` (String) Author of the secret
- `created` (String) Creation timestamp
- `data` (Map of String, Sensitive) Secret data as key-value pairs
- `owner_id` (String) Owner ID of the secret
- `path` (String) Secret path in the vault
- `read_roles` (Block List) List of roles that can read this secret (see [below for nested schema](#nestedblock--read_roles))
- `updated` (String) Last update timestamp
- `updated_by` (String) ID of user who last updated the secret
- `write_roles` (Block List) List of roles that can write to this secret (see [below for nested schema](#nestedblock--write_roles))

<a id="nestedblock--read_roles"></a>
### Nested Schema for `read_roles`

#### Read-Only

- `id` (String) Role ID
- `name` (String) Role name

<a id="nestedblock--write_roles"></a>
### Nested Schema for `write_roles`

#### Read-Only

- `id` (String) Role ID
- `name` (String) Role name

## Security Considerations

1. **Access Control**: You can only read secrets that your PrivX user/role has permission to access.
2. **Sensitive Data**: The `data` field contains sensitive information and should be handled carefully.
3. **Audit Logging**: All secret access is logged by PrivX for security auditing.
4. **State File Security**: Ensure your Terraform state files are properly secured as they will contain the secret data.

## Error Handling

- If the secret doesn't exist, Terraform will return an error
- If you don't have read permissions for the secret, PrivX will return an access denied error
- Network connectivity issues will result in timeout errors

## Notes

- The data source will retrieve the current state of the secret from PrivX
- Secret data is marked as sensitive and will not be displayed in Terraform logs
- Role information includes both ID and name for easier identification
- The `path` field shows the full vault path where the secret is stored