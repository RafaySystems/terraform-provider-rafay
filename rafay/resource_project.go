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
		log.Printf("Project create got error, perform cleanup")
		pr, err := expandProject(d)
		if err != nil {
			log.Printf("Project expandProject error")
			return diags
		}
		auth := config.GetConfig().GetAppAuthProfile()
		client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
		if err != nil {
			return diags
		}

		err = client.SystemV3().Project().Delete(ctx, options.DeleteOptions{
			Name:    pr.Metadata.Name,
			Project: pr.Metadata.Project,
		})
		if err != nil {
			return diags
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

	in := p[0].(map[string]interface{})

	if v, ok := in["cluster_resource_quota"].([]interface{}); ok {
		obj.ClusterResourceQuota = expandProjectResourceQuota(v)
	}

	if v, ok := in["default_cluster_namespace_quota"].([]interface{}); ok {
		obj.DefaultClusterNamespaceQuota = expandProjectResourceQuota(v)
	}

	return obj, nil
}

func expandProjectResourceQuota(p []interface{}) *systempb.ProjectResourceQuota {
	obj := &systempb.ProjectResourceQuota{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["cpu_requests"].(string); ok && len(v) > 0 {
		obj.CpuRequests = expandQuantityString(v)
	}

	if v, ok := in["memory_requests"].(string); ok && len(v) > 0 {
		obj.MemoryRequests = expandQuantityString(v)
	}

	if v, ok := in["cpu_limits"].(string); ok && len(v) > 0 {
		obj.CpuLimits = expandQuantityString(v)
	}

	if v, ok := in["memory_limits"].(string); ok && len(v) > 0 {
		obj.MemoryLimits = expandQuantityString(v)
	}

	if v, ok := in["config_maps"].(string); ok && len(v) > 0 {
		obj.ConfigMaps = expandQuantityString(v)
	}

	if v, ok := in["persistent_volume_claims"].(string); ok && len(v) > 0 {
		obj.PersistentVolumeClaims = expandQuantityString(v)
	}

	if v, ok := in["secrets"].(string); ok && len(v) > 0 {
		obj.Secrets = expandQuantityString(v)
	}

	if v, ok := in["services"].(string); ok && len(v) > 0 {
		obj.Services = expandQuantityString(v)
	}

	if v, ok := in["services_load_balancers"].(string); ok && len(v) > 0 {
		obj.ServicesLoadBalancers = expandQuantityString(v)
	}

	if v, ok := in["services_node_ports"].(string); ok && len(v) > 0 {
		obj.ServicesNodePorts = expandQuantityString(v)
	}

	if v, ok := in["storage_requests"].(string); ok && len(v) > 0 {
		obj.StorageRequests = expandQuantityString(v)
	}

	if v, ok := in["pods"].(string); ok && len(v) > 0 {
		obj.Pods = expandQuantityString(v)
	}

	if v, ok := in["replication_controllers"].(string); ok && len(v) > 0 {
		obj.ReplicationControllers = expandQuantityString(v)
	}

	return obj
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

	if in.ClusterResourceQuota != nil {
		obj["cluster_resource_quota"] = flattenProjectResourceQuota(in.ClusterResourceQuota)
	}

	if in.DefaultClusterNamespaceQuota != nil {
		obj["default_cluster_namespace_quota"] = flattenProjectResourceQuota(in.DefaultClusterNamespaceQuota)
	}

	return []interface{}{obj}, nil
}

func flattenProjectResourceQuota(in *systempb.ProjectResourceQuota) []interface{} {
	if in == nil {
		return nil
	}

	retNil := true
	obj := make(map[string]interface{})

	if in.ConfigMaps != nil {
		obj["config_maps"] = in.ConfigMaps.String()
		retNil = false
	}
	if in.CpuLimits != nil {
		obj["cpu_limits"] = in.CpuLimits.String()
		retNil = false
	}
	if in.CpuRequests != nil {
		obj["cpu_requests"] = in.CpuRequests.String()
		retNil = false
	}
	if in.MemoryLimits != nil {
		obj["memory_limits"] = in.MemoryLimits.String()
		retNil = false
	}
	if in.MemoryRequests != nil {
		obj["memory_requests"] = in.MemoryRequests.String()
		retNil = false
	}
	if in.PersistentVolumeClaims != nil {
		obj["persistent_volume_claims"] = in.PersistentVolumeClaims.String()
		retNil = false
	}
	if in.Pods != nil {
		obj["pods"] = in.Pods.String()
		retNil = false
	}
	if in.ReplicationControllers != nil {
		obj["replication_controllers"] = in.ReplicationControllers.String()
		retNil = false
	}
	if in.Secrets != nil {
		obj["secrets"] = in.Secrets.String()
		retNil = false
	}
	if in.Services != nil {
		obj["services"] = in.Services.String()
		retNil = false
	}
	if in.ServicesLoadBalancers != nil {
		obj["services_load_balancers"] = in.ServicesLoadBalancers.String()
		retNil = false
	}
	if in.ServicesNodePorts != nil {
		obj["services_node_ports"] = in.ServicesNodePorts.String()
		retNil = false
	}
	if in.StorageRequests != nil {
		obj["storage_requests"] = in.StorageRequests.String()
		retNil = false
	}

	if retNil {
		return nil
	}

	return []interface{}{obj}
}

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
