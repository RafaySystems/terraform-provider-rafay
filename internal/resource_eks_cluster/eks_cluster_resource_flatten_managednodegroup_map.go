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

func (v *ManagedNodegroupsMapValue) Flatten(ctx context.Context, in *rafay.ManagedNodeGroup, state ManagedNodegroupsMapValue) diag.Diagnostics {
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
		v.DisableImdsv1 = types.BoolValue(*in.DisableIMDSv1)
	}
	if in.DisablePodIMDS != nil {
		v.DisablePodsImds = types.BoolValue(*in.DisablePodIMDS)
	}
	if in.EFAEnabled != nil {
		v.EfaEnabled = types.BoolValue(*in.EFAEnabled)
	}
	if in.InstanceType != "" {
		v.InstanceType = types.StringValue(in.InstanceType)
	}
	if in.MaxPodsPerNode != nil {
		v.MaxPodsPerNode = types.Int64Value(int64(*in.MaxPodsPerNode))
	}
	if in.MaxSize != nil {
		v.MaxSize = types.Int64Value(int64(*in.MaxSize))
	}
	if in.MinSize != nil {
		v.MinSize = types.Int64Value(int64(*in.MinSize))
	}
	if in.PrivateNetworking != nil {
		v.PrivateNetworking = types.BoolValue(*in.PrivateNetworking)
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
	if in.EBSOptimized != nil {
		v.EbsOptimized = types.BoolValue(*in.EBSOptimized)
	}
	if in.VolumeName != "" {
		v.VolumeName = types.StringValue(in.VolumeName)
	}
	if in.VolumeEncrypted != nil {
		v.VolumeEncrypted = types.BoolValue(*in.VolumeEncrypted)
	}
	if in.VolumeKmsKeyID != "" {
		v.VolumeKmsKeyId = types.StringValue(in.VolumeKmsKeyID)
	}
	if in.OverrideBootstrapCommand != "" {
		v.OverrideBootstrapCommand = types.StringValue(in.OverrideBootstrapCommand)
	}

	if len(in.PreBootstrapCommands) > 0 {
		pbElements := []attr.Value{}
		for _, pb := range in.PreBootstrapCommands {
			pbElements = append(pbElements, types.StringValue(pb))
		}
		v.PreBootstrapCommands, d = types.ListValue(types.StringType, pbElements)
		diags = append(diags, d...)
	} else {
		v.PreBootstrapCommands = types.ListNull(types.StringType)
	}

	if len(in.ASGSuspendProcesses) > 0 {
		aspElements := []attr.Value{}
		for _, asp := range in.ASGSuspendProcesses {
			aspElements = append(aspElements, types.StringValue(asp))
		}
		v.AsgSuspendProcesses, d = types.ListValue(types.StringType, aspElements)
		diags = append(diags, d...)
	} else {
		v.AsgSuspendProcesses = types.ListNull(types.StringType)
	}

	if in.EnableDetailedMonitoring != nil {
		v.EnableDetailedMonitoring = types.BoolPointerValue(in.EnableDetailedMonitoring)
	}
	if len(in.AvailabilityZones) > 0 {
		azElements := []attr.Value{}
		for _, az := range in.AvailabilityZones {
			azElements = append(azElements, types.StringValue(az))
		}
		v.AvailabilityZones, d = types.ListValue(types.StringType, azElements)
		diags = append(diags, d...)
	} else {
		v.AvailabilityZones = types.ListNull(types.StringType)
	}

	if len(in.Subnets) > 0 {
		snElements := []attr.Value{}
		for _, sn := range in.Subnets {
			snElements = append(snElements, types.StringValue(sn))
		}
		v.Subnets, d = types.SetValue(types.StringType, snElements)
		diags = append(diags, d...)
	} else {
		v.Subnets = types.SetNull(types.StringType)
	}

	if in.InstancePrefix != "" {
		v.InstancePrefix = types.StringValue(in.InstancePrefix)
	}

	if in.InstanceName != "" {
		v.InstanceName = types.StringValue(in.InstanceName)
	}

	if len(in.Labels) > 0 {
		lbs := map[string]attr.Value{}
		for key, val := range in.Labels {
			lbs[key] = types.StringValue(val)
		}
		v.Labels, d = types.MapValue(types.StringType, lbs)
		diags = append(diags, d...)
	} else {
		v.Labels = types.MapNull(types.StringType)
	}

	if len(in.Tags) > 0 {
		tag := map[string]attr.Value{}
		for key, val := range in.Tags {
			tag[key] = types.StringValue(val)
		}
		v.Tags, d = types.MapValue(types.StringType, tag)
		diags = append(diags, d...)
	} else {
		v.Tags = types.MapNull(types.StringType)
	}

	if in.AMI != "" {
		v.Ami = types.StringValue(in.AMI)
	}
	if in.Spot != nil {
		v.Spot = types.BoolPointerValue(in.Spot)
	}

	if len(in.InstanceTypes) > 0 {
		instanceTypesList := []attr.Value{}
		for _, val := range in.InstanceTypes {
			instanceTypesList = append(instanceTypesList, types.StringValue(val))
		}
		v.InstanceTypes, d = types.ListValue(types.StringType, instanceTypesList)
		diags = append(diags, d...)
	} else {
		v.InstanceTypes = types.ListNull(types.StringType)
	}

	if in.IAM != nil {
		stIam := Iam5Value{}
		if !state.Iam5.IsNull() {
			stIamObj, d := Iam5Type{}.ValueFromObject(ctx, state.Iam5)
			if !d.HasError() {
				stIam = stIamObj.(Iam5Value)
			} else {
				stIam = NewIam5ValueNull()
			}
		}

		iam := NewIam5ValueNull()
		d = iam.Flatten(ctx, in.IAM, stIam)
		diags = append(diags, d...)
		v.Iam5, d = iam.ToObjectValue(ctx)
		diags = append(diags, d...)
	} else {
		v.Iam5, d = NewIam5ValueNull().ToObjectValue(ctx)
		diags = append(diags, d...)
	}

	if in.SSH != nil {
		ssh := NewSsh5ValueNull()
		d = ssh.Flatten(ctx, in.SSH)
		diags = append(diags, d...)
		v.Ssh5, d = ssh.ToObjectValue(ctx)
		diags = append(diags, d...)
	} else {
		v.Ssh5, d = NewSsh5ValueNull().ToObjectValue(ctx)
		diags = append(diags, d...)
	}

	if in.Placement != nil {
		placement := NewPlacement5ValueNull()
		d = placement.Flatten(ctx, in.Placement)
		diags = append(diags, d...)
		v.Placement5, d = placement.ToObjectValue(ctx)
		diags = append(diags, d...)
	} else {
		v.Placement5, d = NewPlacement5ValueNull().ToObjectValue(ctx)
		diags = append(diags, d...)
	}

	if in.InstanceSelector != nil {
		instanceSel := NewInstanceSelector5ValueNull()
		d = instanceSel.Flatten(ctx, in.InstanceSelector)
		diags = append(diags, d...)
		v.InstanceSelector5, d = instanceSel.ToObjectValue(ctx)
		diags = append(diags, d...)
	} else {
		v.InstanceSelector5, d = NewInstanceSelector5ValueNull().ToObjectValue(ctx)
		diags = append(diags, d...)
	}

	if in.Bottlerocket != nil {
		bottlerkt := NewBottleRocket5ValueNull()
		d = bottlerkt.Flatten(ctx, in.Bottlerocket)
		diags = append(diags, d...)
		v.BottleRocket5, d = bottlerkt.ToObjectValue(ctx)
		diags = append(diags, d...)
	} else {
		v.BottleRocket5, d = NewBottleRocket5ValueNull().ToObjectValue(ctx)
		diags = append(diags, d...)
	}

	if len(in.Taints) > 0 {
		stTaints := []Taints5Value{}
		if !state.Taints5.IsNull() {
			for _, sT := range state.Taints5.Elements() {
				stTaints = append(stTaints, sT.(Taints5Value))
			}
		}

		taintsList := []attr.Value{}
		for _, val := range in.Taints {
			k := val.Key
			e := val.Effect
			var stTaint Taints5Value
			for _, stT := range stTaints {
				if !stT.IsNull() && !stT.Key.IsNull() && !stT.Effect.IsNull() &&
					getStringValue(stT.Key) == k && getStringValue(stT.Effect) == e {
					stTaint = stT
					break
				}
			}

			taint := NewTaints5ValueNull()
			d = taint.Flatten(ctx, val, stTaint)
			diags = append(diags, d...)
			taintsList = append(taintsList, taint)
		}
		v.Taints5, d = types.SetValue(Taints5Value{}.Type(ctx), taintsList)
		diags = append(diags, d...)
	} else {
		v.Taints5 = types.SetNull(Taints5Value{}.Type(ctx))
	}

	if in.UpdateConfig != nil {
		updateConfig := NewUpdateConfig5ValueNull()
		d = updateConfig.Flatten(ctx, in.UpdateConfig)
		diags = append(diags, d...)
		v.UpdateConfig5, d = updateConfig.ToObjectValue(ctx)
		diags = append(diags, d...)
	} else {
		v.UpdateConfig5, d = NewUpdateConfig5ValueNull().ToObjectValue(ctx)
		diags = append(diags, d...)
	}

	if in.SecurityGroups != nil {
		secGroup := NewSecurityGroups5ValueNull()
		d = secGroup.Flatten(ctx, in.SecurityGroups)
		diags = append(diags, d...)
		v.SecurityGroups5, d = secGroup.ToObjectValue(ctx)
		diags = append(diags, d...)
	} else {
		v.SecurityGroups5, d = NewSecurityGroups5ValueNull().ToObjectValue(ctx)
		diags = append(diags, d...)
	}

	if in.LaunchTemplate != nil {
		launchTemplate := NewLaunchTemplate5ValueNull()
		d = launchTemplate.Flatten(ctx, in.LaunchTemplate)
		diags = append(diags, d...)
		v.LaunchTemplate5, d = launchTemplate.ToObjectValue(ctx)
		diags = append(diags, d...)
	} else {
		v.LaunchTemplate5, d = NewLaunchTemplate5ValueNull().ToObjectValue(ctx)
		diags = append(diags, d...)
	}

	if in.NodeRepairConfig != nil {
		nodeRepairConfig := NewNodeRepairConfig5ValueNull()
		d = nodeRepairConfig.Flatten(ctx, in.NodeRepairConfig)
		diags = append(diags, d...)
		v.NodeRepairConfig5, d = nodeRepairConfig.ToObjectValue(ctx)
		diags = append(diags, d...)
	} else {
		v.NodeRepairConfig5, d = NewNodeRepairConfig5ValueNull().ToObjectValue(ctx)
		diags = append(diags, d...)
	}

	v.state = attr.ValueStateKnown
	return diags
}

func (v *Iam5Value) Flatten(ctx context.Context, in *rafay.NodeGroupIAM, state Iam5Value) diag.Diagnostics {
	var diags, d diag.Diagnostics
	if in == nil {
		return diags
	}

	var isPolicyV1, isPolicyV2 bool
	if !state.IsNull() && !state.AttachPolicyV2.IsNull() && !state.AttachPolicyV2.IsUnknown() &&
		getStringValue(state.AttachPolicyV2) != "" {
		isPolicyV2 = true
	}
	if !state.IsNull() && !state.AttachPolicy5.IsNull() && !state.AttachPolicy5.IsUnknown() {
		isPolicyV1 = true
	}

	if in.AttachPolicy != nil {
		if isPolicyV1 && !isPolicyV2 {
			attachPolicy := NewAttachPolicy5ValueNull()
			d = attachPolicy.Flatten(ctx, in.AttachPolicy)
			diags = append(diags, d...)
			v.AttachPolicy5, d = attachPolicy.ToObjectValue(ctx)
			diags = append(diags, d...)

			v.AttachPolicyV2 = types.StringNull()
		} else {
			var json2 = jsoniter.ConfigCompatibleWithStandardLibrary
			jsonBytes, err := json2.Marshal(in.AttachPolicy)
			if err != nil {
				diags.AddError("Attach Policy Marshal Error", err.Error())
			}
			v.AttachPolicyV2 = types.StringValue(string(jsonBytes))

			v.AttachPolicy5, d = NewAttachPolicy5ValueNull().ToObjectValue(ctx)
			diags = append(diags, d...)
		}
	} else {
		v.AttachPolicy5, d = NewAttachPolicy5ValueNull().ToObjectValue(ctx)
		diags = append(diags, d...)
		v.AttachPolicyV2 = types.StringNull()
	}

	if len(in.AttachPolicyARNs) > 0 {
		arns := []attr.Value{}
		for _, arn := range in.AttachPolicyARNs {
			arns = append(arns, types.StringValue(arn))
		}
		v.AttachPolicyArns, d = types.SetValue(types.StringType, arns)
		diags = append(diags, d...)
	} else {
		v.AttachPolicyArns = types.SetNull(types.StringType)
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
		addonPolicies := NewIamNodeGroupWithAddonPolicies5ValueNull()
		d = addonPolicies.Flatten(ctx, in.WithAddonPolicies)
		diags = append(diags, d...)
		v.IamNodeGroupWithAddonPolicies5, d = addonPolicies.ToObjectValue(ctx)
		diags = append(diags, d...)
	} else {
		v.IamNodeGroupWithAddonPolicies5, d = NewIamNodeGroupWithAddonPolicies5ValueNull().ToObjectValue(ctx)
		diags = append(diags, d...)
	}

	v.state = attr.ValueStateKnown
	return diags
}

func (v *AttachPolicy5Value) Flatten(ctx context.Context, in *rafay.InlineDocument) diag.Diagnostics {
	var diags, d diag.Diagnostics
	if in == nil {
		return diags
	}

	if in.Version != "" {
		v.Version = types.StringValue(in.Version)
	}
	if in.Id != "" {
		v.Id = types.StringValue(in.Id)
	}

	if len(in.Statement) > 0 {
		stms := []attr.Value{}
		for _, stm := range in.Statement {
			sv := NewStatement5ValueNull()
			d = sv.Flatten(ctx, stm)
			diags = append(diags, d...)
			stms = append(stms, sv)
		}
		v.Statement5, d = types.SetValue(Statement5Value{}.Type(ctx), stms)
		diags = append(diags, d...)
	} else {
		v.Statement5 = types.SetNull(Statement5Value{}.Type(ctx))
	}

	v.state = attr.ValueStateKnown
	return diags
}

func (v *IamNodeGroupWithAddonPolicies5Value) Flatten(ctx context.Context, in *rafay.NodeGroupIAMAddonPolicies) diag.Diagnostics {
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

func (v *Statement5Value) Flatten(ctx context.Context, in rafay.InlineStatement) diag.Diagnostics {
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
		v.Action, d = types.SetValue(types.StringType, actEle)
		diags = append(diags, d...)
	} else {
		v.Action = types.SetNull(types.StringType)
	}
	if in.NotAction != nil && len(in.NotAction.([]interface{})) > 0 {
		naEle := []attr.Value{}
		for _, na := range in.NotAction.([]interface{}) {
			naEle = append(naEle, types.StringValue(na.(string)))
		}
		v.NotAction, d = types.SetValue(types.StringType, naEle)
		diags = append(diags, d...)
	} else {
		v.NotAction = types.SetNull(types.StringType)
	}
	if len(in.Resource.(string)) > 0 {
		v.Resource = types.StringValue(in.Resource.(string))
	}
	if in.NotResource != nil && len(in.NotResource.([]interface{})) > 0 {
		nrEle := []attr.Value{}
		for _, nr := range in.NotResource.([]interface{}) {
			nrEle = append(nrEle, types.StringValue(nr.(string)))
		}
		v.NotResource, d = types.SetValue(types.StringType, nrEle)
		diags = append(diags, d...)
	} else {
		v.NotResource = types.SetNull(types.StringType)
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

func (v *Ssh5Value) Flatten(ctx context.Context, in *rafay.NodeGroupSSH) diag.Diagnostics {
	var diags, d diag.Diagnostics

	if in.Allow != nil {
		v.Allow = types.BoolPointerValue(in.Allow)
	}
	if in.PublicKeyPath != "" {
		v.PublicKey = types.StringValue(in.PublicKeyPath)
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

func (v *Placement5Value) Flatten(ctx context.Context, in *rafay.Placement) diag.Diagnostics {
	var diags diag.Diagnostics

	if in.GroupName != "" {
		v.Group = types.StringValue(in.GroupName)
	}

	v.state = attr.ValueStateKnown
	return diags
}

func (v *InstanceSelector5Value) Flatten(ctx context.Context, in *rafay.InstanceSelector) diag.Diagnostics {
	var diags diag.Diagnostics

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

func (v *BottleRocket5Value) Flatten(ctx context.Context, in *rafay.NodeGroupBottlerocket) diag.Diagnostics {
	var diags diag.Diagnostics

	if in.EnableAdminContainer != nil {
		v.EnableAdminContainer = types.BoolPointerValue(in.EnableAdminContainer)
	}

	if len(in.Settings) > 0 {
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

func (v *Taints5Value) Flatten(ctx context.Context, in rafay.NodeGroupTaint, state Taints5Value) diag.Diagnostics {
	var diags diag.Diagnostics

	if in.Key != "" {
		v.Key = types.StringValue(in.Key)
	}

	if in.Value != "" {
		v.Value = types.StringValue(in.Value)
	} else {
		// hack: API can not differenciate nil and zero value of Value field. This is to avoid unnecessary diffs.
		if !state.IsNull() && !state.Value.IsNull() {
			v.Value = state.Value
		}
	}
	if in.Effect != "" {
		v.Effect = types.StringValue(in.Effect)
	}

	v.state = attr.ValueStateKnown
	return diags
}

func (v *LaunchTemplate5Value) Flatten(ctx context.Context, in *rafay.LaunchTemplate) diag.Diagnostics {
	var diags diag.Diagnostics
	if in == nil {
		return diags
	}

	if in.ID != "" {
		v.Id = types.StringValue(in.ID)
	}
	if in.Version != "" {
		v.Version = types.StringValue(in.Version)
	}

	v.state = attr.ValueStateKnown
	return diags
}

func (v *UpdateConfig5Value) Flatten(ctx context.Context, in *rafay.NodeGroupUpdateConfig) diag.Diagnostics {
	var diags diag.Diagnostics
	if in == nil {
		return diags
	}

	if in.MaxUnavailable != nil {
		v.MaxUnavailable = types.Int64Value(int64(*in.MaxUnavailable))
	}
	if in.MaxUnavailablePercentage != nil {
		v.MaxUnavailablePercentage = types.Int64Value(int64(*in.MaxUnavailablePercentage))
	}

	v.state = attr.ValueStateKnown
	return diags
}

func (v *SecurityGroups5Value) Flatten(ctx context.Context, in *rafay.NodeGroupSGs) diag.Diagnostics {
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

func (v *NodeRepairConfig5Value) Flatten(ctx context.Context, in *rafay.NodeRepairConfig) diag.Diagnostics {
	var diags diag.Diagnostics
	if in == nil {
		return diags
	}

	if in.Enabled != nil {
		v.Enabled = types.BoolPointerValue(in.Enabled)
	}

	v.state = attr.ValueStateKnown
	return diags
}
