package rafay

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// optionalAddonNames returns the add-on names of an optional_addons set, never
// nil, so an unset or empty selection still goes to the API as [].
func optionalAddonNames(v interface{}) []string {
	var names []interface{}
	if s, ok := v.(*schema.Set); ok && s != nil {
		names = s.List()
	}
	return toArrayString(names)
}

// optionalAddonsBlueprintSchema returns a copy of a blueprint block whose
// optional_addons is a set, so their order is not a change: the selection is
// deployed as a set. It is a copy because the block comes from the generated
// resource.ClusterSchema, which every resource shares and which models the
// add-ons as a list.
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
	optionalAddons.Type = schema.TypeSet
	fields["optional_addons"] = &optionalAddons

	out := *blueprint
	out.Elem = &schema.Resource{Schema: fields}
	return &out
}
