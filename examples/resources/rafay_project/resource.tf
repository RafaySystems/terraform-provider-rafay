# Create a project with resource quota
resource "rafay_project" "tfdemoproject1" {
  metadata {
    name        = "tfdemoproject1"
    description = "tfdemoproject1 description"
  }

  spec {
    # spec default value is fixed to 'false' for now in the controller.
    # Will be allowed to enable in the future.
    default = false
    cluster_resource_quota {
      cpu_requests = "8m"
      memory_requests = "4Mi"
      cpu_limits = "6m"
      memory_limits = "8Mi"
	    config_maps = "10"
	    persistent_volume_claims = "5"
	    secrets = "4"
	    services = "20"	
	    pods = "200"
	    replication_controllers = "10"
    }
    default_cluster_namespace_quota {
      cpu_requests = "4m"
	    memory_requests = "2Mi"
	    cpu_limits = "2m"
	    memory_limits = "4Mi"
	    config_maps = "5"
	    persistent_volume_claims = "2"
	    secrets = "2"
	    services = "10"
	    pods = "20"
	    replication_controllers = "4"
    }
  }
}