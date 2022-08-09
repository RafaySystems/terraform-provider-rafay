resource "rafay_namespace_network_policy_rule" "tfdemonnpr1" {
  metadata {
    name    = "tfdemonnpr1"
    project = "terraform"
  }
  spec {
    version = "example-version"
    sharing {
        enabled = false
    }
  }
}