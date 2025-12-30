# Basic host example for testing
resource "privx_host" "test" {
  common_name     = "test-host"
  addresses       = ["192.168.1.100"]
  source_id       = "test-source-id"
  access_group_id = "test-access-group-id"

  services {
    service = "SSH"
    address = "192.168.1.100"
    port    = 22
    source  = "UI"
  }

  principals {
    principal = "testuser"
    source    = "UI"
    
    roles {
      id   = "test-role-id"
      name = "Test Role"
    }
  }
  tags = ["test", "terraform"]
}

# Basic host with service options
resource "privx_host" "test_with_options" {
  common_name     = "test-host-with-options"
  addresses       = ["192.168.1.101"]
  source_id       = "test-source-id"
  access_group_id = "test-access-group-id"

  services {
    service = "SSH"
    address = "192.168.1.101"
    port    = 22
    source  = "UI"
  }

  principals {
    principal = "testuser"
    source    = "UI"
    
    roles {
      id   = "test-role-id"
      name = "Test Role"
    }

    # Basic service options configuration
    service_options {
      ssh {
        shell         = true
        file_transfer = true
        exec          = true
        tunnels       = true
        x11           = true
        other         = true
      }
    }

    # Basic command restrictions (disabled)
    command_restrictions {
      enabled = false
    }
  }
  
  tags = ["test", "terraform", "service-options"]
}

# Minimal host for debugging
resource "privx_host" "minimal" {
  common_name     = "minimal-host"
  addresses       = ["10.0.0.1"]
  source_id       = "test-source-id"
  access_group_id = "test-access-group-id"
}