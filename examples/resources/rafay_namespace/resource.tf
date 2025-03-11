#Basic example for namespace
resource "rafay_namespace" "tfdemonamespace1" {
  metadata {
    name    = "tfdemonamespace1"
    project = "defaultproject"
  }
  spec {
    drift {
      enabled = true
    }
	resource_quotas {
		config_maps = "10"
		cpu_limits = "8000m"
#		memory_limits = "16384Mi"
#		cpu_requests = "4000m"
#		memory_requests = "8192Mi"
#		gpu_requests = "10"
#		gpu_limits = "10"
#		persistent_volume_claims = "2"
#		pods = "30"
#		replication_controllers = "5"
#		secrets = "10"
#		services = "10"
#		services_load_balancers = "4"
#		services_node_ports = "4"
#		storage_requests = "10Gi"
	}
  }
}
