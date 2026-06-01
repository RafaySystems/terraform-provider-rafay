package resource_eks_cluster

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func modifyClusterConfigNodeGroupPlan(ctx context.Context, configCC, stateCC, planCC *ClusterConfigValue) diag.Diagnostics {
	var diags diag.Diagnostics

	if configCC == nil || planCC == nil {
		return diags
	}

	useNGMap := clusterConfigUsesNodeGroupsMap(*configCC)
	useMNGMap := clusterConfigUsesManagedNodeGroupsMap(*configCC)

	if useNGMap {
		planCC.NodeGroups = types.ListNull(NodeGroupsValue{}.Type(ctx))
	} else if clusterConfigUsesNodeGroupsList(*configCC) {
		stateList := types.ListNull(NodeGroupsValue{}.Type(ctx))
		if stateCC != nil && !stateCC.NodeGroups.IsNull() {
			stateList = stateCC.NodeGroups
		}

		normalized, d := normalizeNodeGroupsPlanList(ctx, configCC.NodeGroups, stateList)
		diags = append(diags, d...)
		planCC.NodeGroups = normalized

		if stateCC != nil && !stateCC.NodeGroupsMap.IsNull() {
			planCC.NodeGroupsMap = stateCC.NodeGroupsMap
		} else {
			planCC.NodeGroupsMap = types.MapNull(NodeGroupsMapValue{}.Type(ctx))
		}
	} else {
		planCC.NodeGroups = types.ListNull(NodeGroupsValue{}.Type(ctx))
		planCC.NodeGroupsMap = types.MapNull(NodeGroupsMapValue{}.Type(ctx))
	}

	if useMNGMap {
		planCC.ManagedNodegroups = types.ListNull(ManagedNodegroupsValue{}.Type(ctx))
	} else if clusterConfigUsesManagedNodeGroupsList(*configCC) {
		stateList := types.ListNull(ManagedNodegroupsValue{}.Type(ctx))
		if stateCC != nil && !stateCC.ManagedNodegroups.IsNull() {
			stateList = stateCC.ManagedNodegroups
		}

		normalized, d := normalizeManagedNodeGroupsPlanList(ctx, configCC.ManagedNodegroups, stateList)
		diags = append(diags, d...)
		planCC.ManagedNodegroups = normalized

		if stateCC != nil && !stateCC.ManagedNodegroupsMap.IsNull() {
			planCC.ManagedNodegroupsMap = stateCC.ManagedNodegroupsMap
		} else {
			planCC.ManagedNodegroupsMap = types.MapNull(ManagedNodegroupsMapValue{}.Type(ctx))
		}
	} else {
		planCC.ManagedNodegroups = types.ListNull(ManagedNodegroupsValue{}.Type(ctx))
		planCC.ManagedNodegroupsMap = types.MapNull(ManagedNodegroupsMapValue{}.Type(ctx))
	}

	return diags
}

func ModifyEksClusterNodeGroupPlan(ctx context.Context, config, state, plan *EksClusterModel) diag.Diagnostics {
	var diags diag.Diagnostics

	configCC := ClusterConfigFromModel(config)
	stateCC := ClusterConfigFromModel(state)
	planCC := ClusterConfigFromModel(plan)
	if planCC == nil {
		return diags
	}

	diags = append(diags, modifyClusterConfigNodeGroupPlan(ctx, configCC, stateCC, planCC)...)

	planCCList := make([]ClusterConfigValue, 0, len(plan.ClusterConfig.Elements()))
	if d := plan.ClusterConfig.ElementsAs(ctx, &planCCList, false); d.HasError() {
		return append(diags, d...)
	}
	if len(planCCList) == 0 {
		planCCList = []ClusterConfigValue{*planCC}
	} else {
		planCCList[0] = *planCC
	}

	updatedCC, d := types.ListValue(ClusterConfigValue{}.Type(ctx), []attr.Value{planCCList[0]})
	diags = append(diags, d...)
	if diags.HasError() {
		return diags
	}
	plan.ClusterConfig = updatedCC

	return diags
}

func ClusterConfigFromModel(model *EksClusterModel) *ClusterConfigValue {
	if model == nil || model.ClusterConfig.IsNull() || model.ClusterConfig.IsUnknown() {
		return nil
	}
	ccList := make([]ClusterConfigValue, 0, len(model.ClusterConfig.Elements()))
	_ = model.ClusterConfig.ElementsAs(context.Background(), &ccList, false)
	if len(ccList) == 0 {
		return nil
	}
	return &ccList[0]
}
