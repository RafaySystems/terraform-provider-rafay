resource "rafay_aks_cluster" "may-tf-1" {
  apiversion = "rafay.io/v1alpha1"
  kind       = "Cluster"
  metadata {
    name    = "may-tf-1"
    project = "mayank"
  }
  spec {
    type          = "aks"
    blueprint     = "minimal"
    cloudprovider = "aks1"
    cluster_config {
      apiversion = "rafay.io/v1alpha1"
      kind       = "aksClusterConfig"
      metadata {
        name = "may-tf-1"
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
            dns_prefix         = "may-tf-1-dns"
            enable_rbac        = true
            kubernetes_version = "1.27.9"
            network_profile {
              network_plugin = "kubenet"
            }
            power_state {
              code = "Running"
            }
            addon_profiles {
              http_application_routing {
                enabled = false
              }
            }
            auto_upgrade_profile {
              upgrade_channel = "patch"
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
            orchestrator_version = "1.27.9"
            os_type              = "Linux"
            type                 = "VirtualMachineScaleSets"
            vm_size              = "Standard_B4ms"
          }
          type     = "Microsoft.ContainerService/managedClusters/agentPools"
        }
        maintenance_configurations {
          api_version = "2024-01-01"
          name = "aksManagedNodeOSUpgradeSchedule"
          properties {
            maintenance_window {
              duration_hours = 6
              schedule {
                daily {
                  interval_days = 1
                }
              }
              start_date = "2024-06-25"
              start_time = "23:57"
            }
          }
          type = "Microsoft.ContainerService/managedClusters/maintenanceConfigurations"
        }
        maintenance_configurations {
          api_version = "2024-01-01"
          name = "aksManagedAutoUpgradeSchedule"
          properties {
            maintenance_window {
              duration_hours = 6
              schedule {
                weekly {
                  day_of_week = "Tuesday"
                  interval_weeks = 1
                }
              }
              start_date = "2024-06-25"
              start_time = "23:57"
            }
          }
          type = "Microsoft.ContainerService/managedClusters/maintenanceConfigurations"
        }
      }
    }
    sharing {
      enabled = true
      projects {
        name = "defaultproject"
      }
    }
  }
}
