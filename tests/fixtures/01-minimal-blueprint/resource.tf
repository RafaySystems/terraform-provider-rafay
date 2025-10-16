resource "rafay_eks_cluster" "ekscluster-minimal-blueprint" {
  cluster {
    kind = "Cluster"
    metadata {
      name    = var.name
      project = var.project
    }
    spec {
      type              = "eks"
      blueprint         = "minimal"
      blueprint_version = "Latest"
      cloud_provider    = var.cloud_provider
      cni_provider      = "aws-cni"
      proxy_config      = {}
    }
  }
  cluster_config {
    apiversion = "rafay.io/v1alpha5"
    kind       = "ClusterConfig"
    metadata {
      name    = var.name
      region  = "us-west-2"
      version = "1.30"
      tags = {
        env   = "dev"
        email = "akshay.gaikwad@rafay.co"
      }
    }
    addons_config {
      disable_ebs_csi_driver = false
    }

    vpc {
      cidr = "192.168.0.0/16"
      cluster_endpoints {
        private_access = true
        public_access  = false
      }
      nat {
        gateway = "Single"
      }
    }
    managed_nodegroups {
      name = "ng-1"
      # ami = "ami-07a1409f173fe796b"
      ami_family = "AmazonLinux2"
      iam {
        iam_node_group_with_addon_policies {
          image_builder = true
          auto_scaler   = true
        }
      }
      instance_type      = var.instance_type
      desired_capacity   = 2
      min_size           = 1
      max_size           = 2
      max_pods_per_node  = 50
      version            = "1.30"
      volume_size        = 80
      volume_type        = var.volume_type
      private_networking = true
      labels = {
        app       = "infra"
        dedicated = "true"
      }
    }
    managed_nodegroups {
      name       = "ng-2"
      ami_family = "AmazonLinux2"
      iam {
        iam_node_group_with_addon_policies {
          image_builder = true
          auto_scaler   = true
        }
      }
      instance_type      = var.instance_type
      desired_capacity   = 2
      min_size           = 1
      max_size           = 2
      max_pods_per_node  = 50
      version            = "1.30"
      volume_size        = 80
      volume_type        = var.volume_type
      private_networking = true
      labels = {
        app       = "infra"
        dedicated = "true"
      }
    }
  }
}