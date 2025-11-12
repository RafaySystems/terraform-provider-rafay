package rafay

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/RafaySystems/rctl/pkg/config"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataKubeConfig() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataKubeConfigRead,
		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(10 * time.Minute),
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
			"username": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"kubeconfig": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func dataKubeConfigUtil(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var err error
	log.Printf("download kube config starts")
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		_ = context.WithValue(ctx, "debug", "true")
	}
	auth := config.GetConfig().GetAppAuthProfile()

	defaultNamespace := d.Get("namespace").(string)
	cluster := d.Get("cluster").(string)

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
			log.Printf("cluser '%s' is not supported when username %s is given", cluster, username)
		}
		if defaultNamespace != "" {
			log.Printf("namespace '%s' is not supported when username %s  is given", defaultNamespace, username)
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

	if err := d.Set("kubeconfig", yaml); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("kubeconfig")
	return diags
}

func dataKubeConfigRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return dataKubeConfigUtil(ctx, d, m)
}
