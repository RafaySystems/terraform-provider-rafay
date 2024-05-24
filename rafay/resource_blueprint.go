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
	"github.com/RafaySystems/rafay-common/proto/types/hub/infrapb"
	"github.com/RafaySystems/rctl/pkg/blueprint"
	bp "github.com/RafaySystems/rctl/pkg/blueprint"
	"github.com/RafaySystems/rctl/pkg/config"
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
		Importer: &schema.ResourceImporter{
			State: resourceBluePrintImport,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema:        resource.BlueprintSchema.Schema,
	}
}

func resourceBluePrintImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	idParts := strings.SplitN(d.Id(), "/", 2)
	log.Println("resourceBluePrintImport idParts:", idParts)
	d_debug := spew.Sprintf("%+v", d)
	log.Println("resourceBluePrintImport d.Id:", d.Id())
	log.Println("resourceBluePrintImport d_debug", d_debug)

	blueprint, err := expandBluePrint(d)
	if err != nil {
		log.Printf("blueprint expandBluePrint error")
		return nil, err
	}

	var metaD commonpb.Metadata
	metaD.Name = idParts[0]
	metaD.Project = idParts[1]
	blueprint.Metadata = &metaD

	err = d.Set("metadata", flattenMetaData(blueprint.Metadata))
	if err != nil {
		return nil, err
	}

	d.SetId(blueprint.Metadata.Name)

	return []*schema.ResourceData{d}, nil
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
			return diags
		}
		auth := config.GetConfig().GetAppAuthProfile()
		client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, TF_USER_AGENT, options.WithInsecureSkipVerify(auth.SkipServerCertValid))
		if err != nil {
			return diags
		}

		err = client.InfraV3().Blueprint().Delete(ctx, options.DeleteOptions{
			Name:    bp.Metadata.Name,
			Project: bp.Metadata.Project,
		})
		if err != nil {
			return diags
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

	if d.State() != nil && d.State().ID != "" {
		n := GetMetaName(d)
		if n != "" && n != d.State().ID {
			log.Printf("metadata name change not supported")
			d.State().Tainted = true
			return diag.FromErr(fmt.Errorf("%s", "metadata name change not supported"))
		}
	}

	blueprint, err := expandBluePrint(d)
	if err != nil {
		log.Printf("blueprint expandBluePrint error")
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, TF_USER_AGENT, options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.InfraV3().Blueprint().Apply(ctx, blueprint, options.ApplyOptions{})
	if err != nil {
		// XXX Debug
		n1 := spew.Sprintf("%+v", blueprint)
		log.Println("blueprint apply blueprint:", n1)
		log.Printf("blueprint apply error: %v", err)
		return diag.FromErr(err)
	}

	d.SetId(blueprint.Metadata.Name)

	//blueprint publish
	projectId, err := config.GetProjectIdByName(blueprint.Metadata.Project)
	if err != nil {
		return diag.FromErr(err)
	}
	bp.PublishBlueprint(blueprint.Metadata.Name, blueprint.Spec.Version, blueprint.Metadata.Description, projectId)

	return diags
}

func resourceBluePrintRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	log.Println("resourceBlueprintRead ")
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	meta := GetMetaData(d)
	if meta == nil {
		return diag.FromErr(fmt.Errorf("%s", "failed to read resource "))
	}
	if d.State() != nil && d.State().ID != "" {
		meta.Name = d.State().ID
	}

	tfBlueprintState, err := expandBluePrint(d)
	if err != nil {
		return diag.FromErr(err)
	}

	// XXX Debug
	// w1 := spew.Sprintf("%+v", tfBlueprintState)
	// log.Println("resourceBluePrintRead tfBlueprintState", w1)

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, TF_USER_AGENT, options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return diag.FromErr(err)
	}

	bp, err := client.InfraV3().Blueprint().Get(ctx, options.GetOptions{
		//Name:    tfBlueprintState.Metadata.Name,
		Name:    meta.Name,
		Project: tfBlueprintState.Metadata.Project,
	})
	if err != nil {
		if strings.Contains(err.Error(), "code 404") {
			log.Println("Resource Read ", "error", err)
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	if tfBlueprintState.Spec != nil && tfBlueprintState.Spec.Sharing != nil && !tfBlueprintState.Spec.Sharing.Enabled && bp.Spec.Sharing == nil {
		bp.Spec.Sharing = &commonpb.SharingSpec{}
		bp.Spec.Sharing.Enabled = false
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
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, TF_USER_AGENT, options.WithInsecureSkipVerify(auth.SkipServerCertValid))
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

	if v, ok := in.Get("metadata").([]interface{}); ok && len(v) > 0 {
		obj.Metadata = expandMetaData(v)
	}

	if v, ok := in.Get("spec").([]interface{}); ok && len(v) > 0 {
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
	} else {
		log.Println("expandBluePrintSpec empty default_addons")
		empt := make([]interface{}, 0)
		obj.DefaultAddons, _ = expandDefaultAddons(empt)
		log.Println("expandBluePrintSpec Ingress ", obj.DefaultAddons.EnableIngress)
	}
	da := spew.Sprintf("%+v", obj.DefaultAddons)
	log.Println("expandBluePrintSpec DefaultAddons ", da)

	if v, ok := in["custom_addons"].([]interface{}); ok && len(v) > 0 {
		obj.CustomAddons = expandCustomAddons(v)
	}
	if v, ok := in["components_criticality"].([]interface{}); ok && len(v) > 0 {
		obj.ComponentsCriticality = expandComponentCriticality(v)
	}
	ca := spew.Sprintf("%+v", obj.CustomAddons)
	log.Println("expandBluePrintSpec CustomAddons ", ca)

	if v, ok := in["psp"].([]interface{}); ok && len(v) > 0 {
		obj.Psp = expandBlueprintPSP(v)
	}

	if v, ok := in["sharing"].([]interface{}); ok && len(v) > 0 {
		obj.Sharing = expandSharingSpec(v)
	}

	if v, ok := in["type"].(string); ok && len(v) > 0 {
		obj.Type = v
	}

	if v, ok := in["drift"].([]interface{}); ok && len(v) > 0 {
		obj.Drift = expandDrift(v)
	}

	if v, ok := in["drift_webhook"].([]interface{}); ok && len(v) > 0 {
		obj.DriftWebhook = expandDriftWebhook(v)
	}

	if v, ok := in["namespace_config"].([]interface{}); ok && len(v) > 0 {
		obj.NamespaceConfig = expandBlueprintNamespaceConfig(v)
	}

	if v, ok := in["network_policy"].([]interface{}); ok && len(v) > 0 {
		obj.NetworkPolicy = expandBlueprintNetworkPolicy(v)
	}

	if v, ok := in["service_mesh"].([]interface{}); ok && len(v) > 0 {
		obj.ServiceMesh = expandBlueprintServiceMesh(v)
	}

	if v, ok := in["cost_profile"].([]interface{}); ok && len(v) > 0 {
		obj.CostProfile = expandBlueprintCostProfile(v)
	}

	if v, ok := in["opa_policy"].([]interface{}); ok && len(v) > 0 {
		obj.OpaPolicy = expandBlueprintOPAPolicy(v)
	}

	if v, ok := in["base"].([]interface{}); ok && len(v) > 0 {
		obj.Base = expandBlueprintBase(v)
	}

	if v, ok := in["private_kube_api_proxies"].([]interface{}); ok && len(v) > 0 {
		obj.PrivateKubeAPIProxies = expandPrivateKubeAPIProxies(v)
	}

	if v, ok := in["placement"].([]interface{}); ok && len(v) > 0 {
		obj.Placement = expandBlueprintPlacement(v)
	}
	pa := spew.Sprintf("%+v", obj.Placement)
	log.Println("expandBluePrintSpec Placement:", pa)

	return obj, nil
}

func expandBlueprintNamespaceConfig(p []interface{}) *infrapb.NsConfig {
	obj := &infrapb.NsConfig{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}

	in := p[0].(map[string]interface{})

	// if v, ok := in["deny_out_of_band_creation"].(bool); ok {
	// 	obj.DenyOutOfBandCreation = v
	// }

	if v, ok := in["enable_sync"].(bool); ok {
		obj.EnableSync = v
	}

	if v, ok := in["sync_type"].(string); ok {
		obj.SyncType = v
	}

	return obj
}

func expandBlueprintPlacement(p []interface{}) *infrapb.BlueprintPlacement {
	obj := &infrapb.BlueprintPlacement{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["auto_publish"].(bool); ok {
		obj.AutoPublish = v
	}

	if v, ok := in["fleet_values"].([]interface{}); ok && len(v) > 0 {
		obj.FleetValues = toArrayStringSorted(v)
	}

	return obj
}

func expandDefaultAddons(p []interface{}) (*infrapb.DefaultAddons, error) {
	obj := &infrapb.DefaultAddons{}
	if len(p) == 0 || p[0] == nil {
		obj.EnableIngress = false
		log.Println("expandDefaultAddons empty")
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

	if v, ok := in["enable_rook_ceph"].(bool); ok {
		obj.EnableRookCeph = v
	}

	if v, ok := in["enable_vm"].(bool); ok {
		obj.EnableVM = v
	}

	if v, ok := in["enable_csi_secret_store"].(bool); ok {
		obj.EnableCsiSecretStore = v
	}

	if v, ok := in["csi_secret_store_config"].([]interface{}); ok && len(v) > 0 {
		obj.CsiSecretStoreConfig = expandCsiSecretStoreConfig(v)
	}

	if v, ok := in["monitoring"].([]interface{}); ok && len(v) > 0 {
		obj.Monitoring = expandMonitoring(v)
		rs := spew.Sprintf("%+v", obj.Monitoring)
		log.Println("expandDefaultAddons Monitoring", rs)
	}

	return obj, nil
}

func expandCsiSecretStoreConfig(p []interface{}) *infrapb.CsiSecretStoreConfig {
	obj := &infrapb.CsiSecretStoreConfig{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["sync_secrets"].(bool); ok {
		obj.SyncSecrets = v
	}

	if v, ok := in["enable_secret_rotation"].(bool); ok {
		obj.EnableSecretRotation = v
	}

	if v, ok := in["rotation_poll_interval"].(string); ok {
		obj.RotationPollInterval = v
	}

	if v, ok := in["providers"].([]interface{}); ok && len(v) > 0 {
		obj.Providers = expandCsiSecretStoreProviders(v)
	}

	return obj
}

func expandCsiSecretStoreProviders(p []interface{}) *infrapb.SecretStoreProvider {
	obj := &infrapb.SecretStoreProvider{}

	if len(p) == 0 || p[0] == nil {
		return obj
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["aws"].(bool); ok {
		obj.Aws = v
	}

	if v, ok := in["vault"].(bool); ok {
		obj.Vault = v
	}

	return obj
}

func expandMonitoring(p []interface{}) *infrapb.MonitoringConfig {
	obj := &infrapb.MonitoringConfig{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["prometheus_adapter"].([]interface{}); ok && len(v) > 0 {
		obj.PrometheusAdapter = expandMonitoringComponent(v)
	}

	if v, ok := in["metrics_server"].([]interface{}); ok && len(v) > 0 {
		obj.MetricsServer = expandMonitoringComponent(v)
	}

	if v, ok := in["kube_state_metrics"].([]interface{}); ok && len(v) > 0 {
		obj.KubeStateMetrics = expandMonitoringComponent(v)
	}

	if v, ok := in["node_exporter"].([]interface{}); ok && len(v) > 0 {
		obj.NodeExporter = expandMonitoringComponent(v)
	}

	if v, ok := in["helm_exporter"].([]interface{}); ok && len(v) > 0 {
		obj.HelmExporter = expandMonitoringComponent(v)
	}

	if v, ok := in["resources"].([]interface{}); ok && len(v) > 0 {
		obj.Resources = expandResources(v)
		rs := spew.Sprintf("%+v", obj.Resources.Limits)
		log.Println("expandMonitoring Resources", rs)
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
	if v, ok := in["customization_enabled"].(bool); ok {
		obj.CustomizationEnabled = v
	}
	if v, ok := in["discovery"].([]interface{}); ok && len(v) > 0 {
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

	if v, ok := in["namespace"].(string); ok && len(v) > 0 {
		obj.Namespace = v
	}

	if v, ok := in["resource"].(string); ok && len(v) > 0 {
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
	if v, ok := in["limits"].([]interface{}); ok && len(v) > 0 {
		obj.Limits = expandResourceQuantity1170(v)
		log.Println("expandResources Limits ", obj.Limits)
	}

	return obj
}

func expandComponentCriticality(p []interface{}) []*infrapb.SnapshotRef {
	if len(p) == 0 || p[0] == nil {
		return []*infrapb.SnapshotRef{}
	}

	out := make([]*infrapb.SnapshotRef, len(p))
	for i := range p {
		obj := infrapb.SnapshotRef{}
		in := p[i].(map[string]interface{})

		if v, ok := in["name"].(string); ok && len(v) > 0 {
			obj.Name = v
		}
		if v, ok := in["version"].(string); ok && len(v) > 0 {
			obj.Version = v
		}
		if v, ok := in["componentType"].(string); ok && len(v) > 0 {
			obj.ComponentType = v
		}
		if v, ok := in["dependsOn"].([]interface{}); ok && len(v) > 0 {
			obj.DependsOn = toArrayString(v)
		}
		if v, ok := in["isGlobal"].(bool); ok {
			obj.IsGlobal = v
		}
		if v, ok := in["publishedGeneration"].(string); ok && len(v) > 0 {
			obj.PublishedGeneration = v
		}
		if v, ok := in["selector"].(string); ok && len(v) > 0 {
			obj.Selector = v
		}
		if v, ok := in["workloadID"].(string); ok && len(v) > 0 {
			obj.WorkloadID = v
		}
		if v, ok := in["refType"].(infrapb.SnapshotRefType); ok {
			obj.RefType = v
		}
		out[i] = &obj
	}
	return out
}

func expandCustomAddons(p []interface{}) []*infrapb.BlueprintAddon {
	if len(p) == 0 || p[0] == nil {
		return []*infrapb.BlueprintAddon{}
	}

	out := make([]*infrapb.BlueprintAddon, len(p))

	for i := range p {
		obj := infrapb.BlueprintAddon{}
		in := p[i].(map[string]interface{})

		if v, ok := in["name"].(string); ok && len(v) > 0 {
			obj.Name = v
		}

		if v, ok := in["version"].(string); ok && len(v) > 0 {
			obj.Version = v
		}

		if v, ok := in["depends_on"].([]interface{}); ok && len(v) > 0 {
			obj.DependsOn = toArrayString(v)
			log.Println("expandCustomAddons depends_on ", obj.DependsOn)
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

		if v, ok := in["name"].(string); ok && len(v) > 0 {
			obj.Name = v
		}

		if v, ok := in["id"].(string); ok && len(v) > 0 {
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

	if v, ok := in["names"].([]interface{}); ok && len(v) > 0 {
		obj.Names = toArrayString(v)
	}

	return obj
}

func expandBlueprintServiceMesh(p []interface{}) *infrapb.ServiceMesh {
	obj := &infrapb.ServiceMesh{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["profile"].([]interface{}); ok && len(v) > 0 {
		obj.Profile = expandBlueprintMeshProfile(v)
	}

	if v, ok := in["policies"].([]interface{}); ok && len(v) > 0 {
		obj.Policies = expandBlueprintClusterMeshPolicies(v)
	}

	return obj
}

func expandBlueprintCostProfile(p []interface{}) *infrapb.CostProfile {
	obj := &infrapb.CostProfile{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["name"].(string); ok && len(v) > 0 {
		obj.Name = v
	}

	if v, ok := in["version"].(string); ok && len(v) > 0 {
		obj.Version = v
	}

	return obj
}

func expandBlueprintNetworkPolicy(p []interface{}) *infrapb.NetworkPolicy {
	obj := &infrapb.NetworkPolicy{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["profile"].([]interface{}); ok && len(v) > 0 {
		obj.Profile = expandBlueprintNetworkPolicyProfile(v)
	}

	if v, ok := in["policies"].([]interface{}); ok && len(v) > 0 {
		obj.Policies = expandBlueprintNetworkPolicyPolicies(v)
	}

	return obj
}

func expandBlueprintOPAPolicies(p []interface{}) []*infrapb.Policy {
	if len(p) == 0 || p[0] == nil {
		return []*infrapb.Policy{}
	}

	out := make([]*infrapb.Policy, len(p))

	for i := range p {
		obj := infrapb.Policy{}
		in := p[i].(map[string]interface{})

		if v, ok := in["name"].(string); ok {
			obj.Name = v
		}

		if v, ok := in["version"].(string); ok {
			obj.Version = v
		}

		out[i] = &obj
	}

	return out
}

func expandBlueprintOPAProfile(p []interface{}) *infrapb.OPAProfile {
	obj := &infrapb.OPAProfile{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["name"].(string); ok && len(v) > 0 {
		obj.Name = v
	}

	if v, ok := in["version"].(string); ok && len(v) > 0 {
		obj.Version = v
	}

	return obj
}

func expandBlueprintOPAPolicy(p []interface{}) *infrapb.OPAPolicy {
	obj := &infrapb.OPAPolicy{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["opa_policy"].([]interface{}); ok && len(v) > 0 {
		obj.OpaPolicy = expandBlueprintOPAPolicies(v)
	}

	if v, ok := in["profile"].([]interface{}); ok && len(v) > 0 {
		obj.Profile = expandBlueprintOPAProfile(v)
	}

	return obj
}

func expandBlueprintNetworkPolicyProfile(p []interface{}) *commonpb.ResourceNameAndVersionRef {
	obj := &commonpb.ResourceNameAndVersionRef{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["name"].(string); ok && len(v) > 0 {
		obj.Name = v
	}

	if v, ok := in["version"].(string); ok && len(v) > 0 {
		obj.Version = v
	}

	return obj
}

func expandBlueprintNetworkPolicyPolicies(p []interface{}) []*commonpb.ResourceNameAndVersionRef {
	if len(p) == 0 || p[0] == nil {
		return []*commonpb.ResourceNameAndVersionRef{}
	}

	out := make([]*commonpb.ResourceNameAndVersionRef, len(p))

	for i := range p {
		obj := commonpb.ResourceNameAndVersionRef{}
		in := p[i].(map[string]interface{})

		if v, ok := in["name"].(string); ok {
			obj.Name = v
		}

		if v, ok := in["version"].(string); ok {
			obj.Version = v
		}

		out[i] = &obj
	}

	return out

}

func expandBlueprintMeshProfile(p []interface{}) *commonpb.ResourceNameAndVersionRef {
	obj := &commonpb.ResourceNameAndVersionRef{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["name"].(string); ok && len(v) > 0 {
		obj.Name = v
	}

	if v, ok := in["version"].(string); ok && len(v) > 0 {
		obj.Version = v
	}

	return obj
}

func expandBlueprintClusterMeshPolicies(p []interface{}) []*commonpb.ResourceNameAndVersionRef {
	if len(p) == 0 || p[0] == nil {
		return []*commonpb.ResourceNameAndVersionRef{}
	}

	out := make([]*commonpb.ResourceNameAndVersionRef, len(p))

	for i := range p {
		obj := commonpb.ResourceNameAndVersionRef{}
		in := p[i].(map[string]interface{})

		if v, ok := in["name"].(string); ok {
			obj.Name = v
		}

		if v, ok := in["version"].(string); ok {
			obj.Version = v
		}

		out[i] = &obj
	}

	return out

}

func expandBlueprintBase(p []interface{}) *infrapb.BlueprintBase {
	obj := &infrapb.BlueprintBase{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["name"].(string); ok && len(v) > 0 {
		obj.Name = v
	}

	if v, ok := in["version"].(string); ok && len(v) > 0 {
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
	w1 := spew.Sprintf("%+v", v)
	log.Println("flattenBlueprint before ", w1)
	var ret []interface{}
	ret, err = flattenBlueprintSpec(in.Spec, v)
	if err != nil {
		return err
	}
	// XXX Debug
	w1 = spew.Sprintf("%+v", ret)
	log.Println("flattenBlueprint after ", w1)

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
		obj["custom_addons"] = flatteCustomAddons(in.CustomAddons, v)
	}

	if in.ComponentsCriticality != nil && len(in.ComponentsCriticality) > 0 {
		v, ok := obj["components_criticality"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["components_criticality"] = flattenComponenetCriticality(in.ComponentsCriticality, v)
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

	obj["sharing"] = flattenSharingSpec(in.Sharing)
	obj["type"] = in.Type

	if in.Drift != nil {
		obj["drift"] = flattenDrift(in.Drift)
	}

	if in.NamespaceConfig != nil {
		v, ok := obj["namespace_config"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["namespace_config"] = flattenBlueprintNamespaceConfig(in.NamespaceConfig, v)
	}

	if in.OpaPolicy != nil {
		v, ok := obj["opa_policy"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["opa_policy"] = flattenBlueprintOpaPolicy(in.OpaPolicy, v)
	} else {
		obj["opa_policy"] = nil
	}

	if in.DriftWebhook != nil {
		v, ok := obj["drift_webhook"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["drift_webhook"] = flattenDriftWebhook(in.DriftWebhook, v)
	}

	if in.NetworkPolicy != nil {
		v, ok := obj["network_policy"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["network_policy"] = flattenBlueprintNetworkPolicy(in.NetworkPolicy, v)
	}

	if in.ServiceMesh != nil {
		v, ok := obj["service_mesh"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["service_mesh"] = flattenBlueprintServiceMesh(in.ServiceMesh, v)
	}

	if in.CostProfile != nil {
		v, ok := obj["cost_profile"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["cost_profile"] = flattenBlueprintCostProfile(in.CostProfile, v)
	}

	if in.Placement != nil {
		v, ok := obj["placement"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["placement"] = flattenBlueprintPlacement(in.Placement, v)
	}

	return []interface{}{obj}, nil
}

func flattenBlueprintOpaPolicy(in *infrapb.OPAPolicy, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	//obj["enabled"] = true

	if in.Profile != nil {
		v, ok := obj["profile"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["profile"] = flattenBlueprintOpaPolicyProfile(in.Profile, v)
	}

	if in.Profile != nil {
		v, ok := obj["opa_policy"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["opa_policy"] = flattenBlueprintOpaPolicies(in.OpaPolicy, v)
	}

	return []interface{}{obj}
}

func flattenDriftWebhook(in *infrapb.DriftWebhook, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	obj["enabled"] = in.Enabled

	return []interface{}{obj}
}

func flattenBlueprintNetworkPolicy(in *infrapb.NetworkPolicy, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if in.Profile != nil {
		v, ok := obj["profile"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["profile"] = flattenBlueprintNetworkPolicyProfile(in.Profile, v)
	}

	if in.Profile != nil {
		v, ok := obj["policies"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["policies"] = flattenBlueprintNetworkPolicies(in.Policies, v)
	}

	return []interface{}{obj}
}

func flattenBlueprintOpaPolicyProfile(in *infrapb.OPAProfile, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.Name) > 0 {
		obj["name"] = in.Name
	}

	if len(in.Version) > 0 {
		obj["version"] = in.Version
	}

	return []interface{}{obj}
}

func flattenBlueprintNetworkPolicyProfile(in *commonpb.ResourceNameAndVersionRef, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.Name) > 0 {
		obj["name"] = in.Name
	}

	if len(in.Version) > 0 {
		obj["version"] = in.Version
	}

	return []interface{}{obj}
}

func flattenBlueprintOpaPolicies(in []*infrapb.Policy, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}

	out := make([]interface{}, len(in))

	for i, in := range in {
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

		out[i] = obj
	}
	return out
}

func flattenBlueprintNetworkPolicies(in []*commonpb.ResourceNameAndVersionRef, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}

	out := make([]interface{}, len(in))

	for i, in := range in {
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
		out[i] = obj
	}
	return out
}

func flattenBlueprintServiceMesh(in *infrapb.ServiceMesh, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if in.Profile != nil {
		v, ok := obj["profile"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["profile"] = flattenBlueprintServiceMeshProfile(in.Profile, v)
	}

	if in.Profile != nil {
		v, ok := obj["policies"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["policies"] = flattenBlueprintServiceMeshPolicies(in.Policies, v)
	}

	return []interface{}{obj}
}

func flattenBlueprintServiceMeshProfile(in *commonpb.ResourceNameAndVersionRef, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.Name) > 0 {
		obj["name"] = in.Name
	}

	if len(in.Version) > 0 {
		obj["version"] = in.Version
	}

	return []interface{}{obj}
}

func flattenBlueprintServiceMeshPolicies(in []*commonpb.ResourceNameAndVersionRef, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}

	out := make([]interface{}, len(in))

	for i, in := range in {
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
		out[i] = obj
	}
	return out
}

func flattenBlueprintCostProfile(in *infrapb.CostProfile, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.Name) > 0 {
		obj["name"] = in.Name
	}

	if len(in.Version) > 0 {
		obj["version"] = in.Version
	}

	return []interface{}{obj}
}

func flattenBlueprintNamespaceConfig(in *infrapb.NsConfig, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}

	obj := make(map[string]interface{})

	// obj["deny_out_of_band_creation"] = in.DenyOutOfBandCreation

	obj["enable_sync"] = in.EnableSync
	obj["sync_type"] = in.SyncType

	return []interface{}{obj}
}

func flattenBlueprintPlacement(in *infrapb.BlueprintPlacement, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}
	if in.AutoPublish {
		obj["auto_publish"] = in.AutoPublish
	}

	if in.FleetValues != nil && len(in.FleetValues) > 0 {
		obj["fleet_values"] = toArrayInterfaceSorted(in.FleetValues)
	}
	return []interface{}{obj}
}

func flattenDefaultAddons(in *infrapb.DefaultAddons, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}

	retNil := false

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	} else {
		retNil = true
	}

	if in.EnableIngress {
		obj["enable_ingress"] = in.EnableIngress
		retNil = false
	}

	if in.EnableLogging {
		obj["enable_logging"] = in.EnableLogging
		retNil = false
	}

	if in.EnableMonitoring {
		obj["enable_monitoring"] = in.EnableMonitoring
		retNil = false

		if in.Monitoring != nil {
			v, ok := obj["monitoring"].([]interface{})
			if !ok {
				v = []interface{}{}
			}
			obj["monitoring"] = flattenMonitoring(in.Monitoring, v)
		}
	}

	if in.EnableCsiSecretStore {
		obj["enable_csi_secret_store"] = in.EnableCsiSecretStore
		retNil = false

		if in.CsiSecretStoreConfig != nil {
			v, ok := obj["csi_secret_store_config"].([]interface{})
			if !ok {
				v = []interface{}{}
			}
			obj["csi_secret_store_config"] = flattenCsiSecretStoreConfig(in.CsiSecretStoreConfig, v)
		}
	}

	if in.EnableRookCeph {
		obj["enable_rook_ceph"] = in.EnableRookCeph
		retNil = false
	}

	if in.EnableVM {
		obj["enable_vm"] = in.EnableVM
		retNil = false
	}

	if retNil {
		return nil
	}

	return []interface{}{obj}
}

func flattenCsiSecretStoreConfig(in *infrapb.CsiSecretStoreConfig, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if in.EnableSecretRotation {
		obj["enable_secret_rotation"] = in.EnableSecretRotation
	}

	if in.SyncSecrets {
		obj["sync_secrets"] = in.SyncSecrets
	}

	if in.RotationPollInterval != "" {
		obj["rotation_poll_interval"] = in.RotationPollInterval
	}

	if in.Providers != nil {
		v, ok := obj["providers"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["providers"] = flattenCsiSecretStoreProviders(in.Providers, v)
		log.Println("flattenCsiSecretStoreConfig in.Providers ", in.Providers)
	}

	return []interface{}{obj}
}

func flattenCsiSecretStoreProviders(in *infrapb.SecretStoreProvider, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}

	obj := make(map[string]interface{})

	if in.Aws {
		obj["aws"] = in.Aws
	}

	if in.Vault {
		obj["vault"] = in.Vault
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
		log.Println("flattenMonitoring in.PrometheusAdapter ", in.PrometheusAdapter)
	}

	if in.MetricsServer != nil {
		v, ok := obj["metrics_server"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["metrics_server"] = flattenMonitoringComponent(in.MetricsServer, v)
		log.Println("flattenMonitoring in.MetricsServer ", in.MetricsServer)
	}

	if in.KubeStateMetrics != nil {
		v, ok := obj["kube_state_metrics"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["kube_state_metrics"] = flattenMonitoringComponent(in.KubeStateMetrics, v)
		log.Println("flattenMonitoring in.KubeStateMetrics ", in.KubeStateMetrics)
	}

	if in.NodeExporter != nil {
		v, ok := obj["node_exporter"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["node_exporter"] = flattenMonitoringComponent(in.NodeExporter, v)
		log.Println("flattenMonitoring in.NodeExporter ", in.NodeExporter)
	}

	if in.HelmExporter != nil {
		v, ok := obj["helm_exporter"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["helm_exporter"] = flattenMonitoringComponent(in.HelmExporter, v)
		log.Println("flattenMonitoring in.HelmExporter ", in.HelmExporter)
	}

	if in.Resources != nil {
		v, ok := obj["resources"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["resources"] = flattenResources(in.Resources, v)
		log.Println("flattenMonitoring in.Resources ", in.Resources)
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
	if in.Enabled {
		obj["enabled"] = in.Enabled
	}
	if in.CustomizationEnabled {
		obj["customization_enabled"] = in.CustomizationEnabled
	}
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
	retNil := true

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.Namespace) > 0 {
		obj["namespace"] = in.Namespace
		retNil = false
	}
	if len(in.Resource) > 0 {
		obj["resource"] = in.Resource
		retNil = false
	}
	if len(in.Labels) > 0 {
		obj["labels"] = toMapInterface(in.Labels)
		retNil = false
	}

	if retNil {
		return nil
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
		obj["limits"] = flattenResourceQuantity1170(in.Limits)
	}

	return []interface{}{obj}
}

func flattenComponenetCriticality(input []*infrapb.SnapshotRef, p []interface{}) []interface{} {
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
		if len(in.ComponentType) > 0 {
			obj["componentType"] = in.ComponentType
		}
		if len(in.DependsOn) > 0 {
			obj["dependsOn"] = in.DependsOn
		}
		if in.IsGlobal {
			obj["isGlobal"] = in.IsGlobal
		}
		if len(in.PublishedGeneration) > 0 {
			obj["publishedGeneration"] = in.PublishedGeneration
		}

		if len(in.Selector) > 0 {
			obj["selector"] = in.Selector
		}
		if len(in.WorkloadID) > 0 {
			obj["workloadID"] = in.WorkloadID
		}
		out[i] = &obj
	}
	return out
}

func flatteCustomAddons(input []*infrapb.BlueprintAddon, p []interface{}) []interface{} {
	log.Println("flatteCustomAddons")
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
