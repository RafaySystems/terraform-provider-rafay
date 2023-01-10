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
	"github.com/RafaySystems/rafay-common/proto/types/hub/infrapb"
	"github.com/RafaySystems/rctl/pkg/config"
	"github.com/RafaySystems/rctl/pkg/versioninfo"
	"github.com/davecgh/go-spew/spew"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	schema "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAKSClusterV3() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAKSClusterV3Create,
		ReadContext:   resourceAKSClusterV3Read,
		UpdateContext: resourceAKSClusterV3Update,
		DeleteContext: resourceAKSClusterV3Delete,
		Importer: &schema.ResourceImporter{
			State: resourceAKSClusterV3Import,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema:        resource.ClusterSchema.Schema,
	}
}

func resourceAKSClusterV3Create(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// calls upsert
	log.Printf(">>>>>>>>>>>>>> Cluster create starts")
	cluster, err := expandClusterV3(d)
	if err != nil {
		log.Printf(">>>>>>>>>>>>>> ERROR")
	}
	log.Println(">>>>>>>>>>>> CLUSTER", cluster)

	return nil
	return resourceAKSClusterV3Upsert(ctx, d, m)
}

func resourceAKSClusterV3Read(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	return diags
}

func resourceAKSClusterV3Update(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// calls upsert
	return resourceAKSClusterV3Upsert(ctx, d, m)
}

func resourceAKSClusterV3Delete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	return diags
}

func resourceAKSClusterV3Import(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	return []*schema.ResourceData{d}, nil
}

func resourceAKSClusterV3Upsert(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("Cluster upsert starts")
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

	cluster, err := expandClusterV3(d)
	if err != nil {
		log.Printf("Cluster expandCluster error")
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.InfraV3().Cluster().Apply(ctx, cluster, options.ApplyOptions{})
	if err != nil {
		// XXX Debug
		n1 := spew.Sprintf("%+v", cluster)
		log.Println("Cluster apply cluster:", n1)
		log.Printf("Cluster apply error")
		return diag.FromErr(err)
	}

	d.SetId(cluster.Metadata.Name)
	return diags
}

func expandClusterV3(in *schema.ResourceData) (*infrapb.Cluster, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "expand credentials empty input")
	}
	obj := &infrapb.Cluster{}

	if v, ok := in.Get("metadata").([]interface{}); ok && len(v) > 0 {
		obj.Metadata = expandMetaData(v)
	}

	if v, ok := in.Get("spec").([]interface{}); ok && len(v) > 0 {
		objSpec, err := expandClusterV3Spec(v)
		if err != nil {
			return nil, err
		}
		log.Println("expandClusterSpec got spec")
		obj.Spec = objSpec
	}

	obj.ApiVersion = "infra.k8smgmt.io/v3"
	obj.Kind = "Cluster"

	return obj, nil
}

func expandClusterV3Spec(p []interface{}) (*infrapb.ClusterSpec, error) {
	obj := &infrapb.ClusterSpec{}
	if len(p) == 0 || p[0] == nil {
		return obj, fmt.Errorf("%s", "expandClusterSpec empty input")
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["type"].(string); ok && len(v) > 0 {
		obj.Type = v
	}

	if obj.Type != "aks" {
		log.Fatalln("Not Implemented")
	}

	if v, ok := in["sharing"].([]interface{}); ok && len(v) > 0 {
		obj.Sharing = expandSharingSpec(v)
	}

	if v, ok := in["blueprint_config"].([]interface{}); ok && len(v) > 0 {
		obj.BlueprintConfig = expandClusterV3Blueprint(v)
	}

	if v, ok := in["cloud_credentials"].(string); ok && len(v) > 0 {
		obj.CloudCredentials = v
	}

	switch obj.Type {
	case "aks":
		if v, ok := in["config"].([]interface{}); ok && len(v) > 0 {
			obj.Config = expandAKSClusterV3Config(v)
		}
	default:
		log.Fatalln("Not Implemented")
	}

	//if v, ok := in["config"].([]interface{}); ok && len(v) > 0 {
	//	obj.Spec.Config = &infrapb.ClusterSpec_Aks{Aks: &infrapb.AksV3ConfigObject{}}
	//}

	// &infrapb.ClusterSpec_Aks{Aks: &infrapb.AksV3ConfigObject{}}

	// TODO: PROXY CONFIG

	return obj, nil
}

func expandClusterV3Blueprint(p []interface{}) *infrapb.BlueprintConfig {
	obj := infrapb.BlueprintConfig{}
	if len(p) == 0 || p[0] == nil {
		return &obj
	}

	in := p[0].(map[string]interface{})
	if v, ok := in["name"].(string); ok {
		obj.Name = v
	}

	if v, ok := in["version"].(string); ok {
		obj.Version = v
	}

	log.Println("expandClusterV3Blueprint obj", obj)
	return &obj
}

func expandAKSClusterV3Config(p []interface{}) *infrapb.ClusterSpec_Aks {
	obj := &infrapb.ClusterSpec_Aks{Aks: &infrapb.AksV3ConfigObject{}}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["api_version"].(string); ok && len(v) > 0 {
		obj.Aks.ApiVersion = v
	}

	if v, ok := in["kind"].(string); ok && len(v) > 0 {
		obj.Aks.Kind = v
	}

	if v, ok := in["metadata"].([]interface{}); ok && len(v) > 0 {
		obj.Aks.Metadata = expandMetaData(v)
	}

	if v, ok := in["spec"].([]interface{}); ok && len(v) > 0 {
		obj.Aks.Spec = expandAKSClusterV3ConfigSpec(v)
	}

	return obj
}
