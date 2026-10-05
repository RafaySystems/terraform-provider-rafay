package provider

import (
	"context"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
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
