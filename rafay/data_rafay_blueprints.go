package rafay

import (
	"context"
	"fmt"
	"time"

	"github.com/RafaySystems/rctl/pkg/project"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/RafaySystems/rctl/pkg/blueprint"
	"github.com/RafaySystems/rctl/pkg/share"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataRafayBlueprints() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataRafayBlueprintsRead,
		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"projectname": {
				Description: "Project name from where blueprints to be listed",
				Type:        schema.TypeString,
				Required:    true,
			},
			"blueprints": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "A list of blueprints",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the blueprint",
						},
						"versions": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Versions count of the blueprint",
						},
						"deployed_clusters": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "deployed clusters count",
						},
						"ownership": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ownership of the blueprint",
						},
					},
				},
			},
		},
	}
}

func dataRafayBlueprintsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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

	bpList, err := blueprint.GetBlueprintSummary(project.ID)
	if err != nil {
		fmt.Printf("project does not exist")
		return diag.Errorf("project %s  does not exist, err: %v", d.Get("projectname").(string), err)
	}

	blueprints := make([]map[string]interface{}, len(bpList))

	for i, bp := range bpList {

		clusters := 0
		for _, bss := range bp.Snapshots {
			clusters += len(bss.Clusters)
		}
		ownershipType := share.GetOwnershipType(project.ID, bp.ProjectID.String())

		blueprints[i] = map[string]interface{}{
			"name":              bp.BlueprintName,
			"versions":          string(rune(len(bp.Snapshots))),
			"deployed_clusters": string(rune(clusters)),
			"ownership":         ownershipType,
		}

	}
	if err := d.Set("blueprints", blueprints); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(project.ID)
	return diags
}
