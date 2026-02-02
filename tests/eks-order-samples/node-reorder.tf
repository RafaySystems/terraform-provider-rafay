# Sample: node_groups reorder scenarios
# Replace provider settings and cloud_provider with real values.

resource "rafay_eks_cluster" "node_reorder" {
  cluster {
    metadata {
      name    = "sample-node-reorder"
      project = "default"
    }

    spec {
      cloud_provider = "your-cloud-credential"
    }
  }

  cluster_config {
    metadata {
      name   = "sample-node-reorder"
      region = "us-west-2"
    }

    # Reorder this list (start/middle/end) to validate stability.
    node_groups {
      name          = "vishal-001"
      instance_type = "m5.large"
      subnets       = ["subnet-a", "subnet-b", "subnet-c"]
    }

    node_groups {
      name          = "vishal-002"
      instance_type = "m5.large"
      subnets       = ["subnet-b", "subnet-c", "subnet-a"]
    }

    node_groups {
      name          = "vishal-003"
      instance_type = "m5.large"
      subnets       = ["subnet-c", "subnet-a", "subnet-b"]
    }

    node_groups {
      name          = "shetty-001"
      instance_type = "m5.large"
      subnets       = ["subnet-a", "subnet-b", "subnet-c"]
    }

    node_groups {
      name          = "shetty-002"
      instance_type = "m5.large"
      subnets       = ["subnet-b", "subnet-c", "subnet-a"]
    }

    # Add shetty-004 at start/middle/end to validate
    # node_groups {
    #   name          = "shetty-004"
    #   instance_type = "m5.large"
    # }
  }
}
