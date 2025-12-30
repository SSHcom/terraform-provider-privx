# PrivX Host Resource Examples

This directory contains examples for using the `privx_host` resource.

## Files

- `basic.tf` - Basic host configuration with minimal required fields
- `resource.tf` - Comprehensive examples showing various host configurations

## Usage Examples

1. **Basic SSH Host**: Creates a simple SSH host with one service and principal
2. **Host with Password Rotation**: Shows how to configure password rotation
3. **AWS EC2 Host**: Example of configuring a cloud-based host with SSH keys
4. **Windows RDP Host with Advanced Features**: Demonstrates comprehensive service options and command restrictions
5. **Database Server**: Shows database-specific configurations with service options

## Required Fields

- `common_name` - Display name for the host
- `addresses` - List of IP addresses or hostnames
- `source_id` - ID of the source that manages this host
- `access_group_id` - ID of the access group for this host

## Optional Configuration

### Basic Configuration
- `services` - Define connection services (SSH, RDP, HTTP, etc.)
- `principals` - Configure user accounts and their roles
- `ssh_host_public_keys` - Add SSH host public keys for verification
- `session_recording_options` - Configure session recording settings
- `tags` - Add metadata tags for organization

### Advanced Principal Configuration

#### Service Options
Configure protocol-specific options for each principal:

- **SSH Options**: Control shell access, file transfer, exec commands, tunnels, X11 forwarding
- **RDP Options**: Configure file transfer, audio, and clipboard access
- **Web Options**: Set file transfer, audio, and clipboard permissions
- **VNC Options**: Control file transfer and clipboard access
- **Database Options**: Set upload/download byte limits for database operations

#### Command Restrictions
Implement fine-grained command control:

- **Whitelists**: Define allowed commands using whitelist references
- **Role-based Access**: Assign different command sets to different roles
- **Audit Controls**: Configure logging for matched and unmatched commands
- **Shell Variants**: Support different shell types (bash, sh, powershell, etc.)
- **Custom Banners**: Display security warnings to users

## Configuration Examples

### Service Options Example
```hcl
service_options {
  ssh {
    shell         = true
    file_transfer = false  # Restrict file transfers
    exec          = true
    tunnels       = false  # Disable tunneling
    x11           = false
    other         = false
  }
  
  rdp {
    file_transfer = true
    audio         = true
    clipboard     = true
  }
  
  db {
    max_bytes_upload   = 1048576   # 1MB limit
    max_bytes_download = 10485760  # 10MB limit
  }
}
```

### Command Restrictions Example
```hcl
command_restrictions {
  enabled         = true
  rshell_variant  = "bash"
  allow_no_match  = false
  audit_match     = true
  audit_no_match  = true
  banner          = "Authorized access only. All activities monitored."
  
  default_whitelist {
    id   = "basic-commands-id"
    name = "basic_linux_commands"
  }
  
  whitelists {
    whitelist {
      id   = "admin-commands-id"
      name = "admin_commands"
    }
    
    roles {
      id   = "admin-role-id"
      name = "System Admin"
    }
  }
}
```

## Security Best Practices

1. **Principle of Least Privilege**: Only enable necessary service options
2. **Command Restrictions**: Use whitelists to control allowed commands
3. **Audit Logging**: Enable audit options for security monitoring
4. **Service Accounts**: Use more restrictive settings for automated accounts
5. **Role-based Access**: Assign different permissions based on user roles

## Limitations

- **Password Rotation**: Password rotation functionality is not currently implemented in this host resource. While the `password_rotation_enabled` field and related configuration options are available in the schema, the actual password rotation logic is not functional.

## Notes

- At least one service should be configured for the host to be usable
- Principals define which users can access the host and with what permissions
- SSH host public keys are recommended for secure connections
- Service options provide granular control over protocol-specific features
- Command restrictions enhance security by limiting executable commands
- Different principals can have different service options and command restrictions