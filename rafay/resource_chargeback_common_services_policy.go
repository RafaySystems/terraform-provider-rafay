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
	"github.com/RafaySystems/rctl/pkg/versioninfo"
	"github.com/davecgh/go-spew/spew"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceChargebackCommonServicesPolicy() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceChargebackCommonServicesPolicyCreate,
		ReadContext:   resourceChargebackCommonServicesPolicyRead,
		UpdateContext: resourceChargebackCommonServicesPolicyUpdate,
		DeleteContext: resourceChargebackCommonServicesPolicyDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema:        resource.ChargebackCommonServicesPolicySchema.Schema,
	}
}

func resourceChargebackCommonServicesPolicyCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("ChargebackCommonServicesPolicy create starts")
	diags := resourceChargebackCommonServicesPolicyUpsert(ctx, d, m)
	if diags.HasError() {
		tflog := os.Getenv("TF_LOG")
		if tflog == "TRACE" || tflog == "DEBUG" {
			ctx = context.WithValue(ctx, "debug", "true")
		}
		log.Printf("chargebackCommonServicesPolicy create got error, perform cleanup")
		mp, err := expandChargebackCommonServicesPolicy(d)
		if err != nil {
			log.Printf("chargebackCommonServicesPolicy expandChargebackCommonServicesPolicy error")
			return diags
		}
		auth := config.GetConfig().GetAppAuthProfile()
		client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent(), options.WithInsecureSkipVerify(auth.SkipServerCertValid))
		if err != nil {
			return diags
		}

		err = client.SystemV3().ChargebackCommonServicesPolicy().Delete(ctx, options.DeleteOptions{
			Name: mp.Metadata.Name,
		})
		if err != nil {
			return diags
		}
	}
	return diags
}

func resourceChargebackCommonServicesPolicyUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("chargebackCommonServicesPolicy update starts")
	return resourceChargebackCommonServicesPolicyUpsert(ctx, d, m)
}

func resourceChargebackCommonServicesPolicyUpsert(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("chargebackCommonServicesPolicy upsert starts")
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

	chargebackCommonServicesPolicy, err := expandChargebackCommonServicesPolicy(d)
	if err != nil {
		log.Printf("chargebackCommonServicesPolicy expandChargebackCommonServicesPolicy error")
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent(), options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.SystemV3().ChargebackCommonServicesPolicy().Apply(ctx, chargebackCommonServicesPolicy, options.ApplyOptions{})
	if err != nil {
		// XXX Debug
		n1 := spew.Sprintf("%+v", chargebackCommonServicesPolicy)
		log.Println("chargebackCommonServicesPolicy apply chargebackCommonServicesPolicy:", n1)
		log.Printf("chargebackCommonServicesPolicy apply error")
		return diag.FromErr(err)
	}

	d.SetId(chargebackCommonServicesPolicy.Metadata.Name)
	return diags

}

func resourceChargebackCommonServicesPolicyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	log.Println("resourceChargebackCommonServicesPolicyRead ")
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

	tfChargebackCommonServicesPolicyState, err := expandChargebackCommonServicesPolicy(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent(), options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return diag.FromErr(err)
	}

	mp, err := client.SystemV3().ChargebackCommonServicesPolicy().Get(ctx, options.GetOptions{
		//Name:    tfChargebackCommonServicesPolicyState.Metadata.Name,
		Name:    meta.Name,
		Project: tfChargebackCommonServicesPolicyState.Metadata.Project,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	err = flattenChargebackCommonServicesPolicy(d, mp)
	if err != nil {
		return diag.FromErr(err)
	}
	return diags

}

func resourceChargebackCommonServicesPolicyDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	mp, err := expandChargebackCommonServicesPolicy(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent(), options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.SystemV3().ChargebackCommonServicesPolicy().Delete(ctx, options.DeleteOptions{
		Name: mp.Metadata.Name,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func expandChargebackCommonServicesPolicy(in *schema.ResourceData) (*systempb.ChargebackCommonServicesPolicy, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "expand chargebackCommonServicesPolicy empty input")
	}
	obj := &systempb.ChargebackCommonServicesPolicy{}

	if v, ok := in.Get("metadata").([]interface{}); ok && len(v) > 0 {
		obj.Metadata = expandMetaData(v)
	}

	if v, ok := in.Get("spec").([]interface{}); ok && len(v) > 0 {
		objSpec, err := expandChargebackCommonServicesPolicySpec(v)
		if err != nil {
			return nil, err
		}
		log.Println("expandChargebackCommonServicesPolicySpec got spec")
		obj.Spec = objSpec
	}

	obj.ApiVersion = "system.k8smgmt.io/v3"
	obj.Kind = "ChargebackCommonServicesPolicy"
	return obj, nil
}

func expandChargebackCommonServicesPolicySpec(p []interface{}) (*systempb.ChargebackCommonServicesPolicySpec, error) {
	obj := &systempb.ChargebackCommonServicesPolicySpec{}
	if len(p) == 0 || p[0] == nil {
		return obj, fmt.Errorf("%s", "expandChargebackCommonServicesPolicySpec empty input")
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["selection_type"].(string); ok && len(v) > 0 {
		obj.SelectionType = v
	}

	if v, ok := in["policy_project"].(string); ok && len(v) > 0 {
		obj.PolicyProject = v
	}

	if v, ok := in["clusters"].([]interface{}); ok && len(v) > 0 {
		obj.Clusters = make([]string, len(v))
		for idx := range v {
			obj.Clusters[idx] = v[idx].(string)
		}
	}

	if v, ok := in["cluster_labels"].([]interface{}); ok && len(v) > 0 {
		obj.ClusterLabels = expandChargebackPolicyLabels(v)
	}

	if v, ok := in["common_services_namespaces"].([]interface{}); ok && len(v) > 0 {
		obj.CommonServicesNamespaces = make([]string, len(v))
		for idx := range v {
			obj.CommonServicesNamespaces[idx] = v[idx].(string)
		}
	}

	if v, ok := in["common_services_namespace_labels"].([]interface{}); ok && len(v) > 0 {
		obj.CommonServicesNamespaceLabels = expandChargebackPolicyLabels(v)
	}

	return obj, nil
}

func expandChargebackPolicyLabels(p []interface{}) []*systempb.ChargebackPolicyLabels {
	if len(p) == 0 {
		return []*systempb.ChargebackPolicyLabels{}
	}

	out := make([]*systempb.ChargebackPolicyLabels, len(p))

	for i := range p {
		if p[i] == nil {
			continue
		}

		obj := systempb.ChargebackPolicyLabels{}
		in := p[i].(map[string]interface{})

		if v, ok := in["key"].(string); ok && len(v) > 0 {
			obj.Key = v
		}

		if v, ok := in["value"].(string); ok && len(v) > 0 {
			obj.Value = v
		}

		out[i] = &obj
	}

	return out

}

// Flatteners

func flattenChargebackCommonServicesPolicy(d *schema.ResourceData, in *systempb.ChargebackCommonServicesPolicy) error {
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
	ret, err = flattenChargebackCommonServicesPolicySpec(in.Spec, v)
	if err != nil {
		return err
	}

	err = d.Set("spec", ret)
	if err != nil {
		return err
	}
	return nil
}

func flattenChargebackCommonServicesPolicySpec(in *systempb.ChargebackCommonServicesPolicySpec, p []interface{}) ([]interface{}, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "flattenChargebackCommonServicesPolicySpec empty input")
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.SelectionType) > 0 {
		obj["selection_type"] = in.SelectionType
	}

	if len(in.PolicyProject) > 0 {
		obj["policy_project"] = in.PolicyProject
	}

	if len(in.Clusters) > 0 {
		obj["clusters"] = in.Clusters
	}

	v, ok := obj["cluster_labels"].([]interface{})
	if !ok {
		v = []interface{}{}
	}
	obj["cluster_labels"] = flattenChargebackCommonServicesPolicyLabels(in.ClusterLabels, v)

	if len(in.CommonServicesNamespaces) > 0 {
		obj["common_services_namespaces"] = in.CommonServicesNamespaces
	}

	v, ok = obj["common_services_namespace_labels"].([]interface{})
	if !ok {
		v = []interface{}{}
	}
	obj["common_services_namespace_labels"] = flattenChargebackCommonServicesPolicyLabels(in.CommonServicesNamespaceLabels, v)

	return []interface{}{obj}, nil
}

func flattenChargebackCommonServicesPolicyLabels(in []*systempb.ChargebackPolicyLabels, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}

	out := make([]interface{}, len(in))
	for i, in := range in {
		obj := map[string]interface{}{}
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}

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
