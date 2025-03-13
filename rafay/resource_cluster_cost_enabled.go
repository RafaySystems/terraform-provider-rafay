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
	"github.com/RafaySystems/rafay-common/proto/types/hub/costpb"
	"github.com/RafaySystems/rctl/pkg/config"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceClusterCostEnabled() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceClusterCostEnabledCreate,
		ReadContext:   resourceClusterCostEnabledRead,
		UpdateContext: resourceClusterCostEnabledUpdate,
		DeleteContext: resourceClusterCostEnabledDelete,
		Importer: &schema.ResourceImporter{
			State: resourceClusterCostEnabledImport,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema:        resource.ClusterCostSchema.Schema,
	}
}

func resourceClusterCostEnabledCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Println("cluster cost create")
	diags := resourceClusterCostEnabledUpsert(ctx, d, m)
	if diags.HasError() {
		tflog := os.Getenv("TF_LOG")
		if tflog == "TRACE" || tflog == "DEBUG" {
			ctx = context.WithValue(ctx, "debug", "true")
		}
		cc, err := expandClusterCostEnabled(d)
		if err != nil {
			return diags
		}
		auth := config.GetConfig().GetAppAuthProfile()
		client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, TF_USER_AGENT, options.WithInsecureSkipVerify(auth.SkipServerCertValid))
		if err != nil {
			return diags
		}

		err = client.CostV1().ClusterCost().Delete(ctx, options.DeleteOptions{
			Name: cc.Metadata.Name,
		})
		if err != nil {
			return diags
		}
	}
	return diags
}

func resourceClusterCostEnabledUpsert(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("cluster cost upsert starts")
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	cc, err := expandClusterCostEnabled(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, TF_USER_AGENT, options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.CostV1().ClusterCost().Apply(ctx, cc, options.ApplyOptions{})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(cc.Metadata.Name)
	return diags
}

func resourceClusterCostEnabledRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Println("cluster cost read starts ")
	meta := GetMetaData(d)
	if meta == nil {
		return diag.FromErr(fmt.Errorf("%s", "failed to read resource "))
	}
	if d.State() != nil && d.State().ID != "" {
		meta.Name = d.State().ID
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, TF_USER_AGENT, options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		log.Println("read client err")
		return diag.FromErr(err)
	}

	ClusterCostEnabled, err := client.CostV1().ClusterCost().Get(ctx, options.GetOptions{
		Name: meta.Name,
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

	err = flattenClusterCostEnabled(d, ClusterCostEnabled)
	if err != nil {
		log.Println("read flatten err")
		return diag.FromErr(err)
	}
	return diags

}

func resourceClusterCostEnabledUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceClusterCostEnabledUpsert(ctx, d, m)
}

func resourceClusterCostEnabledDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Println("cluster cost delete starts")
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

	cc, err := expandClusterCostEnabled(d)
	if err != nil {
		log.Println("error while expanding cluster cost during delete")
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, TF_USER_AGENT, options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return diag.FromErr(err)
	}
	err = client.CostV1().ClusterCost().Delete(ctx, options.DeleteOptions{
		Name: cc.Metadata.Name,
	})

	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func expandClusterCostEnabled(in *schema.ResourceData) (*costpb.ClusterCost, error) {
	log.Println("expand cluster cost resource")
	if in == nil {
		return nil, fmt.Errorf("%s", "expand cluster cost empty input")
	}
	obj := &costpb.ClusterCost{}

	if v, ok := in.Get("metadata").([]interface{}); ok && len(v) > 0 {
		obj.Metadata = expandV3MetaData(v)
	}

	if v, ok := in.Get("spec").([]interface{}); ok && len(v) > 0 {
		objSpec, err := expandClusterCostEnabledSpec(v)
		if err != nil {
			return nil, err
		}
		obj.Spec = objSpec
	}

	obj.ApiVersion = "cost.management.io/v1"
	obj.Kind = "ClusterCost"
	return obj, nil
}

func expandClusterCostEnabledSpec(p []interface{}) (*costpb.ClusterCostSpec, error) {
	log.Println("expand cluster cost spec")
	obj := &costpb.ClusterCostSpec{}
	if len(p) == 0 || p[0] == nil {
		return obj, fmt.Errorf("expand cluster cost spec empty input")
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["selection_type"].(string); ok && len(v) > 0 {
		obj.SelectionType = v
	}

	if v, ok := in["policy_project"].(string); ok && len(v) > 0 {
		obj.PolicyProject = v
	}

	if v, ok := in["clusters"].([]interface{}); ok && len(v) > 0 {
		obj.Clusters = make([]string, len(v))
		for idx := range v {
			obj.Clusters[idx] = v[idx].(string)
		}
	}

	if v, ok := in["cluster_labels"].([]interface{}); ok && len(v) > 0 {
		obj.ClusterLabels = expandClusterCostLabels(v)
	}

	return obj, nil
}

func expandClusterCostLabels(p []interface{}) []*costpb.ClusterCostCostLabels {
	if len(p) == 0 {
		return []*costpb.ClusterCostCostLabels{}
	}

	out := make([]*costpb.ClusterCostCostLabels, len(p))

	for i := range p {
		if p[i] == nil {
			continue
		}

		obj := costpb.ClusterCostCostLabels{}
		in := p[i].(map[string]interface{})

		if v, ok := in["key"].(string); ok && len(v) > 0 {
			obj.Key = v
		}

		if v, ok := in["value"].(string); ok && len(v) > 0 {
			obj.Value = v
		}

		out[i] = &obj
	}

	return out

}

func flattenClusterCostEnabled(d *schema.ResourceData, in *costpb.ClusterCost) error {
	if in == nil {
		return nil
	}

	err := d.Set("metadata", flattenV3MetaData(in.Metadata))
	if err != nil {
		log.Println("flatten metadata err")
		return err
	}

	v, ok := d.Get("spec").([]interface{})
	if !ok {
		v = []interface{}{}
	}

	var ret []interface{}
	ret, err = flattenClusterCostEnabledSpec(in.Spec, v)
	if err != nil {
		log.Println("flatten cluster cost spec err")
		return err
	}

	err = d.Set("spec", ret)
	if err != nil {
		log.Println("set spec err")
		return err
	}
	return nil
}

func flattenClusterCostEnabledSpec(in *costpb.ClusterCostSpec, p []interface{}) ([]interface{}, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "flatten cluster cost spec empty input")
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}
	if len(in.SelectionType) > 0 {
		obj["selection_type"] = in.SelectionType
	}

	if len(in.PolicyProject) > 0 {
		obj["policy_project"] = in.PolicyProject
	}

	if len(in.Clusters) > 0 {
		obj["clusters"] = in.Clusters
	}

	v, ok := obj["cluster_labels"].([]interface{})
	if !ok {
		v = []interface{}{}
	}
	obj["cluster_labels"] = flattenClusterCostEnabledLabels(in.ClusterLabels, v)

	return []interface{}{obj}, nil
}

func flattenClusterCostEnabledLabels(in []*costpb.ClusterCostCostLabels, p []interface{}) []interface{} {
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

		out[i] = obj
	}

	return out
}

func resourceClusterCostEnabledImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {

	log.Printf("Cluster Cost Enabled Import Starts")

	idParts := strings.SplitN(d.Id(), "/", 2)
	log.Println("resourceClusterCostEnabledImport idParts:", idParts)

	log.Println("resourceClusterCostEnabledImport Invoking expandClusterCostEnabled")
	cc, err := expandClusterCostEnabled(d)
	if err != nil {
		log.Printf("resourceClusterCostEnabledImport  expand error %s", err.Error())
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
