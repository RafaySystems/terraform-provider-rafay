resource "rafay_eks_cluster" "basic" {
  cluster {
    metadata {
      name    = var.cluster_name
      project = var.rafay_project
      labels  = {
        environment = "test"
        owner       = "qa"
      }
    }

    spec {
      cloud_provider = "aws"
      type           = "aws-eks"
      cni_provider   = "aws-cni"
    }
  }

  cluster_config {
    metadata2 {
      name   = var.cluster_name
      region = var.aws_region
      version = var.kubernetes_version
      tags = {
        environment = "test"
      }
    }
  }
}