package rafay

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
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

func resourceFleetPlanTrigger() *schema.Resource {
	return &schema.Resource{
		CreateContext: createFleetPlanJob,
		ReadContext:   readFleetPlanJob,
		UpdateContext: updateFleetPlanJob,
		DeleteContext: deleteFleetPlanJob,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(2 * time.Hour),   // 2 hours
			Update: schema.DefaultTimeout(2 * time.Hour),   // 2 hour
			Delete: schema.DefaultTimeout(2 * time.Minute), // 2 minutes
		},
		Schema: map[string]*schema.Schema{
			"fleetplan_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "FleetPlan name",
			},
			"project": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "FleetPlan project",
			},
			"trigger_value": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Enter trigger value to trigger a new job for fleetplan",
			},
		},
	}
}

func createFleetPlanJob(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Println("Creating FleetPlan job")
	return upsertFleetPlanJob(ctx, d)
}

func updateFleetPlanJob(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Println("Updating FleetPlan job")
	return upsertFleetPlanJob(ctx, d)
}

func upsertFleetPlanJob(ctx context.Context, d *schema.ResourceData) diag.Diagnostics {
	log.Println("fleetplan job upsert starts..")
	var diags diag.Diagnostics
	// Implement the logic to upsert a fleet plan job
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	fleetPlanName := d.Get("fleetplan_name").(string)
	fleetPlanProject := d.Get("project").(string)

	log.Printf("upserting fleetplan job for fleetplan: %s, project: %s", fleetPlanName, fleetPlanProject)

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent(), options.WithInsecureSkipVerify(auth.SkipServerCertValid), options.WithConnectionTimeout(CONN_TIMEOUT))
	if err != nil {
		return diag.FromErr(err)
	}

	response, err := client.InfraV3().FleetPlan().ExtApi().ExecuteFleetPlan(ctx, options.ExtOptions{
		Name:    fleetPlanName,
		Project: fleetPlanProject,
	})

	if err != nil {
		return diag.FromErr(err)
	}

	job := &infrapb.FleetPlanJob{}
	err = json.Unmarshal(response.Body, job)
	if err != nil {
		return diag.FromErr(err)
	}

	if job.Metadata != nil && job.Metadata.ID != "" {
		d.SetId(job.Metadata.ID)
	}

	// poll job status
	fleetPlanJobName := job.Metadata.Name
	ticker := time.NewTicker(time.Duration(60) * time.Second)
	defer ticker.Stop()

LOOP:
	for {
		select {
		case <-ctx.Done():
			log.Printf("fleet plan: %s job execution timed out", fleetPlanName)
			return diag.FromErr(fmt.Errorf("fleet plan: %s job execution timed out", fleetPlanName))

		case <-ticker.C:
			response, err := client.InfraV3().FleetPlan().ExtApi().GetJobStatus(ctx, options.ExtOptions{
				Name:    fleetPlanName,
				Project: fleetPlanProject,
				UrlParams: map[string]string{
					"job_name": fleetPlanJobName,
				},
			})
			if err != nil {
				return diag.FromErr(err)
			}

			jobStatus := &infrapb.FleetPlanJobStatus{}
			err = json.Unmarshal(response.Body, jobStatus)
			if err != nil {
				return diag.FromErr(err)
			}

			switch jobStatus.GetJobStatus().GetStatus() {
			case "skipped":
				log.Printf("fleet plan: %s job: %s is skipped\n", fleetPlanName, fleetPlanJobName)
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Warning,
					Summary:  "FleetPlan Job Skipped",
					Detail:   jobStatus.GetJobStatus().GetReason(),
				})
				break LOOP
			case "fail":
				log.Printf("fleet plan: %s job: %s has failed\n", fleetPlanName, fleetPlanJobName)
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Warning,
					Summary:  "FleetPlan Job Failed",
					Detail:   jobStatus.GetJobStatus().GetReason(),
				})
				break LOOP
			case "completed_with_failures":
				log.Printf("fleet plan: %s job: %s is completed with failures\n", fleetPlanName, fleetPlanJobName)
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Warning,
					Summary:  "FleetPlan Job completed with failures",
					Detail:   jobStatus.GetJobStatus().GetReason(),
				})
				break LOOP
			case "success":
				log.Printf("fleet plan: %s job: %s is successful\n", fleetPlanName, fleetPlanJobName)
				break LOOP
			case "cancelled":
				log.Printf("fleet plan: %s job: %s is cancelled\n", fleetPlanName, fleetPlanJobName)
				break LOOP
			default:
				log.Printf("fleet plan: %s job: %s is still running\n", fleetPlanName, fleetPlanJobName)
			}
		}
	}

	log.Println("fleetplan job upsert ends..")
	return diags
}

func readFleetPlanJob(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Implement the logic to read a fleet plan job
	log.Println("read aks workload identity")

	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	fleetPlanName := d.Get("fleetplan_name").(string)
	fleetPlanProject := d.Get("project").(string)

	// Implement the logic to read a fleet plan job
	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent(), options.WithInsecureSkipVerify(auth.SkipServerCertValid), options.WithConnectionTimeout(CONN_TIMEOUT))
	if err != nil {
		return diag.FromErr(err)
	}

	response, err := client.InfraV3().FleetPlan().ExtApi().GetJobs(ctx, options.ExtOptions{
		Name:    fleetPlanName,
		Project: fleetPlanProject,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	jobList := &infrapb.FleetPlanJobList{}
	err = json.Unmarshal(response.Body, jobList)
	if err != nil {
		return diag.FromErr(err)
	}

	if jobList != nil && len(jobList.Items) > 0 {
		latestJobName := jobList.Items[0].Metadata.DisplayName
		d.Set("trigger_value", latestJobName)
	}

	return diags
}

func deleteFleetPlanJob(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Implement the logic to delete a fleet plan job
	log.Println("Delete operation is not supported for fleet plan job. Terraform destroy operation will only remove the fleetplan_job resource from stored state.")

	return diags
}
