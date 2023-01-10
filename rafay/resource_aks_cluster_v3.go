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
	"github.com/RafaySystems/rafay-common/proto/types/hub/infrapb"
	"github.com/RafaySystems/rctl/pkg/config"
	"github.com/RafaySystems/rctl/pkg/versioninfo"
	"github.com/davecgh/go-spew/spew"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	schema "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAKSClusterV3() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAKSClusterV3Create,
		ReadContext:   resourceAKSClusterV3Read,
		UpdateContext: resourceAKSClusterV3Update,
		DeleteContext: resourceAKSClusterV3Delete,
		Importer: &schema.ResourceImporter{
			State: resourceAKSClusterV3Import,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema:        resource.ClusterSchema.Schema,
	}
}

func resourceAKSClusterV3Create(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// calls upsert
	log.Printf(">>>>>>>>>>>>>> Cluster create starts")
	cluster, err := expandClusterV3(d)
	if err != nil {
		log.Printf(">>>>>>>>>>>>>> ERROR")
	}
	log.Println(">>>>>>>>>>>> CLUSTER", cluster)

	return nil
	return resourceAKSClusterV3Upsert(ctx, d, m)
}

func resourceAKSClusterV3Read(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	return diags
}

func resourceAKSClusterV3Update(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// calls upsert
	return resourceAKSClusterV3Upsert(ctx, d, m)
}

func resourceAKSClusterV3Delete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	return diags
}

func resourceAKSClusterV3Import(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	return []*schema.ResourceData{d}, nil
}

func resourceAKSClusterV3Upsert(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("Cluster upsert starts")
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

	cluster, err := expandClusterV3(d)
	if err != nil {
		log.Printf("Cluster expandCluster error")
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.InfraV3().Cluster().Apply(ctx, cluster, options.ApplyOptions{})
	if err != nil {
		// XXX Debug
		n1 := spew.Sprintf("%+v", cluster)
		log.Println("Cluster apply cluster:", n1)
		log.Printf("Cluster apply error")
		return diag.FromErr(err)
	}

	d.SetId(cluster.Metadata.Name)
	return diags
}

func expandClusterV3(in *schema.ResourceData) (*infrapb.Cluster, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "expand credentials empty input")
	}
	obj := &infrapb.Cluster{}

	if v, ok := in.Get("metadata").([]interface{}); ok && len(v) > 0 {
		obj.Metadata = expandMetaData(v)
	}

	if v, ok := in.Get("spec").([]interface{}); ok && len(v) > 0 {
		objSpec, err := expandClusterV3Spec(v)
		if err != nil {
			return nil, err
		}
		log.Println("expandClusterSpec got spec")
		obj.Spec = objSpec
	}

	obj.ApiVersion = "infra.k8smgmt.io/v3"
	obj.Kind = "Cluster"

	return obj, nil
}

func expandClusterV3Spec(p []interface{}) (*infrapb.ClusterSpec, error) {
	obj := &infrapb.ClusterSpec{}
	if len(p) == 0 || p[0] == nil {
		return obj, fmt.Errorf("%s", "expandClusterSpec empty input")
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["type"].(string); ok && len(v) > 0 {
		obj.Type = v
	}

	if obj.Type != "aks" {
		log.Fatalln("Not Implemented")
	}

	if v, ok := in["sharing"].([]interface{}); ok && len(v) > 0 {
		obj.Sharing = expandSharingSpec(v)
	}

	if v, ok := in["blueprint_config"].([]interface{}); ok && len(v) > 0 {
		obj.BlueprintConfig = expandClusterV3Blueprint(v)
	}

	if v, ok := in["cloud_credentials"].(string); ok && len(v) > 0 {
		obj.CloudCredentials = v
	}

	switch obj.Type {
	case "aks":
		if v, ok := in["config"].([]interface{}); ok && len(v) > 0 {
			obj.Config = expandAKSClusterV3Config(v)
		}
	default:
		log.Fatalln("Not Implemented")
	}

	//if v, ok := in["config"].([]interface{}); ok && len(v) > 0 {
	//	obj.Spec.Config = &infrapb.ClusterSpec_Aks{Aks: &infrapb.AksV3ConfigObject{}}
	//}

	// &infrapb.ClusterSpec_Aks{Aks: &infrapb.AksV3ConfigObject{}}

	// TODO: PROXY CONFIG

	return obj, nil
}

func expandClusterV3Blueprint(p []interface{}) *infrapb.BlueprintConfig {
	obj := infrapb.BlueprintConfig{}
	if len(p) == 0 || p[0] == nil {
		return &obj
	}

	in := p[0].(map[string]interface{})
	if v, ok := in["name"].(string); ok {
		obj.Name = v
	}

	if v, ok := in["version"].(string); ok {
		obj.Version = v
	}

	log.Println("expandClusterV3Blueprint obj", obj)
	return &obj
}

func expandAKSClusterV3Config(p []interface{}) *infrapb.ClusterSpec_Aks {
	obj := &infrapb.ClusterSpec_Aks{Aks: &infrapb.AksV3ConfigObject{}}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["api_version"].(string); ok && len(v) > 0 {
		obj.Aks.ApiVersion = v
	}

	if v, ok := in["kind"].(string); ok && len(v) > 0 {
		obj.Aks.Kind = v
	}

	if v, ok := in["metadata"].([]interface{}); ok && len(v) > 0 {
		obj.Aks.Metadata = expandMetaData(v)
	}

	if v, ok := in["spec"].([]interface{}); ok && len(v) > 0 {
		obj.Aks.Spec = expandAKSClusterV3ConfigSpec(v)
	}

	return obj
}

func expandAKSClusterV3ConfigSpec(p []interface{}) *infrapb.AksV3Spec {
	obj := &infrapb.AksV3Spec{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["subscription_id"].(string); ok && len(v) > 0 {
		obj.SubscriptionID = v
	}

	if v, ok := in["resource_group_name"].(string); ok && len(v) > 0 {
		obj.ResourceGroupName = v
	}

	if v, ok := in["managed_cluster"].([]interface{}); ok && len(v) > 0 {
		obj.ManagedCluster = expandAKSConfigManagedClusterV3(v)
	}

	// if v, ok := in["node_pools"].([]interface{}); ok && len(v) > 0 {
	// 	obj.NodePools = expandAKSNodePool(v)
	// }

	return obj
}

func expandAKSConfigManagedClusterV3(p []interface{}) *infrapb.Managedcluster {
	obj := &infrapb.Managedcluster{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["api_version"].(string); ok && len(v) > 0 {
		obj.ApiVersion = v
	}

	if v, ok := in["extended_location"].([]interface{}); ok && len(v) > 0 {
		obj.ExtendedLocation = expandAKSManagedClusterV3ExtendedLocation(v)
	}

	if v, ok := in["identity"].([]interface{}); ok && len(v) > 0 {
		obj.Identity = expandAKSManagedClusterV3Identity(v)
	}

	if v, ok := in["location"].(string); ok && len(v) > 0 {
		obj.Location = v
	}

	if v, ok := in["properties"].([]interface{}); ok && len(v) > 0 {
		obj.Properties = expandAKSManagedClusterProperties(v)
	}

	if v, ok := in["sku"].([]interface{}); ok && len(v) > 0 {
		obj.Sku = expandAKSManagedClusterV3SKU(v)
	}

	if v, ok := in["tags"].(map[string]string); ok {
		obj.Tags = v
	}

	if v, ok := in["type"].(string); ok && len(v) > 0 {
		obj.Type = v
	}

	if v, ok := in["additional_metadata"].([]interface{}); ok && len(v) > 0 {
		obj.AdditionalMetadata = expandAKSManagedClusterV3AdditionalMetadata(v)
	}

	return obj
}

func expandAKSManagedClusterV3ExtendedLocation(p []interface{}) *infrapb.Extendedlocation {
	obj := &infrapb.Extendedlocation{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["name"].(string); ok && len(v) > 0 {
		obj.Name = v
	}

	if v, ok := in["type"].(string); ok && len(v) > 0 {
		obj.Type = v
	}
	return obj
}

func expandAKSManagedClusterV3Identity(p []interface{}) *infrapb.Identity {
	obj := &infrapb.Identity{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["type"].(string); ok && len(v) > 0 {
		obj.Type = v
	}

	// TODO: TEST THIS CASE
	if v, ok := in["user_assigned_identities"].(map[string]interface{}); ok {
		obj.UserAssignedIdentities = toMapByte(v)
	}
	return obj
}

func expandAKSManagedClusterV3SKU(p []interface{}) *infrapb.Sku {
	obj := &infrapb.Sku{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["name"].(string); ok && len(v) > 0 {
		obj.Name = v
	}

	if v, ok := in["tier"].(string); ok && len(v) > 0 {
		obj.Tier = v
	}

	return obj
}

func expandAKSManagedClusterV3AdditionalMetadata(p []interface{}) *infrapb.Additionalmetadata {
	obj := &infrapb.Additionalmetadata{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["acr_profile"].([]interface{}); ok && len(v) > 0 {
		obj.AcrProfile = expandAKSManagedClusterV3AdditionalMetadataACRProfile(v)
	}

	if v, ok := in["oms_workspace_location"].(string); ok && len(v) > 0 {
		obj.OmsWorkspaceLocation = v
	}

	return obj
}

func expandAKSManagedClusterV3AdditionalMetadataACRProfile(p []interface{}) *infrapb.AcrProfile {
	obj := &infrapb.AcrProfile{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["resource_group_name"].(string); ok && len(v) > 0 {
		obj.ResourceGroupName = v
	}

	if v, ok := in["acr_name"].(string); ok && len(v) > 0 {
		obj.AcrName = v
	}

	return obj
}

func expandAKSManagedClusterV3Properties(p []interface{}) *infrapb.ManagedClusterProperties {
	obj := &infrapb.ManagedClusterProperties{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["aad_profile"].([]interface{}); ok && len(v) > 0 {
		obj.AadProfile = expandAKSManagedClusterV3AzureADProfile(v)
	}

	// TODO: GO BACK TO THIS
	// if v, ok := in["addon_profiles"].([]interface{}); ok && len(v) > 0 {
	// 	obj.AddonProfiles = expandAddonProfiles(v)
	// }

	if v, ok := in["api_server_access_profile"].([]interface{}); ok && len(v) > 0 {
		obj.ApiServerAccessProfile = expandAKSManagedClusterV3APIServerAccessProfile(v)
	}

	if v, ok := in["auto_scaler_profile"].([]interface{}); ok && len(v) > 0 {
		obj.AutoScalerProfile = expandAKSManagedClusterV3AutoScalerProfile(v)
	}

	if v, ok := in["auto_upgrade_profile"].([]interface{}); ok && len(v) > 0 {
		obj.AutoUpgradeProfile = expandAKSManagedClusterV3AutoUpgradeProfile(v)
	}

	if v, ok := in["disable_local_accounts"].(bool); ok {
		obj.DisableLocalAccounts = v
	}

	if v, ok := in["disk_encryption_set_id"].(string); ok {
		obj.DiskEncryptionSetID = v
	}

	if v, ok := in["dns_prefix"].(string); ok {
		obj.DnsPrefix = v
	}

	if v, ok := in["enable_pod_security_policy"].(bool); ok {
		obj.EnablePodSecurityPolicy = v
	}

	if v, ok := in["enable_rbac"].(bool); ok {
		obj.EnableRBAC = v
	}

	if v, ok := in["fqdn_subdomain"].(string); ok {
		obj.FqdnSubdomain = v
	}

	if v, ok := in["http_proxy_config"].([]interface{}); ok && len(v) > 0 {
		obj.HttpProxyConfig = expandAKSManagedClusterV3HTTPProxyConfig(v)
	}

	if v, ok := in["identity_profile"].(map[string]interface{}); ok {
		obj.IdentityProfile = toMapString(v)
	}

	if v, ok := in["kubernetes_version"].(string); ok {
		obj.KubernetesVersion = v
	}

	if v, ok := in["linux_profile"].([]interface{}); ok && len(v) > 0 {
		obj.LinuxProfile = expandAKSManagedClusterV3LinuxProfile(v)
	}

	if v, ok := in["network_profile"].([]interface{}); ok && len(v) > 0 {
		obj.NetworkProfile = expandAKSManagedClusterV3NetworkProfile(v)
	}

	if v, ok := in["node_resource_group"].(string); ok {
		obj.NodeResourceGroup = v
	}

	// TODO: STOPPED HERE
	if v, ok := in["pod_identity_profile"].([]interface{}); ok && len(v) > 0 {
		obj.PodIdentityProfile = expandAKSManagedClusterPodIdentityProfile(v)
	}

	if v, ok := in["private_link_resources"].([]interface{}); ok && len(v) > 0 {
		obj.PrivateLinkResources = expandAKSManagedClusterPrivateLinkResources(v)
	}

	if v, ok := in["service_principal_profile"].([]interface{}); ok && len(v) > 0 {
		obj.ServicePrincipalProfile = expandAKSManagedClusterServicePrincipalProfile(v)
	}

	if v, ok := in["windows_profile"].([]interface{}); ok && len(v) > 0 {
		obj.WindowsProfile = expandAKSManagedClusterWindowsProfile(v)
	}

	return obj
}

func expandAKSManagedClusterV3AzureADProfile(p []interface{}) *infrapb.Aadprofile {
	obj := &infrapb.Aadprofile{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["admin_group_object_ids"].([]interface{}); ok && len(v) > 0 {
		obj.AdminGroupObjectIDs = toArrayString(v)
	}

	if v, ok := in["client_app_id"].(string); ok && len(v) > 0 {
		obj.ClientAppID = v
	}

	if v, ok := in["enable_azure_rbac"].(bool); ok {
		obj.EnableAzureRBAC = v
	}

	if v, ok := in["managed"].(bool); ok {
		obj.Managed = v
	}

	if v, ok := in["server_app_id"].(string); ok && len(v) > 0 {
		obj.ServerAppID = v
	}

	if v, ok := in["server_app_secret"].(string); ok && len(v) > 0 {
		obj.ServerAppSecret = v
	}

	if v, ok := in["tenant_id"].(string); ok && len(v) > 0 {
		obj.TenantID = v
	}

	return obj
}

func expandAKSManagedClusterV3APIServerAccessProfile(p []interface{}) *infrapb.Apiserveraccessprofile {
	obj := &infrapb.Apiserveraccessprofile{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["authorized_ipr_ranges"].([]interface{}); ok && len(v) > 0 {
		obj.AuthorizedIPRanges = toArrayString(v)
	}

	if v, ok := in["enable_private_cluster"].(bool); ok {
		obj.EnablePrivateCluster = v
	}

	if v, ok := in["enable_private_cluster_public_fqdn"].(bool); ok {
		obj.EnablePrivateClusterPublicFQDN = v
	}

	if v, ok := in["private_dns_zone"].(string); ok && len(v) > 0 {
		obj.PrivateDNSZone = v
	}
	return obj
}

func expandAKSManagedClusterV3AutoScalerProfile(p []interface{}) *infrapb.Autoscalerprofile {
	obj := &infrapb.Autoscalerprofile{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["balance_similar_node_groups"].(string); ok && len(v) > 0 {
		obj.BalanceSimilarNodeGroups = v
	}

	if v, ok := in["expander"].(string); ok && len(v) > 0 {
		obj.Expander = v
	}

	if v, ok := in["max_empty_bulk_delete"].(string); ok && len(v) > 0 {
		obj.MaxEmptyBulkDelete = v
	}

	if v, ok := in["max_graceful_termination_sec"].(string); ok && len(v) > 0 {
		obj.MaxGracefulTerminationSec = v
	}

	if v, ok := in["max_node_provision_time"].(string); ok && len(v) > 0 {
		obj.MaxNodeProvisionTime = v
	}

	if v, ok := in["max_total_unready_percentage"].(string); ok && len(v) > 0 {
		obj.MaxTotalUnreadyPercentage = v
	}

	if v, ok := in["new_pod_scale_up_delay"].(string); ok && len(v) > 0 {
		obj.NewPodScaleUpDelay = v
	}

	if v, ok := in["ok_total_unready_count"].(string); ok && len(v) > 0 {
		obj.OkTotalUnreadyCount = v
	}

	if v, ok := in["scale_down_delay_after_add"].(string); ok && len(v) > 0 {
		obj.ScaleDownDelayAfterAdd = v
	}

	if v, ok := in["scale_down_delay_after_delete"].(string); ok && len(v) > 0 {
		obj.ScaleDownDelayAfterDelete = v
	}

	if v, ok := in["scale_down_delay_after_failure"].(string); ok && len(v) > 0 {
		obj.ScaleDownDelayAfterFailure = v
	}

	if v, ok := in["scale_down_unneeded_time"].(string); ok && len(v) > 0 {
		obj.ScaleDownUnneededTime = v
	}

	if v, ok := in["scale_down_unready_time"].(string); ok && len(v) > 0 {
		obj.ScaleDownUnreadyTime = v
	}

	if v, ok := in["scale_down_utilization_threshold"].(string); ok && len(v) > 0 {
		obj.ScaleDownUtilizationThreshold = v
	}

	if v, ok := in["scan_interval"].(string); ok && len(v) > 0 {
		obj.ScanInterval = v
	}

	if v, ok := in["skip_nodes_with_local_storage"].(string); ok && len(v) > 0 {
		obj.SkipNodesWithLocalStorage = v
	}

	if v, ok := in["skip_nodes_with_system_pods"].(string); ok && len(v) > 0 {
		obj.SkipNodesWithSystemPods = v
	}
	return obj
}

func expandAKSManagedClusterV3AutoUpgradeProfile(p []interface{}) *infrapb.Autoupgradeprofile {
	obj := &infrapb.Autoupgradeprofile{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["upgrade_channel"].(string); ok && len(v) > 0 {
		obj.UpgradeChannel = v
	}
	return obj
}

func expandAKSManagedClusterV3HTTPProxyConfig(p []interface{}) *infrapb.Httpproxyconfig {
	obj := &infrapb.Httpproxyconfig{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["http_proxy"].(string); ok && len(v) > 0 {
		obj.HttpProxy = v
	}

	if v, ok := in["https_proxy"].(string); ok && len(v) > 0 {
		obj.HttpsProxy = v
	}

	if v, ok := in["no_proxy"].([]interface{}); ok && len(v) > 0 {
		obj.NoProxy = toArrayString(v)
	}

	if v, ok := in["trusted_ca"].(string); ok && len(v) > 0 {
		obj.TrustedCa = v
	}

	return obj
}

func expandAKSManagedClusterV3LinuxProfile(p []interface{}) *infrapb.Linuxprofile {
	obj := &infrapb.Linuxprofile{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["admin_username"].(string); ok && len(v) > 0 {
		obj.AdminUsername = v
	}

	if v, ok := in["ssh"].([]interface{}); ok && len(v) > 0 {
		obj.Ssh = expandAKSManagedClusterV3SSHConfig(v)
	}
	return obj
}

func expandAKSManagedClusterV3SSHConfig(p []interface{}) *infrapb.Ssh {
	obj := &infrapb.Ssh{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["public_keys"].([]interface{}); ok && len(v) > 0 {
		obj.PublicKeys = expandAKSManagedClusterV3LPSSHKeyData(v)
	}

	return obj

}

func expandAKSManagedClusterV3LPSSHKeyData(p []interface{}) []*infrapb.Publickeys {
	if len(p) == 0 || p[0] == nil {
		return []*infrapb.Publickeys{}
	}
	out := make([]*infrapb.Publickeys, len(p))

	for i := range p {
		obj := infrapb.Publickeys{}
		in := p[i].(map[string]interface{})

		if v, ok := in["key_data"].(string); ok {
			obj.KeyData = v
		}
		out[i] = &obj
	}
	return out
}

func expandAKSManagedClusterV3NetworkProfile(p []interface{}) *infrapb.Networkprofile {
	obj := &infrapb.Networkprofile{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["dns_service_ip"].(string); ok && len(v) > 0 {
		obj.DnsServiceIP = v
	}

	if v, ok := in["docker_bridge_cidr"].(string); ok && len(v) > 0 {
		obj.DockerBridgeCidr = v
	}

	if v, ok := in["load_balancer_profile"].([]interface{}); ok && len(v) > 0 {
		obj.LoadBalancerProfile = expandAKSManagedClusterV3NPLoadBalancerProfile(v)
	}

	if v, ok := in["load_balancer_sku"].(string); ok && len(v) > 0 {
		obj.LoadBalancerSku = v
	}

	if v, ok := in["network_mode"].(string); ok && len(v) > 0 {
		obj.NetworkMode = v
	}

	if v, ok := in["network_plugin"].(string); ok && len(v) > 0 {
		obj.NetworkPlugin = v
	}

	if v, ok := in["network_policy"].(string); ok && len(v) > 0 {
		obj.NetworkPolicy = v
	}

	if v, ok := in["outbound_type"].(string); ok && len(v) > 0 {
		obj.OutboundType = v
	}

	if v, ok := in["pod_cidr"].(string); ok && len(v) > 0 {
		obj.PodCidr = v
	}

	if v, ok := in["service_cidr"].(string); ok && len(v) > 0 {
		obj.ServiceCidr = v
	}
	return obj
}

func expandAKSManagedClusterV3NPLoadBalancerProfile(p []interface{}) *infrapb.Loadbalancerprofile {
	obj := &infrapb.Loadbalancerprofile{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["allocated_outbound_ports"].(uint32); ok && v > 0 {
		obj.AllocatedOutboundPorts = v
	}

	if v, ok := in["effective_outbound_ips"].([]interface{}); ok && len(v) > 0 {
		obj.EffectiveOutboundIPs = expandAKSManagedClusterV3NPEffectiveOutboundIPs(v)
	}

	if v, ok := in["idle_timeout_in_minutes"].(int); ok && v > 0 {
		obj.IdleTimeoutInMinutes = uint32(v)
	}

	if v, ok := in["managed_outbound_ips"].([]interface{}); ok && len(v) > 0 {
		obj.ManagedOutboundIPs = expandAKSManagedClusterV3NPManagedOutboundIPs(v)
	}

	if v, ok := in["outbound_ip_prefixes"].([]interface{}); ok && len(v) > 0 {
		obj.OutboundIPPrefixes = expandAKSManagedClusterV3NPOutboundIPPrefixes(v)
	}

	if v, ok := in["outbound_ips"].([]interface{}); ok && len(v) > 0 {
		obj.OutboundIPs = expandAKSManagedClusterV3NPOutboundIPs(v)
	}

	return obj
}

func expandAKSManagedClusterV3NPEffectiveOutboundIPs(p []interface{}) []*infrapb.Effectiveoutboundips {
	if len(p) == 0 || p[0] == nil {
		return []*infrapb.Effectiveoutboundips{}
	}
	out := make([]*infrapb.Effectiveoutboundips, len(p))

	for i := range p {
		obj := infrapb.Effectiveoutboundips{}
		in := p[i].(map[string]interface{})

		if v, ok := in["id"].(string); ok {
			obj.Id = v
		}
		out[i] = &obj
	}
	return out
}

func expandAKSManagedClusterV3NPManagedOutboundIPs(p []interface{}) *infrapb.Managedoutboundips {
	obj := &infrapb.Managedoutboundips{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["count"].(uint32); ok && v > 0 {
		obj.Count = v
	}
	return obj
}

func expandAKSManagedClusterV3NPOutboundIPPrefixes(p []interface{}) *infrapb.Outboundipprefixes {
	obj := &infrapb.Outboundipprefixes{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["public_ip_prefixes"].([]interface{}); ok && len(v) > 0 {
		obj.PublicIPPrefixes = expandAKSManagedClusterV3NPManagedOutboundIPsPublicIpPrefixes(v)
	}
	return obj
}

func expandAKSManagedClusterV3NPManagedOutboundIPsPublicIpPrefixes(p []interface{}) []*infrapb.Publicipprefixes {
	if len(p) == 0 || p[0] == nil {
		return []*infrapb.Publicipprefixes{}
	}
	out := make([]*infrapb.Publicipprefixes, len(p))

	for i := range p {
		obj := infrapb.Publicipprefixes{}
		in := p[i].(map[string]interface{})

		if v, ok := in["id"].(string); ok {
			obj.Id = v
		}
		out[i] = &obj
	}
	return out
}

func expandAKSManagedClusterV3NPOutboundIPs(p []interface{}) *infrapb.Outboundips {
	obj := &infrapb.Outboundips{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["public_ips"].([]interface{}); ok && len(v) > 0 {
		obj.PublicIPs = expandAKSManagedClusterV3NPOutboundIPsPublicIps(v)
	}
	return obj
}

func expandAKSManagedClusterV3NPOutboundIPsPublicIps(p []interface{}) []*infrapb.Publicips {
	if len(p) == 0 || p[0] == nil {
		return []*infrapb.Publicips{}
	}
	out := make([]*infrapb.Publicips, len(p))

	for i := range p {
		obj := infrapb.Publicips{}
		in := p[i].(map[string]interface{})

		if v, ok := in["id"].(string); ok {
			obj.Id = v
		}
		out[i] = &obj
	}
	return out
}
