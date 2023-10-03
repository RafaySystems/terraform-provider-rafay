package rafay

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/RafaySystems/rafay-common/pkg/hub/client/options"
	typed "github.com/RafaySystems/rafay-common/pkg/hub/client/typed"
	"github.com/RafaySystems/rafay-common/pkg/hub/terraform/resource"
	"github.com/RafaySystems/rafay-common/proto/types/hub/commonpb"
	"github.com/RafaySystems/rafay-common/proto/types/hub/eaaspb"
	"github.com/RafaySystems/rctl/pkg/config"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceEnvironment() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceEnvironmentCreate,
		ReadContext:   resourceEnvironmentRead,
		UpdateContext: resourceEnvironmentUpdate,
		DeleteContext: resourceEnvironmentDelete,
		Importer: &schema.ResourceImporter{
			State: resourceEnvironmentImport,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema:        resource.EnvironmentSchema.Schema,
	}
}

func resourceEnvironmentCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Println("environment create")
	diags := environmentUpsert(ctx, d, m)
	if diags.HasError() {
		tflog := os.Getenv("TF_LOG")
		if tflog == "TRACE" || tflog == "DEBUG" {
			ctx = context.WithValue(ctx, "debug", "true")
		}
		environment, err := expandEnvironment(d)
		if err != nil {
			return diags
		}
		auth := config.GetConfig().GetAppAuthProfile()
		client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, TF_USER_AGENT, options.WithInsecureSkipVerify(auth.SkipServerCertValid))
		if err != nil {
			return diags
		}

		err = client.EaasV1().Environment().Delete(ctx, options.DeleteOptions{
			Name:    environment.Metadata.Name,
			Project: environment.Metadata.Project,
		})
		if err != nil {
			return diags
		}
	}
	return diags
}

func environmentUpsert(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("environment upsert starts")
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	environment, err := expandEnvironment(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, TF_USER_AGENT, options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.EaasV1().Environment().Apply(ctx, environment, options.ApplyOptions{})
	if err != nil {
		return diag.FromErr(err)
	}

	response, err := client.EaasV1().Environment().ExtApi().Publish(ctx, options.ExtOptions{
		Name:    environment.Metadata.Name,
		Project: environment.Metadata.Project,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	triggerEvent := &eaaspb.TriggerEvent{}
	err = json.Unmarshal(response.Body, triggerEvent)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("environment published with trigger id %s", triggerEvent.GetId())

	d.SetId(environment.Metadata.Name)
	return diags
}

func resourceEnvironmentRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Println("environment read starts ")
	meta := GetMetaData(d)
	if meta == nil {
		return diag.FromErr(fmt.Errorf("%s", "failed to read environment "))
	}
	if d.State() != nil && d.State().ID != "" {
		meta.Name = d.State().ID
	}

	et, err := expandEnvironment(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, TF_USER_AGENT, options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		log.Println("read client err")
		return diag.FromErr(err)
	}

	environment, err := client.EaasV1().Environment().Get(ctx, options.GetOptions{
		Name:    meta.Name,
		Project: et.Metadata.Project,
	})
	if err != nil {
		log.Println("read get err")
		if strings.Contains(err.Error(), "code 404") {
			log.Println("Resource Read ", "error", err)
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	err = flattenEnvironment(d, environment)
	if err != nil {
		log.Println("read flatten err")
		return diag.FromErr(err)
	}
	return diags

}

func resourceEnvironmentUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return environmentUpsert(ctx, d, m)
}

func resourceEnvironmentDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Println("environment delete starts")
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

	env, err := expandEnvironment(d)
	if err != nil {
		log.Println("error while expanding environment during delete")
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, TF_USER_AGENT, options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.EaasV1().Environment().Delete(ctx, options.DeleteOptions{
		Name:    env.Metadata.Name,
		Project: env.Metadata.Project,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func expandEnvironment(in *schema.ResourceData) (*eaaspb.Environment, error) {
	log.Println("expand environment")
	if in == nil {
		return nil, fmt.Errorf("%s", "expand environment empty input")
	}
	obj := &eaaspb.Environment{}

	if v, ok := in.Get("metadata").([]interface{}); ok && len(v) > 0 {
		obj.Metadata = expandV3MetaData(v)
	}

	if v, ok := in.Get("spec").([]interface{}); ok && len(v) > 0 {
		objSpec, err := expandEnvironmentSpec(v)
		if err != nil {
			return nil, err
		}
		obj.Spec = objSpec
	}

	obj.ApiVersion = "eaas.envmgmt.io/v1"
	obj.Kind = "Environment"
	return obj, nil
}

func expandEnvironmentSpec(p []interface{}) (*eaaspb.EnvironmentSpec, error) {
	log.Println("expand environment spec")
	spec := &eaaspb.EnvironmentSpec{}
	if len(p) == 0 || p[0] == nil {
		return spec, fmt.Errorf("%s", "expand environment spec empty input")
	}

	in := p[0].(map[string]interface{})

	var err error
	if v, ok := in["template"].([]interface{}); ok && len(v) > 0 {
		spec.Template, err = expandTemplate(v)
		if err != nil {
			return spec, err
		}
	}

	if v, ok := in["variables"].([]interface{}); ok && len(v) > 0 {
		spec.Variables = expandVariables(v)
	}

	if v, ok := in["sharing"].([]interface{}); ok && len(v) > 0 {
		spec.Sharing = expandSharingSpec(v)
	}

	return spec, nil
}

func expandTemplate(p []interface{}) (*commonpb.ResourceNameAndVersionRef, error) {
	log.Println("expand template")
	if len(p) == 0 || p[0] == nil {
		return nil, fmt.Errorf("%s", "expand template empty input")
	}

	obj := &commonpb.ResourceNameAndVersionRef{}

	in := p[0].(map[string]interface{})

	if v, ok := in["name"].(string); ok && len(v) > 0 {
		obj.Name = v
	}

	if v, ok := in["version"].(string); ok && len(v) > 0 {
		obj.Version = v
	}

	return obj, nil
}

// Flatteners

func flattenEnvironment(d *schema.ResourceData, in *eaaspb.Environment) error {
	if in == nil {
		return nil
	}

	err := d.Set("metadata", flattenV3MetaData(in.Metadata))
	if err != nil {
		log.Println("flatten metadata err")
		return err
	}

	v, ok := d.Get("spec").([]interface{})
	if !ok {
		v = []interface{}{}
	}

	var ret []interface{}
	ret, err = flattenEnvironmentSpec(in.Spec, v)
	if err != nil {
		log.Println("flatten environment spec err")
		return err
	}

	err = d.Set("spec", ret)
	if err != nil {
		log.Println("set spec err")
		return err
	}
	return nil
}

func flattenEnvironmentSpec(in *eaaspb.EnvironmentSpec, p []interface{}) ([]interface{}, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "flatten environment spec empty input")
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if in.Template != nil {
		v, ok := obj["template"].([]interface{})
		if !ok {
			v = []interface{}{}
		}

		obj["template"] = flattenTemplate(in.Template, v)
	}

	if len(in.Variables) > 0 {
		v, ok := obj["variables"].([]interface{})
		if !ok {
			v = []interface{}{}
		}

		obj["variables"] = flattenVariables(in.Variables, v)
	}

	obj["sharing"] = flattenSharingSpec(in.Sharing)

	return []interface{}{obj}, nil
}

func flattenTemplate(input *commonpb.ResourceNameAndVersionRef, p []interface{}) []interface{} {
	log.Println("flatten template start", input)
	if input == nil {
		return nil
	}

	obj := map[string]interface{}{}
	if len(p) > 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(input.Name) > 0 {
		obj["name"] = input.Name
	}

	if len(input.Version) > 0 {
		obj["version"] = input.Version
	}

	return []interface{}{obj}
}

func resourceEnvironmentImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {

	log.Printf("Environment Import Starts")

	idParts := strings.SplitN(d.Id(), "/", 2)
	log.Println("resourceEnvironmentImport idParts:", idParts)

	log.Println("resourceEnvironmentImport Invoking expandEnvironment")
	env, err := expandEnvironment(d)
	if err != nil {
		log.Printf("resourceEnvironmentImport  expand error %s", err.Error())
	}

	var metaD commonpb.Metadata
	metaD.Name = idParts[0]
	metaD.Project = idParts[1]
	env.Metadata = &metaD

	err = d.Set("metadata", flattenV3MetaData(&metaD))
	if err != nil {
		log.Println("import set metadata err ", err)
		return nil, err
	}
	d.SetId(env.Metadata.Name)
	return []*schema.ResourceData{d}, nil

}
