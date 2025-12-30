# Basic SSH host example
resource "privx_host" "basic_ssh" {
  common_name     = "web-server-01"
  addresses       = ["192.168.1.100", "web01.example.com"]
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

# Host with password rotation enabled
resource "privx_host" "with_password_rotation" {
  common_name               = "db-server-01"
  addresses                 = ["10.0.1.50"]
  source_id                 = "a0ad72dc-b4aa-4a53-b7e0-14902b8b18bd"
  access_group_id           = "bfe74d8b-feda-46f1-7ac3-b37ef5b15e3b"
  password_rotation_enabled = true
  audit_enabled             = true

  services {
    service                   = "SSH"
    address                   = "10.0.1.50"
    port                      = 22
    use_for_password_rotation = true
    source                    = "UI"
  }

  principals {
    principal                 = "dbadmin"
    use_for_password_rotation = true
    source                    = "UI"
    
    roles {
      id   = "db-admin-role-id"
      name = "Database Admin"
    }
  }

  session_recording_options {
    disable_clipboard_recording     = false
    disable_file_transfer_recording = false
  }

  tags = ["production", "database"]
}

# AWS EC2 instance host
resource "privx_host" "aws_ec2" {
  common_name           = "ec2-web-server"
  addresses             = ["ec2-35-179-143-1.eu-west-2.compute.amazonaws.com", "35.179.143.1", "172.31.39.33"]
  external_id           = "i-0b4b5280d20be9fcb"
  source_id             = "a0ad72dc-b4aa-4a53-b7e0-14902b8b18bd"
  access_group_id       = "bfe74d8b-feda-46f1-7ac3-b37ef5b15e3b"
  cloud_provider        = "AWS"
  cloud_provider_region = "eu-west-2"

  services {
    service = "SSH"
    address = "172.31.39.33"
    port    = 22
    source  = "UI"
  }

  principals {
    principal = "ec2-user"
    source    = "UI"
    
    roles {
      id   = "65127d0a-b2df-403c-be19-e2960d10de4d"
      name = "privx-admin"
    }
    
    roles {
      id   = "d576edba-574b-5614-7124-ee42e2486b49"
      name = "db-job"
    }
  }

  ssh_host_public_keys {
    key         = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC4Xa8DIsUR2NRKopCN8zWO9Vc1V+KUPtwMZZ14yJ6QFvNWe99Me+OJYPiP1ruLpMX7VfynoMO2ZD2aMdUTFG3gBQdLZ/9ydMZT91RUmYIxYnpXoNRlQSWPClZwDV0dNW3Q9aXQJzc3CK8H5o7WatylCrLH704ciiptUzsT7XgVv8YE/sn5Ly6tE9z6aIe0JAejfAh6KFs7vowDpyV99s+RLQb+KzyTzrUVK1Jc6mgYbYdd6RxbElKHOULhg+aAdzYWY4GaRf1H9uRyNY0zbSmf5lBoOXZ1BxLvJrRxijSrMT+IET+Um/BhEGhKJ4u0e3Wj8CPeQc9JIBmzBvaHYIJkX4kvmnprsmysv4+YXPZb4vuvZWjv4WnyIYh25dNHjMfhzPpjhZVxEBIsX8oF2Ka1kCVOCIRc7drXQ+6YaSw3YhwOqGtKi3Z98oHSUTevKa/YpvpUeaCmxkLSZZDxa2YpIZRtGv10/OPGXotVS1oFckHi7nlsqnUueqfOPZbQKfs= root@ip-172-31-39-33.eu-west-2.compute.internal"
    fingerprint = "SHA256:Ge1xVW9GzDR175Xh0+BulrKcSlVkMx7TSM3w0UNSLDw"
  }

  ssh_host_public_keys {
    key         = "ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBDuEdFwfwRsYH6sbgq+p0Tju4IG9QI2RT/91Mu53yyMMXxgTh2pzRaTIjBBDYTgNFUwzOVqzeqA2xn2dsamiD8g= root@ip-172-31-39-33.eu-west-2.compute.internal"
    fingerprint = "SHA256:e3M2MprxPLGWAfn0ueUkk7JfBXkYRCRdFnjl72MhSm0"
  }

  tags = ["aws", "ec2", "production"]
}

# Windows RDP host with comprehensive service options and command restrictions
resource "privx_host" "windows_rdp_advanced" {
  common_name     = "windows-server-01"
  addresses       = ["192.168.1.200", "win01.example.com"]
  source_id       = "a0ad72dc-b4aa-4a53-b7e0-14902b8b18bd"
  access_group_id = "bfe74d8b-feda-46f1-7ac3-b37ef5b15e3b"
  host_type       = "Windows Server"
  audit_enabled   = true

  services {
    service = "RDP"
    address = "192.168.1.200"
    port    = 3389
    source  = "UI"
  }

  services {
    service = "SSH"
    address = "192.168.1.200"
    port    = 22
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
      # SSH service options - restrictive for security
      ssh {
        shell         = true
        file_transfer = false  # Disabled for security
        exec          = true
        tunnels       = false  # Disabled for security
        x11           = false  # Not applicable for Windows
        other         = false
      }
      
      # RDP service options - allow multimedia features
      rdp {
        file_transfer = true
        audio         = true
        clipboard     = true
      }
      
      # Web service options - moderate restrictions
      web {
        file_transfer = false
        audio         = true
        clipboard     = true
      }
      
      # VNC service options - basic functionality
      vnc {
        file_transfer = false
        clipboard     = true
      }
      
      # Database service options - with upload/download limits
      db {
        max_bytes_upload   = 1048576   # 1MB
        max_bytes_download = 10485760  # 10MB
      }
    }

    # Configure command restrictions for PowerShell
    command_restrictions {
      enabled         = true
      rshell_variant  = "powershell"
      allow_no_match  = false
      audit_match     = true
      audit_no_match  = true
      banner          = "WARNING: Authorized access only. All activities are monitored and logged."
      
      # Default whitelist for basic Windows commands
      default_whitelist {
        id   = "11c24354-e280-4fb2-791b-9f5c80dd2b01"
        name = "windows_basic_commands"
      }
      
      # Additional whitelists for admin users
      whitelists {
        whitelist {
          id   = "7ea51435-59b9-4e6d-422d-4acec3173260"
          name = "windows_admin_commands"
        }
        
        roles {
          id   = "65127d0a-b2df-403c-be19-e2960d10de4d"
          name = "privx-admin"
        }
        
        roles {
          id   = "windows-admin-role-id"
          name = "Windows Admin"
        }
      }
      
      # Database-specific commands for DB admins
      whitelists {
        whitelist {
          id   = "db-commands-whitelist-id"
          name = "database_commands"
        }
        
        roles {
          id   = "db-admin-role-id"
          name = "Database Admin"
        }
      }
    }
  }

  # Service account with different restrictions
  principals {
    principal = "svc_backup"
    source    = "UI"
    
    roles {
      id   = "backup-service-role-id"
      name = "Backup Service"
    }

    # More restrictive service options for service accounts
    service_options {
      ssh {
        shell         = false  # No interactive shell for service accounts
        file_transfer = true   # Allow file operations
        exec          = true   # Allow command execution
        tunnels       = false
        x11           = false
        other         = false
      }
      
      rdp {
        file_transfer = true
        audio         = false  # No audio for service accounts
        clipboard     = false  # No clipboard for service accounts
      }
    }

    # Strict command restrictions for service accounts
    command_restrictions {
      enabled         = true
      rshell_variant  = "powershell"
      allow_no_match  = false
      audit_match     = true
      audit_no_match  = true
      banner          = "SERVICE ACCOUNT: Automated operations only."
      
      default_whitelist {
        id   = "service-account-whitelist-id"
        name = "service_account_commands"
      }
    }
  }

  session_recording_options {
    disable_clipboard_recording     = false
    disable_file_transfer_recording = false
  }

  tags = ["windows", "production", "rdp", "database"]
}

# Linux database server with database-specific service options
resource "privx_host" "database_server" {
  common_name     = "postgres-db-01"
  addresses       = ["10.0.2.100"]
  source_id       = "a0ad72dc-b4aa-4a53-b7e0-14902b8b18bd"
  access_group_id = "bfe74d8b-feda-46f1-7ac3-b37ef5b15e3b"
  host_type       = "Database Server"

  services {
    service = "SSH"
    address = "10.0.2.100"
    port    = 22
    source  = "UI"
  }

  services {
    service = "POSTGRESQL"
    address = "10.0.2.100"
    port    = 5432
    source  = "UI"
  }

  principals {
    principal = "postgres"
    source    = "UI"
    
    roles {
      id   = "db-admin-role-id"
      name = "Database Admin"
    }

    # Database-focused service options
    service_options {
      ssh {
        shell         = true
        file_transfer = true
        exec          = true
        tunnels       = false
        x11           = false
        other         = false
      }
      
      # Database service options with higher limits for DB operations
      db {
        max_bytes_upload   = 104857600  # 100MB for database imports
        max_bytes_download = 1073741824 # 1GB for database exports
      }
    }

    # Database-specific command restrictions
    command_restrictions {
      enabled         = true
      rshell_variant  = "bash"
      allow_no_match  = false
      audit_match     = true
      audit_no_match  = true
      banner          = "DATABASE SERVER: Use approved database commands only."
      
      default_whitelist {
        id   = "linux-basic-commands-id"
        name = "linux_basic_commands"
      }
      
      whitelists {
        whitelist {
          id   = "postgresql-commands-id"
          name = "postgresql_admin_commands"
        }
        
        roles {
          id   = "db-admin-role-id"
          name = "Database Admin"
        }
      }
    }
  }

  tags = ["linux", "database", "postgresql", "production"]
}