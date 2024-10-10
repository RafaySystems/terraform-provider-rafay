package rafay

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/RafaySystems/rafay-common/pkg/hub/client/options"
	typed "github.com/RafaySystems/rafay-common/pkg/hub/client/typed"
	"github.com/RafaySystems/rctl/pkg/config"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataNamespaces() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataNamespacesRead,
		Schema: map[string]*schema.Schema{
			"metadata": &schema.Schema{
				Description: "Metadata of the namespace resource",
				Elem: &schema.Resource{Schema: map[string]*schema.Schema{
					"project": &schema.Schema{
						Description: "Project of the resource",
						Optional:    true,
						Type:        schema.TypeString,
					},
				}},
				MaxItems: 1,
				MinItems: 1,
				Optional: true,
				Type:     schema.TypeList,
			},
			"namespaces": &schema.Schema{
				Description: "Specification of the namespace resource",
				Elem: &schema.Resource{Schema: map[string]*schema.Schema{
					"name": &schema.Schema{
						Description: "data is the base64 encoded contents of the file",
						Optional:    true,
						Type:        schema.TypeString,
					},
				}},
				Optional: true,
				Type:     schema.TypeList,
			},
		},
	}
}

func dataNamespacesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	log.Println("dataNamespacesRead ")
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	meta := GetMetaData(d)
	if meta == nil {
		return diag.FromErr(fmt.Errorf("%s", "failed to read resource "))
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, TF_USER_AGENT, options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return diag.FromErr(err)
	}

	namespaces, err := client.InfraV3().Namespace().List(ctx, options.ListOptions{
		Project: meta.Project,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	var namespaceList []map[string]interface{}

	for _, ns := range namespaces.Items {
		nsData := map[string]interface{}{
			"name": ns.Metadata.Name,
		}
		namespaceList = append(namespaceList, nsData)
	}

	d.Set("namespaces", namespaceList)
	d.SetId("All")
	return diags

}
