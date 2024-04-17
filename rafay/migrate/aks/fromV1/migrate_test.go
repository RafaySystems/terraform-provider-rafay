package fromV1_test

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/RafaySystems/terraform-provider-rafay/rafay/migrate/aks/fromV1"
)

// spec.0.cluster_config.0.spec.0.managed_cluster.0.properties.0.identity_profile = {}
var preMigrationStateJsonStr string = `
{
	"apiversion": "rafay.io/v1alpha1",
	"id": "rx28oml",
	"kind": "Cluster",
	"metadata": [
	  {
		"labels": {},
		"name": "thiru-tf-2",
		"project": "defaultproject"
	  }
	],
	"spec": [
	  {
		"blueprint": "minimal",
		"blueprintversion": "",
		"cloudprovider": "gautham-azure-creds",
		"cluster_config": [
		  {
			"apiversion": "rafay.io/v1alpha1",
			"kind": "aksClusterConfig",
			"metadata": [
			  {
				"name": "thiru-tf-2"
			  }
			],
			"spec": [
			  {
				"managed_cluster": [
				  {
					"additional_metadata": [],
					"apiversion": "2022-07-01",
					"extended_location": [],
					"identity": [
					  {
						"type": "SystemAssigned",
						"user_assigned_identities": {}
					  }
					],
					"location": "centralindia",
					"properties": [
					  {
						"aad_profile": [],
						"addon_profiles": [],
						"api_server_access_profile": [],
						"auto_scaler_profile": [],
						"auto_upgrade_profile": [],
						"disable_local_accounts": false,
						"disk_encryption_set_id": "",
						"dns_prefix": "testuser-test-dns",
						"enable_pod_security_policy": false,
						"enable_rbac": false,
						"fqdn_subdomain": "",
						"http_proxy_config": [],
						"identity_profile": {},
						"kubernetes_version": "1.28.5",
						"linux_profile": [],
						"network_profile": [
						  {
							"dns_service_ip": "",
							"docker_bridge_cidr": "",
							"load_balancer_profile": [],
							"load_balancer_sku": "Standard",
							"network_mode": "",
							"network_plugin": "kubenet",
							"network_policy": "",
							"outbound_type": "",
							"pod_cidr": "",
							"service_cidr": ""
						  }
						],
						"node_resource_group": "",
						"pod_identity_profile": [],
						"private_link_resources": [],
						"service_principal_profile": [],
						"windows_profile": []
					  }
					],
					"sku": [],
					"tags": {
					  "env": "dev",
					  "user": "thirumal@rafay.co"
					},
					"type": "Microsoft.ContainerService/managedClusters"
				  }
				],
				"node_pools": [
				  {
					"apiversion": "2022-07-01",
					"location": "centralindia",
					"name": "pool1",
					"properties": [
					  {
						"availability_zones": [],
						"count": 1,
						"enable_auto_scaling": true,
						"enable_encryption_at_host": false,
						"enable_fips": false,
						"enable_node_public_ip": false,
						"enable_ultra_ssd": false,
						"gpu_instance_profile": "",
						"kubelet_config": [],
						"kubelet_disk_type": "",
						"linux_os_config": [],
						"max_count": 1,
						"max_pods": 40,
						"min_count": 1,
						"mode": "System",
						"node_labels": {},
						"node_public_ip_prefix_id": "",
						"node_taints": [],
						"orchestrator_version": "1.28.5",
						"os_disk_size_gb": 0,
						"os_disk_type": "",
						"os_sku": "",
						"os_type": "Linux",
						"pod_subnet_id": "",
						"proximity_placement_group_id": "",
						"scale_set_eviction_policy": "Delete",
						"scale_set_priority": "Regular",
						"spot_max_price": 0,
						"tags": {},
						"type": "VirtualMachineScaleSets",
						"upgrade_settings": [],
						"vm_size": "Standard_B2s",
						"vnet_subnet_id": ""
					  }
					],
					"type": "Microsoft.ContainerService/managedClusters/agentPools"
				  }
				],
				"resource_group_name": "gautham-rg-ci",
				"subscription_id": ""
			  }
			]
		  }
		],
		"sharing": [],
		"type": "aks"
	  }
	],
	"timeouts": null
}
`

// spec.0.cluster_config.0.spec.0.managed_cluster.0.properties.0.identity_profile = []
var postMigrationStateJsonStr string = `
{
	"apiversion": "rafay.io/v1alpha1",
	"id": "rx28oml",
	"kind": "Cluster",
	"metadata": [
	  {
		"labels": {},
		"name": "thiru-tf-2",
		"project": "defaultproject"
	  }
	],
	"spec": [
	  {
		"blueprint": "minimal",
		"blueprintversion": "",
		"cloudprovider": "gautham-azure-creds",
		"cluster_config": [
		  {
			"apiversion": "rafay.io/v1alpha1",
			"kind": "aksClusterConfig",
			"metadata": [
			  {
				"name": "thiru-tf-2"
			  }
			],
			"spec": [
			  {
				"managed_cluster": [
				  {
					"additional_metadata": [],
					"apiversion": "2022-07-01",
					"extended_location": [],
					"identity": [
					  {
						"type": "SystemAssigned",
						"user_assigned_identities": {}
					  }
					],
					"location": "centralindia",
					"properties": [
					  {
						"aad_profile": [],
						"addon_profiles": [],
						"api_server_access_profile": [],
						"auto_scaler_profile": [],
						"auto_upgrade_profile": [],
						"disable_local_accounts": false,
						"disk_encryption_set_id": "",
						"dns_prefix": "testuser-test-dns",
						"enable_pod_security_policy": false,
						"enable_rbac": false,
						"fqdn_subdomain": "",
						"http_proxy_config": [],
						"identity_profile": [],
						"kubernetes_version": "1.28.5",
						"linux_profile": [],
						"network_profile": [
						  {
							"dns_service_ip": "",
							"docker_bridge_cidr": "",
							"load_balancer_profile": [],
							"load_balancer_sku": "Standard",
							"network_mode": "",
							"network_plugin": "kubenet",
							"network_policy": "",
							"outbound_type": "",
							"pod_cidr": "",
							"service_cidr": ""
						  }
						],
						"node_resource_group": "",
						"pod_identity_profile": [],
						"private_link_resources": [],
						"service_principal_profile": [],
						"windows_profile": []
					  }
					],
					"sku": [],
					"tags": {
					  "env": "dev",
					  "user": "thirumal@rafay.co"
					},
					"type": "Microsoft.ContainerService/managedClusters"
				  }
				],
				"node_pools": [
				  {
					"apiversion": "2022-07-01",
					"location": "centralindia",
					"name": "pool1",
					"properties": [
					  {
						"availability_zones": [],
						"count": 1,
						"enable_auto_scaling": true,
						"enable_encryption_at_host": false,
						"enable_fips": false,
						"enable_node_public_ip": false,
						"enable_ultra_ssd": false,
						"gpu_instance_profile": "",
						"kubelet_config": [],
						"kubelet_disk_type": "",
						"linux_os_config": [],
						"max_count": 1,
						"max_pods": 40,
						"min_count": 1,
						"mode": "System",
						"node_labels": {},
						"node_public_ip_prefix_id": "",
						"node_taints": [],
						"orchestrator_version": "1.28.5",
						"os_disk_size_gb": 0,
						"os_disk_type": "",
						"os_sku": "",
						"os_type": "Linux",
						"pod_subnet_id": "",
						"proximity_placement_group_id": "",
						"scale_set_eviction_policy": "Delete",
						"scale_set_priority": "Regular",
						"spot_max_price": 0,
						"tags": {},
						"type": "VirtualMachineScaleSets",
						"upgrade_settings": [],
						"vm_size": "Standard_B2s",
						"vnet_subnet_id": ""
					  }
					],
					"type": "Microsoft.ContainerService/managedClusters/agentPools"
				  }
				],
				"resource_group_name": "gautham-rg-ci",
				"subscription_id": ""
			  }
			]
		  }
		],
		"sharing": [],
		"type": "aks"
	  }
	],
	"timeouts": null
}
`

func TestMigrationFromV1(t *testing.T) {
	preMigrationState := make(map[string]interface{}, 0)
	if err := json.Unmarshal([]byte(preMigrationStateJsonStr), &preMigrationState); err != nil {
		t.Fatalf("Failed to unmarshal pre-migration state: %v", err)
	}
	postMigrationState := make(map[string]interface{}, 0)
	if err := json.Unmarshal([]byte(postMigrationStateJsonStr), &postMigrationState); err != nil {
		t.Fatalf("Failed to unmarshal post-migration state: %v", err)
	}
	migratedState, err := fromV1.Migrate(context.TODO(), preMigrationState, nil)
	if err != nil {
		t.Fatalf("Failed to migrate from V1: %v", err)
	}
	if !reflect.DeepEqual(migratedState, postMigrationState) {
		t.Fatalf("Migration failed, expected: %v, got: %v", postMigrationState, migratedState)
	}
}
