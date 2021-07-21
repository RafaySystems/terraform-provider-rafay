package rafay

import (
	"context"
	//"encoding/json"
	//"fmt"
	"log"
	"time"

	//uncommment import statements once you start using them
	//"github.com/RafaySystems/rctl/pkg/config"
	"github.com/RafaySystems/rctl/pkg/group"
	//"github.com/RafaySystems/rctl/pkg/models"
	//"github.com/RafaySystems/rctl/pkg/commands"
	//"github.com/RafaySystems/rctl/pkg/groupassociation"
	//"github.com/RafaySystems/rctl/pkg/project"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceImportCluster() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceGroupCreate,
		ReadContext:   resourceGroupRead,
		UpdateContext: resourceGroupUpdate,
		DeleteContext: resourceGroupDelete,

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
			"environment": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"kubernetes_distribution": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			//go through rest of the import cluster process and figure out what else to add to the schema
		},
	}
}

func resourceImportClusterCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	log.Printf("resource greoup create %s", d.Get("name").(string))
	err := group.CreateGroup(d.Get("name").(string), d.Get("description").(string))
	if err != nil {
		log.Printf("create group error %s", err.Error())
		return diag.FromErr(err)
	}

	resp, err := group.GetGroupByName(d.Get("name").(string))
	if err != nil {
		log.Printf("create group failed to get group, error %s", err.Error())
		return diag.FromErr(err)
	}

	g, err := group.NewGroupFromResponse([]byte(resp))
	if err != nil {
		log.Printf("create group failed to parse get response, error %s", err.Error())
		return diag.FromErr(err)
	} else if g == nil {
		log.Printf("create group failed to parse get response")
		d.SetId("")
		return diags
	}

	log.Printf("resource group created %s", g.ID)
	d.SetId(g.ID)

	return diags
}

func resourceImportClusterRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	//resp, err := project.GetProjectByName(d.Get("name").(string))
	resp, err := getGroupById(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	p, err := getGroupFromResponse([]byte(resp))
	if err != nil {
		log.Printf("get group by id, error %s", err.Error())
		return diag.FromErr(err)
	} else if p == nil {
		log.Printf("get group response parse error")
		d.SetId("")
		return diags
	}

	if err := d.Set("name", p.Name); err != nil {
		log.Printf("get group set name error %s", err.Error())
		return diag.FromErr(err)
	}

	return diags
}

func resourceImportClusterUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	//TODO implement update project
	var diags diag.Diagnostics
	log.Printf("resource group update id %s", d.Id())
	return diags
}

func resourceImportClusterDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	log.Printf("resource group delete id %s", d.Id())
	err := group.DeleteGroupById(d.Id())
	if err != nil {
		log.Printf("delete group error %s", err.Error())
		return diag.FromErr(err)
	}

	return diags
}
