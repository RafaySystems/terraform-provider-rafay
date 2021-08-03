package rafay

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/RafaySystems/rctl/pkg/cluster"
	"github.com/RafaySystems/rctl/pkg/clusteroverride"
	"github.com/RafaySystems/rctl/pkg/commands"
	"github.com/RafaySystems/rctl/pkg/models"
	"github.com/RafaySystems/rctl/pkg/project"
	"github.com/RafaySystems/rctl/utils"
	"gopkg.in/yaml.v3"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceClusterOverride() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceClusterOverrideCreate,
		ReadContext:   resourceClusterOverrideRead,
		UpdateContext: resourceClusterOverrideUpdate,
		DeleteContext: resourceClusterOverrideDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
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
			"cluster_override_filepath": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceClusterOverrideCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	filePath := d.Get("cluster_override_filepath").(string)
	var co commands.ClusterOverrideYamlConfig

	log.Printf("create cluster override resource")
	//get project id with project name, p.id used to refer to project id -> need p.ID for calling createClusterOverride and getClusterOverride
	resp, err := project.GetProjectByName(d.Get("projectname").(string))
	if err != nil {
		log.Printf("project does not exist, error %s", err.Error())
		return diag.FromErr(err)
	}
	p, err := project.NewProjectFromResponse([]byte(resp))
	if err != nil {
		return diag.FromErr(err)
	} else if p == nil {
		d.SetId("")
		return diags
	}
	projectId := p.ID
	//open file path and retirve config spec from yaml file (from run function in commands/create_cluster_override.go)
	//read and capture file from file path
	if f, err := os.Open(filePath); err == nil {
		// capture the entire file
		c, err := ioutil.ReadAll(f)
		//log.GetLogger().Debugf("file is:\n%s", c)
		if err != nil {
			log.Println("error cpaturing file")
		}

		//implement createClusterOverride from commands/create_cluster_override.go -> then call clusteroverride.CreateClusterOverride
		clusterOverrideDefinition := c

		err = yaml.Unmarshal(clusterOverrideDefinition, &co)
		if err != nil {
			log.Printf("Failed Unmarshal correctly")
		}

		// check if project is provided
		spec, err := getClusterOverrideSpecFromYamlConfigSpec(co, filePath)
		if err != nil {
			log.Printf("Failed to get ClusterOverrideSpecFromYamlConfigSpec")
		}
		//create cluster override
		err = clusteroverride.CreateClusterOverride(co.Metadata.Name, projectId, *spec)
		if err != nil {
			log.Printf("Failed to create cluster override: %s\n", co.Metadata.Name)
		} else {
			log.Printf("Successfully created cluster override: %s\n", co.Metadata.Name)
		}

	} else {
		log.Println("Couldn't open file")
	}

	//retrieve cotype variable (must change) -> format cluster_override_spec to access Type variable
	coType := co.Spec.Type

	//get cluster override to ensure cluster override was created properly
	getClus_resp, err := clusteroverride.GetClusterOverride(d.Get("clustername").(string), projectId, coType)
	if err != nil {
		log.Println("get cluster override failed: ", getClus_resp, err)
	}

	//set id to metadata.Name
	d.SetId(co.Metadata.Name)
	return diags
}

func resourceClusterOverrideRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
	c, err := cluster.GetCluster(d.Get("clustername").(string), project.ID)
	if err != nil {
		log.Printf("error in get cluster %s", err.Error())
		return diag.FromErr(err)
	}
	if err := d.Set("clustername", c.Name); err != nil {
		log.Printf("get group set name error %s", err.Error())
		return diag.FromErr(err)
	}

	return diags
}

func resourceClusterOverrideUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("update cluster override resource")
	filePath := d.Get("cluster_override_filepath").(string)
	createIfNotPresent := false //this is how it is set in commands/update_cluster_override

	//retrieve project_id from project name
	resp, err := project.GetProjectByName(d.Get("projectname").(string))
	if err != nil {
		log.Println("error cpaturing file")
	}
	p, err := project.NewProjectFromResponse([]byte(resp))
	if err != nil {
		return diag.FromErr(err)
	} else if p == nil {
		d.SetId("")
		return diags
	}
	project_id := p.ID
	//update cluster implemented from commmands/update_cluster_override -> will call UpdateClusterOverride from cluster_override.go
	// open file and unmarshal the data
	if f, err := os.Open(filePath); err == nil {
		// capture the entire file
		c, err := ioutil.ReadAll(f)
		//log.GetLogger().Debugf("file is:\n%s", c)
		if err != nil {
			log.Println("error reading file")
			return diags
		}

		clusterOverrideDefinition := c
		var co commands.ClusterOverrideYamlConfig
		err = yaml.Unmarshal(clusterOverrideDefinition, &co)
		if err != nil {
			log.Println("error unmarshalling Cluster Override")
			return diags
		}
		// check if project is provided
		spec, err := getClusterOverrideSpecFromYamlConfigSpec(co, filePath)
		if err != nil {
			log.Println("error getting Cluster Override Spec From Yaml Config Spec")
			return diags
		}
		//update cluster
		err = clusteroverride.UpdateClusterOverride(co.Metadata.Name, project_id, *spec, createIfNotPresent)
		if err != nil {
			log.Printf("Failed to update cluster override: %s\n", co.Metadata.Name)
			return diags
		} else {
			log.Printf("Successfully created/updated cluster override: %s\n", co.Metadata.Name)
		}

		return diags

	} else {
		log.Println("error opening file")
		return diags
	}
}

func resourceClusterOverrideDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("resource cluster override delete id %s", d.Id())
	var co commands.ClusterOverrideYamlConfig
	filePath := d.Get("cluster_override_filepath").(string)

	//get project id with project name, p.id used to refer to project id -> need p.ID for calling deleteClusterOverride
	resp, err := project.GetProjectByName(d.Get("projectname").(string))
	if err != nil {
		log.Printf("project does not exist, error %s", err.Error())
		return diag.FromErr(err)
	}
	p, err := project.NewProjectFromResponse([]byte(resp))
	if err != nil {
		return diag.FromErr(err)
	} else if p == nil {
		d.SetId("")
		return diags
	}

	project_id := p.ID
	//open, read, and unmarshal file to retrieve ClusterOverrideYamlConfig Struct to pass in type for delete cluster
	if f, err := os.Open(filePath); err == nil {
		// capture the entire file
		c, err := ioutil.ReadAll(f)
		//log.GetLogger().Debugf("file is:\n%s", c)
		if err != nil {
			log.Println("error cpaturing file")
		}

		//implement createClusterOverride from commands/create_cluster_override.go -> then call clusteroverride.CreateClusterOverride
		clusterOverrideDefinition := c

		err = yaml.Unmarshal(clusterOverrideDefinition, &co)
		if err != nil {
			log.Printf("Failed Unmarshal correctly")
		}
	}
	//format cluster_override_spec to access Type variable
	err = clusteroverride.DeleteClusterOverride(d.Get("clustername").(string), project_id, co.Spec.Type)
	if err != nil {
		fmt.Print("cluster was not deleted")
		log.Printf("cluster was not deleted, error %s", err.Error())
		return diag.FromErr(err)
	}
	return diags
}

func getClusterOverrideSpecFromYamlConfigSpec(co commands.ClusterOverrideYamlConfig, filePath string) (*models.ClusterOverrideSpec, error) {
	var spec models.ClusterOverrideSpec
	spec.ClusterSelector = co.Spec.ClusterSelector
	spec.ResourceSelector = co.Spec.ResourceSelector
	spec.Type = co.Spec.Type
	spec.RepositoryRef = co.Spec.ValueRepoRef
	spec.RepoArtifactMeta = co.Spec.ValueRepoArtifactMeta
	if co.Spec.OverrideValues != "" && co.Spec.ValuesFile != "" {
		return nil, fmt.Errorf("invalid config: both overrideValues and overrideValuesFile were provided")
	}
	if co.Spec.ValuesFile != "" {
		values, err := getOverrideValues(co, filePath)
		if err != nil {
			return nil, fmt.Errorf("invalid config: error fetching the content of the value file from the location provided %s: Error: %s", co.Spec.ValuesFile, err.Error())
		}
		spec.OverrideValues = values
	}
	if co.Spec.OverrideValues != "" {
		spec.OverrideValues = co.Spec.OverrideValues
	}
	if spec.OverrideValues == "" && spec.RepositoryRef == "" {
		return nil, fmt.Errorf("invalid config: neither overrideValues not valueRepoRef were provided")
	}
	if spec.OverrideValues != "" && spec.RepositoryRef != "" {
		return nil, fmt.Errorf("invalid config: both overrideValues and valueRepoRef were provided")
	}
	if spec.RepositoryRef != "" && (spec.RepoArtifactMeta.Git == nil || len(spec.RepoArtifactMeta.Git.RepoArtifactFiles) == 0) {
		return nil, fmt.Errorf("invalid config: exactly one repo artifact file should be provided.\"")
	}
	return &spec, nil
}

func getOverrideValues(co commands.ClusterOverrideYamlConfig, filePath string) (string, error) {
	valueFileLocation := utils.FullPath(filePath, co.Spec.ValuesFile)
	if _, err := os.Stat(valueFileLocation); os.IsNotExist(err) {
		return "", fmt.Errorf("values file doesn't exist '%s'", valueFileLocation)
	}
	valueFileContent, err := ioutil.ReadFile(valueFileLocation)
	if err != nil {
		return "", fmt.Errorf("error in reading the value file %s: %s\n", valueFileLocation, err)
	}
	return string(valueFileContent), nil
}
