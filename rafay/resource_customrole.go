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

func resourceCustomRole() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCustomRoleCreate,
		ReadContext:   resourceCustomRoleRead,
		UpdateContext: resourceCustomRoleUpdate,
		DeleteContext: resourceCustomRoleDelete,
		Importer: &schema.ResourceImporter{
			State: resourceCustomRoleImport,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema:        resource.CustomRoleSchema.Schema,
	}
}

func resourceCustomRoleImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if d.Id() == "" {
		return nil, fmt.Errorf("customRole name not provided, usage e.g terraform import rafay_customrole.resource <customRole-name>")
	}

	role_name := d.Id()

	log.Println("Importing custom_role: ", role_name)

	customRole, err := expandCustomRole(d)
	if err != nil {
		log.Printf("CustomRole expandCustomRole error")
		return nil, err
	}

	var metaD commonpb.Metadata
	metaD.Name = role_name
	customRole.Metadata = &metaD

	err = d.Set("metadata", flattenMetaData(customRole.Metadata))
	if err != nil {
		return nil, err
	}

	d.SetId(role_name)

	return []*schema.ResourceData{d}, nil
}

func resourceCustomRoleCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Println("custom role create")
	diags := resourceCustomRoleUpsert(ctx, d, m)
	if diags.HasError() {
		tflog := os.Getenv("TF_LOG")
		if tflog == "TRACE" || tflog == "DEBUG" {
			ctx = context.WithValue(ctx, "debug", "true")
		}
		zr, err := expandCustomRole(d)
		if err != nil {
			return diags
		}
		auth := config.GetConfig().GetAppAuthProfile()
		client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent(), options.WithInsecureSkipVerify(auth.SkipServerCertValid))
		if err != nil {
			return diags
		}

		err = client.SystemV3().CustomRole().Delete(ctx, options.DeleteOptions{
			Name: zr.Metadata.Name,
		})
		if err != nil {
			return diags
		}
	}
	return diags
}

func resourceCustomRoleUpsert(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("custom role upsert starts")
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

	ac, err := expandCustomRole(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent(), options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.SystemV3().CustomRole().Apply(ctx, ac, options.ApplyOptions{})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(ac.Metadata.Name)
	return diags
}

func resourceCustomRoleRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	log.Println("resource custom role ")

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

	ac, err := client.SystemV3().CustomRole().Get(ctx, options.GetOptions{
		Name: meta.Name,
	})
	if err != nil {
		log.Println("read get err")
		return diag.FromErr(err)
	}

	err = flattenCustomRole(d, ac)
	if err != nil {
		log.Println("read flatten err")
		return diag.FromErr(err)
	}
	return diags
}

func resourceCustomRoleUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceCustomRoleUpsert(ctx, d, m)
}

func resourceCustomRoleDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	addon, err := expandCustomRole(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent(), options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.SystemV3().CustomRole().Delete(ctx, options.DeleteOptions{
		Name: addon.Metadata.Name,
	})

	if err != nil {
		log.Println("custom role delete error")
		return diag.FromErr(err)
	}

	return diags
}

func expandCustomRole(in *schema.ResourceData) (*systempb.CustomRole, error) {
	log.Println("expand custom role")
	if in == nil {
		return nil, fmt.Errorf("%s", "expand CustomRole empty input")
	}
	obj := &systempb.CustomRole{}

	if v, ok := in.Get("metadata").([]interface{}); ok {
		obj.Metadata = expandMetaData(v)
	}

	if v, ok := in.Get("spec").([]interface{}); ok && len(v) > 0 {
		objSpec, err := expandCustomRoleSpec(v)
		if err != nil {
			return nil, err
		}
		obj.Spec = objSpec
	}

	obj.ApiVersion = "system.k8smgmt.io/v3"
	obj.Kind = "CustomRole"

	return obj, nil
}

func expandPolicyList(p []interface{}) []*systempb.PolicyRef {
	if len(p) == 0 || p[0] == nil {
		return []*systempb.PolicyRef{}
	}

	out := make([]*systempb.PolicyRef, len(p))

	for i := range p {
		obj := systempb.PolicyRef{}
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

func expandCustomRoleSpec(p []interface{}) (*systempb.CustomRoleSpec, error) {

	obj := &systempb.CustomRoleSpec{}
	if len(p) == 0 || p[0] == nil {
		return obj, fmt.Errorf("%s", "expandCustomRoleSpec empty input")
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["base_role"].(string); ok && len(v) > 0 {
		obj.BaseRole = v
	}

	if v, ok := in["abac_policy_list"].([]interface{}); ok && len(v) > 0 {
		obj.AbacPolicyList = expandPolicyList(v)
	}

	if v, ok := in["ztka_policy_list"].([]interface{}); ok && len(v) > 0 {
		obj.ZtkaPolicyList = expandPolicyList(v)
	}

	return obj, nil
}

// Flatteners

func flattenCustomRole(d *schema.ResourceData, in *systempb.CustomRole) error {
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
	ret, err = flattenCustomRoleSpec(in.Spec, v)
	if err != nil {
		log.Println("flatten custom role spec err")
		return err
	}

	err = d.Set("spec", ret)
	if err != nil {
		log.Println("set spec err")
		return err
	}
	return nil
}

func flattenCustomRoleSpec(in *systempb.CustomRoleSpec, p []interface{}) ([]interface{}, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "flattenCustomRoleSpec empty input")
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.BaseRole) > 0 {
		obj["base_role"] = in.BaseRole
	}

	if in.ZtkaPolicyList != nil && len(in.ZtkaPolicyList) > 0 {
		v, ok := obj["ztka_policy_list"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["ztka_policy_list"] = flattenPolicyList(in.ZtkaPolicyList, v)
	}

	if in.AbacPolicyList != nil && len(in.AbacPolicyList) > 0 {
		v, ok := obj["abac_policy_list"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["abac_policy_list"] = flattenPolicyList(in.AbacPolicyList, v)
	}

	return []interface{}{obj}, nil
}

func flattenPolicyList(input []*systempb.PolicyRef, p []interface{}) []interface{} {
	log.Println("flattenPolicyList")
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
