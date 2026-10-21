package resource_eks_cluster

import (
	"context"
	"testing"

	"github.com/RafaySystems/terraform-provider-rafay/rafay"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func boolPtr(b bool) *bool { return &b }

// The server omits the key when the flag is not set. Leaving the attribute null
// in that case contradicted an explicit false in config, so Terraform planned a
// change on every run that applying could never satisfy.
func TestFlattenForceUpdateVersion(t *testing.T) {
	cases := []struct {
		name string
		in   *bool
		want types.Bool
	}{
		{name: "absent from the server reads as false", in: nil, want: types.BoolValue(false)},
		{name: "false stays false", in: boolPtr(false), want: types.BoolValue(false)},
		{name: "true stays true", in: boolPtr(true), want: types.BoolValue(true)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var v Metadata2Value
			diags := v.Flatten(context.Background(), &rafay.EKSClusterConfigMetadata{
				Name:               "c",
				Region:             "us-east-1",
				Version:            "1.34",
				ForceUpdateVersion: tc.in,
			})
			if diags.HasError() {
				t.Fatalf("unexpected diagnostics: %v", diags)
			}
			if v.ForceUpdateVersion.IsNull() {
				t.Fatal("force_update_version must never be left null, that is the perpetual-diff bug")
			}
			if !v.ForceUpdateVersion.Equal(tc.want) {
				t.Fatalf("got %v, want %v", v.ForceUpdateVersion, tc.want)
			}
		})
	}
}

// Expand stays true-only: a false is indistinguishable from unset on the wire, and
// sending it would add the key to the stored config for users who never opted in.
func TestExpandForceUpdateVersion(t *testing.T) {
	ctx := context.Background()
	cases := []struct {
		name    string
		value   types.Bool
		wantSet bool
	}{
		{name: "true is sent", value: types.BoolValue(true), wantSet: true},
		{name: "false is not sent", value: types.BoolValue(false)},
		{name: "null is not sent", value: types.BoolNull()},
		{name: "unknown is not sent", value: types.BoolUnknown()},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			v := NewMetadata2ValueMust(Metadata2Value{}.AttributeTypes(ctx), map[string]attr.Value{
				"annotations":          types.MapNull(types.StringType),
				"force_update_version": tc.value,
				"name":                 types.StringValue("c"),
				"region":               types.StringValue("us-east-1"),
				"tags":                 types.MapNull(types.StringType),
				"version":              types.StringValue("1.34"),
			})

			md, diags := v.Expand(ctx)
			if diags.HasError() {
				t.Fatalf("unexpected diagnostics: %v", diags)
			}
			if tc.wantSet {
				if md.ForceUpdateVersion == nil || !*md.ForceUpdateVersion {
					t.Fatalf("expected forceUpdateVersion true on the wire, got %v", md.ForceUpdateVersion)
				}
				return
			}
			if md.ForceUpdateVersion != nil {
				t.Fatalf("expected no forceUpdateVersion on the wire, got %v", *md.ForceUpdateVersion)
			}
		})
	}
}
