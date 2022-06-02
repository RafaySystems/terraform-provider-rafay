resource "rafay_pipeline" "tfdemopipeline1" {
  metadata {
    name    = "email-test"
    project = "terraform"
  }
  spec {
    active = false
    sharing {
      enabled = false
    }
    stages {
      config {
        approvers {
          sso_user  = false
          user_name = "hardik@rafay.co"
        }
        timeout = "2m0s"
        type    = "Email"
      }
      name = "email"
      type = "Approval"
    }
  }
}