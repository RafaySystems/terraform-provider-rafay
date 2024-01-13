package rafay

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/RafaySystems/rafay-common/pkg/hub/client/options"
	typed "github.com/RafaySystems/rafay-common/pkg/hub/client/typed"
	"github.com/RafaySystems/rafay-common/pkg/hub/terraform/resource"
	"github.com/RafaySystems/rctl/pkg/config"
	"github.com/RafaySystems/rctl/pkg/versioninfo"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	schema "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataGKEClusterV3() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataGKEClusterV3Read,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema:        resource.ClusterSchema.Schema,
	}
}

func dataGKEClusterV3Read(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	log.Printf("dataGKEClusterV3Read GKE")
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	meta := GetMetaData(d)
	if meta == nil {
		return diag.FromErr(fmt.Errorf("%s", "failed to read resource "))
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}

	ag, err := client.InfraV3().Cluster().Get(ctx, options.GetOptions{
		Name:    meta.Name,
		Project: meta.Project,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	err = flattenGKEClusterV3(d, ag)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(meta.Name)
	return diags

}
