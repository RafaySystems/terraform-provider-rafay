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
	"github.com/RafaySystems/rafay-common/proto/types/hub/commonpb"
	"github.com/RafaySystems/rafay-common/proto/types/hub/servicemeshpb"
	"github.com/RafaySystems/rctl/pkg/config"
	"github.com/RafaySystems/rctl/pkg/versioninfo"
	"github.com/davecgh/go-spew/spew"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceNamespaceMeshPolicy() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceNamespaceMeshPolicyCreate,
		ReadContext:   resourceNamespaceMeshPolicyRead,
		UpdateContext: resourceNamespaceMeshPolicyUpdate,
		DeleteContext: resourceNamespaceMeshPolicyDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema:        resource.NamespaceMeshPolicySchema.Schema,
	}
}

func resourceNamespaceMeshPolicyCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("NamespaceMeshPolicy create starts")
	diags := resourceNamespaceMeshPolicyUpsert(ctx, d, m)
	if diags.HasError() {
		tflog := os.Getenv("TF_LOG")
		if tflog == "TRACE" || tflog == "DEBUG" {
			ctx = context.WithValue(ctx, "debug", "true")
		}
		log.Printf("namespaceMeshPolicy create got error, perform cleanup")
		nnp, err := expandNamespaceMeshPolicy(d)
		if err != nil {
			log.Printf("namespaceMeshPolicy expandNamespaceMeshPolicy error")
			return diags
		}
		auth := config.GetConfig().GetAppAuthProfile()
		client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
		if err != nil {
			return diags
		}

		err = client.ServicemeshV3().NamespaceMeshPolicy().Delete(ctx, options.DeleteOptions{
			Name:    nnp.Metadata.Name,
			Project: nnp.Metadata.Project,
		})
		if err != nil {
			return diags
		}
	}
	return diags
}

func resourceNamespaceMeshPolicyUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("namespaceMeshPolicy update starts")
	return resourceNamespaceMeshPolicyUpsert(ctx, d, m)
}

func resourceNamespaceMeshPolicyUpsert(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("namespaceMeshPolicy upsert starts")
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

	namespaceMeshPolicy, err := expandNamespaceMeshPolicy(d)
	if err != nil {
		log.Printf("namespaceMeshPolicy expandNamespaceMeshPolicy error")
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.ServicemeshV3().NamespaceMeshPolicy().Apply(ctx, namespaceMeshPolicy, options.ApplyOptions{})
	if err != nil {
		// XXX Debug
		n1 := spew.Sprintf("%+v", namespaceMeshPolicy)
		log.Println("namespaceMeshPolicy apply namespaceMeshPolicy:", n1)
		log.Printf("namespaceMeshPolicy apply error")
		return diag.FromErr(err)
	}

	d.SetId(namespaceMeshPolicy.Metadata.Name)
	return diags

}

func resourceNamespaceMeshPolicyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	log.Println("resourceNamespaceMeshPolicyRead ")
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

	tfNamespaceMeshPolicyState, err := expandNamespaceMeshPolicy(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}

	nnp, err := client.ServicemeshV3().NamespaceMeshPolicy().Get(ctx, options.GetOptions{
		//Name:    tfNamespaceMeshPolicyState.Metadata.Name,
		Name:    meta.Name,
		Project: tfNamespaceMeshPolicyState.Metadata.Project,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	err = flattenNamespaceMeshPolicy(d, nnp)
	if err != nil {
		return diag.FromErr(err)
	}
	return diags

}

func resourceNamespaceMeshPolicyDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	nnp, err := expandNamespaceMeshPolicy(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.ServicemeshV3().NamespaceMeshPolicy().Delete(ctx, options.DeleteOptions{
		Name:    nnp.Metadata.Name,
		Project: nnp.Metadata.Project,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func expandNamespaceMeshPolicy(in *schema.ResourceData) (*servicemeshpb.NamespaceMeshPolicy, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "expand namespaceMeshPolicy empty input")
	}
	obj := &servicemeshpb.NamespaceMeshPolicy{}

	if v, ok := in.Get("metadata").([]interface{}); ok && len(v) > 0 {
		obj.Metadata = expandMetaData(v)
	}

	if v, ok := in.Get("spec").([]interface{}); ok && len(v) > 0 {
		objSpec, err := expandNamespaceMeshPolicySpec(v)
		if err != nil {
			return nil, err
		}
		log.Println("expandNamespaceMeshPolicySpec got spec")
		obj.Spec = objSpec
	}

	obj.ApiVersion = "servicemesh.k8smgmt.io/v3"
	obj.Kind = "NamespaceMeshPolicy"
	return obj, nil
}

func expandNamespaceMeshPolicySpec(p []interface{}) (*servicemeshpb.NamespaceMeshPolicySpec, error) {
	obj := &servicemeshpb.NamespaceMeshPolicySpec{}
	if len(p) == 0 || p[0] == nil {
		return obj, fmt.Errorf("%s", "expandNamespaceMeshPolicySpec empty input")
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["rules"].([]interface{}); ok && len(v) > 0 {
		obj.Rules = expandNamespaceMeshPolicySpecRules(v)
	}

	if v, ok := in["sharing"].([]interface{}); ok && len(v) > 0 {
		obj.Sharing = expandSharingSpec(v)
	}

	if v, ok := in["version"].(string); ok && len(v) > 0 {
		obj.Version = v
	}

	return obj, nil
}

func expandNamespaceMeshPolicySpecRules(p []interface{}) []*commonpb.ResourceNameAndVersionRef {
	if len(p) == 0 || p[0] == nil {
		return []*commonpb.ResourceNameAndVersionRef{}
	}

	out := make([]*commonpb.ResourceNameAndVersionRef, len(p))

	for i := range p {
		obj := commonpb.ResourceNameAndVersionRef{}
		in := p[i].(map[string]interface{})

		if v, ok := in["name"].(string); ok {
			obj.Name = v
		}

		if v, ok := in["version"].(string); ok {
			obj.Version = v
		}

		out[i] = &obj
	}

	return out

}

// Flatteners

func flattenNamespaceMeshPolicy(d *schema.ResourceData, in *servicemeshpb.NamespaceMeshPolicy) error {
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
	ret, err = flattenNamespaceMeshPolicySpec(in.Spec, v)
	if err != nil {
		return err
	}

	err = d.Set("spec", ret)
	if err != nil {
		return err
	}
	return nil
}

func flattenNamespaceMeshPolicySpec(in *servicemeshpb.NamespaceMeshPolicySpec, p []interface{}) ([]interface{}, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "flattenNamespaceMeshPolicySpec empty input")
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if in.Rules != nil {
		v, ok := obj["rules"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["rules"] = flattenNamespaceMeshPolicySpecs(in.Rules, v)
	}

	if in.Sharing != nil {
		obj["sharing"] = flattenSharingSpec(in.Sharing)
	}

	if len(in.Version) > 0 {
		obj["version"] = in.Version
	}

	return []interface{}{obj}, nil
}

func flattenNamespaceMeshPolicySpecs(in []*commonpb.ResourceNameAndVersionRef, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}

	out := make([]interface{}, len(in))
	for i, in := range in {

		obj := map[string]interface{}{}
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}

		if len(in.Name) > 0 {
			obj["name"] = in.Name
		}

		if len(in.Version) > 0 {
			obj["version"] = in.Version
		}

		out[i] = &obj
	}

	return out
}
