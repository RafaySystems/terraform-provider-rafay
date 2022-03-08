package rafay

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/RafaySystems/rafay-common/pkg/hub/client/options"
	typed "github.com/RafaySystems/rafay-common/pkg/hub/client/typed"
	"github.com/RafaySystems/rafay-common/pkg/hub/terraform/resource"
	commonpb "github.com/RafaySystems/rafay-common/proto/types/hub/commonpb"
	"github.com/RafaySystems/rafay-common/proto/types/hub/infrapb"
	"github.com/RafaySystems/rctl/pkg/config"
	"github.com/RafaySystems/rctl/pkg/versioninfo"
	"github.com/davecgh/go-spew/spew"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/pkg/errors"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type namespaceSpecTranspose struct {
	Psp            *infrapb.NamespacePSP            `protobuf:"bytes,1,opt,name=psp,proto3" json:"psp,omitempty"`
	Placement      *commonpb.PlacementSpec          `protobuf:"bytes,2,opt,name=placement,proto3" json:"placement,omitempty"`
	Drift          *commonpb.DriftSpec              `protobuf:"bytes,3,opt,name=drift,proto3" json:"drift,omitempty"`
	ResourceQuotas *infrapb.NamespaceResourceQuotas `protobuf:"bytes,4,opt,name=resourceQuotas,proto3" json:"resourceQuotas,omitempty"`
	LimitRange     *infrapb.NamespaceLimitRange     `protobuf:"bytes,5,opt,name=limitRange,proto3" json:"limitRange,omitempty"`
	// Types that are assignable to Artifact:
	//	*NamespaceSpec_Uploaded
	//	*NamespaceSpec_Repo
	//Artifact isNamespaceSpec_Artifact `protobuf_oneof:"artifact"`

	Artifact struct {
		Repository string `protobuf:"bytes,1,opt,name=repository,proto3" json:"repository,omitempty"`
		Revision   string `protobuf:"bytes,2,opt,name=revision,proto3" json:"revision,omitempty"`
		Path       *File  `protobuf:"bytes,3,opt,name=path,proto3" json:"path,omitempty"`
	} `json:"artifact,omitempty"`
}

func resourceNamespace() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceNamespaceCreate,
		ReadContext:   resourceNamespaceRead,
		UpdateContext: resourceNamespaceUpdate,
		DeleteContext: resourceNamespaceDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema:        resource.NamespaceSchema.Schema,
	}
}

func resourceNamespaceCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("namespace create starts")
	diags := resourceNamespaceUpsert(ctx, d, m)
	if diags.HasError() {
		tflog := os.Getenv("TF_LOG")
		if tflog == "TRACE" || tflog == "DEBUG" {
			ctx = context.WithValue(ctx, "debug", "true")
		}
		log.Printf("namespace create got error, perform cleanup")
		ns, err := expandNamespace(d)
		if err != nil {
			log.Printf("namespace expandNamespace error")
			return diag.FromErr(err)
		}
		auth := config.GetConfig().GetAppAuthProfile()
		client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
		if err != nil {
			return diag.FromErr(err)
		}

		err = client.InfraV3().Namespace().Delete(ctx, options.DeleteOptions{
			Name:    ns.Metadata.Name,
			Project: ns.Metadata.Project,
		})
		if err != nil {
			return diag.FromErr(err)
		}
	}
	return diags
}

func resourceNamespaceUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("namespace update starts")
	return resourceNamespaceUpsert(ctx, d, m)
}

func resourceNamespaceUpsert(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("namespace upsert starts")
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	ns, err := expandNamespace(d)
	if err != nil {
		log.Printf("namespace expandNamespace error")
		return diag.FromErr(err)
	}

	artifactNames, err := ns.ArtifactList()
	if err != nil {
		err = errors.Wrapf(err, "unable to list artifacts")
		return diag.FromErr(err)
	}
	log.Println(" resourceNamespaceUpsert artifactNames ", artifactNames)

	for _, artifactName := range artifactNames {

		if !strings.HasPrefix(artifactName, "file://") {
			continue
		}
		//get full path of artifact
		artifactFullPath := filepath.Join(filepath.Dir("."), artifactName[7:])
		//retrieve artifact data
		artifactData, err := ioutil.ReadFile(artifactFullPath)
		if err != nil {
			err = fmt.Errorf("unable to read artifact at '%s'", artifactFullPath)
			return diag.FromErr(err)
		}
		//set artifact in namespace
		err = ns.ArtifactSet(artifactName, artifactData)
		if err != nil {
			err = fmt.Errorf("unable to set artifact %s at path '%s'", artifactName, artifactFullPath)
			return diag.FromErr(err)
		}
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}
	w1 := spew.Sprintf("%+v", ns)
	log.Println("resourceNamespaceUpsert ns:", w1)
	err = client.InfraV3().Namespace().Apply(ctx, ns, options.ApplyOptions{})
	if err != nil {
		// XXX Debug
		// n1 := spew.Sprintf("%+v", ns)
		// log.Println("namespace apply ns:", n1)
		log.Printf("namespace apply error")
		return diag.FromErr(err)
	}

	// wait for publish
	for {
		time.Sleep(30 * time.Second)
		nsStatus, err := client.InfraV3().Namespace().Status(ctx, options.StatusOptions{
			Name:    ns.Metadata.Name,
			Project: ns.Metadata.Project,
		})
		if err != nil {
			return diag.FromErr(err)
		}
		//check if namespace can be placed on a cluster, if true break out of infinite loop
		if nsStatus.Status.ConditionStatus == commonpb.ConditionStatus_StatusOK ||
			nsStatus.Status.ConditionStatus == commonpb.ConditionStatus_StatusNotSet {
			break
		}
		if nsStatus.Status.ConditionStatus == commonpb.ConditionStatus_StatusFailed {
			return diag.FromErr(fmt.Errorf("%s", "failed to publish namespace"))
		}
		log.Println("nsStatus.Status.ConditionStatus ", nsStatus.Status.ConditionStatus)
	}

	d.SetId(ns.Metadata.Name)
	return diags

}

func resourceNamespaceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("namespace read starts")
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	nsTFState, err := expandNamespace(d)
	if err != nil {
		return diag.FromErr(err)
	}

	// XXX Debug
	n1 := spew.Sprintf("%+v", nsTFState)
	log.Println("resourceNamespaceRead nsTFState ", n1)

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}

	ns, err := client.InfraV3().Namespace().Get(ctx, options.GetOptions{
		Name:    nsTFState.Metadata.Name,
		Project: nsTFState.Metadata.Project,
	})
	if err != nil {
		return diag.FromErr(err)
	}
	log.Println("resourceNamespaceRead remoteState ", ns)
	// XXX Debug
	// n1 = spew.Sprintf("%+v", ns)
	// log.Println("resourceNamespaceRead ns", n1)
	if ns.Spec.ResourceQuotas != nil && ns.Spec.ResourceQuotas.Requests != nil && ns.Spec.ResourceQuotas.Requests.Memory != nil {
		log.Println("resourceNamespaceRead ns.Spec.ResourceQuotas Memory", ns.Spec.ResourceQuotas.Requests.Memory)
	}

	err = flattenNamespace(d, ns)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceNamespaceDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	nsTFState, err := expandNamespace(d)
	if err != nil {
		return diag.FromErr(err)
	}

	// XXX Debug
	// n1 := spew.Sprintf("%+v", nsTFState)
	// log.Println("resourceNamespaceRead nsTFState", n1)

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.InfraV3().Namespace().Delete(ctx, options.DeleteOptions{
		Name:    nsTFState.Metadata.Name,
		Project: nsTFState.Metadata.Project,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func expandNamespace(in *schema.ResourceData) (*infrapb.Namespace, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "expand namespace empty input")
	}
	obj := &infrapb.Namespace{}

	if v, ok := in.Get("metadata").([]interface{}); ok {
		obj.Metadata = expandMetaData(v)
	}
	w1 := spew.Sprintf("%+v", obj.Metadata)
	log.Println("expandNamespace metadata ", w1)

	if v, ok := in.Get("spec").([]interface{}); ok {
		objSpec, err := expandNamespaceSpec(v)
		if err != nil {
			return nil, err
		}
		log.Println("expandNamespace got spec")
		obj.Spec = objSpec
	}

	obj.ApiVersion = "infra.k8smgmt.io/v3"
	obj.Kind = "Namespace"
	return obj, nil
}

func expandNamespaceSpec(p []interface{}) (*infrapb.NamespaceSpec, error) {
	obj := infrapb.NamespaceSpec{}
	if len(p) == 0 || p[0] == nil {
		return &obj, fmt.Errorf("%s", "expandNamespaceSpec empty input")
	}

	nst := namespaceSpecTranspose{}

	in := p[0].(map[string]interface{})
	if v, ok := in["psp"].([]interface{}); ok {
		nst.Psp = expandPsp(v)
	}

	if v, ok := in["placement"].([]interface{}); ok {
		nst.Placement = expandPlacement(v)
	}

	if v, ok := in["drift"].([]interface{}); ok {
		nst.Drift = expandDrift(v)
	}

	if v, ok := in["resource_quotas"].([]interface{}); ok {
		log.Println("resource_quotas v", v)
		nst.ResourceQuotas = expandNamespaceResourceQuotas(v)
		log.Println("nst.ResourceQuotas ", nst.ResourceQuotas)
	}

	if v, ok := in["limit_range"].([]interface{}); ok {
		nst.LimitRange = expandNamespaceLimitRange(v)
	}

	if vp, ok := in["artifact"].([]interface{}); ok && len(vp) > 0 {
		if len(vp) == 0 || vp[0] == nil {
			log.Println("expandArtifact empty artifact")
		} else {
			ina := vp[0].(map[string]interface{})

			if v, ok := ina["path"].([]interface{}); ok && len(v) > 0 {
				nst.Artifact.Path = expandFile(v)
			}

			if v, ok := ina["repository"].(string); ok && len(v) > 0 {
				nst.Artifact.Repository = v
			}

			if v, ok := ina["revision"].(string); ok && len(v) > 0 {
				nst.Artifact.Revision = v
			}
		}
	}
	// XXX Debug
	s := spew.Sprintf("%+v", nst)
	log.Println("expandNamespaceSpec nst", s)

	jsonSpec, err := json.Marshal(nst)
	if err != nil {
		return nil, err
	}

	// XXX Debug
	log.Println("expandNamespaceSpec jsonSpec ", string(jsonSpec))

	err = obj.UnmarshalJSON(jsonSpec)
	if err != nil {
		log.Println("expandNamespaceSpec artifact UnmarshalJSON error ", err)
		return nil, err
	}

	if obj.ResourceQuotas != nil && obj.ResourceQuotas.Requests != nil && obj.ResourceQuotas.Requests.Memory != nil {
		log.Println("expandNamespaceSpec Obj Memory ", obj.ResourceQuotas.Requests.Memory)
	}

	log.Println("expandNamespaceSpec Obj", obj)
	return &obj, nil
}

func expandPsp(p []interface{}) *infrapb.NamespacePSP {
	obj := &infrapb.NamespacePSP{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["name"].(string); ok {
		obj.Name = v
	}

	return obj
}

func expandNamespaceResourceQuotas(p []interface{}) *infrapb.NamespaceResourceQuotas {
	obj := &infrapb.NamespaceResourceQuotas{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}

	in := p[0].(map[string]interface{})
	if v, ok := in["requests"].([]interface{}); ok {
		log.Println("requests v", v)
		obj.Requests = expandResourceQuantity(v)
	}

	if v, ok := in["limits"].([]interface{}); ok {
		log.Println("limits v", v)
		obj.Limits = expandResourceQuantity(v)
	}

	log.Println("expandNamespaceResourceQuotas obj ", obj)
	return obj
}

func expandNamespaceLimitRange(p []interface{}) *infrapb.NamespaceLimitRange {
	obj := &infrapb.NamespaceLimitRange{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}

	//log.Println("expandNamespaceLimitRange")
	in := p[0].(map[string]interface{})
	if v, ok := in["pod"].([]interface{}); ok {
		obj.Pod = expandNamespaceLimitRangeConfig(v)
	}

	if v, ok := in["container"].([]interface{}); ok {
		obj.Container = expandNamespaceLimitRangeConfig(v)
	}

	return obj
}

func expandNamespaceLimitRangeConfig(p []interface{}) *infrapb.NamespaceLimitRangeConfig {
	obj := &infrapb.NamespaceLimitRangeConfig{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}

	in := p[0].(map[string]interface{})
	if v, ok := in["min"].([]interface{}); ok {
		obj.Min = expandResourceQuantity(v)
	}

	if v, ok := in["max"].([]interface{}); ok {
		obj.Max = expandResourceQuantity(v)
	}

	if v, ok := in["default"].([]interface{}); ok {
		obj.Default = expandResourceQuantity(v)
	}

	if v, ok := in["default_request"].([]interface{}); ok {
		obj.DefaultRequest = expandResourceQuantity(v)
	}

	//log.Println("expandNamespaceLimitRangeConfig <<")
	if v, ok := in["ratio"].([]interface{}); ok {
		//log.Println("expandNamespaceLimitRangeConfig ")
		obj.Ratio = expandResourceRatio(v)
	}

	return obj
}

func expandResourceRatio(p []interface{}) *commonpb.ResourceRatio {
	//log.Println("expandResourceRatio ")
	obj := &commonpb.ResourceRatio{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}

	in := p[0].(map[string]interface{})
	//w1 := spew.Sprintf("%+v", in)
	//log.Println("expandResourceRatio << in", w1)
	if v, ok := in["memory"].(float64); ok {
		obj.Memory = float32(v)
	}

	if v, ok := in["cpu"].(float64); ok {
		obj.Cpu = float32(v)
	}

	//log.Println("expandResourceRatio obj ", obj)
	return obj
}

// Flatteners

func flattenNamespace(d *schema.ResourceData, in *infrapb.Namespace) error {
	if in == nil {
		return nil
	}

	err := d.Set("metadata", flattenMetaData(in.Metadata))
	if err != nil {
		return err
	}

	v, ok := d.Get("spec").([]interface{})
	if !ok {
		v = []interface{}{}
	}

	// XXX Debug
	//w1 := spew.Sprintf("%+v", in.Spec)
	//log.Println("flattenNamespaceSpec before ", w1)
	var ret []interface{}
	ret, err = flattenNamespaceSpec(in.Spec, v)
	if err != nil {
		return err
	}
	// XXX Debug
	//w1 = spew.Sprintf("%+v", ret)
	//log.Println("flattenNamespaceSpec after ", w1)

	err = d.Set("spec", ret)
	if err != nil {
		return err
	}
	return nil
}

func flattenPSP(in *infrapb.NamespacePSP) []interface{} {
	if in == nil {
		return nil
	}

	obj := make(map[string]interface{})

	if len(in.Name) > 0 {
		obj["name"] = in.Name
	}

	return []interface{}{obj}
}

func flattenNamespaceResourceQuotas(in *infrapb.NamespaceResourceQuotas) []interface{} {
	if in == nil {
		return nil
	}

	retNil := true
	obj := make(map[string]interface{})

	if in.Requests != nil {
		obj["requests"] = flattenResourceQuantity(in.Requests)
		retNil = false
	}

	if in.Limits != nil {
		obj["limits"] = flattenResourceQuantity(in.Limits)
		retNil = false
	}

	if retNil {
		return nil
	}

	return []interface{}{obj}
}

func flattenNamespaceLimitRange(in *infrapb.NamespaceLimitRange) []interface{} {
	if in == nil {
		return nil
	}
	retNil := true
	obj := make(map[string]interface{})

	if in.Pod != nil {
		obj["pod"] = flattenNamespaceLimitRangeConfig(in.Pod)
		retNil = false
	}

	if in.Container != nil {
		obj["container"] = flattenNamespaceLimitRangeConfig(in.Container)
		retNil = false
	}

	if retNil {
		return nil
	}

	return []interface{}{obj}
}

func flattenNamespaceLimitRangeConfig(in *infrapb.NamespaceLimitRangeConfig) []interface{} {
	if in == nil {
		return nil
	}

	obj := make(map[string]interface{})

	if in.Min != nil {
		obj["min"] = flattenResourceQuantity(in.Min)
	}

	if in.Max != nil {
		obj["max"] = flattenResourceQuantity(in.Max)
	}

	if in.Default != nil {
		obj["default"] = flattenResourceQuantity(in.Default)
	}

	if in.Default != nil {
		obj["default_request"] = flattenResourceQuantity(in.DefaultRequest)
	}

	//log.Println("flattenNamespaceLimitRangeConfig ", in.Ratio)
	if in.Ratio != nil {
		obj["ratio"] = flattenRatio(in.Ratio)
	}

	return []interface{}{obj}
}

func flattenNamespaceSpec(in *infrapb.NamespaceSpec, p []interface{}) ([]interface{}, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "flattenNamespaceSpec empty input")
	}

	jsonBytes, err := in.MarshalJSON()
	if err != nil {
		log.Println("flattenNamespaceSpec MarshalJSON error", err)
		return nil, fmt.Errorf("%s %+v", "flattenNamespaceSpec MarshalJSON error", err)
	}

	log.Println("flattenNamespaceSpec jsonBytes ", string(jsonBytes))

	nsat := namespaceSpecTranspose{}
	err = json.Unmarshal(jsonBytes, &nsat)

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if in.Psp != nil {
		obj["psp"] = flattenPSP(in.Psp)
	}

	if in.Placement != nil {
		obj["placement"] = flattenPlacement(in.Placement)
	}

	if in.Drift != nil {
		obj["drift"] = flattenDrift(in.Drift)
	}

	if in.ResourceQuotas != nil {
		obj["resource_quotas"] = flattenNamespaceResourceQuotas(in.ResourceQuotas)
	}

	if in.LimitRange != nil {
		obj["limit_range"] = flattenNamespaceLimitRange(in.LimitRange)
	}

	v, ok := obj["artifact"].([]interface{})
	if !ok {
		v = []interface{}{}
	}
	// XXX Debug
	w1 := spew.Sprintf("%+v", v)
	log.Println("flattenNamespaceArtifact before ", w1)

	var ret []interface{}
	ret, err = flattenNamespaceArtifact(&nsat, v)
	if err != nil {
		log.Println("flattenNamespaceArtifact error ", err)
		return nil, err
	}

	// XXX Debug
	w1 = spew.Sprintf("%+v", ret)
	log.Println("flattenNamespaceArtifact after ", w1)

	obj["artifact"] = ret

	return []interface{}{obj}, nil
}

func flattenNamespaceArtifact(nsat *namespaceSpecTranspose, p []interface{}) ([]interface{}, error) {
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	retNil := true
	if len(nsat.Artifact.Repository) > 0 {
		obj["repository"] = nsat.Artifact.Repository
		retNil = false
	}

	if len(nsat.Artifact.Revision) > 0 {
		obj["revision"] = nsat.Artifact.Revision
		retNil = false
	}

	if nsat.Artifact.Path != nil {
		obj["path"] = flattenFile(nsat.Artifact.Path)
		retNil = false
	}

	if retNil {
		return nil, nil
	}

	return []interface{}{obj}, nil

}
