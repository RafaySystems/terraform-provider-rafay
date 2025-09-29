package resource_eks_cluster

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var _ basetypes.MapTypable = NullableEmptyMapType{}

type NullableEmptyMapType struct {
	basetypes.MapType
}

func (t NullableEmptyMapType) Equal(o attr.Type) bool {
	other, ok := o.(NullableEmptyMapType)
	if !ok {
		return false
	}
	return t.MapType.Equal(other.MapType)
}

func (t NullableEmptyMapType) String() string {
	return "NullableEmptyMapType"
}

func (t NullableEmptyMapType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.MapType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}

	mapValue, ok := attrValue.(basetypes.MapValue)
	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", attrValue)
	}

	mapValuable, diags := t.ValueFromMap(ctx, mapValue)
	if diags.HasError() {
		return nil, fmt.Errorf("unexpected error converting basetypes.MapValue to basetypes.MapValuable: %v", diags)
	}

	return mapValuable, nil
}

func (t NullableEmptyMapType) ValueFromMap(ctx context.Context, in basetypes.MapValue) (basetypes.MapValuable, diag.Diagnostics) {
	value := NullableEmptyMapValue{
		MapValue: in,
	}
	return value, nil
}

func (t NullableEmptyMapType) ValueType(ctx context.Context) attr.Value {
	return NullableEmptyMapValue{
		MapValue: basetypes.NewMapNull(t.ElemType),
	}
}

// value type

// Ensure the implementation satisfies the expected interfaces.
var _ basetypes.MapValuable = NullableEmptyMapValue{}

type NullableEmptyMapValue struct {
	basetypes.MapValue
}

func (v NullableEmptyMapValue) Equal(o attr.Value) bool {
	other, ok := o.(NullableEmptyMapValue)
	if !ok {
		return false
	}

	fmt.Println("Comparing values:", v, other)

	// Handle the core logic: empty map is equal to null.
	// If both values are null, they are equal.
	if v.IsNull() && other.IsNull() {
		return true
	}

	// If one is null and the other is empty, they are equal.
	if v.IsNull() && len(other.Elements()) == 0 {
		return true
	}

	if len(v.Elements()) == 0 && other.IsNull() {
		fmt.Println("nullable empty map equal: empty vs null")
		return true
	}

	// Otherwise, use the standard map equality.
	return v.MapValue.Equal(other.MapValue)
}

func (v NullableEmptyMapValue) Type(ctx context.Context) attr.Type {
	return NullableEmptyMapType{
		MapType: v.MapValue.Type(ctx).(basetypes.MapType),
	}
}

// semantic equality

var _ basetypes.MapValuableWithSemanticEquals = NullableEmptyMapValue{}

func (v NullableEmptyMapValue) MapSemanticEquals(ctx context.Context, newValuable basetypes.MapValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	newValue, ok := newValuable.(NullableEmptyMapValue)
	if !ok {
		diags.AddError(
			"Semantic Equality Check Error",
			"An unexpected value type was received while performing semantic equality checks. "+
				"Please report this to the provider developers.\n\n"+
				"Expected Value Type: "+fmt.Sprintf("%T", v)+"\n"+
				"Got Value Type: "+fmt.Sprintf("%T", newValuable),
		)
		return false, diags
	}

	tflog.Debug(ctx, "Performing semantic equality check", map[string]interface{}{"value": v, "new_value": newValue})

	// Handle the core logic: empty map is equal to null.
	// If both values are null, they are equal.
	if v.IsNull() && newValue.IsNull() {
		return true, nil
	}

	// If one is null and the other is empty, they are equal.
	if v.IsNull() && len(newValue.Elements()) == 0 {
		return true, nil
	}

	if len(v.Elements()) == 0 && newValue.IsNull() {
		return true, nil
	}

	// Otherwise, use the standard map semantic equality.
	return v.MapValue.Equal(newValue.MapValue), nil
}
