resource "rafay_project" "tfdemoproject1" {
  metadata {
    name        = "tfdemoproject1"
    description = "tfdemoproject1 description"
  }
  spec {
    default = false
  }
}