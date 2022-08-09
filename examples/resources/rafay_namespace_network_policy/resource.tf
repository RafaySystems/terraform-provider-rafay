resource "rafay_namespace_network_policy" "tfdemonnp2" {
  metadata {
    name    = "tfdemonnp2"
    project = "terraform"
  }
  spec {
    version = "example-version"
    sharing {
        enabled = false
    }
  }
}