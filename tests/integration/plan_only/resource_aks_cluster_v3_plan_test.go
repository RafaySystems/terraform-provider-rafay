//go:build planonly
// +build planonly

package rafay_test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// Use the released provider from the Terraform Registry during tests.
var externalProvidersAKSV3 = map[string]resource.ExternalProvider{
	"rafay": {Source: "RafaySystems/rafay"},
}

// Minimal env so provider Configure() is happy in plan-only runs.
func setDummyEnvAKSV3(t *testing.T) {
	_ = os.Setenv("RCTL_API_KEY", "dummy")
	_ = os.Setenv("RCTL_PROJECT", "default")
	_ = os.Setenv("RCTL_REST_ENDPOINT", "console.example.dev")
}

// ------------------ BASIC ------------------

func TestAccAKSClusterV3_Basic_PlanOnly(t *testing.T) {
	setDummyEnvAKSV3(t)

	resource.Test(t, resource.TestCase{
		ExternalProviders: externalProvidersAKSV3,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "rafay" {}

					resource "rafay_aks_cluster_v3" "test" {
					  metadata {
					    name    = "tf-planonly-v3-basic"
					    project = "default"
					  }
					  
					  spec {
					    type             = "aks"
					    cloud_credentials = "azure-cred"
					    
					    blueprint_config {
					      name = "default"
					    }
					    
					    config {
					      api_version = "rafay.io/v1alpha5"
					      kind        = "aksClusterConfig"
					      
					      metadata {
					        name = "tf-planonly-v3-basic"
					      }
					      
					      spec {
					        subscription_id      = "12345678-1234-1234-1234-123456789012"
					        resource_group_name  = "test-rg"
					        
					        managed_cluster {
					          api_version = "2023-01-01"
					          location    = "eastus"
					          
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
					resource.TestCheckResourceAttr("rafay_aks_cluster_v3.test", "metadata.0.name", "tf-planonly-v3-basic"),
					resource.TestCheckResourceAttr("rafay_aks_cluster_v3.test", "metadata.0.project", "default"),
					resource.TestCheckResourceAttr("rafay_aks_cluster_v3.test", "spec.0.type", "aks"),
					resource.TestCheckResourceAttr("rafay_aks_cluster_v3.test", "spec.0.cloud_credentials", "azure-cred"),
				),
			},
		},
	})
}

// ------------------ NODE POOLS ------------------

func TestAccAKSClusterV3_NodePools_PlanOnly(t *testing.T) {
	setDummyEnvAKSV3(t)

	resource.Test(t, resource.TestCase{
		ExternalProviders: externalProvidersAKSV3,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "rafay" {}

					resource "rafay_aks_cluster_v3" "test" {
					  metadata {
					    name    = "tf-planonly-v3-np"
					    project = "default"
					  }
					  
					  spec {
					    type             = "aks"
					    cloud_credentials = "azure-cred"
					    
					    blueprint_config {
					      name = "default"
					    }
					    
					    config {
					      api_version = "rafay.io/v1alpha5"
					      kind        = "aksClusterConfig"
					      
					      metadata {
					        name = "tf-planonly-v3-np"
					      }
					      
					      spec {
					        subscription_id      = "12345678-1234-1234-1234-123456789012"
					        resource_group_name  = "test-rg"
					        
					        managed_cluster {
					          api_version = "2023-01-01"
					          location    = "eastus"
					          
					          properties {
					            kubernetes_version = "1.27.3"
					            dns_prefix        = "tfplanonly"
					          }
					        }
					        
					        node_pools {
					          api_version = "2023-01-01"
					          name        = "nodepool1"
					          location    = "eastus"
					          
					          properties {
					            count                 = 3
					            vm_size              = "Standard_DS2_v2"
					            os_type              = "Linux"
					            mode                 = "System"
					            availability_zones    = ["1", "2", "3"]
					            enable_auto_scaling  = true
					            min_count            = 1
					            max_count            = 5
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
					resource.TestCheckResourceAttr("rafay_aks_cluster_v3.test", "spec.0.config.0.spec.0.node_pools.0.name", "nodepool1"),
					resource.TestCheckResourceAttr("rafay_aks_cluster_v3.test", "spec.0.config.0.spec.0.node_pools.0.properties.0.vm_size", "Standard_DS2_v2"),
					resource.TestCheckResourceAttr("rafay_aks_cluster_v3.test", "spec.0.config.0.spec.0.node_pools.0.properties.0.enable_auto_scaling", "true"),
					resource.TestCheckResourceAttr("rafay_aks_cluster_v3.test", "spec.0.config.0.spec.0.node_pools.0.properties.0.availability_zones.0", "1"),
				),
			},
		},
	})
}

// ------------------ MANAGED CLUSTER ------------------

func TestAccAKSClusterV3_ManagedCluster_PlanOnly(t *testing.T) {
	setDummyEnvAKSV3(t)

	resource.Test(t, resource.TestCase{
		ExternalProviders: externalProvidersAKSV3,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "rafay" {}

					resource "rafay_aks_cluster_v3" "test" {
					  metadata {
					    name    = "tf-planonly-v3-mc"
					    project = "default"
					  }
					  
					  spec {
					    type             = "aks"
					    cloud_credentials = "azure-cred"
					    
					    blueprint_config {
					      name = "default"
					    }
					    
					    config {
					      api_version = "rafay.io/v1alpha5"
					      kind        = "aksClusterConfig"
					      
					      metadata {
					        name = "tf-planonly-v3-mc"
					      }
					      
					      spec {
					        subscription_id      = "12345678-1234-1234-1234-123456789012"
					        resource_group_name  = "test-rg"
					        
					        managed_cluster {
					          api_version = "2023-01-01"
					          location    = "eastus"
					          
					          identity {
					            type = "SystemAssigned"
					          }
					          
					          properties {
					            kubernetes_version = "1.27.3"
					            dns_prefix        = "tfplanonly"
					            enable_rbac       = true
					            node_resource_group = "test-node-rg"
					          }
					          
					          sku {
					            name = "Base"
					            tier = "Standard"
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
					resource.TestCheckResourceAttr("rafay_aks_cluster_v3.test", "spec.0.config.0.spec.0.managed_cluster.0.properties.0.kubernetes_version", "1.27.3"),
					resource.TestCheckResourceAttr("rafay_aks_cluster_v3.test", "spec.0.config.0.spec.0.managed_cluster.0.properties.0.dns_prefix", "tfplanonly"),
					resource.TestCheckResourceAttr("rafay_aks_cluster_v3.test", "spec.0.config.0.spec.0.managed_cluster.0.sku.0.tier", "Standard"),
				),
			},
		},
	})
}

// ------------------ SHARING ------------------

func TestAccAKSClusterV3_Sharing_PlanOnly(t *testing.T) {
	setDummyEnvAKSV3(t)

	resource.Test(t, resource.TestCase{
		ExternalProviders: externalProvidersAKSV3,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "rafay" {}

					resource "rafay_aks_cluster_v3" "test" {
					  metadata {
					    name    = "tf-planonly-v3-share"
					    project = "default"
					  }
					  
					  spec {
					    type             = "aks"
					    cloud_credentials = "azure-cred"
					    
					    blueprint_config {
					      name = "default"
					    }
					    
					    sharing {
					      enabled = true
					      projects {
					        name = "project1"
					      }
					      projects {
					        name = "project2"
					      }
					    }
					    
					    config {
					      api_version = "rafay.io/v1alpha5"
					      kind        = "aksClusterConfig"
					      
					      metadata {
					        name = "tf-planonly-v3-share"
					      }
					      
					      spec {
					        subscription_id      = "12345678-1234-1234-1234-123456789012"
					        resource_group_name  = "test-rg"
					        
					        managed_cluster {
					          api_version = "2023-01-01"
					          location    = "eastus"
					          
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
					resource.TestCheckResourceAttr("rafay_aks_cluster_v3.test", "spec.0.sharing.0.enabled", "true"),
					resource.TestCheckResourceAttr("rafay_aks_cluster_v3.test", "spec.0.sharing.0.projects.0.name", "project1"),
					resource.TestCheckResourceAttr("rafay_aks_cluster_v3.test", "spec.0.sharing.0.projects.1.name", "project2"),
				),
			},
		},
	})
}

// ------------------ SECURITY PROFILE ------------------

func TestAccAKSClusterV3_SecurityProfile_PlanOnly(t *testing.T) {
	setDummyEnvAKSV3(t)

	resource.Test(t, resource.TestCase{
		ExternalProviders: externalProvidersAKSV3,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "rafay" {}

					resource "rafay_aks_cluster_v3" "test" {
					  metadata {
					    name    = "tf-planonly-v3-sec"
					    project = "default"
					  }
					  
					  spec {
					    type             = "aks"
					    cloud_credentials = "azure-cred"
					    
					    blueprint_config {
					      name = "default"
					    }
					    
					    config {
					      api_version = "rafay.io/v1alpha5"
					      kind        = "aksClusterConfig"
					      
					      metadata {
					        name = "tf-planonly-v3-sec"
					      }
					      
					      spec {
					        subscription_id      = "12345678-1234-1234-1234-123456789012"
					        resource_group_name  = "test-rg"
					        
					        managed_cluster {
					          api_version = "2023-01-01"
					          location    = "eastus"
					          
					          properties {
					            kubernetes_version = "1.27.3"
					            dns_prefix        = "tfplanonly"
					            
					            security_profile {
					              workload_identity {
					                enabled = true
					              }
					            }
					            
					            oidc_issuer_profile {
					              enabled = true
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
					resource.TestCheckResourceAttr("rafay_aks_cluster_v3.test", "spec.0.config.0.spec.0.managed_cluster.0.properties.0.security_profile.0.workload_identity.0.enabled", "true"),
					resource.TestCheckResourceAttr("rafay_aks_cluster_v3.test", "spec.0.config.0.spec.0.managed_cluster.0.properties.0.oidc_issuer_profile.0.enabled", "true"),
				),
			},
		},
	})
}

// ------------------ SYSTEM COMPONENTS PLACEMENT ------------------

func TestAccAKSClusterV3_SystemComponents_PlanOnly(t *testing.T) {
	setDummyEnvAKSV3(t)

	resource.Test(t, resource.TestCase{
		ExternalProviders: externalProvidersAKSV3,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "rafay" {}

					resource "rafay_aks_cluster_v3" "test" {
					  metadata {
					    name    = "tf-planonly-v3-sys"
					    project = "default"
					  }
					  
					  spec {
					    type             = "aks"
					    cloud_credentials = "azure-cred"
					    
					    blueprint_config {
					      name = "default"
					    }
					    
					    system_components_placement {
					      node_selector = {
					        "node-type" = "system"
					      }
					      
					      tolerations {
					        key      = "CriticalAddonsOnly"
					        operator = "Exists"
					      }
					    }
					    
					    config {
					      api_version = "rafay.io/v1alpha5"
					      kind        = "aksClusterConfig"
					      
					      metadata {
					        name = "tf-planonly-v3-sys"
					      }
					      
					      spec {
					        subscription_id      = "12345678-1234-1234-1234-123456789012"
					        resource_group_name  = "test-rg"
					        
					        managed_cluster {
					          api_version = "2023-01-01"
					          location    = "eastus"
					          
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
					resource.TestCheckResourceAttr("rafay_aks_cluster_v3.test", "spec.0.system_components_placement.0.node_selector.node-type", "system"),
					resource.TestCheckResourceAttr("rafay_aks_cluster_v3.test", "spec.0.system_components_placement.0.tolerations.0.key", "CriticalAddonsOnly"),
				),
			},
		},
	})
}
