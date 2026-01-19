package helpers

import (
	"embed"
	"path"
	"testing"
)

// LoadFixture reads a fixture from the embedded FS, joining it to the default testdata directory.
func LoadFixture(t *testing.T, fs embed.FS, fileName string) string {
	t.Helper()
	fixturePath := path.Join("testdata", fileName)
	data, err := fs.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("failed to read fixture %s: %v", fixturePath, err)
	}
	return string(data)
}
