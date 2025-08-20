package resource_eks_cluster

import (
	"context"

	"github.com/RafaySystems/terraform-provider-rafay/rafay"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var ngMapInUse = true

func FlattenEksCluster(ctx context.Context, c rafay.EKSCluster, data *EksClusterModel) diag.Diagnostics {
	var diags, d diag.Diagnostics

	cv := NewClusterValueNull()
	d = cv.Flatten(ctx, c)
	diags = append(diags, d...)
	clusterElements := []attr.Value{
		cv,
	}
	data.Cluster, d = types.ListValue(ClusterValue{}.Type(ctx), clusterElements)
	diags = append(diags, d...)

	return diags
}

func FlattenEksClusterConfig(ctx context.Context, c rafay.EKSClusterConfig, data *EksClusterModel) diag.Diagnostics {
	var diags, d diag.Diagnostics

	// check ngMap are used
	ccList := make([]ClusterConfigValue, len(data.ClusterConfig.Elements()))
	d = data.ClusterConfig.ElementsAs(ctx, &ccList, false)
	diags = append(diags, d...)
	cc0 := ccList[0]
	if len(cc0.NodeGroups.Elements()) > 0 {
		ngMapInUse = false
	}

	cc := NewClusterConfigValueNull()
	d = cc.Flatten(ctx, c)
	diags = append(diags, d...)
	clusterElements := []attr.Value{
		cc,
	}
	data.ClusterConfig, d = types.ListValue(ClusterConfigValue{}.Type(ctx), clusterElements)
	diags = append(diags, d...)

	return diags
}

func (v *ClusterValue) Flatten(ctx context.Context, c rafay.EKSCluster) diag.Diagnostics {
	var diags, d diag.Diagnostics
	v.Kind = types.StringValue(c.Kind)

	md := NewMetadataValueNull()
	d = md.Flatten(ctx, c.Metadata)
	diags = append(diags, d...)
	mdElements := []attr.Value{
		md,
	}
	v.Metadata, d = types.ListValue(MetadataValue{}.Type(ctx), mdElements)
	diags = append(diags, d...)

	spec := NewSpecValueNull()
	d = spec.Flatten(ctx, c.Spec)
	diags = append(diags, d...)
	specElements := []attr.Value{
		spec,
	}
	v.Spec, d = types.ListValue(SpecValue{}.Type(ctx), specElements)
	diags = append(diags, d...)

	v.state = attr.ValueStateKnown
	return diags
}

func (v *MetadataValue) Flatten(ctx context.Context, md *rafay.EKSClusterMetadata) diag.Diagnostics {
	var diags, d diag.Diagnostics

	v.Name = types.StringValue(md.Name)
	v.Project = types.StringValue(md.Project)

	lbsMap := types.MapNull(types.StringType)
	if len(md.Labels) != 0 {
		lbs := map[string]attr.Value{}
		for key, val := range md.Labels {
			lbs[key] = types.StringValue(val)
		}
		lbsMap, d = types.MapValue(types.StringType, lbs)
		diags = append(diags, d...)
	}
	v.Labels = lbsMap

	v.state = attr.ValueStateKnown
	return diags
}

func (v *SpecValue) Flatten(ctx context.Context, spec *rafay.EKSSpec) diag.Diagnostics {
	var diags, d diag.Diagnostics
	if spec == nil {
		return diags
	}

	v.Blueprint = types.StringValue(spec.Blueprint)
	v.BlueprintVersion = types.StringValue(spec.BlueprintVersion)
	v.CloudProvider = types.StringValue(spec.CloudProvider)
	v.CniProvider = types.StringValue(spec.CniProvider)
	v.CrossAccountRoleArn = types.StringValue(spec.CrossAccountRoleArn)
	v.SpecType = types.StringValue(spec.Type)

	cp := NewCniParamsValueNull()
	d = cp.Flatten(ctx, spec.CniParams)
	diags = append(diags, d...)
	cpElements := []attr.Value{
		cp,
	}
	v.CniParams, d = types.ListValue(CniParamsValue{}.Type(ctx), cpElements)

	//TODO(Akshay): Flatten the proxy config.
	v.ProxyConfig = types.MapNull(types.StringType)

	scp := NewSystemComponentsPlacementValueNull()
	d = scp.Flatten(ctx, spec.SystemComponentsPlacement)
	diags = append(diags, d...)
	v.SystemComponentsPlacement, d = scp.ToObjectValue(ctx)
	diags = append(diags, d...)

	sh := NewSharingValueNull()
	d = sh.Flatten(ctx, spec.Sharing)
	diags = append(diags, d...)
	v.Sharing, d = sh.ToObjectValue(ctx)
	diags = append(diags, d...)

	v.state = attr.ValueStateKnown
	return diags
}

func (v *CniParamsValue) Flatten(ctx context.Context, cp *rafay.CustomCni) diag.Diagnostics {
	var diags, d diag.Diagnostics
	if cp == nil {
		return diags
	}

	v.CustomCniCidr = types.StringValue(cp.CustomCniCidr)

	csElements := make([]attr.Value, 0, len(cp.CustomCniCrdSpec))
	for name, cniSpec := range cp.CustomCniCrdSpec {
		cs := NewCustomCniCrdSpecValueNull()
		d = cs.Flatten(ctx, name, cniSpec)
		diags = append(diags, d...)
	}
	v.CustomCniCrdSpec, d = types.ListValue(CustomCniCrdSpecValue{}.Type(ctx), csElements)
	diags = append(diags, d...)

	v.state = attr.ValueStateKnown
	return diags
}

func (v *CustomCniCrdSpecValue) Flatten(ctx context.Context, name string, cs []rafay.CustomCniSpec) diag.Diagnostics {
	var diags, d diag.Diagnostics

	v.Name = types.StringValue(name)

	specElements := make([]attr.Value, 0, len(cs))
	for _, spec := range cs {
		s := NewCniSpecValueNull()
		d = s.Flatten(ctx, spec)
		diags = append(diags, d...)
	}
	v.CniSpec, d = types.ListValue(CniSpecValue{}.Type(ctx), specElements)
	diags = append(diags, d...)

	v.state = attr.ValueStateKnown
	return diags
}

func (v *CniSpecValue) Flatten(ctx context.Context, spec rafay.CustomCniSpec) diag.Diagnostics {
	var diags, d diag.Diagnostics

	v.Subnet = types.StringValue(spec.Subnet)
	sgElements := make([]attr.Value, 0, len(spec.SecurityGroups))
	for _, sg := range spec.SecurityGroups {
		sgElements = append(sgElements, types.StringValue(sg))
	}
	v.SecurityGroups, d = types.ListValue(types.StringType, sgElements)
	diags = append(diags, d...)

	v.state = attr.ValueStateKnown
	return diags
}

func (v *ClusterConfigValue) Flatten(ctx context.Context, cc rafay.EKSClusterConfig) diag.Diagnostics {
	var diags, d diag.Diagnostics
	v.Apiversion = types.StringValue(cc.APIVersion)
	v.Kind = types.StringValue(cc.Kind)

	md := NewMetadata2ValueNull()
	d = md.Flatten(ctx, cc.Metadata)
	diags = append(diags, d...)
	mdElements := []attr.Value{
		md,
	}
	v.Metadata2, d = types.ListValue(Metadata2Value{}.Type(ctx), mdElements)
	diags = append(diags, d...)

	if ngMapInUse {
		ngMap := types.MapNull(NodeGroupsMapValue{}.Type(ctx))
		if len(cc.NodeGroups) != 0 {
			nodegrp := map[string]attr.Value{}
			for _, ng := range cc.NodeGroups {
				ngrp := NewNodeGroupsMapValueNull()
				d = ngrp.Flatten(ctx, ng)
				diags = append(diags, d...)
				nodegrp[ng.Name] = ngrp
			}
			ngMap, d = types.MapValue(NodeGroupsMapValue{}.Type(ctx), nodegrp)
			diags = append(diags, d...)
		}

		v.NodeGroupsMap = ngMap
		v.NodeGroups = types.ListNull(NodeGroupsValue{}.Type(ctx))
	} else {
		ngElements := []attr.Value{}
		for _, ng := range cc.NodeGroups {
			ngList := NewNodeGroupsValueNull()
			d = ngList.Flatten(ctx, ng)
			diags = append(diags, d...)
			ngElements = append(ngElements, ngList)
		}
		v.NodeGroups, d = types.ListValue(NodeGroupsValue{}.Type(ctx), ngElements)
		diags = append(diags, d...)
		v.NodeGroupsMap = types.MapNull(NodeGroupsMapValue{}.Type(ctx))
	}

	v.state = attr.ValueStateKnown
	return diags
}

func (v *Metadata2Value) Flatten(ctx context.Context, md *rafay.EKSClusterConfigMetadata) diag.Diagnostics {
	var diags, d diag.Diagnostics

	v.Name = types.StringValue(md.Name)
	v.Region = types.StringValue(md.Region)
	v.Version = types.StringValue(md.Version)

	tagMap := types.MapNull(types.StringType)
	if len(md.Tags) != 0 {
		tag := map[string]attr.Value{}
		for key, val := range md.Tags {
			tag[key] = types.StringValue(val)
		}
		tagMap, d = types.MapValue(types.StringType, tag)
		diags = append(diags, d...)
	}
	v.Tags = tagMap

	v.state = attr.ValueStateKnown
	return diags
}

func (v *NodeGroupsMapValue) Flatten(ctx context.Context, ng *rafay.NodeGroup) diag.Diagnostics {
	var diags, d diag.Diagnostics

	v.AmiFamily = types.StringValue(ng.AMIFamily)
	v.DesiredCapacity = types.Int64Value(int64(*ng.DesiredCapacity))
	v.DisableImdsv1 = types.BoolValue(*ng.DisableIMDSv1)
	v.DisablePodsImds = types.BoolValue(*ng.DisablePodIMDS)
	v.EfaEnabled = types.BoolValue(*ng.EFAEnabled)
	v.InstanceType = types.StringValue(ng.InstanceType)
	v.MaxPodsPerNode = types.Int64Value(int64(ng.MaxPodsPerNode))
	v.MaxSize = types.Int64Value(int64(*ng.MaxSize))
	v.MinSize = types.Int64Value(int64(*ng.MinSize))
	v.PrivateNetworking = types.BoolValue(*ng.PrivateNetworking)
	v.Version = types.StringValue(ng.Version)
	v.VolumeIops = types.Int64Value(int64(*ng.VolumeIOPS))
	v.VolumeSize = types.Int64Value(int64(*ng.VolumeSize))
	v.VolumeThroughput = types.Int64Value(int64(*ng.VolumeThroughput))
	v.VolumeType = types.StringValue(ng.VolumeType)

	iam := NewIam2ValueNull()
	d = iam.Flatten(ctx, ng.IAM)
	diags = append(diags, d...)
	iamElements := []attr.Value{iam}
	v.Iam2, d = types.ListValue(Iam2Value{}.Type(ctx), iamElements)
	diags = append(diags, d...)

	v.state = attr.ValueStateKnown
	return diags
}

func (v *Iam2Value) Flatten(ctx context.Context, iam *rafay.NodeGroupIAM) diag.Diagnostics {
	var diags, d diag.Diagnostics
	if iam == nil {
		return diags
	}

	addonPolicies := NewIamNodeGroupWithAddonPolicies2ValueNull()
	d = addonPolicies.Flatten(ctx, iam.WithAddonPolicies)
	diags = append(diags, d...)
	addonPoliciesElements := []attr.Value{addonPolicies}
	v.IamNodeGroupWithAddonPolicies2, d = types.ListValue(IamNodeGroupWithAddonPolicies2Value{}.Type(ctx), addonPoliciesElements)
	diags = append(diags, d...)

	v.state = attr.ValueStateKnown
	return diags
}

func (v *IamNodeGroupWithAddonPolicies2Value) Flatten(ctx context.Context, addonPolicies *rafay.NodeGroupIAMAddonPolicies) diag.Diagnostics {
	var diags diag.Diagnostics
	if addonPolicies == nil {
		return diags
	}

	v.AlbIngress = types.BoolPointerValue(addonPolicies.AWSLoadBalancerController)
	v.AppMesh = types.BoolPointerValue(addonPolicies.AppMesh)
	v.AppMeshReview = types.BoolPointerValue(addonPolicies.AppMeshPreview)
	v.AutoScaler = types.BoolPointerValue(addonPolicies.AutoScaler)
	v.CertManager = types.BoolPointerValue(addonPolicies.CertManager)
	v.CloudWatch = types.BoolPointerValue(addonPolicies.CloudWatch)
	v.Ebs = types.BoolPointerValue(addonPolicies.EBS)
	v.Efs = types.BoolPointerValue(addonPolicies.EFS)
	v.ExternalDns = types.BoolPointerValue(addonPolicies.ExternalDNS)
	v.Fsx = types.BoolPointerValue(addonPolicies.FSX)
	v.ImageBuilder = types.BoolPointerValue(addonPolicies.ImageBuilder)
	v.Xray = types.BoolPointerValue(addonPolicies.XRay)

	v.state = attr.ValueStateKnown
	return diags
}

func (v *NodeGroupsValue) Flatten(ctx context.Context, ng *rafay.NodeGroup) diag.Diagnostics {
	var diags, d diag.Diagnostics

	v.Name = types.StringValue(ng.Name)
	v.AmiFamily = types.StringValue(ng.AMIFamily)
	v.DesiredCapacity = types.Int64Value(int64(*ng.DesiredCapacity))
	v.DisableImdsv1 = types.BoolValue(*ng.DisableIMDSv1)
	v.DisablePodsImds = types.BoolValue(*ng.DisablePodIMDS)
	v.EfaEnabled = types.BoolValue(*ng.EFAEnabled)
	v.InstanceType = types.StringValue(ng.InstanceType)
	v.MaxPodsPerNode = types.Int64Value(int64(ng.MaxPodsPerNode))
	v.MaxSize = types.Int64Value(int64(*ng.MaxSize))
	v.MinSize = types.Int64Value(int64(*ng.MinSize))
	v.PrivateNetworking = types.BoolValue(*ng.PrivateNetworking)
	v.Version = types.StringValue(ng.Version)
	v.VolumeIops = types.Int64Value(int64(*ng.VolumeIOPS))
	v.VolumeSize = types.Int64Value(int64(*ng.VolumeSize))
	v.VolumeThroughput = types.Int64Value(int64(*ng.VolumeThroughput))
	v.VolumeType = types.StringValue(ng.VolumeType)

	iam := NewIamValueNull()
	d = iam.Flatten(ctx, ng.IAM)
	diags = append(diags, d...)
	iamElements := []attr.Value{iam}
	v.Iam, d = types.ListValue(IamValue{}.Type(ctx), iamElements)
	diags = append(diags, d...)

	v.state = attr.ValueStateKnown
	return diags
}

func (v *IamValue) Flatten(ctx context.Context, iam *rafay.NodeGroupIAM) diag.Diagnostics {
	var diags, d diag.Diagnostics
	if iam == nil {
		return diags
	}

	addonPolicies := NewIamNodeGroupWithAddonPoliciesValueNull()
	d = addonPolicies.Flatten(ctx, iam.WithAddonPolicies)
	diags = append(diags, d...)
	addonPoliciesElements := []attr.Value{addonPolicies}
	v.IamNodeGroupWithAddonPolicies, d = types.ListValue(IamNodeGroupWithAddonPoliciesValue{}.Type(ctx), addonPoliciesElements)
	diags = append(diags, d...)

	v.state = attr.ValueStateKnown
	return diags
}

func (v *IamNodeGroupWithAddonPoliciesValue) Flatten(ctx context.Context, addonPolicies *rafay.NodeGroupIAMAddonPolicies) diag.Diagnostics {
	var diags diag.Diagnostics
	if addonPolicies == nil {
		return diags
	}

	v.AlbIngress = types.BoolPointerValue(addonPolicies.AWSLoadBalancerController)
	v.AppMesh = types.BoolPointerValue(addonPolicies.AppMesh)
	v.AppMeshReview = types.BoolPointerValue(addonPolicies.AppMeshPreview)
	v.AutoScaler = types.BoolPointerValue(addonPolicies.AutoScaler)
	v.CertManager = types.BoolPointerValue(addonPolicies.CertManager)
	v.CloudWatch = types.BoolPointerValue(addonPolicies.CloudWatch)
	v.Ebs = types.BoolPointerValue(addonPolicies.EBS)
	v.Efs = types.BoolPointerValue(addonPolicies.EFS)
	v.ExternalDns = types.BoolPointerValue(addonPolicies.ExternalDNS)
	v.Fsx = types.BoolPointerValue(addonPolicies.FSX)
	v.ImageBuilder = types.BoolPointerValue(addonPolicies.ImageBuilder)
	v.Xray = types.BoolPointerValue(addonPolicies.XRay)

	v.state = attr.ValueStateKnown
	return diags
}

func (v *SystemComponentsPlacementValue) Flatten(ctx context.Context, scp *rafay.SystemComponentsPlacement) diag.Diagnostics {
	var diags, d diag.Diagnostics
	if scp == nil {
		return diags
	}

	ns := types.MapNull(types.StringType)
	if len(scp.NodeSelector) != 0 {
		nodeSelector := map[string]attr.Value{}
		for key, val := range scp.NodeSelector {
			nodeSelector[key] = types.StringValue(val)
		}
		ns, d = types.MapValue(types.StringType, nodeSelector)
		diags = append(diags, d...)
	}
	v.NodeSelector = ns

	tolerations := make([]attr.Value, 0, len(scp.Tolerations))
	for _, t := range scp.Tolerations {
		tol := NewTolerationsValueNull()
		d = tol.Flatten(ctx, t)
		diags = append(diags, d...)
		tolerations = append(tolerations, tol)
	}
	v.Tolerations, d = types.ListValue(TolerationsValue{}.Type(ctx), tolerations)
	diags = append(diags, d...)

	// DaemonsetOverride
	daemonsetOverride := NewDaemonsetOverrideValueNull()
	d = daemonsetOverride.Flatten(ctx, scp.DaemonsetOverride)
	diags = append(diags, d...)
	v.DaemonsetOverride, d = types.ListValue(DaemonsetOverrideValue{}.Type(ctx), []attr.Value{daemonsetOverride})
	diags = append(diags, d...)

	v.state = attr.ValueStateKnown
	return diags
}

func (v *TolerationsValue) Flatten(ctx context.Context, t *rafay.Tolerations) diag.Diagnostics {
	var diags diag.Diagnostics
	if t == nil {
		return diags
	}

	v.Key = types.StringValue(t.Key)
	v.Operator = types.StringValue(t.Operator)
	v.Value = types.StringValue(t.Value)
	v.Effect = types.StringValue(t.Effect)
	if t.TolerationSeconds != nil {
		v.TolerationSeconds = types.Int64Value(int64(*t.TolerationSeconds))
	} else {
		v.TolerationSeconds = types.Int64Null()
	}

	v.state = attr.ValueStateKnown
	return diags
}

func (v *DaemonsetOverrideValue) Flatten(ctx context.Context, dso *rafay.DaemonsetOverride) diag.Diagnostics {
	var diags, d diag.Diagnostics
	if dso == nil {
		return diags
	}

	v.NodeSelectionEnabled = types.BoolPointerValue(dso.NodeSelectionEnabled)

	tolerations := make([]attr.Value, 0, len(dso.Tolerations))
	for _, t := range dso.Tolerations {
		tol := NewTolerations2ValueNull()
		d = tol.Flatten(ctx, t)
		diags = append(diags, d...)
		tolerations = append(tolerations, tol)
	}
	v.Tolerations2, d = types.ListValue(Tolerations2Value{}.Type(ctx), tolerations)
	diags = append(diags, d...)

	v.state = attr.ValueStateKnown
	return diags
}

func (v *Tolerations2Value) Flatten(ctx context.Context, t *rafay.Tolerations) diag.Diagnostics {
	var diags diag.Diagnostics
	if t == nil {
		return diags
	}

	v.Key = types.StringValue(t.Key)
	v.Operator = types.StringValue(t.Operator)
	v.Value = types.StringValue(t.Value)
	v.Effect = types.StringValue(t.Effect)
	if t.TolerationSeconds != nil {
		v.TolerationSeconds = types.Int64Value(int64(*t.TolerationSeconds))
	} else {
		v.TolerationSeconds = types.Int64Null()
	}

	v.state = attr.ValueStateKnown
	return diags
}

func (v *SharingValue) Flatten(ctx context.Context, sh *rafay.V1ClusterSharing) diag.Diagnostics {
	var diags, d diag.Diagnostics
	if sh == nil {
		return diags
	}

	v.Enabled = types.BoolPointerValue(sh.Enabled)

	projects := make([]attr.Value, 0, len(sh.Projects))
	for _, p := range sh.Projects {
		proj := NewProjectsValueNull()
		d = proj.Flatten(ctx, p)
		diags = append(diags, d...)
		projects = append(projects, proj)
	}
	v.Projects, d = types.ListValue(ProjectsValue{}.Type(ctx), projects)
	diags = append(diags, d...)

	v.state = attr.ValueStateKnown
	return diags
}

func (v *ProjectsValue) Flatten(ctx context.Context, p *rafay.V1ClusterSharingProject) diag.Diagnostics {
	var diags diag.Diagnostics
	if p == nil {
		return diags
	}

	v.Name = types.StringValue(p.Name)

	v.state = attr.ValueStateKnown
	return diags
}
