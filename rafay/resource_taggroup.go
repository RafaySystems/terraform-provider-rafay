package rafay

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/RafaySystems/rafay-common/pkg/hub/client/options"
	typed "github.com/RafaySystems/rafay-common/pkg/hub/client/typed"
	"github.com/RafaySystems/rafay-common/pkg/hub/terraform/resource"
	"github.com/RafaySystems/rafay-common/proto/types/hub/tagspb"
	"github.com/RafaySystems/rctl/pkg/config"
	"github.com/RafaySystems/rctl/pkg/versioninfo"
	"github.com/davecgh/go-spew/spew"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTagGroup() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceTagGroupCreate,
		ReadContext:   resourceTagGroupRead,
		UpdateContext: resourceTagGroupUpdate,
		DeleteContext: resourceTagGroupDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema:        resource.TagGroupSchema.Schema,
	}
}

func resourceTagGroupCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("TagGroup create starts")
	diags := resourceTagGroupUpsert(ctx, d, m)
	if diags.HasError() {
		tflog := os.Getenv("TF_LOG")
		if tflog == "TRACE" || tflog == "DEBUG" {
			ctx = context.WithValue(ctx, "debug", "true")
		}
		log.Printf("TagGroup create got error, perform cleanup")
		tagGroup, err := expandTagGroup(d)
		if err != nil {
			log.Printf("TagGroup expandTagGroup error")
			return diags
		}
		auth := config.GetConfig().GetAppAuthProfile()
		client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
		if err != nil {
			return diags
		}

		err = client.TagsV3().TagGroup().Delete(ctx, options.DeleteOptions{
			Name:    tagGroup.Metadata.Name,
			Project: tagGroup.Metadata.Project,
		})
		if err != nil {
			return diags
		}
	}
	return diags
}

func resourceTagGroupUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("TagGroup update starts")
	return resourceTagGroupUpsert(ctx, d, m)
}

func resourceTagGroupUpsert(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("TagGroup upsert starts")
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	if d.State() != nil && d.State().ID != "" {
		n := GetMetaName(d)
		if n != "" && n != d.State().ID {
			log.Printf("TagGroup metadata name change not supported")
			d.State().Tainted = true
			return diag.FromErr(fmt.Errorf("%s", "TagGroup metadata name change not supported"))
		}
	}

	tagGroup, err := expandTagGroup(d)
	if err != nil {
		log.Printf("TagGroup expandTagGroup error")
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.TagsV3().TagGroup().Apply(ctx, tagGroup, options.ApplyOptions{})
	if err != nil {
		// XXX Debug
		n1 := spew.Sprintf("%+v", tagGroup)
		log.Println("TagGroup apply tagGroup:", n1)
		log.Printf("TagGroup apply error")
		return diag.FromErr(err)
	}

	d.SetId(tagGroup.Metadata.Name)
	return diags

}

func resourceTagGroupRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	log.Println("resourceTagGroupRead ")
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	meta := GetMetaData(d)
	if meta == nil {
		return diag.FromErr(fmt.Errorf("%s", "TagGroup failed to read resource "))
	}
	if d.State() != nil && d.State().ID != "" {
		meta.Name = d.State().ID
	}

	tagGroup, err := expandTagGroup(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}

	tagGroup, err = client.TagsV3().TagGroup().Get(ctx, options.GetOptions{
		Name:    meta.Name,
		Project: tagGroup.Metadata.Project,
	})
	if err != nil {
		if strings.Contains(err.Error(), "code 404") {
			log.Println("TagGroup Resource Read ", "error", err)
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	err = flattenTagGroup(d, tagGroup)
	if err != nil {
		return diag.FromErr(err)
	}
	return diags

}

func resourceTagGroupDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	tagGroup, err := expandTagGroup(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.TagsV3().TagGroup().Delete(ctx, options.DeleteOptions{
		Name:    tagGroup.Metadata.Name,
		Project: tagGroup.Metadata.Project,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func expandTagGroup(in *schema.ResourceData) (*tagspb.TagGroup, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "expand tagGroup empty input")
	}
	obj := &tagspb.TagGroup{}

	if v, ok := in.Get("metadata").([]interface{}); ok && len(v) > 0 {
		obj.Metadata = expandMetaData(v)
	}

	if v, ok := in.Get("spec").([]interface{}); ok && len(v) > 0 {
		objSpec, err := expandTagGroupSpec(v)
		if err != nil {
			return nil, err
		}
		log.Println("expandTagGroupSpec got spec")
		obj.Spec = objSpec
	}

	obj.ApiVersion = "tags.k8smgmt.io/v3"
	obj.Kind = "TagGroup"
	return obj, nil
}

func expandTagGroupSpec(p []interface{}) (*tagspb.TagGroupSpec, error) {
	obj := &tagspb.TagGroupSpec{}
	if len(p) == 0 || p[0] == nil {
		return obj, fmt.Errorf("%s", "expandTagGroupSpec empty input")
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["tags"].([]interface{}); ok && len(v) > 0 {
		obj.Tags = expandTagGroupSpecTags(v)
	}

	return obj, nil
}

func expandTagGroupSpecTags(p []interface{}) []*tagspb.TagConfig {
	if len(p) == 0 || p[0] == nil {
		return []*tagspb.TagConfig{}
	}

	out := make([]*tagspb.TagConfig, len(p))

	for i := range p {
		obj := tagspb.TagConfig{}
		in := p[i].(map[string]interface{})

		if v, ok := in["key"].(string); ok {
			obj.Key = v
		}

		if v, ok := in["value"].(string); ok {
			obj.Value = v
		}

		out[i] = &obj
	}

	return out

}

// Flatteners

func flattenTagGroup(d *schema.ResourceData, in *tagspb.TagGroup) error {
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
	ret, err = flattenTagGroupSpec(in.Spec, v)
	if err != nil {
		return err
	}

	err = d.Set("spec", ret)
	if err != nil {
		return err
	}
	return nil
}

func flattenTagGroupSpec(in *tagspb.TagGroupSpec, p []interface{}) ([]interface{}, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "flattenTagGroupSpec empty input")
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if in.Tags != nil {
		v, ok := obj["tags"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["tags"] = flattenTagGroupSpecTags(in.Tags, v)
	}

	return []interface{}{obj}, nil
}

func flattenTagGroupSpecTags(in []*tagspb.TagConfig, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}

	out := make([]interface{}, len(in))
	for i, in := range in {

		obj := map[string]interface{}{}
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}

		if len(in.Key) > 0 {
			obj["key"] = in.Key
		}

		if len(in.Value) > 0 {
			obj["value"] = in.Value
		}

		out[i] = &obj
	}

	return out
}
