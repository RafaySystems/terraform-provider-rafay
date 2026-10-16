package rafay

import (
	"testing"

	commonpb "github.com/RafaySystems/rafay-common/proto/types/hub/commonpb"
)

func names(files []*commonpb.File) []string {
	out := make([]string, 0, len(files))
	for _, f := range files {
		out = append(out, f.Name)
	}
	return out
}

// RC-50217: dropPlaceholderValuesPaths must remove the empty placeholder values
// path emitted on create while preserving genuinely configured values files.
func TestDropPlaceholderValuesPaths(t *testing.T) {
	t.Run("drops empty placeholder on helm_in_git", func(t *testing.T) {
		spec := &commonpb.ArtifactSpec{
			Artifact: &commonpb.ArtifactSpec_HelmInGitRepo{
				HelmInGitRepo: &commonpb.HelmInGitRepo{
					ValuesPaths: []*commonpb.File{{}},
				},
			},
		}
		dropPlaceholderValuesPaths(spec)
		if got := spec.GetHelmInGitRepo().ValuesPaths; got != nil {
			t.Fatalf("expected nil ValuesPaths, got %v", names(got))
		}
	})

	t.Run("keeps configured values file", func(t *testing.T) {
		spec := &commonpb.ArtifactSpec{
			Artifact: &commonpb.ArtifactSpec_HelmInGitRepo{
				HelmInGitRepo: &commonpb.HelmInGitRepo{
					ValuesPaths: []*commonpb.File{{Name: "values/prod.yaml"}, {}},
				},
			},
		}
		dropPlaceholderValuesPaths(spec)
		got := names(spec.GetHelmInGitRepo().ValuesPaths)
		if len(got) != 1 || got[0] != "values/prod.yaml" {
			t.Fatalf("expected [values/prod.yaml], got %v", got)
		}
	})

	t.Run("cleans values_ref placeholder too", func(t *testing.T) {
		spec := &commonpb.ArtifactSpec{
			Artifact: &commonpb.ArtifactSpec_HelmInGitRepo{
				HelmInGitRepo: &commonpb.HelmInGitRepo{
					ValuesRef: &commonpb.OverrideRepoReference{
						ValuesPaths: []*commonpb.File{{}},
					},
				},
			},
		}
		dropPlaceholderValuesPaths(spec)
		if got := spec.GetHelmInGitRepo().GetValuesRef().ValuesPaths; got != nil {
			t.Fatalf("expected nil ValuesRef.ValuesPaths, got %v", names(got))
		}
	})

	t.Run("nil spec is a no-op", func(t *testing.T) {
		dropPlaceholderValuesPaths(nil)
	})
}
