package rafay

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/RafaySystems/rctl/pkg/config"

	"github.com/RafaySystems/rafay-common/pkg/hub/client/options"
	typed "github.com/RafaySystems/rafay-common/pkg/hub/client/typed"
	"github.com/RafaySystems/rafay-common/pkg/hub/terraform/resource"
	"github.com/RafaySystems/rafay-common/proto/types/hub/commonpb"
	"github.com/RafaySystems/rafay-common/proto/types/hub/systempb"
	"github.com/RafaySystems/rctl/pkg/versioninfo"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceZTKAPolicy() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceZTKAPolicyCreate,
		ReadContext:   resourceZTKAPolicyRead,
		UpdateContext: resourceZTKAPolicyUpdate,
		DeleteContext: resourceZTKAPolicyDelete,
		Importer: &schema.ResourceImporter{
			State: resourceZTKAPolicyImport,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema:        resource.ZTKAPolicySchema.Schema,
	}
}

func resourceZTKAPolicyImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if d.Id() == "" {
		return nil, fmt.Errorf("ztkapolicy name not provided, usage e.g terraform import rafay_ztkapolicy.resource <ztkapolicy-name>")
	}

	policy_name := d.Id()

	log.Println("Importing ztka-policy: ", policy_name)

	ztkaPolicy, err := expandZTKAPolicy(d)
	if err != nil {
		log.Printf("ztkaPolicy expandZTKAPolicy error")
		return nil, err
	}

	var metaD commonpb.Metadata
	metaD.Name = policy_name
	ztkaPolicy.Metadata = &metaD

	err = d.Set("metadata", flattenMetaData(ztkaPolicy.Metadata))
	if err != nil {
		return nil, err
	}

	d.SetId(policy_name)

	return []*schema.ResourceData{d}, nil
}

func resourceZTKAPolicyCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Println("ztka policy create")
	create := isZTKAPolicyAlreadyExists(ctx, d)
	diags := resourceZTKAPolicyUpsert(ctx, d, m)
	if diags.HasError() && !create {
		tflog := os.Getenv("TF_LOG")
		if tflog == "TRACE" || tflog == "DEBUG" {
			ctx = context.WithValue(ctx, "debug", "true")
		}
		zr, err := expandZTKAPolicy(d)
		if err != nil {
			return diags
		}
		auth := config.GetConfig().GetAppAuthProfile()
		client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent(), options.WithInsecureSkipVerify(auth.SkipServerCertValid))
		if err != nil {
			return diags
		}

		err = client.SystemV3().ZTKAPolicy().Delete(ctx, options.DeleteOptions{
			Name: zr.Metadata.Name,
		})
		if err != nil {
			return diags
		}
	}
	return diags
}

func resourceZTKAPolicyUpsert(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("ztka policy upsert starts")
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

	ac, err := expandZTKAPolicy(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent(), options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.SystemV3().ZTKAPolicy().Apply(ctx, ac, options.ApplyOptions{})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(ac.Metadata.Name)
	return diags
}

func resourceZTKAPolicyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	log.Println("resource ztka policy ")

	meta := GetMetaData(d)
	if meta == nil {
		return diag.FromErr(fmt.Errorf("%s", "failed to read resource "))
	}
	if d.State() != nil && d.State().ID != "" {
		meta.Name = d.State().ID
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent(), options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		log.Println("read client err")
		return diag.FromErr(err)
	}

	ac, err := client.SystemV3().ZTKAPolicy().Get(ctx, options.GetOptions{
		Name: meta.Name,
	})
	if err != nil {
		log.Println("read get err")
		return diag.FromErr(err)
	}

	err = flattenZTKAPolicy(d, ac)
	if err != nil {
		log.Println("read flatten err")
		return diag.FromErr(err)
	}
	return diags
}

func resourceZTKAPolicyUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceZTKAPolicyUpsert(ctx, d, m)
}

func resourceZTKAPolicyDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	addon, err := expandZTKAPolicy(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent(), options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.SystemV3().ZTKAPolicy().Delete(ctx, options.DeleteOptions{
		Name: addon.Metadata.Name,
	})

	if err != nil {
		log.Println("ztka policy delete error")
		return diag.FromErr(err)
	}

	return diags
}

func expandZTKAPolicy(in *schema.ResourceData) (*systempb.ZTKAPolicy, error) {
	log.Println("expand ztka policy")
	if in == nil {
		return nil, fmt.Errorf("%s", "expand ZTKAPolicy empty input")
	}
	obj := &systempb.ZTKAPolicy{}

	if v, ok := in.Get("metadata").([]interface{}); ok {
		obj.Metadata = expandMetaData(v)
	}

	if v, ok := in.Get("spec").([]interface{}); ok && len(v) > 0 {
		objSpec, err := expandZTKAPolicySpec(v)
		if err != nil {
			return nil, err
		}
		obj.Spec = objSpec
	}

	obj.ApiVersion = "system.k8smgmt.io/v3"
	obj.Kind = "ZTKAPolicy"

	return obj, nil
}

func expandZTKARuleList(p []interface{}) []*systempb.ZTKAPolicyRule {
	if len(p) == 0 || p[0] == nil {
		return []*systempb.ZTKAPolicyRule{}
	}

	out := make([]*systempb.ZTKAPolicyRule, len(p))

	for i := range p {
		obj := systempb.ZTKAPolicyRule{}
		in := p[i].(map[string]interface{})

		if v, ok := in["name"].(string); ok && len(v) > 0 {
			obj.Name = v
		}
		if v, ok := in["version"].(string); ok && len(v) > 0 {
			obj.Version = v
		}

		out[i] = &obj

	}

	return out
}

func expandZTKAPolicySpec(p []interface{}) (*systempb.ZTKAPolicySpec, error) {

	obj := &systempb.ZTKAPolicySpec{}
	if len(p) == 0 || p[0] == nil {
		return obj, fmt.Errorf("%s", "expandZTKAPolicySpec empty input")
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["ztka_rule_list"].([]interface{}); ok && len(v) > 0 {
		obj.ZtkaRuleList = expandZTKARuleList(v)
	}

	if v, ok := in["version"].(string); ok && len(v) > 0 {
		obj.Version = v
	}

	return obj, nil
}

// Flatteners

func flattenZTKAPolicy(d *schema.ResourceData, in *systempb.ZTKAPolicy) error {
	if in == nil {
		return nil
	}

	err := d.Set("metadata", flattenMetaData(in.Metadata))
	if err != nil {
		log.Println("flatten metadata err")
		return err
	}

	v, ok := d.Get("spec").([]interface{})
	if !ok {
		v = []interface{}{}
	}

	var ret []interface{}
	ret, err = flattenZTKAPolicySpec(in.Spec, v)
	if err != nil {
		log.Println("flatten ztka policy spec err")
		return err
	}

	err = d.Set("spec", ret)
	if err != nil {
		log.Println("set spec err")
		return err
	}
	return nil
}

func flattenZTKAPolicySpec(in *systempb.ZTKAPolicySpec, p []interface{}) ([]interface{}, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "flattenZTKAPolicySpec empty input")
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.Version) > 0 {
		obj["version"] = in.Version
	}

	if in.ZtkaRuleList != nil {
		v, ok := obj["ztka_rule_list"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["ztka_rule_list"] = flattenZTKARuleList(in.ZtkaRuleList, v)
	}

	return []interface{}{obj}, nil
}

func flattenZTKARuleList(input []*systempb.ZTKAPolicyRule, p []interface{}) []interface{} {
	log.Println("flattenZTKARuleList")
	if input == nil {
		return nil
	}
	out := make([]interface{}, len(input))
	for i, in := range input {

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

func isZTKAPolicyAlreadyExists(ctx context.Context, d *schema.ResourceData) bool {
	meta := GetMetaData(d)
	if meta == nil {
		return false
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent(), options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return false
	}

	_, err = client.SystemV3().ZTKAPolicy().Get(ctx, options.GetOptions{
		Name: meta.Name,
	})
	if err != nil {
		return false
	}

	return true
}