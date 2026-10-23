package eksint64planmodifier

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// UseStateOrNullForUnknown returns a plan modifier that, when the planned value
// is unknown, copies the prior state value even if it is null. If state is also
// unknown (e.g. create), it sets a known null. This avoids "(known after apply)"
// / unknown-after-apply for computed_optional ints omitted from HCL.
//
// For list nodegroups, the value is copied from the state nodegroup with the
// same name. Copying the array slot would move a deleted nodegroup's
// volume_iops / volume_throughput onto the next nodegroup.
func UseStateOrNullForUnknown() planmodifier.Int64 {
	return useStateOrNullForUnknown{}
}

type useStateOrNullForUnknown struct{}

func (m useStateOrNullForUnknown) Description(_ context.Context) string {
	return "If config omits this value, keep prior state for the same nodegroup name (including null); never leave unknown."
}

func (m useStateOrNullForUnknown) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m useStateOrNullForUnknown) PlanModifyInt64(ctx context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
	if !req.PlanValue.IsUnknown() {
		return
	}

	// List nodegroups are paired by position. Look up state by name so a
	// delete does not keep the removed nodegroup's values on the next one.
	if v, ok := int64FromStateNodegroupByName(ctx, req); ok {
		resp.PlanValue = v
		return
	}

	if !req.StateValue.IsUnknown() {
		resp.PlanValue = req.StateValue
		return
	}
	resp.PlanValue = types.Int64Null()
}

// int64FromStateNodegroupByName handles managed_nodegroups and node_groups
// lists. ok is false for maps, where the existing slot value is already the
// right nodegroup, and for any case we cannot resolve - the caller then falls
// back to the prior slot behaviour rather than discarding the value.
func int64FromStateNodegroupByName(ctx context.Context, req planmodifier.Int64Request) (types.Int64, bool) {
	elementPath := req.Path.ParentPath()
	last, _ := elementPath.Steps().LastStep()
	if _, isList := last.(path.PathStepElementKeyInt); !isList {
		return types.Int64Null(), false
	}

	var configName types.String
	diags := req.Config.GetAttribute(ctx, elementPath.AtName("name"), &configName)
	if diags.HasError() || configName.IsNull() || configName.IsUnknown() {
		return types.Int64Null(), false
	}

	listPath := elementPath.ParentPath()

	var stateList types.List
	diags = req.State.GetAttribute(ctx, listPath, &stateList)
	if diags.HasError() || stateList.IsNull() || stateList.IsUnknown() {
		return types.Int64Null(), false
	}

	attrName := ""
	if step, _ := req.Path.Steps().LastStep(); step != nil {
		if nameStep, ok := step.(path.PathStepAttributeName); ok {
			attrName = string(nameStep)
		}
	}
	if attrName == "" {
		return types.Int64Null(), false
	}

	for i := 0; i < len(stateList.Elements()); i++ {
		var nameVal types.String
		diags = req.State.GetAttribute(ctx, listPath.AtListIndex(i).AtName("name"), &nameVal)
		if diags.HasError() || nameVal.IsNull() || nameVal.IsUnknown() ||
			!strings.EqualFold(nameVal.ValueString(), configName.ValueString()) {
			continue
		}

		var field types.Int64
		diags = req.State.GetAttribute(ctx, listPath.AtListIndex(i).AtName(attrName), &field)
		if diags.HasError() {
			// Could not read the field - fall back to the slot value rather
			// than planning a null and deleting it on the server.
			return types.Int64Null(), false
		}
		if field.IsUnknown() {
			return types.Int64Null(), true
		}
		return field, true
	}

	// This name is not in state yet (new nodegroup). Leave the field null.
	return types.Int64Null(), true
}
