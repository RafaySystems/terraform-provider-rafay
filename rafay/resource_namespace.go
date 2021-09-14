package rafay

import (
	"context"
	"io/ioutil"
	"log"
	"os"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/RafaySystems/rctl/pkg/commands"
	"github.com/RafaySystems/rctl/pkg/models"
	"github.com/RafaySystems/rctl/pkg/namespace"
	"github.com/RafaySystems/rctl/pkg/project"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceNamespace() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceNamespaceCreate,
		ReadContext:   resourceNamespaceRead,
		UpdateContext: resourceNamespaceUpdate,
		DeleteContext: resourceNamespaceDelete,

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
			"namespace_filepath": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

//remove and make proper changes in rctl (make sure to point to right version release)
type namespaceYamlConfig struct {
	Kind     string `yaml:"kind"`
	Metadata struct {
		Name        string            `yaml:"name"`
		Labels      map[string]string `yaml:"labels"`
		Annotations map[string]string `yaml:"annotations"`
		Description string            `yaml:"description"`
	} `yaml:"metadata"`
	Spec struct {
		Type              string                  `yaml:"type"`
		ResourceQuota     map[string]interface{}  `yaml:"resourceQuota"`
		LimitRange        map[string]interface{}  `yaml:"limitRange"`
		Placement         map[string]interface{}  `yaml:"placement"`
		PSP               string                  `yaml:"psp"`
		RepositoryRef     string                  `protobuf:"bytes,4,opt,name=repositoryRef,proto3" json:"repoRef,omitempty" yaml:"repoRef"`
		NamespaceFromFile string                  `yaml:"namespaceFromFile"`
		RepoArtifactMeta  models.RepoArtifactMeta `protobuf:"bytes,4,opt,name=repoArtifactMeta,proto3" yaml:"repoArtifactMeta,omitempty"`
	} `yaml:"spec"`
}

func resourceNamespaceCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	filePath := d.Get("namespace_filepath").(string)
	var n namespaceYamlConfig
	var nst *models.Namespace
	var nstype string
	var filepath string
	//make sure this is the correct file path
	log.Println("filepath: ", filePath)
	log.Printf("create namespace resource")
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
		err = yaml.Unmarshal(c, &n)
		if err != nil {
			log.Println("error unmarhsalling data")
		}
		nstype = n.Spec.Type

		filepath = n.Spec.NamespaceFromFile
		nst, err = commands.ConvertNamespaceYAMLToModel(&n, nstype, filepath, filePath)
		if err != nil {
			log.Printf("Failed to create namespace:%s\n", n.Metadata.Name)
		}
		err = namespace.CreateNamespace(nst, projectId)
		if err != nil {
			log.Printf("Failed to create namespace:%s\n", n.Metadata.Name)
		} else {
			log.Printf("Successfully created namespace:%s\n", n.Metadata.Name)
		}
	}
	//Set metadataname as id
	d.SetId(n.Metadata.Name)
	return diags
}

func resourceNamespaceUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var n commands.NamespaceYamlConfig
	var nstype string
	var filepath string
	filePath := d.Get("namespace_filepath").(string)
	//make sure this is the correct file path
	log.Println("filepath: ", filePath)
	log.Printf("update namespace resource")
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
		err = yaml.Unmarshal(c, &n)
		if err != nil {
			log.Println("error unmarhsalling data")
		}
		nstype = n.Spec.Type

		filepath = n.Spec.NamespaceFromFile
		nst, err := commands.ConvertNamespaceYAMLToModel(&n, nstype, filepath, filePath)
		if err != nil {
			log.Printf("Failed to create namespace:\n", n.Metadata.Name)
		}
		// updaqte the namespace
		err = namespace.UpdateNamespace(nst, d.Id(), projectId)
		if err != nil {
			log.Printf("Failed to update namespace:\n", n.Metadata.Name, err)
		} else {
			log.Printf("Successfully updated namespace:\n", n.Metadata.Name)
		}
	}
	return diags
}
func resourceNamespaceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	return diags
}
func resourceNamespaceDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
	err = namespace.DeleteNamespace(string(d.Id()), projectId)
	if err != nil {
		log.Println("error delete namespace: ", err)
	}
	return diags
}
