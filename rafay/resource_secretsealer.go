package rafay

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/RafaySystems/rafay-common/pkg/hub/client/options"
	typed "github.com/RafaySystems/rafay-common/pkg/hub/client/typed"
	"github.com/RafaySystems/rafay-common/pkg/hub/terraform/resource"
	"github.com/RafaySystems/rafay-common/proto/types/hub/commonpb"
	"github.com/RafaySystems/rafay-common/proto/types/hub/integrationspb"
	"github.com/RafaySystems/rctl/pkg/config"
	"github.com/RafaySystems/rctl/pkg/versioninfo"
	"github.com/davecgh/go-spew/spew"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceSecretSealer() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSecretSealerCreate,
		ReadContext:   resourceSecretSealerRead,
		UpdateContext: resourceSecretSealerUpdate,
		DeleteContext: resourceSecretSealerDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema:        resource.SecretSealerSchema.Schema,
	}
}

func resourceSecretSealerCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("secret sealer create starts")
	diags := resourceSecretSealerUpsert(ctx, d, m)
	if diags.HasError() {
		tflog := os.Getenv("TF_LOG")
		if tflog == "TRACE" || tflog == "DEBUG" {
			ctx = context.WithValue(ctx, "debug", "true")
		}
		log.Printf("secret sealer create got error, perform cleanup")
		ss, err := expandSecretSealer(d)
		if err != nil {
			log.Printf("secret sealer expandSecretSealer error")
			return diags
		}
		auth := config.GetConfig().GetAppAuthProfile()
		client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent(), options.WithInsecureSkipVerify(auth.SkipServerCertValid))
		if err != nil {
			return diags
		}

		err = client.IntegrationsV3().SecretSealer().Delete(ctx, options.DeleteOptions{
			Name:    ss.Metadata.Name,
			Project: ss.Metadata.Project,
		})
		if err != nil {
			return diags
		}
	}
	return diags
}

func resourceSecretSealerUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("secret sealer update starts")
	return resourceSecretSealerUpsert(ctx, d, m)
}

func resourceSecretSealerUpsert(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("secret sealer upsert starts")
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

	secretSealer, err := expandSecretSealer(d)
	if err != nil {
		log.Printf("secret sealer expandSecretSealer error")
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent(), options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.IntegrationsV3().SecretSealer().Apply(ctx, secretSealer, options.ApplyOptions{})
	if err != nil {
		// XXX Debug
		n1 := spew.Sprintf("%+v", secretSealer)
		log.Println("secret sealer apply secret sealer:", n1)
		log.Printf("secret sealer apply error")
		return diag.FromErr(err)
	}

	d.SetId(secretSealer.Metadata.Name)
	return diags

}

func resourceSecretSealerRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	log.Println("resourceSecretSealerRead ")
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

	tfSecretSealerState, err := expandSecretSealer(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent(), options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return diag.FromErr(err)
	}

	ag, err := client.IntegrationsV3().SecretSealer().Get(ctx, options.GetOptions{
		//Name:    tfSecretSealerState.Metadata.Name,
		Name:    meta.Name,
		Project: tfSecretSealerState.Metadata.Project,
	})
	if err != nil {
		if strings.Contains(err.Error(), "code 404") {
			log.Println("Resource Read ", "error", err)
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	if tfSecretSealerState.Spec != nil && tfSecretSealerState.Spec.Sharing != nil && !tfSecretSealerState.Spec.Sharing.Enabled && ag.Spec.Sharing == nil {
		ag.Spec.Sharing = &commonpb.SharingSpec{}
		ag.Spec.Sharing.Enabled = false
	}

	err = flattenSecretSealer(d, ag)
	if err != nil {
		return diag.FromErr(err)
	}
	return diags

}

func resourceSecretSealerDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	ag, err := expandSecretSealer(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent(), options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.IntegrationsV3().SecretSealer().Delete(ctx, options.DeleteOptions{
		Name:    ag.Metadata.Name,
		Project: ag.Metadata.Project,
	})

	return diags
}

func expandSecretSealer(in *schema.ResourceData) (*integrationspb.SecretSealer, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "expand secret sealer empty input")
	}
	obj := &integrationspb.SecretSealer{}

	if v, ok := in.Get("metadata").([]interface{}); ok && len(v) > 0 {
		obj.Metadata = expandMetaData(v)
	}

	if v, ok := in.Get("spec").([]interface{}); ok && len(v) > 0 {
		objSpec, err := expandSecretSealerSpec(v)
		if err != nil {
			return nil, err
		}
		log.Println("expandSecretSealerSpec got spec")
		obj.Spec = objSpec
	}

	obj.ApiVersion = "integrations.k8smgmt.io/v3"
	obj.Kind = "SecretSealer"
	return obj, nil
}

func expandSecretSealerSpec(p []interface{}) (*integrationspb.SecretSealerSpec, error) {
	obj := &integrationspb.SecretSealerSpec{}
	if len(p) == 0 || p[0] == nil {
		return obj, fmt.Errorf("%s", "expandSecretSealerSpec empty input")
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["type"].(string); ok && len(v) > 0 {
		obj.Type = v
	}

	if v, ok := in["sharing"].([]interface{}); ok && len(v) > 0 {
		obj.Sharing = expandSharingSpec(v)
	}

	if v, ok := in["version"].(string); ok && len(v) > 0 {
		obj.Version = v
	}

	return obj, nil
}

// Flatten

func flattenSecretSealer(d *schema.ResourceData, in *integrationspb.SecretSealer) error {
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
	ret, err = flattenSecretSealerSpec(in.Spec, v)
	if err != nil {
		return err
	}

	err = d.Set("spec", ret)
	if err != nil {
		return err
	}
	return nil
}

func flattenSecretSealerSpec(in *integrationspb.SecretSealerSpec, p []interface{}) ([]interface{}, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "flattenSecretSealer empty input")
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.Type) > 0 {
		obj["type"] = in.Type
	}

	obj["sharing"] = flattenSharingSpec(in.Sharing)

	if len(in.Version) > 0 {
		obj["version"] = in.Version
	}

	return []interface{}{obj}, nil
}
