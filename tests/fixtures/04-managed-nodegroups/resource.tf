resource "rafay_eks_cluster" "managed_ng" {
  cluster {
    metadata {
      name    = "${var.cluster_name}"
      project = var.rafay_project
      labels  = { workload = "managed" }
    }

    spec {
      cloud_provider = "aws"
      type           = "aws-eks"
      blueprint      = "default"
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
      "managed-ng-1" = {
        instance_type  = var.node_instance_type
        desired_capacity = var.node_desired
        min_size = var.node_min_size
        max_size = var.node_max_size
        volume_size = 50
        labels = {
          role = "app"
        }
        tags = {
          purpose = "integration-test"
        }
      }
    }
  }
}