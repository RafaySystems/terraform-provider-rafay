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
	"github.com/RafaySystems/rctl/pkg/pipeline"
	"github.com/RafaySystems/rctl/pkg/project"
	"gopkg.in/yaml.v3"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourcePipeline() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourcePipelineCreate,
		ReadContext:   resourcePipelineRead,
		UpdateContext: resourcePipelineUpdate,
		DeleteContext: resourcePipelineDelete,

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
			"pipeline_filepath": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourcePipelineCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	filePath := d.Get("pipeline_filepath").(string)
	var p commands.PipelineYamlConfig
	//get project id with project name, p.id used to refer to project id -> need p.ID for calling createPipeline
	resp, err := project.GetProjectByName(d.Get("projectname").(string))
	if err != nil {
		log.Printf("project does not exist, error %s", err.Error())
		return diag.FromErr(err)
	}
	proj, err := project.NewProjectFromResponse([]byte(resp))
	if err != nil {
		return diag.FromErr(err)
	} else if proj == nil {
		d.SetId("")
		return diags
	}
	projectId := proj.ID
	//open file path and retirve config spec from yaml file (from run function in commands/create_pipeline.go)
	//read and capture file from file path
	if f, err := os.Open(filePath); err == nil {
		c, err := ioutil.ReadAll(f)
		if err != nil {
			log.Println("error cpaturing file")
		}
		//implement createPipeline from commands/create_Pipeline.go -> then call CreatePipeline
		// unmarshal the data
		pipelineDefinition := c
		err = yaml.Unmarshal(pipelineDefinition, &p)
		if err != nil {
			log.Println("error unmarshaling in create")
		}

		// check if project is provided
		projId := projectId
		if p.Metadata.Project != "" {
			projId, err = config.GetProjectIdByName(p.Metadata.Project)
			if err != nil {
				log.Println("error getting project id by name")
			}
		}
		if p.Spec.Stages == nil || p.Spec.Edges == nil {
			log.Println("Atleast one stage and one edge is required for the pipeline")
		}

		var spec models.PipelineSpec
		for _, v := range p.Spec.Variables {
			spec.Variables = append(spec.Variables, models.InlineVariableSpec{
				Name:  v.Name,
				Type:  models.VariableType(v.Type),
				Value: v.Value,
			})
		}
		for _, s := range p.Spec.Stages {
			stageSpec := models.PipelineStageSpec{}
			stageSpec.Name = s.Name
			switch s.StageType {
			case pipeline.ApprovalStage:
				stageSpec.StageType = pipeline.ApprovalStage
				stageSpec.StageConfig.Approval = &models.ApprovalStageConfig{}
				stageSpec.StageConfig.Approval.ApprovalType = models.ApprovalType(s.StageConfig.Approval.ApprovalType)
				stageSpec.StageConfig.Approval.Emails = s.StageConfig.Approval.Emails
				t, err := time.ParseDuration(s.StageConfig.Approval.Timeout)
				if err != nil {
					log.Println("Invalid value for timeout; Valid time units are \"ns\", \"us\" (or \"µs\"), \"ms\", \"s\", \"m\", \"h\". e.g. 10s, 20m, 1h10m")
				}
				stageSpec.StageConfig.Approval.Timeout = models.Duration{Duration: t}
			case pipeline.DeployWorkloadStage:
				stageSpec.StageType = pipeline.DeployWorkloadStage
				stageSpec.StageConfig.Deployment = &models.DeployWorkloadStageConfig{}
				stageSpec.StageConfig.Deployment.WorkloadRef = s.StageConfig.Deployment.WorkloadRef
			case pipeline.DeployWorkloadTemplateStage:
				stageSpec.StageType = pipeline.DeployWorkloadTemplateStage
				placement := models.PlacementSpec{}
				switch s.StageConfig.WorkloadTemplate.Placement.PlacementType {
				case "ClusterSelector":
					if s.StageConfig.WorkloadTemplate.Placement.ClusterSelector == "" {
						log.Println("cluster selector cannot be empty for ClusterSelector placement type")
					}
					placement.ClusterSelector = s.StageConfig.WorkloadTemplate.Placement.ClusterSelector
				case "ClusterLocations", "ClusterLabels", "ClusterSpecific":
					if len(s.StageConfig.WorkloadTemplate.Placement.ClusterLabels) == 0 {
						log.Println("cluster labels cannot be empty for placement type", s.StageConfig.WorkloadTemplate.Placement.PlacementType)
					}
					for _, label := range s.StageConfig.WorkloadTemplate.Placement.ClusterLabels {
						placement.ClusterLabels = append(placement.ClusterLabels, &models.PlacementLabel{
							Key:   label.Key,
							Value: label.Value,
						})
					}
				default:
					log.Println("invalid placement type ", s.StageConfig.WorkloadTemplate.Placement.PlacementType)
				}
				placement.PlacementType = models.PlacementType(s.StageConfig.WorkloadTemplate.Placement.PlacementType)

				switch s.StageConfig.WorkloadTemplate.Placement.DriftAction {
				case "DriftReconcillationActionNotSet", "DriftReconcillationActionNotify", "DriftReconcillationActionDeny":
				default:
					log.Println("Invalid driftAction: Allowed [DriftReconcillationActionNotSet, DriftReconcillationActionNotify, DriftReconcillationActionDeny]")
				}
				placement.DriftAction = models.DriftReconcillationAction(s.StageConfig.WorkloadTemplate.Placement.DriftAction)

				stageSpec.StageConfig.WorkloadTemplate = &models.DeployWorkloadTemplateStageConfig{
					WorkloadTemplateRef: s.StageConfig.WorkloadTemplate.WorkloadTemplateRef,
					Namespace:           s.StageConfig.WorkloadTemplate.Namespace,
					Placement:           placement,
				}
				for _, o := range s.StageConfig.WorkloadTemplate.Overrides {
					var mo models.OverrideTemplate
					mo.RepositoryRef = o.RepositoryRef
					mo.InlineTemplate = o.InlineTemplate
					mo.OverrideType = models.OverrideTemplateType(o.OverrideType)
					mo.OverrideWeight = o.OverrideWeight
					mo.Revision = o.Revision
					for _, rp := range o.RepoPaths {
						mo.RepoPaths = append(mo.RepoPaths, &models.RepositoryPath{
							FolderPath: rp.FolderPath,
							FileFilter: rp.FileFilter,
						})
					}
					mo.TemplateSource = models.OverrideTemplateSource(o.TemplateSoure)

					stageSpec.StageConfig.WorkloadTemplate.Overrides = append(stageSpec.StageConfig.WorkloadTemplate.Overrides, &mo)
				}
			case pipeline.InfraProvisionerStage:
				stageSpec.StageType = pipeline.InfraProvisionerStage
				stageSpec.StageConfig.InfraProvisioner = &models.InfraProvisionerStageConfig{
					InfraProvisionerName: s.StageConfig.InfraProvisioner.InfraProvisionerName,
					GitRevision:          s.StageConfig.InfraProvisioner.GitRevision,
					UseWorkingDirFrom:    s.StageConfig.InfraProvisioner.UseWorkingDirFrom,
					PersistWorkingDir:    s.StageConfig.InfraProvisioner.PersistWorkingDir,
					AgentNames:           s.StageConfig.InfraProvisioner.AgentNames,
				}
				if s.StageConfig.InfraProvisioner.ActionConfig.Terraform != nil {
					terraform := s.StageConfig.InfraProvisioner.ActionConfig.Terraform
					stageSpec.StageConfig.InfraProvisioner.ActionConfig.Terraform = &models.TerraformAction{
						Type:      models.TerraformActionType(terraform.Type),
						NoRefresh: terraform.NoRefresh,
						Targets:   terraform.Targets,
						Destroy:   terraform.Destroy,
					}
				}
				if s.StageConfig.InfraProvisioner.Config.Terraform != nil {
					terraform := s.StageConfig.InfraProvisioner.Config.Terraform
					stageSpec.StageConfig.InfraProvisioner.Config.Terraform = &models.TerraformConfig{
						Version:    terraform.Version,
						TfvarsFile: terraform.TfvarsFile,
					}
					stageSpec.StageConfig.InfraProvisioner.Config.Terraform.InputVars = []models.KeyValue{}
					for _, kv := range terraform.InputVars {
						stageSpec.StageConfig.InfraProvisioner.Config.Terraform.InputVars = append(stageSpec.StageConfig.InfraProvisioner.Config.Terraform.InputVars, models.KeyValue{
							Key:   kv.Key,
							Value: kv.Value,
							Type:  models.ValueType(kv.Type),
						})
					}
					stageSpec.StageConfig.InfraProvisioner.Config.Terraform.EnvVars = []models.KeyValue{}
					for _, kv := range terraform.EnvVars {
						stageSpec.StageConfig.InfraProvisioner.Config.Terraform.EnvVars = append(stageSpec.StageConfig.InfraProvisioner.Config.Terraform.EnvVars, models.KeyValue{
							Key:   kv.Key,
							Value: kv.Value,
							Type:  models.ValueType(kv.Type),
						})
					}
					if terraform.BackendConfig != nil {
						stageSpec.StageConfig.InfraProvisioner.Config.Terraform.BackendConfig = &models.TerraformBackendConfig{
							File: terraform.BackendConfig.File,
						}
						stageSpec.StageConfig.InfraProvisioner.Config.Terraform.BackendConfig.KeyValues = []models.KeyValue{}
						for _, kv := range terraform.BackendConfig.KeyValues {
							stageSpec.StageConfig.InfraProvisioner.Config.Terraform.BackendConfig.KeyValues = append(stageSpec.StageConfig.InfraProvisioner.Config.Terraform.BackendConfig.KeyValues, models.KeyValue{
								Key:   kv.Key,
								Value: kv.Value,
								Type:  models.ValueType(kv.Type),
							})
						}
					}
				}
			default:
				log.Println("invalid stageType, must one of ", s.StageType, strings.Join(pipeline.AllowedStageTypes, "|"))
			}
			stageSpec.StageConfig.OnFailure = models.StageOnFailure(s.StageConfig.OnFailure)
			for _, v := range s.Variables {
				stageSpec.Variables = append(stageSpec.Variables, models.InlineVariableSpec{
					Name:  v.Name,
					Type:  models.VariableType(v.Type),
					Value: v.Value,
				})
			}
			for _, pc := range s.PreConditions {
				stageSpec.PreConditions = append(stageSpec.PreConditions, models.StagePreCondition{
					ConditionType: models.StagePreConditionType(pc.ConditionType),
					Config: models.StagePreConditionConfig{
						Expression: pc.Config.Expression,
					},
				})
			}
			spec.Stages = append(spec.Stages, stageSpec)
		}
		for _, e := range p.Spec.Edges {
			edge := models.PipelineStageEdge{}
			edge.Source = e.Source
			edge.Target = e.Target
			edge.Weight = e.Weight
			spec.Edges = append(spec.Edges, edge)
		}
		err = pipeline.CreatePipeline(p.Metadata.Name, projId, spec)
		if err != nil {
			log.Println("Failed to create Pipeline: ", p.Metadata.Name, "err: ", err)
			return diags
		} else {
			log.Println("Successfully created pipeline: ", p.Metadata.Name)
		}
	}
	//set id to metadata.Name
	d.SetId(p.Metadata.Name)
	return diags
}

func resourcePipelineRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	return diags
}

func resourcePipelineUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	filePath := d.Get("pipeline_filepath").(string)
	var p commands.PipelineYamlConfig
	createIfNotPresent := false
	//get project id with project name, p.id used to refer to project id -> need p.ID for calling UpdatePipeline
	resp, err := project.GetProjectByName(d.Get("projectname").(string))
	if err != nil {
		log.Printf("project does not exist, error %s", err.Error())
		return diag.FromErr(err)
	}
	proj, err := project.NewProjectFromResponse([]byte(resp))
	if err != nil {
		return diag.FromErr(err)
	} else if proj == nil {
		d.SetId("")
		return diags
	}
	projectId := proj.ID
	// read the file
	if f, err := os.Open(filePath); err == nil {
		// capture the entire file
		c, err := ioutil.ReadAll(f)
		if err != nil {
			log.Println("error reading file ")
		}
		pipelineDefinition := c
		//implement updatePipeline to call UpdatePipeline
		err = yaml.Unmarshal(pipelineDefinition, &p)
		if err != nil {
			log.Println("error unmarshalling")
		}

		// check if project is provided
		projId := projectId
		if p.Metadata.Project != "" {
			projId, err = config.GetProjectIdByName(p.Metadata.Project)
			if err != nil {
				log.Println("error getting project id by name")
			}
		}

		if p.Spec.Stages == nil || p.Spec.Edges == nil {
			log.Println("Atleast one stage and one edge is required for the pipeline")
		}

		var spec models.PipelineSpec
		for _, v := range p.Spec.Variables {
			spec.Variables = append(spec.Variables, models.InlineVariableSpec{
				Name:  v.Name,
				Type:  models.VariableType(v.Type),
				Value: v.Value,
			})
		}
		for _, s := range p.Spec.Stages {
			stageSpec := models.PipelineStageSpec{}
			stageSpec.Name = s.Name
			switch s.StageType {
			case pipeline.ApprovalStage:
				stageSpec.StageType = pipeline.ApprovalStage
				stageSpec.StageConfig.Approval = &models.ApprovalStageConfig{}
				stageSpec.StageConfig.Approval.ApprovalType = models.ApprovalType(s.StageConfig.Approval.ApprovalType)
				stageSpec.StageConfig.Approval.Emails = s.StageConfig.Approval.Emails
				t, err := time.ParseDuration(s.StageConfig.Approval.Timeout)
				if err != nil {
					log.Println("Invalid value for timeout; Valid time units are \"ns\", \"us\" (or \"µs\"), \"ms\", \"s\", \"m\", \"h\". e.g. 10s, 20m, 1h10m")
				}
				stageSpec.StageConfig.Approval.Timeout = models.Duration{Duration: t}
			case pipeline.DeployWorkloadStage:
				stageSpec.StageType = pipeline.DeployWorkloadStage
				stageSpec.StageConfig.Deployment = &models.DeployWorkloadStageConfig{}
				stageSpec.StageConfig.Deployment.WorkloadRef = s.StageConfig.Deployment.WorkloadRef
			case pipeline.DeployWorkloadTemplateStage:
				stageSpec.StageType = pipeline.DeployWorkloadTemplateStage

				placement := models.PlacementSpec{}
				switch s.StageConfig.WorkloadTemplate.Placement.PlacementType {
				case "ClusterSelector":
					if s.StageConfig.WorkloadTemplate.Placement.ClusterSelector == "" {
						log.Println("cluster selector cannot be empty for ClusterSelector placement type")
					}
					placement.ClusterSelector = s.StageConfig.WorkloadTemplate.Placement.ClusterSelector
				case "ClusterLocations", "ClusterLabels", "ClusterSpecific":
					if len(s.StageConfig.WorkloadTemplate.Placement.ClusterLabels) == 0 {
						log.Println("cluster labels cannot be empty for placement type", s.StageConfig.WorkloadTemplate.Placement.PlacementType)
					}
					for _, label := range s.StageConfig.WorkloadTemplate.Placement.ClusterLabels {
						placement.ClusterLabels = append(placement.ClusterLabels, &models.PlacementLabel{
							Key:   label.Key,
							Value: label.Value,
						})
					}
				default:
					log.Println("invalid placement type ", s.StageConfig.WorkloadTemplate.Placement.PlacementType)
				}
				placement.PlacementType = models.PlacementType(s.StageConfig.WorkloadTemplate.Placement.PlacementType)

				switch s.StageConfig.WorkloadTemplate.Placement.DriftAction {
				case "DriftReconcillationActionNotSet", "DriftReconcillationActionNotify", "DriftReconcillationActionDeny":
				default:
					log.Println("Invalid driftAction: Allowed [DriftReconcillationActionNotSet, DriftReconcillationActionNotify, DriftReconcillationActionDeny]")
				}
				placement.DriftAction = models.DriftReconcillationAction(s.StageConfig.WorkloadTemplate.Placement.DriftAction)
				stageSpec.StageConfig.WorkloadTemplate = &models.DeployWorkloadTemplateStageConfig{
					WorkloadTemplateRef: s.StageConfig.WorkloadTemplate.WorkloadTemplateRef,
					Namespace:           s.StageConfig.WorkloadTemplate.Namespace,
					Placement:           placement,
				}
				for _, o := range s.StageConfig.WorkloadTemplate.Overrides {
					var mo models.OverrideTemplate
					mo.RepositoryRef = o.RepositoryRef
					mo.InlineTemplate = o.InlineTemplate
					mo.OverrideType = models.OverrideTemplateType(o.OverrideType)
					mo.OverrideWeight = o.OverrideWeight
					mo.Revision = o.Revision
					for _, rp := range o.RepoPaths {
						mo.RepoPaths = append(mo.RepoPaths, &models.RepositoryPath{
							FolderPath: rp.FolderPath,
							FileFilter: rp.FileFilter,
						})
					}
					mo.TemplateSource = models.OverrideTemplateSource(o.TemplateSoure)

					stageSpec.StageConfig.WorkloadTemplate.Overrides = append(stageSpec.StageConfig.WorkloadTemplate.Overrides, &mo)
				}
			case pipeline.InfraProvisionerStage:
				stageSpec.StageType = pipeline.InfraProvisionerStage
				stageSpec.StageConfig.InfraProvisioner = &models.InfraProvisionerStageConfig{
					InfraProvisionerName: s.StageConfig.InfraProvisioner.InfraProvisionerName,
					GitRevision:          s.StageConfig.InfraProvisioner.GitRevision,
					UseWorkingDirFrom:    s.StageConfig.InfraProvisioner.UseWorkingDirFrom,
					PersistWorkingDir:    s.StageConfig.InfraProvisioner.PersistWorkingDir,
					AgentNames:           s.StageConfig.InfraProvisioner.AgentNames,
				}
				if s.StageConfig.InfraProvisioner.ActionConfig.Terraform != nil {
					terraform := s.StageConfig.InfraProvisioner.ActionConfig.Terraform
					stageSpec.StageConfig.InfraProvisioner.ActionConfig.Terraform = &models.TerraformAction{
						Type:      models.TerraformActionType(terraform.Type),
						NoRefresh: terraform.NoRefresh,
						Targets:   terraform.Targets,
						Destroy:   terraform.Destroy,
					}
				}
				if s.StageConfig.InfraProvisioner.Config.Terraform != nil {
					terraform := s.StageConfig.InfraProvisioner.Config.Terraform
					stageSpec.StageConfig.InfraProvisioner.Config.Terraform = &models.TerraformConfig{
						Version:    terraform.Version,
						TfvarsFile: terraform.TfvarsFile,
					}
					stageSpec.StageConfig.InfraProvisioner.Config.Terraform.InputVars = []models.KeyValue{}
					for _, kv := range terraform.InputVars {
						stageSpec.StageConfig.InfraProvisioner.Config.Terraform.InputVars = append(stageSpec.StageConfig.InfraProvisioner.Config.Terraform.InputVars, models.KeyValue{
							Key:   kv.Key,
							Value: kv.Value,
							Type:  models.ValueType(kv.Type),
						})
					}
					stageSpec.StageConfig.InfraProvisioner.Config.Terraform.EnvVars = []models.KeyValue{}
					for _, kv := range terraform.EnvVars {
						stageSpec.StageConfig.InfraProvisioner.Config.Terraform.EnvVars = append(stageSpec.StageConfig.InfraProvisioner.Config.Terraform.EnvVars, models.KeyValue{
							Key:   kv.Key,
							Value: kv.Value,
							Type:  models.ValueType(kv.Type),
						})
					}
					if terraform.BackendConfig != nil {
						stageSpec.StageConfig.InfraProvisioner.Config.Terraform.BackendConfig = &models.TerraformBackendConfig{
							File: terraform.BackendConfig.File,
						}
						stageSpec.StageConfig.InfraProvisioner.Config.Terraform.BackendConfig.KeyValues = []models.KeyValue{}
						for _, kv := range terraform.BackendConfig.KeyValues {
							stageSpec.StageConfig.InfraProvisioner.Config.Terraform.BackendConfig.KeyValues = append(stageSpec.StageConfig.InfraProvisioner.Config.Terraform.BackendConfig.KeyValues, models.KeyValue{
								Key:   kv.Key,
								Value: kv.Value,
								Type:  models.ValueType(kv.Type),
							})
						}
					}
				}
			default:
				log.Println("invalid stageType, must one of", s.StageType, strings.Join(pipeline.AllowedStageTypes, "|"))
			}
			stageSpec.StageConfig.OnFailure = models.StageOnFailure(s.StageConfig.OnFailure)
			for _, v := range s.Variables {
				stageSpec.Variables = append(stageSpec.Variables, models.InlineVariableSpec{
					Name:  v.Name,
					Type:  models.VariableType(v.Type),
					Value: v.Value,
				})
			}
			for _, pc := range s.PreConditions {
				stageSpec.PreConditions = append(stageSpec.PreConditions, models.StagePreCondition{
					ConditionType: models.StagePreConditionType(pc.ConditionType),
					Config: models.StagePreConditionConfig{
						Expression: pc.Config.Expression,
					},
				})
			}

			spec.Stages = append(spec.Stages, stageSpec)
		}
		for _, e := range p.Spec.Edges {
			edge := models.PipelineStageEdge{}
			edge.Source = e.Source
			edge.Target = e.Target
			edge.Weight = e.Weight
			spec.Edges = append(spec.Edges, edge)
		}
		//call update pipeline
		err = pipeline.UpdatePipeline(p.Metadata.Name, projId, spec, createIfNotPresent)
		if err != nil {
			log.Println("Failed to update Pipeline: ", p.Metadata.Name)
		} else {
			log.Println("Successfully created/updated Pipeline: ", p.Metadata.Name)
		}
	} else {
		log.Println("error opening the file")
	}
	return diags
}

func resourcePipelineDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("resource pipeline id %s", d.Id())
	//get project id with project name, p.id used to refer to project id -> need p.ID for calling Deletepipeline
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
	//delete pipeline
	err = pipeline.DeletePipeline(d.Id(), projectId)
	if err != nil {
		log.Println("error deleting pipeline")
	} else {
		log.Println("Deleted pipeline: ", d.Id())
	}
	return diags
}
