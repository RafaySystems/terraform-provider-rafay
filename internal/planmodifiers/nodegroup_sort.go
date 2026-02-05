// Package planmodifiers provides custom plan modifiers for terraform-plugin-framework resources.
package planmodifiers

import (
	"context"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// NodeGroupSortModifier is a plan modifier that sorts nodegroups by name.
// This ensures that reordering nodegroups in HCL doesn't cause spurious diffs
// by comparing nodegroups by name rather than by list index position.
type NodeGroupSortModifier struct{}

// Description returns a human-readable description of what this modifier does.
func (m NodeGroupSortModifier) Description(ctx context.Context) string {
	return "Sorts nodegroups by name to prevent ordering-based diffs. " +
		"When nodegroups are reordered in HCL but have the same content, " +
		"this modifier ensures no changes are detected."
}

// MarkdownDescription returns a markdown description for documentation.
func (m NodeGroupSortModifier) MarkdownDescription(ctx context.Context) string {
	return "Sorts nodegroups by name to prevent ordering-based diffs. " +
		"When nodegroups are reordered in HCL but have the same content, " +
		"this modifier ensures no changes are detected."
}

// PlanModifyList implements the plan modification logic for list attributes.
func (m NodeGroupSortModifier) PlanModifyList(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
	// If the configuration is null or unknown, don't modify
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	// If there's no state (new resource), don't need to sort for comparison
	if req.StateValue.IsNull() {
		return
	}

	tflog.Debug(ctx, "NodeGroupSortModifier: Starting plan modification")

	// Get the config and state values
	configElements := req.ConfigValue.Elements()
	stateElements := req.StateValue.Elements()

	if len(configElements) == 0 {
		return
	}

	// If element counts differ, there are real additions/deletions - don't modify
	if len(configElements) != len(stateElements) {
		tflog.Debug(ctx, "NodeGroupSortModifier: Element counts differ, not modifying plan", map[string]interface{}{
			"configCount": len(configElements),
			"stateCount":  len(stateElements),
		})
		return
	}

	// Build maps by nodegroup name for both config and state
	configByName := BuildNodeGroupMapFromList(ctx, configElements)
	stateByName := BuildNodeGroupMapFromList(ctx, stateElements)

	// Safety check: if we couldn't extract names for all elements, don't modify the plan
	// This can happen if names are unknown (e.g., for newly created nodegroups)
	if len(configByName) != len(configElements) {
		tflog.Debug(ctx, "NodeGroupSortModifier: Could not extract all config nodegroup names, not modifying plan", map[string]interface{}{
			"configElements": len(configElements),
			"configByName":   len(configByName),
		})
		return
	}
	if len(stateByName) != len(stateElements) {
		tflog.Debug(ctx, "NodeGroupSortModifier: Could not extract all state nodegroup names, not modifying plan", map[string]interface{}{
			"stateElements": len(stateElements),
			"stateByName":   len(stateByName),
		})
		return
	}

	tflog.Debug(ctx, "NodeGroupSortModifier: Built maps", map[string]interface{}{
		"configCount": len(configByName),
		"stateCount":  len(stateByName),
	})

	// Check if this is just a reorder (same names, same content)
	if IsReorderOnly(ctx, configByName, stateByName) {
		tflog.Debug(ctx, "NodeGroupSortModifier: Detected reorder-only change, preserving state order")

		// Sort the config elements to match the sorted order
		sortedConfig := SortNodeGroupList(ctx, configElements)
		sortedState := SortNodeGroupList(ctx, stateElements)

		// If sorted versions are equal, use state value to suppress diff
		if ListsEqualAfterSort(ctx, sortedConfig, sortedState, configByName, stateByName) {
			resp.PlanValue = req.StateValue
			tflog.Debug(ctx, "NodeGroupSortModifier: Suppressed reorder diff")
			return
		}
	}

	// Otherwise, sort the plan value for consistent ordering
	sortedPlan := SortNodeGroupList(ctx, configElements)
	if sortedPlan != nil {
		listValue, diags := types.ListValue(req.ConfigValue.ElementType(ctx), sortedPlan)
		resp.Diagnostics.Append(diags...)
		if !resp.Diagnostics.HasError() {
			resp.PlanValue = listValue
			tflog.Debug(ctx, "NodeGroupSortModifier: Applied sorted plan value")
		}
	}
}

// NodeGroupSortListModifier returns a new NodeGroupSortModifier for use in schema definitions.
func NodeGroupSortListModifier() planmodifier.List {
	return NodeGroupSortModifier{}
}

// BuildNodeGroupMapFromList extracts nodegroups into a map keyed by name.
func BuildNodeGroupMapFromList(ctx context.Context, elements []attr.Value) map[string]attr.Value {
	result := make(map[string]attr.Value)

	for _, elem := range elements {
		obj, ok := elem.(types.Object)
		if !ok {
			continue
		}

		attrs := obj.Attributes()
		nameAttr, exists := attrs["name"]
		if !exists {
			continue
		}

		nameStr, ok := nameAttr.(types.String)
		if !ok || nameStr.IsNull() || nameStr.IsUnknown() {
			continue
		}

		result[nameStr.ValueString()] = elem
	}

	return result
}

// IsReorderOnly checks if the only difference between config and state is ordering.
func IsReorderOnly(ctx context.Context, configByName, stateByName map[string]attr.Value) bool {
	// Different count means real changes
	if len(configByName) != len(stateByName) {
		return false
	}

	// Check all names exist in both
	for name := range configByName {
		if _, exists := stateByName[name]; !exists {
			return false
		}
	}

	return true
}

// SortNodeGroupList sorts nodegroups by name.
func SortNodeGroupList(ctx context.Context, elements []attr.Value) []attr.Value {
	if len(elements) == 0 {
		return elements
	}

	type namedElement struct {
		name string
		elem attr.Value
	}

	named := make([]namedElement, 0, len(elements))
	for _, elem := range elements {
		obj, ok := elem.(types.Object)
		if !ok {
			continue
		}

		attrs := obj.Attributes()
		nameAttr, exists := attrs["name"]
		if !exists {
			named = append(named, namedElement{name: "", elem: elem})
			continue
		}

		nameStr, ok := nameAttr.(types.String)
		if !ok || nameStr.IsNull() || nameStr.IsUnknown() {
			named = append(named, namedElement{name: "", elem: elem})
			continue
		}

		named = append(named, namedElement{name: nameStr.ValueString(), elem: elem})
	}

	sort.SliceStable(named, func(i, j int) bool {
		return named[i].name < named[j].name
	})

	result := make([]attr.Value, len(named))
	for i, n := range named {
		result[i] = n.elem
	}

	return result
}

// ListsEqualAfterSort checks if two sorted lists have the same nodegroups by name.
func ListsEqualAfterSort(ctx context.Context, sortedConfig, sortedState []attr.Value, configByName, stateByName map[string]attr.Value) bool {
	if len(sortedConfig) != len(sortedState) {
		return false
	}

	// Compare by name - if all names match and values are compatible, consider equal
	for name, configElem := range configByName {
		stateElem, exists := stateByName[name]
		if !exists {
			return false
		}

		// Compare the essential fields (ignoring computed/default fields)
		if !NodeGroupsEqualByEssentialFields(ctx, configElem, stateElem) {
			return false
		}
	}

	return true
}

// NodeGroupsEqualByEssentialFields compares two nodegroup objects by their essential fields.
// This allows us to ignore computed/default fields that might differ between config and state.
func NodeGroupsEqualByEssentialFields(ctx context.Context, configElem, stateElem attr.Value) bool {
	configObj, ok1 := configElem.(types.Object)
	stateObj, ok2 := stateElem.(types.Object)
	if !ok1 || !ok2 {
		return false
	}

	configAttrs := configObj.Attributes()
	stateAttrs := stateObj.Attributes()

	// List of essential fields to compare (user-specified fields)
	essentialFields := []string{
		"name",
		"ami_family",
		"instance_type",
		"desired_capacity",
		"min_size",
		"max_size",
		"volume_size",
		"volume_type",
		"labels",
		"tags",
		"taints",
		"ssh",
		"iam",
		"subnets",
		"availability_zones",
	}

	for _, field := range essentialFields {
		configVal, configExists := configAttrs[field]
		stateVal, stateExists := stateAttrs[field]

		// If config doesn't specify this field, skip comparison
		if !configExists {
			continue
		}

		// If config specifies it but state doesn't have it, they differ
		if !stateExists {
			// Check if config value is null (meaning not specified)
			if IsNullOrUnknown(configVal) {
				continue
			}
			return false
		}

		// Compare the values
		if !ValuesEqual(configVal, stateVal) {
			return false
		}
	}

	return true
}

// IsNullOrUnknown checks if an attr.Value is null or unknown.
func IsNullOrUnknown(v attr.Value) bool {
	return v.IsNull() || v.IsUnknown()
}

// ValuesEqual compares two attr.Values for equality.
func ValuesEqual(a, b attr.Value) bool {
	// Handle null/unknown
	if a.IsNull() && b.IsNull() {
		return true
	}
	if a.IsNull() || b.IsNull() {
		return false
	}
	if a.IsUnknown() || b.IsUnknown() {
		return true // Unknown values are considered compatible
	}

	// Use string comparison as a simple equality check
	return a.String() == b.String()
}
