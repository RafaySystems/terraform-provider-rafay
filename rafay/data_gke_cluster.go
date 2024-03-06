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
	"github.com/RafaySystems/rctl/pkg/config"
	"github.com/RafaySystems/rctl/pkg/versioninfo"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	schema "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataGKEClusterV3() *schema.Resource {
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

	return &schema.Resource{
		ReadContext: dataGKEClusterV3Read,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
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
		},
	}
}

func dataGKEClusterV3Read(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	log.Printf("dataGKEClusterV3Read GKE")
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	meta := GetMetaData(d)
	if meta == nil {
		return diag.FromErr(fmt.Errorf("%s", "failed to read resource "))
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}

	ag, err := client.InfraV3().Cluster().Get(ctx, options.GetOptions{
		Name:    meta.Name,
		Project: meta.Project,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	err = flattenGKEClusterV3(d, ag)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(meta.Name)
	return diags

}
