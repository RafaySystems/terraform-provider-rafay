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
	APIVersion string              `yaml:"apiVersion,omitempty"`
	Kind       string              `yaml:"kind,omitempty"`
	Metadata   *AKSClusterMetadata `yaml:"metadata,omitempty"`
	Spec       *AKSClusterSpec     `yaml:"spec,omitempty"`
}

type AKSClusterMetadata struct {
	Name    string            `yaml:"name,omitempty"`
	Project string            `yaml:"project,omitempty"`
	Labels  map[string]string `yaml:"labels,omitempty,omitempty"`
}

// AKSClusterSpecProxyConfig is the spec-level proxy configuration for Rafay system components (bootstrap, agents).
type AKSClusterSpecProxyConfig struct {
	HttpProxy              string `yaml:"httpProxy,omitempty"`
	HttpsProxy             string `yaml:"httpsProxy,omitempty"`
	NoProxy                string `yaml:"noProxy,omitempty"`
	Enabled                bool   `yaml:"enabled,omitempty"`
	ProxyAuth              string `yaml:"proxyAuth,omitempty"`
	BootstrapCA            string `yaml:"bootstrapCA,omitempty"`
	AllowInsecureBootstrap bool   `yaml:"allowInsecureBootstrap,omitempty"`
}

type AKSClusterSpec struct {
	Type                      string                     `yaml:"type,omitempty"`
	Blueprint                 string                     `yaml:"blueprint,omitempty"`
	BlueprintVersion          string                     `yaml:"blueprintversion,omitempty"`
	CloudProvider             string                     `yaml:"cloudprovider,omitempty"`
	AKSClusterConfig          *AKSClusterConfig          `yaml:"clusterConfig,omitempty"`
	Sharing                   *V1ClusterSharing          `yaml:"sharing,omitempty"`
	SystemComponentsPlacement *SystemComponentsPlacement `yaml:"systemComponentsPlacement,omitempty"`
	ProxyConfig               *AKSClusterSpecProxyConfig `yaml:"proxyconfig,omitempty"`
}

type AzureRafayMetadata struct {
	SubscriptionID    string `yaml:"subscriptionId,omitempty"`
	ResourceGroupName string `yaml:"resourceGroupName,omitempty"`
}

type AKSClusterConfig struct {
	APIVersion string                    `yaml:"apiVersion,omitempty"`
	Kind       string                    `yaml:"kind,omitempty"`
	Metadata   *AKSClusterConfigMetadata `yaml:"metadata,omitempty"`
	Spec       *AKSClusterConfigSpec     `yaml:"spec,omitempty"`
}

type AKSClusterConfigMetadata struct {
	Name string `yaml:"name,omitempty"`
}

type AKSClusterConfigSpec struct {
	SubscriptionID     string                   `yaml:"subscriptionId,omitempty"`
	ResourceGroupName  string                   `yaml:"resourceGroupName,omitempty"`
	ManagedCluster     *AKSManagedCluster       `yaml:"managedCluster,omitempty"`
	NodePools          []*AKSNodePool           `yaml:"nodePools,omitempty"`
	MaintenanceConfigs []*AKSMaintenanceConfig  `yaml:"maintenanceConfigurations,omitempty"`
	WorkloadIdentities []*AzureWorkloadIdentity `yaml:"workloadIdentities,omitempty"`
	//Internal          *AKSRafayInternal  `yaml:"internal,omitempty"`
}

type AzureWorkloadIdentity struct {
	CreateIdentity     bool                                      `yaml:"createIdentity,omitempty"`
	Metadata           *AzureWorkloadIdentityMetadata            `yaml:"metadata,omitempty"`
	RoleAssignments    []*AzureWorkloadIdentityRoleAssignment    `yaml:"roleAssignments,omitempty"`
	K8sServiceAccounts []*AzureWorkloadIdentityK8sServiceAccount `yaml:"serviceAccounts,omitempty"`
}

type AzureWorkloadIdentityMetadata struct {
	ClientId      string            `yaml:"clientId,omitempty"`
	PrincipalId   string            `yaml:"principalId,omitempty"`
	Name          string            `yaml:"name,omitempty"`
	Location      string            `yaml:"location,omitempty"`
	ResourceGroup string            `yaml:"resourceGroup,omitempty"`
	Tags          map[string]string `yaml:"tags,omitempty"`
}

type AzureWorkloadIdentityRoleAssignment struct {
	Name             string `yaml:"name,omitempty"`
	RoleDefinitionId string `yaml:"roleDefinitionId,omitempty"`
	Scope            string `yaml:"scope,omitempty"`
}

type AzureWorkloadIdentityK8sServiceAccount struct {
	CreateAccount bool                       `yaml:"createAccount"`
	Metadata      *K8sServiceAccountMetadata `yaml:"metadata,omitempty"`
}

type K8sServiceAccountMetadata struct {
	Name        string            `yaml:"name"`
	Namespace   string            `yaml:"namespace"`
	Annotations map[string]string `yaml:"annotations,omitempty"`
	Labels      map[string]string `yaml:"labels,omitempty"`
}

type AKSMaintenanceConfig struct {
	ApiVersion string                          `yaml:"apiVersion,omitempty"`
	Name       string                          `yaml:"name,omitempty"`
	Properties *AKSMaintenanceConfigProperties `yaml:"properties,omitempty"`
	Type       string                          `yaml:"type,omitempty"`
}

type AKSMaintenanceConfigProperties struct {
	MaintenanceWindow *AKSMaintenanceWindow       `yaml:"maintenanceWindow,omitempty"`
	NotAllowedTime    []*AKSMaintenanceTimeSpan   `yaml:"notAllowedTime,omitempty"`
	TimeInWeek        []*AKSMaintenanceTimeInWeek `yaml:"timeInWeek,omitempty"`
}

type AKSMaintenanceWindow struct {
	DurationHours   int                       `yaml:"durationHours,omitempty"`
	NotAllowedDates []*AKSMaintenanceTimeSpan `yaml:"notAllowedDates,omitempty"`
	Schedule        *AKSMaintenanceSchedule   `yaml:"schedule,omitempty"`
	StartDate       string                    `yaml:"startDate,omitempty"`
	StartTime       string                    `yaml:"startTime,omitempty"`
	UtcOffset       string                    `yaml:"utcOffset,omitempty"`
}

type AKSMaintenanceTimeSpan struct {
	End   string `yaml:"end,omitempty"`
	Start string `yaml:"start,omitempty"`
}

type AKSMaintenanceTimeInWeek struct {
	Day       string `yaml:"day,omitempty"`
	HourSlots []int  `yaml:"hourSlots,omitempty"`
}

type AKSMaintenanceSchedule struct {
	AbsoluteMonthlySchedule *AKSMaintenanceAbsoluteMonthlySchedule `yaml:"absoluteMonthly,omitempty"`
	DailySchedule           *AKSMaintenanceDailySchedule           `yaml:"daily,omitempty"`
	RelativeMonthlySchedule *AKSMaintenanceRelativeMonthlySchedule `yaml:"relativeMonthly,omitempty"`
	WeeklySchedule          *AKSMaintenanceWeeklySchedule          `yaml:"weekly,omitempty"`
}

type AKSMaintenanceAbsoluteMonthlySchedule struct {
	DayOfMonth     int `yaml:"dayOfMonth,omitempty"`
	IntervalMonths int `yaml:"intervalMonths,omitempty"`
}

type AKSMaintenanceDailySchedule struct {
	IntervalDays int `yaml:"intervalDays,omitempty"`
}

type AKSMaintenanceRelativeMonthlySchedule struct {
	DayOfWeek      string `yaml:"dayOfWeek,omitempty"`
	IntervalMonths int    `yaml:"intervalMonths,omitempty"`
	WeekIndex      string `yaml:"weekIndex,omitempty"`
}

type AKSMaintenanceWeeklySchedule struct {
	DayOfWeek     string `yaml:"dayOfWeek,omitempty"`
	IntervalWeeks int    `yaml:"intervalWeeks,omitempty"`
}

// type AzureContainerRegistryProfile struct {
// 	ResourceGroupName string `yaml:"resourceGroupName,omitempty"`
// 	RegistryName      string `yaml:"acrName"`
// }

type AKSManagedCluster struct {
	ExtendedLocation   *AKSClusterExtendedLocation          `yaml:"extendedLocation,omitempty"`
	Type               string                               `yaml:"type,omitempty"`
	APIVersion         string                               `yaml:"apiVersion,omitempty"`
	Location           string                               `yaml:"location,omitempty"`
	Identity           *AKSManagedClusterIdentity           `yaml:"identity,omitempty"`
	Properties         *AKSManagedClusterProperties         `yaml:"properties,omitempty"`
	SKU                *AKSManagedClusterSKU                `yaml:"sku,omitempty"`
	Tags               map[string]interface{}               `yaml:"tags,omitempty"`
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
	EnableRBAC              *bool                                    `yaml:"enableRBAC,omitempty"`
	FQDNSubdomain           string                                   `yaml:"fqdnSubdomain,omitempty"`
	DNSPrefix               string                                   `yaml:"dnsPrefix,omitempty"`
	EnablePodSecurityPolicy *bool                                    `yaml:"enablePodSecurityPolicy,omitempty"`
	NodeResourceGroup       string                                   `yaml:"nodeResourceGroup,omitempty"`
	NetworkProfile          *AKSManagedClusterNetworkProfile         `yaml:"networkProfile,omitempty"`
	AzureADProfile          *AKSManagedClusterAzureADProfile         `yaml:"aadProfile,omitempty"`
	APIServerAccessProfile  *AKSManagedClusterAPIServerAccessProfile `yaml:"apiServerAccessProfile,omitempty"`
	DisableLocalAccounts    *bool                                    `yaml:"disableLocalAccounts,omitempty"`
	DiskEncryptionSetID     string                                   `yaml:"diskEncryptionSetID,omitempty"`
	AddonProfiles           *AddonProfiles                           `yaml:"addonProfiles,omitempty"`
	//AddonProfiles           map[string]string                         `yaml:"addonProfiles,omitempty"`
	SecurityProfile         *AKSManagedClusterSecurityProfile         `yaml:"securityProfile,omitempty"`
	ServicePrincipalProfile *AKSManagedClusterServicePrincipalProfile `yaml:"servicePrincipalProfile,omitempty"`
	LinuxProfile            *AKSManagedClusterLinuxProfile            `yaml:"linuxProfile,omitempty"`
	WindowsProfile          *AKSManagedClusterWindowsProfile          `yaml:"windowsProfile,omitempty"`
	HTTPProxyConfig         *AKSManagedClusterHTTPProxyConfig         `yaml:"httpProxyConfig,omitempty"`
	IdentityProfile         *AKSManagedClusterIdentityProfile         `yaml:"identityProfile,omitempty"`
	AutoScalerProfile       *AKSManagedClusterAutoScalerProfile       `yaml:"autoScalerProfile,omitempty"`
	AutoUpgradeProfile      *AKSManagedClusterAutoUpgradeProfile      `yaml:"autoUpgradeProfile,omitempty"`
	OidcIssuerProfile       *AKSManagedClusterOidcIssuerProfile       `yaml:"oidcIssuerProfile,omitempty"`
	PodIdentityProfile      *AKSManagedClusterPodIdentityProfile      `yaml:"podIdentityProfile,omitempty"`
	PrivateLinkResources    *AKSManagedClusterPrivateLinkResources    `yaml:"privateLinkResources,omitempty"`
	PowerState              *AKSManagedClusterPowerState              `yaml:"powerState,omitempty"`
}

type AddonProfiles struct {
	HttpApplicationRouting       *AKSManagedClusterAddonProfile         `yaml:"httpApplicationRouting,omitempty"`
	AzurePolicy                  *AKSManagedClusterAddonProfile         `yaml:"azurePolicy,omitempty"`
	OmsAgent                     *OmsAgentProfile                       `yaml:"omsAgent,omitempty"`
	AzureKeyvaultSecretsProvider *AzureKeyvaultSecretsProviderProfile   `yaml:"azureKeyvaultSecretsProvider,omitempty"`
	IngressApplicationGateway    *IngressApplicationGatewayAddonProfile `yaml:"ingressApplicationGateway,omitempty"`
}

type AKSManagedClusterAddonProfile struct {
	Enabled *bool                  `yaml:"enabled,omitempty"`
	Config  map[string]interface{} `yaml:"config,omitempty"`
}

type OmsAgentProfile struct {
	Enabled *bool           `yaml:"enabled,omitempty"`
	Config  *OmsAgentConfig `yaml:"config,omitempty"`
}

type AzureKeyvaultSecretsProviderProfile struct {
	Enabled *bool                                      `yaml:"enabled,omitempty"`
	Config  *AzureKeyvaultSecretsProviderProfileConfig `yaml:"config,omitempty"`
}

type OmsAgentConfig struct {
	LogAnalyticsWorkspaceResourceID string `yaml:"logAnalyticsWorkspaceResourceID,omitempty"`
}

type AzureKeyvaultSecretsProviderProfileConfig struct {
	EnableSecretRotation string `yaml:"enableSecretRotation,omitempty"`
	RotationPollInterval string `yaml:"rotationPollInterval,omitempty"`
}

type IngressApplicationGatewayAddonProfile struct {
	Enabled *bool                                 `yaml:"enabled,omitempty"`
	Config  *IngressApplicationGatewayAddonConfig `yaml:"config,omitempty"`
}

type IngressApplicationGatewayAddonConfig struct {
	ApplicationGatewayName string `yaml:"applicationGatewayName,omitempty"`
	ApplicationGatewayID   string `yaml:"applicationGatewayId,omitempty"`
	SubnetCIDR             string `yaml:"subnetCIDR,omitempty"`
	SubnetID               string `yaml:"subnetId,omitempty"`
	WatchNamespace         string `yaml:"watchNamespace,omitempty"`
}

type AKSManagedClusterSecurityProfile struct {
	WorkloadIdentity *AKSManagedClusterWorkloadIdentity `yaml:"workloadIdentity,omitempty"`
}

type AKSManagedClusterWorkloadIdentity struct {
	Enabled *bool `yaml:"enabled,omitempty"`
}

type AKSManagedClusterOidcIssuerProfile struct {
	Enabled *bool `yaml:"enabled,omitempty"`
}

type AKSManagedClusterNetworkProfile struct {
	LoadBalancerSKU     string                                  `yaml:"loadBalancerSku,omitempty"`
	NetworkPlugin       string                                  `yaml:"networkPlugin,omitempty"`
	NetworkPluginMode   string                                  `yaml:"networkPluginMode,omitempty"`
	NetworkPolicy       string                                  `yaml:"networkPolicy,omitempty"`
	NetworkDataplane    string                                  `yaml:"networkDataplane,omitempty"`
	DNSServiceIP        string                                  `yaml:"dnsServiceIP,omitempty"`
	DockerBridgeCidr    string                                  `yaml:"dockerBridgeCidr,omitempty"`
	LoadBalancerProfile *AKSManagedClusterNPLoadBalancerProfile `yaml:"loadBalancerProfile,omitempty"`
	NetworkMode         string                                  `yaml:"networkMode,omitempty"`
	OutboundType        string                                  `yaml:"outboundType,omitempty"`
	PodCidr             string                                  `yaml:"podCidr,omitempty"`
	ServiceCidr         string                                  `yaml:"serviceCidr,omitempty"`
}

type AKSManagedClusterNPLoadBalancerProfile struct {
	AllocatedOutboundPorts *int                                       `yaml:"allocatedOutboundPorts,omitempty"`
	EffectiveOutboundIPs   []*AKSManagedClusterNPEffectiveOutboundIPs `yaml:"effectiveOutboundIPs,omitempty"`
	IdleTimeoutInMinutes   *int                                       `yaml:"idleTimeoutInMinutes,omitempty"`
	ManagedOutboundIPs     *AKSManagedClusterNPManagedOutboundIPs     `yaml:"managedOutboundIPs,omitempty"`
	OutboundIPPrefixes     *AKSManagedClusterNPOutboundIPPrefixes     `yaml:"outboundIPPrefixes,omitempty"`
	OutboundIPs            *AKSManagedClusterNPOutboundIPs            `yaml:"outboundIPs,omitempty"`
}

type AKSManagedClusterNPEffectiveOutboundIPs struct {
	ID string `yaml:"id,omitempty"`
}

type AKSManagedClusterNPManagedOutboundIPs struct {
	Count *int `yaml:"count,omitempty"`
}

type AKSManagedClusterNPOutboundIPs struct {
	PublicIPs []*AKSManagedClusterNPOutboundIPsPublicIps `yaml:"publicIPs,omitempty"`
}
type AKSManagedClusterNPOutboundIPsPublicIps struct {
	ID string `yaml:"id,omitempty"`
}

type AKSManagedClusterNPOutboundIPPrefixes struct {
	PublicIPPrefixes []*AKSManagedClusterNPManagedOutboundIPsPublicIpPrefixes `yaml:"publicIPPrefixes,omitempty"`
}

type AKSManagedClusterNPManagedOutboundIPsPublicIpPrefixes struct {
	ID string `yaml:"id,omitempty"`
}

type AKSManagedClusterAzureADProfile struct {
	AdminGroupObjectIDs []string `yaml:"adminGroupObjectIDs,omitempty"`
	ClientAppId         string   `yaml:"clientAppID,omitempty"`
	EnableAzureRbac     *bool    `yaml:"enableAzureRBAC,omitempty"`
	Managed             *bool    `yaml:"managed,omitempty"`
	ServerAppId         string   `yaml:"serverAppID,omitempty"`
	ServerAppSecret     string   `yaml:"serverAppSecret,omitempty"`
	TenantId            string   `yaml:"tenantID,omitempty"`
}

type AKSManagedClusterAPIServerAccessProfile struct {
	AuthorizedIPRanges             []string `yaml:"authorizedIPRanges,omitempty"`
	EnablePrivateCluster           *bool    `yaml:"enablePrivateCluster,omitempty"`
	EnablePrivateClusterPublicFQDN *bool    `yaml:"enablePrivateClusterPublicFQDN,omitempty"`
	PrivateDnsZone                 string   `yaml:"privateDNSZone,omitempty"`
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
	UpgradeChannel       string `yaml:"upgradeChannel,omitempty"`
	NodeOsUpgradeChannel string `yaml:"nodeOsUpgradeChannel,omitempty"`
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

type AKSManagedClusterIdentityProfile struct {
	KubeletIdentity *AKSManagedClusterKubeletIdentity `yaml:"kubeletIdentity,omitempty"`
}

type AKSManagedClusterKubeletIdentity struct {
	ResourceId string `yaml:"resourceId,omitempty"`
}

type AKSManagedClusterPodIdentityProfile struct {
	AllowNetworkPluginKubenet      *bool                                                 `yaml:"allowNetworkPluginKubenet,omitempty"`
	Enabled                        *bool                                                 `yaml:"enabled,omitempty"`
	UserAssignedIdentities         []*AKSManagedClusterPIPUserAssignedIdentities         `yaml:"userAssignedIdentities,omitempty"`
	UserAssignedIdentityExceptions []*AKSManagedClusterPIPUserAssignedIdentityExceptions `yaml:"userAssignedIdentityExceptions,omitempty"`
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

type AKSManagedClusterPowerState struct {
	Code string `yaml:"code,omitempty"`
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
	ResourceGroupName string         `yaml:"resourceGroupName,omitempty"`
	ACRName           string         `yaml:"acrName,omitempty"`
	Registries        []*AksRegistry `yaml:"registries,omitempty"`
}

type AksRegistry struct {
	ACRName           string `yaml:"acrName,omitempty"`
	ResourceGroupName string `yaml:"resourceGroupName,omitempty"`
}

type AKSNodePool struct {
	APIVersion string                 `yaml:"apiVersion,omitempty"`
	Name       string                 `yaml:"name,omitempty"`
	Properties *AKSNodePoolProperties `yaml:"properties,omitempty"`
	Type       string                 `yaml:"type,omitempty"`
	Location   string                 `yaml:"location,omitempty"`
}

type AKSNodePoolProperties struct {
	OsDiskSizeGB              *int                        `yaml:"osDiskSizeGB,omitempty"`
	Mode                      string                      `yaml:"mode,omitempty"`
	AvailabilityZones         []string                    `yaml:"availabilityZones,omitempty"`
	EnableAutoScaling         *bool                       `yaml:"enableAutoScaling,omitempty"`
	Count                     *int                        `yaml:"count,omitempty"`
	MinCount                  *int                        `yaml:"minCount,omitempty"`
	MaxCount                  *int                        `yaml:"maxCount,omitempty"`
	MaxPods                   *int                        `yaml:"maxPods,omitempty"`
	Type                      string                      `yaml:"type,omitempty"`
	EnableNodePublicIP        *bool                       `yaml:"enableNodePublicIP,omitempty"`
	NodeLabels                map[string]string           `yaml:"nodeLabels,omitempty"`
	NodeTaints                []string                    `yaml:"nodeTaints,omitempty"`
	VnetSubnetID              string                      `yaml:"vnetSubnetID,omitempty"`
	UpgradeSettings           *AKSNodePoolUpgradeSettings `yaml:"upgradeSettings,omitempty"`
	ScaleSetPriority          string                      `yaml:"scaleSetPriority,omitempty"`
	ScaleSetEvictionPolicy    string                      `yaml:"scaleSetEvictionPolicy,omitempty"`
	SpotMaxPrice              *float64                    `yaml:"spotMaxPrice,omitempty"`
	EnableEncryptionAtHost    *bool                       `yaml:"enableEncryptionAtHost,omitempty"`
	OrchestratorVersion       string                      `yaml:"orchestratorVersion,omitempty"`
	EnableFIPS                *bool                       `yaml:"enableFIPS,omitempty"`
	EnableUltraSSD            *bool                       `yaml:"enableUltraSsd,omitempty"`
	GpuInstanceProfile        string                      `yaml:"gpuInstanceProfile,omitempty"`
	KubeletConfig             *AKSNodePoolKubeletConfig   `yaml:"kubeletConfig,omitempty"`
	KubeletDiskType           string                      `yaml:"kubeletDiskType,omitempty"`
	LinuxOSConfig             *AKSNodePoolLinuxOsConfig   `yaml:"linuxOsConfig,omitempty"`
	NodePublicIPPrefixID      string                      `yaml:"nodePublicIPPrefixID,omitempty"`
	OsDiskType                string                      `yaml:"osDiskType,omitempty"`
	OsSku                     string                      `yaml:"osSku,omitempty"`
	OsType                    string                      `yaml:"osType,omitempty"`
	PodSubnetID               string                      `yaml:"podSubnetID,omitempty"`
	ProximityPlacementGroupID string                      `yaml:"proximityPlacementGroupID,omitempty"`
	Tags                      map[string]string           `yaml:"tags,omitempty"`
	VmSize                    string                      `yaml:"vmSize,omitempty"`
	CreationData              *AKSNodePoolCreationData    `yaml:"creationData,omitempty"`
}

type AKSNodePoolCreationData struct {
	SourceResourceId string `yaml:"sourceResourceId,omitempty"`
}

type AKSNodePoolUpgradeSettings struct {
	MaxSurge string `yaml:"maxSurge,omitempty"`
}

type AKSNodePoolKubeletConfig struct {
	AllowedUnsafeSysctls  []string `yaml:"allowedUnsafeSysctls,omitempty"`
	ContainerLogMaxFiles  *int     `yaml:"containerLogMaxFiles,omitempty"`
	ContainerLogMaxSizeMB *int     `yaml:"containerLogMaxSizeMB,omitempty"`
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
	SwapFileSizeMB             *int                             `yaml:"swapFileSizeMB,omitempty"`
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
