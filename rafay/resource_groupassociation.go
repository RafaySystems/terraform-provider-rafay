package rafay

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/RafaySystems/rctl/pkg/config"
	"github.com/RafaySystems/rctl/pkg/group"
	"github.com/RafaySystems/rctl/pkg/groupassociation"
	"github.com/RafaySystems/rctl/pkg/commands"
	"github.com/RafaySystems/rctl/pkg/models"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceGroupAssociation() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceGroupAssociationCreate,
		ReadContext:   resourceGroupAssociationRead,
		UpdateContext: resourceGroupAssociationUpdate,
		DeleteContext: resourceGroupAssociationDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		//add in all the parameters needed for create association group function
		//also still need to edit resources/rafay_groupassociation files
		Schema: map[string]*schema.Schema{
			"group": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"project": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"roles": {//figure out best way to declare schema type []string, two diff methods in roles and namespaces 
				Type:     schema.TypeList.String(),
				Required: true,
				ForceNew: true,
			},
			"namespaces": { 
				Type:     schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				Required: true,
				ForceNew: true,
			},//add in the rest of the schema from the struct on rctl 
			//call rctl function to create new group with association, make sure you get the right commands 
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceGroupAssociationCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	log.Printf("resource greoup create %s", d.Get("name").(string))
	err := commands.CreateGroupAssociation(nil, d.Get("group").(string), d.Get("project").(string), d.Get("roles"), d.Get("namespace"))
	if err != nil {
		log.Printf("create group error %s", err.Error())
		return diag.FromErr(err)
	}
	//make sure group exists? might not be necessary 
	resp, err := group.GetGroupByName(d.Get("group").(string))
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

	log.Printf("resource greoup created %s", g.ID)
	d.SetId(g.ID)

	return diags
}

func getGroupAssociationById(id string) (string, error) {
	auth := config.GetConfig().GetAppAuthProfile()
	uri := "/auth/v1/groups/"
	uri = uri + fmt.Sprintf("%s/", id)
	return auth.AuthAndRequest(uri, "GET", nil)
}

func getGroupAssociationFromResponse(json_data []byte) (*models.Group, error) {
	var gr models.Group
	if err := json.Unmarshal(json_data, &gr); err != nil {
		return nil, err
	}
	return &gr, nil
}

func resourceGroupAssociationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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

func resourceGroupAssociationUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	//TODO implement update project
	var diags diag.Diagnostics
	log.Printf("resource group update id %s", d.Id())
	return diags
}

func resourceGroupAssociationDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	log.Printf("resource group delete id %s", d.Id())
	err := group.DeleteGroupById(d.Id())
	if err != nil {
		log.Printf("delete group error %s", err.Error())
		return diag.FromErr(err)
	}

	return diags
}
