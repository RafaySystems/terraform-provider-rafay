package rafay

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/RafaySystems/rafay-common/proto/types/hub/commonpb"
	"github.com/RafaySystems/rctl/pkg/clusteroverride"
	"github.com/RafaySystems/rctl/pkg/commands"
	"github.com/RafaySystems/rctl/pkg/config"
	"github.com/RafaySystems/rctl/pkg/models"
	"github.com/RafaySystems/rctl/pkg/project"
	"github.com/RafaySystems/rctl/utils"
	"github.com/davecgh/go-spew/spew"
	"gopkg.in/yaml.v3"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type clusterOverrideYamlConfig struct {
	Kind       string                     `json:"kind,omitempty" yaml:"kind"`
	ApiVersion string                     `json:"apiversion,omitempty" yaml:"apiVersion"`
	Metadata   *commonpb.Metadata         `json:"metadata,omitempty" yaml:"metadata"`
	Spec       models.ClusterOverrideSpec `json:"spec,omitempty" yaml:"spec"`
}

func resourceClusterOverride() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceClusterOverrideCreate,
		ReadContext:   resourceClusterOverrideRead,
		UpdateContext: resourceClusterOverrideUpdate,
		DeleteContext: resourceClusterOverrideDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"metadata": &schema.Schema{
				Description: "Metadata of the addon resource",
				Elem: &schema.Resource{Schema: map[string]*schema.Schema{
					"annotations": &schema.Schema{
						Description: "annotations of the resource",
						Elem:        &schema.Schema{Type: schema.TypeString},
						Optional:    true,
						Type:        schema.TypeMap,
					},
					"description": &schema.Schema{
						Description: "description of the resource",
						Optional:    true,
						Type:        schema.TypeString,
					},
					"labels": &schema.Schema{
						Description: "labels of the resource",
						Elem:        &schema.Schema{Type: schema.TypeString},
						Optional:    true,
						Type:        schema.TypeMap,
					},
					"name": &schema.Schema{
						Description: "name of the resource",
						Optional:    true,
						Type:        schema.TypeString,
					},
					"project": &schema.Schema{
						Description: "Project of the resource",
						Optional:    true,
						Type:        schema.TypeString,
					},
				}},
				MaxItems: 1,
				MinItems: 1,
				Optional: true,
				Type:     schema.TypeList,
			},
			"spec": &schema.Schema{
				Description: "override specification",
				Elem: &schema.Resource{Schema: map[string]*schema.Schema{
					"cluster_selector": &schema.Schema{
						Description: "cluster selector",
						Optional:    true,
						Type:        schema.TypeString,
					},
					"resource_selector": &schema.Schema{
						Description: "resource selector",
						Optional:    true,
						Type:        schema.TypeString,
					},
					"type": &schema.Schema{
						Description: "override type, accepted values are *ClusterOverrideTypeoverride*, *ClusterOverrideTypeAddon*, *ClusterOverrideTypeNamespace*, *ClusterOverrideTypeBlueprint*",
						Optional:    true,
						Type:        schema.TypeString,
					},
					"override_values": &schema.Schema{
						Description: "override value",
						Optional:    true,
						Type:        schema.TypeString,
					},
					"value_repo_ref": &schema.Schema{
						Description: "value repo ref",
						Optional:    true,
						Type:        schema.TypeString,
					},
					"values_repo_artifact_meta": &schema.Schema{
						Description: "repo information override values",
						Elem: &schema.Resource{Schema: map[string]*schema.Schema{
							"git_options": &schema.Schema{
								Description: "git options",
								Optional:    true,
								Type:        schema.TypeList,
								Elem: &schema.Resource{Schema: map[string]*schema.Schema{
									"revision": &schema.Schema{
										Description: "repository revision",
										Optional:    true,
										Type:        schema.TypeString,
									},
									"repo_artifact_files": &schema.Schema{
										Description: "repository revision",
										Optional:    true,
										Type:        schema.TypeList,
										Elem: &schema.Resource{Schema: map[string]*schema.Schema{
											"name": &schema.Schema{
												Description: "name",
												Optional:    true,
												Type:        schema.TypeString,
											},
											"relative_path": &schema.Schema{
												Description: "file path",
												Optional:    true,
												Type:        schema.TypeString,
											},
											"file_type": &schema.Schema{
												Description: "file type",
												Optional:    true,
												Type:        schema.TypeString,
											},
										}},
									},
								},
								}},
							"helm_options": &schema.Schema{
								Description: "helm options",
								Optional:    true,
								Type:        schema.TypeList,
								Elem: &schema.Resource{Schema: map[string]*schema.Schema{
									"chart_name": &schema.Schema{
										Description: "chart name",
										Optional:    true,
										Type:        schema.TypeString,
									},
									"tag": &schema.Schema{
										Description: "tag for chart",
										Optional:    true,
										Type:        schema.TypeString,
									},
								}},
							},
							"timeouts": &schema.Schema{
								Description: "timeouts",
								Optional:    true,
								Default:     0,
								Type:        schema.TypeInt,
							},
						}},
						MaxItems: 1,
						MinItems: 1,
						Optional: true,
						Type:     schema.TypeList,
					},
					"cluster_placement": &schema.Schema{
						Description: "placement specification of the override resource",
						Elem: &schema.Resource{Schema: map[string]*schema.Schema{
							"cluster_labels": &schema.Schema{
								Description: "list of labels for placement",
								Elem: &schema.Resource{Schema: map[string]*schema.Schema{
									"key": &schema.Schema{
										Description: "Key of the placement label",
										Optional:    true,
										Type:        schema.TypeString,
									},
									"value": &schema.Schema{
										Description: "Value of the placement label",
										Optional:    true,
										Type:        schema.TypeString,
									},
								}},
								MaxItems: 0,
								MinItems: 0,
								Optional: true,
								Type:     schema.TypeList,
							},
							"cluster_selector": &schema.Schema{
								Description: "Kubernetes style label selector",
								Optional:    true,
								Type:        schema.TypeString,
							},
							"placement_type": &schema.Schema{
								Description: "placement type, value ClusterSelector",
								Optional:    true,
								Type:        schema.TypeString,
							},
							"node_grouping_keys": &schema.Schema{
								Description: "node grouping keys",
								Elem:        &schema.Schema{Type: schema.TypeString},
								Optional:    true,
								Type:        schema.TypeList,
							},
						}},
						MaxItems: 1,
						MinItems: 1,
						Optional: true,
						Type:     schema.TypeList,
					},
				}},
				MaxItems: 1,
				MinItems: 1,
				Optional: true,
				Type:     schema.TypeList,
			},
		},
	}
}

func resourceClusterOverrideCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	diags := resourceOverrideUpsert(ctx, d, m)
	if diags.HasError() {
		tflog := os.Getenv("TF_LOG")
		if tflog == "TRACE" || tflog == "DEBUG" {
			ctx = context.WithValue(ctx, "debug", "true")
		}
		log.Printf("override create got error, perform cleanup")
		or, err := expandOverride(d)
		if err != nil {
			log.Printf("override expandNamespace error")
			return diags
		}

		projectId, err := config.GetProjectIdByName(or.Metadata.Project)
		if err != nil {
			return diags
		}

		err = clusteroverride.DeleteClusterOverride(or.Metadata.Name, projectId, or.Spec.Type)
		if err != nil {
			log.Println("failed to delete cluster override ", or.Metadata.Name)
			return diags
		}
	}
	return diags
}

func resourceClusterOverrideUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("cluster override update starts")
	return resourceOverrideUpsert(ctx, d, m)
}

func resourceOverrideUpsert(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("override create starts")
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	if d.State() != nil && d.State().ID != "" {
		n := GetMetaName(d)
		if n != "" && n != d.State().ID {
			log.Printf("metadata name change not supported")
			d.State().Tainted = true
			return diag.FromErr(fmt.Errorf("%s", "metadata name change not supported"))
		}
	}

	or, err := expandOverride(d)
	if err != nil {
		return diag.FromErr(err)
	}

	projectId, err := config.GetProjectIdByName(or.Metadata.Project)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Println("resourceOverrideUpsert ", or.Spec)
	w1 := spew.Sprintf("%+v", or.Spec)
	log.Println("name ", or.Metadata.Name, " project ", projectId)
	log.Println("resourceOverrideUpsert spec: ", w1)
	err = clusteroverride.UpdateClusterOverride(or.Metadata.Name, projectId, or.Spec, true)
	if err != nil {
		log.Println("failed to create/update cluster override ", or.Metadata.Name, " error ", err)
		return diag.FromErr(err)
	}

	d.SetId(or.Metadata.Name)
	return diags
}

func resourceClusterOverrideDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}
	log.Printf("override create got error, perform cleanup")
	or, err := expandOverride(d)
	if err != nil {
		log.Printf("override expandNamespace error")
		return diag.FromErr(err)
	}

	projectId, err := config.GetProjectIdByName(or.Metadata.Project)
	if err != nil {
		return diag.FromErr(err)
	}

	err = clusteroverride.DeleteClusterOverride(or.Metadata.Name, projectId, or.Spec.Type)
	if err != nil {
		log.Println("failed to delete cluster override ", or.Metadata.Name)
		return diag.FromErr(err)
	}

	return diags
}

func resourceClusterOverrideRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	log.Println("resourceBlueprintRead ")
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}
	tfLocalState, err := expandOverride(d)
	if err != nil {
		return diag.FromErr(err)
	}

	// XXX Debug
	// w1 := spew.Sprintf("%+v", tfLocalState)
	// log.Println("resourceBluePrintRead tfLocalState", w1)

	projectId, err := config.GetProjectIdByName(tfLocalState.Metadata.Project)
	if err != nil {
		return diag.FromErr(err)
	}
	remoteOr, err := clusteroverride.GetClusterOverride(tfLocalState.Metadata.Name, projectId, tfLocalState.Spec.Type)
	if err != nil {
		log.Println("get cluster override failed: ", err)
		return diag.FromErr(err)
	}

	// XXX Debug
	w1 := spew.Sprintf("%+v", remoteOr)
	log.Println("resourceClusterOverrideRead remoteOr", w1)

	err = flattenClusterOverride(d, remoteOr, tfLocalState.Metadata.Project)
	if err != nil {
		return diag.FromErr(err)
	}
	return diags

}

func expandOverride(in *schema.ResourceData) (*clusterOverrideYamlConfig, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "expandWorkload empty input")
	}
	obj := &clusterOverrideYamlConfig{}

	if v, ok := in.Get("metadata").([]interface{}); ok {
		obj.Metadata = expandMetaData(v)
	}

	if v, ok := in.Get("spec").([]interface{}); ok {
		objSpec, err := expandOverrideSpec(v)
		if err != nil {
			return nil, err
		}
		obj.Spec = objSpec
	}

	obj.Kind = "ClusterOverride"
	obj.ApiVersion = "config.rafay.dev/v2"
	return obj, nil
}

func expandOverrideSpec(p []interface{}) (models.ClusterOverrideSpec, error) {
	obj := models.ClusterOverrideSpec{}
	if len(p) == 0 || p[0] == nil {
		return obj, fmt.Errorf("%s", "expandOverrideSpec empty input")
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["cluster_selector"].(string); ok && len(v) > 0 {
		obj.ClusterSelector = v
	}

	if v, ok := in["resource_selector"].(string); ok && len(v) > 0 {
		obj.ResourceSelector = v
	}

	if v, ok := in["type"].(string); ok && len(v) > 0 {
		obj.Type = v
	}

	if v, ok := in["override_values"].(string); ok && len(v) > 0 {
		obj.OverrideValues = v
	}

	if v, ok := in["cluster_placement"].([]interface{}); ok && len(v) > 0 {
		obj.ClusterPlacement = expandClusterPlacement(v)
	}

	if v, ok := in["value_repo_ref"].(string); ok && len(v) > 0 {
		obj.RepositoryRef = v
	}

	if v, ok := in["values_repo_artifact_meta"].([]interface{}); ok {
		log.Println("values_repo_artifact_meta")
		obj.RepoArtifactMeta = expandValueRepoArtifactMeta(v)
	}

	return obj, nil
}

func expandClusterPlacement(p []interface{}) models.PlacementSpec {
	obj := models.PlacementSpec{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}

	log.Println("expandClusterPlacement ")
	in := p[0].(map[string]interface{})
	if v, ok := in["placement_type"].(string); ok && len(v) > 0 {
		obj.PlacementType = models.PlacementType(v)
	}

	if v, ok := in["cluster_selector"].(string); ok && len(v) > 0 {
		obj.ClusterSelector = v
	}

	if v, ok := in["cluster_labels"].([]interface{}); ok && len(v) > 0 {
		obj.ClusterLabels = expandClusterLabels(v)
	}
	// XXX Debug
	w1 := spew.Sprintf("%+v", obj)
	log.Println("expandClusterPlacement obj", w1)
	return obj
}

func expandValueRepoArtifactMeta(p []interface{}) models.RepoArtifactMeta {
	obj := models.RepoArtifactMeta{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	log.Println("expandValueRepoArtifactMeta")
	in := p[0].(map[string]interface{})

	if v, ok := in["git_options"].([]interface{}); ok {
		obj.Git = expandGit(v)
	}

	if v, ok := in["helm_options"].([]interface{}); ok {
		obj.Helm = expandHelm(v)
	}

	if v, ok := in["timeouts"].(int64); ok {
		obj.Timeout = v
	}

	return obj
}

func expandGit(p []interface{}) *models.GitOptions {
	obj := &models.GitOptions{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["revision"].(string); ok {
		obj.Revision = v
	}

	if v, ok := in["repo_artifact_files"].([]interface{}); ok {
		obj.RepoArtifactFiles = expandRepoArtifactFiles(v)
	}

	return obj
}

func expandHelm(p []interface{}) *models.HelmOptions {
	obj := &models.HelmOptions{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["chart_name"].(string); ok {
		obj.ChartName = v
	}

	if v, ok := in["tag"].(string); ok {
		obj.Tag = v
	}

	return obj
}

func expandRepoArtifactFiles(p []interface{}) []models.RepoFile {
	if len(p) == 0 || p[0] == nil {
		return []models.RepoFile{}
	}

	out := make([]models.RepoFile, len(p))

	for i := range p {
		obj := models.RepoFile{}
		in := p[i].(map[string]interface{})

		if v, ok := in["name"].(string); ok {
			obj.Name = v
		}

		if v, ok := in["relative_path"].(string); ok {
			obj.RelPath = v
		}

		if v, ok := in["file_type"].(models.FileType); ok {
			obj.FileType = v
		}

		out[i] = obj

	}

	return out
}

func expandClusterLabels(p []interface{}) []*models.PlacementLabel {
	if len(p) == 0 || p[0] == nil {
		return []*models.PlacementLabel{}
	}

	out := make([]*models.PlacementLabel, len(p))

	for i := range p {
		obj := models.PlacementLabel{}
		in := p[i].(map[string]interface{})

		if v, ok := in["key"].(string); ok {
			obj.Key = v
		}

		if v, ok := in["value"].(string); ok {
			obj.Value = v
		}
		w1 := spew.Sprintf("%+v", obj)
		log.Println("expandClusterLabels obj ", w1)

		out[i] = &obj
	}

	w1 := spew.Sprintf("%+v", out)
	log.Println("expandClusterLabels out ", w1)
	return out
}

// Flatteners

func flattenClusterOverride(d *schema.ResourceData, in *models.ClusterOverride, projectName string) error {
	if in == nil {
		return nil
	}

	if len(in.RafayMeta.Name) > 0 {
		m := commonpb.Metadata{}
		m.Name = in.RafayMeta.Name
		m.Labels = in.RafayMeta.Labels
		m.Annotations = in.Annotations
		m.Project = projectName
		log.Println("flattenClusterOverride ", m)
		err := d.Set("metadata", flattenMetaData(&m))
		if err != nil {
			return err
		}
	}

	v, ok := d.Get("spec").([]interface{})
	if !ok {
		v = []interface{}{}
	}

	// XXX Debug
	// w1 := spew.Sprintf("%+v", v)
	// log.Println("flattenBlueprint before ", w1)
	var ret []interface{}
	ret, err := flattenOverrideSpec(in.ClusterOverrideSpec, v)
	if err != nil {
		return err
	}
	// XXX Debug
	// w1 = spew.Sprintf("%+v", ret)
	// log.Println("flattenBlueprint after ", w1)

	err = d.Set("spec", ret)
	if err != nil {
		return err
	}
	return nil
}

func flattenOverrideSpec(in models.ClusterOverrideSpec, p []interface{}) ([]interface{}, error) {

	log.Println("flattenOverrideSpec ")
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.ClusterSelector) > 0 {
		obj["cluster_selector"] = in.ClusterSelector
	}

	if len(in.ResourceSelector) > 0 {
		obj["resource_selector"] = in.ResourceSelector
	}

	if len(in.Type) > 0 {
		obj["type"] = in.Type
	}

	if len(in.RepositoryRef) > 0 {
		obj["value_repo_ref"] = in.RepositoryRef
	}

	v, ok := obj["values_repo_artifact_meta"].([]interface{})
	if !ok {
		v = []interface{}{}
	}
	obj["values_repo_artifact_meta"] = flattenArtifactMeta(in.RepoArtifactMeta, v)

	if in.RepoArtifactMeta.Git == nil {
		log.Println("flattenOverrideSpec ")

		if len(in.OverrideValues) > 0 {
			obj["override_values"] = in.OverrideValues
		}
	}
	return []interface{}{obj}, nil
}

func flattenArtifactMeta(in models.RepoArtifactMeta, p []interface{}) []interface{} {

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	retNIL := true
	if in.Timeout != 0 {
		obj["timeouts"] = in.Timeout
		retNIL = false
	}

	if in.Git != nil {
		v, ok := obj["git_options"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["git_options"] = flattenGitOptions(in.Git, v)
		retNIL = false
	}

	if in.Helm != nil {
		v, ok := obj["helm_options"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["helm_options"] = flattenHelmOptions(in.Helm, v)
		retNIL = false
	}

	if retNIL {
		return nil
	}

	return []interface{}{obj}
}

func flattenGitOptions(in *models.GitOptions, p []interface{}) []interface{} {

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.Revision) > 0 {
		obj["revision"] = in.Revision
	}

	if len(in.RepoArtifactFiles) > 0 {
		v, ok := obj["repo_artifact_files"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["repo_artifact_files"] = flattenRepoArtifactFiles(in.RepoArtifactFiles, v)
	}

	return []interface{}{obj}
}

func flattenRepoArtifactFiles(input []models.RepoFile, p []interface{}) []interface{} {
	if input == nil {
		return nil
	}
	out := make([]interface{}, len(input))
	for i, in := range input {

		obj := map[string]interface{}{}
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}

		if len(in.Name) > 0 {
			obj["name"] = in.Name
		}
		if len(in.RelPath) > 0 {
			obj["relative_path"] = in.RelPath
		}
		if len(in.FileType) > 0 {
			obj["file_type"] = in.FileType
		}

		out[i] = &obj

	}

	return out
}

func flattenHelmOptions(in *models.HelmOptions, p []interface{}) []interface{} {

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.ChartName) > 0 {
		obj["chart_name"] = in.ChartName
	}
	if len(in.Tag) > 0 {
		obj["tag"] = in.Tag
	}

	return []interface{}{obj}
}

func resourceClusterOverrideCreate1(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	filePath := d.Get("cluster_override_filepath").(string)
	var co commands.ClusterOverrideYamlConfig
	//make sure this is the correct file path
	log.Println("filepath: ", filePath)
	log.Printf("create cluster override resource")
	//get project id with project name, p.id used to refer to project id -> need p.ID for calling createClusterOverride and getClusterOverride
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
	//open file path and retirve config spec from yaml file (from run function in commands/create_cluster_override.go)
	//read and capture file from file path
	if f, err := os.Open(filePath); err == nil {
		c, err := ioutil.ReadAll(f)
		if err != nil {
			log.Println("error cpaturing file")
		}
		//implement createClusterOverride from commands/create_cluster_override.go -> then call clusteroverride.CreateClusterOverride
		clusterOverrideDefinition := c
		err = yaml.Unmarshal(clusterOverrideDefinition, &co)
		if err != nil {
			log.Printf("Failed Unmarshal correctly")
		}
		//get cluster override spec from yaml file
		spec, err := getClusterOverrideSpecFromYamlConfigSpec(co, filePath)
		if err != nil {
			log.Printf("Failed to get ClusterOverrideSpecFromYamlConfigSpec")
		}
		//create cluster override
		err = clusteroverride.CreateClusterOverride(co.Metadata.Name, projectId, *spec)
		if err != nil {
			log.Printf("Failed to create cluster override: %s\n", co.Metadata.Name)
		} else {
			log.Printf("Successfully created cluster override: %s\n", co.Metadata.Name)
		}
	} else {
		log.Println("Couldn't open file, err: ", err)
	}
	//get cluster override to ensure cluster override was created properly
	getClus_resp, err := clusteroverride.GetClusterOverride(co.Metadata.Name, projectId, co.Spec.Type)
	if err != nil {
		log.Println("get cluster override failed: ", getClus_resp, err)
	} else {
		log.Println("got newly created cluster override: ", co.Metadata.Name)
	}
	//set id to metadata.Name
	d.SetId(co.Metadata.Name)
	return diags
}

func resourceClusterOverrideRead1(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	filePath := d.Get("cluster_override_filepath").(string)
	var co commands.ClusterOverrideYamlConfig
	log.Printf("create cluster override resource")
	//get project id with project name, p.id used to refer to project id -> need p.ID for calling createClusterOverride and getClusterOverride
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
	//open file path and retirve config spec from yaml file (from run function in commands/create_cluster_override.go)
	//read and capture file from file path
	if f, err := os.Open(filePath); err == nil {
		// capture the entire file
		c, err := ioutil.ReadAll(f)
		if err != nil {
			log.Println("error cpaturing file")
		}
		//unmarshal yaml file to get correct specs
		clusterOverrideDefinition := c
		err = yaml.Unmarshal(clusterOverrideDefinition, &co)
		if err != nil {
			log.Printf("Failed Unmarshal correctly")
		}
		//get cluster override spec from yaml file
		_, err = getClusterOverrideSpecFromYamlConfigSpec(co, filePath)
		if err != nil {
			log.Printf("Failed to get ClusterOverrideSpecFromYamlConfigSpec")
		}
	} else {
		log.Println("Couldn't open file, err: ", err)
	}
	//get cluster override to ensure cluster override was created properly
	getClus_resp, err := clusteroverride.GetClusterOverride(co.Metadata.Name, projectId, co.Spec.Type)
	if err != nil {
		log.Println("get cluster override failed: ", getClus_resp, err)
	} else {
		log.Println("got newly created cluster override: ", co.Metadata.Name)
	}
	return diags
}

func resourceClusterOverrideUpdate1(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("update cluster override resource")
	filePath := d.Get("cluster_override_filepath").(string)
	createIfNotPresent := false //this is how it is set in commands/update_cluster_override
	//retrieve project_id from project name
	resp, err := project.GetProjectByName(d.Get("projectname").(string))
	if err != nil {
		log.Println("error cpaturing file")
	}
	p, err := project.NewProjectFromResponse([]byte(resp))
	if err != nil {
		return diag.FromErr(err)
	} else if p == nil {
		d.SetId("")
		return diags
	}
	project_id := p.ID
	//update cluster implemented from commmands/update_cluster_override -> will call UpdateClusterOverride from cluster_override.go
	// open and read file then unmarshal the data
	if f, err := os.Open(filePath); err == nil {
		// capture the entire file
		c, err := ioutil.ReadAll(f)
		if err != nil {
			log.Println("error reading file")
			return diags
		}
		clusterOverrideDefinition := c
		var co commands.ClusterOverrideYamlConfig
		err = yaml.Unmarshal(clusterOverrideDefinition, &co)
		if err != nil {
			log.Println("error unmarshalling Cluster Override")
			return diags
		}
		// get cluster override spec from yaml file
		spec, err := getClusterOverrideSpecFromYamlConfigSpec(co, filePath)
		if err != nil {
			log.Println("error getting Cluster Override Spec From Yaml Config Spec")
			return diags
		}
		//update cluster
		err = clusteroverride.UpdateClusterOverride(co.Metadata.Name, project_id, *spec, createIfNotPresent)
		if err != nil {
			log.Printf("Failed to update cluster override: %s\n", co.Metadata.Name)
			return diags
		} else {
			log.Printf("Successfully created/updated cluster override: %s\n", co.Metadata.Name)
		}
		return diags
	} else {
		log.Println("error opening file")
		return diags
	}
}

func resourceClusterOverrideDelete1(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("resource cluster override delete id %s", d.Id())
	var co commands.ClusterOverrideYamlConfig
	filePath := d.Get("cluster_override_filepath").(string)
	//get project id with project name, p.id used to refer to project id -> need p.ID for calling deleteClusterOverride
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
	project_id := p.ID
	//open, read, and unmarshal file to retrieve ClusterOverrideYamlConfig Struct to pass in type for delete cluster
	if f, err := os.Open(filePath); err == nil {
		c, err := ioutil.ReadAll(f)
		if err != nil {
			log.Println("error cpaturing file")
		}
		clusterOverrideDefinition := c
		err = yaml.Unmarshal(clusterOverrideDefinition, &co)
		if err != nil {
			log.Printf("Failed Unmarshal correctly")
		}
	}
	//delete cluster override
	err = clusteroverride.DeleteClusterOverride(co.Metadata.Name, project_id, co.Spec.Type)
	if err != nil {
		fmt.Print("cluster was not deleted")
		log.Printf("cluster was not deleted, error %s", err.Error())
		return diag.FromErr(err)
	}
	return diags
}

func getClusterOverrideSpecFromYamlConfigSpec(co commands.ClusterOverrideYamlConfig, filePath string) (*models.ClusterOverrideSpec, error) {
	var spec models.ClusterOverrideSpec
	spec.ClusterSelector = co.Spec.ClusterSelector
	spec.ResourceSelector = co.Spec.ResourceSelector
	spec.Type = co.Spec.Type
	spec.RepositoryRef = co.Spec.ValueRepoRef
	spec.RepoArtifactMeta = co.Spec.ValueRepoArtifactMeta
	if co.Spec.OverrideValues != "" && co.Spec.ValuesFile != "" {
		return nil, fmt.Errorf("invalid config: both overrideValues and overrideValuesFile were provided")
	}
	if co.Spec.ValuesFile != "" {
		values, err := getOverrideValues(co, filePath)
		if err != nil {
			return nil, fmt.Errorf("invalid config: error fetching the content of the value file from the location provided %s: Error: %s", co.Spec.ValuesFile, err.Error())
		}
		spec.OverrideValues = values
	}
	if co.Spec.OverrideValues != "" {
		spec.OverrideValues = co.Spec.OverrideValues
	}
	if spec.OverrideValues == "" && spec.RepositoryRef == "" {
		return nil, fmt.Errorf("invalid config: neither overrideValues not valueRepoRef were provided")
	}
	if spec.OverrideValues != "" && spec.RepositoryRef != "" {
		return nil, fmt.Errorf("invalid config: both overrideValues and valueRepoRef were provided")
	}
	if spec.RepositoryRef != "" && (spec.RepoArtifactMeta.Git == nil || len(spec.RepoArtifactMeta.Git.RepoArtifactFiles) == 0) {
		return nil, fmt.Errorf("invalid config: exactly one repo artifact file should be provided.\"")
	}
	return &spec, nil
}

func getOverrideValues(co commands.ClusterOverrideYamlConfig, filePath string) (string, error) {
	valueFileLocation := utils.FullPath(filePath, co.Spec.ValuesFile)
	if _, err := os.Stat(valueFileLocation); os.IsNotExist(err) {
		return "", fmt.Errorf("values file doesn't exist '%s'", valueFileLocation)
	}
	valueFileContent, err := ioutil.ReadFile(valueFileLocation)
	if err != nil {
		return "", fmt.Errorf("error in reading the value file %s: %s\n", valueFileLocation, err)
	}
	return string(valueFileContent), nil
}
