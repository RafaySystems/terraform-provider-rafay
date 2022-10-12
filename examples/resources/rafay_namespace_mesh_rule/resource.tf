resource "rafay_namespace_mesh_rule" "tfdemonmr1" {
  metadata {
    name    = "tfdemonmr1"
    project = "terraform"
  }
  spec { 
    artifact { 
      type = "Yaml"
      artifact {
        paths { 
          name = "file://artifacts/namespace-mesh-rule.yaml"
        } 
      } 
    }
    version = "v0"
    sharing {
      enabled = false
    }
  }
}
