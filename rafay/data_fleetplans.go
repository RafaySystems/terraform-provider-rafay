package rafay

import (
	"context"
	"os"
	"time"

	"github.com/RafaySystems/rafay-common/pkg/hub/client/options"
	typed "github.com/RafaySystems/rafay-common/pkg/hub/client/typed"
	"github.com/RafaySystems/rafay-common/proto/types/hub/infrapb"
	"github.com/RafaySystems/rctl/pkg/config"
	"github.com/RafaySystems/rctl/pkg/versioninfo"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataFleetplans() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataFleetplansRead,
		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(10 * time.Minute), // 10 minutes
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"project": {
				Type:        schema.TypeString,
				Description: "Project name from where fleetplans to be listed",
				Required:    true,
			},
			"type": {
				Type:        schema.TypeString,
				Description: "Resource type of the fleet plan",
				Default:     "clusters",
				Optional:    true,
			},
			"fleetplans": {
				Type:        schema.TypeList,
				Description: "List of fleetplans",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"fleetplan_name": {
							Type:        schema.TypeString,
							Description: "Name of the fleet plan",
							Computed:    true,
						},
						"status": {
							Type:        schema.TypeList,
							Description: "Status of the fleet plan",
							Computed:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"job_status": {
										Type:        schema.TypeList,
										Description: "Job status of the fleet plan",
										Computed:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"status": {
													Type:        schema.TypeString,
													Description: "Status of the last job",
													Computed:    true,
												},
												"reason": {
													Type:        schema.TypeString,
													Description: "Reason for the last job status",
													Computed:    true,
												},
											},
										},
									},
									"schedule_status": {
										Type:        schema.TypeList,
										Description: "Schedule status of the fleet plan",
										Computed:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"name": {
													Type:        schema.TypeString,
													Description: "Name of the schedule",
													Computed:    true,
												},
												"last_run": {
													Type:        schema.TypeString,
													Description: "Last run timestamp of the schedule",
													Computed:    true,
												},
												"next_run": {
													Type:        schema.TypeString,
													Description: "Next run timestamp of the schedule",
													Computed:    true,
												},
												"times_opted_out": {
													Type:        schema.TypeInt,
													Description: "Number of times opted out of the schedule",
													Computed:    true,
												},
												"opted_out_duration": {
													Type:        schema.TypeString,
													Description: "Duration for which the schedule was last opted out",
													Computed:    true,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func dataFleetplansRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	fleetplanProject := d.Get("project").(string)
	resourceType := d.Get("type").(string)

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent(), options.WithInsecureSkipVerify(auth.SkipServerCertValid), options.WithConnectionTimeout(CONN_TIMEOUT))
	if err != nil {
		return diag.FromErr(err)
	}

	resp, err := client.InfraV3().FleetPlan().List(ctx, options.ListOptions{
		Project: fleetplanProject,
		Order:   resourceType,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	err = flattenFleetplans(d, resp.Items)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fleetplanProject)
	return diags
}

func flattenFleetplans(d *schema.ResourceData, items []*infrapb.FleetPlan) error {
	if err := d.Set("fleetplans", flattenFleetPlanItems(d, items)); err != nil {
		return err
	}
	return nil
}

func flattenFleetPlanItems(d *schema.ResourceData, items []*infrapb.FleetPlan) []interface{} {
	out := make([]interface{}, len(items))
	for i, item := range items {
		obj := map[string]interface{}{
			"fleetplan_name": item.Metadata.Name,
			"status":         flattenFleetPlanStatus(d, item.Status),
		}
		out[i] = obj
	}
	return out
}

func flattenFleetPlanStatus(d *schema.ResourceData, status *infrapb.FleetPlanJobStatus) []interface{} {
	if status == nil {
		return nil
	}
	obj := map[string]interface{}{}

	obj["job_status"] = flattenFleetPlanJobStatus(d, status.JobStatus)
	obj["schedule_status"] = flattenFleetPlanScheduleStatus(d, status.ScheduleStatus)

	return []interface{}{obj}
}

func flattenFleetPlanJobStatus(d *schema.ResourceData, status *infrapb.StatusObj) []interface{} {
	if status == nil {
		return nil
	}
	obj := map[string]interface{}{}

	obj["status"] = status.Status
	obj["reason"] = status.Reason

	return []interface{}{obj}
}

func flattenFleetPlanScheduleStatus(d *schema.ResourceData, status map[string]*infrapb.ScheduleStatus) []interface{} {
	if status == nil {
		return nil
	}
	obj := map[string]interface{}{}

	for k, v := range status {
		obj["name"] = k
		obj["last_run"] = v.LastRun.AsTime().String()
		obj["next_run"] = v.NextRun.AsTime().String()
		obj["times_opted_out"] = v.TimesOptedOut
		obj["opted_out_duration"] = v.OptedOutDuration
	}

	return []interface{}{obj}
}
