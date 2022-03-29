resource "rafay_eks_cluster" "eksclusterbasic" {
  cluster {
    kind = "Cluster"
    metadata {
      name = "test-cluster"
      project = "dev"
    }
    spec {
      type = "eks"
      blueprint = "default"
      cloud_provider = "aws-eks"
      cni_provider = "aws-cni"
      proxy_config = {}
    }
  }
  cluster_config {
    apiversion = "rafay.io/v1alpha5"
    kind = "ClusterConfig"
    metadata {
      name = "test-cluster"
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
/*
resource "rafay_eks_cluster" "eksclustersecretencryption" {
  cluster {
    kind = "Cluster"
    metadata {
      name = "rctl-chai-eks"
      project = "defaultproject"
      labels = {
        "env" : "dev"
        "type" : "eks-workloads"
      }
    }
    spec {
      type = "eks"
      blueprint = "rctl-test-blueprint"
      blueprint_version = "v1.2.x"
      cloud_provider = " yog-test-dev-aws"
      //cni_provider = ""
      //proxy_config = ""
    }
  }
  cluster_config{
    apiversion = "rafay.io/v1alpha5"
    kind = "ClusterConfig"
    metadata {
      name = "rctl-spot-eks"
      region = "us-west-1"
      //version = ""
      tags = {
         "demo": true
      }
      //annotations = ""
    }
    node_groups{
      name = "ng-1"
      instance_type = "t3.large"
      desired_capacity = 1
    }
  }
}

resource "rafay_eks_cluster" "eksclusterspot" {
  cluster {
    kind = "Cluster"
    metadata {
      name = "rctl-spot-eks"
      project = "defaultproject"
      labels = {
        "env" : "dev"
        "type" : "eks-workloads"
      }
    }
    spec {
      type = "eks"
      blueprint = "rctl-test-blueprint"
      //blueprint_version = ""
      cloud_provider = " yog-test-dev-aws"
      //cni_provider = ""
      //proxy_config = ""
    }
  }
  cluster_config{
    apiversion = "rafay.io/v1alpha5"
    kind = "ClusterConfig"
    metadata {
      name = "rctl-spot-eks"
      region = "us-west-1"
      //version = ""
      //tags = ""
      //annotations = ""
    }
    node_groups{
      name = "spot-ng-1"
      min_size = 2
      max_size = 4
      instance_distribution {
        max_price = 0.0017
        instance_types = ["t3.large"]
        on_demand_base_capacity = 0
        on_demand_percentage_above_base_capacity = 50
        spot_instance_pools = 2
      }
    }
  }
}
/*
  apiversion = "rafay.io/v1alpha1"
  kind = "Cluster"
  metadata {
    name = "demo-terraform5"
    project = "upgrade"
    labels = ""
    region = ""
    version = ""
    tags = ""
    annotations = ""
  }
  
  kubernetes_network_config{

  }
  iam {

  }
  identity_providers{

  }
  vpc {

  }
  addons {

  }
  private_cluster{}
  node_groups{}
  managed_nodegroups
  fargate_profiles
  availability_zones
  cloud_watch
  secrets_encryption
  /*
    cluster_config {
      apiversion = "rafay.io/v1alpha1"
      kind = "aksClusterConfig"
      metadata {
        name = "demo-terraform5"
      }
      spec {
        resource_group_name = "hardik-terraform"
        managed_cluster {
          apiversion = "2021-05-01"
          identity {
            type = "SystemAssigned"
          }
          location = "centralindia"
          properties {
            api_server_access_profile {
              enable_private_cluster = true
            }
            dns_prefix = "hardik-test-dns"
            kubernetes_version = "1.21.9"
            network_profile {
              network_plugin = "kubenet"
            }
            service_principle_profile {
              client_id = "3cc2fbb4-6a8b-4c42-93f1-7d5256b3d4d7"
              secret = "zTeXVo0.gV1He8b5QP_Noujdt_BaIlDKe~"
            }
          }
          type = "Microsoft.ContainerService/managedClusters"
        }
        node_pools {
          apiversion = "2021-05-01"
          name = "primary"
          properties {
            count = 2
            enable_auto_scaling = true
            max_count = 2
            max_pods = 40
            min_count = 1
            mode = "System"
            orchestrator_version = "1.21.9"
            os_type = "Linux"
            type = "VirtualMachineScaleSets"
            upgrade_settings {
              max_surge = "40%"
            }
            vm_size = "Standard_DS2_v2"
          }
          type = "Microsoft.ContainerService/managedClusters/agentPools"
        }
      }
    }
  }*/
