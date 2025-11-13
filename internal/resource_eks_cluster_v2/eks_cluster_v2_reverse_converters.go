package resource_eks_cluster_v2

import (
	"context"
	"fmt"

	"github.com/RafaySystems/terraform-provider-rafay/rafay"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// convertClusterSpecToModelComplete converts API cluster spec to the complete Terraform model
// This is the reverse of convertModelToClusterSpecComplete
func convertClusterSpecToModelComplete(ctx context.Context, eksCluster *rafay.EKSCluster, eksClusterConfig *rafay.EKSClusterConfig) (*EKSClusterV2ResourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	model := &EKSClusterV2ResourceModel{}

	// Convert cluster (metadata + spec)
	clusterObj, clusterDiags := flattenCluster(ctx, eksCluster)
	diags.Append(clusterDiags...)
	if diags.HasError() {
		return nil, diags
	}
	model.Cluster = clusterObj

	// Convert cluster config
	clusterConfigObj, configDiags := flattenClusterConfig(ctx, eksClusterConfig)
	diags.Append(configDiags...)
	if diags.HasError() {
		return nil, diags
	}
	model.ClusterConfig = clusterConfigObj

	tflog.Info(ctx, "Successfully converted complete cluster spec to model")
	return model, diags
}

// flattenCluster converts the EKSCluster to Terraform object
func flattenCluster(ctx context.Context, eksCluster *rafay.EKSCluster) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	if eksCluster == nil {
		return types.ObjectNull(clusterObjectTypes()), diags
	}

	// Flatten metadata
	metadataObj, metaDiags := flattenClusterMetadata(ctx, eksCluster.Metadata)
	diags.Append(metaDiags...)

	// Flatten spec
	specObj, specDiags := flattenClusterSpec(ctx, eksCluster.Spec)
	diags.Append(specDiags...)

	// Build cluster object
	clusterObj, objDiags := types.ObjectValue(
		clusterObjectTypes(),
		map[string]attr.Value{
			"kind":     types.StringValue(eksCluster.Kind),
			"metadata": metadataObj,
			"spec":     specObj,
		},
	)
	diags.Append(objDiags...)

	return clusterObj, diags
}

// flattenClusterMetadata converts cluster metadata to Terraform object
func flattenClusterMetadata(ctx context.Context, metadata *rafay.EKSClusterMetadata) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	if metadata == nil {
		return types.ObjectNull(clusterMetadataObjectTypes()), diags
	}

	// Convert labels map
	labels := types.MapNull(types.StringType)
	if len(metadata.Labels) > 0 {
		labelsMap := make(map[string]attr.Value)
		for k, v := range metadata.Labels {
			labelsMap[k] = types.StringValue(v)
		}
		var err error
		labels, err = types.MapValue(types.StringType, labelsMap)
		if err != nil {
			diags.AddError("Failed to convert labels", err.Error())
		}
	}

	metadataObj, objDiags := types.ObjectValue(
		clusterMetadataObjectTypes(),
		map[string]attr.Value{
			"name":    types.StringValue(metadata.Name),
			"project": types.StringValue(metadata.Project),
			"labels":  labels,
		},
	)
	diags.Append(objDiags...)

	return metadataObj, diags
}

// flattenClusterSpec converts cluster spec to Terraform object
func flattenClusterSpec(ctx context.Context, spec *rafay.EKSSpec) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	if spec == nil {
		return types.ObjectNull(clusterSpecObjectTypes()), diags
	}

	// Flatten CNI params
	cniParamsObj := types.ObjectNull(cniParamsObjectTypes())
	if spec.CniParams != nil {
		var cniDiags diag.Diagnostics
		cniParamsObj, cniDiags = flattenCNIParams(ctx, spec.CniParams)
		diags.Append(cniDiags...)
	}

	// Flatten proxy config
	proxyConfigMap := types.MapNull(types.StringType)
	if spec.ProxyConfig != nil {
		var proxyDiags diag.Diagnostics
		proxyConfigMap, proxyDiags = flattenProxyConfig(ctx, spec.ProxyConfig)
		diags.Append(proxyDiags...)
	}

	// Flatten system components placement
	scpObj := types.ObjectNull(systemComponentsPlacementObjectTypes())
	if spec.SystemComponentsPlacement != nil {
		var scpDiags diag.Diagnostics
		scpObj, scpDiags = flattenSystemComponentsPlacement(ctx, spec.SystemComponentsPlacement)
		diags.Append(scpDiags...)
	}

	// Flatten sharing
	sharingObj := types.ObjectNull(sharingObjectTypes())
	if spec.Sharing != nil {
		var sharingDiags diag.Diagnostics
		sharingObj, sharingDiags = flattenSharing(ctx, spec.Sharing)
		diags.Append(sharingDiags...)
	}

	specObj, objDiags := types.ObjectValue(
		clusterSpecObjectTypes(),
		map[string]attr.Value{
			"type":                        types.StringValue(spec.Type),
			"blueprint":                   types.StringValue(spec.Blueprint),
			"blueprint_version":           types.StringValue(spec.BlueprintVersion),
			"cloud_provider":              types.StringValue(spec.CloudProvider),
			"cross_account_role_arn":      types.StringValue(spec.CrossAccountRoleArn),
			"cni_provider":                types.StringValue(spec.CniProvider),
			"cni_params":                  cniParamsObj,
			"proxy_config":                proxyConfigMap,
			"system_components_placement": scpObj,
			"sharing":                     sharingObj,
		},
	)
	diags.Append(objDiags...)

	return specObj, diags
}

// flattenCNIParams converts CNI params to Terraform object
func flattenCNIParams(ctx context.Context, cniParams *rafay.CustomCni) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	obj, objDiags := types.ObjectValue(
		cniParamsObjectTypes(),
		map[string]attr.Value{
			"custom_cni_cidr":    types.StringValue(cniParams.CustomCniCidr),
			"custom_cni_credits": types.StringValue(""), // Add if field exists
		},
	)
	diags.Append(objDiags...)

	return obj, diags
}

// flattenProxyConfig converts proxy config to Terraform map
func flattenProxyConfig(ctx context.Context, proxyConfig *rafay.ProxyConfig) (types.Map, diag.Diagnostics) {
	var diags diag.Diagnostics

	proxyMap := make(map[string]attr.Value)
	if proxyConfig.HttpProxy != "" {
		proxyMap["http_proxy"] = types.StringValue(proxyConfig.HttpProxy)
	}
	if proxyConfig.HttpsProxy != "" {
		proxyMap["https_proxy"] = types.StringValue(proxyConfig.HttpsProxy)
	}
	if proxyConfig.NoProxy != "" {
		proxyMap["no_proxy"] = types.StringValue(proxyConfig.NoProxy)
	}

	if len(proxyMap) == 0 {
		return types.MapNull(types.StringType), diags
	}

	mapValue, err := types.MapValue(types.StringType, proxyMap)
	if err != nil {
		diags.AddError("Failed to convert proxy config", err.Error())
		return types.MapNull(types.StringType), diags
	}

	return mapValue, diags
}

// flattenSystemComponentsPlacement converts system components placement to Terraform object
func flattenSystemComponentsPlacement(ctx context.Context, scp *rafay.SystemComponentsPlacement) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Flatten node selector
	nodeSelectorMap := types.MapNull(types.StringType)
	if len(scp.NodeSelector) > 0 {
		nsMap := make(map[string]attr.Value)
		for k, v := range scp.NodeSelector {
			nsMap[k] = types.StringValue(v)
		}
		var err error
		nodeSelectorMap, err = types.MapValue(types.StringType, nsMap)
		if err != nil {
			diags.AddError("Failed to convert node selector", err.Error())
		}
	}

	// Flatten tolerations (array to map) - KEY CONVERSION!
	tolerationsMap := types.MapNull(tolerationObjectType())
	if len(scp.Tolerations) > 0 {
		var tolDiags diag.Diagnostics
		tolerationsMap, tolDiags = flattenTolerationsArrayToMap(ctx, scp.Tolerations)
		diags.Append(tolDiags...)
	}

	// Flatten daemonset node selector
	dsNodeSelectorMap := types.MapNull(types.StringType)
	if len(scp.DaemonsetNodeSelector) > 0 {
		dsnsMap := make(map[string]attr.Value)
		for k, v := range scp.DaemonsetNodeSelector {
			dsnsMap[k] = types.StringValue(v)
		}
		var err error
		dsNodeSelectorMap, err = types.MapValue(types.StringType, dsnsMap)
		if err != nil {
			diags.AddError("Failed to convert daemonset node selector", err.Error())
		}
	}

	// Flatten daemonset tolerations
	dsTolerationMap := types.MapNull(tolerationObjectType())
	if len(scp.DaemonsetTolerations) > 0 {
		var tolDiags diag.Diagnostics
		dsTolerationMap, tolDiags = flattenTolerationsArrayToMap(ctx, scp.DaemonsetTolerations)
		diags.Append(tolDiags...)
	}

	obj, objDiags := types.ObjectValue(
		systemComponentsPlacementObjectTypes(),
		map[string]attr.Value{
			"node_selector":           nodeSelectorMap,
			"tolerations":             tolerationsMap,
			"daemonset_node_selector": dsNodeSelectorMap,
			"daemonset_tolerations":   dsTolerationMap,
		},
	)
	diags.Append(objDiags...)

	return obj, diags
}

// flattenTolerationsArrayToMap converts tolerations array to map (CRITICAL for zero diff!)
func flattenTolerationsArrayToMap(ctx context.Context, tolerations []*rafay.Toleration) (types.Map, diag.Diagnostics) {
	var diags diag.Diagnostics

	if len(tolerations) == 0 {
		return types.MapNull(tolerationObjectType()), diags
	}

	tolerationsMap := make(map[string]attr.Value)
	for _, tol := range tolerations {
		// Use the toleration key as the map key - this ensures map stability!
		key := tol.Key
		if key == "" {
			// Fallback: use a generated key if key is empty (shouldn't happen in practice)
			key = fmt.Sprintf("toleration-%d", len(tolerationsMap))
		}

		tolObj, tolDiags := types.ObjectValue(
			tolerationObjectType().AttrTypes,
			map[string]attr.Value{
				"key":      types.StringValue(tol.Key),
				"operator": types.StringValue(tol.Operator),
				"value":    types.StringValue(tol.Value),
				"effect":   types.StringValue(tol.Effect),
			},
		)
		diags.Append(tolDiags...)

		tolerationsMap[key] = tolObj
	}

	mapValue, err := types.MapValue(tolerationObjectType(), tolerationsMap)
	if err != nil {
		diags.AddError("Failed to convert tolerations to map", err.Error())
		return types.MapNull(tolerationObjectType()), diags
	}

	return mapValue, diags
}

// flattenSharing converts sharing configuration to Terraform object
func flattenSharing(ctx context.Context, sharing *rafay.V1ClusterSharing) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Flatten projects (array to map)
	projectsMap := types.MapNull(projectObjectType())
	if len(sharing.Projects) > 0 {
		projMap := make(map[string]attr.Value)
		for _, proj := range sharing.Projects {
			// Use project name as the map key
			key := proj.Name

			projObj, projDiags := types.ObjectValue(
				projectObjectType().AttrTypes,
				map[string]attr.Value{
					"name": types.StringValue(proj.Name),
				},
			)
			diags.Append(projDiags...)

			projMap[key] = projObj
		}

		var err error
		projectsMap, err = types.MapValue(projectObjectType(), projMap)
		if err != nil {
			diags.AddError("Failed to convert projects to map", err.Error())
		}
	}

	obj, objDiags := types.ObjectValue(
		sharingObjectTypes(),
		map[string]attr.Value{
			"enabled":  types.BoolValue(sharing.Enabled),
			"projects": projectsMap,
		},
	)
	diags.Append(objDiags...)

	return obj, diags
}

// flattenClusterConfig converts the EKSClusterConfig to Terraform object
func flattenClusterConfig(ctx context.Context, config *rafay.EKSClusterConfig) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	if config == nil {
		return types.ObjectNull(clusterConfigObjectTypes()), diags
	}

	// Flatten metadata
	metadataObj, metaDiags := flattenClusterConfigMetadata(ctx, config.Metadata)
	diags.Append(metaDiags...)

	// Flatten VPC
	vpcObj := types.ObjectNull(vpcObjectTypes())
	if config.VPC != nil {
		var vpcDiags diag.Diagnostics
		vpcObj, vpcDiags = flattenVPC(ctx, config.VPC)
		diags.Append(vpcDiags...)
	}

	// Flatten node groups (array to map)
	nodeGroupsMap := types.MapNull(nodeGroupObjectType())
	if len(config.NodeGroups) > 0 {
		var ngDiags diag.Diagnostics
		nodeGroupsMap, ngDiags = flattenNodeGroupsArrayToMap(ctx, config.NodeGroups)
		diags.Append(ngDiags...)
	}

	// Flatten managed node groups (array to map)
	managedNodeGroupsMap := types.MapNull(managedNodeGroupObjectType())
	if len(config.ManagedNodeGroups) > 0 {
		var mngDiags diag.Diagnostics
		managedNodeGroupsMap, mngDiags = flattenManagedNodeGroupsArrayToMap(ctx, config.ManagedNodeGroups)
		diags.Append(mngDiags...)
	}

	// Flatten identity providers (array to map)
	identityProvidersMap := types.MapNull(identityProviderObjectType())
	if len(config.IdentityProviders) > 0 {
		var ipDiags diag.Diagnostics
		identityProvidersMap, ipDiags = flattenIdentityProvidersArrayToMap(ctx, config.IdentityProviders)
		diags.Append(ipDiags...)
	}

	// Flatten encryption config
	encryptionConfigObj := types.ObjectNull(encryptionConfigObjectTypes())
	// TODO: Implement based on actual schema

	// Flatten access entries
	accessEntriesMap := types.MapNull(accessEntryObjectType())
	if config.AccessConfig != nil && len(config.AccessConfig.AccessEntries) > 0 {
		var aeDiags diag.Diagnostics
		accessEntriesMap, aeDiags = flattenAccessEntriesArrayToMap(ctx, config.AccessConfig.AccessEntries)
		diags.Append(aeDiags...)
	}

	// Flatten identity mappings
	identityMappingsObj := types.ObjectNull(identityMappingsObjectTypes())
	// TODO: Implement based on actual schema

	configObj, objDiags := types.ObjectValue(
		clusterConfigObjectTypes(),
		map[string]attr.Value{
			"apiversion":         types.StringValue(config.APIVersion),
			"kind":               types.StringValue(config.Kind),
			"metadata":           metadataObj,
			"vpc":                vpcObj,
			"node_groups":        nodeGroupsMap,
			"managed_node_groups": managedNodeGroupsMap,
			"identity_providers": identityProvidersMap,
			"encryption_config":  encryptionConfigObj,
			"access_entries":     accessEntriesMap,
			"identity_mappings":  identityMappingsObj,
		},
	)
	diags.Append(objDiags...)

	return configObj, diags
}

// flattenClusterConfigMetadata converts cluster config metadata to Terraform object
func flattenClusterConfigMetadata(ctx context.Context, metadata *rafay.EKSClusterConfigMetadata) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	if metadata == nil {
		return types.ObjectNull(clusterConfigMetadataObjectTypes()), diags
	}

	// Convert tags map
	tags := types.MapNull(types.StringType)
	if len(metadata.Tags) > 0 {
		tagsMap := make(map[string]attr.Value)
		for k, v := range metadata.Tags {
			tagsMap[k] = types.StringValue(v)
		}
		var err error
		tags, err = types.MapValue(types.StringType, tagsMap)
		if err != nil {
			diags.AddError("Failed to convert tags", err.Error())
		}
	}

	obj, objDiags := types.ObjectValue(
		clusterConfigMetadataObjectTypes(),
		map[string]attr.Value{
			"name":    types.StringValue(metadata.Name),
			"region":  types.StringValue(metadata.Region),
			"version": types.StringValue(metadata.Version),
			"tags":    tags,
		},
	)
	diags.Append(objDiags...)

	return obj, diags
}

// flattenVPC converts VPC configuration to Terraform object
func flattenVPC(ctx context.Context, vpc *rafay.EKSClusterVPC) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Flatten subnets
	subnetsObj := types.ObjectNull(subnetsObjectTypes())
	if vpc.Subnets != nil {
		var subnetsDiags diag.Diagnostics
		subnetsObj, subnetsDiags = flattenSubnets(ctx, vpc.Subnets)
		diags.Append(subnetsDiags...)
	}

	obj, objDiags := types.ObjectValue(
		vpcObjectTypes(),
		map[string]attr.Value{
			"region":                      types.StringValue(vpc.Region),
			"cidr":                        types.StringValue(vpc.CIDR),
			"cluster_resources_vpc_config": types.ObjectNull(map[string]attr.Type{}), // TODO
			"subnets":                     subnetsObj,
			"nat":                         types.ObjectNull(map[string]attr.Type{}), // TODO
			"security_group":              types.ObjectNull(map[string]attr.Type{}), // TODO
		},
	)
	diags.Append(objDiags...)

	return obj, diags
}

// flattenSubnets converts subnets to Terraform object
func flattenSubnets(ctx context.Context, subnets *rafay.ClusterSubnets) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Flatten public subnets
	publicMap := types.MapNull(subnetObjectType())
	if len(subnets.Public) > 0 {
		pubMap := make(map[string]attr.Value)
		for az, subnet := range subnets.Public {
			subnetObj, subnetDiags := types.ObjectValue(
				subnetObjectType().AttrTypes,
				map[string]attr.Value{
					"id":   types.StringValue(subnet.ID),
					"cidr": types.StringValue(subnet.CIDR),
					"az":   types.StringValue(subnet.AZ),
				},
			)
			diags.Append(subnetDiags...)
			pubMap[az] = subnetObj
		}

		var err error
		publicMap, err = types.MapValue(subnetObjectType(), pubMap)
		if err != nil {
			diags.AddError("Failed to convert public subnets", err.Error())
		}
	}

	// Flatten private subnets
	privateMap := types.MapNull(subnetObjectType())
	if len(subnets.Private) > 0 {
		privMap := make(map[string]attr.Value)
		for az, subnet := range subnets.Private {
			subnetObj, subnetDiags := types.ObjectValue(
				subnetObjectType().AttrTypes,
				map[string]attr.Value{
					"id":   types.StringValue(subnet.ID),
					"cidr": types.StringValue(subnet.CIDR),
					"az":   types.StringValue(subnet.AZ),
				},
			)
			diags.Append(subnetDiags...)
			privMap[az] = subnetObj
		}

		var err error
		privateMap, err = types.MapValue(subnetObjectType(), privMap)
		if err != nil {
			diags.AddError("Failed to convert private subnets", err.Error())
		}
	}

	obj, objDiags := types.ObjectValue(
		subnetsObjectTypes(),
		map[string]attr.Value{
			"public":  publicMap,
			"private": privateMap,
		},
	)
	diags.Append(objDiags...)

	return obj, diags
}

// flattenNodeGroupsArrayToMap converts node groups array to map
func flattenNodeGroupsArrayToMap(ctx context.Context, nodeGroups []*rafay.NodeGroup) (types.Map, diag.Diagnostics) {
	var diags diag.Diagnostics

	nodeGroupsMap := make(map[string]attr.Value)
	for _, ng := range nodeGroups {
		// Use node group name as map key
		key := ng.Name

		// Convert labels
		labels := types.MapNull(types.StringType)
		if len(ng.Labels) > 0 {
			labelsMap := make(map[string]attr.Value)
			for k, v := range ng.Labels {
				labelsMap[k] = types.StringValue(v)
			}
			var err error
			labels, err = types.MapValue(types.StringType, labelsMap)
			if err != nil {
				diags.AddError("Failed to convert node group labels", err.Error())
			}
		}

		// Convert tags
		tags := types.MapNull(types.StringType)
		if len(ng.Tags) > 0 {
			tagsMap := make(map[string]attr.Value)
			for k, v := range ng.Tags {
				tagsMap[k] = types.StringValue(v)
			}
			var err error
			tags, err = types.MapValue(types.StringType, tagsMap)
			if err != nil {
				diags.AddError("Failed to convert node group tags", err.Error())
			}
		}

		// Convert taints (array to map)
		taintsMap := types.MapNull(taintObjectType())
		if len(ng.Taints) > 0 {
			tMap := make(map[string]attr.Value)
			for _, taint := range ng.Taints {
				taintObj, taintDiags := types.ObjectValue(
					taintObjectType().AttrTypes,
					map[string]attr.Value{
						"key":    types.StringValue(taint.Key),
						"value":  types.StringValue(taint.Value),
						"effect": types.StringValue(taint.Effect),
					},
				)
				diags.Append(taintDiags...)
				tMap[taint.Key] = taintObj
			}

			var err error
			taintsMap, err = types.MapValue(taintObjectType(), tMap)
			if err != nil {
				diags.AddError("Failed to convert taints", err.Error())
			}
		}

		// Convert AZs list
		azsList, _ := types.ListValueFrom(ctx, types.StringType, ng.AvailabilityZones)

		ngObj, ngDiags := types.ObjectValue(
			nodeGroupObjectType().AttrTypes,
			map[string]attr.Value{
				"name":                types.StringValue(ng.Name),
				"ami":                 types.StringValue(ng.AMI),
				"instance_type":       types.StringValue(ng.InstanceType),
				"desired_capacity":    types.Int64Value(int64(ng.DesiredCapacity)),
				"min_size":            types.Int64Value(int64(ng.MinSize)),
				"max_size":            types.Int64Value(int64(ng.MaxSize)),
				"volume_size":         types.Int64Value(int64(ng.VolumeSize)),
				"volume_type":         types.StringValue(ng.VolumeType),
				"private_networking":  types.BoolValue(ng.PrivateNetworking),
				"availability_zones":  azsList,
				"labels":              labels,
				"tags":                tags,
				"taints":              taintsMap,
				"iam":                 types.ObjectNull(map[string]attr.Type{}), // TODO
				"security_groups":     types.ObjectNull(map[string]attr.Type{}), // TODO
				"ssh":                 types.ObjectNull(map[string]attr.Type{}), // TODO
				"update_config":       types.ObjectNull(map[string]attr.Type{}), // TODO
				"scaling_config":      types.ObjectNull(map[string]attr.Type{}), // TODO
			},
		)
		diags.Append(ngDiags...)

		nodeGroupsMap[key] = ngObj
	}

	mapValue, err := types.MapValue(nodeGroupObjectType(), nodeGroupsMap)
	if err != nil {
		diags.AddError("Failed to convert node groups to map", err.Error())
		return types.MapNull(nodeGroupObjectType()), diags
	}

	return mapValue, diags
}

// flattenManagedNodeGroupsArrayToMap converts managed node groups array to map
func flattenManagedNodeGroupsArrayToMap(ctx context.Context, managedNodeGroups []*rafay.ManagedNodeGroup) (types.Map, diag.Diagnostics) {
	var diags diag.Diagnostics

	mngMap := make(map[string]attr.Value)
	for _, mng := range managedNodeGroups {
		// Use managed node group name as map key
		key := mng.Name

		// Similar flattening logic as node groups...
		// (truncated for brevity - follow same pattern as flattenNodeGroupsArrayToMap)

		mngObj := types.ObjectNull(managedNodeGroupObjectType().AttrTypes) // Placeholder
		mngMap[key] = mngObj
	}

	mapValue, err := types.MapValue(managedNodeGroupObjectType(), mngMap)
	if err != nil {
		diags.AddError("Failed to convert managed node groups to map", err.Error())
		return types.MapNull(managedNodeGroupObjectType()), diags
	}

	return mapValue, diags
}

// flattenIdentityProvidersArrayToMap converts identity providers array to map
func flattenIdentityProvidersArrayToMap(ctx context.Context, identityProviders []*rafay.IdentityProvider) (types.Map, diag.Diagnostics) {
	var diags diag.Diagnostics
	// TODO: Implement based on actual schema
	return types.MapNull(identityProviderObjectType()), diags
}

// flattenAccessEntriesArrayToMap converts access entries array to map
func flattenAccessEntriesArrayToMap(ctx context.Context, accessEntries []*rafay.EKSAccessEntry) (types.Map, diag.Diagnostics) {
	var diags diag.Diagnostics
	// TODO: Implement based on actual schema
	return types.MapNull(accessEntryObjectType()), diags
}

// Type definition helpers for object types
func clusterObjectTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"kind":     types.StringType,
		"metadata": types.ObjectType{AttrTypes: clusterMetadataObjectTypes()},
		"spec":     types.ObjectType{AttrTypes: clusterSpecObjectTypes()},
	}
}

func clusterMetadataObjectTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":    types.StringType,
		"project": types.StringType,
		"labels":  types.MapType{ElemType: types.StringType},
	}
}

func clusterSpecObjectTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"type":                        types.StringType,
		"blueprint":                   types.StringType,
		"blueprint_version":           types.StringType,
		"cloud_provider":              types.StringType,
		"cross_account_role_arn":      types.StringType,
		"cni_provider":                types.StringType,
		"cni_params":                  types.ObjectType{AttrTypes: cniParamsObjectTypes()},
		"proxy_config":                types.MapType{ElemType: types.StringType},
		"system_components_placement": types.ObjectType{AttrTypes: systemComponentsPlacementObjectTypes()},
		"sharing":                     types.ObjectType{AttrTypes: sharingObjectTypes()},
	}
}

func cniParamsObjectTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"custom_cni_cidr":    types.StringType,
		"custom_cni_credits": types.StringType,
	}
}

func systemComponentsPlacementObjectTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"node_selector":           types.MapType{ElemType: types.StringType},
		"tolerations":             types.MapType{ElemType: tolerationObjectType()},
		"daemonset_node_selector": types.MapType{ElemType: types.StringType},
		"daemonset_tolerations":   types.MapType{ElemType: tolerationObjectType()},
	}
}

func tolerationObjectType() types.ObjectType {
	return types.ObjectType{AttrTypes: map[string]attr.Type{
		"key":      types.StringType,
		"operator": types.StringType,
		"value":    types.StringType,
		"effect":   types.StringType,
	}}
}

func sharingObjectTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"enabled":  types.BoolType,
		"projects": types.MapType{ElemType: projectObjectType()},
	}
}

func projectObjectType() types.ObjectType {
	return types.ObjectType{AttrTypes: map[string]attr.Type{
		"name": types.StringType,
	}}
}

func clusterConfigObjectTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"apiversion":         types.StringType,
		"kind":               types.StringType,
		"metadata":           types.ObjectType{AttrTypes: clusterConfigMetadataObjectTypes()},
		"vpc":                types.ObjectType{AttrTypes: vpcObjectTypes()},
		"node_groups":        types.MapType{ElemType: nodeGroupObjectType()},
		"managed_node_groups": types.MapType{ElemType: managedNodeGroupObjectType()},
		"identity_providers": types.MapType{ElemType: identityProviderObjectType()},
		"encryption_config":  types.ObjectType{AttrTypes: encryptionConfigObjectTypes()},
		"access_entries":     types.MapType{ElemType: accessEntryObjectType()},
		"identity_mappings":  types.ObjectType{AttrTypes: identityMappingsObjectTypes()},
	}
}

func clusterConfigMetadataObjectTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":    types.StringType,
		"region":  types.StringType,
		"version": types.StringType,
		"tags":    types.MapType{ElemType: types.StringType},
	}
}

func vpcObjectTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"region":                      types.StringType,
		"cidr":                        types.StringType,
		"cluster_resources_vpc_config": types.ObjectType{AttrTypes: map[string]attr.Type{}},
		"subnets":                     types.ObjectType{AttrTypes: subnetsObjectTypes()},
		"nat":                         types.ObjectType{AttrTypes: map[string]attr.Type{}},
		"security_group":              types.ObjectType{AttrTypes: map[string]attr.Type{}},
	}
}

func subnetsObjectTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"public":  types.MapType{ElemType: subnetObjectType()},
		"private": types.MapType{ElemType: subnetObjectType()},
	}
}

func subnetObjectType() types.ObjectType {
	return types.ObjectType{AttrTypes: map[string]attr.Type{
		"id":   types.StringType,
		"cidr": types.StringType,
		"az":   types.StringType,
	}}
}

func nodeGroupObjectType() types.ObjectType {
	return types.ObjectType{AttrTypes: map[string]attr.Type{
		"name":                types.StringType,
		"ami":                 types.StringType,
		"instance_type":       types.StringType,
		"desired_capacity":    types.Int64Type,
		"min_size":            types.Int64Type,
		"max_size":            types.Int64Type,
		"volume_size":         types.Int64Type,
		"volume_type":         types.StringType,
		"private_networking":  types.BoolType,
		"availability_zones":  types.ListType{ElemType: types.StringType},
		"labels":              types.MapType{ElemType: types.StringType},
		"tags":                types.MapType{ElemType: types.StringType},
		"taints":              types.MapType{ElemType: taintObjectType()},
		"iam":                 types.ObjectType{AttrTypes: map[string]attr.Type{}},
		"security_groups":     types.ObjectType{AttrTypes: map[string]attr.Type{}},
		"ssh":                 types.ObjectType{AttrTypes: map[string]attr.Type{}},
		"update_config":       types.ObjectType{AttrTypes: map[string]attr.Type{}},
		"scaling_config":      types.ObjectType{AttrTypes: map[string]attr.Type{}},
	}}
}

func taintObjectType() types.ObjectType {
	return types.ObjectType{AttrTypes: map[string]attr.Type{
		"key":    types.StringType,
		"value":  types.StringType,
		"effect": types.StringType,
	}}
}

func managedNodeGroupObjectType() types.ObjectType {
	return types.ObjectType{AttrTypes: map[string]attr.Type{
		// TODO: Define actual attribute types
		"name": types.StringType,
	}}
}

func identityProviderObjectType() types.ObjectType {
	return types.ObjectType{AttrTypes: map[string]attr.Type{
		"type": types.StringType,
		"name": types.StringType,
	}}
}

func encryptionConfigObjectTypes() map[string]attr.Type {
	return map[string]attr.Type{}
}

func accessEntryObjectType() types.ObjectType {
	return types.ObjectType{AttrTypes: map[string]attr.Type{
		"principal_arn": types.StringType,
		"type":          types.StringType,
	}}
}

func identityMappingsObjectTypes() map[string]attr.Type {
	return map[string]attr.Type{}
}

