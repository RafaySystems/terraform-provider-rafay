resource "rafay_static_resource" "static-resource" {
  metadata {
    name    = var.name
    project = var.project
  }
  spec {
    variables {
      name = "inp1"
      value_type = "text"
      value = "inp1-value"
    }
  }
}