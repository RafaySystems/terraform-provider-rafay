package rafay

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/RafaySystems/rctl/pkg/models"
	"github.com/RafaySystems/rctl/pkg/namespace"
	"github.com/RafaySystems/rctl/pkg/project"
	"github.com/RafaySystems/rctl/utils"

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

//why is this not importable from rctl?
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
		nst, err = convertNamespaceYAMLToModel(&n, nstype, filepath, filePath)
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
	return diags
}
func resourceNamespaceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	return diags
}
func resourceNamespaceDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	return diags
}

func convertNamespaceYAMLToModel(nsYaml *namespaceYamlConfig, nstype string, filepath string, path string) (*models.Namespace, error) {
	var ns models.Namespace
	ns.Metadata.Name = nsYaml.Metadata.Name
	ns.Kind = nsYaml.Kind
	ns.Metadata.Labels = nsYaml.Metadata.Labels
	ns.Spec.Type = nsYaml.Spec.Type
	ns.Metadata.Annotations = nsYaml.Metadata.Annotations
	ns.Spec.ResourceQuota = nsYaml.Spec.ResourceQuota
	ns.Spec.LimitRange = nsYaml.Spec.LimitRange
	ns.Spec.Placement = nsYaml.Spec.Placement
	ns.Spec.PSP = nsYaml.Spec.PSP
	ns.Metadata.Description = nsYaml.Metadata.Description

	if nsYaml.Spec.RepositoryRef != "" && nsYaml.Spec.NamespaceFromFile != "" {
		log.Println("invalid config: both repo and file were provided")
	}
	ns.Spec.RepositoryRef = nsYaml.Spec.RepositoryRef
	ns.Spec.RepoArtifactMeta = nsYaml.Spec.RepoArtifactMeta
	if nsYaml.Spec.NamespaceFromFile != "" {
		nsFileContent, err := getNamespaceFromFile(filepath, path)
		if err != nil {
			return nil, fmt.Errorf("invalid config: error fetching the content of the value file from the location provided %s: Error: %s", path, err.Error())
		}
		ns.Spec.NamespaceFromFile = nsFileContent
	}
	if ns.Spec.RepositoryRef != "" && (ns.Spec.RepoArtifactMeta.Git == nil || len(ns.Spec.RepoArtifactMeta.Git.RepoArtifactFiles) == 0) {
		return nil, fmt.Errorf("invalid config: exactly one repo artifact file should be provided.\"")
	}

	return &ns, nil
}

func getNamespaceFromFile(filepath string, path string) (string, error) {
	{
		nsFileLocation := utils.FullPath(path, filepath)
		if _, err := os.Stat(nsFileLocation); os.IsNotExist(err) {
			return "", fmt.Errorf("namespace file doesn't exist '%s'", nsFileLocation)
		}
		nsFileContent, err := ioutil.ReadFile(nsFileLocation)
		if err != nil {
			return "", fmt.Errorf("error in reading the namespace file %s: %s\n", nsFileLocation, err)
		}
		return string(nsFileContent), nil

	}
}
