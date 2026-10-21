package rafay

import (
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// suppressOptionalAddonsOrder suppresses the diff of an optional_addons list
// that holds the same add-ons in another order. The selection is deployed as a
// set, so neither a reordered config nor an API that returns the add-ons in
// another order is a change.
func suppressOptionalAddonsOrder(k, _, _ string, d *schema.ResourceData) bool {
	// k is the list ("...optional_addons") or one of its keys
	// ("...optional_addons.#", "...optional_addons.0")
	list := k
	if i := strings.LastIndex(k, "."); i >= 0 && strings.HasSuffix(k[:i], "optional_addons") {
		list = k[:i]
	}
	o, n := d.GetChange(list)
	return sameOptionalAddons(o, n)
}

// sameOptionalAddons reports whether two optional_addons values name the same
// add-ons, in any order.
func sameOptionalAddons(a, b interface{}) bool {
	sorted := func(v interface{}) []string {
		l, _ := v.([]interface{})
		out := toArrayString(l)
		slices.Sort(out)
		return out
	}
	return slices.Equal(sorted(a), sorted(b))
}

// optionalAddonsBlueprintSchema returns a copy of a blueprint block whose
// optional_addons ignores the order of the add-ons. It is a copy because the
// block comes from resource.ClusterSchema, which every resource shares.
func optionalAddonsBlueprintSchema(blueprint *schema.Schema) *schema.Schema {
	elem, ok := blueprint.Elem.(*schema.Resource)
	if !ok || elem.Schema["optional_addons"] == nil {
		return blueprint
	}
	fields := make(map[string]*schema.Schema, len(elem.Schema))
	for k, v := range elem.Schema {
		fields[k] = v
	}
	optionalAddons := *elem.Schema["optional_addons"]
	optionalAddons.DiffSuppressFunc = suppressOptionalAddonsOrder
	fields["optional_addons"] = &optionalAddons

	out := *blueprint
	out.Elem = &schema.Resource{Schema: fields}
	return &out
}
