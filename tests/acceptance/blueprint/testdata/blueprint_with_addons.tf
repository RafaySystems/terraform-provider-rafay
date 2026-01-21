

resource "rafay_addon" "tfdemoaddon3" {
  metadata {
    name    = "tfdemoaddon3"
    project = "%s"
    labels = {
      env  = "dev"
      name = "app"
    }
  }
  spec {
    namespace = "tftestnamespace"
    version   = "v0"
    artifact {
      type = "Yaml"
      artifact {
        url = ["https://raw.githubusercontent.com/kubernetes/website/main/content/en/examples/application/nginx-app.yaml"]
      }
    }
    sharing {
      enabled = false
    }
  }
}

resource "rafay_blueprint" "custom-blueprint" {
  depends_on = [rafay_addon.tfdemoaddon3]
  metadata {
    name    = "custom-blueprint"
    project = "%s"
  }
  spec {
    version = "v0"
    base {
      name    = "default"
      version = "%s"
    }
    custom_addons {
      name = "tfdemoaddon3"
      version = "v0"
    }
    default_addons {
      enable_ingress    = true
      enable_monitoring = true
      enable_vm         = false
      monitoring {
        metrics_server {
          enabled = true
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
            cpu = "100m"
            memory= "200Mi"
          }
        }
      }
    }
    drift {
      action  = "Deny"
      enabled = true
    }
  }
}