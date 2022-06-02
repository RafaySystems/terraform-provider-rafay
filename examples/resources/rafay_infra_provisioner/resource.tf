resource "rafay_infra_provisioner" "tfdemoinfraprovisioner1" {
  metadata {
    name    = "tfdemoinfraprovisioner1"
    project = "upgrade"
  }
  spec {
    config {
      backend_file_path {
        name    = "some-name"
        project = "upgrade"
      }
      backend_vars = [
        {
          key   = "string"
          type  = "Plain"
          value = "string"
        },
      ]
      env_vars = [
        {
          key   = "string"
          type  = "Plain"
          value = "string"
        },
      ]
      inputVars = [
        {
          key   = "string"
          type  = "Plain"
          value = "string"
        },
      ]
      tf_vars_file_path {
        name    = "some-name"
        project = "upgrade"
      }
      version = "1.0.0"
    }
    folder_path {
      name    = "some-name"
      project = "upgrade"
    }
    repository = "gitops"
    revision   = "string"
    type       = "Terraform"
  }
}
