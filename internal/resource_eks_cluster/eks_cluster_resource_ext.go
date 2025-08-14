package resource_eks_cluster

import (
	"context"

	"github.com/RafaySystems/terraform-provider-rafay/rafay"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
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
	_ = d

	return &metadata, diags
}

func (v SpecValue) Expand(ctx context.Context) (*rafay.EKSSpec, diag.Diagnostics) {
	var diags diag.Diagnostics
	var spec rafay.EKSSpec

	if v.IsNull() {
		return &rafay.EKSSpec{}, diags
	}

	spec.Blueprint = getStringValue(v.Blueprint)
	spec.CloudProvider = getStringValue(v.CloudProvider)
	spec.CniProvider = getStringValue(v.CniProvider)
	spec.Type = getStringValue(v.SpecType)

	return &spec, diags
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

	vMetadata2List := make([]Metadata2Value, 0, len(v.Metadata2.Elements()))
	diags = v.Metadata2.ElementsAs(ctx, &vMetadata2List, false)
	vMetadata2 := vMetadata2List[0]
	md, d := vMetadata2.Expand(ctx)
	diags = append(diags, d...)
	clusterConfig.Metadata = md

	vNodeGroupsList := make([]NodeGroupsValue, 0, len(v.NodeGroups.Elements()))
	diags = v.NodeGroups.ElementsAs(ctx, &vNodeGroupsList, false)
	ngs := make([]*rafay.NodeGroup, 0, len(vNodeGroupsList))
	for _, vng := range vNodeGroupsList {
		ng, d := vng.Expand(ctx)
		diags = append(diags, d...)
		ngs = append(ngs, ng)
	}
	clusterConfig.NodeGroups = ngs

	vngMap := make(map[string]NodeGroupsMapValue, len(v.NodeGroupsMap.Elements()))
	d = v.NodeGroupsMap.ElementsAs(ctx, &vngMap, false)
	diags = append(diags, d...)

	ngsMap := make([]*rafay.NodeGroup, 0, len(vngMap))
	for ngName, ngMap := range vngMap {
		ngMap, d := ngMap.Expand(ctx)
		diags = append(diags, d...)
		ngMap.Name = ngName

		ngsMap = append(ngsMap, ngMap)
	}
	clusterConfig.NodeGroups = ngsMap

	return &clusterConfig, diags
}

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

	return &md, diags
}

func (v NodeGroupsValue) Expand(ctx context.Context) (*rafay.NodeGroup, diag.Diagnostics) {
	var diags, d diag.Diagnostics
	var ng rafay.NodeGroup

	if v.IsNull() {
		return &rafay.NodeGroup{}, diags
	}

	ng.AMIFamily = getStringValue(v.AmiFamily)

	dc := int(getInt64Value(v.DesiredCapacity))
	ng.DesiredCapacity = &dc

	dimds := getBoolValue(v.DisableImdsv1)
	ng.DisableIMDSv1 = &dimds

	dpodimds := getBoolValue(v.DisablePodsImds)
	ng.DisablePodIMDS = &dpodimds

	efa := getBoolValue(v.EfaEnabled)
	ng.EFAEnabled = &efa

	vIamList := make([]IamValue, 0, len(v.Iam.Elements()))
	diags = v.Iam.ElementsAs(ctx, &vIamList, false)
	vIam := vIamList[0]
	iam, d := vIam.Expand(ctx)
	diags = append(diags, d...)
	ng.IAM = iam

	ng.InstanceType = getStringValue(v.InstanceType)

	ng.MaxPodsPerNode = int(getInt64Value(v.MaxPodsPerNode))

	maxS := int(getInt64Value(v.MaxSize))
	ng.MaxSize = &maxS

	minS := int(getInt64Value(v.MinSize))
	ng.MinSize = &minS

	pn := getBoolValue(v.PrivateNetworking)
	ng.PrivateNetworking = &pn

	ng.Name = getStringValue(v.Name)

	ng.Version = getStringValue(v.Version)

	volumeIOPS := int(getInt64Value(v.VolumeIops))
	ng.VolumeIOPS = &volumeIOPS

	volumeSize := int(getInt64Value(v.VolumeSize))
	ng.VolumeSize = &volumeSize

	volumeThrp := int(getInt64Value(v.VolumeThroughput))
	ng.VolumeThroughput = &volumeThrp

	ng.VolumeType = getStringValue(v.VolumeType)

	// TODO(Akshay): Update later
	f := false
	ng.EBSOptimized = &f
	ng.EnableDetailedMonitoring = &f
	ng.VolumeEncrypted = &f

	return &ng, diags
}

func (v IamValue) Expand(ctx context.Context) (*rafay.NodeGroupIAM, diag.Diagnostics) {
	var diags, d diag.Diagnostics
	var iam rafay.NodeGroupIAM

	vAddonsList := make([]IamNodeGroupWithAddonPoliciesValue, 0, len(v.IamNodeGroupWithAddonPolicies.Elements()))
	diags = v.IamNodeGroupWithAddonPolicies.ElementsAs(ctx, &vAddonsList, false)
	vAddons := vAddonsList[0]
	iam.WithAddonPolicies, d = vAddons.Expand(ctx)
	diags = append(diags, d...)

	return &iam, diags
}

func (v IamNodeGroupWithAddonPoliciesValue) Expand(ctx context.Context) (*rafay.NodeGroupIAMAddonPolicies, diag.Diagnostics) {
	var diags diag.Diagnostics
	var addons rafay.NodeGroupIAMAddonPolicies

	alb := getBoolValue(v.AlbIngress)
	addons.AWSLoadBalancerController = &alb

	am := getBoolValue(v.AppMesh)
	addons.AppMesh = &am

	amp := getBoolValue(v.AppMeshReview)
	addons.AppMeshPreview = &amp

	as := getBoolValue(v.AutoScaler)
	addons.AutoScaler = &as

	cm := getBoolValue(v.CertManager)
	addons.CertManager = &cm

	cw := getBoolValue(v.CloudWatch)
	addons.CloudWatch = &cw

	ebs := getBoolValue(v.Ebs)
	addons.EBS = &ebs

	efs := getBoolValue(v.Efs)
	addons.EFS = &efs

	ed := getBoolValue(v.ExternalDns)
	addons.ExternalDNS = &ed

	fsx := getBoolValue(v.Fsx)
	addons.FSX = &fsx

	im := getBoolValue(v.ImageBuilder)
	addons.ImageBuilder = &im

	xray := getBoolValue(v.Xray)
	addons.XRay = &xray

	return &addons, diags
}

func (v NodeGroupsMapValue) Expand(ctx context.Context) (*rafay.NodeGroup, diag.Diagnostics) {
	var diags, d diag.Diagnostics
	var ng rafay.NodeGroup

	if v.IsNull() {
		return &rafay.NodeGroup{}, diags
	}

	ng.AMIFamily = getStringValue(v.AmiFamily)

	dc := int(getInt64Value(v.DesiredCapacity))
	ng.DesiredCapacity = &dc

	dimds := getBoolValue(v.DisableImdsv1)
	ng.DisableIMDSv1 = &dimds

	dpodimds := getBoolValue(v.DisablePodsImds)
	ng.DisablePodIMDS = &dpodimds

	efa := getBoolValue(v.EfaEnabled)
	ng.EFAEnabled = &efa

	vIam2List := make([]Iam2Value, 0, len(v.Iam2.Elements()))
	diags = v.Iam2.ElementsAs(ctx, &vIam2List, false)
	vIam2 := vIam2List[0]
	iam, d := vIam2.Expand(ctx)
	diags = append(diags, d...)
	ng.IAM = iam

	ng.InstanceType = getStringValue(v.InstanceType)

	ng.MaxPodsPerNode = int(getInt64Value(v.MaxPodsPerNode))

	maxS := int(getInt64Value(v.MaxSize))
	ng.MaxSize = &maxS

	minS := int(getInt64Value(v.MinSize))
	ng.MinSize = &minS

	pn := getBoolValue(v.PrivateNetworking)
	ng.PrivateNetworking = &pn

	ng.Version = getStringValue(v.Version)

	volumeIOPS := int(getInt64Value(v.VolumeIops))
	ng.VolumeIOPS = &volumeIOPS

	volumeSize := int(getInt64Value(v.VolumeSize))
	ng.VolumeSize = &volumeSize

	volumeThrp := int(getInt64Value(v.VolumeThroughput))
	ng.VolumeThroughput = &volumeThrp

	ng.VolumeType = getStringValue(v.VolumeType)

	// TODO(Akshay): Update later
	f := false
	ng.EBSOptimized = &f
	ng.EnableDetailedMonitoring = &f
	ng.VolumeEncrypted = &f

	return &ng, diags
}

func (v Iam2Value) Expand(ctx context.Context) (*rafay.NodeGroupIAM, diag.Diagnostics) {
	var diags, d diag.Diagnostics
	var iam rafay.NodeGroupIAM

	vAddons2List := make([]IamNodeGroupWithAddonPolicies2Value, 0, len(v.IamNodeGroupWithAddonPolicies2.Elements()))
	diags = v.IamNodeGroupWithAddonPolicies2.ElementsAs(ctx, &vAddons2List, false)
	vAddons2 := vAddons2List[0]
	iam.WithAddonPolicies, d = vAddons2.Expand(ctx)
	diags = append(diags, d...)

	return &iam, diags
}

func (v IamNodeGroupWithAddonPolicies2Value) Expand(ctx context.Context) (*rafay.NodeGroupIAMAddonPolicies, diag.Diagnostics) {
	var diags diag.Diagnostics
	var addons rafay.NodeGroupIAMAddonPolicies

	alb := getBoolValue(v.AlbIngress)
	addons.AWSLoadBalancerController = &alb

	am := getBoolValue(v.AppMesh)
	addons.AppMesh = &am

	amp := getBoolValue(v.AppMeshReview)
	addons.AppMeshPreview = &amp

	as := getBoolValue(v.AutoScaler)
	addons.AutoScaler = &as

	cm := getBoolValue(v.CertManager)
	addons.CertManager = &cm

	cw := getBoolValue(v.CloudWatch)
	addons.CloudWatch = &cw

	ebs := getBoolValue(v.Ebs)
	addons.EBS = &ebs

	efs := getBoolValue(v.Efs)
	addons.EFS = &efs

	ed := getBoolValue(v.ExternalDns)
	addons.ExternalDNS = &ed

	fsx := getBoolValue(v.Fsx)
	addons.FSX = &fsx

	im := getBoolValue(v.ImageBuilder)
	addons.ImageBuilder = &im

	xray := getBoolValue(v.Xray)
	addons.XRay = &xray

	return &addons, diags
}
