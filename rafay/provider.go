package rafay

import (
	"context"
	"crypto/tls"
	"log"
	"net/http"
	"os/user"
	"path/filepath"
	"strings"

	rctlconfig "github.com/RafaySystems/rctl/pkg/config"
	rctlcontext "github.com/RafaySystems/rctl/pkg/context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const TF_USER_AGENT = "terraform"

func New(_ string) func() *schema.Provider {
	return func() *schema.Provider {
		p := &schema.Provider{
			Schema: map[string]*schema.Schema{
				"provider_config_file": &schema.Schema{
					Type:        schema.TypeString,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("RAFAY_PROVIDER_CONFIG", "~/.rafay/cli/config.json"),
				},
				"ignore_insecure_tls_error": &schema.Schema{
					Type:     schema.TypeBool,
					Optional: true,
				},
			},
			ResourcesMap: map[string]*schema.Resource{
				"rafay_project":                           resourceProject(),
				"rafay_cloud_credential":                  resourceCloudCredential(),
				"rafay_eks_cluster":                       resourceEKSCluster(),
				"rafay_eks_cluster_spec":                  resourceEKSClusterSpec(),
				"rafay_aks_cluster":                       resourceAKSCluster(),
				"rafay_aks_cluster_v3":                    resourceAKSClusterV3(),
				"rafay_aks_cluster_spec":                  resourceAKSClusterSpec(),
				"rafay_gke_cluster":                       resourceGKEClusterV3(),
				"rafay_addon":                             resourceAddon(),
				"rafay_blueprint":                         resourceBluePrint(),
				"rafay_import_cluster":                    resourceImportCluster(),
				"rafay_cluster_override":                  resourceClusterOverride(),
				"rafay_workload":                          resourceWorkload(),
				"rafay_namespace":                         resourceNamespace(),
				"rafay_repositories":                      resourceRepositories(),
				"rafay_agent":                             resourceAgent(),
				"rafay_pipeline":                          resourcePipeline(),
				"rafay_secret_group":                      resourceSecretGroup(),
				"rafay_workloadtemplate":                  resourceWorkloadTemplate(),
				"rafay_opa_policy":                        resourceOPAPolicy(),
				"rafay_opa_constraint":                    resourceOPAConstraint(),
				"rafay_opa_constraint_template":           resourceOPAConstraintTemplate(),
				"rafay_secretsealer":                      resourceSecretSealer(),
				"rafay_infra_provisioner":                 resourceInfraProvisioner(),
				"rafay_groupassociation":                  resourceGroupAssociation(),
				"rafay_group":                             resourceGroup(),
				"rafay_catalog":                           resourceCatalog(),
				"rafay_secret_provider":                   resourceSecretProvider(),
				"rafay_download_kubeconfig":               downloadKubeConfig(),
				"rafay_cluster_network_policy":            resourceClusterNetworkPolicy(),
				"rafay_cluster_network_policy_rule":       resourceClusterNetworkPolicyRule(),
				"rafay_namespace_network_policy":          resourceNamespaceNetworkPolicy(),
				"rafay_namespace_network_policy_rule":     resourceNamespaceNetworkPolicyRule(),
				"rafay_network_policy_profile":            resourceNetworkPolicyProfile(),
				"rafay_user":                              resourceUser(),
				"rafay_opa_installation_profile":          resourceOPAInstallationProfile(),
				"rafay_access_apikey":                     resourceAccessApikey(),
				"rafay_mesh_profile":                      resourceMeshProfile(),
				"rafay_cluster_mesh_rule":                 resourceClusterMeshRule(),
				"rafay_cluster_mesh_policy":               resourceClusterMeshPolicy(),
				"rafay_namespace_mesh_rule":               resourceNamespaceMeshRule(),
				"rafay_namespace_mesh_policy":             resourceNamespaceMeshPolicy(),
				"rafay_cluster_sharing":                   resourceClusterSharing(),
				"rafay_container_registry":                resourceContainerRegistry(),
				"rafay_cost_profile":                      resourceCostProfile(),
				"rafay_chargeback_group":                  resourceChargebackGroup(),
				"rafay_chargeback_group_report":           resourceChargebackGroupReport(),
				"rafay_chargeback_share":                  resourceChargebackShare(),
				"rafay_cloud_credentials_v3":              resourceCredentials(),
				"rafay_alertconfig":                       resourceAlertConfig(),
				"rafay_organizationalertconfig":           resourceOrganizationAlertConfig(),
				"rafay_tag_group":                         resourceTagGroup(),
				"rafay_project_tags_association":          resourceProjectTagsAssociation(),
				"rafay_static_resource":                   resourceStaticResource(),
				"rafay_config_context":                    resourceConfigContext(),
				"rafay_resource_template":                 resourceResourceTemplate(),
				"rafay_environment_template":              resourceEnvironmentTemplate(),
				"rafay_environment":                       resourceEnvironment(),
				"rafay_fleetplan":                         resourceFleetPlan(),
				"rafay_chargeback_common_services_policy": resourceChargebackCommonServicesPolicy(),
			},
			DataSourcesMap:       map[string]*schema.Resource{},
			ConfigureContextFunc: providerConfigure,
		}

		return p
	}
}

func expandHomeDir(path string) (string, error) {
	if len(path) == 0 || path[0] != '~' {
		return path, nil
	}

	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	return filepath.Join(usr.HomeDir, path[1:]), nil
}

func providerConfigure(ctx context.Context, rd *schema.ResourceData) (interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics

	config_file := rd.Get("provider_config_file").(string)
	ignoreTlsError := rd.Get("ignore_insecure_tls_error").(bool)

	log.Printf("rafay provider config file %s", config_file)
	cliCtx := rctlcontext.GetContext()
	if config_file != "" {
		var err error

		config_file = strings.TrimSpace(config_file)
		if config_file[0] == '~' {
			config_file, err = expandHomeDir(config_file)
		} else {
			config_file, err = filepath.Abs(config_file)
		}

		if err == nil {
			log.Printf("rafay provider config file absolute path %s", config_file)
			configPath := filepath.Dir(config_file)
			fileName := filepath.Base(config_file)
			cliCtx.ConfigFile = fileName
			cliCtx.ConfigDir = configPath
		} else {
			log.Println("failed to get rafay provider config absolute path error:", err)
			log.Println("provider will use default config file ~/.rafay/cli/config.json")
		}
	} else {
		log.Println("provider will use default config file ~/.rafay/cli/config.json")
	}

	err := rctlconfig.InitConfig(cliCtx)

	if err != nil {
		log.Printf("rafay provider config init error %s", err.Error())
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create rafay provider",
			Detail:   "Unable to init config for authenticated rafay provider",
		})
		return nil, diags
	}

	if ignoreTlsError {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	return rctlconfig.GetConfig(), diags

}
