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

func resourceEnvironmentTemplate() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceEnvironmentTemplateCreate,
		ReadContext:   resourceEnvironmentTemplateRead,
		UpdateContext: resourceEnvironmentTemplateUpdate,
		DeleteContext: resourceEnvironmentTemplateDelete,
		Importer: &schema.ResourceImporter{
			State: resourceEnvironmentTemplateImport,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema:        resource.EnvironmentTemplateSchema.Schema,
	}
}

func resourceEnvironmentTemplateCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	log.Println("environment template create")
	return environmentTemplateUpsert(ctx, d, m)
}

func environmentTemplateUpsert(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("environment template upsert starts")
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	environmenttemplate, err := expandEnvironmentTemplate(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, TF_USER_AGENT, options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.EaasV1().EnvironmentTemplate().Apply(ctx, environmenttemplate, options.ApplyOptions{})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(environmenttemplate.Metadata.Name)
	return diags
}

func resourceEnvironmentTemplateRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Println("environment template read starts ")
	meta := GetMetaData(d)
	if meta == nil {
		return diag.FromErr(fmt.Errorf("%s", "failed to read environment template "))
	}
	if d.State() != nil && d.State().ID != "" {
		meta.Name = d.State().ID
	}

	et, err := expandEnvironmentTemplate(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, TF_USER_AGENT, options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		log.Println("read client err")
		return diag.FromErr(err)
	}

	environmenttemplate, err := client.EaasV1().EnvironmentTemplate().Get(ctx, options.GetOptions{
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

	if et.GetSpec().GetSharing() != nil && !et.GetSpec().GetSharing().GetEnabled() && environmenttemplate.GetSpec().GetSharing() == nil {
		environmenttemplate.Spec.Sharing = &commonpb.SharingSpec{}
		environmenttemplate.Spec.Sharing.Enabled = false
		environmenttemplate.Spec.Sharing.Projects = et.GetSpec().GetSharing().GetProjects()
	}

	err = flattenEnvironmentTemplate(d, environmenttemplate)
	if err != nil {
		log.Println("read flatten err")
		return diag.FromErr(err)
	}
	return diags

}

func resourceEnvironmentTemplateUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	return environmentTemplateUpsert(ctx, d, m)
}

func resourceEnvironmentTemplateDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Println("environment template delete starts")
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

	rt, err := expandEnvironmentTemplate(d)
	if err != nil {
		log.Println("error while expanding environment template during delete")
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, TF_USER_AGENT, options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return diag.FromErr(err)
	}
	err = client.EaasV1().EnvironmentTemplate().Delete(ctx, options.DeleteOptions{
		Name:    rt.Metadata.Name,
		Project: rt.Metadata.Project,
	})

	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func expandEnvironmentTemplate(in *schema.ResourceData) (*eaaspb.EnvironmentTemplate, error) {
	log.Println("expand environment template")
	if in == nil {
		return nil, fmt.Errorf("%s", "expand environment template empty input")
	}
	obj := &eaaspb.EnvironmentTemplate{}

	if v, ok := in.Get("metadata").([]any); ok && len(v) > 0 {
		obj.Metadata = expandV3MetaData(v)
	}

	if v, ok := in.Get("spec").([]any); ok && len(v) > 0 {
		objSpec, err := expandEnvironmentTemplateSpec(v)
		if err != nil {
			return nil, err
		}
		obj.Spec = objSpec
	}

	obj.ApiVersion = "eaas.envmgmt.io/v1"
	obj.Kind = "EnvironmentTemplate"
	return obj, nil
}

func expandEnvironmentTemplateSpec(p []any) (*eaaspb.EnvironmentTemplateSpec, error) {
	log.Println("expand environment template spec")
	spec := &eaaspb.EnvironmentTemplateSpec{}
	if len(p) == 0 || p[0] == nil {
		return spec, fmt.Errorf("%s", "expand environment template spec empty input")
	}

	in := p[0].(map[string]any)

	if v, ok := in["version"].(string); ok && len(v) > 0 {
		spec.Version = v
	}

	if vs, ok := in["version_state"].(string); ok && len(vs) > 0 {
		spec.VersionState = vs
	}

	if iconurl, ok := in["icon_url"].(string); ok && len(iconurl) > 0 {
		spec.IconURL = iconurl
	}

	if readme, ok := in["readme"].(string); ok && len(readme) > 0 {
		spec.Readme = readme
	}

	var err error
	if p, ok := in["resources"].([]any); ok && len(p) > 0 {
		spec.Resources, err = expandEnvironmentResources(p)
		if err != nil {
			return nil, err
		}
	}

	if v, ok := in["variables"].([]any); ok && len(v) > 0 {
		spec.Variables = expandVariables(v)
	}

	if h, ok := in["hooks"].([]any); ok && len(h) > 0 {
		spec.Hooks, err = expandEnvironmentHooks(h)
		if err != nil {
			return nil, err
		}
	}

	if ag, ok := in["agents"].([]any); ok && len(ag) > 0 {
		spec.Agents = expandEaasAgents(ag)
	}

	if v, ok := in["sharing"].([]any); ok && len(v) > 0 {
		spec.Sharing = expandSharingSpec(v)
	}

	if v, ok := in["contexts"].([]any); ok && len(v) > 0 {
		spec.Contexts = expandContexts(v)
	}

	if v, ok := in["agent_override"].([]any); ok && len(v) > 0 {
		spec.AgentOverride = expandEaasAgentOverrideOptions(v)
	}

	if s, ok := in["schedules"].([]any); ok && len(s) > 0 {
		spec.Schedules, err = expandSchedules(s)
		if err != nil {
			return nil, err
		}
	}

	if v, ok := in["allow_new_inputs_during_publish"].([]any); ok && len(v) > 0 {
		spec.AllowNewInputsDuringPublish = expandBoolValue(v)
	}

	if s, ok := in["actions"].([]any); ok && len(s) > 0 {
		spec.Actions, err = expandActions(s)
		if err != nil {
			return nil, err
		}
	}

	return spec, nil
}

func expandSchedules(p []any) ([]*eaaspb.Schedules, error) {
	schds := make([]*eaaspb.Schedules, 0)
	if len(p) == 0 || p[0] == nil {
		return schds, nil
	}
	var err error

	for i := range p {
		schd := eaaspb.Schedules{}
		in := p[i].(map[string]any)

		if v, ok := in["name"].(string); ok && len(v) > 0 {
			schd.Name = v
		}

		if v, ok := in["description"].(string); ok && len(v) > 0 {
			schd.Description = v
		}

		if v, ok := in["type"].(string); ok && len(v) > 0 {
			schd.Type = v
		}

		if v, ok := in["cadence"].([]any); ok && len(v) > 0 {
			schd.Cadence = expandCadence(v)
		}

		if v, ok := in["context"].([]any); ok && len(v) > 0 && v[0] != nil {
			schd.Context = expandConfigContextCompoundRef(v[0].(map[string]any))
		}

		if v, ok := in["opt_out_options"].([]any); ok && len(v) > 0 {
			schd.OptOutOptions, err = expandOptOutOptions(v)
			if err != nil {
				return nil, err
			}
		}

		if h, ok := in["workflows"].([]any); ok && len(h) > 0 {
			schd.Workflows, err = expandCustomProviderOptions(h)
			if err != nil {
				return nil, err
			}
		}

		schds = append(schds, &schd)
	}

	return schds, nil
}

func expandOptOutOptions(p []any) (*eaaspb.OptOutOptions, error) {
	ooo := eaaspb.OptOutOptions{}
	if len(p) == 0 || p[0] == nil {
		return &ooo, nil
	}

	var err error
	in := p[0].(map[string]any)
	if h, ok := in["allow_opt_out"].([]any); ok && len(h) > 0 {
		ooo.AllowOptOut = expandBoolValue(h)
	}
	if v, ok := in["max_allowed_duration"].(string); ok && len(v) > 0 {
		ooo.MaxAllowedDuration = v
	}
	if v, ok := in["max_allowed_times"].(int); ok {
		ooo.MaxAllowedTimes = int32(v)
	}
	if h, ok := in["approval"].([]any); ok && len(h) > 0 {
		ooo.Approval, err = expandCustomProviderOptions(h)
		if err != nil {
			return nil, err
		}
	}

	return &ooo, nil
}

func expandCadence(p []any) *eaaspb.ScheduleOptions {
	cadence := eaaspb.ScheduleOptions{}
	if len(p) == 0 || p[0] == nil {
		return &cadence
	}

	in := p[0].(map[string]any)
	if v, ok := in["cron_expression"].(string); ok && len(v) > 0 {
		cadence.CronExpression = v
	}

	if v, ok := in["cron_timezone"].(string); ok && len(v) > 0 {
		cadence.CronTimezone = v
	}

	if v, ok := in["time_to_live"].(string); ok && len(v) > 0 {
		cadence.TimeToLive = v
	}

	if v, ok := in["staggered"].([]any); ok && len(v) > 0 {
		cadence.Staggered = expandStaggered(v)
	}

	return &cadence
}

func expandStaggered(p []any) *eaaspb.Staggered {
	staggered := eaaspb.Staggered{}
	if len(p) == 0 || p[0] == nil {
		return &staggered
	}

	in := p[0].(map[string]any)
	if h, ok := in["enabled"].([]any); ok && len(h) > 0 {
		staggered.Enabled = expandBoolValue(h)
	}
	if v, ok := in["max_interval"].(string); ok && len(v) > 0 {
		staggered.MaxInterval = v
	}

	return &staggered
}

func expandEaasAgentOverrideOptions(p []any) *eaaspb.AgentOverrideOptions {
	agentOverrideOptions := &eaaspb.AgentOverrideOptions{}
	if len(p) == 0 || p[0] == nil {
		return agentOverrideOptions
	}

	in := p[0].(map[string]any)
	if v, ok := in["required"].(bool); ok {
		agentOverrideOptions.Required = v
	}

	if aot, ok := in["type"].(string); ok {
		agentOverrideOptions.Type = aot
	}

	if agnts, ok := in["restricted_agents"].([]any); ok && len(agnts) > 0 {
		agentOverrideOptions.RestrictedAgents = toArrayString(agnts)
	}

	return agentOverrideOptions
}

func expandEnvironmentResources(p []any) ([]*eaaspb.EnvironmentResourceCompoundRef, error) {
	log.Println("expand environment resources")
	if len(p) == 0 || p[0] == nil {
		return nil, fmt.Errorf("%s", "expand environment resources empty input")
	}

	envresources := make([]*eaaspb.EnvironmentResourceCompoundRef, len(p))

	for i := range p {
		obj := eaaspb.EnvironmentResourceCompoundRef{}
		in := p[i].(map[string]any)

		if v, ok := in["type"].(string); ok && len(v) > 0 {
			obj.Type = v
		}

		if v, ok := in["kind"].(string); ok && len(v) > 0 {
			obj.Kind = v
		}

		if v, ok := in["name"].(string); ok && len(v) > 0 {
			obj.Name = v
		}

		if v, ok := in["resource_options"].([]any); ok && len(v) > 0 {
			obj.ResourceOptions = expandResourceOptions(v)
		}

		if v, ok := in["depends_on"].([]any); ok && len(v) > 0 {
			obj.DependsOn = expandDependsOn(v)
		}

		envresources[i] = &obj

	}

	return envresources, nil

}

func expandResourceOptions(p []any) *eaaspb.EnvironmentResourceOptions {
	ro := &eaaspb.EnvironmentResourceOptions{}

	if len(p) == 0 || p[0] == nil {
		return ro
	}

	in := p[0].(map[string]any)

	if dedicated, ok := in["dedicated"].(bool); ok {
		ro.Dedicated = dedicated
	}

	if version, ok := in["version"].(string); ok {
		ro.Version = version
	}

	return ro
}

func expandDependsOn(p []any) []*commonpb.ResourceNameAndVersionRef {
	dependson := make([]*commonpb.ResourceNameAndVersionRef, 0)
	if len(p) == 0 {
		return dependson
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

		dependson = append(dependson, obj)
	}

	return dependson
}

func expandEnvironmentHooks(p []any) (*eaaspb.EnvironmentHooks, error) {
	hooks := &eaaspb.EnvironmentHooks{}

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

	return hooks, nil

}

// Flatteners

func flattenEnvironmentTemplate(d *schema.ResourceData, in *eaaspb.EnvironmentTemplate) error {
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
	ret, err = flattenEnvironmentTemplateSpec(in.Spec, v)
	if err != nil {
		log.Println("flatten environment template spec err")
		return err
	}

	err = d.Set("spec", ret)
	if err != nil {
		log.Println("set spec err")
		return err
	}
	return nil
}

func flattenEnvironmentTemplateSpec(in *eaaspb.EnvironmentTemplateSpec, p []any) ([]any, error) {
	if in == nil {
		return nil, nil
	}

	obj := map[string]any{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]any)
	}
	obj["version"] = in.Version
	obj["version_state"] = in.VersionState
	obj["icon_url"] = in.IconURL
	obj["readme"] = in.Readme

	v, _ := obj["resources"].([]any)
	obj["resources"] = flattenEnvironmentResources(in.Resources, v)

	v, _ = obj["variables"].([]any)
	obj["variables"] = flattenVariables(in.Variables, v)

	v, _ = obj["hooks"].([]any)
	obj["hooks"] = flattenEnvironmentHooks(in.Hooks, v)

	obj["agents"] = flattenEaasAgents(in.Agents)
	obj["sharing"] = flattenSharingSpec(in.Sharing)

	v, _ = obj["contexts"].([]any)
	obj["contexts"] = flattenContexts(in.Contexts, v)

	obj["agent_override"] = flattenEaasAgentOverrideOptions(in.AgentOverride)

	v, _ = obj["schedules"].([]any)
	obj["schedules"] = flattenSchedules(in.Schedules, v)

	obj["allow_new_inputs_during_publish"] = flattenBoolValue(in.AllowNewInputsDuringPublish)

	v, _ = obj["actions"].([]any)
	obj["actions"] = flattenActions(in.Actions, v)

	return []any{obj}, nil
}

func flattenEaasAgentOverrideOptions(in *eaaspb.AgentOverrideOptions) []any {
	if in == nil {
		return nil
	}

	obj := make(map[string]any)
	obj["required"] = in.Required
	obj["type"] = in.Type
	obj["restricted_agents"] = toArrayInterface(in.RestrictedAgents)

	return []any{obj}
}

func flattenSchedules(input []*eaaspb.Schedules, p []any) []any {
	log.Println("flatten schedules start")
	if input == nil {
		return nil
	}

	out := make([]any, len(input))
	for i, in := range input {
		log.Println("flatten schedule ", in)
		obj := map[string]any{}
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]any)
		}
		obj["name"] = in.Name
		obj["description"] = in.Description
		obj["type"] = in.Type

		v, _ := obj["cadence"].([]any)
		obj["cadence"] = flattenCadence(in.Cadence, v)

		context := flattenConfigContextCompoundRef(in.Context)
		if context != nil {
			obj["context"] = []any{context}
		}

		v, _ = obj["opt_out_options"].([]any)
		obj["opt_out_options"] = flattenOptOutOptions(in.OptOutOptions, v)

		v, _ = obj["workflows"].([]any)
		obj["workflows"] = flattenCustomProviderOptions(in.Workflows, v)
		out[i] = &obj
	}

	return out
}

func flattenOptOutOptions(in *eaaspb.OptOutOptions, p []any) []any {
	if in == nil {
		return nil
	}

	obj := make(map[string]any)
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]any)
	}
	obj["allow_opt_out"] = flattenBoolValue(in.AllowOptOut)
	obj["max_allowed_duration"] = in.MaxAllowedDuration
	obj["max_allowed_times"] = in.MaxAllowedTimes
	v, _ := obj["approval"].([]any)
	obj["approval"] = flattenCustomProviderOptions(in.Approval, v)
	return []any{obj}
}

func flattenCadence(in *eaaspb.ScheduleOptions, p []any) []any {
	if in == nil {
		return nil
	}

	obj := make(map[string]any)
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]any)
	}
	obj["cron_expression"] = in.CronExpression
	obj["cron_timezone"] = in.CronTimezone
	obj["time_to_live"] = in.TimeToLive
	v, _ := obj["staggered"].([]any)
	obj["staggered"] = flattenStaggered(in.Staggered, v)
	return []any{obj}
}

func flattenStaggered(in *eaaspb.Staggered, p []any) []any {
	if in == nil {
		return nil
	}

	obj := make(map[string]any)
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]any)
	}
	obj["enabled"] = flattenBoolValue(in.Enabled)
	obj["max_interval"] = in.MaxInterval
	return []any{obj}
}

func flattenEnvironmentHooks(in *eaaspb.EnvironmentHooks, p []any) []any {
	log.Println("flatten environment hooks start")
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

	return []any{obj}
}

func flattenEnvironmentResources(input []*eaaspb.EnvironmentResourceCompoundRef, p []any) []any {
	log.Println("flatten environment resources start")
	if len(input) == 0 {
		return nil
	}

	out := make([]any, len(input))
	for i, in := range input {
		log.Println("flatten environment resource ", in)
		obj := map[string]any{}
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]any)
		}
		obj["type"] = in.Type
		obj["kind"] = in.Kind
		obj["name"] = in.Name
		obj["resource_options"] = flattenResourceOptions(in.ResourceOptions)

		v, _ := obj["depends_on"].([]any)
		obj["depends_on"] = flattenDependsOn(in.DependsOn, v)

		out[i] = &obj
	}

	return out
}

func flattenResourceOptions(input *eaaspb.EnvironmentResourceOptions) []any {
	if input == nil {
		return nil
	}
	obj := map[string]any{
		"dedicated": input.Dedicated,
		"version":   input.Version,
	}
	return []any{obj}
}

func flattenDependsOn(input []*commonpb.ResourceNameAndVersionRef, p []any) []any {
	log.Println("flatten dependson start")
	if input == nil {
		return nil
	}

	out := make([]any, len(input))
	for i, in := range input {
		log.Println("flatten depends on ", in)
		obj := map[string]any{}
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]any)
		}
		obj["name"] = in.Name
		obj["version"] = in.Version
		out[i] = &obj
	}

	return out
}

func resourceEnvironmentTemplateImport(d *schema.ResourceData, m any) ([]*schema.ResourceData, error) {

	log.Printf("Environment Template Import Starts")

	idParts := strings.SplitN(d.Id(), "/", 2)
	log.Println("resourceEnvironmentTemplateImport idParts:", idParts)

	log.Println("resourceEnvironmentTemplateImport Invoking expandEnvironmentTemplate")
	et, err := expandEnvironmentTemplate(d)
	if err != nil {
		log.Printf("resourceEnvironmentTemplateImport  expand error %s", err.Error())
	}

	var metaD commonpb.Metadata
	metaD.Name = idParts[0]
	metaD.Project = idParts[1]
	et.Metadata = &metaD

	err = d.Set("metadata", flattenV3MetaData(&metaD))
	if err != nil {
		log.Println("import set metadata err ", err)
		return nil, err
	}
	d.SetId(et.Metadata.Name)
	return []*schema.ResourceData{d}, nil

}
