package rafay

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/RafaySystems/rctl/pkg/cluster"
	"github.com/RafaySystems/rctl/pkg/clusterctl"
	"github.com/RafaySystems/rctl/pkg/config"
	glogger "github.com/RafaySystems/rctl/pkg/log"
	"github.com/RafaySystems/rctl/pkg/project"
	"github.com/davecgh/go-spew/spew"
	"github.com/hashicorp/go-cty/cty"
	jsoniter "github.com/json-iterator/go"
	"k8s.io/utils/strings/slices"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/go-yaml/yaml"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// go:embed resource_eks_cluster_description.md
var resourceEKSClusterDescription string

func resourceEKSCluster() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceEKSClusterCreate,
		ReadContext:   resourceEKSClusterRead,
		UpdateContext: resourceEKSClusterUpdate,
		DeleteContext: resourceEKSClusterDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Importer: &schema.ResourceImporter{
			StateContext: resourceEKSClusterImport,
		},

		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"cluster": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "Rafay specific cluster configuration",
				Elem: &schema.Resource{
					Schema: clusterMetadataField(),
				},
				MinItems: 1,
				MaxItems: 1,
			},
			"cluster_config": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "EKS specific cluster configuration",
				Elem: &schema.Resource{
					Schema: configField(),
				},
				MinItems: 1,
				MaxItems: 1,
			},
		},
		Description: resourceEKSClusterDescription,
	}
}

// schema input for cluster file
func clusterMetadataField() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"kind": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "Cluster",
			Description: "The type of resource. Supported value is `Cluster`.",
		},
		"metadata": {
			Type:        schema.TypeList,
			Required:    true,
			Description: "Contains data that helps uniquely identify the resource.",
			Elem: &schema.Resource{
				Schema: clusterMetaMetadataFields(),
			},
			MinItems: 1,
			MaxItems: 1,
		},
		"spec": {
			Type:        schema.TypeList,
			Required:    true,
			Description: "The specification associated with the cluster, including cluster networking options.",
			Elem: &schema.Resource{
				Schema: specField(),
			},
			MinItems: 1,
			MaxItems: 1,
		},
	}
	return s
}
func clusterMetaMetadataFields() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The name of the EKS cluster in Rafay console. This must be unique in your organization.",
		},
		"project": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The name of the Rafay project the cluster will be created in.",
		},
		"labels": {
			Type:        schema.TypeMap,
			Optional:    true,
			Description: "The labels for the cluster in Rafay console.",
		},
	}
	return s
}
func specField() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"type": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "aws-eks",
			Description: "The cluster type. Supported value is `eks`.",
		},
		"blueprint": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "default",
			Description: "The blueprint associated with the cluster. A blueprint defines the configuration and policy. Use blueprints to help standardize cluster configurations.",
		},
		"blueprint_version": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The blueprint version associated with the cluster.",
		},
		"cloud_provider": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The cloud credentials provider used to create and manage the cluster.",
		},
		"cross_account_role_arn": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Role ARN of the linked account",
		},
		"cni_provider": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "aws-cni",
			Description: "The container network interface (CNI) provider used to specify different network connectivity options for the cluster.",
		},
		"cni_params": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "The container network interface (CNI) parameters.",
			Elem: &schema.Resource{
				Schema: customCniField(),
			},
		},
		"proxy_config": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "The proxy configuration for the cluster. Use this if the infrastructure uses an outbound proxy.",
			Elem: &schema.Resource{
				Schema: proxyConfigFields(),
			},
		},
		"system_components_placement": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "Configure tolerations and nodeSelector for Rafay system components.",
			Elem: &schema.Resource{
				Schema: systemComponentsPlacementFields(),
			},
		},
		"sharing": {
			Description: "The sharing configuration for the resource. A cluster can be shared with one or more projects. Note: If the resource is not shared, set enabled = false.",
			Elem: &schema.Resource{
				Schema: sharingFields(),
			},
			MaxItems: 1,
			Optional: true,
			Type:     schema.TypeList,
		},
	}
	return s
}

func sharingFields() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"enabled": {
			Description: "Enable sharing for this resource.",
			Optional:    true,
			Type:        schema.TypeBool,
		},
		"projects": {
			Description: "The list of projects this resource is shared with. Note: Required when project sharing is enabled.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"name": {
						Description: "The name of the project to share the resource.",
						Required:    true,
						Type:        schema.TypeString,
					},
				},
			},
			Optional: true,
			Type:     schema.TypeList,
		},
	}
}

func systemComponentsPlacementFields() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"node_selector": {
			Type:        schema.TypeMap,
			Optional:    true,
			Description: "used to tag AWS resources created by the vendor",
		},
		"tolerations": {
			Type: schema.TypeList,
			//Type:        schema.TypeString,
			Optional:    true,
			Description: "contains custom cni networking configurations",
			Elem: &schema.Resource{
				Schema: tolerationsFields(),
			},
		},
		"daemonset_override": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "contains custom cni networking configurations",
			Elem: &schema.Resource{
				Schema: daemonsetOverrideFields(),
			},
		},
	}
	return s
}

func tolerationsFields() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"key": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "the taint key that the toleration applies to",
		},
		"operator": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "represents a key's relationship to the value",
		},
		"value": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "the taint value the toleration matches to",
		},
		"effect": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "indicates the taint effect to match",
		},
		"toleration_seconds": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "represents the period of time the toleration tolerates the taint",
		},
	}
	return s
}

func daemonsetOverrideFields() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"node_selection_enabled": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "enables node selection",
		},
		"tolerations": {
			Type: schema.TypeList,
			//Type:        schema.TypeString,
			Optional:    true,
			Description: "contains custom cni networking configurations",
			Elem: &schema.Resource{
				Schema: tolerationsFields(),
			},
		},
		/*
			"tolerations": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "contains custom cni networking configurations",
				Elem: &schema.Resource{
					Schema: tolerationsFields(),
				},
			},*/
	}
	return s
}

func customCniField() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"custom_cni_cidr": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Secondary IPv4 CIDR block for the VPC. This should be specified if you choose to auto-create VPC and subnets while creating the EKS cluster.",
		},
		"custom_cni_crd_spec": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "The custom container network interface custom resource definition specification. One or more of these blocks should be specified if you choose to use your existing VPC and subnets while creating the EKS cluster.",
			Elem: &schema.Resource{
				Schema: customCniSpecField(),
			},
		},
	}
	return s
}

func proxyConfigFields() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"http_proxy": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "http Proxy",
		},
		"https_proxy": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "http Proxy",
		},
		"no_proxy": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "http Proxy",
		},
		"proxy_auth": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "http Proxy",
		},
		"bootstrap_ca": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "http Proxy",
		},
		"enabled": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "http Proxy",
		},
		"allow_insecure_bootstrap": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "http Proxy",
		},
	}

	return s
}

func customCniSpecField() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The name of the Availability Zone (AZ). The availability zone specified here should be a part of the region specified for the EKS cluster.",
		},
		"cni_spec": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "The custom CNI configuration for this AZ.",
			Elem: &schema.Resource{
				Schema: cniSpecField(),
			},
		},
	}
	return s
}

func cniSpecField() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"subnet": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The subnet associated with secondary ENIs for AWS EC2 nodes.",
		},
		"security_groups": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "The security groups associated with secondary ENIs for AWS EC2 nodes.",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
	}
	return s
}

// schema input for cluster config file
func configField() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"kind": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "ClusterConfig",
			Description: "kind",
		},
		"apiversion": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "rafay.io/v1alpha5",
			Description: "apiversion",
		},
		"metadata": {
			Type:        schema.TypeList,
			Required:    true,
			Description: "contains cluster networking options",
			Elem: &schema.Resource{
				Schema: configMetadataField(),
			},
		},
		"kubernetes_network_config": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "contains cluster networking options",
			Elem: &schema.Resource{
				Schema: kubernetesNetworkConfigField(),
			},
		},
		"iam": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "holds all IAM attributes of a cluster",
			Elem: &schema.Resource{
				Schema: iamFields(),
			},
		},
		"identity_providers": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "holds an identity provider configuration.",
			Elem: &schema.Resource{
				Schema: identityProviderField(),
			},
		},
		"vpc": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "holds global subnet and all child subnets",
			Elem: &schema.Resource{
				Schema: vpcFields(),
			},
		},
		"addons": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "holds the EKS addon configuration",
			Elem: &schema.Resource{
				Schema: addonConfigFields(),
			},
		},
		"private_cluster": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "allows configuring a fully-private cluster in which no node has outbound internet access, and private access to AWS services is enabled via VPC endpoints",
			Elem: &schema.Resource{
				Schema: privateClusterConfigFields(),
			},
		},
		"node_groups": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "holds all nodegroup attributes of a cluster.",
			Elem: &schema.Resource{
				Schema: nodeGroupsConfigFields(),
			},
		},
		"managed_nodegroups": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "holds all nodegroup attributes of a cluster.",
			Elem: &schema.Resource{
				Schema: managedNodeGroupsConfigFields(),
			},
		},
		"fargate_profiles": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "defines the settings used to schedule workload onto Fargate.",
			Elem: &schema.Resource{
				Schema: fargateProfilesConfigField(),
			},
		},
		"availability_zones": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "availability zones of a cluster",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"cloud_watch": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "holds all CloudWatch attributes of a cluster",
			Elem: &schema.Resource{
				Schema: cloudWatchConfigFields(),
			},
		},
		"secrets_encryption": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "defines the configuration for KMS encryption provider",
			Elem: &schema.Resource{
				Schema: secretsEncryptionConfigFields(),
			},
		},
		"identity_mappings": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "maps IAM user/roles to kubenetes RBAC groups",
			Elem: &schema.Resource{
				Schema: identityMappingsConfigFields(),
			},
		},
	}
	return s
}

func configMetadataField() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "EKS Cluster name",
		},
		"region": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "the AWS region hosting this cluster",
		},
		"version": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "1.20",
			Description: "Valid variants are: '1.16', '1.17', '1.18', '1.19', '1.20' (default), '1.21'.",
		},
		"tags": {
			Type:        schema.TypeMap,
			Optional:    true,
			Description: "used to tag AWS resources created by the vendor",
		},
		"annotations": {
			Type:        schema.TypeMap,
			Optional:    true,
			Description: "arbitrary metadata ignored by the vendor",
		},
	}
	return s
}

func kubernetesNetworkConfigField() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"ip_family": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "IPv4",
			Description: "Valid variants are: 'IPv4' defines an IP family of v4 to be used when creating a new VPC and cluster., 'IPv6' defines an IP family of v6 to be used when creating a new VPC and cluster..",
		},
		"service_ipv4_cidr": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "CIDR range from where ClusterIPs are assigned",
		},
	}
	return s
}

func iamFields() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"service_role_arn": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "service role ARN of the cluster",
		},
		"service_role_permission_boundary": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "permissions boundary for all identity-based entities created by the vendor.",
		},
		"fargate_pod_execution_role_arn": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "role used by pods to access AWS APIs. This role is added to the Kubernetes RBAC for authorization.",
		},
		"fargate_pod_execution_role_permissions_boundary": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "permissions boundary for the fargate pod execution role.",
		},
		"with_oidc": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "enables the IAM OIDC provider as well as IRSA for the Amazon CNI plugin",
		},
		"service_accounts": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "service accounts to create in the cluster.",
			Elem: &schema.Resource{
				Schema: serviceAccountsFields(),
			},
		},
		"vpc_resource_controller_policy": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "attaches the IAM policy necessary to run the VPC controller in the control plane",
		},
	}
	return s
}

func serviceAccountsFields() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"metadata": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "metadata for service accounts",
			Elem: &schema.Resource{
				Schema: serviceAccountsMetadata(),
			},
		},
		"attach_policy_arns": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "CIDR range from where ClusterIPs are assigned",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"well_known_policies": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "for attaching common IAM policies",
			Elem: &schema.Resource{
				Schema: serviceAccountsWellKnownPolicyFields(),
			},
		},
		"attach_policy": { //USE THIS FOR ALL INLINEDOCUMENT TYPES
			Type:        schema.TypeString,
			Optional:    true,
			Description: "holds a policy document to attach to this service account",
			/*
				Elem: &schema.Resource{
					Schema: attachPolicyFields(),
				},*/
		},
		"attach_role_arn": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "ARN of the role to attach to the service account",
		},
		"permissions_boundary": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "ARN of the permissions boundary to associate with the service account",
		},
		"status": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "holds status of the IAM service account",
			Elem: &schema.Resource{
				Schema: serviceAccountsStatusFields(),
			},
		},
		"role_name": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Specific role name instead of the Cloudformation-generated role name",
		},
		"role_only": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "Specify if only the IAM Service Account role should be created without creating/annotating the service account",
		},
		"tags": {
			Type:        schema.TypeMap,
			Optional:    true,
			Description: "AWS tags for the service account",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
	}
	return s
}

func serviceAccountsMetadata() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "service account name",
		},
		"namespace": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "service account namespace",
		},
		"labels": {
			Type:        schema.TypeMap,
			Optional:    true,
			Description: "CIDR range from where ClusterIPs are assigned",
		},
		"annotations": {
			Type:        schema.TypeMap,
			Optional:    true,
			Description: "CIDR range from where ClusterIPs are assigned",
		},
	}
	return s
}

// dealing with attach policy inline document object
func attachPolicyFields() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"version": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Attach policy version",
		},
		"statement": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "holds status of the IAM service account",
			Elem: &schema.Resource{
				Schema: statementFields(),
			},
		},
	}
	return s
}

func statementFields() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"effect": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Attach policy effect",
		},
		"action": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Attach policy action",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"resource": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Attach policy resource",
		},
	}
	return s
}

func serviceAccountsWellKnownPolicyFields() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"image_builder": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "allows for full ECR (Elastic Container Registry) access.",
		},
		"auto_scaler": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "service account annotations",
		},
		"aws_load_balancer_controller": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "adds policies for using the aws-load-balancer-controller.",
		},
		"external_dns": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "adds external-dns policies for Amazon Route 53.",
		},
		"cert_manager": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "adds cert-manager policies.",
		},
		"ebs_csi_controller": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "adds cert-manager policies.",
		},
		"efs_csi_controller": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "adds policies for using the ebs-csi-controller.",
		},
	}
	return s
}

func serviceAccountsStatusFields() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"role_arn": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "role ARN associated with the service account.",
		},
	}
	return s
}
func identityProviderField() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"type": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "oidc",
			Description: "Valid variants are: 'oidc': OIDC identity provider",
		},
	}
	return s
}

func vpcFields() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"id": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "AWS VPC ID.",
		},
		"cidr": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "AWS VPC ID.",
		},
		"ipv6_cidr": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "n/a",
		},
		"ipv6_pool": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "n/a",
		},
		"security_group": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "(aka the ControlPlaneSecurityGroup) for communication between control plane and nodes",
		},
		"subnets": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "keyed by AZ for convenience.",
			Elem: &schema.Resource{
				Schema: subnetsConfigFields(),
			},
		},
		"extra_cidrs": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "for additional CIDR associations, e.g. a CIDR for private subnets or any ad-hoc subnets",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"extra_ipv6_cidrs": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "for additional CIDR associations, e.g. a CIDR for private subnets or any ad-hoc subnets",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"shared_node_security_group": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "for pre-defined shared node SG",
		},
		"manage_shared_node_security_group_rules": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "Automatically add security group rules to and from the default cluster security group and the shared node security group. This allows unmanaged nodes to communicate with the control plane and managed nodes. This option cannot be disabled when using vendor created security groups.",
		},
		"auto_allocate_ipv6": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "AutoAllocateIPV6 requests an IPv6 CIDR block with /56 prefix for the VPC",
		},
		"nat": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "NAT config",
			Elem: &schema.Resource{
				Schema: natConfigFields(),
			},
		},
		"cluster_endpoints": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Manage access to the Kubernetes API server endpoints.",
			Elem: &schema.Resource{
				Schema: clusterEndpointsConfigFields(),
			},
		},
		"public_access_cidrs": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "which CIDR blocks to allow access to public k8s API endpoint",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
	}
	return s
}
func subnetsConfigFields() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"private": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "holds subnet to AZ mappings. If the key is an AZ, that also becomes the name of the subnet otherwise use the key to refer to this subnet.",
			Elem: &schema.Resource{
				Schema: subnetSpecConfigFields(),
			},
		},
		"public": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "holds subnet to AZ mappings. If the key is an AZ, that also becomes the name of the subnet otherwise use the key to refer to this subnet.",
			Elem: &schema.Resource{
				Schema: subnetSpecConfigFields(),
			},
		},
	}
	return s
}
func subnetSpecConfigFields() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "name of subnet",
		},
		"id": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "id of subnet",
		},
		"az": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "dont know what this is, not in docs",
		},
		"cidr": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "dont know what this is, not in docs",
		},
	}
	return s
}

func natConfigFields() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"gateway": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "Single",
			Description: "Valid variants are: 'HighlyAvailable' configures a highly available NAT gateway, 'Single' configures a single NAT gateway (default), 'Disable' disables NAT.",
		},
	}
	return s
}

func clusterEndpointsConfigFields() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"private_access": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "enable private access to the Kubernetes API server endpoints.",
		},
		"public_access": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "enable public access to the Kubernetes API server endpoints.",
		},
	}
	return s
}

func addonConfigFields() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "EKS addon name",
		},
		"version": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "EKS addon version",
		},
		"service_account_role_arn": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "service account role ARN",
		},
		"attach_policy_arns": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "list of ARNs of the IAM policies to attach",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"attach_policy": { //USE THIS FOR ALL INLINEDOCUMENT TYPES
			Type:        schema.TypeList,
			Optional:    true,
			Description: "holds a policy document to attach to this service account",
			Elem: &schema.Resource{
				Schema: attachPolicyFields(),
			},
		},
		"permissions_boundary": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "ARN of the permissions boundary to associate",
		},
		"well_known_policies": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "for attaching common IAM policies",
			Elem: &schema.Resource{
				Schema: serviceAccountsWellKnownPolicyFields(),
			},
		},
		"tags": {
			Type:        schema.TypeMap,
			Optional:    true,
			Description: "The metadata to apply to the cluster to assist with categorization and organization. Each tag consists of a key and an optional value, both of which you define.",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"configuration_values": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "configuration values for the addon",
		},
	}
	return s
}

func privateClusterConfigFields() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"enabled": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "enables creation of a fully-private cluster",
		},
		"skip_endpoint_creation": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "skips the creation process for endpoints completely. This is only used in case of an already provided VPC and if the user decided to set it to true.",
		},
		"additional_endpoint_services": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "specifies additional endpoint services that must be enabled for private access. Valid entries are: 'cloudformation', 'autoscaling', 'logs'",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
	}
	return s
}

func nodeGroupsConfigFields() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "name of the node group",
		},
		"ami_family": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Valid variants are: 'AmazonLinux2' (default), 'Ubuntu2004', 'Ubuntu1804', 'Bottlerocket', 'WindowsServer2019CoreContainer', 'WindowsServer2019FullContainer', 'WindowsServer2004CoreContainer'.",
		},
		"instance_type": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "type of instances in the nodegroup",
		},
		"availability_zones": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Limit nodes to specific AZs",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"subnets": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Limit nodes to specific subnets",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"instance_prefix": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "for instances in the nodegroup",
		},
		"instance_name": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "for instances in the nodegroup",
		},
		"desired_capacity": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "Desired capacity of instances in the nodegroup",
		},
		"min_size": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "Minimum size of instances in the nodegroup",
		},
		"max_size": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "Maximum size of instances in the nodegroup",
		},
		"volume_size": {
			Type:        schema.TypeInt,
			Optional:    true,
			Default:     80,
			Description: "volume size in gigabytes",
		},
		"ssh": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "configures ssh access for this nodegroup",
			Elem: &schema.Resource{
				Schema: sshConfigFields(),
			},
		},
		"labels": {
			Type:        schema.TypeMap,
			Optional:    true,
			Description: "labels on nodes in the nodegroup",
		},
		"private_networking": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Enable private networking for nodegroup",
		},
		"tags": {
			Type:        schema.TypeMap,
			Optional:    true,
			Description: "Applied to the Autoscaling Group and to the EC2 instances (unmanaged), Applied to the EKS Nodegroup resource and to the EC2 instances (managed)",
		},
		"iam": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "holds all IAM attributes of a NodeGroup",
			Elem: &schema.Resource{
				Schema: iamNodeGroupConfigFields(),
			},
		},
		"ami": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Specify custom AMIs, auto-ssm, auto, or static",
		},
		"security_groups": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "controls security groups for this nodegroup",
			Elem: &schema.Resource{
				Schema: securityGroupsConfigFields(),
			},
		},
		"max_pods_per_node": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "Maximum pods per node",
		},
		"asg_suspend_processes": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "See relevant AWS docs",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"ebs_optimized": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "enables EBS optimization",
		},
		"volume_type": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "gp3",
			Description: "Valid variants are: 'gp2' is General Purpose SSD, 'gp3' is General Purpose SSD which can be optimised for high throughput (default), 'io1' is Provisioned IOPS SSD, 'sc1' is Cold HDD, 'st1' is Throughput Optimized HDD.",
		},
		"volume_name": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "of volumes attached to instances in the nodegroup",
		},
		"volume_encrypted": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "of volumes attached to instances in the nodegroup",
		},
		"volume_kms_key_id": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "of volumes attached to instances in the nodegroup",
		},
		"volume_iops": {
			Type:        schema.TypeInt,
			Optional:    true,
			Default:     3000,
			Description: "of volumes attached to instances in the nodegroup",
		},
		"volume_throughput": {
			Type:        schema.TypeInt,
			Optional:    true,
			Default:     125,
			Description: "of volumes attached to instances in the nodegroup",
		},
		"pre_bootstrap_commands": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "executed before bootstrapping instances to the cluster",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"override_bootstrap_command": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Override the vendor's bootstrapping script",
		},
		"disable_imdsv1": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "requires requests to the metadata service to use IMDSv2 tokens",
		},
		"disable_pods_imds": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "blocks all IMDS requests from non host networking pods",
		},
		"placement": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "specifies the placement group in which nodes should be spawned",
			Elem: &schema.Resource{
				Schema: placementField(),
			},
		},
		"efa_enabled": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "creates the maximum allowed number of EFA-enabled network cards on nodes in this group.",
		},
		"instance_selector": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "specifies options for EC2 instance selector",
			Elem: &schema.Resource{
				Schema: instanceSelectorFields(),
			},
		},
		"bottle_rocket": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "specifies settings for Bottlerocket nodes",
			Elem: &schema.Resource{
				Schema: bottleRocketFields(),
			},
		},
		"enable_detailed_monitoring": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "Enable EC2 detailed monitoring",
		},
		"instances_distribution": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "holds the configuration for spot instances",
			Elem: &schema.Resource{
				Schema: instanceDistributionFields(),
			},
		},
		"asg_metrics_collection": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "used by the scaling config, see cloudformation docs",
			Elem: &schema.Resource{
				Schema: asgMetricsCollectionFields(),
			},
		},
		"cpu_credits": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "configures T3 Unlimited, valid only for T-type instances",
		},
		"classic_load_balancer_names": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Associate load balancers with auto scaling group",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"target_group_arns": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Associate target group with auto scaling group",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"taints": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "taints to apply to the nodegroup",
			Elem: &schema.Resource{
				Schema: managedNodeGroupTaintConfigFields(),
			},
		},
		"update_config": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "used by the scaling config, see cloudformation docs",
			Elem: &schema.Resource{
				Schema: updateConfigFields(),
			},
		},
		"cluster_dns": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Custom address used for DNS lookups",
		},

		"kubelet_extra_config": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Custom address used for DNS lookups",
			Elem: &schema.Resource{
				Schema: kubeLetExtraConfigFields(),
			},
		},
		"version": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Kuberenetes version for the nodegroup",
		},
		"subnet_cidr": {
			Type:        schema.TypeString, //supposed be of type object?
			Optional:    true,
			Description: "Create new subnet from the CIDR block and limit nodes to this subnet (Applicable only for the WavelenghZone nodes)",
		},
	}
	return s
}

// @@@
func kubeLetExtraConfigFields() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"kube_reserved": {
			Type:        schema.TypeMap,
			Optional:    true,
			Description: "",
		},
		"kube_reserved_cgroup": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "",
		},
		"system_reserved": {
			Type:        schema.TypeMap,
			Optional:    true,
			Description: "",
		},
		"eviction_hard": {
			Type:        schema.TypeMap,
			Optional:    true,
			Description: "",
		},
		"feature_gates": {
			Type:        schema.TypeMap,
			Optional:    true,
			Description: "",
		},
	}
	return s
}

func instanceDistributionFields() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"instance_types": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Enable admin container",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"max_price": {
			Type:        schema.TypeFloat,
			Optional:    true,
			Description: "Maximum bid price in USD",
		},
		"on_demand_base_capacity": {
			Type:        schema.TypeInt,
			Optional:    true,
			Default:     0,
			Description: "base number of on-demand instances (non-negative)",
		},
		"on_demand_percentage_above_base_capacity": {
			Type:        schema.TypeInt,
			Optional:    true,
			Default:     100,
			Description: "Range [0-100]",
		},
		"spot_instance_pools": {
			Type:        schema.TypeInt,
			Optional:    true,
			Default:     2,
			Description: "Range [0-20]",
		},
		"spot_allocation_strategy": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "allocation strategy for spot instances. Valid values are capacity-optimized and lowest-price",
		},
		"capacity_rebalance": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Enable capacity rebalancing for spot instances",
		},
	}
	return s
}

func updateConfigFields() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"max_unavaliable": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "sets the max number of nodes that can become unavailable when updating a nodegroup (specified as number)",
		},
		"max_unavaliable_percetage": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "sets the max number of nodes that can become unavailable when updating a nodegroup (specified as percentage)",
		},
	}
	return s
}

func asgMetricsCollectionFields() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"granularity": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "of metrics collected",
		},
		"metrics": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "specifies a list of metrics",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
	}
	return s
}

func bottleRocketFields() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"enable_admin_container": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "Enable admin container",
		},
		"settings": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "contains any bottlerocket settings",
		},
	}
	return s
}

func instanceSelectorFields() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"vcpus": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "specifies the number of vCPUs",
		},
		"memory": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "specifies the memory The unit defaults to GiB",
		},
		"gpus": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "specifies the number of GPUs. It can be set to 0 to select non-GPU instance types.",
		},
		"cpu_architecture": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "x86_64",
			Description: "CPU Architecture of the EC2 instance type. Valid variants are: 'x86_64' 'amd64' 'arm64'",
		},
	}
	return s
}
func placementField() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"group": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "placement group name ",
		},
	}
	return s
}

func sshConfigFields() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"allow": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "If Allow is true the SSH configuration provided is used, otherwise it is ignored. Only one of PublicKeyPath, PublicKey and PublicKeyName can be configured",
		},
		"public_key": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Public key to be added to the nodes SSH keychain. If Allow is false this value is ignored.",
		},
		"public_key_name": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Public key name in EC2 to be added to the nodes SSH keychain. If Allow is false this value is ignored.",
		},
		"source_security_group_ids": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "source securitgy group IDs",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"enable_ssm": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "Enables the ability to SSH onto nodes using SSM",
		},
	}
	return s
}

func iamNodeGroupConfigFields() map[string]*schema.Schema { //@@@TODO: need to change schema to have attachPolicy(inline object)
	s := map[string]*schema.Schema{
		"attach_policy": { //USE THIS FOR ALL INLINEDOCUMENT TYPES
			Type:        schema.TypeList,
			Optional:    true,
			Description: "holds a policy document to attach to this service account",
			Elem: &schema.Resource{
				Schema: attachPolicyFields(),
			},
		},
		"attach_policy_arns": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "attach polciy ARN",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"instance_profile_arn": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "instance profile ARN",
		},
		"instance_role_arn": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "instance role ARN",
		},
		"instance_role_name": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "instance role Name",
		},
		"instance_role_permission_boundary": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "instance role permissions boundary",
		},
		"iam_node_group_with_addon_policies": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "holds all IAM attributes of a NodeGroup",
			Elem: &schema.Resource{
				Schema: iamNodeGroupWithAddonPoliciesFields(),
			},
		},
	}
	return s
}

func iamNodeGroupWithAddonPoliciesFields() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"image_builder": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
			Description: "allows for full ECR (Elastic Container Registry) access. This is useful for building, for example, a CI server that needs to push images to ECR",
		},
		"auto_scaler": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
			Description: "enables IAM policy for cluster-autoscaler",
		},
		"external_dns": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "adds the external-dns project policies for Amazon Route 53",
		},
		"cert_manager": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "enables the ability to add records to Route 53 in order to solve the DNS01 challenge.",
		},
		"app_mesh": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "enables full access to AppMesh",
		},
		"app_mesh_review": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "enables full access to AppMesh Preview",
		},
		"ebs": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "enables the new EBS CSI (Elastic Block Store Container Storage Interface) driver",
		},
		"fsx": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "enables full access to FSX",
		},
		"efs": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "enables full access to EFS",
		},
		"alb_ingress": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "enables access to ALB Ingress controller",
		},
		"xray": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "enables access to XRay",
		},
		"cloud_watch": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "enables access to cloud watch",
		},
	}
	return s
}

func securityGroupsConfigFields() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"attach_ids": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "attaches additional security groups to the nodegroup",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"with_shared": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
			Description: "attach the security group shared among all nodegroups in the cluster",
		},
		"with_local": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
			Description: "attach a security group local to this nodegroup Not supported for managed nodegroups",
		},
	}
	return s
}
func managedSecurityGroupsConfigFields() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"attach_ids": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "attaches additional security groups to the nodegroup",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"with_shared": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "attach the security group shared among all nodegroups in the cluster",
		},
		"with_local": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "attach a security group local to this nodegroup Not supported for managed nodegroups",
		},
	}
	return s
}

func managedNodeGroupsConfigFields() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "name of the node group",
		},
		"ami_family": {
			Type:     schema.TypeString,
			Optional: true,
			// Default:     "AmazonLinux2",
			Description: "Valid variants are: 'AmazonLinux2'.",
		},
		"instance_type": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "type of instances in the nodegroup",
		},
		"availability_zones": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Limit nodes to specific AZs",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"subnets": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Limit nodes to specific subnets",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"instance_prefix": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "for instances in the nodegroup",
		},
		"instance_name": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "for instances in the nodegroup",
		},
		"desired_capacity": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "desired capacity of instances in the nodegroup",
		},
		"min_size": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "minimum size of instances in the nodegroup",
		},
		"max_size": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "maximum size of instances in the nodegroup",
		},
		"volume_size": {
			Type:        schema.TypeInt,
			Optional:    true,
			Default:     80,
			Description: "in gigabytes",
		},
		"ssh": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "configures ssh access for this nodegroup",
			Elem: &schema.Resource{
				Schema: sshConfigFields(),
			},
		},
		"labels": {
			Type:        schema.TypeMap,
			Optional:    true,
			Description: "labels on nodes in the nodegroup",
		},
		"private_networking": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Enable private networking for nodegroup",
		},
		"tags": {
			Type:        schema.TypeMap,
			Optional:    true,
			Description: "Applied to the Autoscaling Group and to the EC2 instances (unmanaged), Applied to the EKS Nodegroup resource and to the EC2 instances (managed)",
		},
		"iam": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "holds all IAM attributes of a NodeGroup",
			Elem: &schema.Resource{
				Schema: iamNodeGroupConfigFields(),
			},
		},
		"ami": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Specify custom AMIs, auto-ssm, auto, or static",
		},
		"security_groups": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "controls security groups for this nodegroup",
			Elem: &schema.Resource{
				Schema: managedSecurityGroupsConfigFields(),
			},
		},
		"max_pods_per_node": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "Maximum pods per node",
		},
		"asg_suspend_processes": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "See relevant AWS docs",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"ebs_optimized": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "enables EBS optimization",
		},
		"volume_type": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "gp3",
			Description: "Valid variants are: 'gp2' is General Purpose SSD, 'gp3' is General Purpose SSD which can be optimised for high throughput (default), 'io1' is Provisioned IOPS SSD, 'sc1' is Cold HDD, 'st1' is Throughput Optimized HDD.",
		},
		"volume_name": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "of volumes attached to instances in the nodegroup",
		},
		"volume_encrypted": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "of volumes attached to instances in the nodegroup",
		},
		"volume_kms_key_id": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "of volumes attached to instances in the nodegroup",
		},
		"volume_iops": {
			Type:        schema.TypeInt,
			Optional:    true,
			Default:     3000,
			Description: "of volumes attached to instances in the nodegroup",
		},
		"volume_throughput": {
			Type:        schema.TypeInt,
			Optional:    true,
			Default:     125,
			Description: "of volumes attached to instances in the nodegroup",
		},
		"pre_bootstrap_commands": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "executed before bootstrapping instances to the cluster",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"override_bootstrap_command": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Override the vendor's bootstrapping script",
		},
		"disable_imdsv1": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "requires requests to the metadata service to use IMDSv2 tokens",
		},
		"disable_pods_imds": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "blocks all IMDS requests from non host networking pods",
		},
		"placement": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "specifies the placement group in which nodes should be spawned",
			Elem: &schema.Resource{
				Schema: placementField(),
			},
		},
		"efa_enabled": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "creates the maximum allowed number of EFA-enabled network cards on nodes in this group.",
		},
		"instance_selector": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "specifies options for EC2 instance selector",
			Elem: &schema.Resource{
				Schema: instanceSelectorFields(),
			},
		},
		"bottle_rocket": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "specifies settings for Bottlerocket nodes",
			Elem: &schema.Resource{
				Schema: bottleRocketFields(),
			},
		},
		"enable_detailed_monitoring": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "Enable EC2 detailed monitoring",
		},
		"instance_types": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "specifies a list of instance types",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"spot": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "create a spot nodegroup",
		},
		"taints": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "taints to apply to the nodegroup",
			Elem: &schema.Resource{
				Schema: managedNodeGroupTaintConfigFields(),
			},
		},
		"update_config": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "used by the scaling config, see cloudformation docs",
			Elem: &schema.Resource{
				Schema: updateConfigManagedNodeGroupsFields(),
			},
		},
		"launch_template": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "used by the scaling config, see cloudformation docs",
			Elem: &schema.Resource{
				Schema: launchTempelateFields(),
			},
		}, //@@@ check eks_config.go wats this release version, is it in launch tempelate or managedNodeGroups, doc vs eks_config is confusing
		"version": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Kuberenetes version for the nodegroup",
		},
	}
	return s
}

func launchTempelateFields() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"id": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "key of taint",
		},
		"version": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "value of taint",
		},
	}
	return s
}

func managedNodeGroupTaintConfigFields() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"key": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "key of taint",
		},
		"value": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "value of taint",
		},
		"effect": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "effect of taint",
		},
	}
	return s
}

func updateConfigManagedNodeGroupsFields() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"max_unavailable": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "sets the max number of nodes that can become unavailable when updating a nodegroup (specified as number)",
		},
		"max_unavailable_percentage": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "sets the max number of nodes that can become unavailable when updating a nodegroup (specified as percentage)",
		},
	}
	return s
}

func fargateProfilesConfigField() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "name of the fargate profile",
		},
		"pod_execution_role_arn": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "IAM role's ARN to use to run pods onto Fargate.",
		},
		"selectors": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "define the rules to select workload to schedule onto Fargate.",
			Elem: &schema.Resource{
				Schema: selectorsFields(),
			},
		},
		"subnets": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "define the rules to select workload to schedule onto Fargate.",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"tags": {
			Type:        schema.TypeMap,
			Optional:    true,
			Description: "Used to tag the AWS resources",
		},
		"status": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The current status of the Fargate profile.",
		},
	}
	return s
}

func selectorsFields() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"namespace": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Kubernetes namespace from which to select workload.",
		},
		"labels": {
			Type:        schema.TypeMap,
			Optional:    true,
			Description: "Kubernetes label selectors to use to select workload.",
		},
	}
	return s
}

func cloudWatchConfigFields() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"cluster_logging": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "container config parameters related to cluster logging",
			Elem: &schema.Resource{
				Schema: clusterLoggingFields(),
			},
		},
	}
	return s
}

func clusterLoggingFields() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"enable_types": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Types of logging to enable. Valid entries are: 'api', 'audit', 'authenticator', 'controllerManager', 'scheduler', 'all', '*'.",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"log_retention_in_days": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "The number of days you want to retain log events in the specified log group. Possible values are: 1, 3, 5, 7, 14, 30, 60, 90, 120, 150, 180, 365, 400, 545, 731, 1827, and 3653.",
		},
	}
	return s
}

func secretsEncryptionConfigFields() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"key_arn": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "KMS key ARN",
		},
	}
	return s
}

func identityMappingsConfigFields() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"arns": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "List of ARN objects",
			Elem: &schema.Resource{
				Schema: arnFields(),
			},
		},
		"accounts": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "List of IAM accounts to map",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
	}
	return s
}

func arnFields() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"arn": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "ARN of user/role to be mapped",
		},
		"group": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "List of kubernetes groups to be mapped to",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"username": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The username to be used by kubernetes",
		},
	}
	return s
}

func resourceEKSClusterUpsert(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("resourceEKSClusterUpsert")
	return processEKSInputs(ctx, d, m)

}

// expand eks cluster function (completed)
func expandEKSCluster(p []interface{}) *EKSCluster {
	obj := &EKSCluster{}

	if len(p) == 0 || p[0] == nil {
		return obj
	}
	//prefix = prefix + ".0"
	in := p[0].(map[string]interface{})
	if v, ok := in["kind"].(string); ok && len(v) > 0 {
		obj.Kind = v
	}
	if v, ok := in["metadata"].([]interface{}); ok && len(v) > 0 {
		obj.Metadata = expandEKSMetaMetadata(v)
	}
	if v, ok := in["spec"].([]interface{}); ok && len(v) > 0 {
		obj.Spec = expandEKSClusterSpecConfig(v)
	}
	return obj
}

// expand eks cluster function (completed)
func expandEKSClusterConfig(p []interface{}, rawConfig cty.Value) *EKSClusterConfig {
	obj := &EKSClusterConfig{}

	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})
	rawConfig = rawConfig.AsValueSlice()[0]
	if v, ok := in["kind"].(string); ok && len(v) > 0 {
		obj.Kind = v
	}
	if v, ok := in["apiversion"].(string); ok && len(v) > 0 {
		obj.APIVersion = v
	}
	if v, ok := in["metadata"].([]interface{}); ok && len(v) > 0 {
		obj.Metadata = expandEKSSpecMetadata(v)
	}
	if v, ok := in["kubernetes_network_config"].([]interface{}); ok && len(v) > 0 {
		obj.KubernetesNetworkConfig = expandKubernetesNetworkConfig(v)
	}
	if v, ok := in["iam"].([]interface{}); ok && len(v) > 0 {
		obj.IAM = expandIAMFields(v)
	}
	if v, ok := in["identity_providers"].([]interface{}); ok && len(v) > 0 {
		obj.IdentityProviders = expandIdentityProviders(v)
	}

	if v, ok := in["addons"].([]interface{}); ok && len(v) > 0 {
		obj.Addons = expandAddons(v)
	}
	if v, ok := in["private_cluster"].([]interface{}); ok && len(v) > 0 {
		obj.PrivateCluster = expandPrivateCluster(v)
	}
	if v, ok := in["node_groups"].([]interface{}); ok && len(v) > 0 {
		obj.NodeGroups = expandNodeGroups(v)
	}
	if v, ok := in["vpc"].([]interface{}); ok && len(v) > 0 {
		obj.VPC = expandVPC(v, rawConfig.GetAttr("vpc"))
	}
	if v, ok := in["managed_nodegroups"].([]interface{}); ok && len(v) > 0 {
		obj.ManagedNodeGroups = expandManagedNodeGroups(v, rawConfig.GetAttr("managed_nodegroups"))
	}
	if v, ok := in["fargate_profiles"].([]interface{}); ok && len(v) > 0 {
		obj.FargateProfiles = expandFargateProfiles(v)
	}
	if v, ok := in["availability_zones"].([]interface{}); ok && len(v) > 0 {
		obj.AvailabilityZones = toArrayStringSorted(v)
	}
	if v, ok := in["cloud_watch"].([]interface{}); ok && len(v) > 0 {
		obj.CloudWatch = expandCloudWatch(v)
	}
	if v, ok := in["secrets_encryption"].([]interface{}); ok && len(v) > 0 {
		obj.SecretsEncryption = expandSecretEncryption(v)
	}
	if v, ok := in["identity_mappings"].([]interface{}); ok && len(v) > 0 {
		obj.IdentityMappings = expandIdentityMappings(v)
	}
	return obj
}

func processEKSInputs(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	//building cluster and cluster config yaml file
	var yamlCluster *EKSCluster
	var yamlClusterConfig *EKSClusterConfig
	rawConfig := d.GetRawConfig()
	//expand cluster yaml file
	if v, ok := d.Get("cluster").([]interface{}); ok {
		yamlCluster = expandEKSCluster(v)
	} else {
		log.Print("Cluster data unable to be found")
		return diag.FromErr(fmt.Errorf("%s", "Cluster data is missing"))
	}
	//expand cluster config yaml file
	if v, ok := d.Get("cluster_config").([]interface{}); ok {
		yamlClusterConfig = expandEKSClusterConfig(v, rawConfig.GetAttr("cluster_config"))
	} else {
		log.Print("Cluster Config unable to be found")
		return diag.FromErr(fmt.Errorf("%s", "Cluster Config is missing"))
	}

	return processEKSFilebytes(ctx, d, m, yamlCluster, yamlClusterConfig)
}
func processEKSFilebytes(ctx context.Context, d *schema.ResourceData, m interface{}, yamlClusterMetadata *EKSCluster, yamlClusterConfig *EKSClusterConfig) diag.Diagnostics {
	log.Printf("process_filebytes")
	var diags diag.Diagnostics

	clusterName := yamlClusterMetadata.Metadata.Name
	projectName := yamlClusterMetadata.Metadata.Project
	projectID, err := getProjectIDFromName(projectName)
	if err != nil {
		log.Print("error converting project name to id")
		return diag.Errorf("error converting project name to project ID")
	}

	var b bytes.Buffer
	encoder := yaml.NewEncoder(&b)
	if err := encoder.Encode(yamlClusterMetadata); err != nil {
		log.Printf("error encoding cluster: %s", err)
		return diag.FromErr(err)
	}
	if err := encoder.Encode(yamlClusterConfig); err != nil {
		log.Printf("error encoding cluster config: %s", err)
		return diag.FromErr(err)
	}

	logger := glogger.GetLogger()
	rctlConfig := config.GetConfig()

	log.Printf("calling cluster ctl:\n%s", b.String())
	response, err := clusterctl.Apply(logger, rctlConfig, clusterName, b.Bytes(), false, false)
	if err != nil {
		log.Printf("cluster error 1: %s", err)
		return diag.FromErr(err)
	}

	log.Printf("process_filebytes response : %s", response)
	res := clusterCTLResponse{}
	err = json.Unmarshal([]byte(response), &res)
	if err != nil {
		log.Println("response parse error", err)
		return diag.FromErr(err)
	}
	if res.TaskSetID == "" {
		return nil
	}
	time.Sleep(10 * time.Second)
	s, errGet := cluster.GetCluster(clusterName, projectID)
	if errGet != nil {
		log.Printf("error while getCluster %s", errGet.Error())
		return diag.FromErr(errGet)
	}

	log.Println("Cluster Provision may take upto 15-20 Minutes")
	d.SetId(s.ID)
	for { //wait for cluster to provision correctly
		time.Sleep(60 * time.Second)
		check, errGet := cluster.GetCluster(yamlClusterMetadata.Metadata.Name, projectID)
		if errGet != nil {
			log.Printf("error while getCluster %s", errGet.Error())
			return diag.FromErr(errGet)
		}
		rctlConfig.ProjectID = projectID
		statusResp, err := clusterctl.Status(logger, rctlConfig, res.TaskSetID)
		if err != nil {
			log.Println("status response parse error", err)
			return diag.FromErr(err)
		}
		log.Println("statusResp:\n ", statusResp)
		sres := clusterCTLResponse{}
		err = json.Unmarshal([]byte(statusResp), &sres)
		if err != nil {
			log.Println("status response unmarshal error", err)
			return diag.FromErr(err)
		}
		if strings.Contains(sres.Status, "STATUS_COMPLETE") {
			if check.Status == "READY" {
				break
			}
			log.Println("task completed but cluster is not ready")
		}
		if strings.Contains(sres.Status, "STATUS_FAILED") {
			return diag.FromErr(fmt.Errorf("failed to create/update cluster while provisioning cluster %s %s", yamlClusterMetadata.Metadata.Name, statusResp))
		}
	}

	log.Printf("resource eks cluster created/updated %s", s.ID)

	return diags
}
func eksClusterCTLStatus(taskid, projectID string) (string, error) {
	log.Println("eksClusterCTLStatus")
	logger := glogger.GetLogger()
	rctlCfg := config.GetConfig()
	rctlCfg.ProjectID = projectID
	return clusterctl.Status(logger, rctlCfg, taskid)
}

// expand metadat for eks metadata file  (completed)
func expandEKSMetaMetadata(p []interface{}) *EKSClusterMetadata {
	obj := &EKSClusterMetadata{}

	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})
	if v, ok := in["name"].(string); ok && len(v) > 0 {
		obj.Name = v
	}
	if v, ok := in["project"].(string); ok && len(v) > 0 {
		obj.Project = v
	}
	if v, ok := in["labels"].(map[string]interface{}); ok && len(v) > 0 {
		obj.Labels = toMapString(v)
	}
	return obj
}

// expand metadata for eks spec metadata (completed)
func expandEKSSpecMetadata(p []interface{}) *EKSClusterConfigMetadata {
	obj := &EKSClusterConfigMetadata{}

	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})
	if v, ok := in["name"].(string); ok && len(v) > 0 {
		obj.Name = v
	}
	if v, ok := in["region"].(string); ok && len(v) > 0 {
		obj.Region = v
	}
	if v, ok := in["version"].(string); ok && len(v) > 0 {
		obj.Version = v
	}
	if v, ok := in["tags"].(map[string]interface{}); ok && len(v) > 0 {
		obj.Tags = toMapString(v)
	}
	if v, ok := in["annotations"].(map[string]interface{}); ok && len(v) > 0 {
		obj.Annotations = toMapString(v)
	}
	return obj
}

// expand secret encryption (completed)
func expandSecretEncryption(p []interface{}) *SecretsEncryption {
	obj := &SecretsEncryption{}

	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})
	if v, ok := in["key_arn"].(string); ok && len(v) > 0 {
		obj.KeyARN = v
	}
	return obj
}

func expandIdentityMappings(p []interface{}) *EKSClusterIdentityMappings {
	obj := &EKSClusterIdentityMappings{}

	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["arns"].([]interface{}); ok && len(v) > 0 {
		obj.Arns = expandArnFields(v)
	}
	if v, ok := in["accounts"].([]interface{}); ok && len(v) > 0 {
		obj.Accounts = toArrayString(v)
	}

	return obj
}

func expandArnFields(p []interface{}) []*IdentityMappingARN {
	out := make([]*IdentityMappingARN, len(p))

	if len(p) == 0 || p[0] == nil {
		return out
	}

	for i := range p {
		obj := &IdentityMappingARN{}
		in := p[i].(map[string]interface{})

		if v, ok := in["arn"].(string); ok && len(v) > 0 {
			obj.Arn = v
		}

		if v, ok := in["group"].([]interface{}); ok && len(v) > 0 {
			obj.Group = toArrayString(v)
		}

		if v, ok := in["username"].(string); ok && len(v) > 0 {
			obj.Username = v
		}
		out[i] = obj
	}

	return out
}

// expand cloud watch function (completed)
func expandCloudWatch(p []interface{}) *EKSClusterCloudWatch {
	obj := &EKSClusterCloudWatch{}

	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})
	if v, ok := in["cluster_logging"].([]interface{}); ok && len(v) > 0 {
		obj.ClusterLogging = expandCloudWatchClusterLogging(v)
	}
	return obj
}

func expandCloudWatchClusterLogging(p []interface{}) *EKSClusterCloudWatchLogging {
	obj := &EKSClusterCloudWatchLogging{}

	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})
	if v, ok := in["enable_types"].([]interface{}); ok && len(v) > 0 {
		obj.EnableTypes = toArrayString(v)
	}
	if v, ok := in["log_retention_in_days"].(int); ok {
		obj.LogRetentionInDays = v
	}

	return obj
}

// expand fargate profiles (completed)
func expandFargateProfiles(p []interface{}) []*FargateProfile {
	obj := FargateProfile{}
	out := make([]*FargateProfile, len(p))
	if len(p) == 0 || p[0] == nil {
		return out
	}

	for i := range p {
		in := p[i].(map[string]interface{})
		if v, ok := in["name"].(string); ok && len(v) > 0 {
			obj.Name = v
		}
		if v, ok := in["pod_execution_role_arn"].(string); ok && len(v) > 0 {
			obj.PodExecutionRoleARN = v
		}
		if v, ok := in["selectors"].([]interface{}); ok && len(v) > 0 {
			obj.Selectors = expandFargateProfilesSelectors(v)
		}
		if v, ok := in["subnets"].([]interface{}); ok && len(v) > 0 {
			obj.Subnets = toArrayString(v)
		}
		if v, ok := in["tags"].(map[string]interface{}); ok && len(v) > 0 {
			obj.Tags = toMapString(v)
		}
		if v, ok := in["status"].(string); ok && len(v) > 0 {
			obj.Status = v
		}
		out[i] = &obj
	}

	return out
}

func expandFargateProfilesSelectors(p []interface{}) []FargateProfileSelector {
	obj := &FargateProfileSelector{}
	out := make([]FargateProfileSelector, len(p))
	if len(p) == 0 || p[0] == nil {
		return out
	}
	for i := range p {
		in := p[i].(map[string]interface{})
		if v, ok := in["namespace"].(string); ok && len(v) > 0 {
			obj.Namespace = v
		}
		if v, ok := in["labels"].(map[string]interface{}); ok && len(v) > 0 {
			obj.Labels = toMapString(v)
		}
		out[i] = *obj
	}
	return out
}

func expandManagedNodeGroups(p []interface{}, rawConfig cty.Value) []*ManagedNodeGroup { //not completed have questions in comments
	out := make([]*ManagedNodeGroup, len(p))
	outToSort := make([]ManagedNodeGroup, len(p))
	if len(p) == 0 || p[0] == nil {
		return out
	}
	log.Println("got to managed node group")
	for i := range p {
		obj := &ManagedNodeGroup{}
		in := p[i].(map[string]interface{})
		nRawConfig := rawConfig.AsValueSlice()[i]
		if v, ok := in["name"].(string); ok && len(v) > 0 {
			obj.Name = v
		}
		if v, ok := in["ami_family"].(string); ok && len(v) > 0 {
			obj.AMIFamily = v
		}
		if v, ok := in["instance_type"].(string); ok && len(v) > 0 {
			obj.InstanceType = v
		}
		if v, ok := in["availability_zones"].([]interface{}); ok && len(v) > 0 {
			obj.AvailabilityZones = toArrayStringSorted(v)
		}
		if v, ok := in["subnets"].([]interface{}); ok && len(v) > 0 {
			obj.Subnets = toArrayString(v)
		}
		if v, ok := in["instance_prefix"].(string); ok && len(v) > 0 {
			obj.InstancePrefix = v
		}
		if v, ok := in["instance_name"].(string); ok && len(v) > 0 {
			obj.InstanceName = v
		}
		if v, ok := in["desired_capacity"].(int); ok {
			obj.DesiredCapacity = &v
		}
		if v, ok := in["min_size"].(int); ok {
			obj.MinSize = &v
		}
		if v, ok := in["max_size"].(int); ok {
			obj.MaxSize = &v
		}
		if v, ok := in["volume_size"].(int); ok {
			obj.VolumeSize = &v
		}
		if v, ok := in["ssh"].([]interface{}); ok && len(v) > 0 {
			obj.SSH = expandNodeGroupSsh(v, true)
		}
		if v, ok := in["labels"].(map[string]interface{}); ok && len(v) > 0 {
			obj.Labels = toMapString(v)
		}
		if v, ok := in["private_networking"].(bool); ok {
			obj.PrivateNetworking = &v
		}
		if v, ok := in["tags"].(map[string]interface{}); ok && len(v) > 0 {
			obj.Tags = toMapString(v)
		}
		if v, ok := in["iam"].([]interface{}); ok && len(v) > 0 {
			obj.IAM = expandNodeGroupIam(v)
		}
		if v, ok := in["ami"].(string); ok && len(v) > 0 {
			obj.AMI = v
		}
		if v, ok := in["security_groups"].([]interface{}); ok && len(v) > 0 {
			obj.SecurityGroups = expandManagedNodeGroupSecurityGroups(v, nRawConfig.GetAttr("security_groups"))
		}
		if v, ok := in["max_pods_per_node"].(int); ok {
			obj.MaxPodsPerNode = &v
		}
		if v, ok := in["asg_suspend_process"].([]interface{}); ok && len(v) > 0 {
			obj.ASGSuspendProcesses = toArrayString(v)
		}
		if v, ok := in["ebs_optimized"].(bool); ok {
			obj.EBSOptimized = &v
		}
		if v, ok := in["volume_type"].(string); ok && len(v) > 0 {
			obj.VolumeType = v
		}
		if v, ok := in["volume_name"].(string); ok && len(v) > 0 {
			obj.VolumeName = v
		}
		if v, ok := in["volume_encrypted"].(bool); ok {
			obj.VolumeEncrypted = &v
		}
		if v, ok := in["volume_kms_key_id"].(string); ok && len(v) > 0 {
			obj.VolumeKmsKeyID = v
		}
		if v, ok := in["volume_iops"].(int); ok {
			obj.VolumeIOPS = &v
		}
		if v, ok := in["volume_throughput"].(int); ok {
			obj.VolumeThroughput = &v
		}
		if v, ok := in["pre_bootstrap_commands"].([]interface{}); ok && len(v) > 0 {
			obj.PreBootstrapCommands = toArrayString(v)
		}
		if v, ok := in["override_bootstrap_command"].(string); ok && len(v) > 0 {
			obj.OverrideBootstrapCommand = v
		}
		if v, ok := in["disable_imdsv1"].(bool); ok {
			obj.DisableIMDSv1 = &v
		}
		if v, ok := in["disable_pods_imds"].(bool); ok {
			obj.DisablePodIMDS = &v
		}
		if v, ok := in["placement"].([]interface{}); ok && len(v) > 0 {
			obj.Placement = expandNodeGroupPlacement(v)
		}
		if v, ok := in["efa_enabled"].(bool); ok {
			obj.EFAEnabled = &v
		}
		if v, ok := in["instance_selector"].([]interface{}); ok && len(v) > 0 {
			obj.InstanceSelector = expandNodeGroupInstanceSelector(v)
		}
		//additional encrypted volume field not in spec

		if v, ok := in["bottle_rocket"].([]interface{}); ok && len(v) > 0 {
			obj.Bottlerocket = expandNodeGroupBottleRocket(v)
		}
		//doc does not have fields custom ami, enable detailed monitoring, or is wavlength zone but NodeGroupbase struct does (says to remove)
		if v, ok := in["enable_detailed_monitoring"].(bool); ok {
			obj.EnableDetailedMonitoring = &v
		}
		if v, ok := in["instance_types"].([]interface{}); ok && len(v) > 0 {
			obj.InstanceTypes = toArrayString(v)
		}
		if v, ok := in["spot"].(bool); ok {
			obj.Spot = &v
		}
		if v, ok := in["taints"].([]interface{}); ok && len(v) > 0 {
			obj.Taints = expandManagedNodeGroupTaints(v)
		}
		if v, ok := in["update_config"].([]interface{}); ok && len(v) > 0 {
			obj.UpdateConfig = expandNodeGroupUpdateConfig(v)
		}
		if v, ok := in["launch_template"].([]interface{}); ok && len(v) > 0 {
			obj.LaunchTemplate = expandManagedNodeGroupLaunchTempelate(v)
		}
		if v, ok := in["version"].(string); ok && len(v) > 0 {
			obj.Version = v
		}
		//@@@TODO:
		//struct has field ReleaseVersion
		//also has internal field unowned -> will leave blank for now
		//how do i finish this?

		//check if this is how to build array of pointers
		//out[i] = obj
		outToSort[i] = *obj
	}

	sort.Sort(ByManagedNodeGroupName(outToSort))
	for i := range outToSort {
		out[i] = &outToSort[i]
	}

	return out
}

// expand managed node group taints function (completed) (can i use this to expand taints in node group?)
func expandManagedNodeGroupTaints(p []interface{}) []NodeGroupTaint {

	out := make([]NodeGroupTaint, len(p))
	if len(p) == 0 || p[0] == nil {
		return out
	}

	for i := range p {
		in := p[i].(map[string]interface{})
		obj := &NodeGroupTaint{}
		if v, ok := in["key"].(string); ok && len(v) > 0 {
			obj.Key = v
		}
		if v, ok := in["value"].(string); ok && len(v) > 0 {
			obj.Value = v
		}
		if v, ok := in["effect"].(string); ok && len(v) > 0 {
			obj.Effect = v
		}
		out[i] = *obj
	}
	//docs dont have field skip endpoint creation but struct does
	return out
}

// expand managed node group Launch Tempelate function (completed)
func expandManagedNodeGroupLaunchTempelate(p []interface{}) *LaunchTemplate {
	obj := &LaunchTemplate{}

	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})
	if v, ok := in["id"].(string); ok && len(v) > 0 {
		obj.ID = v
	}
	if v, ok := in["version"].(string); ok && len(v) > 0 {
		obj.Version = v
	}
	//docs dont have field skip endpoint creation but struct does
	return obj
}

func expandNodeGroups(p []interface{}) []*NodeGroup { //not completed have questions in comments
	out := make([]*NodeGroup, len(p))
	outToSort := make([]NodeGroup, len(p))

	if len(p) == 0 || p[0] == nil {
		return out
	}

	for i := range p {
		in := p[i].(map[string]interface{})
		obj := NodeGroup{}
		log.Println("expand_nodegroups")
		log.Println("ngs_yaml name: ", in["name"].(string))
		if v, ok := in["name"].(string); ok && len(v) > 0 {
			log.Println("ngs_name: ", v)
			//obj.Name = "bob"
			//log.Println("obj name: ", obj.Name)
			obj.Name = v
		}
		if v, ok := in["ami_family"].(string); ok && len(v) > 0 {
			obj.AMIFamily = v
		}
		if v, ok := in["instance_type"].(string); ok && len(v) > 0 {
			obj.InstanceType = v
		}
		if v, ok := in["availability_zones"].([]interface{}); ok && len(v) > 0 {
			obj.AvailabilityZones = toArrayStringSorted(v)
		}
		if v, ok := in["subnets"].([]interface{}); ok && len(v) > 0 {
			obj.Subnets = toArrayString(v)
		}
		if v, ok := in["instance_prefix"].(string); ok && len(v) > 0 {
			obj.InstancePrefix = v
		}
		if v, ok := in["instance_name"].(string); ok && len(v) > 0 {
			obj.InstanceName = v
		}
		if v, ok := in["desired_capacity"].(int); ok {
			obj.DesiredCapacity = &v
		}
		if v, ok := in["min_size"].(int); ok {
			obj.MinSize = &v
		}
		if v, ok := in["max_size"].(int); ok {
			obj.MaxSize = &v
		}
		if v, ok := in["volume_size"].(int); ok {
			obj.VolumeSize = &v
		}
		if v, ok := in["ssh"].([]interface{}); ok && len(v) > 0 {
			obj.SSH = expandNodeGroupSsh(v, false)
		}
		if v, ok := in["labels"].(map[string]interface{}); ok && len(v) > 0 {
			obj.Labels = toMapString(v)
		}
		if v, ok := in["private_networking"].(bool); ok {
			obj.PrivateNetworking = &v
		}
		if v, ok := in["tags"].(map[string]interface{}); ok && len(v) > 0 {
			obj.Tags = toMapString(v)
		}
		if v, ok := in["iam"].([]interface{}); ok && len(v) > 0 {
			obj.IAM = expandNodeGroupIam(v)
		}
		if v, ok := in["ami"].(string); ok && len(v) > 0 {
			obj.AMI = v
		}
		if v, ok := in["security_groups"].([]interface{}); ok && len(v) > 0 {
			obj.SecurityGroups = expandNodeGroupSecurityGroups(v)
		}
		if v, ok := in["max_pods_per_node"].(int); ok {
			obj.MaxPodsPerNode = v
		}
		if v, ok := in["asg_suspend_process"].([]interface{}); ok && len(v) > 0 {
			obj.ASGSuspendProcesses = toArrayString(v)
		}
		if v, ok := in["ebs_optimized"].(bool); ok {
			obj.EBSOptimized = &v
		}
		if v, ok := in["volume_type"].(string); ok && len(v) > 0 {
			obj.VolumeType = v
		}
		if v, ok := in["volume_name"].(string); ok && len(v) > 0 {
			obj.VolumeName = v
		}
		if v, ok := in["volume_encrypted"].(bool); ok {
			obj.VolumeEncrypted = &v
		}
		if v, ok := in["volume_kms_key_id"].(string); ok && len(v) > 0 {
			obj.VolumeKmsKeyID = v
		}
		if v, ok := in["volume_iops"].(int); ok && v != 0 {
			obj.VolumeIOPS = &v
		}
		if v, ok := in["volume_throughput"].(int); ok && v != 0 {
			obj.VolumeThroughput = &v
		}
		if v, ok := in["pre_bootstrap_commands"].([]interface{}); ok && len(v) > 0 {
			obj.PreBootstrapCommands = toArrayString(v)
		}
		if v, ok := in["override_bootstrap_command"].(string); ok && len(v) > 0 {
			obj.OverrideBootstrapCommand = v
		}
		if v, ok := in["disable_imdsv1"].(bool); ok {
			obj.DisableIMDSv1 = &v
		}
		if v, ok := in["disable_pods_imds"].(bool); ok {
			obj.DisablePodIMDS = &v
		}
		if v, ok := in["placement"].([]interface{}); ok && len(v) > 0 {
			obj.Placement = expandNodeGroupPlacement(v)
		}
		if v, ok := in["efa_enabled"].(bool); ok {
			obj.EFAEnabled = &v
		}
		if v, ok := in["instance_selector"].([]interface{}); ok && len(v) > 0 {
			obj.InstanceSelector = expandNodeGroupInstanceSelector(v)
		}
		//additional encrypted volume field not in spec

		if v, ok := in["bottle_rocket"].([]interface{}); ok && len(v) > 0 {
			obj.Bottlerocket = expandNodeGroupBottleRocket(v)
		}
		//doc does not have fields custom ami, enable detailed monitoring, or is wavlength zone but NodeGroupbase struct does

		if v, ok := in["enable_detailed_monitoring"].(bool); ok {
			obj.EnableDetailedMonitoring = &v
		}
		if v, ok := in["instances_distribution"].([]interface{}); ok && len(v) > 0 {
			obj.InstancesDistribution = expandNodeGroupInstanceDistribution(v)
		}
		if v, ok := in["asg_metrics_collection"].([]interface{}); ok && len(v) > 0 {
			obj.ASGMetricsCollection = expandNodeGroupASGMetricCollection(v)
		}
		if v, ok := in["cpu_credits"].(string); ok && len(v) > 0 {
			obj.CPUCredits = v
		}
		if v, ok := in["classic_load_balancer_names"].([]interface{}); ok && len(v) > 0 {
			obj.ClassicLoadBalancerNames = toArrayString(v)
		}
		if v, ok := in["target_group_arns"].([]interface{}); ok && len(v) > 0 {
			obj.TargetGroupARNs = toArrayString(v)
		}
		if v, ok := in["taints"].([]interface{}); ok && len(v) > 0 {
			obj.Taints = expandManagedNodeGroupTaints(v)
		}

		if v, ok := in["update_config"].([]interface{}); ok && len(v) > 0 {
			obj.UpdateConfig = expandNodeGroupUpdateConfig(v)
		}
		if v, ok := in["cluster_dns"].(string); ok && len(v) > 0 {
			obj.ClusterDNS = v
		}
		//@@@TODO Store terraform input as inline document object correctly
		if v, ok := in["kubelet_extra_config"].([]interface{}); ok && len(v) > 0 {
			obj.KubeletExtraConfig = expandKubeletExtraConfig(v)
		}
		//@@@
		if v, ok := in["version"].(string); ok && len(v) > 0 {
			obj.Version = v
		}
		//struct has field containerRuntime
		//doc has version and subnet cidr
		//how do i finish this?

		//check if this is how to build array of pointers
		outToSort[i] = obj
	}

	sort.Sort(ByNodeGroupName(outToSort))
	for i := range outToSort {
		out[i] = &outToSort[i]
	}

	return out
}

// @@expand KubeletExtraConfig function (completed)
func expandKubeletExtraConfig(p []interface{}) *KubeletExtraConfig {
	obj := &KubeletExtraConfig{}

	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})
	if v, ok := in["kube_reserved"].(map[string]interface{}); ok && len(v) > 0 {
		obj.KubeReserved = toMapString(v)
	}
	if v, ok := in["kube_reserved_cgroup"].(string); ok && len(v) > 0 {
		obj.KubeReservedCGroup = v
	}
	if v, ok := in["system_reserved"].(map[string]interface{}); ok && len(v) > 0 {
		obj.SystemReserved = toMapString(v)
	}
	if v, ok := in["eviction_hard"].(map[string]interface{}); ok && len(v) > 0 {
		obj.EvictionHard = toMapString(v)
	}
	if v, ok := in["feature_gates"].(map[string]interface{}); ok && len(v) > 0 {
		obj.FeatureGates = toMapBool(v)
	}

	return obj
}

// expand node group Update Config function (completed)
func expandNodeGroupUpdateConfig(p []interface{}) *NodeGroupUpdateConfig {
	obj := &NodeGroupUpdateConfig{}

	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})
	if v, ok := in["max_unavaliable"].(int); ok {
		obj.MaxUnavailable = &v
	}
	if v, ok := in["max_unavaliable_percetage"].(int); ok {
		obj.MaxUnavailablePercentage = &v
	}
	//docs dont have field skip endpoint creation but struct does
	return obj
}

// expand node group ASG Metrics Collection function (completed)
func expandNodeGroupASGMetricCollection(p []interface{}) []MetricsCollection {
	out := make([]MetricsCollection, len(p))
	if len(p) == 0 || p[0] == nil {
		return out
	}
	for i := range p {
		obj := &MetricsCollection{}
		in := p[0].(map[string]interface{})
		if v, ok := in["granularity"].(string); ok && len(v) > 0 {
			obj.Granularity = v
		}
		if v, ok := in["metrics"].([]interface{}); ok && len(v) > 0 {
			obj.Metrics = toArrayString(v)
		}
		out[i] = *obj
	}

	return out
}

// expand node group Instance Distribution function (completed)
func expandNodeGroupInstanceDistribution(p []interface{}) *NodeGroupInstancesDistribution {
	obj := &NodeGroupInstancesDistribution{}

	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})
	if v, ok := in["instance_types"].([]interface{}); ok && len(v) > 0 {
		obj.InstanceTypes = toArrayString(v)
	}
	if v, ok := in["max_price"].(float64); ok {
		obj.MaxPrice = &v
	}
	if v, ok := in["on_demand_base_capacity"].(int); ok {
		obj.OnDemandBaseCapacity = &v
	}
	if v, ok := in["on_demand_percentage_above_base_capacity"].(int); ok {
		obj.OnDemandPercentageAboveBaseCapacity = &v
	}
	if v, ok := in["spot_instance_pools"].(int); ok {
		obj.SpotInstancePools = &v
	}
	if v, ok := in["spot_allocation_strategy"].(string); ok {
		obj.SpotAllocationStrategy = v
	}
	if v, ok := in["capacity_rebalance"].(bool); ok {
		obj.CapacityRebalance = &v
	}
	return obj
}

// expand node group Bottle Rocket function (completed)
func expandNodeGroupBottleRocket(p []interface{}) *NodeGroupBottlerocket {
	obj := &NodeGroupBottlerocket{}

	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})
	if v, ok := in["enable_admin_container"].(bool); ok {
		obj.EnableAdminContainer = &v
	}
	////@@@TODO Store terraform input as inline document object correctly
	if v, ok := in["settings"].(string); ok && len(v) > 0 {
		var policyDoc map[string]interface{}
		var json2 = jsoniter.ConfigCompatibleWithStandardLibrary
		//json.Unmarshal(input, &data)
		json2.Unmarshal([]byte(v), &policyDoc)
		obj.Settings = policyDoc
		log.Println("bottle rocket settings expanded correct")
	}
	//docs dont have field skip endpoint creation but struct does
	return obj
}

// expand node group instance selector function (completed)
func expandNodeGroupInstanceSelector(p []interface{}) *InstanceSelector {
	obj := &InstanceSelector{}

	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})
	if v, ok := in["vcpus"].(int); ok {
		obj.VCPUs = &v
	}
	if v, ok := in["memory"].(string); ok && len(v) > 0 {
		obj.Memory = v
	}
	if v, ok := in["gpus"].(int); ok {
		obj.GPUs = &v
	}
	if v, ok := in["cpu_architecture"].(string); ok && len(v) > 0 {
		obj.CPUArchitecture = v
	}

	return obj
}

// expand node group placement function (completed)
func expandNodeGroupPlacement(p []interface{}) *Placement {
	obj := &Placement{}

	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})
	if v, ok := in["group"].(string); ok && len(v) > 0 {
		obj.GroupName = v
	}
	return obj
}

// expand node group security groups function (completed)
func expandNodeGroupSecurityGroups(p []interface{}) *NodeGroupSGs {
	obj := &NodeGroupSGs{}

	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})
	if v, ok := in["attach_ids"].([]interface{}); ok && len(v) > 0 {
		obj.AttachIDs = toArrayString(v)
	}
	if v, ok := in["with_shared"].(bool); ok {
		obj.WithShared = &v
	}
	if v, ok := in["with_local"].(bool); ok {
		obj.WithLocal = &v
	}
	return obj
}

func expandManagedNodeGroupSecurityGroups(p []interface{}, rawConfig cty.Value) *NodeGroupSGs {
	obj := &NodeGroupSGs{}

	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})
	rawConfig = rawConfig.AsValueSlice()[0]

	if v, ok := in["attach_ids"].([]interface{}); ok && len(v) > 0 {
		obj.AttachIDs = toArrayString(v)
	}

	rawWithShared := rawConfig.GetAttr("with_shared")
	if !rawWithShared.IsNull() {
		boolVal := rawWithShared.True()
		obj.WithShared = &boolVal
	}

	rawWithLocal := rawConfig.GetAttr("with_local")
	if !rawWithLocal.IsNull() {
		boolVal := rawWithLocal.True()
		obj.WithLocal = &boolVal
	}

	return obj
}

// expand node group iam function (completed/kind of)
func expandNodeGroupIam(p []interface{}) *NodeGroupIAM {
	obj := &NodeGroupIAM{}

	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})
	//@@@TODO Store terraform input as inline document object correctly
	if v, ok := in["attach_policy"].([]interface{}); ok && len(v) > 0 {
		obj.AttachPolicy = expandAttachPolicy(v)
	}

	if v, ok := in["attach_policy_arns"].([]interface{}); ok && len(v) > 0 {
		obj.AttachPolicyARNs = toArrayString(v)
	}
	if v, ok := in["instance_profile_arn"].(string); ok && len(v) > 0 {
		obj.InstanceProfileARN = v
	}
	if v, ok := in["instance_role_arn"].(string); ok && len(v) > 0 {
		obj.InstanceRoleARN = v
	}
	if v, ok := in["instance_role_name"].(string); ok && len(v) > 0 {
		obj.InstanceRoleName = v
	}
	if v, ok := in["instance_role_permission_boundary"].(string); ok && len(v) > 0 {
		obj.InstanceRolePermissionsBoundary = v
	}
	if v, ok := in["iam_node_group_with_addon_policies"].([]interface{}); ok && len(v) > 0 {
		obj.WithAddonPolicies = expandNodeGroupIAMWithAddonPolicies(v)
	}
	return obj
}

// expand attach policy (completed)@@@
func expandStatement(p []interface{}) InlineStatement {
	obj := InlineStatement{}

	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})
	if v, ok := in["effect"].(string); ok && len(v) > 0 {
		obj.Effect = v
	}
	if v, ok := in["action"].([]interface{}); ok && len(v) > 0 {
		obj.Action = toArrayStringSorted(v)
	}
	if v, ok := in["resource"].(string); ok && len(v) > 0 {
		obj.Resource = v
	}
	return obj
}

// expand attach policy (completed)
func expandAttachPolicy(p []interface{}) *InlineDocument {
	obj := InlineDocument{}

	if len(p) == 0 || p[0] == nil {
		return &obj
	}
	in := p[0].(map[string]interface{})
	if v, ok := in["version"].(string); ok && len(v) > 0 {
		obj.Version = v
	}
	if v, ok := in["statement"].([]interface{}); ok && len(v) > 0 {
		obj.Statement = expandStatement(v)
	}
	return &obj
}

// expand node group IAm With Addon Policies function (completed/kind of)
func expandNodeGroupIAMWithAddonPolicies(p []interface{}) *NodeGroupIAMAddonPolicies {
	obj := NodeGroupIAMAddonPolicies{}

	if len(p) == 0 || p[0] == nil {
		return &obj
	}
	in := p[0].(map[string]interface{})
	n1 := spew.Sprintf("%+v", in)
	log.Println("expandNodeGroupIAMWithAddonPolicies: ", n1)
	if v, ok := in["image_builder"].(bool); ok {
		obj.ImageBuilder = &v
	}
	if v, ok := in["auto_scaler"].(bool); ok {
		obj.AutoScaler = &v
	}
	if v, ok := in["external_dns"].(bool); ok {
		obj.ExternalDNS = &v
	}
	if v, ok := in["cert_manager"].(bool); ok {
		obj.CertManager = &v
	}
	if v, ok := in["app_mesh"].(bool); ok {
		obj.AppMesh = &v
	}
	if v, ok := in["app_mesh_review"].(bool); ok {
		obj.AppMeshPreview = &v
	}
	if v, ok := in["ebs"].(bool); ok {
		obj.EBS = &v
	}
	if v, ok := in["fsx"].(bool); ok {
		obj.FSX = &v
	}
	if v, ok := in["efs"].(bool); ok {
		obj.EFS = &v
	}
	// @@@@ doc says it should be field alb_ingress,
	// struct has field ABSLoadBalancerController?
	if v, ok := in["alb_ingress"].(bool); ok {
		obj.AWSLoadBalancerController = &v
	}

	if v, ok := in["xray"].(bool); ok {
		obj.XRay = &v
	}
	if v, ok := in["cloud_watch"].(bool); ok {
		obj.CloudWatch = &v
	}
	n2 := spew.Sprintf("%+v", obj)
	log.Println("expandNodeGroupIAMWithAddonPolicies obj: ", n2)
	return &obj
}

// expand node group ssh function (completed/ kind of)
func expandNodeGroupSsh(p []interface{}, managed bool) *NodeGroupSSH {
	obj := &NodeGroupSSH{}

	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})
	if v, ok := in["allow"].(bool); ok {
		obj.Allow = &v
	}
	//struct has publicKeypath when the doc does not
	if v, ok := in["public_key"].(string); ok && len(v) > 0 {
		obj.PublicKey = v
	}
	if v, ok := in["public_key_name"].(string); ok && len(v) > 0 {
		obj.PublicKeyName = v
	}
	if v, ok := in["source_security_group_ids"].([]interface{}); ok && len(v) > 0 {
		obj.SourceSecurityGroupIDs = toArrayString(v)
	}
	// Deprecated but still valid to use this API till an alterative is found!

	if v, ok := in["enable_ssm"].(bool); ok && !managed {
		obj.EnableSSM = &v
	}
	//docs dont have field skip endpoint creation but struct does
	return obj
}

// expand private clusters function (completed)
func expandPrivateCluster(p []interface{}) *PrivateCluster {
	obj := &PrivateCluster{}

	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})
	if v, ok := in["enabled"].(bool); ok {
		obj.Enabled = &v
	}
	if v, ok := in["skip_endpoint_creation"].(bool); ok {
		obj.SkipEndpointCreation = &v
	}
	if v, ok := in["additional_endpoint_services"].([]interface{}); ok && len(v) > 0 {
		obj.AdditionalEndpointServices = toArrayString(v)
	}
	//docs dont have field skip endpoint creation but struct does
	return obj
}

// expand addon(completed/kind of)
func expandAddons(p []interface{}) []*Addon { //checkhow to return a []*
	out := make([]*Addon, len(p))
	if len(p) == 0 || p[0] == nil {
		return out
	}
	for i := range p {
		obj := &Addon{}
		in := p[i].(map[string]interface{})
		if v, ok := in["name"].(string); ok && len(v) > 0 {
			obj.Name = v
		}
		if v, ok := in["version"].(string); ok && len(v) > 0 {
			obj.Version = v
		}
		if v, ok := in["service_account_role_arn"].(string); ok && len(v) > 0 {
			obj.ServiceAccountRoleARN = v
		}
		if v, ok := in["attach_policy_arns"].([]interface{}); ok && len(v) > 0 {
			obj.AttachPolicyARNs = toArrayString(v)
		}

		//@@@TODO Store terraform input as inline document object correctly
		if v, ok := in["attach_policy"].([]interface{}); ok && len(v) > 0 {
			obj.AttachPolicy = expandAttachPolicy(v)
		}
		if v, ok := in["permissions_boundary"].(string); ok && len(v) > 0 {
			obj.PermissionsBoundary = v
		}
		if v, ok := in["well_known_policies"].([]interface{}); ok && len(v) > 0 {
			obj.WellKnownPolicies = expandIAMWellKnownPolicies(v)
		}
		if v, ok := in["tags"].(map[string]interface{}); ok && len(v) > 0 {
			obj.Tags = toMapString(v)
		}
		if v, ok := in["configuration_values"].(string); ok && len(v) > 0 {
			obj.ConfigurationValues = v
		}
		//docs dont have force variable but struct does
		out[i] = obj
	}
	return out
}

// expand vpc function
func expandVPC(p []interface{}, rawConfig cty.Value) *EKSClusterVPC {
	obj := &EKSClusterVPC{}

	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})
	rawConfig = rawConfig.AsValueSlice()[0]

	if v, ok := in["id"].(string); ok && len(v) > 0 {
		obj.ID = v
	}
	if v, ok := in["cidr"].(string); ok && len(v) > 0 {
		obj.CIDR = v
	}
	if v, ok := in["ipv6_cidr"].(string); ok && len(v) > 0 {
		obj.IPv6Cidr = v
	}
	if v, ok := in["ipv6_pool"].(string); ok && len(v) > 0 {
		obj.IPv6Pool = v
	}
	if v, ok := in["security_group"].(string); ok && len(v) > 0 {
		obj.SecurityGroup = v
	}
	if v, ok := in["subnets"].([]interface{}); ok && len(v) > 0 {
		obj.Subnets = expandSubnets(v)
	}
	if v, ok := in["extra_ipv6_cidrs"].([]interface{}); ok && len(v) > 0 {
		obj.ExtraIPv6CIDRs = toArrayString(v)
	}
	if v, ok := in["extra_cidrs"].([]interface{}); ok && len(v) > 0 {
		obj.ExtraCIDRs = toArrayString(v)
	}
	if v, ok := in["shared_node_security_group"].(string); ok && len(v) > 0 {
		obj.SharedNodeSecurityGroup = v
	}
	rawManageSharedNodeSecurityGroupRules := rawConfig.GetAttr("manage_shared_node_security_group_rules")
	if !rawManageSharedNodeSecurityGroupRules.IsNull() {
		boolVal := rawManageSharedNodeSecurityGroupRules.True()
		obj.ManageSharedNodeSecurityGroupRules = &boolVal
	}
	if v, ok := in["auto_allocate_ipv6"].(bool); ok {
		obj.AutoAllocateIPv6 = &v
	}
	if v, ok := in["nat"].([]interface{}); ok && len(v) > 0 {
		obj.NAT = expandNat(v)
	}
	if v, ok := in["cluster_endpoints"].([]interface{}); ok && len(v) > 0 {
		obj.ClusterEndpoints = expandClusterEndpoints(v)
	}
	if v, ok := in["public_access_cidrs"].([]interface{}); ok && len(v) > 0 {
		obj.PublicAccessCIDRs = toArrayString(v)
	}
	return obj
}

func expandClusterEndpoints(p []interface{}) *ClusterEndpoints {
	obj := &ClusterEndpoints{}

	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})
	if v, ok := in["private_access"].(bool); ok {
		obj.PrivateAccess = &v
	}
	if v, ok := in["public_access"].(bool); ok {
		obj.PublicAccess = &v
	}
	return obj
}

func expandNat(p []interface{}) *ClusterNAT {
	obj := &ClusterNAT{}

	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})
	if v, ok := in["gateway"].(string); ok && len(v) > 0 {
		obj.Gateway = v
	}
	return obj
}

func expandSubnets(p []interface{}) *ClusterSubnets {
	obj := &ClusterSubnets{}

	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})
	if v, ok := in["private"].([]interface{}); ok && len(v) > 0 {
		obj.Private = expandSubnetSpec(v)
	}
	if v, ok := in["public"].([]interface{}); ok && len(v) > 0 {
		obj.Public = expandSubnetSpec(v)
	}
	return obj
}
func expandSubnetSpec(p []interface{}) AZSubnetMapping {
	obj := make(AZSubnetMapping)

	if len(p) == 0 || p[0] == nil {
		return obj
	}

	for i := range p {
		in := p[i].(map[string]interface{})
		elem2 := AZSubnetSpec{}
		if v, ok := in["id"].(string); ok && len(v) > 0 {
			elem2.ID = v
		}
		if v, ok := in["az"].(string); ok && len(v) > 0 {
			elem2.AZ = v
		}
		if v, ok := in["cidr"].(string); ok && len(v) > 0 {
			elem2.CIDR = v
		}
		if v, ok := in["name"].(string); ok && len(v) > 0 {
			obj[v] = elem2
		}
	}
	return obj
}

// struct IdentityProviders has one extra field not in documentation or the schema
func expandIdentityProviders(p []interface{}) []*IdentityProvider {
	out := make([]*IdentityProvider, len(p))
	if len(p) == 0 || p[0] == nil {
		return out
	}
	for i := range p {
		in := p[i].(map[string]interface{})
		obj := &IdentityProvider{}
		if v, ok := in["type"].(string); ok && len(v) > 0 {
			obj.Type = v
		}
		out[i] = obj
	}
	//obj.type_ = in["type"].(string)
	return out
}

func expandIAMFields(p []interface{}) *EKSClusterIAM {
	obj := &EKSClusterIAM{}

	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["service_role_arn"].(string); ok && len(v) > 0 {
		obj.ServiceRoleARN = v
	}

	if v, ok := in["service_role_permission_boundary"].(string); ok && len(v) > 0 {
		obj.ServiceRolePermissionsBoundary = v
	}

	if v, ok := in["fargate_pod_execution_role_arn"].(string); ok && len(v) > 0 {
		obj.FargatePodExecutionRoleARN = v
	}

	if v, ok := in["fargate_pod_execution_permissions_boundary"].(string); ok && len(v) > 0 {
		obj.FargatePodExecutionRolePermissionsBoundary = v
	}

	if v, ok := in["with_oidc"].(bool); ok {
		obj.WithOIDC = &v
	}

	if v, ok := in["service_accounts"].([]interface{}); ok && len(v) > 0 {
		obj.ServiceAccounts = expandIAMServiceAccountsConfig(v)
	}

	if v, ok := in["vpcResourceControllerPolicy"].(bool); ok {
		obj.VPCResourceControllerPolicy = &v
	}

	return obj
}

func expandServiceAccountsMetadata(p []interface{}) *EKSClusterIAMMeta {
	obj := &EKSClusterIAMMeta{}

	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	//is this okay or do i need to store it in metadata, golang gives me access to the contents inside the metadata struct
	if v, ok := in["name"].(string); ok && len(v) > 0 {
		obj.Name = v
	}
	if v, ok := in["namespace"].(string); ok && len(v) > 0 {
		obj.Namespace = v
	}
	if v, ok := in["labels"].(map[string]interface{}); ok && len(v) > 0 {
		obj.Labels = toMapString(v)
	}
	if v, ok := in["annotations"].(map[string]interface{}); ok && len(v) > 0 {
		obj.Annotations = toMapString(v)
	}
	return obj
}

func expandIAMServiceAccountsConfig(p []interface{}) []*EKSClusterIAMServiceAccount {
	out := make([]*EKSClusterIAMServiceAccount, len(p))
	if len(p) == 0 || p[0] == nil {
		return out
	}
	for i := range p {
		obj := &EKSClusterIAMServiceAccount{}
		in := p[i].(map[string]interface{})
		if v, ok := in["metadata"].([]interface{}); ok && len(v) > 0 {
			obj.Metadata = expandServiceAccountsMetadata(v)
		}
		//finish clusterIAM metadata
		if v, ok := in["attach_policy_arns"].([]interface{}); ok && len(v) > 0 {
			obj.AttachPolicyARNs = toArrayString(v)
		}
		if v, ok := in["well_known_policies"].([]interface{}); ok && len(v) > 0 {
			obj.WellKnownPolicies = expandIAMWellKnownPolicies(v)
		}
		//check for attach policy
		////@@@TODO Store terraform input as inline document object correctly
		if v, ok := in["attach_policy"].(string); ok && len(v) > 0 {
			var policyDoc map[string]interface{}
			var json2 = jsoniter.ConfigCompatibleWithStandardLibrary
			//json.Unmarshal(input, &data)
			json2.Unmarshal([]byte(v), &policyDoc)
			obj.AttachPolicy = policyDoc
			log.Println("attach policy expanded correct")
		}
		if v, ok := in["attach_role_arn"].(string); ok && len(v) > 0 {
			obj.AttachRoleARN = v
		}
		if v, ok := in["permissions_boundary"].(string); ok && len(v) > 0 {
			obj.PermissionsBoundary = v
		}
		if v, ok := in["status"].([]interface{}); ok && len(v) > 0 {
			obj.Status = expandIAMServiceAccountsStatusConfig(v)
		}
		if v, ok := in["role_name"].(string); ok && len(v) > 0 {
			obj.RoleName = v
		}
		if v, ok := in["role_only"].(bool); ok {
			obj.RoleOnly = &v
		}
		if v, ok := in["tags"].(map[string]interface{}); ok && len(v) > 0 {
			obj.Tags = toMapString(v)
		}
		out[i] = obj
	}

	return out
}

func expandIAMWellKnownPolicies(p []interface{}) *WellKnownPolicies {
	obj := WellKnownPolicies{}

	if len(p) == 0 || p[0] == nil {
		return nil
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["image_builder"].(bool); ok {
		obj.ImageBuilder = &v
	}
	if v, ok := in["auto_scaler"].(bool); ok {
		obj.AutoScaler = &v
	}
	if v, ok := in["aws_load_balancer_controller"].(bool); ok {
		obj.AWSLoadBalancerController = &v
	}
	if v, ok := in["external_dns"].(bool); ok {
		obj.ExternalDNS = &v
	}
	if v, ok := in["cert_manager"].(bool); ok {
		obj.CertManager = &v
	}
	if v, ok := in["ebs_csi_controller"].(bool); ok {
		obj.EBSCSIController = &v
	}
	if v, ok := in["efs_csi_controller"].(bool); ok {
		obj.EFSCSIController = &v
	}

	return &obj
}

func expandIAMServiceAccountsStatusConfig(p []interface{}) *ClusterIAMServiceAccountStatus {
	obj := &ClusterIAMServiceAccountStatus{}

	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["role_arn"].(string); ok && len(v) > 0 {
		obj.RoleARN = v
	}
	//obj.RoleARN = in["roleARN"].(string)
	return obj
}

func expandKubernetesNetworkConfig(p []interface{}) *KubernetesNetworkConfig {
	obj := &KubernetesNetworkConfig{}

	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["ip_family"].(string); ok && len(v) > 0 {
		obj.IPFamily = v
	}
	if v, ok := in["service_ipv4_cidr"].(string); ok && len(v) > 0 {
		obj.ServiceIPv4CIDR = v
	}
	//obj.ServiceIPv4CIDR = in["serviceIPv4CIDR"].(string)
	return obj
}

func expandEKSClusterSpecConfig(p []interface{}) *EKSSpec {
	obj := &EKSSpec{}
	log.Println("expandClusterSpec")

	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["type"].(string); ok && len(v) > 0 {
		obj.Type = v
	}
	if v, ok := in["blueprint"].(string); ok && len(v) > 0 {
		obj.Blueprint = v
	}
	if v, ok := in["blueprint_version"].(string); ok && len(v) > 0 {
		obj.BlueprintVersion = v
	}
	if v, ok := in["cloud_provider"].(string); ok && len(v) > 0 {
		obj.CloudProvider = v
	}
	if v, ok := in["cross_account_role_arn"].(string); ok && len(v) > 0 {
		obj.CrossAccountRoleArn = v
	}
	if v, ok := in["cni_provider"].(string); ok && len(v) > 0 {
		obj.CniProvider = v
	}
	if v, ok := in["cni_params"].([]interface{}); ok && len(v) > 0 {
		obj.CniParams = expandCNIParams(v)
	}
	if v, ok := in["proxy_config"].([]interface{}); ok && len(v) > 0 {
		obj.ProxyConfig = expandProxyConfig(v)
	}
	if v, ok := in["system_components_placement"].([]interface{}); ok && len(v) > 0 {
		obj.SystemComponentsPlacement = expandSystemComponentsPlacement(v)
	}
	if v, ok := in["sharing"].([]interface{}); ok && len(v) > 0 {
		obj.Sharing = expandEKSClusterSharing(v)
	}
	log.Println("cluster spec cloud_provider: ", obj.CloudProvider)

	return obj
}

func expandEKSClusterSharing(p []interface{}) *EKSClusterSharing {
	obj := &EKSClusterSharing{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}

	in := p[0].(map[string]interface{})
	if v, ok := in["enabled"].(bool); ok {
		obj.Enabled = &v
	}
	if v, ok := in["projects"].([]interface{}); ok && len(v) > 0 {
		obj.Projects = expandEKSClusterSharingProjects(v)
	}
	return obj
}

func expandEKSClusterSharingProjects(p []interface{}) []*EKSClusterSharingProject {
	if len(p) == 0 {
		return nil
	}
	var res []*EKSClusterSharingProject
	for i := range p {
		in := p[i].(map[string]interface{})
		obj := &EKSClusterSharingProject{}
		if v, ok := in["name"].(string); ok && len(v) > 0 {
			obj.Name = v
		}
		res = append(res, obj)
	}
	return res
}

func expandProxyConfig(p []interface{}) *ProxyConfig {
	obj := &ProxyConfig{}
	log.Println("expandProxyConfig")

	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["http_proxy"].(string); ok && len(v) > 0 {
		obj.HttpProxy = v
	}

	if v, ok := in["https_proxy"].(string); ok && len(v) > 0 {
		obj.HttpsProxy = v
	}

	if v, ok := in["no_proxy"].(string); ok && len(v) > 0 {
		obj.NoProxy = v
	}

	if v, ok := in["proxy_auth"].(string); ok && len(v) > 0 {
		obj.ProxyAuth = v
	}

	if v, ok := in["bootstrap_ca"].(string); ok && len(v) > 0 {
		obj.BootstrapCA = v
	}

	if v, ok := in["enabled"].(bool); ok {
		obj.Enabled = &v
	}

	if v, ok := in["allow_insecure_bootstrap"].(bool); ok {
		obj.AllowInsecureBootstrap = &v
	}

	return obj

}

func expandSystemComponentsPlacement(p []interface{}) *SystemComponentsPlacement {
	obj := &SystemComponentsPlacement{}
	log.Println("expandSystemComponentsPlacement")

	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["node_selector"].(map[string]interface{}); ok && len(v) > 0 {
		obj.NodeSelector = toMapString(v)
	}

	if v, ok := in["tolerations"].([]interface{}); ok && len(v) > 0 {
		obj.Tolerations = expandTolerations(v)
	}
	if v, ok := in["daemonset_override"].([]interface{}); ok && len(v) > 0 {
		obj.DaemonsetOverride = expandDaemonsetOverride(v)
	}
	return obj
}

func expandTolerations(p []interface{}) []*Tolerations {
	out := make([]*Tolerations, len(p))
	if len(p) == 0 || p[0] == nil {
		return out
	}
	for i := range p {
		obj := &Tolerations{}
		in := p[i].(map[string]interface{})

		if v, ok := in["key"].(string); ok && len(v) > 0 {
			obj.Key = v
		}
		if v, ok := in["operator"].(string); ok && len(v) > 0 {
			obj.Operator = v
		}
		if v, ok := in["value"].(string); ok && len(v) > 0 {
			obj.Value = v
		}
		if v, ok := in["effect"].(string); ok && len(v) > 0 {
			obj.Effect = v
		}
		if v, ok := in["toleration_seconds"].(int); ok {
			if v == 0 {
				obj.TolerationSeconds = nil
			} else {
				log.Println("setting toleration seconds")
				obj.TolerationSeconds = &v
			}
		}
		out[i] = obj
	}
	return out
}

func expandDaemonsetOverride(p []interface{}) *DaemonsetOverride {
	obj := &DaemonsetOverride{}
	log.Println("expand CNI params")

	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["node_selection_enabled"].(bool); ok {
		obj.NodeSelectionEnabled = &v
	}
	if v, ok := in["tolerations"].([]interface{}); ok && len(v) > 0 {
		obj.Tolerations = expandTolerations(v)
	}
	return obj
}

func expandCNIParams(p []interface{}) *CustomCni {
	obj := &CustomCni{}
	log.Println("expand CNI params")

	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["custom_cni_cidr"].(string); ok && len(v) > 0 {
		obj.CustomCniCidr = v
	}
	//@@@what to do for expanding map[string][]object
	if v, ok := in["custom_cni_crd_spec"].([]interface{}); ok && len(v) > 0 {
		obj.CustomCniCrdSpec = expandCustomCNISpec(v)
	}
	return obj
}

func expandCustomCNISpec(p []interface{}) map[string][]CustomCniSpec {
	obj := make(map[string][]CustomCniSpec)
	log.Println("expand CNI Mapping")

	if len(p) == 0 || p[0] == nil {
		return obj
	}

	for i := range p {
		in := p[i].(map[string]interface{})
		elem2 := []CustomCniSpec{}

		if v, ok := in["cni_spec"].([]interface{}); ok && len(v) > 0 {
			elem2 = expandCNISpec(v)
		}

		if v, ok := in["name"].(string); ok && len(v) > 0 {
			obj[v] = elem2
		}
	}
	log.Println("Mapping Complete: ", obj)
	return obj
}

func expandCNISpec(p []interface{}) []CustomCniSpec {
	out := make([]CustomCniSpec, len(p))
	if len(p) == 0 || p[0] == nil {
		return out
	}
	for i := range p {
		obj := CustomCniSpec{}
		in := p[i].(map[string]interface{})

		if v, ok := in["subnet"].(string); ok && len(v) > 0 {
			obj.Subnet = v
		}
		if v, ok := in["security_groups"].([]interface{}); ok && len(v) > 0 {
			obj.SecurityGroups = toArrayStringSorted(v)
		}

		out[i] = obj
	}

	return out
}

func flattenEKSCluster(in *EKSCluster, p []interface{}) ([]interface{}, error) {
	obj := map[string]interface{}{}
	if in == nil {
		return nil, fmt.Errorf("empty cluster input")
	}

	if len(in.Kind) > 0 {
		obj["kind"] = in.Kind
	}
	var err error
	//flatten eks cluster metadata
	var ret1 []interface{}
	if in.Metadata != nil {
		v, ok := obj["metadata"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		ret1, err = flattenEKSClusterMetadata(in.Metadata, v)
		log.Println("ret1: ", ret1)
		if err != nil {
			log.Println("flattenEKSClusterMetadata err")
			return nil, err
		}
		obj["metadata"] = ret1
		log.Println("set metadata: ", obj["metadata"])
	}
	//flattening EKSClusterSpec
	var ret2 []interface{}
	if in.Spec != nil {
		v, ok := obj["spec"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		ret2, err = flattenEKSClusterSpec(in.Spec, v)
		if err != nil {
			log.Println("flattenEKSClusterSpec err")
			return nil, err
		}
		obj["spec"] = ret2
		log.Println("set metadata: ", obj["spec"])
	}
	log.Println("flattenEKSCluster finished ")
	return []interface{}{obj}, nil
}
func flattenEKSClusterMetadata(in *EKSClusterMetadata, p []interface{}) ([]interface{}, error) {
	if in == nil {
		log.Println("wrong input")
		return nil, fmt.Errorf("%s", "flattenEKSClusterMetaData empty input")
	}
	obj := map[string]interface{}{}

	if len(in.Name) > 0 {
		obj["name"] = in.Name
	}
	log.Println("md 1")
	if len(in.Project) > 0 {
		obj["project"] = in.Project
	}
	log.Println("md 2")
	if in.Labels != nil && len(in.Labels) > 0 {
		obj["labels"] = toMapInterface(in.Labels)
		log.Println("saving metadata labels: ", in.Labels)
	}
	log.Println("md 3")
	return []interface{}{obj}, nil
}
func flattenEKSClusterSpec(in *EKSSpec, p []interface{}) ([]interface{}, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "flattenEKSClusterMetaData empty input")
	}
	obj := map[string]interface{}{}

	if len(in.Type) > 0 {
		obj["type"] = in.Type
	}
	if len(in.Blueprint) > 0 {
		obj["blueprint"] = in.Blueprint
	}
	if len(in.BlueprintVersion) > 0 {
		obj["blueprint_version"] = in.BlueprintVersion
	}
	if len(in.CloudProvider) > 0 {
		obj["cloud_provider"] = in.CloudProvider
	}
	if len(in.CrossAccountRoleArn) > 0 {
		obj["cross_account_role_arn"] = in.CrossAccountRoleArn
	}
	if len(in.CniProvider) > 0 {
		obj["cni_provider"] = in.CniProvider
	}
	if in.CniParams != nil {
		v, ok := obj["cni_params"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["cni_params"] = flattenCNIParams(in.CniParams, v)
	}
	if in.ProxyConfig != nil {
		obj["proxy_config"] = flattenProxyConfig(in.ProxyConfig)
	}

	if in.SystemComponentsPlacement != nil {
		v, ok := obj["system_components_placement"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["system_components_placement"] = flattenSystemComponentsPlacement(in.SystemComponentsPlacement, v)
	}
	if in.Sharing != nil {
		obj["sharing"] = flattenEKSSharing(in.Sharing)
	}

	return []interface{}{obj}, nil
}

func flattenEKSSharing(in *EKSClusterSharing) []interface{} {
	if in == nil {
		return nil
	}
	obj := make(map[string]interface{})
	if in.Enabled != nil {
		obj["enabled"] = *in.Enabled
	}
	if len(in.Projects) > 0 {
		obj["projects"] = flattenEKSSharingProjects(in.Projects)
	}
	return []interface{}{obj}
}

func flattenEKSSharingProjects(in []*EKSClusterSharingProject) []interface{} {
	if len(in) == 0 {
		return nil
	}
	var out []interface{}
	for _, x := range in {
		obj := make(map[string]interface{})
		if len(x.Name) > 0 {
			obj["name"] = x.Name
		}
		out = append(out, obj)
	}
	return out
}

func flattenProxyConfig(in *ProxyConfig) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	log.Println("got to flatten proxy config:", in)

	if len(in.HttpProxy) > 0 {
		obj["http_proxy"] = in.HttpProxy
	}
	if len(in.HttpsProxy) > 0 {
		obj["https_proxy"] = in.HttpsProxy
	}
	if len(in.NoProxy) > 0 {
		obj["no_proxy"] = in.NoProxy
	}
	if len(in.ProxyAuth) > 0 {
		obj["proxy_auth"] = in.ProxyAuth
	}
	if len(in.BootstrapCA) > 0 {
		obj["bootstrap_ca"] = in.BootstrapCA
	}
	if in.Enabled != nil {
		obj["enabled"] = *in.Enabled
	}
	if in.AllowInsecureBootstrap != nil {
		obj["allow_insecure_bootstrap"] = *in.AllowInsecureBootstrap
	}

	return []interface{}{obj}

}

func flattenSystemComponentsPlacement(in *SystemComponentsPlacement, p []interface{}) []interface{} {
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}
	log.Println("got to flatten system comp:", in)
	log.Println("node_selectopr type: ", reflect.TypeOf(in.NodeSelector))
	if in.NodeSelector != nil && len(in.NodeSelector) > 0 {
		obj["node_selector"] = toMapInterface(in.NodeSelector)
	}
	if in.Tolerations != nil {
		v, ok := obj["tolerations"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		log.Println("type of read tolerations:", reflect.TypeOf(in.Tolerations), in.Tolerations)
		obj["tolerations"] = flattenTolerations(in.Tolerations, v)
	}
	if in.DaemonsetOverride != nil {
		v, ok := obj["daemonset_override"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["daemonset_override"] = flattenDaemonsetOverride(in.DaemonsetOverride, v)
	}

	return []interface{}{obj}
}

func flattenTolerations(in []*Tolerations, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	log.Println("flattenTolerations")
	out := make([]interface{}, len(in))
	for i, in := range in {
		obj := map[string]interface{}{}
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}

		if len(in.Key) > 0 {
			obj["key"] = in.Key
		}
		if len(in.Operator) > 0 {
			obj["operator"] = in.Operator
		}
		if len(in.Value) > 0 {
			obj["value"] = in.Value
		}
		if len(in.Effect) > 0 {
			obj["effect"] = in.Effect
		}
		if in.TolerationSeconds != nil {
			obj["toleration_seconds"] = in.TolerationSeconds
		}

		out[i] = &obj
	}
	return out
}

func flattenDaemonsetOverride(in *DaemonsetOverride, p []interface{}) []interface{} {
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	obj["node_selection_enabled"] = in.NodeSelectionEnabled

	if in.Tolerations != nil {
		v, ok := obj["tolerations"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["tolerations"] = flattenTolerations(in.Tolerations, v)
	}

	return []interface{}{obj}
}

func flattenCNIParams(in *CustomCni, p []interface{}) []interface{} {
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.CustomCniCidr) > 0 {
		obj["custom_cni_cidr"] = in.CustomCniCidr
	}
	if in.CustomCniCrdSpec != nil {
		v, ok := obj["custom_cni_crd_spec"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["custom_cni_crd_spec"] = flattenCustomCNISpec(in.CustomCniCrdSpec, v)
	}

	return []interface{}{obj}
}

func flattenCustomCNISpec(in map[string][]CustomCniSpec, p []interface{}) []interface{} {
	log.Println("got to flatten custom CNI mapping", len(p))
	out := make([]interface{}, len(in))
	i := 0
	for key, elem := range in {
		obj := map[string]interface{}{}
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}
		if elem != nil {
			v, ok := obj["cni_spec"].([]interface{})
			if !ok {
				v = []interface{}{}
			}
			obj["cni_spec"] = flattenCNISpec(elem, v)
		}
		if len(key) > 0 {
			obj["name"] = key
		}
		out[i] = obj
		i += 1
	}
	log.Println("finished customCNI mapping")
	return out
}

func flattenCNISpec(elem []CustomCniSpec, p []interface{}) []interface{} {
	if elem == nil {
		return nil
	}
	out := make([]interface{}, len(elem))
	for i, in := range elem {
		obj := map[string]interface{}{}
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}
		if len(in.Subnet) > 0 {
			obj["subnet"] = in.Subnet
		}
		if len(in.SecurityGroups) > 0 {
			obj["security_groups"] = toArrayInterfaceSorted(in.SecurityGroups)
		}
		out[i] = &obj
	}
	return out
}

func flattenEKSConfigMetadata(in *EKSClusterConfigMetadata, p []interface{}) ([]interface{}, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "flattenEKSClusterMetaData empty input")
	}
	obj := map[string]interface{}{}

	if len(in.Name) > 0 {
		obj["name"] = in.Name
	}
	if len(in.Region) > 0 {
		obj["region"] = in.Region
	}
	if len(in.Version) > 0 {
		obj["version"] = in.Version
	}

	if in.Tags != nil && len(in.Tags) > 0 {
		obj["tags"] = toMapInterface(in.Tags)
	}

	return []interface{}{obj}, nil
}
func flattenEKSClusterConfig(in *EKSClusterConfig, p []interface{}) ([]interface{}, error) {
	if in == nil {
		return nil, fmt.Errorf("empty cluster config input")
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.APIVersion) > 0 {
		obj["apiversion"] = in.APIVersion
	}
	if len(in.Kind) > 0 {
		obj["kind"] = in.Kind
	}
	var err error

	var ret1 []interface{}
	if in.Metadata != nil {
		v, ok := obj["metadata"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		ret1, err = flattenEKSConfigMetadata(in.Metadata, v)
		if err != nil {
			log.Println("flattenEKSClusterConfigMetadata err")
			return nil, err
		}
		obj["metadata"] = ret1
	}
	//setting up flatten KubernetesNetworkConfig
	var ret2 []interface{}
	if in.KubernetesNetworkConfig != nil {
		v, ok := obj["kubernetes_network_config"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		ret2, err = flattenEKSClusterKubernetesNetworkConfig(in.KubernetesNetworkConfig, v)
		if err != nil {
			log.Println("flattenEKSClusterKubernetesNetworkConfig err")
			return nil, err
		}
		obj["kubernetes_network_config"] = ret2
	}
	//setting up flatten IAM
	var ret3 []interface{}
	if in.IAM != nil {
		v, ok := obj["iam"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		ret3, err = flattenEKSClusterIAM(in.IAM, v)
		if err != nil {
			log.Println("flattenEKSClusterIAM err")
			return nil, err
		}
		obj["iam"] = ret3
	}
	//setting up flatten Identity Providers
	var ret4 []interface{}
	if in.IdentityProviders != nil {
		v, ok := obj["identity_providers"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		ret4, err = flattenEKSClusterIdentityProviders(in.IdentityProviders, v)
		if err != nil {
			log.Println("flattenEKSClusterIdentityProviders err")
			return nil, err
		}
		obj["identity_providers"] = ret4
	}
	//setting up flatten VPC
	var ret5 []interface{}
	if in.VPC != nil {
		v, ok := obj["vpc"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		ret5, err = flattenEKSClusterVPC(in.VPC, v)
		if err != nil {
			log.Println("flattenEKSClusterVPC err")
			return nil, err
		}
		obj["vpc"] = ret5
	}
	//setting up flatten Addon
	var ret6 []interface{}
	if in.Addons != nil {
		v, ok := obj["addons"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		ret6, err = flattenEKSClusterAddons(in.Addons, v)
		if err != nil {
			log.Println("flattenEKSClusterAddons err")
			return nil, err
		}
		obj["addons"] = ret6
	}
	//setting up flatten Private Clusters
	var ret7 []interface{}
	if in.PrivateCluster != nil {
		v, ok := obj["private_cluster"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		ret7 = flattenEKSClusterPrivateCluster(in.PrivateCluster, v)
		/*
			if err != nil {
				log.Println("flattenEKSClusterPrivateCluster err")
				return nil, err
			}*/
		obj["private_cluster"] = ret7
	}
	//setting up flatten Node Groups
	var ret8 []interface{}
	if in.NodeGroups != nil {
		v, ok := obj["node_groups"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		ret8 = flattenEKSClusterNodeGroups(in.NodeGroups, v)
		/*
			if err != nil {
				log.Println("flattenEKSClusterNodeGroups err")
				return nil, err
			}*/
		log.Println("flattend node group")
		obj["node_groups"] = ret8
	}
	//setting up flatten Managed Node Groups
	var ret9 []interface{}
	if in.ManagedNodeGroups != nil {
		v, ok := obj["managed_nodegroups"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		ret9, err = flattenEKSClusterManagedNodeGroups(in.ManagedNodeGroups, v)
		if err != nil {
			log.Println("flattenEKSClusterManagedNodeGroups err")
			return nil, err
		}
		obj["managed_nodegroups"] = ret9
		log.Println("flattend managed node group: ", obj["managed_nodegroups"], ret9)
	}
	//setting up flatten Fargate Profiles
	var ret10 []interface{}
	if in.FargateProfiles != nil {
		v, ok := obj["fargate_profiles"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		ret10 = flattenEKSClusterFargateProfiles(in.FargateProfiles, v)
		/*if err != nil {
			log.Println("flattenEKSClusterFargateProfiles err")
			return nil, err
		}*/
		obj["fargate_profiles"] = ret10
	}
	//setting up flatten Availability Zones
	if in.AvailabilityZones != nil && len(in.AvailabilityZones) > 0 {
		obj["availability_zones"] = toArrayInterfaceSorted(in.AvailabilityZones)
	}
	//setting up flatten Cloud Watch
	var ret11 []interface{}
	if in.CloudWatch != nil {
		v, ok := obj["cloud_watch"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		ret11 = flattenEKSClusterCloudWatch(in.CloudWatch, v)
		/*if err != nil {
			log.Println("flattenEKSClusterCloudWatch err")
			return nil, err
		}*/
		obj["cloud_watch"] = ret11
	}
	//setting up flatten Secrets Encryption
	var ret12 []interface{}
	if in.SecretsEncryption != nil {
		v, ok := obj["secrets_encryption"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		ret12 = flattenEKSClusterSecretsEncryption(in.SecretsEncryption, v)
		/*if err != nil {
			log.Println("flattenEKSClusterSecretsEncryption err")
			return nil, err
		}*/
		obj["secrets_encryption"] = ret12
	}
	// setting up flatten identity mappings
	//var ret13 []interface{}
	if in.IdentityMappings != nil {
		v, ok := obj["identity_mappings"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		ret13, err := flattenIdentityMappings(in.IdentityMappings, v)
		if err != nil {
			log.Println("flattenIdentityMapping err")
			return nil, err
		}
		obj["identity_mappings"] = ret13
	}

	log.Println("end of flatten config")

	return []interface{}{obj}, nil
}

func flattenEKSClusterKubernetesNetworkConfig(in *KubernetesNetworkConfig, p []interface{}) ([]interface{}, error) {
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}
	if in == nil {
		return []interface{}{obj}, nil
	}

	if len(in.IPFamily) > 0 {
		obj["ip_family"] = in.IPFamily
	}
	if len(in.ServiceIPv4CIDR) > 0 {
		obj["service_ipv4_cidr"] = in.ServiceIPv4CIDR
	}
	return []interface{}{obj}, nil
}
func flattenEKSClusterIAM(in *EKSClusterIAM, p []interface{}) ([]interface{}, error) {
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}
	if in == nil {
		return []interface{}{obj}, nil
	}

	if len(in.ServiceRoleARN) > 0 {
		obj["service_role_arn"] = in.ServiceRoleARN
	}
	if len(in.ServiceRolePermissionsBoundary) > 0 {
		obj["service_role_permission_boundary"] = in.ServiceRolePermissionsBoundary
	}
	if len(in.FargatePodExecutionRoleARN) > 0 {
		obj["fargate_pod_execution_role_arn"] = in.FargatePodExecutionRoleARN
	}
	if len(in.FargatePodExecutionRolePermissionsBoundary) > 0 {
		obj["fargate_pod_execution_role_permissions_boundary"] = in.FargatePodExecutionRolePermissionsBoundary
	}

	obj["with_oidc"] = in.WithOIDC

	if in.ServiceAccounts != nil {
		v, ok := obj["service_accounts"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["service_accounts"] = flattenIAMServiceAccounts(in.ServiceAccounts, v)
	}

	obj["vpc_resource_controller_policy"] = in.VPCResourceControllerPolicy

	return []interface{}{obj}, nil
}

func flattenIAMServiceAccountMetadata(in *EKSClusterIAMMeta, p []interface{}) []interface{} {
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.Name) > 0 {
		obj["name"] = in.Name
	}
	if len(in.Namespace) > 0 {
		obj["namespace"] = in.Namespace
	}

	if in.Labels != nil && len(in.Labels) > 0 {
		obj["labels"] = toMapInterface(in.Labels)
	}
	if in.Annotations != nil && len(in.Annotations) > 0 {
		obj["annotations"] = toMapInterface(in.Annotations)
	}

	return []interface{}{obj}
}

func flattenIAMServiceAccounts(inp []*EKSClusterIAMServiceAccount, p []interface{}) []interface{} {
	if inp == nil {
		return nil
	}
	out := make([]interface{}, len(inp))
	for i, in := range inp {

		obj := map[string]interface{}{}
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}

		v, ok := obj["metadata"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["metadata"] = flattenIAMServiceAccountMetadata(in.Metadata, v)

		if in.AttachPolicyARNs != nil && len(in.AttachPolicyARNs) > 0 {
			obj["attach_policy_arns"] = toArrayInterface(in.AttachPolicyARNs)
		}

		v, ok = obj["well_known_policies"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["well_known_policies"] = flattenIAMWellKnownPolicies(in.WellKnownPolicies, v)

		//@@@TODO Store inline document object as terraform input correctly
		/*v1, ok := obj["attach_policy"].([]interface{})
		if !ok {
			v1 = []interface{}{}
		}
		obj["attach_policy"] = flattenAttachPolicy(in.AttachPolicy, v1)
		*/
		log.Println("input attach policy:", in.AttachPolicy)
		if in.AttachPolicy != nil && len(in.AttachPolicy) > 0 {
			//log.Println("type:", reflect.TypeOf(in.AttachPolicy))
			var json2 = jsoniter.ConfigCompatibleWithStandardLibrary
			jsonStr, err := json2.Marshal(in.AttachPolicy)
			if err != nil {
				log.Println("attach policy marshal err:", err)
			}
			//log.Println("jsonSTR:", jsonStr)
			obj["attach_policy"] = string(jsonStr)
			//log.Println("attach policy flattened correct:", obj["attach_policy"])
		}
		if len(in.AttachRoleARN) > 0 {
			obj["attach_role_arn"] = in.AttachRoleARN
		}
		if len(in.PermissionsBoundary) > 0 {
			obj["permissions_boundary"] = in.PermissionsBoundary
		}
		if in.Status != nil {
			v, ok := obj["status"].([]interface{})
			if !ok {
				v = []interface{}{}
			}
			obj["status"] = flattenIAMStatus(in.Status, v)
		}
		if len(in.RoleName) > 0 {
			obj["role_name"] = in.RoleName
		}

		obj["role_only"] = in.RoleOnly

		if in.Tags != nil && len(in.Tags) > 0 {
			obj["tags"] = toMapInterface(in.Tags)
		}

		out[i] = &obj
	}

	return out

}

// @@@Flatten attach policy
func flattenAttachPolicy(in *InlineDocument, p []interface{}) []interface{} {
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}
	/*
		if in == nil {
			return []interface{}{obj}
		}*/
	if len(in.Version) > 0 {
		obj["version"] = in.Version
	}

	v, ok := obj["statement"].([]interface{})
	if !ok {
		v = []interface{}{}
	}
	obj["statement"] = flattenStatement(in.Statement, v)

	return []interface{}{obj}
}

func flattenStatement(in InlineStatement, p []interface{}) []interface{} {
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.Effect) > 0 {
		obj["effect"] = in.Effect
	}
	if len(in.Action) > 0 {
		obj["action"] = toArrayInterface(in.Action)
	}
	if len(in.Resource) > 0 {
		obj["resource"] = in.Resource
	}

	return []interface{}{obj}
}
func flattenIAMStatus(in *ClusterIAMServiceAccountStatus, p []interface{}) []interface{} {
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}
	if in == nil {
		return []interface{}{obj}
	}
	if len(in.RoleARN) > 0 {
		obj["role_arn"] = in.RoleARN
	}

	return []interface{}{obj}
}
func flattenIAMWellKnownPolicies(in *WellKnownPolicies, p []interface{}) []interface{} {
	if in == nil {
		return make([]interface{}, 0)
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	obj["image_builder"] = in.ImageBuilder
	obj["auto_scaler"] = in.AutoScaler
	obj["aws_load_balancer_controller"] = in.AWSLoadBalancerController
	obj["external_dns"] = in.ExternalDNS
	obj["cert_manager"] = in.CertManager
	obj["ebs_csi_controller"] = in.EBSCSIController
	obj["efs_csi_controller"] = in.EFSCSIController
	return []interface{}{obj}
}
func flattenEKSClusterIdentityProviders(inp []*IdentityProvider, p []interface{}) ([]interface{}, error) {
	out := make([]interface{}, len(inp))
	if inp == nil {
		return []interface{}{out}, nil
	}
	for i, in := range inp {
		obj := map[string]interface{}{}
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}
		if len(in.Type) > 0 {
			obj["type"] = in.Type
		}

		out[i] = &obj
	}
	return out, nil
}
func flattenEKSClusterVPC(in *EKSClusterVPC, p []interface{}) ([]interface{}, error) {
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}
	if in == nil {
		return []interface{}{obj}, nil
	}
	if len(in.ID) > 0 {
		obj["id"] = in.ID
	}
	if len(in.CIDR) > 0 {
		obj["cidr"] = in.CIDR
	}
	if len(in.IPv6Cidr) > 0 {
		obj["ipv6_cidr"] = in.CIDR
	}
	if len(in.IPv6Pool) > 0 {
		obj["ipv6_pool"] = in.CIDR
	}
	if len(in.SecurityGroup) > 0 {
		obj["security_group"] = in.SecurityGroup
	}

	if in.Subnets != nil {
		v, ok := obj["subnets"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["subnets"] = flattenVPCSubnets(in.Subnets, v)
	}

	if len(in.ExtraCIDRs) > 0 {
		obj["extra_cidrs"] = toArrayInterface(in.ExtraCIDRs)
	}
	if in.ExtraIPv6CIDRs != nil && len(in.ExtraIPv6CIDRs) > 0 {
		obj["extra_ipv6_cidrs"] = toArrayInterface(in.ExtraIPv6CIDRs)
	}
	if len(in.SharedNodeSecurityGroup) > 0 {
		obj["shared_node_security_group"] = in.SharedNodeSecurityGroup
	}

	obj["manage_shared_node_security_group_rules"] = in.ManageSharedNodeSecurityGroupRules
	obj["auto_allocate_ipv6"] = in.AutoAllocateIPv6

	if in.NAT != nil {
		v, ok := obj["nat"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["nat"] = flattenVPCNAT(in.NAT, v)
	}
	if in.ClusterEndpoints != nil {
		v, ok := obj["cluster_endpoints"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["cluster_endpoints"] = flattenVPCClusterEndpoints(in.ClusterEndpoints, v)
	}

	if len(in.PublicAccessCIDRs) > 0 {
		obj["public_access_cidrs"] = toArrayInterface(in.PublicAccessCIDRs)
	}

	return []interface{}{obj}, nil
}
func flattenVPCSubnets(in *ClusterSubnets, p []interface{}) []interface{} {
	log.Println("got to flatten subnet", in)
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}
	if in.Private != nil {
		v, ok := obj["private"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["private"] = flattenSubnetMapping(in.Private, v)
	}
	if in.Public != nil {
		v, ok := obj["public"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["public"] = flattenSubnetMapping(in.Public, v)
	}
	n1 := spew.Sprintf("%+v", obj)
	log.Println("flattenVPCSubnets:", n1)
	return []interface{}{obj}
}
func flattenSubnetMapping(in AZSubnetMapping, p []interface{}) []interface{} {
	log.Println("got to flatten subnet mapping", len(p))
	out := make([]interface{}, len(in))
	i := 0
	orderedSubnetNames := getSubnetNamesOrderFromState(p)

	for idx := 0; idx < len(orderedSubnetNames); idx++ {
		obj := map[string]interface{}{}
		if idx < len(p) && p[idx] != nil {
			obj = p[idx].(map[string]interface{})
		}
		name := orderedSubnetNames[idx]
		if elem, ok := in[name]; ok {
			if len(elem.ID) > 0 {
				obj["id"] = elem.ID
			}
			if len(elem.AZ) > 0 {
				obj["az"] = elem.AZ
			}
			if len(name) > 0 {
				obj["name"] = name
			}
			if len(elem.CIDR) > 0 {
				obj["cidr"] = elem.CIDR
			}
			out[i] = obj
			i += 1
		}
	}
	for key, elem := range in {
		if !slices.Contains(orderedSubnetNames, key) {
			obj := map[string]interface{}{}
			if len(elem.ID) > 0 {
				obj["id"] = elem.ID
			}
			if len(elem.AZ) > 0 {
				obj["az"] = elem.AZ
			}
			if len(key) > 0 {
				obj["name"] = key
			}
			if len(elem.CIDR) > 0 {
				obj["cidr"] = elem.CIDR
			}
			out[i] = obj
			i += 1
		}
	}
	log.Println("finished subnet mapping")
	return out
}

func getSubnetNamesOrderFromState(p []interface{}) []string {
	extractValue := func(obj map[string]interface{}, key string) string {
		if val, ok := obj[key]; ok {
			if val2, ok2 := val.(string); ok2 {
				return val2
			}
		}
		return ""
	}
	res := make([]string, len(p))
	for i := 0; i < len(p); i++ {
		if p[i] != nil {
			if obj, ok := p[i].(map[string]interface{}); ok {
				if x := extractValue(obj, "name"); x != "" {
					res = append(res, obj["name"].(string))
				}
			}
		}
	}
	return res
}
func flattenVPCNAT(in *ClusterNAT, p []interface{}) []interface{} {
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}
	if in == nil {
		return []interface{}{obj}
	}
	if len(in.Gateway) > 0 {
		obj["gateway"] = in.Gateway
	}
	return []interface{}{obj}
}
func flattenVPCClusterEndpoints(in *ClusterEndpoints, p []interface{}) []interface{} {
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}
	if in == nil {
		return []interface{}{obj}
	}
	obj["private_access"] = in.PrivateAccess
	obj["public_access"] = in.PublicAccess
	return []interface{}{obj}
}

func flattenEKSClusterAddons(inp []*Addon, p []interface{}) ([]interface{}, error) {
	if inp == nil {
		return nil, fmt.Errorf("emptyinput flatten addons")
	}
	out := make([]interface{}, len(inp))
	for i, in := range inp {
		obj := map[string]interface{}{}
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}
		if len(in.Name) > 0 {
			obj["name"] = in.Name
		}
		if len(in.Version) > 0 {
			obj["version"] = in.Version
		}
		if len(in.ServiceAccountRoleARN) > 0 {
			obj["service_account_role_arn"] = in.ServiceAccountRoleARN
		}
		if len(in.AttachPolicyARNs) > 0 {
			obj["attach_policy_arns"] = toArrayInterface(in.AttachPolicyARNs)
		} else {
			obj["attach_policy_arns"] = make([]interface{}, 0)
		}
		//@@@TODO Store inline document object as terraform input correctly
		if in.AttachPolicy != nil {
			v1, ok := obj["attach_policy"].([]interface{})
			if !ok {
				v1 = []interface{}{}
			}
			obj["attach_policy"] = flattenAttachPolicy(in.AttachPolicy, v1)
			if len(in.PermissionsBoundary) > 0 {
				obj["permissions_boundary"] = in.PermissionsBoundary
			}
		}
		v, ok := obj["well_known_policies"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["well_known_policies"] = flattenIAMWellKnownPolicies(in.WellKnownPolicies, v)

		obj["tags"] = toMapInterface(in.Tags)
		//Force field for existing addon (not in doc)
		if len(in.ConfigurationValues) > 0 {
			obj["configuration_values"] = in.ConfigurationValues
		}

		out[i] = &obj
	}
	return out, nil
}
func flattenEKSClusterPrivateCluster(in *PrivateCluster, p []interface{}) []interface{} {
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}
	if in == nil {
		return []interface{}{obj}
	}

	obj["enabled"] = in.Enabled
	obj["skip_endpoint_creation"] = in.SkipEndpointCreation

	if len(in.AdditionalEndpointServices) > 0 {
		obj["additional_endpoint_services"] = toArrayInterface(in.AdditionalEndpointServices)
	}
	return []interface{}{obj}
}
func flattenEKSClusterNodeGroups(inp []*NodeGroup, p []interface{}) []interface{} {
	if inp == nil {
		return nil
	}

	inpSorted := make([]NodeGroup, len(inp))
	for i := range inp {
		inpSorted[i] = *inp[i]
	}
	sort.Sort(ByNodeGroupName(inpSorted))

	out := make([]interface{}, len(inp))
	for i, in := range inpSorted {
		obj := map[string]interface{}{}
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}
		if len(in.Name) > 0 {
			obj["name"] = in.Name
		}
		if len(in.AMIFamily) > 0 {
			obj["ami_family"] = in.AMIFamily
		}
		if len(in.InstanceType) > 0 {
			obj["instance_type"] = in.InstanceType
		}
		if len(in.AvailabilityZones) > 0 {
			obj["availability_zones"] = toArrayInterfaceSorted(in.AvailabilityZones)
		}
		if len(in.Subnets) > 0 {
			obj["subnets"] = toArrayInterface(in.Subnets)
		}
		if len(in.InstancePrefix) > 0 {
			obj["instance_prefix"] = in.InstancePrefix
		}
		if len(in.InstanceName) > 0 {
			obj["instance_name"] = in.InstanceName
		}
		obj["desired_capacity"] = in.DesiredCapacity
		obj["min_size"] = in.MinSize
		obj["max_size"] = in.MaxSize
		obj["volume_size"] = in.VolumeSize

		if in.SSH != nil {
			v, ok := obj["ssh"].([]interface{})
			if !ok {
				v = []interface{}{}
			}
			obj["ssh"] = flattenNodeGroupSSH(in.SSH, v)
		}

		if len(in.Labels) > 0 {
			obj["labels"] = toMapInterface(in.Labels)
		}
		obj["private_networking"] = in.PrivateNetworking
		if len(in.Tags) > 0 {
			obj["tags"] = toMapInterface(in.Tags)
		}
		if in.IAM != nil {
			v, ok := obj["iam"].([]interface{})
			if !ok {
				v = []interface{}{}
			}
			obj["iam"] = flattenNodeGroupIAM(in.IAM, v)
		}
		if len(in.AMI) > 0 {
			obj["ami"] = in.AMI
		}
		if in.SecurityGroups != nil {
			v, ok := obj["security_groups"].([]interface{})
			if !ok {
				v = []interface{}{}
			}
			obj["security_groups"] = flattenNodeGroupSecurityGroups(in.SecurityGroups, v)
		}
		obj["max_pods_per_node"] = in.MaxPodsPerNode
		if len(in.ASGSuspendProcesses) > 0 {
			obj["asg_suspend_processes"] = toArrayInterface(in.ASGSuspendProcesses)
		}
		obj["ebs_optimized"] = in.EBSOptimized
		if len(in.VolumeType) > 0 {
			obj["volume_type"] = in.VolumeType
		}
		if len(in.VolumeName) > 0 {
			obj["volume_name"] = in.VolumeName
		}
		obj["volume_encrypted"] = in.VolumeEncrypted
		if len(in.VolumeKmsKeyID) > 0 {
			obj["volume_kms_key_id"] = in.VolumeKmsKeyID
		}
		obj["volume_iops"] = in.VolumeIOPS
		obj["volume_throughput"] = in.VolumeThroughput
		if len(in.PreBootstrapCommands) > 0 {
			obj["pre_bootstrap_commands"] = toArrayInterface(in.PreBootstrapCommands)
		}
		if len(in.OverrideBootstrapCommand) > 0 {
			obj["override_bootstrap_command"] = in.OverrideBootstrapCommand
		}
		obj["disable_imdsv1"] = in.DisableIMDSv1
		obj["disable_pods_imds"] = in.DisablePodIMDS
		if in.Placement != nil {
			v, ok := obj["placement"].([]interface{})
			if !ok {
				v = []interface{}{}
			}
			obj["placement"] = flattenNodeGroupPlacement(in.Placement, v)
		}
		obj["efa_enabled"] = in.EFAEnabled
		if in.InstanceSelector != nil {
			v, ok := obj["instance_selector"].([]interface{})
			if !ok {
				v = []interface{}{}
			}
			obj["instance_selector"] = flattenNodeGroupInstanceSelector(in.InstanceSelector, v)
		}
		//dont have additional encryption volume in doc
		if in.Bottlerocket != nil {
			v, ok := obj["bottle_rocket"].([]interface{})
			if !ok {
				v = []interface{}{}
			}
			obj["bottle_rocket"] = flattenNodeGroupBottlerocket(in.Bottlerocket, v)
		}
		if in.InstancesDistribution != nil {
			v, ok := obj["instances_distribution"].([]interface{})
			if !ok {
				v = []interface{}{}
			}
			obj["instances_distribution"] = flattenNodeGroupInstancesDistribution(in.InstancesDistribution, v)
		}
		if in.ASGMetricsCollection != nil {
			v, ok := obj["asg_metrics_collection"].([]interface{})
			if !ok {
				v = []interface{}{}
			}
			obj["asg_metrics_collection"] = flattenNodeGroupASGMetricsCollection(in.ASGMetricsCollection, v)
		}
		if len(in.CPUCredits) > 0 {
			obj["cpu_credits"] = in.CPUCredits
		}
		if len(in.ClassicLoadBalancerNames) > 0 {
			obj["classic_load_balancer_names"] = toArrayInterface(in.ClassicLoadBalancerNames)
		}
		if len(in.TargetGroupARNs) > 0 {
			obj["target_group_arns"] = toArrayInterface(in.TargetGroupARNs)
		}
		if in.Taints != nil {
			v, ok := obj["taints"].([]interface{})
			if !ok {
				v = []interface{}{}
			}
			obj["taints"] = flattenNodeGroupTaint(in.Taints, v)
		}
		if in.UpdateConfig != nil {
			v, ok := obj["update_config"].([]interface{})
			if !ok {
				v = []interface{}{}
			}
			obj["update_config"] = flattenNodeGroupUpdateConfig(in.UpdateConfig, v)
		}
		if len(in.ClusterDNS) > 0 {
			obj["cluster_dns"] = in.ClusterDNS
		}

		if len(in.Version) > 0 {
			obj["version"] = in.Version
		}
		//@@@TODO Store inline document object as terraform input correctly
		if in.KubeletExtraConfig != nil {
			v, ok := obj["kubelet_extra_config"].([]interface{})
			if !ok {
				v = []interface{}{}
			}
			obj["kubelet_extra_config"] = flattenKubeletExtraConfig(in.KubeletExtraConfig, v)
		}
		//Container Runtime not in doc from struct
		out[i] = &obj
	}
	return out
}

func flattenKubeletExtraConfig(in *KubeletExtraConfig, p []interface{}) []interface{} {
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}
	if in == nil {
		return []interface{}{obj}
	}

	if len(in.KubeReserved) > 0 {
		obj["kube_reserved"] = toMapInterface(in.KubeReserved)
	}
	if len(in.KubeReservedCGroup) > 0 {
		obj["kube_reserved_cgroup"] = in.KubeReservedCGroup
	}
	if len(in.SystemReserved) > 0 {
		obj["system_reserved"] = toMapInterface(in.SystemReserved)
	}
	if len(in.EvictionHard) > 0 {
		obj["eviction_hard"] = toMapInterface(in.EvictionHard)
	}
	if len(in.FeatureGates) > 0 {
		obj["feature_gates"] = toMapBoolInterface(in.FeatureGates)
	}
	return []interface{}{obj}
}

func flattenNodeGroupSSH(in *NodeGroupSSH, p []interface{}) []interface{} {
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}
	if in == nil {
		return []interface{}{obj}
	}

	obj["allow"] = in.Allow
	if len(in.PublicKeyPath) > 0 {
		obj["public_key"] = in.PublicKeyPath
	}
	if len(in.PublicKeyName) > 0 {
		obj["public_key_name"] = in.PublicKeyName
	}
	if len(in.SourceSecurityGroupIDs) > 0 {
		obj["source_security_group_ids"] = toArrayInterface(in.SourceSecurityGroupIDs)
	}
	obj["enable_ssm"] = in.EnableSSM
	return []interface{}{obj}
}
func flattenNodeGroupIAM(in *NodeGroupIAM, p []interface{}) []interface{} {
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}
	if in == nil {
		return []interface{}{obj}
	}
	//@@@TODO Store inline document object as terraform input correctly
	if in.AttachPolicy != nil {
		v1, ok := obj["attach_policy"].([]interface{})
		if !ok {
			v1 = []interface{}{}
		}
		obj["attach_policy"] = flattenAttachPolicy(in.AttachPolicy, v1)
	}

	if len(in.AttachPolicyARNs) > 0 {
		obj["attach_policy_arns"] = toArrayInterface(in.AttachPolicyARNs)
	}
	if len(in.InstanceProfileARN) > 0 {
		obj["instance_profile_arn"] = in.InstanceProfileARN
	}
	if len(in.InstanceRoleARN) > 0 {
		obj["instance_role_arn"] = in.InstanceRoleARN
	}
	if len(in.InstanceRoleName) > 0 {
		obj["instance_role_name"] = in.InstanceRoleName
	}
	if len(in.InstanceRolePermissionsBoundary) > 0 {
		obj["instance_role_permission_boundary"] = in.InstanceRolePermissionsBoundary
	}

	v, ok := obj["iam_node_group_with_addon_policies"].([]interface{})
	if !ok {
		v = []interface{}{}

	}

	if in.WithAddonPolicies != nil {
		obj["iam_node_group_with_addon_policies"] = flattenNodeGroupIAMWithAddonPolicies(in.WithAddonPolicies, v)
	}

	return []interface{}{obj}
}

func flattenNodeGroupIAMWithAddonPolicies(in *NodeGroupIAMAddonPolicies, p []interface{}) []interface{} {
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	obj["image_builder"] = in.ImageBuilder
	obj["auto_scaler"] = in.AutoScaler
	obj["external_dns"] = in.ExternalDNS
	obj["cert_manager"] = in.CertManager
	obj["app_mesh"] = in.AppMesh
	obj["app_mesh_review"] = in.AppMeshPreview
	obj["ebs"] = in.EBS
	obj["fsx"] = in.FSX
	obj["efs"] = in.EFS
	obj["alb_ingress"] = in.AWSLoadBalancerController
	obj["xray"] = in.XRay
	obj["cloud_watch"] = in.CloudWatch

	return []interface{}{obj}
}
func flattenNodeGroupSecurityGroups(in *NodeGroupSGs, p []interface{}) []interface{} {
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}
	if in == nil {
		return []interface{}{obj}
	}
	if len(in.AttachIDs) > 0 {
		obj["attach_ids"] = toArrayInterface(in.AttachIDs)
	}
	obj["with_shared"] = in.WithShared
	obj["with_local"] = in.WithLocal
	return []interface{}{obj}
}
func flattenNodeGroupPlacement(in *Placement, p []interface{}) []interface{} {
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}
	if in == nil {
		return []interface{}{obj}
	}
	if len(in.GroupName) > 0 {
		obj["group"] = in.GroupName
	}
	return []interface{}{obj}
}
func flattenNodeGroupInstanceSelector(in *InstanceSelector, p []interface{}) []interface{} {
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}
	if in == nil {
		return []interface{}{obj}
	}
	obj["vcpus"] = in.VCPUs
	if len(in.Memory) > 0 {
		obj["memory"] = in.Memory
	}
	obj["gpus"] = in.GPUs
	if len(in.CPUArchitecture) > 0 {
		obj["cpu_architecture"] = in.CPUArchitecture
	}
	return []interface{}{obj}
}
func flattenNodeGroupBottlerocket(in *NodeGroupBottlerocket, p []interface{}) []interface{} {
	obj := map[string]interface{}{}
	if in == nil {
		return []interface{}{obj}
	}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}
	obj["enable_admin_container"] = in.EnableAdminContainer

	if in.Settings != nil && len(in.Settings) > 0 {
		var json2 = jsoniter.ConfigCompatibleWithStandardLibrary
		jsonStr, err := json2.Marshal(in.Settings)
		if err != nil {
			log.Println("attach policy marshal err:", err)
		}
		log.Println("jsonSTR:", jsonStr)
		obj["settings"] = string(jsonStr)
		log.Println("bottlerocket settings flattened correct:", obj["settings"])
	}
	return []interface{}{obj}
}
func flattenNodeGroupInstancesDistribution(in *NodeGroupInstancesDistribution, p []interface{}) []interface{} {
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}
	if in == nil {
		return []interface{}{obj}
	}
	if len(in.InstanceTypes) > 0 {
		obj["instance_types"] = toArrayInterface(in.InstanceTypes)
	}
	obj["max_price"] = in.MaxPrice
	obj["on_demand_base_capacity"] = in.OnDemandBaseCapacity
	obj["on_demand_percentage_above_base_capacity"] = in.OnDemandPercentageAboveBaseCapacity
	obj["spot_instance_pools"] = in.SpotInstancePools
	if len(in.SpotAllocationStrategy) > 0 {
		obj["spot_allocation_strategy"] = in.SpotAllocationStrategy
	}
	obj["capacity_rebalance"] = in.CapacityRebalance

	return []interface{}{obj}
}
func flattenNodeGroupASGMetricsCollection(inp []MetricsCollection, p []interface{}) []interface{} {
	if inp == nil {
		return nil
	}
	out := make([]interface{}, len(inp))
	for i, in := range inp {
		obj := map[string]interface{}{}
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}
		if len(in.Granularity) > 0 {
			obj["granularity"] = in.Granularity
		}
		if len(in.Metrics) > 0 {
			obj["metrics"] = toArrayInterface(in.Metrics)
		}
		out[i] = obj
	}
	return out
}
func flattenNodeGroupUpdateConfig(in *NodeGroupUpdateConfig, p []interface{}) []interface{} {
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}
	if in == nil {
		return []interface{}{obj}
	}

	obj["max_unavaliable"] = in.MaxUnavailable
	obj["max_unavaliable_percetage"] = in.MaxUnavailablePercentage
	return []interface{}{obj}
}

// Flatten mnanaged Node Groups
func flattenEKSClusterManagedNodeGroups(inp []*ManagedNodeGroup, p []interface{}) ([]interface{}, error) {
	if inp == nil {
		return nil, fmt.Errorf("empty input for managedNodeGroup")
	}

	inpSorted := make([]ManagedNodeGroup, len(inp))
	for i := range inp {
		inpSorted[i] = *inp[i]
	}
	sort.Sort(ByManagedNodeGroupName(inpSorted))

	out := make([]interface{}, len(inp))
	for i, in := range inpSorted {
		obj := map[string]interface{}{}
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}
		if len(in.Name) > 0 {
			obj["name"] = in.Name
		}
		if len(in.AMIFamily) > 0 {
			obj["ami_family"] = in.AMIFamily
		}
		if len(in.InstanceType) > 0 {
			obj["instance_type"] = in.InstanceType
		}
		if len(in.AvailabilityZones) > 0 {
			obj["availability_zones"] = toArrayInterfaceSorted(in.AvailabilityZones)
		}
		if len(in.Subnets) > 0 {
			obj["subnets"] = toArrayInterfaceSorted(in.Subnets)
		}
		if len(in.InstancePrefix) > 0 {
			obj["instance_prefix"] = in.InstancePrefix
		}
		if len(in.InstanceName) > 0 {
			obj["instance_name"] = in.InstanceName
		}
		obj["desired_capacity"] = in.DesiredCapacity
		obj["min_size"] = in.MinSize
		obj["max_size"] = in.MaxSize
		obj["volume_size"] = in.VolumeSize

		if in.SSH != nil {
			v, ok := obj["ssh"].([]interface{})
			if !ok {
				v = []interface{}{}
			}
			obj["ssh"] = flattenNodeGroupSSH(in.SSH, v)
		}

		if len(in.Labels) > 0 {
			obj["labels"] = toMapInterface(in.Labels)
		}
		obj["private_networking"] = in.PrivateNetworking
		if len(in.Tags) > 0 {
			obj["tags"] = toMapInterface(in.Tags)
		}
		if in.IAM != nil {
			v, ok := obj["iam"].([]interface{})
			if !ok {
				v = []interface{}{}
			}
			obj["iam"] = flattenNodeGroupIAM(in.IAM, v)
		}
		if len(in.AMI) > 0 {
			obj["ami"] = in.AMI
		}
		if in.SecurityGroups != nil {
			v, ok := obj["security_groups"].([]interface{})
			if !ok {
				v = []interface{}{}
			}
			obj["security_groups"] = flattenNodeGroupSecurityGroups(in.SecurityGroups, v)
		}
		obj["max_pods_per_node"] = in.MaxPodsPerNode
		if len(in.ASGSuspendProcesses) > 0 {
			obj["asg_suspend_processes"] = toArrayInterface(in.ASGSuspendProcesses)
		}
		obj["ebs_optimized"] = in.EBSOptimized
		if len(in.VolumeType) > 0 {
			obj["volume_type"] = in.VolumeType
		}
		if len(in.VolumeName) > 0 {
			obj["volume_name"] = in.VolumeName
		}
		obj["volume_encrypted"] = in.VolumeEncrypted
		if len(in.VolumeKmsKeyID) > 0 {
			obj["volume_kms_key_id"] = in.VolumeKmsKeyID
		}
		obj["volume_iops"] = in.VolumeIOPS
		obj["volume_throughput"] = in.VolumeThroughput
		if len(in.PreBootstrapCommands) > 0 {
			obj["pre_bootstrap_commands"] = toArrayInterface(in.PreBootstrapCommands)
		}
		if len(in.OverrideBootstrapCommand) > 0 {
			obj["override_bootstrap_command"] = in.OverrideBootstrapCommand
		}
		obj["disable_imdsv1"] = in.DisableIMDSv1
		obj["disable_pods_imds"] = in.DisablePodIMDS
		if in.Placement != nil {
			v, ok := obj["placement"].([]interface{})
			if !ok {
				v = []interface{}{}
			}
			obj["placement"] = flattenNodeGroupPlacement(in.Placement, v)
		}
		//@@@ efa enabled
		obj["efa_enabled"] = in.EFAEnabled
		//log.Println("input efaEnabled:", *in.EFAEnabled)
		//log.Println("object efaEnabled:", obj["efa_enabled"])
		if in.InstanceSelector != nil {
			v, ok := obj["instance_selector"].([]interface{})
			if !ok {
				v = []interface{}{}
			}
			obj["instance_selector"] = flattenNodeGroupInstanceSelector(in.InstanceSelector, v)
		}
		//dont have additional encryption volume in doc
		if in.Bottlerocket != nil {
			v, ok := obj["bottle_rocket"].([]interface{})
			if !ok {
				v = []interface{}{}
			}
			obj["bottle_rocket"] = flattenNodeGroupBottlerocket(in.Bottlerocket, v)
		}
		if len(in.InstanceTypes) > 0 {
			obj["instance_types"] = toArrayInterface(in.InstanceTypes)
		}
		obj["spot"] = in.Spot
		if in.Taints != nil {
			v, ok := obj["taints"].([]interface{})
			if !ok {
				v = []interface{}{}
			}
			obj["taints"] = flattenNodeGroupTaint(in.Taints, v)
		}
		if in.UpdateConfig != nil {
			v, ok := obj["update_config"].([]interface{})
			if !ok {
				v = []interface{}{}
			}
			obj["update_config"] = flattenNodeGroupUpdateConfig(in.UpdateConfig, v)
		}
		if in.LaunchTemplate != nil {
			v, ok := obj["launch_template"].([]interface{})
			if !ok {
				v = []interface{}{}
			}
			obj["launch_template"] = flattenNodeGroupLaunchTemplate(in.LaunchTemplate, v)
		}
		if len(in.Version) > 0 {
			obj["version"] = in.Version
		}
		out[i] = obj
	}
	return out, nil
}
func flattenNodeGroupTaint(inp []NodeGroupTaint, p []interface{}) []interface{} {
	if inp == nil {
		return nil
	}
	out := make([]interface{}, len(inp))
	for i, in := range inp {
		obj := map[string]interface{}{}
		if len(in.Key) > 0 {
			obj["key"] = in.Key
		}
		if len(in.Value) > 0 {
			obj["value"] = in.Value
		}
		if len(in.Effect) > 0 {
			obj["effect"] = in.Effect
		}
		out[i] = obj
	}
	return out
}
func flattenNodeGroupLaunchTemplate(in *LaunchTemplate, p []interface{}) []interface{} {
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}
	if in == nil {
		return []interface{}{obj}
	}
	if len(in.ID) > 0 {
		obj["id"] = in.ID
	}
	if len(in.Version) > 0 {
		obj["version"] = in.Version
	}
	return []interface{}{obj}
}

// Flatten Fargate Profiles
func flattenEKSClusterFargateProfiles(inp []*FargateProfile, p []interface{}) []interface{} {
	if inp == nil {
		return nil
	}
	out := make([]interface{}, len(inp))
	for i, in := range inp {
		obj := map[string]interface{}{}
		if len(in.Name) > 0 {
			obj["name"] = in.Name
		}
		if len(in.PodExecutionRoleARN) > 0 {
			obj["pod_execution_role_arn"] = in.PodExecutionRoleARN
		}
		if in.Selectors != nil {
			v, ok := obj["selectors"].([]interface{})
			if !ok {
				v = []interface{}{}
			}
			obj["selectors"] = flattenFargateProfileSelectors(in.Selectors, v)
		}
		if len(in.Subnets) > 0 {
			obj["subnets"] = toArrayInterface(in.Subnets)
		}
		if len(in.Tags) > 0 {
			obj["tags"] = toMapInterface(in.Tags)
		}
		if len(in.Status) > 0 {
			obj["status"] = in.Status
		}
		out[i] = obj
	}

	return out
}
func flattenFargateProfileSelectors(inp []FargateProfileSelector, p []interface{}) []interface{} {
	if inp == nil {
		return nil
	}
	out := make([]interface{}, len(inp))
	for i, in := range inp {
		obj := map[string]interface{}{}
		if len(in.Namespace) > 0 {
			obj["namespace"] = in.Namespace
		}
		if len(in.Labels) > 0 {
			obj["labels"] = toMapInterface(in.Labels)
		}
		out[i] = obj
	}
	return out
}

// flatten Cluster Cloudwatch
func flattenEKSClusterCloudWatch(in *EKSClusterCloudWatch, p []interface{}) []interface{} {
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}
	if in == nil {
		return []interface{}{obj}
	}
	if in.ClusterLogging != nil {
		v, ok := obj["cluster_logging"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["cluster_logging"] = flattenClusterCloudWatchLogging(in.ClusterLogging, v)
	}
	return []interface{}{obj}
}
func flattenClusterCloudWatchLogging(in *EKSClusterCloudWatchLogging, p []interface{}) []interface{} {
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}
	if in == nil {
		return []interface{}{obj}
	}
	if len(in.EnableTypes) > 0 {
		obj["enable_types"] = toArrayInterface(in.EnableTypes)
	}
	obj["log_retention_in_days"] = in.LogRetentionInDays
	return []interface{}{obj}
}

// flatten secret encryption
func flattenEKSClusterSecretsEncryption(in *SecretsEncryption, p []interface{}) []interface{} {
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}
	if in == nil {
		return []interface{}{obj}
	}
	if len(in.KeyARN) > 0 {
		obj["key_arn"] = in.KeyARN
	}
	return []interface{}{obj}
}

func flattenIdentityMappings(in *EKSClusterIdentityMappings, p []interface{}) ([]interface{}, error) {
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}
	if in == nil {
		return []interface{}{obj}, nil
	}

	if in.Arns != nil {
		v, ok := obj["arns"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["arns"] = flattenArnFields(in.Arns, v)
	}

	if len(in.Accounts) > 0 {
		obj["accounts"] = in.Accounts
	}

	return []interface{}{obj}, nil
}

func flattenArnFields(inp []*IdentityMappingARN, p []interface{}) []interface{} {
	if inp == nil {
		return nil
	}
	out := make([]interface{}, len(inp))

	for i, in := range inp {
		obj := map[string]interface{}{}
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}

		if len(in.Arn) > 0 {
			obj["arn"] = in.Arn
		}

		if len(in.Group) > 0 {
			obj["group"] = in.Group
		}

		if len(in.Username) > 0 {
			obj["username"] = in.Username
		}

		out[i] = &obj

	}
	return out
}

func getProjectIDFromName(projectName string) (string, error) {
	// derive project id from project name
	resp, err := project.GetProjectByName(projectName)
	if err != nil {
		log.Print("project name missing in the resource")
		return "", err
	}

	project, err := project.NewProjectFromResponse([]byte(resp))
	if err != nil {
		log.Printf("project does not exist")
		return "", err
	}
	return project.ID, nil
}

func resourceEKSClusterCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Println("create EKS cluster resource")
	return resourceEKSClusterUpsert(ctx, d, m)
}

func resourceEKSClusterRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Println("READ eks cluster")
	var diags diag.Diagnostics
	// find cluster name and project name
	clusterName, ok := d.Get("cluster.0.metadata.0.name").(string)
	if !ok || clusterName == "" {
		log.Print("Cluster name unable to be found")
		return diag.FromErr(fmt.Errorf("%s", "cluster name is missing"))
	}
	projectName, ok := d.Get("cluster.0.metadata.0.project").(string)
	if !ok || projectName == "" {
		log.Print("Cluster project name unable to be found")
		return diag.FromErr(fmt.Errorf("%s", "project name is missing"))
	}
	projectID, err := getProjectIDFromName(projectName)
	if err != nil {
		log.Print("error converting project name to id")
		return diag.Errorf("error converting project name to project ID")
	}
	c, err := cluster.GetCluster(clusterName, projectID)
	if err != nil {
		log.Printf("error in get cluster %s", err.Error())
		if strings.Contains(err.Error(), "not found") {
			log.Println("Resource Read ", "error", err)
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}
	log.Println("got cluster from backend")
	logger := glogger.GetLogger()
	rctlCfg := config.GetConfig()
	clusterSpecYaml, err := clusterctl.GetClusterSpec(logger, rctlCfg, c.Name, projectID)
	if err != nil {
		log.Printf("error in get clusterspec %s", err.Error())
		return diag.FromErr(err)
	}
	log.Println("resourceEKSClusterRead clusterSpec ", clusterSpecYaml)

	decoder := yaml.NewDecoder(bytes.NewReader([]byte(clusterSpecYaml)))

	clusterSpec := EKSCluster{}
	if err := decoder.Decode(&clusterSpec); err != nil {
		log.Println("error decoding cluster spec")
		return diag.FromErr(err)
	}

	clusterConfigSpec := EKSClusterConfig{}
	if err := decoder.Decode(&clusterConfigSpec); err != nil {
		log.Println("error decoding cluster config spec")
		return diag.FromErr(err)
	}

	v, ok := d.Get("cluster").([]interface{})
	if !ok {
		v = []interface{}{}
	}
	c1, err := flattenEKSCluster(&clusterSpec, v)
	log.Println("finished flatten eks cluster", c1)
	if err != nil {
		log.Printf("flatten eks cluster error %s", err.Error())
		return diag.FromErr(err)
	}
	err = d.Set("cluster", c1)
	if err != nil {
		log.Printf("err setting cluster %s", err.Error())
		return diag.FromErr(err)
	}

	v2, ok := d.Get("cluster_config").([]interface{})
	if !ok {
		v2 = []interface{}{}
	}
	c2, err := flattenEKSClusterConfig(&clusterConfigSpec, v2)
	if err != nil {
		log.Printf("flatten eks cluster config error %s", err.Error())
		return diag.FromErr(err)
	}
	err = d.Set("cluster_config", c2)
	if err != nil {
		log.Printf("err setting cluster config %s", err.Error())
		return diag.FromErr(err)
	}
	log.Println("flattened cluster fine")
	log.Println("finished read")
	return diags
}

func resourceEKSClusterUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// find cluster name and project name
	clusterName, ok := d.Get("cluster.0.metadata.0.name").(string)
	if !ok || clusterName == "" {
		log.Print("Cluster name unable to be found")
		return diag.FromErr(fmt.Errorf("%s", "cluster name is missing"))
	}
	projectName, ok := d.Get("cluster.0.metadata.0.project").(string)
	if !ok || projectName == "" {
		log.Print("Cluster project name unable to be found")
		return diag.FromErr(fmt.Errorf("%s", "project name is missing"))
	}
	projectID, err := getProjectIDFromName(projectName)
	if err != nil {
		log.Print("error converting project name to id")
		return diag.Errorf("error converting project name to project ID")
	}
	c, err := cluster.GetCluster(clusterName, projectID)
	if err != nil {
		log.Printf("error in get cluster %s", err.Error())
		return diag.FromErr(err)
	}
	if c.ID != d.Id() {
		log.Printf("edge id has changed, state: %s, current: %s", d.Id(), c.ID)
		return diag.Errorf("remote and state id mismatch")
	}
	log.Println("finished update")
	return resourceEKSClusterUpsert(ctx, d, m)
}

func resourceEKSClusterDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// find cluster name and project name
	clusterName, ok := d.Get("cluster.0.metadata.0.name").(string)
	if !ok || clusterName == "" {
		log.Print("Cluster name unable to be found")
		return diag.FromErr(fmt.Errorf("%s", "cluster name is missing"))
	}
	projectName, ok := d.Get("cluster.0.metadata.0.project").(string)
	if !ok || projectName == "" {
		log.Print("Cluster project name unable to be found")
		return diag.FromErr(fmt.Errorf("%s", "project name is missing"))
	}
	projectID, err := getProjectIDFromName(projectName)
	if err != nil {
		log.Print("error converting project name to id")
		return diag.Errorf("error converting project name to project ID")
	}

	errDel := cluster.DeleteCluster(clusterName, projectID, false)
	if errDel != nil {
		log.Printf("delete cluster error %s", errDel.Error())
		return diag.FromErr(errDel)
	}
	for {
		time.Sleep(60 * time.Second)
		check, errGet := cluster.GetCluster(clusterName, projectID)
		if errGet != nil {
			log.Printf("error while getCluster %s, delete success", errGet.Error())
			break
		}
		if check == nil || (check != nil && check.Status != "READY") {
			break
		}
	}
	log.Println("finished delete")

	return diag.Diagnostics{}
}

// Sort EKS Nodepool

// ByNodeGroupName struct
type ByNodeGroupName []NodeGroup

func (np ByNodeGroupName) Len() int      { return len(np) }
func (np ByNodeGroupName) Swap(i, j int) { np[i], np[j] = np[j], np[i] }
func (np ByNodeGroupName) Less(i, j int) bool {
	ret := strings.Compare(np[i].Name, np[j].Name)
	if ret < 0 {
		return true
	} else {
		return false
	}
}

type ByManagedNodeGroupName []ManagedNodeGroup

func (np ByManagedNodeGroupName) Len() int      { return len(np) }
func (np ByManagedNodeGroupName) Swap(i, j int) { np[i], np[j] = np[j], np[i] }
func (np ByManagedNodeGroupName) Less(i, j int) bool {
	ret := strings.Compare(np[i].Name, np[j].Name)
	if ret < 0 {
		return true
	} else {
		return false
	}
}

func resourceEKSClusterImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	idParts := strings.SplitN(d.Id(), "/", 2)
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		log.Printf("Invalid ID passed: %s", d.Id())
		return nil, fmt.Errorf("invalid ID passed: %s", d.Id())
	}

	clusterName := idParts[0]
	projectName := idParts[1]
	projectID, err := getProjectIDFromName(projectName)
	if err != nil {
		log.Printf("error converting project name to id: %s", projectName)
		return nil, fmt.Errorf("error converting project name to project ID")
	}

	s, errGet := cluster.GetCluster(clusterName, projectID)
	if errGet != nil {
		log.Printf("error while getCluster %s", errGet.Error())
		return nil, errGet
	}
	log.Printf("resource eks cluster import id %s", s.ID)

	clusters := []interface{}{
		map[string][]interface{}{
			"metadata": {
				map[string]interface{}{
					"name":    clusterName,
					"project": projectName,
				},
			},
		},
	}

	if err := d.Set("cluster", clusters); err != nil {
		log.Printf("error setting cluster in state to %+v", clusters)
		return nil, err
	}

	log.Printf("resource eks cluster import id %s", s.ID)
	d.SetId(s.ID)

	return []*schema.ResourceData{d}, nil
}
