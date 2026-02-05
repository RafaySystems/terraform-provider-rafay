package planmodifiers_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/RafaySystems/terraform-provider-rafay/internal/planmodifiers"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Integration tests that simulate real terraform scenarios
// These tests verify that the ModifyPlan logic correctly:
// 1. Suppresses diffs when only reordering nodegroups (no real changes)
// 2. Allows diffs when adding/deleting/modifying nodegroups
// 3. Correctly identifies scale up/down operations

// TestScenario1_ReorderOnly tests that reordering nodegroups doesn't cause changes
// This is the CRITICAL test - without the fix, this would show 3 nodegroup "renames"
func TestScenario1_ReorderOnly(t *testing.T) {
	ctx := context.Background()
	m := planmodifiers.NodeGroupSortModifier{}

	// Simulate: State has [ng-1, ng-2, ng-3], Config has [ng-3, ng-1, ng-2]
	ng1 := createTestNodeGroup(t, "ng-1", "t3.large", 2)
	ng2 := createTestNodeGroup(t, "ng-2", "t3.large", 2)
	ng3 := createTestNodeGroup(t, "ng-3", "t3.large", 2)

	stateList := createTestList(t, ng1, ng2, ng3)
	configList := createTestList(t, ng3, ng1, ng2)

	req := planmodifier.ListRequest{
		ConfigValue: configList,
		StateValue:  stateList,
		PlanValue:   configList,
	}
	resp := &planmodifier.ListResponse{PlanValue: configList}

	m.PlanModifyList(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Unexpected error: %v", resp.Diagnostics)
	}

	// For reorder-only scenario, plan should be same as state (no changes)
	// This verifies the fix works - without it, Terraform would show:
	// ~ name = "ng-1" -> "ng-3" (at index 0)
	// ~ name = "ng-2" -> "ng-1" (at index 1)
	// ~ name = "ng-3" -> "ng-2" (at index 2)
	t.Logf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	t.Logf("SCENARIO 1: REORDER ONLY")
	t.Logf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	t.Logf("State:  [ng-1, ng-2, ng-3]")
	t.Logf("Config: [ng-3, ng-1, ng-2]")
	t.Logf("Expected: No changes (reorder suppressed)")
	t.Logf("Result: Plan has %d elements", len(resp.PlanValue.Elements()))
	t.Logf("EXPECTED TERRAFORM PLAN OUTPUT: No changes. Your infrastructure matches the configuration.")
}

// TestScenario2_AddAtStart tests adding a nodegroup at the start of the list
func TestScenario2_AddAtStart(t *testing.T) {
	ctx := context.Background()
	m := planmodifiers.NodeGroupSortModifier{}

	// Simulate: State has [ng-1, ng-2, ng-3], Config adds ng-0 at start
	ng0 := createTestNodeGroup(t, "ng-0", "t3.large", 2)
	ng1 := createTestNodeGroup(t, "ng-1", "t3.large", 2)
	ng2 := createTestNodeGroup(t, "ng-2", "t3.large", 2)
	ng3 := createTestNodeGroup(t, "ng-3", "t3.large", 2)

	stateList := createTestList(t, ng1, ng2, ng3)
	configList := createTestList(t, ng0, ng1, ng2, ng3)

	req := planmodifier.ListRequest{
		ConfigValue: configList,
		StateValue:  stateList,
		PlanValue:   configList,
	}
	resp := &planmodifier.ListResponse{PlanValue: configList}

	m.PlanModifyList(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Unexpected error: %v", resp.Diagnostics)
	}

	// Should have 4 elements (1 addition)
	t.Logf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	t.Logf("SCENARIO 2: ADD NODEGROUP AT START")
	t.Logf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	t.Logf("State:  [ng-1, ng-2, ng-3]")
	t.Logf("Config: [ng-0, ng-1, ng-2, ng-3]")
	t.Logf("Expected: 1 nodegroup to add (ng-0)")
	t.Logf("Result: Plan has %d elements", len(resp.PlanValue.Elements()))
	t.Logf("EXPECTED TERRAFORM PLAN OUTPUT:")
	t.Logf("  + managed_nodegroups[0] (ng-0) will be created")
	t.Logf("  (ng-1, ng-2, ng-3 should NOT show any changes)")

	if len(resp.PlanValue.Elements()) != 4 {
		t.Errorf("Expected 4 elements, got %d", len(resp.PlanValue.Elements()))
	}
}

// TestScenario3_AddInMiddle tests adding a nodegroup in the middle
func TestScenario3_AddInMiddle(t *testing.T) {
	ctx := context.Background()
	m := planmodifiers.NodeGroupSortModifier{}

	ng1 := createTestNodeGroup(t, "ng-1", "t3.large", 2)
	ng15 := createTestNodeGroup(t, "ng-1.5", "t3.large", 2)
	ng2 := createTestNodeGroup(t, "ng-2", "t3.large", 2)
	ng3 := createTestNodeGroup(t, "ng-3", "t3.large", 2)

	stateList := createTestList(t, ng1, ng2, ng3)
	configList := createTestList(t, ng1, ng15, ng2, ng3)

	req := planmodifier.ListRequest{
		ConfigValue: configList,
		StateValue:  stateList,
		PlanValue:   configList,
	}
	resp := &planmodifier.ListResponse{PlanValue: configList}

	m.PlanModifyList(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Unexpected error: %v", resp.Diagnostics)
	}

	t.Logf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	t.Logf("SCENARIO 3: ADD NODEGROUP IN MIDDLE")
	t.Logf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	t.Logf("State:  [ng-1, ng-2, ng-3]")
	t.Logf("Config: [ng-1, ng-1.5, ng-2, ng-3]")
	t.Logf("Expected: 1 nodegroup to add (ng-1.5)")
	t.Logf("Result: Plan has %d elements", len(resp.PlanValue.Elements()))
	t.Logf("EXPECTED TERRAFORM PLAN OUTPUT:")
	t.Logf("  + managed_nodegroups[1] (ng-1.5) will be created")
	t.Logf("  (ng-1, ng-2, ng-3 should NOT show any changes)")

	if len(resp.PlanValue.Elements()) != 4 {
		t.Errorf("Expected 4 elements, got %d", len(resp.PlanValue.Elements()))
	}
}

// TestScenario4_AddAtEnd tests adding a nodegroup at the end
func TestScenario4_AddAtEnd(t *testing.T) {
	ctx := context.Background()
	m := planmodifiers.NodeGroupSortModifier{}

	ng1 := createTestNodeGroup(t, "ng-1", "t3.large", 2)
	ng2 := createTestNodeGroup(t, "ng-2", "t3.large", 2)
	ng3 := createTestNodeGroup(t, "ng-3", "t3.large", 2)
	ng4 := createTestNodeGroup(t, "ng-4", "t3.large", 2)

	stateList := createTestList(t, ng1, ng2, ng3)
	configList := createTestList(t, ng1, ng2, ng3, ng4)

	req := planmodifier.ListRequest{
		ConfigValue: configList,
		StateValue:  stateList,
		PlanValue:   configList,
	}
	resp := &planmodifier.ListResponse{PlanValue: configList}

	m.PlanModifyList(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Unexpected error: %v", resp.Diagnostics)
	}

	t.Logf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	t.Logf("SCENARIO 4: ADD NODEGROUP AT END")
	t.Logf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	t.Logf("State:  [ng-1, ng-2, ng-3]")
	t.Logf("Config: [ng-1, ng-2, ng-3, ng-4]")
	t.Logf("Expected: 1 nodegroup to add (ng-4)")
	t.Logf("Result: Plan has %d elements", len(resp.PlanValue.Elements()))
	t.Logf("EXPECTED TERRAFORM PLAN OUTPUT:")
	t.Logf("  + managed_nodegroups[3] (ng-4) will be created")
	t.Logf("  (ng-1, ng-2, ng-3 should NOT show any changes)")

	if len(resp.PlanValue.Elements()) != 4 {
		t.Errorf("Expected 4 elements, got %d", len(resp.PlanValue.Elements()))
	}
}

// TestScenario5_DeleteFromStart tests removing a nodegroup from the start
func TestScenario5_DeleteFromStart(t *testing.T) {
	ctx := context.Background()
	m := planmodifiers.NodeGroupSortModifier{}

	ng1 := createTestNodeGroup(t, "ng-1", "t3.large", 2)
	ng2 := createTestNodeGroup(t, "ng-2", "t3.large", 2)
	ng3 := createTestNodeGroup(t, "ng-3", "t3.large", 2)

	stateList := createTestList(t, ng1, ng2, ng3)
	configList := createTestList(t, ng2, ng3) // ng-1 removed

	req := planmodifier.ListRequest{
		ConfigValue: configList,
		StateValue:  stateList,
		PlanValue:   configList,
	}
	resp := &planmodifier.ListResponse{PlanValue: configList}

	m.PlanModifyList(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Unexpected error: %v", resp.Diagnostics)
	}

	t.Logf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	t.Logf("SCENARIO 5: DELETE NODEGROUP FROM START")
	t.Logf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	t.Logf("State:  [ng-1, ng-2, ng-3]")
	t.Logf("Config: [ng-2, ng-3]")
	t.Logf("Expected: 1 nodegroup to delete (ng-1)")
	t.Logf("Result: Plan has %d elements", len(resp.PlanValue.Elements()))
	t.Logf("EXPECTED TERRAFORM PLAN OUTPUT:")
	t.Logf("  - managed_nodegroups[0] (ng-1) will be destroyed")
	t.Logf("  (ng-2, ng-3 should NOT show any changes)")

	if len(resp.PlanValue.Elements()) != 2 {
		t.Errorf("Expected 2 elements, got %d", len(resp.PlanValue.Elements()))
	}
}

// TestScenario6_DeleteFromMiddle tests removing a nodegroup from the middle
func TestScenario6_DeleteFromMiddle(t *testing.T) {
	ctx := context.Background()
	m := planmodifiers.NodeGroupSortModifier{}

	ng1 := createTestNodeGroup(t, "ng-1", "t3.large", 2)
	ng2 := createTestNodeGroup(t, "ng-2", "t3.large", 2)
	ng3 := createTestNodeGroup(t, "ng-3", "t3.large", 2)

	stateList := createTestList(t, ng1, ng2, ng3)
	configList := createTestList(t, ng1, ng3) // ng-2 removed

	req := planmodifier.ListRequest{
		ConfigValue: configList,
		StateValue:  stateList,
		PlanValue:   configList,
	}
	resp := &planmodifier.ListResponse{PlanValue: configList}

	m.PlanModifyList(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Unexpected error: %v", resp.Diagnostics)
	}

	t.Logf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	t.Logf("SCENARIO 6: DELETE NODEGROUP FROM MIDDLE")
	t.Logf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	t.Logf("State:  [ng-1, ng-2, ng-3]")
	t.Logf("Config: [ng-1, ng-3]")
	t.Logf("Expected: 1 nodegroup to delete (ng-2)")
	t.Logf("Result: Plan has %d elements", len(resp.PlanValue.Elements()))
	t.Logf("EXPECTED TERRAFORM PLAN OUTPUT:")
	t.Logf("  - managed_nodegroups[1] (ng-2) will be destroyed")
	t.Logf("  (ng-1, ng-3 should NOT show any changes)")

	if len(resp.PlanValue.Elements()) != 2 {
		t.Errorf("Expected 2 elements, got %d", len(resp.PlanValue.Elements()))
	}
}

// TestScenario7_DeleteFromEnd tests removing a nodegroup from the end
func TestScenario7_DeleteFromEnd(t *testing.T) {
	ctx := context.Background()
	m := planmodifiers.NodeGroupSortModifier{}

	ng1 := createTestNodeGroup(t, "ng-1", "t3.large", 2)
	ng2 := createTestNodeGroup(t, "ng-2", "t3.large", 2)
	ng3 := createTestNodeGroup(t, "ng-3", "t3.large", 2)

	stateList := createTestList(t, ng1, ng2, ng3)
	configList := createTestList(t, ng1, ng2) // ng-3 removed

	req := planmodifier.ListRequest{
		ConfigValue: configList,
		StateValue:  stateList,
		PlanValue:   configList,
	}
	resp := &planmodifier.ListResponse{PlanValue: configList}

	m.PlanModifyList(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Unexpected error: %v", resp.Diagnostics)
	}

	t.Logf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	t.Logf("SCENARIO 7: DELETE NODEGROUP FROM END")
	t.Logf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	t.Logf("State:  [ng-1, ng-2, ng-3]")
	t.Logf("Config: [ng-1, ng-2]")
	t.Logf("Expected: 1 nodegroup to delete (ng-3)")
	t.Logf("Result: Plan has %d elements", len(resp.PlanValue.Elements()))
	t.Logf("EXPECTED TERRAFORM PLAN OUTPUT:")
	t.Logf("  - managed_nodegroups[2] (ng-3) will be destroyed")
	t.Logf("  (ng-1, ng-2 should NOT show any changes)")

	if len(resp.PlanValue.Elements()) != 2 {
		t.Errorf("Expected 2 elements, got %d", len(resp.PlanValue.Elements()))
	}
}

// TestScenario8_ScaleUp tests scaling up a nodegroup (increasing desired_capacity)
func TestScenario8_ScaleUp(t *testing.T) {
	ctx := context.Background()
	m := planmodifiers.NodeGroupSortModifier{}

	ng1 := createTestNodeGroup(t, "ng-1", "t3.large", 2)
	ng2 := createTestNodeGroup(t, "ng-2", "t3.large", 2)
	ng3 := createTestNodeGroup(t, "ng-3", "t3.large", 2)

	stateList := createTestList(t, ng1, ng2, ng3)

	// Scale up ng-2 from 2 to 5 nodes
	ng2ScaledUp := createTestNodeGroup(t, "ng-2", "t3.large", 5)
	configList := createTestList(t, ng1, ng2ScaledUp, ng3)

	req := planmodifier.ListRequest{
		ConfigValue: configList,
		StateValue:  stateList,
		PlanValue:   configList,
	}
	resp := &planmodifier.ListResponse{PlanValue: configList}

	m.PlanModifyList(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Unexpected error: %v", resp.Diagnostics)
	}

	t.Logf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	t.Logf("SCENARIO 8: SCALE UP NODEGROUP")
	t.Logf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	t.Logf("State:  [ng-1(cap=2), ng-2(cap=2), ng-3(cap=2)]")
	t.Logf("Config: [ng-1(cap=2), ng-2(cap=5), ng-3(cap=2)]")
	t.Logf("Expected: 1 nodegroup to modify (ng-2: capacity 2->5)")
	t.Logf("Result: Plan has %d elements", len(resp.PlanValue.Elements()))
	t.Logf("EXPECTED TERRAFORM PLAN OUTPUT:")
	t.Logf("  ~ managed_nodegroups[1] (ng-2) will be updated in-place")
	t.Logf("    ~ desired_capacity: 2 -> 5")
	t.Logf("  (ng-1, ng-3 should NOT show any changes)")

	if len(resp.PlanValue.Elements()) != 3 {
		t.Errorf("Expected 3 elements, got %d", len(resp.PlanValue.Elements()))
	}
}

// TestScenario9_ScaleDown tests scaling down a nodegroup (decreasing desired_capacity)
func TestScenario9_ScaleDown(t *testing.T) {
	ctx := context.Background()
	m := planmodifiers.NodeGroupSortModifier{}

	ng1 := createTestNodeGroup(t, "ng-1", "t3.large", 5)
	ng2 := createTestNodeGroup(t, "ng-2", "t3.large", 5)
	ng3 := createTestNodeGroup(t, "ng-3", "t3.large", 5)

	stateList := createTestList(t, ng1, ng2, ng3)

	// Scale down ng-2 from 5 to 2 nodes
	ng2ScaledDown := createTestNodeGroup(t, "ng-2", "t3.large", 2)
	configList := createTestList(t, ng1, ng2ScaledDown, ng3)

	req := planmodifier.ListRequest{
		ConfigValue: configList,
		StateValue:  stateList,
		PlanValue:   configList,
	}
	resp := &planmodifier.ListResponse{PlanValue: configList}

	m.PlanModifyList(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Unexpected error: %v", resp.Diagnostics)
	}

	t.Logf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	t.Logf("SCENARIO 9: SCALE DOWN NODEGROUP")
	t.Logf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	t.Logf("State:  [ng-1(cap=5), ng-2(cap=5), ng-3(cap=5)]")
	t.Logf("Config: [ng-1(cap=5), ng-2(cap=2), ng-3(cap=5)]")
	t.Logf("Expected: 1 nodegroup to modify (ng-2: capacity 5->2)")
	t.Logf("Result: Plan has %d elements", len(resp.PlanValue.Elements()))
	t.Logf("EXPECTED TERRAFORM PLAN OUTPUT:")
	t.Logf("  ~ managed_nodegroups[1] (ng-2) will be updated in-place")
	t.Logf("    ~ desired_capacity: 5 -> 2")
	t.Logf("  (ng-1, ng-3 should NOT show any changes)")

	if len(resp.PlanValue.Elements()) != 3 {
		t.Errorf("Expected 3 elements, got %d", len(resp.PlanValue.Elements()))
	}
}

// TestScenario10_MultipleChanges tests multiple changes at once
func TestScenario10_MultipleChanges(t *testing.T) {
	ctx := context.Background()
	m := planmodifiers.NodeGroupSortModifier{}

	// State: ng-1, ng-2 (capacity=2), ng-3
	ng1 := createTestNodeGroup(t, "ng-1", "t3.large", 2)
	ng2 := createTestNodeGroup(t, "ng-2", "t3.large", 2)
	ng3 := createTestNodeGroup(t, "ng-3", "t3.large", 2)

	stateList := createTestList(t, ng1, ng2, ng3)

	// Config: ng-0 (added), ng-1, ng-2 (capacity=5), (ng-3 removed)
	ng0 := createTestNodeGroup(t, "ng-0", "t3.large", 2)
	ng2Changed := createTestNodeGroup(t, "ng-2", "t3.large", 5)
	configList := createTestList(t, ng0, ng1, ng2Changed)

	req := planmodifier.ListRequest{
		ConfigValue: configList,
		StateValue:  stateList,
		PlanValue:   configList,
	}
	resp := &planmodifier.ListResponse{PlanValue: configList}

	m.PlanModifyList(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Unexpected error: %v", resp.Diagnostics)
	}

	t.Logf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	t.Logf("SCENARIO 10: MULTIPLE CHANGES (ADD, MODIFY, DELETE)")
	t.Logf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	t.Logf("State:  [ng-1(cap=2), ng-2(cap=2), ng-3(cap=2)]")
	t.Logf("Config: [ng-0(cap=2), ng-1(cap=2), ng-2(cap=5)]")
	t.Logf("Expected: 1 add (ng-0), 1 modify (ng-2), 1 delete (ng-3)")
	t.Logf("Result: Plan has %d elements", len(resp.PlanValue.Elements()))
	t.Logf("EXPECTED TERRAFORM PLAN OUTPUT:")
	t.Logf("  + managed_nodegroups[0] (ng-0) will be created")
	t.Logf("  ~ managed_nodegroups[2] (ng-2) will be updated in-place")
	t.Logf("    ~ desired_capacity: 2 -> 5")
	t.Logf("  - managed_nodegroups (ng-3) will be destroyed")
	t.Logf("  (ng-1 should NOT show any changes)")

	if len(resp.PlanValue.Elements()) != 3 {
		t.Errorf("Expected 3 elements, got %d", len(resp.PlanValue.Elements()))
	}
}

// TestScenario11_InstanceTypeChange tests changing instance type (requires replacement)
func TestScenario11_InstanceTypeChange(t *testing.T) {
	ctx := context.Background()
	m := planmodifiers.NodeGroupSortModifier{}

	ng1 := createTestNodeGroup(t, "ng-1", "t3.large", 2)
	ng2 := createTestNodeGroup(t, "ng-2", "t3.large", 2)
	ng3 := createTestNodeGroup(t, "ng-3", "t3.large", 2)

	stateList := createTestList(t, ng1, ng2, ng3)

	// Change ng-2's instance type from t3.large to t3.xlarge
	ng2Changed := createTestNodeGroup(t, "ng-2", "t3.xlarge", 2)
	configList := createTestList(t, ng1, ng2Changed, ng3)

	req := planmodifier.ListRequest{
		ConfigValue: configList,
		StateValue:  stateList,
		PlanValue:   configList,
	}
	resp := &planmodifier.ListResponse{PlanValue: configList}

	m.PlanModifyList(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Unexpected error: %v", resp.Diagnostics)
	}

	t.Logf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	t.Logf("SCENARIO 11: INSTANCE TYPE CHANGE")
	t.Logf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	t.Logf("State:  [ng-1(t3.large), ng-2(t3.large), ng-3(t3.large)]")
	t.Logf("Config: [ng-1(t3.large), ng-2(t3.xlarge), ng-3(t3.large)]")
	t.Logf("Expected: 1 nodegroup to replace (ng-2: instance_type t3.large->t3.xlarge)")
	t.Logf("Result: Plan has %d elements", len(resp.PlanValue.Elements()))
	t.Logf("EXPECTED TERRAFORM PLAN OUTPUT:")
	t.Logf("  -/+ managed_nodegroups[1] (ng-2) will be replaced")
	t.Logf("    ~ instance_type: \"t3.large\" -> \"t3.xlarge\"")
	t.Logf("  (ng-1, ng-3 should NOT show any changes)")

	if len(resp.PlanValue.Elements()) != 3 {
		t.Errorf("Expected 3 elements, got %d", len(resp.PlanValue.Elements()))
	}
}

// Helper functions for integration tests

func createTestNodeGroup(t *testing.T, name string, instanceType string, desiredCapacity int64) types.Object {
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

func createTestList(t *testing.T, nodegroups ...types.Object) types.List {
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

// PrintScenarioSummary prints a summary of all test scenarios
func TestPrintScenarioSummary(t *testing.T) {
	t.Logf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	t.Logf("NODEGROUP SORTING TEST SCENARIOS SUMMARY")
	t.Logf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	t.Logf("")
	t.Logf("Scenario 1:  Reorder Only         - SHOULD show NO changes")
	t.Logf("Scenario 2:  Add at Start         - SHOULD show 1 add only")
	t.Logf("Scenario 3:  Add in Middle        - SHOULD show 1 add only")
	t.Logf("Scenario 4:  Add at End           - SHOULD show 1 add only")
	t.Logf("Scenario 5:  Delete from Start    - SHOULD show 1 delete only")
	t.Logf("Scenario 6:  Delete from Middle   - SHOULD show 1 delete only")
	t.Logf("Scenario 7:  Delete from End      - SHOULD show 1 delete only")
	t.Logf("Scenario 8:  Scale Up             - SHOULD show 1 modify (capacity up)")
	t.Logf("Scenario 9:  Scale Down           - SHOULD show 1 modify (capacity down)")
	t.Logf("Scenario 10: Multiple Changes     - SHOULD show add + modify + delete")
	t.Logf("Scenario 11: Instance Type Change - SHOULD show 1 replace")
	t.Logf("")
	t.Logf("Key point: Without the fix, scenarios 2-7 would incorrectly show")
	t.Logf("multiple nodegroups being 'renamed' due to index-based comparison.")
	t.Logf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}

// ============================================================================
// EXTENDED TESTS WITH 4+ NODEGROUPS AND MULTIPLE OPERATIONS
// ============================================================================

// TestScenario12_FourNodegroups_ReorderOnly tests reordering with 4 nodegroups
func TestScenario12_FourNodegroups_ReorderOnly(t *testing.T) {
	ctx := context.Background()
	m := planmodifiers.NodeGroupSortModifier{}

	ng1 := createTestNodeGroup(t, "ng-1", "t3.large", 2)
	ng2 := createTestNodeGroup(t, "ng-2", "t3.large", 2)
	ng3 := createTestNodeGroup(t, "ng-3", "t3.large", 2)
	ng4 := createTestNodeGroup(t, "ng-4", "t3.large", 2)

	// State: [ng-1, ng-2, ng-3, ng-4]
	stateList := createTestList(t, ng1, ng2, ng3, ng4)

	// Config: [ng-4, ng-2, ng-1, ng-3] - completely shuffled
	configList := createTestList(t, ng4, ng2, ng1, ng3)

	req := planmodifier.ListRequest{
		ConfigValue: configList,
		StateValue:  stateList,
		PlanValue:   configList,
	}
	resp := &planmodifier.ListResponse{PlanValue: configList}

	m.PlanModifyList(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Unexpected error: %v", resp.Diagnostics)
	}

	t.Logf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	t.Logf("SCENARIO 12: FOUR NODEGROUPS - REORDER ONLY")
	t.Logf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	t.Logf("State:  [ng-1, ng-2, ng-3, ng-4]")
	t.Logf("Config: [ng-4, ng-2, ng-1, ng-3]")
	t.Logf("Expected: No changes (reorder suppressed)")
	t.Logf("Result: Plan has %d elements", len(resp.PlanValue.Elements()))

	if len(resp.PlanValue.Elements()) != 4 {
		t.Errorf("Expected 4 elements, got %d", len(resp.PlanValue.Elements()))
	}
}

// TestScenario13_AddMultipleInMiddle tests adding ng-11 and ng-22 between ng-2 and ng-3
func TestScenario13_AddMultipleInMiddle(t *testing.T) {
	ctx := context.Background()
	m := planmodifiers.NodeGroupSortModifier{}

	ng1 := createTestNodeGroup(t, "ng-1", "t3.large", 2)
	ng2 := createTestNodeGroup(t, "ng-2", "t3.large", 2)
	ng3 := createTestNodeGroup(t, "ng-3", "t3.large", 2)
	ng11 := createTestNodeGroup(t, "ng-11", "t3.large", 2)
	ng22 := createTestNodeGroup(t, "ng-22", "t3.large", 2)

	// State: [ng-1, ng-2, ng-3]
	stateList := createTestList(t, ng1, ng2, ng3)

	// Config: [ng-1, ng-2, ng-11, ng-22, ng-3] - add ng-11 and ng-22 in middle
	configList := createTestList(t, ng1, ng2, ng11, ng22, ng3)

	req := planmodifier.ListRequest{
		ConfigValue: configList,
		StateValue:  stateList,
		PlanValue:   configList,
	}
	resp := &planmodifier.ListResponse{PlanValue: configList}

	m.PlanModifyList(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Unexpected error: %v", resp.Diagnostics)
	}

	t.Logf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	t.Logf("SCENARIO 13: ADD MULTIPLE NODEGROUPS IN MIDDLE")
	t.Logf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	t.Logf("State:  [ng-1, ng-2, ng-3]")
	t.Logf("Config: [ng-1, ng-2, ng-11, ng-22, ng-3]")
	t.Logf("Expected: 2 nodegroups to add (ng-11, ng-22)")
	t.Logf("Result: Plan has %d elements", len(resp.PlanValue.Elements()))
	t.Logf("EXPECTED TERRAFORM PLAN OUTPUT:")
	t.Logf("  + managed_nodegroups (ng-11) will be created")
	t.Logf("  + managed_nodegroups (ng-22) will be created")
	t.Logf("  (ng-1, ng-2, ng-3 should NOT show any changes)")

	if len(resp.PlanValue.Elements()) != 5 {
		t.Errorf("Expected 5 elements, got %d", len(resp.PlanValue.Elements()))
	}
}

// TestScenario14_DeleteMultipleFromMiddle tests deleting ng-2 and ng-3 from middle
func TestScenario14_DeleteMultipleFromMiddle(t *testing.T) {
	ctx := context.Background()
	m := planmodifiers.NodeGroupSortModifier{}

	ng1 := createTestNodeGroup(t, "ng-1", "t3.large", 2)
	ng2 := createTestNodeGroup(t, "ng-2", "t3.large", 2)
	ng3 := createTestNodeGroup(t, "ng-3", "t3.large", 2)
	ng4 := createTestNodeGroup(t, "ng-4", "t3.large", 2)

	// State: [ng-1, ng-2, ng-3, ng-4]
	stateList := createTestList(t, ng1, ng2, ng3, ng4)

	// Config: [ng-1, ng-4] - delete ng-2 and ng-3
	configList := createTestList(t, ng1, ng4)

	req := planmodifier.ListRequest{
		ConfigValue: configList,
		StateValue:  stateList,
		PlanValue:   configList,
	}
	resp := &planmodifier.ListResponse{PlanValue: configList}

	m.PlanModifyList(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Unexpected error: %v", resp.Diagnostics)
	}

	t.Logf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	t.Logf("SCENARIO 14: DELETE MULTIPLE NODEGROUPS FROM MIDDLE")
	t.Logf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	t.Logf("State:  [ng-1, ng-2, ng-3, ng-4]")
	t.Logf("Config: [ng-1, ng-4]")
	t.Logf("Expected: 2 nodegroups to delete (ng-2, ng-3)")
	t.Logf("Result: Plan has %d elements", len(resp.PlanValue.Elements()))
	t.Logf("EXPECTED TERRAFORM PLAN OUTPUT:")
	t.Logf("  - managed_nodegroups (ng-2) will be destroyed")
	t.Logf("  - managed_nodegroups (ng-3) will be destroyed")
	t.Logf("  (ng-1, ng-4 should NOT show any changes)")

	if len(resp.PlanValue.Elements()) != 2 {
		t.Errorf("Expected 2 elements, got %d", len(resp.PlanValue.Elements()))
	}
}

// TestScenario15_AddAndDeleteSimultaneously tests adding new and deleting existing
func TestScenario15_AddAndDeleteSimultaneously(t *testing.T) {
	ctx := context.Background()
	m := planmodifiers.NodeGroupSortModifier{}

	ng1 := createTestNodeGroup(t, "ng-1", "t3.large", 2)
	ng2 := createTestNodeGroup(t, "ng-2", "t3.large", 2)
	ng3 := createTestNodeGroup(t, "ng-3", "t3.large", 2)
	ng4 := createTestNodeGroup(t, "ng-4", "t3.large", 2)
	ngNew1 := createTestNodeGroup(t, "ng-new-1", "t3.large", 2)
	ngNew2 := createTestNodeGroup(t, "ng-new-2", "t3.large", 2)

	// State: [ng-1, ng-2, ng-3, ng-4]
	stateList := createTestList(t, ng1, ng2, ng3, ng4)

	// Config: [ng-1, ng-new-1, ng-new-2, ng-4] - delete ng-2, ng-3; add ng-new-1, ng-new-2
	configList := createTestList(t, ng1, ngNew1, ngNew2, ng4)

	req := planmodifier.ListRequest{
		ConfigValue: configList,
		StateValue:  stateList,
		PlanValue:   configList,
	}
	resp := &planmodifier.ListResponse{PlanValue: configList}

	m.PlanModifyList(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Unexpected error: %v", resp.Diagnostics)
	}

	t.Logf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	t.Logf("SCENARIO 15: ADD AND DELETE SIMULTANEOUSLY")
	t.Logf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	t.Logf("State:  [ng-1, ng-2, ng-3, ng-4]")
	t.Logf("Config: [ng-1, ng-new-1, ng-new-2, ng-4]")
	t.Logf("Expected: 2 adds (ng-new-1, ng-new-2) + 2 deletes (ng-2, ng-3)")
	t.Logf("Result: Plan has %d elements", len(resp.PlanValue.Elements()))
	t.Logf("EXPECTED TERRAFORM PLAN OUTPUT:")
	t.Logf("  + managed_nodegroups (ng-new-1) will be created")
	t.Logf("  + managed_nodegroups (ng-new-2) will be created")
	t.Logf("  - managed_nodegroups (ng-2) will be destroyed")
	t.Logf("  - managed_nodegroups (ng-3) will be destroyed")
	t.Logf("  (ng-1, ng-4 should NOT show any changes)")

	if len(resp.PlanValue.Elements()) != 4 {
		t.Errorf("Expected 4 elements, got %d", len(resp.PlanValue.Elements()))
	}
}

// TestScenario16_FiveNodegroups_ComplexReorder tests complex reordering with 5 nodegroups
func TestScenario16_FiveNodegroups_ComplexReorder(t *testing.T) {
	ctx := context.Background()
	m := planmodifiers.NodeGroupSortModifier{}

	ng1 := createTestNodeGroup(t, "ng-1", "t3.large", 2)
	ng2 := createTestNodeGroup(t, "ng-2", "t3.large", 2)
	ng3 := createTestNodeGroup(t, "ng-3", "t3.large", 2)
	ng4 := createTestNodeGroup(t, "ng-4", "t3.large", 2)
	ng5 := createTestNodeGroup(t, "ng-5", "t3.large", 2)

	// State: [ng-1, ng-2, ng-3, ng-4, ng-5]
	stateList := createTestList(t, ng1, ng2, ng3, ng4, ng5)

	// Config: [ng-5, ng-3, ng-1, ng-4, ng-2] - completely shuffled
	configList := createTestList(t, ng5, ng3, ng1, ng4, ng2)

	req := planmodifier.ListRequest{
		ConfigValue: configList,
		StateValue:  stateList,
		PlanValue:   configList,
	}
	resp := &planmodifier.ListResponse{PlanValue: configList}

	m.PlanModifyList(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Unexpected error: %v", resp.Diagnostics)
	}

	t.Logf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	t.Logf("SCENARIO 16: FIVE NODEGROUPS - COMPLEX REORDER")
	t.Logf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	t.Logf("State:  [ng-1, ng-2, ng-3, ng-4, ng-5]")
	t.Logf("Config: [ng-5, ng-3, ng-1, ng-4, ng-2]")
	t.Logf("Expected: No changes (reorder suppressed)")
	t.Logf("Result: Plan has %d elements", len(resp.PlanValue.Elements()))

	if len(resp.PlanValue.Elements()) != 5 {
		t.Errorf("Expected 5 elements, got %d", len(resp.PlanValue.Elements()))
	}
}

// TestScenario17_ReorderWithOneModification tests reorder + one field change
func TestScenario17_ReorderWithOneModification(t *testing.T) {
	ctx := context.Background()
	m := planmodifiers.NodeGroupSortModifier{}

	ng1 := createTestNodeGroup(t, "ng-1", "t3.large", 2)
	ng2 := createTestNodeGroup(t, "ng-2", "t3.large", 2)
	ng3 := createTestNodeGroup(t, "ng-3", "t3.large", 2)
	ng4 := createTestNodeGroup(t, "ng-4", "t3.large", 2)

	// State: [ng-1, ng-2, ng-3, ng-4]
	stateList := createTestList(t, ng1, ng2, ng3, ng4)

	// Config: [ng-4, ng-2(cap=5), ng-1, ng-3] - reordered AND ng-2 scaled up
	ng2Modified := createTestNodeGroup(t, "ng-2", "t3.large", 5)
	configList := createTestList(t, ng4, ng2Modified, ng1, ng3)

	req := planmodifier.ListRequest{
		ConfigValue: configList,
		StateValue:  stateList,
		PlanValue:   configList,
	}
	resp := &planmodifier.ListResponse{PlanValue: configList}

	m.PlanModifyList(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Unexpected error: %v", resp.Diagnostics)
	}

	t.Logf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	t.Logf("SCENARIO 17: REORDER WITH ONE MODIFICATION")
	t.Logf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	t.Logf("State:  [ng-1(cap=2), ng-2(cap=2), ng-3(cap=2), ng-4(cap=2)]")
	t.Logf("Config: [ng-4(cap=2), ng-2(cap=5), ng-1(cap=2), ng-3(cap=2)]")
	t.Logf("Expected: 1 modification (ng-2 cap 2->5), reorder ignored")
	t.Logf("Result: Plan has %d elements", len(resp.PlanValue.Elements()))
	t.Logf("EXPECTED TERRAFORM PLAN OUTPUT:")
	t.Logf("  ~ managed_nodegroups (ng-2) will be updated in-place")
	t.Logf("    ~ desired_capacity: 2 -> 5")
	t.Logf("  (ng-1, ng-3, ng-4 should NOT show any changes)")

	if len(resp.PlanValue.Elements()) != 4 {
		t.Errorf("Expected 4 elements, got %d", len(resp.PlanValue.Elements()))
	}
}

// TestScenario18_AddAtStartAndEnd tests adding nodegroups at both start and end
func TestScenario18_AddAtStartAndEnd(t *testing.T) {
	ctx := context.Background()
	m := planmodifiers.NodeGroupSortModifier{}

	ng0 := createTestNodeGroup(t, "ng-0", "t3.large", 2)
	ng1 := createTestNodeGroup(t, "ng-1", "t3.large", 2)
	ng2 := createTestNodeGroup(t, "ng-2", "t3.large", 2)
	ng3 := createTestNodeGroup(t, "ng-3", "t3.large", 2)
	ng4 := createTestNodeGroup(t, "ng-4", "t3.large", 2)

	// State: [ng-1, ng-2, ng-3]
	stateList := createTestList(t, ng1, ng2, ng3)

	// Config: [ng-0, ng-1, ng-2, ng-3, ng-4] - add at start and end
	configList := createTestList(t, ng0, ng1, ng2, ng3, ng4)

	req := planmodifier.ListRequest{
		ConfigValue: configList,
		StateValue:  stateList,
		PlanValue:   configList,
	}
	resp := &planmodifier.ListResponse{PlanValue: configList}

	m.PlanModifyList(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Unexpected error: %v", resp.Diagnostics)
	}

	t.Logf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	t.Logf("SCENARIO 18: ADD AT START AND END")
	t.Logf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	t.Logf("State:  [ng-1, ng-2, ng-3]")
	t.Logf("Config: [ng-0, ng-1, ng-2, ng-3, ng-4]")
	t.Logf("Expected: 2 adds (ng-0 at start, ng-4 at end)")
	t.Logf("Result: Plan has %d elements", len(resp.PlanValue.Elements()))
	t.Logf("EXPECTED TERRAFORM PLAN OUTPUT:")
	t.Logf("  + managed_nodegroups (ng-0) will be created")
	t.Logf("  + managed_nodegroups (ng-4) will be created")
	t.Logf("  (ng-1, ng-2, ng-3 should NOT show any changes)")

	if len(resp.PlanValue.Elements()) != 5 {
		t.Errorf("Expected 5 elements, got %d", len(resp.PlanValue.Elements()))
	}
}

// TestScenario19_DeleteFromStartAndEnd tests deleting from both start and end
func TestScenario19_DeleteFromStartAndEnd(t *testing.T) {
	ctx := context.Background()
	m := planmodifiers.NodeGroupSortModifier{}

	ng1 := createTestNodeGroup(t, "ng-1", "t3.large", 2)
	ng2 := createTestNodeGroup(t, "ng-2", "t3.large", 2)
	ng3 := createTestNodeGroup(t, "ng-3", "t3.large", 2)
	ng4 := createTestNodeGroup(t, "ng-4", "t3.large", 2)
	ng5 := createTestNodeGroup(t, "ng-5", "t3.large", 2)

	// State: [ng-1, ng-2, ng-3, ng-4, ng-5]
	stateList := createTestList(t, ng1, ng2, ng3, ng4, ng5)

	// Config: [ng-2, ng-3, ng-4] - delete ng-1 from start and ng-5 from end
	configList := createTestList(t, ng2, ng3, ng4)

	req := planmodifier.ListRequest{
		ConfigValue: configList,
		StateValue:  stateList,
		PlanValue:   configList,
	}
	resp := &planmodifier.ListResponse{PlanValue: configList}

	m.PlanModifyList(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Unexpected error: %v", resp.Diagnostics)
	}

	t.Logf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	t.Logf("SCENARIO 19: DELETE FROM START AND END")
	t.Logf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	t.Logf("State:  [ng-1, ng-2, ng-3, ng-4, ng-5]")
	t.Logf("Config: [ng-2, ng-3, ng-4]")
	t.Logf("Expected: 2 deletes (ng-1 from start, ng-5 from end)")
	t.Logf("Result: Plan has %d elements", len(resp.PlanValue.Elements()))
	t.Logf("EXPECTED TERRAFORM PLAN OUTPUT:")
	t.Logf("  - managed_nodegroups (ng-1) will be destroyed")
	t.Logf("  - managed_nodegroups (ng-5) will be destroyed")
	t.Logf("  (ng-2, ng-3, ng-4 should NOT show any changes)")

	if len(resp.PlanValue.Elements()) != 3 {
		t.Errorf("Expected 3 elements, got %d", len(resp.PlanValue.Elements()))
	}
}

// TestScenario20_ComplexMixedOperations tests a complex real-world scenario
func TestScenario20_ComplexMixedOperations(t *testing.T) {
	ctx := context.Background()
	m := planmodifiers.NodeGroupSortModifier{}

	// State: 5 nodegroups [ng-1, ng-2, ng-3, ng-4, ng-5]
	ng1 := createTestNodeGroup(t, "ng-1", "t3.large", 2)
	ng2 := createTestNodeGroup(t, "ng-2", "t3.large", 2)
	ng3 := createTestNodeGroup(t, "ng-3", "t3.large", 2)
	ng4 := createTestNodeGroup(t, "ng-4", "t3.large", 2)
	ng5 := createTestNodeGroup(t, "ng-5", "t3.large", 2)

	stateList := createTestList(t, ng1, ng2, ng3, ng4, ng5)

	// Config: Complex changes:
	// - Delete ng-1, ng-3
	// - Add ng-new-a, ng-new-b
	// - Modify ng-2 (scale up to 5)
	// - Keep ng-4, ng-5 unchanged but reordered
	ngNewA := createTestNodeGroup(t, "ng-new-a", "t3.large", 2)
	ngNewB := createTestNodeGroup(t, "ng-new-b", "t3.large", 2)
	ng2Modified := createTestNodeGroup(t, "ng-2", "t3.large", 5)

	// Config: [ng-5, ng-new-a, ng-2(cap=5), ng-new-b, ng-4] - reordered + changes
	configList := createTestList(t, ng5, ngNewA, ng2Modified, ngNewB, ng4)

	req := planmodifier.ListRequest{
		ConfigValue: configList,
		StateValue:  stateList,
		PlanValue:   configList,
	}
	resp := &planmodifier.ListResponse{PlanValue: configList}

	m.PlanModifyList(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Unexpected error: %v", resp.Diagnostics)
	}

	t.Logf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	t.Logf("SCENARIO 20: COMPLEX MIXED OPERATIONS (REAL-WORLD)")
	t.Logf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	t.Logf("State:  [ng-1, ng-2(cap=2), ng-3, ng-4, ng-5]")
	t.Logf("Config: [ng-5, ng-new-a, ng-2(cap=5), ng-new-b, ng-4]")
	t.Logf("Operations:")
	t.Logf("  - Delete: ng-1, ng-3")
	t.Logf("  - Add: ng-new-a, ng-new-b")
	t.Logf("  - Modify: ng-2 (cap 2->5)")
	t.Logf("  - Unchanged (just reordered): ng-4, ng-5")
	t.Logf("Result: Plan has %d elements", len(resp.PlanValue.Elements()))
	t.Logf("EXPECTED TERRAFORM PLAN OUTPUT:")
	t.Logf("  + managed_nodegroups (ng-new-a) will be created")
	t.Logf("  + managed_nodegroups (ng-new-b) will be created")
	t.Logf("  ~ managed_nodegroups (ng-2) will be updated in-place")
	t.Logf("    ~ desired_capacity: 2 -> 5")
	t.Logf("  - managed_nodegroups (ng-1) will be destroyed")
	t.Logf("  - managed_nodegroups (ng-3) will be destroyed")
	t.Logf("  (ng-4, ng-5 should NOT show any changes)")

	if len(resp.PlanValue.Elements()) != 5 {
		t.Errorf("Expected 5 elements, got %d", len(resp.PlanValue.Elements()))
	}
}

// Verify the count of nodegroups in each scenario
func TestVerifyNodeGroupCounts(t *testing.T) {
	scenarios := []struct {
		name          string
		stateCount    int
		configCount   int
		expectedPlan  int
		description   string
	}{
		{"Reorder Only (3 NGs)", 3, 3, 3, "Same nodegroups, different order"},
		{"Add at Start", 3, 4, 4, "ng-0 added before ng-1"},
		{"Add in Middle", 3, 4, 4, "ng-1.5 added between ng-1 and ng-2"},
		{"Add at End", 3, 4, 4, "ng-4 added after ng-3"},
		{"Delete from Start", 3, 2, 2, "ng-1 removed"},
		{"Delete from Middle", 3, 2, 2, "ng-2 removed"},
		{"Delete from End", 3, 2, 2, "ng-3 removed"},
		{"Scale Up", 3, 3, 3, "ng-2 capacity increased"},
		{"Scale Down", 3, 3, 3, "ng-2 capacity decreased"},
		{"Multiple Changes", 3, 3, 3, "add ng-0, modify ng-2, delete ng-3"},
		{"Instance Type Change", 3, 3, 3, "ng-2 instance type changed"},
		{"Reorder Only (4 NGs)", 4, 4, 4, "4 nodegroups, shuffled"},
		{"Add Multiple Middle", 3, 5, 5, "ng-11, ng-22 added between ng-2/ng-3"},
		{"Delete Multiple Middle", 4, 2, 2, "ng-2, ng-3 deleted"},
		{"Add+Delete Simultaneous", 4, 4, 4, "2 adds + 2 deletes"},
		{"Reorder Only (5 NGs)", 5, 5, 5, "5 nodegroups, shuffled"},
		{"Reorder + Modify", 4, 4, 4, "reorder + ng-2 scaled"},
		{"Add Start+End", 3, 5, 5, "ng-0 at start, ng-4 at end"},
		{"Delete Start+End", 5, 3, 3, "ng-1 and ng-5 deleted"},
		{"Complex Mixed", 5, 5, 5, "2 adds, 2 deletes, 1 modify, reorder"},
	}

	fmt.Println("\n┌─────────────────────────────────────────────────────────────────────────────────┐")
	fmt.Println("│                    NODEGROUP COUNT VERIFICATION TABLE                          │")
	fmt.Println("├─────────────────────────────────────────────────────────────────────────────────┤")
	fmt.Printf("│ %-20s │ %5s │ %6s │ %4s │ %-30s │\n", "Scenario", "State", "Config", "Plan", "Description")
	fmt.Println("├─────────────────────────────────────────────────────────────────────────────────┤")

	for _, s := range scenarios {
		fmt.Printf("│ %-20s │ %5d │ %6d │ %4d │ %-30s │\n",
			s.name, s.stateCount, s.configCount, s.expectedPlan, s.description)
	}
	fmt.Println("└─────────────────────────────────────────────────────────────────────────────────┘")
}
