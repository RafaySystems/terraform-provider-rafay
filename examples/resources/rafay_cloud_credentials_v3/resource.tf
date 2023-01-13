
#Basic example for credentials
resource "rafay_cloud_credentials_v3" "tftestcredentials" {
  metadata {
    name    = "terraform-demo-credentials-2"
    project = "cang-test"
  }
  spec {
    type = "CostManagement"
    provider = "aws"
    credentials {
        type = "AccessKey"
        access_id = "dummy-id"
        secret_key = "dummy-key"
    } 
    sharing {
      enabled = false
    }
  }
}