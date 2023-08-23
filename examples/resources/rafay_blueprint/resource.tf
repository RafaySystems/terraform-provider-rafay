# Example of a custom blueprint resource.
resource "rafay_blueprint" "blueprint" {
  metadata {
    name    = "custom-blueprint"
    project = "terraform"
  }
  spec {
    version = "v0"
    base {
      name    = "default"
      version = "1.16.0"
    }
    default_addons {
      enable_ingress    = true
      enable_logging    = false
      enable_monitoring = true
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

# Blueprint for fleet values of cluster
resource "rafay_blueprint" "blueprint" {
  metadata {
    name    = "custom-blueprint-advanced"
    project = "terraform"
  }
  spec {
    version = "v0"
    base {
      name    = "default"
      version = "1.16.0"
    }
    default_addons {
      enable_ingress    = true
      enable_logging    = false
      enable_monitoring = true
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
            memory ="200Mi"
            cpu = "100m"
          }
        }
      }
    }
    drift {
      action  = "Deny"
      enabled = true
    }
    placement {
      auto_publish = true
      fleet_values = ["value 1","value 2","value 3"]
    }
  }
}
# Blueprint with Rook-Ceph managed add-on and custom add-on
resource "rafay_blueprint" "blueprint" {
  metadata {
    name    = "custom-blueprint-advanced2"
    project = "terraform"
  }
  spec {
    version = "v0"
    base {
      name    = "default"
      version = "1.16.0"
    }
    custom_addons {
      name = "addon1"
      version = "v0"
    }
    custom_addons {
      depends_on = ["addon1"]
      name = "addon2"
      version = "v0"
    }
    default_addons {
      enable_ingress    = true
      enable_logging    = false
      enable_monitoring = true
      enable_rook_ceph = true
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
    sharing {
      enabled = true
      projects {
        name = "terraform"
      }
    }
  
    opa_policy {
      opa_policy {
	      name = "policy_name"
	      version = "policy_version"
      }
      profile {
	      name = "default"
          version = "latest"
      }
    }
  }
}
# Example of a custom golden base blueprint resource.
resource "rafay_blueprint" "golden_base_blueprint" {
  metadata {
    name    = "custom-golden-blueprint"
    project = "terraform"
  }
  spec {
    version = "v0"
    base {
      name    = "default"
      version = "1.16.0"
    }
    type = "golden"
    default_addons {
      enable_ingress    = true
      enable_logging    = false
      enable_monitoring = true
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
# Example of a custom blueprint resource with cost profile
resource "rafay_blueprint" "cost-blueprint" {
  metadata {
    name    = "custom-cost-blueprint"
    project = "terraform"
  }
  spec {
    version = "v0"
    base {
      name    = "default"
      version = "1.19.0"
    }
    default_addons {
      enable_ingress    = true
      enable_logging    = false
      enable_monitoring = true
    }
    drift {
      action  = "Deny"
      enabled = true
    }

    cost_profile {
        name = "tfdemocostprofile1"
        version = "v0"
    }

  }
}
