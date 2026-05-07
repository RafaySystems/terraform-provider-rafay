package rafay

import (
	"strings"

	"github.com/RafaySystems/rafay-common/pkg/hub/terraform/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// suppressComponentsWhenDisabled suppresses diffs on a components list field
// (cloud_monitoring_components / cloud_logging_components) when its paired
// enable_* flag is false. This prevents perpetual plan noise when a user
// keeps component values in HCL but has disabled the feature.
func suppressComponentsWhenDisabled(componentsField, enableField string) schema.SchemaDiffSuppressFunc {
	return func(k, _, _ string, d *schema.ResourceData) bool {
		idx := strings.Index(k, "."+componentsField)
		if idx <= 0 {
			return false
		}
		parent := k[:idx]
		enabled, _ := d.Get(parent + "." + enableField).(bool)
		return !enabled
	}
}

// gkeFeaturesSchema returns a local copy of the upstream features schema with
// DiffSuppressFunc applied to cloud_*_components fields. We deliberately do NOT
// mutate resource.ClusterSchema — that is a package-level singleton shared by
// every resource in this provider and mutating it causes subtle cross-resource bugs.
func gkeFeaturesSchema() *schema.Schema {
	return &schema.Schema{
		Description: "GKE cluster additional features configuration.",
		Type:        schema.TypeList,
		Optional:    true,
		MaxItems:    1,
		MinItems:    1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"cloud_logging_components": {
					Description:      "List of GKE components to enable for cloud logging (e.g., SYSTEM_COMPONENTS, WORKLOADS)",
					Type:             schema.TypeList,
					Optional:         true,
					Elem:             &schema.Schema{Type: schema.TypeString},
					DiffSuppressFunc: suppressComponentsWhenDisabled("cloud_logging_components", "enable_cloud_logging"),
				},
				"cloud_monitoring_components": {
					Description:      "List of GKE components to enable for cloud monitoring (e.g., SYSTEM_COMPONENTS, WORKLOADS)",
					Type:             schema.TypeList,
					Optional:         true,
					Elem:             &schema.Schema{Type: schema.TypeString},
					DiffSuppressFunc: suppressComponentsWhenDisabled("cloud_monitoring_components", "enable_cloud_monitoring"),
				},
				"disable_horizontal_pod_autoscaling": {
					Description: "Disables horizontal pod autoscaling (HPA)",
					Type:        schema.TypeBool,
					Optional:    true,
				},
				"disable_http_load_balancing": {
					Description: "Disables the HTTP (L7) load balancing controller addon for services",
					Type:        schema.TypeBool,
					Optional:    true,
				},
				"enable_application_manager_beta": {
					Description: "Application Manager is a GKE controller for managing the lifecycle of applications. It enables application delivery and updates following Kubernetes and GitOps best practices",
					Type:        schema.TypeBool,
					Optional:    true,
				},
				"enable_backup_for_gke": {
					Description: "Backup for GKE allows you to back up and restore GKE workloads. There is no cost for enabling this feature, but you are charged for backups based on the size of the data and the number of pods you protect",
					Type:        schema.TypeBool,
					Optional:    true,
				},
				"enable_cloud_logging": {
					Description: "Logging collects logs emitted by your applications and by GKE infrastructure",
					Type:        schema.TypeBool,
					Optional:    true,
				},
				"enable_cloud_monitoring": {
					Description: "Monitoring collects metrics emitted by your applications and by GKE infrastructure",
					Type:        schema.TypeBool,
					Optional:    true,
				},
				"enable_compute_engine_persistent_disk_csi_driver": {
					Description: "Enable to automatically deploy and manage the Compute Engine Persistent Disk CSI Driver. This feature is an alternative to using the gcePersistentDisk in-tree volume plugin",
					Type:        schema.TypeBool,
					Optional:    true,
				},
				"enable_filestore_csi_driver": {
					Description: "Enable to automatically deploy and manage the Filestore CSI Driver",
					Type:        schema.TypeBool,
					Optional:    true,
				},
				"enable_gcs_fuse_csi_driver": {
					Description: "Enables the Cloud Storage FUSE CSI driver for mounting GCS buckets",
					Type:        schema.TypeBool,
					Optional:    true,
				},
				"enable_image_streaming": {
					Description: "Image streaming allows your workloads to initialize without waiting for the entire image to download",
					Type:        schema.TypeBool,
					Optional:    true,
				},
				"enable_managed_service_prometheus": {
					Description: "This option deploys managed collectors for Prometheus metrics within this cluster. These collectors must be configured using PodMonitoring resources. To enable Managed Service for Prometheus here, you'll need. Cluster version of 1.21.4-gke.300 or greater",
					Type:        schema.TypeBool,
					Optional:    true,
				},
			},
		},
	}
}

func GKEClusterV3Schema() map[string]*schema.Schema {
	spec := resource.ClusterSchema.Schema["spec"].Elem.(*schema.Resource)
	config := spec.Schema["config"].Elem.(*schema.Resource)

	ctype := spec.Schema["type"]
	blueprint := spec.Schema["blueprint"]
	cc := spec.Schema["cloud_credentials"]
	proxy := spec.Schema["proxy"]
	sharing := spec.Schema["sharing"]

	project := config.Schema["gcp_project"]
	cpVersion := config.Schema["control_plane_version"]
	location := config.Schema["location"]
	network := config.Schema["network"]
	nodePools := config.Schema["node_pools"]
	security := config.Schema["security"]
	pbCommands := config.Schema["pre_bootstrap_commands"]
	resourceLabels := config.Schema["resource_labels"]

	return map[string]*schema.Schema{
		"api_version": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "apiVersion of the resource",
		},
		"kind": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "kind of the resource",
		},
		"metadata": resource.ClusterSchema.Schema["metadata"],
		"spec": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "GKE specific cluster configuration",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"type":              ctype,
					"blueprint":         blueprint,
					"cloud_credentials": cc,
					"proxy":             proxy,
					"sharing":           sharing,
					"config": &schema.Schema{
						Type:        schema.TypeList,
						Optional:    true,
						Description: "GKE cluster config",
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"gcp_project":            project,
								"control_plane_version":  cpVersion,
								"location":               location,
								"network":                network,
								"features":               gkeFeaturesSchema(),
								"node_pools":             nodePools,
								"security":               security,
								"pre_bootstrap_commands": pbCommands,
								"resource_labels":        resourceLabels,
							},
						},
					},
				},
			},
		},
	}
}
