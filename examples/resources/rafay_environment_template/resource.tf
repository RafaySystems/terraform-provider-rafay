resource "rafay_environment_template" "aws-et" {
  metadata {
    name    = var.name
    project = var.project
  }
  spec {
    version = var.r_version
    resources {
      type = "dynamic"
      kind = "resourcetemplate"
      name = var.rt_name
      resource_options {
       version = "v1" 
      }
    }
    resources {
      type = "static"
      kind = "resource"
      name = var.sr_name
    }
    variables {
      name       = "name"
      value_type = "text"
      value      = "rds-envmgr"
      options {
        description = "this is a dummy variable"
        sensitive   = false
        required    = true
        override {
          type = "allowed"
        }
      }
    }
  }
}