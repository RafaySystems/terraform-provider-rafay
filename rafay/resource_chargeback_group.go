package rafay

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/RafaySystems/rafay-common/pkg/hub/client/options"
	typed "github.com/RafaySystems/rafay-common/pkg/hub/client/typed"
	"github.com/RafaySystems/rafay-common/pkg/hub/terraform/resource"
	"github.com/RafaySystems/rafay-common/proto/types/hub/systempb"
	"github.com/RafaySystems/rctl/pkg/config"
	"github.com/davecgh/go-spew/spew"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceChargebackGroup() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceChargebackGroupCreate,
		ReadContext:   resourceChargebackGroupRead,
		UpdateContext: resourceChargebackGroupUpdate,
		DeleteContext: resourceChargebackGroupDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema:        resource.ChargebackGroupSchema.Schema,
	}
}

func resourceChargebackGroupCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("ChargebackGroup create starts")
	diags := resourceChargebackGroupUpsert(ctx, d, m)
	if diags.HasError() {
		tflog := os.Getenv("TF_LOG")
		if tflog == "TRACE" || tflog == "DEBUG" {
			ctx = context.WithValue(ctx, "debug", "true")
		}
		log.Printf("chargebackGroup create got error, perform cleanup")
		mp, err := expandChargebackGroup(d)
		if err != nil {
			log.Printf("chargebackGroup expandChargebackGroup error")
			return diags
		}
		auth := config.GetConfig().GetAppAuthProfile()
		client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, getUserAgent(), options.WithInsecureSkipVerify(auth.SkipServerCertValid))
		if err != nil {
			return diags
		}

		err = client.SystemV3().ChargebackGroup().Delete(ctx, options.DeleteOptions{
			Name: mp.Metadata.Name,
		})
		if err != nil {
			return diags
		}
	}
	return diags
}

func resourceChargebackGroupUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("chargebackGroup update starts")
	return resourceChargebackGroupUpsert(ctx, d, m)
}

func resourceChargebackGroupUpsert(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("chargebackGroup upsert starts")
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	if d.State() != nil && d.State().ID != "" {
		n := GetMetaName(d)
		if n != "" && n != d.State().ID {
			log.Printf("metadata name change not supported")
			d.State().Tainted = true
			return diag.FromErr(fmt.Errorf("%s", "metadata name change not supported"))
		}
	}

	chargebackGroup, err := expandChargebackGroup(d)
	if err != nil {
		log.Printf("chargebackGroup expandChargebackGroup error")
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, getUserAgent(), options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.SystemV3().ChargebackGroup().Apply(ctx, chargebackGroup, options.ApplyOptions{})
	if err != nil {
		// XXX Debug
		n1 := spew.Sprintf("%+v", chargebackGroup)
		log.Println("chargebackGroup apply chargebackGroup:", n1)
		log.Printf("chargebackGroup apply error")
		return diag.FromErr(err)
	}

	d.SetId(chargebackGroup.Metadata.Name)
	return diags

}

func resourceChargebackGroupRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	log.Println("resourceChargebackGroupRead ")
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	meta := GetMetaData(d)
	if meta == nil {
		return diag.FromErr(fmt.Errorf("%s", "failed to read resource "))
	}
	if d.State() != nil && d.State().ID != "" {
		meta.Name = d.State().ID
	}

	tfChargebackGroupState, err := expandChargebackGroup(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, getUserAgent(), options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return diag.FromErr(err)
	}

	mp, err := client.SystemV3().ChargebackGroup().Get(ctx, options.GetOptions{
		//Name:    tfChargebackGroupState.Metadata.Name,
		Name:    meta.Name,
		Project: tfChargebackGroupState.Metadata.Project,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	err = flattenChargebackGroup(d, mp)
	if err != nil {
		return diag.FromErr(err)
	}
	return diags

}

func resourceChargebackGroupDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	mp, err := expandChargebackGroup(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, getUserAgent(), options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.SystemV3().ChargebackGroup().Delete(ctx, options.DeleteOptions{
		Name: mp.Metadata.Name,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func expandChargebackGroup(in *schema.ResourceData) (*systempb.ChargebackGroup, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "expand chargebackGroup empty input")
	}
	obj := &systempb.ChargebackGroup{}

	if v, ok := in.Get("metadata").([]interface{}); ok && len(v) > 0 {
		obj.Metadata = expandMetaData(v)
	}

	if v, ok := in.Get("spec").([]interface{}); ok && len(v) > 0 {
		objSpec, err := expandChargebackGroupSpec(v)
		if err != nil {
			return nil, err
		}
		log.Println("expandChargebackGroupSpec got spec")
		obj.Spec = objSpec
	}

	obj.ApiVersion = "system.k8smgmt.io/v3"
	obj.Kind = "ChargebackGroup"
	return obj, nil
}

func expandChargebackGroupSpec(p []interface{}) (*systempb.ChargebackGroupSpec, error) {
	obj := &systempb.ChargebackGroupSpec{}
	if len(p) == 0 || p[0] == nil {
		return obj, fmt.Errorf("%s", "expandChargebackGroupSpec empty input")
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["type"].(string); ok && len(v) > 0 {
		obj.Type = v
	}

	if v, ok := in["aggregate"].([]interface{}); ok && len(v) > 0 {
		obj.Aggregate = expandChargebackGroupAggregate(v)
	}

	if v, ok := in["inclusions"].([]interface{}); ok && len(v) > 0 {
		obj.Inclusions = expandChargebackGroupFilter(v)
	}

	if v, ok := in["exclusions"].([]interface{}); ok && len(v) > 0 {
		obj.Exclusions = expandChargebackGroupFilter(v)
	}

	return obj, nil
}

func expandChargebackGroupAggregate(p []interface{}) *systempb.ChargebackAggregate {
	obj := &systempb.ChargebackAggregate{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["project"].(bool); ok {
		obj.Project = v
	}

	if v, ok := in["cluster"].(bool); ok {
		obj.Cluster = v
	}

	if v, ok := in["namespace"].(bool); ok {
		obj.Namespace = v
	}

	if v, ok := in["label"].([]interface{}); ok && len(v) > 0 {
		obj.Label = make([]string, len(v))
		for idx := range v {
			if v[idx] != nil {
				obj.Label[idx] = v[idx].(string)
			}
		}
	}

	return obj

}

func expandChargebackGroupFilter(p []interface{}) []*systempb.ChargebackFilter {
	if len(p) == 0 {
		return []*systempb.ChargebackFilter{}
	}

	out := make([]*systempb.ChargebackFilter, len(p))

	for i := range p {
		if p[i] == nil {
			continue
		}

		obj := systempb.ChargebackFilter{}
		in := p[i].(map[string]interface{})

		if v, ok := in["project"].(string); ok && len(v) > 0 {
			obj.Project = v
		}

		if v, ok := in["cluster"].(string); ok && len(v) > 0 {
			obj.Cluster = v
		}

		if v, ok := in["namespace"].(string); ok && len(v) > 0 {
			obj.Namespace = v
		}

		if v, ok := in["label"].([]interface{}); ok && len(v) > 0 {
			obj.Label = make([]string, len(v))
			for idx := range v {
				obj.Label[idx] = v[idx].(string)
			}
		}

		if v, ok := in["project_name"].(string); ok && len(v) > 0 {
			obj.ProjectName = v
		}

		if v, ok := in["cluster_name"].(string); ok && len(v) > 0 {
			obj.ClusterName = v
		}

		out[i] = &obj
	}

	return out

}

// Flatteners

func flattenChargebackGroup(d *schema.ResourceData, in *systempb.ChargebackGroup) error {
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

	var ret []interface{}
	ret, err = flattenChargebackGroupSpec(in.Spec, v)
	if err != nil {
		return err
	}

	err = d.Set("spec", ret)
	if err != nil {
		return err
	}
	return nil
}

func flattenChargebackGroupSpec(in *systempb.ChargebackGroupSpec, p []interface{}) ([]interface{}, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "flattenChargebackGroupSpec empty input")
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.Type) > 0 {
		obj["type"] = in.Type
	}

	v, ok := obj["aggregate"].([]interface{})
	if !ok {
		v = []interface{}{}
	}
	obj["aggregate"] = flattenChargebackGroupSpecAggregate(in.Aggregate, v)

	v, ok = obj["inclusions"].([]interface{})
	if !ok {
		v = []interface{}{}
	}
	obj["inclusions"] = flattenChargebackGroupSpecFilters(in.Inclusions, v)

	v, ok = obj["exclusions"].([]interface{})
	if !ok {
		v = []interface{}{}
	}
	obj["exclusions"] = flattenChargebackGroupSpecFilters(in.Exclusions, v)

	return []interface{}{obj}, nil
}

func flattenChargebackGroupSpecFilters(in []*systempb.ChargebackFilter, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}

	out := make([]interface{}, len(in))
	for i, in := range in {
		obj := map[string]interface{}{}
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}

		if len(in.Project) > 0 {
			obj["project"] = in.Project
		}

		if len(in.ProjectName) > 0 {
			obj["project_name"] = in.ProjectName
		}

		if len(in.Namespace) > 0 {
			obj["namespace"] = in.Namespace
		}

		if len(in.Label) > 0 {
			obj["label"] = in.Label
		}

		if len(in.Cluster) > 0 {
			obj["cluster"] = in.Cluster
		}

		if len(in.ClusterName) > 0 {
			obj["cluster_name"] = in.ClusterName
		}

		out[i] = obj
	}

	return out
}

func flattenChargebackGroupSpecAggregate(in *systempb.ChargebackAggregate, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}

	retNil := true
	obj := make(map[string]interface{})

	if in.Project {
		obj["project"] = in.Project
		retNil = false
	}

	if in.Cluster {
		obj["cluster"] = in.Cluster
		retNil = false
	}

	if in.Namespace {
		obj["namespace"] = in.Namespace
		retNil = false
	}

	if len(in.Label) > 0 {
		obj["label"] = in.Label
		retNil = false
	}

	if retNil {
		return nil
	}

	return []interface{}{obj}

}
