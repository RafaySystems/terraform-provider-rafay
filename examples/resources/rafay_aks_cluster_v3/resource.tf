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
        name = "aks-v3-tf-1"
      }
      spec {
        resource_group_name = "rafay-resource"
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
          tags = {
            "email" = "mayank@rafay.co"
            "env"   = "dev"
          }
          properties {
            api_server_access_profile {
              enable_private_cluster = true
            }
            dns_prefix         = "aks-v3-tf-2401202303-dns"
            kubernetes_version = "1.25.15"
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
          }
          type = "Microsoft.ContainerService/managedClusters"
        }
        node_pools {
          api_version = "2022-07-01"
          name        = "primary"
          location    = "centralindia"
          properties {
            count                = 1
            enable_auto_scaling  = true
            max_count            = 1
            max_pods             = 40
            min_count            = 1
            mode                 = "System"
            orchestrator_version = "1.25.15"
            os_type              = "Linux"
            type                 = "VirtualMachineScaleSets"
            vm_size              = "Standard_DS2_v2"
            # node_labels = {
            #   app       = "infra"
            #   dedicated = "true"
            # }
            # node_taints = ["app=infra:PreferNoSchedule"]
          }
          type = "Microsoft.ContainerService/managedClusters/agentPools"
        }

        node_pools {
          api_version = "2022-07-01"
          name        = "agentpool2"
          location    = "centralindia"
          properties {
            count                = 1
            enable_auto_scaling  = true
            max_count            = 1
            max_pods             = 40
            min_count            = 1
            mode                 = "System"
            orchestrator_version = "1.25.15"
            os_type              = "Linux"
            type                 = "VirtualMachineScaleSets"
            vm_size              = "Standard_B4ms"
            # node_labels = {
            #   app       = "infra"
            #   dedicated = "true"
            # }
            # node_taints = ["app=infra:PreferNoSchedule"]
          }
          type = "Microsoft.ContainerService/managedClusters/agentPools"
        }

      }
    }
  }
}
