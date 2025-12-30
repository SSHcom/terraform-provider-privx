# privx_api_client (Data Source)

Retrieves information about a PrivX API client.

## Example Usage

```terraform
# Get API client by ID
data "privx_api_client" "by_id" {
  id = "api-client-id"
}

# Get API client by name
data "privx_api_client" "by_name" {
  name = "my-api-client"
}
```

## Schema

### Optional

- `id` (String) API Client ID
- `name` (String) Name of the API client

### Read-Only

- `secret` (String, Sensitive) API Client secret
- `created` (String) Creation timestamp
- `updated` (String) Last update timestamp
- `updated_by` (String) User who last updated the API client
- `author` (String) User who created the API client
- `roles` (List of Object) Roles assigned to the API client
- `oauth_client_id` (String) OAuth client ID
- `oauth_client_secret` (String, Sensitive) OAuth client secret

### Nested Schema for `roles`

#### Read-Only

- `id` (String) Role ID
- `name` (String) Role name

## Notes

Either `id` or `name` must be specified to identify the API client. If both are provided, `id` takes precedence.