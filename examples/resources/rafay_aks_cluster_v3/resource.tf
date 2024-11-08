resource "rafay_aks_cluster_v3" "demo-terraform" {
  metadata {
    name    = "aks-v3-tf-1"
    project = "defaultproject"
  }
  spec {
    type = "aks"
    blueprint_config {
      name = "default-aks"
    }
    cloud_credentials = "aks-cred"
    config {
      kind = "aksClusterConfig"
      metadata {
        name = "aks-v3-tf-1"
      }
      spec {
        resource_group_name = "rafay-resource"
        managed_cluster {
          api_version = "2024-01-01"
          sku {
            name = "Base"
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
            kubernetes_version = "1.29.0"
            network_profile {
              network_plugin    = "kubenet"
              load_balancer_sku = "standard"
            }
            power_state {
              code = "Running"
            }
            addon_profiles {
              http_application_routing {
                enabled = true
              }
              azure_policy {
                enabled = true
              }
              azure_keyvault_secrets_provider {
                enabled = true
                config {
                  enable_secret_rotation = false
                  rotation_poll_interval = "2m"
                }
              }
            }
            auto_upgrade_profile {
              upgrade_channel         = "rapid"
              node_os_upgrade_channel = "NodeImage"
            }
          }
          type = "Microsoft.ContainerService/managedClusters"
        }
        node_pools {
          api_version = "2024-01-01"
          name        = "primary"
          properties {
            count                = 1
            enable_auto_scaling  = true
            max_count            = 1
            max_pods             = 40
            min_count            = 1
            mode                 = "System"
            orchestrator_version = "1.29.0"
            os_type              = "Linux"
            type                 = "VirtualMachineScaleSets"
            vm_size              = "Standard_DS2_v2"
          }
          type = "Microsoft.ContainerService/managedClusters/agentPools"
        }
        maintenance_configurations {
          name        = "aksManagedAutoUpgradeSchedule"
          api_version = "2024-01-01"
          properties {
            maintenance_window {
              duration_hours = 4
              schedule {
                weekly {
                  interval_weeks = 1
                  day_of_week    = "Friday"
                }
              }
              start_date = "2024-07-19"
              start_time = "11:35"
              utc_offset = "+05:30"
            }
          }
          type = "Microsoft.ContainerService/managedClusters/maintenanceConfigurations"
        }
        maintenance_configurations {
          name        = "aksManagedNodeOSUpgradeSchedule"
          api_version = "2024-01-01"
          properties {
            maintenance_window {
              duration_hours = 4
              schedule {
                weekly {
                  interval_weeks = 1
                  day_of_week    = "Friday"
                }
              }
              start_date = "2024-07-19"
              start_time = "11:35"
              utc_offset = "+05:30"
            }
          }
          type = "Microsoft.ContainerService/managedClusters/maintenanceConfigurations"
        }
      }
    }
  }
}

resource "rafay_aks_cluster_v3" "demo-terraform2" {
  metadata {
    name    = "aks-v3-tf-2"
    project = "defaultproject"
  }
  spec {
    type = "aks"
    blueprint_config {
      name = "default-aks"
    }
    cloud_credentials = "aks-cred"
    config {
      kind = "aksClusterConfig"
      metadata {
        name = "aks-v3-tf-2"
      }
      spec {
        resource_group_name = "rafay-resource"
        managed_cluster {
          api_version = "2024-01-01"

          additional_metadata {
            acr_profile {
              registries {
                acr_name            = "<acr-name>"
                resource_group_name = "<acr-rg>"
              }
            }
          }
          sku {
            name = "Base"
            tier = "Free"
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
              enable_private_cluster = false
            }
            dns_prefix         = "aks-v3-tf-2-2401202303-dns"
            kubernetes_version = "1.29.0"
            network_profile {
              network_plugin    = "kubenet"
              load_balancer_sku = "standard"
            }
            enable_rbac = true
            identity_profile {
              kubelet_identity {
                resource_id = "/subscriptions/<subscription-id>/resourceGroups/<resource-group>/providers/Microsoft.ManagedIdentity/userAssignedIdentities/<identity-name>"
              }
            }
          }
          type = "Microsoft.ContainerService/managedClusters"
        }
        node_pools {
          api_version = "2024-01-01"
          name        = "primary"
          properties {
            count                = 1
            enable_auto_scaling  = true
            max_count            = 1
            max_pods             = 40
            min_count            = 1
            mode                 = "System"
            orchestrator_version = "1.29.0"
            os_type              = "Linux"
            type                 = "VirtualMachineScaleSets"
            vm_size              = "Standard_B4ms"
          }
          type = "Microsoft.ContainerService/managedClusters/agentPools"
        }
      }
    }
  }
}

resource "rafay_aks_cluster_v3" "demo-terraform3" {
  metadata {
    name    = "aks-v3-tf-3"
    project = "defaultproject"
  }
  spec {
    type = "aks"
    blueprint_config {
      name = "default-aks"
    }
    cloud_credentials = "aks-cred"
    config {
      kind = "aksClusterConfig"
      metadata {
        name = "aks-v3-tf-3"
      }
      spec {
        resource_group_name = "rafay-resource"
        managed_cluster {
          api_version = "2024-01-01"
          sku {
            name = "Base"
            tier = "Free"
          }
          identity {
            type = "SystemAssigned"
          }
          location = "centralindia"
          properties {
            api_server_access_profile {
              enable_private_cluster = false
            }
            dns_prefix         = "aks-v3-tf-3-2401202303-dns"
            kubernetes_version = "1.29.0"
            network_profile {
              network_plugin    = "kubenet"
              load_balancer_sku = "standard"
            }
            enable_rbac            = true
            disable_local_accounts = true
            aad_profile {
              managed           = true
              enable_azure_rbac = true
            }
          }
          type = "Microsoft.ContainerService/managedClusters"
        }
        node_pools {
          api_version = "2024-01-01"
          name        = "primary"
          properties {
            count                = 1
            enable_auto_scaling  = true
            max_count            = 1
            max_pods             = 40
            min_count            = 1
            mode                 = "System"
            orchestrator_version = "1.29.0"
            os_type              = "Linux"
            type                 = "VirtualMachineScaleSets"
            vm_size              = "Standard_B4ms"
          }
          type = "Microsoft.ContainerService/managedClusters/agentPools"
        }
      }
    }
  }
}

resource "rafay_aks_cluster_v3" "demo-terraform4" {
  metadata {
    name    = "aks-v3-tf-4"
    project = "defaultproject"
  }
  spec {
    type = "aks"
    blueprint_config {
      name = "default-aks"
    }
    cloud_credentials = "aks-cred"
    config {
      kind = "aksClusterConfig"
      metadata {
        name = "aks-v3-tf-4"
      }
      spec {
        resource_group_name = "rafay-resource"
        managed_cluster {
          api_version = "2024-01-01"
          sku {
            name = "Base"
            tier = "Free"
          }
          identity {
            type = "SystemAssigned"
          }
          location = "centralindia"
          properties {
            api_server_access_profile {
              enable_private_cluster = false
            }
            dns_prefix         = "aks-v3-tf-4-2401202303-dns"
            kubernetes_version = "1.29.0"
            network_profile {
              network_plugin    = "kubenet"
              load_balancer_sku = "standard"
            }
            enable_rbac            = true
            disable_local_accounts = true
            aad_profile {
              managed                = true
              admin_group_object_ids = ["<aad-group-object-id>"]
            }
          }
          type = "Microsoft.ContainerService/managedClusters"
        }
        node_pools {
          api_version = "2024-01-01"
          name        = "primary"
          properties {
            count                = 1
            enable_auto_scaling  = true
            max_count            = 1
            max_pods             = 40
            min_count            = 1
            mode                 = "System"
            orchestrator_version = "1.29.0"
            os_type              = "Linux"
            type                 = "VirtualMachineScaleSets"
            vm_size              = "Standard_B4ms"
          }
          type = "Microsoft.ContainerService/managedClusters/agentPools"
        }
      }
    }
  }
}

resource "rafay_aks_cluster_v3" "demo-terraform-tf" {
  metadata {
    name    = "demo-terraform-aks-v3"
    project = "defaultproject"
  }
  spec {
    type = "aks"
    blueprint_config {
      name = "default-aks"
    }
    cloud_credentials = "azure-creds"
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
      daemon_set_override {
        node_selection_enabled = false
        tolerations {
          key      = "app1dedicated"
          value    = true
          effect   = "NoSchedule"
          operator = "Equal"
        }
      }
    }
    config {
      kind = "aksClusterConfig"
      metadata {
        name = "demo-terraform"
      }
      spec {
        resource_group_name = "demo-terraform-rg"
        managed_cluster {
          api_version = "2024-01-01"
          sku {
            name = "Base"
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
            kubernetes_version = "1.29.0"
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
                enabled = true
              }
              azure_policy {
                enabled = true
              }
              azure_keyvault_secrets_provider {
                enabled = true
                config {
                  enable_secret_rotation = false
                  rotation_poll_interval = "2m"
                }
              }
            }
          }
          type = "Microsoft.ContainerService/managedClusters"
        }
        node_pools {
          api_version = "2024-01-01"
          name        = "primary"
          properties {
            count                = 1
            enable_auto_scaling  = true
            max_count            = 1
            max_pods             = 40
            min_count            = 1
            mode                 = "System"
            orchestrator_version = "1.29.0"
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

resource "rafay_aks_cluster_v3" "demo-terraform" {
  metadata {
    name    = "gautham-aks-v3-tf-1"
    project = "defaultproject"
  }
  spec {
    type = "aks"
    blueprint_config {
      name = "default-aks"
    }
    cloud_credentials = "gautham-azure-creds"
    config {
      kind = "aksClusterConfig"
      metadata {
        name = "gautham-aks-v3-tf-1"
      }
      spec {
        resource_group_name = "gautham-rg-ci"
        managed_cluster {
          api_version = "2023-11-01"
          sku {
            name = "Base"
            tier = "Free"
          }
          identity {
            type = "SystemAssigned"
          }
          location = "centralindia"
          tags = {
            "email" = "mvgautham@rafay.co"
            "env"   = "dev"
          }
          properties {
            api_server_access_profile {
              enable_private_cluster = true
            }
            dns_prefix         = "gautham-aks-v3-tf-2401202303-dns"
            kubernetes_version = "1.28.9"
            network_profile {
              network_plugin    = "kubenet"
              load_balancer_sku = "standard"
            }
            power_state {
              code = "Running"
            }

            oidc_issuer_profile {
              enabled = true
            }
            security_profile {
              workload_identity {
                enabled = false
              }
            }
          }
          type = "Microsoft.ContainerService/managedClusters"
        }
        node_pools {
          api_version = "2023-11-01"
          name        = "primary"
          location    = "centralindia"
          properties {
            count                = 1
            enable_auto_scaling  = true
            max_count            = 1
            max_pods             = 40
            min_count            = 1
            mode                 = "System"
            orchestrator_version = "1.28.9"
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

resource "rafay_aks_cluster_v3" "demo-terraform" {
  metadata {
    name    = "gautham-tf-wi-1"
    project = "defaultproject"
  }
  spec {
    type = "aks"
    blueprint_config {
      name = "minimal"
    }
    cloud_credentials = "gautham-azure-creds"
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
      daemon_set_override {
        node_selection_enabled = false
        tolerations {
          key      = "app1dedicated"
          value    = true
          effect   = "NoSchedule"
          operator = "Equal"
        }
      }
    }
    config {
      kind = "aksClusterConfig"
      metadata {
        name = "gautham-tf-wi-1"
      }
      spec {
        resource_group_name = "gautham-rg-ci"
        managed_cluster {
          api_version = "2023-11-01"
          sku {
            name = "Base"
            tier = "Free"
          }
          identity {
            type = "SystemAssigned"
          }
          location = "centralindia"
          tags = {
            "email" = "gautham@rafay.co"
            "env"   = "dev"
          }
          properties {
            api_server_access_profile {
              enable_private_cluster = true
            }
            dns_prefix         = "aks-v3-tf-2401202303-dns"
            kubernetes_version = "1.28.9"
            network_profile {
              network_plugin    = "kubenet"
              load_balancer_sku = "standard"
            }
            power_state {
              code = "Running"
            }
            addon_profiles {
              http_application_routing {
                enabled = true
              }
              azure_policy {
                enabled = true
              }
              azure_keyvault_secrets_provider {
                enabled = true
                config {
                  enable_secret_rotation = false
                  rotation_poll_interval = "2m"
                }
              }
            }
            oidc_issuer_profile {
              enabled = true
            }
            security_profile {
              workload_identity {
                enabled = true
              }
            }
          }
          type = "Microsoft.ContainerService/managedClusters"
        }
        node_pools {
          api_version = "2023-11-01"
          name        = "primary"
          location    = "centralindia"
          properties {
            count                = 1
            enable_auto_scaling  = true
            max_count            = 1
            max_pods             = 40
            min_count            = 1
            mode                 = "System"
            orchestrator_version = "1.28.9"
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

resource "rafay_aks_cluster_v3" "demo-terraform-wi-cluster" {
  metadata {
    name    = "gautham-tf-wi-v3"
    project = "defaultproject"
  }
  spec {
    type = "aks"
    blueprint_config {
      name = "minimal"
    }
    cloud_credentials = "gautham-azure-creds"

    config {
      kind = "aksClusterConfig"
      metadata {
        name = "gautham-tf-wi-v3"
      }
      spec {
        resource_group_name = "gautham-rg-ci"
        managed_cluster {
          api_version = "2023-11-01"
          sku {
            name = "Base"
            tier = "Free"
          }
          identity {
            type = "SystemAssigned"
          }
          location = "centralindia"
          tags = {
            "email" = "gautham@rafay.co"
            "env"   = "dev"
          }
          properties {
            api_server_access_profile {
              enable_private_cluster = true
            }
            dns_prefix         = "aks-v3-tf-2401202303-dns"
            kubernetes_version = "1.29.2"
            network_profile {
              network_plugin    = "kubenet"
              load_balancer_sku = "standard"
            }
            power_state {
              code = "Running"
            }
            addon_profiles {
              http_application_routing {
                enabled = true
              }
              azure_policy {
                enabled = true
              }
              azure_keyvault_secrets_provider {
                enabled = true
                config {
                  enable_secret_rotation = false
                  rotation_poll_interval = "2m"
                }
              }
            }
            oidc_issuer_profile {
              enabled = true
            }
            security_profile {
              workload_identity {
                enabled = true
              }
            }
          }
          type = "Microsoft.ContainerService/managedClusters"
        }
        node_pools {
          api_version = "2023-11-01"
          name        = "primary"
          location    = "centralindia"
          properties {
            count                = 2
            enable_auto_scaling  = true
            max_count            = 2
            max_pods             = 40
            min_count            = 2
            mode                 = "System"
            orchestrator_version = "1.29.2"
            os_type              = "Linux"
            type                 = "VirtualMachineScaleSets"
            vm_size              = "Standard_DS2_v2"
          }
          type = "Microsoft.ContainerService/managedClusters/agentPools"
        }
        node_pools {
          api_version = "2023-11-01"
          name        = "secondary"
          location    = "centralindia"
          properties {
            count                = 2
            enable_auto_scaling  = true
            max_count            = 2
            max_pods             = 40
            min_count            = 2
            mode                 = "System"
            orchestrator_version = "1.29.2"
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
resource "rafay_aks_workload_identity" "demo-terraform-wi" {
  metadata {
    cluster_name = "gautham-tf-wi-v3"
    project      = "defaultproject"
  }

  spec {
    create_identity = true

    metadata {
      name           = "gautham-tf-wi-v3-uai-1"
      location       = "centralindia"
      resource_group = "shobhit-rg"
      tags = {
        "owner"      = "gautham"
        "department" = "gautham"
        "app"        = "gautham"
      }
    }

    role_assignments {
      name  = "Key Vault Secrets User"
      scope = "/subscriptions/a2252eb2-7a25-432b-a5ec-e18eba6f26b1/resourceGroups/qa-automation/providers/Microsoft.KeyVault/vaults/gautham-rauto-kv-1"
    }

    service_accounts {
      create_account = true

      metadata {
        name      = "gautham-tf-wi-v3-sa-11"
        namespace = "aks-wi-ns"
        annotations = {
          "role" = "dev"
        }
        labels = {
          "owner"      = "gautham"
          "department" = "gautham"
        }
      }
    }
  }

  depends_on = [rafay_aks_cluster_v3.demo-terraform-wi-cluster]
}
