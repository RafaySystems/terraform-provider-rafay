package rafay

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/RafaySystems/rafay-common/pkg/hub/client/options"
	typed "github.com/RafaySystems/rafay-common/pkg/hub/client/typed"
	"github.com/RafaySystems/rafay-common/pkg/hub/terraform/resource"
	"github.com/RafaySystems/rafay-common/proto/types/hub/commonpb"
	"github.com/RafaySystems/rafay-common/proto/types/hub/gitopspb"
	"github.com/RafaySystems/rctl/pkg/config"
	"github.com/RafaySystems/rctl/pkg/pipeline"
	"github.com/RafaySystems/rctl/pkg/user"
	"github.com/RafaySystems/rctl/pkg/versioninfo"
	"github.com/davecgh/go-spew/spew"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourcePipeline() *schema.Resource {
	modSchema := resource.PipelineSchema.Schema
	modSchema["impersonate"] = &schema.Schema{
		Description: "impersonate user",
		Optional:    true,
		Type:        schema.TypeString,
	}
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
		Schema:        modSchema,
	}
}

//Stage Spec

type stageSpec struct {
	Name          string                      `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Type          string                      `protobuf:"bytes,2,opt,name=type,proto3" json:"type,omitempty"`
	PreConditions []stageSpecPreConditionSpec `protobuf:"bytes,3,opt,name=type,proto3" json:"preConditions,omitempty"`
	Variables     []*gitopspb.VariableSpec    `protobuf:"bytes,4,rep,name=variables,proto3" json:"variables,omitempty"`
	Next          []*gitopspb.NextStage       `protobuf:"bytes,5,rep,name=next,proto3" json:"next,omitempty"`
	// Types that are assignable to Config:
	//    *StageSpec_Approval
	//    *StageSpec_Workload
	//    *StageSpec_WorkloadTemplate
	//    *StageSpec_InfraProvisioner
	//    *StageSpec_SystemSync
	Config stageSpecConfig `json:"config,omitempty"`
}

type stageSpecPreConditionSpec struct {
	Type   string `protobuf:"bytes,1,opt,name=type,proto3" json:"type,omitempty"`
	Config struct {
		Expression string `protobuf:"bytes,1,opt,name=expression,proto3" json:"expression,omitempty"`
	} `json:"config,omitempty"`
}

type stageSpecConfig struct {
	Type                               string                                     `protobuf:"bytes,1,opt,name=type,proto3" json:"type,omitempty"`
	Approvers                          []*gitopspb.Approver                       `protobuf:"bytes,2,rep,name=approvers,proto3" json:"approvers,omitempty"`
	Timeout                            string                                     `protobuf:"bytes,3,opt,name=timeout,proto3" json:"timeout,omitempty"`
	Workload                           string                                     `protobuf:"bytes,1,opt,name=workload,proto3" json:"workload,omitempty"`
	WorkloadTemplate                   string                                     `protobuf:"bytes,1,opt,name=workloadTemplate,proto3" json:"workloadTemplate,omitempty"`
	Namespace                          string                                     `protobuf:"bytes,2,opt,name=namespace,proto3" json:"namespace,omitempty"`
	Placement                          *commonpb.PlacementSpec                    `protobuf:"bytes,3,opt,name=placement,proto3" json:"placement,omitempty"`
	Overrides                          []stageSpecConfigWorkloadTemplateOverrides `protobuf:"bytes,4,rep,name=overrides,proto3" json:"overrides,omitempty"`
	UseRevisionFromWebhookTriggerEvent bool                                       `protobuf:"varint,5,opt,name=useRevisionFromWebhookTriggerEvent,proto3" json:"useRevisionFromWebhookTriggerEvent,omitempty"`
	Provisioner                        string                                     `protobuf:"bytes,2,opt,name=provisioner,proto3" json:"provisioner,omitempty"`
	Revision                           string                                     `protobuf:"bytes,3,opt,name=revision,proto3" json:"revision,omitempty"`
	WorkingDirectory                   string                                     `protobuf:"bytes,4,opt,name=workingDirectory,proto3" json:"workingDirectory,omitempty"`
	PersistWorkingDirectory            bool                                       `protobuf:"varint,5,opt,name=persistWorkingDirectory,proto3" json:"persistWorkingDirectory,omitempty"`
	Agents                             []*gitopspb.AgentMeta                      `protobuf:"bytes,6,rep,name=agents,proto3" json:"agents,omitempty"`
	Action                             struct {
		Action          string                      `protobuf:"bytes,1,opt,name=action,proto3" json:"action,omitempty"`
		Version         string                      `protobuf:"bytes,2,opt,name=version,proto3" json:"version,omitempty"`
		InputVars       []*gitopspb.KeyValue        `protobuf:"bytes,3,rep,name=inputVars,proto3" json:"inputVars,omitempty"`
		TfVarsFilePath  *commonpb.File              `protobuf:"bytes,4,opt,name=tfVarsFilePath,proto3" json:"tfVarsFilePath,omitempty"`
		EnvVars         []*gitopspb.KeyValue        `protobuf:"bytes,5,rep,name=envVars,proto3" json:"envVars,omitempty"`
		BackendVars     []*gitopspb.KeyValue        `protobuf:"bytes,6,rep,name=backendVars,proto3" json:"backendVars,omitempty"`
		BackendFilePath *commonpb.File              `protobuf:"bytes,7,opt,name=backendFilePath,proto3" json:"backendFilePath,omitempty"`
		Refresh         bool                        `protobuf:"varint,8,opt,name=refresh,proto3" json:"refresh,omitempty"`
		Targets         []*gitopspb.TerraformTarget `protobuf:"bytes,9,rep,name=targets,proto3" json:"targets,omitempty"`
		Destroy         bool                        `protobuf:"varint,10,opt,name=destroy,proto3" json:"destroy,omitempty"`
	} `json:"action,omitempty"`
	GitToSystemSync     bool                           `protobuf:"varint,1,opt,name=gitToSystemSync,proto3" json:"gitToSystemSync,omitempty"`
	SystemToGitSync     bool                           `protobuf:"varint,2,opt,name=systemToGitSync,proto3" json:"systemToGitSync,omitempty"`
	IncludedResources   []*gitopspb.SystemSyncResource `protobuf:"bytes,3,rep,name=includedResources,proto3" json:"includedResources,omitempty"`
	ExcludedResources   []*gitopspb.SystemSyncResource `protobuf:"bytes,4,rep,name=excludedResources,proto3" json:"excludedResources,omitempty"`
	SourceRepo          *gitopspb.SystemSyncRepo       `protobuf:"bytes,5,opt,name=sourceRepo,proto3" json:"sourceRepo,omitempty"`
	DestinationRepo     *gitopspb.SystemSyncRepo       `protobuf:"bytes,6,opt,name=destinationRepo,proto3" json:"destinationRepo,omitempty"`
	SourceAsDestination bool                           `protobuf:"varint,7,opt,name=sourceAsDestination,proto3" json:"sourceAsDestination,omitempty"`
}

type stageSpecConfigWorkloadTemplateOverrides struct {
	Type     string `protobuf:"bytes,1,opt,name=type,proto3" json:"type,omitempty"`
	Weight   int32  `protobuf:"zigzag32,2,opt,name=weight,proto3" json:"weight,omitempty"`
	Template struct {
		Inline     string           `protobuf:"bytes,1,opt,name=inline,proto3" json:"inline,omitempty"`
		Repository string           `protobuf:"bytes,1,opt,name=repository,proto3" json:"repository,omitempty"`
		Revision   string           `protobuf:"bytes,2,opt,name=revision,proto3" json:"revision,omitempty"`
		Paths      []*commonpb.File `protobuf:"bytes,3,rep,name=paths,proto3" json:"paths,omitempty"`
	} `json:"template,omitempty"`
}

// TriggerSpec

type triggerSpec struct {
	Type      string                   `protobuf:"bytes,1,opt,name=type,proto3" json:"type,omitempty"`
	Name      string                   `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	Variables []*gitopspb.VariableSpec `protobuf:"bytes,3,rep,name=variables,proto3" json:"variables,omitempty"`
	Config    struct {
		Repo struct {
			Provider     string           `protobuf:"bytes,1,opt,name=provider,proto3" json:"provider,omitempty"`
			Repository   string           `protobuf:"bytes,2,opt,name=repository,proto3" json:"repository,omitempty"`
			Revision     string           `protobuf:"bytes,3,opt,name=revision,proto3" json:"revision,omitempty"`
			Paths        []*commonpb.File `protobuf:"bytes,4,rep,name=paths,proto3" json:"paths,omitempty"`
			ChartName    string           `protobuf:"bytes,2,opt,name=chartName,proto3" json:"chartName,omitempty"`
			ChartVersion string           `protobuf:"bytes,3,opt,name=chartVersion,proto3" json:"chartVersion,omitempty"`
		} `json:"repo,omitempty"`
		CronExpression string `protobuf:"bytes,1,opt,name=cronExpression,proto3" json:"cronExpression,omitempty"`
	}
}

func resourcePipelineCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("pipeline create starts")
	diags := resourcePipelineUpsert(ctx, d, m)
	if diags.HasError() {
		tflog := os.Getenv("TF_LOG")
		if tflog == "TRACE" || tflog == "DEBUG" {
			ctx = context.WithValue(ctx, "debug", "true")
		}
		log.Printf("pipeline create got error, perform cleanup")
		ag, err := expandPipeline(d)
		if err != nil {
			log.Printf("pipeline expandPipeline error")
			return diags
		}

		if v, ok := d.Get("impersonate").(string); ok && len(v) > 0 {
			defer ResetImpersonateUser()
			asUser := d.Get("impersonate").(string)
			// check user role : impersonation not allowed for a user
			// with ORG Admin role
			isOrgAdmin, err := user.IsOrgAdmin(asUser)
			if err != nil {
				return diag.FromErr(err)
			}
			if isOrgAdmin {
				return diag.FromErr(fmt.Errorf("%s", "--as-user cannot have ORGADMIN role"))
			}
			config.ApiKey, config.ApiSecret, err = user.GetUserAPIKey(asUser)
			if err != nil {
				return diag.FromErr(err)
			}
		}

		auth := config.GetConfig().GetAppAuthProfile()
		client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
		if err != nil {
			return diags
		}

		err = client.GitopsV3().Pipeline().Delete(ctx, options.DeleteOptions{
			Name:    ag.Metadata.Name,
			Project: ag.Metadata.Project,
		})
		if err != nil {
			return diags
		}
	}
	return diags
}

func resourcePipelineUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("pipeline update starts")
	return resourcePipelineUpsert(ctx, d, m)
}

func resourcePipelineUpsert(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("pipeline upsert starts")
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	pipeline, err := expandPipeline(d)
	if err != nil {
		log.Printf("pipeline expandPipeline error")
		return diag.FromErr(err)
	}

	if v, ok := d.Get("impersonate").(string); ok && len(v) > 0 {
		defer ResetImpersonateUser()
		asUser := d.Get("impersonate").(string)
		// check user role : impersonation not allowed for a user
		// with ORG Admin role
		isOrgAdmin, err := user.IsOrgAdmin(asUser)
		if err != nil {
			return diag.FromErr(err)
		}
		if isOrgAdmin {
			return diag.FromErr(fmt.Errorf("%s", "--as-user cannot have ORGADMIN role"))
		}
		config.ApiKey, config.ApiSecret, err = user.GetUserAPIKey(asUser)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}

	pipe1 := spew.Sprintf("%+v", pipeline)
	log.Println("pipeline apply pipeline:", pipe1)

	err = client.GitopsV3().Pipeline().Apply(ctx, pipeline, options.ApplyOptions{})
	if err != nil {
		// XXX Debug
		n1 := spew.Sprintf("%+v", pipeline)
		log.Println("pipeline apply pipeline:", n1)
		log.Printf("pipeline apply error")
		return diag.FromErr(err)
	}

	d.SetId(pipeline.Metadata.Name)
	return diags

}

func resourcePipelineRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	log.Println("resourcePipelineRead ")
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}
	tfPipelineState, err := expandPipeline(d)
	if err != nil {
		return diag.FromErr(err)
	}

	if v, ok := d.Get("impersonate").(string); ok && len(v) > 0 {
		defer ResetImpersonateUser()
		asUser := d.Get("impersonate").(string)
		// check user role : impersonation not allowed for a user
		// with ORG Admin role
		isOrgAdmin, err := user.IsOrgAdmin(asUser)
		if err != nil {
			return diag.FromErr(err)
		}
		if isOrgAdmin {
			return diag.FromErr(fmt.Errorf("%s", "--as-user cannot have ORGADMIN role"))
		}
		config.ApiKey, config.ApiSecret, err = user.GetUserAPIKey(asUser)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}

	ag, err := client.GitopsV3().Pipeline().Get(ctx, options.GetOptions{
		Name:    tfPipelineState.Metadata.Name,
		Project: tfPipelineState.Metadata.Project,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	err = flattenPipeline(d, ag)
	if err != nil {
		return diag.FromErr(err)
	}
	return diags

}

func resourcePipelineDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	ag, err := expandPipeline(d)
	if err != nil {
		return diag.FromErr(err)
	}

	if v, ok := d.Get("impersonate").(string); ok && len(v) > 0 {
		defer ResetImpersonateUser()
		asUser := d.Get("impersonate").(string)
		// check user role : impersonation not allowed for a user
		// with ORG Admin role
		isOrgAdmin, err := user.IsOrgAdmin(asUser)
		if err != nil {
			return diag.FromErr(err)
		}
		if isOrgAdmin {
			return diag.FromErr(fmt.Errorf("%s", "--as-user cannot have ORGADMIN role"))
		}
		config.ApiKey, config.ApiSecret, err = user.GetUserAPIKey(asUser)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.GitopsV3().Pipeline().Delete(ctx, options.DeleteOptions{
		Name:    ag.Metadata.Name,
		Project: ag.Metadata.Project,
	})

	if err != nil {
		//v3 spec gave error try v2
		return resourcePipelineV2Delete(ctx, ag)
	}

	return diags
}

func resourcePipelineV2Delete(ctx context.Context, pl *gitopspb.Pipeline) diag.Diagnostics {
	var diags diag.Diagnostics

	projectId, err := config.GetProjectIdByName(pl.Metadata.Project)
	if err != nil {
		return diag.FromErr(err)
	}
	//delete pipeline
	err = pipeline.DeletePipeline(pl.Metadata.Name, projectId)
	if err != nil {
		log.Println("error deleting pipeline")
	} else {
		log.Println("Deleted pipeline: ", pl.Metadata.Name)
	}
	return diags
}

func expandPipeline(in *schema.ResourceData) (*gitopspb.Pipeline, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "expand pipeline empty input")
	}
	obj := &gitopspb.Pipeline{}

	if v, ok := in.Get("metadata").([]interface{}); ok && len(v) > 0 {
		obj.Metadata = expandMetaData(v)
	}

	if v, ok := in.Get("spec").([]interface{}); ok && len(v) > 0 {
		objSpec, err := expandPipelineSpec(v)
		if err != nil {
			return nil, err
		}
		log.Println("expandPipelineSpec got spec")
		obj.Spec = objSpec
	}

	obj.ApiVersion = "gitops.k8smgmt.io/v3"
	obj.Kind = "Pipeline"
	return obj, nil
}

func expandPipelineSpec(p []interface{}) (*gitopspb.PipelineSpec, error) {
	var err error
	obj := &gitopspb.PipelineSpec{}
	if len(p) == 0 || p[0] == nil {
		return obj, fmt.Errorf("%s", "expandPipelineSpec empty input")
	}

	in := p[0].(map[string]interface{})

	// Stages    []*StageSpec          `protobuf:"bytes,1,rep,name=stages,proto3" json:"stages,omitempty"`
	if v, ok := in["stages"].([]interface{}); ok && len(v) > 0 {
		obj.Stages, err = expandStageSpec(v)
		if err != nil {
			return obj, err
		}
	}

	// Variables []*VariableSpec       `protobuf:"bytes,2,rep,name=variables,proto3" json:"variables,omitempty"`
	if v, ok := in["variables"].([]interface{}); ok && len(v) > 0 {
		obj.Variables = expandVariableSpec(v)
	}
	// Triggers  []*TriggerSpec        `protobuf:"bytes,3,rep,name=triggers,proto3" json:"triggers,omitempty"`
	if v, ok := in["triggers"].([]interface{}); ok && len(v) > 0 {
		obj.Triggers = expandTriggerSpec(v)
	}

	if v, ok := in["sharing"].([]interface{}); ok && len(v) > 0 {
		obj.Sharing = expandSharingSpec(v)
	}

	if v, ok := in["active"].(bool); ok {
		obj.Active = v
	}

	if v, ok := in["secret"].([]interface{}); ok {
		obj.Secret = expandCommonpbFile(v)
	}

	return obj, nil
}

func expandApprovers(p []interface{}) []*gitopspb.Approver {
	if len(p) == 0 || p[0] == nil {
		return []*gitopspb.Approver{}
	}

	out := make([]*gitopspb.Approver, len(p))

	for i := range p {
		obj := gitopspb.Approver{}
		in := p[i].(map[string]interface{})

		if v, ok := in["userName"].(string); ok && len(v) > 0 {
			obj.UserName = v
		}

		if v, ok := in["enabled"].(bool); ok {
			obj.SsoUser = v
		}
		out[i] = &obj
	}
	return out
}

func expandStageSpecConfigWorkload(p []interface{}) (*gitopspb.StageSpec_Workload, error) {
	obj := gitopspb.StageSpec_Workload{}
	obj.Workload = &gitopspb.DeployWorkloadConfig{}
	if len(p) == 0 || p[0] == nil {
		log.Println("expandStageSpecConfigWorkload empty")
		return &obj, fmt.Errorf("%s", "expandStageSpecConfigWorkload empty input")
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["workload"].(string); ok && len(v) > 0 {
		log.Println("expandStageSpecConfigWorkload workload")
		obj.Workload.Workload = v
	}

	if v, ok := in["useRevisionFromWebhookTriggerEvent"].(bool); ok {
		log.Println("expandStageSpecConfigWorkload useRevisionFromWebhookTriggerEvent")
		obj.Workload.UseRevisionFromWebhookTriggerEvent = v
	}

	log.Println("expandStageSpecConfigWorkload obj:", obj)

	return &obj, nil
}

func expandIncludedResourceType(p []interface{}) ([]*gitopspb.SystemSyncResource, error) {
	if len(p) == 0 || p[0] == nil {
		return []*gitopspb.SystemSyncResource{}, fmt.Errorf("%s", "expandIncludedResourceType empty input")
	}

	out := make([]*gitopspb.SystemSyncResource, len(p))

	for i := range p {
		obj := gitopspb.SystemSyncResource{}
		in := p[i].(map[string]interface{})

		if v, ok := in["name"].(string); ok && len(v) > 0 {
			obj.Name = v
		}
		out[i] = &obj
	}

	return out, nil
}

func expandExcludedResourceType(p []interface{}) ([]*gitopspb.SystemSyncResource, error) {
	if len(p) == 0 || p[0] == nil {
		return []*gitopspb.SystemSyncResource{}, fmt.Errorf("%s", "expandExcludedResourceType empty input")
	}

	out := make([]*gitopspb.SystemSyncResource, len(p))

	for i := range p {
		obj := gitopspb.SystemSyncResource{}
		in := p[i].(map[string]interface{})

		if v, ok := in["name"].(string); ok && len(v) > 0 {
			obj.Name = v
		}
		out[i] = &obj
	}

	return out, nil
}

// type SystemSyncRepo struct {
// 	state         protoimpl.MessageState
// 	sizeCache     protoimpl.SizeCache
// 	unknownFields protoimpl.UnknownFields

// 	Repository string         `protobuf:"bytes,1,opt,name=repository,proto3" json:"repository,omitempty"`
// 	Revision   string         `protobuf:"bytes,2,opt,name=revision,proto3" json:"revision,omitempty"`
// 	Path       *commonpb.File `protobuf:"bytes,3,opt,name=path,proto3" json:"path,omitempty"`
// }

func expandRepo(p []interface{}) (*gitopspb.SystemSyncRepo, error) {
	obj := &gitopspb.SystemSyncRepo{}
	if len(p) == 0 || p[0] == nil {
		return obj, fmt.Errorf("%s", "expandRepo empty input")
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["repository"].(string); ok && len(v) > 0 {
		obj.Repository = v
	}

	if v, ok := in["revision"].(string); ok && len(v) > 0 {
		obj.Revision = v
	}

	if v, ok := in["path"].([]interface{}); ok {
		obj.Path = expandCommonpbFile(v)
	}

	return obj, nil
}

func expandStageSpecConfigSystemSync(p []interface{}) (*gitopspb.StageSpec_SystemSync, error) {
	var err error

	obj := gitopspb.StageSpec_SystemSync{}
	obj.SystemSync = &gitopspb.SystemSyncConfig{}

	if len(p) == 0 || p[0] == nil {
		return &obj, fmt.Errorf("%s", "expandStageSpecConfigSystemSync empty input")
	}

	in := p[0].(map[string]interface{})

	log.Println("expandStageSpecConfigSystemSync")

	if v, ok := in["git_to_system_sync"].(bool); ok {
		log.Println("expandStageSpecConfigSystemSync git_to_system_sync")
		obj.SystemSync.GitToSystemSync = v
	}

	if v, ok := in["system_to_git_sync"].(bool); ok {
		obj.SystemSync.SystemToGitSync = v
	}

	if v, ok := in["included_resources"].([]interface{}); ok && len(v) > 0 {
		obj.SystemSync.IncludedResources, err = expandIncludedResourceType(v)
		if err != nil {
			return &obj, err
		}
	}

	if v, ok := in["excluded_resources"].([]interface{}); ok && len(v) > 0 {
		obj.SystemSync.ExcludedResources, err = expandExcludedResourceType(v)
		if err != nil {
			return &obj, err
		}
	}

	if v, ok := in["source_repo"].([]interface{}); ok && len(v) > 0 {
		obj.SystemSync.SourceRepo, err = expandRepo(v)
		log.Println("expandStageSpecConfigSystemSync sourceRepo: ", obj.SystemSync.SourceRepo, " ERR ", err)
	}

	if v, ok := in["destination_repo"].([]interface{}); ok && len(v) > 0 {
		obj.SystemSync.DestinationRepo, err = expandRepo(v)
		log.Println("expandStageSpecConfigSystemSync destinationRepo: ", obj.SystemSync.DestinationRepo, " ERR ", err)

	}

	if v, ok := in["source_as_destination"].(bool); ok {
		obj.SystemSync.SourceAsDestination = v
	}

	log.Println("expandStageSpecConfigSystemSync obj: ", obj.SystemSync)

	return &obj, nil
}

func expandStageSpecConfigApproval(p []interface{}) (*gitopspb.StageSpec_Approval, error) {
	obj := gitopspb.StageSpec_Approval{}
	obj.Approval = &gitopspb.ApprovalConfig{}
	if len(p) == 0 || p[0] == nil {
		return &obj, fmt.Errorf("%s", "expandStageSpecConfigApproval empty input")
	}

	in := p[0].(map[string]interface{})

	// Stages    []*StageSpec          `protobuf:"bytes,1,rep,name=stages,proto3" json:"stages,omitempty"`
	if v, ok := in["type"].(string); ok && len(v) > 0 {
		obj.Approval.Type = v
	}

	if v, ok := in["timeout"].(string); ok && len(v) > 0 {
		obj.Approval.Timeout = v
	}

	if v, ok := in["approvers"].([]interface{}); ok && len(v) > 0 {
		obj.Approval.Approvers = expandApprovers(v)
	}

	return &obj, nil
}

func expandStageSpecConfigWorkloadTemplate(p []interface{}) (*gitopspb.StageSpec_WorkloadTemplate, error) {
	obj := gitopspb.StageSpec_WorkloadTemplate{}
	obj.WorkloadTemplate = &gitopspb.DeployWorkloadTemplateConfig{}
	if len(p) == 0 || p[0] == nil {
		return &obj, fmt.Errorf("%s", "expandStageSpecConfigWorkloadTemplate empty input")
	}

	in := p[0].(map[string]interface{})

	if _, ok := in["overrides"].([]interface{}); ok {
		return &obj, fmt.Errorf("%s", "expandStageSpecConfigWorkloadTemplate overrides not supported in terraform")
	}

	if v, ok := in["workload_template"].(string); ok && len(v) > 0 {
		obj.WorkloadTemplate.WorkloadTemplate = v
	}

	if v, ok := in["namespace"].(string); ok && len(v) > 0 {
		obj.WorkloadTemplate.Namespace = v
	}

	if v, ok := in["placement"].([]interface{}); ok {
		obj.WorkloadTemplate.Placement = expandPlacement(v)
	}

	if v, ok := in["use_revision_from_webhook_trigger_event"].(bool); ok {
		obj.WorkloadTemplate.UseRevisionFromWebhookTriggerEvent = v
	}

	return &obj, nil
}

// Expand Stage Spec Start
func expandStageSpec(p []interface{}) ([]*gitopspb.StageSpec, error) {
	if len(p) == 0 || p[0] == nil {
		return []*gitopspb.StageSpec{}, fmt.Errorf("%s", "expandStageSpec empty input")
	}

	out := make([]*gitopspb.StageSpec, len(p))

	for i := range p {
		var stageType string
		obj := gitopspb.StageSpec{}
		in := p[i].(map[string]interface{})

		if v, ok := in["name"].(string); ok && len(v) > 0 {
			obj.Name = v
		}

		if v, ok := in["type"].(string); ok && len(v) > 0 {
			obj.Type = v
			stageType = v
			log.Println("expandStageSpec stageType:", stageType)
		}

		if v, ok := in["pre_conditions"].([]interface{}); ok && len(v) > 0 {
			obj.PreConditions = expandStageSpecPreConditions(v)
		}

		if v, ok := in["variables"].([]interface{}); ok && len(v) > 0 {
			obj.Variables = expandVariableSpec(v)
		}

		if v, ok := in["next"].([]interface{}); ok && len(v) > 0 {
			obj.Next = expandStageSpecNext(v)
		}

		if v, ok := in["config"].([]interface{}); ok && len(v) > 0 {
			log.Println("expandStageSpec config")

			if stageType == "SystemSync" {
				var err error
				obj.Config, err = expandStageSpecConfigSystemSync(v)
				if err != nil {
					log.Println("expandStageSpec SystemSync err ", err)
					return []*gitopspb.StageSpec{}, err
				}
				log.Println("expandStageSpec SystemSync obj.Config ", obj.Config)
			}

			if stageType == "Approval" {
				var err error
				obj.Config, err = expandStageSpecConfigApproval(v)
				if err != nil {
					return []*gitopspb.StageSpec{}, err
				}
			}

			if stageType == "Workload" {
				var err error
				log.Println("expandStageSpec got Workload")
				obj.Config, err = expandStageSpecConfigWorkload(v)
				if err != nil {
					return []*gitopspb.StageSpec{}, err
				}
			}

			if stageType == "WorkloadTemplate" {
				var err error
				log.Println("expandStageSpec got Workload")
				obj.Config, err = expandStageSpecConfigWorkloadTemplate(v)
				if err != nil {
					return []*gitopspb.StageSpec{}, err
				}
			}

		}

		// XXX Debug
		s1 := spew.Sprintf("%+v", obj)
		log.Println("expandStageSpec obj", s1)

		out[i] = &obj

	}

	return out, nil
}

func expandStageSpecPreConditions(p []interface{}) []*gitopspb.PreConditionSpec {
	if len(p) == 0 || p[0] == nil {
		return []*gitopspb.PreConditionSpec{}
	}

	out := make([]*gitopspb.PreConditionSpec, len(p))

	for i := range p {
		//precSpec := stageSpecPreConditionSpec{}
		obj := gitopspb.PreConditionSpec{}
		in := p[i].(map[string]interface{})

		if v, ok := in["type"].(string); ok && len(v) > 0 {
			//precSpec.Type = v
			obj.Type = v
		}
		if vp, ok := in["config"].([]interface{}); ok && len(vp) > 0 {
			if len(vp) == 0 || vp[0] == nil {
				fmt.Printf("%s", "expandStageSpecPreConditions empty ")
			}
			inp := vp[0].(map[string]interface{})

			if v, ok := inp["expression"].(string); ok && len(v) > 0 {
				//precSpec.Config.Expression = v
				obj.Config = nil
			}
		}

		// XXX Debug
		s := spew.Sprintf("%+v", obj)
		log.Println("expandStageSpecPreConditions repoSpec", s)

		out[i] = &obj
	}

	return out

}

func expandStageSpecNext(p []interface{}) []*gitopspb.NextStage {
	if len(p) == 0 || p[0] == nil {
		return []*gitopspb.NextStage{}
	}

	out := make([]*gitopspb.NextStage, len(p))

	for i := range p {
		obj := gitopspb.NextStage{}
		in := p[i].(map[string]interface{})

		if v, ok := in["name"].(string); ok && len(v) > 0 {
			obj.Name = v
		}

		if v, ok := in["weight"].(int32); ok && v > 0 {
			obj.Weight = v
		}

		out[i] = &obj

	}

	return out
}

// func expandIncludedResourceType(p []interface{}) string {
// 	if len(p) == 0 || p[0] == nil {
// 		return ""
// 	}
// 	in := p[0].(map[string]interface{})
// 	if v, ok := in["name"].(string); ok && len(v) > 0 {
// 		log.Println("expandIncludedResourceType name:", v)
// 		return v
// 	}
// 	return ""
// }
// func getConfigType(p []interface{}) string {
// 	if len(p) == 0 || p[0] == nil {
// 		return ""
// 	}
// 	in := p[0].(map[string]interface{})

// 	if v, ok := in["included_resources"].([]interface{}); ok && len(v) > 0 {
// 		log.Println("included_resources")
// 		return expandIncludedResourceType(v)
// 	}
// 	return ""
// }

// // Stage Spec Config
// func expandStageSpecConfig(p []interface{}) stageSpecConfig {
// 	configSpec := stageSpecConfig{}
// 	if len(p) == 0 || p[0] == nil {
// 		return configSpec
// 	}

// 	in := p[0].(map[string]interface{})

// 	if v, ok := in["type"].(string); ok && len(v) > 0 {
// 		configSpec.Type = v
// 	}

// 	if v, ok := in["approvers"].([]interface{}); ok && len(v) > 0 {
// 		configSpec.Approvers = expandStageSpecConfigApprovers(v)
// 	}

// 	if v, ok := in["timeout"].(string); ok && len(v) > 0 {
// 		configSpec.Timeout = v
// 	}

// 	if v, ok := in["workload"].(string); ok && len(v) > 0 {
// 		configSpec.Workload = v
// 	}

// 	if v, ok := in["use_revision_from_webhook_trigger_event"].(bool); ok {
// 		configSpec.UseRevisionFromWebhookTriggerEvent = v
// 	}

// 	if v, ok := in["workload_template"].(string); ok && len(v) > 0 {
// 		configSpec.WorkloadTemplate = v
// 	}

// 	if v, ok := in["namespace"].(string); ok && len(v) > 0 {
// 		configSpec.Namespace = v
// 	}

// 	if v, ok := in["placement"].([]interface{}); ok {
// 		configSpec.Placement = expandPlacement(v)
// 	}

// 	if v, ok := in["overrides"].([]interface{}); ok && len(v) > 0 {
// 		configSpec.Overrides = expandStageSpecConfigWorkloadTemplateOverrides(v)
// 	}

// 	if v, ok := in["use_revision_from_webhook_trigger_event"].(bool); ok {
// 		configSpec.UseRevisionFromWebhookTriggerEvent = v
// 	}

// 	if v, ok := in["provisioner"].(string); ok && len(v) > 0 {
// 		configSpec.Provisioner = v
// 	}

// 	if v, ok := in["revision"].(string); ok && len(v) > 0 {
// 		configSpec.Revision = v
// 	}

// 	if v, ok := in["working_directory"].(string); ok && len(v) > 0 {
// 		configSpec.WorkingDirectory = v
// 	}

// 	if v, ok := in["persist_working_directory"].(bool); ok {
// 		configSpec.PersistWorkingDirectory = v
// 	}

// 	if v, ok := in["agents"].([]interface{}); ok && len(v) > 0 {
// 		configSpec.Agents = expandStageSpecConfigAgents(v)
// 	}

// 	if va, ok := in["action"].([]interface{}); ok && len(va) > 0 {
// 		if len(va) == 0 || va[0] == nil {
// 			fmt.Printf("%s", "expandStageSpecConfigAction empty ")
// 		}
// 		inp := va[0].(map[string]interface{})
// 		if v, ok := inp["action"].(string); ok && len(v) > 0 {
// 			configSpec.Action.Action = v
// 		}

// 		if v, ok := inp["version"].(string); ok && len(v) > 0 {
// 			configSpec.Action.Version = v
// 		}

// 		if v, ok := inp["input_vars"].([]interface{}); ok && len(v) > 0 {
// 			configSpec.Action.InputVars = expandStageSpecKeyValue(v)
// 		}

// 		if v, ok := in["tf_vars_file_path"].([]interface{}); ok {
// 			configSpec.Action.TfVarsFilePath = expandCommonpbFile(v)
// 		}

// 		if v, ok := inp["env_vars"].([]interface{}); ok && len(v) > 0 {
// 			configSpec.Action.EnvVars = expandStageSpecKeyValue(v)
// 		}

// 		if v, ok := inp["backend_vars"].([]interface{}); ok && len(v) > 0 {
// 			configSpec.Action.BackendVars = expandStageSpecKeyValue(v)
// 		}

// 		if v, ok := in["backend_file_path"].([]interface{}); ok {
// 			configSpec.Action.BackendFilePath = expandCommonpbFile(v)
// 		}

// 		if v, ok := in["refresh"].(bool); ok {
// 			configSpec.Action.Refresh = v
// 		}

// 		if v, ok := in["targets"].([]interface{}); ok && len(v) > 0 {
// 			configSpec.Action.Targets = expandStageSpecConfigActionTargets(v)
// 		}

// 		if v, ok := in["destroy"].(bool); ok {
// 			configSpec.Action.Destroy = v
// 		}

// 	}

// 	if v, ok := in["git_to_system_sync"].(bool); ok {
// 		configSpec.GitToSystemSync = v
// 	}

// 	if v, ok := in["system_to_git_sync"].(bool); ok {
// 		configSpec.SystemToGitSync = v
// 	}

// 	if v, ok := in["included_resources"].([]interface{}); ok && len(v) > 0 {
// 		configSpec.IncludedResources = expandStageSpecSystemSyncResources(v)
// 	}

// 	if v, ok := in["excluded_resources"].([]interface{}); ok && len(v) > 0 {
// 		configSpec.ExcludedResources = expandStageSpecSystemSyncResources(v)
// 	}

// 	if v, ok := in["source_repo"].([]interface{}); ok && len(v) > 0 {
// 		configSpec.SourceRepo = expandStageSpecSystemSyncRepo(v)
// 	}

// 	if v, ok := in["destination_repo"].([]interface{}); ok && len(v) > 0 {
// 		configSpec.DestinationRepo = expandStageSpecSystemSyncRepo(v)
// 	}

// 	if v, ok := in["source_as_destination"].(bool); ok {
// 		configSpec.SourceAsDestination = v
// 	}

// 	return configSpec
// }

// func expandStageSpecConfigApprovers(p []interface{}) []*gitopspb.Approver {
// 	if len(p) == 0 || p[0] == nil {
// 		return []*gitopspb.Approver{}
// 	}
// 	out := make([]*gitopspb.Approver, len(p))

// 	for i := range p {
// 		obj := gitopspb.Approver{}
// 		in := p[i].(map[string]interface{})

// 		if v, ok := in["user_name"].(string); ok && len(v) > 0 {
// 			obj.UserName = v
// 		}

// 		if v, ok := in["sso_user"].(bool); ok {
// 			obj.SsoUser = v
// 		}

// 		out[i] = &obj

// 	}

// 	return out

// }

// func expandStageSpecConfigWorkloadTemplateOverrides(p []interface{}) []stageSpecConfigWorkloadTemplateOverrides {
// 	if len(p) == 0 || p[0] == nil {
// 		return []stageSpecConfigWorkloadTemplateOverrides{}
// 	}

// 	out := make([]stageSpecConfigWorkloadTemplateOverrides, len(p))

// 	for i := range p {
// 		obj := stageSpecConfigWorkloadTemplateOverrides{}
// 		in := p[i].(map[string]interface{})

// 		if v, ok := in["type"].(string); ok && len(v) > 0 {
// 			obj.Type = v
// 		}

// 		if v, ok := in["weight"].(int32); ok && v > 0 {
// 			obj.Weight = v
// 		}

// 		if vp, ok := in["template"].([]interface{}); ok && len(vp) > 0 {
// 			if len(vp) == 0 || vp[0] == nil {
// 				fmt.Printf("%s", "expandStageSpecConfigWorkloadTemplateOverridesTemplate empty ")
// 			}
// 			inp := vp[0].(map[string]interface{})
// 			if v, ok := inp["inline"].(string); ok && len(v) > 0 {
// 				obj.Template.Inline = v
// 			}

// 			if v, ok := inp["repository"].(string); ok && len(v) > 0 {
// 				obj.Template.Repository = v
// 			}

// 			if v, ok := inp["inline"].(string); ok && len(v) > 0 {
// 				obj.Template.Revision = v
// 			}

// 			if v, ok := in["paths"].([]interface{}); ok {
// 				obj.Template.Paths = expandCommonpbFiles(v)
// 			}

// 		}

// 		out[i] = obj

// 	}

// 	return out
// }

// func expandStageSpecConfigAgents(p []interface{}) []*gitopspb.AgentMeta {
// 	if len(p) == 0 || p[0] == nil {
// 		return []*gitopspb.AgentMeta{}
// 	}
// 	out := make([]*gitopspb.AgentMeta, len(p))

// 	for i := range p {
// 		obj := gitopspb.AgentMeta{}
// 		in := p[i].(map[string]interface{})

// 		if v, ok := in["name"].(string); ok && len(v) > 0 {
// 			obj.Name = v
// 		}

// 		if v, ok := in["id"].(string); ok && len(v) > 0 {
// 			obj.Id = v
// 		}

// 		out[i] = &obj

// 	}

// 	return out

// }

// func expandStageSpecKeyValue(p []interface{}) []*gitopspb.KeyValue {
// 	if len(p) == 0 || p[0] == nil {
// 		return nil
// 	}

// 	out := make([]*gitopspb.KeyValue, len(p))

// 	for i := range p {
// 		obj := gitopspb.KeyValue{}
// 		in := p[i].(map[string]interface{})

// 		if v, ok := in["key"].(string); ok && len(v) > 0 {
// 			obj.Key = v
// 		}

// 		if v, ok := in["value"].(string); ok && len(v) > 0 {
// 			obj.Value = v
// 		}

// 		if v, ok := in["type"].(string); ok && len(v) > 0 {
// 			obj.Type = v
// 		}
// 		out[i] = &obj
// 	}
// 	return out

// }

// func expandStageSpecSystemSyncResources(p []interface{}) []*gitopspb.SystemSyncResource {
// 	if len(p) == 0 || p[0] == nil {
// 		return nil
// 	}

// 	out := make([]*gitopspb.SystemSyncResource, len(p))

// 	for i := range p {
// 		obj := gitopspb.SystemSyncResource{}
// 		in := p[i].(map[string]interface{})

// 		if v, ok := in["name"].(string); ok && len(v) > 0 {
// 			obj.Name = v
// 		}

// 		out[i] = &obj
// 	}
// 	return out

// }

// func expandStageSpecSystemSyncRepo(p []interface{}) *gitopspb.SystemSyncRepo {
// 	obj := &gitopspb.SystemSyncRepo{}
// 	if len(p) == 0 || p[0] == nil {
// 		return obj
// 	}

// 	in := p[0].(map[string]interface{})
// 	if v, ok := in["repository"].(string); ok && len(v) > 0 {
// 		obj.Repository = v
// 	}

// 	if v, ok := in["revision"].(string); ok && len(v) > 0 {
// 		obj.Revision = v
// 	}

// 	if v, ok := in["path"].([]interface{}); ok {
// 		obj.Path = expandCommonpbFile(v)
// 	}

// 	return obj

// }

// func expandStageSpecConfigActionTargets(p []interface{}) []*gitopspb.TerraformTarget {
// 	if len(p) == 0 || p[0] == nil {
// 		return []*gitopspb.TerraformTarget{}
// 	}
// 	out := make([]*gitopspb.TerraformTarget, len(p))

// 	for i := range p {
// 		obj := gitopspb.TerraformTarget{}
// 		in := p[i].(map[string]interface{})

// 		if v, ok := in["name"].(string); ok && len(v) > 0 {
// 			obj.Name = v
// 		}

// 		out[i] = &obj

// 	}

// 	return out

// }

// Stage Spec Expand End

// Variable Spec Expand Start
func expandVariableSpec(p []interface{}) []*gitopspb.VariableSpec {
	if len(p) == 0 || p[0] == nil {
		return []*gitopspb.VariableSpec{}
	}

	out := make([]*gitopspb.VariableSpec, len(p))

	for i := range p {
		obj := gitopspb.VariableSpec{}
		in := p[i].(map[string]interface{})

		if v, ok := in["name"].(string); ok && len(v) > 0 {
			obj.Name = v
		}

		if v, ok := in["type"].(string); ok && len(v) > 0 {
			obj.Type = v
		}

		if v, ok := in["value"].(string); ok && len(v) > 0 {
			obj.Value = v
		}

		out[i] = &obj

	}

	return out
}

// Variable Spec Expand End
func expandWebhookTriggerHelm(p []interface{}) *gitopspb.WebhookTriggerConfig_Helm {
	obj := &gitopspb.WebhookTriggerConfig_Helm{}
	obj.Helm = &gitopspb.HelmRepoConfig{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["chart_name"].(string); ok && len(v) > 0 {
		obj.Helm.ChartName = v
	}

	if v, ok := in["chart_version"].(string); ok && len(v) > 0 {
		obj.Helm.ChartVersion = v
	}

	if v, ok := in["repository"].(string); ok && len(v) > 0 {
		obj.Helm.Repository = v
	}

	return obj
}

func expandWebhookTriggerGit(p []interface{}) *gitopspb.WebhookTriggerConfig_Git {
	obj := &gitopspb.WebhookTriggerConfig_Git{}
	obj.Git = &gitopspb.GitRepoConfig{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["provider"].(string); ok && len(v) > 0 {
		obj.Git.Provider = v
	}

	if v, ok := in["repository"].(string); ok && len(v) > 0 {
		obj.Git.Repository = v
	}

	if v, ok := in["revision"].(string); ok && len(v) > 0 {
		obj.Git.Revision = v
	}

	if v, ok := in["provider"].(string); ok && len(v) > 0 {
		obj.Git.Paths = expandCommonpbFiles(p)
	}

	return obj
}

func expandWebhookTrigger(p []interface{}) *gitopspb.WebhookTriggerConfig {
	obj := &gitopspb.WebhookTriggerConfig{}

	if len(p) == 0 || p[0] == nil {
		return obj
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["provider"].(string); ok && len(v) > 0 {
		if v == "Github" || v == "Gitlab" || v == "Bitbucket" || v == "AzureRepos" {
			obj.Repo = expandWebhookTriggerGit(p)
		}

		if v == "Helm" {
			obj.Repo = expandWebhookTriggerGit(p)
		}
	}
	return obj
}

func expandCronTriggerHelm(p []interface{}) *gitopspb.CronTriggerConfig_Helm {
	obj := &gitopspb.CronTriggerConfig_Helm{}
	obj.Helm = &gitopspb.HelmRepoConfig{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["chart_name"].(string); ok && len(v) > 0 {
		obj.Helm.ChartName = v
	}

	if v, ok := in["chart_version"].(string); ok && len(v) > 0 {
		obj.Helm.ChartVersion = v
	}

	if v, ok := in["repository"].(string); ok && len(v) > 0 {
		obj.Helm.Repository = v
	}

	return obj
}

func expandCronTriggerGit(p []interface{}) *gitopspb.CronTriggerConfig_Git {
	obj := &gitopspb.CronTriggerConfig_Git{}
	obj.Git = &gitopspb.GitRepoConfig{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["provider"].(string); ok && len(v) > 0 {
		obj.Git.Provider = v
	}

	if v, ok := in["repository"].(string); ok && len(v) > 0 {
		obj.Git.Repository = v
	}

	if v, ok := in["revision"].(string); ok && len(v) > 0 {
		obj.Git.Revision = v
	}

	if v, ok := in["provider"].(string); ok && len(v) > 0 {
		obj.Git.Paths = expandCommonpbFiles(p)
	}

	return obj
}

func expandCronTrigger(p []interface{}) *gitopspb.CronTriggerConfig {
	obj := &gitopspb.CronTriggerConfig{}

	if len(p) == 0 || p[0] == nil {
		return obj
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["provider"].(string); ok && len(v) > 0 {
		if v == "Github" {
			obj.Repo = expandCronTriggerGit(p)
		}

		if v == "Helm" {
			obj.Repo = expandCronTriggerHelm(p)
		}
	}
	return obj
}

func expandTriggerWebhookSpec(p []interface{}) *gitopspb.TriggerSpec_Webhook {
	obj := &gitopspb.TriggerSpec_Webhook{}
	obj.Webhook = &gitopspb.WebhookTriggerConfig{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["repo"].([]interface{}); ok && len(v) > 0 {
		obj.Webhook = expandWebhookTrigger(v)
	}

	return obj
}

func expandTriggerCronSpec(p []interface{}) *gitopspb.TriggerSpec_Cron {
	obj := &gitopspb.TriggerSpec_Cron{}
	obj.Cron = &gitopspb.CronTriggerConfig{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["cron_expression"].(string); ok && len(v) > 0 {
		obj.Cron.CronExpression = v
	}

	if v, ok := in["cron_expression"].(string); ok && len(v) > 0 {
		obj.Cron.CronExpression = v
	}

	if v, ok := in["repo"].([]interface{}); ok && len(v) > 0 {
		obj.Cron = expandCronTrigger(v)
	}

	return obj
}

// Trigger Spec Expand Start
func expandTriggerSpec(p []interface{}) []*gitopspb.TriggerSpec {
	if len(p) == 0 || p[0] == nil {
		return []*gitopspb.TriggerSpec{}
	}

	out := make([]*gitopspb.TriggerSpec, len(p))

	for i := range p {
		obj := gitopspb.TriggerSpec{}
		in := p[i].(map[string]interface{})

		if v, ok := in["type"].(string); ok && len(v) > 0 {
			obj.Type = v
		}

		if v, ok := in["name"].(string); ok && len(v) > 0 {
			obj.Name = v
		}

		if v, ok := in["variables"].([]interface{}); ok && len(v) > 0 {
			obj.Variables = expandVariableSpec(v)
		}

		if vp, ok := in["config"].([]interface{}); ok && len(vp) > 0 {
			if obj.Type == "Webhook" {
				obj.Config = expandTriggerWebhookSpec(vp)
			}

			if obj.Type == "Chron" {
				obj.Config = expandTriggerCronSpec(vp)
			}

			// if len(vp) == 0 || vp[0] == nil {
			// 	return []*gitopspb.TriggerSpec{}
			// }
			// inp := vp[0].(map[string]interface{})

			// if vpo, ok := in["repo"].([]interface{}); ok && len(vpo) > 0 {
			// 	if len(vpo) == 0 || vpo[0] == nil {
			// 		return []*gitopspb.TriggerSpec{}
			// 	}
			// 	inpo := vpo[0].(map[string]interface{})

			// 	if v, ok := inpo["provider"].(string); ok && len(v) > 0 {
			// 		obj.Config.Repo.Provider = v
			// 	}

			// 	if v, ok := inpo["repository"].(string); ok && len(v) > 0 {
			// 		obj.Config.Repo.Repository = v
			// 	}

			// 	if v, ok := inpo["revision"].(string); ok && len(v) > 0 {
			// 		obj.Config.Repo.Revision = v
			// 	}

			// 	if v, ok := inpo["paths"].([]interface{}); ok {
			// 		//obj.Secret = expandCommonpbFile(v)
			// 		obj.Config.Repo.Paths = expandCommonpbFiles(v)
			// 	}

			// 	if v, ok := inpo["chart_name"].(string); ok && len(v) > 0 {
			// 		obj.Config.Repo.ChartName = v
			// 	}

			// 	if v, ok := inpo["chart_version"].(string); ok && len(v) > 0 {
			// 		obj.Config.Repo.ChartVersion = v
			// 	}
			// }

			// if v, ok := inp["cron_expression"].(string); ok && len(v) > 0 {
			// 	obj.Config.CronExpression = v
			// }
		}

		// log.Println("expandTriggerSpec jsonSpec ", string(jsonSpec))

		// err = trueObj.UnmarshalJSON(jsonSpec)
		// if err != nil {
		// 	log.Println("expandTriggerSpec UnmarshalJSON error ", err)
		// 	return nil
		// }

		out[i] = &obj

	}
	return out
}

// Flatteners

func flattenPipeline(d *schema.ResourceData, in *gitopspb.Pipeline) error {
	if in == nil {
		return nil
	}

	err := d.Set("metadata", flattenMetaData(in.Metadata))
	if err != nil {
		return err
	}

	v, ok := d.Get("spec").([]interface{})
	if !ok {
		v = []interface{}{}
	}

	var ret []interface{}
	ret, err = flattenPipelineSpec(in.Spec, v)
	if err != nil {
		return err
	}

	err = d.Set("spec", ret)
	if err != nil {
		return err
	}
	return nil
}

func flattenPipelineSpec(in *gitopspb.PipelineSpec, p []interface{}) ([]interface{}, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "flattenPipeline empty input")
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	// Stages    []*StageSpec          `protobuf:"bytes,1,rep,name=stages,proto3" json:"stages,omitempty"`
	if in.Stages != nil && len(in.Stages) > 0 {
		v, ok := obj["stages"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["stages"] = flattenStageSpec(in.Stages, v)
	}

	// Variables []*VariableSpec       `protobuf:"bytes,2,rep,name=variables,proto3" json:"variables,omitempty"`
	if in.Variables != nil && len(in.Variables) > 0 {
		v, ok := obj["variables"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["variables"] = flattenVariableSpec(in.Variables, v)
	}
	// Triggers  []*TriggerSpec        `protobuf:"bytes,3,rep,name=triggers,proto3" json:"triggers,omitempty"`
	if in.Triggers != nil && len(in.Triggers) > 0 {
		v, ok := obj["triggers"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["triggers"] = flattenTriggerSpec(in.Triggers, v)
	}

	if in.Sharing != nil {
		obj["sharing"] = flattenSharingSpec(in.Sharing)
	}

	obj["active"] = in.Active

	if in.Secret != nil {
		obj["secrets"] = flattenCommonpbFile(in.Secret)
	}

	return []interface{}{obj}, nil
}

func flattenStageSpec(input []*gitopspb.StageSpec, p []interface{}) []interface{} {
	log.Println("flattenStageSpec")
	if input == nil {
		return nil
	}

	out := make([]interface{}, len(input))
	for i, in := range input {
		log.Println("flattenStageSpec in ", in)

		jsonBytes, err := in.MarshalJSON()
		if err != nil {
			log.Println("flattenStageSpec MarshalJSON error", err)
			return nil
		}
		log.Println("flattenStageSpec jsonBytes ", string(jsonBytes))

		stSpec := stageSpec{}
		err = json.Unmarshal(jsonBytes, &stSpec)

		obj := map[string]interface{}{}
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}

		if len(in.Name) > 0 {
			obj["name"] = in.Name
		}

		if len(in.Type) > 0 {
			obj["type"] = in.Type
		}

		//PRE CONDITION FLATTEN
		v, ok := obj["pre_conditions"].([]interface{})
		if !ok {
			v = []interface{}{}
		}

		// XXX Debug
		w1 := spew.Sprintf("%+v", v)
		log.Println("flattenPreConditions before ", w1)

		var ret []interface{}
		ret, err = flattenPreConditions(&stSpec, v)
		if err != nil {
			log.Println("flattenPreConditions error ", err)
			return nil
		}

		// XXX Debug
		w1 = spew.Sprintf("%+v", ret)
		log.Println("flattenPreConditions after ", w1)

		obj["pre_conditions"] = ret

		//Variables Flatten
		if len(in.Variables) > 0 {
			v, ok := obj["variables"].([]interface{})
			if !ok {
				v = []interface{}{}
			}

			obj["variables"] = flattenVariableSpec(in.Variables, v)
		} else {
			obj["variables"] = nil
		}

		//Next Flatten
		if len(in.Next) > 0 {
			v, ok := obj["next"].([]interface{})
			if !ok {
				v = []interface{}{}
			}

			obj["next"] = flattenNextSpec(in.Next, v)
		} else {
			obj["next"] = nil
		}

		//Stage Spec Config Flatten
		v2, ok := obj["config"].([]interface{})
		if !ok {
			v2 = []interface{}{}
		}

		// XXX Debug
		w2 := spew.Sprintf("%+v", v2)
		log.Println("flattenStageSpecConfig before ", w2)

		var ret2 []interface{}
		ret2, err = flattenStageSpecConfig(&stSpec, v2)
		if err != nil {
			log.Println("flattenStageSpecConfig error ", err)
			return nil
		}

		// XXX Debug
		w2 = spew.Sprintf("%+v", ret2)
		log.Println("flattenStageSpecConfig after ", w2)

		obj["config"] = ret2

		//Put together

		out[i] = &obj
	}

	return out
}

func flattenPreConditions(stSpec *stageSpec, p []interface{}) ([]interface{}, error) {
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	retNil := true

	out := make([]interface{}, len(stSpec.PreConditions))
	for i, in := range stSpec.PreConditions {

		obj := map[string]interface{}{}
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}

		if len(in.Type) > 0 {
			obj["type"] = in.Type
			retNil = false
		}

		if len(in.Config.Expression) > 0 {
			obj["expression"] = in.Config.Expression
			retNil = false
		}

		out[i] = &obj
	}

	if retNil {
		return nil, nil
	}

	return []interface{}{obj}, nil

}

func flattenVariableSpec(input []*gitopspb.VariableSpec, p []interface{}) []interface{} {
	log.Println("flattenVariableSpec")
	if input == nil {
		return nil
	}

	out := make([]interface{}, len(input))
	for i, in := range input {
		log.Println("flattenVariableSpec in ", in)
		obj := map[string]interface{}{}
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}

		if len(in.Name) > 0 {
			obj["name"] = in.Name
		}

		if len(in.Type) > 0 {
			obj["type"] = in.Type
		}

		if len(in.Value) > 0 {
			obj["value"] = in.Value
		}

		out[i] = &obj
	}

	return out
}

func flattenNextSpec(input []*gitopspb.NextStage, p []interface{}) []interface{} {
	log.Println("flattenNextSpec")
	if input == nil {
		return nil
	}

	out := make([]interface{}, len(input))
	for i, in := range input {
		log.Println("flattenNextSpec in ", in)
		obj := map[string]interface{}{}
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}

		if len(in.Name) > 0 {
			obj["name"] = in.Name
		}

		obj["weight"] = in.Weight

		out[i] = &obj
	}

	return out
}

func flattenStageSpecConfig(stSpec *stageSpec, p []interface{}) ([]interface{}, error) {
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	retNil := true

	// var start

	if len(stSpec.Config.Type) > 0 {
		obj["type"] = stSpec.Config.Type
		retNil = false
	}

	if len(stSpec.Config.Approvers) > 0 {
		v, ok := obj["approvers"].([]interface{})
		if !ok {
			v = []interface{}{}
		}

		obj["approvers"] = flattenStageSpecApprovers(stSpec.Config.Approvers, v)
	} else {
		obj["approvers"] = nil
	}

	if len(stSpec.Config.Timeout) > 0 {
		obj["timeout"] = stSpec.Config.Timeout
		retNil = false
	}

	if len(stSpec.Config.Workload) > 0 {
		obj["workload"] = stSpec.Config.Workload
		retNil = false
	}

	if len(stSpec.Config.WorkloadTemplate) > 0 {
		obj["workload_template"] = stSpec.Config.WorkloadTemplate
		retNil = false
	}

	if len(stSpec.Config.Namespace) > 0 {
		obj["namespace"] = stSpec.Config.Namespace
		retNil = false
	}

	if stSpec.Config.Placement != nil {
		obj["placement"] = flattenPlacement(stSpec.Config.Placement)
	}

	if len(stSpec.Config.Overrides) > 0 {
		v, ok := obj["overrides"].([]interface{})
		if !ok {
			v = []interface{}{}
		}

		obj["overrides"] = flattenStageSpecOverrides(stSpec.Config.Overrides, v)
	} else {
		obj["overrides"] = nil
	}

	if stSpec.Config.UseRevisionFromWebhookTriggerEvent {
		obj["use_revision_from_webhook_trigger_event"] = stSpec.Config.UseRevisionFromWebhookTriggerEvent
		retNil = false
	}

	if len(stSpec.Config.Provisioner) > 0 {
		obj["provisioner"] = stSpec.Config.Provisioner
		retNil = false
	}

	if len(stSpec.Config.Revision) > 0 {
		obj["revision"] = stSpec.Config.Revision
		retNil = false
	}

	if len(stSpec.Config.WorkingDirectory) > 0 {
		obj["working_directory"] = stSpec.Config.WorkingDirectory
		retNil = false
	}

	if stSpec.Config.PersistWorkingDirectory {
		obj["persist_working_directory"] = stSpec.Config.PersistWorkingDirectory
		retNil = false
	}

	if len(stSpec.Config.Agents) > 0 {
		v, ok := obj["agents"].([]interface{})
		if !ok {
			v = []interface{}{}
		}

		obj["agents"] = flattenStageSpecAgents(stSpec.Config.Agents, v)
	} else {
		obj["agents"] = nil
	}

	obj["action"] = flattenStageSpecAction(stSpec)

	if stSpec.Config.GitToSystemSync {
		obj["git_to_system_sync"] = stSpec.Config.GitToSystemSync
		retNil = false
	}

	if stSpec.Config.SystemToGitSync {
		obj["system_to_git_sync"] = stSpec.Config.SystemToGitSync
		retNil = false
	}

	if len(stSpec.Config.IncludedResources) > 0 {
		v, ok := obj["included_resources"].([]interface{})
		if !ok {
			v = []interface{}{}
		}

		obj["included_resources"] = flattenStageSpecSystemSyncResource(stSpec.Config.IncludedResources, v)
	} else {
		obj["included_resources"] = nil
	}

	if len(stSpec.Config.ExcludedResources) > 0 {
		v, ok := obj["excluded_resources"].([]interface{})
		if !ok {
			v = []interface{}{}
		}

		obj["excluded_resources"] = flattenStageSpecSystemSyncResource(stSpec.Config.ExcludedResources, v)
	} else {
		obj["excluded_resources"] = nil
	}

	if stSpec.Config.SourceRepo != nil {
		obj["source_repo"] = flattenStageSpecSystemSyncRepo(stSpec.Config.SourceRepo)
	}

	if stSpec.Config.DestinationRepo != nil {
		obj["destination_repo"] = flattenStageSpecSystemSyncRepo(stSpec.Config.DestinationRepo)
	}

	if stSpec.Config.SourceAsDestination {
		obj["source_as_destination"] = stSpec.Config.SourceAsDestination
		retNil = false
	}

	if retNil {
		return nil, nil
	}

	return []interface{}{obj}, nil

}

func flattenStageSpecApprovers(input []*gitopspb.Approver, p []interface{}) []interface{} {
	log.Println("flattenStageSpecApprovers")
	if input == nil {
		return nil
	}

	out := make([]interface{}, len(input))
	for i, in := range input {
		log.Println("flattenStageSpecApprovers in ", in)
		obj := map[string]interface{}{}
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}

		if len(in.UserName) > 0 {
			obj["user_name"] = in.UserName
		}

		if in.SsoUser {
			obj["sso_user"] = in.SsoUser
		}

		out[i] = &obj
	}

	return out
}

func flattenStageSpecOverrides(input []stageSpecConfigWorkloadTemplateOverrides, p []interface{}) []interface{} {
	log.Println("flattenStageSpecOverrides")
	if input == nil {
		return nil
	}

	out := make([]interface{}, len(input))
	for i, in := range input {
		log.Println("flattenStageSpecOverrides in ", in)
		obj := map[string]interface{}{}
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}

		if len(in.Type) > 0 {
			obj["type"] = in.Type
		}

		obj["weight"] = in.Weight

		obj["template"] = flattenStageSpecConfigOverridesTemplate(in)

		out[i] = &obj
	}

	return out
}

func flattenStageSpecConfigOverridesTemplate(in stageSpecConfigWorkloadTemplateOverrides) []interface{} {
	retNil := true
	obj := make(map[string]interface{})

	if len(in.Template.Inline) > 0 {
		obj["inline"] = in.Template.Inline
		retNil = false
	}

	if len(in.Template.Repository) > 0 {
		obj["repository"] = in.Template.Repository
		retNil = false

	}

	if len(in.Template.Revision) > 0 {
		obj["revision"] = in.Template.Revision
		retNil = false

	}

	if in.Template.Paths != nil {
		obj["paths"] = flattenCommonpbFiles(in.Template.Paths)
		retNil = false
	}

	if retNil {
		return nil
	}

	return []interface{}{obj}
}

func flattenStageSpecAgents(input []*gitopspb.AgentMeta, p []interface{}) []interface{} {
	log.Println("flattenStageSpecAgents")
	if input == nil {
		return nil
	}

	out := make([]interface{}, len(input))
	for i, in := range input {
		log.Println("flattenStageSpecAgents in ", in)
		obj := map[string]interface{}{}
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}

		if len(in.Name) > 0 {
			obj["name"] = in.Name
		}

		if len(in.Name) > 0 {
			obj["id"] = in.Id
		}

		out[i] = &obj
	}

	return out
}

func flattenStageSpecAction(in *stageSpec) []interface{} {
	if in == nil {
		return nil
	}
	retNil := true
	obj := make(map[string]interface{})

	if len(in.Config.Action.Action) > 0 {
		obj["action"] = (in.Config.Action.Action)
		retNil = false
	}

	if len(in.Config.Action.Version) > 0 {
		obj["version"] = (in.Config.Action.Version)
		retNil = false
	}

	if len(in.Config.Action.InputVars) > 0 {
		retNil = false
		v, ok := obj["input_vars"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["input_vars"] = flattenStageSpecConfigActionKeyValue(in.Config.Action.InputVars, v)
	} else {
		obj["input_vars"] = nil
		retNil = false
	}

	if in.Config.Action.TfVarsFilePath != nil {
		obj["tf_vars_file_path"] = flattenCommonpbFile(in.Config.Action.TfVarsFilePath)
		retNil = false
	}

	if len(in.Config.Action.EnvVars) > 0 {
		retNil = false
		v, ok := obj["input_vars"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["env_vars"] = flattenStageSpecConfigActionKeyValue(in.Config.Action.InputVars, v)
	} else {
		obj["env_vars"] = nil
		retNil = false
	}

	if len(in.Config.Action.BackendVars) > 0 {
		retNil = false
		v, ok := obj["backend_vars"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["backend_vars"] = flattenStageSpecConfigActionKeyValue(in.Config.Action.InputVars, v)
	} else {
		obj["backend_vars"] = nil
		retNil = false
	}

	if in.Config.Action.BackendFilePath != nil {
		obj["backend_file_path"] = flattenCommonpbFile(in.Config.Action.BackendFilePath)
		retNil = false
	}

	if in.Config.Action.Refresh {
		obj["refresh"] = in.Config.Action.Refresh
		retNil = false
	}

	if len(in.Config.Action.Targets) > 0 {
		retNil = false
		v, ok := obj["targets"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["targets"] = flattenStageSpecConfigActionTargets(in.Config.Action.Targets, v)
	} else {
		obj["targets"] = nil
		retNil = false
	}

	if in.Config.Action.Destroy {
		obj["destroy"] = in.Config.Action.Destroy
		retNil = false
	}

	if retNil {
		return nil
	}

	return []interface{}{obj}
}

func flattenStageSpecConfigActionKeyValue(input []*gitopspb.KeyValue, p []interface{}) []interface{} {
	log.Println("flattenStageSpecConfigActionKeyValue")
	if input == nil {
		return nil
	}

	out := make([]interface{}, len(input))
	for i, in := range input {
		log.Println("flattenStageSpecConfigActionKeyValue in ", in)
		obj := map[string]interface{}{}
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}

		if len(in.Key) > 0 {
			obj["key"] = in.Key
		}

		if len(in.Value) > 0 {
			obj["value"] = in.Value
		}

		if len(in.Type) > 0 {
			obj["Type"] = in.Type
		}

		out[i] = &obj
	}

	return out
}

func flattenStageSpecConfigActionTargets(input []*gitopspb.TerraformTarget, p []interface{}) []interface{} {
	log.Println("flattenStageSpecConfigActionTargets")
	if input == nil {
		return nil
	}

	out := make([]interface{}, len(input))
	for i, in := range input {
		log.Println("flattenStageSpecConfigActionTargets in ", in)
		obj := map[string]interface{}{}
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}

		if len(in.Name) > 0 {
			obj["name"] = in.Name
		}
		out[i] = &obj
	}

	return out
}

func flattenStageSpecSystemSyncResource(input []*gitopspb.SystemSyncResource, p []interface{}) []interface{} {
	log.Println("flattenStageSpecSystemSyncResource")
	if input == nil {
		return nil
	}

	out := make([]interface{}, len(input))
	for i, in := range input {
		log.Println("flattenStageSpecSystemSyncResource in ", in)
		obj := map[string]interface{}{}
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}

		if len(in.Name) > 0 {
			obj["name"] = in.Name
		}

		out[i] = &obj
	}

	return out
}

func flattenStageSpecSystemSyncRepo(in *gitopspb.SystemSyncRepo) []interface{} {
	if in == nil {
		return nil
	}
	retNil := true
	obj := make(map[string]interface{})

	if len(in.Repository) > 0 {
		obj["pod"] = (in.Repository)
		retNil = false
	}

	if len(in.Revision) > 0 {
		obj["name"] = (in.Revision)
		retNil = false
	}

	if in.Path != nil {
		obj["paths"] = flattenCommonpbFile(in.Path)
		retNil = false
	}

	if retNil {
		return nil
	}

	return []interface{}{obj}
}

func flattenTriggerSpec(input []*gitopspb.TriggerSpec, p []interface{}) []interface{} {
	if input == nil {
		return nil
	}

	out := make([]interface{}, len(input))
	for i, in := range input {
		log.Println("flattenTriggerSpec in ", in)

		jsonBytes, err := in.MarshalJSON()
		if err != nil {
			log.Println("flattenTriggerSpec MarshalJSON error", err)
			return nil
		}
		log.Println("flattenTriggerSpec jsonBytes ", string(jsonBytes))

		tSpec := triggerSpec{}
		err = json.Unmarshal(jsonBytes, &tSpec)

		obj := map[string]interface{}{}
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}

		if len(in.Type) > 0 {
			obj["type"] = in.Type
		}

		if len(in.Name) > 0 {
			obj["name"] = in.Name
		}

		if len(in.Variables) > 0 {
			v, ok := obj["variables"].([]interface{})
			if !ok {
				v = []interface{}{}
			}

			obj["variables"] = flattenVariableSpec(in.Variables, v)
		} else {
			obj["variables"] = nil
		}

		v, ok := obj["config"].([]interface{})
		if !ok {
			v = []interface{}{}
		}

		// XXX Debug
		w1 := spew.Sprintf("%+v", v)
		log.Println("flattenTriggerConfig before ", w1)

		var ret []interface{}
		ret, err = flattenTriggerConfig(&tSpec, v)
		if err != nil {
			log.Println("flattenTriggerConfig error ", err)
			return nil
		}

		// XXX Debug
		w1 = spew.Sprintf("%+v", ret)
		log.Println("flattenTriggerConfig after ", w1)

		obj["config"] = ret
		out[i] = &obj
	}

	return []interface{}{out}
}

func flattenTriggerConfig(tSpec *triggerSpec, p []interface{}) ([]interface{}, error) {
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	retNil := true
	if len(tSpec.Config.Repo.Provider) > 0 {
		obj["provider"] = tSpec.Config.Repo.Provider
		retNil = false
	}

	if len(tSpec.Config.Repo.Repository) > 0 {
		obj["repository"] = tSpec.Config.Repo.Repository
		retNil = false
	}

	if len(tSpec.Config.Repo.Revision) > 0 {
		obj["revision"] = tSpec.Config.Repo.Revision
		retNil = false
	}

	if tSpec.Config.Repo.Paths != nil {
		obj["paths"] = flattenCommonpbFiles(tSpec.Config.Repo.Paths)
	}

	if len(tSpec.Config.Repo.ChartName) > 0 {
		obj["chart_name"] = tSpec.Config.Repo.ChartName
		retNil = false
	}

	if len(tSpec.Config.Repo.ChartVersion) > 0 {
		obj["chart_version"] = tSpec.Config.Repo.ChartVersion
		retNil = false
	}

	if len(tSpec.Config.CronExpression) > 0 {
		obj["cron_expression"] = tSpec.Config.CronExpression
		retNil = false
	}

	if retNil {
		return nil, nil
	}

	return []interface{}{obj}, nil

}
