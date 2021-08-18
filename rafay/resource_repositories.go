package rafay

import (
	"context"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/RafaySystems/rctl/pkg/commands"
	"github.com/RafaySystems/rctl/pkg/config"
	"github.com/RafaySystems/rctl/pkg/models"
	"github.com/RafaySystems/rctl/pkg/project"
	"github.com/RafaySystems/rctl/pkg/repository"
	"gopkg.in/yaml.v3"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceRepositories() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRepositoriesCreate,
		ReadContext:   resourceRepositoriesRead,
		UpdateContext: resourceRepositoriesUpdate,
		DeleteContext: resourceRepositoriesDelete,

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
			"repositories_filepath": {
				Type:     schema.TypeString,
				Required: true,
			},
			"delete_repositories": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceRepositoriesCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	filePath := d.Get("agent_filepath").(string)
	var r commands.RepositoryYamlConfig
	//make sure this is the correct file path
	log.Println("filepath: ", filePath)
	log.Printf("create integrations repositories resource")
	//get project id with project name, p.id used to refer to project id -> need p.ID for calling createRepositories
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
		c, err := ioutil.ReadAll(f)
		if err != nil {
			log.Println("error cpaturing file")
		}
		//implement createClusterOverride from commands/create_cluster_override.go -> then call clusteroverride.CreateClusterOverride
		repositoryDefinition := c
		err = yaml.Unmarshal(repositoryDefinition, &r)
		if err != nil {
			log.Printf("Failed Unmarshal correctly")
		}
		// check if project is provided from yaml file
		if r.Metadata.Project != "" {
			projectId, err = config.GetProjectIdByName(r.Metadata.Project)
			if err != nil {
				log.Println("error getting project ID from yaml file")
			}
		}
		// create the agent
		var spec models.RepositorySpec
		spec.Endpoint = r.Spec.Endpoint
		spec.Insecure = r.Spec.Insecure
		spec.CaCert = r.Spec.CACert
		switch r.Spec.RepositoryType {
		case repository.GitRepository:
			spec.RepositoryType = repository.GitRepository
		case repository.HelmRepository:
			spec.RepositoryType = repository.HelmRepository
		default:
			log.Println("invalid repositoryType, must one of", r.Spec.RepositoryType, strings.Join(repository.AllowedTypes, "|"))
		}

		switch r.Spec.CredentialType {
		case repository.SSHCredential:
			spec.CredentialType = repository.SSHCredential
			spec.Credentials.Ssh = &models.RepoSSHCredentials{}
			if r.Spec.Credentials.SSH.SSHPrivateKey != "" && r.Spec.Credentials.SSH.SSHPrivateKeyFile != "" {
				log.Printf("Found multiple values for ssh private key. Please specify either \"sshPrivateKey\" or \"sshPrivateKeyFile\"")
			}
			if r.Spec.Credentials.SSH.SSHPrivateKeyFile != "" {
				// read sshPrivateKeyFile
				if f, err := os.Open(r.Spec.Credentials.SSH.SSHPrivateKeyFile); err == nil {
					c, err := ioutil.ReadAll(f)
					log.Println("sshPrivateKeyFile Content is:\n", c)
					if err != nil {
						log.Println("error reading ssh private key files")
					}
					spec.Credentials.Ssh.SshPrivateKey = string(c)
				} else {
					log.Println("error opening ssh private key files")
				}
			} else {
				spec.Credentials.Ssh.SshPrivateKey = r.Spec.Credentials.SSH.SSHPrivateKey
			}
		case repository.UserPassCredential:
			spec.CredentialType = repository.UserPassCredential
			spec.Credentials.UserPass = &models.RepoUserPassCredentials{}
			spec.Credentials.UserPass.Username = r.Spec.Credentials.UserPass.Username
			spec.Credentials.UserPass.Password = r.Spec.Credentials.UserPass.Password
		case repository.CredentialTypeNotSet:
			spec.CredentialType = repository.CredentialTypeNotSet
		default:
			log.Println("invalid repository.CredentialType, must one of", r.Spec.CredentialType, strings.Join(repository.AllowedCrentialTypes, "|"))
		}
		spec.AgentNames = r.Spec.Agents

		err = repository.CreateRepository(r.Metadata.Name, projectId, spec)
		if err != nil {
			log.Printf("Failed to create Repository: %s\n", r.Metadata.Name)
		} else {
			log.Printf("Successfully created repository: %s\n", r.Metadata.Name)
		}
	}
	//set id to metadata.Name
	d.SetId(r.Metadata.Name)
	return diags
}

func resourceRepositoriesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	/*
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
		}*/
	return diags
}

func resourceRepositoriesUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Println("rctl doesn't have update functionality for agent")
	return diags
}

func resourceRepositoriesDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("deleting repositories")

	//convert namesapce interface to passable list for function
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
	//get project id with project name, p.id used to refer to project id -> need p.ID for calling DeleteRepository
	if d.Get("delete_repositories") != nil {
		deleterepositoriesList := d.Get("delete_repositories").([]interface{})
		deleterepositories := make([]string, len(deleterepositoriesList))
		for i, raw := range deleterepositoriesList {
			deleterepositories[i] = raw.(string)
		}
		// delete the specified repositories
		for _, r := range deleterepositories {
			if err := repository.DeleteRepository(r, projectId); err != nil {
				log.Println("error deleting repositories")
			}
			log.Println("Deleted repositories: ", r)
		}
	}
	return diags
}
