package resource_eks_cluster

import (
	"context"

	"github.com/RafaySystems/terraform-provider-rafay/rafay"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	jsoniter "github.com/json-iterator/go"
)

// --- Managed Node Group Map Expand ---
func (v ManagedNodegroupsMapValue) Expand(ctx context.Context) (*rafay.ManagedNodeGroup, diag.Diagnostics) {
	var diags, d diag.Diagnostics
	var mng rafay.ManagedNodeGroup
	if v.IsNull() {
		return &rafay.ManagedNodeGroup{}, diags
	}

	if !v.Version.IsNull() && !v.Version.IsUnknown() {
		mng.Version = getStringValue(v.Version)
	}
	if !v.EnableDetailedMonitoring.IsNull() && !v.EnableDetailedMonitoring.IsUnknown() {
		edm := getBoolValue(v.EnableDetailedMonitoring)
		mng.EnableDetailedMonitoring = &edm
	}
	if !v.AmiFamily.IsNull() && !v.AmiFamily.IsUnknown() {
		mng.AMIFamily = getStringValue(v.AmiFamily)
	}
	if !v.InstanceType.IsNull() && !v.InstanceType.IsUnknown() {
		mng.InstanceType = getStringValue(v.InstanceType)
	}
	if !v.MaxPodsPerNode.IsNull() && !v.MaxPodsPerNode.IsUnknown() {
		mpn := int(getInt64Value(v.MaxPodsPerNode))
		mng.MaxPodsPerNode = &mpn
	}
	if !v.VolumeType.IsNull() && !v.VolumeType.IsUnknown() {
		mng.VolumeType = getStringValue(v.VolumeType)
	}
	if !v.VolumeSize.IsNull() && !v.VolumeSize.IsUnknown() {
		volSize := int(getInt64Value(v.VolumeSize))
		mng.VolumeSize = &volSize
	}
	if !v.VolumeIops.IsNull() && !v.VolumeIops.IsUnknown() {
		volIops := int(getInt64Value(v.VolumeIops))
		mng.VolumeIOPS = &volIops
	}
	if !v.VolumeThroughput.IsNull() && !v.VolumeThroughput.IsUnknown() {
		volThroughput := int(getInt64Value(v.VolumeThroughput))
		mng.VolumeThroughput = &volThroughput
	}
	if !v.DesiredCapacity.IsNull() && !v.DesiredCapacity.IsUnknown() {
		desiredCap := int(getInt64Value(v.DesiredCapacity))
		mng.DesiredCapacity = &desiredCap
	}
	if !v.MaxSize.IsNull() && !v.MaxSize.IsUnknown() {
		maxSize := int(getInt64Value(v.MaxSize))
		mng.MaxSize = &maxSize
	}
	if !v.MinSize.IsNull() && !v.MinSize.IsUnknown() {
		minSize := int(getInt64Value(v.MinSize))
		mng.MinSize = &minSize
	}
	if !v.PrivateNetworking.IsNull() && !v.PrivateNetworking.IsUnknown() {
		privNet := getBoolValue(v.PrivateNetworking)
		mng.PrivateNetworking = &privNet
	}
	if !v.DisableImdsv1.IsNull() && !v.DisableImdsv1.IsUnknown() {
		disImdsv1 := getBoolValue(v.DisableImdsv1)
		mng.DisableIMDSv1 = &disImdsv1
	}
	if !v.DisablePodsImds.IsNull() && !v.DisablePodsImds.IsUnknown() {
		disPodImds := getBoolValue(v.DisablePodsImds)
		mng.DisablePodIMDS = &disPodImds
	}
	if !v.EfaEnabled.IsNull() && !v.EfaEnabled.IsUnknown() {
		efaEnabled := getBoolValue(v.EfaEnabled)
		mng.EFAEnabled = &efaEnabled
	}
	if !v.Labels.IsNull() && !v.Labels.IsUnknown() {
		mng.Labels = make(map[string]string, len(v.Labels.Elements()))
		vLabels := make(map[string]types.String, len(v.Labels.Elements()))
		d = v.Labels.ElementsAs(ctx, &vLabels, false)
		diags = append(diags, d...)
		for k, val := range vLabels {
			mng.Labels[k] = getStringValue(val)
		}
	}
	if !v.Tags.IsNull() && !v.Tags.IsUnknown() {
		mng.Tags = make(map[string]string, len(v.Tags.Elements()))
		vTags := make(map[string]types.String, len(v.Tags.Elements()))
		d = v.Tags.ElementsAs(ctx, &vTags, false)
		diags = append(diags, d...)
		for k, val := range vTags {
			mng.Tags[k] = getStringValue(val)
		}
	}
	if !v.EnableDetailedMonitoring.IsNull() && !v.EnableDetailedMonitoring.IsUnknown() {
		enableDM := getBoolValue(v.EnableDetailedMonitoring)
		mng.EnableDetailedMonitoring = &enableDM
	}
	if !v.AvailabilityZones.IsNull() && !v.AvailabilityZones.IsUnknown() {
		azList := make([]types.String, 0, len(v.AvailabilityZones.Elements()))
		d = v.AvailabilityZones.ElementsAs(ctx, &azList, false)
		diags = append(diags, d...)
		azs := make([]string, 0, len(azList))
		for _, az := range azList {
			azs = append(azs, getStringValue(az))
		}
		mng.AvailabilityZones = azs
	}
	if !v.Subnets.IsNull() && !v.Subnets.IsUnknown() {
		subnetsList := make([]types.String, 0, len(v.Subnets.Elements()))
		d = v.Subnets.ElementsAs(ctx, &subnetsList, false)
		diags = append(diags, d...)
		subnets := make([]string, 0, len(subnetsList))
		for _, s := range subnetsList {
			subnets = append(subnets, getStringValue(s))
		}
		mng.Subnets = subnets
	}
	if !v.InstancePrefix.IsNull() && !v.InstancePrefix.IsUnknown() {
		mng.InstancePrefix = getStringValue(v.InstancePrefix)
	}
	if !v.InstanceName.IsNull() && !v.InstanceName.IsUnknown() {
		mng.InstanceName = getStringValue(v.InstanceName)
	}
	if !v.Ami.IsNull() && !v.Ami.IsUnknown() {
		mng.AMI = getStringValue(v.Ami)
	}
	if !v.AsgSuspendProcesses.IsNull() && !v.AsgSuspendProcesses.IsUnknown() {
		asgSuspendList := make([]types.String, 0, len(v.AsgSuspendProcesses.Elements()))
		d = v.AsgSuspendProcesses.ElementsAs(ctx, &asgSuspendList, false)
		asgSuspend := make([]string, 0, len(asgSuspendList))
		for _, p := range asgSuspendList {
			asgSuspend = append(asgSuspend, getStringValue(p))
		}
		mng.ASGSuspendProcesses = asgSuspend
	}
	if !v.EbsOptimized.IsNull() && !v.EbsOptimized.IsUnknown() {
		ebsOpt := getBoolValue(v.EbsOptimized)
		mng.EBSOptimized = &ebsOpt
	}
	if !v.VolumeName.IsNull() && !v.VolumeName.IsUnknown() {
		mng.VolumeName = getStringValue(v.VolumeName)
	}
	if !v.VolumeEncrypted.IsNull() && !v.VolumeEncrypted.IsUnknown() {
		volEncrypted := getBoolValue(v.VolumeEncrypted)
		mng.VolumeEncrypted = &volEncrypted
	}
	if !v.VolumeKmsKeyId.IsNull() && !v.VolumeKmsKeyId.IsUnknown() {
		mng.VolumeKmsKeyID = getStringValue(v.VolumeKmsKeyId)
	}
	if !v.PreBootstrapCommands.IsNull() && !v.PreBootstrapCommands.IsUnknown() {
		preBootstrapList := make([]types.String, 0, len(v.PreBootstrapCommands.Elements()))
		d = v.PreBootstrapCommands.ElementsAs(ctx, &preBootstrapList, false)
		preBootstrap := make([]string, 0, len(preBootstrapList))
		for _, cmd := range preBootstrapList {
			preBootstrap = append(preBootstrap, getStringValue(cmd))
		}
		mng.PreBootstrapCommands = preBootstrap
	}
	if !v.OverrideBootstrapCommand.IsNull() && !v.OverrideBootstrapCommand.IsUnknown() {
		mng.OverrideBootstrapCommand = getStringValue(v.OverrideBootstrapCommand)
	}
	if !v.InstanceTypes.IsNull() && !v.InstanceTypes.IsUnknown() {
		itsList := make([]types.String, 0, len(v.InstanceTypes.Elements()))
		d = v.InstanceTypes.ElementsAs(ctx, &itsList, false)
		diags = append(diags, d...)
		its := make([]string, 0, len(itsList))
		for _, it := range itsList {
			its = append(its, getStringValue(it))
		}
		mng.InstanceTypes = its
	}
	if !v.Spot.IsNull() && !v.Spot.IsUnknown() {
		spot := getBoolValue(v.Spot)
		mng.Spot = &spot
	}

	if !v.Iam5.IsNull() && !v.Iam5.IsUnknown() {
		var iamTypes Iam5Type
		tfIamValue, d := iamTypes.ValueFromObject(ctx, v.Iam5)
		if d.HasError() {
			diags = append(diags, d...)
			return &mng, diags
		}
		mng.IAM, d = tfIamValue.(Iam5Value).Expand(ctx)
		diags = append(diags, d...)
	}

	if !v.Ssh5.IsNull() && !v.Ssh5.IsUnknown() {
		var sshTypes SshType
		tfSshValue, d := sshTypes.ValueFromObject(ctx, v.Ssh5)
		if d.HasError() {
			diags = append(diags, d...)
			return &mng, diags
		}
		mng.SSH, d = tfSshValue.(Ssh5Value).Expand(ctx)
		diags = append(diags, d...)
	}

	if !v.Placement5.IsNull() && !v.Placement5.IsUnknown() {
		var placementTypes PlacementType
		tfPlacementValue, d := placementTypes.ValueFromObject(ctx, v.Placement5)
		if d.HasError() {
			diags = append(diags, d...)
			return &mng, diags
		}
		mng.Placement, d = tfPlacementValue.(Placement5Value).Expand(ctx)
		diags = append(diags, d...)
	}

	if !v.InstanceSelector5.IsNull() && !v.InstanceSelector5.IsUnknown() {
		var instanceSelectorTypes InstanceSelectorType
		tfInstanceSelectorValue, d := instanceSelectorTypes.ValueFromObject(ctx, v.InstanceSelector5)
		if d.HasError() {
			diags = append(diags, d...)
			return &mng, diags
		}
		mng.InstanceSelector, d = tfInstanceSelectorValue.(InstanceSelector5Value).Expand(ctx)
		diags = append(diags, d...)
	}

	if !v.BottleRocket5.IsNull() && !v.BottleRocket5.IsUnknown() {
		var bottleRocketTypes BottleRocketType
		tfBottleRocketValue, d := bottleRocketTypes.ValueFromObject(ctx, v.BottleRocket5)
		if d.HasError() {
			diags = append(diags, d...)
			return &mng, diags
		}
		mng.Bottlerocket, d = tfBottleRocketValue.(BottleRocket5Value).Expand(ctx)
		diags = append(diags, d...)
	}

	for _, taint := range v.Taints5.Elements() {
		t, d := taint.(Taints5Value).Expand(ctx)
		diags = append(diags, d...)
		mng.Taints = append(mng.Taints, t)

	}

	if !v.UpdateConfig5.IsNull() && !v.UpdateConfig5.IsUnknown() {
		var updateConfigTypes UpdateConfigType
		tfUpdateConfigValue, d := updateConfigTypes.ValueFromObject(ctx, v.UpdateConfig5)
		if d.HasError() {
			diags = append(diags, d...)
			return &mng, diags
		}
		mng.UpdateConfig, d = tfUpdateConfigValue.(UpdateConfig5Value).Expand(ctx)
		diags = append(diags, d...)
	}

	if !v.LaunchTemplate5.IsNull() && !v.LaunchTemplate5.IsUnknown() {
		var launchTemplateTypes LaunchTemplate5Type
		tfLaunchTemplateValue, d := launchTemplateTypes.ValueFromObject(ctx, v.LaunchTemplate5)
		if d.HasError() {
			diags = append(diags, d...)
			return &mng, diags
		}
		mng.LaunchTemplate, d = tfLaunchTemplateValue.(LaunchTemplate5Value).Expand(ctx)
		diags = append(diags, d...)
	}

	if !v.SecurityGroups5.IsNull() && !v.SecurityGroups5.IsUnknown() {
		var securityGroupsTypes SecurityGroups5Type
		tfSecurityGroupsValue, d := securityGroupsTypes.ValueFromObject(ctx, v.SecurityGroups5)
		if d.HasError() {
			diags = append(diags, d...)
			return &mng, diags
		}
		mng.SecurityGroups, d = tfSecurityGroupsValue.(SecurityGroups5Value).Expand(ctx)
		diags = append(diags, d...)
	}

	return &mng, diags
}

func (v Iam5Value) Expand(ctx context.Context) (*rafay.NodeGroupIAM, diag.Diagnostics) {
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

	// Map iam_node_group_with_addon_policies block
	if !v.IamNodeGroupWithAddonPolicies5.IsNull() && !v.IamNodeGroupWithAddonPolicies5.IsUnknown() {
		var iamNodeGroupPolicies IamNodeGroupWithAddonPolicies5Type
		tfIamNodeGroupPoliciesValue, d := iamNodeGroupPolicies.ValueFromObject(ctx, v.IamNodeGroupWithAddonPolicies5)
		if d.HasError() {
			diags = append(diags, d...)
			return &iam, diags
		}
		iam.WithAddonPolicies, d = tfIamNodeGroupPoliciesValue.(IamNodeGroupWithAddonPolicies5Value).Expand(ctx)
		diags = append(diags, d...)
	}

	// Map attach_policy block
	if !v.AttachPolicy5.IsNull() && !v.AttachPolicy5.IsUnknown() {
		var attachPolicy AttachPolicy5Type
		tfAttachPolicyValue, d := attachPolicy.ValueFromObject(ctx, v.AttachPolicy5)
		if d.HasError() {
			diags = append(diags, d...)
			return &iam, diags
		}
		iam.AttachPolicy, d = tfAttachPolicyValue.(AttachPolicy5Value).Expand(ctx)
		diags = append(diags, d...)
	}

	return &iam, diags
}

func (v IamNodeGroupWithAddonPolicies5Value) Expand(ctx context.Context) (*rafay.NodeGroupIAMAddonPolicies, diag.Diagnostics) {
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

func (v AttachPolicy5Value) Expand(ctx context.Context) (*rafay.InlineDocument, diag.Diagnostics) {
	var diags diag.Diagnostics
	var policy rafay.InlineDocument

	if v.IsNull() {
		return &rafay.InlineDocument{}, diags
	}

	if !v.Version.IsNull() && !v.Version.IsUnknown() {
		policy.Version = getStringValue(v.Version)
	}

	if !v.Statement5.IsNull() && !v.Statement5.IsUnknown() {
		for _, stmt := range v.Statement5.Elements() {
			s, d := stmt.(Statement5Value).Expand(ctx)
			diags = append(diags, d...)
			policy.Statement = append(policy.Statement, s)
		}
	}

	if !v.Id.IsNull() && !v.Id.IsUnknown() {
		policy.Id = getStringValue(v.Id)
	}

	return &policy, diags
}

func (v Statement5Value) Expand(ctx context.Context) (rafay.InlineStatement, diag.Diagnostics) {
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

func (v Ssh5Value) Expand(ctx context.Context) (*rafay.NodeGroupSSH, diag.Diagnostics) {
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
func (v Placement5Value) Expand(ctx context.Context) (*rafay.Placement, diag.Diagnostics) {
	var diags diag.Diagnostics
	var placement rafay.Placement

	if !v.Group.IsNull() && !v.Group.IsUnknown() {
		group := getStringValue(v.Group)
		placement.GroupName = group
	}

	return &placement, diags
}

func (v InstanceSelector5Value) Expand(ctx context.Context) (*rafay.InstanceSelector, diag.Diagnostics) {
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

func (v BottleRocket5Value) Expand(ctx context.Context) (*rafay.NodeGroupBottlerocket, diag.Diagnostics) {
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

func (v Taints5Value) Expand(ctx context.Context) (rafay.NodeGroupTaint, diag.Diagnostics) {
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

func (v UpdateConfig5Value) Expand(ctx context.Context) (*rafay.NodeGroupUpdateConfig, diag.Diagnostics) {
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

func (v LaunchTemplate5Value) Expand(ctx context.Context) (*rafay.LaunchTemplate, diag.Diagnostics) {
	var diags diag.Diagnostics
	var lt rafay.LaunchTemplate

	if !v.Id.IsNull() && !v.Id.IsUnknown() {
		lt.ID = getStringValue(v.Id)
	}

	if !v.Version.IsNull() && !v.Version.IsUnknown() {
		lt.Version = getStringValue(v.Version)
	}

	return &lt, diags
}

// SecurityGroups5Value Expand
func (v SecurityGroups5Value) Expand(ctx context.Context) (*rafay.NodeGroupSGs, diag.Diagnostics) {
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
