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

func resourceNodeCost() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceNodeCostCreate,
		ReadContext:   resourceNodeCostRead,
		UpdateContext: resourceNodeCostUpdate,
		DeleteContext: resourceNodeCostDelete,
		Importer: &schema.ResourceImporter{
			State: resourceNodeCostImport,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema:        resource.NodeCostSchema.Schema,
	}
}

func resourceNodeCostCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Println("node cost create")
	diags := resourceNodeCostUpsert(ctx, d, m)
	if diags.HasError() {
		tflog := os.Getenv("TF_LOG")
		if tflog == "TRACE" || tflog == "DEBUG" {
			ctx = context.WithValue(ctx, "debug", "true")
		}
		cc, err := expandNodeCost(d)
		if err != nil {
			return diags
		}
		auth := config.GetConfig().GetAppAuthProfile()
		client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, TF_USER_AGENT, options.WithInsecureSkipVerify(auth.SkipServerCertValid))
		if err != nil {
			return diags
		}

		err = client.CostV1().NodeCost().Delete(ctx, options.DeleteOptions{
			Name: cc.Metadata.Name,
			// Project: cc.Metadata.Project,
		})
		if err != nil {
			return diags
		}
	}
	return diags
}

func resourceNodeCostUpsert(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("config context upsert starts")
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	cc, err := expandNodeCost(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, TF_USER_AGENT, options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.CostV1().NodeCost().Apply(ctx, cc, options.ApplyOptions{})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(cc.Metadata.Name)
	return diags
}

func resourceNodeCostRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Println("config context read starts ")
	meta := GetMetaData(d)
	if meta == nil {
		return diag.FromErr(fmt.Errorf("%s", "failed to read resource "))
	}
	if d.State() != nil && d.State().ID != "" {
		meta.Name = d.State().ID
	}

	// cc, err := expandNodeCost(d)
	// if err != nil {
	// 	return diag.FromErr(err)
	// }

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, TF_USER_AGENT, options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		log.Println("read client err")
		return diag.FromErr(err)
	}

	nodecost, err := client.CostV1().NodeCost().Get(ctx, options.GetOptions{
		Name: meta.Name,
		// Project: cc.Metadata.Project,
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

	err = flattenNodeCost(d, nodecost)
	if err != nil {
		log.Println("read flatten err")
		return diag.FromErr(err)
	}
	return diags

}

func resourceNodeCostUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceNodeCostUpsert(ctx, d, m)
}

func resourceNodeCostDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Println("config context delete starts")
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

	cc, err := expandNodeCost(d)
	if err != nil {
		log.Println("error while expanding config context during delete")
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, TF_USER_AGENT, options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return diag.FromErr(err)
	}
	err = client.CostV1().NodeCost().Delete(ctx, options.DeleteOptions{
		Name: cc.Metadata.Name,
		// Project: cc.Metadata.Project,
	})

	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func expandNodeCost(in *schema.ResourceData) (*costpb.NodeCost, error) {
	log.Println("expand config context resource")
	if in == nil {
		return nil, fmt.Errorf("%s", "expand config context empty input")
	}
	obj := &costpb.NodeCost{}

	if v, ok := in.Get("metadata").([]interface{}); ok && len(v) > 0 {
		obj.Metadata = expandV3MetaData(v)
	}

	if v, ok := in.Get("spec").([]interface{}); ok && len(v) > 0 {
		objSpec, err := expandNodeCostSpec(v)
		if err != nil {
			return nil, err
		}
		obj.Spec = objSpec
	}

	obj.ApiVersion = "cost.management.io/v1"
	obj.Kind = "NodeCost"
	return obj, nil
}

func expandNodeCostSpec(p []interface{}) (*costpb.NodeCostSpec, error) {
	log.Println("expand config context spec")
	spec := &costpb.NodeCostSpec{}
	if len(p) == 0 || p[0] == nil {
		return spec, fmt.Errorf("expand config context spec empty input")
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["cost_values"].([]interface{}); ok && len(v) > 0 {
		spec.CostValues = expandCostValue(v)
	}

	if v, ok := in["node_labels"].([]interface{}); ok && len(v) > 0 {
		spec.NodeLabels = expandNodeCostLabels(v)
	}

	if v, ok := in["currency"].([]interface{}); ok && len(v) > 0 {
		spec.Currency = expandCurrencyType(v)
	}

	// if v, ok := in["sharing"].([]interface{}); ok && len(v) > 0 {
	// 	spec.Sharing = expandSharingSpec(v)
	// }

	return spec, nil
}
func expandCostValue(p []interface{}) *costpb.NodeCostValue {
	obj := &costpb.NodeCostValue{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["cpu"].(string); ok && len(v) > 0 {
		obj.Cpu = v
	}

	if v, ok := in["gpu"].(string); ok && len(v) > 0 {
		obj.Gpu = v
	}

	if v, ok := in["memory"].(string); ok && len(v) > 0 {
		obj.Memory = v
	}

	return obj

}
func expandCurrencyType(p []interface{}) *costpb.NodeCostCurrency {
	obj := &costpb.NodeCostCurrency{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["type"].(string); ok && len(v) > 0 {
		obj.Type = v
	}

	return obj

}
func expandNodeCostLabels(p []interface{}) []*costpb.NodeCostLabels {
	if len(p) == 0 {
		return []*costpb.NodeCostLabels{}
	}

	out := make([]*costpb.NodeCostLabels, len(p))

	for i := range p {
		if p[i] == nil {
			continue
		}

		obj := costpb.NodeCostLabels{}
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

func flattenNodeCost(d *schema.ResourceData, in *costpb.NodeCost) error {
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
	ret, err = flattenNodeCostSpec(in.Spec, v)
	if err != nil {
		log.Println("flatten config context spec err")
		return err
	}

	err = d.Set("spec", ret)
	if err != nil {
		log.Println("set spec err")
		return err
	}
	return nil
}

func flattenNodeCostSpec(in *costpb.NodeCostSpec, p []interface{}) ([]interface{}, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "flatten config context spec empty input")
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}
	if in.Currency != nil {
		v, ok := obj["currency"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["currency"] = flattenNodeCostCurreny(in.Currency, v)
	}
	if in.CostValues != nil {
		v, ok := obj["cost_values"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["cost_values"] = flattenNodeCostValue(in.CostValues, v)
	}
	if len(in.NodeLabels) > 0 {
		v, ok := obj["node_labels"].([]interface{})
		if !ok {
			v = []interface{}{}
		}

		obj["node_labels"] = flattenNodeCostLabels(in.NodeLabels, v)
	}

	return []interface{}{obj}, nil
}
func flattenNodeCostValue(in *costpb.NodeCostValue, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.Cpu) > 0 {
		obj["cpu"] = in.Cpu
	}

	if len(in.Gpu) > 0 {
		obj["gpu"] = in.Gpu
	}

	if len(in.Memory) > 0 {
		obj["memory"] = in.Memory
	}

	return []interface{}{obj}
}
func flattenNodeCostCurreny(in *costpb.NodeCostCurrency, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.Type) > 0 {
		obj["type"] = in.Type
	}

	return []interface{}{obj}
}

func flattenNodeCostLabels(in []*costpb.NodeCostLabels, p []interface{}) []interface{} {
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

func resourceNodeCostImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {

	log.Printf("Config Context Import Starts")

	idParts := strings.SplitN(d.Id(), "/", 2)
	log.Println("resourceConfigContextImport idParts:", idParts)

	log.Println("resourceConfigContextImport Invoking expandConfigContext")
	cc, err := expandNodeCost(d)
	if err != nil {
		log.Printf("resourceConfigContextImport  expand error %s", err.Error())
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
