# privx_secret (Resource)

Manages a PrivX secret resource. Secrets are used to store sensitive information like passwords, API keys, and certificates that can be accessed by authorized roles.

## Example Usage

```terraform
# Create a basic secret
resource "privx_secret" "database_password" {
  name = "db-prod-password"
  
  data = {
    _privx_schema = "credentials"
    username = "dbadmin"
    password = "super-secret-password"
    host     = "db.example.com"
    port     = "5432"
  }

  read_roles = [{
    id   = "db-admin-role-id"
    name = "Database Admin"
  }]

  write_roles = [{
    id   = "db-admin-role-id"
    name = "Database Admin"
  }]
}

# Create an API key secret
resource "privx_secret" "api_keys" {
  name = "external-api-keys"
  
  data = {
    github_token    = "ghp_xxxxxxxxxxxxxxxxxxxx"
    slack_webhook   = "https://hooks.slack.com/services/xxx/xxx/xxx"
    aws_access_key  = "AKIAIOSFODNN7EXAMPLE"
    aws_secret_key  = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
  }

  read_roles = [{
    id   = "devops-role-id"
    name = "DevOps Team"
  },
  {
    id   = "developer-role-id"
    name = "Developers"
  }]

  write_roles {
    id   = "devops-role-id"
    name = "DevOps Team"
  }
}

# Create a certificate secret
resource "privx_secret" "ssl_certificate" {
  name     = "web-server-ssl-cert"
  owner_id = "certificate-manager-user-id"
  
  data = {
    certificate = file("${path.module}/certs/server.crt")
    private_key = file("${path.module}/certs/server.key")
    ca_bundle   = file("${path.module}/certs/ca-bundle.crt")
  }

  read_roles = [{
    id   = "web-admin-role-id"
    name = "Web Administrators"
  }]

  write_roles = [{
    id   = "cert-manager-role-id"
    name = "Certificate Managers"
  }]
}

# Example with API client permissions for Terraform
resource "privx_secret" "terraform_managed" {
  name = "terraform-managed-secret"
  
  data = {
    api_key = "secret-api-key-value"
    token   = "secret-token-value"
  }

  # Include API client role for both read and write access
  read_roles = [{
    id   = "terraform-api-client-role-id"  # API client role
    name = "Terraform API Client"
  },
  {
    id   = "application-role-id"
    name = "Application Role"
  }]

  write_roles = [{
    id   = "terraform-api-client-role-id"  # API client role (required for updates)
    name = "Terraform API Client"
  },
  {
    id   = "admin-role-id"
    name = "Administrators"
  }]
}
```

## Schema

### Required

- `name` (String) Secret name (used as identifier). Cannot be changed after creation.

### Optional

- `data` (Map of String, Sensitive) Secret data as key-value pairs. All values are stored securely.
- `owner_id` (String) Owner ID of the secret. Defaults to `""`.
- `read_roles` (Block List) List of roles that can read this secret (see [below for nested schema](#nestedblock--read_roles))
- `write_roles` (Block List) List of roles that can write to this secret (see [below for nested schema](#nestedblock--write_roles))

### Read-Only

- `author` (String) Author of the secret
- `created` (String) Creation timestamp
- `path` (String) Secret path in the vault
- `updated` (String) Last update timestamp
- `updated_by` (String) ID of user who last updated the secret

<a id="nestedblock--read_roles"></a>
### Nested Schema for `read_roles`

#### Required

- `id` (String) Role ID

#### Optional

- `name` (String) Role name

<a id="nestedblock--write_roles"></a>
### Nested Schema for `write_roles`

#### Required

- `id` (String) Role ID

#### Optional

- `name` (String) Role name

## Security Considerations

1. **Sensitive Data**: All data in the `data` map is marked as sensitive and will not be displayed in Terraform logs or output.
2. **Role-based Access**: Use `read_roles` and `write_roles` to control who can access the secret.
3. **Least Privilege**: Only grant the minimum necessary permissions to roles.
4. **Audit Trail**: All secret access and modifications are logged by PrivX.
5. **API Client Permissions**: The PrivX API client used by Terraform must ideally have read access to secrets it creates so it can read them back after create/update operations.
6. **Write Permissions Required**: The API client must have write permissions (be included in `write_roles`) to update secret data. Without write permissions, updates will fail with "PERMISSION_DENIED" error.

## Import

Secrets can be imported using their name:

```shell
terraform import privx_secret.example my-secret-name
```

**Note**: When importing secrets, the `data` field will be populated with the current secret data from PrivX. If your Terraform configuration doesn't specify the `data` field, the existing data will be preserved during subsequent updates. If you want to manage the secret data through Terraform, make sure to include the `data` field in your configuration after import.

## Notes

- Secret names must be unique within the PrivX vault
- The `data` field is sensitive and its contents will not be displayed in Terraform output
- Changes to the secret data will trigger an update operation
- Deleting a secret resource will permanently remove it from the PrivX vault
- Role permissions are enforced by PrivX - ensure the API client has appropriate permissions
- The API client must be included in `write_roles` to update secret data, otherwise updates will fail with permission errors