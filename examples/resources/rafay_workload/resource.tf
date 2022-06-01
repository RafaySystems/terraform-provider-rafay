resource "rafay_workload" "tfdemoworkload1" {
  metadata {
    name    = "tfdemoworkload1"
    project = "bharath"
    labels = {
      env  = "dev"
      name = "app"
    }
  }
  spec {
    namespace = "tfdemonamespace1"
    placement {
      selector = "rafay.dev/clusterName=kuber-test"
    }
    drift {
      action  = "Deny"
      enabled = true
    }
    artifact {
      type = "Yaml"
      artifact {
        paths {
          name = "sport/test.yaml"
        }
        repository = "tfdemorepository1"
        revision   = "main"
      }
    }
  }
}
