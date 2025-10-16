resource "rafay_eks_cluster" "ha_multi_az" {
  cluster {
    metadata {
      name    = "${var.cluster_name}"
      project = var.rafay_project
      labels  = { availability = "multi-az" }
    }

    spec {
      cloud_provider = "aws"
      type           = "aws-eks"
      blueprint      = "high-availability"
    }
  }

  cluster_config {
    metadata2 {
      name   = "${var.cluster_name}"
      region = var.aws_region
      version = var.kubernetes_version
    }

    node_groups_map = {
      "frontend-ng" = {
        instance_type = "t3.large"
        desired_capacity = 4
        min_size = 2
        max_size = 6
        availability_zones2 = ["us-west-2a","us-west-2b","us-west-2c"]
        max_pods_per_node = 110
      }
      "backend-ng" = {
        instance_type = "m5.large"
        desired_capacity = 3
        min_size = 2
        max_size = 5
        availability_zones2 = ["us-west-2a","us-west-2b"]
      }
    }
  }
}