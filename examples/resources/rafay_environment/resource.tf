resource "rafay_environment" "eks-rds-env" {
  metadata {
    name    = var.name
    project = var.project
  }
  spec {
    template {
      name    = var.et_name
      version = var.et_version
    }
    variables {
      name       = "name"
      value_type = "text"
      value      = "dev-env-resource"
      options {
        description = "this is the name of resource created"
        sensitive   = false
        required    = true
      }
    }
  }
}