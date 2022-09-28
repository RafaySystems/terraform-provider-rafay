package rafay

import (
	"context"
	"encoding/json"
	"errors"
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
		CloudProvider    string `yaml:"cloudprovider"`
	} `yaml:"spec"`
}

func resourceAKSClusterSpec() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAKSClusterSpecCreate,
		ReadContext:   resourceAKSClusterSpecRead,
		UpdateContext: resourceAKSClusterSpecUpdate,
		DeleteContext: resourceAKSClusterSpecDelete,

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

func aksClusterSpecCTL(config *config.Config, rafayConfigs, clusterConfigs [][]byte, dryRun bool) (string, []error) {
	var errs []error
	logger := glogger.GetLogger()
	configMap, errs := collateConfigsByName(rafayConfigs, clusterConfigs)
	// Make request
	for clusterName, configBytes := range configMap {
		/* only suppoort one cluster */
		rsponse, err := clusterctl.Apply(logger, config, clusterName, configBytes, dryRun, false)

		if err != nil {
			log.Println("error performing apply on cluster: ", clusterName, err)
			errs = append(errs, fmt.Errorf("Error performing apply on cluster %s: %s", clusterName, err))
			return rsponse, errs
		}
		return rsponse, nil
		// if _, err := clusterctl.Apply(logger, config, clusterName, configBytes, dryRun); err != nil {
		// 	errs = append(errs, fmt.Errorf(`Error performing apply on cluster "%s": %s`, clusterName, err))
		// 	continue
		// }
	}
	return "", nil
}

func resourceAKSClusterSpecUpsert(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var errs []error

	rctlCfg := config.GetConfig()

	YamlConfigFilePath := d.Get("yamlfilepath").(string)

	fileBytes, err := utils.ReadYAMLFileContents(YamlConfigFilePath)
	if err != nil {
		return diag.FromErr(err)
	}

	// split the file and update individual resources
	cfgList, _, err := utils.SplitYamlAndGetListByKind(fileBytes)
	if err != nil {
		return diag.FromErr(err)
	}
	if len(cfgList) > 1 {
		fmt.Printf("found more than one cluster config in %s", YamlConfigFilePath)
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

	// clusters
	response := ""
	if _, ok := cfgList["Cluster"]; ok {
		response, errs = aksClusterSpecCTL(rctlCfg, cfgList["Cluster"], cfgList["ClusterConfig"], false)
		if errs != nil && len(errs) > 0 {
			s := ""
			for index, err := range errs {
				if index != 0 {
					s += "\n"
				}
				s += err.Error()
			}
			return diag.FromErr(errors.New(s))
		}

	}
	log.Printf("process_filebytes response : %s", response)
	res := clusterCTLResponse{}
	err = json.Unmarshal([]byte(response), &res)
	if err != nil {
		log.Println("response parse error", err)
		return diag.FromErr(err)
	}
	if res.TaskSetID == "" {
		return nil
	}

	time.Sleep(10 * time.Second)
	s, errGet := cluster.GetCluster(d.Get("name").(string), project.ID)
	if errGet != nil {
		log.Printf("error while getCluster %s", errGet.Error())
		return diag.FromErr(errGet)
	}

	d.SetId(s.ID)
	if d.Get("waitflag").(string) == "1" {
		log.Printf("Cluster Provision may take upto 15-20 Minutes")
		for {
			time.Sleep(60 * time.Second)
			check, errGet := cluster.GetCluster(d.Get("name").(string), project.ID)
			if errGet != nil {
				log.Printf("error while getCluster %s", errGet.Error())
				return diag.FromErr(errGet)
			}

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
				if check.Status == "READY" {
					break
				}
				log.Println("task completed but cluster is not ready")
			}
			if strings.Contains(sres.Status, "STATUS_FAILED") {
				return diag.FromErr(fmt.Errorf("failed to create/update cluster while ",
					"provisioning cluster %s %s",
					d.Get("projectname"), statusResp))
			}
		}
	}
	log.Printf("resource aks cluster created/updated %s", s.ID)

	return diags
}

func resourceAKSClusterSpecCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("create AKS cluster resource")
	return resourceAKSClusterSpecUpsert(ctx, d, m)
}

func resourceAKSClusterSpecRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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

func resourceAKSClusterSpecUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("update EKS cluster resource")

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

	YamlConfigFilePath := d.Get("yamlfilepath").(string)

	fileBytes, err := utils.ReadYAMLFileContents(YamlConfigFilePath)
	if err != nil {
		return diag.FromErr(err)
	}

	var c clusterYamlConfig
	if err = yaml.Unmarshal(fileBytes, &c); err != nil {
		return diag.FromErr(err)
	}

	return resourceAKSClusterSpecUpsert(ctx, d, m)
}

func resourceAKSClusterSpecDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
