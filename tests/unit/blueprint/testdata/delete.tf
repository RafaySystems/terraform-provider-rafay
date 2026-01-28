resource "rafay_blueprint" "tftest" {
  metadata {
    name    = "test-blueprint-delete"
    project = "test-project"
  }
  spec {
    version = "v1"
  }
}
