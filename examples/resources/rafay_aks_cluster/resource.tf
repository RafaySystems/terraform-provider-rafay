resource "rafay_aks_cluster" "demo-terraform" {
  apiversion = "rafay.io/v1alpha1"
  kind       = "Cluster"
  metadata {
    name    = "demo-terraform"
    project = "terraform"
  }
  spec {
    type          = "aks"
    blueprint     = "default-aks"
    cloudprovider = "testuser-azure"
    cluster_config {
      apiversion = "rafay.io/v1alpha1"
      kind       = "aksClusterConfig"
      metadata {
        name = "demo-terraform"
      }
      spec {
        resource_group_name = "testuser-terraform"
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
            dns_prefix         = "testuser-test-dns"
            kubernetes_version = "1.21.9"
            network_profile {
              network_plugin = "kubenet"
            }
            addon_profiles {
              oms_agent {
                enabled = false
                config {
                  log_analytics_workspace_resource_id = ""
                }
              }
            }
          }
          type = "Microsoft.ContainerService/managedClusters"
        }
        node_pools {
          apiversion = "2021-05-01"
          name       = "primary"
          properties {
            count                = 2
            enable_auto_scaling  = true
            max_count            = 2
            max_pods             = 40
            min_count            = 1
            mode                 = "System"
            orchestrator_version = "1.21.9"
            os_type              = "Linux"
            type                 = "VirtualMachineScaleSets"
            vm_size              = "Standard_DS2_v2"
          }
          type = "Microsoft.ContainerService/managedClusters/agentPools"
          location = "centralindia"
        }
      }
    }
  }
}


resource "rafay_aks_cluster" "demo-terraform1" {
  apiversion = "rafay.io/v1alpha1"
  kind       = "Cluster"
  metadata {
    name    = "demo-terraform1"
    project = "terraform"
  }
  spec {
    type          = "aks"
    blueprint     = "default-aks"
    cloudprovider = "testuser-azure"
    cluster_config {
      apiversion = "rafay.io/v1alpha1"
      kind       = "aksClusterConfig"
      metadata {
        name = "demo-terraform1"
      }
      spec {
        resource_group_name = "testuser-terraform"
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
            dns_prefix         = "testuser-test-dns"
            kubernetes_version = "1.21.9"
            network_profile {
              network_plugin = "kubenet"
              network_policy = "calico"
            }
          }
          type = "Microsoft.ContainerService/managedClusters"
        }
        node_pools {
          apiversion = "2021-05-01"
          name       = "primary"
          properties {
            count                = 2
            enable_auto_scaling  = true
            max_count            = 2
            max_pods             = 40
            min_count            = 1
            mode                 = "System"
            orchestrator_version = "1.21.9"
            os_type              = "Linux"
            type                 = "VirtualMachineScaleSets"
            vm_size = "Standard_DS2_v2"
          }
          type = "Microsoft.ContainerService/managedClusters/agentPools"
          location = "centralindia"
        }
        node_pools {
          apiversion = "2021-05-01"
          name       = "secondary"
          properties {
            count                = 2
            enable_auto_scaling  = false
            max_pods             = 40
            mode                 = "User"
            orchestrator_version = "1.21.9"
            os_type              = "Windows"
            type                 = "VirtualMachineScaleSets"
            vm_size = "Standard_DS2_v2"
          }
          type = "Microsoft.ContainerService/managedClusters/agentPools"
          location = "centralindia"
        }
      }
    }
  }
}

resource "rafay_aks_cluster" "demo-terraform-existing-vnet" {
  apiversion = "rafay.io/v1alpha1"
  kind       = "Cluster"
  metadata {
    name    = "demo-terraform2"
    project = "terraform"
  }
  spec {
    type          = "aks"
    blueprint     = "default-aks"
    cloudprovider = "testuser-azure"
    cluster_config {
      apiversion = "rafay.io/v1alpha1"
      kind       = "aksClusterConfig"
      metadata {
        name = "demo-terraform1"
      }
      spec {
        resource_group_name = "testuser-terraform"
        managed_cluster {
          apiversion = "2021-05-01"
          identity {
            type = "UserAssigned"
             user_assigned_identities = {
                "/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.ManagedIdentity/userAssignedIdentities/{identityName}": "{}"
                }
          }
          location = "centralindia"
          properties {
            api_server_access_profile {
              enable_private_cluster = true
              private_dns_zone  = "/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.Network/privateDnsZones/{dnsZoneName}"
            }
            dns_prefix         = "testuser-test-dns"
            kubernetes_version = "1.21.9"
            network_profile {
              network_plugin = "kubenet"
              network_policy = "calico"
            }
          }
          type = "Microsoft.ContainerService/managedClusters"
        }
        node_pools {
          apiversion = "2021-05-01"
          name       = "primary"
          properties {
            count                = 2
            enable_auto_scaling  = true
            max_count            = 2
            max_pods             = 40
            min_count            = 1
            mode                 = "System"
            orchestrator_version = "1.21.9"
            os_type              = "Linux"
            type                 = "VirtualMachineScaleSets"
            vm_size = "Standard_DS2_v2"
            vnet_subnet_id = "/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.Network/virtualNetworks/{virtualNetworkName}/subnets/{subnetName}"
          }
          type = "Microsoft.ContainerService/managedClusters/agentPools"
          location = "centralindia"
        }
        node_pools {
          apiversion = "2021-05-01"
          name       = "secondary"
          properties {
            count                = 2
            enable_auto_scaling  = false
            max_pods             = 40
            mode                 = "User"
            orchestrator_version = "1.21.9"
            os_type              = "Windows"
            type                 = "VirtualMachineScaleSets"
            vm_size = "Standard_DS2_v2"
            vnet_subnet_id = "/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.Network/virtualNetworks/{virtualNetworkName}/subnets/{subnetName}"
          }
          type = "Microsoft.ContainerService/managedClusters/agentPools"
          location = "centralindia"
        }
      }
    }
  }
}