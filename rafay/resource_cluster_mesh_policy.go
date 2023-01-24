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

func resourceClusterMeshPolicy() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceClusterMeshPolicyCreate,
		ReadContext:   resourceClusterMeshPolicyRead,
		UpdateContext: resourceClusterMeshPolicyUpdate,
		DeleteContext: resourceClusterMeshPolicyDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema:        resource.ClusterMeshPolicySchema.Schema,
	}
}

func resourceClusterMeshPolicyCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("ClusterMeshPolicy create starts")
	diags := resourceClusterMeshPolicyUpsert(ctx, d, m)
	if diags.HasError() {
		tflog := os.Getenv("TF_LOG")
		if tflog == "TRACE" || tflog == "DEBUG" {
			ctx = context.WithValue(ctx, "debug", "true")
		}
		log.Printf("clusterMeshPolicy create got error, perform cleanup")
		nnp, err := expandClusterMeshPolicy(d)
		if err != nil {
			log.Printf("clusterMeshPolicy expandClusterMeshPolicy error")
			return diags
		}
		auth := config.GetConfig().GetAppAuthProfile()
		client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
		if err != nil {
			return diags
		}

		err = client.ServicemeshV3().ClusterMeshPolicy().Delete(ctx, options.DeleteOptions{
			Name:    nnp.Metadata.Name,
			Project: nnp.Metadata.Project,
		})
		if err != nil {
			return diags
		}
	}
	return diags
}

func resourceClusterMeshPolicyUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("clusterMeshPolicy update starts")
	return resourceClusterMeshPolicyUpsert(ctx, d, m)
}

func resourceClusterMeshPolicyUpsert(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("clusterMeshPolicy upsert starts")
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

	clusterMeshPolicy, err := expandClusterMeshPolicy(d)
	if err != nil {
		log.Printf("clusterMeshPolicy expandClusterMeshPolicy error")
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.ServicemeshV3().ClusterMeshPolicy().Apply(ctx, clusterMeshPolicy, options.ApplyOptions{})
	if err != nil {
		// XXX Debug
		n1 := spew.Sprintf("%+v", clusterMeshPolicy)
		log.Println("clusterMeshPolicy apply clusterMeshPolicy:", n1)
		log.Printf("clusterMeshPolicy apply error")
		return diag.FromErr(err)
	}

	d.SetId(clusterMeshPolicy.Metadata.Name)
	return diags

}

func resourceClusterMeshPolicyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	log.Println("resourceClusterMeshPolicyRead ")
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

	tfClusterMeshPolicyState, err := expandClusterMeshPolicy(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}

	nnp, err := client.ServicemeshV3().ClusterMeshPolicy().Get(ctx, options.GetOptions{
		//Name:    tfClusterMeshPolicyState.Metadata.Name,
		Name:    meta.Name,
		Project: tfClusterMeshPolicyState.Metadata.Project,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	err = flattenClusterMeshPolicy(d, nnp)
	if err != nil {
		return diag.FromErr(err)
	}
	return diags

}

func resourceClusterMeshPolicyDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	nnp, err := expandClusterMeshPolicy(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.ServicemeshV3().ClusterMeshPolicy().Delete(ctx, options.DeleteOptions{
		Name:    nnp.Metadata.Name,
		Project: nnp.Metadata.Project,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func expandClusterMeshPolicy(in *schema.ResourceData) (*servicemeshpb.ClusterMeshPolicy, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "expand clusterMeshPolicy empty input")
	}
	obj := &servicemeshpb.ClusterMeshPolicy{}

	if v, ok := in.Get("metadata").([]interface{}); ok && len(v) > 0 {
		obj.Metadata = expandMetaData(v)
	}

	if v, ok := in.Get("spec").([]interface{}); ok && len(v) > 0 {
		objSpec, err := expandClusterMeshPolicySpec(v)
		if err != nil {
			return nil, err
		}
		log.Println("expandClusterMeshPolicySpec got spec")
		obj.Spec = objSpec
	}

	obj.ApiVersion = "servicemesh.k8smgmt.io/v3"
	obj.Kind = "ClusterMeshPolicy"
	return obj, nil
}

func expandClusterMeshPolicySpec(p []interface{}) (*servicemeshpb.ClusterMeshPolicySpec, error) {
	obj := &servicemeshpb.ClusterMeshPolicySpec{}
	if len(p) == 0 || p[0] == nil {
		return obj, fmt.Errorf("%s", "expandClusterMeshPolicySpec empty input")
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["rules"].([]interface{}); ok && len(v) > 0 {
		obj.Rules = expandClusterMeshPolicySpecRules(v)
	}

	if v, ok := in["sharing"].([]interface{}); ok && len(v) > 0 {
		obj.Sharing = expandSharingSpec(v)
	}

	if v, ok := in["version"].(string); ok && len(v) > 0 {
		obj.Version = v
	}

	return obj, nil
}

func expandClusterMeshPolicySpecRules(p []interface{}) []*commonpb.ResourceNameAndVersionRef {
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

func flattenClusterMeshPolicy(d *schema.ResourceData, in *servicemeshpb.ClusterMeshPolicy) error {
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
	ret, err = flattenClusterMeshPolicySpec(in.Spec, v)
	if err != nil {
		return err
	}

	err = d.Set("spec", ret)
	if err != nil {
		return err
	}
	return nil
}

func flattenClusterMeshPolicySpec(in *servicemeshpb.ClusterMeshPolicySpec, p []interface{}) ([]interface{}, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "flattenClusterMeshPolicySpec empty input")
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
		obj["rules"] = flattenClusterMeshPolicySpecs(in.Rules, v)
	}

	obj["sharing"] = flattenSharingSpec(in.Sharing)

	if len(in.Version) > 0 {
		obj["version"] = in.Version
	}

	return []interface{}{obj}, nil
}

func flattenClusterMeshPolicySpecs(in []*commonpb.ResourceNameAndVersionRef, p []interface{}) []interface{} {
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
