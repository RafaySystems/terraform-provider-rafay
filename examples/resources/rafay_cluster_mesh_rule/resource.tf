resource "rafay_cluster_mesh_rule" "tfdemocmr1" {
  metadata {
    name    = "tfdemocmr1"
    project = "terraform"
  }
  spec { 
    artifact { 
      type = "Yaml"
      artifact {
        paths { 
          name = "file://artifacts/cluster-mesh-rule.yaml"
        } 
      } 
    }
    version = "v0"
    sharing {
      enabled = true
      projects {
        name = "terraformproject2"
      }
    }
  }
}
