package rafay

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
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

func resourceGKEClusterV3() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceGKEClusterV3Create,
		ReadContext:   resourceGKEClusterV3Read,
		UpdateContext: resourceGKEClusterV3Update,
		DeleteContext: resourceGKEClusterV3Delete,
		//	Importer: &schema.ResourceImporter{
		//		State: resourceAKSClusterV3Import,
		//	},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(90 * time.Minute),
			Update: schema.DefaultTimeout(90 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema:        resource.ClusterSchema.Schema,
	}
}

func resourceGKEClusterV3Upsert(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("GKE Cluster upsert starts")
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	if d.State() != nil && d.State().ID != "" {
		n := GetMetaName(d)
		if n != "" && n != d.State().ID {
			log.Printf("cluster name change not supported")
			d.State().Tainted = true
			return diag.FromErr(fmt.Errorf("%s", "cluster name change not supported"))
		}
	}

	//cluster, err := expandClusterV3(d)
	cluster, err := expandGKEClusterToV3(d)
	if err != nil {
		log.Printf("Cluster expandCluster error " + err.Error())
		return diag.FromErr(err)
	}

	log.Println(">>>>>> CLUSTER: ", cluster)

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

	// // wait for cluster creation
	// ticker := time.NewTicker(time.Duration(60) * time.Second)
	// defer ticker.Stop()
	// timeout := time.After(time.Duration(90) * time.Minute)

	// edgeName := cluster.Metadata.Name
	// projectName := cluster.Metadata.Project
	// d.SetId(cluster.Metadata.Name)

	return diags
}

func resourceGKEClusterV3Create(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("GKE Cluster create starts")

	diags := resourceGKEClusterV3Upsert(ctx, d, m)
	return diags

}

func resourceGKEClusterV3Read(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	return diags

}

func resourceGKEClusterV3Update(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("GKE Cluster update starts")

	var diags diag.Diagnostics

	diags = resourceGKEClusterV3Upsert(ctx, d, m)
	return diags
}

func resourceGKEClusterV3Delete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("GKE Cluster delete starts")

	return diags
}

func expandGKEClusterToV3(in *schema.ResourceData) (*infrapb.Cluster, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "expand cluster invoked with empty input")
	}
	obj := &infrapb.Cluster{}

	obj.ApiVersion = V3_CLUSTER_APIVERSION
	obj.Kind = V3_CLUSTER_KIND

	if v, ok := in.Get("metadata").([]interface{}); ok && len(v) > 0 {
		obj.Metadata = expandMetaData(v)
	}

	// spec
	if v, ok := in.Get("spec").([]interface{}); ok && len(v) > 0 {
		objSpec, err := expandGKEClusterToV3Spec(v)
		if err != nil {
			return nil, err
		}
		log.Println("expandClusterSpec got spec")
		obj.Spec = objSpec
	}

	return obj, nil
}

func expandGKEClusterToV3Spec(p []interface{}) (*infrapb.ClusterSpec, error) {
	// expandGKESpec??
	/*
		type
		sharing
		cloudCredentials
		blueprint
		proxy
		config --- gke

	*/

	obj := &infrapb.ClusterSpec{}
	if len(p) == 0 || p[0] == nil {
		return obj, fmt.Errorf("%s", "expandClusterSpec empty input")
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["sharing"].([]interface{}); ok && len(v) > 0 {
		obj.Sharing = expandSharingSpecV3(v)
	}

	if v, ok := in["blueprint"].([]interface{}); ok && len(v) > 0 {
		var err error
		obj.Blueprint, err = expandGKEClusterToV3Blueprint(v)
		if err != nil {
			return nil, fmt.Errorf("failed to expand blueprint " + err.Error())
		}
	}

	if v, ok := in["cloud_credentials"].(string); ok && len(v) > 0 {
		obj.CloudCredentials = v
	}

	// TODO: Proxy

	if v, ok := in["type"].(string); ok && len(v) > 0 {
		obj.Type = v
	}

	if !strings.EqualFold(obj.Type, GKE_CLUSTER_TYPE) {
		return nil, errors.New("cluster type not implemented")
	}

	if strings.EqualFold(obj.Type, GKE_CLUSTER_TYPE) {
		if v, ok := in["config"].([]interface{}); ok && len(v) > 0 {
			var err error
			obj.Config, err = expandToV3GkeConfigObject(v)
			if err != nil {
				return nil, fmt.Errorf("failed to expand to gke config " + err.Error())
			}
		}
	}

	return obj, nil
}

func expandGKEClusterToV3Blueprint(p []interface{}) (*infrapb.ClusterBlueprint, error) {
	obj := &infrapb.ClusterBlueprint{}
	if len(p) == 0 || p[0] == nil {
		return obj, errors.New("empty blueprint in input")
	}

	in := p[0].(map[string]interface{})
	if v, ok := in["name"].(string); ok {
		obj.Name = v
	} else if !ok {
		return nil, errors.New("missing blueprint name")
	}

	if v, ok := in["version"].(string); ok {
		obj.Version = v
	} else if !ok {
		return nil, errors.New("missing blueprint version")
	}

	log.Println("expandGKEClusterToV3Blueprint obj", obj)
	return obj, nil

}
