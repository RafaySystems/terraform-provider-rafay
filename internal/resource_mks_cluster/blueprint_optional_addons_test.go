package resource_mks_cluster

import (
	"context"
	"testing"

	"github.com/RafaySystems/rafay-common/proto/types/hub/infrapb"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestBlueprintFromHubOptionalAddons(t *testing.T) {
	ctx := context.Background()
	empty := types.SetValueMust(types.StringType, []attr.Value{})
	tests := []struct {
		name     string
		prior    types.Set
		hub      []string
		wantNull bool
		wantLen  int
	}{
		{"empty set in config stays empty", empty, nil, false, 0},
		{"unset stays null", types.SetNull(types.StringType), nil, true, 0},
		{"api value wins", empty, []string{"a"}, false, 1},
	}
	// a set, so the API returning the selection in another order is no change
	prior := types.SetValueMust(types.StringType, []attr.Value{types.StringValue("a"), types.StringValue("b")})
	v := BlueprintValue{OptionalAddons: prior}
	obj, diags := v.FromHub(ctx, &infrapb.ClusterBlueprint{Name: "bp", Version: "v2", OptionalAddons: []string{"b", "a"}})
	if diags.HasError() {
		t.Fatal(diags)
	}
	if got := obj.Attributes()["optional_addons"].(types.Set); !got.Equal(prior) {
		t.Errorf("api returns another order: got %v, want %v", got, prior)
	}
	for _, tt := range tests {
		v := BlueprintValue{OptionalAddons: tt.prior}
		obj, diags := v.FromHub(ctx, &infrapb.ClusterBlueprint{Name: "bp", Version: "v2", OptionalAddons: tt.hub})
		if diags.HasError() {
			t.Fatalf("%s: %v", tt.name, diags)
		}
		got := obj.Attributes()["optional_addons"].(types.Set)
		if got.IsNull() != tt.wantNull || len(got.Elements()) != tt.wantLen {
			t.Errorf("%s: got %v", tt.name, got)
		}
	}
}
