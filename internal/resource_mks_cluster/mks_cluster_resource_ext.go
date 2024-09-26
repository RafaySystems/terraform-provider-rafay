// Contains the conversion methods to convert the Terraform types to the Hub types and vice versa
// For each Tf type, we're extending it with ToHub and FromHub signature
// The state of each type is set to known after the conversion

package resource_mks_cluster

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/RafaySystems/rafay-common/proto/types/hub/commonpb"
	"github.com/RafaySystems/rafay-common/proto/types/hub/infrapb"
	v1 "k8s.io/api/core/v1"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	dynamic "github.com/RafaySystems/rafay-common/pkg/hub/client/dynamic"
	typed "github.com/RafaySystems/rafay-common/pkg/hub/client/typed"

	"github.com/RafaySystems/rafay-common/pkg/hub/client/options"
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

func (v MetadataValue) ToHub(ctx context.Context) (*commonpb.Metadata, diag.Diagnostics) {
	var diags diag.Diagnostics

	var hub commonpb.Metadata

	hub.Name = getStringValue(v.Name)
	hub.Project = getStringValue(v.Project)

	if !v.Annotations.IsNull() && !v.Annotations.IsUnknown() {
		hub.Annotations = convertFromTfMap(v.Annotations)
	}

	if !v.Description.IsNull() && !v.Description.IsUnknown() {
		hub.Description = getStringValue(v.Description)
	}

	if !v.Labels.IsNull() && !v.Labels.IsUnknown() {
		hub.Labels = convertFromTfMap(v.Labels)
	}

	return &hub, diags
}

func (v MetadataValue) FromHub(ctx context.Context, hub *commonpb.Metadata) (MetadataValue, diag.Diagnostics) {

	var diags diag.Diagnostics

	v.Name = types.StringValue(hub.Name)
	v.Project = types.StringValue(hub.Project)

	if hub.Description != "" {
		v.Description = types.StringValue(hub.Description)
	}

	if hub.Annotations != nil {
		v.Annotations = convertToTfMap(hub.Annotations)
	}
	if hub.Labels != nil {
		v.Labels = convertToTfMap(hub.Labels)
	}

	v.state = attr.ValueStateKnown
	return v, diags
}

func (v CniValue) ToHub(ctx context.Context) (*infrapb.Cni, diag.Diagnostics) {
	hub := &infrapb.Cni{}

	hub.Name = getStringValue(v.Name)
	hub.Version = getStringValue(v.Version)
	return hub, nil
}

func (v CniValue) FromHub(ctx context.Context, hub *infrapb.Cni) (basetypes.ObjectValue, diag.Diagnostics) {
	// Convert the hub object to terraform object
	v.Name = types.StringValue(hub.Name)
	v.Version = types.StringValue(hub.Version)

	v.state = attr.ValueStateKnown
	return v.ToObjectValue(ctx)
}

func (v Ipv6Value) ToHub(ctx context.Context) (*infrapb.MksSubnet, diag.Diagnostics) {
	var hub infrapb.MksSubnet

	hub.PodSubnet = getStringValue(v.PodSubnet)
	hub.ServiceSubnet = getStringValue(v.ServiceSubnet)

	return &hub, nil
}

func (v Ipv6Value) FromHub(ctx context.Context, hub *infrapb.MksSubnet) (basetypes.ObjectValue, diag.Diagnostics) {
	// Convert the hub object to terraform object
	if hub.PodSubnet != "" {
		v.PodSubnet = types.StringValue(hub.PodSubnet)
	}
	if hub.ServiceSubnet != "" {
		v.ServiceSubnet = types.StringValue(hub.ServiceSubnet)
	}

	v.state = attr.ValueStateKnown
	return v.ToObjectValue(ctx)
}

func (v NetworkValue) ToHub(ctx context.Context) (*infrapb.MksClusterNetworking, diag.Diagnostics) {
	var cniType CniType

	var diags, d diag.Diagnostics

	// Get the value from the object
	tfCni, d := cniType.ValueFromObject(ctx, v.Cni)
	if diags.HasError() {
		diags = append(diags, d...)
		return nil, diags
	}

	hub := &infrapb.MksClusterNetworking{}
	hub.Cni, d = tfCni.(CniValue).ToHub(ctx)
	diags = append(diags, d...)

	hub.PodSubnet = getStringValue(v.PodSubnet)
	hub.ServiceSubnet = getStringValue(v.ServiceSubnet)

	if !v.Ipv6.IsNull() && !v.Ipv6.IsUnknown() {
		// Handle IPv6
		var ipv6Type Ipv6Type
		tfIpv6Value, d := ipv6Type.ValueFromObject(ctx, v.Ipv6)
		if diags.HasError() {
			diags = append(diags, d...)
			return hub, diags
		}
		hub.Ipv6, d = tfIpv6Value.(Ipv6Value).ToHub(ctx)
		diags = append(diags, d...)

	}

	return hub, diags
}

func (v NetworkValue) FromHub(ctx context.Context, hub *infrapb.MksClusterNetworking) (basetypes.ObjectValue, diag.Diagnostics) {
	var diags, d diag.Diagnostics

	tfCni, d := NewCniValue(v.Cni.AttributeTypes(ctx), v.Cni.Attributes())
	if d.HasError() {
		tfCni = NewCniValueNull()
	}
	v.Cni, d = tfCni.FromHub(ctx, hub.Cni)
	diags = append(diags, d...)

	v.PodSubnet = types.StringValue(hub.PodSubnet)
	v.ServiceSubnet = types.StringValue(hub.ServiceSubnet)

	// Handle IPv6
	if hub.Ipv6 != nil {
		tfIpv6, d := NewIpv6Value(v.Ipv6.AttributeTypes(ctx), v.Ipv6.Attributes())
		if d.HasError() {
			tfIpv6 = NewIpv6ValueNull()
		}
		v.Ipv6, diags = tfIpv6.FromHub(ctx, hub.Ipv6)
		diags = append(diags, d...)
	} else {
		v.Ipv6, d = NewIpv6ValueNull().ToObjectValue(ctx)
		diags = append(diags, d...)
	}

	v.state = attr.ValueStateKnown

	obj, d := v.ToObjectValue(ctx)
	diags = append(diags, d...)

	return obj, diags

}

func (v TaintsValue) ToHub(ctx context.Context) (*v1.Taint, diag.Diagnostics) {
	var hub v1.Taint

	hub.Effect = v1.TaintEffect(getStringValue(v.Effect))
	hub.Key = getStringValue(v.Key)
	hub.Value = getStringValue(v.Value)
	return &hub, nil
}

func (v TaintsValue) FromHub(ctx context.Context, hub *v1.Taint) (TaintsValue, diag.Diagnostics) {
	// Convert the hub object to terraform object
	if hub.Effect != "" {
		v.Effect = types.StringValue(string(hub.Effect))
	}
	if hub.Key != "" {
		v.Key = types.StringValue(hub.Key)
	}
	if hub.Value != "" {
		v.Value = types.StringValue(hub.Value)
	}

	v.state = attr.ValueStateKnown
	return v, nil
}

func (v TolerationsValue) ToHub(ctx context.Context) (*v1.Toleration, diag.Diagnostics) {
	var hub v1.Toleration

	hub.Effect = v1.TaintEffect(getStringValue(v.Effect))
	hub.Key = getStringValue(v.Key)
	hub.Value = getStringValue(v.Value)
	hub.Operator = v1.TolerationOperator(getStringValue(v.Operator))
	hub.TolerationSeconds = v.TolerationSeconds.ValueInt64Pointer()

	return &hub, nil
}

func (v TolerationsValue) FromHub(ctx context.Context, hub *v1.Toleration) (TolerationsValue, diag.Diagnostics) {
	// Convert the hub object to terraform object
	if hub.Effect != "" {
		v.Effect = types.StringValue(string(hub.Effect))
	}
	if hub.Key != "" {
		v.Key = types.StringValue(hub.Key)
	}
	if hub.Value != "" {
		v.Value = types.StringValue(hub.Value)
	}
	if hub.Operator != "" {
		v.Operator = types.StringValue(string(hub.Operator))
	}
	if hub.TolerationSeconds != nil {
		v.TolerationSeconds = types.Int64PointerValue(hub.TolerationSeconds)
	}

	v.state = attr.ValueStateKnown
	return v, nil
}

func (v DaemonSetTolerationsValue) ToHub(ctx context.Context) (*v1.Toleration, diag.Diagnostics) {
	var hub v1.Toleration

	hub.Effect = v1.TaintEffect(getStringValue(v.Effect))
	hub.Key = getStringValue(v.Key)
	hub.Value = getStringValue(v.Value)
	hub.Operator = v1.TolerationOperator(getStringValue(v.Operator))
	hub.TolerationSeconds = v.TolerationSeconds.ValueInt64Pointer()

	return &hub, nil
}

func (v DaemonSetTolerationsValue) FromHub(ctx context.Context, hub *v1.Toleration) (DaemonSetTolerationsValue, diag.Diagnostics) {
	// Convert the hub object to terraform object
	if hub.Effect != "" {
		v.Effect = types.StringValue(string(hub.Effect))
	}
	if hub.Key != "" {
		v.Key = types.StringValue(hub.Key)
	}
	if hub.Value != "" {
		v.Value = types.StringValue(hub.Value)
	}
	if hub.Operator != "" {
		v.Operator = types.StringValue(string(hub.Operator))
	}
	if hub.TolerationSeconds != nil {
		v.TolerationSeconds = types.Int64PointerValue(hub.TolerationSeconds)
	}

	v.state = attr.ValueStateKnown
	return v, nil
}

func (v DaemonSetOverrideValue) ToHub(ctx context.Context) (*infrapb.DaemonSetOverride, diag.Diagnostics) {
	var diags diag.Diagnostics
	hub := &infrapb.DaemonSetOverride{}

	hub.NodeSelectionEnabled = getBoolValue(v.NodeSelectionEnabled)

	for _, toleration := range v.DaemonSetTolerations.Elements() {
		h, d := toleration.(DaemonSetTolerationsValue).ToHub(ctx)
		diags = append(diags, d...)
		hub.Tolerations = append(hub.Tolerations, h)
	}

	return hub, diags
}

func (v DaemonSetOverrideValue) FromHub(ctx context.Context, hub *infrapb.DaemonSetOverride) (basetypes.ObjectValue, diag.Diagnostics) {
	// Convert the hub object to terraform object
	var diags, d diag.Diagnostics

	if hub.NodeSelectionEnabled {
		v.NodeSelectionEnabled = types.BoolValue(hub.NodeSelectionEnabled)
	}

	var tfDaemonSetTolerations []attr.Value

	tfDaemonSetTolerationsType := DaemonSetTolerationsType{
		ObjectType: types.ObjectType{
			AttrTypes: DaemonSetTolerationsValue{}.AttributeTypes(ctx),
		},
	}

	if hub.Tolerations != nil {
		// loop through the hub tolerations and convert them to terraform tolerations
		for _, hub := range hub.Tolerations {
			tfDsTol := &DaemonSetTolerationsValue{}
			h, d := tfDsTol.FromHub(ctx, hub)
			diags = append(diags, d...)
			tfDaemonSetTolerations = append(tfDaemonSetTolerations, h)
		}

		v.DaemonSetTolerations, d = types.SetValue(tfDaemonSetTolerationsType, tfDaemonSetTolerations)
		diags = append(diags, d...)
	} else {
		v.DaemonSetTolerations = types.SetNull(tfDaemonSetTolerationsType)
	}

	v.state = attr.ValueStateKnown
	obj, d := v.ToObjectValue(ctx)
	diags = append(diags, d...)

	return obj, diags
}

func (v ProxyValue) ToHub(ctx context.Context) (*infrapb.ClusterProxy, diag.Diagnostics) {

	hub := &infrapb.ClusterProxy{}
	hub.AllowInsecureBootstrap = getBoolValue(v.AllowInsecureBootstrap)
	hub.BootstrapCA = getStringValue(v.BootstrapCa)
	hub.Enabled = getBoolValue(v.Enabled)
	hub.HttpProxy = getStringValue(v.HttpProxy)
	hub.HttpsProxy = getStringValue(v.HttpsProxy)
	hub.NoProxy = getStringValue(v.NoProxy)
	hub.ProxyAuth = getStringValue(v.ProxyAuth)

	return hub, nil
}

func (v ProxyValue) FromHub(ctx context.Context, hub *infrapb.ClusterProxy) (basetypes.ObjectValue, diag.Diagnostics) {
	// Convert the hub object to terraform object
	if hub.AllowInsecureBootstrap {

		v.AllowInsecureBootstrap = types.BoolValue(hub.AllowInsecureBootstrap)
	}
	if hub.BootstrapCA != "" {
		v.BootstrapCa = types.StringValue(hub.BootstrapCA)
	}
	if hub.Enabled {
		v.Enabled = types.BoolValue(hub.Enabled)
	}
	if hub.HttpProxy != "" {
		v.HttpProxy = types.StringValue(hub.HttpProxy)
	}
	if hub.HttpsProxy != "" {
		v.HttpsProxy = types.StringValue(hub.HttpsProxy)
	}
	if hub.NoProxy != "" {
		v.NoProxy = types.StringValue(hub.NoProxy)
	}
	if hub.ProxyAuth != "" {
		v.ProxyAuth = types.StringValue(hub.ProxyAuth)
	}

	v.state = attr.ValueStateKnown
	return v.ToObjectValue(ctx)
}

func (v SharingValue) ToHub(ctx context.Context) (*infrapb.Sharing, diag.Diagnostics) {

	hub := &infrapb.Sharing{}
	hub.Enabled = getBoolValue(v.Enabled)

	for _, project := range v.Projects.Elements() {
		hub.Projects = append(hub.Projects, &infrapb.Projects{
			Name: getStringValue(project.(types.String)),
		})
	}

	return hub, nil
}

func (v SharingValue) FromHub(ctx context.Context, hub *infrapb.Sharing) (basetypes.ObjectValue, diag.Diagnostics) {
	// Convert the hub object to terraform object

	v.Enabled = types.BoolValue(hub.Enabled)

	var tfProjects []attr.Value
	for _, project := range hub.Projects {
		tfProjects = append(tfProjects, types.StringValue(project.Name))

	}
	v.Projects, _ = types.SetValue(types.StringType, tfProjects)

	v.state = attr.ValueStateKnown
	return v.ToObjectValue(ctx)
}

func (v SystemComponentsPlacementValue) ToHub(ctx context.Context) (*infrapb.SystemComponentsPlacement, diag.Diagnostics) {
	var diags diag.Diagnostics
	hub := &infrapb.SystemComponentsPlacement{}

	hub.NodeSelector = convertFromTfMap(v.NodeSelector)

	for _, toleration := range v.Tolerations.Elements() {
		h, d := toleration.(TolerationsValue).ToHub(ctx)
		diags = append(diags, d...)
		hub.Tolerations = append(hub.Tolerations, h)
	}

	if !v.DaemonSetOverride.IsNull() && !v.DaemonSetOverride.IsUnknown() {
		var daemonSetType DaemonSetOverrideType
		tfDaemonSetValue, d := daemonSetType.ValueFromObject(ctx, v.DaemonSetOverride)
		if d.HasError() {
			diags = append(diags, d...)
			return hub, diags
		}
		hub.DaemonSetOverride, diags = tfDaemonSetValue.(DaemonSetOverrideValue).ToHub(ctx)
	}

	return hub, diags
}

func (v SystemComponentsPlacementValue) FromHub(ctx context.Context, hub *infrapb.SystemComponentsPlacement) (basetypes.ObjectValue, diag.Diagnostics) {

	var diags, d diag.Diagnostics
	if hub.NodeSelector != nil {
		v.NodeSelector = convertToTfMap(hub.NodeSelector)
	}
	var tfTolerations []attr.Value

	tfTolerationsType := TolerationsType{
		ObjectType: types.ObjectType{
			AttrTypes: TolerationsValue{}.AttributeTypes(ctx),
		},
	}

	if hub.Tolerations != nil {
		// loop through the hub tolerations and convert them to terraform tolerations
		for _, hub := range hub.Tolerations {
			tfTol := &TolerationsValue{}
			h, d := tfTol.FromHub(ctx, hub)
			diags = append(diags, d...)
			tfTolerations = append(tfTolerations, h)
		}

		v.Tolerations, d = types.SetValue(tfTolerationsType, tfTolerations)
		diags = append(diags, d...)
	} else {
		v.Tolerations = types.SetNull(tfTolerationsType)
	}

	if hub.DaemonSetOverride != nil {
		tfDaemonSetOverride, d := NewDaemonSetOverrideValue(v.DaemonSetOverride.AttributeTypes(ctx), v.DaemonSetOverride.Attributes())
		if d.HasError() {
			tfDaemonSetOverride = NewDaemonSetOverrideValueNull()
		}
		v.DaemonSetOverride, d = tfDaemonSetOverride.FromHub(ctx, hub.DaemonSetOverride)
		diags = append(diags, d...)
	} else {
		v.DaemonSetOverride, d = NewDaemonSetOverrideValueNull().ToObjectValue(ctx)
		diags = append(diags, d...)
	}

	v.state = attr.ValueStateKnown

	obj, d := v.ToObjectValue(ctx)
	diags = append(diags, d...)

	return obj, diags
}

func (v ClusterSshValue) ToHub(ctx context.Context) (*infrapb.MksClusterSshConfig, diag.Diagnostics) {
	hub := &infrapb.MksClusterSshConfig{}

	hub.PrivateKeyPath = getStringValue(v.PrivateKeyPath)
	hub.Username = getStringValue(v.Username)
	hub.Port = getStringValue(v.Port)
	hub.Passphrase = getStringValue(v.Passphrase)

	return hub, nil
}

func (v ClusterSshValue) FromHub(ctx context.Context, hub *infrapb.MksClusterSshConfig) (basetypes.ObjectValue, diag.Diagnostics) {

	if hub.PrivateKeyPath != "" {
		v.PrivateKeyPath = types.StringValue(hub.PrivateKeyPath)
	}
	if hub.Username != "" {
		v.Username = types.StringValue(hub.Username)
	}
	if hub.Port != "" {
		v.Port = types.StringValue(hub.Port)
	}
	if hub.Passphrase != "" {
		v.Passphrase = types.StringValue(hub.Passphrase)
	}

	v.state = attr.ValueStateKnown
	return v.ToObjectValue(ctx)
}

func (v SshValue) ToHub(ctx context.Context) (*infrapb.MksNodeSshConfig, diag.Diagnostics) {
	hub := &infrapb.MksNodeSshConfig{}

	hub.IpAddress = getStringValue(v.IpAddress)
	hub.Passphrase = getStringValue(v.Passphrase)
	hub.Port = getStringValue(v.Port)
	hub.PrivateKeyPath = getStringValue(v.PrivateKeyPath)
	hub.Username = getStringValue(v.Username)

	return hub, nil
}

func (v SshValue) FromHub(ctx context.Context, hub *infrapb.MksNodeSshConfig) (basetypes.ObjectValue, diag.Diagnostics) {
	if hub.IpAddress != "" {
		v.IpAddress = types.StringValue(hub.IpAddress)
	}
	if hub.Passphrase != "" {
		v.Passphrase = types.StringValue(hub.Passphrase)
	}
	if hub.Port != "" {
		v.Port = types.StringValue(hub.Port)
	}
	if hub.PrivateKeyPath != "" {
		v.PrivateKeyPath = types.StringValue(hub.PrivateKeyPath)
	}
	if hub.Username != "" {
		v.Username = types.StringValue(hub.Username)
	}

	v.state = attr.ValueStateKnown
	return v.ToObjectValue(ctx)
}

func (v NodesValue) ToHub(ctx context.Context) (*infrapb.MksNode, diag.Diagnostics) {
	var diags diag.Diagnostics

	hub := &infrapb.MksNode{}

	hub.Arch = getStringValue(v.Arch)
	hub.Hostname = getStringValue(v.Hostname)
	hub.OperatingSystem = getStringValue(v.OperatingSystem)
	hub.PrivateIP = getStringValue(v.PrivateIp)

	hub.Interface = v.Interface.ValueString()

	for _, role := range v.Roles.Elements() {
		hub.Roles = append(hub.Roles, getStringValue(role.(types.String)))
	}

	hub.Labels = convertFromTfMap(v.Labels)

	for _, taint := range v.Taints.Elements() {
		h, d := taint.(TaintsValue).ToHub(ctx)
		diags = append(diags, d...)
		hub.Taints = append(hub.Taints, h)
	}

	if !v.Ssh.IsNull() && !v.Ssh.IsUnknown() {
		var sshType SshType
		tfSshValue, d := sshType.ValueFromObject(ctx, v.Ssh)
		if d.HasError() {
			diags = append(diags, d...)
			return hub, diags
		}
		hub.Ssh, d = tfSshValue.(SshValue).ToHub(ctx)
		diags = append(diags, d...)
	}

	return hub, diags
}

func (v NodesValue) FromHub(ctx context.Context, hub *infrapb.MksNode) (NodesValue, diag.Diagnostics) {

	var diags, d diag.Diagnostics

	v.Arch = types.StringValue(hub.Arch)
	v.Hostname = types.StringValue(hub.Hostname)
	v.OperatingSystem = types.StringValue(hub.OperatingSystem)
	v.PrivateIp = types.StringValue(hub.PrivateIP)

	if hub.Interface != "" {
		v.Interface = types.StringValue(hub.Interface)
	}

	var tfRoles []attr.Value
	for _, role := range hub.Roles {
		tfRoles = append(tfRoles, types.StringValue(role))
	}

	v.Roles, d = types.SetValue(types.StringType, tfRoles)
	diags = append(diags, d...)

	v.Labels = convertToTfMap(hub.Labels)

	var tfTaints []attr.Value

	tfTaintsType := TaintsType{
		ObjectType: types.ObjectType{
			AttrTypes: TaintsValue{}.AttributeTypes(ctx),
		},
	}

	if hub.Taints != nil {
		// loop through the hub taints and convert them to terraform taints
		for _, hub := range hub.Taints {
			tfTaint := &TaintsValue{}
			h, d := tfTaint.FromHub(ctx, hub)
			diags = append(diags, d...)
			tfTaints = append(tfTaints, h)
		}

		v.Taints, d = types.SetValue(tfTaintsType, tfTaints)
		diags = append(diags, d...)
	} else {
		v.Taints = types.SetNull(tfTaintsType)
	}

	if hub.Ssh != nil {
		tfSsh, d := NewSshValue(v.Ssh.AttributeTypes(ctx), v.Ssh.Attributes())
		if d.HasError() {
			tfSsh = NewSshValueNull()
		}
		v.Ssh, d = tfSsh.FromHub(ctx, hub.Ssh)
		diags = append(diags, d...)
	} else {
		v.Ssh, d = NewSshValueNull().ToObjectValue(ctx)
		diags = append(diags, d...)
	}

	v.state = attr.ValueStateKnown

	return v, diags
}

func (v ConfigValue) ToHub(ctx context.Context) (*infrapb.MksV3ConfigObject, diag.Diagnostics) {
	var diags, d diag.Diagnostics

	hub := &infrapb.MksV3ConfigObject{}

	hub.Location = getStringValue(v.Location)
	hub.AutoApproveNodes = getBoolValue(v.AutoApproveNodes)
	hub.DedicatedControlPlane = getBoolValue(v.DedicatedControlPlane)
	hub.HighAvailability = getBoolValue(v.HighAvailability)
	hub.KubernetesVersion = getStringValue(v.KubernetesVersion)

	var networkType NetworkType

	tfNetworkValue, d := networkType.ValueFromObject(ctx, v.Network)
	if d.HasError() {
		diags = append(diags, d...)
		return hub, diags
	}

	hub.Network, d = tfNetworkValue.(NetworkValue).ToHub(ctx)
	diags = append(diags, d...)

	for tfHostName, node := range v.Nodes.Elements() {
		h, d := node.(NodesValue).ToHub(ctx)
		diags = append(diags, d...)
		if tfHostName != h.Hostname {
			diags.AddAttributeError(path.Root(fmt.Sprintf("spec.config.nodes.%s", tfHostName)),
				"Mismatch in the node configuration",
				"We strongly enforce using same hostname as key in node configuration")
		}
		hub.Nodes = append(hub.Nodes, h)
	}

	if !v.ClusterSsh.IsNull() && !v.ClusterSsh.IsUnknown() {
		var sshType ClusterSshType
		tfSshValue, d := sshType.ValueFromObject(ctx, v.ClusterSsh)
		if d.HasError() {
			diags = append(diags, d...)
			return hub, diags
		}
		hub.Ssh, d = tfSshValue.(ClusterSshValue).ToHub(ctx)
		diags = append(diags, d...)
	}

	return hub, diags

}

func (v ConfigValue) FromHub(ctx context.Context, hub *infrapb.MksV3ConfigObject) (basetypes.ObjectValue, diag.Diagnostics) {
	var diags, d diag.Diagnostics

	if hub.Location != "" {
		v.Location = types.StringValue(hub.Location)
	}
	if hub.AutoApproveNodes {
		v.AutoApproveNodes = types.BoolValue(hub.AutoApproveNodes)
	}
	if hub.DedicatedControlPlane {
		v.DedicatedControlPlane = types.BoolValue(hub.DedicatedControlPlane)
	}
	if hub.HighAvailability {
		v.HighAvailability = types.BoolValue(hub.HighAvailability)
	}

	v.KubernetesVersion = types.StringValue(hub.KubernetesVersion)

	network, d := NewNetworkValue(v.Network.AttributeTypes(ctx), v.Network.Attributes())
	if d.HasError() {
		network = NewNetworkValueNull()
	}
	v.Network, d = network.FromHub(ctx, hub.Network)
	diags = append(diags, d...)

	tfNodeMap := make(map[string]NodesValue)
	for _, node := range v.Nodes.Elements() {
		tfNode := node.(NodesValue)
		tfNodeMap[getStringValue(tfNode.Hostname)] = tfNode
	}

	newTfNodes := make(map[string]attr.Value)
	// Compare the nodes in the hub and terraform
	for _, hubNode := range hub.Nodes {
		tfNode, ok := tfNodeMap[hubNode.Hostname]
		if !ok {
			tfNode = NewNodesValueNull()
		}
		h, d := tfNode.FromHub(ctx, hubNode)
		diags = append(diags, d...)
		newTfNodes[hubNode.Hostname] = h
	}

	tfNodeType := NodesType{
		ObjectType: types.ObjectType{
			AttrTypes: NodesValue{}.AttributeTypes(ctx),
		},
	}

	v.Nodes, d = types.MapValue(tfNodeType, newTfNodes)
	diags = append(diags, d...)

	if hub.Ssh != nil {
		tfSsh, d := NewClusterSshValue(v.ClusterSsh.AttributeTypes(ctx), v.ClusterSsh.Attributes())
		if d.HasError() {
			tfSsh = NewClusterSshValueNull()
		}
		v.ClusterSsh, d = tfSsh.FromHub(ctx, hub.Ssh)
		diags = append(diags, d...)
	} else {
		v.ClusterSsh, d = NewClusterSshValueNull().ToObjectValue(ctx)
		diags = append(diags, d...)
	}

	v.state = attr.ValueStateKnown

	obj, d := v.ToObjectValue(ctx)
	diags = append(diags, d...)

	return obj, diags
}

func (v BlueprintValue) ToHub(ctx context.Context) (*infrapb.ClusterBlueprint, diag.Diagnostics) {
	hub := &infrapb.ClusterBlueprint{}

	hub.Name = getStringValue(v.Name)
	hub.Version = getStringValue(v.Version)

	return hub, nil
}

func (v BlueprintValue) FromHub(ctx context.Context, hub *infrapb.ClusterBlueprint) (basetypes.ObjectValue, diag.Diagnostics) {
	v.Name = types.StringValue(hub.Name)
	v.Version = types.StringValue(hub.Version)

	v.state = attr.ValueStateKnown
	return v.ToObjectValue(ctx)
}

func ConvertMksClusterToHub(ctx context.Context, v MksClusterModel) (*infrapb.Cluster, diag.Diagnostics) {
	var diags, d diag.Diagnostics

	hub := &infrapb.Cluster{}

	hub.Kind = getStringValue(v.Kind)
	hub.ApiVersion = getStringValue(v.ApiVersion)

	hub.Metadata, d = v.Metadata.ToHub(ctx)
	diags = append(diags, d...)

	hub.Spec = &infrapb.ClusterSpec{}
	hub.Spec.Type = getStringValue(v.Spec.SpecType)

	var blueprintType BlueprintType

	tfBlueprintValue, d := blueprintType.ValueFromObject(ctx, v.Spec.Blueprint)
	if d.HasError() {
		diags = append(diags, d...)
		return hub, diags
	}

	hub.Spec.Blueprint, d = tfBlueprintValue.(BlueprintValue).ToHub(ctx)
	diags = append(diags, d...)

	if !v.Spec.CloudCredentials.IsNull() && !v.Spec.CloudCredentials.IsUnknown() {
		hub.Spec.CloudCredentials = getStringValue(v.Spec.CloudCredentials)
	}

	if !v.Spec.Sharing.IsNull() && !v.Spec.Sharing.IsUnknown() {
		var sharingType SharingType
		tfSharingValue, d := sharingType.ValueFromObject(ctx, v.Spec.Sharing)
		if d.HasError() {
			diags = append(diags, d...)
			return hub, diags
		}
		hub.Spec.Sharing, d = tfSharingValue.(SharingValue).ToHub(ctx)
		diags = append(diags, d...)
	}

	if !v.Spec.SystemComponentsPlacement.IsNull() && !v.Spec.SystemComponentsPlacement.IsUnknown() {
		var scpType SystemComponentsPlacementType
		scp, d := scpType.ValueFromObject(ctx, v.Spec.SystemComponentsPlacement)
		if d.HasError() {
			diags = append(diags, d...)
			return hub, diags
		}
		hub.Spec.SystemComponentsPlacement, d = scp.(SystemComponentsPlacementValue).ToHub(ctx)
		diags = append(diags, d...)
	}

	if !v.Spec.Proxy.IsNull() && !v.Spec.Proxy.IsUnknown() {
		var proxyType ProxyType
		tfProxyValue, d := proxyType.ValueFromObject(ctx, v.Spec.Proxy)
		if d.HasError() {
			diags = append(diags, d...)
			return hub, diags
		}
		hub.Spec.Proxy, d = tfProxyValue.(ProxyValue).ToHub(ctx)
		diags = append(diags, d...)
	}

	var configType ConfigType

	tfConfigValue, d := configType.ValueFromObject(ctx, v.Spec.Config)
	if d.HasError() {
		diags = append(diags, d...)
		return hub, diags
	}

	mksConfig, d := tfConfigValue.(ConfigValue).ToHub(ctx)
	diags = append(diags, d...)

	hub.Spec.Config = &infrapb.ClusterSpec_Mks{
		Mks: mksConfig,
	}

	return hub, diags
}

func ConvertMksClusterFromHub(ctx context.Context, hub *infrapb.Cluster, tf *MksClusterModel) diag.Diagnostics {
	// Convert the hub object to terraform object
	var diags, d diag.Diagnostics

	tf.Kind = types.StringValue(hub.Kind)
	tf.ApiVersion = types.StringValue(hub.ApiVersion)

	tf.Metadata, d = tf.Metadata.FromHub(ctx, hub.Metadata)
	diags = append(diags, d...)

	tf.Spec.SpecType = types.StringValue(hub.Spec.Type)

	if hub.Spec.Blueprint != nil {
		tfBlueprint, d := NewBlueprintValue(tf.Spec.Blueprint.AttributeTypes(ctx), tf.Spec.Blueprint.Attributes())
		if d.HasError() {
			tfBlueprint = NewBlueprintValueNull()
		}
		tf.Spec.Blueprint, d = tfBlueprint.FromHub(ctx, hub.Spec.Blueprint)
		diags = append(diags, d...)
	}

	if hub.Spec.CloudCredentials != "" {
		tf.Spec.CloudCredentials = types.StringValue(hub.Spec.CloudCredentials)
	}

	if hub.Spec.Sharing != nil {
		tfSharing, d := NewSharingValue(tf.Spec.Sharing.AttributeTypes(ctx), tf.Spec.Sharing.Attributes())
		if d.HasError() {
			tfSharing = NewSharingValueNull()
		}
		tf.Spec.Sharing, d = tfSharing.FromHub(ctx, hub.Spec.Sharing)
		diags = append(diags, d...)
	} else {
		tf.Spec.Sharing, d = NewSharingValueNull().ToObjectValue(ctx)
		diags = append(diags, d...)
	}

	if hub.Spec.SystemComponentsPlacement != nil {
		tfSystemComponentsPlacement, d := NewSystemComponentsPlacementValue(
			tf.Spec.SystemComponentsPlacement.AttributeTypes(ctx),
			tf.Spec.SystemComponentsPlacement.Attributes(),
		)
		if d.HasError() {
			tfSystemComponentsPlacement = NewSystemComponentsPlacementValueNull()
		}
		tf.Spec.SystemComponentsPlacement, d = tfSystemComponentsPlacement.FromHub(ctx, hub.Spec.SystemComponentsPlacement)

		diags = append(diags, d...)
	} else {
		tf.Spec.SystemComponentsPlacement, d = NewSystemComponentsPlacementValueNull().ToObjectValue(ctx)
		diags = append(diags, d...)
	}

	if hub.Spec.Proxy != nil {
		tfProxy, d := NewProxyValue(tf.Spec.Proxy.AttributeTypes(ctx), tf.Spec.Proxy.Attributes())
		if d.HasError() {
			tfProxy = NewProxyValueNull()
		}
		tf.Spec.Proxy, d = tfProxy.FromHub(ctx, hub.Spec.Proxy)
		diags = append(diags, d...)
	} else {
		tf.Spec.Proxy, d = NewProxyValueNull().ToObjectValue(ctx)
		diags = append(diags, d...)
	}

	if hub.Spec.Config != nil {
		hubConfig := hub.Spec.Config.(*infrapb.ClusterSpec_Mks).Mks
		tfConfig, d := NewConfigValue(tf.Spec.Config.AttributeTypes(ctx), tf.Spec.Config.Attributes())
		if d.HasError() {
			tfConfig = NewConfigValueNull()
		}
		tf.Spec.Config, d = tfConfig.FromHub(ctx, hubConfig)
		diags = append(diags, d...)
	}

	tf.Spec.state = attr.ValueStateKnown

	return diags
}

func WaitForClusterApplyOperation(ctx context.Context, client typed.Client, cluster *infrapb.Cluster, timeout <-chan time.Time, ticker *time.Ticker) diag.Diagnostics {
	var diags diag.Diagnostics
	for {
		select {
		case <-timeout:
			// Timeout reached
			diags.AddError("Timeout reached while waiting for cluster operation to complete", "")
			return diags

		case <-ticker.C:
			uCluster, err := client.InfraV3().Cluster().Status(ctx, options.StatusOptions{
				Name:    cluster.Metadata.Name,
				Project: cluster.Metadata.Project,
			})
			if err != nil {
				// Error occurred while fetching cluster status
				diags.AddError("Error occurred while fetching cluster status", err.Error())
				return diags
			}

			if uCluster == nil {
				continue
			}

			if uCluster.Status != nil && uCluster.Status.Mks != nil {
				uClusterCommonStatus := uCluster.Status.CommonStatus
				switch uClusterCommonStatus.ConditionStatus {
				case commonpb.ConditionStatus_StatusSubmitted:
					// Submitted
					continue
				case commonpb.ConditionStatus_StatusOK:
					// Completed
					return diags
				case commonpb.ConditionStatus_StatusFailed:
					failureReason := uClusterCommonStatus.Reason
					diags.AddError("Cluster operation failed", failureReason)
					return diags
				}
			}
		}
	}
}

func WaitForClusterDeleteOperation(ctx context.Context, client typed.Client, name string, project string, timeout <-chan time.Time, ticker *time.Ticker) diag.Diagnostics {
	var diags diag.Diagnostics
	for {
		select {
		case <-timeout:
			// Timeout reached
			diags.AddError("Timeout reached while deleting the cluster resource", "")
			return diags

		case <-ticker.C:
			_, err := client.InfraV3().Cluster().Get(ctx, options.GetOptions{
				Name:    name,
				Project: project,
			})
			if err, ok := err.(*dynamic.DynamicClientGetError); ok && err != nil {
				switch err.StatusCode {
				case http.StatusNotFound:
					return diags
				default:
					diags.AddError("Cluster Deletion failed", err.Error())
					return diags
				}
			}
		}
	}
}
