package rafay

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/RafaySystems/rafay-common/proto/types/hub/commonpb"
	"github.com/RafaySystems/rctl/pkg/cluster"
	"github.com/RafaySystems/rctl/pkg/config"
	"github.com/RafaySystems/rctl/pkg/models"
	"github.com/RafaySystems/rctl/pkg/project"
	"github.com/RafaySystems/rctl/pkg/share"
	"github.com/davecgh/go-spew/spew"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceClusterSharing() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceClusterSharingCreate,
		ReadContext:   resourceClusterSharingRead,
		UpdateContext: resourceClusterSharingUpdate,
		DeleteContext: resourceClusterSharingDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"clustername": {
				Type:     schema.TypeString,
				Required: true,
			},
			"project": {
				Type:     schema.TypeString,
				Required: true,
			},
			"sharing": &schema.Schema{
				Description: "cluster sharing configuration",
				Elem: &schema.Resource{Schema: map[string]*schema.Schema{
					"all": &schema.Schema{
						Description: "flag to specify if shared to all projects",
						Optional:    true,
						Type:        schema.TypeBool,
					},
					"projects": &schema.Schema{
						Description: "list of projects this cluster is shared to",
						Elem: &schema.Resource{Schema: map[string]*schema.Schema{
							"name": &schema.Schema{
								Description: "name of the project",
								Optional:    true,
								Type:        schema.TypeString,
							},
						}},
						// 						MaxItems: 0,
						// 						MinItems: 0,
						Optional: true,
						Type:     schema.TypeSet,
					},
				}},
				MaxItems: 1,
				MinItems: 1,
				Optional: true,
				Type:     schema.TypeList,
			},
		},
	}
}

func resourceClusterSharingCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceClusterSharingUpsert(ctx, d, true)
}

func resourceClusterSharingUpsert(ctx context.Context, d *schema.ResourceData, create bool) diag.Diagnostics {
	var diags diag.Diagnostics
	var sharingSpec *commonpb.SharingSpec
	var projs []*commonpb.ProjectMeta
	var sortprojs []*commonpb.ProjectMeta
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

	cse := clusterObj.Settings[clusterSharingExtKey]
	tflog.Info(ctx, "Got cluster from backend", map[string]any{clusterSharingExtKey: cse})
	if cse == "false" {
		// Cluster is using `spec.sharing` for sharing management.
		return diag.Errorf("cluster sharing is managed from rafay_eks_cluster itself.")
	}

	if v, ok := d.Get("sharing").([]interface{}); ok && len(v) > 0 {
		sharingSpec = expandClusterSharingSpec(v)
	}

	if sharingSpec == nil {
		// no sharing info. so set to none
		_, err = cluster.UnassignClusterFromProjects(clusterObj.ID, projectObj.ID, share.ShareModeAll, []string{}, "", false)
		if err != nil {
			log.Printf("cluster failed to unshare form all projects %s ", clusterName)
			return diag.FromErr(err)
		}
		d.SetId(clusterName)
		return diags
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
			fmt.Printf("project does not exist")
			return diag.FromErr(err)
		} else {
			var prj commonpb.ProjectMeta
			prj.Id = p.ProjectID
			prj.Name = pName
			log.Println("cluster project list info: ", p.ProjectID, pName)
			projs = append(projs, &prj)
		}
	}

	// try to order cluster list based on local state
	if len(projs) > 0 && sharingSpec.Projects != nil {
		for _, p := range sharingSpec.Projects {
			for _, pi := range projs {
				if pi.Id == p.Id {
					sortprojs = append(sortprojs, p)
				}
			}
		}
		for _, p := range projs {
			found1 := false
			for _, pi := range sharingSpec.Projects {
				if pi.Id == p.Id {
					found1 = true
					break
				}
			}
			if !found1 {
				sortprojs = append(sortprojs, p)
			}
		}
	}

	if sharingSpec.Enabled {
		if len(sharingSpec.Projects) > 0 {
			log.Printf("when 'all' is true, do not specify projects list")
			d.State().Tainted = true
			return diag.FromErr(fmt.Errorf("when 'all' is true, do not specify projects list"))
		}
		log.Printf("enabling sharing to all projects")
		if clusterObj.ShareMode == share.ShareModeAll {
			log.Printf("cluster is already shared to ALL projects")
		} else {
			// share to all projects
			log.Println("call AssignClusterToProjects", clusterObj.ID, projectObj.ID, clusterName)
			_, err := cluster.AssignClusterToProjects(clusterObj.ID, projectObj.ID, share.ShareModeAll, []string{}, uaDef, clusterSharingExt)
			if err != nil {
				log.Printf("failed to share cluster to ALL projects")
				return diag.FromErr(err)
			}
		}
		d.SetId(clusterName)
		return diags
	}

	if clusterObj.ShareMode == share.ShareModeAll {
		log.Println("cluster share mode is 'all' so first unassign from 'all'")
		// cluster share mode is 'all' so first unassign from 'all'
		_, err := cluster.UnassignClusterFromProjects(clusterObj.ID, projectObj.ID, share.ShareModeAll, []string{}, uaDef, clusterSharingExt)
		if err != nil {
			log.Printf("cluster share setting had all, but failed to unshare form all projects")
			return diag.FromErr(err)
		}
	}

	// compare incoming spec with cluster sharing data
	// find new shared project
	var newIds []string
	for _, inProj := range sharingSpec.Projects {
		if inProj.Id == "" || len(inProj.Id) <= 0 {
			log.Println("failed to get project id by name ", inProj.Name)
			return diag.FromErr(fmt.Errorf("failed to get project id by name %s", inProj.Name))
		}

		if inProj.Id == projectObj.ID {
			// same as parent project
			continue
		}

		found := false
		for _, cProj := range projs {
			if cProj.Id == inProj.Id {
				found = true
				break
			}
		}

		if !found {
			// does not exist in cluster project list
			newIds = append(newIds, inProj.Id)
			log.Println("append newIds", inProj.Id)
		}
	}

	if len(newIds) > 0 {
		log.Println("cluster share to project ids ", newIds)
		_, err = cluster.AssignClusterToProjects(clusterObj.ID, projectObj.ID, share.ShareModeCustom, newIds, uaDef, clusterSharingExt)
		if err != nil {
			log.Printf("failed to share cluster to new projects")
			return diag.FromErr(err)
		}
	}

	// find projects to unshare
	var oldIds []string
	for _, cProj := range projs {
		if cProj.Id == projectObj.ID {
			// same as parent project
			continue
		}

		found := false
		for _, inProj := range sharingSpec.Projects {
			if cProj.Id == inProj.Id {
				found = true
				break
			}
		}

		if !found {
			// does not exist in incoming project list
			oldIds = append(oldIds, cProj.Id)
		}
	}

	if len(oldIds) > 0 {
		log.Println("cluster unshare from project ids ", oldIds)
		_, err = cluster.UnassignClusterFromProjects(clusterObj.ID, projectObj.ID, share.ShareModeCustom, oldIds, uaDef, clusterSharingExt)
		if err != nil {
			log.Printf("failed to un share cluster from old projects")
			return diag.FromErr(err)
		}
	}

	d.SetId(clusterName)
	return diags
}

func resourceClusterSharingRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var projs []*commonpb.ProjectMeta
	var sortprojs []*commonpb.ProjectMeta
	var sharingSpec *commonpb.SharingSpec

	clusterName := d.Get("clustername").(string)
	projectName := d.Get("project").(string)

	if d.State() != nil && d.State().ID != "" {
		if clusterName != d.State().ID {
			log.Println("detected clusterName change ", clusterName, d.State().ID)
			return diag.FromErr(fmt.Errorf("cannot change name of the cluster"))
		}
	}

	if v, ok := d.Get("sharing").([]interface{}); ok && len(v) > 0 {
		sharingSpec = expandClusterSharingSpec(v)
	}

	// get project details
	resp, err := project.GetProjectByName(d.Get("project").(string))
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

	log.Println("clusterObj share type", clusterObj.ShareMode)
	for _, p := range clusterObj.Projects {
		if p.ProjectID == projectObj.ID {
			//skip owner/parent projects
			continue
		}
		pName, err := config.GetProjectNameById(p.ProjectID)
		if err == nil {
			var prj commonpb.ProjectMeta
			prj.Id = p.ProjectID
			prj.Name = pName
			log.Println("cluster project list info: ", p.ProjectID, pName)
			projs = append(projs, &prj)
		} else {
			log.Println("get project name from cluster project list failed: ", p.ProjectID, err.Error())
		}
	}

	// try to order cluster list based on local state
	if len(projs) > 0 {
		for _, p := range sharingSpec.Projects {
			for _, pi := range projs {
				if pi.Id == p.Id {
					sortprojs = append(sortprojs, p)
				}
			}
		}
		for _, p := range projs {
			found1 := false
			for _, pi := range sharingSpec.Projects {
				if pi.Id == p.Id {
					found1 = true
					break
				}
			}
			if !found1 {
				sortprojs = append(sortprojs, p)
			}
		}
	}

	err = d.Set("clustername", clusterName)
	if err != nil {
		return diag.FromErr(err)
	}
	err = d.Set("project", projectName)
	if err != nil {
		return diag.FromErr(err)
	}

	err = flattenClusterSharing(d, clusterName, projectName, projectObj.ID, clusterObj, sortprojs)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func flattenClusterSharingSpec(in *commonpb.SharingSpec, p []interface{}) ([]interface{}, error) {
	if in == nil {
		return nil, fmt.Errorf("flattenClusterSharingSpec empty input")
	}

	// Always start with a new map to reflect the remote state exactly.
	obj := map[string]interface{}{}
	obj["all"] = in.Enabled

	// If there are any projects (other than the parent), add them; if not, explicitly set an empty list.
	if len(in.Projects) > 0 {
		obj["projects"] = flattenProjectMeta(in.Projects, true)
	} else {
		obj["projects"] = []interface{}{}
	}

	return []interface{}{obj}, nil
}

func flattenClusterSharing(d *schema.ResourceData, clusterName string, projectName string, projectID string, in *models.ClusterDetails, projs []*commonpb.ProjectMeta) error {
	var inSharing commonpb.SharingSpec

	if in.ShareMode == share.ShareModeAll {
		inSharing.Enabled = true
	} else {
		inSharing.Enabled = false
	}

	// Instead of only setting projects when > 1, do so when there’s at least one.
	if len(in.Projects) > 0 {
		inSharing.Projects = projs
	}

	v, ok := d.Get("sharing").([]interface{})
	if !ok {
		v = []interface{}{}
	}

	var ret []interface{}
	ret, err := flattenClusterSharingSpec(&inSharing, v)
	if err != nil {
		return err
	}

	log.Println("flattenClusterSharing after ", spew.Sprintf("%+v", ret))

	if err = d.Set("sharing", ret); err != nil {
		log.Println("failed to set sharing")
		return err
	}

	return nil
}

func resourceClusterSharingUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("resource user update id %s", d.Id())
	return resourceClusterSharingUpsert(ctx, d, false)
	//return diag.FromErr(fmt.Errorf("%s", "update not supported for user. Use group association to alter groups"))
}

func resourceClusterSharingDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
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

	_, err = cluster.UnassignClusterFromProjects(clusterObj.ID, projectObj.ID, share.ShareModeAll, []string{}, "", false)
	if err != nil {
		log.Printf("cluster share setting had all, but failed to unshare form all projects")
		return diag.FromErr(err)
	}

	return diags
}

func expandClusterSharingSpec(p []interface{}) *commonpb.SharingSpec {
	obj := commonpb.SharingSpec{}
	if len(p) == 0 || p[0] == nil {
		return &obj
	}

	in := p[0].(map[string]interface{})
	if v, ok := in["all"].(bool); ok {
		obj.Enabled = v
	}

	if v, ok := in["projects"].([]interface{}); ok && len(v) > 0 {
		obj.Projects = expandProjectMeta(v)
	} else if v, ok := in["projects"].(*schema.Set); ok && v != nil && v.Len() > 0 {
		obj.Projects = expandProjectMeta(v.List())
	}
	for _, inProj := range obj.Projects {
		pID, err := config.GetProjectIdByName(inProj.Name)
		if err != nil {
			log.Println("failed to get project id by name ", inProj.Name)
		} else {
			inProj.Id = pID
		}
	}

	log.Println("expandClusterSharingSpec obj", obj)
	return &obj
}
