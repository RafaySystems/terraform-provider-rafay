package resource_eks_cluster_test

import (
	"context"
	"testing"

	"github.com/RafaySystems/terraform-provider-rafay/internal/resource_eks_cluster"
	"github.com/RafaySystems/terraform-provider-rafay/rafay"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func testNodeGroup(name string, desired int64) resource_eks_cluster.NodeGroupsValue {
	return resource_eks_cluster.NewTestNodeGroupsValue(name, desired)
}

func testManagedNodeGroup(name string, desired int64) resource_eks_cluster.ManagedNodegroupsValue {
	return resource_eks_cluster.NewTestManagedNodegroupsValue(name, desired)
}

func mustNodeGroupList(t *testing.T, ctx context.Context, items ...resource_eks_cluster.NodeGroupsValue) basetypes.ListValue {
	t.Helper()

	vals := make([]attr.Value, len(items))
	for i, item := range items {
		vals[i] = item
	}

	list, diags := types.ListValue(resource_eks_cluster.NodeGroupsValue{}.Type(ctx), vals)
	if diags.HasError() {
		t.Fatalf("building node group list: %v", diags)
	}
	return list
}

func mustManagedNodeGroupList(t *testing.T, ctx context.Context, items ...resource_eks_cluster.ManagedNodegroupsValue) basetypes.ListValue {
	t.Helper()

	vals := make([]attr.Value, len(items))
	for i, item := range items {
		vals[i] = item
	}

	list, diags := types.ListValue(resource_eks_cluster.ManagedNodegroupsValue{}.Type(ctx), vals)
	if diags.HasError() {
		t.Fatalf("building managed node group list: %v", diags)
	}
	return list
}

func listNodeGroupNames(t *testing.T, ctx context.Context, list basetypes.ListValue) []string {
	t.Helper()

	items := make([]resource_eks_cluster.NodeGroupsValue, 0, len(list.Elements()))
	if diags := list.ElementsAs(ctx, &items, false); diags.HasError() {
		t.Fatalf("reading node group list: %v", diags)
	}

	names := make([]string, 0, len(items))
	for _, item := range items {
		names = append(names, item.Name.ValueString())
	}
	return names
}

func TestNormalizeNodeGroupsPlanList_ReorderOnly(t *testing.T) {
	ctx := context.Background()

	state := mustNodeGroupList(t, ctx,
		testNodeGroup("ng-a", 1),
		testNodeGroup("ng-b", 2),
	)
	config := mustNodeGroupList(t, ctx,
		testNodeGroup("ng-b", 2),
		testNodeGroup("ng-a", 1),
	)

	plan, diags := resource_eks_cluster.NormalizeNodeGroupsPlanList(ctx, config, state)
	if diags.HasError() {
		t.Fatalf("normalize plan list: %v", diags)
	}

	if !plan.Equal(state) {
		t.Fatalf("expected reorder-only plan to match state order, got names %v", listNodeGroupNames(t, ctx, plan))
	}
}

func TestNormalizeNodeGroupsPlanList_AddOne(t *testing.T) {
	ctx := context.Background()

	state := mustNodeGroupList(t, ctx,
		testNodeGroup("ng-a", 1),
		testNodeGroup("ng-b", 2),
	)
	config := mustNodeGroupList(t, ctx,
		testNodeGroup("ng-a", 1),
		testNodeGroup("ng-c", 3),
		testNodeGroup("ng-b", 2),
	)

	plan, diags := resource_eks_cluster.NormalizeNodeGroupsPlanList(ctx, config, state)
	if diags.HasError() {
		t.Fatalf("normalize plan list: %v", diags)
	}

	got := listNodeGroupNames(t, ctx, plan)
	want := []string{"ng-a", "ng-c", "ng-b"}
	if len(got) != len(want) {
		t.Fatalf("expected %d node groups, got %d (%v)", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("expected names %v, got %v", want, got)
		}
	}
}

func TestNormalizeNodeGroupsPlanList_RemoveOne(t *testing.T) {
	ctx := context.Background()

	state := mustNodeGroupList(t, ctx,
		testNodeGroup("ng-a", 1),
		testNodeGroup("ng-b", 2),
	)
	config := mustNodeGroupList(t, ctx,
		testNodeGroup("ng-a", 1),
	)

	plan, diags := resource_eks_cluster.NormalizeNodeGroupsPlanList(ctx, config, state)
	if diags.HasError() {
		t.Fatalf("normalize plan list: %v", diags)
	}

	got := listNodeGroupNames(t, ctx, plan)
	if len(got) != 1 || got[0] != "ng-a" {
		t.Fatalf("expected [ng-a], got %v", got)
	}
}

func TestNormalizeNodeGroupsPlanList_ChangeOne(t *testing.T) {
	ctx := context.Background()

	state := mustNodeGroupList(t, ctx,
		testNodeGroup("ng-a", 1),
		testNodeGroup("ng-b", 2),
	)
	config := mustNodeGroupList(t, ctx,
		testNodeGroup("ng-a", 1),
		testNodeGroup("ng-b", 5),
	)

	plan, diags := resource_eks_cluster.NormalizeNodeGroupsPlanList(ctx, config, state)
	if diags.HasError() {
		t.Fatalf("normalize plan list: %v", diags)
	}

	items := make([]resource_eks_cluster.NodeGroupsValue, 0, len(plan.Elements()))
	if diags := plan.ElementsAs(ctx, &items, false); diags.HasError() {
		t.Fatalf("reading plan list: %v", diags)
	}

	if items[1].DesiredCapacity.ValueInt64() != 5 {
		t.Fatalf("expected changed desired capacity 5, got %d", items[1].DesiredCapacity.ValueInt64())
	}
	if listNodeGroupNames(t, ctx, plan)[0] != "ng-a" || listNodeGroupNames(t, ctx, plan)[1] != "ng-b" {
		t.Fatalf("expected stable order [ng-a ng-b], got %v", listNodeGroupNames(t, ctx, plan))
	}
}

func TestNormalizeNodeGroupsPlanList_AddAtEnd(t *testing.T) {
	ctx := context.Background()

	state := mustNodeGroupList(t, ctx,
		testNodeGroup("ng-a", 1),
		testNodeGroup("ng-b", 2),
	)
	config := mustNodeGroupList(t, ctx,
		testNodeGroup("ng-a", 1),
		testNodeGroup("ng-b", 2),
		testNodeGroup("ng-c", 3),
	)

	plan, diags := resource_eks_cluster.NormalizeNodeGroupsPlanList(ctx, config, state)
	if diags.HasError() {
		t.Fatalf("normalize plan list: %v", diags)
	}

	got := listNodeGroupNames(t, ctx, plan)
	want := []string{"ng-a", "ng-b", "ng-c"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("expected names %v, got %v", want, got)
		}
	}
}

func TestNormalizeManagedNodeGroupsPlanList_AddAtEnd(t *testing.T) {
	ctx := context.Background()

	state := mustManagedNodeGroupList(t, ctx,
		testManagedNodeGroup("mng-a", 1),
		testManagedNodeGroup("mng-b", 2),
	)
	config := mustManagedNodeGroupList(t, ctx,
		testManagedNodeGroup("mng-a", 1),
		testManagedNodeGroup("mng-b", 2),
		testManagedNodeGroup("mng-c", 3),
	)

	plan, diags := resource_eks_cluster.NormalizeManagedNodeGroupsPlanList(ctx, config, state)
	if diags.HasError() {
		t.Fatalf("normalize managed plan list: %v", diags)
	}

	got := listManagedNodeGroupNames(t, ctx, plan)
	want := []string{"mng-a", "mng-b", "mng-c"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("expected names %v, got %v", want, got)
		}
	}
}

func TestNormalizeManagedNodeGroupsPlanList_ReorderOnly(t *testing.T) {
	ctx := context.Background()

	state := mustManagedNodeGroupList(t, ctx,
		testManagedNodeGroup("mng-a", 1),
		testManagedNodeGroup("mng-b", 2),
	)
	config := mustManagedNodeGroupList(t, ctx,
		testManagedNodeGroup("mng-b", 2),
		testManagedNodeGroup("mng-a", 1),
	)

	plan, diags := resource_eks_cluster.NormalizeManagedNodeGroupsPlanList(ctx, config, state)
	if diags.HasError() {
		t.Fatalf("normalize managed plan list: %v", diags)
	}

	if !plan.Equal(state) {
		t.Fatal("expected reorder-only managed plan to match state")
	}
}

func TestNormalizeManagedNodeGroupsPlanList_AddOne(t *testing.T) {
	ctx := context.Background()

	state := mustManagedNodeGroupList(t, ctx,
		testManagedNodeGroup("mng-a", 1),
		testManagedNodeGroup("mng-b", 2),
	)
	config := mustManagedNodeGroupList(t, ctx,
		testManagedNodeGroup("mng-a", 1),
		testManagedNodeGroup("mng-c", 3),
		testManagedNodeGroup("mng-b", 2),
	)

	plan, diags := resource_eks_cluster.NormalizeManagedNodeGroupsPlanList(ctx, config, state)
	if diags.HasError() {
		t.Fatalf("normalize managed plan list: %v", diags)
	}

	got := listManagedNodeGroupNames(t, ctx, plan)
	want := []string{"mng-a", "mng-c", "mng-b"}
	if len(got) != len(want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("expected names %v, got %v", want, got)
		}
	}
}

func TestNormalizeManagedNodeGroupsPlanList_RemoveOne(t *testing.T) {
	ctx := context.Background()

	state := mustManagedNodeGroupList(t, ctx,
		testManagedNodeGroup("mng-a", 1),
		testManagedNodeGroup("mng-b", 2),
	)
	config := mustManagedNodeGroupList(t, ctx,
		testManagedNodeGroup("mng-a", 1),
	)

	plan, diags := resource_eks_cluster.NormalizeManagedNodeGroupsPlanList(ctx, config, state)
	if diags.HasError() {
		t.Fatalf("normalize managed plan list: %v", diags)
	}

	got := listManagedNodeGroupNames(t, ctx, plan)
	if len(got) != 1 || got[0] != "mng-a" {
		t.Fatalf("expected [mng-a], got %v", got)
	}
}

func TestNormalizeManagedNodeGroupsPlanList_ChangeOne(t *testing.T) {
	ctx := context.Background()

	state := mustManagedNodeGroupList(t, ctx,
		testManagedNodeGroup("mng-a", 1),
		testManagedNodeGroup("mng-b", 2),
	)
	config := mustManagedNodeGroupList(t, ctx,
		testManagedNodeGroup("mng-a", 1),
		testManagedNodeGroup("mng-b", 5),
	)

	plan, diags := resource_eks_cluster.NormalizeManagedNodeGroupsPlanList(ctx, config, state)
	if diags.HasError() {
		t.Fatalf("normalize managed plan list: %v", diags)
	}

	items := make([]resource_eks_cluster.ManagedNodegroupsValue, 0, len(plan.Elements()))
	if diags := plan.ElementsAs(ctx, &items, false); diags.HasError() {
		t.Fatalf("reading managed plan list: %v", diags)
	}

	if items[1].DesiredCapacity.ValueInt64() != 5 {
		t.Fatalf("expected changed desired capacity 5, got %d", items[1].DesiredCapacity.ValueInt64())
	}
	if listManagedNodeGroupNames(t, ctx, plan)[0] != "mng-a" || listManagedNodeGroupNames(t, ctx, plan)[1] != "mng-b" {
		t.Fatalf("expected stable order [mng-a mng-b], got %v", listManagedNodeGroupNames(t, ctx, plan))
	}
}

func listManagedNodeGroupNames(t *testing.T, ctx context.Context, list basetypes.ListValue) []string {
	t.Helper()

	items := make([]resource_eks_cluster.ManagedNodegroupsValue, 0, len(list.Elements()))
	if diags := list.ElementsAs(ctx, &items, false); diags.HasError() {
		t.Fatalf("reading managed node group list: %v", diags)
	}

	names := make([]string, 0, len(items))
	for _, item := range items {
		names = append(names, item.Name.ValueString())
	}
	return names
}

func TestOrderNodeGroupsFromState_PreservesStateOrder(t *testing.T) {
	ctx := context.Background()

	state := mustNodeGroupList(t, ctx,
		testNodeGroup("ng-a", 1),
		testNodeGroup("ng-b", 2),
	)

	apiGroups := []*rafay.NodeGroup{
		{Name: "ng-b"},
		{Name: "ng-a"},
	}

	ordered, diags := resource_eks_cluster.OrderNodeGroupsFromState(ctx, apiGroups, state)
	if diags.HasError() {
		t.Fatalf("order node groups: %v", diags)
	}

	if len(ordered) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(ordered))
	}
	if ordered[0].Name != "ng-a" || ordered[1].Name != "ng-b" {
		t.Fatalf("expected [ng-a ng-b], got [%s %s]", ordered[0].Name, ordered[1].Name)
	}
}


func TestEksClusterResourceSchemaPatched_MarksMapFieldsComputed(t *testing.T) {
	ctx := context.Background()
	schemaDef := resource_eks_cluster.EksClusterResourceSchemaPatched(ctx)

	idAttr, ok := schemaDef.Attributes["id"].(schema.StringAttribute)
	if !ok {
		t.Fatal("id attribute missing or not StringAttribute")
	}
	if len(idAttr.PlanModifiers) == 0 {
		t.Fatal("expected id to use plan modifiers for stable computed value")
	}

	ccBlock, ok := schemaDef.Blocks["cluster_config"].(schema.ListNestedBlock)
	if !ok {
		t.Fatal("cluster_config block missing")
	}

	for _, name := range []string{"node_groups_map", "managed_nodegroups_map"} {
		attr, ok := ccBlock.NestedObject.Attributes[name]
		if !ok {
			t.Fatalf("%s attribute missing", name)
		}
		mapAttr, ok := attr.(schema.MapNestedAttribute)
		if !ok {
			t.Fatalf("%s is not MapNestedAttribute", name)
		}
		if !mapAttr.Computed || !mapAttr.Optional {
			t.Fatalf("%s expected Computed+Optional, got computed=%v optional=%v", name, mapAttr.Computed, mapAttr.Optional)
		}
	}
}
