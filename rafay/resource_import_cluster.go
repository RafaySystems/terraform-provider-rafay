package rafay

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/RafaySystems/rctl/pkg/cluster"
	"github.com/RafaySystems/rctl/pkg/config"
	"github.com/RafaySystems/rctl/pkg/models"
	"github.com/RafaySystems/rctl/pkg/project"
	"github.com/RafaySystems/rctl/pkg/rerror"
	"k8s.io/utils/strings/slices"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var (
	supportedK8sProviderList    []string = []string{"AKS", "EKS", "GKE", "OPENSHIFT", "OTHER", "RKE", "EKSANYWHERE"}
	supportedProvisionEnvList   []string = []string{"CLOUD", "ONPREM"}
	environmentManagerLabelsKey []string = []string{"rafay.dev/envRun", "rafay.dev/k8sVersion"}
)

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
			"clustername": {
				Type:     schema.TypeString,
				Required: true,
			},
			"projectname": {
				Type:     schema.TypeString,
				Required: true,
			},
			"blueprint": {
				Type:     schema.TypeString,
				Required: true,
			},
			"blueprint_version": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"location": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"kubeconfig_path": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"bootstrap_path": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					// Suppress the diff if the value is still the same (or is computed to be the same)
					// We suppress the diff because the value is expected to be computed and shouldn't change unexpectedly.
					return old == new
				},
			},
			"values_path": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					// Suppress the diff if the value is still the same (or is computed to be the same)
					// We suppress the diff because the value is expected to be computed and shouldn't change unexpectedly.
					return old == new
				},
			},
			"values_data": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bootstrap_data": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"labels": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"kubernetes_provider": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "OTHER",
				ValidateDiagFunc: func(i interface{}, p cty.Path) diag.Diagnostics {
					k8sProvider := i.(string)
					if !slices.Contains(supportedK8sProviderList, k8sProvider) {
						return diag.Errorf("Unsupported K8s Provider.Please refer list of K8s Provider supported: %v", supportedK8sProviderList)
					}
					return diag.Diagnostics{}
				},
			},
			"provision_environment": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "CLOUD",
				ValidateDiagFunc: func(i interface{}, p cty.Path) diag.Diagnostics {
					provisionEnv := i.(string)
					if !slices.Contains(supportedProvisionEnvList, provisionEnv) {
						return diag.Errorf("Unsupported Provision Environment.Please refer list of Environment supported: %v", supportedProvisionEnvList)
					}
					return diag.Diagnostics{}
				},
			},
		},
	}
}

func getClusterlabels(name, projectId string) (map[string]string, error) {
	uri := fmt.Sprintf("/v2/scheduler/project/%s/cluster/%s", projectId, name)
	auth := config.GetConfig().GetAppAuthProfile()
	respString, err := auth.AuthAndRequest(uri, "GET", nil)
	if err != nil {
		return nil, rerror.CrudErr{
			Type: "cluster labels",
			Name: name,
			Op:   "get",
		}
	}

	var resp struct {
		Metadata struct {
			Labels map[string]string `json:"labels,omitempty"`
		} `json:"metadata,omitempty"`
	}

	if err := json.Unmarshal([]byte(respString), &resp); err != nil {
		log.Printf("Error unmarshaling response from get v2 cluster: %s", err)
		return nil, err
	}

	labels := map[string]string{}
	for k, v := range resp.Metadata.Labels {
		if !strings.HasPrefix(k, "rafay.dev/") || slices.Contains(environmentManagerLabelsKey, k) {
			labels[k] = v
		}
	}
	return labels, nil
}

func updateClusterLabels(name, edgeId, projectId string, labels map[string]string) error {
	uri := fmt.Sprintf("/edge/v1/projects/%s/edges/%s/labels/", projectId, edgeId)
	uri += fmt.Sprintf("?user_agent=%s", uaDef)
	auth := config.GetConfig().GetAppAuthProfile()
	_, err := auth.AuthAndRequest(uri, "PUT", labels)
	if err != nil {
		return rerror.CrudErr{
			Type: "cluster labels",
			Name: name,
			Op:   "update",
		}
	}
	return nil
}

func GetValuesFile(name, project string) (string, error) {
	auth := config.GetConfig().GetAppAuthProfile()
	uri := fmt.Sprintf("/v2/scheduler/project/%s/cluster/%s/download/valuesyaml", project, name)
	resp, err := auth.AuthAndRequest(uri, "GET", nil)
	if err != nil {
		return "", rerror.CrudErr{
			Type: "cluster bootstrap",
			Name: name,
			Op:   "get",
		}
	}

	f := &models.BootstrapFileDownload{}
	err = json.Unmarshal([]byte(resp), f)
	if err != nil {
		return "", err
	}

	b, err := base64.StdEncoding.DecodeString(f.Data)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func resourceImportClusterCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var bootstrap_path, values_path string

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
	_, err = cluster.NewImportClusterWithProvisionParams(d.Get("clustername").(string), d.Get("blueprint").(string), d.Get("location").(string), project_id, d.Get("blueprint_version").(string), d.Get("provision_environment").(string), d.Get("kubernetes_provider").(string))
	if err != nil {
		log.Printf("create import cluster failed to create (check parameters passed in), error %s", err.Error())
		return diag.FromErr(err)
	}

	time.Sleep(10 * time.Second)
	//if error with get cluster add a sleep to wait for cluster creation
	//make sure new imported cluster was created by calling get cluster and checking for no errors
	cluster_resp, err := cluster.GetCluster(d.Get("clustername").(string), project_id, "")
	if err != nil {
		log.Printf("imported cluster was not created, error %s", err.Error())
		return diag.FromErr(err)
	}

	//set ID for imported cluster id, d.SetID()
	d.SetId(cluster_resp.ID)
	if d.Get("blueprint_version").(string) != "" {
		cluster_resp.ClusterBlueprintVersion = d.Get("blueprint_version").(string)
		err = cluster.UpdateCluster(cluster_resp, uaDef)
		if err != nil {
			log.Printf("setting cluster blueprint version failed, error %s", err.Error())
			return diag.FromErr(err)
		}
	}

	if labelsX, ok := d.Get("labels").(map[string]interface{}); ok && len(labelsX) > 0 {
		labels := map[string]string{}
		for k, v := range labelsX {
			labels[k] = v.(string)
		}
		if err := updateClusterLabels(cluster_resp.Name, cluster_resp.ID, cluster_resp.ProjectID, labels); err != nil {
			log.Printf("error setting labels on the cluster: %s", err.Error())
			return diag.FromErr(err)
		}
	}

	values_file, err := GetValuesFile(d.Get("clustername").(string), project_id)
	if err != nil {
		log.Printf("values yaml file was not obtained correctly, error %s", err.Error())
		return diag.FromErr(err)
	}
	log.Println("values_filepath fetched correctly: \n", values_file)
	//write values file into values file path
	if (d.Get("values_path").(string)) != "" {
		log.Printf("Saving values file to: %s", d.Get("values_path").(string))
		values_path = d.Get("values_path").(string)
	} else {
		values_filename := d.Get("clustername").(string) + "-values.yaml"
		values_path, _ = filepath.Abs(values_filename)
		log.Printf("Saving values file to: %s", values_path)
		d.Set("values_path", values_path)
	}
	fv, err := os.Create(values_path)

	if err != nil {
		log.Fatal(err)
	}

	defer fv.Close()

	_, err2 := fv.WriteString(values_file)

	if err2 != nil {
		log.Printf("values yaml file was not written correctly, error %s", err2.Error())
		return diag.FromErr(err2)
	}
	d.Set("values_data", values_file)

	//then retrieve bootstrap yaml file, call GetBootstrapFile() -> make sure this function downloads the bootstrap file locally (i think the url request does)
	bootstrap_file, err := cluster.GetBootstrapFile(d.Get("clustername").(string), project_id)
	if err != nil {
		log.Printf("bootstrap yaml file was not obtained correctly, error %s", err.Error())
		return diag.FromErr(err)
	}
	log.Println("bootstrap_filepath got correctly: \n", bootstrap_file)
	//write bootstrap file into bootstrap file path
	if (d.Get("bootstrap_path").(string)) != "" {
		log.Printf("Saving bootstrap file to: %s", d.Get("bootstrap_path").(string))
		bootstrap_path = d.Get("bootstrap_path").(string)
	} else {
		bootstrap_path, _ = filepath.Abs("bootstrap.yaml")
		log.Printf("Saving bootstrap file to: %s", bootstrap_path)
		d.Set("bootstrap_path", bootstrap_path)
	}
	f, err := os.Create(bootstrap_path)

	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	_, err2 = f.WriteString(bootstrap_file)

	if err2 != nil {
		log.Printf("bootstrap yaml file was not written correctly, error %s", err2.Error())
		return diag.FromErr(err2)
	}
	d.Set("bootstrap_data", bootstrap_file)
	//pass in bootstrap file path into exec command
	// bootstrap_filepath, _ := filepath.Abs("bootstrap.yaml")
	//figure out how to apply bootstrap yaml file to created cluster STILL NEED TO COMPLETE
	//add kube_config file as optional schema, call os/exec to cal kubectl apply on the filepath to kube config
	time.Sleep(60 * time.Second)
	if (d.Get("kubeconfig_path").(string)) != "" {
		cmd := exec.Command("kubectl", "--kubeconfig", d.Get("kubeconfig_path").(string), "apply", "-f", bootstrap_path)
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

	return diags

}

func resourceImportClusterRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	resp, err := project.GetProjectByName(d.Get("projectname").(string))
	if err != nil {
		fmt.Print("project name missing in the resource")
		return diags
	}

	project, err := project.NewProjectFromResponse([]byte(resp))
	if err != nil {
		fmt.Printf("project does not exist")
		return diags
	}
	c, err := cluster.GetCluster(d.Get("clustername").(string), project.ID, "")
	if err != nil {
		log.Printf("error in get cluster %s", err.Error())
		if strings.Contains(err.Error(), "not found") {
			log.Println("Resource Read ", "error", err)
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}
	if err := d.Set("clustername", c.Name); err != nil {
		log.Printf("get group set name error %s", err.Error())
		return diag.FromErr(err)
	}

	labels, err := getClusterlabels(c.Name, c.ProjectID)
	if err != nil {
		log.Printf("error getting cluster v2 labels: %s", err.Error())
		return diag.FromErr(err)
	}
	if err := d.Set("labels", labels); err != nil {
		log.Printf("set labels error %s", err.Error())
		return diag.FromErr(err)
	}

	return diags
}

func resourceImportClusterUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("update imported cluster resource")

	//retrieve project_id from project name for calling get_cluster
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
	//retrieve cluster_details from get cluster to pass into update cluster
	cluster_resp, err := cluster.GetCluster(d.Get("clustername").(string), project_id, "")
	if err != nil {
		log.Printf("imported cluster was not created, error %s", err.Error())
		return diag.FromErr(err)
	}
	oldClusterBlueprint := cluster_resp.ClusterBlueprint
	// read the blueprint name
	if d.Get("blueprint").(string) != "" {
		cluster_resp.ClusterBlueprint = d.Get("blueprint").(string)
	}
	// read the blueprint version
	oldClusterBlueprintVersion := cluster_resp.ClusterBlueprintVersion
	if d.Get("blueprint_version").(string) != "" {
		cluster_resp.ClusterBlueprintVersion = d.Get("blueprint_version").(string)
	}
	//update cluster to send updated cluster details to core
	err = cluster.UpdateCluster(cluster_resp, uaDef)
	if err != nil {
		log.Printf("cluster was not updated, error %s", err.Error())
		return diag.FromErr(err)
	}

	//publish cluster bp
	if (cluster_resp.ClusterBlueprint != oldClusterBlueprint) || (cluster_resp.ClusterBlueprintVersion != oldClusterBlueprintVersion) {
		err = cluster.PublishClusterBlueprint(d.Get("clustername").(string), project_id, false)
		if err != nil {
			log.Printf("cluster was not updated, error %s", err.Error())
			return diag.FromErr(err)
		}
	}

	// update labels
	labels := map[string]string{}
	if labelsX, ok := d.Get("labels").(map[string]interface{}); ok && len(labelsX) > 0 {
		for k, v := range labelsX {
			labels[k] = v.(string)
		}
	}
	existingLabels, err := getClusterlabels(cluster_resp.Name, cluster_resp.ProjectID)
	if err != nil {
		log.Printf("error getting cluster v2 labels: %s", err.Error())
		return diag.FromErr(err)
	}
	log.Println("Debug--- existing labels: ", existingLabels)
	log.Println("Debug--- new labels: ", labels)

	for k := range labels {
		log.Println("Debug--- new label key: ", k)
		if strings.HasPrefix(k, "rafay.dev/") {
			if _, ok := existingLabels[k]; !ok {
				errMsg := "cannot edit system labels during update operation"
				log.Printf("error setting labels: %s", errMsg)
				return diag.Errorf("error setting labels: %s", errMsg)
			}
		}
	}
	if err := updateClusterLabels(cluster_resp.Name, cluster_resp.ID, cluster_resp.ProjectID, labels); err != nil {
		log.Printf("error setting labels on the cluster: %s", err.Error())
		return diag.FromErr(err)
	}

	return diags
}

func resourceImportClusterDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("resource imported cluster delete id %s", d.Id())

	//get project id with project name, p.id used to refer to project id -> need p.ID for calling deleteCluster
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
	//delete cluster once project id is retrieved correctly
	err = cluster.DeleteCluster(d.Get("clustername").(string), project_id, false, "")
	if err != nil {
		fmt.Print("cluster was not deleted")
		log.Printf("cluster was not deleted, error %s", err.Error())
		return diag.FromErr(err)
	}

	return diags
}
