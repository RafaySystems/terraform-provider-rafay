package rafay

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/RafaySystems/rafay-common/pkg/hub/client/options"
	typed "github.com/RafaySystems/rafay-common/pkg/hub/client/typed"
	"github.com/RafaySystems/rafay-common/pkg/hub/terraform/resource"
	commonpb "github.com/RafaySystems/rafay-common/proto/types/hub/commonpb"
	"github.com/RafaySystems/rafay-common/proto/types/hub/infrapb"
	"github.com/RafaySystems/rctl/pkg/config"
	"github.com/RafaySystems/rctl/pkg/namespace"
	"github.com/RafaySystems/rctl/pkg/project"
	"github.com/RafaySystems/rctl/pkg/versioninfo"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

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
	}
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
	return resourceNamespaceUpsert(ctx, d, m)
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
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.InfraV3().Namespace().Apply(ctx, ns, options.ApplyOptions{})
	if err != nil {
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
	}

	d.SetId(ns.Metadata.Name)
	return diags

}

func resourceNamespaceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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

	ns, err := client.InfraV3().Namespace().Get(ctx, options.GetOptions{
		Name:    nsTFState.Metadata.Name,
		Project: nsTFState.Metadata.Project,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	// XXX Debug
	// n1 = spew.Sprintf("%+v", ns)
	// log.Println("resourceNamespaceRead ns", n1)

	err = flattenNamespace(d, ns)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceNamespaceDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	//get project id with project name, p.id used to refer to project id -> need p.ID for calling Create Namespace
	resp, err := project.GetProjectByName(d.Get("projectname").(string))
	if err != nil {
		log.Printf("project does not exist, error %s", err.Error())
		return diag.FromErr(err)
	}
	p, err := project.NewProjectFromResponse([]byte(resp))
	if err != nil {
		return diag.FromErr(err)
	} else if p == nil {
		d.SetId("")
		return diags
	}
	projectId := p.ID
	//Delete namespace
	err = namespace.DeleteNamespace(string(d.Id()), projectId)
	if err != nil {
		log.Println("error delete namespace: ", err)
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

	if v, ok := in.Get("spec").([]interface{}); ok {
		objSpec, err := expandNamespaceSpec(v)
		if err != nil {
			return nil, err
		}
		obj.Spec = objSpec
	}

	obj.ApiVersion = "apps.k8smgmt.io/v3"
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
		nst.ResourceQuotas = expandNamespaceResourceQuotas(v)
	}

	if v, ok := in["limit_range"].([]interface{}); ok {
		nst.LimitRange = expandNamespaceLimitRange(v)
	}

	if vp, ok := in["artifact"].([]interface{}); ok && len(vp) > 0 {
		if len(vp) == 0 || vp[0] == nil {
			return nil, fmt.Errorf("%s", "expandArtifact empty artifact")
		}
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
	// XXX Debug
	// s := spew.Sprintf("%+v", at)
	// log.Println("expandNamespaceSpec at", s)

	jsonSpec, err := json.Marshal(nst)
	if err != nil {
		return nil, err
	}

	// XXX Debug
	// log.Println("expandNamespaceSpec jsonSpec ", string(jsonSpec))

	err = obj.UnmarshalJSON(jsonSpec)
	if err != nil {
		log.Println("expandNamespaceSpec artifact UnmarshalJSON error ", err)
		return nil, err
	}

	// XXX Debug
	// s1 := spew.Sprintf("%+v", obj)
	// log.Println("expandNamespaceSpec obj", s1)

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
		obj.Requests = expandResourceQuantity(v)
	}

	if v, ok := in["limits"].([]interface{}); ok {
		obj.Limits = expandResourceQuantity(v)
	}
	return obj
}

func expandNamespaceLimitRange(p []interface{}) *infrapb.NamespaceLimitRange {
	obj := &infrapb.NamespaceLimitRange{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}

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

	if v, ok := in["ratio"].([]interface{}); ok {
		obj.Ratio = expandResourceRatio(v)
	}

	return obj
}

func expandResourceRatio(p []interface{}) *commonpb.ResourceRatio {
	obj := &commonpb.ResourceRatio{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}

	in := p[0].(map[string]interface{})
	if v, ok := in["memory"].(float32); ok {
		obj.Memory = v
	}

	if v, ok := in["cpu"].(float32); ok {
		obj.Cpu = v
	}

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
	// w1 := spew.Sprintf("%+v", v)
	// log.Println("flattenWorkload before ", w1)
	var ret []interface{}
	ret, err = flattenNamespaceSpec(in.Spec, v)
	if err != nil {
		return err
	}
	// XXX Debug
	// w1 = spew.Sprintf("%+v", ret)
	// log.Println("flattenWorkload after ", w1)

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

	obj := make(map[string]interface{})

	if in.Requests != nil {
		obj["requests"] = flattenResourceQuantity(in.Requests)
	}

	if in.Limits != nil {
		obj["limits"] = flattenResourceQuantity(in.Limits)
	}

	return []interface{}{obj}
}

func flattenNamespaceLimitRange(in *infrapb.NamespaceLimitRange) []interface{} {
	if in == nil {
		return nil
	}

	obj := make(map[string]interface{})

	if in.Pod != nil {
		obj["pod"] = flattenNamespaceLimitRangeConfig(in.Pod)
	}

	if in.Container != nil {
		obj["container"] = flattenNamespaceLimitRangeConfig(in.Container)
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
		log.Println("FlattenArtifactSpec MarshalJSON error", err)
		return nil, fmt.Errorf("%s %+v", "FlattenArtifactSpec MarshalJSON error", err)
	}

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
	// w1 := spew.Sprintf("%+v", v)
	// log.Println("flattenNamespaceSpec before ", w1)

	var ret []interface{}
	ret, err = flattenNamespaceArtifact(&nsat, v)
	if err != nil {
		log.Println("flattenNamespaceSpec error ", err)
		return nil, err
	}

	// XXX Debug
	// w1 = spew.Sprintf("%+v", ret)
	// log.Println("flattenNamespaceSpec after ", w1)

	obj["artifact"] = ret

	return []interface{}{obj}, nil
}

func flattenNamespaceArtifact(nsat *namespaceSpecTranspose, p []interface{}) ([]interface{}, error) {
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(nsat.Artifact.Repository) > 0 {
		obj["repository"] = nsat.Artifact.Repository
	}

	if len(nsat.Artifact.Revision) > 0 {
		obj["revision"] = nsat.Artifact.Revision
	}

	if nsat.Artifact.Path != nil {
		obj["path"] = flattenFile(nsat.Artifact.Path)
	}

	return []interface{}{obj}, nil

}
