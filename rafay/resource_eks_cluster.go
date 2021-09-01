package rafay

import (
	"context"
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

type configMetadata struct {
	Name    string `yaml:"name"`
	Project string `yaml:"project"`
	Version string `yaml:"version"`
}

type configResourceType struct {
	Meta *configMetadata `yaml:"metadata"`
}

type blueprintSpec struct {
	Blueprint        string `yaml:"blueprint"`
	Blueprintversion string `yaml:"blueprintversion"`
}

type blueprintType struct {
	Spec *blueprintSpec `yaml:"spec"`
}

func resourceEKSCluster() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceEKSClusterCreate,
		ReadContext:   resourceEKSClusterRead,
		UpdateContext: resourceEKSClusterUpdate,
		DeleteContext: resourceEKSClusterDelete,

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

func findResourceNameFromConfig(configBytes []byte) (string, string, string, error) {
	var config configResourceType
	if err := yaml.Unmarshal(configBytes, &config); err != nil {
		return "", "", "", nil
	} else if config.Meta == nil {
		return "", "", "", fmt.Errorf("%s", "Invalid resource: No metadata found")
	} else if config.Meta.Name == "" {
		return "", "", "", fmt.Errorf("%s", "Invalid resource: No name specified in metadata")
	}
	return config.Meta.Name, config.Meta.Project, config.Meta.Version, nil
}

func findBlueprintName(configBytes []byte) (string, string, error) {
	var blueprint blueprintType
	if err := yaml.Unmarshal(configBytes, &blueprint); err != nil {
		return "", "", nil
	} else if blueprint.Spec == nil {
		return "", "", fmt.Errorf("%s", "Invalid resource: No spec found")
	} else if blueprint.Spec.Blueprint == "" {
		return "", "", fmt.Errorf("%s", "Invalid resource: No name specified in spec")
	}

	return blueprint.Spec.Blueprint, blueprint.Spec.Blueprintversion, nil

}
func collateConfigsByName(rafayConfigs, clusterConfigs [][]byte) (map[string][]byte, []error) {
	var errs []error
	configsMap := make(map[string][][]byte)
	// First find all rafay spec configurations
	for _, config := range rafayConfigs {
		name, _, _, err := findResourceNameFromConfig(config)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		if _, ok := configsMap[name]; ok {
			errs = append(errs, fmt.Errorf(`duplicate "cluster" resource with name "%s" found`, name))
			continue
		}
		configsMap[name] = append(configsMap[name], config)
	}
	// Then append the cluster specific configurations
	for _, config := range clusterConfigs {
		name, _, _, err := findResourceNameFromConfig(config)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		if _, ok := configsMap[name]; !ok {
			errs = append(errs, fmt.Errorf(`error finding "Cluster" configuration for name "%s"`, name))
			continue
		}
		configsMap[name] = append(configsMap[name], config)
	}
	// Remove any configs that don't have the tail end (cluster related configs)
	result := make(map[string][]byte)
	for name, configs := range configsMap {
		if len(configs) <= 1 {
			errs = append(errs, fmt.Errorf(`no "ClusterConfig" found for cluster "%s"`, name))
			continue
		}
		collatedConfigBytes, err := utils.JoinYAML(configs)
		if err != nil {
			errs = append(errs, fmt.Errorf(`error collating YAML files for cluster "%s": %s`, name, err))
			continue
		}
		result[name] = collatedConfigBytes
		log.Printf(`final Configuration for cluster "%s": %#v`, name, string(collatedConfigBytes))
	}
	return result, errs
}

func resourceEKSClusterCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	log.Printf("create EKS cluster resource")
	c := config.GetConfig()
	logger := glogger.GetLogger()

	YamlConfigFilePath := d.Get("yamlfilepath").(string)

	fileBytes, err := utils.ReadYAMLFileContents(YamlConfigFilePath)
	if err != nil {
		return diag.FromErr(err)
	}

	// split the file and update individual resources
	y, uerr := utils.SplitYamlAndGetListByKind(fileBytes)
	if uerr != nil {
		return diag.FromErr(err)
	}

	var rafayConfigs, clusterConfigs [][]byte
	rafayConfigs = y["Cluster"]
	clusterConfigs = y["ClusterConfig"]
	if len(rafayConfigs) > 1 {
		return diag.FromErr(fmt.Errorf("%s", "only one cluster per config is supported"))
	}
	for _, yi := range rafayConfigs {
		log.Println("rafayConfig:", string(yi))
		name, project, _, err := findResourceNameFromConfig(yi)
		if err != nil {
			return diag.FromErr(fmt.Errorf("%s", "failed to get cluster name"))
		}
		log.Println("rafayConfig name:", name, "project:", project)
		if name != d.Get("name").(string) {
			return diag.FromErr(fmt.Errorf("%s", "cluster name does not match config file "))
		}
		if project != d.Get("projectname").(string) {
			return diag.FromErr(fmt.Errorf("%s", "project name does not match config file"))
		}
	}

	for _, yi := range clusterConfigs {
		log.Println("clusterConfig", string(yi))
		name, _, _, err := findResourceNameFromConfig(yi)
		if err != nil {
			return diag.FromErr(fmt.Errorf("%s", "failed to get cluster name"))
		}
		if name != d.Get("name").(string) {
			return diag.FromErr(fmt.Errorf("%s", "ClusterConfig name does not match config file"))
		}
	}

	// get project details
	resp, err := project.GetProjectByName(d.Get("projectname").(string))
	if err != nil {
		fmt.Print("project does not exist")
		return diags
	}
	project, err := project.NewProjectFromResponse([]byte(resp))
	if err != nil {
		fmt.Printf("project does not exist")
		return diags
	}

	// override config project
	c.ProjectID = project.ID

	configMap, errs := collateConfigsByName(rafayConfigs, clusterConfigs)
	if len(errs) > 0 {
		for _, err := range errs {
			log.Println("error in collateConfigsByName", err)
		}
		return diag.FromErr(fmt.Errorf("%s", "failed in collateConfigsByName"))
	}

	// Make request
	for clusterName, configBytes := range configMap {
		log.Println("create cluster:", clusterName, "config:", string(configBytes), "projectID :", c.ProjectID)
		if err := clusterctl.Apply(logger, c, clusterName, configBytes, false); err != nil {
			return diag.FromErr(fmt.Errorf("error performing apply on cluster %s: %s", clusterName, err))
		}
	}

	s, err := cluster.GetCluster(d.Get("name").(string), project.ID)
	if err != nil {
		log.Printf("error while getCluster %s", err.Error())
		return diag.FromErr(err)
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

	log.Printf("resource eks cluster created %s", s.ID)
	d.SetId(s.ID)

	return diags
}

func resourceEKSClusterRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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

func resourceEKSClusterUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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

	y, uerr := utils.SplitYamlAndGetListByKind(fileBytes)
	if uerr != nil {
		return diag.FromErr(err)
	}

	rafayConfigs := y["Cluster"]
	if len(rafayConfigs) > 1 {
		return diag.FromErr(fmt.Errorf("%s", "only one cluster per config is supported"))
	}

	var blueprintName, blueprintversion string
	for _, yi := range rafayConfigs {
		var err error
		blueprintName, blueprintversion, err = findBlueprintName(yi)
		if err != nil {
			return diag.FromErr(fmt.Errorf("%s", "failed to get blueprint name"))
		}
		log.Printf("blueprint name %s", blueprintName)
	}

	if cluster_resp.ClusterBlueprint != blueprintName || cluster_resp.ClusterBlueprintVersion != blueprintversion {
		cluster_resp.ClusterBlueprint = blueprintName

		if blueprintversion != "" {
			cluster_resp.ClusterBlueprintVersion = blueprintversion
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

	var versionstr, name string
	clusterConfigs := y["ClusterConfig"]
	for _, yi := range clusterConfigs {
		log.Println("clusterConfig", string(yi))
		name, _, versionstr, err = findResourceNameFromConfig(yi)
		if err != nil {
			return diag.FromErr(fmt.Errorf("%s", "failed to get cluster name"))
		}
		if name != d.Get("name").(string) {
			return diag.FromErr(fmt.Errorf("%s", "ClusterConfig name does not match config file"))
		}
	}

	log.Println("versionstr ", versionstr)
	if versionstr != "" {
		logger := glogger.GetLogger()
		err = cluster.UpgradeClusterEks(d.Get("name").(string),
			versionstr, project.ID, logger)
		if err != nil {
			log.Println("cluster upgrade response ", err)
		}
	}

	return diags
}

func resourceEKSClusterDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
