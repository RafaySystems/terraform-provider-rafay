package rafay

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/RafaySystems/rctl/pkg/commands"
	"github.com/RafaySystems/rctl/pkg/group"
	"github.com/RafaySystems/rctl/pkg/groupassociation"
	"github.com/RafaySystems/rctl/pkg/models"
	"github.com/RafaySystems/rctl/pkg/namespace"
	"github.com/RafaySystems/rctl/pkg/project"
	"github.com/RafaySystems/rctl/pkg/user"
	"github.com/RafaySystems/rctl/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceGroupAssociation() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceGroupAssociationCreate,
		ReadContext:   resourceGroupAssociationRead,
		UpdateContext: resourceGroupAssociationUpdate,
		DeleteContext: resourceGroupAssociationDelete,
		Importer: &schema.ResourceImporter{
			State: resourceGroupAssociationImport,
		},

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
				Optional: true,
				// Required: false,
				ForceNew: true,
			},
			"roles": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				AtLeastOneOf: []string{"custom_roles", "roles"},
			},
			"custom_roles": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				AtLeastOneOf: []string{"custom_roles", "roles"},
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

func RemoveDuplicatesFromSlice(strSlice []string) []string {
	uniqueStrings := []string{}
	keys := make(map[string]bool)
	for _, item := range strSlice {
		if _, value := keys[item]; !value {
			keys[item] = true
			uniqueStrings = append(uniqueStrings, item)
		}
	}
	return uniqueStrings
}

func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func returnValidNamespaceNames(namespace_id_list []string, projectID string) ([]string, error) {

	var namespace_name_list []string

	namespaceList, _, err := namespace.ListAllNamespaces(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch namespace details %s", err)
	}

	for _, namespace := range namespaceList {
		if StringInSlice(namespace.Metadata.ID, namespace_id_list) {
			namespace_name_list = append(namespace_name_list, namespace.Metadata.Name)
		}
	}

	return namespace_name_list, nil
}

func resourceGroupAssociationImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	s := strings.Split(d.Id(), "/")
	if len(s) < 2 {
		return nil, fmt.Errorf("group name or project name not provided, usage e.g terraform import resource group-name-project-name")
	}

	group_name := s[0]
	project_name := s[1]

	log.Println("Importing groupassociation for group name: ", group_name, "project name: ", project_name)

	// convert group name to group id

	resp, err := group.GetGroupByName(group_name)
	if err != nil {
		log.Printf("Failed to get group by name, error %s", err.Error())
		return nil, fmt.Errorf("failed to get group by name, error %s", err.Error())
	}

	//checking response of GetGroupByName
	currGroup, err := group.NewGroupFromResponse([]byte(resp), group_name)
	if err != nil {
		log.Printf("Failed to get group by name, error %s", err.Error())
		return nil, fmt.Errorf("failed to get group by name, error %s", err.Error())
	}

	// get project by name
	resp, err = project.GetProjectByName(project_name)
	if err != nil {
		log.Printf("Failed to get project by name, error %s", err.Error())
		return nil, fmt.Errorf("failed to get project by name, error %s", err.Error())
	}
	p, err := project.NewProjectFromResponse([]byte(resp))
	if err != nil {
		return nil, fmt.Errorf("failed to get project by name, error %s", err.Error())
	}

	d.SetId(currGroup.ID + "-" + p.ID)
	log.Printf("ID set by import handler - %s", d.Id())

	return []*schema.ResourceData{d}, nil
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
	//convert custom roles interface to passable list for function
	customRolesList := d.Get("custom_roles").([]interface{})
	customRoles := make([]string, len(customRolesList))
	for i, raw := range customRolesList {
		customRoles[i] = raw.(string)
	}
	//create group association
	log.Printf("resource group assocation create %s", d.Get("group").(string))
	log.Println("roles: ", roles, "namespace: ", namespace)
	if d.Get("project") != nil && d.Get("project").(string) != "" {
		err := commands.CreateProjectAssociation(nil, d.Get("group").(string), d.Get("project").(string), roles, namespace, customRoles)
		if err != nil {
			log.Printf("create group association error %s", err.Error())
			if strings.Contains(err.Error(), "already assigned") {
				// try to update the association
				err := commands.UpdateProjectAssociation(nil, d.Get("group").(string), d.Get("project").(string), roles, namespace, customRoles)
				if err != nil {
					log.Printf("update group association error %s", err.Error())
					if !strings.Contains(err.Error(), "already assigned") {
						return diag.FromErr(err)
					}
				}
			} else {
				return diag.FromErr(err)
			}
		}
		//make sure group project association gets created
		_, err = groupassociation.GetProjectAssociatedWithGroup(d.Get("group").(string))
		if err != nil {
			log.Printf("create group association failed to get group, error %s", err.Error())
			return diag.FromErr(err)
		}
	} else {
		// admin role
		isAdminRole := false
		if utils.StringInSlice("ADMIN", roles) || utils.StringInSlice("ADMINISTRATOR_READ_ONLY", roles) || utils.StringInSlice("FINOPS_ADMIN", roles) {
			isAdminRole = true
		} else if len(customRoles) > 0 {
			// custom role could have an admin base role, not breaking the flow
			isAdminRole = true
		}
		if !isAdminRole {
			return diag.FromErr(fmt.Errorf("project name is missing for non admin/project scoped role"))
		}
		err := commands.CreateAdminRoleAssociation(nil, d.Get("group").(string), roles, customRoles)
		if err != nil {
			log.Printf("create admin role group association error %s", err.Error())
			if strings.Contains(err.Error(), "already assigned") {
				// try to update the association
				err := commands.UpdateAdminRoleAssociation(nil, d.Get("group").(string), roles, customRoles)
				if err != nil {
					log.Printf("update admin role group association error %s", err.Error())
					if !strings.Contains(err.Error(), "already assigned") {
						return diag.FromErr(err)
					}
				}
			} else {
				return diag.FromErr(err)
			}
		}
	}
	//checking the id of the group
	groupResp, err := group.GetGroupByName(d.Get("group").(string))
	if err != nil {
		log.Printf("create group failed to get group, error %s", err.Error())
		return diag.FromErr(err)
	}
	//checking response of group id
	currGroup, err := group.NewGroupFromResponse([]byte(groupResp), d.Get("group").(string))
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

	projectID := ""

	if d.Get("project") != nil && d.Get("project").(string) != "" {
		//checking response of project id
		p, err := project.NewProjectFromResponse([]byte(resp))
		if err != nil {
			return diag.FromErr(err)
		} else if p == nil {
			d.SetId("")
			return diags
		}
		projectID = p.ID
	} else {
		projectID = "all_projects"
	}
	//create user association to group if users are included in resources
	if d.Get("add_users") != nil {
		var addUsers []string
		//convert users interface to passable list for function create
		usersList := d.Get("add_users").([]interface{})
		users := make([]string, len(usersList))
		for i, raw := range usersList {
			users[i] = raw.(string)
		}
		// check user group associationalready exists
		for _, usr := range users {
			grps, err := user.GetUserGroups(usr)
			if err == nil {
				found := false
				for _, grp := range grps {
					if grp == d.Get("group").(string) {
						log.Println("user already associated with group")
						found = true
						break
					}
				}
				if !found {
					addUsers = append(addUsers, usr)
				}
			} else {
				addUsers = append(addUsers, usr)
			}
		}
		//call create user association
		if d.Get("idp_user").(bool) {
			err = groupassociation.CreateIDPUserAssociation(d.Get("group").(string), addUsers)
		} else {
			err = commands.CreateUserAssociation(nil, d.Get("group").(string), addUsers)
		}
		if err != nil {
			log.Println("user association create DID NOT WORK")
		} else {
			log.Println("user association create was created properly to group")
		}

	}
	//creating association id by combining group and project id
	d.SetId(currGroup.ID + "-" + projectID)

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

	//getting project name from id
	if s[1] != "all_projects" {
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
		//check if there is a group association
		respRoles, err := groupassociation.GetProjectAssociatedWithGroup(g.Name)
		if err != nil {
			log.Printf("read group association failed to get group, error %s", err.Error())
			return diag.FromErr(err)
		} else {
			var roleLst []string
			var customRoleLst []string
			var namespace_id_list []string
			gaList := []models.GroupAssociationRoles{}
			err = json.Unmarshal([]byte(respRoles), &gaList)
			if err != nil {
				log.Printf("read group association failed to get roles, error %s", err.Error())
				return diag.FromErr(err)
			}
			for _, sn := range gaList {
				if sn.Project.Name == p.Name {
					projectRoleMap := make(map[string]int)
					for _, cr := range sn.CustomRoles {
						customRoleLst = append(customRoleLst, cr.CustomRole.Name)
						count := 1
						if len(cr.Namespaces) > 0 {
							count = len(cr.Namespaces)
						}
						projectRoleMap[cr.CustomRole.BaseRoleName] = projectRoleMap[cr.CustomRole.BaseRoleName] + count
					}
					for _, cp := range sn.Roles {
						if v, ok := projectRoleMap[cp.Role.Name]; ok && v > 0 {
							projectRoleMap[cp.Role.Name] = v - 1
						} else {
							roleLst = append(roleLst, cp.Role.Name)
						}
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
			if len(customRoleLst) > 0 {
				if err := d.Set("custom_roles", RemoveDuplicatesFromSlice(customRoleLst)); err != nil {
					log.Printf("get group association set custom role error %s", err.Error())
					return diag.FromErr(err)
				}
			} else {
				if err := d.Set("custom_roles", nil); err != nil {
					log.Printf("get group association set custom role error %s", err.Error())
					return diag.FromErr(err)
				}
			}

			if len(namespace_id_list) > 0 {
				namespace_names, _ := returnValidNamespaceNames(namespace_id_list, p.ID)
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
	}
	//setting the group name
	if err := d.Set("group", g.Name); err != nil {
		log.Printf("get group association set error %s", err.Error())
		return diag.FromErr(err)
	}

	return diags
}

func resourceGroupAssociationUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var namespace []string
	var err error
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
	//convert custom roles interface to passable list for function
	customRolesList := d.Get("custom_roles").([]interface{})
	customRoles := make([]string, len(customRolesList))
	for i, raw := range customRolesList {
		customRoles[i] = raw.(string)
	}
	if d.Get("project") != nil && d.Get("project").(string) != "" {
		//update group association
		err := commands.UpdateProjectAssociation(nil, d.Get("group").(string), d.Get("project").(string), roles, namespace, customRoles)
		if err != nil {
			log.Printf("update group association error %s", err.Error())
			return diag.FromErr(err)
		}
	} else {
		// admin role
		isAdminRole := false
		if utils.StringInSlice("ADMIN", roles) || utils.StringInSlice("ADMINISTRATOR_READ_ONLY", roles) || utils.StringInSlice("FINOPS_ADMIN", roles) {
			isAdminRole = true
		} else if len(customRoles) > 0 {
			// custom role could have an admin base role, not breaking the flow
			isAdminRole = true
		}
		if !isAdminRole {
			return diag.FromErr(fmt.Errorf("project name is missing for non admin/project scoped role"))
		}
		err := commands.UpdateAdminRoleAssociation(nil, d.Get("group").(string), roles, customRoles)
		if err != nil {
			log.Printf("update admin role group association error %s", err.Error())
			return diag.FromErr(err)
		}
	}

	if d.Get("remove_users") != nil || d.Get("add_users") != nil {
		var usersToAdd []string
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

		// check user group associationalready exists
		for _, usr := range addUsers {
			grps, err := user.GetUserGroups(usr)
			if err == nil {
				found := false
				for _, grp := range grps {
					if grp == d.Get("group").(string) {
						log.Println("user already associated with group")
						found = true
						break
					}
				}
				if !found {
					usersToAdd = append(usersToAdd, usr)
				}
			} else {
				usersToAdd = append(usersToAdd, usr)
			}
		}
		//call create user association
		if d.Get("idp_user").(bool) {
			err = groupassociation.UpdateIDPUserAssociation(d.Get("group").(string), usersToAdd, removeUsers)
		} else {
			err = commands.UpdateUserAssociation(nil, d.Get("group").(string), usersToAdd, removeUsers)
		}
		log.Println("users to add: ", usersToAdd)
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
	var err error
	//delete association with group name and project name
	//both should be parsed correctly from the response in read function
	log.Printf("group name: %s, project name: %s", d.Get("group").(string), d.Get("project").(string))
	if d.Get("project") != nil && d.Get("project").(string) != "" {
		err = commands.DeleteProjectAssociation(nil, d.Get("group").(string), d.Get("project").(string))
	} else {
		log.Println("Dissociating admin role from group")
		err = commands.DeleteProjectAssociation(nil, d.Get("group").(string), "ALL_PROJECTS")
	}
	if err != nil {
		log.Printf("delete group error %s", err.Error())
		return diag.FromErr(err)
	}
	if d.Get("remove_users") != nil || d.Get("add_users") != nil {
		//convert users interface to passable list for function create
		usersList := d.Get("remove_users").([]interface{})
		users := make([]string, len(usersList))
		for i, raw := range usersList {
			users[i] = raw.(string)
		}

		usersList = d.Get("add_users").([]interface{})
		for _, raw := range usersList {
			users = append(users, raw.(string))
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
