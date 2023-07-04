resource "rafay_config_context" "aws-rds-config" {
  metadata {
    name    = var.name
    project = var.project
    description = "this is a test config context created from terraform"
    annotations = {
      key = "my-ann-key"
      value = "my-ann-value"
    }
  }
  spec {
    envs {
      key       = "name2"
      value     = "my-value"
      sensitive = true
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
        override {
          type = "allowed"
        }
        description = "this is a dummy variable"
        sensitive   = false
        required    = true
      }
    }
  }
}