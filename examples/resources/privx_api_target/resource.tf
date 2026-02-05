
    data "privx_role" "existing_role" {
      name = "privx-user"
    }

    resource "privx_api_target" "k8s" {
      name    = "tf-api-target-xbldna"
      comment = "created by acc test"

      roles = [{
        id   = data.privx_role.existing_role.id
        name = data.privx_role.existing_role.name
      }]

      authorized_endpoints = [{
        host      = "kubernetes.example.com:6443"
        protocols = ["*"]
        methods   = ["*"]
        paths     = ["**"]
      }]

      target_credential = {
        type         = "token"
        bearer_token = "<KUBERNETES_API_TARGET_TOKEN>"
      }

      tls_trust_anchors = file("/tmp/trust-anchor.pem")

      tls_insecure_skip_verify = false
    }
