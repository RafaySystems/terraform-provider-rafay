package mks_cluster

import (
	"context"
	"fmt"
	"log"

	"github.com/RafaySystems/rafay-common/proto/types/hub/commonpb"
	"github.com/RafaySystems/rafay-common/proto/types/hub/infrapb"
	v1 "k8s.io/api/core/v1"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

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

func convertFromTfMap(tfMap types.Map) map[string]string {
	result := make(map[string]string)
	for k, v := range tfMap.Elements() {
		result[k] = getStringValue(v.(types.String))
	}
	log.Println("Converted Map", result)
	return result
}

func convertToTfMap(goMap map[string]string) types.Map {
	elements := make(map[string]attr.Value)

	for k, v := range goMap {
		elements[k] = basetypes.NewStringValue(v)
	}
	tfMap, _ := basetypes.NewMapValue(types.StringType, elements)
	return tfMap

}

// Conversion of Terrafrom MetadataModel to commonpb.Metadata
func ConvertMetadataToHub(ctx context.Context, tf *MetadataValue) (*commonpb.Metadata, diag.Diagnostics) {

	var diag diag.Diagnostics

	hub := &commonpb.Metadata{}

	if !tf.Annotations.IsNull() && !tf.Annotations.IsUnknown() {
		hub.Annotations = convertFromTfMap(tf.Annotations)
	}

	if !tf.CreatedBy.IsNull() && !tf.CreatedBy.IsUnknown() {
		var createdByType CreatedByType

		tfCreatedBy, diag := createdByType.ValueFromObject(ctx, tf.CreatedBy)
		if diag.HasError() {
			return nil, diag
		}
		tfCreatedByValue := tfCreatedBy.(CreatedByValue)
		hub.CreatedBy = &commonpb.UserMeta{
			Id:        getStringValue(tfCreatedByValue.Id),
			IsSSOUser: getBoolValue(tfCreatedByValue.IsSsouser),
			Username:  getStringValue(tfCreatedByValue.Username),
		}

	}

	hub.Description = getStringValue(tf.Description)
	hub.DisplayName = getStringValue(tf.DisplayName)
	fmt.Println("TF Metadata lables", tf.Labels)

	hub.Labels = convertFromTfMap(tf.Labels)

	fmt.Println("Hub Metadata lables", hub.Labels)

	if !tf.ModifiedBy.IsNull() && !tf.ModifiedBy.IsUnknown() {
		var modifiedByType ModifiedByType

		tfModifiedBy, diag := modifiedByType.ValueFromObject(ctx, tf.ModifiedBy)
		if diag.HasError() {
			return nil, diag
		}

		tfModifiedByValue := tfModifiedBy.(ModifiedByValue)
		hub.ModifiedBy = &commonpb.UserMeta{
			Id:        getStringValue(tfModifiedByValue.Id),
			IsSSOUser: getBoolValue(tfModifiedByValue.IsSsouser),
			Username:  getStringValue(tfModifiedByValue.Username),
		}

	}
	hub.Name = getStringValue(tf.Name)
	hub.Project = getStringValue(tf.Project)

	return hub, diag
}

func ConvertMksNetworkToHub(ctx context.Context, tf basetypes.ObjectValue) (*infrapb.MksClusterNetworking, diag.Diagnostics) {
	var diags diag.Diagnostics

	var tfNetworkType NetworkType

	tfNetworkValue, err := tfNetworkType.ValueFromObject(ctx, tf)
	if err != nil {
		return nil, err
	}

	tfNetwork := tfNetworkValue.(NetworkValue)
	var cniType CniType
	tfCniValue, diag := cniType.ValueFromObject(ctx, tfNetwork.Cni)

	if diag.HasError() {
		return nil, diag
	}

	tfCni := tfCniValue.(CniValue)
	hubCni := &infrapb.Cni{
		Name:    getStringValue(tfCni.Name),
		Version: getStringValue(tfCni.Version),
	}

	hub := &infrapb.MksClusterNetworking{
		Cni: hubCni,
	}
	if !tfNetwork.PodSubnet.IsNull() && !tfNetwork.PodSubnet.IsUnknown() {
		hub.PodSubnet = getStringValue(tfNetwork.PodSubnet)
	}

	if !tfNetwork.ServiceSubnet.IsNull() && !tfNetwork.ServiceSubnet.IsUnknown() {
		hub.ServiceSubnet = getStringValue(tfNetwork.ServiceSubnet)
	}

	if !tfNetwork.Ipv6.IsNull() && !tfNetwork.Ipv6.IsUnknown() {
		// Handle IPv6
		var ipv6Type Ipv6Type
		tfIpv6Value, err := ipv6Type.ValueFromObject(ctx, tfNetwork.Ipv6)
		if err != nil {
			return nil, err
		}

		tfIpv6 := tfIpv6Value.(Ipv6Value)

		hub.Ipv6 = &infrapb.MksSubnet{
			PodSubnet:     getStringValue(tfIpv6.PodSubnet),
			ServiceSubnet: getStringValue(tfIpv6.ServiceSubnet),
		}
	}
	return hub, diags
}

func ConvertMksNodeToHub(ctx context.Context, tf NodesValue) (*infrapb.MksNode, diag.Diagnostics) {
	var diag diag.Diagnostics

	hub := &infrapb.MksNode{
		Arch:            getStringValue(tf.Arch),
		Hostname:        getStringValue(tf.Hostname),
		OperatingSystem: getStringValue(tf.OperatingSystem),
		PrivateIP:       getStringValue(tf.PrivateIp),
	}

	if !tf.Interface.IsNull() && !tf.Interface.IsUnknown() {
		hub.Interface = getStringValue(tf.Interface)
	}

	if tf.Roles.IsNull() || tf.Roles.IsUnknown() {
		diag.AddError("roles", "Roles are required for a node")
		return nil, diag
	}

	for _, role := range tf.Roles.Elements() {
		hub.Roles = append(hub.Roles, getStringValue(role.(types.String)))
	}

	if !tf.Labels.IsNull() && !tf.Labels.IsUnknown() {
		hub.Labels = convertFromTfMap(tf.Labels)
	}

	if !tf.Taints.IsNull() && !tf.Taints.IsUnknown() {
		for _, taint := range tf.Taints.Elements() {
			t, diag := ConvertTaintToHub(ctx, taint.(TaintsValue))
			if diag.HasError() {
				return nil, diag
			}
			hub.Taints = append(hub.Taints, t)
		}
	}

	if !tf.Ssh.IsNull() && !tf.Ssh.IsUnknown() {
		var sshType SshType

		tfSshValue, err := sshType.ValueFromObject(ctx, tf.Ssh)
		if err != nil {
			return nil, err
		}

		tfSsh := tfSshValue.(SshValue)
		hub.Ssh = &infrapb.MksNodeSshConfig{
			IpAddress:      getStringValue(tfSsh.IpAddress),
			Passphrase:     getStringValue(tfSsh.Passphrase),
			Port:           getStringValue(tfSsh.Port),
			PrivateKeyPath: getStringValue(tfSsh.PrivateKeyPath),
			Username:       getStringValue(tfSsh.Username),
		}
	}

	return hub, diag
}

func ConvertTaintToHub(ctx context.Context, tf TaintsValue) (*v1.Taint, diag.Diagnostics) {

	hub := &v1.Taint{
		Effect: v1.TaintEffect(getStringValue(tf.Effect)),
		Key:    getStringValue(tf.Key),
		Value:  getStringValue(tf.Value),
	}
	return hub, nil
}

func ConvertMksProxyToHub(ctx context.Context, tf basetypes.ObjectValue) (*infrapb.ClusterProxy, diag.Diagnostics) {
	var proxyType ProxyType

	tfProxyValue, err := proxyType.ValueFromObject(ctx, tf)
	if err != nil {
		return nil, err
	}
	tfProxy := tfProxyValue.(ProxyValue)

	hub := &infrapb.ClusterProxy{
		AllowInsecureBootstrap: getBoolValue(tfProxy.AllowInsecureBootstrap),
		BootstrapCA:            getStringValue(tfProxy.BootstrapCa),
		Enabled:                getBoolValue(tfProxy.Enabled),
		HttpProxy:              getStringValue(tfProxy.HttpProxy),
		HttpsProxy:             getStringValue(tfProxy.HttpsProxy),
		NoProxy:                getStringValue(tfProxy.NoProxy),
		ProxyAuth:              getStringValue(tfProxy.ProxyAuth),
	}

	return hub, nil
}

func ConvertMksSharingToHub(ctx context.Context, tf basetypes.ObjectValue) (*infrapb.Sharing, diag.Diagnostics) {
	var sharingType SharingType

	tfSharingValue, err := sharingType.ValueFromObject(ctx, tf)
	if err != nil {
		return nil, err
	}
	tfSharing := tfSharingValue.(SharingValue)

	hub := &infrapb.Sharing{
		Enabled: getBoolValue(tfSharing.Enabled),
	}

	diag := tfSharing.Projects.ElementsAs(ctx, hub.Projects, true)

	return hub, diag
}

func ConvertMksSystemComponentsPlacementToHub(ctx context.Context, tf basetypes.ObjectValue) (*infrapb.SystemComponentsPlacement, diag.Diagnostics) {
	var systemComponentsPlacementType SystemComponentsPlacementType

	var diag diag.Diagnostics

	tfObject, err := systemComponentsPlacementType.ValueFromObject(ctx, tf)
	if err != nil {
		return nil, err
	}
	tfValue := tfObject.(SystemComponentsPlacementValue)

	hub := &infrapb.SystemComponentsPlacement{
		NodeSelector: convertFromTfMap(tfValue.NodeSelector),
	}

	for _, toleration := range tfValue.Tolerations.Elements() {
		t, diag := ConvertTolerationsToHub(ctx, toleration.(TolerationsValue))
		if diag.HasError() {
			return nil, diag
		}
		hub.Tolerations = append(hub.Tolerations, t)
	}

	var daemonSetType DaemonSetOverrideType

	tfDaemonSetValue, err := daemonSetType.ValueFromObject(ctx, tfValue.DaemonSetOverride)

	if err != nil {
		return nil, err
	}

	tfDaemonSet := tfDaemonSetValue.(DaemonSetOverrideValue)

	hub.DaemonSetOverride = &infrapb.DaemonSetOverride{
		NodeSelectionEnabled: getBoolValue(tfDaemonSet.NodeSelectionEnabled),
	}

	for _, toleration := range tfDaemonSet.DaemonSetTolerations.Elements() {
		t, diag := ConvertDaemonSetTolerationsToHub(ctx, toleration.(DaemonSetTolerationsValue))
		if diag.HasError() {
			return nil, diag
		}
		hub.DaemonSetOverride.Tolerations = append(hub.DaemonSetOverride.Tolerations, t)
	}

	return hub, diag
}

func ConvertTolerationsToHub(ctx context.Context, tf TolerationsValue) (*v1.Toleration, diag.Diagnostics) {

	hub := &v1.Toleration{
		Effect:            v1.TaintEffect(getStringValue(tf.Effect)),
		Key:               getStringValue(tf.Key),
		Operator:          v1.TolerationOperator(getStringValue(tf.Operator)),
		TolerationSeconds: tf.TolerationSeconds.ValueInt64Pointer(),
		Value:             getStringValue(tf.Value),
	}

	return hub, nil

}

func ConvertDaemonSetTolerationsToHub(ctx context.Context, tf DaemonSetTolerationsValue) (*v1.Toleration, diag.Diagnostics) {

	hub := &v1.Toleration{
		Effect:            v1.TaintEffect(getStringValue(tf.Effect)),
		Key:               getStringValue(tf.Key),
		Operator:          v1.TolerationOperator(getStringValue(tf.Operator)),
		TolerationSeconds: tf.TolerationSeconds.ValueInt64Pointer(),
		Value:             getStringValue(tf.Value),
	}

	return hub, nil

}

func ConvertMksConfigToHub(ctx context.Context, tf basetypes.ObjectValue) (*infrapb.MksV3ConfigObject, diag.Diagnostics) {
	var diags diag.Diagnostics

	var configType ConfigType

	tfMksConfigValue, err := configType.ValueFromObject(ctx, tf)
	if err != nil {
		return nil, err
	}

	tfMksConfig := tfMksConfigValue.(ConfigValue)

	hub := &infrapb.MksV3ConfigObject{
		Location:              getStringValue(tfMksConfig.Location),
		AutoApproveNodes:      getBoolValue(tfMksConfig.AutoApproveNodes),
		DedicatedControlPlane: getBoolValue(tfMksConfig.DedicatedControlPlane),
		HighAvailability:      getBoolValue(tfMksConfig.HighAvailability),
		KubernetesVersion:     getStringValue(tfMksConfig.KubernetesVersion),
	}

	hub.Network, diags = ConvertMksNetworkToHub(ctx, tfMksConfig.Network)
	if diags.HasError() {
		return nil, diags
	}

	for _, node := range tfMksConfig.Nodes.Elements() {
		val := node.(NodesValue)
		n, diag := ConvertMksNodeToHub(ctx, val)
		if diag.HasError() {
			return nil, diag
		}
		hub.Nodes = append(hub.Nodes, n)
	}

	if !tfMksConfig.ClusterSsh.IsNull() && !tfMksConfig.ClusterSsh.IsUnknown() {

		var sshType ClusterSshType

		tfSshValue, err := sshType.ValueFromObject(ctx, tfMksConfig.ClusterSsh)
		if err != nil {
			return nil, err
		}
		tfSsh := tfSshValue.(ClusterSshValue)

		hub.Ssh = &infrapb.MksClusterSshConfig{
			PrivateKeyPath: getStringValue(tfSsh.PrivateKeyPath),
			Username:       getStringValue(tfSsh.Username),
			Port:           getStringValue(tfSsh.Port),
			Passphrase:     getStringValue(tfSsh.Passphrase),
		}

	}

	return hub, diags
}

func ConvertMksClusterToHub(ctx context.Context, tf *MksClusterModel, hub *infrapb.Cluster) diag.Diagnostics {
	var diags diag.Diagnostics

	hub.Kind = tf.Kind.ValueString()
	hub.ApiVersion = tf.ApiVersion.ValueString()

	hub.Metadata, diags = ConvertMetadataToHub(ctx, &tf.Metadata)
	if diags.HasError() {
		return diags
	}

	var blueprintType BlueprintType

	tfBlueprintValue, err := blueprintType.ValueFromObject(ctx, tf.Spec.Blueprint)

	if err != nil {
		return err
	}
	tfBlueprint := tfBlueprintValue.(BlueprintValue)

	hub.Spec = &infrapb.ClusterSpec{
		Blueprint: &infrapb.ClusterBlueprint{
			Name:    getStringValue(tfBlueprint.Name),
			Version: getStringValue(tfBlueprint.Version),
		},
	}

	hub.Spec.Type = getStringValue(tf.Spec.SpecType)

	if !tf.Spec.CloudCredentials.IsNull() && !tf.Spec.CloudCredentials.IsUnknown() {
		hub.Spec.CloudCredentials = getStringValue(tf.Spec.CloudCredentials)
	}

	if !tf.Spec.Sharing.IsNull() && !tf.Spec.Sharing.IsUnknown() {
		hub.Spec.Sharing, diags = ConvertMksSharingToHub(ctx, tf.Spec.Sharing)
		if diags.HasError() {
			return diags
		}
	}

	if !tf.Spec.SystemComponentsPlacement.IsNull() && !tf.Spec.SystemComponentsPlacement.IsUnknown() {
		hub.Spec.SystemComponentsPlacement, diags = ConvertMksSystemComponentsPlacementToHub(ctx, tf.Spec.SystemComponentsPlacement)
		if diags.HasError() {
			fmt.Println("Error in System Components Placement", diags)
			return diags
		}
	}

	if tf.Spec.Config.IsNull() || tf.Spec.Config.IsUnknown() {
		diags.AddError("config", "Config is required")
		return diags
	}

	fmt.Println("TF Config", tf.Spec.Config.Attributes())

	mksConfig, diags := ConvertMksConfigToHub(ctx, tf.Spec.Config)
	if diags.HasError() {
		return diags
	}

	hub.Spec.Config = &infrapb.ClusterSpec_Mks{
		Mks: mksConfig,
	}

	return diags

}

// Conversion of Hub Metadata to Terraform MetadataModel

func ConvertMetadataFromHub(ctx context.Context, hub *commonpb.Metadata) (MetadataValue, diag.Diagnostics) {

	var diags diag.Diagnostics

	tf := MetadataValue{
		Annotations: convertToTfMap(hub.Annotations),
		Description: basetypes.NewStringValue(hub.Description),
		DisplayName: basetypes.NewStringValue(hub.DisplayName),
		Labels:      convertToTfMap(hub.Labels),
		Name:        basetypes.NewStringValue(hub.Name),
		Project:     basetypes.NewStringValue(hub.Project),
	}

	if hub.CreatedBy != nil {
		tfCreatedByValue := &CreatedByValue{
			Id:        basetypes.NewStringValue(hub.CreatedBy.Id),
			IsSsouser: basetypes.NewBoolValue(hub.CreatedBy.IsSSOUser),
			Username:  basetypes.NewStringValue(hub.CreatedBy.Username),
		}
		tfObj, diag := tfCreatedByValue.ToObjectValue(ctx)
		if diag.HasError() {
			return NewMetadataValueNull(), diag
		}
		tf.CreatedBy = tfObj
	}

	if hub.ModifiedBy != nil {
		tfModifiedByValue := &ModifiedByValue{
			Id:        basetypes.NewStringValue(hub.ModifiedBy.Id),
			IsSsouser: basetypes.NewBoolValue(hub.ModifiedBy.IsSSOUser),
			Username:  basetypes.NewStringValue(hub.ModifiedBy.Username),
		}
		tfObj, diag := tfModifiedByValue.ToObjectValue(ctx)
		if diag.HasError() {
			return NewMetadataValueNull(), diag
		}
		tf.ModifiedBy = tfObj
	}

	return tf, diags
}

func ConvertMksNetworkFromHub(ctx context.Context, hub *infrapb.MksClusterNetworking) (*NetworkValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	var tfNetwork NetworkValue

	// Handle CNI
	if hub.Cni != nil {
		tfCni := CniValue{
			Name:    types.StringValue(hub.Cni.Name),
			Version: types.StringValue(hub.Cni.Version),
		}

		tfCniObject, diag := tfCni.ToObjectValue(ctx)
		if diag.HasError() {
			return nil, diag
		}
		tfNetwork.Cni = tfCniObject
	}

	// Handle PodSubnet
	if hub.PodSubnet != "" {
		tfNetwork.PodSubnet = types.StringValue(hub.PodSubnet)
	}

	// Handle ServiceSubnet
	if hub.ServiceSubnet != "" {
		tfNetwork.ServiceSubnet = types.StringValue(hub.ServiceSubnet)
	}

	// Handle IPv6
	if hub.Ipv6 != nil {
		tfIpv6 := Ipv6Value{
			PodSubnet:     types.StringValue(hub.Ipv6.PodSubnet),
			ServiceSubnet: types.StringValue(hub.Ipv6.ServiceSubnet),
		}

		tfIpv6Object, diag := tfIpv6.ToObjectValue(ctx)
		if diag.HasError() {
			return nil, diag
		}
		tfNetwork.Ipv6 = tfIpv6Object
	} else {
		tfNetwork.Ipv6, diags = NewIpv6ValueNull().ToObjectValue(ctx)
		if diags.HasError() {
			return &tfNetwork, diags
		}
	}

	return &tfNetwork, diags
}

func ConvertMksNodeFromHub(ctx context.Context, hub *infrapb.MksNode) (*NodesValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	tf := &NodesValue{
		Arch:            types.StringValue(hub.Arch),
		Hostname:        types.StringValue(hub.Hostname),
		OperatingSystem: types.StringValue(hub.OperatingSystem),
		PrivateIp:       types.StringValue(hub.PrivateIP),
	}

	if hub.Interface != "" {
		tf.Interface = types.StringValue(hub.Interface)
	}

	var tfRoles []attr.Value // Create a slice of attr.Value (which includes types.String)

	for _, role := range hub.Roles {
		tfRoles = append(tfRoles, types.StringValue(role))
		// Convert slice to ListValue
		tf.Roles, diags = types.ListValue(types.StringType, tfRoles)
		if diags.HasError() {
			return nil, diags
		}
	}

	if hub.Labels != nil {
		tf.Labels = convertToTfMap(hub.Labels)
	}

	if hub.Taints != nil {
		var tfTaints []attr.Value
		for _, taint := range hub.Taints {
			t, diag := ConvertTaintFromHub(ctx, taint)
			if diag.HasError() {
				return nil, diag
			}
			tfTaints = append(tfTaints, t)
		}
		tf.Taints, diags = types.ListValue(TaintsType{}, tfTaints)
		if diags.HasError() {
			log.Println("ConvertFrom Error in Taints", diags)
			return nil, diags
		}

	}

	if hub.Ssh != nil {
		tfSsh := SshValue{
			IpAddress:      types.StringValue(hub.Ssh.IpAddress),
			Passphrase:     types.StringValue(hub.Ssh.Passphrase),
			Port:           types.StringValue(hub.Ssh.Port),
			PrivateKeyPath: types.StringValue(hub.Ssh.PrivateKeyPath),
			Username:       types.StringValue(hub.Ssh.Username),
		}

		tfSshObject, diag := tfSsh.ToObjectValue(ctx)
		if diag.HasError() {
			return nil, diag
		}
		tf.Ssh = tfSshObject
	}

	return tf, diags
}

func ConvertTaintFromHub(ctx context.Context, hub *v1.Taint) (*TaintsValue, diag.Diagnostics) {
	tf := &TaintsValue{
		Effect: types.StringValue(string(hub.Effect)),
		Key:    types.StringValue(hub.Key),
		Value:  types.StringValue(hub.Value),
	}

	return tf, nil
}

func ConvertMksProxyFromHub(ctx context.Context, hub *infrapb.ClusterProxy) (*ProxyValue, diag.Diagnostics) {
	tf := &ProxyValue{
		AllowInsecureBootstrap: types.BoolValue(hub.AllowInsecureBootstrap),
		BootstrapCa:            types.StringValue(hub.BootstrapCA),
		Enabled:                types.BoolValue(hub.Enabled),
		HttpProxy:              types.StringValue(hub.HttpProxy),
		HttpsProxy:             types.StringValue(hub.HttpsProxy),
		NoProxy:                types.StringValue(hub.NoProxy),
		ProxyAuth:              types.StringValue(hub.ProxyAuth),
	}

	return tf, nil
}

func ConvertMksSharingFromHub(ctx context.Context, hub *infrapb.Sharing) (SharingValue, diag.Diagnostics) {
	var diag diag.Diagnostics

	tf := SharingValue{
		Enabled: types.BoolValue(hub.Enabled),
	}

	var tfProjects []attr.Value

	for _, project := range hub.Projects {
		tfProjects = append(tfProjects, types.StringValue(project.Name))
	}

	tf.Projects, diag = types.ListValue(types.StringType, tfProjects)

	return tf, diag
}

func ConvertMksSystemComponentsPlacementFromHub(ctx context.Context, hub *infrapb.SystemComponentsPlacement) (basetypes.ObjectValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	tf := &SystemComponentsPlacementValue{
		NodeSelector: convertToTfMap(hub.NodeSelector),
	}

	var tfTolerations []attr.Value

	for _, toleration := range hub.Tolerations {
		t, diag := ConvertTolerationsFromHub(ctx, toleration)
		if diag.HasError() {
			return NewSystemComponentsPlacementValueNull().ToObjectValue(ctx)
		}
		tfTolerations = append(tfTolerations, t)
	}

	tf.Tolerations, diags = types.ListValue(TolerationsType{}, tfTolerations)
	if diags.HasError() {
		return NewSystemComponentsPlacementValueNull().ToObjectValue(ctx)
	}

	if hub.DaemonSetOverride != nil {
		tfDaemonSet := DaemonSetOverrideValue{
			NodeSelectionEnabled: types.BoolValue(hub.DaemonSetOverride.NodeSelectionEnabled),
		}

		var tfDaemonSetTolerations []attr.Value

		for _, toleration := range hub.DaemonSetOverride.Tolerations {
			t, diag := ConvertDaemonSetTolerationsFromHub(ctx, toleration)
			if diag.HasError() {
				return NewSystemComponentsPlacementValueNull().ToObjectValue(ctx)
			}
			tfDaemonSetTolerations = append(tfDaemonSetTolerations, t)
		}

		tfDaemonSet.DaemonSetTolerations, diags = types.ListValue(DaemonSetTolerationsType{}, tfDaemonSetTolerations)
		if diags.HasError() {
			return NewSystemComponentsPlacementValueNull().ToObjectValue(ctx)
		}

		tf.DaemonSetOverride, diags = tfDaemonSet.ToObjectValue(ctx)
		if diags.HasError() {
			return NewSystemComponentsPlacementValueNull().ToObjectValue(ctx)
		}
	}

	return tf.ToObjectValue(ctx)
}

func ConvertTolerationsFromHub(ctx context.Context, hub *v1.Toleration) (*TolerationsValue, diag.Diagnostics) {
	tf := &TolerationsValue{
		Effect:            types.StringValue(string(hub.Effect)),
		Key:               types.StringValue(hub.Key),
		Operator:          types.StringValue(string(hub.Operator)),
		TolerationSeconds: types.Int64PointerValue(hub.TolerationSeconds),
		Value:             types.StringValue(hub.Value),
	}

	return tf, nil
}

func ConvertDaemonSetTolerationsFromHub(ctx context.Context, hub *v1.Toleration) (*DaemonSetTolerationsValue, diag.Diagnostics) {
	tf := &DaemonSetTolerationsValue{
		Effect:            types.StringValue(string(hub.Effect)),
		Key:               types.StringValue(hub.Key),
		Operator:          types.StringValue(string(hub.Operator)),
		TolerationSeconds: types.Int64PointerValue(hub.TolerationSeconds),
		Value:             types.StringValue(hub.Value),
	}

	return tf, nil
}

func ConvertMksConfigFromHub(ctx context.Context, hub *infrapb.MksV3ConfigObject) (*ConfigValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	tf := &ConfigValue{
		Location:              types.StringValue(hub.Location),
		AutoApproveNodes:      types.BoolValue(hub.AutoApproveNodes),
		DedicatedControlPlane: types.BoolValue(hub.DedicatedControlPlane),
		HighAvailability:      types.BoolValue(hub.HighAvailability),
		KubernetesVersion:     types.StringValue(hub.KubernetesVersion),
	}

	tfNetwork, diags := ConvertMksNetworkFromHub(ctx, hub.Network)
	if diags.HasError() {
		return tf, diags
	}

	tf.Network, diags = tfNetwork.ToObjectValue(ctx)
	if diags.HasError() {
		return tf, diags
	}

	var tfNodes []attr.Value

	for _, node := range hub.Nodes {
		n, diag := ConvertMksNodeFromHub(ctx, node)
		if diag.HasError() {
			return tf, diag
		}
		tfNodes = append(tfNodes, n)
	}

	tf.Nodes, diags = types.ListValue(NodesType{}, tfNodes)
	if diags.HasError() {
		return tf, diags
	}

	if hub.Ssh != nil {
		tfSsh := ClusterSshValue{
			PrivateKeyPath: types.StringValue(hub.Ssh.PrivateKeyPath),
			Username:       types.StringValue(hub.Ssh.Username),
			Port:           types.StringValue(hub.Ssh.Port),
			Passphrase:     types.StringValue(hub.Ssh.Passphrase),
		}

		tfSshObject, diag := tfSsh.ToObjectValue(ctx)
		if diag.HasError() {
			return tf, diag
		}
		tf.ClusterSsh = tfSshObject
	}

	return tf, diags
}

func ConvertMksClusterFromHub(ctx context.Context, hub *infrapb.Cluster, tf *MksClusterModel) diag.Diagnostics {
	var diags diag.Diagnostics

	tf.Kind = types.StringValue(hub.Kind)
	tf.ApiVersion = types.StringValue(hub.ApiVersion)

	tf.Metadata, diags = ConvertMetadataFromHub(ctx, hub.Metadata)
	if diags.HasError() {
		return diags
	}

	tfBlueprint := BlueprintValue{
		Name:    types.StringValue(hub.Spec.Blueprint.Name),
		Version: types.StringValue(hub.Spec.Blueprint.Version),
	}

	tf.Spec.SpecType = types.StringValue(hub.Spec.Type)

	tf.Spec.Blueprint, diags = tfBlueprint.ToObjectValue(ctx)
	if diags.HasError() {
		return diags
	}

	if hub.Spec.CloudCredentials != "" {
		tf.Spec.CloudCredentials = types.StringValue(hub.Spec.CloudCredentials)
	}

	if hub.Spec.Sharing != nil {
		sharingValue, diags := ConvertMksSharingFromHub(ctx, hub.Spec.Sharing)
		if diags.HasError() {
			return diags
		}
		tf.Spec.Sharing, diags = sharingValue.ToObjectValue(ctx)
		if diags.HasError() {
			return diags
		}
	}

	if hub.Spec.SystemComponentsPlacement != nil {
		tf.Spec.SystemComponentsPlacement, diags = ConvertMksSystemComponentsPlacementFromHub(ctx, hub.Spec.SystemComponentsPlacement)
		if diags.HasError() {
			return diags
		}

	}

	if hub.Spec.Config != nil {
		hubConfig := hub.Spec.Config.(*infrapb.ClusterSpec_Mks).Mks

		tfConfig, diags := ConvertMksConfigFromHub(ctx, hubConfig)
		if diags.HasError() {
			return diags
		}
		tf.Spec.Config, diags = tfConfig.ToObjectValue(ctx)
		if diags.HasError() {
			return diags
		}
	}

	return diags
}
