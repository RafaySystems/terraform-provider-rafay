package rafay

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/RafaySystems/rafay-common/pkg/hub/client/options"
	typed "github.com/RafaySystems/rafay-common/pkg/hub/client/typed"
	"github.com/RafaySystems/rctl/pkg/project"
	"github.com/RafaySystems/rctl/pkg/versioninfo"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/RafaySystems/rctl/pkg/config"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataRafayNamespaces() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataRafayNamespaceRead,
		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"projectname": {
				Type:     schema.TypeString,
				Required: true,
			},
			"namespaces": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "A list of namespaces",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the namespace.",
						},
						"type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Type of the namespace.",
						},
						"deployed_status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "status of the namespace.",
						},
					},
				},
			},
		},
	}
}

func dataRafayNamespaceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent(), options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return diag.FromErr(err)
	}

	ns, err := client.InfraV3().Namespace().List(ctx, options.ListOptions{
		//Name:    nsTFState.Metadata.Name,
		Project: project.Name,
	})
	if err != nil {
		if strings.Contains(err.Error(), "code 404") {
			log.Println("Resource Read ", "error", err)
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	namespaces := make([]map[string]interface{}, len(ns.Items))

	for i, n := range ns.Items {
		tempType := ""
		var managedClusterList strings.Builder
		if n.Spec.Artifact != nil {
			switch {
			case n.Spec.GetRepo() != nil:
				tempType = "Repo"
			case n.Spec.GetUploaded() != nil:
				tempType = "Uploaded"
			}
		} else {
			tempType = "Wizard"
		}
		for i, ClusterName := range n.Status.DeployedClusters {
			managedClusterList.WriteString(ClusterName)
			if i != len(n.Status.DeployedClusters)-1 {
				managedClusterList.WriteString(", ")
			}
		}
		namespaces[i] = map[string]interface{}{
			"name":            n.Metadata.Name,
			"type":            tempType,
			"deployed_status": managedClusterList.String(),
		}

	}
	if err := d.Set("namespaces", namespaces); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(project.ID)
	return diags
}
