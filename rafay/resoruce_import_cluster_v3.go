package rafay

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"

	dynamic "github.com/RafaySystems/rafay-common/pkg/hub/client/dynamic"
	"github.com/RafaySystems/rafay-common/pkg/hub/client/options"
	typed "github.com/RafaySystems/rafay-common/pkg/hub/client/typed"
	"github.com/RafaySystems/rafay-common/pkg/hub/terraform/resource"
	"github.com/RafaySystems/rafay-common/proto/types/hub/commonpb"
	"github.com/RafaySystems/rafay-common/proto/types/hub/infrapb"
	"github.com/RafaySystems/rctl/pkg/config"
	"github.com/RafaySystems/rctl/pkg/versioninfo"
	"github.com/davecgh/go-spew/spew"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	schema "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
)

func resourceImportClusterV3() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceImportClusterV3Create,
		ReadContext:   resourceImportClusterV3Read,
		UpdateContext: resourceImportClusterV3Update,
		DeleteContext: resourceImportClusterV3Delete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(100 * time.Minute),
			Update: schema.DefaultTimeout(130 * time.Minute),
			Delete: schema.DefaultTimeout(70 * time.Minute),
		},

		SchemaVersion: 1,
		Schema:        getImportedV3ResourceSchema(),
	}
}

func getImportedV3ResourceSchema() map[string]*schema.Schema {
	clusterSchema := resource.ClusterSchema.Schema
	additionalSchema := map[string]*schema.Schema{
		"bootstrap_path": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Specify bootstrap file path to store rafay bootstrap file content.",
		},
		"kubeconfig_path": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Specify kubeconfig file path.",
		},
	}
	for k, v := range additionalSchema {
		if _, ok := clusterSchema[k]; !ok {
			clusterSchema[k] = v
		}
	}
	return clusterSchema
}

func resourceImportClusterV3Create(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("Resource Import cluster creation starts...")

	diags := resourceImportClusterV3Upsert(ctx, d, m)
	return diags
}

func resourceImportClusterV3Read(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	log.Println("Resource Import cluster reading starts...")
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}
	projectName, ok := d.Get("metadata.0.project").(string)
	if !ok || projectName == "" {
		return diag.FromErr(errors.New("project name unable to be found"))
	}

	resourceName, ok := d.Get("metadata.0.name").(string)
	if !ok || resourceName == "" {
		return diag.FromErr(errors.New("resource name unable to be found"))
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent(), options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return diag.FromErr(err)
	}

	resourceRemoteData, err := client.InfraV3().Cluster().Get(ctx, options.GetOptions{
		Name:    resourceName,
		Project: projectName,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	err = flattenImportClusterV3(d, resourceRemoteData)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Println("Resource Import cluster reading finished...")

	return diags
}

func resourceImportClusterV3Update(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("Resource Import cluster update starts...")
	return resourceImportClusterV3Upsert(ctx, d, m)
}

func resourceImportClusterV3Delete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("Resource Import cluster delete starts...")

	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	projectName, ok := d.Get("metadata.0.project").(string)
	if !ok || projectName == "" {
		return diag.FromErr(errors.New("project name unable to be found"))
	}

	resourceName, ok := d.Get("metadata.0.name").(string)
	if !ok || resourceName == "" {
		return diag.FromErr(errors.New("resource name unable to be found"))
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent(), options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.InfraV3().Cluster().Delete(ctx, options.DeleteOptions{
		Name:    resourceName,
		Project: projectName,
	})
	if err != nil {
		log.Printf("cluster delete failed for edgename: %s and projectname: %s", resourceName, projectName)
		return diag.FromErr(err)
	}

	ticker := time.NewTicker(time.Duration(60) * time.Second)
	defer ticker.Stop()

LOOP:
	for {
		select {
		case <-ctx.Done():
			log.Printf("Cluster Deletion for edgename: %s and projectname: %s got timeout out.", resourceName, projectName)
			return diag.FromErr(fmt.Errorf("cluster deletion for edgename: %s and projectname: %s got timeout out", resourceName, projectName))
		case <-ticker.C:
			_, err := client.InfraV3().Cluster().Get(ctx, options.GetOptions{
				Name:    resourceName,
				Project: projectName,
			})
			if dErr, ok := err.(*dynamic.DynamicClientGetError); ok && dErr != nil {
				switch dErr.StatusCode {
				case http.StatusNotFound:
					log.Printf("Cluster Deletion completes for edgename: %s and projectname: %s", resourceName, projectName)
					break LOOP
				default:
					log.Printf("Cluster Deletion failed for edgename: %s and projectname: %s with error: %s", resourceName, projectName, dErr.Error())
					return diag.FromErr(dErr)
				}
			}
			log.Printf("Cluster Deletion is in progress for edgename: %s and projectname: %s", resourceName, projectName)
		}
	}

	return diags
}

func resourceImportClusterV3Upsert(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("Cluster upsert starts")
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	rawConfig := d.GetRawConfig()
	if d.State() != nil && d.State().ID != "" {
		n := GetMetaName(d)
		if n != "" && n != d.State().ID {
			log.Printf("metadata name change not supported")
			d.State().Tainted = true
			return diag.FromErr(fmt.Errorf("%s", "metadata name change not supported"))
		}
	}
	// rawConfig := d.GetRawConfig()

	cluster, err := expandImportClusterV3(d)
	if err != nil {
		log.Printf("Cluster expandCluster error")
		return diag.FromErr(err)
	}

	resourceName := cluster.Metadata.Name
	projectName := cluster.Metadata.Project
	log.Println(">>>>>> CLUSTER: ", cluster)

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent(), options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return diag.FromErr(err)
	}

	resourceAlreadyExists := false
	resourceId, projectId := "", ""
	resourceRemoteData, _ := client.InfraV3().Cluster().Status(ctx, options.StatusOptions{
		Name:    resourceName,
		Project: projectName,
	})
	if resourceRemoteData != nil {
		resourceAlreadyExists = true
		resourceId = resourceRemoteData.Status.Id
		projectId, err = getProjectIDFromName(projectName)
		if err != nil {
			log.Print("error converting project name to id")
			return diag.Errorf("error converting project name to project ID")
		}
	}

	err = client.InfraV3().Cluster().Apply(ctx, cluster, options.ApplyOptions{})
	if err != nil {
		// XXX Debug
		n1 := spew.Sprintf("%+v", cluster)
		log.Println("Cluster apply cluster:", n1)
		log.Printf("Cluster apply error")
		return diag.FromErr(err)
	}

	applyBootstrap := true
	bootstrapFilePath := ""
	if !resourceAlreadyExists || isBlueprintSyncPending(resourceId, projectId) {
		// Fetch Bootstrap from remote
		getBootstrapContentResp, err := client.InfraV3().Cluster().ExtApi().Bootstrap(
			ctx, options.ExtOptions{
				Name:    resourceName,
				Project: projectName,
			},
		)
		if err != nil {
			log.Println("Failed to fetch Bootstrap content from ExtApi Client.")
			return diag.FromErr(fmt.Errorf("failed to fetch Bootstrap content for resource: %s in project: %s", cluster.Metadata.Name, cluster.Metadata.Project))
		}

		// encodedBootstrapContent := string(getBootstrapContentResp.Body)
		// log.Println("Debug--- ", "getBootstrapContentResp", getBootstrapContentResp)
		// log.Println("Debug--- ", "encodedBootstrapContent", encodedBootstrapContent)
		// bootstrapContentBytes, err := base64.StdEncoding.DecodeString(encodedBootstrapContent)
		// if err != nil {
		// 	log.Println("error decoding bootstrap content.")
		// 	return diag.FromErr(fmt.Errorf("internal error: Invalid bootstrap content found for resource: %s in project: %s. Error: %s", cluster.Metadata.Name, cluster.Metadata.Project, err.Error()))
		// }
		bootstrapContent := string(getBootstrapContentResp.Body)
		rawBootstrapPath := rawConfig.GetAttr("bootstrap_path")
		if rawBootstrapPath.IsNull() || len(rawBootstrapPath.AsString()) == 0 {
			bootstrapFilePath = ClusterBootstrapFileName
		} else {
			bootstrapFilePath = rawBootstrapPath.AsString()
		}
		if len(bootstrapContent) > 0 {
			// write bootstrapContent in File specified by bootstrap_filepath
			f, err := os.Create(bootstrapFilePath)
			if err != nil {
				return diag.FromErr(fmt.Errorf("failed to create bootstrap file on specified path: %s", bootstrapFilePath))
			}
			defer f.Close()
			log.Printf("started writing bootstrap content in file: %s", bootstrapFilePath)
			_, err = f.WriteString(bootstrapContent)
			if err != nil {
				log.Printf("Failed to write bootstrap content: %s in file created on specified path: %s", bootstrapContent, bootstrapFilePath)
				return diag.FromErr(fmt.Errorf("failed to write bootstrap content in file created on specified path: %s", bootstrapFilePath))
			}
			log.Printf("finished writing bootstrap content in file: %s", bootstrapFilePath)
		}

		// look for kubeconfig submitted if any and apply bootstrap accordingly
		rawKubeconfigPath := rawConfig.GetAttr("kubeconfig_path")
		if rawKubeconfigPath.IsNull() || len(rawKubeconfigPath.AsString()) == 0 {
			log.Println("kubeconfig_path not set for resource to apply bootstrap. Kindly specify kubeconfig_path and retrigger apply to apply bootstrap.")
			applyBootstrap = false
		} else {
			kubeConfigFilePath := rawKubeconfigPath.AsString()
			cmd := exec.Command("kubectl", "--kubeconfig", kubeConfigFilePath, "apply", "-f", bootstrapFilePath)
			b, err := cmd.CombinedOutput()
			if err != nil {
				log.Println("kubectl command failed to apply bootstrap yaml file", string(b))
				return diag.FromErr(fmt.Errorf("kubectl command failed to apply bootstrap yaml file with error: %s", err.Error()))
			}
		}
	}
	d.SetId(cluster.Metadata.Name)
	// wait for cluster creation
	ticker := time.NewTicker(time.Duration(60) * time.Second)
	defer ticker.Stop()

LOOP:
	for {
		select {
		case <-ctx.Done():
			log.Printf("Cluster operation timed out for resource: %s and projectname: %s", resourceName, projectName)
			return diag.FromErr(fmt.Errorf("cluster operation timed out for resource: %s and projectname: %s", resourceName, projectName))
		case <-ticker.C:
			if !applyBootstrap {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Warning,
					Summary:  "Unable to apply bootstrap as no kubeconfig_path found in resource configuration. Kindly refer to " + bootstrapFilePath + "for generated bootstrap.",
				})
				break LOOP
			}
			resourceRemoteData, err = client.InfraV3().Cluster().Status(ctx, options.StatusOptions{
				Name:    resourceName,
				Project: projectName,
			})
			if err != nil {
				log.Printf("Fetching cluster having resource: %s and projectname: %s failing due to err: %v", resourceName, projectName, err)
				return diag.FromErr(err)
			} else if resourceRemoteData == nil {
				log.Printf("Cluster operation has not started with resource: %s and projectname: %s", resourceName, projectName)
			} else if resourceRemoteData.Status != nil && resourceRemoteData.Status.Imported != nil && resourceRemoteData.Status.CommonStatus != nil {
				resourceId := resourceRemoteData.Status.Id
				projectId, err := getProjectIDFromName(projectName)
				if err != nil {
					log.Print("error converting project name to id")
					return diag.Errorf("error converting project name to project ID")
				}
				resourceCommonStatus := resourceRemoteData.Status.CommonStatus
				switch resourceCommonStatus.ConditionStatus {
				case commonpb.ConditionStatus_StatusSubmitted:
					log.Printf("Cluster operation not completed for resource: %s and projectname: %s. Waiting 60 seconds more for cluster to complete the operation.", resourceName, projectName)
				case commonpb.ConditionStatus_StatusOK:
					log.Println("Checking in cluster conditions for blueprint sync success..")
					conditionsFailure, clusterReadiness, err := getClusterConditions(resourceId, projectId)
					if err != nil {
						log.Printf("error while getCluster %s", err.Error())
						return diag.FromErr(err)
					}
					if conditionsFailure {
						log.Printf("blueprint sync failed for resource: %s and projectname: %s", resourceName, projectName)
						return diag.FromErr(fmt.Errorf("blueprint sync failed for resource: %s and projectname: %s", resourceName, projectName))
					} else if clusterReadiness {
						log.Printf("Cluster operation completed for resource: %s and projectname: %s", resourceName, projectName)
						break LOOP
					} else {
						log.Println("Cluster Provisiong is Complete. Waiting for cluster to be Ready...")
					}
				case commonpb.ConditionStatus_StatusFailed:
					// log.Printf("Cluster operation failed for edgename: %s and projectname: %s with failure reason: %s", edgeName, projectName, uClusterCommonStatus.Reason)
					return diag.Errorf("Cluster operation failed for resource: %s and projectname: %s with failure reasons: %s", resourceName, projectName, resourceRemoteData.Status.ProvisionStatusReason)
				}
			}
		}
	}
	return diags
}

func expandImportClusterV3(in *schema.ResourceData) (*infrapb.Cluster, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "resource config found empty")
	}
	// rawConfig = rawConfig.AsValueSlice()[0]
	obj := &infrapb.Cluster{}

	if v, ok := in.Get("metadata").([]interface{}); ok && len(v) > 0 {
		obj.Metadata = expandMetaData(v)
	}

	if v, ok := in.Get("spec").([]interface{}); ok && len(v) > 0 {
		objSpec, err := expandImportClusterV3Spec(v)
		if err != nil {
			return nil, err
		}
		log.Println("expandImportClusterV3Spec got spec")
		obj.Spec = objSpec
	}

	obj.ApiVersion = "infra.k8smgmt.io/v3"
	obj.Kind = "Cluster"

	return obj, nil
}

func expandImportClusterV3Spec(p []interface{}) (*infrapb.ClusterSpec, error) {
	obj := &infrapb.ClusterSpec{}
	if len(p) == 0 || p[0] == nil {
		return obj, fmt.Errorf("%s", "resource config spec found empty")
	}
	// rawConfig = rawConfig.AsValueSlice()[0]
	in := p[0].(map[string]interface{})

	if v, ok := in["type"].(string); ok && len(v) > 0 {
		obj.Type = v
	}

	if obj.Type != ImportedClusterType {
		return nil, fmt.Errorf("invalid cluster type found. Please use type: %s with this resource", ImportedClusterType)
	} else if v, ok := in["config"].([]interface{}); ok && len(v) > 0 {
		obj.Config = expandImportClusterV3Config(v)
	}

	if v, ok := in["sharing"].([]interface{}); ok && len(v) > 0 {
		obj.Sharing = expandSharingSpecV3(v)
	}

	if v, ok := in["blueprint_config"].([]interface{}); ok && len(v) > 0 {
		obj.BlueprintConfig = expandImportClusterV3Blueprint(v)
	}

	if v, ok := in["cloud_credentials"].(string); ok && len(v) > 0 {
		obj.CloudCredentials = v
	}

	if v, ok := in["system_components_placement"].([]interface{}); ok && len(v) > 0 {
		obj.SystemComponentsPlacement = expandV3SystemComponentsPlacement(v)
	}

	if v, ok := in["proxy_config"].([]interface{}); ok && len(v) > 0 {
		obj.ProxyConfig = expandV3ProxyConfig(v)
	}
	return obj, nil
}

func expandImportClusterV3Blueprint(p []interface{}) *infrapb.BlueprintConfig {
	obj := infrapb.BlueprintConfig{}
	if len(p) == 0 || p[0] == nil {
		return &obj
	}

	in := p[0].(map[string]interface{})
	if v, ok := in["name"].(string); ok {
		obj.Name = v
	}

	if v, ok := in["version"].(string); ok {
		obj.Version = v
	}

	log.Println("expandClusterV3Blueprint obj", obj)
	return &obj
}

func expandImportClusterV3Config(p []interface{}) *infrapb.ClusterSpec_Imported {
	obj := &infrapb.ClusterSpec_Imported{Imported: &infrapb.ImportedV3ConfigObject{}}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["provision_environment"].(string); ok && len(v) > 0 {
		obj.Imported.ProvisionEnvironment = v
	}

	if v, ok := in["kubernetes_provider"].(string); ok && len(v) > 0 {
		obj.Imported.KubernetesProvider = v
	}

	if v, ok := in["imported_cluster_location"].(string); ok && len(v) > 0 {
		obj.Imported.ImportedClusterLocation = v
	}
	return obj
}

func flattenImportClusterV3(d *schema.ResourceData, in *infrapb.Cluster) error {
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
	var err error

	var ret1 []interface{}
	if in.Metadata != nil {
		v, ok := obj["metadata"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		ret1 = flattenMetadataV3(in.Metadata, v)
	}

	err = d.Set("metadata", ret1)
	if err != nil {
		return err
	}

	var ret2 []interface{}
	if in.Spec != nil {
		v, ok := obj["spec"].([]interface{})
		if !ok {
			v = []interface{}{}
		}

		ret2 = flattenImportClusterV3Spec(in.Spec, v)
	}

	err = d.Set("spec", ret2)
	if err != nil {
		return err
	}

	return nil

}

func flattenImportClusterV3Spec(in *infrapb.ClusterSpec, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.Type) > 0 {
		obj["type"] = in.Type
	}

	if in.BlueprintConfig != nil {
		obj["blueprint_config"] = flattenClusterV3Blueprint(in.BlueprintConfig)
	}

	if len(in.CloudCredentials) > 0 {
		obj["cloud_credentials"] = in.CloudCredentials
	}

	if in.GetImported() != nil {
		v, ok := obj["config"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["config"] = flattenImportClusterV3Config(in.GetImported(), v)
	}

	if in.Sharing != nil {
		obj["sharing"] = flattenSharingSpecV3(in.Sharing)
	}

	if in.SystemComponentsPlacement != nil {
		v, ok := obj["system_components_placement"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["system_components_placement"] = flattenV3SystemComponentsPlacement(in.SystemComponentsPlacement, v)
	}

	if in.ProxyConfig != nil {
		v, ok := obj["proxy_config"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["proxy_config"] = flattenV3ProxyConfig(in.ProxyConfig, v)
	}

	return []interface{}{obj}
}

func flattenImportClusterV3Config(in *infrapb.ImportedV3ConfigObject, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.KubernetesProvider) > 0 {
		obj["kubernetes_provider"] = in.KubernetesProvider
	}

	if len(in.ProvisionEnvironment) > 0 {
		obj["provision_environment"] = in.ProvisionEnvironment
	}

	if len(in.ImportedClusterLocation) > 0 {
		obj["imported_cluster_location"] = in.ImportedClusterLocation
	}

	return []interface{}{obj}
}
