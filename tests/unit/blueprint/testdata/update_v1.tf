resource "rafay_blueprint" "tftest" {
  metadata {
    name    = "test-blueprint-update"
    project = "test-project"
  }
  spec {
    version = "v1"
  }
}
