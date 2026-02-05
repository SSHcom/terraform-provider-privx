resource "privx_whitelist" "example" {
  name    = "example-whitelist"
  comment = "Example whitelist for command restrictions"
  type    = "glob"
  whitelist_patterns = [
    "ls -la",
    "cat /etc/passwd",
    "grep -r 'pattern' /var/log/",
    "systemctl status *",
    "ps aux"
  ]
}

# Whitelist with minimal configuration
resource "privx_whitelist" "minimal" {
  name = "minimal-whitelist"
  type = "glob"
}

# Whitelist for development environment
resource "privx_whitelist" "dev_commands" {
  name    = "dev-environment-whitelist"
  comment = "Allowed commands for development environment"
  type    = "glob"
  whitelist_patterns = [
    "git *",
    "npm *",
    "yarn *",
    "docker *",
    "kubectl *",
    "helm *"
  ]
}