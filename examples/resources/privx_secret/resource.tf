resource "privx_secret" "example" {
  name = "bar"
  data = jsonencode({})
  write_roles = [
    {
      id   = "4f22ba9d-d3a0-47f9-b5f8-0345e869e7ea"
      name = "role_[Default]_[ADMIN]"
    }
  ]
  read_roles = [
    {
      id   = "1379d493-e746-4308-608c-ba465b416787"
      name = "role_[Default]_[USER]"
    }
  ]
}