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
	"github.com/RafaySystems/rafay-common/proto/types/hub/infrapb"
	"github.com/RafaySystems/rctl/pkg/config"
	"github.com/RafaySystems/rctl/pkg/versioninfo"
	"github.com/davecgh/go-spew/spew"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"google.golang.org/protobuf/types/known/structpb"
)

func resourceFleetPlan() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceFleetPlanCreate,
		UpdateContext: resourceFleetPlanUpdate,
		ReadContext:   resourceFleetPlanRead,
		DeleteContext: resourceFleetPlanDelete,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		SchemaVersion: 1,
		Schema:        resource.FleetPlanSchema.Schema,
	}
}

func resourceFleetPlanCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceFleetPlanUpsert(context.Background(), d)
}

func resourceFleetPlanUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceFleetPlanUpsert(context.Background(), d)
}

func resourceFleetPlanUpsert(ctx context.Context, d *schema.ResourceData) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("fleetplan upsert starts")
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

	fleetplan, err := expandFleetPlan(d)
	if err != nil {
		log.Printf("fleetplan expandFleetPlan error")
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.InfraV3().FleetPlan().Apply(ctx, fleetplan, options.ApplyOptions{})
	if err != nil {
		if strings.Contains(err.Error(), "Updated FleetPlan but could not execute the job") {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  err.Error(),
			})
		} else {
			n1 := spew.Sprintf("%+v", fleetplan)
			log.Println("fleetplan apply fleetplan:", n1)
			log.Printf("fleetplan apply error: %v", err)
			return diag.FromErr(err)
		}
	}

	d.SetId(fleetplan.Metadata.Name)

	return diags
}

func resourceFleetPlanRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}

	meta := GetMetaData(d)
	if meta == nil || meta.Name == "" {
		return diag.FromErr(fmt.Errorf("%s", "failed to read resource "))
	}

	fleetplan, err := client.InfraV3().FleetPlan().Get(ctx, options.GetOptions{
		Name:    meta.Name,
		Project: meta.Project,
	})

	if err != nil {
		return diag.FromErr(err)
	}

	if fleetplan == nil {
		d.SetId("")
		return diags
	}

	d.SetId(fleetplan.Metadata.Name)
	err = d.Set("metadata", flattenMetaData(fleetplan.Metadata))
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("spec", flattenFleetPlanSpec(fleetplan.Spec))
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceFleetPlanDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	var diags diag.Diagnostics
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	fp, err := expandFleetPlan(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent(), options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.InfraV3().FleetPlan().Delete(ctx, options.DeleteOptions{
		Name:    fp.Metadata.Name,
		Project: fp.Metadata.Project,
	})

	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func flattenFleetPlanSpec(spec *infrapb.FleetPlanSpec) []interface{} {
	if spec == nil {
		return []interface{}{}
	}

	obj := make(map[string]interface{})
	obj["fleet"] = flattenFleet(spec.Fleet)
	obj["operation_workflow"] = flattenOperationWorkflow(spec.OperationWorkflow)
	obj["agents"] = flattenFleetPlanAgents(spec.Agents)

	return []interface{}{obj}
}

func flattenFleet(fs *infrapb.FleetSpec) []interface{} {
	if fs == nil {
		return []interface{}{}
	}

	obj := make(map[string]interface{})
	obj["kind"] = fs.Kind
	obj["labels"] = fs.Labels
	obj["projects"] = flattenProjects(fs.Projects)
	return []interface{}{obj}
}

func flattenProjects(projects []*infrapb.ProjectFilter) []interface{} {
	if projects == nil {
		return []interface{}{}
	}
	obj := make([]interface{}, len(projects))
	for i, v := range projects {
		obj[i] = map[string]interface{}{
			"name": v.Name,
		}
	}
	return obj
}

func flattenOperationWorkflow(workflow *infrapb.OperationWorkflowSpec) []interface{} {
	if workflow == nil {
		return []interface{}{}
	}
	obj := make(map[string]interface{})
	obj["operations"] = flattenOperations(workflow.Operations)
	return []interface{}{obj}
}

func flattenOperations(operations []*infrapb.OperationSpec) []interface{} {
	if operations == nil {
		return []interface{}{}
	}
	obj := make([]interface{}, len(operations))
	for i, v := range operations {
		obj[i] = flattenOperation(v)
	}
	return obj
}

func flattenOperation(operation *infrapb.OperationSpec) map[string]interface{} {
	if operation == nil {
		return map[string]interface{}{}
	}
	obj := make(map[string]interface{})
	obj["action"] = flattenAction(operation.Action)
	obj["prehooks"] = flattenHooks(operation.Prehooks)
	obj["posthooks"] = flattenHooks(operation.Posthooks)
	return obj
}

func flattenAction(action *infrapb.ActionSpec) []interface{} {
	if action == nil {
		return []interface{}{}
	}
	obj := make(map[string]interface{})
	obj["name"] = action.Name
	obj["type"] = action.Type
	obj["description"] = action.Description
	obj["control_plane_upgrade_config"] = flattenControlPlaneUpgradeConfig(action.ControlPlaneUpgradeConfig)
	obj["node_groups_and_control_plane_upgrade_config"] = flattenNodeGroupsAndControlPlaneUpgradeConfig(action.NodeGroupsAndControlPlaneUpgradeConfig)
	obj["node_groups_upgrade_config"] = flattenNodeGroupsUpgradeConfig(action.NodeGroupsUpgradeConfig)
	// obj["patch_config"] = flattenPatchConfig(action.PatchConfig)

	return []interface{}{obj}
}

func flattenNodeGroupsUpgradeConfig(config *infrapb.NodeGroupsUpgradeConfigSpec) []interface{} {
	if config == nil {
		return []interface{}{}
	}
	obj := make(map[string]interface{})
	obj["names"] = config.Names
	obj["version"] = config.Version
	return []interface{}{obj}
}

func flattenNodeGroupsAndControlPlaneUpgradeConfig(config *infrapb.NodeGroupsAndControlPlaneUpgradeConfigSpec) []interface{} {
	if config == nil {
		return []interface{}{}
	}
	obj := make(map[string]interface{})
	obj["version"] = config.Version
	return []interface{}{obj}
}

func flattenControlPlaneUpgradeConfig(config *infrapb.ControlPlaneUpgradeConfigSpec) []interface{} {
	if config == nil {
		return []interface{}{}
	}
	obj := make(map[string]interface{})
	obj["version"] = config.Version
	return []interface{}{obj}
}

func flattenFleetPlanAgents(agents []*infrapb.Agent) []interface{} {
	if agents == nil {
		return []interface{}{}
	}
	obj := make([]interface{}, len(agents))
	for i, v := range agents {
		obj[i] = flattenFleetPlanAgent(v)
	}
	return obj
}

func flattenFleetPlanAgent(agent *infrapb.Agent) map[string]interface{} {
	if agent == nil {
		return make(map[string]interface{})
	}
	obj := make(map[string]interface{})

	obj["name"] = agent.Name

	return obj
}

func flattenHooks(hooks []*infrapb.HookSpec) []interface{} {
	if hooks == nil {
		return []interface{}{}
	}
	obj := make([]interface{}, len(hooks))
	for i, v := range hooks {
		obj[i] = flattenHook(v)
	}
	return obj
}

func flattenHook(hook *infrapb.HookSpec) map[string]interface{} {
	if hook == nil {
		return make(map[string]interface{})
	}
	obj := make(map[string]interface{})
	obj["description"] = hook.Description
	obj["inject"] = hook.Inject
	obj["name"] = hook.Name
	obj["container_config"] = flattenContainerConfig(hook.ContainerConfig)
	obj["http_config"] = flattenHTTPConfig(hook.HttpConfig)
	return obj
}

func flattenContainerConfig(config *infrapb.ContainerConfigSpec) []interface{} {
	if config == nil {
		return []interface{}{}
	}
	obj := make(map[string]interface{})
	obj["runner"] = config.Runner
	obj["image"] = config.Image
	if config.Env != nil {
		obj["env"] = config.Env
	}
	if config.Arguments != nil {
		obj["arguments"] = config.Arguments
	}
	if config.Commands != nil {
		obj["commands"] = config.Commands
	}

	return []interface{}{obj}
}

func flattenHTTPConfig(config *infrapb.HttpConfigSpec) []interface{} {
	if config == nil {
		return []interface{}{}
	}
	obj := make(map[string]interface{})

	obj["endpoint"] = config.Endpoint
	obj["method"] = config.Method
	if config.Headers != nil {
		obj["headers"] = config.Headers
	}
	obj["body"] = config.Body

	return []interface{}{obj}

}

func expandFleetPlan(d *schema.ResourceData) (*infrapb.FleetPlan, error) {
	if d == nil {
		return nil, fmt.Errorf("%s", "expand fleetplan empty input")
	}
	obj := &infrapb.FleetPlan{}

	if v, ok := d.Get("metadata").([]interface{}); ok && len(v) > 0 {
		obj.Metadata = expandMetaData(v)
	}

	if v, ok := d.Get("spec").([]interface{}); ok && len(v) > 0 {
		objSpec, err := expandFleetPlanSpec(v)
		if err != nil {
			return nil, err
		}
		log.Println("expandFleetPlanSpec got spec")
		obj.Spec = objSpec
	}

	obj.ApiVersion = "infra.k8smgmt.io/v3"
	obj.Kind = "FleetPlan"
	return obj, nil
}

func expandFleetPlanSpec(p []interface{}) (*infrapb.FleetPlanSpec, error) {
	obj := &infrapb.FleetPlanSpec{}
	if len(p) == 0 || p[0] == nil {
		return nil, fmt.Errorf("%s", "expandFleetPlanSpec empty input")
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["fleet"].([]interface{}); ok && len(v) > 0 {
		obj.Fleet = expandFleetSpec(v)
	}

	if v, ok := in["operation_workflow"].([]interface{}); ok && len(v) > 0 {
		obj.OperationWorkflow = expandOperationWorkflow(v)
	}

	if v, ok := in["agents"].([]interface{}); ok && len(v) > 0 {
		obj.Agents = expandFleetPlanAgent(v)
	}

	return obj, nil

}

func expandFleetSpec(p []interface{}) *infrapb.FleetSpec {

	obj := &infrapb.FleetSpec{}
	if len(p) == 0 || p[0] == nil {
		return nil
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["kind"].(string); ok {
		obj.Kind = v
	}

	if v, ok := in["labels"].(map[string]interface{}); ok && len(v) > 0 {
		obj.Labels = toMapString(v)
	}

	if v, ok := in["projects"].([]interface{}); ok {
		obj.Projects = expandProjects(v)
	}
	return obj
}

func expandProjects(v []interface{}) []*infrapb.ProjectFilter {

	var projects []*infrapb.ProjectFilter

	for _, project := range v {
		project := project.(map[string]interface{})

		projects = append(projects, &infrapb.ProjectFilter{
			Name: project["name"].(string),
		})
	}

	return projects

}

func expandOperationWorkflow(v []interface{}) *infrapb.OperationWorkflowSpec {

	obj := &infrapb.OperationWorkflowSpec{}
	if len(v) == 0 || v[0] == nil {
		return nil
	}

	in := v[0].(map[string]interface{})

	if operations, ok := in["operations"]; ok && len(in) > 0 {
		obj.Operations = expandOperations(operations.([]interface{}))
	}

	return obj
}

func expandOperations(v []interface{}) []*infrapb.OperationSpec {
	var operations []*infrapb.OperationSpec

	for _, operation := range v {

		op := operation.(map[string]interface{})
		//TODO
		operations = append(operations, &infrapb.OperationSpec{
			Name:      op["name"].(string),
			Action:    expandAction(op["action"].([]interface{})),
			Prehooks:  expandHooks(op["prehooks"].([]interface{})),
			Posthooks: expandHooks(op["posthooks"].([]interface{})),
		})
	}
	return operations
}

func expandAction(in []interface{}) *infrapb.ActionSpec {
	obj := &infrapb.ActionSpec{}
	if len(in) == 0 || in[0] == nil {
		return nil
	}

	v := in[0].(map[string]interface{})

	if v, ok := v["name"].(string); ok {
		obj.Name = v
	}

	if v, ok := v["type"].(string); ok {
		obj.Type = v
	}

	if v, ok := v["description"].(string); ok {
		obj.Description = v
	}

	if v, ok := v["control_plane_upgrade_config"].([]interface{}); ok {
		obj.ControlPlaneUpgradeConfig = expandControlPlaneUpgradeConfig(v)
	}

	if v, ok := v["node_groups_and_control_plane_upgrade_config"].([]interface{}); ok {
		obj.NodeGroupsAndControlPlaneUpgradeConfig = expandNodeGroupsAndControlPlaneUpgradeConfig(v)
	}

	if v, ok := v["node_groups_upgrade_config"].([]interface{}); ok {
		obj.NodeGroupsUpgradeConfig = expandNodeGroupsUpgradeConfig(v)
	}

	if v, ok := v["patch_config"].([]interface{}); ok {
		obj.PatchConfig = expandPatchConfig(v)
	}

	return obj

}

func expandControlPlaneUpgradeConfig(in []interface{}) *infrapb.ControlPlaneUpgradeConfigSpec {
	obj := &infrapb.ControlPlaneUpgradeConfigSpec{}
	if len(in) == 0 || in[0] == nil {
		return nil
	}

	v := in[0].(map[string]interface{})
	if v, ok := v["version"].(string); ok {
		obj.Version = v
	}

	return obj
}

func expandNodeGroupsAndControlPlaneUpgradeConfig(in []interface{}) *infrapb.NodeGroupsAndControlPlaneUpgradeConfigSpec {
	obj := &infrapb.NodeGroupsAndControlPlaneUpgradeConfigSpec{}
	if len(in) == 0 || in[0] == nil {
		return nil
	}

	v := in[0].(map[string]interface{})
	if v, ok := v["version"].(string); ok {
		obj.Version = v
	}

	return obj
}

func expandNodeGroupsUpgradeConfig(in []interface{}) *infrapb.NodeGroupsUpgradeConfigSpec {
	obj := &infrapb.NodeGroupsUpgradeConfigSpec{}
	if len(in) == 0 || in[0] == nil {
		return nil
	}

	v := in[0].(map[string]interface{})
	if v, ok := v["version"].(string); ok {
		obj.Version = v
	}

	if v, ok := v["names"].([]interface{}); ok && len(v) > 0 {
		obj.Names = toArrayString(v)
	}

	return obj
}

func expandPatchConfig(in []interface{}) []*infrapb.PatchConfigSpec {
	obj := make([]*infrapb.PatchConfigSpec, 0)

	if len(in) == 0 || in[0] == nil {
		return nil
	}

	for _, patch := range in {

		p := patch.(map[string]interface{})
		sVal := &structpb.Value{}
		if valString, ok := p["value"].(string); ok {
			sVal = convertValStrToVal(valString)
		}
		var op, path string
		if v, ok := p["op"].(string); ok {
			op = v
		}
		if v, ok := p["path"].(string); ok {
			path = v
		}

		obj = append(obj, &infrapb.PatchConfigSpec{
			Op:    op,
			Path:  path,
			Value: sVal,
		})
	}

	return obj
}

func expandHooks(in []interface{}) []*infrapb.HookSpec {
	var outHooks []*infrapb.HookSpec

	for _, inHook := range in {
		outHook := &infrapb.HookSpec{}
		inH := inHook.(map[string]interface{})
		if val, ok := inH["name"].(string); ok {
			outHook.Name = val
		}
		if val, ok := inH["description"].(string); ok {
			outHook.Description = val
		}
		if val, ok := inH["inject"].([]interface{}); ok {
			outHook.Inject = toArrayString(val)
		}
		if val, ok := inH["container_config"]; ok {
			outHook.ContainerConfig = expandContainerConfig(val.([]interface{}))
		}
		if val, ok := inH["http_config"]; ok {
			outHook.HttpConfig = expandHttpConfig(val.([]interface{}))
		}

		outHooks = append(outHooks, outHook)

	}

	return outHooks
}

func expandContainerConfig(in []interface{}) *infrapb.ContainerConfigSpec {
	if len(in) == 0 || in[0] == nil {
		return nil
	}
	obj := &infrapb.ContainerConfigSpec{}

	v := in[0].(map[string]interface{})

	if v, ok := v["runner"].(string); ok {
		obj.Runner = v
	}

	if v, ok := v["image"].(string); ok {
		obj.Image = v
	}

	if v, ok := v["env"].(map[string]interface{}); ok && len(v) > 0 {
		obj.Env = toMapString(v)
	}

	if v, ok := v["arguments"].([]interface{}); ok {
		obj.Arguments = toArrayString(v)
	}

	if v, ok := v["commands"].([]interface{}); ok {
		obj.Commands = toArrayString(v)
	}

	return obj
}

func expandHttpConfig(in []interface{}) *infrapb.HttpConfigSpec {
	if len(in) == 0 || in[0] == nil {
		return nil
	}
	obj := &infrapb.HttpConfigSpec{}

	v := in[0].(map[string]interface{})

	if v, ok := v["endpoint"].(string); ok {
		obj.Endpoint = v
	}

	if v, ok := v["method"].(string); ok {
		obj.Method = v
	}

	if v, ok := v["headers"].(map[string]interface{}); ok && len(v) > 0 {
		obj.Headers = toMapString(v)
	}

	if v, ok := v["body"].(string); ok {
		obj.Body = v
	}

	return obj
}

func expandFleetPlanAgent(in []interface{}) []*infrapb.Agent {

	if len(in) == 0 || in[0] == nil {
		return nil
	}

	var outAgents []*infrapb.Agent

	for _, inAgent := range in {

		outAgent := &infrapb.Agent{}
		inA := inAgent.(map[string]interface{})
		if val, ok := inA["name"]; ok {
			outAgent.Name = val.(string)
		}
		outAgents = append(outAgents, outAgent)

	}

	return outAgents
}

func convertValStrToVal(valSt string) *structpb.Value {

	var val interface{}
	err := json.Unmarshal([]byte(valSt), &val)
	if err != nil {
		log.Printf("convertValStrToVal err %s\n", err)
	}
	sVal, err := structpb.NewValue(val)
	if err != nil {
		log.Printf("convertValStrToVal err %s\n", err)
	}

	return sVal

}
