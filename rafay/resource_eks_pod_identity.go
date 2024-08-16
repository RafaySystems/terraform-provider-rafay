package rafay

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/RafaySystems/rctl/pkg/cluster"
	"github.com/RafaySystems/rctl/pkg/config"
	"github.com/RafaySystems/rctl/pkg/project"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceEKSPodIdentity() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceEksPodIdentityCreate,
		ReadContext:   resourceEksPodIdentityRead,
		UpdateContext: resourceEksPodIdentityUpdate,
		DeleteContext: resourceEksPodIdentityDelete,
		// Importer: &schema.ResourceImporter{
		// 	State: resourceEksPodIdentityImport,
		// },
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"spec": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: podIdentityAssociationsFields(),
				},
				MinItems: 1,
				MaxItems: 1,
			},
			"host_metadata": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: hostMetadataFields(),
				},
				MinItems: 1,
				MaxItems: 1,
			},
		},
		//		Description:   resourceEksPodIdentityDescription,
	}
}

func hostMetadataFields() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"cluster_name": {
			Type:     schema.TypeString,
			Required: true,
		},
		"project_name": {
			Type:     schema.TypeString,
			Required: true,
		},
	}

	return s
}

func resourceEksPodIdentityCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceEksPodIdentityUpsert(ctx, d, m)
}

func resourceEksPodIdentityUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceEksPodIdentityUpsert(ctx, d, m)
}

func resourceEksPodIdentityUpsert(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("pod identity upsert starts")
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	var podIdentity []*IAMPodIdentityAssociation
	metadata := &HostMetadata{}
	//rawConfig := d.GetRawConfig()

	if v, ok := d.Get("spec").([]interface{}); ok && len(v) > 0 {
		podIdentity = expandIAMPodIdentityAssociationsConfig(v)
	} else {
		return diag.FromErr(fmt.Errorf("spec not specified"))
	}

	if v, ok := d.Get("host_metadata").([]interface{}); ok && len(v) > 0 {
		in := v[0].(map[string]interface{})
		metadata.clusterName = in["cluster_name"].(string)
		metadata.projectName = in["project_name"].(string)
	}

	resp, err := project.GetProjectByName(metadata.projectName)
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

	cluster_resp, err := cluster.GetCluster(metadata.clusterName, project_id)
	if err != nil {
		log.Printf("imported cluster was not created, error %s", err.Error())
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	uri := fmt.Sprintf("/edge/v1/projects/%s/edges/%s/podidentity", project_id, cluster_resp.ID)

	// payload, err := json.Marshal(podIdentity)
	// if err != nil {
	// 	return diag.FromErr(err)
	// }

	log.Printf("payload response : %s", podIdentity[0].Namespace)

	response, err := auth.AuthAndRequest(uri, "POST", podIdentity)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("process_filebytes response : %s", response)

	return diags

}

func resourceEksPodIdentityRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	clusterName, ok := d.Get("host_metadata.0.cluster_name").(string)
	if !ok || clusterName == "" {
		log.Print("Cluster name unable to be found")
		return diag.FromErr(fmt.Errorf("%s", "cluster name is missing"))
	}
	projectName, ok := d.Get("host_metadata.0.project_name").(string)
	if !ok || projectName == "" {
		log.Print("Cluster project name unable to be found")
		return diag.FromErr(fmt.Errorf("%s", "project name is missing"))
	}
	project_id, edge_id, err := getIdFromName(clusterName, projectName)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	uri := fmt.Sprintf("/edge/v1/projects/%s/edges/%s/podidentity", project_id, edge_id)

	response, err := auth.AuthAndRequest(uri, "GET", "")
	if err != nil {
		return diag.FromErr(err)
	}

	log.Println("pod identity get response : ", response)

	decoder := json.NewDecoder(bytes.NewReader([]byte(response)))

	piaSpec := []*IAMPodIdentityAssociation{}

	if err := decoder.Decode(&piaSpec); err != nil {
		log.Println("error decoding pod identity spec")
		return diag.FromErr(err)
	}

	v, ok := d.Get("spec").([]interface{})
	if !ok {
		v = []interface{}{}
	}

	pia := flattenIAMPodIdentityAssociations(piaSpec, v)

	err = d.Set("spec", pia)
	if err != nil {
		log.Printf("err setting pia spec %s", err.Error())
		return diag.FromErr(err)
	}

	return diags
}

func resourceEksPodIdentityDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	clusterName, ok := d.Get("host_metadata.0.cluster_name").(string)
	if !ok || clusterName == "" {
		log.Print("Cluster name unable to be found")
		return diag.FromErr(fmt.Errorf("%s", "cluster name is missing"))
	}
	projectName, ok := d.Get("host_metadata.0.project_name").(string)
	if !ok || projectName == "" {
		log.Print("Cluster project name unable to be found")
		return diag.FromErr(fmt.Errorf("%s", "project name is missing"))
	}

	project_id, edge_id, err := getIdFromName(clusterName, projectName)
	if err != nil {
		return diag.FromErr(err)
	}

	var podIdentity *IAMPodIdentityAssociation
	//rawConfig := d.GetRawConfig()

	if v, ok := d.Get("spec").([]interface{}); ok && len(v) > 0 {
		podIdentity = expandIAMPodIdentityAssociationsConfig(v)[0]
	} else {
		return diag.FromErr(fmt.Errorf("spec not specified"))
	}

	payload, err := json.Marshal(podIdentity)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	uri := fmt.Sprintf("/edge/v1/projects/%s/edges/%s/podidentity", project_id, edge_id)

	response, err := auth.AuthAndRequest(uri, "DELETE", string(payload))
	if err != nil {
		return diag.FromErr(err)
	}

	log.Println("pod identity get response : ", response)

	return diags
}

func getIdFromName(clusterName string, projectName string) (string, string, error) {
	resp, err := project.GetProjectByName(projectName)
	if err != nil {
		log.Printf("project does not exist, error %s", err.Error())
		return "", "", err
	}
	p, err := project.NewProjectFromResponse([]byte(resp))
	if err != nil {
		return "", "", err
	} else if p == nil {
		return "", "", err
	}
	project_id := p.ID

	cluster_resp, err := cluster.GetCluster(clusterName, project_id)
	if err != nil {
		log.Printf("imported cluster was not created, error %s", err.Error())
		return "", "", err
	}

	return project_id, cluster_resp.ID, nil
}
