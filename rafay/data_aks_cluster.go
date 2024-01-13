package rafay

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/RafaySystems/rctl/pkg/cluster"
	"github.com/RafaySystems/rctl/pkg/clusterctl"
	"github.com/RafaySystems/rctl/pkg/config"
	glogger "github.com/RafaySystems/rctl/pkg/log"
	"github.com/RafaySystems/rctl/pkg/project"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	// Yaml pkg that have no limit for key length
	"github.com/go-yaml/yaml"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataAKSCluster() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataAKSClusterRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"apiversion": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "apiversion",
			},
			"kind": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "Cluster",
				Description: "kind",
			},
			"metadata": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "AKS specific cluster configuration",
				Elem: &schema.Resource{
					Schema: clusterAKSClusterMetadata(),
				},
			},
			"spec": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "AKS specific cluster configuration",
				Elem: &schema.Resource{
					Schema: clusterAKSClusterSpec(),
				},
			},
		},
	}
}

func dataAKSClusterRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Println("dataAKSClusterRead")
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		_ = context.WithValue(ctx, "debug", "true")
	}

	obj := &AKSCluster{}
	if v, ok := d.Get("apiversion").(string); ok {
		obj.APIVersion = v
	} else {
		fmt.Print("apiversion unable to be found")
		return diag.FromErr(fmt.Errorf("%s", "Apiversion is missing"))
	}

	if v, ok := d.Get("kind").(string); ok {
		obj.Kind = v
	} else {
		fmt.Print("kind unable to be found")
		return diag.FromErr(fmt.Errorf("%s", "Kind is missing"))
	}

	if v, ok := d.Get("metadata").([]interface{}); ok {
		obj.Metadata = expandAKSClusterMetadata(v)
	} else {
		fmt.Print("metadata unable to be found")
		return diag.FromErr(fmt.Errorf("%s", "Metadata is missing"))
	}

	//project details
	resp, err := project.GetProjectByName(obj.Metadata.Project)
	if err != nil {
		fmt.Print("project name missing in the resource")
		return diags
	}

	project, err := project.NewProjectFromResponse([]byte(resp))
	if err != nil {
		fmt.Printf("project does not exist")
		return diags
	}

	c, err := cluster.GetCluster(obj.Metadata.Name, project.ID)
	if err != nil {
		log.Printf("error in get cluster %s", err.Error())
		if strings.Contains(err.Error(), "not found") {
			log.Println("Resource Read ", "error", err)
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	// XXX Debug
	// debugCluster := spew.Sprintf("%+v", c)
	// log.Println("dataAKSClusterRead cluster", debugCluster)

	// another
	logger := glogger.GetLogger()
	rctlCfg := config.GetConfig()
	clusterSpecYaml, err := clusterctl.GetClusterSpec(logger, rctlCfg, c.Name, project.ID)
	if err != nil {
		log.Printf("error in get clusterspec %s", err.Error())
		return diag.FromErr(err)
	}

	cluster, err := cluster.GetCluster(c.Name, project.ID)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("Cluster get cluster 2 worked")

	//log.Printf("Cluster from name: %s", cluster)

	fmt.Println(cluster.ClusterType)

	var respGetCfgFile ResponseGetClusterSpec
	if err := json.Unmarshal([]byte(resp), &respGetCfgFile); err != nil {
		return diag.FromErr(err)
	}

	clusterSpec := AKSCluster{}
	err = yaml.Unmarshal([]byte(clusterSpecYaml), &clusterSpec)
	if err != nil {
		return diag.FromErr(err)
	}
	err = flattenAKSCluster(d, &clusterSpec)
	if err != nil {
		log.Printf("get aks cluster set error %s", err.Error())
		return diag.FromErr(err)
	}

	d.SetId(cluster.ID)
	return diags
}
