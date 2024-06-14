package rafay

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/RafaySystems/rafay-common/proto/types/hub/commonpb"
	"github.com/RafaySystems/rctl/pkg/clusteroverride"
	"github.com/RafaySystems/rctl/pkg/config"
	"github.com/RafaySystems/rctl/pkg/models"
	"github.com/RafaySystems/rctl/pkg/project"
	"github.com/RafaySystems/rctl/pkg/share"
	"github.com/davecgh/go-spew/spew"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceClusterOverrideSharing() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceClusterOverrideSharingCreate,
		ReadContext:   resourceClusterOverrideSharingRead,
		UpdateContext: resourceClusterOverrideSharingUpdate,
		DeleteContext: resourceClusterOverrideSharingDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"clusteroverridename": {
				Type:     schema.TypeString,
				Required: true,
			},
			"project": {
				Type:     schema.TypeString,
				Required: true,
			},
			"clusteroverridetype": {
				Type:     schema.TypeString,
				Required: true,
			},
			"sharing": &schema.Schema{
				Description: "cluster override sharing configuration",
				Elem: &schema.Resource{Schema: map[string]*schema.Schema{
					"all": &schema.Schema{
						Description: "flag to specify if shared to all projects",
						Optional:    true,
						Type:        schema.TypeBool,
					},
					"projects": &schema.Schema{
						Description: "list of projects this cluster override is shared to",
						Elem: &schema.Resource{Schema: map[string]*schema.Schema{
							"name": &schema.Schema{
								Description: "name of the project",
								Optional:    true,
								Type:        schema.TypeString,
							},
							"id": &schema.Schema{
								Description: "id of the project",
								Optional:    true,
								Type:        schema.TypeString,
								Sensitive:   true,
								Computed:    true,
							},
						}},
						MaxItems: 0,
						MinItems: 0,
						Optional: true,
						Type:     schema.TypeList,
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

func resourceClusterOverrideSharingCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceClusterOverrideSharingUpsert(ctx, d, true)
}

func resourceClusterOverrideSharingUpsert(ctx context.Context, d *schema.ResourceData, create bool) diag.Diagnostics {
	var diags diag.Diagnostics
	var sharingSpec *commonpb.SharingSpec
	var projs []*commonpb.ProjectMeta
	var sortprojs []*commonpb.ProjectMeta
	var err error

	clusterOverrideName := d.Get("clusteroverridename").(string)
	projectName := d.Get("project").(string)
	overrideType := d.Get("clusteroverridetype").(string)

	if d.State() != nil && d.State().ID != "" {
		if clusterOverrideName != "" && clusterOverrideName != d.State().ID {
			log.Printf("clusterOverrideName change not supported")
			d.State().Tainted = true
			return diag.FromErr(fmt.Errorf("%s", "clusterOverrideName change not supported"))
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

	clusterOverrideObj, errGet := clusteroverride.GetClusterOverride(clusterOverrideName, projectObj.ID, overrideType)
	if errGet != nil {
		log.Printf("failed to get cluster override info %s", errGet.Error())
		return diag.FromErr(errGet)
	}
	if clusterOverrideObj == nil {
		log.Printf("failed to get cluster override info")
		return diag.FromErr(fmt.Errorf("failed to get cluster info"))
	}

	if v, ok := d.Get("sharing").([]interface{}); ok && len(v) > 0 {
		sharingSpec = expandClusterOverrideSharingSpec(v)
	}

	if sharingSpec == nil {
		// no sharing info. so set to none
		_, err = clusteroverride.UnassignClusterOverrideFromProjects(clusterOverrideObj.Name, overrideType, projectObj.ID, share.ShareModeAll, []string{})
		if err != nil {
			log.Printf("cluster override failed to unshare from all projects %s ", clusterOverrideObj.Name)
			return diag.FromErr(err)
		}
		d.SetId(clusterOverrideObj.Name)
		return diags
	}

	log.Println("clusterOverrideObj share type", clusterOverrideObj.ShareMode)
	for _, p := range clusterOverrideObj.Projects {
		if string(p.ProjectID) == projectObj.ID {
			//skip owner/parent projects
			continue
		}
		pName, err := config.GetProjectNameById(string(p.ProjectID))
		if err != nil {
			log.Println("get project name from cluster override project list failed ", p.ProjectID, err.Error())
		} else {
			var prj commonpb.ProjectMeta
			prj.Id = string(p.ProjectID)
			prj.Name = pName
			log.Println("cluster override project list info: ", p.ProjectID, pName)
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
		if clusterOverrideObj.ShareMode == share.ShareModeAll {
			log.Printf("cluster override is already shared to ALL projects")
		} else {
			// share to all projects
			log.Println("call AssignClusterOverrideToProjects", clusterOverrideObj.ID, projectObj.ID, clusterOverrideName)
			_, err := clusteroverride.AssignClusterOverrideToProjects(clusterOverrideObj.Name, overrideType, projectObj.ID, share.ShareModeAll, []string{})
			if err != nil {
				log.Printf("failed to share cluster override to ALL projects")
				return diag.FromErr(err)
			}
		}
		d.SetId(clusterOverrideName)
		return diags
	}

	if clusterOverrideObj.ShareMode == share.ShareModeAll {
		log.Println("cluster override share mode is 'all' so first unassign from 'all'")
		// cluster override share mode is 'all' so first unassign from 'all'
		_, err := clusteroverride.UnassignClusterOverrideFromProjects(clusterOverrideObj.Name, overrideType, projectObj.ID, share.ShareModeAll, []string{})
		if err != nil {
			log.Printf("cluster override share setting had all, but failed to unshare form all projects")
			return diag.FromErr(err)
		}
	}

	// compare incoming spec with cluster override sharing data
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
			// does not exist in cluster override project list
			newIds = append(newIds, inProj.Id)
			log.Println("append newIds", inProj.Id)
		}
	}

	if len(newIds) > 0 {
		log.Println("cluster override share to project ids ", newIds)
		_, err = clusteroverride.AssignClusterOverrideToProjects(clusterOverrideObj.Name, overrideType, projectObj.ID, share.ShareModeCustom, newIds)
		if err != nil {
			log.Printf("failed to share cluster override to new projects")
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
		log.Println("cluster override unshare from project ids ", oldIds)
		_, err = clusteroverride.UnassignClusterOverrideFromProjects(clusterOverrideObj.Name, overrideType, projectObj.ID, share.ShareModeCustom, oldIds)
		if err != nil {
			log.Printf("failed to unshare cluster override from old projects")
			return diag.FromErr(err)
		}
	}

	d.SetId(clusterOverrideName)
	return diags
}

func resourceClusterOverrideSharingRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var projs []*commonpb.ProjectMeta
	var sortprojs []*commonpb.ProjectMeta
	var sharingSpec *commonpb.SharingSpec

	clusterOverrideName := d.Get("clusteroverridename").(string)
	projectName := d.Get("project").(string)
	clusterOverrideType := d.Get("clusteroverridetype").(string)

	if d.State() != nil && d.State().ID != "" {
		if clusterOverrideName != d.State().ID {
			log.Println("detected clusterOverrideName change ", clusterOverrideName, d.State().ID)
			return diag.FromErr(fmt.Errorf("cannot change name of the clusteroverride"))
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

	clusterOverrideObj, errGet := clusteroverride.GetClusterOverride(clusterOverrideName, projectObj.ID, clusterOverrideType)
	if errGet != nil {
		log.Printf("failed to get cluster override info %s", errGet.Error())
		return diag.FromErr(errGet)
	}
	if clusterOverrideObj == nil {
		log.Printf("failed to get cluster override info")
		return diag.FromErr(fmt.Errorf("failed to get cluster override info"))
	}

	log.Println("clusterOverrideObj share type", clusterOverrideObj.ShareMode)
	for _, p := range clusterOverrideObj.Projects {
		if string(p.ProjectID) == projectObj.ID {
			//skip owner/parent projects
			continue
		}
		pName, err := config.GetProjectNameById(string(p.ProjectID))
		if err == nil {
			var prj commonpb.ProjectMeta
			prj.Id = string(p.ProjectID)
			prj.Name = pName
			log.Println("cluster override project list info: ", p.ProjectID, pName)
			projs = append(projs, &prj)
		} else {
			log.Println("get project name from cluster override project list failed: ", p.ProjectID, err.Error())
		}
	}

	// try to order cluster override list based on local state
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

	err = d.Set("clusteroverridename", clusterOverrideName)
	if err != nil {
		return diag.FromErr(err)
	}
	err = d.Set("project", projectName)
	if err != nil {
		return diag.FromErr(err)
	}

	err = flattenClusterOverrideSharing(d, clusterOverrideObj, sortprojs)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func flattenClusterOverrideSharingSpec(in *commonpb.SharingSpec, p []interface{}) ([]interface{}, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "flattenClusterOverrideSharingSpec empty input")
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}
	obj["all"] = in.Enabled
	if len(in.Projects) > 0 {
		obj["projects"] = flattenProjectMeta(in.Projects, true)
	}

	return []interface{}{obj}, nil
}

func flattenClusterOverrideSharing(d *schema.ResourceData, in *models.ClusterOverride, projs []*commonpb.ProjectMeta) error {
	var inSharing commonpb.SharingSpec

	if in.ShareMode == share.ShareModeAll {
		inSharing.Enabled = true
	} else {
		inSharing.Enabled = false
	}

	if len(in.Projects) > 1 {
		inSharing.Projects = projs
	}

	v, ok := d.Get("sharing").([]interface{})
	if !ok {
		v = []interface{}{}
	}

	var ret []interface{}
	ret, err := flattenClusterOverrideSharingSpec(&inSharing, v)
	if err != nil {
		return err
	}

	// XXX Debug
	w1 := spew.Sprintf("%+v", ret)
	log.Println("flattenClusterOverrideSharing after ", w1)

	err = d.Set("sharing", ret)
	if err != nil {
		log.Println("failed to set sharing")
		return err
	}

	return nil
}

func resourceClusterOverrideSharingUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("resource user update id %s", d.Id())
	return resourceClusterOverrideSharingUpsert(ctx, d, false)
}

func resourceClusterOverrideSharingDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	clusterOverrideName := d.Get("clusteroverridename").(string)
	projectName := d.Get("project").(string)
	clusterOverrideType := d.Get("clusteroverridetype").(string)

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

	clusterOverrideObj, errGet := clusteroverride.GetClusterOverride(clusterOverrideName, projectObj.ID, clusterOverrideType)
	if errGet != nil {
		log.Printf("failed to get cluster override info %s", errGet.Error())
		return diag.FromErr(errGet)
	}
	if clusterOverrideObj == nil {
		log.Printf("failed to get cluster override info")
		return diag.FromErr(fmt.Errorf("failed to get cluster override info"))
	}

	_, err = clusteroverride.UnassignClusterOverrideFromProjects(clusterOverrideObj.Name, clusterOverrideType, projectObj.ID, share.ShareModeAll, []string{})
	if err != nil {
		log.Printf("cluster override share setting had all, but failed to unshare form all projects")
		return diag.FromErr(err)
	}

	return diags
}

func expandClusterOverrideSharingSpec(p []interface{}) *commonpb.SharingSpec {
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
		for _, inProj := range obj.Projects {
			pID, err := config.GetProjectIdByName(inProj.Name)
			if err != nil {
				log.Println("failed to get project id by name ", inProj.Name)
			} else {
				inProj.Id = pID
			}
		}
	}

	log.Println("expandClusterOverrideSharingSpec obj", obj)
	return &obj
}
