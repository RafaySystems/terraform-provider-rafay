#Basic example for credentials
resource "rafay_credentials" "tftestcredentials1" {
  metadata {
    name    = "tftestcredentials1"
    project = "terraform"
  }
  spec {
    type = "CloudProvisioning"
    provider = "aws"
    credentials {
      type = "access-based"
     access_id = "dummy-id"
     secret_key = "dummy-key"
     session_token = "dummy-session"
    } 
    sharing {
      enabled = false
    }
  }
}