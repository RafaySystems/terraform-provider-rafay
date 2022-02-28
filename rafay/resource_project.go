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
	"github.com/RafaySystems/rafay-common/pkg/hub/terraform/resource"
	"github.com/RafaySystems/rafay-common/proto/types/hub/systempb"
	"github.com/RafaySystems/rctl/pkg/config"
	"github.com/RafaySystems/rctl/pkg/models"
	"github.com/RafaySystems/rctl/pkg/project"
	"github.com/RafaySystems/rctl/pkg/versioninfo"
	"github.com/davecgh/go-spew/spew"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceProject() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceProjectCreate,
		ReadContext:   resourceProjectRead,
		UpdateContext: resourceProjectUpdate,
		DeleteContext: resourceProjectDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema:        resource.ProjectSchema.Schema,
	}
}

func resourceProjectCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("project create starts")
	diags := resourceProjectUpsert(ctx, d, m)
	if diags.HasError() {
		tflog := os.Getenv("TF_LOG")
		if tflog == "TRACE" || tflog == "DEBUG" {
			ctx = context.WithValue(ctx, "debug", "true")
		}
		log.Printf("project create got error, perform cleanup")
		bp, err := expandProject(d)
		if err != nil {
			log.Printf("project expandProject error")
			return diag.FromErr(err)
		}
		auth := config.GetConfig().GetAppAuthProfile()
		client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
		if err != nil {
			return diag.FromErr(err)
		}

		err = client.SystemV3().Project().Delete(ctx, options.DeleteOptions{
			Name:    bp.Metadata.Name,
			Project: bp.Metadata.Project,
		})
		if err != nil {
			return diag.FromErr(err)
		}
	}
	return diags
}

func resourceProjectUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("project update starts")
	return resourceProjectUpsert(ctx, d, m)
}

func resourceProjectUpsert(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("project upsert starts")
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	project, err := expandProject(d)
	if err != nil {
		log.Printf("project expandProject error")
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.SystemV3().Project().Apply(ctx, project, options.ApplyOptions{})
	if err != nil {
		// XXX Debug
		n1 := spew.Sprintf("%+v", project)
		log.Println("project apply project:", n1)
		log.Printf("project apply error")
		return diag.FromErr(err)
	}

	d.SetId(project.Metadata.Name)
	return diags

}

func resourceProjectRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	log.Println("resourceProjectRead ")
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}
	tfProjectState, err := expandProject(d)
	if err != nil {
		return diag.FromErr(err)
	}

	// XXX Debug
	// w1 := spew.Sprintf("%+v", tfProjectState)
	// log.Println("resourceProjectRead tfProjectState", w1)

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}

	project, err := client.SystemV3().Project().Get(ctx, options.GetOptions{
		Name:    tfProjectState.Metadata.Name,
		Project: tfProjectState.Metadata.Name,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	// XXX Debug
	// w1 = spew.Sprintf("%+v", wl)
	// log.Println("resourceProjectRead wl", w1)

	err = flattenProject(d, project)
	if err != nil {
		return diag.FromErr(err)
	}
	return diags

}

func resourceProjectDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	project, err := expandProject(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.SystemV3().Project().Delete(ctx, options.DeleteOptions{
		Name:    project.Metadata.Name,
		Project: project.Metadata.Name,
	})

	if err != nil {
		//v3 spec gave error try v2
		return resourceProjectV2Delete(ctx, project)
	}

	return diags
}

func resourceProjectV2Delete(ctx context.Context, projectp *systempb.Project) diag.Diagnostics {
	var diags diag.Diagnostics

	projectId, err := config.GetProjectIdByName(projectp.Metadata.Name)
	if err != nil {
		return diag.FromErr(err)
	}

	errDel := project.DeleteProjectByName(projectId)
	if errDel != nil {
		fmt.Printf("error while deleting project %s", errDel.Error())
		return diag.FromErr(errDel)
	}

	return diags
}

// expand functions

func expandProject(in *schema.ResourceData) (*systempb.Project, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "expand project empty input")
	}
	obj := &systempb.Project{}

	if v, ok := in.Get("metadata").([]interface{}); ok {
		obj.Metadata = expandMetaData(v)
	}

	if v, ok := in.Get("spec").([]interface{}); ok {
		objSpec, err := expandProjectSpec(v)
		if err != nil {
			return nil, err
		}
		log.Println("expandProjectSpec got spec")
		obj.Spec = objSpec
	}

	obj.ApiVersion = "infra.k8smgmt.io/v3"
	obj.Kind = "Project"
	return obj, nil
}

func expandProjectSpec(p []interface{}) (*systempb.ProjectSpec, error) {
	obj := &systempb.ProjectSpec{}
	if len(p) == 0 || p[0] == nil {
		return obj, fmt.Errorf("%s", "expandProjectSpec empty input")
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["default"].(bool); ok {
		obj.Default = v
	}

	return obj, nil
}

// flatten functions

func flattenProject(d *schema.ResourceData, in *systempb.Project) error {
	if in == nil {
		return nil
	}

	err := d.Set("metadata", flattenMetaData(in.Metadata))
	if err != nil {
		return err
	}

	v, ok := d.Get("spec").([]interface{})
	if !ok {
		v = []interface{}{}
	}

	// XXX Debug
	// w1 := spew.Sprintf("%+v", v)
	// log.Println("flattenProject before ", w1)
	var ret []interface{}
	ret, err = flattenProjectSpec(in.Spec, v)
	if err != nil {
		return err
	}
	// XXX Debug
	// w1 = spew.Sprintf("%+v", ret)
	// log.Println("flattenProject after ", w1)

	err = d.Set("spec", ret)
	if err != nil {
		return err
	}
	return nil
}

func flattenProjectSpec(in *systempb.ProjectSpec, p []interface{}) ([]interface{}, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "flattenProjectSpec empty input")
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if in.Default {
		obj["default"] = in.Default
	}

	// XXX Debug
	// w1 := spew.Sprintf("%+v", v)
	// log.Println("flattenProjectSpec before ", w1)

	// XXX Debug
	// w1 = spew.Sprintf("%+v", ret)
	// log.Println("flattenProjectSpec after ", w1)

	return []interface{}{obj}, nil
}

// to be deprecated as we upgrade all terraform resources to v3.

func getProjectById(id string) (string, error) {
	log.Printf("get project by id %s", id)
	auth := config.GetConfig().GetAppAuthProfile()
	uri := "/auth/v1/projects/"
	uri = uri + fmt.Sprintf("%s/", id)
	return auth.AuthAndRequest(uri, "GET", nil)
}

func getProjectFromResponse(json_data []byte) (*models.Project, error) {
	var pr models.Project
	if err := json.Unmarshal(json_data, &pr); err != nil {
		return nil, err
	}
	return &pr, nil
}
