#Basic example for credentials
resource "rafay_cloud_credentials_v3" "tftestcredentials" {
  metadata {
    name    = "terraform-demo-credentials-3"
    project = "defaultproject"
  }
  spec {
    type = "ClusterProvisioning"
    provider = "aws"
    credentials {
        type = "AccessKey"
        access_id = "dummy-id"
        secret_key = "dummy-key"
        session_token = "fake-token"
    } 
    sharing {
      enabled = false
    }
  }
}