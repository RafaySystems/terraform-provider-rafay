package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"

	fw "github.com/RafaySystems/terraform-provider-rafay/internal/resource_mks_cluster"
)

// the blueprint block's attributes must match its CustomType, or reading the
// data source fails converting the object
func TestMksClusterDataSourceBlueprintMatchesCustomType(t *testing.T) {
	ctx := context.Background()
	spec := MksClusterDataSourceSchema(ctx).Attributes["spec"].(schema.SingleNestedAttribute)
	bp := spec.Attributes["blueprint"].(schema.SingleNestedAttribute)

	want := fw.BlueprintValue{}.AttributeTypes(ctx)
	if len(bp.Attributes) != len(want) {
		t.Errorf("schema has %d blueprint attributes, custom type has %d", len(bp.Attributes), len(want))
	}
	for name, typ := range want {
		a, ok := bp.Attributes[name]
		if !ok {
			t.Errorf("blueprint attribute %q missing from data source schema", name)
			continue
		}
		if !a.GetType().Equal(typ) {
			t.Errorf("blueprint attribute %q: schema type %s, custom type %s", name, a.GetType(), typ)
		}
	}
}
