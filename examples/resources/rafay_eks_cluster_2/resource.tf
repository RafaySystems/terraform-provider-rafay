resource "rafay_eks_cluster" "ekscluster_with_node_repair" {
  cluster {
    kind = "Cluster"
    metadata {
      name    = "gopi-eks-node-repair-2"
      project = "defaultproject"
    }
    spec {
      type              = "eks"
      blueprint         = "default"
      blueprint_version = "Latest"
      cloud_provider    = "gopi-eks"
      cni_provider      = "aws-cni"
      proxy_config      = {}
    }
  }
  cluster_config {
    apiversion = "rafay.io/v1alpha5"
    kind       = "ClusterConfig"
    metadata {
      name    = "gopi-eks-node-repair-2"
      region  = "us-west-2"
      version = "1.32"
      tags = {
        env = "dev"
        email = "gopi.mallela@rafay.co"
      }
    }
    managed_nodegroups_map = {
      "ng-with-repair" = {
        ami_family         = "AmazonLinux2"
        instance_type      = "t3.large"
        desired_capacity   = 2
        min_size           = 1
        max_size           = 2
        version            = "1.32"
        volume_size        = 80
        volume_type        = "gp3"
        private_networking = true

        node_repair_config = {
          enabled                                 = true
          max_unhealthy_node_threshold_percentage = 40
          max_parallel_nodes_repaired_percentage  = 25

          node_repair_config_overrides = [
            {
              node_monitoring_condition = "NodeNotReady"
              node_unhealthy_reason     = "KubeletNotReady"
              min_repair_wait_time_mins = 20
              repair_action            = "Replace"
            },
            {
              node_monitoring_condition = "NodeNotReady"
              node_unhealthy_reason     = "NetworkUnavailable"
              min_repair_wait_time_mins = 10
              repair_action            = "Replace"
            },
            {
              node_monitoring_condition = "NodeNotReady"
              node_unhealthy_reason     = "MemoryPressure"
              min_repair_wait_time_mins = 10
              repair_action            = "Replace"
            }
          ]
        }
      }
    }

    zonal_shift_config {
      enabled = true
    }

    auto_zonal_shift_config {
      enabled          = true
      outcome_alarms   = ["arn:aws:cloudwatch:us-west-2:679196758854:alarm:harish-fmac-controller-database_connection"]
      blocking_alarms  = ["arn:aws:cloudwatch:us-west-2:679196758854:alarm:harish-fmac-controller-database_connection"]
      # blocked_windows  = ["Sun:02:00-Sun:06:00"]
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
  }
}
