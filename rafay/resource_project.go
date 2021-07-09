package rafay

import (
	"context"
	"fmt"
	"time"

	"github.com/RafaySystems/rctl/pkg/config"
	"github.com/RafaySystems/rctl/pkg/project"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceProject() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceProjectCreate,
		ReadContext:   resourceProjectRead,
		UpdateContext: resourceProjectUpdate,
		DeleteContext: resourceProjectDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceProjectCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	err := project.CreateProject(d.Get("name").(string), d.Get("description").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	resp, err := project.GetProjectByName(d.Get("name").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	p, err := project.NewProjectFromResponse([]byte(resp))
	if err != nil {
		return diag.FromErr(err)
	} else if p == nil {
		d.SetId("")
		return diags
	}

	d.SetId(p.ID)

	return diags
}

func getProjectById(id string) (string, error) {
	auth := config.GetConfig().GetAppAuthProfile()
	uri := "/auth/v1/projects/"
	uri = uri + fmt.Sprintf("%s/", id)
	return auth.AuthAndRequest(uri, "GET", nil)
}

func resourceProjectRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	//resp, err := project.GetProjectByName(d.Get("name").(string))
	resp, err := getProjectById(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	p, err := project.NewProjectFromResponse([]byte(resp))
	if err != nil {
		return diag.FromErr(err)
	} else if p == nil {
		d.SetId("")
		return diags
	}

	if err := d.Set("name", p.Name); err != nil {
		return diag.FromErr(err)
	}

	if len(p.Description) > 0 {
		if err := d.Set("description", p.Description); err != nil {
			return diag.FromErr(err)
		}
	}

	return diags
}

func resourceProjectUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	//TODO implement update project
	var diags diag.Diagnostics
	return diags
}

func resourceProjectDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	err := project.DeleteProjectById(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}
