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
	"github.com/RafaySystems/rafay-common/proto/types/hub/appspb"
	"github.com/RafaySystems/rctl/pkg/config"
	"github.com/RafaySystems/rctl/pkg/user"
	"github.com/RafaySystems/rctl/pkg/versioninfo"
	"github.com/RafaySystems/rctl/pkg/workloadtemplate"
	"github.com/davecgh/go-spew/spew"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceWorkloadTemplate() *schema.Resource {
	modSchema := resource.WorkloadTemplateSchema.Schema
	modSchema["impersonate"] = &schema.Schema{
		Description: "impersonate user",
		Optional:    true,
		Type:        schema.TypeString,
	}
	return &schema.Resource{
		CreateContext: resourceWorkloadTemplateCreate,
		ReadContext:   resourceWorkloadTemplateRead,
		UpdateContext: resourceWorkloadTemplateUpdate,
		DeleteContext: resourceWorkloadTemplateDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema:        modSchema,
	}
}

func resourceWorkloadTemplateCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("workloadtemplate create starts")
	diags := resourceWorkloadTemplateUpsert(ctx, d, m)
	if diags.HasError() {
		tflog := os.Getenv("TF_LOG")
		if tflog == "TRACE" || tflog == "DEBUG" {
			ctx = context.WithValue(ctx, "debug", "true")
		}
		log.Printf("workloadtemplate create got error, perform cleanup")
		ag, err := expandWorkloadTemplate(d)
		if err != nil {
			log.Printf("workloadtemplate expandWorkloadTemplate error")
			return diags
		}

		if v, ok := d.Get("impersonate").(string); ok && len(v) > 0 {
			defer ResetImpersonateUser()
			asUser := d.Get("impersonate").(string)
			// check user role : impersonation not allowed for a user
			// with ORG Admin role
			isOrgAdmin, err := user.IsOrgAdmin(asUser)
			if err != nil {
				return diag.FromErr(err)
			}
			if isOrgAdmin {
				return diag.FromErr(fmt.Errorf("%s", "--as-user cannot have ORGADMIN role"))
			}
			config.ApiKey, config.ApiSecret, err = user.GetUserAPIKey(asUser)
			if err != nil {
				return diag.FromErr(err)
			}
		}

		auth := config.GetConfig().GetAppAuthProfile()
		client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
		if err != nil {
			return diags
		}

		err = client.AppsV3().WorkloadTemplate().Delete(ctx, options.DeleteOptions{
			Name:    ag.Metadata.Name,
			Project: ag.Metadata.Project,
		})
		if err != nil {
			return diags
		}
	}
	return diags
}

func resourceWorkloadTemplateUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("workloadtemplate update starts")
	return resourceWorkloadTemplateUpsert(ctx, d, m)
}

func resourceWorkloadTemplateUpsert(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("workloadtemplate upsert starts")
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

	workloadtemplate, err := expandWorkloadTemplate(d)
	if err != nil {
		log.Printf("workloadtemplate expandWorkloadTemplate error")
		return diag.FromErr(err)
	}

	if v, ok := d.Get("impersonate").(string); ok && len(v) > 0 {
		defer ResetImpersonateUser()
		asUser := d.Get("impersonate").(string)
		// check user role : impersonation not allowed for a user
		// with ORG Admin role
		isOrgAdmin, err := user.IsOrgAdmin(asUser)
		if err != nil {
			return diag.FromErr(err)
		}
		if isOrgAdmin {
			return diag.FromErr(fmt.Errorf("%s", "--as-user cannot have ORGADMIN role"))
		}
		config.ApiKey, config.ApiSecret, err = user.GetUserAPIKey(asUser)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.AppsV3().WorkloadTemplate().Apply(ctx, workloadtemplate, options.ApplyOptions{})
	if err != nil {
		// XXX Debug
		n1 := spew.Sprintf("%+v", workloadtemplate)
		log.Println("workloadtemplate apply workloadtemplate:", n1)
		log.Printf("workloadtemplate apply error")
		return diag.FromErr(err)
	}

	d.SetId(workloadtemplate.Metadata.Name)
	return diags

}

func resourceWorkloadTemplateRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	log.Println("resourceWorkloadTemplateRead ")
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

	tfWorkloadTemplateState, err := expandWorkloadTemplate(d)
	if err != nil {
		return diag.FromErr(err)
	}

	if v, ok := d.Get("impersonate").(string); ok && len(v) > 0 {
		defer ResetImpersonateUser()
		asUser := d.Get("impersonate").(string)
		// check user role : impersonation not allowed for a user
		// with ORG Admin role
		isOrgAdmin, err := user.IsOrgAdmin(asUser)
		if err != nil {
			return diag.FromErr(err)
		}
		if isOrgAdmin {
			return diag.FromErr(fmt.Errorf("%s", "--as-user cannot have ORGADMIN role"))
		}
		config.ApiKey, config.ApiSecret, err = user.GetUserAPIKey(asUser)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}

	wt, err := client.AppsV3().WorkloadTemplate().Get(ctx, options.GetOptions{
		//Name:    tfWorkloadTemplateState.Metadata.Name,
		Name:    meta.Name,
		Project: tfWorkloadTemplateState.Metadata.Project,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	err = flattenWorkloadTemplate(d, wt)
	if err != nil {
		return diag.FromErr(err)
	}
	return diags

}

func resourceWorkloadTemplateDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	wt, err := expandWorkloadTemplate(d)
	if err != nil {
		return diag.FromErr(err)
	}

	if v, ok := d.Get("impersonate").(string); ok && len(v) > 0 {
		defer ResetImpersonateUser()
		asUser := d.Get("impersonate").(string)
		// check user role : impersonation not allowed for a user
		// with ORG Admin role
		isOrgAdmin, err := user.IsOrgAdmin(asUser)
		if err != nil {
			return diag.FromErr(err)
		}
		if isOrgAdmin {
			return diag.FromErr(fmt.Errorf("%s", "--as-user cannot have ORGADMIN role"))
		}
		config.ApiKey, config.ApiSecret, err = user.GetUserAPIKey(asUser)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.AppsV3().WorkloadTemplate().Delete(ctx, options.DeleteOptions{
		Name:    wt.Metadata.Name,
		Project: wt.Metadata.Project,
	})

	if err != nil {
		//v3 spec gave error try v2
		return resourceWorkloadTemplateV2Delete(ctx, wt)
	}

	return diags
}

func resourceWorkloadTemplateV2Delete(ctx context.Context, ag *appspb.WorkloadTemplate) diag.Diagnostics {
	var diags diag.Diagnostics

	projectId, err := config.GetProjectIdByName(ag.Metadata.Project)
	if err != nil {
		return diag.FromErr(err)
	}
	//delete workloadtemplate
	err = workloadtemplate.DeleteWorkloadTemplate(ag.Metadata.Name, projectId)
	if err != nil {
		log.Println("error deleting workloadtemplate")
	} else {
		log.Println("Deleted workloadtemplate: ", ag.Metadata.Name)
	}
	return diags
}

func expandWorkloadTemplate(in *schema.ResourceData) (*appspb.WorkloadTemplate, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "expand workloadtemplate empty input")
	}
	obj := &appspb.WorkloadTemplate{}

	if v, ok := in.Get("metadata").([]interface{}); ok && len(v) > 0 {
		obj.Metadata = expandMetaData(v)
	}

	if v, ok := in.Get("spec").([]interface{}); ok && len(v) > 0 {
		objSpec, err := expandWorkloadTemplateSpec(v)
		if err != nil {
			return nil, err
		}
		log.Println("expandWorkloadTemplateSpec got spec")
		obj.Spec = objSpec
	}

	obj.ApiVersion = "apps.k8smgmt.io/v3"
	obj.Kind = "WorkloadTemplate"
	return obj, nil
}

func expandWorkloadTemplateSpec(p []interface{}) (*appspb.WorkloadTemplateSpec, error) {
	obj := &appspb.WorkloadTemplateSpec{}
	if len(p) == 0 || p[0] == nil {
		return obj, fmt.Errorf("%s", "expandWorkloadTemplateSpec empty input")
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["artifact"].([]interface{}); ok {
		objArtifact, err := ExpandArtifactSpec(v)
		if err != nil {
			return nil, err
		}
		obj.Artifact = objArtifact
	}

	if v, ok := in["sharing"].([]interface{}); ok && len(v) > 0 {
		obj.Sharing = expandSharingSpec(v)
	}

	return obj, nil
}

// Flatteners

func flattenWorkloadTemplate(d *schema.ResourceData, in *appspb.WorkloadTemplate) error {
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
	ret, err = flattenWorkloadTemplateSpec(in.Spec, v)
	if err != nil {
		return err
	}

	err = d.Set("spec", ret)
	if err != nil {
		return err
	}
	return nil
}

func flattenWorkloadTemplateSpec(in *appspb.WorkloadTemplateSpec, p []interface{}) ([]interface{}, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "flattenWorkloadTemplateSpec empty input")
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	v, ok := obj["artifact"].([]interface{})
	if !ok {
		v = []interface{}{}
	}

	var ret []interface{}
	var err error
	ret, err = FlattenArtifactSpec(in.Artifact, v)
	if err != nil {
		log.Println("FlattenArtifactSpec error ", err)
		return nil, err
	}

	obj["artifact"] = ret

	obj["sharing"] = flattenSharingSpec(in.Sharing)

	return []interface{}{obj}, nil
}
