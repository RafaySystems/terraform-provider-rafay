resource "rafay_eks_cluster" "eks" {
  cluster {
    kind = "Cluster"
    metadata {
      name    = "eks8"
      project = "dev"
      labels  = {
        "env" = "prod"
        "type" = "eks"
        "provider" = "cloud"
      }
    }
    spec {
      type           = "eks"
      blueprint      = "default"
      blueprint_version = "1.13.0"
      cloud_provider = "hardik-eks-role"
      cni_provider   = "aws-cni"
      //cni_params {
      //  custom_cni_crd_spec {
      //    cni_spec {}
      //  }
      //}
      proxy_config   = {}
    }
  }
  cluster_config {
    apiversion = "rafay.io/v1alpha5"
    kind       = "ClusterConfig"
    metadata {
      name    = "eks8"
      region  = "us-west-2"
      version = "1.21"
      tags = {
        "user" = "italia"
        "created-by" = "terraform"
      }
    }
    
    vpc {
      subnets {
        private {
          name = "private-01"
          id   = "subnet-04eaf5d33c8885b7f"
        }
        private {
          name = "private-02"
          id   = "subnet-0d2ab11c56fe5aa18"
        }
        public {
          name = "public-01"
          id   = "subnet-090683485d02afe81"
        }
        public {
          name = "public-02"
          id   = "subnet-063eaa2fa47340675"
        }
      }
      cluster_endpoints {
        private_access = true
        public_access  = false
      }
    }
    managed_nodegroups {
      name       = "mng-1"
      ami = "ami-013d96a3d7ea18879"
      ami_family = "AmazonLinux2"
      iam {
         iam_node_group_with_addon_policies {
           cloud_watch = true
           alb_ingress = true
           auto_scaler = true
           external_dns = true
           ebs =  true
         }
      }
      instance_type    = "t3.xlarge"
      desired_capacity = 1
      min_size         = 1
      max_size         = 2
      version          = "1.21"
      volume_size      = 80
      volume_type      = "gp3"
      volume_iops = 3000
      volume_throughput = 125
      //availability_zones  = ["us-west-2a", "us-west-2b"]
      private_networking = true
      //override_bootstrap_command = file("ami-docker-config.txt")
      subnets = ["subnet-05362ea9b9e324351", "subnet-00035e67864913317"]
      ssh {
        allow = true
        public_key_name = "hardik-ssh-1"

      }
      labels = {
        "node" = "worker"
      }
      tags = {
        "user" = "hardik"
      }
    }
    managed_nodegroups {
      name       = "mng-2"
      ami = "ami-013d96a3d7ea18879"
      ami_family = "AmazonLinux2"
      iam {
         iam_node_group_with_addon_policies {
           cloud_watch = true
           alb_ingress = true
           auto_scaler = true
           external_dns = true
           ebs =  true
         }
      }
      instance_type    = "t3.xlarge"
      desired_capacity = 1
      min_size         = 1
      max_size         = 2
      version          = "1.21"
      volume_size      = 80
      volume_type      = "gp3"
      volume_iops = 3000
      volume_throughput = 125
      //availability_zones  = ["us-west-2a", "us-west-2b"]
      private_networking = true
      //override_bootstrap_command = file("ami-docker-config.txt")
      subnets = ["subnet-05362ea9b9e324351", "subnet-00035e67864913317"]
      ssh {
        allow = true
        public_key_name = "hardik-ssh-1"
        enable_ssm = false

      }
      labels = {
        "node" = "worker"
      }
      tags = {
        "user" = "hardik"
      }
    }
    
  }
}

/*
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
*/