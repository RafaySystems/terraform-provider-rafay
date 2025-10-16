resource "rafay_eks_cluster" "private" {
  cluster {
    metadata {
      name    = "${var.cluster_name}"
      project = var.rafay_project
      labels  = { network = "private" }
    }

    spec {
      cloud_provider = "aws"
      type           = "aws-eks"
      cni_provider   = "aws-cni"
      proxy_config = {
        http_proxy  = ""
        https_proxy = ""
      }
    }
  }

  cluster_config {
    metadata2 {
      name   = "${var.cluster_name}"
      region = var.aws_region
      version = var.kubernetes_version
    }

    node_groups_map = {
      "private-ng" = {
        instance_type = var.node_instance_type
        desired_capacity = var.node_desired
        private_networking = true
        subnets = [
          "subnet-0a1b2c3d4e",
          "subnet-1a2b3c4d5e"
        ]
        availability_zones = ["us-west-2a", "us-west-2b"]
      }
    }
  }
}