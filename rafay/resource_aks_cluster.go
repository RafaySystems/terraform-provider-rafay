package rafay

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/RafaySystems/edge-common/pkg/models/edge"
	"github.com/RafaySystems/rctl/pkg/cluster"
	"github.com/RafaySystems/rctl/pkg/clusterctl"
	"github.com/RafaySystems/rctl/pkg/config"
	glogger "github.com/RafaySystems/rctl/pkg/log"
	"github.com/RafaySystems/rctl/pkg/project"
	"github.com/RafaySystems/terraform-provider-rafay/rafay/migrate/aks/fromV1"
	"github.com/davecgh/go-spew/spew"
	jsoniter "github.com/json-iterator/go"
	"go.uber.org/zap"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	// Yaml pkg that have no limit for key length
	"github.com/go-yaml/yaml"
	yamlf "github.com/goccy/go-yaml"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type clusterCTLResponse struct {
	TaskSetID  string                 `json:"taskset_id,omitempty"`
	Status     string                 `json:"status,omitempty"`
	Operations []*clusterCTLOperation `json:"operations"`
	Error      *errorResponse         `json:"error,omitempty"`
}

type clusterCTLOperation struct {
	Operation    string         `json:"operation,omitempty"`
	ResourceName string         `json:"resource_name,omitempty"`
	Status       string         `json:"status,omitempty"`
	Error        *errorResponse `json:"error,omitempty"`
}

type errorResponse struct {
	Type   string                 `json:"type,omitempty"`
	Status int                    `json:"status,omitempty"`
	Title  string                 `json:"title,omitempty"`
	Detail map[string]interface{} `json:"detail,omitempty"`
}

func resourceAKSCluster() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAKSClusterCreate,
		ReadContext:   resourceAKSClusterRead,
		UpdateContext: resourceAKSClusterUpdate,
		DeleteContext: resourceAKSClusterDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(100 * time.Minute),
			Update: schema.DefaultTimeout(130 * time.Minute),
			Delete: schema.DefaultTimeout(70 * time.Minute),
		},

		SchemaVersion: 2,
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
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    fromV1.Resource().CoreConfigSchema().ImpliedType(),
				Upgrade: fromV1.Migrate,
				Version: fromV1.Version,
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
		"proxy_config": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "Proxy configuration for Rafay system components (bootstrap, agents). Use this if the infrastructure uses an outbound proxy.",
			Elem: &schema.Resource{
				Schema: clusterAKSClusterSpecProxyConfig(),
			},
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

func clusterAKSClusterSpecProxyConfig() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"http_proxy": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "HTTP proxy URL for outbound traffic.",
		},
		"https_proxy": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "HTTPS proxy URL for outbound traffic.",
		},
		"no_proxy": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Comma-separated list of hosts or CIDRs that should bypass the proxy (e.g., 10.0.0.0/16,localhost,127.0.0.1,.svc,.cluster.local).",
		},
		"enabled": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "Whether proxy is enabled for Rafay system components.",
		},
		"proxy_auth": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Proxy authentication (e.g., user:password).",
		},
		"bootstrap_ca": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "CA certificate for proxy TLS/bootstrap.",
		},
		"allow_insecure_bootstrap": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "Allow insecure bootstrap when using proxy.",
		},
	}
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
		"maintenance_configurations": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "The Aks Auto-Upgrade Channels maintenance configurations",
			Elem: &schema.Resource{
				Schema: clusterAKSMaintenanceConfig(),
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
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Identities associated with the cluster",
			Elem: &schema.Resource{
				Schema: clusterAKSManagedClusterIdentityProfile(),
			},
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
			Description: "Profile of OpenID Connect configuration.",
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
		"power_state": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Cluster Power State",
			Elem: &schema.Resource{
				Schema: clusterAKSManagedClusterPowerState(),
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
		"node_os_upgrade_channel": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Valid values are None, Unmanaged, NodeImage",
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

func clusterAKSManagedClusterIdentityProfile() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"kubelet_identity": {
			Type:        schema.TypeList,
			Required:    true,
			Description: "Kubelet Identity for managed cluster identity profile",
			Elem: &schema.Resource{
				Schema: clusterAKSManagedClusterKubeletIdentity(),
			},
		},
	}
	return s
}

func clusterAKSManagedClusterKubeletIdentity() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"resource_id": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "value must be ARM resource ID in the form: /subscriptions/<subscription-id>/resourceGroups/<resource-group>/providers/Microsoft.ManagedIdentity/userAssignedIdentities/<identity-name>",
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
		"network_plugin_mode": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Network plugin mode used for building the Azure CNI. Valid values are 'overlay'",
		},
		"network_dataplane": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Network dataplane used in the Kubernetes cluster. Valid values are azure, cilium.",
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
			Description: "Whether to enable OIDC Issuer",
			Default:     false,
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

func clusterAKSManagedClusterPowerState() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"code": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Whether the cluster is running or stopped",
			// ValidateFunc: schema.SchemaValidateFunc(func(v interface{}, k string) (ws []string, errors []error) {

			// },
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
				Schema: clusterAKSManagedClusterWorkloadIdentity(),
			},
		},
	}
	return s
}

func clusterAKSManagedClusterWorkloadIdentity() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"enabled": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "Whether to enable workload identity",
			Default:     false,
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
			Optional:    true,
			Description: "The name of the Azure Container Registry resource.",
		},
		"registries": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "The list of Azure Container Registry Profiles",
			Elem: &schema.Resource{
				Schema: clusterAKSManagedClusterAdditionalMetadataACRProfiles(),
			},
		},
	}
	return s
}

func clusterAKSManagedClusterAdditionalMetadataACRProfiles() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"acr_name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The name of the Azure Container Registry resource.",
		},
		"resource_group_name": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The location of the Azure Container Registry resource.",
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
			Optional:    true,
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
		"creation_data": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "The creation data for the node pool VMSS when using a custom image.",
			Elem: &schema.Resource{
				Schema: clusterAKSNodePoolCreationData(),
			},
		},
	}
	return s
}

func clusterAKSNodePoolCreationData() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"source_resource_id": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The resource ID of the custom image to be used for the node pool VMSS.",
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

func clusterAKSMaintenanceConfig() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"api_version": {
			Description: "",
			Required:    true,
			Type:        schema.TypeString,
		},
		"name": {
			Description: "",
			Required:    true,
			Type:        schema.TypeString,
		},
		"type": {
			Description: "",
			Optional:    true,
			Type:        schema.TypeString,
		},
		"properties": {
			Description: "",
			Required:    true,
			Type:        schema.TypeList,
			Elem: &schema.Resource{
				Schema: clusterAKSMaintenanceConfigProperties(),
			},
		},
	}
	return s
}

func clusterAKSMaintenanceConfigProperties() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"maintenance_window": {
			Description: "",
			Optional:    true,
			Type:        schema.TypeList,
			Elem: &schema.Resource{
				Schema: clusterAKSMaintenanceWindow(),
			},
		},
		"not_allowed_time": {
			Description: "",
			Optional:    true,
			Type:        schema.TypeList,
			Elem: &schema.Resource{
				Schema: clusterAKSMaintenanceNotAllowedTime(),
			},
		},
		"time_in_week": {
			Description: "",
			Optional:    true,
			Type:        schema.TypeList,
			Elem: &schema.Resource{
				Schema: clusterAKSMaintenanceTimeInWeek(),
			},
		},
	}
	return s
}

func clusterAKSMaintenanceWindow() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"duration_hours": {
			Description: "",
			Optional:    true,
			Type:        schema.TypeInt,
		},
		"not_allowed_dates": {
			Description: "",
			Optional:    true,
			Type:        schema.TypeList,
			Elem: &schema.Resource{
				Schema: clusterAKSMaintenanceNotAllowedTime(),
			},
		},
		"start_date": {
			Description: "",
			Optional:    true,
			Type:        schema.TypeString,
		},
		"start_time": {
			Description: "",
			Optional:    true,
			Type:        schema.TypeString,
		},
		"utc_offset": {
			Description: "",
			Optional:    true,
			Type:        schema.TypeString,
		},
		"schedule": {
			Description: "",
			Optional:    true,
			Type:        schema.TypeList,
			Elem: &schema.Resource{
				Schema: clusterAKSMaintenanceConfigSchedule(),
			},
		},
	}
	return s
}

func clusterAKSMaintenanceConfigSchedule() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"absolute_monthly": {
			Description: "",
			Optional:    true,
			Type:        schema.TypeList,
			Elem: &schema.Resource{
				Schema: clusterAKSMaintenanceScheduleAbsoluteMonthly(),
			},
		},
		"daily": {
			Description: "",
			Optional:    true,
			Type:        schema.TypeList,
			Elem: &schema.Resource{
				Schema: clusterAKSMaintenanceScheduleDaily(),
			},
		},
		"relative_monthly": {
			Description: "",
			Optional:    true,
			Type:        schema.TypeList,
			Elem: &schema.Resource{
				Schema: clusterAKSMaintenanceScheduleRelativeMonthly(),
			},
		},
		"weekly": {
			Description: "",
			Optional:    true,
			Type:        schema.TypeList,
			Elem: &schema.Resource{
				Schema: clusterAKSMaintenanceScheduleWeekly(),
			},
		},
	}
	return s
}

func clusterAKSMaintenanceScheduleAbsoluteMonthly() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"day_of_month": {
			Description: "",
			Optional:    true,
			Type:        schema.TypeInt,
		},
		"interval_months": {
			Description: "",
			Optional:    true,
			Type:        schema.TypeInt,
		},
	}
	return s
}
func clusterAKSMaintenanceScheduleDaily() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"interval_days": {
			Description: "",
			Optional:    true,
			Type:        schema.TypeInt,
		},
	}
	return s
}
func clusterAKSMaintenanceScheduleRelativeMonthly() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"day_of_week": {
			Description: "",
			Optional:    true,
			Type:        schema.TypeString,
		},
		"interval_months": {
			Description: "",
			Optional:    true,
			Type:        schema.TypeInt,
		},
		"week_index": {
			Description: "",
			Optional:    true,
			Type:        schema.TypeString,
		},
	}
	return s
}
func clusterAKSMaintenanceScheduleWeekly() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"day_of_week": {
			Description: "",
			Optional:    true,
			Type:        schema.TypeString,
		},
		"interval_weeks": {
			Description: "",
			Optional:    true,
			Type:        schema.TypeInt,
		},
	}
	return s
}

func clusterAKSMaintenanceNotAllowedTime() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"end": {
			Description: "",
			Optional:    true,
			Type:        schema.TypeString,
		},
		"start": {
			Description: "",
			Optional:    true,
			Type:        schema.TypeString,
		},
	}
	return s
}

func clusterAKSMaintenanceTimeInWeek() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"day": {
			Description: "",
			Optional:    true,
			Type:        schema.TypeString,
		},
		"hour_slots": {
			Description: "",
			Elem:        &schema.Schema{Type: schema.TypeInt},
			Optional:    true,
			Type:        schema.TypeList,
		},
	}
	return s
}

func expandAKSClusterMetadata(p []interface{}) *AKSClusterMetadata {
	obj := &AKSClusterMetadata{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["name"].(string); ok && len(v) > 0 {
		obj.Name = v
	}

	if v, ok := in["project"].(string); ok && len(v) > 0 {
		obj.Project = v
	}

	if v, ok := in["labels"].(map[string]interface{}); ok {
		obj.Labels = toMapString(v)
	}

	return obj
}

func expandAKSClusterSpec(p []interface{}, rawConfig cty.Value) *AKSClusterSpec {
	obj := &AKSClusterSpec{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})
	rawConfig = rawConfig.AsValueSlice()[0]

	if v, ok := in["type"].(string); ok && len(v) > 0 {
		obj.Type = v
	}

	if v, ok := in["blueprint"].(string); ok && len(v) > 0 {
		obj.Blueprint = v
	}

	if v, ok := in["blueprintversion"].(string); ok && len(v) > 0 {
		obj.BlueprintVersion = v
	}

	if v, ok := in["cloudprovider"].(string); ok && len(v) > 0 {
		obj.CloudProvider = v
	}

	if v, ok := in["cluster_config"].([]interface{}); ok && len(v) > 0 {
		obj.AKSClusterConfig = expandAKSClusterConfig(v, rawConfig.GetAttr("cluster_config"))
	}

	if v, ok := in["sharing"].([]interface{}); ok && len(v) > 0 {
		obj.Sharing = expandV1ClusterSharing(v)
	}

	if v, ok := in["system_components_placement"].([]interface{}); ok && len(v) > 0 {
		obj.SystemComponentsPlacement = expandSystemComponentsPlacement(v)
	}

	if v, ok := in["proxy_config"].([]interface{}); ok && len(v) > 0 {
		obj.ProxyConfig = expandAKSClusterSpecProxyConfig(v)
	}

	return obj
}

func expandAKSClusterSpecProxyConfig(p []interface{}) *AKSClusterSpecProxyConfig {
	obj := &AKSClusterSpecProxyConfig{}
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
	if v, ok := in["no_proxy"].(string); ok && len(v) > 0 {
		obj.NoProxy = v
	}
	if v, ok := in["enabled"].(bool); ok {
		obj.Enabled = v
	}
	if v, ok := in["proxy_auth"].(string); ok && len(v) > 0 {
		obj.ProxyAuth = v
	}
	if v, ok := in["bootstrap_ca"].(string); ok && len(v) > 0 {
		obj.BootstrapCA = v
	}
	if v, ok := in["allow_insecure_bootstrap"].(bool); ok {
		obj.AllowInsecureBootstrap = v
	}
	if len(obj.HttpProxy) > 0 || len(obj.HttpsProxy) > 0 {
		obj.Enabled = true
	}
	return obj
}

func expandAKSClusterConfig(p []interface{}, rawConfig cty.Value) *AKSClusterConfig {
	obj := &AKSClusterConfig{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})
	rawConfig = rawConfig.AsValueSlice()[0]
	if v, ok := in["apiversion"].(string); ok && len(v) > 0 {
		obj.APIVersion = v
	}

	if v, ok := in["kind"].(string); ok && len(v) > 0 {
		obj.Kind = v
	}

	if v, ok := in["metadata"].([]interface{}); ok && len(v) > 0 {
		obj.Metadata = expandAKSClusterConfigMetadata(v)
	}

	if v, ok := in["spec"].([]interface{}); ok && len(v) > 0 {
		obj.Spec = expandAKSClusterConfigSpec(v, rawConfig.GetAttr("spec"))
	}

	return obj
}

func expandAKSClusterConfigMetadata(p []interface{}) *AKSClusterConfigMetadata {
	obj := &AKSClusterConfigMetadata{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["name"].(string); ok && len(v) > 0 {
		obj.Name = v
	}
	return obj
}

func expandAKSClusterConfigSpec(p []interface{}, rawConfig cty.Value) *AKSClusterConfigSpec {
	obj := &AKSClusterConfigSpec{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})
	rawConfig = rawConfig.AsValueSlice()[0]
	if v, ok := in["subscription_id"].(string); ok && len(v) > 0 {
		obj.SubscriptionID = v
	}

	if v, ok := in["resource_group_name"].(string); ok && len(v) > 0 {
		obj.ResourceGroupName = v
	}

	if v, ok := in["managed_cluster"].([]interface{}); ok && len(v) > 0 {
		obj.ManagedCluster = expandAKSConfigManagedCluster(v)
	}

	if v, ok := in["node_pools"].([]interface{}); ok && len(v) > 0 {
		obj.NodePools = expandAKSNodePool(v, rawConfig.GetAttr("node_pools"))
	}

	if v, ok := in["maintenance_configurations"].([]interface{}); ok && len(v) > 0 {
		obj.MaintenanceConfigs = expandAKSMaintenanceConfigs(v)
	}

	return obj
}

func expandAKSConfigManagedCluster(p []interface{}) *AKSManagedCluster {
	obj := &AKSManagedCluster{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["apiversion"].(string); ok && len(v) > 0 {
		obj.APIVersion = v
	}

	if v, ok := in["extended_location"].([]interface{}); ok && len(v) > 0 {
		obj.ExtendedLocation = expandAKSManagedClusterExtendedLocation(v)
	}

	if v, ok := in["identity"].([]interface{}); ok && len(v) > 0 {
		obj.Identity = expandAKSManagedClusterIdentity(v)
	}

	if v, ok := in["location"].(string); ok && len(v) > 0 {
		obj.Location = v
	}

	if v, ok := in["properties"].([]interface{}); ok && len(v) > 0 {
		obj.Properties = expandAKSManagedClusterProperties(v)
	}

	if v, ok := in["sku"].([]interface{}); ok && len(v) > 0 {
		obj.SKU = expandAKSManagedClusterSKU(v)
	}

	if v, ok := in["tags"].(map[string]interface{}); ok {
		obj.Tags = v
	}

	if v, ok := in["type"].(string); ok && len(v) > 0 {
		obj.Type = v
	}

	if v, ok := in["additional_metadata"].([]interface{}); ok && len(v) > 0 {
		obj.AdditionalMetadata = expandAKSManagedClusterAdditionalMetadata(v)
	}

	return obj
}

func expandAKSManagedClusterExtendedLocation(p []interface{}) *AKSClusterExtendedLocation {
	obj := &AKSClusterExtendedLocation{}
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

func expandAKSManagedClusterIdentity(p []interface{}) *AKSManagedClusterIdentity {
	obj := &AKSManagedClusterIdentity{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["type"].(string); ok && len(v) > 0 {
		obj.Type = v
	}

	if v, ok := in["user_assigned_identities"].(map[string]interface{}); ok {
		//obj.UserAssignedIdentities = toMapString(v)
		obj.UserAssignedIdentities = toMapEmptyObject(v)
	}
	return obj
}

func expandAKSManagedClusterProperties(p []interface{}) *AKSManagedClusterProperties {
	obj := &AKSManagedClusterProperties{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["aad_profile"].([]interface{}); ok && len(v) > 0 {
		obj.AzureADProfile = expandAKSManagedClusterAzureADProfile(v)
	}

	/*
		if v, ok := in["addon_profiles"].(map[string]interface{}); ok {
			obj.AddonProfiles = toMapString(v)
		}*/

	if v, ok := in["addon_profiles"].([]interface{}); ok && len(v) > 0 {
		obj.AddonProfiles = expandAddonProfiles(v)
	}

	if v, ok := in["api_server_access_profile"].([]interface{}); ok && len(v) > 0 {
		obj.APIServerAccessProfile = expandAKSManagedClusterAPIServerAccessProfile(v)
	}

	if v, ok := in["auto_scaler_profile"].([]interface{}); ok && len(v) > 0 {
		obj.AutoScalerProfile = expandAKSManagedClusterAutoScalerProfile(v)
	}

	if v, ok := in["auto_upgrade_profile"].([]interface{}); ok && len(v) > 0 {
		obj.AutoUpgradeProfile = expandAKSManagedClusterAutoUpgradeProfile(v)
	}

	if v, ok := in["disable_local_accounts"].(bool); ok {
		obj.DisableLocalAccounts = &v
	}

	if v, ok := in["disk_encryption_set_id"].(string); ok {
		obj.DiskEncryptionSetID = v
	}

	if v, ok := in["dns_prefix"].(string); ok {
		obj.DNSPrefix = v
	}

	if v, ok := in["enable_pod_security_policy"].(bool); ok {
		obj.EnablePodSecurityPolicy = &v
	}

	if v, ok := in["enable_rbac"].(bool); ok {
		obj.EnableRBAC = &v
	}

	if v, ok := in["fqdn_subdomain"].(string); ok {
		obj.FQDNSubdomain = v
	}

	if v, ok := in["http_proxy_config"].([]interface{}); ok && len(v) > 0 {
		obj.HTTPProxyConfig = expandAKSManagedClusterHTTPProxyConfig(v)
	}

	if v, ok := in["identity_profile"].([]interface{}); ok && len(v) > 0 {
		obj.IdentityProfile = expandAKSManagedClusterIdentityProfile(v)
	}

	if v, ok := in["kubernetes_version"].(string); ok {
		obj.KubernetesVersion = v
	}

	if v, ok := in["linux_profile"].([]interface{}); ok && len(v) > 0 {
		obj.LinuxProfile = expandAKSManagedClusterLinuxProfile(v)
	}

	if v, ok := in["network_profile"].([]interface{}); ok && len(v) > 0 {
		obj.NetworkProfile = expandAKSManagedClusterNetworkProfile(v)
	}

	if v, ok := in["node_resource_group"].(string); ok {
		obj.NodeResourceGroup = v
	}

	if v, ok := in["oidc_issuer_profile"].([]interface{}); ok && len(v) > 0 {
		obj.OidcIssuerProfile = expandAKSManagedClusterOidcIssuerProfile(v)
	}

	if v, ok := in["pod_identity_profile"].([]interface{}); ok && len(v) > 0 {
		obj.PodIdentityProfile = expandAKSManagedClusterPodIdentityProfile(v)
	}

	if v, ok := in["private_link_resources"].([]interface{}); ok && len(v) > 0 {
		obj.PrivateLinkResources = expandAKSManagedClusterPrivateLinkResources(v)
	}

	if v, ok := in["power_state"].([]interface{}); ok && len(v) > 0 {
		obj.PowerState = expandAKSManagedClusterPowerState(v)
	}

	if v, ok := in["security_profile"].([]interface{}); ok && len(v) > 0 {
		obj.SecurityProfile = expandAKSManagedClusterSecurityProfile(v)
	}

	if v, ok := in["service_principal_profile"].([]interface{}); ok && len(v) > 0 {
		obj.ServicePrincipalProfile = expandAKSManagedClusterServicePrincipalProfile(v)
	}

	if v, ok := in["windows_profile"].([]interface{}); ok && len(v) > 0 {
		obj.WindowsProfile = expandAKSManagedClusterWindowsProfile(v)
	}

	return obj
}

func expandAddonProfiles(p []interface{}) *AddonProfiles {
	obj := &AddonProfiles{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["http_application_routing"].([]interface{}); ok && len(v) > 0 {
		obj.HttpApplicationRouting = expandAKSManagedClusterAddonProfile(v)
	}
	if v, ok := in["azure_policy"].([]interface{}); ok && len(v) > 0 {
		obj.AzurePolicy = expandAKSManagedClusterAddonProfile(v)
	}
	if v, ok := in["oms_agent"].([]interface{}); ok && len(v) > 0 {
		obj.OmsAgent = expandAKSManagedClusterAddonOmsAgentProfile(v)
	}
	if v, ok := in["azure_keyvault_secrets_provider"].([]interface{}); ok && len(v) > 0 {
		obj.AzureKeyvaultSecretsProvider = expandAKSManagedClusterAddonAzureKeyvaultSecretsProviderProfile(v)
	}
	if v, ok := in["ingress_application_gateway"].([]interface{}); ok && len(v) > 0 {
		obj.IngressApplicationGateway = expandAKSManagedClusterAddonIngressApplicationGatewayProfile(v)
	}

	return obj
}

func expandAKSManagedClusterAddonProfile(p []interface{}) *AKSManagedClusterAddonProfile {
	obj := &AKSManagedClusterAddonProfile{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["enabled"].(bool); ok {
		obj.Enabled = &v
	}
	//convert string input into json object (map[string]interfgace{})
	if v, ok := in["config"].(string); ok && len(v) > 0 {
		var policyDoc map[string]interface{}
		var json2 = jsoniter.ConfigCompatibleWithStandardLibrary
		json2.Unmarshal([]byte(v), &policyDoc)
		obj.Config = policyDoc
		log.Println("addon profile config expanded correct")
	}

	return obj
}

func expandAKSManagedClusterAddonOmsAgentProfile(p []interface{}) *OmsAgentProfile {
	obj := &OmsAgentProfile{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["enabled"].(bool); ok {
		obj.Enabled = &v
	}
	if v, ok := in["config"].([]interface{}); ok && len(v) > 0 {
		obj.Config = expandAKSManagedClusterAddonOmsAgentConfigProfile(v)
	}

	return obj
}

func expandAKSManagedClusterAddonAzureKeyvaultSecretsProviderProfile(p []interface{}) *AzureKeyvaultSecretsProviderProfile {
	obj := &AzureKeyvaultSecretsProviderProfile{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["enabled"].(bool); ok {
		obj.Enabled = &v
	}
	if v, ok := in["config"].([]interface{}); ok && len(v) > 0 {
		obj.Config = expandAKSManagedClusterAddonAzureKeyvaultSecretsProviderConfigProfile(v)
	}

	return obj
}

func expandAKSManagedClusterAddonOmsAgentConfigProfile(p []interface{}) *OmsAgentConfig {
	obj := &OmsAgentConfig{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["log_analytics_workspace_resource_id"].(string); ok && len(v) > 0 {
		obj.LogAnalyticsWorkspaceResourceID = v
	}

	return obj
}

func expandAKSManagedClusterAddonAzureKeyvaultSecretsProviderConfigProfile(p []interface{}) *AzureKeyvaultSecretsProviderProfileConfig {
	obj := &AzureKeyvaultSecretsProviderProfileConfig{}
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

func expandAKSManagedClusterAddonIngressApplicationGatewayProfile(p []interface{}) *IngressApplicationGatewayAddonProfile {
	obj := &IngressApplicationGatewayAddonProfile{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["enabled"].(bool); ok {
		obj.Enabled = &v
	}
	if v, ok := in["config"].([]interface{}); ok && len(v) > 0 {
		obj.Config = expandAKSManagedClusterAddonIngressApplicationGatewayConfig(v)
	}

	return obj
}

func expandAKSManagedClusterAddonIngressApplicationGatewayConfig(p []interface{}) *IngressApplicationGatewayAddonConfig {
	obj := &IngressApplicationGatewayAddonConfig{}
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

func expandAKSManagedClusterAzureADProfile(p []interface{}) *AKSManagedClusterAzureADProfile {
	obj := &AKSManagedClusterAzureADProfile{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["admin_group_object_ids"].([]interface{}); ok && len(v) > 0 {
		obj.AdminGroupObjectIDs = toArrayString(v)
	}

	if v, ok := in["client_app_id"].(string); ok && len(v) > 0 {
		obj.ClientAppId = v
	}

	if v, ok := in["enable_azure_rbac"].(bool); ok {
		obj.EnableAzureRbac = &v
	}

	if v, ok := in["managed"].(bool); ok {
		obj.Managed = &v
	}

	if v, ok := in["server_app_id"].(string); ok && len(v) > 0 {
		obj.ServerAppId = v
	}

	if v, ok := in["server_app_secret"].(string); ok && len(v) > 0 {
		obj.ServerAppSecret = v
	}

	if v, ok := in["tenant_id"].(string); ok && len(v) > 0 {
		obj.TenantId = v
	}

	return obj
}

func expandAKSManagedClusterAPIServerAccessProfile(p []interface{}) *AKSManagedClusterAPIServerAccessProfile {
	obj := &AKSManagedClusterAPIServerAccessProfile{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["authorized_ipr_ranges"].([]interface{}); ok && len(v) > 0 {
		obj.AuthorizedIPRanges = toArrayString(v)
	}

	if v, ok := in["enable_private_cluster"].(bool); ok {
		obj.EnablePrivateCluster = &v
	}

	if v, ok := in["enable_private_cluster_public_fqdn"].(bool); ok {
		obj.EnablePrivateClusterPublicFQDN = &v
	}

	if v, ok := in["private_dns_zone"].(string); ok && len(v) > 0 {
		obj.PrivateDnsZone = v
	}
	return obj
}

func expandAKSManagedClusterAutoScalerProfile(p []interface{}) *AKSManagedClusterAutoScalerProfile {
	obj := &AKSManagedClusterAutoScalerProfile{}
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

func expandAKSManagedClusterAutoUpgradeProfile(p []interface{}) *AKSManagedClusterAutoUpgradeProfile {
	obj := &AKSManagedClusterAutoUpgradeProfile{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["upgrade_channel"].(string); ok && len(v) > 0 {
		obj.UpgradeChannel = v
	}
	if v, ok := in["node_os_upgrade_channel"].(string); ok && len(v) > 0 {
		obj.NodeOsUpgradeChannel = v
	}
	return obj
}

func expandAKSManagedClusterHTTPProxyConfig(p []interface{}) *AKSManagedClusterHTTPProxyConfig {
	obj := &AKSManagedClusterHTTPProxyConfig{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["http_proxy"].(string); ok && len(v) > 0 {
		obj.HTTPProxy = v
	}

	if v, ok := in["https_proxy"].(string); ok && len(v) > 0 {
		obj.HTTPSProxy = v
	}

	if v, ok := in["no_proxy"].([]interface{}); ok && len(v) > 0 {
		obj.NoProxy = toArrayString(v)
	}

	if v, ok := in["trusted_ca"].(string); ok && len(v) > 0 {
		obj.TrustedCA = v
	}

	return obj
}

func expandAKSManagedClusterIdentityProfile(p []interface{}) *AKSManagedClusterIdentityProfile {
	obj := &AKSManagedClusterIdentityProfile{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["kubelet_identity"].([]interface{}); ok && len(v) > 0 {
		obj.KubeletIdentity = expandAKSManagedClusterIdentityProfileKubeletIdentity(v)
	}

	return obj
}

func expandAKSManagedClusterIdentityProfileKubeletIdentity(p []interface{}) *AKSManagedClusterKubeletIdentity {
	obj := &AKSManagedClusterKubeletIdentity{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["resource_id"].(string); ok && len(v) > 0 {
		obj.ResourceId = v
	}

	return obj
}

func expandAKSManagedClusterLinuxProfile(p []interface{}) *AKSManagedClusterLinuxProfile {
	obj := &AKSManagedClusterLinuxProfile{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["admin_username"].(string); ok && len(v) > 0 {
		obj.AdminUsername = v
	}

	if v, ok := in["ssh"].([]interface{}); ok && len(v) > 0 {
		obj.SSH = expandAKSManagedClusterSSHConfig(v)
	}

	if v, ok := in["no_proxy"].([]interface{}); ok && len(v) > 0 {
		obj.NoProxy = toArrayString(v)
	}

	if v, ok := in["trusted_ca"].(string); ok && len(v) > 0 {
		obj.TrustedCa = v
	}
	return obj
}

func expandAKSManagedClusterSSHConfig(p []interface{}) *AKSManagedClusterSSHConfig {
	obj := &AKSManagedClusterSSHConfig{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["public_keys"].([]interface{}); ok && len(v) > 0 {
		obj.PublicKeys = expandAKSManagedClusterLPSSHKeyData(v)
	}

	return obj

}

func expandAKSManagedClusterLPSSHKeyData(p []interface{}) []*AKSManagedClusterSSHKeyData {
	if len(p) == 0 || p[0] == nil {
		return []*AKSManagedClusterSSHKeyData{}
	}
	out := make([]*AKSManagedClusterSSHKeyData, len(p))

	for i := range p {
		obj := AKSManagedClusterSSHKeyData{}
		in := p[i].(map[string]interface{})

		if v, ok := in["key_data"].(string); ok {
			obj.KeyData = v
		}
		out[i] = &obj
	}
	return out
}

func expandAKSManagedClusterNetworkProfile(p []interface{}) *AKSManagedClusterNetworkProfile {
	obj := &AKSManagedClusterNetworkProfile{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["dns_service_ip"].(string); ok && len(v) > 0 {
		obj.DNSServiceIP = v
	}

	if v, ok := in["docker_bridge_cidr"].(string); ok && len(v) > 0 {
		obj.DockerBridgeCidr = v
	}

	if v, ok := in["load_balancer_profile"].([]interface{}); ok && len(v) > 0 {
		obj.LoadBalancerProfile = expandAKSManagedClusterNPLoadBalancerProfile(v)
	}

	if v, ok := in["load_balancer_sku"].(string); ok && len(v) > 0 {
		obj.LoadBalancerSKU = v
	}

	if v, ok := in["network_mode"].(string); ok && len(v) > 0 {
		obj.NetworkMode = v
	}

	if v, ok := in["network_plugin"].(string); ok && len(v) > 0 {
		obj.NetworkPlugin = v
	}

	if v, ok := in["network_plugin_mode"].(string); ok && len(v) > 0 {
		obj.NetworkPluginMode = v
	}

	if v, ok := in["network_policy"].(string); ok && len(v) > 0 {
		obj.NetworkPolicy = v
	}

	if v, ok := in["network_dataplane"].(string); ok && len(v) > 0 {
		obj.NetworkDataplane = v
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

func expandAKSManagedClusterNPLoadBalancerProfile(p []interface{}) *AKSManagedClusterNPLoadBalancerProfile {
	obj := &AKSManagedClusterNPLoadBalancerProfile{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["allocated_outbound_ports"].(int); ok && v > 0 {
		obj.AllocatedOutboundPorts = &v
	}

	if v, ok := in["effective_outbound_ips"].([]interface{}); ok && len(v) > 0 {
		obj.EffectiveOutboundIPs = expandAKSManagedClusterNPEffectiveOutboundIPs(v)
	}

	if v, ok := in["idle_timeout_in_minutes"].(int); ok && v > 0 {
		obj.IdleTimeoutInMinutes = &v
	}

	if v, ok := in["managed_outbound_ips"].([]interface{}); ok && len(v) > 0 {
		obj.ManagedOutboundIPs = expandAKSManagedClusterNPManagedOutboundIPs(v)
	}

	if v, ok := in["outbound_ip_prefixes"].([]interface{}); ok && len(v) > 0 {
		obj.OutboundIPPrefixes = expandAKSManagedClusterNPOutboundIPPrefixes(v)
	}

	if v, ok := in["outbound_ips"].([]interface{}); ok && len(v) > 0 {
		obj.OutboundIPs = expandAKSManagedClusterNPOutboundIPs(v)
	}

	return obj
}

func expandAKSManagedClusterNPEffectiveOutboundIPs(p []interface{}) []*AKSManagedClusterNPEffectiveOutboundIPs {
	if len(p) == 0 || p[0] == nil {
		return []*AKSManagedClusterNPEffectiveOutboundIPs{}
	}
	out := make([]*AKSManagedClusterNPEffectiveOutboundIPs, len(p))

	for i := range p {
		obj := AKSManagedClusterNPEffectiveOutboundIPs{}
		in := p[i].(map[string]interface{})

		if v, ok := in["id"].(string); ok {
			obj.ID = v
		}
		out[i] = &obj
	}
	return out
}

func expandAKSManagedClusterNPManagedOutboundIPs(p []interface{}) *AKSManagedClusterNPManagedOutboundIPs {
	obj := &AKSManagedClusterNPManagedOutboundIPs{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["count"].(int); ok && v > 0 {
		obj.Count = &v
	}
	return obj
}

func expandAKSManagedClusterNPOutboundIPPrefixes(p []interface{}) *AKSManagedClusterNPOutboundIPPrefixes {
	obj := &AKSManagedClusterNPOutboundIPPrefixes{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["public_ip_prefixes"].([]interface{}); ok && len(v) > 0 {
		obj.PublicIPPrefixes = expandAKSManagedClusterNPManagedOutboundIPsPublicIpPrefixes(v)
	}
	return obj
}

func expandAKSManagedClusterNPManagedOutboundIPsPublicIpPrefixes(p []interface{}) []*AKSManagedClusterNPManagedOutboundIPsPublicIpPrefixes {
	if len(p) == 0 || p[0] == nil {
		return []*AKSManagedClusterNPManagedOutboundIPsPublicIpPrefixes{}
	}
	out := make([]*AKSManagedClusterNPManagedOutboundIPsPublicIpPrefixes, len(p))

	for i := range p {
		obj := AKSManagedClusterNPManagedOutboundIPsPublicIpPrefixes{}
		in := p[i].(map[string]interface{})

		if v, ok := in["id"].(string); ok {
			obj.ID = v
		}
		out[i] = &obj
	}
	return out
}

func expandAKSManagedClusterNPOutboundIPs(p []interface{}) *AKSManagedClusterNPOutboundIPs {
	obj := &AKSManagedClusterNPOutboundIPs{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["public_ips"].([]interface{}); ok && len(v) > 0 {
		obj.PublicIPs = expandAKSManagedClusterNPOutboundIPsPublicIps(v)
	}
	return obj
}

func expandAKSManagedClusterNPOutboundIPsPublicIps(p []interface{}) []*AKSManagedClusterNPOutboundIPsPublicIps {
	if len(p) == 0 || p[0] == nil {
		return []*AKSManagedClusterNPOutboundIPsPublicIps{}
	}
	out := make([]*AKSManagedClusterNPOutboundIPsPublicIps, len(p))

	for i := range p {
		obj := AKSManagedClusterNPOutboundIPsPublicIps{}
		in := p[i].(map[string]interface{})

		if v, ok := in["id"].(string); ok {
			obj.ID = v
		}
		out[i] = &obj
	}
	return out
}

func expandAKSManagedClusterOidcIssuerProfile(p []interface{}) *AKSManagedClusterOidcIssuerProfile {
	obj := &AKSManagedClusterOidcIssuerProfile{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["enabled"].(bool); ok {
		obj.Enabled = &v
	}

	return obj
}

func expandAKSManagedClusterPodIdentityProfile(p []interface{}) *AKSManagedClusterPodIdentityProfile {
	obj := &AKSManagedClusterPodIdentityProfile{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["allow_network_plugin_kubenet"].(bool); ok {
		obj.AllowNetworkPluginKubenet = &v
	}

	if v, ok := in["enabled"].(bool); ok {
		obj.Enabled = &v
	}

	if v, ok := in["user_assigned_identities"].([]interface{}); ok && len(v) > 0 {
		obj.UserAssignedIdentities = expandAKSManagedClusterPIPUserAssignedIdentities(v)
	}

	if v, ok := in["user_assigned_identity_exceptions"].([]interface{}); ok && len(v) > 0 {
		obj.UserAssignedIdentityExceptions = expandAKSManagedClusterPIPUserAssignedIdentityExceptions(v)
	}

	return obj
}

func expandAKSManagedClusterPIPUserAssignedIdentities(p []interface{}) []*AKSManagedClusterPIPUserAssignedIdentities {
	out := make([]*AKSManagedClusterPIPUserAssignedIdentities, len(p))
	if len(p) == 0 || p[0] == nil {
		return out
	}
	for i := range p {
		obj := &AKSManagedClusterPIPUserAssignedIdentities{}
		in := p[i].(map[string]interface{})

		if v, ok := in["binding_selector"].(string); ok && len(v) > 0 {
			obj.BindingSelector = v
		}

		if v, ok := in["identity"].([]interface{}); ok && len(v) > 0 {
			obj.Identity = expandAKSManagedClusterUAIIdentity(v)
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

func expandAKSManagedClusterUAIIdentity(p []interface{}) *AKSManagedClusterUAIIdentity {
	obj := &AKSManagedClusterUAIIdentity{}
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

func expandAKSManagedClusterPIPUserAssignedIdentityExceptions(p []interface{}) []*AKSManagedClusterPIPUserAssignedIdentityExceptions {
	out := make([]*AKSManagedClusterPIPUserAssignedIdentityExceptions, len(p))
	if len(p) == 0 || p[0] == nil {
		return out
	}
	for i := range p {
		obj := &AKSManagedClusterPIPUserAssignedIdentityExceptions{}
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

func expandAKSManagedClusterPrivateLinkResources(p []interface{}) *AKSManagedClusterPrivateLinkResources {
	obj := &AKSManagedClusterPrivateLinkResources{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["group_id"].(string); ok && len(v) > 0 {
		obj.GroupId = v
	}

	if v, ok := in["id"].(string); ok && len(v) > 0 {
		obj.ID = v
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

	return obj
}

func expandAKSManagedClusterPowerState(p []interface{}) *AKSManagedClusterPowerState {
	obj := &AKSManagedClusterPowerState{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["code"].(string); ok && len(v) > 0 {
		obj.Code = v
	}

	return obj
}

func expandAKSManagedClusterSecurityProfile(p []interface{}) *AKSManagedClusterSecurityProfile {
	obj := &AKSManagedClusterSecurityProfile{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["workload_identity"].([]interface{}); ok && len(v) > 0 {
		obj.WorkloadIdentity = expandAKSManagedClusterWorkloadIdentity(v)
	}

	return obj
}

func expandAKSManagedClusterWorkloadIdentity(p []interface{}) *AKSManagedClusterWorkloadIdentity {
	obj := &AKSManagedClusterWorkloadIdentity{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["enabled"].(bool); ok {
		obj.Enabled = &v
	}

	return obj
}

func expandAKSManagedClusterServicePrincipalProfile(p []interface{}) *AKSManagedClusterServicePrincipalProfile {
	obj := &AKSManagedClusterServicePrincipalProfile{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["client_id"].(string); ok && len(v) > 0 {
		obj.ClientID = v
	}

	if v, ok := in["secret"].(string); ok && len(v) > 0 {
		obj.Secret = v
	}

	return obj
}

func expandAKSManagedClusterWindowsProfile(p []interface{}) *AKSManagedClusterWindowsProfile {
	obj := &AKSManagedClusterWindowsProfile{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["admin_username"].(string); ok && len(v) > 0 {
		obj.AdminUsername = v
	}

	if v, ok := in["enable_csi_proxy"].(bool); ok {
		obj.EnableCSIProxy = &v
	}

	if v, ok := in["license_type"].(string); ok && len(v) > 0 {
		obj.LicenseType = v
	}
	return obj
}

func expandAKSManagedClusterSKU(p []interface{}) *AKSManagedClusterSKU {
	obj := &AKSManagedClusterSKU{}
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

func expandAKSManagedClusterAdditionalMetadata(p []interface{}) *AKSManagedClusterAdditionalMetadata {
	obj := &AKSManagedClusterAdditionalMetadata{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["acr_profile"].([]interface{}); ok && len(v) > 0 {
		obj.ACRProfile = expandAKSManagedClusterAdditionalMetadataACRProfile(v)
	}

	if v, ok := in["oms_workspace_location"].(string); ok && len(v) > 0 {
		obj.OmsWorkspaceLocation = v
	}

	return obj
}

func expandAKSManagedClusterAdditionalMetadataACRProfile(p []interface{}) *AKSManagedClusterAdditionalMetadataACRProfile {
	obj := &AKSManagedClusterAdditionalMetadataACRProfile{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["resource_group_name"].(string); ok && len(v) > 0 {
		obj.ResourceGroupName = v
	}

	if v, ok := in["acr_name"].(string); ok && len(v) > 0 {
		obj.ACRName = v
	}
	if v, ok := in["registries"].([]interface{}); ok && len(v) > 0 {
		obj.Registries = expandAKSManagedClusterAdditionalMetadataACRProfiles(v)
	}

	return obj
}

func expandAKSManagedClusterAdditionalMetadataACRProfiles(p []interface{}) []*AksRegistry {
	if len(p) == 0 || p[0] == nil {
		return []*AksRegistry{}
	}
	out := make([]*AksRegistry, len(p))

	for i := range p {
		obj := AksRegistry{}
		in := p[i].(map[string]interface{})

		if v, ok := in["acr_name"].(string); ok && len(v) > 0 {
			obj.ACRName = v
		}

		if v, ok := in["resource_group_name"].(string); ok && len(v) > 0 {
			obj.ResourceGroupName = v
		}
		out[i] = &obj
	}
	return out
}

func expandAKSMaintenanceConfigs(p []interface{}) []*AKSMaintenanceConfig {
	if len(p) == 0 || p[0] == nil {
		return []*AKSMaintenanceConfig{}
	}

	out := make([]*AKSMaintenanceConfig, len(p))
	for i := range p {
		obj := AKSMaintenanceConfig{}
		in := p[i].(map[string]interface{})

		if v, ok := in["api_version"].(string); ok && len(v) > 0 {
			obj.ApiVersion = v
		}

		if v, ok := in["name"].(string); ok && len(v) > 0 {
			obj.Name = v
		}

		if v, ok := in["type"].(string); ok && len(v) > 0 {
			obj.Type = v
		}

		if v, ok := in["properties"].([]interface{}); ok && len(v) > 0 {
			obj.Properties = expandAKSMaintenanceConfigProperties(v)
		}
		out[i] = &obj
	}
	return out
}

func expandAKSMaintenanceConfigProperties(p []interface{}) *AKSMaintenanceConfigProperties {
	obj := &AKSMaintenanceConfigProperties{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["maintenance_window"].([]interface{}); ok && len(v) > 0 {
		obj.MaintenanceWindow = expandAKSMCMaintenanceWindow(v)
	}
	if v, ok := in["not_allowed_time"].([]interface{}); ok && len(v) > 0 {
		obj.NotAllowedTime = expandAKSMCTimeSpan(v)
	}
	if v, ok := in["time_in_week"].([]interface{}); ok && len(v) > 0 {
		obj.TimeInWeek = expandAKSMCTimeInWeek(v)
	}
	return obj
}

func expandAKSMCTimeSpan(p []interface{}) []*AKSMaintenanceTimeSpan {
	if len(p) == 0 || p[0] == nil {
		return []*AKSMaintenanceTimeSpan{}
	}

	out := make([]*AKSMaintenanceTimeSpan, len(p))
	for i := range p {
		obj := AKSMaintenanceTimeSpan{}
		in := p[i].(map[string]interface{})
		if v, ok := in["end"].(string); ok && len(v) > 0 {
			obj.End = v
		}
		if v, ok := in["start"].(string); ok && len(v) > 0 {
			obj.Start = v
		}
		out[i] = &obj
	}
	return out
}

func expandAKSMCTimeInWeek(p []interface{}) []*AKSMaintenanceTimeInWeek {
	if len(p) == 0 || p[0] == nil {
		return []*AKSMaintenanceTimeInWeek{}
	}

	out := make([]*AKSMaintenanceTimeInWeek, len(p))
	for i := range p {
		obj := AKSMaintenanceTimeInWeek{}
		in := p[i].(map[string]interface{})
		if v, ok := in["day"].(string); ok && len(v) > 0 {
			obj.Day = v
		}

		if v, ok := in["hour_slots"].([]interface{}); ok && len(v) > 0 {
			obj.HourSlots = toArrayInt(v)
		}
		out[i] = &obj
	}
	return out
}

func expandAKSMCMaintenanceWindow(p []interface{}) *AKSMaintenanceWindow {
	obj := &AKSMaintenanceWindow{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["duration_hours"].(int); ok && v > 0 {
		obj.DurationHours = v
	}

	if v, ok := in["not_allowed_dates"].([]interface{}); ok && len(v) > 0 {
		obj.NotAllowedDates = expandAKSMCTimeSpan(v)
	}

	if v, ok := in["start_date"].(string); ok && len(v) > 0 {
		obj.StartDate = v
	}

	if v, ok := in["start_time"].(string); ok && len(v) > 0 {
		obj.StartTime = v
	}

	if v, ok := in["utc_offset"].(string); ok && len(v) > 0 {
		obj.UtcOffset = v
	}

	if v, ok := in["schedule"].([]interface{}); ok && len(v) > 0 {
		obj.Schedule = expandAKSMCSchedule(v)
	}
	return obj
}

func expandAKSMCSchedule(p []interface{}) *AKSMaintenanceSchedule {
	obj := &AKSMaintenanceSchedule{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["absolute_monthly"].([]interface{}); ok && len(v) > 0 {
		obj.AbsoluteMonthlySchedule = expandAKSMCAbsoluteMonthlySchedule(v)
	}

	if v, ok := in["daily"].([]interface{}); ok && len(v) > 0 {
		obj.DailySchedule = expandAKSMCDailySchedule(v)
	}

	if v, ok := in["relative_monthly"].([]interface{}); ok && len(v) > 0 {
		obj.RelativeMonthlySchedule = expandAKSMCRelativeMonthlySchedule(v)
	}

	if v, ok := in["weekly"].([]interface{}); ok && len(v) > 0 {
		obj.WeeklySchedule = expandAKSMCWeeklySchedule(v)
	}
	return obj
}

func expandAKSMCWeeklySchedule(p []interface{}) *AKSMaintenanceWeeklySchedule {
	obj := &AKSMaintenanceWeeklySchedule{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["day_of_week"].(string); ok && len(v) > 0 {
		obj.DayOfWeek = v
	}

	if v, ok := in["interval_weeks"].(int); ok && v > 0 {
		obj.IntervalWeeks = v
	}
	return obj
}

func expandAKSMCRelativeMonthlySchedule(p []interface{}) *AKSMaintenanceRelativeMonthlySchedule {
	obj := &AKSMaintenanceRelativeMonthlySchedule{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["day_of_week"].(string); ok && len(v) > 0 {
		obj.DayOfWeek = v
	}

	if v, ok := in["interval_months"].(int); ok && v > 0 {
		obj.IntervalMonths = v
	}

	if v, ok := in["week_index"].(string); ok && len(v) > 0 {
		obj.WeekIndex = v
	}
	return obj
}

func expandAKSMCDailySchedule(p []interface{}) *AKSMaintenanceDailySchedule {
	obj := &AKSMaintenanceDailySchedule{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["interval_days"].(int); ok && v > 0 {
		obj.IntervalDays = v
	}
	return obj
}

func expandAKSMCAbsoluteMonthlySchedule(p []interface{}) *AKSMaintenanceAbsoluteMonthlySchedule {
	obj := &AKSMaintenanceAbsoluteMonthlySchedule{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["day_of_month"].(int); ok && v > 0 {
		obj.DayOfMonth = v
	}

	if v, ok := in["interval_months"].(int); ok && v > 0 {
		obj.IntervalMonths = v
	}
	return obj
}

func expandAKSNodePool(p []interface{}, rawConfig cty.Value) []*AKSNodePool {
	if len(p) == 0 || p[0] == nil {
		return []*AKSNodePool{}
	}

	out := make([]*AKSNodePool, len(p))
	for i := range p {
		obj := AKSNodePool{}
		in := p[i].(map[string]interface{})
		nRawConfig := rawConfig.AsValueSlice()[i]

		if v, ok := in["apiversion"].(string); ok && len(v) > 0 {
			obj.APIVersion = v
		}

		if v, ok := in["name"].(string); ok && len(v) > 0 {
			obj.Name = v
		}

		if v, ok := in["properties"].([]interface{}); ok && len(v) > 0 {
			obj.Properties = expandAKSNodePoolProperties(v, nRawConfig.GetAttr("properties"))
		}

		if v, ok := in["type"].(string); ok && len(v) > 0 {
			obj.Type = v
		}

		if v, ok := in["location"].(string); ok && len(v) > 0 {
			obj.Location = v
		}
		out[i] = &obj
	}
	n1 := spew.Sprintf("%+v", out)
	log.Println("expand sorted node pools:", n1)

	return out
}

func expandAKSNodePoolProperties(p []interface{}, rawConfig cty.Value) *AKSNodePoolProperties {
	obj := &AKSNodePoolProperties{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})
	rawConfig = rawConfig.AsValueSlice()[0]

	if v, ok := in["availability_zones"].([]interface{}); ok && len(v) > 0 {
		obj.AvailabilityZones = toArrayString(v)
	}

	rawCount := rawConfig.GetAttr("count")
	if !rawCount.IsNull() && rawCount.Type().IsPrimitiveType() && rawCount.AsBigFloat().IsInt() {
		val64, _ := rawCount.AsBigFloat().Int64()
		val := int(val64)
		obj.Count = &val
	}

	if v, ok := in["enable_auto_scaling"].(bool); ok {
		obj.EnableAutoScaling = &v
	}

	if v, ok := in["enable_encryption_at_host"].(bool); ok {
		obj.EnableEncryptionAtHost = &v
	}

	if v, ok := in["enable_fips"].(bool); ok {
		obj.EnableFIPS = &v
	}

	if v, ok := in["enable_node_public_ip"].(bool); ok {
		obj.EnableNodePublicIP = &v
	}

	if v, ok := in["enable_ultra_ssd"].(bool); ok {
		obj.EnableUltraSSD = &v
	}

	if v, ok := in["gpu_instance_profile"].(string); ok && len(v) > 0 {
		obj.GpuInstanceProfile = v
	}

	if v, ok := in["kubelet_config"].([]interface{}); ok && len(v) > 0 {
		obj.KubeletConfig = expandAKSNodePoolKubeletConfig(v)
	}

	if v, ok := in["kubelet_disk_type"].(string); ok && len(v) > 0 {
		obj.KubeletDiskType = v
	}

	if v, ok := in["linux_os_config"].([]interface{}); ok && len(v) > 0 {
		obj.LinuxOSConfig = expandAKSNodePoolLinuxOsConfig(v)
	}

	rawCount = rawConfig.GetAttr("max_count")
	if !rawCount.IsNull() && rawCount.Type().IsPrimitiveType() && rawCount.AsBigFloat().IsInt() {
		val64, _ := rawCount.AsBigFloat().Int64()
		val := int(val64)
		obj.MaxCount = &val
	}

	if v, ok := in["max_pods"].(int); ok && v > 0 {
		obj.MaxPods = &v
	}

	rawCount = rawConfig.GetAttr("min_count")
	if !rawCount.IsNull() && rawCount.Type().IsPrimitiveType() && rawCount.AsBigFloat().IsInt() {
		val64, _ := rawCount.AsBigFloat().Int64()
		val := int(val64)
		obj.MinCount = &val
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
		obj.OsDiskSizeGB = &v
	}

	if v, ok := in["os_disk_type"].(string); ok && len(v) > 0 {
		obj.OsDiskType = v
	}

	if v, ok := in["os_sku"].(string); ok && len(v) > 0 {
		obj.OsSku = v
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
	rawScaleSetEvictionPolicy := rawConfig.GetAttr("scale_set_eviction_policy")
	if v, ok := in["scale_set_eviction_policy"].(string); ok && len(v) > 0 && !rawScaleSetEvictionPolicy.IsNull() {
		obj.ScaleSetEvictionPolicy = v
	}
	rawScaleSetPriority := rawConfig.GetAttr("scale_set_priority")
	if v, ok := in["scale_set_priority"].(string); ok && len(v) > 0 && !rawScaleSetPriority.IsNull() {
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
		obj.UpgradeSettings = expandAKSNodePoolUpgradeSettings(v)
	}

	if v, ok := in["vm_size"].(string); ok && len(v) > 0 {
		obj.VmSize = v
	}

	if v, ok := in["vnet_subnet_id"].(string); ok && len(v) > 0 {
		obj.VnetSubnetID = v
	}

	if v, ok := in["creation_data"].([]interface{}); ok && len(v) > 0 {
		obj.CreationData = expandAKSNodePoolCreationData(v)
	}

	return obj
}

func expandAKSNodePoolKubeletConfig(p []interface{}) *AKSNodePoolKubeletConfig {
	obj := &AKSNodePoolKubeletConfig{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["allowed_unsafe_sysctls"].([]interface{}); ok && len(v) > 0 {
		obj.AllowedUnsafeSysctls = toArrayString(v)
	}

	if v, ok := in["container_log_max_files"].(int); ok && v > 0 {
		obj.ContainerLogMaxFiles = &v
	}

	if v, ok := in["container_log_max_size_mb"].(int); ok && v > 0 {
		obj.ContainerLogMaxSizeMB = &v
	}

	if v, ok := in["cpu_cfs_quota"].(bool); ok {
		obj.CpuCfsQuota = &v
	}

	if v, ok := in["cpu_cfs_quota_period"].(string); ok && len(v) > 0 {
		obj.CpuCfsQuotaPeriod = v
	}

	if v, ok := in["cpu_manager_policy"].(string); ok && len(v) > 0 {
		obj.CpuManagerPolicy = v
	}

	if v, ok := in["fail_swap_on"].(bool); ok {
		obj.FailSwapOn = &v
	}

	if v, ok := in["image_gc_high_threshold"].(int); ok && v > 0 {
		obj.ImageGcHighThreshold = &v
	}

	if v, ok := in["image_gc_low_threshold"].(int); ok && v > 0 {
		obj.ImageGcLowThreshold = &v
	}

	if v, ok := in["pod_max_pids"].(int); ok && v > 0 {
		obj.PodMaxPids = &v
	}

	if v, ok := in["topology_manager_policy"].(string); ok && len(v) > 0 {
		obj.TopologyManagerPolicy = v
	}

	return obj
}

func expandAKSNodePoolLinuxOsConfig(p []interface{}) *AKSNodePoolLinuxOsConfig {
	obj := &AKSNodePoolLinuxOsConfig{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["swap_file_size_mb"].(int); ok && v > 0 {
		obj.SwapFileSizeMB = &v
	}

	if v, ok := in["sysctls"].([]interface{}); ok && len(v) > 0 {
		obj.Sysctls = expandAKSNodePoolLinuxOsConfigSysctls(v)
	}

	if v, ok := in["transparent_huge_page_defrag"].(string); ok && len(v) > 0 {
		obj.TransparentHugePageDefrag = v
	}

	if v, ok := in["transparent_huge_page_enabled"].(string); ok && len(v) > 0 {
		obj.TransparentHugePageEnabled = v
	}
	return obj
}

func expandAKSNodePoolLinuxOsConfigSysctls(p []interface{}) *AKSNodePoolLinuxOsConfigSysctls {
	obj := &AKSNodePoolLinuxOsConfigSysctls{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["fs_aio_max_nr"].(int); ok && v > 0 {
		obj.FsAioMaxNr = &v
	}

	if v, ok := in["fs_file_max"].(int); ok && v > 0 {
		obj.FsFileMax = &v
	}

	if v, ok := in["fs_inotify_max_user_watches"].(int); ok && v > 0 {
		obj.FsInotifyMaxUserWatches = &v
	}

	if v, ok := in["fs_nr_open"].(int); ok && v > 0 {
		obj.FsNrOpen = &v
	}

	if v, ok := in["kernel_threads_max"].(int); ok && v > 0 {
		obj.KernelThreadsMax = &v
	}

	if v, ok := in["net_core_netdev_max_backlog"].(int); ok && v > 0 {
		obj.NetCoreNetdevMaxBacklog = &v
	}

	if v, ok := in["net_core_optmem_max"].(int); ok && v > 0 {
		obj.NetCoreOptmemMax = &v
	}

	if v, ok := in["net_core_rmem_default"].(int); ok && v > 0 {
		obj.NetCoreRmemDefault = &v
	}

	if v, ok := in["net_core_rmem_max"].(int); ok && v > 0 {
		obj.NetCoreRmemMax = &v
	}

	if v, ok := in["net_core_somaxconn"].(int); ok && v > 0 {
		obj.NetCoreSomaxconn = &v
	}

	if v, ok := in["net_core_wmem_default"].(int); ok && v > 0 {
		obj.NetCoreWmemDefault = &v
	}

	if v, ok := in["net_core_wmem_max"].(int); ok && v > 0 {
		obj.NetCoreWmemMax = &v
	}

	if v, ok := in["net_ipv4_ip_local_port_range"].(string); ok && len(v) > 0 {
		obj.NetIpv4IpLocalPortRange = v
	}

	if v, ok := in["net_ipv4_neigh_default_gc_thresh1"].(int); ok && v > 0 {
		obj.NetIpv4NeighDefaultGcThresh1 = &v
	}

	if v, ok := in["net_ipv4_neigh_default_gc_thresh2"].(int); ok && v > 0 {
		obj.NetIpv4NeighDefaultGcThresh2 = &v
	}

	if v, ok := in["net_ipv4_neigh_default_gc_thresh3"].(int); ok && v > 0 {
		obj.NetIpv4NeighDefaultGcThresh3 = &v
	}

	if v, ok := in["net_ipv4_tcp_fin_timeout"].(int); ok && v > 0 {
		obj.NetIpv4TcpFinTimeout = &v
	}

	if v, ok := in["net_ipv4_tcpkeepalive_intvl"].(int); ok && v > 0 {
		obj.NetIpv4TcpkeepaliveIntvl = &v
	}

	if v, ok := in["net_ipv4_tcp_keepalive_probes"].(int); ok && v > 0 {
		obj.NetIpv4TcpKeepaliveProbes = &v
	}

	if v, ok := in["net_ipv4_tcp_keepalive_time"].(int); ok && v > 0 {
		obj.NetIpv4TcpKeepaliveTime = &v
	}

	if v, ok := in["net_ipv4_tcp_max_syn_backlog"].(int); ok && v > 0 {
		obj.NetIpv4TcpMaxSynBacklog = &v
	}

	if v, ok := in["net_ipv4_tcp_max_tw_buckets"].(int); ok && v > 0 {
		obj.NetIpv4TcpMaxTwBuckets = &v
	}

	if v, ok := in["net_ipv4_tcp_tw_reuse"].(bool); ok {
		obj.NetIpv4TcpTwReuse = &v
	}

	if v, ok := in["net_netfilter_nf_conntrack_buckets"].(int); ok && v > 0 {
		obj.NetNetfilterNfConntrackBuckets = &v
	}

	if v, ok := in["net_netfilter_nf_conntrack_max"].(int); ok && v > 0 {
		obj.NetNetfilterNfConntrackMax = &v
	}

	if v, ok := in["vm_max_map_count"].(int); ok && v > 0 {
		obj.VmMaxMapCount = &v
	}

	if v, ok := in["vm_swappiness"].(int); ok && v > 0 {
		obj.VmSwappiness = &v
	}

	if v, ok := in["vm_vfs_cache_pressure"].(int); ok && v > 0 {
		obj.VmVfsCachePressure = &v
	}

	return obj
}

func expandAKSNodePoolCreationData(p []interface{}) *AKSNodePoolCreationData {
	obj := &AKSNodePoolCreationData{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})
	if v, ok := in["source_resource_id"].(string); ok && len(v) > 0 {
		obj.SourceResourceId = v
	}
	return obj
}

func expandAKSNodePoolUpgradeSettings(p []interface{}) *AKSNodePoolUpgradeSettings {
	obj := &AKSNodePoolUpgradeSettings{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["max_surge"].(string); ok && len(v) > 0 {
		obj.MaxSurge = v
	}
	return obj
}

// Flatten f

func flattenAKSCluster(d *schema.ResourceData, in *AKSCluster) error {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	rawState := d.GetRawState()

	if len(in.APIVersion) > 0 {
		obj["apiversion"] = in.APIVersion
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
		ret1 = flattenAKSClusterMetadata(in.Metadata, v)
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
		var nRawState cty.Value
		if !rawState.IsNull() {
			nRawState = rawState.GetAttr("spec")
		}
		ret2 = flattenAKSClusterSpec(in.Spec, v, nRawState)
	}

	err = d.Set("spec", ret2)
	if err != nil {
		return err
	}

	return nil

}

func flattenAKSClusterMetadata(in *AKSClusterMetadata, p []interface{}) []interface{} {
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

func flattenAKSClusterSpec(in *AKSClusterSpec, p []interface{}, rawState cty.Value) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if !rawState.IsNull() && len(rawState.AsValueSlice()) > 0 {
		rawState = rawState.AsValueSlice()[0]
	}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.Type) > 0 {
		obj["type"] = in.Type
	}
	if len(in.Blueprint) > 0 {
		obj["blueprint"] = in.Blueprint
	}

	if len(in.BlueprintVersion) > 0 {
		obj["blueprintversion"] = in.BlueprintVersion
	}

	if len(in.CloudProvider) > 0 {
		obj["cloudprovider"] = in.CloudProvider
	}

	if in.AKSClusterConfig != nil {
		v, ok := obj["cluster_config"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		var nRawState cty.Value
		if !rawState.IsNull() {
			nRawState = rawState.GetAttr("cluster_config")
		}
		obj["cluster_config"] = flattenAKSClusterConfig(in.AKSClusterConfig, v, nRawState)
	}

	if in.Sharing != nil {
		obj["sharing"] = flattenV1ClusterSharing(in.Sharing)
	}

	if in.SystemComponentsPlacement != nil {
		v, ok := obj["system_components_placement"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["system_components_placement"] = flattenSystemComponentsPlacement(in.SystemComponentsPlacement, v)
	}

	if in.ProxyConfig != nil {
		obj["proxy_config"] = flattenAKSClusterSpecProxyConfig(in.ProxyConfig)
	}

	return []interface{}{obj}
}

func flattenAKSClusterSpecProxyConfig(in *AKSClusterSpecProxyConfig) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(in.HttpProxy) > 0 {
		obj["http_proxy"] = in.HttpProxy
	}
	if len(in.HttpsProxy) > 0 {
		obj["https_proxy"] = in.HttpsProxy
	}
	if len(in.NoProxy) > 0 {
		obj["no_proxy"] = in.NoProxy
	}
	if !in.Enabled {
		obj["enabled"] = in.Enabled
	}
	if len(in.ProxyAuth) > 0 {
		obj["proxy_auth"] = in.ProxyAuth
	}
	if len(in.BootstrapCA) > 0 {
		obj["bootstrap_ca"] = in.BootstrapCA
	}
	if in.AllowInsecureBootstrap {
		obj["allow_insecure_bootstrap"] = in.AllowInsecureBootstrap
	}
	return []interface{}{obj}
}

func flattenAKSClusterConfig(in *AKSClusterConfig, p []interface{}, rawState cty.Value) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if !rawState.IsNull() && len(rawState.AsValueSlice()) > 0 {
		rawState = rawState.AsValueSlice()[0]
	}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.APIVersion) > 0 {
		obj["apiversion"] = in.APIVersion
	}

	if len(in.Kind) > 0 {
		obj["kind"] = in.Kind
	}

	if in.Metadata != nil {
		v, ok := obj["metadata"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["metadata"] = flattenAKSClusterConfigMetadata(in.Metadata, v)
	}

	if in.Spec != nil {
		v, ok := obj["spec"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		var nRawState cty.Value
		if !rawState.IsNull() {
			nRawState = rawState.GetAttr("spec")
		}
		obj["spec"] = flattenAKSClusterConfigSpec(in.Spec, v, nRawState)
	}

	return []interface{}{obj}
}

func flattenAKSClusterConfigMetadata(in *AKSClusterConfigMetadata, p []interface{}) []interface{} {
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

func flattenAKSClusterConfigSpec(in *AKSClusterConfigSpec, p []interface{}, rawState cty.Value) []interface{} {
	if in == nil {
		return nil
	}
	if !rawState.IsNull() && len(rawState.AsValueSlice()) > 0 {
		rawState = rawState.AsValueSlice()[0]
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
		obj["managed_cluster"] = flattenAKSManagedCluster(in.ManagedCluster, v)
	}

	// @@@@@@@
	if in.NodePools != nil && len(in.NodePools) > 0 {
		v, ok := obj["node_pools"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		var nRawState cty.Value
		if !rawState.IsNull() {
			nRawState = rawState.GetAttr("node_pools")
		}
		obj["node_pools"] = flattenAKSNodePool(in.NodePools, v, nRawState)
	}

	if in.MaintenanceConfigs != nil && len(in.MaintenanceConfigs) > 0 {
		v, ok := obj["maintenance_configurations"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["maintenance_configurations"] = flattenAKSMaintenanceConfigs(in.MaintenanceConfigs, v)
	}

	return []interface{}{obj}

}

func flattenAKSManagedCluster(in *AKSManagedCluster, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.APIVersion) > 0 {
		obj["apiversion"] = in.APIVersion
	}

	if in.ExtendedLocation != nil {
		v, ok := obj["extended_location"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["extended_location"] = flattenAKSManagedClusterExtendedLocation(in.ExtendedLocation, v)
	}

	if in.Identity != nil {
		v, ok := obj["identity"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["identity"] = flattenAKSManagedClusterIdentity(in.Identity, v)
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
		obj["properties"] = flattenAKSManagedClusterProperties(in.Properties, v)
	}

	if in.SKU != nil {
		v, ok := obj["sku"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["sku"] = flattenAKSManagedClusterSKU(in.SKU, v)
	}

	if in.Tags != nil && len(in.Tags) > 0 {
		obj["tags"] = in.Tags
	}

	if len(in.Type) > 0 {
		obj["type"] = in.Type
	}

	if in.AdditionalMetadata != nil {
		v, ok := obj["additional_metadata"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["additional_metadata"] = flattenAKSManagedClusterAdditionalMetadata(in.AdditionalMetadata, v)
	}

	return []interface{}{obj}

}

func flattenAKSManagedClusterExtendedLocation(in *AKSClusterExtendedLocation, p []interface{}) []interface{} {
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

func flattenAKSManagedClusterIdentity(in *AKSManagedClusterIdentity, p []interface{}) []interface{} {
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
		//obj["user_assigned_identities"] = toMapInterface(in.UserAssignedIdentities)
		obj["user_assigned_identities"] = toMapInterfaceObject(in.UserAssignedIdentities)
	}

	return []interface{}{obj}

}

func flattenAKSManagedClusterProperties(in *AKSManagedClusterProperties, p []interface{}) []interface{} {
	log.Printf("Entered flattenAKSManagedClusterProperties")
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if in.AzureADProfile != nil {
		v, ok := obj["aad_profile"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["aad_profile"] = flattenAKSManagedClusterAzureADProfile(in.AzureADProfile, v)
	}
	/*
		if in.AddonProfiles != nil && len(in.AddonProfiles) > 0 {
			obj["addon_profiles"] = toMapInterface(in.AddonProfiles)
		}*/
	if in.AddonProfiles != nil {
		v, ok := obj["addon_profiles"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["addon_profiles"] = flattenAddonProfile(in.AddonProfiles, v)
	}

	if in.APIServerAccessProfile != nil {
		v, ok := obj["api_server_access_profile"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["api_server_access_profile"] = flattenAKSManagedClusterAPIServerAccessProfile(in.APIServerAccessProfile, v)
	}

	if in.AutoScalerProfile != nil {
		v, ok := obj["auto_scaler_profile"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["auto_scaler_profile"] = flattenAKSManagedClusterAutoScalerProfile(in.AutoScalerProfile, v)
	}

	if in.AutoUpgradeProfile != nil {
		v, ok := obj["auto_upgrade_profile"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["auto_upgrade_profile"] = flattenAKSManagedClusterAutoUpgradeProfile(in.AutoUpgradeProfile, v)
	}

	obj["disable_local_accounts"] = in.DisableLocalAccounts

	if len(in.DiskEncryptionSetID) > 0 {
		obj["disk_encryption_set_id"] = in.DiskEncryptionSetID
	}

	if len(in.DNSPrefix) > 0 {
		obj["dns_prefix"] = in.DNSPrefix
	}

	obj["enable_pod_security_policy"] = in.EnablePodSecurityPolicy

	obj["enable_rbac"] = in.EnableRBAC

	if len(in.FQDNSubdomain) > 0 {
		obj["fqdn_subdomain"] = in.FQDNSubdomain
	}

	if in.HTTPProxyConfig != nil {
		v, ok := obj["http_proxy_config"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["http_proxy_config"] = flattenAKSManagedClusterHTTPProxyConfig(in.HTTPProxyConfig, v)
	}

	if in.IdentityProfile != nil {
		v, ok := obj["identity_profile"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["identity_profile"] = flattenAKSManagedClusterIdentityProfile(in.IdentityProfile, v)
	}

	if len(in.KubernetesVersion) > 0 {
		obj["kubernetes_version"] = in.KubernetesVersion
	}

	if in.LinuxProfile != nil {
		v, ok := obj["linux_profile"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["linux_profile"] = flattenAKSManagedClusterLinuxProfile(in.LinuxProfile, v)
	}

	if in.NetworkProfile != nil {
		v, ok := obj["network_profile"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["network_profile"] = flattenAKSMCPropertiesNetworkProfile(in.NetworkProfile, v)
	}

	if len(in.NodeResourceGroup) > 0 {
		obj["node_resource_group"] = in.NodeResourceGroup
	}

	if in.OidcIssuerProfile != nil {
		v, ok := obj["oidc_issuer_profile"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["oidc_issuer_profile"] = flattenAKSMCPropertiesOidcIssuerProfile(in.OidcIssuerProfile, v)
	}

	if in.PodIdentityProfile != nil {
		v, ok := obj["pod_identity_profile"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["pod_identity_profile"] = flattenAKSManagedClusterPodIdentityProfile(in.PodIdentityProfile, v)
	}

	if in.PowerState != nil {
		v, ok := obj["power_state"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["power_state"] = flattenAKSManagedClusterPowerState(in.PowerState, v)
	}

	if in.PrivateLinkResources != nil {
		v, ok := obj["private_link_resources"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["private_link_resources"] = flattenAKSManagedClusterPrivateLinkResources(in.PrivateLinkResources, v)
	}

	if in.SecurityProfile != nil {
		v, ok := obj["security_profile"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["security_profile"] = flattenAKSMCPropertiesSecurityProfile(in.SecurityProfile, v)
	}

	if in.ServicePrincipalProfile != nil {
		v, ok := obj["service_principal_profile"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["service_principal_profile"] = flattenAKSManagedClusterServicePrincipalProfile(in.ServicePrincipalProfile, v)
	}

	if in.WindowsProfile != nil {
		v, ok := obj["windows_profile"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["windows_profile"] = flattenAKSManagedClusterWindowsProfile(in.WindowsProfile, v)
	}

	return []interface{}{obj}

}

func flattenAddonProfile(in *AddonProfiles, p []interface{}) []interface{} {
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
		obj["http_application_routing"] = flattenAKSManagedClusterAddonProfile(in.HttpApplicationRouting, v)
	}
	if in.AzurePolicy != nil {
		v, ok := obj["azure_policy"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["azure_policy"] = flattenAKSManagedClusterAddonProfile(in.AzurePolicy, v)
	}

	if in.OmsAgent != nil {
		v, ok := obj["oms_agent"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["oms_agent"] = flattenAKSManagedClusterAddonOmsAgentProfile(in.OmsAgent, v)
	}

	if in.AzureKeyvaultSecretsProvider != nil {
		v, ok := obj["azure_keyvault_secrets_provider"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["azure_keyvault_secrets_provider"] = flattenAKSManagedClusterAddonAzureKeyvaultSecretsProviderProfile(in.AzureKeyvaultSecretsProvider, v)
	}

	if in.IngressApplicationGateway != nil {
		v, ok := obj["ingress_application_gateway"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["ingress_application_gateway"] = flattenAKSManagedClusterAddonIngressApplicationGatewayProfile(in.IngressApplicationGateway, v)
	}

	return []interface{}{obj}
}

func flattenAKSManagedClusterAddonProfile(in *AKSManagedClusterAddonProfile, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	obj["enabled"] = in.Enabled

	if in.Config != nil && len(in.Config) > 0 {
		//log.Println("type:", reflect.TypeOf(in.AttachPolicy))
		var json2 = jsoniter.ConfigCompatibleWithStandardLibrary
		jsonStr, err := json2.Marshal(in.Config)
		if err != nil {
			log.Println("Config marshal err:", err)
		}
		//log.Println("jsonSTR:", jsonStr)
		obj["config"] = string(jsonStr)
		//log.Println("attach policy flattened correct:", obj["attach_policy"])
	}

	return []interface{}{obj}
}

func flattenAKSManagedClusterAddonOmsAgentProfile(in *OmsAgentProfile, p []interface{}) []interface{} {
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
		obj["config"] = flattenAKSManagedClusterAddonOmsAgentConfigProfile(in.Config, v)
	}

	return []interface{}{obj}
}

func flattenAKSManagedClusterAddonAzureKeyvaultSecretsProviderProfile(in *AzureKeyvaultSecretsProviderProfile, p []interface{}) []interface{} {
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
		obj["config"] = flattenAKSManagedClusterAddonAzureKeyvaultSecretsProviderProfileConfigProfile(in.Config, v)
	}

	return []interface{}{obj}
}

func flattenAKSManagedClusterAddonIngressApplicationGatewayProfile(in *IngressApplicationGatewayAddonProfile, p []interface{}) []interface{} {
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
		obj["config"] = flattenAKSManagedClusterAddonIngressApplicationGatewayConfig(in.Config, v)
	}

	return []interface{}{obj}
}

func flattenAKSManagedClusterAddonIngressApplicationGatewayConfig(in *IngressApplicationGatewayAddonConfig, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.ApplicationGatewayName) > 0 {
		obj["application_gateway_name"] = in.ApplicationGatewayName
	}
	if len(in.ApplicationGatewayID) > 0 {
		obj["application_gateway_id"] = in.ApplicationGatewayID
	}
	if len(in.SubnetCIDR) > 0 {
		obj["subnet_cidr"] = in.SubnetCIDR
	}
	if len(in.SubnetID) > 0 {
		obj["subnet_id"] = in.SubnetID
	}
	if len(in.WatchNamespace) > 0 {
		obj["watch_namespace"] = in.WatchNamespace
	}

	return []interface{}{obj}
}

func flattenAKSManagedClusterAddonOmsAgentConfigProfile(in *OmsAgentConfig, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.LogAnalyticsWorkspaceResourceID) > 0 {
		obj["log_analytics_workspace_resource_id"] = in.LogAnalyticsWorkspaceResourceID
	}

	return []interface{}{obj}
}

func flattenAKSManagedClusterAddonAzureKeyvaultSecretsProviderProfileConfigProfile(in *AzureKeyvaultSecretsProviderProfileConfig, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.EnableSecretRotation) > 0 {
		obj["enable_secret_rotation"] = in.EnableSecretRotation
	}
	if len(in.RotationPollInterval) > 0 {
		obj["rotation_poll_interval"] = in.RotationPollInterval
	}

	return []interface{}{obj}
}

func flattenAKSManagedClusterAzureADProfile(in *AKSManagedClusterAzureADProfile, p []interface{}) []interface{} {
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

	if len(in.ClientAppId) > 0 {
		obj["client_app_id"] = in.ClientAppId
	}

	obj["enable_azure_rbac"] = in.EnableAzureRbac

	obj["managed"] = in.Managed

	if len(in.ServerAppId) > 0 {
		obj["server_app_id"] = in.ServerAppId
	}

	if len(in.ServerAppSecret) > 0 {
		obj["server_app_id_secret"] = in.ServerAppSecret
	}

	if len(in.TenantId) > 0 {
		obj["tenant_id"] = in.TenantId
	}

	return []interface{}{obj}

}

func flattenAKSManagedClusterAPIServerAccessProfile(in *AKSManagedClusterAPIServerAccessProfile, p []interface{}) []interface{} {
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

	if len(in.PrivateDnsZone) > 0 {
		obj["private_dns_zone"] = in.PrivateDnsZone
	}

	return []interface{}{obj}

}

func flattenAKSManagedClusterAutoScalerProfile(in *AKSManagedClusterAutoScalerProfile, p []interface{}) []interface{} {
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

func flattenAKSManagedClusterAutoUpgradeProfile(in *AKSManagedClusterAutoUpgradeProfile, p []interface{}) []interface{} {
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

	if len(in.NodeOsUpgradeChannel) > 0 {
		obj["node_os_upgrade_channel"] = in.NodeOsUpgradeChannel
	}

	return []interface{}{obj}

}

func flattenAKSManagedClusterIdentityProfile(in *AKSManagedClusterIdentityProfile, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if in.KubeletIdentity != nil {
		v, ok := obj["kubelet_identity"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["kubelet_identity"] = flattenAKSManagedClusterIdentityProfileKubeletIdentity(in.KubeletIdentity, v)
	}

	return []interface{}{obj}
}

func flattenAKSManagedClusterIdentityProfileKubeletIdentity(in *AKSManagedClusterKubeletIdentity, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.ResourceId) > 0 {
		obj["resource_id"] = in.ResourceId
	}
	return []interface{}{obj}
}

func flattenAKSManagedClusterHTTPProxyConfig(in *AKSManagedClusterHTTPProxyConfig, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.HTTPProxy) > 0 {
		obj["http_proxy"] = in.HTTPProxy
	}

	if len(in.HTTPProxy) > 0 {
		obj["https_proxy"] = in.HTTPSProxy
	}

	if in.NoProxy != nil && len(in.NoProxy) > 0 {
		obj["no_proxy"] = toArrayInterface(in.NoProxy)
	}

	if len(in.TrustedCA) > 0 {
		obj["trusted_ca"] = in.TrustedCA
	}

	return []interface{}{obj}

}

func flattenAKSManagedClusterLinuxProfile(in *AKSManagedClusterLinuxProfile, p []interface{}) []interface{} {
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

	if in.SSH != nil {
		v, ok := obj["ssh"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["ssh"] = flattenAKSManagedClusterSSHConfig(in.SSH, v)
	}

	if in.NoProxy != nil && len(in.NoProxy) > 0 {
		obj["no_proxy"] = toArrayInterface(in.NoProxy)
	}

	if len(in.TrustedCa) > 0 {
		obj["trusted_ca"] = in.TrustedCa
	}

	return []interface{}{obj}

}

func flattenAKSManagedClusterSSHConfig(in *AKSManagedClusterSSHConfig, p []interface{}) []interface{} {
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
		obj["public_keys"] = flattenAKSManagedClusterSSHKeyData(in.PublicKeys, v)
	}

	return []interface{}{obj}

}

func flattenAKSManagedClusterSSHKeyData(in []*AKSManagedClusterSSHKeyData, p []interface{}) []interface{} {
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

func flattenAKSMCPropertiesNetworkProfile(in *AKSManagedClusterNetworkProfile, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.DNSServiceIP) > 0 {
		obj["dns_service_ip"] = in.DNSServiceIP
	}

	if len(in.DockerBridgeCidr) > 0 {
		obj["docker_bridge_cidr"] = in.DockerBridgeCidr
	}

	if in.LoadBalancerProfile != nil {
		v, ok := obj["load_balancer_profile"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["load_balancer_profile"] = flattenAKSManagedClusterNPLoadBalancerProfile(in.LoadBalancerProfile, v)
	}

	if len(in.LoadBalancerSKU) > 0 {
		obj["load_balancer_sku"] = in.LoadBalancerSKU
	}

	if len(in.NetworkMode) > 0 {
		obj["network_mode"] = in.NetworkMode
	}

	if len(in.NetworkPlugin) > 0 {
		obj["network_plugin"] = in.NetworkPlugin
	}

	if len(in.NetworkPluginMode) > 0 {
		obj["network_plugin_mode"] = in.NetworkPluginMode
	}

	if len(in.NetworkPolicy) > 0 {
		obj["network_policy"] = in.NetworkPolicy
	}

	if len(in.NetworkDataplane) > 0 {
		obj["network_dataplane"] = in.NetworkDataplane
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

func flattenAKSManagedClusterNPLoadBalancerProfile(in *AKSManagedClusterNPLoadBalancerProfile, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if in.AllocatedOutboundPorts != nil {
		obj["allocated_outbound_ports"] = *in.AllocatedOutboundPorts
	}

	if in.EffectiveOutboundIPs != nil && len(in.EffectiveOutboundIPs) > 0 {
		v, ok := obj["effective_outbound_ips"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["effective_outbound_ips"] = flattenAKSManagedClusterNPEffectiveOutboundIPs(in.EffectiveOutboundIPs, v)
	}

	if in.IdleTimeoutInMinutes != nil {
		obj["idle_timeout_in_minutes"] = *in.IdleTimeoutInMinutes
	}

	if in.ManagedOutboundIPs != nil {
		v, ok := obj["managed_outbound_ips"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["managed_outbound_ips"] = flattenAKSManagedClusterNPManagedOutboundIPs(in.ManagedOutboundIPs, v)
	}

	if in.OutboundIPPrefixes != nil {
		v, ok := obj["outbound_ip_prefixes"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["outbound_ip_prefixes"] = flattenAKSManagedClusterNPOutboundIPPrefixes(in.OutboundIPPrefixes, v)
	}

	if in.OutboundIPs != nil {
		v, ok := obj["outbound_ips"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["outbound_ips"] = flattenAKSManagedClusterNPOutboundIPs(in.OutboundIPs, v)
	}

	return []interface{}{obj}

}

func flattenAKSManagedClusterNPEffectiveOutboundIPs(in []*AKSManagedClusterNPEffectiveOutboundIPs, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	out := make([]interface{}, len(in))
	for i, in := range in {

		obj := map[string]interface{}{}
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}

		if len(in.ID) > 0 {
			obj["id"] = in.ID
		}

		out[i] = &obj
	}

	return out

}

func flattenAKSManagedClusterNPManagedOutboundIPs(in *AKSManagedClusterNPManagedOutboundIPs, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if in.Count != nil {
		obj["count"] = *in.Count
	}

	return []interface{}{obj}

}

func flattenAKSManagedClusterNPOutboundIPPrefixes(in *AKSManagedClusterNPOutboundIPPrefixes, p []interface{}) []interface{} {
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
		obj["public_ip_prefixes"] = flattenAKSManagedClusterNPOutboundIPsPublicIPPrefixes(in.PublicIPPrefixes, v)
	}

	return []interface{}{obj}

}

func flattenAKSManagedClusterNPOutboundIPsPublicIPPrefixes(in []*AKSManagedClusterNPManagedOutboundIPsPublicIpPrefixes, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	out := make([]interface{}, len(in))
	for i, in := range in {

		obj := map[string]interface{}{}
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}

		if len(in.ID) > 0 {
			obj["id"] = in.ID
		}

		out[i] = &obj
	}

	return out

}

func flattenAKSManagedClusterNPOutboundIPs(in *AKSManagedClusterNPOutboundIPs, p []interface{}) []interface{} {
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
		obj["public_ips"] = flattenAKSManagedClusterNPOutboundIPsPublicIPs(in.PublicIPs, v)
	}

	return []interface{}{obj}

}

func flattenAKSManagedClusterNPOutboundIPsPublicIPs(in []*AKSManagedClusterNPOutboundIPsPublicIps, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	out := make([]interface{}, len(in))
	for i, in := range in {

		obj := map[string]interface{}{}
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}

		if len(in.ID) > 0 {
			obj["id"] = in.ID
		}

		out[i] = &obj
	}

	return out

}

func flattenAKSMCPropertiesOidcIssuerProfile(in *AKSManagedClusterOidcIssuerProfile, p []interface{}) []interface{} {
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

func flattenAKSManagedClusterPodIdentityProfile(in *AKSManagedClusterPodIdentityProfile, p []interface{}) []interface{} {
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
		obj["user_assigned_identities"] = flattenAKSManagedClusterPIPUserAssignedIdentities(in.UserAssignedIdentities, v)
	}

	if in.UserAssignedIdentityExceptions != nil {
		v, ok := obj["user_assigned_identity_exceptions"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["user_assigned_identity_exceptions"] = flattenAKSManagedClusterPIPUserAssignedIdentityExceptions(in.UserAssignedIdentityExceptions, v)
	}

	return []interface{}{obj}
}

func flattenAKSManagedClusterPowerState(in *AKSManagedClusterPowerState, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	obj["code"] = in.Code

	return []interface{}{obj}
}

func flattenAKSManagedClusterPIPUserAssignedIdentities(inp []*AKSManagedClusterPIPUserAssignedIdentities, p []interface{}) []interface{} {
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
			obj["identity"] = flattenAKSManagedClusterUAIIdentity(in.Identity, v)
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

func flattenAKSManagedClusterUAIIdentity(in *AKSManagedClusterUAIIdentity, p []interface{}) []interface{} {
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

func flattenAKSManagedClusterPIPUserAssignedIdentityExceptions(inp []*AKSManagedClusterPIPUserAssignedIdentityExceptions, p []interface{}) []interface{} {
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

func flattenAKSMCPropertiesSecurityProfile(in *AKSManagedClusterSecurityProfile, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if in.WorkloadIdentity != nil {
		v, ok := obj["workload_identity"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["workload_identity"] = flattenAKSManagedClusterWorkloadIdentity(in.WorkloadIdentity, v)
	}

	return []interface{}{obj}

}

func flattenAKSManagedClusterWorkloadIdentity(in *AKSManagedClusterWorkloadIdentity, p []interface{}) []interface{} {
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

func flattenAKSManagedClusterServicePrincipalProfile(in *AKSManagedClusterServicePrincipalProfile, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.ClientID) > 0 {
		obj["client_id"] = in.ClientID
	}

	if len(in.Secret) > 0 {
		obj["secret"] = in.Secret
	}

	return []interface{}{obj}

}

func flattenAKSManagedClusterWindowsProfile(in *AKSManagedClusterWindowsProfile, p []interface{}) []interface{} {
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

func flattenAKSManagedClusterPrivateLinkResources(in *AKSManagedClusterPrivateLinkResources, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.GroupId) > 0 {
		obj["group_id"] = in.GroupId
	}

	if len(in.ID) > 0 {
		obj["id"] = in.ID
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

	return []interface{}{obj}

}

func flattenAKSManagedClusterSKU(in *AKSManagedClusterSKU, p []interface{}) []interface{} {
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

func flattenAKSManagedClusterAdditionalMetadata(in *AKSManagedClusterAdditionalMetadata, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if in.ACRProfile != nil {
		v, ok := obj["acr_profile"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["acr_profile"] = flattenAKSManagedClusterAdditionalMetadataACRProfile(in.ACRProfile, v)
	}

	if len(in.OmsWorkspaceLocation) > 0 {
		obj["oms_workspace_location"] = in.OmsWorkspaceLocation
	}

	return []interface{}{obj}

}

func flattenAKSManagedClusterAdditionalMetadataACRProfile(in *AKSManagedClusterAdditionalMetadataACRProfile, p []interface{}) []interface{} {
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

	if len(in.ACRName) > 0 {
		obj["acr_name"] = in.ACRName
	}

	if in.Registries != nil && len(in.Registries) > 0 {
		v, ok := obj["registries"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["registries"] = flattenAKSManagedClusterAdditionalMetadataACRProfiles(in.Registries, v)
	}

	return []interface{}{obj}

}

func flattenAKSManagedClusterAdditionalMetadataACRProfiles(in []*AksRegistry, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	out := make([]interface{}, len(in))
	for i, in := range in {
		obj := map[string]interface{}{}
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}

		if len(in.ACRName) > 0 {
			obj["acr_name"] = in.ACRName
		}

		if len(in.ResourceGroupName) > 0 {
			obj["resource_group_name"] = in.ResourceGroupName
		}

		out[i] = &obj
	}

	return out

}

func flattenAKSMaintenanceConfigs(in []*AKSMaintenanceConfig, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
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

		if len(in.Type) > 0 {
			obj["type"] = in.Type
		}

		if in.Properties != nil {
			v, ok := obj["properties"].([]interface{})
			if !ok {
				v = []interface{}{}
			}
			obj["properties"] = flattenAKSMaintenanceConfigProperties(in.Properties, v)
		}
		out[i] = obj
	}
	return out
}

func flattenAKSMaintenanceConfigProperties(in *AKSMaintenanceConfigProperties, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if in.MaintenanceWindow != nil {
		v, ok := obj["maintenance_window"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["maintenance_window"] = flattenAKSMaintenanceWindow(in.MaintenanceWindow, v)
	}
	if in.NotAllowedTime != nil && len(in.NotAllowedTime) > 0 {
		v, ok := obj["not_allowed_time"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["not_allowed_time"] = flattenAKSMCTimeSpan(in.NotAllowedTime, v)
	}
	if in.TimeInWeek != nil {
		v, ok := obj["time_in_week"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["time_in_week"] = flattenAKSMCTimeInWeek(in.TimeInWeek, v)
	}
	return []interface{}{obj}
}

func flattenAKSMCTimeSpan(in []*AKSMaintenanceTimeSpan, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}

	out := make([]interface{}, len(in))
	for i, elem := range in {
		obj := map[string]interface{}{}
		if len(p) != 0 && p[0] != nil {
			obj = p[0].(map[string]interface{})
		}

		if len(elem.End) > 0 {
			obj["end"] = elem.End
		}

		if len(elem.Start) > 0 {
			obj["start"] = elem.Start
		}
		out[i] = obj
	}
	return out
}

func flattenAKSMCTimeInWeek(in []*AKSMaintenanceTimeInWeek, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	out := make([]interface{}, len(in))
	for i, elem := range in {
		obj := map[string]interface{}{}
		if len(p) != 0 && p[0] != nil {
			obj = p[0].(map[string]interface{})
		}

		if len(elem.Day) > 0 {
			obj["day"] = elem.Day
		}

		if elem.HourSlots != nil && len(elem.HourSlots) > 0 {
			obj["hour_slots"] = intArraytoInterfaceArray(elem.HourSlots)
		}
		out[i] = obj
	}
	return out
}

func flattenAKSMaintenanceWindow(in *AKSMaintenanceWindow, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if in.DurationHours > 0 {
		obj["duration_hours"] = in.DurationHours
	}

	if in.NotAllowedDates != nil && len(in.NotAllowedDates) > 0 {
		v, ok := obj["not_allowed_dates"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["not_allowed_dates"] = flattenAKSMCTimeSpan(in.NotAllowedDates, v)
	}

	if in.Schedule != nil {
		v, ok := obj["schedule"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["schedule"] = flattenAKSMCSchedule(in.Schedule, v)
	}

	if len(in.StartDate) > 0 {
		obj["start_date"] = in.StartDate
	}

	if len(in.StartTime) > 0 {
		obj["start_time"] = in.StartTime
	}

	if len(in.UtcOffset) > 0 {
		obj["utc_offset"] = in.UtcOffset
	}
	return []interface{}{obj}
}

func flattenAKSMCSchedule(in *AKSMaintenanceSchedule, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if in.AbsoluteMonthlySchedule != nil {
		v, ok := obj["absolute_monthly"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["absolute_monthly"] = flattenAKSMCAbsoluteMonthlySchedule(in.AbsoluteMonthlySchedule, v)
	}

	if in.DailySchedule != nil {
		v, ok := obj["daily"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["daily"] = flattenAKSMCDailySchedule(in.DailySchedule, v)
	}

	if in.RelativeMonthlySchedule != nil {
		v, ok := obj["relative_monthly"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["relative_monthly"] = flattenAKSMCRelativeMonthlySchedule(in.RelativeMonthlySchedule, v)
	}

	if in.WeeklySchedule != nil {
		v, ok := obj["weekly"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["weekly"] = flattenAKSMCWeeklySchedule(in.WeeklySchedule, v)
	}

	return []interface{}{obj}
}

func flattenAKSMCAbsoluteMonthlySchedule(in *AKSMaintenanceAbsoluteMonthlySchedule, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if in.DayOfMonth > 0 {
		obj["day_of_month"] = in.DayOfMonth
	}

	if in.IntervalMonths > 0 {
		obj["interval_months"] = in.IntervalMonths
	}
	return []interface{}{obj}
}

func flattenAKSMCDailySchedule(in *AKSMaintenanceDailySchedule, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if in.IntervalDays > 0 {
		obj["interval_days"] = in.IntervalDays
	}
	return []interface{}{obj}
}

func flattenAKSMCRelativeMonthlySchedule(in *AKSMaintenanceRelativeMonthlySchedule, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.DayOfWeek) > 0 {
		obj["day_of_week"] = in.DayOfWeek
	}

	if len(in.WeekIndex) > 0 {
		obj["week_index"] = in.WeekIndex
	}

	if in.IntervalMonths > 0 {
		obj["interval_months"] = in.IntervalMonths
	}
	return []interface{}{obj}
}

func flattenAKSMCWeeklySchedule(in *AKSMaintenanceWeeklySchedule, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if in.IntervalWeeks > 0 {
		obj["interval_weeks"] = in.IntervalWeeks
	}

	if len(in.DayOfWeek) > 0 {
		obj["day_of_week"] = in.DayOfWeek
	}
	return []interface{}{obj}
}

func flattenAKSNodePool(in []*AKSNodePool, p []interface{}, rawState cty.Value) []interface{} {
	if in == nil {
		return nil
	}

	// sort the incoming nodepools
	// inToSort := make([]AKSNodePool, len(in))
	// for i := range in {
	// 	inToSort[i] = *in[i]
	// }
	// sort.Sort(ByNodepoolName(inToSort))
	// for i := range inToSort {
	// 	in[i] = &inToSort[i]
	// }
	n1 := spew.Sprintf("%+v", in)
	log.Println("flatten sorted node pools:", n1)
	//log.Println("sorted node pools:", in)
	out := make([]interface{}, len(in))
	for i, in := range in {
		var nRawState cty.Value
		if !rawState.IsNull() && len(rawState.AsValueSlice()) > i {
			nRawState = rawState.AsValueSlice()[0]
		}
		obj := map[string]interface{}{}
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}

		if len(in.APIVersion) > 0 {
			obj["apiversion"] = in.APIVersion
		}

		if len(in.Name) > 0 {
			obj["name"] = in.Name
		}

		if in.Properties != nil {
			v, ok := obj["properties"].([]interface{})
			if !ok {
				v = []interface{}{}
			}
			var propRawState cty.Value
			if !nRawState.IsNull() {
				propRawState = nRawState.GetAttr("properties")
			}
			obj["properties"] = flattenAKSNodePoolProperties(in.Properties, v, propRawState)
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

func flattenAKSNodePoolProperties(in *AKSNodePoolProperties, p []interface{}, rawState cty.Value) []interface{} {
	if in == nil {
		return nil
	}
	if !rawState.IsNull() && len(rawState.AsValueSlice()) > 0 {
		rawState = rawState.AsValueSlice()[0]
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if in.AvailabilityZones != nil && len(in.AvailabilityZones) > 0 {
		obj["availability_zones"] = toArrayInterface(in.AvailabilityZones)
	}

	obj["count"] = in.Count

	if in.EnableAutoScaling != nil {
		obj["enable_auto_scaling"] = *in.EnableAutoScaling
	}

	if in.EnableEncryptionAtHost != nil {
		obj["enable_encryption_at_host"] = *in.EnableEncryptionAtHost
	}

	if in.EnableFIPS != nil {
		obj["enable_fips"] = *in.EnableFIPS
	}

	if in.EnableNodePublicIP != nil {
		obj["enable_node_public_ip"] = *in.EnableNodePublicIP
	}

	if in.EnableUltraSSD != nil {
		obj["enable_ultra_ssd"] = *in.EnableUltraSSD
	}

	if len(in.GpuInstanceProfile) > 0 {
		obj["gpu_instance_profile"] = in.GpuInstanceProfile
	}

	if in.KubeletConfig != nil {
		v, ok := obj["kubelet_config"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["kubelet_config"] = flattenAKSNodePoolKubeletConfig(in.KubeletConfig, v)
	}

	if len(in.KubeletDiskType) > 0 {
		obj["kubelet_disk_type"] = in.KubeletDiskType
	}

	if in.LinuxOSConfig != nil {
		v, ok := obj["linux_os_config"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["linux_os_config"] = flattenAKSNodePoolLinuxOsConfig(in.LinuxOSConfig, v)
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

	if len(in.OsSku) > 0 {
		obj["os_sku"] = in.OsSku
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
	} else if !rawState.IsNull() {
		rawStateScaleSetEvictionPolicy := rawState.GetAttr("scale_set_eviction_policy")
		obj["scale_set_eviction_policy"] = rawStateScaleSetEvictionPolicy.AsString()
	}

	if len(in.ScaleSetPriority) > 0 {
		obj["scale_set_priority"] = in.ScaleSetPriority
	} else if !rawState.IsNull() {
		rawStateScaleSetPriority := rawState.GetAttr("scale_set_priority")
		obj["scale_set_priority"] = rawStateScaleSetPriority.AsString()
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
		obj["upgrade_settings"] = flattenAKSNodePoolUpgradeSettings(in.UpgradeSettings, v)
	}

	if len(in.VmSize) > 0 {
		obj["vm_size"] = in.VmSize
	}

	if len(in.VnetSubnetID) > 0 {
		obj["vnet_subnet_id"] = in.VnetSubnetID
	}

	if in.CreationData != nil {
		v, ok := obj["creation_data"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["creation_data"] = flattenAKSNodePoolCreationData(in.CreationData, v)
	}

	return []interface{}{obj}

}

func flattenAKSNodePoolKubeletConfig(in *AKSNodePoolKubeletConfig, p []interface{}) []interface{} {
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

	if in.CpuCfsQuota != nil {
		obj["cpu_cfs_quota"] = *in.CpuCfsQuota
	}

	if len(in.CpuCfsQuotaPeriod) > 0 {
		obj["cpu_cfs_quota_period"] = in.CpuCfsQuotaPeriod
	}

	if len(in.CpuManagerPolicy) > 0 {
		obj["cpu_manager_policy"] = in.CpuManagerPolicy
	}

	if in.FailSwapOn != nil {
		obj["fail_swap_on"] = *in.FailSwapOn
	}

	obj["image_gc_high_threshold"] = in.ImageGcHighThreshold

	obj["image_gc_low_threshold"] = in.ImageGcLowThreshold

	obj["pod_max_pids"] = in.PodMaxPids

	if len(in.TopologyManagerPolicy) > 0 {
		obj["topology_manager_policy"] = in.TopologyManagerPolicy
	}

	return []interface{}{obj}

}

func flattenAKSNodePoolLinuxOsConfig(in *AKSNodePoolLinuxOsConfig, p []interface{}) []interface{} {
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
		obj["sysctls"] = flattenAKSNodePoolLinuxOsConfigSysctls(in.Sysctls, v)
	}

	if len(in.TransparentHugePageDefrag) > 0 {
		obj["transparent_huge_page_defrag"] = in.TransparentHugePageDefrag
	}

	if len(in.TransparentHugePageEnabled) > 0 {
		obj["transparent_huge_page_enabled"] = in.TransparentHugePageEnabled
	}

	return []interface{}{obj}

}

func flattenAKSNodePoolLinuxOsConfigSysctls(in *AKSNodePoolLinuxOsConfigSysctls, p []interface{}) []interface{} {
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

	if in.NetIpv4TcpTwReuse != nil {
		obj["net_ipv4_tcp_tw_reuse"] = *in.NetIpv4TcpTwReuse
	}

	obj["net_netfilter_nf_conntrack_buckets"] = in.NetNetfilterNfConntrackBuckets

	obj["net_netfilter_nf_conntrack_max"] = in.NetNetfilterNfConntrackMax

	obj["vm_max_map_count"] = in.VmMaxMapCount

	obj["vm_swappiness"] = in.VmSwappiness

	obj["vm_vfs_cache_pressure"] = in.VmVfsCachePressure

	return []interface{}{obj}

}

func flattenAKSNodePoolCreationData(in *AKSNodePoolCreationData, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.SourceResourceId) > 0 {
		obj["source_resource_id"] = in.SourceResourceId
	}

	return []interface{}{obj}
}

func flattenAKSNodePoolUpgradeSettings(in *AKSNodePoolUpgradeSettings, p []interface{}) []interface{} {
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

func aksClusterCTL(config *config.Config, clusterName string, configBytes []byte, dryRun bool, cse string) (string, error) {
	log.Printf("aks cluster ctl start")
	glogger.SetLevel(zap.DebugLevel)
	logger := glogger.GetLogger()
	return clusterctl.Apply(logger, config, clusterName, configBytes, dryRun, false, false, false, uaDef, cse)
}

func aksClusterCTLStatus(taskid, projectID string) (string, error) {
	logger := glogger.GetLogger()
	rctlCfg := config.GetConfig()
	rctlCfg.ProjectID = projectID
	return clusterctl.Status(logger, rctlCfg, taskid)
}

func processInputs(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Println("AKS process inputs")

	desiredObj, err := expandAksCluster(d)
	if err != nil {
		log.Println("error while expanding aks cluster", err)
		return diag.FromErr(err)
	}

	// Only proceed with stitching if the cluster resource already exists
	if d.Id() != "" {
		// ============== Stitching Start ==============

		log.Println("Including first class edge resources in desired spec")

		deployedObj, err := getDeployedClusterSpec(ctx, d)
		if err != nil {
			log.Println("error while reading aks cluster", err)
			return diag.FromErr(err)
		}

		if len(deployedObj.Spec.AKSClusterConfig.Spec.WorkloadIdentities) > 0 {
			// Copy over the WorkloadIdentities from the deployed cluster spec

			desiredObj.Spec.AKSClusterConfig.Spec.WorkloadIdentities = deployedObj.Spec.AKSClusterConfig.Spec.WorkloadIdentities
		}

		// ============== Stitching End ==============
	}

	out, err := yamlf.Marshal(desiredObj)
	if err != nil {
		log.Println("err marshall:", err)
		return diag.FromErr(err)
	}
	log.Printf("AKS Cluster YAML SPEC \n---\n%s\n----\n", out)
	return process_filebytes(ctx, d, m, out, desiredObj)
}

func expandAksCluster(d *schema.ResourceData) (*AKSCluster, error) {
	obj := &AKSCluster{}
	rawConfig := d.GetRawConfig()

	if v, ok := d.Get("apiversion").(string); ok {
		obj.APIVersion = v
	} else {
		log.Println("apiversion unable to be found")
		return obj, fmt.Errorf("%s", "Apiversion is missing")
	}

	if v, ok := d.Get("kind").(string); ok {
		obj.Kind = v
	} else {
		log.Println("kind unable to be found")
		return obj, fmt.Errorf("%s", "Kind is missing")
	}

	if v, ok := d.Get("metadata").([]interface{}); ok {
		obj.Metadata = expandAKSClusterMetadata(v)
		log.Println("md:", obj.Metadata)
	} else {
		log.Println("metadata unable to be found")
		return obj, fmt.Errorf("%s", "Metadata is missing")
	}

	if v, ok := d.Get("spec").([]interface{}); ok {
		obj.Spec = expandAKSClusterSpec(v, rawConfig.GetAttr("spec"))
	} else {
		log.Println("Cluster spec unable to be found")
		return obj, fmt.Errorf("%s", "Spec is missing")
	}

	projectName := obj.Metadata.Project
	_, err := project.GetProjectByName(projectName)
	if err != nil {
		log.Println("Cluster project name is invalid", err)
		return obj, fmt.Errorf("%s", "Cluster project name is invalid")
	}

	if obj.Metadata.Name != obj.Spec.AKSClusterConfig.Metadata.Name {
		return obj, fmt.Errorf("%s", "ClusterConfig name does not match config file")
	}

	return obj, nil
}

func resourceAKSClusterUpsert(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("resourceAKSClusterUpsert")

	return processInputs(ctx, d, m)

}

func process_filebytes(ctx context.Context, d *schema.ResourceData, m interface{}, fileBytes []byte, obj *AKSCluster) diag.Diagnostics {
	log.Printf("process_filebytes")
	var diags diag.Diagnostics
	rctlCfg := config.GetConfig()

	// get project details
	resp, err := project.GetProjectByName(obj.Metadata.Project)
	if err != nil {
		fmt.Printf("project does not exist")
		return diag.FromErr(fmt.Errorf("project does not exist. Error: %s", err.Error()))
	}
	project, err := project.NewProjectFromResponse([]byte(resp))
	if err != nil {
		fmt.Printf("project does not exist")
		return diag.FromErr(fmt.Errorf("project does not exist. Error: %s", err.Error()))
	}

	var cse string
	if obj.Spec.Sharing != nil {
		cse = "false"
	}

	// cluster
	clusterName := obj.Metadata.Name
	response, err := aksClusterCTL(rctlCfg, clusterName, fileBytes, false, cse)
	if err != nil {
		log.Printf("cluster error 1: %s", err)
		return diag.FromErr(err)
	}

	log.Printf("process_filebytes cluster create response : %s", response)
	res := clusterCTLResponse{}
	err = json.Unmarshal([]byte(response), &res)
	if err != nil {
		log.Println("response parse error", err)
		return diag.FromErr(err)
	}
	if res.TaskSetID == "" {
		return nil
	}
	time.Sleep(20 * time.Second)
	s, errGet := cluster.GetCluster(obj.Metadata.Name, project.ID, uaDef)
	if errGet != nil {
		log.Printf("error while getCluster for %s %s", obj.Metadata.Name, errGet.Error())
		return diag.FromErr(errGet)
	}

	log.Printf("Cluster Provision may take upto 15-20 Minutes")
	d.SetId(s.ID)

	ticker := time.NewTicker(time.Duration(60) * time.Second)
	defer ticker.Stop()

	var warnings []string
LOOP:
	for {
		//Check for cluster operation timeout
		select {
		case <-ctx.Done():
			log.Println("Cluster operation stopped due to operation timeout.")
			return diag.Errorf("cluster operation stopped for cluster: `%s` due to operation timeout", clusterName)
		case <-ticker.C:
			check, errGet := cluster.GetCluster(obj.Metadata.Name, project.ID, uaDef)
			if errGet != nil {
				log.Printf("error while getCluster for %s %s", obj.Metadata.Name, errGet.Error())
				return diag.FromErr(errGet)
			}
			edgeId := check.ID
			statusResp, err := aksClusterCTLStatus(res.TaskSetID, project.ID)
			if err != nil {
				log.Println("status response parse error", err)
				return diag.FromErr(err)
			}

			log.Println("statusResp ", statusResp)
			sres := clusterCTLResponse{}
			err = json.Unmarshal([]byte(statusResp), &sres)
			if err != nil {
				log.Println("status response unmarshal error", err)
				return diag.FromErr(err)
			}
			if strings.Contains(sres.Status, "STATUS_COMPLETE") {
				log.Println("Checking in cluster conditions for blueprint sync success..")
				conditionsFailure, clusterReadiness, err := getClusterConditions(edgeId, project.ID)
				if err != nil {
					log.Printf("error while getCluster %s", err.Error())
					return diag.FromErr(err)
				}
				if conditionsFailure {
					log.Printf("blueprint sync failed for edgename: %s and projectname: %s", clusterName, project.Name)
					return diag.FromErr(fmt.Errorf("blueprint sync failed for edgename: %s and projectname: %s", clusterName, project.Name))
				} else if clusterReadiness {
					log.Printf("Cluster operation completed for edgename: %s and projectname: %s", clusterName, project.Name)
					for _, op := range sres.Operations {
						if op == nil {
							continue
						}
						if strings.Compare(op.Operation, edge.ClusterUpgrade.String()) == 0 && op.Error != nil {
							warnings = append(warnings, op.Error.Title)
						}
					}
					break LOOP
				} else {
					log.Println("Cluster Provisiong is Complete. Waiting for cluster to be Ready...")
				}
			} else if strings.Contains(sres.Status, "STATUS_FAILED") {
				failureReasons, err := collectAKSUpsertErrors(sres.Operations)
				if err != nil {
					return diag.FromErr(err)
				}
				return diag.Errorf("Cluster operation failed for edgename: %s and projectname: %s with failure reasons: %s", clusterName, project.Name, failureReasons)
			} else {
				log.Printf("Cluster operation not completed for edgename: %s and projectname: %s. Waiting 60 seconds more for cluster to complete the operation.", clusterName, project.Name)
			}

		}
	}

	edgeDb, err := cluster.GetCluster(obj.Metadata.Name, project.ID, uaDef)
	if err != nil {
		log.Printf("error while getCluster for %s %s", obj.Metadata.Name, err.Error())
		tflog.Error(ctx, "failed to get cluster", map[string]any{"name": obj.Metadata.Name, "pid": project.ID})
		return diag.Errorf("Failed to fetch cluster: %s", err)
	}

	cseFromDb := edgeDb.Settings[clusterSharingExtKey]
	if cseFromDb != "true" {
		if obj.Spec.Sharing == nil && cseFromDb != "" {
			// reset cse as sharing is removed
			edgeDb.Settings[clusterSharingExtKey] = ""
			err := cluster.UpdateCluster(edgeDb, uaDef)
			if err != nil {
				tflog.Error(ctx, "failed to update cluster", map[string]any{"edgeObj": edgeDb})
				return diag.Errorf("Unable to update the edge object, got error: %s", err)
			}
			tflog.Error(ctx, "cse removed successfully")
		}
		if obj.Spec.Sharing != nil && cseFromDb != "false" {
			// explicitly set cse to false
			edgeDb.Settings[clusterSharingExtKey] = "false"
			err := cluster.UpdateCluster(edgeDb, uaDef)
			if err != nil {
				tflog.Error(ctx, "failed to update cluster", map[string]any{"edgeObj": edgeDb})
				return diag.Errorf("Unable to update the edge object, got error: %s", err)
			}
			tflog.Error(ctx, "cse set to false")
		}
	}

	if len(warnings) > 0 {
		diags = make([]diag.Diagnostic, len(warnings))
		for i, message := range warnings {
			diags[i].Severity = diag.Warning
			diags[i].Summary = message
		}
	}
	return diags
}

func collectAKSUpsertErrors(operations []*clusterCTLOperation) (string, error) {
	collectedErrors := AksUpsertErrorFormatter{}
	for _, operation := range operations {
		if strings.Contains(operation.Status, "STATUS_FAILED") && operation.Error != nil {
			collectedErrors.FailureReason += operation.Error.Title + "\n"
		}
	}
	collectedErrsFormattedBytes, err := json.MarshalIndent(collectedErrors, "", "    ")
	if err != nil {
		return "", err
	}
	collectErrs := strings.ReplaceAll(string(collectedErrsFormattedBytes), "\\n", "\n")
	fmt.Println("After MarshalIndent: ", "collectedErrsFormattedBytes", collectErrs)

	return "\n" + collectErrs, nil
}

func resourceAKSClusterCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("create AKS cluster resource")
	return resourceAKSClusterUpsert(ctx, d, m)
}

type ResponseGetClusterSpec struct {
	ClusterYaml string `json:"cluster_yaml"`
}

func resourceAKSClusterRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Println("resourceAKSClusterRead")

	clusterSpec, err := getDeployedClusterSpec(ctx, d)
	if err != nil {
		log.Printf("error in get cluster spec %s", err.Error())
		return diag.FromErr(err)
	}

	// ============== Unfurl Start ==============

	log.Println("Excluding first class edge resources in deployed spec")

	// Remove the cluster associated but externalized edge resources from the deployed cluster
	if len(clusterSpec.Spec.AKSClusterConfig.Spec.WorkloadIdentities) > 0 {
		// WorkloadIdentities is not part of the terraform cluster resource schema

		log.Println("Removing deployed workload identities from deployed cluster spec")
		clusterSpec.Spec.AKSClusterConfig.Spec.WorkloadIdentities = nil
	}

	// ============== Unfurl End =================

	err = flattenAKSCluster(d, clusterSpec)
	if err != nil {
		log.Printf("get aks cluster set error %s", err.Error())
		return diag.FromErr(err)
	}

	return diags
}

func getDeployedClusterSpec(ctx context.Context, d *schema.ResourceData) (*AKSCluster, error) {
	clusterSpec := &AKSCluster{}

	projectName, ok := d.Get("metadata.0.project").(string)
	if !ok || projectName == "" {
		return clusterSpec, errors.New("project name unable to be found")
	}

	clusterName, ok := d.Get("metadata.0.name").(string)
	if !ok || clusterName == "" {
		return clusterSpec, errors.New("cluster name unable to be found")
	}

	fmt.Printf("Found project_name: %s, cluster_name: %s", projectName, clusterName)

	//project details
	projectId, err := getProjectIDFromName(projectName)
	if err != nil {
		fmt.Print("Cluster project name is invalid")
		return clusterSpec, fmt.Errorf("cluster project name is invalid. Error: %s", err.Error())
	}

	c, err := cluster.GetCluster(clusterName, projectId, uaDef)
	if err != nil {
		log.Printf("error in get cluster %s", err.Error())
		if strings.Contains(err.Error(), "not found") {
			log.Println("Resource Read ", "error", err)
			d.SetId("")
			return clusterSpec, fmt.Errorf("resource read failed, cluster not found. Error: %s", err.Error())
		}
		return clusterSpec, err
	}

	cse := c.Settings[clusterSharingExtKey]
	tflog.Info(ctx, "Got cluster from backend", map[string]any{clusterSharingExtKey: cse})

	// another
	logger := glogger.GetLogger()
	rctlCfg := config.GetConfig()
	clusterSpecYaml, err := clusterctl.GetClusterSpec(logger, rctlCfg, c.Name, projectId, uaDef)
	if err != nil {
		log.Printf("error in get clusterspec %s", err.Error())
		return clusterSpec, err
	}
	log.Println("clusterSpecYaml from getClusterSpec:", clusterSpecYaml)

	err = yaml.Unmarshal([]byte(clusterSpecYaml), &clusterSpec)
	if err != nil {
		return clusterSpec, err
	}

	// For backward compatibilty, if the cluster sharing is
	// managed by separately then ignore sharing in cluster
	// config. Both should not be present at a same time.
	if cse == "true" {
		clusterSpec.Spec.Sharing = nil
	}

	log.Println("unmarshalled clusterSpec from getClusterSpec:", clusterSpec)

	return clusterSpec, nil
}

func resourceAKSClusterUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("update AKS cluster resource")

	projectName, ok := d.Get("metadata.0.project").(string)
	if !ok || projectName == "" {
		return diag.FromErr(errors.New("project name unable to be found."))
	}

	clusterName, ok := d.Get("metadata.0.name").(string)
	if !ok || clusterName == "" {
		return diag.FromErr(errors.New("cluster name unable to be found."))
	}

	fmt.Printf("Found project_name: %s, cluster_name: %s", projectName, clusterName)

	projectId, err := getProjectIDFromName(projectName)
	if err != nil {
		fmt.Print("Cluster project name is invalid")
		return diag.FromErr(fmt.Errorf("Cluster project name is invalid. Error: %s", err.Error()))
	}

	c, err := cluster.GetCluster(clusterName, projectId, uaDef)
	if err != nil {
		log.Printf("error in get cluster %s", err.Error())
		return diag.FromErr(err)
	}

	cse := c.Settings[clusterSharingExtKey]
	tflog.Error(ctx, "Cluster fetched", map[string]any{clusterSharingExtKey: cse})

	// Check if cse == true and `spec.sharing` specified. then
	// Error out here only before procedding. The next Upsert is
	// called by "Create" flow as well which is explicitly setting
	// cse to false if `spec.sharing` provided.
	if cse == "true" {
		if d.HasChange("spec.0.sharing") {
			_, new := d.GetChange("spec.0.sharing")
			if new != nil {
				return diag.Errorf("Cluster sharing is currently managed through the external 'rafay_cluster_sharing' resource. To prevent configuration conflicts, please remove the sharing settings from the 'rafay_aks_cluster' resource and manage sharing exclusively via the external resource.")
			}
		}
	}

	return resourceAKSClusterUpsert(ctx, d, m)
}

func resourceAKSClusterDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("resource cluster delete id %s", d.Id())
	projectName, ok := d.Get("metadata.0.project").(string)
	if !ok || projectName == "" {
		return diag.FromErr(errors.New("project name unable to be found."))
	}

	clusterName, ok := d.Get("metadata.0.name").(string)
	if !ok || clusterName == "" {
		return diag.FromErr(errors.New("cluster name unable to be found."))
	}

	fmt.Printf("Found project_name: %s, cluster_name: %s", projectName, clusterName)

	projectId, err := getProjectIDFromName(projectName)
	if err != nil {
		fmt.Print("Cluster project name is invalid")
		return diag.FromErr(fmt.Errorf("Cluster project name is invalid. Error: %s", err.Error()))
	}

	errDel := cluster.DeleteCluster(clusterName, projectId, false, uaDef)
	if errDel != nil {
		log.Printf("delete cluster error %s", errDel.Error())
		return diag.FromErr(errDel)
	}

	ticker := time.NewTicker(time.Duration(60) * time.Second)
	defer ticker.Stop()
LOOP:
	for {
		select {
		case <-ctx.Done():
			log.Printf("Cluster Deletion for edgename: %s and projectname: %s got timeout out.", clusterName, projectName)
			return diag.FromErr(fmt.Errorf("cluster deletion for edgename: %s and projectname: %s got timeout out", clusterName, projectName))
		case <-ticker.C:
			check, errGet := cluster.GetCluster(clusterName, projectId, uaDef)
			if errGet != nil {
				log.Printf("error while getCluster %s, delete success", errGet.Error())
				break LOOP
			}
			if check == nil {
				log.Printf("Cluster Deletion completes for edgename: %s and projectname: %s", clusterName, projectName)
				break LOOP
			}
			log.Printf("Cluster Deletion is in progress for edgename: %s and projectname: %s. Waiting 60 seconds more for operation to complete.", clusterName, projectName)
		}
	}
	return diags
}

// Sort AKS Nodepool

type ByNodepoolName []AKSNodePool

func (np ByNodepoolName) Len() int      { return len(np) }
func (np ByNodepoolName) Swap(i, j int) { np[i], np[j] = np[j], np[i] }
func (np ByNodepoolName) Less(i, j int) bool {
	ret := strings.Compare(np[i].Name, np[j].Name)
	if ret < 0 {
		return true
	} else {
		return false
	}
}
