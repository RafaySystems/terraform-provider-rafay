package rafay

import (
	"context"
	"encoding/json"
	"log"
	"os"
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

func dataFleetplanJob() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataFleetplanJobRead,
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
			"name": {
				Description: "FleetPlan job name",
				Type:        schema.TypeString,
				Required:    true,
			},
			"status": {
				Description: "Fleetplan job",
				Computed:    true,
				Type:        schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Description: "Fleet plan job name",
							Computed:    true,
						},
						"status": {
							Type:        schema.TypeString,
							Description: "Fleet plan job status",
							Computed:    true,
						},
						"reason": {
							Type:        schema.TypeString,
							Description: "Fleet plan job reason",
							Computed:    true,
						},
						"targets": {
							Type:        schema.TypeList,
							Description: "Fleet plan job targets",
							Computed:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:        schema.TypeString,
										Description: "Target name",
										Computed:    true,
									},
									"project": {
										Type:        schema.TypeString,
										Description: "Target project",
										Computed:    true,
									},
									"operations": {
										Type:        schema.TypeList,
										Description: "Target operations",
										Computed:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"name": {
													Type:        schema.TypeString,
													Description: "Operation name",
													Computed:    true,
												},
												"status": {
													Type:        schema.TypeString,
													Description: "Operation status",
													Computed:    true,
												},
												"reason": {
													Type:        schema.TypeString,
													Description: "Operation reason",
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

func dataFleetplanJobRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	fleetplanProject := d.Get("project").(string)
	fleetplanName := d.Get("fleetplan_name").(string)
	fleetplanJobName := d.Get("name").(string)

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent(), options.WithInsecureSkipVerify(auth.SkipServerCertValid), options.WithConnectionTimeout(CONN_TIMEOUT))
	if err != nil {
		return diag.FromErr(err)
	}

	resp, err := client.InfraV3().FleetPlan().ExtApi().GetJobStatus(ctx, options.ExtOptions{
		Name:    fleetplanName,
		Project: fleetplanProject,
		UrlParams: map[string]string{
			"job_name": fleetplanJobName,
		},
	})
	if err != nil {
		log.Println("Resource Read ", "error", err)
		if strings.Contains(err.Error(), "code 404") {
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	jobStatus := &infrapb.FleetPlanJobStatus{}
	err = json.Unmarshal(resp.Body, jobStatus)
	if err != nil {
		log.Println("Failed to read response body", "body", resp.Body)
		return diag.FromErr(err)
	}

	err = flattenFleetPlanJobAndTargetStatus(d, jobStatus)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fleetplanName + "-" + fleetplanJobName)
	return diags
}

func flattenFleetPlanJobAndTargetStatus(d *schema.ResourceData, jobStatus *infrapb.FleetPlanJobStatus) error {
	obj := map[string]interface{}{}

	obj["name"] = jobStatus.JobStatus.Name
	obj["status"] = jobStatus.JobStatus.Status
	obj["reason"] = jobStatus.JobStatus.Reason

	obj["targets"] = flattenFleetPlanJobTargets(d, jobStatus.ResourcesStatus)

	err := d.Set("status", []interface{}{obj})
	if err != nil {
		return err
	}
	return nil
}

func flattenFleetPlanJobTargets(d *schema.ResourceData, resourceStatus []*infrapb.ResourceStatusInfo) []interface{} {
	if resourceStatus == nil {
		return nil
	}
	out := make([]interface{}, len(resourceStatus))
	for i, resource := range resourceStatus {
		target := map[string]interface{}{
			"name":       resource.Name,
			"project":    resource.Project,
			"operations": flattenFleetPlanJobTargetOperations(d, resource.Operations),
		}
		out[i] = target
	}
	return out
}

func flattenFleetPlanJobTargetOperations(d *schema.ResourceData, operations []*infrapb.OperationStatus) []interface{} {
	if operations == nil {
		return nil
	}
	out := make([]interface{}, len(operations))
	for i, operation := range operations {
		op := map[string]interface{}{
			"name": operation.Name,
		}
		if operation.Action != nil {
			op["status"] = operation.Action.Status
			op["reason"] = operation.Action.Reason
		}
		out[i] = op
	}
	return out
}
