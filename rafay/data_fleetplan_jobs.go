package rafay

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/RafaySystems/rafay-common/pkg/hub/client/options"
	typed "github.com/RafaySystems/rafay-common/pkg/hub/client/typed"
	"github.com/RafaySystems/rafay-common/proto/types/hub/infrapb"
	"github.com/RafaySystems/rctl/pkg/config"
	"github.com/RafaySystems/rctl/pkg/versioninfo"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataFleetplanJobs() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataFleetplanJobsRead,
		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"project": {
				Description: "Project name from where environments to be listed",
				Type:        schema.TypeString,
				Required:    true,
			},
			"fleetplan_name": {
				Description: "FleetPlan name",
				Type:        schema.TypeString,
				Required:    true,
			},
			"jobs": {
				Type:        schema.TypeList,
				Description: "List of fleetplan jobs",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"job_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "name of the job",
						},
						"created_at": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "creation timestamp of the job",
						},
						"execution_duration": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "execution duration of the job",
						},
						"resource_count": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "number of resources used by the job",
						},
						"state": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "current state of the job",
						},
						"reason": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "reason for the job's current state",
						},
						"workflow_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the workflow",
						},
					},
				},
			},
		},
	}
}

func dataFleetplanJobsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	fleetplanProject := d.Get("project").(string)
	fleetplanName := d.Get("fleetplan_name").(string)

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent(), options.WithInsecureSkipVerify(auth.SkipServerCertValid), options.WithConnectionTimeout(CONN_TIMEOUT))
	if err != nil {
		return diag.FromErr(err)
	}

	resp, err := client.InfraV3().FleetPlan().ExtApi().GetJobs(ctx, options.ExtOptions{
		Name:    fleetplanName,
		Project: fleetplanProject,
	})
	if err != nil {
		if strings.Contains(err.Error(), "code 404") {
			log.Println("Resource Read ", "error", err)
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	jobList := &infrapb.FleetPlanJobList{}
	err = json.Unmarshal(resp.Body, jobList)
	if err != nil {
		return diag.FromErr(err)
	}

	var fleetplanId string
	fleetplanJobs := make([]map[string]interface{}, len(jobList.Items))
	for i, job := range jobList.Items {
		fleetplanJobs[i] = map[string]interface{}{
			"job_name":           job.Metadata.Name,
			"created_at":         job.Metadata.CreatedAt.AsTime().String(),
			"execution_duration": job.ExecutionDuration,
			"resource_count":     job.ResourceCount,
			"state":              job.State,
			"reason":             job.Reason,
			"workflow_id":        job.WorkflowId,
		}
		fleetplanId = job.FleetPlanId
	}

	if err := d.Set("jobs", fleetplanJobs); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(fleetplanId)
	return diags
}
