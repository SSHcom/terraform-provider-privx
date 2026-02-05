resource "privx_api_client" "example" {
  name = "test-provider-terraform"
  roles = [
    {
      id   = "dc012ffa-540c-563a-5293-51a644ce273b"
      name = "test-role"
    }
  ]
}
