package resource_eks_cluster

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func managedPlanNames(t *testing.T, ctx context.Context, list basetypes.ListValue) []string {
	t.Helper()

	items := make([]ManagedNodegroupsValue, 0, len(list.Elements()))
	if diags := list.ElementsAs(ctx, &items, false); diags.HasError() {
		t.Fatalf("reading managed node group list: %v", diags)
	}

	names := make([]string, 0, len(items))
	for _, item := range items {
		names = append(names, item.Name.ValueString())
	}
	return names
}

func runModifyManagedNodeGroupPlan(
	t *testing.T,
	ctx context.Context,
	stateList, configList basetypes.ListValue,
) ClusterConfigValue {
	t.Helper()

	stateCC := NewTestClusterConfigValueWithManaged(ctx, stateList)
	configCC := NewTestClusterConfigValueWithManaged(ctx, configList)

	stateCCList, diags := types.ListValue(ClusterConfigValue{}.Type(ctx), []attr.Value{stateCC})
	if diags.HasError() {
		t.Fatalf("building state cluster_config: %v", diags)
	}
	configCCList, diags := types.ListValue(ClusterConfigValue{}.Type(ctx), []attr.Value{configCC})
	if diags.HasError() {
		t.Fatalf("building config cluster_config: %v", diags)
	}

	state := EksClusterModel{ClusterConfig: stateCCList}
	config := EksClusterModel{ClusterConfig: configCCList}
	plan := EksClusterModel{ClusterConfig: stateCCList}

	if diags := ModifyEksClusterNodeGroupPlan(ctx, &config, &state, &plan); diags.HasError() {
		t.Fatalf("modify plan: %v", diags)
	}

	planCC := ClusterConfigFromModel(&plan)
	if planCC == nil {
		t.Fatal("expected cluster_config in plan")
	}
	return *planCC
}

func TestClusterConfigExpand_PrefersListOverComputedMap(t *testing.T) {
	ctx := context.Background()

	nodeGroups, diags := types.ListValue(NodeGroupsValue{}.Type(ctx), []attr.Value{
		NewTestNodeGroupsValue("ng-a", 1),
	})
	if diags.HasError() {
		t.Fatalf("building node groups: %v", diags)
	}

	configCC := NewTestClusterConfigValue(ctx, nodeGroups)

	planCC := configCC
	mapMirror := map[string]attr.Value{
		"ng-b": NewNodeGroupsMapValueNull(),
	}
	planCC.NodeGroupsMap, _ = types.MapValue(NodeGroupsMapValue{}.Type(ctx), mapMirror)

	opts := clusterConfigExpandOptsFromConfig(configCC)
	cfg, diags := planCC.Expand(ctx, opts)
	if diags.HasError() {
		t.Fatalf("expand cluster config: %v", diags)
	}

	if len(cfg.NodeGroups) != 1 || cfg.NodeGroups[0].Name != "ng-a" {
		t.Fatalf("expected list-based ng-a, got %+v", cfg.NodeGroups)
	}
}

func TestModifyEksClusterNodeGroupPlan_ManagedReorder(t *testing.T) {
	ctx := context.Background()

	stateList, diags := types.ListValue(ManagedNodegroupsValue{}.Type(ctx), []attr.Value{
		NewTestManagedNodegroupsValue("ng-1", 1),
		NewTestManagedNodegroupsValue("ng-2", 2),
	})
	if diags.HasError() {
		t.Fatalf("building state list: %v", diags)
	}

	configList, diags := types.ListValue(ManagedNodegroupsValue{}.Type(ctx), []attr.Value{
		NewTestManagedNodegroupsValue("ng-2", 2),
		NewTestManagedNodegroupsValue("ng-1", 1),
	})
	if diags.HasError() {
		t.Fatalf("building config list: %v", diags)
	}

	stateCC := NewTestClusterConfigValueWithManaged(ctx, stateList)
	planCC := runModifyManagedNodeGroupPlan(t, ctx, stateList, configList)
	if !planCC.ManagedNodegroups.Equal(stateCC.ManagedNodegroups) {
		t.Fatalf("expected plan to match state order on reorder-only, got names %v", managedPlanNames(t, ctx, planCC.ManagedNodegroups))
	}
}

func TestModifyEksClusterNodeGroupPlan_ManagedAddAtEnd(t *testing.T) {
	ctx := context.Background()

	stateList, diags := types.ListValue(ManagedNodegroupsValue{}.Type(ctx), []attr.Value{
		NewTestManagedNodegroupsValue("ng-1", 1),
		NewTestManagedNodegroupsValue("ng-2", 2),
	})
	if diags.HasError() {
		t.Fatalf("building state list: %v", diags)
	}

	configList, diags := types.ListValue(ManagedNodegroupsValue{}.Type(ctx), []attr.Value{
		NewTestManagedNodegroupsValue("ng-1", 1),
		NewTestManagedNodegroupsValue("ng-2", 2),
		NewTestManagedNodegroupsValue("ng-3", 3),
	})
	if diags.HasError() {
		t.Fatalf("building config list: %v", diags)
	}

	planCC := runModifyManagedNodeGroupPlan(t, ctx, stateList, configList)
	got := managedPlanNames(t, ctx, planCC.ManagedNodegroups)
	want := []string{"ng-1", "ng-2", "ng-3"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("expected names %v, got %v", want, got)
		}
	}
}

func TestModifyEksClusterNodeGroupPlan_ManagedAddOne(t *testing.T) {
	ctx := context.Background()

	stateList, diags := types.ListValue(ManagedNodegroupsValue{}.Type(ctx), []attr.Value{
		NewTestManagedNodegroupsValue("ng-1", 1),
		NewTestManagedNodegroupsValue("ng-2", 2),
	})
	if diags.HasError() {
		t.Fatalf("building state list: %v", diags)
	}

	configList, diags := types.ListValue(ManagedNodegroupsValue{}.Type(ctx), []attr.Value{
		NewTestManagedNodegroupsValue("ng-1", 1),
		NewTestManagedNodegroupsValue("ng-3", 3),
		NewTestManagedNodegroupsValue("ng-2", 2),
	})
	if diags.HasError() {
		t.Fatalf("building config list: %v", diags)
	}

	planCC := runModifyManagedNodeGroupPlan(t, ctx, stateList, configList)
	got := managedPlanNames(t, ctx, planCC.ManagedNodegroups)
	want := []string{"ng-1", "ng-3", "ng-2"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("expected names %v, got %v", want, got)
		}
	}
}

func TestModifyEksClusterNodeGroupPlan_ManagedRemoveOne(t *testing.T) {
	ctx := context.Background()

	stateList, diags := types.ListValue(ManagedNodegroupsValue{}.Type(ctx), []attr.Value{
		NewTestManagedNodegroupsValue("ng-1", 1),
		NewTestManagedNodegroupsValue("ng-2", 2),
	})
	if diags.HasError() {
		t.Fatalf("building state list: %v", diags)
	}

	configList, diags := types.ListValue(ManagedNodegroupsValue{}.Type(ctx), []attr.Value{
		NewTestManagedNodegroupsValue("ng-1", 1),
	})
	if diags.HasError() {
		t.Fatalf("building config list: %v", diags)
	}

	planCC := runModifyManagedNodeGroupPlan(t, ctx, stateList, configList)
	got := managedPlanNames(t, ctx, planCC.ManagedNodegroups)
	if len(got) != 1 || got[0] != "ng-1" {
		t.Fatalf("expected [ng-1], got %v", got)
	}
}

func TestModifyEksClusterNodeGroupPlan_ManagedChangeOne(t *testing.T) {
	ctx := context.Background()

	stateList, diags := types.ListValue(ManagedNodegroupsValue{}.Type(ctx), []attr.Value{
		NewTestManagedNodegroupsValue("ng-1", 1),
		NewTestManagedNodegroupsValue("ng-2", 2),
	})
	if diags.HasError() {
		t.Fatalf("building state list: %v", diags)
	}

	configList, diags := types.ListValue(ManagedNodegroupsValue{}.Type(ctx), []attr.Value{
		NewTestManagedNodegroupsValue("ng-1", 1),
		NewTestManagedNodegroupsValue("ng-2", 5),
	})
	if diags.HasError() {
		t.Fatalf("building config list: %v", diags)
	}

	planCC := runModifyManagedNodeGroupPlan(t, ctx, stateList, configList)
	items := make([]ManagedNodegroupsValue, 0, len(planCC.ManagedNodegroups.Elements()))
	if diags := planCC.ManagedNodegroups.ElementsAs(ctx, &items, false); diags.HasError() {
		t.Fatalf("reading plan list: %v", diags)
	}
	if items[1].DesiredCapacity.ValueInt64() != 5 {
		t.Fatalf("expected desired capacity 5, got %d", items[1].DesiredCapacity.ValueInt64())
	}
}

func TestModifyEksClusterNodeGroupPlan_ManagedReorderAndAdd(t *testing.T) {
	ctx := context.Background()

	stateList, diags := types.ListValue(ManagedNodegroupsValue{}.Type(ctx), []attr.Value{
		NewTestManagedNodegroupsValue("ng-1", 1),
		NewTestManagedNodegroupsValue("ng-2", 2),
	})
	if diags.HasError() {
		t.Fatalf("building state list: %v", diags)
	}

	configList, diags := types.ListValue(ManagedNodegroupsValue{}.Type(ctx), []attr.Value{
		NewTestManagedNodegroupsValue("ng-2", 2),
		NewTestManagedNodegroupsValue("ng-3", 3),
		NewTestManagedNodegroupsValue("ng-1", 1),
	})
	if diags.HasError() {
		t.Fatalf("building config list: %v", diags)
	}

	planCC := runModifyManagedNodeGroupPlan(t, ctx, stateList, configList)
	got := managedPlanNames(t, ctx, planCC.ManagedNodegroups)
	want := []string{"ng-2", "ng-3", "ng-1"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("expected names %v, got %v", want, got)
		}
	}
}
