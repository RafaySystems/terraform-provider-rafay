package rafay

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"
	"net/url"
	"strings"



	"github.com/RafaySystems/rctl/pkg/config"
	"github.com/RafaySystems/rctl/pkg/models"
	"github.com/RafaySystems/rctl/pkg/workload"

	"github.com/RafaySystems/rctl/utils"
	"gopkg.in/yaml.v3"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type Workload struct {
	Name             string           `json:"name,omitempty" yaml:"name,omitempty"`
	Description      string           `json:"description,omitempty" yaml:"description,omitempty"`
	Namespace        string           `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	Project          string           `json:"project,omitempty" yaml:"project,omitempty"`
	Type             string           `json:"type,omitempty" yaml:"type,omitempty"`
	Clusters         string           `json:"clusters,omitempty" yaml:"clusters,omitempty"`
	Locations        string           `json:"cluster_locations,omitempty" yaml:"cluster_locations,omitempty"`
	Labels           string           `json:"cluster_label_selectors,omitempty" yaml:"cluster_label_selectors,omitempty"`
	PayloadFile      string           `json:"payload,omitempty" yaml:"payload,omitempty"`
	ValuesFile       string           `json:"values,omitempty" yaml:"values,omitempty"`
	RepositoryRef    string           `json:"repository_ref,omitempty" yaml:"repository_ref,omitempty"`
	RepoArtifactMeta RepoArtifactMeta `json:"repo_artifact_meta,omitempty" yaml:"repo_artifact_meta,omitempty"`
	DriftAction      string           `json:"driftaction,omitempty" yaml:"driftaction,omitempty"`
}

func resourceWorkload() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceWorkloadCreate,
		ReadContext:   resourceWorkloadRead,
		UpdateContext: resourceWorkloadUpdate,
		DeleteContext: resourceWorkloadDelete,

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
			"projectname": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func validateValuesFiles(path string) error {
	if path == "" {
		return nil
	}
	allValuesFiles := strings.Split(path, ",")
	for _, valuesFileFullPath := range allValuesFiles {
		if _, err := os.Stat(valuesFileFullPath); os.IsNotExist(err) {
			return fmt.Errorf("Values file doesn't exist '%s'", valuesFileFullPath)
		}
	}
	return nil
}

func resourceWorkloadCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("works here")

	YamlConfigFilePath := d.Get("yamlfilepath").(string)
	log.Printf("YamlConfigFile %s", YamlConfigFilePath)

	if !utils.FileExists(YamlConfigFilePath) {
		log.Printf("file %s not exist", YamlConfigFilePath)
		return diags
	}
	if filepath.Ext(YamlConfigFilePath) != ".yml" && filepath.Ext(YamlConfigFilePath) != ".yaml" {
		log.Printf("file must a yaml file, file type is %s", filepath.Ext(YamlConfigFilePath))
		return diags
	}
	f, err := os.Open(YamlConfigFilePath)
	if err != nil {
		log.Printf("Error while open Yaml %s", YamlConfigFilePath)
		return diag.FromErr(err)
	}

	c, err := ioutil.ReadAll(f)
	if err != nil {
		log.Printf("error while Reading file")
		return diag.FromErr(err)
	}
	file := string(c)
	var wl models.Workload
	err = yaml.Unmarshal(c, wl)
	if err != nil {
		log.Printf("error while unmarshal Yaml file ")
		return diag.FromErr(err)
	}
	auth := config.GetConfig().GetAppAuthProfile()
	payloadFile := wl.PayloadFile
	projectId, _ := config.GetProjectIdByName(wl.Project)
	if err != nil {
		log.Printf("Project by name '%s' is not present (or you don't have access to it yet). Please use valid project in meta file.", wl.Project)
		return diag.FromErr(err)
	}
	if projectId == "" {
		projectId = config.GetConfig().ProjectID
	}
	if projectId == "" {
		log.Printf("project context couldn't be determined. Please use --project argument or init rctl with the project context using \"rctl config set project <project name>\"")
		return diag.FromErr(err)
	}
	wl.Project = projectId

	// 
	switch wl.Type {
	case "", "Rafay":
		//This is a deprecated method of using the CLI to support backward compatible commands - using the input file as payload for V1 workload
		if payloadFile == "" {
			log.Printf("payloadfile is not part of the input file. This is a deprecated method of using this command. Assuming it is a Wizard workload payload and proceeding...")
			payloadFile = file
		} else {
			//Verify that the payloadFile pointed in the metafile exists considering its fullPath relative to input meta file
			payloadFile = utils.FullPath(file, payloadFile)
			if _, err := os.Stat(payloadFile); os.IsNotExist(err) {
				log.Printf("payload file doesn't exist '%s'", payloadFile)
				return diag.FromErr(err)

			}
		}
		params := url.Values{}
		params.Add("cli_call", "true")
		uri := fmt.Sprintf("/config/v1/projects/%s/workloads/?", projectId) + params.Encode()
		resp, err := auth.PostRequestFromFile(uri, payloadFile)
		if err != nil {
			log.Printf("failed to create workload from %s, resp %s", payloadFile, resp)
			return diag.FromErr(err)

		}
		log.Printf("End %s resp %s", file, resp)
	case "Helm", "NativeYaml", "Helm3":
		valuesFileFullPath := utils.FullPaths(file, wl.ValuesFile)
		// if it's git repo or helm repo
		if wl.RepositoryRef != "" {
			log.Printf("Creating helmInGitRepo, YamlInGitRepo or HelmInHelmRepo workload {%v}", wl)
			workload.CreateWorkloadWithRepo(&wl, valuesFileFullPath)
			return nil
		}

		if payloadFile == "" {
			log.Printf("payload file information doesn't exist. A field by the name 'payload' is required which points to the payload location")
			return diag.FromErr(err)
		}
		payloadFileFullPath := utils.FullPath(file, payloadFile)
		log.Printf("payloadFile, %s, valuesFile %s", payloadFileFullPath, valuesFileFullPath)
		if _, err := os.Stat(payloadFileFullPath); os.IsNotExist(err) {
			log.Printf("Payload file doesn't exist '%s'", payloadFileFullPath)
			return diag.FromErr(err)
		}
		err = validateValuesFiles(valuesFileFullPath)
		if err != nil {
			return diag.FromErr(err)
		}
		log.Printf("Creating helm or native yaml workload {%v}", wl)
		err := workload.CreateV2Workload(&wl, payloadFileFullPath, valuesFileFullPath)
		if err != nil {
			log.Printf("Workload create failed")
			return diag.FromErr(err)
		}
	default:
		log.Printf("unsupported workload type %s. Supported types Helm, NativeYaml, Helm3, (Wizard workloads if empty), ", wl.Type)
		return diag.FromErr(err)
	}

	log.Printf("Workload created successfully")
	return diags
}

func resourceWorkloadRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
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

	//get workload
	wl := d 
	wlid := d.ID
	wltype, err := workload.getWorkloadType(wl)
	if err != nil {
		return fmt.Errorf("failed to determine the workload id")
	}
	shouldShowPublished, err := cmd.Flags().GetBool("published")
	actionEndPoint := "current_view"
	if shouldShowPublished {
		actionEndPoint = "published_view"
	}
	uri := fmt.Sprintf("/config/v1/projects/%s/workloads/%s/%s/?", projectId, wlid, actionEndPoint)
	auth := config.GetConfig().GetAppAuthProfile()
	resp, err := auth.AuthAndRequest(uri, "GET", nil)
	workloadNotPublishedErrorMessage := "Workload is not published yet."
	if err != nil {
		if wltype != "NativeHelm" {
			if strings.Contains(err.Error(), authprofile.ResourceNotExists.Error()) {
				fmt.Println(workloadNotPublishedErrorMessage)
				return nil
			}
			log.Printf("failed to get workload components")
			return diag.FromErr(err)

		}
		//XXX TODO when backend supports current_view/published_view apis for helm3
		return nil
	}
	resp = strings.TrimSpace(resp)
	if resp == "" {
		if shouldShowPublished {
			fmt.Println(workloadNotPublishedErrorMessage)
		} else {
			fmt.Println("Workload is not published yet. To get view please publish workload.")
		}
	} else {
		fmt.Println(resp)
	}
	log.Printf("End [%s %s]", cmd.CommandPath(), wl)
	return diags
}

func resourceWorkloadUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	//TODO implement update workload
	var diags diag.Diagnostics
	//log.Printf("resource workload update id %s", d.Id())

	return diags
}

func resourceWorkloadDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	//get project id with project name, p.id
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
	wl := d

	wlid = metadata.name
	if err != nil {
		return log.Printf("Failed to determine the workload id")
	}
	auth := config.GetConfig().GetAppAuthProfile()
	uri := fmt.Sprintf("/config/v1/projects/%s/workloads/%s/", projectId, wlid)
	_, err = auth.AuthAndRequest(uri, "DELETE", "")
	if err != nil {
		return log.Printf("Failed to delete workload %s", args[0] //workload)
	}
	log.Printf("End %s", wl)
	return diags
}
