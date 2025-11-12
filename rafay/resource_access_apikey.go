package rafay

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/RafaySystems/rctl/pkg/commands"
	"github.com/RafaySystems/rctl/pkg/config"
	"github.com/RafaySystems/rctl/pkg/exit"
	"github.com/RafaySystems/rctl/pkg/models"
	"github.com/RafaySystems/rctl/pkg/user"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type Role struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	IsGlobal        bool   `json:"is_global"`
	Scope           string `json:"scope"`
	NamespaceIDList any    `json:"namespace_id_list"`
	RoleType        string `json:"role_type"`
}

type Group struct {
	ID         string `json:"id"`
	Type       string `json:"type"`
	Name       string `json:"name"`
	CreatedAt  any    `json:"created_at"`
	ModifiedAt any    `json:"modified_at"`
}

type Roles []struct {
	Project      any   `json:"project"`
	Roles        Role  `json:"role"`
	Group        Group `json:"group"`
	NamespaceAgg any   `json:"namespace_agg"`
	Organization any   `json:"organization"`
	Partner      any   `json:"partner"`
}

type Account struct {
	ID            string `json:"id"`
	CreatedAt     string `json:"createdAt"`
	ModifiedAt    string `json:"modifiedAt"`
	LastLogin     string `json:"last_login"`
	Username      string `json:"username"`
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
	EmailVerified bool   `json:"emailVerified"`
	UserType      string `json:"user_type"`
	Password      string `json:"password"`
	IsSso         bool   `json:"is_sso"`
}

type UserProfile struct {
	Roles        Roles   `json:"roles"`
	Account      Account `json:"account"`
	Organization any     `json:"organization"`
	Partner      any     `json:"partner"`
}

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
	usrProf, err := checkConfigRole(ctx, userName)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Println("resourceAccessApiUpsert ", userName)

	if d.State() != nil && d.State().ID != "" {
		if userName != "" && userName != d.State().ID {
			log.Printf("username change not supported")
			d.State().Tainted = true
			return diag.FromErr(fmt.Errorf("%s", "username change not supported"))
		}
	}

	if usrProf != nil {
		log.Println("resourceAccessApiUpsert ", userName, "Account", usrProf.Account)
		api, secret, err = createUserAPIKeyByID(userName, usrProf.Account.ID)
	} else {
		log.Println("resourceAccessApiUpsert ", userName)
		api, secret, err = commands.CreateUserAPIKey(userName)
	}
	if err != nil {
		log.Println("resourceAccessApiUpsert error", err)
		return diag.FromErr(err)
	}

	log.Println("resourceAccessApiUpsert  len(api) ", len(api), len(secret))

	if len(api) > 0 {
		if err := d.Set("apikey", api); err != nil {
			return diag.FromErr(err)
		}
	}
	if len(secret) > 0 {
		if err := d.Set("api_secret", secret); err != nil {
			return diag.FromErr(err)
		}
	}

	d.SetId(userName)
	return diags
}

func resourceAccessApiRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var found bool
	var userAccount models.UserResponse

	apikey := d.Get("apikey").(string)

	s := strings.Split(apikey, ".")
	if len(s) > 1 && s[0] == "ra2" {
		apikey = s[0] + "." + s[1]
	}

	userName := d.Get("user_name").(string)
	usrProf, err := checkConfigRole(ctx, userName)
	if err != nil {
		return diag.FromErr(err)
	}
	if usrProf != nil {
		userAccount.Account.Username = userName
		userAccount.Account.ID = usrProf.Account.ID
	} else {
		ua, err := user.GetUser(userName)
		if err != nil {
			log.Printf("get user account, error %s", err.Error())
			return diag.FromErr(err)
		}
		userAccount = *ua
		userName = userAccount.Account.Username
	}

	log.Println("resourceAccessApiRead ", userName, " apikey ", len(apikey))

	if d.State() != nil && d.State().ID != "" {
		if userName != d.State().ID {
			log.Println("detected uername change ", userName, d.State().ID)
			userName = d.State().ID
		}
	}

	log.Println("userAccount ", userAccount)

	if len(apikey) > 0 {
		var apikeys []models.UserAPIKeyStatus
		// there is an api key in the state. check key exist in controller
		if usrProf != nil {
			// get current user api keys
			apikeys, err = getUserAPIKeysByID(userName, usrProf.Account.ID)
		} else {
			apikeys, err = user.GetUserAPIKeys(userName)
		}
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

	err = flattenAccessApi(d, &userAccount, apikey)
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
		if err := d.Set("apikey", "use 'terraform apply -replace=resource-name' to recreate"); err != nil {
			return err
		}
		if err := d.Set("api_secret", "use 'terraform apply -replace=resource-name' to recreate"); err != nil {
			return err
		}
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
	usrProf, err := checkConfigRole(ctx, userName)
	if err != nil {
		return diag.FromErr(err)
	}

	apiKey := d.Get("apikey").(string)

	s := strings.Split(apiKey, ".")
	if len(s) > 1 && s[0] == "ra2" {
		apiKey = s[0] + "." + s[1]
	}

	log.Printf("resource user delete id %s", userName)
	if usrProf != nil {
		err = deleteUserAPIKeyByUserID(userName, usrProf.Account.ID, apiKey)
	} else {
		err = commands.DeleteUserAPIKey(userName, apiKey)
	}

	if err != nil {
		log.Printf("delete apikey error %s", err.Error())
		return diag.FromErr(err)
	}

	return diags
}

func resourceAccessApiGetCurrentUser(ctx context.Context) (*UserProfile, diag.Diagnostics, error) {
	var diags diag.Diagnostics

	// Get current user profile based on API key
	auth := config.GetConfig().GetAppAuthProfile()

	uri := "/auth/v1/users/-/profile"
	resp, err := auth.AuthAndRequest(uri, "GET", nil)
	if err != nil {
		log.Println("get user profile uri", uri, "error", err)
		return nil, diag.FromErr(err), err
	}

	var respGetProfile UserProfile
	if err := json.Unmarshal([]byte(resp), &respGetProfile); err != nil {
		return nil, diag.FromErr(err), err
	}

	log.Println("get user profile", respGetProfile)

	return &respGetProfile, diags, nil
}

func getUserAPIKeysByID(userName, id string) ([]models.UserAPIKeyStatus, error) {
	var userApiKeys []models.UserAPIKeyStatus

	auth := config.GetConfig().GetAppAuthProfile()
	uriAPIUrl := fmt.Sprintf("/auth/v1/users/%s/apikeys/", id)
	resp, err := auth.AuthAndRequest(uriAPIUrl, "GET", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get user %s %v", userName, err)
	}

	errVal := json.Unmarshal([]byte(resp), &userApiKeys)
	if errVal != nil {
		exit.SetExitWithError(err, fmt.Sprintf("Internal CLI error. Error: %s", errVal))
		return nil, fmt.Errorf("failed to get user %s apikeys %v", userName, errVal)
	}

	if len(userApiKeys) <= 0 {
		return nil, fmt.Errorf("empty api key for user %s", userName)
	}

	return userApiKeys, nil
}

func deleteUserAPIKeyByUserID(username, id, apikey string) error {
	auth := config.GetConfig().GetAppAuthProfile()

	userApiKeys, err := getUserAPIKeysByID(username, id)
	if err != nil {
		return fmt.Errorf("failed to get api keys for user %s err %v", username, err)
	}

	for _, ak := range userApiKeys {
		if apikey == ak.Key {
			uriAPIUrl := fmt.Sprintf("/auth/v1/apikeys/%s/", ak.ID)
			_, err := auth.AuthAndRequest(uriAPIUrl, "DELETE", uriAPIUrl)
			if err != nil {
				return fmt.Errorf("failed to delete apikey for user %s error %v", username, err)
			}
			return nil
		}
	}

	return nil
}

func createUserAPIKeyByID(username, userId string) (string, string, error) {
	var userApiKey models.UserAPIKeyStatus

	uriAPIUrl := fmt.Sprintf("/auth/v1/users/%s/apikey/", userId)
	postApiKeyVal := commands.NewAPIKeyPost{
		Name: "dynamic",
	}

	auth := config.GetConfig().GetAppAuthProfile()
	resp, err := auth.AuthAndRequest(uriAPIUrl, "POST", postApiKeyVal)
	if err != nil {
		log.Println("create user api key error", err)
		return "", "", fmt.Errorf("user %s api key create error %v", username, err)
	}

	err = json.Unmarshal([]byte(resp), &userApiKey)
	if err != nil {
		return "", "", fmt.Errorf("user %s api creation failed %v", username, err)
	}

	return userApiKey.Key, userApiKey.Secret, nil

}

func checkConfigRole(ctx context.Context, username string) (*UserProfile, error) {
	_, err := user.GetUserIDByName(username)
	if err == nil {
		// has permission to manage other users
		return nil, nil
	}

	usrProf, _, err := resourceAccessApiGetCurrentUser(ctx)
	if err != nil {
		return nil, err
	}

	if usrProf == nil {
		return nil, fmt.Errorf("failed to get current user %s", username)
	}

	if usrProf.Account.Username != username {
		return nil, fmt.Errorf("forbidden: failed to manage user %s", username)
	}
	return usrProf, nil
}
