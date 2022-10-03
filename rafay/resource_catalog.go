package rafay

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/RafaySystems/rctl/pkg/config"

	"github.com/RafaySystems/rafay-common/pkg/hub/client/options"
	typed "github.com/RafaySystems/rafay-common/pkg/hub/client/typed"
	"github.com/RafaySystems/rafay-common/pkg/hub/terraform/resource"
	"github.com/RafaySystems/rafay-common/proto/types/hub/appspb"
	"github.com/RafaySystems/rctl/pkg/versioninfo"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceCatalog() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCatalogCreate,
		ReadContext:   resourceCatalogRead,
		UpdateContext: resourceCatalogUpdate,
		DeleteContext: resourceCatalogDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema:        resource.CatalogSchema.Schema,
	}
}

func resourceCatalogCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Println("catalog create")
	diags := resourceCatalogUpsert(ctx, d, m)
	if diags.HasError() {
		tflog := os.Getenv("TF_LOG")
		if tflog == "TRACE" || tflog == "DEBUG" {
			ctx = context.WithValue(ctx, "debug", "true")
		}
		cg, err := expandCatalog(d)
		if err != nil {
			return diags
		}
		auth := config.GetConfig().GetAppAuthProfile()
		client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
		if err != nil {
			return diags
		}

		err = client.AppsV3().Catalog().Delete(ctx, options.DeleteOptions{
			Name:    cg.Metadata.Name,
			Project: cg.Metadata.Project,
		})
		if err != nil {
			return diags
		}
	}
	return diags
}
func resourceCatalogUpsert(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("catalog upsert starts")
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	cg, err := expandCatalog(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.AppsV3().Catalog().Apply(ctx, cg, options.ApplyOptions{})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(cg.Metadata.Name)
	return diags
}

func resourceCatalogRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	log.Println("resourceCatalogRead ")

	tfWorkloadState, err := expandCatalog(d)
	if err != nil {
		return diag.FromErr(err)
	}

	// XXX Debug
	// w1 := spew.Sprintf("%+v", tfWorkloadState)
	// log.Println("resourceWorkloadRead tfWorkloadState", w1)

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		log.Println("read client err")
		return diag.FromErr(err)
	}

	cg, err := client.AppsV3().Catalog().Get(ctx, options.GetOptions{
		Name:    tfWorkloadState.Metadata.Name,
		Project: tfWorkloadState.Metadata.Project,
	})
	if err != nil {
		log.Println("read get err")
		return diag.FromErr(err)
	}

	// XXX Debug
	// w1 = spew.Sprintf("%+v", wl)
	// log.Println("resourceWorkloadRead wl", w1)

	err = flattenCatalog(d, cg)
	if err != nil {
		log.Println("read flatten err")
		return diag.FromErr(err)
	}
	return diags

}

func resourceCatalogUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceCatalogUpsert(ctx, d, m)
}

func resourceCatalogDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Println("resourceCatalogDelete")
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

	cg, err := expandCatalog(d)
	if err != nil {
		log.Println("delete expand err")
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}
	err = client.AppsV3().Catalog().Delete(ctx, options.DeleteOptions{
		Name:    cg.Metadata.Name,
		Project: cg.Metadata.Project,
	})

	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func expandCatalog(in *schema.ResourceData) (*appspb.Catalog, error) {
	log.Println("expand catalog")
	if in == nil {
		return nil, fmt.Errorf("%s", "expandWorkload empty input")
	}
	obj := &appspb.Catalog{}

	if v, ok := in.Get("metadata").([]interface{}); ok {
		obj.Metadata = expandMetaData(v)
	}

	if v, ok := in.Get("spec").([]interface{}); ok {
		objSpec, err := expandCatalogSpec(v)
		if err != nil {
			return nil, err
		}
		obj.Spec = objSpec
	}

	obj.ApiVersion = "apps.k8smgmt.io/v3"
	obj.Kind = "Catalog"
	return obj, nil
}

func expandCatalogSpec(p []interface{}) (*appspb.CatalogSpec, error) {
	log.Println("expand catalog spec")
	obj := &appspb.CatalogSpec{}
	if len(p) == 0 || p[0] == nil {
		return obj, fmt.Errorf("%s", "expandCatalogSpec empty input")
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["auto_sync"].(bool); ok {
		obj.AutoSync = v
	}
	if v, ok := in["icon_url"].(string); ok && len(v) > 0 {
		obj.IconURL = v
	}
	if v, ok := in["repository"].(string); ok && len(v) > 0 {
		obj.Repository = v
	}

	if v, ok := in["sharing"].([]interface{}); ok {
		obj.Sharing = expandSharingSpec(v)
	}

	if v, ok := in["type"].(string); ok && len(v) > 0 {
		obj.Type = v
	}

	return obj, nil
}

// Flatteners

func flattenCatalog(d *schema.ResourceData, in *appspb.Catalog) error {
	if in == nil {
		return nil
	}

	err := d.Set("metadata", flattenMetaData(in.Metadata))
	if err != nil {
		log.Println("flatten metadata err")
		return err
	}

	v, ok := d.Get("spec").([]interface{})
	if !ok {
		v = []interface{}{}
	}

	var ret []interface{}
	ret, err = flattenCatalogSpec(in.Spec, v)
	if err != nil {
		log.Println("flatten catalog spec err")
		return err
	}

	err = d.Set("spec", ret)
	if err != nil {
		log.Println("set spec err")
		return err
	}
	return nil
}

func flattenCatalogSpec(in *appspb.CatalogSpec, p []interface{}) ([]interface{}, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "flattenCatalogSpec empty input")
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if in.AutoSync {
		obj["auto_sync"] = in.AutoSync
	}
	if len(in.IconURL) > 0 {
		obj["icon_url"] = in.IconURL
	}
	if len(in.Repository) > 0 {
		obj["repository"] = in.Repository
	}

	if in.Sharing != nil {
		obj["sharing"] = flattenSharingSpec(in.Sharing)
	}

	if len(in.Type) > 0 {
		obj["type"] = in.Type
	}

	return []interface{}{obj}, nil
}
