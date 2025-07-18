package rafay

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/RafaySystems/rafay-common/proto/types/hub/commonpb"
	"github.com/RafaySystems/rafay-common/proto/types/hub/commonpb/datatypes"
	"github.com/RafaySystems/rafay-common/proto/types/hub/eaaspb"
	"github.com/RafaySystems/rafay-common/proto/types/hub/gitopspb"
	"github.com/RafaySystems/rafay-common/proto/types/hub/infrapb"
	"github.com/RafaySystems/rafay-common/proto/types/hub/integrationspb"
	"github.com/RafaySystems/rctl/pkg/config"
	"github.com/RafaySystems/rctl/utils"
	"github.com/davecgh/go-spew/spew"
	"github.com/go-yaml/yaml"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"google.golang.org/protobuf/types/known/structpb"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/validation"
)

type File struct {
	Name      string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Sensitive bool   `protobuf:"bytes,1,opt,name=sensitive,proto3" json:"sensitive,omitempty"`
	Data      []byte `protobuf:"bytes,2,opt,name=data,proto3" json:"data,omitempty"`
}

type HubError struct {
	Internal string `json:"internal"`
	Code     int    `json:"code"`
	External string `json:"external"`
}

/*
type UploadedYAMLArtifact struct {
	Paths []*File `protobuf:"bytes,1,rep,name=paths,proto3" json:"paths,omitempty"`
}

type UploadedHelmArtifact struct {
	ChartPath   *File   `protobuf:"bytes,1,opt,name=chartPath,proto3" json:"chartPath,omitempty"`
	ValuesPaths []*File `protobuf:"bytes,2,rep,name=valuesPaths,proto3" json:"valuesPaths,omitempty"`
}

type YAMLInGitRepo struct {
	Repository string  `protobuf:"bytes,1,opt,name=repository,proto3" json:"repository,omitempty"`
	Revision   string  `protobuf:"bytes,2,opt,name=revision,proto3" json:"revision,omitempty"`
	Paths      []*File `protobuf:"bytes,3,rep,name=paths,proto3" json:"paths,omitempty"`
}

type HelmInGitRepo struct {
	Repository  string  `protobuf:"bytes,1,opt,name=repository,proto3" json:"repository,omitempty"`
	Revision    string  `protobuf:"bytes,2,opt,name=revision,proto3" json:"revision,omitempty"`
	ChartPath   *File   `protobuf:"bytes,3,opt,name=chartPath,proto3" json:"chartPath,omitempty"`
	ValuesPaths []*File `protobuf:"bytes,4,rep,name=valuesPaths,proto3" json:"valuesPaths,omitempty"`
}

type HelmInHelmRepo struct {
	Repository   string  `protobuf:"bytes,1,opt,name=repository,proto3" json:"repository,omitempty"`
	ChartName    string  `protobuf:"bytes,2,opt,name=chartName,proto3" json:"chartName,omitempty"`
	ChartVersion string  `protobuf:"bytes,3,opt,name=chartVersion,proto3" json:"chartVersion,omitempty"`
	ValuesPaths  []*File `protobuf:"bytes,4,rep,name=valuesPaths,proto3" json:"valuesPaths,omitempty"`
}

type ManagedAlertManager struct {
	Configmap     *File `protobuf:"bytes,1,opt,name=configmap,proto3" json:"configmap,omitempty"`
	Secret        *File `protobuf:"bytes,2,opt,name=secret,proto3" json:"secret,omitempty"`
	Configuration *File `protobuf:"bytes,3,opt,name=configuration,proto3" json:"configuration,omitempty"`
	Statefulset   *File `protobuf:"bytes,4,opt,name=statefulset,proto3" json:"statefulset,omitempty"`
}
*/

func toArrayString(in []interface{}) []string {
	out := make([]string, len(in))
	for i, v := range in {
		if v == nil {
			out[i] = ""
			continue
		}
		out[i] = v.(string)
	}
	return out
}

func toArrayInt(in []interface{}) []int {
	out := make([]int, len(in))
	for i, v := range in {
		if v == nil {
			out[i] = 0
			continue
		}
		out[i] = v.(int)
	}
	return out
}

func toArrayInt32(in []interface{}) []int32 {
	out := make([]int32, len(in))
	for i, v := range in {
		if v == nil {
			out[i] = 0
			continue
		}
		nv := v.(int)
		out[i] = int32(nv)
	}
	return out
}

func toArrayStringSorted(in []interface{}) []string {
	if in == nil {
		return nil
	}
	out := toArrayString(in)
	sort.Strings(out)
	return out
}

func toArrayInterface(in []string) []interface{} {
	out := make([]interface{}, len(in))
	for i, v := range in {
		out[i] = v
	}
	return out
}

func intArraytoInterfaceArray(in []int) []interface{} {
	out := make([]interface{}, len(in))
	for i, v := range in {
		out[i] = v
	}
	return out
}

func toArrayInterfaceSorted(in []string) []interface{} {
	if in == nil {
		return nil
	}
	sort.Strings(in)
	out := toArrayInterface(in)
	return out
}

func toMapString(in map[string]interface{}) map[string]string {
	out := make(map[string]string)
	for i, v := range in {
		if v == nil {
			out[i] = ""
			continue
		}
		out[i] = v.(string)
	}
	return out
}

func toMapEmptyObject(in map[string]interface{}) map[string]interface{} {
	type x struct{}
	out := make(map[string]interface{})
	for i, v := range in {
		if v == nil {
			out[i] = ""
			continue
		}
		out[i] = x{}
	}
	return out
}

func toMapBool(in map[string]interface{}) map[string]bool {
	out := make(map[string]bool)
	for i, v := range in {
		if v == nil {
			out[i] = false
			continue
		}
		out[i] = v.(bool)
	}
	return out
}

func toMapByte(in map[string]interface{}) map[string][]byte {
	out := make(map[string][]byte)
	for i, v := range in {
		if v == nil {
			out[i] = []byte{}
			continue
		}
		if vstr, ok := v.(string); ok {
			out[i] = []byte(vstr)
		}
	}
	return out
}

func toMapByteInterface(in map[string][]byte) map[string]interface{} {
	out := make(map[string]interface{})
	for i, v := range in {
		if v == nil {
			out[i] = []byte{}
			continue
		}
		out[i] = bytes.NewBuffer(v).String()
	}
	return out
}

func toMapInterface(in map[string]string) map[string]interface{} {
	out := make(map[string]interface{})
	for i, v := range in {
		out[i] = v
	}
	return out
}

func toMapInterfaceObject(in map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{})
	log.Println("toMapInterfaceObject:", in)
	for i, v := range in {
		log.Println("toMapInterfaceObject v :", v)
		out[i] = "{}"
	}
	log.Println("toMapInterfaceObject: out:", out)
	return out
}

func toMapBoolInterface(in map[string]bool) map[string]interface{} {
	out := make(map[string]interface{})
	for i, v := range in {
		out[i] = v
	}
	return out
}

// Expanders

func expandMetaData(p []interface{}) *commonpb.Metadata {
	obj := &commonpb.Metadata{}
	if p == nil || len(p) == 0 || p[0] == nil {
		return obj
	}

	in := p[0].(map[string]interface{})
	if v, ok := in["name"].(string); ok && len(v) > 0 {
		obj.Name = v
	}
	if v, ok := in["description"].(string); ok && len(v) > 0 {
		obj.Description = v
	}
	if v, ok := in["project"].(string); ok && len(v) > 0 {
		obj.Project = v
	}
	if v, ok := in["projectID"].(string); ok && len(v) > 0 {
		obj.ProjectID = v
	}
	if v, ok := in["id"].(string); ok && len(v) > 0 {
		obj.Id = v
	}

	if v, ok := in["labels"].(map[string]interface{}); ok && len(v) > 0 {
		obj.Labels = toMapString(v)
	} else {
		obj.Labels = nil
	}

	log.Println("expandMetaData")
	if v, ok := in["annotations"].(map[string]interface{}); ok && len(v) > 0 {
		w1 := spew.Sprintf("%+v", v)
		log.Println("expandMetaData annotations ", w1)
		obj.Annotations = toMapString(v)
	}
	return obj
}

func expandDrift(p []interface{}) *commonpb.DriftSpec {
	obj := &commonpb.DriftSpec{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["enabled"].(bool); ok {
		obj.Enabled = v
	}

	if v, ok := in["action"].(string); ok && len(v) > 0 {
		obj.Action = v
	}

	return obj
}

func expandDriftWebhook(p []interface{}) *infrapb.DriftWebhook {
	obj := &infrapb.DriftWebhook{}
	if len(p) == 0 || p[0] == nil {
		return nil
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["enabled"].(bool); ok {
		obj.Enabled = v
		return obj
	}

	return nil
}

func expandPlacementLabels(p []interface{}) []*commonpb.PlacementLabel {
	if len(p) == 0 || p[0] == nil {
		return nil
	}

	obj := make([]*commonpb.PlacementLabel, len(p))
	for i := range p {
		in := p[i].(map[string]interface{})
		label := commonpb.PlacementLabel{}

		if v, ok := in["key"].(string); ok {
			label.Key = v
		}
		if v, ok := in["value"].(string); ok {
			label.Value = v
		}
		obj[i] = &label
	}

	return obj
}

func expandPlacement(p []interface{}) *commonpb.PlacementSpec {
	obj := &commonpb.PlacementSpec{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}

	in := p[0].(map[string]interface{})
	if v, ok := in["selector"].(string); ok && len(v) > 0 {
		obj.Selector = v
	}

	if v, ok := in["labels"].([]interface{}); ok {
		obj.Labels = expandPlacementLabels(v)
	}

	if v, ok := in["environment"].([]any); ok {
		obj.Environment = expandEnvironmentPlacement(v)
	}

	return obj
}

func expandEnvironmentPlacement(p []any) *commonpb.Environment {
	if len(p) == 0 || p[0] == nil {
		return nil
	}
	obj := &commonpb.Environment{}
	in := p[0].(map[string]any)
	if v, ok := in["name"].(string); ok && len(v) > 0 {
		obj.Name = v
	}
	return obj
}

func expandOverridesRepo(p []interface{}) (*gitopspb.OverrideTemplate_Repo, error) {
	obj := gitopspb.OverrideTemplate_Repo{}
	obj.Repo = &gitopspb.RepoOverrideTemplate{}

	if len(p) == 0 || p[0] == nil {
		return &obj, fmt.Errorf("%s", "expandOverridesRepo empty input")
	}

	in := p[0].(map[string]interface{})

	log.Println("expandOverridesRepo")

	if v, ok := in["repository"].(string); ok && len(v) > 0 {
		obj.Repo.Repository = v
	}

	if v, ok := in["revision"].(string); ok && len(v) > 0 {
		obj.Repo.Revision = v
	}

	if v, ok := in["revision"].(string); ok && len(v) > 0 {
		obj.Repo.Revision = v
	}

	if v, ok := in["paths"].([]interface{}); ok && len(v) > 0 {
		obj.Repo.Paths = expandCommonpbFiles(v)
	}

	return &obj, nil
}

func expandOverridesInline(p []interface{}) (*gitopspb.OverrideTemplate_Inline, error) {
	obj := gitopspb.OverrideTemplate_Inline{}
	obj.Inline = &gitopspb.InlineOverrideTemplate{}

	if len(p) == 0 || p[0] == nil {
		return &obj, fmt.Errorf("%s", "expandOverridesRepo empty input")
	}

	in := p[0].(map[string]interface{})

	log.Println("expandOverridesRepo")

	if v, ok := in["inline"].(string); ok && len(v) > 0 {
		obj.Inline.Inline = v
	}

	return &obj, nil
}

func expandOverrides(p []interface{}) []*gitopspb.OverrideTemplate {
	if len(p) == 0 || p[0] == nil {
		return []*gitopspb.OverrideTemplate{}
	}

	out := make([]*gitopspb.OverrideTemplate, len(p))
	for i := range p {
		obj := gitopspb.OverrideTemplate{}
		in := p[i].(map[string]interface{})

		if v, ok := in["type"].(string); ok && len(v) > 0 {
			obj.Type = v
		}

		if v, ok := in["weight"].(int); ok {
			obj.Weight = int32(v)
		}

		if vp, ok := in["template"].([]interface{}); ok && len(vp) > 0 {
			if len(vp) == 0 || vp[0] == nil {
				return nil
			}
			in := vp[0].(map[string]interface{})
			if v, ok := in["inline"].(string); ok && len(v) > 0 {
				obj.Template, _ = expandOverridesInline(vp)
			} else {
				obj.Template, _ = expandOverridesRepo(vp)
			}
		}

		out[i] = &obj
	}

	return out
}

func expandAgents(p []interface{}) []*integrationspb.AgentMeta {
	if len(p) == 0 || p[0] == nil {
		return []*integrationspb.AgentMeta{}
	}

	out := make([]*integrationspb.AgentMeta, len(p))

	for i := range p {
		obj := &integrationspb.AgentMeta{}
		in := p[i].(map[string]interface{})

		if v, ok := in["name"].(string); ok && len(v) > 0 {
			obj.Name = v
		}

		if v, ok := in["id"].(string); ok && len(v) > 0 {
			obj.Id = v
		}
		out[i] = obj
	}

	return out
}

func expandFile(p []interface{}) *File {
	obj := File{}
	if len(p) == 0 || p[0] == nil {
		return nil
	}

	in := p[0].(map[string]interface{})
	if v, ok := in["name"].(string); ok && len(v) > 0 {
		obj.Name = v
	}

	if v, ok := in["sensitive"].(bool); ok {
		obj.Sensitive = v
	}

	if strings.HasPrefix(obj.Name, "file://") {
		//get full path of artifact
		artifactFullPath := filepath.Join(filepath.Dir("."), obj.Name[7:])
		//retrieve artifact data
		artifactData, err := ioutil.ReadFile(artifactFullPath)
		if err != nil {
			log.Println("unable to read artifact at ", artifactFullPath)
		} else {
			obj.Data = artifactData
		}
	}

	return &obj
}

func expandCommonpbFile(p []interface{}) *commonpb.File {
	obj := commonpb.File{}
	if len(p) == 0 || p[0] == nil {
		return nil
	}

	in := p[0].(map[string]interface{})
	if v, ok := in["name"].(string); ok && len(v) > 0 {
		obj.Name = v
	}

	if strings.HasPrefix(obj.Name, "file://") {
		//get full path of artifact
		artifactFullPath := filepath.Join(filepath.Dir("."), obj.Name[7:])
		//retrieve artifact data
		artifactData, err := ioutil.ReadFile(artifactFullPath)
		if err != nil {
			log.Println("unable to read artifact at ", artifactFullPath)
		} else {
			obj.Data = artifactData
		}
	}

	return &obj
}

func expandCommonpbFiles(p []interface{}) []*commonpb.File {
	if len(p) == 0 || p[0] == nil {
		return nil
	}
	out := make([]*commonpb.File, len(p))

	for i := range p {
		obj := commonpb.File{}
		in := p[i].(map[string]interface{})

		if name, ok := in["name"].(string); ok && len(name) > 0 {

			if strings.HasPrefix(name, "file://") {
				//get full path of artifact
				artifactFullPath := filepath.Join(filepath.Dir("."), name[7:])
				//retrieve artifact data
				artifactData, err := ioutil.ReadFile(artifactFullPath)
				if err != nil {
					log.Println("unable to read artifact at ", artifactFullPath)
				} else {
					obj.Data = artifactData
				}
				obj.Name = strings.TrimPrefix(name, "file://")
			} else {
				obj.Name = name
				if data, ok := in["data"].(string); ok && len(data) > 0 {
					obj.Data = []byte(data)
				}
			}
		}

		if mp, ok := in["mount_path"].(string); ok && len(mp) > 0 {
			obj.MountPath = mp
		}

		if v, ok := in["sensitive"].(bool); ok {
			obj.Sensitive = v
		}

		if v, ok := in["options"].([]interface{}); ok && len(v) > 0 {
			obj.Options = expandFileOptions(v)
		}

		out[i] = &obj
	}

	return out
}

func expandFiles(p []interface{}) ([]*File, error) {
	if len(p) == 0 || p[0] == nil {
		return nil, fmt.Errorf("%s", "expandFiles empty input")
	}

	obj := make([]*File, len(p))
	for i := range p {
		of := File{}
		in := p[i].(map[string]interface{})
		if v, ok := in["name"].(string); ok && len(v) > 0 {
			of.Name = v
		}

		if strings.HasPrefix(of.Name, "file://") {
			//get full path of artifact
			artifactFullPath := filepath.Join(filepath.Dir("."), of.Name[7:])
			//retrieve artifact data
			artifactData, err := ioutil.ReadFile(artifactFullPath)
			if err != nil {
				log.Println("unable to read artifact at ", artifactFullPath)
				return nil, err
			} else {
				of.Data = artifactData
			}
		} else if strings.HasPrefix(of.Name, "temp://") {
			//get full path of artifact
			artifactFullPath := filepath.Join(filepath.Dir("."), of.Name[7:])
			//retrieve artifact data
			artifactData, err := ioutil.ReadFile(artifactFullPath)
			if err != nil {
				log.Println("unable to read artifact at ", artifactFullPath)
			}
			of.Data = artifactData
		}

		obj[i] = &of
	}
	return obj, nil
}

func expandQuantity(p []interface{}) *resource.Quantity {
	if len(p) == 0 || p[0] == nil {
		return nil
	}
	in := p[0].(map[string]interface{})
	if v, ok := in["string"].(string); ok {
		log.Println("string v", v)
		ob, err := resource.ParseQuantity(v)
		if err == nil {
			log.Println("string v error", err, " ob ", ob)
			return &ob
		}
		log.Println("string v error", err)
	}

	return nil
}

func expandQuantityString(str string) *resource.Quantity {
	if len(str) == 0 {
		return nil
	}
	ob, err := resource.ParseQuantity(str)
	if err == nil {
		log.Println("string v error", err, " ob ", ob)
		return &ob
	}
	log.Println("string v error", err)

	return nil
}

// func expandResourceQuantity(p []interface{}) *commonpb.ResourceQuantity {
// 	obj := commonpb.ResourceQuantity{}
// 	if len(p) == 0 || p[0] == nil {
// 		log.Println("expandResourceQuantity empty input")
// 		return &obj
// 	}
// 	in := p[0].(map[string]interface{})
// 	if v, ok := in["memory"].([]interface{}); ok {
// 		obj.Memory = expandQuantity(v)
// 		log.Println("expandResourceQuantity memory", obj.Memory)
// 	}

// 	if v, ok := in["cpu"].([]interface{}); ok {
// 		obj.Cpu = expandQuantity(v)
// 		log.Println("expandResourceQuantity CPU", obj.Cpu)
// 	}

// 	log.Println("expandResourceQuantity obj", obj)
// 	return &obj
// }

func expandResourceQuantityString(p []interface{}) *commonpb.ResourceQuantity {
	obj := commonpb.ResourceQuantity{}
	if len(p) == 0 || p[0] == nil {
		log.Println("expandResourceQuantity empty input")
		return &obj
	}
	in := p[0].(map[string]interface{})
	if v, ok := in["memory"].(string); ok {
		obj.Memory = expandQuantity1140(v)
		//obj.Memory = v
		log.Println("expandResourceQuantity memory", obj.Memory)
	}

	if v, ok := in["cpu"].(string); ok {
		obj.Cpu = expandQuantity1140(v)
		//obj.Cpu = v
		log.Println("expandResourceQuantity CPU", obj.Cpu)
	}

	log.Println("expandResourceQuantity obj", obj)
	return &obj
}

func expandProjectMeta(p []interface{}) []*commonpb.ProjectMeta {
	if len(p) == 0 {
		return []*commonpb.ProjectMeta{}
	}
	var sortByName []string
	out := make([]*commonpb.ProjectMeta, len(p))
	for i := range p {
		if p[i] == nil {
			continue
		}
		in := p[i].(map[string]interface{})
		obj := commonpb.ProjectMeta{}

		if v, ok := in["name"].(string); ok && len(v) > 0 {
			obj.Name = v
			sortByName = append(sortByName, v)
		}
		if v, ok := in["id"].(string); ok && len(v) > 0 {
			obj.Id = v
		}

		out[i] = &obj
	}

	var sortedOut []*commonpb.ProjectMeta
	for _, name := range sortByName {
		for _, val := range out {
			if name == val.Name {
				sortedOut = append(sortedOut, val)
			}
		}
	}

	log.Println("expandProjectMeta out", sortedOut)
	return sortedOut
}

func expandProjectMetaV3(p []interface{}) []*infrapb.Projects {
	if len(p) == 0 {
		return []*infrapb.Projects{}
	}
	var sortByName []string
	out := make([]*infrapb.Projects, len(p))
	for i := range p {
		in := p[i].(map[string]interface{})
		obj := infrapb.Projects{}

		if v, ok := in["name"].(string); ok && len(v) > 0 {
			obj.Name = v
			sortByName = append(sortByName, v)
		}

		out[i] = &obj
	}

	var sortedOut []*infrapb.Projects
	for _, name := range sortByName {
		for _, val := range out {
			if name == val.Name {
				sortedOut = append(sortedOut, val)
			}
		}
	}

	log.Println("expandProjectMeta out", sortedOut)
	return sortedOut
}

func expandSharingSpec(p []interface{}) *commonpb.SharingSpec {
	obj := commonpb.SharingSpec{}
	if len(p) == 0 || p[0] == nil {
		return &obj
	}

	in := p[0].(map[string]interface{})
	if v, ok := in["enabled"].(bool); ok {
		obj.Enabled = v
	}

	if v, ok := in["projects"].([]interface{}); ok && len(v) > 0 {
		obj.Projects = expandProjectMeta(v)
	} else if v, ok := in["projects"].(*schema.Set); ok && v != nil && v.Len() > 0 {
		obj.Projects = expandProjectMeta(v.List())
	}

	log.Println("expandSharingSpec obj", obj)
	return &obj
}

func expandSharingSpecV3(p []interface{}) *infrapb.Sharing {
	obj := infrapb.Sharing{}
	if len(p) == 0 || p[0] == nil {
		return &obj
	}

	in := p[0].(map[string]interface{})
	if v, ok := in["enabled"].(bool); ok {
		obj.Enabled = v
	}

	if v, ok := in["projects"].(*schema.Set); ok && v.Len() > 0 {
		obj.Projects = expandProjectMetaV3(v.List())
	}

	log.Println("expandSharingSpec obj", obj)
	return &obj
}

// Flatteners

func flattenMetaData(in *commonpb.Metadata) []interface{} {
	if in == nil {
		return nil
	}
	log.Println("flatten metadata: ", in)
	obj := make(map[string]interface{})

	if len(in.Name) > 0 {
		obj["name"] = in.Name
	}

	if len(in.Description) > 0 {
		obj["description"] = in.Description
	}

	if len(in.Project) > 0 {
		obj["project"] = in.Project
	}

	if len(in.ProjectID) > 0 {
		obj["projectID"] = in.ProjectID
	}

	if len(in.Id) > 0 {
		obj["id"] = in.Id
	}

	if len(in.Labels) > 0 {
		obj["labels"] = toMapInterface(in.Labels)
	}

	if len(in.Annotations) > 0 {
		obj["annotations"] = toMapInterface(in.Annotations)
	}

	return []interface{}{obj}
}

func flattenPlacementLabels(input []*commonpb.PlacementLabel) []interface{} {
	if input == nil {
		return nil
	}

	out := make([]interface{}, len(input))
	for i, in := range input {
		obj := map[string]interface{}{}

		if len(in.Key) > 0 {
			obj["key"] = in.Key
		}
		if len(in.Value) > 0 {
			obj["value"] = in.Value
		}
		out[i] = obj
	}

	return out
}

func flattenPlacement(in *commonpb.PlacementSpec) []interface{} {
	if in == nil {
		return nil
	}

	obj := make(map[string]interface{})
	if len(in.Labels) > 0 {
		obj["labels"] = flattenPlacementLabels(in.Labels)
	}

	if len(in.Selector) > 0 {
		obj["selector"] = in.Selector
	}

	if in.Environment != nil {
		obj["environment"] = flattenEnvironmentPlacement(in.Environment)
	}

	return []interface{}{obj}
}

func flattenEnvironmentPlacement(in *commonpb.Environment) []any {
	if in == nil {
		return nil
	}

	obj := make(map[string]any)
	if len(in.Name) > 0 {
		obj["name"] = in.Name
	}

	return []any{obj}
}

func flattenFile(in *File) []interface{} {
	if in == nil {
		return nil
	}

	obj := make(map[string]interface{})
	if len(in.Name) > 0 {
		obj["name"] = in.Name
	}
	obj["sensitive"] = in.Sensitive
	return []interface{}{obj}
}

func flattenCommonpbFile(in *commonpb.File) []interface{} {
	if in == nil {
		return nil
	}

	obj := make(map[string]interface{})
	if len(in.Name) > 0 {
		obj["name"] = in.Name
	}
	obj["sensitive"] = in.Sensitive
	return []interface{}{obj}
}

func flattenFiles(input []*File) []interface{} {
	if input == nil {
		return nil
	}

	out := make([]interface{}, len(input))
	for i, in := range input {
		obj := map[string]interface{}{}
		if len(in.Name) > 0 {
			obj["name"] = in.Name
		}
		out[i] = obj
	}

	return out
}

func flattenCommonpbFiles(input []*commonpb.File) []interface{} {
	if input == nil {
		return nil
	}

	out := make([]interface{}, len(input))
	for i, in := range input {
		obj := map[string]interface{}{}
		obj["name"] = in.Name
		obj["sensitive"] = in.Sensitive
		if !in.Sensitive {
			obj["data"] = string(in.Data)
		}
		obj["mount_path"] = in.MountPath
		obj["options"] = flattenFileOptions(in.Options)

		out[i] = obj
	}

	return out
}

func flattenRatio(in *commonpb.ResourceRatio) []interface{} {
	if in == nil {
		return nil
	}
	//log.Println("flattenRatio ", in)
	obj := make(map[string]interface{})
	obj["memory"] = in.Memory
	obj["cpu"] = in.Cpu

	return []interface{}{obj}
}

func flattenProjectMeta(input []*commonpb.ProjectMeta, includeProjectId bool) []interface{} {
	if input == nil {
		return nil
	}

	out := make([]interface{}, len(input))
	for i, in := range input {
		obj := map[string]interface{}{}
		if len(in.Name) > 0 {
			obj["name"] = in.Name
		}
		if includeProjectId && len(in.Id) > 0 {
			obj["id"] = in.Id
		}
		out[i] = obj
	}

	return out
}

func flattenProjectMetaV3(input []*infrapb.Projects) []interface{} {
	if input == nil {
		return nil
	}

	out := make([]interface{}, len(input))
	for i, in := range input {
		obj := map[string]interface{}{}
		if len(in.Name) > 0 {
			obj["name"] = in.Name
		}
		out[i] = obj
	}

	return out
}

func flattenSharingSpec(in *commonpb.SharingSpec) []interface{} {
	if in == nil {
		return nil
	}

	obj := make(map[string]interface{})
	obj["enabled"] = in.Enabled
	if len(in.Projects) > 0 {
		obj["projects"] = flattenProjectMeta(in.Projects, false)
	}

	return []interface{}{obj}
}

func flattenSharingSpecV3(in *infrapb.Sharing) []interface{} {
	if in == nil {
		return nil
	}

	obj := make(map[string]interface{})
	obj["enabled"] = in.Enabled
	if len(in.Projects) > 0 {
		obj["projects"] = flattenProjectMetaV3(in.Projects)
	}

	return []interface{}{obj}
}

func flattenDrift(in *commonpb.DriftSpec) []interface{} {
	if in == nil {
		return nil
	}

	obj := make(map[string]interface{})

	obj["enabled"] = in.Enabled

	if len(in.Action) > 0 {
		obj["action"] = in.Action
	}

	return []interface{}{obj}
}

// Cluster Spec file processing
type configMetadata struct {
	Name    string `yaml:"name"`
	Project string `yaml:"project"`
	Version string `yaml:"version"`
}

type configResourceType struct {
	Meta *configMetadata `yaml:"metadata"`
}

func findResourceNameFromConfig(configBytes []byte) (string, string, string, error) {
	var config configResourceType
	if err := yaml.Unmarshal(configBytes, &config); err != nil {
		return "", "", "", nil
	} else if config.Meta == nil {
		return "", "", "", fmt.Errorf("%s", "Invalid resource: No metadata found")
	} else if config.Meta.Name == "" {
		return "", "", "", fmt.Errorf("%s", "Invalid resource: No name specified in metadata")
	}
	return config.Meta.Name, config.Meta.Project, config.Meta.Version, nil
}

func collateConfigsByName(rafayConfigs, clusterConfigs [][]byte) (map[string][]byte, []error) {
	var errs []error
	configsMap := make(map[string][][]byte)
	// First find all rafay spec configurations
	for _, config := range rafayConfigs {
		name, _, _, err := findResourceNameFromConfig(config)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		if _, ok := configsMap[name]; ok {
			errs = append(errs, fmt.Errorf(`duplicate "cluster" resource with name "%s" found`, name))
			continue
		}
		configsMap[name] = append(configsMap[name], config)
	}
	// Then append the cluster specific configurations
	for _, config := range clusterConfigs {
		name, _, _, err := findResourceNameFromConfig(config)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		if _, ok := configsMap[name]; !ok {
			errs = append(errs, fmt.Errorf(`error finding "Cluster" configuration for name "%s"`, name))
			continue
		}
		configsMap[name] = append(configsMap[name], config)
	}
	// Remove any configs that don't have the tail end (cluster related configs)
	result := make(map[string][]byte)
	for name, configs := range configsMap {
		if len(configs) <= 0 {
			errs = append(errs, fmt.Errorf(`no "ClusterConfig" found for cluster "%s"`, name))
			continue
		}
		collatedConfigBytes, err := utils.JoinYAML(configs)
		if err != nil {
			errs = append(errs, fmt.Errorf(`error collating YAML files for cluster "%s": %s`, name, err))
			continue
		}
		result[name] = collatedConfigBytes
		log.Printf(`final Configuration for cluster "%s": %#v`, name, string(collatedConfigBytes))
	}
	return result, errs
}

func expandResourceQuantity1170(p []interface{}) *commonpb.ResourceQuantity {
	obj := commonpb.ResourceQuantity{}
	if len(p) == 0 || p[0] == nil {
		log.Println("expandResourceQuantity1170 empty input")
		return &obj
	}
	in := p[0].(map[string]interface{})
	if v, ok := in["memory"].(string); ok {
		obj.Memory = expandQuantity1170(v).String()
		//obj.Memory = v
		log.Println("expandResourceQuantity1170 memory", obj.Memory)
	}

	if v, ok := in["cpu"].(string); ok {
		obj.Cpu = expandQuantity1140(v)
		//obj.Cpu = v
		log.Println("expandResourceQuantity1170 CPU", obj.Cpu)
	}

	log.Println("expandResourceQuantity1170 obj", obj)
	return &obj
}

func expandQuantity1170(p string) *resource.Quantity {
	if p == "" {
		return nil
	}
	ob, err := resource.ParseQuantity(p)
	if err == nil {
		log.Println("expandQuantity1170 ob: ", ob)
		return &ob
	}
	log.Println("expandQuantity1170 error", err)
	return nil
}

func expandQuantity1140(p string) string {
	if p == "" {
		return ""
	}
	ob, err := resource.ParseQuantity(p)
	if err == nil {
		log.Println("expandQuantity1140 ob: ", ob)
		return ob.String()
	}
	log.Println("expandQuantity1140 error", err)
	return ""
}

func flattenResourceQuantity1170(in *commonpb.ResourceQuantity) []interface{} {
	if in == nil {
		return nil
	}

	obj := make(map[string]interface{})
	if in.Memory != "" {
		// for i := 0; i < 10; i++ {
		// 	in.GetMemory().Add(*in.GetMemory())
		// 	//log.Println("adding ", in.GetMemory().String())
		// }
		obj["memory"] = in.Memory
		//log.Println("flattenResourceQuantity1170 memory string ", in.GetMemory().String())
	}

	if in.Cpu != "" {
		// cq := *in.Cpu
		// for i := 0; i < 999; i++ {
		// 	in.GetCpu().Add(cq)
		// 	//log.Println("adding ", in.GetCpu().String())
		// }
		// in.GetCpu().RoundUp(resource.Micro)
		// obj["cpu"] = in.GetCpu().String()
		obj["cpu"] = in.Cpu
		//log.Println("flattenResourceQuantity1170 cpu string ", in.GetCpu().String(), " => ")
	}

	log.Println("flattenResourceQuantityV101 obj", obj)
	return []interface{}{obj}
}

func flattenResourceQuantity(in *commonpb.ResourceQuantity) []interface{} {
	if in == nil {
		return nil
	}

	obj := make(map[string]interface{})
	if in.Memory != "" {
		var m resource.QuantityValue
		m.Set(in.GetMemory())
		for i := 0; i < 10; i++ {
			m.Add(m.Quantity)
			//in.GetMemory().Add(*in.GetMemory())
			log.Println("adding ", m.String())
		}
		if strings.Contains(m.String(), "Gi") {
			val := m.String()[:len(m.String())-2]
			log.Println("flattenResourceQuantity adjust Gi ", m.String(), val)
			intVar, err := strconv.Atoi(val)
			if err == nil {
				intVar = 1024 * intVar
				val1 := strconv.Itoa(intVar)
				obj["memory"] = val1 + "Mi"
				log.Println("flattenResourceQuantity memory string ", val1+"Mi")
			} else {
				obj["memory"] = m.String()
			}
		} else {
			obj["memory"] = m.String()
			log.Println("flattenResourceQuantity memory string ", m.String())
		}
	}

	if in.Cpu != "" {
		var cp resource.QuantityValue
		cp.Set(in.GetCpu())
		cp1 := cp
		for i := 0; i < 999; i++ {
			cp.Add(cp1.Quantity)
			log.Println("adding ", cp.String())
		}
		cp.RoundUp(resource.Micro)
		obj["cpu"] = cp.String()
		log.Println("flattenResourceQuantity cpu string ", cp.String())
	}

	log.Println("flattenResourceQuantity obj", obj)
	return []interface{}{obj}
}

func ResetImpersonateUser() {
	log.Println("ResetImpersonateUser")
	config.ApiKey = ""
	config.ApiSecret = ""
	config.ResetOrigConfig()
}

func ReadMetaName(p []interface{}) string {
	if p == nil || len(p) == 0 || p[0] == nil {
		return ""
	}

	in := p[0].(map[string]interface{})
	if v, ok := in["name"].(string); ok && len(v) > 0 {
		return v
	}
	return ""
}

func GetMetaName(in *schema.ResourceData) string {
	if in == nil {
		return ""
	}

	if v, ok := in.Get("metadata").([]interface{}); ok {
		return ReadMetaName(v)
	}

	return ""
}

func ReadMeta(p []interface{}) *commonpb.Metadata {
	if p == nil || len(p) == 0 || p[0] == nil {
		return nil
	}

	var meta commonpb.Metadata

	in := p[0].(map[string]interface{})
	if v, ok := in["name"].(string); ok && len(v) > 0 {
		meta.Name = v
	}

	if v, ok := in["project"].(string); ok && len(v) > 0 {
		meta.Project = v
	}

	return &meta
}
func GetMetaData(in *schema.ResourceData) *commonpb.Metadata {
	if in == nil {
		return nil
	}

	if v, ok := in.Get("metadata").([]interface{}); ok {
		return ReadMeta(v)
	}

	return nil
}

func expandVariables(p []interface{}) []*eaaspb.Variable {
	if len(p) == 0 || p[0] == nil {
		return []*eaaspb.Variable{}
	}
	log.Println("expand variables start")
	vars := make([]*eaaspb.Variable, len(p))

	for i := range p {
		obj := eaaspb.Variable{}
		in := p[i].(map[string]interface{})

		if v, ok := in["name"].(string); ok && len(v) > 0 {
			obj.Name = v
		}

		if v, ok := in["value_type"].(string); ok && len(v) > 0 {
			obj.ValueType = v
		}

		if v, ok := in["value"].(string); ok && len(v) > 0 {
			obj.Value = v
		}

		if v, ok := in["options"].([]interface{}); ok && len(v) > 0 {
			obj.Options = expandVariableOptions(v)
		}

		vars[i] = &obj
	}

	return vars
}

func expandVariableOptions(p []interface{}) *eaaspb.VariableOptions {
	if len(p) == 0 || p[0] == nil {
		return nil
	}

	options := &eaaspb.VariableOptions{}
	opts := p[0].(map[string]interface{})

	if v, ok := opts["description"].(string); ok && len(v) > 0 {
		options.Description = v
	}

	if v, ok := opts["sensitive"].(bool); ok {
		options.Sensitive = v
	}

	if v, ok := opts["required"].(bool); ok {
		options.Required = v
	}

	if v, ok := opts["immutable"].(bool); ok {
		options.Immutable = v
	}

	if v, ok := opts["override"].([]interface{}); ok && len(v) > 0 {
		options.Override = expandVariableOverrideOptions(v)
	}
	if v, ok := opts["display_metadata"].(string); ok && len(v) > 0 {
		options.DisplayMetadata = expandDisplayMetadata(v)
	}

	if v, ok := opts["schema"].([]interface{}); ok && len(v) > 0 {
		options.Schema = expandCustomSchema(v)
	}

	return options

}

func expandDisplayMetadata(v string) *structpb.Struct {
	if v == "" {
		return nil
	}
	newMap := map[string]interface{}{}
	if err := json.Unmarshal([]byte(v), &newMap); err == nil {
		s, err := structpb.NewStruct(newMap)
		if err != nil {
			return nil
		}
		return s
	}
	return nil
}

func expandVariableOverrideOptions(p []interface{}) *eaaspb.VariableOverrideOptions {
	if len(p) == 0 || p[0] == nil {
		return nil
	}

	override := &eaaspb.VariableOverrideOptions{}
	in := p[0].(map[string]interface{})

	if v, ok := in["type"].(string); ok && len(v) > 0 {
		override.Type = v
	}

	if vals, ok := in["restricted_values"].([]interface{}); ok && len(vals) > 0 {
		override.RestrictedValues = toArrayString(vals)
	}

	if vals, ok := in["selectors"].([]interface{}); ok && len(vals) > 0 {
		override.Selectors = toArrayString(vals)
	}

	if do, ok := in["display_overridden"].(bool); ok {
		override.DisplayOverridden = do
	}

	if rkv, ok := in["restricted_key_values"].(map[string]any); ok && len(rkv) > 0 {
		override.RestrictedKeyValues = toMapString(rkv)
	}

	return override
}

func flattenVariables(input []*eaaspb.Variable, p []interface{}) []interface{} {
	log.Println("flatten variables start")
	if len(input) == 0 {
		return nil
	}

	out := make([]interface{}, len(input))
	for i, in := range input {
		log.Println("flatten variable ", in)
		obj := map[string]interface{}{}
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}
		obj["name"] = in.Name
		obj["value_type"] = in.ValueType
		obj["value"] = in.Value
		obj["options"] = flattenVariableOptions(in.Options)
		out[i] = &obj
	}

	return out
}

func flattenVariableOptions(input *eaaspb.VariableOptions) []interface{} {
	log.Println("flatten variable options")
	if input == nil {
		return nil
	}
	obj := map[string]any{
		"description":      input.Description,
		"sensitive":        input.Sensitive,
		"required":         input.Required,
		"immutable":        input.Immutable,
		"display_metadata": flattenDisplayMetadata(input.DisplayMetadata),
		"override":         flattenVariableOverrideOptions(input.Override),
		"schema":           flattenCustomSchema(input.Schema),
	}
	return []interface{}{obj}
}

func flattenVariableOverrideOptions(input *eaaspb.VariableOverrideOptions) []interface{} {
	log.Println("flatten variable override options")
	if input == nil {
		return nil
	}
	obj := map[string]any{
		"type":                  input.Type,
		"restricted_values":     toArrayInterface(input.RestrictedValues),
		"selectors":             toArrayInterface(input.Selectors),
		"display_overridden":    input.DisplayOverridden,
		"restricted_key_values": toMapInterface(input.RestrictedKeyValues),
	}
	return []interface{}{obj}
}

func expandEaasHooks(p []interface{}) ([]*eaaspb.Hook, error) {
	hooks := make([]*eaaspb.Hook, 0)
	if len(p) == 0 {
		return hooks, nil
	}

	for indx := range p {
		if p[indx] == nil {
			return nil, nil
		}
		hook := &eaaspb.Hook{}
		in := p[indx].(map[string]interface{})

		if n, ok := in["name"].(string); ok && len(n) > 0 {
			hook.Name = n
		}

		if d, ok := in["description"].(string); ok && len(d) > 0 {
			hook.Description = d
		}

		if t, ok := in["type"].(string); ok && len(t) > 0 {
			hook.Type = t
		}

		if ho, ok := in["options"].([]interface{}); ok {
			hook.Options = expandHookOptions(ho)
		}

		if ag, ok := in["agents"].([]interface{}); ok {
			hook.Agents = expandEaasAgents(ag)
		}

		if d, ok := in["timeout_seconds"].(int); ok {
			hook.TimeoutSeconds = int64(d)
		}

		if n, ok := in["on_failure"].(string); ok && len(n) > 0 {
			hook.OnFailure = n
		}

		var err error
		if n, ok := in["driver"].([]interface{}); ok && len(n) > 0 {
			log.Println("WARN: driver is deprecated, please use workflow_handler instead")
			hook.Driver, err = expandWorkflowHandlerCompoundRef(n)
			if err != nil {
				return nil, err
			}
		}

		if n, ok := in["workflow_handler"].([]interface{}); ok && len(n) > 0 {
			hook.WorkflowHandler, err = expandWorkflowHandlerCompoundRef(n)
			if err != nil {
				return nil, err
			}
		}

		if n, ok := in["depends_on"].([]interface{}); ok && len(n) > 0 {
			hook.DependsOn = toArrayString(n)
		}

		if d, ok := in["execute_once"].(bool); ok {
			hook.ExecuteOnce = d
		}

		if skip, ok := in["skip_config"].([]any); ok && len(skip) > 0 {
			hook.SkipConfig = expandSkipConfig(skip)
		}

		hooks = append(hooks, hook)
	}

	return hooks, nil
}

func expandSkipConfig(p []any) *eaaspb.SkipConfig {
	if len(p) == 0 || p[0] == nil {
		return nil
	}

	obj := &eaaspb.SkipConfig{}
	in := p[0].(map[string]any)

	if v, ok := in["condition"].(string); ok {
		obj.Condition = v
	}

	if v, ok := in["skip_on_destroy"].(bool); ok {
		obj.SkipOnDestroy = v
	}

	return obj
}

func expandHookOptions(p []interface{}) *eaaspb.HookOptions {
	if len(p) == 0 || p[0] == nil {
		return nil
	}

	in := p[0].(map[string]interface{})
	ho := &eaaspb.HookOptions{}

	if ao, ok := in["approval"].([]interface{}); ok && len(ao) > 0 {
		ho.Approval = expandApprovalOptions(ao)
	}

	if no, ok := in["notification"].([]interface{}); ok && len(no) > 0 {
		ho.Notification = expandNotificationOptions(no)
	}

	if so, ok := in["script"].([]interface{}); ok && len(so) > 0 {
		ho.Script = expandScriptOptions(so)
	}

	if co, ok := in["container"].([]interface{}); ok && len(co) > 0 {
		ho.Container = expandContainerOptions(co)
	}

	if o, ok := in["http"].([]interface{}); ok && len(o) > 0 {
		ho.Http = expandHttpOptions(o)
	}

	return ho
}

func expandApprovalOptions(p []interface{}) *eaaspb.ApprovalOptions {
	ao := &eaaspb.ApprovalOptions{}
	if len(p) == 0 || p[0] == nil {
		return ao
	}

	in := p[0].(map[string]interface{})

	if t, ok := in["type"].(string); ok && len(t) > 0 {
		ao.Type = t
	}

	if iao, ok := in["internal"].([]interface{}); ok && len(iao) > 0 {
		ao.Internal = expandInternalApprovalOptions(iao)
	}

	if eao, ok := in["email"].([]interface{}); ok && len(eao) > 0 {
		ao.Email = expandEmailApprovalOptions(eao)
	}

	if jao, ok := in["jira"].([]interface{}); ok && len(jao) > 0 {
		ao.Jira = expandJiraApprovalOptions(jao)
	}

	if ghao, ok := in["github_pull_request"].([]interface{}); ok && len(ghao) > 0 {
		ao.GithubPullRequest = expandGithubPRApprovalOptions(ghao)
	}

	return ao
}

func expandInternalApprovalOptions(p []interface{}) *eaaspb.InternalApprovalOptions {
	if len(p) == 0 || p[0] == nil {
		return nil
	}
	iao := &eaaspb.InternalApprovalOptions{}
	in := p[0].(map[string]interface{})

	if emails, ok := in["emails"].([]interface{}); ok && len(emails) > 0 {
		iao.Emails = toArrayString(emails)
	}

	return iao
}

func expandEmailApprovalOptions(p []interface{}) *eaaspb.EmailApprovalOptions {
	eao := &eaaspb.EmailApprovalOptions{}
	if len(p) == 0 || p[0] == nil {
		return eao
	}

	return eao
}

func expandJiraApprovalOptions(p []interface{}) *eaaspb.JiraApprovalOptions {
	jao := &eaaspb.JiraApprovalOptions{}
	if len(p) == 0 || p[0] == nil {
		return jao
	}

	return jao
}

func expandGithubPRApprovalOptions(p []interface{}) *eaaspb.GithubPullRequestApprovalOptions {
	ghao := &eaaspb.GithubPullRequestApprovalOptions{}
	if len(p) == 0 || p[0] == nil {
		return ghao
	}

	return ghao
}

func expandNotificationOptions(p []interface{}) *eaaspb.NotificationOptions {
	no := &eaaspb.NotificationOptions{}
	if len(p) == 0 || p[0] == nil {
		return no
	}

	return no
}

func expandScriptOptions(p []interface{}) *eaaspb.ShellScriptOptions {
	so := &eaaspb.ShellScriptOptions{}
	if len(p) == 0 || p[0] == nil {
		return so
	}

	in := p[0].(map[string]interface{})

	if s, ok := in["script"].(string); ok && len(s) > 0 {
		so.Script = s
	}

	if ev, ok := in["envvars"].(map[string]string); ok && len(ev) > 0 {
		so.Envvars = ev
	}

	if c, ok := in["cpu_limit_milli"].(string); ok && len(c) > 0 {
		so.CpuLimitMilli = c
	}

	if m, ok := in["memory_limit_mb"].(string); ok && len(m) > 0 {
		so.MemoryLimitMB = m
	}

	if s, ok := in["success_condition"].(string); ok && len(s) > 0 {
		so.SuccessCondition = s
	}

	return so
}

func expandContainerOptions(p []interface{}) *eaaspb.ContainerOptions {
	co := &eaaspb.ContainerOptions{}
	if len(p) == 0 || p[0] == nil {
		return co
	}

	in := p[0].(map[string]interface{})

	if i, ok := in["image"].(string); ok && len(i) > 0 {
		co.Image = i
	}

	if args, ok := in["arguments"].([]interface{}); ok && len(args) > 0 {
		co.Arguments = toArrayString(args)
	}

	if cmds, ok := in["commands"].([]interface{}); ok && len(cmds) > 0 {
		co.Commands = toArrayString(cmds)
	}

	if ev, ok := in["envvars"].(map[string]interface{}); ok && len(ev) > 0 {
		co.Envvars = toMapString(ev)
	}

	if wdp, ok := in["working_dir_path"].(string); ok && len(wdp) > 0 {
		co.WorkingDirPath = wdp
	}

	if c, ok := in["cpu_limit_milli"].(string); ok && len(c) > 0 {
		co.CpuLimitMilli = c
	}

	if m, ok := in["memory_limit_mb"].(string); ok && len(m) > 0 {
		co.MemoryLimitMB = m
	}

	if s, ok := in["success_condition"].(string); ok && len(s) > 0 {
		co.SuccessCondition = s
	}

	return co
}

func expandHttpOptions(p []interface{}) *eaaspb.HttpOptions {
	ho := &eaaspb.HttpOptions{}
	if len(p) == 0 || p[0] == nil {
		return ho
	}

	in := p[0].(map[string]interface{})

	if ep, ok := in["endpoint"].(string); ok && len(ep) > 0 {
		ho.Endpoint = ep
	}

	if m, ok := in["method"].(string); ok && len(m) > 0 {
		ho.Method = m
	}

	if h, ok := in["headers"].(map[string]interface{}); ok && len(h) > 0 {
		ho.Headers = toMapString(h)
	}

	if b, ok := in["body"].(string); ok && len(b) > 0 {
		ho.Body = b
	}

	if s, ok := in["success_condition"].(string); ok && len(s) > 0 {
		ho.SuccessCondition = s
	}

	return ho
}

func flattenEaasHooks(input []*eaaspb.Hook, p []interface{}) []interface{} {
	if len(input) == 0 {
		return nil
	}

	out := make([]interface{}, len(input))
	for i, in := range input {
		log.Println("flatten eaas hook ", in)
		obj := map[string]interface{}{}
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}
		obj["name"] = in.Name
		obj["description"] = in.Description
		obj["type"] = in.Type
		v, _ := obj["options"].([]interface{})
		obj["options"] = flattenHookOptions(in.Options, v)
		obj["agents"] = flattenEaasAgents(in.Agents)
		obj["timeout_seconds"] = in.TimeoutSeconds
		obj["on_failure"] = in.OnFailure
		if v, ok := obj["driver"].([]interface{}); ok && len(v) > 0 {
			obj["type"] = "driver"
			obj["driver"] = flattenWorkflowHandlerCompoundRef(in.WorkflowHandler)
		} else if v, ok = obj["workflow_handler"].([]interface{}); ok && len(v) > 0 {
			obj["workflow_handler"] = flattenWorkflowHandlerCompoundRef(in.WorkflowHandler)
		}
		obj["depends_on"] = toArrayInterface(in.DependsOn)
		obj["execute_once"] = in.ExecuteOnce
		obj["skip_config"] = flattenSkipConfig(in.SkipConfig)
		out[i] = &obj
		log.Println("flatten hook setting object ", out[i])
	}

	return out
}

func flattenSkipConfig(in *eaaspb.SkipConfig) []any {
	if in == nil {
		return nil
	}
	obj := map[string]any{
		"condition":       in.Condition,
		"skip_on_destroy": in.SkipOnDestroy,
	}
	return []any{obj}
}

func flattenHookOptions(input *eaaspb.HookOptions, p []interface{}) []interface{} {
	if input == nil || cmp.Equal(input, &eaaspb.HookOptions{}, cmpopts.IgnoreUnexported(eaaspb.HookOptions{})) {
		return nil
	}

	obj := map[string]interface{}{}
	if len(p) > 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	v, _ := obj["approval"].([]interface{})
	obj["approval"] = flattenApprovalOptions(input.Approval, v)

	obj["notification"] = flattenNotificationOptions(input.Notification)

	v, _ = obj["script"].([]interface{})
	obj["script"] = flattenScriptOptions(input.Script, v)

	v, _ = obj["container"].([]interface{})
	obj["container"] = flattenContainerOptions(input.Container, v)

	v, _ = obj["http"].([]interface{})
	obj["http"] = flattenHttpOptions(input.Http, v)

	return []interface{}{obj}
}

func flattenApprovalOptions(input *eaaspb.ApprovalOptions, p []interface{}) []interface{} {
	if input == nil {
		return nil
	}

	obj := map[string]interface{}{}
	if len(p) > 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	obj["type"] = input.Type
	obj["internal"] = flattenInternalApprovalOptions(input.Internal)
	obj["email"] = flattenEmailApprovalOptions(input.Email)
	obj["jira"] = flattenJiraApprovalOptions(input.Jira)
	obj["github_pull_request"] = flattenGithubPRApprovalOptions(input.GithubPullRequest)

	return []interface{}{obj}
}

func flattenInternalApprovalOptions(input *eaaspb.InternalApprovalOptions) []interface{} {
	if input == nil {
		return nil
	}

	if len(input.Emails) == 0 {
		return nil
	}

	obj := map[string]interface{}{}
	obj["emails"] = toArrayInterface(input.Emails)

	return []interface{}{obj}
}

func flattenEmailApprovalOptions(input *eaaspb.EmailApprovalOptions) []interface{} {
	if input == nil {
		return nil
	}

	obj := map[string]interface{}{}
	return []interface{}{obj}
}

func flattenJiraApprovalOptions(input *eaaspb.JiraApprovalOptions) []interface{} {
	if input == nil {
		return nil
	}

	obj := map[string]interface{}{}
	return []interface{}{obj}
}

func flattenGithubPRApprovalOptions(input *eaaspb.GithubPullRequestApprovalOptions) []interface{} {
	if input == nil {
		return nil
	}

	obj := map[string]interface{}{}
	return []interface{}{obj}
}

func flattenNotificationOptions(input *eaaspb.NotificationOptions) []interface{} {
	if input == nil {
		return nil
	}

	obj := map[string]interface{}{}
	return []interface{}{obj}
}

func flattenScriptOptions(input *eaaspb.ShellScriptOptions, p []interface{}) []interface{} {
	if input == nil {
		return nil
	}

	obj := map[string]interface{}{}
	if len(p) > 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	obj["script"] = input.Script
	obj["envvars"] = toMapInterface(input.Envvars)
	obj["cpu_limit_milli"] = input.CpuLimitMilli
	obj["memory_limit_mb"] = input.MemoryLimitMB
	obj["success_condition"] = input.SuccessCondition

	return []interface{}{obj}
}

func flattenContainerOptions(input *eaaspb.ContainerOptions, p []interface{}) []interface{} {
	if input == nil {
		return nil
	}

	obj := map[string]interface{}{}
	if len(p) > 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	obj["image"] = input.Image
	obj["arguments"] = toArrayInterface(input.Arguments)
	obj["commands"] = toArrayInterface(input.Commands)
	obj["envvars"] = toMapInterface(input.Envvars)
	obj["working_dir_path"] = input.WorkingDirPath
	obj["cpu_limit_milli"] = input.CpuLimitMilli
	obj["memory_limit_mb"] = input.MemoryLimitMB
	obj["success_condition"] = input.SuccessCondition

	return []interface{}{obj}
}

func flattenHttpOptions(input *eaaspb.HttpOptions, p []interface{}) []interface{} {
	if input == nil {
		return nil
	}

	obj := map[string]interface{}{}
	if len(p) > 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	obj["endpoint"] = input.Endpoint
	obj["method"] = input.Method
	obj["headers"] = toMapInterface(input.Headers)
	obj["body"] = input.Body
	obj["success_condition"] = input.SuccessCondition

	return []interface{}{obj}
}

func expandBoolValue(in []interface{}) *datatypes.BoolValue {
	if len(in) == 0 {
		return nil
	}

	bv := in[0].(map[string]interface{})
	return datatypes.NewBool(bv["value"].(bool))
}

func flattenBoolValue(in *datatypes.BoolValue) []interface{} {
	if in == nil {
		return nil
	}

	obj := make(map[string]interface{})
	obj["value"] = in.Value

	return []interface{}{obj}
}

func expandV3MetaData(p []interface{}) *commonpb.Metadata {
	obj := &commonpb.Metadata{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}

	in := p[0].(map[string]interface{})
	if v, ok := in["name"].(string); ok && len(v) > 0 {
		obj.Name = v
	}
	if v, ok := in["description"].(string); ok && len(v) > 0 {
		obj.Description = v
	}
	if v, ok := in["project"].(string); ok && len(v) > 0 {
		obj.Project = v
	}

	if v, ok := in["labels"].(map[string]interface{}); ok && len(v) > 0 {
		obj.Labels = toMapString(v)
	}

	if v, ok := in["annotations"].(map[string]interface{}); ok && len(v) > 0 {
		obj.Annotations = toMapString(v)
	}

	if v, ok := in["display_name"].(string); ok && len(v) > 0 {
		obj.DisplayName = v
	}
	return obj
}

func flattenV3MetaData(in *commonpb.Metadata) []interface{} {
	if in == nil {
		return nil
	}
	log.Println("flatten metadata: ", in)
	obj := map[string]any{
		"name":         in.Name,
		"description":  in.Description,
		"project":      in.Project,
		"labels":       toMapInterface(in.Labels),
		"annotations":  toMapInterface(in.Annotations),
		"display_name": in.DisplayName,
	}
	return []interface{}{obj}
}

func validateResourceName(name string) error {
	errs := validation.IsDNS1123Subdomain(name)
	if len(errs) != 0 {
		return fmt.Errorf("%s", strings.Join(errs, " "))
	}
	return nil
}

func checkStandardInputTextError(input string) bool {
	dns1123ValidationErrMsg := "a lowercase RFC 1123 subdomain must consist of lower case alphanumeric characters"
	return strings.Contains(input, dns1123ValidationErrMsg)
}

func expandEnvVarOptions(p []interface{}) *eaaspb.EnvVarOptions {
	if len(p) == 0 || p[0] == nil {
		return nil
	}

	options := &eaaspb.EnvVarOptions{}
	opts := p[0].(map[string]interface{})

	if v, ok := opts["description"].(string); ok && len(v) > 0 {
		options.Description = v
	}

	if v, ok := opts["sensitive"].(bool); ok {
		options.Sensitive = v
	}

	if v, ok := opts["required"].(bool); ok {
		options.Required = v
	}

	if v, ok := opts["immutable"].(bool); ok {
		options.Immutable = v
	}

	if v, ok := opts["display_metadata"].(string); ok && len(v) > 0 {
		options.DisplayMetadata = expandDisplayMetadata(v)
	}

	if v, ok := opts["override"].([]interface{}); ok && len(v) > 0 {
		options.Override = expandEnvVarOverrideOptions(v)
	}

	if v, ok := opts["schema"].([]interface{}); ok && len(v) > 0 {
		options.Schema = expandCustomSchema(v)
	}

	return options

}

func expandEnvVarOverrideOptions(p []interface{}) *eaaspb.EnvVarOverrideOptions {
	if len(p) == 0 || p[0] == nil {
		return nil
	}

	override := &eaaspb.EnvVarOverrideOptions{}
	in := p[0].(map[string]interface{})

	if v, ok := in["type"].(string); ok && len(v) > 0 {
		override.Type = v
	}

	if vals, ok := in["restricted_values"].([]interface{}); ok && len(vals) > 0 {
		override.RestrictedValues = toArrayString(vals)
	}

	if vals, ok := in["selectors"].([]interface{}); ok && len(vals) > 0 {
		override.Selectors = toArrayString(vals)
	}

	if v, ok := in["display_overridden"].(bool); ok {
		override.DisplayOverridden = v
	}

	if vals, ok := in["restricted_key_values"].(map[string]any); ok && len(vals) > 0 {
		override.RestrictedKeyValues = toMapString(vals)
	}

	return override
}

func flattenEnvVarOptions(input *eaaspb.EnvVarOptions) []interface{} {
	log.Println("flatten envvar options")
	cmpOpt := cmpopts.IgnoreUnexported(eaaspb.EnvVarOptions{})
	if input == nil || cmp.Equal(input, &eaaspb.EnvVarOptions{}, cmpOpt) ||
		cmp.Equal(input, &eaaspb.EnvVarOptions{Mask: true}, cmpOpt) {
		return nil
	}

	obj := map[string]interface{}{
		"description":      input.Description,
		"sensitive":        input.Sensitive,
		"required":         input.Required,
		"immutable":        input.Immutable,
		"override":         flattenEnvVarOverrideOptions(input.Override),
		"schema":           flattenCustomSchema(input.Schema),
		"display_metadata": flattenDisplayMetadata(input.DisplayMetadata),
	}
	return []interface{}{obj}
}

func flattenDisplayMetadata(in *structpb.Struct) string {
	if in == nil {
		return ""
	}
	b, err := in.MarshalJSON()
	if err != nil {
		return ""
	}
	return string(b)
}

func flattenEnvVarOverrideOptions(input *eaaspb.EnvVarOverrideOptions) []interface{} {
	log.Println("flatten envvar override options")
	if input == nil {
		return nil
	}
	obj := map[string]interface{}{
		"type":                  input.Type,
		"restricted_values":     toArrayInterface(input.RestrictedValues),
		"selectors":             toArrayInterface(input.Selectors),
		"display_overridden":    input.DisplayOverridden,
		"restricted_key_values": toMapInterface(input.RestrictedKeyValues),
	}
	return []interface{}{obj}
}

func expandFileOptions(p []interface{}) *commonpb.FileOptions {
	if len(p) == 0 || p[0] == nil {
		return nil
	}

	options := &commonpb.FileOptions{}
	opts := p[0].(map[string]interface{})

	if v, ok := opts["description"].(string); ok && len(v) > 0 {
		options.Description = v
	}

	if v, ok := opts["sensitive"].(bool); ok {
		options.Sensitive = v
	}

	if v, ok := opts["required"].(bool); ok {
		options.Required = v
	}

	if v, ok := opts["override"].([]interface{}); ok && len(v) > 0 {
		options.Override = expandFileOverrideOptions(v)
	}

	if v, ok := opts["display_metadata"].(string); ok && len(v) > 0 {
		options.DisplayMetadata = expandDisplayMetadata(v)
	}

	if v, ok := opts["schema"].([]interface{}); ok && len(v) > 0 {
		options.Schema = expandCustomSchema(v)
	}

	return options

}

func expandFileOverrideOptions(p []interface{}) *commonpb.FileOverrideOptions {
	if len(p) == 0 || p[0] == nil {
		return nil
	}

	override := &commonpb.FileOverrideOptions{}
	in := p[0].(map[string]interface{})

	if v, ok := in["type"].(string); ok && len(v) > 0 {
		override.Type = v
	}

	return override
}

func flattenFileOptions(input *commonpb.FileOptions) []interface{} {
	log.Println("flatten file options")
	ignoreOpt := cmpopts.IgnoreUnexported(eaaspb.EnvVarOptions{})
	if input == nil ||
		cmp.Equal(input, &eaaspb.EnvVarOptions{}, ignoreOpt) ||
		cmp.Equal(input, &eaaspb.EnvVarOptions{Mask: true}, ignoreOpt) {
		return nil
	}
	obj := map[string]interface{}{
		"description":      input.Description,
		"sensitive":        input.Sensitive,
		"required":         input.Required,
		"override":         flattenFileOverrideOptions(input.Override),
		"display_metadata": flattenDisplayMetadata(input.DisplayMetadata),
		"schema":           flattenCustomSchema(input.Schema),
	}

	return []interface{}{obj}
}

func flattenFileOverrideOptions(input *commonpb.FileOverrideOptions) []interface{} {
	log.Println("flatten file override options")
	if input == nil {
		return nil
	}
	obj := map[string]any{
		"type": input.Type,
	}
	return []interface{}{obj}
}

func expandCustomSchema(p []interface{}) *commonpb.Schema {
	log.Println("expand custom schema")
	if len(p) == 0 || p[0] == nil {
		return nil
	}
	schema := &commonpb.Schema{}
	in := p[0].(map[string]interface{})

	if v, ok := in["jsonschema"].(string); ok && len(v) > 0 {
		schema.Jsonschema = expandJsonUISchema(v)
	}

	if v, ok := in["uischema"].(string); ok && len(v) > 0 {
		schema.Uischema = expandJsonUISchema(v)
	}

	return schema
}

func flattenCustomSchema(input *commonpb.Schema) []interface{} {
	log.Println("flatten custom schema")
	if input == nil {
		return nil
	}
	obj := map[string]any{
		"jsonschema": flattenJsonUISchema(input.Jsonschema),
		"uischema":   flattenJsonUISchema(input.Uischema),
	}
	return []interface{}{obj}
}

func expandJsonUISchema(p string) *structpb.Struct {
	if len(p) == 0 {
		return nil
	}

	jsonSchemaMap := map[string]interface{}{}
	if err := json.Unmarshal([]byte(p), &jsonSchemaMap); err == nil {
		s, err := structpb.NewStruct(jsonSchemaMap)
		if err != nil {
			return nil
		}

		return s
	}
	return nil
}

func flattenJsonUISchema(in *structpb.Struct) string {
	if in == nil {
		return ""
	}
	b, _ := in.MarshalJSON()
	return string(b)
}

func expandActions(p []interface{}) ([]*eaaspb.Action, error) {
	actions := make([]*eaaspb.Action, 0)
	if len(p) == 0 || p[0] == nil {
		return actions, nil
	}
	var err error

	for i := range p {
		action := eaaspb.Action{}
		in := p[i].(map[string]interface{})

		if v, ok := in["name"].(string); ok && len(v) > 0 {
			action.Name = v
		}

		if v, ok := in["description"].(string); ok && len(v) > 0 {
			action.Description = v
		}

		if v, ok := in["type"].(string); ok && len(v) > 0 {
			action.Type = v
		}

		if v, ok := in["context"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
			action.Context = expandConfigContextCompoundRef(v[0].(map[string]any))
		}

		if h, ok := in["workflows"].([]interface{}); ok && len(h) > 0 {
			action.Workflows, err = expandCustomProviderOptions(h)
			if err != nil {
				return nil, err
			}
		}

		if rr, ok := in["reconcile_resources"].([]interface{}); ok && len(rr) > 0 {
			action.ReconcileResources = expandReconcileResources(rr)
		}

		actions = append(actions, &action)
	}

	return actions, nil
}

func flattenActions(input []*eaaspb.Action, p []interface{}) []interface{} {
	log.Println("flatten actions start")
	if input == nil {
		return nil
	}

	out := make([]interface{}, len(input))
	for i, in := range input {
		log.Println("flatten action ", in)
		obj := map[string]interface{}{}
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}
		obj["name"] = in.Name
		obj["description"] = in.Description
		obj["type"] = in.Type
		obj["context"] = []interface{}{flattenConfigContextCompoundRef(in.Context)}
		v, _ := obj["workflows"].([]interface{})
		obj["workflows"] = flattenCustomProviderOptions(in.Workflows, v)
		obj["reconcile_resources"] = flattenReconcileResources(in.ReconcileResources)
		out[i] = &obj
	}

	return out
}

func expandReconcileResources(p []interface{}) []*commonpb.ResourceNameAndVersionRef {
	reconcileResources := make([]*commonpb.ResourceNameAndVersionRef, 0)
	if len(p) == 0 {
		return reconcileResources
	}

	for indx := range p {
		obj := &commonpb.ResourceNameAndVersionRef{}

		in := p[indx].(map[string]interface{})

		if v, ok := in["name"].(string); ok && len(v) > 0 {
			obj.Name = v
		}

		if v, ok := in["version"].(string); ok && len(v) > 0 {
			obj.Version = v
		}

		reconcileResources = append(reconcileResources, obj)
	}

	return reconcileResources
}

func flattenReconcileResources(input []*commonpb.ResourceNameAndVersionRef) []interface{} {
	log.Println("flatten reconcile resources start")
	if len(input) == 0 {
		return nil
	}

	out := make([]interface{}, len(input))
	for i, in := range input {
		obj := map[string]any{
			"name": in.Name,
		}
		out[i] = &obj
	}

	return out
}

func IsResourceNotFoundErr(err error) bool {
	errStr := err.Error()
	switch true {
	case strings.Contains(errStr, "code 404"):
		return true
	case strings.Contains(errStr, "not found"):
		return true
	case strings.Contains(errStr, "does not exist"):
		return true
	default:
		return false
	}
}
