package fromV1

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const Version = 1

func resolve(rawState map[string]interface{}, path, field string) (map[string]interface{}, error) {
	// find and extract the field from the raw state
	attr, ok := rawState[field]
	if !ok {
		return nil, fmt.Errorf("field %s at path %s: not found", field, path)
	}
	// we know attr is an array, so we can cast it
	attrArray, ok := attr.([]interface{})
	if !ok {
		return nil, fmt.Errorf("field %s at path %s: not an array", field, path)
	}
	// we know there is only one element in the array
	if len(attrArray) != 1 {
		return nil, fmt.Errorf("field %s at path %s: invalid array of length %d", field, path, len(attrArray))
	}
	// we know the element is a map
	attrMap, ok := attrArray[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("first attr of field %s at path %s: not a map", field, path)
	}
	return attrMap, nil
}

func Migrate(ctx context.Context, rawState map[string]any, meta any) (map[string]any, error) {
	// spec.0.cluster_config.0.spec.0.managed_cluster.0.properties.0.identity_profile
	spec, err := resolve(rawState, "", "spec")
	if err != nil {
		return nil, err
	}
	clusterConfig, err := resolve(spec, "spec", "cluster_config")
	if err != nil {
		return nil, err
	}
	clusterConfigSpec, err := resolve(clusterConfig, "spec.cluster_config", "spec")
	if err != nil {
		return nil, err
	}
	managedCluster, err := resolve(clusterConfigSpec, "spec.cluster_config.spec", "managed_cluster")
	if err != nil {
		return nil, err
	}
	properties, err := resolve(managedCluster, "spec.cluster_config.spec.managed_cluster", "properties")
	if err != nil {
		return nil, err
	}
	// Migrate if the state has identity profile as a map
	if identityProfile, ok := properties["identity_profile"]; ok {
		if m, ok := identityProfile.(map[string]interface{}); ok && len(m) == 0 {
			// identity profile is a map, we need to migrate
			properties["identity_profile"] = []interface{}{}
		}
	}
	return rawState, nil
}

// Copied over from https://github.com/RafaySystems/terraform-provider-rafay/blob/v2.4.x/rafay/resource_aks_cluster.go

func Resource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"apiversion": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "apiversion",
			},
			"kind": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "Cluster",
				Description: "kind",
			},
			"metadata": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "AKS specific cluster configuration",
				Elem: &schema.Resource{
					Schema: clusterAKSClusterMetadata(),
				},
			},
			"spec": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "AKS specific cluster configuration",
				Elem: &schema.Resource{
					Schema: clusterAKSClusterSpec(),
				},
			},
		},
	}
}

func clusterAKSClusterMetadata() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "AKS Cluster name",
		},
		"project": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Project for the cluster",
		},
		"labels": {
			Type:        schema.TypeMap,
			Optional:    true,
			Description: "labels for the cluster",
		},
	}
	return s
}

func clusterAKSClusterSpec() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"type": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "AKS Cluster type",
		},
		"blueprint": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "default-aks",
			Description: "Blueprint to be associated with the cluster. Default will be default-aks",
		},
		"blueprintversion": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Blueprint version to be associated with the cluster. Default will be the latest version",
		},
		"cloudprovider": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Cloud credentials provider used to create and manage the cluster.",
		},
		"system_components_placement": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "Configure tolerations and nodeSelector for Rafay system components.",
			Elem: &schema.Resource{
				Schema: systemComponentsPlacementFields(),
			},
		},
		"cluster_config": {
			Type:        schema.TypeList,
			Required:    true,
			Description: "AKS specific cluster configuration	",
			Elem: &schema.Resource{
				Schema: clusterAKSClusterConfig(),
			},
		},
		"sharing": &schema.Schema{
			Description: "blueprint sharing configuration",
			Elem: &schema.Resource{Schema: map[string]*schema.Schema{
				"enabled": &schema.Schema{
					Description: "flag to specify if sharing is enabled for resource",
					Optional:    true,
					Type:        schema.TypeBool,
				},
				"projects": &schema.Schema{
					Description: "list of projects this resource is shared to",
					Elem: &schema.Resource{Schema: map[string]*schema.Schema{"name": &schema.Schema{
						Description: "name of the project",
						Optional:    true,
						Type:        schema.TypeString,
					}}},
					MaxItems: 0,
					MinItems: 0,
					Optional: true,
					Type:     schema.TypeList,
				},
			}},
			MaxItems: 1,
			MinItems: 1,
			Optional: true,
			Type:     schema.TypeList,
		},
	}
	return s
}

func clusterAKSClusterConfig() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"apiversion": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "apiversion",
		},
		"kind": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "aksClusterConfig",
			Description: "kind",
		},
		"metadata": {
			Type:        schema.TypeList,
			Required:    true,
			Description: "AKS specific cluster configuration metadata",
			Elem: &schema.Resource{
				Schema: clusterAKSClusterConfigMetadata(),
			},
		},
		"spec": {
			Type:        schema.TypeList,
			Required:    true,
			Description: "AKS specific cluster configuration spec",
			Elem: &schema.Resource{
				Schema: clusterAKSClusterConfigSpec(),
			},
		},
	}
	return s
}

func clusterAKSClusterConfigMetadata() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "AKS cluster name",
		},
	}
	return s
}

func clusterAKSClusterConfigSpec() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"subscription_id": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The AKS subscription id",
		},
		"resource_group_name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Resource Group for the cluster",
		},
		"managed_cluster": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "The AKS managed cluster",
			Elem: &schema.Resource{
				Schema: clusterAKSManagedCluster(),
			},
		},
		"node_pools": {
			Type:        schema.TypeList,
			Required:    true,
			Description: "The Aks Node Pool",
			Elem: &schema.Resource{
				Schema: clusterAKSNodePool(),
			},
		},
	}
	return s
}

func clusterAKSManagedCluster() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"apiversion": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Azure resource managed cluster api version.",
		},
		"extended_location": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "The AKS managed cluster extended location",
			Elem: &schema.Resource{
				Schema: clusterAKSManagedClusterExtendedLocation(),
			},
		},
		"identity": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "The AKS managed cluster extended location",
			Elem: &schema.Resource{
				Schema: clusterAKSManagedClusterIdentity(),
			},
		},
		"location": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "AKS cluster location",
		},
		"properties": {
			Type:        schema.TypeList,
			Required:    true,
			Description: "Properties of the managed cluster.",
			Elem: &schema.Resource{
				Schema: clusterAKSManagedClusterProperties(),
			},
		},
		"sku": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "The SKU of a Managed Cluster.",
			Elem: &schema.Resource{
				Schema: clusterAKSManagedClusterSKU(),
			},
		},
		"tags": {
			Type:        schema.TypeMap,
			Optional:    true,
			Description: "Resource tags",
		},
		"type": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "Microsoft.ContainerService/managedClusters",
			Description: "Type",
		},
		"additional_metadata": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Additional metadata associated with the managed cluster.",
			Elem: &schema.Resource{
				Schema: clusterAKSManagedClusterAdditionalMetadata(),
			},
		},
	}
	return s
}

func clusterAKSManagedClusterExtendedLocation() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The AKS managed cluster extended location name",
		},
		"type": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The AKS managed cluster extended location type",
		},
	}
	return s
}

func clusterAKSManagedClusterIdentity() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"type": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Identity type for the AKS cluster. For more information see use managed identities in AKS. Valid values are SystemAssigned, UserAssigned, None.",
		},
		"user_assigned_identities": {
			Type:        schema.TypeMap,
			Optional:    true,
			Description: "Arm Resource Ids",
		},
	}
	return s
}

func clusterAKSManagedClusterProperties() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"aad_profile": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "The AKS managed cluster AAD Profile",
			Elem: &schema.Resource{
				Schema: clusterAKSManagedClusterPropertiesAadProfile(),
			},
		}, /*
			"addon_profiles": { //make change to string json like attach policy
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "The AKS managed cluster addon profiles",
			},*/
		"addon_profiles": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "The AKS managed cluster addon profiles",
			Elem: &schema.Resource{
				Schema: addonProfileFields(),
			},
		},
		"api_server_access_profile": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "The AKS managed cluster api server access profile",
			Elem: &schema.Resource{
				Schema: clusterAKSManagedClusterAPIServerAccessProfile(),
			},
		},
		"auto_scaler_profile": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Parameters to be applied to the cluster-autoscaler when enabled",
			Elem: &schema.Resource{
				Schema: clusterAKSManagedClusterAutoScalerProfile(),
			},
		},
		"auto_upgrade_profile": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "The AKS managed cluster autoupgrade profile",
			Elem: &schema.Resource{
				Schema: clusterAKSManagedClusterAutoUpgradeProfile(),
			},
		},
		"disable_local_accounts": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "The AKS managed cluster addon profiles",
		},
		"disk_encryption_set_id": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "This is of the form: /subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.Compute/diskEncryptionSets/{encryptionSetName}",
		},
		"dns_prefix": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "This cannot be updated once the Managed Cluster has been created.",
		},
		"enable_pod_security_policy": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "(DEPRECATED) Whether to enable Kubernetes pod security policy (preview). This feature is set for removal on October 15th, 2020.",
		},
		"enable_rbac": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "Whether to enable Kubernetes Role-Based Access Control.",
		},
		"fqdn_subdomain": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "This cannot be updated once the Managed Cluster has been created",
		},
		"http_proxy_config": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Cluster HTTP proxy configuration.",
			Elem: &schema.Resource{
				Schema: clusterAKSManagedClusterHTTPProxyConfig(),
			},
		},
		"identity_profile": {
			Type:        schema.TypeMap,
			Optional:    true,
			Description: "Identities associated with the cluster",
		},
		"kubernetes_version": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Kubernetes version",
		},
		"linux_profile": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Profile for Linux VMs in the container service cluster.",
			Elem: &schema.Resource{
				Schema: clusterAKSManagedClusterLinuxProfile(),
			},
		},
		"network_profile": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Profile of network configuration.",
			Elem: &schema.Resource{
				Schema: clusterAKSManagedClusterNetworkProfile(),
			},
		},
		"node_resource_group": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The name of the resource group containing agent pool nodes.",
		},
		"oidc_issuer_profile": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Profile of OIDC Issuer configuration.",
			Elem: &schema.Resource{
				Schema: clusterAKSManagedClusterOidcIssuerProfile(),
			},
		},
		"pod_identity_profile": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Aspect of pod identity integration.",
			Elem: &schema.Resource{
				Schema: clusterAKSManagedClusterPodIdentityProfile(),
			},
		},
		"private_link_resources": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Private link resources associated with the cluster.",
			Elem: &schema.Resource{
				Schema: clusterAKSManagedClusterPrivateLinkResources(),
			},
		},
		"security_profile": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Profile of security configuration.",
			Elem: &schema.Resource{
				Schema: clusterAKSManagedClusterSecurityProfile(),
			},
		},
		"service_principal_profile": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Information about a service principal identity for the cluster to use for manipulating Azure APIs.",
			Elem: &schema.Resource{
				Schema: clusterAKSManagedClusterServicePrincipalProfile(),
			},
		},
		"windows_profile": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Profile for Windows VMs in the managed cluster.",
			Elem: &schema.Resource{
				Schema: clusterAKSManagedClusterWindowsProfile(),
			},
		},
	}
	return s
}

func addonProfileFields() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"http_application_routing": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Config for HTTP Application Routing Addon Profile",
			Elem: &schema.Resource{
				Schema: aKSManagedClusterAddonProfile(),
			},
		},
		"azure_policy": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Config for Azure Policy in Addon Profile",
			Elem: &schema.Resource{
				Schema: aKSManagedClusterAddonProfile(),
			},
		},
		"oms_agent": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Config for OMS Agent in Addon Profile",
			Elem: &schema.Resource{
				Schema: aKSManagedClusterAddonOmsAgentProfile(),
			},
		},
		"azure_keyvault_secrets_provider": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Azure Keyvault Secrets Provider for AKS",
			Elem: &schema.Resource{
				Schema: aKSManagedClusterAddonAzureKeyvaultSecretsProviderProfile(),
			},
		},
		"ingress_application_gateway": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Azure Ingress Application Gateway Addon for AKS",
			Elem: &schema.Resource{
				Schema: aKSManagedClusterAddonIngressApplicationGatewayProfile(),
			},
		},
	}
	return s
}

func aKSManagedClusterAddonProfile() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"enabled": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "Whether to enable HTTP Application Routing or Azure Policy in Addon Profile",
		},
		"config": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Config for HTTP Application Routing or Azure Policy in Addon Profile",
		},
	}
	return s
}

func aKSManagedClusterAddonOmsAgentProfile() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"enabled": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "Whether to enable OMS Agent in Addon Profile",
		},
		"config": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Config for OMS Agent in Addon Profile",
			Elem: &schema.Resource{
				Schema: aKSManagedClusterAddonOmsAgentConfigProfile(),
			},
		},
	}
	return s
}

func aKSManagedClusterAddonIngressApplicationGatewayProfile() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"enabled": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "Whether to enable Ingress Application Gateway in Addon Profile",
		},
		"config": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Config for Ingress Application Gateway in Addon Profile",
			Elem: &schema.Resource{
				Schema: aKSManagedClusterAddonIngressApplicationGatewayConfigProfile(),
			},
		},
	}
	return s
}

func aKSManagedClusterAddonIngressApplicationGatewayConfigProfile() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"application_gateway_name": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Name of the application gateway to create/use in the node resource group.",
		},
		"application_gateway_id": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Resource Id of an existing Application Gateway to use with AGIC.",
		},
		"subnet_cidr": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Subnet CIDR to use for a new subnet created to deploy the Application Gateway.",
		},
		"subnet_id": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Resource Id of an existing Subnet used to deploy the Application Gateway.",
		},
		"watch_namespace": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Specify the namespace, which AGIC should watch. This could be a single string value, or a comma-separated list of namespaces.",
		},
	}
	return s
}

func aKSManagedClusterAddonAzureKeyvaultSecretsProviderProfile() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"enabled": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "Whether to enable Azure Key Vault Secrets Provider in Addon Profile",
		},
		"config": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Config Azure Key Vault Secrets Provider in Addon Profile",
			Elem: &schema.Resource{
				Schema: aKSManagedClusterAddonAzureKeyvaultSecretsProviderConfigProfile(),
			},
		},
	}
	return s
}

func aKSManagedClusterAddonOmsAgentConfigProfile() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"log_analytics_workspace_resource_id": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "ID of the log analytics workspace",
		},
	}
	return s
}

func aKSManagedClusterAddonAzureKeyvaultSecretsProviderConfigProfile() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"enable_secret_rotation": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Whether to enable Secret Rotation",
		},
		"rotation_poll_interval": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Interval to poll for secret rotation",
		},
	}
	return s
}

func clusterAKSManagedClusterPropertiesAadProfile() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"admin_group_object_ids": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "The AKS managed cluster aad profile admin group object ids",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"client_app_id": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The AKS managed cluster aad profile client app id",
		},
		"enable_azure_rbac": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "Whether or not to enable azure rbac for kubernetes authorization",
		},
		"managed": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "Whether or not to enable managed aad",
		},
		"server_app_id": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The server AAD application ID.",
		},
		"server_app_secret": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The AKS managed cluster aad profile server app secret",
		},
		"tenant_id": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The AKS managed cluster tenant id",
		},
	}
	return s
}

func clusterAKSManagedClusterAPIServerAccessProfile() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"authorized_ipr_ranges": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "The AKS managed cluster properties server access profile server access profile",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"enable_private_cluster": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
			Description: "Enable private cluster",
		},
		"enable_private_cluster_public_fqdn": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "Whether or not to create additional public fqdn for private cluster",
		},
		"private_dns_zone": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "System",
			Description: "The AKS managed cluster properties private dns zone",
		},
	}
	return s
}

func clusterAKSManagedClusterAutoScalerProfile() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"balance_similar_node_groups": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Valid values are true or false",
		},
		"expander": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "random",
			Description: "Valid values are least-waste, most-pods, priority, random",
		},
		"max_empty_bulk_delete": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "10",
			Description: "Max empty bulk delete",
		},
		"max_graceful_termination_sec": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "600",
			Description: "Max graceful termination sec",
		},
		"max_node_provision_time": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "15m",
			Description: "Values must be an integer followed by an m. No unit of time other than minutes (m) is supported",
		},
		"max_total_unready_percentage": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "45",
			Description: "The maximum is 100 and the minimum is 0",
		},
		"new_pod_scale_up_delay": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "0s",
			Description: "For scenarios like burst/batch scale where you don't want CA to act before the kubernetes scheduler could schedule all the pods, you can tell CA to ignore unscheduled pods before they're a certain age.",
		},
		//@@@@@@@@@@@@@ Listed as string in schema @@@@@@@@@@@@
		"ok_total_unready_count": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     3,
			Description: "This must be an integer.",
		},
		"scale_down_delay_after_add": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "10m",
			Description: "Values must be an integer followed by an m. No unit of time other than minutes (m) is supported",
		},
		"scale_down_delay_after_delete": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The default is the scan-interval. Values must be an integer followed by an m",
		},
		"scale_down_delay_after_failure": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "3m",
			Description: " Values must be an integer followed by an m",
		},
		"scale_down_unneeded_time": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "10m",
			Description: "Values must be an integer followed by an m",
		},
		"scale_down_unready_time": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "20m",
			Description: "Values must be an integer followed by an m",
		},
		"scale_down_utilization_threshold": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "0.5",
			Description: "The scale down utilization threshold",
		},
		"scan_interval": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "10",
			Description: "The default is 10. Values must be an integer number of seconds",
		},
		"skip_nodes_with_local_storage": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "true",
			Description: "Skip nodes with local storage",
		},
		"skip_nodes_with_system_pods": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "true",
			Description: "Skip nodes with system pods",
		},
	}
	return s
}

func clusterAKSManagedClusterAutoUpgradeProfile() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"upgrade_channel": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Valid values are rapid, stable, patch, node-image, none",
		},
	}
	return s
}

func clusterAKSManagedClusterHTTPProxyConfig() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"http_proxy": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The HTTP proxy server endpoint to use.",
		},
		"https_proxy": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The HTTPs proxy server endpoint to use.",
		},
		"no_proxy": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "The endpoints that should not go through proxy.",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"trusted_ca": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Alternative CA cert to use for connecting to proxy servers.",
		},
	}
	return s
}

func clusterAKSManagedClusterLinuxProfile() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"admin_username": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The administrator username to use for Linux VMs.",
		},
		"ssh": {
			Type:        schema.TypeList,
			Required:    true,
			Description: "SSH configuration for Linux-based VMs running on Azure.",
			Elem: &schema.Resource{
				Schema: clusterAKSManagedClusterSSHConfig(),
			},
		},
		"no_proxy": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "The endpoints that should not go through proxy.",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"trusted_ca": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Alternative CA cert to use for connecting to proxy servers.",
		},
	}
	return s
}

func clusterAKSManagedClusterSSHConfig() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"public_keys": {
			Type:        schema.TypeList,
			Required:    true,
			MaxItems:    1,
			MinItems:    1,
			Description: "The list of SSH public keys used to authenticate with Linux-based VMs. A maximum of 1 key may be specified.",
			Elem: &schema.Resource{
				Schema: clusterAKSManagedClusterSSHKeyData(),
			},
		},
	}
	return s
}

func clusterAKSManagedClusterSSHKeyData() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"key_data": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Certificate public key used to authenticate with VMs through SSH. The certificate must be in PEM format with or without headers.",
		},
	}
	return s
}

func clusterAKSManagedClusterNetworkProfile() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"dns_service_ip": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "An IP address assigned to the Kubernetes DNS service.",
		},
		"docker_bridge_cidr": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "A CIDR notation IP range assigned to the Docker bridge network.",
		},
		"load_balancer_profile": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Profile of the managed cluster load balancer.",
			Elem: &schema.Resource{
				Schema: clusterAKSManagedClusterNPLoadBalancerProfile(),
			},
		},
		"load_balancer_sku": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "Standard",
			Description: "Valid values are standard, basic.",
		},
		"network_mode": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "This cannot be specified if networkPlugin is anything other than azure.",
		},
		"network_plugin": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "kubenet",
			Description: "Network plugin used for building the Kubernetes network. Valid values are azure, kubenet.",
		},
		"network_policy": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Network policy used for building the Kubernetes network. Valid values are calico, azure.",
		},
		"outbound_type": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "This can only be set at cluster creation time and cannot be changed later. Valid values are loadBalancer, userDefinedRouting.",
		},
		"pod_cidr": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "A CIDR notation IP range from which to assign pod IPs when kubenet is used.",
		},
		"service_cidr": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "A CIDR notation IP range from which to assign service cluster IPs.",
		},
	}
	return s
}

func clusterAKSManagedClusterNPLoadBalancerProfile() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"allocated_outbound_ports": {
			Type:        schema.TypeInt,
			Optional:    true,
			Default:     0,
			Description: "The desired number of allocated SNAT ports per VM.",
		},
		"effective_outbound_ips": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "The effective outbound IP resources of the cluster load balancer.",
			Elem: &schema.Resource{
				Schema: clusterAKSManagedClusterNPEffectiveOutboundIPs(),
			},
		},
		"idle_timeout_in_minutes": {
			Type:        schema.TypeInt,
			Optional:    true,
			Default:     30,
			Description: "Desired outbound flow idle timeout in minutes.",
		},
		"managed_outbound_ips": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Desired managed outbound IPs for the cluster load balancer.",
			Elem: &schema.Resource{
				Schema: clusterAKSManagedClusterNPManagedOutboundIPs(),
			},
		},
		"outbound_ip_prefixes": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Desired managed outbound IPs for the cluster load balancer.",
			Elem: &schema.Resource{
				Schema: clusterAKSManagedClusterNPOutboundIPPrefixes(),
			},
		},
		"outbound_ips": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Desired outbound IP resources for the cluster load balancer.",
			Elem: &schema.Resource{
				Schema: clusterAKSManagedClusterNPOutboundIPs(),
			},
		},
	}
	return s
}

func clusterAKSManagedClusterNPEffectiveOutboundIPs() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"id": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The fully qualified Azure resource id.",
		},
	}
	return s
}

func clusterAKSManagedClusterNPManagedOutboundIPs() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"count": {
			Type:        schema.TypeInt,
			Optional:    true,
			Default:     1,
			Description: "The desired number of outbound IPs created/managed by Azure for the cluster load balancer.",
		},
	}
	return s
}

func clusterAKSManagedClusterNPOutboundIPPrefixes() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"public_ip_prefixes": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "A list of public IP prefix resources.",
			Elem: &schema.Resource{
				Schema: clusterAKSManagedClusterNPManagedOutboundIPsPublicIpPrefixes(),
			},
		},
	}
	return s
}

func clusterAKSManagedClusterNPManagedOutboundIPsPublicIpPrefixes() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"id": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The fully qualified Azure resource id.",
		},
	}
	return s
}

func clusterAKSManagedClusterNPOutboundIPs() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"public_ips": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "A list of public IP resources.",
			Elem: &schema.Resource{
				Schema: clusterAKSManagedClusterNPOutboundIPsPublicIps(),
			},
		},
	}
	return s
}

func clusterAKSManagedClusterNPOutboundIPsPublicIps() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"id": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: " 	The fully qualified Azure resource id",
		},
	}
	return s
}

func clusterAKSManagedClusterOidcIssuerProfile() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"enabled": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "Whether to enable the OIDC issuer for the cluster.",
		},
	}
	return s
}

func clusterAKSManagedClusterPodIdentityProfile() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"allow_network_plugin_kubenet": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "Running in Kubenet is disabled by default due to the security related nature of AAD Pod Identity and the risks of IP spoofing.",
		},
		"enabled": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "Whether the pod identity addon is enabled.",
		},
		"user_assigned_identities": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "The pod identities to use in the cluster.",
			Elem: &schema.Resource{
				Schema: clusterAKSManagedClusterPIPUserAssignedIdentities(),
			},
		},
		"user_assigned_identity_exceptions": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "The pod identity exceptions to allow.",
			Elem: &schema.Resource{
				Schema: clusterAKSManagedClusterPIPUserAssignedIdentityExceptions(),
			},
		},
	}
	return s
}

func clusterAKSManagedClusterPIPUserAssignedIdentities() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"binding_selector": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The binding selector to use for the AzureIdentityBinding resource.",
		},
		"identity": {
			Type:        schema.TypeList,
			Required:    true,
			Description: "Details about a user assigned identity.",
			Elem: &schema.Resource{
				Schema: clusterAKSManagedClusterUAIIdentity(),
			},
		},
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The name of the pod identity.",
		},
		"namespace": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The namespace of the pod identity.",
		},
	}
	return s
}

func clusterAKSManagedClusterUAIIdentity() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"client_id": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The client ID of the user assigned identity.",
		},
		"object_id": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The object ID of the user assigned identity.",
		},
		"resource_id": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The resource ID of the user assigned identity.",
		},
	}

	return s
}

func clusterAKSManagedClusterPIPUserAssignedIdentityExceptions() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The name of the pod identity.",
		},
		"namespace": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The namespace of the pod identity.",
		},
		"pod_labels": {
			Type:        schema.TypeMap,
			Required:    true,
			Description: "The pod labels to match.",
		},
	}
	return s
}

func clusterAKSManagedClusterPrivateLinkResources() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"group_id": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The group ID of the resource.",
		},
		"id": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The ID of the private link resource.",
		},
		"name": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The name of the private link resource.",
		},
		"required_members": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "The RequiredMembers of the resource",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"type": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The resource type.",
		},
	}
	return s
}

func clusterAKSManagedClusterSecurityProfile() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"workload_identity": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Profile of the managed cluster workload identity.",
			Elem: &schema.Resource{
				Schema: clusterAKSManagedClusterWorkloadIdentityProfile(),
			},
		},
	}
	return s
}

func clusterAKSManagedClusterWorkloadIdentityProfile() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"enabled": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Whether to enable the Workload Identity for the cluster.",
		},
	}
	return s
}

func clusterAKSManagedClusterServicePrincipalProfile() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"client_id": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "FORMATTED:The ID for the service principal. If specified, must be set to `[parameters('servicePrincipalClientId')]`. This would be set to the cloud credential's client ID during cluster deployment.",
		},
		"secret": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The secret password associated with the service principal in plain text.",
		},
	}
	return s
}

func clusterAKSManagedClusterWindowsProfile() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"admin_username": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Specifies the name of the administrator account.",
		},
		"enable_csi_proxy": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "Enable CSI proxy",
		},
		"license_type": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The license type to use for Windows VMs.",
		},
	}
	return s
}

func clusterAKSManagedClusterSKU() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The name of a managed cluster SKU.",
		},
		"tier": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "Free",
			Description: " Valid values are Paid, Free.",
		},
	}
	return s
}

func clusterAKSManagedClusterAdditionalMetadata() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"acr_profile": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Profile for Azure Container Registry configuration",
			Elem: &schema.Resource{
				Schema: clusterAKSManagedClusterAdditionalMetadataACRProfile(),
			},
		},
		"oms_workspace_location": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "If not specified, defaults to the resource group of the managed cluster. Valid only if the Log analytics workspace is specified.",
		},
	}
	return s
}

func clusterAKSManagedClusterAdditionalMetadataACRProfile() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"resource_group_name": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "If not specified, defaults to the resource group of the managed cluster",
		},
		"acr_name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The name of the Azure Container Registry resource.",
		},
	}
	return s
}

func clusterAKSNodePool() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"apiversion": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The AKS node pool api version",
		},
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The AKS node pool name",
		},
		"properties": {
			Type:        schema.TypeList,
			Required:    true,
			Description: "The AKS managed cluster",
			Elem: &schema.Resource{
				Schema: clusterAKSNodePoolProperties(),
			},
		},
		"type": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "Microsoft.ContainerService/managedClusters/agentPools",
			Description: "The AKS node pool type",
		},
		"location": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "AKS cluster location",
		},
	}

	return s
}

func clusterAKSNodePoolProperties() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"availability_zones": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "The list of Availability zones to use for nodes. This can only be specified if the AgentPoolType property is VirtualMachineScaleSets.",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"count": {
			Type:        schema.TypeInt,
			Optional:    true,
			Default:     1,
			Description: "Number of agents (VMs) to host docker containers. Allowed values must be in the range of 0 to 1000 (inclusive) for user pools and in the range of 1 to 1000 (inclusive) for system pools. The default value is 1.",
		},
		"enable_auto_scaling": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "Whether to enable auto-scaler",
		},
		"enable_encryption_at_host": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "This is only supported on certain VM sizes and in certain Azure regions.",
		},
		"enable_fips": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "See Add a FIPS-enabled node pool for more details.",
		},
		"enable_node_public_ip": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "Some scenarios may require nodes in a node pool to receive their own dedicated public IP addresses. A common scenario is for gaming workloads, where a console needs to make a direct connection to a cloud virtual machine to minimize hops. For more information see assigning a public IP per node. The default is false.",
		},
		"enable_ultra_ssd": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "Whether to enable UltraSSD",
		},
		"gpu_instance_profile": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "GPUInstanceProfile to be used to specify GPU MIG instance profile for supported GPU VM SKU.",
		},
		"kubelet_config": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "See AKS custom node configuration for more details.",
			Elem: &schema.Resource{
				Schema: clusterAKSNodePoolKubeletConfig(),
			},
		},
		"kubelet_disk_type": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Valid values are OS, Temporary.",
		},
		"linux_os_config": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "See AKS custom node configuration for more details.",
			Elem: &schema.Resource{
				Schema: clusterAKSNodePoolLinuxOsConfig(),
			},
		},
		"max_count": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "The maximum number of nodes for auto-scaling.",
		},
		"max_pods": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "The maximum number of pods that can run on a node.",
		},
		"mode": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "System",
			Description: "The mode for a node pool which defines a node pool's primary function. If set as 'System', AKS prefers system pods scheduling to node pools with mode System. Accepted values: System, User",
		},
		"min_count": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "The minimum number of nodes for auto-scaling",
		},
		"node_labels": {
			Type:     schema.TypeMap,
			Optional: true,
			// COMPUTED OPTION?
			//Computed:    true,
			Description: "Valid values are System, User.",
		},
		"node_public_ip_prefix_id": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "This is of the form: /subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.Network/publicIPPrefixes/{publicIPPrefixName}",
		},
		"node_taints": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "The taints added to new nodes during node pool create and scale. For example, key=value:NoSchedule.",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"orchestrator_version": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The AKS node pool Kubernetes version",
		},
		"os_disk_size_gb": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "OS Disk Size in GB to be used to specify the disk size for every machine in the master/agent pool. ",
		},
		"os_disk_type": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Valid values are Managed, Ephemeral.",
		},
		"os_sku": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Valid values are Ubuntu, CBLMariner.",
		},
		"os_type": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "Linux",
			Description: "Valid values are Linux, Windows.",
		},
		"pod_subnet_id": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "If omitted, pod IPs are statically assigned on the node subnet (see vnetSubnetID for more details). This is of the form: /subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.Network/virtualNetworks/{virtualNetworkName}/subnets/{subnetName}",
		},
		"proximity_placement_group_id": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The ID for Proximity Placement Group.",
		},
		"scale_set_eviction_policy": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "Delete",
			Description: "This cannot be specified unless the scaleSetPriority is Spot",
		},
		"scale_set_priority": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "Regular",
			Description: "The Virtual Machine Scale Set priority.",
		},
		"spot_max_price": {
			Type:        schema.TypeFloat,
			Optional:    true,
			Description: "Possible values are any decimal value greater than zero or -1 which indicates the willingness to pay any on-demand price. For more details on spot pricing, see spot VMs pricing",
		},
		"tags": {
			Type:        schema.TypeMap,
			Optional:    true,
			Description: "The tags to be persisted on the agent pool virtual machine scale set.",
		},
		"type": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "VirtualMachineScaleSets",
			Description: "Valid values are VirtualMachineScaleSets, AvailabilitySet.",
		},
		"upgrade_settings": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Settings for upgrading an agentpool",
			Elem: &schema.Resource{
				Schema: clusterAKSNodePoolUpgradeSettings(),
			},
		},
		"vm_size": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "Standard_DS2_v2",
			Description: "The AKS node pool VM size",
		},
		"vnet_subnet_id": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "If this is not specified, a VNET and subnet will be generated and used. If no podSubnetID is specified, this applies to nodes and pods, otherwise it applies to just nodes.",
		},
	}
	return s
}

func clusterAKSNodePoolKubeletConfig() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"allowed_unsafe_sysctls": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Allowed list of unsafe sysctls or unsafe sysctl patterns (ending in *).",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"container_log_max_files": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "The maximum number of container log files that can be present for a container. The number must be ≥ 2.",
		},
		"container_log_max_size_mb": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "The maximum size (e.g. 10Mi) of container log file before it is rotated.",
		},
		"cpu_cfs_quota": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "The default is true.",
		},
		"cpu_cfs_quota_period": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "100ms",
			Description: "Valid values are a sequence of decimal numbers with an optional fraction and a unit suffix.",
		},
		"cpu_manager_policy": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "none",
			Description: "See Kubernetes CPU management policies for more information",
		},
		"fail_swap_on": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "If set to true it will make the Kubelet fail to start if swap is enabled on the node.",
		},
		"image_gc_high_threshold": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "To disable image garbage collection, set to 100. The default is 85%",
		},
		"image_gc_low_threshold": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "This cannot be set higher than imageGcHighThreshold. The default is 80%",
		},
		"pod_max_pids": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "The maximum number of processes per pod.",
		},
		"topology_manager_policy": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Allowed values are none, best-effort, restricted, and single-numa-node.",
		},
	}
	return s
}

func clusterAKSNodePoolLinuxOsConfig() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"swap_file_size_mb": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "The size in MB of a swap file that will be created on each node.",
		},
		"sysctls": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Sysctl settings for Linux agent nodes.",
			Elem: &schema.Resource{
				Schema: clusterAKSNodePoolLinuxOsConfigSysctls(),
			},
		},
		"transparent_huge_page_defrag": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "madvise",
			Description: "Valid values are always, defer, defer+madvise, madvise and never.",
		},
		"transparent_huge_page_enabled": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "always",
			Description: "Valid values are always, madvise, and never.",
		},
	}
	return s

}

func clusterAKSNodePoolLinuxOsConfigSysctls() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"fs_aio_max_nr": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "Sysctl setting fs.aio-max-nr.",
		},
		"fs_file_max": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "Sysctl setting fs.file-max.",
		},
		"fs_inotify_max_user_watches": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "Sysctl setting fs.inotify.max_user_watches.",
		},
		"fs_nr_open": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "Sysctl setting fs.nr_open.",
		},
		"kernel_threads_max": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "Sysctl setting kernel.threads-max.",
		},
		"net_core_netdev_max_backlog": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "Sysctl setting net.core.netdev_max_backlog.",
		},
		"net_core_optmem_max": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "Sysctl setting net.core.optmem_max.",
		},
		"net_core_rmem_default": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "Sysctl setting net.core.rmem_default.",
		},
		"net_core_rmem_max": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "Sysctl setting net.core.rmem_max.",
		},
		"net_core_somaxconn": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "Sysctl setting net.core.somaxconn.",
		},
		"net_core_wmem_default": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "Sysctl setting net.core.wmem_default.",
		},
		"net_core_wmem_max": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "Sysctl setting net.core.wmem_max.",
		},
		"net_ipv4_ip_local_port_range": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Sysctl setting net.ipv4.ip_local_port_range.",
		},
		"net_ipv4_neigh_default_gc_thresh1": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "Sysctl setting net.ipv4.neigh.default.gc_thresh1.",
		},
		"net_ipv4_neigh_default_gc_thresh2": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "Sysctl setting net.ipv4.neigh.default.gc_thresh2.",
		},
		"net_ipv4_neigh_default_gc_thresh3": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "Sysctl setting net.ipv4.neigh.default.gc_thresh3.",
		},
		"net_ipv4_tcp_fin_timeout": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "Sysctl setting net.ipv4.tcp_fin_timeout.",
		},
		"net_ipv4_tcpkeepalive_intvl": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "Sysctl setting net.ipv4.tcp_keepalive_intvl.",
		},
		"net_ipv4_tcp_keepalive_probes": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "Sysctl setting net.ipv4.tcp_keepalive_probes.",
		},
		"net_ipv4_tcp_keepalive_time": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "Sysctl setting net.ipv4.tcp_keepalive_time.",
		},
		"net_ipv4_tcp_max_syn_backlog": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "Sysctl setting net.ipv4.tcp_max_syn_backlog.",
		},
		"net_ipv4_tcp_max_tw_buckets": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "Sysctl setting net.ipv4.tcp_max_tw_buckets.",
		},
		"net_ipv4_tcp_tw_reuse": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "Sysctl setting net.ipv4.tcp_tw_reuse.",
		},
		"net_netfilter_nf_conntrack_buckets": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "Sysctl setting net.netfilter.nf_conntrack_buckets.",
		},
		"net_netfilter_nf_conntrack_max": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "Sysctl setting net.netfilter.nf_conntrack_max.",
		},
		"vm_max_map_count": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "Sysctl setting vm.max_map_count.",
		},
		"vm_swappiness": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "Sysctl setting vm.swappiness.",
		},
		"vm_vfs_cache_pressure": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "Sysctl setting vm.vfs_cache_pressure.",
		},
	}
	return s
}

func clusterAKSNodePoolUpgradeSettings() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"max_surge": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "This can either be set to an integer (e.g. 5) or a percentage (e.g. 50%)",
		},
	}
	return s
}

// Copied over from https://github.com/RafaySystems/terraform-provider-rafay/blob/v2.4.x/rafay/cluster_util.go

func systemComponentsPlacementFields() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"node_selector": {
			Type:        schema.TypeMap,
			Optional:    true,
			Description: "Key-Value pairs insuring pods to be scheduled on desired nodes.",
		},
		"tolerations": {
			Type: schema.TypeList,
			//Type:        schema.TypeString,
			Optional:    true,
			Description: "Enables the kuberenetes scheduler to schedule pods with matching taints.",
			Elem: &schema.Resource{
				Schema: tolerationsFields(),
			},
		},
		"daemonset_override": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Allows users to override the default behaviour of DaemonSet for specific nodes, enabling the addition of additional tolerations for Daemonsets to match the taints available on the nodes.",
			Elem: &schema.Resource{
				Schema: daemonsetOverrideFields(),
			},
		},
	}
	return s
}

func tolerationsFields() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"key": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "the taint key that the toleration applies to",
		},
		"operator": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "represents a key's relationship to the value",
		},
		"value": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "the taint value the toleration matches to",
		},
		"effect": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "indicates the taint effect to match",
		},
		"toleration_seconds": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "represents the period of time the toleration tolerates the taint",
		},
	}
	return s
}

func daemonsetOverrideFields() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"node_selection_enabled": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "enables node selection",
		},
		"tolerations": {
			Type: schema.TypeList,
			//Type:        schema.TypeString,
			Optional:    true,
			Description: "Additional tolerations for Daemonsets to match the taints available on the nodes",
			Elem: &schema.Resource{
				Schema: tolerationsFields(),
			},
		},
	}
	return s
}
