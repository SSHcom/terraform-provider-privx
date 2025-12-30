# privx_host (Data Source)

Retrieves information about a PrivX host. You can look up a host by either its ID or common name.

## Example Usage

```terraform
# Get host by ID
data "privx_host" "by_id" {
  id = "295dbfa2-c291-4021-6caa-70ff2750bb8d"
}

# Get host by common name
data "privx_host" "by_name" {
  common_name = "example-server"
}

# Use the host data in other resources
resource "privx_role" "host_admin" {
  name = "Admin for ${data.privx_host.by_name.common_name}"
  
  # Reference host information
  # host_id = data.privx_host.by_name.id
}
```

## Schema

### Optional

- `common_name` (String) Host common name. Either id or common_name must be specified.
- `id` (String) Host UUID. Either id or common_name must be specified.

### Read-Only

- `access_group_id` (String) Access Group ID for the host
- `addresses` (List of String) List of host addresses
- `audit_enabled` (Boolean) Whether audit is enabled for the host
- `cloud_provider` (String) Cloud provider for the host
- `cloud_provider_region` (String) Cloud provider region for the host
- `comment` (String) Comment for the host
- `contact_address` (String) Contact address for the host
- `created` (String) Creation timestamp
- `deployable` (Boolean) Whether the host is deployable
- `disabled` (String) Whether the host is disabled
- `distinguished_name` (String) Distinguished name for the host
- `external_id` (String) External ID for the host
- `host_classification` (String) Host classification
- `host_type` (String) Host type
- `instance_id` (String) Instance ID for the host
- `organization` (String) Organization for the host
- `organizational_unit` (String) Organizational unit for the host
- `password_rotation_enabled` (Boolean) Whether password rotation is enabled
- `principals` (Block List) List of principals for the host (see [below for nested schema](#nestedblock--principals))
- `services` (Block List) List of services for the host (see [below for nested schema](#nestedblock--services))
- `session_recording_options` (Block) Session recording options (see [below for nested schema](#nestedblock--session_recording_options))
- `source_id` (String) Source ID for the host
- `ssh_host_public_keys` (Block List) List of SSH host public keys (see [below for nested schema](#nestedblock--ssh_host_public_keys))
- `tags` (List of String) List of tags for the host (sorted alphabetically)
- `toch` (Boolean) TOCH setting
- `tofu` (Boolean) TOFU (Trust On First Use) setting
- `updated` (String) Last update timestamp
- `updated_by` (String) ID of user who last updated the host
- `user_message` (String) User message for the host
- `zone` (String) Zone for the host

<a id="nestedblock--principals"></a>
### Nested Schema for `principals`

#### Read-Only

- `applications` (List of String) List of applications for the principal
- `command_restrictions` (Block) Command restrictions for the principal (see [below for nested schema](#nestedblock--principals--command_restrictions))
- `passphrase` (String) Principal passphrase (masked by API)
- `principal` (String) Principal name
- `roles` (Block List) List of roles for the principal (see [below for nested schema](#nestedblock--principals--roles))
- `rotate` (Boolean) Whether to rotate the principal
- `service_options` (Block) Service options for the principal (see [below for nested schema](#nestedblock--principals--service_options))
- `source` (String) Principal source
- `use_for_password_rotation` (Boolean) Use this principal for password rotation
- `use_user_account` (Boolean) Use user account
- `username_attribute` (String) Username attribute

<a id="nestedblock--principals--command_restrictions"></a>
### Nested Schema for `principals.command_restrictions`

#### Read-Only

- `allow_no_match` (Boolean) Allow commands that don't match any whitelist
- `audit_match` (Boolean) Audit commands that match whitelists
- `audit_no_match` (Boolean) Audit commands that don't match whitelists
- `banner` (String) Banner message to display
- `default_whitelist` (Block) Default whitelist (see [below for nested schema](#nestedblock--principals--command_restrictions--default_whitelist))
- `enabled` (Boolean) Enable command restrictions
- `rshell_variant` (String) Shell variant (e.g., bash, sh, powershell)
- `whitelists` (Block List) List of whitelists with roles (see [below for nested schema](#nestedblock--principals--command_restrictions--whitelists))

<a id="nestedblock--principals--command_restrictions--default_whitelist"></a>
### Nested Schema for `principals.command_restrictions.default_whitelist`

#### Read-Only

- `id` (String) Whitelist ID
- `name` (String) Whitelist name

<a id="nestedblock--principals--command_restrictions--whitelists"></a>
### Nested Schema for `principals.command_restrictions.whitelists`

#### Read-Only

- `roles` (Block List) Roles for this whitelist (see [below for nested schema](#nestedblock--principals--command_restrictions--whitelists--roles))
- `whitelist` (Block) Whitelist reference (see [below for nested schema](#nestedblock--principals--command_restrictions--whitelists--whitelist))

<a id="nestedblock--principals--command_restrictions--whitelists--roles"></a>
### Nested Schema for `principals.command_restrictions.whitelists.roles`

#### Read-Only

- `id` (String) Role ID
- `name` (String) Role name

<a id="nestedblock--principals--command_restrictions--whitelists--whitelist"></a>
### Nested Schema for `principals.command_restrictions.whitelists.whitelist`

#### Read-Only

- `id` (String) Whitelist ID
- `name` (String) Whitelist name

<a id="nestedblock--principals--roles"></a>
### Nested Schema for `principals.roles`

#### Read-Only

- `id` (String) Role ID
- `name` (String) Role name

<a id="nestedblock--principals--service_options"></a>
### Nested Schema for `principals.service_options`

#### Read-Only

- `db` (Block) Database service options (see [below for nested schema](#nestedblock--principals--service_options--db))
- `rdp` (Block) RDP service options (see [below for nested schema](#nestedblock--principals--service_options--rdp))
- `ssh` (Block) SSH service options (see [below for nested schema](#nestedblock--principals--service_options--ssh))
- `vnc` (Block) VNC service options (see [below for nested schema](#nestedblock--principals--service_options--vnc))
- `web` (Block) Web service options (see [below for nested schema](#nestedblock--principals--service_options--web))

<a id="nestedblock--principals--service_options--db"></a>
### Nested Schema for `principals.service_options.db`

#### Read-Only

- `max_bytes_download` (Number) Maximum bytes for download
- `max_bytes_upload` (Number) Maximum bytes for upload

<a id="nestedblock--principals--service_options--rdp"></a>
### Nested Schema for `principals.service_options.rdp`

#### Read-Only

- `audio` (Boolean) Allow audio
- `clipboard` (Boolean) Allow clipboard
- `file_transfer` (Boolean) Allow file transfer

<a id="nestedblock--principals--service_options--ssh"></a>
### Nested Schema for `principals.service_options.ssh`

#### Read-Only

- `exec` (Boolean) Allow exec commands
- `file_transfer` (Boolean) Allow file transfer
- `other` (Boolean) Allow other SSH features
- `shell` (Boolean) Allow shell access
- `tunnels` (Boolean) Allow tunnels
- `x11` (Boolean) Allow X11 forwarding

<a id="nestedblock--principals--service_options--vnc"></a>
### Nested Schema for `principals.service_options.vnc`

#### Read-Only

- `clipboard` (Boolean) Allow clipboard
- `file_transfer` (Boolean) Allow file transfer

<a id="nestedblock--principals--service_options--web"></a>
### Nested Schema for `principals.service_options.web`

#### Read-Only

- `audio` (Boolean) Allow audio
- `clipboard` (Boolean) Allow clipboard
- `file_transfer` (Boolean) Allow file transfer

<a id="nestedblock--services"></a>
### Nested Schema for `services`

#### Read-Only

- `address` (String) Service address
- `port` (Number) Service port
- `service` (String) Service type (e.g., SSH, RDP, HTTP)
- `source` (String) Service source
- `ssh_tunnel_port` (Number) SSH tunnel port
- `status` (String) Service status
- `status_updated` (String) Service status last updated
- `use_for_password_rotation` (Boolean) Use this service for password rotation
- `use_plaintext_vnc` (Boolean) Use plaintext VNC

<a id="nestedblock--session_recording_options"></a>
### Nested Schema for `session_recording_options`

#### Read-Only

- `disable_clipboard_recording` (Boolean) Disable clipboard recording
- `disable_file_transfer_recording` (Boolean) Disable file transfer recording

<a id="nestedblock--ssh_host_public_keys"></a>
### Nested Schema for `ssh_host_public_keys`

#### Read-Only

- `fingerprint` (String) SSH key fingerprint
- `key` (String) SSH public key