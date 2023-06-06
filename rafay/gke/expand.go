package gke

import (
	"fmt"

	"github.com/RafaySystems/rafay-common/proto/types/hub/infrapb"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// takes input given in the format of the terraform schema and populate the backend structure for that resource.
// convert from tf schema --> V3 schema in rafay-common proto


func ExpandClusterV3Config(p []interface{}) *infrapb.ClusterSpec_Gke {
	obj := &infrapb.ClusterSpec_Gke{
		Gke: &infrapb.GkeV3ConfigObject{}}

	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["api_version"].(string); ok && len(v) > 0 {
		//	obj.Gke.ApiVersion = v
	}

	return obj

	// obj := &infrapb.ClusterSpec_Aks{Aks: &infrapb.AksV3ConfigObject{}}
	// 	if len(p) == 0 || p[0] == nil {
	// 		return obj
	// 	}
	// 	in := p[0].(map[string]interface{})

	// 	if v, ok := in["api_version"].(string); ok && len(v) > 0 {
	// 		obj.Aks.ApiVersion = v
	// 	}

	// 	if v, ok := in["kind"].(string); ok && len(v) > 0 {
	// 		obj.Aks.Kind = v
	// 	}

	// 	if v, ok := in["metadata"].([]interface{}); ok && len(v) > 0 {
	// 		obj.Aks.Metadata = expandAKSClusterV3ConfigMetaData(v)
	// 	}

	// 	if v, ok := in["spec"].([]interface{}); ok && len(v) > 0 {
	// 		obj.Aks.Spec = expandAKSClusterV3ConfigSpec(v)
	// 	}

	// return obj
}
