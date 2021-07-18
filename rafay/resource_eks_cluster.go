package rafay

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/RafaySystems/rctl/pkg/project"
	"github.com/RafaySystems/rctl/utils"
	"github.com/RafaySystems/rctl/pkg/config"
	"github.com/RafaySystems/rctl/pkg/clusterctl"
	"github.com/RafaySystems/rctl/pkg/cluster"
	glogger "github.com/RafaySystems/rctl/pkg/log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/go-yaml/yaml"
)

type configMetadata struct {
        Name string `yaml:"name"`
}

type configResourceType struct {
        Meta *configMetadata `yaml:"metadata"`
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
			"name": {
				Type: schema.TypeString,
				Required: true,
			},
			"projectname": {
				Type: schema.TypeString,
				Required: true,
			},
		},
	}
}

func findResourceNameFromConfig(configBytes []byte) (string, error) {
        var config configResourceType
        if err := yaml.Unmarshal(configBytes, &config); err != nil {
                return "", nil
        } else if config.Meta == nil {
                return "", fmt.Errorf("Invalid resource: No metadata found")
        } else if config.Meta.Name == "" {
                return "", fmt.Errorf("Invalid resource: No name specified in metadata")
        }
        return config.Meta.Name, nil
}


func collateConfigsByName(rafayConfigs, clusterConfigs [][]byte) (map[string][]byte, []error) {
        var errs []error
        configsMap := make(map[string][][]byte)
        // First find all rafay spec configurations
        for _, config := range rafayConfigs {
                name, err := findResourceNameFromConfig(config)
                if err != nil {
                        errs = append(errs, err)
                        continue
                }
                if _, ok := configsMap[name]; ok {
                        errs = append(errs, fmt.Errorf(`Duplicate "Cluster" resource with name "%s" found`, name))
                        continue
                }
                configsMap[name] = append(configsMap[name], config)
        }
        // Then append the cluster specific configurations
        for _, config := range clusterConfigs {
                name, err := findResourceNameFromConfig(config)
                if err != nil {
                        errs = append(errs, err)
                        continue
                }
                if _, ok := configsMap[name]; !ok {
                        errs = append(errs, fmt.Errorf(` Error finding "Cluster" configuration for name "%s"`, name))
                        continue
                }
                configsMap[name] = append(configsMap[name], config)
        }
        // Remove any configs that don't have the tail end (cluster related configs)
        result := make(map[string][]byte)
        for name, configs := range configsMap {
                if len(configs) <= 1 {
                        errs = append(errs, fmt.Errorf(`No "ClusterConfig" found for cluster "%s"`, name))
                        continue
                }
                collatedConfigBytes, err := utils.JoinYAML(configs)
                if err != nil {
                        errs = append(errs, fmt.Errorf(`Error collating YAML files for cluster "%s": %s`, name, err))
                        continue
                }
                result[name] = collatedConfigBytes
                log.Printf(`Final Configuration for cluster "%s": %#v`, name, string(collatedConfigBytes))
        }
        return result, errs
}

func resourceEKSClusterCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := config.GetConfig()
        logger := glogger.GetLogger()

	YamlConfigFilePath := d.Get("yamlfilepath").(string)

        log.Printf("Yaml FilePath  %s",YamlConfigFilePath )

	fileBytes, err := utils.ReadYAMLFileContents(YamlConfigFilePath)
        if err != nil {
                return diags
        }

        // split the file and update individual resources
       m, erru := utils.SplitYamlAndGetListByKind(fileBytes)
        if erru != nil {
                return diags
        }

	test, ok := m.(map[string]interface{}) //["Cluster"]
	test2, ok := m.(map[string]interface{}) //["ClusterConfig"]
	if ok {
		fmt.Printf("err")
	}

	configMap, errs := collateConfigsByName( test["Cluster"].([][]byte), test2["ClusterConfig"].([][]byte) )
//	configMap, errs := collateConfigsByName( test["Cluster"] , test2["ClusterConfig"] )
//	configMap, errs := collateConfigsByName( *m["Cluster"], *m["ClusterConfig"] )
        // Make request
        for clusterName, configBytes := range configMap {
                if err := clusterctl.Apply(logger,c, clusterName, configBytes, true ); err != nil {
                        errs = append(errs, fmt.Errorf(`Error performing apply on cluster "%s": %s`, clusterName, err))
                        continue
                }
        }
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
        s, err := cluster.GetCluster(d.Get("name").(string), project.ID )
        if err != nil {
                log.Printf("error while getCluster %s", err.Error())
                return diag.FromErr(err)
        }

        log.Printf("resource eks cluster created %s", s.ID)
        d.SetId(s.ID)

	return diags
}

func resourceEKSClusterRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
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
	c, err := cluster.GetCluster(d.Get("name").(string), project.ID )
	if err != nil {
		log.Printf("error while getCluster %s", err.Error())
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
	log.Printf("get cloudprovider " )
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

	errDel := cluster.DeleteCluster(d.Get("name").(string),project.ID)
	if errDel != nil {
		log.Printf("delete cluster error %s", errDel.Error())
		return diag.FromErr(errDel)
	}

	return diags
}
