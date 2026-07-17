package rafay

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	commonpb "github.com/RafaySystems/rafay-common/proto/types/hub/commonpb"
)

// Allowed values for the Helm4 strategy-style options. These mirror the
// helm v4 contract (and the schema field descriptions) so that only valid
// strategies reach the backend. Empty / unset is always allowed and means
// "use the Helm default".
var (
	helm4WaitStrategyValues    = []string{"watcher", "legacy", "hookOnly"}
	helm4DryRunStrategyValues  = []string{"none", "client", "server"}
	helm4ServerSideApplyValues = []string{"true", "false", "auto"}
)

// validateHelm4Option returns an error when value is not one of the allowed
// values for the given option. The check is case-sensitive to match the helm
// v4 contract (e.g. "hookOnly").
func validateHelm4Option(field, value string, allowed []string) error {
	for _, a := range allowed {
		if value == a {
			return nil
		}
	}
	return fmt.Errorf("invalid value %q for option %q; valid values are: %s", value, field, strings.Join(allowed, ", "))
}

// helm4ArtifactTranspose mirrors the proto JSON shape emitted by
// ArtifactSpec.MarshalJSON when Type == "Helm4". The artifact wire types
// (UploadedHelm, HelmInGitRepo, HelmInHelmRepo, HelmInCatalog) are shared
// with Helm3, so the Artifact field carries the same fields. Options is
// Helm4Options-shaped and is intentionally separate from artifactTranspose.
type helm4ArtifactTranspose struct {
	Type string `json:"type,omitempty"`

	Artifact struct {
		Catalog      string  `json:"catalog,omitempty"`
		Repository   string  `json:"repository,omitempty"`
		Revision     string  `json:"revision,omitempty"`
		ChartPath    *File   `json:"chartPath,omitempty"`
		ValuesPaths  []*File `json:"valuesPaths,omitempty"`
		ChartName    string  `json:"chartName,omitempty"`
		ChartVersion string  `json:"chartVersion,omitempty"`
		Project      string  `json:"project,omitempty"`
		ValuesRef    struct {
			Repository  string  `json:"repository,omitempty"`
			Revision    string  `json:"revision,omitempty"`
			ValuesPaths []*File `json:"valuesPaths,omitempty"`
		} `json:"valuesRef,omitempty"`
	} `json:"artifact,omitempty"`

	Options helm4OptionsTranspose `json:"options,omitempty"`
}

// helm4OptionsTranspose mirrors commonpb.Helm4Options on the wire.
// Field JSON tags must match the proto-generated tags in artifacts.pb.go.
type helm4OptionsTranspose struct {
	Set                      []string          `json:"set,omitempty"`
	Labels                   map[string]string `json:"labels,omitempty"`
	WaitStrategy             string            `json:"waitStrategy,omitempty"`
	WaitForJobs              bool              `json:"waitForJobs,omitempty"`
	Timeout                  string            `json:"timeout,omitempty"`
	DryRunStrategy           string            `json:"dryRunStrategy,omitempty"`
	HideSecret               bool              `json:"hideSecret,omitempty"`
	DisableHooks             bool              `json:"disableHooks,omitempty"`
	SubNotes                 bool              `json:"subNotes,omitempty"`
	HideNotes                bool              `json:"hideNotes,omitempty"`
	SkipCrds                 bool              `json:"skipCrds,omitempty"`
	SkipSchemaValidation     bool              `json:"skipSchemaValidation,omitempty"`
	DisableOpenAPIValidation bool              `json:"disableOpenAPIValidation,omitempty"`
	ServerSideApply          string            `json:"serverSideApply,omitempty"`
	ForceReplace             bool              `json:"forceReplace,omitempty"`
	ForceConflicts           bool              `json:"forceConflicts,omitempty"`
	TakeOwnership            bool              `json:"takeOwnership,omitempty"`
	Replace                  bool              `json:"replace,omitempty"`
	MaxHistory               int32             `json:"maxHistory,omitempty"`
	RollbackOnFailure        bool              `json:"rollbackOnFailure,omitempty"`
	CleanupOnFail            bool              `json:"cleanupOnFail,omitempty"`
	ResetValues              bool              `json:"resetValues,omitempty"`
	ReuseValues              bool              `json:"reuseValues,omitempty"`
	ResetThenReuseValues     bool              `json:"resetThenReuseValues,omitempty"`
	Description              string            `json:"description,omitempty"`
	DependencyUpdate         bool              `json:"dependencyUpdate,omitempty"`
	EnableDns                bool              `json:"enableDns,omitempty"`
}

// ExpandHelm4Artifact builds an ArtifactSpec for type "Helm4". It mirrors
// ExpandArtifact but reads the Helm4-specific fields from the options block.
func ExpandHelm4Artifact(ap []interface{}) (*commonpb.ArtifactSpec, error) {
	if len(ap) == 0 || ap[0] == nil {
		return nil, fmt.Errorf("%s", "ExpandHelm4Artifact empty input")
	}

	obj := commonpb.ArtifactSpec{}
	at := helm4ArtifactTranspose{}
	at.Type = commonpb.ArtifactTypeHelm4
	var err error

	inp, ok := ap[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("%s", "ExpandHelm4Artifact input is not a map")
	}
	if vp, ok := inp["artifact"].([]interface{}); ok && len(vp) > 0 {
		if vp[0] == nil {
			return nil, fmt.Errorf("%s", "ExpandHelm4Artifact empty artifact")
		}
		in, ok := vp[0].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("%s", "ExpandHelm4Artifact artifact is not a map")
		}

		if v, ok := in["catalog"].(string); ok && len(v) > 0 {
			at.Artifact.Catalog = v
		}
		if v, ok := in["chart_name"].(string); ok && len(v) > 0 {
			at.Artifact.ChartName = v
		}
		if v, ok := in["chart_path"].([]interface{}); ok && len(v) > 0 {
			at.Artifact.ChartPath = expandFile(v)
		}
		if v, ok := in["chart_version"].(string); ok && len(v) > 0 {
			at.Artifact.ChartVersion = v
		}
		if v, ok := in["repository"].(string); ok && len(v) > 0 {
			at.Artifact.Repository = v
		}
		if v, ok := in["revision"].(string); ok && len(v) > 0 {
			at.Artifact.Revision = v
		}
		if v, ok := in["project"].(string); ok && len(v) > 0 {
			at.Artifact.Project = v
		}
		if v, ok := readTerraformValuesPathsBlocks(in); ok {
			at.Artifact.ValuesPaths, err = expandTerraformValuesPathsBlocks(v)
			if err != nil {
				return nil, err
			}
		}
		if v, ok := in["values_ref"].([]interface{}); ok && len(v) > 0 {
			if v[0] != nil {
				inVref := v[0].(map[string]interface{})
				if v, ok := inVref["repository"].(string); ok && len(v) > 0 {
					at.Artifact.ValuesRef.Repository = v
				}
				if v, ok := inVref["revision"].(string); ok && len(v) > 0 {
					at.Artifact.ValuesRef.Revision = v
				}
				if v, ok := readTerraformValuesPathsBlocks(inVref); ok {
					at.Artifact.ValuesRef.ValuesPaths, err = expandTerraformValuesPathsBlocks(v)
					if err != nil {
						return nil, err
					}
				}
			}
		}
	}

	if vp, ok := inp["options"].([]interface{}); ok && len(vp) > 0 && vp[0] != nil {
		in := vp[0].(map[string]interface{})
		if v, ok := in["set"].([]interface{}); ok && len(v) > 0 {
			for _, value := range v {
				if value != nil && value.(string) != "" {
					at.Options.Set = append(at.Options.Set, value.(string))
				}
			}
		}
		if v, ok := in["labels"].(map[string]interface{}); ok && len(v) > 0 {
			at.Options.Labels = toMapString(v)
		}
		if v, ok := in["wait_strategy"].(string); ok && len(v) > 0 {
			if err := validateHelm4Option("wait_strategy", v, helm4WaitStrategyValues); err != nil {
				return nil, err
			}
			at.Options.WaitStrategy = v
		}
		if v, ok := in["wait_for_jobs"].(bool); ok {
			at.Options.WaitForJobs = v
		}
		if v, ok := in["timeout"].(string); ok && len(v) > 0 {
			at.Options.Timeout = v
		}
		if v, ok := in["dry_run_strategy"].(string); ok && len(v) > 0 {
			if err := validateHelm4Option("dry_run_strategy", v, helm4DryRunStrategyValues); err != nil {
				return nil, err
			}
			at.Options.DryRunStrategy = v
		}
		if v, ok := in["hide_secret"].(bool); ok {
			at.Options.HideSecret = v
		}
		if v, ok := in["disable_hooks"].(bool); ok {
			at.Options.DisableHooks = v
		}
		if v, ok := in["sub_notes"].(bool); ok {
			at.Options.SubNotes = v
		}
		if v, ok := in["hide_notes"].(bool); ok {
			at.Options.HideNotes = v
		}
		if v, ok := in["skip_crds"].(bool); ok {
			at.Options.SkipCrds = v
		}
		if v, ok := in["skip_schema_validation"].(bool); ok {
			at.Options.SkipSchemaValidation = v
		}
		if v, ok := in["disable_open_api_validation"].(bool); ok {
			at.Options.DisableOpenAPIValidation = v
		}
		if v, ok := in["server_side_apply"].(string); ok && len(v) > 0 {
			if err := validateHelm4Option("server_side_apply", v, helm4ServerSideApplyValues); err != nil {
				return nil, err
			}
			at.Options.ServerSideApply = v
		}
		if v, ok := in["force_replace"].(bool); ok {
			at.Options.ForceReplace = v
		}
		if v, ok := in["force_conflicts"].(bool); ok {
			at.Options.ForceConflicts = v
		}
		if v, ok := in["take_ownership"].(bool); ok {
			at.Options.TakeOwnership = v
		}
		if v, ok := in["replace"].(bool); ok {
			at.Options.Replace = v
		}
		if v, ok := in["max_history"].(int); ok {
			at.Options.MaxHistory = int32(v)
		}
		if v, ok := in["rollback_on_failure"].(bool); ok {
			at.Options.RollbackOnFailure = v
		}
		if v, ok := in["cleanup_on_fail"].(bool); ok {
			at.Options.CleanupOnFail = v
		}
		if v, ok := in["reset_values"].(bool); ok {
			at.Options.ResetValues = v
		}
		if v, ok := in["reuse_values"].(bool); ok {
			at.Options.ReuseValues = v
		}
		if v, ok := in["reset_then_reuse_values"].(bool); ok {
			at.Options.ResetThenReuseValues = v
		}
		if v, ok := in["description"].(string); ok && len(v) > 0 {
			at.Options.Description = v
		}
		if v, ok := in["dependency_update"].(bool); ok {
			at.Options.DependencyUpdate = v
		}
		if v, ok := in["enable_dns"].(bool); ok {
			at.Options.EnableDns = v
		}
	}

	jsonSpec, err := json.Marshal(at)
	if err != nil {
		return nil, err
	}

	if err := obj.UnmarshalJSON(jsonSpec); err != nil {
		log.Println("ExpandHelm4Artifact UnmarshalJSON error ", err)
		return nil, err
	}

	return &obj, nil
}

// FlattenHelm4Artifact populates the artifact map for type Helm4.
func FlattenHelm4Artifact(at *helm4ArtifactTranspose, p []interface{}) ([]interface{}, error) {
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(at.Artifact.Repository) > 0 {
		obj["repository"] = at.Artifact.Repository
	}
	if len(at.Artifact.Revision) > 0 {
		obj["revision"] = at.Artifact.Revision
	}
	if len(at.Artifact.Project) > 0 {
		obj["project"] = at.Artifact.Project
	}
	if at.Artifact.ChartPath != nil {
		obj["chart_path"] = flattenFile(at.Artifact.ChartPath)
	}
	if at.Artifact.ValuesPaths != nil {
		flat := flattenValuesPathsForState(
			at.Artifact.ValuesPaths,
			priorHasValuesPathsBlock(obj, "values_paths"),
		)
		if flat == nil {
			delete(obj, "values_paths")
		} else {
			obj["values_paths"] = flat
		}
	}
	if len(at.Artifact.Catalog) > 0 {
		obj["catalog"] = at.Artifact.Catalog
	}
	if len(at.Artifact.ChartName) > 0 {
		obj["chart_name"] = at.Artifact.ChartName
	}
	if len(at.Artifact.ChartVersion) > 0 {
		obj["chart_version"] = at.Artifact.ChartVersion
	}

	v, ok := obj["values_ref"].([]interface{})
	if !ok {
		v = []interface{}{}
	}
	if at.Artifact.ValuesRef.Repository != "" {
		obj["values_ref"] = flattenHelm4ValuesRef(at, v)
	}

	return []interface{}{obj}, nil
}

func flattenHelm4ValuesRef(at *helm4ArtifactTranspose, p []interface{}) []interface{} {
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(at.Artifact.ValuesRef.Repository) > 0 {
		obj["repository"] = at.Artifact.ValuesRef.Repository
	}
	if len(at.Artifact.ValuesRef.Revision) > 0 {
		obj["revision"] = at.Artifact.ValuesRef.Revision
	}
	if at.Artifact.ValuesRef.ValuesPaths != nil {
		flat := flattenValuesPathsForState(
			at.Artifact.ValuesRef.ValuesPaths,
			priorHasValuesPathsBlock(obj, "values_paths"),
		)
		if flat == nil {
			delete(obj, "values_paths")
		} else {
			obj["values_paths"] = flat
		}
	}
	return []interface{}{obj}
}

// FlattenHelm4ArtifactOptions populates the options block in TF state.
func FlattenHelm4ArtifactOptions(at *helm4ArtifactTranspose, p []interface{}) ([]interface{}, error) {
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(at.Options.Set) > 0 {
		obj["set"] = toArrayInterface(at.Options.Set)
	}
	if len(at.Options.Labels) > 0 {
		obj["labels"] = toMapInterface(at.Options.Labels)
	}
	if len(at.Options.WaitStrategy) > 0 {
		obj["wait_strategy"] = at.Options.WaitStrategy
	}
	obj["wait_for_jobs"] = at.Options.WaitForJobs
	if len(at.Options.Timeout) > 0 {
		// The backend normalizes an unset timeout to "0s"; writing that
		// back when the config never set a timeout causes a perpetual
		// "0s" -> null diff on every plan.
		if prior, _ := obj["timeout"].(string); at.Options.Timeout != "0s" || prior != "" {
			obj["timeout"] = at.Options.Timeout
		}
	}
	if len(at.Options.DryRunStrategy) > 0 {
		obj["dry_run_strategy"] = at.Options.DryRunStrategy
	}
	obj["hide_secret"] = at.Options.HideSecret
	obj["disable_hooks"] = at.Options.DisableHooks
	obj["sub_notes"] = at.Options.SubNotes
	obj["hide_notes"] = at.Options.HideNotes
	obj["skip_crds"] = at.Options.SkipCrds
	obj["skip_schema_validation"] = at.Options.SkipSchemaValidation
	obj["disable_open_api_validation"] = at.Options.DisableOpenAPIValidation
	if len(at.Options.ServerSideApply) > 0 {
		obj["server_side_apply"] = at.Options.ServerSideApply
	}
	obj["force_replace"] = at.Options.ForceReplace
	obj["force_conflicts"] = at.Options.ForceConflicts
	obj["take_ownership"] = at.Options.TakeOwnership
	obj["replace"] = at.Options.Replace
	obj["max_history"] = int(at.Options.MaxHistory)
	obj["rollback_on_failure"] = at.Options.RollbackOnFailure
	obj["cleanup_on_fail"] = at.Options.CleanupOnFail
	obj["reset_values"] = at.Options.ResetValues
	obj["reuse_values"] = at.Options.ReuseValues
	obj["reset_then_reuse_values"] = at.Options.ResetThenReuseValues
	if len(at.Options.Description) > 0 {
		obj["description"] = at.Options.Description
	}
	obj["dependency_update"] = at.Options.DependencyUpdate
	obj["enable_dns"] = at.Options.EnableDns

	return []interface{}{obj}, nil
}
