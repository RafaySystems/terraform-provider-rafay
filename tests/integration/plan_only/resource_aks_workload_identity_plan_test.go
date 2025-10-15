//go:build planonly
// +build planonly

package rafay_test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// Use the released provider from the Terraform Registry during tests.
var externalProvidersWI = map[string]resource.ExternalProvider{
	"rafay": {Source: "RafaySystems/rafay"},
}

// Minimal env so provider Configure() is happy in plan-only runs.
func setDummyEnvWI(t *testing.T) {
	_ = os.Setenv("RCTL_API_KEY", "dummy")
	_ = os.Setenv("RCTL_PROJECT", "default")
	_ = os.Setenv("RCTL_REST_ENDPOINT", "console.example.dev")
}

// ------------------ BASIC ------------------

func TestAccAKSWorkloadIdentity_Basic_PlanOnly(t *testing.T) {
	setDummyEnvWI(t)

	resource.Test(t, resource.TestCase{
		ExternalProviders: externalProvidersWI,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "rafay" {}

					resource "rafay_aks_workload_identity" "test" {
					  metadata {
					    name        = "test-wi"
					    cluster_name = "test-cluster"
					    project     = "default"
					  }
					  
					  spec {
					    metadata {
					      name           = "test-wi"
					      resource_group = "test-rg"
					      tenant_id      = "12345678-1234-1234-1234-123456789012"
					    }
					  }
					}
				`,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rafay_aks_workload_identity.test", "metadata.0.name", "test-wi"),
					resource.TestCheckResourceAttr("rafay_aks_workload_identity.test", "metadata.0.clustername", "test-cluster"),
					resource.TestCheckResourceAttr("rafay_aks_workload_identity.test", "metadata.0.project", "default"),
					resource.TestCheckResourceAttr("rafay_aks_workload_identity.test", "spec.0.metadata.0.name", "test-wi"),
					resource.TestCheckResourceAttr("rafay_aks_workload_identity.test", "spec.0.metadata.0.resource_group", "test-rg"),
				),
			},
		},
	})
}

// ------------------ WITH ROLE ASSIGNMENTS ------------------

func TestAccAKSWorkloadIdentity_WithRoleAssignments_PlanOnly(t *testing.T) {
	setDummyEnvWI(t)

	resource.Test(t, resource.TestCase{
		ExternalProviders: externalProvidersWI,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "rafay" {}

					resource "rafay_aks_workload_identity" "test" {
					  metadata {
					    name        = "test-wi-roles"
					    cluster_name = "test-cluster"
					    project     = "default"
					  }
					  
					  spec {
					    metadata {
					      name           = "test-wi-roles"
					      resource_group = "test-rg"
					      tenant_id      = "12345678-1234-1234-1234-123456789012"
					    }
					    
					    role_assignments {
					      scope                = "/subscriptions/12345678-1234-1234-1234-123456789012/resourceGroups/test-rg"
					      role_definition_name = "Storage Blob Data Reader"
					    }
					    
					    role_assignments {
					      scope                = "/subscriptions/12345678-1234-1234-1234-123456789012/resourceGroups/test-rg/providers/Microsoft.KeyVault/vaults/test-kv"
					      role_definition_name = "Key Vault Secrets User"
					    }
					  }
					}
				`,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rafay_aks_workload_identity.test", "metadata.0.name", "test-wi-roles"),
					resource.TestCheckResourceAttr("rafay_aks_workload_identity.test", "spec.0.role_assignments.0.role_definition_name", "Storage Blob Data Reader"),
					resource.TestCheckResourceAttr("rafay_aks_workload_identity.test", "spec.0.role_assignments.1.role_definition_name", "Key Vault Secrets User"),
				),
			},
		},
	})
}

// ------------------ WITH SERVICE ACCOUNTS ------------------

func TestAccAKSWorkloadIdentity_WithServiceAccounts_PlanOnly(t *testing.T) {
	setDummyEnvWI(t)

	resource.Test(t, resource.TestCase{
		ExternalProviders: externalProvidersWI,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "rafay" {}

					resource "rafay_aks_workload_identity" "test" {
					  metadata {
					    name        = "test-wi-sa"
					    cluster_name = "test-cluster"
					    project     = "default"
					  }
					  
					  spec {
					    metadata {
					      name           = "test-wi-sa"
					      resource_group = "test-rg"
					      tenant_id      = "12345678-1234-1234-1234-123456789012"
					    }
					    
					    service_accounts {
					      k8s_metadata {
					        name      = "app-sa"
					        namespace = "app-namespace"
					        
					        labels = {
					          "app" = "myapp"
					        }
					        
					        annotations = {
					          "custom.annotation" = "value"
					        }
					      }
					    }
					    
					    service_accounts {
					      k8s_metadata {
					        name      = "worker-sa"
					        namespace = "worker-namespace"
					      }
					    }
					  }
					}
				`,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rafay_aks_workload_identity.test", "metadata.0.name", "test-wi-sa"),
					resource.TestCheckResourceAttr("rafay_aks_workload_identity.test", "spec.0.service_accounts.0.k8s_metadata.0.name", "app-sa"),
					resource.TestCheckResourceAttr("rafay_aks_workload_identity.test", "spec.0.service_accounts.0.k8s_metadata.0.namespace", "app-namespace"),
					resource.TestCheckResourceAttr("rafay_aks_workload_identity.test", "spec.0.service_accounts.0.k8s_metadata.0.labels.app", "myapp"),
					resource.TestCheckResourceAttr("rafay_aks_workload_identity.test", "spec.0.service_accounts.1.k8s_metadata.0.name", "worker-sa"),
				),
			},
		},
	})
}

// ------------------ WITH TAGS ------------------

func TestAccAKSWorkloadIdentity_WithTags_PlanOnly(t *testing.T) {
	setDummyEnvWI(t)

	resource.Test(t, resource.TestCase{
		ExternalProviders: externalProvidersWI,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "rafay" {}

					resource "rafay_aks_workload_identity" "test" {
					  metadata {
					    name        = "test-wi-tags"
					    cluster_name = "test-cluster"
					    project     = "default"
					  }
					  
					  spec {
					    metadata {
					      name           = "test-wi-tags"
					      resource_group = "test-rg"
					      tenant_id      = "12345678-1234-1234-1234-123456789012"
					    }
					    
					    tags = {
					      "environment" = "test"
					      "team"        = "platform"
					      "cost-center" = "engineering"
					    }
					  }
					}
				`,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rafay_aks_workload_identity.test", "metadata.0.name", "test-wi-tags"),
					resource.TestCheckResourceAttr("rafay_aks_workload_identity.test", "spec.0.tags.environment", "test"),
					resource.TestCheckResourceAttr("rafay_aks_workload_identity.test", "spec.0.tags.team", "platform"),
					resource.TestCheckResourceAttr("rafay_aks_workload_identity.test", "spec.0.tags.cost-center", "engineering"),
				),
			},
		},
	})
}
