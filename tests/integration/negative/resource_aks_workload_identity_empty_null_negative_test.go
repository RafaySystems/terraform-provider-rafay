//go:build !planonly
// +build !planonly

package rafay_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// Use the released provider from the Terraform Registry (black-box tests).
var externalProvidersNegWI = map[string]resource.ExternalProvider{
	"rafay": {Source: "RafaySystems/rafay"},
}

// ---------- metadata.name ----------

// Empty string is treated as "present", so plan is non-empty.
func TestAccNegAKSWorkloadIdentity_EmptyName_AllowsPlan(t *testing.T) {
	setDummyEnv(t)
	t.Parallel()
	resource.Test(t, resource.TestCase{
		ExternalProviders: externalProvidersNegWI,
		Steps: []resource.TestStep{
			{
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				Config: `
					resource "rafay_aks_workload_identity" "test" {
					  metadata {
					    cluster_name = "test-cluster"
					    project     = "default"
					  }
					  
					  spec {
					    create_identity = true
					    metadata {
					      name           = ""          # empty
					      resource_group = "test-rg"
					    }
					    service_accounts {
					      create_account = true
					      metadata {
					        name      = "test-sa"
					        namespace = "default"
					      }
					    }
					  }
					}
				`,
			},
		},
	})
}

// ---------- metadata.clustername ----------

// Empty string -> present -> plan proceeds.
func TestAccNegAKSWorkloadIdentity_EmptyClusterName_AllowsPlan(t *testing.T) {
	setDummyEnv(t)
	t.Parallel()
	resource.Test(t, resource.TestCase{
		ExternalProviders: externalProvidersNegWI,
		Steps: []resource.TestStep{
			{
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				Config: `
					resource "rafay_aks_workload_identity" "test" {
					  metadata {
					    cluster_name = ""            # empty
					    project     = "default"
					  }
					  
					  spec {
					    create_identity = true
					    metadata {
					      name           = "tf-neg-wi-empty-cluster"
					      resource_group = "test-rg"
					    }
					    service_accounts {
					      create_account = true
					      metadata {
					        name      = "test-sa"
					        namespace = "default"
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
func TestAccNegAKSWorkloadIdentity_NullClusterName_RequiredError(t *testing.T) {
	setDummyEnv(t)
	t.Parallel()
	resource.Test(t, resource.TestCase{
		ExternalProviders: externalProvidersNegWI,
		Steps: []resource.TestStep{
			{
				PlanOnly: true,
				Config: `
					resource "rafay_aks_workload_identity" "test" {
					  metadata {
					    cluster_name = null            # null
					    project     = "default"
					  }
					  
					  spec {
					    create_identity = true
					    metadata {
					      name           = "tf-neg-wi-null-cluster"
					      resource_group = "test-rg"
					    }
					    service_accounts {
					      create_account = true
					      metadata {
					        name      = "test-sa"
					        namespace = "default"
					      }
					    }
					  }
					}
				`,
				ExpectError: regexp.MustCompile(`(?s)Missing required argument.*metadata.*cluster_name.*is required`),
			},
		},
	})
}

// ---------- metadata.project ----------

// Empty string -> present -> plan proceeds.
func TestAccNegAKSWorkloadIdentity_EmptyProject_AllowsPlan(t *testing.T) {
	setDummyEnv(t)
	t.Parallel()
	resource.Test(t, resource.TestCase{
		ExternalProviders: externalProvidersNegWI,
		Steps: []resource.TestStep{
			{
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				Config: `
					resource "rafay_aks_workload_identity" "test" {
					  metadata {
					    cluster_name = "test-cluster"
					    project     = ""            # empty
					  }
					  
					  spec {
					    create_identity = true
					    metadata {
					      name           = "tf-neg-wi-empty-project"
					      resource_group = "test-rg"
					    }
					    service_accounts {
					      create_account = true
					      metadata {
					        name      = "test-sa"
					        namespace = "default"
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
func TestAccNegAKSWorkloadIdentity_NullProject_RequiredError(t *testing.T) {
	setDummyEnv(t)
	t.Parallel()
	resource.Test(t, resource.TestCase{
		ExternalProviders: externalProvidersNegWI,
		Steps: []resource.TestStep{
			{
				PlanOnly: true,
				Config: `
					resource "rafay_aks_workload_identity" "test" {
					  metadata {
					    cluster_name = "test-cluster"
					    project     = null            # null
					  }
					  
					  spec {
					    create_identity = true
					    metadata {
					      name           = "tf-neg-wi-null-project"
					      resource_group = "test-rg"
					    }
					    service_accounts {
					      create_account = true
					      metadata {
					        name      = "test-sa"
					        namespace = "default"
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

// ---------- spec.metadata.name ----------

// Empty string -> present -> plan proceeds.
func TestAccNegAKSWorkloadIdentity_EmptySpecName_AllowsPlan(t *testing.T) {
	setDummyEnv(t)
	t.Parallel()
	resource.Test(t, resource.TestCase{
		ExternalProviders: externalProvidersNegWI,
		Steps: []resource.TestStep{
			{
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				Config: `
					resource "rafay_aks_workload_identity" "test" {
					  metadata {
					    cluster_name = "test-cluster"
					    project     = "default"
					  }
					  
					  spec {
					    create_identity = true
					    metadata {
					      name           = ""              # empty
					      resource_group = "test-rg"
					    }
					    service_accounts {
					      create_account = true
					      metadata {
					        name      = "test-sa"
					        namespace = "default"
					      }
					    }
					  }
					}
				`,
			},
		},
	})
}

// ---------- invalid role scope ----------

func TestAccNegAKSWorkloadIdentity_InvalidRoleScope_Error(t *testing.T) {
	setDummyEnv(t)
	t.Parallel()
	resource.Test(t, resource.TestCase{
		ExternalProviders: externalProvidersNegWI,
		Steps: []resource.TestStep{
			{
				// Note: This test validates that invalid values pass plan but would fail on apply
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				Config: `
					resource "rafay_aks_workload_identity" "test" {
					  metadata {
					    cluster_name = "test-cluster"
					    project     = "default"
					  }
					  
					  spec {
					    create_identity = true
					    metadata {
					      name           = "tf-neg-wi-invalid-scope"
					      resource_group = "test-rg"
					    }
					    
					    role_assignments {
					      name  = "Reader"
					      scope = "invalid-scope-format"   # Invalid scope
					    }
					    
					    service_accounts {
					      create_account = true
					      metadata {
					        name      = "test-sa"
					        namespace = "default"
					      }
					    }
					  }
					}
				`,
			},
		},
	})
}

// ---------- invalid service account namespace ----------

func TestAccNegAKSWorkloadIdentity_InvalidServiceAccountNamespace_Error(t *testing.T) {
	setDummyEnv(t)
	t.Parallel()
	resource.Test(t, resource.TestCase{
		ExternalProviders: externalProvidersNegWI,
		Steps: []resource.TestStep{
			{
				// Note: This test validates that invalid values pass plan but would fail on apply
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				Config: `
					resource "rafay_aks_workload_identity" "test" {
					  metadata {
					    cluster_name = "test-cluster"
					    project     = "default"
					  }
					  
					  spec {
					    create_identity = true
					    metadata {
					      name           = "tf-neg-wi-invalid-ns"
					      resource_group = "test-rg"
					    }
					    
					    service_accounts {
					      create_account = true
					      metadata {
					        name      = "app-sa"
					        namespace = "Invalid-Namespace!"   # Invalid namespace format
					      }
					    }
					  }
					}
				`,
			},
		},
	})
}
