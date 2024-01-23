# Example of a custom blueprint resource.
resource "rafay_blueprint" "blueprint" {
  metadata {
    name    = "custom-blueprint10"
    project = "vivek"
  }
  spec {
    version = "v0"
    base {
      name    = "default"
      version = "1.16.0"
    }
    namespace_config {
      sync_type = "managed"
      enable_sync = true
    }
    default_addons {
      enable_ingress    = true
      enable_csi_secret_store = true
      enable_logging    = false
      enable_monitoring = true
      enable_vm         = false

      csi_secret_store_config {
        enable_secret_rotation = true
        sync_secrets = true
        rotation_poll_interval = "2m"
        providers {
          aws = true
        }
      }
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
      enabled = true
    }
    placement {
      auto_publish = false
    }
  }
}
