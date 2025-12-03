//go:build !planonly
// +build !planonly

package rafay_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// Use the released provider from the Terraform Registry (black-box tests).
var externalProvidersNegAKSSpec = map[string]resource.ExternalProvider{
	"rafay": {Source: "RafaySystems/rafay"},
}

// ---------- yamlfilepath ----------

// Empty string is treated as "present", so plan is non-empty.
func TestAccNegAKSClusterSpec_EmptyYamlFilePath_AllowsPlan(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		ExternalProviders: externalProvidersNegAKSSpec,
		Steps: []resource.TestStep{
			{
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				Config: `
					resource "rafay_aks_cluster_spec" "test" {
					  yamlfilepath    = ""          # empty
					  yamlfileversion = "v1"
					  name            = "tf-neg-spec-empty-path"
					  projectname     = "default"
					}
				`,
			},
		},
	})
}

// null is treated as missing; expect "Missing required argument".
func TestAccNegAKSClusterSpec_NullYamlFilePath_RequiredError(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		ExternalProviders: externalProvidersNegAKSSpec,
		Steps: []resource.TestStep{
			{
				PlanOnly: true,
				Config: `
					resource "rafay_aks_cluster_spec" "test" {
					  yamlfilepath    = null       # null
					  yamlfileversion = "v1"
					  name            = "tf-neg-spec-null-path"
					  projectname     = "default"
					}
				`,
				ExpectError: regexp.MustCompile(`(?s)Missing required argument.*"yamlfilepath" is required`),
			},
		},
	})
}

// ---------- yamlfileversion ----------

// Empty string -> present -> plan proceeds.
func TestAccNegAKSClusterSpec_EmptyYamlFileVersion_AllowsPlan(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		ExternalProviders: externalProvidersNegAKSSpec,
		Steps: []resource.TestStep{
			{
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				Config: `
					resource "rafay_aks_cluster_spec" "test" {
					  yamlfilepath    = "/tmp/aks-cluster.yaml"
					  yamlfileversion = ""            # empty
					  name            = "tf-neg-spec-empty-version"
					  projectname     = "default"
					}
				`,
			},
		},
	})
}

// null -> missing -> error.
func TestAccNegAKSClusterSpec_NullYamlFileVersion_RequiredError(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		ExternalProviders: externalProvidersNegAKSSpec,
		Steps: []resource.TestStep{
			{
				PlanOnly: true,
				Config: `
					resource "rafay_aks_cluster_spec" "test" {
					  yamlfilepath    = "/tmp/aks-cluster.yaml"
					  yamlfileversion = null            # null
					  name            = "tf-neg-spec-null-version"
					  projectname     = "default"
					}
				`,
				ExpectError: regexp.MustCompile(`(?s)Missing required argument.*"yamlfileversion" is required`),
			},
		},
	})
}

// ---------- name ----------

// Empty string -> present -> plan proceeds.
func TestAccNegAKSClusterSpec_EmptyName_AllowsPlan(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		ExternalProviders: externalProvidersNegAKSSpec,
		Steps: []resource.TestStep{
			{
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				Config: `
					resource "rafay_aks_cluster_spec" "test" {
					  yamlfilepath    = "/tmp/aks-cluster.yaml"
					  yamlfileversion = "v1"
					  name            = ""              # empty
					  projectname     = "default"
					}
				`,
			},
		},
	})
}

// null -> missing -> error.
func TestAccNegAKSClusterSpec_NullName_RequiredError(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		ExternalProviders: externalProvidersNegAKSSpec,
		Steps: []resource.TestStep{
			{
				PlanOnly: true,
				Config: `
					resource "rafay_aks_cluster_spec" "test" {
					  yamlfilepath    = "/tmp/aks-cluster.yaml"
					  yamlfileversion = "v1"
					  name            = null              # null
					  projectname     = "default"
					}
				`,
				ExpectError: regexp.MustCompile(`(?s)Missing required argument.*"name" is required`),
			},
		},
	})
}

// ---------- projectname ----------

// Empty string -> present -> plan proceeds.
func TestAccNegAKSClusterSpec_EmptyProjectName_AllowsPlan(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		ExternalProviders: externalProvidersNegAKSSpec,
		Steps: []resource.TestStep{
			{
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				Config: `
					resource "rafay_aks_cluster_spec" "test" {
					  yamlfilepath    = "/tmp/aks-cluster.yaml"
					  yamlfileversion = "v1"
					  name            = "tf-neg-spec-empty-project"
					  projectname     = ""            # empty
					}
				`,
			},
		},
	})
}

// null -> missing -> error.
func TestAccNegAKSClusterSpec_NullProjectName_RequiredError(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		ExternalProviders: externalProvidersNegAKSSpec,
		Steps: []resource.TestStep{
			{
				PlanOnly: true,
				Config: `
					resource "rafay_aks_cluster_spec" "test" {
					  yamlfilepath    = "/tmp/aks-cluster.yaml"
					  yamlfileversion = "v1"
					  name            = "tf-neg-spec-null-project"
					  projectname     = null            # null
					}
				`,
				ExpectError: regexp.MustCompile(`(?s)Missing required argument.*"projectname" is required`),
			},
		},
	})
}

// ---------- invalid yaml path ----------

func TestAccNegAKSClusterSpec_InvalidYamlPath_Error(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		ExternalProviders: externalProvidersNegAKSSpec,
		Steps: []resource.TestStep{
			{
				// Note: This test validates that invalid path passes plan but would fail on apply
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				Config: `
					resource "rafay_aks_cluster_spec" "test" {
					  yamlfilepath    = "/nonexistent/path/to/file.yaml"   # Invalid path
					  yamlfileversion = "v1"
					  name            = "tf-neg-spec-invalid-path"
					  projectname     = "default"
					}
				`,
			},
		},
	})
}

// ---------- mismatched cluster name ----------

func TestAccNegAKSClusterSpec_MismatchedClusterName_Error(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		ExternalProviders: externalProvidersNegAKSSpec,
		Steps: []resource.TestStep{
			{
				// Note: This test validates that mismatched names pass plan but would fail on apply
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				Config: `
					resource "rafay_aks_cluster_spec" "test" {
					  yamlfilepath    = "/tmp/aks-cluster.yaml"
					  yamlfileversion = "v1"
					  name            = "cluster-name-mismatch"   # Would mismatch YAML content
					  projectname     = "default"
					}
				`,
			},
		},
	})
}
