variable "initial_password" {
  type      = string
  sensitive = true
}

resource "privx_local_user" "example" {
  username  = "example-user"
  full_name = "Terraform provider Test User"
  job_title = "worker"
  email     = "example-user@example.com"

  password                  = var.initial_password
  password_change_required  = true
  tags                      = ["team-a", "oncall"]
}

