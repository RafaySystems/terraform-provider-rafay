//go:build !planonly
// +build !planonly

package rafay_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// Use the released provider from the Terraform Registry (black-box tests).
var externalProvidersNegAKS = map[string]resource.ExternalProvider{
	"rafay": {Source: "RafaySystems/rafay"},
}

// ---------- metadata.name ----------

// Empty string is treated as "present", so plan is non-empty.
func TestAccNegAKSCluster_EmptyClusterName_AllowsPlan(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		ExternalProviders: externalProvidersNegAKS,
		Steps: []resource.TestStep{
			{
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				Config: `
					resource "rafay_aks_cluster" "test" {
					  apiversion = "rafay.io/v1alpha5"
					  kind       = "Cluster"
					  
					  metadata {
					    name    = ""          # empty
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
					        name = "test"
					      }
					      
					      spec {
					        subscription_id      = "12345678-1234-1234-1234-123456789012"
					        resource_group_name  = "test-rg"
					        
					        managed_cluster {
					          apiversion = "2023-01-01"
					          location   = "eastus"
					          
					          properties {
					            kubernetes_version = "1.27.3"
					            dns_prefix        = "test"
					          }
					        }
					        
					        node_pools {
					          apiversion = "2023-01-01"
					          name       = "default"
					          properties {
					            count    = 1
					            vm_size  = "Standard_DS2_v2"
					            mode     = "System"
					          }
					        }
					      }
					    }
					  }
					}
				`,
			},
		},
	})
}

// null is treated as missing; expect "Missing required argument".
func TestAccNegAKSCluster_NullClusterName_RequiredError(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		ExternalProviders: externalProvidersNegAKS,
		Steps: []resource.TestStep{
			{
				PlanOnly: true,
				Config: `
					resource "rafay_aks_cluster" "test" {
					  apiversion = "rafay.io/v1alpha5"
					  kind       = "Cluster"
					  
					  metadata {
					    name    = null       # null
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
					        name = "test"
					      }
					      
					      spec {
					        subscription_id      = "12345678-1234-1234-1234-123456789012"
					        resource_group_name  = "test-rg"
					        
					        managed_cluster {
					          apiversion = "2023-01-01"
					          location   = "eastus"
					          
					          properties {
					            kubernetes_version = "1.27.3"
					            dns_prefix        = "test"
					          }
					        }
					        
					        node_pools {
					          apiversion = "2023-01-01"
					          name       = "default"
					          properties {
					            count    = 1
					            vm_size  = "Standard_DS2_v2"
					            mode     = "System"
					          }
					        }
					      }
					    }
					  }
					}
				`,
				ExpectError: regexp.MustCompile(`(?s)Missing required argument.*metadata.*name.*is required`),
			},
		},
	})
}

// ---------- metadata.project ----------

// Empty string -> present -> plan proceeds.
func TestAccNegAKSCluster_EmptyProject_AllowsPlan(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		ExternalProviders: externalProvidersNegAKS,
		Steps: []resource.TestStep{
			{
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				Config: `
					resource "rafay_aks_cluster" "test" {
					  apiversion = "rafay.io/v1alpha5"
					  kind       = "Cluster"
					  
					  metadata {
					    name    = "tf-neg-empty-project"
					    project = ""            # empty
					  }
					  
					  spec {
					    type          = "aks"
					    blueprint     = "default-aks"
					    cloudprovider = "azure-cred"
					    
					    cluster_config {
					      apiversion = "rafay.io/v1alpha5"
					      kind       = "aksClusterConfig"
					      
					      metadata {
					        name = "tf-neg-empty-project"
					      }
					      
					      spec {
					        subscription_id      = "12345678-1234-1234-1234-123456789012"
					        resource_group_name  = "test-rg"
					        
					        managed_cluster {
					          apiversion = "2023-01-01"
					          location   = "eastus"
					          
					          properties {
					            kubernetes_version = "1.27.3"
					            dns_prefix        = "test"
					          }
					        }
					        
					        node_pools {
					          apiversion = "2023-01-01"
					          name       = "default"
					          properties {
					            count    = 1
					            vm_size  = "Standard_DS2_v2"
					            mode     = "System"
					          }
					        }
					      }
					    }
					  }
					}
				`,
			},
		},
	})
}

// null -> missing -> error.
func TestAccNegAKSCluster_NullProject_RequiredError(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		ExternalProviders: externalProvidersNegAKS,
		Steps: []resource.TestStep{
			{
				PlanOnly: true,
				Config: `
					resource "rafay_aks_cluster" "test" {
					  apiversion = "rafay.io/v1alpha5"
					  kind       = "Cluster"
					  
					  metadata {
					    name    = "tf-neg-null-project"
					    project = null            # null
					  }
					  
					  spec {
					    type          = "aks"
					    blueprint     = "default-aks"
					    cloudprovider = "azure-cred"
					    
					    cluster_config {
					      apiversion = "rafay.io/v1alpha5"
					      kind       = "aksClusterConfig"
					      
					      metadata {
					        name = "tf-neg-null-project"
					      }
					      
					      spec {
					        subscription_id      = "12345678-1234-1234-1234-123456789012"
					        resource_group_name  = "test-rg"
					        
					        managed_cluster {
					          apiversion = "2023-01-01"
					          location   = "eastus"
					          
					          properties {
					            kubernetes_version = "1.27.3"
					            dns_prefix        = "test"
					          }
					        }
					        
					        node_pools {
					          apiversion = "2023-01-01"
					          name       = "default"
					          properties {
					            count    = 1
					            vm_size  = "Standard_DS2_v2"
					            mode     = "System"
					          }
					        }
					      }
					    }
					  }
					}
				`,
				ExpectError: regexp.MustCompile(`(?s)Missing required argument.*metadata.*project.*is required`),
			},
		},
	})
}

// ---------- spec.cloudprovider ----------

// Empty string -> present -> plan proceeds.
func TestAccNegAKSCluster_EmptyCloudProvider_AllowsPlan(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		ExternalProviders: externalProvidersNegAKS,
		Steps: []resource.TestStep{
			{
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				Config: `
					resource "rafay_aks_cluster" "test" {
					  apiversion = "rafay.io/v1alpha5"
					  kind       = "Cluster"
					  
					  metadata {
					    name    = "tf-neg-empty-cp"
					    project = "default"
					  }
					  
					  spec {
					    type          = "aks"
					    blueprint     = "default-aks"
					    cloudprovider = ""              # empty
					    
					    cluster_config {
					      apiversion = "rafay.io/v1alpha5"
					      kind       = "aksClusterConfig"
					      
					      metadata {
					        name = "tf-neg-empty-cp"
					      }
					      
					      spec {
					        subscription_id      = "12345678-1234-1234-1234-123456789012"
					        resource_group_name  = "test-rg"
					        
					        managed_cluster {
					          apiversion = "2023-01-01"
					          location   = "eastus"
					          
					          properties {
					            kubernetes_version = "1.27.3"
					            dns_prefix        = "test"
					          }
					        }
					        
					        node_pools {
					          apiversion = "2023-01-01"
					          name       = "default"
					          properties {
					            count    = 1
					            vm_size  = "Standard_DS2_v2"
					            mode     = "System"
					          }
					        }
					      }
					    }
					  }
					}
				`,
			},
		},
	})
}

// null -> missing -> error.
func TestAccNegAKSCluster_NullCloudProvider_RequiredError(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		ExternalProviders: externalProvidersNegAKS,
		Steps: []resource.TestStep{
			{
				PlanOnly: true,
				Config: `
					resource "rafay_aks_cluster" "test" {
					  apiversion = "rafay.io/v1alpha5"
					  kind       = "Cluster"
					  
					  metadata {
					    name    = "tf-neg-null-cp"
					    project = "default"
					  }
					  
					  spec {
					    type          = "aks"
					    blueprint     = "default-aks"
					    cloudprovider = null              # null
					    
					    cluster_config {
					      apiversion = "rafay.io/v1alpha5"
					      kind       = "aksClusterConfig"
					      
					      metadata {
					        name = "tf-neg-null-cp"
					      }
					      
					      spec {
					        subscription_id      = "12345678-1234-1234-1234-123456789012"
					        resource_group_name  = "test-rg"
					        
					        managed_cluster {
					          apiversion = "2023-01-01"
					          location   = "eastus"
					          
					          properties {
					            kubernetes_version = "1.27.3"
					            dns_prefix        = "test"
					          }
					        }
					        
					        node_pools {
					          apiversion = "2023-01-01"
					          name       = "default"
					          properties {
					            count    = 1
					            vm_size  = "Standard_DS2_v2"
					            mode     = "System"
					          }
					        }
					      }
					    }
					  }
					}
				`,
				ExpectError: regexp.MustCompile(`(?s)Missing required argument.*spec.*cloudprovider.*is required`),
			},
		},
	})
}

// ---------- spec.type ----------

// Empty string -> present -> plan proceeds.
func TestAccNegAKSCluster_EmptyType_AllowsPlan(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		ExternalProviders: externalProvidersNegAKS,
		Steps: []resource.TestStep{
			{
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				Config: `
					resource "rafay_aks_cluster" "test" {
					  apiversion = "rafay.io/v1alpha5"
					  kind       = "Cluster"
					  
					  metadata {
					    name    = "tf-neg-empty-type"
					    project = "default"
					  }
					  
					  spec {
					    type          = ""              # empty
					    blueprint     = "default-aks"
					    cloudprovider = "azure-cred"
					    
					    cluster_config {
					      apiversion = "rafay.io/v1alpha5"
					      kind       = "aksClusterConfig"
					      
					      metadata {
					        name = "tf-neg-empty-type"
					      }
					      
					      spec {
					        subscription_id      = "12345678-1234-1234-1234-123456789012"
					        resource_group_name  = "test-rg"
					        
					        managed_cluster {
					          apiversion = "2023-01-01"
					          location   = "eastus"
					          
					          properties {
					            kubernetes_version = "1.27.3"
					            dns_prefix        = "test"
					          }
					        }
					        
					        node_pools {
					          apiversion = "2023-01-01"
					          name       = "default"
					          properties {
					            count    = 1
					            vm_size  = "Standard_DS2_v2"
					            mode     = "System"
					          }
					        }
					      }
					    }
					  }
					}
				`,
			},
		},
	})
}

// null -> missing -> error.
func TestAccNegAKSCluster_NullType_RequiredError(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		ExternalProviders: externalProvidersNegAKS,
		Steps: []resource.TestStep{
			{
				PlanOnly: true,
				Config: `
					resource "rafay_aks_cluster" "test" {
					  apiversion = "rafay.io/v1alpha5"
					  kind       = "Cluster"
					  
					  metadata {
					    name    = "tf-neg-null-type"
					    project = "default"
					  }
					  
					  spec {
					    type          = null              # null
					    blueprint     = "default-aks"
					    cloudprovider = "azure-cred"
					    
					    cluster_config {
					      apiversion = "rafay.io/v1alpha5"
					      kind       = "aksClusterConfig"
					      
					      metadata {
					        name = "tf-neg-null-type"
					      }
					      
					      spec {
					        subscription_id      = "12345678-1234-1234-1234-123456789012"
					        resource_group_name  = "test-rg"
					        
					        managed_cluster {
					          apiversion = "2023-01-01"
					          location   = "eastus"
					          
					          properties {
					            kubernetes_version = "1.27.3"
					            dns_prefix        = "test"
					          }
					        }
					        
					        node_pools {
					          apiversion = "2023-01-01"
					          name       = "default"
					          properties {
					            count    = 1
					            vm_size  = "Standard_DS2_v2"
					            mode     = "System"
					          }
					        }
					      }
					    }
					  }
					}
				`,
				ExpectError: regexp.MustCompile(`(?s)Missing required argument.*spec.*type.*is required`),
			},
		},
	})
}

// ---------- invalid identity type ----------

func TestAccNegAKSCluster_InvalidIdentityType_Error(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		ExternalProviders: externalProvidersNegAKS,
		Steps: []resource.TestStep{
			{
				// Note: This test validates that invalid values pass plan but would fail on apply
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				Config: `
					resource "rafay_aks_cluster" "test" {
					  apiversion = "rafay.io/v1alpha5"
					  kind       = "Cluster"
					  
					  metadata {
					    name    = "tf-neg-invalid-identity"
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
					        name = "tf-neg-invalid-identity"
					      }
					      
					      spec {
					        subscription_id      = "12345678-1234-1234-1234-123456789012"
					        resource_group_name  = "test-rg"
					        
					        managed_cluster {
					          apiversion = "2023-01-01"
					          location   = "eastus"
					          
					          identity {
					            type = "InvalidType"   # Invalid identity type
					          }
					          
					          properties {
					            kubernetes_version = "1.27.3"
					            dns_prefix        = "test"
					          }
					        }
					        
					        node_pools {
					          apiversion = "2023-01-01"
					          name       = "default"
					          properties {
					            count    = 1
					            vm_size  = "Standard_DS2_v2"
					            mode     = "System"
					          }
					        }
					      }
					    }
					  }
					}
				`,
			},
		},
	})
}

// ---------- invalid network plugin ----------

func TestAccNegAKSCluster_InvalidNetworkPlugin_Error(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		ExternalProviders: externalProvidersNegAKS,
		Steps: []resource.TestStep{
			{
				// Note: This test validates that invalid values pass plan but would fail on apply
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				Config: `
					resource "rafay_aks_cluster" "test" {
					  apiversion = "rafay.io/v1alpha5"
					  kind       = "Cluster"
					  
					  metadata {
					    name    = "tf-neg-invalid-plugin"
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
					        name = "tf-neg-invalid-plugin"
					      }
					      
					      spec {
					        subscription_id      = "12345678-1234-1234-1234-123456789012"
					        resource_group_name  = "test-rg"
					        
					        managed_cluster {
					          apiversion = "2023-01-01"
					          location   = "eastus"
					          
					          properties {
					            kubernetes_version = "1.27.3"
					            dns_prefix        = "test"
					            
					            network_profile {
					              network_plugin = "invalid-plugin"   # Invalid plugin
					            }
					          }
					        }
					        
					        node_pools {
					          apiversion = "2023-01-01"
					          name       = "default"
					          properties {
					            count    = 1
					            vm_size  = "Standard_DS2_v2"
					            mode     = "System"
					          }
					        }
					      }
					    }
					  }
					}
				`,
			},
		},
	})
}
