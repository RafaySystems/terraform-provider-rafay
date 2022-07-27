package rafay

import (
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/RafaySystems/rctl/pkg/config"
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
		},
	}
}

func downloadKubeConfigCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("agent create starts")
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}
	auth := config.GetConfig().GetAppAuthProfile()

	defaultNamespace := d.Get("namespace").(string)
	cluster := d.Get("cluster").(string)
	params := url.Values{}
	if defaultNamespace != "" {
		params.Add("namespace", defaultNamespace)
	}
	if cluster != "" {
		params.Add("opts.selector", fmt.Sprintf("rafay.dev/clusterName=%s", cluster))
	}

	uri := fmt.Sprintf("/v2/sentry/kubeconfig/user?%s", params.Encode())
	resp, err := auth.AuthAndRequestFullResponse(uri, "GET", nil)
	if err != nil {
		log.Printf("failed to get kubeconfig")
	}

	jsonData := &struct {
		Data string `json:"data"`
	}{}

	err = resp.JSON(jsonData)
	if err != nil {
		log.Printf("failed to get kubeconfig")
	}

	decoded, err := base64.StdEncoding.DecodeString(jsonData.Data)
	if err != nil {
		log.Printf("failed to get kubeconfig")
	}
	yaml := string(decoded)

	toFile := "kubeconfig-file"
	err = ioutil.WriteFile(toFile, []byte(yaml), 0644)
	if err != nil {
		log.Printf("Failed to store the downloaded kubeconfig file ")
	}
	fmt.Println(fmt.Sprintf("kubeconfig downloaded to file - %s", toFile))

	d.SetId(yaml)
	return diags

}

func downloadKubeConfigUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
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

	return diags
}

func upsertDownload(namespace string, clusterName string, d *schema.ResourceData) error {

	return nil
}
