resource "rafay_eks_cluster" "eksclusterbasic" {
  cluster {
    kind = "Cluster"
    metadata {
      name = "test-cluster2"
      project = "dev"
    }
    spec {
      type = "eks"
      blueprint = "default"
      cloud_provider = "hardik-eks-role"
      cni_provider = "aws-cni"
      proxy_config = {}
    }
  }
  cluster_config {
    apiversion = "rafay.io/v1alpha5"
    kind = "ClusterConfig"
    metadata {
      name = "test-cluster2"
      region = "us-west-2"
      version = "1.21"
    }
    node_groups{
      name = "ng-57658a87"
      ami_family = "AmazonLinux2"
      iam {
        iam_node_group_with_addon_policies {
          image_builder = true
          auto_scaler = true
        }
      }
      instance_type = "t3.xlarge"
      desired_capacity = 1
      min_size = 1
      max_size = 2
      volume_size = 80
      volume_type = "gp3"
    }
    vpc {
      cidr = "192.168.0.0/16"
      cluster_endpoints {
        private_access = true
        public_access = false
      }
      nat {
        gateway = "Single"
      }
    }
  }
}

resource "rafay_eks_cluster" "eksspot" {
  cluster {
    kind = "Cluster"
    metadata {
      name = "test-spot"
      project = "dev"
    }
    spec {
      type = "eks"
      blueprint = "default"
      cloud_provider = "hardik-eks-role"
      cni_provider = "aws-cni"
      proxy_config = {}
    }
  }
  cluster_config {
    apiversion = "rafay.io/v1alpha5"
    kind = "ClusterConfig"
    metadata {
      name = "test-spot"
      region = "us-west-2"
      version = "1.21"
    }
    node_groups{
      name = "spot-ng-1"
      min_size = 0
      max_size = 4
      instances_distribution {
        max_price = 0.017
        instance_types = ["t3.xlarge"]
        on_demand_base_capacity = 0
        on_demand_percentage_above_base_capacity = 0
        spot_instance_pools = 2
      }
    }
  }
}

resource "rafay_eks_cluster" "eksmanagedcustom" {
  cluster {
    kind = "Cluster"
    metadata {
      name = "test-managed-custom2"
      project = "dev"
    }
    spec {
      type = "eks"
      blueprint = "default"
      cloud_provider = "hardik-eks-role"
      cni_provider = "aws-cni"
      proxy_config = {}
    }
  }
  cluster_config {
    apiversion = "rafay.io/v1alpha5"
    kind = "ClusterConfig"
    metadata {
      name = "test-managed-custom2"
      region = "us-west-2"
      version = "1.21"
    }
    vpc {
      subnets {
        private {
          name = "subnet-1" 
          id = "subnet-06e2a2cea8270483d"
        }
        private {
          name = "subnet-2" 
          id = "subnet-032b5640e54cee1d2"
        }
        public {
          name = "subnet-3" 
          id = "subnet-076ae102f8593bc63"
        }
        public {
          name = "subnet-4" 
          id = "subnet-05389b9c0829a4c30"
        }
      }
    }
    managed_nodegroups {
      name = "managed-ng-1"
      instance_type = "t3.large"
      desired_capacity = 1
    }
  }
}