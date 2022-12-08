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

func resourceClusterNetworkPolicy() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceClusterNetworkPolicyCreate,
		ReadContext:   resourceClusterNetworkPolicyRead,
		UpdateContext: resourceClusterNetworkPolicyUpdate,
		DeleteContext: resourceClusterNetworkPolicyDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema:        resource.ClusterNetworkPolicySchema.Schema,
	}
}

func resourceClusterNetworkPolicyCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("ClusterNetworkPolicy create starts")
	diags := resourceClusterNetworkPolicyUpsert(ctx, d, m)
	if diags.HasError() {
		tflog := os.Getenv("TF_LOG")
		if tflog == "TRACE" || tflog == "DEBUG" {
			ctx = context.WithValue(ctx, "debug", "true")
		}
		log.Printf("clusterNetworkPolicy create got error, perform cleanup")
		cnp, err := expandClusterNetworkPolicy(d)
		if err != nil {
			log.Printf("clusterNetworkPolicy expandClusterNetworkPolicy error")
			return diags
		}
		auth := config.GetConfig().GetAppAuthProfile()
		client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
		if err != nil {
			return diags
		}

		err = client.SecurityV3().ClusterNetworkPolicy().Delete(ctx, options.DeleteOptions{
			Name:    cnp.Metadata.Name,
			Project: cnp.Metadata.Project,
		})
		if err != nil {
			return diags
		}
	}
	return diags
}

func resourceClusterNetworkPolicyUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("clusterNetworkPolicy update starts")
	return resourceClusterNetworkPolicyUpsert(ctx, d, m)
}

func resourceClusterNetworkPolicyUpsert(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("clusterNetworkPolicy upsert starts")
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

	clusterNetworkPolicy, err := expandClusterNetworkPolicy(d)
	if err != nil {
		log.Printf("clusterNetworkPolicy expandClusterNetworkPolicy error")
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.SecurityV3().ClusterNetworkPolicy().Apply(ctx, clusterNetworkPolicy, options.ApplyOptions{})
	if err != nil {
		// XXX Debug
		n1 := spew.Sprintf("%+v", clusterNetworkPolicy)
		log.Println("clusterNetworkPolicy apply clusterNetworkPolicy:", n1)
		log.Printf("clusterNetworkPolicy apply error")
		return diag.FromErr(err)
	}

	d.SetId(clusterNetworkPolicy.Metadata.Name)
	return diags

}

func resourceClusterNetworkPolicyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	log.Println("resourceClusterNetworkPolicyRead ")
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

	tfClusterNetworkPolicyState, err := expandClusterNetworkPolicy(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}

	cnp, err := client.SecurityV3().ClusterNetworkPolicy().Get(ctx, options.GetOptions{
		//Name:    tfClusterNetworkPolicyState.Metadata.Name,
		Name:    meta.Name,
		Project: tfClusterNetworkPolicyState.Metadata.Project,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	err = flattenClusterNetworkPolicy(d, cnp)
	if err != nil {
		return diag.FromErr(err)
	}
	return diags

}

func resourceClusterNetworkPolicyDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	cnp, err := expandClusterNetworkPolicy(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.SecurityV3().ClusterNetworkPolicy().Delete(ctx, options.DeleteOptions{
		Name:    cnp.Metadata.Name,
		Project: cnp.Metadata.Project,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func expandClusterNetworkPolicy(in *schema.ResourceData) (*securitypb.ClusterNetworkPolicy, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "expand clusterNetworkPolicy empty input")
	}
	obj := &securitypb.ClusterNetworkPolicy{}

	if v, ok := in.Get("metadata").([]interface{}); ok && len(v) > 0 {
		obj.Metadata = expandMetaData(v)
	}

	if v, ok := in.Get("spec").([]interface{}); ok && len(v) > 0 {
		objSpec, err := expandClusterNetworkPolicySpec(v)
		if err != nil {
			return nil, err
		}
		log.Println("expandClusterNetworkPolicySpec got spec")
		obj.Spec = objSpec
	}

	obj.ApiVersion = "security.k8smgmt.io/v3"
	obj.Kind = "ClusterNetworkPolicy"
	return obj, nil
}

func expandClusterNetworkPolicySpec(p []interface{}) (*securitypb.ClusterNetworkPolicySpec, error) {
	obj := &securitypb.ClusterNetworkPolicySpec{}
	if len(p) == 0 || p[0] == nil {
		return obj, fmt.Errorf("%s", "expandClusterNetworkPolicySpec empty input")
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["rules"].([]interface{}); ok && len(v) > 0 {
		obj.Rules = expandClusterNetworkPolicySpecRules(v)
	}

	if v, ok := in["sharing"].([]interface{}); ok && len(v) > 0 {
		obj.Sharing = expandSharingSpec(v)
	}

	if v, ok := in["version"].(string); ok && len(v) > 0 {
		obj.Version = v
	}

	return obj, nil
}

func expandClusterNetworkPolicySpecRules(p []interface{}) []*commonpb.ResourceNameAndVersionRef {
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

func flattenClusterNetworkPolicy(d *schema.ResourceData, in *securitypb.ClusterNetworkPolicy) error {
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
	ret, err = flattenClusterNetworkPolicySpec(in.Spec, v)
	if err != nil {
		return err
	}

	err = d.Set("spec", ret)
	if err != nil {
		return err
	}
	return nil
}

func flattenClusterNetworkPolicySpec(in *securitypb.ClusterNetworkPolicySpec, p []interface{}) ([]interface{}, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "flattenClusterNetworkPolicySpec empty input")
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
		obj["rules"] = flattenClusterNetworkPolicySpecRules(in.Rules, v)
	}

	obj["sharing"] = flattenSharingSpec(in.Sharing)

	if len(in.Version) > 0 {
		obj["version"] = in.Version
	}

	return []interface{}{obj}, nil
}

func flattenClusterNetworkPolicySpecRules(in []*commonpb.ResourceNameAndVersionRef, p []interface{}) []interface{} {
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
