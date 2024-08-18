resource "rafay_environment_template" "aws-et" {
  metadata {
    name    = "tf-cred-cluster-et"
    project = "defaultproject"
  }
  spec {
    version = "v2"
    resources {
      type = "dynamic"
      kind = "resourcetemplate"
      name = "tf-cred-rt"
      resource_options {
        version   = "v2"
      }
    }
    resources {
      type = "dynamic"
      kind = "resourcetemplate"
      name = "tf-cluster-rt"
      resource_options {
        version   = "v1"
      }
      depends_on {
        name = "tf-cred-rt"
      }
    }
  }
}