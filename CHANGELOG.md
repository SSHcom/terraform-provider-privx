## 1.45.1 (Released)

Tested against PrivX 45.

### Bug Fixes
- `privx_host` data source now populates `id` and all other attributes (addresses, tags, services, principals, SSH host public keys, session recording options, timestamps). In 1.45.0 the data source was registered but returned empty state, so `id` was never set. Principal `passphrase` is always null because the API returns it masked.

### Tests
- Added `TestHostDataSource_populateModel`, a unit test that maps an API host through the real data source schema without needing a live PrivX.

## 1.45.0 (Released)

Tested against PrivX 45.

### Security
- Stopped logging API secrets. `api_client_secret`, `api_oauth_client_secret` and `api_bearer_token` were added to the Terraform log context in the provider, and also printed by an unconditional `log.Println` in the client. Credentials could appear in logs at `TF_LOG=INFO` or higher, or without `TF_LOG` at all in the client case. If you ran with debug logging enabled on earlier versions, consider rotating those credentials.

### Features
- Added data source `privx_ca_certificate`, which exposes the PrivX authorizer CA certificates and public keys. It can be used to detect CA rotation.
- Registered the `privx_host` data source (it was implemented but not registered). Note: it returned empty state in this release; fixed in 1.45.1.

### Bug Fixes
- Provider error messages now name the correct environment variables, `PRIVX_API_OAUTH_CLIENT_ID` and `PRIVX_API_OAUTH_CLIENT_SECRET`. They previously said `PRIVX_API_OAUTH_ID` and `PRIVX_API_OAUTH_SECRET`.
- Corrected the `api_client_id` and `api_client_secret` provider attribute descriptions.

### Breaking / Action Required
- The provider source address is now `SSHcom/privx` (`registry.terraform.io/SSHcom/privx`). The previous `hashicorp/privx` address was incorrect. Update `required_providers`:
  ```hcl
  terraform {
    required_providers {
      privx = {
        source  = "SSHcom/privx"
        version = "~> 1.45.0"
      }
    }
  }
  ```

### Documentation
- Fixed `examples/resources/privx_host/basic.tf`, which used removed block syntax. It now uses list attributes.
- Corrected the host example README, which claimed password rotation was not implemented.
- Removed the `privx_license` examples, which had no backing implementation.
- Standardized the provider source address across the README, registry README and docs.
- Regenerated provider documentation.

### Build & Development
- CI lint now uses the repository `.golangci.yaml` instead of `--no-config`; the config was migrated to the golangci-lint v2 schema.
- Pinned golangci-lint to the prebuilt `v2.5.0` binary (`goinstall @latest` failed against Go 1.25).
- Fixed the `task doc` flags (`--provider-name`, `--rendered-website-dir`).
- Removed the unsupported `trust_anchor` option from `scripts/generate-provider-tf.sh`.
- Removed the reference to a non-existent acceptance-tests workflow from `DEVELOPMENT.md`. Acceptance tests run from a separate private repository.

### Tests
- Added acceptance tests for the `privx_host` and `privx_ca_certificate` data sources.

---

## 1.44.0 (Released)

Tested against PrivX 44.

### Features
- Added `access-roles-manage` permission to role resource, introduced in PrivX 44

## 1.43.0 (Released)

Tested against PrivX 43.

### Improvements
- Role resource: gracefully handle principal key creation failures with a warning instead of aborting the entire resource creation
- Workflow resource: automatically remove from Terraform state when deleted out-of-band (not-found handling)
- Whitelist resource: use shared `IsPrivxNotFound` utility for consistent not-found detection

### Bug Fixes
- Fixed potential crash in role resource when `CreatePrincipalKey` returns an error before the key ID is available

### Tests
- Added `TestAccRoleResource_basicCreateUpdateDelete` acceptance test covering full create → update → delete lifecycle (comment, permissions, permit_agent)
- Acceptance test list now references individual role test cases

### Build & Development
- Updated Go module dependencies (golang.org/x/crypto v0.46.0, grpc v1.79.3, protobuf v1.36.10, opentelemetry v1.39.0, and others)
- Fixed `mermaid-cli` Nix package reference in shell.nix

---

## 1.42.0 (Released)

### Features
- Added new data sources: `privx_api_proxy_config`, `privx_api_target`, `privx_password_policy`, `privx_script_template`
- Added new resources: `privx_api_proxy_credential`, `privx_api_target`, `privx_local_user`, `privx_local_user_password`, `privx_network_target`
- Enhanced workflow resource with additional configuration options
- Improved host resource with expanded attribute support
- Added comprehensive acceptance test coverage for all resources

### Improvements
- Refactored client connection handling for better reliability
- Enhanced provider configuration with improved error handling
- Simplified example configurations across all resources and data sources
- Updated Go module dependencies to latest versions
- Improved resource import functionality
- Enhanced data source filtering and querying capabilities

### Documentation
- Completely restructured README.md with comprehensive provider documentation
- Added clear sections for resources and data sources listing
- Included detailed "How to Use the Provider" section with local build instructions
- Added disclaimer section emphasizing testing in non-production environments
- Enhanced "How to Contribute" section with development setup and guidelines
- Added note about upcoming Terraform Registry publication
- Improved provider configuration examples with environment variables
- Updated all resource and data source documentation with current schemas
- Added Overview and Technologies sections for better context
- Generated comprehensive documentation using terraform-plugin-docs

### Build & Development
- Added Taskfile.yml for improved build automation
- Enhanced shell.nix with additional development dependencies
- Updated CI/CD workflows for better testing coverage
- Added acceptance test helpers for consistent testing
- Improved development environment setup

### Breaking Changes
- Updated provider schema to match latest PrivX API specifications
- Simplified resource configurations (removed deprecated attributes)
- Updated minimum Go version requirement
- Changed some attribute names for consistency across resources

### Internal
- Major refactoring of provider code structure
- Improved error handling and logging throughout
- Enhanced test coverage with comprehensive acceptance tests
- Updated internal utilities and helper functions
- Streamlined resource lifecycle management
