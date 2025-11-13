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

// Complete implementation of convertModelToClusterSpec
// This replaces the partial implementation in eks_cluster_v2_helpers.go

// convertModelToClusterSpecComplete converts the full Terraform model to API cluster spec
func convertModelToClusterSpecComplete(ctx context.Context, data *EKSClusterV2ResourceModel) (*rafay.EKSCluster, *rafay.EKSClusterConfig, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Extract cluster metadata
	var clusterModel ClusterModel
	diags.Append(data.Cluster.As(ctx, &clusterModel, types.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil, nil, diags
	}

	// Build EKSCluster (first YAML document)
	eksCluster := &rafay.EKSCluster{
		Kind: clusterModel.Kind.ValueString(),
	}

	// Convert metadata
	eksCluster.Metadata = convertClusterMetadata(ctx, clusterModel.Metadata, &diags)
	if diags.HasError() {
		return nil, nil, diags
	}

	// Convert spec
	eksCluster.Spec = convertClusterSpec(ctx, clusterModel.Spec, &diags)
	if diags.HasError() {
		return nil, nil, diags
	}

	// Build EKSClusterConfig (second YAML document)
	var clusterConfigModel ClusterConfigModel
	diags.Append(data.ClusterConfig.As(ctx, &clusterConfigModel, types.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil, nil, diags
	}

	eksClusterConfig := convertClusterConfig(ctx, &clusterConfigModel, &diags)
	if diags.HasError() {
		return nil, nil, diags
	}

	tflog.Info(ctx, "Successfully converted complete model to cluster spec")
	return eksCluster, eksClusterConfig, diags
}

// convertClusterMetadata converts cluster metadata
func convertClusterMetadata(ctx context.Context, metadataObj types.Object, diags *diag.Diagnostics) *rafay.EKSClusterMetadata {
	if metadataObj.IsNull() || metadataObj.IsUnknown() {
		return nil
	}

	var metadataModel ClusterMetadataModel
	diags.Append(metadataObj.As(ctx, &metadataModel, types.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}

	metadata := &rafay.EKSClusterMetadata{
		Name:    metadataModel.Name.ValueString(),
		Project: metadataModel.Project.ValueString(),
	}

	// Extract labels map
	if !metadataModel.Labels.IsNull() && !metadataModel.Labels.IsUnknown() {
		labels := make(map[string]string)
		*diags = metadataModel.Labels.ElementsAs(ctx, &labels, false)
		if diags.HasError() {
			return nil
		}
		metadata.Labels = labels
	}

	return metadata
}

// convertClusterSpec converts cluster spec
func convertClusterSpec(ctx context.Context, specObj types.Object, diags *diag.Diagnostics) *rafay.EKSSpec {
	if specObj.IsNull() || specObj.IsUnknown() {
		return nil
	}

	var specModel ClusterSpecModel
	diags.Append(specObj.As(ctx, &specModel, types.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}

	spec := &rafay.EKSSpec{
		Type:                specModel.Type.ValueString(),
		Blueprint:           specModel.Blueprint.ValueString(),
		BlueprintVersion:    specModel.BlueprintVersion.ValueString(),
		CloudProvider:       specModel.CloudProvider.ValueString(),
		CrossAccountRoleArn: specModel.CrossAccountRoleArn.ValueString(),
		CniProvider:         specModel.CniProvider.ValueString(),
	}

	// Convert CNI params
	if !specModel.CniParams.IsNull() && !specModel.CniParams.IsUnknown() {
		spec.CniParams = convertCNIParams(ctx, specModel.CniParams, diags)
	}

	// Convert proxy config
	if !specModel.ProxyConfig.IsNull() && !specModel.ProxyConfig.IsUnknown() {
		spec.ProxyConfig = convertProxyConfig(ctx, specModel.ProxyConfig, diags)
	}

	// Convert system components placement
	if !specModel.SystemComponentsPlacement.IsNull() && !specModel.SystemComponentsPlacement.IsUnknown() {
		spec.SystemComponentsPlacement = convertSystemComponentsPlacement(ctx, specModel.SystemComponentsPlacement, diags)
	}

	// Convert sharing
	if !specModel.Sharing.IsNull() && !specModel.Sharing.IsUnknown() {
		spec.Sharing = convertSharing(ctx, specModel.Sharing, diags)
	}

	return spec
}

// convertCNIParams converts CNI parameters
func convertCNIParams(ctx context.Context, cniParamsObj types.Object, diags *diag.Diagnostics) *rafay.CustomCni {
	var cniParams CNIParamsModel
	diags.Append(cniParamsObj.As(ctx, &cniParams, types.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}

	return &rafay.CustomCni{
		CustomCniCidr: cniParams.CustomCniCidr.ValueString(),
		// CustomCniCredits would go here if needed
	}
}

// convertProxyConfig converts proxy configuration from map
func convertProxyConfig(ctx context.Context, proxyConfigMap types.Map, diags *diag.Diagnostics) *rafay.ProxyConfig {
	proxyConfig := make(map[string]string)
	*diags = proxyConfigMap.ElementsAs(ctx, &proxyConfig, false)
	if diags.HasError() {
		return nil
	}

	return &rafay.ProxyConfig{
		HttpProxy:  proxyConfig["http_proxy"],
		HttpsProxy: proxyConfig["https_proxy"],
		NoProxy:    proxyConfig["no_proxy"],
	}
}

// convertSystemComponentsPlacement converts system components placement
func convertSystemComponentsPlacement(ctx context.Context, scpObj types.Object, diags *diag.Diagnostics) *rafay.SystemComponentsPlacement {
	var scpModel SystemComponentsPlacementModel
	diags.Append(scpObj.As(ctx, &scpModel, types.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}

	scp := &rafay.SystemComponentsPlacement{}

	// Convert node selector (map)
	if !scpModel.NodeSelector.IsNull() && !scpModel.NodeSelector.IsUnknown() {
		nodeSelector := make(map[string]string)
		*diags = scpModel.NodeSelector.ElementsAs(ctx, &nodeSelector, false)
		if !diags.HasError() {
			scp.NodeSelector = nodeSelector
		}
	}

	// Convert tolerations (map to array for API)
	if !scpModel.Tolerations.IsNull() && !scpModel.Tolerations.IsUnknown() {
		scp.Tolerations = convertTolerationsMapToArray(ctx, scpModel.Tolerations, diags)
	}

	// Convert daemonset node selector
	if !scpModel.DaemonsetNodeSelector.IsNull() && !scpModel.DaemonsetNodeSelector.IsUnknown() {
		dsNodeSelector := make(map[string]string)
		*diags = scpModel.DaemonsetNodeSelector.ElementsAs(ctx, &dsNodeSelector, false)
		if !diags.HasError() {
			scp.DaemonsetNodeSelector = dsNodeSelector
		}
	}

	// Convert daemonset tolerations
	if !scpModel.DaemonsetTolerations.IsNull() && !scpModel.DaemonsetTolerations.IsUnknown() {
		scp.DaemonsetTolerations = convertTolerationsMapToArray(ctx, scpModel.DaemonsetTolerations, diags)
	}

	return scp
}

// convertTolerationsMapToArray converts tolerations from map to array
func convertTolerationsMapToArray(ctx context.Context, tolerationsMap types.Map, diags *diag.Diagnostics) []*rafay.Toleration {
	tolerationsMapData := make(map[string]types.Object)
	*diags = tolerationsMap.ElementsAs(ctx, &tolerationsMapData, false)
	if diags.HasError() {
		return nil
	}

	tolerations := make([]*rafay.Toleration, 0, len(tolerationsMapData))
	for _, tolerationObj := range tolerationsMapData {
		var toleration TolerationModel
		diags.Append(tolerationObj.As(ctx, &toleration, types.ObjectAsOptions{})...)
		if diags.HasError() {
			continue
		}

		tolerations = append(tolerations, &rafay.Toleration{
			Key:      toleration.Key.ValueString(),
			Operator: toleration.Operator.ValueString(),
			Value:    toleration.Value.ValueString(),
			Effect:   toleration.Effect.ValueString(),
		})
	}

	return tolerations
}

// convertSharing converts sharing configuration
func convertSharing(ctx context.Context, sharingObj types.Object, diags *diag.Diagnostics) *rafay.V1ClusterSharing {
	var sharingModel SharingModel
	diags.Append(sharingObj.As(ctx, &sharingModel, types.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}

	sharing := &rafay.V1ClusterSharing{
		Enabled: sharingModel.Enabled.ValueBool(),
	}

	// Convert projects map to array for API
	if !sharingModel.Projects.IsNull() && !sharingModel.Projects.IsUnknown() {
		projectsMap := make(map[string]types.Object)
		*diags = sharingModel.Projects.ElementsAs(ctx, &projectsMap, false)
		if diags.HasError() {
			return sharing
		}

		projects := make([]*rafay.SharingProject, 0, len(projectsMap))
		for _, projectObj := range projectsMap {
			var project ProjectModel
			diags.Append(projectObj.As(ctx, &project, types.ObjectAsOptions{})...)
			if diags.HasError() {
				continue
			}

			projects = append(projects, &rafay.SharingProject{
				Name: project.Name.ValueString(),
			})
		}
		sharing.Projects = projects
	}

	return sharing
}

// convertClusterConfig converts the full cluster config
func convertClusterConfig(ctx context.Context, configModel *ClusterConfigModel, diags *diag.Diagnostics) *rafay.EKSClusterConfig {
	config := &rafay.EKSClusterConfig{
		APIVersion: configModel.APIVersion.ValueString(),
		Kind:       configModel.Kind.ValueString(),
	}

	// Convert metadata
	config.Metadata = convertClusterConfigMetadata(ctx, configModel.Metadata, diags)

	// Convert VPC
	if !configModel.VPC.IsNull() && !configModel.VPC.IsUnknown() {
		config.VPC = convertVPC(ctx, configModel.VPC, diags)
	}

	// Convert node groups (map to array)
	if !configModel.NodeGroups.IsNull() && !configModel.NodeGroups.IsUnknown() {
		config.NodeGroups = convertNodeGroupsMapToArray(ctx, configModel.NodeGroups, diags)
	}

	// Convert managed node groups (map to array)
	if !configModel.ManagedNodeGroups.IsNull() && !configModel.ManagedNodeGroups.IsUnknown() {
		config.ManagedNodeGroups = convertManagedNodeGroupsMapToArray(ctx, configModel.ManagedNodeGroups, diags)
	}

	// Convert identity providers (map to array)
	if !configModel.IdentityProviders.IsNull() && !configModel.IdentityProviders.IsUnknown() {
		config.IdentityProviders = convertIdentityProvidersMapToArray(ctx, configModel.IdentityProviders, diags)
	}

	// Convert encryption config
	if !configModel.EncryptionConfig.IsNull() && !configModel.EncryptionConfig.IsUnknown() {
		config.SecretsEncryption = convertEncryptionConfig(ctx, configModel.EncryptionConfig, diags)
	}

	// Convert access entries (map to array)
	if !configModel.AccessEntries.IsNull() && !configModel.AccessEntries.IsUnknown() {
		config.AccessConfig = convertAccessConfig(ctx, configModel.AccessEntries, diags)
	}

	// Convert identity mappings
	if !configModel.IdentityMappings.IsNull() && !configModel.IdentityMappings.IsUnknown() {
		config.IdentityMappings = convertIdentityMappings(ctx, configModel.IdentityMappings, diags)
	}

	return config
}

// convertClusterConfigMetadata converts cluster config metadata
func convertClusterConfigMetadata(ctx context.Context, metadataObj types.Object, diags *diag.Diagnostics) *rafay.EKSClusterConfigMetadata {
	if metadataObj.IsNull() || metadataObj.IsUnknown() {
		return nil
	}

	var configMetadata ClusterConfigMetadataModel
	diags.Append(metadataObj.As(ctx, &configMetadata, types.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}

	metadata := &rafay.EKSClusterConfigMetadata{
		Name:    configMetadata.Name.ValueString(),
		Region:  configMetadata.Region.ValueString(),
		Version: configMetadata.Version.ValueString(),
	}

	// Extract tags map
	if !configMetadata.Tags.IsNull() && !configMetadata.Tags.IsUnknown() {
		tags := make(map[string]string)
		*diags = configMetadata.Tags.ElementsAs(ctx, &tags, false)
		if !diags.HasError() {
			metadata.Tags = tags
		}
	}

	return metadata
}

// convertVPC converts VPC configuration
func convertVPC(ctx context.Context, vpcObj types.Object, diags *diag.Diagnostics) *rafay.EKSClusterVPC {
	var vpcModel VPCModel
	diags.Append(vpcObj.As(ctx, &vpcModel, types.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}

	vpc := &rafay.EKSClusterVPC{
		Region: vpcModel.Region.ValueString(),
		CIDR:   vpcModel.CIDR.ValueString(),
	}

	// Convert subnets (map to array by AZ)
	if !vpcModel.Subnets.IsNull() && !vpcModel.Subnets.IsUnknown() {
		vpc.Subnets = convertSubnets(ctx, vpcModel.Subnets, diags)
	}

	// Note: Additional VPC fields (NAT, SecurityGroup, ClusterResourcesVpcConfig) would be converted here
	// For brevity, showing the pattern - you'd add similar conversions for each field

	return vpc
}

// convertSubnets converts subnets from map structure to array
func convertSubnets(ctx context.Context, subnetsObj types.Object, diags *diag.Diagnostics) *rafay.ClusterSubnets {
	var subnetsModel SubnetsModel
	diags.Append(subnetsObj.As(ctx, &subnetsModel, types.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}

	subnets := &rafay.ClusterSubnets{}

	// Convert public subnets map to array
	if !subnetsModel.Public.IsNull() && !subnetsModel.Public.IsUnknown() {
		publicMap := make(map[string]types.Object)
		*diags = subnetsModel.Public.ElementsAs(ctx, &publicMap, false)
		if !diags.HasError() {
			public := make(map[string]*rafay.SubnetSpec)
			for az, subnetObj := range publicMap {
				var subnet SubnetModel
				diags.Append(subnetObj.As(ctx, &subnet, types.ObjectAsOptions{})...)
				if !diags.HasError() {
					public[az] = &rafay.SubnetSpec{
						ID:   subnet.ID.ValueString(),
						CIDR: subnet.CIDR.ValueString(),
						AZ:   subnet.AZ.ValueString(),
					}
				}
			}
			subnets.Public = public
		}
	}

	// Convert private subnets map to array
	if !subnetsModel.Private.IsNull() && !subnetsModel.Private.IsUnknown() {
		privateMap := make(map[string]types.Object)
		*diags = subnetsModel.Private.ElementsAs(ctx, &privateMap, false)
		if !diags.HasError() {
			private := make(map[string]*rafay.SubnetSpec)
			for az, subnetObj := range privateMap {
				var subnet SubnetModel
				diags.Append(subnetObj.As(ctx, &subnet, types.ObjectAsOptions{})...)
				if !diags.HasError() {
					private[az] = &rafay.SubnetSpec{
						ID:   subnet.ID.ValueString(),
						CIDR: subnet.CIDR.ValueString(),
						AZ:   subnet.AZ.ValueString(),
					}
				}
			}
			subnets.Private = private
		}
	}

	return subnets
}

// convertNodeGroupsMapToArray converts node groups from map to array
func convertNodeGroupsMapToArray(ctx context.Context, nodeGroupsMap types.Map, diags *diag.Diagnostics) []*rafay.NodeGroup {
	nodeGroupsMapData := make(map[string]types.Object)
	*diags = nodeGroupsMap.ElementsAs(ctx, &nodeGroupsMapData, false)
	if diags.HasError() {
		return nil
	}

	nodeGroups := make([]*rafay.NodeGroup, 0, len(nodeGroupsMapData))
	for _, ngObj := range nodeGroupsMapData {
		var ng NodeGroupModel
		diags.Append(ngObj.As(ctx, &ng, types.ObjectAsOptions{})...)
		if diags.HasError() {
			continue
		}

		nodeGroup := &rafay.NodeGroup{
			Name:              ng.Name.ValueString(),
			AMI:               ng.AMI.ValueString(),
			InstanceType:      ng.InstanceType.ValueString(),
			DesiredCapacity:   int(ng.DesiredCapacity.ValueInt64()),
			MinSize:           int(ng.MinSize.ValueInt64()),
			MaxSize:           int(ng.MaxSize.ValueInt64()),
			VolumeSize:        int(ng.VolumeSize.ValueInt64()),
			VolumeType:        ng.VolumeType.ValueString(),
			PrivateNetworking: ng.PrivateNetworking.ValueBool(),
		}

		// Convert labels
		if !ng.Labels.IsNull() && !ng.Labels.IsUnknown() {
			labels := make(map[string]string)
			*diags = ng.Labels.ElementsAs(ctx, &labels, false)
			if !diags.HasError() {
				nodeGroup.Labels = labels
			}
		}

		// Convert tags
		if !ng.Tags.IsNull() && !ng.Tags.IsUnknown() {
			tags := make(map[string]string)
			*diags = ng.Tags.ElementsAs(ctx, &tags, false)
			if !diags.HasError() {
				nodeGroup.Tags = tags
			}
		}

		// Convert taints (map to array)
		if !ng.Taints.IsNull() && !ng.Taints.IsUnknown() {
			taintsMap := make(map[string]types.Object)
			*diags = ng.Taints.ElementsAs(ctx, &taintsMap, false)
			if !diags.HasError() {
				taints := make([]rafay.NodeGroupTaint, 0, len(taintsMap))
				for _, taintObj := range taintsMap {
					var taint TaintModel
					diags.Append(taintObj.As(ctx, &taint, types.ObjectAsOptions{})...)
					if !diags.HasError() {
						taints = append(taints, rafay.NodeGroupTaint{
							Key:    taint.Key.ValueString(),
							Value:  taint.Value.ValueString(),
							Effect: taint.Effect.ValueString(),
						})
					}
				}
				nodeGroup.Taints = taints
			}
		}

		// Convert availability zones (list)
		if !ng.AvailabilityZones.IsNull() && !ng.AvailabilityZones.IsUnknown() {
			var azs []string
			diags.Append(ng.AvailabilityZones.ElementsAs(ctx, &azs, false)...)
			if !diags.HasError() {
				nodeGroup.AvailabilityZones = azs
			}
		}

		nodeGroups = append(nodeGroups, nodeGroup)
	}

	return nodeGroups
}

// convertManagedNodeGroupsMapToArray converts managed node groups from map to array
func convertManagedNodeGroupsMapToArray(ctx context.Context, mngMap types.Map, diags *diag.Diagnostics) []*rafay.ManagedNodeGroup {
	mngMapData := make(map[string]types.Object)
	*diags = mngMap.ElementsAs(ctx, &mngMapData, false)
	if diags.HasError() {
		return nil
	}

	managedNodeGroups := make([]*rafay.ManagedNodeGroup, 0, len(mngMapData))
	for _, mngObj := range mngMapData {
		var mng ManagedNodeGroupModel
		diags.Append(mngObj.As(ctx, &mng, types.ObjectAsOptions{})...)
		if diags.HasError() {
			continue
		}

		managedNodeGroup := &rafay.ManagedNodeGroup{
			Name:        mng.Name.ValueString(),
			AMIType:     mng.AMIType.ValueString(),
			DesiredSize: int(mng.DesiredSize.ValueInt64()),
			MinSize:     int(mng.MinSize.ValueInt64()),
			MaxSize:     int(mng.MaxSize.ValueInt64()),
			DiskSize:    int(mng.DiskSize.ValueInt64()),
		}

		// Convert instance types (list)
		if !mng.InstanceTypes.IsNull() && !mng.InstanceTypes.IsUnknown() {
			var instanceTypes []string
			diags.Append(mng.InstanceTypes.ElementsAs(ctx, &instanceTypes, false)...)
			if !diags.HasError() {
				managedNodeGroup.InstanceTypes = instanceTypes
			}
		}

		// Convert labels
		if !mng.Labels.IsNull() && !mng.Labels.IsUnknown() {
			labels := make(map[string]string)
			*diags = mng.Labels.ElementsAs(ctx, &labels, false)
			if !diags.HasError() {
				managedNodeGroup.Labels = labels
			}
		}

		// Convert tags
		if !mng.Tags.IsNull() && !mng.Tags.IsUnknown() {
			tags := make(map[string]string)
			*diags = mng.Tags.ElementsAs(ctx, &tags, false)
			if !diags.HasError() {
				managedNodeGroup.Tags = tags
			}
		}

		// Convert taints (map to array)
		if !mng.Taints.IsNull() && !mng.Taints.IsUnknown() {
			taintsMap := make(map[string]types.Object)
			*diags = mng.Taints.ElementsAs(ctx, &taintsMap, false)
			if !diags.HasError() {
				taints := make([]*rafay.NodeGroupTaint, 0, len(taintsMap))
				for _, taintObj := range taintsMap {
					var taint TaintModel
					diags.Append(taintObj.As(ctx, &taint, types.ObjectAsOptions{})...)
					if !diags.HasError() {
						taints = append(taints, &rafay.NodeGroupTaint{
							Key:    taint.Key.ValueString(),
							Value:  taint.Value.ValueString(),
							Effect: taint.Effect.ValueString(),
						})
					}
				}
				managedNodeGroup.Taints = taints
			}
		}

		managedNodeGroups = append(managedNodeGroups, managedNodeGroup)
	}

	return managedNodeGroups
}

// convertIdentityProvidersMapToArray converts identity providers from map to array
func convertIdentityProvidersMapToArray(ctx context.Context, ipMap types.Map, diags *diag.Diagnostics) []*rafay.IdentityProvider {
	ipMapData := make(map[string]types.Object)
	*diags = ipMap.ElementsAs(ctx, &ipMapData, false)
	if diags.HasError() {
		return nil
	}

	identityProviders := make([]*rafay.IdentityProvider, 0, len(ipMapData))
	for _, ipObj := range ipMapData {
		var ip IdentityProviderModel
		diags.Append(ipObj.As(ctx, &ip, types.ObjectAsOptions{})...)
		if diags.HasError() {
			continue
		}

		identityProvider := &rafay.IdentityProvider{
			Type: ip.Type.ValueString(),
			// Name, IssuerURL, ClientID etc. would be added here based on the type
		}

		identityProviders = append(identityProviders, identityProvider)
	}

	return identityProviders
}

// convertEncryptionConfig converts encryption configuration
func convertEncryptionConfig(ctx context.Context, encConfigObj types.Object, diags *diag.Diagnostics) *rafay.SecretsEncryption {
	// Placeholder - implement based on actual schema
	return &rafay.SecretsEncryption{}
}

// convertAccessConfig converts access configuration
func convertAccessConfig(ctx context.Context, accessEntriesMap types.Map, diags *diag.Diagnostics) *rafay.EKSClusterAccess {
	accessEntriesMapData := make(map[string]types.Object)
	*diags = accessEntriesMap.ElementsAs(ctx, &accessEntriesMapData, false)
	if diags.HasError() {
		return nil
	}

	accessEntries := make([]*rafay.EKSAccessEntry, 0, len(accessEntriesMapData))
	for _, aeObj := range accessEntriesMapData {
		var ae AccessEntryModel
		diags.Append(aeObj.As(ctx, &ae, types.ObjectAsOptions{})...)
		if diags.HasError() {
			continue
		}

		accessEntry := &rafay.EKSAccessEntry{
			PrincipalARN: ae.PrincipalARN.ValueString(),
			Type:         ae.Type.ValueString(),
		}

		accessEntries = append(accessEntries, accessEntry)
	}

	return &rafay.EKSClusterAccess{
		AccessEntries: accessEntries,
	}
}

// convertIdentityMappings converts identity mappings
func convertIdentityMappings(ctx context.Context, imObj types.Object, diags *diag.Diagnostics) *rafay.EKSClusterIdentityMappings {
	// Placeholder - implement based on actual schema
	return &rafay.EKSClusterIdentityMappings{}
}

// Additional model types needed for conversion
type SubnetModel struct {
	ID   types.String `tfsdk:"id"`
	CIDR types.String `tfsdk:"cidr"`
	AZ   types.String `tfsdk:"az"`
}

type TaintModel struct {
	Key    types.String `tfsdk:"key"`
	Value  types.String `tfsdk:"value"`
	Effect types.String `tfsdk:"effect"`
}

type ManagedNodeGroupModel struct {
	Name          types.String `tfsdk:"name"`
	AMIType       types.String `tfsdk:"ami_type"`
	InstanceTypes types.List   `tfsdk:"instance_types"`
	DesiredSize   types.Int64  `tfsdk:"desired_size"`
	MinSize       types.Int64  `tfsdk:"min_size"`
	MaxSize       types.Int64  `tfsdk:"max_size"`
	DiskSize      types.Int64  `tfsdk:"disk_size"`
	Labels        types.Map    `tfsdk:"labels"`
	Tags          types.Map    `tfsdk:"tags"`
	Taints        types.Map    `tfsdk:"taints"`
}

type IdentityProviderModel struct {
	Type types.String `tfsdk:"type"`
	Name types.String `tfsdk:"name"`
}

type AccessEntryModel struct {
	PrincipalARN types.String `tfsdk:"principal_arn"`
	Type         types.String `tfsdk:"type"`
}

