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
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
)

type eksPodIdentityResourceAssociation struct {
	Namespace              string                 `json:"namespace" yaml:"namespace"`
	ServiceAccountName     string                 `json:"serviceAccountName" yaml:"serviceAccountName"`
	RoleARN                string                 `json:"roleARN,omitempty" yaml:"roleARN,omitempty"`
	CreateServiceAccount   *bool                  `json:"createServiceAccount,omitempty" yaml:"createServiceAccount,omitempty"`
	RoleName               string                 `json:"roleName,omitempty" yaml:"roleName,omitempty"`
	PermissionsBoundaryARN string                 `json:"permissionsBoundaryARN,omitempty" yaml:"permissionsBoundaryARN,omitempty"`
	PermissionPolicyARNs   []string               `json:"permissionPolicyARNs,omitempty" yaml:"permissionPolicyARNs,omitempty"`
	PermissionPolicy       map[string]interface{} `json:"permissionPolicy,omitempty" yaml:"permissionPolicy,omitempty"`
	WellKnownPolicies      *WellKnownPolicies     `json:"wellKnownPolicies,omitempty" yaml:"wellKnownPolicies,omitempty"`
	Tags                   map[string]string      `json:"tags,omitempty" yaml:"tags,omitempty"`
}

func eksPodIdentityResourceFields() map[string]*schema.Schema {
	fields := podIdentityAssociationsFields()
	fields["create_service_account"] = &schema.Schema{
		Type:        schema.TypeBool,
		Optional:    true,
		Computed:    true,
		Description: "Creates the Kubernetes service account when unset or true, and deletes a Rafay-created service account when false. An existing value is kept when the argument is removed.",
	}
	return fields
}

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
					Schema: eksPodIdentityResourceFields(),
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

func expandEKSPodIdentityResource(p []interface{}, rawSpec cty.Value) []*eksPodIdentityResourceAssociation {
	out := make([]*eksPodIdentityResourceAssociation, len(p))
	rawItems, haveRaw := podIdentityRawSlice(rawSpec)
	for i := range p {
		obj := &eksPodIdentityResourceAssociation{}
		in := p[i].(map[string]interface{})
		obj.Namespace, _ = in["namespace"].(string)
		obj.ServiceAccountName, _ = in["service_account_name"].(string)
		obj.RoleARN, _ = in["role_arn"].(string)
		// d.Get reports false for an omitted optional bool. The raw config keeps that null distinct from an explicit false.
		if haveRaw && i < len(rawItems) {
			if enabled, ok := configuredBool(rawItems[i], "create_service_account"); ok {
				obj.CreateServiceAccount = &enabled
			}
		}
		obj.RoleName, _ = in["role_name"].(string)
		obj.PermissionsBoundaryARN, _ = in["permission_boundary_arn"].(string)
		if value, ok := in["permission_policy"].(string); ok && value != "" {
			_ = json.Unmarshal([]byte(value), &obj.PermissionPolicy)
		}
		if value, ok := in["permission_policy_arns"].([]interface{}); ok {
			obj.PermissionPolicyARNs = toArrayString(value)
		}
		if value, ok := in["well_known_policies"].([]interface{}); ok && len(value) > 0 {
			obj.WellKnownPolicies = expandIAMWellKnownPolicies(value)
		}
		if value, ok := in["tags"].(map[string]interface{}); ok {
			obj.Tags = toMapString(value)
		}
		out[i] = obj
	}
	return out
}

// podIdentityRawSpec returns spec only when raw config is a known object that has it.
// Destroy passes a null config after the instance leaves configuration, and GetAttr panics on that null.
func podIdentityRawSpec(rawConfig cty.Value) cty.Value {
	if rawConfig.IsNull() || !rawConfig.IsKnown() || !rawConfig.Type().IsObjectType() || !rawConfig.Type().HasAttribute("spec") {
		return cty.NullVal(cty.DynamicPseudoType)
	}
	return rawConfig.GetAttr("spec")
}

func podIdentityRawSlice(value cty.Value) ([]cty.Value, bool) {
	if value.IsNull() || !value.IsKnown() {
		return nil, false
	}
	valueType := value.Type()
	if !valueType.IsListType() && !valueType.IsTupleType() && !valueType.IsSetType() {
		return nil, false
	}
	return value.AsValueSlice(), true
}

func configuredBool(obj cty.Value, name string) (bool, bool) {
	if obj.IsNull() || !obj.IsKnown() || !obj.Type().IsObjectType() || !obj.Type().HasAttribute(name) {
		return false, false
	}
	attr := obj.GetAttr(name)
	if attr.IsNull() || !attr.IsKnown() || !attr.Type().Equals(cty.Bool) {
		return false, false
	}
	return attr.True(), true
}

func flattenEKSPodIdentityResource(
	inp []*eksPodIdentityResourceAssociation,
	previous []interface{},
) []interface{} {
	if inp == nil {
		return nil
	}
	out := make([]interface{}, len(inp))
	for i, in := range inp {
		obj := map[string]interface{}{}
		if i < len(previous) && previous[i] != nil {
			obj = previous[i].(map[string]interface{})
		}
		obj["namespace"] = in.Namespace
		obj["service_account_name"] = in.ServiceAccountName
		obj["role_arn"] = in.RoleARN
		obj["role_name"] = in.RoleName
		obj["permission_boundary_arn"] = in.PermissionsBoundaryARN
		obj["permission_policy_arns"] = toArrayInterface(in.PermissionPolicyARNs)
		obj["tags"] = toMapInterface(in.Tags)
		if in.PermissionPolicy != nil {
			value, err := json.Marshal(in.PermissionPolicy)
			if err == nil {
				obj["permission_policy"] = string(value)
			}
		}
		if in.WellKnownPolicies != nil {
			value, ok := obj["well_known_policies"].([]interface{})
			if !ok {
				value = []interface{}{}
			}
			obj["well_known_policies"] = flattenIAMWellKnownPolicies(in.WellKnownPolicies, value)
		}
		if in.CreateServiceAccount != nil {
			obj["create_service_account"] = *in.CreateServiceAccount
		}
		out[i] = obj
	}
	return out
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

	var podIdentity []*eksPodIdentityResourceAssociation
	metadata := &Metadata{}

	if v, ok := d.Get("spec").([]interface{}); ok && len(v) > 0 {
		podIdentity = expandEKSPodIdentityResource(v, podIdentityRawSpec(d.GetRawConfig()))
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

	var podIdentity []*eksPodIdentityResourceAssociation
	metadata := &Metadata{}

	if v, ok := d.Get("spec").([]interface{}); ok && len(v) > 0 {
		podIdentity = expandEKSPodIdentityResource(v, podIdentityRawSpec(d.GetRawConfig()))
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

func getPodIdentityStatus(podIdentity *eksPodIdentityResourceAssociation, projectId, clusterId string) (string, string, error) {

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
		log.Printf("resourceEksPodIdentityRead getIdFromName error %s", err.Error())
		if IsResourceNotFoundErr(err) {
			log.Println("resourceEksPodIdentityRead: cluster not found, treating as drift", "error", err)
			d.SetId("")
			return diags
		}
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
		log.Printf("resourceEksPodIdentityRead get pod identity error %s", err.Error())
		if IsResourceNotFoundErr(err) {
			log.Println("resourceEksPodIdentityRead: pod identity not found, treating as drift", "error", err)
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	log.Println("pod identity get response : ", response)

	decoder := json.NewDecoder(bytes.NewReader([]byte(response)))

	piaSpec := []*eksPodIdentityResourceAssociation{}

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

	spec := flattenEKSPodIdentityResource(piaSpec, v)
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

	podIdentity := []*eksPodIdentityResourceAssociation{}

	if v, ok := d.Get("spec").([]interface{}); ok && len(v) > 0 {
		podIdentity = expandEKSPodIdentityResource(v, podIdentityRawSpec(d.GetRawConfig()))
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
