package rafay

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/RafaySystems/rctl/pkg/cluster"
	"github.com/RafaySystems/rctl/pkg/config"
	"github.com/RafaySystems/rctl/pkg/project"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
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
			"metadata": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: metadataFields(),
				},
				MinItems: 1,
				MaxItems: 1,
			},
		},
		//		Description:   resourceEksPodIdentityDescription,
	}
}

func metadataFields() map[string]*schema.Schema {
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
	var diags diag.Diagnostics
	log.Printf("pod identity update starts")
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	var podIdentity []*IAMPodIdentityAssociation
	metadata := &Metadata{}
	//rawConfig := d.GetRawConfig()

	if v, ok := d.Get("spec").([]interface{}); ok && len(v) > 0 {
		podIdentity = expandIAMPodIdentityAssociationsConfig(v)
	} else {
		return diag.FromErr(fmt.Errorf("spec not specified"))
	}

	if d.HasChange("spec") {
		oldRaw, newRaw := d.GetChange("spec")

		oldFlag := extractCreateServiceAccount(oldRaw)
		newFlag := extractCreateServiceAccount(newRaw)

		if oldFlag != newFlag {
			return diag.Errorf(
				"create_service_account is immutable. ")
		}
	}

	if len(podIdentity) == 0 {
		return diag.FromErr(errors.New("could not get pod identity associations"))
	}

	if v, ok := d.Get("metadata").([]interface{}); ok && len(v) > 0 {
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

	cluster_resp, err := cluster.GetCluster(metadata.clusterName, project_id, "terraform")
	if err != nil {
		log.Printf("imported cluster was not created, error %s", err.Error())
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	uri := fmt.Sprintf("/edge/v1/projects/%s/edges/%s/podidentity?user_agent=%s", project_id, cluster_resp.ID, "terraform")

	log.Printf("payload response : %s", podIdentity[0].Namespace)

	response, err := auth.AuthAndRequest(uri, "PUT", podIdentity)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("Update Pod Identity response : %s", response)

	ticker := time.NewTicker(time.Duration(5) * time.Second)
	defer ticker.Stop()

	uniqueId := podIdentity[0].ServiceAccountName + "/" + podIdentity[0].Namespace
	d.SetId(uniqueId)

LOOP:
	for {
		select {
		case <-ctx.Done():
			log.Printf("pod identity %s operation timed out", uniqueId)
			return diag.FromErr(fmt.Errorf("pod identity %s operation timed out", uniqueId))
		case <-ticker.C:
			status, comments, err := getPodIdentityStatus(podIdentity[0], project_id, cluster_resp.ID)
			if err != nil {
				log.Println("error in getting pod identity status", err)
				return diag.FromErr(err)
			}

			switch status {
			case "POD_IDENTITY_UPDATION_COMPLETE":
				log.Printf("pod identity %s operation completed", uniqueId)
				break LOOP

			case "POD_IDENTITY_UPDATION_FAILED":
				log.Printf("pod identity %s operation failed", uniqueId)
				return diag.Errorf("pod identity %s operation failed with errors: %s", uniqueId, comments)

			case "POD_IDENTITY_UPDATION_IN_PROGRESS", "POD_IDENTITY_UPDATION_PENDING":
				log.Printf("pod identity %s operation", uniqueId)

			}
		}
	}
	return diags
}

func resourceEksPodIdentityUpsert(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("pod identity upsert starts")
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	var podIdentity []*IAMPodIdentityAssociation
	metadata := &Metadata{}
	//rawConfig := d.GetRawConfig()

	if v, ok := d.Get("spec").([]interface{}); ok && len(v) > 0 {
		podIdentity = expandIAMPodIdentityAssociationsConfig(v)
	} else {
		return diag.FromErr(fmt.Errorf("spec not specified"))
	}

	if len(podIdentity) == 0 {
		return diag.FromErr(errors.New("could not get pod identity associations"))
	}

	if v, ok := d.Get("metadata").([]interface{}); ok && len(v) > 0 {
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

	cluster_resp, err := cluster.GetCluster(metadata.clusterName, project_id, "terraform")
	if err != nil {
		log.Printf("imported cluster was not created, error %s", err.Error())
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	uri := fmt.Sprintf("/edge/v1/projects/%s/edges/%s/podidentity?user_agent=%s", project_id, cluster_resp.ID, "terraform")

	log.Printf("payload response : %s", podIdentity[0].Namespace)

	response, err := auth.AuthAndRequest(uri, "POST", podIdentity)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("Create Pod Identity response : %s", response)

	ticker := time.NewTicker(time.Duration(5) * time.Second)
	defer ticker.Stop()

	uniqueId := podIdentity[0].ServiceAccountName + "/" + podIdentity[0].Namespace
	d.SetId(uniqueId)
	time.Sleep(5 * time.Second)

LOOP:
	for {
		select {
		case <-ctx.Done():
			log.Printf("pod identity %s operation timed out", uniqueId)
			return diag.FromErr(fmt.Errorf("pod identity %s operation timed out", uniqueId))
		case <-ticker.C:
			status, comments, err := getPodIdentityStatus(podIdentity[0], project_id, cluster_resp.ID)
			if err != nil {
				log.Println("error in getting pod identity status", err)
				return diag.FromErr(err)
			}

			switch status {
			case "POD_IDENTITY_CREATION_COMPLETE":
				log.Printf("pod identity %s operation completed", uniqueId)
				break LOOP

			case "POD_IDENTITY_CREATION_FAILED":
				log.Printf("pod identity %s operation failed", uniqueId)
				d.SetId("")
				return diag.Errorf("pod identity %s operation failed with errors: %s", uniqueId, comments)

			case "POD_IDENTITY_CREATION_IN_PROGRESS", "POD_IDENTITY_CREATION_PENDING":
				log.Printf("pod identity %s operation", uniqueId)

			}
		}
	}
	return diags

}

func getPodIdentityStatus(podIdentity *IAMPodIdentityAssociation, projectId, clusterId string) (string, string, error) {

	auth := config.GetConfig().GetAppAuthProfile()
	uri := fmt.Sprintf("/edge/v1/projects/%s/edges/%s/podidentity/%s/%s", projectId, clusterId, podIdentity.Namespace, podIdentity.ServiceAccountName)

	response, err := auth.AuthAndRequest(uri, "GET", "")
	if err != nil {
		return "", "", err
	}
	log.Printf("Get Pod Identity response : %s", response)

	decoder := json.NewDecoder(bytes.NewReader([]byte(response)))

	piaSpec := []*IAMPodIdentityAssociationOutput{}

	if err := decoder.Decode(&piaSpec); err != nil {
		log.Println("error decoding pod identity spec")
		return "", "", err
	}

	if len(piaSpec) == 0 {
		return "", "", err
	}

	status := piaSpec[0].Status
	comments := piaSpec[0].Comments

	return status, comments, nil
}

func resourceEksPodIdentityRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Println("READ POD IDENTITY START")
	var diags diag.Diagnostics
	clusterName, ok := d.Get("metadata.0.cluster_name").(string)
	if !ok || clusterName == "" {
		log.Print("Cluster name unable to be found")
		return diag.FromErr(fmt.Errorf("%s", "cluster name is missing"))
	}
	projectName, ok := d.Get("metadata.0.project_name").(string)
	if !ok || projectName == "" {
		log.Print("Cluster project name unable to be found")
		return diag.FromErr(fmt.Errorf("%s", "project name is missing"))
	}
	project_id, edge_id, err := getIdFromName(clusterName, projectName)
	if err != nil {
		return diag.FromErr(err)
	}

	svc_name, ok := d.Get("spec.0.service_account_name").(string)
	if !ok || svc_name == "" {
		log.Print("Svc name unable to be found")
		return diag.FromErr(fmt.Errorf("%s", "service account name is missing"))
	}
	namespace, ok := d.Get("spec.0.namespace").(string)
	if !ok || namespace == "" {
		log.Print("namespace unable to be found")
		return diag.FromErr(fmt.Errorf("%s", "namespace is missing"))
	}
	auth := config.GetConfig().GetAppAuthProfile()
	uri := fmt.Sprintf("/edge/v1/projects/%s/edges/%s/podidentity/%s/%s", project_id, edge_id, namespace, svc_name)

	response, err := auth.AuthAndRequest(uri, "GET", "")
	if err != nil {
		return diag.FromErr(err)
	}

	log.Println("pod identity get response : ", response)

	decoder := json.NewDecoder(bytes.NewReader([]byte(response)))

	piaSpec := []*IAMPodIdentityAssociation{}

	piaStatus := []*IAMPodIdentityAssociationOutput{}

	if err := decoder.Decode(&piaSpec); err != nil {
		log.Println("error decoding pod identity spec")
		return diag.FromErr(err)
	}

	decoder = json.NewDecoder(bytes.NewReader([]byte(response)))

	if err := decoder.Decode(&piaStatus); err != nil {
		log.Println("error decoding pod identity status spec")
		return diag.FromErr(err)
	}

	v, ok := d.Get("spec").([]interface{})
	if !ok {
		v = []interface{}{}
	}

	spec := flattenIAMPodIdentityAssociations(piaSpec, v)
	log.Printf("After flatten spec %s", spec)
	err = d.Set("spec", spec)
	if err != nil {
		log.Printf("err setting pia spec %s", err.Error())
		return diag.FromErr(err)
	}

	return diags
}

func resourceEksPodIdentityDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	clusterName, ok := d.Get("metadata.0.cluster_name").(string)
	if !ok || clusterName == "" {
		log.Print("Cluster name unable to be found")
		return diag.FromErr(fmt.Errorf("%s", "cluster name is missing"))
	}
	projectName, ok := d.Get("metadata.0.project_name").(string)
	if !ok || projectName == "" {
		log.Print("Cluster project name unable to be found")
		return diag.FromErr(fmt.Errorf("%s", "project name is missing"))
	}

	project_id, edge_id, err := getIdFromName(clusterName, projectName)
	if err != nil {
		return diag.FromErr(err)
	}

	var podIdentity []*IAMPodIdentityAssociation

	//rawConfig := d.GetRawConfig()

	if v, ok := d.Get("spec").([]interface{}); ok && len(v) > 0 {
		podIdentity = expandIAMPodIdentityAssociationsConfig(v)
	} else {
		return diag.FromErr(fmt.Errorf("spec not specified"))
	}

	if len(podIdentity) == 0 {
		return diag.FromErr(errors.New("could not get pod identity associations"))
	}

	log.Println("Delete started, pod identity to delete ", podIdentity[0].Namespace)

	// payload, err := json.Marshal(podIdentity)
	// if err != nil {
	// 	return diag.FromErr(err)
	// }

	auth := config.GetConfig().GetAppAuthProfile()
	uri := fmt.Sprintf("/edge/v1/projects/%s/edges/%s/podidentity?user_agent=%s", project_id, edge_id, "terraform")

	//log.Println("Delete started, pod identity to delete ", string(payload))

	response, err := auth.AuthAndRequest(uri, "DELETE", podIdentity)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Println("pod identity get response : ", response)

	return diags
}

func extractCreateServiceAccount(raw interface{}) bool {
	if l, ok := raw.([]interface{}); ok && len(l) > 0 && l[0] != nil {
		if m, ok := l[0].(map[string]interface{}); ok {
			if v, ok := m["create_service_account"].(bool); ok {
				return v
			}
		}
	}
	return false
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

	cluster_resp, err := cluster.GetCluster(clusterName, project_id, "terraform")
	if err != nil {
		log.Printf("imported cluster was not created, error %s", err.Error())
		return "", "", err
	}

	return project_id, cluster_resp.ID, nil
}
