package resource_eks_cluster_v2

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/RafaySystems/rctl/pkg/cluster"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// convertModelToClusterSpec converts the Terraform model to Rafay cluster specification
func convertModelToClusterSpec(ctx context.Context, data *EKSClusterV2ResourceModel) (*cluster.Cluster, error) {
	tflog.Debug(ctx, "Converting model to cluster spec")

	clusterSpec := &cluster.Cluster{
		APIVersion: "rafay.io/v1alpha5",
		Kind:       "Cluster",
	}

	// Extract cluster metadata
	var clusterModel ClusterModel
	diags := data.Cluster.As(ctx, &clusterModel, types.ObjectAsOptions{})
	if diags.HasError() {
		return nil, fmt.Errorf("failed to extract cluster model: %v", diags)
	}

	// Extract metadata
	var metadata ClusterMetadataModel
	diags = clusterModel.Metadata.As(ctx, &metadata, types.ObjectAsOptions{})
	if diags.HasError() {
		return nil, fmt.Errorf("failed to extract metadata: %v", diags)
	}

	clusterSpec.Metadata = &cluster.Metadata{
		Name:    metadata.Name.ValueString(),
		Project: metadata.Project.ValueString(),
	}

	// Extract labels map
	if !metadata.Labels.IsNull() && !metadata.Labels.IsUnknown() {
		labels := make(map[string]string)
		diags = metadata.Labels.ElementsAs(ctx, &labels, false)
		if diags.HasError() {
			return nil, fmt.Errorf("failed to extract labels: %v", diags)
		}
		clusterSpec.Metadata.Labels = labels
	}

	// Extract spec
	var spec ClusterSpecModel
	diags = clusterModel.Spec.As(ctx, &spec, types.ObjectAsOptions{})
	if diags.HasError() {
		return nil, fmt.Errorf("failed to extract spec: %v", diags)
	}

	clusterSpec.Spec = &cluster.Spec{
		Type:                   spec.Type.ValueString(),
		Blueprint:              spec.Blueprint.ValueString(),
		BlueprintVersion:       spec.BlueprintVersion.ValueString(),
		CloudProvider:          spec.CloudProvider.ValueString(),
		CrossAccountRoleArn:    spec.CrossAccountRoleArn.ValueString(),
		CniProvider:            spec.CniProvider.ValueString(),
	}

	// Extract proxy config
	if !spec.ProxyConfig.IsNull() && !spec.ProxyConfig.IsUnknown() {
		proxyConfig := make(map[string]string)
		diags = spec.ProxyConfig.ElementsAs(ctx, &proxyConfig, false)
		if diags.HasError() {
			return nil, fmt.Errorf("failed to extract proxy config: %v", diags)
		}
		clusterSpec.Spec.ProxyConfig = proxyConfig
	}

	// TODO: Extract additional nested configurations
	// - CNI params
	// - System components placement
	// - Sharing

	return clusterSpec, nil
}

// convertModelToClusterConfig converts the Terraform model to EKS cluster configuration
func convertModelToClusterConfig(ctx context.Context, data *EKSClusterV2ResourceModel) (map[string]interface{}, error) {
	tflog.Debug(ctx, "Converting model to cluster config")

	var configModel ClusterConfigModel
	diags := data.ClusterConfig.As(ctx, &configModel, types.ObjectAsOptions{})
	if diags.HasError() {
		return nil, fmt.Errorf("failed to extract cluster config: %v", diags)
	}

	config := make(map[string]interface{})
	config["apiVersion"] = configModel.APIVersion.ValueString()
	config["kind"] = configModel.Kind.ValueString()

	// Extract metadata
	var metadata map[string]interface{}
	if !configModel.Metadata.IsNull() {
		metadata = make(map[string]interface{})
		// TODO: Extract metadata fields
	}
	config["metadata"] = metadata

	// Extract VPC configuration
	if !configModel.VPC.IsNull() && !configModel.VPC.IsUnknown() {
		vpc, err := extractVPCConfig(ctx, configModel.VPC)
		if err != nil {
			return nil, fmt.Errorf("failed to extract VPC config: %w", err)
		}
		config["vpc"] = vpc
	}

	// Extract node groups (map-based)
	if !configModel.NodeGroups.IsNull() && !configModel.NodeGroups.IsUnknown() {
		nodeGroups, err := extractNodeGroupsMap(ctx, configModel.NodeGroups)
		if err != nil {
			return nil, fmt.Errorf("failed to extract node groups: %w", err)
		}
		config["nodeGroups"] = nodeGroups
	}

	// Extract managed node groups (map-based)
	if !configModel.ManagedNodeGroups.IsNull() && !configModel.ManagedNodeGroups.IsUnknown() {
		managedNodeGroups, err := extractManagedNodeGroupsMap(ctx, configModel.ManagedNodeGroups)
		if err != nil {
			return nil, fmt.Errorf("failed to extract managed node groups: %w", err)
		}
		config["managedNodeGroups"] = managedNodeGroups
	}

	return config, nil
}

// extractVPCConfig extracts VPC configuration from the model
func extractVPCConfig(ctx context.Context, vpcObj types.Object) (map[string]interface{}, error) {
	var vpc VPCModel
	diags := vpcObj.As(ctx, &vpc, types.ObjectAsOptions{})
	if diags.HasError() {
		return nil, fmt.Errorf("failed to extract VPC model: %v", diags)
	}

	config := make(map[string]interface{})
	
	if !vpc.Region.IsNull() {
		config["region"] = vpc.Region.ValueString()
	}
	if !vpc.CIDR.IsNull() {
		config["cidr"] = vpc.CIDR.ValueString()
	}

	// Extract subnets (map-based)
	if !vpc.Subnets.IsNull() && !vpc.Subnets.IsUnknown() {
		subnets, err := extractSubnetsMap(ctx, vpc.Subnets)
		if err != nil {
			return nil, fmt.Errorf("failed to extract subnets: %w", err)
		}
		config["subnets"] = subnets
	}

	return config, nil
}

// extractSubnetsMap extracts subnet configuration from maps
func extractSubnetsMap(ctx context.Context, subnetsObj types.Object) (map[string]interface{}, error) {
	var subnets SubnetsModel
	diags := subnetsObj.As(ctx, &subnets, types.ObjectAsOptions{})
	if diags.HasError() {
		return nil, fmt.Errorf("failed to extract subnets model: %v", diags)
	}

	config := make(map[string]interface{})

	// Extract public subnets map
	if !subnets.Public.IsNull() && !subnets.Public.IsUnknown() {
		publicSubnets := make(map[string]interface{})
		// Iterate through map elements
		for key, value := range subnets.Public.Elements() {
			subnetObj, ok := value.(types.Object)
			if !ok {
				continue
			}
			// Extract subnet details
			subnetConfig := make(map[string]interface{})
			// TODO: Extract subnet fields (id, cidr, az)
			publicSubnets[key] = subnetConfig
		}
		config["public"] = publicSubnets
	}

	// Extract private subnets map
	if !subnets.Private.IsNull() && !subnets.Private.IsUnknown() {
		privateSubnets := make(map[string]interface{})
		// Iterate through map elements
		for key, value := range subnets.Private.Elements() {
			subnetObj, ok := value.(types.Object)
			if !ok {
				continue
			}
			// Extract subnet details
			subnetConfig := make(map[string]interface{})
			// TODO: Extract subnet fields (id, cidr, az)
			privateSubnets[key] = subnetConfig
		}
		config["private"] = privateSubnets
	}

	return config, nil
}

// extractNodeGroupsMap extracts self-managed node groups from a map
func extractNodeGroupsMap(ctx context.Context, nodeGroupsMap types.Map) ([]map[string]interface{}, error) {
	var nodeGroups []map[string]interface{}

	for name, value := range nodeGroupsMap.Elements() {
		nodeGroupObj, ok := value.(types.Object)
		if !ok {
			continue
		}

		var ng NodeGroupModel
		diags := nodeGroupObj.As(ctx, &ng, types.ObjectAsOptions{})
		if diags.HasError() {
			return nil, fmt.Errorf("failed to extract node group %s: %v", name, diags)
		}

		ngConfig := make(map[string]interface{})
		ngConfig["name"] = ng.Name.ValueString()
		
		if !ng.AMI.IsNull() {
			ngConfig["ami"] = ng.AMI.ValueString()
		}
		if !ng.InstanceType.IsNull() {
			ngConfig["instanceType"] = ng.InstanceType.ValueString()
		}
		if !ng.DesiredCapacity.IsNull() {
			ngConfig["desiredCapacity"] = ng.DesiredCapacity.ValueInt64()
		}
		if !ng.MinSize.IsNull() {
			ngConfig["minSize"] = ng.MinSize.ValueInt64()
		}
		if !ng.MaxSize.IsNull() {
			ngConfig["maxSize"] = ng.MaxSize.ValueInt64()
		}
		if !ng.VolumeSize.IsNull() {
			ngConfig["volumeSize"] = ng.VolumeSize.ValueInt64()
		}
		if !ng.VolumeType.IsNull() {
			ngConfig["volumeType"] = ng.VolumeType.ValueString()
		}
		if !ng.PrivateNetworking.IsNull() {
			ngConfig["privateNetworking"] = ng.PrivateNetworking.ValueBool()
		}

		// Extract labels map
		if !ng.Labels.IsNull() && !ng.Labels.IsUnknown() {
			labels := make(map[string]string)
			diags := ng.Labels.ElementsAs(ctx, &labels, false)
			if diags.HasError() {
				return nil, fmt.Errorf("failed to extract labels for node group %s: %v", name, diags)
			}
			ngConfig["labels"] = labels
		}

		// Extract tags map
		if !ng.Tags.IsNull() && !ng.Tags.IsUnknown() {
			tags := make(map[string]string)
			diags := ng.Tags.ElementsAs(ctx, &tags, false)
			if diags.HasError() {
				return nil, fmt.Errorf("failed to extract tags for node group %s: %v", name, diags)
			}
			ngConfig["tags"] = tags
		}

		// Extract taints map
		if !ng.Taints.IsNull() && !ng.Taints.IsUnknown() {
			taints, err := extractTaintsMap(ctx, ng.Taints)
			if err != nil {
				return nil, fmt.Errorf("failed to extract taints for node group %s: %w", name, err)
			}
			ngConfig["taints"] = taints
		}

		nodeGroups = append(nodeGroups, ngConfig)
	}

	return nodeGroups, nil
}

// extractManagedNodeGroupsMap extracts EKS managed node groups from a map
func extractManagedNodeGroupsMap(ctx context.Context, nodeGroupsMap types.Map) ([]map[string]interface{}, error) {
	var nodeGroups []map[string]interface{}

	for name, value := range nodeGroupsMap.Elements() {
		nodeGroupObj, ok := value.(types.Object)
		if !ok {
			continue
		}

		// Extract managed node group configuration
		ngConfig := make(map[string]interface{})
		ngConfig["name"] = name

		// TODO: Extract all managed node group fields

		nodeGroups = append(nodeGroups, ngConfig)
	}

	return nodeGroups, nil
}

// extractTaintsMap extracts taints from a map structure
func extractTaintsMap(ctx context.Context, taintsMap types.Map) ([]map[string]interface{}, error) {
	var taints []map[string]interface{}

	for key, value := range taintsMap.Elements() {
		taintObj, ok := value.(types.Object)
		if !ok {
			continue
		}

		taint := make(map[string]interface{})
		taint["key"] = key

		// Extract taint value and effect
		// TODO: Extract from taintObj

		taints = append(taints, taint)
	}

	return taints, nil
}

// convertClusterSpecToModel converts Rafay cluster spec back to Terraform model
func convertClusterSpecToModel(ctx context.Context, clusterSpec *cluster.Cluster, config map[string]interface{}) (*EKSClusterV2ResourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	
	tflog.Debug(ctx, "Converting cluster spec to model")

	model := &EKSClusterV2ResourceModel{}

	// Set ID
	if clusterSpec.Metadata != nil {
		model.ID = types.StringValue(fmt.Sprintf("%s/%s", clusterSpec.Metadata.Project, clusterSpec.Metadata.Name))
	}

	// TODO: Convert cluster spec to ClusterModel
	// TODO: Convert config to ClusterConfigModel
	// TODO: Set model.Cluster and model.ClusterConfig

	return model, diags
}

// waitForClusterReady waits for the cluster to reach ready state
func waitForClusterReady(ctx context.Context, clusterName, projectName string, timeout int) error {
	tflog.Debug(ctx, "Waiting for cluster to be ready", map[string]interface{}{
		"cluster": clusterName,
		"project": projectName,
		"timeout": timeout,
	})

	// TODO: Implement cluster status polling
	// TODO: Check cluster health
	// TODO: Handle timeout

	return nil
}

// waitForClusterDeleted waits for the cluster to be deleted
func waitForClusterDeleted(ctx context.Context, clusterName, projectName string, timeout int) error {
	tflog.Debug(ctx, "Waiting for cluster to be deleted", map[string]interface{}{
		"cluster": clusterName,
		"project": projectName,
		"timeout": timeout,
	})

	// TODO: Implement deletion status polling
	// TODO: Handle timeout

	return nil
}

// getClusterStatus retrieves the current status of a cluster
func getClusterStatus(ctx context.Context, clusterName, projectName string) (string, error) {
	tflog.Debug(ctx, "Getting cluster status", map[string]interface{}{
		"cluster": clusterName,
		"project": projectName,
	})

	// TODO: Call Rafay API to get cluster status

	return "READY", nil
}

// createMapAttribute creates a MapAttribute from a map[string]string
func createMapAttribute(ctx context.Context, data map[string]string) (types.Map, diag.Diagnostics) {
	if data == nil || len(data) == 0 {
		return types.MapNull(types.StringType), nil
	}

	elements := make(map[string]attr.Value)
	for k, v := range data {
		elements[k] = types.StringValue(v)
	}

	return types.MapValue(types.StringType, elements)
}

// createObjectAttribute creates an ObjectAttribute from a map
func createObjectAttribute(ctx context.Context, attrTypes map[string]attr.Type, data map[string]interface{}) (types.Object, diag.Diagnostics) {
	if data == nil || len(data) == 0 {
		return types.ObjectNull(attrTypes), nil
	}

	elements := make(map[string]attr.Value)
	
	// Convert data map to attr.Value map based on types
	for key, attrType := range attrTypes {
		if value, ok := data[key]; ok {
			// Convert based on type
			switch attrType {
			case types.StringType:
				if strVal, ok := value.(string); ok {
					elements[key] = types.StringValue(strVal)
				}
			case types.Int64Type:
				if intVal, ok := value.(int64); ok {
					elements[key] = types.Int64Value(intVal)
				} else if intVal, ok := value.(int); ok {
					elements[key] = types.Int64Value(int64(intVal))
				}
			case types.BoolType:
				if boolVal, ok := value.(bool); ok {
					elements[key] = types.BoolValue(boolVal)
				}
			}
		}
	}

	return types.ObjectValue(attrTypes, elements)
}

// marshalJSON marshals data to JSON string for API calls
func marshalJSON(data interface{}) (string, error) {
	bytes, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal data: %w", err)
	}
	return string(bytes), nil
}

// unmarshalJSON unmarshals JSON string to data structure
func unmarshalJSON(jsonStr string, target interface{}) error {
	err := json.Unmarshal([]byte(jsonStr), target)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	return nil
}

