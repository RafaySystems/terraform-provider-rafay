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
	"github.com/RafaySystems/rafay-common/proto/types/hub/eaaspb"
	"github.com/RafaySystems/rctl/pkg/config"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceResourceTemplate() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceTemplateCreate,
		ReadContext:   resourceTemplateRead,
		UpdateContext: resourceTemplateUpdate,
		DeleteContext: resourceTemplateDelete,
		Importer: &schema.ResourceImporter{
			State: resourceResourceTemplateImport,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema:        resource.ResourceTemplateSchema.Schema,
	}
}

func resourceTemplateCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Println("resource template create")
	diags := resourceTemplateUpsert(ctx, d, m)
	if diags.HasError() {
		tflog := os.Getenv("TF_LOG")
		if tflog == "TRACE" || tflog == "DEBUG" {
			ctx = context.WithValue(ctx, "debug", "true")
		}
		resourcetemplate, err := expandResourceTemplate(d)
		if err != nil {
			return diags
		}
		auth := config.GetConfig().GetAppAuthProfile()
		client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, TF_USER_AGENT, options.WithInsecureSkipVerify(auth.SkipServerCertValid))
		if err != nil {
			return diags
		}

		err = client.EaasV1().ResourceTemplate().Delete(ctx, options.DeleteOptions{
			Name:    resourcetemplate.Metadata.Name,
			Project: resourcetemplate.Metadata.Project,
		})
		if err != nil {
			return diags
		}
	}
	return diags
}

func resourceTemplateUpsert(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("resource template upsert starts")
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	resourcetemplate, err := expandResourceTemplate(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, TF_USER_AGENT, options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.EaasV1().ResourceTemplate().Apply(ctx, resourcetemplate, options.ApplyOptions{})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(resourcetemplate.Metadata.Name)
	return diags
}

func resourceTemplateRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Println("resource template read starts ")
	meta := GetMetaData(d)
	if meta == nil {
		return diag.FromErr(fmt.Errorf("%s", "failed to read resource template "))
	}
	if d.State() != nil && d.State().ID != "" {
		meta.Name = d.State().ID
	}

	rt, err := expandResourceTemplate(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, TF_USER_AGENT, options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		log.Println("read client err")
		return diag.FromErr(err)
	}

	resourcetemplate, err := client.EaasV1().ResourceTemplate().Get(ctx, options.GetOptions{
		Name:    meta.Name,
		Project: rt.Metadata.Project,
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

	err = flattenResourceTemplate(d, resourcetemplate)
	if err != nil {
		log.Println("read flatten err")
		return diag.FromErr(err)
	}
	return diags

}

func resourceTemplateUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceTemplateUpsert(ctx, d, m)
}

func resourceTemplateDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Println("resource template delete starts")
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

	rt, err := expandResourceTemplate(d)
	if err != nil {
		log.Println("error while expanding resource template during delete")
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, TF_USER_AGENT, options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return diag.FromErr(err)
	}
	err = client.EaasV1().ResourceTemplate().Delete(ctx, options.DeleteOptions{
		Name:    rt.Metadata.Name,
		Project: rt.Metadata.Project,
	})

	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func expandResourceTemplate(in *schema.ResourceData) (*eaaspb.ResourceTemplate, error) {
	log.Println("expand resource template")
	if in == nil {
		return nil, fmt.Errorf("%s", "expand resource template empty input")
	}
	obj := &eaaspb.ResourceTemplate{}

	if v, ok := in.Get("metadata").([]interface{}); ok && len(v) > 0 {
		obj.Metadata = expandV3MetaData(v)
	}

	if v, ok := in.Get("spec").([]interface{}); ok && len(v) > 0 {
		objSpec, err := expandResourceTemplateSpec(v)
		if err != nil {
			return nil, err
		}
		obj.Spec = objSpec
	}

	obj.ApiVersion = "eaas.envmgmt.io/v1"
	obj.Kind = "ResourceTemplate"
	return obj, nil
}

func expandResourceTemplateSpec(p []interface{}) (*eaaspb.ResourceTemplateSpec, error) {
	log.Println("expand resource template spec")
	spec := &eaaspb.ResourceTemplateSpec{}
	if len(p) == 0 || p[0] == nil {
		return spec, fmt.Errorf("%s", "expand resource template spec empty input")
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["version"].(string); ok && len(v) > 0 {
		spec.Version = v
	}

	if p, ok := in["provider"].(string); ok && len(p) > 0 {
		spec.Provider = p
	}

	if po, ok := in["provider_options"].([]interface{}); ok {
		spec.ProviderOptions = expandProviderOptions(po)
	}

	if ro, ok := in["repository_options"].([]interface{}); ok {
		spec.RepositoryOptions = expandResourceTemplateRepositoryOptions(ro)
	}

	if v, ok := in["contexts"].([]interface{}); ok {
		spec.Contexts = expandContexts(v)
	}

	if v, ok := in["variables"].([]interface{}); ok {
		spec.Variables = expandVariables(v)
	}

	if h, ok := in["hooks"].([]interface{}); ok {
		spec.Hooks = expandResourceHooks(h)
	}

	if ag, ok := in["agents"].([]interface{}); ok {
		spec.Agents = expandEaasAgents(ag)
	}

	if v, ok := in["sharing"].([]interface{}); ok {
		spec.Sharing = expandSharingSpec(v)
	}

	return spec, nil
}

func expandProviderOptions(p []interface{}) *eaaspb.ResourceTemplateProviderOptions {
	po := &eaaspb.ResourceTemplateProviderOptions{}
	if len(p) == 0 || p[0] == nil {
		return po
	}
	in := p[0].(map[string]interface{})

	if tp, ok := in["terraform"].([]interface{}); ok && len(tp) > 0 {
		po.Terraform = expandTerraformProviderOptions(tp)
	}

	if s, ok := in["system"].([]interface{}); ok && len(s) > 0 {
		po.System = expandSystemProviderOptions(s)
	}

	if t, ok := in["terragrunt"].([]interface{}); ok && len(t) > 0 {
		po.Terragrunt = expandTerragruntProviderOptions(t)
	}

	if p, ok := in["pulumi"].([]interface{}); ok && len(p) > 0 {
		po.Pulumi = expandPulumiProviderOptions(p)
	}

	return po

}

func expandResourceTemplateRepositoryOptions(p []interface{}) *eaaspb.ResourceTemplateRepositoryOptions {
	ro := &eaaspb.ResourceTemplateRepositoryOptions{}
	if len(p) == 0 || p[0] == nil {
		return ro
	}
	in := p[0].(map[string]interface{})

	if n, ok := in["name"].(string); ok && len(n) > 0 {
		ro.Name = n
	}

	if b, ok := in["branch"].(string); ok && len(b) > 0 {
		ro.Branch = b
	}

	if dp, ok := in["directory_path"].(string); ok && len(dp) > 0 {
		ro.DirectoryPath = dp
	}

	return ro
}

func expandContexts(p []interface{}) []*commonpb.ResourceNameAndVersionRef {
	ctxs := make([]*commonpb.ResourceNameAndVersionRef, 0)
	if len(p) == 0 {
		return ctxs
	}

	for indx := range p {
		obj := &commonpb.ResourceNameAndVersionRef{}

		in := p[indx].(map[string]interface{})

		if v, ok := in["name"].(string); ok && len(v) > 0 {
			obj.Name = v
		}

		ctxs = append(ctxs, obj)
	}

	return ctxs
}

func expandResourceHooks(p []interface{}) *eaaspb.ResourceHooks {
	hooks := &eaaspb.ResourceHooks{}

	if len(p) == 0 || p[0] == nil {
		return hooks
	}

	in := p[0].(map[string]interface{})

	if h, ok := in["on_completion"].([]interface{}); ok && len(h) > 0 {
		hooks.OnCompletion = expandEaasHooks(h)
	}

	if h, ok := in["on_success"].([]interface{}); ok && len(h) > 0 {
		hooks.OnSuccess = expandEaasHooks(h)
	}

	if h, ok := in["on_failure"].([]interface{}); ok && len(h) > 0 {
		hooks.OnFailure = expandEaasHooks(h)
	}

	if h, ok := in["on_init"].([]interface{}); ok && len(h) > 0 {
		hooks.OnInit = expandEaasHooks(h)
	}

	if h, ok := in["provider"].([]interface{}); ok && len(h) > 0 {
		hooks.Provider = expandProviderHooks(h)
	}

	return hooks

}

func expandEaasAgents(p []interface{}) []*commonpb.ResourceNameAndVersionRef {
	agents := make([]*commonpb.ResourceNameAndVersionRef, 0)
	if len(p) == 0 {
		return agents
	}

	for indx := range p {
		obj := &commonpb.ResourceNameAndVersionRef{}

		in := p[indx].(map[string]interface{})

		if v, ok := in["name"].(string); ok && len(v) > 0 {
			obj.Name = v
		}

		if v, ok := in["version"].(string); ok && len(v) > 0 {
			obj.Version = v
		}

		agents = append(agents, obj)
	}

	return agents
}

func expandTerraformProviderOptions(p []interface{}) *eaaspb.TerraformProviderOptions {
	tpo := &eaaspb.TerraformProviderOptions{}
	if len(p) == 0 || p[0] == nil {
		return tpo
	}

	in := p[0].(map[string]interface{})

	if h, ok := in["version"].(string); ok {
		tpo.Version = h
	}

	if h, ok := in["use_system_state_store"].([]interface{}); ok && len(h) > 0 {
		tpo.UseSystemStateStore = expandBoolValue(h)
	}

	if vfiles, ok := in["var_files"].([]interface{}); ok && len(vfiles) > 0 {
		tpo.VarFiles = toArrayString(vfiles)
	}

	if bcfgs, ok := in["backend_configs"].([]interface{}); ok && len(bcfgs) > 0 {
		tpo.BackendConfigs = toArrayString(bcfgs)
	}

	if h, ok := in["refresh"].([]interface{}); ok && len(h) > 0 {
		tpo.Refresh = expandBoolValue(h)
	}

	if h, ok := in["lock"].([]interface{}); ok && len(h) > 0 {
		tpo.Lock = expandBoolValue(h)
	}

	if h, ok := in["lock_timeout_seconds"].(int); ok {
		tpo.LockTimeoutSeconds = uint64(h)
	}

	if pdirs, ok := in["plugin_dirs"].([]interface{}); ok && len(pdirs) > 0 {
		tpo.PluginDirs = toArrayString(pdirs)
	}

	if tgtrs, ok := in["target_resources"].([]interface{}); ok && len(tgtrs) > 0 {
		tpo.TargetResources = toArrayString(tgtrs)
	}

	return tpo
}

func expandSystemProviderOptions(p []interface{}) *eaaspb.SystemProviderOptions {
	spo := &eaaspb.SystemProviderOptions{}
	if len(p) == 0 || p[0] == nil {
		return spo
	}

	return spo

}

func expandTerragruntProviderOptions(p []interface{}) *eaaspb.TerragruntProviderOptions {
	tpo := &eaaspb.TerragruntProviderOptions{}
	if len(p) == 0 || p[0] == nil {
		return tpo
	}

	return tpo
}

func expandPulumiProviderOptions(p []interface{}) *eaaspb.PulumiProviderOptions {
	ppo := &eaaspb.PulumiProviderOptions{}
	if len(p) == 0 || p[0] == nil {
		return ppo
	}

	return ppo
}

func expandProviderHooks(p []interface{}) *eaaspb.ResourceTemplateProviderHooks {
	rtph := &eaaspb.ResourceTemplateProviderHooks{}
	if len(p) == 0 || p[0] == nil {
		return rtph
	}

	in := p[0].(map[string]interface{})

	if h, ok := in["terraform"].([]interface{}); ok && len(h) > 0 {
		rtph.Terraform = expandTerraformProviderHooks(h)
	}

	if h, ok := in["pulumi"].([]interface{}); ok && len(h) > 0 {
		rtph.Pulumi = expandPulumiProviderHooks(h)
	}

	return rtph

}

func expandTerraformProviderHooks(p []interface{}) *eaaspb.TerraformProviderHooks {
	tph := &eaaspb.TerraformProviderHooks{}
	if len(p) == 0 || p[0] == nil {
		return tph
	}

	in := p[0].(map[string]interface{})

	if h, ok := in["deploy"].([]interface{}); ok && len(h) > 0 {
		tph.Deploy = expandTerraformDeployHooks(h)
	}

	if h, ok := in["destroy"].([]interface{}); ok && len(h) > 0 {
		tph.Destroy = expandTerraformDestroyHooks(h)
	}

	return tph
}

func expandPulumiProviderHooks(p []interface{}) *eaaspb.PulumiProviderHooks {
	pph := &eaaspb.PulumiProviderHooks{}
	if len(p) == 0 || p[0] == nil {
		return pph
	}

	in := p[0].(map[string]interface{})

	if h, ok := in["deploy"].([]interface{}); ok {
		pph.Deploy = expandPulumiDeployHooks(h)
	}

	if h, ok := in["destroy"].([]interface{}); ok {
		pph.Destroy = expandPulumiDestroyHooks(h)
	}

	return pph
}

func expandTerraformDeployHooks(p []interface{}) *eaaspb.TerraformDeployHooks {
	tdh := &eaaspb.TerraformDeployHooks{}
	if len(p) == 0 || p[0] == nil {
		return tdh
	}

	in := p[0].(map[string]interface{})

	if h, ok := in["init"].([]interface{}); ok && len(h) > 0 {
		tdh.Init = expandLifecycleEventHooks(h)
	}

	if h, ok := in["plan"].([]interface{}); ok && len(h) > 0 {
		tdh.Plan = expandLifecycleEventHooks(h)
	}

	if h, ok := in["apply"].([]interface{}); ok && len(h) > 0 {
		tdh.Apply = expandLifecycleEventHooks(h)
	}

	if h, ok := in["output"].([]interface{}); ok && len(h) > 0 {
		tdh.Output = expandLifecycleEventHooks(h)
	}

	return tdh
}

func expandTerraformDestroyHooks(p []interface{}) *eaaspb.TerraformDestroyHooks {
	tdh := &eaaspb.TerraformDestroyHooks{}
	if len(p) == 0 || p[0] == nil {
		return tdh
	}

	in := p[0].(map[string]interface{})

	if h, ok := in["init"].([]interface{}); ok && len(h) > 0 {
		tdh.Init = expandLifecycleEventHooks(h)
	}

	if h, ok := in["plan"].([]interface{}); ok && len(h) > 0 {
		tdh.Plan = expandLifecycleEventHooks(h)
	}

	if h, ok := in["destroy"].([]interface{}); ok && len(h) > 0 {
		tdh.Destroy = expandLifecycleEventHooks(h)
	}

	return tdh
}

func expandPulumiDeployHooks(p []interface{}) *eaaspb.PulumiDeployHooks {
	pdh := &eaaspb.PulumiDeployHooks{}
	if len(p) == 0 || p[0] == nil {
		return pdh
	}

	in := p[0].(map[string]interface{})

	if h, ok := in["login"].([]interface{}); ok && len(h) > 0 {
		pdh.Login = expandLifecycleEventHooks(h)
	}

	if h, ok := in["preview"].([]interface{}); ok && len(h) > 0 {
		pdh.Preview = expandLifecycleEventHooks(h)
	}

	if h, ok := in["up"].([]interface{}); ok && len(h) > 0 {
		pdh.Up = expandLifecycleEventHooks(h)
	}

	if h, ok := in["output"].([]interface{}); ok && len(h) > 0 {
		pdh.Output = expandLifecycleEventHooks(h)
	}

	return pdh
}

func expandPulumiDestroyHooks(p []interface{}) *eaaspb.PulumiDestroyHooks {
	pdh := &eaaspb.PulumiDestroyHooks{}
	if len(p) == 0 || p[0] == nil {
		return pdh
	}

	in := p[0].(map[string]interface{})

	if h, ok := in["login"].([]interface{}); ok && len(h) > 0 {
		pdh.Login = expandLifecycleEventHooks(h)
	}

	if h, ok := in["preview"].([]interface{}); ok && len(h) > 0 {
		pdh.Preview = expandLifecycleEventHooks(h)
	}

	if h, ok := in["destroy"].([]interface{}); ok && len(h) > 0 {
		pdh.Destroy = expandLifecycleEventHooks(h)
	}

	return pdh
}

func expandLifecycleEventHooks(p []interface{}) *eaaspb.LifecycleEventHooks {
	lh := &eaaspb.LifecycleEventHooks{}
	if len(p) == 0 || p[0] == nil {
		return lh
	}

	in := p[0].(map[string]interface{})

	if h, ok := in["before"].([]interface{}); ok && len(h) > 0 {
		lh.Before = expandEaasHooks(h)
	}

	if h, ok := in["after"].([]interface{}); ok && len(h) > 0 {
		lh.After = expandEaasHooks(h)
	}

	return lh
}

// Flatteners

func flattenResourceTemplate(d *schema.ResourceData, in *eaaspb.ResourceTemplate) error {
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
	ret, err = flattenResourceTemplateSpec(in.Spec, v)
	if err != nil {
		log.Println("flatten resource template spec err")
		return err
	}

	err = d.Set("spec", ret)
	if err != nil {
		log.Println("set spec err")
		return err
	}
	return nil
}

func flattenResourceTemplateSpec(in *eaaspb.ResourceTemplateSpec, p []interface{}) ([]interface{}, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "flatten resource spec empty input")
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	obj["version"] = in.Version
	obj["provider"] = in.Provider
	obj["provider_options"] = flattenProviderOptions(in.ProviderOptions)
	obj["repository_options"] = flattenRepositoryOptions(in.RepositoryOptions)

	if len(in.Contexts) > 0 {
		v, ok := obj["contexts"].([]interface{})
		if !ok {
			v = []interface{}{}
		}

		obj["contexts"] = flattenContexts(in.Contexts, v)
	}

	if len(in.Variables) > 0 {
		v, ok := obj["variables"].([]interface{})
		if !ok {
			v = []interface{}{}
		}

		obj["variables"] = flattenVariables(in.Variables, v)
	}

	if in.Hooks != nil {
		v, ok := obj["hooks"].([]interface{})
		if !ok {
			v = []interface{}{}
		}

		obj["hooks"] = flattenResourceHooks(in.Hooks, v)
	}

	obj["agents"] = flattenEaasAgents(in.Agents)
	obj["sharing"] = flattenSharingSpec(in.Sharing)

	return []interface{}{obj}, nil
}

func flattenProviderOptions(in *eaaspb.ResourceTemplateProviderOptions) []interface{} {
	if in == nil {
		return nil
	}

	obj := make(map[string]interface{})
	obj["terraform"] = flattenTerraformProviderOptions(in.Terraform)
	obj["system"] = flattenSystemProviderOptions(in.System)
	obj["terragrunt"] = flattenTerragruntProviderOptions(in.Terragrunt)
	obj["pulumi"] = flattenPulumiProviderOptions(in.Pulumi)

	return []interface{}{obj}
}

func flattenTerraformProviderOptions(in *eaaspb.TerraformProviderOptions) []interface{} {
	if in == nil {
		return nil
	}

	obj := make(map[string]interface{})
	obj["version"] = in.Version
	obj["use_system_state_store"] = flattenBoolValue(in.UseSystemStateStore)
	obj["var_files"] = toArrayInterface(in.VarFiles)
	obj["backend_configs"] = toArrayInterface(in.BackendConfigs)
	obj["refresh"] = flattenBoolValue(in.Refresh)
	obj["lock"] = flattenBoolValue(in.Lock)
	obj["lock_timeout_seconds"] = in.LockTimeoutSeconds
	obj["plugin_dirs"] = toArrayInterface(in.PluginDirs)
	obj["target_resources"] = toArrayInterface(in.TargetResources)

	return []interface{}{obj}
}

func flattenSystemProviderOptions(in *eaaspb.SystemProviderOptions) []interface{} {
	if in == nil {
		return nil
	}

	obj := make(map[string]interface{})
	return []interface{}{obj}
}

func flattenTerragruntProviderOptions(in *eaaspb.TerragruntProviderOptions) []interface{} {
	if in == nil {
		return nil
	}

	obj := make(map[string]interface{})
	return []interface{}{obj}
}

func flattenPulumiProviderOptions(in *eaaspb.PulumiProviderOptions) []interface{} {
	if in == nil {
		return nil
	}

	obj := make(map[string]interface{})
	return []interface{}{obj}
}

func flattenRetryOptions(in *eaaspb.RetryOptions) []interface{} {
	if in == nil {
		return nil
	}

	obj := make(map[string]interface{})
	obj["enabled"] = in.Enabled
	obj["max_count"] = in.MaxCount

	return []interface{}{obj}
}

func flattenRepositoryOptions(in *eaaspb.ResourceTemplateRepositoryOptions) []interface{} {
	if in == nil {
		return nil
	}

	obj := make(map[string]interface{})
	obj["name"] = in.Name
	obj["branch"] = in.Branch
	obj["directory_path"] = in.DirectoryPath

	return []interface{}{obj}
}

func flattenContexts(input []*commonpb.ResourceNameAndVersionRef, p []interface{}) []interface{} {
	log.Println("flatten contexts start")
	if input == nil {
		return nil
	}

	out := make([]interface{}, len(input))
	for i, in := range input {
		log.Println("flatten context ", in)
		obj := map[string]interface{}{}
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}

		if len(in.Name) > 0 {
			obj["name"] = in.Name
		}

		out[i] = &obj
	}

	return out
}

func flattenResourceHooks(in *eaaspb.ResourceHooks, p []interface{}) []interface{} {
	log.Println("flatten resource hooks start")
	if in == nil {
		return nil
	}

	obj := make(map[string]interface{})
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.OnCompletion) > 0 {
		v, ok := obj["on_completion"].([]interface{})
		if !ok {
			v = []interface{}{}
		}

		obj["on_completion"] = flattenEaasHooks(in.OnCompletion, v)
	}

	if len(in.OnSuccess) > 0 {
		v, ok := obj["on_success"].([]interface{})
		if !ok {
			v = []interface{}{}
		}

		obj["on_success"] = flattenEaasHooks(in.OnSuccess, v)
	}

	if len(in.OnFailure) > 0 {
		v, ok := obj["on_failure"].([]interface{})
		if !ok {
			v = []interface{}{}
		}

		obj["on_failure"] = flattenEaasHooks(in.OnFailure, v)
	}

	if len(in.OnInit) > 0 {
		v, ok := obj["on_init"].([]interface{})
		if !ok {
			v = []interface{}{}
		}

		obj["on_init"] = flattenEaasHooks(in.OnInit, v)
	}

	if in.Provider != nil {
		v, ok := obj["provider"].([]interface{})
		if !ok {
			v = []interface{}{}
		}

		obj["provider"] = flattenProviderHooks(in.Provider, v)
	}
	return []interface{}{obj}
}

func flattenProviderHooks(input *eaaspb.ResourceTemplateProviderHooks, p []interface{}) []interface{} {
	log.Println("flatten provider hooks start")
	if input == nil {
		return nil
	}

	obj := make(map[string]interface{})
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if input.Terraform != nil {
		v, ok := obj["terraform"].([]interface{})
		if !ok {
			v = []interface{}{}
		}

		obj["terraform"] = flattenTerraformProviderHooks(input.Terraform, v)
	}

	if input.Pulumi != nil {
		v, ok := obj["pulumi"].([]interface{})
		if !ok {
			v = []interface{}{}
		}

		obj["pulumi"] = flattenPulumiProviderHooks(input.Pulumi, v)
	}

	return []interface{}{obj}
}

func flattenTerraformProviderHooks(input *eaaspb.TerraformProviderHooks, p []interface{}) []interface{} {
	log.Println("flatten terraform provider hooks start")
	if input == nil {
		return nil
	}

	obj := make(map[string]interface{})
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if input.Deploy != nil {
		v, ok := obj["deploy"].([]interface{})
		if !ok {
			v = []interface{}{}
		}

		obj["deploy"] = flattenTerraformDeployHooks(input.Deploy, v)
	}

	if input.Destroy != nil {
		v, ok := obj["destroy"].([]interface{})
		if !ok {
			v = []interface{}{}
		}

		obj["destroy"] = flattenTerraformDestroyHooks(input.Destroy, v)
	}

	return []interface{}{obj}
}

func flattenPulumiProviderHooks(input *eaaspb.PulumiProviderHooks, p []interface{}) []interface{} {
	log.Println("flatten pulumi provider hooks start")
	if input == nil {
		return nil
	}

	obj := make(map[string]interface{})
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if input.Deploy != nil {
		v, ok := obj["deploy"].([]interface{})
		if !ok {
			v = []interface{}{}
		}

		obj["deploy"] = flattenPulumiDeployHooks(input.Deploy, v)
	}

	if input.Destroy != nil {
		v, ok := obj["destroy"].([]interface{})
		if !ok {
			v = []interface{}{}
		}

		obj["destroy"] = flattenPulumiDestroyHooks(input.Destroy, v)
	}

	return []interface{}{obj}
}

func flattenTerraformDeployHooks(input *eaaspb.TerraformDeployHooks, p []interface{}) []interface{} {
	log.Println("flatten terraform deploy hooks start")
	if input == nil {
		return nil
	}

	obj := make(map[string]interface{})
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if input.Init != nil {
		v, ok := obj["init"].([]interface{})
		if !ok {
			v = []interface{}{}
		}

		obj["init"] = flattenLifecycleEventHooks(input.Init, v)
	}

	if input.Plan != nil {
		v, ok := obj["plan"].([]interface{})
		if !ok {
			v = []interface{}{}
		}

		obj["plan"] = flattenLifecycleEventHooks(input.Plan, v)
	}

	if input.Apply != nil {
		v, ok := obj["apply"].([]interface{})
		if !ok {
			v = []interface{}{}
		}

		obj["apply"] = flattenLifecycleEventHooks(input.Apply, v)
	}

	if input.Output != nil {
		v, ok := obj["output"].([]interface{})
		if !ok {
			v = []interface{}{}
		}

		obj["output"] = flattenLifecycleEventHooks(input.Output, v)
	}

	return []interface{}{obj}
}

func flattenTerraformDestroyHooks(input *eaaspb.TerraformDestroyHooks, p []interface{}) []interface{} {
	log.Println("flatten terraform destroy hooks start")
	if input == nil {
		return nil
	}

	obj := make(map[string]interface{})
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if input.Init != nil {
		v, ok := obj["init"].([]interface{})
		if !ok {
			v = []interface{}{}
		}

		obj["init"] = flattenLifecycleEventHooks(input.Init, v)
	}

	if input.Plan != nil {
		v, ok := obj["plan"].([]interface{})
		if !ok {
			v = []interface{}{}
		}

		obj["plan"] = flattenLifecycleEventHooks(input.Plan, v)
	}

	if input.Destroy != nil {
		v, ok := obj["destroy"].([]interface{})
		if !ok {
			v = []interface{}{}
		}

		obj["destroy"] = flattenLifecycleEventHooks(input.Destroy, v)
	}

	return []interface{}{obj}
}

func flattenPulumiDeployHooks(input *eaaspb.PulumiDeployHooks, p []interface{}) []interface{} {
	log.Println("flatten pulumi deploy hooks start")
	if input == nil {
		return nil
	}

	obj := make(map[string]interface{})
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if input.Login != nil {
		v, ok := obj["login"].([]interface{})
		if !ok {
			v = []interface{}{}
		}

		obj["login"] = flattenLifecycleEventHooks(input.Login, v)
	}

	if input.Preview != nil {
		v, ok := obj["preview"].([]interface{})
		if !ok {
			v = []interface{}{}
		}

		obj["preview"] = flattenLifecycleEventHooks(input.Preview, v)
	}

	if input.Up != nil {
		v, ok := obj["up"].([]interface{})
		if !ok {
			v = []interface{}{}
		}

		obj["up"] = flattenLifecycleEventHooks(input.Up, v)
	}

	if input.Output != nil {
		v, ok := obj["output"].([]interface{})
		if !ok {
			v = []interface{}{}
		}

		obj["output"] = flattenLifecycleEventHooks(input.Output, v)
	}

	return []interface{}{obj}
}

func flattenPulumiDestroyHooks(input *eaaspb.PulumiDestroyHooks, p []interface{}) []interface{} {
	log.Println("flatten pulumi destroy hooks start")
	if input == nil {
		return nil
	}

	obj := make(map[string]interface{})
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if input.Login != nil {
		v, ok := obj["login"].([]interface{})
		if !ok {
			v = []interface{}{}
		}

		obj["login"] = flattenLifecycleEventHooks(input.Login, v)
	}

	if input.Preview != nil {
		v, ok := obj["preview"].([]interface{})
		if !ok {
			v = []interface{}{}
		}

		obj["preview"] = flattenLifecycleEventHooks(input.Preview, v)
	}

	if input.Destroy != nil {
		v, ok := obj["destroy"].([]interface{})
		if !ok {
			v = []interface{}{}
		}

		obj["destroy"] = flattenLifecycleEventHooks(input.Destroy, v)
	}

	return []interface{}{obj}
}

func flattenLifecycleEventHooks(input *eaaspb.LifecycleEventHooks, p []interface{}) []interface{} {
	if input == nil {
		return nil
	}

	obj := make(map[string]interface{})
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(input.Before) > 0 {
		v, ok := obj["before"].([]interface{})
		if !ok {
			v = []interface{}{}
		}

		obj["before"] = flattenEaasHooks(input.Before, v)
	}

	if len(input.After) > 0 {
		v, ok := obj["after"].([]interface{})
		if !ok {
			v = []interface{}{}
		}

		obj["after"] = flattenEaasHooks(input.After, v)
	}

	return []interface{}{obj}
}

func flattenEaasAgents(input []*commonpb.ResourceNameAndVersionRef) []interface{} {
	log.Println("flatten agents start")
	if len(input) == 0 {
		return nil
	}

	out := make([]interface{}, len(input))
	for i, in := range input {
		log.Println("flatten agent ", in)
		obj := map[string]interface{}{}

		if len(in.Name) > 0 {
			obj["name"] = in.Name
		}

		out[i] = &obj
	}

	return out
}

func resourceResourceTemplateImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {

	log.Printf("Resource Template Import Starts")

	idParts := strings.SplitN(d.Id(), "/", 2)
	log.Println("resourceResourceTemplateImport idParts:", idParts)

	log.Println("resourceResourceTemplateImport Invoking expandResourceTemplate")
	rt, err := expandResourceTemplate(d)
	if err != nil {
		log.Printf("resourceResourceTemplateImport  expand error %s", err.Error())
	}

	var metaD commonpb.Metadata
	metaD.Name = idParts[0]
	metaD.Project = idParts[1]
	rt.Metadata = &metaD

	err = d.Set("metadata", flattenV3MetaData(&metaD))
	if err != nil {
		log.Println("import set metadata err ", err)
		return nil, err
	}
	d.SetId(rt.Metadata.Name)
	return []*schema.ResourceData{d}, nil

}
