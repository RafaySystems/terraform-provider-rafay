package rafay

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
	APIVersion string              `yaml:"apiversion"`
	Kind       string              `yaml:"kind"`
	Metadata   *AKSClusterMetadata `yaml:"metadata"`
	Spec       *AKSClusterSpec     `yaml:"spec"`
}

type AKSClusterMetadata struct {
	Name    string            `yaml:"name"`
	Project string            `yaml:"project"`
	Labels  map[string]string `yaml:"labels,omitempty"`
}

type AKSClusterSpec struct {
	Type             string            `yaml:"type"`
	Blueprint        string            `yaml:"blueprint"`
	BlueprintVersion string            `yaml:"blueprintversion"`
	CloudProvider    string            `yaml:"cloudprovider"`
	AKSClusterConfig *AKSClusterConfig `yaml:"clusterConfig"`
}

type AzureRafayMetadata struct {
	SubscriptionID    string `yaml:"subscriptionId,omitempty"`
	ResourceGroupName string `yaml:"resourceGroupName,omitempty"`
}

type AKSClusterConfig struct {
	APIVersion string                    `yaml:"apiversion"`
	Kind       string                    `yaml:"kind"`
	Metadata   *AKSClusterConfigMetadata `yaml:"metadata"`
	Spec       *AKSClusterConfigSpec     `yaml:"spec"`
}

type AKSClusterConfigMetadata struct {
	Name string `yaml:"name"`
}

type AKSClusterConfigSpec struct {
	SubscriptionID    string             `yaml:"subscriptionId,omitempty"`
	ResourceGroupName string             `yaml:"resourceGroupName,omitempty"`
	ManagedCluster    *AKSManagedCluster `yaml:"managedCluster,omitempty"`
	NodePools         []*AKSNodePool     `yaml:"nodePools,omitempty"`
	//Internal          *AKSRafayInternal  `yaml:"internal,omitempty"`
}

// type AzureContainerRegistryProfile struct {
// 	ResourceGroupName string `yaml:"resource_group_name,omitempty"`
// 	RegistryName      string `yaml:"acrName"`
// }

type AKSManagedCluster struct {
	ExtendedLocation *AKSClusterExtendedLocation `yaml:"extendedLocation,omitempty"`
	//Metadata         *AzureRafayClusterMetadata   `yaml:"additionalMetadata,omitempty"`
	Type       string                       `yaml:"type"`
	APIVersion string                       `yaml:"apiversion"`
	Location   string                       `yaml:"location"`
	Identity   *AKSManagedClusterIdentity   `yaml:"identity,omitempty"`
	Properties *AKSManagedClusterProperties `yaml:"properties"`
	SKU        *AKSManagedClusterSKU        `yaml:"sku,omitempty"`
	Tags       map[string]string            `yaml:"tags,omitempty"`
}

type AKSClusterExtendedLocation struct {
	Name string `yaml:"name,omitempty"`
	Type string `yaml:"type,omitempty"`
}

// type AzureRafayClusterMetadata struct {
// 	ACRProfile                  *AzureContainerRegistryProfile `yaml:"acrProfile,omitempty"`
// 	ServicePrincipalCredentials string                         `yaml:"service_principal_credential,omitempty"`
// 	OMSWorkspaceLocation        string                         `yaml:"oms_workspace_location,omitempty"`
// 	WindowsAdminCredentials     string                         `yaml:"windows_admin_credentials,omitempty"`
// }

type AKSManagedClusterIdentity struct {
	Type                   string            `yaml:"type,omitempty"`
	UserAssignedIdentities map[string]string `yaml:"userAssignedIdentities,omitempty"`
}

type AKSManagedClusterProperties struct {
	KubernetesVersion       string                                    `yaml:"kubernetesVersion,omitempty"`
	EnableRBAC              *bool                                     `yaml:"enableRbac,omitempty"`
	FQDNSubdomain           string                                    `yaml:"fqdnSubdomain,omitempty"`
	DNSPrefix               string                                    `yaml:"dnsPrefix,omitempty"`
	EnablePodSecurityPolicy *bool                                     `yaml:"enablePodSecurityPolicy,omitempty"`
	NodeResourceGroup       string                                    `yaml:"nodeResourceGroup,omitempty"`
	NetworkProfile          *AKSManagedClusterNetworkProfile          `yaml:"networkProfile,omitempty"`
	AzureADProfile          *AKSManagedClusterAzureADProfile          `yaml:"aadProfile,omitempty"`
	APIServerAccessProfile  *AKSManagedClusterAPIServerAccessProfile  `yaml:"apiServerAccessProfile,omitempty"`
	DisableLocalAccounts    *bool                                     `yaml:"disableLocalAccounts,omitempty"`
	DiskEncryptionSetID     string                                    `yaml:"diskEncryptionSetId,omitempty"`
	AddonProfiles           map[string]string                         `yaml:"addonProfiles,omitempty"`
	ServicePrincipalProfile *AKSManagedClusterServicePrincipalProfile `yaml:"servicePrincipalProfile,omitempty"`
	LinuxProfile            *AKSManagedClusterLinuxProfile            `yaml:"linuxProfile,omitempty"`
	WindowsProfile          *AKSManagedClusterWindowsProfile          `yaml:"windowsProfile,omitempty"`
	HTTPProxyConfig         *AKSManagedClusterHTTPProxyConfig         `yaml:"httpProxyConfig,omitempty"`
	IdentityProfile         map[string]string                         `yaml:"identityProfile,omitempty"`
	AutoScalerProfile       *AKSManagedClusterAutoScalerProfile       `yaml:"autoScalerProfile,omitempty"`
	AutoUpgradeProfile      *AKSManagedClusterAutoUpgradeProfile      `yaml:"autoUpgradeProfile,omitempty"`
	PodIdentityProfile      *AKSManagedClusterPodIdentityProfile      `yaml:"podIdentityProfile,omitempty"`
	PrivateLinkResources    []*AKSManagedClusterPrivateLinkResources  `yaml:"privateLinkResources,omitempty"`
}

type AKSManagedClusterNetworkProfile struct {
	LoadBalancerSKU     string                                  `yaml:"loadBalancerSku,omitempty"`
	NetworkPlugin       string                                  `yaml:"networkPlugin,omitempty"`
	NetworkPolicy       string                                  `yaml:"networkPolicy,omitempty"`
	ServiceCIDR         string                                  `yaml:"serviceCidr,omitempty"`
	DNSServiceIP        string                                  `yaml:"dnsServiceIp,omitempty"`
	DockerBridgeCIDR    string                                  `yaml:"dockerBridgeCidr,omitempty"`
	LoadBalancerProfile *AKSManagedClusterNPLoadBalancerProfile `yaml:"loadBalancerProfile,omitempty"`
	NetworkMode         string                                  `yaml:"networkMode,omitempty"`
	OutboundType        string                                  `yaml:"outboundType,omitempty"`
	PodCidr             string                                  `yaml:"podCidr,omitempty"`
	ServiceCidr         string                                  `yaml:"serviceCidr,omitempty"`
}

type AKSManagedClusterNPLoadBalancerProfile struct {
	AllocatedOutboundPorts *int                                       `yaml:"allocated_outbound_ports,omitempty"`
	EffectiveOutboundIPs   []*AKSManagedClusterNPEffectiveOutboundIPs `yaml:"effective_outboundIps,omitempty"`
	IdleTimeoutInMinutes   *int                                       `yaml:"idle_timeout_in_minutes,omitempty"`
	ManagedOutboundIPs     *AKSManagedClusterNPManagedOutboundIPs     `yaml:"managed_outboundIps,omitempty"`
	OutboundIPPrefixes     *AKSManagedClusterNPOutboundIPPrefixes     `yaml:"outboundIp_prefixes,omitempty"`
	OutboundIPs            *AKSManagedClusterNPOutboundIPs            `yaml:"outboundIps,omitempty"`
}

type AKSManagedClusterNPEffectiveOutboundIPs struct {
	ID string `yaml:"id,omitempty"`
}

type AKSManagedClusterNPManagedOutboundIPs struct {
	Count *int `yaml:"count,omitempty"`
}

type AKSManagedClusterNPOutboundIPs struct {
	PublicIPs *AKSManagedClusterNPOutboundIPsPublicIps `yaml:"publicIps,omitempty"`
}
type AKSManagedClusterNPOutboundIPsPublicIps struct {
	ID string `yaml:"id,omitempty"`
}

type AKSManagedClusterNPOutboundIPPrefixes struct {
	PublicIPPrefixes *AKSManagedClusterNPManagedOutboundIPsPublicIpPrefixes `yaml:"publicIp_prefixes,omitempty"`
}

type AKSManagedClusterNPManagedOutboundIPsPublicIpPrefixes struct {
	ID string `yaml:"id,omitempty"`
}

type AKSManagedClusterAzureADProfile struct {
	AdminGroupObjectIDs []string `yaml:"admin_group_objectIds,omitempty"`
	ClientAppId         string   `yaml:"client_appId,omitempty"`
	EnableAzureRbac     *bool    `yaml:"enable_azure_rbac,omitempty"`
	Managed             *bool    `yaml:"managed,omitempty"`
	ServerAppId         string   `yaml:"server_appId,omitempty"`
	ServerAppSecret     string   `yaml:"server_app_secret,omitempty"`
	TenantId            string   `yaml:"tenantId,omitempty"`
}

type AKSManagedClusterAPIServerAccessProfile struct {
	AuthorizedIPRanges             []string `yaml:"authorized_ipr_ranges,omitempty"`
	EnablePrivateCluster           *bool    `yaml:"enable_private_cluster,omitempty"`
	EnablePrivateClusterPublicFQDN *bool    `yaml:"enable_private_cluster_public_fqdn,omitempty"`
	PrivateDnsZone                 string   `yaml:"private_dns_zone,omitempty"`
}

type AKSManagedClusterAutoScalerProfile struct {
	BalanceSimilarNodeGroups      string `yaml:"balance_similar_node_groups,omitempty"`
	Expander                      string `yaml:"expander,omitempty"`
	MaxEmptyBulkDelete            string `yaml:"max_empty_bulk_delete,omitempty"`
	MaxGracefulTerminationSec     string `yaml:"max_graceful_termination_sec,omitempty"`
	MaxNodeProvisionTime          string `yaml:"max_node_provision_time,omitempty"`
	MaxTotalUnreadyPercentage     string `yaml:"max_total_unready_percentage,omitempty"`
	NewPodScaleUpDelay            string `yaml:"new_pod_scale_up_delay,omitempty"`
	OkTotalUnreadyCount           *int   `yaml:"ok_total_unready_count,omitempty"`
	ScaleDownDelayAfterAdd        string `yaml:"scale_down_delay_after_add,omitempty"`
	ScaleDownDelayAfterDelete     string `yaml:"scale_down_delay_after_delete,omitempty"`
	ScaleDownDelayAfterFailure    string `yaml:"scale_down_delay_after_failure,omitempty"`
	ScaleDownUnneededTime         string `yaml:"scale_down_unneeded_time,omitempty"`
	ScaleDownUnreadyTime          string `yaml:"scale_down_unready_time,omitempty"`
	ScaleDownUtilizationThreshold string `yaml:"scale_down_utilization_threshold,omitempty"`
	ScanInterval                  string `yaml:"scan_interval,omitempty"`
	SkipNodesWithLocalStorage     string `yaml:"skip_nodes_with_local_storage,omitempty"`
	SkipNodesWithSystemPods       string `yaml:"skip_nodes_with_system_pods,omitempty"`
}

type AKSManagedClusterAutoUpgradeProfile struct {
	UpgradeChannel string `yaml:"upgrade_channel,omitempty"`
}

type AKSManagedClusterAddonProfile struct {
	Enabled *bool                  `yaml:"enabled"`
	Config  map[string]interface{} `yaml:"config,omitempty"`
}

type AKSManagedClusterServicePrincipalProfile struct {
	ClientID string `yaml:"clientId"`
	Secret   string `yaml:"secret,omitempty"`
}

type AKSManagedClusterLinuxProfile struct {
	AdminUsername string                      `yaml:"admin_username"`
	SSH           *AKSManagedClusterSSHConfig `yaml:"ssh"`
	NoProxy       []string                    `yaml:"no_proxy"`
	TrustedCa     string                      `yaml:"trusted_ca"`
}

type AKSManagedClusterSSHConfig struct {
	PublicKeys []*AKSManagedClusterSSHKeyData `yaml:"public_keys"`
}

type AKSManagedClusterSSHKeyData struct {
	KeyData string `yaml:"key_data"`
}

type AKSManagedClusterWindowsProfile struct {
	AdminUsername  string `yaml:"admin_username"`
	AdminPassword  string `yaml:"admin_password,omitempty"`
	LicenseType    string `yaml:"licenseType,omitempty"`
	EnableCSIProxy *bool  `yaml:"enable_csi_proxy,omitempty"`
}

type AKSManagedClusterHTTPProxyConfig struct {
	HTTPProxy  string   `yaml:"http_proxy,omitempty"`
	HTTPSProxy string   `yaml:"https_proxy,omitempty"`
	NoProxy    []string `yaml:"no_proxy,omitempty"`
	TrustedCA  string   `yaml:"trusted_ca,omitempty"`
}

type AKSManagedClusterPodIdentityProfile struct {
	AllowNetworkPluginKubenet      *bool                                               `yaml:"allow_network_plugin_kubenet,omitempty"`
	Enabled                        *bool                                               `yaml:"enabled,omitempty"`
	UserAssignedIdentities         *AKSManagedClusterPIPUserAssignedIdentities         `yaml:"user_assignedIdentities,omitempty"`
	UserAssignedIdentityExceptions *AKSManagedClusterPIPUserAssignedIdentityExceptions `yaml:"user_assignedIdentity_exceptions,omitempty"`
}

type AKSManagedClusterPIPUserAssignedIdentities struct {
	BindingSelector string                        `yaml:"binding_selector,omitempty"`
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
	PodLabels map[string]string `yaml:"pod_labels,omitempty"`
}

type AKSManagedClusterPrivateLinkResources struct {
	GroupId         string   `yaml:"groupId,omitempty"`
	ID              string   `yaml:"id,omitempty"`
	Name            string   `yaml:"name,omitempty"`
	RequiredMembers []string `yaml:"required_members,omitempty"`
	Type            string   `yaml:"type,omitempty"`
}

type AKSManagedClusterSKU struct {
	Name string `yaml:"name,omitempty"`
	Tier string `yaml:"tier,omitempty"`
}

type AKSNodePool struct {
	Type       string `yaml:"type"`
	APIVersion string `yaml:"apiversion"`
	Name       string `yaml:"name"`
	// Location   string                 `yaml:"location"`
	Properties *AKSNodePoolProperties `yaml:"properties"`
}

type AKSNodePoolProperties struct {
	VMSize                    string                      `yaml:"vm_size,omitempty"`
	OsDiskSizeGB              *int                        `yaml:"os_disk_size_gb,omitempty"`
	Mode                      string                      `yaml:"mode,omitempty"`
	AvailabilityZones         []string                    `yaml:"availability_zones,omitempty"`
	EnableAutoScaling         *bool                       `yaml:"enable_auto_scaling,omitempty"`
	Count                     *int                        `yaml:"count,omitempty"`
	MinCount                  *int                        `yaml:"min_count,omitempty"`
	MaxCount                  *int                        `yaml:"max_count,omitempty"`
	MaxPods                   *int                        `yaml:"max_pods,omitempty"`
	Type                      string                      `yaml:"type,omitempty"`
	EnableNodePublicIP        *bool                       `yaml:"enable_node_publicIp,omitempty"`
	NodeLabels                map[string]string           `yaml:"node_labels,omitempty"`
	NodeTaints                []string                    `yaml:"node_taints,omitempty"`
	VnetSubnetID              string                      `yaml:"vnet_subnetId,omitempty"`
	UpgradeSettings           *AKSNodePoolUpgradeSettings `yaml:"upgrade_settings,omitempty"`
	ScaleSetPriority          string                      `yaml:"scale_set_priority,omitempty"`
	ScaleSetEvictionPolicy    string                      `yaml:"scale_set_eviction_policy,omitempty"`
	SpotMaxPrice              *float64                    `yaml:"spot_max_price,omitempty"`
	EnableEncryptionAtHost    *bool                       `yaml:"enable_encryption_at_host,omitempty"`
	OrchestratorVersion       string                      `yaml:"orchestrator_version,omitempty"`
	EnableFIPS                *bool                       `yaml:"enable_fips,omitempty"`
	EnableUltraSSD            *bool                       `yaml:"enable_ultra_ssd,omitempty"`
	GpuInstanceProfile        string                      `yaml:"gpu_instanceProfile,omitempty"`
	KubeletConfig             *AKSNodePoolKubeletConfig   `yaml:"kubelet_config,omitempty"`
	KubeletDiskType           string                      `yaml:"kubelet_diskType,omitempty"`
	LinuxOSConfig             *AKSNodePoolLinuxOsConfig   `yaml:"linux_os_config,omitempty"`
	NodePublicIPPrefixID      string                      `yaml:"node_publicIp_prefixId,omitempty"`
	OsDiskType                string                      `yaml:"os_diskType,omitempty"`
	OsSku                     string                      `yaml:"os_sku,omitempty"`
	OsType                    string                      `yaml:"osType,omitempty"`
	PodSubnetID               string                      `yaml:"pod_subnetId,omitempty"`
	ProximityPlacementGroupID string                      `yaml:"proximity_placement_groupId,omitempty"`
	Tags                      map[string]string           `yaml:"tags,omitempty"`
	VmSize                    string                      `yaml:"vm_size,omitempty"`
}

type AKSNodePoolUpgradeSettings struct {
	MaxSurge string `yaml:"maxSurge,omitempty"`
}

type AKSNodePoolKubeletConfig struct {
	AllowedUnsafeSysctls  []string `yaml:"allowed_unsafe_sysctls,omitempty"`
	ContainerLogMaxFiles  *int     `yaml:"container_log_max_files,omitempty"`
	ContainerLogMaxSizeMB *int     `yaml:"container_log_max_size_mb,omitempty"`
	CpuCfsQuota           *bool    `yaml:"cpu_cfs_quota,omitempty"`
	CpuCfsQuotaPeriod     string   `yaml:"cpu_cfs_quota_period,omitempty"`
	CpuManagerPolicy      string   `yaml:"cpu_manager_policy,omitempty"`
	FailSwapOn            *bool    `yaml:"fail_swap_on,omitempty"`
	ImageGcHighThreshold  *int     `yaml:"image_gc_high_threshold,omitempty"`
	ImageGcLowThreshold   *int     `yaml:"image_gc_low_threshold,omitempty"`
	PodMaxPids            *int     `yaml:"pod_max_pids,omitempty"`
	TopologyManagerPolicy string   `yaml:"topology_manager_policy,omitempty"`
}

type AKSNodePoolLinuxOsConfig struct {
	SwapFileSizeMB             *int                             `yaml:"swap_file_size_mb,omitempty"`
	Sysctls                    *AKSNodePoolLinuxOsConfigSysctls `yaml:"sysctls,omitempty"`
	TransparentHugePageDefrag  string                           `yaml:"transparent_huge_page_defrag,omitempty"`
	TransparentHugePageEnabled string                           `yaml:"transparent_huge_page_enabled,omitempty"`
}

type AKSNodePoolLinuxOsConfigSysctls struct {
	FsAioMaxNr                     *int   `yaml:"fs_aio_max_nr,omitempty"`
	FsFileMax                      *int   `yaml:"fs_file_max,omitempty"`
	FsInotifyMaxUserWatches        *int   `yaml:"fs_inotify_max_user_watches,omitempty"`
	FsNrOpen                       *int   `yaml:"fs_nr_open,omitempty"`
	KernelThreadsMax               *int   `yaml:"kernel_threads_max,omitempty"`
	NetCoreNetdevMaxBacklog        *int   `yaml:"net_core_netdev_max_backlog,omitempty"`
	NetCoreOptmemMax               *int   `yaml:"net_core_optmem_max,omitempty"`
	NetCoreRmemDefault             *int   `yaml:"net_core_rmem_default,omitempty"`
	NetCoreRmemMax                 *int   `yaml:"net_core_rmem_max,omitempty"`
	NetCoreSomaxconn               *int   `yaml:"net_core_somaxconn,omitempty"`
	NetCoreWmemDefault             *int   `yaml:"net_core_wmem_default,omitempty"`
	NetCoreWmemMax                 *int   `yaml:"net_core_wmem_max,omitempty"`
	NetIpv4IpLocalPortRange        string `yaml:"netIpv4Ip_local_port_range,omitempty"`
	NetIpv4NeighDefaultGcThresh1   *int   `yaml:"netIpv4_neigh_default_gc_thresh1,omitempty"`
	NetIpv4NeighDefaultGcThresh2   *int   `yaml:"netIpv4_neigh_default_gc_thresh2,omitempty"`
	NetIpv4NeighDefaultGcThresh3   *int   `yaml:"netIpv4_neigh_default_gc_thresh3,omitempty"`
	NetIpv4TcpFinTimeout           *int   `yaml:"netIpv4_tcp_fin_timeout,omitempty"`
	NetIpv4TcpkeepaliveIntvl       *int   `yaml:"netIpv4_tcpkeepalive_intvl,omitempty"`
	NetIpv4TcpKeepaliveProbes      *int   `yaml:"netIpv4_tcp_keepalive_probes,omitempty"`
	NetIpv4TcpKeepaliveTime        *int   `yaml:"netIpv4_tcp_keepalive_time,omitempty"`
	NetIpv4TcpMaxSynBacklog        *int   `yaml:"netIpv4_tcp_max_syn_backlog,omitempty"`
	NetIpv4TcpMaxTwBuckets         *int   `yaml:"netIpv4_tcp_max_tw_buckets,omitempty"`
	NetIpv4TcpTwReuse              *bool  `yaml:"netIpv4_tcp_tw_reuse,omitempty"`
	NetNetfilterNfConntrackBuckets *int   `yaml:"net_netfilter_nf_conntrack_buckets,omitempty"`
	NetNetfilterNfConntrackMax     *int   `yaml:"net_netfilter_nf_conntrack_max,omitempty"`
	VmMaxMapCount                  *int   `yaml:"vm_max_map_count,omitempty"`
	VmSwappiness                   *int   `yaml:"vm_swappiness,omitempty"`
	VmVfsCachePressure             *int   `yaml:"vm_vfs_cache_pressure,omitempty"`
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
