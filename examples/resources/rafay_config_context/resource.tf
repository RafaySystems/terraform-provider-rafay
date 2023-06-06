resource "rafay_config_context" "aws-rds-config" {
  metadata {
    name    = var.name
    project = var.project
  }
  spec {
    envs {
      key       = "name"
      value     = "envmgr-aws-rds"
      sensitive = false
    }
    files {
      name      = "file://variables.tf"
      sensitive = true
    }
    variables {
      name       = "my-variable"
      value_type = "text"
      value      = "my-value"
      options {
        description = "this is a dummy variable"
        sensitive   = false
        required    = true
      }
    }
  }
}