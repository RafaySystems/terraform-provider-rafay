package rafay

const AKSClusterAPIVersion = "rafay.io/v1alpha1"
const AKSClusterKind = "Cluster"

const AKSClusterConfigAPIVersion = "rafay.io/v1alpha1"
const AKSClusterConfigKind = "aksClusterConfig"

const AKSManagedClusterType = "Microsoft.ContainerService/managedClusters"
const AKSNodePoolType = "Microsoft.ContainerService/managedClusters/agentPools"
const AKSManagedClusterAPIVersion = "2021-05-01"
const AKSManagedClusterNetworkProfileLoadBalancerSKU = "standard"

const AzureVnetRoleAssignmentType = "Microsoft.Network/virtualNetworks/subnets/providers/roleAssignments"
const AzureVnetRoleAssignmentAPIVersion = "2018-09-01-preview"

const AzureOMSRoleAssignmentType = "Microsoft.ContainerService/managedClusters/providers/roleAssignments"
const AzureOMSRoleAssignmentAPIVersion = "2018-01-01-preview"

const AzureACRRoleAssignmentType = "Microsoft.ContainerRegistry/registries/providers/roleAssignments"
const AzureACRRoleAssignmentAPIVersion = "2018-09-01-preview"

const AzureOperationsManagementSolutionType = "Microsoft.OperationsManagement/solutions"
const AzureOperationsManagementSolutionAPIVersion = "2015-11-01-preview"

const AzureRoleIDNetworkContributor = "4d97b98b-1d4f-4787-a291-c67834d212e7"
const AzureRoleIDMonitoringMetricsPublisher = "3913510d-42f4-4e42-8a64-420c390055eb"
const AzureRoleIDContributor = "b24988ac-6180-42a0-ab88-20f7382dd24c"

type AKSCluster struct {
	APIVersion string              `yaml:"apiversion,omitempty"`
	Kind       string              `yaml:"kind,omitempty"`
	Metadata   *AKSClusterMetadata `yaml:"metadata,omitempty"`
	Spec       *AKSClusterSpec     `yaml:"spec,omitempty"`
}

type AKSClusterMetadata struct {
	Name    string            `yaml:"name,omitempty"`
	Project string            `yaml:"project,omitempty"`
	Labels  map[string]string `yaml:"labels,omitempty,omitempty"`
}

type AKSClusterSpec struct {
	Type             string            `yaml:"type,omitempty"`
	Blueprint        string            `yaml:"blueprint,omitempty"`
	BlueprintVersion string            `yaml:"blueprintversion,omitempty"`
	CloudProvider    string            `yaml:"cloudprovider,omitempty"`
	AKSClusterConfig *AKSClusterConfig `yaml:"clusterConfig,omitempty"`
}

type AzureRafayMetadata struct {
	SubscriptionID    string `yaml:"subscriptionId,omitempty"`
	ResourceGroupName string `yaml:"resourceGroupName,omitempty"`
}

type AKSClusterConfig struct {
	APIVersion string                    `yaml:"apiversion,omitempty"`
	Kind       string                    `yaml:"kind,omitempty"`
	Metadata   *AKSClusterConfigMetadata `yaml:"metadata,omitempty"`
	Spec       *AKSClusterConfigSpec     `yaml:"spec,omitempty"`
}

type AKSClusterConfigMetadata struct {
	Name string `yaml:"name,omitempty"`
}

type AKSClusterConfigSpec struct {
	SubscriptionID    string             `yaml:"subscriptionId,omitempty"`
	ResourceGroupName string             `yaml:"resourceGroupName,omitempty"`
	ManagedCluster    *AKSManagedCluster `yaml:"managedCluster,omitempty"`
	NodePools         []*AKSNodePool     `yaml:"nodePools,omitempty"`
	//Internal          *AKSRafayInternal  `yaml:"internal,omitempty"`
}

// type AzureContainerRegistryProfile struct {
// 	ResourceGroupName string `yaml:"resourceGroupName,omitempty"`
// 	RegistryName      string `yaml:"acrName"`
// }

type AKSManagedCluster struct {
	ExtendedLocation   *AKSClusterExtendedLocation          `yaml:"extendedLocation,omitempty"`
	Type               string                               `yaml:"type,omitempty"`
	APIVersion         string                               `yaml:"apiversion,omitempty"`
	Location           string                               `yaml:"location,omitempty"`
	Identity           *AKSManagedClusterIdentity           `yaml:"identity,omitempty"`
	Properties         *AKSManagedClusterProperties         `yaml:"properties,omitempty"`
	SKU                *AKSManagedClusterSKU                `yaml:"sku,omitempty"`
	Tags               map[string]string                    `yaml:"tags,omitempty"`
	AdditionalMetadata *AKSManagedClusterAdditionalMetadata `yaml:"additionalMetadata,omitempty"`
}

type AKSClusterExtendedLocation struct {
	Name string `yaml:"name,omitempty"`
	Type string `yaml:"type,omitempty"`
}

// type AzureRafayClusterMetadata struct {
// 	ACRProfile                  *AzureContainerRegistryProfile `yaml:"acrProfile,omitempty"`
// 	ServicePrincipalCredentials string                         `yaml:"servicePrincipalCredential,omitempty"`
// 	OMSWorkspaceLocation        string                         `yaml:"omsWorkspaceLocation,omitempty"`
// 	WindowsAdminCredentials     string                         `yaml:"windowsAdminCredentials,omitempty"`
// }

type AKSManagedClusterIdentity struct {
	Type string `yaml:"type,omitempty"`
	//UserAssignedIdentities map[string]string `yaml:"userAssignedIdentities,omitempty"`
	UserAssignedIdentities map[string]interface{} `yaml:"userAssignedIdentities,omitempty"`
}

type AKSManagedClusterProperties struct {
	KubernetesVersion       string                                   `yaml:"kubernetesVersion,omitempty"`
	EnableRBAC              *bool                                    `yaml:"enableRbac,omitempty"`
	FQDNSubdomain           string                                   `yaml:"fqdnSubdomain,omitempty"`
	DNSPrefix               string                                   `yaml:"dnsPrefix,omitempty"`
	EnablePodSecurityPolicy *bool                                    `yaml:"enablePodSecurityPolicy,omitempty"`
	NodeResourceGroup       string                                   `yaml:"nodeResourceGroup,omitempty"`
	NetworkProfile          *AKSManagedClusterNetworkProfile         `yaml:"networkProfile,omitempty"`
	AzureADProfile          *AKSManagedClusterAzureADProfile         `yaml:"aadProfile,omitempty"`
	APIServerAccessProfile  *AKSManagedClusterAPIServerAccessProfile `yaml:"apiServerAccessProfile,omitempty"`
	DisableLocalAccounts    *bool                                    `yaml:"disableLocalAccounts,omitempty"`
	DiskEncryptionSetID     string                                   `yaml:"diskEncryptionSetId,omitempty"`
	AddonProfiles           *AddonProfiles                           `yaml:"addonProfiles,omitempty"`
	//AddonProfiles           map[string]string                         `yaml:"addonProfiles,omitempty"`
	ServicePrincipalProfile *AKSManagedClusterServicePrincipalProfile `yaml:"servicePrincipalProfile,omitempty"`
	LinuxProfile            *AKSManagedClusterLinuxProfile            `yaml:"linuxProfile,omitempty"`
	WindowsProfile          *AKSManagedClusterWindowsProfile          `yaml:"windowsProfile,omitempty"`
	HTTPProxyConfig         *AKSManagedClusterHTTPProxyConfig         `yaml:"httpProxyConfig,omitempty"`
	IdentityProfile         map[string]string                         `yaml:"identityProfile,omitempty"`
	AutoScalerProfile       *AKSManagedClusterAutoScalerProfile       `yaml:"autoScalerProfile,omitempty"`
	AutoUpgradeProfile      *AKSManagedClusterAutoUpgradeProfile      `yaml:"autoUpgradeProfile,omitempty"`
	PodIdentityProfile      *AKSManagedClusterPodIdentityProfile      `yaml:"podIdentityProfile,omitempty"`
	PrivateLinkResources    *AKSManagedClusterPrivateLinkResources    `yaml:"privateLinkResources,omitempty"`
}

type AddonProfiles struct {
	HttpApplicationRouting *AKSManagedClusterAddonProfile `yaml:"httpApplicationRouting,omitempty"`
	AzurePolicy            *AKSManagedClusterAddonProfile `yaml:"azurePolicy,omitempty"`
	//OmsAgent               *AKSManagedClusterAddonProfile `yaml:"omsAgent,omitempty"`
}

type AKSManagedClusterAddonProfile struct {
	Enabled *bool                  `yaml:"enabled,omitempty"`
	Config  map[string]interface{} `yaml:"config,omitempty"`
}

type AKSManagedClusterNetworkProfile struct {
	LoadBalancerSKU     string                                  `yaml:"loadBalancerSku,omitempty"`
	NetworkPlugin       string                                  `yaml:"networkPlugin,omitempty"`
	NetworkPolicy       string                                  `yaml:"networkPolicy,omitempty"`
	DNSServiceIP        string                                  `yaml:"dnsServiceIp,omitempty"`
	DockerBridgeCidr    string                                  `yaml:"dockerBridgeCidr,omitempty"`
	LoadBalancerProfile *AKSManagedClusterNPLoadBalancerProfile `yaml:"loadBalancerProfile,omitempty"`
	NetworkMode         string                                  `yaml:"networkMode,omitempty"`
	OutboundType        string                                  `yaml:"outboundType,omitempty"`
	PodCidr             string                                  `yaml:"podCidr,omitempty"`
	ServiceCidr         string                                  `yaml:"serviceCidr,omitempty"`
}

type AKSManagedClusterNPLoadBalancerProfile struct {
	AllocatedOutboundPorts *int                                       `yaml:"allocatedOutboundPorts,omitempty"`
	EffectiveOutboundIPs   []*AKSManagedClusterNPEffectiveOutboundIPs `yaml:"effectiveOutboundIps,omitempty"`
	IdleTimeoutInMinutes   *int                                       `yaml:"idleTimeoutInMinutes,omitempty"`
	ManagedOutboundIPs     *AKSManagedClusterNPManagedOutboundIPs     `yaml:"managedOutboundIps,omitempty"`
	OutboundIPPrefixes     *AKSManagedClusterNPOutboundIPPrefixes     `yaml:"outboundIpPrefixes,omitempty"`
	OutboundIPs            *AKSManagedClusterNPOutboundIPs            `yaml:"outboundIps,omitempty"`
}

type AKSManagedClusterNPEffectiveOutboundIPs struct {
	ID string `yaml:"id,omitempty"`
}

type AKSManagedClusterNPManagedOutboundIPs struct {
	Count *int `yaml:"count,omitempty"`
}

type AKSManagedClusterNPOutboundIPs struct {
	PublicIPs []*AKSManagedClusterNPOutboundIPsPublicIps `yaml:"publicIps,omitempty"`
}
type AKSManagedClusterNPOutboundIPsPublicIps struct {
	ID string `yaml:"id,omitempty"`
}

type AKSManagedClusterNPOutboundIPPrefixes struct {
	PublicIPPrefixes []*AKSManagedClusterNPManagedOutboundIPsPublicIpPrefixes `yaml:"publicIpPrefixes,omitempty"`
}

type AKSManagedClusterNPManagedOutboundIPsPublicIpPrefixes struct {
	ID string `yaml:"id,omitempty"`
}

type AKSManagedClusterAzureADProfile struct {
	AdminGroupObjectIDs []string `yaml:"adminGroupObjectIds,omitempty"`
	ClientAppId         string   `yaml:"clientAppId,omitempty"`
	EnableAzureRbac     *bool    `yaml:"enableAzureRbac,omitempty"`
	Managed             *bool    `yaml:"managed,omitempty"`
	ServerAppId         string   `yaml:"serverAppId,omitempty"`
	ServerAppSecret     string   `yaml:"serverAppSecret,omitempty"`
	TenantId            string   `yaml:"tenantId,omitempty"`
}

type AKSManagedClusterAPIServerAccessProfile struct {
	AuthorizedIPRanges             []string `yaml:"authorizedIprRanges,omitempty"`
	EnablePrivateCluster           *bool    `yaml:"enablePrivateCluster,omitempty"`
	EnablePrivateClusterPublicFQDN *bool    `yaml:"enablePrivateClusterPublicFqdn,omitempty"`
	PrivateDnsZone                 string   `yaml:"privateDnsZone,omitempty"`
}

type AKSManagedClusterAutoScalerProfile struct {
	BalanceSimilarNodeGroups      string `yaml:"balance-similar-node-groups,omitempty"`
	Expander                      string `yaml:"expander,omitempty"`
	MaxEmptyBulkDelete            string `yaml:"max-empty-bulk-delete,omitempty"`
	MaxGracefulTerminationSec     string `yaml:"max-graceful-termination-sec,omitempty"`
	MaxNodeProvisionTime          string `yaml:"max-node-provision-time,omitempty"`
	MaxTotalUnreadyPercentage     string `yaml:"max-total-unready-percentage,omitempty"`
	NewPodScaleUpDelay            string `yaml:"new-pod-scale-up-delay,omitempty"`
	OkTotalUnreadyCount           string `yaml:"ok-total-unready-count,omitempty"`
	ScaleDownDelayAfterAdd        string `yaml:"scale-down-delay-after-add,omitempty"`
	ScaleDownDelayAfterDelete     string `yaml:"scale-down-delay-after-delete,omitempty"`
	ScaleDownDelayAfterFailure    string `yaml:"scale-down-delay-after-failure,omitempty"`
	ScaleDownUnneededTime         string `yaml:"scale-down-unneeded-time,omitempty"`
	ScaleDownUnreadyTime          string `yaml:"scale-down-unready-time,omitempty"`
	ScaleDownUtilizationThreshold string `yaml:"scale-down-utilization-threshold,omitempty"`
	ScanInterval                  string `yaml:"scan-interval,omitempty"`
	SkipNodesWithLocalStorage     string `yaml:"skip-nodes-with-local-storage,omitempty"`
	SkipNodesWithSystemPods       string `yaml:"skip-nodes-with-system-pods,omitempty"`
}

type AKSManagedClusterAutoUpgradeProfile struct {
	UpgradeChannel string `yaml:"upgradeChannel,omitempty"`
}

type AKSManagedClusterServicePrincipalProfile struct {
	ClientID string `yaml:"clientId,omitempty"`
	Secret   string `yaml:"secret,omitempty"`
}

type AKSManagedClusterLinuxProfile struct {
	AdminUsername string                      `yaml:"adminUsername,omitempty"`
	SSH           *AKSManagedClusterSSHConfig `yaml:"ssh,omitempty"`
	NoProxy       []string                    `yaml:"noProxy,omitempty"`
	TrustedCa     string                      `yaml:"trustedCa,omitempty"`
}

type AKSManagedClusterSSHConfig struct {
	PublicKeys []*AKSManagedClusterSSHKeyData `yaml:"publicKeys,omitempty"`
}

type AKSManagedClusterSSHKeyData struct {
	KeyData string `yaml:"keyData,omitempty"`
}

type AKSManagedClusterWindowsProfile struct {
	AdminUsername  string `yaml:"adminUsername,omitempty"`
	AdminPassword  string `yaml:"adminPassword,omitempty"`
	LicenseType    string `yaml:"licenseType,omitempty"`
	EnableCSIProxy *bool  `yaml:"enableCsiProxy,omitempty"`
}

type AKSManagedClusterHTTPProxyConfig struct {
	HTTPProxy  string   `yaml:"httpProxy,omitempty"`
	HTTPSProxy string   `yaml:"httpsProxy,omitempty"`
	NoProxy    []string `yaml:"noProxy,omitempty"`
	TrustedCA  string   `yaml:"trustedCa,omitempty"`
}

type AKSManagedClusterPodIdentityProfile struct {
	AllowNetworkPluginKubenet      *bool                                               `yaml:"allowNetworkPluginKubenet,omitempty"`
	Enabled                        *bool                                               `yaml:"enabled,omitempty"`
	UserAssignedIdentities         *AKSManagedClusterPIPUserAssignedIdentities         `yaml:"userAssignedIdentities,omitempty"`
	UserAssignedIdentityExceptions *AKSManagedClusterPIPUserAssignedIdentityExceptions `yaml:"userAssignedIdentityExceptions,omitempty"`
}

type AKSManagedClusterPIPUserAssignedIdentities struct {
	BindingSelector string                        `yaml:"bindingSelector,omitempty"`
	Identity        *AKSManagedClusterUAIIdentity `yaml:"identity,omitempty"`
	Name            string                        `yaml:"name,omitempty"`
	Namespace       string                        `yaml:"namespace,omitempty"`
}

type AKSManagedClusterUAIIdentity struct {
	ClientId   string `yaml:"clientId,omitempty"`
	ObjectId   string `yaml:"objectId,omitempty"`
	ResourceId string `yaml:"resourceId,omitempty"`
}

type AKSManagedClusterPIPUserAssignedIdentityExceptions struct {
	Name      string            `yaml:"name,omitempty"`
	Namespace string            `yaml:"namespace,omitempty"`
	PodLabels map[string]string `yaml:"podLabels,omitempty"`
}

type AKSManagedClusterPrivateLinkResources struct {
	GroupId         string   `yaml:"groupId,omitempty"`
	ID              string   `yaml:"id,omitempty"`
	Name            string   `yaml:"name,omitempty"`
	RequiredMembers []string `yaml:"requiredMembers,omitempty"`
	Type            string   `yaml:"type,omitempty"`
}

type AKSManagedClusterSKU struct {
	Name string `yaml:"name,omitempty"`
	Tier string `yaml:"tier,omitempty"`
}

type AKSManagedClusterAdditionalMetadata struct {
	ACRProfile           *AKSManagedClusterAdditionalMetadataACRProfile `yaml:"acrProfile,omitempty"`
	OmsWorkspaceLocation string                                         `yaml:"oms_workspace_location,omitempty"`
}

type AKSManagedClusterAdditionalMetadataACRProfile struct {
	ResourceGroupName string `yaml:"resourceGroupName,omitempty"`
	ACRName           string `yaml:"acrName,omitempty"`
}

type AKSNodePool struct {
	APIVersion string                 `yaml:"apiversion,omitempty"`
	Name       string                 `yaml:"name,omitempty"`
	Properties *AKSNodePoolProperties `yaml:"properties,omitempty"`
	Type       string                 `yaml:"type,omitempty"`
	Location   string                 `yaml:"location,omitempty"`
}

type AKSNodePoolProperties struct {
	OsDiskSizeGB              *int                        `yaml:"osDiskSizeGb,omitempty"`
	Mode                      string                      `yaml:"mode,omitempty"`
	AvailabilityZones         []string                    `yaml:"availabilityZones,omitempty"`
	EnableAutoScaling         *bool                       `yaml:"enableAutoScaling,omitempty"`
	Count                     *int                        `yaml:"count,omitempty"`
	MinCount                  *int                        `yaml:"minCount,omitempty"`
	MaxCount                  *int                        `yaml:"maxCount,omitempty"`
	MaxPods                   *int                        `yaml:"maxPods,omitempty"`
	Type                      string                      `yaml:"type,omitempty"`
	EnableNodePublicIP        *bool                       `yaml:"enableNodePublicIp,omitempty"`
	NodeLabels                map[string]string           `yaml:"nodeLabels,omitempty"`
	NodeTaints                []string                    `yaml:"nodeTaints,omitempty"`
	VnetSubnetID              string                      `yaml:"vnetSubnetId,omitempty"`
	UpgradeSettings           *AKSNodePoolUpgradeSettings `yaml:"upgradeSettings,omitempty"`
	ScaleSetPriority          string                      `yaml:"scaleSetPriority,omitempty"`
	ScaleSetEvictionPolicy    string                      `yaml:"scaleSetEvictionPolicy,omitempty"`
	SpotMaxPrice              *float64                    `yaml:"spotMaxPrice,omitempty"`
	EnableEncryptionAtHost    *bool                       `yaml:"enableEncryptionAtHost,omitempty"`
	OrchestratorVersion       string                      `yaml:"orchestratorVersion,omitempty"`
	EnableFIPS                *bool                       `yaml:"enableFips,omitempty"`
	EnableUltraSSD            *bool                       `yaml:"enableUltraSsd,omitempty"`
	GpuInstanceProfile        string                      `yaml:"gpuInstanceProfile,omitempty"`
	KubeletConfig             *AKSNodePoolKubeletConfig   `yaml:"kubeletConfig,omitempty"`
	KubeletDiskType           string                      `yaml:"kubeletDiskType,omitempty"`
	LinuxOSConfig             *AKSNodePoolLinuxOsConfig   `yaml:"linuxOsConfig,omitempty"`
	NodePublicIPPrefixID      string                      `yaml:"nodePublicIpPrefixId,omitempty"`
	OsDiskType                string                      `yaml:"osDiskType,omitempty"`
	OsSku                     string                      `yaml:"osSku,omitempty"`
	OsType                    string                      `yaml:"osType,omitempty"`
	PodSubnetID               string                      `yaml:"podSubnetId,omitempty"`
	ProximityPlacementGroupID string                      `yaml:"proximityPlacementGroupId,omitempty"`
	Tags                      map[string]string           `yaml:"tags,omitempty"`
	VmSize                    string                      `yaml:"vmSize,omitempty"`
}

type AKSNodePoolUpgradeSettings struct {
	MaxSurge string `yaml:"maxSurge,omitempty"`
}

type AKSNodePoolKubeletConfig struct {
	AllowedUnsafeSysctls  []string `yaml:"allowedUnsafeSysctls,omitempty"`
	ContainerLogMaxFiles  *int     `yaml:"containerLogMaxFiles,omitempty"`
	ContainerLogMaxSizeMB *int     `yaml:"containerLogMaxSizeMb,omitempty"`
	CpuCfsQuota           *bool    `yaml:"cpuCfsQuota,omitempty"`
	CpuCfsQuotaPeriod     string   `yaml:"cpuCfsQuotaPeriod,omitempty"`
	CpuManagerPolicy      string   `yaml:"cpuManagerPolicy,omitempty"`
	FailSwapOn            *bool    `yaml:"failSwapOn,omitempty"`
	ImageGcHighThreshold  *int     `yaml:"imageGcHighThreshold,omitempty"`
	ImageGcLowThreshold   *int     `yaml:"imageGcLowThreshold,omitempty"`
	PodMaxPids            *int     `yaml:"podMaxPids,omitempty"`
	TopologyManagerPolicy string   `yaml:"topologyManagerPolicy,omitempty"`
}

type AKSNodePoolLinuxOsConfig struct {
	SwapFileSizeMB             *int                             `yaml:"swapFileSizeMb,omitempty"`
	Sysctls                    *AKSNodePoolLinuxOsConfigSysctls `yaml:"sysctls,omitempty"`
	TransparentHugePageDefrag  string                           `yaml:"transparentHugePageDefrag,omitempty"`
	TransparentHugePageEnabled string                           `yaml:"transparentHugePageEnabled,omitempty"`
}

type AKSNodePoolLinuxOsConfigSysctls struct {
	FsAioMaxNr                     *int   `yaml:"fsAioMaxNr,omitempty"`
	FsFileMax                      *int   `yaml:"fsFileMax,omitempty"`
	FsInotifyMaxUserWatches        *int   `yaml:"fsInotifyMaxUserWatches,omitempty"`
	FsNrOpen                       *int   `yaml:"fsNrOpen,omitempty"`
	KernelThreadsMax               *int   `yaml:"kernelThreadsMax,omitempty"`
	NetCoreNetdevMaxBacklog        *int   `yaml:"netCoreNetdevMaxBacklog,omitempty"`
	NetCoreOptmemMax               *int   `yaml:"netCoreOptmemMax,omitempty"`
	NetCoreRmemDefault             *int   `yaml:"netCoreRmemDefault,omitempty"`
	NetCoreRmemMax                 *int   `yaml:"netCoreRmemMax,omitempty"`
	NetCoreSomaxconn               *int   `yaml:"netCoreSomaxconn,omitempty"`
	NetCoreWmemDefault             *int   `yaml:"netCoreWmemDefault,omitempty"`
	NetCoreWmemMax                 *int   `yaml:"netCoreWmemMax,omitempty"`
	NetIpv4IpLocalPortRange        string `yaml:"netIpv4IpLocalPortRange,omitempty"`
	NetIpv4NeighDefaultGcThresh1   *int   `yaml:"netIpv4NeighDefaultGcThresh1,omitempty"`
	NetIpv4NeighDefaultGcThresh2   *int   `yaml:"netIpv4NeighDefaultGcThresh2,omitempty"`
	NetIpv4NeighDefaultGcThresh3   *int   `yaml:"netIpv4NeighDefaultGcThresh3,omitempty"`
	NetIpv4TcpFinTimeout           *int   `yaml:"netIpv4TcpFinTimeout,omitempty"`
	NetIpv4TcpkeepaliveIntvl       *int   `yaml:"netIpv4TcpkeepaliveIntvl,omitempty"`
	NetIpv4TcpKeepaliveProbes      *int   `yaml:"netIpv4TcpKeepaliveProbes,omitempty"`
	NetIpv4TcpKeepaliveTime        *int   `yaml:"netIpv4TcpKeepaliveTime,omitempty"`
	NetIpv4TcpMaxSynBacklog        *int   `yaml:"netIpv4TcpMaxSynBacklog,omitempty"`
	NetIpv4TcpMaxTwBuckets         *int   `yaml:"netIpv4TcpMaxTwBuckets,omitempty"`
	NetIpv4TcpTwReuse              *bool  `yaml:"netIpv4TcpTwReuse,omitempty"`
	NetNetfilterNfConntrackBuckets *int   `yaml:"netNetfilterNfConntrackBuckets,omitempty"`
	NetNetfilterNfConntrackMax     *int   `yaml:"netNetfilterNfConntrackMax,omitempty"`
	VmMaxMapCount                  *int   `yaml:"vmMaxMapCount,omitempty"`
	VmSwappiness                   *int   `yaml:"vmSwappiness,omitempty"`
	VmVfsCachePressure             *int   `yaml:"vmVfsCachePressure,omitempty"`
}

// type AKSRafayInternal struct {
// 	Parameters map[string]*AzureParameter `yaml:"parameters,omitempty"`
// 	Resources  *AKSRafayInternalResources `yaml:"resources,omitempty"`
// }

// type AzureParameter struct {
// 	Type string `yaml:"type"`
// }

// type AKSRafayInternalResources struct {
// 	OperationsManagementSolutions []*AzureOperationsManagementSolution `yaml:"omSolutions,omitempty"`
// 	RoleAssignments               []*AzureRoleAssignment               `yaml:"roleAssignments"`
// }

// type AzureRoleAssignment struct {
// 	Metadata   *AzureRafayMetadata            `yaml:"additionalMetadata"`
// 	Type       string                         `yaml:"type"`
// 	APIVersion string                         `yaml:"apiVersion"`
// 	Name       string                         `yaml:"name"`
// 	DependsOn  []string                       `yaml:"dependsOn,omitempty"`
// 	Properties *AzureRoleAssignmentProperties `yaml:"properties"`
// }

// type AzureRoleAssignmentProperties struct {
// 	Scope            string `yaml:"scope,omitempty"`
// 	RoleDefinitionID string `yaml:"roleDefinitionId"`
// 	PrincipalID      string `yaml:"principalId"`
// }

// type AzureOperationsManagementSolution struct {
// 	Metadata   *AzureRafayMetadata                          `yaml:"additionalMetadata"`
// 	Type       string                                       `yaml:"type"`
// 	APIVersion string                                       `yaml:"apiVersion"`
// 	Name       string                                       `yaml:"name"`
// 	Location   string                                       `yaml:"location,omitempty"`
// 	Tags       map[string]string                            `yaml:"tags,omitempty"`
// 	Plan       *AzureOperationsManagementSolutionPlan       `yaml:"plan"`
// 	Properties *AzureOperationsManagementSolutionProperties `yaml:"properties"`
// }

// type AzureOperationsManagementSolutionPlan struct {
// 	Name          string `yaml:"name,omitempty"`
// 	Publisher     string `yaml:"publisher,omitempty"`
// 	Product       string `yaml:"product,omitempty"`
// 	PromotionCode string `yaml:"promotionCode"`
// }

// type AzureOperationsManagementSolutionProperties struct {
// 	WorkspaceResourceID string `yaml:"workspaceResourceId"`
// }
