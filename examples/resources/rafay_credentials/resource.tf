
#Basic example for credentials
resource "rafay_credentials" "tftestcredentials1" {
  metadata {
    name    = "tftestcredentials1"
    project = "defaultproject"
  }
  spec {
    type = "ClusterProvisioning"
    provider = "aws"
    credentials {
        type = "AccessKey"
        access_id = "dummy-id"
        secret_key = "dummy-key"
        session_token = "dummy-session"
    } 
    sharing {
      enabled = false
    }
  }
}