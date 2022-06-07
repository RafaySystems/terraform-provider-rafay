package rafay

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/RafaySystems/rctl/pkg/commands"
	"github.com/RafaySystems/rctl/pkg/group"
	"github.com/RafaySystems/rctl/pkg/groupassociation"
	"github.com/RafaySystems/rctl/pkg/project"

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
				Required: true,
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

func resourceGroupAssociationCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var namespace []string
	//schema List returns interface
	//convert roles interface to passable list for function
	rolesList := d.Get("roles").([]interface{})
	roles := make([]string, len(rolesList))
	for i, raw := range rolesList {
		roles[i] = raw.(string)
	}
	//convert namesapce interface to passable list for function
	if d.Get("namespaces") != nil {
		namespaceList := d.Get("namespaces").([]interface{})
		namespace = make([]string, len(namespaceList))
		for i, raw := range namespaceList {
			namespace[i] = raw.(string)
		}
	}
	//create group association
	log.Printf("resource group assocation create %s", d.Get("group").(string))
	log.Println("roles: ", roles, "namespace: ", namespace)
	err := commands.CreateProjectAssociation(nil, d.Get("group").(string), d.Get("project").(string), roles, namespace)
	if err != nil {
		log.Printf("create group association error %s", err.Error())
		return diag.FromErr(err)
	}
	//make sure group project association gets created
	_, err = groupassociation.GetProjectAssociatedWithGroup(d.Get("group").(string))
	if err != nil {
		log.Printf("create group association failed to get group, error %s", err.Error())
		return diag.FromErr(err)
	}
	//checking the id of the group
	groupResp, err := group.GetGroupByName(d.Get("group").(string))
	if err != nil {
		log.Printf("create group failed to get group, error %s", err.Error())
		return diag.FromErr(err)
	}
	//checking response of group id
	currGroup, err := group.NewGroupFromResponse([]byte(groupResp))
	if err != nil {
		log.Printf("create group failed to get group, error %s", err.Error())
		return diag.FromErr(err)
	}
	//checking id of the project
	resp, err := project.GetProjectByName(d.Get("project").(string))
	if err != nil {
		log.Printf("get project after creation failed, error %s", err.Error())
		return diag.FromErr(err)
	}
	//checking response of project id
	p, err := project.NewProjectFromResponse([]byte(resp))
	if err != nil {
		return diag.FromErr(err)
	} else if p == nil {
		d.SetId("")
		return diags
	}
	//create user association to group if users are included in resources
	if d.Get("add_users") != nil {
		//convert users interface to passable list for function create
		usersList := d.Get("add_users").([]interface{})
		users := make([]string, len(usersList))
		for i, raw := range usersList {
			users[i] = raw.(string)
		}
		//call create user association
		if d.Get("idp_user").(bool) {
			err = groupassociation.CreateIDPUserAssociation(d.Get("group").(string), users)
		} else {
			err = commands.CreateUserAssociation(nil, d.Get("group").(string), users)
		}
		if err != nil {
			log.Println("user association create DID NOT WORK")
		} else {
			log.Println("user association create was created properly to group")
		}

	}
	//creating association id by combining group and project id
	d.SetId(currGroup.ID + "-" + p.ID)

	return diags
}

func resourceGroupAssociationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("resource group association read id %s", d.Id())
	//splice association id to get group id s[0] and project id s[1]
	s := strings.Split(d.Id(), "-")
	if len(s) < 2 {
		return diag.FromErr(fmt.Errorf("invalid group association id"))
	}
	//retireve group by id, already created
	resp, err := getGroupById(s[0])
	if err != nil {
		return diag.FromErr(err)
	}
	//check response from group id
	g, err := getGroupFromResponse([]byte(resp))
	if err != nil {
		log.Printf("get group by id, error %s", err.Error())
		return diag.FromErr(err)
	} else if g == nil {
		log.Printf("get group response parse error")
		d.SetId("")
		return diags
	}
	//check if there is a group association
	_, err = groupassociation.GetProjectAssociatedWithGroup(g.Name)
	if err != nil {
		log.Printf("create group association failed to get group, error %s", err.Error())
		return diag.FromErr(err)
	}
	//seeting the group name
	if err := d.Set("group", g.Name); err != nil {
		log.Printf("get group association set error %s", err.Error())
		return diag.FromErr(err)
	}
	//getting project name from id
	resp, err = getProjectById(s[1])
	if err != nil {
		log.Printf("failed to get project ID, error %s", err.Error())
		return diag.FromErr(err)
	}
	//parse response to retrieve only project name
	p, err := getProjectFromResponse([]byte(resp))
	if err != nil {
		log.Printf("get project response error %s", err.Error())
		return diag.FromErr(err)
	} else if p == nil {
		d.SetId("")
		return diags
	}
	//setting project name to p.Name
	if err := d.Set("project", p.Name); err != nil {
		log.Printf("get group association set error %s", err.Error())
		return diag.FromErr(err)
	}

	return diags
}

func resourceGroupAssociationUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var namespace []string
	//schema List returns interface
	//convert roles interface to passable list for function
	rolesList := d.Get("roles").([]interface{})
	roles := make([]string, len(rolesList))
	for i, raw := range rolesList {
		roles[i] = raw.(string)
	}
	//convert namesapce interface to passable list for function
	if d.Get("namespaces") != nil {
		namespaceList := d.Get("namespaces").([]interface{})
		namespace = make([]string, len(namespaceList))
		for i, raw := range namespaceList {
			namespace[i] = raw.(string)
		}
	}
	err := commands.UpdateProjectAssociation(nil, d.Get("group").(string), d.Get("project").(string), roles, namespace)
	if err != nil {
		log.Printf("update group association error %s", err.Error())
		return diag.FromErr(err)
	}
	if d.Get("remove_users") != nil || d.Get("add_users") != nil {
		//convert remove users interface to passable list for function create
		removeUsersList := d.Get("remove_users").([]interface{})
		removeUsers := make([]string, len(removeUsersList))
		for i, raw := range removeUsersList {
			removeUsers[i] = raw.(string)
		}
		//convert remove users interface to passable list for function create
		addUsersList := d.Get("add_users").([]interface{})
		addUsers := make([]string, len(addUsersList))
		for i, raw := range addUsersList {
			addUsers[i] = raw.(string)
		}

		//call create user association
		if d.Get("idp_user").(bool) {
			err = groupassociation.UpdateIDPUserAssociation(d.Get("group").(string), addUsers, removeUsers)
		} else {
			err = commands.UpdateUserAssociation(nil, d.Get("group").(string), addUsers, removeUsers)
		}
		log.Println("users to add: ", addUsers)
		log.Println("users to delete: ", removeUsers)
		if err != nil {
			log.Println("user association update DID NOT WORK: ", err)
		} else {
			log.Println("user association update was created properly to group")
		}

	}
	return diags
}

func resourceGroupAssociationDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	//delete association with group name and project name
	//both should be parsed correctly from the response in read function
	log.Printf("group name: %s, project name: %s", d.Get("group").(string), d.Get("project").(string))
	err := commands.DeleteProjectAssociation(nil, d.Get("group").(string), d.Get("project").(string))
	if err != nil {
		log.Printf("delete group error %s", err.Error())
		return diag.FromErr(err)
	}
	if d.Get("remove_users") != nil {
		//convert users interface to passable list for function create
		usersList := d.Get("remove_users").([]interface{})
		users := make([]string, len(usersList))
		for i, raw := range usersList {
			users[i] = raw.(string)
		}

		//call create user association
		if d.Get("idp_user").(bool) {
			err = groupassociation.DeleteIDPUserAssociation(d.Get("group").(string), users)
		} else {
			err = commands.DeleteUsersAssociation(nil, d.Get("group").(string), users)
		}

		if err != nil {
			log.Println("user association delete DID NOT WORK")
		} else {
			log.Println("user association delete was created properly to group")
		}
	}
	return diags
}
