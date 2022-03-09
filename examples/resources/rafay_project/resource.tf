resource "rafay_project" "tfdemoproject1" {
  metadata {
    name        = "tfdemoproject1"
    description = "tfdemoproject1 description"
  }

  spec {
    # spec default value is fixed to 'false' for now.
    # Enable support will be added in the future.
    default = false
  }
}
