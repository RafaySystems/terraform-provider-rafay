package rafay

import (
	"context"
	"testing"

	"github.com/RafaySystems/rafay-common/proto/types/hub/infrapb"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func aksV3BlueprintBlock() []interface{} {
	return []interface{}{map[string]interface{}{
		"name": "bp", "version": "v3", "optional_addons": []interface{}{"a"},
	}}
}

// either blueprint block reaches the API; before, a blueprint block was dropped
// and the server answered "no blueprint config found"
func TestExpandAKSV3SpecBlueprintBlocks(t *testing.T) {
	spec, err := expandClusterV3Spec([]interface{}{map[string]interface{}{"type": "aks", "blueprint": aksV3BlueprintBlock()}})
	require.NoError(t, err)
	require.NotNil(t, spec.Blueprint)
	assert.Equal(t, "bp", spec.Blueprint.Name)
	assert.Equal(t, "v3", spec.Blueprint.Version)
	assert.Equal(t, []string{"a"}, spec.Blueprint.OptionalAddons)
	assert.Nil(t, spec.BlueprintConfig)

	spec, err = expandClusterV3Spec([]interface{}{map[string]interface{}{"type": "aks", "blueprint_config": aksV3BlueprintBlock()}})
	require.NoError(t, err)
	require.NotNil(t, spec.BlueprintConfig)
	assert.Equal(t, "bp", spec.BlueprintConfig.Name)
	assert.Nil(t, spec.Blueprint)
}

// the server value lands in the block the state already uses, whichever field
// the server returned it in
func TestFlattenAKSV3SpecBlueprintBlocks(t *testing.T) {
	stale := []interface{}{map[string]interface{}{"name": "old", "version": "v1"}}
	fromServer := map[string]*infrapb.ClusterSpec{
		"blueprint only":       {Blueprint: &infrapb.ClusterBlueprint{Name: "bp", Version: "v3", OptionalAddons: []string{"a"}}},
		"blueprintConfig only": {BlueprintConfig: &infrapb.BlueprintConfig{Name: "bp", Version: "v3", OptionalAddons: []string{"a"}}},
		"both, blueprint wins": {Blueprint: &infrapb.ClusterBlueprint{Name: "bp", Version: "v3", OptionalAddons: []string{"a"}}, BlueprintConfig: &infrapb.BlueprintConfig{Name: "other"}},
	}
	for name, in := range fromServer {
		for _, block := range []string{"blueprint", "blueprint_config"} {
			t.Run(name+"/state uses "+block, func(t *testing.T) {
				prior := []interface{}{map[string]interface{}{block: stale}}
				got := flattenClusterV3Spec(in, prior)[0].(map[string]interface{})
				assert.Equal(t, aksV3BlueprintBlock(), got[block])
			})
		}
	}

	// nothing in state yet (import): the long-standing blueprint_config block
	got := flattenClusterV3Spec(fromServer["blueprint only"], nil)[0].(map[string]interface{})
	assert.Equal(t, aksV3BlueprintBlock(), got["blueprint_config"])
	assert.Nil(t, got["blueprint"])
}

func TestAKSV3PlanRejectsBothBlueprintBlocks(t *testing.T) {
	r := resourceAKSClusterV3()
	plan := func(spec map[string]interface{}) error {
		spec["type"] = "aks"
		cfg := terraform.NewResourceConfigRaw(map[string]interface{}{
			"metadata": []interface{}{map[string]interface{}{"name": "c1", "project": "p1"}},
			"spec":     []interface{}{spec},
		})
		_, err := r.Diff(context.Background(), nil, cfg, nil)
		return err
	}

	err := plan(map[string]interface{}{"blueprint": aksV3BlueprintBlock(), "blueprint_config": aksV3BlueprintBlock()})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "either blueprint or blueprint_config")

	assert.NoError(t, plan(map[string]interface{}{"blueprint": aksV3BlueprintBlock()}))
	assert.NoError(t, plan(map[string]interface{}{"blueprint_config": aksV3BlueprintBlock()}))
}
