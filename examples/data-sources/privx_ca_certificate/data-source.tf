# Read the PrivX authorizer root CA certificates.
# Useful as a drift guard: if a CA is rotated in PrivX, the exposed
# public keys change and dependent configuration produces a plan diff.
data "privx_ca_certificate" "all" {}

# Optionally filter by access group.
data "privx_ca_certificate" "default_group" {
  access_group_id = "00000000-0000-0000-0000-000000000000"
}

output "privx_ca_public_keys" {
  description = "Public keys of the PrivX authorizer CA certificates"
  value       = [for ca in data.privx_ca_certificate.all.certificates : ca.public_key]
}
