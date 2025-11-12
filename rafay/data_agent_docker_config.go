package rafay

import (
	"context"
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"time"

	"github.com/RafaySystems/rctl/pkg/config"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataAgentDockerConfig() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataAgentDockerConfigRead,
		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"project": {
				Type:     schema.TypeString,
				Required: true,
			},
			"agent_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"download_config_files": {
				Type:     schema.TypeBool,
				Required: true,
			},
			"download_directory": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"docker_compose": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"config": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"agent_id_hash": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"start_command": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"stop_command": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"compose_file_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"config_file_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataAgentDockerConfigRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var err error
	log.Printf("download docker agent config files start")
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		_ = context.WithValue(ctx, "debug", "true")
	}

	projectName, ok := d.Get("project").(string)
	if !ok || projectName == "" {
		log.Print("cannot find project name")
		return diag.FromErr(fmt.Errorf("%s", "project name is missing"))
	}

	projectId, err := getProjectIDFromName(projectName)
	if err != nil {
		log.Print("error converting project name to id")
		return diag.FromErr(err)
	}

	agentName, ok := d.Get("agent_name").(string)
	if !ok || agentName == "" {
		log.Print("Cannot find agent name")
		return diag.FromErr(fmt.Errorf("%s", "agent name is missing"))
	}

	dockerCompose, err := getDockerCompose(projectId, agentName)
	if err != nil {
		if strings.Contains(err.Error(), "401") && strings.Contains(err.Error(), "agent is not in the scope of this project") {
			return diag.FromErr(fmt.Errorf("agent %s does not exist", agentName))
		}
		return diag.FromErr(err)
	}

	relayConfig, err := getRelayConfig(projectId, agentName)
	if err != nil {
		return diag.FromErr(err)
	}

	agentId, err := getAgentId(projectId, agentName)
	if err != nil {
		return diag.FromErr(err)
	}

	downloadDirectory := "./"
	if d.Get("download_directory").(string) != "" {
		downloadDirectory = d.Get("download_directory").(string)
	}

	dockerComposeFileName := fmt.Sprintf("docker-compose-%s.yaml", agentId)
	relayConfigFileName := fmt.Sprintf("relayConfigData-%s.json", agentId)

	dockerComposeFilePath := path.Join(downloadDirectory, dockerComposeFileName)
	relayConfigFilePath := path.Join(downloadDirectory, relayConfigFileName)

	if d.Get("download_config_files").(bool) {
		if _, err := os.Stat(downloadDirectory); os.IsNotExist(err) {
			err = os.MkdirAll(downloadDirectory, 0755)
			if err != nil {
				return diag.FromErr(err)
			}
		}

		err = os.WriteFile(dockerComposeFilePath, []byte(dockerCompose), 0644)
		if err != nil {
			return diag.FromErr(err)
		}

		err = os.WriteFile(relayConfigFilePath, []byte(relayConfig), 0644)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if err := d.Set("docker_compose", dockerCompose); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("config", relayConfig); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("agent_id_hash", agentId); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("start_command", fmt.Sprintf("docker compose -f %s up -d", dockerComposeFilePath)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("stop_command", fmt.Sprintf("docker compose -f %s down", dockerComposeFilePath)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("compose_file_name", dockerComposeFileName); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("config_file_name", relayConfigFileName); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(fmt.Sprintf("%s-%s", projectName, agentName))

	return diags
}

func getAgentId(projectId, agentName string) (string, error) {
	auth := config.GetConfig().GetAppAuthProfile()

	uri := fmt.Sprintf("/v2/config/project/%s/agent/%s", projectId, agentName)
	resp, err := auth.AuthAndRequestFullResponse(uri, "GET", nil)
	if err != nil {
		log.Printf("failed to get agent id")
		return "", err
	}

	defer func() { _ = resp.Close() }()

	if resp.StatusCode != 200 {
		log.Printf("failed to get agent id")
		return "", err
	}

	var agent map[string]interface{}
	err = resp.JSON(&agent)
	if err != nil {
		log.Printf("failed to serialize agent")
		return "", err
	}

	id := agent["metadata"].(map[string]interface{})["id"].(string)
	if id == "" {
		log.Printf("failed to get agent id from agent object")
		return "", err
	}

	return id, nil
}

func getDockerCompose(projectId, agentName string) (string, error) {
	auth := config.GetConfig().GetAppAuthProfile()

	uri := fmt.Sprintf("/v2/config/project/%s/agent/%s/docker-compose", projectId, agentName)
	resp, err := auth.AuthAndRequestFullResponse(uri, "GET", nil)
	if err != nil {
		log.Printf("failed to get docker compose")
		return "", err
	}

	defer func() { _ = resp.Close() }()

	if resp.StatusCode != 200 {
		log.Printf("failed to get docker compose")
		return "", err
	}

	return string(resp.String()), nil
}

func getRelayConfig(projectId, agentName string) (string, error) {
	auth := config.GetConfig().GetAppAuthProfile()

	uri := fmt.Sprintf("/v2/config/project/%s/agent/%s/relay-config", projectId, agentName)
	resp, err := auth.AuthAndRequestFullResponse(uri, "GET", nil)
	if err != nil {
		log.Printf("failed to get relay config")
		return "", err
	}

	defer func() { _ = resp.Close() }()

	if resp.StatusCode != 200 {
		log.Printf("failed to get relay config")
		return "", err
	}

	return string(resp.String()), nil
}
