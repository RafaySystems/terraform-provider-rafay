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
	"github.com/RafaySystems/rafay-common/proto/types/hub/systempb"
	"github.com/RafaySystems/rctl/pkg/config"
	"github.com/davecgh/go-spew/spew"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceChargebackShare() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceChargebackShareCreate,
		ReadContext:   resourceChargebackShareRead,
		UpdateContext: resourceChargebackShareUpdate,
		DeleteContext: resourceChargebackShareDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema:        resource.ChargebackShareSchema.Schema,
	}
}

func resourceChargebackShareCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("ChargebackShare create starts")
	diags := resourceChargebackShareUpsert(ctx, d, m)
	if diags.HasError() {
		tflog := os.Getenv("TF_LOG")
		if tflog == "TRACE" || tflog == "DEBUG" {
			ctx = context.WithValue(ctx, "debug", "true")
		}
		log.Printf("chargebackShare create got error, perform cleanup")
		mp, err := expandChargebackShare(d)
		if err != nil {
			log.Printf("chargebackShare expandChargebackShare error")
			return diags
		}
		auth := config.GetConfig().GetAppAuthProfile()
		client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, getUserAgent(), options.WithInsecureSkipVerify(auth.SkipServerCertValid))
		if err != nil {
			return diags
		}

		err = client.SystemV3().ChargebackShare().Delete(ctx, options.DeleteOptions{
			Name:    mp.Metadata.Name,
			Project: mp.Metadata.Project,
		})
		if err != nil {
			return diags
		}
	}
	return diags
}

func resourceChargebackShareUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("chargebackShare update starts")
	return resourceChargebackShareUpsert(ctx, d, m)
}

func resourceChargebackShareUpsert(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("chargebackShare upsert starts")
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

	chargebackShare, err := expandChargebackShare(d)
	if err != nil {
		log.Printf("chargebackShare expandChargebackShare error")
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, getUserAgent(), options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.SystemV3().ChargebackShare().Apply(ctx, chargebackShare, options.ApplyOptions{})
	if err != nil {
		// XXX Debug
		n1 := spew.Sprintf("%+v", chargebackShare)
		log.Println("chargebackShare apply chargebackShare:", n1)
		log.Printf("chargebackShare apply error")
		return diag.FromErr(err)
	}

	d.SetId(chargebackShare.Metadata.Name)
	return diags

}

func resourceChargebackShareRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	log.Println("resourceChargebackShareRead ")
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

	tfChargebackShareState, err := expandChargebackShare(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, getUserAgent(), options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return diag.FromErr(err)
	}

	mp, err := client.SystemV3().ChargebackShare().Get(ctx, options.GetOptions{
		//Name:    tfChargebackShareState.Metadata.Name,
		Name:    meta.Name,
		Project: tfChargebackShareState.Metadata.Project,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	err = flattenChargebackShare(d, mp)
	if err != nil {
		return diag.FromErr(err)
	}
	return diags

}

func resourceChargebackShareDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	err := fmt.Errorf("ChargebackShare Delete is not supported")
	return diag.FromErr(err)
}

func expandChargebackShare(in *schema.ResourceData) (*systempb.ChargebackShare, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "expand chargebackShare empty input")
	}
	obj := &systempb.ChargebackShare{}

	if v, ok := in.Get("metadata").([]interface{}); ok && len(v) > 0 {
		obj.Metadata = expandMetaData(v)
	}

	if v, ok := in.Get("spec").([]interface{}); ok && len(v) > 0 {
		objSpec, err := expandChargebackShareSpec(v)
		if err != nil {
			return nil, err
		}
		log.Println("expandChargebackShareSpec got spec")
		obj.Spec = objSpec
	}

	obj.ApiVersion = "system.k8smgmt.io/v3"
	obj.Kind = "ChargebackShare"
	return obj, nil
}

func expandChargebackShareSpec(p []interface{}) (*systempb.ChargebackShareSpec, error) {
	obj := &systempb.ChargebackShareSpec{}
	if len(p) == 0 || p[0] == nil {
		return obj, fmt.Errorf("%s", "expandChargebackShareSpec empty input")
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["share_enabled"].(bool); ok {
		obj.ShareEnabled = v
	}

	if v, ok := in["share_type"].(string); ok && len(v) > 0 {
		obj.ShareType = v
	}

	return obj, nil
}

// Flatteners

func flattenChargebackShare(d *schema.ResourceData, in *systempb.ChargebackShare) error {
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
	ret, err = flattenChargebackShareSpec(in.Spec, v)
	if err != nil {
		return err
	}

	err = d.Set("spec", ret)
	if err != nil {
		return err
	}
	return nil
}

func flattenChargebackShareSpec(in *systempb.ChargebackShareSpec, p []interface{}) ([]interface{}, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "flattenChargebackShareSpec empty input")
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	obj["share_enabled"] = in.ShareEnabled

	if len(in.ShareType) > 0 {
		obj["share_type"] = in.ShareType
	}

	return []interface{}{obj}, nil
}
