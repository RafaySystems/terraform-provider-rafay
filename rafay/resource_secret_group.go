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
	"github.com/RafaySystems/rafay-common/proto/types/hub/gitopspb"
	"github.com/RafaySystems/rctl/pkg/config"
	"github.com/RafaySystems/rctl/pkg/versioninfo"
	"github.com/davecgh/go-spew/spew"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceSecretGroup() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSecretGroupCreate,
		ReadContext:   resourceSecretGroupRead,
		UpdateContext: resourceSecretGroupUpdate,
		DeleteContext: resourceSecretGroupDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema:        resource.SecretGroupSchema.Schema,
	}
}

func resourceSecretGroupCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("SecretGroup create starts")
	diags := resourceSecretGroupUpsert(ctx, d, m)
	if diags.HasError() {
		tflog := os.Getenv("TF_LOG")
		if tflog == "TRACE" || tflog == "DEBUG" {
			ctx = context.WithValue(ctx, "debug", "true")
		}
		log.Printf("SecretGroup create got error, perform cleanup")
		ag, err := expandSecretGroup(d)
		if err != nil {
			log.Printf("SecretGroup expandSecretGroup error")
			return diags
		}
		auth := config.GetConfig().GetAppAuthProfile()
		client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
		if err != nil {
			return diags
		}

		err = client.GitopsV3().SecretGroup().Delete(ctx, options.DeleteOptions{
			Name:    ag.Metadata.Name,
			Project: ag.Metadata.Project,
		})
		if err != nil {
			return diags
		}
	}
	return diags
}

func resourceSecretGroupUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("SecretGroup update starts")
	return resourceSecretGroupUpsert(ctx, d, m)
}

func resourceSecretGroupUpsert(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("SecretGroup upsert starts")
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

	secretGroup, err := expandSecretGroup(d)
	if err != nil {
		log.Printf("pipeline expandPipeline error")
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.GitopsV3().SecretGroup().Apply(ctx, secretGroup, options.ApplyOptions{})
	if err != nil {
		// XXX Debug
		n1 := spew.Sprintf("%+v", secretGroup)
		log.Println("secretGroup apply pipeline:", n1)
		log.Printf("secretGroup apply error")
		return diag.FromErr(err)
	}

	d.SetId(secretGroup.Metadata.Name)
	return diags
}

func resourceSecretGroupRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	log.Println("resourceSecretGroupRead ")
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

	secretGroup, err := expandSecretGroup(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}

	ag, err := client.GitopsV3().SecretGroup().Get(ctx, options.GetOptions{
		//Name:    secretGroup.Metadata.Name,
		Name:    meta.Name,
		Project: secretGroup.Metadata.Project,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	err = flattenSecretGroup(d, ag)
	if err != nil {
		return diag.FromErr(err)
	}
	return diags

}

func resourceSecretGroupDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	secretGroup, err := expandSecretGroup(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.GitopsV3().SecretGroup().Delete(ctx, options.DeleteOptions{
		Name:    secretGroup.Metadata.Name,
		Project: secretGroup.Metadata.Project,
	})

	return diags
}

func expandSecretGroup(in *schema.ResourceData) (*gitopspb.SecretGroup, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "expand SecretGroup empty input")
	}
	obj := &gitopspb.SecretGroup{}

	if v, ok := in.Get("metadata").([]interface{}); ok && len(v) > 0 {
		obj.Metadata = expandMetaData(v)
	}

	if v, ok := in.Get("spec").([]interface{}); ok && len(v) > 0 {
		objSpec, err := expandSecretGroupSpec(v)
		if err != nil {
			return nil, err
		}
		log.Println("expandSecretGroupSpec got spec")
		obj.Spec = objSpec
	}

	obj.ApiVersion = "gitops.k8smgmt.io/v3"
	obj.Kind = "SecretGroup"
	return obj, nil
}

func expandSecretGroupSpec(p []interface{}) (*gitopspb.SecretGroupSpec, error) {
	obj := &gitopspb.SecretGroupSpec{}
	if len(p) == 0 || p[0] == nil {
		return obj, fmt.Errorf("%s", "expandSecretGroupSpec empty input")
	}

	in := p[0].(map[string]interface{})

	//how to expand secret?
	if v, ok := in["secret"].([]interface{}); ok {
		obj.Secret = expandCommonpbFile(v)
	}

	if v, ok := in["secrets"].([]interface{}); ok && len(v) > 0 {
		obj.Secrets = expandSecrets(v)
	}

	return obj, nil
}

func expandSecrets(p []interface{}) []*gitopspb.Secret {
	if len(p) == 0 || p[0] == nil {
		return []*gitopspb.Secret{}
	}

	out := make([]*gitopspb.Secret, len(p))

	for i := range p {
		obj := gitopspb.Secret{}
		in := p[i].(map[string]interface{})

		if v, ok := in["file_path"].(string); ok && len(v) > 0 {
			obj.FilePath = v
		}

		if v, ok := in["secret"].(string); ok && len(v) > 0 {
			obj.Secret = v
		}

		out[i] = &obj

	}
	return out
}

func flattenSecretGroup(d *schema.ResourceData, in *gitopspb.SecretGroup) error {
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
	ret, err = flattenSecretGroupSpec(in.Spec, v)
	if err != nil {
		return err
	}

	err = d.Set("spec", ret)
	if err != nil {
		return err
	}
	return nil
}

func flattenSecretGroupSpec(in *gitopspb.SecretGroupSpec, p []interface{}) ([]interface{}, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "flattenPipeline empty input")
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if in.Secret != nil {
		obj["secrets"] = flattenCommonpbFile(in.Secret)
	}

	if in.Secrets != nil && len(in.Secrets) > 0 {
		v, ok := obj["secrets"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["secrets"] = flattenSecrets(in.Secrets, v)
	}

	return []interface{}{obj}, nil
}

func flattenSecrets(input []*gitopspb.Secret, p []interface{}) []interface{} {
	log.Println("flattenVariableSpec")
	if input == nil {
		return nil
	}

	out := make([]interface{}, len(input))
	for i, in := range input {
		log.Println("flattenSecrets in ", in)
		obj := map[string]interface{}{}
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}

		if len(in.Secret) > 0 {
			obj["secret"] = in.Secret
		}

		if len(in.FilePath) > 0 {
			obj["file_path"] = in.FilePath
		}

		out[i] = &obj
	}

	return out
}
