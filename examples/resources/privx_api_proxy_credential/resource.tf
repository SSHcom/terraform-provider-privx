resource "privx_api_proxy_credential" "cfg" {
  user_id   = "d0f6418b-728b-4ffc-8b9e-4f01ba1e659b"
  name      = "tf-api_proxy_credential"
  target_id = "17679ec8-9926-4584-be5a-d64b4628d90a"

  not_before = timestamp()
  not_after  = timeadd(timestamp(), "8760h")

  enabled        = true
  type           = "token"
  comment        = "test"
  source_address = []
}