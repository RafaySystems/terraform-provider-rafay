resource "rafay_eks_cluster" "spot_ng" {
  cluster {
    metadata {
      name    = "${var.cluster_name}"
      project = var.rafay_project
      labels  = { cost = "optimized" }
    }

    spec {
      cloud_provider = "aws"
      type           = "aws-eks"
      cni_provider   = "aws-cni"
    }
  }

  cluster_config {
    metadata2 {
      name   = "${var.cluster_name}"
      region = var.aws_region
      version = var.kubernetes_version
    }

    managed_nodegroups_map = {
      "spot-ng" = {
        instance_types = ["t3.medium", "t3.small"]
        spot = true
        desired_capacity = 2
        min_size = 1
        max_size = 4
        instances_distribution6 = {
          instance_types = ["t3.medium", "t3.small"]
          on_demand_base_capacity = 0
          on_demand_percentage_above_base_capacity = 0
          spot_instance_pools = 2
        }
        tags = { created_by = "automation-tests" }
      }
    }
  }
}