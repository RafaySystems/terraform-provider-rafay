resource "rafay_cluster_network_policy" "tfdemocnp2" {
  metadata {
    name    = "tfdemocnp2"
    project = "terraform"
  }
  spec {
    version = "example-version"
    sharing {
        enabled = false
    }
  }
}