resource "rafay_eks_cluster" "ekscluster-basic" {
  cluster {
    kind = "Cluster"
    metadata {
      name    = "gopimallela-terraform01"
      project = "defaultproject"
    }
    spec {
      type           = "eks"
      blueprint      = "default"
      blueprint_version = "Latest"
      cloud_provider = "gopi-aws"
      cni_provider   = "aws-cni"
      proxy_config   = {}
    }
  }
  cluster_config {

    kind       = "ClusterConfig"
    metadata {
      name    = "gopimallela-terraform01"
      region  = "us-west-2"
      version = "1.22"
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
    identity_mappings {
      arns {
        arn = "arn:aws:iam::679196758854:user/gopi01@rafay.co"
        group = ["dummy"]
        username = "gopi01"
      }
    }
    node_groups {
      name       = "ng-1"
      ami_family = "AmazonLinux2"
      iam {
        iam_node_group_with_addon_policies {
          image_builder = true
          auto_scaler   = true
        }
      }
      instance_type    = "t3.xlarge"
      desired_capacity = 1
      min_size         = 1
      max_size         = 2
      max_pods_per_node = 50
      version          = "1.22"
      volume_size      = 80
      volume_type      = "gp3"
      private_networking = true
    }
  }
}