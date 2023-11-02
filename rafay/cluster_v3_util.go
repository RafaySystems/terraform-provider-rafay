package rafay

import (
	"github.com/RafaySystems/rafay-common/proto/types/hub/commonpb"
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
