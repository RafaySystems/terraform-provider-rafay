package rafay

import (
	"encoding/json"
	"fmt"
	"log"

	commonpb "github.com/RafaySystems/rafay-common/proto/types/hub/commonpb"
	"github.com/davecgh/go-spew/spew"
)

type artifactTranspose struct {
	Type string `json:"type,omitempty"`

	Artifact struct {
		Catalog       string  `protobuf:"bytes,2,opt,name=catalog,proto3" json:"catalog,omitempty"`
		Repository    string  `protobuf:"bytes,1,opt,name=repository,proto3" json:"repository,omitempty"`
		Revision      string  `protobuf:"bytes,2,opt,name=revision,proto3" json:"revision,omitempty"`
		ChartPath     *File   `protobuf:"bytes,3,opt,name=chart_path,proto3" json:"chartPath,omitempty"`
		ValuesPaths   []*File `protobuf:"bytes,4,rep,name=values_paths,proto3" json:"valuesPaths,omitempty"`
		ChartName     string  `protobuf:"bytes,2,opt,name=chart_name,proto3" json:"chartName,omitempty"`
		ChartVersion  string  `protobuf:"bytes,3,opt,name=chart_version,proto3" json:"chartVersion,omitempty"`
		Paths         []*File `protobuf:"bytes,3,rep,name=paths,proto3" json:"paths,omitempty"`
		Configmap     *File   `protobuf:"bytes,1,opt,name=configmap,proto3" json:"configmap,omitempty"`
		Secret        *File   `protobuf:"bytes,2,opt,name=secret,proto3" json:"secret,omitempty"`
		Configuration *File   `protobuf:"bytes,3,opt,name=configuration,proto3" json:"configuration,omitempty"`
		Statefulset   *File   `protobuf:"bytes,4,opt,name=statefulset,proto3" json:"statefulset,omitempty"`
		Project       string  `protobuf:"bytes,4,opt,name=project,proto3" json:"project,omitempty"`
		ValuesRef     struct {
			Repository  string  `protobuf:"bytes,1,opt,name=repository,proto3" json:"repository,omitempty"`
			Revision    string  `protobuf:"bytes,2,opt,name=revision,proto3" json:"revision,omitempty"`
			ValuesPaths []*File `protobuf:"bytes,3,rep,name=valuesPaths,proto3" json:"valuesPaths,omitempty"`
		} `json:"valuesRef,omitempty"`
	} `json:"artifact,omitempty"`

	Options struct {
		Atomic                   bool     `protobuf:"varint,1,opt,name=atomic,proto3" json:"atomic,omitempty"`
		Wait                     bool     `protobuf:"varint,2,opt,name=wait,proto3" json:"wait,omitempty"`
		Force                    bool     `protobuf:"varint,3,opt,name=force,proto3" json:"force,omitempty"`
		NoHooks                  bool     `protobuf:"varint,4,opt,name=no_hooks,proto3" json:"noHooks,omitempty"`
		MaxHistory               int      `protobuf:"zigzag32,5,opt,name=max_history,proto3" json:"maxHistory,omitempty"`
		RenderSubChartNotes      bool     `protobuf:"varint,6,opt,name=render_sub_chart_notes,proto3" json:"renderSubChartNotes,omitempty"`
		ResetValues              bool     `protobuf:"varint,7,opt,name=reset_values,proto3" json:"resetValues,omitempty"`
		ReuseValues              bool     `protobuf:"varint,8,opt,name=reuse_values,proto3" json:"reuseValues,omitempty"`
		SetString                []string `protobuf:"bytes,9,rep,name=set_string,proto3" json:"setString,omitempty"`
		SkipCRDs                 bool     `protobuf:"varint,10,opt,name=skip_cr_ds,proto3" json:"skipCRDs,omitempty"`
		Timeout                  string   `protobuf:"bytes,11,opt,name=timeout,proto3" json:"timeout,omitempty"`
		CleanUpOnFail            bool     `protobuf:"varint,12,opt,name=clean_up_on_fail,proto3" json:"cleanUpOnFail,omitempty"`
		Description              string   `protobuf:"bytes,13,opt,name=description,proto3" json:"description,omitempty"`
		DisableOpenAPIValidation bool     `protobuf:"varint,14,opt,name=disable_open_api_validation,proto3" json:"disableOpenAPIValidation,omitempty"`
		KeepHistory              bool     `protobuf:"varint,15,opt,name=keep_history,proto3" json:"keepHistory,omitempty"`
		WaitForJobs              bool     `protobuf:"varint,16,opt,name=waitForJobs,proto3" json:"waitForJobs" yaml:"waitForJobs"`
	} `json:"options,omitempty"`
}

// ExpandArtifact expands tf state to ArtifactSpec
func ExpandArtifact(artifactType string, ap []interface{}) (*commonpb.ArtifactSpec, error) {
	if len(ap) == 0 || ap[0] == nil {
		return nil, fmt.Errorf("%s", "expandArtifact empty input")
	}

	obj := commonpb.ArtifactSpec{}
	at := artifactTranspose{}
	at.Type = artifactType
	var err error

	inp := ap[0].(map[string]interface{})
	if vp, ok := inp["artifact"].([]interface{}); ok && len(vp) > 0 {
		if len(vp) == 0 || vp[0] == nil {
			return nil, fmt.Errorf("%s", "expandArtifact empty artifact")
		}
		in := vp[0].(map[string]interface{})

		artfct := spew.Sprintf("%+v", in)
		log.Println("ExpandArtifact in ", artfct)
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

		if v, ok := in["configmap"].([]interface{}); ok && len(v) > 0 {
			at.Artifact.Configmap = expandFile(v)
		}

		if v, ok := in["configuration"].([]interface{}); ok && len(v) > 0 {
			at.Artifact.Configuration = expandFile(v)
			artfct = spew.Sprintf("%+v", at.Artifact.Configuration)
			log.Println("ExpandArtifact  at.Artifact.Configuration ", artfct)
		}

		if v, ok := in["paths"].([]interface{}); ok && len(v) > 0 {
			at.Artifact.Paths, err = expandFiles(v)
			if err != nil {
				return nil, err
			}
			artfct = spew.Sprintf("%+v", at.Artifact.Paths)
			log.Println("ExpandArtifact  at.Artifact.Paths ", artfct)
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

		if v, ok := in["secret"].([]interface{}); ok && len(v) > 0 {
			at.Artifact.Secret = expandFile(v)
		}

		if v, ok := in["statefulset"].([]interface{}); ok && len(v) > 0 {
			at.Artifact.Statefulset = expandFile(v)
		}

		if v, ok := in["values_paths"].([]interface{}); ok && len(v) > 0 {
			at.Artifact.ValuesPaths, err = expandFiles(v)
			if err != nil {
				return nil, err
			}
			artfct = spew.Sprintf("%+v", at.Artifact.ValuesPaths)
			log.Println("ExpandArtifact  at.Artifact.ValuesPaths ", artfct)
		}

		if v, ok := in["revision"].(string); ok && len(v) > 0 {
			at.Artifact.Revision = v
		}

		if v, ok := in["values_ref"].([]interface{}); ok && len(v) > 0 {
			//at.Artifact.Configmap = expandValuesRef(v)
			if v[0] == nil {
				log.Println("expandValuesRef empty options")
			} else {
				inVref := v[0].(map[string]interface{})
				if v, ok := inVref["repository"].(string); ok && len(v) > 0 {
					at.Artifact.ValuesRef.Repository = v
				}

				if v, ok := inVref["revision"].(string); ok && len(v) > 0 {
					at.Artifact.ValuesRef.Revision = v
				}

				if v, ok := inVref["values_paths"].([]interface{}); ok && len(v) > 0 {
					at.Artifact.ValuesRef.ValuesPaths, err = expandFiles(v)
					if err != nil {
						return nil, err
					}
					artfct = spew.Sprintf("%+v", at.Artifact.ValuesRef.ValuesPaths)
					log.Println("at.Artifact.ValuesRef.ValuesPaths ", artfct)
				}
			}
		}

		artfct = spew.Sprintf("%+v", at.Artifact)
		log.Println("ExpandArtifact  at.Artifact ", artfct)
	}

	if vp, ok := inp["options"].([]interface{}); ok && len(vp) > 0 {
		if len(vp) == 0 || vp[0] == nil {
			log.Println("expandArtifact empty options")
		} else {
			in := vp[0].(map[string]interface{})
			if v, ok := in["atomic"].(bool); ok {
				at.Options.Atomic = v
			}
			if v, ok := in["clean_up_on_fail"].(bool); ok {
				at.Options.CleanUpOnFail = v
			}
			if v, ok := in["description"].(string); ok && len(v) > 0 {
				at.Options.Description = v
			}
			if v, ok := in["disable_open_api_validation"].(bool); ok {
				at.Options.DisableOpenAPIValidation = v
			}
			if v, ok := in["force"].(bool); ok {
				at.Options.Force = v
			}
			if v, ok := in["keep_history"].(bool); ok {
				at.Options.KeepHistory = v
			}
			if v, ok := in["max_history"].(int); ok {
				at.Options.MaxHistory = v
				log.Println("ExpandArtifact max_history ", at.Options.MaxHistory)
			}
			if v, ok := in["no_hooks"].(bool); ok {
				at.Options.NoHooks = v
			}
			if v, ok := in["render_sub_chart_notes"].(bool); ok {
				at.Options.RenderSubChartNotes = v
			}
			if v, ok := in["reset_values"].(bool); ok {
				at.Options.ResetValues = v
			}
			if v, ok := in["reuse_values"].(bool); ok {
				at.Options.ReuseValues = v
			}
			if v, ok := in["skip_crd"].(bool); ok {
				at.Options.SkipCRDs = v
			}
			if v, ok := in["set_string"].([]string); ok && len(v) > 0 {
				at.Options.SetString = v
			}
			if v, ok := in["timeout"].(string); ok && len(v) > 0 {
				at.Options.Timeout = v
			}
			if v, ok := in["wait"].(bool); ok {
				at.Options.Wait = v
			}
			if v, ok := in["wait_for_jobs"].(bool); ok {
				at.Options.WaitForJobs = v
			}
			ops := spew.Sprintf("%+v", at.Options)
			log.Println("ExpandArtifact ops ", ops)
		}
	}

	// XXX Debug
	s := spew.Sprintf("%+v", at)
	log.Println("ExpandArtifact at", s)

	jsonSpec, err := json.Marshal(at)
	if err != nil {
		return nil, err
	}

	// XXX Debug
	log.Println("ExpandArtifact jsonSpec ", string(jsonSpec))

	err = obj.UnmarshalJSON(jsonSpec)
	if err != nil {
		log.Println("expandArtifact artifact UnmarshalJSON error ", err)
		return nil, err
	}

	// XXX Debug
	s1 := spew.Sprintf("%+v", obj)
	log.Println("ExpandArtifact obj", s1)

	return &obj, nil
}

func ExpandArtifactSpec(p []interface{}) (*commonpb.ArtifactSpec, error) {
	var err error
	var obj *commonpb.ArtifactSpec

	if len(p) == 0 || p[0] == nil {
		return nil, fmt.Errorf("%s", "ExpandArtifactSpec empty input")
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["type"].(string); ok && len(v) > 0 {
		artifactType := v
		obj, err = ExpandArtifact(artifactType, p)
		if err != nil {
			return nil, err
		}

	}

	// XXX Debug
	s1 := spew.Sprintf("%+v", obj)
	log.Println("ExpandArtifactSpec obj", s1)

	return obj, nil
}

// Flatten

func flattenValuesRef(at *artifactTranspose, p []interface{}) []interface{} {
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
		obj["values_paths"] = flattenFiles(at.Artifact.ValuesRef.ValuesPaths)
	}

	return []interface{}{obj}
}

// FlattenArtifact ArtifactSpec to TF State
func FlattenArtifact(at *artifactTranspose, p []interface{}) ([]interface{}, error) {
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
		obj["values_paths"] = flattenFiles(at.Artifact.ValuesPaths)
	}

	if len(at.Artifact.ChartName) > 0 {
		obj["chart_name"] = at.Artifact.ChartName
	}

	if len(at.Artifact.ChartVersion) > 0 {
		obj["chart_version"] = at.Artifact.ChartVersion
	}

	if at.Artifact.Paths != nil {
		obj["paths"] = flattenFiles(at.Artifact.Paths)
	}

	if at.Artifact.Configmap != nil {
		obj["configmap"] = flattenFile(at.Artifact.Configmap)
	}

	if at.Artifact.Secret != nil {
		obj["secret"] = flattenFile(at.Artifact.Secret)
	}

	if at.Artifact.Configuration != nil {
		obj["configuration"] = flattenFile(at.Artifact.Configuration)
	}

	if at.Artifact.Statefulset != nil {
		obj["statefulset"] = flattenFile(at.Artifact.Statefulset)
	}

	v, ok := obj["values_ref"].([]interface{})
	if !ok {
		v = []interface{}{}
	}
	if at.Artifact.ValuesRef.Repository != "" {
		obj["values_ref"] = flattenValuesRef(at, v)
	}

	s1 := spew.Sprintf("%+v", obj)
	log.Println("FlattenArtifact obj", s1)

	return []interface{}{obj}, nil
}

// FlattenArtifactOptions ArtifactSpec to TF State
func FlattenArtifactOptions(at *artifactTranspose, p []interface{}) ([]interface{}, error) {
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	obj["atomic"] = at.Options.Atomic
	obj["wait"] = at.Options.Wait
	obj["wait_for_jobs"] = at.Options.WaitForJobs
	obj["force"] = at.Options.Force
	obj["no_hooks"] = at.Options.NoHooks
	obj["max_history"] = at.Options.MaxHistory
	obj["render_sub_chart_notes"] = at.Options.RenderSubChartNotes
	obj["reset_values"] = at.Options.ResetValues
	obj["reuse_values"] = at.Options.ReuseValues
	if len(at.Options.SetString) > 0 {
		obj["set_string"] = toArrayInterface(at.Options.SetString)
	}
	obj["skip_crd"] = at.Options.SkipCRDs
	if len(at.Options.Timeout) > 0 {
		obj["timeout"] = at.Options.Timeout
	}
	obj["clean_up_on_fail"] = at.Options.CleanUpOnFail
	if len(at.Options.Description) > 0 {
		obj["description"] = at.Options.Description
	}
	obj["disable_open_api_validation"] = at.Options.DisableOpenAPIValidation
	obj["keep_history"] = at.Options.KeepHistory

	return []interface{}{obj}, nil
}

// FlattenArtifactSpec ArtifactSpec to TF State
func FlattenArtifactSpec(dataResource bool, in *commonpb.ArtifactSpec, p []interface{}) ([]interface{}, error) {
	if in == nil {
		log.Println("FlattenArtifactSpec empty input")
		return nil, fmt.Errorf("%s", "FlattenArtifactSpec empty input")
	}

	// XXX Debug
	// ob := spew.Sprintf("%+v", p)
	// log.Println("FlattenArtifactSpec p", ob)

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.Type) > 0 {
		obj["type"] = in.Type
	}

	jsonBytes, err := in.MarshalJSON()
	if err != nil {
		log.Println("FlattenArtifactSpec MarshalJSON error", err)
		return nil, fmt.Errorf("%s %+v", "FlattenArtifactSpec MarshalJSON error", err)
	}

	at := artifactTranspose{}
	err = json.Unmarshal(jsonBytes, &at)
	if err != nil {
		return nil, fmt.Errorf("%s %+v", "FlattenArtifactSpec json unmarshal error", err)
	}

	// XXX Debug
	log.Println("FlattenArtifactSpec jsonBytes:", string(jsonBytes))
	s1 := spew.Sprintf("%+v", at)
	log.Println("FlattenArtifactSpec at", s1)

	v, ok := obj["artifact"].([]interface{})
	if !ok {
		v = []interface{}{}
	}
	artfct, err := FlattenArtifact(&at, v)
	if dataResource && err == nil {
		obj["artifact"] = artfct
	}

	v, ok = obj["options"].([]interface{})
	if !ok {
		v = []interface{}{}
	}
	artfctOpts, err := FlattenArtifactOptions(&at, v)
	if dataResource && err == nil {
		obj["options"] = artfctOpts
	}

	// XXX Debug
	// ob = spew.Sprintf("%+v", p)
	// log.Println("FlattenArtifactSpec obj", ob)

	return []interface{}{obj}, nil
}
