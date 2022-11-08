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
	"github.com/RafaySystems/rafay-common/proto/types/hub/servicemeshpb"
	"github.com/RafaySystems/rctl/pkg/config"
	"github.com/RafaySystems/rctl/pkg/versioninfo"
	"github.com/davecgh/go-spew/spew"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceMeshProfile() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceMeshProfileCreate,
		ReadContext:   resourceMeshProfileRead,
		UpdateContext: resourceMeshProfileUpdate,
		DeleteContext: resourceMeshProfileDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema:        resource.MeshProfileSchema.Schema,
	}
}

func resourceMeshProfileCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("MeshProfile create starts")
	diags := resourceMeshProfileUpsert(ctx, d, m)
	if diags.HasError() {
		tflog := os.Getenv("TF_LOG")
		if tflog == "TRACE" || tflog == "DEBUG" {
			ctx = context.WithValue(ctx, "debug", "true")
		}
		log.Printf("meshProfile create got error, perform cleanup")
		mp, err := expandMeshProfile(d)
		if err != nil {
			log.Printf("meshProfile expandMeshProfile error")
			return diags
		}
		auth := config.GetConfig().GetAppAuthProfile()
		client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
		if err != nil {
			return diags
		}

		err = client.ServicemeshV3().MeshProfile().Delete(ctx, options.DeleteOptions{
			Name:    mp.Metadata.Name,
			Project: mp.Metadata.Project,
		})
		if err != nil {
			return diags
		}
	}
	return diags
}

func resourceMeshProfileUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("meshProfile update starts")
	return resourceMeshProfileUpsert(ctx, d, m)
}

func resourceMeshProfileUpsert(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("meshProfile upsert starts")
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	if d.State() != nil && d.State().ID != "" {
		n := GetMetaName(d)
		if n != "" && n != d.State().ID {
			log.Printf("metadata name change not supported")
			d.State().Tainted = true
			return diag.FromErr(fmt.Errorf("%s", "metadata name change not supported"))
		}
	}

	meshProfile, err := expandMeshProfile(d)
	if err != nil {
		log.Printf("meshProfile expandMeshProfile error")
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.ServicemeshV3().MeshProfile().Apply(ctx, meshProfile, options.ApplyOptions{})
	if err != nil {
		// XXX Debug
		n1 := spew.Sprintf("%+v", meshProfile)
		log.Println("meshProfile apply meshProfile:", n1)
		log.Printf("meshProfile apply error")
		return diag.FromErr(err)
	}

	d.SetId(meshProfile.Metadata.Name)
	return diags

}

func resourceMeshProfileRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	log.Println("resourceMeshProfileRead ")
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

	tfMeshProfileState, err := expandMeshProfile(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}

	mp, err := client.ServicemeshV3().MeshProfile().Get(ctx, options.GetOptions{
		//Name:    tfMeshProfileState.Metadata.Name,
		Name:    meta.Name,
		Project: tfMeshProfileState.Metadata.Project,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	err = flattenMeshProfile(d, mp)
	if err != nil {
		return diag.FromErr(err)
	}
	return diags

}

func resourceMeshProfileDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	mp, err := expandMeshProfile(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.ServicemeshV3().MeshProfile().Delete(ctx, options.DeleteOptions{
		Name:    mp.Metadata.Name,
		Project: mp.Metadata.Project,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func expandMeshProfile(in *schema.ResourceData) (*servicemeshpb.MeshProfile, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "expand meshProfile empty input")
	}
	obj := &servicemeshpb.MeshProfile{}

	if v, ok := in.Get("metadata").([]interface{}); ok && len(v) > 0 {
		obj.Metadata = expandMetaData(v)
	}

	if v, ok := in.Get("spec").([]interface{}); ok && len(v) > 0 {
		objSpec, err := expandMeshProfileSpec(v)
		if err != nil {
			return nil, err
		}
		log.Println("expandMeshProfileSpec got spec")
		obj.Spec = objSpec
	}

	obj.ApiVersion = "servicemesh.k8smgmt.io/v3"
	obj.Kind = "MeshProfile"
	return obj, nil
}

func expandMeshProfileSpec(p []interface{}) (*servicemeshpb.MeshProfileSpec, error) {
	obj := &servicemeshpb.MeshProfileSpec{}
	if len(p) == 0 || p[0] == nil {
		return obj, fmt.Errorf("%s", "expandMeshProfileSpec empty input")
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["sharing"].([]interface{}); ok && len(v) > 0 {
		obj.Sharing = expandSharingSpec(v)
	}

	if v, ok := in["version"].(string); ok && len(v) > 0 {
		obj.Version = v
	}

	if v, ok := in["installation_params"].([]interface{}); ok && len(v) > 0 {
		obj.InstallationParams = expandMeshProfileIP(v)
	}

	return obj, nil
}

func expandMeshProfileIP(p []interface{}) *servicemeshpb.InstallationParams {
	obj := &servicemeshpb.InstallationParams{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["cert_type"].(string); ok {
		obj.CertType = v
	}

	if v, ok := in["enable_ingress"].(bool); ok {
		obj.EnableIngress = v
	}

	if v, ok := in["enable_namespaces_by_default"].(bool); ok {
		obj.EnableNamespacesByDefault = v
	}

	if v, ok := in["resource_quota"].([]interface{}); ok && len(v) > 0 {
		log.Println("resource_quotas v", v)
		obj.ResourceQuota = expandMeshResourceQuotas(v)
		log.Println("obj.ResourceQuotas ", obj.ResourceQuota)
	}

	return obj

}

func expandMeshResourceQuotas(p []interface{}) *servicemeshpb.MeshResourceQuotas {
	obj := &servicemeshpb.MeshResourceQuotas{}
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

	log.Println("expandNamespaceResourceQuotas obj ", obj)
	return obj
}

// Flatteners

func flattenMeshProfile(d *schema.ResourceData, in *servicemeshpb.MeshProfile) error {
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
	ret, err = flattenMeshProfileSpec(in.Spec, v)
	if err != nil {
		return err
	}

	err = d.Set("spec", ret)
	if err != nil {
		return err
	}
	return nil
}

func flattenMeshProfileSpec(in *servicemeshpb.MeshProfileSpec, p []interface{}) ([]interface{}, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "flattenMeshProfileSpec empty input")
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if in.Sharing != nil {
		obj["sharing"] = flattenSharingSpec(in.Sharing)
	}

	if len(in.Version) > 0 {
		obj["version"] = in.Version
	}

	if in.InstallationParams != nil {
		v, ok := obj["installation_params"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["installation_params"] = flattenMeshProfileSpecIP(in.InstallationParams, v)
	}

	return []interface{}{obj}, nil
}

func flattenMeshProfileSpecIP(in *servicemeshpb.InstallationParams, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.CertType) > 0 {
		obj["cert_type"] = in.CertType
	}

	if in.EnableIngress {
		obj["enable_ingress"] = in.EnableIngress
	}

	if in.EnableNamespacesByDefault {
		obj["enable_namespaces_by_default"] = in.EnableNamespacesByDefault
	}

	if in.ResourceQuota != nil {
		obj["resource_quota"] = flattenMeshResourceQuotas(in.ResourceQuota)
	}

	return []interface{}{obj}
}

func flattenMeshResourceQuotas(in *servicemeshpb.MeshResourceQuotas) []interface{} {
	if in == nil {
		return nil
	}

	retNil := true
	obj := make(map[string]interface{})

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

	if retNil {
		return nil
	}

	return []interface{}{obj}
}
