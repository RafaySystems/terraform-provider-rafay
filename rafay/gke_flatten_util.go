package rafay

import (
	"fmt"
	"log"

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
	log.Println("flattenGKEClusterV3 in", in)
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

	// fmt.Printf("flattenGKEClusterV3 flattenMetadataV3 returned %+v\n", ret1)
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

	fmt.Printf("flattenGKEClusterV3 flattenGKEV3Spec returned %+v\n", ret2)
	err = d.Set("spec", ret2)
	if err != nil {
		return err
	}

	//	fmt.Printf("flattenGKEClusterV3 after d.Set, d.GetSpec %+v\n", d.Get("spec"))
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

	if in.Proxy != nil {
		obj["proxy"] = flattenGkeV3Proxy(in.Proxy)
	}

	if in.GetGke() != nil {
		v, ok := obj["config"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["config"] = flattenGKEV3Config(in.GetGke(), v)
	}

	return []interface{}{obj}
}

func flattenGkeV3Proxy(in *infrapb.ClusterProxy) []interface{} {
	if in == nil {
		return nil
	}

	obj := map[string]interface{}{}

	obj["enabled"] = in.Enabled
	obj["http_proxy"] = in.HttpProxy
	obj["https_proxy"] = in.HttpsProxy
	obj["no_proxy"] = in.NoProxy
	obj["proxy_auth"] = in.ProxyAuth
	obj["allow_insecure_bootstrap"] = in.AllowInsecureBootstrap
	obj["bootstrap_ca"] = in.BootstrapCA

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

	//	log.Println("flattenGKEV3Config len of prebootstrapcommands", len(in.PreBootstrapCommands))
	if in.PreBootstrapCommands != nil && len(in.PreBootstrapCommands) > 0 {
		//		log.Println("flattenGKEV3Config populating prebootstrapcommands")
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
		v, ok := obj["security"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["security"] = flattenGKEV3Security(in.Security, v)
	}

	if in.Features != nil {
		v, ok := obj["features"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["features"] = flattenGKEV3Features(in.Features, v)
	}
	fmt.Printf("flattenGKEV3Config complete %+v\n", obj)

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

	if in.GetRegional() != nil {
		v, ok := obj["config"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["config"] = flattenGKEV3RegionalConfig(in.GetRegional(), v)
	} else if in.GetZonal() != nil {
		v, ok := obj["config"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["config"] = flattenGKEV3ZonalConfig(in.GetZonal(), v)
	}

	return []interface{}{obj}
}

func flattenGKEV3RegionalConfig(in *infrapb.GkeRegionalCluster, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	obj["region"] = in.Region
	obj["zone"] = in.Zone

	return []interface{}{obj}
}

func flattenGKEV3ZonalConfig(in *infrapb.GkeZonalCluster, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}

	obj["zone"] = in.Zone

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

	obj["enabled"] = in.Enabled

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
	obj["pod_secondary_range_name"] = in.PodSecondaryRangeName
	obj["service_secondary_range_name"] = in.ServiceSecondaryRangeName

	obj["max_pods_per_node"] = in.MaxPodsPerNode

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

	if in.GetPrivate() != nil {
		v, ok := obj["config"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["config"] = flattenGKEV3PrivateCluster(in.GetPrivate(), v)
	}
	// else if in.GetPublic() != nil {
	// TODO in future when we have Public cluster specific config. Today this is empty
	//}

	return []interface{}{obj}
}

func flattenGKEV3PrivateCluster(in *infrapb.GkePrivateCluster, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.ControlPlaneIPRange) > 0 {
		obj["control_plane_ip_range"] = in.ControlPlaneIPRange
	}

	obj["enable_access_control_plane_external_ip"] = in.EnableAccessControlPlaneExternalIP
	obj["enable_access_control_plane_global"] = in.EnableAccessControlPlaneGlobal
	obj["disable_snat"] = in.DisableSNAT

	if in.FirewallRules != nil && len(in.FirewallRules) > 0 {
		v, ok := obj["firewall_rules"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["firewall_rules"] = flattenGKEV3Firewalls(in.FirewallRules, v)
	}

	return []interface{}{}
}

func flattenGKEV3FirewallRules(in []*infrapb.Rule, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	out := make([]interface{}, len(in))
	for i, j := range in {
		obj := map[string]interface{}{}
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}

		obj["protocol"] = j.Protocol
		obj["ports"] = toArrayInterface(j.Ports)

		out[i] = &obj
	}
	return out
}

func flattenGKEV3Firewalls(in []*infrapb.FirewallRule, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	out := make([]interface{}, len(in))
	for i, j := range in {
		obj := map[string]interface{}{}
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}

		obj["action"] = j.Action
		obj["description"] = j.Description
		obj["destination_ranges"] = j.DestinationRanges
		obj["direction"] = j.Direction
		obj["name"] = j.Name
		obj["network"] = j.Network
		obj["priority"] = j.Priority
		obj["source_ranges"] = j.SourceRanges
		obj["target_tags"] = j.TargetTags
		if j.Rules != nil && len(j.Rules) > 0 {
			v, ok := obj["rules"].([]interface{})
			if !ok {
				v = []interface{}{}
			}
			obj["rules"] = flattenGKEV3FirewallRules(j.Rules, v)
		}

		out[i] = &obj
	}

	return out
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
	for i, j := range in {
		obj := map[string]interface{}{}
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}

		obj["name"] = j.Name
		obj["cidr"] = j.Cidr

		out[i] = &obj
	}

	return out
}

func flattenGKEV3Security(in *infrapb.GkeSecurity, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	obj["enable_workload_identity"] = in.EnableWorkloadIdentity
	obj["enable_google_groups_for_rbac"] = in.EnableGoogleGroupsForRbac
	if in.EnableGoogleGroupsForRbac {
		obj["security_group"] = in.SecurityGroup
	}
	obj["enable_legacy_authorization"] = in.EnableLegacyAuthorization
	obj["issue_client_certificate"] = in.IssueClientCertificate

	return []interface{}{obj}
}

func flattenGKEV3Features(in *infrapb.GkeFeatures, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	obj["enable_cloud_logging"] = in.EnableCloudLogging
	if in.CloudLoggingComponents != nil && len(in.CloudLoggingComponents) > 0 {
		obj["cloud_logging_components"] = toArrayInterface(in.CloudLoggingComponents)
	}

	obj["enable_cloud_monitoring"] = in.EnableCloudMonitoring
	if in.CloudMonitoringComponents != nil && len(in.CloudMonitoringComponents) > 0 {
		obj["cloud_monitoring_components"] = toArrayInterface(in.CloudMonitoringComponents)
	}

	obj["enable_managed_service_prometheus"] = in.EnableManagedServicePrometheus
	obj["enable_application_manager_beta"] = in.EnableApplicationManagerBeta
	obj["enable_backup_for_gke"] = in.EnableBackupForGke
	obj["enable_compute_engine_persistent_disk_csi_driver"] = in.EnableComputeEnginePersistentDiskCSIDriver
	obj["enable_filestore_csi_driver"] = in.EnableFilestoreCSIDriver
	obj["enable_image_streaming"] = in.EnableImageStreaming

	return []interface{}{obj}
}

func flattenGKEV3Nodepools(in []*infrapb.GkeNodePool, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	out := make([]interface{}, len(in))
	for i, j := range in {
		obj := map[string]interface{}{}
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}

		obj["name"] = j.Name
		obj["node_version"] = j.NodeVersion
		obj["size"] = j.Size

		if j.NodeLocations != nil {
			v, ok := obj["node_locations"].([]interface{})
			if !ok {
				v = []interface{}{}
			}
			obj["node_locations"] = flattenGKEV3NodeLocations(j.NodeLocations, v)
		}

		if j.AutoScaling != nil {
			v, ok := obj["auto_scaling"].([]interface{})
			if !ok {
				v = []interface{}{}
			}
			obj["auto_scaling"] = flattenGKEV3NodePoolAutoScaling(j.AutoScaling, v)
		}

		if j.MachineConfig != nil {
			v, ok := obj["machine_config"].([]interface{})
			if !ok {
				v = []interface{}{}
			}
			obj["machine_config"] = flattenGKEV3NodeMachineConfig(j.MachineConfig, v)
		}

		if j.Networking != nil {
			v, ok := obj["networking"].([]interface{})
			if !ok {
				v = []interface{}{}
			}
			obj["networking"] = flattenGKEV3NodeNetworking(j.Networking, v)
		}

		if j.Security != nil {
			v, ok := obj["security"].([]interface{})
			if !ok {
				v = []interface{}{}
			}
			obj["security"] = flattenGKEV3NodeSecurity(j.Security, v)

		}

		if j.Metadata != nil {
			v, ok := obj["metadata"].([]interface{})
			if !ok {
				v = []interface{}{}
			}
			obj["metadata"] = flattenGKEV3NodeMetadata(j.Metadata, v)
		}

		if j.Management != nil {
			v, ok := obj["management"].([]interface{})
			if !ok {
				v = []interface{}{}
			}
			obj["management"] = flattenGKEV3NodeManagement(j.Management, v)
		}

		if j.UpgradeSettings != nil {
			v, ok := obj["upgrade_settings"].([]interface{})
			if !ok {
				v = []interface{}{}
			}
			obj["upgrade_settings"] = flattenGKEV3NodeUpgradeSettings(j.UpgradeSettings, v)
		}

		out[i] = &obj
	}

	return out
}

func flattenGKEV3NodeLocations(in *infrapb.GkeNodeLocation, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	obj["enabled"] = in.Enabled
	if in.Zones != nil && len(in.Zones) > 0 {
		obj["zones"] = toArrayInterface(in.Zones)
	}

	return []interface{}{obj}
}

func flattenGKEV3NodePoolAutoScaling(in *infrapb.GkeNodeAutoScale, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	obj["min_nodes"] = in.MinNodes
	obj["max_nodes"] = in.MaxNodes

	return []interface{}{obj}
}

func flattenGKEV3NodeMachineConfig(in *infrapb.GkeNodeMachineConfig, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	obj["image_type"] = in.ImageType
	obj["machine_type"] = in.MachineType
	obj["boot_disk_size"] = in.BootDiskSize
	obj["boot_disk_type"] = in.BootDiskType

	if in.ReservationAffinity != nil {
		v, ok := obj["reservation_affinity"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["reservation_affinity"] = flattenGKEV3NodeReservationAffinity(in.GetReservationAffinity(), v)
	}

	if in.Accelerators != nil && len(in.Accelerators) > 0 {
		v, ok := obj["accelerators"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["accelerators"] = flattenGKEV3NodeAccelerators(in.Accelerators, v)
	}

	return []interface{}{obj}
}

func flattenGKEV3NodeAccelerators(in []*infrapb.GkeNodeAccelerator, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	out := make([]interface{}, len(in))
	for i, j := range in {
		obj := map[string]interface{}{}
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}

		obj["type"] = j.Type
		obj["count"] = j.Count
		obj["gpu_partition_size"] = j.GpuPartitionSize

		if j.AcceleratorSharing != nil {
			v, ok := obj["accelerator_sharing"].([]interface{})
			if !ok {
				v = []interface{}{}
			}
			obj["accelerator_sharing"] = flattenGKEV3NodeAcceleratorSharing(j.AcceleratorSharing, v)
		}

		if j.GpuDriverInstallation != nil {
			v, ok := obj["gpu_driver_installation"].([]interface{})
			if !ok {
				v = []interface{}{}
			}
			obj["gpu_driver_installation"] = flattenGKEV3NodeGpuDriverInstallation(j.GpuDriverInstallation, v)
		}
	}

	return out
}

func flattenGKEV3NodeGpuDriverInstallation(in *infrapb.GPUDriverInstallation, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}
	obj["type"] = in.Type

	if in.GetGoogleManaged() != nil {
		v, ok := obj["config"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["config"] = flattenGKEV3NodeGpuDriverGoogleInstallationConfig(in.GetGoogleManaged(), v)
	} else if in.GetUserManaged() != nil {
		v, ok := obj["config"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["config"] = flattenGKEV3NodeGpuDriverUserInstallationConfig(in.GetUserManaged(), v)
	}

	return []interface{}{obj}
}

func flattenGKEV3NodeGpuDriverGoogleInstallationConfig(in *infrapb.GPUGoogleManagedDriverInstallation, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}
	obj["version"] = in.Version

	return []interface{}{obj}
}

func flattenGKEV3NodeGpuDriverUserInstallationConfig(in *infrapb.GPUUserManagedDriverInstallation, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	// obj := map[string]interface{}{}
	// if len(p) != 0 && p[0] != nil {
	// 	obj = p[0].(map[string]interface{})
	// }
	return nil
}

func flattenGKEV3NodeAcceleratorSharing(in *infrapb.GkeAcceleratorSharing, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}
	obj["max_shared_clients"] = in.MaxSharedClients
	obj["strategy"] = in.Strategy

	return []interface{}{obj}
}

func flattenGKEV3NodeReservationAffinity(in *infrapb.GkeNodeReservationAffinity, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	obj["consume_reservation_type"] = in.ConsumeReservationType
	obj["reservation_name"] = in.ReservationName

	return []interface{}{obj}
}

func flattenGKEV3NodeNetworking(in *infrapb.GkeNodeNetworking, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	obj["max_pods_per_node"] = in.MaxPodsPerNode

	if in.NetworkTags != nil && len(in.NetworkTags) > 0 {
		obj["network_tags"] = toArrayInterface(in.NetworkTags)
	}

	return []interface{}{obj}
}

func flattenGKEV3NodeSecurity(in *infrapb.GkeNodeSecurity, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	obj["enable_secure_boot"] = in.EnableSecureBoot
	obj["enable_integrity_monitoring"] = in.EnableIntegrityMonitoring

	return []interface{}{obj}
}

func flattenGKEV3NodeMetadata(in *infrapb.GkeNodeMetadata, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if in.KubernetesLabels != nil && len(in.KubernetesLabels) > 0 {
		v, ok := obj["kubernetes_labels"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["kubernetes_labels"] = flattenGKEV3KubernetesLabels(in.KubernetesLabels, v)
	}

	if in.NodeTaints != nil && len(in.NodeTaints) > 0 {
		v, ok := obj["node_taints"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["node_taints"] = flattenGKEV3NodeTaints(in.NodeTaints, v)
	}

	if in.GceInstanceMetadata != nil && len(in.GceInstanceMetadata) > 0 {
		v, ok := obj["gce_instance_metadata"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["gce_instance_metadata"] = flattenGKEV3GceInstanceMetadata(in.GceInstanceMetadata, v)
	}

	return []interface{}{obj}
}

func flattenGKEV3NodeTaints(in []*infrapb.GkeNodeTaint, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	out := make([]interface{}, len(in))
	for i, j := range in {
		obj := map[string]interface{}{}
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}

		obj["key"] = j.Key
		obj["value"] = j.Value
		obj["effect"] = j.Effect

		out[i] = &obj
	}

	return out
}

func flattenGKEV3KubernetesLabels(in []*infrapb.GkeKubernetesLabel, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	out := make([]interface{}, len(in))
	for i, j := range in {
		obj := map[string]interface{}{}
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}

		obj["key"] = j.Key
		obj["value"] = j.Value

		out[i] = &obj
	}

	return out
}

func flattenGKEV3GceInstanceMetadata(in []*infrapb.GkeGCEInstanceMetadata, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	out := make([]interface{}, len(in))
	for i, j := range in {
		obj := map[string]interface{}{}
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}

		obj["key"] = j.Key
		obj["value"] = j.Value

		out[i] = &obj
	}

	return out
}

func flattenGKEV3NodeManagement(in *infrapb.GkeNodeManagement, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}
	obj["auto_upgrade"] = in.AutoUpgrade

	return []interface{}{obj}
}

func flattenGKEV3NodeUpgradeSettings(in *infrapb.GkeNodeUpgradeSettings, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	obj["strategy"] = in.Strategy

	if in.GetSurge() != nil {
		v, ok := obj["config"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["config"] = flattenGKEV3NodeSurgeSettings(in.GetSurge(), v)
	} else if in.GetBlueGreen() != nil {
		v, ok := obj["config"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["config"] = flattenGKEV3NodeBlueGreenSettings(in.GetBlueGreen(), v)
	}

	return []interface{}{obj}
}

func flattenGKEV3NodeSurgeSettings(in *infrapb.GkeNodeSurgeSettings, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	obj["max_surge"] = in.MaxSurge
	obj["max_unavailable"] = in.MaxUnavailable

	return []interface{}{obj}
}

func flattenGKEV3NodeBlueGreenSettings(in *infrapb.GkeNodeBlueGreenSettings, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	obj["batch_node_count"] = in.BatchNodeCount
	obj["batch_soak_duration"] = in.BatchSoakDuration
	obj["node_pool_soak_duration"] = in.NodePoolSoakDuration

	return []interface{}{obj}
}
