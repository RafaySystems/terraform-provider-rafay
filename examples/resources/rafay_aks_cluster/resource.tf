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
          sku {
            name = "Basic"
            tier = "Free"
          }
          location = "centralindia"
          properties {
            api_server_access_profile {
              enable_private_cluster = true
              enable_private_cluster_public_fqdn = false
              private_dns_zone  = "/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.Network/privateDnsZones/{dnsZoneName}"
            }
            dns_prefix          = "testuser-test-dns"
            kubernetes_version  = "1.21.9"
            node_resource_group = "node-resource-name"
            linux_profile  {
              admin_username = "adminuser"
              ssh {
                public_keys {
                  key_data = "ssh_public_key"
                }
              }
            }
            network_profile {
              network_plugin = "kubenet"
              network_policy = "calico"
              outbound_type  = "loadBalancer"
            }
            aad_profile {
              managed                = true
              admin_group_object_ids = ["admin_group_id"]
              enable_azure_rbac      = false
            }
            addon_profiles {
              http_application_routing {
                enabled = false
              }
              azure_policy  {
                enabled = false
              }
              oms_agent {
                enabled = true
                config {
                  log_analytics_workspace_resource_id = "/subscriptions/{subscriptionId}/resourcegroups/{resourceGroupName}/providers/microsoft.operationalinsights/workspaces/{loganalyticsWorkspaceName}"
                }
              }
              azure_keyvault_secrets_provider {
                enabled = true
                config {
                  enable_secret_rotation = "true"
                  rotation_poll_interval = "2m"
                }
              }
            }
            auto_scaler_profile {
              balance_similar_node_groups      = "false"
              expander                         = "random"
              max_graceful_termination_sec     = "600"
              max_node_provision_time          = "15m"
              ok_total_unready_count           =  "3"
              max_total_unready_percentage     = "45"
              new_pod_scale_up_delay           = "10s"
              scale_down_delay_after_add       = "10m"
              scale_down_delay_after_delete    = "60s"
              scale_down_delay_after_failure   = "3m"
              scan_interval                    = "10s"
              scale_down_unneeded_time         = "10m"
              scale_down_unready_time          = "20m"
              scale_down_utilization_threshold = "0.5"
              max_empty_bulk_delete            = "10"
              skip_nodes_with_local_storage    = "true"
              skip_nodes_with_system_pods      = "true"
            }
          }
          type = "Microsoft.ContainerService/managedClusters"
           tags = {
            "key" = "value"
          }
          additional_metadata {
            oms_workspace_location = "centralindia"
          }
        }
        node_pools {
          apiversion = "2021-05-01"
          name       = "primary"
          properties {
            count                 = 2
            enable_auto_scaling   = true
            max_count             = 2
            max_pods              = 40
            min_count             = 1
            mode                  = "System"
            orchestrator_version  = "1.21.9"
            os_type               = "Linux"
            os_disk_size_gb       = 30
            type                  = "VirtualMachineScaleSets"
            availability_zones    = [1, 2, 3]
            enable_node_public_ip = false
            vm_size = "Standard_DS2_v2"
            vnet_subnet_id = "/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.Network/virtualNetworks/{virtualNetworkName}/subnets/{subnetName}"
            node_labels = {
              "key" = "value"
            }
            tags = {
              "key" = "value"
            }
          }
          type = "Microsoft.ContainerService/managedClusters/agentPools"
          location = "centralindia"
        }
        node_pools {
          apiversion = "2021-05-01"
          name       = "secondary"
          properties {
            count                 = 2
            enable_auto_scaling   = false
            max_pods              = 40
            mode                  = "User"
            orchestrator_version  = "1.21.9"
            os_type               = "Windows"
            os_disk_size_gb       = 30
            type                  = "VirtualMachineScaleSets"
            availability_zones    = [1, 2, 3]
            enable_node_public_ip = false
            vm_size = "Standard_DS2_v2"
            vnet_subnet_id = "/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.Network/virtualNetworks/{virtualNetworkName}/subnets/{subnetName}"
            node_labels = {
              "key" = "value"
            }
            tags = {
              "key" = "value"
            }
          }
          type = "Microsoft.ContainerService/managedClusters/agentPools"
          location = "centralindia"
        }
      }
    }
  }
}
