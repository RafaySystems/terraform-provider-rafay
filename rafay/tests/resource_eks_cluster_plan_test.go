//go:build planonly
// +build planonly

package rafay_test

import (
	"os"
	"testing"

	"github.com/RafaySystems/terraform-provider-rafay/rafay"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Use the released provider from the Terraform Registry during tests.
var externalProviders = map[string]resource.ExternalProvider{
	"rafay": {Source: "RafaySystems/rafay"},
}

var localProviderFactories = map[string]func() (*schema.Provider, error){
	"rafay": func() (*schema.Provider, error) {
		return rafay.New("test")(), nil
	},
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

// ------------------ ORDERING (LOCAL PROVIDER FACTORY) ------------------

func TestAccEKSCluster_ManagedNodeGroups_Order_PlanOnly_Local(t *testing.T) {
	setDummyEnv(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories: localProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "rafay" {}

					resource "rafay_eks_cluster" "test" {
					  cluster {
					    metadata {
					      name    = "tf-planonly-mng-order"
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
					      name   = "tf-planonly-mng-order"
					      region = "us-west-2"
					    }

					    managed_nodegroups {
					      name           = "zz-mng"
					      instance_types = ["m6a.large", "m5a.large", "m7i.large"]
					      security_groups {
					        attach_ids = ["sg-3", "sg-1", "sg-2"]
					      }
					      taints {
					        key    = "b"
					        effect = "NoSchedule"
					        value  = "two"
					      }
					      taints {
					        key    = "a"
					        effect = "NoExecute"
					        value  = "one"
					      }
					    }

					    managed_nodegroups {
					      name = "aa-mng"
					    }
					  }
					}
				`,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rafay_eks_cluster.test", "cluster_config.0.managed_nodegroups.0.name", "aa-mng"),
					resource.TestCheckResourceAttr("rafay_eks_cluster.test", "cluster_config.0.managed_nodegroups.1.name", "zz-mng"),
					resource.TestCheckResourceAttr("rafay_eks_cluster.test", "cluster_config.0.managed_nodegroups.1.instance_types.0", "m5a.large"),
					resource.TestCheckResourceAttr("rafay_eks_cluster.test", "cluster_config.0.managed_nodegroups.1.security_groups.0.attach_ids.0", "sg-1"),
					resource.TestCheckResourceAttr("rafay_eks_cluster.test", "cluster_config.0.managed_nodegroups.1.taints.0.key", "a"),
				),
			},
		},
	})
}

func TestAccEKSCluster_NodeGroups_Order_PlanOnly_Local(t *testing.T) {
	setDummyEnv(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories: localProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "rafay" {}

					resource "rafay_eks_cluster" "test" {
					  cluster {
					    metadata {
					      name    = "tf-planonly-ng-order"
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
					      name   = "tf-planonly-ng-order"
					      region = "us-west-2"
					    }

					    node_groups {
					      name          = "zz-ng"
					      instance_type = "t3.medium"
					      instances_distribution {
					        instance_types = ["m6a.large", "m5a.large", "m7i.large"]
					      }
					      security_groups {
					        attach_ids = ["sg-3", "sg-1", "sg-2"]
					      }
					      taints {
					        key    = "b"
					        effect = "NoSchedule"
					        value  = "two"
					      }
					      taints {
					        key    = "a"
					        effect = "NoExecute"
					        value  = "one"
					      }
					    }

					    node_groups {
					      name          = "aa-ng"
					      instance_type = "t3.medium"
					    }
					  }
					}
				`,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rafay_eks_cluster.test", "cluster_config.0.node_groups.0.name", "aa-ng"),
					resource.TestCheckResourceAttr("rafay_eks_cluster.test", "cluster_config.0.node_groups.1.name", "zz-ng"),
					resource.TestCheckResourceAttr("rafay_eks_cluster.test", "cluster_config.0.node_groups.1.instances_distribution.0.instance_types.0", "m5a.large"),
					resource.TestCheckResourceAttr("rafay_eks_cluster.test", "cluster_config.0.node_groups.1.security_groups.0.attach_ids.0", "sg-1"),
					resource.TestCheckResourceAttr("rafay_eks_cluster.test", "cluster_config.0.node_groups.1.taints.0.key", "a"),
				),
			},
		},
	})
}

// ------------------ ORDERING / CANONICALIZATION ------------------

func TestAccEKSCluster_ManagedNodeGroups_Order_PlanOnly(t *testing.T) {
	setDummyEnv(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories: localProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "rafay" {}

					resource "rafay_eks_cluster" "test" {
					  cluster {
					    metadata {
					      name    = "tf-planonly-mng-order"
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
					      name   = "tf-planonly-mng-order"
					      region = "us-west-2"
					    }

					    # Intentionally out of order
					    managed_nodegroups {
					      name           = "b-ng"
					      instance_types = ["m5a.large", "m7i.large", "m6a.large"]
					      security_groups {
					        attach_ids = ["sg-3", "sg-1", "sg-2"]
					      }
					      taints {
					        key    = "b"
					        effect = "NoSchedule"
					        value  = "two"
					      }
					      taints {
					        key    = "a"
					        effect = "NoExecute"
					        value  = "one"
					      }
					    }
					    managed_nodegroups {
					      name           = "a-ng"
					      instance_types = ["t3.medium"]
					    }
					  }
					}
				`,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					// list ordering by name
					resource.TestCheckResourceAttr("rafay_eks_cluster.test", "cluster_config.0.managed_nodegroups.0.name", "a-ng"),
					resource.TestCheckResourceAttr("rafay_eks_cluster.test", "cluster_config.0.managed_nodegroups.1.name", "b-ng"),
					// nested list ordering
					resource.TestCheckResourceAttr("rafay_eks_cluster.test", "cluster_config.0.managed_nodegroups.1.instance_types.0", "m5a.large"),
					resource.TestCheckResourceAttr("rafay_eks_cluster.test", "cluster_config.0.managed_nodegroups.1.instance_types.1", "m6a.large"),
					resource.TestCheckResourceAttr("rafay_eks_cluster.test", "cluster_config.0.managed_nodegroups.1.instance_types.2", "m7i.large"),
					resource.TestCheckResourceAttr("rafay_eks_cluster.test", "cluster_config.0.managed_nodegroups.1.security_groups.0.attach_ids.0", "sg-1"),
					resource.TestCheckResourceAttr("rafay_eks_cluster.test", "cluster_config.0.managed_nodegroups.1.security_groups.0.attach_ids.1", "sg-2"),
					resource.TestCheckResourceAttr("rafay_eks_cluster.test", "cluster_config.0.managed_nodegroups.1.security_groups.0.attach_ids.2", "sg-3"),
					// taints ordered by key/effect/value
					resource.TestCheckResourceAttr("rafay_eks_cluster.test", "cluster_config.0.managed_nodegroups.1.taints.0.key", "a"),
					resource.TestCheckResourceAttr("rafay_eks_cluster.test", "cluster_config.0.managed_nodegroups.1.taints.1.key", "b"),
				),
			},
		},
	})
}

func TestAccEKSCluster_NodeGroups_Order_PlanOnly(t *testing.T) {
	setDummyEnv(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories: localProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "rafay" {}

					resource "rafay_eks_cluster" "test" {
					  cluster {
					    metadata {
					      name    = "tf-planonly-ng-order"
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
					      name   = "tf-planonly-ng-order"
					      region = "us-west-2"
					    }

					    node_groups {
					      name = "b-ng"
					      instances_distribution {
					        instance_types = ["m6a.large", "m5a.large", "m7i.large"]
					      }
					      security_groups {
					        attach_ids = ["sg-3", "sg-1", "sg-2"]
					      }
					      taints {
					        key    = "b"
					        effect = "NoSchedule"
					        value  = "two"
					      }
					      taints {
					        key    = "a"
					        effect = "NoExecute"
					        value  = "one"
					      }
					    }
					    node_groups {
					      name = "a-ng"
					    }
					  }
					}
				`,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					// list ordering by name
					resource.TestCheckResourceAttr("rafay_eks_cluster.test", "cluster_config.0.node_groups.0.name", "a-ng"),
					resource.TestCheckResourceAttr("rafay_eks_cluster.test", "cluster_config.0.node_groups.1.name", "b-ng"),
					// nested list ordering
					resource.TestCheckResourceAttr("rafay_eks_cluster.test", "cluster_config.0.node_groups.1.instances_distribution.0.instance_types.0", "m5a.large"),
					resource.TestCheckResourceAttr("rafay_eks_cluster.test", "cluster_config.0.node_groups.1.instances_distribution.0.instance_types.1", "m6a.large"),
					resource.TestCheckResourceAttr("rafay_eks_cluster.test", "cluster_config.0.node_groups.1.instances_distribution.0.instance_types.2", "m7i.large"),
					resource.TestCheckResourceAttr("rafay_eks_cluster.test", "cluster_config.0.node_groups.1.security_groups.0.attach_ids.0", "sg-1"),
					resource.TestCheckResourceAttr("rafay_eks_cluster.test", "cluster_config.0.node_groups.1.security_groups.0.attach_ids.1", "sg-2"),
					resource.TestCheckResourceAttr("rafay_eks_cluster.test", "cluster_config.0.node_groups.1.security_groups.0.attach_ids.2", "sg-3"),
					// taints ordered by key/effect/value
					resource.TestCheckResourceAttr("rafay_eks_cluster.test", "cluster_config.0.node_groups.1.taints.0.key", "a"),
					resource.TestCheckResourceAttr("rafay_eks_cluster.test", "cluster_config.0.node_groups.1.taints.1.key", "b"),
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
