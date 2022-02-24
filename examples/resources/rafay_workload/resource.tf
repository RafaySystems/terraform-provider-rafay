resource "rafay_workload" "workload" {
  metadata {
    name    = "benny-workload1"
    project = "upgrade"
    labels = {
      env  = "dev"
      name = "app"
    }
  }
  spec {
    namespace = "benny-test1"
    placement {
      selector = "rafay.dev/clusterName=hardik-qc-mks-3"
    }
    drift {
      action  = "Deny"
      enabled = true
    }
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
  }
}
