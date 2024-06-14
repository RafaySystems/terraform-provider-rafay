package rafay

import (
	"github.com/RafaySystems/rafay-common/pkg/hub/terraform/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

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
	features := config.Schema["features"]
	nodePools := config.Schema["node_pools"]
	security := config.Schema["security"]
	pbCommands := config.Schema["pre_bootstrap_commands"]

	return map[string]*schema.Schema{
		"api_version": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "apiVersion of the resource",
		},
		"kind": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "Cluster",
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
								"features":               features,
								"node_pools":             nodePools,
								"security":               security,
								"pre_bootstrap_commands": pbCommands,
							},
						},
					},
				},
			},
		},
	}
}
