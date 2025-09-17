package resource_eks_cluster

import (
	"context"

	"github.com/RafaySystems/terraform-provider-rafay/rafay"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	jsoniter "github.com/json-iterator/go"
)

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

func getInt64Value(tfInt types.Int64) int64 {
	if tfInt.IsNull() || tfInt.IsUnknown() {
		return 0
	}
	return tfInt.ValueInt64()
}

func getFloat64Value(tfFloat types.Float64) float64 {
	if tfFloat.IsNull() || tfFloat.IsUnknown() {
		return 0
	}
	return tfFloat.ValueFloat64()
}

func ExpandEksCluster(ctx context.Context, v EksClusterModel) (*rafay.EKSCluster, diag.Diagnostics) {
	var diags, d diag.Diagnostics
	cluster := &rafay.EKSCluster{}

	if v.Cluster.IsNull() {
		return &rafay.EKSCluster{}, diags
	}

	vClusterList := make([]ClusterValue, 0, len(v.Cluster.Elements()))
	diags = v.Cluster.ElementsAs(ctx, &vClusterList, false)
	if len(vClusterList) > 0 {
		cluster, d = vClusterList[0].Expand(ctx)
		diags = append(diags, d...)
	}

	return cluster, diags
}

func ExpandEksClusterConfig(ctx context.Context, v EksClusterModel) (*rafay.EKSClusterConfig, diag.Diagnostics) {
	var diags, d diag.Diagnostics
	var clusterConfig *rafay.EKSClusterConfig

	if v.ClusterConfig.IsNull() {
		return &rafay.EKSClusterConfig{}, diags
	}

	vClusterConfigList := make([]ClusterConfigValue, 0, len(v.ClusterConfig.Elements()))
	diags = v.ClusterConfig.ElementsAs(ctx, &vClusterConfigList, false)
	if len(vClusterConfigList) > 0 {
		clusterConfig, d = vClusterConfigList[0].Expand(ctx)
		diags = append(diags, d...)
	}

	return clusterConfig, diags
}

// Cluster Expand
func (v ClusterValue) Expand(ctx context.Context) (*rafay.EKSCluster, diag.Diagnostics) {
	var diags, d diag.Diagnostics
	var cluster rafay.EKSCluster

	if v.IsNull() {
		return &rafay.EKSCluster{}, diags
	}

	if !v.Kind.IsNull() && !v.Kind.IsUnknown() {
		cluster.Kind = getStringValue(v.Kind)
	}

	vMetadataList := make([]MetadataValue, 0, len(v.Metadata.Elements()))
	diags = v.Metadata.ElementsAs(ctx, &vMetadataList, false)
	if len(vMetadataList) > 0 {
		md := rafay.EKSClusterMetadata{}
		cluster.Metadata = &md
		cluster.Metadata, d = vMetadataList[0].Expand(ctx)
		diags = append(diags, d...)
	}

	vSpecList := make([]SpecValue, 0, len(v.Spec.Elements()))
	diags = v.Spec.ElementsAs(ctx, &vSpecList, false)
	if len(vSpecList) > 0 {
		cluster.Spec, d = vSpecList[0].Expand(ctx)
		diags = append(diags, d...)
	}

	return &cluster, diags
}

func (v MetadataValue) Expand(ctx context.Context) (*rafay.EKSClusterMetadata, diag.Diagnostics) {
	var diags, d diag.Diagnostics
	var metadata rafay.EKSClusterMetadata

	if v.IsNull() {
		return &rafay.EKSClusterMetadata{}, diags
	}

	if !v.Name.IsNull() && !v.Name.IsUnknown() {
		metadata.Name = getStringValue(v.Name)
	}
	if !v.Project.IsNull() && !v.Project.IsUnknown() {
		metadata.Project = getStringValue(v.Project)
	}

	vLabels := make(map[string]types.String, len(v.Labels.Elements()))
	d = v.Labels.ElementsAs(ctx, &vLabels, false)
	diags = append(diags, d...)
	if len(vLabels) > 0 {
		metadata.Labels = make(map[string]string, len(vLabels))
		for k, val := range vLabels {
			metadata.Labels[k] = getStringValue(val)
		}
	}

	return &metadata, diags
}

func (v SpecValue) Expand(ctx context.Context) (*rafay.EKSSpec, diag.Diagnostics) {
	var diags, d diag.Diagnostics
	var spec rafay.EKSSpec

	if v.IsNull() {
		return &rafay.EKSSpec{}, diags
	}

	if !v.Blueprint.IsNull() && !v.Blueprint.IsUnknown() {
		spec.Blueprint = getStringValue(v.Blueprint)
	}
	if !v.BlueprintVersion.IsNull() && !v.BlueprintVersion.IsUnknown() {
		spec.BlueprintVersion = getStringValue(v.BlueprintVersion)
	}
	if !v.CloudProvider.IsNull() && !v.CloudProvider.IsUnknown() {
		spec.CloudProvider = getStringValue(v.CloudProvider)
	}
	if !v.CniProvider.IsNull() && !v.CniProvider.IsUnknown() {
		spec.CniProvider = getStringValue(v.CniProvider)
	}
	if !v.SpecType.IsNull() && !v.SpecType.IsUnknown() {
		spec.Type = getStringValue(v.SpecType)
	}
	if !v.CrossAccountRoleArn.IsNull() && !v.CrossAccountRoleArn.IsUnknown() {
		spec.CrossAccountRoleArn = getStringValue(v.CrossAccountRoleArn)
	}

	proxyConfig := make(map[string]types.String, len(v.ProxyConfig.Elements()))
	d = v.ProxyConfig.ElementsAs(ctx, &proxyConfig, false)
	diags = append(diags, d...)
	if http, ok := proxyConfig["http_proxy"]; ok && getStringValue(http) != "" {
		spec.ProxyConfig.HttpProxy = getStringValue(http)
	}
	if https, ok := proxyConfig["https_proxy"]; ok && getStringValue(https) != "" {
		spec.ProxyConfig.HttpsProxy = getStringValue(https)
	}
	if noProxy, ok := proxyConfig["no_proxy"]; ok && getStringValue(noProxy) != "" {
		spec.ProxyConfig.NoProxy = getStringValue(noProxy)
	}
	if proxyAuth, ok := proxyConfig["proxy_auth"]; ok && getStringValue(proxyAuth) != "" {
		spec.ProxyConfig.ProxyAuth = getStringValue(proxyAuth)
	}
	if bootstrapCA, ok := proxyConfig["bootstrap_ca"]; ok && getStringValue(bootstrapCA) != "" {
		spec.ProxyConfig.BootstrapCA = getStringValue(bootstrapCA)
	}
	if enabled, ok := proxyConfig["enabled"]; ok && getStringValue(enabled) != "" {
		spec.ProxyConfig.Enabled = getStringValue(enabled) == "true"
	}
	if allowInsecureBootstrap, ok := proxyConfig["allow_insecure_bootstrap"]; ok && getStringValue(allowInsecureBootstrap) != "" {
		spec.ProxyConfig.AllowInsecureBootstrap = getStringValue(allowInsecureBootstrap) == "true"
	}

	if !v.CniParams.IsNull() && !v.CniParams.IsUnknown() {
		vCniParams := make([]CniParamsValue, 0, len(v.CniParams.Elements()))
		d = v.CniParams.ElementsAs(ctx, &vCniParams, false)
		diags = append(diags, d...)
		if len(vCniParams) > 0 {
			spec.CniParams, d = vCniParams[0].Expand(ctx)
			diags = append(diags, d...)
		}
	}

	if !v.SystemComponentsPlacement.IsNull() && !v.SystemComponentsPlacement.IsUnknown() {
		vSCPList := make([]SystemComponentsPlacementValue, 0, len(v.SystemComponentsPlacement.Elements()))
		d = v.SystemComponentsPlacement.ElementsAs(ctx, &vSCPList, false)
		diags = append(diags, d...)
		if len(vSCPList) > 0 {
			spec.SystemComponentsPlacement, d = vSCPList[0].Expand(ctx)
			diags = append(diags, d...)
		}
	}

	if !v.Sharing.IsNull() && !v.Sharing.IsUnknown() {
		vSharingList := make([]SharingValue, 0, len(v.Sharing.Elements()))
		d = v.Sharing.ElementsAs(ctx, &vSharingList, false)
		diags = append(diags, d...)
		if len(vSharingList) > 0 {
			spec.Sharing, d = vSharingList[0].Expand(ctx)
			diags = append(diags, d...)
		}
	}

	return &spec, diags
}

func (v CniParamsValue) Expand(ctx context.Context) (*rafay.CustomCni, diag.Diagnostics) {
	var diags, d diag.Diagnostics
	var cniParams rafay.CustomCni

	if v.IsNull() {
		return &rafay.CustomCni{}, diags
	}

	if !v.CustomCniCidr.IsNull() && !v.CustomCniCidr.IsUnknown() {
		cniParams.CustomCniCidr = getStringValue(v.CustomCniCidr)
	}

	if !v.CustomCniCrdSpec.IsNull() && !v.CustomCniCrdSpec.IsUnknown() {
		vCustomCniCrdSpec := make([]CustomCniCrdSpecValue, 0, len(v.CustomCniCrdSpec.Elements()))
		d = v.CustomCniCrdSpec.ElementsAs(ctx, &vCustomCniCrdSpec, false)
		diags = append(diags, d...)
		if len(vCustomCniCrdSpec) > 0 {
			cniParams.CustomCniCrdSpec = make(map[string][]rafay.CustomCniSpec, len(vCustomCniCrdSpec))
			for _, crdSpec := range vCustomCniCrdSpec {
				nm, ele, d := crdSpec.Expand(ctx)
				diags = append(diags, d...)
				if nm != "" && len(ele) > 0 {
					cniParams.CustomCniCrdSpec[nm] = ele
				}
			}

		}
	}
	return &cniParams, diags
}

func (v CustomCniCrdSpecValue) Expand(ctx context.Context) (string, []rafay.CustomCniSpec, diag.Diagnostics) {
	var diags, d diag.Diagnostics
	var spec []rafay.CustomCniSpec

	if v.IsNull() {
		return "", spec, diags
	}

	var name string
	if !v.Name.IsNull() && !v.Name.IsUnknown() {
		name = getStringValue(v.Name)
	}

	if !v.CniSpec.IsNull() && !v.CniSpec.IsUnknown() {
		vCniSpec := make([]CniSpecValue, 0, len(v.CniSpec.Elements()))
		d = v.CniSpec.ElementsAs(ctx, &vCniSpec, false)
		diags = append(diags, d...)
		if len(vCniSpec) > 0 {
			for _, vcs := range vCniSpec {
				cs, d := vcs.Expand(ctx)
				diags = append(diags, d...)
				spec = append(spec, cs)
			}
		}
	}

	return name, spec, diags
}

func (v CniSpecValue) Expand(ctx context.Context) (rafay.CustomCniSpec, diag.Diagnostics) {
	var diags, d diag.Diagnostics
	var spec rafay.CustomCniSpec
	if v.IsNull() {
		return rafay.CustomCniSpec{}, diags
	}

	if !v.Subnet.IsNull() && !v.Subnet.IsUnknown() {
		spec.Subnet = getStringValue(v.Subnet)
	}

	if !v.SecurityGroups2.IsNull() && !v.SecurityGroups2.IsUnknown() {
		vSecurityGroups := make([]types.String, 0, len(v.SecurityGroups2.Elements()))
		d = v.SecurityGroups2.ElementsAs(ctx, &vSecurityGroups, false)
		diags = append(diags, d...)
		if len(vSecurityGroups) > 0 {
			sg := make([]string, 0, len(vSecurityGroups))
			for _, s := range vSecurityGroups {
				sg = append(sg, getStringValue(s))
			}
			spec.SecurityGroups = sg
		}
	}

	return spec, diags
}

func (v SharingValue) Expand(ctx context.Context) (*rafay.V1ClusterSharing, diag.Diagnostics) {
	var diags, d diag.Diagnostics
	var sharing rafay.V1ClusterSharing

	if v.IsNull() {
		return &rafay.V1ClusterSharing{}, diags
	}

	if !v.Enabled.IsNull() && !v.Enabled.IsUnknown() {
		b := getBoolValue(v.Enabled)
		sharing.Enabled = &b
	}

	vProjectsList := make([]ProjectsValue, 0, len(v.Projects.Elements()))
	d = v.Projects.ElementsAs(ctx, &vProjectsList, false)
	diags = append(diags, d...)
	if len(vProjectsList) > 0 {
		prjs := make([]*rafay.V1ClusterSharingProject, 0, len(vProjectsList))
		for _, prj := range vProjectsList {
			p, d := prj.Expand(ctx)
			diags = append(diags, d...)
			prjs = append(prjs, p)
		}
		sharing.Projects = prjs
	}

	return &sharing, diags
}

func (v ProjectsValue) Expand(ctx context.Context) (*rafay.V1ClusterSharingProject, diag.Diagnostics) {
	var diags diag.Diagnostics
	var project rafay.V1ClusterSharingProject

	if v.IsNull() {
		return &rafay.V1ClusterSharingProject{}, diags
	}

	if !v.Name.IsNull() && !v.Name.IsUnknown() {
		project.Name = getStringValue(v.Name)
	}

	return &project, diags
}

func (v SystemComponentsPlacementValue) Expand(ctx context.Context) (*rafay.SystemComponentsPlacement, diag.Diagnostics) {
	var diags, d diag.Diagnostics
	var scp rafay.SystemComponentsPlacement

	if v.IsNull() {
		return &rafay.SystemComponentsPlacement{}, diags
	}

	vnsel := make(map[string]types.String, len(v.NodeSelector.Elements()))
	d = v.NodeSelector.ElementsAs(ctx, &vnsel, false)
	diags = append(diags, d...)
	if len(vnsel) > 0 {
		nsel := make(map[string]string, len(vnsel))
		for k, val := range vnsel {
			nsel[k] = getStringValue(val)
		}
		scp.NodeSelector = nsel
	}

	vTolerationList := make([]TolerationsValue, 0, len(v.Tolerations.Elements()))
	d = v.Tolerations.ElementsAs(ctx, &vTolerationList, false)
	diags = append(diags, d...)
	if len(vTolerationList) > 0 {
		tols := make([]*rafay.Tolerations, 0, len(vTolerationList))
		for _, tl := range vTolerationList {
			t, d := tl.Expand(ctx)
			diags = append(diags, d...)
			tols = append(tols, t)
		}
		scp.Tolerations = tols
	}

	vDaemonsetOverrideList := make([]DaemonsetOverrideValue, 0, len(v.DaemonsetOverride.Elements()))
	d = v.DaemonsetOverride.ElementsAs(ctx, &vDaemonsetOverrideList, false)
	diags = append(diags, d...)
	if len(vDaemonsetOverrideList) > 0 {
		do, d := vDaemonsetOverrideList[0].Expand(ctx)
		diags = append(diags, d...)
		scp.DaemonsetOverride = do
	}

	return &scp, diags
}

func (v DaemonsetOverrideValue) Expand(ctx context.Context) (*rafay.DaemonsetOverride, diag.Diagnostics) {
	var diags, d diag.Diagnostics
	var do rafay.DaemonsetOverride

	if v.IsNull() {
		return &rafay.DaemonsetOverride{}, diags
	}

	if !v.NodeSelectionEnabled.IsNull() && !v.NodeSelectionEnabled.IsUnknown() {
		nse := getBoolValue(v.NodeSelectionEnabled)
		do.NodeSelectionEnabled = &nse
	}

	vTolerationList := make([]Tolerations2Value, 0, len(v.Tolerations2.Elements()))
	d = v.Tolerations2.ElementsAs(ctx, &vTolerationList, false)
	diags = append(diags, d...)
	if len(vTolerationList) > 0 {
		tols := make([]*rafay.Tolerations, 0, len(vTolerationList))
		for _, tl := range vTolerationList {
			t, d := tl.Expand(ctx)
			diags = append(diags, d...)
			tols = append(tols, t)
		}
		do.Tolerations = tols
	}

	return &do, diags
}

func (v *Tolerations2Value) Expand(ctx context.Context) (*rafay.Tolerations, diag.Diagnostics) {
	var diags diag.Diagnostics
	var tol rafay.Tolerations

	if v.IsNull() {
		return &rafay.Tolerations{}, diags
	}

	if !v.Key.IsNull() && !v.Key.IsUnknown() {
		tol.Key = getStringValue(v.Key)
	}
	if !v.Operator.IsNull() && !v.Operator.IsUnknown() {
		tol.Operator = getStringValue(v.Operator)
	}
	if !v.Value.IsNull() && !v.Value.IsUnknown() {
		tol.Value = getStringValue(v.Value)
	}
	if !v.Effect.IsNull() && !v.Effect.IsUnknown() {
		tol.Effect = getStringValue(v.Effect)
	}
	if !v.TolerationSeconds.IsNull() && !v.TolerationSeconds.IsUnknown() {
		d := int(getInt64Value(v.TolerationSeconds))
		tol.TolerationSeconds = &d
	}

	return &tol, diags
}

func (v *TolerationsValue) Expand(ctx context.Context) (*rafay.Tolerations, diag.Diagnostics) {
	var diags diag.Diagnostics
	var tol rafay.Tolerations

	if v.IsNull() {
		return &rafay.Tolerations{}, diags
	}

	if !v.Key.IsNull() && !v.Key.IsUnknown() {
		tol.Key = getStringValue(v.Key)
	}
	if !v.Operator.IsNull() && !v.Operator.IsUnknown() {
		tol.Operator = getStringValue(v.Operator)
	}
	if !v.Value.IsNull() && !v.Value.IsUnknown() {
		tol.Value = getStringValue(v.Value)
	}
	if !v.Effect.IsNull() && !v.Effect.IsUnknown() {
		tol.Effect = getStringValue(v.Effect)
	}
	if !v.TolerationSeconds.IsNull() && !v.TolerationSeconds.IsUnknown() {
		d := int(getInt64Value(v.TolerationSeconds))
		tol.TolerationSeconds = &d
	}

	return &tol, diags
}

// ClusterConfig Expand

func (v ClusterConfigValue) Expand(ctx context.Context) (*rafay.EKSClusterConfig, diag.Diagnostics) {
	var diags, d diag.Diagnostics
	var clusterConfig rafay.EKSClusterConfig

	if v.IsNull() {
		return &rafay.EKSClusterConfig{}, diags
	}

	if !v.Apiversion.IsNull() && !v.Apiversion.IsUnknown() {
		clusterConfig.APIVersion = getStringValue(v.Apiversion)
	}
	if !v.Kind.IsNull() && !v.Kind.IsUnknown() {
		clusterConfig.Kind = getStringValue(v.Kind)
	}

	// metadata2 block
	vMetadata2List := make([]Metadata2Value, 0, len(v.Metadata2.Elements()))
	diags = v.Metadata2.ElementsAs(ctx, &vMetadata2List, false)
	if len(vMetadata2List) > 0 {
		md, d := vMetadata2List[0].Expand(ctx)
		diags = append(diags, d...)
		clusterConfig.Metadata = md
	}

	// node_groups block (deprecated)
	vNodeGroupsList := make([]NodeGroupsValue, 0, len(v.NodeGroups.Elements()))
	diags = v.NodeGroups.ElementsAs(ctx, &vNodeGroupsList, false)
	ngs := make([]*rafay.NodeGroup, 0, len(vNodeGroupsList))
	for _, vng := range vNodeGroupsList {
		ng, d := vng.Expand(ctx)
		diags = append(diags, d...)
		ngs = append(ngs, ng)
	}
	if len(ngs) > 0 {
		clusterConfig.NodeGroups = ngs
	}

	// node_groups_map block
	vngMap := make(map[string]NodeGroupsMapValue, len(v.NodeGroupsMap.Elements()))
	d = v.NodeGroupsMap.ElementsAs(ctx, &vngMap, false)
	diags = append(diags, d...)
	ngsMap := make([]*rafay.NodeGroup, 0, len(vngMap))
	for ngName, ngMap := range vngMap {
		ngObj, d := ngMap.Expand(ctx)
		diags = append(diags, d...)
		ngObj.Name = ngName
		ngsMap = append(ngsMap, ngObj)
	}
	if len(ngsMap) > 0 {
		clusterConfig.NodeGroups = ngsMap
	}

	// managed_nodegroups block (deprecated)
	vManagedNodeGroupsList := make([]ManagedNodegroupsValue, 0, len(v.ManagedNodegroups.Elements()))
	diags = v.ManagedNodegroups.ElementsAs(ctx, &vManagedNodeGroupsList, false)
	mngs := make([]*rafay.ManagedNodeGroup, 0, len(vManagedNodeGroupsList))
	for _, mng := range vManagedNodeGroupsList {
		mngObj, d := mng.Expand(ctx)
		diags = append(diags, d...)
		mngs = append(mngs, mngObj)
	}
	if len(mngs) > 0 {
		clusterConfig.ManagedNodeGroups = mngs
	}

	// managed_nodegroups_map block
	vmngMap := make(map[string]ManagedNodegroupsMapValue, len(v.ManagedNodegroupsMap.Elements()))
	d = v.ManagedNodegroupsMap.ElementsAs(ctx, &vmngMap, false)
	diags = append(diags, d...)
	mngsMap := make([]*rafay.ManagedNodeGroup, 0, len(vmngMap))
	for mngName, mngMap := range vmngMap {
		mngObj, d := mngMap.Expand(ctx)
		diags = append(diags, d...)
		mngObj.Name = mngName
		mngsMap = append(mngsMap, mngObj)
	}
	if len(mngsMap) > 0 {
		clusterConfig.ManagedNodeGroups = mngsMap
	}

	// availability_zones (list of strings)
	azList := make([]types.String, 0, len(v.AvailabilityZones.Elements()))
	d = v.AvailabilityZones.ElementsAs(ctx, &azList, false)
	diags = append(diags, d...)
	azs := make([]string, 0, len(azList))
	for _, az := range azList {
		azs = append(azs, getStringValue(az))
	}
	if len(azs) > 0 {
		clusterConfig.AvailabilityZones = azs
	}

	// kubernetes_network_config block
	vKNCList := make([]KubernetesNetworkConfigValue, 0, len(v.KubernetesNetworkConfig.Elements()))
	diags = v.KubernetesNetworkConfig.ElementsAs(ctx, &vKNCList, false)
	if len(vKNCList) > 0 {
		knc, d := vKNCList[0].Expand(ctx)
		diags = append(diags, d...)
		clusterConfig.KubernetesNetworkConfig = knc
	}

	// iam3 block
	vIAM3List := make([]Iam3Value, 0, len(v.Iam3.Elements()))
	diags = v.Iam3.ElementsAs(ctx, &vIAM3List, false)
	if len(vIAM3List) > 0 {
		iam, d := vIAM3List[0].Expand(ctx)
		diags = append(diags, d...)
		clusterConfig.IAM = iam
	}

	// identity_providers block
	vIdpList := make([]IdentityProvidersValue, 0, len(v.IdentityProviders.Elements()))
	diags = v.IdentityProviders.ElementsAs(ctx, &vIdpList, false)
	idps := make([]*rafay.IdentityProvider, 0, len(vIdpList))
	for _, idp := range vIdpList {
		idpObj, d := idp.Expand(ctx)
		diags = append(diags, d...)
		idps = append(idps, idpObj)
	}
	if len(idps) > 0 {
		clusterConfig.IdentityProviders = idps
	}

	// vpc block
	vVpcList := make([]VpcValue, 0, len(v.Vpc.Elements()))
	diags = v.Vpc.ElementsAs(ctx, &vVpcList, false)
	if len(vVpcList) > 0 {
		vpc, d := vVpcList[0].Expand(ctx)
		diags = append(diags, d...)
		clusterConfig.VPC = vpc
	}

	// addons block
	vAddonsList := make([]AddonsValue, 0, len(v.Addons.Elements()))
	diags = v.Addons.ElementsAs(ctx, &vAddonsList, false)
	addons := make([]*rafay.Addon, 0, len(vAddonsList))
	for _, addon := range vAddonsList {
		addonObj, d := addon.Expand(ctx)
		diags = append(diags, d...)
		addons = append(addons, addonObj)
	}
	if len(addons) > 0 {
		clusterConfig.Addons = addons
	}

	// private_cluster block
	vPrivateClusterList := make([]PrivateClusterValue, 0, len(v.PrivateCluster.Elements()))
	diags = v.PrivateCluster.ElementsAs(ctx, &vPrivateClusterList, false)
	if len(vPrivateClusterList) > 0 {
		pc, d := vPrivateClusterList[0].Expand(ctx)
		diags = append(diags, d...)
		clusterConfig.PrivateCluster = pc
	}

	// fargate_profiles block
	vFargateProfilesList := make([]FargateProfilesValue, 0, len(v.FargateProfiles.Elements()))
	diags = v.FargateProfiles.ElementsAs(ctx, &vFargateProfilesList, false)
	fps := make([]*rafay.FargateProfile, 0, len(vFargateProfilesList))
	for _, fp := range vFargateProfilesList {
		fpObj, d := fp.Expand(ctx)
		diags = append(diags, d...)
		fps = append(fps, fpObj)
	}
	if len(fps) > 0 {
		clusterConfig.FargateProfiles = fps
	}

	// cloud_watch block
	vCloudWatchList := make([]CloudWatchValue, 0, len(v.CloudWatch.Elements()))
	diags = v.CloudWatch.ElementsAs(ctx, &vCloudWatchList, false)
	if len(vCloudWatchList) > 0 {
		cw, d := vCloudWatchList[0].Expand(ctx)
		diags = append(diags, d...)
		clusterConfig.CloudWatch = cw
	}

	// secrets_encryption block
	vSecretsEncryptionList := make([]SecretsEncryptionValue, 0, len(v.SecretsEncryption.Elements()))
	diags = v.SecretsEncryption.ElementsAs(ctx, &vSecretsEncryptionList, false)
	if len(vSecretsEncryptionList) > 0 {
		se, d := vSecretsEncryptionList[0].Expand(ctx)
		diags = append(diags, d...)
		clusterConfig.SecretsEncryption = se
	}

	// identity_mappings block
	vIdentityMappingsList := make([]IdentityMappingsValue, 0, len(v.IdentityMappings.Elements()))
	diags = v.IdentityMappings.ElementsAs(ctx, &vIdentityMappingsList, false)
	if len(vIdentityMappingsList) > 0 {
		im, d := vIdentityMappingsList[0].Expand(ctx)
		diags = append(diags, d...)
		clusterConfig.IdentityMappings = im
	}

	// access_config block
	vAccessConfigList := make([]AccessConfigValue, 0, len(v.AccessConfig.Elements()))
	diags = v.AccessConfig.ElementsAs(ctx, &vAccessConfigList, false)
	if len(vAccessConfigList) > 0 {
		ac, d := vAccessConfigList[0].Expand(ctx)
		diags = append(diags, d...)
		clusterConfig.AccessConfig = ac
	}

	// addons_config block
	vAddonsConfigList := make([]AddonsConfigValue, 0, len(v.AddonsConfig.Elements()))
	diags = v.AddonsConfig.ElementsAs(ctx, &vAddonsConfigList, false)
	if len(vAddonsConfigList) > 0 {
		acfg, d := vAddonsConfigList[0].Expand(ctx)
		diags = append(diags, d...)
		clusterConfig.AddonsConfig = acfg
	}

	// auto_mode_config block
	vAutoModeConfigList := make([]AutoModeConfigValue, 0, len(v.AutoModeConfig.Elements()))
	diags = v.AutoModeConfig.ElementsAs(ctx, &vAutoModeConfigList, false)
	if len(vAutoModeConfigList) > 0 {
		amc, d := vAutoModeConfigList[0].Expand(ctx)
		diags = append(diags, d...)
		clusterConfig.AutoModeConfig = amc
	}

	return &clusterConfig, diags
}

// Dedicated Expand functions for each block type
func (v Metadata2Value) Expand(ctx context.Context) (*rafay.EKSClusterConfigMetadata, diag.Diagnostics) {
	var diags, d diag.Diagnostics
	var md rafay.EKSClusterConfigMetadata
	if v.IsNull() {
		return &rafay.EKSClusterConfigMetadata{}, diags
	}
	md.Name = getStringValue(v.Name)
	md.Region = getStringValue(v.Region)
	md.Version = getStringValue(v.Version)

	tags := make(map[string]string, len(v.Tags.Elements()))
	vTags := make(map[string]types.String, len(v.Tags.Elements()))
	d = v.Tags.ElementsAs(ctx, &vTags, false)
	diags = append(diags, d...)
	for k, val := range vTags {
		tags[k] = getStringValue(val)
	}
	md.Tags = tags

	annotations := make(map[string]string, len(v.Annotations.Elements()))
	vAnnotations := make(map[string]types.String, len(v.Annotations.Elements()))
	d = v.Annotations.ElementsAs(ctx, &vAnnotations, false)
	diags = append(diags, d...)
	for k, val := range vAnnotations {
		annotations[k] = getStringValue(val)
	}
	md.Annotations = annotations

	return &md, diags
}

func (v KubernetesNetworkConfigValue) Expand(ctx context.Context) (*rafay.KubernetesNetworkConfig, diag.Diagnostics) {
	var diags diag.Diagnostics
	var knc rafay.KubernetesNetworkConfig

	if v.IsNull() {
		return &rafay.KubernetesNetworkConfig{}, diags
	}

	// Map string fields
	if !v.IpFamily.IsNull() && !v.IpFamily.IsUnknown() {
		knc.IPFamily = getStringValue(v.IpFamily)
	}

	if !v.ServiceIpv4Cidr.IsNull() && !v.ServiceIpv4Cidr.IsUnknown() {
		knc.ServiceIPv4CIDR = getStringValue(v.ServiceIpv4Cidr)
	}

	return &knc, diags
}

func (v Iam3Value) Expand(ctx context.Context) (*rafay.EKSClusterIAM, diag.Diagnostics) {
	var diags, d diag.Diagnostics
	var iam rafay.EKSClusterIAM

	if v.IsNull() {
		return &rafay.EKSClusterIAM{}, diags
	}

	// Map string fields
	if !v.ServiceRoleArn.IsNull() && !v.ServiceRoleArn.IsUnknown() {
		iam.ServiceRoleARN = getStringValue(v.ServiceRoleArn)
	}

	if !v.ServiceRolePermissionBoundary.IsNull() && !v.ServiceRolePermissionBoundary.IsUnknown() {
		iam.ServiceRolePermissionsBoundary = getStringValue(v.ServiceRolePermissionBoundary)
	}

	if !v.FargatePodExecutionRoleArn.IsNull() && !v.FargatePodExecutionRoleArn.IsUnknown() {
		iam.FargatePodExecutionRoleARN = getStringValue(v.FargatePodExecutionRoleArn)
	}

	if !v.FargatePodExecutionRolePermissionsBoundary.IsNull() && !v.FargatePodExecutionRolePermissionsBoundary.IsUnknown() {
		iam.FargatePodExecutionRolePermissionsBoundary = getStringValue(v.FargatePodExecutionRolePermissionsBoundary)
	}

	// Map boolean fields
	if !v.WithOidc.IsNull() && !v.WithOidc.IsUnknown() {
		withOidc := getBoolValue(v.WithOidc)
		iam.WithOIDC = &withOidc
	}

	if !v.VpcResourceControllerPolicy.IsNull() && !v.VpcResourceControllerPolicy.IsUnknown() {
		vpcResourceControllerPolicy := getBoolValue(v.VpcResourceControllerPolicy)
		iam.VPCResourceControllerPolicy = &vpcResourceControllerPolicy
	}

	// Map service_accounts block
	if !v.ServiceAccounts.IsNull() && !v.ServiceAccounts.IsUnknown() {
		vServiceAccountsList := make([]ServiceAccountsValue, 0, len(v.ServiceAccounts.Elements()))
		d = v.ServiceAccounts.ElementsAs(ctx, &vServiceAccountsList, false)
		diags = append(diags, d...)
		serviceAccounts := make([]*rafay.EKSClusterIAMServiceAccount, 0, len(vServiceAccountsList))
		for _, sa := range vServiceAccountsList {
			saObj, d := sa.Expand(ctx)
			diags = append(diags, d...)
			serviceAccounts = append(serviceAccounts, saObj)
		}
		if len(serviceAccounts) > 0 {
			iam.ServiceAccounts = serviceAccounts
		}
	}

	// Map pod_identity_associations block
	if !v.PodIdentityAssociations.IsNull() && !v.PodIdentityAssociations.IsUnknown() {
		vPodIdentityAssociationsList := make([]PodIdentityAssociationsValue, 0, len(v.PodIdentityAssociations.Elements()))
		d = v.PodIdentityAssociations.ElementsAs(ctx, &vPodIdentityAssociationsList, false)
		diags = append(diags, d...)
		podIdentityAssociations := make([]*rafay.IAMPodIdentityAssociation, 0, len(vPodIdentityAssociationsList))
		for _, pia := range vPodIdentityAssociationsList {
			piaObj, d := pia.Expand(ctx)
			diags = append(diags, d...)
			podIdentityAssociations = append(podIdentityAssociations, piaObj)
		}
		if len(podIdentityAssociations) > 0 {
			iam.PodIdentityAssociations = podIdentityAssociations
		}
	}

	return &iam, diags
}

// PodIdentityAssociationsValue Expand
func (v PodIdentityAssociationsValue) Expand(ctx context.Context) (*rafay.IAMPodIdentityAssociation, diag.Diagnostics) {
	var diags, d diag.Diagnostics
	var pia rafay.IAMPodIdentityAssociation

	if v.IsNull() {
		return &rafay.IAMPodIdentityAssociation{}, diags
	}

	// Map string fields
	if !v.Namespace.IsNull() && !v.Namespace.IsUnknown() {
		pia.Namespace = getStringValue(v.Namespace)
	}

	if !v.ServiceAccountName.IsNull() && !v.ServiceAccountName.IsUnknown() {
		pia.ServiceAccountName = getStringValue(v.ServiceAccountName)
	}

	if !v.RoleArn.IsNull() && !v.RoleArn.IsUnknown() {
		pia.RoleARN = getStringValue(v.RoleArn)
	}

	if !v.RoleName.IsNull() && !v.RoleName.IsUnknown() {
		pia.RoleName = getStringValue(v.RoleName)
	}

	if !v.PermissionBoundaryArn.IsNull() && !v.PermissionBoundaryArn.IsUnknown() {
		pia.PermissionsBoundaryARN = getStringValue(v.PermissionBoundaryArn)
	}

	// Map boolean fields
	if !v.CreateServiceAccount.IsNull() && !v.CreateServiceAccount.IsUnknown() {
		pia.CreateServiceAccount = getBoolValue(v.CreateServiceAccount)
	}

	// Map permission_policy (JSON string to map)
	if !v.PermissionPolicy.IsNull() && !v.PermissionPolicy.IsUnknown() {
		policyStr := getStringValue(v.PermissionPolicy)
		if policyStr != "" {
			var policyDoc map[string]interface{}
			var json2 = jsoniter.ConfigCompatibleWithStandardLibrary
			json2.Unmarshal([]byte(policyStr), &policyDoc)
			pia.PermissionPolicy = policyDoc
		}
	}

	// Map permission_policy_arns (list of strings)
	if !v.PermissionPolicyArns.IsNull() && !v.PermissionPolicyArns.IsUnknown() {
		policyArnsList := make([]types.String, 0, len(v.PermissionPolicyArns.Elements()))
		d = v.PermissionPolicyArns.ElementsAs(ctx, &policyArnsList, false)
		diags = append(diags, d...)
		policyArns := make([]string, 0, len(policyArnsList))
		for _, arn := range policyArnsList {
			policyArns = append(policyArns, getStringValue(arn))
		}
		if len(policyArns) > 0 {
			pia.PermissionPolicyARNs = policyArns
		}
	}

	// Map well_known_policies block
	if !v.WellKnownPolicies.IsNull() && !v.WellKnownPolicies.IsUnknown() {
		vWellKnownPoliciesList := make([]WellKnownPoliciesValue, 0, len(v.WellKnownPolicies.Elements()))
		d = v.WellKnownPolicies.ElementsAs(ctx, &vWellKnownPoliciesList, false)
		diags = append(diags, d...)
		if len(vWellKnownPoliciesList) > 0 {
			wellKnownPolicies, d := vWellKnownPoliciesList[0].Expand(ctx)
			diags = append(diags, d...)
			pia.WellKnownPolicies = wellKnownPolicies
		}
	}

	// Map tags (map of strings)
	if !v.Tags.IsNull() && !v.Tags.IsUnknown() {
		tags := make(map[string]string, len(v.Tags.Elements()))
		vTags := make(map[string]types.String, len(v.Tags.Elements()))
		d = v.Tags.ElementsAs(ctx, &vTags, false)
		diags = append(diags, d...)
		for k, val := range vTags {
			tags[k] = getStringValue(val)
		}
		if len(tags) > 0 {
			pia.Tags = tags
		}
	}

	return &pia, diags
}

// WellKnownPoliciesValue Expand
func (v WellKnownPoliciesValue) Expand(ctx context.Context) (*rafay.WellKnownPolicies, diag.Diagnostics) {
	var diags diag.Diagnostics
	var wp rafay.WellKnownPolicies

	if v.IsNull() {
		return &rafay.WellKnownPolicies{}, diags
	}

	// Map boolean fields to pointers (following the same pattern as expandIAMWellKnownPolicies)
	if !v.ImageBuilder.IsNull() && !v.ImageBuilder.IsUnknown() {
		imageBuilder := getBoolValue(v.ImageBuilder)
		wp.ImageBuilder = &imageBuilder
	}

	if !v.AutoScaler.IsNull() && !v.AutoScaler.IsUnknown() {
		autoScaler := getBoolValue(v.AutoScaler)
		wp.AutoScaler = &autoScaler
	}

	if !v.AwsLoadBalancerController.IsNull() && !v.AwsLoadBalancerController.IsUnknown() {
		awsLoadBalancerController := getBoolValue(v.AwsLoadBalancerController)
		wp.AWSLoadBalancerController = &awsLoadBalancerController
	}

	if !v.ExternalDns.IsNull() && !v.ExternalDns.IsUnknown() {
		externalDns := getBoolValue(v.ExternalDns)
		wp.ExternalDNS = &externalDns
	}

	if !v.CertManager.IsNull() && !v.CertManager.IsUnknown() {
		certManager := getBoolValue(v.CertManager)
		wp.CertManager = &certManager
	}

	if !v.EbsCsiController.IsNull() && !v.EbsCsiController.IsUnknown() {
		ebsCsiController := getBoolValue(v.EbsCsiController)
		wp.EBSCSIController = &ebsCsiController
	}

	if !v.EfsCsiController.IsNull() && !v.EfsCsiController.IsUnknown() {
		efsCsiController := getBoolValue(v.EfsCsiController)
		wp.EFSCSIController = &efsCsiController
	}

	return &wp, diags
}

// ServiceAccountsValue Expand
func (v ServiceAccountsValue) Expand(ctx context.Context) (*rafay.EKSClusterIAMServiceAccount, diag.Diagnostics) {
	var diags, d diag.Diagnostics
	var sa rafay.EKSClusterIAMServiceAccount

	if v.IsNull() {
		return &rafay.EKSClusterIAMServiceAccount{}, diags
	}

	// Map metadata3 block
	if !v.Metadata3.IsNull() && !v.Metadata3.IsUnknown() {
		vMetadata3List := make([]Metadata3Value, 0, len(v.Metadata3.Elements()))
		d = v.Metadata3.ElementsAs(ctx, &vMetadata3List, false)
		diags = append(diags, d...)
		if len(vMetadata3List) > 0 {
			metadata, d := vMetadata3List[0].Expand(ctx)
			diags = append(diags, d...)
			sa.Metadata = metadata
		}
	}

	// Map attach_policy_arns2 (list of strings)
	if !v.AttachPolicyArns2.IsNull() && !v.AttachPolicyArns2.IsUnknown() {
		policyArnsList := make([]types.String, 0, len(v.AttachPolicyArns2.Elements()))
		d = v.AttachPolicyArns2.ElementsAs(ctx, &policyArnsList, false)
		diags = append(diags, d...)
		policyArns := make([]string, 0, len(policyArnsList))
		for _, arn := range policyArnsList {
			policyArns = append(policyArns, getStringValue(arn))
		}
		if len(policyArns) > 0 {
			sa.AttachPolicyARNs = policyArns
		}
	}

	// Map well_known_policies2 block
	if !v.WellKnownPolicies2.IsNull() && !v.WellKnownPolicies2.IsUnknown() {
		vWellKnownPolicies2List := make([]WellKnownPolicies2Value, 0, len(v.WellKnownPolicies2.Elements()))
		d = v.WellKnownPolicies2.ElementsAs(ctx, &vWellKnownPolicies2List, false)
		diags = append(diags, d...)
		if len(vWellKnownPolicies2List) > 0 {
			wellKnownPolicies, d := vWellKnownPolicies2List[0].Expand(ctx)
			diags = append(diags, d...)
			sa.WellKnownPolicies = wellKnownPolicies
		}
	}

	// Map attach_policy (JSON string to map)
	if !v.AttachPolicy.IsNull() && !v.AttachPolicy.IsUnknown() {
		policyStr := getStringValue(v.AttachPolicy)
		if policyStr != "" {
			var policyDoc map[string]interface{}
			var json2 = jsoniter.ConfigCompatibleWithStandardLibrary
			json2.Unmarshal([]byte(policyStr), &policyDoc)
			sa.AttachPolicy = policyDoc
		}
	}

	// Map string fields
	if !v.AttachRoleArn.IsNull() && !v.AttachRoleArn.IsUnknown() {
		sa.AttachRoleARN = getStringValue(v.AttachRoleArn)
	}

	if !v.PermissionsBoundary.IsNull() && !v.PermissionsBoundary.IsUnknown() {
		sa.PermissionsBoundary = getStringValue(v.PermissionsBoundary)
	}

	if !v.RoleName.IsNull() && !v.RoleName.IsUnknown() {
		sa.RoleName = getStringValue(v.RoleName)
	}

	// Map boolean fields
	if !v.RoleOnly.IsNull() && !v.RoleOnly.IsUnknown() {
		roleOnly := getBoolValue(v.RoleOnly)
		sa.RoleOnly = &roleOnly
	}

	// Map status block
	if !v.Status.IsNull() && !v.Status.IsUnknown() {
		vStatusList := make([]StatusValue, 0, len(v.Status.Elements()))
		d = v.Status.ElementsAs(ctx, &vStatusList, false)
		diags = append(diags, d...)
		if len(vStatusList) > 0 {
			status, d := vStatusList[0].Expand(ctx)
			diags = append(diags, d...)
			sa.Status = status
		}
	}

	// Map tags3 (map of strings)
	if !v.Tags3.IsNull() && !v.Tags3.IsUnknown() {
		tags := make(map[string]string, len(v.Tags3.Elements()))
		vTags := make(map[string]types.String, len(v.Tags3.Elements()))
		d = v.Tags3.ElementsAs(ctx, &vTags, false)
		diags = append(diags, d...)
		for k, val := range vTags {
			tags[k] = getStringValue(val)
		}
		if len(tags) > 0 {
			sa.Tags = tags
		}
	}

	return &sa, diags
}

// Metadata3Value Expand
func (v Metadata3Value) Expand(ctx context.Context) (*rafay.EKSClusterIAMMeta, diag.Diagnostics) {
	var diags diag.Diagnostics
	var meta rafay.EKSClusterIAMMeta

	if v.IsNull() {
		return &rafay.EKSClusterIAMMeta{}, diags
	}

	// Map string fields
	if !v.Name.IsNull() && !v.Name.IsUnknown() {
		meta.Name = getStringValue(v.Name)
	}

	if !v.Namespace.IsNull() && !v.Namespace.IsUnknown() {
		meta.Namespace = getStringValue(v.Namespace)
	}

	// Map labels (map of strings)
	if !v.Labels.IsNull() && !v.Labels.IsUnknown() {
		labels := make(map[string]string, len(v.Labels.Elements()))
		vLabels := make(map[string]types.String, len(v.Labels.Elements()))
		d := v.Labels.ElementsAs(ctx, &vLabels, false)
		diags = append(diags, d...)
		for k, val := range vLabels {
			labels[k] = getStringValue(val)
		}
		if len(labels) > 0 {
			meta.Labels = labels
		}
	}

	// Map annotations (map of strings)
	if !v.Annotations.IsNull() && !v.Annotations.IsUnknown() {
		annotations := make(map[string]string, len(v.Annotations.Elements()))
		vAnnotations := make(map[string]types.String, len(v.Annotations.Elements()))
		d := v.Annotations.ElementsAs(ctx, &vAnnotations, false)
		diags = append(diags, d...)
		for k, val := range vAnnotations {
			annotations[k] = getStringValue(val)
		}
		if len(annotations) > 0 {
			meta.Annotations = annotations
		}
	}

	return &meta, diags
}

// WellKnownPolicies2Value Expand
func (v WellKnownPolicies2Value) Expand(ctx context.Context) (*rafay.WellKnownPolicies, diag.Diagnostics) {
	var diags diag.Diagnostics
	var wp rafay.WellKnownPolicies

	if v.IsNull() {
		return &rafay.WellKnownPolicies{}, diags
	}

	// Map boolean fields to pointers (following the same pattern as WellKnownPoliciesValue.Expand)
	if !v.ImageBuilder.IsNull() && !v.ImageBuilder.IsUnknown() {
		imageBuilder := getBoolValue(v.ImageBuilder)
		wp.ImageBuilder = &imageBuilder
	}

	if !v.AutoScaler.IsNull() && !v.AutoScaler.IsUnknown() {
		autoScaler := getBoolValue(v.AutoScaler)
		wp.AutoScaler = &autoScaler
	}

	if !v.AwsLoadBalancerController.IsNull() && !v.AwsLoadBalancerController.IsUnknown() {
		awsLoadBalancerController := getBoolValue(v.AwsLoadBalancerController)
		wp.AWSLoadBalancerController = &awsLoadBalancerController
	}

	if !v.ExternalDns.IsNull() && !v.ExternalDns.IsUnknown() {
		externalDns := getBoolValue(v.ExternalDns)
		wp.ExternalDNS = &externalDns
	}

	if !v.CertManager.IsNull() && !v.CertManager.IsUnknown() {
		certManager := getBoolValue(v.CertManager)
		wp.CertManager = &certManager
	}

	if !v.EbsCsiController.IsNull() && !v.EbsCsiController.IsUnknown() {
		ebsCsiController := getBoolValue(v.EbsCsiController)
		wp.EBSCSIController = &ebsCsiController
	}

	if !v.EfsCsiController.IsNull() && !v.EfsCsiController.IsUnknown() {
		efsCsiController := getBoolValue(v.EfsCsiController)
		wp.EFSCSIController = &efsCsiController
	}

	return &wp, diags
}

// StatusValue Expand
func (v StatusValue) Expand(ctx context.Context) (*rafay.ClusterIAMServiceAccountStatus, diag.Diagnostics) {
	var diags diag.Diagnostics
	var status rafay.ClusterIAMServiceAccountStatus

	if v.IsNull() {
		return &rafay.ClusterIAMServiceAccountStatus{}, diags
	}

	// Map role_arn field
	if !v.RoleArn.IsNull() && !v.RoleArn.IsUnknown() {
		status.RoleARN = getStringValue(v.RoleArn)
	}

	return &status, diags
}

func (v IdentityProvidersValue) Expand(ctx context.Context) (*rafay.IdentityProvider, diag.Diagnostics) {
	var diags diag.Diagnostics
	var idp rafay.IdentityProvider

	if v.IsNull() {
		return &rafay.IdentityProvider{}, diags
	}

	// Map string fields
	if !v.IdentityProvidersType.IsNull() && !v.IdentityProvidersType.IsUnknown() {
		idp.Type = getStringValue(v.IdentityProvidersType)
	}

	return &idp, diags
}

func (v VpcValue) Expand(ctx context.Context) (*rafay.EKSClusterVPC, diag.Diagnostics) {
	var diags diag.Diagnostics
	var vpc rafay.EKSClusterVPC
	if v.IsNull() {
		return &rafay.EKSClusterVPC{}, diags
	}

	if !v.Id.IsNull() && !v.Id.IsUnknown() {
		vpc.ID = getStringValue(v.Id)
	}
	if !v.Cidr.IsNull() && !v.Cidr.IsUnknown() {
		vpc.CIDR = getStringValue(v.Cidr)
	}
	if !v.Ipv6Cidr.IsNull() && !v.Ipv6Cidr.IsUnknown() {
		vpc.IPv6Cidr = getStringValue(v.Ipv6Cidr)
	}
	if !v.Ipv6Pool.IsNull() && !v.Ipv6Pool.IsUnknown() {
		vpc.IPv6Pool = getStringValue(v.Ipv6Pool)
	}
	if !v.SecurityGroup.IsNull() && !v.SecurityGroup.IsUnknown() {
		vpc.SecurityGroup = getStringValue(v.SecurityGroup)
	}
	if !v.ExtraCidrs.IsNull() && !v.ExtraCidrs.IsUnknown() {
		extraCidrsList := make([]types.String, 0, len(v.ExtraCidrs.Elements()))
		d := v.ExtraCidrs.ElementsAs(ctx, &extraCidrsList, false)
		diags = append(diags, d...)
		extraCidrs := make([]string, 0, len(extraCidrsList))
		for _, cidr := range extraCidrsList {
			extraCidrs = append(extraCidrs, getStringValue(cidr))
		}
		if len(extraCidrs) > 0 {
			vpc.ExtraCIDRs = extraCidrs
		}
	}
	if !v.ExtraIpv6Cidrs.IsNull() && !v.ExtraIpv6Cidrs.IsUnknown() {
		extraIpv6CidrsList := make([]types.String, 0, len(v.ExtraIpv6Cidrs.Elements()))
		d := v.ExtraIpv6Cidrs.ElementsAs(ctx, &extraIpv6CidrsList, false)
		diags = append(diags, d...)
		extraIpv6Cidrs := make([]string, 0, len(extraIpv6CidrsList))
		for _, cidr := range extraIpv6CidrsList {
			extraIpv6Cidrs = append(extraIpv6Cidrs, getStringValue(cidr))
		}
		if len(extraIpv6Cidrs) > 0 {
			vpc.ExtraIPv6CIDRs = extraIpv6Cidrs
		}
	}
	if !v.SharedNodeSecurityGroup.IsNull() && !v.SharedNodeSecurityGroup.IsUnknown() {
		vpc.SharedNodeSecurityGroup = getStringValue(v.SharedNodeSecurityGroup)
	}
	if !v.ManageSharedNodeSecurityGroupRules.IsNull() && !v.ManageSharedNodeSecurityGroupRules.IsUnknown() {
		manage := getBoolValue(v.ManageSharedNodeSecurityGroupRules)
		vpc.ManageSharedNodeSecurityGroupRules = &manage
	}
	if !v.AutoAllocateIpv6.IsNull() && !v.AutoAllocateIpv6.IsUnknown() {
		autoallocate := getBoolValue(v.AutoAllocateIpv6)
		vpc.AutoAllocateIPv6 = &autoallocate
	}
	if !v.PublicAccessCidrs.IsNull() && !v.PublicAccessCidrs.IsUnknown() {
		publicAccessCidrsList := make([]types.String, 0, len(v.PublicAccessCidrs.Elements()))
		d := v.PublicAccessCidrs.ElementsAs(ctx, &publicAccessCidrsList, false)
		diags = append(diags, d...)
		publicAccessCidrs := make([]string, 0, len(publicAccessCidrsList))
		for _, cidr := range publicAccessCidrsList {
			publicAccessCidrs = append(publicAccessCidrs, getStringValue(cidr))
		}
		if len(publicAccessCidrs) > 0 {
			vpc.PublicAccessCIDRs = publicAccessCidrs
		}
	}

	if !v.Subnets3.IsNull() && !v.Subnets3.IsUnknown() {
		vSubnets := make([]Subnets3Value, 0, len(v.Subnets3.Elements()))
		d := v.Subnets3.ElementsAs(ctx, &vSubnets, false)
		diags = append(diags, d...)
		if len(vSubnets) > 0 {
			sn, d := vSubnets[0].Expand(ctx)
			diags = append(diags, d...)
			vpc.Subnets = sn
		}
	}

	if !v.Nat.IsNull() && !v.Nat.IsUnknown() {
		vNatList := make([]NatValue, 0, len(v.Nat.Elements()))
		d := v.Nat.ElementsAs(ctx, &vNatList, false)
		diags = append(diags, d...)
		if len(vNatList) > 0 {
			nat, d := vNatList[0].Expand(ctx)
			diags = append(diags, d...)
			vpc.NAT = nat
		}
	}

	if !v.ClusterEndpoints.IsNull() && !v.ClusterEndpoints.IsUnknown() {
		vClusterEndpointsList := make([]ClusterEndpointsValue, 0, len(v.ClusterEndpoints.Elements()))
		d := v.ClusterEndpoints.ElementsAs(ctx, &vClusterEndpointsList, false)
		diags = append(diags, d...)
		if len(vClusterEndpointsList) > 0 {
			ce, d := vClusterEndpointsList[0].Expand(ctx)
			diags = append(diags, d...)
			vpc.ClusterEndpoints = ce
		}
	}

	return &vpc, diags
}

func (v Subnets3Value) Expand(ctx context.Context) (*rafay.ClusterSubnets, diag.Diagnostics) {
	var diags diag.Diagnostics
	var subnets rafay.ClusterSubnets
	if v.IsNull() {
		return &rafay.ClusterSubnets{}, diags
	}
	if !v.Private.IsNull() && !v.Private.IsUnknown() {
		privateList := make([]PrivateValue, 0, len(v.Private.Elements()))
		d := v.Private.ElementsAs(ctx, &privateList, false)
		diags = append(diags, d...)
		if len(privateList) > 0 {
			snmapping := make(rafay.AZSubnetMapping)
			for _, ps := range privateList {
				nm, azspec, d := ps.Expand(ctx)
				diags = append(diags, d...)
				snmapping[nm] = *azspec
			}
			subnets.Private = snmapping
		}
	}

	if !v.Public.IsNull() && !v.Public.IsUnknown() {
		publicList := make([]PublicValue, 0, len(v.Public.Elements()))
		d := v.Public.ElementsAs(ctx, &publicList, false)
		diags = append(diags, d...)
		if len(publicList) > 0 {
			snmapping := make(rafay.AZSubnetMapping)
			for _, ps := range publicList {
				nm, azspec, d := ps.Expand(ctx)
				diags = append(diags, d...)
				snmapping[nm] = *azspec
			}
			subnets.Public = snmapping
		}
	}

	return &subnets, diags
}

func (v PrivateValue) Expand(ctx context.Context) (string, *rafay.AZSubnetSpec, diag.Diagnostics) {
	var diags diag.Diagnostics
	var azspec rafay.AZSubnetSpec
	if v.IsNull() {
		return "", &rafay.AZSubnetSpec{}, diags
	}
	var name string
	if !v.Name.IsNull() && !v.Name.IsUnknown() {
		name = getStringValue(v.Name)
	}
	if !v.Id.IsNull() && !v.Id.IsUnknown() {
		azspec.ID = getStringValue(v.Id)
	}
	if !v.Cidr.IsNull() && !v.Cidr.IsUnknown() {
		azspec.CIDR = getStringValue(v.Cidr)
	}
	if !v.Az.IsNull() && !v.Az.IsUnknown() {
		azspec.AZ = getStringValue(v.Az)
	}

	return name, &azspec, diags
}

func (v PublicValue) Expand(ctx context.Context) (string, *rafay.AZSubnetSpec, diag.Diagnostics) {
	var diags diag.Diagnostics
	var azspec rafay.AZSubnetSpec
	if v.IsNull() {
		return "", &rafay.AZSubnetSpec{}, diags
	}
	var name string
	if !v.Name.IsNull() && !v.Name.IsUnknown() {
		name = getStringValue(v.Name)
	}
	if !v.Id.IsNull() && !v.Id.IsUnknown() {
		azspec.ID = getStringValue(v.Id)
	}
	if !v.Cidr.IsNull() && !v.Cidr.IsUnknown() {
		azspec.CIDR = getStringValue(v.Cidr)
	}
	if !v.Az.IsNull() && !v.Az.IsUnknown() {
		azspec.AZ = getStringValue(v.Az)
	}
	return name, &azspec, diags
}

func (v NatValue) Expand(ctx context.Context) (*rafay.ClusterNAT, diag.Diagnostics) {
	var diags diag.Diagnostics
	var nat rafay.ClusterNAT
	if v.IsNull() {
		return &rafay.ClusterNAT{}, diags
	}

	if !v.Gateway.IsNull() && !v.Gateway.IsUnknown() {
		nat.Gateway = getStringValue(v.Gateway)
	}

	return &nat, diags
}

func (v ClusterEndpointsValue) Expand(ctx context.Context) (*rafay.ClusterEndpoints, diag.Diagnostics) {
	var diags diag.Diagnostics
	var ce rafay.ClusterEndpoints
	if v.IsNull() {
		return &rafay.ClusterEndpoints{}, diags
	}

	if !v.PublicAccess.IsNull() && !v.PublicAccess.IsUnknown() {
		publicAccess := getBoolValue(v.PublicAccess)
		ce.PublicAccess = &publicAccess
	}

	if !v.PrivateAccess.IsNull() && !v.PrivateAccess.IsUnknown() {
		privateAccess := getBoolValue(v.PrivateAccess)
		ce.PrivateAccess = &privateAccess
	}
	return &ce, diags
}

func (v AddonsValue) Expand(ctx context.Context) (*rafay.Addon, diag.Diagnostics) {
	var diags, d diag.Diagnostics
	var addon rafay.Addon
	if v.IsNull() {
		return &rafay.Addon{}, diags
	}
	if !v.Name.IsNull() && !v.Name.IsUnknown() {
		addon.Name = getStringValue(v.Name)
	}
	if !v.Version.IsNull() && !v.Version.IsUnknown() {
		addon.Version = getStringValue(v.Version)
	}
	if !v.ServiceAccountRoleArn.IsNull() && !v.ServiceAccountRoleArn.IsUnknown() {
		addon.ServiceAccountRoleARN = getStringValue(v.ServiceAccountRoleArn)
	}
	if !v.AttachPolicyArns3.IsNull() && !v.AttachPolicyArns3.IsUnknown() {
		policyArnsList := make([]types.String, 0, len(v.AttachPolicyArns3.Elements()))
		d := v.AttachPolicyArns3.ElementsAs(ctx, &policyArnsList, false)
		diags = append(diags, d...)
		policyArns := make([]string, 0, len(policyArnsList))
		for _, arn := range policyArnsList {
			policyArns = append(policyArns, getStringValue(arn))
		}
		if len(policyArns) > 0 {
			addon.AttachPolicyARNs = policyArns
		}
	}
	if !v.AttachPolicyV22.IsNull() && !v.AttachPolicyV22.IsUnknown() {
		var policyDoc *rafay.InlineDocument
		var json2 = jsoniter.ConfigCompatibleWithStandardLibrary
		json2.Unmarshal([]byte(getStringValue(v.AttachPolicyV22)), &policyDoc)
		addon.AttachPolicy = policyDoc
	}
	if !v.PermissionsBoundary2.IsNull() && !v.PermissionsBoundary2.IsUnknown() {
		addon.PermissionsBoundary = getStringValue(v.PermissionsBoundary2)
	}
	if !v.Tags4.IsNull() && !v.Tags4.IsUnknown() {
		tags := make(map[string]string, len(v.Tags4.Elements()))
		vTags := make(map[string]types.String, len(v.Tags4.Elements()))
		d := v.Tags4.ElementsAs(ctx, &vTags, false)
		diags = append(diags, d...)
		for key, val := range vTags {
			tags[key] = getStringValue(val)
		}
		if len(tags) > 0 {
			addon.Tags = tags
		}
	}
	if !v.ConfigurationValues.IsNull() && !v.ConfigurationValues.IsUnknown() {
		addon.ConfigurationValues = getStringValue(v.ConfigurationValues)
	}
	if !v.UseDefaultPodIdentityAssociations.IsNull() && !v.UseDefaultPodIdentityAssociations.IsUnknown() {
		addon.UseDefaultPodIdentityAssociations = getBoolValue(v.UseDefaultPodIdentityAssociations)
	}

	if !v.AttachPolicy3.IsNull() && !v.AttachPolicy3.IsUnknown() {
		vAttachPolicy := make([]AttachPolicy3Value, 0, len(v.AttachPolicy3.Elements()))
		d = v.AttachPolicy3.ElementsAs(ctx, &vAttachPolicy, false)
		diags = append(diags, d...)
		if len(vAttachPolicy) > 0 {
			policy, d := vAttachPolicy[0].Expand(ctx)
			diags = append(diags, d...)
			addon.AttachPolicy = policy
		}
	}

	if !v.WellKnownPolicies3.IsNull() && !v.WellKnownPolicies3.IsUnknown() {
		vWellKnownPolicies := make([]WellKnownPolicies3Value, 0, len(v.WellKnownPolicies3.Elements()))
		d = v.WellKnownPolicies3.ElementsAs(ctx, &vWellKnownPolicies, false)
		diags = append(diags, d...)
		if len(vWellKnownPolicies) > 0 {
			wellKnownPolicies, d := vWellKnownPolicies[0].Expand(ctx)
			diags = append(diags, d...)
			addon.WellKnownPolicies = wellKnownPolicies
		}
	}

	if !v.PodIdentityAssociations2.IsNull() && !v.PodIdentityAssociations2.IsUnknown() {
		vPodIdentityAssociations := make([]PodIdentityAssociations2Value, 0, len(v.PodIdentityAssociations2.Elements()))
		d = v.PodIdentityAssociations2.ElementsAs(ctx, &vPodIdentityAssociations, false)
		diags = append(diags, d...)
		podIdentityAssociations := make([]*rafay.IAMPodIdentityAssociation, 0, len(vPodIdentityAssociations))
		for _, pia := range vPodIdentityAssociations {
			piaObj, d := pia.Expand(ctx)
			diags = append(diags, d...)
			podIdentityAssociations = append(podIdentityAssociations, piaObj)
		}
		if len(podIdentityAssociations) > 0 {
			addon.PodIdentityAssociations = podIdentityAssociations
		}
	}

	return &addon, diags
}

func (v AttachPolicy3Value) Expand(ctx context.Context) (*rafay.InlineDocument, diag.Diagnostics) {
	var diags diag.Diagnostics
	var policyDoc rafay.InlineDocument
	if v.IsNull() {
		return &rafay.InlineDocument{}, diags
	}
	if !v.Version.IsNull() && !v.Version.IsUnknown() {
		policyDoc.Version = getStringValue(v.Version)
	}
	if !v.Id.IsNull() && !v.Id.IsUnknown() {
		policyDoc.Id = getStringValue(v.Id)
	}
	if !v.Statement2.IsNull() && !v.Statement2.IsUnknown() {
		vStatement := make([]StatementValue, 0, len(v.Statement2.Elements()))
		d := v.Statement2.ElementsAs(ctx, &vStatement, false)
		diags = append(diags, d...)
		statements := make([]rafay.InlineStatement, 0, len(vStatement))
		for _, stmt := range vStatement {
			stmtObj, d := stmt.Expand(ctx)
			diags = append(diags, d...)
			statements = append(statements, stmtObj)
		}
		if len(statements) > 0 {
			policyDoc.Statement = statements
		}
	}

	return &policyDoc, diags
}

func (v WellKnownPolicies3Value) Expand(ctx context.Context) (*rafay.WellKnownPolicies, diag.Diagnostics) {
	var diags diag.Diagnostics
	var wp rafay.WellKnownPolicies

	if v.IsNull() {
		return &rafay.WellKnownPolicies{}, diags
	}

	if !v.ImageBuilder.IsNull() && !v.ImageBuilder.IsUnknown() {
		imageBuilder := getBoolValue(v.ImageBuilder)
		wp.ImageBuilder = &imageBuilder
	}

	if !v.AutoScaler.IsNull() && !v.AutoScaler.IsUnknown() {
		autoScaler := getBoolValue(v.AutoScaler)
		wp.AutoScaler = &autoScaler
	}

	if !v.AwsLoadBalancerController.IsNull() && !v.AwsLoadBalancerController.IsUnknown() {
		awsLoadBalancerController := getBoolValue(v.AwsLoadBalancerController)
		wp.AWSLoadBalancerController = &awsLoadBalancerController
	}

	if !v.ExternalDns.IsNull() && !v.ExternalDns.IsUnknown() {
		externalDns := getBoolValue(v.ExternalDns)
		wp.ExternalDNS = &externalDns
	}

	if !v.CertManager.IsNull() && !v.CertManager.IsUnknown() {
		certManager := getBoolValue(v.CertManager)
		wp.CertManager = &certManager
	}

	if !v.EbsCsiController.IsNull() && !v.EbsCsiController.IsUnknown() {
		ebsCsiController := getBoolValue(v.EbsCsiController)
		wp.EBSCSIController = &ebsCsiController
	}

	if !v.EfsCsiController.IsNull() && !v.EfsCsiController.IsUnknown() {
		efsCsiController := getBoolValue(v.EfsCsiController)
		wp.EFSCSIController = &efsCsiController
	}

	return &wp, diags
}

func (v PodIdentityAssociations2Value) Expand(ctx context.Context) (*rafay.IAMPodIdentityAssociation, diag.Diagnostics) {
	var diags, d diag.Diagnostics
	var pia rafay.IAMPodIdentityAssociation

	if v.IsNull() {
		return &rafay.IAMPodIdentityAssociation{}, diags
	}

	// Map string fields
	if !v.Namespace.IsNull() && !v.Namespace.IsUnknown() {
		pia.Namespace = getStringValue(v.Namespace)
	}

	if !v.ServiceAccountName.IsNull() && !v.ServiceAccountName.IsUnknown() {
		pia.ServiceAccountName = getStringValue(v.ServiceAccountName)
	}

	if !v.RoleArn.IsNull() && !v.RoleArn.IsUnknown() {
		pia.RoleARN = getStringValue(v.RoleArn)
	}

	if !v.CreateServiceAccount.IsNull() && !v.CreateServiceAccount.IsUnknown() {
		pia.CreateServiceAccount = getBoolValue(v.CreateServiceAccount)
	}

	if !v.RoleName.IsNull() && !v.RoleName.IsUnknown() {
		pia.RoleName = getStringValue(v.RoleName)
	}

	if !v.PermissionBoundaryArn.IsNull() && !v.PermissionBoundaryArn.IsUnknown() {
		pia.PermissionsBoundaryARN = getStringValue(v.PermissionBoundaryArn)
	}

	if !v.PermissionPolicyArns.IsNull() && !v.PermissionPolicyArns.IsUnknown() {
		policyArnsList := make([]types.String, 0, len(v.PermissionPolicyArns.Elements()))
		d = v.PermissionPolicyArns.ElementsAs(ctx, &policyArnsList, false)
		diags = append(diags, d...)
		policyArns := make([]string, 0, len(policyArnsList))
		for _, arn := range policyArnsList {
			policyArns = append(policyArns, getStringValue(arn))
		}
		if len(policyArns) > 0 {
			pia.PermissionPolicyARNs = policyArns
		}
	}

	if !v.PermissionPolicy.IsNull() && !v.PermissionPolicy.IsUnknown() {
		var policyDoc map[string]interface{}
		var json2 = jsoniter.ConfigCompatibleWithStandardLibrary
		json2.Unmarshal([]byte(getStringValue(v.PermissionPolicy)), &policyDoc)
		pia.PermissionPolicy = policyDoc
	}

	if !v.Tags.IsNull() && !v.Tags.IsUnknown() {
		tags := make(map[string]string, len(v.Tags.Elements()))
		vTags := make(map[string]types.String, len(v.Tags.Elements()))
		d = v.Tags.ElementsAs(ctx, &vTags, false)
		diags = append(diags, d...)
		for k, val := range vTags {
			tags[k] = getStringValue(val)
		}
		if len(tags) > 0 {
			pia.Tags = tags
		}
	}

	if !v.WellKnownPolicies4.IsNull() && !v.WellKnownPolicies4.IsUnknown() {
		vWellKnownPoliciesList := make([]WellKnownPolicies4Value, 0, len(v.WellKnownPolicies4.Elements()))
		d = v.WellKnownPolicies4.ElementsAs(ctx, &vWellKnownPoliciesList, false)
		diags = append(diags, d...)
		if len(vWellKnownPoliciesList) > 0 {
			wellKnownPolicies, d := vWellKnownPoliciesList[0].Expand(ctx)
			diags = append(diags, d...)
			pia.WellKnownPolicies = wellKnownPolicies
		}
	}

	return &pia, diags
}

func (v WellKnownPolicies4Value) Expand(ctx context.Context) (*rafay.WellKnownPolicies, diag.Diagnostics) {
	var diags diag.Diagnostics
	var wp rafay.WellKnownPolicies

	if v.IsNull() {
		return &rafay.WellKnownPolicies{}, diags
	}

	if !v.ImageBuilder.IsNull() && !v.ImageBuilder.IsUnknown() {
		imageBuilder := getBoolValue(v.ImageBuilder)
		wp.ImageBuilder = &imageBuilder
	}

	if !v.AutoScaler.IsNull() && !v.AutoScaler.IsUnknown() {
		autoScaler := getBoolValue(v.AutoScaler)
		wp.AutoScaler = &autoScaler
	}

	if !v.AwsLoadBalancerController.IsNull() && !v.AwsLoadBalancerController.IsUnknown() {
		awsLoadBalancerController := getBoolValue(v.AwsLoadBalancerController)
		wp.AWSLoadBalancerController = &awsLoadBalancerController
	}

	if !v.ExternalDns.IsNull() && !v.ExternalDns.IsUnknown() {
		externalDns := getBoolValue(v.ExternalDns)
		wp.ExternalDNS = &externalDns
	}

	if !v.CertManager.IsNull() && !v.CertManager.IsUnknown() {
		certManager := getBoolValue(v.CertManager)
		wp.CertManager = &certManager
	}

	if !v.EbsCsiController.IsNull() && !v.EbsCsiController.IsUnknown() {
		ebsCsiController := getBoolValue(v.EbsCsiController)
		wp.EBSCSIController = &ebsCsiController
	}

	if !v.EfsCsiController.IsNull() && !v.EfsCsiController.IsUnknown() {
		efsCsiController := getBoolValue(v.EfsCsiController)
		wp.EFSCSIController = &efsCsiController
	}

	return &wp, diags
}

func (v PrivateClusterValue) Expand(ctx context.Context) (*rafay.PrivateCluster, diag.Diagnostics) {
	var diags diag.Diagnostics
	var pc rafay.PrivateCluster
	if v.IsNull() {
		return &rafay.PrivateCluster{}, diags
	}

	// Map boolean fields to pointers
	if !v.Enabled.IsNull() && !v.Enabled.IsUnknown() {
		enabled := getBoolValue(v.Enabled)
		pc.Enabled = &enabled
	}

	if !v.SkipEndpointCreation.IsNull() && !v.SkipEndpointCreation.IsUnknown() {
		skipEndpointCreation := getBoolValue(v.SkipEndpointCreation)
		pc.SkipEndpointCreation = &skipEndpointCreation
	}

	// Map additional_endpoint_services (list of strings)
	if !v.AdditionalEndpointServices.IsNull() && !v.AdditionalEndpointServices.IsUnknown() {
		endpointServicesList := make([]types.String, 0, len(v.AdditionalEndpointServices.Elements()))
		d := v.AdditionalEndpointServices.ElementsAs(ctx, &endpointServicesList, false)
		diags = append(diags, d...)
		endpointServices := make([]string, 0, len(endpointServicesList))
		for _, service := range endpointServicesList {
			endpointServices = append(endpointServices, getStringValue(service))
		}
		if len(endpointServices) > 0 {
			pc.AdditionalEndpointServices = endpointServices
		}
	}

	return &pc, diags
}

func (v FargateProfilesValue) Expand(ctx context.Context) (*rafay.FargateProfile, diag.Diagnostics) {
	var diags, d diag.Diagnostics
	var fp rafay.FargateProfile

	if v.IsNull() {
		return &rafay.FargateProfile{}, diags
	}

	// Map string fields
	if !v.Name.IsNull() && !v.Name.IsUnknown() {
		fp.Name = getStringValue(v.Name)
	}

	if !v.PodExecutionRoleArn.IsNull() && !v.PodExecutionRoleArn.IsUnknown() {
		fp.PodExecutionRoleARN = getStringValue(v.PodExecutionRoleArn)
	}

	if !v.Status.IsNull() && !v.Status.IsUnknown() {
		fp.Status = getStringValue(v.Status)
	}

	// Map selectors block
	if !v.Selectors.IsNull() && !v.Selectors.IsUnknown() {
		vSelectorsList := make([]SelectorsValue, 0, len(v.Selectors.Elements()))
		d = v.Selectors.ElementsAs(ctx, &vSelectorsList, false)
		diags = append(diags, d...)
		selectors := make([]rafay.FargateProfileSelector, 0, len(vSelectorsList))
		for _, selectorValue := range vSelectorsList {
			selector, d := selectorValue.Expand(ctx)
			diags = append(diags, d...)
			selectors = append(selectors, selector)
		}
		if len(selectors) > 0 {
			fp.Selectors = selectors
		}
	}

	// Map subnets (list of strings)
	if !v.Subnets.IsNull() && !v.Subnets.IsUnknown() {
		subnetsList := make([]types.String, 0, len(v.Subnets.Elements()))
		d = v.Subnets.ElementsAs(ctx, &subnetsList, false)
		diags = append(diags, d...)
		subnets := make([]string, 0, len(subnetsList))
		for _, subnet := range subnetsList {
			subnets = append(subnets, getStringValue(subnet))
		}
		if len(subnets) > 0 {
			fp.Subnets = subnets
		}
	}

	// Map tags (map of strings)
	if !v.Tags.IsNull() && !v.Tags.IsUnknown() {
		tags := make(map[string]string, len(v.Tags.Elements()))
		vTags := make(map[string]types.String, len(v.Tags.Elements()))
		d = v.Tags.ElementsAs(ctx, &vTags, false)
		diags = append(diags, d...)
		for k, val := range vTags {
			tags[k] = getStringValue(val)
		}
		if len(tags) > 0 {
			fp.Tags = tags
		}
	}

	return &fp, diags
}

// SelectorsValue Expand
func (v SelectorsValue) Expand(ctx context.Context) (rafay.FargateProfileSelector, diag.Diagnostics) {
	var diags, d diag.Diagnostics
	var selector rafay.FargateProfileSelector

	if v.IsNull() {
		return rafay.FargateProfileSelector{}, diags
	}

	// Map namespace field
	if !v.Namespace.IsNull() && !v.Namespace.IsUnknown() {
		selector.Namespace = getStringValue(v.Namespace)
	}

	// Map labels (map of strings)
	if !v.Labels.IsNull() && !v.Labels.IsUnknown() {
		labels := make(map[string]string, len(v.Labels.Elements()))
		vLabels := make(map[string]types.String, len(v.Labels.Elements()))
		d = v.Labels.ElementsAs(ctx, &vLabels, false)
		diags = append(diags, d...)
		for k, val := range vLabels {
			labels[k] = getStringValue(val)
		}
		if len(labels) > 0 {
			selector.Labels = labels
		}
	}

	return selector, diags
}

func (v CloudWatchValue) Expand(ctx context.Context) (*rafay.EKSClusterCloudWatch, diag.Diagnostics) {
	var diags, d diag.Diagnostics
	var cw rafay.EKSClusterCloudWatch

	if v.IsNull() {
		return &rafay.EKSClusterCloudWatch{}, diags
	}

	// Map cluster_logging block
	if !v.ClusterLogging.IsNull() && !v.ClusterLogging.IsUnknown() {
		vClusterLoggingList := make([]ClusterLoggingValue, 0, len(v.ClusterLogging.Elements()))
		d = v.ClusterLogging.ElementsAs(ctx, &vClusterLoggingList, false)
		diags = append(diags, d...)
		if len(vClusterLoggingList) > 0 {
			clusterLogging, d := vClusterLoggingList[0].Expand(ctx)
			diags = append(diags, d...)
			cw.ClusterLogging = clusterLogging
		}
	}

	return &cw, diags
}

// CloudLoggingValue Expand
func (v ClusterLoggingValue) Expand(ctx context.Context) (*rafay.EKSClusterCloudWatchLogging, diag.Diagnostics) {
	var diags, d diag.Diagnostics
	var logging rafay.EKSClusterCloudWatchLogging

	if v.IsNull() {
		return &rafay.EKSClusterCloudWatchLogging{}, diags
	}

	// Map enable_types (list of strings)
	if !v.EnableTypes.IsNull() && !v.EnableTypes.IsUnknown() {
		enableTypesList := make([]types.String, 0, len(v.EnableTypes.Elements()))
		d = v.EnableTypes.ElementsAs(ctx, &enableTypesList, false)
		diags = append(diags, d...)
		enableTypes := make([]string, 0, len(enableTypesList))
		for _, enableType := range enableTypesList {
			enableTypes = append(enableTypes, getStringValue(enableType))
		}
		if len(enableTypes) > 0 {
			logging.EnableTypes = enableTypes
		}
	}

	// Map log_retention_in_days (int64 field)
	if !v.LogRetentionInDays.IsNull() && !v.LogRetentionInDays.IsUnknown() {
		logging.LogRetentionInDays = int(getInt64Value(v.LogRetentionInDays))
	}

	return &logging, diags
}

func (v SecretsEncryptionValue) Expand(ctx context.Context) (*rafay.SecretsEncryption, diag.Diagnostics) {
	var diags diag.Diagnostics
	var se rafay.SecretsEncryption

	if v.IsNull() {
		return &rafay.SecretsEncryption{}, diags
	}

	// Map key_arn field
	if !v.KeyArn.IsNull() && !v.KeyArn.IsUnknown() {
		se.KeyARN = getStringValue(v.KeyArn)
	}

	// Map encrypt_existing_secrets field (bool to pointer)
	if !v.EncryptExistingSecrets.IsNull() && !v.EncryptExistingSecrets.IsUnknown() {
		encryptExisting := getBoolValue(v.EncryptExistingSecrets)
		se.EncryptExistingSecrets = &encryptExisting
	}

	return &se, diags
}

func (v IdentityMappingsValue) Expand(ctx context.Context) (*rafay.EKSClusterIdentityMappings, diag.Diagnostics) {
	var diags, d diag.Diagnostics
	var im rafay.EKSClusterIdentityMappings

	if v.IsNull() {
		return &rafay.EKSClusterIdentityMappings{}, diags
	}

	// Map arns block
	if !v.Arns.IsNull() && !v.Arns.IsUnknown() {
		vArnsList := make([]ArnsValue, 0, len(v.Arns.Elements()))
		d = v.Arns.ElementsAs(ctx, &vArnsList, false)
		diags = append(diags, d...)
		arns := make([]*rafay.IdentityMappingARN, 0, len(vArnsList))
		for _, arnValue := range vArnsList {
			arn, d := arnValue.Expand(ctx)
			diags = append(diags, d...)
			arns = append(arns, arn)
		}
		if len(arns) > 0 {
			im.Arns = arns
		}
	}

	// Map accounts (list of strings)
	if !v.Accounts.IsNull() && !v.Accounts.IsUnknown() {
		accountsList := make([]types.String, 0, len(v.Accounts.Elements()))
		d = v.Accounts.ElementsAs(ctx, &accountsList, false)
		diags = append(diags, d...)
		accounts := make([]string, 0, len(accountsList))
		for _, account := range accountsList {
			accounts = append(accounts, getStringValue(account))
		}
		if len(accounts) > 0 {
			im.Accounts = accounts
		}
	}

	return &im, diags
}

// ArnsValue Expand
func (v ArnsValue) Expand(ctx context.Context) (*rafay.IdentityMappingARN, diag.Diagnostics) {
	var diags, d diag.Diagnostics
	var arn rafay.IdentityMappingARN

	if v.IsNull() {
		return &rafay.IdentityMappingARN{}, diags
	}

	// Map arn field
	if !v.Arn.IsNull() && !v.Arn.IsUnknown() {
		arn.Arn = getStringValue(v.Arn)
	}

	// Map username field
	if !v.Username.IsNull() && !v.Username.IsUnknown() {
		arn.Username = getStringValue(v.Username)
	}

	// Map group (list of strings)
	if !v.Group.IsNull() && !v.Group.IsUnknown() {
		groupList := make([]types.String, 0, len(v.Group.Elements()))
		d = v.Group.ElementsAs(ctx, &groupList, false)
		diags = append(diags, d...)
		group := make([]string, 0, len(groupList))
		for _, groupItem := range groupList {
			group = append(group, getStringValue(groupItem))
		}
		if len(group) > 0 {
			arn.Group = group
		}
	}

	return &arn, diags
}

func (v AccessConfigValue) Expand(ctx context.Context) (*rafay.EKSClusterAccess, diag.Diagnostics) {
	var diags, d diag.Diagnostics
	var ac rafay.EKSClusterAccess

	if v.IsNull() {
		return &rafay.EKSClusterAccess{}, diags
	}

	// Map bootstrap_cluster_creator_admin_permissions field
	if !v.BootstrapClusterCreatorAdminPermissions.IsNull() && !v.BootstrapClusterCreatorAdminPermissions.IsUnknown() {
		ac.BootstrapClusterCreatorAdminPermissions = getBoolValue(v.BootstrapClusterCreatorAdminPermissions)
	}

	// Map authentication_mode field
	if !v.AuthenticationMode.IsNull() && !v.AuthenticationMode.IsUnknown() {
		ac.AuthenticationMode = getStringValue(v.AuthenticationMode)
	}

	// Map access_entries block
	if !v.AccessEntries.IsNull() && !v.AccessEntries.IsUnknown() {
		vAccessEntriesList := make([]AccessEntriesValue, 0, len(v.AccessEntries.Elements()))
		d = v.AccessEntries.ElementsAs(ctx, &vAccessEntriesList, false)
		diags = append(diags, d...)
		accessEntries := make([]*rafay.EKSAccessEntry, 0, len(vAccessEntriesList))
		for _, accessEntryValue := range vAccessEntriesList {
			accessEntry, d := accessEntryValue.Expand(ctx)
			diags = append(diags, d...)
			accessEntries = append(accessEntries, accessEntry)
		}
		if len(accessEntries) > 0 {
			ac.AccessEntries = accessEntries
		}
	}

	return &ac, diags
}

// AccessEntriesValue Expand
func (v AccessEntriesValue) Expand(ctx context.Context) (*rafay.EKSAccessEntry, diag.Diagnostics) {
	var diags, d diag.Diagnostics
	var ae rafay.EKSAccessEntry

	if v.IsNull() {
		return &rafay.EKSAccessEntry{}, diags
	}

	// Map principal_arn field
	if !v.PrincipalArn.IsNull() && !v.PrincipalArn.IsUnknown() {
		ae.PrincipalARN = getStringValue(v.PrincipalArn)
	}

	// Map type field
	if !v.AccessEntriesType.IsNull() && !v.AccessEntriesType.IsUnknown() {
		ae.Type = getStringValue(v.AccessEntriesType)
	}

	// Map kubernetes_username field
	if !v.KubernetesUsername.IsNull() && !v.KubernetesUsername.IsUnknown() {
		ae.KubernetesUsername = getStringValue(v.KubernetesUsername)
	}

	// Map kubernetes_groups (list of strings)
	if !v.KubernetesGroups.IsNull() && !v.KubernetesGroups.IsUnknown() {
		kubernetesGroupsList := make([]types.String, 0, len(v.KubernetesGroups.Elements()))
		d = v.KubernetesGroups.ElementsAs(ctx, &kubernetesGroupsList, false)
		diags = append(diags, d...)
		kubernetesGroups := make([]string, 0, len(kubernetesGroupsList))
		for _, group := range kubernetesGroupsList {
			kubernetesGroups = append(kubernetesGroups, getStringValue(group))
		}
		if len(kubernetesGroups) > 0 {
			ae.KubernetesGroups = kubernetesGroups
		}
	}

	// Map tags (map of strings)
	if !v.Tags.IsNull() && !v.Tags.IsUnknown() {
		tags := make(map[string]string, len(v.Tags.Elements()))
		vTags := make(map[string]types.String, len(v.Tags.Elements()))
		d = v.Tags.ElementsAs(ctx, &vTags, false)
		diags = append(diags, d...)
		for k, val := range vTags {
			tags[k] = getStringValue(val)
		}
		if len(tags) > 0 {
			ae.Tags = tags
		}
	}

	// Map access_policies block
	if !v.AccessPolicies.IsNull() && !v.AccessPolicies.IsUnknown() {
		vAccessPoliciesList := make([]AccessPoliciesValue, 0, len(v.AccessPolicies.Elements()))
		d = v.AccessPolicies.ElementsAs(ctx, &vAccessPoliciesList, false)
		diags = append(diags, d...)
		accessPolicies := make([]*rafay.EKSAccessPolicy, 0, len(vAccessPoliciesList))
		for _, accessPolicyValue := range vAccessPoliciesList {
			accessPolicy, d := accessPolicyValue.Expand(ctx)
			diags = append(diags, d...)
			accessPolicies = append(accessPolicies, accessPolicy)
		}
		if len(accessPolicies) > 0 {
			ae.AccessPolicies = accessPolicies
		}
	}

	return &ae, diags
}

// AccessPoliciesValue Expand
func (v AccessPoliciesValue) Expand(ctx context.Context) (*rafay.EKSAccessPolicy, diag.Diagnostics) {
	var diags, d diag.Diagnostics
	var ap rafay.EKSAccessPolicy

	if v.IsNull() {
		return &rafay.EKSAccessPolicy{}, diags
	}

	// Map policy_arn field
	if !v.PolicyArn.IsNull() && !v.PolicyArn.IsUnknown() {
		ap.PolicyARN = getStringValue(v.PolicyArn)
	}

	// Map access_scope block
	if !v.AccessScope.IsNull() && !v.AccessScope.IsUnknown() {
		vAccessScopeList := make([]AccessScopeValue, 0, len(v.AccessScope.Elements()))
		d = v.AccessScope.ElementsAs(ctx, &vAccessScopeList, false)
		diags = append(diags, d...)
		if len(vAccessScopeList) > 0 {
			accessScope, d := vAccessScopeList[0].Expand(ctx)
			diags = append(diags, d...)
			ap.AccessScope = accessScope
		}
	}

	return &ap, diags
}

// AccessScopeValue Expand
func (v AccessScopeValue) Expand(ctx context.Context) (*rafay.EKSAccessScope, diag.Diagnostics) {
	var diags, d diag.Diagnostics
	var as rafay.EKSAccessScope

	if v.IsNull() {
		return &rafay.EKSAccessScope{}, diags
	}

	// Map type field
	if !v.AccessScopeType.IsNull() && !v.AccessScopeType.IsUnknown() {
		as.Type = getStringValue(v.AccessScopeType)
	}

	// Map namespaces (list of strings)
	if !v.Namespaces.IsNull() && !v.Namespaces.IsUnknown() {
		namespacesList := make([]types.String, 0, len(v.Namespaces.Elements()))
		d = v.Namespaces.ElementsAs(ctx, &namespacesList, false)
		diags = append(diags, d...)
		namespaces := make([]string, 0, len(namespacesList))
		for _, namespace := range namespacesList {
			namespaces = append(namespaces, getStringValue(namespace))
		}
		if len(namespaces) > 0 {
			as.Namespaces = namespaces
		}
	}

	return &as, diags
}

func (v AddonsConfigValue) Expand(ctx context.Context) (*rafay.EKSAddonsConfig, diag.Diagnostics) {
	var diags diag.Diagnostics
	var acfg rafay.EKSAddonsConfig

	if v.IsNull() {
		return &rafay.EKSAddonsConfig{}, diags
	}

	// Map auto_apply_pod_identity_associations field
	if !v.AutoApplyPodIdentityAssociations.IsNull() && !v.AutoApplyPodIdentityAssociations.IsUnknown() {
		acfg.AutoApplyPodIdentityAssociations = getBoolValue(v.AutoApplyPodIdentityAssociations)
	}

	// Map disable_ebs_csi_driver field
	if !v.DisableEbsCsiDriver.IsNull() && !v.DisableEbsCsiDriver.IsUnknown() {
		acfg.DisableEBSCSIDriver = getBoolValue(v.DisableEbsCsiDriver)
	}

	return &acfg, diags
}

func (v AutoModeConfigValue) Expand(ctx context.Context) (*rafay.EKSAutoModeConfig, diag.Diagnostics) {
	var diags, d diag.Diagnostics
	var amc rafay.EKSAutoModeConfig

	if v.IsNull() {
		return &rafay.EKSAutoModeConfig{}, diags
	}

	// Map enabled field
	if !v.Enabled.IsNull() && !v.Enabled.IsUnknown() {
		amc.Enabled = getBoolValue(v.Enabled)
	}

	// Map node_role_arn field
	if !v.NodeRoleArn.IsNull() && !v.NodeRoleArn.IsUnknown() {
		amc.NodeRoleARN = getStringValue(v.NodeRoleArn)
	}

	// Map node_pools (list of strings)
	if !v.NodePools.IsNull() && !v.NodePools.IsUnknown() {
		nodePoolsList := make([]types.String, 0, len(v.NodePools.Elements()))
		d = v.NodePools.ElementsAs(ctx, &nodePoolsList, false)
		diags = append(diags, d...)
		nodePools := make([]string, 0, len(nodePoolsList))
		for _, nodePool := range nodePoolsList {
			nodePools = append(nodePools, getStringValue(nodePool))
		}
		if len(nodePools) > 0 {
			amc.NodePools = nodePools
		}
	}

	return &amc, diags
}
