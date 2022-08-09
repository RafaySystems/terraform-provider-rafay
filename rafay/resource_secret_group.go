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
	log.Printf("secret group create starts")
	diags := resourceSecretGroupUpsert(ctx, d, m)
	if diags.HasError() {
		tflog := os.Getenv("TF_LOG")
		if tflog == "TRACE" || tflog == "DEBUG" {
			ctx = context.WithValue(ctx, "debug", "true")
		}
		log.Printf("secret group create got error, perform cleanup")
		ss, err := expandSecretGroup(d)
		if err != nil {
			log.Printf("secret group expandSecretGroup error")
			return diags
		}
		auth := config.GetConfig().GetAppAuthProfile()
		client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
		if err != nil {
			return diags
		}

		err = client.GitopsV3().SecretGroup().Delete(ctx, options.DeleteOptions{
			Name:    ss.Metadata.Name,
			Project: ss.Metadata.Project,
		})
		if err != nil {
			return diags
		}
	}
	return diags
}

func resourceSecretGroupUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("secret group update starts")
	return resourceSecretGroupUpsert(ctx, d, m)
}

func resourceSecretGroupUpsert(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("secret group upsert starts")
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	secretGroup, err := expandSecretGroup(d)
	if err != nil {
		log.Printf("secret group expandSecretGroup error")
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
		log.Println("secret group apply secret group:", n1)
		log.Printf("secret group apply error")
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
	tfSecretGroupState, err := expandSecretGroup(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}

	ag, err := client.GitopsV3().SecretGroup().Get(ctx, options.GetOptions{
		Name:    tfSecretGroupState.Metadata.Name,
		Project: tfSecretGroupState.Metadata.Project,
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

	ag, err := expandSecretGroup(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.GitopsV3().SecretGroup().Delete(ctx, options.DeleteOptions{
		Name:    ag.Metadata.Name,
		Project: ag.Metadata.Project,
	})

	return diags
}

func expandSecretGroup(in *schema.ResourceData) (*gitopspb.SecretGroup, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "expand secret group empty input")
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

	obj.ApiVersion = "integrations.k8smgmt.io/v3"
	obj.Kind = "SecretGroup"
	return obj, nil
}

func expandSecretGroupSpec(p []interface{}) (*gitopspb.SecretGroupSpec, error) {
	obj := &gitopspb.SecretGroupSpec{}
	if len(p) == 0 || p[0] == nil {
		return obj, fmt.Errorf("%s", "expandSecretGroupSpec empty input")
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["secret"].([]interface{}); ok && len(v) > 0 {
		obj.Secret = expandCommonpbFile(v)
	}

	if v, ok := in["secrets"].([]interface{}); ok && len(v) > 0 {
		obj.Secrets = expandSecretGroupSecrets(v)
	}

	return obj, nil
}

func expandSecretGroupSecrets(p []interface{}) []*gitopspb.Secret {
	if len(p) == 0 || p[0] == nil {
		return []*gitopspb.Secret{}
	}

	out := make([]*gitopspb.Secret, len(p))

	for i := range p {
		obj := gitopspb.Secret{}
		in := p[i].(map[string]interface{})

		if v, ok := in["secret"].(string); ok {
			obj.Secret = v
		}

		if v, ok := in["filepath"].(string); ok {
			obj.FilePath = v
		}

		out[i] = &obj
	}

	return out

}

// Flatten

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
		return nil, fmt.Errorf("%s", "flattenSecretGroup empty input")
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if in.Secret != nil {
		obj["secret"] = flattenCommonpbFile(in.Secret)
	}

	if in.Secrets != nil {
		v, ok := obj["secrets"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["secrets"] = flattenSecretGroupSecrets(in.Secrets, v)
	}

	return []interface{}{obj}, nil
}

func flattenSecretGroupSecrets(in []*gitopspb.Secret, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}

	out := make([]interface{}, len(in))
	for i, in := range in {

		obj := map[string]interface{}{}
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}

		if len(in.Secret) > 0 {
			obj["secret"] = in.Secret
		}

		if len(in.FilePath) > 0 {
			obj["filepath"] = in.FilePath
		}

		out[i] = &obj
	}

	return out
}
