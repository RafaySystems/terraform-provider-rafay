package rafay

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"sort"
	"strings"

	commonpb "github.com/RafaySystems/rafay-common/proto/types/hub/commonpb"
	"github.com/RafaySystems/rafay-common/proto/types/hub/gitopspb"
	"github.com/RafaySystems/rafay-common/proto/types/hub/integrationspb"
	"github.com/RafaySystems/rctl/pkg/config"
	"github.com/RafaySystems/rctl/utils"
	"github.com/davecgh/go-spew/spew"
	"github.com/go-yaml/yaml"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"k8s.io/apimachinery/pkg/api/resource"
)

type File struct {
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Data []byte `protobuf:"bytes,2,opt,name=data,proto3" json:"data,omitempty"`
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
		value := v.(string)
		out[i] = []byte(value)
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
	}

	log.Println("expandMetaData")
	if v, ok := in["annotations"].(map[string]interface{}); ok && len(v) > 0 {
		w1 := spew.Sprintf("%+v", v)
		log.Println("expandMetaData annotations ", w1)
		obj.Annotations = toMapString(v)
	}
	return obj
}

func expandV1MetaData(p []interface{}) *commonpb.Metadata {
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

	obj.Labels = nil

	obj.Annotations = nil

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

func expandResourceQuantity(p []interface{}) *commonpb.ResourceQuantity {
	obj := commonpb.ResourceQuantity{}
	if len(p) == 0 || p[0] == nil {
		log.Println("expandResourceQuantity empty input")
		return &obj
	}
	in := p[0].(map[string]interface{})
	if v, ok := in["memory"].([]interface{}); ok {
		obj.Memory = expandQuantity(v)
		log.Println("expandResourceQuantity memory", obj.Memory)
	}

	if v, ok := in["cpu"].([]interface{}); ok {
		obj.Cpu = expandQuantity(v)
		log.Println("expandResourceQuantity CPU", obj.Cpu)
	}

	log.Println("expandResourceQuantity obj", obj)
	return &obj
}

func expandResourceQuantityString(p []interface{}) *commonpb.ResourceQuantity {
	obj := commonpb.ResourceQuantity{}
	if len(p) == 0 || p[0] == nil {
		log.Println("expandResourceQuantity empty input")
		return &obj
	}
	in := p[0].(map[string]interface{})
	if v, ok := in["memory"].(string); ok {
		obj.Memory = expandQuantityString(v)
		log.Println("expandResourceQuantity memory", obj.Memory)
	}

	if v, ok := in["cpu"].(string); ok {
		obj.Cpu = expandQuantityString(v)
		log.Println("expandResourceQuantity CPU", obj.Cpu)
	}

	log.Println("expandResourceQuantity obj", obj)
	return &obj
}

func expandProjectMeta(p []interface{}) []*commonpb.ProjectMeta {
	if len(p) == 0 {
		return []*commonpb.ProjectMeta{}
	}
	out := make([]*commonpb.ProjectMeta, len(p))
	for i := range p {
		in := p[i].(map[string]interface{})
		obj := commonpb.ProjectMeta{}

		if v, ok := in["name"].(string); ok && len(v) > 0 {
			obj.Name = v
		}
		if v, ok := in["id"].(string); ok && len(v) > 0 {
			obj.Id = v
		}

		out[i] = &obj
	}

	log.Println("expandProjectMeta out", out)
	return out
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
	}

	log.Println("expandSharingSpec obj", obj)
	return &obj
}

// Flatteners

func flattenMetaData(in *commonpb.Metadata) []interface{} {
	if in == nil {
		return nil
	}

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

func flattenV1MetaData(in *commonpb.Metadata) []interface{} {
	if in == nil {
		return nil
	}

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

	return []interface{}{obj}
}

func flattenFile(in *File) []interface{} {
	if in == nil {
		return nil
	}

	obj := make(map[string]interface{})
	if len(in.Name) > 0 {
		obj["name"] = in.Name
	}
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
		if len(in.Name) > 0 {
			obj["name"] = in.Name
		}
		out[i] = obj
	}

	return out
}

func flattenResourceQuantity(in *commonpb.ResourceQuantity) []interface{} {
	if in == nil {
		return nil
	}

	obj := make(map[string]interface{})
	if in.Memory != nil {
		obj1 := make([]interface{}, 1)
		obj2 := make(map[string]interface{})
		obj2["string"] = in.GetMemory().String()
		obj1[0] = obj2
		obj["memory"] = obj1
	}

	if in.Cpu != nil {
		obj1 := make([]interface{}, 1)
		obj2 := make(map[string]interface{})
		obj2["string"] = in.GetCpu().String()
		obj1[0] = obj2
		obj["cpu"] = obj1
	}

	log.Println("flattenResourceQuantity obj", obj)
	return []interface{}{obj}
}

func flattenResourceQuantities(in *commonpb.ResourceQuantity) []interface{} {
	if in == nil {
		return nil
	}
	objRoot := make([]interface{}, 1)

	obj := make(map[string]interface{})
	if in.Memory != nil {
		obj1 := make([]interface{}, 1)
		obj2 := make(map[string]interface{})
		obj2["string"] = in.GetMemory()
		obj1[0] = obj2
		obj["memory"] = obj1
	}

	if in.Cpu != nil {
		obj1 := make([]interface{}, 1)
		obj2 := make(map[string]interface{})
		obj2["string"] = in.GetCpu()
		obj1[0] = obj2
		obj["cpu"] = obj1
	}

	objRoot[0] = obj
	log.Println("flattenResourceQuantity obj", obj)
	return []interface{}{objRoot}
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

func flattenProjectMeta(input []*commonpb.ProjectMeta) []interface{} {
	if input == nil {
		return nil
	}

	out := make([]interface{}, len(input))
	for i, in := range input {
		obj := map[string]interface{}{}
		if len(in.Name) > 0 {
			obj["name"] = in.Name
		}
		if len(in.Id) > 0 {
			obj["id"] = in.Id
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
		obj["projects"] = flattenProjectMeta(in.Projects)
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
		obj.Memory = expandQuantity1170(v)
		log.Println("expandResourceQuantity1170 memory", obj.Memory)
	}

	if v, ok := in["cpu"].(string); ok {
		obj.Cpu = expandQuantity1170(v)
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

func flattenResourceQuantity1170(in *commonpb.ResourceQuantity) []interface{} {
	if in == nil {
		return nil
	}

	obj := make(map[string]interface{})
	if in.Memory != nil {
		for i := 0; i < 10; i++ {
			in.GetMemory().Add(*in.GetMemory())
			//log.Println("adding ", in.GetMemory().String())
		}
		obj["memory"] = in.GetMemory().String()
		//log.Println("flattenResourceQuantity1170 memory string ", in.GetMemory().String())
	}

	if in.Cpu != nil {
		cq := *in.Cpu
		for i := 0; i < 999; i++ {
			in.GetCpu().Add(cq)
			//log.Println("adding ", in.GetCpu().String())
		}
		in.GetCpu().RoundUp(resource.Micro)
		obj["cpu"] = in.GetCpu().String()
		//log.Println("flattenResourceQuantity1170 cpu string ", in.GetCpu().String(), " => ")
	}

	log.Println("flattenResourceQuantityV101 obj", obj)
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
