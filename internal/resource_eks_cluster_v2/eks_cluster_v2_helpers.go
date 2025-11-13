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

// convertModelToClusterSpec converts the Terraform model to the API cluster spec
func convertModelToClusterSpec(ctx context.Context, data *EKSClusterV2ResourceModel) (*rafay.EKSCluster, *rafay.EKSClusterConfig, diag.Diagnostics) {
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
	var metadataModel ClusterMetadataModel
	diags.Append(clusterModel.Metadata.As(ctx, &metadataModel, types.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil, nil, diags
	}

	eksCluster.Metadata = &rafay.EKSClusterMetadata{
		Name:    metadataModel.Name.ValueString(),
		Project: metadataModel.Project.ValueString(),
	}

	// Extract labels map
	if !metadataModel.Labels.IsNull() && !metadataModel.Labels.IsUnknown() {
		labels := make(map[string]string)
		diags = metadataModel.Labels.ElementsAs(ctx, &labels, false)
		if diags.HasError() {
			return nil, nil, diags
		}
		eksCluster.Metadata.Labels = labels
	}

	// Convert spec
	var specModel ClusterSpecModel
	diags.Append(clusterModel.Spec.As(ctx, &specModel, types.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil, nil, diags
	}

	eksCluster.Spec = &rafay.EKSSpec{
		Type:                specModel.Type.ValueString(),
		Blueprint:           specModel.Blueprint.ValueString(),
		BlueprintVersion:    specModel.BlueprintVersion.ValueString(),
		CloudProvider:       specModel.CloudProvider.ValueString(),
		CrossAccountRoleArn: specModel.CrossAccountRoleArn.ValueString(),
		CniProvider:         specModel.CniProvider.ValueString(),
	}

	// Convert CNI params
	if !specModel.CniParams.IsNull() && !specModel.CniParams.IsUnknown() {
		var cniParams CNIParamsModel
		diags.Append(specModel.CniParams.As(ctx, &cniParams, types.ObjectAsOptions{})...)
		if diags.HasError() {
			return nil, nil, diags
		}
		eksCluster.Spec.CniParams = &rafay.CustomCni{
			CustomCniCidr: cniParams.CustomCniCidr.ValueString(),
		}
	}

	// Convert proxy config (map)
	if !specModel.ProxyConfig.IsNull() && !specModel.ProxyConfig.IsUnknown() {
		proxyConfig := make(map[string]string)
		diags = specModel.ProxyConfig.ElementsAs(ctx, &proxyConfig, false)
		if diags.HasError() {
			return nil, nil, diags
		}
		eksCluster.Spec.ProxyConfig = &rafay.ProxyConfig{
			HttpProxy:  proxyConfig["http_proxy"],
			HttpsProxy: proxyConfig["https_proxy"],
			NoProxy:    proxyConfig["no_proxy"],
		}
	}

	// Convert system components placement
	if !specModel.SystemComponentsPlacement.IsNull() && !specModel.SystemComponentsPlacement.IsUnknown() {
		var scpModel SystemComponentsPlacementModel
		diags.Append(specModel.SystemComponentsPlacement.As(ctx, &scpModel, types.ObjectAsOptions{})...)
		if diags.HasError() {
			return nil, nil, diags
		}

		eksCluster.Spec.SystemComponentsPlacement = &rafay.SystemComponentsPlacement{}

		// Convert node selector (map)
		if !scpModel.NodeSelector.IsNull() && !scpModel.NodeSelector.IsUnknown() {
			nodeSelector := make(map[string]string)
			diags = scpModel.NodeSelector.ElementsAs(ctx, &nodeSelector, false)
			if diags.HasError() {
				return nil, nil, diags
			}
			eksCluster.Spec.SystemComponentsPlacement.NodeSelector = nodeSelector
		}

		// Convert tolerations (map to array for API)
		if !scpModel.Tolerations.IsNull() && !scpModel.Tolerations.IsUnknown() {
			tolerationsMap := make(map[string]types.Object)
			diags = scpModel.Tolerations.ElementsAs(ctx, &tolerationsMap, false)
			if diags.HasError() {
				return nil, nil, diags
			}

			tolerations := make([]*rafay.Toleration, 0, len(tolerationsMap))
			for _, tolerationObj := range tolerationsMap {
				var toleration TolerationModel
				diags.Append(tolerationObj.As(ctx, &toleration, types.ObjectAsOptions{})...)
				if diags.HasError() {
					return nil, nil, diags
				}

				tolerations = append(tolerations, &rafay.Toleration{
					Key:      toleration.Key.ValueString(),
					Operator: toleration.Operator.ValueString(),
					Value:    toleration.Value.ValueString(),
					Effect:   toleration.Effect.ValueString(),
				})
			}
			eksCluster.Spec.SystemComponentsPlacement.Tolerations = tolerations
		}

		// Convert daemonset tolerations (map to array for API)
		if !scpModel.DaemonsetTolerations.IsNull() && !scpModel.DaemonsetTolerations.IsUnknown() {
			dsTolerationsMap := make(map[string]types.Object)
			diags = scpModel.DaemonsetTolerations.ElementsAs(ctx, &dsTolerationsMap, false)
			if diags.HasError() {
				return nil, nil, diags
			}

			dsTolerations := make([]*rafay.Toleration, 0, len(dsTolerationsMap))
			for _, tolerationObj := range dsTolerationsMap {
				var toleration TolerationModel
				diags.Append(tolerationObj.As(ctx, &toleration, types.ObjectAsOptions{})...)
				if diags.HasError() {
					return nil, nil, diags
				}

				dsTolerations = append(dsTolerations, &rafay.Toleration{
					Key:      toleration.Key.ValueString(),
					Operator: toleration.Operator.ValueString(),
					Value:    toleration.Value.ValueString(),
					Effect:   toleration.Effect.ValueString(),
				})
			}
			eksCluster.Spec.SystemComponentsPlacement.DaemonsetTolerations = dsTolerations
		}

		// Convert daemonset node selector
		if !scpModel.DaemonsetNodeSelector.IsNull() && !scpModel.DaemonsetNodeSelector.IsUnknown() {
			dsNodeSelector := make(map[string]string)
			diags = scpModel.DaemonsetNodeSelector.ElementsAs(ctx, &dsNodeSelector, false)
			if diags.HasError() {
				return nil, nil, diags
			}
			eksCluster.Spec.SystemComponentsPlacement.DaemonsetNodeSelector = dsNodeSelector
		}
	}

	// Convert sharing
	if !specModel.Sharing.IsNull() && !specModel.Sharing.IsUnknown() {
		var sharingModel SharingModel
		diags.Append(specModel.Sharing.As(ctx, &sharingModel, types.ObjectAsOptions{})...)
		if diags.HasError() {
			return nil, nil, diags
		}

		eksCluster.Spec.Sharing = &rafay.V1ClusterSharing{
			Enabled: sharingModel.Enabled.ValueBool(),
		}

		// Convert projects map to array for API
		if !sharingModel.Projects.IsNull() && !sharingModel.Projects.IsUnknown() {
			projectsMap := make(map[string]types.Object)
			diags = sharingModel.Projects.ElementsAs(ctx, &projectsMap, false)
			if diags.HasError() {
				return nil, nil, diags
			}

			projects := make([]*rafay.SharingProject, 0, len(projectsMap))
			for _, projectObj := range projectsMap {
				var project ProjectModel
				diags.Append(projectObj.As(ctx, &project, types.ObjectAsOptions{})...)
				if diags.HasError() {
					return nil, nil, diags
				}

				projects = append(projects, &rafay.SharingProject{
					Name: project.Name.ValueString(),
				})
			}
			eksCluster.Spec.Sharing.Projects = projects
		}
	}

	// Build EKSClusterConfig (second YAML document)
	var clusterConfigModel ClusterConfigModel
	diags.Append(data.ClusterConfig.As(ctx, &clusterConfigModel, types.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil, nil, diags
	}

	eksClusterConfig := &rafay.EKSClusterConfig{
		APIVersion: clusterConfigModel.APIVersion.ValueString(),
		Kind:       clusterConfigModel.Kind.ValueString(),
	}

	// Convert cluster config metadata
	var configMetadataModel ClusterConfigMetadataModel
	diags.Append(clusterConfigModel.Metadata.As(ctx, &configMetadataModel, types.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil, nil, diags
	}

	eksClusterConfig.Metadata = &rafay.EKSClusterConfigMetadata{
		Name:    configMetadataModel.Name.ValueString(),
		Region:  configMetadataModel.Region.ValueString(),
		Version: configMetadataModel.Version.ValueString(),
	}

	// Extract tags map
	if !configMetadataModel.Tags.IsNull() && !configMetadataModel.Tags.IsUnknown() {
		tags := make(map[string]string)
		diags = configMetadataModel.Tags.ElementsAs(ctx, &tags, false)
		if diags.HasError() {
			return nil, nil, diags
		}
		eksClusterConfig.Metadata.Tags = tags
	}

	tflog.Info(ctx, "Successfully converted model to cluster spec")
	return eksCluster, eksClusterConfig, diags
}

// convertClusterSpecToModel converts API cluster spec to the Terraform model
func convertClusterSpecToModel(ctx context.Context, eksCluster *rafay.EKSCluster, eksClusterConfig *rafay.EKSClusterConfig) (*EKSClusterV2ResourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	model := &EKSClusterV2ResourceModel{}

	// Convert cluster metadata
	labels := types.MapNull(types.StringType)
	if eksCluster.Metadata != nil && len(eksCluster.Metadata.Labels) > 0 {
		labelsMap := make(map[string]attr.Value)
		for k, v := range eksCluster.Metadata.Labels {
			labelsMap[k] = types.StringValue(v)
		}
		var err error
		labels, err = types.MapValue(types.StringType, labelsMap)
		if err != nil {
			diags.AddError("Failed to convert labels", err.Error())
			return nil, diags
		}
	}

	metadataObj, metaDiags := types.ObjectValue(
		map[string]attr.Type{
			"name":    types.StringType,
			"project": types.StringType,
			"labels":  types.MapType{ElemType: types.StringType},
		},
		map[string]attr.Value{
			"name":    types.StringValue(eksCluster.Metadata.Name),
			"project": types.StringValue(eksCluster.Metadata.Project),
			"labels":  labels,
		},
	)
	diags.Append(metaDiags...)

	// TODO: Convert spec (system components, sharing, etc.)
	// For now, create a minimal spec object
	specObj := types.ObjectNull(map[string]attr.Type{
		"type":                       types.StringType,
		"blueprint":                  types.StringType,
		"blueprint_version":          types.StringType,
		"cloud_provider":             types.StringType,
		"cross_account_role_arn":     types.StringType,
		"cni_provider":               types.StringType,
		"cni_params":                 types.ObjectType{AttrTypes: map[string]attr.Type{}},
		"proxy_config":               types.MapType{ElemType: types.StringType},
		"system_components_placement": types.ObjectType{AttrTypes: map[string]attr.Type{}},
		"sharing":                    types.ObjectType{AttrTypes: map[string]attr.Type{}},
	})

	clusterObj, clusterDiags := types.ObjectValue(
		map[string]attr.Type{
			"kind":     types.StringType,
			"metadata": types.ObjectType{AttrTypes: map[string]attr.Type{}},
			"spec":     types.ObjectType{AttrTypes: map[string]attr.Type{}},
		},
		map[string]attr.Value{
			"kind":     types.StringValue(eksCluster.Kind),
			"metadata": metadataObj,
			"spec":     specObj,
		},
	)
	diags.Append(clusterDiags...)

	model.Cluster = clusterObj

	// TODO: Convert cluster config
	// For now, create a minimal cluster config object
	model.ClusterConfig = types.ObjectNull(map[string]attr.Type{
		"apiversion": types.StringType,
		"kind":       types.StringType,
		"metadata":   types.ObjectType{AttrTypes: map[string]attr.Type{}},
		// Add other fields...
	})

	tflog.Info(ctx, "Successfully converted cluster spec to model")
	return model, diags
}

// TolerationModel represents a toleration
type TolerationModel struct {
	Key      types.String `tfsdk:"key"`
	Operator types.String `tfsdk:"operator"`
	Value    types.String `tfsdk:"value"`
	Effect   types.String `tfsdk:"effect"`
}

// ProjectModel represents a sharing project
type ProjectModel struct {
	Name types.String `tfsdk:"name"`
}

// ClusterConfigMetadataModel represents cluster config metadata
type ClusterConfigMetadataModel struct {
	Name    types.String `tfsdk:"name"`
	Region  types.String `tfsdk:"region"`
	Version types.String `tfsdk:"version"`
	Tags    types.Map    `tfsdk:"tags"`
}
