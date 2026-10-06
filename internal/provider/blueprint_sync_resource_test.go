package provider

import (
	"context"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func blueprintSyncConfig(t *testing.T, optionalAddons tftypes.Value) tfsdk.Config {
	ctx := context.Background()
	var sr resource.SchemaResponse
	(&BlueprintSyncResource{}).Schema(ctx, resource.SchemaRequest{}, &sr)
	typ := sr.Schema.Type().TerraformType(ctx).(tftypes.Object)
	vals := map[string]tftypes.Value{}
	for name, at := range typ.AttributeTypes {
		vals[name] = tftypes.NewValue(at, nil)
	}
	vals["optional_addons"] = optionalAddons
	return tfsdk.Config{Schema: sr.Schema, Raw: tftypes.NewValue(typ, vals)}
}

// unset and explicit [] must stay distinguishable: [] deselects, unset keeps
func TestReadStringListFromConfigNullVsEmpty(t *testing.T) {
	list := tftypes.List{ElementType: tftypes.String}
	cases := map[string]struct {
		in   tftypes.Value
		want []string
	}{
		"null":  {tftypes.NewValue(list, nil), nil},
		"empty": {tftypes.NewValue(list, []tftypes.Value{}), []string{}},
		"set":   {tftypes.NewValue(list, []tftypes.Value{tftypes.NewValue(tftypes.String, "a")}), []string{"a"}},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got, diags := readStringListFromConfig(context.Background(), blueprintSyncConfig(t, tc.in), "optional_addons")
			if diags.HasError() {
				t.Fatal(diags)
			}
			if !reflect.DeepEqual(got, tc.want) || (got == nil) != (tc.want == nil) {
				t.Fatalf("got %#v, want %#v", got, tc.want)
			}
		})
	}
}

func TestBlueprintSyncPublishSelection(t *testing.T) {
	current := []string{"cur"}
	cases := []struct {
		name           string
		addons         []string
		optionalAddons []string
		wantSelective  bool
		wantSelection  []string
	}{
		{"nothing set keeps selection via full publish", nil, nil, false, current},
		{"selective addons sync keeps current selection", []string{"x"}, nil, true, current},
		{"explicit [] deselects", nil, []string{}, true, []string{}},
		{"explicit list selects", nil, []string{"a"}, true, []string{"a"}},
		{"addons and list", []string{"x"}, []string{"a"}, true, []string{"a"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			selective, selection := blueprintSyncPublishSelection(tc.addons, tc.optionalAddons, current)
			if selective != tc.wantSelective || !reflect.DeepEqual(selection, tc.wantSelection) {
				t.Fatalf("got (%v, %#v), want (%v, %#v)", selective, selection, tc.wantSelective, tc.wantSelection)
			}
		})
	}
}

// optional_addons must be kept in state (and refreshed by Read) for a plan to
// show a change in the selection; a write-only attribute never shows in a plan.
func TestBlueprintSyncOptionalAddonsInState(t *testing.T) {
	var sr resource.SchemaResponse
	(&BlueprintSyncResource{}).Schema(context.Background(), resource.SchemaRequest{}, &sr)
	attr, ok := sr.Schema.Attributes["optional_addons"].(schema.ListAttribute)
	if !ok {
		t.Fatalf("optional_addons is not a list attribute")
	}
	if attr.WriteOnly {
		t.Error("optional_addons is write-only, so terraform plan can never show a change to it")
	}
	if !attr.Optional || !attr.Computed {
		t.Error("optional_addons must be optional and computed: unset keeps the cluster's selection without a diff")
	}
}

func TestOptionalAddonsState(t *testing.T) {
	ctx := context.Background()
	empty := optionalAddonsState(nil)
	if empty.IsNull() || len(empty.Elements()) != 0 {
		t.Errorf("no selection on the cluster must be [] in state, got %v", empty)
	}

	observed := []string{"simple-addon"}
	var got []string
	appliedOptionalAddons(types.ListUnknown(types.StringType), observed).ElementsAs(ctx, &got, false)
	if !reflect.DeepEqual(got, observed) {
		t.Errorf("unknown plan (unset on create) must record the selection on the cluster, got %v", got)
	}

	planned := optionalAddonsState([]string{"nginx-yaml"})
	if applied := appliedOptionalAddons(planned, observed); !applied.Equal(planned) {
		t.Errorf("a known plan must be stored as planned, got %v", applied)
	}
}

// A plan with nothing changed must be empty; force_sync=true still asks for a
// re-sync on every apply.
func TestBlueprintSyncModifyPlanForcesOnlyWithForceSync(t *testing.T) {
	ctx := context.Background()
	var sr resource.SchemaResponse
	(&BlueprintSyncResource{}).Schema(ctx, resource.SchemaRequest{}, &sr)
	typ := sr.Schema.Type().TerraformType(ctx).(tftypes.Object)

	value := func(forceSync interface{}) tftypes.Value {
		vals := map[string]tftypes.Value{}
		for name, at := range typ.AttributeTypes {
			vals[name] = tftypes.NewValue(at, nil)
		}
		vals["id"] = tftypes.NewValue(tftypes.String, "import1/defaultproject")
		vals["cluster_name"] = tftypes.NewValue(tftypes.String, "import1")
		vals["project"] = tftypes.NewValue(tftypes.String, "defaultproject")
		vals["force_sync"] = tftypes.NewValue(tftypes.Bool, forceSync)
		return tftypes.NewValue(typ, vals)
	}

	for _, c := range []struct {
		forceSync interface{}
		forced    bool
	}{{nil, false}, {false, false}, {true, true}} {
		planned := value(nil)
		req := resource.ModifyPlanRequest{
			Config: tfsdk.Config{Schema: sr.Schema, Raw: value(c.forceSync)},
			Plan:   tfsdk.Plan{Schema: sr.Schema, Raw: planned},
			State:  tfsdk.State{Schema: sr.Schema, Raw: value(nil)},
		}
		resp := resource.ModifyPlanResponse{Plan: tfsdk.Plan{Schema: sr.Schema, Raw: planned}}
		(&BlueprintSyncResource{}).ModifyPlan(ctx, req, &resp)
		if resp.Diagnostics.HasError() {
			t.Fatal(resp.Diagnostics)
		}
		var id types.String
		resp.Plan.GetAttribute(ctx, path.Root("id"), &id)
		if id.IsUnknown() != c.forced {
			t.Errorf("force_sync=%v: id unknown (update forced) = %v, want %v", c.forceSync, id.IsUnknown(), c.forced)
		}
	}
}

type fakePrivate map[string][]byte

func (f fakePrivate) GetKey(_ context.Context, key string) ([]byte, diag.Diagnostics) {
	return f[key], nil
}

// Removing optional_addons from a config that set it deselects them; a config
// that never set it keeps the cluster's selection, as before.
func TestPlannedOptionalAddons(t *testing.T) {
	ctx := context.Background()
	fromState := optionalAddonsState([]string{"simple-addon"})
	configured := fakePrivate{optionalAddonsConfiguredKey: []byte("true")}
	neverSet := fakePrivate{}

	if got := plannedOptionalAddons(ctx, types.ListNull(types.StringType), configured, fromState); len(got.Elements()) != 0 || got.IsNull() {
		t.Errorf("removed from the config: want [], got %v", got)
	}
	if got := plannedOptionalAddons(ctx, types.ListNull(types.StringType), neverSet, fromState); !got.Equal(fromState) {
		t.Errorf("never set: want the cluster's selection kept, got %v", got)
	}
	set := optionalAddonsState([]string{"nginx-yaml"})
	if got := plannedOptionalAddons(ctx, set, configured, set); !got.Equal(set) {
		t.Errorf("set in the config: want it planned as is, got %v", got)
	}
}

// A set optional_addons whose "set" flag is not in private state yet (set with
// nothing else to change, or before the flag existed) needs one update to
// record it, or removing it later would keep the selection.
func TestRecordOptionalAddonsConfigured(t *testing.T) {
	ctx := context.Background()
	set := optionalAddonsState([]string{"nginx-yaml"})
	unset := types.ListNull(types.StringType)
	recorded := fakePrivate{optionalAddonsConfiguredKey: []byte("true")}
	recordedUnset := fakePrivate{optionalAddonsConfiguredKey: []byte("false")}

	for _, c := range []struct {
		name       string
		configured types.List
		private    privateState
		want       bool
	}{
		{"set, flag missing", set, fakePrivate{}, true},
		{"set, flag recorded as unset", set, recordedUnset, true},
		{"set, flag recorded", set, recorded, false},
		{"unset, flag missing", unset, fakePrivate{}, false},
	} {
		if got := mustRecordOptionalAddonsConfigured(ctx, c.configured, c.private); got != c.want {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
}

// The update that only records the flag must not publish a sync.
func TestBlueprintSyncNeeded(t *testing.T) {
	state := BlueprintSyncModel{
		BlueprintName:    types.StringValue("test1"),
		BlueprintVersion: types.StringValue("v1"),
		OptionalAddons:   optionalAddonsState([]string{"nginx-yaml"}),
	}
	same := state
	if blueprintSyncNeeded(same, state, false) {
		t.Error("nothing changed and no force_sync: no sync")
	}
	if !blueprintSyncNeeded(same, state, true) {
		t.Error("force_sync: sync")
	}
	changed := state
	changed.OptionalAddons = optionalAddonsState(nil)
	if !blueprintSyncNeeded(changed, state, false) {
		t.Error("selection changed: sync")
	}
	moved := state
	moved.BlueprintVersion = types.StringValue("v2")
	if !blueprintSyncNeeded(moved, state, false) {
		t.Error("blueprint version changed: sync")
	}
}
