package planmodifiers

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Helper function to create a nodegroup object value
func createNodeGroupObject(t *testing.T, name string, instanceType string, desiredCapacity int64) types.Object {
	t.Helper()

	attrTypes := map[string]attr.Type{
		"name":             types.StringType,
		"instance_type":    types.StringType,
		"desired_capacity": types.Int64Type,
		"min_size":         types.Int64Type,
		"max_size":         types.Int64Type,
	}

	attrValues := map[string]attr.Value{
		"name":             types.StringValue(name),
		"instance_type":    types.StringValue(instanceType),
		"desired_capacity": types.Int64Value(desiredCapacity),
		"min_size":         types.Int64Value(1),
		"max_size":         types.Int64Value(10),
	}

	obj, diags := types.ObjectValue(attrTypes, attrValues)
	if diags.HasError() {
		t.Fatalf("Failed to create nodegroup object: %v", diags)
	}
	return obj
}

// Helper function to create a list of nodegroups
func createNodeGroupList(t *testing.T, nodegroups ...types.Object) types.List {
	t.Helper()

	if len(nodegroups) == 0 {
		return types.ListNull(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"name":             types.StringType,
				"instance_type":    types.StringType,
				"desired_capacity": types.Int64Type,
				"min_size":         types.Int64Type,
				"max_size":         types.Int64Type,
			},
		})
	}

	elements := make([]attr.Value, len(nodegroups))
	for i, ng := range nodegroups {
		elements[i] = ng
	}

	listValue, diags := types.ListValue(nodegroups[0].Type(context.Background()), elements)
	if diags.HasError() {
		t.Fatalf("Failed to create nodegroup list: %v", diags)
	}
	return listValue
}

func TestNodeGroupSortModifier_Description(t *testing.T) {
	m := NodeGroupSortModifier{}
	ctx := context.Background()

	desc := m.Description(ctx)
	if desc == "" {
		t.Error("Description should not be empty")
	}

	mdDesc := m.MarkdownDescription(ctx)
	if mdDesc == "" {
		t.Error("MarkdownDescription should not be empty")
	}
}

func TestNodeGroupSortModifier_ReorderOnly(t *testing.T) {
	ctx := context.Background()
	m := NodeGroupSortModifier{}

	// Create nodegroups
	ng1 := createNodeGroupObject(t, "ng-1", "t3.large", 2)
	ng2 := createNodeGroupObject(t, "ng-2", "t3.large", 2)
	ng3 := createNodeGroupObject(t, "ng-3", "t3.large", 2)

	// State has [ng-1, ng-2, ng-3] (sorted)
	stateList := createNodeGroupList(t, ng1, ng2, ng3)

	// Config has [ng-3, ng-1, ng-2] (different order)
	configList := createNodeGroupList(t, ng3, ng1, ng2)

	req := planmodifier.ListRequest{
		ConfigValue: configList,
		StateValue:  stateList,
		PlanValue:   configList,
	}

	resp := &planmodifier.ListResponse{
		PlanValue: configList,
	}

	m.PlanModifyList(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("Unexpected error: %v", resp.Diagnostics)
	}

	// The plan should preserve state order since it's just a reorder
	// (or at least be sorted consistently)
	resultElements := resp.PlanValue.Elements()
	if len(resultElements) != 3 {
		t.Errorf("Expected 3 elements, got %d", len(resultElements))
	}
}

func TestNodeGroupSortModifier_AddNodeGroup(t *testing.T) {
	ctx := context.Background()
	m := NodeGroupSortModifier{}

	// Create nodegroups
	ng0 := createNodeGroupObject(t, "ng-0", "t3.large", 2)
	ng1 := createNodeGroupObject(t, "ng-1", "t3.large", 2)
	ng2 := createNodeGroupObject(t, "ng-2", "t3.large", 2)
	ng3 := createNodeGroupObject(t, "ng-3", "t3.large", 2)

	// State has [ng-1, ng-2, ng-3]
	stateList := createNodeGroupList(t, ng1, ng2, ng3)

	// Config has [ng-0, ng-1, ng-2, ng-3] (ng-0 added at start)
	configList := createNodeGroupList(t, ng0, ng1, ng2, ng3)

	req := planmodifier.ListRequest{
		ConfigValue: configList,
		StateValue:  stateList,
		PlanValue:   configList,
	}

	resp := &planmodifier.ListResponse{
		PlanValue: configList,
	}

	m.PlanModifyList(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("Unexpected error: %v", resp.Diagnostics)
	}

	// The plan should have 4 elements (the addition)
	resultElements := resp.PlanValue.Elements()
	if len(resultElements) != 4 {
		t.Errorf("Expected 4 elements, got %d", len(resultElements))
	}
}

func TestNodeGroupSortModifier_DeleteNodeGroup(t *testing.T) {
	ctx := context.Background()
	m := NodeGroupSortModifier{}

	// Create nodegroups
	ng1 := createNodeGroupObject(t, "ng-1", "t3.large", 2)
	ng2 := createNodeGroupObject(t, "ng-2", "t3.large", 2)
	ng3 := createNodeGroupObject(t, "ng-3", "t3.large", 2)

	// State has [ng-1, ng-2, ng-3]
	stateList := createNodeGroupList(t, ng1, ng2, ng3)

	// Config has [ng-2, ng-3] (ng-1 removed)
	configList := createNodeGroupList(t, ng2, ng3)

	req := planmodifier.ListRequest{
		ConfigValue: configList,
		StateValue:  stateList,
		PlanValue:   configList,
	}

	resp := &planmodifier.ListResponse{
		PlanValue: configList,
	}

	m.PlanModifyList(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("Unexpected error: %v", resp.Diagnostics)
	}

	// The plan should have 2 elements (the deletion)
	resultElements := resp.PlanValue.Elements()
	if len(resultElements) != 2 {
		t.Errorf("Expected 2 elements, got %d", len(resultElements))
	}
}

func TestNodeGroupSortModifier_FieldChange(t *testing.T) {
	ctx := context.Background()
	m := NodeGroupSortModifier{}

	// Create nodegroups
	ng1 := createNodeGroupObject(t, "ng-1", "t3.large", 2)
	ng2 := createNodeGroupObject(t, "ng-2", "t3.large", 2)
	ng3 := createNodeGroupObject(t, "ng-3", "t3.large", 2)

	// State has [ng-1, ng-2, ng-3] with desired_capacity=2
	stateList := createNodeGroupList(t, ng1, ng2, ng3)

	// Config has ng-2 with different desired_capacity
	ng2Changed := createNodeGroupObject(t, "ng-2", "t3.large", 5) // Changed from 2 to 5
	configList := createNodeGroupList(t, ng1, ng2Changed, ng3)

	req := planmodifier.ListRequest{
		ConfigValue: configList,
		StateValue:  stateList,
		PlanValue:   configList,
	}

	resp := &planmodifier.ListResponse{
		PlanValue: configList,
	}

	m.PlanModifyList(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("Unexpected error: %v", resp.Diagnostics)
	}

	// The plan should reflect the change
	resultElements := resp.PlanValue.Elements()
	if len(resultElements) != 3 {
		t.Errorf("Expected 3 elements, got %d", len(resultElements))
	}
}

func TestNodeGroupSortModifier_NullConfig(t *testing.T) {
	ctx := context.Background()
	m := NodeGroupSortModifier{}

	// Create nodegroups for state
	ng1 := createNodeGroupObject(t, "ng-1", "t3.large", 2)
	stateList := createNodeGroupList(t, ng1)

	// Config is null
	configList := types.ListNull(ng1.Type(ctx))

	req := planmodifier.ListRequest{
		ConfigValue: configList,
		StateValue:  stateList,
		PlanValue:   configList,
	}

	resp := &planmodifier.ListResponse{
		PlanValue: configList,
	}

	m.PlanModifyList(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("Unexpected error: %v", resp.Diagnostics)
	}

	// Plan should remain null
	if !resp.PlanValue.IsNull() {
		t.Error("Plan should be null when config is null")
	}
}

func TestNodeGroupSortModifier_NewResource(t *testing.T) {
	ctx := context.Background()
	m := NodeGroupSortModifier{}

	// Create nodegroups
	ng1 := createNodeGroupObject(t, "ng-1", "t3.large", 2)
	ng2 := createNodeGroupObject(t, "ng-2", "t3.large", 2)

	// State is null (new resource)
	stateList := types.ListNull(ng1.Type(ctx))

	// Config has nodegroups
	configList := createNodeGroupList(t, ng2, ng1) // Unsorted

	req := planmodifier.ListRequest{
		ConfigValue: configList,
		StateValue:  stateList,
		PlanValue:   configList,
	}

	resp := &planmodifier.ListResponse{
		PlanValue: configList,
	}

	m.PlanModifyList(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("Unexpected error: %v", resp.Diagnostics)
	}

	// Plan should have 2 elements
	resultElements := resp.PlanValue.Elements()
	if len(resultElements) != 2 {
		t.Errorf("Expected 2 elements, got %d", len(resultElements))
	}
}

func TestBuildNodeGroupMapFromList(t *testing.T) {
	ctx := context.Background()

	ng1 := createNodeGroupObject(t, "ng-1", "t3.large", 2)
	ng2 := createNodeGroupObject(t, "ng-2", "t3.large", 2)

	elements := []attr.Value{ng1, ng2}

	result := BuildNodeGroupMapFromList(ctx, elements)

	if len(result) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(result))
	}

	if _, exists := result["ng-1"]; !exists {
		t.Error("Expected ng-1 to exist in map")
	}

	if _, exists := result["ng-2"]; !exists {
		t.Error("Expected ng-2 to exist in map")
	}
}

func TestIsReorderOnly(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		config   map[string]attr.Value
		state    map[string]attr.Value
		expected bool
	}{
		{
			name:     "Same names - reorder only",
			config:   map[string]attr.Value{"ng-1": nil, "ng-2": nil, "ng-3": nil},
			state:    map[string]attr.Value{"ng-1": nil, "ng-2": nil, "ng-3": nil},
			expected: true,
		},
		{
			name:     "Different count - addition",
			config:   map[string]attr.Value{"ng-0": nil, "ng-1": nil, "ng-2": nil, "ng-3": nil},
			state:    map[string]attr.Value{"ng-1": nil, "ng-2": nil, "ng-3": nil},
			expected: false,
		},
		{
			name:     "Different names - replacement",
			config:   map[string]attr.Value{"ng-1": nil, "ng-2": nil, "ng-4": nil},
			state:    map[string]attr.Value{"ng-1": nil, "ng-2": nil, "ng-3": nil},
			expected: false,
		},
		{
			name:     "Empty both",
			config:   map[string]attr.Value{},
			state:    map[string]attr.Value{},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsReorderOnly(ctx, tt.config, tt.state)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestSortNodeGroupList(t *testing.T) {
	ctx := context.Background()

	ng1 := createNodeGroupObject(t, "ng-1", "t3.large", 2)
	ng2 := createNodeGroupObject(t, "ng-2", "t3.large", 2)
	ng3 := createNodeGroupObject(t, "ng-3", "t3.large", 2)

	// Unsorted list
	elements := []attr.Value{ng3, ng1, ng2}

	sorted := SortNodeGroupList(ctx, elements)

	if len(sorted) != 3 {
		t.Fatalf("Expected 3 elements, got %d", len(sorted))
	}

	// Verify order
	expectedOrder := []string{"ng-1", "ng-2", "ng-3"}
	for i, elem := range sorted {
		obj := elem.(types.Object)
		nameAttr := obj.Attributes()["name"].(types.String)
		if nameAttr.ValueString() != expectedOrder[i] {
			t.Errorf("Position %d: expected %s, got %s", i, expectedOrder[i], nameAttr.ValueString())
		}
	}
}

// Benchmark tests
func BenchmarkSortNodeGroupList(b *testing.B) {
	ctx := context.Background()

	// Create 100 nodegroups
	elements := make([]attr.Value, 100)
	for i := 0; i < 100; i++ {
		attrTypes := map[string]attr.Type{
			"name":             types.StringType,
			"instance_type":    types.StringType,
			"desired_capacity": types.Int64Type,
			"min_size":         types.Int64Type,
			"max_size":         types.Int64Type,
		}
		attrValues := map[string]attr.Value{
			"name":             types.StringValue(string(rune('z'-i%26)) + "-ng"),
			"instance_type":    types.StringValue("t3.large"),
			"desired_capacity": types.Int64Value(2),
			"min_size":         types.Int64Value(1),
			"max_size":         types.Int64Value(10),
		}
		obj, _ := types.ObjectValue(attrTypes, attrValues)
		elements[i] = obj
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SortNodeGroupList(ctx, elements)
	}
}
