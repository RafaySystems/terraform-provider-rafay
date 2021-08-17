package rafay

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/RafaySystems/rctl/pkg/agent"
	"github.com/RafaySystems/rctl/pkg/clusteroverride"
	"github.com/RafaySystems/rctl/pkg/commands"
	"github.com/RafaySystems/rctl/pkg/config"
	"github.com/RafaySystems/rctl/pkg/models"
	"github.com/RafaySystems/rctl/pkg/project"
	"gopkg.in/yaml.v3"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAgent() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAgentCreate,
		ReadContext:   resourceAgentRead,
		UpdateContext: resourceAgentUpdate,
		DeleteContext: resourceAgentDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"projectname": {
				Type:     schema.TypeString,
				Required: true,
			},
			"agent_filepath": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceAgentCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	filePath := d.Get("agent_filepath").(string)
	var agentYaml commands.AgentYamlConfig
	//make sure this is the correct file path
	log.Println("filepath: ", filePath)
	log.Printf("create gitops agent resource")
	//get project id with project name, p.id used to refer to project id -> need p.ID for calling createAgent
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
	//open file path and retirve config spec from yaml file (from run function in commands/create_agent.go)
	//read and capture file from file path
	if f, err := os.Open(filePath); err == nil {
		c, err := ioutil.ReadAll(f)
		if err != nil {
			log.Println("error cpaturing file")
		}
		//implement createClusterOverride from commands/create_cluster_override.go -> then call clusteroverride.CreateClusterOverride
		agentDefinition := c
		err = yaml.Unmarshal(agentDefinition, &agentYaml)
		if err != nil {
			log.Printf("Failed Unmarshal correctly")
		}
		// check if project is provided from yaml file
		if agentYaml.Metadata.Project != "" {
			projectId, err = config.GetProjectIdByName(agentYaml.Metadata.Project)
			if err != nil {
				log.Println("error getting project ID from yaml file")
			}
		}
		// create the agent
		if strings.ToLower(agentYaml.Spec.Template.Type) == strings.ToLower(string(agent.ClusterAgent)) {
			spec := models.AgentSpec{
				AgentType:   models.AgentType(agent.ClusterAgent),
				ClusterName: agentYaml.Spec.Template.ClusterName,
			}
			if agentYaml.Spec.Template.ClusterName == "" {
				log.Println("you must provide a clusterName in the config")
			}
			err = agent.CreateAgent(agentYaml.Metadata.Name, projectId, spec)
		}

		if err != nil {
			log.Printf("Failed to create Agent: %s\n", agentYaml.Metadata.Name)
		} else {
			log.Printf("Successfully created Agent: %s\n", agentYaml.Metadata.Name)
		}
	}
	//set id to metadata.Name
	d.SetId(agentYaml.Metadata.Name)
	return diags
}

func resourceAgentRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
		if err != nil {
			log.Println("error cpaturing file")
		}
		//unmarshal yaml file to get correct specs
		clusterOverrideDefinition := c
		err = yaml.Unmarshal(clusterOverrideDefinition, &co)
		if err != nil {
			log.Printf("Failed Unmarshal correctly")
		}
		//get cluster override spec from yaml file
		_, err = getClusterOverrideSpecFromYamlConfigSpec(co, filePath)
		if err != nil {
			log.Printf("Failed to get ClusterOverrideSpecFromYamlConfigSpec")
		}
	} else {
		log.Println("Couldn't open file, err: ", err)
	}
	//get cluster override to ensure cluster override was created properly
	getClus_resp, err := clusteroverride.GetClusterOverride(co.Metadata.Name, projectId, co.Spec.Type)
	if err != nil {
		log.Println("get cluster override failed: ", getClus_resp, err)
	} else {
		log.Println("got newly created cluster override: ", co.Metadata.Name)
	}
	return diags
}

func resourceAgentUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
	// open and read file then unmarshal the data
	if f, err := os.Open(filePath); err == nil {
		// capture the entire file
		c, err := ioutil.ReadAll(f)
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
		// get cluster override spec from yaml file
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

func resourceAgentDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
		c, err := ioutil.ReadAll(f)
		if err != nil {
			log.Println("error cpaturing file")
		}
		clusterOverrideDefinition := c
		err = yaml.Unmarshal(clusterOverrideDefinition, &co)
		if err != nil {
			log.Printf("Failed Unmarshal correctly")
		}
	}
	//delete cluster override
	err = clusteroverride.DeleteClusterOverride(co.Metadata.Name, project_id, co.Spec.Type)
	if err != nil {
		fmt.Print("cluster was not deleted")
		log.Printf("cluster was not deleted, error %s", err.Error())
		return diag.FromErr(err)
	}
	return diags
}
