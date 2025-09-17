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

func (v *NodeGroupsMapValue) Flatten(ctx context.Context, in *rafay.NodeGroup, state NodeGroupsMapValue) diag.Diagnostics {
	var diags, d diag.Diagnostics
	if in == nil {
		return diags
	}

	if in.AMIFamily != "" {
		v.AmiFamily = types.StringValue(in.AMIFamily)
	}
	if in.DesiredCapacity != nil {
		v.DesiredCapacity = types.Int64Value(int64(*in.DesiredCapacity))
	}
	if in.DisableIMDSv1 != nil {
		v.DisableImdsv1 = types.BoolPointerValue(in.DisableIMDSv1)
	}
	if in.DisablePodIMDS != nil {
		v.DisablePodsImds = types.BoolPointerValue(in.DisablePodIMDS)
	}
	if in.EFAEnabled != nil {
		v.EfaEnabled = types.BoolPointerValue(in.EFAEnabled)
	}
	if in.InstanceType != "" {
		v.InstanceType = types.StringValue(in.InstanceType)
	}
	v.MaxPodsPerNode = types.Int64Value(int64(in.MaxPodsPerNode))
	if in.MaxSize != nil {
		v.MaxSize = types.Int64Value(int64(*in.MaxSize))
	}
	if in.MinSize != nil {
		v.MinSize = types.Int64Value(int64(*in.MinSize))
	}
	if in.PrivateNetworking != nil {
		v.PrivateNetworking = types.BoolPointerValue(in.PrivateNetworking)
	}
	if in.Version != "" {
		v.Version = types.StringValue(in.Version)
	}
	if in.VolumeIOPS != nil {
		v.VolumeIops = types.Int64Value(int64(*in.VolumeIOPS))
	}
	if in.VolumeSize != nil {
		v.VolumeSize = types.Int64Value(int64(*in.VolumeSize))
	}
	if in.VolumeThroughput != nil {
		v.VolumeThroughput = types.Int64Value(int64(*in.VolumeThroughput))
	}
	if in.VolumeType != "" {
		v.VolumeType = types.StringValue(in.VolumeType)
	}
	if in.ClusterDNS != "" {
		v.ClusterDns = types.StringValue(in.ClusterDNS)
	}
	if in.EBSOptimized != nil {
		v.EbsOptimized = types.BoolPointerValue(in.EBSOptimized)
	}
	if in.VolumeName != "" {
		v.VolumeName = types.StringValue(in.VolumeName)
	}
	if in.VolumeEncrypted != nil {
		v.VolumeEncrypted = types.BoolPointerValue(in.VolumeEncrypted)
	}
	if in.VolumeKmsKeyID != "" {
		v.VolumeKmsKeyId = types.StringValue(in.VolumeKmsKeyID)
	}
	if in.OverrideBootstrapCommand != "" {
		v.OverrideBootstrapCommand = types.StringValue(in.OverrideBootstrapCommand)
	}

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

	if in.CPUCredits != "" {
		v.CpuCredits = types.StringValue(in.CPUCredits)
	}
	if in.EnableDetailedMonitoring != nil {
		v.EnableDetailedMonitoring = types.BoolPointerValue(in.EnableDetailedMonitoring)
	}
	if in.InstanceType != "" {
		v.InstanceType = types.StringValue(in.InstanceType)
	}

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

	subnets := types.SetNull(types.StringType)
	if len(in.Subnets) > 0 {
		snElements := []attr.Value{}
		for _, sn := range in.Subnets {
			snElements = append(snElements, types.StringValue(sn))
		}
		subnets, d = types.SetValue(types.StringType, snElements)
		diags = append(diags, d...)
	}
	v.Subnets = subnets

	if in.InstancePrefix != "" {
		v.InstancePrefix = types.StringValue(in.InstancePrefix)
	}
	if in.InstanceName != "" {
		v.InstanceName = types.StringValue(in.InstanceName)
	}

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

	if in.AMI != "" {
		v.Ami = types.StringValue(in.AMI)
	}

	if in.IAM != nil {
		// get state iam
		var stIam Iam6Value
		stIamObj, d := Iam6Type{}.ValueFromObject(ctx, state.Iam6)
		diags = append(diags, d...)
		if !d.HasError() {
			stIam = stIamObj.(Iam6Value)
		}

		iam := NewIam6ValueNull()
		d = iam.Flatten(ctx, in.IAM, stIam)
		diags = append(diags, d...)
		v.Iam6, d = iam.ToObjectValue(ctx)
		diags = append(diags, d...)
	} else {
		v.Iam6, d = NewIam6ValueNull().ToObjectValue(ctx)
		diags = append(diags, d...)
	}

	if in.SSH != nil {
		ssh := NewSsh6ValueNull()
		d = ssh.Flatten(ctx, in.SSH)
		diags = append(diags, d...)
		v.Ssh6, d = ssh.ToObjectValue(ctx)
		diags = append(diags, d...)
	} else {
		v.Ssh6, d = NewSsh6ValueNull().ToObjectValue(ctx)
		diags = append(diags, d...)
	}

	if in.Placement != nil {
		placement := NewPlacement6ValueNull()
		d = placement.Flatten(ctx, in.Placement)
		diags = append(diags, d...)
		v.Placement6, d = placement.ToObjectValue(ctx)
		diags = append(diags, d...)
	} else {
		v.Placement6, d = NewPlacement6ValueNull().ToObjectValue(ctx)
		diags = append(diags, d...)
	}

	if in.InstanceSelector != nil {
		instanceSel := NewInstanceSelector6ValueNull()
		d = instanceSel.Flatten(ctx, in.InstanceSelector)
		diags = append(diags, d...)
		v.InstanceSelector6, d = instanceSel.ToObjectValue(ctx)
		diags = append(diags, d...)
	} else {
		v.InstanceSelector6, d = NewInstanceSelector6ValueNull().ToObjectValue(ctx)
		diags = append(diags, d...)
	}

	if in.Bottlerocket != nil {
		bottlerkt := NewBottleRocket6ValueNull()
		d = bottlerkt.Flatten(ctx, in.Bottlerocket)
		diags = append(diags, d...)
		v.BottleRocket6, d = bottlerkt.ToObjectValue(ctx)
		diags = append(diags, d...)
	} else {
		v.BottleRocket6, d = NewBottleRocket6ValueNull().ToObjectValue(ctx)
		diags = append(diags, d...)
	}

	if in.InstancesDistribution != nil {
		instDistribution := NewInstancesDistribution6ValueNull()
		d = instDistribution.Flatten(ctx, in.InstancesDistribution)
		diags = append(diags, d...)
		v.InstancesDistribution6, d = instDistribution.ToObjectValue(ctx)
		diags = append(diags, d...)
	} else {
		v.InstancesDistribution6, d = NewInstancesDistribution6ValueNull().ToObjectValue(ctx)
		diags = append(diags, d...)
	}

	if len(in.ASGMetricsCollection) > 0 {
		amcList := []attr.Value{}
		for _, val := range in.ASGMetricsCollection {
			amc := NewAsgMetricsCollection6ValueNull()
			d = amc.Flatten(ctx, val)
			diags = append(diags, d...)
			amcList = append(amcList, amc)
		}
		v.AsgMetricsCollection6, d = types.SetValue(AsgMetricsCollectionValue{}.Type(ctx), amcList)
		diags = append(diags, d...)
	} else {
		v.AsgMetricsCollection6 = types.SetNull(AsgMetricsCollection6Value{}.Type(ctx))
	}

	if len(in.Taints) > 0 {
		taintsList := []attr.Value{}
		for _, val := range in.Taints {
			taint := NewTaints6ValueNull()
			d = taint.Flatten(ctx, val)
			diags = append(diags, d...)
			taintsList = append(taintsList, taint)
		}
		v.Taints6, d = types.SetValue(TaintsValue{}.Type(ctx), taintsList)
		diags = append(diags, d...)
	} else {
		v.Taints6 = types.SetNull(TaintsValue{}.Type(ctx))
	}

	if in.UpdateConfig != nil {
		updateConfig := NewUpdateConfig6ValueNull()
		d = updateConfig.Flatten(ctx, in.UpdateConfig)
		diags = append(diags, d...)
		v.UpdateConfig6, d = updateConfig.ToObjectValue(ctx)
		diags = append(diags, d...)
	} else {
		v.UpdateConfig6, d = NewUpdateConfig6ValueNull().ToObjectValue(ctx)
		diags = append(diags, d...)
	}

	if in.KubeletExtraConfig != nil {
		kubeletExtraConfig := NewKubeletExtraConfig6ValueNull()
		d = kubeletExtraConfig.Flatten(ctx, in.KubeletExtraConfig)
		diags = append(diags, d...)
		v.KubeletExtraConfig6, d = kubeletExtraConfig.ToObjectValue(ctx)
		diags = append(diags, d...)
	} else {
		v.KubeletExtraConfig6, d = NewKubeletExtraConfig6ValueNull().ToObjectValue(ctx)
		diags = append(diags, d...)
	}

	if in.SecurityGroups != nil {
		secGroup := NewSecurityGroups6ValueNull()
		d = secGroup.Flatten(ctx, in.SecurityGroups)
		diags = append(diags, d...)
		v.SecurityGroups6, d = secGroup.ToObjectValue(ctx)
		diags = append(diags, d...)
	} else {
		v.SecurityGroups6, d = NewSecurityGroups6ValueNull().ToObjectValue(ctx)
		diags = append(diags, d...)
	}

	v.state = attr.ValueStateKnown
	return diags
}

func (v *Iam6Value) Flatten(ctx context.Context, in *rafay.NodeGroupIAM, state Iam6Value) diag.Diagnostics {
	var diags, d diag.Diagnostics
	if in == nil {
		return diags
	}

	var isPolicyV1, isPolicyV2 bool
	if !state.IsNull() && !state.AttachPolicyV2.IsNull() && !state.AttachPolicyV2.IsUnknown() &&
		getStringValue(state.AttachPolicyV2) != "" {
		isPolicyV2 = true
	}
	if !state.IsNull() && !state.AttachPolicy6.IsNull() && !state.AttachPolicy6.IsUnknown() {
		isPolicyV1 = true
	}

	if in.AttachPolicy != nil {
		if isPolicyV1 && !isPolicyV2 {
			attachPolicy := NewAttachPolicy6ValueNull()
			d := attachPolicy.Flatten(ctx, in.AttachPolicy)
			diags = append(diags, d...)
			v.AttachPolicy6, d = attachPolicy.ToObjectValue(ctx)
			diags = append(diags, d...)
		} else {
			var json2 = jsoniter.ConfigCompatibleWithStandardLibrary
			jsonBytes, err := json2.Marshal(in.AttachPolicy)
			if err != nil {
				diags.AddError("Attach Policy Marshal Error", err.Error())
			}
			v.AttachPolicyV2 = types.StringValue(string(jsonBytes))
		}
	} else {
		v.AttachPolicy6, d = NewAttachPolicy6ValueNull().ToObjectValue(ctx)
		diags = append(diags, d...)
		v.AttachPolicyV2 = types.StringNull()
	}

	if len(in.AttachPolicyARNs) > 0 {
		aparns := []attr.Value{}
		for _, arn := range in.AttachPolicyARNs {
			aparns = append(aparns, types.StringValue(arn))
		}
		v.AttachPolicyArns, d = types.ListValue(types.StringType, aparns)
		diags = append(diags, d...)
	} else {
		v.AttachPolicyArns = types.ListNull(types.StringType)
	}

	if in.InstanceProfileARN != "" {
		v.InstanceProfileArn = types.StringValue(in.InstanceProfileARN)
	}
	if in.InstanceRoleARN != "" {
		v.InstanceRoleArn = types.StringValue(in.InstanceRoleARN)
	}
	if in.InstanceRoleName != "" {
		v.InstanceRoleName = types.StringValue(in.InstanceRoleName)
	}
	if in.InstanceRolePermissionsBoundary != "" {
		v.InstanceRolePermissionBoundary = types.StringValue(in.InstanceRolePermissionsBoundary)
	}
	if in.WithAddonPolicies != nil {
		addonPolicies := NewIamNodeGroupWithAddonPolicies6ValueNull()
		d = addonPolicies.Flatten(ctx, in.WithAddonPolicies)
		diags = append(diags, d...)
		v.IamNodeGroupWithAddonPolicies6, d = addonPolicies.ToObjectValue(ctx)
		diags = append(diags, d...)
	} else {
		v.IamNodeGroupWithAddonPolicies6, d = NewIamNodeGroupWithAddonPolicies6ValueNull().ToObjectValue(ctx)
		diags = append(diags, d...)
	}

	v.state = attr.ValueStateKnown
	return diags
}
func (v *IamNodeGroupWithAddonPolicies6Value) Flatten(ctx context.Context, in *rafay.NodeGroupIAMAddonPolicies) diag.Diagnostics {
	var diags diag.Diagnostics
	if in == nil {
		return diags
	}
	if in.AWSLoadBalancerController != nil {
		v.AlbIngress = types.BoolPointerValue(in.AWSLoadBalancerController)
	}
	if in.AppMesh != nil {
		v.AppMesh = types.BoolPointerValue(in.AppMesh)
	}
	if in.AppMeshPreview != nil {
		v.AppMeshReview = types.BoolPointerValue(in.AppMeshPreview)
	}
	if in.AutoScaler != nil {
		v.AutoScaler = types.BoolPointerValue(in.AutoScaler)
	}
	if in.CertManager != nil {
		v.CertManager = types.BoolPointerValue(in.CertManager)
	}
	if in.CloudWatch != nil {
		v.CloudWatch = types.BoolPointerValue(in.CloudWatch)
	}
	if in.EBS != nil {
		v.Ebs = types.BoolPointerValue(in.EBS)
	}
	if in.EFS != nil {
		v.Efs = types.BoolPointerValue(in.EFS)
	}
	if in.ExternalDNS != nil {
		v.ExternalDns = types.BoolPointerValue(in.ExternalDNS)
	}
	if in.FSX != nil {
		v.Fsx = types.BoolPointerValue(in.FSX)
	}
	if in.ImageBuilder != nil {
		v.ImageBuilder = types.BoolPointerValue(in.ImageBuilder)
	}
	if in.XRay != nil {
		v.Xray = types.BoolPointerValue(in.XRay)
	}
	v.state = attr.ValueStateKnown
	return diags
}
func (v *AttachPolicy6Value) Flatten(ctx context.Context, in *rafay.InlineDocument) diag.Diagnostics {
	var diags, d diag.Diagnostics
	if in == nil {
		return diags
	}
	if len(in.Version) > 0 {
		v.Version = types.StringValue(in.Version)
	}
	if len(in.Id) > 0 {
		v.Id = types.StringValue(in.Id)
	}

	if len(in.Statement) > 0 {
		var tfStatements []attr.Value
		for _, stmt := range in.Statement {
			tfStatement := NewStatement6ValueNull()
			d = tfStatement.Flatten(ctx, stmt)
			diags = append(diags, d...)
			tfStatementObj, d := tfStatement.ToObjectValue(ctx)
			diags = append(diags, d...)
			tfStatements = append(tfStatements, tfStatementObj)
		}
		v.Statement6, d = types.SetValue(Statement6Value{}.Type(ctx), tfStatements)
	} else {
		v.Statement6 = types.SetNull(Statement6Value{}.Type(ctx))
	}

	v.state = attr.ValueStateKnown
	return diags
}
func (v *Statement6Value) Flatten(ctx context.Context, in rafay.InlineStatement) diag.Diagnostics {
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

func (v *Ssh6Value) Flatten(ctx context.Context, in *rafay.NodeGroupSSH) diag.Diagnostics {
	var diags, d diag.Diagnostics
	if in == nil {
		return diags
	}
	if in.Allow != nil {
		v.Allow = types.BoolPointerValue(in.Allow)
	}
	if in.PublicKey != "" {
		v.PublicKey = types.StringValue(in.PublicKey)
	}
	if in.PublicKeyName != "" {
		v.PublicKeyName = types.StringValue(in.PublicKeyName)
	}
	if len(in.SourceSecurityGroupIDs) > 0 {
		ids := []attr.Value{}
		for _, id := range in.SourceSecurityGroupIDs {
			ids = append(ids, types.StringValue(id))
		}
		v.SourceSecurityGroupIds, d = types.ListValue(types.StringType, ids)
		diags = append(diags, d...)
	} else {
		v.SourceSecurityGroupIds = types.ListNull(types.StringType)
	}
	if in.EnableSSM != nil {
		v.EnableSsm = types.BoolPointerValue(in.EnableSSM)
	}
	v.state = attr.ValueStateKnown
	return diags
}
func (v *Placement6Value) Flatten(ctx context.Context, in *rafay.Placement) diag.Diagnostics {
	var diags diag.Diagnostics
	if in == nil {
		return diags
	}
	if in.GroupName != "" {
		v.Group = types.StringValue(in.GroupName)
	}
	v.state = attr.ValueStateKnown
	return diags
}
func (v *InstanceSelector6Value) Flatten(ctx context.Context, in *rafay.InstanceSelector) diag.Diagnostics {
	var diags diag.Diagnostics
	if in == nil {
		return diags
	}
	if in.VCPUs != nil {
		v.Vcpus = types.Int64Value(int64(*in.VCPUs))
	}
	if in.Memory != "" {
		v.Memory = types.StringValue(in.Memory)
	}
	if in.GPUs != nil {
		v.Gpus = types.Int64Value(int64(*in.GPUs))
	}
	if in.CPUArchitecture != "" {
		v.CpuArchitecture = types.StringValue(in.CPUArchitecture)
	}
	v.state = attr.ValueStateKnown
	return diags
}
func (v *BottleRocket6Value) Flatten(ctx context.Context, in *rafay.NodeGroupBottlerocket) diag.Diagnostics {
	var diags diag.Diagnostics
	if in == nil {
		return diags
	}
	if in.EnableAdminContainer != nil {
		v.EnableAdminContainer = types.BoolPointerValue(in.EnableAdminContainer)
	}
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

func (v *InstancesDistribution6Value) Flatten(ctx context.Context, in *rafay.NodeGroupInstancesDistribution) diag.Diagnostics {
	var diags, d diag.Diagnostics
	if in == nil {
		return diags
	}

	if len(in.InstanceTypes) > 0 {
		instanceTypesList := []attr.Value{}
		for _, it := range in.InstanceTypes {
			instanceTypesList = append(instanceTypesList, types.StringValue(it))
		}
		v.InstanceTypes, d = types.ListValue(types.StringType, instanceTypesList)
		diags = append(diags, d...)
	} else {
		v.InstanceTypes = types.ListNull(types.StringType)
	}
	if in.MaxPrice != nil {
		v.MaxPrice = types.Float64PointerValue(in.MaxPrice)
	}
	if in.OnDemandBaseCapacity != nil {
		v.OnDemandBaseCapacity = types.Int64Value(int64(*in.OnDemandBaseCapacity))
	}
	if in.OnDemandPercentageAboveBaseCapacity != nil {
		v.OnDemandPercentageAboveBaseCapacity = types.Int64Value(int64(*in.OnDemandPercentageAboveBaseCapacity))
	}
	if in.SpotInstancePools != nil {
		v.SpotInstancePools = types.Int64Value(int64(*in.SpotInstancePools))
	}
	if in.SpotAllocationStrategy != "" {
		v.SpotAllocationStrategy = types.StringValue(in.SpotAllocationStrategy)
	}
	if in.CapacityRebalance != nil {
		v.CapacityRebalance = types.BoolPointerValue(in.CapacityRebalance)
	}
	v.state = attr.ValueStateKnown
	return diags
}

func (v *AsgMetricsCollection6Value) Flatten(ctx context.Context, in rafay.MetricsCollection) diag.Diagnostics {
	var diags, d diag.Diagnostics
	if in.Granularity != "" {
		v.Granularity = types.StringValue(in.Granularity)
	}
	if len(in.Metrics) > 0 {
		metricsList := []attr.Value{}
		for _, metric := range in.Metrics {
			metricsList = append(metricsList, types.StringValue(metric))
		}
		v.Metrics, d = types.ListValue(types.StringType, metricsList)
		diags = append(diags, d...)
	} else {
		v.Metrics = types.ListNull(types.StringType)
	}
	v.state = attr.ValueStateKnown
	return diags
}

func (v *Taints6Value) Flatten(ctx context.Context, in rafay.NodeGroupTaint) diag.Diagnostics {
	var diags diag.Diagnostics
	if in.Key != "" {
		v.Key = types.StringValue(in.Key)
	}
	if in.Value != "" {
		v.Value = types.StringValue(in.Value)
	}
	if in.Effect != "" {
		v.Effect = types.StringValue(in.Effect)
	}
	v.state = attr.ValueStateKnown
	return diags
}

func (v *UpdateConfig6Value) Flatten(ctx context.Context, in *rafay.NodeGroupUpdateConfig) diag.Diagnostics {
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

func (v *KubeletExtraConfig6Value) Flatten(ctx context.Context, in *rafay.KubeletExtraConfig) diag.Diagnostics {
	var diags, d diag.Diagnostics
	if in == nil {
		return diags
	}
	if in.KubeReserved != nil && len(in.KubeReserved) > 0 {
		kr := map[string]attr.Value{}
		for key, val := range in.KubeReserved {
			kr[key] = types.StringValue(val)
		}
		v.KubeReserved, d = types.MapValue(types.StringType, kr)
		diags = append(diags, d...)
	} else {
		v.KubeReserved = types.MapNull(types.StringType)
	}
	if in.KubeReservedCGroup != "" {
		v.KubeReservedCgroup = types.StringValue(in.KubeReservedCGroup)
	}

	if len(in.SystemReserved) > 0 {
		sr := map[string]attr.Value{}
		for key, val := range in.SystemReserved {
			sr[key] = types.StringValue(val)
		}
		v.SystemReserved, d = types.MapValue(types.StringType, sr)
		diags = append(diags, d...)
	} else {
		v.SystemReserved = types.MapNull(types.StringType)
	}

	if len(in.EvictionHard) > 0 {
		eh := map[string]attr.Value{}
		for key, val := range in.EvictionHard {
			eh[key] = types.StringValue(val)
		}
		v.EvictionHard, d = types.MapValue(types.StringType, eh)
		diags = append(diags, d...)
	} else {
		v.EvictionHard = types.MapNull(types.StringType)
	}

	if len(in.FeatureGates) > 0 {
		fg := map[string]attr.Value{}
		for key, val := range in.FeatureGates {
			fg[key] = types.BoolValue(val)
		}
		v.FeatureGates, d = types.MapValue(types.BoolType, fg)
		diags = append(diags, d...)
	} else {
		v.FeatureGates = types.MapNull(types.BoolType)
	}
	v.state = attr.ValueStateKnown
	return diags
}

func (v *SecurityGroups6Value) Flatten(ctx context.Context, in *rafay.NodeGroupSGs) diag.Diagnostics {
	var diags, d diag.Diagnostics
	if in == nil {
		return diags
	}
	if in.WithShared != nil {
		v.WithShared = types.BoolPointerValue(in.WithShared)
	}
	if in.WithLocal != nil {
		v.WithLocal = types.BoolPointerValue(in.WithLocal)
	}

	if len(in.AttachIDs) > 0 {
		aidsElements := []attr.Value{}
		for _, aid := range in.AttachIDs {
			aidsElements = append(aidsElements, types.StringValue(aid))
		}
		v.AttachIds, d = types.ListValue(types.StringType, aidsElements)
		diags = append(diags, d...)
	} else {
		v.AttachIds = types.ListNull(types.StringType)
	}

	v.state = attr.ValueStateKnown
	return diags
}
