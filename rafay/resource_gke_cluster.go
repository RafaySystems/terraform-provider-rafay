package rafay

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	dynamic "github.com/RafaySystems/rafay-common/pkg/hub/client/dynamic"
	"github.com/RafaySystems/rafay-common/pkg/hub/client/options"
	typed "github.com/RafaySystems/rafay-common/pkg/hub/client/typed"
	"github.com/RafaySystems/rafay-common/proto/types/common"
	"github.com/RafaySystems/rafay-common/proto/types/hub/commonpb"
	"github.com/RafaySystems/rafay-common/proto/types/hub/infrapb"
	"github.com/RafaySystems/rctl/pkg/cluster"
	"github.com/RafaySystems/rctl/pkg/config"
	"github.com/RafaySystems/rctl/pkg/versioninfo"
	"github.com/davecgh/go-spew/spew"
	_tflog "github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	schema "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
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
		Schema:        GKEClusterV3Schema(),
	}
}

func resourceGKEClusterV3Import(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	log.Printf("GKE Cluster Import starts")

	idParts := strings.SplitN(d.Id(), "/", 2)
	if len(idParts) != 2 {
		return nil, fmt.Errorf("invalid id %s, expected name/project", d.Id())
	}
	log.Println("resourceGKEClusterV3Import idParts:", idParts)

	metadata := commonpb.Metadata{
		Name:    idParts[0],
		Project: idParts[1],
	}

	err := d.Set("metadata", flattenMetaData(&metadata))
	if err != nil {
		log.Println("import set metadata err ", err)
		return nil, err
	}
	d.SetId(metadata.Name)
	return []*schema.ResourceData{d}, nil

}

func resourceGKEClusterV3Upsert(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("GKE Cluster upsert starts")

	var diags diag.Diagnostics
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	if d.State() != nil && d.State().ID != "" {
		n := GetMetaName(d)
		if n != "" && n != d.State().ID {
			log.Printf("cluster name change not supported")
			d.State().Tainted = true
			return diag.FromErr(fmt.Errorf("%s", "cluster name change not supported"))
		}
	}

	c, err := expandGKEClusterToV3(d)
	if err != nil {
		log.Printf("Cluster expandCluster error : %v", err)
		return diag.FromErr(err)
	}

	if c == nil {
		log.Printf("Cluster is nil")
		return diag.FromErr(fmt.Errorf("cluster is nil"))
	}

	log.Println(">>>>>> CLUSTER: ", c)

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}

	log.Println("GKE Cluster upsert: Invoking V3 Cluster Apply")
	err = client.InfraV3().Cluster().Apply(ctx, c, options.ApplyOptions{})
	if err != nil {
		// XXX Debug
		n1 := spew.Sprintf("%+v", c)
		log.Println("GKE Cluster apply cluster:", n1, err)
		return diag.FromErr(err)
	}

	// wait for cluster creation
	ticker := time.NewTicker(time.Duration(60) * time.Second)
	defer ticker.Stop()
	timeout := time.After(time.Duration(90) * time.Minute)

	cName := c.Metadata.Name
	pName := c.Metadata.Project
	d.SetId(c.Metadata.Name)

LOOP:
	for {
		select {
		case <-timeout:
			log.Printf("Cluster operation timed out for clusterName: %s and projectname: %s", cName, pName)
			return diag.FromErr(fmt.Errorf("cluster operation timed out for clusterName: %s and projectname: %s", cName, pName))
		case <-ticker.C:
			uCluster, err2 := client.InfraV3().Cluster().Status(ctx, options.StatusOptions{
				Name:    cName,
				Project: pName,
			})
			if err2 != nil {
				log.Printf("Unable to fetch cluster: %s with projectname: %s . failing due to err: %v", cName, pName, err2)
				return diag.FromErr(err2)
			} else if uCluster == nil {
				log.Printf("Cluster operation has not started. cluster: %s and projectname: %s", cName, pName)
			} else if uCluster.Status != nil && uCluster.Status.Gke != nil {
				gkeStatus := uCluster.Status.Gke
				uClusterCommonStatus := uCluster.Status.CommonStatus
				switch uClusterCommonStatus.ConditionStatus {
				case commonpb.ConditionStatus_StatusSubmitted:
					log.Printf("Cluster operation not completed for cluster: %s and projectname: %s. Waiting 60 seconds more for the operation to complete.", cName, pName)
				case commonpb.ConditionStatus_StatusOK:
					log.Printf("Cluster operation completed for cluster: %s and projectname: %s", cName, pName)
					break LOOP
				case commonpb.ConditionStatus_StatusFailed:
					failureReasons, err := collectGKEUpsertErrors(gkeStatus)
					if err != nil {
						return diag.Errorf("Cluster operation failed for cluster: %s and projectname: %s. Error collecting reasons: %s", cName, pName, err)
					}
					log.Printf("Cluster operation failed for cluster: %s and projectname: %s with failure reason: %s", cName, pName, uClusterCommonStatus.Reason)
					return diag.Errorf("Cluster operation failed for cluster: %s and projectname: %s with failure reasons: %s", cName, pName, failureReasons)
				}

			}
		}
	}

	pid, err := getProjectIDFromName(c.Metadata.Project)
	if err != nil {
		_tflog.Error(ctx, "failed to get project id from name", map[string]any{"name": c.Metadata.Project})
		return diag.FromErr(err)
	}
	edge, err := cluster.GetCluster(c.Metadata.Name, pid, uaDef)
	if err != nil {
		_tflog.Error(ctx, "failed to edge cluster", map[string]any{"name": c.Metadata.Name, "pid": pid})
		return diag.FromErr(err)
	}
	cse := edge.Settings[clusterSharingExtKey]

	// for update, sharing is removed. Unset cse flag.
	if c.GetSpec().Sharing == nil && cse == "false" {
		edge.Settings[clusterSharingExtKey] = ""
		err = cluster.UpdateCluster(edge, uaDef)
		if err != nil {
			_tflog.Error(ctx, "failed to update cluster", map[string]any{"edgeObj": edge})
			return diag.Errorf("Unable to update the edge object, got error: %s", err)
		}
		_tflog.Debug(ctx, "cse flag unset successfully")
	}

	// for create, set cse flag
	if c.GetSpec().Sharing != nil && cse != "false" {
		edge.Settings[clusterSharingExtKey] = "false"
		err = cluster.UpdateCluster(edge, uaDef)
		if err != nil {
			_tflog.Error(ctx, "failed to update cluster", map[string]any{"edgeObj": edge})
			return diag.Errorf("Unable to update the edge object, got error: %s", err)
		}
		_tflog.Debug(ctx, "cse is set to false")
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

	log.Printf("resourceGKEClusterV3Read GKE")
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	log.Printf("resourceGKEClusterV3Read GKE. Invoking expandGKEClusterToV3")
	tfClusterState, err := expandGKEClusterToV3(d)
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

	// Get cluster from edgedb to identity how sharing is managed.
	pid, err := getProjectIDFromName(ag.Metadata.Project)
	if err != nil {
		_tflog.Error(ctx, "failed to get project id from name", map[string]any{"name": ag.Metadata.Project})
		return diag.FromErr(err)
	}
	edge, err := cluster.GetCluster(ag.Metadata.Name, pid, uaDef)
	if err != nil {
		_tflog.Error(ctx, "failed to edge cluster", map[string]any{"name": ag.Metadata.Name, "pid": pid})
		return diag.FromErr(err)
	}
	cse := edge.Settings[clusterSharingExtKey]
	_tflog.Info(ctx, "Got edge obj from upstream", map[string]any{"cse": cse})
	if cse == "true" {
		ag.Spec.Sharing = nil
	}

	err = flattenGKEClusterV3(d, ag)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags

}

func resourceGKEClusterV3Update(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("GKE Cluster update starts")

	prjName, ok := d.Get("metadata.0.project").(string)
	if !ok || prjName == "" {
		return diag.FromErr(errors.New("unable to find project name"))
	}
	clusterName, ok := d.Get("metadata.0.name").(string)
	if !ok || clusterName == "" {
		return diag.FromErr(errors.New("unable to find cluster name"))
	}
	prjId, err := getProjectIDFromName(prjName)
	if err != nil {
		return diag.Errorf("Failed to get project id from name. Error: %s", err)
	}
	c, err := cluster.GetCluster(clusterName, prjId, uaDef)
	if err != nil {
		return diag.Errorf("Failed to get cluster. Error: %s", err)
	}
	cse := c.Settings[clusterSharingExtKey]
	// Check if cse == true and `spec.sharing` specified. then
	// Error out here only before procedding.
	if cse == "true" {
		if d.HasChange("spec.0.sharing") {
			_, new := d.GetChange("spec.0.sharing")
			if new != nil {
				return diag.Errorf("Cluster sharing is currently managed through the external 'rafay_cluster_sharing' resource. To prevent configuration conflicts, please remove the sharing settings from the 'rafay_gke_cluster' resource and manage sharing exclusively via the external resource.")
			}
		}
	}

	diags := resourceGKEClusterV3Upsert(ctx, d, m)
	return diags
}

func resourceGKEClusterV3Delete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("GKE Cluster delete starts")

	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	log.Printf("GKE Cluster delete: Invoking expandGKEClusterToV3")
	ag, err := expandGKEClusterToV3(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent(), options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("GKE Cluster delete: Invoking V3 Cluster Delete")
	err = client.InfraV3().Cluster().Delete(ctx, options.DeleteOptions{
		Name:    ag.Metadata.Name,
		Project: ag.Metadata.Project,
	})
	if err != nil {
		log.Printf("cluster delete failed for cluster: %s and projectname: %s", ag.Metadata.Name, ag.Metadata.Project)
		return diag.FromErr(err)
	}

	ticker := time.NewTicker(time.Duration(30) * time.Second)
	defer ticker.Stop()
	timeout := time.After(time.Duration(10) * time.Minute)

	edgeName := ag.Metadata.Name
	projectName := ag.Metadata.Project

LOOP:
	for {
		select {
		case <-timeout:
			log.Printf("Cluster Deletion for cluster: %s and projectname: %s got timeout out.", edgeName, projectName)
			return diag.FromErr(fmt.Errorf("cluster deletion for cluster: %s and projectname: %s got timeout out", edgeName, projectName))
		case <-ticker.C:
			_, err := client.InfraV3().Cluster().Get(ctx, options.GetOptions{
				Name:    edgeName,
				Project: projectName,
			})
			if dErr, ok := err.(*dynamic.DynamicClientGetError); ok && dErr != nil {
				switch dErr.StatusCode {
				case http.StatusNotFound:
					log.Printf("Cluster Deletion completes for cluster: %s and projectname: %s", edgeName, projectName)
					break LOOP
				default:
					log.Printf("Cluster Deletion failed for cluster: %s and projectname: %s with error: %s", edgeName, projectName, dErr.Error())
					return diag.FromErr(dErr)
				}
			}
			log.Printf("Cluster Deletion is in progress for cluster: %s and projectname: %s", edgeName, projectName)
		}
	}

	return diags
}

func collectGKEUpsertErrors(gkeStatus *infrapb.GkeStatus) (string, error) {
	if gkeStatus == nil {
		return "", fmt.Errorf("gkeStatus is nil")
	}

	// Defining local struct just to collect errors in json-prettify format to display the same to end user for better visualization.
	type GkeUpsertErrorFormatter struct {
		Name          string `json:"name"`
		Type          string `json:"condition"`
		FailureReason string `json:"failureReason"`
	}

	// adding errors into GkeUpsertErrorFormatter
	collectedErrors := GkeUpsertErrorFormatter{}

	for _, c := range gkeStatus.Conditions {
		if c.Status == common.Failed.String() {
			collectedErrors.Name = "Cluster"
			collectedErrors.Type = c.Type
			collectedErrors.FailureReason = c.Reason
		}
	}

	for _, np := range gkeStatus.Nodepools {
		for _, npc := range np.Conditions {
			if npc.Status == common.Failed.String() {
				collectedErrors.Name = "NodePool-" + np.Name
				collectedErrors.Type = npc.Type
				collectedErrors.FailureReason = npc.Reason
			}
		}
	}

	// Using MarshalIndent to indent the errors in json formatted bytes
	collectedErrsFormattedBytes, err := json.MarshalIndent(collectedErrors, "", "    ")
	if err != nil {
		return "", err
	}
	fmt.Println("After MarshalIndent: ", "collectedErrsFormattedBytes", string(collectedErrsFormattedBytes))
	return "\n" + string(collectedErrsFormattedBytes), nil
}
