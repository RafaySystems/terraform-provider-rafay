package rafay

import (
	"context"
	"fmt"
	"log"
	"os"
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
			"docker_compose": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"relay_config": {
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
		return diag.FromErr(err)
	}

	relayConfig, err := getRelayConfig(projectId, agentName)
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("docker_compose", dockerCompose)
	d.Set("relay_config", relayConfig)
	d.SetId(fmt.Sprintf("%s-%s", projectName, agentName))

	return diags
}

func getDockerCompose(projectId, agentName string) (string, error) {
	auth := config.GetConfig().GetAppAuthProfile()

	uri := fmt.Sprintf("/v2/config/project/%s/agent/%s/docker-compose", projectId, agentName)
	resp, err := auth.AuthAndRequestFullResponse(uri, "GET", nil)
	if err != nil {
		log.Printf("failed to get docker compose")
		return "", err
	}

	defer resp.Close()

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

	defer resp.Close()

	if resp.StatusCode != 200 {
		log.Printf("failed to get relay config")
		return "", err
	}

	return string(resp.String()), nil
}
