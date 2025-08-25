package resource_eks_cluster

import (
	"context"
	"log"

	"github.com/RafaySystems/terraform-provider-rafay/rafay"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	jsoniter "github.com/json-iterator/go"
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

	proxycfgMap := types.MapNull(types.StringType)
	if spec.ProxyConfig != nil {
		pc := map[string]attr.Value{}
		if len(spec.ProxyConfig.HttpProxy) > 0 {
			pc["http_proxy"] = types.StringValue(spec.ProxyConfig.HttpProxy)
		}
		if len(spec.ProxyConfig.HttpsProxy) > 0 {
			pc["https_proxy"] = types.StringValue(spec.ProxyConfig.HttpsProxy)
		}
		if len(spec.ProxyConfig.NoProxy) > 0 {
			pc["no_proxy"] = types.StringValue(spec.ProxyConfig.NoProxy)
		}
		if len(spec.ProxyConfig.ProxyAuth) > 0 {
			pc["proxy_auth"] = types.StringValue(spec.ProxyConfig.ProxyAuth)
		}
		if len(spec.ProxyConfig.BootstrapCA) > 0 {
			pc["bootstrap_ca"] = types.StringValue(spec.ProxyConfig.BootstrapCA)
		}
		if spec.ProxyConfig.Enabled {
			pc["enabled"] = types.StringValue("true")
		}
		if spec.ProxyConfig.AllowInsecureBootstrap {
			pc["allow_insecure_bootstrap"] = types.StringValue("true")
		}
	}
	v.ProxyConfig = proxycfgMap

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
		// TODO(Akshay): Update later
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

	azElements := []attr.Value{}
	for _, az := range cc.AvailabilityZones {
		azElements = append(azElements, types.StringValue(az))
	}
	v.AvailabilityZones, d = types.ListValue(types.StringType, azElements)
	diags = append(diags, d...)

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

	antsMap := types.MapNull(types.StringType)
	if len(md.Annotations) != 0 {
		ants := map[string]attr.Value{}
		for key, val := range md.Annotations {
			ants[key] = types.StringValue(val)
		}
		antsMap, d = types.MapValue(types.StringType, ants)
		diags = append(diags, d...)
	}
	v.Annotations = antsMap

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
	// TODO: SubnetCidr is missing
	v.ClusterDns = types.StringValue(ng.ClusterDNS)
	v.EbsOptimized = types.BoolValue(*ng.EBSOptimized)
	v.VolumeName = types.StringValue(ng.VolumeName)
	v.VolumeEncrypted = types.BoolValue(*ng.VolumeEncrypted)
	v.VolumeKmsKeyId = types.StringValue(ng.VolumeKmsKeyID)
	v.OverrideBootstrapCommand = types.StringValue(ng.OverrideBootstrapCommand)

	pbElements := []attr.Value{}
	for _, pb := range ng.PreBootstrapCommands {
		pbElements = append(pbElements, types.StringValue(pb))
	}
	v.PreBootstrapCommands, d = types.ListValue(types.StringType, pbElements)
	diags = append(diags, d...)

	aspElements := []attr.Value{}
	for _, asp := range ng.ASGSuspendProcesses {
		aspElements = append(aspElements, types.StringValue(asp))
	}
	v.AsgSuspendProcesses, d = types.ListValue(types.StringType, aspElements)
	diags = append(diags, d...)

	tgaElements := []attr.Value{}
	for _, tga := range ng.TargetGroupARNs {
		tgaElements = append(tgaElements, types.StringValue(tga))
	}
	v.TargetGroupArns, d = types.ListValue(types.StringType, tgaElements)
	diags = append(diags, d...)

	clbElements := []attr.Value{}
	for _, clb := range ng.ClassicLoadBalancerNames {
		clbElements = append(clbElements, types.StringValue(clb))
	}
	v.ClassicLoadBalancerNames, d = types.ListValue(types.StringType, clbElements)
	diags = append(diags, d...)

	v.CpuCredits = types.StringValue(ng.CPUCredits)
	v.EnableDetailedMonitoring = types.BoolValue(*ng.EnableDetailedMonitoring)
	v.InstanceType = types.StringValue(ng.InstanceType)

	azElements := []attr.Value{}
	for _, az := range ng.AvailabilityZones {
		azElements = append(azElements, types.StringValue(az))
	}
	v.AvailabilityZones2, d = types.ListValue(types.StringType, azElements)
	diags = append(diags, d...)

	snElements := []attr.Value{}
	for _, sn := range ng.Subnets {
		snElements = append(snElements, types.StringValue(sn))
	}
	v.Subnets, d = types.ListValue(types.StringType, snElements)
	diags = append(diags, d...)

	v.InstancePrefix = types.StringValue(ng.InstancePrefix)
	v.InstanceName = types.StringValue(ng.InstanceName)

	lbsMap := types.MapNull(types.StringType)
	if len(ng.Labels) != 0 {
		lbs := map[string]attr.Value{}
		for key, val := range ng.Labels {
			lbs[key] = types.StringValue(val)
		}
		lbsMap, d = types.MapValue(types.StringType, lbs)
		diags = append(diags, d...)
	}
	v.Labels2 = lbsMap

	tagMap := types.MapNull(types.StringType)
	if len(ng.Tags) != 0 {
		tag := map[string]attr.Value{}
		for key, val := range ng.Tags {
			tag[key] = types.StringValue(val)
		}
		tagMap, d = types.MapValue(types.StringType, tag)
		diags = append(diags, d...)
	}
	v.Tags2 = tagMap

	v.Ami = types.StringValue(ng.AMI)

	// security groups
	if ng.SecurityGroups != nil {
		sg := NewSecurityGroups2ValueNull()
		d = sg.Flatten(ctx, ng.SecurityGroups)
		diags = append(d...)
		sgElements := []attr.Value{sg}
		v.SecurityGroups2, d = types.ListValue(SecurityGroups2Value{}.Type(ctx), sgElements)
		diags = append(diags, d...)
	}

	iam := NewIamValueNull()
	d = iam.Flatten(ctx, ng.IAM)
	diags = append(diags, d...)
	iamElements := []attr.Value{iam}
	v.Iam, d = types.ListValue(IamValue{}.Type(ctx), iamElements)
	diags = append(diags, d...)

	v.state = attr.ValueStateKnown
	return diags
}

func (v *SecurityGroups2Value) Flatten(ctx context.Context, sg *rafay.NodeGroupSGs) diag.Diagnostics {
	var diags, d diag.Diagnostics
	if sg == nil {
		return diags
	}

	v.WithShared = types.BoolValue(*sg.WithShared)
	v.WithLocal = types.BoolValue(*sg.WithLocal)

	aidsElements := []attr.Value{}
	for _, aid := range sg.AttachIDs {
		aidsElements = append(aidsElements, types.StringValue(aid))
	}
	v.AttachIds, d = types.ListValue(types.StringType, aidsElements)
	diags = append(diags, d...)

	v.state = attr.ValueStateKnown
	return diags
}

func (v *IamValue) Flatten(ctx context.Context, iam *rafay.NodeGroupIAM) diag.Diagnostics {
	var diags, d diag.Diagnostics
	if iam == nil {
		return diags
	}

	// TODO(Akshay): Check if Attach Policy v2. based on that populate attach_policy and attach_policy_v2
	var json2 = jsoniter.ConfigCompatibleWithStandardLibrary
	jsonBytes, err := json2.Marshal(iam.AttachPolicy)
	if err != nil {
		diags.AddError("Attach Policy Marshal Error", err.Error())
	}
	v.AttachPolicyV2 = types.StringValue(string(jsonBytes))

	attachPolicy := NewAttachPolicyValueNull()
	d = attachPolicy.Flatten(ctx, iam.AttachPolicy)
	diags = append(diags, d...)
	v.AttachPolicy, d = types.ListValue(AttachPolicyValue{}.Type(ctx), []attr.Value{attachPolicy})
	diags = append(diags, d...)

	arns := []attr.Value{}
	for _, arn := range iam.AttachPolicyARNs {
		arns = append(arns, types.StringValue(arn))
	}
	v.AttachPolicyArns, d = types.ListValue(types.StringType, arns)
	diags = append(diags, d...)

	v.InstanceProfileArn = types.StringValue(iam.InstanceProfileARN)
	v.InstanceRoleArn = types.StringValue(iam.InstanceRoleARN)
	v.InstanceRoleName = types.StringValue(iam.InstanceRoleName)
	v.InstanceRolePermissionBoundary = types.StringValue(iam.InstanceRolePermissionsBoundary)

	addonPolicies := NewIamNodeGroupWithAddonPoliciesValueNull()
	d = addonPolicies.Flatten(ctx, iam.WithAddonPolicies)
	diags = append(diags, d...)
	addonPoliciesElements := []attr.Value{addonPolicies}
	v.IamNodeGroupWithAddonPolicies, d = types.ListValue(IamNodeGroupWithAddonPoliciesValue{}.Type(ctx), addonPoliciesElements)
	diags = append(diags, d...)

	v.state = attr.ValueStateKnown
	return diags
}

func (v *AttachPolicyValue) Flatten(ctx context.Context, attachpol *rafay.InlineDocument) diag.Diagnostics {
	var diags, d diag.Diagnostics
	if attachpol == nil {
		return diags
	}

	v.Version = types.StringValue(attachpol.Version)
	v.Id = types.StringValue(attachpol.Id)

	stms := []attr.Value{}
	for _, stm := range attachpol.Statement {
		sv := NewStatementValueNull()
		d = sv.Flatten(ctx, stm)
		diags = append(diags, d...)
		stms = append(stms, sv)
	}
	v.Statement, d = types.ListValue(StatementValue{}.Type(ctx), stms)
	diags = append(diags, d...)

	v.state = attr.ValueStateKnown
	return diags
}

func (v *StatementValue) Flatten(ctx context.Context, stm rafay.InlineStatement) diag.Diagnostics {
	var diags, d diag.Diagnostics

	if len(stm.Effect) > 0 {
		v.Effect = types.StringValue(stm.Effect)
	}
	if len(stm.Sid) > 0 {
		v.Sid = types.StringValue(stm.Sid)
	}
	if stm.Action != nil && len(stm.Action.([]interface{})) > 0 {
		actEle := []attr.Value{}
		for _, act := range stm.Action.([]interface{}) {
			actEle = append(actEle, types.StringValue(act.(string)))
		}
		v.Action, d = types.ListValue(types.StringType, actEle)
		diags = append(diags, d...)
	}
	if stm.NotAction != nil && len(stm.NotAction.([]interface{})) > 0 {
		naEle := []attr.Value{}
		for _, na := range stm.NotAction.([]interface{}) {
			naEle = append(naEle, types.StringValue(na.(string)))
		}
		v.NotAction, d = types.ListValue(types.StringType, naEle)
		diags = append(diags, d...)
	}
	if len(stm.Resource.(string)) > 0 {
		v.Resource = types.StringValue(stm.Resource.(string))
	}
	if stm.NotResource != nil && len(stm.NotResource.([]interface{})) > 0 {
		nrEle := []attr.Value{}
		for _, nr := range stm.NotResource.([]interface{}) {
			nrEle = append(nrEle, types.StringValue(nr.(string)))
		}
		v.NotResource, d = types.ListValue(types.StringType, nrEle)
		diags = append(diags, d...)
	}

	if len(stm.Condition) > 0 {
		var json2 = jsoniter.ConfigCompatibleWithStandardLibrary
		jsonStr, err := json2.Marshal(stm.Condition)
		if err != nil {
			diags.AddError("condition marshal error", err.Error())
		}
		v.Condition = types.StringValue(string(jsonStr))
	}

	if len(stm.Principal) > 0 {
		var json2 = jsoniter.ConfigCompatibleWithStandardLibrary
		jsonStr, err := json2.Marshal(stm.Principal)
		if err != nil {
			log.Println("attach policy marshal err:", err)
		}
		v.Principal = types.StringValue(string(jsonStr))

	}

	if len(stm.NotPrincipal) > 0 {
		var json2 = jsoniter.ConfigCompatibleWithStandardLibrary
		jsonStr, err := json2.Marshal(stm.NotPrincipal)
		if err != nil {
			log.Println("attach policy marshal err:", err)
		}
		v.NotPrincipal = types.StringValue(string(jsonStr))
	}

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
