resource "rafay_eks_cluster" "shared" {
  cluster {
    metadata {
      name    = "${var.cluster_name}"
      project = var.rafay_project
      labels  = { shared = "true" }
    }

    spec {
      cloud_provider = "aws"
      type           = "aws-eks"
    }
  }

  cluster_config {
    metadata2 {
      name   = "${var.cluster_name}"
      region = var.aws_region
      version = var.kubernetes_version
    }

    sharing = {
      enabled = true
      projects = [
        { name = "project-a" },
        { name = "project-b" }
      ]
    }
  }
}