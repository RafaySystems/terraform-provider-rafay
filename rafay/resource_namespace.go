package rafay

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/RafaySystems/rctl/pkg/namespace"
	"github.com/RafaySystems/rctl/pkg/project"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceNamespace() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceNamespaceCreate,
		ReadContext:   resourceNamespaceRead,
		UpdateContext: resourceNamespaceUpdate,
		DeleteContext: resourceNamespaceDelete,

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
			},
			"projectname": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"psp": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceNamespaceCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	log.Printf("create namespace with name %s projectname %s", d.Get("name").(string), d.Get("projectname").(string))

	resp, err := project.GetProjectByName(d.Get("projectname").(string))
	if err != nil {
		log.Printf("project does not exist, error %s", err.Error())
		return diag.FromErr(err)
	}
	p, err := project.NewProjectFromResponse([]byte(resp))
	if err != nil {
		return diag.FromErr(err)
	} else if p == nil {
		d.SetId("")
		return diags
	}
	project_id := p.ID

	log.Printf("create namespace project_id %s description '%s' psp '%s' ", project_id, d.Get("description").(string), d.Get("psp").(string))
	err = namespace.CreateNamespace(d.Get("name").(string), d.Get("description").(string), d.Get("psp").(string), project_id)
	if err != nil {
		log.Printf("create name space error %s", err.Error())
		return diag.FromErr(err)
	}

	log.Printf("created namespace with name %s", d.Get("name").(string))
	d.SetId(d.Get("name").(string) + ":" + project_id)

	return diags
}

func resourceNamespaceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	log.Printf("resource namespace read id %s", d.Id())
	s := strings.Split(d.Id(), ":")
	if len(s) < 2 {
		log.Printf("invalid namespace respurce %s", d.Id())
		return diag.FromErr(fmt.Errorf("invalid namespace respurce %s", d.Id()))
	}
	namespace := s[0]
	project_id := s[1]

	resp, err := getProjectById(project_id)
	if err != nil {
		log.Printf("get project by id, error %s", err.Error())
		return diag.FromErr(err)
	}

	p, err := getProjectFromResponse([]byte(resp))
	if err != nil {
		log.Printf("get project response error %s", err.Error())
		return diag.FromErr(err)
	} else if p == nil {
		log.Printf("get project response error ")
		return diags
	}

	if err := d.Set("name", namespace); err != nil {
		log.Printf("read namespace set name error %s", err.Error())
		return diag.FromErr(err)
	}

	if err := d.Set("projectname", p.Name); err != nil {
		log.Printf("read namespace set project error %s", err.Error())
		return diag.FromErr(err)
	}

	if len(p.Description) > 0 {
		if err := d.Set("description", p.Description); err != nil {
			log.Printf("read namespace set description error %s", err.Error())
			return diag.FromErr(err)
		}
	}

	return diags
}

func resourceNamespaceUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	//TODO implement update namespace
	var diags diag.Diagnostics
	log.Printf("resource namespace update id %s", d.Id())
	return diags
}

func resourceNamespaceDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("resource namespace delete id %s", d.Id())

	s := strings.Split(d.Id(), ":")
	if len(s) < 2 {
		log.Printf("invalid namespace respurce %s", d.Id())
		return diag.FromErr(fmt.Errorf("invalid namespace respurce %s", d.Id()))
	}
	ns := s[0]
	project_id := s[1]

	err := namespace.DeleteNamespaceByName(ns, project_id)
	if err != nil {
		log.Printf("delete namespace error %s", err.Error())
		return diag.FromErr(err)
	}

	return diags
}
