resource "rafay_blueprint" "tfdemoblueprint1" {
  metadata {
    name    = "tfdemoblueprint1"
    project = "upgrade"
    labels = {
      env  = "dev"
      name = "app"
    }
  }
  spec {
    version = "v1.1"
    base {
      name    = "default"
      version = "1.11.0"
    }
    default_addons {
      enable_ingress    = true
      enable_logging    = false
      enable_monitoring = true
      enable_vm         = false
      monitoring {
        metrics_server {
          enabled = false
          discovery {}
        }
        helm_exporter {
          enabled = false
        }

        kube_state_metrics {
          enabled = false
        }

        node_exporter {
          enabled = false
        }

        prometheus_adapter {
          enabled = false
        }

        resources {
        }
      }
    }

    drift {
      action  = "Deny"
      enabled = true
    }

    sharing {
      enabled = true
      projects {
        name = "demo"
      }
    }
  }
}



resource "rafay_blueprint" "tfdemoblueprint2" {
  metadata {
    annotations = {}
    labels = {
      "env"  = "dev"
      "name" = "app"
    }
    name    = "tfdemoblueprint2"
    project = "upgrade"
  }

  spec {
    version = "v1.1"

    base {
      name    = "default"
      version = "1.11.0"
    }

    custom_addons {
      depends_on = []
      name       = "tomcat1"
      version    = "v1"
    }

    custom_addons {
      depends_on = []
      name       = "gold-pinger"
      version    = "v0"
    }

    default_addons {
      enable_ingress    = false
      enable_logging    = false
      enable_monitoring = true
      enable_vm         = false

      monitoring {
        helm_exporter {
          enabled = false
        }

        kube_state_metrics {
          enabled = false
        }

        metrics_server {
          enabled = false

          discovery {}
        }

        node_exporter {
          enabled = false
        }

        prometheus_adapter {
          enabled = true
        }

        resources {
          limits {
            memory {
              string = "2Gi"

            }
            cpu {
              string = "2"
            }
          }
        }
      }
    }

    drift {
      action  = "Deny"
      enabled = true
    }

    sharing {
      enabled = false
    }
  }
}




resource "rafay_blueprint" "tfdemoblueprint3" {
  metadata {
    name    = "tfdemoblueprint3"
    project = "upgrade"
    labels = {
      env  = "dev"
      name = "app"
    }
  }
  spec {
    version = "v1.1"
    base {
      name    = "default"
      version = "1.11.0"
    }
    
    custom_addons {
      depends_on = [
        "gold-pinger"
      ]
      name = "voyager"
      version = "v0"
    }
    custom_addons {
      depends_on = []
      name       = "gold-pinger"
      version    = "v0"
    }
    drift {
      action  = "Deny"
      enabled = true
    }

    sharing {
      enabled = false
    }
  }
}
