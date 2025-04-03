# Basic project example
resource "rafay_project" "tfdemoproject1" {
  metadata {
    name        = "terraform"
    description = "terraform project"
  }
  spec {
    default = false
  }
}

# Project with resource quota
resource "rafay_project" "tfdemoproject2" {
  metadata {
    name        = "terraform-quota"
    description = "terraform quota project"
  }
  spec {
    default                  = false
    sync_excluded_namespaces = ["namespace1", "namespace2", "namespace3"]
    cluster_resource_quota {
      cpu_requests             = "8m"
      memory_requests          = "4Mi"
      cpu_limits               = "6m"
      gpu_requests             = "10"
      gpu_limits               = "10"
      memory_limits            = "8Mi"
      config_maps              = "10"
      persistent_volume_claims = "5"
      secrets                  = "4"
      services                 = "20"
      pods                     = "200"
      replication_controllers  = "10"
      services_load_balancers  = "3"
      services_node_ports      = "10"
      storage_requests         = "10Gi"
    }
    default_cluster_namespace_quota {
      cpu_requests             = "4m"
      memory_requests          = "2Mi"
      cpu_limits               = "2m"
      gpu_requests             = "10"
      gpu_limits               = "10"
      memory_limits            = "4Mi"
      config_maps              = "5"
      persistent_volume_claims = "2"
      secrets                  = "2"
      services                 = "10"
      pods                     = "20"
      replication_controllers  = "4"
      services_load_balancers  = "3"
      services_node_ports      = "10"
      storage_requests         = "10Gi"
    }
    drift_webhook {
      enabled = true
    }
  }
}