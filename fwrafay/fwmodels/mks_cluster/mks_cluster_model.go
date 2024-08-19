package mks_cluster

import (
	"github.com/RafaySystems/rafay-common/proto/types/hub/commonpb"
	"github.com/RafaySystems/rafay-common/proto/types/hub/infrapb"
	"github.com/hashicorp/terraform-plugin-framework/types"
	v1 "k8s.io/api/core/v1"
)

// Main cluster model definition for Terraform
type MksClusterModel struct {
	APIVersion types.String  `tfsdk:"api_version"`
	Kind       types.String  `tfsdk:"kind"`
	Metadata   MetadataModel `tfsdk:"metadata"`
	Spec       SpecModel     `tfsdk:"spec"`
}

// Metadata related to the cluster model
type MetadataModel struct {
	Annotations types.Map       `tfsdk:"annotations"`
	CreatedBy   CreatedByModel  `tfsdk:"created_by"`
	Description types.String    `tfsdk:"description"`
	DisplayName types.String    `tfsdk:"display_name"`
	Labels      types.Map       `tfsdk:"labels"`
	ModifiedBy  ModifiedByModel `tfsdk:"modified_by"`
	Name        types.String    `tfsdk:"name"`
	Project     types.String    `tfsdk:"project"`
}

type CreatedByModel struct {
	ID        types.String `tfsdk:"id"`
	IsSSOUser types.Bool   `tfsdk:"is_ssouser"`
	Username  types.String `tfsdk:"username"`
}

type ModifiedByModel struct {
	ID        types.String `tfsdk:"id"`
	IsSSOUser types.Bool   `tfsdk:"is_ssouser"`
	Username  types.String `tfsdk:"username"`
}

// Specification model containing detailed configuration
type SpecModel struct {
	Blueprint                 BlueprintModel                 `tfsdk:"blueprint"`
	CloudCredentials          types.String                   `tfsdk:"cloud_credentials"`
	Config                    ConfigModel                    `tfsdk:"config"`
	Proxy                     ProxyModel                     `tfsdk:"proxy"`
	Sharing                   SharingModel                   `tfsdk:"sharing"`
	SystemComponentsPlacement SystemComponentsPlacementModel `tfsdk:"system_components_placement"`
}

type BlueprintModel struct {
	Name    types.String `tfsdk:"name"`
	Version types.String `tfsdk:"version"`
}

// Cluster configuration details
type ConfigModel struct {
	AutoApproveNodes      types.Bool             `tfsdk:"auto_approve_nodes"`
	SSH                   ClusterSSHModel        `tfsdk:"cluster_ssh"`
	DedicatedControlPlane types.Bool             `tfsdk:"dedicated_control_plane"`
	HighAvailability      types.Bool             `tfsdk:"high_availability"`
	KubernetesUpgrade     KubernetesUpgradeModel `tfsdk:"kubernetes_upgrade"`
	KubernetesVersion     types.String           `tfsdk:"kubernetes_version"`
	Location              types.String           `tfsdk:"location"`
	Network               NetworkModel           `tfsdk:"network"`
	Nodes                 []NodeModel            `tfsdk:"nodes"`
}

type ClusterSSHModel struct {
	Passphrase     types.String `tfsdk:"passphrase"`
	Port           types.String `tfsdk:"port"`
	PrivateKeyPath types.String `tfsdk:"private_key_path"`
	Username       types.String `tfsdk:"username"`
}

type KubernetesUpgradeModel struct {
	Params   ParamsModel  `tfsdk:"params"`
	Strategy types.String `tfsdk:"strategy"`
}

type ParamsModel struct {
	WorkerConcurrency types.String `tfsdk:"worker_concurrency"`
}

// Network configuration model
type NetworkModel struct {
	CNI           CNIModel     `tfsdk:"cni"`
	IPv6          IPv6Model    `tfsdk:"ipv6"`
	PodSubnet     types.String `tfsdk:"pod_subnet"`
	ServiceSubnet types.String `tfsdk:"service_subnet"`
}

type CNIModel struct {
	Name    types.String `tfsdk:"name"`
	Version types.String `tfsdk:"version"`
}

type IPv6Model struct {
	PodSubnet     types.String `tfsdk:"pod_subnet"`
	ServiceSubnet types.String `tfsdk:"service_subnet"`
}

// Node model with SSH configuration and taints
type NodeModel struct {
	Arch            types.String   `tfsdk:"arch"`
	Hostname        types.String   `tfsdk:"hostname"`
	Interface       types.String   `tfsdk:"interface"`
	Labels          types.Map      `tfsdk:"labels"`
	OperatingSystem types.String   `tfsdk:"operating_system"`
	PrivateIP       types.String   `tfsdk:"private_ip"`
	Roles           []types.String `tfsdk:"roles"`
	SSH             NodeSSHModel   `tfsdk:"ssh"`
	Taints          []TaintModel   `tfsdk:"taints"`
}

type NodeSSHModel struct {
	IPAddress      types.String `tfsdk:"ip_address"`
	Passphrase     types.String `tfsdk:"passphrase"`
	Port           types.String `tfsdk:"port"`
	PrivateKeyPath types.String `tfsdk:"private_key_path"`
	Username       types.String `tfsdk:"username"`
}

type TaintModel struct {
	Effect types.String `tfsdk:"effect"`
	Key    types.String `tfsdk:"key"`
	Value  types.String `tfsdk:"value"`
}

// Proxy configuration model
type ProxyModel struct {
	AllowInsecureBootstrap types.Bool   `tfsdk:"allow_insecure_bootstrap"`
	BootstrapCA            types.String `tfsdk:"bootstrap_ca"`
	Enabled                types.Bool   `tfsdk:"enabled"`
	HTTPProxy              types.String `tfsdk:"http_proxy"`
	HTTPSProxy             types.String `tfsdk:"https_proxy"`
	NoProxy                types.String `tfsdk:"no_proxy"`
	ProxyAuth              types.String `tfsdk:"proxy_auth"`
}

// Sharing configuration model
type SharingModel struct {
	Enabled  types.Bool     `tfsdk:"enabled"`
	Projects []ProjectModel `tfsdk:"projects"`
}

type ProjectModel struct {
	Name types.String `tfsdk:"name"`
}

// System components placement model
type SystemComponentsPlacementModel struct {
	DaemonSetOverride DaemonSetOverrideModel `tfsdk:"daemon_set_override"`
	NodeSelector      types.Map              `tfsdk:"node_selector"`
	Tolerations       []TolerationModel      `tfsdk:"tolerations"`
}

type DaemonSetOverrideModel struct {
	DaemonSetTolerations []DaemonSetTolerationModel `tfsdk:"tolerations"`
	NodeSelectionEnabled types.Bool                 `tfsdk:"node_selection_enabled"`
}

type DaemonSetTolerationModel struct {
	Effect            types.String `tfsdk:"effect"`
	Key               types.String `tfsdk:"key"`
	Operator          types.String `tfsdk:"operator"`
	TolerationSeconds types.Int64  `tfsdk:"toleration_seconds"`
	Value             types.String `tfsdk:"value"`
}

type TolerationModel struct {
	Effect            types.String `tfsdk:"effect"`
	Key               types.String `tfsdk:"key"`
	Operator          types.String `tfsdk:"operator"`
	TolerationSeconds types.Int64  `tfsdk:"toleration_seconds"`
	Value             types.String `tfsdk:"value"`
}

// Utility functions to handle conversion of Terraform types to native types
func getStringValue(tfString types.String) string {
	if tfString.IsNull() || tfString.IsUnknown() {
		return ""
	}
	return tfString.ValueString()
}

func getBoolValue(tfBool types.Bool) bool {
	if tfBool.IsNull() || tfBool.IsUnknown() {
		return false
	}
	return tfBool.ValueBool()
}

func convertMap(tfMap types.Map) map[string]string {
	result := make(map[string]string)
	for k, v := range tfMap.Elements() {
		if !v.IsNull() && !v.IsUnknown() {
			result[k] = v.String()
		}
	}
	return result
}

// Main conversion function from Terraform model to Hub model
func ConvertToHubMksCluster(tf *MksClusterModel, hub *infrapb.Cluster) {
	hub.ApiVersion = getStringValue(tf.APIVersion)
	hub.Kind = getStringValue(tf.Kind)
	ConvertToHubMetadata(&tf.Metadata, hub.Metadata)
	ConvertToHubSpec(&tf.Spec, hub.Spec)
}

// Conversion of SpecModel to infrapb.ClusterSpec
func ConvertToHubSpec(tf *SpecModel, hub *infrapb.ClusterSpec) {
	hub.Blueprint = &infrapb.ClusterBlueprint{
		Name:    getStringValue(tf.Blueprint.Name),
		Version: getStringValue(tf.Blueprint.Version),
	}
	hub.CloudCredentials = getStringValue(tf.CloudCredentials)

	mksConfig := &infrapb.MksV3ConfigObject{
		AutoApproveNodes:      getBoolValue(tf.Config.AutoApproveNodes),
		DedicatedControlPlane: getBoolValue(tf.Config.DedicatedControlPlane),
		HighAvailability:      getBoolValue(tf.Config.HighAvailability),
		KubernetesUpgrade: &infrapb.KubernetesUpgrade{
			Params: &infrapb.KubernetesUpgradeParams{
				WorkerConcurrency: getStringValue(tf.Config.KubernetesUpgrade.Params.WorkerConcurrency),
			},
			Strategy: getStringValue(tf.Config.KubernetesUpgrade.Strategy),
		},
		KubernetesVersion: getStringValue(tf.Config.KubernetesVersion),
		Location:          getStringValue(tf.Config.Location),
		Network: &infrapb.MksClusterNetworking{
			Cni: &infrapb.Cni{
				Name:    getStringValue(tf.Config.Network.CNI.Name),
				Version: getStringValue(tf.Config.Network.CNI.Version),
			},
			PodSubnet:     getStringValue(tf.Config.Network.PodSubnet),
			ServiceSubnet: getStringValue(tf.Config.Network.ServiceSubnet),
			Ipv6: &infrapb.MksSubnet{
				PodSubnet:     getStringValue(tf.Config.Network.IPv6.PodSubnet),
				ServiceSubnet: getStringValue(tf.Config.Network.IPv6.ServiceSubnet),
			},
		},
	}

	for _, node := range tf.Config.Nodes {
		n := &infrapb.MksNode{
			Arch:            getStringValue(node.Arch),
			Hostname:        getStringValue(node.Hostname),
			Interface:       getStringValue(node.Interface),
			OperatingSystem: getStringValue(node.OperatingSystem),
			PrivateIP:       getStringValue(node.PrivateIP),
			Labels:          convertMap(node.Labels),
			Ssh: &infrapb.MksNodeSshConfig{
				IpAddress:      getStringValue(node.SSH.IPAddress),
				Passphrase:     getStringValue(node.SSH.Passphrase),
				Port:           getStringValue(node.SSH.Port),
				PrivateKeyPath: getStringValue(node.SSH.PrivateKeyPath),
				Username:       getStringValue(node.SSH.Username),
			},
		}

		for _, role := range node.Roles {
			n.Roles = append(n.Roles, role.ValueString())
		}

		for _, t := range node.Taints {
			n.Taints = append(n.Taints, &v1.Taint{
				Effect: v1.TaintEffect(t.Effect.ValueString()),
				Key:    t.Key.ValueString(),
				Value:  t.Value.ValueString(),
			})
		}
		mksConfig.Nodes = append(mksConfig.Nodes, n)
	}

	hub.Config = &infrapb.ClusterSpec_Mks{
		Mks: mksConfig,
	}

	hub.Proxy = &infrapb.ClusterProxy{
		AllowInsecureBootstrap: getBoolValue(tf.Proxy.AllowInsecureBootstrap),
		BootstrapCA:            getStringValue(tf.Proxy.BootstrapCA),
		Enabled:                getBoolValue(tf.Proxy.Enabled),
		HttpProxy:              getStringValue(tf.Proxy.HTTPProxy),
		HttpsProxy:             getStringValue(tf.Proxy.HTTPSProxy),
		NoProxy:                getStringValue(tf.Proxy.NoProxy),
		ProxyAuth:              getStringValue(tf.Proxy.ProxyAuth),
	}

	hub.Sharing = &infrapb.Sharing{
		Enabled: getBoolValue(tf.Sharing.Enabled),
	}

	for _, proj := range tf.Sharing.Projects {
		hub.Sharing.Projects = append(hub.Sharing.Projects, &infrapb.Projects{
			Name: getStringValue(proj.Name),
		})
	}

	hub.SystemComponentsPlacement = &infrapb.SystemComponentsPlacement{

		NodeSelector: convertMap(tf.SystemComponentsPlacement.NodeSelector),
	}

	for _, tol := range tf.SystemComponentsPlacement.Tolerations {
		hub.SystemComponentsPlacement.Tolerations = append(hub.SystemComponentsPlacement.Tolerations, &v1.Toleration{
			Effect:            v1.TaintEffect(tol.Effect.ValueString()),
			Key:               tol.Key.ValueString(),
			Operator:          v1.TolerationOperator(tol.Operator.ValueString()),
			TolerationSeconds: tol.TolerationSeconds.ValueInt64Pointer(),
			Value:             tol.Value.ValueString(),
		})
	}
}

// Conversion of MetadataModel to infrapb.ClusterMetadata
func ConvertToHubMetadata(tf *MetadataModel, hub *commonpb.Metadata) {
	hub.Annotations = convertMap(tf.Annotations)
	hub.CreatedBy = &commonpb.UserMeta{
		Id:        getStringValue(tf.CreatedBy.ID),
		IsSSOUser: getBoolValue(tf.CreatedBy.IsSSOUser),
		Username:  getStringValue(tf.CreatedBy.Username),
	}
	hub.Description = getStringValue(tf.Description)
	hub.DisplayName = getStringValue(tf.DisplayName)
	hub.Labels = convertMap(tf.Labels)
	hub.ModifiedBy = &commonpb.UserMeta{
		Id:        getStringValue(tf.ModifiedBy.ID),
		IsSSOUser: getBoolValue(tf.ModifiedBy.IsSSOUser),
		Username:  getStringValue(tf.ModifiedBy.Username),
	}
	hub.Name = getStringValue(tf.Name)
	hub.Project = getStringValue(tf.Project)
}
