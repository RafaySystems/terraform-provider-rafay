package convertor

import (
	"context"
	"fmt"

	"github.com/RafaySystems/rafay-common/proto/types/hub/commonpb"
	"github.com/RafaySystems/rafay-common/proto/types/hub/infrapb"
	v1 "k8s.io/api/core/v1"

	fw "github.com/RafaySystems/terraform-provider-rafay/internal/gen/resource_mks_cluster"
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

	if tfMap.IsNull() || tfMap.IsUnknown() {
		return result
	}
	for k, v := range tfMap.Elements() {
		result[k] = getStringValue(v.(types.String))
	}
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

func tfTaintKey(taint *fw.TaintsValue) string {
	return fmt.Sprintf("%s:%s:%s", taint.Key.ValueString(), taint.Value.ValueString(), taint.Effect.ValueString())
}

func hubTaintKey(taint *v1.Taint) string {
	return fmt.Sprintf("%s:%s:%s", taint.Key, taint.Value, taint.Effect)
}

func tfTolerationsKey(t *fw.TolerationsValue) string {
	return fmt.Sprintf("%s:%s:%s:%s", t.Key.ValueString(), t.Value.ValueString(), t.Operator.ValueString(), t.Effect.ValueString())
}

func hubTolerationsKey(t *v1.Toleration) string {
	return fmt.Sprintf("%s:%s:%s:%s", t.Key, t.Value, t.Operator, t.Effect)
}

func tfDaemonSetTolerationsKey(t *fw.DaemonSetTolerationsValue) string {
	return fmt.Sprintf("%s:%s:%s:%s", t.Key.ValueString(), t.Value.ValueString(), t.Operator.ValueString(), t.Effect.ValueString())
}

// Conversion of Terrafrom MetadataModel to commonpb.Metadata
func convertMksMetadataToHub(tf *fw.MetadataValue, hub *commonpb.Metadata) (diags diag.Diagnostics) {

	hub.Name = getStringValue(tf.Name)
	hub.Project = getStringValue(tf.Project)

	if !tf.Annotations.IsNull() && !tf.Annotations.IsUnknown() {
		hub.Annotations = convertFromTfMap(tf.Annotations)
	}

	if !tf.Description.IsNull() && !tf.Description.IsUnknown() {
		hub.Description = getStringValue(tf.Description)
	}

	if !tf.Labels.IsNull() && !tf.Labels.IsUnknown() {
		hub.Labels = convertFromTfMap(tf.Labels)
	}

	return
}

func convertMksNetworkToHub(ctx context.Context, tf basetypes.ObjectValue, hub *infrapb.MksClusterNetworking) (diags diag.Diagnostics) {

	var tfNetworkType fw.NetworkType

	tfNetworkValue, diags := tfNetworkType.ValueFromObject(ctx, tf)
	if diags.HasError() {
		return diags
	}

	tfNetwork := tfNetworkValue.(fw.NetworkValue)
	var cniType fw.CniType
	tfCniValue, diags := cniType.ValueFromObject(ctx, tfNetwork.Cni)

	if diags.HasError() {
		return diags
	}

	tfCni := tfCniValue.(fw.CniValue)
	hubCni := &infrapb.Cni{
		Name:    getStringValue(tfCni.Name),
		Version: getStringValue(tfCni.Version),
	}

	hub.Cni = hubCni

	hub.PodSubnet = getStringValue(tfNetwork.PodSubnet)
	hub.ServiceSubnet = getStringValue(tfNetwork.ServiceSubnet)

	if !tfNetwork.Ipv6.IsNull() && !tfNetwork.Ipv6.IsUnknown() {
		// Handle IPv6
		var ipv6Type fw.Ipv6Type
		tfIpv6Value, diags := ipv6Type.ValueFromObject(ctx, tfNetwork.Ipv6)
		if diags.HasError() {
			return diags
		}
		tfIpv6 := tfIpv6Value.(fw.Ipv6Value)

		hub.Ipv6 = &infrapb.MksSubnet{
			PodSubnet:     getStringValue(tfIpv6.PodSubnet),
			ServiceSubnet: getStringValue(tfIpv6.ServiceSubnet),
		}
	}
	return diags
}

func convertMksNodeToHub(ctx context.Context, tf *fw.NodesValue, hub *infrapb.MksNode) (diags diag.Diagnostics) {
	var d diag.Diagnostics

	hub.Arch = getStringValue(tf.Arch)
	hub.Hostname = getStringValue(tf.Hostname)
	hub.OperatingSystem = getStringValue(tf.OperatingSystem)
	hub.PrivateIP = getStringValue(tf.PrivateIp)

	if !tf.Interface.IsNull() && !tf.Interface.IsUnknown() {
		hub.Interface = tf.Interface.ValueString()
	}

	for _, role := range tf.Roles.Elements() {
		hub.Roles = append(hub.Roles, getStringValue(role.(types.String)))
	}

	if !tf.Labels.IsNull() && !tf.Labels.IsUnknown() {
		hub.Labels = convertFromTfMap(tf.Labels)
	}

	if !tf.Taints.IsNull() && !tf.Taints.IsUnknown() {
		for _, taint := range tf.Taints.Elements() {
			var t v1.Taint
			d = convertMksNodeTaintToHub(taint.(fw.TaintsValue), &t)
			diags = append(diags, d...)
			hub.Taints = append(hub.Taints, &t)
		}
	}

	if !tf.Ssh.IsNull() && !tf.Ssh.IsUnknown() {
		hub.Ssh = &infrapb.MksNodeSshConfig{}
		d = convertMksNodeSshToHub(ctx, tf.Ssh, hub.Ssh)
		diags = append(diags, d...)
	}

	return
}

func convertMksNodeSshToHub(ctx context.Context, tf basetypes.ObjectValue, hub *infrapb.MksNodeSshConfig) (diags diag.Diagnostics) {
	var sshType fw.SshType

	tfSshValue, d := sshType.ValueFromObject(ctx, tf)
	if d.HasError() {
		return d
	}

	tfSsh := tfSshValue.(fw.SshValue)
	if !tfSsh.IpAddress.IsNull() && !tfSsh.IpAddress.IsUnknown() {
		hub.IpAddress = getStringValue(tfSsh.IpAddress)
	}
	if !tfSsh.Passphrase.IsNull() && !tfSsh.Passphrase.IsUnknown() {
		hub.Passphrase = getStringValue(tfSsh.Passphrase)
	}
	if !tfSsh.Port.IsNull() && !tfSsh.Port.IsUnknown() {
		hub.Port = getStringValue(tfSsh.Port)
	}
	if !tfSsh.PrivateKeyPath.IsNull() && !tfSsh.PrivateKeyPath.IsUnknown() {
		hub.PrivateKeyPath = getStringValue(tfSsh.PrivateKeyPath)
	}
	if !tfSsh.Username.IsNull() && !tfSsh.Username.IsUnknown() {
		hub.Username = getStringValue(tfSsh.Username)
	}

	return
}

func convertMksNodeTaintToHub(tf fw.TaintsValue, hub *v1.Taint) (diags diag.Diagnostics) {
	if !tf.Effect.IsNull() && !tf.Effect.IsUnknown() {
		hub.Effect = v1.TaintEffect(getStringValue(tf.Effect))
	}

	if !tf.Key.IsNull() && !tf.Key.IsUnknown() {
		hub.Key = getStringValue(tf.Key)
	}

	if !tf.Value.IsNull() && !tf.Value.IsUnknown() {
		hub.Value = getStringValue(tf.Value)
	}

	return
}

func convertMksProxyToHub(ctx context.Context, tf basetypes.ObjectValue, hub *infrapb.ClusterProxy) (diags diag.Diagnostics) {
	var proxyType fw.ProxyType

	tfProxyValue, d := proxyType.ValueFromObject(ctx, tf)
	if d.HasError() {
		return d
	}

	tfProxy := tfProxyValue.(fw.ProxyValue)

	if !tfProxy.AllowInsecureBootstrap.IsNull() && !tfProxy.AllowInsecureBootstrap.IsUnknown() {
		hub.AllowInsecureBootstrap = getBoolValue(tfProxy.AllowInsecureBootstrap)
	}

	if !tfProxy.BootstrapCa.IsNull() && !tfProxy.BootstrapCa.IsUnknown() {
		hub.BootstrapCA = getStringValue(tfProxy.BootstrapCa)
	}

	if !tfProxy.Enabled.IsNull() && !tfProxy.Enabled.IsUnknown() {
		hub.Enabled = getBoolValue(tfProxy.Enabled)
	}

	if !tfProxy.HttpProxy.IsNull() && !tfProxy.HttpProxy.IsUnknown() {
		hub.HttpProxy = getStringValue(tfProxy.HttpProxy)
	}

	if !tfProxy.HttpsProxy.IsNull() && !tfProxy.HttpsProxy.IsUnknown() {
		hub.HttpsProxy = getStringValue(tfProxy.HttpsProxy)
	}

	if !tfProxy.NoProxy.IsNull() && !tfProxy.NoProxy.IsUnknown() {
		hub.NoProxy = getStringValue(tfProxy.NoProxy)
	}
	return
}

func convertMksSharingToHub(ctx context.Context, tf basetypes.ObjectValue, hub *infrapb.Sharing) (diags diag.Diagnostics) {
	var sharingType fw.SharingType

	tfSharingValue, err := sharingType.ValueFromObject(ctx, tf)
	if err != nil {
		return err
	}
	tfSharing := tfSharingValue.(fw.SharingValue)

	hub.Enabled = getBoolValue(tfSharing.Enabled)

	for _, project := range tfSharing.Projects.Elements() {
		hub.Projects = append(hub.Projects, &infrapb.Projects{
			Name: getStringValue(project.(types.String)),
		})
	}

	return
}

func convertMksSystemComponentsPlacementToHub(ctx context.Context, tf basetypes.ObjectValue, hub *infrapb.SystemComponentsPlacement) diag.Diagnostics {
	var systemComponentsPlacementType fw.SystemComponentsPlacementType

	tfObject, diag := systemComponentsPlacementType.ValueFromObject(ctx, tf)
	if diag.HasError() {
		return diag
	}

	tfValue := tfObject.(fw.SystemComponentsPlacementValue)

	if !tfValue.NodeSelector.IsNull() && !tfValue.NodeSelector.IsUnknown() {
		hub.NodeSelector = convertFromTfMap(tfValue.NodeSelector)
	}

	for _, toleration := range tfValue.Tolerations.Elements() {
		t := &v1.Toleration{}
		diag := convertMksTolerationsToHub(toleration.(fw.TolerationsValue), t)
		if diag.HasError() {
			return diag
		}
		hub.Tolerations = append(hub.Tolerations, t)
	}

	if !tfValue.DaemonSetOverride.IsNull() && !tfValue.DaemonSetOverride.IsUnknown() {
		var daemonSetType fw.DaemonSetOverrideType

		tfDaemonSetValue, daig := daemonSetType.ValueFromObject(ctx, tfValue.DaemonSetOverride)
		if daig.HasError() {
			return daig
		}

		tfDaemonSet := tfDaemonSetValue.(fw.DaemonSetOverrideValue)

		hub.DaemonSetOverride = &infrapb.DaemonSetOverride{
			NodeSelectionEnabled: getBoolValue(tfDaemonSet.NodeSelectionEnabled),
		}

		for _, toleration := range tfDaemonSet.DaemonSetTolerations.Elements() {
			t := &v1.Toleration{}
			diag := convertMksDaemonSetTolerationsToHub(toleration.(fw.DaemonSetTolerationsValue), t)
			if diag.HasError() {
				return diag
			}
			hub.DaemonSetOverride.Tolerations = append(hub.DaemonSetOverride.Tolerations, t)
		}
	}

	return diag
}

func convertMksTolerationsToHub(tf fw.TolerationsValue, hub *v1.Toleration) (daig diag.Diagnostics) {
	if !tf.Effect.IsNull() && !tf.Effect.IsUnknown() {
		hub.Effect = v1.TaintEffect(getStringValue(tf.Effect))
	}

	if !tf.Key.IsNull() && !tf.Key.IsUnknown() {
		hub.Key = getStringValue(tf.Key)
	}

	if !tf.Value.IsNull() && !tf.Value.IsUnknown() {
		hub.Value = getStringValue(tf.Value)
	}

	if !tf.Operator.IsNull() && !tf.Operator.IsUnknown() {
		hub.Operator = v1.TolerationOperator(getStringValue(tf.Operator))
	}

	if !tf.TolerationSeconds.IsNull() && !tf.TolerationSeconds.IsUnknown() {
		hub.TolerationSeconds = tf.TolerationSeconds.ValueInt64Pointer()
	}
	return
}

func convertMksDaemonSetTolerationsToHub(tf fw.DaemonSetTolerationsValue, hub *v1.Toleration) (diags diag.Diagnostics) {

	if !tf.Effect.IsNull() && !tf.Effect.IsUnknown() {
		hub.Effect = v1.TaintEffect(getStringValue(tf.Effect))
	}

	if !tf.Key.IsNull() && !tf.Key.IsUnknown() {
		hub.Key = getStringValue(tf.Key)
	}

	if !tf.Value.IsNull() && !tf.Value.IsUnknown() {
		hub.Value = getStringValue(tf.Value)
	}

	if !tf.Operator.IsNull() && !tf.Operator.IsUnknown() {
		hub.Operator = v1.TolerationOperator(getStringValue(tf.Operator))
	}

	if !tf.TolerationSeconds.IsNull() && !tf.TolerationSeconds.IsUnknown() {
		hub.TolerationSeconds = tf.TolerationSeconds.ValueInt64Pointer()
	}

	return

}

func convertMksConfigToHub(ctx context.Context, tf basetypes.ObjectValue) (*infrapb.MksV3ConfigObject, diag.Diagnostics) {

	var configType fw.ConfigType

	tfMksConfigValue, diags := configType.ValueFromObject(ctx, tf)
	if diags.HasError() {
		return nil, diags
	}

	tfMksConfig := tfMksConfigValue.(fw.ConfigValue)
	hub := &infrapb.MksV3ConfigObject{
		Location:              getStringValue(tfMksConfig.Location),
		AutoApproveNodes:      getBoolValue(tfMksConfig.AutoApproveNodes),
		DedicatedControlPlane: getBoolValue(tfMksConfig.DedicatedControlPlane),
		HighAvailability:      getBoolValue(tfMksConfig.HighAvailability),
		KubernetesVersion:     getStringValue(tfMksConfig.KubernetesVersion),
	}

	hub.Network = &infrapb.MksClusterNetworking{}
	diags = convertMksNetworkToHub(ctx, tfMksConfig.Network, hub.Network)
	if diags.HasError() {
		return nil, diags
	}

	for _, node := range tfMksConfig.Nodes.Elements() {
		val := node.(fw.NodesValue)
		n := &infrapb.MksNode{}
		diags = convertMksNodeToHub(ctx, &val, n)
		if diags.HasError() {
			return nil, diags
		}
		hub.Nodes = append(hub.Nodes, n)
	}

	if !tfMksConfig.ClusterSsh.IsNull() && !tfMksConfig.ClusterSsh.IsUnknown() {
		hub.Ssh = &infrapb.MksClusterSshConfig{}
		diags = convertMksClusterSshToHub(ctx, tfMksConfig.ClusterSsh, hub.Ssh)
		if diags.HasError() {
			return nil, diags
		}

	}

	return hub, diags
}

func convertMksClusterSshToHub(ctx context.Context, tf basetypes.ObjectValue, hub *infrapb.MksClusterSshConfig) (diags diag.Diagnostics) {

	var sshType fw.ClusterSshType

	tfSshValue, d := sshType.ValueFromObject(ctx, tf)
	if d.HasError() {
		return d
	}

	tfSsh := tfSshValue.(fw.ClusterSshValue)

	if !tfSsh.PrivateKeyPath.IsNull() && !tfSsh.PrivateKeyPath.IsUnknown() {
		hub.PrivateKeyPath = getStringValue(tfSsh.PrivateKeyPath)
	}

	if !tfSsh.Username.IsNull() && !tfSsh.Username.IsUnknown() {
		hub.Username = getStringValue(tfSsh.Username)
	}

	if !tfSsh.Port.IsNull() && !tfSsh.Port.IsUnknown() {
		hub.Port = getStringValue(tfSsh.Port)
	}

	if !tfSsh.Passphrase.IsNull() && !tfSsh.Passphrase.IsUnknown() {
		hub.Passphrase = getStringValue(tfSsh.Passphrase)
	}

	return
}

// Conversion of Hub Metadata to Terraform MetadataModel
func convertMetadataFromHub(hub *commonpb.Metadata, tf *fw.MetadataValue) diag.Diagnostics {
	var diags diag.Diagnostics

	tf.Name = types.StringValue(hub.Name)
	tf.Project = types.StringValue(hub.Project)
	if hub.Description != "" {
		tf.Description = types.StringValue(hub.Description)
	}

	tf.Annotations = convertToTfMap(hub.Annotations)
	tf.Labels = convertToTfMap(hub.Labels)

	return diags
}

func convertMksNetworkFromHub(ctx context.Context, hub *infrapb.MksClusterNetworking, tf *basetypes.ObjectValue) (basetypes.ObjectValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	newTf, d := fw.NewNetworkValue(tf.AttributeTypes(ctx), tf.Attributes())
	if d.HasError() {
		// return old object and diagnostics
		return *tf, d
	}

	tfCni, d := fw.NewCniValue(newTf.Cni.AttributeTypes(ctx), newTf.Cni.Attributes())
	if d.HasError() {
		// return old object and diagnostics
		return *tf, d
	}
	// Handle CNI
	tfCni.Name = types.StringValue(hub.Cni.Name)
	tfCni.Version = types.StringValue(hub.Cni.Version)

	tfCniObj, d := tfCni.ToObjectValue(ctx)
	diags = append(diags, d...)
	newTf.Cni = tfCniObj

	newTf.PodSubnet = types.StringValue(hub.PodSubnet)
	newTf.ServiceSubnet = types.StringValue(hub.ServiceSubnet)

	// Handle IPv6
	if hub.Ipv6 != nil {
		tfIpv6 := fw.Ipv6Value{
			PodSubnet:     types.StringValue(hub.Ipv6.PodSubnet),
			ServiceSubnet: types.StringValue(hub.Ipv6.ServiceSubnet),
		}

		tfIpv6Object, d := tfIpv6.ToObjectValue(ctx)
		diags = append(diags, d...)
		newTf.Ipv6 = tfIpv6Object
	}
	newTfObj, d := newTf.ToObjectValue(ctx)
	diags = append(diags, d...)
	return newTfObj, diags
}

func convertMksNodeFromHub(ctx context.Context, hub *infrapb.MksNode, tf *fw.NodesValue) (diags diag.Diagnostics) {
	var d diag.Diagnostics

	tf.Arch = types.StringValue(hub.Arch)
	tf.Hostname = types.StringValue(hub.Hostname)
	tf.OperatingSystem = types.StringValue(hub.OperatingSystem)
	tf.PrivateIp = types.StringValue(hub.PrivateIP)

	if hub.Interface != "" {
		tf.Interface = types.StringValue(hub.Interface)
	}

	tfRoles := []attr.Value{}
	for _, role := range hub.Roles {
		tfRoles = append(tfRoles, types.StringValue(role))
		tf.Roles, diags = types.SetValue(types.StringType, tfRoles)
	}

	if hub.Labels != nil {
		tf.Labels = convertToTfMap(hub.Labels)
	}

	// Construct map with current taints in terraform state
	tfTaintMap := make(map[string]*fw.TaintsValue)
	for _, taint := range tf.Taints.Elements() {
		tfTaint := taint.(fw.TaintsValue)
		tfTaintMap[tfTaintKey(&tfTaint)] = &tfTaint
	}

	if hub.Taints != nil {
		var tfTaints []attr.Value
		for _, hub := range hub.Taints {
			tfTaint, ok := tfTaintMap[hubTaintKey(hub)]
			if !ok {
				tfTaint = &fw.TaintsValue{}
			}
			d = convertTaintFromHub(hub, tfTaint)
			if d.HasError() {
				diags = append(diags, d...)
			}
			tfTaints = append(tfTaints, tfTaint)
		}

		tfTaintsType := fw.TaintsType{
			ObjectType: types.ObjectType{
				AttrTypes: fw.TaintsValue{}.AttributeTypes(ctx),
			},
		}
		tf.Taints, d = types.SetValue(tfTaintsType, tfTaints)
		diags = append(diags, d...)
	}

	if hub.Ssh != nil {
		tf.Ssh, d = convertMksNodeSshFromHub(ctx, hub.Ssh, &tf.Ssh)
		diags = append(diags, d...)
	}
	return diags

}

func convertMksNodeSshFromHub(ctx context.Context, hub *infrapb.MksNodeSshConfig, tf *basetypes.ObjectValue) (basetypes.ObjectValue, diag.Diagnostics) {
	newTf, d := fw.NewSshValue(tf.AttributeTypes(ctx), tf.Attributes())
	if d.HasError() {
		// return old object and diagnostics
		return *tf, d
	}

	if hub.IpAddress != "" {
		newTf.IpAddress = types.StringValue(hub.IpAddress)
	}

	if hub.Passphrase != "" {
		newTf.Passphrase = types.StringValue(hub.Passphrase)
	}

	if hub.Port != "" {
		newTf.Port = types.StringValue(hub.Port)
	}

	if hub.PrivateKeyPath != "" {
		newTf.PrivateKeyPath = types.StringValue(hub.PrivateKeyPath)
	}

	if hub.Username != "" {
		newTf.Username = types.StringValue(hub.Username)
	}

	return newTf.ToObjectValue(ctx)
}

func convertTaintFromHub(hub *v1.Taint, tf *fw.TaintsValue) diag.Diagnostics {
	var diags diag.Diagnostics

	if hub.Effect != "" {
		tf.Effect = types.StringValue(string(hub.Effect))
	}
	if hub.Key != "" {
		tf.Key = types.StringValue(hub.Key)
	}
	if hub.Value != "" {
		tf.Value = types.StringValue(hub.Value)
	}

	return diags
}

func convertMksProxyFromHub(ctx context.Context, hub *infrapb.ClusterProxy, tf *basetypes.ObjectValue) (basetypes.ObjectValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	newTf, d := fw.NewProxyValue(tf.AttributeTypes(ctx), tf.Attributes())
	if d.HasError() {
		// return old object and diagnostics
		return *tf, d
	}

	if hub.AllowInsecureBootstrap {
		newTf.AllowInsecureBootstrap = types.BoolPointerValue(&hub.AllowInsecureBootstrap)
	}

	if hub.BootstrapCA != "" {
		newTf.BootstrapCa = types.StringValue(hub.BootstrapCA)
	}

	if hub.Enabled {
		newTf.Enabled = types.BoolPointerValue(&hub.Enabled)
	}

	if hub.HttpProxy != "" {
		newTf.HttpProxy = types.StringValue(hub.HttpProxy)
	}

	if hub.HttpsProxy != "" {
		newTf.HttpsProxy = types.StringValue(hub.HttpsProxy)
	}

	if hub.NoProxy != "" {
		newTf.NoProxy = types.StringValue(hub.NoProxy)
	}

	if hub.ProxyAuth != "" {
		newTf.ProxyAuth = types.StringValue(hub.ProxyAuth)
	}

	newTfObj, d := newTf.ToObjectValue(ctx)
	diags = append(diags, d...)
	return newTfObj, diags
}

func convertMksSharingFromHub(ctx context.Context, hub *infrapb.Sharing, tf *basetypes.ObjectValue) (basetypes.ObjectValue, diag.Diagnostics) {

	var diag diag.Diagnostics
	newTf, d := fw.NewSharingValue(tf.AttributeTypes(ctx), tf.Attributes())
	if d.HasError() {
		return *tf, d
	}

	newTf.Enabled = types.BoolPointerValue(&hub.Enabled)
	var tfProjects []attr.Value

	for _, project := range hub.Projects {
		tfProjects = append(tfProjects, types.StringValue(project.Name))
	}

	newTf.Projects, diag = types.SetValue(types.StringType, tfProjects)

	newTfObj, d := newTf.ToObjectValue(ctx)
	diag = append(diag, d...)
	return newTfObj, diag
}

func convertMksSystemComponentsPlacementFromHub(ctx context.Context, hub *infrapb.SystemComponentsPlacement, tf *basetypes.ObjectValue) (basetypes.ObjectValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	newTf, d := fw.NewSystemComponentsPlacementValue(tf.AttributeTypes(ctx), tf.Attributes())
	if d.HasError() {
		// return old object and diagnostics
		return *tf, d
	}
	if hub.NodeSelector != nil {
		newTf.NodeSelector = convertToTfMap(hub.NodeSelector)
	}

	// construct map with current tolerations in terraform state
	tfTolerationMap := make(map[string]*fw.TolerationsValue)
	for _, toleration := range newTf.Tolerations.Elements() {
		tfToleration := toleration.(fw.TolerationsValue)
		tfTolerationMap[tfTolerationsKey(&tfToleration)] = &tfToleration
	}

	if hub.Tolerations != nil {
		var tfTolerations []attr.Value

		for _, hub := range hub.Tolerations {
			tfToleration, ok := tfTolerationMap[hubTolerationsKey(hub)]
			if !ok {
				tfToleration = &fw.TolerationsValue{}
			}
			d := convertTolerationsFromHub(hub, tfToleration)
			diags = append(diags, d...)
			tfTolerations = append(tfTolerations, tfToleration)
		}

		tfTolerationsType := fw.TolerationsType{
			ObjectType: types.ObjectType{
				AttrTypes: fw.TolerationsValue{}.AttributeTypes(ctx),
			},
		}
		newTf.Tolerations, d = types.SetValue(tfTolerationsType, tfTolerations)

		diags = append(diags, d...)
	}

	if hub.DaemonSetOverride != nil {
		tfDaemonSet, d := fw.NewDaemonSetOverrideValue(newTf.DaemonSetOverride.AttributeTypes(ctx), newTf.DaemonSetOverride.Attributes())
		if d.HasError() {
			// return old object and diagnostics
			diags = append(diags, d...)
			return *tf, diags
		}

		tfDaemonSet.NodeSelectionEnabled = types.BoolPointerValue(&hub.DaemonSetOverride.NodeSelectionEnabled)

		tfDaemonSetTolerationsMap := make(map[string]*fw.DaemonSetTolerationsValue)
		for _, toleration := range tfDaemonSet.DaemonSetTolerations.Elements() {
			tfDaemonSetTol := toleration.(fw.DaemonSetTolerationsValue)
			tfDaemonSetTolerationsMap[tfDaemonSetTolerationsKey(&tfDaemonSetTol)] = &tfDaemonSetTol
		}

		var tfDaemonSetTolerations []attr.Value
		for _, hub := range hub.DaemonSetOverride.Tolerations {
			tfDsTol, ok := tfDaemonSetTolerationsMap[hubTolerationsKey(hub)]
			if !ok {
				tfDsTol = &fw.DaemonSetTolerationsValue{}
			}
			d = convertDaemonSetTolerationsFromHub(hub, tfDsTol)
			diags = append(diags, d...)
			tfDaemonSetTolerations = append(tfDaemonSetTolerations, tfDsTol)
		}

		tfDaemonSetTolerationsType := fw.DaemonSetTolerationsType{
			ObjectType: types.ObjectType{
				AttrTypes: fw.DaemonSetTolerationsValue{}.AttributeTypes(ctx),
			},
		}

		tfDaemonSet.DaemonSetTolerations, d = types.SetValue(tfDaemonSetTolerationsType, tfDaemonSetTolerations)
		diags = append(diags, d...)

		newTf.DaemonSetOverride, d = tfDaemonSet.ToObjectValue(ctx)
		diags = append(diags, d...)
	}

	newTfObj, d := newTf.ToObjectValue(ctx)
	diags = append(diags, d...)

	return newTfObj, diags
}

func convertTolerationsFromHub(hub *v1.Toleration, tf *fw.TolerationsValue) (diags diag.Diagnostics) {
	if hub.Effect != "" {
		tf.Effect = types.StringValue(string(hub.Effect))
	}
	if hub.Key != "" {
		tf.Key = types.StringValue(hub.Key)
	}

	if hub.Value != "" {
		tf.Value = types.StringValue(hub.Value)
	}
	if hub.Operator != "" {
		tf.Operator = types.StringValue(string(hub.Operator))
	}
	if hub.TolerationSeconds != nil {
		tf.TolerationSeconds = types.Int64PointerValue(hub.TolerationSeconds)
	}

	return
}

func convertDaemonSetTolerationsFromHub(hub *v1.Toleration, tf *fw.DaemonSetTolerationsValue) (diags diag.Diagnostics) {
	if hub.Effect != "" {
		tf.Effect = types.StringValue(string(hub.Effect))
	}
	if hub.Key != "" {
		tf.Key = types.StringValue(hub.Key)
	}

	if hub.Value != "" {
		tf.Value = types.StringValue(hub.Value)
	}
	if hub.Operator != "" {
		tf.Operator = types.StringValue(string(hub.Operator))
	}
	if hub.TolerationSeconds != nil {
		tf.TolerationSeconds = types.Int64PointerValue(hub.TolerationSeconds)
	}

	return
}

func convertMksConfigFromHub(ctx context.Context, hub *infrapb.MksV3ConfigObject, tf *basetypes.ObjectValue) (basetypes.ObjectValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	newTf, d := fw.NewConfigValue(tf.AttributeTypes(ctx), tf.Attributes())

	if d.HasError() {
		// return old object and diagnostics
		return *tf, d
	}
	if hub.Location != "" {
		newTf.Location = types.StringValue(hub.Location)
	}
	if hub.AutoApproveNodes {
		newTf.AutoApproveNodes = types.BoolPointerValue(&hub.AutoApproveNodes)
	}
	if hub.DedicatedControlPlane {
		newTf.DedicatedControlPlane = types.BoolPointerValue(&hub.DedicatedControlPlane)
	}
	if hub.HighAvailability {
		newTf.HighAvailability = types.BoolPointerValue(&hub.HighAvailability)
	}
	newTf.KubernetesVersion = types.StringValue(hub.KubernetesVersion)

	newTf.Network, d = convertMksNetworkFromHub(ctx, hub.Network, &newTf.Network)
	diags = append(diags, d...)

	// Incoming Nodes-source of truth
	hubNodeMap := make(map[string]*infrapb.MksNode)
	for _, node := range hub.GetNodes() {
		hubNodeMap[node.Hostname] = node
	}

	// Current Nodes in terraform
	tfNodeMap := make(map[string]*fw.NodesValue)
	for _, node := range newTf.Nodes.Elements() {
		tfNode := node.(fw.NodesValue)
		tfNodeMap[getStringValue(tfNode.Hostname)] = &tfNode
	}

	newTfNodes := make(map[string]attr.Value)

	// Compare the nodes in the hub and terraform
	for hostname, hubNode := range hubNodeMap {
		tfNode, ok := tfNodeMap[hostname]
		if !ok {
			tfNode = &fw.NodesValue{}
		}
		d = convertMksNodeFromHub(ctx, hubNode, tfNode)
		if d.HasError() {
			diags = append(diags, d...)
		} else {
			newTfNodes[hostname] = tfNode
		}

	}

	tfNodeType := fw.NodesType{
		ObjectType: types.ObjectType{
			AttrTypes: fw.NodesValue{}.AttributeTypes(ctx),
		},
	}
	newTf.Nodes, d = types.MapValue(tfNodeType, newTfNodes)
	diags = append(diags, d...)

	if hub.Ssh != nil {
		newTf.ClusterSsh, d = convertClusterSshFromHub(ctx, hub.Ssh, &newTf.ClusterSsh)
		diags = append(diags, d...)
	}

	newTfObj, d := newTf.ToObjectValue(ctx)
	diags = append(diags, d...)
	return newTfObj, diags
}

func convertClusterSshFromHub(ctx context.Context, hub *infrapb.MksClusterSshConfig, tf *basetypes.ObjectValue) (basetypes.ObjectValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	newTf, d := fw.NewClusterSshValue(tf.AttributeTypes(ctx), tf.Attributes())
	if d.HasError() {
		// return old object and diagnostics
		return *tf, d
	}

	newTf.PrivateKeyPath = types.StringValue(hub.PrivateKeyPath)
	newTf.Username = types.StringValue(hub.Username)
	newTf.Port = types.StringValue(hub.Port)
	newTf.Passphrase = types.StringValue(hub.Passphrase)

	newTfObj, d := newTf.ToObjectValue(ctx)
	diags = append(diags, d...)
	return newTfObj, diags
}

// ConvertMksClusterToHub converts a Terraform MksClusterModel to a hub Cluster
func ConvertMksClusterToHub(ctx context.Context, tf *fw.MksClusterModel, hub *infrapb.Cluster) (diags diag.Diagnostics) {
	var d diag.Diagnostics

	hub.Kind = tf.Kind.ValueString()
	hub.ApiVersion = tf.ApiVersion.ValueString()

	hub.Metadata = &commonpb.Metadata{}
	d = convertMksMetadataToHub(&tf.Metadata, hub.Metadata)

	diags = append(diags, d...)

	var blueprintType fw.BlueprintType

	tfBlueprintValue, d := blueprintType.ValueFromObject(ctx, tf.Spec.Blueprint)
	if d.HasError() {
		diags = append(diags, d...)
		return diags
	}

	tfBlueprint := tfBlueprintValue.(fw.BlueprintValue)
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
		hub.Spec.Sharing = &infrapb.Sharing{}
		d = convertMksSharingToHub(ctx, tf.Spec.Sharing, hub.Spec.Sharing)
		diags = append(diags, d...)
	}

	if !tf.Spec.SystemComponentsPlacement.IsNull() && !tf.Spec.SystemComponentsPlacement.IsUnknown() {
		hub.Spec.SystemComponentsPlacement = &infrapb.SystemComponentsPlacement{}
		d = convertMksSystemComponentsPlacementToHub(ctx, tf.Spec.SystemComponentsPlacement, hub.Spec.SystemComponentsPlacement)
		diags = append(diags, d...)
	}

	if !tf.Spec.Proxy.IsNull() && !tf.Spec.Proxy.IsUnknown() {
		hub.Spec.Proxy = &infrapb.ClusterProxy{}
		d = convertMksProxyToHub(ctx, tf.Spec.Proxy, hub.Spec.Proxy)
		diags = append(diags, d...)
	}

	mksConfig, d := convertMksConfigToHub(ctx, tf.Spec.Config)
	if d.HasError() {
		diags = append(diags, d...)
		return diags
	}

	hub.Spec.Config = &infrapb.ClusterSpec_Mks{
		Mks: mksConfig,
	}

	return diags

}

// ConvertMksClusterFromHub converts a hub Cluster to a Terraform MksClusterModel
func ConvertMksClusterFromHub(ctx context.Context, hub *infrapb.Cluster, tf *fw.MksClusterModel) diag.Diagnostics {
	var diags diag.Diagnostics

	tf.Kind = types.StringValue(hub.Kind)
	tf.ApiVersion = types.StringValue(hub.ApiVersion)

	d := convertMetadataFromHub(hub.Metadata, &tf.Metadata)
	diags.Append(d...)

	tf.Spec.SpecType = types.StringValue(hub.Spec.Type)

	tfBlueprint, d := fw.NewBlueprintValue(tf.Spec.Blueprint.AttributeTypes(ctx), tf.Spec.Blueprint.Attributes())
	if d.HasError() {
		return d
	}

	tfBlueprint.Name = types.StringValue(hub.Spec.Blueprint.Name)
	if hub.Spec.Blueprint.Version != "" {
		tfBlueprint.Version = types.StringValue(hub.Spec.Blueprint.Version)
	}

	tf.Spec.Blueprint, d = tfBlueprint.ToObjectValue(ctx)

	diags.Append(d...)

	if hub.Spec.CloudCredentials != "" {
		tf.Spec.CloudCredentials = types.StringValue(hub.Spec.CloudCredentials)
	}

	if hub.Spec.Sharing != nil {
		tf.Spec.Sharing, d = convertMksSharingFromHub(ctx, hub.Spec.Sharing, &tf.Spec.Sharing)
		diags.Append(d...)
	}

	if hub.Spec.SystemComponentsPlacement != nil {
		tf.Spec.SystemComponentsPlacement, d = convertMksSystemComponentsPlacementFromHub(ctx, hub.Spec.SystemComponentsPlacement, &tf.Spec.SystemComponentsPlacement)
		diags.Append(d...)
	} else {
		tf.Spec.SystemComponentsPlacement, d = fw.NewSystemComponentsPlacementValueNull().ToObjectValue(ctx)
		diags.Append(d...)
	}

	if hub.Spec.Proxy != nil {
		tf.Spec.Proxy, d = convertMksProxyFromHub(ctx, hub.Spec.Proxy, &tf.Spec.Proxy)
		diags.Append(d...)
	}

	if hub.Spec.Config != nil {
		hubConfig := hub.Spec.Config.(*infrapb.ClusterSpec_Mks).Mks

		tf.Spec.Config, d = convertMksConfigFromHub(ctx, hubConfig, &tf.Spec.Config)
		diags.Append(d...)

	}

	return diags
}
