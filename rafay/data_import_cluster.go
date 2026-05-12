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
		Description: "Reads an imported cluster's configuration from the Rafay platform.",
		ReadContext: dataImportClusterRead,
		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"clustername": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the cluster to read.",
			},
			"projectname": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the Rafay project containing the cluster.",
			},
			"blueprint": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Name of the blueprint applied to the cluster.",
			},
			"blueprint_version": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Version of the blueprint applied to the cluster.",
			},
			"operational_status": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Current operational status of the cluster.",
			},
			"provision_status": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Current provisioning status of the cluster.",
			},
			"blueprint_sync": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Blueprint sync status of the cluster.",
			},
			"location": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Location or region of the cluster.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the cluster.",
			},
			"labels": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "Key-value labels on the cluster.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"kubernetes_provider": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "OTHER",
				Description: "The Kubernetes provider type. Valid values: AKS, EKS, GKE, OPENSHIFT, OTHER, RKE, EKSANYWHERE.",
				ValidateDiagFunc: func(i interface{}, p cty.Path) diag.Diagnostics {
					k8sProvider := i.(string)
					if !slices.Contains(supportedK8sProviderList, k8sProvider) {
						return diag.Errorf("Unsupported K8s Provider.Please refer list of K8s Provider supported: %v", supportedK8sProviderList)
					}
					return diag.Diagnostics{}
				},
			},
			"provision_environment": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "CLOUD",
				Description: "The provisioning environment. Valid values: CLOUD, ONPREM.",
				ValidateDiagFunc: func(i interface{}, p cty.Path) diag.Diagnostics {
					provisionEnv := i.(string)
					if !slices.Contains(supportedProvisionEnvList, provisionEnv) {
						return diag.Errorf("Unsupported Provision Environment.Please refer list of Environment supported: %v", supportedProvisionEnvList)
					}
					return diag.Diagnostics{}
				},
			},
			"cluster_health": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Overall health status of the cluster.",
			},
			"nodes": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "List of nodes in the cluster.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Name of the node.",
						},
						"created_at": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Timestamp when the node was created.",
						},
						"modified_at": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Timestamp when the node was last modified.",
						},
						"node_id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Unique identifier of the node.",
						},
						"private_ip": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Private IP address of the node.",
						},
						"num_cores": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "Number of CPU cores on the node.",
						},
						"total_memory": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "Total memory (in bytes) on the node.",
						},
						"edge_id": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Edge identifier of the node.",
						},
						"roles": {
							Type:        schema.TypeList,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Optional:    true,
							Description: "List of roles assigned to the node (e.g., master, worker).",
						},
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Internal ID of the node.",
						},
						"approved": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Whether the node has been approved.",
						},
						"status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Current status of the node.",
						},
						"labels": {
							Type:        schema.TypeMap,
							Optional:    true,
							Description: "Key-value labels on the node.",
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
						"tags": {
							Type:        schema.TypeMap,
							Optional:    true,
							Description: "Key-value tags on the node.",
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
						"storage_devices_all": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "List of all storage devices on the node.",
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
						"ipv4_interfaces_all": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "List of all IPv4 network interfaces on the node.",
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
						"ipv4_interface": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "Primary IPv4 network interfaces of the node.",
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
						"additional_storage_devices": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "List of additional storage devices on the node.",
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
						"ips": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "List of IP addresses assigned to the node.",
							Elem:        &schema.Schema{Type: schema.TypeString},
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
