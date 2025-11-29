package rafay

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/RafaySystems/rctl/pkg/config"
	"github.com/RafaySystems/rctl/pkg/user"

	"github.com/RafaySystems/rafay-common/pkg/hub/client/options"
	typed "github.com/RafaySystems/rafay-common/pkg/hub/client/typed"
	"github.com/RafaySystems/rafay-common/pkg/hub/terraform/resource"
	"github.com/RafaySystems/rafay-common/proto/types/hub/commonpb"
	"github.com/RafaySystems/rctl/pkg/versioninfo"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataWorkload() *schema.Resource {
	modSchema := resource.WorkloadSchema.Schema
	modSchema["impersonate"] = &schema.Schema{
		Description: "impersonate user",
		Optional:    true,
		Type:        schema.TypeString,
	}
	modSchema["condition"] = &schema.Schema{
		Description: "workload condition",
		Optional:    true,
		Type:        schema.TypeString,
	}
	modSchema["condition_status"] = &schema.Schema{
		Description: "workload status",
		Optional:    true,
		Type:        schema.TypeString,
	}
	modSchema["reason"] = &schema.Schema{
		Description: "workload reason",
		Optional:    true,
		Type:        schema.TypeString,
	}
	// modSchema["clusters"] = &schema.Schema{
	// 	Description: "workload deployed clusters",
	// 	Optional:    true,
	// 	Type:        schema.TypeList,
	// 	Elem: &schema.Schema{
	// 		Type: schema.TypeString,
	// 	},
	// }

	return &schema.Resource{
		ReadContext: dataWorkloadRead,
		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema:        modSchema,
	}
}

func dataWorkloadRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	log.Println("dataWorkloadRead ")

	meta := GetMetaData(d)
	if meta == nil {
		return diag.FromErr(fmt.Errorf("%s", "failed to read resource "))
	}

	// XXX Debug
	// w1 := spew.Sprintf("%+v", tfWorkloadState)
	// log.Println("dataWorkloadRead tfWorkloadState", w1)

	if v, ok := d.Get("impersonate").(string); ok && len(v) > 0 {
		defer ResetImpersonateUser()
		asUser := d.Get("impersonate").(string)
		// check user role : impersonation not allowed for a user
		// with ORG Admin role
		isOrgAdmin, err := user.IsOrgAdmin(asUser)
		if err != nil {
			return diag.FromErr(err)
		}
		if isOrgAdmin {
			return diag.FromErr(fmt.Errorf("%s", "--as-user cannot have ORGADMIN role"))
		}
		config.ApiKey, config.ApiSecret, err = user.GetUserAPIKey(asUser)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent(), options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return diag.FromErr(err)
	}

	wl, err := client.AppsV3().Workload().Get(ctx, options.GetOptions{
		Name:    meta.Name,
		Project: meta.Project,
	})
	if err != nil {
		if strings.Contains(err.Error(), "request code 404") {
			var ret []interface{}
			err = d.Set("metadata", ret)
			if err != nil {
				return diags
			}
			err = d.Set("spec", ret)
			if err != nil {
				return diags
			}
			return diags
		}
		return diag.FromErr(err)
	}

	// XXX Debug
	// w1 = spew.Sprintf("%+v", wl)
	// log.Println("dataWorkloadRead wl", w1)

	err = flattenWorkload(d, wl, true)
	if err != nil {
		return diag.FromErr(err)
	}

	wls, err := client.AppsV3().Workload().Status(ctx, options.StatusOptions{
		Name:    wl.Metadata.Name,
		Project: wl.Metadata.Project,
	})
	if err != nil {
		return diag.FromErr(err)
	}
	log.Println("wls.Status", wls.Status)
	status := "unknown"
	if wls.Status != nil {
		if wls.Status.ConditionStatus == commonpb.ConditionStatus_StatusOK {
			status = "ok"
		} else if wls.Status.ConditionStatus == commonpb.ConditionStatus_StatusFailed {
			status = "failed"
		} else if wls.Status.ConditionStatus == commonpb.ConditionStatus_StatusNotSet {
			status = "notset"
		} else if wls.Status.ConditionStatus == commonpb.ConditionStatus_StatusSubmitted {
			status = "submitted"
		}
		if err := d.Set("condition", wls.Status.ConditionType); err != nil {
			log.Println("failed to set condition error", err)
		}
		if err := d.Set("condition_status", status); err != nil {
			log.Println("failed to set condition_status error", err)
		}
		if err := d.Set("reason", wls.Status.Reason); err != nil {
			log.Println("failed to set reason error", err)
		}
		// d.Set("clusters", wls.Status.DeployedClusters)
	}
	if err := d.Set("condition_status", status); err != nil {
		log.Println("failed to set condition_status error", err)
	}
	d.SetId(wl.Metadata.Name)

	return diags

}
