resource "rafay_namespace_mesh_policy" "tfdemonmp1" {
  metadata {
    name    = "tfdemonmp1"
    project = "terraform"
  }
  spec {
    version = "v0"
    rules {
      name = "tfdemonmr1"
      version = "v0"
    }
    sharing {
      enabled = true
      projects {
        name = "terraformproject2"
      }
    }
  }
}
