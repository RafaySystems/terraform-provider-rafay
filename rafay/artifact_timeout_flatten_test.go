package rafay

import (
	"testing"
)

// Prior TF state for a Helm4 artifact spec whose config has an options
// block without a timeout.
func priorHelm4ArtifactSpecState() []interface{} {
	return []interface{}{map[string]interface{}{
		"type": "Helm4",
		"artifact": []interface{}{map[string]interface{}{
			"chart_path": []interface{}{map[string]interface{}{"name": "file://chart.tgz"}},
		}},
		"options": []interface{}{map[string]interface{}{
			"max_history": 10,
		}},
	}}
}

func optionsMap(t *testing.T, spec []interface{}) map[string]interface{} {
	t.Helper()
	opts := spec[0].(map[string]interface{})["options"].([]interface{})
	return opts[0].(map[string]interface{})
}

// A resource read (dataResource false) must not leak backend-normalized
// option values (timeout "0s") into the prior state maps.
func TestFlattenHelm4ResourceReadDoesNotLeakTimeout(t *testing.T) {
	prior := priorHelm4ArtifactSpecState()

	spec, err := ExpandArtifactSpec(prior)
	if err != nil {
		t.Fatalf("ExpandArtifactSpec: %v", err)
	}
	spec.GetHelm4Options().Timeout = "0s"

	ret, err := FlattenArtifactSpec(false, spec, prior)
	if err != nil {
		t.Fatalf("FlattenArtifactSpec: %v", err)
	}

	if v, ok := optionsMap(t, prior)["timeout"]; ok {
		t.Errorf("prior state options mutated with timeout=%v", v)
	}
	if v, ok := optionsMap(t, ret)["timeout"]; ok {
		t.Errorf("returned options contain timeout=%v", v)
	}
}

// A data-source read (dataResource true) must skip a backend timeout of
// "0s" when the prior state has no timeout, but keep real values.
func TestFlattenHelm4DataResourceTimeoutZero(t *testing.T) {
	for _, tc := range []struct {
		backend string
		want    string // "" means the key must be absent
	}{
		{backend: "0s", want: ""},
		{backend: "5m0s", want: "5m0s"},
	} {
		prior := priorHelm4ArtifactSpecState()

		spec, err := ExpandArtifactSpec(prior)
		if err != nil {
			t.Fatalf("ExpandArtifactSpec: %v", err)
		}
		spec.GetHelm4Options().Timeout = tc.backend

		ret, err := FlattenArtifactSpec(true, spec, prior)
		if err != nil {
			t.Fatalf("FlattenArtifactSpec: %v", err)
		}

		got, ok := optionsMap(t, ret)["timeout"]
		if tc.want == "" && ok {
			t.Errorf("backend timeout %q written to state as %v, want absent", tc.backend, got)
		}
		if tc.want != "" && got != tc.want {
			t.Errorf("state timeout = %v, want %q", got, tc.want)
		}
	}
}

// An explicitly configured "0s" must survive a data-source/import read so
// it does not flip-flop for users who set timeout = "0s" in config.
func TestFlattenHelm4KeepsExplicitZeroTimeout(t *testing.T) {
	prior := priorHelm4ArtifactSpecState()
	optionsMap(t, prior)["timeout"] = "0s"

	spec, err := ExpandArtifactSpec(prior)
	if err != nil {
		t.Fatalf("ExpandArtifactSpec: %v", err)
	}
	spec.GetHelm4Options().Timeout = "0s"

	ret, err := FlattenArtifactSpec(true, spec, prior)
	if err != nil {
		t.Fatalf("FlattenArtifactSpec: %v", err)
	}

	if got := optionsMap(t, ret)["timeout"]; got != "0s" {
		t.Errorf("state timeout = %v, want explicit \"0s\" kept", got)
	}
}
