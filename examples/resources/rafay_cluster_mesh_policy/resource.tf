resource "rafay_cluster_mesh_policy" "tfdemocmp1" {
  metadata {
    name    = "tfdemocmp1"
    project = "terraform"
  }
  spec {
    version = "v0"
    rules {
      name = "tfdemocmr1"
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
