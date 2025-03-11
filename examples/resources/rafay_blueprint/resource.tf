# Example of a custom blueprint resource.
resource "rafay_blueprint" "blueprint" {
  metadata {
    name    = "custom-blueprint3"
    project = "defaultproject"
  }
  spec {
    version = "v0"
    base {
      name    = "default"
      version = "3.2.0"
    }
    sharing {
      enabled = true
      #projects {
      #  name = "project1"
      #}
      projects {
        name = "project10"
      }
      projects {
        name = "project2"
      }
      projects {
        name = "project3"
      }
    }
    default_addons {
      enable_ingress    = true
      enable_logging    = false
      enable_monitoring = true
      enable_vm         = false
      monitoring {
        metrics_server {
          enabled = true
          discovery {}
        }
        helm_exporter {
          enabled = true
        }
        kube_state_metrics {
          enabled = true
        }
        node_exporter {
          enabled = true
        }
        prometheus_adapter {
          enabled = true
        }
        resources {
          limits {
            memory = "200Mi"
            cpu  = "100m"
          }
        }
      }
    }
    drift {
      action  = "Deny"
      enabled = true
    }
    placement {
      auto_publish = false
    }
  }
}

