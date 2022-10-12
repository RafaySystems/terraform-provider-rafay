resource "rafay_mesh_profile" "tfdemomeshprofile1" {
  metadata {
    name    = "tfdemomeshprofile1"
    project = "terraform"
  }
  spec {
    version = "example-version"
    sharing {
        enabled = false
    }
  }
}
