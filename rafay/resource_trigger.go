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
	"github.com/RafaySystems/rctl/pkg/trigger"
	"github.com/RafaySystems/rctl/utils"
	"gopkg.in/yaml.v3"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTrigger() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceTriggerCreate,
		ReadContext:   resourceTriggerRead,
		UpdateContext: resourceTriggerUpdate,
		DeleteContext: resourceTriggerDelete,

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
			"trigger_filepath": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceTriggerCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	filePath := d.Get("trigger_filepath").(string)
	var t commands.TriggerYamlConfig
	//make sure this is the correct file path
	log.Println("filepath: ", filePath)
	log.Printf("create gitops Trigger resource")
	//get project id with project name, p.id used to refer to project id -> need p.ID for calling createTrigger
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
	//open file path and retirve config spec from yaml file (from run function in commands/create_agent.go)
	//read and capture file from file path
	if f, err := os.Open(filePath); err == nil {
		c, err := ioutil.ReadAll(f)
		if err != nil {
			log.Println("error cpaturing file")
		}
		//implement createClusterOverride from commands/create_agent.go -> then call clusteroverride.Createagent
		triggerDefinition := c
		// unmarshal the data
		err = yaml.Unmarshal(triggerDefinition, &t)
		if err != nil {
			log.Println("unmarshalling error")
		}

		// check if project is provided
		projId := projectId
		if t.Metadata.Project != "" {
			projId, err = config.GetProjectIdByName(t.Metadata.Project)
			if err != nil {
				log.Println("get project by id name error")
			}
		}

		for _, v := range trigger.AllowedTypes {
			if utils.StringsAreEqualCaseInsensitive(t.Spec.TriggerType, v) {
				t.Spec.TriggerType = v
			}
		}
		spec := models.TriggerSpec{}
		spec.PipelineRef = t.Spec.PipelineRef
		spec.RepositoryRef = t.Spec.RepositoryRef
		spec.TriggerConfig = &models.TriggerConfig{}
		switch t.Spec.TriggerType {
		case trigger.WebhookTrigger:
			spec.TriggerType = trigger.WebhookTrigger
			spec.TriggerConfig.Webhook = &models.WebhookTriggerConfig{}
			spec.TriggerConfig.Webhook.ConfigType = models.WebhookTriggerConfigType(t.Spec.TriggerConfig.Webhook.ConfigType)
			spec.TriggerConfig.Webhook.PayloadType = models.WebhookPayloadType(t.Spec.TriggerConfig.Webhook.PayloadType)
		case trigger.PeriodicTrigger:
			spec.TriggerType = trigger.PeriodicTrigger
			spec.TriggerConfig.Periodic = &models.PeriodicTriggerConfig{}
			spec.TriggerConfig.Periodic.Periodicity = t.Spec.TriggerConfig.Periodic.Periodicity
			spec.TriggerConfig.Periodic.CronExpression = t.Spec.TriggerConfig.Periodic.CronExpression
		case trigger.PeriodicSCMTrigger:
			spec.TriggerType = trigger.PeriodicSCMTrigger
			spec.TriggerConfig.PeriodicSCM = &models.PeriodicSCMTriggerConfig{}
			spec.TriggerConfig.PeriodicSCM.Periodicity = t.Spec.TriggerConfig.PeriodicSCM.Periodicity
			spec.TriggerConfig.PeriodicSCM.CronExpression = t.Spec.TriggerConfig.PeriodicSCM.CronExpression
		default:
			log.Println("invalid triggerType, must one of", t.Spec.TriggerType, strings.Join(trigger.AllowedTypes, "|"))
		}

		if t.Spec.RepositoryConfig.Git.Revision != "" {
			spec.RepositoryConfig.Git = &models.GitRepositoryConfig{}
			spec.RepositoryConfig.Git.Revision = t.Spec.RepositoryConfig.Git.Revision
			spec.RepositoryConfig.Git.Paths = t.Spec.RepositoryConfig.Git.Paths
		} else if t.Spec.RepositoryConfig.Helm.ChartName != "" {
			spec.RepositoryConfig.Helm = &models.HelmRepositoryConfig{}
			spec.RepositoryConfig.Helm.ChartName = t.Spec.RepositoryConfig.Helm.ChartName
			spec.RepositoryConfig.Helm.Revision = t.Spec.RepositoryConfig.Helm.Revision
		} else {
			log.Println("Required respositoryConfig.git or respositoryConfig.helm")
		}

		for _, v := range t.Spec.Variables {
			spec.Variables = append(spec.Variables, models.InlineVariableSpec{
				Name:  v.Name,
				Type:  models.VariableType(v.Type),
				Value: v.Value,
			})
		}
		//create trigger
		err = trigger.CreateTrigger(t.Metadata.Name, projId, spec)
		if err != nil {
			log.Println("Failed to create Trigger: \n", t.Metadata.Name, "err: ", err)
		} else {
			log.Println("Successfully created trigger: \n", t.Metadata.Name)
		}
	}
	//set id to metadata.Name
	d.SetId(t.Metadata.Name)
	return diags
}

func resourceTriggerRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	return diags
}

func resourceTriggerUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	createIfNotPresent := false
	log.Println("update trigger")
	filePath := d.Get("trigger_filepath").(string)
	var t commands.TriggerYamlConfig
	//make sure this is the correct file path
	//make sure this is the correct file path
	log.Println("filepath: ", filePath)
	log.Printf("create gitops trigger resource")
	//get project id with project name, p.id used to refer to project id -> need p.ID for calling createAgent
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
		triggerDefinition := c
		// unmarshal the data
		err = yaml.Unmarshal(triggerDefinition, &t)
		if err != nil {
			log.Println("unmarshalling error")
		}

		for _, v := range trigger.AllowedTypes {
			if utils.StringsAreEqualCaseInsensitive(t.Spec.TriggerType, v) {
				t.Spec.TriggerType = v
			}
		}
		spec := models.TriggerSpec{}
		spec.PipelineRef = t.Spec.PipelineRef
		spec.RepositoryRef = t.Spec.RepositoryRef
		spec.TriggerConfig = &models.TriggerConfig{}
		switch t.Spec.TriggerType {
		case trigger.WebhookTrigger:
			spec.TriggerType = trigger.WebhookTrigger
			spec.TriggerConfig.Webhook = &models.WebhookTriggerConfig{}
			spec.TriggerConfig.Webhook.ConfigType = models.WebhookTriggerConfigType(t.Spec.TriggerConfig.Webhook.ConfigType)
			spec.TriggerConfig.Webhook.PayloadType = models.WebhookPayloadType(t.Spec.TriggerConfig.Webhook.PayloadType)
		case trigger.PeriodicTrigger:
			spec.TriggerType = trigger.PeriodicTrigger
			spec.TriggerConfig.Periodic = &models.PeriodicTriggerConfig{}
			spec.TriggerConfig.Periodic.Periodicity = t.Spec.TriggerConfig.Periodic.Periodicity
			spec.TriggerConfig.Periodic.CronExpression = t.Spec.TriggerConfig.Periodic.CronExpression
		case trigger.PeriodicSCMTrigger:
			spec.TriggerType = trigger.PeriodicSCMTrigger
			spec.TriggerConfig.PeriodicSCM = &models.PeriodicSCMTriggerConfig{}
			spec.TriggerConfig.PeriodicSCM.Periodicity = t.Spec.TriggerConfig.PeriodicSCM.Periodicity
			spec.TriggerConfig.PeriodicSCM.CronExpression = t.Spec.TriggerConfig.PeriodicSCM.CronExpression
		default:
			log.Println("invalid triggerType , must one of ", t.Spec.TriggerType, strings.Join(trigger.AllowedTypes, "|"))
		}

		if t.Spec.RepositoryConfig.Git.Revision != "" {
			spec.RepositoryConfig.Git = &models.GitRepositoryConfig{}
			spec.RepositoryConfig.Git.Revision = t.Spec.RepositoryConfig.Git.Revision
			spec.RepositoryConfig.Git.Paths = t.Spec.RepositoryConfig.Git.Paths
		} else if t.Spec.RepositoryConfig.Helm.ChartName != "" {
			spec.RepositoryConfig.Helm = &models.HelmRepositoryConfig{}
			spec.RepositoryConfig.Helm.ChartName = t.Spec.RepositoryConfig.Helm.ChartName
			spec.RepositoryConfig.Helm.Revision = t.Spec.RepositoryConfig.Helm.Revision
		} else {
			log.Println("Failed to create/update Trigger. Required respositoryConfig.git or respositoryConfig.helm")
		}
		for _, v := range t.Spec.Variables {
			spec.Variables = append(spec.Variables, models.InlineVariableSpec{
				Name:  v.Name,
				Type:  models.VariableType(v.Type),
				Value: v.Value,
			})
		}

		err = trigger.UpdateTrigger(t.Metadata.Name, projectId, spec, createIfNotPresent)
		if err != nil {
			log.Println("Failed to update Trigger: \n", t.Metadata.Name)
		} else {
			log.Println("Successfully created/updated Trigger: \n", t.Metadata.Name)
		}
	}
	return diags
}

func resourceTriggerDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("resource agent delete id %s", d.Id())

	//get project id with project name, p.id used to refer to project id -> need p.ID for calling DeleteTrigger
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
	//delete Trigger
	err = trigger.DeleteTrigger(d.Id(), projectId)
	if err != nil {
		log.Println("error deleting trigger")
	} else {
		log.Println("Deleted trigger: ", d.Id())
	}
	return diags
}
