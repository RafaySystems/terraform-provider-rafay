resource "rafay_aks_cluster_v3" "demo-terraform" {
  metadata {
    name    = "aks-v3-tf-2401202303"
    project = "defaultproject"
  }
  spec {
    type          = "aks"
    blueprint_config {
      name = "default-aks"
      version = "1.21.0"
    }
    cloud_credentials = "pj-azure"
    config {
      kind       = "aksClusterConfig"
      metadata {
        name = "aks-v3-tf-2401202303"
        project = "defaultproject"
      }
      spec {
        resource_group_name = "rafay-atlantis-rg"
        managed_cluster {
          api_version = "2022-07-01"
          sku {
            name = "Basic"
            tier = "Free"
          }
          identity {
            type = "SystemAssigned"
          }
          location = "centralindia"
          properties {
            api_server_access_profile {
              enable_private_cluster = true
            }
            dns_prefix         = "aks-v3-tf-2401202303-dns"
            kubernetes_version = "1.23.12"
            network_profile {
              network_plugin = "kubenet"
              load_balancer_sku = "standard"
            }
          }
          type = "Microsoft.ContainerService/managedClusters"
        }
        node_pools {
          api_version = "2022-07-01"
          name       = "primary"
          location = "centralindia"
          properties {
            count                = 1
            enable_auto_scaling  = true
            max_count            = 1
            max_pods             = 40
            min_count            = 1
            mode                 = "System"
            orchestrator_version = "1.23.12"
            os_type              = "Linux"
            type                 = "VirtualMachineScaleSets"
            vm_size              = "Standard_DS2_v2"
          }
          type = "Microsoft.ContainerService/managedClusters/agentPools"
        }
        
        #node_pools {
        #  api_version = "2022-07-01"
        #  name       = "agentpool2"
        #  location = "centralindia"
        #  properties {
        #    count                = 1
        #    enable_auto_scaling  = true
        #    max_count            = 1
        #    max_pods             = 40
        #    min_count            = 1
        #    mode                 = "System"
        #    orchestrator_version = "1.23.12"
        #    os_type              = "Linux"
        #    type                 = "VirtualMachineScaleSets"
        #    vm_size              = "Standard_B4ms"
        #  }
        #  type = "Microsoft.ContainerService/managedClusters/agentPools"
        #}

      }
    }
  }
}
