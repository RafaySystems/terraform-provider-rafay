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

type AzureRafayMetadata struct {
	SubscriptionID    string `yaml:"subscriptionID,omitempty"`
	ResourceGroupName string `yaml:"resourceGroupName,omitempty"`
}

type AKSClusterConfig struct {
	APIVersion string                    `yaml:"apiVersion"`
	Kind       string                    `yaml:"kind"`
	Metadata   *AKSClusterConfigMetadata `yaml:"metadata"`
	Spec       *AKSClusterConfigSpec     `yaml:"spec"`
}

type AKSClusterConfigMetadata struct {
	Name string `yaml:"name"`
}

type AKSClusterConfigSpec struct {
	ResourceGroupName string             `yaml:"resourceGroupName,omitempty"`
	ManagedCluster    *AKSManagedCluster `yaml:"managedCluster,omitempty"`
	NodePools         *[]AKSNodePool     `yaml:"nodePools,omitempty"`
	Internal          *AKSRafayInternal  `yaml:"internal,omitempty"`
}

type AzureContainerRegistryProfile struct {
	ResourceGroupName string `yaml:"resourceGroupName,omitempty"`
	RegistryName      string `yaml:"acrName"`
}

type AKSManagedCluster struct {
	Metadata   *AzureRafayClusterMetadata   `yaml:"additionalMetadata,omitempty"`
	Type       string                       `yaml:"type"`
	APIVersion string                       `yaml:"apiVersion"`
	Location   string                       `yaml:"location"`
	Identity   *AKSManagedClusterIdentity   `yaml:"identity,omitempty"`
	Properties *AKSManagedClusterProperties `yaml:"properties"`
	SKU        *AKSManagedClusterSKU        `yaml:"sku,omitempty"`
	Tags       map[string]string            `yaml:"tags,omitempty"`
}
type AzureRafayClusterMetadata struct {
	ACRProfile                  *AzureContainerRegistryProfile `yaml:"acrProfile,omitempty"`
	ServicePrincipalCredentials string                         `yaml:"service_principal_credential,omitempty"`
	OMSWorkspaceLocation        string                         `yaml:"oms_workspace_location,omitempty"`
	WindowsAdminCredentials     string                         `yaml:"windows_admin_credentials,omitempty"`
}

type AKSManagedClusterIdentity struct {
	Type string `yaml:"type"`
}

type AKSManagedClusterProperties struct {
	KubernetesVersion       string                                    `yaml:"kubernetesVersion,omitempty"`
	EnableRBAC              *bool                                     `yaml:"enableRBAC,omitempty"`
	DNSPrefix               string                                    `yaml:"dnsPrefix,omitempty"`
	NodeResourceGroup       string                                    `yaml:"nodeResourceGroup,omitempty"`
	NetworkProfile          *AKSManagedClusterNetworkProfile          `yaml:"networkProfile,omitempty"`
	AzureADProfile          *AKSManagedClusterAzureADProfile          `yaml:"aadProfile,omitempty"`
	APIServerAccessProfile  *AKSManagedClusterAPIServerAccessProfile  `yaml:"apiServerAccessProfile,omitempty"`
	DiskEncryptionSetID     string                                    `yaml:"diskEncryptionSetID,omitempty"`
	AddonProfiles           map[string]*AKSManagedClusterAddonProfile `yaml:"addonProfiles,omitempty"`
	ServicePrincipalProfile *AKSManagedClusterServicePrincipalProfile `yaml:"servicePrincipalProfile,omitempty"`
	LinuxProfile            *AKSManagedClusterLinuxProfile            `yaml:"linuxProfile,omitempty"`
	WindowsProfile          *AKSManagedClusterWindowsProfile          `yaml:"windowsProfile,omitempty"`
	HTTPProxyConfig         *AKSManagedClusterHTTPProxyConfig         `yaml:"httpProxyConfig,omitempty"`
}

type AKSManagedClusterNetworkProfile struct {
	LoadBalancerSKU  string `yaml:"loadBalancerSku,omitempty"`
	NetworkPlugin    string `yaml:"networkPlugin,omitempty"`
	NetworkPolicy    string `yaml:"networkPolicy,omitempty"`
	ServiceCIDR      string `yaml:"serviceCidr,omitempty"`
	DNSServiceIP     string `yaml:"dnsServiceIP,omitempty"`
	DockerBridgeCIDR string `yaml:"dockerBridgeCidr,omitempty"`
}

type AKSManagedClusterAzureADProfile struct {
	Managed             *bool    `yaml:"managed,omitempty"`
	AdminGroupObjectIDs []string `yaml:"adminGroupObjectIDs,omitempty"`
}

type AKSManagedClusterAPIServerAccessProfile struct {
	AuthorizedIPRanges   []string `yaml:"authorizedIPRanges,omitempty"`
	EnablePrivateCluster *bool    `yaml:"enablePrivateCluster,omitempty"`
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
	AdminUsername string                      `yaml:"adminUsername"`
	SSH           *AKSManagedClusterSSHConfig `yaml:"ssh"`
}

type AKSManagedClusterSSHConfig struct {
	PublicKeys []string `yaml:"publicKeys"`
}

type AKSManagedClusterWindowsProfile struct {
	AdminUsername  string `yaml:"adminUsername"`
	AdminPassword  string `yaml:"adminPassword,omitempty"`
	LicenseType    string `yaml:"licenseType,omitempty"`
	EnableCSIProxy *bool  `yaml:"enableCSIProxy,omitempty"`
}

type AKSManagedClusterHTTPProxyConfig struct {
	HTTPProxy  string   `yaml:"httpProxy,omitempty"`
	HTTPSProxy string   `yaml:"httpsProxy,omitempty"`
	NoProxy    []string `yaml:"noProxy,omitempty"`
	TrustedCA  string   `yaml:"trustedCa,omitempty"`
}

type AKSManagedClusterSKU struct {
	Name string `yaml:"name,omitempty"`
	Tier string `yaml:"tier,omitempty"`
}

type AKSNodePool struct {
	Type       string                 `yaml:"type"`
	APIVersion string                 `yaml:"apiVersion"`
	Name       string                 `yaml:"name"`
	Location   string                 `yaml:"location"`
	Properties *AKSNodePoolProperties `yaml:"properties"`
}

type AKSNodePoolProperties struct {
	VMSize                 string                      `yaml:"vmSize,omitempty"`
	OSDiskSizeGB           *int                        `yaml:"osDiskSizeGB,omitempty"`
	Mode                   string                      `yaml:"mode,omitempty"`
	OSType                 string                      `yaml:"osType,omitempty"`
	AvailabilityZones      []string                    `yaml:"availabilityZones,omitempty"`
	EnableAutoScaling      *bool                       `yaml:"enableAutoScaling,omitempty"`
	Count                  *int                        `yaml:"count,omitempty"`
	MinCount               *int                        `yaml:"minCount,omitempty"`
	MaxCount               *int                        `yaml:"maxCount,omitempty"`
	MaxPods                *int                        `yaml:"maxPods,omitempty"`
	Type                   string                      `yaml:"type,omitempty"`
	EnableNodePublicIP     *bool                       `yaml:"enableNodePublicIP,omitempty"`
	NodeLabels             map[string]string           `yaml:"nodeLabels,omitempty"`
	NodeTaints             []string                    `yaml:"nodeTaints,omitempty"`
	VnetSubnetID           string                      `yaml:"vnetSubnetID,omitempty"`
	UpgradeSettings        *AKSNodePoolUpgradeSettings `yaml:"upgradeSettings,omitempty"`
	ScaleSetPriority       string                      `yaml:"scaleSetPriority,omitempty"`
	ScaleSetEvictionPolicy string                      `yaml:"scaleSetEvictionPolicy,omitempty"`
	SpotMaxPrice           *float64                    `yaml:"spotMaxPrice,omitempty"`
	EnableEncryptionAtHost *bool                       `yaml:"enableEncryptionAtHost,omitempty"`
	OrchestratorVersion    string                      `yaml:"orchestratorVersion,omitempty"`
}

type AKSNodePoolUpgradeSettings struct {
	MaxSurge string `yaml:"maxSurge,omitempty"`
}

type AKSRafayInternal struct {
	Parameters map[string]*AzureParameter `yaml:"parameters,omitempty"`
	Resources  *AKSRafayInternalResources `yaml:"resources,omitempty"`
}

type AzureParameter struct {
	Type string `yaml:"type"`
}

type AKSRafayInternalResources struct {
	OperationsManagementSolutions []*AzureOperationsManagementSolution `yaml:"omSolutions,omitempty"`
	RoleAssignments               []*AzureRoleAssignment               `yaml:"roleAssignments"`
}

type AzureRoleAssignment struct {
	Metadata   *AzureRafayMetadata            `yaml:"additionalMetadata"`
	Type       string                         `yaml:"type"`
	APIVersion string                         `yaml:"apiVersion"`
	Name       string                         `yaml:"name"`
	DependsOn  []string                       `yaml:"dependsOn,omitempty"`
	Properties *AzureRoleAssignmentProperties `yaml:"properties"`
}

type AzureRoleAssignmentProperties struct {
	Scope            string `yaml:"scope,omitempty"`
	RoleDefinitionID string `yaml:"roleDefinitionId"`
	PrincipalID      string `yaml:"principalId"`
}

type AzureOperationsManagementSolution struct {
	Metadata   *AzureRafayMetadata                          `yaml:"additionalMetadata"`
	Type       string                                       `yaml:"type"`
	APIVersion string                                       `yaml:"apiVersion"`
	Name       string                                       `yaml:"name"`
	Location   string                                       `yaml:"location,omitempty"`
	Tags       map[string]string                            `yaml:"tags,omitempty"`
	Plan       *AzureOperationsManagementSolutionPlan       `yaml:"plan"`
	Properties *AzureOperationsManagementSolutionProperties `yaml:"properties"`
}

type AzureOperationsManagementSolutionPlan struct {
	Name          string `yaml:"name,omitempty"`
	Publisher     string `yaml:"publisher,omitempty"`
	Product       string `yaml:"product,omitempty"`
	PromotionCode string `yaml:"promotionCode"`
}

type AzureOperationsManagementSolutionProperties struct {
	WorkspaceResourceID string `yaml:"workspaceResourceId"`
}
