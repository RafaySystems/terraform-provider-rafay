//go:build !planonly
// +build !planonly

package rafay_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// Use the released provider from the Terraform Registry (black-box tests).
var externalProvidersNeg = map[string]resource.ExternalProvider{
	"rafay": {Source: "RafaySystems/rafay"},
}

// ---------- cluster.metadata.name ----------

// Empty string is treated as "present", so plan is non-empty.
func TestAccNegEKSCluster_EmptyClusterName_AllowsPlan(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		ExternalProviders: externalProvidersNeg,
		Steps: []resource.TestStep{
			{
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				Config: `
					resource "rafay_eks_cluster" "test" {
					  cluster {
					    kind = "Cluster"
					    metadata {
					      name    = ""             # empty
					      project = "default"
					    }
					    spec {
					      blueprint      = "default"
					      cloud_provider = "AWS"
					      cni_provider   = "aws-cni"
					      type           = "aws-eks"
					    }
					  }
					  cluster_config {
					    apiversion = "rafay.io/v1alpha5"
					    kind       = "ClusterConfig"
					    metadata {
					      name    = "tf-neg-empty-name"
					      region  = "us-west-2"
					      version = "1.20"
					    }
					  }
					}
				`,
			},
		},
	})
}

// null is treated as missing; expect "Missing required argument".
func TestAccNegEKSCluster_NullClusterName_RequiredError(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		ExternalProviders: externalProvidersNeg,
		Steps: []resource.TestStep{
			{
				PlanOnly: true,
				Config: `
					resource "rafay_eks_cluster" "test" {
					  cluster {
					    kind = "Cluster"
					    metadata {
					      name    = null           # null
					      project = "default"
					    }
					    spec {
					      blueprint      = "default"
					      cloud_provider = "AWS"
					      cni_provider   = "aws-cni"
					      type           = "aws-eks"
					    }
					  }
					  cluster_config {
					    apiversion = "rafay.io/v1alpha5"
					    kind       = "ClusterConfig"
					    metadata {
					      name    = "tf-neg-null-name"
					      region  = "us-west-2"
					      version = "1.20"
					    }
					  }
					}
				`,
				ExpectError: regexp.MustCompile(`(?s)Missing required argument.*"cluster\.0\.metadata\.0\.name" is required`),
			},
		},
	})
}

// ---------- cluster.metadata.project ----------

// Empty string -> present -> plan proceeds.
func TestAccNegEKSCluster_EmptyProject_AllowsPlan(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		ExternalProviders: externalProvidersNeg,
		Steps: []resource.TestStep{
			{
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				Config: `
					resource "rafay_eks_cluster" "test" {
					  cluster {
					    kind = "Cluster"
					    metadata {
					      name    = "tf-neg-empty-project"
					      project = ""            # empty
					    }
					    spec {
					      blueprint      = "default"
					      cloud_provider = "AWS"
					      cni_provider   = "aws-cni"
					      type           = "aws-eks"
					    }
					  }
					  cluster_config {
					    apiversion = "rafay.io/v1alpha5"
					    kind       = "ClusterConfig"
					    metadata {
					      name    = "tf-neg-empty-project"
					      region  = "us-west-2"
					      version = "1.20"
					    }
					  }
					}
				`,
			},
		},
	})
}

// ---------- cluster.spec.cloud_provider ----------

// Empty string -> present -> plan proceeds.
func TestAccNegEKSCluster_EmptyCloudProvider_AllowsPlan(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		ExternalProviders: externalProvidersNeg,
		Steps: []resource.TestStep{
			{
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				Config: `
					resource "rafay_eks_cluster" "test" {
					  cluster {
					    kind = "Cluster"
					    metadata {
					      name    = "tf-neg-empty-cloudp"
					      project = "default"
					    }
					    spec {
					      blueprint      = "default"
					      cloud_provider = ""       # empty (still present)
					      cni_provider   = "aws-cni"
					      type           = "aws-eks"
					    }
					  }
					  cluster_config {
					    apiversion = "rafay.io/v1alpha5"
					    kind       = "ClusterConfig"
					    metadata {
					      name    = "tf-neg-empty-cloudp"
					      region  = "us-west-2"
					      version = "1.20"
					    }
					  }
					}
				`,
			},
		},
	})
}

// null -> missing -> required error.
func TestAccNegEKSCluster_NullCloudProvider_RequiredError(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		ExternalProviders: externalProvidersNeg,
		Steps: []resource.TestStep{
			{
				PlanOnly: true,
				Config: `
					resource "rafay_eks_cluster" "test" {
					  cluster {
					    kind = "Cluster"
					    metadata {
					      name    = "tf-neg-null-cloudp"
					      project = "default"
					    }
					    spec {
					      blueprint      = "default"
					      cloud_provider = null     # null
					      cni_provider   = "aws-cni"
					      type           = "aws-eks"
					    }
					  }
					  cluster_config {
					    apiversion = "rafay.io/v1alpha5"
					    kind       = "ClusterConfig"
					    metadata {
					      name    = "tf-neg-null-cloudp"
					      region  = "us-west-2"
					      version = "1.20"
					    }
					  }
					}
				`,
				ExpectError: regexp.MustCompile(`(?s)Missing required argument.*"cluster\.0\.spec\.0\.cloud_provider" is required`),
			},
		},
	})
}

// ---------- cluster_config.metadata.region ----------

// Empty string -> present -> plan proceeds.
func TestAccNegEKSCluster_EmptyRegion_AllowsPlan(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		ExternalProviders: externalProvidersNeg,
		Steps: []resource.TestStep{
			{
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				Config: `
					resource "rafay_eks_cluster" "test" {
					  cluster {
					    kind = "Cluster"
					    metadata {
					      name    = "tf-neg-empty-region"
					      project = "default"
					    }
					    spec {
					      blueprint      = "default"
					      cloud_provider = "AWS"
					      cni_provider   = "aws-cni"
					      type           = "aws-eks"
					    }
					  }
					  cluster_config {
					    apiversion = "rafay.io/v1alpha5"
					    kind       = "ClusterConfig"
					    metadata {
					      name    = "tf-neg-empty-region"
					      region  = ""           # empty
					      version = "1.20"
					    }
					  }
					}
				`,
			},
		},
	})
}

// null -> missing -> required error.
func TestAccNegEKSCluster_NullRegion_RequiredError(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		ExternalProviders: externalProvidersNeg,
		Steps: []resource.TestStep{
			{
				PlanOnly: true,
				Config: `
					resource "rafay_eks_cluster" "test" {
					  cluster {
					    kind = "Cluster"
					    metadata {
					      name    = "tf-neg-null-region"
					      project = "default"
					    }
					    spec {
					      blueprint      = "default"
					      cloud_provider = "AWS"
					      cni_provider   = "aws-cni"
					      type           = "aws-eks"
					    }
					  }
					  cluster_config {
					    apiversion = "rafay.io/v1alpha5"
					    kind       = "ClusterConfig"
					    metadata {
					      name    = "tf-neg-null-region"
					      region  = null         # null
					      version = "1.20"
					    }
					  }
					}
				`,
				ExpectError: regexp.MustCompile(`(?s)Missing required argument.*"cluster_config\.0\.metadata\.0\.region" is required`),
			},
		},
	})
}

// ---------- cluster_config.kubernetes_network_config ----------

// Empty optional field -> plan proceeds.
func TestAccNegEKSCluster_EmptyServiceCIDR_AllowsPlan(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		ExternalProviders: externalProvidersNeg,
		Steps: []resource.TestStep{
			{
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				Config: `
					resource "rafay_eks_cluster" "test" {
					  cluster {
					    kind = "Cluster"
					    metadata {
					      name    = "tf-neg-empty-svc-cidr"
					      project = "default"
					    }
					    spec {
					      blueprint      = "default"
					      cloud_provider = "AWS"
					      cni_provider   = "aws-cni"
					      type           = "aws-eks"
					    }
					  }
					  cluster_config {
					    apiversion = "rafay.io/v1alpha5"
					    kind       = "ClusterConfig"
					    kubernetes_network_config {
					      ip_family         = "IPv4"
					      service_ipv4_cidr = ""   # empty; optional
					    }
					    metadata {
					      name    = "tf-neg-empty-svc-cidr"
					      region  = "us-west-2"
					      version = "1.20"
					    }
					  }
					}
				`,
			},
		},
	})
}

// Null optional field -> plan proceeds (defaults to IPv4).
func TestAccNegEKSCluster_NullIPFamily_AllowsPlan(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		ExternalProviders: externalProvidersNeg,
		Steps: []resource.TestStep{
			{
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				Config: `
					resource "rafay_eks_cluster" "test" {
					  cluster {
					    kind = "Cluster"
					    metadata {
					      name    = "tf-neg-null-ipfam"
					      project = "default"
					    }
					    spec {
					      blueprint      = "default"
					      cloud_provider = "AWS"
					      cni_provider   = "aws-cni"
					      type           = "aws-eks"
					    }
					  }
					  cluster_config {
					    apiversion = "rafay.io/v1alpha5"
					    kind       = "ClusterConfig"
					    kubernetes_network_config {
					      ip_family = null        # null; optional
					    }
					    metadata {
					      name    = "tf-neg-null-ipfam"
					      region  = "us-west-2"
					      version = "1.20"
					    }
					  }
					}
				`,
			},
		},
	})
}

// ---------- cluster_config.access_config ----------

// Empty/omitted optional -> plan proceeds.
func TestAccNegEKSCluster_EmptyAuthMode_AllowsPlan(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		ExternalProviders: externalProvidersNeg,
		Steps: []resource.TestStep{
			{
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				Config: `
					resource "rafay_eks_cluster" "test" {
					  cluster {
					    kind = "Cluster"
					    metadata {
					      name    = "tf-neg-empty-authmode"
					      project = "default"
					    }
					    spec {
					      blueprint      = "default"
					      cloud_provider = "AWS"
					      cni_provider   = "aws-cni"
					      type           = "aws-eks"
					    }
					  }
					  cluster_config {
					    apiversion = "rafay.io/v1alpha5"
					    kind       = "ClusterConfig"
					    access_config {
					      authentication_mode = ""  # empty; optional in schema
					    }
					    metadata {
					      name    = "tf-neg-empty-authmode"
					      region  = "us-west-2"
					      version = "1.20"
					    }
					  }
					}
				`,
			},
		},
	})
}
