package rafay

import (
	"strings"
	"testing"

	yamlf "github.com/goccy/go-yaml"
	"github.com/hashicorp/go-cty/cty"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// the cluster resource is declarative: an unset or empty optional_addons must
// still reach the API as an explicit [] so the selection is cleared
func TestExpandAKSClusterSpecAlwaysSendsOptionalAddons(t *testing.T) {
	rawConfig := cty.ListVal([]cty.Value{cty.EmptyObjectVal})
	for name, in := range map[string]map[string]interface{}{
		"unset": {"type": "aks"},
		"empty": {"type": "aks", "optional_addons": []interface{}{}},
	} {
		t.Run(name, func(t *testing.T) {
			spec := expandAKSClusterSpec([]interface{}{in}, rawConfig)
			out, err := yamlf.Marshal(&AKSCluster{Spec: spec})
			require.NoError(t, err)
			assert.True(t, strings.Contains(string(out), "optionalAddons: []\n"), "want optionalAddons: [] in yaml, got:\n%s", out)
		})
	}
}

func TestFlattenAKSClusterSpecReadsOptionalAddons(t *testing.T) {
	prior := []interface{}{map[string]interface{}{"optional_addons": []interface{}{"stale"}}}
	rawState := cty.NullVal(cty.Object(map[string]cty.Type{}))

	got := flattenAKSClusterSpec(&AKSClusterSpec{OptionalAddons: []string{"a", "b"}}, prior, rawState)
	assert.Equal(t, []interface{}{"a", "b"}, got[0].(map[string]interface{})["optional_addons"])

	prior = []interface{}{map[string]interface{}{"optional_addons": []interface{}{"stale"}}}
	got = flattenAKSClusterSpec(&AKSClusterSpec{}, prior, rawState)
	assert.Empty(t, got[0].(map[string]interface{})["optional_addons"])
}
