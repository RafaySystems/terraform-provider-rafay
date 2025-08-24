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

func resourceTemplateCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	log.Println("resource template create")
	diags := resourceTemplateUpsert(ctx, d, m)
	return diags
}

func resourceTemplateUpsert(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
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

func resourceTemplateRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
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

	if rt.GetSpec().GetSharing() != nil && !rt.GetSpec().GetSharing().GetEnabled() && resourcetemplate.GetSpec().GetSharing() == nil {
		resourcetemplate.Spec.Sharing = &commonpb.SharingSpec{}
		resourcetemplate.Spec.Sharing.Enabled = false
		resourcetemplate.Spec.Sharing.Projects = rt.GetSpec().GetSharing().GetProjects()
	}

	err = flattenResourceTemplate(d, resourcetemplate)
	if err != nil {
		log.Println("read flatten err")
		return diag.FromErr(err)
	}
	return diags

}

func resourceTemplateUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	return resourceTemplateUpsert(ctx, d, m)
}

func resourceTemplateDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
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

	if v, ok := in.Get("metadata").([]any); ok && len(v) > 0 {
		obj.Metadata = expandV3MetaData(v)
	}

	if v, ok := in.Get("spec").([]any); ok && len(v) > 0 {
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

func expandResourceTemplateSpec(p []any) (*eaaspb.ResourceTemplateSpec, error) {
	log.Println("expand resource template spec")
	spec := &eaaspb.ResourceTemplateSpec{}
	if len(p) == 0 || p[0] == nil {
		return spec, fmt.Errorf("%s", "expand resource template spec empty input")
	}

	in := p[0].(map[string]any)

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
	if po, ok := in["provider_options"].([]any); ok && len(po) > 0 {
		spec.ProviderOptions, err = expandProviderOptions(po)
		if err != nil {
			return nil, err
		}
	}

	if ro, ok := in["repository_options"].([]any); ok && len(ro) > 0 {
		spec.RepositoryOptions = expandResourceTemplateRepositoryOptions(ro)
	}

	if v, ok := in["contexts"].([]any); ok && len(v) > 0 {
		spec.Contexts = expandContexts(v)
	}

	if v, ok := in["variables"].([]any); ok && len(v) > 0 {
		spec.Variables = expandVariables(v)
	}

	if h, ok := in["hooks"].([]any); ok && len(h) > 0 {
		spec.Hooks, err = expandResourceHooks(h)
		if err != nil {
			return nil, err
		}
	}

	if ag, ok := in["agents"].([]any); ok && len(ag) > 0 {
		spec.Agents = expandEaasAgents(ag)
	}

	if ap, ok := in["agent_pools"].([]any); ok && len(ap) > 0 {
		spec.AgentPools = expandEaasAgents(ap)
	}

	if v, ok := in["sharing"].([]any); ok && len(v) > 0 {
		spec.Sharing = expandSharingSpec(v)
	}

	if ad, ok := in["artifact_driver"].([]any); ok && len(ad) > 0 {
		log.Println("WARN: artifact_driver is deprecated, use artifact_workflow_handler instead")
		spec.ArtifactDriver, err = expandWorkflowHandlerCompoundRef(ad)
		if err != nil {
			return nil, err
		}
	}

	if ad, ok := in["artifact_workflow_handler"].([]any); ok && len(ad) > 0 {
		spec.ArtifactWorkflowHandler, err = expandWorkflowHandlerCompoundRef(ad)
		if err != nil {
			return nil, err
		}
	}

	if s, ok := in["actions"].([]any); ok && len(s) > 0 {
		spec.Actions, err = expandActions(s)
		if err != nil {
			return nil, err
		}
	}

	return spec, nil
}

func expandProviderOptions(p []any) (*eaaspb.ResourceTemplateProviderOptions, error) {
	po := &eaaspb.ResourceTemplateProviderOptions{}
	if len(p) == 0 || p[0] == nil {
		return po, nil
	}
	in := p[0].(map[string]any)

	var err error
	if tp, ok := in["terraform"].([]any); ok && len(tp) > 0 {
		po.Terraform = expandTerraformProviderOptions(tp)
	}

	if s, ok := in["system"].([]any); ok && len(s) > 0 {
		po.System = expandSystemProviderOptions(s)
	}

	if t, ok := in["terragrunt"].([]any); ok && len(t) > 0 {
		po.Terragrunt = expandTerragruntProviderOptions(t)
	}

	if p, ok := in["pulumi"].([]any); ok && len(p) > 0 {
		po.Pulumi = expandPulumiProviderOptions(p)
	}

	if p, ok := in["driver"].([]any); ok && len(p) > 0 {
		log.Println("WARN: spec.provider_options.driver is deprecated, use spec.provider_options.workflow_handler instead")
		po.Driver, err = expandWorkflowHandlerCompoundRef(p)
		if err != nil {
			return nil, err
		}
	}

	if p, ok := in["workflow_handler"].([]any); ok && len(p) > 0 {
		po.WorkflowHandler, err = expandWorkflowHandlerCompoundRef(p)
		if err != nil {
			return nil, err
		}
	}

	if p, ok := in["open_tofu"].([]any); ok && len(p) > 0 {
		po.OpenTofu = expandOpenTofuProviderOptions(p)
	}

	if w, ok := in["custom"].([]any); ok && len(p) > 0 {
		po.Custom, err = expandCustomProviderOptions(w)
		if err != nil {
			return nil, err
		}
	}

	if p, ok := in["hcp_terraform"].([]any); ok && len(p) > 0 {
		po.HcpTerraform = expandHcpTerraformProviderOptions(p)
	}

	return po, nil

}

func expandResourceTemplateRepositoryOptions(p []any) *eaaspb.ResourceTemplateRepositoryOptions {
	ro := &eaaspb.ResourceTemplateRepositoryOptions{}
	if len(p) == 0 || p[0] == nil {
		return ro
	}
	in := p[0].(map[string]any)

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

func expandContexts(p []any) []*eaaspb.ConfigContextCompoundRef {
	ctxs := make([]*eaaspb.ConfigContextCompoundRef, 0)
	if len(p) == 0 {
		return ctxs
	}

	for indx := range p {
		obj := &eaaspb.ConfigContextCompoundRef{}

		in := p[indx].(map[string]any)

		if v, ok := in["name"].(string); ok && len(v) > 0 {
			obj.Name = v
		}

		if v, ok := in["data"].([]any); ok && len(v) > 0 {
			obj.Data = expandConfigContextInline(v)
		}

		ctxs = append(ctxs, obj)
	}

	return ctxs
}

func expandResourceHooks(p []any) (*eaaspb.ResourceHooks, error) {
	hooks := &eaaspb.ResourceHooks{}

	if len(p) == 0 || p[0] == nil {
		return hooks, nil
	}

	in := p[0].(map[string]any)

	var err error
	if h, ok := in["on_completion"].([]any); ok && len(h) > 0 {
		hooks.OnCompletion, err = expandEaasHooks(h)
		if err != nil {
			return nil, err
		}
	}

	if h, ok := in["on_success"].([]any); ok && len(h) > 0 {
		hooks.OnSuccess, err = expandEaasHooks(h)
		if err != nil {
			return nil, err
		}
	}

	if h, ok := in["on_failure"].([]any); ok && len(h) > 0 {
		hooks.OnFailure, err = expandEaasHooks(h)
		if err != nil {
			return nil, err
		}
	}

	if h, ok := in["on_init"].([]any); ok && len(h) > 0 {
		hooks.OnInit, err = expandEaasHooks(h)
		if err != nil {
			return nil, err
		}
	}

	if h, ok := in["provider"].([]any); ok && len(h) > 0 {
		hooks.Provider, err = expandProviderHooks(h)
		if err != nil {
			return nil, err
		}
	}

	return hooks, nil

}

func expandCustomProviderOptions(p []any) (*eaaspb.CustomProviderOptions, error) {
	if len(p) == 0 || p[0] == nil {
		return nil, nil
	}
	wfProviderOptions := &eaaspb.CustomProviderOptions{}
	in := p[0].(map[string]any)

	var err error
	if h, ok := in["tasks"].([]any); ok && len(h) > 0 {
		wfProviderOptions.Tasks, err = expandEaasHooks(h)
		if err != nil {
			return nil, err
		}
	}
	wfProviderOptions.ReverseOnDestroy = in["reverse_on_destroy"].(bool)
	return wfProviderOptions, nil

}

func expandEaasAgents(p []any) []*commonpb.ResourceNameAndVersionRef {
	agents := make([]*commonpb.ResourceNameAndVersionRef, 0)
	if len(p) == 0 {
		return agents
	}

	for indx := range p {
		obj := &commonpb.ResourceNameAndVersionRef{}

		in := p[indx].(map[string]any)

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

func expandTerraformProviderOptions(p []any) *eaaspb.TerraformProviderOptions {
	tpo := &eaaspb.TerraformProviderOptions{}
	if len(p) == 0 || p[0] == nil {
		return tpo
	}

	in := p[0].(map[string]any)

	if h, ok := in["version"].(string); ok {
		tpo.Version = h
	}

	if h, ok := in["use_system_state_store"].([]any); ok && len(h) > 0 {
		tpo.UseSystemStateStore = expandBoolValue(h)
	}

	if vfiles, ok := in["var_files"].([]any); ok && len(vfiles) > 0 {
		tpo.VarFiles = toArrayString(vfiles)
	}

	if bcfgs, ok := in["backend_configs"].([]any); ok && len(bcfgs) > 0 {
		tpo.BackendConfigs = toArrayString(bcfgs)
	}

	if bt, ok := in["backend_type"].(string); ok {
		tpo.BackendType = bt
	}

	if h, ok := in["refresh"].([]any); ok && len(h) > 0 {
		tpo.Refresh = expandBoolValue(h)
	}

	if h, ok := in["lock"].([]any); ok && len(h) > 0 {
		tpo.Lock = expandBoolValue(h)
	}

	if h, ok := in["lock_timeout_seconds"].(int); ok {
		tpo.LockTimeoutSeconds = uint64(h)
	}

	if pdirs, ok := in["plugin_dirs"].([]any); ok && len(pdirs) > 0 {
		tpo.PluginDirs = toArrayString(pdirs)
	}

	if tgtrs, ok := in["target_resources"].([]any); ok && len(tgtrs) > 0 {
		tpo.TargetResources = toArrayString(tgtrs)
	}

	if h, ok := in["with_terraform_cloud"].([]any); ok && len(h) > 0 {
		tpo.WithTerraformCloud = expandBoolValue(h)
	}

	if v, ok := in["volumes"].([]any); ok && len(v) > 0 {
		tpo.Volumes = expandProviderVolumeOptions(v)
	}

	if h, ok := in["timeout_seconds"].(int); ok {
		tpo.TimeoutSeconds = int64(h)
	}

	return tpo
}

func expandOpenTofuProviderOptions(p []any) *eaaspb.OpenTofuProviderOptions {
	if len(p) == 0 || p[0] == nil {
		return nil
	}
	tpo := &eaaspb.OpenTofuProviderOptions{}

	in := p[0].(map[string]any)

	if h, ok := in["version"].(string); ok {
		tpo.Version = h
	}

	if vfiles, ok := in["var_files"].([]any); ok && len(vfiles) > 0 {
		tpo.VarFiles = toArrayString(vfiles)
	}

	if bcfgs, ok := in["backend_configs"].([]any); ok && len(bcfgs) > 0 {
		tpo.BackendConfigs = toArrayString(bcfgs)
	}

	if bt, ok := in["backend_type"].(string); ok {
		tpo.BackendType = bt
	}

	if h, ok := in["refresh"].([]any); ok && len(h) > 0 {
		tpo.Refresh = expandBoolValue(h)
	}

	if h, ok := in["lock"].([]any); ok && len(h) > 0 {
		tpo.Lock = expandBoolValue(h)
	}

	if h, ok := in["lock_timeout_seconds"].(int); ok {
		tpo.LockTimeoutSeconds = uint64(h)
	}

	if pdirs, ok := in["plugin_dirs"].([]any); ok && len(pdirs) > 0 {
		tpo.PluginDirs = toArrayString(pdirs)
	}

	if tgtrs, ok := in["target_resources"].([]any); ok && len(tgtrs) > 0 {
		tpo.TargetResources = toArrayString(tgtrs)
	}

	if v, ok := in["volumes"].([]any); ok && len(v) > 0 {
		tpo.Volumes = expandProviderVolumeOptions(v)
	}

	if h, ok := in["timeout_seconds"].(int); ok {
		tpo.TimeoutSeconds = int64(h)
	}

	return tpo
}

func expandHcpTerraformProviderOptions(p []any) *eaaspb.HCPTerraformProviderOptions {
	hcpTFOpts := &eaaspb.HCPTerraformProviderOptions{}
	if len(p) == 0 || p[0] == nil {
		return hcpTFOpts
	}
	in := p[0].(map[string]any)

	if vfiles, ok := in["var_files"].([]any); ok && len(vfiles) > 0 {
		hcpTFOpts.VarFiles = toArrayString(vfiles)
	}

	if h, ok := in["refresh"].([]any); ok && len(h) > 0 {
		hcpTFOpts.Refresh = expandBoolValue(h)
	}

	if h, ok := in["lock"].([]any); ok && len(h) > 0 {
		hcpTFOpts.Lock = expandBoolValue(h)
	}

	if h, ok := in["lock_timeout_seconds"].(int); ok {
		hcpTFOpts.LockTimeoutSeconds = uint64(h)
	}

	if pdirs, ok := in["plugin_dirs"].([]any); ok && len(pdirs) > 0 {
		hcpTFOpts.PluginDirs = toArrayString(pdirs)
	}

	if tgtrs, ok := in["target_resources"].([]any); ok && len(tgtrs) > 0 {
		hcpTFOpts.TargetResources = toArrayString(tgtrs)
	}

	if v, ok := in["volumes"].([]any); ok && len(v) > 0 {
		hcpTFOpts.Volumes = expandProviderVolumeOptions(v)
	}

	if h, ok := in["timeout_seconds"].(int); ok {
		hcpTFOpts.TimeoutSeconds = int64(h)
	}

	return hcpTFOpts

}

func expandSystemProviderOptions(p []any) *eaaspb.SystemProviderOptions {
	spo := &eaaspb.SystemProviderOptions{}
	if len(p) == 0 || p[0] == nil {
		return spo
	}
	in := p[0].(map[string]any)

	if h, ok := in["kind"].(string); ok {
		spo.Kind = h
	}

	return spo
}

func expandTerragruntProviderOptions(p []any) *eaaspb.TerragruntProviderOptions {
	if len(p) == 0 || p[0] == nil {
		return nil
	}
	tpo := &eaaspb.TerragruntProviderOptions{}
	return tpo
}

func expandPulumiProviderOptions(p []any) *eaaspb.PulumiProviderOptions {
	if len(p) == 0 || p[0] == nil {
		return nil
	}
	ppo := &eaaspb.PulumiProviderOptions{}
	return ppo
}

func expandProviderHooks(p []any) (*eaaspb.ResourceTemplateProviderHooks, error) {
	rtph := &eaaspb.ResourceTemplateProviderHooks{}
	if len(p) == 0 || p[0] == nil {
		return rtph, nil
	}

	in := p[0].(map[string]any)

	var err error
	if h, ok := in["terraform"].([]any); ok && len(h) > 0 {
		rtph.Terraform, err = expandTerraformProviderHooks(h)
		if err != nil {
			return nil, err
		}
	}

	if h, ok := in["pulumi"].([]any); ok && len(h) > 0 {
		rtph.Pulumi, err = expandPulumiProviderHooks(h)
		if err != nil {
			return nil, err
		}
	}

	if h, ok := in["open_tofu"].([]any); ok && len(h) > 0 {
		rtph.OpenTofu, err = expandOpenTofuProviderHooks(h)
		if err != nil {
			return nil, err
		}
	}

	if h, oj := in["hcp_terraform"].([]any); oj && len(h) > 0 {
		rtph.HcpTerraform, err = expandHcpTerraformProviderHooks(h)
		if err != nil {
			return nil, err
		}
	}

	if h, oj := in["system"].([]any); oj && len(h) > 0 {
		rtph.System, err = expandSystemProviderHooks(h)
		if err != nil {
			return nil, err
		}
	}

	return rtph, nil

}

func expandTerraformProviderHooks(p []any) (*eaaspb.TerraformProviderHooks, error) {
	tph := &eaaspb.TerraformProviderHooks{}
	if len(p) == 0 || p[0] == nil {
		return tph, nil
	}

	in := p[0].(map[string]any)

	var err error
	if h, ok := in["deploy"].([]any); ok && len(h) > 0 {
		tph.Deploy, err = expandTerraformDeployHooks(h)
		if err != nil {
			return nil, err
		}
	}

	if h, ok := in["destroy"].([]any); ok && len(h) > 0 {
		tph.Destroy, err = expandTerraformDestroyHooks(h)
		if err != nil {
			return nil, err
		}
	}

	return tph, nil
}

func expandPulumiProviderHooks(p []any) (*eaaspb.PulumiProviderHooks, error) {
	pph := &eaaspb.PulumiProviderHooks{}
	if len(p) == 0 || p[0] == nil {
		return pph, nil
	}

	in := p[0].(map[string]any)

	var err error
	if h, ok := in["deploy"].([]any); ok {
		pph.Deploy, err = expandPulumiDeployHooks(h)
		if err != nil {
			return nil, err
		}
	}

	if h, ok := in["destroy"].([]any); ok {
		pph.Destroy, err = expandPulumiDestroyHooks(h)
		if err != nil {
			return nil, err
		}
	}

	return pph, nil
}

func expandOpenTofuProviderHooks(p []any) (*eaaspb.OpenTofuProviderHooks, error) {
	tph := &eaaspb.OpenTofuProviderHooks{}
	if len(p) == 0 || p[0] == nil {
		return tph, nil
	}

	in := p[0].(map[string]any)

	var err error
	if h, ok := in["deploy"].([]any); ok && len(h) > 0 {
		tph.Deploy, err = expandOpenTofuDeployHooks(h)
		if err != nil {
			return nil, err
		}
	}

	if h, ok := in["destroy"].([]any); ok && len(h) > 0 {
		tph.Destroy, err = expandOpenTofuDestroyHooks(h)
		if err != nil {
			return nil, err
		}
	}

	return tph, nil
}

func expandHcpTerraformProviderHooks(p []any) (*eaaspb.HCPTerraformProviderHooks, error) {
	tph := &eaaspb.HCPTerraformProviderHooks{}
	if len(p) == 0 || p[0] == nil {
		return tph, nil
	}

	in := p[0].(map[string]any)

	var err error
	if h, ok := in["deploy"].([]any); ok && len(h) > 0 {
		tph.Deploy, err = expandHcpTerraformDeployHooks(h)
		if err != nil {
			return nil, err
		}
	}

	if h, ok := in["destroy"].([]any); ok && len(h) > 0 {
		tph.Destroy, err = expandHcpTerraformDestroyHooks(h)
		if err != nil {
			return nil, err
		}
	}

	return tph, nil
}
func expandSystemProviderHooks(p []any) (*eaaspb.SystemProviderHooks, error) {
	sph := &eaaspb.SystemProviderHooks{}
	if len(p) == 0 || p[0] == nil {
		return sph, nil
	}

	in := p[0].(map[string]any)

	var err error
	if h, ok := in["deploy"].([]any); ok && len(h) > 0 {
		sph.Deploy, err = expandSystemDeployHooks(h)
		if err != nil {
			return nil, err
		}
	}

	if h, ok := in["destroy"].([]any); ok && len(h) > 0 {
		sph.Destroy, err = expandSystemDestroyHooks(h)
		if err != nil {
			return nil, err
		}
	}

	return sph, nil
}

func expandTerraformDeployHooks(p []any) (*eaaspb.TerraformDeployHooks, error) {
	tdh := &eaaspb.TerraformDeployHooks{}
	if len(p) == 0 || p[0] == nil {
		return tdh, nil
	}

	in := p[0].(map[string]any)

	var err error
	if h, ok := in["init"].([]any); ok && len(h) > 0 {
		tdh.Init, err = expandLifecycleEventHooks(h)
		if err != nil {
			return nil, err
		}
	}

	if h, ok := in["plan"].([]any); ok && len(h) > 0 {
		tdh.Plan, err = expandLifecycleEventHooks(h)
		if err != nil {
			return nil, err
		}
	}

	if h, ok := in["apply"].([]any); ok && len(h) > 0 {
		tdh.Apply, err = expandLifecycleEventHooks(h)
		if err != nil {
			return nil, err
		}
	}

	if h, ok := in["output"].([]any); ok && len(h) > 0 {
		tdh.Output, err = expandLifecycleEventHooks(h)
		if err != nil {
			return nil, err
		}
	}

	return tdh, nil
}

func expandTerraformDestroyHooks(p []any) (*eaaspb.TerraformDestroyHooks, error) {
	tdh := &eaaspb.TerraformDestroyHooks{}
	if len(p) == 0 || p[0] == nil {
		return tdh, nil
	}

	in := p[0].(map[string]any)

	var err error
	if h, ok := in["init"].([]any); ok && len(h) > 0 {
		tdh.Init, err = expandLifecycleEventHooks(h)
		if err != nil {
			return nil, err
		}
	}

	if h, ok := in["plan"].([]any); ok && len(h) > 0 {
		tdh.Plan, err = expandLifecycleEventHooks(h)
		if err != nil {
			return nil, err
		}
	}

	if h, ok := in["destroy"].([]any); ok && len(h) > 0 {
		tdh.Destroy, err = expandLifecycleEventHooks(h)
		if err != nil {
			return nil, err
		}
	}

	return tdh, nil
}

func expandOpenTofuDeployHooks(p []any) (*eaaspb.OpenTofuDeployHooks, error) {
	tdh := &eaaspb.OpenTofuDeployHooks{}
	if len(p) == 0 || p[0] == nil {
		return tdh, nil
	}

	in := p[0].(map[string]any)

	var err error
	if h, ok := in["init"].([]any); ok && len(h) > 0 {
		tdh.Init, err = expandLifecycleEventHooks(h)
		if err != nil {
			return nil, err
		}
	}

	if h, ok := in["plan"].([]any); ok && len(h) > 0 {
		tdh.Plan, err = expandLifecycleEventHooks(h)
		if err != nil {
			return nil, err
		}
	}

	if h, ok := in["apply"].([]any); ok && len(h) > 0 {
		tdh.Apply, err = expandLifecycleEventHooks(h)
		if err != nil {
			return nil, err
		}
	}

	if h, ok := in["output"].([]any); ok && len(h) > 0 {
		tdh.Output, err = expandLifecycleEventHooks(h)
		if err != nil {
			return nil, err
		}
	}

	return tdh, nil
}

func expandOpenTofuDestroyHooks(p []any) (*eaaspb.OpenTofuDestroyHooks, error) {
	tdh := &eaaspb.OpenTofuDestroyHooks{}
	if len(p) == 0 || p[0] == nil {
		return tdh, nil
	}

	in := p[0].(map[string]any)

	var err error
	if h, ok := in["init"].([]any); ok && len(h) > 0 {
		tdh.Init, err = expandLifecycleEventHooks(h)
		if err != nil {
			return nil, err
		}
	}

	if h, ok := in["plan"].([]any); ok && len(h) > 0 {
		tdh.Plan, err = expandLifecycleEventHooks(h)
		if err != nil {
			return nil, err
		}
	}

	if h, ok := in["destroy"].([]any); ok && len(h) > 0 {
		tdh.Destroy, err = expandLifecycleEventHooks(h)
		if err != nil {
			return nil, err
		}
	}

	return tdh, nil
}

func expandHcpTerraformDeployHooks(p []any) (*eaaspb.HCPTerraformDeployHooks, error) {
	tdh := &eaaspb.HCPTerraformDeployHooks{}
	if len(p) == 0 || p[0] == nil {
		return tdh, nil
	}

	in := p[0].(map[string]any)

	var err error
	if h, ok := in["init"].([]any); ok && len(h) > 0 {
		tdh.Init, err = expandLifecycleEventHooks(h)
		if err != nil {
			return nil, err
		}
	}

	if h, ok := in["plan"].([]any); ok && len(h) > 0 {
		tdh.Plan, err = expandLifecycleEventHooks(h)
		if err != nil {
			return nil, err
		}
	}

	if h, ok := in["apply"].([]any); ok && len(h) > 0 {
		tdh.Apply, err = expandLifecycleEventHooks(h)
		if err != nil {
			return nil, err
		}
	}

	if h, ok := in["output"].([]any); ok && len(h) > 0 {
		tdh.Output, err = expandLifecycleEventHooks(h)
		if err != nil {
			return nil, err
		}
	}

	return tdh, nil
}

func expandHcpTerraformDestroyHooks(p []any) (*eaaspb.HCPTerraformDestroyHooks, error) {
	tdh := &eaaspb.HCPTerraformDestroyHooks{}
	if len(p) == 0 || p[0] == nil {
		return tdh, nil
	}

	in := p[0].(map[string]any)

	var err error
	if h, ok := in["init"].([]any); ok && len(h) > 0 {
		tdh.Init, err = expandLifecycleEventHooks(h)
		if err != nil {
			return nil, err
		}
	}

	if h, ok := in["plan"].([]any); ok && len(h) > 0 {
		tdh.Plan, err = expandLifecycleEventHooks(h)
		if err != nil {
			return nil, err
		}
	}

	if h, ok := in["destroy"].([]any); ok && len(h) > 0 {
		tdh.Destroy, err = expandLifecycleEventHooks(h)
		if err != nil {
			return nil, err
		}
	}

	return tdh, nil
}

func expandSystemDeployHooks(p []any) (*eaaspb.SystemDeployHooks, error) {
	sdh := &eaaspb.SystemDeployHooks{}
	if len(p) == 0 || p[0] == nil {
		return sdh, nil
	}

	in := p[0].(map[string]any)

	var err error
	if h, ok := in["apply"].([]any); ok && len(h) > 0 {
		sdh.Apply, err = expandLifecycleEventHooks(h)
		if err != nil {
			return nil, err
		}
	}

	return sdh, nil
}

func expandSystemDestroyHooks(p []any) (*eaaspb.SystemDestroyHooks, error) {
	sdh := &eaaspb.SystemDestroyHooks{}
	if len(p) == 0 || p[0] == nil {
		return sdh, nil
	}

	in := p[0].(map[string]any)

	var err error
	if h, ok := in["destroy"].([]any); ok && len(h) > 0 {
		sdh.Destroy, err = expandLifecycleEventHooks(h)
		if err != nil {
			return nil, err
		}
	}

	return sdh, nil
}

func expandPulumiDeployHooks(p []any) (*eaaspb.PulumiDeployHooks, error) {
	pdh := &eaaspb.PulumiDeployHooks{}
	if len(p) == 0 || p[0] == nil {
		return pdh, nil
	}

	in := p[0].(map[string]any)

	var err error
	if h, ok := in["login"].([]any); ok && len(h) > 0 {
		pdh.Login, err = expandLifecycleEventHooks(h)
		if err != nil {
			return nil, err
		}
	}

	if h, ok := in["preview"].([]any); ok && len(h) > 0 {
		pdh.Preview, err = expandLifecycleEventHooks(h)
		if err != nil {
			return nil, err
		}
	}

	if h, ok := in["up"].([]any); ok && len(h) > 0 {
		pdh.Up, err = expandLifecycleEventHooks(h)
		if err != nil {
			return nil, err
		}
	}

	if h, ok := in["output"].([]any); ok && len(h) > 0 {
		pdh.Output, err = expandLifecycleEventHooks(h)
		if err != nil {
			return nil, err
		}
	}

	return pdh, nil
}

func expandPulumiDestroyHooks(p []any) (*eaaspb.PulumiDestroyHooks, error) {
	pdh := &eaaspb.PulumiDestroyHooks{}
	if len(p) == 0 || p[0] == nil {
		return pdh, nil
	}

	in := p[0].(map[string]any)

	var err error
	if h, ok := in["login"].([]any); ok && len(h) > 0 {
		pdh.Login, err = expandLifecycleEventHooks(h)
		if err != nil {
			return nil, err
		}
	}

	if h, ok := in["preview"].([]any); ok && len(h) > 0 {
		pdh.Preview, err = expandLifecycleEventHooks(h)
		if err != nil {
			return nil, err
		}
	}

	if h, ok := in["destroy"].([]any); ok && len(h) > 0 {
		pdh.Destroy, err = expandLifecycleEventHooks(h)
		if err != nil {
			return nil, err
		}
	}

	return pdh, nil
}

func expandLifecycleEventHooks(p []any) (*eaaspb.LifecycleEventHooks, error) {
	lh := &eaaspb.LifecycleEventHooks{}
	if len(p) == 0 || p[0] == nil {
		return lh, nil
	}

	in := p[0].(map[string]any)

	var err error
	if h, ok := in["before"].([]any); ok && len(h) > 0 {
		lh.Before, err = expandEaasHooks(h)
		if err != nil {
			return nil, err
		}
	}

	if h, ok := in["after"].([]any); ok && len(h) > 0 {
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

	v, ok := d.Get("spec").([]any)
	if !ok {
		v = []any{}
	}

	var ret []any
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

func flattenResourceTemplateSpec(in *eaaspb.ResourceTemplateSpec, p []any) ([]any, error) {
	if in == nil {
		return nil, nil
	}

	obj := map[string]any{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]any)
	}

	obj["version"] = in.Version
	obj["version_state"] = in.VersionState
	obj["provider"] = in.Provider

	v, _ := obj["provider_options"].([]any)
	obj["provider_options"] = flattenProviderOptions(in.ProviderOptions, v)

	obj["repository_options"] = flattenRepositoryOptions(in.RepositoryOptions)

	v, _ = obj["contexts"].([]any)
	obj["contexts"] = flattenContexts(in.Contexts, v)

	v, _ = obj["variables"].([]any)
	obj["variables"] = flattenVariables(in.Variables, v)

	v, _ = obj["hooks"].([]any)
	obj["hooks"] = flattenResourceHooks(in.Hooks, v)

	obj["agents"] = flattenEaasAgents(in.Agents)
	obj["agent_pools"] = flattenEaasAgents(in.AgentPools)
	obj["sharing"] = flattenSharingSpec(in.Sharing)

	if w, ok := obj["artifact_driver"].([]any); ok && len(w) > 0 {
		obj["artifact_driver"] = flattenWorkflowHandlerCompoundRef(in.ArtifactWorkflowHandler)
	} else {
		obj["artifact_workflow_handler"] = flattenWorkflowHandlerCompoundRef(in.ArtifactWorkflowHandler)
	}

	v, _ = obj["actions"].([]any)
	obj["actions"] = flattenActions(in.Actions, v)

	return []any{obj}, nil
}

func flattenProviderOptions(in *eaaspb.ResourceTemplateProviderOptions, p []any) []any {
	if in == nil {
		return nil
	}
	obj := make(map[string]any)
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]any)
	}

	obj["terraform"] = flattenTerraformProviderOptions(in.Terraform)
	obj["system"] = flattenSystemProviderOptions(in.System)
	obj["terragrunt"] = flattenTerragruntProviderOptions(in.Terragrunt)
	obj["pulumi"] = flattenPulumiProviderOptions(in.Pulumi)
	if v, ok := obj["driver"].([]any); ok && len(v) > 0 {
		obj["driver"] = flattenWorkflowHandlerCompoundRef(in.WorkflowHandler)
	} else {
		obj["workflow_handler"] = flattenWorkflowHandlerCompoundRef(in.WorkflowHandler)
	}
	obj["open_tofu"] = flattenOpenTofuProviderOptions(in.OpenTofu)
	v, _ := obj["custom"].([]any)
	obj["custom"] = flattenCustomProviderOptions(in.Custom, v)
	obj["hcp_terraform"] = flattenHcpTerraformProviderOptions(in.HcpTerraform)

	return []any{obj}
}

func flattenOpenTofuProviderOptions(in *eaaspb.OpenTofuProviderOptions) []any {
	if in == nil {
		return nil
	}

	obj := make(map[string]any)
	obj["version"] = in.Version
	obj["var_files"] = toArrayInterface(in.VarFiles)
	obj["backend_configs"] = toArrayInterface(in.BackendConfigs)
	obj["backend_type"] = in.BackendType
	obj["refresh"] = flattenBoolValue(in.Refresh)
	obj["lock"] = flattenBoolValue(in.Lock)
	obj["lock_timeout_seconds"] = in.LockTimeoutSeconds
	obj["plugin_dirs"] = toArrayInterface(in.PluginDirs)
	obj["target_resources"] = toArrayInterface(in.TargetResources)
	v, _ := obj["volumes"].([]any)
	obj["volumes"] = flattenProviderVolumeOptions(in.Volumes, v)
	obj["timeout_seconds"] = in.TimeoutSeconds
	return []any{obj}
}

func flattenCustomProviderOptions(in *eaaspb.CustomProviderOptions, p []any) []any {
	if in == nil {
		return nil
	}
	obj := map[string]any{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]any)
	}
	v, _ := obj["tasks"].([]any)
	obj["tasks"] = flattenEaasHooks(in.Tasks, v)
	obj["reverse_on_destroy"] = in.ReverseOnDestroy
	return []any{obj}
}

func flattenTerraformProviderOptions(in *eaaspb.TerraformProviderOptions) []any {
	if in == nil {
		return nil
	}

	obj := map[string]any{
		"version":                in.Version,
		"use_system_state_store": flattenBoolValue(in.UseSystemStateStore),
		"var_files":              toArrayInterface(in.VarFiles),
		"backend_configs":        toArrayInterface(in.BackendConfigs),
		"backend_type":           in.BackendType,
		"refresh":                flattenBoolValue(in.Refresh),
		"lock":                   flattenBoolValue(in.Lock),
		"lock_timeout_seconds":   in.LockTimeoutSeconds,
		"plugin_dirs":            toArrayInterface(in.PluginDirs),
		"target_resources":       toArrayInterface(in.TargetResources),
		"with_terraform_cloud":   flattenBoolValue(in.WithTerraformCloud),
		"volumes":                flattenProviderVolumeOptions(in.Volumes, nil),
		"timeout_seconds":        in.TimeoutSeconds,
	}
	return []any{obj}
}

func flattenHcpTerraformProviderOptions(in *eaaspb.HCPTerraformProviderOptions) []any {
	if in == nil {
		return nil
	}

	obj := map[string]any{
		"var_files":            toArrayInterface(in.VarFiles),
		"refresh":              flattenBoolValue(in.Refresh),
		"lock":                 flattenBoolValue(in.Lock),
		"lock_timeout_seconds": in.LockTimeoutSeconds,
		"plugin_dirs":          toArrayInterface(in.PluginDirs),
		"target_resources":     toArrayInterface(in.TargetResources),
		"volumes":              flattenProviderVolumeOptions(in.Volumes, nil),
		"timeout_seconds":      in.TimeoutSeconds,
	}
	return []any{obj}
}

func flattenSystemProviderOptions(in *eaaspb.SystemProviderOptions) []any {
	if in == nil {
		return nil
	}

	obj := map[string]any{
		"kind": in.Kind,
	}
	return []any{obj}
}

func flattenTerragruntProviderOptions(in *eaaspb.TerragruntProviderOptions) []any {
	if in == nil {
		return nil
	}

	obj := make(map[string]any)
	return []any{obj}
}

func flattenPulumiProviderOptions(in *eaaspb.PulumiProviderOptions) []any {
	if in == nil {
		return nil
	}

	obj := make(map[string]any)
	return []any{obj}
}

func flattenRepositoryOptions(in *eaaspb.ResourceTemplateRepositoryOptions) []any {
	if in == nil {
		return nil
	}

	obj := map[string]any{
		"name":           in.Name,
		"branch":         in.Branch,
		"directory_path": in.DirectoryPath,
	}
	return []any{obj}
}

func flattenContexts(input []*eaaspb.ConfigContextCompoundRef, p []any) []any {
	log.Println("flatten contexts start")
	if input == nil {
		return nil
	}

	out := make([]any, len(input))
	for i, in := range input {
		log.Println("flatten context ", in)
		obj := map[string]any{}
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]any)
		}
		obj["name"] = in.Name
		obj["data"] = flattenConfigContextInline(in.Data)
		out[i] = &obj
	}

	return out
}

func flattenResourceHooks(in *eaaspb.ResourceHooks, p []any) []any {
	if in == nil {
		return nil
	}

	obj := make(map[string]any)
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]any)
	}

	v, _ := obj["on_completion"].([]any)
	obj["on_completion"] = flattenEaasHooks(in.OnCompletion, v)

	v, _ = obj["on_success"].([]any)
	obj["on_success"] = flattenEaasHooks(in.OnSuccess, v)

	v, _ = obj["on_failure"].([]any)
	obj["on_failure"] = flattenEaasHooks(in.OnFailure, v)

	v, _ = obj["on_init"].([]any)
	obj["on_init"] = flattenEaasHooks(in.OnInit, v)

	v, _ = obj["provider"].([]any)
	obj["provider"] = flattenProviderHooks(in.Provider, v)

	return []any{obj}
}

func flattenProviderHooks(input *eaaspb.ResourceTemplateProviderHooks, p []any) []any {
	if input == nil {
		return nil
	}

	obj := make(map[string]any)
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]any)
	}

	v, _ := obj["terraform"].([]any)
	obj["terraform"] = flattenTerraformProviderHooks(input.Terraform, v)

	v, _ = obj["open_tofu"].([]any)
	obj["open_tofu"] = flattenOpenTofuProviderHooks(input.OpenTofu, v)

	v, _ = obj["pulumi"].([]any)
	obj["pulumi"] = flattenPulumiProviderHooks(input.Pulumi, v)

	v, _ = obj["hcp_terraform"].([]any)
	obj["hcp_terraform"] = flattenHcpTerraformProviderHooks(input.HcpTerraform, v)

	v, _ = obj["system"].([]any)
	obj["system"] = flattenSystemProviderHooks(input.System, v)

	return []any{obj}
}

func flattenTerraformProviderHooks(input *eaaspb.TerraformProviderHooks, p []any) []any {
	if input == nil {
		return nil
	}

	obj := make(map[string]any)
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]any)
	}

	v, _ := obj["deploy"].([]any)
	obj["deploy"] = flattenTerraformDeployHooks(input.Deploy, v)

	v, _ = obj["destroy"].([]any)
	obj["destroy"] = flattenTerraformDestroyHooks(input.Destroy, v)

	return []any{obj}
}

func flattenOpenTofuProviderHooks(input *eaaspb.OpenTofuProviderHooks, p []any) []any {
	if input == nil {
		return nil
	}

	obj := make(map[string]any)
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]any)
	}

	v, _ := obj["deploy"].([]any)
	obj["deploy"] = flattenOpenTofuDeployHooks(input.Deploy, v)

	v, _ = obj["destroy"].([]any)
	obj["destroy"] = flattenOpenTofuDestroyHooks(input.Destroy, v)

	return []any{obj}
}

func flattenHcpTerraformProviderHooks(input *eaaspb.HCPTerraformProviderHooks, p []any) []any {
	if input == nil {
		return nil
	}

	obj := make(map[string]any)
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]any)
	}

	v, _ := obj["deploy"].([]any)
	obj["deploy"] = flattenHcpTerraformDeployHooks(input.Deploy, v)

	v, _ = obj["destroy"].([]any)
	obj["destroy"] = flattenHcpTerraformDestroyHooks(input.Destroy, v)

	return []any{obj}
}

func flattenSystemProviderHooks(input *eaaspb.SystemProviderHooks, p []any) []any {
	if input == nil {
		return nil
	}

	obj := make(map[string]any)
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]any)
	}

	v, _ := obj["deploy"].([]any)
	obj["deploy"] = flattenSystemDeployHooks(input.Deploy, v)

	v, _ = obj["destroy"].([]any)
	obj["destroy"] = flattenSystemDestroyHooks(input.Destroy, v)

	return []any{obj}
}

func flattenPulumiProviderHooks(input *eaaspb.PulumiProviderHooks, p []any) []any {
	if input == nil {
		return nil
	}

	obj := make(map[string]any)
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]any)
	}

	v, _ := obj["deploy"].([]any)
	obj["deploy"] = flattenPulumiDeployHooks(input.Deploy, v)

	v, _ = obj["destroy"].([]any)
	obj["destroy"] = flattenPulumiDestroyHooks(input.Destroy, v)

	return []any{obj}
}

func flattenTerraformDeployHooks(input *eaaspb.TerraformDeployHooks, p []any) []any {
	log.Println("flatten terraform deploy hooks start")
	if input == nil {
		return nil
	}

	obj := make(map[string]any)
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]any)
	}

	v, _ := obj["init"].([]any)
	obj["init"] = flattenLifecycleEventHooks(input.Init, v)

	v, _ = obj["plan"].([]any)
	obj["plan"] = flattenLifecycleEventHooks(input.Plan, v)

	v, _ = obj["apply"].([]any)
	obj["apply"] = flattenLifecycleEventHooks(input.Apply, v)

	v, _ = obj["output"].([]any)
	obj["output"] = flattenLifecycleEventHooks(input.Output, v)
	return []any{obj}
}

func flattenTerraformDestroyHooks(input *eaaspb.TerraformDestroyHooks, p []any) []any {
	log.Println("flatten terraform destroy hooks start")
	if input == nil {
		return nil
	}

	obj := make(map[string]any)
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]any)
	}

	v, _ := obj["init"].([]any)
	obj["init"] = flattenLifecycleEventHooks(input.Init, v)

	v, _ = obj["plan"].([]any)
	obj["plan"] = flattenLifecycleEventHooks(input.Plan, v)

	v, _ = obj["destroy"].([]any)
	obj["destroy"] = flattenLifecycleEventHooks(input.Destroy, v)

	return []any{obj}
}

func flattenOpenTofuDeployHooks(input *eaaspb.OpenTofuDeployHooks, p []any) []any {
	log.Println("flatten opentofu deploy hooks start")
	if input == nil {
		return nil
	}

	obj := make(map[string]any)
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]any)
	}

	v, _ := obj["init"].([]any)
	obj["init"] = flattenLifecycleEventHooks(input.Init, v)

	v, _ = obj["plan"].([]any)
	obj["plan"] = flattenLifecycleEventHooks(input.Plan, v)

	v, _ = obj["apply"].([]any)
	obj["apply"] = flattenLifecycleEventHooks(input.Apply, v)

	v, _ = obj["output"].([]any)
	obj["output"] = flattenLifecycleEventHooks(input.Output, v)
	return []any{obj}
}

func flattenOpenTofuDestroyHooks(input *eaaspb.OpenTofuDestroyHooks, p []any) []any {
	log.Println("flatten opentofu destroy hooks start")
	if input == nil {
		return nil
	}

	obj := make(map[string]any)
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]any)
	}

	v, _ := obj["init"].([]any)
	obj["init"] = flattenLifecycleEventHooks(input.Init, v)

	v, _ = obj["plan"].([]any)
	obj["plan"] = flattenLifecycleEventHooks(input.Plan, v)

	v, _ = obj["destroy"].([]any)
	obj["destroy"] = flattenLifecycleEventHooks(input.Destroy, v)

	return []any{obj}
}

func flattenHcpTerraformDeployHooks(input *eaaspb.HCPTerraformDeployHooks, p []any) []any {
	log.Println("flatten hcp terraform deploy hooks start")
	if input == nil {
		return nil
	}

	obj := make(map[string]any)
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]any)
	}

	v, _ := obj["init"].([]any)
	obj["init"] = flattenLifecycleEventHooks(input.Init, v)

	v, _ = obj["plan"].([]any)
	obj["plan"] = flattenLifecycleEventHooks(input.Plan, v)

	v, _ = obj["apply"].([]any)
	obj["apply"] = flattenLifecycleEventHooks(input.Apply, v)

	v, _ = obj["output"].([]any)
	obj["output"] = flattenLifecycleEventHooks(input.Output, v)
	return []any{obj}
}

func flattenHcpTerraformDestroyHooks(input *eaaspb.HCPTerraformDestroyHooks, p []any) []any {
	log.Println("flatten hcp terraform destroy hooks start")
	if input == nil {
		return nil
	}

	obj := make(map[string]any)
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]any)
	}

	v, _ := obj["init"].([]any)
	obj["init"] = flattenLifecycleEventHooks(input.Init, v)

	v, _ = obj["plan"].([]any)
	obj["plan"] = flattenLifecycleEventHooks(input.Plan, v)

	v, _ = obj["destroy"].([]any)
	obj["destroy"] = flattenLifecycleEventHooks(input.Destroy, v)
	return []any{obj}
}

func flattenSystemDeployHooks(input *eaaspb.SystemDeployHooks, p []any) []any {
	log.Println("flatten system deploy hooks start")
	if input == nil {
		return nil
	}

	obj := make(map[string]any)
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]any)
	}

	v, _ := obj["apply"].([]any)
	obj["apply"] = flattenLifecycleEventHooks(input.Apply, v)
	return []any{obj}
}

func flattenSystemDestroyHooks(input *eaaspb.SystemDestroyHooks, p []any) []any {
	log.Println("flatten system destroy hooks start")
	if input == nil {
		return nil
	}

	obj := make(map[string]any)
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]any)
	}

	v, _ := obj["destroy"].([]any)
	obj["destroy"] = flattenLifecycleEventHooks(input.Destroy, v)
	return []any{obj}
}

func flattenPulumiDeployHooks(input *eaaspb.PulumiDeployHooks, p []any) []any {
	log.Println("flatten pulumi deploy hooks start")
	if input == nil {
		return nil
	}

	obj := make(map[string]any)
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]any)
	}
	v, _ := obj["login"].([]any)
	obj["login"] = flattenLifecycleEventHooks(input.Login, v)

	v, _ = obj["preview"].([]any)
	obj["preview"] = flattenLifecycleEventHooks(input.Preview, v)

	v, _ = obj["up"].([]any)
	obj["up"] = flattenLifecycleEventHooks(input.Up, v)

	v, _ = obj["output"].([]any)
	obj["output"] = flattenLifecycleEventHooks(input.Output, v)

	return []any{obj}
}

func flattenPulumiDestroyHooks(input *eaaspb.PulumiDestroyHooks, p []any) []any {
	log.Println("flatten pulumi destroy hooks start")
	if input == nil {
		return nil
	}

	obj := make(map[string]any)
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]any)
	}
	v, _ := obj["login"].([]any)
	obj["login"] = flattenLifecycleEventHooks(input.Login, v)

	v, _ = obj["preview"].([]any)
	obj["preview"] = flattenLifecycleEventHooks(input.Preview, v)

	v, _ = obj["destroy"].([]any)
	obj["destroy"] = flattenLifecycleEventHooks(input.Destroy, v)
	return []any{obj}
}

func flattenLifecycleEventHooks(input *eaaspb.LifecycleEventHooks, p []any) []any {
	if input == nil {
		return nil
	}

	obj := make(map[string]any)
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]any)
	}
	v, _ := obj["before"].([]any)
	obj["before"] = flattenEaasHooks(input.Before, v)

	v, _ = obj["after"].([]any)
	obj["after"] = flattenEaasHooks(input.After, v)
	return []any{obj}
}

func flattenEaasAgents(input []*commonpb.ResourceNameAndVersionRef) []any {
	log.Println("flatten agents start")
	if len(input) == 0 {
		return nil
	}

	out := make([]any, len(input))
	for i, in := range input {
		log.Println("flatten agent ", in)
		obj := map[string]any{
			"name": in.Name,
		}
		out[i] = &obj
	}

	return out
}

func expandProviderVolumeOptions(p []any) []*eaaspb.ProviderVolumeOptions {
	volumes := make([]*eaaspb.ProviderVolumeOptions, 0)
	if len(p) == 0 {
		return volumes
	}

	for indx := range p {
		volume := &eaaspb.ProviderVolumeOptions{}
		if p[indx] == nil {
			return volumes
		}
		in := p[indx].(map[string]any)

		if mp, ok := in["mount_path"].(string); ok && len(mp) > 0 {
			volume.MountPath = mp
		}

		if pvcsz, ok := in["pvc_size_gb"].(string); ok && len(pvcsz) > 0 {
			volume.PvcSizeGB = pvcsz
		}

		if pvcsc, ok := in["pvc_storage_class"].(string); ok && len(pvcsc) > 0 {
			volume.PvcStorageClass = pvcsc
		}

		if usepvc, ok := in["use_pvc"].([]any); ok && len(usepvc) > 0 {
			volume.UsePVC = expandBoolValue(usepvc)
		}

		if enableBackupAndRestore, ok := in["enable_backup_and_restore"].(bool); ok {
			volume.EnableBackupAndRestore = enableBackupAndRestore
		}

		volumes = append(volumes, volume)

	}

	return volumes
}

func flattenProviderVolumeOptions(input []*eaaspb.ProviderVolumeOptions, p []any) []any {
	if len(input) == 0 {
		return nil
	}

	out := make([]any, len(input))
	for i, in := range input {
		log.Println("flatten provider volume options", in)
		obj := map[string]any{}
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]any)
		}
		obj["use_pvc"] = flattenBoolValue(in.UsePVC)
		obj["mount_path"] = in.MountPath
		obj["pvc_size_gb"] = in.PvcSizeGB
		obj["pvc_storage_class"] = in.PvcStorageClass
		obj["enable_backup_and_restore"] = in.EnableBackupAndRestore
		out[i] = &obj
	}

	return out
}

func resourceResourceTemplateImport(d *schema.ResourceData, m any) ([]*schema.ResourceData, error) {

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
