# PrivX Terraform Provider

The PrivX Terraform Provider allows you to manage PrivX configuration using Terraform,
including hosts, roles, workflows, access groups, and password rotation settings.

This provider is intended for teams who want Infrastructure as Code (IaC) control over PrivX access management, approvals, and automation.


### Installation

```
terraform {
  required_providers {
    privx = {
      source  = "hashicorp/privx"
      version = "~> 1.42.0"
    }
  }
}
```
Replace the version constraint with the latest published version.


## Provider Configuration
### Authentication

The PrivX provider authenticates using OAuth2 client credentials.
Configuration is typically supplied via environment variables.

Required Environment Variables

| Variable | Description |
|--------|-------------|
| `PRIVX_API_BASE_URL` | Base URL of the PrivX API (for example: `https://privx.example.com`) |
| `PRIVX_API_OAUTH_CLIENT_ID` | OAuth 2.0 client ID used to authenticate with PrivX |
| `PRIVX_API_OAUTH_CLIENT_SECRET` | OAuth 2.0 client secret used to authenticate with PrivX |
| `PRIVX_API_CLIENT_ID` | PrivX API client identifier (used for API access after OAuth authentication) |
| `PRIVX_API_CLIENT_SECRET` | PrivX API client secret associated with the API client |


### Provider Block
No arguments are required in the provider block when environment variables are used.

```
provider "privx" {}

resource "privx_access_group" "example" {
    name    = "example-group"
    comment = "Managed by Terraform"
}

resource "privx_role" "example" {
    name            = "example-role"
    access_group_id = privx_access_group.example.id
    
    permissions  = ["users-view"]
    permit_agent = false
}
```

### Resources and Data Sources

See the Documentation sidebar for full resource and data source references.

### Importing Existing Resources

Most PrivX resources support Terraform import.

### Example: Import a Host
```
terraform import privx_host.example <HOST_ID>
```

After importing, run:

```
terraform plan
```
to review any detected differences.

### Troubleshooting
#### Duplicate Name Errors

PrivX enforces uniqueness on several fields, including:
- role names
- workflow names
- host service addresses

Symptoms:
```
VALUE_DUPLICATE, message: Duplicate name
```

Resolution:
Ensure resource names are unique
For tests or automation, include a random or environment-specific suffix


#### Service Address Not Unique (Hosts)

Symptoms:
```VALUE_DUPLICATE, message: Service Address not unique```

Cause:
A host with the same service address already exists in PrivX.

Resolution:

Use unique IPs or hostnames per host

Avoid reusing fixed IPs across test runs

#### Permission / Authorization Errors
```
FORBIDDEN
```

Resolution:

- Ensure the OAuth client has sufficient PrivX permissions

- Verify access to the relevant domains (roles, workflows, hosts)


### Compatibility

| Component           | Supported           |
| ------------------- |---------------------|
| PrivX Server        | PrivX v42 and newer |
| Terraform           | Terraform 1.3+      |

