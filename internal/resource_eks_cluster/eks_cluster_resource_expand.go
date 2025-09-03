package resource_eks_cluster

import (
	"context"

	"github.com/RafaySystems/terraform-provider-rafay/rafay"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
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
	var cluster *rafay.EKSCluster

	if v.Cluster.IsNull() {
		return &rafay.EKSCluster{}, diags
	}

	vClusterList := make([]ClusterValue, 0, len(v.Cluster.Elements()))
	diags = v.Cluster.ElementsAs(ctx, &vClusterList, false)
	vCluster := vClusterList[0]
	cluster, d = vCluster.Expand(ctx)
	diags = append(diags, d...)

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
	vClusterConfig := vClusterConfigList[0]
	clusterConfig, d = vClusterConfig.Expand(ctx)
	diags = append(diags, d...)

	return clusterConfig, diags
}

// Cluster Expand

func (v ClusterValue) Expand(ctx context.Context) (*rafay.EKSCluster, diag.Diagnostics) {
	var diags, d diag.Diagnostics
	var cluster rafay.EKSCluster

	if v.IsNull() {
		return &rafay.EKSCluster{}, diags
	}

	cluster.Kind = getStringValue(v.Kind)

	vMetadataList := make([]MetadataValue, 0, len(v.Metadata.Elements()))
	diags = v.Metadata.ElementsAs(ctx, &vMetadataList, false)
	vMetadata := vMetadataList[0]
	md, d := vMetadata.Expand(ctx)
	diags = append(diags, d...)
	cluster.Metadata = md

	vSpecList := make([]SpecValue, 0, len(v.Spec.Elements()))
	diags = v.Spec.ElementsAs(ctx, &vSpecList, false)
	vSpec := vSpecList[0]
	spec, d := vSpec.Expand(ctx)
	diags = append(diags, d...)
	cluster.Spec = spec

	return &cluster, diags
}

func (v MetadataValue) Expand(ctx context.Context) (*rafay.EKSClusterMetadata, diag.Diagnostics) {
	var diags, d diag.Diagnostics
	var metadata rafay.EKSClusterMetadata

	if v.IsNull() {
		return &rafay.EKSClusterMetadata{}, diags
	}

	metadata.Name = getStringValue(v.Name)
	metadata.Project = getStringValue(v.Project)

	labels := make(map[string]string, len(v.Labels.Elements()))
	vLabels := make(map[string]types.String, len(v.Labels.Elements()))
	d = v.Labels.ElementsAs(ctx, &vLabels, false)
	diags = append(diags, d...)
	for k, val := range vLabels {
		labels[k] = getStringValue(val)
	}
	metadata.Labels = labels

	return &metadata, diags
}

func (v SpecValue) Expand(ctx context.Context) (*rafay.EKSSpec, diag.Diagnostics) {
	var diags, d diag.Diagnostics
	var spec rafay.EKSSpec

	if v.IsNull() {
		return &rafay.EKSSpec{}, diags
	}

	spec.Blueprint = getStringValue(v.Blueprint)
	spec.BlueprintVersion = getStringValue(v.BlueprintVersion)
	spec.CloudProvider = getStringValue(v.CloudProvider)
	spec.CniProvider = getStringValue(v.CniProvider)
	spec.Type = getStringValue(v.SpecType)
	spec.CrossAccountRoleArn = getStringValue(v.CrossAccountRoleArn)

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

	var vSCP SystemComponentsPlacementValue
	d = v.SystemComponentsPlacement.As(ctx, &vSCP, basetypes.ObjectAsOptions{})
	diags = append(diags, d...)
	spec.SystemComponentsPlacement, d = vSCP.Expand(ctx)
	diags = append(diags, d...)

	var sharing SharingValue
	d = v.Sharing.As(ctx, &sharing, basetypes.ObjectAsOptions{})
	diags = append(diags, d...)
	spec.Sharing, d = sharing.Expand(ctx)
	diags = append(diags, d...)

	return &spec, diags
}

func (v SharingValue) Expand(ctx context.Context) (*rafay.V1ClusterSharing, diag.Diagnostics) {
	var diags, d diag.Diagnostics
	var sharing rafay.V1ClusterSharing

	if v.IsNull() {
		return &rafay.V1ClusterSharing{}, diags
	}

	b := getBoolValue(v.Enabled)
	sharing.Enabled = &b

	vProjectsList := make([]ProjectsValue, 0, len(v.Projects.Elements()))
	d = v.Projects.ElementsAs(ctx, &vProjectsList, false)
	diags = append(diags, d...)
	prjs := make([]*rafay.V1ClusterSharingProject, 0, len(vProjectsList))
	for _, prj := range vProjectsList {
		p, d := prj.Expand(ctx)
		diags = append(diags, d...)
		prjs = append(prjs, p)
	}
	sharing.Projects = prjs

	return &sharing, diags
}

func (v ProjectsValue) Expand(ctx context.Context) (*rafay.V1ClusterSharingProject, diag.Diagnostics) {
	var diags diag.Diagnostics
	var project rafay.V1ClusterSharingProject

	if v.IsNull() {
		return &rafay.V1ClusterSharingProject{}, diags
	}

	project.Name = getStringValue(v.Name)

	return &project, diags
}

func (v SystemComponentsPlacementValue) Expand(ctx context.Context) (*rafay.SystemComponentsPlacement, diag.Diagnostics) {
	var diags, d diag.Diagnostics
	var scp rafay.SystemComponentsPlacement

	if v.IsNull() {
		return &rafay.SystemComponentsPlacement{}, diags
	}

	nsel := make(map[string]string, len(v.NodeSelector.Elements()))
	vnsel := make(map[string]types.String, len(v.NodeSelector.Elements()))
	d = v.NodeSelector.ElementsAs(ctx, &vnsel, false)
	diags = append(diags, d...)
	for k, val := range vnsel {
		nsel[k] = getStringValue(val)
	}
	scp.NodeSelector = nsel

	vTolerationList := make([]TolerationsValue, 0, len(v.Tolerations.Elements()))
	d = v.Tolerations.ElementsAs(ctx, &vTolerationList, false)
	diags = append(diags, d...)
	tols := make([]*rafay.Tolerations, 0, len(vTolerationList))
	for _, tl := range vTolerationList {
		t, d := tl.Expand(ctx)
		diags = append(diags, d...)
		tols = append(tols, t)
	}
	scp.Tolerations = tols

	vDaemonsetOverrideList := make([]DaemonsetOverrideValue, 0, len(v.DaemonsetOverride.Elements()))
	d = v.DaemonsetOverride.ElementsAs(ctx, &vDaemonsetOverrideList, false)
	diags = append(diags, d...)
	do, d := vDaemonsetOverrideList[0].Expand(ctx)
	diags = append(diags, d...)
	scp.DaemonsetOverride = do

	return &scp, diags
}

func (v DaemonsetOverrideValue) Expand(ctx context.Context) (*rafay.DaemonsetOverride, diag.Diagnostics) {
	var diags, d diag.Diagnostics
	var do rafay.DaemonsetOverride

	if v.IsNull() {
		return &rafay.DaemonsetOverride{}, diags
	}

	nse := getBoolValue(v.NodeSelectionEnabled)
	do.NodeSelectionEnabled = &nse

	vTolerationList := make([]Tolerations2Value, 0, len(v.Tolerations2.Elements()))
	d = v.Tolerations2.ElementsAs(ctx, &vTolerationList, false)
	diags = append(diags, d...)
	tols := make([]*rafay.Tolerations, 0, len(vTolerationList))
	for _, tl := range vTolerationList {
		t, d := tl.Expand(ctx)
		diags = append(diags, d...)
		tols = append(tols, t)
	}
	do.Tolerations = tols

	return &do, diags
}

func (v *Tolerations2Value) Expand(ctx context.Context) (*rafay.Tolerations, diag.Diagnostics) {
	var diags diag.Diagnostics
	var tol rafay.Tolerations

	if v.IsNull() {
		return &rafay.Tolerations{}, diags
	}

	tol.Key = getStringValue(v.Key)
	tol.Operator = getStringValue(v.Operator)
	tol.Value = getStringValue(v.Value)
	tol.Effect = getStringValue(v.Effect)

	d := int(getInt64Value(v.TolerationSeconds))
	tol.TolerationSeconds = &d

	return &tol, diags
}

func (v *TolerationsValue) Expand(ctx context.Context) (*rafay.Tolerations, diag.Diagnostics) {
	var diags diag.Diagnostics
	var tol rafay.Tolerations

	if v.IsNull() {
		return &rafay.Tolerations{}, diags
	}

	tol.Key = getStringValue(v.Key)
	tol.Operator = getStringValue(v.Operator)
	tol.Value = getStringValue(v.Value)
	tol.Effect = getStringValue(v.Effect)

	d := int(getInt64Value(v.TolerationSeconds))
	tol.TolerationSeconds = &d

	return &tol, diags
}

// ClusterConfig Expand

func (v ClusterConfigValue) Expand(ctx context.Context) (*rafay.EKSClusterConfig, diag.Diagnostics) {
	var diags, d diag.Diagnostics
	var clusterConfig rafay.EKSClusterConfig

	if v.IsNull() {
		return &rafay.EKSClusterConfig{}, diags
	}

	clusterConfig.APIVersion = getStringValue(v.Apiversion)
	clusterConfig.Kind = getStringValue(v.Kind)

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
// TODO: Implement Expand functions for: VpcValue, AddonsValue, PrivateClusterValue, ManagedNodeGroupsValue, FargateProfilesValue, CloudWatchValue, SecretsEncryptionValue, IdentityMappingsValue, AccessConfigValue, AddonsConfigValue, AutoModeConfigValue, NodeGroupsMapValue

// Stub Expand methods for all referenced block types
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

func (v NodeGroupsValue) Expand(ctx context.Context) (*rafay.NodeGroup, diag.Diagnostics) {
	var diags, d diag.Diagnostics
	var ng rafay.NodeGroup
	if v.IsNull() {
		return &rafay.NodeGroup{}, diags
	}

	if !v.Name.IsNull() && !v.Name.IsUnknown() {
		ng.Name = getStringValue(v.Name)
	}
	if !v.Version.IsNull() && !v.Version.IsUnknown() {
		ng.Version = getStringValue(v.Version)
	}
	if !v.AmiFamily.IsNull() && !v.AmiFamily.IsUnknown() {
		ng.AMIFamily = getStringValue(v.AmiFamily)
	}
	if !v.InstanceType.IsNull() && !v.InstanceType.IsUnknown() {
		ng.InstanceType = getStringValue(v.InstanceType)
	}
	if !v.MaxPodsPerNode.IsNull() && !v.MaxPodsPerNode.IsUnknown() {
		ng.MaxPodsPerNode = int(getInt64Value(v.MaxPodsPerNode))
	}
	if !v.VolumeType.IsNull() && !v.VolumeType.IsUnknown() {
		ng.VolumeType = getStringValue(v.VolumeType)
	}
	if !v.VolumeSize.IsNull() && !v.VolumeSize.IsUnknown() {
		volSize := int(getInt64Value(v.VolumeSize))
		ng.VolumeSize = &volSize
	}
	if !v.VolumeIops.IsNull() && !v.VolumeIops.IsUnknown() {
		volIops := int(getInt64Value(v.VolumeIops))
		ng.VolumeIOPS = &volIops
	}
	if !v.VolumeThroughput.IsNull() && !v.VolumeThroughput.IsUnknown() {
		volThroughput := int(getInt64Value(v.VolumeThroughput))
		ng.VolumeThroughput = &volThroughput
	}
	if !v.DesiredCapacity.IsNull() && !v.DesiredCapacity.IsUnknown() {
		desiredCap := int(getInt64Value(v.DesiredCapacity))
		ng.DesiredCapacity = &desiredCap
	}
	if !v.MaxSize.IsNull() && !v.MaxSize.IsUnknown() {
		maxSize := int(getInt64Value(v.MaxSize))
		ng.MaxSize = &maxSize
	}
	if !v.MinSize.IsNull() && !v.MinSize.IsUnknown() {
		minSize := int(getInt64Value(v.MinSize))
		ng.MinSize = &minSize
	}
	if !v.PrivateNetworking.IsNull() && !v.PrivateNetworking.IsUnknown() {
		privNet := getBoolValue(v.PrivateNetworking)
		ng.PrivateNetworking = &privNet
	}
	if !v.DisableImdsv1.IsNull() && !v.DisableImdsv1.IsUnknown() {
		disImdsv1 := getBoolValue(v.DisableImdsv1)
		ng.DisableIMDSv1 = &disImdsv1
	}
	if !v.DisablePodsImds.IsNull() && !v.DisablePodsImds.IsUnknown() {
		disPodImds := getBoolValue(v.DisablePodsImds)
		ng.DisablePodIMDS = &disPodImds
	}
	if !v.EfaEnabled.IsNull() && !v.EfaEnabled.IsUnknown() {
		efaEnabled := getBoolValue(v.EfaEnabled)
		ng.EFAEnabled = &efaEnabled
	}
	if !v.Labels2.IsNull() && !v.Labels2.IsUnknown() {
		ng.Labels = make(map[string]string, len(v.Labels2.Elements()))
		vLabels := make(map[string]types.String, len(v.Labels2.Elements()))
		d = v.Labels2.ElementsAs(ctx, &vLabels, false)
		diags = append(diags, d...)
		for k, val := range vLabels {
			ng.Labels[k] = getStringValue(val)
		}
	}
	if !v.Tags2.IsNull() && !v.Tags2.IsUnknown() {
		ng.Tags = make(map[string]string, len(v.Tags2.Elements()))
		vTags := make(map[string]types.String, len(v.Tags2.Elements()))
		d = v.Tags2.ElementsAs(ctx, &vTags, false)
		diags = append(diags, d...)
		for k, val := range vTags {
			ng.Tags[k] = getStringValue(val)
		}
	}
	if !v.ClusterDns.IsNull() && !v.ClusterDns.IsUnknown() {
		ng.ClusterDNS = getStringValue(v.ClusterDns)
	}
	if !v.TargetGroupArns.IsNull() && !v.TargetGroupArns.IsUnknown() {
		tgArnsList := make([]types.String, 0, len(v.TargetGroupArns.Elements()))
		d = v.TargetGroupArns.ElementsAs(ctx, &tgArnsList, false)
		diags = append(diags, d...)
		tgArns := make([]string, 0, len(tgArnsList))
		for _, tg := range tgArnsList {
			tgArns = append(tgArns, getStringValue(tg))
		}
		ng.TargetGroupARNs = tgArns
	}
	if !v.ClassicLoadBalancerNames.IsNull() && !v.ClassicLoadBalancerNames.IsUnknown() {
		lbNamesList := make([]types.String, 0, len(v.ClassicLoadBalancerNames.Elements()))
		d = v.ClassicLoadBalancerNames.ElementsAs(ctx, &lbNamesList, false)
		diags = append(diags, d...)
		lbNames := make([]string, 0, len(lbNamesList))
		for _, lb := range lbNamesList {
			lbNames = append(lbNames, getStringValue(lb))
		}
		ng.ClassicLoadBalancerNames = lbNames
	}
	if !v.CpuCredits.IsNull() && !v.CpuCredits.IsUnknown() {
		ng.CPUCredits = getStringValue(v.CpuCredits)
	}
	if !v.EnableDetailedMonitoring.IsNull() && !v.EnableDetailedMonitoring.IsUnknown() {
		enableDM := getBoolValue(v.EnableDetailedMonitoring)
		ng.EnableDetailedMonitoring = &enableDM
	}
	if !v.AvailabilityZones2.IsNull() && !v.AvailabilityZones2.IsUnknown() {
		azList := make([]types.String, 0, len(v.AvailabilityZones2.Elements()))
		d = v.AvailabilityZones2.ElementsAs(ctx, &azList, false)
		diags = append(diags, d...)
		azs := make([]string, 0, len(azList))
		for _, az := range azList {
			azs = append(azs, getStringValue(az))
		}
		ng.AvailabilityZones = azs
	}
	if !v.Subnets.IsNull() && !v.Subnets.IsUnknown() {
		subnetsList := make([]types.String, 0, len(v.Subnets.Elements()))
		d = v.Subnets.ElementsAs(ctx, &subnetsList, false)
		diags = append(diags, d...)
		subnets := make([]string, 0, len(subnetsList))
		for _, s := range subnetsList {
			subnets = append(subnets, getStringValue(s))
		}
		ng.Subnets = subnets
	}
	if !v.InstancePrefix.IsNull() && !v.InstancePrefix.IsUnknown() {
		ng.InstancePrefix = getStringValue(v.InstancePrefix)
	}
	if !v.InstanceName.IsNull() && !v.InstanceName.IsUnknown() {
		ng.InstanceName = getStringValue(v.InstanceName)
	}
	if !v.Ami.IsNull() && !v.Ami.IsUnknown() {
		ng.AMI = getStringValue(v.Ami)
	}
	if !v.AsgSuspendProcesses.IsNull() && !v.AsgSuspendProcesses.IsUnknown() {
		asgSuspendList := make([]types.String, 0, len(v.AsgSuspendProcesses.Elements()))
		d = v.AsgSuspendProcesses.ElementsAs(ctx, &asgSuspendList, false)
		asgSuspend := make([]string, 0, len(asgSuspendList))
		for _, p := range asgSuspendList {
			asgSuspend = append(asgSuspend, getStringValue(p))
		}
		ng.ASGSuspendProcesses = asgSuspend
	}
	if !v.EbsOptimized.IsNull() && !v.EbsOptimized.IsUnknown() {
		ebsOpt := getBoolValue(v.EbsOptimized)
		ng.EBSOptimized = &ebsOpt
	}
	if !v.VolumeName.IsNull() && !v.VolumeName.IsUnknown() {
		ng.VolumeName = getStringValue(v.VolumeName)
	}
	if !v.VolumeEncrypted.IsNull() && !v.VolumeEncrypted.IsUnknown() {
		volEncrypted := getBoolValue(v.VolumeEncrypted)
		ng.VolumeEncrypted = &volEncrypted
	}
	if !v.VolumeKmsKeyId.IsNull() && !v.VolumeKmsKeyId.IsUnknown() {
		ng.VolumeKmsKeyID = getStringValue(v.VolumeKmsKeyId)
	}
	if !v.PreBootstrapCommands.IsNull() && !v.PreBootstrapCommands.IsUnknown() {
		preBootstrapList := make([]types.String, 0, len(v.PreBootstrapCommands.Elements()))
		d = v.PreBootstrapCommands.ElementsAs(ctx, &preBootstrapList, false)
		preBootstrap := make([]string, 0, len(preBootstrapList))
		for _, cmd := range preBootstrapList {
			preBootstrap = append(preBootstrap, getStringValue(cmd))
		}
		ng.PreBootstrapCommands = preBootstrap
	}
	if !v.OverrideBootstrapCommand.IsNull() && !v.OverrideBootstrapCommand.IsUnknown() {
		ng.OverrideBootstrapCommand = getStringValue(v.OverrideBootstrapCommand)
	}

	// node group's block starts here
	if !v.Iam.IsNull() && !v.Iam.IsUnknown() {
		vIamList := make([]IamValue, 0, len(v.Iam.Elements()))
		d = v.Iam.ElementsAs(ctx, &vIamList, false)
		diags = append(diags, d...)
		if len(vIamList) > 0 {
			ng.IAM, d = vIamList[0].Expand(ctx)
			diags = append(diags, d...)
		}
	}
	if !v.Ssh.IsNull() && !v.Ssh.IsUnknown() {
		vSshList := make([]SshValue, 0, len(v.Ssh.Elements()))
		d = v.Ssh.ElementsAs(ctx, &vSshList, false)
		diags = append(diags, d...)
		if len(vSshList) > 0 {
			ng.SSH, d = vSshList[0].Expand(ctx)
			diags = append(diags, d...)
		}
	}

	if !v.Placement.IsNull() && !v.Placement.IsUnknown() {
		vPlacementList := make([]PlacementValue, 0, len(v.Placement.Elements()))
		d = v.Placement.ElementsAs(ctx, &vPlacementList, false)
		diags = append(diags, d...)
		if len(vPlacementList) > 0 {
			ng.Placement, d = vPlacementList[0].Expand(ctx)
			diags = append(diags, d...)
		}
	}

	if !v.InstanceSelector.IsNull() && !v.InstanceSelector.IsUnknown() {
		vInstanceSelectorList := make([]InstanceSelectorValue, 0, len(v.InstanceSelector.Elements()))
		d = v.InstanceSelector.ElementsAs(ctx, &vInstanceSelectorList, false)
		diags = append(diags, d...)
		if len(vInstanceSelectorList) > 0 {
			ng.InstanceSelector, d = vInstanceSelectorList[0].Expand(ctx)
			diags = append(diags, d...)
		}
	}

	if !v.BottleRocket.IsNull() && !v.BottleRocket.IsUnknown() {
		vBottleRocketList := make([]BottleRocketValue, 0, len(v.BottleRocket.Elements()))
		d = v.BottleRocket.ElementsAs(ctx, &vBottleRocketList, false)
		diags = append(diags, d...)
		if len(vBottleRocketList) > 0 {
			ng.Bottlerocket, d = vBottleRocketList[0].Expand(ctx)
			diags = append(diags, d...)
		}
	}
	if !v.InstancesDistribution.IsNull() && !v.InstancesDistribution.IsUnknown() {
		vInstancesDistributionList := make([]InstancesDistributionValue, 0, len(v.InstancesDistribution.Elements()))
		d = v.InstancesDistribution.ElementsAs(ctx, &vInstancesDistributionList, false)
		diags = append(diags, d...)
		if len(vInstancesDistributionList) > 0 {
			ng.InstancesDistribution, d = vInstancesDistributionList[0].Expand(ctx)
			diags = append(diags, d...)
		}
	}

	if !v.AsgMetricsCollection.IsNull() && !v.AsgMetricsCollection.IsUnknown() {
		vASGMetricsCollectionList := make([]AsgMetricsCollectionValue, 0, len(v.AsgMetricsCollection.Elements()))
		d = v.AsgMetricsCollection.ElementsAs(ctx, &vASGMetricsCollectionList, false)
		diags = append(diags, d...)
		if len(vASGMetricsCollectionList) > 0 {
			for _, m := range vASGMetricsCollectionList {
				metric, d := m.Expand(ctx)
				diags = append(diags, d...)
				ng.ASGMetricsCollection = append(ng.ASGMetricsCollection, metric)
			}
		}
	}

	if !v.Taints.IsNull() && !v.Taints.IsUnknown() {
		vTaintsList := make([]TaintsValue, 0, len(v.Taints.Elements()))
		d = v.Taints.ElementsAs(ctx, &vTaintsList, false)
		diags = append(diags, d...)
		taints := make([]rafay.NodeGroupTaint, 0, len(vTaintsList))
		for _, t := range vTaintsList {
			taint, d := t.Expand(ctx)
			diags = append(diags, d...)
			taints = append(taints, taint)
		}
		ng.Taints = taints
	}

	if !v.UpdateConfig.IsNull() && !v.UpdateConfig.IsUnknown() {
		vUpdateConfigList := make([]UpdateConfigValue, 0, len(v.UpdateConfig.Elements()))
		d = v.UpdateConfig.ElementsAs(ctx, &vUpdateConfigList, false)
		diags = append(diags, d...)
		if len(vUpdateConfigList) > 0 {
			ng.UpdateConfig, d = vUpdateConfigList[0].Expand(ctx)
			diags = append(diags, d...)
		}
	}

	if !v.KubeletExtraConfig.IsNull() && !v.KubeletExtraConfig.IsUnknown() {
		vKubeletExtraConfigList := make([]KubeletExtraConfigValue, 0, len(v.KubeletExtraConfig.Elements()))
		d = v.KubeletExtraConfig.ElementsAs(ctx, &vKubeletExtraConfigList, false)
		diags = append(diags, d...)
		if len(vKubeletExtraConfigList) > 0 {
			ng.KubeletExtraConfig, d = vKubeletExtraConfigList[0].Expand(ctx)
			diags = append(diags, d...)
		}
	}

	if !v.SecurityGroups2.IsNull() && !v.SecurityGroups2.IsUnknown() {
		vSecurityGroups2List := make([]SecurityGroups2Value, 0, len(v.SecurityGroups2.Elements()))
		d = v.SecurityGroups2.ElementsAs(ctx, &vSecurityGroups2List, false)
		diags = append(diags, d...)
		if len(vSecurityGroups2List) > 0 {
			ng.SecurityGroups, d = vSecurityGroups2List[0].Expand(ctx)
			diags = append(diags, d...)
		}
	}

	return &ng, diags
}

func (v NodeGroupsMapValue) Expand(ctx context.Context) (*rafay.NodeGroup, diag.Diagnostics) {
	var diags diag.Diagnostics
	var ng rafay.NodeGroup
	if v.IsNull() {
		return &rafay.NodeGroup{}, diags
	}
	// TODO: Map fields appropriately
	return &ng, diags
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
	// TODO: Map fields appropriately
	return &vpc, diags
}

func (v AddonsValue) Expand(ctx context.Context) (*rafay.Addon, diag.Diagnostics) {
	var diags diag.Diagnostics
	var addon rafay.Addon
	if v.IsNull() {
		return &rafay.Addon{}, diags
	}
	// TODO: Map fields appropriately
	return &addon, diags
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

func (v ManagedNodegroupsValue) Expand(ctx context.Context) (*rafay.ManagedNodeGroup, diag.Diagnostics) {
	var diags diag.Diagnostics
	var mng rafay.ManagedNodeGroup
	if v.IsNull() {
		return &rafay.ManagedNodeGroup{}, diags
	}
	// TODO: Map fields appropriately
	return &mng, diags
}

func (v FargateProfilesValue) Expand(ctx context.Context) (*rafay.FargateProfile, diag.Diagnostics) {
	var diags diag.Diagnostics
	var fp rafay.FargateProfile
	if v.IsNull() {
		return &rafay.FargateProfile{}, diags
	}
	// TODO: Map fields appropriately
	return &fp, diags
}

func (v CloudWatchValue) Expand(ctx context.Context) (*rafay.EKSClusterCloudWatch, diag.Diagnostics) {
	var diags diag.Diagnostics
	var cw rafay.EKSClusterCloudWatch
	if v.IsNull() {
		return &rafay.EKSClusterCloudWatch{}, diags
	}
	// TODO: Map fields appropriately
	return &cw, diags
}

func (v SecretsEncryptionValue) Expand(ctx context.Context) (*rafay.SecretsEncryption, diag.Diagnostics) {
	var diags diag.Diagnostics
	var se rafay.SecretsEncryption
	if v.IsNull() {
		return &rafay.SecretsEncryption{}, diags
	}
	// TODO: Map fields appropriately
	return &se, diags
}

func (v IdentityMappingsValue) Expand(ctx context.Context) (*rafay.EKSClusterIdentityMappings, diag.Diagnostics) {
	var diags diag.Diagnostics
	var im rafay.EKSClusterIdentityMappings
	if v.IsNull() {
		return &rafay.EKSClusterIdentityMappings{}, diags
	}
	// TODO: Map fields appropriately
	return &im, diags
}

func (v AccessConfigValue) Expand(ctx context.Context) (*rafay.EKSClusterAccess, diag.Diagnostics) {
	var diags diag.Diagnostics
	var ac rafay.EKSClusterAccess
	if v.IsNull() {
		return &rafay.EKSClusterAccess{}, diags
	}
	// TODO: Map fields appropriately
	return &ac, diags
}

func (v AddonsConfigValue) Expand(ctx context.Context) (*rafay.EKSAddonsConfig, diag.Diagnostics) {
	var diags diag.Diagnostics
	var acfg rafay.EKSAddonsConfig
	if v.IsNull() {
		return &rafay.EKSAddonsConfig{}, diags
	}
	// TODO: Map fields appropriately
	return &acfg, diags
}

func (v AutoModeConfigValue) Expand(ctx context.Context) (*rafay.EKSAutoModeConfig, diag.Diagnostics) {
	var diags diag.Diagnostics
	var amc rafay.EKSAutoModeConfig
	if v.IsNull() {
		return &rafay.EKSAutoModeConfig{}, diags
	}
	// TODO: Map fields appropriately
	return &amc, diags
}

// --- NodeGroup Block Expand Stubs ---
func (v IamValue) Expand(ctx context.Context) (*rafay.NodeGroupIAM, diag.Diagnostics) {
	var diags, d diag.Diagnostics
	var iam rafay.NodeGroupIAM

	// Map string fields
	if !v.AttachPolicyV2.IsNull() && !v.AttachPolicyV2.IsUnknown() {
		var policyDoc *rafay.InlineDocument
		var json2 = jsoniter.ConfigCompatibleWithStandardLibrary
		json2.Unmarshal([]byte(getStringValue(v.AttachPolicyV2)), &policyDoc)
		iam.AttachPolicy = policyDoc
	}
	if !v.InstanceProfileArn.IsNull() && !v.InstanceProfileArn.IsUnknown() {
		iam.InstanceProfileARN = getStringValue(v.InstanceProfileArn)
	}
	if !v.InstanceRoleArn.IsNull() && !v.InstanceRoleArn.IsUnknown() {
		iam.InstanceRoleARN = getStringValue(v.InstanceRoleArn)
	}
	if !v.InstanceRoleName.IsNull() && !v.InstanceRoleName.IsUnknown() {
		iam.InstanceRoleName = getStringValue(v.InstanceRoleName)
	}
	if !v.InstanceRolePermissionBoundary.IsNull() && !v.InstanceRolePermissionBoundary.IsUnknown() {
		iam.InstanceRolePermissionsBoundary = getStringValue(v.InstanceRolePermissionBoundary)
	}

	// Map attach_policy_arns (list of strings)
	if !v.AttachPolicyArns.IsNull() && !v.AttachPolicyArns.IsUnknown() {
		policyArnsList := make([]types.String, 0, len(v.AttachPolicyArns.Elements()))
		d = v.AttachPolicyArns.ElementsAs(ctx, &policyArnsList, false)
		diags = append(diags, d...)
		policyArns := make([]string, 0, len(policyArnsList))
		for _, arn := range policyArnsList {
			policyArns = append(policyArns, getStringValue(arn))
		}
		if len(policyArns) > 0 {
			iam.AttachPolicyARNs = policyArns
		}
	}

	// Map iam_node_group_with_addon_policies block (list)
	if !v.IamNodeGroupWithAddonPolicies.IsNull() && !v.IamNodeGroupWithAddonPolicies.IsUnknown() {
		addonPoliciesList := make([]IamNodeGroupWithAddonPoliciesValue, 0, len(v.IamNodeGroupWithAddonPolicies.Elements()))
		d = v.IamNodeGroupWithAddonPolicies.ElementsAs(ctx, &addonPoliciesList, false)
		diags = append(diags, d...)
		if len(addonPoliciesList) > 0 {
			apExpanded, d := addonPoliciesList[0].Expand(ctx)
			diags = append(diags, d...)
			iam.WithAddonPolicies = apExpanded
		}
	}

	// Map attach_policy block (list)
	if !v.AttachPolicy.IsNull() && !v.AttachPolicy.IsUnknown() {
		attachPolicyList := make([]AttachPolicyValue, 0, len(v.AttachPolicy.Elements()))
		d = v.AttachPolicy.ElementsAs(ctx, &attachPolicyList, false)
		diags = append(diags, d...)
		if len(attachPolicyList) > 0 {
			apExpanded, d := attachPolicyList[0].Expand(ctx)
			diags = append(diags, d...)
			iam.AttachPolicy = apExpanded
		}
	}

	return &iam, diags
}

func (v IamNodeGroupWithAddonPoliciesValue) Expand(ctx context.Context) (*rafay.NodeGroupIAMAddonPolicies, diag.Diagnostics) {
	var diags diag.Diagnostics
	var ap rafay.NodeGroupIAMAddonPolicies

	if v.IsNull() {
		return &rafay.NodeGroupIAMAddonPolicies{}, diags
	}

	if !v.AlbIngress.IsNull() && !v.AlbIngress.IsUnknown() {
		albIngress := getBoolValue(v.AlbIngress)
		ap.AWSLoadBalancerController = &albIngress
	}

	if !v.AppMesh.IsNull() && !v.AppMesh.IsUnknown() {
		appMesh := getBoolValue(v.AppMesh)
		ap.AppMesh = &appMesh
	}

	if !v.AppMeshReview.IsNull() && !v.AppMeshReview.IsUnknown() {
		appMeshReview := getBoolValue(v.AppMeshReview)
		ap.AppMeshPreview = &appMeshReview
	}

	if !v.CertManager.IsNull() && !v.CertManager.IsUnknown() {
		certManager := getBoolValue(v.CertManager)
		ap.CertManager = &certManager
	}

	if !v.CloudWatch.IsNull() && !v.CloudWatch.IsUnknown() {
		cloudWatch := getBoolValue(v.CloudWatch)
		ap.CloudWatch = &cloudWatch
	}

	if !v.Ebs.IsNull() && !v.Ebs.IsUnknown() {
		ebs := getBoolValue(v.Ebs)
		ap.EBS = &ebs
	}

	if !v.Efs.IsNull() && !v.Efs.IsUnknown() {
		efs := getBoolValue(v.Efs)
		ap.EFS = &efs
	}

	if !v.ExternalDns.IsNull() && !v.ExternalDns.IsUnknown() {
		externalDns := getBoolValue(v.ExternalDns)
		ap.ExternalDNS = &externalDns
	}

	if !v.Fsx.IsNull() && !v.Fsx.IsUnknown() {
		fsx := getBoolValue(v.Fsx)
		ap.FSX = &fsx
	}

	if !v.Xray.IsNull() && !v.Xray.IsUnknown() {
		xray := getBoolValue(v.Xray)
		ap.XRay = &xray
	}

	if !v.ImageBuilder.IsNull() && !v.ImageBuilder.IsUnknown() {
		imageBuilder := getBoolValue(v.ImageBuilder)
		ap.ImageBuilder = &imageBuilder
	}

	if !v.AutoScaler.IsNull() && !v.AutoScaler.IsUnknown() {
		autoScaler := getBoolValue(v.AutoScaler)
		ap.AutoScaler = &autoScaler
	}

	return &ap, diags
}

func (v AttachPolicyValue) Expand(ctx context.Context) (*rafay.InlineDocument, diag.Diagnostics) {
	var diags diag.Diagnostics
	var policy rafay.InlineDocument

	if v.IsNull() {
		return &rafay.InlineDocument{}, diags
	}

	if !v.Version.IsNull() || !v.Version.IsUnknown() {
		policy.Version = getStringValue(v.Version)
	}

	if !v.Statement.IsNull() || !v.Statement.IsUnknown() {
		statementsList := make([]StatementValue, 0, len(v.Statement.Elements()))
		d := v.Statement.ElementsAs(ctx, &statementsList, false)
		diags = append(diags, d...)
		statements := make([]rafay.InlineStatement, 0, len(statementsList))
		for _, stmt := range statementsList {
			stmtMap, d := stmt.Expand(ctx)
			diags = append(diags, d...)
			statements = append(statements, stmtMap)
		}
		if len(statements) > 0 {
			policy.Statement = statements
		}
	}

	if !v.Id.IsNull() || !v.Id.IsUnknown() {
		policy.Id = getStringValue(v.Id)
	}

	return &policy, diags
}

func (v StatementValue) Expand(ctx context.Context) (rafay.InlineStatement, diag.Diagnostics) {
	var diags diag.Diagnostics
	var stmt rafay.InlineStatement

	if v.IsNull() {
		return rafay.InlineStatement{}, diags
	}
	// Map string fields
	if !v.Effect.IsNull() && !v.Effect.IsUnknown() {
		stmt.Effect = getStringValue(v.Effect)
	}
	if !v.Sid.IsNull() && !v.Sid.IsUnknown() {
		stmt.Sid = getStringValue(v.Sid)
	}

	if !v.Action.IsNull() && !v.Action.IsUnknown() {
		actsList := make([]types.String, 0, len(v.Action.Elements()))
		d := v.Action.ElementsAs(ctx, &actsList, false)
		diags = append(diags, d...)
		if len(actsList) > 0 {
			actionStrs := make([]string, 0, len(actsList))
			for _, act := range actsList {
				actionStrs = append(actionStrs, getStringValue(act))
			}
			stmt.Action = actionStrs
		}
	}

	if !v.NotAction.IsNull() && !v.NotAction.IsUnknown() {
		nactsList := make([]types.String, 0, len(v.NotAction.Elements()))
		d := v.NotAction.ElementsAs(ctx, &nactsList, false)
		diags = append(diags, d...)
		if len(nactsList) > 0 {
			notActionStrs := make([]string, 0, len(nactsList))
			for _, act := range nactsList {
				notActionStrs = append(notActionStrs, getStringValue(act))
			}
			stmt.NotAction = notActionStrs
		}
	}

	if !v.NotResource.IsNull() && !v.NotResource.IsUnknown() {
		nresList := make([]types.String, 0, len(v.NotResource.Elements()))
		d := v.NotResource.ElementsAs(ctx, &nresList, false)
		diags = append(diags, d...)
		if len(nresList) > 0 {
			notResourceStrs := make([]string, 0, len(nresList))
			for _, res := range nresList {
				notResourceStrs = append(notResourceStrs, getStringValue(res))
			}
			stmt.NotResource = notResourceStrs
		}
	}

	if !v.Principal.IsNull() && !v.Principal.IsUnknown() {
		var policyDoc map[string]interface{}
		var json2 = jsoniter.ConfigCompatibleWithStandardLibrary
		json2.Unmarshal([]byte(getStringValue(v.Condition)), &policyDoc)
		stmt.Principal = policyDoc
	}

	if !v.Resource.IsNull() && !v.Resource.IsUnknown() {
		stmt.Resource = getStringValue(v.Resource)
	}

	if !v.NotPrincipal.IsNull() && !v.NotPrincipal.IsUnknown() {
		var policyDoc map[string]interface{}
		var json2 = jsoniter.ConfigCompatibleWithStandardLibrary
		json2.Unmarshal([]byte(getStringValue(v.Condition)), &policyDoc)
		stmt.NotPrincipal = policyDoc
	}

	if !v.Condition.IsNull() && !v.Condition.IsUnknown() {
		var policyDoc map[string]interface{}
		var json2 = jsoniter.ConfigCompatibleWithStandardLibrary
		json2.Unmarshal([]byte(getStringValue(v.Condition)), &policyDoc)
		stmt.Condition = policyDoc
	}

	return stmt, diags
}

func (v SshValue) Expand(ctx context.Context) (*rafay.NodeGroupSSH, diag.Diagnostics) {
	var diags, d diag.Diagnostics
	var ssh rafay.NodeGroupSSH

	// Map allow (bool)
	if !v.Allow.IsNull() && !v.Allow.IsUnknown() {
		allow := getBoolValue(v.Allow)
		ssh.Allow = &allow
	}

	// Map public_key (string)
	if !v.PublicKey.IsNull() && !v.PublicKey.IsUnknown() {
		ssh.PublicKey = getStringValue(v.PublicKey)
	}

	// Map public_key_name (string)
	if !v.PublicKeyName.IsNull() && !v.PublicKeyName.IsUnknown() {
		ssh.PublicKeyName = getStringValue(v.PublicKeyName)
	}

	// Map source_security_group_ids (list of strings)
	if !v.SourceSecurityGroupIds.IsNull() && !v.SourceSecurityGroupIds.IsUnknown() {
		groupIdsList := make([]types.String, 0, len(v.SourceSecurityGroupIds.Elements()))
		d = v.SourceSecurityGroupIds.ElementsAs(ctx, &groupIdsList, false)
		diags = append(diags, d...)
		groupIds := make([]string, 0, len(groupIdsList))
		for _, gid := range groupIdsList {
			groupIds = append(groupIds, getStringValue(gid))
		}
		if len(groupIds) > 0 {
			ssh.SourceSecurityGroupIDs = groupIds
		}
	}

	// Map enable_ssm (bool)
	if !v.EnableSsm.IsNull() && !v.EnableSsm.IsUnknown() {
		enableSsm := getBoolValue(v.EnableSsm)
		ssh.EnableSSM = &enableSsm
	}

	return &ssh, diags
}

// PlacementValue Expand
func (v PlacementValue) Expand(ctx context.Context) (*rafay.Placement, diag.Diagnostics) {
	var diags diag.Diagnostics
	var placement rafay.Placement

	if !v.Group.IsNull() && !v.Group.IsUnknown() {
		group := getStringValue(v.Group)
		placement.GroupName = group
	}

	return &placement, diags
}

func (v InstanceSelectorValue) Expand(ctx context.Context) (*rafay.InstanceSelector, diag.Diagnostics) {
	var diags diag.Diagnostics
	var ins rafay.InstanceSelector

	if !v.Vcpus.IsNull() && !v.Vcpus.IsUnknown() {
		vcpus := int(getInt64Value(v.Vcpus))
		ins.VCPUs = &vcpus
	}

	if !v.Memory.IsNull() && !v.Memory.IsUnknown() {
		memory := getStringValue(v.Memory)
		ins.Memory = memory
	}

	if !v.Gpus.IsNull() && !v.Gpus.IsUnknown() {
		gpus := int(getInt64Value(v.Gpus))
		ins.GPUs = &gpus
	}

	if !v.CpuArchitecture.IsNull() && !v.CpuArchitecture.IsUnknown() {
		cpuArch := getStringValue(v.CpuArchitecture)
		ins.CPUArchitecture = cpuArch
	}

	return &ins, diags
}

func (v BottleRocketValue) Expand(ctx context.Context) (*rafay.NodeGroupBottlerocket, diag.Diagnostics) {
	var diags diag.Diagnostics
	var br rafay.NodeGroupBottlerocket

	if !v.EnableAdminContainer.IsNull() && !v.EnableAdminContainer.IsUnknown() {
		enableAdminContainer := getBoolValue(v.EnableAdminContainer)
		br.EnableAdminContainer = &enableAdminContainer
	}

	if !v.Settings.IsNull() && !v.Settings.IsUnknown() {
		var policyDoc map[string]interface{}
		var json2 = jsoniter.ConfigCompatibleWithStandardLibrary
		json2.Unmarshal([]byte(getStringValue(v.Settings)), &policyDoc)
		br.Settings = policyDoc
	}

	return &br, diags
}

func (v InstancesDistributionValue) Expand(ctx context.Context) (*rafay.NodeGroupInstancesDistribution, diag.Diagnostics) {
	var diags, d diag.Diagnostics
	var dist rafay.NodeGroupInstancesDistribution

	// Map instance_types (list of strings)
	if !v.InstanceTypes.IsNull() && !v.InstanceTypes.IsUnknown() {
		instanceTypesList := make([]types.String, 0, len(v.InstanceTypes.Elements()))
		d = v.InstanceTypes.ElementsAs(ctx, &instanceTypesList, false)
		diags = append(diags, d...)
		instanceTypes := make([]string, 0, len(instanceTypesList))
		for _, it := range instanceTypesList {
			instanceTypes = append(instanceTypes, getStringValue(it))
		}
		if len(instanceTypes) > 0 {
			dist.InstanceTypes = instanceTypes
		}
	}

	// Map max_price (float64)
	if !v.MaxPrice.IsNull() && !v.MaxPrice.IsUnknown() {
		maxPrice := getFloat64Value(v.MaxPrice)
		dist.MaxPrice = &maxPrice
	}

	// Map on_demand_base_capacity (int64)
	if !v.OnDemandBaseCapacity.IsNull() && !v.OnDemandBaseCapacity.IsUnknown() {
		baseCap := int(getInt64Value(v.OnDemandBaseCapacity))
		dist.OnDemandBaseCapacity = &baseCap
	}

	// Map on_demand_percentage_above_base_capacity (int64)
	if !v.OnDemandPercentageAboveBaseCapacity.IsNull() && !v.OnDemandPercentageAboveBaseCapacity.IsUnknown() {
		pct := int(getInt64Value(v.OnDemandPercentageAboveBaseCapacity))
		dist.OnDemandPercentageAboveBaseCapacity = &pct
	}

	// Map spot_instance_pools (int64)
	if !v.SpotInstancePools.IsNull() && !v.SpotInstancePools.IsUnknown() {
		pools := int(getInt64Value(v.SpotInstancePools))
		dist.SpotInstancePools = &pools
	}

	// Map spot_allocation_strategy (string)
	if !v.SpotAllocationStrategy.IsNull() && !v.SpotAllocationStrategy.IsUnknown() {
		dist.SpotAllocationStrategy = getStringValue(v.SpotAllocationStrategy)
	}

	// Map capacity_rebalance (bool)
	if !v.CapacityRebalance.IsNull() && !v.CapacityRebalance.IsUnknown() {
		rebalance := getBoolValue(v.CapacityRebalance)
		dist.CapacityRebalance = &rebalance
	}

	return &dist, diags
}

func (v AsgMetricsCollectionValue) Expand(ctx context.Context) (rafay.MetricsCollection, diag.Diagnostics) {
	var diags diag.Diagnostics
	var metrics rafay.MetricsCollection

	if !v.Granularity.IsNull() && !v.Granularity.IsUnknown() {
		metrics.Granularity = getStringValue(v.Granularity)
	}

	if !v.Metrics.IsNull() && !v.Metrics.IsUnknown() {
		metricsList := make([]types.String, 0, len(v.Metrics.Elements()))
		d := v.Metrics.ElementsAs(ctx, &metricsList, false)
		diags = append(diags, d...)
		if len(metricsList) > 0 {
			metricStrs := make([]string, 0, len(metricsList))
			for _, m := range metricsList {
				metricStrs = append(metricStrs, getStringValue(m))
			}
			metrics.Metrics = metricStrs
		}
	}

	return metrics, diags
}

func (v TaintsValue) Expand(ctx context.Context) (rafay.NodeGroupTaint, diag.Diagnostics) {
	var diags diag.Diagnostics
	var taint rafay.NodeGroupTaint

	if !v.Key.IsNull() && !v.Key.IsUnknown() {
		taint.Key = getStringValue(v.Key)
	}

	if !v.Value.IsNull() && !v.Value.IsUnknown() {
		taint.Value = getStringValue(v.Value)
	}

	if !v.Effect.IsNull() && !v.Effect.IsUnknown() {
		taint.Effect = getStringValue(v.Effect)
	}

	return taint, diags
}

func (v UpdateConfigValue) Expand(ctx context.Context) (*rafay.NodeGroupUpdateConfig, diag.Diagnostics) {
	var diags diag.Diagnostics
	var updateConfig rafay.NodeGroupUpdateConfig

	if !v.MaxUnavaliable.IsNull() && !v.MaxUnavaliable.IsUnknown() {
		maxUnavailable := int(getInt64Value(v.MaxUnavaliable))
		updateConfig.MaxUnavailable = &maxUnavailable
	}

	if !v.MaxUnavaliablePercetage.IsNull() && !v.MaxUnavaliablePercetage.IsUnknown() {
		maxUnavailablePct := int(getInt64Value(v.MaxUnavaliablePercetage))
		updateConfig.MaxUnavailablePercentage = &maxUnavailablePct
	}

	return &updateConfig, diags
}

func (v KubeletExtraConfigValue) Expand(ctx context.Context) (*rafay.KubeletExtraConfig, diag.Diagnostics) {
	var diags, d diag.Diagnostics
	var kec rafay.KubeletExtraConfig

	// Map kube_reserved (map[string]string)
	if !v.KubeReserved.IsNull() && !v.KubeReserved.IsUnknown() {
		kubeReserved := make(map[string]string, len(v.KubeReserved.Elements()))
		vKubeReserved := make(map[string]types.String, len(v.KubeReserved.Elements()))
		d = v.KubeReserved.ElementsAs(ctx, &vKubeReserved, false)
		diags = append(diags, d...)
		for k, val := range vKubeReserved {
			kubeReserved[k] = getStringValue(val)
		}
		if len(kubeReserved) > 0 {
			kec.KubeReserved = kubeReserved
		}
	}

	// Map kube_reserved_cgroup (string)
	if !v.KubeReservedCgroup.IsNull() && !v.KubeReservedCgroup.IsUnknown() {
		kec.KubeReservedCGroup = getStringValue(v.KubeReservedCgroup)
	}

	// Map system_reserved (map[string]string)
	if !v.SystemReserved.IsNull() && !v.SystemReserved.IsUnknown() {
		systemReserved := make(map[string]string, len(v.SystemReserved.Elements()))
		vSystemReserved := make(map[string]types.String, len(v.SystemReserved.Elements()))
		d = v.SystemReserved.ElementsAs(ctx, &vSystemReserved, false)
		diags = append(diags, d...)
		for k, val := range vSystemReserved {
			systemReserved[k] = getStringValue(val)
		}
		if len(systemReserved) > 0 {
			kec.SystemReserved = systemReserved
		}
	}

	// Map eviction_hard (map[string]string)
	if !v.EvictionHard.IsNull() && !v.EvictionHard.IsUnknown() {
		evictionHard := make(map[string]string, len(v.EvictionHard.Elements()))
		vEvictionHard := make(map[string]types.String, len(v.EvictionHard.Elements()))
		d = v.EvictionHard.ElementsAs(ctx, &vEvictionHard, false)
		diags = append(diags, d...)
		for k, val := range vEvictionHard {
			evictionHard[k] = getStringValue(val)
		}
		if len(evictionHard) > 0 {
			kec.EvictionHard = evictionHard
		}
	}

	// Map feature_gates (map[string]bool)
	if !v.FeatureGates.IsNull() && !v.FeatureGates.IsUnknown() {
		featureGates := make(map[string]bool, len(v.FeatureGates.Elements()))
		vFeatureGates := make(map[string]types.Bool, len(v.FeatureGates.Elements()))
		d = v.FeatureGates.ElementsAs(ctx, &vFeatureGates, false)
		diags = append(diags, d...)
		for k, val := range vFeatureGates {
			featureGates[k] = getBoolValue(val)
		}
		if len(featureGates) > 0 {
			kec.FeatureGates = featureGates
		}
	}

	return &kec, diags
}

// SecurityGroups2Value Expand
func (v SecurityGroups2Value) Expand(ctx context.Context) (*rafay.NodeGroupSGs, diag.Diagnostics) {
	var diags, d diag.Diagnostics
	var sgs rafay.NodeGroupSGs

	// Map with_shared (bool)
	if !v.WithShared.IsNull() && !v.WithShared.IsUnknown() {
		withShared := getBoolValue(v.WithShared)
		sgs.WithShared = &withShared
	}

	// Map with_local (bool)
	if !v.WithLocal.IsNull() && !v.WithLocal.IsUnknown() {
		withLocal := getBoolValue(v.WithLocal)
		sgs.WithLocal = &withLocal
	}

	// Map attach_ids (list of strings)
	if !v.AttachIds.IsNull() && !v.AttachIds.IsUnknown() {
		attachIdsList := make([]types.String, 0, len(v.AttachIds.Elements()))
		d = v.AttachIds.ElementsAs(ctx, &attachIdsList, false)
		diags = append(diags, d...)
		attachIds := make([]string, 0, len(attachIdsList))
		for _, id := range attachIdsList {
			attachIds = append(attachIds, getStringValue(id))
		}
		if len(attachIds) > 0 {
			sgs.AttachIDs = attachIds
		}
	}

	return &sgs, diags
}
