#Basic example for namespace
resource "rafay_namespace" "tfdemonamespace1" {
  metadata {
    name        = "tfdemonamespace1"
    project     = "tfdemoproject1"
  }
  spec {
    drift {
      enabled = false
    }
    placement {
      labels {
        key = "rafay.dev/clusterName"
        value = "cluster_name"
      }
    }
  }
}

#Namespace example with resource quotas & limit ranges
resource "rafay_namespace" "namespace" {
  metadata {
    name        = "tfdemonamespace2"
    project     = "tfdemoproject1"
    labels = {
      "env"  = "prod"
    }
    annotations = {
      "logging" = "enabled"
    }
  }
  spec {
    drift {
      enabled = false
    }
    limit_range {
      pod {
        max {
          cpu {
            string = "2"
          }
          memory {
            string = "2Gi"
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
            string = "2"
          }

          memory {
            string = "2Gi"
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
    resource_quotas {
      limits {
        cpu {
          string = "8"
        }
        memory {
          string = "16Gi"
        }

      }
      requests {
        cpu {
          string = "4"
        }
        memory {
          string = "8Gi"
        }
      }

    }
    placement {
      labels {
        key = "rafay.dev/clusterName"
        value = "cluster_name"
      }
    }
  }
}

resource "rafay_namespace" "namespace" {
  metadata {
    name    = "tfdemonamespace3"
    project = "dev"
    labels = {
      "env" = "prod"
    }
    annotations = {
      "logging" = "enabled"
    }
  }
  spec {
    drift {
      enabled = true
    }
    limit_range {
      pod {
        max {
          cpu {
            string = "2"
          }
          memory {
            string = "2Gi"
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
            string = "2"
          }

          memory {
            string = "2Gi"
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
    resource_quotas {
      limits {
        cpu {
          string = "8"
        }
        memory {
          string = "16Gi"
        }

      }
      requests {
        cpu {
          string = "4"
        }
        memory {
          string = "8Gi"
        }
      }

    }
    
    placement {
      labels {
        key   = "rafay.dev/clusterName"
        value = "cluster_name"
      }
    }
  }
}