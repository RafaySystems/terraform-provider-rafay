#Basic example for namespace
resource "rafay_namespace" "tfdemonamespace1" {
  metadata {
    name    = "tfdemonamespace1"
    project = "terraform"
  }
  spec {
    drift {
      enabled = false
    }
    placement {
      labels {
        key   = "rafay.dev/clusterName"
        value = "cluster_name"
      }
    }
  }
}

#Namespace example with resource quotas & limit ranges
resource "rafay_namespace" "namespace" {
  metadata {
    name    = "cloudops"
    project = "terraform"
    labels = {
      "env" = "prod"
    }
    annotations = {
      "logging" = "enabled"
    }
  }
  spec {
    drift {
      enabled = false
    }
    placement {
      labels {
        key   = "rafay.dev/clusterName"
        value = "cluster_name"
      }
    }
    limit_range {
      pod {
        max {
          cpu  = "2"
          memory = "2Gi"
        }
        min {
          cpu  = "1"
          memory = "1Gi"
        }
        ratio {
          cpu    = 1
          memory = 1
        }
      }
      container {
        default {
          cpu = "1"   
          memory = "1Gi"
        }
        default_request {
          cpu = "1"
          memory = "1Gi"
        }

        max {
          cpu = "2"
          memory = "2Gi"
        }
        min {
          cpu = "1"   
          memory = "2Gi"
        }
        ratio {
          cpu    = 1
          memory = 1
        }
      }
    }
    resource_quotas {
      config_maps = "10"
      cpu_limits = "8"
      memory_limits = "16Gi"
      cpu_requests = "4"
      memory_requests = "8Gi"
      persistent_volume_claims = "2"
      pods = "30"
      replication_controllers = "5"
      secrets = "10"
      services = "10"
      services_load_balancers = "3"
      services_node_ports = "10"
      storage_requests = "10Gi"
      }
    }
}