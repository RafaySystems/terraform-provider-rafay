package rafay

import (
	"errors"
	"fmt"
	"strings"

	"github.com/RafaySystems/rafay-common/proto/types/hub/infrapb"
)

// takes input given in the format of the terraform schema and populate the backend structure for that resource.
// convert from tf schema --> V3 schema in rafay-common proto

// GkeV3ConfigObject
func expandToV3GkeConfigObject(p []interface{}) (*infrapb.ClusterSpec_Gke, error) {
	obj := &infrapb.ClusterSpec_Gke{
		Gke: &infrapb.GkeV3ConfigObject{}}

	if len(p) == 0 || p[0] == nil {
		return obj, errors.New("got nil or empty object for gke config") // TODO: review this: Does it matter whether we return nil or obj here?
	}
	in := p[0].(map[string]interface{})

	/*
		gcp project
		location
		controlplaneversion
		network
		security
		Feature
		nodepools
		prebootstrapcommands
	*/

	if v, ok := in["gcp_project"].(string); ok && len(v) > 0 {
		obj.Gke.GcpProject = v
	} else if !ok {
		return nil, errors.New("missing gcp project name")
	}

	if v, ok := in["control_plane_version"].(string); ok && len(v) > 0 {
		obj.Gke.ControlPlaneVersion = v
	} else if !ok {
		return nil, errors.New("missing controlplane version in input")
	}

	var err error
	// location
	if v, ok := in["location"].([]interface{}); ok && len(v) > 0 {
		obj.Gke.Location, err = expandToV3GkeLocation(v)
		if err != nil {
			return nil, fmt.Errorf("failed to expand gke location from schema " + err.Error())
		}
	}

	// network
	if v, ok := in["network"].([]interface{}); ok && len(v) > 0 {
		obj.Gke.Network, err = expandToV3GkeNetwork(v)
		if err != nil {
			return nil, fmt.Errorf("failed to expand gke network from schema " + err.Error())
		}
	}

	// security

	// feature

	// nodepools
	if v, ok := in["node_pools"].([]interface{}); ok && len(v) > 0 {
		obj.Gke.NodePools, err = expandToV3GkeNodepools(v)
	}

	// prebootstrapCommands
	if v, ok := in["pre_bootstrap_commands"].([]interface{}); ok && len(v) > 0 {
		obj.Gke.PreBootstrapCommands = toArrayString(v)
	}

	return obj, nil

}

func expandToV3GkeLocation(p []interface{}) (*infrapb.GkeLocation, error) {
	obj := &infrapb.GkeLocation{}

	if len(p) == 0 || p[0] == nil {
		return obj, errors.New("got nil for gke location")
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["type"].(string); ok && len(v) > 0 {
		obj.Type = v
	}

	var err error
	// GkeDefaultNodeLocation
	if v, ok := in["default_node_locations"].([]interface{}); ok && len(v) > 0 {
		obj.DefaultNodeLocations, err = expandToV3GkeDefaultNodeLocation(v)
		if err != nil {
			return obj, fmt.Errorf("failed to expand gke default node locations " + err.Error())
		}
	}

	// zonal/regional
	if v, ok := in["config"].([]interface{}); ok && len(v) > 0 {
		if strings.EqualFold(obj.Type, GKE_ZONAL_CLUSTER_TYPE) {
			obj.Config, err = expandToV3GkeZonalCluster(v)
			return nil, fmt.Errorf("failed to expand gke zonal cluster " + err.Error())
		} else if strings.EqualFold(obj.Type, GKE_REGIONAL_CLUSTER_TYPE) {
			obj.Config, err = expandToV3GkeRegionalCluster(v)
			return nil, fmt.Errorf("failed to expand gke regional cluster " + err.Error())
		}
	}

	return obj, nil
}

func expandToV3GkeDefaultNodeLocation(p []interface{}) (*infrapb.GkeDefaultNodeLocation, error) {
	if len(p) == 0 || p[0] == nil {
		return nil, errors.New("got nil for gke default node locations")
	}

	obj := &infrapb.GkeDefaultNodeLocation{}

	in := p[0].(map[string]interface{})

	if v, ok := in["enabled"].(bool); ok {
		obj.Enabled = v
	}

	//if obj.Enabled {
	if v, ok := in["zones"].([]interface{}); ok && len(v) > 0 {
		obj.Zones = toArrayString(v)
	}
	//}

	return obj, nil

}

func expandToV3GkeZonalCluster(p []interface{}) (*infrapb.GkeLocation_Zonal, error) {
	obj := &infrapb.GkeLocation_Zonal{
		Zonal: &infrapb.GkeZonalCluster{},
	}

	if len(p) == 0 || p[0] == nil {
		return obj, errors.New("got nil for gke zonal cluster")
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["zone"].(string); ok && len(v) > 0 {
		obj.Zonal.Zone = v
	}

	return obj, nil
}

func expandToV3GkeRegionalCluster(p []interface{}) (*infrapb.GkeLocation_Regional, error) {
	obj := &infrapb.GkeLocation_Regional{
		Regional: &infrapb.GkeRegionalCluster{},
	}

	if len(p) == 0 || p[0] == nil {
		return obj, errors.New("got nil for gke regional cluster")
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["region"].(string); ok && len(v) > 0 {
		obj.Regional.Region = v
	}

	if v, ok := in["zone"].(string); ok && len(v) > 0 {
		obj.Regional.Zone = v
	}

	return obj, nil
}

// GkeNetwork
func expandToV3GkeNetwork(p []interface{}) (*infrapb.GkeNetwork, error) {
	if len(p) == 0 || p[0] == nil {
		return nil, errors.New("got nil for gke network config")
	}

	obj := &infrapb.GkeNetwork{}
	in := p[0].(map[string]interface{})

	if v, ok := in["name"].(string); ok && len(v) > 0 {
		obj.Name = v
	}

	if v, ok := in["subnet_name"].(string); ok && len(v) > 0 {
		obj.SubnetName = v
	}

	var err error
	// access
	if v, ok := in["access"].([]interface{}); ok && len(v) > 0 {
		obj.Access, err = expandToV3GkeAccess(v)
		if err != nil {
			return obj, fmt.Errorf("failed to expand gke cluster access " + err.Error())
		}
	}

	// enable_vpc_nativetraffic
	if v, ok := in["enable_vpc_nativetraffic"].(bool); ok {
		obj.EnableVPCNativetraffic = v
	}

	if v, ok := in["max_pods_per_node"].(int); ok && v > 0 {
		obj.MaxPodsPerNode = int64(v)
	}

	// pod_address_range
	if v, ok := in["pod_address_range"].(string); ok {
		obj.PodAddressRange = v
	}

	if v, ok := in["service_address_range"].(string); ok {
		obj.ServiceAddressRange = v
	}

	if v, ok := in["control_plane_authorized_network"].([]interface{}); ok && len(v) > 0 {
		obj.ControlPlaneAuthorizedNetwork, err = expandToV3GkeControlPlaneAuthorizedNetwork(v)
		if err != nil {
			return obj, fmt.Errorf("failed to expand control plane authorized network " + err.Error())
		}
	}

	return obj, nil
}

func expandToV3GkeAccess(p []interface{}) (*infrapb.GkeAccess, error) {
	if len(p) == 0 || p[0] == nil {
		return nil, errors.New("got nil for gke network config")
	}

	obj := &infrapb.GkeAccess{}
	in := p[0].(map[string]interface{})

	if v, ok := in["type"].(string); ok && len(v) > 0 {
		obj.Type = v
	}

	var err error
	// public or private cluster
	if v, ok := in["config"].([]interface{}); ok && len(v) > 0 {
		if strings.EqualFold(obj.Type, GKE_PRIVATE_CLUSTER_TYPE) {
			obj.Config, err = expandToV3GkePrivateCluster(v)
			if err != nil {
				return obj, fmt.Errorf("failed to expand gke private cluster config " + err.Error())
			}
		}
		// else if strings.EqualFold(obj.Type, GKE_PUBLIC_CLUSTER_TYPE) {
		// 	// no public cluster specific config as of now
		// }
	}

	return obj, nil
}

func expandToV3GkePrivateCluster(p []interface{}) (*infrapb.GkeAccess_Private, error) {
	if len(p) == 0 || p[0] == nil {
		return nil, errors.New("got nil for gke private network config")
	}

	obj := &infrapb.GkeAccess_Private{
		Private: &infrapb.GkePrivateCluster{},
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["control_plane_ip_range"].(string); ok && len(v) > 0 {
		obj.Private.ControlPlaneIPRange = v
	}

	if v, ok := in["enable_access_control_plane_external_ip"].(bool); ok {
		obj.Private.EnableAccessControlPlaneExternalIP = v
	}

	if v, ok := in["enable_access_control_plane_global"].(bool); ok {
		obj.Private.EnableAccessControlPlaneGlobal = v
	}

	if v, ok := in["disable_snat"].(bool); ok {
		obj.Private.DisableSNAT = v
	}

	return obj, nil
}

// GkeControlPlaneAuthorizedNetwork
func expandToV3GkeControlPlaneAuthorizedNetwork(p []interface{}) (*infrapb.GkeControlPlaneAuthorizedNetwork, error) {
	if len(p) == 0 || p[0] == nil {
		return nil, errors.New("got nil for gke control plane authorized network")
	}

	obj := &infrapb.GkeControlPlaneAuthorizedNetwork{}
	in := p[0].(map[string]interface{})

	if v, ok := in["enabled"].(bool); ok {
		obj.Enabled = v
	}

	var err error
	if v, ok := in["authorized_network"].([]interface{}); ok && len(v) > 0 {
		//obj.Zones = toArrayString(v)
		obj.AuthorizedNetwork, err = expandToV3GkeAuthorizedNetwork(v)
		if err != nil {
			return obj, fmt.Errorf("failed to expand Gke authorized networks " + err.Error())
		}
	}

	return obj, nil
}

// expandToV3GkeAuthorizedNetwork
func expandToV3GkeAuthorizedNetwork(p []interface{}) ([]*infrapb.GkeAuthorizedNetwork, error) {
	if len(p) == 0 || p[0] == nil {
		return nil, errors.New("got nil for gke authorized network config")
	}

	out := make([]*infrapb.GkeAuthorizedNetwork, len(p))

	for i := range p {
		obj := &infrapb.GkeAuthorizedNetwork{}
		in := p[i].(map[string]interface{})

		if v, ok := in["name"].(string); ok {
			obj.Name = v
		}

		if v, ok := in["cidr"].(string); ok {
			obj.Name = v
		}
		out[i] = obj
	}

	return out, nil
}

// expandToV3GkeNodepools
func expandToV3GkeNodepools(p []interface{}) ([]*infrapb.GkeNodePool, error) {
	if len(p) == 0 || p[0] == nil {
		return nil, errors.New("got nil for gke network config")
	}

	out := make([]*infrapb.GkeNodePool, len(p))
	for i := range p {
		obj := &infrapb.GkeNodePool{}

		in := p[i].(map[string]interface{})

		if v, ok := in["name"].(string); ok {
			obj.Name = v
		}

		if v, ok := in["node_version"].(string); ok {
			obj.Name = v
		}

		if v, ok := in["size"].(int); ok && v > 0 {
			obj.Size = int64(v)
		}

		// GkeNodeLocation
		var err error
		if v, ok := in["node_locations"].([]interface{}); ok && len(v) > 0 {
			obj.NodeLocations, err = expandToV3GkeNodeLocation(v)
			if err != nil {
				return nil, fmt.Errorf("failed to expand Gke node locations " + err.Error())
			}
		}

		// GkeNodeAutoScale
		if v, ok := in["auto_scaling"].([]interface{}); ok && len(v) > 0 {
			obj.AutoScaling, err = expandToV3GkeNodeAutoScale(v)
			if err != nil {
				return nil, fmt.Errorf("failed to expand Gke autoscaling " + err.Error())
			}
		}

		// GkeNodeMachineConfig
		if v, ok := in["machine_config"].([]interface{}); ok && len(v) > 0 {
			obj.MachineConfig, err = expandToV3GkeNodeMachineConfig(v)
			if err != nil {
				return nil, fmt.Errorf("failed to expand Gke machine config " + err.Error())
			}
		}

		// GkeNodeNetworking
		if v, ok := in["networking"].([]interface{}); ok && len(v) > 0 {
			obj.Networking, err = expandToV3GkeNodeNetworking(v)
			if err != nil {
				return nil, fmt.Errorf("failed to expand Gke node networking " + err.Error())
			}
		}

		// GkeNodeSecurity
		if v, ok := in["security"].([]interface{}); ok && len(v) > 0 {
			obj.Security, err = expandToV3GkeNodeSecurity(v)
			if err != nil {
				return nil, fmt.Errorf("failed to expand Gke node security " + err.Error())
			}
		}

		// TODO GkeNodeMetadata
		// 		if v, ok := in[""].([]interface{}); ok && len(v) > 0 {
		// 	if err != nil {
		// 		return obj, fmt.Errorf("failed to expand Gke authorized networks " + err.Error())
		// 	}
		// }

	}

	return out, nil
}

// GkeNodeLocation
func expandToV3GkeNodeLocation(p []interface{}) (*infrapb.GkeNodeLocation, error) {
	if len(p) == 0 || p[0] == nil {
		return nil, errors.New("got nil for gke node location config")
	}

	obj := &infrapb.GkeNodeLocation{}
	in := p[0].(map[string]interface{})

	if v, ok := in["enabled"].(bool); ok {
		obj.Enabled = v
	}

	if v, ok := in["zones"].([]interface{}); ok && len(v) > 0 {
		obj.Zones = toArrayString(v)
	}

	return obj, nil
}

// GkeNodeAutoScale
func expandToV3GkeNodeAutoScale(p []interface{}) (*infrapb.GkeNodeAutoScale, error) {
	if len(p) == 0 || p[0] == nil {
		return nil, errors.New("got nil for gke node autoscale config")
	}

	obj := &infrapb.GkeNodeAutoScale{}
	in := p[0].(map[string]interface{})

	if v, ok := in["min_nodes"].(int); ok && v > 0 {
		obj.MinNodes = int64(v)
	}

	if v, ok := in["max_nodes"].(int); ok && v > 0 {
		obj.MaxNodes = int64(v)
	}

	return obj, nil
}

// GkeNodeMachineConfig
func expandToV3GkeNodeMachineConfig(p []interface{}) (*infrapb.GkeNodeMachineConfig, error) {
	if len(p) == 0 || p[0] == nil {
		return nil, errors.New("got nil for gke node machine config")
	}

	obj := &infrapb.GkeNodeMachineConfig{}
	in := p[0].(map[string]interface{})

	if v, ok := in["machine_type"].(string); ok {
		obj.MachineType = v
	}

	if v, ok := in["image_type"].(string); ok {
		obj.ImageType = v
	}

	if v, ok := in["boot_disk_type"].(string); ok {
		obj.BootDiskType = v
	}

	if v, ok := in["boot_disk_size"].(int); ok && v > 0 {
		obj.BootDiskSize = int64(v)
	}

	return obj, nil
}

// GkeNodeNetworking
func expandToV3GkeNodeNetworking(p []interface{}) (*infrapb.GkeNodeNetworking, error) {
	if len(p) == 0 || p[0] == nil {
		return nil, errors.New("got nil for gke node network config")
	}

	obj := &infrapb.GkeNodeNetworking{}
	in := p[0].(map[string]interface{})

	if v, ok := in["max_pods_per_node"].(int); ok && v > 0 {
		obj.MaxPodsPerNode = int64(v)
	}

	if v, ok := in["network_tags"].([]interface{}); ok && len(v) > 0 {
		obj.NetworkTags = toArrayString(v)
	}

	return obj, nil
}

// GkeNodeSecurity
func expandToV3GkeNodeSecurity(p []interface{}) (*infrapb.GkeNodeSecurity, error) {
	if len(p) == 0 || p[0] == nil {
		return nil, errors.New("got nil for gke node security config")
	}

	obj := &infrapb.GkeNodeSecurity{}
	in := p[0].(map[string]interface{})

	if v, ok := in["enable_integrity_monitoring"].(bool); ok {
		obj.EnableIntegrityMonitoring = v
	}

	if v, ok := in["enable_secure_boot"].(bool); ok {
		obj.EnableSecureBoot = v
	}

	return obj, nil
}

// func expandToV3GkeNetwork(p []interface{}) (*infrapb.GkeNetwork, error) {
// 	if len(p) == 0 || p[0] == nil {
// 		return nil, errors.New("got nil for gke network config")
// 	}

// 	obj := &infrapb.GkeNetwork{}
// in := p[0].(map[string]interface{})

// 	return obj, nil
// }
