package resource_eks_cluster

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func indexOfString(item string, list []string) int {
	for i, v := range list {
		if v == item {
			return i
		}
	}
	return -1
}

// flattenStringListWithStateOrder preserves the order from prior Terraform state when
// flattening string lists returned by the API in a different order (e.g. alphabetically).
func flattenStringListWithStateOrder(remote []string, stateList types.List) []string {
	if len(remote) == 0 {
		return nil
	}

	var localOrder []string
	if !stateList.IsNull() && !stateList.IsUnknown() {
		for _, elem := range stateList.Elements() {
			if s, ok := elem.(types.String); ok && !s.IsNull() {
				localOrder = append(localOrder, s.ValueString())
			}
		}
	}

	out := make([]string, 0, len(remote))
	remoteOnly := make([]string, 0)
	for _, elem := range remote {
		if indexOfString(elem, localOrder) < 0 {
			remoteOnly = append(remoteOnly, elem)
		}
	}
	for _, elem := range localOrder {
		if indexOfString(elem, remote) >= 0 {
			out = append(out, elem)
		} else if len(remoteOnly) > 0 {
			out = append(out, remoteOnly[0])
			remoteOnly = remoteOnly[1:]
		}
	}
	out = append(out, remoteOnly...)
	return out
}
