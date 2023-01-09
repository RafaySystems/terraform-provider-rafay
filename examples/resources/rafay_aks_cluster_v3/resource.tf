resource "rafay_aks_cluster_v3" "demo-terraform" {
  api_version = "infra.k8smgmt.io/v3"
  kind       = "Cluster"
  metadata {
    name    = "rafay-aks-v3-test"
    project = "defaultproject"
  }
  spec {
    type          = "aks"
    blueprint_config {
      name = "default-aks"
      version = "1.21.0"
    }
    cloudprovider = "azure-key-jon"
    config {
      api_version = "rafay.io/v1alpha1" # TODO: FIX THIS
      kind       = "aksClusterConfig"
      metadata {
        name = "rafay-aks-v3-test"
        project = "defaultproject"
      }
      spec {
        resource_group_name = "rafay-atlantis-rg"
        managed_cluster {
          api_version = "2022-07-01"
          identity {
            type = "SystemAssigned"
          }
          location = "centralindia"
          properties {
            api_server_access_profile {
              enable_private_cluster = true
            }
            dns_prefix         = "rafay-aks-v3-dns"
            kubernetes_version = "1.23.12"
            network_profile {
              network_plugin = "kubenet"
              load_balancer_sku = "standard"
            }
            sku {
              name = "Basic"
              tier = "Free"
            }
          }
          type = "Microsoft.ContainerService/managedClusters"
        }
        node_pools {
          api_version = "2022-07-01"
          location = "centralindia"
          name       = "primary"
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
      }
    }
  }
}
