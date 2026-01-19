resource "rafay_blueprint" "tftest" {
  metadata {
    name    = "test-blueprint-read"
    project = "test-project"
  }
  spec {
    version = "v1"
  }
}
