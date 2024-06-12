package rafay

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/RafaySystems/rafay-common/proto/types/hub/infrapb"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// takes input given in the format of the terraform schema and populate the backend structure for that resource.
// convert from tf schema --> V3 schema in rafay-common proto
func expandGKEClusterToV3(in *schema.ResourceData) (*infrapb.Cluster, error) {
	/*
		Cluster:
		- apiversion
		- kind
		- metadata
		- spec
	*/
	if in == nil {
		return nil, fmt.Errorf("%s", "expand cluster invoked with empty input")
	}
	obj := &infrapb.Cluster{}

	obj.ApiVersion = V3_CLUSTER_APIVERSION
	obj.Kind = V3_CLUSTER_KIND

	v, ok := in.Get("metadata").([]interface{})
	if !ok || len(v) == 0 {
		return nil, fmt.Errorf("%s", "expand cluster invoked with empty metadata")
	}

	if ok && len(v) > 0 {
		obj.Metadata = expandMetaData(v)
	}

	// spec
	if v, ok := in.Get("spec").([]interface{}); ok && len(v) > 0 {
		objSpec, err := expandGKEClusterToV3Spec(v)
		if err != nil {
			return nil, err
		}
		obj.Spec = objSpec
	}

	log.Println("In expandGKEClusterToV3. infrapb obj", obj)

	return obj, nil
}

func expandGKEClusterToV3Spec(p []interface{}) (*infrapb.ClusterSpec, error) {
	/*
		Spec:
		- type
		- sharing
		- cloudCredentials
		- blueprint
		- proxy
		- config --- gke

	*/

	obj := &infrapb.ClusterSpec{}
	if len(p) == 0 || p[0] == nil {
		return obj, fmt.Errorf("%s", "expandClusterSpec empty input")
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["sharing"].([]interface{}); ok && len(v) > 0 {
		obj.Sharing = expandSharingSpecV3(v)
	}

	if v, ok := in["blueprint"].([]interface{}); ok && len(v) > 0 {
		var err error
		obj.Blueprint, err = expandGKEClusterToV3Blueprint(v)
		if err != nil {
			return nil, fmt.Errorf("failed to expand blueprint " + err.Error())
		}
	}

	if v, ok := in["cloud_credentials"].(string); ok && len(v) > 0 {
		obj.CloudCredentials = v
	}

	if v, ok := in["proxy"].([]interface{}); ok && len(v) > 0 {
		var err error
		obj.Proxy, err = expandToV3GKEProxy(v)
		if err != nil {
			return nil, fmt.Errorf("failed to expand proxy " + err.Error())
		}
	}
	if v, ok := in["type"].(string); ok && len(v) > 0 {
		obj.Type = v
	}

	if !strings.EqualFold(obj.Type, GKE_CLUSTER_TYPE) {
		log.Println("In expandGKEClusterToV3Spec. Got non-GKE cluster. cluster type not implemented")
		return nil, errors.New("expandGKEClusterToV3Spec got non-GKE cluster. cluster type not implemented")
	}

	if strings.EqualFold(obj.Type, GKE_CLUSTER_TYPE) {
		if v, ok := in["config"].([]interface{}); ok && len(v) > 0 {
			var err error
			obj.Config, err = expandToV3GkeConfigObject(v)
			if err != nil {
				return nil, fmt.Errorf("failed to expand to gke config " + err.Error())
			}
		}
	}

	return obj, nil
}

func expandGKEClusterToV3Blueprint(p []interface{}) (*infrapb.ClusterBlueprint, error) {
	obj := &infrapb.ClusterBlueprint{}
	if len(p) == 0 || p[0] == nil {
		return obj, errors.New("empty blueprint in input")
	}

	in := p[0].(map[string]interface{})
	if v, ok := in["name"].(string); ok {
		obj.Name = v
	} else if !ok {
		return nil, errors.New("missing blueprint name")
	}

	if v, ok := in["version"].(string); ok {
		obj.Version = v
	} else if !ok {
		return nil, errors.New("missing blueprint version")
	}

	log.Println("expandGKEClusterToV3Blueprint obj", obj)
	return obj, nil
}

// GkeProxy
func expandToV3GKEProxy(p []interface{}) (*infrapb.ClusterProxy, error) {
	obj := &infrapb.ClusterProxy{}
	if len(p) == 0 || p[0] == nil {
		return obj, errors.New("empty proxy in input")
	}

	in := p[0].(map[string]interface{})
	if v, ok := in["enabled"].(bool); ok {
		obj.Enabled = v
	}

	if v, ok := in["http_proxy"].(string); ok && len(v) > 0 {
		obj.HttpProxy = v
	}

	if v, ok := in["https_proxy"].(string); ok && len(v) > 0 {
		obj.HttpsProxy = v
	}

	if v, ok := in["no_proxy"].(string); ok && len(v) > 0 {
		obj.NoProxy = v
	}

	if v, ok := in["proxy_auth"].(string); ok && len(v) > 0 {
		obj.ProxyAuth = v
	}

	if v, ok := in["bootstrap_ca"].(string); ok && len(v) > 0 {
		obj.BootstrapCA = v
	}

	if v, ok := in["allow_insecure_bootstrap"].(bool); ok {
		obj.AllowInsecureBootstrap = v
	}

	return obj, nil
}

// GkeV3ConfigObject
func expandToV3GkeConfigObject(p []interface{}) (*infrapb.ClusterSpec_Gke, error) {
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

	obj := &infrapb.ClusterSpec_Gke{
		Gke: &infrapb.GkeV3ConfigObject{},
	}

	if len(p) == 0 || p[0] == nil {
		return obj, errors.New("got nil or empty object for gke config")
	}
	in := p[0].(map[string]interface{})

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
	if v, ok := in["security"].([]interface{}); ok && len(v) > 0 {
		obj.Gke.Security, err = expandToV3GkeSecurity(v)
		if err != nil {
			return nil, fmt.Errorf("failed to expand gke security from schema " + err.Error())
		}
	}

	// feature
	if v, ok := in["features"].([]interface{}); ok && len(v) > 0 {
		obj.Gke.Features, err = expandToV3GkeFeatures(v)
		if err != nil {
			return nil, fmt.Errorf("failed to expand gke feature from schema " + err.Error())
		}
	}

	// nodepools
	if v, ok := in["node_pools"].([]interface{}); ok && len(v) > 0 {
		obj.Gke.NodePools, err = expandToV3GkeNodepools(v)
		if err != nil {
			return obj, fmt.Errorf("failed to expand gke nodepool " + err.Error())
		}
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
		obj.DefaultNodeLocations, err = expandToV3GkeDefaultNodeLocations(v)
		if err != nil {
			return obj, fmt.Errorf("failed to expand gke default node locations " + err.Error())
		}
	}

	// zonal/regional
	if v, ok := in["config"].([]interface{}); ok && len(v) > 0 {
		if strings.EqualFold(obj.Type, GKE_ZONAL_CLUSTER_TYPE) {
			log.Printf("Invoking expandToV3GkeZonalCluster %s ", v)
			//	zonalConfig := &infrapb.GkeLocation_Zonal{}
			obj.Config, err = expandToV3GkeZonalCluster(v)
			if err != nil {
				return nil, fmt.Errorf("failed to expand gke zonal cluster " + err.Error())
			}
		} else if strings.EqualFold(obj.Type, GKE_REGIONAL_CLUSTER_TYPE) {
			obj.Config, err = expandToV3GkeRegionalCluster(v)
			if err != nil {
				return nil, fmt.Errorf("failed to expand gke regional cluster " + err.Error())
			}
		}
	}

	return obj, nil
}

func expandToV3GkeDefaultNodeLocations(p []interface{}) (*infrapb.GkeDefaultNodeLocation, error) {
	if len(p) == 0 || p[0] == nil {
		return nil, errors.New("got nil for gke default node locations")
	}

	obj := &infrapb.GkeDefaultNodeLocation{}

	in := p[0].(map[string]interface{})

	if v, ok := in["enabled"].(bool); ok {
		obj.Enabled = v
	}

	if v, ok := in["zones"].([]interface{}); ok && len(v) > 0 {
		obj.Zones = toArrayString(v)
	}

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

func expandToV3GkeSecurity(p []interface{}) (*infrapb.GkeSecurity, error) {
	if len(p) == 0 || p[0] == nil {
		return nil, errors.New("got nil for gke security config")
	}

	obj := &infrapb.GkeSecurity{}
	in := p[0].(map[string]interface{})

	if v, ok := in["enable_workload_identity"].(bool); ok {
		obj.EnableWorkloadIdentity = v
	}

	if v, ok := in["enable_google_groups_for_rbac"].(bool); ok {
		obj.EnableGoogleGroupsForRbac = v
	}

	if v, ok := in["security_group"].(string); ok && len(v) > 0 {
		obj.SecurityGroup = v
	}

	if v, ok := in["enable_legacy_authorization"].(bool); ok {
		obj.EnableLegacyAuthorization = v
	}

	if v, ok := in["issue_client_certificate"].(bool); ok {
		obj.IssueClientCertificate = v
	}

	return obj, nil
}

func expandToV3GkeFeatures(p []interface{}) (*infrapb.GkeFeatures, error) {
	if len(p) == 0 || p[0] == nil {
		return nil, errors.New("got nil for gke security config")
	}

	obj := &infrapb.GkeFeatures{}
	in := p[0].(map[string]interface{})

	if v, ok := in["enable_cloud_logging"].(bool); ok {
		obj.EnableCloudLogging = v
	}

	if v, ok := in["cloud_logging_components"].([]interface{}); ok && len(v) > 0 {
		obj.CloudLoggingComponents = toArrayString(v)
	}

	if v, ok := in["enable_cloud_monitoring"].(bool); ok {
		obj.EnableCloudMonitoring = v
	}

	if v, ok := in["cloud_monitoring_components"].([]interface{}); ok && len(v) > 0 {
		obj.CloudMonitoringComponents = toArrayString(v)
	}

	if v, ok := in["enable_managed_service_prometheus"].(bool); ok {
		obj.EnableManagedServicePrometheus = v
	}

	if v, ok := in["enable_application_manager_beta"].(bool); ok {
		obj.EnableApplicationManagerBeta = v
	}

	if v, ok := in["enable_backup_for_gke"].(bool); ok {
		obj.EnableBackupForGke = v
	}

	if v, ok := in["enable_compute_engine_persistent_disk_csi_driver"].(bool); ok {
		obj.EnableComputeEnginePersistentDiskCSIDriver = v
	}

	if v, ok := in["enable_filestore_csi_driver"].(bool); ok {
		obj.EnableFilestoreCSIDriver = v
	}

	if v, ok := in["enable_image_streaming"].(bool); ok {
		obj.EnableImageStreaming = v
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
	if v, ok := in["pod_address_range"].(string); ok && len(v) > 0 {
		obj.PodAddressRange = v
	}

	if v, ok := in["service_address_range"].(string); ok && len(v) > 0 {
		obj.ServiceAddressRange = v
	}

	if v, ok := in["pod_secondary_range_name"].(string); ok && len(v) > 0 {
		obj.PodSecondaryRangeName = v
	}

	if v, ok := in["service_secondary_range_name"].(string); ok && len(v) > 0 {
		obj.ServiceSecondaryRangeName = v
	}

	if v, ok := in["data_plane_v_2"].(string); ok && len(v) > 0 {
		obj.DataPlaneV2 = v
	}

	if v, ok := in["enable_data_plane_v_2_metrics"].(bool); ok {
		obj.EnableDataPlaneV2Metrics = v
	}

	if v, ok := in["enable_data_plane_v_2_observability"].(bool); ok {
		obj.EnableDataPlaneV2Observability = v
	}

	if v, ok := in["network_policy_config"].(bool); ok {
		obj.NetworkPolicyConfig = v
	}

	if v, ok := in["network_policy"].(string); ok && len(v) > 0 {
		obj.NetworkPolicy = v
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

	var err error
	if v, ok := in["firewall_rules"].([]interface{}); ok && len(v) > 0 {
		obj.Private.FirewallRules, err = expandToV3GkeFirewalls(v)
		if err != nil {
			return obj, fmt.Errorf("failed to expand gke private cluster firewall rules " + err.Error())
		}
	}

	return obj, nil
}

// expandToV3GkeFirewalls
func expandToV3GkeFirewalls(p []interface{}) ([]*infrapb.FirewallRule, error) {
	if len(p) == 0 || p[0] == nil {
		return nil, errors.New("got nil for gke firewall rule config")
	}

	out := make([]*infrapb.FirewallRule, len(p))

	for i := range p {
		obj := &infrapb.FirewallRule{}
		in := p[i].(map[string]interface{})

		if v, ok := in["action"].(string); ok && len(v) > 0 {
			obj.Action = v
		}

		if v, ok := in["description"].(string); ok && len(v) > 0 {
			obj.Description = v
		}

		if v, ok := in["destination_ranges"].([]interface{}); ok && len(v) > 0 {
			obj.DestinationRanges = toArrayString(v)
		}

		if v, ok := in["direction"].(string); ok && len(v) > 0 {
			obj.Direction = v
		}

		if v, ok := in["name"].(string); ok && len(v) > 0 {
			obj.Name = v
		}

		if v, ok := in["network"].(string); ok && len(v) > 0 {
			obj.Network = v
		}

		if v, ok := in["priority"].(int); ok {
			obj.Priority = int32(v)
		}

		if v, ok := in["source_ranges"].([]interface{}); ok && len(v) > 0 {
			obj.SourceRanges = toArrayString(v)
		}

		if v, ok := in["target_tags"].([]interface{}); ok && len(v) > 0 {
			obj.TargetTags = toArrayString(v)
		}

		var err error
		if v, ok := in["rules"].([]interface{}); ok && len(v) > 0 {
			obj.Rules, err = expandToV3GkeFirewallRules(v)
			if err != nil {
				return nil, fmt.Errorf("failed to expand gke private cluster firewall rules " + err.Error())
			}
		}

		out[i] = obj
	}

	return out, nil
}

func expandToV3GkeFirewallRules(p []interface{}) ([]*infrapb.Rule, error) {
	if len(p) == 0 || p[0] == nil {
		return nil, errors.New("got nil for gke firewall rule config")
	}

	out := make([]*infrapb.Rule, len(p))
	for i := range p {
		obj := &infrapb.Rule{}
		in := p[i].(map[string]interface{})

		if v, ok := in["protocol"].(string); ok && len(v) > 0 {
			obj.Protocol = v
		}

		if v, ok := in["ports"].([]interface{}); ok && len(v) > 0 {
			obj.Ports = toArrayString(v)
		}
		out[i] = obj
	}

	return out, nil
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

		if v, ok := in["name"].(string); ok && len(v) > 0 {
			obj.Name = v
		}

		if v, ok := in["cidr"].(string); ok && len(v) > 0 {
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

		if v, ok := in["name"].(string); ok && len(v) > 0 {
			obj.Name = v
		}

		if v, ok := in["node_version"].(string); ok && len(v) > 0 {
			obj.NodeVersion = v
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

		// GkeNodeMetadata
		if v, ok := in["metadata"].([]interface{}); ok && len(v) > 0 {
			obj.Metadata, err = expandToV3GkeNodeMetaData(v)
			if err != nil {
				return nil, fmt.Errorf("failed to expand Gke authorized networks " + err.Error())
			}
		}

		// GkeManagement
		if v, ok := in["management"].([]interface{}); ok && len(v) > 0 {
			obj.Management, err = expandToV3GkeNodeManagement(v)
			if err != nil {
				return nil, fmt.Errorf("failed to expand Gke node management " + err.Error())
			}
		}

		// GkeUpgradeSettings
		if v, ok := in["upgrade_settings"].([]interface{}); ok && len(v) > 0 {
			obj.UpgradeSettings, err = expandToV3GkeNodeUpgradeSettings(v)
			if err != nil {
				return nil, fmt.Errorf("failed to expand Gke node upgrade settings " + err.Error())
			}
		}

		out[i] = obj
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
	var err error
	if len(p) == 0 || p[0] == nil {
		return nil, errors.New("got nil for gke node machine config")
	}

	obj := &infrapb.GkeNodeMachineConfig{}
	in := p[0].(map[string]interface{})

	if v, ok := in["machine_type"].(string); ok && len(v) > 0 {
		obj.MachineType = v
	}

	if v, ok := in["image_type"].(string); ok && len(v) > 0 {
		obj.ImageType = v
	}

	if v, ok := in["boot_disk_type"].(string); ok && len(v) > 0 {
		obj.BootDiskType = v
	}

	if v, ok := in["boot_disk_size"].(int); ok && v > 0 {
		obj.BootDiskSize = int64(v)
	}

	// GkeNodeReservationAffinity
	if v, ok := in["reservation_affinity"].([]interface{}); ok && len(v) > 0 {
		obj.ReservationAffinity, err = expandToV3GkeNodeReservationAffinity(v)
		if err != nil {
			return nil, fmt.Errorf("failed to expand Gke reservation affinity " + err.Error())
		}
	}

	// GkeNodeAccelerators
	if v, ok := in["accelerators"].([]interface{}); ok && len(v) > 0 {
		obj.Accelerators, err = expandToV3GkeNodeAccelerators(v)
		if err != nil {
			return nil, fmt.Errorf("failed to expand Gke node accelerators " + err.Error())
		}
	}

	return obj, nil
}

// GkeNodeReservationAffinity
func expandToV3GkeNodeReservationAffinity(p []interface{}) (*infrapb.GkeNodeReservationAffinity, error) {
	if len(p) == 0 || p[0] == nil {
		return nil, errors.New("got nil for gke node reservation affinity")
	}

	obj := &infrapb.GkeNodeReservationAffinity{}
	in := p[0].(map[string]interface{})

	if v, ok := in["consume_reservation_type"].(string); ok && len(v) > 0 {
		obj.ConsumeReservationType = v
	}
	if v, ok := in["reservation_name"].(string); ok && len(v) > 0 {
		obj.ReservationName = v
	}
	return obj, nil
}

func expandToV3GkeNodeAccelerators(p []interface{}) ([]*infrapb.GkeNodeAccelerator, error) {
	var err error
	if len(p) == 0 || p[0] == nil {
		return nil, errors.New("got nil for gke node accelerators config")
	}

	out := make([]*infrapb.GkeNodeAccelerator, len(p))
	for i := range p {
		obj := &infrapb.GkeNodeAccelerator{}

		in := p[i].(map[string]interface{})

		if v, ok := in["type"].(string); ok && len(v) > 0 {
			obj.Type = v
		}

		if v, ok := in["count"].(int); ok && v > 0 {
			obj.Count = int64(v)
		}

		if v, ok := in["accelerator_sharing"].([]interface{}); ok && len(v) > 0 {
			obj.AcceleratorSharing, err = expandToV3GkeNodeAcceleratorSharing(v)
			if err != nil {
				return nil, err
			}
		}

		if v, ok := in["gpu_driver_installation"].([]interface{}); ok && len(v) > 0 {
			obj.GpuDriverInstallation, err = expandToV3GkeNodeGpuDriverInstallation(v)
			if err != nil {
				return nil, err
			}
		}

		if v, ok := in["gpu_partition_size"].(string); ok && len(v) > 0 {
			obj.GpuPartitionSize = v
		}

		out[i] = obj
	}

	return out, nil
}

func expandToV3GkeNodeGpuDriverInstallation(p []interface{}) (*infrapb.GPUDriverInstallation, error) {
	if len(p) == 0 || p[0] == nil {
		return nil, errors.New("got nil for gke node gpu driver installation config")
	}

	obj := &infrapb.GPUDriverInstallation{}
	in := p[0].(map[string]interface{})

	if v, ok := in["type"].(string); ok {
		obj.Type = v
	}

	var err error
	if strings.EqualFold(obj.Type, GKE_ACCELERATOR_GOOGLE_DRIVER_INSTALATION) {
		if v, ok := in["config"].([]interface{}); ok && len(v) > 0 {
			obj.Config, err = expandToV3GoogleDriverInstallation(v)
			if err != nil {
				return nil, fmt.Errorf("failed to expand Google driver installation settings" + err.Error())
			}
		}
	}
	if strings.EqualFold(obj.Type, GKE_ACCELERATOR_USER_DRIVER_INSTALATION) {
		if v, ok := in["config"].([]interface{}); ok && len(v) > 0 {
			obj.Config, err = expandToV3UserDriverInstallation(v)
			if err != nil {
				return nil, fmt.Errorf("failed to expand User driver installation settings" + err.Error())
			}
		}
	}

	return obj, nil
}

func expandToV3UserDriverInstallation(p []interface{}) (*infrapb.GPUDriverInstallation_UserManaged, error) {
	if len(p) == 0 || p[0] == nil {
		return nil, errors.New("got nil for gke node gpu driver installation config")
	}

	obj := &infrapb.GPUDriverInstallation_UserManaged{}
	// in := p[0].(map[string]interface{})

	return obj, nil
}

func expandToV3GoogleDriverInstallation(p []interface{}) (*infrapb.GPUDriverInstallation_GoogleManaged, error) {
	if len(p) == 0 || p[0] == nil {
		return nil, errors.New("got nil for gke node gpu driver installation config")
	}

	obj := &infrapb.GPUDriverInstallation_GoogleManaged{
		GoogleManaged: &infrapb.GPUGoogleManagedDriverInstallation{},
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["version"].(string); ok {
		obj.GoogleManaged.Version = v
	}

	return obj, nil
}

func expandToV3GkeNodeAcceleratorSharing(p []interface{}) (*infrapb.GkeAcceleratorSharing, error) {
	if len(p) == 0 || p[0] == nil {
		return nil, errors.New("got nil for gke node accelerator sharing config")
	}

	obj := &infrapb.GkeAcceleratorSharing{}
	in := p[0].(map[string]interface{})

	if v, ok := in["max_shared_clients"].(int); ok {
		obj.MaxSharedClients = int64(v)
	}

	if v, ok := in["strategy"].(string); ok {
		obj.Strategy = v
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

func expandToV3GkeNodeMetaData(p []interface{}) (*infrapb.GkeNodeMetadata, error) {
	if len(p) == 0 || p[0] == nil {
		return nil, errors.New("got nil for gke node metadata config")
	}

	obj := &infrapb.GkeNodeMetadata{}
	in := p[0].(map[string]interface{})

	var err error

	if v, ok := in["kubernetes_labels"].([]interface{}); ok && len(v) > 0 {
		obj.KubernetesLabels, err = expandToV3GkeKubernetesLabels(v)
		if err != nil {
			return nil, fmt.Errorf("failed to expand Gke kubernetes labels " + err.Error())
		}
	}

	if v, ok := in["gce_instance_metadata"].([]interface{}); ok && len(v) > 0 {
		obj.GceInstanceMetadata, err = expandToV3GkeGceInstanceMetadata(v)
		if err != nil {
			return nil, fmt.Errorf("failed to expand Gke gce instance metadata " + err.Error())
		}
	}

	if v, ok := in["node_taints"].([]interface{}); ok && len(v) > 0 {
		obj.NodeTaints, err = expandToV3GkeNodeTaints(v)
		if err != nil {
			return nil, fmt.Errorf("failed to expand Gke node taints " + err.Error())
		}
	}

	return obj, nil
}

func expandToV3GkeNodeManagement(p []interface{}) (*infrapb.GkeNodeManagement, error) {
	if len(p) == 0 || p[0] == nil {
		return nil, errors.New("got nil for gke node management config")
	}

	obj := &infrapb.GkeNodeManagement{}
	in := p[0].(map[string]interface{})

	if v, ok := in["auto_upgrade"].(bool); ok {
		obj.AutoUpgrade = v
	}

	return obj, nil
}

func expandToV3GkeNodeUpgradeSettings(p []interface{}) (*infrapb.GkeNodeUpgradeSettings, error) {
	if len(p) == 0 || p[0] == nil {
		return nil, errors.New("got nil for gke node upgrade settings config")
	}

	obj := &infrapb.GkeNodeUpgradeSettings{}
	in := p[0].(map[string]interface{})

	if v, ok := in["strategy"].(string); ok {
		obj.Strategy = v
	}

	var err error
	if strings.EqualFold(obj.Strategy, GKE_NODEPOOL_UPGRADE_STRATEGY_SURGE) {
		if v, ok := in["config"].([]interface{}); ok && len(v) > 0 {
			obj.Config, err = expandToV3GkeNodeSurgeSettings(v)
			if err != nil {
				return nil, fmt.Errorf("failed to expand Gke Node surge settings " + err.Error())
			}
		}
	}
	if strings.EqualFold(obj.Strategy, GKE_NODEPOOL_UPGRADE_STRATEGY_BLUE_GREEN) {
		if v, ok := in["config"].([]interface{}); ok && len(v) > 0 {
			obj.Config, err = expandToV3GkeNodeBlueGreenSettings(v)
			if err != nil {
				return nil, fmt.Errorf("failed to expand Gke Node blue green settings " + err.Error())
			}
		}
	}

	return obj, nil
}

func expandToV3GkeNodeSurgeSettings(p []interface{}) (*infrapb.GkeNodeUpgradeSettings_Surge, error) {
	if len(p) == 0 || p[0] == nil {
		return nil, errors.New("got nil for gke node surge settings config")
	}

	obj := &infrapb.GkeNodeUpgradeSettings_Surge{
		Surge: &infrapb.GkeNodeSurgeSettings{},
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["max_surge"].(int); ok {
		obj.Surge.MaxSurge = int64(v)
	}

	if v, ok := in["max_unavailable"].(int); ok {
		obj.Surge.MaxUnavailable = int64(v)
	}

	return obj, nil
}

func expandToV3GkeNodeBlueGreenSettings(p []interface{}) (*infrapb.GkeNodeUpgradeSettings_BlueGreen, error) {
	if len(p) == 0 || p[0] == nil {
		return nil, errors.New("got nil for gke blue green settings config")
	}

	obj := &infrapb.GkeNodeUpgradeSettings_BlueGreen{
		BlueGreen: &infrapb.GkeNodeBlueGreenSettings{},
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["batch_node_count"].(int); ok {
		obj.BlueGreen.BatchNodeCount = int64(v)
	}

	if v, ok := in["batch_soak_duration"].(string); ok {
		obj.BlueGreen.BatchSoakDuration = v
	}
	if v, ok := in["node_pool_soak_duration"].(string); ok {
		obj.BlueGreen.NodePoolSoakDuration = v
	}
	return obj, nil
}

func expandToV3GkeKubernetesLabels(p []interface{}) ([]*infrapb.GkeKubernetesLabel, error) {
	if len(p) == 0 || p[0] == nil {
		return nil, errors.New("got nil for gke kubernetes lables config")
	}

	out := make([]*infrapb.GkeKubernetesLabel, len(p))
	for i := range p {
		obj := &infrapb.GkeKubernetesLabel{}

		in := p[i].(map[string]interface{})

		if v, ok := in["key"].(string); ok && len(v) > 0 {
			obj.Key = v
		}

		if v, ok := in["value"].(string); ok && len(v) > 0 {
			obj.Value = v
		}

		out[i] = obj
	}

	return out, nil
}

func expandToV3GkeGceInstanceMetadata(p []interface{}) ([]*infrapb.GkeGCEInstanceMetadata, error) {
	if len(p) == 0 || p[0] == nil {
		return nil, errors.New("got nil for gke gce instance metadata config")
	}

	out := make([]*infrapb.GkeGCEInstanceMetadata, len(p))
	for i := range p {
		obj := &infrapb.GkeGCEInstanceMetadata{}

		in := p[i].(map[string]interface{})

		log.Println("In expandToV3GkeGceInstanceMetadata ", in)

		if v, ok := in["key"].(string); ok && len(v) > 0 {
			obj.Key = v
		}

		if v, ok := in["value"].(string); ok && len(v) > 0 {
			obj.Value = v
		}

		out[i] = obj
	}

	return out, nil
}

func expandToV3GkeNodeTaints(p []interface{}) ([]*infrapb.GkeNodeTaint, error) {
	if len(p) == 0 || p[0] == nil {
		return nil, errors.New("got nil for gke node taints config")
	}

	out := make([]*infrapb.GkeNodeTaint, len(p))
	for i := range p {
		obj := &infrapb.GkeNodeTaint{}

		in := p[i].(map[string]interface{})

		if v, ok := in["key"].(string); ok && len(v) > 0 {
			obj.Key = v
		}

		if v, ok := in["value"].(string); ok && len(v) > 0 {
			obj.Value = v
		}

		if v, ok := in["effect"].(string); ok && len(v) > 0 {
			obj.Effect = v
		}

		out[i] = obj
	}

	return out, nil
}
