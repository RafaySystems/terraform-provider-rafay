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
	empty := types.ListValueMust(types.StringType, []attr.Value{})
	tests := []struct {
		name     string
		prior    types.List
		hub      []string
		wantNull bool
		wantLen  int
	}{
		{"empty list in config stays empty", empty, nil, false, 0},
		{"unset stays null", types.ListNull(types.StringType), nil, true, 0},
		{"api value wins", empty, []string{"a"}, false, 1},
	}
	for _, tt := range tests {
		v := BlueprintValue{OptionalAddons: tt.prior}
		obj, diags := v.FromHub(ctx, &infrapb.ClusterBlueprint{Name: "bp", Version: "v2", OptionalAddons: tt.hub})
		if diags.HasError() {
			t.Fatalf("%s: %v", tt.name, diags)
		}
		got := obj.Attributes()["optional_addons"].(types.List)
		if got.IsNull() != tt.wantNull || len(got.Elements()) != tt.wantLen {
			t.Errorf("%s: got %v", tt.name, got)
		}
	}
}
