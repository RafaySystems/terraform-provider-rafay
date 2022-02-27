resource "rafay_addon" "tfdemoaddon1" {
  metadata {
    name    = "tfdemoaddon1"
    project = "upgrade"
    labels = {
      env  = "dev"
      name = "app"
    }
  }
  spec {
    namespace = "benny-test1"
    version   = "v1.0"
    artifact {
      type = "Yaml"
      artifact {
        paths {
          name = "yaml/qc_app_yaml_with_annotations.yaml"
        }
        repository = "release-check-ssh"
        revision   = "main"
      }
    }
    sharing {
      enabled = false
    }
  }
}