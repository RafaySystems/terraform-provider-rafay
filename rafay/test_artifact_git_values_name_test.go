package rafay

import (
	"testing"
)

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

// RC-50217: a git-sourced values_ref must reject an explicitly configured empty
// file name, while a populated name and an omitted block remain valid.
func TestGitValuesRefEmptyName(t *testing.T) {
	t.Run("git values_ref with empty name errors", func(t *testing.T) {
		input := valuesRefSpecInput("my-git-repo", []interface{}{
			map[string]interface{}{"name": ""},
		})
		if _, err := ExpandArtifactSpec(input); err == nil {
			t.Fatalf("expected error for empty git values path name, got nil")
		}
	})

	t.Run("git values_ref with whitespace name errors", func(t *testing.T) {
		input := valuesRefSpecInput("my-git-repo", []interface{}{
			map[string]interface{}{"name": "   "},
		})
		if _, err := ExpandArtifactSpec(input); err == nil {
			t.Fatalf("expected error for whitespace git values path name, got nil")
		}
	})

	t.Run("git values_ref with valid name is accepted", func(t *testing.T) {
		input := valuesRefSpecInput("my-git-repo", []interface{}{
			map[string]interface{}{"name": "envs/prod/values.yaml"},
		})
		if _, err := ExpandArtifactSpec(input); err != nil {
			t.Fatalf("expected no error for valid git values path, got %v", err)
		}
	})

	t.Run("git values_ref with no values_paths block is accepted", func(t *testing.T) {
		input := valuesRefSpecInput("my-git-repo", []interface{}{})
		if _, err := ExpandArtifactSpec(input); err != nil {
			t.Fatalf("expected no error when values_paths is omitted, got %v", err)
		}
	})

	t.Run("values_ref without repository does not trigger git validation", func(t *testing.T) {
		input := valuesRefSpecInput("", []interface{}{
			map[string]interface{}{"name": ""},
		})
		if _, err := ExpandArtifactSpec(input); err != nil {
			t.Fatalf("expected no error when repository is unset (not git-sourced), got %v", err)
		}
	})
}
