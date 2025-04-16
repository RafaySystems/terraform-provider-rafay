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

	if vs, ok := in["version_state"].(string); ok && len(vs) > 0 {
		spec.VersionState = vs
	}

	if p, ok := in["provider"].(string); ok && len(p) > 0 {
		spec.Provider = p
	}

	var err error
	if po, ok := in["provider_options"].([]interface{}); ok && len(po) > 0 {
		spec.ProviderOptions, err = expandProviderOptions(po)
		if err != nil {
			return nil, err
		}
	}

	if ro, ok := in["repository_options"].([]interface{}); ok && len(ro) > 0 {
		spec.RepositoryOptions = expandResourceTemplateRepositoryOptions(ro)
	}

	if v, ok := in["contexts"].([]interface{}); ok && len(v) > 0 {
		spec.Contexts = expandContexts(v)
	}

	if v, ok := in["variables"].([]interface{}); ok && len(v) > 0 {
		spec.Variables = expandVariables(v)
	}

	if h, ok := in["hooks"].([]interface{}); ok && len(h) > 0 {
		spec.Hooks, err = expandResourceHooks(h)
		if err != nil {
			return nil, err
		}
	}

	if ag, ok := in["agents"].([]interface{}); ok && len(ag) > 0 {
		spec.Agents = expandEaasAgents(ag)
	}

	if v, ok := in["sharing"].([]interface{}); ok && len(v) > 0 {
		spec.Sharing = expandSharingSpec(v)
	}

	if ad, ok := in["artifact_driver"].([]interface{}); ok && len(ad) > 0 {
		log.Println("WARN: artifact_driver is deprecated, use artifact_workflow_handler instead")
		spec.ArtifactDriver, err = expandWorkflowHandlerCompoundRef(ad)
		if err != nil {
			return nil, err
		}
	}

	if ad, ok := in["artifact_workflow_handler"].([]interface{}); ok && len(ad) > 0 {
		spec.ArtifactWorkflowHandler, err = expandWorkflowHandlerCompoundRef(ad)
		if err != nil {
			return nil, err
		}
	}

	if s, ok := in["actions"].([]interface{}); ok && len(s) > 0 {
		spec.Actions, err = expandActions(s)
		if err != nil {
			return nil, err
		}
	}

	return spec, nil
}

func expandProviderOptions(p []interface{}) (*eaaspb.ResourceTemplateProviderOptions, error) {
	po := &eaaspb.ResourceTemplateProviderOptions{}
	if len(p) == 0 || p[0] == nil {
		return po, nil
	}
	in := p[0].(map[string]interface{})

	var err error
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

	if p, ok := in["driver"].([]interface{}); ok && len(p) > 0 {
		log.Println("WARN: spec.provider_options.driver is deprecated, use spec.provider_options.workflow_handler instead")
		po.Driver, err = expandWorkflowHandlerCompoundRef(p)
		if err != nil {
			return nil, err
		}
	}

	if p, ok := in["workflow_handler"].([]interface{}); ok && len(p) > 0 {
		po.WorkflowHandler, err = expandWorkflowHandlerCompoundRef(p)
		if err != nil {
			return nil, err
		}
	}

	if p, ok := in["open_tofu"].([]interface{}); ok && len(p) > 0 {
		po.OpenTofu = expandOpenTofuProviderOptions(p)
	}

	if w, ok := in["custom"].([]interface{}); ok && len(p) > 0 {
		po.Custom, err = expandCustomProviderOptions(w)
		if err != nil {
			return nil, err
		}
	}

	if p, ok := in["hcp_terraform"].([]interface{}); ok && len(p) > 0 {
		po.HcpTerraform = expandHcpTerraformProviderOptions(p)
	}

	return po, nil

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

func expandContexts(p []interface{}) []*eaaspb.ConfigContextCompoundRef {
	ctxs := make([]*eaaspb.ConfigContextCompoundRef, 0)
	if len(p) == 0 {
		return ctxs
	}

	for indx := range p {
		obj := &eaaspb.ConfigContextCompoundRef{}

		in := p[indx].(map[string]interface{})

		if v, ok := in["name"].(string); ok && len(v) > 0 {
			obj.Name = v
		}

		ctxs = append(ctxs, obj)
	}

	return ctxs
}

func expandResourceHooks(p []interface{}) (*eaaspb.ResourceHooks, error) {
	hooks := &eaaspb.ResourceHooks{}

	if len(p) == 0 || p[0] == nil {
		return hooks, nil
	}

	in := p[0].(map[string]interface{})

	var err error
	if h, ok := in["on_completion"].([]interface{}); ok && len(h) > 0 {
		hooks.OnCompletion, err = expandEaasHooks(h)
		if err != nil {
			return nil, err
		}
	}

	if h, ok := in["on_success"].([]interface{}); ok && len(h) > 0 {
		hooks.OnSuccess, err = expandEaasHooks(h)
		if err != nil {
			return nil, err
		}
	}

	if h, ok := in["on_failure"].([]interface{}); ok && len(h) > 0 {
		hooks.OnFailure, err = expandEaasHooks(h)
		if err != nil {
			return nil, err
		}
	}

	if h, ok := in["on_init"].([]interface{}); ok && len(h) > 0 {
		hooks.OnInit, err = expandEaasHooks(h)
		if err != nil {
			return nil, err
		}
	}

	if h, ok := in["provider"].([]interface{}); ok && len(h) > 0 {
		hooks.Provider, err = expandProviderHooks(h)
		if err != nil {
			return nil, err
		}
	}

	return hooks, nil

}

func expandCustomProviderOptions(p []interface{}) (*eaaspb.CustomProviderOptions, error) {
	if len(p) == 0 || p[0] == nil {
		return nil, nil
	}
	wfProviderOptions := &eaaspb.CustomProviderOptions{}
	in := p[0].(map[string]interface{})

	var err error
	if h, ok := in["tasks"].([]interface{}); ok && len(h) > 0 {
		wfProviderOptions.Tasks, err = expandEaasHooks(h)
		if err != nil {
			return nil, err
		}
	}
	wfProviderOptions.ReverseOnDestroy = in["reverse_on_destroy"].(bool)
	return wfProviderOptions, nil

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

	if bt, ok := in["backend_type"].(string); ok {
		tpo.BackendType = bt
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

	if h, ok := in["with_terraform_cloud"].([]interface{}); ok && len(h) > 0 {
		tpo.WithTerraformCloud = expandBoolValue(h)
	}

	if v, ok := in["volumes"].([]interface{}); ok && len(v) > 0 {
		tpo.Volumes = expandProviderVolumeOptions(v)
	}

	if h, ok := in["timeout_seconds"].(int); ok {
		tpo.TimeoutSeconds = int64(h)
	}

	return tpo
}

func expandOpenTofuProviderOptions(p []interface{}) *eaaspb.OpenTofuProviderOptions {
	if len(p) == 0 || p[0] == nil {
		return nil
	}
	tpo := &eaaspb.OpenTofuProviderOptions{}

	in := p[0].(map[string]interface{})

	if h, ok := in["version"].(string); ok {
		tpo.Version = h
	}

	if vfiles, ok := in["var_files"].([]interface{}); ok && len(vfiles) > 0 {
		tpo.VarFiles = toArrayString(vfiles)
	}

	if bcfgs, ok := in["backend_configs"].([]interface{}); ok && len(bcfgs) > 0 {
		tpo.BackendConfigs = toArrayString(bcfgs)
	}

	if bt, ok := in["backend_type"].(string); ok {
		tpo.BackendType = bt
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

	if v, ok := in["volumes"].([]interface{}); ok && len(v) > 0 {
		tpo.Volumes = expandProviderVolumeOptions(v)
	}

	if h, ok := in["timeout_seconds"].(int); ok {
		tpo.TimeoutSeconds = int64(h)
	}

	return tpo
}

func expandHcpTerraformProviderOptions(p []interface{}) *eaaspb.HCPTerraformProviderOptions {
	hcpTFOpts := &eaaspb.HCPTerraformProviderOptions{}
	if len(p) == 0 || p[0] == nil {
		return hcpTFOpts
	}
	in := p[0].(map[string]interface{})

	if vfiles, ok := in["var_files"].([]interface{}); ok && len(vfiles) > 0 {
		hcpTFOpts.VarFiles = toArrayString(vfiles)
	}

	if h, ok := in["refresh"].([]interface{}); ok && len(h) > 0 {
		hcpTFOpts.Refresh = expandBoolValue(h)
	}

	if h, ok := in["lock"].([]interface{}); ok && len(h) > 0 {
		hcpTFOpts.Lock = expandBoolValue(h)
	}

	if h, ok := in["lock_timeout_seconds"].(int); ok {
		hcpTFOpts.LockTimeoutSeconds = uint64(h)
	}

	if pdirs, ok := in["plugin_dirs"].([]interface{}); ok && len(pdirs) > 0 {
		hcpTFOpts.PluginDirs = toArrayString(pdirs)
	}

	if tgtrs, ok := in["target_resources"].([]interface{}); ok && len(tgtrs) > 0 {
		hcpTFOpts.TargetResources = toArrayString(tgtrs)
	}

	if v, ok := in["volumes"].([]interface{}); ok && len(v) > 0 {
		hcpTFOpts.Volumes = expandProviderVolumeOptions(v)
	}

	if h, ok := in["timeout_seconds"].(int); ok {
		hcpTFOpts.TimeoutSeconds = int64(h)
	}

	return hcpTFOpts

}

func expandSystemProviderOptions(p []interface{}) *eaaspb.SystemProviderOptions {
	spo := &eaaspb.SystemProviderOptions{}
	if len(p) == 0 || p[0] == nil {
		return spo
	}
	in := p[0].(map[string]interface{})

	if h, ok := in["kind"].(string); ok {
		spo.Kind = h
	}

	return spo
}

func expandTerragruntProviderOptions(p []interface{}) *eaaspb.TerragruntProviderOptions {
	if len(p) == 0 || p[0] == nil {
		return nil
	}
	tpo := &eaaspb.TerragruntProviderOptions{}
	return tpo
}

func expandPulumiProviderOptions(p []interface{}) *eaaspb.PulumiProviderOptions {
	if len(p) == 0 || p[0] == nil {
		return nil
	}
	ppo := &eaaspb.PulumiProviderOptions{}
	return ppo
}

func expandProviderHooks(p []interface{}) (*eaaspb.ResourceTemplateProviderHooks, error) {
	rtph := &eaaspb.ResourceTemplateProviderHooks{}
	if len(p) == 0 || p[0] == nil {
		return rtph, nil
	}

	in := p[0].(map[string]interface{})

	var err error
	if h, ok := in["terraform"].([]interface{}); ok && len(h) > 0 {
		rtph.Terraform, err = expandTerraformProviderHooks(h)
		if err != nil {
			return nil, err
		}
	}

	if h, ok := in["pulumi"].([]interface{}); ok && len(h) > 0 {
		rtph.Pulumi, err = expandPulumiProviderHooks(h)
		if err != nil {
			return nil, err
		}
	}

	if h, ok := in["open_tofu"].([]interface{}); ok && len(h) > 0 {
		rtph.OpenTofu, err = expandOpenTofuProviderHooks(h)
		if err != nil {
			return nil, err
		}
	}

	if h, oj := in["hcp_terraform"].([]interface{}); oj && len(h) > 0 {
		rtph.HcpTerraform, err = expandHcpTerraformProviderHooks(h)
		if err != nil {
			return nil, err
		}
	}

	if h, oj := in["system"].([]interface{}); oj && len(h) > 0 {
		rtph.System, err = expandSystemProviderHooks(h)
		if err != nil {
			return nil, err
		}
	}

	return rtph, nil

}

func expandTerraformProviderHooks(p []interface{}) (*eaaspb.TerraformProviderHooks, error) {
	tph := &eaaspb.TerraformProviderHooks{}
	if len(p) == 0 || p[0] == nil {
		return tph, nil
	}

	in := p[0].(map[string]interface{})

	var err error
	if h, ok := in["deploy"].([]interface{}); ok && len(h) > 0 {
		tph.Deploy, err = expandTerraformDeployHooks(h)
		if err != nil {
			return nil, err
		}
	}

	if h, ok := in["destroy"].([]interface{}); ok && len(h) > 0 {
		tph.Destroy, err = expandTerraformDestroyHooks(h)
		if err != nil {
			return nil, err
		}
	}

	return tph, nil
}

func expandPulumiProviderHooks(p []interface{}) (*eaaspb.PulumiProviderHooks, error) {
	pph := &eaaspb.PulumiProviderHooks{}
	if len(p) == 0 || p[0] == nil {
		return pph, nil
	}

	in := p[0].(map[string]interface{})

	var err error
	if h, ok := in["deploy"].([]interface{}); ok {
		pph.Deploy, err = expandPulumiDeployHooks(h)
		if err != nil {
			return nil, err
		}
	}

	if h, ok := in["destroy"].([]interface{}); ok {
		pph.Destroy, err = expandPulumiDestroyHooks(h)
		if err != nil {
			return nil, err
		}
	}

	return pph, nil
}

func expandOpenTofuProviderHooks(p []interface{}) (*eaaspb.OpenTofuProviderHooks, error) {
	tph := &eaaspb.OpenTofuProviderHooks{}
	if len(p) == 0 || p[0] == nil {
		return tph, nil
	}

	in := p[0].(map[string]interface{})

	var err error
	if h, ok := in["deploy"].([]interface{}); ok && len(h) > 0 {
		tph.Deploy, err = expandOpenTofuDeployHooks(h)
		if err != nil {
			return nil, err
		}
	}

	if h, ok := in["destroy"].([]interface{}); ok && len(h) > 0 {
		tph.Destroy, err = expandOpenTofuDestroyHooks(h)
		if err != nil {
			return nil, err
		}
	}

	return tph, nil
}

func expandHcpTerraformProviderHooks(p []interface{}) (*eaaspb.HCPTerraformProviderHooks, error) {
	tph := &eaaspb.HCPTerraformProviderHooks{}
	if len(p) == 0 || p[0] == nil {
		return tph, nil
	}

	in := p[0].(map[string]interface{})

	var err error
	if h, ok := in["deploy"].([]interface{}); ok && len(h) > 0 {
		tph.Deploy, err = expandHcpTerraformDeployHooks(h)
		if err != nil {
			return nil, err
		}
	}

	if h, ok := in["destroy"].([]interface{}); ok && len(h) > 0 {
		tph.Destroy, err = expandHcpTerraformDestroyHooks(h)
		if err != nil {
			return nil, err
		}
	}

	return tph, nil
}
func expandSystemProviderHooks(p []interface{}) (*eaaspb.SystemProviderHooks, error) {
	sph := &eaaspb.SystemProviderHooks{}
	if len(p) == 0 || p[0] == nil {
		return sph, nil
	}

	in := p[0].(map[string]interface{})

	var err error
	if h, ok := in["deploy"].([]interface{}); ok && len(h) > 0 {
		sph.Deploy, err = expandSystemDeployHooks(h)
		if err != nil {
			return nil, err
		}
	}

	if h, ok := in["destroy"].([]interface{}); ok && len(h) > 0 {
		sph.Destroy, err = expandSystemDestroyHooks(h)
		if err != nil {
			return nil, err
		}
	}

	return sph, nil
}

func expandTerraformDeployHooks(p []interface{}) (*eaaspb.TerraformDeployHooks, error) {
	tdh := &eaaspb.TerraformDeployHooks{}
	if len(p) == 0 || p[0] == nil {
		return tdh, nil
	}

	in := p[0].(map[string]interface{})

	var err error
	if h, ok := in["init"].([]interface{}); ok && len(h) > 0 {
		tdh.Init, err = expandLifecycleEventHooks(h)
		if err != nil {
			return nil, err
		}
	}

	if h, ok := in["plan"].([]interface{}); ok && len(h) > 0 {
		tdh.Plan, err = expandLifecycleEventHooks(h)
		if err != nil {
			return nil, err
		}
	}

	if h, ok := in["apply"].([]interface{}); ok && len(h) > 0 {
		tdh.Apply, err = expandLifecycleEventHooks(h)
		if err != nil {
			return nil, err
		}
	}

	if h, ok := in["output"].([]interface{}); ok && len(h) > 0 {
		tdh.Output, err = expandLifecycleEventHooks(h)
		if err != nil {
			return nil, err
		}
	}

	return tdh, nil
}

func expandTerraformDestroyHooks(p []interface{}) (*eaaspb.TerraformDestroyHooks, error) {
	tdh := &eaaspb.TerraformDestroyHooks{}
	if len(p) == 0 || p[0] == nil {
		return tdh, nil
	}

	in := p[0].(map[string]interface{})

	var err error
	if h, ok := in["init"].([]interface{}); ok && len(h) > 0 {
		tdh.Init, err = expandLifecycleEventHooks(h)
		if err != nil {
			return nil, err
		}
	}

	if h, ok := in["plan"].([]interface{}); ok && len(h) > 0 {
		tdh.Plan, err = expandLifecycleEventHooks(h)
		if err != nil {
			return nil, err
		}
	}

	if h, ok := in["destroy"].([]interface{}); ok && len(h) > 0 {
		tdh.Destroy, err = expandLifecycleEventHooks(h)
		if err != nil {
			return nil, err
		}
	}

	return tdh, nil
}

func expandOpenTofuDeployHooks(p []interface{}) (*eaaspb.OpenTofuDeployHooks, error) {
	tdh := &eaaspb.OpenTofuDeployHooks{}
	if len(p) == 0 || p[0] == nil {
		return tdh, nil
	}

	in := p[0].(map[string]interface{})

	var err error
	if h, ok := in["init"].([]interface{}); ok && len(h) > 0 {
		tdh.Init, err = expandLifecycleEventHooks(h)
		if err != nil {
			return nil, err
		}
	}

	if h, ok := in["plan"].([]interface{}); ok && len(h) > 0 {
		tdh.Plan, err = expandLifecycleEventHooks(h)
		if err != nil {
			return nil, err
		}
	}

	if h, ok := in["apply"].([]interface{}); ok && len(h) > 0 {
		tdh.Apply, err = expandLifecycleEventHooks(h)
		if err != nil {
			return nil, err
		}
	}

	if h, ok := in["output"].([]interface{}); ok && len(h) > 0 {
		tdh.Output, err = expandLifecycleEventHooks(h)
		if err != nil {
			return nil, err
		}
	}

	return tdh, nil
}

func expandOpenTofuDestroyHooks(p []interface{}) (*eaaspb.OpenTofuDestroyHooks, error) {
	tdh := &eaaspb.OpenTofuDestroyHooks{}
	if len(p) == 0 || p[0] == nil {
		return tdh, nil
	}

	in := p[0].(map[string]interface{})

	var err error
	if h, ok := in["init"].([]interface{}); ok && len(h) > 0 {
		tdh.Init, err = expandLifecycleEventHooks(h)
		if err != nil {
			return nil, err
		}
	}

	if h, ok := in["plan"].([]interface{}); ok && len(h) > 0 {
		tdh.Plan, err = expandLifecycleEventHooks(h)
		if err != nil {
			return nil, err
		}
	}

	if h, ok := in["destroy"].([]interface{}); ok && len(h) > 0 {
		tdh.Destroy, err = expandLifecycleEventHooks(h)
		if err != nil {
			return nil, err
		}
	}

	return tdh, nil
}

func expandHcpTerraformDeployHooks(p []interface{}) (*eaaspb.HCPTerraformDeployHooks, error) {
	tdh := &eaaspb.HCPTerraformDeployHooks{}
	if len(p) == 0 || p[0] == nil {
		return tdh, nil
	}

	in := p[0].(map[string]interface{})

	var err error
	if h, ok := in["init"].([]interface{}); ok && len(h) > 0 {
		tdh.Init, err = expandLifecycleEventHooks(h)
		if err != nil {
			return nil, err
		}
	}

	if h, ok := in["plan"].([]interface{}); ok && len(h) > 0 {
		tdh.Plan, err = expandLifecycleEventHooks(h)
		if err != nil {
			return nil, err
		}
	}

	if h, ok := in["apply"].([]interface{}); ok && len(h) > 0 {
		tdh.Apply, err = expandLifecycleEventHooks(h)
		if err != nil {
			return nil, err
		}
	}

	if h, ok := in["output"].([]interface{}); ok && len(h) > 0 {
		tdh.Output, err = expandLifecycleEventHooks(h)
		if err != nil {
			return nil, err
		}
	}

	return tdh, nil
}

func expandHcpTerraformDestroyHooks(p []interface{}) (*eaaspb.HCPTerraformDestroyHooks, error) {
	tdh := &eaaspb.HCPTerraformDestroyHooks{}
	if len(p) == 0 || p[0] == nil {
		return tdh, nil
	}

	in := p[0].(map[string]interface{})

	var err error
	if h, ok := in["init"].([]interface{}); ok && len(h) > 0 {
		tdh.Init, err = expandLifecycleEventHooks(h)
		if err != nil {
			return nil, err
		}
	}

	if h, ok := in["plan"].([]interface{}); ok && len(h) > 0 {
		tdh.Plan, err = expandLifecycleEventHooks(h)
		if err != nil {
			return nil, err
		}
	}

	if h, ok := in["destroy"].([]interface{}); ok && len(h) > 0 {
		tdh.Destroy, err = expandLifecycleEventHooks(h)
		if err != nil {
			return nil, err
		}
	}

	return tdh, nil
}

func expandSystemDeployHooks(p []interface{}) (*eaaspb.SystemDeployHooks, error) {
	sdh := &eaaspb.SystemDeployHooks{}
	if len(p) == 0 || p[0] == nil {
		return sdh, nil
	}

	in := p[0].(map[string]interface{})

	var err error
	if h, ok := in["apply"].([]interface{}); ok && len(h) > 0 {
		sdh.Apply, err = expandLifecycleEventHooks(h)
		if err != nil {
			return nil, err
		}
	}

	return sdh, nil
}

func expandSystemDestroyHooks(p []interface{}) (*eaaspb.SystemDestroyHooks, error) {
	sdh := &eaaspb.SystemDestroyHooks{}
	if len(p) == 0 || p[0] == nil {
		return sdh, nil
	}

	in := p[0].(map[string]interface{})

	var err error
	if h, ok := in["destroy"].([]interface{}); ok && len(h) > 0 {
		sdh.Destroy, err = expandLifecycleEventHooks(h)
		if err != nil {
			return nil, err
		}
	}

	return sdh, nil
}

func expandPulumiDeployHooks(p []interface{}) (*eaaspb.PulumiDeployHooks, error) {
	pdh := &eaaspb.PulumiDeployHooks{}
	if len(p) == 0 || p[0] == nil {
		return pdh, nil
	}

	in := p[0].(map[string]interface{})

	var err error
	if h, ok := in["login"].([]interface{}); ok && len(h) > 0 {
		pdh.Login, err = expandLifecycleEventHooks(h)
		if err != nil {
			return nil, err
		}
	}

	if h, ok := in["preview"].([]interface{}); ok && len(h) > 0 {
		pdh.Preview, err = expandLifecycleEventHooks(h)
		if err != nil {
			return nil, err
		}
	}

	if h, ok := in["up"].([]interface{}); ok && len(h) > 0 {
		pdh.Up, err = expandLifecycleEventHooks(h)
		if err != nil {
			return nil, err
		}
	}

	if h, ok := in["output"].([]interface{}); ok && len(h) > 0 {
		pdh.Output, err = expandLifecycleEventHooks(h)
		if err != nil {
			return nil, err
		}
	}

	return pdh, nil
}

func expandPulumiDestroyHooks(p []interface{}) (*eaaspb.PulumiDestroyHooks, error) {
	pdh := &eaaspb.PulumiDestroyHooks{}
	if len(p) == 0 || p[0] == nil {
		return pdh, nil
	}

	in := p[0].(map[string]interface{})

	var err error
	if h, ok := in["login"].([]interface{}); ok && len(h) > 0 {
		pdh.Login, err = expandLifecycleEventHooks(h)
		if err != nil {
			return nil, err
		}
	}

	if h, ok := in["preview"].([]interface{}); ok && len(h) > 0 {
		pdh.Preview, err = expandLifecycleEventHooks(h)
		if err != nil {
			return nil, err
		}
	}

	if h, ok := in["destroy"].([]interface{}); ok && len(h) > 0 {
		pdh.Destroy, err = expandLifecycleEventHooks(h)
		if err != nil {
			return nil, err
		}
	}

	return pdh, nil
}

func expandLifecycleEventHooks(p []interface{}) (*eaaspb.LifecycleEventHooks, error) {
	lh := &eaaspb.LifecycleEventHooks{}
	if len(p) == 0 || p[0] == nil {
		return lh, nil
	}

	in := p[0].(map[string]interface{})

	var err error
	if h, ok := in["before"].([]interface{}); ok && len(h) > 0 {
		lh.Before, err = expandEaasHooks(h)
		if err != nil {
			return nil, err
		}
	}

	if h, ok := in["after"].([]interface{}); ok && len(h) > 0 {
		lh.After, err = expandEaasHooks(h)
		if err != nil {
			return nil, err
		}
	}

	return lh, nil
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
		return nil, nil
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	obj["version"] = in.Version
	obj["version_state"] = in.VersionState
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
	obj["artifact_workflow_handler"] = flattenWorkflowHandlerCompoundRef(in.ArtifactWorkflowHandler)

	if len(in.Actions) > 0 {
		v, ok := obj["actions"].([]interface{})
		if !ok {
			v = []interface{}{}
		}

		obj["actions"] = flattenActions(in.Actions, v)
	}

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
	obj["workflow_handler"] = flattenWorkflowHandlerCompoundRef(in.WorkflowHandler)
	obj["open_tofu"] = flattenOpenTofuProviderOptions(in.OpenTofu)
	obj["custom"] = flattenCustomProviderOptions(in.Custom)
	obj["hcp_terraform"] = flattenHcpTerraformProviderOptions(in.HcpTerraform)

	return []interface{}{obj}
}

func flattenOpenTofuProviderOptions(in *eaaspb.OpenTofuProviderOptions) []interface{} {
	if in == nil {
		return nil
	}

	obj := make(map[string]interface{})
	obj["version"] = in.Version
	// obj["use_system_state_store"] = flattenBoolValue(in.UseSystemStateStore)
	obj["var_files"] = toArrayInterface(in.VarFiles)
	obj["backend_configs"] = toArrayInterface(in.BackendConfigs)
	obj["backend_type"] = in.BackendType
	obj["refresh"] = flattenBoolValue(in.Refresh)
	obj["lock"] = flattenBoolValue(in.Lock)
	obj["lock_timeout_seconds"] = in.LockTimeoutSeconds
	obj["plugin_dirs"] = toArrayInterface(in.PluginDirs)
	obj["target_resources"] = toArrayInterface(in.TargetResources)
	// obj["with_terraform_cloud"] = flattenBoolValue(in.WithTerraformCloud)
	if len(in.Volumes) > 0 {
		v, ok := obj["volumes"].([]interface{})
		if !ok {
			v = []interface{}{}
		}

		obj["volumes"] = flattenProviderVolumeOptions(in.Volumes, v)
	}
	obj["timeout_seconds"] = in.TimeoutSeconds

	return []interface{}{obj}
}

func flattenCustomProviderOptions(in *eaaspb.CustomProviderOptions) []interface{} {
	if in == nil {
		return nil
	}
	obj := make(map[string]interface{})
	if len(in.Tasks) > 0 {
		v, ok := obj["tasks"].([]interface{})
		if !ok {
			v = []interface{}{}
		}

		obj["tasks"] = flattenEaasHooks(in.Tasks, v)
		obj["reverse_on_destroy"] = in.ReverseOnDestroy
	}

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
	obj["backend_type"] = in.BackendType
	obj["refresh"] = flattenBoolValue(in.Refresh)
	obj["lock"] = flattenBoolValue(in.Lock)
	obj["lock_timeout_seconds"] = in.LockTimeoutSeconds
	obj["plugin_dirs"] = toArrayInterface(in.PluginDirs)
	obj["target_resources"] = toArrayInterface(in.TargetResources)
	obj["with_terraform_cloud"] = flattenBoolValue(in.WithTerraformCloud)
	if len(in.Volumes) > 0 {
		v, ok := obj["volumes"].([]interface{})
		if !ok {
			v = []interface{}{}
		}

		obj["volumes"] = flattenProviderVolumeOptions(in.Volumes, v)
	}
	obj["timeout_seconds"] = in.TimeoutSeconds

	return []interface{}{obj}
}

func flattenHcpTerraformProviderOptions(in *eaaspb.HCPTerraformProviderOptions) []interface{} {
	if in == nil {
		return nil
	}

	obj := make(map[string]interface{})
	obj["var_files"] = toArrayInterface(in.VarFiles)
	obj["refresh"] = flattenBoolValue(in.Refresh)
	obj["lock"] = flattenBoolValue(in.Lock)
	obj["lock_timeout_seconds"] = in.LockTimeoutSeconds
	obj["plugin_dirs"] = toArrayInterface(in.PluginDirs)
	obj["target_resources"] = toArrayInterface(in.TargetResources)
	if len(in.Volumes) > 0 {
		v, ok := obj["volumes"].([]interface{})
		if !ok {
			v = []interface{}{}
		}

		obj["volumes"] = flattenProviderVolumeOptions(in.Volumes, v)
	}
	obj["timeout_seconds"] = in.TimeoutSeconds

	return []interface{}{obj}
}

func flattenSystemProviderOptions(in *eaaspb.SystemProviderOptions) []interface{} {
	if in == nil {
		return nil
	}

	obj := make(map[string]interface{})
	obj["kind"] = in.Kind
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

func flattenContexts(input []*eaaspb.ConfigContextCompoundRef, p []interface{}) []interface{} {
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

	if input.OpenTofu != nil {
		v, ok := obj["open_tofu"].([]interface{})
		if !ok {
			v = []interface{}{}
		}

		obj["open_tofu"] = flattenOpenTofuProviderHooks(input.OpenTofu, v)
	}

	if input.Pulumi != nil {
		v, ok := obj["pulumi"].([]interface{})
		if !ok {
			v = []interface{}{}
		}

		obj["pulumi"] = flattenPulumiProviderHooks(input.Pulumi, v)
	}

	if input.HcpTerraform != nil {
		v, ok := obj["hcp_terraform"].([]interface{})
		if !ok {
			v = []interface{}{}
		}

		obj["hcp_terraform"] = flattenHcpTerraformProviderHooks(input.HcpTerraform, v)
	}

	if input.System != nil {
		v, ok := obj["system"].([]interface{})
		if !ok {
			v = []interface{}{}
		}

		obj["system"] = flattenSystemProviderHooks(input.System, v)
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

func flattenOpenTofuProviderHooks(input *eaaspb.OpenTofuProviderHooks, p []interface{}) []interface{} {
	log.Println("flatten opentofu provider hooks start")
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

		obj["deploy"] = flattenOpenTofuDeployHooks(input.Deploy, v)
	}

	if input.Destroy != nil {
		v, ok := obj["destroy"].([]interface{})
		if !ok {
			v = []interface{}{}
		}

		obj["destroy"] = flattenOpenTofuDestroyHooks(input.Destroy, v)
	}

	return []interface{}{obj}
}

func flattenHcpTerraformProviderHooks(input *eaaspb.HCPTerraformProviderHooks, p []interface{}) []interface{} {
	log.Println("flatten hcp terraform provider hooks start")
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

		obj["deploy"] = flattenHcpTerraformDeployHooks(input.Deploy, v)
	}

	if input.Destroy != nil {
		v, ok := obj["destroy"].([]interface{})
		if !ok {
			v = []interface{}{}
		}

		obj["destroy"] = flattenHcpTerraformDestroyHooks(input.Destroy, v)
	}

	return []interface{}{obj}
}

func flattenSystemProviderHooks(input *eaaspb.SystemProviderHooks, p []interface{}) []interface{} {
	log.Println("flatten system provider hooks start")
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

		obj["deploy"] = flattenSystemDeployHooks(input.Deploy, v)
	}

	if input.Destroy != nil {
		v, ok := obj["destroy"].([]interface{})
		if !ok {
			v = []interface{}{}
		}

		obj["destroy"] = flattenSystemDestroyHooks(input.Destroy, v)
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

func flattenOpenTofuDeployHooks(input *eaaspb.OpenTofuDeployHooks, p []interface{}) []interface{} {
	log.Println("flatten opentofu deploy hooks start")
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

func flattenOpenTofuDestroyHooks(input *eaaspb.OpenTofuDestroyHooks, p []interface{}) []interface{} {
	log.Println("flatten opentofu destroy hooks start")
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

func flattenHcpTerraformDeployHooks(input *eaaspb.HCPTerraformDeployHooks, p []interface{}) []interface{} {
	log.Println("flatten hcp terraform deploy hooks start")
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

func flattenHcpTerraformDestroyHooks(input *eaaspb.HCPTerraformDestroyHooks, p []interface{}) []interface{} {
	log.Println("flatten hcp terraform destroy hooks start")
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

func flattenSystemDeployHooks(input *eaaspb.SystemDeployHooks, p []interface{}) []interface{} {
	log.Println("flatten system deploy hooks start")
	if input == nil {
		return nil
	}

	obj := make(map[string]interface{})
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if input.Apply != nil {
		v, ok := obj["apply"].([]interface{})
		if !ok {
			v = []interface{}{}
		}

		obj["apply"] = flattenLifecycleEventHooks(input.Apply, v)
	}

	return []interface{}{obj}
}

func flattenSystemDestroyHooks(input *eaaspb.SystemDestroyHooks, p []interface{}) []interface{} {
	log.Println("flatten system destroy hooks start")
	if input == nil {
		return nil
	}

	obj := make(map[string]interface{})
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
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

func expandProviderVolumeOptions(p []interface{}) []*eaaspb.ProviderVolumeOptions {
	volumes := make([]*eaaspb.ProviderVolumeOptions, 0)
	if len(p) == 0 {
		return volumes
	}

	for indx := range p {
		volume := &eaaspb.ProviderVolumeOptions{}
		if p[indx] == nil {
			return volumes
		}
		in := p[indx].(map[string]interface{})

		if mp, ok := in["mount_path"].(string); ok && len(mp) > 0 {
			volume.MountPath = mp
		}

		if pvcsz, ok := in["pvc_size_gb"].(string); ok && len(pvcsz) > 0 {
			volume.PvcSizeGB = pvcsz
		}

		if pvcsc, ok := in["pvc_storage_class"].(string); ok && len(pvcsc) > 0 {
			volume.PvcStorageClass = pvcsc
		}

		if usepvc, ok := in["use_pvc"].([]interface{}); ok && len(usepvc) > 0 {
			volume.UsePVC = expandBoolValue(usepvc)
		}

		if enableBackupAndRestore, ok := in["enable_backup_and_restore"].(bool); ok {
			volume.EnableBackupAndRestore = enableBackupAndRestore
		}

		volumes = append(volumes, volume)

	}

	return volumes
}

func flattenProviderVolumeOptions(input []*eaaspb.ProviderVolumeOptions, p []interface{}) []interface{} {
	if len(input) == 0 {
		return nil
	}

	out := make([]interface{}, len(input))
	for i, in := range input {
		log.Println("flatten provider volume options", in)
		obj := map[string]interface{}{}
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}
		obj["use_pvc"] = flattenBoolValue(in.UsePVC)

		if len(in.MountPath) > 0 {
			obj["mount_path"] = in.MountPath
		}

		if len(in.PvcSizeGB) > 0 {
			obj["pvc_size_gb"] = in.PvcSizeGB
		}

		if len(in.PvcStorageClass) > 0 {
			obj["pvc_storage_class"] = in.PvcStorageClass
		}

		obj["enable_backup_and_restore"] = in.EnableBackupAndRestore

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
