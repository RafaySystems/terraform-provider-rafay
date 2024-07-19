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
          apiversion = "2024-01-01"
          identity {
            type = "SystemAssigned"
          }
          location = "centralindia"
          properties {
            api_server_access_profile {
              enable_private_cluster = true
            }
            dns_prefix         = "testuser-test-dns"
            enable_rbac        = true
            kubernetes_version = "1.29.0"
            network_profile {
              network_plugin = "kubenet"
            }
            power_state {
              code = "Running"
            }
            addon_profiles {
              oms_agent {
                enabled = false
                config {
                  log_analytics_workspace_resource_id = ""
                }
              }
            }
            auto_upgrade_profile {
              upgrade_channel = "rapid"
              node_os_upgrade_channel = "NodeImage"
            }
          }
          type = "Microsoft.ContainerService/managedClusters"
        }
        node_pools {
          apiversion = "2024-01-01"
          name       = "primary"
          properties {
            count                = 2
            enable_auto_scaling  = true
            max_count            = 2
            max_pods             = 40
            min_count            = 1
            mode                 = "System"
            orchestrator_version = "1.29.0"
            os_type              = "Linux"
            type                 = "VirtualMachineScaleSets"
            vm_size              = "Standard_DS2_v2"
          }
          type     = "Microsoft.ContainerService/managedClusters/agentPools"
          
        }
        maintenance_configurations {
          api_version = "2024-01-01"
          name = "aksManagedNodeOSUpgradeSchedule"
          properties {
            maintenance_window {
              duration_hours = 4
              schedule {
                daily {
                  interval_days = 1
                }
              }
              start_date = "2024-07-19"
              start_time = "11:38"
              utc_offset = "+05:30"
            }
          }
          type = "Microsoft.ContainerService/managedClusters/maintenanceConfigurations"
        }
        maintenance_configurations {
          api_version = "2024-01-01"
          name = "aksManagedAutoUpgradeSchedule"
          properties {
            maintenance_window {
              duration_hours = 4
              schedule {
                weekly {
                  day_of_week = "Tuesday"
                  interval_weeks = 1
                }
              }
              start_date = "2024-07-19"
              start_time = "11:38"
              utc_offset = "+05:30"
            }
          }
          type = "Microsoft.ContainerService/managedClusters/maintenanceConfigurations"
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
          apiversion = "2024-01-01"
          identity {
            type = "SystemAssigned"
          }
          location = "centralindia"
          properties {
            api_server_access_profile {
              enable_private_cluster = true
            }
            dns_prefix         = "testuser-test-dns"
            enable_rbac        = true
            kubernetes_version = "1.29.0"
            network_profile {
              network_plugin = "kubenet"
              network_policy = "calico"
            }
          }
          type = "Microsoft.ContainerService/managedClusters"
        }
        node_pools {
          apiversion = "2024-01-01"
          name       = "primary"
          properties {
            count                = 2
            enable_auto_scaling  = true
            max_count            = 2
            max_pods             = 40
            min_count            = 1
            mode                 = "System"
            orchestrator_version = "1.29.0"
            os_type              = "Linux"
            type                 = "VirtualMachineScaleSets"
            vm_size              = "Standard_DS2_v2"
          }
          type     = "Microsoft.ContainerService/managedClusters/agentPools"
          
        }
        node_pools {
          apiversion = "2024-01-01"
          name       = "secondary"
          properties {
            count                = 2
            enable_auto_scaling  = false
            max_pods             = 40
            mode                 = "User"
            orchestrator_version = "1.29.0"
            os_type              = "Windows"
            type                 = "VirtualMachineScaleSets"
            vm_size              = "Standard_DS2_v2"
          }
          type     = "Microsoft.ContainerService/managedClusters/agentPools"
          
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
          apiversion = "2024-01-01"
          identity {
            type = "UserAssigned"
            user_assigned_identities = {
              "/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.ManagedIdentity/userAssignedIdentities/{identityName}" : "{}"
            }
          }
          sku {
            name = "Base"
            tier = "Free"
          }
          location = "centralindia"
          properties {
            api_server_access_profile {
              enable_private_cluster             = true
              enable_private_cluster_public_fqdn = false
              private_dns_zone                   = "/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.Network/privateDnsZones/{dnsZoneName}"
            }
            dns_prefix          = "testuser-test-dns"
            enable_rbac         = true
            kubernetes_version  = "1.29.0"
            node_resource_group = "node-resource-name"
            pod_identity_profile {
              enabled                      = true
              allow_network_plugin_kubenet = true
              user_assigned_identities {
                binding_selector = "selector-name"
                identity {
                  client_id   = "CLIENT_ID"
                  object_id   = "OBJECT_ID"
                  resource_id = "resource_id = /subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.ManagedIdentity/userAssignedIdentities/{identityName}"
                }
                name      = "pod-identity-name"
                namespace = "namespace-name"
              }
              user_assigned_identity_exceptions {
                name      = "exception-name"
                namespace = "namespace-name"
                pod_labels = {
                  "key" = "value"
                }
              }
            }
            linux_profile {
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
              azure_policy {
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
              ok_total_unready_count           = "3"
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
          apiversion = "2024-01-01"
          name       = "primary"
          properties {
            count                 = 2
            enable_auto_scaling   = true
            max_count             = 2
            max_pods              = 40
            min_count             = 1
            mode                  = "System"
            orchestrator_version  = "1.29.0"
            os_type               = "Linux"
            os_disk_size_gb       = 30
            type                  = "VirtualMachineScaleSets"
            availability_zones    = [1, 2, 3]
            enable_node_public_ip = false
            vm_size               = "Standard_DS2_v2"
            vnet_subnet_id        = "/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.Network/virtualNetworks/{virtualNetworkName}/subnets/{subnetName}"
            node_labels = {
              "key" = "value"
            }
            tags = {
              "key" = "value"
            }
          }
          type     = "Microsoft.ContainerService/managedClusters/agentPools"
          
        }
        node_pools {
          apiversion = "2024-01-01"
          name       = "secondary"
          properties {
            count                 = 2
            enable_auto_scaling   = false
            max_pods              = 40
            mode                  = "User"
            orchestrator_version  = "1.29.0"
            os_type               = "Windows"
            os_disk_size_gb       = 30
            type                  = "VirtualMachineScaleSets"
            availability_zones    = [1, 2, 3]
            enable_node_public_ip = false
            vm_size               = "Standard_DS2_v2"
            vnet_subnet_id        = "/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.Network/virtualNetworks/{virtualNetworkName}/subnets/{subnetName}"
            node_labels = {
              "key" = "value"
            }
            tags = {
              "key" = "value"
            }
          }
          type     = "Microsoft.ContainerService/managedClusters/agentPools"
          
        }
      }
    }
  }
}

resource "rafay_aks_cluster" "demo-terraform-scp" {
  apiversion = "rafay.io/v1alpha1"
  kind       = "Cluster"
  metadata {
    name    = "demo-terraform-scp"
    project = "terraform"
  }
  spec {
    type          = "aks"
    blueprint     = "default-aks"
    cloudprovider = "azure-cred"
    cluster_config {
      apiversion = "rafay.io/v1alpha1"
      kind       = "aksClusterConfig"
      metadata {
        name = "demo-terraform-scp"
      }
      spec {
        resource_group_name = "testuser-terraform"
        managed_cluster {
          apiversion = "2024-01-01"
          identity {
            type = "SystemAssigned"
          }
          location = "centralindia"
          properties {
            api_server_access_profile {
              enable_private_cluster = true
            }
            dns_prefix         = "testuser-test-dns"
            enable_rbac        = true
            kubernetes_version = "1.25.6"
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
          apiversion = "2024-01-01"
          name       = "primary"
          properties {
            count                = 2
            enable_auto_scaling  = true
            max_count            = 2
            max_pods             = 40
            min_count            = 1
            mode                 = "System"
            orchestrator_version = "1.25.6"
            os_type              = "Linux"
            type                 = "VirtualMachineScaleSets"
            vm_size              = "Standard_DS2_v2"
            node_labels = {
              app       = "infra"
              dedicated = "true"
            }
            node_taints = ["app=infra:PreferNoSchedule"]
          }
          type     = "Microsoft.ContainerService/managedClusters/agentPools"
          
        }
      }
    }
    system_components_placement {
      node_selector = {
        app       = "infra"
        dedicated = "true"
      }
      tolerations {
        effect   = "PreferNoSchedule"
        key      = "app"
        operator = "Equal"
        value    = "infra"
      }

      daemonset_override {
        node_selection_enabled = false
        tolerations {
          key      = "app1dedicated"
          value    = true
          effect   = "NoSchedule"
          operator = "Equal"
        }
      }
    }
  }
}

resource "rafay_aks_cluster" "demo-terraform-authType-localAccounts-k8sRBAC" {
  apiversion = "rafay.io/v1alpha1"
  kind       = "Cluster"
  metadata {
    name    = "demo-terraform3"
    project = "terraform"
  }
  spec {
    type          = "aks"
    blueprint     = "minimal"
    cloudprovider = "aks-cred"
    cluster_config {
      apiversion = "rafay.io/v1alpha1"
      kind       = "aksClusterConfig"
      metadata {
        name = "demo-terraform3"
      }
      spec {
        resource_group_name = "aks-resourcegroup"
        managed_cluster {
          apiversion = "2024-01-01"
          identity {
            type = "SystemAssigned"
          }
          location = "centralindia"
          properties {
            api_server_access_profile {
              enable_private_cluster = true
            }
            dns_prefix         = "demo-terraform3-test-dns"
            enable_rbac        = true
            kubernetes_version = "1.29.0"
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
          apiversion = "2024-01-01"
          name       = "primary"
          properties {
            count                = 2
            enable_auto_scaling  = true
            max_count            = 2
            max_pods             = 40
            min_count            = 1
            mode                 = "System"
            orchestrator_version = "1.29.0"
            os_type              = "Linux"
            type                 = "VirtualMachineScaleSets"
            vm_size              = "Standard_B4ms"
          }
          type     = "Microsoft.ContainerService/managedClusters/agentPools"
          
        }
      }
    }
  }
}

resource "rafay_aks_cluster" "demo-terraform-authType-azureAuthentication-k8sRBAC" {
  apiversion = "rafay.io/v1alpha1"
  kind       = "Cluster"
  metadata {
    name    = "demo-terraform5"
    project = "terraform"
  }
  spec {
    type          = "aks"
    blueprint     = "minimal"
    cloudprovider = "azure-cred"
    cluster_config {
      apiversion = "rafay.io/v1alpha1"
      kind       = "aksClusterConfig"
      metadata {
        name = "demo-terraform5"
      }
      spec {
        resource_group_name = "azure-resourcegroup"
        managed_cluster {
          apiversion = "2024-01-01"
          identity {
            type = "SystemAssigned"
          }
          location = "centralindia"
          properties {
            api_server_access_profile {
              enable_private_cluster = true
            }
            dns_prefix             = "demo-terraform5-test-dns"
            enable_rbac            = true
            disable_local_accounts = true
            aad_profile {
              managed                = true
              admin_group_object_ids = ["<aad-group-object-id>"]
            }
            kubernetes_version = "1.29.0"
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
          apiversion = "2024-01-01"
          name       = "primary"
          properties {
            count                = 2
            enable_auto_scaling  = true
            max_count            = 2
            max_pods             = 40
            min_count            = 1
            mode                 = "System"
            orchestrator_version = "1.29.0"
            os_type              = "Linux"
            type                 = "VirtualMachineScaleSets"
            vm_size              = "Standard_B4ms"
          }
          type     = "Microsoft.ContainerService/managedClusters/agentPools"
          
        }
      }
    }
  }
}

resource "rafay_aks_cluster" "demo-terraform-authType-azureAuthentication-azureRBAC" {
  apiversion = "rafay.io/v1alpha1"
  kind       = "Cluster"
  metadata {
    name    = "demo-terraform6"
    project = "terraform"
  }
  spec {
    type          = "aks"
    blueprint     = "minimal"
    cloudprovider = "azure-cred"
    cluster_config {
      apiversion = "rafay.io/v1alpha1"
      kind       = "aksClusterConfig"
      metadata {
        name = "demo-terraform6"
      }
      spec {
        resource_group_name = "mayank-rg"
        managed_cluster {
          apiversion = "2024-01-01"
          identity {
            type = "SystemAssigned"
          }
          location = "centralindia"
          properties {
            api_server_access_profile {
              enable_private_cluster = true
            }
            dns_prefix             = "demo-terraform6-test-dns"
            enable_rbac            = true
            disable_local_accounts = true
            aad_profile {
              managed           = true
              enable_azure_rbac = true
            }
            kubernetes_version = "1.29.0"
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
          apiversion = "2024-01-01"
          name       = "primary"
          properties {
            count                = 2
            enable_auto_scaling  = true
            max_count            = 2
            max_pods             = 40
            min_count            = 1
            mode                 = "System"
            orchestrator_version = "1.29.0"
            os_type              = "Linux"
            type                 = "VirtualMachineScaleSets"
            vm_size              = "Standard_B4ms"
          }
          type     = "Microsoft.ContainerService/managedClusters/agentPools"
          
        }
      }
    }
  }
}

resource "rafay_aks_cluster" "demo-terraform-multiple-ACR" {
  apiversion = "rafay.io/v1alpha1"
  kind       = "Cluster"
  metadata {
    name    = "demo-terraform7"
    project = "terraform"
  }
  spec {
    type          = "aks"
    blueprint     = "default-aks"
    cloudprovider = "azure-cred"
    cluster_config {
      apiversion = "rafay.io/v1alpha1"
      kind       = "aksClusterConfig"
      metadata {
        name = "demo-terraform7"
      }
      spec {
        resource_group_name = "azure-resourcegroup"
        managed_cluster {
          apiversion = "2024-01-01"
          additional_metadata {
            acr_profile {
              registries {
                acr_name            = "<acr-name>"
                resource_group_name = "<acr-resourcegroup>"
              }
            }
          }
          identity {
            type = "UserAssigned"
            user_assigned_identities = {
              "/subscriptions/<subscription-id>/resourceGroups/<resource-group>/providers/Microsoft.ManagedIdentity/userAssignedIdentities/<identity-name>" = "{}"
            }
          }
          location = "centralindia"
          properties {
            api_server_access_profile {
              enable_private_cluster = true
            }
            dns_prefix         = "testuser-test-dns"
            enable_rbac        = true
            kubernetes_version = "1.29.0"
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
            identity_profile {
              kubelet_identity {
                resource_id = "/subscriptions/<subscription-id>/resourceGroups/<resource-group>/providers/Microsoft.ManagedIdentity/userAssignedIdentities/<identity-name>"
              }
            }
          }
          type = "Microsoft.ContainerService/managedClusters"
        }
        node_pools {
          apiversion = "2024-01-01"
          name       = "primary"
          properties {
            count                = 2
            enable_auto_scaling  = true
            max_count            = 2
            max_pods             = 40
            min_count            = 1
            mode                 = "System"
            orchestrator_version = "1.29.0"
            os_type              = "Linux"
            type                 = "VirtualMachineScaleSets"
            vm_size              = "Standard_B4ms"
          }
          type     = "Microsoft.ContainerService/managedClusters/agentPools"
          
        }
      }
    }
  }
}

resource "rafay_aks_cluster" "aks_cluster_azure_cni_overlay" {
  apiversion = "rafay.io/v1alpha1"
  kind       = "Cluster"
  metadata {
    name    = "aks_cluster_azure_cni_overlay"
    project = "defaultproject"
  }
  spec {
    type          = "aks"
    blueprint     = "minimal"
    cloudprovider = "azure-creds"
    cluster_config {
      apiversion = "rafay.io/v1alpha1"
      kind       = "aksClusterConfig"
      metadata {
        name = "aks_cluster_azure_cni_overlay"
      }
      spec {
        resource_group_name = "gautham-rg-ci"
        managed_cluster {
          apiversion = "2023-11-01"
          identity {
            type = "SystemAssigned"
          }
          location = "centralindia"
          properties {
            api_server_access_profile {
              enable_private_cluster = true
            }
            dns_prefix         = "testuser-test-dns"
            enable_rbac        = true
            kubernetes_version = "1.28.3"
            network_profile {
              network_plugin      = "azure"
              load_balancer_sku   = "standard"
              network_plugin_mode = "overlay"
              pod_cidr            = "192.168.0.0/16"
              service_cidr        = "10.0.0.0/16"
              dns_service_ip      = "10.0.0.10"
            }
            power_state {
              code = "Running"
            }
            addon_profiles {
              http_application_routing {
                enabled = false
              }
              azure_policy {
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
          }
          type = "Microsoft.ContainerService/managedClusters"
        }
        node_pools {
          apiversion = "2023-11-01"
          name       = "primary"
          properties {
            count                = 2
            enable_auto_scaling  = true
            max_count            = 2
            max_pods             = 40
            min_count            = 1
            mode                 = "System"
            orchestrator_version = "1.28.3"
            os_type              = "Linux"
            type                 = "VirtualMachineScaleSets"
            vm_size              = "Standard_DS2_v2"
          }
          type     = "Microsoft.ContainerService/managedClusters/agentPools"
          
        }
      }
    }
  }
}