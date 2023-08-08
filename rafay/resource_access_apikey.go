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

func resourceAccessApikey() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAccessApiCreate,
		ReadContext:   resourceAccessApiRead,
		UpdateContext: resourceAccessApiUpdate,
		DeleteContext: resourceAccessApiDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"user_name": {
				Description: "User name for the API Keys that allow to interact with the system via the RESTful API exposed by the platform.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"apikey": {
				Description: "The API Keys that allow to interact with the system",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
				Sensitive:   true,
			},
			"api_secret": {
				Description: "The API secret that allow to interact with the system",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
				Sensitive:   true,
			},
		},
	}
}

func resourceAccessApiCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("resource user create id %s", d.Id())
	return resourceAccessApiUpsert(ctx, d, true)
}

func resourceAccessApiUpsert(ctx context.Context, d *schema.ResourceData, create bool) diag.Diagnostics {
	var diags diag.Diagnostics
	var err error
	var api, secret string

	userName := d.Get("user_name").(string)

	log.Println("resourceAccessApiUpsert ", userName)

	if d.State() != nil && d.State().ID != "" {
		if userName != "" && userName != d.State().ID {
			log.Printf("username change not supported")
			d.State().Tainted = true
			return diag.FromErr(fmt.Errorf("%s", "username change not supported"))
		}
	}

	log.Println("resourceAccessApiUpsert ", userName)
	api, secret, err = commands.CreateUserAPIKey(userName)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Println("resourceAccessApiUpsert  len(api) ", len(api), len(secret))

	if len(api) > 0 {
		d.Set("apikey", api)
	}
	if len(secret) > 0 {
		d.Set("api_secret", secret)
	}

	d.SetId(userName)
	return diags
}

func resourceAccessApiRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var found bool

	userName := d.Get("user_name").(string)
	apikey := d.Get("apikey").(string)

	s := strings.Split(apikey, ".")
	if len(s) > 1 && s[0] == "ra2" {
		apikey = s[0] + "." + s[1]
	}

	log.Println("resourceAccessApiRead ", userName, " apikey ", len(apikey))

	if d.State() != nil && d.State().ID != "" {
		if userName != d.State().ID {
			log.Println("detected uername change ", userName, d.State().ID)
			userName = d.State().ID
		}
	}

	userAccount, err := user.GetUser(userName)
	if err != nil {
		log.Printf("get user account, error %s", err.Error())
		return diag.FromErr(err)
	}

	log.Println("userAccount ", userAccount)

	if len(apikey) > 0 {
		// there is an api key in the state. check key exist in controller
		apikeys, err := user.GetUserAPIKeys(userName)
		if err != nil {
			log.Println("resourceAccessApiRead ", "error", err)
			found = false
		} else {
			for _, ak := range apikeys {
				if ak.Key == apikey {
					found = true
				}
			}
		}
	}

	if !found {
		// apikey is not found set to empty to reflect the state
		apikey = ""
	}

	err = flattenAccessApi(d, userAccount, apikey)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func flattenAccessApi(d *schema.ResourceData, in *models.UserResponse, api string) error {
	if in == nil {
		return fmt.Errorf("%s", "failed to get user account(empty)")
	}

	// if len(in.Account.Username) > 0 {
	// 	err := d.Set("user_name", in.Account.Username)
	// 	if err != nil {
	// 		return err
	// 	}
	// }

	if len(api) <= 0 {
		d.Set("apikey", "use 'terraform apply -replace=resource-name' to recreate")
		d.Set("api_secret", "use 'terraform apply -replace=resource-name' to recreate")
		return nil
	}

	// err := d.Set("apikey", api)
	// if err != nil {
	// 	return err
	// }
	return nil
}

func resourceAccessApiUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("resource user update id %s", d.Id())
	return resourceAccessApiUpsert(ctx, d, false)
	//return diag.FromErr(fmt.Errorf("%s", "update not supported for user. Use group association to alter groups"))
}

func resourceAccessApiDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	userName := d.Get("user_name").(string)
	apiKey := d.Get("apikey").(string)

	log.Printf("resource user delete id %s", userName)
	err := commands.DeleteUserAPIKey(userName, apiKey)
	if err != nil {
		log.Printf("delete apikey error %s", err.Error())
		return diag.FromErr(err)
	}

	return diags
}
