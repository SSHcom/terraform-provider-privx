variable "privx_license_code" {
  description = "PrivX license code"
  type        = string
  sensitive   = true
}

resource "privx_license" "this" {
  license_code = var.privx_license_code

  statistics_optin  = true
  refresh_on_update = true
}