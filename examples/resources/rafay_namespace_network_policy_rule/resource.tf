resource "rafay_namespace_network_policy_rule" "tfdemonnpr1" {
  metadata {
    name    = "tfdemonnpr1"
    project = "terraform"
  }
  spec {
    artifact {
      type = "Yaml"
      artifact {
        paths {
          name = "file://artifacts/namespace-network-policy.yaml"
        }
      }
    }
    version = "v0"
    sharing {
      enabled = false
    }
  }
}
