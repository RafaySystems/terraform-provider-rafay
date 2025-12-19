package rafay

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/RafaySystems/rafay-common/pkg/hub/client/options"
	typed "github.com/RafaySystems/rafay-common/pkg/hub/client/typed"
	"github.com/RafaySystems/rafay-common/pkg/hub/terraform/resource"
	"github.com/RafaySystems/rctl/pkg/config"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataBluePrint() *schema.Resource {
	s := copySchemaMap(resource.BlueprintSchema.Schema)
	return &schema.Resource{
		ReadContext: dataBluePrintRead,
		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(10 * time.Minute),
		},
		SchemaVersion: 1,
		Schema:        s,
	}
}

func dataBluePrintRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	log.Println("dataBluePrintRead ")
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	meta := GetMetaData(d)
	if meta == nil {
		return diag.FromErr(fmt.Errorf("%s", "failed to read resource "))
	}
	if d.State() != nil && d.State().ID != "" {
		meta.Name = d.State().ID
	}

	// XXX Debug
	// w1 := spew.Sprintf("%+v", tfBlueprintState)
	// log.Println("dataBluePrintRead tfBlueprintState", w1)

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, TF_USER_AGENT, options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return diag.FromErr(err)
	}

	bp, err := client.InfraV3().Blueprint().Get(ctx, options.GetOptions{
		Name:    meta.Name,
		Project: meta.Project,
	})
	if err != nil {
		if strings.Contains(err.Error(), "code 404") {
			log.Println("Resource Read ", "error", err)
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	// XXX Debug
	// w1 = spew.Sprintf("%+v", wl)
	// log.Println("dataBluePrintRead wl", w1)

	err = flattenBlueprint(d, bp)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(bp.Metadata.Name)

	return diags

}
