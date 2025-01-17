package rafay

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/RafaySystems/rafay-common/proto/types/hub/commonpb"
	"github.com/RafaySystems/rctl/pkg/cluster"
	"github.com/RafaySystems/rctl/pkg/config"
	"github.com/RafaySystems/rctl/pkg/project"
	"github.com/RafaySystems/rctl/pkg/share"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

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
							Computed: true,
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
	if errGet != nil {
		log.Printf("failed to get cluster info %s", errGet.Error())
		return diag.FromErr(errGet)
	}
	if clusterObj == nil {
		log.Printf("failed to get cluster info")
		return diag.FromErr(fmt.Errorf("failed to get cluster info"))
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
			_, err = cluster.AssignClusterToProjects(clusterObj.ID, projectObj.ID, share.ShareModeCustom, []string{addProject.Id})
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
					_, err = cluster.UnassignClusterFromProjects(clusterObj.ID, projectObj.ID, share.ShareModeCustom, []string{oldProjectID})
					if err != nil {
						log.Printf("failed to remove cluster from old project: %v", oldProjectName)
						return diag.FromErr(err)
					}
					projs = removeProjects(oldProjectName, projs)
				}

				// Add the cluster to the new project
				if !isProjectShared {
					_, err = cluster.AssignClusterToProjects(clusterObj.ID, projectObj.ID, share.ShareModeCustom, []string{addProject.Id})
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
	if errGet != nil {
		log.Printf("failed to get cluster info %s", errGet.Error())
		return diag.FromErr(errGet)
	}
	if clusterObj == nil {
		log.Printf("failed to get cluster info")
		return diag.FromErr(fmt.Errorf("failed to get cluster info"))
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
	if errGet != nil {
		log.Printf("failed to get cluster info %s", errGet.Error())
		return diag.FromErr(errGet)
	}
	if clusterObj == nil {
		log.Printf("failed to get cluster info")
		return diag.FromErr(fmt.Errorf("failed to get cluster info"))
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

	_, err = cluster.UnassignClusterFromProjects(clusterObj.ID, projectObj.ID, share.ShareModeCustom, []string{addProject.Id})
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
