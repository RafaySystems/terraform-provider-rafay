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
	"github.com/RafaySystems/rafay-common/proto/types/hub/settingspb"
	"github.com/RafaySystems/rctl/pkg/versioninfo"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAlertConfig() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAlertConfigCreate,
		ReadContext:   resourceAlertConfigRead,
		UpdateContext: resourceAlertConfigUpdate,
		DeleteContext: resourceAlertConfigDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema:        resource.AlertConfigurationSchema.Schema,
	}
}

func resourceAlertConfigCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Println("alert config create")
	diags := resourceAlertConfigUpsert(ctx, d, m)
	if diags.HasError() {
		tflog := os.Getenv("TF_LOG")
		if tflog == "TRACE" || tflog == "DEBUG" {
			ctx = context.WithValue(ctx, "debug", "true")
		}
		cg, err := expandAlertConfig(d)
		if err != nil {
			return diags
		}
		auth := config.GetConfig().GetAppAuthProfile()
		client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
		if err != nil {
			return diags
		}

		err = client.SettingsV3().AlertConfiguration().Delete(ctx, options.DeleteOptions{
			Name:    cg.Metadata.Name,
			Project: cg.Metadata.Project,
		})
		if err != nil {
			return diags
		}
	}
	return diags
}
func resourceAlertConfigUpsert(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("alertconfig upsert starts")
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	ac, err := expandAlertConfig(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.SettingsV3().AlertConfiguration().Apply(ctx, ac, options.ApplyOptions{})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(ac.Metadata.Name)
	return diags
}

func resourceAlertConfigRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	log.Println("resourceAlertConfigRead ")

	meta := GetMetaData(d)
	if meta == nil {
		return diag.FromErr(fmt.Errorf("%s", "failed to read resource "))
	}
	if d.State() != nil && d.State().ID != "" {
		meta.Name = d.State().ID
	}

	tfAlertconfigState, err := expandAlertConfig(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		log.Println("read client err")
		return diag.FromErr(err)
	}

	ac, err := client.SettingsV3().AlertConfiguration().Get(ctx, options.GetOptions{
		Name:    meta.Name,
		Project: tfAlertconfigState.Metadata.Project,
	})
	if err != nil {
		log.Println("read get err")
		return diag.FromErr(err)
	}

	err = flattenAlertConfig(d, ac)
	if err != nil {
		log.Println("read flatten err")
		return diag.FromErr(err)
	}
	return diags

}

func resourceAlertConfigUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceAlertConfigUpsert(ctx, d, m)
}

func resourceAlertConfigDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	err := fmt.Errorf("AlertConfig Delete is not supported")
	return diag.FromErr(err)
}

func expandAlertConfig(in *schema.ResourceData) (*settingspb.AlertConfiguration, error) {
	log.Println("expand alertconfig")
	if in == nil {
		return nil, fmt.Errorf("%s", "expandAlertConfig empty input")
	}
	obj := &settingspb.AlertConfiguration{}

	if v, ok := in.Get("metadata").([]interface{}); ok {
		obj.Metadata = expandMetaData(v)
	}

	if v, ok := in.Get("spec").([]interface{}); ok {
		objSpec, err := expandAlertConfigSpec(v)
		if err != nil {
			return nil, err
		}
		obj.Spec = objSpec
	}

	obj.ApiVersion = "settings.k8smgmt.io/v3"
	obj.Kind = "AlertConfiguration"

	return obj, nil
}

func expandAlertConfigSpec(p []interface{}) (*settingspb.AlertConfigSpec, error) {
	log.Println("expand alertconfig spec")
	obj := &settingspb.AlertConfigSpec{}
	if len(p) == 0 || p[0] == nil {
		return obj, fmt.Errorf("%s", "expandAlertConfigSpec empty input")
	}

	in := p[0].(map[string]interface{})

	list_email_interfaces := in["emails"].([]interface{})
	list_email := []string{}
	for x := range list_email_interfaces {
		list_email = append(list_email, list_email_interfaces[x].(string))
	}

	obj.Emails = list_email

	alerts_map := map[string]bool{}
	alerts_interfaces := in["alerts"].(map[string]interface{})
	for key, value := range alerts_interfaces {
		alerts_map[key] = value.(bool)
	}
	obj.Alerts = alerts_map

	return obj, nil
}

// Flatteners

func flattenAlertConfig(d *schema.ResourceData, in *settingspb.AlertConfiguration) error {
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
	ret, err = flattenAlertConfigSpec(in.Spec, v)
	if err != nil {
		log.Println("flatten alertconfig spec err")
		return err
	}

	err = d.Set("spec", ret)
	if err != nil {
		log.Println("set spec err")
		return err
	}
	return nil
}

func flattenAlertConfigSpec(in *settingspb.AlertConfigSpec, p []interface{}) ([]interface{}, error) {

	if in == nil {
		return nil, fmt.Errorf("%s", "flattenAlertconfigSpec empty input")
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.Emails) > 0 {
		obj["emails"] = in.Emails
	}

	if len(in.Alerts) > 0 {
		obj["alerts"] = in.Alerts
	}

	return []interface{}{obj}, nil
}
