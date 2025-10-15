//go:build planonly
// +build planonly

package rafay_test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// Use the released provider from the Terraform Registry during tests.
var externalProvidersAKSSpec = map[string]resource.ExternalProvider{
	"rafay": {Source: "RafaySystems/rafay"},
}

// Minimal env so provider Configure() is happy in plan-only runs.
func setDummyEnvAKSSpec(t *testing.T) {
	_ = os.Setenv("RCTL_API_KEY", "dummy")
	_ = os.Setenv("RCTL_PROJECT", "default")
	_ = os.Setenv("RCTL_REST_ENDPOINT", "console.example.dev")
}

// ------------------ BASIC ------------------

func TestAccAKSClusterSpec_Basic_PlanOnly(t *testing.T) {
	setDummyEnvAKSSpec(t)

	resource.Test(t, resource.TestCase{
		ExternalProviders: externalProvidersAKSSpec,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "rafay" {}

					resource "rafay_aks_cluster_spec" "test" {
					  yamlfilepath    = "/tmp/aks-cluster.yaml"
					  yamlfileversion = "v1"
					  name            = "tf-planonly-spec"
					  projectname     = "default"
					}
				`,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rafay_aks_cluster_spec.test", "yamlfilepath", "/tmp/aks-cluster.yaml"),
					resource.TestCheckResourceAttr("rafay_aks_cluster_spec.test", "yamlfileversion", "v1"),
					resource.TestCheckResourceAttr("rafay_aks_cluster_spec.test", "name", "tf-planonly-spec"),
					resource.TestCheckResourceAttr("rafay_aks_cluster_spec.test", "projectname", "default"),
				),
			},
		},
	})
}

// ------------------ WITH WAIT FLAG ------------------

func TestAccAKSClusterSpec_WithWaitFlag_PlanOnly(t *testing.T) {
	setDummyEnvAKSSpec(t)

	resource.Test(t, resource.TestCase{
		ExternalProviders: externalProvidersAKSSpec,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "rafay" {}

					resource "rafay_aks_cluster_spec" "test" {
					  yamlfilepath    = "/tmp/aks-cluster.yaml"
					  yamlfileversion = "v1"
					  name            = "tf-planonly-wait"
					  projectname     = "default"
					  waitflag        = "0"
					}
				`,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rafay_aks_cluster_spec.test", "name", "tf-planonly-wait"),
					resource.TestCheckResourceAttr("rafay_aks_cluster_spec.test", "waitflag", "0"),
				),
			},
			{
				Config: `
					provider "rafay" {}

					resource "rafay_aks_cluster_spec" "test" {
					  yamlfilepath    = "/tmp/aks-cluster.yaml"
					  yamlfileversion = "v1"
					  name            = "tf-planonly-wait-default"
					  projectname     = "default"
					  # waitflag not specified, should default to "1"
					}
				`,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rafay_aks_cluster_spec.test", "name", "tf-planonly-wait-default"),
					resource.TestCheckResourceAttr("rafay_aks_cluster_spec.test", "waitflag", "1"),
				),
			},
		},
	})
}

// ------------------ WITH CHECK DIFF ------------------

func TestAccAKSClusterSpec_WithCheckDiff_PlanOnly(t *testing.T) {
	setDummyEnvAKSSpec(t)

	resource.Test(t, resource.TestCase{
		ExternalProviders: externalProvidersAKSSpec,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "rafay" {}

					resource "rafay_aks_cluster_spec" "test" {
					  yamlfilepath    = "/tmp/aks-cluster.yaml"
					  yamlfileversion = "v1"
					  name            = "tf-planonly-checkdiff"
					  projectname     = "default"
					  checkdiff       = true
					}
				`,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rafay_aks_cluster_spec.test", "name", "tf-planonly-checkdiff"),
					resource.TestCheckResourceAttr("rafay_aks_cluster_spec.test", "checkdiff", "true"),
				),
			},
			{
				Config: `
					provider "rafay" {}

					resource "rafay_aks_cluster_spec" "test" {
					  yamlfilepath    = "/tmp/aks-cluster.yaml"
					  yamlfileversion = "v1"
					  name            = "tf-planonly-no-checkdiff"
					  projectname     = "default"
					  checkdiff       = false
					}
				`,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rafay_aks_cluster_spec.test", "name", "tf-planonly-no-checkdiff"),
					resource.TestCheckResourceAttr("rafay_aks_cluster_spec.test", "checkdiff", "false"),
				),
			},
		},
	})
}
