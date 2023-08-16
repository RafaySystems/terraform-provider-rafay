package rafay

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/RafaySystems/rctl/pkg/config"
	"github.com/RafaySystems/rctl/pkg/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func downloadKubeConfig() *schema.Resource {
	return &schema.Resource{
		CreateContext: downloadKubeConfigCreate,
		UpdateContext: downloadKubeConfigUpdate,
		ReadContext:   downloadKubeConfigRead,
		DeleteContext: downloadKubeConfigDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"cluster": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"namespace": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"output_folder_path": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"filename": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"username": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func downloadKubeConfigCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return downloadKubeConfigUtil(ctx, d, m)
}

func downloadKubeConfigUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return downloadKubeConfigUtil(ctx, d, m)
}

func getUserDetails(username string) (accountId string, err error) {
	params := url.Values{}
	params.Add("q", username)
	uri := fmt.Sprintf("/auth/v1/users/?%s", params.Encode())
	auth := config.GetConfig().GetAppAuthProfile()

	log.Println("getUserDetails uri ", uri)
	resp, err := auth.AuthAndRequestFullResponse(uri, "GET", nil)
	if err != nil {
		log.Println("failed to get user details ", username, "resp", resp)
		return "", err
	}
	var usr models.UsersFullResponse
	if err := json.Unmarshal(resp.Bytes(), &usr); err != nil {
		log.Println("failed to get user details ", username, "resp", resp)
		return "", err
	}
	log.Println("download kubeconfig user getUserDetails:", usr, "resp", resp)

	if len(usr.Users) <= 0 {
		log.Println("failed to get user details got empty user", username, "resp", resp)
		return "", fmt.Errorf("error /auth/v1/users/ resp: %s", resp)
	}
	accountId = usr.Users[0].Account.ID
	return accountId, nil
}

func downloadKubeConfigUtil(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var err error
	log.Printf("download kube config starts")
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}
	auth := config.GetConfig().GetAppAuthProfile()

	defaultNamespace := d.Get("namespace").(string)
	cluster := d.Get("cluster").(string)
	filepath := ""
	if d.Get("output_folder_path").(string) != "" {
		filepath = d.Get("output_folder_path").(string)
	}

	filename := "kubeconfig-file"
	if d.Get("filename").(string) != "" {
		filename = d.Get("filename").(string)

	}

	username := ""
	accountID := ""
	if d.Get("username").(string) != "" {
		username = d.Get("username").(string)
		accountID, err = getUserDetails(username)
		if err != nil {
			log.Printf("failed to get kubeconfig for user %s", username)
			return diags
		}
	}

	if accountID != "" && (defaultNamespace != "" || cluster != "") {
		if cluster != "" {
			log.Printf("cluser '%s' is not suppoerted when username %s is given", cluster, username)
		}
		if defaultNamespace != "" {
			log.Printf("namespace '%s' is not suppoerted when username %s  is given", defaultNamespace, username)
		}
		return diags
	}

	params := url.Values{}
	if defaultNamespace != "" {
		params.Add("namespace", defaultNamespace)
	}
	if cluster != "" {
		params.Add("opts.selector", fmt.Sprintf("rafay.dev/clusterName=%s", cluster))
	}

	uri := ""
	if accountID != "" {
		uri = fmt.Sprintf("/v2/sentry/kubeconfig/user/%s/download", accountID)
	} else {
		uri = fmt.Sprintf("/v2/sentry/kubeconfig/user?%s", params.Encode())
	}

	resp, err := auth.AuthAndRequestFullResponse(uri, "GET", nil)
	if err != nil {
		log.Printf("failed to get kubeconfig")
	}

	jsonData := &struct {
		Data string `json:"data"`
	}{}

	err = resp.JSON(jsonData)
	if err != nil {
		log.Println("failed to unmarshal kubeconfig jsonData error", err)
	}

	decoded, err := base64.StdEncoding.DecodeString(jsonData.Data)
	if err != nil {
		log.Println("failed to decode kubeconfig error", err)
	}
	yaml := string(decoded)

	fileLocation := filepath + "/" + filename
	err = ioutil.WriteFile(fileLocation, []byte(yaml), 0644)
	if err != nil {
		log.Printf("Failed to store the downloaded kubeconfig file ")
	}
	fmt.Printf("kubeconfig downloaded to file location - %s", fileLocation)

	d.SetId(fileLocation)
	return diags
}

func downloadKubeConfigRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	return diags
}

func downloadKubeConfigDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}
	os.Remove(d.Id())
	return diags
}
