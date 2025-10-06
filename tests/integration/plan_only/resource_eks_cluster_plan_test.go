//go:build planonly
// +build planonly

package rafay_test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// Use the released provider from the Terraform Registry during tests.
var externalProviders = map[string]resource.ExternalProvider{
	"rafay": {Source: "RafaySystems/rafay"},
}

// Minimal env so provider Configure() is happy in plan-only runs.
func setDummyEnv(t *testing.T) {
	_ = os.Setenv("RCTL_API_KEY", "dummy")
	_ = os.Setenv("RCTL_PROJECT", "default")
	_ = os.Setenv("RCTL_REST_ENDPOINT", "console.example.dev")
}

// ------------------ BASIC ------------------

func TestAccEKSCluster_Basic_PlanOnly(t *testing.T) {
	setDummyEnv(t)

	resource.Test(t, resource.TestCase{
		ExternalProviders: externalProviders,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "rafay" {}

					resource "rafay_eks_cluster" "test" {
					  cluster {
					    metadata {
					      name    = "tf-planonly-basic"
					      project = "default"
					    }
					    spec {
					      cloud_provider = "AWS"
					      type           = "aws-eks"
					      blueprint      = "default"
					      cni_provider   = "aws-cni"
					    }
					  }

					  cluster_config {
					    metadata {
					      name    = "tf-planonly-basic"
					      region  = "us-west-2"
					      version = "1.20"
					    }
					  }
					}
				`,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rafay_eks_cluster.test", "cluster.0.metadata.0.name", "tf-planonly-basic"),
					resource.TestCheckResourceAttr("rafay_eks_cluster.test", "cluster.0.metadata.0.project", "default"),
					resource.TestCheckResourceAttr("rafay_eks_cluster.test", "cluster.0.spec.0.cloud_provider", "AWS"),
					resource.TestCheckResourceAttr("rafay_eks_cluster.test", "cluster.0.spec.0.type", "aws-eks"),
					resource.TestCheckResourceAttr("rafay_eks_cluster.test", "cluster.0.spec.0.blueprint", "default"),
					resource.TestCheckResourceAttr("rafay_eks_cluster.test", "cluster.0.spec.0.cni_provider", "aws-cni"),
					resource.TestCheckResourceAttr("rafay_eks_cluster.test", "cluster_config.0.metadata.0.region", "us-west-2"),
					resource.TestCheckResourceAttr("rafay_eks_cluster.test", "cluster_config.0.metadata.0.version", "1.20"),
				),
			},
		},
	})
}

// ------------------ NODE GROUPS (UNMANAGED) ------------------

func TestAccEKSCluster_NodeGroups_Defaults_PlanOnly(t *testing.T) {
	setDummyEnv(t)

	resource.Test(t, resource.TestCase{
		ExternalProviders: externalProviders,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "rafay" {}

					resource "rafay_eks_cluster" "test" {
					  cluster {
					    metadata {
					      name    = "tf-planonly-ng"
					      project = "default"
					    }
					    spec {
					      cloud_provider = "AWS"
					      type           = "aws-eks"
					      blueprint      = "default"
					    }
					  }

					  cluster_config {
					    metadata {
					      name   = "tf-planonly-ng"
					      region = "us-west-2"
					    }

					    node_groups {
					      name             = "ng1"
					      instance_type    = "t3.medium"
					      desired_capacity = 1
					      # Expect volume_* defaults to be set in plan
					    }
					  }
					}
				`,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rafay_eks_cluster.test", "cluster_config.0.node_groups.0.volume_size", "80"),
					resource.TestCheckResourceAttr("rafay_eks_cluster.test", "cluster_config.0.node_groups.0.volume_type", "gp3"),
					resource.TestCheckResourceAttr("rafay_eks_cluster.test", "cluster_config.0.node_groups.0.volume_iops", "3000"),
					resource.TestCheckResourceAttr("rafay_eks_cluster.test", "cluster_config.0.node_groups.0.volume_throughput", "125"),
				),
			},
		},
	})
}

// ------------------ MANAGED NODE GROUPS ------------------

func TestAccEKSCluster_ManagedNodeGroups_PlanOnly(t *testing.T) {
	setDummyEnv(t)

	resource.Test(t, resource.TestCase{
		ExternalProviders: externalProviders,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "rafay" {}

					resource "rafay_eks_cluster" "test" {
					  cluster {
					    metadata {
					      name    = "tf-planonly-mng"
					      project = "default"
					    }
					    spec {
					      cloud_provider = "AWS"
					      type           = "aws-eks"
					      blueprint      = "default"
					    }
					  }

					  cluster_config {
					    metadata {
					      name   = "tf-planonly-mng"
					      region = "us-west-2"
					    }

					    managed_nodegroups {
					      name             = "mng1"
					      instance_types   = ["t3.medium"]
					      desired_capacity = 1
					    }
					  }
					}
				`,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rafay_eks_cluster.test", "cluster_config.0.managed_nodegroups.0.volume_size", "80"),
					resource.TestCheckResourceAttr("rafay_eks_cluster.test", "cluster_config.0.managed_nodegroups.0.volume_type", "gp3"),
					resource.TestCheckResourceAttr("rafay_eks_cluster.test", "cluster_config.0.managed_nodegroups.0.volume_iops", "3000"),
					resource.TestCheckResourceAttr("rafay_eks_cluster.test", "cluster_config.0.managed_nodegroups.0.volume_throughput", "125"),
				),
			},
		},
	})
}

// ------------------ VPC / PRIVATE / K8S NETWORK ------------------

func TestAccEKSCluster_VPC_Private_PlanOnly(t *testing.T) {
	setDummyEnv(t)

	resource.Test(t, resource.TestCase{
		ExternalProviders: externalProviders,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "rafay" {}

					resource "rafay_eks_cluster" "test" {
					  cluster {
					    metadata {
					      name    = "tf-planonly-priv"
					      project = "default"
					    }
					    spec {
					      cloud_provider = "AWS"
					      type           = "aws-eks"
					      blueprint      = "default"
					    }
					  }

					  cluster_config {
					    metadata {
					      name   = "tf-planonly-priv"
					      region = "us-west-2"
					    }

					    kubernetes_network_config {
					      ip_family         = "IPv4"
					      service_ipv4_cidr = "172.20.0.0/16"
					    }

					    vpc {
					      cluster_endpoints {
					        private_access = true
					        public_access  = false
					      }
					    }

					    private_cluster {
					      enabled                = true
					      skip_endpoint_creation = false
					    }
					  }
					}
				`,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rafay_eks_cluster.test", "cluster_config.0.kubernetes_network_config.0.ip_family", "IPv4"),
					resource.TestCheckResourceAttr("rafay_eks_cluster.test", "cluster_config.0.kubernetes_network_config.0.service_ipv4_cidr", "172.20.0.0/16"),
					resource.TestCheckResourceAttr("rafay_eks_cluster.test", "cluster_config.0.vpc.0.cluster_endpoints.0.private_access", "true"),
					resource.TestCheckResourceAttr("rafay_eks_cluster.test", "cluster_config.0.vpc.0.cluster_endpoints.0.public_access", "false"),
					resource.TestCheckResourceAttr("rafay_eks_cluster.test", "cluster_config.0.private_cluster.0.enabled", "true"),
					resource.TestCheckResourceAttr("rafay_eks_cluster.test", "cluster_config.0.private_cluster.0.skip_endpoint_creation", "false"),
				),
			},
		},
	})
}

// ------------------ IAM / OIDC / SERVICE ACCOUNTS ------------------

func TestAccEKSCluster_IAM_OIDC_ServiceAccounts_PlanOnly(t *testing.T) {
	setDummyEnv(t)

	resource.Test(t, resource.TestCase{
		ExternalProviders: externalProviders,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "rafay" {}

					resource "rafay_eks_cluster" "test" {
					  cluster {
					    metadata {
					      name    = "tf-planonly-iam"
					      project = "default"
					    }
					    spec {
					      cloud_provider = "AWS"
					      type           = "aws-eks"
					      blueprint      = "default"
					    }
					  }

					  cluster_config {
					    metadata {
					      name   = "tf-planonly-iam"
					      region = "us-west-2"
					    }

					    iam {
					      with_oidc = true

					      service_accounts {
					        metadata {
					          name      = "sa-logs"
					          namespace = "kube-system"
					        }
					        attach_role_arn = "arn:aws:iam::111111111111:role/example"
					      }

					      pod_identity_associations {
					        namespace              = "kube-system"
					        service_account_name   = "sa-logs"
					        role_arn               = "arn:aws:iam::111111111111:role/example"
					        create_service_account = false
					      }
					    }
					  }
					}
				`,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rafay_eks_cluster.test", "cluster_config.0.iam.0.with_oidc", "true"),
					resource.TestCheckResourceAttr("rafay_eks_cluster.test", "cluster_config.0.iam.0.service_accounts.0.metadata.0.name", "sa-logs"),
					resource.TestCheckResourceAttr("rafay_eks_cluster.test", "cluster_config.0.iam.0.service_accounts.0.metadata.0.namespace", "kube-system"),
					resource.TestCheckResourceAttr("rafay_eks_cluster.test", "cluster_config.0.iam.0.pod_identity_associations.0.namespace", "kube-system"),
					resource.TestCheckResourceAttr("rafay_eks_cluster.test", "cluster_config.0.iam.0.pod_identity_associations.0.service_account_name", "sa-logs"),
					resource.TestCheckResourceAttr("rafay_eks_cluster.test", "cluster_config.0.iam.0.pod_identity_associations.0.create_service_account", "false"),
				),
			},
		},
	})
}

// ------------------ ACCESS CONFIG ------------------

func TestAccEKSCluster_AccessConfig_PlanOnly(t *testing.T) {
	setDummyEnv(t)

	resource.Test(t, resource.TestCase{
		ExternalProviders: externalProviders,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "rafay" {}

					resource "rafay_eks_cluster" "test" {
					  cluster {
					    metadata {
					      name    = "tf-planonly-access"
					      project = "default"
					    }
					    spec {
					      cloud_provider = "AWS"
					      type           = "aws-eks"
					      blueprint      = "default"
					    }
					  }

					  cluster_config {
					    metadata {
					      name   = "tf-planonly-access"
					      region = "us-west-2"
					    }

					    access_config {
					      bootstrap_cluster_creator_admin_permissions = true
					      authentication_mode = "API_AND_CONFIG_MAP"

					      access_entries {
					        principal_arn = "arn:aws:iam::111111111111:user/example"
					        type          = "USER"

					        access_policies {
					          policy_arn = "arn:aws:eks::aws:cluster-access-policy/AmazonEKSAdminPolicy"
					          access_scope {
					            type       = "namespace"
					            namespaces = ["kube-system"]
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
					resource.TestCheckResourceAttr("rafay_eks_cluster.test", "cluster_config.0.access_config.0.bootstrap_cluster_creator_admin_permissions", "true"),
					resource.TestCheckResourceAttr("rafay_eks_cluster.test", "cluster_config.0.access_config.0.authentication_mode", "API_AND_CONFIG_MAP"),
					resource.TestCheckResourceAttr("rafay_eks_cluster.test", "cluster_config.0.access_config.0.access_entries.0.type", "USER"),
					resource.TestCheckResourceAttr("rafay_eks_cluster.test", "cluster_config.0.access_config.0.access_entries.0.access_policies.0.access_scope.0.type", "namespace"),
					resource.TestCheckResourceAttr("rafay_eks_cluster.test", "cluster_config.0.access_config.0.access_entries.0.access_policies.0.access_scope.0.namespaces.0", "kube-system"),
				),
			},
		},
	})
}

// ------------------ ADDONS ------------------

func TestAccEKSCluster_Addons_PlanOnly(t *testing.T) {
	setDummyEnv(t)

	resource.Test(t, resource.TestCase{
		ExternalProviders: externalProviders,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "rafay" {}

					resource "rafay_eks_cluster" "test" {
					  cluster {
					    metadata {
					      name    = "tf-planonly-addons"
					      project = "default"
					    }
					    spec {
					      cloud_provider = "AWS"
					      type           = "aws-eks"
					      blueprint      = "default"
					    }
					  }

					  cluster_config {
					    metadata {
					      name   = "tf-planonly-addons"
					      region = "us-west-2"
					    }

					    addons {
					      name                 = "vpc-cni"
					      version              = "v1.15.3-eksbuild.1"
					      configuration_values = "{\"enablePodENI\": true}"

					      pod_identity_associations {
					        namespace              = "kube-system"
					        service_account_name   = "aws-node"
					        role_arn               = "arn:aws:iam::111111111111:role/example"
					        create_service_account = false
					      }
					    }
					  }
					}
				`,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rafay_eks_cluster.test", "cluster_config.0.addons.0.name", "vpc-cni"),
					resource.TestCheckResourceAttr("rafay_eks_cluster.test", "cluster_config.0.addons.0.version", "v1.15.3-eksbuild.1"),
					resource.TestCheckResourceAttr("rafay_eks_cluster.test", "cluster_config.0.addons.0.configuration_values", "{\"enablePodENI\": true}"),
					resource.TestCheckResourceAttr("rafay_eks_cluster.test", "cluster_config.0.addons.0.pod_identity_associations.0.namespace", "kube-system"),
					resource.TestCheckResourceAttr("rafay_eks_cluster.test", "cluster_config.0.addons.0.pod_identity_associations.0.service_account_name", "aws-node"),
				),
			},
		},
	})
}
