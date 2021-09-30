package rafay

import (
	"context"
	"io/ioutil"
	"log"
	"os"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/RafaySystems/rctl/pkg/commands"
	"github.com/RafaySystems/rctl/pkg/config"
	"github.com/RafaySystems/rctl/pkg/infraprovisioner"
	"github.com/RafaySystems/rctl/pkg/project"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceInfrastuctureProvisioner() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceInfrastuctureProvisionerCreate,
		ReadContext:   resourceInfrastuctureProvisionerRead,
		UpdateContext: resourceInfrastuctureProvisionerUpdate,
		DeleteContext: resourceInfrastuctureProvisionerDelete,

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
			"infrastucture_provisioner_filepath": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceInfrastuctureProvisionerCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	filePath := d.Get("infrastucture_provisioner_filepath").(string)
	var ipYamlConfig commands.InfraProvisionerYamlConfig
	//make sure this is the correct file path
	log.Println("filepath: ", filePath)
	log.Printf("create infrasturcture filepath resource")
	//get project id with project name, p.id used to refer to project id -> need p.ID for calling Create Infrastructure Provisioner
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
	//open file path and retirve config spec from yaml file (from run function in commands/create_infrastructure.go)
	//cant call function but need to take code from there because we need metadata.name to set it
	//read and capture file from file path
	if f, err := os.Open(filePath); err == nil {
		// capture the entire file
		ipYaml, err := ioutil.ReadAll(f)
		if err != nil {
			log.Println("error reading file")
		}
		err = yaml.Unmarshal(ipYaml, &ipYamlConfig)
		if err != nil {
			log.Println("err unmarhsalling data")
		}

		// check if project is provided
		projId := projectId
		if ipYamlConfig.Metadata.Project != "" {
			projId, err = config.GetProjectIdByName(ipYamlConfig.Metadata.Project)
			if err != nil {
				log.Println("error getting project id by name:", err)
			}
		}

		mwt, err := commands.ConvertInfraProvisionerYAMLToModel(&ipYamlConfig)
		if err != nil {
			log.Println("error converting infra provisioner yaml to model:", err)
		}

		err = infraprovisioner.CreateInfraProvisioner(projId, mwt)
		if err != nil {
			log.Println("error creating infra provisioner:", err)
		} else {
			log.Println("Succesfully created infra provisioner!")
		}
	}
	//Set metadataname as id
	d.SetId(ipYamlConfig.Metadata.Name)
	return diags
}

func resourceInfrastuctureProvisionerUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var ip commands.InfraProvisionerYamlConfig
	createIfNotPresent := false
	filePath := d.Get("infrastucture_provisioner_filepath").(string)
	//make sure this is the correct file path
	log.Println("filepath: ", filePath)
	log.Printf("update infrastructure provisioner resource")
	//get project id with project name, p.id used to refer to project id -> need p.ID for calling Create Namespace
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
	//open file path and retirve config spec from yaml file (from run function in commands/create_repositories.go)
	//read and capture file from file path
	if f, err := os.Open(filePath); err == nil {
		// capture the entire file
		c, err := ioutil.ReadAll(f)
		if err != nil {
			log.Println("error reading file")
		}
		//unmarshal the data
		err = yaml.Unmarshal(c, &ip)
		if err != nil {
			log.Println("error unmarhsalling data")
		}
		// check if project is provided
		projId := projectId
		if ip.Metadata.Project != "" {
			projId, err = config.GetProjectIdByName(ip.Metadata.Project)
			if err != nil {
				log.Println("err getting prohject id by name:", err)
			}
		}

		mip, err := commands.ConvertInfraProvisionerYAMLToModel(&ip)
		if err != nil {
			log.Println("error converting infra provisioner yaml to model:", err)
		}

		err = infraprovisioner.UpdateInfraProvisioner(projId, mip, createIfNotPresent)
		if err != nil {
			log.Println("error creating infra provisioner:", err)
		} else {
			log.Println("Succesfully updated infra provisioner!")
		}
	}
	return diags
}

func resourceInfrastuctureProvisionerRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	return diags
}
func resourceInfrastuctureProvisionerDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	//get project id with project name, p.id used to refer to project id -> need p.ID for calling Create Namespace
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
	//Delete namespace
	err = infraprovisioner.DeleteInfraProvisioner(string(d.Id()), projectId)
	if err != nil {
		log.Println("error delete infra provisioner: ", err)
	} else {
		log.Println("Succesfully deleted infra provisioner")
	}
	return diags
}
