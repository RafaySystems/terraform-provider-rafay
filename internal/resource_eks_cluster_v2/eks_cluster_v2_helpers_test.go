package resource_eks_cluster_v2

import (
	"context"
	"testing"

	"github.com/RafaySystems/terraform-provider-rafay/rafay"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConvertModelToClusterSpec_Basic tests basic cluster metadata conversion
func TestConvertModelToClusterSpec_Basic(t *testing.T) {
	ctx := context.Background()

	// Create test model
	labels, _ := types.MapValue(types.StringType, map[string]attr.Value{
		"environment": types.StringValue("test"),
		"team":        types.StringValue("platform"),
	})

	metadataObj, _ := types.ObjectValue(
		map[string]attr.Type{
			"name":    types.StringType,
			"project": types.StringType,
			"labels":  types.MapType{ElemType: types.StringType},
		},
		map[string]attr.Value{
			"name":    types.StringValue("test-cluster"),
			"project": types.StringValue("defaultproject"),
			"labels":  labels,
		},
	)

	specObj, _ := types.ObjectValue(
		map[string]attr.Type{
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
		},
		map[string]attr.Value{
			"type":                        types.StringValue("aws-eks"),
			"blueprint":                   types.StringValue("default"),
			"blueprint_version":           types.StringValue("latest"),
			"cloud_provider":              types.StringValue("aws-creds"),
			"cross_account_role_arn":      types.StringValue("arn:aws:iam::123456789012:role/CrossAccountRole"),
			"cni_provider":                types.StringValue("aws-cni"),
			"cni_params":                  types.ObjectNull(map[string]attr.Type{}),
			"proxy_config":                types.MapNull(types.StringType),
			"system_components_placement": types.ObjectNull(map[string]attr.Type{}),
			"sharing":                     types.ObjectNull(map[string]attr.Type{}),
		},
	)

	clusterObj, _ := types.ObjectValue(
		map[string]attr.Type{
			"kind":     types.StringType,
			"metadata": types.ObjectType{AttrTypes: map[string]attr.Type{}},
			"spec":     types.ObjectType{AttrTypes: map[string]attr.Type{}},
		},
		map[string]attr.Value{
			"kind":     types.StringValue("Cluster"),
			"metadata": metadataObj,
			"spec":     specObj,
		},
	)

	model := &EKSClusterV2ResourceModel{
		Cluster:       clusterObj,
		ClusterConfig: types.ObjectNull(map[string]attr.Type{}),
	}

	// Convert
	eksCluster, eksClusterConfig, diags := convertModelToClusterSpec(ctx, model)

	// Assert
	require.False(t, diags.HasError(), "Should not have errors")
	require.NotNil(t, eksCluster, "Should return cluster")
	require.NotNil(t, eksClusterConfig, "Should return cluster config")

	assert.Equal(t, "Cluster", eksCluster.Kind)
	assert.Equal(t, "test-cluster", eksCluster.Metadata.Name)
	assert.Equal(t, "defaultproject", eksCluster.Metadata.Project)
	assert.Equal(t, "test", eksCluster.Metadata.Labels["environment"])
	assert.Equal(t, "platform", eksCluster.Metadata.Labels["team"])

	assert.Equal(t, "aws-eks", eksCluster.Spec.Type)
	assert.Equal(t, "default", eksCluster.Spec.Blueprint)
	assert.Equal(t, "aws-creds", eksCluster.Spec.CloudProvider)
}

// TestConvertModelToClusterSpec_WithTolerations tests toleration conversion from map to array
func TestConvertModelToClusterSpec_WithTolerations(t *testing.T) {
	ctx := context.Background()

	// Create toleration objects
	toleration1Obj, _ := types.ObjectValue(
		map[string]attr.Type{
			"key":      types.StringType,
			"operator": types.StringType,
			"value":    types.StringType,
			"effect":   types.StringType,
		},
		map[string]attr.Value{
			"key":      types.StringValue("node-role"),
			"operator": types.StringValue("Equal"),
			"value":    types.StringValue("system"),
			"effect":   types.StringValue("NoSchedule"),
		},
	)

	toleration2Obj, _ := types.ObjectValue(
		map[string]attr.Type{
			"key":      types.StringType,
			"operator": types.StringType,
			"value":    types.StringType,
			"effect":   types.StringType,
		},
		map[string]attr.Value{
			"key":      types.StringValue("gpu"),
			"operator": types.StringValue("Exists"),
			"value":    types.StringValue(""),
			"effect":   types.StringValue("NoSchedule"),
		},
	)

	tolerationsMap, _ := types.MapValue(
		types.ObjectType{AttrTypes: map[string]attr.Type{
			"key":      types.StringType,
			"operator": types.StringType,
			"value":    types.StringType,
			"effect":   types.StringType,
		}},
		map[string]attr.Value{
			"node-role": toleration1Obj,
			"gpu":       toleration2Obj,
		},
	)

	scpObj, _ := types.ObjectValue(
		map[string]attr.Type{
			"node_selector":            types.MapType{ElemType: types.StringType},
			"tolerations":              types.MapType{ElemType: types.ObjectType{AttrTypes: map[string]attr.Type{}}},
			"daemonset_node_selector":  types.MapType{ElemType: types.StringType},
			"daemonset_tolerations":    types.MapType{ElemType: types.ObjectType{AttrTypes: map[string]attr.Type{}}},
		},
		map[string]attr.Value{
			"node_selector":           types.MapNull(types.StringType),
			"tolerations":             tolerationsMap,
			"daemonset_node_selector": types.MapNull(types.StringType),
			"daemonset_tolerations":   types.MapNull(types.ObjectType{AttrTypes: map[string]attr.Type{}}),
		},
	)

	// Create minimal model with system components placement
	metadataObj, _ := types.ObjectValue(
		map[string]attr.Type{
			"name":    types.StringType,
			"project": types.StringType,
			"labels":  types.MapType{ElemType: types.StringType},
		},
		map[string]attr.Value{
			"name":    types.StringValue("test-cluster"),
			"project": types.StringValue("defaultproject"),
			"labels":  types.MapNull(types.StringType),
		},
	)

	specObj, _ := types.ObjectValue(
		map[string]attr.Type{
			"type":                        types.StringType,
			"blueprint":                   types.StringType,
			"blueprint_version":           types.StringType,
			"cloud_provider":              types.StringType,
			"cross_account_role_arn":      types.StringType,
			"cni_provider":                types.StringType,
			"cni_params":                  types.ObjectType{AttrTypes: map[string]attr.Type{}},
			"proxy_config":                types.MapType{ElemType: types.StringType},
			"system_components_placement": types.ObjectType{AttrTypes: map[string]attr.Type{}},
			"sharing":                     types.ObjectType{AttrTypes: map[string]attr.Type{}},
		},
		map[string]attr.Value{
			"type":                        types.StringValue("aws-eks"),
			"blueprint":                   types.StringValue("default"),
			"blueprint_version":           types.StringValue(""),
			"cloud_provider":              types.StringValue("aws-creds"),
			"cross_account_role_arn":      types.StringValue(""),
			"cni_provider":                types.StringValue("aws-cni"),
			"cni_params":                  types.ObjectNull(map[string]attr.Type{}),
			"proxy_config":                types.MapNull(types.StringType),
			"system_components_placement": scpObj,
			"sharing":                     types.ObjectNull(map[string]attr.Type{}),
		},
	)

	clusterObj, _ := types.ObjectValue(
		map[string]attr.Type{
			"kind":     types.StringType,
			"metadata": types.ObjectType{AttrTypes: map[string]attr.Type{}},
			"spec":     types.ObjectType{AttrTypes: map[string]attr.Type{}},
		},
		map[string]attr.Value{
			"kind":     types.StringValue("Cluster"),
			"metadata": metadataObj,
			"spec":     specObj,
		},
	)

	model := &EKSClusterV2ResourceModel{
		Cluster:       clusterObj,
		ClusterConfig: types.ObjectNull(map[string]attr.Type{}),
	}

	// Convert
	eksCluster, _, diags := convertModelToClusterSpec(ctx, model)

	// Assert
	require.False(t, diags.HasError(), "Should not have errors")
	require.NotNil(t, eksCluster.Spec.SystemComponentsPlacement, "Should have system components placement")
	require.NotNil(t, eksCluster.Spec.SystemComponentsPlacement.Tolerations, "Should have tolerations")
	assert.Len(t, eksCluster.Spec.SystemComponentsPlacement.Tolerations, 2, "Should have 2 tolerations")

	// Check tolerations were converted from map to array
	var foundNodeRole, foundGPU bool
	for _, tol := range eksCluster.Spec.SystemComponentsPlacement.Tolerations {
		if tol.Key == "node-role" {
			foundNodeRole = true
			assert.Equal(t, "Equal", tol.Operator)
			assert.Equal(t, "system", tol.Value)
			assert.Equal(t, "NoSchedule", tol.Effect)
		}
		if tol.Key == "gpu" {
			foundGPU = true
			assert.Equal(t, "Exists", tol.Operator)
			assert.Equal(t, "NoSchedule", tol.Effect)
		}
	}

	assert.True(t, foundNodeRole, "Should have node-role toleration")
	assert.True(t, foundGPU, "Should have gpu toleration")
}

// TestConvertClusterSpecToModel_Basic tests basic cluster spec to model conversion
func TestConvertClusterSpecToModel_Basic(t *testing.T) {
	ctx := context.Background()

	// Create test cluster spec
	eksCluster := &rafay.EKSCluster{
		Kind: "Cluster",
		Metadata: &rafay.EKSClusterMetadata{
			Name:    "test-cluster",
			Project: "defaultproject",
			Labels: map[string]string{
				"environment": "test",
				"team":        "platform",
			},
		},
		Spec: &rafay.EKSSpec{
			Type:          "aws-eks",
			Blueprint:     "default",
			CloudProvider: "aws-creds",
			CniProvider:   "aws-cni",
		},
	}

	eksClusterConfig := &rafay.EKSClusterConfig{
		APIVersion: "rafay.io/v1alpha5",
		Kind:       "ClusterConfig",
		Metadata: &rafay.EKSClusterConfigMetadata{
			Name:    "test-cluster",
			Region:  "us-west-2",
			Version: "1.28",
			Tags: map[string]string{
				"Environment": "test",
			},
		},
	}

	// Convert
	model, diags := convertClusterSpecToModel(ctx, eksCluster, eksClusterConfig)

	// Assert
	require.False(t, diags.HasError(), "Should not have errors")
	require.NotNil(t, model, "Should return model")
	assert.False(t, model.Cluster.IsNull(), "Should have cluster object")
}

// TestConvertModelToClusterSpec_NullFields tests handling of null fields
func TestConvertModelToClusterSpec_NullFields(t *testing.T) {
	ctx := context.Background()

	// Create model with null optional fields
	metadataObj, _ := types.ObjectValue(
		map[string]attr.Type{
			"name":    types.StringType,
			"project": types.StringType,
			"labels":  types.MapType{ElemType: types.StringType},
		},
		map[string]attr.Value{
			"name":    types.StringValue("test-cluster"),
			"project": types.StringValue("defaultproject"),
			"labels":  types.MapNull(types.StringType), // Null labels
		},
	)

	specObj, _ := types.ObjectValue(
		map[string]attr.Type{
			"type":                        types.StringType,
			"blueprint":                   types.StringType,
			"blueprint_version":           types.StringType,
			"cloud_provider":              types.StringType,
			"cross_account_role_arn":      types.StringType,
			"cni_provider":                types.StringType,
			"cni_params":                  types.ObjectType{AttrTypes: map[string]attr.Type{}},
			"proxy_config":                types.MapType{ElemType: types.StringType},
			"system_components_placement": types.ObjectType{AttrTypes: map[string]attr.Type{}},
			"sharing":                     types.ObjectType{AttrTypes: map[string]attr.Type{}},
		},
		map[string]attr.Value{
			"type":                        types.StringValue("aws-eks"),
			"blueprint":                   types.StringValue("default"),
			"blueprint_version":           types.StringValue(""),
			"cloud_provider":              types.StringValue("aws-creds"),
			"cross_account_role_arn":      types.StringValue(""),
			"cni_provider":                types.StringValue("aws-cni"),
			"cni_params":                  types.ObjectNull(map[string]attr.Type{}), // Null CNI params
			"proxy_config":                types.MapNull(types.StringType),          // Null proxy
			"system_components_placement": types.ObjectNull(map[string]attr.Type{}), // Null SCP
			"sharing":                     types.ObjectNull(map[string]attr.Type{}), // Null sharing
		},
	)

	clusterObj, _ := types.ObjectValue(
		map[string]attr.Type{
			"kind":     types.StringType,
			"metadata": types.ObjectType{AttrTypes: map[string]attr.Type{}},
			"spec":     types.ObjectType{AttrTypes: map[string]attr.Type{}},
		},
		map[string]attr.Value{
			"kind":     types.StringValue("Cluster"),
			"metadata": metadataObj,
			"spec":     specObj,
		},
	)

	model := &EKSClusterV2ResourceModel{
		Cluster:       clusterObj,
		ClusterConfig: types.ObjectNull(map[string]attr.Type{}),
	}

	// Convert
	eksCluster, _, diags := convertModelToClusterSpec(ctx, model)

	// Assert - should handle null fields gracefully
	require.False(t, diags.HasError(), "Should not have errors")
	assert.Nil(t, eksCluster.Metadata.Labels, "Labels should be nil")
	assert.Nil(t, eksCluster.Spec.CniParams, "CNI params should be nil")
	assert.Nil(t, eksCluster.Spec.ProxyConfig, "Proxy config should be nil")
	assert.Nil(t, eksCluster.Spec.SystemComponentsPlacement, "System components placement should be nil")
	assert.Nil(t, eksCluster.Spec.Sharing, "Sharing should be nil")
}

