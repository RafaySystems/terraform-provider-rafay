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
			"repository_filepath": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceRepositoriesCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	filePath := d.Get("repository_filepath").(string)
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
	return diags
}

func resourceRepositoriesUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	createIfNotPresent := false
	log.Println("update repository")
	filePath := d.Get("repository_filepath").(string)
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
				log.Println("Found multiple values for ssh private key. Please specify either \"sshPrivateKey\" or \"sshPrivateKeyFile\"")
			}
			if r.Spec.Credentials.SSH.SSHPrivateKeyFile != "" {
				// read sshPrivateKeyFile
				if f, err := os.Open(r.Spec.Credentials.SSH.SSHPrivateKeyFile); err == nil {
					c, err := ioutil.ReadAll(f)
					log.Println("sshPrivateKeyFile Content is:\n", c)
					if err != nil {
						log.Println("error reading ssh private key file")
					}
					spec.Credentials.Ssh.SshPrivateKey = string(c)
				} else {
					log.Println("error opening ssh private key file")
				}
			} else {
				spec.Credentials.Ssh.SshPrivateKey = r.Spec.Credentials.SSH.SSHPrivateKey
			}
		case repository.UserPassCredential:
			spec.CredentialType = repository.UserPassCredential
			spec.Credentials.UserPass = &models.RepoUserPassCredentials{}
			spec.Credentials.UserPass.Username = r.Spec.Credentials.UserPass.Username
			spec.Credentials.UserPass.Password = r.Spec.Credentials.UserPass.Password
		default:
			log.Println("invalid repository.CredentialType, must one of", r.Spec.CredentialType, strings.Join(repository.AllowedCrentialTypes, "|"))
		}
		spec.AgentNames = r.Spec.Agents
		err = repository.UpdateRepository(r.Metadata.Name, projectId, spec, createIfNotPresent)
		if err != nil {
			log.Println("Failed to update Repository: ", r.Metadata.Name)
		} else {
			log.Printf("Successfully created/updated Repository: %s", r.Metadata.Name)
		}
	}
	return diags
}

func resourceRepositoriesDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("deleting repositories")

	//get project id with project name, p.id used to refer to project id -> need p.ID for calling DeleteRepository
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
	//delete repository
	err = repository.DeleteRepository(d.Id(), projectId)
	if err != nil {
		log.Println("error deleting agent")
	} else {
		log.Println("Deleted agent: ", d.Id())
	}
	return diags
}
