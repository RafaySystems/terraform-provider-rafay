package rafay

import (
	"context"
	"log"
	"time"

	"github.com/RafaySystems/rctl/pkg/commands"
	"github.com/RafaySystems/rctl/pkg/user"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spf13/cobra"
)

func resourceUser() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceUserCreate,
		ReadContext:   resourceUserRead,
		UpdateContext: resourceUserUpdate,
		DeleteContext: resourceUserDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"user_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"first_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"last_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"phone": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"groups": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceUserCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	cmd := &cobra.Command{
		Use:     "user <group-name>",
		Aliases: []string{"u"},
		Short:   "Create a new user",
		Long:    "Create a new user",
		Example: `
Using command:
	rctl create user john.doe@example.com 
	rctl create user john.doe@example.com --console John, Doe
	rctl create user john.doe@example.com  --groups testingGroup, productionGroup --console John, Doe, 4089382091
`,

		Args: nil,
		RunE: nil,
	}
	var groups []string
	var consoleAccessInputs []string
	consoleAccessInputs = make([]string, 3)
	userName := d.Get("user_name").(string)
	first := d.Get("first_name").(string)
	last := d.Get("last_name").(string)
	phone := d.Get("phone").(string)

	//convert groups interface to passable list for function
	if v, ok := d.Get("groups").([]interface{}); ok && len(v) > 0 {
		groups = toArrayStringSorted(v)
	}
	/*
		if d.Get("groups") != nil {
			groupsList := d.Get("groups").([]interface{})
			groups = make([]string, len(groupsList))
			for i, raw := range groupsList {
				groups[i] = raw.(string)
			}
		}*/
	//create console access input
	if (first != "" || last != "") || phone != "" {
		consoleAccessInputs[0] = first
		consoleAccessInputs[1] = last
		consoleAccessInputs[2] = phone
	}
	//create user
	log.Println("resource user create: ", userName, groups, consoleAccessInputs)
	err := commands.CreateUser(cmd, userName, groups, consoleAccessInputs)
	if err != nil {
		log.Printf("create user error %s", err.Error())
		return diag.FromErr(err)
	}
	//checking the id of the group
	userID, err := user.GetUserIDByName(userName)
	if err != nil {
		log.Printf("create user by id failed to get user, error %s", err.Error())
		return diag.FromErr(err)
	}
	/*
		//checking response of group id
		currUser, err := user.NewUserFromResponse([]byte(userResp))
		if err != nil {
			log.Printf("create user from repsonse failed to get group, error %s", err.Error())
			return diag.FromErr(err)
		}*/
	d.SetId(userID)
	return diags
}

func resourceUserRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	userName := d.Get("user_name").(string)

	//checking the id of the group
	userResp, err := user.GetUserIDByName(userName)
	if err != nil {
		log.Printf("get user id failed to get group, error %s", err.Error())
		return diag.FromErr(err)
	}

	log.Println("user id:", userResp)

	return diags
}

func resourceUserUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	//TODO implement update user
	var diags diag.Diagnostics
	log.Printf("resource user update id %s", d.Id())
	return diags
}

func resourceUserDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	userName := d.Get("user_name").(string)

	log.Printf("resource user delete id %s", userName)
	err := commands.DeleteUser(nil, userName)
	if err != nil {
		log.Printf("delete user error %s", err.Error())
		return diag.FromErr(err)
	}

	return diags
}
