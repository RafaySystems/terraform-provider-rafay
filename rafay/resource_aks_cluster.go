package rafay

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/RafaySystems/rctl/pkg/cluster"
	"github.com/RafaySystems/rctl/pkg/project"
	"github.com/RafaySystems/rctl/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/go-yaml/yaml"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)


type clusterYamlConfig struct {
	Kind     string `yaml:"kind"`
	Metadata struct {
		Labels  map[string]string `yaml:"labels"`
		Name    string            `yaml:"name"`
		Project string            `yaml:"project"`
	} `yaml:"metadata"`
	Spec struct {
		Type             string `yaml:"type"`
		Blueprint        string `yaml:"blueprint"`
		BlueprintVersion string `yaml:"blueprintversion"`
		Location         string `yaml:"location"`
		//AzureResourceGroup string `yaml:"resourcegroup"`
		//AzureTemplateFile  string `yaml:"templatefile"`
		//AzureTemplateURI   string `yaml:"templateuri"`
		//AzureParameters    string `yaml:"parameters"`
		CloudProvider string `yaml:"cloudprovider"`
	} `yaml:"spec"`
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
			"yamlfilepath": {
				Type:     schema.TypeString,
				Required: true,
			},
			"yamlfileversion": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"projectname": {
				Type:     schema.TypeString,
				Required: true,
			},
			"waitflag": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceAKSClusterCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	log.Printf("create AKS cluster resource")

	YamlConfigFilePath := d.Get("yamlfilepath").(string)

	fileBytes, err := utils.ReadYAMLFileContents(YamlConfigFilePath)
	if err != nil {
		return diag.FromErr(err)
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

	if err = cluster.NewAKSCluster(c.Metadata.Name, c.Spec.Blueprint, c.Spec.BlueprintVersion, c.Spec.CloudProvider, project.ID, string(fileBytes), nil); err != nil {
                return diag.FromErr(err)
        }
	s, errGet := cluster.GetCluster(d.Get("name").(string), project.ID)
	if errGet != nil {
		log.Printf("error while getCluster %s", errGet.Error())
		return diag.FromErr(errGet)
	}

	if d.Get("waitflag").(string) == "1" {
		log.Printf("Cluster Provision may take upto 15-20 Minutes")
		for {
			check, errGet := cluster.GetCluster(d.Get("name").(string), project.ID)
			if errGet != nil {
				log.Printf("error while getCluster %s", errGet.Error())
				return diag.FromErr(errGet)
			}
			if check.Status == "READY" {
				break
			}
			if strings.Contains(check.Provision.Status, "FAILED") {
				return diag.FromErr(fmt.Errorf("Failed to create cluster while cluster provisioning"))
			}
			time.Sleep(40 * time.Second)
		}
	}

	log.Printf("resource aks cluster created %s", s.ID)
	d.SetId(s.ID)

	return diags
}

func resourceAKSClusterRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
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
	if err := d.Set("name", c.Name); err != nil {
		log.Printf("get group set name error %s", err.Error())
		return diag.FromErr(err)
	}

	return diags
}

func resourceAKSClusterUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("update EKS cluster resource")

	resp, err := project.GetProjectByName(d.Get("projectname").(string))

	if err != nil {
		return diag.FromErr(fmt.Errorf("project does not exist"))
	}
	project, err := project.NewProjectFromResponse([]byte(resp))
	if err != nil {
		return diag.FromErr(fmt.Errorf("project does not exist"))
	}

	cluster_resp, err := cluster.GetCluster(d.Get("name").(string), project.ID)
	if err != nil {
		log.Printf("error in get cluster %s", err.Error())
		return diag.FromErr(err)
	}

	YamlConfigFilePath := d.Get("yamlfilepath").(string)

	fileBytes, err := utils.ReadYAMLFileContents(YamlConfigFilePath)
	if err != nil {
		return diag.FromErr(err)
	}

	var c clusterYamlConfig
        if err = yaml.Unmarshal(fileBytes, &c); err != nil {
                return diag.FromErr(err)
        }

	if cluster_resp.ClusterBlueprint != c.Spec.Blueprint || cluster_resp.ClusterBlueprintVersion != c.Spec.BlueprintVersion {
		cluster_resp.ClusterBlueprint = c.Spec.Blueprint

		if c.Spec.BlueprintVersion != "" {
			cluster_resp.ClusterBlueprintVersion = c.Spec.BlueprintVersion
		}

		erru := cluster.UpdateCluster(cluster_resp)
		if erru != nil {
			log.Printf("cluster was not updated, error %s", erru.Error())
			return diag.FromErr(erru)
		}
		errp := cluster.PublishClusterBlueprint(d.Get("name").(string), project.ID)
		if errp != nil {
			log.Printf("cluster was not published, error %s", errp.Error())
			return diag.FromErr(errp)
		}
	}

	return diags
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
