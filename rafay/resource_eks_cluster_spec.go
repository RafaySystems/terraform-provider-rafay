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
	"github.com/google/go-cmp/cmp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/go-yaml/yaml"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type configMetadataSpec struct {
	Name    string `yaml:"name"`
	Project string `yaml:"project"`
	Version string `yaml:"version"`
}

type configResourceTypeSpec struct {
	Meta *configMetadataSpec `yaml:"metadata"`
}

type blueprintClusterSpec struct {
	Blueprint        string `yaml:"blueprint"`
	Blueprintversion string `yaml:"blueprintversion"`
}

type blueprintTypeSpec struct {
	Spec *blueprintClusterSpec `yaml:"spec"`
}

func resourceEKSClusterSpec() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceEKSClusterSpecCreate,
		ReadContext:   resourceEKSClusterSpecRead,
		UpdateContext: resourceEKSClusterSpecUpdate,
		DeleteContext: resourceEKSClusterSpecDelete,

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
				Default:  1,
			},
			"checkdiff": {
				Type:     schema.TypeBool,
				Optional: true,
			},
		},
	}
}

func findResourceNameFromConfigSpec(configBytes []byte) (string, string, string, error) {
	var config configResourceTypeSpec
	if err := yaml.Unmarshal(configBytes, &config); err != nil {
		return "", "", "", nil
	} else if config.Meta == nil {
		return "", "", "", fmt.Errorf("%s", "Invalid resource: No metadata found")
	} else if config.Meta.Name == "" {
		return "", "", "", fmt.Errorf("%s", "Invalid resource: No name specified in metadata")
	}
	return config.Meta.Name, config.Meta.Project, config.Meta.Version, nil
}

func findBlueprintNameSpec(configBytes []byte) (string, string, error) {
	var blueprint blueprintTypeSpec
	if err := yaml.Unmarshal(configBytes, &blueprint); err != nil {
		return "", "", nil
	} else if blueprint.Spec == nil {
		return "", "", fmt.Errorf("%s", "Invalid resource: No spec found")
	} else if blueprint.Spec.Blueprint == "" {
		return "", "", fmt.Errorf("%s", "Invalid resource: No name specified in spec")
	}

	return blueprint.Spec.Blueprint, blueprint.Spec.Blueprintversion, nil

}
func collateConfigsByNameSpec(rafayConfigs, clusterConfigs [][]byte) (map[string][]byte, []error) {
	var errs []error
	configsMap := make(map[string][][]byte)
	// First find all rafay spec configurations
	for _, config := range rafayConfigs {
		name, _, _, err := findResourceNameFromConfigSpec(config)
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
		name, _, _, err := findResourceNameFromConfigSpec(config)
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
		if len(configs) <= 0 {
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

func resourceEKSClusterSpecCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
	y, _, uerr := utils.SplitYamlAndGetListByKind(fileBytes)
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
		name, project, _, err := findResourceNameFromConfigSpec(yi)
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
		name, _, _, err := findResourceNameFromConfigSpec(yi)
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

	configMap, errs := collateConfigsByNameSpec(rafayConfigs, clusterConfigs)
	if len(errs) > 0 {
		for _, err := range errs {
			log.Println("error in collateConfigsByNameSpec", err)
		}
		return diag.FromErr(fmt.Errorf("%s", "failed in collateConfigsByNameSpec"))
	}

	// Make request
	response := ""
	for clusterName, configBytes := range configMap {
		/* support only one cluster per spec */
		log.Println("create cluster:", clusterName, "config:", string(configBytes), "projectID :", c.ProjectID)
		response, err = clusterctl.Apply(logger, c, clusterName, configBytes, false, false)
		if err != nil {
			return diag.FromErr(fmt.Errorf("error performing apply on cluster %s: %s", clusterName, err))
		}
		break
	}

	s, err := cluster.GetCluster(d.Get("name").(string), project.ID)
	if err != nil {
		log.Printf("error while getCluster %s", err.Error())
		return diag.FromErr(err)
	}

	d.SetId(s.ID)
	if d.Get("waitflag").(string) == "1" {
		log.Printf("Cluster Provision may take upto 15-20 Minutes")
		res := clusterCTLResponse{}
		err = json.Unmarshal([]byte(response), &res)
		if err != nil {
			log.Println("response parse error", err)
			return diag.FromErr(err)
		}
		if res.TaskSetID == "" {
			return nil
		}

		for { //wait for cluster to provision correctly
			time.Sleep(60 * time.Second)
			check, errGet := cluster.GetCluster(d.Get("name").(string), project.ID)
			if errGet != nil {
				log.Printf("error while getCluster %s", errGet.Error())
				return diag.FromErr(errGet)
			}

			statusResp, err := eksClusterCTLStatus(res.TaskSetID, project.ID)
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
				if check.Status == "READY" {
					break
				}
				log.Println("task completed but cluster is not ready")
			}
			if strings.Contains(sres.Status, "STATUS_FAILED") {
				return diag.FromErr(fmt.Errorf("failed to create/update cluster while provisioning cluster %s %s", d.Get("name").(string), statusResp))
			}
		}
	}

	log.Printf("resource eks cluster created %s", s.ID)

	return diags
}

func resourceEKSClusterSpecRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
		if strings.Contains(err.Error(), "not found") {
			log.Println("Resource Read ", "error", err)
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	check := d.Get("checkdiff").(bool)
	if check {
		logger := glogger.GetLogger()
		rctlCfg := config.GetConfig()
		clusterSpecYaml, err := clusterctl.GetClusterSpec(logger, rctlCfg, c.Name, project.ID)
		if err != nil {
			log.Printf("error in get clusterspec %s", err.Error())
			return diag.FromErr(err)
		}

		YamlConfigFilePath := d.Get("yamlfilepath").(string)
		fileBytes, err := utils.ReadYAMLFileContents(YamlConfigFilePath)
		if err != nil {
			return diag.FromErr(err)
		}
		localYaml := string(fileBytes)
		log.Println("localYaml:", localYaml)
		log.Println("clusterSpecYaml:", clusterSpecYaml)
		if diff := cmp.Diff(localYaml, clusterSpecYaml); diff != "" {
			log.Println("cmp.Diff: ", diff)
			diags = make([]diag.Diagnostic, 1)
			diags[0].Severity = diag.Warning
			diags[0].Summary = fmt.Sprintf("Your infrastructure may be drifted from configuration.\n +<<diff \n%+v\n-<<diff\n", diff)
		}
	}

	if err := d.Set("name", c.Name); err != nil {
		log.Printf("get group set name error %s", err.Error())
		return diag.FromErr(err)
	}

	return diags
}

func resourceEKSClusterSpecUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceEKSClusterSpecCreate(ctx, d, m)
}

func resourceEKSClusterSpecDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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

	errDel := cluster.DeleteCluster(d.Get("name").(string), project.ID, false)
	if errDel != nil {
		log.Printf("delete cluster error %s", errDel.Error())
		return diag.FromErr(errDel)
	}

	if d.Get("waitflag").(string) == "1" {
		for {
			time.Sleep(60 * time.Second)
			check, errGet := cluster.GetCluster(d.Get("name").(string), project.ID)
			if errGet != nil {
				log.Printf("error while getCluster %s, delete success", errGet.Error())
				break
			}
			if check == nil || (check != nil && check.Status != "READY") {
				break
			}
		}
	}

	return diags
}
