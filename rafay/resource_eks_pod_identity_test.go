package rafay

import (
	"context"
	"testing"

	"github.com/go-yaml/yaml"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func podIdentityBoolPtr(value bool) *bool {
	return &value
}

func TestFlattenIAMPodIdentityAssociationsCreateServiceAccount(t *testing.T) {
	previous := []interface{}{
		map[string]interface{}{
			"namespace":              "ns",
			"service_account_name":   "sa",
			"create_service_account": true,
		},
	}

	falseAssoc := flattenEKSPodIdentityResource([]*eksPodIdentityResourceAssociation{{
		Namespace:            "ns",
		ServiceAccountName:   "sa",
		CreateServiceAccount: podIdentityBoolPtr(false),
	}}, previous)
	require.Len(t, falseAssoc, 1)
	assert.Equal(t, false, falseAssoc[0].(map[string]interface{})["create_service_account"])

	trueAssoc := flattenEKSPodIdentityResource([]*eksPodIdentityResourceAssociation{{
		Namespace:            "ns",
		ServiceAccountName:   "sa",
		CreateServiceAccount: podIdentityBoolPtr(true),
	}}, nil)
	require.Len(t, trueAssoc, 1)
	assert.Equal(t, true, trueAssoc[0].(map[string]interface{})["create_service_account"])
}

func TestCreateServiceAccountChangeIsPlanned(t *testing.T) {
	resource := resourceEKSPodIdentity()
	state := &terraform.InstanceState{
		ID: "sa/ns",
		Attributes: map[string]string{
			"metadata.#":                    "1",
			"metadata.0.cluster_name":       "cluster-x",
			"metadata.0.project_name":       "defaultproject",
			"spec.#":                        "1",
			"spec.0.namespace":              "ns",
			"spec.0.service_account_name":   "sa",
			"spec.0.create_service_account": "true",
		},
	}
	config := terraform.NewResourceConfigRaw(map[string]interface{}{
		"metadata": []interface{}{
			map[string]interface{}{
				"cluster_name": "cluster-x",
				"project_name": "defaultproject",
			},
		},
		"spec": []interface{}{
			map[string]interface{}{
				"namespace":              "ns",
				"service_account_name":   "sa",
				"create_service_account": false,
			},
		},
	})

	diff, err := resource.Diff(context.Background(), state, config, nil)
	require.NoError(t, err)
	require.NotNil(t, diff)
	change, ok := diff.Attributes["spec.0.create_service_account"]
	require.True(t, ok, "create_service_account change was not planned")
	assert.Equal(t, "true", change.Old)
	assert.Equal(t, "false", change.New)
}

func TestCreateServiceAccountYAMLPreservesExistingStateValues(t *testing.T) {
	// Stored values from an existing cluster state: explicit true and false must round-trip.
	for _, value := range []bool{true, false, true} {
		encoded, err := yaml.Marshal(eksPodIdentityResourceAssociation{
			Namespace:            "rafay-demo",
			ServiceAccountName:   "svc",
			RoleARN:              "arn:aws:iam::679196758854:role/gopi-podidentity",
			CreateServiceAccount: podIdentityBoolPtr(value),
		})
		require.NoError(t, err)
		if value {
			assert.Contains(t, string(encoded), "createServiceAccount: true")
		} else {
			assert.Contains(t, string(encoded), "createServiceAccount: false")
		}
		var decoded eksPodIdentityResourceAssociation
		require.NoError(t, yaml.Unmarshal(encoded, &decoded))
		require.NotNil(t, decoded.CreateServiceAccount)
		assert.Equal(t, value, *decoded.CreateServiceAccount)
	}
}

func TestOmittedCreateServiceAccountStaysUnset(t *testing.T) {
	encoded, err := yaml.Marshal(eksPodIdentityResourceAssociation{
		Namespace:          "rafay-demo",
		ServiceAccountName: "svc",
	})
	require.NoError(t, err)
	assert.NotContains(t, string(encoded), "createServiceAccount")

	var decoded eksPodIdentityResourceAssociation
	require.NoError(t, yaml.Unmarshal(encoded, &decoded))
	assert.Nil(t, decoded.CreateServiceAccount)

	flattened := flattenEKSPodIdentityResource([]*eksPodIdentityResourceAssociation{{
		Namespace:          "ns",
		ServiceAccountName: "sa",
	}}, nil)
	require.Len(t, flattened, 1)
	_, present := flattened[0].(map[string]interface{})["create_service_account"]
	assert.False(t, present, "an omitted field was stored as false")
}

func TestOmittedCreateServiceAccountKeepsExistingState(t *testing.T) {
	resource := resourceEKSPodIdentity()
	for _, existing := range []string{"true", "false"} {
		state := &terraform.InstanceState{
			ID: "sa/ns",
			Attributes: map[string]string{
				"metadata.#":                    "1",
				"metadata.0.cluster_name":       "cluster-x",
				"metadata.0.project_name":       "defaultproject",
				"spec.#":                        "1",
				"spec.0.namespace":              "ns",
				"spec.0.service_account_name":   "sa",
				"spec.0.create_service_account": existing,
			},
		}
		config := terraform.NewResourceConfigRaw(map[string]interface{}{
			"metadata": []interface{}{
				map[string]interface{}{
					"cluster_name": "cluster-x",
					"project_name": "defaultproject",
				},
			},
			"spec": []interface{}{
				map[string]interface{}{
					"namespace":            "ns",
					"service_account_name": "sa",
				},
			},
		})

		diff, err := resource.Diff(context.Background(), state, config, nil)
		require.NoError(t, err)
		got := existing
		if diff != nil {
			if change, ok := diff.Attributes["spec.0.create_service_account"]; ok {
				got = change.New
			}
		}
		assert.Equal(t, existing, got, "removing the argument changed stored create_service_account from %s", existing)
	}
}

func TestExpandOmittedCreateServiceAccountStaysNil(t *testing.T) {
	association := func(name string, create cty.Value) cty.Value {
		return cty.ObjectVal(map[string]cty.Value{
			"namespace":              cty.StringVal("ns"),
			"service_account_name":   cty.StringVal(name),
			"role_arn":               cty.StringVal("arn"),
			"create_service_account": create,
		})
	}
	raw := cty.ListVal([]cty.Value{
		association("omitted", cty.NullVal(cty.Bool)),
		association("explicit-false", cty.False),
		association("explicit-true", cty.True),
	})
	// d.Get fills an omitted optional bool with false, matching an explicit false.
	got := expandEKSPodIdentityResource([]interface{}{
		map[string]interface{}{"namespace": "ns", "service_account_name": "omitted", "role_arn": "arn", "create_service_account": false},
		map[string]interface{}{"namespace": "ns", "service_account_name": "explicit-false", "role_arn": "arn", "create_service_account": false},
		map[string]interface{}{"namespace": "ns", "service_account_name": "explicit-true", "role_arn": "arn", "create_service_account": true},
	}, raw)

	require.Len(t, got, 3)
	assert.Nil(t, got[0].CreateServiceAccount)
	require.NotNil(t, got[1].CreateServiceAccount)
	assert.False(t, *got[1].CreateServiceAccount)
	require.NotNil(t, got[2].CreateServiceAccount)
	assert.True(t, *got[2].CreateServiceAccount)
}
