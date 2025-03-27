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

func resourceEnvironmentTemplateCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Println("environment template create")
	diags := environmentTemplateUpsert(ctx, d, m)

	// Note: No need to delete the environment template object because upsert is atomic
	// otherwise if version creation fails, entire object get deleted

	// if diags.HasError() {
	// 	tflog := os.Getenv("TF_LOG")
	// 	if tflog == "TRACE" || tflog == "DEBUG" {
	// 		ctx = context.WithValue(ctx, "debug", "true")
	// 	}
	// 	environmenttemplate, err := expandEnvironmentTemplate(d)
	// 	if err != nil {
	// 		return diags
	// 	}
	// 	auth := config.GetConfig().GetAppAuthProfile()
	// 	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, TF_USER_AGENT, options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	// 	if err != nil {
	// 		return diags
	// 	}

	// 	err = client.EaasV1().EnvironmentTemplate().Delete(ctx, options.DeleteOptions{
	// 		Name:    environmenttemplate.Metadata.Name,
	// 		Project: environmenttemplate.Metadata.Project,
	// 	})
	// 	if err != nil {
	// 		return diags
	// 	}
	// }
	return diags
}

func environmentTemplateUpsert(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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

func resourceEnvironmentTemplateRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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

	if !et.GetSpec().GetSharing().GetEnabled() && environmenttemplate.GetSpec().GetSharing() == nil {
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

func resourceEnvironmentTemplateUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return environmentTemplateUpsert(ctx, d, m)
}

func resourceEnvironmentTemplateDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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

	if v, ok := in.Get("metadata").([]interface{}); ok && len(v) > 0 {
		obj.Metadata = expandV3MetaData(v)
	}

	if v, ok := in.Get("spec").([]interface{}); ok && len(v) > 0 {
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

func expandEnvironmentTemplateSpec(p []interface{}) (*eaaspb.EnvironmentTemplateSpec, error) {
	log.Println("expand environment template spec")
	spec := &eaaspb.EnvironmentTemplateSpec{}
	if len(p) == 0 || p[0] == nil {
		return spec, fmt.Errorf("%s", "expand environment template spec empty input")
	}

	in := p[0].(map[string]interface{})

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
	if p, ok := in["resources"].([]interface{}); ok && len(p) > 0 {
		spec.Resources, err = expandEnvironmentResources(p)
		if err != nil {
			return nil, err
		}
	}

	if v, ok := in["variables"].([]interface{}); ok && len(v) > 0 {
		spec.Variables = expandVariables(v)
	}

	if h, ok := in["hooks"].([]interface{}); ok && len(h) > 0 {
		spec.Hooks, err = expandEnvironmentHooks(h)
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

	if v, ok := in["contexts"].([]interface{}); ok && len(v) > 0 {
		spec.Contexts = expandContexts(v)
	}

	if v, ok := in["agent_override"].([]interface{}); ok && len(v) > 0 {
		spec.AgentOverride = expandEaasAgentOverrideOptions(v)
	}

	if s, ok := in["schedules"].([]interface{}); ok && len(s) > 0 {
		spec.Schedules, err = expandSchedules(s)
		if err != nil {
			return nil, err
		}
	}

	if v, ok := in["allow_new_inputs_during_publish"].([]interface{}); ok && len(v) > 0 {
		spec.AllowNewInputsDuringPublish = expandBoolValue(v)
	}

	if s, ok := in["actions"].([]interface{}); ok && len(s) > 0 {
		spec.Actions, err = expandActions(s)
		if err != nil {
			return nil, err
		}
	}

	return spec, nil
}

func expandSchedules(p []interface{}) ([]*eaaspb.Schedules, error) {
	schds := make([]*eaaspb.Schedules, 0)
	if len(p) == 0 || p[0] == nil {
		return schds, nil
	}
	var err error

	for i := range p {
		schd := eaaspb.Schedules{}
		in := p[i].(map[string]interface{})

		if v, ok := in["name"].(string); ok && len(v) > 0 {
			schd.Name = v
		}

		if v, ok := in["description"].(string); ok && len(v) > 0 {
			schd.Description = v
		}

		if v, ok := in["type"].(string); ok && len(v) > 0 {
			schd.Type = v
		}

		if v, ok := in["cadence"].([]interface{}); ok && len(v) > 0 {
			schd.Cadence = expandCadence(v)
		}

		if v, ok := in["context"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
			schd.Context = expandConfigContextCompoundRef(v[0].(map[string]any))
		}

		if v, ok := in["opt_out_options"].([]interface{}); ok && len(v) > 0 {
			schd.OptOutOptions, err = expandOptOutOptions(v)
			if err != nil {
				return nil, err
			}
		}

		if h, ok := in["workflows"].([]interface{}); ok && len(h) > 0 {
			schd.Workflows, err = expandCustomProviderOptions(h)
			if err != nil {
				return nil, err
			}
		}

		schds = append(schds, &schd)
	}

	return schds, nil
}

func expandOptOutOptions(p []interface{}) (*eaaspb.OptOutOptions, error) {
	ooo := eaaspb.OptOutOptions{}
	if len(p) == 0 || p[0] == nil {
		return &ooo, nil
	}

	var err error
	in := p[0].(map[string]interface{})
	if h, ok := in["allow_opt_out"].([]interface{}); ok && len(h) > 0 {
		ooo.AllowOptOut = expandBoolValue(h)
	}
	if v, ok := in["max_allowed_duration"].(string); ok && len(v) > 0 {
		ooo.MaxAllowedDuration = v
	}
	if v, ok := in["max_allowed_times"].(int); ok {
		ooo.MaxAllowedTimes = int32(v)
	}
	if h, ok := in["approval"].([]interface{}); ok && len(h) > 0 {
		ooo.Approval, err = expandCustomProviderOptions(h)
		if err != nil {
			return nil, err
		}
	}

	return &ooo, nil
}

func expandCadence(p []interface{}) *eaaspb.ScheduleOptions {
	cadence := eaaspb.ScheduleOptions{}
	if len(p) == 0 || p[0] == nil {
		return &cadence
	}

	in := p[0].(map[string]interface{})
	if v, ok := in["cron_expression"].(string); ok && len(v) > 0 {
		cadence.CronExpression = v
	}

	if v, ok := in["cron_timezone"].(string); ok && len(v) > 0 {
		cadence.CronTimezone = v
	}

	if v, ok := in["time_to_live"].(string); ok && len(v) > 0 {
		cadence.TimeToLive = v
	}

	return &cadence
}

func expandEaasAgentOverrideOptions(p []interface{}) *eaaspb.AgentOverrideOptions {
	agentOverrideOptions := &eaaspb.AgentOverrideOptions{}
	if len(p) == 0 || p[0] == nil {
		return agentOverrideOptions
	}

	in := p[0].(map[string]interface{})
	if v, ok := in["required"].(bool); ok {
		agentOverrideOptions.Required = v
	}

	if aot, ok := in["type"].(string); ok {
		agentOverrideOptions.Type = aot
	}

	if agnts, ok := in["restricted_agents"].([]interface{}); ok && len(agnts) > 0 {
		agentOverrideOptions.RestrictedAgents = toArrayString(agnts)
	}

	return agentOverrideOptions
}

func expandEnvironmentResources(p []interface{}) ([]*eaaspb.EnvironmentResourceCompoundRef, error) {
	log.Println("expand environment resources")
	if len(p) == 0 || p[0] == nil {
		return nil, fmt.Errorf("%s", "expand environment resources empty input")
	}

	envresources := make([]*eaaspb.EnvironmentResourceCompoundRef, len(p))

	for i := range p {
		obj := eaaspb.EnvironmentResourceCompoundRef{}
		in := p[i].(map[string]interface{})

		if v, ok := in["type"].(string); ok && len(v) > 0 {
			obj.Type = v
		}

		if v, ok := in["kind"].(string); ok && len(v) > 0 {
			obj.Kind = v
		}

		if v, ok := in["name"].(string); ok && len(v) > 0 {
			obj.Name = v
		}

		if v, ok := in["resource_options"].([]interface{}); ok && len(v) > 0 {
			obj.ResourceOptions = expandResourceOptions(v)
		}

		if v, ok := in["depends_on"].([]interface{}); ok && len(v) > 0 {
			obj.DependsOn = expandDependsOn(v)
		}

		envresources[i] = &obj

	}

	return envresources, nil

}

func expandResourceOptions(p []interface{}) *eaaspb.EnvironmentResourceOptions {
	ro := &eaaspb.EnvironmentResourceOptions{}

	if len(p) == 0 || p[0] == nil {
		return ro
	}

	in := p[0].(map[string]interface{})

	if dedicated, ok := in["dedicated"].(bool); ok {
		ro.Dedicated = dedicated
	}

	if version, ok := in["version"].(string); ok {
		ro.Version = version
	}

	return ro
}

func expandDependsOn(p []interface{}) []*commonpb.ResourceNameAndVersionRef {
	dependson := make([]*commonpb.ResourceNameAndVersionRef, 0)
	if len(p) == 0 {
		return dependson
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

		dependson = append(dependson, obj)
	}

	return dependson
}

func expandEnvironmentHooks(p []interface{}) (*eaaspb.EnvironmentHooks, error) {
	hooks := &eaaspb.EnvironmentHooks{}

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

	v, ok := d.Get("spec").([]interface{})
	if !ok {
		v = []interface{}{}
	}

	var ret []interface{}
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

func flattenEnvironmentTemplateSpec(in *eaaspb.EnvironmentTemplateSpec, p []interface{}) ([]interface{}, error) {
	if in == nil {
		return nil, nil
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}
	obj["version"] = in.Version
	obj["version_state"] = in.VersionState
	obj["icon_url"] = in.IconURL
	obj["readme"] = in.Readme
	obj["resources"] = flattenEnvironmentResources(in.Resources, obj["resources"].([]interface{}))
	obj["variables"] = flattenVariables(in.Variables, obj["variables"].([]interface{}))
	obj["hooks"] = flattenEnvironmentHooks(in.Hooks, obj["hooks"].([]interface{}))
	obj["agents"] = flattenEaasAgents(in.Agents)
	obj["sharing"] = flattenSharingSpec(in.Sharing)
	obj["contexts"] = flattenContexts(in.Contexts, obj["contexts"].([]interface{}))
	obj["agent_override"] = flattenEaasAgentOverrideOptions(in.AgentOverride)
	obj["schedules"] = flattenSchedules(in.Schedules, obj["schedules"].([]interface{}))
	obj["allow_new_inputs_during_publish"] = flattenBoolValue(in.AllowNewInputsDuringPublish)
	obj["actions"] = flattenActions(in.Actions, obj["actions"].([]interface{}))
	return []interface{}{obj}, nil
}

func flattenEaasAgentOverrideOptions(in *eaaspb.AgentOverrideOptions) []interface{} {
	if in == nil {
		return nil
	}

	obj := make(map[string]interface{})
	obj["required"] = in.Required
	obj["type"] = in.Type
	obj["restricted_agents"] = toArrayInterface(in.RestrictedAgents)

	return []interface{}{obj}
}

func flattenSchedules(input []*eaaspb.Schedules, p []interface{}) []interface{} {
	log.Println("flatten schedules start")
	if input == nil {
		return nil
	}

	out := make([]interface{}, len(input))
	for i, in := range input {
		log.Println("flatten schedule ", in)
		obj := map[string]interface{}{}
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}

		if len(in.Name) > 0 {
			obj["name"] = in.Name
		}

		if len(in.Description) > 0 {
			obj["description"] = in.Description
		}

		if len(in.Type) > 0 {
			obj["type"] = in.Type
		}

		if in.Cadence != nil {
			v, ok := obj["cadence"].([]interface{})
			if !ok {
				v = []interface{}{}
			}
			obj["cadence"] = flattenCadence(in.Cadence, v)
		}
		if in.Context != nil {
			cc := flattenConfigContextCompoundRef(in.Context)
			obj["context"] = []interface{}{cc}
		}

		if in.OptOutOptions != nil {
			v, ok := obj["opt_out_options"].([]interface{})
			if !ok {
				v = []interface{}{}
			}
			obj["opt_out_options"] = flattenOptOutOptions(in.OptOutOptions, v)
		}
		obj["workflows"] = flattenCustomProviderOptions(in.Workflows)

		out[i] = &obj
	}

	return out
}

func flattenOptOutOptions(in *eaaspb.OptOutOptions, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}

	obj := make(map[string]interface{})
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	obj["allow_opt_out"] = flattenBoolValue(in.AllowOptOut)
	obj["max_allowed_duration"] = in.MaxAllowedDuration
	obj["max_allowed_times"] = in.MaxAllowedTimes
	obj["approval"] = flattenCustomProviderOptions(in.Approval)

	return []interface{}{obj}
}

func flattenCadence(in *eaaspb.ScheduleOptions, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}

	obj := make(map[string]interface{})
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.CronExpression) > 0 {
		obj["cron_expression"] = in.CronExpression
	}

	if len(in.CronTimezone) > 0 {
		obj["cron_timezone"] = in.CronTimezone
	}

	if len(in.TimeToLive) > 0 {
		obj["time_to_live"] = in.TimeToLive
	}

	return []interface{}{obj}
}

func flattenEnvironmentHooks(in *eaaspb.EnvironmentHooks, p []interface{}) []interface{} {
	log.Println("flatten environment hooks start")
	if in == nil {
		return nil
	}

	obj := make(map[string]interface{})
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}
	obj["on_completion"] = flattenEaasHooks(in.OnCompletion, obj["on_completion"].([]interface{}))
	obj["on_success"] = flattenEaasHooks(in.OnSuccess, obj["on_success"].([]interface{}))
	obj["on_failure"] = flattenEaasHooks(in.OnFailure, obj["on_failure"].([]interface{}))
	obj["on_init"] = flattenEaasHooks(in.OnInit, obj["on_init"].([]interface{}))
	return []interface{}{obj}
}

func flattenEnvironmentResources(input []*eaaspb.EnvironmentResourceCompoundRef, p []interface{}) []interface{} {
	log.Println("flatten environment resources start")
	if len(input) == 0 {
		return nil
	}

	out := make([]interface{}, len(input))
	for i, in := range input {
		log.Println("flatten environment resource ", in)
		obj := map[string]interface{}{}
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}

		if len(in.Type) > 0 {
			obj["type"] = in.Type
		}

		if len(in.Kind) > 0 {
			obj["kind"] = in.Kind
		}

		if len(in.Name) > 0 {
			obj["name"] = in.Name
		}

		if in.ResourceOptions != nil {
			obj["resource_options"] = flattenResourceOptions(in.ResourceOptions)
		}

		if len(in.DependsOn) > 0 {
			v, ok := obj["depends_on"].([]interface{})
			if !ok {
				v = []interface{}{}
			}

			obj["depends_on"] = flattenDependsOn(in.DependsOn, v)
		}

		out[i] = &obj
	}

	return out
}

func flattenResourceOptions(input *eaaspb.EnvironmentResourceOptions) []interface{} {
	if input == nil {
		return nil
	}

	obj := map[string]interface{}{}
	obj["dedicated"] = input.Dedicated
	obj["version"] = input.Version

	return []interface{}{obj}
}

func flattenDependsOn(input []*commonpb.ResourceNameAndVersionRef, p []interface{}) []interface{} {
	log.Println("flatten dependson start")
	if input == nil {
		return nil
	}

	out := make([]interface{}, len(input))
	for i, in := range input {
		log.Println("flatten depends on ", in)
		obj := map[string]interface{}{}
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}

		if len(in.Name) > 0 {
			obj["name"] = in.Name
		}

		if len(in.Version) > 0 {
			obj["version"] = in.Version
		}

		out[i] = &obj
	}

	return out
}

func resourceEnvironmentTemplateImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {

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
