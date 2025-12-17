package resource_eks_cluster

import (
	"context"
	"reflect"
	"testing"

	"github.com/RafaySystems/terraform-provider-rafay/rafay"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestFlattenMetadata(t *testing.T) {
	cases := []struct {
		name     string
		in       *rafay.EKSClusterMetadata
		expected MetadataValue
	}{
		{
			name: "only name",
			in: &rafay.EKSClusterMetadata{
				Name: "my-cluster",
			},
			expected: MetadataValue{
				Name:   types.StringValue("my-cluster"),
				Labels: types.MapNull(types.StringType),
				state:  attr.ValueStateKnown,
			},
		},
		{
			name: "name and labels",
			in: &rafay.EKSClusterMetadata{
				Name: "my-cluster",
				Labels: map[string]string{
					"env":  "prod",
					"team": "devops",
				},
			},
			expected: MetadataValue{
				Name: types.StringValue("my-cluster"),
				Labels: types.MapValueMust(
					types.StringType,
					map[string]attr.Value{
						"env":  types.StringValue("prod"),
						"team": types.StringValue("devops"),
					},
				),
				state: attr.ValueStateKnown,
			},
		},
		{
			name: "nil input",
			in:   nil,
			expected: MetadataValue{
				state: attr.ValueStateNull,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var v MetadataValue
			diags := v.Flatten(context.TODO(), tc.in)
			if diags.HasError() {
				t.Fatalf("unexpected diags: %v", diags)
			}
			if !reflect.DeepEqual(v, tc.expected) {
				t.Fatalf("expected %#v, got %#v", tc.expected, v)
			}
		})
	}
}
