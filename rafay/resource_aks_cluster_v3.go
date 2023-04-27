package rafay

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	dynamic "github.com/RafaySystems/rafay-common/pkg/hub/client/dynamic"
	"github.com/RafaySystems/rafay-common/pkg/hub/client/options"
	typed "github.com/RafaySystems/rafay-common/pkg/hub/client/typed"
	"github.com/RafaySystems/rafay-common/pkg/hub/terraform/resource"
	"github.com/RafaySystems/rafay-common/proto/types/hub/commonpb"
	"github.com/RafaySystems/rafay-common/proto/types/hub/infrapb"
	"github.com/RafaySystems/rctl/pkg/config"
	"github.com/RafaySystems/rctl/pkg/versioninfo"
	"github.com/davecgh/go-spew/spew"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	schema "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	structpb "google.golang.org/protobuf/types/known/structpb"
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
			Create: schema.DefaultTimeout(90 * time.Minute),
			Update: schema.DefaultTimeout(90 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema:        resource.ClusterSchema.Schema,
	}
}

func resourceAKSClusterV3Create(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("Cluster create starts")

	diags := resourceAKSClusterV3Upsert(ctx, d, m)
	return diags
}

func resourceAKSClusterV3Read(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	log.Println("resourceClusterRead ")
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}
	tfClusterState, err := expandClusterV3(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}

	ag, err := client.InfraV3().Cluster().Get(ctx, options.GetOptions{
		Name:    tfClusterState.Metadata.Name,
		Project: tfClusterState.Metadata.Project,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	err = flattenAKSClusterV3(d, ag)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceAKSClusterV3Update(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("Cluster update starts")
	return resourceAKSClusterV3Upsert(ctx, d, m)
}

func resourceAKSClusterV3Delete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("Cluster delete starts")

	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	ag, err := expandClusterV3(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.InfraV3().Cluster().Delete(ctx, options.DeleteOptions{
		Name:    ag.Metadata.Name,
		Project: ag.Metadata.Project,
	})
	if err != nil {
		log.Printf("cluster delete failed for edgename: %s and projectname: %s", ag.Metadata.Name, ag.Metadata.Project)
		return diag.FromErr(err)
	}

	ticker := time.NewTicker(time.Duration(30) * time.Second)
	defer ticker.Stop()
	timeout := time.After(time.Duration(10) * time.Minute)

	edgeName := ag.Metadata.Name
	projectName := ag.Metadata.Project

LOOP:
	for {
		select {
		case <-timeout:
			log.Printf("Cluster Deletion for edgename: %s and projectname: %s got timeout out.", edgeName, projectName)
			return diag.FromErr(fmt.Errorf("cluster deletion for edgename: %s and projectname: %s got timeout out", edgeName, projectName))
		case <-ticker.C:
			_, err := client.InfraV3().Cluster().Get(ctx, options.GetOptions{
				Name:    edgeName,
				Project: projectName,
			})
			if dErr, ok := err.(*dynamic.DynamicClientGetError); ok && dErr != nil {
				switch dErr.StatusCode {
				case http.StatusNotFound:
					log.Printf("Cluster Deletion completes for edgename: %s and projectname: %s", edgeName, projectName)
					break LOOP
				default:
					log.Printf("Cluster Deletion failed for edgename: %s and projectname: %s with error: %s", edgeName, projectName, dErr.Error())
					return diag.FromErr(dErr)
				}
			}
			log.Printf("Cluster Deletion is in progress for edgename: %s and projectname: %s", edgeName, projectName)
		}
	}

	return diags
}

func resourceAKSClusterV3Import(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	idParts := strings.SplitN(d.Id(), "/", 2)
	log.Println("resourceAKSClusterV3 idParts:", idParts)

	cluster, err := expandClusterV3(d)
	if err != nil {
		log.Printf("resourceCluster expand error")
	}

	var metaD commonpb.Metadata
	metaD.Name = idParts[0]
	metaD.Project = idParts[1]
	cluster.Metadata = &metaD

	err = d.Set("metadata", flattenMetaData(cluster.Metadata))
	if err != nil {
		log.Println("import set err")
		return nil, err
	}
	d.SetId(cluster.Metadata.Name)
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

	log.Println(">>>>>> CLUSTER: ", cluster)

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
	// wait for cluster creation
	ticker := time.NewTicker(time.Duration(60) * time.Second)
	defer ticker.Stop()
	timeout := time.After(time.Duration(90) * time.Minute)

	edgeName := cluster.Metadata.Name
	projectName := cluster.Metadata.Project
	d.SetId(cluster.Metadata.Name)

LOOP:
	for {
		select {
		case <-timeout:
			log.Printf("Cluster operation timed out for edgeName: %s and projectname: %s", edgeName, projectName)
			return diag.FromErr(fmt.Errorf("cluster operation timed out for edgeName: %s and projectname: %s", edgeName, projectName))
		case <-ticker.C:
			uCluster, err2 := client.InfraV3().Cluster().Status(ctx, options.StatusOptions{
				Name:    edgeName,
				Project: projectName,
			})
			if err2 != nil {
				log.Printf("Fetching cluster having edgename: %s and projectname: %s failing due to err: %v", edgeName, projectName, err2)
				return diag.FromErr(err2)
			} else if uCluster == nil {
				log.Printf("Cluster operation has not started with edgename: %s and projectname: %s", edgeName, projectName)
			} else if uCluster.Status != nil && uCluster.Status.Aks != nil && uCluster.Status.CommonStatus != nil {
				aksStatus := uCluster.Status.Aks
				uClusterCommonStatus := uCluster.Status.CommonStatus
				switch uClusterCommonStatus.ConditionStatus {
				case commonpb.ConditionStatus_StatusSubmitted:
					log.Printf("Cluster operation not completed for edgename: %s and projectname: %s. Waiting 60 seconds more for cluster to complete the operation.", edgeName, projectName)
				case commonpb.ConditionStatus_StatusOK:
					log.Printf("Cluster operation completed for edgename: %s and projectname: %s", edgeName, projectName)
					break LOOP
				case commonpb.ConditionStatus_StatusFailed:
					// log.Printf("Cluster operation failed for edgename: %s and projectname: %s with failure reason: %s", edgeName, projectName, uClusterCommonStatus.Reason)
					failureReasons, err := collectAKSUpsertErrors(aksStatus.Nodepools, uCluster.Status.LastProvisionFailureReason, uCluster.Status.ProvisionStatus)
					if err != nil {
						return diag.FromErr(err)
					}
					return diag.Errorf("Cluster operation failed for edgename: %s and projectname: %s with failure reasons: %s", edgeName, projectName, failureReasons)
				}
			}
		}
	}
	return diags
}

func collectAKSUpsertErrors(nodepools []*infrapb.NodepoolStatus, lastProvisionFailureReason string, provisionStatus string) (string, error) {
	// Defining local struct just to collect errors in json-prettify format to display the same to end user for better visualization.
	type AksNodepoolsErrorFormatter struct {
		Name          string `json:"name"`
		FailureReason string `json:"failureReason"`
	}

	type AksUpsertErrorFormatter struct {
		FailureReason string                       `json:"failureReason"`
		Nodepools     []AksNodepoolsErrorFormatter `json:"nodepools"`
	}

	// adding errors in AksUpsertErrorFormatter
	collectedErrors := AksUpsertErrorFormatter{}
	if len(lastProvisionFailureReason) > 0 || provisionStatus == "cluster operation failed" {
		collectedErrors.FailureReason = lastProvisionFailureReason
	}
	collectedErrors.Nodepools = []AksNodepoolsErrorFormatter{}
	for _, ng := range nodepools {
		if len(ng.LastProvisionFailureReason) > 0 || ng.ProvisionStatus == "nodegroup operation failed" {
			collectedErrors.Nodepools = append(collectedErrors.Nodepools, AksNodepoolsErrorFormatter{
				Name:          ng.Name,
				FailureReason: ng.LastProvisionFailureReason,
			})
		}
	}
	// Using MarshalIndent to indent the errors in json formatted bytes
	collectedErrsFormattedBytes, err := json.MarshalIndent(collectedErrors, "", "    ")
	if err != nil {
		return "", err
	}
	fmt.Println("After MarshalIndent: ", "collectedErrsFormattedBytes", string(collectedErrsFormattedBytes))
	return string(collectedErrsFormattedBytes), nil
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
		return nil, errors.New("cluster type not implemented")
	}

	if v, ok := in["sharing"].([]interface{}); ok && len(v) > 0 {
		obj.Sharing = expandSharingSpecV3(v)
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
		obj.Aks.Metadata = expandAKSClusterV3ConfigMetaData(v)
	}

	if v, ok := in["spec"].([]interface{}); ok && len(v) > 0 {
		obj.Aks.Spec = expandAKSClusterV3ConfigSpec(v)
	}

	return obj
}

func expandAKSClusterV3ConfigMetaData(p []interface{}) *infrapb.AksV3ConfigMetadata {
	obj := &infrapb.AksV3ConfigMetadata{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["name"].(string); ok && len(v) > 0 {
		obj.Name = v
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

	if v, ok := in["node_pools"].([]interface{}); ok && len(v) > 0 {
		obj.NodePools = expandAKSV3NodePool(v)
	}

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
		obj.Properties = expandAKSManagedClusterV3Properties(v)
	}

	if v, ok := in["sku"].([]interface{}); ok && len(v) > 0 {
		obj.Sku = expandAKSManagedClusterV3SKU(v)
	}

	if v, ok := in["tags"].(map[string]interface{}); ok {
		obj.Tags = toMapString(v)
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
	if v, ok := in["user_assigned_identities"].(map[string]interface{}); ok {
		obj.UserAssignedIdentities = make(map[string]*structpb.Struct)
		for identity := range v {
			x, _ := structpb.NewStruct(map[string]interface{}{})
			obj.UserAssignedIdentities[identity] = x
		}
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

	if v, ok := in["addon_profiles"].([]interface{}); ok && len(v) > 0 {
		obj.AddonProfiles = expandManagedClusterAddonProfile(v)
	}

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

	if v, ok := in["pod_identity_profile"].([]interface{}); ok && len(v) > 0 {
		obj.PodIdentityProfile = expandAKSManagedClusterV3PodIdentityProfile(v)
	}

	if v, ok := in["private_link_resources"].([]interface{}); ok && len(v) > 0 {
		obj.PrivateLinkResources = expandAKSV3ManagedClusterPrivateLinkResources(v)
	}

	if v, ok := in["service_principal_profile"].([]interface{}); ok && len(v) > 0 {
		obj.ServicePrincipalProfile = expandAKSManagedClusterV3ServicePrincipalProfile(v)
	}

	if v, ok := in["windows_profile"].([]interface{}); ok && len(v) > 0 {
		obj.WindowsProfile = expandAKSManagedClusterV3WindowsProfile(v)
	}

	return obj
}

func expandManagedClusterAddonProfile(p []interface{}) *infrapb.ManagedClusterAddonProfile {
	obj := &infrapb.ManagedClusterAddonProfile{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["http_application_routing"].([]interface{}); ok && len(v) > 0 {
		obj.HttpApplicationRouting = expandAddonProfileGeneric(v)
	}

	if v, ok := in["azure_policy"].([]interface{}); ok && len(v) > 0 {
		obj.AzurePolicy = expandAddonProfileGeneric(v)
	}

	if v, ok := in["azure_keyvault_secrets_provider"].([]interface{}); ok && len(v) > 0 {
		obj.AzureKeyvaultSecretsProvider = expandAddonProfileAzureKeyvaultSecretsProvider(v)
	}

	if v, ok := in["oms_agent"].([]interface{}); ok && len(v) > 0 {
		obj.OmsAgent = expandAddonProfileOmsAgent(v)
	}

	if v, ok := in["ingress_application_gateway"].([]interface{}); ok && len(v) > 0 {
		obj.IngressApplicationGateway = expandAddonProfileIngressApplicationGateway(v)
	}
	return obj
}

func expandAddonProfileGeneric(p []interface{}) *infrapb.AddonProfileGeneric {
	obj := &infrapb.AddonProfileGeneric{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["enabled"].(bool); ok {
		obj.Enabled = v
	}
	return obj
}

func expandAddonProfileAzureKeyvaultSecretsProvider(p []interface{}) *infrapb.AddonProfileAzureKeyvaultSecretsProvider {
	obj := &infrapb.AddonProfileAzureKeyvaultSecretsProvider{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["enabled"].(bool); ok {
		obj.Enabled = v
	}

	if v, ok := in["config"].([]interface{}); ok && len(v) > 0 {
		obj.Config = expandAddonProfileAzureKeyvaultSecretsProviderConfig(v)
	}
	return obj
}

func expandAddonProfileAzureKeyvaultSecretsProviderConfig(p []interface{}) *infrapb.AzureKeyvaultSecretsProviderProfileConfig {
	obj := &infrapb.AzureKeyvaultSecretsProviderProfileConfig{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["enable_secret_rotation"].(string); ok && len(v) > 0 {
		obj.EnableSecretRotation = v
	}

	if v, ok := in["rotation_poll_interval"].(string); ok && len(v) > 0 {
		obj.RotationPollInterval = v
	}
	return obj
}

func expandAddonProfileOmsAgent(p []interface{}) *infrapb.AddonProfileOmsAgent {
	obj := &infrapb.AddonProfileOmsAgent{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["enabled"].(bool); ok {
		obj.Enabled = v
	}

	if v, ok := in["config"].([]interface{}); ok && len(v) > 0 {
		obj.Config = expandAddonProfileOmsAgentConfig(v)
	}
	return obj
}

func expandAddonProfileOmsAgentConfig(p []interface{}) *infrapb.AddonProfileOmsAgentConfig {
	obj := &infrapb.AddonProfileOmsAgentConfig{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["log_analytics_workspace_resource_id"].(string); ok && len(v) > 0 {
		obj.LogAnalyticsWorkspaceResourceID = v
	}
	return obj
}

func expandAddonProfileIngressApplicationGateway(p []interface{}) *infrapb.AddonProfileIngressApplicationGateway {
	obj := &infrapb.AddonProfileIngressApplicationGateway{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["enabled"].(bool); ok {
		obj.Enabled = v
	}

	if v, ok := in["config"].([]interface{}); ok && len(v) > 0 {
		obj.Config = expandAddonProfileIngressApplicationGatewayConfig(v)
	}
	return obj
}

func expandAddonProfileIngressApplicationGatewayConfig(p []interface{}) *infrapb.AddonProfileIngressApplicationGatewayConfig {
	obj := &infrapb.AddonProfileIngressApplicationGatewayConfig{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["application_gateway_name"].(string); ok && len(v) > 0 {
		obj.ApplicationGatewayName = v
	}

	if v, ok := in["application_gateway_id"].(string); ok && len(v) > 0 {
		obj.ApplicationGatewayID = v
	}

	if v, ok := in["subnet_cidr"].(string); ok && len(v) > 0 {
		obj.SubnetCIDR = v
	}

	if v, ok := in["subnet_id"].(string); ok && len(v) > 0 {
		obj.SubnetID = v
	}

	if v, ok := in["watch_namespace"].(string); ok && len(v) > 0 {
		obj.WatchNamespace = v
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

	if v, ok := in["allocated_outbound_ports"].(int); ok && v > 0 {
		obj.AllocatedOutboundPorts = uint32(v)
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

	if v, ok := in["count"].(int); ok && v > 0 {
		obj.Count = uint32(v)
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

func expandAKSManagedClusterV3PodIdentityProfile(p []interface{}) *infrapb.Podidentityprofile {
	obj := &infrapb.Podidentityprofile{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["allow_network_plugin_kubenet"].(bool); ok {
		obj.AllowNetworkPluginKubenet = v
	}

	if v, ok := in["enabled"].(bool); ok {
		obj.Enabled = v
	}

	if v, ok := in["user_assigned_identities"].([]interface{}); ok && len(v) > 0 {
		obj.UserAssignedIdentities = expandAKSManagedClusterV3PIPUserAssignedIdentities(v)
	}

	if v, ok := in["user_assigned_identity_exceptions"].([]interface{}); ok && len(v) > 0 {
		obj.UserAssignedIdentityExceptions = expandAKSManagedClusterV3PIPUserAssignedIdentityExceptions(v)
	}

	return obj
}

func expandAKSManagedClusterV3PIPUserAssignedIdentities(p []interface{}) []*infrapb.PodUserassignedidentities {
	out := make([]*infrapb.PodUserassignedidentities, len(p))
	if len(p) == 0 || p[0] == nil {
		return out
	}
	for i := range p {
		obj := &infrapb.PodUserassignedidentities{}
		in := p[i].(map[string]interface{})

		if v, ok := in["binding_selector"].(string); ok && len(v) > 0 {
			obj.BindingSelector = v
		}

		if v, ok := in["identity"].([]interface{}); ok && len(v) > 0 {
			obj.Identity = expandAKSManagedClusterV3UAIIdentity(v)
		}

		if v, ok := in["name"].(string); ok && len(v) > 0 {
			obj.Name = v
		}

		if v, ok := in["namespace"].(string); ok && len(v) > 0 {
			obj.Namespace = v
		}
		out[i] = obj
	}
	return out
}

func expandAKSManagedClusterV3UAIIdentity(p []interface{}) *infrapb.PodIdentity {
	obj := &infrapb.PodIdentity{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["client_id"].(string); ok && len(v) > 0 {
		obj.ClientId = v
	}

	if v, ok := in["object_id"].(string); ok && len(v) > 0 {
		obj.ObjectId = v
	}

	if v, ok := in["resource_id"].(string); ok && len(v) > 0 {
		obj.ResourceId = v
	}
	return obj
}

func expandAKSManagedClusterV3PIPUserAssignedIdentityExceptions(p []interface{}) []*infrapb.Userassignedidentityexceptions {
	out := make([]*infrapb.Userassignedidentityexceptions, len(p))
	if len(p) == 0 || p[0] == nil {
		return out
	}
	for i := range p {
		obj := &infrapb.Userassignedidentityexceptions{}
		in := p[i].(map[string]interface{})

		if v, ok := in["name"].(string); ok && len(v) > 0 {
			obj.Name = v
		}

		if v, ok := in["namespace"].(string); ok && len(v) > 0 {
			obj.Namespace = v
		}

		if v, ok := in["pod_labels"].(map[string]interface{}); ok {
			obj.PodLabels = toMapString(v)
		}
		out[i] = obj
	}
	return out
}

func expandAKSV3ManagedClusterPrivateLinkResources(p []interface{}) []*infrapb.Privatelinkresources {
	out := make([]*infrapb.Privatelinkresources, len(p))
	if len(p) == 0 || p[0] == nil {
		return out
	}

	for i := range p {
		obj := &infrapb.Privatelinkresources{}
		in := p[i].(map[string]interface{})
		if v, ok := in["group_id"].(string); ok && len(v) > 0 {
			obj.GroupId = v
		}

		if v, ok := in["id"].(string); ok && len(v) > 0 {
			obj.Id = v
		}

		if v, ok := in["name"].(string); ok && len(v) > 0 {
			obj.Name = v
		}

		if v, ok := in["required_members"].([]interface{}); ok && len(v) > 0 {
			obj.RequiredMembers = toArrayString(v)
		}

		if v, ok := in["type"].(string); ok && len(v) > 0 {
			obj.Type = v
		}

		out[i] = obj
	}

	return out
}

func expandAKSManagedClusterV3ServicePrincipalProfile(p []interface{}) *infrapb.Serviceprincipalprofile {
	obj := &infrapb.Serviceprincipalprofile{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["client_id"].(string); ok && len(v) > 0 {
		obj.ClientId = v
	}

	if v, ok := in["secret"].(string); ok && len(v) > 0 {
		obj.Secret = v
	}

	return obj
}

func expandAKSManagedClusterV3WindowsProfile(p []interface{}) *infrapb.Windowsprofile {
	obj := &infrapb.Windowsprofile{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["admin_username"].(string); ok && len(v) > 0 {
		obj.AdminUsername = v
	}

	if v, ok := in["enable_csi_proxy"].(bool); ok {
		obj.EnableCSIProxy = v
	}

	if v, ok := in["license_type"].(string); ok && len(v) > 0 {
		obj.LicenseType = v
	}
	return obj
}

func expandAKSV3NodePool(p []interface{}) []*infrapb.Nodepool {
	if len(p) == 0 || p[0] == nil {
		return []*infrapb.Nodepool{}
	}

	out := make([]*infrapb.Nodepool, len(p))
	for i := range p {
		obj := infrapb.Nodepool{}
		in := p[i].(map[string]interface{})

		if v, ok := in["api_version"].(string); ok && len(v) > 0 {
			obj.ApiVersion = v
		}

		if v, ok := in["name"].(string); ok && len(v) > 0 {
			obj.Name = v
		}

		if v, ok := in["properties"].([]interface{}); ok && len(v) > 0 {
			obj.Properties = expandAKSV3NodePoolProperties(v)
		}

		if v, ok := in["type"].(string); ok && len(v) > 0 {
			obj.Type = v
		}

		if v, ok := in["location"].(string); ok && len(v) > 0 {
			obj.Location = v
		}
		out[i] = &obj
	}

	return out
}

func expandAKSV3NodePoolProperties(p []interface{}) *infrapb.NodePoolProperties {
	obj := &infrapb.NodePoolProperties{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["availability_zones"].([]interface{}); ok && len(v) > 0 {
		obj.AvailabilityZones = toArrayString(v)
	}

	if v, ok := in["count"].(int); ok && v > 0 {
		obj.Count = int32(v)
	}

	if v, ok := in["enable_auto_scaling"].(bool); ok {
		obj.EnableAutoScaling = v
	}

	if v, ok := in["enable_encryption_at_host"].(bool); ok {
		obj.EnableEncryptionAtHost = v
	}

	if v, ok := in["enable_fips"].(bool); ok {
		obj.EnableFIPS = v
	}

	if v, ok := in["enable_node_public_ip"].(bool); ok {
		obj.EnableNodePublicIP = v
	}

	if v, ok := in["enable_ultra_ssd"].(bool); ok {
		obj.EnableUltraSSD = v
	}

	if v, ok := in["gpu_instance_profile"].(string); ok && len(v) > 0 {
		obj.GpuInstanceProfile = v
	}

	if v, ok := in["kubelet_config"].([]interface{}); ok && len(v) > 0 {
		obj.KubeletConfig = expandAKSV3NodePoolKubeletConfig(v)
	}

	if v, ok := in["kubelet_disk_type"].(string); ok && len(v) > 0 {
		obj.KubeletDiskType = v
	}

	if v, ok := in["linux_os_config"].([]interface{}); ok && len(v) > 0 {
		obj.LinuxOSConfig = expandAKSV3NodePoolLinuxOsConfig(v)
	}

	if v, ok := in["max_count"].(int); ok && v > 0 {
		obj.MaxCount = int32(v)
	}

	if v, ok := in["max_pods"].(int); ok && v > 0 {
		obj.MaxPods = int32(v)
	}

	if v, ok := in["min_count"].(int); ok && v > 0 {
		obj.MinCount = int32(v)
	}

	if v, ok := in["mode"].(string); ok && len(v) > 0 {
		obj.Mode = v
	}

	if v, ok := in["node_labels"].(map[string]interface{}); ok {
		obj.NodeLabels = toMapString(v)
	}

	if v, ok := in["node_public_ip_prefix_id"].(string); ok && len(v) > 0 {
		obj.NodePublicIPPrefixID = v
	}

	if v, ok := in["node_taints"].([]interface{}); ok && len(v) > 0 {
		obj.NodeTaints = toArrayString(v)
	}

	if v, ok := in["orchestrator_version"].(string); ok && len(v) > 0 {
		obj.OrchestratorVersion = v
	}

	if v, ok := in["os_disk_size_gb"].(int); ok && v > 0 {
		obj.OsDiskSizeGB = int32(v)
	}

	if v, ok := in["os_disk_type"].(string); ok && len(v) > 0 {
		obj.OsDiskType = v
	}

	if v, ok := in["os_sku"].(string); ok && len(v) > 0 {
		obj.OsSKU = v
	}

	if v, ok := in["os_type"].(string); ok && len(v) > 0 {
		obj.OsType = v
	}

	if v, ok := in["pod_subnet_id"].(string); ok && len(v) > 0 {
		obj.PodSubnetID = v
	}

	if v, ok := in["proximity_placement_group_id"].(string); ok && len(v) > 0 {
		obj.ProximityPlacementGroupID = v
	}

	if v, ok := in["scale_set_eviction_policy"].(string); ok && len(v) > 0 {
		obj.ScaleSetEvictionPolicy = v
	}

	if v, ok := in["scale_set_priority"].(string); ok && len(v) > 0 {
		obj.ScaleSetPriority = v
	}

	if v, ok := in["spot_max_price"].(string); ok && len(v) > 0 {
		obj.Type = v
	}

	if v, ok := in["tags"].(map[string]interface{}); ok {
		obj.Tags = toMapString(v)
	}

	if v, ok := in["type"].(string); ok && len(v) > 0 {
		obj.Type = v
	}

	if v, ok := in["upgrade_settings"].([]interface{}); ok && len(v) > 0 {
		obj.UpgradeSettings = expandAKSV3NodePoolUpgradeSettings(v)
	}

	if v, ok := in["vm_size"].(string); ok && len(v) > 0 {
		obj.VmSize = v
	}

	if v, ok := in["vnet_subnet_id"].(string); ok && len(v) > 0 {
		obj.VnetSubnetID = v
	}

	return obj
}

func expandAKSV3NodePoolKubeletConfig(p []interface{}) *infrapb.Kubeletconfig {
	obj := &infrapb.Kubeletconfig{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["allowed_unsafe_sysctls"].([]interface{}); ok && len(v) > 0 {
		obj.AllowedUnsafeSysctls = toArrayString(v)
	}

	if v, ok := in["container_log_max_files"].(int); ok && v > 0 {
		obj.ContainerLogMaxFiles = int32(v)
	}

	if v, ok := in["container_log_max_size_mb"].(int); ok && v > 0 {
		obj.ContainerLogMaxSizeMB = int32(v)
	}

	if v, ok := in["cpu_cfs_quota"].(bool); ok {
		obj.CpuCfsQuota = v
	}

	if v, ok := in["cpu_cfs_quota_period"].(string); ok && len(v) > 0 {
		obj.CpuCfsQuotaPeriod = v
	}

	if v, ok := in["cpu_manager_policy"].(string); ok && len(v) > 0 {
		obj.CpuManagerPolicy = v
	}

	if v, ok := in["fail_swap_on"].(bool); ok {
		obj.FailSwapOn = v
	}

	if v, ok := in["image_gc_high_threshold"].(int); ok && v > 0 {
		obj.ImageGcHighThreshold = int32(v)
	}

	if v, ok := in["image_gc_low_threshold"].(int); ok && v > 0 {
		obj.ImageGcLowThreshold = int32(v)
	}

	if v, ok := in["pod_max_pids"].(int); ok && v > 0 {
		obj.PodMaxPids = int32(v)
	}

	if v, ok := in["topology_manager_policy"].(string); ok && len(v) > 0 {
		obj.TopologyManagerPolicy = v
	}

	return obj
}

func expandAKSV3NodePoolLinuxOsConfig(p []interface{}) *infrapb.Linuxosconfig {
	obj := &infrapb.Linuxosconfig{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["swap_file_size_mb"].(int); ok && v > 0 {
		obj.SwapFileSizeMB = int32(v)
	}

	if v, ok := in["sysctls"].([]interface{}); ok && len(v) > 0 {
		obj.Sysctls = expandAKSV3NodePoolLinuxOsConfigSysctls(v)
	}

	if v, ok := in["transparent_huge_page_defrag"].(string); ok && len(v) > 0 {
		obj.TransparentHugePageDefrag = v
	}

	if v, ok := in["transparent_huge_page_enabled"].(string); ok && len(v) > 0 {
		obj.TransparentHugePageEnabled = v
	}
	return obj
}

func expandAKSV3NodePoolLinuxOsConfigSysctls(p []interface{}) *infrapb.Sysctls {
	obj := &infrapb.Sysctls{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["fs_aio_max_nr"].(int); ok && v > 0 {
		obj.FsAioMaxNr = int32(v)
	}

	if v, ok := in["fs_file_max"].(int); ok && v > 0 {
		obj.FsFileMax = int32(v)
	}

	if v, ok := in["fs_inotify_max_user_watches"].(int); ok && v > 0 {
		obj.FsInotifyMaxUserWatches = int32(v)
	}

	if v, ok := in["fs_nr_open"].(int); ok && v > 0 {
		obj.FsNrOpen = int32(v)
	}

	if v, ok := in["kernel_threads_max"].(int); ok && v > 0 {
		obj.KernelThreadsMax = int32(v)
	}

	if v, ok := in["net_core_netdev_max_backlog"].(int); ok && v > 0 {
		obj.NetCoreNetdevMaxBacklog = int32(v)
	}

	if v, ok := in["net_core_optmem_max"].(int); ok && v > 0 {
		obj.NetCoreOptmemMax = int32(v)
	}

	if v, ok := in["net_core_rmem_default"].(int); ok && v > 0 {
		obj.NetCoreRmemDefault = int32(v)
	}

	if v, ok := in["net_core_rmem_max"].(int); ok && v > 0 {
		obj.NetCoreRmemMax = int32(v)
	}

	if v, ok := in["net_core_somaxconn"].(int); ok && v > 0 {
		obj.NetCoreSomaxconn = int32(v)
	}

	if v, ok := in["net_core_wmem_default"].(int); ok && v > 0 {
		obj.NetCoreWmemDefault = int32(v)
	}

	if v, ok := in["net_core_wmem_max"].(int); ok && v > 0 {
		obj.NetCoreWmemMax = int32(v)
	}

	if v, ok := in["net_ipv4_ip_local_port_range"].(string); ok && len(v) > 0 {
		obj.NetIpv4IpLocalPortRange = v
	}

	if v, ok := in["net_ipv4_neigh_default_gc_thresh1"].(int); ok && v > 0 {
		obj.NetIpv4NeighDefaultGcThresh1 = int32(v)
	}

	if v, ok := in["net_ipv4_neigh_default_gc_thresh2"].(int); ok && v > 0 {
		obj.NetIpv4NeighDefaultGcThresh2 = int32(v)
	}

	if v, ok := in["net_ipv4_neigh_default_gc_thresh3"].(int); ok && v > 0 {
		obj.NetIpv4NeighDefaultGcThresh3 = int32(v)
	}

	if v, ok := in["net_ipv4_tcp_fin_timeout"].(int); ok && v > 0 {
		obj.NetIpv4TcpFinTimeout = int32(v)
	}

	if v, ok := in["net_ipv4_tcpkeepalive_intvl"].(int); ok && v > 0 {
		obj.NetIpv4TcpkeepaliveIntvl = int32(v)
	}

	if v, ok := in["net_ipv4_tcp_keepalive_probes"].(int); ok && v > 0 {
		obj.NetIpv4TcpKeepaliveProbes = int32(v)
	}

	if v, ok := in["net_ipv4_tcp_keepalive_time"].(int); ok && v > 0 {
		obj.NetIpv4TcpKeepaliveTime = int32(v)
	}

	if v, ok := in["net_ipv4_tcp_max_syn_backlog"].(int); ok && v > 0 {
		obj.NetIpv4TcpMaxSynBacklog = int32(v)
	}

	if v, ok := in["net_ipv4_tcp_max_tw_buckets"].(int); ok && v > 0 {
		obj.NetIpv4TcpMaxTwBuckets = int32(v)
	}

	if v, ok := in["net_ipv4_tcp_tw_reuse"].(bool); ok {
		obj.NetIpv4TcpTwReuse = v
	}

	if v, ok := in["net_netfilter_nf_conntrack_buckets"].(int); ok && v > 0 {
		obj.NetNetfilterNfConntrackBuckets = int32(v)
	}

	if v, ok := in["net_netfilter_nf_conntrack_max"].(int); ok && v > 0 {
		obj.NetNetfilterNfConntrackMax = int32(v)
	}

	if v, ok := in["vm_max_map_count"].(int); ok && v > 0 {
		obj.VmMaxMapCount = int32(v)
	}

	if v, ok := in["vm_swappiness"].(int); ok && v > 0 {
		obj.VmSwappiness = int32(v)
	}

	if v, ok := in["vm_vfs_cache_pressure"].(int); ok && v > 0 {
		obj.VmVfsCachePressure = int32(v)
	}

	return obj
}

func expandAKSV3NodePoolUpgradeSettings(p []interface{}) *infrapb.Upgradesettings {
	obj := &infrapb.Upgradesettings{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["max_surge"].(string); ok && len(v) > 0 {
		obj.MaxSurge = v
	}
	return obj
}

// Sort AKS Nodepool
type ByNodepoolNameV3 []infrapb.Nodepool

func (np ByNodepoolNameV3) Len() int      { return len(np) }
func (np ByNodepoolNameV3) Swap(i, j int) { np[i], np[j] = np[j], np[i] }
func (np ByNodepoolNameV3) Less(i, j int) bool {
	ret := strings.Compare(np[i].Name, np[j].Name)
	if ret < 0 {
		return true
	} else {
		return false
	}
}

func flattenAKSClusterV3(d *schema.ResourceData, in *infrapb.Cluster) error {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}

	if len(in.ApiVersion) > 0 {
		obj["api_version"] = in.ApiVersion
	}
	if len(in.Kind) > 0 {
		obj["kind"] = in.Kind
	}
	var err error

	var ret1 []interface{}
	if in.Metadata != nil {
		v, ok := obj["metadata"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		ret1 = flattenMetadataV3(in.Metadata, v)
	}

	err = d.Set("metadata", ret1)
	if err != nil {
		return err
	}

	var ret2 []interface{}
	if in.Spec != nil {
		v, ok := obj["spec"].([]interface{})
		if !ok {
			v = []interface{}{}
		}

		ret2 = flattenClusterV3Spec(in.Spec, v)
	}

	err = d.Set("spec", ret2)
	if err != nil {
		return err
	}

	return nil

}

func flattenMetadataV3(in *commonpb.Metadata, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}

	if len(in.Name) > 0 {
		obj["name"] = in.Name
	}

	if len(in.Project) > 0 {
		obj["project"] = in.Project
	}

	if in.Labels != nil && len(in.Labels) > 0 {
		obj["labels"] = toMapInterface(in.Labels)
	}

	return []interface{}{obj}
}

func flattenClusterV3Spec(in *infrapb.ClusterSpec, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.Type) > 0 {
		obj["type"] = in.Type
	}

	if in.BlueprintConfig != nil {
		obj["blueprint_config"] = flattenClusterV3Blueprint(in.BlueprintConfig)
	}

	if len(in.CloudCredentials) > 0 {
		obj["cloud_credentials"] = in.CloudCredentials
	}

	if in.GetAks() != nil {
		v, ok := obj["config"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["config"] = flattenAKSClusterV3Config(in.GetAks(), v)
	}

	if in.Sharing != nil {
		obj["sharing"] = flattenSharingSpecV3(in.Sharing)
	}

	return []interface{}{obj}
}

func flattenAKSClusterV3Config(in *infrapb.AksV3ConfigObject, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.ApiVersion) > 0 {
		obj["api_version"] = in.ApiVersion
	}

	if len(in.Kind) > 0 {
		obj["kind"] = in.Kind
	}

	if in.Metadata != nil {
		v, ok := obj["metadata"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["metadata"] = flattenAKSV3ClusterConfigMetadata(in.Metadata, v)
	}

	if in.Spec != nil {
		v, ok := obj["spec"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["spec"] = flattenAKSV3ClusterConfigSpec(in.Spec, v)
	}

	return []interface{}{obj}
}

func flattenAKSV3ClusterConfigMetadata(in *infrapb.AksV3ConfigMetadata, p []interface{}) []interface{} {
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
	return []interface{}{obj}
}

func flattenAKSV3ClusterConfigSpec(in *infrapb.AksV3Spec, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.SubscriptionID) > 0 {
		obj["subscription_id"] = in.SubscriptionID
	}

	if len(in.ResourceGroupName) > 0 {
		obj["resource_group_name"] = in.ResourceGroupName
	}

	if in.ManagedCluster != nil {
		v, ok := obj["managed_cluster"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["managed_cluster"] = flattenAKSV3ManagedCluster(in.ManagedCluster, v)
	}

	// @@@@@@@
	if in.NodePools != nil && len(in.NodePools) > 0 {
		v, ok := obj["node_pools"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["node_pools"] = flattenAKSV3NodePool(in.NodePools, v)
	}

	return []interface{}{obj}

}

func flattenAKSV3ManagedCluster(in *infrapb.Managedcluster, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.ApiVersion) > 0 {
		obj["api_version"] = in.ApiVersion
	}

	if in.ExtendedLocation != nil {
		v, ok := obj["extended_location"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["extended_location"] = flattenAKSManagedClusterV3ExtendedLocation(in.ExtendedLocation, v)
	}

	if in.Identity != nil {
		v, ok := obj["identity"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["identity"] = flattenAKSV3ManagedClusterIdentity(in.Identity, v)
	}

	if len(in.Location) > 0 {
		obj["location"] = in.Location
	}

	if in.Properties != nil {
		v, ok := obj["properties"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		log.Printf("About to enter flattenAKSManagedClusterProperties")
		obj["properties"] = flattenAKSV3ManagedClusterProperties(in.Properties, v)
	}

	if in.Sku != nil {
		v, ok := obj["sku"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["sku"] = flattenAKSV3ManagedClusterSKU(in.Sku, v)
	}

	if in.Tags != nil && len(in.Tags) > 0 {
		obj["tags"] = toMapInterface(in.Tags)
	}

	if len(in.Type) > 0 {
		obj["type"] = in.Type
	}

	if in.AdditionalMetadata != nil {
		v, ok := obj["additional_metadata"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["additional_metadata"] = flattenAKSV3ManagedClusterAdditionalMetadata(in.AdditionalMetadata, v)
	}

	return []interface{}{obj}

}

func flattenAKSManagedClusterV3ExtendedLocation(in *infrapb.Extendedlocation, p []interface{}) []interface{} {
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

	if len(in.Type) > 0 {
		obj["type"] = in.Type
	}

	return []interface{}{obj}

}

func flattenAKSV3ManagedClusterIdentity(in *infrapb.Identity, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.Type) > 0 {
		obj["type"] = in.Type
	}

	if in.UserAssignedIdentities != nil && len(in.UserAssignedIdentities) > 0 {
		identity := map[string]string{}
		for k := range in.UserAssignedIdentities {
			identity[k] = ""
		}
		obj["user_assigned_identities"] = identity
	}

	return []interface{}{obj}

}

func flattenAKSV3ManagedClusterProperties(in *infrapb.ManagedClusterProperties, p []interface{}) []interface{} {
	log.Printf("Entered flattenAKSV3ManagedClusterProperties")
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if in.AadProfile != nil {
		v, ok := obj["aad_profile"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["aad_profile"] = flattenAKSV3ManagedClusterAzureADProfile(in.AadProfile, v)
	}

	// TODO: REVIEW
	if in.AddonProfiles != nil {
		v, ok := obj["addon_profiles"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["addon_profiles"] = flattenAKSV3ManagedClusterAddonProfile(in.AddonProfiles, v)
	}

	if in.ApiServerAccessProfile != nil {
		v, ok := obj["api_server_access_profile"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["api_server_access_profile"] = flattenAKSV3ManagedClusterAPIServerAccessProfile(in.ApiServerAccessProfile, v)
	}

	if in.AutoScalerProfile != nil {
		v, ok := obj["auto_scaler_profile"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["auto_scaler_profile"] = flattenAKSV3ManagedClusterAutoScalerProfile(in.AutoScalerProfile, v)
	}

	if in.AutoUpgradeProfile != nil {
		v, ok := obj["auto_upgrade_profile"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["auto_upgrade_profile"] = flattenAKSV3ManagedClusterAutoUpgradeProfile(in.AutoUpgradeProfile, v)
	}

	obj["disable_local_accounts"] = in.DisableLocalAccounts

	if len(in.DiskEncryptionSetID) > 0 {
		obj["disk_encryption_set_id"] = in.DiskEncryptionSetID
	}

	if len(in.DnsPrefix) > 0 {
		obj["dns_prefix"] = in.DnsPrefix
	}

	obj["enable_pod_security_policy"] = in.EnablePodSecurityPolicy

	obj["enable_rbac"] = in.EnableRBAC

	if len(in.FqdnSubdomain) > 0 {
		obj["fqdn_subdomain"] = in.FqdnSubdomain
	}

	if in.HttpProxyConfig != nil {
		v, ok := obj["http_proxy_config"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["http_proxy_config"] = flattenAKSV3ManagedClusterHTTPProxyConfig(in.HttpProxyConfig, v)
	}

	if in.IdentityProfile != nil && len(in.IdentityProfile) > 0 {
		obj["identity_profile"] = toMapInterface(in.IdentityProfile)
	}

	if len(in.KubernetesVersion) > 0 {
		obj["kubernetes_version"] = in.KubernetesVersion
	}

	if in.LinuxProfile != nil {
		v, ok := obj["linux_profile"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["linux_profile"] = flattenAKSV3ManagedClusterLinuxProfile(in.LinuxProfile, v)
	}

	if in.NetworkProfile != nil {
		v, ok := obj["network_profile"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["network_profile"] = flattenAKSV3MCPropertiesNetworkProfile(in.NetworkProfile, v)
	}

	if len(in.NodeResourceGroup) > 0 {
		obj["node_resource_group"] = in.NodeResourceGroup
	}

	if in.PodIdentityProfile != nil {
		v, ok := obj["pod_identity_profile"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["pod_identity_profile"] = flattenAKSV3ManagedClusterPodIdentityProfile(in.PodIdentityProfile, v)
	}

	if in.PrivateLinkResources != nil {
		v, ok := obj["private_link_resources"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["private_link_resources"] = flattenAKSV3ManagedClusterPrivateLinkResources(in.PrivateLinkResources, v)
	}

	if in.ServicePrincipalProfile != nil {
		v, ok := obj["service_principal_profile"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["service_principal_profile"] = flattenAKSV3ManagedClusterServicePrincipalProfile(in.ServicePrincipalProfile, v)
	}

	if in.WindowsProfile != nil {
		v, ok := obj["windows_profile"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["windows_profile"] = flattenAKSV3ManagedClusterWindowsProfile(in.WindowsProfile, v)
	}

	return []interface{}{obj}

}

func flattenAKSV3ManagedClusterAddonProfile(in *infrapb.ManagedClusterAddonProfile, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if in.HttpApplicationRouting != nil {
		v, ok := obj["http_application_routing"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["http_application_routing"] = flattenAKSV3ManagedClusterAddonOnGenericProfile(in.HttpApplicationRouting, v)
	}

	if in.AzurePolicy != nil {
		v, ok := obj["azure_policy"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["azure_policy"] = flattenAKSV3ManagedClusterAddonOnGenericProfile(in.AzurePolicy, v)
	}

	if in.AzureKeyvaultSecretsProvider != nil {
		v, ok := obj["azure_keyvault_secrets_provider"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["azure_keyvault_secrets_provider"] = flattenAKSV3ManagedClusterAzureKeyvaultSecretsProviderProfile(in.AzureKeyvaultSecretsProvider, v)
	}

	if in.OmsAgent != nil {
		v, ok := obj["oms_agent"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["oms_agent"] = flattenAKSV3ManagedClusterOmsAgentProfile(in.OmsAgent, v)
	}

	if in.IngressApplicationGateway != nil {
		v, ok := obj["ingress_application_gateway"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["ingress_application_gateway"] = flattenAKSV3ManagedClusterIngressApplicationGatewayProfile(in.IngressApplicationGateway, v)
	}

	return []interface{}{obj}
}

func flattenAKSV3ManagedClusterAddonOnGenericProfile(in *infrapb.AddonProfileGeneric, p []interface{}) []interface{} {
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

func flattenAKSV3ManagedClusterAzureKeyvaultSecretsProviderProfile(in *infrapb.AddonProfileAzureKeyvaultSecretsProvider, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	obj["enabled"] = in.Enabled

	if in.Config != nil {
		v, ok := obj["config"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["config"] = flattenAKSV3ManagedClusterAzureKeyvaultSecretsProviderProfileConfig(in.Config, v)
	}
	return []interface{}{obj}
}

func flattenAKSV3ManagedClusterAzureKeyvaultSecretsProviderProfileConfig(in *infrapb.AzureKeyvaultSecretsProviderProfileConfig, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	obj["enable_secret_rotation"] = in.EnableSecretRotation
	obj["rotation_poll_interval"] = in.RotationPollInterval

	return []interface{}{obj}
}

func flattenAKSV3ManagedClusterOmsAgentProfile(in *infrapb.AddonProfileOmsAgent, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	obj["enabled"] = in.Enabled
	if in.Config != nil {
		v, ok := obj["config"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["config"] = flattenAKSV3ManagedClusterOmsAgentProfileConfig(in.Config, v)
	}

	return []interface{}{obj}
}

func flattenAKSV3ManagedClusterOmsAgentProfileConfig(in *infrapb.AddonProfileOmsAgentConfig, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	obj["log_analytics_workspace_resource_id"] = in.LogAnalyticsWorkspaceResourceID
	return []interface{}{obj}
}

func flattenAKSV3ManagedClusterIngressApplicationGatewayProfile(in *infrapb.AddonProfileIngressApplicationGateway, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	obj["enabled"] = in.Enabled
	if in.Config != nil {
		v, ok := obj["config"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["config"] = flattenAKSV3ManagedClusterIngressApplicationGatewayProfileConfig(in.Config, v)
	}
	return []interface{}{obj}
}

func flattenAKSV3ManagedClusterIngressApplicationGatewayProfileConfig(in *infrapb.AddonProfileIngressApplicationGatewayConfig, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	obj["application_gateway_id"] = in.ApplicationGatewayID
	obj["application_gateway_name"] = in.ApplicationGatewayName
	obj["subnet_cidr"] = in.SubnetCIDR
	obj["subnet_id"] = in.SubnetID
	obj["watch_namespace"] = in.WatchNamespace

	return []interface{}{obj}
}

func flattenAKSV3ManagedClusterAzureADProfile(in *infrapb.Aadprofile, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if in.AdminGroupObjectIDs != nil && len(in.AdminGroupObjectIDs) > 0 {
		obj["admin_group_object_ids"] = toArrayInterface(in.AdminGroupObjectIDs)
	}

	if len(in.ClientAppID) > 0 {
		obj["client_app_id"] = in.ClientAppID
	}

	obj["enable_azure_rbac"] = in.EnableAzureRBAC

	obj["managed"] = in.Managed

	if len(in.ServerAppID) > 0 {
		obj["server_app_id"] = in.ServerAppID
	}

	if len(in.ServerAppSecret) > 0 {
		obj["server_app_id_secret"] = in.ServerAppSecret
	}

	if len(in.TenantID) > 0 {
		obj["tenant_id"] = in.TenantID
	}

	return []interface{}{obj}

}

func flattenAKSV3ManagedClusterAPIServerAccessProfile(in *infrapb.Apiserveraccessprofile, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if in.AuthorizedIPRanges != nil && len(in.AuthorizedIPRanges) > 0 {
		obj["authorized_ipr_ranges"] = toArrayInterface(in.AuthorizedIPRanges)
	}

	obj["enable_private_cluster"] = in.EnablePrivateCluster

	obj["enable_private_cluster_public_fqdn"] = in.EnablePrivateClusterPublicFQDN

	if len(in.PrivateDNSZone) > 0 {
		obj["private_dns_zone"] = in.PrivateDNSZone
	}

	return []interface{}{obj}
}

func flattenAKSV3ManagedClusterAutoScalerProfile(in *infrapb.Autoscalerprofile, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.BalanceSimilarNodeGroups) > 0 {
		obj["balance_similar_node_groups"] = in.BalanceSimilarNodeGroups
	}

	if len(in.Expander) > 0 {
		obj["expander"] = in.Expander
	}

	if len(in.MaxEmptyBulkDelete) > 0 {
		obj["max_empty_bulk_delete"] = in.MaxEmptyBulkDelete
	}

	if len(in.MaxGracefulTerminationSec) > 0 {
		obj["max_graceful_termination_sec"] = in.MaxGracefulTerminationSec
	}

	if len(in.MaxNodeProvisionTime) > 0 {
		obj["max_node_provision_time"] = in.MaxNodeProvisionTime
	}

	if len(in.MaxTotalUnreadyPercentage) > 0 {
		obj["max_total_unready_percentage"] = in.MaxTotalUnreadyPercentage
	}

	if len(in.NewPodScaleUpDelay) > 0 {
		obj["new_pod_scale_up_delay"] = in.NewPodScaleUpDelay
	}

	if len(in.OkTotalUnreadyCount) > 0 {
		obj["ok_total_unready_count"] = in.OkTotalUnreadyCount
	}
	/*
		if in.OkTotalUnreadyCount != nil {
			obj["ok_total_unready_count"] = *in.OkTotalUnreadyCount
		}
	*/
	if len(in.ScaleDownDelayAfterAdd) > 0 {
		obj["scale_down_delay_after_add"] = in.ScaleDownDelayAfterAdd
	}

	if len(in.ScaleDownDelayAfterDelete) > 0 {
		obj["scale_down_delay_after_delete"] = in.ScaleDownDelayAfterDelete
	}

	if len(in.ScaleDownDelayAfterFailure) > 0 {
		obj["scale_down_delay_after_failure"] = in.ScaleDownDelayAfterFailure
	}

	if len(in.ScaleDownUnneededTime) > 0 {
		obj["scale_down_unneeded_time"] = in.ScaleDownUnneededTime
	}

	if len(in.ScaleDownUnreadyTime) > 0 {
		obj["scale_down_unready_time"] = in.ScaleDownUnreadyTime
	}

	if len(in.ScaleDownUtilizationThreshold) > 0 {
		obj["scale_down_utilization_threshold"] = in.ScaleDownUtilizationThreshold
	}

	if len(in.ScanInterval) > 0 {
		obj["scan_interval"] = in.ScanInterval
	}

	if len(in.SkipNodesWithLocalStorage) > 0 {
		obj["skip_nodes_with_local_storage"] = in.SkipNodesWithLocalStorage
	}

	if len(in.SkipNodesWithSystemPods) > 0 {
		obj["skip_nodes_with_system_pods"] = in.SkipNodesWithSystemPods
	}

	return []interface{}{obj}

}

func flattenAKSV3ManagedClusterAutoUpgradeProfile(in *infrapb.Autoupgradeprofile, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.UpgradeChannel) > 0 {
		obj["upgrade_channel"] = in.UpgradeChannel
	}

	return []interface{}{obj}

}

func flattenAKSV3ManagedClusterHTTPProxyConfig(in *infrapb.Httpproxyconfig, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.HttpProxy) > 0 {
		obj["http_proxy"] = in.HttpProxy
	}

	if len(in.HttpProxy) > 0 {
		obj["https_proxy"] = in.HttpProxy
	}

	if in.NoProxy != nil && len(in.NoProxy) > 0 {
		obj["no_proxy"] = toArrayInterface(in.NoProxy)
	}

	if len(in.TrustedCa) > 0 {
		obj["trusted_ca"] = in.TrustedCa
	}

	return []interface{}{obj}

}

func flattenAKSV3ManagedClusterLinuxProfile(in *infrapb.Linuxprofile, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.AdminUsername) > 0 {
		obj["admin_username"] = in.AdminUsername
	}

	if in.Ssh != nil {
		v, ok := obj["ssh"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["ssh"] = flattenAKSV3ManagedClusterSSHConfig(in.Ssh, v)
	}

	return []interface{}{obj}

}

func flattenAKSV3ManagedClusterSSHConfig(in *infrapb.Ssh, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if in.PublicKeys != nil && len(in.PublicKeys) > 0 {
		v, ok := obj["ssh"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["public_keys"] = flattenAKSV3ManagedClusterSSHKeyData(in.PublicKeys, v)
	}

	return []interface{}{obj}

}

func flattenAKSV3ManagedClusterSSHKeyData(in []*infrapb.Publickeys, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	out := make([]interface{}, len(in))
	for i, in := range in {

		obj := map[string]interface{}{}
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}

		if len(in.KeyData) > 0 {
			obj["key_data"] = in.KeyData
		}

		out[i] = &obj
	}

	return out

}

func flattenAKSV3MCPropertiesNetworkProfile(in *infrapb.Networkprofile, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.DnsServiceIP) > 0 {
		obj["dns_service_ip"] = in.DnsServiceIP
	}

	if len(in.DockerBridgeCidr) > 0 {
		obj["docker_bridge_cidr"] = in.DockerBridgeCidr
	}

	if in.LoadBalancerProfile != nil {
		v, ok := obj["load_balancer_profile"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["load_balancer_profile"] = flattenAKSV3ManagedClusterNPLoadBalancerProfile(in.LoadBalancerProfile, v)
	}

	if len(in.LoadBalancerSku) > 0 {
		obj["load_balancer_sku"] = in.LoadBalancerSku
	}

	if len(in.NetworkMode) > 0 {
		obj["network_mode"] = in.NetworkMode
	}

	if len(in.NetworkPlugin) > 0 {
		obj["network_plugin"] = in.NetworkPlugin
	}

	if len(in.NetworkPolicy) > 0 {
		obj["network_policy"] = in.NetworkPolicy
	}

	if len(in.OutboundType) > 0 {
		obj["outbound_type"] = in.OutboundType
	}

	if len(in.PodCidr) > 0 {
		obj["pod_cidr"] = in.PodCidr
	}

	if len(in.ServiceCidr) > 0 {
		obj["service_cidr"] = in.ServiceCidr
	}

	return []interface{}{obj}

}

func flattenAKSV3ManagedClusterNPLoadBalancerProfile(in *infrapb.Loadbalancerprofile, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	// TODO: REVIEW THIS
	if in.AllocatedOutboundPorts != 0 {
		obj["allocated_outbound_ports"] = in.AllocatedOutboundPorts
	}

	if in.EffectiveOutboundIPs != nil && len(in.EffectiveOutboundIPs) > 0 {
		v, ok := obj["effective_outbound_ips"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["effective_outbound_ips"] = flattenAKSV3ManagedClusterNPEffectiveOutboundIPs(in.EffectiveOutboundIPs, v)
	}

	// TODO: REVIEW THIS
	if in.IdleTimeoutInMinutes != 0 {
		obj["idle_timeout_in_minutes"] = in.IdleTimeoutInMinutes
	}

	// TODO: FIX
	if in.ManagedOutboundIPs != nil {
		v, ok := obj["managed_outbound_ips"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["managed_outbound_ips"] = flattenAKSV3ManagedClusterNPManagedOutboundIPs(in.ManagedOutboundIPs, v)
	}

	if in.OutboundIPPrefixes != nil {
		v, ok := obj["outbound_ip_prefixes"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["outbound_ip_prefixes"] = flattenAKSV3ManagedClusterNPOutboundIPPrefixes(in.OutboundIPPrefixes, v)
	}

	if in.OutboundIPs != nil {
		v, ok := obj["outbound_ips"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["outbound_ips"] = flattenAKSV3ManagedClusterNPOutboundIPs(in.OutboundIPs, v)
	}

	return []interface{}{obj}

}

func flattenAKSV3ManagedClusterNPEffectiveOutboundIPs(in []*infrapb.Effectiveoutboundips, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	out := make([]interface{}, len(in))
	for i, in := range in {

		obj := map[string]interface{}{}
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}

		if len(in.Id) > 0 {
			obj["id"] = in.Id
		}

		out[i] = &obj
	}

	return out
}

func flattenAKSV3ManagedClusterNPManagedOutboundIPs(in *infrapb.Managedoutboundips, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if in.Count > 0 {
		obj["count"] = in.Count
	}

	return []interface{}{obj}

}

func flattenAKSV3ManagedClusterNPOutboundIPPrefixes(in *infrapb.Outboundipprefixes, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if in.PublicIPPrefixes != nil && len(in.PublicIPPrefixes) > 0 {
		v, ok := obj["public_ip_prefixes"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["public_ip_prefixes"] = flattenAKSV3ManagedClusterNPOutboundIPsPublicIPPrefixes(in.PublicIPPrefixes, v)
	}

	return []interface{}{obj}

}

func flattenAKSV3ManagedClusterNPOutboundIPsPublicIPPrefixes(in []*infrapb.Publicipprefixes, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	out := make([]interface{}, len(in))
	for i, in := range in {

		obj := map[string]interface{}{}
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}

		if len(in.Id) > 0 {
			obj["id"] = in.Id
		}

		out[i] = &obj
	}

	return out
}

func flattenAKSV3ManagedClusterNPOutboundIPs(in *infrapb.Outboundips, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if in.PublicIPs != nil && len(in.PublicIPs) > 0 {
		v, ok := obj["public_ips"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["public_ips"] = flattenAKSV3ManagedClusterNPOutboundIPsPublicIPs(in.PublicIPs, v)
	}

	return []interface{}{obj}

}

func flattenAKSV3ManagedClusterNPOutboundIPsPublicIPs(in []*infrapb.Publicips, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	out := make([]interface{}, len(in))
	for i, in := range in {

		obj := map[string]interface{}{}
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}

		if len(in.Id) > 0 {
			obj["id"] = in.Id
		}

		out[i] = &obj
	}

	return out

}

func flattenAKSV3ManagedClusterPodIdentityProfile(in *infrapb.Podidentityprofile, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	obj["allow_network_plugin_kubenet"] = in.AllowNetworkPluginKubenet

	obj["enabled"] = in.Enabled

	if in.UserAssignedIdentities != nil {
		v, ok := obj["user_assigned_identities"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["user_assigned_identities"] = flattenAKSV3ManagedClusterPIPUserAssignedIdentities(in.UserAssignedIdentities, v)
	}

	if in.UserAssignedIdentityExceptions != nil {
		v, ok := obj["user_assigned_identity_exceptions"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["user_assigned_identity_exceptions"] = flattenAKSV3ManagedClusterPIPUserAssignedIdentityExceptions(in.UserAssignedIdentityExceptions, v)
	}

	return []interface{}{obj}
}

func flattenAKSV3ManagedClusterPIPUserAssignedIdentities(inp []*infrapb.PodUserassignedidentities, p []interface{}) []interface{} {
	if inp == nil {
		return nil
	}
	out := make([]interface{}, len(inp))
	for i, in := range inp {
		obj := map[string]interface{}{}
		if len(p) != 0 && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}

		if len(in.BindingSelector) > 0 {
			obj["binding_selector"] = in.BindingSelector
		}

		if in.Identity != nil {
			v, ok := obj["identity"].([]interface{})
			if !ok {
				v = []interface{}{}
			}
			obj["identity"] = flattenAKSV3ManagedClusterUAIIdentity(in.Identity, v)
		}

		if len(in.Name) > 0 {
			obj["name"] = in.Name
		}

		if len(in.Namespace) > 0 {
			obj["namespace"] = in.Namespace
		}
		out[i] = &obj
	}

	return out
}

func flattenAKSV3ManagedClusterUAIIdentity(in *infrapb.PodIdentity, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.ClientId) > 0 {
		obj["client_id"] = in.ClientId
	}

	if len(in.ObjectId) > 0 {
		obj["object_id"] = in.ObjectId
	}

	if len(in.ResourceId) > 0 {
		obj["resource_id"] = in.ResourceId
	}

	return []interface{}{obj}

}

func flattenAKSV3ManagedClusterPIPUserAssignedIdentityExceptions(inp []*infrapb.Userassignedidentityexceptions, p []interface{}) []interface{} {
	if inp == nil {
		return nil
	}
	out := make([]interface{}, len(inp))
	for i, in := range inp {
		obj := map[string]interface{}{}
		if len(p) != 0 && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}

		if len(in.Name) > 0 {
			obj["name"] = in.Name
		}

		if len(in.Namespace) > 0 {
			obj["namespace"] = in.Namespace
		}

		if in.PodLabels != nil && len(in.PodLabels) > 0 {
			obj["pod_labels"] = toMapInterface(in.PodLabels)
		}
		out[i] = &obj
	}
	return out
}

func flattenAKSV3ManagedClusterPrivateLinkResources(in []*infrapb.Privatelinkresources, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	out := make([]interface{}, len(in))

	for i, in := range in {
		obj := map[string]interface{}{}

		if len(in.GroupId) > 0 {
			obj["group_id"] = in.GroupId
		}

		if len(in.Id) > 0 {
			obj["id"] = in.Id
		}

		if len(in.Name) > 0 {
			obj["name"] = in.Name
		}

		if in.RequiredMembers != nil && len(in.RequiredMembers) > 0 {
			obj["required_members"] = toArrayInterface(in.RequiredMembers)
		}

		if len(in.Type) > 0 {
			obj["type"] = in.Type
		}

		out[i] = &obj
	}

	return out

}

func flattenAKSV3ManagedClusterServicePrincipalProfile(in *infrapb.Serviceprincipalprofile, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.ClientId) > 0 {
		obj["client_id"] = in.ClientId
	}

	if len(in.Secret) > 0 {
		obj["secret"] = in.Secret
	}

	return []interface{}{obj}

}

func flattenAKSV3ManagedClusterWindowsProfile(in *infrapb.Windowsprofile, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.AdminUsername) > 0 {
		obj["admin_username"] = in.AdminUsername
	}

	if len(in.LicenseType) > 0 {
		obj["license_type"] = in.LicenseType
	}

	obj["enable_csi_proxy"] = in.EnableCSIProxy

	return []interface{}{obj}

}

func flattenAKSV3ManagedClusterSKU(in *infrapb.Sku, p []interface{}) []interface{} {
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

	if len(in.Tier) > 0 {
		obj["tier"] = in.Tier
	}

	return []interface{}{obj}

}

func flattenAKSV3ManagedClusterAdditionalMetadata(in *infrapb.Additionalmetadata, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if in.AcrProfile != nil {
		v, ok := obj["acr_profile"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["acr_profile"] = flattenAKSV3ManagedClusterAdditionalMetadataACRProfile(in.AcrProfile, v)
	}

	if len(in.OmsWorkspaceLocation) > 0 {
		obj["oms_workspace_location"] = in.OmsWorkspaceLocation
	}

	return []interface{}{obj}

}

func flattenAKSV3ManagedClusterAdditionalMetadataACRProfile(in *infrapb.AcrProfile, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.ResourceGroupName) > 0 {
		obj["resource_group_name"] = in.ResourceGroupName
	}

	if len(in.AcrName) > 0 {
		obj["acr_name"] = in.AcrName
	}

	return []interface{}{obj}

}

func flattenAKSV3NodePool(in []*infrapb.Nodepool, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}

	// TODO: TEST
	// sort the incoming nodepools
	// sortedIn := make([]infrapb.Nodepool, len(in))
	// for i := range in {
	// 	sortedIn[i] = *in[i]
	// }
	// sort.Sort(ByNodepoolNameV3(sortedIn))

	out := make([]interface{}, len(in))
	for i, in := range in {

		obj := map[string]interface{}{}
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}

		if len(in.ApiVersion) > 0 {
			obj["api_version"] = in.ApiVersion
		}

		if len(in.Name) > 0 {
			obj["name"] = in.Name
		}

		if in.Properties != nil {
			v, ok := obj["properties"].([]interface{})
			if !ok {
				v = []interface{}{}
			}
			obj["properties"] = flattenAKSV3NodePoolProperties(in.Properties, v)
		}

		if len(in.Type) > 0 {
			obj["type"] = in.Type
		}

		if len(in.Location) > 0 {
			obj["location"] = in.Location
		}

		out[i] = obj
	}
	return out
}

func flattenAKSV3NodePoolProperties(in *infrapb.NodePoolProperties, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if in.AvailabilityZones != nil && len(in.AvailabilityZones) > 0 {
		obj["availability_zones"] = toArrayInterface(in.AvailabilityZones)
	}

	obj["count"] = in.Count

	if in.EnableAutoScaling {
		obj["enable_auto_scaling"] = in.EnableAutoScaling
	}

	if in.EnableEncryptionAtHost {
		obj["enable_encryption_at_host"] = in.EnableEncryptionAtHost
	}

	if in.EnableFIPS {
		obj["enable_fips"] = in.EnableFIPS
	}

	if in.EnableNodePublicIP {
		obj["enable_node_public_ip"] = in.EnableNodePublicIP
	}

	if in.EnableUltraSSD {
		obj["enable_ultra_ssd"] = in.EnableUltraSSD
	}

	if len(in.GpuInstanceProfile) > 0 {
		obj["gpu_instance_profile"] = in.GpuInstanceProfile
	}

	if in.KubeletConfig != nil {
		v, ok := obj["kubelet_config"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["kubelet_config"] = flattenAKSV3NodePoolKubeletConfig(in.KubeletConfig, v)
	}

	if len(in.KubeletDiskType) > 0 {
		obj["kubelet_disk_type"] = in.KubeletDiskType
	}

	if in.LinuxOSConfig != nil {
		v, ok := obj["linux_os_config"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["linux_os_config"] = flattenAKSV3NodePoolLinuxOsConfig(in.LinuxOSConfig, v)
	}

	obj["max_count"] = in.MaxCount

	obj["max_pods"] = in.MaxPods

	obj["min_count"] = in.MinCount

	if len(in.Mode) > 0 {
		obj["mode"] = in.Mode
	}

	if in.NodeLabels != nil && len(in.NodeLabels) > 0 {
		obj["node_labels"] = toMapInterface(in.NodeLabels)
	}

	if len(in.NodePublicIPPrefixID) > 0 {
		obj["node_public_ip_prefix_id"] = in.NodePublicIPPrefixID
	}

	if in.NodeTaints != nil && len(in.NodeTaints) > 0 {
		obj["node_taints"] = toArrayInterface(in.NodeTaints)
	}

	if len(in.OrchestratorVersion) > 0 {
		obj["orchestrator_version"] = in.OrchestratorVersion
	}

	obj["os_disk_size_gb"] = in.OsDiskSizeGB

	if len(in.OsDiskType) > 0 {
		obj["os_disk_type"] = in.OsDiskType
	}

	if len(in.OsSKU) > 0 {
		obj["os_sku"] = in.OsSKU
	}

	if len(in.OsType) > 0 {
		obj["os_type"] = in.OsType
	}

	if len(in.PodSubnetID) > 0 {
		obj["pod_subnet_id"] = in.PodSubnetID
	}

	if len(in.ProximityPlacementGroupID) > 0 {
		obj["proximity_placement_group_id"] = in.ProximityPlacementGroupID
	}

	if len(in.ScaleSetEvictionPolicy) > 0 {
		obj["scale_set_eviction_policy"] = in.ScaleSetEvictionPolicy
	}

	if len(in.ScaleSetPriority) > 0 {
		obj["scale_set_priority"] = in.ScaleSetPriority
	}

	obj["spot_max_price"] = in.SpotMaxPrice

	if in.Tags != nil && len(in.Tags) > 0 {
		obj["tags"] = toMapInterface(in.Tags)
	}

	if len(in.Type) > 0 {
		obj["type"] = in.Type
	}

	if in.UpgradeSettings != nil {
		v, ok := obj["upgrade_settings"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["upgrade_settings"] = flattenAKSV3NodePoolUpgradeSettings(in.UpgradeSettings, v)
	}

	if len(in.VmSize) > 0 {
		obj["vm_size"] = in.VmSize
	}

	if len(in.VnetSubnetID) > 0 {
		obj["vnet_subnet_id"] = in.VnetSubnetID
	}

	return []interface{}{obj}

}

func flattenAKSV3NodePoolKubeletConfig(in *infrapb.Kubeletconfig, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if in.AllowedUnsafeSysctls != nil && len(in.AllowedUnsafeSysctls) > 0 {
		obj["allowed_unsafe_sysctls"] = toArrayInterface(in.AllowedUnsafeSysctls)
	}

	obj["container_log_max_files"] = in.ContainerLogMaxFiles

	obj["container_log_max_size_mb"] = in.ContainerLogMaxSizeMB

	if in.CpuCfsQuota {
		obj["cpu_cfs_quota"] = in.CpuCfsQuota
	}

	if len(in.CpuCfsQuotaPeriod) > 0 {
		obj["cpu_cfs_quota_period"] = in.CpuCfsQuotaPeriod
	}

	if len(in.CpuManagerPolicy) > 0 {
		obj["cpu_manager_policy"] = in.CpuManagerPolicy
	}

	if in.FailSwapOn {
		obj["fail_swap_on"] = in.FailSwapOn
	}

	obj["image_gc_high_threshold"] = in.ImageGcHighThreshold

	obj["image_gc_low_threshold"] = in.ImageGcLowThreshold

	obj["pod_max_pids"] = in.PodMaxPids

	if len(in.TopologyManagerPolicy) > 0 {
		obj["topology_manager_policy"] = in.TopologyManagerPolicy
	}

	return []interface{}{obj}
}

func flattenAKSV3NodePoolLinuxOsConfig(in *infrapb.Linuxosconfig, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if in.Sysctls != nil {
		v, ok := obj["sysctls"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["sysctls"] = flattenAKSV3NodePoolLinuxOsConfigSysctls(in.Sysctls, v)
	}

	if len(in.TransparentHugePageDefrag) > 0 {
		obj["transparent_huge_page_defrag"] = in.TransparentHugePageDefrag
	}

	if len(in.TransparentHugePageEnabled) > 0 {
		obj["transparent_huge_page_enabled"] = in.TransparentHugePageEnabled
	}

	return []interface{}{obj}

}

func flattenAKSV3NodePoolLinuxOsConfigSysctls(in *infrapb.Sysctls, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	obj["fs_aio_max_nr"] = in.FsAioMaxNr

	obj["fs_file_max"] = in.FsFileMax

	obj["fs_inotify_max_user_watches"] = in.FsInotifyMaxUserWatches

	obj["fs_nr_open"] = in.FsNrOpen

	obj["kernel_threads_max"] = in.KernelThreadsMax

	obj["net_core_netdev_max_backlog"] = in.NetCoreNetdevMaxBacklog

	obj["net_core_optmem_max"] = in.NetCoreOptmemMax

	obj["net_core_rmem_default"] = in.NetCoreRmemDefault

	obj["net_core_optmem_max"] = in.NetCoreRmemMax

	obj["net_core_somaxconn"] = in.NetCoreSomaxconn

	obj["net_core_wmem_default"] = in.NetCoreWmemDefault

	obj["net_core_wmem_max"] = in.NetCoreWmemMax

	if len(in.NetIpv4IpLocalPortRange) > 0 {
		obj["net_ipv4_ip_local_port_range"] = in.NetIpv4IpLocalPortRange
	}

	obj["net_ipv4_neigh_default_gc_thresh1"] = in.NetIpv4NeighDefaultGcThresh1

	obj["net_ipv4_neigh_default_gc_thresh2"] = in.NetIpv4NeighDefaultGcThresh2

	obj["net_ipv4_neigh_default_gc_thresh3"] = in.NetIpv4NeighDefaultGcThresh3

	obj["net_ipv4_tcp_fin_timeout"] = in.NetIpv4TcpFinTimeout

	obj["net_ipv4_tcpkeepalive_intvl"] = in.NetIpv4TcpkeepaliveIntvl

	obj["net_ipv4_tcp_keepalive_probes"] = in.NetIpv4TcpKeepaliveProbes

	obj["net_ipv4_tcp_keepalive_time"] = in.NetIpv4TcpKeepaliveTime

	obj["net_ipv4_tcp_max_syn_backlog"] = in.NetIpv4TcpMaxSynBacklog

	obj["net_ipv4_tcp_max_tw_buckets"] = in.NetIpv4TcpMaxTwBuckets

	if in.NetIpv4TcpTwReuse {
		obj["net_ipv4_tcp_tw_reuse"] = in.NetIpv4TcpTwReuse
	}

	obj["net_netfilter_nf_conntrack_buckets"] = in.NetNetfilterNfConntrackBuckets

	obj["net_netfilter_nf_conntrack_max"] = in.NetNetfilterNfConntrackMax

	obj["vm_max_map_count"] = in.VmMaxMapCount

	obj["vm_swappiness"] = in.VmSwappiness

	obj["vm_vfs_cache_pressure"] = in.VmVfsCachePressure

	return []interface{}{obj}

}

func flattenAKSV3NodePoolUpgradeSettings(in *infrapb.Upgradesettings, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.MaxSurge) > 0 {
		obj["max_surge"] = in.MaxSurge
	}

	return []interface{}{obj}

}

func flattenClusterV3Blueprint(in *infrapb.BlueprintConfig) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}

	if len(in.Name) > 0 {
		obj["name"] = in.Name
	}

	if len(in.Version) > 0 {
		obj["version"] = in.Version
	}

	return []interface{}{obj}
}
