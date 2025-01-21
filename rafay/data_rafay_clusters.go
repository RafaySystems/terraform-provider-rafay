package rafay

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/RafaySystems/rctl/pkg/cluster"
	"github.com/RafaySystems/rctl/pkg/project"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/RafaySystems/rctl/pkg/share"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataRafayClusters() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataRafayClusterRead,
		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"projectname": {
				Description: "Project name from where clusters to be listed",
				Type:        schema.TypeString,
				Required:    true,
			},
			"clusters": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "A list of clusters with their names",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"clustername": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the cluster.",
						},
						"clustertype": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Type of the cluster.",
						},
						"ownership": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ownership of the cluster.",
						},
					},
				},
			},
		},
	}
}

func dataRafayClusterRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	resp, err := project.GetProjectByName(d.Get("projectname").(string))
	if err != nil {
		fmt.Print("project name missing in the resource")
		return diag.Errorf("project name missing in the resource, err: %v", err)
	}

	project, err := project.NewProjectFromResponse([]byte(resp))
	if err != nil {
		fmt.Printf("project does not exist")
		return diag.Errorf("project %s  does not exist, err: %v", d.Get("projectname").(string), err)
	}
	clusterList, err := cluster.ListAllClusters(project.ID, "", "")
	if err != nil {
		log.Printf("error in get cluster %s", err.Error())
		return diag.FromErr(err)
	}
	clusters := make([]map[string]interface{}, len(*clusterList))

	for i, cluster := range *clusterList {
		clusters[i] = map[string]interface{}{
			"clustername": cluster.Name,
			"clustertype": cluster.ClusterType,
			"ownership":   share.GetOwnershipType(project.ID, cluster.ProjectID),
		}
	}

	if err := d.Set("clusters", clusters); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(project.ID)

	return diags
}
