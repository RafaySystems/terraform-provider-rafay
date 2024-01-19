package rafay

import (
	"context"
	"log"
	"time"

	"github.com/RafaySystems/rctl/pkg/group"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataGroup() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataGroupRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func dataGroupRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	resp, err := group.GetGroupByName(d.Get("name").(string))
	if err != nil {
		log.Printf("create group failed to get group, error %s", err.Error())
		return diag.FromErr(err)
	}

	g, err := group.NewGroupFromResponse([]byte(resp))
	if err != nil {
		log.Printf("create group failed to parse get response, error %s", err.Error())
		return diag.FromErr(err)
	} else if g == nil {
		log.Printf("create group failed to parse get response")
		d.SetId("")
		return diags
	}

	log.Printf("resource group created %s", g.ID)
	if err := d.Set("name", g.Name); err != nil {
		log.Printf("get group set name error %s", err.Error())
		return diag.FromErr(err)
	}

	d.SetId(g.ID)

	return diags
}
