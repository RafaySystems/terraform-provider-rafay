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

func (v *NodeGroupsValue) Flatten(ctx context.Context, in *rafay.NodeGroup, state NodeGroupsValue) diag.Diagnostics {
	var diags, d diag.Diagnostics

	v.Name = types.StringValue(in.Name)
	v.AmiFamily = types.StringValue(in.AMIFamily)
	if in.DesiredCapacity != nil {
		v.DesiredCapacity = types.Int64Value(int64(*in.DesiredCapacity))
	}
	v.DisableImdsv1 = types.BoolPointerValue(in.DisableIMDSv1)
	v.DisablePodsImds = types.BoolPointerValue(in.DisablePodIMDS)
	v.EfaEnabled = types.BoolPointerValue(in.EFAEnabled)
	v.InstanceType = types.StringValue(in.InstanceType)
	v.MaxPodsPerNode = types.Int64Value(int64(in.MaxPodsPerNode))
	if in.MaxSize != nil {
		v.MaxSize = types.Int64Value(int64(*in.MaxSize))
	}
	if in.MinSize != nil {
		v.MinSize = types.Int64Value(int64(*in.MinSize))
	}
	v.PrivateNetworking = types.BoolPointerValue(in.PrivateNetworking)
	v.Version = types.StringValue(in.Version)
	if in.VolumeIOPS != nil {
		v.VolumeIops = types.Int64Value(int64(*in.VolumeIOPS))
	}
	if in.VolumeSize != nil {
		v.VolumeSize = types.Int64Value(int64(*in.VolumeSize))
	}
	if in.VolumeThroughput != nil {
		v.VolumeThroughput = types.Int64Value(int64(*in.VolumeThroughput))
	}
	v.VolumeType = types.StringValue(in.VolumeType)
	// TODO: SubnetCidr is missing
	v.ClusterDns = types.StringValue(in.ClusterDNS)
	v.EbsOptimized = types.BoolPointerValue(in.EBSOptimized)
	v.VolumeName = types.StringValue(in.VolumeName)
	v.VolumeEncrypted = types.BoolPointerValue(in.VolumeEncrypted)
	v.VolumeKmsKeyId = types.StringValue(in.VolumeKmsKeyID)
	v.OverrideBootstrapCommand = types.StringValue(in.OverrideBootstrapCommand)

	preBootstrapCommands := types.ListNull(types.StringType)
	if len(in.PreBootstrapCommands) > 0 {
		pbElements := []attr.Value{}
		for _, pb := range in.PreBootstrapCommands {
			pbElements = append(pbElements, types.StringValue(pb))
		}
		preBootstrapCommands, d = types.ListValue(types.StringType, pbElements)
		diags = append(diags, d...)
	}
	v.PreBootstrapCommands = preBootstrapCommands

	asgSuspendProcess := types.ListNull(types.StringType)
	if len(in.ASGSuspendProcesses) > 0 {
		aspElements := []attr.Value{}
		for _, asp := range in.ASGSuspendProcesses {
			aspElements = append(aspElements, types.StringValue(asp))
		}
		asgSuspendProcess, d = types.ListValue(types.StringType, aspElements)
		diags = append(diags, d...)
	}
	v.AsgSuspendProcesses = asgSuspendProcess

	targetGroupArns := types.ListNull(types.StringType)
	if len(in.TargetGroupARNs) > 0 {
		tgaElements := []attr.Value{}
		for _, tga := range in.TargetGroupARNs {
			tgaElements = append(tgaElements, types.StringValue(tga))
		}
		targetGroupArns, d = types.ListValue(types.StringType, tgaElements)
		diags = append(diags, d...)
	}
	v.TargetGroupArns = targetGroupArns

	classicLoadBalancerNames := types.ListNull(types.StringType)
	if len(in.ClassicLoadBalancerNames) > 0 {
		clbElements := []attr.Value{}
		for _, clb := range in.ClassicLoadBalancerNames {
			clbElements = append(clbElements, types.StringValue(clb))
		}
		classicLoadBalancerNames, d = types.ListValue(types.StringType, clbElements)
		diags = append(diags, d...)
	}
	v.ClassicLoadBalancerNames = classicLoadBalancerNames

	v.CpuCredits = types.StringValue(in.CPUCredits)
	v.EnableDetailedMonitoring = types.BoolPointerValue(in.EnableDetailedMonitoring)
	v.InstanceType = types.StringValue(in.InstanceType)

	availabilityZones2 := types.ListNull(types.StringType)
	if len(in.AvailabilityZones) > 0 {
		azElements := []attr.Value{}
		for _, az := range in.AvailabilityZones {
			azElements = append(azElements, types.StringValue(az))
		}
		availabilityZones2, d = types.ListValue(types.StringType, azElements)
		diags = append(diags, d...)
	}
	v.AvailabilityZones2 = availabilityZones2

	subnets := types.ListNull(types.StringType)
	if len(in.Subnets) > 0 {
		snElements := []attr.Value{}
		for _, sn := range in.Subnets {
			snElements = append(snElements, types.StringValue(sn))
		}
		subnets, d = types.ListValue(types.StringType, snElements)
		diags = append(diags, d...)
	}
	v.Subnets = subnets

	v.InstancePrefix = types.StringValue(in.InstancePrefix)
	v.InstanceName = types.StringValue(in.InstanceName)

	lbsMap := types.MapNull(types.StringType)
	if len(in.Labels) != 0 {
		lbs := map[string]attr.Value{}
		for key, val := range in.Labels {
			lbs[key] = types.StringValue(val)
		}
		lbsMap, d = types.MapValue(types.StringType, lbs)
		diags = append(diags, d...)
	}
	v.Labels2 = lbsMap

	tagMap := types.MapNull(types.StringType)
	if len(in.Tags) != 0 {
		tag := map[string]attr.Value{}
		for key, val := range in.Tags {
			tag[key] = types.StringValue(val)
		}
		tagMap, d = types.MapValue(types.StringType, tag)
		diags = append(diags, d...)
	}
	v.Tags2 = tagMap

	v.Ami = types.StringValue(in.AMI)

	if in.IAM != nil {
		// get state iam
		stIams := make([]IamValue, 0, len(state.Iam.Elements()))
		d = state.Iam.ElementsAs(ctx, &stIams, false)
		diags = append(diags, d...)

		var stIam IamValue
		if len(stIams) > 0 {
			stIam = stIams[0]
		}

		iam := NewIamValueNull()
		d = iam.Flatten(ctx, in.IAM, stIam)
		diags = append(diags, d...)
		iamElements := []attr.Value{iam}
		v.Iam, d = types.ListValue(IamValue{}.Type(ctx), iamElements)
		diags = append(diags, d...)
	}

	ssh := NewSshValueNull()
	d = ssh.Flatten(ctx, in.SSH)
	diags = append(diags, d...)
	v.Ssh, d = types.ListValue(SshValue{}.Type(ctx), []attr.Value{ssh})
	diags = append(diags, d...)

	placement := NewPlacementValueNull()
	d = placement.Flatten(ctx, in.Placement)
	diags = append(diags, d...)
	v.Placement, d = types.ListValue(PlacementValue{}.Type(ctx), []attr.Value{placement})
	diags = append(diags, d...)

	instanceSel := NewInstanceSelectorValueNull()
	d = instanceSel.Flatten(ctx, in.InstanceSelector)
	diags = append(diags, d...)
	v.InstanceSelector, d = types.ListValue(InstanceSelectorValue{}.Type(ctx), []attr.Value{instanceSel})
	diags = append(diags, d...)

	bottlerkt := NewBottleRocketValueNull()
	d = bottlerkt.Flatten(ctx, in.Bottlerocket)
	diags = append(diags, d...)
	v.BottleRocket, d = types.ListValue(BottleRocketValue{}.Type(ctx), []attr.Value{bottlerkt})
	diags = append(diags, d...)

	instDistribution := NewInstancesDistributionValueNull()
	d = instDistribution.Flatten(ctx, in.InstancesDistribution)
	diags = append(diags, d...)
	v.InstancesDistribution, d = types.ListValue(InstancesDistributionValue{}.Type(ctx), []attr.Value{instDistribution})
	diags = append(diags, d...)

	asgMetricsCollection := types.ListNull(AsgMetricsCollectionValue{}.Type(ctx))
	if len(in.ASGMetricsCollection) > 0 {
		amcList := []attr.Value{}
		for _, val := range in.ASGMetricsCollection {
			amc := NewAsgMetricsCollectionValueNull()
			d = amc.Flatten(ctx, val)
			diags = append(diags, d...)
			amcList = append(amcList, amc)
		}
		asgMetricsCollection, d = types.ListValue(AsgMetricsCollectionValue{}.Type(ctx), amcList)
		diags = append(diags, d...)
	}
	v.AsgMetricsCollection = asgMetricsCollection

	taints := types.ListNull(TaintsValue{}.Type(ctx))
	if len(in.Taints) > 0 {
		taintsList := []attr.Value{}
		for _, val := range in.Taints {
			taint := NewTaintsValueNull()
			d = taint.Flatten(ctx, val)
			diags = append(diags, d...)
			taintsList = append(taintsList, taint)
		}
		taints, d = types.ListValue(TaintsValue{}.Type(ctx), taintsList)
		diags = append(diags, d...)
	}
	v.Taints = taints

	updateConfig := NewUpdateConfigValueNull()
	d = updateConfig.Flatten(ctx, in.UpdateConfig)
	diags = append(diags, d...)
	v.UpdateConfig, d = types.ListValue(UpdateConfigValue{}.Type(ctx), []attr.Value{updateConfig})
	diags = append(diags, d...)

	kubeletExtraConfig := NewKubeletExtraConfigValueNull()
	d = kubeletExtraConfig.Flatten(ctx, in.KubeletExtraConfig)
	diags = append(diags, d...)
	v.KubeletExtraConfig, d = types.ListValue(KubeletExtraConfigValue{}.Type(ctx), []attr.Value{kubeletExtraConfig})
	diags = append(diags, d...)

	secGroup := NewSecurityGroups2ValueNull()
	d = secGroup.Flatten(ctx, in.SecurityGroups)
	diags = append(diags, d...)
	v.SecurityGroups2, d = types.ListValue(SecurityGroups2Value{}.Type(ctx), []attr.Value{secGroup})
	diags = append(diags, d...)

	v.state = attr.ValueStateKnown
	return diags
}

func (v *KubeletExtraConfigValue) Flatten(ctx context.Context, in *rafay.KubeletExtraConfig) diag.Diagnostics {
	var diags, d diag.Diagnostics
	if in == nil {
		return diags
	}

	kubeReserved := types.MapNull(types.StringType)
	if len(in.KubeReserved) > 0 {
		kr := map[string]attr.Value{}
		for key, val := range in.KubeReserved {
			kr[key] = types.StringValue(val)
		}
		kubeReserved, d = types.MapValue(types.StringType, kr)
		diags = append(diags, d...)
	}
	v.KubeReserved = kubeReserved
	v.KubeReservedCgroup = types.StringValue(in.KubeReservedCGroup)

	sysReserved := types.MapNull(types.StringType)
	if len(in.SystemReserved) > 0 {
		sr := map[string]attr.Value{}
		for key, val := range in.SystemReserved {
			sr[key] = types.StringValue(val)
		}
		sysReserved, d = types.MapValue(types.StringType, sr)
		diags = append(diags, d...)
	}
	v.SystemReserved = sysReserved

	evictionHard := types.MapNull(types.StringType)
	if len(in.EvictionHard) > 0 {
		eh := map[string]attr.Value{}
		for key, val := range in.EvictionHard {
			eh[key] = types.StringValue(val)
		}
		evictionHard, d = types.MapValue(types.StringType, eh)
		diags = append(diags, d...)
	}
	v.EvictionHard = evictionHard

	featureGates := types.MapNull(types.BoolType)
	if len(in.FeatureGates) > 0 {
		fg := map[string]attr.Value{}
		for key, val := range in.FeatureGates {
			fg[key] = types.BoolValue(val)
		}
		featureGates, d = types.MapValue(types.BoolType, fg)
		diags = append(diags, d...)
	}
	v.FeatureGates = featureGates

	v.state = attr.ValueStateKnown
	return diags
}

func (v *UpdateConfigValue) Flatten(ctx context.Context, in *rafay.NodeGroupUpdateConfig) diag.Diagnostics {
	var diags diag.Diagnostics
	if in == nil {
		return diags
	}

	if in.MaxUnavailable != nil {
		v.MaxUnavaliable = types.Int64Value(int64(*in.MaxUnavailable))
	}
	if in.MaxUnavailablePercentage != nil {
		v.MaxUnavaliablePercetage = types.Int64Value(int64(*in.MaxUnavailablePercentage))
	}

	v.state = attr.ValueStateKnown
	return diags
}

func (v *TaintsValue) Flatten(ctx context.Context, in rafay.NodeGroupTaint) diag.Diagnostics {
	var diags diag.Diagnostics

	v.Key = types.StringValue(in.Key)
	v.Value = types.StringValue(in.Value)
	v.Effect = types.StringValue(in.Effect)

	v.state = attr.ValueStateKnown
	return diags
}

func (v *AsgMetricsCollectionValue) Flatten(ctx context.Context, in rafay.MetricsCollection) diag.Diagnostics {
	var diags, d diag.Diagnostics

	v.Granularity = types.StringValue(in.Granularity)

	metrics := types.ListNull(types.StringType)
	if len(in.Metrics) > 0 {
		metricsList := []attr.Value{}
		for _, metric := range in.Metrics {
			metricsList = append(metricsList, types.StringValue(metric))
		}
		metrics, d = types.ListValue(types.StringType, metricsList)
		diags = append(diags, d...)
	}
	v.Metrics = metrics

	v.state = attr.ValueStateKnown
	return diags
}

func (v *InstancesDistributionValue) Flatten(ctx context.Context, in *rafay.NodeGroupInstancesDistribution) diag.Diagnostics {
	var diags, d diag.Diagnostics
	if in == nil {
		return diags
	}

	instanceTypes := types.ListNull(types.StringType)
	if len(in.InstanceTypes) > 0 {
		instanceTypesList := []attr.Value{}
		for _, it := range in.InstanceTypes {
			instanceTypesList = append(instanceTypesList, types.StringValue(it))
		}
		instanceTypes, d = types.ListValue(types.StringType, instanceTypesList)
		diags = append(diags, d...)
	}
	v.InstanceTypes = instanceTypes

	v.MaxPrice = types.Float64PointerValue(in.MaxPrice)
	if in.OnDemandBaseCapacity != nil {
		v.OnDemandBaseCapacity = types.Int64Value(int64(*in.OnDemandBaseCapacity))
	}

	if in.OnDemandPercentageAboveBaseCapacity != nil {
		v.OnDemandPercentageAboveBaseCapacity = types.Int64Value(int64(*in.OnDemandPercentageAboveBaseCapacity))
	}
	if in.SpotInstancePools != nil {
		v.SpotInstancePools = types.Int64Value(int64(*in.SpotInstancePools))
	}
	v.SpotAllocationStrategy = types.StringValue(in.SpotAllocationStrategy)
	v.CapacityRebalance = types.BoolPointerValue(in.CapacityRebalance)

	v.state = attr.ValueStateKnown
	return diags
}

func (v *BottleRocketValue) Flatten(ctx context.Context, in *rafay.NodeGroupBottlerocket) diag.Diagnostics {
	var diags diag.Diagnostics
	if in == nil {
		return diags
	}

	v.EnableAdminContainer = types.BoolPointerValue(in.EnableAdminContainer)

	if in.Settings != nil && len(in.Settings) > 0 {
		var json2 = jsoniter.ConfigCompatibleWithStandardLibrary
		jsonStr, err := json2.Marshal(in.Settings)
		if err != nil {
			diags.AddError("Bottle rocket marshal error", err.Error())
		}
		v.Settings = types.StringValue(string(jsonStr))
	}

	v.state = attr.ValueStateKnown
	return diags
}

func (v *InstanceSelectorValue) Flatten(ctx context.Context, in *rafay.InstanceSelector) diag.Diagnostics {
	var diags diag.Diagnostics

	if in == nil {
		return diags
	}

	if in.VCPUs != nil {
		v.Vcpus = types.Int64Value(int64(*in.VCPUs))
	}
	v.Memory = types.StringValue(in.Memory)

	if in.GPUs != nil {
		v.Gpus = types.Int64Value(int64(*in.GPUs))
	}
	v.CpuArchitecture = types.StringValue(in.CPUArchitecture)

	v.state = attr.ValueStateKnown
	return diags
}

func (v *PlacementValue) Flatten(ctx context.Context, in *rafay.Placement) diag.Diagnostics {
	var diags diag.Diagnostics

	if in == nil {
		return diags
	}

	v.Group = types.StringValue(in.GroupName)

	v.state = attr.ValueStateKnown
	return diags
}

func (v *SshValue) Flatten(ctx context.Context, in *rafay.NodeGroupSSH) diag.Diagnostics {
	var diags, d diag.Diagnostics
	if in == nil {
		return diags
	}

	v.Allow = types.BoolPointerValue(in.Allow)
	v.PublicKey = types.StringValue(in.PublicKey)
	v.PublicKeyName = types.StringValue(in.PublicKeyName)

	sourceSecurityGroupIds := types.ListNull(types.StringType)
	if len(in.SourceSecurityGroupIDs) > 0 {
		ids := []attr.Value{}
		for _, id := range in.SourceSecurityGroupIDs {
			ids = append(ids, types.StringValue(id))
		}
		sourceSecurityGroupIds, d = types.ListValue(types.StringType, ids)
		diags = append(diags, d...)
	}
	v.SourceSecurityGroupIds = sourceSecurityGroupIds

	v.EnableSsm = types.BoolPointerValue(in.EnableSSM)

	v.state = attr.ValueStateKnown
	return diags
}

func (v *SecurityGroups2Value) Flatten(ctx context.Context, in *rafay.NodeGroupSGs) diag.Diagnostics {
	var diags, d diag.Diagnostics
	if in == nil {
		return diags
	}

	v.WithShared = types.BoolPointerValue(in.WithShared)
	v.WithLocal = types.BoolPointerValue(in.WithLocal)

	attachIds := types.ListNull(types.StringType)
	if len(in.AttachIDs) > 0 {
		aidsElements := []attr.Value{}
		for _, aid := range in.AttachIDs {
			aidsElements = append(aidsElements, types.StringValue(aid))
		}
		attachIds, d = types.ListValue(types.StringType, aidsElements)
		diags = append(diags, d...)
	}
	v.AttachIds = attachIds

	v.state = attr.ValueStateKnown
	return diags
}

func (v *IamValue) Flatten(ctx context.Context, in *rafay.NodeGroupIAM, state IamValue) diag.Diagnostics {
	var diags, d diag.Diagnostics
	if in == nil {
		return diags
	}

	var isPolicyV1, isPolicyV2 bool
	if !state.IsNull() && !state.AttachPolicyV2.IsNull() && !state.AttachPolicyV2.IsUnknown() &&
		getStringValue(state.AttachPolicyV2) != "" {
		isPolicyV2 = true
	}
	if !state.IsNull() && !state.AttachPolicy.IsNull() && !state.AttachPolicy.IsUnknown() {
		isPolicyV1 = true
	}

	if in.AttachPolicy != nil {
		if isPolicyV1 && !isPolicyV2 {
			attachPolicy := NewAttachPolicyValueNull()
			d = attachPolicy.Flatten(ctx, in.AttachPolicy)
			diags = append(diags, d...)
			v.AttachPolicy, d = types.ListValue(AttachPolicyValue{}.Type(ctx), []attr.Value{attachPolicy})
			diags = append(diags, d...)
		} else {
			var json2 = jsoniter.ConfigCompatibleWithStandardLibrary
			jsonBytes, err := json2.Marshal(in.AttachPolicy)
			if err != nil {
				diags.AddError("Attach Policy Marshal Error", err.Error())
			}
			v.AttachPolicyV2 = types.StringValue(string(jsonBytes))
		}
	}

	attachPolicyArns := types.ListNull(types.StringType)
	if len(in.AttachPolicyARNs) > 0 {
		arns := []attr.Value{}
		for _, arn := range in.AttachPolicyARNs {
			arns = append(arns, types.StringValue(arn))
		}
		attachPolicyArns, d = types.ListValue(types.StringType, arns)
		diags = append(diags, d...)
	}
	v.AttachPolicyArns = attachPolicyArns

	v.InstanceProfileArn = types.StringValue(in.InstanceProfileARN)
	v.InstanceRoleArn = types.StringValue(in.InstanceRoleARN)
	v.InstanceRoleName = types.StringValue(in.InstanceRoleName)
	v.InstanceRolePermissionBoundary = types.StringValue(in.InstanceRolePermissionsBoundary)

	addonPolicies := NewIamNodeGroupWithAddonPoliciesValueNull()
	d = addonPolicies.Flatten(ctx, in.WithAddonPolicies)
	diags = append(diags, d...)
	addonPoliciesElements := []attr.Value{addonPolicies}
	v.IamNodeGroupWithAddonPolicies, d = types.ListValue(IamNodeGroupWithAddonPoliciesValue{}.Type(ctx), addonPoliciesElements)
	diags = append(diags, d...)

	v.state = attr.ValueStateKnown
	return diags
}

func (v *AttachPolicyValue) Flatten(ctx context.Context, in *rafay.InlineDocument) diag.Diagnostics {
	var diags, d diag.Diagnostics
	if in == nil {
		return diags
	}

	v.Version = types.StringValue(in.Version)
	v.Id = types.StringValue(in.Id)

	statement := types.ListNull(StatementValue{}.Type(ctx))
	if len(in.Statement) > 0 {
		stms := []attr.Value{}
		for _, stm := range in.Statement {
			sv := NewStatementValueNull()
			d = sv.Flatten(ctx, stm)
			diags = append(diags, d...)
			stms = append(stms, sv)
		}
		statement, d = types.ListValue(StatementValue{}.Type(ctx), stms)
		diags = append(diags, d...)
	}
	v.Statement = statement

	v.state = attr.ValueStateKnown
	return diags
}

func (v *StatementValue) Flatten(ctx context.Context, in rafay.InlineStatement) diag.Diagnostics {
	var diags, d diag.Diagnostics

	if len(in.Effect) > 0 {
		v.Effect = types.StringValue(in.Effect)
	}
	if len(in.Sid) > 0 {
		v.Sid = types.StringValue(in.Sid)
	}
	if in.Action != nil && len(in.Action.([]interface{})) > 0 {
		actEle := []attr.Value{}
		for _, act := range in.Action.([]interface{}) {
			actEle = append(actEle, types.StringValue(act.(string)))
		}
		v.Action, d = types.ListValue(types.StringType, actEle)
		diags = append(diags, d...)
	}
	if in.NotAction != nil && len(in.NotAction.([]interface{})) > 0 {
		naEle := []attr.Value{}
		for _, na := range in.NotAction.([]interface{}) {
			naEle = append(naEle, types.StringValue(na.(string)))
		}
		v.NotAction, d = types.ListValue(types.StringType, naEle)
		diags = append(diags, d...)
	}
	if len(in.Resource.(string)) > 0 {
		v.Resource = types.StringValue(in.Resource.(string))
	}
	if in.NotResource != nil && len(in.NotResource.([]interface{})) > 0 {
		nrEle := []attr.Value{}
		for _, nr := range in.NotResource.([]interface{}) {
			nrEle = append(nrEle, types.StringValue(nr.(string)))
		}
		v.NotResource, d = types.ListValue(types.StringType, nrEle)
		diags = append(diags, d...)
	}

	if len(in.Condition) > 0 {
		var json2 = jsoniter.ConfigCompatibleWithStandardLibrary
		jsonStr, err := json2.Marshal(in.Condition)
		if err != nil {
			diags.AddError("condition marshal error", err.Error())
		}
		v.Condition = types.StringValue(string(jsonStr))
	}

	if len(in.Principal) > 0 {
		var json2 = jsoniter.ConfigCompatibleWithStandardLibrary
		jsonStr, err := json2.Marshal(in.Principal)
		if err != nil {
			log.Println("attach policy marshal err:", err)
		}
		v.Principal = types.StringValue(string(jsonStr))

	}

	if len(in.NotPrincipal) > 0 {
		var json2 = jsoniter.ConfigCompatibleWithStandardLibrary
		jsonStr, err := json2.Marshal(in.NotPrincipal)
		if err != nil {
			log.Println("attach policy marshal err:", err)
		}
		v.NotPrincipal = types.StringValue(string(jsonStr))
	}

	v.state = attr.ValueStateKnown
	return diags
}

func (v *IamNodeGroupWithAddonPoliciesValue) Flatten(ctx context.Context, in *rafay.NodeGroupIAMAddonPolicies) diag.Diagnostics {
	var diags diag.Diagnostics
	if in == nil {
		return diags
	}

	v.AlbIngress = types.BoolPointerValue(in.AWSLoadBalancerController)
	v.AppMesh = types.BoolPointerValue(in.AppMesh)
	v.AppMeshReview = types.BoolPointerValue(in.AppMeshPreview)
	v.AutoScaler = types.BoolPointerValue(in.AutoScaler)
	v.CertManager = types.BoolPointerValue(in.CertManager)
	v.CloudWatch = types.BoolPointerValue(in.CloudWatch)
	v.Ebs = types.BoolPointerValue(in.EBS)
	v.Efs = types.BoolPointerValue(in.EFS)
	v.ExternalDns = types.BoolPointerValue(in.ExternalDNS)
	v.Fsx = types.BoolPointerValue(in.FSX)
	v.ImageBuilder = types.BoolPointerValue(in.ImageBuilder)
	v.Xray = types.BoolPointerValue(in.XRay)

	v.state = attr.ValueStateKnown
	return diags
}
