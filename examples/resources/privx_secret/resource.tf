# Basic secret example
resource "privx_secret" "basic_secret" {
  name = "basic-credentials"
  
  data = {
    username = "admin"
    password = "secure-password-123"
  }

  read_roles {
    id   = "admin-role-id"
    name = "Administrators"
  }

  write_roles {
    id   = "admin-role-id"
    name = "Administrators"
  }
}

# Database connection secret
resource "privx_secret" "database_config" {
  name = "production-database"
  
  data = {
    host     = "db.prod.example.com"
    port     = "5432"
    database = "myapp_production"
    username = "app_user"
    password = "db-password-xyz789"
    ssl_mode = "require"
  }

  read_roles {
    id   = "app-developers-role-id"
    name = "Application Developers"
  }

  read_roles {
    id   = "devops-role-id"
    name = "DevOps Team"
  }

  write_roles {
    id   = "devops-role-id"
    name = "DevOps Team"
  }
}


# SSL/TLS certificates secret
resource "privx_secret" "ssl_certificates" {
  name     = "web-server-certificates"
  owner_id = "cert-manager-user-id"
  
  data = {
    server_cert = <<-EOT
      -----BEGIN CERTIFICATE-----
      MIIDXTCCAkWgAwIBAgIJAKoK/heBjcOuMA0GCSqGSIb3DQEBBQUAMEUxCzAJBgNV
      BAYTAkFVMRMwEQYDVQQIDApTb21lLVN0YXRlMSEwHwYDVQQKDBhJbnRlcm5ldCBX
      aWRnaXRzIFB0eSBMdGQwHhcNMTcwODI3MjM1NzU5WhcNMTgwODI3MjM1NzU5WjBF
      MQswCQYDVQQGEwJBVTETMBEGA1UECAwKU29tZS1TdGF0ZTEhMB8GA1UECgwYSW50
      ZXJuZXQgV2lkZ2l0cyBQdHkgTHRkMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIB
      CgKCAQEAwU1/1QcHhJ1hW62VwP2OlQAGZhYpuUYuAjxVIdLhLz1cWgjdX7oKmST5
      De/BjgnQZocKx/fQg4QFy+InN0DE2kkInVWHSiviLvDHyeHcwlM/s+wuVuoCpGrH
      L0Pa1dTkpQPjqA0AMxHzpBuOIA==
      -----END CERTIFICATE-----
    EOT
    
    private_key = <<-EOT
      -----BEGIN PRIVATE KEY-----
      MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQDBTX/VBweEnWFb
      rZXA/Y6VAAZmFim5Ri4CPFUh0uEvPVxaCN1fugqZJPkN78GOCdBmhwrH99CDhAXL
      4ic3QMTaSSidVYdKK+Iu8MfJ4dzCUz+z7C5W6gKkascvQ9rV1OSlA+OoDQAzEfOk
      G44gDQABAAECggEAMIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIB
      AQDBTZ/VBweEnWFbrZXA/Y6VAAZmFim5Ri4CPFUh0uEvPVxaCN1fugqZJPkN78GO
      CdBmhwrH99CDhAXL4ic3QMTaSSidVYdKK+Iu8MfJ4dzCUz+z7C5W6gKkascvQ9rV
      1OSlA+OoDQAzEfOkG44gDQABAAECggEAMIIEvQIBADANBgkqhkiG9w0BAQEFAASC
      BKcwggSjAgEAAoIBAQDBTX/VBweEnWFbrZXA/Y6VAAZmFim5Ri4CPFUh0uEvPVxa
      CN1fugqZJPkN78GOCdBmhwrH99CDhAXL4ic3QMTaSSidVYdKK+Iu8MfJ4dzCUz+z
      7C5W6gKkascvQ9rV1OSlA+OoDQAzEfOkG44gDQABAAECggEA
      -----END PRIVATE KEY-----
    EOT
    
    ca_bundle = <<-EOT
      -----BEGIN CERTIFICATE-----
      MIIDXTCCAkWgAwIBAgIJAKoK/heBjcOuMA0GCSqGSIb3DQEBBQUAMEUxCzAJBgNV
      BAYTAkFVMRMwEQYDVQQIDApTb21lLVN0YXRlMSEwHwYDVQQKDBhJbnRlcm5ldCBX
      aWRnaXRzIFB0eSBMdGQwHhcNMTcwODI3MjM1NzU5WhcNMTgwODI3MjM1NzU5WjBF
      MQswCQYDVQQGEwJBVTETMBEGA1UECAwKU29tZS1TdGF0ZTEhMB8GA1UECgwYSW50
      ZXJuZXQgV2lkZ2l0cyBQdHkgTHRkMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIB
      CgKCAQEAwU1/1QcHhJ1hW62VwP2OlQAGZhYpuUYuAjxVIdLhLz1cWgjdX7oKmST5
      De/BjgnQZocKx/fQg4QFy+InN0DE2kkInVWHSiviLvDHyeHcwlM/s+wuVuoCpGrH
      L0Pa1dTkpQPjqA0AMxHzpBuOIA==
      -----END CERTIFICATE-----
    EOT
  }

  read_roles {
    id   = "web-admin-role-id"
    name = "Web Administrators"
  }

  read_roles {
    id   = "security-team-role-id"
    name = "Security Team"
  }

  write_roles {
    id   = "cert-manager-role-id"
    name = "Certificate Managers"
  }
}

# Application configuration secret
resource "privx_secret" "app_config" {
  name = "microservice-config"
  
  data = {
    jwt_secret           = "super-secret-jwt-signing-key-12345"
    encryption_key       = "aes-256-encryption-key-67890"
    redis_url           = "redis://redis.internal:6379/0"
    elasticsearch_url   = "https://elastic.internal:9200"
    log_level          = "info"
    debug_mode         = "false"
    feature_flag_api   = "https://features.internal/api/v1"
    metrics_endpoint   = "https://metrics.internal/prometheus"
  }

  read_roles {
    id   = "backend-developers-role-id"
    name = "Backend Developers"
  }

  read_roles {
    id   = "sre-role-id"
    name = "Site Reliability Engineers"
  }

  write_roles {
    id   = "platform-team-role-id"
    name = "Platform Team"
  }
}

# Shared service account credentials
resource "privx_secret" "service_accounts" {
  name = "shared-service-accounts"
  
  data = {
    monitoring_user     = "monitoring_svc"
    monitoring_password = "mon-pass-xyz123"
    backup_user        = "backup_svc"
    backup_password    = "backup-pass-abc456"
    deploy_user        = "deploy_svc"
    deploy_ssh_key     = <<-EOT
      -----BEGIN OPENSSH PRIVATE KEY-----
      b3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAAAFwAAAAdzc2gtcn
      NhAAAAAwEAAQAAAQEAwU1/1QcHhJ1hW62VwP2OlQAGZhYpuUYuAjxVIdLhLz1cWgjdX7
      oKmST5De/BjgnQZocKx/fQg4QFy+InN0DE2kkInVWHSiviLvDHyeHcwlM/s+wuVuoC
      pGrHL0Pa1dTkpQPjqA0AMxHzpBuOIAAAAAwEAAQAAAQEAwU1/1QcHhJ1hW62VwP2OlQ
      AGZhYpuUYuAjxVIdLhLz1cWgjdX7oKmST5De/BjgnQZocKx/fQg4QFy+InN0DE2kkI
      nVWHSiviLvDHyeHcwlM/s+wuVuoCpGrHL0Pa1dTkpQPjqA0AMxHzpBuOIAAAAAwEAAQ
      AAAQEAwU1/1QcHhJ1hW62VwP2OlQAGZhYpuUYuAjxVIdLhLz1cWgjdX7oKmST5De/B
      jgnQZocKx/fQg4QFy+InN0DE2kkInVWHSiviLvDHyeHcwlM/s+wuVuoCpGrHL0Pa1d
      TkpQPjqA0AMxHzpBuOIAAAAAwEAAQAAAQEAwU1/1QcHhJ1hW62VwP2OlQAGZhYpuUY
      uAjxVIdLhLz1cWgjdX7oKmST5De/BjgnQZocKx/fQg4QFy+InN0DE2kkInVWHSivi
      LvDHyeHcwlM/s+wuVuoCpGrHL0Pa1dTkpQPjqA0AMxHzpBuOIAAAAA
      -----END OPENSSH PRIVATE KEY-----
    EOT
  }

  read_roles {
    id   = "automation-role-id"
    name = "Automation Systems"
  }

  write_roles {
    id   = "security-admin-role-id"
    name = "Security Administrators"
  }
}