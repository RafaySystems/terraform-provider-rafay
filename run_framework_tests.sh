#!/bin/bash
# Comprehensive test runner for framework-based nodegroup sorting
#
# This script runs all tests including:
# - Unit tests for the PlanModifier
# - Integration tests with state/config scenarios
# - Race condition tests
# - Benchmark tests
#
# Usage: ./run_framework_tests.sh [--verbose] [--quick]

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Parse arguments
VERBOSE="-v"  # Always verbose to show test output
QUICK=""
while [[ $# -gt 0 ]]; do
    case $1 in
        --quick|-q)
            QUICK="true"
            shift
            ;;
        *)
            shift
            ;;
    esac
done

# Export required environment variable
export GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore

# Track test results
TOTAL_TESTS=0
TESTS_PASSED=0
TESTS_FAILED=0

# Function to run a test section
run_test_section() {
    local section_name="$1"
    local test_cmd="$2"

    echo ""
    echo -e "${BLUE}════════════════════════════════════════════════════════════════${NC}"
    echo -e "${BLUE}  $section_name${NC}"
    echo -e "${BLUE}════════════════════════════════════════════════════════════════${NC}"
    echo ""

    ((TOTAL_TESTS++))
    if eval "$test_cmd"; then
        echo -e "${GREEN}✓ PASSED: $section_name${NC}"
        ((TESTS_PASSED++))
        return 0
    else
        echo -e "${RED}✗ FAILED: $section_name${NC}"
        ((TESTS_FAILED++))
        return 1
    fi
}

echo ""
echo -e "${CYAN}╔══════════════════════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${CYAN}║                 TERRAFORM PROVIDER RAFAY - NODEGROUP SORTING FIX                 ║${NC}"
echo -e "${CYAN}║                            Provider Version: 4.4.4                               ║${NC}"
echo -e "${CYAN}╚══════════════════════════════════════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "${YELLOW}Problem Being Solved:${NC}"
echo "  When nodegroups are reordered in HCL, Terraform's index-based list comparison"
echo "  incorrectly shows all nodegroups as being 'renamed', causing unnecessary"
echo "  delete/recreate operations."
echo ""
echo -e "${YELLOW}Solution:${NC}"
echo "  ModifyPlan compares nodegroups by NAME instead of by list INDEX."
echo "  If only ordering changed (same nodegroups, different order), the diff is suppressed."
echo ""
echo "Test Suite Date: $(date)"
echo ""

# ============================================================================
# PART 1: BUILD AND COMPILE CHECKS
# ============================================================================

echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}PART 1: BUILD AND COMPILE CHECKS${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

run_test_section "Go Build - Full Provider" "go build ./..."

run_test_section "Go Vet - PlanModifiers" "go vet ./internal/planmodifiers/..."

run_test_section "Go Vet - Provider" "go vet ./internal/provider/..."

# ============================================================================
# PART 2: UNIT TESTS
# ============================================================================

echo ""
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}PART 2: UNIT TESTS (PlanModifier Logic)${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

run_test_section "All Unit Tests" \
    "go test $VERBOSE ./internal/planmodifiers/... -run 'Test[^S][^c]' -timeout 60s"

# ============================================================================
# PART 3: INTEGRATION TESTS - SCENARIOS
# ============================================================================

echo ""
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}PART 3: INTEGRATION TESTS (State/Config Scenarios)${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo "These tests simulate real terraform scenarios with state and config files."
echo "Each test shows expected terraform plan output for manual verification."
echo ""

run_test_section "Scenario 1: Reorder Only (CRITICAL FIX TEST)" \
    "go test $VERBOSE ./internal/planmodifiers/... -run 'TestScenario1_ReorderOnly' -timeout 30s"

run_test_section "Scenario 2: Add NodeGroup at Start" \
    "go test $VERBOSE ./internal/planmodifiers/... -run 'TestScenario2_AddAtStart' -timeout 30s"

run_test_section "Scenario 3: Add NodeGroup in Middle" \
    "go test $VERBOSE ./internal/planmodifiers/... -run 'TestScenario3_AddInMiddle' -timeout 30s"

run_test_section "Scenario 4: Add NodeGroup at End" \
    "go test $VERBOSE ./internal/planmodifiers/... -run 'TestScenario4_AddAtEnd' -timeout 30s"

run_test_section "Scenario 5: Delete NodeGroup from Start" \
    "go test $VERBOSE ./internal/planmodifiers/... -run 'TestScenario5_DeleteFromStart' -timeout 30s"

run_test_section "Scenario 6: Delete NodeGroup from Middle" \
    "go test $VERBOSE ./internal/planmodifiers/... -run 'TestScenario6_DeleteFromMiddle' -timeout 30s"

run_test_section "Scenario 7: Delete NodeGroup from End" \
    "go test $VERBOSE ./internal/planmodifiers/... -run 'TestScenario7_DeleteFromEnd' -timeout 30s"

run_test_section "Scenario 8: Scale Up NodeGroup" \
    "go test $VERBOSE ./internal/planmodifiers/... -run 'TestScenario8_ScaleUp' -timeout 30s"

run_test_section "Scenario 9: Scale Down NodeGroup" \
    "go test $VERBOSE ./internal/planmodifiers/... -run 'TestScenario9_ScaleDown' -timeout 30s"

run_test_section "Scenario 10: Multiple Changes (Add + Modify + Delete)" \
    "go test $VERBOSE ./internal/planmodifiers/... -run 'TestScenario10_MultipleChanges' -timeout 30s"

run_test_section "Scenario 11: Instance Type Change" \
    "go test $VERBOSE ./internal/planmodifiers/... -run 'TestScenario11_InstanceTypeChange' -timeout 30s"

echo ""
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}PART 3b: EXTENDED TESTS (4+ Nodegroups, Multiple Operations)${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

run_test_section "Scenario 12: Four NodeGroups - Reorder Only" \
    "go test $VERBOSE ./internal/planmodifiers/... -run 'TestScenario12_FourNodegroups_ReorderOnly' -timeout 30s"

run_test_section "Scenario 13: Add Multiple NodeGroups in Middle (ng-11, ng-22)" \
    "go test $VERBOSE ./internal/planmodifiers/... -run 'TestScenario13_AddMultipleInMiddle' -timeout 30s"

run_test_section "Scenario 14: Delete Multiple NodeGroups from Middle" \
    "go test $VERBOSE ./internal/planmodifiers/... -run 'TestScenario14_DeleteMultipleFromMiddle' -timeout 30s"

run_test_section "Scenario 15: Add and Delete Simultaneously" \
    "go test $VERBOSE ./internal/planmodifiers/... -run 'TestScenario15_AddAndDeleteSimultaneously' -timeout 30s"

run_test_section "Scenario 16: Five NodeGroups - Complex Reorder" \
    "go test $VERBOSE ./internal/planmodifiers/... -run 'TestScenario16_FiveNodegroups_ComplexReorder' -timeout 30s"

run_test_section "Scenario 17: Reorder With One Modification" \
    "go test $VERBOSE ./internal/planmodifiers/... -run 'TestScenario17_ReorderWithOneModification' -timeout 30s"

run_test_section "Scenario 18: Add at Start and End" \
    "go test $VERBOSE ./internal/planmodifiers/... -run 'TestScenario18_AddAtStartAndEnd' -timeout 30s"

run_test_section "Scenario 19: Delete from Start and End" \
    "go test $VERBOSE ./internal/planmodifiers/... -run 'TestScenario19_DeleteFromStartAndEnd' -timeout 30s"

run_test_section "Scenario 20: Complex Mixed Operations (Real-World)" \
    "go test $VERBOSE ./internal/planmodifiers/... -run 'TestScenario20_ComplexMixedOperations' -timeout 30s"

# ============================================================================
# PART 4: SUMMARY AND VERIFICATION
# ============================================================================

echo ""
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}PART 4: TEST SCENARIO SUMMARY${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

run_test_section "Print Scenario Summary" \
    "go test $VERBOSE ./internal/planmodifiers/... -run 'TestPrintScenarioSummary' -timeout 30s"

run_test_section "Verify NodeGroup Counts" \
    "go test $VERBOSE ./internal/planmodifiers/... -run 'TestVerifyNodeGroupCounts' -timeout 30s"

# ============================================================================
# PART 5: RACE CONDITION TESTS
# ============================================================================

if [ "$QUICK" != "true" ]; then
    echo ""
    echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${YELLOW}PART 5: RACE CONDITION AND BENCHMARK TESTS${NC}"
    echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

    run_test_section "Race Condition Tests" \
        "go test -race ./internal/planmodifiers/... -timeout 120s"

    run_test_section "Benchmark Tests" \
        "go test -bench=. ./internal/planmodifiers/... -benchmem -run='^$' -timeout 60s"
fi

# ============================================================================
# FINAL SUMMARY
# ============================================================================

echo ""
echo -e "${CYAN}╔══════════════════════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${CYAN}║                              FINAL TEST SUMMARY                                  ║${NC}"
echo -e "${CYAN}╚══════════════════════════════════════════════════════════════════════════════════╝${NC}"
echo ""
echo "Total Test Sections: $TOTAL_TESTS"
echo -e "${GREEN}Passed: $TESTS_PASSED${NC}"
echo -e "${RED}Failed: $TESTS_FAILED${NC}"
echo ""

echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${CYAN}EXPECTED TERRAFORM BEHAVIOR AFTER FIX:${NC}"
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo "BEFORE FIX (index-based comparison):"
echo "  Reordering nodegroups [ng-1, ng-2, ng-3] to [ng-3, ng-1, ng-2] would show:"
echo "    ~ name = \"ng-1\" -> \"ng-3\"  (at index 0)"
echo "    ~ name = \"ng-2\" -> \"ng-1\"  (at index 1)"
echo "    ~ name = \"ng-3\" -> \"ng-2\"  (at index 2)"
echo "  This causes ALL nodegroups to be recreated!"
echo ""
echo "AFTER FIX (name-based comparison):"
echo "  Reordering nodegroups shows:"
echo "    No changes. Your infrastructure matches the configuration."
echo "  Only actual changes (add/delete/modify) are shown."
echo ""
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

if [ $TESTS_FAILED -eq 0 ]; then
    echo ""
    echo -e "${GREEN}╔══════════════════════════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║                           ALL TESTS PASSED!                                      ║${NC}"
    echo -e "${GREEN}║                                                                                  ║${NC}"
    echo -e "${GREEN}║   The nodegroup sorting fix is working correctly.                                ║${NC}"
    echo -e "${GREEN}║   Provider version 4.4.4 is ready for use.                                       ║${NC}"
    echo -e "${GREEN}╚══════════════════════════════════════════════════════════════════════════════════╝${NC}"
    exit 0
else
    echo ""
    echo -e "${RED}╔══════════════════════════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${RED}║                     SOME TESTS FAILED - SEE OUTPUT ABOVE                         ║${NC}"
    echo -e "${RED}╚══════════════════════════════════════════════════════════════════════════════════╝${NC}"
    exit 1
fi
