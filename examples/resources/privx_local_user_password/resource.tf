resource "privx_local_user_password" "rotate" {
  user_id  = privx_local_user.test.id
  password = "new_password"
}