package rafay

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/RafaySystems/rctl/pkg/group"
	"github.com/RafaySystems/rctl/pkg/groupassociation"
	"github.com/RafaySystems/rctl/pkg/models"
	"github.com/RafaySystems/rctl/pkg/project"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataGroupAssociation() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataGroupAssociationRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
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
			"roles": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"custom_roles": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"namespaces": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"idp_user": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "IDP users vs Local users",
			},
			"add_users": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"remove_users": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func dataGroupAssociationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("resource group association read id %s", d.Id())

	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

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

	//project details
	resp, err = project.GetProjectByName(d.Get("project").(string))
	if err != nil {
		fmt.Print("project name missing in the resource")
		return diags
	}

	project, err := project.NewProjectFromResponse([]byte(resp))
	if err != nil {
		fmt.Printf("project does not exist")
		return diags
	}

	//check if there is a group association
	respRoles, err := groupassociation.GetProjectAssociatedWithGroup(g.Name)
	if err != nil {
		log.Printf("read group association failed to get group, error %s", err.Error())
		return diag.FromErr(err)
	} else {
		var roleLst []string
		var namespace_id_list []string
		gaList := []models.GroupAssociationRoles{}

		// XXX Debug
		// w1 := spew.Sprintf("%+v", respRoles)
		// log.Println("dataGroupAssociationRead respRoles", w1)

		err = json.Unmarshal([]byte(respRoles), &gaList)
		if err != nil {
			log.Printf("read group association failed to get roles, error %s", err.Error())
			return diag.FromErr(err)
		}
		for _, sn := range gaList {
			if sn.Project.Name == project.Name {
				for _, cp := range sn.Roles {
					roleLst = append(roleLst, cp.Role.Name)
					// get namespace from namespace_id
					if cp.NamespaceID != "" {
						namespace_id_list = append(namespace_id_list, cp.NamespaceID)
					}
				}
			}
		}
		if len(roleLst) > 0 {
			//sort.Strings(roleLst)
			if err := d.Set("roles", RemoveDuplicatesFromSlice(roleLst)); err != nil {
				log.Printf("get group association set role error %s", err.Error())
				return diag.FromErr(err)
			}
		} else {
			if err := d.Set("roles", nil); err != nil {
				log.Printf("get group association set role error %s", err.Error())
				return diag.FromErr(err)
			}
		}

		if len(namespace_id_list) > 0 {
			namespace_names, _ := returnValidNamespaceNames(namespace_id_list, project.ID)
			if err := d.Set("namespaces", RemoveDuplicatesFromSlice(namespace_names)); err != nil {
				log.Printf("get group association set namespace error %s", err.Error())
				return diag.FromErr(err)
			}
		} else {
			if err := d.Set("namespaces", nil); err != nil {
				log.Printf("get group association set namespace error %s", err.Error())
				return diag.FromErr(err)
			}
		}
	}
	//setting the group name
	if err := d.Set("group", g.Name); err != nil {
		log.Printf("get group association set error %s", err.Error())
		return diag.FromErr(err)
	}

	d.SetId(g.ID)

	return diags
}
