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

func resourceCostOptimisation() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCostOptimisationCreate,
		ReadContext:   resourceCostOptimisationRead,
		UpdateContext: resourceCostOptimisationUpdate,
		DeleteContext: resourceCostOptimisationDelete,
		Importer: &schema.ResourceImporter{
			State: resourceCostOptimisationImport,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema:        resource.CostOptimisationSchema.Schema,
	}
}

func resourceCostOptimisationCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Println("cost optimisation create")
	diags := resourceCostOptimisationUpsert(ctx, d, m)
	if diags.HasError() {
		tflog := os.Getenv("TF_LOG")
		if tflog == "TRACE" || tflog == "DEBUG" {
			ctx = context.WithValue(ctx, "debug", "true")
		}
		cc, err := expandCostOptimisation(d)
		if err != nil {
			return diags
		}
		auth := config.GetConfig().GetAppAuthProfile()
		client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, TF_USER_AGENT, options.WithInsecureSkipVerify(auth.SkipServerCertValid))
		if err != nil {
			return diags
		}

		err = client.CostV1().CostOptimisation().Delete(ctx, options.DeleteOptions{
			Name: cc.Metadata.Name,
		})
		if err != nil {
			return diags
		}
	}
	return diags
}

func resourceCostOptimisationUpsert(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("cost optimisation upsert starts")
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	cc, err := expandCostOptimisation(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, TF_USER_AGENT, options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.CostV1().CostOptimisation().Apply(ctx, cc, options.ApplyOptions{})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(cc.Metadata.Name)
	return diags
}

func resourceCostOptimisationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Println("cost optimisation read starts ")
	meta := GetMetaData(d)
	if meta == nil {
		return diag.FromErr(fmt.Errorf("%s", "failed to read resource "))
	}
	if d.State() != nil && d.State().ID != "" {
		meta.Name = d.State().ID
	}

	_, err := expandCostOptimisation(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, TF_USER_AGENT, options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		log.Println("read client err")
		return diag.FromErr(err)
	}

	costoptimisation, err := client.CostV1().CostOptimisation().Get(ctx, options.GetOptions{
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

	err = flattenCostOptimisation(d, costoptimisation)
	if err != nil {
		log.Println("read flatten err")
		return diag.FromErr(err)
	}
	return diags

}

func resourceCostOptimisationUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceCostOptimisationUpsert(ctx, d, m)
}

func resourceCostOptimisationDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Println("cost optimisation delete starts")
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

	cc, err := expandCostOptimisation(d)
	if err != nil {
		log.Println("error while expanding cost optimisation during delete")
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, TF_USER_AGENT, options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return diag.FromErr(err)
	}
	err = client.CostV1().CostOptimisation().Delete(ctx, options.DeleteOptions{
		Name: cc.Metadata.Name,
	})

	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func expandCostOptimisation(in *schema.ResourceData) (*costpb.CostOptimisation, error) {
	log.Println("expand cost optimisation resource")
	if in == nil {
		return nil, fmt.Errorf("%s", "expand cost optimisation empty input")
	}
	obj := &costpb.CostOptimisation{}

	if v, ok := in.Get("metadata").([]interface{}); ok && len(v) > 0 {
		obj.Metadata = expandV3MetaData(v)
	}

	if v, ok := in.Get("spec").([]interface{}); ok && len(v) > 0 {
		objSpec, err := expandCostOptimisationSpec(v)
		if err != nil {
			return nil, err
		}
		obj.Spec = objSpec
	}

	obj.ApiVersion = "cost.optimisation.io/v1"
	obj.Kind = "CostOptimisation"
	return obj, nil
}

func expandCostOptimisationSpec(p []interface{}) (*costpb.CostOptimisationSpec, error) {
	log.Println("expand cost optimisation spec")
	spec := &costpb.CostOptimisationSpec{}
	if len(p) == 0 || p[0] == nil {
		return spec, fmt.Errorf("expand cost optimisation spec empty input")
	}

	in := p[0].(map[string]interface{})

	spec.SelectionType = in["selection_type"].(string)
	spec.ConfigProject = in["config_project"].(string)

	if v, ok := in["clusters"].([]interface{}); ok && len(v) > 0 {
		var clusters []string
		for _, cluster := range v {
			clusters = append(clusters, cluster.(string))
		}
		spec.Clusters = clusters
	}

	spec.Period = int64(in["period"].(int))
	spec.Mode = in["mode"].(string)
	spec.Recommended = int64(in["recommended"].(int))

	if v, ok := in["cluster_labels"].([]interface{}); ok && len(v) > 0 {
		spec.ClusterLabels = expandCostOptClusterLabels(v)
	}

	if v, ok := in["inclusions"].([]interface{}); ok && len(v) > 0 {
		spec.Inclusions = expandCostOptFilter(v)
	}

	if v, ok := in["exclusions"].([]interface{}); ok && len(v) > 0 {
		spec.Exclusions = expandCostOptFilter(v)
	}

	if v, ok := in["bound"].([]interface{}); ok && len(v) > 0 {
		spec.Bound = expandCostOptBound(v)
	}

	if v, ok := in["min_threshold"].([]interface{}); ok && len(v) > 0 {
		spec.MinThreshold = expandCostOptMinimumThreshold(v)
	}

	return spec, nil
}

func expandCostOptClusterLabels(p []interface{}) []*costpb.CostOptimisationLabels {
	if len(p) == 0 || p[0] == nil {
		return []*costpb.CostOptimisationLabels{}
	}

	clusterLables := make([]*costpb.CostOptimisationLabels, len(p))

	for i := range p {
		obj := costpb.CostOptimisationLabels{}
		in := p[i].(map[string]interface{})

		if v, ok := in["key"].(string); ok && len(v) > 0 {
			obj.Key = v
		}

		if v, ok := in["value"].(string); ok && len(v) > 0 {
			obj.Value = v
		}

		clusterLables[i] = &obj

	}

	return clusterLables
}

func expandCostOptFilter(p []interface{}) []*costpb.CostOptimisationFilter {
	if len(p) == 0 || p[0] == nil {
		return []*costpb.CostOptimisationFilter{}
	}

	filter := make([]*costpb.CostOptimisationFilter, len(p))

	for i := range p {
		obj := costpb.CostOptimisationFilter{}
		in := p[i].(map[string]interface{})

		if v, ok := in["namespace"].(string); ok && len(v) > 0 {
			obj.Namespace = v
		}

		if v, ok := in["namespace_label"].([]interface{}); ok && len(v) > 0 {
			var labels []string
			for _, label := range v {
				labels = append(labels, label.(string))
			}
			obj.NamespaceLabel = labels
		}

		filter[i] = &obj

	}

	return filter
}

func expandCostOptBound(p []interface{}) *costpb.CostOptimisationBound {
	bound := &costpb.CostOptimisationBound{
		Cpu:    &costpb.CostOptimisationCPUBound{},
		Memory: &costpb.CostOptimisationMemoryBound{},
	}
	if len(p) == 0 || p[0] == nil {
		return bound
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["cpu"].([]interface{}); ok && len(v) > 0 {
		cpu := v[0].(map[string]interface{})
		if min, ok := cpu["minimum"].(string); ok && len(min) > 0 {
			bound.Cpu.Minimum = min
		}
		if max, ok := cpu["maximum"].(string); ok && len(max) > 0 {
			bound.Cpu.Maximum = max
		}
	}

	if v, ok := in["memory"].([]interface{}); ok && len(v) > 0 {
		memory := v[0].(map[string]interface{})
		if min, ok := memory["minimum"].(string); ok && len(min) > 0 {
			bound.Memory.Minimum = min
		}
		if max, ok := memory["maximum"].(string); ok && len(max) > 0 {
			bound.Memory.Maximum = max
		}
	}

	return bound
}

func expandCostOptMinimumThreshold(p []interface{}) *costpb.CostOptimisationMinimumThreshold {
	minThreshold := &costpb.CostOptimisationMinimumThreshold{
		Cpu:    &costpb.CostOptimisationCPUThreshold{},
		Memory: &costpb.CostOptimisationMemoryThreshold{},
	}
	if len(p) == 0 || p[0] == nil {
		return minThreshold
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["cpu"].([]interface{}); ok && len(v) > 0 {
		cpu := v[0].(map[string]interface{})
		if pct, ok := cpu["percentage"].(string); ok && len(pct) > 0 {
			minThreshold.Cpu.Percentage = pct
		}
		if unit, ok := cpu["unit"].(string); ok && len(unit) > 0 {
			minThreshold.Cpu.Unit = unit
		}
	}

	if v, ok := in["memory"].([]interface{}); ok && len(v) > 0 {
		memory := v[0].(map[string]interface{})
		if pct, ok := memory["percentage"].(string); ok && len(pct) > 0 {
			minThreshold.Memory.Percentage = pct
		}
		if unit, ok := memory["unit"].(string); ok && len(unit) > 0 {
			minThreshold.Memory.Unit = unit
		}
	}

	return minThreshold
}

// Flatteners

func flattenCostOptimisation(d *schema.ResourceData, in *costpb.CostOptimisation) error {
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
	ret, err = flattenCostOptimisationSpec(in.Spec, v)
	if err != nil {
		log.Println("flatten cost optimisation spec err")
		return err
	}

	err = d.Set("spec", ret)
	if err != nil {
		log.Println("set spec err")
		return err
	}
	return nil
}

func flattenCostOptimisationSpec(in *costpb.CostOptimisationSpec, p []interface{}) ([]interface{}, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "flatten cost optimisation spec empty input")
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.SelectionType) > 0 {
		obj["selection_type"] = in.SelectionType
	}

	if len(in.ConfigProject) > 0 {
		obj["config_project"] = in.ConfigProject
	}

	if len(in.Clusters) > 0 {
		obj["clusters"] = in.Clusters
	}

	if in.Period != 0 {
		obj["period"] = in.Period
	}

	if len(in.Mode) > 0 {
		obj["mode"] = in.Mode
	}

	if in.Recommended != 0 {
		obj["recommended"] = in.Recommended
	}

	if len(in.ClusterLabels) > 0 {
		v, ok := obj["cluster_labels"].([]interface{})
		if !ok {
			v = []interface{}{}
		}

		obj["cluster_labels"] = flattenCostOptClusterLabels(in.ClusterLabels, v)
	}

	if len(in.Inclusions) > 0 {
		v, ok := obj["inclusions"].([]interface{})
		if !ok {
			v = []interface{}{}
		}

		obj["inclusions"] = flattenCostOptFilter(in.Inclusions, v)
	}

	if len(in.Exclusions) > 0 {
		v, ok := obj["exclusions"].([]interface{})
		if !ok {
			v = []interface{}{}
		}

		obj["exclusions"] = flattenCostOptFilter(in.Exclusions, v)
	}

	if in.Bound != nil {
		v, ok := obj["bound"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["bound"] = flattenCostOptBound(in.Bound, v)
	}

	if in.MinThreshold != nil {
		v, ok := obj["min_threshold"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["min_threshold"] = flattenCostOptMinimumThreshold(in.MinThreshold, v)
	}

	return []interface{}{obj}, nil
}

func flattenCostOptClusterLabels(input []*costpb.CostOptimisationLabels, p []interface{}) []interface{} {
	if input == nil {
		return nil
	}

	out := make([]interface{}, len(input))

	for i, in := range input {
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

func flattenCostOptFilter(input []*costpb.CostOptimisationFilter, p []interface{}) []interface{} {
	if input == nil {
		return nil
	}

	out := make([]interface{}, len(input))

	for i, in := range input {
		obj := map[string]interface{}{}
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}

		if len(in.Namespace) > 0 {
			obj["namespace"] = in.Namespace
		}

		if len(in.NamespaceLabel) > 0 {
			obj["namespace_label"] = in.NamespaceLabel
		}

		out[i] = &obj
	}

	return out
}

func flattenCostOptBound(in *costpb.CostOptimisationBound, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if in.Cpu != nil {
		v, ok := obj["cpu"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["cpu"] = flattenCostOptCPUBound(in.Cpu, v)
	}

	if in.Memory != nil {
		v, ok := obj["memory"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["memory"] = flattenCostOptMemoryBound(in.Memory, v)
	}

	return []interface{}{obj}
}

func flattenCostOptCPUBound(in *costpb.CostOptimisationCPUBound, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.Minimum) > 0 {
		obj["minimum"] = in.Minimum
	}

	if len(in.Maximum) > 0 {
		obj["maximum"] = in.Maximum
	}

	return []interface{}{obj}
}

func flattenCostOptMemoryBound(in *costpb.CostOptimisationMemoryBound, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.Minimum) > 0 {
		obj["minimum"] = in.Minimum
	}

	if len(in.Maximum) > 0 {
		obj["maximum"] = in.Maximum
	}

	return []interface{}{obj}
}

func flattenCostOptMinimumThreshold(in *costpb.CostOptimisationMinimumThreshold, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if in.Cpu != nil {
		v, ok := obj["cpu"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["cpu"] = flattenCostOptCPUThreshold(in.Cpu, v)
	}

	if in.Memory != nil {
		v, ok := obj["memory"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["memory"] = flattenCostOptMemoryThreshold(in.Memory, v)
	}

	return []interface{}{obj}
}

func flattenCostOptCPUThreshold(in *costpb.CostOptimisationCPUThreshold, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.Percentage) > 0 {
		obj["percentage"] = in.Percentage
	}

	if len(in.Unit) > 0 {
		obj["unit"] = in.Unit
	}

	return []interface{}{obj}
}

func flattenCostOptMemoryThreshold(in *costpb.CostOptimisationMemoryThreshold, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.Percentage) > 0 {
		obj["percentage"] = in.Percentage
	}

	if len(in.Unit) > 0 {
		obj["unit"] = in.Unit
	}

	return []interface{}{obj}
}

func resourceCostOptimisationImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {

	log.Printf("Cost Optimisation Import Starts")

	idParts := strings.SplitN(d.Id(), "/", 2)
	log.Println("resourceCostOptimisationImport idParts:", idParts)

	log.Println("resourceCostOptimisationImport Invoking expandCostOptimisation")
	cc, err := expandCostOptimisation(d)
	if err != nil {
		log.Printf("resourceCostOptimisationImport  expand error %s", err.Error())
	}

	var metaD commonpb.Metadata
	metaD.Name = idParts[0]
	cc.Metadata = &metaD

	err = d.Set("metadata", flattenV3MetaData(&metaD))
	if err != nil {
		log.Println("import set metadata err ", err)
		return nil, err
	}
	d.SetId(cc.Metadata.Name)
	return []*schema.ResourceData{d}, nil

}
