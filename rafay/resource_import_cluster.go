package rafay

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/RafaySystems/rctl/pkg/cluster"
	"github.com/RafaySystems/rctl/pkg/project"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
		},
	}
}

func resourceImportClusterCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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

	//set ID for imported cluster id, d.SetID()
	d.SetId(cluster_resp.ID)
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
	time.Sleep(60 * time.Second)
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
	c, err := cluster.GetCluster(d.Get("clustername").(string), project.ID)
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
	cluster_resp, err := cluster.GetCluster(d.Get("clustername").(string), project_id)
	if err != nil {
		log.Printf("imported cluster was not created, error %s", err.Error())
		return diag.FromErr(err)
	}
	// read the blueprint name
	if d.Get("blueprint").(string) != "" {
		cluster_resp.ClusterBlueprint = d.Get("blueprint").(string)
	}
	// read the blueprint version
	if d.Get("blueprint_version").(string) != "" {
		cluster_resp.ClusterBlueprintVersion = d.Get("blueprint_version").(string)
	}
	//update cluster to send updated cluster details to core
	err = cluster.UpdateCluster(cluster_resp)
	if err != nil {
		log.Printf("cluster was not updated, error %s", err.Error())
		return diag.FromErr(err)
	}

	//publish cluster bp
	err = cluster.PublishClusterBlueprint(d.Get("clustername").(string), project_id)
	if err != nil {
		log.Printf("cluster was not updated, error %s", err.Error())
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
	err = cluster.DeleteCluster(d.Get("clustername").(string), project_id, false)
	if err != nil {
		fmt.Print("cluster was not deleted")
		log.Printf("cluster was not deleted, error %s", err.Error())
		return diag.FromErr(err)
	}

	return diags
}
