package rafay

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"k8s.io/utils/strings/slices"

	"github.com/RafaySystems/rctl/pkg/cluster"
	"github.com/RafaySystems/rctl/pkg/models"
	"github.com/RafaySystems/rctl/pkg/project"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataImportCluster() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataImportClusterRead,
		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"clustername": {
				Type:     schema.TypeString,
				Required: true,
			},
			"projectname": {
				Type:     schema.TypeString,
				Required: true,
			},
			"blueprint": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"blueprint_version": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"operational_status": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"provision_status": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"blueprint_sync": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"location": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"labels": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"kubernetes_provider": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "OTHER",
				ValidateDiagFunc: func(i interface{}, p cty.Path) diag.Diagnostics {
					k8sProvider := i.(string)
					if !slices.Contains(supportedK8sProviderList, k8sProvider) {
						return diag.Errorf("Unsupported K8s Provider.Please refer list of K8s Provider supported: %v", supportedK8sProviderList)
					}
					return diag.Diagnostics{}
				},
			},
			"provision_environment": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "CLOUD",
				ValidateDiagFunc: func(i interface{}, p cty.Path) diag.Diagnostics {
					provisionEnv := i.(string)
					if !slices.Contains(supportedProvisionEnvList, provisionEnv) {
						return diag.Errorf("Unsupported Provision Environment.Please refer list of Environment supported: %v", supportedProvisionEnvList)
					}
					return diag.Diagnostics{}
				},
			},
			"cluster_health": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"nodes": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"created_at": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"modified_at": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"node_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"private_ip": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"num_cores": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"total_memory": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"edge_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"roles": {
							Type:     schema.TypeList,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Optional: true,
						},
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"approved": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"labels": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"tags": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"storage_devices_all": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"ipv4_interfaces_all": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"ipv4_interface": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"additional_storage_devices": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"ips": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
		},
	}
}

func dataImportClusterRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var health string
	resp, err := project.GetProjectByName(d.Get("projectname").(string))
	if err != nil {
		fmt.Print("project name missing in the resource")
		return diags
	}

	project, err := project.NewProjectFromResponse([]byte(resp))
	if err != nil {
		fmt.Printf("project does not exist")
		return diags
	}
	c, err := cluster.GetCluster(d.Get("clustername").(string), project.ID, "")
	if err != nil {
		log.Printf("error in get cluster %s", err.Error())
		if strings.Contains(err.Error(), "not found") {
			log.Println("Resource Read ", "error", err)
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	c, err = cluster.GetClusterWithEdgeID(c.ID, c.ProjectID, "")
	if err != nil {
		log.Printf("error in get cluster %s", err.Error())
		if strings.Contains(err.Error(), "not found") {
			log.Println("Resource Read ", "error", err)
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	if err := d.Set("clustername", c.Name); err != nil {
		log.Printf("get group set name error %s", err.Error())
		return diag.FromErr(err)
	}

	labels, err := getClusterlabels(c.Name, c.ProjectID)
	if err != nil {
		log.Printf("error getting cluster v2 labels: %s", err.Error())
		return diag.FromErr(err)
	}
	if err := d.Set("labels", labels); err != nil {
		log.Printf("set labels error %s", err.Error())
		return diag.FromErr(err)
	}
	if err := d.Set("blueprint", c.ClusterBlueprint); err != nil {
		log.Printf("set blueprint  error %s", err.Error())
		return diag.FromErr(err)
	}

	if err := d.Set("blueprint_version", c.ClusterBlueprintVersion); err != nil {
		log.Printf("set blueprint_version  error %s", err.Error())
		return diag.FromErr(err)
	}

	if err := d.Set("kubernetes_provider", c.ClusterProvisionParams.KubernetesProvider); err != nil {
		log.Printf("set kubernetes_provider  error %s", err.Error())
		return diag.FromErr(err)
	}

	if err := d.Set("provision_environment", c.ClusterProvisionParams.ProvisionEnvironment); err != nil {
		log.Printf("set provision_environment  error %s", err.Error())
		return diag.FromErr(err)
	}
	if err := d.Set("operational_status", c.Status); err != nil {
		log.Printf("set provision_environment  error %s", err.Error())
		return diag.FromErr(err)
	}
	if err := d.Set("provision_status", c.Provision.Status); err != nil {
		log.Printf("set provision_environment  error %s", err.Error())
		return diag.FromErr(err)
	}

	for _, condition := range c.Cluster.ClusterStatus.Conditions {
		if condition.Type == models.ClusterReady {
			log.Println("conditions", string(condition.Status))
			if err := d.Set("blueprint_sync", string(condition.Status)); err != nil {
				log.Printf("set provision_environment  error %s", err.Error())
				return diag.FromErr(err)
			}
			break
		}
	}
	if c.Health == 1 {
		health = "HEALTHY"
	} else if c.Health == 2 {
		health = "UNHEALTHY"
	} else {
		health = "HEALTH UNKNOWN"
	}
	if err := d.Set("cluster_health", health); err != nil {
		log.Printf("set provision_environment  error %s", err.Error())
		return diag.FromErr(err)
	}

	data, _ := json.Marshal(c.Nodes)
	var nodeList []interface{}
	err = json.Unmarshal(data, &nodeList)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("nodes", nodeList); err != nil {
		log.Printf("set provision_environment  error %s", err.Error())
		return diag.FromErr(err)
	}

	d.SetId(c.ID)

	return diags
}
