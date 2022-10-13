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

func resourceNamespaceMeshRule() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceNamespaceMeshRuleCreate,
		ReadContext:   resourceNamespaceMeshRuleRead,
		UpdateContext: resourceNamespaceMeshRuleUpdate,
		DeleteContext: resourceNamespaceMeshRuleDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema:        resource.NamespaceMeshRuleSchema.Schema,
	}
}

func resourceNamespaceMeshRuleCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("NamespaceMeshRule create starts")
	diags := resourceNamespaceMeshRuleUpsert(ctx, d, m)
	if diags.HasError() {
		tflog := os.Getenv("TF_LOG")
		if tflog == "TRACE" || tflog == "DEBUG" {
			ctx = context.WithValue(ctx, "debug", "true")
		}
		log.Printf("namespaceMeshRule create got error, perform cleanup")
		cnpr, err := expandNamespaceMeshRule(d)
		if err != nil {
			log.Printf("namespaceMeshRule expandNamespaceMeshRule error")
			return diags
		}
		auth := config.GetConfig().GetAppAuthProfile()
		client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
		if err != nil {
			return diags
		}

		err = client.ServicemeshV3().NamespaceMeshRule().Delete(ctx, options.DeleteOptions{
			Name:    cnpr.Metadata.Name,
			Project: cnpr.Metadata.Project,
		})
		if err != nil {
			return diags
		}
	}
	return diags
}

func resourceNamespaceMeshRuleUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("namespaceMeshRule update starts")
	return resourceNamespaceMeshRuleUpsert(ctx, d, m)
}

func resourceNamespaceMeshRuleUpsert(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("namespaceMeshRule upsert starts")
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

	namespaceMeshRule, err := expandNamespaceMeshRule(d)
	if err != nil {
		log.Printf("namespaceMeshRule expandNamespaceMeshRule error")
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.ServicemeshV3().NamespaceMeshRule().Apply(ctx, namespaceMeshRule, options.ApplyOptions{})
	if err != nil {
		// XXX Debug
		n1 := spew.Sprintf("%+v", namespaceMeshRule)
		log.Println("namespaceMeshRule apply namespaceMeshRule:", n1)
		log.Printf("namespaceMeshRule apply error")
		return diag.FromErr(err)
	}

	d.SetId(namespaceMeshRule.Metadata.Name)
	return diags

}

func resourceNamespaceMeshRuleRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	log.Println("resourceNamespaceMeshRuleRead ")
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

	tfNamespaceMeshRuleState, err := expandNamespaceMeshRule(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}

	cnp, err := client.ServicemeshV3().NamespaceMeshRule().Get(ctx, options.GetOptions{
		//Name:    tfNamespaceMeshRuleState.Metadata.Name,
		Name:    meta.Name,
		Project: tfNamespaceMeshRuleState.Metadata.Project,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	err = flattenNamespaceMeshRule(d, cnp)
	if err != nil {
		return diag.FromErr(err)
	}
	return diags

}

func resourceNamespaceMeshRuleDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	cnpr, err := expandNamespaceMeshRule(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.ServicemeshV3().NamespaceMeshRule().Delete(ctx, options.DeleteOptions{
		Name:    cnpr.Metadata.Name,
		Project: cnpr.Metadata.Project,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func expandNamespaceMeshRule(in *schema.ResourceData) (*servicemeshpb.NamespaceMeshRule, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "expand namespaceMeshRule empty input")
	}
	obj := &servicemeshpb.NamespaceMeshRule{}

	if v, ok := in.Get("metadata").([]interface{}); ok && len(v) > 0 {
		obj.Metadata = expandMetaData(v)
	}

	if v, ok := in.Get("spec").([]interface{}); ok && len(v) > 0 {
		objSpec, err := expandNamespaceMeshRuleSpec(v)
		if err != nil {
			return nil, err
		}
		log.Println("expandNamespaceMeshRuleSpec got spec")
		obj.Spec = objSpec
	}

	obj.ApiVersion = "servicemesh.k8smgmt.io/v3"
	obj.Kind = "NamespaceMeshRule"
	return obj, nil
}

func expandNamespaceMeshRuleSpec(p []interface{}) (*servicemeshpb.NamespaceMeshRuleSpec, error) {
	obj := &servicemeshpb.NamespaceMeshRuleSpec{}
	if len(p) == 0 || p[0] == nil {
		return obj, fmt.Errorf("%s", "expandNamespaceMeshRuleSpec empty input")
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

func flattenNamespaceMeshRule(d *schema.ResourceData, in *servicemeshpb.NamespaceMeshRule) error {
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
	ret, err = flattenNamespaceMeshRuleSpec(in.Spec, v)
	if err != nil {
		return err
	}

	err = d.Set("spec", ret)
	if err != nil {
		return err
	}
	return nil
}

func flattenNamespaceMeshRuleSpec(in *servicemeshpb.NamespaceMeshRuleSpec, p []interface{}) ([]interface{}, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "flattenNamespaceMeshRuleSpec empty input")
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if in.Sharing != nil {
		obj["sharing"] = flattenSharingSpec(in.Sharing)
	}

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
