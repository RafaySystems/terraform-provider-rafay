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
	"github.com/RafaySystems/rafay-common/proto/types/hub/systempb"
	"github.com/RafaySystems/rctl/pkg/config"
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
	log.Printf("Project create starts")
	diags := resourceProjectUpsert(ctx, d, m)
	if diags.HasError() {
		tflog := os.Getenv("TF_LOG")
		if tflog == "TRACE" || tflog == "DEBUG" {
			ctx = context.WithValue(ctx, "debug", "true")
		}
		log.Printf("Project create got error, perform cleanup")
		pr, err := expandProject(d)
		if err != nil {
			log.Printf("Project expandProject error")
			return diag.FromErr(err)
		}
		auth := config.GetConfig().GetAppAuthProfile()
		client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
		if err != nil {
			return diag.FromErr(err)
		}

		err = client.SystemV3().Project().Delete(ctx, options.DeleteOptions{
			Name:    pr.Metadata.Name,
			Project: pr.Metadata.Project,
		})
		if err != nil {
			return diag.FromErr(err)
		}
	}
	return diags
}

func resourceProjectUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("Project update starts")
	return resourceProjectUpsert(ctx, d, m)
}

func resourceProjectUpsert(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("Project upsert starts")
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	pr, err := expandProject(d)
	if err != nil {
		log.Printf("Project expandProject error")
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.SystemV3().Project().Apply(ctx, pr, options.ApplyOptions{})
	if err != nil {
		// XXX Debug
		n1 := spew.Sprintf("%+v", pr)
		log.Println("Project apply Project:", n1)
		log.Printf("Project apply error")
		return diag.FromErr(err)
	}

	d.SetId(pr.Metadata.Name)
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

	// println("resourceProjectDelete project ", project)
	// auth := config.GetConfig().GetAppAuthProfile()
	// client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	// if err != nil {
	// 	return diag.FromErr(err)
	// }

	// err = client.SystemV3().Project().Delete(ctx, options.DeleteOptions{
	// 	Name:    Project.Metadata.Name,
	// 	Project: Project.Metadata.Project,
	// })
	// log.Printf("resourceProjectDelete ", err)

	//v3 spec gave error try v2
	return resourceProjectV2Delete(ctx, project)

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

	Project, err := client.SystemV3().Project().Get(ctx, options.GetOptions{
		Name:    tfProjectState.Metadata.Name,
		Project: tfProjectState.Metadata.Project,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	// XXX Debug
	// w1 = spew.Sprintf("%+v", wl)
	// log.Println("resourceProjectRead wl", w1)

	err = flattenProject(d, Project)
	if err != nil {
		return diag.FromErr(err)
	}
	return diags

}

func expandProject(in *schema.ResourceData) (*systempb.Project, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "expand Project empty input")
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

	obj.ApiVersion = "system.k8smgmt.io/v3"
	obj.Kind = "Project"
	return obj, nil
}

func expandProjectSpec(p []interface{}) (*systempb.ProjectSpec, error) {
	obj := &systempb.ProjectSpec{}
	if len(p) == 0 || p[0] == nil {
		return obj, fmt.Errorf("%s", "expandProjectSpec empty input")
	}

	// Force dafult to false, to avoid conflict with system default project
	obj.Default = false

	return obj, nil
}

func resourceProjectV2Delete(ctx context.Context, projectp *systempb.Project) diag.Diagnostics {
	var diags diag.Diagnostics

	//log.Printf("resourceProjectV2Delete")
	projectId, err := config.GetProjectIdByName(projectp.Metadata.Name)
	if err != nil {
		return diag.FromErr(err)
	}
	err = project.DeleteProjectById(projectId)
	if err != nil {
		log.Printf("delete project error %s", err.Error())
		return diag.FromErr(err)
	}

	return diags
}

// Flatteners

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

	var ret []interface{}
	ret, err = flattenProjectSpec(in.Spec, v)
	if err != nil {
		return err
	}

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

	obj["default"] = false

	return []interface{}{obj}, nil
}
