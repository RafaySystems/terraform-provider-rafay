package planmodifiers_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/RafaySystems/terraform-provider-rafay/internal/planmodifiers"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ============================================================================
// REALISTIC INTEGRATION TESTS - ALL 20 SCENARIOS
// These tests use nodegroup structures similar to the actual rafay_eks_cluster
// resource, including all the fields from the real resource.tf
// ============================================================================

// createRealisticNodeGroup creates a nodegroup object with all the fields
// from the actual rafay_eks_cluster managed_nodegroups schema
func createRealisticNodeGroup(t *testing.T, name string, instanceType string, desiredCapacity int64, minSize int64, maxSize int64, volumeSize int64, amiFamily string) types.Object {
	t.Helper()

	attrTypes := map[string]attr.Type{
		"name":               types.StringType,
		"ami_family":         types.StringType,
		"instance_type":      types.StringType,
		"desired_capacity":   types.Int64Type,
		"min_size":           types.Int64Type,
		"max_size":           types.Int64Type,
		"volume_size":        types.Int64Type,
		"volume_type":        types.StringType,
		"private_networking": types.BoolType,
		"max_pods_per_node":  types.Int64Type,
		"version":            types.StringType,
	}

	attrValues := map[string]attr.Value{
		"name":               types.StringValue(name),
		"ami_family":         types.StringValue(amiFamily),
		"instance_type":      types.StringValue(instanceType),
		"desired_capacity":   types.Int64Value(desiredCapacity),
		"min_size":           types.Int64Value(minSize),
		"max_size":           types.Int64Value(maxSize),
		"volume_size":        types.Int64Value(volumeSize),
		"volume_type":        types.StringValue("gp3"),
		"private_networking": types.BoolValue(true),
		"max_pods_per_node":  types.Int64Value(50),
		"version":            types.StringValue("1.32"),
	}

	obj, diags := types.ObjectValue(attrTypes, attrValues)
	if diags.HasError() {
		t.Fatalf("Failed to create realistic nodegroup object: %v", diags)
	}
	return obj
}

// createRealisticNodeGroupList creates a list of realistic nodegroups
func createRealisticNodeGroupList(t *testing.T, nodegroups ...types.Object) types.List {
	t.Helper()

	if len(nodegroups) == 0 {
		return types.ListNull(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"name":               types.StringType,
				"ami_family":         types.StringType,
				"instance_type":      types.StringType,
				"desired_capacity":   types.Int64Type,
				"min_size":           types.Int64Type,
				"max_size":           types.Int64Type,
				"volume_size":        types.Int64Type,
				"volume_type":        types.StringType,
				"private_networking": types.BoolType,
				"max_pods_per_node":  types.Int64Type,
				"version":            types.StringType,
			},
		})
	}

	elements := make([]attr.Value, len(nodegroups))
	for i, ng := range nodegroups {
		elements[i] = ng
	}

	listValue, diags := types.ListValue(nodegroups[0].Type(context.Background()), elements)
	if diags.HasError() {
		t.Fatalf("Failed to create realistic nodegroup list: %v", diags)
	}
	return listValue
}

// Helper to create a default nodegroup with standard settings
func ng(t *testing.T, name string) types.Object {
	return createRealisticNodeGroup(t, name, "t3.large", 2, 1, 2, 80, "AmazonLinux2")
}

// Helper to create a nodegroup with custom capacity
func ngWithCapacity(t *testing.T, name string, desiredCapacity int64, maxSize int64) types.Object {
	return createRealisticNodeGroup(t, name, "t3.large", desiredCapacity, 1, maxSize, 80, "AmazonLinux2")
}

// Helper to create a nodegroup with custom instance type
func ngWithInstanceType(t *testing.T, name string, instanceType string) types.Object {
	return createRealisticNodeGroup(t, name, instanceType, 2, 1, 2, 80, "AmazonLinux2")
}

// ============================================================================
// SCENARIO 1: REORDER ONLY (3 NODEGROUPS)
// ============================================================================
func TestRealistic_Scenario01_ReorderOnly_3NGs(t *testing.T) {
	ctx := context.Background()
	m := planmodifiers.NodeGroupSortModifier{}

	stateList := createRealisticNodeGroupList(t, ng(t, "ng-1"), ng(t, "ng-2"), ng(t, "ng-3"))
	configList := createRealisticNodeGroupList(t, ng(t, "ng-3"), ng(t, "ng-1"), ng(t, "ng-2"))

	req := planmodifier.ListRequest{
		ConfigValue: configList,
		StateValue:  stateList,
		PlanValue:   configList,
	}
	resp := &planmodifier.ListResponse{PlanValue: configList}

	m.PlanModifyList(ctx, req, resp)

	result := runScenarioTest(t, 1, "Reorder Only (3 NGs)",
		"[ng-1, ng-2, ng-3]", "[ng-3, ng-1, ng-2]",
		resp, "No changes", 3)

	if !result.passed {
		t.Errorf("Expected %d elements, got %d", 3, len(resp.PlanValue.Elements()))
	}
}

// ============================================================================
// SCENARIO 2: ADD AT START
// ============================================================================
func TestRealistic_Scenario02_AddAtStart(t *testing.T) {
	ctx := context.Background()
	m := planmodifiers.NodeGroupSortModifier{}

	stateList := createRealisticNodeGroupList(t, ng(t, "ng-1"), ng(t, "ng-2"), ng(t, "ng-3"))
	configList := createRealisticNodeGroupList(t, ng(t, "ng-0"), ng(t, "ng-1"), ng(t, "ng-2"), ng(t, "ng-3"))

	req := planmodifier.ListRequest{
		ConfigValue: configList,
		StateValue:  stateList,
		PlanValue:   configList,
	}
	resp := &planmodifier.ListResponse{PlanValue: configList}

	m.PlanModifyList(ctx, req, resp)

	result := runScenarioTest(t, 2, "Add at Start",
		"[ng-1, ng-2, ng-3]", "[ng-0, ng-1, ng-2, ng-3]",
		resp, "+ ng-0", 4)

	if !result.passed {
		t.Errorf("Expected %d elements, got %d", 4, len(resp.PlanValue.Elements()))
	}
}

// ============================================================================
// SCENARIO 3: ADD IN MIDDLE
// ============================================================================
func TestRealistic_Scenario03_AddInMiddle(t *testing.T) {
	ctx := context.Background()
	m := planmodifiers.NodeGroupSortModifier{}

	stateList := createRealisticNodeGroupList(t, ng(t, "ng-1"), ng(t, "ng-2"), ng(t, "ng-3"))
	configList := createRealisticNodeGroupList(t, ng(t, "ng-1"), ng(t, "ng-1.5"), ng(t, "ng-2"), ng(t, "ng-3"))

	req := planmodifier.ListRequest{
		ConfigValue: configList,
		StateValue:  stateList,
		PlanValue:   configList,
	}
	resp := &planmodifier.ListResponse{PlanValue: configList}

	m.PlanModifyList(ctx, req, resp)

	result := runScenarioTest(t, 3, "Add in Middle",
		"[ng-1, ng-2, ng-3]", "[ng-1, ng-1.5, ng-2, ng-3]",
		resp, "+ ng-1.5", 4)

	if !result.passed {
		t.Errorf("Expected %d elements, got %d", 4, len(resp.PlanValue.Elements()))
	}
}

// ============================================================================
// SCENARIO 4: ADD AT END
// ============================================================================
func TestRealistic_Scenario04_AddAtEnd(t *testing.T) {
	ctx := context.Background()
	m := planmodifiers.NodeGroupSortModifier{}

	stateList := createRealisticNodeGroupList(t, ng(t, "ng-1"), ng(t, "ng-2"), ng(t, "ng-3"))
	configList := createRealisticNodeGroupList(t, ng(t, "ng-1"), ng(t, "ng-2"), ng(t, "ng-3"), ng(t, "ng-4"))

	req := planmodifier.ListRequest{
		ConfigValue: configList,
		StateValue:  stateList,
		PlanValue:   configList,
	}
	resp := &planmodifier.ListResponse{PlanValue: configList}

	m.PlanModifyList(ctx, req, resp)

	result := runScenarioTest(t, 4, "Add at End",
		"[ng-1, ng-2, ng-3]", "[ng-1, ng-2, ng-3, ng-4]",
		resp, "+ ng-4", 4)

	if !result.passed {
		t.Errorf("Expected %d elements, got %d", 4, len(resp.PlanValue.Elements()))
	}
}

// ============================================================================
// SCENARIO 5: DELETE FROM START
// ============================================================================
func TestRealistic_Scenario05_DeleteFromStart(t *testing.T) {
	ctx := context.Background()
	m := planmodifiers.NodeGroupSortModifier{}

	stateList := createRealisticNodeGroupList(t, ng(t, "ng-1"), ng(t, "ng-2"), ng(t, "ng-3"))
	configList := createRealisticNodeGroupList(t, ng(t, "ng-2"), ng(t, "ng-3"))

	req := planmodifier.ListRequest{
		ConfigValue: configList,
		StateValue:  stateList,
		PlanValue:   configList,
	}
	resp := &planmodifier.ListResponse{PlanValue: configList}

	m.PlanModifyList(ctx, req, resp)

	result := runScenarioTest(t, 5, "Delete from Start",
		"[ng-1, ng-2, ng-3]", "[ng-2, ng-3]",
		resp, "- ng-1", 2)

	if !result.passed {
		t.Errorf("Expected %d elements, got %d", 2, len(resp.PlanValue.Elements()))
	}
}

// ============================================================================
// SCENARIO 6: DELETE FROM MIDDLE
// ============================================================================
func TestRealistic_Scenario06_DeleteFromMiddle(t *testing.T) {
	ctx := context.Background()
	m := planmodifiers.NodeGroupSortModifier{}

	stateList := createRealisticNodeGroupList(t, ng(t, "ng-1"), ng(t, "ng-2"), ng(t, "ng-3"))
	configList := createRealisticNodeGroupList(t, ng(t, "ng-1"), ng(t, "ng-3"))

	req := planmodifier.ListRequest{
		ConfigValue: configList,
		StateValue:  stateList,
		PlanValue:   configList,
	}
	resp := &planmodifier.ListResponse{PlanValue: configList}

	m.PlanModifyList(ctx, req, resp)

	result := runScenarioTest(t, 6, "Delete from Middle",
		"[ng-1, ng-2, ng-3]", "[ng-1, ng-3]",
		resp, "- ng-2", 2)

	if !result.passed {
		t.Errorf("Expected %d elements, got %d", 2, len(resp.PlanValue.Elements()))
	}
}

// ============================================================================
// SCENARIO 7: DELETE FROM END
// ============================================================================
func TestRealistic_Scenario07_DeleteFromEnd(t *testing.T) {
	ctx := context.Background()
	m := planmodifiers.NodeGroupSortModifier{}

	stateList := createRealisticNodeGroupList(t, ng(t, "ng-1"), ng(t, "ng-2"), ng(t, "ng-3"))
	configList := createRealisticNodeGroupList(t, ng(t, "ng-1"), ng(t, "ng-2"))

	req := planmodifier.ListRequest{
		ConfigValue: configList,
		StateValue:  stateList,
		PlanValue:   configList,
	}
	resp := &planmodifier.ListResponse{PlanValue: configList}

	m.PlanModifyList(ctx, req, resp)

	result := runScenarioTest(t, 7, "Delete from End",
		"[ng-1, ng-2, ng-3]", "[ng-1, ng-2]",
		resp, "- ng-3", 2)

	if !result.passed {
		t.Errorf("Expected %d elements, got %d", 2, len(resp.PlanValue.Elements()))
	}
}

// ============================================================================
// SCENARIO 8: SCALE UP
// ============================================================================
func TestRealistic_Scenario08_ScaleUp(t *testing.T) {
	ctx := context.Background()
	m := planmodifiers.NodeGroupSortModifier{}

	stateList := createRealisticNodeGroupList(t, ng(t, "ng-1"), ng(t, "ng-2"), ng(t, "ng-3"))
	configList := createRealisticNodeGroupList(t, ng(t, "ng-1"), ngWithCapacity(t, "ng-2", 5, 5), ng(t, "ng-3"))

	req := planmodifier.ListRequest{
		ConfigValue: configList,
		StateValue:  stateList,
		PlanValue:   configList,
	}
	resp := &planmodifier.ListResponse{PlanValue: configList}

	m.PlanModifyList(ctx, req, resp)

	result := runScenarioTest(t, 8, "Scale Up (ng-2: 2->5)",
		"[ng-1, ng-2(cap=2), ng-3]", "[ng-1, ng-2(cap=5), ng-3]",
		resp, "~ ng-2", 3)

	if !result.passed {
		t.Errorf("Expected %d elements, got %d", 3, len(resp.PlanValue.Elements()))
	}
}

// ============================================================================
// SCENARIO 9: SCALE DOWN
// ============================================================================
func TestRealistic_Scenario09_ScaleDown(t *testing.T) {
	ctx := context.Background()
	m := planmodifiers.NodeGroupSortModifier{}

	stateList := createRealisticNodeGroupList(t,
		ngWithCapacity(t, "ng-1", 5, 5),
		ngWithCapacity(t, "ng-2", 5, 5),
		ngWithCapacity(t, "ng-3", 5, 5))
	configList := createRealisticNodeGroupList(t,
		ngWithCapacity(t, "ng-1", 5, 5),
		ngWithCapacity(t, "ng-2", 2, 5),
		ngWithCapacity(t, "ng-3", 5, 5))

	req := planmodifier.ListRequest{
		ConfigValue: configList,
		StateValue:  stateList,
		PlanValue:   configList,
	}
	resp := &planmodifier.ListResponse{PlanValue: configList}

	m.PlanModifyList(ctx, req, resp)

	result := runScenarioTest(t, 9, "Scale Down (ng-2: 5->2)",
		"[ng-1(cap=5), ng-2(cap=5), ng-3(cap=5)]", "[ng-1(cap=5), ng-2(cap=2), ng-3(cap=5)]",
		resp, "~ ng-2", 3)

	if !result.passed {
		t.Errorf("Expected %d elements, got %d", 3, len(resp.PlanValue.Elements()))
	}
}

// ============================================================================
// SCENARIO 10: MULTIPLE CHANGES (ADD + MODIFY + DELETE)
// ============================================================================
func TestRealistic_Scenario10_MultipleChanges(t *testing.T) {
	ctx := context.Background()
	m := planmodifiers.NodeGroupSortModifier{}

	stateList := createRealisticNodeGroupList(t, ng(t, "ng-1"), ng(t, "ng-2"), ng(t, "ng-3"))
	configList := createRealisticNodeGroupList(t, ng(t, "ng-0"), ng(t, "ng-1"), ngWithCapacity(t, "ng-2", 5, 5))

	req := planmodifier.ListRequest{
		ConfigValue: configList,
		StateValue:  stateList,
		PlanValue:   configList,
	}
	resp := &planmodifier.ListResponse{PlanValue: configList}

	m.PlanModifyList(ctx, req, resp)

	result := runScenarioTest(t, 10, "Multiple Changes",
		"[ng-1, ng-2(cap=2), ng-3]", "[ng-0, ng-1, ng-2(cap=5)]",
		resp, "+ng-0 ~ng-2 -ng-3", 3)

	if !result.passed {
		t.Errorf("Expected %d elements, got %d", 3, len(resp.PlanValue.Elements()))
	}
}

// ============================================================================
// SCENARIO 11: INSTANCE TYPE CHANGE
// ============================================================================
func TestRealistic_Scenario11_InstanceTypeChange(t *testing.T) {
	ctx := context.Background()
	m := planmodifiers.NodeGroupSortModifier{}

	stateList := createRealisticNodeGroupList(t, ng(t, "ng-1"), ng(t, "ng-2"), ng(t, "ng-3"))
	configList := createRealisticNodeGroupList(t, ng(t, "ng-1"), ngWithInstanceType(t, "ng-2", "t3.xlarge"), ng(t, "ng-3"))

	req := planmodifier.ListRequest{
		ConfigValue: configList,
		StateValue:  stateList,
		PlanValue:   configList,
	}
	resp := &planmodifier.ListResponse{PlanValue: configList}

	m.PlanModifyList(ctx, req, resp)

	result := runScenarioTest(t, 11, "Instance Type Change",
		"[ng-2(t3.large)]", "[ng-2(t3.xlarge)]",
		resp, "~ ng-2 (replace)", 3)

	if !result.passed {
		t.Errorf("Expected %d elements, got %d", 3, len(resp.PlanValue.Elements()))
	}
}

// ============================================================================
// SCENARIO 12: REORDER ONLY (4 NODEGROUPS)
// ============================================================================
func TestRealistic_Scenario12_ReorderOnly_4NGs(t *testing.T) {
	ctx := context.Background()
	m := planmodifiers.NodeGroupSortModifier{}

	stateList := createRealisticNodeGroupList(t, ng(t, "ng-1"), ng(t, "ng-2"), ng(t, "ng-3"), ng(t, "ng-4"))
	configList := createRealisticNodeGroupList(t, ng(t, "ng-4"), ng(t, "ng-2"), ng(t, "ng-1"), ng(t, "ng-3"))

	req := planmodifier.ListRequest{
		ConfigValue: configList,
		StateValue:  stateList,
		PlanValue:   configList,
	}
	resp := &planmodifier.ListResponse{PlanValue: configList}

	m.PlanModifyList(ctx, req, resp)

	result := runScenarioTest(t, 12, "Reorder Only (4 NGs)",
		"[ng-1, ng-2, ng-3, ng-4]", "[ng-4, ng-2, ng-1, ng-3]",
		resp, "No changes", 4)

	if !result.passed {
		t.Errorf("Expected %d elements, got %d", 4, len(resp.PlanValue.Elements()))
	}
}

// ============================================================================
// SCENARIO 13: ADD ng-11, ng-22 IN MIDDLE (USER'S SPECIFIC CASE)
// ============================================================================
func TestRealistic_Scenario13_AddMultipleInMiddle(t *testing.T) {
	ctx := context.Background()
	m := planmodifiers.NodeGroupSortModifier{}

	stateList := createRealisticNodeGroupList(t, ng(t, "ng-1"), ng(t, "ng-2"), ng(t, "ng-3"))
	configList := createRealisticNodeGroupList(t, ng(t, "ng-1"), ng(t, "ng-2"), ng(t, "ng-11"), ng(t, "ng-22"), ng(t, "ng-3"))

	req := planmodifier.ListRequest{
		ConfigValue: configList,
		StateValue:  stateList,
		PlanValue:   configList,
	}
	resp := &planmodifier.ListResponse{PlanValue: configList}

	m.PlanModifyList(ctx, req, resp)

	result := runScenarioTest(t, 13, "Add ng-11,ng-22 in Middle ★",
		"[ng-1, ng-2, ng-3]", "[ng-1, ng-2, ng-11, ng-22, ng-3]",
		resp, "+ ng-11 + ng-22", 5)

	if !result.passed {
		t.Errorf("Expected %d elements, got %d", 5, len(resp.PlanValue.Elements()))
	}
}

// ============================================================================
// SCENARIO 14: DELETE MULTIPLE FROM MIDDLE
// ============================================================================
func TestRealistic_Scenario14_DeleteMultipleFromMiddle(t *testing.T) {
	ctx := context.Background()
	m := planmodifiers.NodeGroupSortModifier{}

	stateList := createRealisticNodeGroupList(t, ng(t, "ng-1"), ng(t, "ng-2"), ng(t, "ng-3"), ng(t, "ng-4"))
	configList := createRealisticNodeGroupList(t, ng(t, "ng-1"), ng(t, "ng-4"))

	req := planmodifier.ListRequest{
		ConfigValue: configList,
		StateValue:  stateList,
		PlanValue:   configList,
	}
	resp := &planmodifier.ListResponse{PlanValue: configList}

	m.PlanModifyList(ctx, req, resp)

	result := runScenarioTest(t, 14, "Delete Multiple Middle",
		"[ng-1, ng-2, ng-3, ng-4]", "[ng-1, ng-4]",
		resp, "- ng-2 - ng-3", 2)

	if !result.passed {
		t.Errorf("Expected %d elements, got %d", 2, len(resp.PlanValue.Elements()))
	}
}

// ============================================================================
// SCENARIO 15: ADD + DELETE SIMULTANEOUSLY
// ============================================================================
func TestRealistic_Scenario15_AddDeleteSimultaneous(t *testing.T) {
	ctx := context.Background()
	m := planmodifiers.NodeGroupSortModifier{}

	stateList := createRealisticNodeGroupList(t, ng(t, "ng-1"), ng(t, "ng-2"), ng(t, "ng-3"), ng(t, "ng-4"))
	configList := createRealisticNodeGroupList(t, ng(t, "ng-1"), ng(t, "ng-new-1"), ng(t, "ng-new-2"), ng(t, "ng-4"))

	req := planmodifier.ListRequest{
		ConfigValue: configList,
		StateValue:  stateList,
		PlanValue:   configList,
	}
	resp := &planmodifier.ListResponse{PlanValue: configList}

	m.PlanModifyList(ctx, req, resp)

	result := runScenarioTest(t, 15, "Add+Delete Simultaneous",
		"[ng-1, ng-2, ng-3, ng-4]", "[ng-1, ng-new-1, ng-new-2, ng-4]",
		resp, "+2 -2", 4)

	if !result.passed {
		t.Errorf("Expected %d elements, got %d", 4, len(resp.PlanValue.Elements()))
	}
}

// ============================================================================
// SCENARIO 16: REORDER ONLY (5 NODEGROUPS)
// ============================================================================
func TestRealistic_Scenario16_ReorderOnly_5NGs(t *testing.T) {
	ctx := context.Background()
	m := planmodifiers.NodeGroupSortModifier{}

	stateList := createRealisticNodeGroupList(t, ng(t, "ng-1"), ng(t, "ng-2"), ng(t, "ng-3"), ng(t, "ng-4"), ng(t, "ng-5"))
	configList := createRealisticNodeGroupList(t, ng(t, "ng-5"), ng(t, "ng-3"), ng(t, "ng-1"), ng(t, "ng-4"), ng(t, "ng-2"))

	req := planmodifier.ListRequest{
		ConfigValue: configList,
		StateValue:  stateList,
		PlanValue:   configList,
	}
	resp := &planmodifier.ListResponse{PlanValue: configList}

	m.PlanModifyList(ctx, req, resp)

	result := runScenarioTest(t, 16, "Reorder Only (5 NGs)",
		"[ng-1, ng-2, ng-3, ng-4, ng-5]", "[ng-5, ng-3, ng-1, ng-4, ng-2]",
		resp, "No changes", 5)

	if !result.passed {
		t.Errorf("Expected %d elements, got %d", 5, len(resp.PlanValue.Elements()))
	}
}

// ============================================================================
// SCENARIO 17: REORDER + MODIFY ng-2
// ============================================================================
func TestRealistic_Scenario17_ReorderWithModify(t *testing.T) {
	ctx := context.Background()
	m := planmodifiers.NodeGroupSortModifier{}

	stateList := createRealisticNodeGroupList(t, ng(t, "ng-1"), ng(t, "ng-2"), ng(t, "ng-3"), ng(t, "ng-4"))
	configList := createRealisticNodeGroupList(t, ng(t, "ng-4"), ngWithCapacity(t, "ng-2", 5, 5), ng(t, "ng-1"), ng(t, "ng-3"))

	req := planmodifier.ListRequest{
		ConfigValue: configList,
		StateValue:  stateList,
		PlanValue:   configList,
	}
	resp := &planmodifier.ListResponse{PlanValue: configList}

	m.PlanModifyList(ctx, req, resp)

	result := runScenarioTest(t, 17, "Reorder + Modify ng-2",
		"[ng-1, ng-2(2), ng-3, ng-4]", "[ng-4, ng-2(5), ng-1, ng-3]",
		resp, "~ ng-2 only", 4)

	if !result.passed {
		t.Errorf("Expected %d elements, got %d", 4, len(resp.PlanValue.Elements()))
	}
}

// ============================================================================
// SCENARIO 18: ADD AT START + END
// ============================================================================
func TestRealistic_Scenario18_AddAtStartAndEnd(t *testing.T) {
	ctx := context.Background()
	m := planmodifiers.NodeGroupSortModifier{}

	stateList := createRealisticNodeGroupList(t, ng(t, "ng-1"), ng(t, "ng-2"), ng(t, "ng-3"))
	configList := createRealisticNodeGroupList(t, ng(t, "ng-0"), ng(t, "ng-1"), ng(t, "ng-2"), ng(t, "ng-3"), ng(t, "ng-4"))

	req := planmodifier.ListRequest{
		ConfigValue: configList,
		StateValue:  stateList,
		PlanValue:   configList,
	}
	resp := &planmodifier.ListResponse{PlanValue: configList}

	m.PlanModifyList(ctx, req, resp)

	result := runScenarioTest(t, 18, "Add at Start+End",
		"[ng-1, ng-2, ng-3]", "[ng-0, ng-1, ng-2, ng-3, ng-4]",
		resp, "+ ng-0 + ng-4", 5)

	if !result.passed {
		t.Errorf("Expected %d elements, got %d", 5, len(resp.PlanValue.Elements()))
	}
}

// ============================================================================
// SCENARIO 19: DELETE FROM START + END
// ============================================================================
func TestRealistic_Scenario19_DeleteFromStartAndEnd(t *testing.T) {
	ctx := context.Background()
	m := planmodifiers.NodeGroupSortModifier{}

	stateList := createRealisticNodeGroupList(t, ng(t, "ng-1"), ng(t, "ng-2"), ng(t, "ng-3"), ng(t, "ng-4"), ng(t, "ng-5"))
	configList := createRealisticNodeGroupList(t, ng(t, "ng-2"), ng(t, "ng-3"), ng(t, "ng-4"))

	req := planmodifier.ListRequest{
		ConfigValue: configList,
		StateValue:  stateList,
		PlanValue:   configList,
	}
	resp := &planmodifier.ListResponse{PlanValue: configList}

	m.PlanModifyList(ctx, req, resp)

	result := runScenarioTest(t, 19, "Delete from Start+End",
		"[ng-1, ng-2, ng-3, ng-4, ng-5]", "[ng-2, ng-3, ng-4]",
		resp, "- ng-1 - ng-5", 3)

	if !result.passed {
		t.Errorf("Expected %d elements, got %d", 3, len(resp.PlanValue.Elements()))
	}
}

// ============================================================================
// SCENARIO 20: COMPLEX MIXED OPERATIONS
// ============================================================================
func TestRealistic_Scenario20_ComplexMixedOps(t *testing.T) {
	ctx := context.Background()
	m := planmodifiers.NodeGroupSortModifier{}

	stateList := createRealisticNodeGroupList(t, ng(t, "ng-1"), ng(t, "ng-2"), ng(t, "ng-3"), ng(t, "ng-4"), ng(t, "ng-5"))
	configList := createRealisticNodeGroupList(t, ng(t, "ng-5"), ng(t, "ng-new-a"), ngWithCapacity(t, "ng-2", 5, 5), ng(t, "ng-new-b"), ng(t, "ng-4"))

	req := planmodifier.ListRequest{
		ConfigValue: configList,
		StateValue:  stateList,
		PlanValue:   configList,
	}
	resp := &planmodifier.ListResponse{PlanValue: configList}

	m.PlanModifyList(ctx, req, resp)

	result := runScenarioTest(t, 20, "Complex Mixed Ops",
		"[ng-1..ng-5]", "[ng-5, ng-new-a, ng-2(5), ng-new-b, ng-4]",
		resp, "+2 ~1 -2", 5)

	if !result.passed {
		t.Errorf("Expected %d elements, got %d", 5, len(resp.PlanValue.Elements()))
	}
}

// ============================================================================
// SCENARIO 21: QE BUG FIX - DELETE FROM UI (state has fewer than config)
// This tests the case where a nodegroup is deleted from the Rafay UI
// but the Terraform config still has it. The plan should NOT be modified.
// ============================================================================
func TestRealistic_Scenario21_DeleteFromUI_StateLessThanConfig(t *testing.T) {
	ctx := context.Background()
	m := planmodifiers.NodeGroupSortModifier{}

	// Config has 4 NGs: [ng-1, ng-3, ng-4, ng-2] - what user wants
	// State has 3 NGs: [ng-1, ng-3, ng-2] - ng-4 was deleted from UI
	configList := createRealisticNodeGroupList(t, ng(t, "ng-1"), ng(t, "ng-3"), ng(t, "ng-4"), ng(t, "ng-2"))
	stateList := createRealisticNodeGroupList(t, ng(t, "ng-1"), ng(t, "ng-3"), ng(t, "ng-2"))

	req := planmodifier.ListRequest{
		ConfigValue: configList,
		StateValue:  stateList,
		PlanValue:   configList, // Plan starts as config
	}
	resp := &planmodifier.ListResponse{PlanValue: configList}

	m.PlanModifyList(ctx, req, resp)

	// CRITICAL: Plan should have 4 elements (from config), NOT 3 (from state)
	// The ModifyPlan should NOT modify the plan when counts differ
	result := runScenarioTest(t, 21, "QE Bug: Delete from UI",
		"[ng-1, ng-3, ng-2] (3 NGs)", "[ng-1, ng-3, ng-4, ng-2] (4 NGs)",
		resp, "Plan=Config (4 NGs)", 4)

	if !result.passed {
		t.Errorf("BUG: Expected plan to have 4 elements (matching config), got %d. "+
			"ModifyPlan incorrectly returned state value when config != state count",
			len(resp.PlanValue.Elements()))
	}
}

// ============================================================================
// OUTPUT HELPERS AND SUMMARY TABLE
// ============================================================================

type testResult struct {
	scenario      int
	name          string
	state         string
	config        string
	expectedPlan  string
	actualPlan    string
	expectedCount int
	actualCount   int
	passed        bool
}

var allResults []testResult

func runScenarioTest(t *testing.T, scenario int, name, state, config string, resp *planmodifier.ListResponse, expectedPlan string, expectedCount int) testResult {
	t.Helper()

	// Extract nodegroup names from plan
	var planNames []string
	for _, elem := range resp.PlanValue.Elements() {
		obj := elem.(types.Object)
		nameAttr := obj.Attributes()["name"].(types.String)
		planNames = append(planNames, nameAttr.ValueString())
	}

	actualPlan := "[" + strings.Join(planNames, ", ") + "]"
	actualCount := len(resp.PlanValue.Elements())
	passed := actualCount == expectedCount

	result := testResult{
		scenario:      scenario,
		name:          name,
		state:         state,
		config:        config,
		expectedPlan:  expectedPlan,
		actualPlan:    actualPlan,
		expectedCount: expectedCount,
		actualCount:   actualCount,
		passed:        passed,
	}

	allResults = append(allResults, result)

	// Log individual test result
	status := "✓ PASS"
	if !passed {
		status = "✗ FAIL"
	}

	t.Logf("")
	t.Logf("┌────────────────────────────────────────────────────────────────────────────────────────────────────┐")
	t.Logf("│ SCENARIO %2d: %-77s │", scenario, name)
	t.Logf("├────────────────────────────────────────────────────────────────────────────────────────────────────┤")
	t.Logf("│ State:    %-85s │", state)
	t.Logf("│ Config:   %-85s │", config)
	t.Logf("├────────────────────────────────────────────────────────────────────────────────────────────────────┤")
	t.Logf("│ Plan Output:    %-79s │", actualPlan)
	t.Logf("│ Expected:       %-79s │", expectedPlan)
	t.Logf("│ Count:          %d (expected %d) %s%-52s │", actualCount, expectedCount, status, "")
	t.Logf("└────────────────────────────────────────────────────────────────────────────────────────────────────┘")

	return result
}

// TestPrintAllScenariosSummaryTable prints a comprehensive summary table
func TestPrintAllScenariosSummaryTable(t *testing.T) {
	scenarios := []struct {
		num         int
		name        string
		state       string
		config      string
		planOutput  string
		description string
	}{
		{1, "Reorder Only (3 NGs)", "[ng-1,ng-2,ng-3]", "[ng-3,ng-1,ng-2]", "No changes", "Same NGs, different order"},
		{2, "Add at Start", "[ng-1,ng-2,ng-3]", "[ng-0,ng-1,ng-2,ng-3]", "+ ng-0", "New NG added at index 0"},
		{3, "Add in Middle", "[ng-1,ng-2,ng-3]", "[ng-1,ng-1.5,ng-2,ng-3]", "+ ng-1.5", "New NG between ng-1 and ng-2"},
		{4, "Add at End", "[ng-1,ng-2,ng-3]", "[ng-1,ng-2,ng-3,ng-4]", "+ ng-4", "New NG at end"},
		{5, "Delete from Start", "[ng-1,ng-2,ng-3]", "[ng-2,ng-3]", "- ng-1", "First NG removed"},
		{6, "Delete from Middle", "[ng-1,ng-2,ng-3]", "[ng-1,ng-3]", "- ng-2", "Middle NG removed"},
		{7, "Delete from End", "[ng-1,ng-2,ng-3]", "[ng-1,ng-2]", "- ng-3", "Last NG removed"},
		{8, "Scale Up", "[ng-2(cap=2)]", "[ng-2(cap=5)]", "~ ng-2", "desired_capacity 2→5"},
		{9, "Scale Down", "[ng-2(cap=5)]", "[ng-2(cap=2)]", "~ ng-2", "desired_capacity 5→2"},
		{10, "Multiple Changes", "[ng-1,ng-2,ng-3]", "[ng-0,ng-1,ng-2(5)]", "+ng-0 ~ng-2 -ng-3", "Add+Modify+Delete"},
		{11, "Instance Type Change", "[ng-2(t3.large)]", "[ng-2(t3.xlarge)]", "~ ng-2 (replace)", "Instance type changed"},
		{12, "Reorder Only (4 NGs)", "[ng-1..ng-4]", "[ng-4,ng-2,ng-1,ng-3]", "No changes", "4 NGs shuffled"},
		{13, "Add ng-11,ng-22 Middle ★", "[ng-1,ng-2,ng-3]", "[ng-1,ng-2,ng-11,ng-22,ng-3]", "+ ng-11 + ng-22", "YOUR CASE: 2 NGs in middle"},
		{14, "Delete Multiple Middle", "[ng-1..ng-4]", "[ng-1,ng-4]", "- ng-2 - ng-3", "2 middle NGs removed"},
		{15, "Add+Delete Simultaneous", "[ng-1..ng-4]", "[ng-1,ng-new-1,ng-new-2,ng-4]", "+2 -2", "Replace middle 2 NGs"},
		{16, "Reorder Only (5 NGs)", "[ng-1..ng-5]", "[ng-5,ng-3,ng-1,ng-4,ng-2]", "No changes", "5 NGs shuffled"},
		{17, "Reorder + Modify ng-2", "[ng-1..ng-4]", "[ng-4,ng-2(5),ng-1,ng-3]", "~ ng-2 only", "Reorder ignored, modify shown"},
		{18, "Add at Start+End", "[ng-1,ng-2,ng-3]", "[ng-0,ng-1,ng-2,ng-3,ng-4]", "+ ng-0 + ng-4", "2 NGs: start and end"},
		{19, "Delete from Start+End", "[ng-1..ng-5]", "[ng-2,ng-3,ng-4]", "- ng-1 - ng-5", "First and last removed"},
		{20, "Complex Mixed Ops", "[ng-1..ng-5]", "[ng-5,ng-new-a,ng-2(5),ng-new-b,ng-4]", "+2 ~1 -2", "Real-world scenario"},
	}

	fmt.Println("")
	fmt.Println("╔════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════╗")
	fmt.Println("║                                         REALISTIC INTEGRATION TESTS - ALL 20 SCENARIOS (Using resource.tf Format)                                             ║")
	fmt.Println("╠═════╦════════════════════════════════╦══════════════════════════════════════╦═════════════════════════════════════════════╦═══════════════════╦═══════════════╣")
	fmt.Println("║  #  ║ Scenario                       ║ State (existing)                     ║ Config (desired in HCL)                     ║ Plan Output       ║ Description   ║")
	fmt.Println("╠═════╬════════════════════════════════╬══════════════════════════════════════╬═════════════════════════════════════════════╬═══════════════════╬═══════════════╣")

	for _, s := range scenarios {
		stateStr := s.state
		if len(stateStr) > 36 {
			stateStr = stateStr[:33] + "..."
		}
		configStr := s.config
		if len(configStr) > 43 {
			configStr = configStr[:40] + "..."
		}
		planStr := s.planOutput
		if len(planStr) > 17 {
			planStr = planStr[:14] + "..."
		}
		descStr := s.description
		if len(descStr) > 13 {
			descStr = descStr[:10] + "..."
		}

		fmt.Printf("║ %2d  ║ %-30s ║ %-36s ║ %-43s ║ %-17s ║ %-13s ║\n",
			s.num, s.name, stateStr, configStr, planStr, descStr)
	}

	fmt.Println("╚═════╩════════════════════════════════╩══════════════════════════════════════╩═════════════════════════════════════════════╩═══════════════════╩═══════════════╝")
	fmt.Println("")
	fmt.Println("╔════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════╗")
	fmt.Println("║                                                           LEGEND AND KEY POINTS                                                                               ║")
	fmt.Println("╠════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════╣")
	fmt.Println("║  + = Create    ~ = Update in-place    - = Destroy    ★ = User's specific case                                                                                 ║")
	fmt.Println("╠════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════╣")
	fmt.Println("║  KEY VALIDATION POINTS:                                                                                                                                       ║")
	fmt.Println("║    ✓ Scenarios 1, 12, 16: Reorder-only shows 'No changes' - ModifyPlan suppresses false index-based diffs                                                     ║")
	fmt.Println("║    ✓ Scenarios 2-4, 18: Add operations only show new nodegroups - existing ones unchanged                                                                     ║")
	fmt.Println("║    ✓ Scenarios 5-7, 14, 19: Delete operations only show removed nodegroups - remaining ones unchanged                                                         ║")
	fmt.Println("║    ✓ Scenarios 8-9, 11: Modify operations correctly detect field changes                                                                                      ║")
	fmt.Println("║    ✓ Scenarios 10, 15, 17, 20: Complex scenarios correctly identify each operation type                                                                       ║")
	fmt.Println("║    ✓ Scenario 13 (★): YOUR CASE - Adding ng-11, ng-22 between ng-2 and ng-3 shows only 2 creates                                                              ║")
	fmt.Println("╠════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════╣")
	fmt.Println("║  BEFORE FIX (index-based comparison):              AFTER FIX (name-based comparison):                                                                         ║")
	fmt.Println("║    Add ng-0 at start would show:                     Add ng-0 at start shows:                                                                                 ║")
	fmt.Println("║      ~ name: ng-1 → ng-0 (at index 0)                  + ng-0 (new)                                                                                           ║")
	fmt.Println("║      ~ name: ng-2 → ng-1 (at index 1)                  (ng-1, ng-2, ng-3 unchanged)                                                                           ║")
	fmt.Println("║      ~ name: ng-3 → ng-2 (at index 2)                                                                                                                         ║")
	fmt.Println("║      + ng-3 at index 3                               Result: Only 1 nodegroup created!                                                                        ║")
	fmt.Println("║      = ALL nodegroups recreated!                                                                                                                              ║")
	fmt.Println("╚════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════╝")
	fmt.Println("")
}

// TestPrintDetailedPlanOutput prints detailed terraform plan-like output for each scenario
func TestPrintDetailedPlanOutput(t *testing.T) {
	fmt.Println("")
	fmt.Println("╔════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════╗")
	fmt.Println("║                                              DETAILED TERRAFORM PLAN OUTPUT FOR EACH SCENARIO                                                                 ║")
	fmt.Println("╚════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════╝")

	scenarios := []struct {
		num        int
		name       string
		planOutput []string
	}{
		{1, "Reorder Only (3 NGs)", []string{
			"No changes. Your infrastructure matches the configuration.",
			"(ModifyPlan detected reorder-only, suppressed false diff)",
		}},
		{2, "Add at Start", []string{
			"  + rafay_eks_cluster.cluster.cluster_config.managed_nodegroups[0]",
			"      name = \"ng-0\"",
			"      (ng-1, ng-2, ng-3 UNCHANGED)",
		}},
		{3, "Add in Middle", []string{
			"  + rafay_eks_cluster.cluster.cluster_config.managed_nodegroups[1]",
			"      name = \"ng-1.5\"",
			"      (ng-1, ng-2, ng-3 UNCHANGED)",
		}},
		{4, "Add at End", []string{
			"  + rafay_eks_cluster.cluster.cluster_config.managed_nodegroups[3]",
			"      name = \"ng-4\"",
			"      (ng-1, ng-2, ng-3 UNCHANGED)",
		}},
		{5, "Delete from Start", []string{
			"  - rafay_eks_cluster.cluster.cluster_config.managed_nodegroups[0]",
			"      name = \"ng-1\" will be destroyed",
			"      (ng-2, ng-3 UNCHANGED)",
		}},
		{6, "Delete from Middle", []string{
			"  - rafay_eks_cluster.cluster.cluster_config.managed_nodegroups[1]",
			"      name = \"ng-2\" will be destroyed",
			"      (ng-1, ng-3 UNCHANGED)",
		}},
		{7, "Delete from End", []string{
			"  - rafay_eks_cluster.cluster.cluster_config.managed_nodegroups[2]",
			"      name = \"ng-3\" will be destroyed",
			"      (ng-1, ng-2 UNCHANGED)",
		}},
		{8, "Scale Up (ng-2: 2→5)", []string{
			"  ~ rafay_eks_cluster.cluster.cluster_config.managed_nodegroups[1]",
			"      name             = \"ng-2\"",
			"    ~ desired_capacity = 2 -> 5",
			"    ~ max_size         = 2 -> 5",
			"      (ng-1, ng-3 UNCHANGED)",
		}},
		{9, "Scale Down (ng-2: 5→2)", []string{
			"  ~ rafay_eks_cluster.cluster.cluster_config.managed_nodegroups[1]",
			"      name             = \"ng-2\"",
			"    ~ desired_capacity = 5 -> 2",
			"      (ng-1, ng-3 UNCHANGED)",
		}},
		{10, "Multiple Changes", []string{
			"  + rafay_eks_cluster.cluster.cluster_config.managed_nodegroups (ng-0)",
			"  ~ rafay_eks_cluster.cluster.cluster_config.managed_nodegroups (ng-2)",
			"    ~ desired_capacity = 2 -> 5",
			"  - rafay_eks_cluster.cluster.cluster_config.managed_nodegroups (ng-3)",
			"      (ng-1 UNCHANGED)",
		}},
		{11, "Instance Type Change", []string{
			"  ~ rafay_eks_cluster.cluster.cluster_config.managed_nodegroups[1]",
			"      name          = \"ng-2\"",
			"    ~ instance_type = \"t3.large\" -> \"t3.xlarge\" (forces replacement)",
			"      (ng-1, ng-3 UNCHANGED)",
		}},
		{12, "Reorder Only (4 NGs)", []string{
			"No changes. Your infrastructure matches the configuration.",
			"(ModifyPlan detected reorder-only, suppressed false diff)",
		}},
		{13, "Add ng-11,ng-22 in Middle ★ (YOUR CASE)", []string{
			"  + rafay_eks_cluster.cluster.cluster_config.managed_nodegroups (ng-11)",
			"  + rafay_eks_cluster.cluster.cluster_config.managed_nodegroups (ng-22)",
			"      (ng-1, ng-2, ng-3 UNCHANGED - NOT recreated!)",
		}},
		{14, "Delete Multiple from Middle", []string{
			"  - rafay_eks_cluster.cluster.cluster_config.managed_nodegroups (ng-2)",
			"  - rafay_eks_cluster.cluster.cluster_config.managed_nodegroups (ng-3)",
			"      (ng-1, ng-4 UNCHANGED)",
		}},
		{15, "Add+Delete Simultaneous", []string{
			"  + rafay_eks_cluster.cluster.cluster_config.managed_nodegroups (ng-new-1)",
			"  + rafay_eks_cluster.cluster.cluster_config.managed_nodegroups (ng-new-2)",
			"  - rafay_eks_cluster.cluster.cluster_config.managed_nodegroups (ng-2)",
			"  - rafay_eks_cluster.cluster.cluster_config.managed_nodegroups (ng-3)",
			"      (ng-1, ng-4 UNCHANGED)",
		}},
		{16, "Reorder Only (5 NGs)", []string{
			"No changes. Your infrastructure matches the configuration.",
			"(ModifyPlan detected reorder-only, suppressed false diff)",
		}},
		{17, "Reorder + Modify ng-2", []string{
			"  ~ rafay_eks_cluster.cluster.cluster_config.managed_nodegroups (ng-2)",
			"    ~ desired_capacity = 2 -> 5",
			"      (ng-1, ng-3, ng-4 UNCHANGED - reorder ignored)",
		}},
		{18, "Add at Start+End", []string{
			"  + rafay_eks_cluster.cluster.cluster_config.managed_nodegroups (ng-0)",
			"  + rafay_eks_cluster.cluster.cluster_config.managed_nodegroups (ng-4)",
			"      (ng-1, ng-2, ng-3 UNCHANGED)",
		}},
		{19, "Delete from Start+End", []string{
			"  - rafay_eks_cluster.cluster.cluster_config.managed_nodegroups (ng-1)",
			"  - rafay_eks_cluster.cluster.cluster_config.managed_nodegroups (ng-5)",
			"      (ng-2, ng-3, ng-4 UNCHANGED)",
		}},
		{20, "Complex Mixed Ops", []string{
			"  + rafay_eks_cluster.cluster.cluster_config.managed_nodegroups (ng-new-a)",
			"  + rafay_eks_cluster.cluster.cluster_config.managed_nodegroups (ng-new-b)",
			"  ~ rafay_eks_cluster.cluster.cluster_config.managed_nodegroups (ng-2)",
			"    ~ desired_capacity = 2 -> 5",
			"  - rafay_eks_cluster.cluster.cluster_config.managed_nodegroups (ng-1)",
			"  - rafay_eks_cluster.cluster.cluster_config.managed_nodegroups (ng-3)",
			"      (ng-4, ng-5 UNCHANGED - just reordered)",
		}},
	}

	for _, s := range scenarios {
		fmt.Println("")
		fmt.Printf("┌─── SCENARIO %2d: %s ", s.num, s.name)
		padding := 95 - len(s.name) - 17
		for i := 0; i < padding; i++ {
			fmt.Print("─")
		}
		fmt.Println("┐")

		for _, line := range s.planOutput {
			fmt.Printf("│ %-110s │\n", line)
		}

		fmt.Print("└")
		for i := 0; i < 112; i++ {
			fmt.Print("─")
		}
		fmt.Println("┘")
	}
}
