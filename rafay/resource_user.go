package rafay

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/RafaySystems/rctl/pkg/commands"
	"github.com/RafaySystems/rctl/pkg/models"
	"github.com/RafaySystems/rctl/pkg/user"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceUser() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceUserCreate,
		ReadContext:   resourceUserRead,
		UpdateContext: resourceUserUpdate,
		DeleteContext: resourceUserDelete,
		Importer: &schema.ResourceImporter{
			State: resourceUserImport,
		},

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
				//ForceNew: true,
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
			"generate_apikey": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"console_access": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"groups": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"apikey": {
				Type:      schema.TypeString,
				Optional:  true,
				Computed:  true,
				ForceNew:  true,
				Sensitive: true,
			},
			"api_secret": {
				Type:      schema.TypeString,
				Optional:  true,
				Computed:  true,
				ForceNew:  true,
				Sensitive: true,
			},
		},
	}
}

func resourceUserCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceUserUpsert(ctx, d, true)
}

func resourceUserUpsert(ctx context.Context, d *schema.ResourceData, create bool) diag.Diagnostics {
	var diags diag.Diagnostics
	var err error
	var groups []string
	var consoleAccessInputs []string
	var api, secret string

	isGroup := false
	isAPi := false
	isConsole := false
	consoleAccessInputs = make([]string, 3)
	userName := d.Get("user_name").(string)
	first := d.Get("first_name").(string)
	last := d.Get("last_name").(string)
	phone := d.Get("phone").(string)

	if d.State() != nil && d.State().ID != "" {
		if userName != "" && userName != d.State().ID {
			log.Printf("username change not supported")
			d.State().Tainted = true
			return diag.FromErr(fmt.Errorf("%s", "username change not supported"))
		}
	}

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
	if len(groups) > 0 {
		isGroup = true
	}

	if d.Get("generate_apikey").(bool) {
		log.Println("create user generate_apikey ")
		isAPi = true
	}

	if d.State() != nil && d.State().ID != "" {
		log.Println(" console_access ", d.State().Attributes["console_access"])
		userAccount, err := user.GetUser(userName)
		if err != nil {
			log.Printf("get user account, error %s", err.Error())
			return diag.FromErr(err)
		}
		if userAccount.Account.UserType != "CONSOLE" && d.Get("console_access").(bool) {
			return diag.FromErr(fmt.Errorf("%s", "console_access change not supported"))
		}
		if userAccount.Account.UserType == "CONSOLE" && !d.Get("console_access").(bool) {
			return diag.FromErr(fmt.Errorf("%s", "console_access change not supported"))
		}
	}

	if d.Get("console_access").(bool) {
		log.Println("create user console_access ")
		isConsole = true
	}

	var account *commands.CreateAccount
	if create {
		account, api, secret, err = commands.CreateUserTF(userName, groups, consoleAccessInputs, isGroup, isAPi, isConsole)
		if len(api) > 0 {
			d.Set("apikey", api)
		}
		if len(secret) > 0 {
			d.Set("api_secret", secret)
		}
	} else {
		account, api, secret, err = commands.UpdateUserTF(userName, groups, consoleAccessInputs, isGroup, isAPi, isConsole)
		if len(api) > 0 {
			d.Set("apikey", api)
		}
		if len(secret) > 0 {
			d.Set("api_secret", secret)
		}
	}

	log.Println(" UsertUser ", userName, groups, consoleAccessInputs, isGroup, isAPi, isConsole)
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

	log.Println(" CreateUserTF account", account, userID)

	// if d.Get("generate_apikey").(bool) {
	// 	apiKey, apiSecret, err := user.GetUserAPIKey(userName)
	// 	if err == nil {
	// 		d.Set("apikey", apiKey)
	// 		if len(apiSecret) > 0 {
	// 			d.Set("api_secret", apiSecret)
	// 		}
	// 	}
	// }

	d.SetId(userName)
	return diags
}

func resourceUserRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	userName := d.Get("user_name").(string)

	//checking the id of the group
	// userResp, err := user.GetUserIDByName(userName)
	// if err != nil {
	// 	log.Printf("get user id failed to get group, error %s", err.Error())
	// 	return diag.FromErr(err)
	// }
	if d.State() != nil && d.State().ID != "" {
		if userName != d.State().ID {
			log.Println("detected uername change ", userName, d.State().ID)
			userName = d.State().ID
		}
	}

	userAccount, err := user.GetUser(userName)
	if err != nil {
		log.Printf("get user account, error %s", err.Error())
		if strings.Contains(err.Error(), "does not exist") {
			log.Println("Resource Read ", "error", err)
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	log.Println("userAccount ", userAccount)

	err = flattenUser(d, userAccount)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func flattenUser(d *schema.ResourceData, in *models.UserResponse) error {
	if in == nil {
		return fmt.Errorf("%s", "failed to get user account(empty)")
	}

	if len(in.Account.Username) > 0 {
		err := d.Set("user_name", in.Account.Username)
		if err != nil {
			return err
		}
	}
	if len(in.Account.UserType) > 0 {
		if in.Account.UserType == "CONSOLE" {
			err := d.Set("console_access", true)
			if err != nil {
				return err
			}
		} else {
			err := d.Set("console_access", false)
			if err != nil {
				return err
			}
		}

	}
	if len(in.Account.FirstName) > 0 {
		err := d.Set("first_name", in.Account.FirstName)
		if err != nil {
			return err
		}
	}
	if len(in.Account.LastName) > 0 {
		err := d.Set("last_name", in.Account.LastName)
		if err != nil {
			return err
		}
	}
	if len(in.Account.Phone) > 0 {
		err := d.Set("phone", in.Account.Phone)
		if err != nil {
			return err
		}
	}

	grps, err := user.GetUserGroups(in.Account.Username)
	if err != nil {
		return err
	}
	log.Println("grps", len(grps))
	if len(grps) > 0 {
		err := d.Set("groups", grps)
		if err != nil {
			return err
		}
	}

	return nil
}

func resourceUserUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("resource user update id %s", d.Id())
	return resourceUserUpsert(ctx, d, false)
	//return diag.FromErr(fmt.Errorf("%s", "update not supported for user. Use group association to alter groups"))
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

func resourceUserImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	d.SetId(d.Id())
	d.Set("user_name", d.Id())
	log.Println("user_name", d.Id())
	return []*schema.ResourceData{d}, nil
}
