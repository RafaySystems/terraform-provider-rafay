package rafay

import (
	"os"
	"testing"
)

// expandFiles must strip the "file://" scheme from the file name after reading
// the local data, so the stored name/relPath is a clean path rather than a
// literal "file://..." string (which a git-sourced artifact would treat as a
// repo path and fail to resolve).
func TestExpandFilesStripsFileScheme(t *testing.T) {
	if err := os.WriteFile("values.yaml", []byte("replicas: 1\n"), 0o600); err != nil {
		t.Fatalf("setup: %v", err)
	}
	defer os.Remove("values.yaml")

	got, err := expandFiles([]interface{}{
		map[string]interface{}{"name": "file://values.yaml"},
	})
	if err != nil {
		t.Fatalf("expandFiles: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 file, got %d", len(got))
	}
	if got[0].Name != "values.yaml" {
		t.Fatalf("expected stripped name %q, got %q", "values.yaml", got[0].Name)
	}
	if string(got[0].Data) != "replicas: 1\n" {
		t.Fatalf("expected local file data to be read, got %q", string(got[0].Data))
	}
}

// A plain (non-scheme) name is a repo-relative path and must pass through unchanged.
func TestExpandFilesKeepsPlainName(t *testing.T) {
	got, err := expandFiles([]interface{}{
		map[string]interface{}{"name": "artifacts/values.yaml"},
	})
	if err != nil {
		t.Fatalf("expandFiles: %v", err)
	}
	if got[0].Name != "artifacts/values.yaml" {
		t.Fatalf("expected name unchanged, got %q", got[0].Name)
	}
}
