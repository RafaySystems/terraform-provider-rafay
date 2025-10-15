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

// ----------------------------------------------------------------------------
// NOTE ON EXPECTATIONS
// - Setting a required attribute to `null` -> Terraform Core errors with
//   "Missing required argument. The argument \"<path>\" is required"
// - Setting it to "" (empty string) counts as present, so the plan proceeds.
//   For those, we assert PlanOnly + ExpectNonEmptyPlan.
// ----------------------------------------------------------------------------

// ---------- metadata.name ----------

// Empty string is treated as "present", so plan is non-empty.
func TestAccNegAKSWorkloadIdentity_EmptyName_AllowsPlan(t *testing.T) {
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
					    name        = ""          # empty
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
			},
		},
	})
}

// null is treated as missing; expect "Missing required argument".
func TestAccNegAKSWorkloadIdentity_NullName_RequiredError(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		ExternalProviders: externalProvidersNegWI,
		Steps: []resource.TestStep{
			{
				PlanOnly: true,
				Config: `
					resource "rafay_aks_workload_identity" "test" {
					  metadata {
					    name        = null       # null
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
				ExpectError: regexp.MustCompile(`(?s)Missing required argument.*"metadata\.0\.name" is required`),
			},
		},
	})
}

// ---------- metadata.clustername ----------

// Empty string -> present -> plan proceeds.
func TestAccNegAKSWorkloadIdentity_EmptyClusterName_AllowsPlan(t *testing.T) {
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
					    name        = "tf-neg-wi-empty-cluster"
					    clustername = ""            # empty
					    project     = "default"
					  }
					  
					  spec {
					    metadata {
					      name           = "tf-neg-wi-empty-cluster"
					      resource_group = "test-rg"
					      tenant_id      = "12345678-1234-1234-1234-123456789012"
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
	t.Parallel()
	resource.Test(t, resource.TestCase{
		ExternalProviders: externalProvidersNegWI,
		Steps: []resource.TestStep{
			{
				PlanOnly: true,
				Config: `
					resource "rafay_aks_workload_identity" "test" {
					  metadata {
					    name        = "tf-neg-wi-null-cluster"
					    clustername = null            # null
					    project     = "default"
					  }
					  
					  spec {
					    metadata {
					      name           = "tf-neg-wi-null-cluster"
					      resource_group = "test-rg"
					      tenant_id      = "12345678-1234-1234-1234-123456789012"
					    }
					  }
					}
				`,
				ExpectError: regexp.MustCompile(`(?s)Missing required argument.*"metadata\.0\.clustername" is required`),
			},
		},
	})
}

// ---------- metadata.project ----------

// Empty string -> present -> plan proceeds.
func TestAccNegAKSWorkloadIdentity_EmptyProject_AllowsPlan(t *testing.T) {
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
					    name        = "tf-neg-wi-empty-project"
					    cluster_name = "test-cluster"
					    project     = ""            # empty
					  }
					  
					  spec {
					    metadata {
					      name           = "tf-neg-wi-empty-project"
					      resource_group = "test-rg"
					      tenant_id      = "12345678-1234-1234-1234-123456789012"
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
	t.Parallel()
	resource.Test(t, resource.TestCase{
		ExternalProviders: externalProvidersNegWI,
		Steps: []resource.TestStep{
			{
				PlanOnly: true,
				Config: `
					resource "rafay_aks_workload_identity" "test" {
					  metadata {
					    name        = "tf-neg-wi-null-project"
					    cluster_name = "test-cluster"
					    project     = null            # null
					  }
					  
					  spec {
					    metadata {
					      name           = "tf-neg-wi-null-project"
					      resource_group = "test-rg"
					      tenant_id      = "12345678-1234-1234-1234-123456789012"
					    }
					  }
					}
				`,
				ExpectError: regexp.MustCompile(`(?s)Missing required argument.*"metadata\.0\.project" is required`),
			},
		},
	})
}

// ---------- spec.metadata.name ----------

// Empty string -> present -> plan proceeds.
func TestAccNegAKSWorkloadIdentity_EmptySpecName_AllowsPlan(t *testing.T) {
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
					    name        = "tf-neg-wi-empty-spec-name"
					    cluster_name = "test-cluster"
					    project     = "default"
					  }
					  
					  spec {
					    metadata {
					      name           = ""              # empty
					      resource_group = "test-rg"
					      tenant_id      = "12345678-1234-1234-1234-123456789012"
					    }
					  }
					}
				`,
			},
		},
	})
}

// null -> missing -> error.
func TestAccNegAKSWorkloadIdentity_NullSpecName_RequiredError(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		ExternalProviders: externalProvidersNegWI,
		Steps: []resource.TestStep{
			{
				PlanOnly: true,
				Config: `
					resource "rafay_aks_workload_identity" "test" {
					  metadata {
					    name        = "tf-neg-wi-null-spec-name"
					    cluster_name = "test-cluster"
					    project     = "default"
					  }
					  
					  spec {
					    metadata {
					      name           = null              # null
					      resource_group = "test-rg"
					      tenant_id      = "12345678-1234-1234-1234-123456789012"
					    }
					  }
					}
				`,
				ExpectError: regexp.MustCompile(`(?s)Missing required argument.*"spec\.0\.metadata\.0\.name" is required`),
			},
		},
	})
}

// ---------- invalid role scope ----------

func TestAccNegAKSWorkloadIdentity_InvalidRoleScope_Error(t *testing.T) {
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
					    name        = "tf-neg-wi-invalid-scope"
					    cluster_name = "test-cluster"
					    project     = "default"
					  }
					  
					  spec {
					    metadata {
					      name           = "tf-neg-wi-invalid-scope"
					      resource_group = "test-rg"
					      tenant_id      = "12345678-1234-1234-1234-123456789012"
					    }
					    
					    role_assignments {
					      scope                = "invalid-scope-format"   # Invalid scope
					      role_definition_name = "Reader"
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
					    name        = "tf-neg-wi-invalid-ns"
					    cluster_name = "test-cluster"
					    project     = "default"
					  }
					  
					  spec {
					    metadata {
					      name           = "tf-neg-wi-invalid-ns"
					      resource_group = "test-rg"
					      tenant_id      = "12345678-1234-1234-1234-123456789012"
					    }
					    
					    service_accounts {
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
