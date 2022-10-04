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
	"github.com/RafaySystems/rafay-common/proto/types/hub/securitypb"
	"github.com/RafaySystems/rctl/pkg/config"
	"github.com/RafaySystems/rctl/pkg/versioninfo"
	"github.com/davecgh/go-spew/spew"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceNamespaceNetworkPolicy() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceNamespaceNetworkPolicyCreate,
		ReadContext:   resourceNamespaceNetworkPolicyRead,
		UpdateContext: resourceNamespaceNetworkPolicyUpdate,
		DeleteContext: resourceNamespaceNetworkPolicyDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema:        resource.NamespaceNetworkPolicySchema.Schema,
	}
}

func resourceNamespaceNetworkPolicyCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("NamespaceNetworkPolicy create starts")
	diags := resourceNamespaceNetworkPolicyUpsert(ctx, d, m)
	if diags.HasError() {
		tflog := os.Getenv("TF_LOG")
		if tflog == "TRACE" || tflog == "DEBUG" {
			ctx = context.WithValue(ctx, "debug", "true")
		}
		log.Printf("namespaceNetworkPolicy create got error, perform cleanup")
		nnp, err := expandNamespaceNetworkPolicy(d)
		if err != nil {
			log.Printf("namespaceNetworkPolicy expandNamespaceNetworkPolicy error")
			return diags
		}
		auth := config.GetConfig().GetAppAuthProfile()
		client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
		if err != nil {
			return diags
		}

		err = client.SecurityV3().NamespaceNetworkPolicy().Delete(ctx, options.DeleteOptions{
			Name:    nnp.Metadata.Name,
			Project: nnp.Metadata.Project,
		})
		if err != nil {
			return diags
		}
	}
	return diags
}

func resourceNamespaceNetworkPolicyUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("namespaceNetworkPolicy update starts")
	return resourceNamespaceNetworkPolicyUpsert(ctx, d, m)
}

func resourceNamespaceNetworkPolicyUpsert(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("namespaceNetworkPolicy upsert starts")
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

	namespaceNetworkPolicy, err := expandNamespaceNetworkPolicy(d)
	if err != nil {
		log.Printf("namespaceNetworkPolicy expandNamespaceNetworkPolicy error")
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.SecurityV3().NamespaceNetworkPolicy().Apply(ctx, namespaceNetworkPolicy, options.ApplyOptions{})
	if err != nil {
		// XXX Debug
		n1 := spew.Sprintf("%+v", namespaceNetworkPolicy)
		log.Println("namespaceNetworkPolicy apply namespaceNetworkPolicy:", n1)
		log.Printf("namespaceNetworkPolicy apply error")
		return diag.FromErr(err)
	}

	d.SetId(namespaceNetworkPolicy.Metadata.Name)
	return diags

}

func resourceNamespaceNetworkPolicyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	log.Println("resourceNamespaceNetworkPolicyRead ")
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

	tfNamespaceNetworkPolicyState, err := expandNamespaceNetworkPolicy(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}

	nnp, err := client.SecurityV3().NamespaceNetworkPolicy().Get(ctx, options.GetOptions{
		//Name:    tfNamespaceNetworkPolicyState.Metadata.Name,
		Name:    meta.Name,
		Project: tfNamespaceNetworkPolicyState.Metadata.Project,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	err = flattenNamespaceNetworkPolicy(d, nnp)
	if err != nil {
		return diag.FromErr(err)
	}
	return diags

}

func resourceNamespaceNetworkPolicyDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	nnp, err := expandNamespaceNetworkPolicy(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.SecurityV3().NamespaceNetworkPolicy().Delete(ctx, options.DeleteOptions{
		Name:    nnp.Metadata.Name,
		Project: nnp.Metadata.Project,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func expandNamespaceNetworkPolicy(in *schema.ResourceData) (*securitypb.NamespaceNetworkPolicy, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "expand namespaceNetworkPolicy empty input")
	}
	obj := &securitypb.NamespaceNetworkPolicy{}

	if v, ok := in.Get("metadata").([]interface{}); ok && len(v) > 0 {
		obj.Metadata = expandMetaData(v)
	}

	if v, ok := in.Get("spec").([]interface{}); ok && len(v) > 0 {
		objSpec, err := expandNamespaceNetworkPolicySpec(v)
		if err != nil {
			return nil, err
		}
		log.Println("expandNamespaceNetworkPolicySpec got spec")
		obj.Spec = objSpec
	}

	obj.ApiVersion = "security.k8smgmt.io/v3"
	obj.Kind = "NamespaceNetworkPolicy"
	return obj, nil
}

func expandNamespaceNetworkPolicySpec(p []interface{}) (*securitypb.NamespaceNetworkPolicySpec, error) {
	obj := &securitypb.NamespaceNetworkPolicySpec{}
	if len(p) == 0 || p[0] == nil {
		return obj, fmt.Errorf("%s", "expandNamespaceNetworkPolicySpec empty input")
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["rules"].([]interface{}); ok && len(v) > 0 {
		obj.Rules = expandNamespaceNetworkPolicySpecRules(v)
	}

	if v, ok := in["sharing"].([]interface{}); ok && len(v) > 0 {
		obj.Sharing = expandSharingSpec(v)
	}

	if v, ok := in["version"].(string); ok && len(v) > 0 {
		obj.Version = v
	}

	return obj, nil
}

func expandNamespaceNetworkPolicySpecRules(p []interface{}) []*commonpb.ResourceNameAndVersionRef {
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

func flattenNamespaceNetworkPolicy(d *schema.ResourceData, in *securitypb.NamespaceNetworkPolicy) error {
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
	ret, err = flattenNamespaceNetworkPolicySpec(in.Spec, v)
	if err != nil {
		return err
	}

	err = d.Set("spec", ret)
	if err != nil {
		return err
	}
	return nil
}

func flattenNamespaceNetworkPolicySpec(in *securitypb.NamespaceNetworkPolicySpec, p []interface{}) ([]interface{}, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "flattenNamespaceNetworkPolicySpec empty input")
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
		obj["rules"] = flattenNamespaceNetworkPolicySpecs(in.Rules, v)
	}

	if in.Sharing != nil {
		obj["sharing"] = flattenSharingSpec(in.Sharing)
	}

	if len(in.Version) > 0 {
		obj["version"] = in.Version
	}

	return []interface{}{obj}, nil
}

func flattenNamespaceNetworkPolicySpecs(in []*commonpb.ResourceNameAndVersionRef, p []interface{}) []interface{} {
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
