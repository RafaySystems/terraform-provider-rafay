resource "rafay_config_context" "aws-rds-config" {
  metadata {
    name        = var.name
    project     = var.project
    description = "this is a test config context created from terraform"
    annotations = {
      key   = "my-ann-key"
      value = "my-ann-value"
    }
  }
  spec {
    envs {
      key       = "name-modified"
      value     = "modified-value"
      options {
        sensitive = true
      }
    }
    envs {
      key       = "name-new"
      value     = "new-value"
      options {
        required = true
      }
    }
    files {
      name      = "file://variables.tf"
      mount_path = "/local/tmp"
      options {
        description = "file with default input variables"
        override {
          type = "allowed"
        }
      }
    }
    variables {
      name       = "new-variable"
      value_type = "text"
      value      = "new-value"
      options {
        override {
          type              = "restricted"
          restricted_values = ["new-value", "modified-value"]
        }
        description = "this is a dummy variable"
        sensitive   = false
        required    = true
      }
    }
  }
}