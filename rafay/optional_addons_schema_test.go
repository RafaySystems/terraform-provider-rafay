package rafay

import (
	"context"
	"slices"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// The selection is deployed as a set, so the same add-ons in another order,
// whether reordered in the config or returned so by the API, plan no change,
// while adding, removing or replacing one still does.
func TestOptionalAddonsSetDiff(t *testing.T) {
	r := &schema.Resource{Schema: map[string]*schema.Schema{
		// as on the AKS and imported cluster resources
		"optional_addons": {Type: schema.TypeSet, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}},
		// as in spec.blueprint of the GKE and AKS v3 resources, from the shared list schema
		"spec": {Type: schema.TypeList, Optional: true, Elem: &schema.Resource{Schema: map[string]*schema.Schema{
			"blueprint": optionalAddonsBlueprintSchema(&schema.Schema{Type: schema.TypeList, Optional: true, Elem: &schema.Resource{Schema: map[string]*schema.Schema{
				"name":            {Type: schema.TypeString, Optional: true},
				"optional_addons": {Type: schema.TypeList, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}},
			}}}),
		}}},
	}}
	raw := func(top, nested []interface{}) map[string]interface{} {
		return map[string]interface{}{
			"optional_addons": top,
			"spec":            []interface{}{map[string]interface{}{"blueprint": []interface{}{map[string]interface{}{"name": "bp", "optional_addons": nested}}}},
		}
	}
	// state as terraform stores it after applying [a b] at both levels
	d := schema.TestResourceDataRaw(t, r.Schema, raw([]interface{}{"a", "b"}, []interface{}{"a", "b"}))
	d.SetId("c1")
	state := d.State()

	diff := func(top, nested []interface{}) map[string]*terraform.ResourceAttrDiff {
		got, err := r.SimpleDiff(context.Background(), state, terraform.NewResourceConfigRaw(raw(top, nested)), nil)
		if err != nil {
			t.Fatal(err)
		}
		if got == nil {
			return nil
		}
		return got.Attributes
	}

	if got := diff([]interface{}{"b", "a"}, []interface{}{"b", "a"}); len(got) != 0 {
		t.Errorf("reordered add-ons planned a change: %v", got)
	}
	for name, c := range map[string][2][]interface{}{
		"top level add-on replaced": {{"a", "c"}, {"a", "b"}},
		"top level add-on removed":  {{"a"}, {"a", "b"}},
		"nested add-on replaced":    {{"a", "b"}, {"c", "a"}},
		"nested add-on added":       {{"a", "b"}, {"a", "b", "c"}},
	} {
		if got := diff(c[0], c[1]); len(got) == 0 {
			t.Errorf("%s: planned no change", name)
		}
	}
}

// Every optional_addons of every SDKv2 cluster resource and data source is a
// set, including the blueprint blocks copied from the shared list schema.
func TestOptionalAddonsAreSets(t *testing.T) {
	var find func(path string, s map[string]*schema.Schema, found map[string]*schema.Schema)
	find = func(path string, s map[string]*schema.Schema, found map[string]*schema.Schema) {
		for k, v := range s {
			if k == "optional_addons" {
				found[path+k] = v
			}
			if r, ok := v.Elem.(*schema.Resource); ok {
				find(path+k+".", r.Schema, found)
			}
		}
	}
	for name, s := range map[string]map[string]*schema.Schema{
		"rafay_aks_cluster":                resourceAKSCluster().Schema,
		"rafay_import_cluster":             resourceImportCluster().Schema,
		"rafay_import_cluster data source": dataImportCluster().Schema,
		"rafay_gke_cluster":                GKEClusterV3Schema(),
		"rafay_aks_cluster_v3":             aksV3ClusterSchema(),
	} {
		found := map[string]*schema.Schema{}
		find("", s, found)
		if len(found) == 0 {
			t.Errorf("%s: no optional_addons in the schema", name)
		}
		for path, v := range found {
			if v.Type != schema.TypeSet {
				t.Errorf("%s: %s is %v, want a set", name, path, v.Type)
			}
		}
	}
}

// An unset or empty selection must still reach the API as [] so it clears the
// selection, never as nil.
func TestOptionalAddonNames(t *testing.T) {
	for name, v := range map[string]interface{}{
		"unset": nil,
		"empty": schema.NewSet(schema.HashString, nil),
	} {
		if got := optionalAddonNames(v); got == nil || len(got) != 0 {
			t.Errorf("%s: got %#v, want an empty non-nil slice", name, got)
		}
	}
	got := optionalAddonNames(schema.NewSet(schema.HashString, []interface{}{"b", "a"}))
	slices.Sort(got)
	if !slices.Equal(got, []string{"a", "b"}) {
		t.Errorf("got %v, want [a b]", got)
	}
}
