resource "rafay_environment" "eks_rds_env" {
  count = var.num_environments

  metadata {
    name    = "${var.name_prefix}-${count.index + 100}"
    project = var.project
  }
  spec {
    template {
      name    = var.et_name
      version = var.et_version
    }
    agents {
      name = var.agent
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