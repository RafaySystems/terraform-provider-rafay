resource "rafay_eks_cluster" "bottlerocket_gpu" {
  cluster {
    metadata {
      name    = "${var.cluster_name}"
      project = var.rafay_project
      labels  = { os = "bottlerocket", gpu = "true" }
    }

    spec {
      cloud_provider = "aws"
      type           = "aws-eks"
      blueprint      = "gpu-optimized"
    }
  }

  cluster_config {
    metadata2 {
      name   = "${var.cluster_name}"
      region = var.aws_region
      version = var.kubernetes_version
    }

    node_groups_map = {
      "gpu-ng" = {
        instance_type = "p3.2xlarge"
        desired_capacity = 1
        min_size = 1
        max_size = 2
        bottle_rocket5 = {
          enable_admin_container = true
          settings = "kubernetes.enable=true"
        }
        instance_selector5 = {
          gpus = 1
        }
        labels = { workload = "gpu" }
      }
    }
  }
}