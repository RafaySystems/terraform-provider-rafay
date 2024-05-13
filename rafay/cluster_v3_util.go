package rafay

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/RafaySystems/rafay-common/pkg/hub/client/options"
	typed "github.com/RafaySystems/rafay-common/pkg/hub/client/typed"
	"github.com/RafaySystems/rafay-common/proto/types/hub/commonpb"
	"github.com/RafaySystems/rafay-common/proto/types/hub/infrapb"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
)

const (
	V3_CLUSTER_APIVERSION = "infra.k8smgmt.io/v3"
	V3_CLUSTER_KIND       = "Cluster"

	GKE_CLUSTER_TYPE = "gke"

	GKE_ZONAL_CLUSTER_TYPE    = "zonal"
	GKE_REGIONAL_CLUSTER_TYPE = "regional"

	GKE_PRIVATE_CLUSTER_TYPE = "private"
	GKE_PUBLIC_CLUSTER_TYPE  = "public"

	GKE_NODEPOOL_UPGRADE_STRATEGY_SURGE      = "SURGE"
	GKE_NODEPOOL_UPGRADE_STRATEGY_BLUE_GREEN = "BLUE_GREEN"

	GKE_ACCELERATOR_GOOGLE_DRIVER_INSTALATION = "google-managed"
	GKE_ACCELERATOR_USER_DRIVER_INSTALATION   = "user-managed"
)

type AksNodepoolsErrorFormatter struct {
	Name          string `json:"name,omitempty"`
	FailureReason string `json:"failureReason,omitempty"`
}

type AksUpsertErrorFormatter struct {
	FailureReason string                       `json:"failureReason,omitempty"`
	Nodepools     []AksNodepoolsErrorFormatter `json:"nodepools,omitempty"`
}

func flattenMetadataV3(in *commonpb.Metadata, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}

	if len(in.Name) > 0 {
		obj["name"] = in.Name
	}

	if len(in.Project) > 0 {
		obj["project"] = in.Project
	}

	if in.Labels != nil && len(in.Labels) > 0 {
		obj["labels"] = toMapInterface(in.Labels)
	}

	return []interface{}{obj}
}

func expandV3SystemComponentsPlacement(p []interface{}) *infrapb.SystemComponentsPlacement {
	obj := infrapb.SystemComponentsPlacement{}
	if len(p) == 0 || p[0] == nil {
		return &obj
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["node_selector"].(map[string]interface{}); ok && len(v) > 0 {
		obj.NodeSelector = toMapString(v)
	} else {
		obj.NodeSelector = nil
	}
	if v, ok := in["tolerations"].([]interface{}); ok {
		obj.Tolerations = expandV3Tolerations(v)
	}

	if v, ok := in["daemon_set_override"].([]interface{}); ok {
		obj.DaemonSetOverride = expandV3DaemonsetOverride(v)
	}

	return &obj
}

func expandV3Tolerations(p []interface{}) []*v1.Toleration {
	out := make([]*v1.Toleration, len(p))
	if len(p) == 0 || p[0] == nil {
		return out
	}
	for i := range p {
		obj := v1.Toleration{}
		in := p[i].(map[string]interface{})

		if v, ok := in["key"].(string); ok && len(v) > 0 {
			obj.Key = v
		}
		if v, ok := in["operator"].(string); ok && len(v) > 0 {
			obj.Operator = v1.TolerationOperator(v)
		}
		if v, ok := in["value"].(string); ok && len(v) > 0 {
			obj.Value = v
		}
		if v, ok := in["effect"].(string); ok && len(v) > 0 {
			obj.Effect = v1.TaintEffect(v)
		}
		if v, ok := in["toleration_seconds"].(int); ok {
			if v == 0 {
				obj.TolerationSeconds = nil
			} else {
				ts := int64(v)
				log.Println("setting toleration seconds")
				obj.TolerationSeconds = &ts
			}
		}
		out[i] = &obj
	}
	return out
}

func expandV3DaemonsetOverride(p []interface{}) *infrapb.DaemonSetOverride {
	obj := infrapb.DaemonSetOverride{}
	if len(p) == 0 || p[0] == nil {
		return &obj
	}

	in := p[0].(map[string]interface{})
	if v, ok := in["node_selection_enabled"].(bool); ok {
		obj.NodeSelectionEnabled = v
	}
	if v, ok := in["tolerations"].([]interface{}); ok {
		obj.Tolerations = expandV3Tolerations(v)
	}
	return &obj
}

func flattenV3SystemComponentsPlacement(in *infrapb.SystemComponentsPlacement, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if in.NodeSelector != nil && len(in.NodeSelector) > 0 {
		obj["node_selector"] = toMapInterface(in.NodeSelector)
	}

	if in.Tolerations != nil {
		v, ok := obj["tolerations"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["tolerations"] = flattenV3Tolerations(in.Tolerations, v)
	}

	if in.DaemonSetOverride != nil {
		v, ok := obj["daemon_set_override"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["daemon_set_override"] = flattenV3DaemonSetOverride(in.DaemonSetOverride, v)
	}

	return []interface{}{obj}
}

func flattenV3DaemonSetOverride(in *infrapb.DaemonSetOverride, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	obj["node_selection_enabled"] = in.NodeSelectionEnabled

	if in.Tolerations != nil {
		v, ok := obj["tolerations"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["tolerations"] = flattenV3Tolerations(in.Tolerations, v)
	}
	return []interface{}{obj}
}

func flattenV3Tolerations(in []*v1.Toleration, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	out := make([]interface{}, len(in))

	for i, t := range in {
		obj := map[string]interface{}{}
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}
		if len(t.Key) > 0 {
			obj["key"] = t.Key
		}
		if len(t.Operator) > 0 {
			obj["operator"] = t.Operator
		}
		if len(t.Value) > 0 {
			obj["value"] = t.Value
		}
		if len(t.Effect) > 0 {
			obj["effect"] = t.Effect
		}
		if t.TolerationSeconds != nil {
			obj["toleration_seconds"] = t.TolerationSeconds
		}

		out[i] = &obj
	}

	return out
}

func expandV3ProxyConfig(p []interface{}) *infrapb.ProxyConfig {
	obj := infrapb.ProxyConfig{}
	if len(p) == 0 || p[0] == nil {
		return &obj
	}
	// rawConfig = rawConfig.AsValueSlice()[0]
	in := p[0].(map[string]interface{})

	// rawAllowInsecureBootstrap := rawConfig.GetAttr("allow_insecure_bootstrap")
	if v, ok := in["allow_insecure_bootstrap"].(bool); ok {
		obj.AllowInsecureBootstrap = v
	}
	// rawEnabled := rawConfig.GetAttr("enabled")
	if v, ok := in["enabled"].(bool); ok {
		obj.Enabled = v
	}
	if v, ok := in["bootstrap_ca"].(string); ok {
		obj.BootstrapCA = v
	}
	if v, ok := in["http_proxy"].(string); ok {
		obj.HttpProxy = v
	}
	if v, ok := in["https_proxy"].(string); ok {
		obj.HttpsProxy = v
	}
	if v, ok := in["proxy_auth"].(string); ok {
		obj.NoProxy = v
	}
	if v, ok := in["bootstrap_ca"].(string); ok {
		obj.ProxyAuth = v
	}

	return &obj
}

func flattenV3ProxyConfig(in *infrapb.ProxyConfig, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	obj["allow_insecure_bootstrap"] = in.AllowInsecureBootstrap
	obj["enabled"] = in.Enabled

	if len(in.BootstrapCA) > 0 {
		obj["bootstrap_ca"] = in.BootstrapCA
	}
	if len(in.HttpProxy) > 0 {
		obj["http_proxy"] = in.HttpProxy
	}
	if len(in.HttpsProxy) > 0 {
		obj["https_proxy"] = in.HttpsProxy
	}
	if len(in.NoProxy) > 0 {
		obj["no_proxy"] = in.NoProxy
	}
	if len(in.ProxyAuth) > 0 {
		obj["proxy_auth"] = in.ProxyAuth
	}

	return []interface{}{obj}
}

func clusterV3UpsertWaiter(ctx context.Context, client typed.Client, ticker *time.Ticker, resourceName, projectName string) error {
LOOP:
	for {
		select {
		case <-ctx.Done():
			log.Printf("Cluster operation timed out for resource: %s and projectname: %s", resourceName, projectName)
			return fmt.Errorf("cluster operation timed out for resource: %s and projectname: %s", resourceName, projectName)
		case <-ticker.C:
			resourceRemoteData, err := client.InfraV3().Cluster().Status(ctx, options.StatusOptions{
				Name:    resourceName,
				Project: projectName,
			})
			if err != nil {
				log.Printf("Fetching cluster having resource: %s and projectname: %s failing due to err: %v", resourceName, projectName, err)
				return err
			} else if resourceRemoteData == nil {
				log.Printf("Cluster operation has not started with resource: %s and projectname: %s", resourceName, projectName)
			} else if resourceRemoteData.Status != nil && resourceRemoteData.Status.Imported != nil && resourceRemoteData.Status.CommonStatus != nil {
				resourceId := resourceRemoteData.Status.Id
				projectId, err := getProjectIDFromName(projectName)
				if err != nil {
					log.Print("error converting project name to id")
					return errors.New("error converting project name to project ID")
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
						return err
					}
					if conditionsFailure {
						log.Printf("blueprint sync failed for resource: %s and projectname: %s", resourceName, projectName)
						return fmt.Errorf("blueprint sync failed for resource: %s and projectname: %s", resourceName, projectName)
					} else if clusterReadiness {
						log.Printf("Cluster operation completed for resource: %s and projectname: %s", resourceName, projectName)
						break LOOP
					} else {
						log.Println("Cluster Provisiong is Complete. Waiting for cluster to be Ready...")
					}
				case commonpb.ConditionStatus_StatusFailed:
					// log.Printf("Cluster operation failed for edgename: %s and projectname: %s with failure reason: %s", edgeName, projectName, uClusterCommonStatus.Reason)
					return fmt.Errorf("cluster operation failed for resource: %s and projectname: %s with failure reasons: %s", resourceName, projectName, resourceRemoteData.Status.ProvisionStatusReason)
				}
			}
		}
	}
	return nil
}
