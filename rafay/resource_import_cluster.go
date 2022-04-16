package rafay

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/RafaySystems/rctl/pkg/cluster"
	"github.com/RafaySystems/rctl/pkg/project"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type ImportCluster struct {
	// +required
	Kind     string                 `yaml:"kind,omitempty"`
	Metadata *ImportClusterMetadata `yaml: "metdatdata,omitemty"`
	Spec     *ImportClusterSpec     `yaml:"spec,omitempty"`
}
type ImportClusterMetadata struct {
	Name    string            `yaml:"name,omitempty"`
	Project string            `yaml:"project,omitempty"`
	Labels  map[string]string `yaml:"labels,omitempty"`
}
type ImportClusterSpec struct {
	// +required
	Type             string `yaml:"type,omitempty"`
	Blueprint        string `yaml:"blueprint,omitempty"`
	BlueprintVersion string `yaml:"blueprintVersion,omitempty"`
	Location         string `yaml:"location,omitempty"`
	KubeConfigPath   string `yaml:"kubeConfigPath,omitempty"`
	Description      string `yaml:"description,omitempty"`
}

func resourceImportCluster() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceImportClusterCreate,
		ReadContext:   resourceImportClusterRead,
		UpdateContext: resourceImportClusterUpdate,
		DeleteContext: resourceImportClusterDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			/*hardcode kind
			"kind": {
				Type:     schema.TypeString,
				Required: true,
			},*/
			"metadata": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "cluster yaml file",
				Elem: &schema.Resource{
					Schema: metadataFields(),
				},
			},
			"spec": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "cluster yaml file",
				Elem: &schema.Resource{
					Schema: specFields(),
				},
			},
		},
	}
}

func metadataFields() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "EKS Cluster name",
		},
		"project": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Project for the cluster",
		},
		"labels": {
			Type:        schema.TypeMap,
			Optional:    true,
			Description: "Cluster Labels",
		},
	}
	return s
}

func specFields() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"type": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "can be 'imported', 'aws-ec2', 'gcp', 'aws-eks', 'on-prem'",
		},
		"blueprint": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "field is optional, if not specified, default value is 'default'",
		},
		"blueprint_version": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"location": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "location, can be custom or predefined",
		},
		"kubeconfig_path": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"description": {
			Type:     schema.TypeString,
			Optional: true,
		},
	}
	return s
}

func resourceImportClusterCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Println("create import cluster resource")
	return resourceImportClusterUpsert(ctx, d, m)
}

func resourceImportClusterUpsert(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("import cluster starts")
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	importCluster, err := expandImportCluster(d)
	if err != nil {
		log.Printf("expand importCluster error")
		return diag.FromErr(err)
	}
	log.Printf("expand importCluster:", importCluster)

	//get project id with project name, p.id used to refer to project id -> need p.ID for calling getCluster and GetBootstrapFile and NewImportCluster
	resp, err := project.GetProjectByName(importCluster.Metadata.Project)
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

	//create imported cluster
	resp, err = cluster.NewImportCluster(importCluster.Metadata.Name, importCluster.Spec.Blueprint, importCluster.Spec.Location, importCluster.Metadata.Project, importCluster.Spec.BlueprintVersion)
	if err != nil {
		log.Printf("create import cluster failed to create (check parameters passed in), error %s", err.Error())
		return diag.FromErr(err)
	}

	time.Sleep(10 * time.Second)
	//if error with get cluster add a sleep to wait for cluster creation
	//make sure new imported cluster was created by calling get cluster and checking for no errors
	cluster_resp, err := cluster.GetCluster(importCluster.Metadata.Name, project_id)
	if err != nil {
		log.Printf("imported cluster was not created, error %s", err.Error())
		return diag.FromErr(err)
	}

	if importCluster.Spec.BlueprintVersion != "" {
		cluster_resp.ClusterBlueprintVersion = importCluster.Spec.BlueprintVersion
		err = cluster.UpdateCluster(cluster_resp)
		if err != nil {
			log.Printf("setting cluster blueprint version failed, error %s", err.Error())
			return diag.FromErr(err)
		}
	}

	//then retrieve bootstrap yaml file, call GetBootstrapFile() -> make sure this function downloads the bootstrap file locally (i think the url request does)
	bootsrap_file, err := cluster.GetBootstrapFile(importCluster.Metadata.Name, project_id)
	if err != nil {
		log.Printf("bootstrap yaml file was not obtained correctly, error %s", err.Error())
		return diag.FromErr(err)
	}
	log.Println("bootstrap_filepath got correctly: \n", bootsrap_file)
	//write bootstrap file into bootstrap file path
	f, err := os.Create("bootstrap.yaml")

	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	_, err2 := f.WriteString(bootsrap_file)

	if err2 != nil {
		log.Printf("bootstrap yaml file was not written correctly, error %s", err2.Error())
		return diag.FromErr(err2)
	}
	//pass in bootstrap file path into exec command
	bootstrap_filepath, _ := filepath.Abs("bootstrap.yaml")
	//figure out how to apply bootstrap yaml file to created cluster STILL NEED TO COMPLETE
	//add kube_config file as optional schema, call os/exec to cal kubectl apply on the filepath to kube config
	if (d.Get("kubeconfig_path").(string)) != "" {
		cmd := exec.Command("kubectl", "--kubeconfig", importCluster.Spec.KubeConfigPath, "apply", "-f", bootstrap_filepath)
		var out bytes.Buffer

		//cmd.Stdout = &out
		log.Println("load client", "id", project_id, "command", cmd)
		b, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Print("failed to apply bootstrap yaml to cluster")
			log.Println("kubectl command failed to apply bootstrap yaml file", string(b))
			log.Println("command", "id", project_id, "error", err, "out", out.String())
		}
	}

	//set ID for imported cluster id, d.SetID()
	d.SetId(cluster_resp.ID)
	return diags

}

func expandImportCluster(in *schema.ResourceData) (*ImportCluster, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "expandWorkload empty input")
	}
	obj := &ImportCluster{}

	if v, ok := in.Get("metadata").([]interface{}); ok {
		obj.Metadata = expandMetadata(v)
	}

	if v, ok := in.Get("spec").([]interface{}); ok {
		objSpec, err := expandSpec(v)
		if err != nil {
			return nil, err
		}
		obj.Spec = objSpec
	}

	obj.Kind = "ClusterOverride"
	//obj.ApiVersion = "config.rafay.dev/v2"
	return obj, nil
}

func expandMetadata(p []interface{}) *ImportClusterMetadata {
	obj := &ImportClusterMetadata{}
	if p == nil || len(p) == 0 || p[0] == nil {
		return obj
	}

	in := p[0].(map[string]interface{})
	if v, ok := in["name"].(string); ok && len(v) > 0 {
		obj.Name = v
	}
	if v, ok := in["project"].(string); ok && len(v) > 0 {
		obj.Project = v
	}
	if v, ok := in["labels"].(map[string]interface{}); ok && len(v) > 0 {
		obj.Labels = toMapString(v)
	}

	log.Println("expandMetaData")
	return obj
}

func expandSpec(p []interface{}) (*ImportClusterSpec, error) {
	obj := &ImportClusterSpec{}
	if len(p) == 0 || p[0] == nil {
		return obj, fmt.Errorf("%s", "expandOverrideSpec empty input")
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["type"].(string); ok && len(v) > 0 {
		obj.Type = v
	}
	if v, ok := in["blueprint"].(string); ok && len(v) > 0 {
		obj.Blueprint = v
	}
	if v, ok := in["blueprint_version"].(string); ok && len(v) > 0 {
		obj.BlueprintVersion = v
	}
	if v, ok := in["location"].(string); ok && len(v) > 0 {
		obj.Location = v
	}
	if v, ok := in["kubeconfig_path"].(string); ok && len(v) > 0 {
		obj.Blueprint = v
	}
	if v, ok := in["description"].(string); ok && len(v) > 0 {
		obj.Description = v
	}

	return obj, nil
}

func flattenImportCluster(in *ImportCluster, p []interface{}) ([]interface{}, error) {
	obj := map[string]interface{}{}
	if in == nil {
		return nil, fmt.Errorf("empty cluster input")
	}

	if len(in.Kind) > 0 {
		obj["kind"] = in.Kind
	}
	var err error
	//flatten eks cluster metadata
	var ret1 []interface{}
	if in.Metadata != nil {
		v, ok := obj["metadata"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		ret1, err = flattenMetadata(in.Metadata, v)
		log.Println("ret1: ", ret1)
		if err != nil {
			log.Println("flattenMetadata err")
			return nil, err
		}
		obj["metadata"] = ret1
		log.Println("set metadata: ", obj["metadata"])
	}
	//flattening EKSClusterSpec
	var ret2 []interface{}
	if in.Spec != nil {
		v, ok := obj["spec"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		ret2, err = flattenSpec(in.Spec, v)
		if err != nil {
			log.Println("flattenEKSClusterSpec err")
			return nil, err
		}
		obj["spec"] = ret2
		log.Println("set metadata: ", obj["spec"])
	}
	log.Println("flattenImportCluster finished ")
	return []interface{}{obj}, nil
}

func flattenMetadata(in *ImportClusterMetadata, p []interface{}) ([]interface{}, error) {
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}
	if in == nil {
		return []interface{}{obj}, nil
	}

	if len(in.Name) > 0 {
		obj["name"] = in.Name
	}
	log.Println("md 1")
	if len(in.Project) > 0 {
		obj["project"] = in.Project
	}
	log.Println("md 2")
	if in.Labels != nil && len(in.Labels) > 0 {
		obj["labels"] = toMapInterface(in.Labels)
		log.Println("saving metadata labels: ", in.Labels)
	}
	log.Println("md 3")
	return []interface{}{obj}, nil
}

func flattenSpec(in *ImportClusterSpec, p []interface{}) ([]interface{}, error) {
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}
	if in == nil {
		return []interface{}{obj}, nil
	}

	if len(in.Blueprint) > 0 {
		obj["blueprint"] = in.Blueprint
	}
	log.Println("md 1")
	if len(in.BlueprintVersion) > 0 {
		obj["project"] = in.BlueprintVersion
	}
	log.Println("md 2")
	if len(in.Type) > 0 {
		obj["type"] = in.Type
	}
	log.Println("md 3")
	if len(in.Location) > 0 {
		obj["location"] = in.Location
	}
	if len(in.KubeConfigPath) > 0 {
		obj["kubeconfig_path"] = in.KubeConfigPath
	}
	if len(in.Description) > 0 {
		obj["description"] = in.Description
	}
	return []interface{}{obj}, nil
}

func resourceImportClusterCreate1(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	log.Printf("create import cluster resource")
	//get project id with project name, p.id used to refer to project id -> need p.ID for calling getCluster and GetBootstrapFile and NewImportCluster
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

	//create imported cluster
	resp, err = cluster.NewImportCluster(d.Get("clustername").(string), d.Get("blueprint").(string), d.Get("location").(string), project_id, d.Get("blueprint_version").(string))
	if err != nil {
		log.Printf("create import cluster failed to create (check parameters passed in), error %s", err.Error())
		return diag.FromErr(err)
	}

	time.Sleep(10 * time.Second)
	//if error with get cluster add a sleep to wait for cluster creation
	//make sure new imported cluster was created by calling get cluster and checking for no errors
	cluster_resp, err := cluster.GetCluster(d.Get("clustername").(string), project_id)
	if err != nil {
		log.Printf("imported cluster was not created, error %s", err.Error())
		return diag.FromErr(err)
	}

	if d.Get("blueprint_version").(string) != "" {
		cluster_resp.ClusterBlueprintVersion = d.Get("blueprint_version").(string)
		err = cluster.UpdateCluster(cluster_resp)
		if err != nil {
			log.Printf("setting cluster blueprint version failed, error %s", err.Error())
			return diag.FromErr(err)
		}
	}

	//then retrieve bootstrap yaml file, call GetBootstrapFile() -> make sure this function downloads the bootstrap file locally (i think the url request does)
	bootsrap_file, err := cluster.GetBootstrapFile(d.Get("clustername").(string), project_id)
	if err != nil {
		log.Printf("bootstrap yaml file was not obtained correctly, error %s", err.Error())
		return diag.FromErr(err)
	}
	log.Println("bootstrap_filepath got correctly: \n", bootsrap_file)
	//write bootstrap file into bootstrap file path
	f, err := os.Create("bootstrap.yaml")

	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	_, err2 := f.WriteString(bootsrap_file)

	if err2 != nil {
		log.Printf("bootstrap yaml file was not written correctly, error %s", err2.Error())
		return diag.FromErr(err2)
	}
	//pass in bootstrap file path into exec command
	bootstrap_filepath, _ := filepath.Abs("bootstrap.yaml")
	//figure out how to apply bootstrap yaml file to created cluster STILL NEED TO COMPLETE
	//add kube_config file as optional schema, call os/exec to cal kubectl apply on the filepath to kube config
	if (d.Get("kubeconfig_path").(string)) != "" {
		cmd := exec.Command("kubectl", "--kubeconfig", d.Get("kubeconfig_path").(string), "apply", "-f", bootstrap_filepath)
		var out bytes.Buffer

		//cmd.Stdout = &out
		log.Println("load client", "id", project_id, "command", cmd)
		b, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Print("failed to apply bootstrap yaml to cluster")
			log.Println("kubectl command failed to apply bootstrap yaml file", string(b))
			log.Println("command", "id", project_id, "error", err, "out", out.String())
		}
	}

	//set ID for imported cluster id, d.SetID()
	d.SetId(cluster_resp.ID)
	return diags

}

func resourceImportClusterRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	log.Println("resourceImportClusterRead ")
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}
	importCluster, err := expandImportCluster(d)
	if err != nil {
		return diag.FromErr(err)
	}

	resp, err := project.GetProjectByName(importCluster.Metadata.Project)
	if err != nil {
		fmt.Print("project name missing in the resource")
		return diags
	}

	project, err := project.NewProjectFromResponse([]byte(resp))
	if err != nil {
		fmt.Printf("project does not exist")
		return diags
	}
	c, err := cluster.GetCluster(importCluster.Metadata.Name, project.ID)
	if err != nil {
		log.Printf("error in get cluster %s", err.Error())
		return diag.FromErr(err)
	}
	if err := d.Set("clustername", c.Name); err != nil {
		log.Printf("get group set name error %s", err.Error())
		return diag.FromErr(err)
	}

	return diags
}

func resourceImportClusterUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("update imported cluster resource")

	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	importCluster, err := expandImportCluster(d)
	if err != nil {
		return diag.FromErr(err)
	}

	//retrieve project_id from project name for calling get_cluster
	resp, err := project.GetProjectByName(importCluster.Metadata.Project)
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
	//retrieve cluster_details from get cluster to pass into update cluster
	cluster_resp, err := cluster.GetCluster(importCluster.Metadata.Name, project_id)
	if err != nil {
		log.Printf("imported cluster was not created, error %s", err.Error())
		return diag.FromErr(err)
	}
	// read the blueprint name
	if importCluster.Spec.Blueprint != "" {
		cluster_resp.ClusterBlueprint = d.Get("blueprint").(string)
	}
	// read the blueprint version
	if importCluster.Spec.BlueprintVersion != "" {
		cluster_resp.ClusterBlueprintVersion = importCluster.Spec.BlueprintVersion
	}
	//update cluster to send updated cluster details to core
	err = cluster.UpdateCluster(cluster_resp)
	if err != nil {
		log.Printf("cluster was not updated, error %s", err.Error())
		return diag.FromErr(err)
	}

	return diags
}

func resourceImportClusterDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("resource imported cluster delete id %s", d.Id())

	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}
	importCluster, err := expandImportCluster(d)

	//get project id with project name, p.id used to refer to project id -> need p.ID for calling deleteCluster
	resp, err := project.GetProjectByName(importCluster.Metadata.Project)
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
	//delete cluster once project id is retrieved correctly
	err = cluster.DeleteCluster(importCluster.Metadata.Name, project_id)
	if err != nil {
		fmt.Print("cluster was not deleted")
		log.Printf("cluster was not deleted, error %s", err.Error())
		return diag.FromErr(err)
	}

	return diags
}
