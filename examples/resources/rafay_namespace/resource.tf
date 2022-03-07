#rafay_namespace.tfdemonamespace1:
resource "rafay_namespace" "tfdemonamespace1" {
  metadata {
    labels = {
      "env"  = "dev"
      "name" = "app"
    }
    annotations = {
      "env"  = "dev"
      "name" = "app"
    }
    
    name    = "tfdemonamespace1"
    project = "upgrade"
  }

  spec {
    artifact {
    }

    drift {
      enabled = false
    }

    limit_range {
      container {
        default {
          cpu {
            string = "1"
          }

          memory {
            string = "1Gi"
          }
        }

        default_request {
          cpu {
            string = "1"
          }

          memory {
            string = "1Gi"
          }
        }

        max {
          cpu {
            string = "1"
          }

          memory {
            string = "1Gi"
          }
        }

        min {
          cpu {
            string = "1"
          }

          memory {
            string = "1Gi"
          }
        }

        ratio {
          cpu    = 1
          memory = 1
        }
      }

      pod {

        max {
          cpu {
            string = "1"
          }

          memory {
            string = "1Gi"
          }
        }

        min {
          cpu {
            string = "1"
          }

          memory {
            string = "1Gi"
          }
        }

        ratio {
          cpu    = 1
          memory = 1
        }
      }
    }

    placement {
       labels {
        key = "tftest"
        value = "nstest"
      }
    }


    resource_quotas {
      limits {
        cpu {
          string = "1"
        }

        memory {
          string = "1Gi"
        }
      }

      requests {
        cpu {
          string = "1"
        }

        memory {
          string = "1Gi"
        }
      }
    }
  }
  timeouts {
    create = "1m"
    delete = "1m"
    update = "1m" 
  }
}


#rafay_namespace.tfdemonamespace2:
resource "rafay_namespace" "tfdemonamespace2" {

  metadata {
    name    = "tfdemonamespace2"
    project = "upgrade"
    labels = {
      env  = "dev"
      name = "app"
    }
  }
  spec {
    placement {
      selector = "rafay.dev/clusterName=hardik-qc-mks-1"
    }
    drift {
      enabled = false
    }
    artifact {
      path {
        name = "yaml/tfns.yaml"
      }
      repository = "release-check-ssh"
      revision   = "main"

    }
  }
}
