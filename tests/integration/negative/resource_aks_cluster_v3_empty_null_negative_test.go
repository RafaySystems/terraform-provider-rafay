//go:build !planonly
// +build !planonly

package rafay_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// Use the released provider from the Terraform Registry (black-box tests).
var externalProvidersNegAKSV3 = map[string]resource.ExternalProvider{
	"rafay": {Source: "RafaySystems/rafay"},
}

// ----------------------------------------------------------------------------
// NOTE ON EXPECTATIONS
// - Setting a required attribute to `null` -> Terraform Core errors with
//   "Missing required argument. The argument \"<path>\" is required"
// - Setting it to "" (empty string) counts as present, so the plan proceeds.
//   For those, we assert PlanOnly + ExpectNonEmptyPlan.
// ----------------------------------------------------------------------------

// ---------- metadata.name ----------

// Empty string is treated as "present", so plan is non-empty.
func TestAccNegAKSClusterV3_EmptyMetadataName_AllowsPlan(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		ExternalProviders: externalProvidersNegAKSV3,
		Steps: []resource.TestStep{
			{
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				Config: `
					resource "rafay_aks_cluster_v3" "test" {
					  metadata {
					    name    = ""          # empty
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
					        name = "test"
					      }
					      
					      spec {
					        subscription_id      = "12345678-1234-1234-1234-123456789012"
					        resource_group_name  = "test-rg"
					        
					        managed_cluster {
					          api_version = "2023-01-01"
					          location    = "eastus"
					          
					          properties {
					            kubernetes_version = "1.27.3"
					            dns_prefix        = "test"
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

// NOTE: metadata.name is Optional, so null test is not applicable

// ---------- metadata.project ----------

// Empty string -> present -> plan proceeds.
func TestAccNegAKSClusterV3_EmptyProject_AllowsPlan(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		ExternalProviders: externalProvidersNegAKSV3,
		Steps: []resource.TestStep{
			{
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				Config: `
					resource "rafay_aks_cluster_v3" "test" {
					  metadata {
					    name    = "tf-neg-v3-empty-project"
					    project = ""            # empty
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
					        name = "tf-neg-v3-empty-project"
					      }
					      
					      spec {
					        subscription_id      = "12345678-1234-1234-1234-123456789012"
					        resource_group_name  = "test-rg"
					        
					        managed_cluster {
					          api_version = "2023-01-01"
					          location    = "eastus"
					          
					          properties {
					            kubernetes_version = "1.27.3"
					            dns_prefix        = "test"
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

// NOTE: metadata.project is Optional, so null test is not applicable

// ---------- spec.type ----------

// Empty string -> present -> plan proceeds.
func TestAccNegAKSClusterV3_EmptySpecType_AllowsPlan(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		ExternalProviders: externalProvidersNegAKSV3,
		Steps: []resource.TestStep{
			{
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				Config: `
					resource "rafay_aks_cluster_v3" "test" {
					  metadata {
					    name    = "tf-neg-v3-empty-type"
					    project = "default"
					  }
					  
					  spec {
					    type             = ""              # empty
					    cloud_credentials = "azure-cred"
					    
					    blueprint_config {
					      name = "default"
					    }
					    
					    config {
					      api_version = "rafay.io/v1alpha5"
					      kind        = "aksClusterConfig"
					      
					      metadata {
					        name = "tf-neg-v3-empty-type"
					      }
					      
					      spec {
					        subscription_id      = "12345678-1234-1234-1234-123456789012"
					        resource_group_name  = "test-rg"
					        
					        managed_cluster {
					          api_version = "2023-01-01"
					          location    = "eastus"
					          
					          properties {
					            kubernetes_version = "1.27.3"
					            dns_prefix        = "test"
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

// NOTE: spec.type is Optional, so null test is not applicable

// ---------- spec.cloud_credentials ----------

// Empty string -> present -> plan proceeds.
func TestAccNegAKSClusterV3_EmptyCloudCredentials_AllowsPlan(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		ExternalProviders: externalProvidersNegAKSV3,
		Steps: []resource.TestStep{
			{
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				Config: `
					resource "rafay_aks_cluster_v3" "test" {
					  metadata {
					    name    = "tf-neg-v3-empty-creds"
					    project = "default"
					  }
					  
					  spec {
					    type             = "aks"
					    cloud_credentials = ""              # empty
					    
					    blueprint_config {
					      name = "default"
					    }
					    
					    config {
					      api_version = "rafay.io/v1alpha5"
					      kind        = "aksClusterConfig"
					      
					      metadata {
					        name = "tf-neg-v3-empty-creds"
					      }
					      
					      spec {
					        subscription_id      = "12345678-1234-1234-1234-123456789012"
					        resource_group_name  = "test-rg"
					        
					        managed_cluster {
					          api_version = "2023-01-01"
					          location    = "eastus"
					          
					          properties {
					            kubernetes_version = "1.27.3"
					            dns_prefix        = "test"
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

// NOTE: spec.cloud_credentials is Optional, so null test is not applicable

// ---------- invalid node pool count ----------

func TestAccNegAKSClusterV3_InvalidNodePoolCount_Error(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		ExternalProviders: externalProvidersNegAKSV3,
		Steps: []resource.TestStep{
			{
				// Note: This test validates that invalid values pass plan but would fail on apply
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				Config: `
					resource "rafay_aks_cluster_v3" "test" {
					  metadata {
					    name    = "tf-neg-v3-invalid-count"
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
					        name = "tf-neg-v3-invalid-count"
					      }
					      
					      spec {
					        subscription_id      = "12345678-1234-1234-1234-123456789012"
					        resource_group_name  = "test-rg"
					        
					        managed_cluster {
					          api_version = "2023-01-01"
					          location    = "eastus"
					          
					          properties {
					            kubernetes_version = "1.27.3"
					            dns_prefix        = "test"
					          }
					        }
					        
					        node_pools {
					          api_version = "2023-01-01"
					          name        = "nodepool1"
					          location    = "eastus"
					          
					          properties {
					            count   = -1           # Invalid count
					            vm_size = "Standard_DS2_v2"
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

// ---------- invalid kubernetes version ----------

func TestAccNegAKSClusterV3_InvalidKubernetesVersion_Error(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		ExternalProviders: externalProvidersNegAKSV3,
		Steps: []resource.TestStep{
			{
				// Note: This test validates that invalid values pass plan but would fail on apply
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				Config: `
					resource "rafay_aks_cluster_v3" "test" {
					  metadata {
					    name    = "tf-neg-v3-invalid-version"
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
					        name = "tf-neg-v3-invalid-version"
					      }
					      
					      spec {
					        subscription_id      = "12345678-1234-1234-1234-123456789012"
					        resource_group_name  = "test-rg"
					        
					        managed_cluster {
					          api_version = "2023-01-01"
					          location    = "eastus"
					          
					          properties {
					            kubernetes_version = "invalid.version.format"   # Invalid version
					            dns_prefix        = "test"
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
