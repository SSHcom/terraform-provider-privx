## 1.42.0 (Unreleased)

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
