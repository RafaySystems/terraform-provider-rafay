resource "rafay_environment_template" "aws-et" {
  metadata {
    name    = "aws-et"
    project = "terraform"
  }
  spec {
    version = "v1"
    resources {
      type = "dynamic"
      kind = "resourcetemplate"
      name = "aws-elasticache"
    }
    resources {
      type = "static"
      kind = "resource"
      name = "static-resource"
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