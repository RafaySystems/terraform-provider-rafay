package rafay

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/RafaySystems/rafay-common/pkg/hub/client/options"
	typed "github.com/RafaySystems/rafay-common/pkg/hub/client/typed"
	"github.com/RafaySystems/rafay-common/proto/types/hub/commonpb"
	"github.com/RafaySystems/rafay-common/proto/types/hub/infrapb"
	"github.com/RafaySystems/rctl/pkg/cluster"
	"github.com/RafaySystems/rctl/pkg/config"
	"github.com/RafaySystems/rctl/pkg/project"
	"github.com/RafaySystems/rctl/pkg/rerror"
	"github.com/RafaySystems/rctl/pkg/share"
	"github.com/RafaySystems/rctl/pkg/versioninfo"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// clusterSharingClusterNotFound reports whether err indicates the cluster no longer exists
// (e.g. deleted outside Terraform). See rctl/pkg/cluster.GetCluster.
func clusterSharingClusterNotFound(err error) bool {
	var nf rerror.ResourceNotFound
	return err != nil && errors.As(err, &nf)
}

func resourceClusterSharingSingle() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceClusterSharingSingleCreate,
		ReadContext:   resourceClusterSharingSingleRead,
		UpdateContext: resourceClusterSharingSingleUpdate,
		DeleteContext: resourceClusterSharingSingleDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"clustername": {
				Description: "Name of the cluster to be shared",
				Type:        schema.TypeString,
				Required:    true,
			},
			"project": {
				Description: "Name of the project where cluster is created",
				Type:        schema.TypeString,
				Required:    true,
			},
			"sharing": &schema.Schema{
				Description: "cluster sharing configuration",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"projectname": {
							Description: "Name of the project the cluster is shared to",
							Type:        schema.TypeString,
							Required:    true,
						},
						"projects_list": {
							Description: "List of projects cluster shared with",
							Type:        schema.TypeList,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Description: "Name of the project the cluster is shared to",
										Type:        schema.TypeString,
										Computed:    true,
									},
									"id": {
										Description: "Id of the project the cluster is shared to",
										Type:        schema.TypeString,
										Computed:    true,
									},
								},
							},
							Computed:  true,
							Sensitive: true,
						},
					},
				},
				Required: true,
				Type:     schema.TypeList,
				MaxItems: 1,
			},
		},
	}
}

func resourceClusterSharingSingleCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceClusterSharingSingleUpsert(ctx, d, true)
}

func resourceClusterSharingSingleUpsert(ctx context.Context, d *schema.ResourceData, create bool) diag.Diagnostics {
	var diags diag.Diagnostics
	var projs []*commonpb.ProjectMeta
	var addProject commonpb.ProjectMeta
	var err error

	clusterName := d.Get("clustername").(string)
	projectName := d.Get("project").(string)

	if d.State() != nil && d.State().ID != "" {
		if clusterName != "" && clusterName != d.State().ID {
			log.Printf("clusterName change not supported")
			d.State().Tainted = true
			return diag.FromErr(fmt.Errorf("%s", "clusterName change not supported"))
		}
	}

	// get project details
	resp, err := project.GetProjectByName(projectName)
	if err != nil {
		fmt.Printf("project does not exist")
		return diags
	}
	projectObj, err := project.NewProjectFromResponse([]byte(resp))
	if err != nil {
		fmt.Printf("project does not exist")
		return diags
	}

	clusterObj, errGet := cluster.GetCluster(clusterName, projectObj.ID, uaDef)
	if clusterSharingClusterNotFound(errGet) {
		if create {
			return diag.FromErr(errGet)
		}
		tflog.Warn(ctx, "cluster not found during update; removing resource from Terraform state", map[string]any{"cluster": clusterName})
		d.SetId("")
		return diags
	}
	if errGet != nil {
		log.Printf("failed to get cluster info %s", errGet.Error())
		return diag.FromErr(errGet)
	}
	if clusterObj == nil {
		log.Printf("failed to get cluster info")
		return diag.FromErr(fmt.Errorf("failed to get cluster info"))
	}

	cse := clusterObj.Settings[clusterSharingExtKey]
	// TODO(Akshay): convert to Info later
	tflog.Error(ctx, "Got cluster from backend", map[string]any{clusterSharingExtKey: cse})

	if cse == "false" {
		// Cluster is using `spec.sharing` for sharing management.
		return diag.Errorf("Detected conflicting cluster sharing configurations in both 'rafay_*_cluster' and 'rafay_cluster_sharing' resources. Please consolidate the sharing settings into a single resource to ensure consistent cluster sharing behavior.")
	}

	// For infra-apiserver managed clusters route through the typed V3 client.
	if isInfraV3ManagedCluster(clusterObj.ClusterType) {
		return upsertClusterSharingSingleInfraV3(ctx, d, clusterName, projectName, clusterObj.ID, projectObj.ID, create)
	}

	if v, ok := d.Get("sharing").([]interface{}); ok && len(v) > 0 {
		if n, ok1 := v[0].(map[string]interface{})["projectname"].(string); ok1 {
			addProject.Name = n
		} else {
			return diag.Errorf("projectname should not be empty")
		}
		respShared, err := project.GetProjectByName(addProject.Name)
		if err != nil {
			fmt.Printf("project does not exist")
			return diags
		}
		projectObjShared, err := project.NewProjectFromResponse([]byte(respShared))
		if err != nil {
			fmt.Printf("project does not exist")
			return diags
		} else {
			log.Println("projectdetails--------------------", projectObjShared)
			addProject.Id = projectObjShared.ID
		}

	} else {
		return diag.Errorf("sharing spec should not be empty")
	}

	log.Println("clusterObj share type", clusterObj.ShareMode)
	for _, p := range clusterObj.Projects {
		if p.ProjectID == projectObj.ID {
			//skip owner/parent projects
			continue
		}
		pName, err := config.GetProjectNameById(p.ProjectID)
		if err != nil {
			log.Println("get project name from cluster project list failed ", p.ProjectID, err.Error())
		} else {
			var prj commonpb.ProjectMeta
			prj.Id = p.ProjectID
			prj.Name = pName
			log.Println("cluster project list info: ", p.ProjectID, pName)
			projs = append(projs, &prj)
		}
	}
	isProjectShared := false
	if len(projs) > 0 {
		for _, p := range projs {
			if p.Name == addProject.Name {
				isProjectShared = true
			}
		}
	}

	if clusterObj.ShareMode == share.ShareModeAll {
		log.Println("cluster shared mode all, no action required ")
		d.SetId(clusterName)
		return diags
	}

	if addProject.Id == projectObj.ID {
		log.Println("cluster cannot be shared to same project")
		return diags
	}
	if create {
		if !isProjectShared {
			_, err = cluster.AssignClusterToProjects(clusterObj.ID, projectObj.ID, share.ShareModeCustom, []string{addProject.Id}, uaDef, clusterSharingExt)
			if err != nil {
				log.Printf("failed to share cluster to new project")
				return diag.FromErr(err)
			}
			projs = append(projs, &addProject)
		}
	} else {
		if d.HasChange("sharing") {
			old, new := d.GetChange("sharing")
			oldProjectName := old.([]interface{})[0].(map[string]interface{})["projectname"].(string)
			newProjectName := new.([]interface{})[0].(map[string]interface{})["projectname"].(string)

			if oldProjectName != newProjectName {
				// Remove the cluster from the old project
				oldProjectID, err := config.GetProjectIdByName(oldProjectName)
				if err == nil {
					_, err = cluster.UnassignClusterFromProjects(clusterObj.ID, projectObj.ID, share.ShareModeCustom, []string{oldProjectID}, uaDef, clusterSharingExt)
					if err != nil {
						log.Printf("failed to remove cluster from old project: %v", oldProjectName)
						return diag.FromErr(err)
					}
					projs = removeProjects(oldProjectName, projs)
				}

				// Add the cluster to the new project
				if !isProjectShared {
					_, err = cluster.AssignClusterToProjects(clusterObj.ID, projectObj.ID, share.ShareModeCustom, []string{addProject.Id}, uaDef, clusterSharingExt)
					if err != nil {
						log.Printf("failed to share cluster to new project")
						return diag.FromErr(err)
					}
					projs = append(projs, &addProject)
				}
			}
		}
	}

	d.Set("sharing", []interface{}{
		map[string]interface{}{
			"projectname":   addProject.Name,
			"projects_list": getProjectList(projs),
		},
	})
	d.SetId(clusterName)
	return diags
}

func resourceClusterSharingSingleRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var projs []*commonpb.ProjectMeta
	var addProject commonpb.ProjectMeta
	var err error

	clusterName := d.Get("clustername").(string)
	projectName := d.Get("project").(string)

	if d.State() != nil && d.State().ID != "" {
		if clusterName != "" && clusterName != d.State().ID {
			log.Printf("clusterName change not supported")
			d.State().Tainted = true
			return diag.FromErr(fmt.Errorf("%s", "clusterName change not supported"))
		}
	}

	// get project details
	resp, err := project.GetProjectByName(projectName)
	if err != nil {
		fmt.Printf("project does not exist")
		return diags
	}
	projectObj, err := project.NewProjectFromResponse([]byte(resp))
	if err != nil {
		fmt.Printf("project does not exist")
		return diags
	}

	clusterObj, errGet := cluster.GetCluster(clusterName, projectObj.ID, uaDef)
	if clusterSharingClusterNotFound(errGet) {
		tflog.Warn(ctx, "cluster not found during read; removing resource from Terraform state", map[string]any{"cluster": clusterName})
		d.SetId("")
		return diags
	}
	if errGet != nil {
		log.Printf("failed to get cluster info %s", errGet.Error())
		return diag.FromErr(errGet)
	}
	if clusterObj == nil {
		log.Printf("failed to get cluster info")
		return diag.FromErr(fmt.Errorf("failed to get cluster info"))
	}

	// For infra-apiserver managed clusters read via the typed V3 client.
	if isInfraV3ManagedCluster(clusterObj.ClusterType) {
		return readClusterSharingSingleInfraV3(ctx, d, clusterName, projectName)
	}

	if v, ok := d.Get("sharing").([]interface{}); ok && len(v) > 0 {
		if n, ok1 := v[0].(map[string]interface{})["projectname"].(string); ok1 {
			addProject.Name = n
		} else {
			return diag.Errorf("projectname should not be empty")
		}
		respShared, err := project.GetProjectByName(addProject.Name)
		if err != nil {
			fmt.Printf("project does not exist")
			return diags
		}
		projectObjShared, err := project.NewProjectFromResponse([]byte(respShared))
		if err != nil {
			fmt.Printf("project does not exist")
			return diags
		} else {
			addProject.Id = projectObjShared.ID
		}

	} else {
		return diag.Errorf("sharing spec should not be empty")
	}
	log.Println("clusterObj share type", clusterObj.ShareMode)
	for _, p := range clusterObj.Projects {
		if p.ProjectID == projectObj.ID {
			//skip owner/parent projects
			continue
		}
		pName, err := config.GetProjectNameById(p.ProjectID)
		if err != nil {
			log.Println("get project name from cluster project list failed ", p.ProjectID, err.Error())
		} else {
			var prj commonpb.ProjectMeta
			prj.Id = p.ProjectID
			prj.Name = pName
			log.Println("cluster project list info: ", p.ProjectID, pName)
			projs = append(projs, &prj)
		}
	}
	isProjectShared := false
	if len(projs) > 0 {
		for _, p := range projs {
			if p.Name == addProject.Name {
				isProjectShared = true
			}
		}
	}
	if !isProjectShared {
		addProject.Name = ""
		addProject.Id = ""
	}
	d.Set("sharing", []interface{}{
		map[string]interface{}{
			"projectname":   addProject.Name,
			"projects_list": getProjectList(projs),
		},
	})
	err = d.Set("clustername", clusterName)
	if err != nil {
		return diag.FromErr(err)
	}
	err = d.Set("project", projectName)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceClusterSharingSingleUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("resource user update id %s", d.Id())
	return resourceClusterSharingSingleUpsert(ctx, d, false)
	//return diag.FromErr(fmt.Errorf("%s", "update not supported for user. Use group association to alter groups"))
}

func resourceClusterSharingSingleDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var addProject commonpb.ProjectMeta
	clusterName := d.Get("clustername").(string)
	projectName := d.Get("project").(string)

	// get project details
	resp, err := project.GetProjectByName(projectName)
	if err != nil {
		fmt.Printf("project does not exist")
		return diags
	}
	projectObj, err := project.NewProjectFromResponse([]byte(resp))
	if err != nil {
		fmt.Printf("project does not exist")
		return diags
	}

	clusterObj, errGet := cluster.GetCluster(clusterName, projectObj.ID, uaDef)
	if clusterSharingClusterNotFound(errGet) {
		tflog.Warn(ctx, "cluster not found during delete; nothing to unshare", map[string]any{"cluster": clusterName})
		d.SetId("")
		return diags
	}
	if errGet != nil {
		log.Printf("failed to get cluster info %s", errGet.Error())
		return diag.FromErr(errGet)
	}
	if clusterObj == nil {
		log.Printf("failed to get cluster info")
		return diag.FromErr(fmt.Errorf("failed to get cluster info"))
	}

	// For infra-apiserver managed clusters delete via the typed V3 client.
	if isInfraV3ManagedCluster(clusterObj.ClusterType) {
		return deleteClusterSharingSingleInfraV3(ctx, d, clusterName, projectName, clusterObj.ID, projectObj.ID)
	}

	if v, ok := d.Get("sharing").([]interface{}); ok && len(v) > 0 {
		if n, ok1 := v[0].(map[string]interface{})["projectname"].(string); ok1 {
			addProject.Name = n
		} else {
			return diag.Errorf("projectname should not be empty")
		}
		respShared, err := project.GetProjectByName(addProject.Name)
		if err != nil {
			fmt.Printf("project does not exist")
			return diags
		}
		projectObjShared, err := project.NewProjectFromResponse([]byte(respShared))
		if err != nil {
			fmt.Printf("project does not exist")
			return diags
		} else {
			addProject.Id = projectObjShared.ID
		}

	} else {
		return diag.Errorf("sharing spec should not be empty")
	}

	_, err = cluster.UnassignClusterFromProjects(clusterObj.ID, projectObj.ID, share.ShareModeCustom, []string{addProject.Id}, uaDef, "")
	if err != nil {
		log.Printf("cluster share setting had all, but failed to unshare form all projects")
		return diag.FromErr(err)
	}

	return diags

}

func getProjectList(projs []*commonpb.ProjectMeta) []map[string]interface{} {
	var projectsList []map[string]interface{}
	if len(projs) == 0 {
		return projectsList
	}
	for _, p := range projs {
		projectsList = append(projectsList, map[string]interface{}{
			"name": p.Name,
			"id":   p.Id,
		})
	}
	return projectsList
}

func removeProjects(projectNameToRemove string, projs []*commonpb.ProjectMeta) []*commonpb.ProjectMeta {
	var updatedProjs []*commonpb.ProjectMeta

	for _, p := range projs {
		// Exclude the project to be removed
		if p.Name != projectNameToRemove {
			updatedProjs = append(updatedProjs, p)
		} else {
			log.Printf("Removing project: %s (ID: %s)", p.Name, p.Id)
		}
	}

	return updatedProjs
}

// upsertClusterSharingSingleInfraV3 handles create/update for cluster_sharing_single on
// infra-apiserver-managed clusters. It modifies only the target project in spec.Sharing,
// preserving any other projects already shared.
func upsertClusterSharingSingleInfraV3(ctx context.Context, d *schema.ResourceData, clusterName, projectName, clusterID, projectID string, create bool) diag.Diagnostics {
	var diags diag.Diagnostics

	newProjectName := ""
	if v, ok := d.Get("sharing").([]interface{}); ok && len(v) > 0 {
		if n, ok1 := v[0].(map[string]interface{})["projectname"].(string); ok1 {
			newProjectName = n
		}
	}
	if newProjectName == "" {
		return diag.Errorf("cluster_sharing_single: projectname must not be empty")
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}
	ag, err := client.InfraV3().Cluster().Get(ctx, options.GetOptions{
		Name:    clusterName,
		Project: projectName,
	})
	if err != nil {
		return diag.FromErr(fmt.Errorf("cluster_sharing_single infraV3 get failed: %w", err))
	}

	// Build updated project list: start from current sharing, modify target entry.
	existing := []*infrapb.Projects{}
	if ag.Spec.Sharing != nil && ag.Spec.Sharing.Enabled {
		for _, p := range ag.Spec.Sharing.Projects {
			if p.Name != "*" {
				existing = append(existing, p)
			}
		}
	}

	if !create && d.HasChange("sharing") {
		old, _ := d.GetChange("sharing")
		oldName := ""
		if ov, ok := old.([]interface{}); ok && len(ov) > 0 {
			if n, ok1 := ov[0].(map[string]interface{})["projectname"].(string); ok1 {
				oldName = n
			}
		}
		// Remove old project from the list.
		updated := []*infrapb.Projects{}
		for _, p := range existing {
			if p.Name != oldName {
				updated = append(updated, p)
			}
		}
		existing = updated
	}

	// Add new project if not already present.
	found := false
	for _, p := range existing {
		if p.Name == newProjectName {
			found = true
			break
		}
	}
	if !found {
		existing = append(existing, &infrapb.Projects{Name: newProjectName})
	}

	ag.Spec.Sharing = &infrapb.Sharing{Enabled: true, Projects: existing}
	if err = client.InfraV3().Cluster().Apply(ctx, ag, options.ApplyOptions{}); err != nil {
		return diag.FromErr(fmt.Errorf("cluster_sharing_single infraV3 apply failed: %w", err))
	}

	// Set v1 cluster_sharing_external flag so rafay_gke_cluster read suppresses sharing.
	var projectIDs []string
	for _, p := range existing {
		pID, e := config.GetProjectIdByName(p.Name)
		if e == nil {
			projectIDs = append(projectIDs, pID)
		}
	}
	cluster.AssignClusterToProjects(clusterID, projectID, share.ShareModeCustom, projectIDs, uaDef, clusterSharingExt)

	d.Set("sharing", []interface{}{map[string]interface{}{
		"projectname":   newProjectName,
		"projects_list": func() []map[string]interface{} {
			var out []map[string]interface{}
			for _, p := range existing {
				out = append(out, map[string]interface{}{"name": p.Name, "id": ""})
			}
			return out
		}(),
	}})
	d.SetId(clusterName)
	return diags
}

// readClusterSharingSingleInfraV3 reads the current sharing state and reflects whether
// the target projectname is still shared.
func readClusterSharingSingleInfraV3(ctx context.Context, d *schema.ResourceData, clusterName, projectName string) diag.Diagnostics {
	var diags diag.Diagnostics

	targetName := ""
	if v, ok := d.Get("sharing").([]interface{}); ok && len(v) > 0 {
		if n, ok1 := v[0].(map[string]interface{})["projectname"].(string); ok1 {
			targetName = n
		}
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}
	ag, err := client.InfraV3().Cluster().Get(ctx, options.GetOptions{
		Name:    clusterName,
		Project: projectName,
	})
	if err != nil {
		return diag.FromErr(fmt.Errorf("cluster_sharing_single infraV3 read failed: %w", err))
	}

	allProjs := []map[string]interface{}{}
	isShared := false
	if ag.Spec.Sharing != nil && ag.Spec.Sharing.Enabled {
		for _, p := range ag.Spec.Sharing.Projects {
			if p.Name == "*" {
				continue
			}
			allProjs = append(allProjs, map[string]interface{}{"name": p.Name, "id": ""})
			if p.Name == targetName {
				isShared = true
			}
		}
	}
	if !isShared {
		targetName = ""
	}

	d.Set("sharing", []interface{}{map[string]interface{}{
		"projectname":   targetName,
		"projects_list": allProjs,
	}})
	if err2 := d.Set("clustername", clusterName); err2 != nil {
		return diag.FromErr(err2)
	}
	if err2 := d.Set("project", projectName); err2 != nil {
		return diag.FromErr(err2)
	}
	return diags
}

// deleteClusterSharingSingleInfraV3 removes only the target project from sharing.
// If no projects remain after removal, sharing is set to nil (disabled).
func deleteClusterSharingSingleInfraV3(ctx context.Context, d *schema.ResourceData, clusterName, projectName, clusterID, projectID string) diag.Diagnostics {
	var diags diag.Diagnostics

	targetName := ""
	if v, ok := d.Get("sharing").([]interface{}); ok && len(v) > 0 {
		if n, ok1 := v[0].(map[string]interface{})["projectname"].(string); ok1 {
			targetName = n
		}
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}
	ag, err := client.InfraV3().Cluster().Get(ctx, options.GetOptions{
		Name:    clusterName,
		Project: projectName,
	})
	if err != nil {
		return diag.FromErr(fmt.Errorf("cluster_sharing_single infraV3 get failed on delete: %w", err))
	}

	remaining := []*infrapb.Projects{}
	if ag.Spec.Sharing != nil && ag.Spec.Sharing.Enabled {
		for _, p := range ag.Spec.Sharing.Projects {
			if p.Name != targetName && p.Name != "*" {
				remaining = append(remaining, p)
			}
		}
	}

	if len(remaining) == 0 {
		ag.Spec.Sharing = nil
		// Clear the v1 cluster_sharing_external flag.
		cluster.UnassignClusterFromProjects(clusterID, projectID, share.ShareModeAll, []string{}, uaDef, "")
	} else {
		ag.Spec.Sharing = &infrapb.Sharing{Enabled: true, Projects: remaining}
		// Keep flag set; update project list in v1.
		var projectIDs []string
		for _, p := range remaining {
			pID, e := config.GetProjectIdByName(p.Name)
			if e == nil {
				projectIDs = append(projectIDs, pID)
			}
		}
		cluster.AssignClusterToProjects(clusterID, projectID, share.ShareModeCustom, projectIDs, uaDef, clusterSharingExt)
	}

	if err = client.InfraV3().Cluster().Apply(ctx, ag, options.ApplyOptions{}); err != nil {
		return diag.FromErr(fmt.Errorf("cluster_sharing_single infraV3 apply failed on delete: %w", err))
	}
	d.SetId("")
	return diags
}
