# Create a blueprint, sharing across project disabled
resource "rafay_blueprint" "tfdemoblueprint1" {
  metadata {
    name    = "tfdemoblueprint1"
    project = "tfdemoproject1"
    labels = {
      env  = "dev"
      name = "app"
    }
  }
  spec {
    version = "v1.1"
    base {
      name    = "default"
      version = "1.14.0"
    }
    type              = "golden"
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
      enabled = false
    }
    sharing {
      enabled = false
      projects {
        name = "demoproject"
      }
    }
    placement {
      auto_publish = false
    }
  }
}
# Blueprint for fleet values of cluster
resource "rafay_blueprint" "tfdemoblueprint2" {
  metadata {
    name    = "tfdemoblueprint2"
    project = "tfdemoproject2"
    labels = {
      env  = "dev"
      name = "app"
    }
  }
  spec {
    version = "v1.1"
    base {
      name    = "default"
      version = "1.14.0"
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
      enabled = false
    }
    sharing {
      enabled = false
      projects {
        name = "demoproject"
      }
    }
    placement {
      auto_publish = true
      fleet_values = ["value 1","value 2","value 3"]
    }
  }
}
# Blueprint with Rook-Ceph managed add-on and custom add-on
resource "rafay_blueprint" "tfdemoblueprint3" {
  metadata {
    name    = "tfdemoblueprint3"
    project = "tfdemoproject3"
    labels = {
      env  = "dev"
      name = "app"
    }
  }
  spec {
    version = "v1.1"
    base {
      name    = "default-upstream"
      version = "1.14.0"
    }
    custom_addons {
      name       = "add-on name"
      version    = "version"
    }
    default_addons {
      enable_ingress    = true
      enable_logging    = false
      enable_monitoring = true
      enable_vm         = false
      enable_rook_ceph = true
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
      enabled = false
    }
    sharing {
      enabled = false
      projects {
        name = "demoproject"
      }
    }
    placement {
      auto_publish = false
    }
  }
}