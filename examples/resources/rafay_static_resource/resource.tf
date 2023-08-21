resource "rafay_static_resource" "static-resource" {
  metadata {
    name    = var.name
    project = var.project
  }
  spec {
    variables {
    }
  }
}