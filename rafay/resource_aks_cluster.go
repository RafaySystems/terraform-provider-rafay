package rafay

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/RafaySystems/rctl/pkg/cluster"
	"github.com/RafaySystems/rctl/pkg/clusterctl"
	"github.com/RafaySystems/rctl/pkg/config"
	glogger "github.com/RafaySystems/rctl/pkg/log"
	"github.com/RafaySystems/rctl/pkg/project"
	"github.com/RafaySystems/rctl/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/go-yaml/yaml"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type Metadata struct {
	Labels  map[string]string `yaml:"labels,omitempty"`
	Name    string            `yaml:"name"`
	Project string            `yaml:"project"`
}

type Spec struct {
	Type             string            `yaml:"type"`
	Blueprint        string            `yaml:"blueprint"`
	BlueprintVersion string            `yaml:"blueprintversion,omitempty"`
	Location         string            `yaml:"location,omitempty"`
	CloudProvider    string            `yaml:"cloudprovider"`
	ClusterConfig    *AKSClusterConfig `yaml:"clusterConfig"`
}

type clusterYamlConfig struct {
	APIVersion string    `yaml:"apiversion"`
	Kind       string    `yaml:"kind"`
	Metadata   *Metadata `yaml:"metadata"`
	Spec       *Spec     `yaml:"spec"`
}

type clusterCTLResponse struct {
	TaskSetID string `json:"taskset_id,omitempty"`
	Status    string `json:"status,omitempty"`
}

func clusterAKSConfigNodePoolsFields() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The AKS node group name",
		},
		"location": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The AKS node pool locations",
		},
		"count": {
			Type:        schema.TypeInt,
			Optional:    true,
			Default:     1,
			Description: "The AKS node pool count",
		},
		"enable_autoscaling": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Is AKS node pool auto scaling enabled?",
		},
		"max_count": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "The AKS node pool max count",
		},
		"max_pods": {
			Type:        schema.TypeInt,
			Optional:    true,
			Default:     110,
			Description: "The AKS node pool max pods",
		},
		"min_count": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "The AKS node pool min count",
		},
		"mode": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "System",
			Description: "The AKS node pool mode",
		},
		"orchestrator_version": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The AKS node pool orchestrator version",
		},
		"os_type": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "Linux",
			Description: "Enable AKS node pool os type",
		},
		"vm_size": {
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			Description: "The AKS node pool vm size",
		},
		"type": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "Microsoft.ContainerService/managedClusters/agentPools",
			Description: "The AKS node pool type",
		},
		"node_labels": {
			Type:        schema.TypeMap,
			Optional:    true,
			Computed:    true,
			Description: "The AKS node pool labels",
		},
		"property_type": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "VirtualMachineScaleSets",
			Description: "The AKS node pool type",
		},
		"apiversion": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "2021-05-01",
			Description: "API Version",
		},
		"availability_zones": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "The AKS node pool availability zones",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
	}

	return s
}

func clusterAKSConfigFields() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"identity_type": {
			Type:     schema.TypeString,
			Optional: true,
			Default:  "SystemAssigned",
		},
		"enable_private_cluster": {
			Type:     schema.TypeBool,
			Optional: true,
		},
		"location": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"dnsprefix": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"kubernetesversion": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"loadbalancer_sku": {
			Type:     schema.TypeString,
			Optional: true,
			Default:  "standard",
		},
		"network_plugin": {
			Type:     schema.TypeString,
			Optional: true,
			Default:  "kubenet",
		},
		"network_policy": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"sku_name": {
			Type:     schema.TypeString,
			Optional: true,
			Default:  "Basic",
		},
		"sku_tier": {
			Type:     schema.TypeString,
			Optional: true,
			Default:  "Free",
		},
		"tags": {
			Type:        schema.TypeMap,
			Optional:    true,
			Computed:    true,
			Description: "The AKS cluster tags",
		},
		"type": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "Microsoft.ContainerService/managedClusters",
			Description: "Type",
		},
		"apiversion": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "2021-05-01",
			Description: "API Version",
		},
		"resource_group_name": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"node_pools": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "The AKS node pools to use",
			Elem: &schema.Resource{
				Schema: clusterAKSConfigNodePoolsFields(),
			},
		},
	}
	return s
}

func resourceAKSCluster() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAKSClusterCreate,
		ReadContext:   resourceAKSClusterRead,
		UpdateContext: resourceAKSClusterUpdate,
		DeleteContext: resourceAKSClusterDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
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
			"blueprintversion": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"cloudprovider": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"cluster_config": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "The AKS cluster config to use",
				Elem: &schema.Resource{
					Schema: clusterAKSConfigFields(),
				},
			},
		},
	}
}

func aksClusterCTL(config *config.Config, rafayConfigs, clusterConfigs [][]byte, dryRun bool) (string, error) {
	logger := glogger.GetLogger()
	configMap, errs := collateConfigsByName(rafayConfigs, clusterConfigs)
	if len(errs) == 0 && len(configMap) > 0 {
		// Make request
		for clusterName, configBytes := range configMap {
			return clusterctl.Apply(logger, config, clusterName, configBytes, dryRun)
		}
	}
	return "", fmt.Errorf("%s", "config collate error")
}

func aksClusterCTLStatus(taskid string) (string, error) {
	logger := glogger.GetLogger()
	rctlCfg := config.GetConfig()
	return clusterctl.Status(logger, rctlCfg, taskid)
}

func expandClusterAKSNodePools(p []interface{}) []AKSNodePool {
	if len(p) == 0 {
		return []AKSNodePool{}
	}
	out := make([]AKSNodePool, len(p))
	for i := range p {
		in := p[i].(map[string]interface{})
		obj := AKSNodePool{
			Properties: &AKSNodePoolProperties{},
		}

		if v, ok := in["name"].(string); ok {
			obj.Name = v
		}
		if v, ok := in["location"].(string); ok {
			obj.Location = v
		}
		if v, ok := in["apiversion"].(string); ok {
			obj.APIVersion = v
		}
		if v, ok := in["type"].(string); ok {
			obj.Type = v
		}
		if v, ok := in["count"].(int); ok {
			obj.Properties.Count = &v
		}
		if v, ok := in["enable_autoscaling"].(bool); ok {
			obj.Properties.EnableAutoScaling = &v
		}
		if v, ok := in["max_count"].(int); ok {
			obj.Properties.MaxCount = &v
		}
		if v, ok := in["max_pods"].(int); ok {
			obj.Properties.MaxPods = &v
		}
		if v, ok := in["min_count"].(int); ok {
			obj.Properties.MinCount = &v
		}
		if v, ok := in["mode"].(string); ok {
			obj.Properties.Mode = v
		}
		if v, ok := in["orchestrator_version"].(string); ok {
			obj.Properties.OrchestratorVersion = v
		}
		if v, ok := in["os_type"].(string); ok {
			obj.Properties.OSType = v
		}
		if v, ok := in["vm_size"].(string); ok {
			obj.Properties.VMSize = v
		}
		if v, ok := in["node_labels"].(map[string]interface{}); ok && len(v) > 0 {
			obj.Properties.NodeLabels = toMapString(v)
		}
		if v, ok := in["property_type"].(string); ok {
			obj.Properties.Type = v
		}
		if v, ok := in["availability_zones"].([]interface{}); ok {
			availabilityZones := toArrayString(v)
			obj.Properties.AvailabilityZones = availabilityZones
		}
		out[i] = obj
	}
	return out
}

func expandClusterAKSConfig(p []interface{}) *AKSClusterConfig {
	var tags map[string]string
	obj := &AKSClusterConfig{}
	objSpec := &AKSClusterConfigSpec{}

	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	identity_type := in["identity_type"].(string)
	enablePrivateCluster := in["enable_private_cluster"].(bool)
	location := in["location"].(string)
	dnsPrefix := in["dnsprefix"].(string)
	kubernetesVersion := in["kubernetesversion"].(string)
	loadBalancerSku := in["loadbalancer_sku"].(string)
	networkPlugin := in["network_plugin"].(string)
	networkPolicy := in["network_policy"].(string)
	sku_name := in["sku_name"].(string)
	sku_tier := in["sku_tier"].(string)

	if v, ok := in["tags"].(map[string]interface{}); ok && len(v) > 0 {
		tags = toMapString(v)
	}

	aksClustertype := in["type"].(string)
	resourceGroupName := in["resource_group_name"].(string)
	apiVersion := in["apiversion"].(string)

	aksManagedClusterIdentity := AKSManagedClusterIdentity{
		Type: identity_type,
	}
	aksManagedClusterAPIServerAccessProfile := AKSManagedClusterAPIServerAccessProfile{
		EnablePrivateCluster: &enablePrivateCluster,
	}
	aksManagedClusterNetworkProfile := AKSManagedClusterNetworkProfile{
		LoadBalancerSKU: loadBalancerSku,
		NetworkPlugin:   networkPlugin,
		NetworkPolicy:   networkPolicy,
	}
	aksManagedClusterProperties := AKSManagedClusterProperties{
		KubernetesVersion:      kubernetesVersion,
		APIServerAccessProfile: &aksManagedClusterAPIServerAccessProfile,
		DNSPrefix:              dnsPrefix,
		NetworkProfile:         &aksManagedClusterNetworkProfile,
	}
	aksManagedClusterSKU := AKSManagedClusterSKU{
		Name: sku_name,
		Tier: sku_tier,
	}
	aksManagedCluster := AKSManagedCluster{
		Type:       aksClustertype,
		APIVersion: apiVersion,
		Location:   location,
		Identity:   &aksManagedClusterIdentity,
		Properties: &aksManagedClusterProperties,
		SKU:        &aksManagedClusterSKU,
		Tags:       tags,
	}

	objSpec.ResourceGroupName = resourceGroupName
	objSpec.ManagedCluster = &aksManagedCluster

	if v, ok := in["node_pools"].([]interface{}); ok && len(v) > 0 {
		nodePools := expandClusterAKSNodePools(v)
		objSpec.NodePools = &nodePools
	}
	obj.Spec = objSpec
	return obj
}

func processInputs(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	projectName := d.Get("projectname").(string)
	resp, err := project.GetProjectByName(projectName)
	if err != nil {
		fmt.Print("project name missing in the resource")
		return diag.FromErr(fmt.Errorf("%s", "Project name missing in the resource"))
	}

	_, err = project.NewProjectFromResponse([]byte(resp))
	if err != nil {
		fmt.Printf("project does not exist")
		return diag.FromErr(fmt.Errorf("%s", "Project does not exist"))
	}

	clusterName := d.Get("name").(string)
	metadata := AKSClusterConfigMetadata{
		Name: clusterName,
	}

	blueprint := d.Get("blueprint").(string)
	blueprintversion := d.Get("blueprintversion").(string)
	cloudprovider := d.Get("cloudprovider").(string)

	clusterConfig := expandClusterAKSConfig(d.Get("cluster_config").([]interface{}))
	clusterConfig.Metadata = &metadata
	clusterConfig.APIVersion = "rafay.io/v1alpha1"
	clusterConfig.Kind = "aksClusterConfig"

	yamlConfig := clusterYamlConfig{
		APIVersion: "rafay.io/v1alpha1",
		Kind:       "Cluster",
	}
	yamlConfig.Metadata = &Metadata{}
	yamlConfig.Spec = &Spec{}
	yamlConfig.Metadata.Name = clusterName
	yamlConfig.Metadata.Project = projectName
	yamlConfig.Spec.Blueprint = blueprint
	yamlConfig.Spec.BlueprintVersion = blueprintversion
	yamlConfig.Spec.CloudProvider = cloudprovider
	yamlConfig.Spec.ClusterConfig = clusterConfig
	yamlConfig.Spec.Type = "aks"
	//log.Printf("AKS Cluster yamlConfig %v", yamlConfig)

	out, err := yaml.Marshal(yamlConfig)
	if err != nil {
		return diag.FromErr(err)
	}
	//log.Printf("AKS Cluster YAML SPEC \n---\n%s\n----\n", out)
	return process_filebytes(ctx, d, m, out)
}

func resourceAKSClusterUpsert(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("resourceAKSClusterUpsert")

	return processInputs(ctx, d, m)

}

func process_filebytes(ctx context.Context, d *schema.ResourceData, m interface{}, fileBytes []byte) diag.Diagnostics {
	var diags diag.Diagnostics
	rctlCfg := config.GetConfig()
	// split the file and update individual resources
	cfgList, err := utils.SplitYamlAndGetListByKind(fileBytes)
	if err != nil {
		return diag.FromErr(err)
	}

	if len(cfgList) < 1 {
		fmt.Printf("no cluster in the config")
		return diags

	}

	if len(cfgList) > 1 {
		fmt.Printf("found more than one cluster config in the cluster config")
		return diags
	}

	var c clusterYamlConfig
	if err = yaml.Unmarshal(fileBytes, &c); err != nil {
		return diag.FromErr(err)
	}

	if c.Spec.Type != "aks" {
		fmt.Printf("cluster types is not aks, type is %s", c.Spec.Type)
		return diags
	}

	if c.Metadata.Name == "" {
		return diag.FromErr(fmt.Errorf("cluster name is not provided in yaml file"))
	}

	if c.Metadata.Name != d.Get("name").(string) {
		return diag.FromErr(fmt.Errorf("%s", "ClusterConfig name does not match config file"))
	}

	if c.Metadata.Project != d.Get("projectname").(string) {
		return diag.FromErr(fmt.Errorf("%s", "ClusterConfig projectname does not match config file"))
	}
	// get project details
	resp, err := project.GetProjectByName(d.Get("projectname").(string))
	if err != nil {
		fmt.Printf("project does not exist")
		return diags
	}
	project, err := project.NewProjectFromResponse([]byte(resp))
	if err != nil {
		fmt.Printf("project does not exist")
		return diags
	}

	// cluster
	response, err := aksClusterCTL(rctlCfg, cfgList["Cluster"], cfgList["ClusterConfig"], false)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("process_filebytes response : %s", response)
	res := clusterCTLResponse{}
	err = json.Unmarshal([]byte(response), &res)
	if err != nil {
		log.Println("response parse error", err)
		return diag.FromErr(err)
	}

	time.Sleep(10 * time.Second)
	s, errGet := cluster.GetCluster(d.Get("name").(string), project.ID)
	if errGet != nil {
		log.Printf("error while getCluster %s", errGet.Error())
		return diag.FromErr(errGet)
	}

	log.Printf("Cluster Provision may take upto 15-20 Minutes")
	for {
		time.Sleep(60 * time.Second)
		statusResp, err := aksClusterCTLStatus(res.TaskSetID)
		if err != nil {
			log.Println("status response parse error", err)
			return diag.FromErr(err)
		}
		log.Println("statusResp ", statusResp)
		sres := clusterCTLResponse{}
		err = json.Unmarshal([]byte(statusResp), &sres)
		if err != nil {
			log.Println("status response unmarshal error", err)
			return diag.FromErr(err)
		}
		if strings.Contains(sres.Status, "STATUS_COMPLETE") {
			break
		}
		if strings.Contains(sres.Status, "STATUS_FAILED") {
			return diag.FromErr(fmt.Errorf("failed to create/update cluster while provisioning cluster %s", d.Get("name").(string)))
		}
	}
	log.Printf("resource aks cluster created/updated %s", s.ID)
	d.SetId(s.ID)

	return diags
}

func resourceAKSClusterCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("create AKS cluster resource")
	return resourceAKSClusterUpsert(ctx, d, m)
}

func resourceAKSClusterRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	log.Println("resourceAKSClusterRead")
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
	c, err := cluster.GetCluster(d.Get("name").(string), project.ID)
	if err != nil {
		log.Printf("error in get cluster %s", err.Error())
		return diag.FromErr(err)
	}

	logger := glogger.GetLogger()
	rctlCfg := config.GetConfig()
	clusterSpecYaml, err := clusterctl.GetClusterSpec(logger, rctlCfg, c.Name, project.ID)
	if err != nil {
		log.Printf("error in get clusterspec %s", err.Error())
		return diag.FromErr(err)
	}
	log.Println("resourceAKSClusterRead clusterSpec ", clusterSpecYaml)

	clusterSpec := clusterYamlConfig{}
	err = yaml.Unmarshal([]byte(clusterSpecYaml), &clusterSpec)
	if err != nil {
		return diag.FromErr(err)
	}
	err = flattenAKSCluster(d, &clusterSpec)
	if err != nil {
		log.Printf("get aks cluster set error %s", err.Error())
		return diag.FromErr(err)
	}

	return diags
}

func resourceAKSClusterUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("update AKS cluster resource")

	resp, err := project.GetProjectByName(d.Get("projectname").(string))

	if err != nil {
		return diag.FromErr(fmt.Errorf("project does not exist"))
	}
	project, err := project.NewProjectFromResponse([]byte(resp))
	if err != nil {
		return diag.FromErr(fmt.Errorf("project does not exist"))
	}

	_, err = cluster.GetCluster(d.Get("name").(string), project.ID)
	if err != nil {
		log.Printf("error in get cluster %s", err.Error())
		return diag.FromErr(err)
	}

	return resourceAKSClusterUpsert(ctx, d, m)
}

func resourceAKSClusterDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("resource cluster delete id %s", d.Id())

	resp, err := project.GetProjectByName(d.Get("projectname").(string))
	if err != nil {
		fmt.Print("project  does not exist")
		return diags
	}

	project, err := project.NewProjectFromResponse([]byte(resp))
	if err != nil {
		fmt.Printf("project  does not exist")
		return diags
	}

	errDel := cluster.DeleteCluster(d.Get("name").(string), project.ID)
	if errDel != nil {
		log.Printf("delete cluster error %s", errDel.Error())
		return diag.FromErr(errDel)
	}

	return diags
}

// Flatteners
func flattenClusterAKSConfigNodePools(input []AKSNodePool, p []interface{}) []interface{} {
	log.Println("flattenClusterAKSConfigNodePools")
	if input == nil {
		return nil
	}
	out := make([]interface{}, len(input))
	for i, in := range input {
		//log.Println("flattenClusterAKSConfigNodePools in ", in)

		obj := map[string]interface{}{}
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}

		if len(in.Name) > 0 {
			obj["name"] = in.Name
		}
		if len(in.Type) > 0 {
			obj["type"] = in.Type
		}
		if len(in.Location) > 0 {
			obj["location"] = in.Location
		}
		if len(in.APIVersion) > 0 {
			obj["apiversion"] = in.APIVersion
		}
		if in.Properties != nil {

			if len(in.Properties.OrchestratorVersion) > 0 {
				obj["orchestrator_version"] = in.Properties.OrchestratorVersion
			}
			if len(in.Properties.OSType) > 0 {
				obj["os_type"] = in.Properties.OSType
			}
			if len(in.Properties.Mode) > 0 {
				obj["mode"] = in.Properties.Mode
			}
			if len(in.Properties.VMSize) > 0 {
				obj["vm_size"] = in.Properties.VMSize
			}
			if in.Properties.Count != nil {
				obj["count"] = *in.Properties.Count
			}
			if in.Properties.EnableAutoScaling != nil {
				obj["enable_autoscaling"] = *in.Properties.EnableAutoScaling
			}
			if in.Properties.MaxCount != nil {
				obj["max_count"] = *in.Properties.MaxCount
			}
			if in.Properties.MaxPods != nil {
				obj["max_pods"] = *in.Properties.MaxPods
			}
			if in.Properties.MinCount != nil {
				obj["min_count"] = *in.Properties.MinCount
			}

			if in.Properties.NodeLabels != nil && len(in.Properties.NodeLabels) > 0 {
				obj["node_labels"] = toMapInterface(in.Properties.NodeLabels)
			}

			if in.Properties.AvailabilityZones != nil && len(in.Properties.AvailabilityZones) > 0 {
				obj["availability_zones"] = toArrayInterface(in.Properties.AvailabilityZones)
			}

		}
		//log.Println("flattenClusterAKSConfigNodePools obj ", obj)
		out[i] = obj
	}

	//log.Println("flattenClusterAKSConfigNodePools out ", out)
	return out
}

func flattenAKSClusterConfigSpec(in *AKSClusterConfigSpec, p []interface{}) interface{} {
	if in == nil {
		return nil
	}

	log.Println("flattenAKSClusterConfigSpec ", in)
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.ResourceGroupName) > 0 {
		obj["resource_group_name"] = in.ResourceGroupName
	}
	if in.ManagedCluster != nil {
		if len(in.ManagedCluster.Location) > 0 {
			obj["location"] = in.ManagedCluster.Location
		}
		if in.ManagedCluster.Identity != nil && len(in.ManagedCluster.Identity.Type) > 0 {
			obj["identity_type"] = in.ManagedCluster.Identity.Type
		}
		if in.ManagedCluster.Properties != nil {
			if in.ManagedCluster.Properties.APIServerAccessProfile != nil {
				obj["enable_private_cluster"] = *in.ManagedCluster.Properties.APIServerAccessProfile.EnablePrivateCluster
			}

			if len(in.ManagedCluster.Properties.DNSPrefix) > 0 {
				obj["dnsprefix"] = in.ManagedCluster.Properties.DNSPrefix
			}

			if len(in.ManagedCluster.Properties.KubernetesVersion) > 0 {
				obj["kubernetesversion"] = in.ManagedCluster.Properties.KubernetesVersion
			}

			if in.ManagedCluster.Properties.NetworkProfile != nil {
				if len(in.ManagedCluster.Properties.NetworkProfile.NetworkPlugin) > 0 {
					obj["network_plugin"] = in.ManagedCluster.Properties.NetworkProfile.NetworkPlugin
				}
				if len(in.ManagedCluster.Properties.NetworkProfile.LoadBalancerSKU) > 0 {
					obj["loadbalancer_sku"] = in.ManagedCluster.Properties.NetworkProfile.LoadBalancerSKU
				}
				if len(in.ManagedCluster.Properties.NetworkProfile.NetworkPolicy) > 0 {
					obj["network_policy"] = in.ManagedCluster.Properties.NetworkProfile.NetworkPolicy
				}
			}

			if in.ManagedCluster.SKU != nil {
				if len(in.ManagedCluster.SKU.Name) > 0 {
					obj["sku_name"] = in.ManagedCluster.SKU.Name
				}
				if len(in.ManagedCluster.SKU.Name) > 0 {
					obj["sku_tier"] = in.ManagedCluster.SKU.Tier
				}
			}

			if in.ManagedCluster.Tags != nil && len(in.ManagedCluster.Tags) > 0 {
				obj["tags"] = toMapInterface(in.ManagedCluster.Tags)
			}
		}
	}
	//log.Println("flattenAKSClusterConfigSpec obj", obj)

	if in.NodePools != nil && len(*in.NodePools) > 0 {
		v, ok := obj["node_pools"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["node_pools"] = flattenClusterAKSConfigNodePools(*in.NodePools, v)
	}

	return []interface{}{obj}
}

func flattenAKSCluster(d *schema.ResourceData, in *clusterYamlConfig) error {
	var err error
	if in == nil {
		return nil
	}

	log.Println("flattenAKSCluster ", in)

	if in.Metadata != nil {
		if len(in.Metadata.Name) > 0 {
			err = d.Set("name", in.Metadata.Name)
			if err != nil {
				return err
			}
		}
		if len(in.Metadata.Project) > 0 {
			err = d.Set("projectname", in.Metadata.Project)
			if err != nil {
				return err
			}
		}
	}

	if in.Spec != nil {
		if len(in.Spec.Blueprint) > 0 {
			err = d.Set("blueprint", in.Spec.Blueprint)
			if err != nil {
				return err
			}
		}
		if len(in.Spec.BlueprintVersion) > 0 {
			err = d.Set("blueprintversion", in.Spec.BlueprintVersion)
			if err != nil {
				return err
			}
		}
		if len(in.Spec.Location) > 0 {
			err = d.Set("location", in.Spec.Location)
			if err != nil {
				return err
			}
		}
		if len(in.Spec.CloudProvider) > 0 {
			err = d.Set("cloudprovider", in.Spec.CloudProvider)
			if err != nil {
				return err
			}
		}
		if in.Spec.ClusterConfig != nil &&
			in.Spec.ClusterConfig.Spec != nil {
			v, ok := d.Get("cluster_config").([]interface{})
			if !ok {
				v = []interface{}{}
			}
			//log.Println("flattenAKSCluster cluster_config", v)
			aksConfig := flattenAKSClusterConfigSpec(in.Spec.ClusterConfig.Spec, v)
			//log.Println("flattenAKSCluster aksConfig", aksConfig)
			err = d.Set("cluster_config", aksConfig)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
