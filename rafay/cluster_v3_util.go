package rafay

import (
	"log"

	"github.com/RafaySystems/rafay-common/proto/types/hub/commonpb"
	"github.com/RafaySystems/rafay-common/proto/types/hub/infrapb"
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

	log.Println("expandClusterV3Blueprint obj", obj)
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
