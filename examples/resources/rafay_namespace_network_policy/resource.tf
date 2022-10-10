resource "rafay_namespace_network_policy" "tfdemonnp2" {
  metadata {
    name    = "tfdemonnp2"
    project = "terraform"
  }
  spec {
    version = "v0"
    rules {
      name = "namespace-network-rule"
      version = "v0"
    }
    sharing {
      enabled = false
    }
  }
}
