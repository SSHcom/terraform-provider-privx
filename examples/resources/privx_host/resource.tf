terraform {
  required_providers {
    privx = {
      source = "hashicorp/privx"
    }
  }
}

provider "privx" {}

variable "host_address" {
  type    = string
  default = "10.0.0.10"
}

variable "initial_passphrase" {
  type      = string
  sensitive = true
}

data "privx_script_template" "st" {
  name = "Linux per account command template"
}

data "privx_password_policy" "pp" {
  name = "PrivX default password policy"
}

data "privx_access_group" "ag" {
  name = "Default"
}

resource "privx_host" "test" {
  common_name     = "tf-acc-host-rotation"
  addresses       = [var.host_address]
  access_group_id = data.privx_access_group.ag.id

  external_id = "tf-ext-demo"
  comment     = "created by terraform"

  services = [{
    service                   = "SSH"
    address                   = var.host_address
    port                      = 22
    use_for_password_rotation = true
  }]

  password_rotation_enabled = true

  password_rotation = {
    access_group_id                = data.privx_access_group.ag.id
    use_main_account               = true
    operating_system               = "LINUX"
    protocol                       = "SSH"
    certificate_validation_options = "DISABLED"
    password_policy_id             = data.privx_password_policy.pp.id
    script_template_id             = data.privx_script_template.st.id
  }

  principals = [{
    principal                 = "tf-acc-rotate"
    rotate                    = true
    use_for_password_rotation = true
    use_user_account          = false
    passphrase                = var.initial_passphrase
  }]
}
