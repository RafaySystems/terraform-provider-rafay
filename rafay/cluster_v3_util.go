package rafay

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/RafaySystems/rafay-common/proto/types/hub/infrapb"
	"github.com/RafaySystems/terraform-provider-rafay/rafay/gke"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func expandClusterV3(in *schema.ResourceData) (*infrapb.Cluster, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "expand cluster invoked with empty input")
	}
	obj := &infrapb.Cluster{}

	if v, ok := in.Get("metadata").([]interface{}); ok && len(v) > 0 {
		obj.Metadata = expandMetaData(v)
	}

	if v, ok := in.Get("spec").([]interface{}); ok && len(v) > 0 {
		objSpec, err := expandClusterV3Spec(v)
		if err != nil {
			return nil, err
		}
		log.Println("expandClusterSpec got spec")
		obj.Spec = objSpec
	}

	obj.ApiVersion = "infra.k8smgmt.io/v3"
	obj.Kind = "Cluster"

	return obj, nil
}

func expandClusterV3Spec(p []interface{}) (*infrapb.ClusterSpec, error) {
	obj := &infrapb.ClusterSpec{}
	if len(p) == 0 || p[0] == nil {
		return obj, fmt.Errorf("%s", "expandClusterSpec empty input")
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["type"].(string); ok && len(v) > 0 {
		obj.Type = v
	}

	if obj.Type != "aks" || strings.EqualFold(obj.Type, "gke") { // TODO: update cluster type with consts
		return nil, errors.New("cluster type not implemented")
	}

	if v, ok := in["sharing"].([]interface{}); ok && len(v) > 0 {
		obj.Sharing = expandSharingSpecV3(v)
	}

	if v, ok := in["blueprint_config"].([]interface{}); ok && len(v) > 0 {
		obj.BlueprintConfig = expandClusterV3Blueprint(v)
	}

	if v, ok := in["cloud_credentials"].(string); ok && len(v) > 0 {
		obj.CloudCredentials = v
	}

	if strings.EqualFold(obj.Type, "gke") { // TODO: update cluster type with consts
		if v, ok := in["config"].([]interface{}); ok && len(v) > 0 {
			//	obj.Config = expandAKSClusterV3Config(v)
			// TODO
			obj.Config = gke.ExpandClusterV3Config(v)
		}
		return obj, nil
	}

	switch obj.Type {
	case "aks":
		if v, ok := in["config"].([]interface{}); ok && len(v) > 0 {
			obj.Config = expandAKSClusterV3Config(v)
		}
	default:
		log.Fatalln("Not Implemented")
	}

	return obj, nil
}
