package rafay

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/RafaySystems/rafay-common/pkg/hub/client/options"
	"github.com/RafaySystems/rafay-common/pkg/hub/client/typed"
	"github.com/RafaySystems/rafay-common/pkg/hub/terraform/resource"
	"github.com/RafaySystems/rafay-common/proto/types/hub/commonpb"
	"github.com/RafaySystems/rafay-common/proto/types/hub/eaaspb"
	"github.com/RafaySystems/rctl/pkg/config"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceDriver() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDriverCreate,
		ReadContext:   resourceDriverRead,
		UpdateContext: resourceDriverUpdate,
		DeleteContext: resourceDriverDelete,
		Importer: &schema.ResourceImporter{
			State: resourceDriverImport,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema:        resource.DriverSchema.Schema,
	}
}

func resourceDriverCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	log.Println("driver create")
	diags := resourceDriverUpsert(ctx, d, m)
	if diags.HasError() {
		tflog := os.Getenv("TF_LOG")
		if tflog == "TRACE" || tflog == "DEBUG" {
			ctx = context.WithValue(ctx, "debug", "true")
		}
		cc, err := expandDriver(d)
		if err != nil {
			return diags
		}
		auth := config.GetConfig().GetAppAuthProfile()
		client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, TF_USER_AGENT, options.WithInsecureSkipVerify(auth.SkipServerCertValid))
		if err != nil {
			return diags
		}

		err = client.EaasV1().Driver().Delete(ctx, options.DeleteOptions{
			Name:    cc.Metadata.Name,
			Project: cc.Metadata.Project,
		})
		if err != nil {
			return diags
		}
	}
	return diags
}

func resourceDriverUpsert(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("driver upsert starts")
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	cc, err := expandDriver(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, TF_USER_AGENT, options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.EaasV1().Driver().Apply(ctx, cc, options.ApplyOptions{})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(cc.Metadata.Name)
	return diags
}

func resourceDriverRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Println("driver read starts ")
	meta := GetMetaData(d)
	if meta == nil {
		return diag.FromErr(fmt.Errorf("%s", "failed to read resource "))
	}
	if d.State() != nil && d.State().ID != "" {
		meta.Name = d.State().ID
	}

	cc, err := expandDriver(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, TF_USER_AGENT, options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		log.Println("read client err")
		return diag.FromErr(err)
	}

	driver, err := client.EaasV1().Driver().Get(ctx, options.GetOptions{
		Name:    meta.Name,
		Project: cc.Metadata.Project,
	})
	if err != nil {
		log.Println("read get err")
		if strings.Contains(err.Error(), "code 404") {
			log.Println("Resource Read ", "error", err)
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	if cc.GetSpec().GetSharing() != nil && !cc.GetSpec().GetSharing().GetEnabled() && driver.GetSpec().GetSharing() == nil {
		driver.Spec.Sharing = &commonpb.SharingSpec{}
		driver.Spec.Sharing.Enabled = false
		driver.Spec.Sharing.Projects = cc.GetSpec().GetSharing().GetProjects()
	}

	err = flattenDriver(d, driver)
	if err != nil {
		log.Println("read flatten err")
		return diag.FromErr(err)
	}
	return diags

}

func resourceDriverUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	return resourceDriverUpsert(ctx, d, m)
}

func resourceDriverDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Println("driver delete starts")
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

	cc, err := expandDriver(d)
	if err != nil {
		log.Println("error while expanding driver during delete")
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, TF_USER_AGENT, options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return diag.FromErr(err)
	}
	err = client.EaasV1().Driver().Delete(ctx, options.DeleteOptions{
		Name:    cc.Metadata.Name,
		Project: cc.Metadata.Project,
	})

	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func expandDriver(in *schema.ResourceData) (*eaaspb.Driver, error) {
	log.Println("expand driver resource")
	if in == nil {
		return nil, fmt.Errorf("%s", "expand driver empty input")
	}
	obj := &eaaspb.Driver{}

	if v, ok := in.Get("metadata").([]any); ok && len(v) > 0 {
		obj.Metadata = expandV3MetaData(v)
	}

	if v, ok := in.Get("spec").([]any); ok && len(v) > 0 {
		objSpec, err := expandDriverSpec(v)
		if err != nil {
			return nil, err
		}
		obj.Spec = objSpec
	}

	obj.ApiVersion = "eaas.envmgmt.io/v1"
	obj.Kind = "Driver"
	return obj, nil
}

func expandDriverSpec(p []any) (*eaaspb.DriverSpec, error) {
	log.Println("expand driver spec")
	spec := &eaaspb.DriverSpec{}
	if len(p) == 0 || p[0] == nil {
		return spec, fmt.Errorf("%s", "expand driver spec empty input")
	}

	in := p[0].(map[string]any)

	if c, ok := in["config"].([]any); ok && len(c) > 0 {
		spec.Config = expandDriverConfig(c)
	}

	if v, ok := in["sharing"].([]any); ok && len(v) > 0 {
		spec.Sharing = expandSharingSpec(v)
	}

	if v, ok := in["inputs"].([]any); ok && len(v) > 0 {
		spec.Inputs = expandConfigContextCompoundRefs(v)
	}

	var err error
	if v, ok := in["outputs"].(string); ok && len(v) > 0 {
		spec.Outputs, err = expandWorkflowHandlerOutputs(v)
		if err != nil {
			return nil, err
		}
	}

	return spec, nil
}

func expandDriverConfig(p []any) *eaaspb.DriverConfig {
	driverConfig := eaaspb.DriverConfig{}
	if len(p) == 0 || p[0] == nil {
		return &driverConfig
	}

	in := p[0].(map[string]any)

	if typ, ok := in["type"].(string); ok && len(typ) > 0 {
		driverConfig.Type = typ
	}

	if ts, ok := in["timeout_seconds"].(int); ok {
		driverConfig.TimeoutSeconds = int64(ts)
	}

	if sc, ok := in["success_condition"].(string); ok && len(sc) > 0 {
		driverConfig.SuccessCondition = sc
	}

	if ts, ok := in["max_retry_count"].(int); ok {
		driverConfig.MaxRetryCount = int32(ts)
	}

	if v, ok := in["container"].([]any); ok && len(v) > 0 {
		driverConfig.Container = expandWorkflowHandlerContainerConfig(v)
	}

	if v, ok := in["http"].([]any); ok && len(v) > 0 {
		driverConfig.Http = expandWorkflowHandlerHttpConfig(v)
	}

	return &driverConfig
}

// Flatteners

func flattenDriver(d *schema.ResourceData, in *eaaspb.Driver) error {
	if in == nil {
		return nil
	}

	err := d.Set("metadata", flattenV3MetaData(in.Metadata))
	if err != nil {
		log.Println("flatten metadata err")
		return err
	}

	v, ok := d.Get("spec").([]any)
	if !ok {
		v = []any{}
	}

	var ret []any
	ret, err = flattenDriverSpec(in.Spec, v)
	if err != nil {
		log.Println("flatten driver spec err")
		return err
	}

	err = d.Set("spec", ret)
	if err != nil {
		log.Println("set spec err")
		return err
	}
	return nil
}

func flattenDriverSpec(in *eaaspb.DriverSpec, p []any) ([]any, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "flatten driver spec empty input")
	}

	obj := map[string]any{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]any)
	}
	obj["config"] = flattenDriverConfig(in.Config, obj["config"].([]any))
	obj["sharing"] = flattenSharingSpec(in.Sharing)
	obj["inputs"] = flattenConfigContextCompoundRefs(in.Inputs)
	obj["outputs"] = flattenWorkflowHandlerOutputs(in.Outputs)
	return []any{obj}, nil
}

func flattenDriverConfig(input *eaaspb.DriverConfig, p []any) []any {
	log.Println("flatten driver config start", input)
	if input == nil {
		return nil
	}

	obj := map[string]any{}
	if len(p) > 0 && p[0] != nil {
		obj = p[0].(map[string]any)
	}

	obj["type"] = input.Type
	obj["timeout_seconds"] = input.TimeoutSeconds
	obj["success_condition"] = input.SuccessCondition
	obj["max_retry_count"] = input.MaxRetryCount
	obj["container"] = flattenWorkflowHandlerContainerConfig(input.Container, obj["container"].([]any))
	obj["http"] = flattenWorkflowHandlerHttpConfig(input.Http, obj["http"].([]any))

	return []any{obj}
}

func resourceDriverImport(d *schema.ResourceData, m any) ([]*schema.ResourceData, error) {
	log.Printf("Driver Import Starts")

	idParts := strings.SplitN(d.Id(), "/", 2)
	log.Println("resourceDriverImport idParts:", idParts)

	log.Println("resourceDriverImport Invoking expandDriver")
	cc, err := expandDriver(d)
	if err != nil {
		log.Printf("resourceDriverImport  expand error %s", err.Error())
	}

	var metaD commonpb.Metadata
	metaD.Name = idParts[0]
	metaD.Project = idParts[1]
	cc.Metadata = &metaD

	err = d.Set("metadata", flattenV3MetaData(&metaD))
	if err != nil {
		log.Println("import set metadata err ", err)
		return nil, err
	}
	d.SetId(cc.Metadata.Name)
	return []*schema.ResourceData{d}, nil
}
