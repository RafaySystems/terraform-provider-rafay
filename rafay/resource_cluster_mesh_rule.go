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
	"github.com/RafaySystems/rafay-common/proto/types/hub/servicemeshpb"
	"github.com/RafaySystems/rctl/pkg/config"
	"github.com/RafaySystems/rctl/pkg/versioninfo"
	"github.com/davecgh/go-spew/spew"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceClusterMeshRule() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceClusterMeshRuleCreate,
		ReadContext:   resourceClusterMeshRuleRead,
		UpdateContext: resourceClusterMeshRuleUpdate,
		DeleteContext: resourceClusterMeshRuleDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema:        resource.ClusterMeshRuleSchema.Schema,
	}
}

func resourceClusterMeshRuleCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("ClusterMeshRule create starts")
	diags := resourceClusterMeshRuleUpsert(ctx, d, m)
	if diags.HasError() {
		tflog := os.Getenv("TF_LOG")
		if tflog == "TRACE" || tflog == "DEBUG" {
			ctx = context.WithValue(ctx, "debug", "true")
		}
		log.Printf("clusterMeshRule create got error, perform cleanup")
		cnpr, err := expandClusterMeshRule(d)
		if err != nil {
			log.Printf("clusterMeshRule expandClusterMeshRule error")
			return diags
		}
		auth := config.GetConfig().GetAppAuthProfile()
		client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
		if err != nil {
			return diags
		}

		err = client.ServicemeshV3().ClusterMeshRule().Delete(ctx, options.DeleteOptions{
			Name:    cnpr.Metadata.Name,
			Project: cnpr.Metadata.Project,
		})
		if err != nil {
			return diags
		}
	}
	return diags
}

func resourceClusterMeshRuleUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("clusterMeshRule update starts")
	return resourceClusterMeshRuleUpsert(ctx, d, m)
}

func resourceClusterMeshRuleUpsert(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("clusterMeshRule upsert starts")
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

	clusterMeshRule, err := expandClusterMeshRule(d)
	if err != nil {
		log.Printf("clusterMeshRule expandClusterMeshRule error")
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.ServicemeshV3().ClusterMeshRule().Apply(ctx, clusterMeshRule, options.ApplyOptions{})
	if err != nil {
		// XXX Debug
		n1 := spew.Sprintf("%+v", clusterMeshRule)
		log.Println("clusterMeshRule apply clusterMeshRule:", n1)
		log.Printf("clusterMeshRule apply error")
		return diag.FromErr(err)
	}

	d.SetId(clusterMeshRule.Metadata.Name)
	return diags

}

func resourceClusterMeshRuleRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	log.Println("resourceClusterMeshRuleRead ")
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

	tfClusterMeshRuleState, err := expandClusterMeshRule(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}

	cnp, err := client.ServicemeshV3().ClusterMeshRule().Get(ctx, options.GetOptions{
		//Name:    tfClusterMeshRuleState.Metadata.Name,
		Name:    meta.Name,
		Project: tfClusterMeshRuleState.Metadata.Project,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	err = flattenClusterMeshRule(d, cnp)
	if err != nil {
		return diag.FromErr(err)
	}
	return diags

}

func resourceClusterMeshRuleDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	cnpr, err := expandClusterMeshRule(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.ServicemeshV3().ClusterMeshRule().Delete(ctx, options.DeleteOptions{
		Name:    cnpr.Metadata.Name,
		Project: cnpr.Metadata.Project,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func expandClusterMeshRule(in *schema.ResourceData) (*servicemeshpb.ClusterMeshRule, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "expand clusterMeshRule empty input")
	}
	obj := &servicemeshpb.ClusterMeshRule{}

	if v, ok := in.Get("metadata").([]interface{}); ok && len(v) > 0 {
		obj.Metadata = expandMetaData(v)
	}

	if v, ok := in.Get("spec").([]interface{}); ok && len(v) > 0 {
		objSpec, err := expandClusterMeshRuleSpec(v)
		if err != nil {
			return nil, err
		}
		log.Println("expandClusterMeshRuleSpec got spec")
		obj.Spec = objSpec
	}

	obj.ApiVersion = "servicemesh.k8smgmt.io/v3"
	obj.Kind = "ClusterMeshRule"
	return obj, nil
}

func expandClusterMeshRuleSpec(p []interface{}) (*servicemeshpb.ClusterMeshRuleSpec, error) {
	obj := &servicemeshpb.ClusterMeshRuleSpec{}
	if len(p) == 0 || p[0] == nil {
		return obj, fmt.Errorf("%s", "expandClusterMeshRuleSpec empty input")
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["sharing"].([]interface{}); ok && len(v) > 0 {
		obj.Sharing = expandSharingSpec(v)
	}

	if v, ok := in["version"].(string); ok && len(v) > 0 {
		obj.Version = v
	}

	if v, ok := in["artifact"].([]interface{}); ok {
		objArtifact, err := ExpandArtifactSpec(v)
		if err != nil {
			return nil, err
		}
		obj.Artifact = objArtifact
	}

	return obj, nil
}

// Flatteners

func flattenClusterMeshRule(d *schema.ResourceData, in *servicemeshpb.ClusterMeshRule) error {
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
	ret, err = flattenClusterMeshRuleSpec(in.Spec, v)
	if err != nil {
		return err
	}

	err = d.Set("spec", ret)
	if err != nil {
		return err
	}
	return nil
}

func flattenClusterMeshRuleSpec(in *servicemeshpb.ClusterMeshRuleSpec, p []interface{}) ([]interface{}, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "flattenClusterMeshRuleSpec empty input")
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	obj["sharing"] = flattenSharingSpec(in.Sharing)

	if len(in.Version) > 0 {
		obj["version"] = in.Version
	}

	v, ok := obj["artifact"].([]interface{})
	if !ok {
		v = []interface{}{}
	}

	var ret []interface{}
	var err error
	ret, err = FlattenArtifactSpec(in.Artifact, v)
	if err != nil {
		log.Println("FlattenArtifactSpec error ", err)
		return nil, err
	}

	obj["artifact"] = ret

	return []interface{}{obj}, nil
}
