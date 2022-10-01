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
	"github.com/RafaySystems/rafay-common/proto/types/hub/commonpb"
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
		Importer: &schema.ResourceImporter{
			State: resourceProjectImport,
		},

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

	if d.State() != nil && d.State().ID != "" {
		n := GetMetaName(d)
		if n != "" && n != d.State().ID {
			log.Printf("Project name change not supported")
			d.State().Tainted = true
			return diag.FromErr(fmt.Errorf("%s", "project name change not supported"))
		}
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

	// projectId, err := config.GetProjectIdByName(pr.Metadata.Name)
	// if err != nil {
	// 	return diag.FromErr(err)
	// }

	d.SetId(pr.Metadata.Name)
	return diags

}

func resourceProjectDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	//var diags diag.Diagnostics
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
}

func resourceProjectRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	log.Println("resourceProjectRead ")
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	meta := GetMetaData(d)
	if meta == nil {
		return diag.FromErr(fmt.Errorf("%s", "failed to read resource "))
	}
	if d.State() != nil && d.State().ID != "" {
		meta.Name = d.State().ID
	}

	// tfProjectState, err := expandProject(d)
	// if err != nil {
	// 	return diag.FromErr(err)
	// }

	// XXX Debug
	// w1 := spew.Sprintf("%+v", tfProjectState)
	// log.Println("resourceProjectRead tfProjectState", w1)

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}

	Project, err := client.SystemV3().Project().Get(ctx, options.GetOptions{
		Name: meta.Name,
		//Project: meta.Project,
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
		//obj.CpuRequests = expandQuantityString(v)
		obj.CpuRequests = v

	}

	if v, ok := in["memory_requests"].(string); ok && len(v) > 0 {
		//obj.MemoryRequests = expandQuantityString(v)
		obj.MemoryRequests = v
	}

	if v, ok := in["cpu_limits"].(string); ok && len(v) > 0 {
		//obj.CpuLimits = expandQuantityString(v)
		obj.CpuLimits = v
	}

	if v, ok := in["memory_limits"].(string); ok && len(v) > 0 {
		//obj.MemoryLimits = expandQuantityString(v)
		obj.MemoryLimits = v
	}

	if v, ok := in["config_maps"].(string); ok && len(v) > 0 {
		//obj.ConfigMaps = expandQuantityString(v)
		obj.ConfigMaps = v
	}

	if v, ok := in["persistent_volume_claims"].(string); ok && len(v) > 0 {
		//obj.PersistentVolumeClaims = expandQuantityString(v)
		obj.PersistentVolumeClaims = v
	}

	if v, ok := in["secrets"].(string); ok && len(v) > 0 {
		//obj.Secrets = expandQuantityString(v)
		obj.Secrets = v
	}

	if v, ok := in["services"].(string); ok && len(v) > 0 {
		//obj.Services = expandQuantityString(v)
		obj.Services = v
	}

	if v, ok := in["services_load_balancers"].(string); ok && len(v) > 0 {
		//obj.ServicesLoadBalancers = expandQuantityString(v)
		obj.ServicesLoadBalancers = v
	}

	if v, ok := in["services_node_ports"].(string); ok && len(v) > 0 {
		//obj.ServicesNodePorts = expandQuantityString(v)
		obj.ServicesNodePorts = v
	}

	if v, ok := in["storage_requests"].(string); ok && len(v) > 0 {
		//obj.StorageRequests = expandQuantityString(v)
		obj.StorageRequests = v
	}

	if v, ok := in["pods"].(string); ok && len(v) > 0 {
		//obj.Pods = expandQuantityString(v)
		obj.Pods = v
	}

	if v, ok := in["replication_controllers"].(string); ok && len(v) > 0 {
		//obj.ReplicationControllers = expandQuantityString(v)
		obj.ReplicationControllers = v
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

	// if in.ConfigMaps != nil {
	// 	obj["config_maps"] = in.ConfigMaps.String()
	// 	retNil = false
	// }
	// if in.CpuLimits != nil {
	// 	obj["cpu_limits"] = in.CpuLimits.String()
	// 	retNil = false
	// }
	// if in.CpuRequests != nil {
	// 	obj["cpu_requests"] = in.CpuRequests.String()
	// 	retNil = false
	// }
	// if in.MemoryLimits != nil {
	// 	obj["memory_limits"] = in.MemoryLimits.String()
	// 	retNil = false
	// }
	// if in.MemoryRequests != nil {
	// 	obj["memory_requests"] = in.MemoryRequests.String()
	// 	retNil = false
	// }
	// if in.PersistentVolumeClaims != nil {
	// 	obj["persistent_volume_claims"] = in.PersistentVolumeClaims.String()
	// 	retNil = false
	// }
	// if in.Pods != nil {
	// 	obj["pods"] = in.Pods.String()
	// 	retNil = false
	// }
	// if in.ReplicationControllers != nil {
	// 	obj["replication_controllers"] = in.ReplicationControllers.String()
	// 	retNil = false
	// }
	// if in.Secrets != nil {
	// 	obj["secrets"] = in.Secrets.String()
	// 	retNil = false
	// }
	// if in.Services != nil {
	// 	obj["services"] = in.Services.String()
	// 	retNil = false
	// }
	// if in.ServicesLoadBalancers != nil {
	// 	obj["services_load_balancers"] = in.ServicesLoadBalancers.String()
	// 	retNil = false
	// }
	// if in.ServicesNodePorts != nil {
	// 	obj["services_node_ports"] = in.ServicesNodePorts.String()
	// 	retNil = false
	// }
	// if in.StorageRequests != nil {
	// 	obj["storage_requests"] = in.StorageRequests.String()
	// 	retNil = false
	// }

	if len(in.ConfigMaps) > 0 {
		obj["type"] = in.ConfigMaps
		retNil = false
	}

	if len(in.CpuLimits) > 0 {
		obj["cpu_limits"] = in.CpuLimits
		retNil = false
	}
	if len(in.CpuRequests) > 0 {
		obj["cpu_requests"] = in.CpuRequests
		retNil = false
	}
	if len(in.MemoryLimits) > 0 {
		obj["memory_limits"] = in.MemoryLimits
		retNil = false
	}
	if len(in.MemoryRequests) > 0 {
		obj["memory_requests"] = in.MemoryRequests
		retNil = false
	}
	if len(in.PersistentVolumeClaims) > 0 {
		obj["persistent_volume_claims"] = in.PersistentVolumeClaims
		retNil = false
	}
	if len(in.Pods) > 0 {
		obj["pods"] = in.Pods
		retNil = false
	}
	if len(in.ReplicationControllers) > 0 {
		obj["replication_controllers"] = in.ReplicationControllers
		retNil = false
	}
	if len(in.Secrets) > 0 {
		obj["secrets"] = in.Secrets
		retNil = false
	}
	if len(in.Services) > 0 {
		obj["services"] = in.Services
		retNil = false
	}
	if len(in.ServicesLoadBalancers) > 0 {
		obj["services_load_balancers"] = in.ServicesLoadBalancers
		retNil = false
	}
	if len(in.ServicesNodePorts) > 0 {
		obj["services_node_ports"] = in.ServicesNodePorts
		retNil = false
	}
	if len(in.StorageRequests) > 0 {
		obj["storage_requests"] = in.StorageRequests
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

func resourceProjectImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	//d_debug := spew.Sprintf("%+v", d)
	log.Println("resourceProjectImport d.Id:", d.Id())
	//log.Println("resourceProjectImport d_debug", d_debug)

	project := &systempb.Project{}

	var metaD commonpb.Metadata
	metaD.Name = d.Id()
	project.Metadata = &metaD

	err := d.Set("metadata", flattenMetaData(project.Metadata))
	if err != nil {
		return nil, err
	}

	d.SetId(project.Metadata.Name)

	return []*schema.ResourceData{d}, nil
}
