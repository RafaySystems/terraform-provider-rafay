package rafay

import (
	"fmt"
	"math/rand"
	"time"
)

// EKSNGInfoProvider interface provides node group information
type EKSNGInfoProvider interface {
	GetRegion() string
	GetVersion() string
	GetTags() map[string]string
	GetEnableFullAccessToEcr() bool
	GetEnableAsgAccess() bool
	GetEnableExternalDnsAccess() bool
	GetEnableAccessToAppmesh() bool
	GetEnableAccessForAlbIngressController() bool
	GetEnableEfs() bool
	GetInstanceRoleArn() string
	GetInstanceProfileArn() string
	GetInstanceRolePermissionsBoundary() string
	GetManaged() bool
	GetNodes() int64
	GetNodesMin() int64
	GetNodesMax() int64
	GetSshAccess() bool
	GetSshPublicKey() string
	GetNodegroupName() string
	GetNodeAmiFamily() string
	GetInstanceType() string
	GetNodeZones() []string
	GetNodeVolumeSize() int64
	GetNodeLabels() map[string]string
	GetNodePrivateNetworking() bool
	GetNodeSecurityGroups() []string
	GetNodeAmi() string
	GetMaxPodsPerNode() int64
	GetNodeVolumeType() string
	GetVolumeEncrypted() bool
	GetVolumeKmsKeyId() string
	GetInstanceTypes() []string
	GetMaxPrice() string
	GetOnDemandBaseCapacity() int64
	GetOnDemandPercentageAboveBaseCapacity() int64
	GetSpotInstancePools() int64
	GetSpotAllocationStrategy() string
	GetSpot() bool
	GetBootstrapCommands() []string
	GetNodeTags() map[string]string
	GetNodeSubnets() []string
	GetSubnetCidr() string
}

// GenerateNodeGroupName generates a random nodegroup name
func GenerateNodeGroupName() string {
	const nameLength = 8
	const components = "abcdef0123456789"
	var r = rand.New(rand.NewSource(time.Now().UnixNano()))
	name := make([]byte, nameLength)
	for i := 0; i < nameLength; i++ {
		name[i] = components[r.Intn(len(components))]
	}
	return fmt.Sprintf("ng-%s", string(name))
}

const (
	// EKSConfigKind represents the kind of the config file
	EKSConfigKind string = "ClusterConfig"

	// EKSConfigAPIVersion represents the current API version of the YAML file
	EKSConfigAPIVersion string = "rafay.io/v1alpha5"
)

const (
	// EKSDefaultVPCClusterEndpointsPublicAccess holds the default value for cfg.vpc.ClusterEndpoints.PublicAccess (Rafay Override)
	EKSDefaultVPCClusterEndpointsPublicAccess = "false"
	// EKSDefaultVPCClusterEndpointsPrivateAccess holds the default value for cfg.vpc.ClusterEndpoints.PrivateAccess (Rafay Override)
	EKSDefaultVPCClusterEndpointsPrivateAccess = "true"
	// EKSDefaultNodegroupBaseVolumeSize holds the default value for cfg.*nodegroups[].volumeSize (in GB) (EKSCTL Default as of now)
	EKSDefaultNodegroupBaseVolumeSize = 80
	// EKSDefaultNodegroupBaseInstanceType holds the default value for cfg.*nodegroups[].instanceType (EKSCTL default as of now)
	EKSDefaultNodegroupBaseInstanceType = "m5.xlarge"
	// EKSDefaultNodegroupBaseVolumeType holds the default value for cfg.*nodegroups[].volumeType (EKSCTL default as of now)
	EKSDefaultNodegroupBaseVolumeType = "gp3"
	// EKSDefaultNodegroupBaseAMIFamily holds the default value for cfg.*nodegroups[].amiFamily (EKSCTL default as of now)
	EKSDefaultNodegroupBaseAMIFamily = "AmazonLinux2"
	// EKSDefaultNodegroupCount holds the default value for cfg.*nodegroups[].desiredCapacity (EKSCTL Default as of now)
	EKSDefaultNodegroupCount = 2
)

//struct for eks cluster metadata (first part of the yaml file kind:cluster)
type EKSCluster struct {
	Kind     string              `yaml:"kind,omitempty"`
	Metadata *EKSClusterMetadata `yaml:"metadata,omitempty"`
	Spec     *EKSSpec            `yaml:"spec,omitempty"`
}

type EKSSpec struct {
	Type             string            `yaml:"type,omitempty"`
	Blueprint        string            `yaml:"blueprint,omitempty"`
	BlueprintVersion string            `yaml:"blueprint_version,omitempty"`
	CloudProvider    string            `yaml:"cloud_provider,omitempty"`
	CniProvider      string            `yaml:"cni_provider,omitempty"`
	ProxyConfig      map[string]string `yaml:"proxy_config,omitempty"`
}

type EKSClusterMetadata struct {
	Name    string            `yaml:"name,omitempty"`
	Project string            `yaml:"project,omitempty"`
	Labels  map[string]string `yaml:"labels,omitempty"`
}

//struct for eks cluster config sped (second part of the yaml file kind:clusterConfig)
type KubernetesNetworkConfig struct {
	IPFamily        string `yaml:"ipFamily,omitempty"`
	ServiceIPv4CIDR string `yaml:"serviceIPv4CIDR,omitempty"`
}

type EKSClusterConfig struct {
	APIVersion              string                    `yaml:"apiversion,omitempty"`
	Kind                    string                    `yaml:"kind,omitempty"`
	Metadata                *EKSClusterConfigMetadata `yaml:"metadata,omitempty"`
	KubernetesNetworkConfig *KubernetesNetworkConfig  `yaml:"kubernetes_network_config,omitempty"`
	IAM                     *EKSClusterIAM            `yaml:"iam,omitempty,omitempty"`
	IdentityProviders       []IdentityProvider        `yaml:"identity_providers,omitempty"`
	VPC                     *EKSClusterVPC            `yaml:"vpc,omitempty"`
	// +optional
	Addons []*Addon `yaml:"addons,omitempty"`
	// +optional
	PrivateCluster    *PrivateCluster       `yaml:"private_cluster,omitempty"`
	NodeGroups        []*NodeGroup          `yaml:"node_groups,omitempty"`
	ManagedNodeGroups []*ManagedNodeGroup   `yaml:"managed_nodegroups,omitempty"`
	FargateProfiles   []*FargateProfile     `yaml:"fargate_profiles,omitempty"`
	AvailabilityZones []string              `yaml:"availability_zones,omitempty"`
	CloudWatch        *EKSClusterCloudWatch `yaml:"cloud_watch,omitempty"`
	SecretsEncryption *SecretsEncryption    `yaml:"secrets_encryption,omitempty"`
	//do i need this? not in docs
	//Karpenter *Karpenter `yaml:"karpenter,omitempty"`
}

/*Took this struct and modified it to fit documentation
// EKSClusterConfig struct -> cfg
type EKSClusterConfig struct {
	APIVersion  string          `yaml:"apiVersion"`
	Kind        string          `yaml:"kind"`
	ClusterMeta *EKSClusterMeta `yaml:"metadata"`
	IAM         *EKSClusterIAM  `yaml:"iam,omitempty"`
	// +optional
	IdentityProviders []IdentityProvider     `yaml:"identityProviders,omitempty"`
	VPC               *EKSClusterVPC         `yaml:"vpc,omitempty"`
	NodeGroups        []*EKSNodeGroup        `yaml:"nodeGroups,omitempty"`
	ManagedNodeGroups []*EKSManagedNodeGroup `yaml:"managedNodeGroups,omitempty"`
	CloudWatch        *EKSClusterCloudWatch  `yaml:"cloudWatch,omitempty"`

	AvailabilityZones []string `yaml:"availabilityZones,omitempty"`
}
*/
type AWSPolicyInlineDocument map[string]interface{}

// EKSClusterMeta struct -> cfg.EKSClusterMeta
type EKSClusterConfigMetadata struct {
	Name        string            `yaml:"name,omitempty"`
	Region      string            `yaml:"region,omitempty"`
	Version     string            `yaml:"version,omitempty"`
	Tags        map[string]string `yaml:"tags,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty"`
}

// EKSClusterIAM struct -> cfg.IAM.ServiceAccounts
type EKSClusterIAMMeta struct {
	Name        string            `yaml:"name,omitempty"`
	Namespace   string            `yaml:"namespace,omitempty"`
	Labels      map[string]string `yaml:"labels,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty"`
}

/*
type EKSClusterIAMServiceAccount struct {
	EKSClusterIAMMeta   `yaml:"metadata,omitempty"`
	AttachPolicyARNs    []string                `yaml:"attachPolicyARNs,omitempty"`
	AttachPolicy        AWSPolicyInlineDocument `yaml:"attachPolicy,omitempty"`
	PermissionsBoundary string                  `yaml:"permissionsBoundary,omitempty"`
	RoleOnly            *bool                   `yaml:"roleOnly,omitempty"`
	Tags                map[string]string       `yaml:"tags,omitempty"`
	// RoleName string `yaml:"roleName,omitempty"`
}
*/
type IdentityProvider struct {
	// Valid variants are:
	// `"oidc"`: OIDC identity provider
	// +required
	type_ string `yaml:"type,omitempty"` //nolint
	//Inner IdentityProviderInterface
}

// EKSClusterIAM struct -> cfg.IAM
type EKSClusterIAM struct {
	// +optional
	ServiceRoleARN string `yaml:"serviceRoservice_role_arnleARN,omitempty"`

	// permissions boundary for all identity-based entities created by eksctl.
	// See [AWS Permission Boundary](https://docs.aws.amazon.com/IAM/latest/UserGuide/access_policies_boundaries.html)
	// +optional
	ServiceRolePermissionsBoundary string `yaml:"service_role_permission_boundary,omitempty"`

	// role used by pods to access AWS APIs. This role is added to the Kubernetes RBAC for authorization.
	// See [Pod Execution Role](https://docs.aws.amazon.com/eks/latest/userguide/pod-execution-role.html)
	// +optional
	FargatePodExecutionRoleARN string `yaml:"fargate_pod_execution_role_arn,omitempty"`

	// permissions boundary for the fargate pod execution role`. See [EKS Fargate Support](/usage/fargate-support/)
	// +optional
	FargatePodExecutionRolePermissionsBoundary string `yaml:"fargate_pod_execution_role_permissions_boundary,omitempty"`

	// enables the IAM OIDC provider as well as IRSA for the Amazon CNI plugin
	// +optional
	WithOIDC *bool `yaml:"with_oidc,omitempty"`

	// service accounts to create in the cluster.
	// See [IAM Service Accounts](/iamserviceaccounts/#usage-with-config-files)
	// +optional
	ServiceAccounts []*EKSClusterIAMServiceAccount `yaml:"service_accounts,omitempty"`

	// VPCResourceControllerPolicy attaches the IAM policy
	// necessary to run the VPC controller in the control plane
	// Defaults to `true`
	VPCResourceControllerPolicy *bool `yaml:"vpc_resource_controller_policy,omitempty"`
}

// ClusterIAMServiceAccount holds an IAM service account metadata and configuration
type EKSClusterIAMServiceAccount struct {
	//KSClusterIAMMeta `yaml:"metadata,omitempty"`
	Name        string            `yaml:"name,omitempty"`
	Namespace   string            `yaml:"namespace,omitempty"`
	Labels      map[string]string `yaml:"labels,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty"`

	// list of ARNs of the IAM policies to attach
	// +optional
	AttachPolicyARNs []string `yaml:"attach_policy_arns,omitempty"`

	WellKnownPolicies WellKnownPolicies `yaml:"well_known_policies,omitempty"`

	// AttachPolicy holds a policy document to attach to this service account
	// +optional
	//AttachPolicy map[string]string `yaml:"attachPolicy,omitempty"`
	AttachPolicy InlineDocument `yaml:"attach_policy,omitempty"`

	// ARN of the role to attach to the service account
	AttachRoleARN string `yaml:"attach_role_arn,omitempty"`

	// ARN of the permissions boundary to associate with the service account
	// +optional
	PermissionsBoundary string `yaml:"permissions_boundary,omitempty"`

	// +optional
	Status *ClusterIAMServiceAccountStatus `yaml:"status,omitempty"`

	// Specific role name instead of the Cloudformation-generated role name
	// +optional
	RoleName string `yaml:"role_name,omitempty"`

	// Specify if only the IAM Service Account role should be created without creating/annotating the service account
	// +optional
	RoleOnly *bool `yaml:"role_only,omitempty"`

	// AWS tags for the service account
	// +optional
	Tags map[string]string `yaml:"tags,omitempty"`
}

type WellKnownPolicies struct {
	// ImageBuilder allows for full ECR (Elastic Container Registry) access.
	ImageBuilder *bool `yaml:"image_builder,inline,omitempty"`
	// AutoScaler adds policies for cluster-autoscaler. See [autoscaler AWS
	// docs](https://docs.aws.amazon.com/eks/latest/userguide/cluster-autoscaler.html).
	AutoScaler *bool `yaml:"auto_scaler,inline,omitempty"`
	// AWSLoadBalancerController adds policies for using the
	// aws-load-balancer-controller. See [Load Balancer
	// docs](https://docs.aws.amazon.com/eks/latest/userguide/aws-load-balancer-controller.html).
	AWSLoadBalancerController *bool `yaml:"aws_load_balancer_controller,inline,omitempty"`
	// ExternalDNS adds external-dns policies for Amazon Route 53.
	// See [external-dns
	// docs](https://github.com/kubernetes-sigs/external-dns/blob/master/docs/tutorials/aws.md).
	ExternalDNS *bool `yaml:"external_dns,inline,omitempty"`
	// CertManager adds cert-manager policies. See [cert-manager
	// docs](https://cert-manager.io/docs/configuration/acme/dns01/route53).
	CertManager *bool `yaml:"cert_manager,inline,omitempty"`
	// EBSCSIController adds policies for using the
	// ebs-csi-controller. See [aws-ebs-csi-driver
	// docs](https://github.com/kubernetes-sigs/aws-ebs-csi-driver#set-up-driver-permission).
	EBSCSIController *bool `yaml:"ebs_csi_controller,inline,omitempty"`
	// EFSCSIController adds policies for using the
	// efs-csi-controller. See [aws-efs-csi-driver
	// docs](https://aws.amazon.com/blogs/containers/introducing-efs-csi-dynamic-provisioning).
	EFSCSIController *bool `yaml:"efs_csi_controller,inline,omitempty"`
}

//type InlineDocument map[string]interface{}
type InlineDocument struct {
	Version   string          `yaml:"version,omitempty"`
	Statement InlineStatement `yaml:"statement,omitempty"`
}
type InlineStatement struct {
	Effect   string   `yaml:"effect,omitempty"`
	Action   []string `yaml:"action,omitempty"`
	Resource string   `yaml:"resource,omitempty"`
}

type ClusterIAMServiceAccountStatus struct {
	// +optional
	RoleARN string `yaml:"role_arn,omitempty"`
}
type AZSubnetMapping map[string]AZSubnetSpec
type TFAZSubnetMapping map[string]TFAZSubnetSpec
type (
	// ClusterVPC holds global subnet and all child subnets
	EKSClusterVPC struct {
		// global CIDR and VPC ID
		// +optional
		//Network
		// +optional
		ID string `yaml:"id,omitempty"`
		// +optional
		//can i just make this a string?
		CIDR string `yaml:"cidr,omitempty"`
		//CIDR *ipnet.IPNet `yaml:"cidr,omitempty"`
		// +optional
		IPv6Cidr string `yaml:"ipv6_cidr,omitempty"`
		// +optional
		IPv6Pool string `yaml:"ipv6_pool,omitempty"`
		// SecurityGroup (aka the ControlPlaneSecurityGroup) for communication between control plane and nodes
		// +optional
		SecurityGroup string `yaml:"security_group,omitempty"`
		// Subnets are keyed by AZ for convenience.
		// See [this example](/examples/reusing-iam-and-vpc/)
		// as well as [using existing
		// VPCs](/usage/vpc-networking/#use-existing-vpc-other-custom-configuration).
		// +optional
		Subnets *ClusterSubnets `yaml:"subnets,omitempty"`
		// for additional CIDR associations, e.g. a CIDR for
		// private subnets or any ad-hoc subnets
		// +optional
		ExtraCIDRs []string `yaml:"extra_cidrs,omitempty"`
		// for additional IPv6 CIDR associations, e.g. a CIDR for
		// private subnets or any ad-hoc subnets
		// +optional
		ExtraIPv6CIDRs []string `yaml:"extra_ipv6_cidrs,omitempty"`
		// for pre-defined shared node SG
		SharedNodeSecurityGroup string `yaml:"shared_node_security_group,omitempty"`
		// Automatically add security group rules to and from the default
		// cluster security group and the shared node security group.
		// This allows unmanaged nodes to communicate with the control plane
		// and managed nodes.
		// This option cannot be disabled when using eksctl created security groups.
		// Defaults to `true`
		// +optional
		ManageSharedNodeSecurityGroupRules *bool `yaml:"manage_shared_node_security_group_rules,omitempty"`
		// AutoAllocateIPV6 requests an IPv6 CIDR block with /56 prefix for the VPC
		// +optional
		AutoAllocateIPv6 *bool `yaml:"auto_allocate_ipv6,omitempty"`
		// +optional
		NAT *ClusterNAT `yaml:"nat,omitempty"`
		// See [managing access to API](/usage/vpc-networking/#managing-access-to-the-kubernetes-api-server-endpoints)
		// +optional
		ClusterEndpoints *ClusterEndpoints `yaml:"cluster_endpoints,omitempty"`
		// PublicAccessCIDRs are which CIDR blocks to allow access to public
		// k8s API endpoint
		// +optional
		PublicAccessCIDRs []string `yaml:"public_access_cidrs,omitempty"`
	}
	// ClusterSubnets holds private and public subnets
	TFClusterSubnets struct {
		Private TFAZSubnetMapping `yaml:"private,omitempty"`
		Public  TFAZSubnetMapping `yaml:"public,omitempty"`
		//Private AZSubnetMapping `yaml:"private,omitempty"`
		//Public  AZSubnetMapping `yaml:"public,omitempty"`
	}
	ClusterSubnets struct {
		Private AZSubnetMapping `yaml:"private,omitempty"`
		Public  AZSubnetMapping `yaml:"public,omitempty"`
	}
	// SubnetTopology can be SubnetTopologyPrivate or SubnetTopologyPublic
	SubnetTopology string
	TFAZSubnetSpec struct {
		Name string `yaml:"name,omitempty"`
		// +optional
		ID string `yaml:"id,omitempty"`
		// AZ can be omitted if the key is an AZ
		// +optional
		AZ string `yaml:"az,omitempty"`
		// +optional
		//can i just make this a string?
		CIDR string `yaml:"cidr,omitempty"`
		//CIDR *ipnet.IPNet `yaml:"cidr,omitempty"`
	}
	AZSubnetSpec struct {
		// +optional
		ID string `yaml:"id,omitempty"`
		// AZ can be omitted if the key is an AZ
		// +optional
		AZ string `yaml:"az,omitempty"`
		// +optional
		//can i just make this a string?
		CIDR string `yaml:"cidr,omitempty"`
		//CIDR *ipnet.IPNet `yaml:"cidr,omitempty"`
	}
	// Network holds ID and CIDR
	Network struct {
		// +optional
		ID string `yaml:"id,omitempty"`
		// +optional
		//can i just make this a string?
		CIDR string `yaml:"cidr,omitempty"`
		//CIDR *ipnet.IPNet `yaml:"cidr,omitempty"`
		// +optional
		IPv6Cidr string `yaml:"ipv6Cidr,omitempty"`
		// +optional
		IPv6Pool string `yaml:"ipv6Pool,omitempty"`
	}
	// ClusterNAT NAT config
	ClusterNAT struct {
		// Valid variants are `ClusterNAT` constants
		Gateway string `yaml:"gateway,omitempty"`
	}

	// ClusterEndpoints holds cluster api server endpoint access information
	ClusterEndpoints struct {
		PrivateAccess *bool `yaml:"privateAccess,omitempty"`
		PublicAccess  *bool `yaml:"publicAccess,omitempty"`
	}
)
type Addon struct {
	// +required
	Name string `yaml:"name,omitempty"`
	// +optional
	Version string `yaml:"version,omitempty"`
	// +optional
	ServiceAccountRoleARN string `yaml:"service_account_role_arn,omitempty"`
	// list of ARNs of the IAM policies to attach
	// +optional
	AttachPolicyARNs []string `yaml:"attach_policy_arns,omitempty"`
	// AttachPolicy holds a policy document to attach
	// +optional
	AttachPolicy InlineDocument `yaml:"attach_policy,omitempty"`
	// ARN of the permissions' boundary to associate
	// +optional
	PermissionsBoundary string `yaml:"permissions_boundary,omitempty"`
	// WellKnownPolicies for attaching common IAM policies
	//WellKnown Policies not in documentation for addon? (same field as IAM wellknow-policies)
	WellKnownPolicies WellKnownPolicies `yaml:"well_known_policies,omitempty"`
	// The metadata to apply to the cluster to assist with categorization and organization.
	// Each tag consists of a key and an optional value, both of which you define.
	// +optional
	Tags map[string]string `yaml:"tags,omitempty"`
	// Force applies the add-on to overwrite an existing add-on
	Force *bool `yaml:"-"`
}

// PrivateCluster defines the configuration for a fully-private cluster
type PrivateCluster struct {

	// Enabled enables creation of a fully-private cluster
	Enabled *bool `yaml:"enabled"`

	// SkipEndpointCreation skips the creation process for endpoints completely. This is only used in case of an already
	// provided VPC and if the user decided to set it to true.
	SkipEndpointCreation *bool `yaml:"skip_endpoint_creation"`

	// AdditionalEndpointServices specifies additional endpoint services that
	// must be enabled for private access.
	// Valid entries are `AdditionalEndpointServices` constants
	AdditionalEndpointServices []string `yaml:"additional_endpoint_services,omitempty"`
}
type NodeGroup struct {
	// +required
	Name string `yaml:"name"`

	// Valid variants are `NodeAMIFamily` constants
	// +optional
	AMIFamily string `yaml:"ami_family,omitempty"`
	// +optional
	InstanceType string `yaml:"instance_type,omitempty"`
	// Limit [nodes to specific
	// AZs](/usage/autoscaling/#zone-aware-auto-scaling)
	// +optional
	AvailabilityZones []string `yaml:"avalability_zones,omitempty"`
	// Limit nodes to specific subnets
	// +optional
	Subnets []string `yaml:"subnets,omitempty"`

	// +optional
	InstancePrefix string `yaml:"instance_prefix,omitempty"`
	// +optional
	InstanceName string `yaml:"instance_name,omitempty"`

	// +optional
	//ScalingConfig
	// +optional
	DesiredCapacity *int `yaml:"desired_capacity,omitempty"`
	// +optional
	MinSize *int `yaml:"min_size,omitempty"`
	// +optional
	MaxSize *int `yaml:"max_size,omitempty"`

	// +optional
	// VolumeSize gigabytes
	// Defaults to `80`
	VolumeSize *int `yaml:"volume_size,omitempty"`
	// +optional
	// SSH configures ssh access for this nodegroup
	SSH *NodeGroupSSH `yaml:"ssh,omitempty"`
	// +optional
	Labels map[string]string `yaml:"labels,omitempty"`
	// Enable [private
	// networking](/usage/vpc-networking/#use-private-subnets-for-initial-nodegroup)
	// for nodegroup
	// +optional
	PrivateNetworking *bool `yaml:"private_networking"`
	// Applied to the Autoscaling Group and to the EC2 instances (unmanaged),
	// Applied to the EKS Nodegroup resource and to the EC2 instances (managed)
	// +optional
	Tags map[string]string `yaml:"tags,omitempty"`
	// +optional
	IAM *NodeGroupIAM `yaml:"iam,omitempty"`

	// Specify [custom AMIs](/usage/custom-ami-support/), `auto-ssm`, `auto`, or `static`
	// +optional
	AMI string `yaml:"ami,omitempty"`

	// +optional
	SecurityGroups *NodeGroupSGs `yaml:"security_groups,omitempty"`

	// +optional
	MaxPodsPerNode int `yaml:"max_pods_per_node,omitempty"`

	// See [relevant AWS
	// docs](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-attribute-updatepolicy.html#cfn-attributes-updatepolicy-rollingupdate-suspendprocesses)
	// +optional
	ASGSuspendProcesses []string `yaml:"asg_suspend_processes,omitempty"`

	// EBSOptimized enables [EBS
	// optimization](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ebs-optimized.html)
	// +optional
	EBSOptimized *bool `yaml:"ebs_optimized,omitempty"`

	// Valid variants are `VolumeType` constants
	// +optional
	VolumeType string `yaml:"volume_type,omitempty"`
	// +optional
	VolumeName string `yaml:"volume_name,omitempty"`
	// +optional
	VolumeEncrypted *bool `yaml:"volume_encrypted,omitempty"`
	// +optional
	VolumeKmsKeyID string `yaml:"volume_kms_key_id,omitempty"`
	// +optional
	VolumeIOPS *int `yaml:"volume_iops,omitempty"`
	// +optional
	VolumeThroughput *int `yaml:"volume_throughput,omitempty"`

	// PreBootstrapCommands are executed before bootstrapping instances to the
	// cluster
	// +optional
	PreBootstrapCommands []string `yaml:"pre_bootstrap_commands,omitempty"`

	// Override `eksctl`'s bootstrapping script
	// +optional
	OverrideBootstrapCommand string `yaml:"override_bootstrap_command,omitempty"`

	// DisableIMDSv1 requires requests to the metadata service to use IMDSv2 tokens
	// Defaults to `false`
	// +optional
	DisableIMDSv1 *bool `yaml:"disable_imdsv1,omitempty"`

	// DisablePodIMDS blocks all IMDS requests from non host networking pods
	// Defaults to `false`
	// +optional
	DisablePodIMDS *bool `yaml:"disable_pods_imds,omitempty"`

	// Placement specifies the placement group in which nodes should
	// be spawned
	// +optional
	Placement *Placement `yaml:"placement,omitempty"`

	// EFAEnabled creates the maximum allowed number of EFA-enabled network
	// cards on nodes in this group.
	// +optional
	EFAEnabled *bool `yaml:"efa_enabled,omitempty"`

	// InstanceSelector specifies options for EC2 instance selector
	InstanceSelector *InstanceSelector `yaml:"instance_selector,omitempty"`

	// Internal fields
	// Some AMIs (bottlerocket) have a separate volume for the OS
	AdditionalEncryptedVolume string `yaml:"-"`

	// Bottlerocket specifies settings for Bottlerocket nodes
	// +optional
	Bottlerocket *NodeGroupBottlerocket `yaml:"bottle_rocket,omitempty"`

	// TODO remove this
	// This is a hack, will be removed shortly. When this is true for Ubuntu and
	// AL2 images a legacy bootstrapper will be used.
	CustomAMI *bool `yaml:"-"`

	// Enable EC2 detailed monitoring
	// +optional
	EnableDetailedMonitoring *bool `yaml:"enable_detailed_monitoring,omitempty"`
	// Rafay changes - start
	// Internal
	IsWavelengthZone *bool `yaml:"-"`
	// Rafay changes - end

	//+optional
	InstancesDistribution *NodeGroupInstancesDistribution `yaml:"instances_distribution,omitempty"`

	// +optional
	ASGMetricsCollection []MetricsCollection `yaml:"asg_metrics_collection,omitempty"`

	// CPUCredits configures [T3 Unlimited](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/burstable-performance-instances-unlimited-mode.html), valid only for T-type instances
	// +optional
	CPUCredits string `yaml:"cpu_credits,omitempty"`

	// Associate load balancers with auto scaling group
	// +optional
	ClassicLoadBalancerNames []string `yaml:"classic_load_balancer_names,omitempty"`

	// Associate target group with auto scaling group
	// +optional
	TargetGroupARNs []string `yaml:"target_group_arns,omitempty"`

	// Taints taints to apply to the nodegroup
	// +optional
	Taints []NodeGroupTaint `yaml:"taints,omitempty"`

	// UpdateConfig configures how to update NodeGroups.
	// +optional
	UpdateConfig *NodeGroupUpdateConfig `yaml:"update_config,omitempty"`

	// [Custom
	// address](/usage/vpc-networking/#custom-cluster-dns-address) used for DNS
	// lookups
	// +optional
	ClusterDNS string `yaml:"cluster_dns,omitempty"`

	// [Customize `kubelet` config](/usage/customizing-the-kubelet/)
	// +optional

	KubeletExtraConfig *KubeletExtraConfig `yaml:"kubelet_extra_config,omitempty"`

	// ContainerRuntime defines the runtime (CRI) to use for containers on the node
	// +optional
	ContainerRuntime string `yaml:"container_runtime,omitempty"`
}

//@@@ added new kubeletExtraConfig Struct
type KubeletExtraConfig struct {
	KubeReserved       map[string]string `yaml:"kube_reserved,omitempty"`
	KubeReservedCGroup string            `yaml:"kube_reserved_cgroup,omitempty"`
	SystemReserved     map[string]string `yaml:"system_reserved,omitempty"`
	EvictionHard       map[string]string `yaml:"eviction_hard,omitempty"`
	//@@@double check if this is correct for feature gates (or should it be preset with struct containing RoatateKubletServerCert)
	FeatureGates map[string]bool `yaml:"feature_gates,omitempty"`
}

// NodeGroupBase represents the base nodegroup config for self-managed and managed nodegroups
type NodeGroupBase struct {
	// +required
	Name string `yaml:"name"`

	// Valid variants are `NodeAMIFamily` constants
	// +optional
	AMIFamily string `yaml:"amiFamily,omitempty"`
	// +optional
	InstanceType string `yaml:"instanceType,omitempty"`
	// Limit [nodes to specific
	// AZs](/usage/autoscaling/#zone-aware-auto-scaling)
	// +optional
	AvailabilityZones []string `yaml:"availabilityZones,omitempty"`
	// Limit nodes to specific subnets
	// +optional
	Subnets []string `yaml:"subnets,omitempty"`

	// +optional
	InstancePrefix string `yaml:"instancePrefix,omitempty"`
	// +optional
	InstanceName string `yaml:"instanceName,omitempty"`

	// +optional
	ScalingConfig

	// +optional
	// VolumeSize gigabytes
	// Defaults to `80`
	VolumeSize *int `yaml:"volumeSize,omitempty"`
	// +optional
	// SSH configures ssh access for this nodegroup
	SSH *NodeGroupSSH `yaml:"ssh,omitempty"`
	// +optional
	Labels map[string]string `yaml:"labels,omitempty"`
	// Enable [private
	// networking](/usage/vpc-networking/#use-private-subnets-for-initial-nodegroup)
	// for nodegroup
	// +optional
	PrivateNetworking *bool `yaml:"privateNetworking"`
	// Applied to the Autoscaling Group and to the EC2 instances (unmanaged),
	// Applied to the EKS Nodegroup resource and to the EC2 instances (managed)
	// +optional
	Tags map[string]string `yaml:"tags,omitempty"`
	// +optional
	IAM *NodeGroupIAM `yaml:"iam,omitempty"`

	// Specify [custom AMIs](/usage/custom-ami-support/), `auto-ssm`, `auto`, or `static`
	// +optional
	AMI string `yaml:"ami,omitempty"`

	// +optional
	SecurityGroups *NodeGroupSGs `yaml:"securityGroups,omitempty"`

	// +optional
	MaxPodsPerNode *int `yaml:"maxPodsPerNode,omitempty"`

	// See [relevant AWS
	// docs](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-attribute-updatepolicy.html#cfn-attributes-updatepolicy-rollingupdate-suspendprocesses)
	// +optional
	ASGSuspendProcesses []string `yaml:"asgSuspendProcesses,omitempty"`

	// EBSOptimized enables [EBS
	// optimization](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ebs-optimized.html)
	// +optional
	EBSOptimized *bool `yaml:"ebsOptimized,omitempty"`

	// Valid variants are `VolumeType` constants
	// +optional
	VolumeType string `yaml:"volumeType,omitempty"`
	// +optional
	VolumeName string `yaml:"volumeName,omitempty"`
	// +optional
	VolumeEncrypted *bool `yaml:"volumeEncrypted,omitempty"`
	// +optional
	VolumeKmsKeyID string `yaml:"volumeKmsKeyID,omitempty"`
	// +optional
	VolumeIOPS *int `yaml:"volumeIOPS,omitempty"`
	// +optional
	VolumeThroughput *int `yaml:"volumeThroughput,omitempty"`

	// PreBootstrapCommands are executed before bootstrapping instances to the
	// cluster
	// +optional
	PreBootstrapCommands []string `yaml:"preBootstrapCommands,omitempty"`

	// Override `eksctl`'s bootstrapping script
	// +optional
	OverrideBootstrapCommand string `yaml:"overrideBootstrapCommand,omitempty"`

	// DisableIMDSv1 requires requests to the metadata service to use IMDSv2 tokens
	// Defaults to `false`
	// +optional
	DisableIMDSv1 *bool `yaml:"disableIMDSv1,omitempty"`

	// DisablePodIMDS blocks all IMDS requests from non host networking pods
	// Defaults to `false`
	// +optional
	DisablePodIMDS *bool `yaml:"disablePodIMDS,omitempty"`

	// Placement specifies the placement group in which nodes should
	// be spawned
	// +optional
	Placement *Placement `yaml:"placement,omitempty"`

	// EFAEnabled creates the maximum allowed number of EFA-enabled network
	// cards on nodes in this group.
	// +optional
	EFAEnabled *bool `yaml:"efaEnabled,omitempty"`

	// InstanceSelector specifies options for EC2 instance selector
	InstanceSelector *InstanceSelector `yaml:"instanceSelector,omitempty"`

	// Internal fields
	// Some AMIs (bottlerocket) have a separate volume for the OS
	AdditionalEncryptedVolume string `yaml:"-"`

	// Bottlerocket specifies settings for Bottlerocket nodes
	// +optional
	Bottlerocket *NodeGroupBottlerocket `yaml:"bottlerocket,omitempty"`

	// TODO remove this
	// This is a hack, will be removed shortly. When this is true for Ubuntu and
	// AL2 images a legacy bootstrapper will be used.
	CustomAMI *bool `yaml:"-"`

	// Enable EC2 detailed monitoring
	// +optional
	EnableDetailedMonitoring *bool `yaml:"enableDetailedMonitoring,omitempty"`
	// Rafay changes - start
	// Internal
	IsWavelengthZone *bool `yaml:"-"`
	// Rafay changes - end
}
type InstanceSelector struct {
	// VCPUs specifies the number of vCPUs
	VCPUs *int `yaml:"vcpus,omitempty"`
	// Memory specifies the memory
	// The unit defaults to GiB
	Memory string `yaml:"memory,omitempty"`
	// GPUs specifies the number of GPUs.
	// It can be set to 0 to select non-GPU instance types.
	GPUs *int `yaml:"gpus,omitempty"`
	// CPU Architecture of the EC2 instance type.
	// Valid variants are:
	// `"x86_64"`
	// `"amd64"`
	// `"arm64"`
	CPUArchitecture string `yaml:"cpu_architecture,omitempty"`
}
type Placement struct {
	GroupName string `yaml:"group,omitempty"`
}
type ScalingConfig struct {
	// +optional
	DesiredCapacity *int `yaml:"desired_capacity,omitempty"`
	// +optional
	MinSize *int `yaml:"min_size,omitempty"`
	// +optional
	MaxSize *int `yaml:"max_size,omitempty"`
}
type MetricsCollection struct {
	// +required
	Granularity string `yaml:"granularity"`
	// +optional
	Metrics []string `yaml:"metrics,omitempty"`
}
type taintsWrapper []NodeGroupTaint
type NodeGroupTaint struct {
	Key    string `yaml:"key,omitempty"`
	Value  string `yaml:"value,omitempty"`
	Effect string `yaml:"effect,omitempty"`
}
type (
	// NodeGroupSGs controls security groups for this nodegroup
	NodeGroupSGs struct {
		// AttachIDs attaches additional security groups to the nodegroup
		// +optional
		AttachIDs []string `yaml:"attach_ids,omitempty"`
		// WithShared attach the security group
		// shared among all nodegroups in the cluster
		// Defaults to `true`
		// +optional
		WithShared *bool `yaml:"with_shared"`
		// WithLocal attach a security group
		// local to this nodegroup
		// Not supported for managed nodegroups
		// Defaults to `true`
		// +optional
		WithLocal *bool `yaml:"with_local"`
	}
	// NodeGroupIAM holds all IAM attributes of a NodeGroup
	NodeGroupIAM struct {
		// AttachPolicy holds a policy document to attach
		// +optional
		AttachPolicy InlineDocument `yaml:"attach_policy,omitempty"`
		// list of ARNs of the IAM policies to attach
		// +optional
		AttachPolicyARNs []string `yaml:"attach_policy_arns,omitempty"`
		// +optional
		InstanceProfileARN string `yaml:"instance_profile_arn,omitempty"`
		// +optional
		InstanceRoleARN string `yaml:"instance_role_arn,omitempty"`
		// +optional
		InstanceRoleName string `yaml:"instance_role_name,omitempty"`
		// +optional
		InstanceRolePermissionsBoundary string `yaml:"instance_role_permission_boundary,omitempty"`
		// +optional
		WithAddonPolicies NodeGroupIAMAddonPolicies `yaml:"iam_node_group_with_addon_policies,omitempty"`
	}
	// NodeGroupIAMAddonPolicies holds all IAM addon policies
	NodeGroupIAMAddonPolicies struct {
		// +optional
		// ImageBuilder allows for full ECR (Elastic Container Registry) access. This is useful for building, for
		// example, a CI server that needs to push images to ECR
		ImageBuilder *bool `yaml:"image_builder,omitempty"`
		// +optional
		// AutoScaler enables IAM policy for cluster-autoscaler
		AutoScaler *bool `yaml:"auto_scaler,omitempty"`
		// +optional
		// ExternalDNS adds the external-dns project policies for Amazon Route 53
		ExternalDNS *bool `yaml:"external_dns,omitempty"`
		// +optional
		// CertManager enables the ability to add records to Route 53 in order to solve the DNS01 challenge. More information can be found
		// [here](https://cert-manager.io/docs/configuration/acme/dns01/route53/#set-up-a-iam-role)
		CertManager *bool `yaml:"cert_manager,omitempty"`
		// +optional
		// AppMesh enables full access to AppMesh
		AppMesh *bool `yaml:"app_mesh,omitempty"`
		// +optional
		// AppMeshPreview enables full access to AppMesh Preview
		AppMeshPreview *bool `yaml:"app_mesh_review,omitempty"`
		// +optional
		// EBS enables the new EBS CSI (Elastic Block Store Container Storage Interface) driver
		EBS *bool `yaml:"ebs,omitempty"`
		// +optional
		FSX *bool `yaml:"fsx,omitempty"`
		// +optional
		EFS *bool `yaml:"efs,omitempty"`
		// +optional
		AWSLoadBalancerController *bool `yaml:"alb_ingress,omitempty"`
		// +optional
		XRay *bool `yaml:"xray,omitempty"`
		// +optional
		CloudWatch *bool `yaml:"cloud_watch,omitempty"`
	}

	// NodeGroupSSH holds all the ssh access configuration to a NodeGroup
	NodeGroupSSH struct {
		// +optional If Allow is true the SSH configuration provided is used, otherwise it is ignored. Only one of
		// PublicKeyPath, PublicKey and PublicKeyName can be configured
		Allow *bool `yaml:"allow,omitempty"`
		// +optional The path to the SSH public key to be added to the nodes SSH keychain. If Allow is true this value
		// defaults to "~/.ssh/id_rsa.pub", otherwise the value is ignored.
		PublicKeyPath string `yaml:"publicKeyPath,omitempty"`
		// +optional Public key to be added to the nodes SSH keychain. If Allow is false this value is ignored.
		PublicKey string `yaml:"public_key,omitempty"`
		// +optional Public key name in EC2 to be added to the nodes SSH keychain. If Allow is false this value
		// is ignored.
		PublicKeyName string `yaml:"public_key_name,omitempty"`
		// +optional
		SourceSecurityGroupIDs []string `yaml:"source_security_group_ids,omitempty"`
		// Enables the ability to [SSH onto nodes using SSM](/introduction#ssh-access)
		// +optional
		EnableSSM *bool `yaml:"enable_ssm,omitempty"`
	}

	// NodeGroupInstancesDistribution holds the configuration for [spot
	// instances](/usage/spot-instances/)
	NodeGroupInstancesDistribution struct {
		// +required
		InstanceTypes []string `yaml:"instance_types,omitempty"`
		// Defaults to `on demand price`
		// +optional
		MaxPrice *float64 `yaml:"max_price,omitempty"`
		// Defaults to `0`
		// +optional
		OnDemandBaseCapacity *int `yaml:"on_demand_base_capacity,omitempty"`
		// Range [0-100]
		// Defaults to `100`
		// +optional
		OnDemandPercentageAboveBaseCapacity *int `yaml:"on_demand_percentage_above_base_capacity,omitempty"`
		// Range [1-20]
		// Defaults to `2`
		// +optional
		SpotInstancePools *int `yaml:"spot_instance_pools,omitempty"`
		// +optional
		SpotAllocationStrategy string `yaml:"spot_allocation_strategy,omitempty"`
		// Enable [capacity
		// rebalancing](https://docs.aws.amazon.com/autoscaling/ec2/userguide/capacity-rebalance.html)
		// for spot instances
		// +optional
		CapacityRebalance *bool `yaml:"capacity_rebalance"`
	}

	// NodeGroupBottlerocket holds the configuration for Bottlerocket based
	// NodeGroups.
	NodeGroupBottlerocket struct {
		// +optional
		EnableAdminContainer *bool `yaml:"enable_admin_container,omitempty"`
		// Settings contains any [bottlerocket
		// settings](https://github.com/bottlerocket-os/bottlerocket/#description-of-settings)
		// +optional
		Settings map[string]string `yaml:"settings,omitempty"`
		//Settings *InlineDocument `yaml:"settings,omitempty"`
	}

	// NodeGroupUpdateConfig contains the configuration for updating NodeGroups.
	NodeGroupUpdateConfig struct {
		// MaxUnavailable sets the max number of nodes that can become unavailable
		// when updating a nodegroup (specified as number)
		// +optional
		MaxUnavailable *int `yaml:"max_unavailable,omitempty"`

		// MaxUnavailablePercentage sets the max number of nodes that can become unavailable
		// when updating a nodegroup (specified as percentage)
		// +optional
		MaxUnavailablePercentage *int `yaml:"max_unavailable_percentage,omitempty"`
	}
)
type ManagedNodeGroup struct {
	// +required
	Name string `yaml:"name"`

	// Valid variants are `NodeAMIFamily` constants
	// +optional
	AMIFamily string `yaml:"ami_family,omitempty"`
	// +optional
	InstanceType string `yaml:"instance_type,omitempty"`
	// Limit [nodes to specific
	// AZs](/usage/autoscaling/#zone-aware-auto-scaling)
	// +optional
	AvailabilityZones []string `yaml:"avalability_zones,omitempty"`
	// Limit nodes to specific subnets
	// +optional
	Subnets []string `yaml:"subnets,omitempty"`

	// +optional
	InstancePrefix string `yaml:"instance_prefix,omitempty"`
	// +optional
	InstanceName string `yaml:"instance_name,omitempty"`

	// +optional
	//ScalingConfig
	// +optional
	DesiredCapacity *int `yaml:"desired_capacity,omitempty"`
	// +optional
	MinSize *int `yaml:"min_size,omitempty"`
	// +optional
	MaxSize *int `yaml:"max_size,omitempty"`

	// +optional
	// VolumeSize gigabytes
	// Defaults to `80`
	VolumeSize *int `yaml:"volume_size,omitempty"`
	// +optional
	// SSH configures ssh access for this nodegroup
	SSH *NodeGroupSSH `yaml:"ssh,omitempty"`
	// +optional
	Labels map[string]string `yaml:"labels,omitempty"`
	// Enable [private
	// networking](/usage/vpc-networking/#use-private-subnets-for-initial-nodegroup)
	// for nodegroup
	// +optional
	PrivateNetworking *bool `yaml:"private_networking"`
	// Applied to the Autoscaling Group and to the EC2 instances (unmanaged),
	// Applied to the EKS Nodegroup resource and to the EC2 instances (managed)
	// +optional
	Tags map[string]string `yaml:"tags,omitempty"`
	// +optional
	IAM *NodeGroupIAM `yaml:"iam,omitempty"`

	// Specify [custom AMIs](/usage/custom-ami-support/), `auto-ssm`, `auto`, or `static`
	// +optional
	AMI string `yaml:"ami,omitempty"`

	// +optional
	SecurityGroups *NodeGroupSGs `yaml:"security_groups,omitempty"`

	// +optional
	MaxPodsPerNode *int `yaml:"max_pods_per_node,omitempty"`

	// See [relevant AWS
	// docs](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-attribute-updatepolicy.html#cfn-attributes-updatepolicy-rollingupdate-suspendprocesses)
	// +optional
	ASGSuspendProcesses []string `yaml:"asg_suspend_processes,omitempty"`

	// EBSOptimized enables [EBS
	// optimization](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ebs-optimized.html)
	// +optional
	EBSOptimized *bool `yaml:"ebs_optimized,omitempty"`

	// Valid variants are `VolumeType` constants
	// +optional
	VolumeType string `yaml:"volume_type,omitempty"`
	// +optional
	VolumeName string `yaml:"volume_name,omitempty"`
	// +optional
	VolumeEncrypted *bool `yaml:"volume_encrypted,omitempty"`
	// +optional
	VolumeKmsKeyID string `yaml:"volume_kms_key_id,omitempty"`
	// +optional
	VolumeIOPS *int `yaml:"volume_iops,omitempty"`
	// +optional
	VolumeThroughput *int `yaml:"volume_throughput,omitempty"`

	// PreBootstrapCommands are executed before bootstrapping instances to the
	// cluster
	// +optional
	PreBootstrapCommands []string `yaml:"pre_bootstrap_commands,omitempty"`

	// Override `eksctl`'s bootstrapping script
	// +optional
	OverrideBootstrapCommand string `yaml:"override_bootstrap_command,omitempty"`

	// DisableIMDSv1 requires requests to the metadata service to use IMDSv2 tokens
	// Defaults to `false`
	// +optional
	DisableIMDSv1 *bool `yaml:"disable_imdsv1,omitempty"`

	// DisablePodIMDS blocks all IMDS requests from non host networking pods
	// Defaults to `false`
	// +optional
	DisablePodIMDS *bool `yaml:"disable_pods_imds,omitempty"`

	// Placement specifies the placement group in which nodes should
	// be spawned
	// +optional
	Placement *Placement `yaml:"placement,omitempty"`

	// EFAEnabled creates the maximum allowed number of EFA-enabled network
	// cards on nodes in this group.
	// +optional
	EFAEnabled *bool `yaml:"efa_enabled,omitempty"`

	// InstanceSelector specifies options for EC2 instance selector
	InstanceSelector *InstanceSelector `yaml:"instance_selector,omitempty"`

	// Internal fields
	// Some AMIs (bottlerocket) have a separate volume for the OS
	AdditionalEncryptedVolume string `yaml:"-"`

	// Bottlerocket specifies settings for Bottlerocket nodes
	// +optional
	Bottlerocket *NodeGroupBottlerocket `yaml:"bottle_rocket,omitempty"`

	// TODO remove this
	// This is a hack, will be removed shortly. When this is true for Ubuntu and
	// AL2 images a legacy bootstrapper will be used.
	CustomAMI *bool `yaml:"-"`

	// Enable EC2 detailed monitoring
	// +optional
	EnableDetailedMonitoring *bool `yaml:"enable_detailed_monitoring,omitempty"`
	// Rafay changes - start
	// Internal
	IsWavelengthZone *bool `yaml:"-"`
	// Rafay changes - end

	// InstanceTypes specifies a list of instance types
	InstanceTypes []string `yaml:"instance_types,omitempty"`

	// Spot creates a spot nodegroup
	Spot *bool `yaml:"spot,omitempty"`

	// Taints taints to apply to the nodegroup
	Taints []NodeGroupTaint `yaml:"taints,omitempty"`

	// UpdateConfig configures how to update NodeGroups.
	// +optional
	UpdateConfig *NodeGroupUpdateConfig `yaml:"update_config,omitempty"`

	// LaunchTemplate specifies an existing launch template to use
	// for the nodegroup
	LaunchTemplate *LaunchTemplate `yaml:"launch_tempelate,omitempty"`

	// ReleaseVersion the AMI version of the EKS optimized AMI to use
	ReleaseVersion string `yaml:"release_version,omitempty"`

	// Internal fields

	Unowned bool `yaml:"-"`
}
type LaunchTemplate struct {
	// Launch template ID
	// +required
	ID string `yaml:"id,omitempty"`
	// Launch template version
	// Defaults to the default launch template version
	// TODO support $Default, $Latest
	Version string `yaml:"version,omitempty"`
}
type FargateProfile struct {

	// Name of the Fargate profile.
	// +required
	Name string `yaml:"name"`

	// PodExecutionRoleARN is the IAM role's ARN to use to run pods onto Fargate.
	PodExecutionRoleARN string `yaml:"pod_execution_role_arn,omitempty"`

	// Selectors define the rules to select workload to schedule onto Fargate.
	Selectors []FargateProfileSelector `yaml:"selectors,omitempty"`

	// Subnets which Fargate should use to do network placement of the selected workload.
	// If none provided, all subnets for the cluster will be used.
	// +optional
	Subnets []string `yaml:"subnets,omitempty"`

	// Used to tag the AWS resources
	// +optional
	Tags map[string]string `yaml:"tags,omitempty"`

	// The current status of the Fargate profile.
	Status string `yaml:"status,omitempty"`
}

// FargateProfileSelector defines rules to select workload to schedule onto Fargate.
type FargateProfileSelector struct {

	// Namespace is the Kubernetes namespace from which to select workload.
	// +required
	Namespace string `yaml:"namespace,omitempty"`

	// Labels are the Kubernetes label selectors to use to select workload.
	// +optional
	Labels map[string]string `yaml:"labels,omitempty"`
}

/*
// EKSClusterVPC struct -> cfg.vpc
type EKSClusterVPC struct {
	CIDR             string               `yaml:"cidr,omitempty" yaml:"cidr"`
	Subnets          *EKSClusterSubnets   `yaml:"subnets,omitempty"`
	NAT              *EKSClusterNAT       `yaml:"nat,omitempty"`
	ClusterEndpoints *EKSClusterEndpoints `yaml:"clusterEndpoints,omitempty"`
}

// EKSClusterNAT struct -> cfg.vpc.nat
type EKSClusterNAT struct {
	Gateway string `yaml:"gateway,omitempty"`
}
*/
// EKSClusterSubnets struct -> cfg.vpc.subnets
type EKSClusterSubnets struct {
	Private map[string]EKSAZSubnetSpec `yaml:"private,omitempty"`
	Public  map[string]EKSAZSubnetSpec `yaml:"public,omitempty"`
}

// EKSAZSubnetSpec struct -> cfg.vpc.subnets.(private|public)[randomKey]
type EKSAZSubnetSpec struct {
	ID string `yaml:"id,omitempty"`
}

// EKSClusterEndpoints struct -> cfg.vpc.clusterEndpoints
type EKSClusterEndpoints struct {
	PrivateAccess *bool `yaml:"privateAccess,omitempty"`
	PublicAccess  *bool `yaml:"publicAccess,omitempty"`
}

// EKSScalingConfig struct -> embedded into NodeGroupBase
type EKSScalingConfig struct {
	DesiredCapacity *int64 `yaml:"desiredCapacity,omitempty"`
	MinSize         *int64 `yaml:"minSize,omitempty"`
	MaxSize         *int64 `yaml:"maxSize,omitempty"`
}

// EKSNodeGroupBase struct -> embedded into cfg.nodeGroups[]
type EKSNodeGroupBase struct {
	Name              string   `yaml:"name"`
	AMIFamily         string   `yaml:"amiFamily,omitempty"`
	InstanceType      string   `yaml:"instanceType,omitempty"`
	AvailabilityZones []string `yaml:"availabilityZones,omitempty"`
	*EKSScalingConfig
	VolumeSize           *int64            `yaml:"volumeSize,omitempty"`
	SSH                  *EKSNodeGroupSSH  `yaml:"ssh,omitempty"`
	Labels               map[string]string `yaml:"labels,omitempty"`
	PrivateNetworking    *bool             `yaml:"privateNetworking,omitempty"`
	Tags                 map[string]string `yaml:"tags,omitempty"`
	IAM                  *EKSNodeGroupIAM  `yaml:"iam,omitempty"`
	AMI                  string            `yaml:"ami,omitempty"`
	MaxPodsPerNode       *int64            `yaml:"maxPodsPerNode,omitempty"`
	SecurityGroups       *EKSNodeGroupSGs  `yaml:"securityGroups,omitempty"`
	VolumeType           string            `yaml:"volumeType,omitempty"`
	VolumeEncrypted      *bool             `yaml:"volumeEncrypted,omitempty"`
	VolumeKmsKeyID       string            `yaml:"volumeKmsKeyID,omitempty"`
	PreBootstrapCommands []string          `yaml:"preBootstrapCommands,omitempty"`
	Subnets              []string          `yaml:"subnets,omitempty"`
}

// EKSNodeGroup struct -> cfg.nodeGroups[]
type EKSNodeGroup struct {
	*EKSNodeGroupBase
	InstancesDistribution *EKSNodeGroupInstancesDistribution `yaml:"instancesDistribution,omitempty"`
	SubnetCidr            string                             `yaml:"subnetCidr,omitempty"`
}

// EKSManagedNodeGroup struct -> cfg.
type EKSManagedNodeGroup struct {
	*EKSNodeGroupBase
	InstanceTypes []string `yaml:"instanceTypes,omitempty"`
	Spot          *bool    `yaml:"spot,omitempty"`
}

// EKSNodeGroupIAM struct -> cfg.nodeGroups[].iam
type EKSNodeGroupIAM struct {
	InstanceProfileARN              string                        `yaml:"instanceProfileARN,omitempty"`
	InstanceRoleARN                 string                        `yaml:"instanceRoleARN,omitempty"`
	InstanceRolePermissionsBoundary string                        `yaml:"instanceRolePermissionsBoundary,omitempty"`
	WithAddonPolicies               *EKSNodeGroupIAMAddonPolicies `yaml:"withAddonPolicies,omitempty"`
}

// EKSNodeGroupIAMAddonPolicies struct -> cfg.nodeGroups[].iam.withAddonPolicies
type EKSNodeGroupIAMAddonPolicies struct {
	ImageBuilder              *bool `yaml:"imageBuilder,omitempty"`
	AutoScaler                *bool `yaml:"autoScaler,omitempty"`
	ExternalDNS               *bool `yaml:"externalDNS,omitempty"`
	AppMesh                   *bool `yaml:"appMesh,omitempty"`
	AWSLoadBalancerController *bool `yaml:"albIngress,omitempty"`
	EFS                       *bool `yaml:"efs,omitempty"`
}

// Make sure to update hasAtLeastOneEnabled()

// EKSNodeGroupSGs struct -> cfg.nodeGroups[].SecurityGroups
type EKSNodeGroupSGs struct {
	AttachIDs []string `yaml:"attachIDs,omitempty"`
}

// EKSNodeGroupSSH struct -> cfg.nodeGroups[].ssh
type EKSNodeGroupSSH struct {
	Allow         *bool  `yaml:"allow"`
	PublicKeyName string `yaml:"publicKeyName,omitempty"`
}

// EKSNodeGroupInstancesDistribution struct -> cfg.nodeGroups[].instancesDistribution
type EKSNodeGroupInstancesDistribution struct {
	InstanceTypes                       []string `yaml:"instanceTypes,omitempty"`
	MaxPrice                            *float64 `yaml:"maxPrice,omitempty"`
	OnDemandBaseCapacity                *int64   `yaml:"onDemandBaseCapacity,omitempty"`
	OnDemandPercentageAboveBaseCapacity *int64   `yaml:"onDemandPercentageAboveBaseCapacity,omitempty"`
	SpotInstancePools                   *int64   `yaml:"spotInstancePools,omitempty"`
	SpotAllocationStrategy              string   `yaml:"spotAllocationStrategy,omitempty"`
}

type EKSClusterCloudWatch struct {
	//+optional
	ClusterLogging *EKSClusterCloudWatchLogging `yaml:"cluster_logging,omitempty"`
}

// Values for `CloudWatchLogging`
const (
	APILogging               = "api"
	AuditLogging             = "audit"
	AuthenticatorLogging     = "authenticator"
	ControllerManagerLogging = "controllerManager"
	SchedulerLogging         = "scheduler"
	AllLogging               = "all"
	WildcardLogging          = "*"
)

// ClusterCloudWatchLogging container config parameters related to cluster logging
type EKSClusterCloudWatchLogging struct {

	// Types of logging to enable (see [CloudWatch docs](/usage/cloudwatch-cluster-logging/#clusterconfig-examples)).
	// Valid entries are `CloudWatchLogging` constants
	//+optional
	EnableTypes []string `yaml:"enable_types,omitempty"`
}

// SecretsEncryption defines the configuration for KMS encryption provider
type SecretsEncryption struct {
	// +required
	KeyARN string `yaml:"key_arn,omitempty"`
}
