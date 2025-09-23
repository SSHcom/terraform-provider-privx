# Lookup whitelist by name
data "privx_whitelist" "example" {
  name = "example-whitelist"
}

# Lookup whitelist by ID
data "privx_whitelist" "by_id" {
  id = "12345678-1234-1234-1234-123456789012"
}

# Output the whitelist details
output "whitelist_name" {
  description = "Name of the whitelist"
  value       = data.privx_whitelist.example.name
}

output "whitelist_id" {
  description = "ID of the whitelist"
  value       = data.privx_whitelist.example.id
}

output "whitelist_comment" {
  description = "Comment/description of the whitelist"
  value       = data.privx_whitelist.example.comment
}

output "whitelist_type" {
  description = "Type of the whitelist"
  value       = data.privx_whitelist.example.type
}

output "whitelist_patterns" {
  description = "Patterns defined in the whitelist"
  value       = data.privx_whitelist.example.whitelist_patterns
}