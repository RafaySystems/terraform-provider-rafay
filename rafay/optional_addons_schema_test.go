package rafay

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// The selection is deployed as a set, so the same add-ons in another order,
// whether reordered in the config or returned so by the API, plan no change,
// while adding or removing one still does.
func TestSuppressOptionalAddonsOrder(t *testing.T) {
	addons := func() *schema.Schema {
		return &schema.Schema{Type: schema.TypeList, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}}
	}
	topLevel := addons()
	topLevel.DiffSuppressFunc = suppressOptionalAddonsOrder
	r := &schema.Resource{Schema: map[string]*schema.Schema{
		// as on the AKS and imported cluster resources
		"optional_addons": topLevel,
		// as in spec.blueprint of the GKE and AKS v3 resources
		"spec": {Type: schema.TypeList, Optional: true, Elem: &schema.Resource{Schema: map[string]*schema.Schema{
			"blueprint": optionalAddonsBlueprintSchema(&schema.Schema{Type: schema.TypeList, Optional: true, Elem: &schema.Resource{Schema: map[string]*schema.Schema{
				"name":            {Type: schema.TypeString, Optional: true},
				"optional_addons": addons(),
			}}}),
		}}},
	}}
	state := &terraform.InstanceState{ID: "c1", Attributes: map[string]string{
		"optional_addons.#": "2", "optional_addons.0": "a", "optional_addons.1": "b",
		"spec.#": "1", "spec.0.blueprint.#": "1", "spec.0.blueprint.0.name": "bp",
		"spec.0.blueprint.0.optional_addons.#": "2", "spec.0.blueprint.0.optional_addons.0": "a", "spec.0.blueprint.0.optional_addons.1": "b",
	}}
	config := func(top, nested []interface{}) *terraform.ResourceConfig {
		return terraform.NewResourceConfigRaw(map[string]interface{}{
			"optional_addons": top,
			"spec":            []interface{}{map[string]interface{}{"blueprint": []interface{}{map[string]interface{}{"name": "bp", "optional_addons": nested}}}},
		})
	}
	diff := func(c *terraform.ResourceConfig) map[string]*terraform.ResourceAttrDiff {
		d, err := r.SimpleDiff(context.Background(), state, c, nil)
		if err != nil {
			t.Fatal(err)
		}
		if d == nil {
			return nil
		}
		return d.Attributes
	}

	if got := diff(config([]interface{}{"b", "a"}, []interface{}{"b", "a"})); len(got) != 0 {
		t.Errorf("reordered add-ons planned a change: %v", got)
	}
	for name, c := range map[string]*terraform.ResourceConfig{
		"top level add-on replaced": config([]interface{}{"a", "c"}, []interface{}{"a", "b"}),
		"top level add-on removed":  config([]interface{}{"a"}, []interface{}{"a", "b"}),
		"nested add-on replaced":    config([]interface{}{"a", "b"}, []interface{}{"c", "a"}),
		"nested add-on added":       config([]interface{}{"a", "b"}, []interface{}{"a", "b", "c"}),
	} {
		if got := diff(c); len(got) == 0 {
			t.Errorf("%s: planned no change", name)
		}
	}
}

// Every optional_addons of every cluster resource ignores the order of the add-ons.
func TestOptionalAddonsOrderIgnored(t *testing.T) {
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
		"rafay_aks_cluster":    resourceAKSCluster().Schema,
		"rafay_import_cluster": resourceImportCluster().Schema,
		"rafay_gke_cluster":    GKEClusterV3Schema(),
		"rafay_aks_cluster_v3": aksV3ClusterSchema(),
	} {
		found := map[string]*schema.Schema{}
		find("", s, found)
		if len(found) == 0 {
			t.Errorf("%s: no optional_addons in the schema", name)
		}
		for path, v := range found {
			if v.DiffSuppressFunc == nil {
				t.Errorf("%s: %s does not ignore the order of the add-ons", name, path)
			}
		}
	}
}
