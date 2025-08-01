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
			Create: schema.DefaultTimeout(8 * time.Hour),
			Update: schema.DefaultTimeout(8 * time.Hour),
			Delete: schema.DefaultTimeout(8 * time.Hour),
		},

		SchemaVersion: 1,
		Schema:        resource.EnvironmentSchema.Schema,
	}
}

func resourceEnvironmentCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
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

func environmentUpsert(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
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

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	// wait for publish
	for {
		select {
		case <-ticker.C:
		case <-ctx.Done():
			return diag.FromErr(fmt.Errorf("context cancelled"))
		}
		envs, err := client.EaasV1().Environment().Status(ctx, options.StatusOptions{
			Name:    environment.Metadata.Name,
			Project: environment.Metadata.Project,
		})
		if err != nil {
			return diag.FromErr(err)
		}
		if envs.GetStatus() != nil {
			//check if environment provisioning, if true break out of loop
			if status := envs.GetStatus().GetDigestedStatus().GetConditionStatus(); status == commonpb.ConditionStatus_StatusOK ||
				status == commonpb.ConditionStatus_StatusNotSet {
				break
			}
			if envs.GetStatus().GetDigestedStatus().GetConditionStatus() == commonpb.ConditionStatus_StatusFailed {
				return diag.FromErr(fmt.Errorf("%s %s", "failed to publish environment", envs.GetStatus().GetDigestedStatus().GetReason()))
			}
			if envs.GetStatus().GetDigestedStatus().GetConditionStatus() == commonpb.ConditionStatus_StatusSubmitted {
				if strings.Contains(envs.GetStatus().GetDigestedStatus().GetReason(), "trigger not processed") {
					return diag.FromErr(fmt.Errorf("%s %s", "failed to publish environment", envs.GetStatus().GetLatestEvents()[0].GetTriggerDetails().GetReason()))
				}
			}
		} else {
			break
		}

	}

	d.SetId(environment.Metadata.Name)
	return diags
}

func resourceEnvironmentRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
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

	if et.GetSpec().GetSharing() != nil && !et.GetSpec().GetSharing().GetEnabled() && environment.GetSpec().GetSharing() == nil {
		environment.Spec.Sharing = &commonpb.SharingSpec{}
		environment.Spec.Sharing.Enabled = false
		environment.Spec.Sharing.Projects = et.GetSpec().GetSharing().GetProjects()
	}

	err = flattenEnvironment(d, environment)
	if err != nil {
		log.Println("read flatten err")
		return diag.FromErr(err)
	}
	return diags

}

func resourceEnvironmentUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	return environmentUpsert(ctx, d, m)
}

func resourceEnvironmentDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
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

	// wait for destroy
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	// wait for publish
	for {
		select {
		case <-ticker.C:
		case <-ctx.Done():
			return diag.FromErr(fmt.Errorf("context cancelled"))
		}
		envs, err := client.EaasV1().Environment().Status(ctx, options.StatusOptions{
			Name:    env.Metadata.Name,
			Project: env.Metadata.Project,
		})
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				log.Printf("environment %s deleted successfully", env.Metadata.Name)
				return diags
			}

			log.Printf("environment %s deletion failed with error: %s", env.Metadata.Name, err.Error())
			return diag.FromErr(err)
		}

		if envs.GetStatus() != nil {
			//check if environment provisioning, if true break out of loop
			if status := envs.GetStatus().GetDigestedStatus().GetConditionStatus(); status == commonpb.ConditionStatus_StatusOK ||
				status == commonpb.ConditionStatus_StatusNotSet {
				break
			}
			if envs.GetStatus().GetDigestedStatus().GetConditionStatus() == commonpb.ConditionStatus_StatusFailed {
				return diag.FromErr(fmt.Errorf("%s %s", "failed to destroy environment", envs.GetStatus().GetDigestedStatus().GetReason()))
			}
		} else {
			break
		}

	}

	return diags
}

func expandEnvironment(in *schema.ResourceData) (*eaaspb.Environment, error) {
	log.Println("expand environment")
	if in == nil {
		return nil, fmt.Errorf("%s", "expand environment empty input")
	}
	obj := &eaaspb.Environment{}

	if v, ok := in.Get("metadata").([]any); ok && len(v) > 0 {
		obj.Metadata = expandV3MetaData(v)
	}

	if v, ok := in.Get("spec").([]any); ok && len(v) > 0 {
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

func expandEnvironmentSpec(p []any) (*eaaspb.EnvironmentSpec, error) {
	log.Println("expand environment spec")
	spec := &eaaspb.EnvironmentSpec{}
	if len(p) == 0 || p[0] == nil {
		return spec, fmt.Errorf("%s", "expand environment spec empty input")
	}

	in := p[0].(map[string]any)

	var err error
	if v, ok := in["template"].([]any); ok && len(v) > 0 {
		spec.Template, err = expandTemplate(v)
		if err != nil {
			return spec, err
		}
	}

	if v, ok := in["variables"].([]any); ok && len(v) > 0 {
		spec.Variables = expandVariables(v)
	}

	if v, ok := in["sharing"].([]any); ok && len(v) > 0 {
		spec.Sharing = expandSharingSpec(v)
	}

	if ag, ok := in["agents"].([]any); ok && len(ag) > 0 {
		spec.Agents = expandEaasAgents(ag)
	}

	if ev, ok := in["env_vars"].([]any); ok && len(ev) > 0 {
		spec.EnvVars = expandEnvVariables(ev)
	}

	if f, ok := in["files"].([]any); ok && len(f) > 0 {
		spec.Files = expandCommonpbFiles(f)
	}

	if so, ok := in["schedule_optouts"].([]any); ok && len(so) > 0 {
		spec.ScheduleOptouts = expandScheduleOptOuts(so)
	}

	if rr, ok := in["reconcile_resources"].([]any); ok && len(rr) > 0 {
		spec.ReconcileResources = expandReconcileResources(rr)
	}

	return spec, nil
}

func expandTemplate(p []any) (*eaaspb.EnvironmentTemplateCompoundRef, error) {
	log.Println("expand template")
	if len(p) == 0 || p[0] == nil {
		return nil, fmt.Errorf("%s", "expand template empty input")
	}

	obj := &eaaspb.EnvironmentTemplateCompoundRef{}

	in := p[0].(map[string]any)

	if v, ok := in["name"].(string); ok && len(v) > 0 {
		obj.Name = v
	}

	if v, ok := in["version"].(string); ok && len(v) > 0 {
		obj.Version = v
	}

	return obj, nil
}

func expandScheduleOptOuts(p []any) []*eaaspb.ScheduleOptOut {
	soo := make([]*eaaspb.ScheduleOptOut, 0)
	if len(p) == 0 || p[0] == nil {
		return soo
	}

	for i := range p {
		obj := eaaspb.ScheduleOptOut{}
		in := p[i].(map[string]any)

		if v, ok := in["name"].(string); ok && len(v) > 0 {
			obj.Name = v
		}

		if v, ok := in["duration"].(string); ok && len(v) > 0 {
			obj.Duration = v
		}

		soo = append(soo, &obj)

	}

	return soo
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

	v, ok := d.Get("spec").([]any)
	if !ok {
		v = []any{}
	}

	var ret []any
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

func flattenEnvironmentSpec(in *eaaspb.EnvironmentSpec, p []any) ([]any, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "flatten environment spec empty input")
	}

	obj := map[string]any{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]any)
	}

	v, _ := obj["template"].([]any)
	obj["template"] = flattenTemplate(in.Template, v)

	v, _ = obj["variables"].([]any)
	obj["variables"] = flattenVariables(in.Variables, v)

	obj["sharing"] = flattenSharingSpec(in.Sharing)
	obj["agents"] = flattenEaasAgents(in.Agents)

	v, _ = obj["env_vars"].([]any)
	obj["env_vars"] = flattenEnvVariables(in.EnvVars, v)

	obj["files"] = flattenCommonpbFiles(in.Files)

	v, _ = obj["schedule_optouts"].([]any)
	obj["schedule_optouts"] = flattenScheduleOptOuts(in.ScheduleOptouts, v)

	obj["reconcile_resources"] = flattenReconcileResources(in.ReconcileResources)
	return []any{obj}, nil
}

func flattenTemplate(input *eaaspb.EnvironmentTemplateCompoundRef, p []any) []any {
	log.Println("flatten template start", input)
	if input == nil {
		return nil
	}

	obj := map[string]any{}
	if len(p) > 0 && p[0] != nil {
		obj = p[0].(map[string]any)
	}
	obj["name"] = input.Name
	obj["version"] = input.Version
	return []any{obj}
}

func flattenScheduleOptOuts(input []*eaaspb.ScheduleOptOut, p []any) []any {
	log.Println("flatten schedule optout start")
	if input == nil {
		return nil
	}

	out := make([]any, len(input))
	for i, in := range input {
		log.Println("flatten schedule optout ", in)
		obj := map[string]any{}
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]any)
		}
		obj["name"] = in.Name
		obj["duration"] = in.Duration
		out[i] = &obj
	}

	return out
}

func resourceEnvironmentImport(d *schema.ResourceData, m any) ([]*schema.ResourceData, error) {

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
