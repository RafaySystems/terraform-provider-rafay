package rafay

import (
	"context"
	"fmt"
	"time"

	"github.com/RafaySystems/rctl/pkg/cloudprovider"
	"github.com/RafaySystems/rctl/pkg/project"
	"github.com/RafaySystems/rctl/pkg/share"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	schema "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataCloudCredentials() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataCloudCredentialsRead,
		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"projectname": {
				Description: "Project name from where credentials to be listed",
				Type:        schema.TypeString,
				Required:    true,
			},
			"credentials": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "A list of credentials",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the credential",
						},
						"cloud": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The cloud provider of the credential",
						},
						"ownership": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Ownership of the credential",
						},
					},
				},
			},
		},
	}
}

func dataCloudCredentialsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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

	cpList, err := cloudprovider.ListAllCloudProviders(project.ID)
	if err != nil {
		fmt.Printf("project does not exist")
		return diag.Errorf("project %s  does not exist, err: %v", d.Get("projectname").(string), err)
	}

	cloudproviders := make([]map[string]interface{}, len(cpList))

	for i, cp := range cpList {

		ownershipType := share.GetOwnershipType(project.ID, cp.ProjectID)
		providerStr, err := cloudprovider.ConvertProviderIntToStr(cp.Provider)
		if err != nil {
			return diag.Errorf("failed to convert provider int to string, err: %v", err)
		}

		cloudproviders[i] = map[string]interface{}{
			"name":      cp.Name,
			"cloud":     providerStr,
			"ownership": ownershipType,
		}

	}

	if err := d.Set("credentials", cloudproviders); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(project.ID)
	return diags

}
