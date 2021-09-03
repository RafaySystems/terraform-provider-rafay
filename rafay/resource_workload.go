package rafay

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/RafaySystems/rctl/pkg/config"
	"github.com/RafaySystems/rctl/pkg/project"
	"github.com/RafaySystems/rctl/pkg/workload"

	"github.com/RafaySystems/rctl/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

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
			"workloadname": {
				Type:     schema.TypeString,
				Optional: true,
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
			return fmt.Errorf("values file doesn't exist '%s'", valuesFileFullPath)
		}
	}
	return nil
}

func resourceWorkloadCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("workload create starts here")

	YamlConfigFilePath := d.Get("yamlfilepath").(string)
	log.Printf("YamlConfigFile %s", YamlConfigFilePath)

	wl, file, err := workload.GetWorkload(YamlConfigFilePath)
	if err != nil {
		log.Printf("Get workload command call failed")
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
		valuesFileFullPath := utils.FullPaths(YamlConfigFilePath, wl.ValuesFile)
		// if it's git repo or helm repo
		if wl.RepositoryRef != "" {
			log.Printf("Creating helmInGitRepo, YamlInGitRepo or HelmInHelmRepo workload {%v}", wl)
			workload.CreateWorkloadWithRepo(wl, valuesFileFullPath)
			return nil
		}

		if payloadFile == "" {
			log.Printf("payload file information doesn't exist. A field by the name 'payload' is required which points to the payload location")
			return diag.FromErr(err)
		}

		payloadFileFullPath := utils.FullPath(YamlConfigFilePath, payloadFile)
		if _, err := os.Stat(payloadFileFullPath); os.IsNotExist(err) {
			log.Printf("Payload file doesn't exist '%s'", payloadFileFullPath)
			return diag.FromErr(err)
		}
		err = validateValuesFiles(valuesFileFullPath)
		if err != nil {
			return diag.FromErr(err)
		}
		log.Printf("Creating helm or native yaml workload {%v}", wl)
		err := workload.CreateV2Workload(wl, payloadFileFullPath, valuesFileFullPath)
		//err := createLocalWorkload(wl, payloadFileFullPath, valuesFileFullPath)
		if err != nil {
			log.Printf("v2Workload create failed")
			return diag.FromErr(err)
		}
	default:
		log.Printf("unsupported workload type %s. Supported types Helm, NativeYaml, Helm3, (Wizard workloads if empty), ", wl.Type)
		return diag.FromErr(err)
	}

	log.Printf("Workload created successfully")

	//get workloadId
	wlid, err := workload.GetWorkloadId(wl.Name, projectId)
	if err != nil {
		log.Printf("Get workloadid failed")
		return diag.FromErr(err)
	}

	//publish workload
	uri := fmt.Sprintf("/config/v1/projects/%s/workloads/%s/publish/", projectId, wlid)
	_, err = auth.AuthAndRequest(uri, "POST", "")
	if err != nil {
		log.Printf("failed to publish workload %s", wlid)
		return diag.FromErr(err)
	}

	d.SetId(wlid)
	return diags
}

func resourceWorkloadRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	//find projectid
	resp, err := project.GetProjectByName(d.Get("projectname").(string))
	if err != nil {
		log.Printf("project does not exist, error %s", err.Error())
		return diag.FromErr(err)
	}
	p, err := project.NewProjectFromResponse([]byte(resp))
	if err != nil {
		return diag.FromErr(err)
	} else if p == nil {
		return diags
	}
	projectId := p.ID

	//call get
	auth := config.GetConfig().GetAppAuthProfile()
	uri := fmt.Sprintf("/config/v1/projects/%s/workloads/%s/?", projectId, d.Id())
	resp, err = auth.AuthAndRequest(uri, "GET", nil)
	if err != nil {
		log.Printf("Failed to get workload %s", d.Id())
	}
	return diags

}

func resourceWorkloadUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	//TODO implement update workload
	var diags diag.Diagnostics

	log.Printf("Two arguments were passed. This is a deprecated method of using the command (workload update name file) allowed by legacy RCTL version")
	YamlConfigFilePath := d.Get("yamlfilepath").(string)
	log.Printf("YamlConfigFile %s", YamlConfigFilePath)

	wl, file, err := workload.GetWorkload(YamlConfigFilePath)
	if err != nil {
		log.Printf("Get workload command call failed")
		return diag.FromErr(err)
	}
	if file == "" {
		log.Printf("Get file command call failed")
		return diag.FromErr(err)
	}

	if wl.PayloadFile == "" && wl.ValuesFile == "" && wl.Clusters == "" && wl.Labels == "" && wl.Locations == "" {
		log.Printf("workload values empty")
		return diag.FromErr(err)
	}

	if wl.Name == "" {
		log.Printf("invalid input. Workload name can't be determined. Please check the input file")
		return diags
	}

	//projectid
	resp, err := project.GetProjectByName(d.Get("projectname").(string))
	if err != nil {
		log.Printf("project does not exist, error %s", err.Error())
		return diag.FromErr(err)
	}
	p, err := project.NewProjectFromResponse([]byte(resp))
	if err != nil {
		return diag.FromErr(err)
	} else if p == nil {
		return diags
	}
	projectId := p.ID

	//Project ID input order: Flag/Arguments > Meta File > Config File
	if projectId == "" && wl.Project != "" {
		projectId, err = config.GetProjectIdByName(wl.Project)
		if err != nil {
			log.Printf("project by name '%s' is not present (or you don't have access to it yet). Please use valid project in meta file", wl.Project)
			return diag.FromErr(err)
		}
	}
	if projectId == "" {
		projectId = config.GetConfig().ProjectID
	}
	if projectId == "" {
		log.Printf("project context couldn't be determined. Please use --project argument or init rctl with the project context using \"rctl config set project <project name>\"")
		return diag.FromErr(err)
	}
	wlid := d.Id()
	wl.Project = projectId
	switch wl.Type {
	case "", "Rafay":
		payloadFile := wl.PayloadFile
		//Using meta file pointing to the actual payload, verify that the file exists considering relative path of the input meta file
		if d.Get("workloadname") != nil {
			payloadFile = utils.FullPath(YamlConfigFilePath, payloadFile)
			if _, err := os.Stat(payloadFile); os.IsNotExist(err) {
				log.Printf("inside blank and rafay")
				log.Printf("Payload file doesn't exist '%s'", payloadFile)
				return diag.FromErr(err)
			}
		}
		params := url.Values{}
		params.Add("cli_call", "true")
		uri := fmt.Sprintf("/config/v1/projects/%s/workloads/%s/?%s", projectId, wlid, params.Encode())
		auth := config.GetConfig().GetAppAuthProfile()

		resp, err = auth.RequestFromFile(uri, "PUT", payloadFile)
		if err != nil {
			log.Printf("Failed to update workload %s from %s", wl.Name, payloadFile)
			return diag.FromErr(err)
		}

	case "Helm", "NativeYaml", "Helm3":
		if wl.RepositoryRef != "" {
			log.Printf("Creating helmInGitRepo, YamlInGitRepo or HelmInHelmRepo workload {%v}", wl)
			workload.UpdateWorkloadWithRepo(wl, wlid)
			return diag.FromErr(err)
		}
		wlType := wl.Type
		payloadFile := utils.FullPath(YamlConfigFilePath, wl.PayloadFile)
		log.Printf("payloadfile %s", payloadFile)
		log.Printf("wl payloadfile %s", wl.PayloadFile)

		if _, err := os.Stat(payloadFile); wl.PayloadFile != "" && os.IsNotExist(err) {
			log.Printf("inside helm/nativeyaml and hel3")
			log.Printf("Payload file doesn't exist '%s'", payloadFile)
			return diag.FromErr(err)
		}
		valuesFile := utils.FullPaths(YamlConfigFilePath, wl.ValuesFile)
		err = validateValuesFiles(valuesFile)
		if err != nil {
			return diag.FromErr(err)
		}
		if wlType == "Helm3" {
			wlType = "NativeHelm"
		}
		workload.UpdateV2Workload(nil, wlid, projectId, wlType, payloadFile, valuesFile, wl)

	}
	//publish workload
	auth := config.GetConfig().GetAppAuthProfile()
	uri := fmt.Sprintf("/config/v1/projects/%s/workloads/%s/publish/", projectId, wlid)
	_, err = auth.AuthAndRequest(uri, "POST", "")
	if err != nil {
		log.Printf("failed to publish workload %s", wlid)
		return diag.FromErr(err)
	}

	return diags
}

func resourceWorkloadDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	//find projectid
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

	//call delete
	auth := config.GetConfig().GetAppAuthProfile()
	uri := fmt.Sprintf("/config/v1/projects/%s/workloads/%s/", projectId, d.Id())
	_, err = auth.AuthAndRequest(uri, "DELETE", "")
	if err != nil {
		log.Printf("Failed to delete workload %s", d.Id())
	}

	return diags
}
