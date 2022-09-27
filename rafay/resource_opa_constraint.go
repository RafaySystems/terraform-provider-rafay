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
	"github.com/RafaySystems/rafay-common/proto/types/hub/opapb"
	"github.com/RafaySystems/rctl/pkg/config"
	"github.com/RafaySystems/rctl/pkg/versioninfo"
	"github.com/davecgh/go-spew/spew"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceOPAConstraint() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOPAConstraintCreate,
		ReadContext:   resourceOPAConstraintRead,
		UpdateContext: resourceOPAConstraintUpdate,
		DeleteContext: resourceOPAConstraintDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema:        resource.OPAConstraintSchema.Schema,
	}
}

func resourceOPAConstraintCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("resourceOPAConstraintCreate reate starts")
	diags := resourceOPAConstraintUpsert(ctx, d, m)
	if diags.HasError() {
		tflog := os.Getenv("TF_LOG")
		if tflog == "TRACE" || tflog == "DEBUG" {
			ctx = context.WithValue(ctx, "debug", "true")
		}
		log.Printf("Opa constraint create got error, perform cleanup")
		ss, err := expandOPAConstraint(d)
		if err != nil {
			log.Printf("Opa constraint expandOPAConstraint error")
			return diags
		}
		auth := config.GetConfig().GetAppAuthProfile()
		client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
		if err != nil {
			return diags
		}

		err = client.OpaV3().OPAConstraint().Delete(ctx, options.DeleteOptions{
			Name:    ss.Metadata.Name,
			Project: ss.Metadata.Project,
		})
		if err != nil {
			return diags
		}
	}
	return diags
}

func resourceOPAConstraintUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("Opa constraint update starts")
	return resourceOPAConstraintUpsert(ctx, d, m)
}

func resourceOPAConstraintUpsert(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("Opa constraint upsert starts")
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	opaConstraint, err := expandOPAConstraint(d)
	if err != nil {
		log.Printf("Opa constraint expandOPAConstraint error")
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.OpaV3().OPAConstraint().Apply(ctx, opaConstraint, options.ApplyOptions{})
	if err != nil {
		// XXX Debug
		n1 := spew.Sprintf("%+v", opaConstraint)
		log.Println("Opa constraint apply Opa constraint:", n1)
		log.Printf("Opa constraint apply error")
		return diag.FromErr(err)
	}

	d.SetId(opaConstraint.Metadata.Name)
	return diags

}

func resourceOPAConstraintRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	log.Println("resourceOPAConstraintRead ")
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}
	tfOPAConstraintState, err := expandOPAConstraint(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}

	ag, err := client.OpaV3().OPAConstraint().Get(ctx, options.GetOptions{
		Name:    tfOPAConstraintState.Metadata.Name,
		Project: tfOPAConstraintState.Metadata.Project,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	err = flattenOPAConstraint(d, ag)
	if err != nil {
		return diag.FromErr(err)
	}
	return diags

}

func resourceOPAConstraintDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	ag, err := expandOPAConstraint(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.OpaV3().OPAConstraint().Delete(ctx, options.DeleteOptions{
		Name:    ag.Metadata.Name,
		Project: ag.Metadata.Project,
	})

	return diags
}

func expandOPAConstraint(in *schema.ResourceData) (*opapb.OPAConstraint, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "expand Opa constraint empty input")
	}
	obj := &opapb.OPAConstraint{}

	if v, ok := in.Get("metadata").([]interface{}); ok && len(v) > 0 {
		obj.Metadata = expandMetaData(v)
	}

	if v, ok := in.Get("spec").([]interface{}); ok && len(v) > 0 {
		objSpec, err := expandOPAConstraintSpec(v)
		if err != nil {
			return nil, err
		}
		log.Println("expandOPAConstraintSpec got spec")
		obj.Spec = objSpec
	}

	obj.ApiVersion = "opa.k8smgmt.io/v3"
	obj.Kind = "OPAConstraint"
	return obj, nil
}

func expandOPAConstraintSpec(p []interface{}) (*opapb.OPAConstraintSpec, error) {
	obj := &opapb.OPAConstraintSpec{}
	if len(p) == 0 || p[0] == nil {
		return obj, fmt.Errorf("%s", "expandOPAConstraintSpec empty input")
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["template_name"].(string); ok && len(v) > 0 {
		obj.TemplateName = v
	}

	if v, ok := in["version"].(string); ok && len(v) > 0 {
		obj.Version = v
	}

	// if v, ok := in["published"].(bool); ok {
	// 	obj.Published = v
	// }

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

func flattenOPAConstraint(d *schema.ResourceData, in *opapb.OPAConstraint) error {
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
	ret, err = flattenOPAConstraintSpec(in.Spec, v)
	if err != nil {
		return err
	}

	err = d.Set("spec", ret)
	if err != nil {
		return err
	}
	return nil
}

func flattenOPAConstraintSpec(in *opapb.OPAConstraintSpec, p []interface{}) ([]interface{}, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "flattenOPAConstraint empty input")
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.TemplateName) > 0 {
		obj["template_name"] = in.TemplateName
	}

	if len(in.Version) > 0 {
		obj["version"] = in.Version
	}

	obj["published"] = in.Published

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
