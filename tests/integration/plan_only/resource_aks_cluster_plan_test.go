//go:build planonly
// +build planonly

package rafay_test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// Use the released provider from the Terraform Registry during tests.
var externalProvidersAKS = map[string]resource.ExternalProvider{
	"rafay": {Source: "RafaySystems/rafay"},
}

// Minimal env so provider Configure() is happy in plan-only runs.
func setDummyEnvAKS(t *testing.T) {
	_ = os.Setenv("RCTL_API_KEY", "dummy")
	_ = os.Setenv("RCTL_PROJECT", "default")
	_ = os.Setenv("RCTL_REST_ENDPOINT", "console.example.dev")
}

// ------------------ BASIC ------------------

func TestAccAKSCluster_Basic_PlanOnly(t *testing.T) {
	setDummyEnvAKS(t)

	resource.Test(t, resource.TestCase{
		ExternalProviders: externalProvidersAKS,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "rafay" {}

					resource "rafay_aks_cluster" "test" {
					  apiversion = "rafay.io/v1alpha5"
					  kind       = "Cluster"
					  
					  metadata {
					    name    = "tf-planonly-basic"
					    project = "default"
					  }
					  
					  spec {
					    type          = "aks"
					    blueprint     = "default-aks"
					    cloudprovider = "azure-cred"
					    
					    cluster_config {
					      apiversion = "rafay.io/v1alpha5"
					      kind       = "aksClusterConfig"
					      
					      metadata {
					        name = "tf-planonly-basic"
					      }
					      
					      spec {
					        subscription_id      = "12345678-1234-1234-1234-123456789012"
					        resource_group_name  = "test-rg"
					        
					        managed_cluster {
					          apiversion = "2023-01-01"
					          location   = "eastus"
					          
					          properties {
					            kubernetes_version = "1.27.3"
					            dns_prefix        = "tfplanonly"
					          }
					        }
					      }
					    }
					  }
					}
				`,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rafay_aks_cluster.test", "metadata.0.name", "tf-planonly-basic"),
					resource.TestCheckResourceAttr("rafay_aks_cluster.test", "metadata.0.project", "default"),
					resource.TestCheckResourceAttr("rafay_aks_cluster.test", "spec.0.type", "aks"),
					resource.TestCheckResourceAttr("rafay_aks_cluster.test", "spec.0.blueprint", "default-aks"),
					resource.TestCheckResourceAttr("rafay_aks_cluster.test", "spec.0.cloudprovider", "azure-cred"),
				),
			},
		},
	})
}

// ------------------ NODE POOLS ------------------

func TestAccAKSCluster_NodePools_Defaults_PlanOnly(t *testing.T) {
	setDummyEnvAKS(t)

	resource.Test(t, resource.TestCase{
		ExternalProviders: externalProvidersAKS,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "rafay" {}

					resource "rafay_aks_cluster" "test" {
					  apiversion = "rafay.io/v1alpha5"
					  kind       = "Cluster"
					  
					  metadata {
					    name    = "tf-planonly-np"
					    project = "default"
					  }
					  
					  spec {
					    type          = "aks"
					    blueprint     = "default-aks"
					    cloudprovider = "azure-cred"
					    
					    cluster_config {
					      apiversion = "rafay.io/v1alpha5"
					      kind       = "aksClusterConfig"
					      
					      metadata {
					        name = "tf-planonly-np"
					      }
					      
					      spec {
					        subscription_id      = "12345678-1234-1234-1234-123456789012"
					        resource_group_name  = "test-rg"
					        
					        managed_cluster {
					          apiversion = "2023-01-01"
					          location   = "eastus"
					          
					          properties {
					            kubernetes_version = "1.27.3"
					            dns_prefix        = "tfplanonly"
					          }
					        }
					        
					        node_pools {
					          apiversion = "2023-01-01"
					          name      = "nodepool1"
					          
					          properties {
					            count   = 1
					            vm_size = "Standard_DS2_v2"
					            os_type = "Linux"
					            mode    = "System"
					          }
					        }
					      }
					    }
					  }
					}
				`,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rafay_aks_cluster.test", "spec.0.config.0.spec.0.node_pools.0.name", "nodepool1"),
					resource.TestCheckResourceAttr("rafay_aks_cluster.test", "spec.0.config.0.spec.0.node_pools.0.properties.0.vm_size", "Standard_DS2_v2"),
					resource.TestCheckResourceAttr("rafay_aks_cluster.test", "spec.0.config.0.spec.0.node_pools.0.properties.0.os_type", "Linux"),
				),
			},
		},
	})
}

// ------------------ MANAGED CLUSTER ------------------

func TestAccAKSCluster_ManagedCluster_PlanOnly(t *testing.T) {
	setDummyEnvAKS(t)

	resource.Test(t, resource.TestCase{
		ExternalProviders: externalProvidersAKS,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "rafay" {}

					resource "rafay_aks_cluster" "test" {
					  apiversion = "rafay.io/v1alpha5"
					  kind       = "Cluster"
					  
					  metadata {
					    name    = "tf-planonly-mc"
					    project = "default"
					  }
					  
					  spec {
					    type          = "aks"
					    blueprint     = "default-aks"
					    cloudprovider = "azure-cred"
					    
					    cluster_config {
					      apiversion = "rafay.io/v1alpha5"
					      kind       = "aksClusterConfig"
					      
					      metadata {
					        name = "tf-planonly-mc"
					      }
					      
					      spec {
					        subscription_id      = "12345678-1234-1234-1234-123456789012"
					        resource_group_name  = "test-rg"
					        
					        managed_cluster {
					          apiversion = "2023-01-01"
					          location   = "eastus"
					          
					          identity {
					            type = "SystemAssigned"
					          }
					          
					          properties {
					            kubernetes_version = "1.27.3"
					            dns_prefix        = "tfplanonly"
					            enable_rbac       = true
					          }
					          
					          sku {
					            name = "Base"
					            tier = "Free"
					          }
					        }
					      }
					    }
					  }
					}
				`,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rafay_aks_cluster.test", "spec.0.config.0.spec.0.managed_cluster.0.identity.0.type", "SystemAssigned"),
					resource.TestCheckResourceAttr("rafay_aks_cluster.test", "spec.0.config.0.spec.0.managed_cluster.0.properties.0.enable_rbac", "true"),
					resource.TestCheckResourceAttr("rafay_aks_cluster.test", "spec.0.config.0.spec.0.managed_cluster.0.sku.0.tier", "Free"),
				),
			},
		},
	})
}

// ------------------ NETWORK PROFILE ------------------

func TestAccAKSCluster_NetworkProfile_PlanOnly(t *testing.T) {
	setDummyEnvAKS(t)

	resource.Test(t, resource.TestCase{
		ExternalProviders: externalProvidersAKS,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "rafay" {}

					resource "rafay_aks_cluster" "test" {
					  apiversion = "rafay.io/v1alpha5"
					  kind       = "Cluster"
					  
					  metadata {
					    name    = "tf-planonly-net"
					    project = "default"
					  }
					  
					  spec {
					    type          = "aks"
					    blueprint     = "default-aks"
					    cloudprovider = "azure-cred"
					    
					    cluster_config {
					      apiversion = "rafay.io/v1alpha5"
					      kind       = "aksClusterConfig"
					      
					      metadata {
					        name = "tf-planonly-net"
					      }
					      
					      spec {
					        subscription_id      = "12345678-1234-1234-1234-123456789012"
					        resource_group_name  = "test-rg"
					        
					        managed_cluster {
					          apiversion = "2023-01-01"
					          location   = "eastus"
					          
					          properties {
					            kubernetes_version = "1.27.3"
					            dns_prefix        = "tfplanonly"
					            
					            network_profile {
					              network_plugin = "azure"
					              network_policy = "calico"
					              service_cidr   = "10.0.0.0/16"
					              dns_service_ip = "10.0.0.10"
					              pod_cidr       = "10.244.0.0/16"
					            }
					          }
					        }
					      }
					    }
					  }
					}
				`,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rafay_aks_cluster.test", "spec.0.config.0.spec.0.managed_cluster.0.properties.0.network_profile.0.network_plugin", "azure"),
					resource.TestCheckResourceAttr("rafay_aks_cluster.test", "spec.0.config.0.spec.0.managed_cluster.0.properties.0.network_profile.0.network_policy", "calico"),
					resource.TestCheckResourceAttr("rafay_aks_cluster.test", "spec.0.config.0.spec.0.managed_cluster.0.properties.0.network_profile.0.service_cidr", "10.0.0.0/16"),
				),
			},
		},
	})
}

// ------------------ IDENTITY ------------------

func TestAccAKSCluster_Identity_PlanOnly(t *testing.T) {
	setDummyEnvAKS(t)

	resource.Test(t, resource.TestCase{
		ExternalProviders: externalProvidersAKS,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "rafay" {}

					resource "rafay_aks_cluster" "test" {
					  apiversion = "rafay.io/v1alpha5"
					  kind       = "Cluster"
					  
					  metadata {
					    name    = "tf-planonly-id"
					    project = "default"
					  }
					  
					  spec {
					    type          = "aks"
					    blueprint     = "default-aks"
					    cloudprovider = "azure-cred"
					    
					    cluster_config {
					      apiversion = "rafay.io/v1alpha5"
					      kind       = "aksClusterConfig"
					      
					      metadata {
					        name = "tf-planonly-id"
					      }
					      
					      spec {
					        subscription_id      = "12345678-1234-1234-1234-123456789012"
					        resource_group_name  = "test-rg"
					        
					        managed_cluster {
					          apiversion = "2023-01-01"
					          location   = "eastus"
					          
					          identity {
					            type = "UserAssigned"
					            user_assigned_identities = {
					              "/subscriptions/12345678-1234-1234-1234-123456789012/resourceGroups/test-rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/test-identity" = "{}"
					            }
					          }
					          
					          properties {
					            kubernetes_version = "1.27.3"
					            dns_prefix        = "tfplanonly"
					          }
					        }
					      }
					    }
					  }
					}
				`,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rafay_aks_cluster.test", "spec.0.config.0.spec.0.managed_cluster.0.identity.0.type", "UserAssigned"),
				),
			},
		},
	})
}

// ------------------ ADDONS ------------------

func TestAccAKSCluster_Addons_PlanOnly(t *testing.T) {
	setDummyEnvAKS(t)

	resource.Test(t, resource.TestCase{
		ExternalProviders: externalProvidersAKS,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "rafay" {}

					resource "rafay_aks_cluster" "test" {
					  apiversion = "rafay.io/v1alpha5"
					  kind       = "Cluster"
					  
					  metadata {
					    name    = "tf-planonly-addons"
					    project = "default"
					  }
					  
					  spec {
					    type          = "aks"
					    blueprint     = "default-aks"
					    cloudprovider = "azure-cred"
					    
					    cluster_config {
					      apiversion = "rafay.io/v1alpha5"
					      kind       = "aksClusterConfig"
					      
					      metadata {
					        name = "tf-planonly-addons"
					      }
					      
					      spec {
					        subscription_id      = "12345678-1234-1234-1234-123456789012"
					        resource_group_name  = "test-rg"
					        
					        managed_cluster {
					          apiversion = "2023-01-01"
					          location   = "eastus"
					          
					          properties {
					            kubernetes_version = "1.27.3"
					            dns_prefix        = "tfplanonly"
					            
					            addon_profiles {
					              azure_policy {
					                enabled = true
					              }
					              
					              http_application_routing {
					                enabled = false
					              }
					            }
					          }
					        }
					      }
					    }
					  }
					}
				`,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rafay_aks_cluster.test", "spec.0.config.0.spec.0.managed_cluster.0.properties.0.addon_profiles.0.azure_policy.0.enabled", "true"),
					resource.TestCheckResourceAttr("rafay_aks_cluster.test", "spec.0.config.0.spec.0.managed_cluster.0.properties.0.addon_profiles.0.http_application_routing.0.enabled", "false"),
				),
			},
		},
	})
}

// ------------------ MAINTENANCE WINDOWS ------------------

func TestAccAKSCluster_MaintenanceWindows_PlanOnly(t *testing.T) {
	setDummyEnvAKS(t)

	resource.Test(t, resource.TestCase{
		ExternalProviders: externalProvidersAKS,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "rafay" {}

					resource "rafay_aks_cluster" "test" {
					  apiversion = "rafay.io/v1alpha5"
					  kind       = "Cluster"
					  
					  metadata {
					    name    = "tf-planonly-maint"
					    project = "default"
					  }
					  
					  spec {
					    type          = "aks"
					    blueprint     = "default-aks"
					    cloudprovider = "azure-cred"
					    
					    cluster_config {
					      apiversion = "rafay.io/v1alpha5"
					      kind       = "aksClusterConfig"
					      
					      metadata {
					        name = "tf-planonly-maint"
					      }
					      
					      spec {
					        subscription_id      = "12345678-1234-1234-1234-123456789012"
					        resource_group_name  = "test-rg"
					        
					        managed_cluster {
					          apiversion = "2023-01-01"
					          location   = "eastus"
					          
					          properties {
					            kubernetes_version = "1.27.3"
					            dns_prefix        = "tfplanonly"
					          }
					        }
					        
					        maintenance_configurations {
					          apiversion = "2023-03-02-preview"
					          name       = "default"
					          type       = "Microsoft.ContainerService/managedClusters/maintenanceConfigurations"
					          
					          properties {
					            time_in_week {
					              day        = "Monday"
					              hour_slots = [1, 2, 3]
					            }
					          }
					        }
					      }
					    }
					  }
					}
				`,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rafay_aks_cluster.test", "spec.0.config.0.spec.0.maintenance_configurations.0.name", "default"),
					resource.TestCheckResourceAttr("rafay_aks_cluster.test", "spec.0.config.0.spec.0.maintenance_configurations.0.properties.0.time_in_week.0.day", "Monday"),
				),
			},
		},
	})
}

// ------------------ AAD PROFILE ------------------

func TestAccAKSCluster_AADProfile_PlanOnly(t *testing.T) {
	setDummyEnvAKS(t)

	resource.Test(t, resource.TestCase{
		ExternalProviders: externalProvidersAKS,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "rafay" {}

					resource "rafay_aks_cluster" "test" {
					  apiversion = "rafay.io/v1alpha5"
					  kind       = "Cluster"
					  
					  metadata {
					    name    = "tf-planonly-aad"
					    project = "default"
					  }
					  
					  spec {
					    type          = "aks"
					    blueprint     = "default-aks"
					    cloudprovider = "azure-cred"
					    
					    cluster_config {
					      apiversion = "rafay.io/v1alpha5"
					      kind       = "aksClusterConfig"
					      
					      metadata {
					        name = "tf-planonly-aad"
					      }
					      
					      spec {
					        subscription_id      = "12345678-1234-1234-1234-123456789012"
					        resource_group_name  = "test-rg"
					        
					        managed_cluster {
					          apiversion = "2023-01-01"
					          location   = "eastus"
					          
					          properties {
					            kubernetes_version = "1.27.3"
					            dns_prefix        = "tfplanonly"
					            
					            aad_profile {
					              managed          = true
					              enable_azure_rbac = true
					              admin_group_object_ids = [
					                "12345678-1234-1234-1234-123456789012"
					              ]
					            }
					          }
					        }
					      }
					    }
					  }
					}
				`,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rafay_aks_cluster.test", "spec.0.config.0.spec.0.managed_cluster.0.properties.0.aad_profile.0.managed", "true"),
					resource.TestCheckResourceAttr("rafay_aks_cluster.test", "spec.0.config.0.spec.0.managed_cluster.0.properties.0.aad_profile.0.enable_azure_rbac", "true"),
				),
			},
		},
	})
}
