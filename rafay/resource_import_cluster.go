package rafay

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/RafaySystems/rctl/pkg/cluster"
	"github.com/RafaySystems/rctl/pkg/project"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

/*
type configMetadata struct {
	Name    string `yaml:"name"`
	Project string `yaml:"project"`
}

type configResourceType struct {
	Meta *configMetadata `yaml:"metadata"`
}
*/
func resourceImportCluster() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceEKSClusterCreate,
		ReadContext:   resourceEKSClusterRead,
		UpdateContext: resourceEKSClusterUpdate,
		DeleteContext: resourceEKSClusterDelete,

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
			"location": {
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
	//c := config.GetConfig()
	//logger := glogger.GetLogger()

	/*YamlConfigFilePath := d.Get("yamlfilepath").(string)

	fileBytes, err := utils.ReadYAMLFileContents(YamlConfigFilePath)
	if err != nil {
		return diag.FromErr(err)
	}

	// split the file and update individual resources
	y, uerr := utils.SplitYamlAndGetListByKind(fileBytes)
	if uerr != nil {
		return diag.FromErr(err)
	}*/
	//create imported cluster
	resp, err := cluster.NewImportCluster(d.Get("name").(string), d.Get("blueprint").(string), d.Get("location").(string), d.Get("projectname").(string))
	if err != nil {
		log.Printf("create import cluster failed to create (check parameters passed in), error %s", err.Error())
		return diag.FromErr(err)
	}
	//get project id with project name, p.id used to refer to project id -> need p.ID for calling getCluster and GetBootstrapFile
	resp, err = project.GetProjectByName(d.Get("name").(string))
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
	//make sure new imported cluster was created by calling get cluster and checking for no errors
	cluster_resp, err := cluster.GetCluster(d.Get("name").(string), project_id)
	if err != nil {
		log.Printf("imported cluster was not created, error %s", err.Error())
		return diag.FromErr(err)
	}
	//then retrieve bootstrap yaml file, call GetBootstrapFile() -> make sure this function downloads the bootstrap file locally (i think the url request does)
	_, err = cluster.GetBootstrapFile(d.Get("name").(string), project_id)
	if err != nil {
		log.Printf("bootstrap yaml file was not obtained correctly, error %s", err.Error())
		return diag.FromErr(err)
	}
	//figure out how to apply bootstrap yaml file to created cluster STILL NEED TO COMPLETE

	//set ID for imported cluster id, d.SetID()
	d.SetId(cluster_resp.ID)
	return diags
	/*
		var rafayConfigs, clusterConfigs [][]byte
		rafayConfigs = y["Cluster"]
		clusterConfigs = y["ClusterConfig"]
		if len(rafayConfigs) > 1 {
			return diag.FromErr(fmt.Errorf("%s", "only one cluster per config is supported"))
		}
		for _, yi := range rafayConfigs {
			log.Println("rafayConfig:", string(yi))
			name, project, err := findResourceNameFromConfig(yi)
			if err != nil {
				return diag.FromErr(fmt.Errorf("%s", "failed to get cluster name"))
			}
			log.Println("rafayConfig name:", name, "project:", project)
			if name != d.Get("name").(string) {
				return diag.FromErr(fmt.Errorf("%s", "cluster name does not match config file "))
			}
			if project != d.Get("projectname").(string) {
				return diag.FromErr(fmt.Errorf("%s", "project name does not match config file"))
			}
		}

		for _, yi := range clusterConfigs {
			log.Println("clusterConfig", string(yi))
			name, _, err := findResourceNameFromConfig(yi)
			if err != nil {
				return diag.FromErr(fmt.Errorf("%s", "failed to get cluster name"))
			}
			if name != d.Get("name").(string) {
				return diag.FromErr(fmt.Errorf("%s", "ClusterConfig name does not match config file"))
			}
		}

		// get project details
		resp, err := project.GetProjectByName(d.Get("projectname").(string))
		if err != nil {
			fmt.Print("project does not exist")
			return diags
		}
		project, err := project.NewProjectFromResponse([]byte(resp))
		if err != nil {
			fmt.Printf("project does not exist")
			return diags
		}

		// override config project
		c.ProjectID = project.ID

		configMap, errs := collateConfigsByName(rafayConfigs, clusterConfigs)
		if len(errs) > 0 {
			for _, err := range errs {
				log.Println("error in collateConfigsByName", err)
			}
			return diag.FromErr(fmt.Errorf("%s", "failed in collateConfigsByName"))
		}

		// Make request
		for clusterName, configBytes := range configMap {
			log.Println("create cluster:", clusterName, "config:", string(configBytes), "projectID :", c.ProjectID)
			if err := clusterctl.Apply(logger, c, clusterName, configBytes, false); err != nil {
				return diag.FromErr(fmt.Errorf("error performing apply on cluster %s: %s", clusterName, err))
			}
		}

		s, err := cluster.GetCluster(d.Get("name").(string), project.ID)
		if err != nil {
			log.Printf("error while getCluster %s", err.Error())
			return diag.FromErr(err)
		}

		log.Printf("resource eks cluster created %s", s.ID)
		d.SetId(s.ID)
	*/

}

//NOT COMPLETE!!!!!
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
	c, err := cluster.GetCluster(d.Get("name").(string), project.ID)
	if err != nil {
		log.Printf("error in get cluster %s", err.Error())
		return diag.FromErr(err)
	}
	if err := d.Set("name", c.Name); err != nil {
		log.Printf("get group set name error %s", err.Error())
		return diag.FromErr(err)
	}

	return diags
}

func resourceImportClusterUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("update imported cluster resource")

	//retrieve project_id from project name for calling get_cluster
	resp, err := project.GetProjectByName(d.Get("name").(string))
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
	cluster_resp, err := cluster.GetCluster(d.Get("name").(string), project_id)
	if err != nil {
		log.Printf("imported cluster was not created, error %s", err.Error())
		return diag.FromErr(err)
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

	//get project id with project name, p.id used to refer to project id -> need p.ID for calling deleteCluster
	resp, err := project.GetProjectByName(d.Get("name").(string))
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

	err = cluster.DeleteCluster(d.Get("name").(string), project_id)
	if err != nil {
		fmt.Print("cluster was not deleted")
		return diags
	}

	return diags
}
