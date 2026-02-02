# Sample: managed_nodegroups reorder scenarios
# Replace provider settings and cloud_provider with real values.

resource "rafay_eks_cluster" "managed_reorder" {
  cluster {
    metadata {
      name    = "sample-managed-reorder"
      project = "default"
    }

    spec {
      cloud_provider = "your-cloud-credential"
    }
  }

  cluster_config {
    metadata {
      name   = "sample-managed-reorder"
      region = "us-west-2"
    }

    # Reorder this list (start/middle/end) to validate stability.
    managed_nodegroups {
      name          = "vishal-001"
      instance_type = "m5.large"
      subnets       = ["subnet-a", "subnet-b", "subnet-c"]
    }

    managed_nodegroups {
      name          = "vishal-002"
      instance_type = "m5.large"
      subnets       = ["subnet-b", "subnet-c", "subnet-a"]
    }

    managed_nodegroups {
      name          = "vishal-003"
      instance_type = "m5.large"
      subnets       = ["subnet-c", "subnet-a", "subnet-b"]
    }

    managed_nodegroups {
      name          = "shetty-001"
      instance_type = "m5.large"
      subnets       = ["subnet-a", "subnet-b", "subnet-c"]
    }

    managed_nodegroups {
      name          = "shetty-002"
      instance_type = "m5.large"
      subnets       = ["subnet-b", "subnet-c", "subnet-a"]
    }

    # Add shetty-004 at start/middle/end to validate
    # managed_nodegroups {
    #   name          = "shetty-004"
    #   instance_type = "m5.large"
    # }
  }
}
