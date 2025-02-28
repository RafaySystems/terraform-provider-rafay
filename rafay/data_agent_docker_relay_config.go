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

func dataAgentDockerRelayConfig() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataDockerRelayConfigRead,
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
			"content": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataDockerRelayConfigRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var err error
	log.Printf("download docker agent relay config file starts")
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		_ = context.WithValue(ctx, "debug", "true")
	}
	auth := config.GetConfig().GetAppAuthProfile()

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

	uri := fmt.Sprintf("/v2/config/project/%s/agent/%s/relay-config", projectId, agentName)

	resp, err := auth.AuthAndRequestFullResponse(uri, "GET", nil)
	if err != nil {
		log.Printf("failed to get docker agent relay config")
		return diag.FromErr(err)
	}

	defer resp.Close()

	if resp.StatusCode != 200 {
		log.Printf("failed to get docker agent relay config")
		return diag.FromErr(fmt.Errorf("failed to get docker agent relay config"))
	}

	relayConfig := string(resp.String())

	d.Set("content", relayConfig)
	d.SetId(fmt.Sprintf("%s-%s", projectName, agentName))

	return diags
}
