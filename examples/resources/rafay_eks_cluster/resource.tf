resource "rafay_eks_cluster" "ekscluster-basic" {
  cluster {
    kind = "Cluster"
    metadata {
      name    = "eks-cluster-1"
      project = "terraform"
    }
    spec {
      type           = "eks"
      blueprint      = "default"
      blueprint_version = "1.12.0"
      cloud_provider = "aws-eks-role"
      cni_provider   = "regression-eks-role"
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
      instance_type    = "m5.xlarge"
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


resource "rafay_eks_cluster" "ekscluste-advanced" {
  cluster {
    kind = "Cluster"
    metadata {
      name    = "eks-cluster-2"
      project = "terraform"
    }
    spec {
      type           = "eks"
      blueprint      = "default"
      blueprint_version = "1.12.0"
      cloud_provider = "eks-role"
      cni_provider   = "aws-cni"
      proxy_config   = {}
    }
  }
  cluster_config {
    apiversion = "rafay.io/v1alpha5"
    kind       = "ClusterConfig"
    metadata {
      name    = "eks-cluster-2"
      region  = "us-west-2"
      version = "1.21"
    }
    vpc {
      subnets {
        private {
          name = "private-01"
          id   = "private-subnet-id-0"
        }
        private {
          name = "private-02"
          id   = "private-subnet-id-0"
        }
        public {
          name = "public-01"
          id   = "public-subnet-id-0"
        }
        public {
          name = "public-02"
          id   = "public-subnet-id-0"
        }
      }
      cluster_endpoints {
        private_access = true
        public_access  = false
      }
    }
    managed_nodegroups {
      name       = "managed-ng-1"
      ami_family = "AmazonLinux2"
      iam {
        instance_profile_arn = "arn:aws:iam::<AWS_ACCOUNT_ID>:instance-profile/role_name"
        instance_role_arn = "arn:aws:iam::<AWS_ACCOUNT_ID>:role/role_name"
      }
      instance_type    = "m5.xlarge"
      desired_capacity = 1
      min_size         = 1
      max_size         = 2
      max_pods_per_node = 50
      security_groups {
        attach_ids = ["sg-id-1", "sg-id-2"]
      }
      subnets = ["subnet-id-1", "subnet-id-2"]
      version          = "1.21"
      volume_size      = 80
      volume_type      = "gp3"
      volume_iops      = 3000
      volume_throughput = 125
      private_networking = true
    }
  }
}

resource "rafay_eks_cluster" "ekscluster-custom-cni" {
  cluster {
    kind = "Cluster"
    metadata {
      name    = "eks-cluster-3"
      project = "terraform"
    }
    spec {
      type           = "eks"
      blueprint      = "default"
      blueprint_version = "1.12.0"
      cloud_provider = "eks-role"
      cni_provider   = "aws-cni"
      cni_params {
        custom_cni_crd_spec {
          name = "us-west-2a"
          cni_spec {
            security_groups = ["sg-xxxxxx", "sg-yyyyyy"]
            subnet = "subnet-zzz"
          }
          cni_spec {
            security_groups = ["sg-cccccc", "sg-dddddd"]
            subnet = "subnet-kkk"
          }
        }
        custom_cni_crd_spec {
          name = "us-west-2b"
          cni_spec {
            security_groups = ["sg-aaaaaa", "sg-xxxxxx"]
            subnet = "subnet-qqq"
          }
          cni_spec {
            security_groups = ["sg-cccccc", "sg-dddddd"]
            subnet = "subnet-www"
          }
        }
      }
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
      instance_type    = "m5.xlarge"
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
