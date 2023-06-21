package rafay

import (
	"github.com/RafaySystems/rafay-common/proto/types/hub/infrapb"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// convert from V3 schema in rafay-common proto --> tf schema
func flattenGKEClusterV3(d *schema.ResourceData, in *infrapb.Cluster) error {
	if in == nil {
		return nil
	}
	// obj := map[string]interface{}{}

	// if len(in.ApiVersion) > 0 {
	// 	obj["api_version"] = in.ApiVersion
	// }
	// if len(in.Kind) > 0 {
	// 	obj["kind"] = in.Kind
	// }
	// var err error

	// var ret1 []interface{}
	// if in.Metadata != nil {
	// 	v, ok := obj["metadata"].([]interface{})
	// 	if !ok {
	// 		v = []interface{}{}
	// 	}
	// 	ret1 = flattenMetadataV3(in.Metadata, v)
	// }

	// err = d.Set("metadata", ret1)
	// if err != nil {
	// 	return err
	// }

	// var ret2 []interface{}
	// if in.Spec != nil {
	// 	v, ok := obj["spec"].([]interface{})
	// 	if !ok {
	// 		v = []interface{}{}
	// 	}

	// 	//	ret2 = flattenClusterV3Spec(in.Spec, v)
	// }

	// err = d.Set("spec", ret2)
	// if err != nil {
	// 	return err
	// }

	return nil
}
