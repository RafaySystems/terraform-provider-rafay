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
	"github.com/RafaySystems/rafay-common/proto/types/hub/integrationspb"
	"github.com/RafaySystems/rctl/pkg/config"
	"github.com/RafaySystems/rctl/pkg/versioninfo"
	"github.com/davecgh/go-spew/spew"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceSecretProvider() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSecretProviderCreate,
		ReadContext:   resourceSecretProviderRead,
		UpdateContext: resourceSecretProviderUpdate,
		DeleteContext: resourceSecretProviderDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema:        resource.SecretProviderClassSchema.Schema,
	}
}

func resourceSecretProviderCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("secret sealer create starts")
	diags := resourceSecretProviderUpsert(ctx, d, m)
	if diags.HasError() {
		tflog := os.Getenv("TF_LOG")
		if tflog == "TRACE" || tflog == "DEBUG" {
			ctx = context.WithValue(ctx, "debug", "true")
		}
		log.Printf("secret sealer create got error, perform cleanup")
		ss, err := expandSecretProvider(d)
		if err != nil {
			log.Printf("secret sealer expandSecretSealer error")
			return diags
		}
		auth := config.GetConfig().GetAppAuthProfile()
		client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent(), options.WithInsecureSkipVerify(auth.SkipServerCertValid))
		if err != nil {
			return diags
		}

		err = client.IntegrationsV3().SecretProviderClass().Delete(ctx, options.DeleteOptions{
			Name:    ss.Metadata.Name,
			Project: ss.Metadata.Project,
		})
		if err != nil {
			return diags
		}
	}
	return diags
}

func resourceSecretProviderUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("secret sealer update starts")
	return resourceSecretProviderUpsert(ctx, d, m)
}

func resourceSecretProviderUpsert(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("secret provider upsert starts")
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

	secretSealer, err := expandSecretProvider(d)
	if err != nil {
		log.Printf("secret sealer expandSecretSealer error")
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent(), options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.IntegrationsV3().SecretProviderClass().Apply(ctx, secretSealer, options.ApplyOptions{})
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

func resourceSecretProviderRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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

	tfSecretSealerState, err := expandSecretProvider(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent(), options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return diag.FromErr(err)
	}

	ag, err := client.IntegrationsV3().SecretProviderClass().Get(ctx, options.GetOptions{
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

	err = flattenSecretProvider(d, ag)
	if err != nil {
		return diag.FromErr(err)
	}
	return diags

}

func resourceSecretProviderDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	ag, err := expandSecretProvider(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent(), options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.IntegrationsV3().SecretProviderClass().Delete(ctx, options.DeleteOptions{
		Name:    ag.Metadata.Name,
		Project: ag.Metadata.Project,
	})

	return diags
}

func expandSecretProvider(in *schema.ResourceData) (*integrationspb.SecretProviderClass, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "expand secret sealer empty input")
	}
	obj := &integrationspb.SecretProviderClass{}

	if v, ok := in.Get("metadata").([]interface{}); ok && len(v) > 0 {
		obj.Metadata = expandMetaData(v)
	}

	if v, ok := in.Get("spec").([]interface{}); ok && len(v) > 0 {
		objSpec, err := expandSecretProviderSpec(v)
		if err != nil {
			return nil, err
		}
		log.Println("expandSecretSealerSpec got spec")
		obj.Spec = objSpec
	}

	obj.ApiVersion = "integrations.k8smgmt.io/v3"
	obj.Kind = "SecretProviderClass"
	return obj, nil
}

func expandSecretProviderSpec(p []interface{}) (*integrationspb.SecretProviderClassSpec, error) {
	obj := &integrationspb.SecretProviderClassSpec{}
	if len(p) == 0 || p[0] == nil {
		return obj, fmt.Errorf("%s", "expandSecretSealerSpec empty input")
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["provider"].(string); ok && len(v) > 0 {
		obj.Provider = v
	}

	if v, ok := in["artifact"].([]interface{}); ok && len(v) > 0 {
		objArtifact, err := ExpandArtifactSpec(v)
		if err != nil {
			return nil, err
		}
		obj.Artifact = objArtifact
	}
	artfct := spew.Sprintf("%+v", obj.Artifact)
	log.Println("expandSecretClassProviderSpec Artifact ater expand ", artfct)

	if v, ok := in["parameters"].(map[string]interface{}); ok && len(v) > 0 {
		obj.Parameters = toMapString(v)
	}

	return obj, nil
}

// Flatten

func flattenSecretProvider(d *schema.ResourceData, in *integrationspb.SecretProviderClass) error {
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
	ret, err = flattenSecretProviderSpec(in.Spec, v)
	if err != nil {
		return err
	}

	err = d.Set("spec", ret)
	if err != nil {
		return err
	}
	return nil
}

func flattenSecretProviderSpec(in *integrationspb.SecretProviderClassSpec, p []interface{}) ([]interface{}, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "flattenSecretSealer empty input")
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.Provider) > 0 {
		obj["provider"] = in.Provider
	}

	if in.Artifact != nil {
		v, ok := obj["artifact"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		// XXX Debug
		// w1 := spew.Sprintf("%+v", v)
		// log.Println("flattenAddonSpec before ", w1)

		var ret []interface{}
		var err error
		ret, err = FlattenArtifactSpec(in.Artifact, v)
		if err != nil {
			log.Println("FlattenArtifactSpec error ", err)
			return nil, err
		}
		obj["artifact"] = ret
	}

	if in.Parameters != nil && len(in.Parameters) > 0 {
		obj["parameters"] = toMapInterface(in.Parameters)
	}

	return []interface{}{obj}, nil
}
