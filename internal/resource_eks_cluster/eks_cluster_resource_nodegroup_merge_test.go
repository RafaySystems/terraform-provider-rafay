package resource_eks_cluster

import (
	"context"
	"testing"

	"github.com/RafaySystems/terraform-provider-rafay/rafay"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestAppendStateOnlyManagedNodeGroups_PreservesMissingFromAPI(t *testing.T) {
	ctx := context.Background()

	apiElements := []attr.Value{
		NewTestManagedNodegroupsValue("ng-1", 1),
	}
	apiNames := map[string]bool{"ng-1": true}

	stateList, diags := types.ListValue(ManagedNodegroupsValue{}.Type(ctx), []attr.Value{
		NewTestManagedNodegroupsValue("ng-1", 1),
		NewTestManagedNodegroupsValue("ng-2", 2),
	})
	if diags.HasError() {
		t.Fatalf("building state list: %v", diags)
	}

	stMngs := make([]ManagedNodegroupsValue, 0, len(stateList.Elements()))
	if diags := stateList.ElementsAs(ctx, &stMngs, false); diags.HasError() {
		t.Fatalf("reading state list: %v", diags)
	}

	merged := AppendStateOnlyManagedNodeGroupsListElements(apiElements, apiNames, stMngs)
	if len(merged) != 2 {
		t.Fatalf("expected 2 managed node groups, got %d", len(merged))
	}

	second, ok := merged[1].(ManagedNodegroupsValue)
	if !ok {
		t.Fatal("expected ManagedNodegroupsValue at index 1")
	}
	if second.Name.ValueString() != "ng-2" {
		t.Fatalf("expected preserved ng-2, got %s", second.Name.ValueString())
	}
}

func TestAppendStateOnlyNodeGroups_PreservesMissingFromAPI(t *testing.T) {
	ctx := context.Background()

	apiElements := []attr.Value{
		NewTestNodeGroupsValue("ng-a", 1),
	}
	apiNames := map[string]bool{"ng-a": true}

	stateList, diags := types.ListValue(NodeGroupsValue{}.Type(ctx), []attr.Value{
		NewTestNodeGroupsValue("ng-a", 1),
		NewTestNodeGroupsValue("ng-b", 2),
	})
	if diags.HasError() {
		t.Fatalf("building state list: %v", diags)
	}

	stNgs := make([]NodeGroupsValue, 0, len(stateList.Elements()))
	if diags := stateList.ElementsAs(ctx, &stNgs, false); diags.HasError() {
		t.Fatalf("reading state list: %v", diags)
	}

	merged := AppendStateOnlyNodeGroupsListElements(apiElements, apiNames, stNgs)
	if len(merged) != 2 {
		t.Fatalf("expected 2 node groups, got %d", len(merged))
	}

	second, ok := merged[1].(NodeGroupsValue)
	if !ok {
		t.Fatal("expected NodeGroupsValue at index 1")
	}
	if second.Name.ValueString() != "ng-b" {
		t.Fatalf("expected preserved ng-b, got %s", second.Name.ValueString())
	}
}

func TestBuildManagedNodeGroupsMapFromListElements_IncludesStateOnly(t *testing.T) {
	ctx := context.Background()

	list, diags := types.ListValue(ManagedNodegroupsValue{}.Type(ctx), []attr.Value{
		NewTestManagedNodegroupsValue("ng-1", 1),
		NewTestManagedNodegroupsValue("ng-2", 2),
	})
	if diags.HasError() {
		t.Fatalf("building list: %v", diags)
	}

	m, diags := buildManagedNodeGroupsMapFromListElements(ctx, list.Elements(), types.MapNull(ManagedNodegroupsMapValue{}.Type(ctx)))
	if diags.HasError() {
		t.Fatalf("building map mirror: %v", diags)
	}

	if len(m.Elements()) != 2 {
		t.Fatalf("expected map with 2 entries, got %d", len(m.Elements()))
	}
	for _, name := range []string{"ng-1", "ng-2"} {
		if _, ok := m.Elements()[name]; !ok {
			t.Fatalf("expected map key %q", name)
		}
	}
}

func TestOrderManagedNodeGroupsFromState_PreservesStateOrder(t *testing.T) {
	ctx := context.Background()

	stateList, diags := types.ListValue(ManagedNodegroupsValue{}.Type(ctx), []attr.Value{
		NewTestManagedNodegroupsValue("mng-a", 1),
		NewTestManagedNodegroupsValue("mng-b", 2),
	})
	if diags.HasError() {
		t.Fatalf("building state list: %v", diags)
	}

	apiGroups := []*rafay.ManagedNodeGroup{
		{Name: "mng-b"},
		{Name: "mng-a"},
	}

	ordered, diags := OrderManagedNodeGroupsFromState(ctx, apiGroups, stateList)
	if diags.HasError() {
		t.Fatalf("order managed node groups: %v", diags)
	}

	if len(ordered) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(ordered))
	}
	if ordered[0].Name != "mng-a" || ordered[1].Name != "mng-b" {
		t.Fatalf("expected [mng-a mng-b], got [%s %s]", ordered[0].Name, ordered[1].Name)
	}
}
