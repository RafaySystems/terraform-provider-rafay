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

func resourceProjectTagsAssociation() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceProjectTagsAssociationCreate,
		ReadContext:   resourceProjectTagsAssociationRead,
		UpdateContext: resourceProjectTagsAssociationUpdate,
		DeleteContext: resourceProjectTagsAssociationDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema:        resource.ProjectTagsAssociationSchema.Schema,
	}
}

func resourceProjectTagsAssociationCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("ProjectTagsAssociation create starts")
	diags := resourceProjectTagsAssociationUpsert(ctx, d, m)
	if diags.HasError() {
		tflog := os.Getenv("TF_LOG")
		if tflog == "TRACE" || tflog == "DEBUG" {
			ctx = context.WithValue(ctx, "debug", "true")
		}
		log.Printf("ProjectTagsAssociation create got error, perform cleanup")
		tagGroupAssociation, err := expandProjectTagsAssociation(d)
		if err != nil {
			log.Printf("ProjectTagsAssociation expandProjectTagsAssociation error")
			return diags
		}
		auth := config.GetConfig().GetAppAuthProfile()
		client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent(), options.WithInsecureSkipVerify(auth.SkipServerCertValid))
		if err != nil {
			return diags
		}

		err = client.TagsV3().ProjectTagsAssociation().Delete(ctx, options.DeleteOptions{
			Name:    tagGroupAssociation.Metadata.Name,
			Project: tagGroupAssociation.Metadata.Project,
		})
		if err != nil {
			return diags
		}
	}
	return diags
}

func resourceProjectTagsAssociationUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("ProjectTagsAssociation update starts")
	return resourceProjectTagsAssociationUpsert(ctx, d, m)
}

func resourceProjectTagsAssociationUpsert(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("ProjectTagsAssociation upsert starts")
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	if d.State() != nil && d.State().ID != "" {
		n := GetMetaName(d)
		if n != "" && n != d.State().ID {
			log.Printf("ProjectTagsAssociation metadata name change not supported")
			d.State().Tainted = true
			return diag.FromErr(fmt.Errorf("%s", "ProjectTagsAssociation metadata name change not supported"))
		}
	}

	tagGroupAssociation, err := expandProjectTagsAssociation(d)
	if err != nil {
		log.Printf("ProjectTagsAssociation expandProjectTagsAssociation error")
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent(), options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.TagsV3().ProjectTagsAssociation().Apply(ctx, tagGroupAssociation, options.ApplyOptions{})
	if err != nil {
		// XXX Debug
		n1 := spew.Sprintf("%+v", tagGroupAssociation)
		log.Println("ProjectTagsAssociation apply tagGroupAssociation:", n1)
		log.Printf("ProjectTagsAssociation apply error")
		return diag.FromErr(err)
	}

	d.SetId(tagGroupAssociation.Metadata.Name)
	return diags

}

func resourceProjectTagsAssociationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	log.Println("resourceProjectTagsAssociationRead ")
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	meta := GetMetaData(d)
	if meta == nil {
		return diag.FromErr(fmt.Errorf("%s", "ProjectTagsAssociation failed to read resource "))
	}
	if d.State() != nil && d.State().ID != "" {
		meta.Name = d.State().ID
	}

	tagGroupAssociation, err := expandProjectTagsAssociation(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent(), options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return diag.FromErr(err)
	}

	tagGroupAssociation, err = client.TagsV3().ProjectTagsAssociation().Get(ctx, options.GetOptions{
		Name:    meta.Name,
		Project: tagGroupAssociation.Metadata.Project,
	})
	if err != nil {
		if strings.Contains(err.Error(), "code 404") {
			log.Println("ProjectTagsAssociation Resource Read ", "error", err)
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	err = flattenProjectTagsAssociation(d, tagGroupAssociation)
	if err != nil {
		return diag.FromErr(err)
	}
	return diags

}

func resourceProjectTagsAssociationDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	tagGroupAssociation, err := expandProjectTagsAssociation(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent(), options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.TagsV3().ProjectTagsAssociation().Delete(ctx, options.DeleteOptions{
		Name:    tagGroupAssociation.Metadata.Name,
		Project: tagGroupAssociation.Metadata.Project,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func expandProjectTagsAssociation(in *schema.ResourceData) (*tagspb.ProjectTagsAssociation, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "expand tagGroupAssociation empty input")
	}
	obj := &tagspb.ProjectTagsAssociation{}

	if v, ok := in.Get("metadata").([]interface{}); ok && len(v) > 0 {
		obj.Metadata = expandMetaData(v)
	}

	if v, ok := in.Get("spec").([]interface{}); ok && len(v) > 0 {
		objSpec, err := expandProjectTagsAssociationSpec(v)
		if err != nil {
			return nil, err
		}
		log.Println("expandProjectTagsAssociationSpec got spec")
		obj.Spec = objSpec
	}

	obj.ApiVersion = "tags.k8smgmt.io/v3"
	obj.Kind = "ProjectTagsAssociation"
	return obj, nil
}

func expandProjectTagsAssociationSpec(p []interface{}) (*tagspb.ProjectTagAssociationSpec, error) {
	obj := &tagspb.ProjectTagAssociationSpec{}
	if len(p) == 0 || p[0] == nil {
		return obj, fmt.Errorf("%s", "expandProjectTagsAssociationSpec empty input")
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["associations"].([]interface{}); ok && len(v) > 0 {
		obj.Associations = expandProjectTagsAssociationSpecAssociations(v)
	}

	return obj, nil
}

func expandProjectTagsAssociationSpecAssociations(p []interface{}) []*tagspb.TagAssociation {
	if len(p) == 0 || p[0] == nil {
		return []*tagspb.TagAssociation{}
	}

	out := make([]*tagspb.TagAssociation, len(p))

	for i := range p {
		obj := tagspb.TagAssociation{}
		in := p[i].(map[string]interface{})

		if v, ok := in["tag_key"].(string); ok {
			obj.TagKey = v
		}

		if v, ok := in["tag_type"].(string); ok {
			obj.TagType = v
		}

		out[i] = &obj
	}

	return out

}

// Flatteners

func flattenProjectTagsAssociation(d *schema.ResourceData, in *tagspb.ProjectTagsAssociation) error {
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
	ret, err = flattenProjectTagsAssociationSpec(in.Spec, v)
	if err != nil {
		return err
	}

	err = d.Set("spec", ret)
	if err != nil {
		return err
	}
	return nil
}

func flattenProjectTagsAssociationSpec(in *tagspb.ProjectTagAssociationSpec, p []interface{}) ([]interface{}, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "flattenProjectTagsAssociationSpec empty input")
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if in.Associations != nil {
		v, ok := obj["associations"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["associations"] = flattenProjectTagsAssociationSpecAssociations(in.Associations, v)
	}

	return []interface{}{obj}, nil
}

func flattenProjectTagsAssociationSpecAssociations(in []*tagspb.TagAssociation, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}

	out := make([]interface{}, len(in))
	for i, in := range in {

		obj := map[string]interface{}{}
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}

		if len(in.TagKey) > 0 {
			obj["tag_key"] = in.TagKey
		}

		if len(in.TagType) > 0 {
			obj["tag_type"] = in.TagType
		}

		out[i] = &obj
	}

	return out
}
