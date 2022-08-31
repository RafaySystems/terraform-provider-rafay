package rafay

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/RafaySystems/rctl/pkg/cluster"
	"github.com/RafaySystems/rctl/pkg/clusterctl"
	"github.com/RafaySystems/rctl/pkg/config"
	glogger "github.com/RafaySystems/rctl/pkg/log"
	"github.com/RafaySystems/rctl/pkg/project"
	"go.uber.org/zap"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	// Yaml pkg that have no limit for key length
	"github.com/go-yaml/yaml"
	yamlf "github.com/goccy/go-yaml"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type clusterCTLResponse struct {
	TaskSetID string `json:"taskset_id,omitempty"`
	Status    string `json:"status,omitempty"`
}

func resourceAKSCluster() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAKSClusterCreate,
		ReadContext:   resourceAKSClusterRead,
		UpdateContext: resourceAKSClusterUpdate,
		DeleteContext: resourceAKSClusterDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
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
		"cluster_config": {
			Type:     schema.TypeList,
			Required: true,
			Description: "AKS specific cluster configuration	",
			Elem: &schema.Resource{
				Schema: clusterAKSClusterConfig(),
			},
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
		},
		"addon_profiles": {
			Type:        schema.TypeMap,
			Optional:    true,
			Description: "The AKS managed cluster addon profiles",
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
			Type:        schema.TypeInt,
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
			Type:     schema.TypeString,
			Optional: true,
			Description: " 	The fully qualified Azure resource id",
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

///Changes

// New Expand Functions
func expandAKSCluster(p []interface{}) *AKSCluster {
	obj := &AKSCluster{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["apiversion"].(string); ok && len(v) > 0 {
		obj.APIVersion = v
	}

	if v, ok := in["kind"].(string); ok && len(v) > 0 {
		obj.Kind = v
	}

	if v, ok := in["metadata"].([]interface{}); ok && len(v) > 0 {
		obj.Metadata = expandAKSClusterMetadata(v)
	}

	if v, ok := in["spec"].([]interface{}); ok && len(v) > 0 {
		obj.Spec = expandAKSClusterSpec(v)
	}

	return obj
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

func expandAKSClusterSpec(p []interface{}) *AKSClusterSpec {
	obj := &AKSClusterSpec{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

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
		obj.AKSClusterConfig = expandAKSClusterConfig(v)
	}

	return obj
}

func expandAKSClusterConfig(p []interface{}) *AKSClusterConfig {
	obj := &AKSClusterConfig{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

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
		obj.Spec = expandAKSClusterConfigSpec(v)
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

func expandAKSClusterConfigSpec(p []interface{}) *AKSClusterConfigSpec {
	obj := &AKSClusterConfigSpec{}
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
		obj.ManagedCluster = expandAKSConfigManagedCluster(v)
	}

	if v, ok := in["node_pools"].([]interface{}); ok && len(v) > 0 {
		obj.NodePools = expandAKSNodePool(v)
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
		obj.Tags = toMapString(v)
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

	if v, ok := in["addon_profiles"].(map[string]interface{}); ok {
		obj.AddonProfiles = toMapString(v)
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

	if v, ok := in["identity_profile"].(map[string]interface{}); ok {
		obj.IdentityProfile = toMapString(v)
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

	if v, ok := in["ok_total_unready_count"].(int); ok && v > 0 {
		obj.OkTotalUnreadyCount = &v
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

func expandAKSManagedClusterPIPUserAssignedIdentities(p []interface{}) *AKSManagedClusterPIPUserAssignedIdentities {
	obj := &AKSManagedClusterPIPUserAssignedIdentities{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

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

	return obj
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

func expandAKSManagedClusterPIPUserAssignedIdentityExceptions(p []interface{}) *AKSManagedClusterPIPUserAssignedIdentityExceptions {
	obj := &AKSManagedClusterPIPUserAssignedIdentityExceptions{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["name"].(string); ok && len(v) > 0 {
		obj.Name = v
	}

	if v, ok := in["namespace"].(string); ok && len(v) > 0 {
		obj.Namespace = v
	}

	if v, ok := in["pod_labels"].(map[string]interface{}); ok {
		obj.PodLabels = toMapString(v)
	}

	return obj
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

	return obj
}

func expandAKSNodePool(p []interface{}) []*AKSNodePool {
	if len(p) == 0 || p[0] == nil {
		return []*AKSNodePool{}
	}

	out := make([]*AKSNodePool, len(p))
	outToSort := make([]AKSNodePool, len(p))
	for i := range p {
		obj := AKSNodePool{}
		in := p[i].(map[string]interface{})

		if v, ok := in["apiversion"].(string); ok && len(v) > 0 {
			obj.APIVersion = v
		}

		if v, ok := in["name"].(string); ok && len(v) > 0 {
			obj.Name = v
		}

		if v, ok := in["properties"].([]interface{}); ok && len(v) > 0 {
			obj.Properties = expandAKSNodePoolProperties(v)
		}

		if v, ok := in["type"].(string); ok && len(v) > 0 {
			obj.Type = v
		}

		if v, ok := in["location"].(string); ok && len(v) > 0 {
			obj.Location = v
		}
		outToSort[i] = obj
	}

	sort.Sort(ByNodepoolName(outToSort))
	for i := range outToSort {
		out[i] = &outToSort[i]
	}

	return out
}

func expandAKSNodePoolProperties(p []interface{}) *AKSNodePoolProperties {
	obj := &AKSNodePoolProperties{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["availability_zones"].([]interface{}); ok && len(v) > 0 {
		obj.AvailabilityZones = toArrayString(v)
	}

	if v, ok := in["count"].(int); ok && v > 0 {
		obj.Count = &v
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

	if v, ok := in["max_count"].(int); ok && v > 0 {
		obj.MaxCount = &v
	}

	if v, ok := in["max_pods"].(int); ok && v > 0 {
		obj.MaxPods = &v
	}

	if v, ok := in["min_count"].(int); ok && v > 0 {
		obj.MinCount = &v
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
		obj.UpgradeSettings = expandAKSNodePoolUpgradeSettings(v)
	}

	if v, ok := in["vm_size"].(string); ok && len(v) > 0 {
		obj.VmSize = v
	}

	if v, ok := in["vnet_subnet_id"].(string); ok && len(v) > 0 {
		obj.VnetSubnetID = v
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

		ret2 = flattenAKSClusterSpec(in.Spec, v)
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

func flattenAKSClusterSpec(in *AKSClusterSpec, p []interface{}) []interface{} {
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
		obj["cluster_config"] = flattenAKSClusterConfig(in.AKSClusterConfig, v)
	}

	return []interface{}{obj}
}

func flattenAKSClusterConfig(in *AKSClusterConfig, p []interface{}) []interface{} {
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
		obj["spec"] = flattenAKSClusterConfigSpec(in.Spec, v)
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

func flattenAKSClusterConfigSpec(in *AKSClusterConfigSpec, p []interface{}) []interface{} {
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
		obj["managed_cluster"] = flattenAKSManagedCluster(in.ManagedCluster, v)
	}

	// @@@@@@@
	if in.NodePools != nil && len(in.NodePools) > 0 {
		v, ok := obj["node_pools"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["node_pools"] = flattenAKSNodePool(in.NodePools, v)
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

	if in.AddonProfiles != nil && len(in.AddonProfiles) > 0 {
		obj["addon_profiles"] = toMapInterface(in.AddonProfiles)
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
		obj["linux_profile"] = flattenAKSManagedClusterLinuxProfile(in.LinuxProfile, v)
	}

	if in.NetworkProfile != nil {
		v, ok := obj["network_profile"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["network_profile"] = flattenAKSMCPropertiesNetworkProfile(in.NetworkProfile, v)
	}

	if len(in.FQDNSubdomain) > 0 {
		obj["node_resource_group"] = in.NodeResourceGroup
	}

	if in.PodIdentityProfile != nil {
		v, ok := obj["pod_identity_profile"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["pod_identity_profile"] = flattenAKSManagedClusterPodIdentityProfile(in.PodIdentityProfile, v)
	}

	if in.PrivateLinkResources != nil {
		v, ok := obj["private_link_resources"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["private_link_resources"] = flattenAKSManagedClusterPrivateLinkResources(in.PrivateLinkResources, v)
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

	if in.OkTotalUnreadyCount != nil {
		obj["ok_total_unready_count"] = *in.OkTotalUnreadyCount
	}

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

func flattenAKSManagedClusterPIPUserAssignedIdentities(in *AKSManagedClusterPIPUserAssignedIdentities, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
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

	return []interface{}{obj}
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

func flattenAKSManagedClusterPIPUserAssignedIdentityExceptions(in *AKSManagedClusterPIPUserAssignedIdentityExceptions, p []interface{}) []interface{} {
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

	if len(in.Namespace) > 0 {
		obj["namespace"] = in.Namespace
	}

	if in.PodLabels != nil && len(in.PodLabels) > 0 {
		obj["pod_labels"] = toMapInterface(in.PodLabels)
	}

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

	return []interface{}{obj}

}

func flattenAKSNodePool(in []*AKSNodePool, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}

	// sort the incoming nodepools
	inToSort := make([]AKSNodePool, len(in))
	for i := range in {
		inToSort[i] = *in[i]
	}
	sort.Sort(ByNodepoolName(inToSort))
	for i := range inToSort {
		in[i] = &inToSort[i]
	}

	out := make([]interface{}, len(in))
	for i, in := range in {

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
			obj["properties"] = flattenAKSNodePoolProperties(in.Properties, v)
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

func flattenAKSNodePoolProperties(in *AKSNodePoolProperties, p []interface{}) []interface{} {
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
		obj["upgrade_settings"] = flattenAKSNodePoolUpgradeSettings(in.UpgradeSettings, v)
	}

	if len(in.VmSize) > 0 {
		obj["vm_size"] = in.VmSize
	}

	if len(in.VnetSubnetID) > 0 {
		obj["vnet_subnet_id"] = in.VnetSubnetID
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

func aksClusterCTL(config *config.Config, clusterName string, configBytes []byte, dryRun bool) (string, error) {
	log.Printf("aks cluster ctl start")
	glogger.SetLevel(zap.DebugLevel)
	logger := glogger.GetLogger()
	return clusterctl.Apply(logger, config, clusterName, configBytes, dryRun, false)
}

func aksClusterCTLStatus(taskid string) (string, error) {
	logger := glogger.GetLogger()
	rctlCfg := config.GetConfig()
	return clusterctl.Status(logger, rctlCfg, taskid)
}

func processInputs(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	obj := &AKSCluster{}

	if v, ok := d.Get("apiversion").(string); ok {
		obj.APIVersion = v
	} else {
		fmt.Print("apiversion unable to be found")
		return diag.FromErr(fmt.Errorf("%s", "Apiversion is missing"))
	}

	if v, ok := d.Get("kind").(string); ok {
		obj.Kind = v
	} else {
		fmt.Print("kind unable to be found")
		return diag.FromErr(fmt.Errorf("%s", "Kind is missing"))
	}

	if v, ok := d.Get("metadata").([]interface{}); ok {
		obj.Metadata = expandAKSClusterMetadata(v)
	} else {
		fmt.Print("metadata unable to be found")
		return diag.FromErr(fmt.Errorf("%s", "Metadata is missing"))
	}

	if v, ok := d.Get("spec").([]interface{}); ok {
		obj.Spec = expandAKSClusterSpec(v)
	} else {
		fmt.Print("Cluster spec unable to be found")
		return diag.FromErr(fmt.Errorf("%s", "Spec is missing"))
	}

	projectName := obj.Metadata.Project
	_, err := project.GetProjectByName(projectName)
	if err != nil {
		fmt.Print("project name missing in the resource")
		return diag.FromErr(fmt.Errorf("%s", "Project name missing in the resource"))
	}

	if obj.Metadata.Name != obj.Spec.AKSClusterConfig.Metadata.Name {
		return diag.FromErr(fmt.Errorf("%s", "ClusterConfig name does not match config file"))
	}

	out, err := yamlf.Marshal(obj)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("AKS Cluster YAML SPEC \n---\n%s\n----\n", out)
	return process_filebytes(ctx, d, m, out, obj)
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
		return diags
	}
	project, err := project.NewProjectFromResponse([]byte(resp))
	if err != nil {
		fmt.Printf("project does not exist")
		return diags
	}

	// cluster
	clusterName := obj.Metadata.Name
	response, err := aksClusterCTL(rctlCfg, clusterName, fileBytes, false)
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
		log.Println("response res.TaskSetID is empty")
		return diag.FromErr(fmt.Errorf("%s", "response TaskSetID is empty"))
	}
	time.Sleep(20 * time.Second)
	s, errGet := cluster.GetCluster(obj.Metadata.Name, project.ID)
	if errGet != nil {
		log.Printf("error while getCluster for %s %s", obj.Metadata.Name, errGet.Error())
		return diag.FromErr(errGet)
	}

	log.Printf("Cluster Provision may take upto 15-20 Minutes")

	for {
		time.Sleep(60 * time.Second)
		check, errGet := cluster.GetCluster(obj.Metadata.Name, project.ID)
		if errGet != nil {
			log.Printf("error while getCluster for %s %s", obj.Metadata.Name, errGet.Error())
			return diag.FromErr(errGet)
		}

		statusResp, err := aksClusterCTLStatus(res.TaskSetID)
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
			if check.Status == "READY" {
				break
			}
			log.Println("task completed but cluster is not ready")
		}
		if strings.Contains(sres.Status, "STATUS_FAILED") {
			return diag.FromErr(fmt.Errorf("failed to create/update cluster while provisioning cluster %s %s", obj.Metadata.Name, statusResp))
		}
	}
	log.Printf("resource aks cluster created/updated %s", s.ID)
	d.SetId(s.ID)

	return diags
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

	obj := &AKSCluster{}
	if v, ok := d.Get("apiversion").(string); ok {
		obj.APIVersion = v
	} else {
		fmt.Print("apiversion unable to be found")
		return diag.FromErr(fmt.Errorf("%s", "Apiversion is missing"))
	}

	if v, ok := d.Get("kind").(string); ok {
		obj.Kind = v
	} else {
		fmt.Print("kind unable to be found")
		return diag.FromErr(fmt.Errorf("%s", "Kind is missing"))
	}

	if v, ok := d.Get("metadata").([]interface{}); ok {
		obj.Metadata = expandAKSClusterMetadata(v)
	} else {
		fmt.Print("metadata unable to be found")
		return diag.FromErr(fmt.Errorf("%s", "Metadata is missing"))
	}

	if v, ok := d.Get("spec").([]interface{}); ok {
		obj.Spec = expandAKSClusterSpec(v)
	} else {
		fmt.Print("Cluster spec unable to be found")
		return diag.FromErr(fmt.Errorf("%s", "Spec is missing"))
	}

	//project details
	resp, err := project.GetProjectByName(obj.Metadata.Project)
	if err != nil {
		fmt.Print("project name missing in the resource")
		return diags
	}

	project, err := project.NewProjectFromResponse([]byte(resp))
	if err != nil {
		fmt.Printf("project does not exist")
		return diags
	}

	c, err := cluster.GetCluster(obj.Metadata.Name, project.ID)
	if err != nil {
		log.Printf("error in get cluster %s", err.Error())
		return diag.FromErr(err)
	}

	// another
	logger := glogger.GetLogger()
	rctlCfg := config.GetConfig()
	clusterSpecYaml, err := clusterctl.GetClusterSpec(logger, rctlCfg, c.Name, project.ID)
	if err != nil {
		log.Printf("error in get clusterspec %s", err.Error())
		return diag.FromErr(err)
	}

	cluster, err := cluster.GetCluster(c.Name, project.ID)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("Cluster get cluster 2 worked")

	//log.Printf("Cluster from name: %s", cluster)

	fmt.Println(cluster.ClusterType)

	var respGetCfgFile ResponseGetClusterSpec
	if err := json.Unmarshal([]byte(resp), &respGetCfgFile); err != nil {
		return diag.FromErr(err)
	}

	clusterSpec := AKSCluster{}
	err = yaml.Unmarshal([]byte(clusterSpecYaml), &clusterSpec)
	if err != nil {
		return diag.FromErr(err)
	}
	err = flattenAKSCluster(d, &clusterSpec)
	if err != nil {
		log.Printf("get aks cluster set error %s", err.Error())
		return diag.FromErr(err)
	}

	return diags
}

func resourceAKSClusterUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("update AKS cluster resource")
	var diags diag.Diagnostics
	obj := &AKSCluster{}

	if v, ok := d.Get("apiversion").(string); ok {
		obj.APIVersion = v
	} else {
		fmt.Print("apiversion unable to be found")
		return diag.FromErr(fmt.Errorf("%s", "Apiversion is missing"))
	}

	if v, ok := d.Get("kind").(string); ok {
		obj.Kind = v
	} else {
		fmt.Print("kind unable to be found")
		return diag.FromErr(fmt.Errorf("%s", "Kind is missing"))
	}

	if v, ok := d.Get("metadata").([]interface{}); ok {
		obj.Metadata = expandAKSClusterMetadata(v)
	} else {
		fmt.Print("metadata unable to be found")
		return diag.FromErr(fmt.Errorf("%s", "Metadata is missing"))
	}

	if v, ok := d.Get("spec").([]interface{}); ok {
		obj.Spec = expandAKSClusterSpec(v)
	} else {
		fmt.Print("Cluster spec unable to be found")
		return diag.FromErr(fmt.Errorf("%s", "Spec is missing"))
	}

	resp, err := project.GetProjectByName(obj.Metadata.Project)
	if err != nil {
		fmt.Print("project name missing in the resource")
		return diags
	}

	project, err := project.NewProjectFromResponse([]byte(resp))
	if err != nil {
		fmt.Printf("project does not exist")
		return diags
	}
	_, err = cluster.GetCluster(obj.Metadata.Name, project.ID)
	if err != nil {
		log.Printf("error in get cluster %s", err.Error())
		return diag.FromErr(err)
	}

	return resourceAKSClusterUpsert(ctx, d, m)
}

func resourceAKSClusterDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("resource cluster delete id %s", d.Id())
	obj := &AKSCluster{}

	if v, ok := d.Get("apiversion").(string); ok {
		obj.APIVersion = v
	} else {
		fmt.Print("apiversion unable to be found")
		return diag.FromErr(fmt.Errorf("%s", "Apiversion is missing"))
	}

	if v, ok := d.Get("kind").(string); ok {
		obj.Kind = v
	} else {
		fmt.Print("kind unable to be found")
		return diag.FromErr(fmt.Errorf("%s", "Kind is missing"))
	}

	if v, ok := d.Get("metadata").([]interface{}); ok {
		obj.Metadata = expandAKSClusterMetadata(v)
	} else {
		fmt.Print("metadata unable to be found")
		return diag.FromErr(fmt.Errorf("%s", "Metadata is missing"))
	}

	if v, ok := d.Get("spec").([]interface{}); ok {
		obj.Spec = expandAKSClusterSpec(v)
	} else {
		fmt.Print("Cluster spec unable to be found")
		return diag.FromErr(fmt.Errorf("%s", "Spec is missing"))
	}

	resp, err := project.GetProjectByName(obj.Metadata.Project)
	if err != nil {
		fmt.Print("project  does not exist")
		return diags
	}

	project, err := project.NewProjectFromResponse([]byte(resp))
	if err != nil {
		fmt.Printf("project  does not exist")
		return diags
	}

	errDel := cluster.DeleteCluster(obj.Metadata.Name, project.ID, false)
	if errDel != nil {
		log.Printf("delete cluster error %s", errDel.Error())
		return diag.FromErr(errDel)
	}

	for {
		time.Sleep(60 * time.Second)
		check, errGet := cluster.GetCluster(obj.Metadata.Name, project.ID)
		if errGet != nil {
			log.Printf("error while getCluster %s, delete success", errGet.Error())
			break
		}
		if check == nil || (check != nil && check.Status != "READY") {
			break
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
