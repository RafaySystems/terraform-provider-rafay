# Example of a custom blueprint resource.
resource "rafay_blueprint" "blueprint" {
  metadata {
    name    = "custom-blueprint"
    project = "terraform"
  }
  spec {
    version = "v1"
    base {
      name    = "default"
      version = "1.28.0"
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
    drift_webhook {
      enabled = false
    }
    placement {
      auto_publish = false
    }
  }
}
