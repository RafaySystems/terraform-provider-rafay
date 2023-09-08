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
	"github.com/RafaySystems/rafay-common/proto/types/hub/systempb"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceOrganizationAlertConfig() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOrganizationAlertConfigCreate,
		ReadContext:   resourceOrganizationAlertConfigRead,
		UpdateContext: resourceOrganizationAlertConfigUpdate,
		DeleteContext: resourceOrganizationAlertConfigDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema:        resource.AlertConfigurationSchema.Schema,
	}
}

func resourceOrganizationAlertConfigCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Println("organization alert config create")
	diags := resourceOrganizationAlertConfigUpsert(ctx, d, m)
	if diags.HasError() {
		tflog := os.Getenv("TF_LOG")
		if tflog == "TRACE" || tflog == "DEBUG" {
			ctx = context.WithValue(ctx, "debug", "true")
		}
		cg, err := expandOrganizationAlertConfig(d)
		if err != nil {
			return diags
		}
		auth := config.GetConfig().GetAppAuthProfile()
		client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, getUserAgent(), options.WithInsecureSkipVerify(auth.SkipServerCertValid))
		if err != nil {
			return diags
		}

		err = client.SystemV3().OrganizationAlertConfiguration().Delete(ctx, options.DeleteOptions{
			Name: cg.Metadata.Name,
		})
		if err != nil {
			return diags
		}
	}
	return diags
}
func resourceOrganizationAlertConfigUpsert(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("alertconfig upsert starts")
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	ac, err := expandOrganizationAlertConfig(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, getUserAgent(), options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.SystemV3().OrganizationAlertConfiguration().Apply(ctx, ac, options.ApplyOptions{})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(ac.Metadata.Name)
	return diags
}

func resourceOrganizationAlertConfigRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	log.Println("resourceOrganizationAlertConfigRead ")

	meta := GetMetaData(d)
	if meta == nil {
		return diag.FromErr(fmt.Errorf("%s", "failed to read resource "))
	}
	if d.State() != nil && d.State().ID != "" {
		meta.Name = d.State().ID
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, getUserAgent(), options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		log.Println("read client err")
		return diag.FromErr(err)
	}

	ac, err := client.SystemV3().OrganizationAlertConfiguration().Get(ctx, options.GetOptions{
		Name: meta.Name,
	})
	if err != nil {
		log.Println("read get err")
		return diag.FromErr(err)
	}

	err = flattenOrganizationAlertConfig(d, ac)
	if err != nil {
		log.Println("read flatten err")
		return diag.FromErr(err)
	}
	return diags
}

func resourceOrganizationAlertConfigUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceOrganizationAlertConfigUpsert(ctx, d, m)
}

func resourceOrganizationAlertConfigDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	err := fmt.Errorf("Organization AlertConfig Delete is not supported")
	return diag.FromErr(err)
}

func expandOrganizationAlertConfig(in *schema.ResourceData) (*systempb.OrganizationAlertConfiguration, error) {
	log.Println("expand organiztionalertconfig")
	if in == nil {
		return nil, fmt.Errorf("%s", "expandOrganizationAlertConfig empty input")
	}
	obj := &systempb.OrganizationAlertConfiguration{}

	if v, ok := in.Get("metadata").([]interface{}); ok {
		obj.Metadata = expandMetaData(v)
	}

	if v, ok := in.Get("spec").([]interface{}); ok {
		objSpec, err := expandOrganizationAlertConfigSpec(v)
		if err != nil {
			return nil, err
		}
		obj.Spec = objSpec
	}

	obj.ApiVersion = "system.k8smgmt.io/v3"
	obj.Kind = "OrganizationAlertConfiguration"

	return obj, nil
}

func expandOrganizationAlertConfigSpec(p []interface{}) (*systempb.AlertConfigSpec, error) {

	obj := &systempb.AlertConfigSpec{}
	if len(p) == 0 || p[0] == nil {
		return obj, fmt.Errorf("%s", "expandOrganizationAlertConfigSpec empty input")
	}

	in := p[0].(map[string]interface{})

	list_email_interfaces := in["emails"].([]interface{})
	list_email := []string{}
	for x := range list_email_interfaces {
		list_email = append(list_email, list_email_interfaces[x].(string))
	}

	obj.Emails = list_email

	alerts_map := map[string]bool{}
	alerts_interface := in["alerts"].([]interface{})
	alerts_interfaces := alerts_interface[0].(map[string]interface{})
	for key, value := range alerts_interfaces {
		alerts_map[key] = value.(bool)
	}

	alerts := &systempb.AlertsConfig{
		Pod:         alerts_map["pod"],
		Pvc:         alerts_map["pvc"],
		Cluster:     alerts_map["cluster"],
		Node:        alerts_map["node"],
		AgentHealth: alerts_map["agent_health"],
	}

	obj.Alerts = alerts

	return obj, nil
}

// Flatteners

func flattenOrganizationAlertConfig(d *schema.ResourceData, in *systempb.OrganizationAlertConfiguration) error {
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
	ret, err = flattenOrganizationAlertConfigSpec(in.Spec, v)
	if err != nil {
		log.Println("flatten organizationalertconfig spec err")
		return err
	}

	err = d.Set("spec", ret)
	if err != nil {
		log.Println("set spec err")
		return err
	}
	return nil
}

func flattenOrganizationAlertConfigSpec(in *systempb.AlertConfigSpec, p []interface{}) ([]interface{}, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "flattenOrganizationAlertconfigSpec empty input")
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.Emails) > 0 {
		obj["emails"] = in.Emails
	}

	return []interface{}{obj}, nil
}
