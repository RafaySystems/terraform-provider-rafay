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
	"github.com/RafaySystems/rctl/pkg/project"
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
		//return createClusterOverride(projectId, c, cmd, d.Get("cluster_override_filepath").(string))
		clusterOverrideDefinition := c

		err = yaml.Unmarshal(clusterOverrideDefinition, &co)
		if err != nil {
			log.Printf("Failed Unmarshal correctly")
		}

		// check if project is provided
		spec, err := commands.getClusterOverrideSpecFromYamlConfigSpec(co, filePath)
		if err != nil {
			log.Printf("Failed to get ClusterOverrideSpecFromYamlConfigSpec")
		}
		err = clusteroverride.CreateClusterOverride(co.Metadata.Name, projectId, *spec)
		if err != nil {
			log.Printf("Failed to create cluster override: %s\n", co.Metadata.Name)
			//return err
		} else {
			log.Printf("Successfully created cluster override: %s\n", co.Metadata.Name)
		}

	} else {
		log.Println("Couldn't open file")
	}
	/*create cluster override
	err = clusteroverride.CreateClusterOverride(d.Get("clustername").(string), projectId, d.Get("cluster_override_filepath").(string))
	if err != nil {
		log.Printf("Cluster Override Creation fail, error %s", err.Error())
		return diag.FromErr(err)
	}*/
	//retrieve cotype variable (must change) -> format cluster_override_spec to access Type variable
	coType := co.Type
	//get cluster override to ensure cluster override was created properly
	getClus_resp, err := clusteroverride.GetClusterOverride(d.Get("clustername").(string), projectId, coType)
	if err != nil {
		log.Println("get cluster override failed: ", getClus_resp, err)
	}
	//figure out what to set id too
	d.SetId("")
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

	//retrieve project_id from project name for calling get_cluster
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
	//retrieve cluster_details from get cluster to pass into update cluster
	cluster_resp, err := cluster.GetCluster(d.Get("clustername").(string), project_id)
	if err != nil {
		log.Printf("imported cluster was not created, error %s", err.Error())
		return diag.FromErr(err)
	}
	// read the blueprint name
	if d.Get("blueprint").(string) != "" {
		cluster_resp.ClusterBlueprint = d.Get("blueprint").(string)
	}
	// read the blueprint version
	if d.Get("blueprint_version").(string) != "" {
		cluster_resp.ClusterBlueprintVersion = d.Get("blueprint_version").(string)
	}
	//update cluster to send updated cluster details to core
	err = cluster.UpdateCluster(cluster_resp)
	if err != nil {
		log.Printf("cluster was not updated, error %s", err.Error())
		return diag.FromErr(err)
	}

	return diags
}

func resourceClusterOverrideDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("resource cluster override delete id %s", d.Id())

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
	//delete cluster once project id is retrieved correctly
	//format cluster_override_spec to access Type variable
	err = clusteroverride.DeleteClusterOverride(d.Get("clustername").(string), project_id, d.Get("cluster_override_spec").Type)
	if err != nil {
		fmt.Print("cluster was not deleted")
		log.Printf("cluster was not deleted, error %s", err.Error())
		return diag.FromErr(err)
	}
	return diags
}
