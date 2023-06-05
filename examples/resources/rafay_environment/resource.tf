resource "rafay_environment" "eks-rds-env" {
  metadata {
    name    = "eks-rds-env"
    project = "terraform"
  }
  spec {
    template {
      name    = "aws-et"
      version = "v1"
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