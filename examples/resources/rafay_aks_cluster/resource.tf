resource "rafay_aks_cluster" "demo-terraform" {
  apiversion = "rafay.io/v1alpha1"
  kind = "Cluster"
  metadata {
    name = "demo-terraform20"
    project = "upgrade"
  }
  spec {
    type = "aks"
    blueprint = "default-aks"
    cloudprovider = "hardik-azure"
    cluster_config {
      apiversion = "rafay.io/v1alpha1"
      kind = "aksClusterConfig"
      metadata {
        name = "demo-terraform20"
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
            enable_auto_scaling = false
            max_pods = 40
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
  }
}