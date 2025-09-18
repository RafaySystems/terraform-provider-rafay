package resource_eks_cluster

import (
	"context"

	"github.com/RafaySystems/terraform-provider-rafay/rafay"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	jsoniter "github.com/json-iterator/go"
)

// --- Nodegroup map Expand ---
func (v NodeGroupsMapValue) Expand(ctx context.Context) (*rafay.NodeGroup, diag.Diagnostics) {
	var diags, d diag.Diagnostics
	var ng rafay.NodeGroup
	if v.IsNull() {
		return &rafay.NodeGroup{}, diags
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
	if !v.Labels.IsNull() && !v.Labels.IsUnknown() {
		ng.Labels = make(map[string]string, len(v.Labels.Elements()))
		vLabels := make(map[string]types.String, len(v.Labels.Elements()))
		d = v.Labels.ElementsAs(ctx, &vLabels, false)
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
	if !v.Iam6.IsNull() && !v.Iam6.IsUnknown() {
		var iamType Iam6Type
		tfIamValue, d := iamType.ValueFromObject(ctx, v.Iam6)
		if d.HasError() {
			diags = append(diags, d...)
			return &ng, diags
		}
		ng.IAM, d = tfIamValue.(Iam6Value).Expand(ctx)
		diags = append(diags, d...)
	}

	if !v.Ssh6.IsNull() && !v.Ssh6.IsUnknown() {
		var sshTypes Ssh6Type
		tfSshValue, d := sshTypes.ValueFromObject(ctx, v.Ssh6)
		if d.HasError() {
			diags = append(diags, d...)
			return &ng, diags
		}
		ng.SSH, d = tfSshValue.(Ssh6Value).Expand(ctx)
		diags = append(diags, d...)
	}

	if !v.Placement6.IsNull() && !v.Placement6.IsUnknown() {
		var placementTypes Placement6Type
		tfPlacementValue, d := placementTypes.ValueFromObject(ctx, v.Placement6)
		if d.HasError() {
			diags = append(diags, d...)
			return &ng, diags
		}
		ng.Placement, d = tfPlacementValue.(Placement6Value).Expand(ctx)
		diags = append(diags, d...)
	}

	if !v.InstanceSelector6.IsNull() && !v.InstanceSelector6.IsUnknown() {
		var instanceSelectorTypes InstanceSelector6Type
		tfInstanceSelectorValue, d := instanceSelectorTypes.ValueFromObject(ctx, v.InstanceSelector6)
		if d.HasError() {
			diags = append(diags, d...)
			return &ng, diags
		}
		ng.InstanceSelector, d = tfInstanceSelectorValue.(InstanceSelector6Value).Expand(ctx)
		diags = append(diags, d...)
	}

	if !v.BottleRocket6.IsNull() && !v.BottleRocket6.IsUnknown() {
		var bottleRocketTypes BottleRocket6Type
		tfBottleRocketValue, d := bottleRocketTypes.ValueFromObject(ctx, v.BottleRocket6)
		if d.HasError() {
			diags = append(diags, d...)
			return &ng, diags
		}
		ng.Bottlerocket, d = tfBottleRocketValue.(BottleRocket6Value).Expand(ctx)
		diags = append(diags, d...)
	}
	if !v.InstancesDistribution6.IsNull() && !v.InstancesDistribution6.IsUnknown() {
		var instancesDistributionTypes InstancesDistribution6Type
		tfInstancesDistributionValue, d := instancesDistributionTypes.ValueFromObject(ctx, v.InstancesDistribution6)
		if d.HasError() {
			diags = append(diags, d...)
			return &ng, diags
		}
		ng.InstancesDistribution, d = tfInstancesDistributionValue.(InstancesDistribution6Value).Expand(ctx)
		diags = append(diags, d...)
	}

	if !v.AsgMetricsCollection6.IsNull() && !v.AsgMetricsCollection6.IsUnknown() {
		for _, amc := range v.AsgMetricsCollection6.Elements() {
			a, d := amc.(AsgMetricsCollection6Value).Expand(ctx)
			diags = append(diags, d...)
			ng.ASGMetricsCollection = append(ng.ASGMetricsCollection, a)
		}
	}

	if !v.Taints6.IsNull() && !v.Taints6.IsUnknown() {
		for _, taint := range v.Taints6.Elements() {
			t, d := taint.(Taints6Value).Expand(ctx)
			diags = append(diags, d...)
			ng.Taints = append(ng.Taints, t)

		}
	}

	if !v.UpdateConfig6.IsNull() && !v.UpdateConfig6.IsUnknown() {
		var updateConfigTypes UpdateConfig6Type
		tfUpdateConfigValue, d := updateConfigTypes.ValueFromObject(ctx, v.UpdateConfig6)
		if d.HasError() {
			diags = append(diags, d...)
			return &ng, diags
		}
		ng.UpdateConfig, d = tfUpdateConfigValue.(UpdateConfig6Value).Expand(ctx)
		diags = append(diags, d...)
	}

	if !v.KubeletExtraConfig6.IsNull() && !v.KubeletExtraConfig6.IsUnknown() {
		var kubeletExtraConfigTypes KubeletExtraConfig6Type
		tfKubeletExtraConfigValue, d := kubeletExtraConfigTypes.ValueFromObject(ctx, v.KubeletExtraConfig6)
		if d.HasError() {
			diags = append(diags, d...)
			return &ng, diags
		}
		ng.KubeletExtraConfig, d = tfKubeletExtraConfigValue.(KubeletExtraConfig6Value).Expand(ctx)
		diags = append(diags, d...)
	}

	if !v.SecurityGroups6.IsNull() && !v.SecurityGroups6.IsUnknown() {
		var securityGroupsTypes SecurityGroups6Type
		tfSecurityGroupsValue, d := securityGroupsTypes.ValueFromObject(ctx, v.SecurityGroups6)
		if d.HasError() {
			diags = append(diags, d...)
			return &ng, diags
		}
		ng.SecurityGroups, d = tfSecurityGroupsValue.(SecurityGroups6Value).Expand(ctx)
		diags = append(diags, d...)

	}

	return &ng, diags
}

func (v Iam6Value) Expand(ctx context.Context) (*rafay.NodeGroupIAM, diag.Diagnostics) {
	var diags, d diag.Diagnostics
	var iam rafay.NodeGroupIAM

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
	if !v.IamNodeGroupWithAddonPolicies6.IsNull() && !v.IamNodeGroupWithAddonPolicies6.IsUnknown() {
		var iamNodeGroupWithAddonPoliciesTypes IamNodeGroupWithAddonPolicies6Type
		tfIamNodeGroupWithAddonPoliciesValue, d := iamNodeGroupWithAddonPoliciesTypes.ValueFromObject(ctx, v.IamNodeGroupWithAddonPolicies6)
		if d.HasError() {
			diags = append(diags, d...)
			return &iam, diags
		}
		iam.WithAddonPolicies, d = tfIamNodeGroupWithAddonPoliciesValue.(IamNodeGroupWithAddonPolicies6Value).Expand(ctx)
		diags = append(diags, d...)
	}

	// Map attach_policy block (list)
	if !v.AttachPolicy6.IsNull() && !v.AttachPolicy6.IsUnknown() {
		var attachPolicyTypes AttachPolicy6Type
		tfAttachPolicyValue, d := attachPolicyTypes.ValueFromObject(ctx, v.AttachPolicy6)
		if d.HasError() {
			diags = append(diags, d...)
			return &iam, diags
		}
		iam.AttachPolicy, d = tfAttachPolicyValue.(AttachPolicy6Value).Expand(ctx)
		diags = append(diags, d...)
	}

	return &iam, diags
}

func (v IamNodeGroupWithAddonPolicies6Value) Expand(ctx context.Context) (*rafay.NodeGroupIAMAddonPolicies, diag.Diagnostics) {
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

func (v AttachPolicy6Value) Expand(ctx context.Context) (*rafay.InlineDocument, diag.Diagnostics) {
	var diags diag.Diagnostics
	var policy rafay.InlineDocument

	if v.IsNull() {
		return &rafay.InlineDocument{}, diags
	}

	if !v.Version.IsNull() && !v.Version.IsUnknown() {
		policy.Version = getStringValue(v.Version)
	}

	if !v.Statement6.IsNull() && !v.Statement6.IsUnknown() {
		for _, stmt := range v.Statement6.Elements() {
			s, d := stmt.(Statement6Value).Expand(ctx)
			diags = append(diags, d...)
			policy.Statement = append(policy.Statement, s)
		}
	}

	if !v.Id.IsNull() && !v.Id.IsUnknown() {
		policy.Id = getStringValue(v.Id)
	}

	return &policy, diags
}

func (v Statement6Value) Expand(ctx context.Context) (rafay.InlineStatement, diag.Diagnostics) {
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

func (v Ssh6Value) Expand(ctx context.Context) (*rafay.NodeGroupSSH, diag.Diagnostics) {
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
func (v Placement6Value) Expand(ctx context.Context) (*rafay.Placement, diag.Diagnostics) {
	var diags diag.Diagnostics
	var placement rafay.Placement

	if !v.Group.IsNull() && !v.Group.IsUnknown() {
		group := getStringValue(v.Group)
		placement.GroupName = group
	}

	return &placement, diags
}

func (v InstanceSelector6Value) Expand(ctx context.Context) (*rafay.InstanceSelector, diag.Diagnostics) {
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

func (v BottleRocket6Value) Expand(ctx context.Context) (*rafay.NodeGroupBottlerocket, diag.Diagnostics) {
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

func (v InstancesDistribution6Value) Expand(ctx context.Context) (*rafay.NodeGroupInstancesDistribution, diag.Diagnostics) {
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

func (v AsgMetricsCollection6Value) Expand(ctx context.Context) (rafay.MetricsCollection, diag.Diagnostics) {
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

func (v Taints6Value) Expand(ctx context.Context) (rafay.NodeGroupTaint, diag.Diagnostics) {
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

func (v UpdateConfig6Value) Expand(ctx context.Context) (*rafay.NodeGroupUpdateConfig, diag.Diagnostics) {
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

func (v KubeletExtraConfig6Value) Expand(ctx context.Context) (*rafay.KubeletExtraConfig, diag.Diagnostics) {
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

// SecurityGroups6Value Expand
func (v SecurityGroups6Value) Expand(ctx context.Context) (*rafay.NodeGroupSGs, diag.Diagnostics) {
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
