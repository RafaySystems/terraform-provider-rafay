package resource_eks_cluster

import (
	"context"
	"strings"
	"testing"

	"github.com/RafaySystems/terraform-provider-rafay/rafay"
	"github.com/go-yaml/yaml"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// the cluster resource is declarative: an unset or empty optional_addons must
// still reach the API as an explicit [] so the selection is cleared
func TestSpecExpandAlwaysSendsOptionalAddons(t *testing.T) {
	cases := map[string]types.List{
		"null":  types.ListNull(types.StringType),
		"empty": types.ListValueMust(types.StringType, []attr.Value{}),
	}
	for name, oa := range cases {
		t.Run(name, func(t *testing.T) {
			v := SpecValue{OptionalAddons: oa, ProxyConfig: types.MapNull(types.StringType), state: attr.ValueStateKnown}
			spec, diags := v.Expand(context.Background())
			if diags.HasError() {
				t.Fatalf("expand: %v", diags)
			}
			out, err := yaml.Marshal(&rafay.EKSCluster{Spec: spec})
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(out), "optionalAddons: []\n") {
				t.Fatalf("want optionalAddons: [] in yaml, got:\n%s", out)
			}
		})
	}
}

func TestSpecFlattenKeepsEmptyOptionalAddons(t *testing.T) {
	empty := types.ListValueMust(types.StringType, []attr.Value{})
	cases := []struct {
		name  string
		in    []string
		state types.List
		want  types.List
	}{
		{"config [] api none", nil, empty, empty},
		{"config unset api none", nil, types.ListNull(types.StringType), types.ListNull(types.StringType)},
		{"config [] api has a", []string{"a"}, empty, types.ListValueMust(types.StringType, []attr.Value{types.StringValue("a")})},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var v SpecValue
			if d := v.Flatten(context.Background(), &rafay.EKSSpec{OptionalAddons: tc.in}, SpecValue{OptionalAddons: tc.state}); d.HasError() {
				t.Fatalf("flatten: %v", d)
			}
			if !v.OptionalAddons.Equal(tc.want) {
				t.Fatalf("got %v, want %v", v.OptionalAddons, tc.want)
			}
		})
	}
}
