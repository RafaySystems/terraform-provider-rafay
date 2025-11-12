package rafay

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/RafaySystems/rctl/pkg/cluster"
	"github.com/RafaySystems/rctl/pkg/clusterctl"
	"github.com/RafaySystems/rctl/pkg/config"
	glogger "github.com/RafaySystems/rctl/pkg/log"
	"github.com/davecgh/go-spew/spew"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	jsoniter "github.com/json-iterator/go"
	"k8s.io/utils/strings/slices"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/go-yaml/yaml"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

//go:embed resource_eks_cluster_description.md
var resourceEKSClusterDescription string

func resourceEKSCluster() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceEKSClusterCreate,
		ReadContext:   resourceEKSClusterRead,
		UpdateContext: resourceEKSClusterUpdate,
		DeleteContext: resourceEKSClusterDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(100 * time.Minute),
			Update: schema.DefaultTimeout(130 * time.Minute),
			Delete: schema.DefaultTimeout(70 * time.Minute),
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
			Type:        schema.TypeMap,
			Optional:    true,
			Description: "The proxy configuration for the cluster. Use this if the infrastructure uses an outbound proxy.",
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
		"access_config": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "controls how IAM principals can access this cluster",
			Elem: &schema.Resource{
				Schema: accessConfigFields(),
			},
		},
		"addons_config": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "addon config fields",
			Elem: &schema.Resource{
				Schema: addonsConfigurationsField(),
			},
		},
		"auto_mode_config": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "auto mode config fields",
			Elem: &schema.Resource{
				Schema: autoModeConfigurationsField(),
			},
		},
	}
	return s
}

func accessConfigFields() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"bootstrap_cluster_creator_admin_permissions": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "choose whether the IAM principal creating the cluster has Kubernetes cluster administrator access",
		},
		"authentication_mode": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "configure which source the cluster will use for authenticated IAM principals. API or API_AND_CONFIG_MAP (default) or CONFIG_MAP",
		},
		"access_entries": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "specifies a list of access entries for the cluster",
			Elem: &schema.Resource{
				Schema: accessEntryFields(),
			},
		},
	}
	return s
}

func accessEntryFields() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"principal_arn": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "the IAM principal that you want to grant access to Kubernetes objects on your cluster",
		},
		"type": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "EC2_LINUX, EC2_WINDOWS, FARGATE_LINUX or STANDARD",
		},
		"kubernetes_username": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "username to map to the principal ARN",
		},
		"kubernetes_groups": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "set of Kubernetes groups to map to the principal ARN",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"tags": {
			Type:        schema.TypeMap,
			Optional:    true,
			Description: "applied to the access entries",
		},
		"access_policies": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "set of policies to associate with an access entry",
			Elem: &schema.Resource{
				Schema: accessPolicyFields(),
			},
		},
	}
	return s
}

func accessPolicyFields() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"policy_arn": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "the ARN of the policy to attach to the access entry",
		},
		"access_scope": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "defines the scope of an access policy",
			Elem: &schema.Resource{
				Schema: accessScopeFields(),
			},
		},
	}
	return s
}

func accessScopeFields() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"type": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "namespace or cluster",
		},
		"namespaces": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Scope access to namespace(s)",
			Elem: &schema.Schema{
				Type: schema.TypeString,
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
		"pod_identity_associations": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "pod identity associations",
			Elem: &schema.Resource{
				Schema: podIdentityAssociationsFields(),
			},
		},
	}
	return s
}

func podIdentityAssociationsFields() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"namespace": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "namespace of service account",
		},
		"service_account_name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "name of service account",
		},
		"role_arn": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "role ARN of AWS role to associate with service account",
		},
		"create_service_account": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "enable flag to create service account",
			DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
				// During CREATE the resource is still "new" → allow the diff.
				// Afterwards, suppress any attempted change so TF doesn't even
				// try to plan it; the Update code will throw a hard error.
				return !d.IsNewResource()
			},
		},
		"role_name": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "aws role name to associate",
		},
		"permission_boundary_arn": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "permission boundary ARN",
		},
		"permission_policy_arns": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "permission policy ARNs",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"permission_policy": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "permission policy document",
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
			Description: "AWS tags for the service account",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
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
		"id": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Attach policy id",
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
		"condition": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Attach policy Statement",
		},
		"sid": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Sid of policy",
		},
		"not_action": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Attach policy NotAction",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"not_resource": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Attach policy NotResource",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"principal": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Attach policy principal",
		},
		"not_principal": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Attach policy NotPrincipal",
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

func addonsConfigurationsField() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"auto_apply_pod_identity_associations": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "Flag to create pod identity by default for managed addons",
		},
		"disable_ebs_csi_driver": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "flag to enable or disable ebs csi driver",
		},
	}
	return s
}

func autoModeConfigurationsField() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"enabled": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "Enable auto mode in EKS",
		},
		"node_role_arn": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "RoleARN of the nodes",
		},
		"node_pools": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "List of default nodepools (general-purpose,system)",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
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
		"attach_policy_v2": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "holds a policy document to attach to this addon in json string format",
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
		"pod_identity_associations": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "pod identity associations",
			Elem: &schema.Resource{
				Schema: podIdentityAssociationsFields(),
			},
		},
		"use_default_pod_identity_associations": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "Flag to create pod identity association by default",
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
		"attach_policy_v2": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "attach policy in json string format ",
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
		"encrypt_existing_secrets": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "Flag to encrypt existing secrets. Default is true",
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
func expandEKSClusterConfig(p []interface{}, rawConfig cty.Value) (*EKSClusterConfig, error) {
	obj := &EKSClusterConfig{}

	if len(p) == 0 || p[0] == nil {
		return obj, nil
	}
	in := p[0].(map[string]interface{})
	if !rawConfig.IsNull() && len(rawConfig.AsValueSlice()) > 0 {
		rawConfig = rawConfig.AsValueSlice()[0]
	}
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
		var nRawConfig cty.Value
		if !rawConfig.IsNull() {
			nRawConfig = rawConfig.GetAttr("vpc")
		}
		vpc, err := expandVPC(v, nRawConfig)
		if err != nil {
			return nil, err
		}
		obj.VPC = vpc
	}
	if v, ok := in["managed_nodegroups"].([]interface{}); ok && len(v) > 0 {
		var nRawConfig cty.Value
		if !rawConfig.IsNull() {
			nRawConfig = rawConfig.GetAttr("managed_nodegroups")
		}
		obj.ManagedNodeGroups = expandManagedNodeGroups(v, nRawConfig)
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
	if v, ok := in["access_config"].([]interface{}); ok && len(v) > 0 {
		obj.AccessConfig = expandAccessConfig(v)
	}
	if v, ok := in["addons_config"].([]interface{}); ok && len(v) > 0 {
		obj.AddonsConfig = expandAddonsConfig(v)
	}
	if v, ok := in["auto_mode_config"].([]interface{}); ok && len(v) > 0 {
		obj.AutoModeConfig = expandAutoModeConfig(v)
	}
	return obj, nil
}

func processEKSInputs(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	//building cluster and cluster config yaml file
	var yamlCluster *EKSCluster
	var yamlClusterConfig *EKSClusterConfig
	var err error
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
		yamlClusterConfig, err = expandEKSClusterConfig(v, rawConfig.GetAttr("cluster_config"))
		if err != nil {
			log.Print("Invalid cluster config found")
			return diag.FromErr(err)
		}
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
	log.Printf("process_filebytes %s", projectName)
	projectID, err := getProjectIDFromName(projectName)
	if err != nil {
		log.Printf("error converting project name to id %s", err.Error())
		return diag.Errorf("error converting project name to project ID")
	}

	// Specific to create flow: If `spec.sharing` specified then
	// set "cluster_sharing_external" to false.
	var cse string
	if yamlClusterMetadata.Spec.Sharing != nil {
		cse = "false"
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
	response, err := clusterctl.Apply(logger, rctlConfig, clusterName, b.Bytes(), false, false, false, false, uaDef, cse)
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
	s, errGet := cluster.GetCluster(clusterName, projectID, uaDef)
	if errGet != nil {
		log.Printf("error while getCluster %s", errGet.Error())
		return diag.FromErr(errGet)
	}

	log.Println("Cluster Provision may take upto 15-20 Minutes")
	d.SetId(s.ID)

	ticker := time.NewTicker(time.Duration(60) * time.Second)
	defer ticker.Stop()
LOOP:
	for {
		//Check for cluster operation timeout
		select {
		case <-ctx.Done():
			log.Println("Cluster operation stopped due to operation timeout.")
			return diag.Errorf("cluster operation stopped for cluster: `%s` due to operation timeout", clusterName)
		case <-ticker.C:
			log.Printf("Cluster operation not completed for edgename: %s and projectname: %s. Waiting 60 seconds more for cluster to complete the operation.", clusterName, projectName)
			check, errGet := cluster.GetCluster(yamlClusterMetadata.Metadata.Name, projectID, uaDef)
			if errGet != nil {
				log.Printf("error while getCluster %s", errGet.Error())
				return diag.FromErr(errGet)
			}
			edgeId := check.ID
			check, errGet = cluster.GetClusterWithEdgeID(edgeId, projectID, uaDef)
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
				log.Println("Checking in cluster conditions for blueprint sync success..")
				conditionsFailure, clusterReadiness, err := getClusterConditions(edgeId, projectID)
				if err != nil {
					log.Printf("error while getCluster %s", err.Error())
					return diag.FromErr(err)
				}
				if conditionsFailure {
					log.Printf("blueprint sync failed for edgename: %s and projectname: %s", clusterName, projectName)
					return diag.FromErr(fmt.Errorf("blueprint sync failed for edgename: %s and projectname: %s", clusterName, projectName))
				} else if clusterReadiness {
					log.Printf("Cluster operation completed for edgename: %s and projectname: %s", clusterName, projectName)
					break LOOP
				} else {
					log.Println("Cluster Provisiong is Complete. Waiting for cluster to be Ready...")
				}
			} else if strings.Contains(sres.Status, "STATUS_FAILED") {
				return diag.FromErr(fmt.Errorf("failed to create/update cluster while provisioning cluster %s %s", clusterName, statusResp))
			} else {
				log.Printf("Cluster operation not completed for edgename: %s and projectname: %s. Waiting 60 seconds more for cluster to complete the operation.", clusterName, projectName)
			}
		}
	}

	edgeDb, err := cluster.GetCluster(clusterName, projectID, uaDef)
	if err != nil {
		tflog.Error(ctx, "failed to get cluster", map[string]any{"name": clusterName, "pid": projectID})
		return diag.Errorf("Failed to fetch cluster: %s", err)
	}
	cseFromDb := edgeDb.Settings[clusterSharingExtKey]
	if cseFromDb != "true" {
		if yamlClusterMetadata.Spec.Sharing == nil && cseFromDb != "" {
			// reset cse as sharing is removed
			edgeDb.Settings[clusterSharingExtKey] = ""
			err := cluster.UpdateCluster(edgeDb, uaDef)
			if err != nil {
				tflog.Error(ctx, "failed to update cluster", map[string]any{"edgeObj": edgeDb})
				return diag.Errorf("Unable to update the edge object, got error: %s", err)
			}
			tflog.Error(ctx, "cse removed successfully")
		}
		if yamlClusterMetadata.Spec.Sharing != nil && cseFromDb != "false" {
			// explicitly set cse to false
			edgeDb.Settings[clusterSharingExtKey] = "false"
			err := cluster.UpdateCluster(edgeDb, uaDef)
			if err != nil {
				tflog.Error(ctx, "failed to update cluster", map[string]any{"edgeObj": edgeDb})
				return diag.Errorf("Unable to update the edge object, got error: %s", err)
			}
			tflog.Error(ctx, "cse set to false")
		}
	}

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

func expandAccessConfig(p []interface{}) *EKSClusterAccess {
	obj := &EKSClusterAccess{}

	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["bootstrap_cluster_creator_admin_permissions"].(bool); ok {
		obj.BootstrapClusterCreatorAdminPermissions = v
	}

	if v, ok := in["authentication_mode"].(string); ok && len(v) > 0 {
		obj.AuthenticationMode = v
	}

	if v, ok := in["access_entries"].([]interface{}); ok && len(v) > 0 {
		obj.AccessEntries = expandAccessEntries(v)
	}

	return obj
}

func expandAccessEntries(p []interface{}) []*EKSAccessEntry {
	out := make([]*EKSAccessEntry, len(p))
	if len(p) == 0 || p[0] == nil {
		return out
	}

	for i := range p {
		in := p[i].(map[string]interface{})
		obj := &EKSAccessEntry{}
		if v, ok := in["principal_arn"].(string); ok && len(v) > 0 {
			obj.PrincipalARN = v
		}
		if v, ok := in["type"].(string); ok && len(v) > 0 {
			obj.Type = v
		}
		if v, ok := in["kubernetes_username"].(string); ok && len(v) > 0 {
			obj.KubernetesUsername = v
		}
		if v, ok := in["kubernetes_groups"].([]interface{}); ok && len(v) > 0 {
			obj.KubernetesGroups = toArrayString(v)
		}
		if v, ok := in["tags"].(map[string]interface{}); ok && len(v) > 0 {
			obj.Tags = toMapString(v)
		}
		if v, ok := in["access_policies"].([]interface{}); ok && len(v) > 0 {
			obj.AccessPolicies = expandAccessPolicies(v)
		}

		out[i] = obj
	}

	return out
}

func expandAccessPolicies(p []interface{}) []*EKSAccessPolicy {

	out := make([]*EKSAccessPolicy, len(p))
	if len(p) == 0 || p[0] == nil {
		return out
	}

	for i := range p {
		in := p[i].(map[string]interface{})
		obj := &EKSAccessPolicy{}
		if v, ok := in["policy_arn"].(string); ok && len(v) > 0 {
			obj.PolicyARN = v
		}
		if v, ok := in["access_scope"].([]interface{}); ok && len(v) > 0 {
			obj.AccessScope = expandAccessScope(v)
		}
		out[i] = obj
	}

	return out
}

func expandAccessScope(p []interface{}) *EKSAccessScope {
	obj := &EKSAccessScope{}

	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["type"].(string); ok && len(v) > 0 {
		obj.Type = v
	}
	if v, ok := in["namespaces"].([]interface{}); ok && len(v) > 0 {
		obj.Namespaces = toArrayString(v)
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

	if v, ok := in["encrypt_existing_secrets"].(bool); ok {
		obj.EncryptExistingSecrets = &v
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
	if len(p) == 0 || p[0] == nil {
		return out
	}
	log.Println("got to managed node group")
	for i := range p {
		obj := &ManagedNodeGroup{}
		in := p[i].(map[string]interface{})
		// nRawConfig := rawConfig.AsValueSlice()[i]
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
			var nRawConfig cty.Value
			if !rawConfig.IsNull() && i < len(rawConfig.AsValueSlice()) {
				nRawConfig = rawConfig.AsValueSlice()[i].GetAttr("security_groups")
			}
			obj.SecurityGroups = expandManagedNodeGroupSecurityGroups(v, nRawConfig)
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
		out[i] = obj
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
		out[i] = &obj
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
		if err := json2.Unmarshal([]byte(v), &policyDoc); err != nil {
			log.Printf("warning: failed to unmarshal bottle rocket settings: %v", err)
		} else {
			obj.Settings = policyDoc
			log.Println("bottle rocket settings expanded correct")
		}
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
	if !rawConfig.IsNull() && len(rawConfig.AsValueSlice()) > 0 {
		rawConfig = rawConfig.AsValueSlice()[0]
	}

	if v, ok := in["attach_ids"].([]interface{}); ok && len(v) > 0 {
		obj.AttachIDs = toArrayString(v)
	}

	var rawWithShared cty.Value
	if !rawConfig.IsNull() {
		rawWithShared = rawConfig.GetAttr("with_shared")
	}
	if !rawWithShared.IsNull() {
		boolVal := rawWithShared.True()
		obj.WithShared = &boolVal
	}

	var rawWithLocal cty.Value
	if !rawConfig.IsNull() {
		rawWithLocal = rawConfig.GetAttr("with_shared")
	}
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

	if v, ok := in["attach_policy_v2"].(string); ok && len(v) > 0 {
		var policyDoc *InlineDocument
		var json2 = jsoniter.ConfigCompatibleWithStandardLibrary
		if err := json2.Unmarshal([]byte(v), &policyDoc); err != nil {
			log.Printf("warning: failed to unmarshal attach policy: %v", err)
		} else {
			obj.AttachPolicy = policyDoc
			//log.Println("attach policy expanded correct")
		}
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
func expandStatement(p []interface{}) []InlineStatement {
	out := make([]InlineStatement, len(p))

	if len(p) == 0 || p[0] == nil {
		return out
	}

	for i := range p {
		obj := &InlineStatement{}
		in := p[0].(map[string]interface{})
		if v, ok := in["effect"].(string); ok && len(v) > 0 {
			obj.Effect = v
		}
		if v, ok := in["action"].([]interface{}); ok && len(v) > 0 {
			obj.Action = toArrayStringSorted(v)
		}
		if v, ok := in["not_action"].([]interface{}); ok && len(v) > 0 {
			obj.NotAction = toArrayStringSorted(v)
		}
		if v, ok := in["resource"].(string); ok && len(v) > 0 {
			obj.Resource = v
		}
		if v, ok := in["not_resource"].([]interface{}); ok && len(v) > 0 {
			obj.NotResource = toArrayStringSorted(v)
		}
		if v, ok := in["sid"].(string); ok && len(v) > 0 {
			obj.Sid = v
		}
		if v, ok := in["condition"].(string); ok && len(v) > 0 {
			var policyDoc map[string]interface{}
			var json2 = jsoniter.ConfigCompatibleWithStandardLibrary
			//json.Unmarshal(input, &data)
			json2.Unmarshal([]byte(v), &policyDoc)
			obj.Condition = policyDoc
		}

		if v, ok := in["principal"].(string); ok && len(v) > 0 {
			var policyDoc map[string]interface{}
			var json2 = jsoniter.ConfigCompatibleWithStandardLibrary
			//json.Unmarshal(input, &data)
			json2.Unmarshal([]byte(v), &policyDoc)
			obj.Principal = policyDoc
		}

		if v, ok := in["not_principal"].(string); ok && len(v) > 0 {
			var policyDoc map[string]interface{}
			var json2 = jsoniter.ConfigCompatibleWithStandardLibrary
			//json.Unmarshal(input, &data)
			json2.Unmarshal([]byte(v), &policyDoc)
			obj.NotPrincipal = policyDoc
		}

		out[i] = *obj
	}

	return out
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
	if v, ok := in["id"].(string); ok && len(v) > 0 {
		obj.Id = v
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

func expandAddonsConfig(p []interface{}) *EKSAddonsConfig {
	obj := &EKSAddonsConfig{}

	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})
	if v, ok := in["auto_apply_pod_identity_associations"].(bool); ok {
		obj.AutoApplyPodIdentityAssociations = v
	}
	if v, ok := in["disable_ebs_csi_driver"].(bool); ok {
		obj.DisableEBSCSIDriver = v
	}

	return obj
}

func expandAutoModeConfig(p []interface{}) *EKSAutoModeConfig {
	obj := &EKSAutoModeConfig{}

	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})
	if v, ok := in["enabled"].(bool); ok {
		obj.Enabled = v
	}
	if v, ok := in["node_role_arn"].(string); ok {
		obj.NodeRoleARN = v
	}
	if v, ok := in["node_pools"].([]interface{}); ok && len(v) > 0 {
		obj.NodePools = toArrayString(v)
	}

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

		if v, ok := in["attach_policy_v2"].(string); ok && len(v) > 0 {
			var policyDoc *InlineDocument
			var json2 = jsoniter.ConfigCompatibleWithStandardLibrary
			//json.Unmarshal(input, &data)
			json2.Unmarshal([]byte(v), &policyDoc)
			obj.AttachPolicy = policyDoc
			//log.Println("attach policy expanded correct")
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
		if v, ok := in["pod_identity_associations"].([]interface{}); ok && len(v) > 0 {
			obj.PodIdentityAssociations = expandIAMPodIdentityAssociationsConfig(v)
		}
		if v, ok := in["use_default_pod_identity_associations"].(bool); ok {
			obj.UseDefaultPodIdentityAssociations = v
		}
		//docs dont have force variable but struct does
		out[i] = obj
	}
	return out
}

// expand vpc function
func expandVPC(p []interface{}, rawConfig cty.Value) (*EKSClusterVPC, error) {
	obj := &EKSClusterVPC{}

	if len(p) == 0 || p[0] == nil {
		return obj, nil
	}
	in := p[0].(map[string]interface{})
	if !rawConfig.IsNull() && len(rawConfig.AsValueSlice()) > 0 {
		rawConfig = rawConfig.AsValueSlice()[0]
	}

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
		clusterSubnets, err := expandSubnets(v)
		if err != nil {
			return nil, err
		}
		obj.Subnets = clusterSubnets
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
	var rawManageSharedNodeSecurityGroupRules cty.Value
	if !rawConfig.IsNull() {
		rawManageSharedNodeSecurityGroupRules = rawConfig.GetAttr("manage_shared_node_security_group_rules")
	}
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
	return obj, nil
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

func expandSubnets(p []interface{}) (*ClusterSubnets, error) {
	obj := &ClusterSubnets{}

	if len(p) == 0 || p[0] == nil {
		return obj, nil
	}
	in := p[0].(map[string]interface{})
	if v, ok := in["private"].([]interface{}); ok && len(v) > 0 {
		subnets, err := expandSubnetSpec(v)
		if err != nil {
			return nil, err
		}
		obj.Private = subnets
	}
	if v, ok := in["public"].([]interface{}); ok && len(v) > 0 {
		subnets, err := expandSubnetSpec(v)
		if err != nil {
			return nil, err
		}
		obj.Public = subnets
	}
	return obj, nil
}
func expandSubnetSpec(p []interface{}) (AZSubnetMapping, error) {
	obj := make(AZSubnetMapping)
	namesFrequency := make(map[string]int)
	duplicateNames := make([]string, 0)

	if len(p) == 0 || p[0] == nil {
		return obj, nil
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
			namesFrequency[v]++
			if namesFrequency[v] == 2 {
				duplicateNames = append(duplicateNames, v)
			}
		}
	}

	if len(duplicateNames) > 0 {
		return nil, fmt.Errorf("duplicate subnet names found: %v. Kindly use unique name for every subnet configured for better experience", duplicateNames)
	}

	return obj, nil
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

	if v, ok := in["pod_identity_associations"].([]interface{}); ok && len(v) > 0 {
		obj.PodIdentityAssociations = expandIAMPodIdentityAssociationsConfig(v)
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

func expandIAMPodIdentityAssociationsConfig(p []interface{}) []*IAMPodIdentityAssociation {
	out := make([]*IAMPodIdentityAssociation, len(p))
	if len(p) == 0 || p[0] == nil {
		return out
	}
	for i := range p {
		obj := &IAMPodIdentityAssociation{}
		in := p[i].(map[string]interface{})
		if v, ok := in["namespace"].(string); ok && len(v) > 0 {
			obj.Namespace = v
		}
		if v, ok := in["service_account_name"].(string); ok && len(v) > 0 {
			obj.ServiceAccountName = v
		}
		if v, ok := in["role_arn"].(string); ok && len(v) > 0 {
			obj.RoleARN = v
		}
		if v, ok := in["create_service_account"].(bool); ok {
			obj.CreateServiceAccount = v
		}
		if v, ok := in["role_name"].(string); ok && len(v) > 0 {
			obj.RoleName = v
		}
		if v, ok := in["permission_boundary_arn"].(string); ok && len(v) > 0 {
			obj.PermissionsBoundaryARN = v
		}
		if v, ok := in["permission_policy"].(string); ok && len(v) > 0 {
			var policyDoc map[string]interface{}
			var json2 = jsoniter.ConfigCompatibleWithStandardLibrary
			//json.Unmarshal(input, &data)
			json2.Unmarshal([]byte(v), &policyDoc)
			obj.PermissionPolicy = policyDoc
		}
		if v, ok := in["permission_policy_arns"].([]interface{}); ok && len(v) > 0 {
			obj.PermissionPolicyARNs = toArrayString(v)
		}
		if v, ok := in["well_known_policies"].([]interface{}); ok && len(v) > 0 {
			obj.WellKnownPolicies = expandIAMWellKnownPolicies(v)
		}
		if v, ok := in["tags"].(map[string]interface{}); ok && len(v) > 0 {
			obj.Tags = toMapString(v)
		}
		out[i] = obj
	}
	return out
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
	if v, ok := in["proxy_config"].(map[string]interface{}); ok && len(v) > 0 {
		obj.ProxyConfig = expandProxyConfig(v)
	}
	if v, ok := in["system_components_placement"].([]interface{}); ok && len(v) > 0 {
		obj.SystemComponentsPlacement = expandSystemComponentsPlacement(v)
	}
	if v, ok := in["sharing"].([]interface{}); ok && len(v) > 0 {
		obj.Sharing = expandV1ClusterSharing(v)
	}
	log.Println("cluster spec cloud_provider: ", obj.CloudProvider)

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

func expandProxyConfig(p map[string]interface{}) *ProxyConfig {
	obj := &ProxyConfig{}
	log.Println("expandProxyConfig")

	if len(p) == 0 {
		return obj
	}
	in := p

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

	if v, ok := in["enabled"].(string); ok {
		x, _ := strconv.ParseBool(v)
		obj.Enabled = x
	}

	if v, ok := in["allow_insecure_bootstrap"].(string); ok {
		x, _ := strconv.ParseBool(v)
		obj.AllowInsecureBootstrap = x
	}

	return obj

}

func flattenEKSCluster(in *EKSCluster, p []interface{}, rawState cty.Value) ([]interface{}, error) {
	obj := map[string]interface{}{}
	if in == nil {
		return nil, fmt.Errorf("empty cluster input")
	}
	if !rawState.IsNull() && len(rawState.AsValueSlice()) > 0 {
		rawState = rawState.AsValueSlice()[0]
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
		var nRawState cty.Value
		if !rawState.IsNull() && len(rawState.AsValueSlice()) > 0 {
			nRawState = rawState.GetAttr("spec")
		}
		ret2, err = flattenEKSClusterSpec(in.Spec, v, nRawState)
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
func flattenEKSClusterSpec(in *EKSSpec, p []interface{}, rawState cty.Value) ([]interface{}, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "flattenEKSClusterMetaData empty input")
	}
	obj := map[string]interface{}{}

	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}
	if !rawState.IsNull() && len(rawState.AsValueSlice()) > 0 {
		rawState = rawState.AsValueSlice()[0]
	}

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
		var nRawState cty.Value
		// if !rawState.IsNull() {
		// 	if rawState.Type().IsObjectType() {
		// 		if rawState.Type().HasAttribute("cni_params") {
		// 			cniRaw := rawState.GetAttr("cni_params")
		// 			if cniRaw.Type().IsListType() || cniRaw.Type().IsTupleType() {
		// 				nRawState = cniRaw
		// 				log.Println("Rawstate found for cni_params")
		// 			}
		// 		}
		// 	} else if rawState.Type().IsListType() || rawState.Type().IsTupleType() {
		// 		nRawState = rawState
		// 	}
		// }
		if !rawState.IsNull() && (rawState.Type().IsListType() || rawState.Type().IsTupleType()) {
			nRawState = rawState
		}
		obj["cni_params"] = flattenCNIParams(in.CniParams, v, nRawState)
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
		obj["sharing"] = flattenV1ClusterSharing(in.Sharing)
	}

	return []interface{}{obj}, nil
}

func flattenCNIParams(in *CustomCni, p []interface{}, rawState cty.Value) []interface{} {
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}
	if !rawState.IsNull() && len(rawState.AsValueSlice()) > 0 {
		rawState = rawState.AsValueSlice()[0]
	}

	if len(in.CustomCniCidr) > 0 {
		obj["custom_cni_cidr"] = in.CustomCniCidr
	}
	if in.CustomCniCrdSpec != nil {
		v, ok := obj["custom_cni_crd_spec"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		var nRawState cty.Value
		// if !rawState.IsNull() {
		// 	if rawState.Type().IsObjectType() && rawState.Type().HasAttribute("custom_cni_crd_spec") {
		// 		attr := rawState.GetAttr("custom_cni_crd_spec")
		// 		if attr.Type().IsListType() || attr.Type().IsTupleType() {
		// 			nRawState = attr
		// 		}
		// 	} else if rawState.Type().IsListType() || rawState.Type().IsTupleType() {
		// 		nRawState = rawState
		// 	}
		// }
		if !rawState.IsNull() && (rawState.Type().IsListType() || rawState.Type().IsTupleType()) {
			nRawState = rawState
		}
		obj["custom_cni_crd_spec"] = flattenCustomCNISpec(in.CustomCniCrdSpec, v, nRawState)
	}

	return []interface{}{obj}
}

func flattenCustomCNISpec(in map[string][]CustomCniSpec, p []interface{}, rawState cty.Value) []interface{} {
	log.Println("got to flatten custom CNI mapping", len(p))

	findLocalOrder := func(rawState cty.Value) []string {
		var order []string
		if !rawState.IsNull() {
			for _, crdSpec := range rawState.AsValueSlice() {
				if subnetValue, ok := crdSpec.AsValueMap()["name"]; ok {
					order = append(order, subnetValue.AsString())
				}
			}
		}
		return order
	}

	indexOf := func(item string, list []string) int {
		for i, v := range list {
			if v == item {
				return i
			}
		}
		return -1
	}

	azOrderInState := findLocalOrder(rawState)

	var out []interface{}
	for _, key := range azOrderInState {
		elem, ok := in[key]
		if !ok {
			// found only in local, ignore this.
			continue
		}
		obj := map[string]interface{}{
			"name":     key,
			"cni_spec": flattenCNISpec(elem, []interface{}{}),
		}
		out = append(out, obj)
	}
	for key, elem := range in {
		if i := indexOf(key, azOrderInState); i < 0 {
			// not found in local copy. this subnet was removed by the user
			obj := map[string]interface{}{
				"name":     key,
				"cni_spec": flattenCNISpec(elem, []interface{}{}),
			}
			out = append(out, obj)
		}
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

func flattenProxyConfig(in *ProxyConfig) map[string]interface{} {
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
	if in.Enabled {
		obj["enabled"] = "true"
	}
	if in.AllowInsecureBootstrap {
		obj["allow_insecure_bootstrap"] = "true"
	}

	return obj

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
func flattenEKSClusterConfig(in *EKSClusterConfig, rawState cty.Value, p []interface{}) ([]interface{}, error) {
	if in == nil {
		return nil, fmt.Errorf("empty cluster config input")
	}
	obj := map[string]interface{}{}
	if !rawState.IsNull() && len(rawState.AsValueSlice()) > 0 {
		rawState = rawState.AsValueSlice()[0]
	}
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
		var nRawState cty.Value
		if !rawState.IsNull() && len(rawState.AsValueSlice()) > 0 {
			nRawState = rawState.GetAttr("iam")
		}
		ret3, err = flattenEKSClusterIAM(in.IAM, nRawState, v)
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
		var nRawState cty.Value
		if !rawState.IsNull() && len(rawState.AsValueSlice()) > 0 {
			nRawState = rawState.GetAttr("addons")
		}
		ret6, err = flattenEKSClusterAddons(in.Addons, nRawState, v)
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
		var nRawState cty.Value
		if !rawState.IsNull() && len(rawState.AsValueSlice()) > 0 {
			nRawState = rawState.GetAttr("node_groups")
		}
		ret8 = flattenEKSClusterNodeGroups(in.NodeGroups, nRawState, v)
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
		var nRawState cty.Value
		if !rawState.IsNull() && len(rawState.AsValueSlice()) > 0 {
			nRawState = rawState.GetAttr("managed_nodegroups")
		}
		ret9, err = flattenEKSClusterManagedNodeGroups(in.ManagedNodeGroups, nRawState, v)
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
	var ret13 []interface{}
	if in.IdentityMappings != nil {
		v, ok := obj["identity_mappings"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		if len(in.IdentityMappings.Arns) != 0 || len(in.IdentityMappings.Accounts) != 0 {
			ret13, err = flattenIdentityMappings(in.IdentityMappings, v)
			if err != nil {
				log.Println("flattenIdentityMapping err")
				return nil, err
			}
		}
		obj["identity_mappings"] = ret13
	}

	if in.AccessConfig != nil {
		v, ok := obj["access_config"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		ret14, err := flattenEKSClusterAccess(in.AccessConfig, v)
		if err != nil {
			log.Println("flattenEKSClusterAccess err")
			return nil, err
		}
		obj["access_config"] = ret14
	}

	var ret14 []interface{}
	if in.AddonsConfig != nil {
		v, ok := obj["addons_config"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		ret14 = flattenEKSClusterAddonsConfig(in.AddonsConfig, v)
		/*if err != nil {
			log.Println("flattenEKSClusterSecretsEncryption err")
			return nil, err
		}*/
		obj["addons_config"] = ret14
	}

	var ret15 []interface{}
	if in.AutoModeConfig != nil {
		v, ok := obj["auto_mode_config"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		ret15 = flattenAutoModeConfig(in.AutoModeConfig, v)
		obj["auto_mode_config"] = ret15
	}

	log.Println("end of flatten config")

	return []interface{}{obj}, nil
}

func flattenEKSClusterAccess(in *EKSClusterAccess, p []interface{}) ([]interface{}, error) {
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}
	if in == nil {
		return []interface{}{obj}, nil
	}

	obj["bootstrap_cluster_creator_admin_permissions"] = in.BootstrapClusterCreatorAdminPermissions

	if in.AuthenticationMode != "" {
		obj["authentication_mode"] = in.AuthenticationMode
	}

	if in.AccessEntries != nil {
		v, ok := obj["access_entries"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		ret, err := flattenEKSAccessEntry(in.AccessEntries, v)
		if err != nil {
			log.Println("flattenEKSAccessEntry err")
			return nil, err
		}
		obj["access_entries"] = ret
	}

	return []interface{}{obj}, nil
}

func flattenEKSAccessEntry(inp []*EKSAccessEntry, p []interface{}) ([]interface{}, error) {

	out := make([]interface{}, len(inp))
	if inp == nil {
		return []interface{}{out}, nil
	}

	for i, in := range inp {
		obj := map[string]interface{}{}

		if len(in.PrincipalARN) > 0 {
			obj["principal_arn"] = in.PrincipalARN
		}
		if len(in.Type) > 0 {
			obj["type"] = in.Type
		}
		if len(in.KubernetesUsername) > 0 {
			obj["kubernetes_username"] = in.KubernetesUsername
		}

		if in.KubernetesGroups != nil && len(in.KubernetesGroups) > 0 {
			obj["kubernetes_groups"] = toArrayInterfaceSorted(in.KubernetesGroups)
		}
		if in.Tags != nil && len(in.Tags) > 0 {
			obj["tags"] = toMapInterface(in.Tags)
		}
		if in.AccessPolicies != nil {
			v, ok := obj["access_policies"].([]interface{})
			if !ok {
				v = []interface{}{}
			}
			ret, err := flattenEKSAccessPolicy(in.AccessPolicies, v)
			if err != nil {
				log.Println("flattenEKSAccessPolicy err")
				return nil, err
			}
			obj["access_policies"] = ret

		}

		out[i] = obj
	}

	return out, nil
}

func flattenEKSAccessPolicy(inp []*EKSAccessPolicy, p []interface{}) ([]interface{}, error) {
	out := make([]interface{}, len(inp))
	if inp == nil {
		return []interface{}{out}, nil
	}

	for i, in := range inp {
		obj := map[string]interface{}{}
		if in.PolicyARN != "" {
			obj["policy_arn"] = in.PolicyARN
		}

		if in.AccessScope != nil {
			v, ok := obj["access_scope"].([]interface{})
			if !ok {
				v = []interface{}{}
			}
			ret, err := flattenEKSAccessScope(in.AccessScope, v)
			if err != nil {
				log.Println("flattenEKSAccessScope err")
				return nil, err
			}
			obj["access_scope"] = ret
		}

		out[i] = obj
	}
	return out, nil
}

func flattenEKSAccessScope(in *EKSAccessScope, p []interface{}) ([]interface{}, error) {
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}
	if in == nil {
		return []interface{}{obj}, nil
	}

	if in.Type != "" {
		obj["type"] = in.Type
	}

	if in.Namespaces != nil && len(in.Namespaces) > 0 {
		obj["namespaces"] = toArrayInterfaceSorted(in.Namespaces)
	}

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
func flattenEKSClusterIAM(in *EKSClusterIAM, rawState cty.Value, p []interface{}) ([]interface{}, error) {
	obj := map[string]interface{}{}
	if !rawState.IsNull() && len(rawState.AsValueSlice()) > 0 {
		rawState = rawState.AsValueSlice()[0]
	}
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

	if in.ServiceAccounts != nil && len(in.ServiceAccounts) > 0 {
		v, ok := obj["service_accounts"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		var nRawState cty.Value
		if !rawState.IsNull() && rawState.Type().IsObjectType() {
			if _, ok := rawState.Type().AttributeTypes()["service_accounts"]; ok {
				nRawState = rawState.GetAttr("service_accounts")
			}
		}
		obj["service_accounts"] = flattenIAMServiceAccounts(in.ServiceAccounts, nRawState, v)
	}

	if in.PodIdentityAssociations != nil {
		v, ok := obj["pod_identity_associations"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["pod_identity_associations"] = flattenIAMPodIdentityAssociations(in.PodIdentityAssociations, v)
	}

	obj["vpc_resource_controller_policy"] = in.VPCResourceControllerPolicy

	return []interface{}{obj}, nil
}

func flattenIAMServiceAccountMetadata(in *EKSClusterIAMMeta, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
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

func flattenSingleIAMServiceAccount(in *EKSClusterIAMServiceAccount) map[string]interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{
		"metadata":            flattenIAMServiceAccountMetadata(in.Metadata, []interface{}{}),
		"well_known_policies": flattenIAMWellKnownPolicies(in.WellKnownPolicies, []interface{}{}),
		"role_only":           in.RoleOnly,
	}
	if in.AttachPolicyARNs != nil && len(in.AttachPolicyARNs) > 0 {
		obj["attach_policy_arns"] = toArrayInterface(in.AttachPolicyARNs)
	}
	if in.AttachPolicy != nil && len(in.AttachPolicy) > 0 {
		//log.Println("type:", reflect.TypeOf(in.AttachPolicy))
		var json2 = jsoniter.ConfigCompatibleWithStandardLibrary
		jsonStr, err := json2.Marshal(in.AttachPolicy)
		if err != nil {
			log.Println("attach policy marshal err:", err)
		}
		obj["attach_policy"] = string(jsonStr)
	}
	if len(in.AttachRoleARN) > 0 {
		obj["attach_role_arn"] = in.AttachRoleARN
	}
	if len(in.PermissionsBoundary) > 0 {
		obj["permissions_boundary"] = in.PermissionsBoundary
	}
	if in.Status != nil {
		obj["status"] = flattenIAMStatus(in.Status, []interface{}{})
	}
	if len(in.RoleName) > 0 {
		obj["role_name"] = in.RoleName
	}

	if in.Tags != nil && len(in.Tags) > 0 {
		obj["tags"] = toMapInterface(in.Tags)
	}
	return obj
}

func flattenIAMPodIdentityAssociations(inp []*IAMPodIdentityAssociation, p []interface{}) []interface{} {
	if inp == nil {
		return nil
	}
	out := make([]interface{}, len(inp))
	for i, in := range inp {
		obj := map[string]interface{}{}
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}
		if len(in.ServiceAccountName) > 0 {
			obj["service_account_name"] = in.ServiceAccountName
		}
		if len(in.Namespace) > 0 {
			obj["namespace"] = in.Namespace
		}
		if len(in.RoleARN) > 0 {
			obj["role_arn"] = in.RoleARN
		}
		if len(in.RoleName) > 0 {
			obj["role_name"] = in.RoleName
		}
		if len(in.PermissionPolicy) > 0 {
			var json2 = jsoniter.ConfigCompatibleWithStandardLibrary
			jsonStr, err := json2.Marshal(in.PermissionPolicy)
			if err != nil {
				log.Println("permission policy marshal err:", err)
			}
			//log.Println("jsonSTR:", jsonStr)
			obj["permission_policy"] = string(jsonStr)
		}
		if in.WellKnownPolicies != nil {
			v, ok := obj["well_known_policies"].([]interface{})
			if !ok {
				v = []interface{}{}
			}
			obj["well_known_policies"] = flattenIAMWellKnownPolicies(in.WellKnownPolicies, v)
		}
		if len(in.PermissionPolicyARNs) > 0 {
			obj["permission_policy_arns"] = toArrayInterface(in.PermissionPolicyARNs)
		}
		if len(in.PermissionsBoundaryARN) > 0 {
			obj["permissions_boundary_arn"] = in.PermissionsBoundaryARN
		}
		if len(in.Tags) > 0 {
			obj["tags"] = toMapInterface(in.Tags)
		}
		if in.CreateServiceAccount {
			obj["create_service_account"] = in.CreateServiceAccount
		}

		out[i] = obj
	}
	return out
}

func flattenIAMServiceAccounts(inp []*EKSClusterIAMServiceAccount, rawState cty.Value, p []interface{}) []interface{} {
	if inp == nil {
		return nil
	}
	var out []interface{}

	indexOf := func(item string, list []string) int {
		for i, v := range list {
			if v == item {
				return i
			}
		}
		return -1
	}
	findLocalOrder := func(rawState cty.Value) []string {
		var order []string
		if !rawState.IsNull() {
			for _, crdSpec := range rawState.AsValueSlice() {
				item := ""
				if saName, ok := crdSpec.AsValueMap()["name"]; ok {
					item += fmt.Sprintf("%s/", saName.AsString())
				}
				if saNamespace, ok := crdSpec.AsValueMap()["namespace"]; ok {
					item += saNamespace.AsString()
				}
				order = append(order, item)
			}
		}
		return order
	}
	findRemoteOrder := func(inp []*EKSClusterIAMServiceAccount) []string {
		var order []string
		for _, in := range inp {
			item := ""
			if in.Metadata != nil {
				item += fmt.Sprintf("%s/", in.Metadata.Name)
				item += in.Metadata.Namespace
			}
			order = append(order, item)
		}
		return order
	}
	findInRemoteOnly := func(local []string, remote []string, inp []*EKSClusterIAMServiceAccount) []*EKSClusterIAMServiceAccount {
		res := make([]*EKSClusterIAMServiceAccount, 0)
		for i, item := range remote {
			if indexOf(item, local) < 0 {
				res = append(res, inp[i])
			}
		}
		return res
	}

	saOrderInState := findLocalOrder(rawState)
	saOrderInRemote := findRemoteOrder(inp)
	saInRemoteOnly := findInRemoteOnly(saOrderInState, saOrderInRemote, inp)

	for _, key := range saOrderInState {
		if i := indexOf(key, saOrderInRemote); i >= 0 {
			in := inp[i]
			obj := flattenSingleIAMServiceAccount(in)
			out = append(out, obj)
		} else if len(saInRemoteOnly) > 0 {
			removedEntry := saInRemoteOnly[0]
			saInRemoteOnly = saInRemoteOnly[1:]
			obj := flattenSingleIAMServiceAccount(removedEntry)
			out = append(out, obj)
		}
	}
	for _, in := range saInRemoteOnly {
		obj := flattenSingleIAMServiceAccount(in)
		out = append(out, obj)
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

	if len(in.Id) > 0 {
		obj["id"] = in.Id
	}

	v, ok := obj["statement"].([]interface{})
	if !ok {
		v = []interface{}{}
	}
	obj["statement"] = flattenStatement(in.Statement, v)

	return []interface{}{obj}
}

func flattenStatement(in []InlineStatement, p []interface{}) []interface{} {

	if in == nil {
		return nil
	}

	out := make([]interface{}, len(in))

	for i, in := range in {
		obj := map[string]interface{}{}

		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}
		if len(in.Effect) > 0 {
			obj["effect"] = in.Effect
		}
		if len(in.Sid) > 0 {
			obj["sid"] = in.Sid
		}
		if in.Action != nil && len(in.Action.([]interface{})) > 0 {
			obj["action"] = in.Action
		}
		if in.NotAction != nil && len(in.NotAction.([]interface{})) > 0 {
			obj["not_action"] = in.NotAction
		}
		if len(in.Resource.(string)) > 0 {
			obj["resource"] = in.Resource.(string)
		}
		if in.NotResource != nil && len(in.NotResource.([]interface{})) > 0 {
			obj["not_resource"] = in.NotResource
		}

		if len(in.Condition) > 0 {
			var json2 = jsoniter.ConfigCompatibleWithStandardLibrary
			jsonStr, err := json2.Marshal(in.Condition)
			if err != nil {
				log.Println("attach policy marshal err:", err)
			}
			obj["condition"] = string(jsonStr)

		}
		if len(in.Principal) > 0 {
			var json2 = jsoniter.ConfigCompatibleWithStandardLibrary
			jsonStr, err := json2.Marshal(in.Principal)
			if err != nil {
				log.Println("attach policy marshal err:", err)
			}
			obj["principal"] = string(jsonStr)

		}
		if len(in.NotPrincipal) > 0 {
			var json2 = jsoniter.ConfigCompatibleWithStandardLibrary
			jsonStr, err := json2.Marshal(in.NotPrincipal)
			if err != nil {
				log.Println("attach policy marshal err:", err)
			}
			obj["not_principal"] = string(jsonStr)

		}
		out[i] = obj
	}
	return out
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
	out := make([]interface{}, 0)
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
			out = append(out, obj)
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
			out = append(out, obj)
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
	uniqueKeys := make(map[string]bool)
	res := make([]string, len(p))
	for i := 0; i < len(p); i++ {
		if p[i] != nil {
			if obj, ok := p[i].(map[string]interface{}); ok {
				if x := extractValue(obj, "name"); x != "" && !uniqueKeys[x] {
					uniqueKeys[x] = true
					res = append(res, x)
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

func flattenEKSClusterAddons(inp []*Addon, rawState cty.Value, p []interface{}) ([]interface{}, error) {
	if inp == nil {
		return nil, fmt.Errorf("emptyinput flatten addons")
	}

	isPolicyV2 := func(rawState cty.Value, name string) bool {
		if !rawState.IsNull() {
			for _, addon := range rawState.AsValueSlice() {
				if addonName, ok := addon.AsValueMap()["name"]; ok {
					if attachPolicyVersion, ok := addon.AsValueMap()["attach_policy_v2"]; ok {
						// log.Println("isPolicyV2 check:", addonName.AsString(), name, attachPolicyVersion.AsString())
						if addonName.AsString() == name && attachPolicyVersion.AsString() != "" {
							return true
						}
					}
				}
			}
		}
		return false
	}
	isPolicyV1 := func(rawState cty.Value, name string) bool {
		if !rawState.IsNull() {
			for _, addon := range rawState.AsValueSlice() {
				if addonName, ok := addon.AsValueMap()["name"]; ok {
					if attachPolicyVersion, ok := addon.AsValueMap()["attach_policy"]; ok {
						//log.Println("isPolicyV2 check:", addonName.AsString(), name, attachPolicyVersion.AsString())
						if addonName.AsString() == name && len(attachPolicyVersion.AsValueSlice()) != 0 {
							return true
						}
					}
				}
			}
		}
		return false
	}

	isSetInState := func(rawState cty.Value, name string) bool {
		if !rawState.IsNull() {
			for _, addon := range rawState.AsValueSlice() {
				if addonName, ok := addon.AsValueMap()["name"]; ok {
					if addonName.AsString() == name {
						return true
					}
				}
			}
		}
		return false
	}

	filterAddon := make([]*Addon, 0)
	for _, addon := range inp {
		if isSetInState(rawState, addon.Name) {
			filterAddon = append(filterAddon, addon)
		}
	}

	out := make([]interface{}, len(filterAddon))
	for i, in := range filterAddon {

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
			if !isPolicyV2(rawState, in.Name) && isPolicyV1(rawState, in.Name) {
				v1, ok := obj["attach_policy"].([]interface{})
				if !ok {
					v1 = []interface{}{}
				}
				obj["attach_policy"] = flattenAttachPolicy(in.AttachPolicy, v1)
			} else {
				var json2 = jsoniter.ConfigCompatibleWithStandardLibrary
				jsonStr, err := json2.Marshal(in.AttachPolicy)
				if err != nil {
					log.Println("attach policy marshal err:", err)
				}
				//log.Println("jsonSTR:", jsonStr)
				obj["attach_policy_v2"] = string(jsonStr)

			}
		}

		if len(in.PermissionsBoundary) > 0 {
			obj["permissions_boundary"] = in.PermissionsBoundary
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

		if in.PodIdentityAssociations != nil {
			v, ok := obj["pod_identity_associations"].([]interface{})
			if !ok {
				v = []interface{}{}
			}
			obj["pod_identity_associations"] = flattenIAMPodIdentityAssociations(in.PodIdentityAssociations, v)
		}

		if in.UseDefaultPodIdentityAssociations {
			obj["use_default_pod_identity_associations"] = in.UseDefaultPodIdentityAssociations
		}

		out[i] = &obj
	}

	log.Println("Flatten eks addons", out)
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
func flattenEKSClusterNodeGroups(inp []*NodeGroup, rawState cty.Value, p []interface{}) []interface{} {
	if inp == nil {
		return nil
	}
	indexOf := func(item string, list []string) int {
		for i, v := range list {
			if v == item {
				return i
			}
		}
		return -1
	}
	findLocalOrder := func(rawState cty.Value) []string {
		var order []string
		if !rawState.IsNull() {
			for _, val := range rawState.AsValueSlice() {
				order = append(order, val.AsString())
			}
		}
		return order
	}
	flattenListOfString := func(inp []string, rawState cty.Value) []interface{} {
		out := make([]interface{}, 0)
		localStateOrder := findLocalOrder(rawState)
		remoteStateOrder := inp
		remoteOnlyOrder := make([]string, 0)
		for _, elem := range remoteStateOrder {
			if indexOf(elem, localStateOrder) < 0 {
				remoteOnlyOrder = append(remoteOnlyOrder, elem)
			}
		}
		for _, elem := range localStateOrder {
			if i := indexOf(elem, remoteStateOrder); i >= 0 {
				out = append(out, elem)
			} else if len(remoteOnlyOrder) > 0 {
				out = append(out, remoteOnlyOrder[0])
				remoteOnlyOrder = remoteOnlyOrder[1:]
			}
		}
		for _, elem := range remoteOnlyOrder {
			out = append(out, elem)
		}
		return out
	}

	out := make([]interface{}, len(inp))
	for i, in := range inp {
		obj := map[string]interface{}{}
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}
		if in == nil {
			out[i] = &obj
			continue
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
			var nRawState cty.Value
			if !rawState.IsNull() && i < len(rawState.AsValueSlice()) {
				nRawState = rawState.AsValueSlice()[i].GetAttr("availability_zones")
			}
			obj["availability_zones"] = flattenListOfString(in.AvailabilityZones, nRawState)
		}
		if len(in.Subnets) > 0 {
			var nRawState cty.Value
			if !rawState.IsNull() && i < len(rawState.AsValueSlice()) {
				nRawState = rawState.AsValueSlice()[i].GetAttr("subnets")
			}
			obj["subnets"] = flattenListOfString(in.Subnets, nRawState)
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
			var nRawState cty.Value
			if !rawState.IsNull() && i < len(rawState.AsValueSlice()) {
				nRawState = rawState.AsValueSlice()[i].GetAttr("iam")
			}
			obj["iam"] = flattenNodeGroupIAM(in.IAM, nRawState, v)
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
func flattenNodeGroupIAM(in *NodeGroupIAM, rawState cty.Value, p []interface{}) []interface{} {
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}
	if in == nil {
		return []interface{}{obj}
	}

	isPolicyV2 := func(rawState cty.Value) bool {
		if !rawState.IsNull() && len(rawState.AsValueSlice()) > 0 {
			iamSpec := rawState.AsValueSlice()[0]
			if attachPolicyV2, ok := iamSpec.AsValueMap()["attach_policy_v2"]; ok {
				return attachPolicyV2.AsString() != ""
			}
		}
		return false
	}

	//@@@TODO Store inline document object as terraform input correctly
	if in.AttachPolicy != nil {
		if !isPolicyV2(rawState) {
			v1, ok := obj["attach_policy"].([]interface{})
			if !ok {
				v1 = []interface{}{}
			}
			obj["attach_policy"] = flattenAttachPolicy(in.AttachPolicy, v1)
		} else {
			var json2 = jsoniter.ConfigCompatibleWithStandardLibrary
			jsonStr, err := json2.Marshal(in.AttachPolicy)
			if err != nil {
				log.Println("attach policy marshal err:", err)
			}
			//log.Println("jsonSTR:", jsonStr)
			obj["attach_policy_v2"] = string(jsonStr)
			//log.Println("jsonSTR: for v2 nodegroup", obj)
		}

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
func flattenEKSClusterManagedNodeGroups(inp []*ManagedNodeGroup, rawState cty.Value, p []interface{}) ([]interface{}, error) {
	if inp == nil {
		return nil, fmt.Errorf("empty input for managedNodeGroup")
	}

	indexOf := func(item string, list []string) int {
		for i, v := range list {
			if v == item {
				return i
			}
		}
		return -1
	}
	findLocalOrder := func(rawState cty.Value) []string {
		var order []string
		if !rawState.IsNull() {
			for _, val := range rawState.AsValueSlice() {
				order = append(order, val.AsString())
			}
		}
		return order
	}
	flattenListOfString := func(inp []string, rawState cty.Value) []interface{} {
		out := make([]interface{}, 0)
		localStateOrder := findLocalOrder(rawState)
		remoteStateOrder := inp
		remoteOnlyOrder := make([]string, 0)
		for _, elem := range remoteStateOrder {
			if indexOf(elem, localStateOrder) < 0 {
				remoteOnlyOrder = append(remoteOnlyOrder, elem)
			}
		}
		for _, elem := range localStateOrder {
			if i := indexOf(elem, remoteStateOrder); i >= 0 {
				out = append(out, elem)
			} else if len(remoteOnlyOrder) > 0 {
				out = append(out, remoteOnlyOrder[0])
				remoteOnlyOrder = remoteOnlyOrder[1:]
			}
		}
		for _, elem := range remoteOnlyOrder {
			out = append(out, elem)
		}
		return out
	}

	out := make([]interface{}, len(inp))
	for i, in := range inp {
		obj := map[string]interface{}{}
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}
		if in == nil {
			out[i] = &obj
			continue
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
			var nRawState cty.Value
			if !rawState.IsNull() && i < len(rawState.AsValueSlice()) {
				nRawState = rawState.AsValueSlice()[i].GetAttr("availability_zones")
			}
			obj["availability_zones"] = flattenListOfString(in.AvailabilityZones, nRawState)
		}
		if len(in.Subnets) > 0 {
			var nRawState cty.Value
			if !rawState.IsNull() && i < len(rawState.AsValueSlice()) {
				nRawState = rawState.AsValueSlice()[i].GetAttr("subnets")
			}
			obj["subnets"] = flattenListOfString(in.Subnets, nRawState)
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
			var nRawState cty.Value
			if !rawState.IsNull() && i < len(rawState.AsValueSlice()) {
				nRawState = rawState.AsValueSlice()[i].GetAttr("iam")
			}
			obj["iam"] = flattenNodeGroupIAM(in.IAM, nRawState, v)
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
		out[i] = &obj
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
		if in.LogRetentionInDays != 0 {
			obj["log_retention_in_days"] = in.LogRetentionInDays
		}
	}
	return []interface{}{obj}
}

func flattenEKSClusterAddonsConfig(in *EKSAddonsConfig, p []interface{}) []interface{} {
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}
	if in == nil {
		return []interface{}{obj}
	}
	obj["auto_apply_pod_identity_associations"] = in.AutoApplyPodIdentityAssociations
	obj["disable_ebs_csi_driver"] = in.DisableEBSCSIDriver

	return []interface{}{obj}

}

func flattenAutoModeConfig(in *EKSAutoModeConfig, p []interface{}) []interface{} {
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}
	if in == nil {
		return []interface{}{obj}
	}
	obj["enabled"] = in.Enabled

	if len(in.NodeRoleARN) > 0 {
		obj["node_role_arn"] = in.NodeRoleARN
	}

	if len(in.NodePools) > 0 {
		obj["node_pools"] = in.NodePools
	}
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

	if in.EncryptExistingSecrets != nil && *in.EncryptExistingSecrets {
		obj["encrypt_existing_secrets"] = true
	}

	if in.EncryptExistingSecrets != nil && !*in.EncryptExistingSecrets {
		obj["encrypt_existing_secrets"] = false
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

	if in.Arns != nil && len(in.Arns) > 0 {
		v, ok := obj["arns"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["arns"] = flattenArnFields(in.Arns, v)
	}

	if in.Accounts != nil && len(in.Accounts) > 0 {
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

func resourceEKSClusterCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Println("create EKS cluster resource")
	return resourceEKSClusterUpsert(ctx, d, m)
}

func resourceEKSClusterRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Println("READ eks cluster")
	rawState := d.GetRawState()
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
		log.Print("Cluster project name is invalid")
		return diag.Errorf("Cluster project name is invalid")
	}
	c, err := cluster.GetCluster(clusterName, projectID, uaDef)
	if err != nil {
		log.Printf("error in get cluster %s", err.Error())
		if strings.Contains(err.Error(), "not found") {
			log.Println("Resource Read ", "error", err)
			d.SetId("")
			return diag.FromErr(fmt.Errorf("resource read failed, cluster not found. Error: %s", err.Error()))
		}
		return diag.FromErr(err)
	}

	cse := c.Settings[clusterSharingExtKey]
	tflog.Info(ctx, "Got cluster from backend", map[string]any{clusterSharingExtKey: cse})

	logger := glogger.GetLogger()
	rctlCfg := config.GetConfig()
	clusterSpecYaml, err := clusterctl.GetClusterSpec(logger, rctlCfg, c.Name, projectID, uaDef)
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

	// If the cluster sharing is managed by separate resource then
	// don't consider sharing from `rafay_eks_cluster`. Both
	// should not be present simultaneously.
	if cse == "true" {
		clusterSpec.Spec.Sharing = nil
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
	c1, err := flattenEKSCluster(&clusterSpec, v, rawState.GetAttr("cluster"))
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
	c2, err := flattenEKSClusterConfig(&clusterConfigSpec, rawState.GetAttr("cluster_config"), v2)
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
	c, err := cluster.GetCluster(clusterName, projectID, uaDef)
	if err != nil {
		log.Printf("error in get cluster %s", err.Error())
		return diag.FromErr(err)
	}
	if c.ID != d.Id() {
		log.Printf("edge id has changed, state: %s, current: %s", d.Id(), c.ID)
		return diag.Errorf("remote and state id mismatch")
	}
	cse := c.Settings[clusterSharingExtKey]
	tflog.Error(ctx, "##### Fetched cluster", map[string]any{clusterSharingExtKey: cse})

	// Check if cse == true and `spec.sharing` specified. then
	// Error out here only before procedding. The next Upsert is
	// called by "Create" flow as well which is explicitly setting
	// cse to false if `spec.sharing` provided.
	if cse == "true" {
		if d.HasChange("cluster.0.spec.0.sharing") {
			_, new := d.GetChange("cluster.0.spec.0.sharing")
			if new != nil {
				return diag.Errorf("Cluster sharing is currently managed through the external 'rafay_cluster_sharing' resource. To prevent configuration conflicts, please remove the sharing settings from the 'rafay_eks_cluster' resource and manage sharing exclusively via the external resource.")
			}
		}
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

	errDel := cluster.DeleteCluster(clusterName, projectID, false, uaDef)
	if errDel != nil {
		log.Printf("delete cluster error %s", errDel.Error())
		return diag.FromErr(errDel)
	}

	ticker := time.NewTicker(time.Duration(60) * time.Second)
	defer ticker.Stop()

LOOP:
	for {
		select {
		case <-ctx.Done():
			log.Printf("Cluster Deletion for edgename: %s and projectname: %s got timeout out.", clusterName, projectName)
			return diag.FromErr(fmt.Errorf("cluster deletion for edgename: %s and projectname: %s got timeout out", clusterName, projectName))
		case <-ticker.C:
			check, errGet := cluster.GetCluster(clusterName, projectID, uaDef)
			if errGet != nil {
				log.Printf("error while getCluster %s, delete success", errGet.Error())
				break LOOP
			}
			if check == nil {
				break LOOP
			}
			log.Printf("Cluster Deletion is in progress for edgename: %s and projectname: %s. Waiting 60 seconds more for operation to complete.", clusterName, projectName)
		}
	}
	log.Printf("Cluster Deletion completes for edgename: %s and projectname: %s", clusterName, projectName)
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

	s, errGet := cluster.GetCluster(clusterName, projectID, uaDef)
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
