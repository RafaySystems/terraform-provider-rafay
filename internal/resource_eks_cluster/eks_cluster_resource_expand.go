package resource_eks_cluster

import (
	"context"

	"github.com/RafaySystems/terraform-provider-rafay/rafay"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
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
// TODO: Implement Expand functions for: KubernetesNetworkConfigValue, Iam3Value, IdentityProvidersValue, VpcValue, AddonsValue, PrivateClusterValue, ManagedNodeGroupsValue, FargateProfilesValue, CloudWatchValue, SecretsEncryptionValue, IdentityMappingsValue, AccessConfigValue, AddonsConfigValue, AutoModeConfigValue, NodeGroupsMapValue

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
	ng.Name = getStringValue(v.Name)
	ng.Version = getStringValue(v.Version)
	ng.AMIFamily = getStringValue(v.AmiFamily)
	ng.InstanceType = getStringValue(v.InstanceType)
	ng.MaxPodsPerNode = int(getInt64Value(v.MaxPodsPerNode))
	ng.VolumeType = getStringValue(v.VolumeType)

	volSize := int(getInt64Value(v.VolumeSize))
	ng.VolumeSize = &volSize
	volIops := int(getInt64Value(v.VolumeIops))
	ng.VolumeIOPS = &volIops
	volThroughput := int(getInt64Value(v.VolumeThroughput))
	ng.VolumeThroughput = &volThroughput
	desiredCap := int(getInt64Value(v.DesiredCapacity))
	ng.DesiredCapacity = &desiredCap
	maxSize := int(getInt64Value(v.MaxSize))
	ng.MaxSize = &maxSize
	minSize := int(getInt64Value(v.MinSize))
	ng.MinSize = &minSize
	privNet := getBoolValue(v.PrivateNetworking)
	ng.PrivateNetworking = &privNet
	disImdsv1 := getBoolValue(v.DisableImdsv1)
	ng.DisableIMDSv1 = &disImdsv1
	disPodImds := getBoolValue(v.DisablePodsImds)
	ng.DisablePodIMDS = &disPodImds
	efaEnabled := getBoolValue(v.EfaEnabled)
	ng.EFAEnabled = &efaEnabled

	ng.Labels = make(map[string]string, len(v.Labels2.Elements()))
	vLabels := make(map[string]types.String, len(v.Labels2.Elements()))
	d = v.Labels2.ElementsAs(ctx, &vLabels, false)
	diags = append(diags, d...)
	for k, val := range vLabels {
		ng.Labels[k] = getStringValue(val)
	}
	ng.Tags = make(map[string]string, len(v.Tags2.Elements()))
	vTags := make(map[string]types.String, len(v.Tags2.Elements()))
	d = v.Tags2.ElementsAs(ctx, &vTags, false)
	diags = append(diags, d...)
	for k, val := range vTags {
		ng.Tags[k] = getStringValue(val)
	}

	// Map additional fields from spec
	// ng.SubnetCidr = getStringValue(v.SubnetCidr) // Field not present in struct, TODO: add if needed
	ng.ClusterDNS = getStringValue(v.ClusterDns)

	// target_group_arns (list of strings)
	tgArnsList := make([]types.String, 0, len(v.TargetGroupArns.Elements()))
	d = v.TargetGroupArns.ElementsAs(ctx, &tgArnsList, false)
	diags = append(diags, d...)
	tgArns := make([]string, 0, len(tgArnsList))
	for _, tg := range tgArnsList {
		tgArns = append(tgArns, getStringValue(tg))
	}
	ng.TargetGroupARNs = tgArns

	// classic_load_balancer_names (list of strings)
	lbNamesList := make([]types.String, 0, len(v.ClassicLoadBalancerNames.Elements()))
	d = v.ClassicLoadBalancerNames.ElementsAs(ctx, &lbNamesList, false)
	diags = append(diags, d...)
	lbNames := make([]string, 0, len(lbNamesList))
	for _, lb := range lbNamesList {
		lbNames = append(lbNames, getStringValue(lb))
	}
	ng.ClassicLoadBalancerNames = lbNames

	ng.CPUCredits = getStringValue(v.CpuCredits)
	enableDM := getBoolValue(v.EnableDetailedMonitoring)
	ng.EnableDetailedMonitoring = &enableDM

	// availability_zones2 (list of strings)
	// Field not present in struct, TODO: add if needed
	azList := make([]types.String, 0, len(v.AvailabilityZones2.Elements()))
	d = v.AvailabilityZones2.ElementsAs(ctx, &azList, false)
	diags = append(diags, d...)
	azs := make([]string, 0, len(azList))
	for _, az := range azList {
		azs = append(azs, getStringValue(az))
	}
	ng.AvailabilityZones = azs

	// subnets (list of strings)
	subnetsList := make([]types.String, 0, len(v.Subnets.Elements()))
	d = v.Subnets.ElementsAs(ctx, &subnetsList, false)
	diags = append(diags, d...)
	subnets := make([]string, 0, len(subnetsList))
	for _, s := range subnetsList {
		subnets = append(subnets, getStringValue(s))
	}
	ng.Subnets = subnets

	ng.InstancePrefix = getStringValue(v.InstancePrefix)
	ng.InstanceName = getStringValue(v.InstanceName)
	ng.AMI = getStringValue(v.Ami)

	// asg_suspend_processes (list of strings)
	asgSuspendList := make([]types.String, 0, len(v.AsgSuspendProcesses.Elements()))
	d = v.AsgSuspendProcesses.ElementsAs(ctx, &asgSuspendList, false)
	asgSuspend := make([]string, 0, len(asgSuspendList))
	diags = append(diags, d...)
	for _, p := range asgSuspendList {
		asgSuspend = append(asgSuspend, getStringValue(p))
	}
	ng.ASGSuspendProcesses = asgSuspend

	ebsOpt := getBoolValue(v.EbsOptimized)
	ng.EBSOptimized = &ebsOpt
	ng.VolumeName = getStringValue(v.VolumeName)
	volEncrypted := getBoolValue(v.VolumeEncrypted)
	ng.VolumeEncrypted = &volEncrypted
	ng.VolumeKmsKeyID = getStringValue(v.VolumeKmsKeyId)

	// pre_bootstrap_commands (list of strings)
	preBootstrapList := make([]types.String, 0, len(v.PreBootstrapCommands.Elements()))
	d = v.PreBootstrapCommands.ElementsAs(ctx, &preBootstrapList, false)
	diags = append(diags, d...)
	preBootstrap := make([]string, 0, len(preBootstrapList))
	for _, cmd := range preBootstrapList {
		preBootstrap = append(preBootstrap, getStringValue(cmd))
	}
	ng.PreBootstrapCommands = preBootstrap

	ng.OverrideBootstrapCommand = getStringValue(v.OverrideBootstrapCommand)

	// TODO: Map blocks (IAM, SSH, SecurityGroups, Taints, ScalingConfig, InstancesDistribution, etc.)

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
	// TODO: Map fields appropriately
	return &knc, diags
}

func (v Iam3Value) Expand(ctx context.Context) (*rafay.EKSClusterIAM, diag.Diagnostics) {
	var diags diag.Diagnostics
	var iam rafay.EKSClusterIAM
	if v.IsNull() {
		return &rafay.EKSClusterIAM{}, diags
	}
	// TODO: Map fields appropriately
	return &iam, diags
}

func (v IdentityProvidersValue) Expand(ctx context.Context) (*rafay.IdentityProvider, diag.Diagnostics) {
	var diags diag.Diagnostics
	var idp rafay.IdentityProvider
	if v.IsNull() {
		return &rafay.IdentityProvider{}, diags
	}
	// TODO: Map fields appropriately
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
	// TODO: Map fields appropriately
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
