# Terraform Provider for PrivX

This repository is the official Terraform provider for PrivX, enabling Infrastructure as Code management of PrivX resources through declarative Terraform configurations.

## Overview

The current version is based on:

- [privx-sdk-go](https://github.com/SSHcom/privx-sdk-go) v2.42.0
- [Terraform Plugin Framework](https://github.com/hashicorp/terraform-plugin-framework) v1.15.1

## Technologies

- [Go](https://go.dev/)
- [Terraform](https://developer.hashicorp.com/terraform)
- Terraform Plugin Framework
- Terraform Plugin SDK

## Resources and Data Sources

### Resources

- `privx_access_group` - Manage PrivX access groups
- `privx_api_client` - Manage API clients
- `privx_api_proxy_credential` - Manage API proxy credentials
- `privx_api_target` - Manage API targets
- `privx_carrier` - Manage PrivX carriers deployment
- `privx_extender` - Manage PrivX extenders deployment
- `privx_host` - Manage PrivX hosts
- `privx_local_user` - Manage local user accounts
- `privx_local_user_password` - Reset local user passwords
- `privx_network_target` - Manage network targets
- `privx_role` - Manage PrivX roles and permissions
- `privx_secret` - Manage PrivX secrets
- `privx_source` - Manage user sources and identity providers
- `privx_whitelist` - Manage PrivX command whitelists
- `privx_workflow` - Manage PrivX workflows

### Data Sources

- `privx_access_group` - Read access group information
- `privx_api_client` - Read API client information
- `privx_api_proxy_config` - Read API proxy configuration
- `privx_api_target` - Read API targets
- `privx_carrier` - Read carrier information
- `privx_carrier_config` - Read carrier configuration
- `privx_extender` - Read extender information
- `privx_extender_config` - Read extender configuration
- `privx_host` - Read host information
- `privx_password_policy` - Read password policy information
- `privx_role` - Read role information
- `privx_script_template` - Read script template information
- `privx_secret` - Read secret information
- `privx_source` - Read user source information
- `privx_webproxy` - Read web proxy information
- `privx_webproxy_config` - Read web proxy configuration
- `privx_whitelist` - Read command whitelist information
- `privx_workflow` - Read workflow information

## How to Use the Provider

**Note:** This provider will soon be published to the Terraform Registry. For now, you need to clone the repository and build it locally.

### Local Build and Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/SSHcom/terraform-provider-privx.git
   cd terraform-provider-privx
   ```

2. Build the provider:
   ```bash
   go build -o terraform-provider-privx_v1.42.0
   ```

3. Create the local provider directory:
   ```bash
   mkdir -p <terraform_code_dir>/.terraform/providers/local.terraform.com/local/privx/1.42.0/darwin_arm64/
   ```
   
   Note: Replace `darwin_amd64` with your platform (e.g., `linux_amd64`, `windows_amd64`)

4. Copy the built provider:
   ```bash
   cp terraform-provider-privx_v1.42.0 <terraform_code_dir>/.terraform/providers/local.terraform.com/local/privx/1.42.0/darwin_arm64/
   ```

### Basic Configuration

```hcl
terraform {
  required_providers {
    privx = {
      source  = "local.terraform.com/local/privx"
      #source = "sshcom/privx"
      version = "1.42.0"
    }
  }
}

provider "privx" {
  api_base_url         = "https://your-privx-instance.com"
  api_client_id        = "your-client-id"
  api_client_secret    = "your-client-secret"
  api_oauth_client_id  = "your-oauth-client-id"
  api_oauth_client_secret = "your-oauth-client-secret"
}
```

### Environment Variables

You can also configure the provider using environment variables:

- `PRIVX_API_BASE_URL` - PrivX API base URL
- `PRIVX_API_CLIENT_ID` - API client ID
- `PRIVX_API_CLIENT_SECRET` - API client secret
- `PRIVX_API_OAUTH_CLIENT_ID` - OAuth client ID
- `PRIVX_API_OAUTH_CLIENT_SECRET` - OAuth client secret

### Examples

Example configurations for all resources and data sources can be found in the [examples](examples/) directory:

- **Resources**: [examples/resources/](examples/resources/)
- **Data Sources**: [examples/data-sources/](examples/data-sources/)
- **Provider Configuration**: [examples/provider/](examples/provider/)

## Disclaimer

**⚠️ Important: Testing and Environment Usage**

- **Always test in non-production environments first** before using provider in production PrivX
- This provider makes direct API calls to PrivX and can modify your PrivX configuration
- Ensure you have proper backups and rollback procedures in place
- Test all configurations thoroughly in development/staging environments
- Review all Terraform plans carefully before applying changes
- Use version control for your Terraform configurations
- Consider using Terraform workspaces to separate environments

## How to Contribute

We welcome contributions to improve the PrivX Terraform provider!

### Getting Started

1. Fork the repository
2. Clone your fork: `git clone https://github.com/your-username/terraform-provider-privx.git`
3. Create a feature branch: `git checkout -b feature/your-feature-name`

### Development Setup

Refer to [Development instructions](docs/manual/DEVELOPMENT.md)

### Submitting Changes

1. Ensure all tests pass
2. Add tests for new functionality
3. Update documentation as needed
4. Commit your changes with clear commit messages
5. Push to your fork and submit a pull request

### Development Resources

- [Installation requirements](docs/manual/REQUIREMENTS.md)
- [High level design](docs/manual/DESIGN.md)
- [Development instructions](docs/manual/DEVELOPMENT.md)

### Reporting Issues

Please report bugs and feature requests through [GitHub Issues](https://github.com/SSHcom/terraform-provider-privx/issues).
