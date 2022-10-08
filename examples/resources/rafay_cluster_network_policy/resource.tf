resource "rafay_cluster_network_policy" "tfdemocnp2" {
  metadata {
    name    = "tfdemocnp2"
    project = "terraform"
  }
  spec {
    version = "v0"
    rules {
      name = "cluster-network-rule"
      version = "v0"
    }
  }
}