provider "privx" {
  api_base_url = "https://<hostname>"
  /* Oauth auth can be replaced by token */
  //api_bearer_token        = ""
  api_oauth_client_id     = "privx-external"
  api_oauth_client_secret = ""
  api_client_id           = ""
  api_client_secret       = ""
  debug                   = false
}
