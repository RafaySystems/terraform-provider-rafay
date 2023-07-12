package rafay

import (
	"github.com/RafaySystems/rafay-common/proto/types/hub/infrapb"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// convert from V3 schema in rafay-common proto --> tf schema
func flattenGKEClusterV3(d *schema.ResourceData, in *infrapb.Cluster) error {
	/*
		Cluster:
		- apiversion
		- kind
		- metadata
		- spec
	*/
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

		ret2 = flattenGKEV3Spec(in.Spec, v)
	}

	err = d.Set("spec", ret2)
	if err != nil {
		return err
	}

	return nil
}

func flattenGKEV3Spec(in *infrapb.ClusterSpec, p []interface{}) []interface{} {
	/*
		Spec:
		- type
		- sharing
		- cloudCredentials
		- blueprint
		- proxy
		- config --- gke

	*/

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

	if in.Sharing != nil {
		obj["sharing"] = flattenSharingSpecV3(in.Sharing)
	}

	if len(in.CloudCredentials) > 0 {
		obj["cloud_credentials"] = in.CloudCredentials
	}

	if in.Blueprint != nil {
		obj["blueprint"] = flattenClusterGKEV3Blueprint(in.Blueprint)
	}

	// TODO: proxy

	if in.GetGke() != nil {
		v, ok := obj["config"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["config"] = flattenGKEV3Config(in.GetGke(), v)
	}

	return []interface{}{obj}

}

// Note: this uses ClusterBlueprint
func flattenClusterGKEV3Blueprint(in *infrapb.ClusterBlueprint) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}

	if len(in.Name) > 0 {
		obj["name"] = in.Name
	}

	if len(in.Version) > 0 {
		obj["version"] = in.Version
	}

	return []interface{}{obj}
}

func flattenGKEV3Config(in *infrapb.GkeV3ConfigObject, p []interface{}) []interface{} {
	/*
		Config (gke/GkeV3ConfigObject):
		- gcp project
		- controlplaneversion
		- prebootstrapcommands
		- location
		- network
		- nodepools
		- security
		- Feature
	*/

	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.GcpProject) > 0 {
		obj["gcp_project"] = in.GcpProject
	}

	if len(in.ControlPlaneVersion) > 0 {
		obj["control_plane_version"] = in.ControlPlaneVersion
	}

	if in.PreBootstrapCommands != nil && len(in.PreBootstrapCommands) > 0 {
		obj["pre_bootstrap_commands"] = toArrayInterface(in.PreBootstrapCommands)
	}

	if in.Location != nil {
		v, ok := obj["location"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["location"] = flattenGKEV3Location(in.Location, v)
	}

	// network
	if in.Network != nil {
		v, ok := obj["network"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["network"] = flattenGKEV3Network(in.Network, v)

	}

	// nodepools
	if in.NodePools != nil && len(in.NodePools) > 0 {
		v, ok := obj["node_pools"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["node_pools"] = flattenGKEV3Nodepools(in.NodePools, v)
	}

	if in.Security != nil {

	}

	if in.Features != nil {

	}

	return []interface{}{obj}
}

func flattenGKEV3Location(in *infrapb.GkeLocation, p []interface{}) []interface{} {
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

	if in.DefaultNodeLocations != nil {
		v, ok := obj["default_node_locations"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["default_node_locations"] = flattenGKEV3DefaultNodeLocations(in.DefaultNodeLocations, v)
	}

	if in.GetRegional() != nil { // ??? TODO

	}

	return []interface{}{obj}
}

func flattenGKEV3DefaultNodeLocations(in *infrapb.GkeDefaultNodeLocation, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	obj["enabled"] = in.Enabled // TODO: check if this works

	if in.Zones != nil && len(in.Zones) > 0 {
		obj["zones"] = toArrayInterface(in.Zones)
	}

	return []interface{}{obj}
}

func flattenGKEV3Network(in *infrapb.GkeNetwork, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	obj["name"] = in.Name
	obj["subnet_name"] = in.SubnetName
	obj["enable_vpc_nativetraffic"] = in.EnableVPCNativetraffic
	obj["pod_address_range"] = in.PodAddressRange
	obj["service_address_range"] = in.ServiceAddressRange

	// max_pods_per_node
	obj["max_pods_per_node"] = in.MaxPodsPerNode // TODO: check if this works

	// access
	if in.Access != nil {
		v, ok := obj["access"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["access"] = flattenGKEV3NetworkAccess(in.Access, v)
	}

	// control_plane_authorized_network
	if in.ControlPlaneAuthorizedNetwork != nil {
		v, ok := obj["control_plane_authorized_network"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["control_plane_authorized_network"] = flattenGKEV3ControlPlaneAuthorizedNetwork(in.ControlPlaneAuthorizedNetwork, v)
	}

	return []interface{}{obj}
}

func flattenGKEV3NetworkAccess(in *infrapb.GkeAccess, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	obj["type"] = in.Type

	if in.GetPrivate() != nil { // TODO

	}

	return []interface{}{obj}
}

func flattenGKEV3ControlPlaneAuthorizedNetwork(in *infrapb.GkeControlPlaneAuthorizedNetwork, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	obj["enabled"] = in.Enabled

	if in.AuthorizedNetwork != nil && len(in.AuthorizedNetwork) > 0 {
		v, ok := obj["authorized_network"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["authorized_network"] = flattenGKEV3AuthorizedNetwork(in.AuthorizedNetwork, v)
	}

	return []interface{}{obj}
}

func flattenGKEV3AuthorizedNetwork(in []*infrapb.GkeAuthorizedNetwork, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	out := make([]interface{}, len(in))
	for i, in := range in {
		obj := map[string]interface{}{}
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}

		obj["name"] = in.Name
		obj["cidr"] = in.Cidr
	}

	return out
}

func flattenGKEV3Nodepools(in []*infrapb.Nodepool, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	out := make([]interface{}, len(in))
	for i, in := range in {
		obj := map[string]interface{}{}
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}

		// TODO all np fields

	}

	return out

}

// func flattenGKEV3--(in *infrapb., p []interface{}) []interface{} {
// 	if in == nil {
// 		return nil
// 	}
// 	obj := map[string]interface{}{}
// 	if len(p) != 0 && p[0] != nil {
// 		obj = p[0].(map[string]interface{})
// 	}

// 	// TODO

// 	return []interface{}{obj}
// }
