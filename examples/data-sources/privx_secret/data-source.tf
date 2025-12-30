# Get secret by name
data "privx_secret" "database_config" {
  name = "production-database"
}

# Get API keys secret
data "privx_secret" "api_keys" {
  name = "external-service-keys"
}

# Get SSL certificates
data "privx_secret" "ssl_certs" {
  name = "web-server-certificates"
}

# Output secret metadata (non-sensitive information)
output "database_secret_info" {
  description = "Database secret metadata"
  value = {
    name       = data.privx_secret.database_config.name
    author     = data.privx_secret.database_config.author
    created    = data.privx_secret.database_config.created
    updated    = data.privx_secret.database_config.updated
    path       = data.privx_secret.database_config.path
    owner_id   = data.privx_secret.database_config.owner_id
  }
}

# Output role information
output "database_secret_permissions" {
  description = "Database secret access permissions"
  value = {
    read_roles  = data.privx_secret.database_config.read_roles
    write_roles = data.privx_secret.database_config.write_roles
  }
}

# Use secret data in other resources (example with Kubernetes)
resource "kubernetes_secret" "app_database" {
  metadata {
    name      = "app-database-credentials"
    namespace = "production"
  }

  data = {
    host     = data.privx_secret.database_config.data["host"]
    port     = data.privx_secret.database_config.data["port"]
    database = data.privx_secret.database_config.data["database"]
    username = data.privx_secret.database_config.data["username"]
    password = data.privx_secret.database_config.data["password"]
    ssl_mode = data.privx_secret.database_config.data["ssl_mode"]
  }

  type = "Opaque"
}

# Use API keys in application deployment
resource "kubernetes_deployment" "api_service" {
  metadata {
    name      = "api-service"
    namespace = "production"
  }

  spec {
    replicas = 3

    selector {
      match_labels = {
        app = "api-service"
      }
    }

    template {
      metadata {
        labels = {
          app = "api-service"
        }
      }

      spec {
        container {
          name  = "api-service"
          image = "myapp/api-service:latest"

          env {
            name  = "GITHUB_TOKEN"
            value = data.privx_secret.api_keys.data["github_token"]
          }

          env {
            name  = "SLACK_WEBHOOK_URL"
            value = data.privx_secret.api_keys.data["slack_webhook_url"]
          }

          env {
            name  = "AWS_ACCESS_KEY_ID"
            value = data.privx_secret.api_keys.data["aws_access_key_id"]
          }

          env {
            name  = "AWS_SECRET_ACCESS_KEY"
            value = data.privx_secret.api_keys.data["aws_secret_key"]
          }
        }
      }
    }
  }
}

# Create TLS secret from PrivX certificate data
resource "kubernetes_secret" "tls_certificate" {
  metadata {
    name      = "web-server-tls"
    namespace = "production"
  }

  data = {
    "tls.crt" = data.privx_secret.ssl_certs.data["server_cert"]
    "tls.key" = data.privx_secret.ssl_certs.data["private_key"]
    "ca.crt"  = data.privx_secret.ssl_certs.data["ca_bundle"]
  }

  type = "kubernetes.io/tls"
}

# Use secret data to configure external providers
provider "datadog" {
  api_key = data.privx_secret.api_keys.data["datadog_api_key"]
}

# Example: Create AWS resources using credentials from PrivX
data "privx_secret" "aws_credentials" {
  name = "aws-production-account"
}

provider "aws" {
  alias      = "production"
  access_key = data.privx_secret.aws_credentials.data["access_key_id"]
  secret_key = data.privx_secret.aws_credentials.data["secret_access_key"]
  region     = data.privx_secret.aws_credentials.data["region"]
}

# Output specific secret keys (be careful with sensitive data)
output "database_host" {
  description = "Database host from secret"
  value       = data.privx_secret.database_config.data["host"]
  sensitive   = false  # Host is not sensitive
}

output "api_endpoints" {
  description = "Non-sensitive API endpoint information"
  value = {
    slack_configured = length(data.privx_secret.api_keys.data["slack_webhook_url"]) > 0
    github_configured = length(data.privx_secret.api_keys.data["github_token"]) > 0
  }
}

# Example: Conditional resource creation based on secret content
resource "aws_s3_bucket" "backup_bucket" {
  count = length(data.privx_secret.aws_credentials.data["access_key_id"]) > 0 ? 1 : 0
  
  bucket = "app-backups-${random_id.bucket_suffix.hex}"
  
  provider = aws.production
}

resource "random_id" "bucket_suffix" {
  byte_length = 4
}

# Local values using secret data
locals {
  database_url = "postgresql://${data.privx_secret.database_config.data["username"]}:${data.privx_secret.database_config.data["password"]}@${data.privx_secret.database_config.data["host"]}:${data.privx_secret.database_config.data["port"]}/${data.privx_secret.database_config.data["database"]}?sslmode=${data.privx_secret.database_config.data["ssl_mode"]}"
  
  api_config = {
    github_enabled = length(data.privx_secret.api_keys.data["github_token"]) > 0
    slack_enabled  = length(data.privx_secret.api_keys.data["slack_webhook_url"]) > 0
    aws_enabled    = length(data.privx_secret.api_keys.data["aws_access_key_id"]) > 0
  }
}