package rafay

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/weaveworks/eksctl/pkg/utils/ipnet"
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
type EKSClusterMetadata struct {
	Kind     string                  `yaml:"kind"`
	Metadata *EKSClusterMetaMetadata `yaml:"metadata"`
	Spec     *EKSSpec                `yaml:"spec"`
}

type EKSSpec struct {
	Type             string            `yaml:"type"`
	Blueprint        string            `yaml:"blueprint"`
	BlueprintVersion string            `yaml:"blueprintversion,omitempty"`
	CloudProvider    string            `yaml:"cloudprovider"`
	CniProvider      string            `yaml:"cniprovider"`
	ProxyConfig      map[string]string `yaml:"labels"`
}

type EKSClusterMetaMetadata struct {
	Name    string            `yaml:"name"`
	Project string            `yaml:"project"`
	Labels  map[string]string `yaml:"labels,omitempty"`
}

//struct for eks cluster config sped (second part of the yaml file kind:clusterConfig)
type KubernetesNetworkConfig struct {
	ServiceIPv4CIDR string `yaml:"serviceIPv4CIDR"`
}

type EKSClusterYamlConfig struct {
	APIVersion              string                   `yaml:"apiversion"`
	Kind                    string                   `yaml:"kind"`
	Metadata                *EKSClusterMeta          `yaml:"metadata"`
	KubernetesNetworkConfig *KubernetesNetworkConfig `yaml:"kubernetesNetworkConfig"`
	IAM                     *EKSClusterIAM           `yaml:"iam,omitempty"`
	IdentityProviders       []IdentityProvider       `json:"identityProviders,omitempty"`
	VPC                     *EKSClusterVPC           `yaml:"vpc,omitempty"`
	// +optional
	Addons []*Addon `yaml:"addons,omitempty"`
	// +optional
	PrivateCluster    *PrivateCluster       `yaml:"privateCluster,omitempty"`
	NodeGroups        []*NodeGroup          `yaml:"nodeGroups,omitempty"`
	ManagedNodeGroups []*ManagedNodeGroup   `yaml:"managedNodeGroups,omitempty"`
	FargateProfiles   []*FargateProfile     `yaml:"fargateProfiles,omitempty"`
	AvailabilityZones []string              `yaml:"availabilityZones,omitempty"`
	CloudWatch        *EKSClusterCloudWatch `yaml:"cloudWatch,omitempty"`
	SecretsEncryption *SecretsEncryption    `yaml:"secretsEncryption,omitempty"`
	//do i need this? not in docs
	//Karpenter *Karpenter `yaml:"karpenter,omitempty"`
}

/*Took this struct and modified it to fit documentation
// EKSClusterConfig struct -> cfg
type EKSClusterConfig struct {
	APIVersion  string          `json:"apiVersion"`
	Kind        string          `json:"kind"`
	ClusterMeta *EKSClusterMeta `json:"metadata"`
	IAM         *EKSClusterIAM  `json:"iam,omitempty"`
	// +optional
	IdentityProviders []IdentityProvider     `json:"identityProviders,omitempty"`
	VPC               *EKSClusterVPC         `json:"vpc,omitempty"`
	NodeGroups        []*EKSNodeGroup        `json:"nodeGroups,omitempty"`
	ManagedNodeGroups []*EKSManagedNodeGroup `json:"managedNodeGroups,omitempty"`
	CloudWatch        *EKSClusterCloudWatch  `json:"cloudWatch,omitempty"`

	AvailabilityZones []string `json:"availabilityZones,omitempty"`
}
*/
type AWSPolicyInlineDocument map[string]interface{}

// EKSClusterMeta struct -> cfg.EKSClusterMeta
type EKSClusterMeta struct {
	Name        string            `json:"name"`
	Region      string            `json:"region"`
	Version     string            `json:"version,omitempty"`
	Tags        map[string]string `json:"tags,omitempty"`
	Annotations map[string]string `json:"tags,omitempty"`
}

// EKSClusterIAM struct -> cfg.IAM.ServiceAccounts
type EKSClusterIAMMeta struct {
	Name        string            `json:"name,omitempty"`
	Namespace   string            `json:"namespace,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

/*
type EKSClusterIAMServiceAccount struct {
	EKSClusterIAMMeta   `json:"metadata,omitempty"`
	AttachPolicyARNs    []string                `json:"attachPolicyARNs,omitempty"`
	AttachPolicy        AWSPolicyInlineDocument `json:"attachPolicy,omitempty"`
	PermissionsBoundary string                  `json:"permissionsBoundary,omitempty"`
	RoleOnly            *bool                   `json:"roleOnly,omitempty"`
	Tags                map[string]string       `json:"tags,omitempty"`
	// RoleName string `json:"roleName,omitempty"`
}
*/
type IdentityProvider struct {
	// Valid variants are:
	// `"oidc"`: OIDC identity provider
	// +required
	type_ string `json:"type"` //nolint
	//Inner IdentityProviderInterface
}

// EKSClusterIAM struct -> cfg.IAM
type EKSClusterIAM struct {
	// +optional
	ServiceRoleARN string `json:"serviceRoleARN,omitempty"`

	// permissions boundary for all identity-based entities created by eksctl.
	// See [AWS Permission Boundary](https://docs.aws.amazon.com/IAM/latest/UserGuide/access_policies_boundaries.html)
	// +optional
	ServiceRolePermissionsBoundary string `json:"serviceRolePermissionsBoundary,omitempty"`

	// role used by pods to access AWS APIs. This role is added to the Kubernetes RBAC for authorization.
	// See [Pod Execution Role](https://docs.aws.amazon.com/eks/latest/userguide/pod-execution-role.html)
	// +optional
	FargatePodExecutionRoleARN string `json:"fargatePodExecutionRoleARN,omitempty"`

	// permissions boundary for the fargate pod execution role`. See [EKS Fargate Support](/usage/fargate-support/)
	// +optional
	FargatePodExecutionRolePermissionsBoundary string `json:"fargatePodExecutionRolePermissionsBoundary,omitempty"`

	// enables the IAM OIDC provider as well as IRSA for the Amazon CNI plugin
	// +optional
	WithOIDC bool `json:"withOIDC,omitempty"`

	// service accounts to create in the cluster.
	// See [IAM Service Accounts](/iamserviceaccounts/#usage-with-config-files)
	// +optional
	ServiceAccounts []*EKSClusterIAMServiceAccount `json:"serviceAccounts,omitempty"`

	// VPCResourceControllerPolicy attaches the IAM policy
	// necessary to run the VPC controller in the control plane
	// Defaults to `true`
	VPCResourceControllerPolicy bool `json:"vpcResourceControllerPolicy,omitempty"`
}

// ClusterIAMServiceAccount holds an IAM service account metadata and configuration
type EKSClusterIAMServiceAccount struct {
	EKSClusterIAMMeta `json:"metadata,omitempty"`

	// list of ARNs of the IAM policies to attach
	// +optional
	AttachPolicyARNs []string `json:"attachPolicyARNs,omitempty"`

	WellKnownPolicies WellKnownPolicies `json:"wellKnownPolicies,omitempty"`

	// AttachPolicy holds a policy document to attach to this service account
	// +optional
	//AttachPolicy map[string]string `json:"attachPolicy,omitempty"`
	AttachPolicy InlineDocument `json:"attachPolicy,omitempty"`

	// ARN of the role to attach to the service account
	AttachRoleARN string `json:"attachRoleARN,omitempty"`

	// ARN of the permissions boundary to associate with the service account
	// +optional
	PermissionsBoundary string `json:"permissionsBoundary,omitempty"`

	// +optional
	Status *ClusterIAMServiceAccountStatus `json:"status,omitempty"`

	// Specific role name instead of the Cloudformation-generated role name
	// +optional
	RoleName string `json:"roleName,omitempty"`

	// Specify if only the IAM Service Account role should be created without creating/annotating the service account
	// +optional
	RoleOnly *bool `json:"roleOnly,omitempty"`

	// AWS tags for the service account
	// +optional
	Tags map[string]string `json:"tags,omitempty"`
}

type WellKnownPolicies struct {
	// ImageBuilder allows for full ECR (Elastic Container Registry) access.
	ImageBuilder *bool `json:"imageBuilder,inline"`
	// AutoScaler adds policies for cluster-autoscaler. See [autoscaler AWS
	// docs](https://docs.aws.amazon.com/eks/latest/userguide/cluster-autoscaler.html).
	AutoScaler *bool `json:"autoScaler,inline"`
	// AWSLoadBalancerController adds policies for using the
	// aws-load-balancer-controller. See [Load Balancer
	// docs](https://docs.aws.amazon.com/eks/latest/userguide/aws-load-balancer-controller.html).
	AWSLoadBalancerController *bool `json:"awsLoadBalancerController,inline"`
	// ExternalDNS adds external-dns policies for Amazon Route 53.
	// See [external-dns
	// docs](https://github.com/kubernetes-sigs/external-dns/blob/master/docs/tutorials/aws.md).
	ExternalDNS *bool `json:"externalDNS,inline"`
	// CertManager adds cert-manager policies. See [cert-manager
	// docs](https://cert-manager.io/docs/configuration/acme/dns01/route53).
	CertManager *bool `json:"certManager,inline"`
	// EBSCSIController adds policies for using the
	// ebs-csi-controller. See [aws-ebs-csi-driver
	// docs](https://github.com/kubernetes-sigs/aws-ebs-csi-driver#set-up-driver-permission).
	EBSCSIController *bool `json:"ebsCSIController,inline"`
	// EFSCSIController adds policies for using the
	// efs-csi-controller. See [aws-efs-csi-driver
	// docs](https://aws.amazon.com/blogs/containers/introducing-efs-csi-dynamic-provisioning).
	EFSCSIController *bool `json:"efsCSIController,inline"`
}

type InlineDocument map[string]interface{}

type ClusterIAMServiceAccountStatus struct {
	// +optional
	RoleARN string `json:"roleARN,omitempty"`
}
type AZSubnetMapping map[string]AZSubnetSpec
type (
	// ClusterVPC holds global subnet and all child subnets
	EKSClusterVPC struct {
		// global CIDR and VPC ID
		// +optional
		Network
		// SecurityGroup (aka the ControlPlaneSecurityGroup) for communication between control plane and nodes
		// +optional
		SecurityGroup string `json:"securityGroup,omitempty"`
		// Subnets are keyed by AZ for convenience.
		// See [this example](/examples/reusing-iam-and-vpc/)
		// as well as [using existing
		// VPCs](/usage/vpc-networking/#use-existing-vpc-other-custom-configuration).
		// +optional
		Subnets *ClusterSubnets `json:"subnets,omitempty"`
		// for additional CIDR associations, e.g. a CIDR for
		// private subnets or any ad-hoc subnets
		// +optional
		ExtraCIDRs []string `json:"extraCIDRs,omitempty"`
		// for additional IPv6 CIDR associations, e.g. a CIDR for
		// private subnets or any ad-hoc subnets
		// +optional
		ExtraIPv6CIDRs []string `json:"extraIPv6CIDRs,omitempty"`
		// for pre-defined shared node SG
		SharedNodeSecurityGroup string `json:"sharedNodeSecurityGroup,omitempty"`
		// Automatically add security group rules to and from the default
		// cluster security group and the shared node security group.
		// This allows unmanaged nodes to communicate with the control plane
		// and managed nodes.
		// This option cannot be disabled when using eksctl created security groups.
		// Defaults to `true`
		// +optional
		ManageSharedNodeSecurityGroupRules *bool `json:"manageSharedNodeSecurityGroupRules,omitempty"`
		// AutoAllocateIPV6 requests an IPv6 CIDR block with /56 prefix for the VPC
		// +optional
		AutoAllocateIPv6 *bool `json:"autoAllocateIPv6,omitempty"`
		// +optional
		NAT *ClusterNAT `json:"nat,omitempty"`
		// See [managing access to API](/usage/vpc-networking/#managing-access-to-the-kubernetes-api-server-endpoints)
		// +optional
		ClusterEndpoints *ClusterEndpoints `json:"clusterEndpoints,omitempty"`
		// PublicAccessCIDRs are which CIDR blocks to allow access to public
		// k8s API endpoint
		// +optional
		PublicAccessCIDRs []string `json:"publicAccessCIDRs,omitempty"`
	}
	// ClusterSubnets holds private and public subnets
	ClusterSubnets struct {
		Private AZSubnetMapping `json:"private,omitempty"`
		Public  AZSubnetMapping `json:"public,omitempty"`
	}
	// SubnetTopology can be SubnetTopologyPrivate or SubnetTopologyPublic
	SubnetTopology string
	AZSubnetSpec   struct {
		// +optional
		ID string `json:"id,omitempty"`
		// AZ can be omitted if the key is an AZ
		// +optional
		AZ string `json:"az,omitempty"`
		// +optional
		//can i just make this a string?
		//CIDR string `yaml:"cidr"`
		CIDR *ipnet.IPNet `json:"cidr,omitempty"`
	}
	// Network holds ID and CIDR
	Network struct {
		// +optional
		ID string `json:"id,omitempty"`
		// +optional
		//can i just make this a string?
		CIDR string `yaml:"cidr"`
		//CIDR *ipnet.IPNet `json:"cidr,omitempty"`
		// +optional
		IPv6Cidr string `json:"ipv6Cidr,omitempty"`
		// +optional
		IPv6Pool string `json:"ipv6Pool,omitempty"`
	}
	// ClusterNAT NAT config
	ClusterNAT struct {
		// Valid variants are `ClusterNAT` constants
		Gateway string `json:"gateway,omitempty"`
	}

	// ClusterEndpoints holds cluster api server endpoint access information
	ClusterEndpoints struct {
		PrivateAccess *bool `json:"privateAccess,omitempty"`
		PublicAccess  *bool `json:"publicAccess,omitempty"`
	}
)
type Addon struct {
	// +required
	Name string `json:"name,omitempty"`
	// +optional
	Version string `json:"version,omitempty"`
	// +optional
	ServiceAccountRoleARN string `json:"serviceAccountRoleARN,omitempty"`
	// list of ARNs of the IAM policies to attach
	// +optional
	AttachPolicyARNs []string `json:"attachPolicyARNs,omitempty"`
	// AttachPolicy holds a policy document to attach
	// +optional
	AttachPolicy InlineDocument `json:"attachPolicy,omitempty"`
	// ARN of the permissions' boundary to associate
	// +optional
	PermissionsBoundary string `json:"permissionsBoundary,omitempty"`
	// WellKnownPolicies for attaching common IAM policies
	//WellKnown Policies not in documentation for addon? (same field as IAM wellknow-policies)
	//WellKnownPolicies WellKnownPolicies `json:"wellKnownPolicies,omitempty"`
	// The metadata to apply to the cluster to assist with categorization and organization.
	// Each tag consists of a key and an optional value, both of which you define.
	// +optional
	Tags map[string]string `json:"tags,omitempty"`
	// Force applies the add-on to overwrite an existing add-on
	Force bool `json:"-"`
}

// PrivateCluster defines the configuration for a fully-private cluster
type PrivateCluster struct {

	// Enabled enables creation of a fully-private cluster
	Enabled bool `json:"enabled"`

	// SkipEndpointCreation skips the creation process for endpoints completely. This is only used in case of an already
	// provided VPC and if the user decided to set it to true.
	SkipEndpointCreation bool `json:"skipEndpointCreation"`

	// AdditionalEndpointServices specifies additional endpoint services that
	// must be enabled for private access.
	// Valid entries are `AdditionalEndpointServices` constants
	AdditionalEndpointServices []string `json:"additionalEndpointServices,omitempty"`
}
type NodeGroup struct {
	*NodeGroupBase

	//+optional
	InstancesDistribution *NodeGroupInstancesDistribution `json:"instancesDistribution,omitempty"`

	// +optional
	ASGMetricsCollection []MetricsCollection `json:"asgMetricsCollection,omitempty"`

	// CPUCredits configures [T3 Unlimited](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/burstable-performance-instances-unlimited-mode.html), valid only for T-type instances
	// +optional
	CPUCredits string `json:"cpuCredits,omitempty"`

	// Associate load balancers with auto scaling group
	// +optional
	ClassicLoadBalancerNames []string `json:"classicLoadBalancerNames,omitempty"`

	// Associate target group with auto scaling group
	// +optional
	TargetGroupARNs []string `json:"targetGroupARNs,omitempty"`

	// Taints taints to apply to the nodegroup
	// +optional
	Taints taintsWrapper `json:"taints,omitempty"`

	// UpdateConfig configures how to update NodeGroups.
	// +optional
	UpdateConfig *NodeGroupUpdateConfig `json:"updateConfig,omitempty"`

	// [Custom
	// address](/usage/vpc-networking/#custom-cluster-dns-address) used for DNS
	// lookups
	// +optional
	ClusterDNS string `json:"clusterDNS,omitempty"`

	// [Customize `kubelet` config](/usage/customizing-the-kubelet/)
	// +optional

	KubeletExtraConfig *InlineDocument `json:"kubeletExtraConfig,omitempty"`

	// ContainerRuntime defines the runtime (CRI) to use for containers on the node
	// +optional
	ContainerRuntime string `json:"containerRuntime,omitempty"`
}

// NodeGroupBase represents the base nodegroup config for self-managed and managed nodegroups
type NodeGroupBase struct {
	// +required
	Name string `json:"name"`

	// Valid variants are `NodeAMIFamily` constants
	// +optional
	AMIFamily string `json:"amiFamily,omitempty"`
	// +optional
	InstanceType string `json:"instanceType,omitempty"`
	// Limit [nodes to specific
	// AZs](/usage/autoscaling/#zone-aware-auto-scaling)
	// +optional
	AvailabilityZones []string `json:"availabilityZones,omitempty"`
	// Limit nodes to specific subnets
	// +optional
	Subnets []string `json:"subnets,omitempty"`

	// +optional
	InstancePrefix string `json:"instancePrefix,omitempty"`
	// +optional
	InstanceName string `json:"instanceName,omitempty"`

	// +optional
	*ScalingConfig

	// +optional
	// VolumeSize gigabytes
	// Defaults to `80`
	VolumeSize *int `json:"volumeSize,omitempty"`
	// +optional
	// SSH configures ssh access for this nodegroup
	SSH *NodeGroupSSH `json:"ssh,omitempty"`
	// +optional
	Labels map[string]string `json:"labels,omitempty"`
	// Enable [private
	// networking](/usage/vpc-networking/#use-private-subnets-for-initial-nodegroup)
	// for nodegroup
	// +optional
	PrivateNetworking bool `json:"privateNetworking"`
	// Applied to the Autoscaling Group and to the EC2 instances (unmanaged),
	// Applied to the EKS Nodegroup resource and to the EC2 instances (managed)
	// +optional
	Tags map[string]string `json:"tags,omitempty"`
	// +optional
	IAM *NodeGroupIAM `json:"iam,omitempty"`

	// Specify [custom AMIs](/usage/custom-ami-support/), `auto-ssm`, `auto`, or `static`
	// +optional
	AMI string `json:"ami,omitempty"`

	// +optional
	SecurityGroups *NodeGroupSGs `json:"securityGroups,omitempty"`

	// +optional
	MaxPodsPerNode int `json:"maxPodsPerNode,omitempty"`

	// See [relevant AWS
	// docs](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-attribute-updatepolicy.html#cfn-attributes-updatepolicy-rollingupdate-suspendprocesses)
	// +optional
	ASGSuspendProcesses []string `json:"asgSuspendProcesses,omitempty"`

	// EBSOptimized enables [EBS
	// optimization](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ebs-optimized.html)
	// +optional
	EBSOptimized *bool `json:"ebsOptimized,omitempty"`

	// Valid variants are `VolumeType` constants
	// +optional
	VolumeType string `json:"volumeType,omitempty"`
	// +optional
	VolumeName string `json:"volumeName,omitempty"`
	// +optional
	VolumeEncrypted *bool `json:"volumeEncrypted,omitempty"`
	// +optional
	VolumeKmsKeyID string `json:"volumeKmsKeyID,omitempty"`
	// +optional
	VolumeIOPS *int `json:"volumeIOPS,omitempty"`
	// +optional
	VolumeThroughput *int `json:"volumeThroughput,omitempty"`

	// PreBootstrapCommands are executed before bootstrapping instances to the
	// cluster
	// +optional
	PreBootstrapCommands []string `json:"preBootstrapCommands,omitempty"`

	// Override `eksctl`'s bootstrapping script
	// +optional
	OverrideBootstrapCommand string `json:"overrideBootstrapCommand,omitempty"`

	// DisableIMDSv1 requires requests to the metadata service to use IMDSv2 tokens
	// Defaults to `false`
	// +optional
	DisableIMDSv1 *bool `json:"disableIMDSv1,omitempty"`

	// DisablePodIMDS blocks all IMDS requests from non host networking pods
	// Defaults to `false`
	// +optional
	DisablePodIMDS *bool `json:"disablePodIMDS,omitempty"`

	// Placement specifies the placement group in which nodes should
	// be spawned
	// +optional
	Placement *Placement `json:"placement,omitempty"`

	// EFAEnabled creates the maximum allowed number of EFA-enabled network
	// cards on nodes in this group.
	// +optional
	EFAEnabled *bool `json:"efaEnabled,omitempty"`

	// InstanceSelector specifies options for EC2 instance selector
	InstanceSelector *InstanceSelector `json:"instanceSelector,omitempty"`

	// Internal fields
	// Some AMIs (bottlerocket) have a separate volume for the OS
	AdditionalEncryptedVolume string `json:"-"`

	// Bottlerocket specifies settings for Bottlerocket nodes
	// +optional
	Bottlerocket *NodeGroupBottlerocket `json:"bottlerocket,omitempty"`

	// TODO remove this
	// This is a hack, will be removed shortly. When this is true for Ubuntu and
	// AL2 images a legacy bootstrapper will be used.
	CustomAMI bool `json:"-"`

	// Enable EC2 detailed monitoring
	// +optional
	EnableDetailedMonitoring *bool `json:"enableDetailedMonitoring,omitempty"`
	// Rafay changes - start
	// Internal
	IsWavelengthZone bool `json:"-"`
	// Rafay changes - end
}
type InstanceSelector struct {
	// VCPUs specifies the number of vCPUs
	VCPUs int `json:"vCPUs,omitempty"`
	// Memory specifies the memory
	// The unit defaults to GiB
	Memory string `json:"memory,omitempty"`
	// GPUs specifies the number of GPUs.
	// It can be set to 0 to select non-GPU instance types.
	GPUs int `json:"gpus,omitempty"`
	// CPU Architecture of the EC2 instance type.
	// Valid variants are:
	// `"x86_64"`
	// `"amd64"`
	// `"arm64"`
	CPUArchitecture string `json:"cpuArchitecture,omitempty"`
}
type Placement struct {
	GroupName string `json:"groupName,omitempty"`
}
type ScalingConfig struct {
	// +optional
	DesiredCapacity *int `json:"desiredCapacity,omitempty"`
	// +optional
	MinSize *int `json:"minSize,omitempty"`
	// +optional
	MaxSize *int `json:"maxSize,omitempty"`
}
type MetricsCollection struct {
	// +required
	Granularity string `json:"granularity"`
	// +optional
	Metrics []string `json:"metrics,omitempty"`
}
type taintsWrapper []NodeGroupTaint
type NodeGroupTaint struct {
	Key    string `json:"key,omitempty"`
	Value  string `json:"value,omitempty"`
	Effect string `json:"effect,omitempty"`
}
type (
	// NodeGroupSGs controls security groups for this nodegroup
	NodeGroupSGs struct {
		// AttachIDs attaches additional security groups to the nodegroup
		// +optional
		AttachIDs []string `json:"attachIDs,omitempty"`
		// WithShared attach the security group
		// shared among all nodegroups in the cluster
		// Defaults to `true`
		// +optional
		WithShared *bool `json:"withShared"`
		// WithLocal attach a security group
		// local to this nodegroup
		// Not supported for managed nodegroups
		// Defaults to `true`
		// +optional
		WithLocal *bool `json:"withLocal"`
	}
	// NodeGroupIAM holds all IAM attributes of a NodeGroup
	NodeGroupIAM struct {
		// AttachPolicy holds a policy document to attach
		// +optional
		AttachPolicy InlineDocument `json:"attachPolicy,omitempty"`
		// list of ARNs of the IAM policies to attach
		// +optional
		AttachPolicyARNs []string `json:"attachPolicyARNs,omitempty"`
		// +optional
		InstanceProfileARN string `json:"instanceProfileARN,omitempty"`
		// +optional
		InstanceRoleARN string `json:"instanceRoleARN,omitempty"`
		// +optional
		InstanceRoleName string `json:"instanceRoleName,omitempty"`
		// +optional
		InstanceRolePermissionsBoundary string `json:"instanceRolePermissionsBoundary,omitempty"`
		// +optional
		WithAddonPolicies NodeGroupIAMAddonPolicies `json:"withAddonPolicies,omitempty"`
	}
	// NodeGroupIAMAddonPolicies holds all IAM addon policies
	NodeGroupIAMAddonPolicies struct {
		// +optional
		// ImageBuilder allows for full ECR (Elastic Container Registry) access. This is useful for building, for
		// example, a CI server that needs to push images to ECR
		ImageBuilder *bool `json:"imageBuilder"`
		// +optional
		// AutoScaler enables IAM policy for cluster-autoscaler
		AutoScaler *bool `json:"autoScaler"`
		// +optional
		// ExternalDNS adds the external-dns project policies for Amazon Route 53
		ExternalDNS *bool `json:"externalDNS"`
		// +optional
		// CertManager enables the ability to add records to Route 53 in order to solve the DNS01 challenge. More information can be found
		// [here](https://cert-manager.io/docs/configuration/acme/dns01/route53/#set-up-a-iam-role)
		CertManager *bool `json:"certManager"`
		// +optional
		// AppMesh enables full access to AppMesh
		AppMesh *bool `json:"appMesh"`
		// +optional
		// AppMeshPreview enables full access to AppMesh Preview
		AppMeshPreview *bool `json:"appMeshPreview"`
		// +optional
		// EBS enables the new EBS CSI (Elastic Block Store Container Storage Interface) driver
		EBS *bool `json:"ebs"`
		// +optional
		FSX *bool `json:"fsx"`
		// +optional
		EFS *bool `json:"efs"`
		// +optional
		AWSLoadBalancerController *bool `json:"albIngress"`
		// +optional
		XRay *bool `json:"xRay"`
		// +optional
		CloudWatch *bool `json:"cloudWatch"`
	}

	// NodeGroupSSH holds all the ssh access configuration to a NodeGroup
	NodeGroupSSH struct {
		// +optional If Allow is true the SSH configuration provided is used, otherwise it is ignored. Only one of
		// PublicKeyPath, PublicKey and PublicKeyName can be configured
		Allow *bool `json:"allow"`
		// +optional The path to the SSH public key to be added to the nodes SSH keychain. If Allow is true this value
		// defaults to "~/.ssh/id_rsa.pub", otherwise the value is ignored.
		PublicKeyPath string `json:"publicKeyPath,omitempty"`
		// +optional Public key to be added to the nodes SSH keychain. If Allow is false this value is ignored.
		PublicKey string `json:"publicKey,omitempty"`
		// +optional Public key name in EC2 to be added to the nodes SSH keychain. If Allow is false this value
		// is ignored.
		PublicKeyName string `json:"publicKeyName,omitempty"`
		// +optional
		SourceSecurityGroupIDs []string `json:"sourceSecurityGroupIds,omitempty"`
		// Enables the ability to [SSH onto nodes using SSM](/introduction#ssh-access)
		// +optional
		EnableSSM *bool `json:"enableSsm,omitempty"`
	}

	// NodeGroupInstancesDistribution holds the configuration for [spot
	// instances](/usage/spot-instances/)
	NodeGroupInstancesDistribution struct {
		// +required
		InstanceTypes []string `json:"instanceTypes,omitempty"`
		// Defaults to `on demand price`
		// +optional
		MaxPrice *float64 `json:"maxPrice,omitempty"`
		// Defaults to `0`
		// +optional
		OnDemandBaseCapacity *int `json:"onDemandBaseCapacity,omitempty"`
		// Range [0-100]
		// Defaults to `100`
		// +optional
		OnDemandPercentageAboveBaseCapacity *int `json:"onDemandPercentageAboveBaseCapacity,omitempty"`
		// Range [1-20]
		// Defaults to `2`
		// +optional
		SpotInstancePools *int `json:"spotInstancePools,omitempty"`
		// +optional
		SpotAllocationStrategy string `json:"spotAllocationStrategy,omitempty"`
		// Enable [capacity
		// rebalancing](https://docs.aws.amazon.com/autoscaling/ec2/userguide/capacity-rebalance.html)
		// for spot instances
		// +optional
		CapacityRebalance bool `json:"capacityRebalance"`
	}

	// NodeGroupBottlerocket holds the configuration for Bottlerocket based
	// NodeGroups.
	NodeGroupBottlerocket struct {
		// +optional
		EnableAdminContainer *bool `json:"enableAdminContainer,omitempty"`
		// Settings contains any [bottlerocket
		// settings](https://github.com/bottlerocket-os/bottlerocket/#description-of-settings)
		// +optional
		Settings *InlineDocument `json:"settings,omitempty"`
	}

	// NodeGroupUpdateConfig contains the configuration for updating NodeGroups.
	NodeGroupUpdateConfig struct {
		// MaxUnavailable sets the max number of nodes that can become unavailable
		// when updating a nodegroup (specified as number)
		// +optional
		MaxUnavailable *int `json:"maxUnavailable,omitempty"`

		// MaxUnavailablePercentage sets the max number of nodes that can become unavailable
		// when updating a nodegroup (specified as percentage)
		// +optional
		MaxUnavailablePercentage *int `json:"maxUnavailablePercentage,omitempty"`
	}
)
type ManagedNodeGroup struct {
	*NodeGroupBase

	// InstanceTypes specifies a list of instance types
	InstanceTypes []string `json:"instanceTypes,omitempty"`

	// Spot creates a spot nodegroup
	Spot bool `json:"spot,omitempty"`

	// Taints taints to apply to the nodegroup
	Taints []NodeGroupTaint `json:"taints,omitempty"`

	// UpdateConfig configures how to update NodeGroups.
	// +optional
	UpdateConfig *NodeGroupUpdateConfig `json:"updateConfig,omitempty"`

	// LaunchTemplate specifies an existing launch template to use
	// for the nodegroup
	LaunchTemplate *LaunchTemplate `json:"launchTemplate,omitempty"`

	// ReleaseVersion the AMI version of the EKS optimized AMI to use
	ReleaseVersion string `json:"releaseVersion"`

	// Internal fields

	Unowned bool `json:"-"`
}
type LaunchTemplate struct {
	// Launch template ID
	// +required
	ID string `json:"id,omitempty"`
	// Launch template version
	// Defaults to the default launch template version
	// TODO support $Default, $Latest
	Version string `json:"version,omitempty"`
}
type FargateProfile struct {

	// Name of the Fargate profile.
	// +required
	Name string `json:"name"`

	// PodExecutionRoleARN is the IAM role's ARN to use to run pods onto Fargate.
	PodExecutionRoleARN string `json:"podExecutionRoleARN,omitempty"`

	// Selectors define the rules to select workload to schedule onto Fargate.
	Selectors []FargateProfileSelector `json:"selectors"`

	// Subnets which Fargate should use to do network placement of the selected workload.
	// If none provided, all subnets for the cluster will be used.
	// +optional
	Subnets []string `json:"subnets,omitempty"`

	// Used to tag the AWS resources
	// +optional
	Tags map[string]string `json:"tags,omitempty"`

	// The current status of the Fargate profile.
	Status string `json:"status"`
}

// FargateProfileSelector defines rules to select workload to schedule onto Fargate.
type FargateProfileSelector struct {

	// Namespace is the Kubernetes namespace from which to select workload.
	// +required
	Namespace string `json:"namespace"`

	// Labels are the Kubernetes label selectors to use to select workload.
	// +optional
	Labels map[string]string `json:"labels,omitempty"`
}

/*
// EKSClusterVPC struct -> cfg.vpc
type EKSClusterVPC struct {
	CIDR             string               `json:"cidr,omitempty" yaml:"cidr"`
	Subnets          *EKSClusterSubnets   `json:"subnets,omitempty"`
	NAT              *EKSClusterNAT       `json:"nat,omitempty"`
	ClusterEndpoints *EKSClusterEndpoints `json:"clusterEndpoints,omitempty"`
}

// EKSClusterNAT struct -> cfg.vpc.nat
type EKSClusterNAT struct {
	Gateway string `json:"gateway,omitempty"`
}
*/
// EKSClusterSubnets struct -> cfg.vpc.subnets
type EKSClusterSubnets struct {
	Private map[string]EKSAZSubnetSpec `json:"private,omitempty"`
	Public  map[string]EKSAZSubnetSpec `json:"public,omitempty"`
}

// EKSAZSubnetSpec struct -> cfg.vpc.subnets.(private|public)[randomKey]
type EKSAZSubnetSpec struct {
	ID string `json:"id,omitempty"`
}

// EKSClusterEndpoints struct -> cfg.vpc.clusterEndpoints
type EKSClusterEndpoints struct {
	PrivateAccess *bool `json:"privateAccess"`
	PublicAccess  *bool `json:"publicAccess"`
}

// EKSScalingConfig struct -> embedded into NodeGroupBase
type EKSScalingConfig struct {
	DesiredCapacity *int64 `json:"desiredCapacity,omitempty"`
	MinSize         *int64 `json:"minSize,omitempty"`
	MaxSize         *int64 `json:"maxSize,omitempty"`
}

// EKSNodeGroupBase struct -> embedded into cfg.nodeGroups[]
type EKSNodeGroupBase struct {
	Name              string   `json:"name"`
	AMIFamily         string   `json:"amiFamily,omitempty"`
	InstanceType      string   `json:"instanceType,omitempty"`
	AvailabilityZones []string `json:"availabilityZones,omitempty"`
	*EKSScalingConfig
	VolumeSize           *int64            `json:"volumeSize,omitempty"`
	SSH                  *EKSNodeGroupSSH  `json:"ssh,omitempty"`
	Labels               map[string]string `json:"labels,omitempty"`
	PrivateNetworking    *bool             `json:"privateNetworking,omitempty"`
	Tags                 map[string]string `json:"tags,omitempty"`
	IAM                  *EKSNodeGroupIAM  `json:"iam,omitempty"`
	AMI                  string            `json:"ami,omitempty"`
	MaxPodsPerNode       *int64            `json:"maxPodsPerNode,omitempty"`
	SecurityGroups       *EKSNodeGroupSGs  `json:"securityGroups,omitempty"`
	VolumeType           string            `json:"volumeType,omitempty"`
	VolumeEncrypted      *bool             `json:"volumeEncrypted,omitempty"`
	VolumeKmsKeyID       string            `json:"volumeKmsKeyID,omitempty"`
	PreBootstrapCommands []string          `json:"preBootstrapCommands,omitempty"`
	Subnets              []string          `json:"subnets,omitempty"`
}

// EKSNodeGroup struct -> cfg.nodeGroups[]
type EKSNodeGroup struct {
	*EKSNodeGroupBase
	InstancesDistribution *EKSNodeGroupInstancesDistribution `json:"instancesDistribution,omitempty"`
	SubnetCidr            string                             `json:"subnetCidr,omitempty"`
}

// EKSManagedNodeGroup struct -> cfg.
type EKSManagedNodeGroup struct {
	*EKSNodeGroupBase
	InstanceTypes []string `json:"instanceTypes,omitempty"`
	Spot          *bool    `json:"spot,omitempty"`
}

// EKSNodeGroupIAM struct -> cfg.nodeGroups[].iam
type EKSNodeGroupIAM struct {
	InstanceProfileARN              string                        `json:"instanceProfileARN,omitempty"`
	InstanceRoleARN                 string                        `json:"instanceRoleARN,omitempty"`
	InstanceRolePermissionsBoundary string                        `json:"instanceRolePermissionsBoundary,omitempty"`
	WithAddonPolicies               *EKSNodeGroupIAMAddonPolicies `json:"withAddonPolicies,omitempty"`
}

// EKSNodeGroupIAMAddonPolicies struct -> cfg.nodeGroups[].iam.withAddonPolicies
type EKSNodeGroupIAMAddonPolicies struct {
	ImageBuilder              *bool `json:"imageBuilder,omitempty"`
	AutoScaler                *bool `json:"autoScaler,omitempty"`
	ExternalDNS               *bool `json:"externalDNS,omitempty"`
	AppMesh                   *bool `json:"appMesh,omitempty"`
	AWSLoadBalancerController *bool `json:"albIngress,omitempty"`
	EFS                       *bool `json:"efs,omitempty"`
}

// Make sure to update hasAtLeastOneEnabled()

// EKSNodeGroupSGs struct -> cfg.nodeGroups[].SecurityGroups
type EKSNodeGroupSGs struct {
	AttachIDs []string `json:"attachIDs,omitempty"`
}

// EKSNodeGroupSSH struct -> cfg.nodeGroups[].ssh
type EKSNodeGroupSSH struct {
	Allow         *bool  `json:"allow"`
	PublicKeyName string `json:"publicKeyName,omitempty"`
}

// EKSNodeGroupInstancesDistribution struct -> cfg.nodeGroups[].instancesDistribution
type EKSNodeGroupInstancesDistribution struct {
	InstanceTypes                       []string `json:"instanceTypes,omitempty"`
	MaxPrice                            *float64 `json:"maxPrice,omitempty"`
	OnDemandBaseCapacity                *int64   `json:"onDemandBaseCapacity,omitempty"`
	OnDemandPercentageAboveBaseCapacity *int64   `json:"onDemandPercentageAboveBaseCapacity,omitempty"`
	SpotInstancePools                   *int64   `json:"spotInstancePools,omitempty"`
	SpotAllocationStrategy              string   `json:"spotAllocationStrategy,omitempty"`
}

type EKSClusterCloudWatch struct {
	//+optional
	ClusterLogging *EKSClusterCloudWatchLogging `json:"clusterLogging,omitempty"`
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
	EnableTypes []string `json:"enableTypes"`
}

// SecretsEncryption defines the configuration for KMS encryption provider
type SecretsEncryption struct {
	// +required
	KeyARN string `json:"keyARN,omitempty"`
}
