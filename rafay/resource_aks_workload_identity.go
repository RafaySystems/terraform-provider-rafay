package rafay

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/RafaySystems/rafay-common/pkg/hub/client/options"
	typed "github.com/RafaySystems/rafay-common/pkg/hub/client/typed"
	"github.com/RafaySystems/rafay-common/pkg/hub/terraform/resource"
	"github.com/RafaySystems/rafay-common/proto/types/hub/infrapb"
	"github.com/RafaySystems/rctl/pkg/config"
	"github.com/RafaySystems/rctl/pkg/versioninfo"
	"github.com/davecgh/go-spew/spew"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	schema "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	CONN_TIMEOUT int64 = 120
)

func resourceAKSWorkloadIdentity() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAKSWorkloadIdentityCreate,
		ReadContext:   resourceAKSWorkloadIdentityRead,
		UpdateContext: resourceAKSWorkloadIdentityUpdate,
		DeleteContext: resourceAKSWorkloadIdentityDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(2 * time.Minute),
			Delete: schema.DefaultTimeout(2 * time.Minute),
		},

		SchemaVersion: 1,
		Schema:        resource.WorkloadIdentitySchema.Schema,
	}
}

func resourceAKSWorkloadIdentityCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Println("create aks workload identity")

	diags := resourceAKSWorkloadIdentityUpsert(ctx, d)
	return diags
}

func resourceAKSWorkloadIdentityRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	log.Println("read aks workload identity")

	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	desiredInfraAksWorkloadIdentity, err := expandAksWorkloadIdentity(d)
	if err != nil {
		return diag.FromErr(err)
	}

	wiName := desiredInfraAksWorkloadIdentity.Spec.Metadata.Name
	wiClusterName := desiredInfraAksWorkloadIdentity.Metadata.Clustername
	wiProjectName := desiredInfraAksWorkloadIdentity.Metadata.Project

	deployedAksInfraWorkloadIdentity, err := getAksWorkloadIdentity(ctx, wiName, wiClusterName, wiProjectName)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			log.Printf("workload identity %s not found, removing from state", wiName)
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	err = flattenAksWorkloadIdentity(d, deployedAksInfraWorkloadIdentity)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func getAksWorkloadIdentity(ctx context.Context, name, clusterName, project string) (*infrapb.AksWorkloadIdentity, error) {
	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithTimeout(auth.URL, auth.Key, versioninfo.GetUserAgent(), CONN_TIMEOUT, options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return nil, err
	}

	extResponse, err := client.InfraV3().Cluster().ExtApi().GetAksWorkloadIdentity(ctx, options.ExtOptions{
		Name:    clusterName,
		Project: project,
		UrlParams: map[string]string{
			"identity_name": name,
		},
	})
	if err != nil {
		return nil, err
	}

	var deployedAksInfraWorkloadIdentity infrapb.AksWorkloadIdentity
	if err = json.Unmarshal(extResponse.Body, &deployedAksInfraWorkloadIdentity); err != nil {
		return nil, err
	}

	log.Println("deployedAksInfraWorkloadIdentity from controller", spew.Sprintf("%+v", &deployedAksInfraWorkloadIdentity))

	return &deployedAksInfraWorkloadIdentity, nil

}

func listAksWorkloadIdentity(ctx context.Context, clusterName, project string) (*infrapb.AksWorkloadIdentityList, error) {
	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithTimeout(auth.URL, auth.Key, versioninfo.GetUserAgent(), CONN_TIMEOUT, options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return nil, err
	}

	extResponse, err := client.InfraV3().Cluster().ExtApi().ListAksWorkloadIdentities(ctx, options.ExtOptions{
		Name:    clusterName,
		Project: project,
	})
	if err != nil {
		return nil, err
	}

	var aksInfraWorkloadIdentityList infrapb.AksWorkloadIdentityList
	if err = json.Unmarshal(extResponse.Body, &aksInfraWorkloadIdentityList); err != nil {
		return nil, err
	}

	log.Println("aksInfraWorkloadIdentityList from controller", spew.Sprintf("%+v", &aksInfraWorkloadIdentityList))

	return &aksInfraWorkloadIdentityList, nil

}

func resourceAKSWorkloadIdentityUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Println("update aks workload identity")

	return resourceAKSWorkloadIdentityUpsert(ctx, d)
}

func resourceAKSWorkloadIdentityDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	log.Println("delete ak workload identity")

	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	desiredInfraAksWorkloadIdentity, err := expandAksWorkloadIdentity(d)
	if err != nil {
		return diag.FromErr(err)
	}

	wiName := desiredInfraAksWorkloadIdentity.Spec.Metadata.Name
	wiClusterName := desiredInfraAksWorkloadIdentity.Metadata.Clustername
	wiProjectName := desiredInfraAksWorkloadIdentity.Metadata.Project

	log.Printf("deleting workload identity: %s for edgename: %s and projectname: %s", wiName, wiClusterName, wiProjectName)

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithTimeout(auth.URL, auth.Key, versioninfo.GetUserAgent(), CONN_TIMEOUT, options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = client.InfraV3().Cluster().ExtApi().DeleteAksWorkloadIdentity(ctx, options.ExtOptions{
		Name:    wiClusterName,
		Project: wiProjectName,
		UrlParams: map[string]string{
			"identity_name": wiName,
		},
	})
	if err != nil {
		return diag.FromErr(err)
	}

	ticker := time.NewTicker(time.Duration(60) * time.Second)
	defer ticker.Stop()

LOOP:
	for {
		select {
		case <-ctx.Done():
			log.Printf("workload identity deletion %s timed out", wiName)
			return diag.FromErr(fmt.Errorf("workload identity deletion %s timed out", wiName))

		case <-ticker.C:
			aksWorkloadIdentityList, err := listAksWorkloadIdentity(ctx, wiClusterName, wiProjectName)
			if err != nil {
				return diag.FromErr(err)
			}

			for _, aksWorkloadIdentity := range aksWorkloadIdentityList.Items {
				if aksWorkloadIdentity.Spec.Metadata.Name == wiName {
					log.Printf("workload identity %s deletion in progress", wiName)
					continue LOOP
				}
			}

			log.Printf("workload identity %s deletion complete", wiName)
			break LOOP
		}
	}

	return diags
}

func resourceAKSWorkloadIdentityUpsert(ctx context.Context, d *schema.ResourceData) diag.Diagnostics {
	log.Println("upsert aks workload identity")

	var diags diag.Diagnostics

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

	desiredInfraAksWorkloadIdentity, err := expandAksWorkloadIdentity(d)
	if err != nil {
		log.Println("error in expanding aks workload identity", err)
		return diag.FromErr(err)
	}

	wiName := desiredInfraAksWorkloadIdentity.Spec.Metadata.Name
	wiClusterName := desiredInfraAksWorkloadIdentity.Metadata.Clustername
	wiProjectName := desiredInfraAksWorkloadIdentity.Metadata.Project

	log.Printf("upserting workload identity: %s for edgename: %s and projectname: %s", wiName, wiClusterName, wiProjectName)

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithTimeout(auth.URL, auth.Key, versioninfo.GetUserAgent(), CONN_TIMEOUT, options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return diag.FromErr(err)
	}

	desiredInfraAksWorkloadIdentityBytes, err := json.Marshal(desiredInfraAksWorkloadIdentity)
	if err != nil {
		return diag.FromErr(err)
	}

	response, err := client.InfraV3().Cluster().ExtApi().ApplyAksWorkloadIdentity(ctx, options.ExtOptions{
		Name:    wiClusterName,
		Project: wiProjectName,
		Body:    desiredInfraAksWorkloadIdentityBytes,
	})
	if err != nil {
		log.Println("workload identity apply error", err)
		return diag.FromErr(err)
	}

	var applyAksWorkloadIdentityResponse infrapb.ApplyAksWorkloadIdentityResponse
	if err = json.Unmarshal(response.Body, &applyAksWorkloadIdentityResponse); err != nil {
		return diag.FromErr(err)
	}

	log.Println("applyAksWorkloadIdentityResponse from controller", spew.Sprintf("%+v", &applyAksWorkloadIdentityResponse))

	if applyAksWorkloadIdentityResponse.TasksetId == "0lk5wke" {
		log.Println("Taskset ID is 0, implies no-op. Workload Identity already exists")
		d.SetId(wiName)
		return diags
	}

	ticker := time.NewTicker(time.Duration(5) * time.Second)
	defer ticker.Stop()

LOOP:
	for {
		select {
		case <-ctx.Done():
			log.Printf("workload identity %s operation timed out", wiName)
			return diag.FromErr(fmt.Errorf("workload identity %s operation timed out", wiName))
		case <-ticker.C:
			statusCluster, err := client.InfraV3().Cluster().Status(ctx, options.StatusOptions{
				Name:    wiClusterName,
				Project: wiProjectName,
			})
			if err != nil {
				log.Println("error in getting cluster status", err)
				return diag.FromErr(err)
			}

			if len(statusCluster.Status.LastTasksets) == 0 {
				log.Printf("workload identity %s operation not started", wiName)
				continue
			}

			for _, taskset := range statusCluster.Status.LastTasksets {
				if taskset.TasksetId != applyAksWorkloadIdentityResponse.TasksetId {
					continue
				}
				log.Printf("taskset is %s\n", spew.Sprintf("%+v", taskset))

				tasksetStatus := taskset.TasksetStatus

				switch tasksetStatus {
				case "PROVISION_TASKSET_STATUS_COMPLETE":
					log.Printf("workload identity %s operation completed", wiName)
					break LOOP

				case "PROVISION_TASKSET_STATUS_FAILED":
					log.Printf("workload identity %s operation failed", wiName)
					combinedErrors := ""

					var errorSummaries []string
					for _, task := range taskset.TasksetOperations {
						if task.ErrorSummary != "" {
							errorSummaries = append(errorSummaries, task.ErrorSummary)
						}
					}
					if len(errorSummaries) > 0 {
						combinedErrors += "\nTaskset Errors:\n" + strings.Join(errorSummaries, "\n") + "\n"
					}

					if statusCluster.Status.Aks != nil {
						edgeResourceErrors, err := collectAKSV3UpsertEdgeResourceErrors(desiredInfraAksWorkloadIdentity, statusCluster.Status.Aks.EdgeResources)
						if err != nil {
							return diag.FromErr(err)
						}
						if edgeResourceErrors != "" {
							combinedErrors += "\nEdge Resource Errors:" + edgeResourceErrors + "\n"
						}
					}
					return diag.Errorf("workload identity %s operation failed with errors: %s", wiName, combinedErrors)

				case "PROVISION_TASKSET_STATUS_IN_PROGRESS", "PROVISION_TASKSET_STATUS_PENDING":
					log.Printf("workload identity %s operation", wiName)
				}
			}
		}
	}

	d.SetId(wiName)
	return diags
}

func collectAKSV3UpsertEdgeResourceErrors(desiredInfraAksWorkloadIdentity *infrapb.AksWorkloadIdentity, edgeResources []*infrapb.EdgeResourceStatus) (string, error) {
	raSet := make(map[string]struct{})
	for _, ra := range desiredInfraAksWorkloadIdentity.Spec.RoleAssignments {
		raSet[ra.Name] = struct{}{}
	}

	saSet := make(map[string]struct{})
	for _, sa := range desiredInfraAksWorkloadIdentity.Spec.ServiceAccounts {
		saSet[sa.Metadata.Name] = struct{}{}
	}

	var found bool
	collectedErrors := AksUpsertErrorFormatter{}

	collectedErrors.WorkloadIdentities = []AksWorkloadIdentityErrorFormatter{}
	for _, er := range edgeResources {
		if er.EdgeResourceType != "AksWorkloadIdentity" && er.Name != desiredInfraAksWorkloadIdentity.Spec.Metadata.Name {
			continue
		}
		found = true

		if strings.Contains(er.ProvisionStatus, "FAILED") {
			collectedErrors.WorkloadIdentities = append(collectedErrors.WorkloadIdentities, AksWorkloadIdentityErrorFormatter{
				Name:          er.Name,
				FailureReason: er.ProvisionStatusReason,
			})
		}

		for _, ra := range er.AksWorkloadIdentityStatus.RoleAssignmentsStatus {
			if _, found := raSet[ra.Name]; !found {
				continue
			}
			if strings.Contains(ra.ProvisionStatus, "FAILED") {
				collectedErrors.WorkloadIdentities = append(collectedErrors.WorkloadIdentities, AksWorkloadIdentityErrorFormatter{
					Name:          er.Name,
					FailureReason: ra.ProvisionStatusReason,
				})
			}
		}

		for _, sa := range er.AksWorkloadIdentityStatus.ServiceAccountsStatus {
			if _, found := saSet[sa.Name]; !found {
				continue
			}
			if strings.Contains(sa.ProvisionStatus, "FAILED") {
				collectedErrors.WorkloadIdentities = append(collectedErrors.WorkloadIdentities, AksWorkloadIdentityErrorFormatter{
					Name:          er.Name,
					FailureReason: sa.ProvisionStatusReason,
				})
			}
		}
	}

	if !found {
		collectedErrors.WorkloadIdentities = append(collectedErrors.WorkloadIdentities, AksWorkloadIdentityErrorFormatter{
			Name:          desiredInfraAksWorkloadIdentity.Spec.Metadata.Name,
			FailureReason: "workload identity not found in the cluster",
		})
	}

	if len(collectedErrors.WorkloadIdentities) == 0 {
		return "", nil
	}

	collectedErrsFormattedBytes, err := json.MarshalIndent(collectedErrors, "", "    ")
	if err != nil {
		return "", err
	}
	collectErrs := strings.ReplaceAll(string(collectedErrsFormattedBytes), "\\n", "\n")

	fmt.Println("after MarshalIndent: ", "collectErrs", collectErrs)
	return "\n" + collectErrs, nil
}

func expandAksWorkloadIdentity(in *schema.ResourceData) (*infrapb.AksWorkloadIdentity, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "expand empty aks workload identity input")
	}
	obj := &infrapb.AksWorkloadIdentity{}

	if v, ok := in.Get("metadata").([]interface{}); ok && len(v) > 0 {
		obj.Metadata = expandAksWorkloadIdentityMetadata(v)
	}

	if v, ok := in.Get("spec").([]interface{}); ok && len(v) > 0 {
		obj.Spec = expandAksWorkloadIdentitySpec(v)
	}

	log.Println("desiredInfraAksWorkloadIdentity from expandAksWorkloadIdentity", spew.Sprintf("%+v", obj))

	return obj, nil

}

func expandAksWorkloadIdentityMetadata(p []interface{}) *infrapb.AksWorkloadIdentityMetadata {
	obj := infrapb.AksWorkloadIdentityMetadata{}

	if len(p) == 0 || p[0] == nil {
		return &obj
	}

	m := p[0].(map[string]interface{})

	if v, ok := m["name"].(string); ok && v != "" {
		obj.Name = v
	}

	if v, ok := m["project"].(string); ok && v != "" {
		obj.Project = v
	}

	if v, ok := m["cluster_name"].(string); ok && v != "" {
		obj.Clustername = v
	}

	return &obj
}

func expandAksWorkloadIdentitySpec(p []interface{}) *infrapb.AksWorkloadIdentitySpec {
	obj := &infrapb.AksWorkloadIdentitySpec{}

	if len(p) == 0 || p[0] == nil {
		return nil
	}

	m := p[0].(map[string]interface{})

	if v, ok := m["create_identity"].(bool); ok {
		obj.CreateIdentity = v
	}

	if v, ok := m["metadata"].([]interface{}); ok && len(v) > 0 {
		obj.Metadata = expandAzureWorkloadIdentityMetadata(v)
	}

	if v, ok := m["role_assignments"].([]interface{}); ok && len(v) > 0 {
		obj.RoleAssignments = expandRoleAssignments(v)
	}

	if v, ok := m["service_accounts"].([]interface{}); ok && len(v) > 0 {
		obj.ServiceAccounts = expandServiceAccounts(v)
	}

	return obj
}

func expandAzureWorkloadIdentityMetadata(p []interface{}) *infrapb.AzureWorkloadIdentityMetadata {
	obj := infrapb.AzureWorkloadIdentityMetadata{}

	if len(p) == 0 || p[0] == nil {
		return &obj
	}

	m := p[0].(map[string]interface{})

	if v, ok := m["client_id"].(string); ok && v != "" {
		obj.ClientId = v
	}

	if v, ok := m["principal_id"].(string); ok && v != "" {
		obj.PrincipalId = v
	}

	if v, ok := m["name"].(string); ok && v != "" {
		obj.Name = v
	}

	if v, ok := m["location"].(string); ok && v != "" {
		obj.Location = v
	}

	if v, ok := m["resource_group"].(string); ok && v != "" {
		obj.ResourceGroup = v
	}

	if v, ok := m["tags"].(map[string]interface{}); ok {
		obj.Tags = expandTags(v)
	}

	return &obj
}

func expandRoleAssignments(p []interface{}) []*infrapb.AzureWorkloadIdentityRoleAssignment {
	var roleAssignments []*infrapb.AzureWorkloadIdentityRoleAssignment

	for _, item := range p {
		m := item.(map[string]interface{})

		roleAssignment := &infrapb.AzureWorkloadIdentityRoleAssignment{}

		if v, ok := m["name"].(string); ok && v != "" {
			roleAssignment.Name = v
		}

		if v, ok := m["role_definition_id"].(string); ok && v != "" {
			roleAssignment.RoleDefinitionId = v
		}

		if v, ok := m["scope"].(string); ok && v != "" {
			roleAssignment.Scope = v
		}

		roleAssignments = append(roleAssignments, roleAssignment)
	}

	return roleAssignments
}

func expandServiceAccounts(p []interface{}) []*infrapb.AzureWorkloadIdentityK8SServiceAccount {
	var serviceAccounts []*infrapb.AzureWorkloadIdentityK8SServiceAccount

	for _, item := range p {
		m := item.(map[string]interface{})

		serviceAccount := &infrapb.AzureWorkloadIdentityK8SServiceAccount{}

		if v, ok := m["metadata"].([]interface{}); ok && len(v) > 0 {
			serviceAccount.Metadata = expandServiceAccountMetadata(v)
		}

		if v, ok := m["create_account"].(bool); ok {
			serviceAccount.CreateAccount = v
		}

		serviceAccounts = append(serviceAccounts, serviceAccount)
	}

	return serviceAccounts
}

func expandServiceAccountMetadata(p []interface{}) *infrapb.K8SServiceAccountMetadata {
	obj := infrapb.K8SServiceAccountMetadata{}

	if len(p) == 0 || p[0] == nil {
		return &obj
	}

	m := p[0].(map[string]interface{})

	if v, ok := m["name"].(string); ok && v != "" {
		obj.Name = v
	}

	if v, ok := m["namespace"].(string); ok && v != "" {
		obj.Namespace = v
	}

	if v, ok := m["annotations"].(map[string]interface{}); ok {
		obj.Annotations = expandAnnotations(v)
	}

	if v, ok := m["labels"].(map[string]interface{}); ok {
		obj.Labels = expandLabels(v)
	}

	return &obj
}

func expandTags(m map[string]interface{}) map[string]string {
	tags := make(map[string]string)

	for k, v := range m {
		if s, ok := v.(string); ok && s != "" {
			tags[k] = s
		}
	}

	return tags
}

func expandAnnotations(m map[string]interface{}) map[string]string {
	annotations := make(map[string]string)

	for k, v := range m {
		if s, ok := v.(string); ok && s != "" {
			annotations[k] = s
		}
	}

	return annotations
}

func expandLabels(m map[string]interface{}) map[string]string {
	labels := make(map[string]string)

	for k, v := range m {
		if s, ok := v.(string); ok && s != "" {
			labels[k] = s
		}
	}

	return labels
}

func flattenAksWorkloadIdentity(d *schema.ResourceData, in *infrapb.AksWorkloadIdentity) error {
	if in == nil {
		return nil
	}

	obj := map[string]interface{}{}

	if len(in.ApiVersion) > 0 {
		obj["api_version"] = in.ApiVersion
	}
	if len(in.Kind) > 0 {
		obj["kind"] = in.Kind
	}

	var metadata []interface{}
	if in.Metadata != nil {
		metadata = flattenMetadata(in.Metadata)
	}

	log.Println("metadata from flattenAksWorkloadIdentity", spew.Sprintf("%+v", metadata))

	if err := d.Set("metadata", metadata); err != nil {
		return err
	}

	var spec []interface{}
	if in.Spec != nil {
		spec = flattenSpec(in.Spec)
	}

	log.Println("spec from flattenAksWorkloadIdentity", spew.Sprintf("%+v", spec))

	if err := d.Set("spec", spec); err != nil {
		return err
	}

	return nil
}

func flattenMetadata(in *infrapb.AksWorkloadIdentityMetadata) []interface{} {
	if in == nil {
		return nil
	}

	obj := make(map[string]interface{})

	if len(in.Clustername) > 0 {
		obj["cluster_name"] = in.Clustername
	}
	if len(in.Project) > 0 {
		obj["project"] = in.Project
	}
	// if len(in.Name) > 0 {
	// 	obj["name"] = in.Name
	// }

	return []interface{}{obj}
}

func flattenSpec(in *infrapb.AksWorkloadIdentitySpec) []interface{} {
	if in == nil {
		return nil
	}

	obj := make(map[string]interface{})

	obj["create_identity"] = in.CreateIdentity

	var metadata []interface{}
	if in.Metadata != nil {
		metadata = flattenSpecMetadata(in.Metadata)
	}
	obj["metadata"] = metadata

	var roleAssignments []interface{}
	for _, ra := range in.RoleAssignments {
		roleAssignments = append(roleAssignments, flattenRoleAssignment(ra)...)
	}
	obj["role_assignments"] = roleAssignments

	var serviceAccounts []interface{}
	for _, sa := range in.ServiceAccounts {
		serviceAccounts = append(serviceAccounts, flattenServiceAccount(sa)...)
	}
	obj["service_accounts"] = serviceAccounts

	return []interface{}{obj}
}

func flattenSpecMetadata(in *infrapb.AzureWorkloadIdentityMetadata) []interface{} {
	if in == nil {
		return nil
	}

	obj := make(map[string]interface{})

	if len(in.Name) > 0 {
		obj["name"] = in.Name
	}
	if len(in.Location) > 0 {
		obj["location"] = in.Location
	}
	if len(in.ResourceGroup) > 0 {
		obj["resource_group"] = in.ResourceGroup
	}

	var tags map[string]interface{}
	if in.Tags != nil {
		tags = flattenStringMap(in.Tags)
	}
	obj["tags"] = tags

	return []interface{}{obj}
}

func flattenRoleAssignment(in *infrapb.AzureWorkloadIdentityRoleAssignment) []interface{} {
	if in == nil {
		return nil
	}

	obj := make(map[string]interface{})

	if len(in.Name) > 0 {
		obj["name"] = in.Name
	}
	if len(in.Scope) > 0 {
		obj["scope"] = in.Scope
	}

	return []interface{}{obj}
}

func flattenServiceAccount(in *infrapb.AzureWorkloadIdentityK8SServiceAccount) []interface{} {
	if in == nil {
		return nil
	}

	obj := make(map[string]interface{})

	obj["create_account"] = in.CreateAccount

	if in.Metadata != nil {
		obj["metadata"] = flattenServiceAccountMetadata(in.Metadata)
	}

	return []interface{}{obj}
}

func flattenServiceAccountMetadata(in *infrapb.K8SServiceAccountMetadata) []interface{} {
	if in == nil {
		return nil
	}

	obj := make(map[string]interface{})

	if len(in.Name) > 0 {
		obj["name"] = in.Name
	}
	if len(in.Namespace) > 0 {
		obj["namespace"] = in.Namespace
	}

	if in.Annotations != nil {
		obj["annotations"] = flattenStringMap(in.Annotations)
	}

	if in.Labels != nil {
		obj["labels"] = flattenStringMap(in.Labels)
	}

	return []interface{}{obj}
}

func flattenStringMap(in map[string]string) map[string]interface{} {
	if in == nil {
		return nil
	}

	out := make(map[string]interface{})
	for k, v := range in {
		out[k] = v
	}

	return out
}
