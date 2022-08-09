resource "rafay_cluster_network_policy_rule" "tfdemocnpr1" {
  metadata {
    name    = "tfdemocnpr1"
    project = "terraform"
  }
  spec {
    version = "example-version"
    sharing {
        enabled = false
    }
  }
}