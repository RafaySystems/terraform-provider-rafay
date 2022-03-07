package rafay

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/RafaySystems/rafay-common/pkg/hub/client/options"
	typed "github.com/RafaySystems/rafay-common/pkg/hub/client/typed"
	"github.com/RafaySystems/rafay-common/pkg/hub/terraform/resource"
	"github.com/RafaySystems/rafay-common/proto/types/hub/commonpb"
	"github.com/RafaySystems/rafay-common/proto/types/hub/infrapb"
	"github.com/RafaySystems/rctl/pkg/blueprint"
	"github.com/RafaySystems/rctl/pkg/config"
	"github.com/RafaySystems/rctl/pkg/versioninfo"
	"github.com/davecgh/go-spew/spew"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	ClusterScoped   = "cluster-scoped"
	NamespaceScoped = "namespace-scoped"
)

func resourceBluePrint() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceBluePrintCreate,
		ReadContext:   resourceBluePrintRead,
		UpdateContext: resourceBluePrintUpdate,
		DeleteContext: resourceBluePrintDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema:        resource.BlueprintSchema.Schema,
	}
}

func resourceBluePrintCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("blueprint create starts")
	diags := resourceBluePrintUpsert(ctx, d, m)
	if diags.HasError() {
		tflog := os.Getenv("TF_LOG")
		if tflog == "TRACE" || tflog == "DEBUG" {
			ctx = context.WithValue(ctx, "debug", "true")
		}
		log.Printf("blueprint create got error, perform cleanup")
		bp, err := expandBluePrint(d)
		if err != nil {
			log.Printf("blueprint expandBluePrint error")
			return diag.FromErr(err)
		}
		auth := config.GetConfig().GetAppAuthProfile()
		client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
		if err != nil {
			return diag.FromErr(err)
		}

		err = client.InfraV3().Blueprint().Delete(ctx, options.DeleteOptions{
			Name:    bp.Metadata.Name,
			Project: bp.Metadata.Project,
		})
		if err != nil {
			return diag.FromErr(err)
		}
	}
	return diags
}

func resourceBluePrintUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("blueprint update starts")
	return resourceBluePrintUpsert(ctx, d, m)
}

func resourceBluePrintUpsert(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("blueprint upsert starts")
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	blueprint, err := expandBluePrint(d)
	if err != nil {
		log.Printf("blueprint expandBluePrint error")
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.InfraV3().Blueprint().Apply(ctx, blueprint, options.ApplyOptions{})
	if err != nil {
		// XXX Debug
		n1 := spew.Sprintf("%+v", blueprint)
		log.Println("blueprint apply blueprint:", n1)
		log.Printf("blueprint apply error")
		return diag.FromErr(err)
	}

	d.SetId(blueprint.Metadata.Name)
	return diags

}

func resourceBluePrintRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	log.Println("resourceBlueprintRead ")
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}
	tfBlueprintState, err := expandBluePrint(d)
	if err != nil {
		return diag.FromErr(err)
	}

	// XXX Debug
	// w1 := spew.Sprintf("%+v", tfBlueprintState)
	// log.Println("resourceBluePrintRead tfBlueprintState", w1)

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}

	bp, err := client.InfraV3().Blueprint().Get(ctx, options.GetOptions{
		Name:    tfBlueprintState.Metadata.Name,
		Project: tfBlueprintState.Metadata.Project,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	// XXX Debug
	// w1 = spew.Sprintf("%+v", wl)
	// log.Println("resourceBluePrintRead wl", w1)

	err = flattenBlueprint(d, bp)
	if err != nil {
		return diag.FromErr(err)
	}
	return diags

}

func resourceBluePrintDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	bp, err := expandBluePrint(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.InfraV3().Blueprint().Delete(ctx, options.DeleteOptions{
		Name:    bp.Metadata.Name,
		Project: bp.Metadata.Project,
	})

	if err != nil {
		//v3 spec gave error try v2
		return resourceBlueprintV2Delete(ctx, bp)
	}

	return diags
}

func resourceBlueprintV2Delete(ctx context.Context, bp *infrapb.Blueprint) diag.Diagnostics {
	var diags diag.Diagnostics

	projectId, err := config.GetProjectIdByName(bp.Metadata.Project)
	if err != nil {
		return diag.FromErr(err)
	}

	errDel := blueprint.DeleteBlueprint(bp.Metadata.Name, projectId)
	if errDel != nil {
		fmt.Printf("error while deleting blueprint %s", errDel.Error())
		return diag.FromErr(errDel)
	}

	return diags
}

func expandBluePrint(in *schema.ResourceData) (*infrapb.Blueprint, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "expand blueprint empty input")
	}
	obj := &infrapb.Blueprint{}

	if v, ok := in.Get("metadata").([]interface{}); ok {
		obj.Metadata = expandMetaData(v)
	}

	if v, ok := in.Get("spec").([]interface{}); ok {
		objSpec, err := expandBluePrintSpec(v)
		if err != nil {
			return nil, err
		}
		log.Println("expandBluePrintSpec got spec")
		obj.Spec = objSpec
	}

	obj.ApiVersion = "infra.k8smgmt.io/v3"
	obj.Kind = "Blueprint"
	return obj, nil
}

func expandBluePrintSpec(p []interface{}) (*infrapb.BlueprintSpec, error) {
	obj := &infrapb.BlueprintSpec{}
	if len(p) == 0 || p[0] == nil {
		return obj, fmt.Errorf("%s", "expandAddonSpec empty input")
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["version"].(string); ok && len(v) > 0 {
		obj.Version = v
	}

	if v, ok := in["default_addons"].([]interface{}); ok && len(v) > 0 {
		obj.DefaultAddons, _ = expandDefaultAddons(v)
	}

	if v, ok := in["custom_addons"].([]interface{}); ok && len(v) > 0 {
		obj.CustomAddons = expandCustomAddons(v)
	}

	if v, ok := in["psp"].([]interface{}); ok && len(v) > 0 {
		obj.Psp = expandBlueprintPSP(v)
	}

	if v, ok := in["sharing"].([]interface{}); ok {
		obj.Sharing = expandSharingSpec(v)
	}

	if v, ok := in["drift"].([]interface{}); ok {
		obj.Drift = expandDrift(v)
	}

	if v, ok := in["base"].([]interface{}); ok {
		obj.Base = expandBlueprintBase(v)
	}

	if v, ok := in["private_kube_api_proxies"].([]interface{}); ok && len(v) > 0 {
		obj.PrivateKubeAPIProxies = expandPrivateKubeAPIProxies(v)
	}

	return obj, nil
}

func expandDefaultAddons(p []interface{}) (*infrapb.DefaultAddons, error) {
	obj := &infrapb.DefaultAddons{}
	if len(p) == 0 || p[0] == nil {
		return obj, fmt.Errorf("%s", "expandDefaultAddons empty input")
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["enable_ingress"].(bool); ok {
		obj.EnableIngress = v
	}

	if v, ok := in["enable_logging"].(bool); ok {
		obj.EnableLogging = v
	}

	if v, ok := in["enable_monitoring"].(bool); ok {
		obj.EnableMonitoring = v
	}

	if v, ok := in["enable_vm"].(bool); ok {
		obj.EnableVM = v
	}

	if v, ok := in["monitoring"].([]interface{}); ok {
		obj.Monitoring = expandMonitoring(v)
	}

	return obj, nil
}

func expandMonitoring(p []interface{}) *infrapb.MonitoringConfig {
	obj := &infrapb.MonitoringConfig{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["prometheus_adapter"].([]interface{}); ok {
		obj.PrometheusAdapter = expandMonitoringComponent(v)
	}

	if v, ok := in["metrics_server"].([]interface{}); ok {
		obj.MetricsServer = expandMonitoringComponent(v)
	}

	if v, ok := in["kube_state_metrics"].([]interface{}); ok {
		obj.KubeStateMetrics = expandMonitoringComponent(v)
	}

	if v, ok := in["node_exporter"].([]interface{}); ok {
		obj.NodeExporter = expandMonitoringComponent(v)
	}

	if v, ok := in["helm_exporter"].([]interface{}); ok {
		obj.HelmExporter = expandMonitoringComponent(v)
	}

	if v, ok := in["resources"].([]interface{}); ok {
		obj.Resources = expandResources(v)
	}

	return obj
}

func expandMonitoringComponent(p []interface{}) *infrapb.MonitoringComponent {
	obj := &infrapb.MonitoringComponent{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["enabled"].(bool); ok {
		obj.Enabled = v
	}

	if v, ok := in["discovery"].([]interface{}); ok {
		obj.Discovery = expandDiscovery(v)
	}

	return obj
}

func expandDiscovery(p []interface{}) *infrapb.MonitoringDiscoveryConfig {
	obj := &infrapb.MonitoringDiscoveryConfig{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["namespace"].(string); ok {
		obj.Namespace = v
	}

	if v, ok := in["resource"].(string); ok {
		obj.Resource = v
	}

	if v, ok := in["labels"].(map[string]interface{}); ok && len(v) > 0 {
		obj.Labels = toMapString(v)
	}

	return obj
}

func expandResources(p []interface{}) *commonpb.ResourceRequirements {
	obj := &commonpb.ResourceRequirements{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}

	in := p[0].(map[string]interface{})
	if v, ok := in["resources"].([]interface{}); ok {
		obj.Limits = expandResourceQuantity(v)
	}

	return obj
}

func expandCustomAddons(p []interface{}) []*infrapb.BlueprintAddon {
	if len(p) == 0 || p[0] == nil {
		return []*infrapb.BlueprintAddon{}
	}

	out := make([]*infrapb.BlueprintAddon, len(p))

	for i := range p {
		obj := infrapb.BlueprintAddon{}
		in := p[i].(map[string]interface{})

		if v, ok := in["name"].(string); ok {
			obj.Name = v
		}

		if v, ok := in["version"].(string); ok {
			obj.Version = v
		}

		if v, ok := in["depends_on"].([]interface{}); ok {
			obj.DependsOn = toArrayString(v)
		}

		out[i] = &obj

	}

	return out
}

func expandPrivateKubeAPIProxies(p []interface{}) []*infrapb.KubeAPIProxyNetwork {
	if len(p) == 0 || p[0] == nil {
		return []*infrapb.KubeAPIProxyNetwork{}
	}

	out := make([]*infrapb.KubeAPIProxyNetwork, len(p))

	for i := range p {
		obj := infrapb.KubeAPIProxyNetwork{}
		in := p[i].(map[string]interface{})

		if v, ok := in["name"].(string); ok {
			obj.Name = v
		}

		if v, ok := in["id"].(string); ok {
			obj.Id = v
		}

		out[i] = &obj

	}

	return out
}

func expandBlueprintPSP(p []interface{}) *infrapb.BlueprintPSP {
	obj := &infrapb.BlueprintPSP{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["enabled"].(bool); ok {
		obj.Enabled = v
	}

	if v, ok := in["scope"].(string); ok && len(v) > 0 {
		obj.Scope = v
	}

	if v, ok := in["names"].([]interface{}); ok {
		obj.Names = toArrayString(v)
	}

	return obj
}

func expandBlueprintBase(p []interface{}) *infrapb.BlueprintBase {
	obj := &infrapb.BlueprintBase{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["name"].(string); ok {
		obj.Name = v
	}

	if v, ok := in["version"].(string); ok {
		obj.Version = v
	}

	return obj
}

// Flatteners

func flattenBlueprint(d *schema.ResourceData, in *infrapb.Blueprint) error {
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

	// XXX Debug
	// w1 := spew.Sprintf("%+v", v)
	// log.Println("flattenBlueprint before ", w1)
	var ret []interface{}
	ret, err = flattenBlueprintSpec(in.Spec, v)
	if err != nil {
		return err
	}
	// XXX Debug
	// w1 = spew.Sprintf("%+v", ret)
	// log.Println("flattenBlueprint after ", w1)

	err = d.Set("spec", ret)
	if err != nil {
		return err
	}
	return nil
}

func flattenBlueprintSpec(in *infrapb.BlueprintSpec, p []interface{}) ([]interface{}, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "flattenBlueprintSpec empty input")
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.Version) > 0 {
		obj["version"] = in.Version
	}

	if in.DefaultAddons != nil {
		v, ok := obj["default_addons"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["default_addons"] = flattenDefaultAddons(in.DefaultAddons, v)
	}

	if in.CustomAddons != nil && len(in.CustomAddons) > 0 {
		v, ok := obj["custom_addons"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["custom_addons"] = flattenCustomAddons(in.CustomAddons, v)
	}

	if in.Base != nil {
		v, ok := obj["base"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["base"] = flattenBlueprintBase(in.Base, v)
	}

	if in.PrivateKubeAPIProxies != nil {
		v, ok := obj["private_kube_api_proxies"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["private_kube_api_proxies"] = flattenKubeAPIProxyNetwork(in.PrivateKubeAPIProxies, v)
	}

	if in.Sharing != nil {
		obj["sharing"] = flattenSharingSpec(in.Sharing)
	}

	if in.Drift != nil {
		obj["drift"] = flattenDrift(in.Drift)
	}

	return []interface{}{obj}, nil
}

func flattenDefaultAddons(in *infrapb.DefaultAddons, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	obj["enable_ingress"] = in.EnableIngress
	obj["enable_logging"] = in.EnableLogging
	obj["enable_monitoring"] = in.EnableMonitoring
	obj["enable_vm"] = in.EnableVM

	if in.Monitoring != nil {
		v, ok := obj["monitoring"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["monitoring"] = flattenMonitoring(in.Monitoring, v)
	}

	return []interface{}{obj}
}

func flattenMonitoring(in *infrapb.MonitoringConfig, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if in.PrometheusAdapter != nil {
		v, ok := obj["prometheus_adapter"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["prometheus_adapter"] = flattenMonitoringComponent(in.PrometheusAdapter, v)
	}

	if in.MetricsServer != nil {
		v, ok := obj["metrics_server"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["metrics_server"] = flattenMonitoringComponent(in.MetricsServer, v)
	}

	if in.KubeStateMetrics != nil {
		v, ok := obj["kube_state_metrics"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["kube_state_metrics"] = flattenMonitoringComponent(in.KubeStateMetrics, v)
	}

	if in.NodeExporter != nil {
		v, ok := obj["node_exporter"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["node_exporter"] = flattenMonitoringComponent(in.NodeExporter, v)
	}

	if in.HelmExporter != nil {
		v, ok := obj["helm_exporter"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["helm_exporter"] = flattenMonitoringComponent(in.HelmExporter, v)
	}

	if in.Resources != nil {
		v, ok := obj["resources"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["resources"] = flattenResources(in.Resources, v)
	}

	return []interface{}{obj}
}

func flattenMonitoringComponent(in *infrapb.MonitoringComponent, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	obj["enabled"] = in.Enabled
	if in.Discovery != nil {
		v, ok := obj["discovery"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["discovery"] = flattenDiscovery(in.Discovery, v)
	}
	return []interface{}{obj}
}

func flattenDiscovery(in *infrapb.MonitoringDiscoveryConfig, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.Namespace) > 0 {
		obj["namespace"] = in.Namespace
	}
	if len(in.Resource) > 0 {
		obj["resource"] = in.Resource
	}
	if len(in.Labels) > 0 {
		obj["labels"] = toMapInterface(in.Labels)
	}

	return []interface{}{obj}
}

func flattenResources(in *commonpb.ResourceRequirements, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if in.Limits != nil {
		obj["limits"] = flattenResourceQuantity(in.Limits)
	}

	return []interface{}{obj}
}

func flattenCustomAddons(input []*infrapb.BlueprintAddon, p []interface{}) []interface{} {
	log.Println("flattenCustomAddons")
	if input == nil {
		return nil
	}
	out := make([]interface{}, len(input))
	for i, in := range input {

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

		if len(in.DependsOn) > 0 {
			obj["depends_on"] = toArrayInterface(in.DependsOn)
		}

		out[i] = &obj
	}

	return out
}

func flattenBlueprintPSP(in *infrapb.BlueprintPSP, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	obj["enabled"] = in.Enabled
	if len(in.Scope) > 0 {
		obj["scope"] = in.Scope
	}
	if len(in.Names) > 0 {
		obj["names"] = toArrayInterface(in.Names)
	}

	return []interface{}{obj}
}

func flattenBlueprintBase(in *infrapb.BlueprintBase, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.Version) > 0 {
		obj["version"] = in.Version
	}
	if len(in.Name) > 0 {
		obj["name"] = in.Name
	}

	return []interface{}{obj}
}

func flattenKubeAPIProxyNetwork(input []*infrapb.KubeAPIProxyNetwork, p []interface{}) []interface{} {
	log.Println("flattenKubeAPIProxyNetwork")
	if input == nil {
		return nil
	}
	out := make([]interface{}, len(input))
	for i, in := range input {

		obj := map[string]interface{}{}
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}

		if len(in.Id) > 0 {
			obj["version"] = in.Id
		}
		if len(in.Name) > 0 {
			obj["name"] = in.Name
		}

		out[i] = &obj
	}

	return out
}
