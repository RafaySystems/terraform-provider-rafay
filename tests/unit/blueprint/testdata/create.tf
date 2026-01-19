resource "rafay_blueprint" "tftest" {
  metadata {
    name    = "test-blueprint-create"
    project = "test-project"
  }
  spec {
    version = "v1"
    default_addons {
      enable_ingress = true
    }
  }
}
