package eksint64planmodifier

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// UseStateOrNullForUnknown returns a plan modifier that, when the planned value
// is unknown, copies the prior state value even if it is null. If state is also
// unknown (e.g. create), it sets a known null. This avoids "(known after apply)"
// / unknown-after-apply for computed_optional ints omitted from HCL.
func UseStateOrNullForUnknown() planmodifier.Int64 {
	return useStateOrNullForUnknown{}
}

type useStateOrNullForUnknown struct{}

func (m useStateOrNullForUnknown) Description(_ context.Context) string {
	return "If config omits this value, keep prior state (including null); never leave unknown."
}

func (m useStateOrNullForUnknown) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m useStateOrNullForUnknown) PlanModifyInt64(_ context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
	if !req.PlanValue.IsUnknown() {
		return
	}
	// Prefer known state (concrete value or null).
	if !req.StateValue.IsUnknown() {
		resp.PlanValue = req.StateValue
		return
	}
	// Create / no usable state → known null.
	resp.PlanValue = types.Int64Null()
}
