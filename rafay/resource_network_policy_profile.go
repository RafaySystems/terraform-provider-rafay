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
	"github.com/RafaySystems/rafay-common/proto/types/hub/securitypb"
	"github.com/RafaySystems/rctl/pkg/config"
	"github.com/RafaySystems/rctl/pkg/versioninfo"
	"github.com/davecgh/go-spew/spew"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceNetworkPolicyProfile() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceNetworkPolicyProfileCreate,
		ReadContext:   resourceNetworkPolicyProfileRead,
		UpdateContext: resourceNetworkPolicyProfileUpdate,
		DeleteContext: resourceNetworkPolicyProfileDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema:        resource.NetworkPolicyProfileSchema.Schema,
	}
}

func resourceNetworkPolicyProfileCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("NetworkPolicyProfile create starts")
	diags := resourceNetworkPolicyProfileUpsert(ctx, d, m)
	if diags.HasError() {
		tflog := os.Getenv("TF_LOG")
		if tflog == "TRACE" || tflog == "DEBUG" {
			ctx = context.WithValue(ctx, "debug", "true")
		}
		log.Printf("namespaceNetworkPolicy create got error, perform cleanup")
		npp, err := expandNetworkPolicyProfile(d)
		if err != nil {
			log.Printf("namespaceNetworkPolicy expandNetworkPolicyProfile error")
			return diags
		}
		auth := config.GetConfig().GetAppAuthProfile()
		client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
		if err != nil {
			return diags
		}

		err = client.SecurityV3().NetworkPolicyProfile().Delete(ctx, options.DeleteOptions{
			Name:    npp.Metadata.Name,
			Project: npp.Metadata.Project,
		})
		if err != nil {
			return diags
		}
	}
	return diags
}

func resourceNetworkPolicyProfileUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("namespaceNetworkPolicy update starts")
	return resourceNetworkPolicyProfileUpsert(ctx, d, m)
}

func resourceNetworkPolicyProfileUpsert(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("namespaceNetworkPolicy upsert starts")
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

	namespaceNetworkPolicy, err := expandNetworkPolicyProfile(d)
	if err != nil {
		log.Printf("namespaceNetworkPolicy expandNetworkPolicyProfile error")
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.SecurityV3().NetworkPolicyProfile().Apply(ctx, namespaceNetworkPolicy, options.ApplyOptions{})
	if err != nil {
		// XXX Debug
		n1 := spew.Sprintf("%+v", namespaceNetworkPolicy)
		log.Println("namespaceNetworkPolicy apply namespaceNetworkPolicy:", n1)
		log.Printf("namespaceNetworkPolicy apply error")
		return diag.FromErr(err)
	}

	d.SetId(namespaceNetworkPolicy.Metadata.Name)
	return diags

}

func resourceNetworkPolicyProfileRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	log.Println("resourceNetworkPolicyProfileRead ")
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}
	tfNetworkPolicyProfileState, err := expandNetworkPolicyProfile(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}

	npp, err := client.SecurityV3().NetworkPolicyProfile().Get(ctx, options.GetOptions{
		Name:    tfNetworkPolicyProfileState.Metadata.Name,
		Project: tfNetworkPolicyProfileState.Metadata.Project,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	err = flattenNetworkPolicyProfile(d, npp)
	if err != nil {
		return diag.FromErr(err)
	}
	return diags

}

func resourceNetworkPolicyProfileDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	npp, err := expandNetworkPolicyProfile(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.SecurityV3().NetworkPolicyProfile().Delete(ctx, options.DeleteOptions{
		Name:    npp.Metadata.Name,
		Project: npp.Metadata.Project,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func expandNetworkPolicyProfile(in *schema.ResourceData) (*securitypb.NetworkPolicyProfile, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "expand namespaceNetworkPolicy empty input")
	}
	obj := &securitypb.NetworkPolicyProfile{}

	if v, ok := in.Get("metadata").([]interface{}); ok && len(v) > 0 {
		obj.Metadata = expandMetaData(v)
	}

	if v, ok := in.Get("spec").([]interface{}); ok && len(v) > 0 {
		objSpec, err := expandNetworkPolicyProfileSpec(v)
		if err != nil {
			return nil, err
		}
		log.Println("expandNetworkPolicyProfileSpec got spec")
		obj.Spec = objSpec
	}

	obj.ApiVersion = "security.k8smgmt.io/v3"
	obj.Kind = "NetworkPolicyProfile"
	return obj, nil
}

func expandNetworkPolicyProfileSpec(p []interface{}) (*securitypb.NetworkPolicyProfileSpec, error) {
	obj := &securitypb.NetworkPolicyProfileSpec{}
	if len(p) == 0 || p[0] == nil {
		return obj, fmt.Errorf("%s", "expandNetworkPolicyProfileSpec empty input")
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["sharing"].([]interface{}); ok && len(v) > 0 {
		obj.Sharing = expandSharingSpec(v)
	}

	if v, ok := in["version"].(string); ok && len(v) > 0 {
		obj.Version = v
	}

	if v, ok := in["installation_params"].([]interface{}); ok && len(v) > 0 {
		obj.InstallationParams = expandNetworkPolicyProfileIP(v)
	}

	return obj, nil
}

func expandNetworkPolicyProfileIP(p []interface{}) *securitypb.InstallationParams {
	obj := &securitypb.InstallationParams{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["policy_enforcement_mode"].(string); ok && len(v) > 0 {
		obj.PolicyEnforcementMode = v
	}

	return obj

}

// Flatteners

func flattenNetworkPolicyProfile(d *schema.ResourceData, in *securitypb.NetworkPolicyProfile) error {
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
	ret, err = flattenNetworkPolicyProfileSpec(in.Spec, v)
	if err != nil {
		return err
	}

	err = d.Set("spec", ret)
	if err != nil {
		return err
	}
	return nil
}

func flattenNetworkPolicyProfileSpec(in *securitypb.NetworkPolicyProfileSpec, p []interface{}) ([]interface{}, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "flattenNetworkPolicyProfileSpec empty input")
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
		obj["installation_params"] = flattenNetworkPolicyProfileSpecIP(in.InstallationParams, v)
	}

	return []interface{}{obj}, nil
}

func flattenNetworkPolicyProfileSpecIP(in *securitypb.InstallationParams, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.PolicyEnforcementMode) > 0 {
		obj["policy_enforcement_mode"] = in.PolicyEnforcementMode
	}

	return []interface{}{obj}
}
