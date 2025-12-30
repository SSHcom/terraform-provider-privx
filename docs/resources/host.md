# privx_host (Resource)

Manages a PrivX host resource. Hosts represent target systems that users can connect to through PrivX.

## Example Usage

```terraform
# Create a basic SSH host
resource "privx_host" "example" {
  common_name     = "example-server"
  addresses       = ["192.168.1.100", "server.example.com"]
  source_id       = "a0ad72dc-b4aa-4a53-b7e0-14902b8b18bd"
  access_group_id = "bfe74d8b-feda-46f1-7ac3-b37ef5b15e3b"

  services {
    service = "SSH"
    address = "192.168.1.100"
    port    = 22
    source  = "UI"
  }

  principals {
    principal = "ubuntu"
    source    = "UI"
    
    roles {
      id   = "e2dcbeb4-6b07-50ad-788a-5af830da74ca"
      name = "Linux-admin"
    }
  }

  tags = ["production", "web-server"]
}

# Create a host with password rotation enabled
resource "privx_host" "with_rotation" {
  common_name                = "db-server"
  addresses                  = ["10.0.1.50"]
  source_id                  = "a0ad72dc-b4aa-4a53-b7e0-14902b8b18bd"
  access_group_id            = "bfe74d8b-feda-46f1-7ac3-b37ef5b15e3b"
  password_rotation_enabled  = true

  services {
    service                    = "SSH"
    address                    = "10.0.1.50"
    port                       = 22
    use_for_password_rotation  = true
    source                     = "UI"
  }

  principals {
    principal                  = "dbadmin"
    use_for_password_rotation  = true
    source                     = "UI"
    
    roles {
      id   = "db-admin-role-id"
      name = "Database Admin"
    }
  }

  session_recording_options {
    disable_clipboard_recording     = false
    disable_file_transfer_recording = false
  }
}

# Create a Windows RDP host with service options and command restrictions
resource "privx_host" "windows_rdp" {
  common_name     = "windows-server"
  addresses       = ["192.168.1.200"]
  source_id       = "a0ad72dc-b4aa-4a53-b7e0-14902b8b18bd"
  access_group_id = "bfe74d8b-feda-46f1-7ac3-b37ef5b15e3b"

  services {
    service = "RDP"
    address = "192.168.1.200"
    port    = 3389
    source  = "UI"
  }

  principals {
    principal = "Administrator"
    source    = "UI"
    
    roles {
      id   = "windows-admin-role-id"
      name = "Windows Admin"
    }

    # Configure service options for different protocols
    service_options {
      ssh {
        shell         = true
        file_transfer = true
        exec          = true
        tunnels       = false
        x11           = false
        other         = true
      }
      
      rdp {
        file_transfer = true
        audio         = true
        clipboard     = true
      }
      
      web {
        file_transfer = false
        audio         = true
        clipboard     = true
      }
      
      vnc {
        file_transfer = false
        clipboard     = true
      }
      
      db {
        max_bytes_upload   = 1048576  # 1MB
        max_bytes_download = 10485760 # 10MB
      }
    }

    # Configure command restrictions
    command_restrictions {
      enabled         = true
      rshell_variant  = "powershell"
      allow_no_match  = false
      audit_match     = true
      audit_no_match  = true
      banner          = "Authorized access only. All activities are monitored."
      
      default_whitelist {
        id   = "11c24354-e280-4fb2-791b-9f5c80dd2b01"
        name = "windows_basic_commands"
      }
      
      whitelists {
        whitelist {
          id   = "7ea51435-59b9-4e6d-422d-4acec3173260"
          name = "admin_commands"
        }
        
        roles {
          id   = "65127d0a-b2df-403c-be19-e2960d10de4d"
          name = "privx-admin"
        }
      }
    }
  }

  tags = ["windows", "production"]
}
```

## Schema

### Required

- `addresses` (List of String) List of host addresses (IP addresses or hostnames)
- `access_group_id` (String) Access Group ID for the host
- `common_name` (String) Host common name (display name)
- `source_id` (String) Source ID for the host

### Optional

- `audit_enabled` (Boolean) Whether audit is enabled for the host. Defaults to `false`.
- `cloud_provider` (String) Cloud provider for the host. Defaults to `""`.
- `cloud_provider_region` (String) Cloud provider region for the host. Defaults to `""`.
- `comment` (String) Comment for the host. Defaults to `""`.
- `contact_address` (String) Contact address for the host. Defaults to `""`.
- `deployable` (Boolean) Whether the host is deployable. Defaults to `false`.
- `disabled` (String) Whether the host is disabled. Defaults to `"FALSE"`.
- `distinguished_name` (String) Distinguished name for the host. Defaults to `""`.
- `external_id` (String) External ID for the host. Defaults to `""`.
- `host_classification` (String) Host classification. Defaults to `""`.
- `host_type` (String) Host type. Defaults to `""`.
- `instance_id` (String) Instance ID for the host. Defaults to `""`.
- `organization` (String) Organization for the host. Defaults to `""`.
- `organizational_unit` (String) Organizational unit for the host. Defaults to `""`.
- `password_rotation_enabled` (Boolean) Whether password rotation is enabled. Defaults to `false`.
- `principals` (Block List) List of principals for the host (see [below for nested schema](#nestedblock--principals))
- `services` (Block List) List of services for the host (see [below for nested schema](#nestedblock--services))
- `session_recording_options` (Block) Session recording options (see [below for nested schema](#nestedblock--session_recording_options))
- `ssh_host_public_keys` (Block List) List of SSH host public keys (see [below for nested schema](#nestedblock--ssh_host_public_keys))
- `tags` (List of String) List of tags for the host (order preserved when possible, sorted when tags change)
- `toch` (Boolean) TOCH setting. Defaults to `false`.
- `tofu` (Boolean) TOFU (Trust On First Use) setting. Defaults to `false`.
- `user_message` (String) User message for the host. Defaults to `""`.
- `zone` (String) Zone for the host. Defaults to `""`.

### Read-Only

- `created` (String) Creation timestamp
- `id` (String) Host ID
- `updated` (String) Last update timestamp
- `updated_by` (String) ID of user who last updated the host

<a id="nestedblock--principals"></a>
### Nested Schema for `principals`

#### Required

- `principal` (String) Principal name (username)

#### Optional

- `applications` (List of String) List of applications for the principal
- `command_restrictions` (Block) Command restrictions for the principal (see [below for nested schema](#nestedblock--principals--command_restrictions))
- `passphrase` (String) Principal passphrase (write-only, API returns masked value). Defaults to `""`.
- `roles` (Block List) List of roles for the principal (see [below for nested schema](#nestedblock--principals--roles))
- `rotate` (Boolean) Whether to rotate the principal. Defaults to `false`.
- `service_options` (Block) Service options for the principal (see [below for nested schema](#nestedblock--principals--service_options))
- `source` (String) Principal source. Defaults to `"UI"`.
- `use_for_password_rotation` (Boolean) Use this principal for password rotation. Defaults to `false`.
- `use_user_account` (Boolean) Use user account. Defaults to `false`.
- `username_attribute` (String) Username attribute. Defaults to `""`.

<a id="nestedblock--principals--command_restrictions"></a>
### Nested Schema for `principals.command_restrictions`

#### Optional

- `allow_no_match` (Boolean) Allow commands that don't match any whitelist. Defaults to `false`.
- `audit_match` (Boolean) Audit commands that match whitelists. Defaults to `false`.
- `audit_no_match` (Boolean) Audit commands that don't match whitelists. Defaults to `false`.
- `banner` (String) Banner message to display. Defaults to `""`.
- `default_whitelist` (Block) Default whitelist (see [below for nested schema](#nestedblock--principals--command_restrictions--default_whitelist))
- `enabled` (Boolean) Enable command restrictions. Defaults to `false`.
- `rshell_variant` (String) Shell variant (e.g., bash, sh, powershell). Defaults to `""`.
- `whitelists` (Block List) List of whitelists with roles (see [below for nested schema](#nestedblock--principals--command_restrictions--whitelists))

<a id="nestedblock--principals--command_restrictions--default_whitelist"></a>
### Nested Schema for `principals.command_restrictions.default_whitelist`

#### Optional

- `id` (String) Whitelist ID. Defaults to `""`.
- `name` (String) Whitelist name. Defaults to `""`.

<a id="nestedblock--principals--command_restrictions--whitelists"></a>
### Nested Schema for `principals.command_restrictions.whitelists`

#### Required

- `whitelist` (Block) Whitelist reference (see [below for nested schema](#nestedblock--principals--command_restrictions--whitelists--whitelist))

#### Optional

- `roles` (Block List) Roles for this whitelist (see [below for nested schema](#nestedblock--principals--command_restrictions--whitelists--roles))

<a id="nestedblock--principals--command_restrictions--whitelists--roles"></a>
### Nested Schema for `principals.command_restrictions.whitelists.roles`

#### Required

- `id` (String) Role ID

#### Optional

- `name` (String) Role name

<a id="nestedblock--principals--command_restrictions--whitelists--whitelist"></a>
### Nested Schema for `principals.command_restrictions.whitelists.whitelist`

#### Optional

- `id` (String) Whitelist ID. Defaults to `""`.
- `name` (String) Whitelist name. Defaults to `""`.

<a id="nestedblock--principals--roles"></a>
### Nested Schema for `principals.roles`

#### Required

- `id` (String) Role ID

#### Optional

- `name` (String) Role name

<a id="nestedblock--principals--service_options"></a>
### Nested Schema for `principals.service_options`

#### Optional

- `db` (Block) Database service options (see [below for nested schema](#nestedblock--principals--service_options--db))
- `rdp` (Block) RDP service options (see [below for nested schema](#nestedblock--principals--service_options--rdp))
- `ssh` (Block) SSH service options (see [below for nested schema](#nestedblock--principals--service_options--ssh))
- `vnc` (Block) VNC service options (see [below for nested schema](#nestedblock--principals--service_options--vnc))
- `web` (Block) Web service options (see [below for nested schema](#nestedblock--principals--service_options--web))

<a id="nestedblock--principals--service_options--db"></a>
### Nested Schema for `principals.service_options.db`

#### Optional

- `max_bytes_download` (Number) Maximum bytes for download. Defaults to `0`.
- `max_bytes_upload` (Number) Maximum bytes for upload. Defaults to `0`.

<a id="nestedblock--principals--service_options--rdp"></a>
### Nested Schema for `principals.service_options.rdp`

#### Optional

- `audio` (Boolean) Allow audio. Defaults to `false`.
- `clipboard` (Boolean) Allow clipboard. Defaults to `false`.
- `file_transfer` (Boolean) Allow file transfer. Defaults to `false`.

<a id="nestedblock--principals--service_options--ssh"></a>
### Nested Schema for `principals.service_options.ssh`

#### Optional

- `exec` (Boolean) Allow exec commands. Defaults to `true`.
- `file_transfer` (Boolean) Allow file transfer. Defaults to `true`.
- `other` (Boolean) Allow other SSH features. Defaults to `true`.
- `shell` (Boolean) Allow shell access. Defaults to `true`.
- `tunnels` (Boolean) Allow tunnels. Defaults to `true`.
- `x11` (Boolean) Allow X11 forwarding. Defaults to `true`.

<a id="nestedblock--principals--service_options--vnc"></a>
### Nested Schema for `principals.service_options.vnc`

#### Optional

- `clipboard` (Boolean) Allow clipboard. Defaults to `false`.
- `file_transfer` (Boolean) Allow file transfer. Defaults to `false`.

<a id="nestedblock--principals--service_options--web"></a>
### Nested Schema for `principals.service_options.web`

#### Optional

- `audio` (Boolean) Allow audio. Defaults to `false`.
- `clipboard` (Boolean) Allow clipboard. Defaults to `false`.
- `file_transfer` (Boolean) Allow file transfer. Defaults to `false`.

<a id="nestedblock--services"></a>
### Nested Schema for `services`

#### Required

- `service` (String) Service type (e.g., SSH, RDP, HTTP)

#### Optional

- `address` (String) Service address
- `port` (Number) Service port. Defaults to `22`.
- `source` (String) Service source. Defaults to `"UI"`.
- `ssh_tunnel_port` (Number) SSH tunnel port. Defaults to `0`.
- `use_for_password_rotation` (Boolean) Use this service for password rotation. Defaults to `false`.
- `use_plaintext_vnc` (Boolean) Use plaintext VNC. Defaults to `false`.

<a id="nestedblock--session_recording_options"></a>
### Nested Schema for `session_recording_options`

#### Optional

- `disable_clipboard_recording` (Boolean) Disable clipboard recording. Defaults to `false`.
- `disable_file_transfer_recording` (Boolean) Disable file transfer recording. Defaults to `false`.

<a id="nestedblock--ssh_host_public_keys"></a>
### Nested Schema for `ssh_host_public_keys`

#### Required

- `key` (String) SSH public key

#### Optional

- `fingerprint` (String) SSH key fingerprint

## Import

Hosts can be imported using their ID:

```shell
terraform import privx_host.example 295dbfa2-c291-4021-6caa-70ff2750bb8d
```