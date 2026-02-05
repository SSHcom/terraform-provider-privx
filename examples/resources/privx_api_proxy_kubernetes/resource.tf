

data "privx_role" "users" {
  name = "privx-user"
}

variable "kubernetes_api_host" {
  description = "Kubernetes API address in host:port form (for API target Authorized Endpoints)"
  type        = string
  default     = "my-eks.amazonaws.com:443"
}

variable "kubernetes_api_token" {
  description = "Bearer token PrivX uses to authenticate to the Kubernetes API"
  type        = string
  sensitive   = true
}

variable "kubernetes_cluster_ca_pem_path" {
  description = "Path to EKS cluster CA PEM file (trust anchor for PrivX -> Kubernetes)"
  type        = string
  default     = "./eks-cluster-ca.pem"
}

variable "privx_user_id" {
  type        = string
  description = "PrivX user ID (UUID). Leave empty to use current user."
  default = ""
}

resource "privx_api_target" "kube" {
  name    = "tf-kube-8932818406016930558"
  comment = "test"

  roles = [
    {
      id   = data.privx_role.users.id
      name = data.privx_role.users.name
    }
  ]

  authorized_endpoints = [
    {
      host      = var.kubernetes_api_host
      protocols = ["https"]
      methods   = ["*"]
      paths     = ["**"]
    }
  ]

   target_credential = {
      type         = "token"
      bearer_token = var.kubernetes_api_token
   }

    tls_trust_anchors = file(var.kubernetes_cluster_ca_pem_path)

  audit_enabled = false
}

resource "privx_api_proxy_credential" "kube" {
  user_id   = var.privx_user_id
  name      = "tf-kube-8932818406016930558"
  target_id = privx_api_target.kube.id

  not_before = timestamp()
  not_after  = timeadd(timestamp(), "8760h")

  enabled        = true
  type           = "token"
  comment        = "test"
  source_address = []
}

data "privx_api_proxy_config" "this" {}

# Proxy URL (pick first public address)
output "privx_proxy_url" {
  description = "Use as kubeconfig cluster.proxy-url"
  value       = data.privx_api_proxy_config.this.addresses[0]
}

# Upstream Kubernetes API server (your real cluster API endpoint)
output "kube_api_server" {
  description = "Use as kubeconfig cluster.server"
  value       = "https://${privx_api_target.kube.authorized_endpoints[0].host}"
}

# PrivX API Proxy CA cert chain (PEM) as returned by data source
output "privx_proxy_ca_pem" {
  description = "PrivX API Proxy CA certificate chain (PEM)"
  value       = data.privx_api_proxy_config.this.ca_certificate_chain
}

# Base64 form for kubeconfig certificate-authority-data
output "privx_proxy_ca_b64" {
  description = "Use as kubeconfig cluster.certificate-authority-data"
  value       = base64encode(data.privx_api_proxy_config.this.ca_certificate_chain)
}

# User token for kubeconfig users[].user.token
# In your provider schema this is exposed as `secret` (you asserted it in the acc test).
output "privx_kubectl_token" {
  description = "Use as kubeconfig users[].user.token"
  value       = privx_api_proxy_credential.kube.secret
  sensitive   = true
}
