package rafay

import "testing"

// helper to build the ExpandArtifactSpec input for a Helm artifact whose
// values_ref carries the given repository and values_paths entries.
func valuesRefSpecInput(repository string, valuesPaths []interface{}) []interface{} {
	return []interface{}{
		map[string]interface{}{
			"type": "Helm",
			"artifact": []interface{}{
				map[string]interface{}{
					"chart_name": "nginx",
					"values_ref": []interface{}{
						map[string]interface{}{
							"repository":   repository,
							"revision":     "main",
							"values_paths": valuesPaths,
						},
					},
				},
			},
		},
	}
}

// RC-50217: a git-sourced values_ref must resolve to at least one usable values
// path. The important case is the empty LIST: Terraform prunes an all-empty
// `values_paths {}` block before it reaches the provider, so on create an empty
// git path arrives as an empty list rather than as an entry with an empty name.
// Asserting on the empty list reproduces that real pruned shape (the previous
// test injected `{"name": ""}` directly, which never happens through Terraform
// and so gave false confidence that create was covered).
func TestGitValuesRefRequiresUsablePath(t *testing.T) {
	mustError := func(label string, vp []interface{}) {
		t.Run(label, func(t *testing.T) {
			if _, err := ExpandArtifactSpec(valuesRefSpecInput("my-git-repo", vp)); err == nil {
				t.Fatalf("expected error, got nil")
			}
		})
	}

	// Pruned `values_paths {}` on create arrives as an empty list.
	mustError("git values_ref with no usable path (pruned empty list)", []interface{}{})
	// Explicit empty name survives on update (real path -> "").
	mustError("git values_ref with empty name", []interface{}{
		map[string]interface{}{"name": ""},
	})
	// Whitespace-only name is also not a usable path.
	mustError("git values_ref with whitespace name", []interface{}{
		map[string]interface{}{"name": "   "},
	})

	t.Run("git values_ref with a valid path is accepted", func(t *testing.T) {
		input := valuesRefSpecInput("my-git-repo", []interface{}{
			map[string]interface{}{"name": "envs/prod/values.yaml"},
		})
		if _, err := ExpandArtifactSpec(input); err != nil {
			t.Fatalf("expected no error for valid git values path, got %v", err)
		}
	})

	t.Run("non-git values_ref (no repository) is not validated", func(t *testing.T) {
		// Not git-sourced: an empty/omitted path is allowed (it is the upload-type
		// clear-on-update signal, not a git fetch path).
		if _, err := ExpandArtifactSpec(valuesRefSpecInput("", []interface{}{})); err != nil {
			t.Fatalf("expected no error when repository is unset (empty list), got %v", err)
		}
		if _, err := ExpandArtifactSpec(valuesRefSpecInput("", []interface{}{
			map[string]interface{}{"name": ""},
		})); err != nil {
			t.Fatalf("expected no error when repository is unset (empty name), got %v", err)
		}
	})
}

// gitChartSpecInput builds a HelmInGitRepo artifact (chart_path inside a git
// repository) with the given direct values_paths entries.
func gitChartSpecInput(valuesPaths []interface{}) []interface{} {
	artifact := map[string]interface{}{
		"repository": "my-git-repo",
		"revision":   "main",
		"chart_path": []interface{}{
			map[string]interface{}{"name": "charts/my-chart-0.1.0.tgz"},
		},
	}
	if valuesPaths != nil {
		artifact["values_paths"] = valuesPaths
	}
	return []interface{}{
		map[string]interface{}{
			"type":     "Helm",
			"artifact": []interface{}{artifact},
		},
	}
}

// RC-50217: a values_paths entry on a git-sourced Helm chart names a
// repo-relative path, so a blank name is illegal and must be rejected instead of
// silently applied. Values remain optional - an omitted/empty block is allowed.
func TestGitChartValuesPathRejectsBlank(t *testing.T) {
	mustError := func(label string, vp []interface{}) {
		t.Run(label, func(t *testing.T) {
			if _, err := ExpandArtifactSpec(gitChartSpecInput(vp)); err == nil {
				t.Fatalf("expected error, got nil")
			}
		})
	}
	mustOK := func(label string, vp []interface{}) {
		t.Run(label, func(t *testing.T) {
			if _, err := ExpandArtifactSpec(gitChartSpecInput(vp)); err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}

	// A configured-but-blank git values path is rejected (the case that
	// survives Terraform pruning: explicit empty name on update, whitespace name).
	mustError("git chart values_paths with empty name", []interface{}{
		map[string]interface{}{"name": ""},
	})
	mustError("git chart values_paths with whitespace name", []interface{}{
		map[string]interface{}{"name": "   "},
	})
	mustError("git chart values_paths with one good and one blank", []interface{}{
		map[string]interface{}{"name": "values/prod.yaml"},
		map[string]interface{}{"name": ""},
	})

	// Values are optional for a git chart: an omitted or pruned-empty block uses
	// the chart defaults and must be accepted.
	mustOK("git chart with a real values path", []interface{}{
		map[string]interface{}{"name": "values/prod.yaml"},
	})
	mustOK("git chart with omitted values_paths", nil)
	mustOK("git chart with pruned empty values_paths", []interface{}{})
}
