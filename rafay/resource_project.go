package rafay

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/RafaySystems/rctl/pkg/config"
	"github.com/RafaySystems/rctl/pkg/models"
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

	log.Printf("create project with name %s", d.Get("name").(string))
	err := project.CreateProject(d.Get("name").(string), d.Get("description").(string))
	if err != nil {
		log.Printf("create project error %s", err.Error())
		return diag.FromErr(err)
	}

	resp, err := project.GetProjectByName(d.Get("name").(string))
	if err != nil {
		log.Printf("get project after creation failed, error %s", err.Error())
		return diag.FromErr(err)
	}

	p, err := project.NewProjectFromResponse([]byte(resp))
	if err != nil {
		return diag.FromErr(err)
	} else if p == nil {
		d.SetId("")
		return diags
	}

	log.Printf("created project with id %s", p.ID)
	d.SetId(p.ID)

	return diags
}

func getProjectById(id string) (string, error) {
	log.Printf("get project by id %s", id)
	auth := config.GetConfig().GetAppAuthProfile()
	uri := "/auth/v1/projects/"
	uri = uri + fmt.Sprintf("%s/", id)
	return auth.AuthAndRequest(uri, "GET", nil)
}

func getProjectFromResponse(json_data []byte) (*models.Project, error) {
	var pr models.Project
	if err := json.Unmarshal(json_data, &pr); err != nil {
		return nil, err
	}
	return &pr, nil
}

func resourceProjectRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	log.Printf("resource project read id %s", d.Id())
	//resp, err := project.GetProjectByName(d.Get("name").(string))
	resp, err := getProjectById(d.Id())
	if err != nil {
		log.Printf("get project by id, error %s", err.Error())
		return diag.FromErr(err)
	}

	p, err := getProjectFromResponse([]byte(resp))
	if err != nil {
		log.Printf("get project response error %s", err.Error())
		return diag.FromErr(err)
	} else if p == nil {
		d.SetId("")
		return diags
	}

	if err := d.Set("name", p.Name); err != nil {
		log.Printf("read project set name error %s", err.Error())
		return diag.FromErr(err)
	}

	if len(p.Description) > 0 {
		if err := d.Set("description", p.Description); err != nil {
			log.Printf("read project set description error %s", err.Error())
			return diag.FromErr(err)
		}
	}

	return diags
}

func resourceProjectUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	//TODO implement update project
	var diags diag.Diagnostics
	log.Printf("resource project update id %s", d.Id())
	return diags
}

func resourceProjectDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("resource project delete id %s", d.Id())

	err := project.DeleteProjectById(d.Id())
	if err != nil {
		log.Printf("delete project error %s", err.Error())
		return diag.FromErr(err)
	}

	return diags
}
