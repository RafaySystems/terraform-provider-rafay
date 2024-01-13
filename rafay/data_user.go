package rafay

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/RafaySystems/rctl/pkg/user"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataUser() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataUserRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(10 * time.Minute),
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

func dataUserRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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

	d.SetId(userName)
	return diags
}
