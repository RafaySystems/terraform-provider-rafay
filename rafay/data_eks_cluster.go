package rafay

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/RafaySystems/rctl/pkg/cluster"
	"github.com/RafaySystems/rctl/pkg/clusterctl"
	"github.com/RafaySystems/rctl/pkg/config"
	glogger "github.com/RafaySystems/rctl/pkg/log"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/go-yaml/yaml"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// schema input for cluster file
func dataClusterMetadataField() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"kind": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "Cluster",
			Description: "The type of resource. Supported value is `Cluster`.",
		},
		"metadata": {
			Type:        schema.TypeList,
			Required:    true,
			Description: "Contains data that helps uniquely identify the resource.",
			Elem: &schema.Resource{
				Schema: clusterMetaMetadataFields(),
			},
			MinItems: 1,
			MaxItems: 1,
		},
		"spec": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "The specification associated with the cluster, including cluster networking options.",
			Elem: &schema.Resource{
				Schema: specField(),
			},
			MinItems: 1,
			MaxItems: 1,
		},
	}
	return s
}

func dataEKSCluster() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataEKSClusterRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"cluster": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Rafay specific cluster configuration",
				Elem: &schema.Resource{
					Schema: dataClusterMetadataField(),
				},
				MinItems: 1,
				MaxItems: 1,
			},
			"cluster_config": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "EKS specific cluster configuration",
				Elem: &schema.Resource{
					Schema: configField(),
				},
				MinItems: 1,
				MaxItems: 1,
			},
		},
		Description: resourceEKSClusterDescription,
	}
}

func dataEKSClusterRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Println("READ eks cluster")
	rawState := d.GetRawState()
	var diags diag.Diagnostics
	// find cluster name and project name
	clusterName, ok := d.Get("cluster.0.metadata.0.name").(string)
	if !ok || clusterName == "" {
		log.Print("Cluster name unable to be found")
		return diag.FromErr(fmt.Errorf("%s", "cluster name is missing"))
	}
	projectName, ok := d.Get("cluster.0.metadata.0.project").(string)
	if !ok || projectName == "" {
		log.Print("Cluster project name unable to be found")
		return diag.FromErr(fmt.Errorf("%s", "project name is missing"))
	}
	projectID, err := getProjectIDFromName(projectName)
	if err != nil {
		log.Print("error converting project name to id")
		return diag.Errorf("error converting project name to project ID")
	}
	c, err := cluster.GetCluster(clusterName, projectID, uaDef)
	if err != nil {
		log.Printf("error in get cluster %s", err.Error())
		if strings.Contains(err.Error(), "not found") {
			log.Println("Resource Read ", "error", err)
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}
	log.Println("got cluster from backend")
	logger := glogger.GetLogger()
	rctlCfg := config.GetConfig()
	clusterSpecYaml, err := clusterctl.GetClusterSpec(logger, rctlCfg, c.Name, projectID, uaDef)
	if err != nil {
		log.Printf("error in get clusterspec %s", err.Error())
		return diag.FromErr(err)
	}
	log.Println("resourceEKSClusterRead clusterSpec ", clusterSpecYaml)

	decoder := yaml.NewDecoder(bytes.NewReader([]byte(clusterSpecYaml)))

	clusterSpec := EKSCluster{}
	if err := decoder.Decode(&clusterSpec); err != nil {
		log.Println("error decoding cluster spec")
		return diag.FromErr(err)
	}

	clusterConfigSpec := EKSClusterConfig{}
	if err := decoder.Decode(&clusterConfigSpec); err != nil {
		log.Println("error decoding cluster config spec")
		return diag.FromErr(err)
	}

	v, ok := d.Get("cluster").([]interface{})
	if !ok {
		v = []interface{}{}
	}
	var clusterState cty.Value
	if !rawState.IsNull() {
		clusterState = rawState.GetAttr("cluster")
	}
	c1, err := flattenEKSCluster(&clusterSpec, v, clusterState)
	log.Println("finished flatten eks cluster", c1)
	if err != nil {
		log.Printf("flatten eks cluster error %s", err.Error())
		return diag.FromErr(err)
	}
	err = d.Set("cluster", c1)
	if err != nil {
		log.Printf("err setting cluster %s", err.Error())
		return diag.FromErr(err)
	}

	v2, ok := d.Get("cluster_config").([]interface{})
	if !ok {
		v2 = []interface{}{}
	}
	var clusterConfigState cty.Value
	if !rawState.IsNull() {
		clusterConfigState = rawState.GetAttr("cluster")
	}
	c2, err := flattenEKSClusterConfig(&clusterConfigSpec, clusterConfigState, v2)
	if err != nil {
		log.Printf("flatten eks cluster config error %s", err.Error())
		return diag.FromErr(err)
	}
	err = d.Set("cluster_config", c2)
	if err != nil {
		log.Printf("err setting cluster config %s", err.Error())
		return diag.FromErr(err)
	}
	log.Println("flattened cluster fine")
	log.Println("finished read")

	d.SetId(c.ID)
	return diags
}
