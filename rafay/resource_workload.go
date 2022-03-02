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
	commonpb "github.com/RafaySystems/rafay-common/proto/types/hub/commonpb"
	"github.com/RafaySystems/rctl/pkg/versioninfo"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceWorkload() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceWorkloadCreate,
		ReadContext:   resourceWorkloadRead,
		UpdateContext: resourceWorkloadUpdate,
		DeleteContext: resourceWorkloadDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema:        resource.WorkloadSchema.Schema,
	}
}

func resourceWorkloadCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	diags := resourceWorkloadUpsert(ctx, d, m)
	if diags.HasError() {
		tflog := os.Getenv("TF_LOG")
		if tflog == "TRACE" || tflog == "DEBUG" {
			ctx = context.WithValue(ctx, "debug", "true")
		}
		log.Printf("workload create got error, perform cleanup")
		wl, err := expandWorkload(d)
		if err != nil {
			log.Printf("workload expandNamespace error")
			return diags
		}
		auth := config.GetConfig().GetAppAuthProfile()
		client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
		if err != nil {
			return diags
		}

		err = client.AppsV3().Workload().Delete(ctx, options.DeleteOptions{
			Name:    wl.Metadata.Name,
			Project: wl.Metadata.Project,
		})
		if err != nil {
			return diags
		}
	}
	return diags
}
func resourceWorkloadUpsert(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("workload create starts")
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	wl, err := expandWorkload(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.AppsV3().Workload().Apply(ctx, wl, options.ApplyOptions{})
	if err != nil {
		return diag.FromErr(err)
	}

	// wait for publish
	for {
		time.Sleep(30 * time.Second)
		wls, err := client.AppsV3().Workload().Status(ctx, options.StatusOptions{
			Name:    wl.Metadata.Name,
			Project: wl.Metadata.Project,
		})
		if err != nil {
			return diag.FromErr(err)
		}
		log.Println("wls.Status", wls.Status)
		if wls.Status != nil {
			//check if workload can be placed on a cluster, if true break out of loop
			if wls.Status.ConditionStatus == commonpb.ConditionStatus_StatusOK ||
				wls.Status.ConditionStatus == commonpb.ConditionStatus_StatusNotSet {
				break
			}
			if wls.Status.ConditionStatus == commonpb.ConditionStatus_StatusFailed {
				return diag.FromErr(fmt.Errorf("%s", "failed to publish workload"))
			}
		} else {
			break
		}

	}

	d.SetId(wl.Metadata.Name)
	return diags
}

func resourceWorkloadRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	log.Println("resourceWorkloadRead ")

	tfWorkloadState, err := expandWorkload(d)
	if err != nil {
		return diag.FromErr(err)
	}

	// XXX Debug
	// w1 := spew.Sprintf("%+v", tfWorkloadState)
	// log.Println("resourceWorkloadRead tfWorkloadState", w1)

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}

	wl, err := client.AppsV3().Workload().Get(ctx, options.GetOptions{
		Name:    tfWorkloadState.Metadata.Name,
		Project: tfWorkloadState.Metadata.Project,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	// XXX Debug
	// w1 = spew.Sprintf("%+v", wl)
	// log.Println("resourceWorkloadRead wl", w1)

	err = flattenWorkload(d, wl)
	if err != nil {
		return diag.FromErr(err)
	}
	return diags

}

func resourceWorkloadUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceWorkloadUpsert(ctx, d, m)
}

func resourceWorkloadDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}
	wl, err := expandWorkload(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.AppsV3().Workload().Delete(ctx, options.DeleteOptions{
		Name:    wl.Metadata.Name,
		Project: wl.Metadata.Project,
	})

	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func expandWorkload(in *schema.ResourceData) (*appspb.Workload, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "expandWorkload empty input")
	}
	obj := &appspb.Workload{}

	if v, ok := in.Get("metadata").([]interface{}); ok {
		obj.Metadata = expandMetaData(v)
	}

	if v, ok := in.Get("spec").([]interface{}); ok {
		objSpec, err := expandWorkloadSpec(v)
		if err != nil {
			return nil, err
		}
		obj.Spec = objSpec
	}

	obj.ApiVersion = "apps.k8smgmt.io/v3"
	obj.Kind = "Workload"
	return obj, nil
}

func expandDrift(p []interface{}) *commonpb.DriftSpec {
	obj := &commonpb.DriftSpec{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["enabled"].(bool); ok {
		obj.Enabled = v
	}

	if v, ok := in["action"].(string); ok && len(v) > 0 {
		obj.Action = v
	}

	return obj
}

func expandWorkloadSpec(p []interface{}) (*appspb.WorkloadSpec, error) {
	obj := &appspb.WorkloadSpec{}
	if len(p) == 0 || p[0] == nil {
		return obj, fmt.Errorf("%s", "expandWorkloadSpec empty input")
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["namespace"].(string); ok && len(v) > 0 {
		obj.Namespace = v
	}

	if v, ok := in["placement"].([]interface{}); ok {
		obj.Placement = expandPlacement(v)
	}

	if v, ok := in["version"].(string); ok {
		obj.Version = v
	}

	if v, ok := in["drift"].([]interface{}); ok {
		obj.Drift = expandDrift(v)
	}

	if v, ok := in["artifact"].([]interface{}); ok {
		objArtifact, err := ExpandArtifactSpec(v)
		if err != nil {
			return nil, err
		}
		obj.Artifact = objArtifact
	}

	return obj, nil
}

// Flatteners

func flattenWorkload(d *schema.ResourceData, in *appspb.Workload) error {
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

	// XXX Debug
	// w1 := spew.Sprintf("%+v", v)
	// log.Println("flattenWorkload before ", w1)
	var ret []interface{}
	ret, err = flattenWorkloadSpec(in.Spec, v)
	if err != nil {
		return err
	}
	// XXX Debug
	// w1 = spew.Sprintf("%+v", ret)
	// log.Println("flattenWorkload after ", w1)

	err = d.Set("spec", ret)
	if err != nil {
		return err
	}
	return nil
}

func flattenDrift(in *commonpb.DriftSpec) []interface{} {
	if in == nil {
		return nil
	}

	obj := make(map[string]interface{})

	obj["enabled"] = in.Enabled

	if len(in.Action) > 0 {
		obj["action"] = in.Action
	}

	return []interface{}{obj}
}

func flattenWorkloadSpec(in *appspb.WorkloadSpec, p []interface{}) ([]interface{}, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "flattenWorkloadSpec empty input")
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.Namespace) > 0 {
		obj["namespace"] = in.Namespace
	}

	if in.Placement != nil {
		obj["placement"] = flattenPlacement(in.Placement)
	}

	if in.Drift != nil {
		obj["drift"] = flattenDrift(in.Drift)
	}

	if len(in.Version) > 0 {
		obj["version"] = in.Version
	}

	v, ok := obj["artifact"].([]interface{})
	if !ok {
		v = []interface{}{}
	}
	// XXX Debug
	// w1 := spew.Sprintf("%+v", v)
	// log.Println("flattenWorkloadSpec before ", w1)

	var ret []interface{}
	var err error
	ret, err = FlattenArtifactSpec(in.Artifact, v)
	if err != nil {
		log.Println("FlattenArtifactSpec error ", err)
		return nil, err
	}

	// XXX Debug
	// w1 = spew.Sprintf("%+v", ret)
	// log.Println("flattenWorkloadSpec after ", w1)

	obj["artifact"] = ret

	return []interface{}{obj}, nil
}
