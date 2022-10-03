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
	"github.com/RafaySystems/rafay-common/proto/types/hub/commonpb"
	"github.com/RafaySystems/rafay-common/proto/types/hub/opapb"
	"github.com/RafaySystems/rctl/pkg/config"
	"github.com/RafaySystems/rctl/pkg/versioninfo"
	"github.com/davecgh/go-spew/spew"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceOPAConstraintTemplate() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOPAConstraintTemplateCreate,
		ReadContext:   resourceOPAConstraintTemplateRead,
		UpdateContext: resourceOPAConstraintTemplateUpdate,
		DeleteContext: resourceOPAConstraintTemplateDelete,
		Importer: &schema.ResourceImporter{
			State: resourceNamespaceImport,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema:        resource.OPAConstraintTemplateSchema.Schema,
	}
}

func resourceOPAConstraintTemplateImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	idParts := strings.SplitN(d.Id(), "/", 2)
	log.Println("resourceNamespaceImport idParts:", idParts)
	d_debug := spew.Sprintf("%+v", d)
	log.Println("resourceNamespaceImport d.Id:", d.Id())
	log.Println("resourceNamespaceImport d_debug", d_debug)

	opaCT, err := expandOPAConstraintTemplate(d)
	if err != nil {
		log.Printf("namespace expandNamespace error")
		//return nil, err
	}
	log.Println("import1")
	var metaD commonpb.Metadata
	metaD.Name = idParts[0]
	metaD.Project = idParts[1]
	opaCT.Metadata = &metaD
	log.Println("import pre flatten")
	err = d.Set("metadata", flattenMetaData(opaCT.Metadata))
	if err != nil {
		log.Println("import set err")
		return nil, err
	}
	log.Println("import post flatten")
	d.SetId(opaCT.Metadata.Name)
	log.Println("import post set id")
	return []*schema.ResourceData{d}, nil
}

func resourceOPAConstraintTemplateCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("resourceOPAConstraintTemplateCreate reate starts")
	diags := resourceOPAConstraintTemplateUpsert(ctx, d, m)
	if diags.HasError() {
		tflog := os.Getenv("TF_LOG")
		if tflog == "TRACE" || tflog == "DEBUG" {
			ctx = context.WithValue(ctx, "debug", "true")
		}
		log.Printf("Opa constraint create got error, perform cleanup")
		ss, err := expandOPAConstraintTemplate(d)
		if err != nil {
			log.Printf("Opa constraint expandOPAConstraintTemplate error")
			return diags
		}
		auth := config.GetConfig().GetAppAuthProfile()
		client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
		if err != nil {
			return diags
		}

		err = client.OpaV3().OPAConstraintTemplate().Delete(ctx, options.DeleteOptions{
			Name:    ss.Metadata.Name,
			Project: ss.Metadata.Project,
		})
		if err != nil {
			return diags
		}
	}
	return diags
}

func resourceOPAConstraintTemplateUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("Opa constraint Template update starts")
	return resourceOPAConstraintTemplateUpsert(ctx, d, m)
}

func resourceOPAConstraintTemplateUpsert(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("Opa constraint Template upsert starts")
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

	opaConstraintTemplate, err := expandOPAConstraintTemplate(d)
	if err != nil {
		log.Printf("Opa constraint expandOPAConstraintTemplate error")
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.OpaV3().OPAConstraintTemplate().Apply(ctx, opaConstraintTemplate, options.ApplyOptions{})
	if err != nil {
		// XXX Debug
		n1 := spew.Sprintf("%+v", opaConstraintTemplate)
		log.Println("Opa constraint apply Opa constraint:", n1)
		log.Printf("Opa constraint apply error")
		return diag.FromErr(err)
	}

	d.SetId(opaConstraintTemplate.Metadata.Name)
	return diags

}

func resourceOPAConstraintTemplateRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	log.Println("resourceOPAConstraintTemplateRead ")
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}
	tfOPAConstraintTemplateState, err := expandOPAConstraintTemplate(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}

	ag, err := client.OpaV3().OPAConstraintTemplate().Get(ctx, options.GetOptions{
		Name:    tfOPAConstraintTemplateState.Metadata.Name,
		Project: tfOPAConstraintTemplateState.Metadata.Project,
	})
	if err != nil {
		return diag.FromErr(err)
	}
	log.Println("opct:", ag)

	err = flattenOPAConstraintTemplate(d, ag)
	if err != nil {
		return diag.FromErr(err)
	}
	return diags

}

func resourceOPAConstraintTemplateDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	ag, err := expandOPAConstraintTemplate(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.OpaV3().OPAConstraintTemplate().Delete(ctx, options.DeleteOptions{
		Name:    ag.Metadata.Name,
		Project: ag.Metadata.Project,
	})

	return diags
}

func expandOPAConstraintTemplate(in *schema.ResourceData) (*opapb.OPAConstraintTemplate, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "expand Opa constraint empty input")
	}
	obj := &opapb.OPAConstraintTemplate{}

	if v, ok := in.Get("metadata").([]interface{}); ok && len(v) > 0 {
		obj.Metadata = expandMetaData(v)
	}

	if v, ok := in.Get("spec").([]interface{}); ok && len(v) > 0 {
		objSpec, err := expandOPAConstraintTemplateSpec(v)
		if err != nil {
			return nil, err
		}
		log.Println("expandOPAConstraintTemplateSpec got spec")
		obj.Spec = objSpec
	}

	obj.ApiVersion = "opa.k8smgmt.io/v3"
	obj.Kind = "OPAConstraintTemplate"
	return obj, nil
}

func expandOPAConstraintTemplateSpec(p []interface{}) (*opapb.OPAConstraintTemplateSpec, error) {
	obj := &opapb.OPAConstraintTemplateSpec{}
	if len(p) == 0 || p[0] == nil {
		return obj, fmt.Errorf("%s", "expandOPAConstraintTemplateSpec empty input")
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["artifact"].([]interface{}); ok {
		objArtifact, err := ExpandArtifactSpec(v)
		if err != nil {
			return nil, err
		}
		obj.Artifact = objArtifact
	}

	return obj, nil
}

// Flatten

func flattenOPAConstraintTemplate(d *schema.ResourceData, in *opapb.OPAConstraintTemplate) error {
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
	ret, err = flattenOPAConstraintTemplateSpec(in.Spec, v)
	if err != nil {
		return err
	}

	err = d.Set("spec", ret)
	if err != nil {
		return err
	}
	return nil
}

func flattenOPAConstraintTemplateSpec(in *opapb.OPAConstraintTemplateSpec, p []interface{}) ([]interface{}, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "flattenOPAConstraint empty input")
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

	return []interface{}{obj}, nil
}
