resource "rafay_network_policy_profile" "tfdemonetworkpolicyprofile1" {
  metadata {
    name    = "tfdemonetworkpolicyprofile1"
    project = "terraform"
  }
  spec {
    version = "example-version"
    sharing {
      enabled = false
    }
    installation_params {
      policy_enforcement_mode = "default"
    }
  }
}