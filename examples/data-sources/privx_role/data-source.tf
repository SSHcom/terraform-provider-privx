output "privx_admin_role_name" {
  description = "Name of the PrivX admin role"
  value       = data.privx_role.privx-admin.name
}
