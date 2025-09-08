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

func (v *ManagedNodegroupsValue) Flatten(ctx context.Context, in *rafay.ManagedNodeGroup) diag.Diagnostics {
	var diags, d diag.Diagnostics

	v.Name = types.StringValue(in.Name)
	v.AmiFamily = types.StringValue(in.AMIFamily)
	v.DesiredCapacity = types.Int64Value(int64(*in.DesiredCapacity))
	v.DisableImdsv1 = types.BoolValue(*in.DisableIMDSv1)
	v.DisablePodsImds = types.BoolValue(*in.DisablePodIMDS)
	v.EfaEnabled = types.BoolValue(*in.EFAEnabled)
	v.InstanceType = types.StringValue(in.InstanceType)
	v.MaxPodsPerNode = types.Int64Value(int64(*in.MaxPodsPerNode))
	v.MaxSize = types.Int64Value(int64(*in.MaxSize))
	v.MinSize = types.Int64Value(int64(*in.MinSize))
	v.PrivateNetworking = types.BoolValue(*in.PrivateNetworking)
	v.Version = types.StringValue(in.Version)
	v.VolumeIops = types.Int64Value(int64(*in.VolumeIOPS))
	v.VolumeSize = types.Int64Value(int64(*in.VolumeSize))
	v.VolumeThroughput = types.Int64Value(int64(*in.VolumeThroughput))
	v.VolumeType = types.StringValue(in.VolumeType)
	v.EbsOptimized = types.BoolValue(*in.EBSOptimized)
	v.VolumeName = types.StringValue(in.VolumeName)
	v.VolumeEncrypted = types.BoolValue(*in.VolumeEncrypted)
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

	asgSuspendProcesses := types.ListNull(types.StringType)
	if len(in.ASGSuspendProcesses) > 0 {
		aspElements := []attr.Value{}
		for _, asp := range in.ASGSuspendProcesses {
			aspElements = append(aspElements, types.StringValue(asp))
		}
		asgSuspendProcesses, d = types.ListValue(types.StringType, aspElements)
		diags = append(diags, d...)
	}
	v.AsgSuspendProcesses = asgSuspendProcesses

	v.EnableDetailedMonitoring = types.BoolPointerValue(in.EnableDetailedMonitoring)

	availabilityZones := types.ListNull(types.StringType)
	if len(in.AvailabilityZones) > 0 {
		azElements := []attr.Value{}
		for _, az := range in.AvailabilityZones {
			azElements = append(azElements, types.StringValue(az))
		}
		availabilityZones, d = types.ListValue(types.StringType, azElements)
		diags = append(diags, d...)
	}
	v.AvailabilityZones = availabilityZones
	snElements := []attr.Value{}
	for _, sn := range in.Subnets {
		snElements = append(snElements, types.StringValue(sn))
	}
	v.Subnets, d = types.ListValue(types.StringType, snElements)
	diags = append(diags, d...)
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
	v.Labels = lbsMap
	tagMap := types.MapNull(types.StringType)
	if len(in.Tags) != 0 {
		tag := map[string]attr.Value{}
		for key, val := range in.Tags {
			tag[key] = types.StringValue(val)
		}
		tagMap, d = types.MapValue(types.StringType, tag)
		diags = append(diags, d...)
	}
	v.Tags = tagMap
	v.Ami = types.StringValue(in.AMI)
	v.Spot = types.BoolPointerValue(in.Spot)
	instanceTypes := types.ListNull(types.StringType)
	if len(in.InstanceTypes) > 0 {
		instanceTypesList := []attr.Value{}
		for _, val := range in.InstanceTypes {
			instanceTypesList = append(instanceTypesList, types.StringValue(val))
		}
		instanceTypes, d = types.ListValue(types.StringType, instanceTypesList)
		diags = append(diags, d...)
	}
	v.InstanceTypes = instanceTypes

	// blocks start here
	iam := NewIam4ValueNull()
	d = iam.Flatten(ctx, in.IAM)
	diags = append(diags, d...)
	iamElements := []attr.Value{iam}
	v.Iam4, d = types.ListValue(Iam4Value{}.Type(ctx), iamElements)
	diags = append(diags, d...)

	ssh := NewSsh4ValueNull()
	d = ssh.Flatten(ctx, in.SSH)
	diags = append(diags, d...)
	v.Ssh4, d = types.ListValue(Ssh4Value{}.Type(ctx), []attr.Value{ssh})
	diags = append(diags, d...)

	placement := NewPlacement4ValueNull()
	d = placement.Flatten(ctx, in.Placement)
	diags = append(diags, d...)
	v.Placement4, d = types.ListValue(Placement4Value{}.Type(ctx), []attr.Value{placement})
	diags = append(diags, d...)

	instanceSel := NewInstanceSelector4ValueNull()
	d = instanceSel.Flatten(ctx, in.InstanceSelector)
	diags = append(diags, d...)
	v.InstanceSelector4, d = types.ListValue(InstanceSelector4Value{}.Type(ctx), []attr.Value{instanceSel})
	diags = append(diags, d...)

	bottlerkt := NewBottleRocket4ValueNull()
	d = bottlerkt.Flatten(ctx, in.Bottlerocket)
	diags = append(diags, d...)
	v.BottleRocket4, d = types.ListValue(BottleRocket4Value{}.Type(ctx), []attr.Value{bottlerkt})
	diags = append(diags, d...)

	taints := types.ListNull(Taints4Value{}.Type(ctx))
	if len(in.Taints) > 0 {
		taintsList := []attr.Value{}
		for _, val := range in.Taints {
			taint := NewTaints4ValueNull()
			d = taint.Flatten(ctx, val)
			diags = append(diags, d...)
			taintsList = append(taintsList, taint)
		}
		taints, d = types.ListValue(Taints4Value{}.Type(ctx), taintsList)
		diags = append(diags, d...)
	}
	v.Taints4 = taints

	updateConfig := NewUpdateConfig4ValueNull()
	d = updateConfig.Flatten(ctx, in.UpdateConfig)
	diags = append(diags, d...)
	v.UpdateConfig4, d = types.ListValue(UpdateConfig4Value{}.Type(ctx), []attr.Value{updateConfig})
	diags = append(diags, d...)

	if in.SecurityGroups != nil {
		secGroup := NewSecurityGroups4ValueNull()
		d = secGroup.Flatten(ctx, in.SecurityGroups)
		diags = append(diags, d...)
		v.SecurityGroups4, d = types.ListValue(SecurityGroups4Value{}.Type(ctx), []attr.Value{secGroup})
		diags = append(diags, d...)
	} else {
		v.SecurityGroups4 = types.ListNull(SecurityGroups4Value{}.Type(ctx))
	}

	if in.LaunchTemplate != nil {
		launchTemplate := NewLaunchTemplate4ValueNull()
		d = launchTemplate.Flatten(ctx, in.LaunchTemplate)
		diags = append(diags, d...)
		v.LaunchTemplate4, d = types.ListValue(LaunchTemplate4Value{}.Type(ctx), []attr.Value{launchTemplate})
		diags = append(diags, d...)
	} else {
		v.LaunchTemplate4 = types.ListNull(LaunchTemplate4Value{}.Type(ctx))
	}

	v.state = attr.ValueStateKnown
	return diags
}

// --- managed node groups ---
func (v *UpdateConfig4Value) Flatten(ctx context.Context, in *rafay.NodeGroupUpdateConfig) diag.Diagnostics {
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

func (v *Taints4Value) Flatten(ctx context.Context, in rafay.NodeGroupTaint) diag.Diagnostics {
	var diags diag.Diagnostics

	v.Key = types.StringValue(in.Key)
	v.Value = types.StringValue(in.Value)
	v.Effect = types.StringValue(in.Effect)

	v.state = attr.ValueStateKnown
	return diags
}

func (v *BottleRocket4Value) Flatten(ctx context.Context, in *rafay.NodeGroupBottlerocket) diag.Diagnostics {
	var diags diag.Diagnostics

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

func (v *InstanceSelector4Value) Flatten(ctx context.Context, in *rafay.InstanceSelector) diag.Diagnostics {
	var diags diag.Diagnostics

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

func (v *Placement4Value) Flatten(ctx context.Context, in *rafay.Placement) diag.Diagnostics {
	var diags diag.Diagnostics

	v.Group = types.StringValue(in.GroupName)

	v.state = attr.ValueStateKnown
	return diags
}

func (v *Ssh4Value) Flatten(ctx context.Context, in *rafay.NodeGroupSSH) diag.Diagnostics {
	var diags, d diag.Diagnostics

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

func (v *Iam4Value) Flatten(ctx context.Context, in *rafay.NodeGroupIAM) diag.Diagnostics {
	var diags, d diag.Diagnostics
	if in == nil {
		return diags
	}

	// TODO(Akshay): Check if Attach Policy v2. based on that populate attach_policy and attach_policy_v2
	if in.AttachPolicy != nil {
		var json2 = jsoniter.ConfigCompatibleWithStandardLibrary
		jsonBytes, err := json2.Marshal(in.AttachPolicy)
		if err != nil {
			diags.AddError("Attach Policy Marshal Error", err.Error())
		}
		v.AttachPolicyV2 = types.StringValue(string(jsonBytes))
	}

	attachPolicy := NewAttachPolicyValueNull()
	d = attachPolicy.Flatten(ctx, in.AttachPolicy)
	diags = append(diags, d...)
	v.AttachPolicy4, d = types.ListValue(AttachPolicy4Value{}.Type(ctx), []attr.Value{attachPolicy})
	diags = append(diags, d...)

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

	addonPolicies := NewIamNodeGroupWithAddonPolicies4ValueNull()
	d = addonPolicies.Flatten(ctx, in.WithAddonPolicies)
	diags = append(diags, d...)
	addonPoliciesElements := []attr.Value{addonPolicies}
	v.IamNodeGroupWithAddonPolicies4, d = types.ListValue(IamNodeGroupWithAddonPolicies4Value{}.Type(ctx), addonPoliciesElements)
	diags = append(diags, d...)

	v.state = attr.ValueStateKnown
	return diags
}

func (v *AttachPolicy4Value) Flatten(ctx context.Context, in *rafay.InlineDocument) diag.Diagnostics {
	var diags, d diag.Diagnostics
	if in == nil {
		return diags
	}

	v.Version = types.StringValue(in.Version)
	v.Id = types.StringValue(in.Id)

	statement4 := types.ListNull(Statement4Value{}.Type(ctx))
	if len(in.Statement) > 0 {
		stms := []attr.Value{}
		for _, stm := range in.Statement {
			sv := NewStatement4ValueNull()
			d = sv.Flatten(ctx, stm)
			diags = append(diags, d...)
			stms = append(stms, sv)
		}
		statement4, d = types.ListValue(Statement4Value{}.Type(ctx), stms)
		diags = append(diags, d...)
	}
	v.Statement4 = statement4

	v.state = attr.ValueStateKnown
	return diags
}

func (v *Statement4Value) Flatten(ctx context.Context, in rafay.InlineStatement) diag.Diagnostics {
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

func (v *IamNodeGroupWithAddonPolicies4Value) Flatten(ctx context.Context, in *rafay.NodeGroupIAMAddonPolicies) diag.Diagnostics {
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

func (v *LaunchTemplate4Value) Flatten(ctx context.Context, in *rafay.LaunchTemplate) diag.Diagnostics {
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

func (v *SecurityGroups4Value) Flatten(ctx context.Context, in *rafay.NodeGroupSGs) diag.Diagnostics {
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
