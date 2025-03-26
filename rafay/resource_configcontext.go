package rafay

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/RafaySystems/rafay-common/pkg/hub/client/options"
	"github.com/RafaySystems/rafay-common/pkg/hub/client/typed"
	"github.com/RafaySystems/rafay-common/pkg/hub/terraform/resource"
	"github.com/RafaySystems/rafay-common/proto/types/hub/commonpb"
	"github.com/RafaySystems/rafay-common/proto/types/hub/eaaspb"
	"github.com/RafaySystems/rctl/pkg/config"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceConfigContext() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceConfigContextCreate,
		ReadContext:   resourceConfigContextRead,
		UpdateContext: resourceConfigContextUpdate,
		DeleteContext: resourceConfigContextDelete,
		Importer: &schema.ResourceImporter{
			State: resourceConfigContextImport,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema:        resource.ConfigContextSchema.Schema,
	}
}

func resourceConfigContextCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Println("config context create")
	diags := resourceConfigContextUpsert(ctx, d, m)
	if diags.HasError() {
		tflog := os.Getenv("TF_LOG")
		if tflog == "TRACE" || tflog == "DEBUG" {
			ctx = context.WithValue(ctx, "debug", "true")
		}
		cc, err := expandConfigContext(d)
		if err != nil {
			return diags
		}
		auth := config.GetConfig().GetAppAuthProfile()
		client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, TF_USER_AGENT, options.WithInsecureSkipVerify(auth.SkipServerCertValid))
		if err != nil {
			return diags
		}

		err = client.EaasV1().ConfigContext().Delete(ctx, options.DeleteOptions{
			Name:    cc.Metadata.Name,
			Project: cc.Metadata.Project,
		})
		if err != nil {
			return diags
		}
	}
	return diags
}

func resourceConfigContextUpsert(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("config context upsert starts")
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	cc, err := expandConfigContext(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, TF_USER_AGENT, options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.EaasV1().ConfigContext().Apply(ctx, cc, options.ApplyOptions{})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(cc.Metadata.Name)
	return diags
}

func resourceConfigContextRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Println("config context read starts ")
	meta := GetMetaData(d)
	if meta == nil {
		return diag.FromErr(fmt.Errorf("%s", "failed to read resource "))
	}
	if d.State() != nil && d.State().ID != "" {
		meta.Name = d.State().ID
	}

	cc, err := expandConfigContext(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, TF_USER_AGENT, options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		log.Println("read client err")
		return diag.FromErr(err)
	}

	configcontext, err := client.EaasV1().ConfigContext().Get(ctx, options.GetOptions{
		Name:    meta.Name,
		Project: cc.Metadata.Project,
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

	if !cc.GetSpec().GetSharing().GetEnabled() && configcontext.GetSpec().GetSharing() == nil {
		configcontext.Spec.Sharing = &commonpb.SharingSpec{}
		configcontext.Spec.Sharing.Enabled = false
		configcontext.Spec.Sharing.Projects = cc.Spec.Sharing.Projects
	}

	err = flattenConfigContext(d, configcontext)
	if err != nil {
		log.Println("read flatten err")
		return diag.FromErr(err)
	}
	return diags

}

func resourceConfigContextUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceConfigContextUpsert(ctx, d, m)
}

func resourceConfigContextDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Println("config context delete starts")
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

	cc, err := expandConfigContext(d)
	if err != nil {
		log.Println("error while expanding config context during delete")
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, TF_USER_AGENT, options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return diag.FromErr(err)
	}
	err = client.EaasV1().ConfigContext().Delete(ctx, options.DeleteOptions{
		Name:    cc.Metadata.Name,
		Project: cc.Metadata.Project,
	})

	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func expandConfigContext(in *schema.ResourceData) (*eaaspb.ConfigContext, error) {
	log.Println("expand config context resource")
	if in == nil {
		return nil, fmt.Errorf("%s", "expand config context empty input")
	}
	obj := &eaaspb.ConfigContext{}

	if v, ok := in.Get("metadata").([]interface{}); ok && len(v) > 0 {
		obj.Metadata = expandV3MetaData(v)
	}

	if v, ok := in.Get("spec").([]interface{}); ok && len(v) > 0 {
		objSpec, err := expandConfigContextSpec(v)
		if err != nil {
			return nil, err
		}
		obj.Spec = objSpec
	}

	obj.ApiVersion = "eaas.envmgmt.io/v1"
	obj.Kind = "ConfigContext"
	return obj, nil
}

func expandConfigContextSpec(p []interface{}) (*eaaspb.ConfigContextSpec, error) {
	log.Println("expand config context spec")
	spec := &eaaspb.ConfigContextSpec{}
	if len(p) == 0 || p[0] == nil {
		return spec, fmt.Errorf("expand config context spec empty input")
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["envs"].([]interface{}); ok && len(v) > 0 {
		spec.Envs = expandEnvVariables(v)
	}

	if v, ok := in["files"].([]interface{}); ok && len(v) > 0 {
		spec.Files = expandCommonpbFiles(v)
	}

	if v, ok := in["variables"].([]interface{}); ok && len(v) > 0 {
		spec.Variables = expandVariables(v)
	}

	if v, ok := in["sharing"].([]interface{}); ok && len(v) > 0 {
		spec.Sharing = expandSharingSpec(v)
	}

	return spec, nil
}

func expandEnvVariables(p []interface{}) []*eaaspb.EnvData {
	if len(p) == 0 || p[0] == nil {
		return []*eaaspb.EnvData{}
	}

	envvars := make([]*eaaspb.EnvData, len(p))

	for i := range p {
		obj := eaaspb.EnvData{}
		in := p[i].(map[string]interface{})

		if v, ok := in["key"].(string); ok && len(v) > 0 {
			obj.Key = v
		}

		if v, ok := in["value"].(string); ok && len(v) > 0 {
			obj.Value = v
		}

		if v, ok := in["sensitive"].(bool); ok {
			obj.Sensitive = v
		}

		if v, ok := in["options"].([]interface{}); ok && len(v) > 0 {
			obj.Options = expandEnvVarOptions(v)
		}

		envvars[i] = &obj

	}

	return envvars
}

func expandConfigContextCompoundRefs(p []interface{}) []*eaaspb.ConfigContextCompoundRef {
	var ccs []*eaaspb.ConfigContextCompoundRef
	if len(p) == 0 {
		return ccs
	}

	for i := range p {
		cc := expandConfigContextCompoundRef(p[i].(map[string]any))
		ccs = append(ccs, cc)
	}

	return ccs
}

func expandConfigContextCompoundRef(p map[string]any) *eaaspb.ConfigContextCompoundRef {
	cc := &eaaspb.ConfigContextCompoundRef{}
	if len(p) == 0 {
		return cc
	}

	if v, ok := p["name"].(string); ok && len(v) > 0 {
		cc.Name = v
	}

	if v, ok := p["data"].([]interface{}); ok && len(v) > 0 {
		cc.Data = expandConfigContextInline(v)
	}

	return cc
}

func expandConfigContextInline(p []interface{}) *eaaspb.ConfigContextInline {
	cc := &eaaspb.ConfigContextInline{}
	if len(p) == 0 || p[0] == nil {
		return cc
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["envs"].([]interface{}); ok && len(v) > 0 {
		cc.Envs = expandEnvVariables(v)
	}

	if v, ok := in["files"].([]interface{}); ok && len(v) > 0 {
		cc.Files = expandCommonpbFiles(v)
	}

	if v, ok := in["variables"].([]interface{}); ok && len(v) > 0 {
		cc.Variables = expandVariables(v)
	}

	return cc
}

// Flatteners

func flattenConfigContext(d *schema.ResourceData, in *eaaspb.ConfigContext) error {
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
	ret, err = flattenConfigContextSpec(in.Spec, v)
	if err != nil {
		log.Println("flatten config context spec err")
		return err
	}

	err = d.Set("spec", ret)
	if err != nil {
		log.Println("set spec err")
		return err
	}
	return nil
}

func flattenConfigContextSpec(in *eaaspb.ConfigContextSpec, p []interface{}) ([]interface{}, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "flatten config context spec empty input")
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	obj["envs"] = flattenEnvVariables(in.Envs, obj["envs"].([]interface{}))
	obj["files"] = flattenCommonpbFiles(in.Files)
	obj["variables"] = flattenVariables(in.Variables, obj["variables"].([]interface{}))
	obj["sharing"] = flattenSharingSpec(in.Sharing)

	return []interface{}{obj}, nil
}

func flattenEnvVariables(input []*eaaspb.EnvData, p []interface{}) []interface{} {
	log.Println("flatten environment variables start")
	if input == nil {
		return nil
	}

	out := make([]interface{}, len(input))
	for i, in := range input {
		log.Println("flatten environment variable ", in)
		obj := map[string]interface{}{}
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}
		obj["key"] = in.Key
		obj["value"] = in.Value
		obj["sensitive"] = in.Sensitive
		obj["options"] = flattenEnvVarOptions(in.Options)
		out[i] = &obj
	}

	return out
}

func flattenConfigContextCompoundRefs(input []*eaaspb.ConfigContextCompoundRef) []interface{} {
	if input == nil {
		return nil
	}

	var ccs []interface{}
	for _, cc := range input {
		ccs = append(ccs, flattenConfigContextCompoundRef(cc))
	}
	return ccs
}

func flattenConfigContextCompoundRef(input *eaaspb.ConfigContextCompoundRef) map[string]any {
	if input == nil {
		return nil
	}

	cc := make(map[string]any)
	if len(input.Name) > 0 {
		cc["name"] = input.Name
	}

	cc["data"] = flattenConfigContextInline(input.Data)

	return cc
}

func flattenConfigContextInline(input *eaaspb.ConfigContextInline) []interface{} {
	if input == nil {
		return nil
	}
	return []any{
		map[string]any{
			"envs":      flattenEnvVariables(input.Envs, nil),
			"files":     flattenCommonpbFiles(input.Files),
			"variables": flattenVariables(input.Variables, nil),
		},
	}
}

func resourceConfigContextImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {

	log.Printf("Config Context Import Starts")

	idParts := strings.SplitN(d.Id(), "/", 2)
	log.Println("resourceConfigContextImport idParts:", idParts)

	log.Println("resourceConfigContextImport Invoking expandConfigContext")
	cc, err := expandConfigContext(d)
	if err != nil {
		log.Printf("resourceConfigContextImport  expand error %s", err.Error())
	}

	var metaD commonpb.Metadata
	metaD.Name = idParts[0]
	metaD.Project = idParts[1]
	cc.Metadata = &metaD

	err = d.Set("metadata", flattenV3MetaData(&metaD))
	if err != nil {
		log.Println("import set metadata err ", err)
		return nil, err
	}
	d.SetId(cc.Metadata.Name)
	return []*schema.ResourceData{d}, nil

}
