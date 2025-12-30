# privx_api_client (Resource)

Manages a PrivX API client. API clients are used for programmatic access to the PrivX API.

## Example Usage

```terraform
resource "privx_api_client" "example" {
  name = "my-api-client"
  
  roles = [
    {
      id = "role-id-1"
    },
    {
      id = "role-id-2"
      name = "role-name-2"
    }
  ]
}
```

## Schema

### Required

- `name` (String) Name of the API client
- `roles` (List of Object) Roles assigned to the API client

### Optional

None.

### Read-Only

- `id` (String) API Client ID
- `secret` (String, Sensitive) API Client secret
- `created` (String) Creation timestamp
- `updated` (String) Last update timestamp
- `updated_by` (String) User who last updated the API client
- `author` (String) User who created the API client
- `oauth_client_id` (String) OAuth client ID
- `oauth_client_secret` (String, Sensitive) OAuth client secret

### Nested Schema for `roles`

#### Required

- `id` (String) Role ID

#### Optional

- `name` (String) Role name

## Import

API clients can be imported using their ID:

```shell
terraform import privx_api_client.example api-client-id
```