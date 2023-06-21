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
	"github.com/RafaySystems/rafay-common/proto/types/hub/commonpb"
	"github.com/RafaySystems/rctl/pkg/config"
	"github.com/RafaySystems/rctl/pkg/versioninfo"
	"github.com/davecgh/go-spew/spew"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	schema "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceGKEClusterV3() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceGKEClusterV3Create,
		ReadContext:   resourceGKEClusterV3Read,
		UpdateContext: resourceGKEClusterV3Update,
		DeleteContext: resourceGKEClusterV3Delete,
		Importer: &schema.ResourceImporter{
			State: resourceGKEClusterV3Import,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(90 * time.Minute),
			Update: schema.DefaultTimeout(90 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema:        resource.ClusterSchema.Schema,
	}
}

func resourceGKEClusterV3Import(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	// TODO

	return []*schema.ResourceData{d}, nil
}

func resourceGKEClusterV3Upsert(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("GKE Cluster upsert starts")
	// tflog := os.Getenv("TF_LOG")
	//if tflog == "TRACE" || tflog == "DEBUG" {
	ctx = context.WithValue(ctx, "debug", "true")
	//}

	tflog.Info(ctx, "In resourceGKEClusterV3Upsert")

	if d.State() != nil && d.State().ID != "" {
		n := GetMetaName(d)
		if n != "" && n != d.State().ID {
			log.Printf("cluster name change not supported")
			d.State().Tainted = true
			return diag.FromErr(fmt.Errorf("%s", "cluster name change not supported"))
		}
	}

	//cluster, err := expandClusterV3(d)
	cluster, err := expandGKEClusterToV3(d)
	if err != nil {
		log.Printf("Cluster expandCluster error " + err.Error())
		return diag.FromErr(err)
	}

	log.Println(">>>>>> CLUSTER: ", cluster)

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.InfraV3().Cluster().Apply(ctx, cluster, options.ApplyOptions{})
	if err != nil {
		// XXX Debug
		n1 := spew.Sprintf("%+v", cluster)
		log.Println("Cluster apply cluster:", n1)
		log.Printf("Cluster apply error")
		return diag.FromErr(err)
	}

	// wait for cluster creation
	ticker := time.NewTicker(time.Duration(60) * time.Second)
	defer ticker.Stop()
	timeout := time.After(time.Duration(90) * time.Minute)

	edgeName := cluster.Metadata.Name
	projectName := cluster.Metadata.Project
	d.SetId(cluster.Metadata.Name)

LOOP:
	for {
		select {
		case <-timeout:
			log.Printf("Cluster operation timed out for edgeName: %s and projectname: %s", edgeName, projectName)
			return diag.FromErr(fmt.Errorf("cluster operation timed out for edgeName: %s and projectname: %s", edgeName, projectName))
		case <-ticker.C:
			uCluster, err2 := client.InfraV3().Cluster().Status(ctx, options.StatusOptions{
				Name:    edgeName,
				Project: projectName,
			})
			if err2 != nil {
				log.Printf("Fetching cluster having edgename: %s and projectname: %s failing due to err: %v", edgeName, projectName, err2)
				return diag.FromErr(err2)
			} else if uCluster == nil {
				log.Printf("Cluster operation has not started with edgename: %s and projectname: %s", edgeName, projectName)
			} else if uCluster.Status != nil && uCluster.Status.Gke != nil {
				// TODO: revisit this for gke status part
				//	gkeStatus := uCluster.Status.Gke
				//	gkeConditions := uCluster.Status.Gke.Conditions
				uClusterCommonStatus := uCluster.Status.CommonStatus
				switch uClusterCommonStatus.ConditionStatus {
				case commonpb.ConditionStatus_StatusSubmitted:
					log.Printf("Cluster operation not completed for edgename: %s and projectname: %s. Waiting 60 seconds more for cluster to complete the operation.", edgeName, projectName)
				case commonpb.ConditionStatus_StatusOK:
					log.Printf("Cluster operation completed for edgename: %s and projectname: %s", edgeName, projectName)
					break LOOP
				case commonpb.ConditionStatus_StatusFailed:
					log.Printf("Cluster operation failed for edgename: %s and projectname: %s with failure reason: %s", edgeName, projectName, uClusterCommonStatus.Reason)
					// failureReasons, err := collectGKEUpsertErrors(gkeStatus.Nodepools, uCluster.Status.ProvisionStatusReason, uCluster.Status.ProvisionStatus)
					// if err != nil {
					// 	return diag.FromErr(err)
					// }
					//return diag.Errorf("Cluster operation failed for edgename: %s and projectname: %s with failure reasons: %s", edgeName, projectName, failureReasons)
					return diag.Errorf("Cluster operation failed for edgename: %s and projectname: %s", edgeName, projectName)
				}

			}

			// else if uCluster.Status != nil && uCluster.Status.Aks != nil && uCluster.Status.CommonStatus != nil {
			// 	aksStatus := uCluster.Status.Aks
			// 	uClusterCommonStatus := uCluster.Status.CommonStatus
			// 	switch uClusterCommonStatus.ConditionStatus {
			// 	case commonpb.ConditionStatus_StatusSubmitted:
			// 		log.Printf("Cluster operation not completed for edgename: %s and projectname: %s. Waiting 60 seconds more for cluster to complete the operation.", edgeName, projectName)
			// 	case commonpb.ConditionStatus_StatusOK:
			// 		log.Printf("Cluster operation completed for edgename: %s and projectname: %s", edgeName, projectName)
			// 		break LOOP
			// 	case commonpb.ConditionStatus_StatusFailed:
			// 		// log.Printf("Cluster operation failed for edgename: %s and projectname: %s with failure reason: %s", edgeName, projectName, uClusterCommonStatus.Reason)
			// 		failureReasons, err := collectAKSUpsertErrors(aksStatus.Nodepools, uCluster.Status.ProvisionStatusReason, uCluster.Status.ProvisionStatus)
			// 		if err != nil {
			// 			return diag.FromErr(err)
			// 		}
			// 		return diag.Errorf("Cluster operation failed for edgename: %s and projectname: %s with failure reasons: %s", edgeName, projectName, failureReasons)
			// 	}
			// }
		}
	}

	return diags
}

func resourceGKEClusterV3Create(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("GKE Cluster create starts")

	diags := resourceGKEClusterV3Upsert(ctx, d, m)
	return diags

}

func resourceGKEClusterV3Read(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	log.Println("resourceClusterRead GKE")
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}
	tfClusterState, err := expandClusterV3(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}

	ag, err := client.InfraV3().Cluster().Get(ctx, options.GetOptions{
		Name:    tfClusterState.Metadata.Name,
		Project: tfClusterState.Metadata.Project,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	err = flattenGKEClusterV3(d, ag)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags

}

func resourceGKEClusterV3Update(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("GKE Cluster update starts")

	var diags diag.Diagnostics

	diags = resourceGKEClusterV3Upsert(ctx, d, m)
	return diags
}

func resourceGKEClusterV3Delete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("GKE Cluster delete starts")

	return diags
}
