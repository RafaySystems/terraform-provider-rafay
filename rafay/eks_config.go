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
type EKSCluster struct {
	Kind     string              `yaml:"kind,omitempty"`
	Metadata *EKSClusterMetadata `yaml:"metadata,omitempty"`
	Spec     *EKSSpec            `yaml:"spec,omitempty"`
}

type EKSSpec struct {
	Type             string            `yaml:"type,omitempty"`
	Blueprint        string            `yaml:"blueprint,omitempty"`
	BlueprintVersion string            `yaml:"blueprintversion,omitempty"`
	CloudProvider    string            `yaml:"cloudprovider,omitempty"`
	CniProvider      string            `yaml:"cniprovider,omitempty"`
	ProxyConfig      map[string]string `yaml:"labels,omitempty"`
}

type EKSClusterMetadata struct {
	Name    string            `yaml:"name,omitempty"`
	Project string            `yaml:"project,omitempty"`
	Labels  map[string]string `yaml:"labels,omitempty"`
}

//struct for eks cluster config sped (second part of the yaml file kind:clusterConfig)
type KubernetesNetworkConfig struct {
	ServiceIPv4CIDR string `yaml:"serviceIPv4CIDR,omitempty"`
}

type EKSClusterConfig struct {
	APIVersion              string                    `yaml:"apiversion,omitempty"`
	Kind                    string                    `yaml:"kind,omitempty"`
	Metadata                *EKSClusterConfigMetadata `yaml:"metadata,omitempty"`
	KubernetesNetworkConfig *KubernetesNetworkConfig  `yaml:"kubernetesNetworkConfig,omitempty"`
	IAM                     *EKSClusterIAM            `yaml:"iam,omitempty,omitempty"`
	IdentityProviders       []IdentityProvider        `yaml:"identityProviders,omitempty"`
	VPC                     *EKSClusterVPC            `yaml:"vpc,omitempty"`
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
	ServiceRoleARN string `yaml:"serviceRoleARN,omitempty"`

	// permissions boundary for all identity-based entities created by eksctl.
	// See [AWS Permission Boundary](https://docs.aws.amazon.com/IAM/latest/UserGuide/access_policies_boundaries.html)
	// +optional
	ServiceRolePermissionsBoundary string `yaml:"serviceRolePermissionsBoundary,omitempty"`

	// role used by pods to access AWS APIs. This role is added to the Kubernetes RBAC for authorization.
	// See [Pod Execution Role](https://docs.aws.amazon.com/eks/latest/userguide/pod-execution-role.html)
	// +optional
	FargatePodExecutionRoleARN string `yaml:"fargatePodExecutionRoleARN,omitempty"`

	// permissions boundary for the fargate pod execution role`. See [EKS Fargate Support](/usage/fargate-support/)
	// +optional
	FargatePodExecutionRolePermissionsBoundary string `yaml:"fargatePodExecutionRolePermissionsBoundary,omitempty"`

	// enables the IAM OIDC provider as well as IRSA for the Amazon CNI plugin
	// +optional
	WithOIDC bool `yaml:"withOIDC,omitempty"`

	// service accounts to create in the cluster.
	// See [IAM Service Accounts](/iamserviceaccounts/#usage-with-config-files)
	// +optional
	ServiceAccounts []*EKSClusterIAMServiceAccount `yaml:"serviceAccounts,omitempty"`

	// VPCResourceControllerPolicy attaches the IAM policy
	// necessary to run the VPC controller in the control plane
	// Defaults to `true`
	VPCResourceControllerPolicy bool `yaml:"vpcResourceControllerPolicy,omitempty"`
}

// ClusterIAMServiceAccount holds an IAM service account metadata and configuration
type EKSClusterIAMServiceAccount struct {
	EKSClusterIAMMeta `yaml:"metadata,omitempty"`

	// list of ARNs of the IAM policies to attach
	// +optional
	AttachPolicyARNs []string `yaml:"attachPolicyARNs,omitempty"`

	WellKnownPolicies WellKnownPolicies `yaml:"wellKnownPolicies,omitempty"`

	// AttachPolicy holds a policy document to attach to this service account
	// +optional
	//AttachPolicy map[string]string `yaml:"attachPolicy,omitempty"`
	AttachPolicy InlineDocument `yaml:"attachPolicy,omitempty"`

	// ARN of the role to attach to the service account
	AttachRoleARN string `yaml:"attachRoleARN,omitempty"`

	// ARN of the permissions boundary to associate with the service account
	// +optional
	PermissionsBoundary string `yaml:"permissionsBoundary,omitempty"`

	// +optional
	Status *ClusterIAMServiceAccountStatus `yaml:"status,omitempty"`

	// Specific role name instead of the Cloudformation-generated role name
	// +optional
	RoleName string `yaml:"roleName,omitempty"`

	// Specify if only the IAM Service Account role should be created without creating/annotating the service account
	// +optional
	RoleOnly *bool `yaml:"roleOnly,omitempty"`

	// AWS tags for the service account
	// +optional
	Tags map[string]string `yaml:"tags,omitempty"`
}

type WellKnownPolicies struct {
	// ImageBuilder allows for full ECR (Elastic Container Registry) access.
	ImageBuilder *bool `yaml:"imageBuilder,inline,omitempty"`
	// AutoScaler adds policies for cluster-autoscaler. See [autoscaler AWS
	// docs](https://docs.aws.amazon.com/eks/latest/userguide/cluster-autoscaler.html).
	AutoScaler *bool `yaml:"autoScaler,inline"`
	// AWSLoadBalancerController adds policies for using the
	// aws-load-balancer-controller. See [Load Balancer
	// docs](https://docs.aws.amazon.com/eks/latest/userguide/aws-load-balancer-controller.html).
	AWSLoadBalancerController *bool `yaml:"awsLoadBalancerController,inline"`
	// ExternalDNS adds external-dns policies for Amazon Route 53.
	// See [external-dns
	// docs](https://github.com/kubernetes-sigs/external-dns/blob/master/docs/tutorials/aws.md).
	ExternalDNS *bool `yaml:"externalDNS,inline"`
	// CertManager adds cert-manager policies. See [cert-manager
	// docs](https://cert-manager.io/docs/configuration/acme/dns01/route53).
	CertManager *bool `yaml:"certManager,inline"`
	// EBSCSIController adds policies for using the
	// ebs-csi-controller. See [aws-ebs-csi-driver
	// docs](https://github.com/kubernetes-sigs/aws-ebs-csi-driver#set-up-driver-permission).
	EBSCSIController *bool `yaml:"ebsCSIController,inline"`
	// EFSCSIController adds policies for using the
	// efs-csi-controller. See [aws-efs-csi-driver
	// docs](https://aws.amazon.com/blogs/containers/introducing-efs-csi-dynamic-provisioning).
	EFSCSIController *bool `yaml:"efsCSIController,inline"`
}

type InlineDocument map[string]interface{}

type ClusterIAMServiceAccountStatus struct {
	// +optional
	RoleARN string `yaml:"roleARN,omitempty"`
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
		SecurityGroup string `yaml:"securityGroup,omitempty"`
		// Subnets are keyed by AZ for convenience.
		// See [this example](/examples/reusing-iam-and-vpc/)
		// as well as [using existing
		// VPCs](/usage/vpc-networking/#use-existing-vpc-other-custom-configuration).
		// +optional
		Subnets *ClusterSubnets `yaml:"subnets,omitempty"`
		// for additional CIDR associations, e.g. a CIDR for
		// private subnets or any ad-hoc subnets
		// +optional
		ExtraCIDRs []string `yaml:"extraCIDRs,omitempty"`
		// for additional IPv6 CIDR associations, e.g. a CIDR for
		// private subnets or any ad-hoc subnets
		// +optional
		ExtraIPv6CIDRs []string `yaml:"extraIPv6CIDRs,omitempty"`
		// for pre-defined shared node SG
		SharedNodeSecurityGroup string `yaml:"sharedNodeSecurityGroup,omitempty"`
		// Automatically add security group rules to and from the default
		// cluster security group and the shared node security group.
		// This allows unmanaged nodes to communicate with the control plane
		// and managed nodes.
		// This option cannot be disabled when using eksctl created security groups.
		// Defaults to `true`
		// +optional
		ManageSharedNodeSecurityGroupRules *bool `yaml:"manageSharedNodeSecurityGroupRules,omitempty"`
		// AutoAllocateIPV6 requests an IPv6 CIDR block with /56 prefix for the VPC
		// +optional
		AutoAllocateIPv6 *bool `yaml:"autoAllocateIPv6,omitempty"`
		// +optional
		NAT *ClusterNAT `yaml:"nat,omitempty"`
		// See [managing access to API](/usage/vpc-networking/#managing-access-to-the-kubernetes-api-server-endpoints)
		// +optional
		ClusterEndpoints *ClusterEndpoints `yaml:"clusterEndpoints,omitempty"`
		// PublicAccessCIDRs are which CIDR blocks to allow access to public
		// k8s API endpoint
		// +optional
		PublicAccessCIDRs []string `yaml:"publicAccessCIDRs,omitempty"`
	}
	// ClusterSubnets holds private and public subnets
	ClusterSubnets struct {
		Private AZSubnetMapping `yaml:"private,omitempty"`
		Public  AZSubnetMapping `yaml:"public,omitempty"`
	}
	// SubnetTopology can be SubnetTopologyPrivate or SubnetTopologyPublic
	SubnetTopology string
	AZSubnetSpec   struct {
		// +optional
		ID string `yaml:"id,omitempty"`
		// AZ can be omitted if the key is an AZ
		// +optional
		AZ string `yaml:"az,omitempty"`
		// +optional
		//can i just make this a string?
		//CIDR string `yaml:"cidr"`
		CIDR *ipnet.IPNet `yaml:"cidr,omitempty"`
	}
	// Network holds ID and CIDR
	Network struct {
		// +optional
		ID string `yaml:"id,omitempty"`
		// +optional
		//can i just make this a string?
		CIDR string `yaml:"cidr"`
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
	ServiceAccountRoleARN string `yaml:"serviceAccountRoleARN,omitempty"`
	// list of ARNs of the IAM policies to attach
	// +optional
	AttachPolicyARNs []string `yaml:"attachPolicyARNs,omitempty"`
	// AttachPolicy holds a policy document to attach
	// +optional
	AttachPolicy InlineDocument `yaml:"attachPolicy,omitempty"`
	// ARN of the permissions' boundary to associate
	// +optional
	PermissionsBoundary string `yaml:"permissionsBoundary,omitempty"`
	// WellKnownPolicies for attaching common IAM policies
	//WellKnown Policies not in documentation for addon? (same field as IAM wellknow-policies)
	//WellKnownPolicies WellKnownPolicies `yaml:"wellKnownPolicies,omitempty"`
	// The metadata to apply to the cluster to assist with categorization and organization.
	// Each tag consists of a key and an optional value, both of which you define.
	// +optional
	Tags map[string]string `yaml:"tags,omitempty"`
	// Force applies the add-on to overwrite an existing add-on
	Force bool `yaml:"-"`
}

// PrivateCluster defines the configuration for a fully-private cluster
type PrivateCluster struct {

	// Enabled enables creation of a fully-private cluster
	Enabled bool `yaml:"enabled"`

	// SkipEndpointCreation skips the creation process for endpoints completely. This is only used in case of an already
	// provided VPC and if the user decided to set it to true.
	SkipEndpointCreation bool `yaml:"skipEndpointCreation"`

	// AdditionalEndpointServices specifies additional endpoint services that
	// must be enabled for private access.
	// Valid entries are `AdditionalEndpointServices` constants
	AdditionalEndpointServices []string `yaml:"additionalEndpointServices,omitempty"`
}
type NodeGroup struct {
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
	PrivateNetworking bool `yaml:"privateNetworking"`
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
	MaxPodsPerNode int `yaml:"maxPodsPerNode,omitempty"`

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
	CustomAMI bool `yaml:"-"`

	// Enable EC2 detailed monitoring
	// +optional
	EnableDetailedMonitoring *bool `yaml:"enableDetailedMonitoring,omitempty"`
	// Rafay changes - start
	// Internal
	IsWavelengthZone bool `yaml:"-"`
	// Rafay changes - end

	//+optional
	InstancesDistribution *NodeGroupInstancesDistribution `yaml:"instancesDistribution,omitempty"`

	// +optional
	ASGMetricsCollection []MetricsCollection `yaml:"asgMetricsCollection,omitempty"`

	// CPUCredits configures [T3 Unlimited](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/burstable-performance-instances-unlimited-mode.html), valid only for T-type instances
	// +optional
	CPUCredits string `yaml:"cpuCredits,omitempty"`

	// Associate load balancers with auto scaling group
	// +optional
	ClassicLoadBalancerNames []string `yaml:"classicLoadBalancerNames,omitempty"`

	// Associate target group with auto scaling group
	// +optional
	TargetGroupARNs []string `yaml:"targetGroupARNs,omitempty"`

	// Taints taints to apply to the nodegroup
	// +optional
	Taints taintsWrapper `yaml:"taints,omitempty"`

	// UpdateConfig configures how to update NodeGroups.
	// +optional
	UpdateConfig *NodeGroupUpdateConfig `yaml:"updateConfig,omitempty"`

	// [Custom
	// address](/usage/vpc-networking/#custom-cluster-dns-address) used for DNS
	// lookups
	// +optional
	ClusterDNS string `yaml:"clusterDNS,omitempty"`

	// [Customize `kubelet` config](/usage/customizing-the-kubelet/)
	// +optional

	KubeletExtraConfig *InlineDocument `yaml:"kubeletExtraConfig,omitempty"`

	// ContainerRuntime defines the runtime (CRI) to use for containers on the node
	// +optional
	ContainerRuntime string `yaml:"containerRuntime,omitempty"`
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
	PrivateNetworking bool `yaml:"privateNetworking"`
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
	MaxPodsPerNode int `yaml:"maxPodsPerNode,omitempty"`

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
	CustomAMI bool `yaml:"-"`

	// Enable EC2 detailed monitoring
	// +optional
	EnableDetailedMonitoring *bool `yaml:"enableDetailedMonitoring,omitempty"`
	// Rafay changes - start
	// Internal
	IsWavelengthZone bool `yaml:"-"`
	// Rafay changes - end
}
type InstanceSelector struct {
	// VCPUs specifies the number of vCPUs
	VCPUs int `yaml:"vCPUs,omitempty"`
	// Memory specifies the memory
	// The unit defaults to GiB
	Memory string `yaml:"memory,omitempty"`
	// GPUs specifies the number of GPUs.
	// It can be set to 0 to select non-GPU instance types.
	GPUs int `yaml:"gpus,omitempty"`
	// CPU Architecture of the EC2 instance type.
	// Valid variants are:
	// `"x86_64"`
	// `"amd64"`
	// `"arm64"`
	CPUArchitecture string `yaml:"cpuArchitecture,omitempty"`
}
type Placement struct {
	GroupName string `yaml:"groupName,omitempty"`
}
type ScalingConfig struct {
	// +optional
	DesiredCapacity *int `yaml:"desiredCapacity,omitempty"`
	// +optional
	MinSize *int `yaml:"minSize,omitempty"`
	// +optional
	MaxSize *int `yaml:"maxSize,omitempty"`
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
		AttachIDs []string `yaml:"attachIDs,omitempty"`
		// WithShared attach the security group
		// shared among all nodegroups in the cluster
		// Defaults to `true`
		// +optional
		WithShared *bool `yaml:"withShared"`
		// WithLocal attach a security group
		// local to this nodegroup
		// Not supported for managed nodegroups
		// Defaults to `true`
		// +optional
		WithLocal *bool `yaml:"withLocal"`
	}
	// NodeGroupIAM holds all IAM attributes of a NodeGroup
	NodeGroupIAM struct {
		// AttachPolicy holds a policy document to attach
		// +optional
		AttachPolicy InlineDocument `yaml:"attachPolicy,omitempty"`
		// list of ARNs of the IAM policies to attach
		// +optional
		AttachPolicyARNs []string `yaml:"attachPolicyARNs,omitempty"`
		// +optional
		InstanceProfileARN string `yaml:"instanceProfileARN,omitempty"`
		// +optional
		InstanceRoleARN string `yaml:"instanceRoleARN,omitempty"`
		// +optional
		InstanceRoleName string `yaml:"instanceRoleName,omitempty"`
		// +optional
		InstanceRolePermissionsBoundary string `yaml:"instanceRolePermissionsBoundary,omitempty"`
		// +optional
		WithAddonPolicies NodeGroupIAMAddonPolicies `yaml:"withAddonPolicies,omitempty"`
	}
	// NodeGroupIAMAddonPolicies holds all IAM addon policies
	NodeGroupIAMAddonPolicies struct {
		// +optional
		// ImageBuilder allows for full ECR (Elastic Container Registry) access. This is useful for building, for
		// example, a CI server that needs to push images to ECR
		ImageBuilder *bool `yaml:"imageBuilder"`
		// +optional
		// AutoScaler enables IAM policy for cluster-autoscaler
		AutoScaler *bool `yaml:"autoScaler"`
		// +optional
		// ExternalDNS adds the external-dns project policies for Amazon Route 53
		ExternalDNS *bool `yaml:"externalDNS"`
		// +optional
		// CertManager enables the ability to add records to Route 53 in order to solve the DNS01 challenge. More information can be found
		// [here](https://cert-manager.io/docs/configuration/acme/dns01/route53/#set-up-a-iam-role)
		CertManager *bool `yaml:"certManager"`
		// +optional
		// AppMesh enables full access to AppMesh
		AppMesh *bool `yaml:"appMesh"`
		// +optional
		// AppMeshPreview enables full access to AppMesh Preview
		AppMeshPreview *bool `yaml:"appMeshPreview"`
		// +optional
		// EBS enables the new EBS CSI (Elastic Block Store Container Storage Interface) driver
		EBS *bool `yaml:"ebs"`
		// +optional
		FSX *bool `yaml:"fsx"`
		// +optional
		EFS *bool `yaml:"efs"`
		// +optional
		AWSLoadBalancerController *bool `yaml:"albIngress"`
		// +optional
		XRay *bool `yaml:"xRay"`
		// +optional
		CloudWatch *bool `yaml:"cloudWatch"`
	}

	// NodeGroupSSH holds all the ssh access configuration to a NodeGroup
	NodeGroupSSH struct {
		// +optional If Allow is true the SSH configuration provided is used, otherwise it is ignored. Only one of
		// PublicKeyPath, PublicKey and PublicKeyName can be configured
		Allow *bool `yaml:"allow"`
		// +optional The path to the SSH public key to be added to the nodes SSH keychain. If Allow is true this value
		// defaults to "~/.ssh/id_rsa.pub", otherwise the value is ignored.
		PublicKeyPath string `yaml:"publicKeyPath,omitempty"`
		// +optional Public key to be added to the nodes SSH keychain. If Allow is false this value is ignored.
		PublicKey string `yaml:"publicKey,omitempty"`
		// +optional Public key name in EC2 to be added to the nodes SSH keychain. If Allow is false this value
		// is ignored.
		PublicKeyName string `yaml:"publicKeyName,omitempty"`
		// +optional
		SourceSecurityGroupIDs []string `yaml:"sourceSecurityGroupIds,omitempty"`
		// Enables the ability to [SSH onto nodes using SSM](/introduction#ssh-access)
		// +optional
		EnableSSM *bool `yaml:"enableSsm,omitempty"`
	}

	// NodeGroupInstancesDistribution holds the configuration for [spot
	// instances](/usage/spot-instances/)
	NodeGroupInstancesDistribution struct {
		// +required
		InstanceTypes []string `yaml:"instanceTypes,omitempty"`
		// Defaults to `on demand price`
		// +optional
		MaxPrice *float64 `yaml:"maxPrice,omitempty"`
		// Defaults to `0`
		// +optional
		OnDemandBaseCapacity *int `yaml:"onDemandBaseCapacity,omitempty"`
		// Range [0-100]
		// Defaults to `100`
		// +optional
		OnDemandPercentageAboveBaseCapacity *int `yaml:"onDemandPercentageAboveBaseCapacity,omitempty"`
		// Range [1-20]
		// Defaults to `2`
		// +optional
		SpotInstancePools *int `yaml:"spotInstancePools,omitempty"`
		// +optional
		SpotAllocationStrategy string `yaml:"spotAllocationStrategy,omitempty"`
		// Enable [capacity
		// rebalancing](https://docs.aws.amazon.com/autoscaling/ec2/userguide/capacity-rebalance.html)
		// for spot instances
		// +optional
		CapacityRebalance bool `yaml:"capacityRebalance"`
	}

	// NodeGroupBottlerocket holds the configuration for Bottlerocket based
	// NodeGroups.
	NodeGroupBottlerocket struct {
		// +optional
		EnableAdminContainer *bool `yaml:"enableAdminContainer,omitempty"`
		// Settings contains any [bottlerocket
		// settings](https://github.com/bottlerocket-os/bottlerocket/#description-of-settings)
		// +optional
		Settings *InlineDocument `yaml:"settings,omitempty"`
	}

	// NodeGroupUpdateConfig contains the configuration for updating NodeGroups.
	NodeGroupUpdateConfig struct {
		// MaxUnavailable sets the max number of nodes that can become unavailable
		// when updating a nodegroup (specified as number)
		// +optional
		MaxUnavailable *int `yaml:"maxUnavailable,omitempty"`

		// MaxUnavailablePercentage sets the max number of nodes that can become unavailable
		// when updating a nodegroup (specified as percentage)
		// +optional
		MaxUnavailablePercentage *int `yaml:"maxUnavailablePercentage,omitempty"`
	}
)
type ManagedNodeGroup struct {
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
	PrivateNetworking bool `yaml:"privateNetworking"`
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
	MaxPodsPerNode int `yaml:"maxPodsPerNode,omitempty"`

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
	CustomAMI bool `yaml:"-"`

	// Enable EC2 detailed monitoring
	// +optional
	EnableDetailedMonitoring *bool `yaml:"enableDetailedMonitoring,omitempty"`
	// Rafay changes - start
	// Internal
	IsWavelengthZone bool `yaml:"-"`
	// Rafay changes - end

	// InstanceTypes specifies a list of instance types
	InstanceTypes []string `yaml:"instanceTypes,omitempty"`

	// Spot creates a spot nodegroup
	Spot bool `yaml:"spot,omitempty"`

	// Taints taints to apply to the nodegroup
	Taints []NodeGroupTaint `yaml:"taints,omitempty"`

	// UpdateConfig configures how to update NodeGroups.
	// +optional
	UpdateConfig *NodeGroupUpdateConfig `yaml:"updateConfig,omitempty"`

	// LaunchTemplate specifies an existing launch template to use
	// for the nodegroup
	LaunchTemplate *LaunchTemplate `yaml:"launchTemplate,omitempty"`

	// ReleaseVersion the AMI version of the EKS optimized AMI to use
	ReleaseVersion string `yaml:"releaseVersion,omitempty"`

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
	PodExecutionRoleARN string `yaml:"podExecutionRoleARN,omitempty"`

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
	ClusterLogging *EKSClusterCloudWatchLogging `yaml:"clusterLogging,omitempty"`
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
	EnableTypes []string `yaml:"enableTypes,omitempty"`
}

// SecretsEncryption defines the configuration for KMS encryption provider
type SecretsEncryption struct {
	// +required
	KeyARN string `yaml:"keyARN,omitempty"`
}
