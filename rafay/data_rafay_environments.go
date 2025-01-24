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

func dataRafayEnvironments() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataRafayEnvironmentRead,
		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"projectname": {
				Description: "Project name from where environments to be listed",
				Type:        schema.TypeString,
				Required:    true,
			},
			"environments": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "A list of environments",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"environment_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "name of the environment",
						},
						"environment_template_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "name of the environment template",
						},
						"template_version": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "version name of the template",
						},
						"status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "status of the environment",
						},
					},
				},
			},
		},
	}
}

func dataRafayEnvironmentRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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

	environments, err := client.EaasV1().Environment().List(ctx, options.ListOptions{
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

	envList := make([]map[string]interface{}, len(environments.Items))

	for i, e := range environments.Items {
		envStatus := "unkown"
		if (e.Status != nil) && (e.Status.DigestedStatus != nil) {
			envStatus = e.Status.DigestedStatus.ConditionStatus.Enum().String()
		}
		envList[i] = map[string]interface{}{
			"environment_name":          e.Metadata.Name,
			"environment_template_name": e.Spec.Template.Name,
			"template_version":          e.Spec.Template.Version,
			"status":                    envStatus,
		}
	}
	if err := d.Set("environments", envList); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(project.ID)
	return diags
}
