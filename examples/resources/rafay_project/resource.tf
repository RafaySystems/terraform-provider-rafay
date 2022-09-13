# Basic project example
resource "rafay_project" "tfdemoproject1" {
  metadata {
    name        = "terraform"
    description = "terraform project"
  }
  spec {
    default = false
  }
}

resource "rafay_eks_cluster" "ekscluster-basic" {
  cluster {
    kind = "Cluster"
    metadata {
      name    = "eks-cluster-1"
      project = "defaultproject"
    }
    spec {
      type           = "eks"
      blueprint      = "default"
      blueprint_version = "1.12.0"
      cloud_provider = "gopi-aws"
      cni_provider   = "aws-cni"
      proxy_config   = {}
    }
  }
  cluster_config {
    apiversion = "rafay.io/v1alpha5"
    kind       = "ClusterConfig"
    metadata {
      name    = "eks-cluster-1"
      region  = "us-west-2"
      version = "1.21"
    }
    iam {
      service_accounts {
        attach_policy = <<EOF
        {
          "Version": "2012-10-17",
          "Statement": [
            {
              "Effect": "Allow",
              "Action": "ec2:Describe*",
              "Resource": "*"
            },
            {
              "Effect": "Allow",
              "Action": "ec2:AttachVolume",
              "Resource": "*"
            },
            {
              "Effect": "Allow",
              "Action": "ec2:DetachVolume",
              "Resource": "*"
            },
            {
              "Effect": "Allow",
              "Action": ["ec2:*"],
              "Resource": ["*"]
            },
            {
              "Effect": "Allow",
              "Action": ["elasticloadbalancing:*"],
              "Resource": ["*"]
            }
          ]
        }
        EOF
      }
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
    node_groups {
      name       = "ng-1"
      ami_family = "AmazonLinux2"
      iam {
        iam_node_group_with_addon_policies {
          image_builder = true
          auto_scaler   = true
        }
      }
      instance_type    = "t3.2xlarge"
      desired_capacity = 1
      min_size         = 1
      max_size         = 2
      max_pods_per_node = 50
      version          = "1.21"
      volume_size      = 80
      volume_type      = "gp3"
      private_networking = true
    }
  }
}