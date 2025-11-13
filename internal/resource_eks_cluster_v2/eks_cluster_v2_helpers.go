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
// This function now delegates to the complete converters in eks_cluster_v2_converters_complete.go
func convertModelToClusterSpec(ctx context.Context, data *EKSClusterV2ResourceModel) (*rafay.EKSCluster, *rafay.EKSClusterConfig, diag.Diagnostics) {
	// For now, use the complete converter
	// TODO: Once fully tested, can inline this or keep separate for modularity
	return convertModelToClusterSpecComplete(ctx, data)
}

// convertClusterSpecToModel converts API cluster spec to the Terraform model
// This function now delegates to the complete reverse converters in eks_cluster_v2_reverse_converters.go
func convertClusterSpecToModel(ctx context.Context, eksCluster *rafay.EKSCluster, eksClusterConfig *rafay.EKSClusterConfig) (*EKSClusterV2ResourceModel, diag.Diagnostics) {
	// Use the complete reverse converter
	return convertClusterSpecToModelComplete(ctx, eksCluster, eksClusterConfig)
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
