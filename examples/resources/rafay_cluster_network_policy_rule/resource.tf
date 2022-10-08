resource "rafay_cluster_network_policy_rule" "tfdemocnpr1" {
  metadata {
    name    = "tfdemocnpr1"
    project = "terraform"
  }
  spec { 
    artifact { 
      type = "Yaml"
      artifact {
        paths { 
          name = "file://artifacts/cluster-network-policy.yaml" 
        } 
      } 
    }
    version = "v0"
  }
}