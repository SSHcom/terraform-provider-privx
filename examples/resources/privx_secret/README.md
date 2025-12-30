# PrivX Secret Resource Examples

This directory contains examples for using the `privx_secret` resource.

## Files

- `resource.tf` - Comprehensive examples showing various secret configurations

## Usage Examples

1. **Basic Secret**: Simple username/password credentials
2. **Database Configuration**: Complete database connection parameters
3. **External API Keys**: Multiple API keys and tokens for external services
4. **SSL/TLS Certificates**: Certificate bundles with private keys
5. **Application Configuration**: Runtime configuration with sensitive values
6. **Service Account Credentials**: Shared service account access

## Required Fields

- `name` - Unique identifier for the secret (cannot be changed after creation)

## Optional Configuration

- `data` - Key-value pairs containing the secret information (sensitive)
- `read_roles` - Roles that can read the secret
- `write_roles` - Roles that can modify the secret
- `owner_id` - Specific owner for the secret

## Security Best Practices

### Access Control
1. **Principle of Least Privilege**: Only grant necessary permissions to roles
2. **Separate Read/Write Access**: Use different roles for read and write operations when appropriate
3. **Role-based Organization**: Group related permissions by functional roles

### Secret Organization
1. **Logical Grouping**: Group related secrets (e.g., all database configs, all API keys)
2. **Environment Separation**: Use different secrets for different environments
3. **Descriptive Naming**: Use clear, consistent naming conventions

### Data Management
1. **Structured Data**: Use consistent key naming within secret data
2. **Avoid Hardcoding**: Don't include environment-specific values in examples
3. **Certificate Handling**: Store complete certificate chains when needed

## Configuration Patterns

### Database Secrets
```hcl
resource "privx_secret" "database" {
  name = "app-database-prod"
  
  data = {
    host     = "db.prod.example.com"
    port     = "5432"
    database = "myapp"
    username = "app_user"
    password = "secure-password"
    ssl_mode = "require"
  }
  
  read_roles {
    id = "app-role-id"
  }
}
```

### API Key Collections
```hcl
resource "privx_secret" "api_keys" {
  name = "external-apis"
  
  data = {
    service1_key = "key1"
    service2_key = "key2"
    webhook_url  = "https://..."
  }
  
  read_roles {
    id = "integration-role-id"
  }
}
```

### Certificate Bundles
```hcl
resource "privx_secret" "certificates" {
  name = "ssl-certificates"
  
  data = {
    certificate = file("cert.pem")
    private_key = file("key.pem")
    ca_bundle   = file("ca.pem")
  }
  
  read_roles {
    id = "web-server-role-id"
  }
}
```

## Integration Examples

### Using with Kubernetes
```hcl
# Create PrivX secret
resource "privx_secret" "k8s_config" {
  name = "kubernetes-config"
  data = {
    username = "k8s-user"
    password = "k8s-password"
  }
}

# Use in Kubernetes secret
resource "kubernetes_secret" "app_secret" {
  metadata {
    name = "app-credentials"
  }
  
  data = {
    username = data.privx_secret.k8s_config.data["username"]
    password = data.privx_secret.k8s_config.data["password"]
  }
}
```

### Using with AWS
```hcl
# Store AWS credentials in PrivX
resource "privx_secret" "aws_creds" {
  name = "aws-credentials"
  data = {
    access_key_id     = "AKIA..."
    secret_access_key = "..."
    region           = "us-west-2"
  }
}

# Configure AWS provider
provider "aws" {
  access_key = data.privx_secret.aws_creds.data["access_key_id"]
  secret_key = data.privx_secret.aws_creds.data["secret_access_key"]
  region     = data.privx_secret.aws_creds.data["region"]
}
```

## Notes

- All secret data is marked as sensitive and won't appear in Terraform logs
- Secret names must be unique within the PrivX vault
- Changes to secret data trigger updates, not recreation
- Role permissions are enforced by PrivX
- Ensure your Terraform state files are properly secured
- Use `terraform import` to manage existing secrets