package mocks

import (
	"context"

	"github.com/RafaySystems/terraform-provider-rafay/rafay"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// MockTestData contains all the mock data structures for testing
type MockTestData struct {
	EKS MockEKSData
	AKS MockAKSData
}

// MockEKSData contains mock data for EKS cluster testing
type MockEKSData struct {
	ClusterInput       []interface{}
	ClusterConfigInput []interface{}
	MetadataInput      []interface{}
	AccessConfigInput  []interface{}
	VPCInput           []interface{}
	NodeGroupsInput    []interface{}
	IAMInput           []interface{}

	ClusterStruct       *rafay.EKSCluster
	ClusterConfigStruct *rafay.EKSClusterConfig
	MetadataStruct      *rafay.EKSClusterMetadata
	AccessConfigStruct  *rafay.EKSClusterAccess
	VPCStruct           *rafay.EKSClusterVPC
	NodeGroupsStruct    []*rafay.NodeGroup
	IAMStruct           *rafay.EKSClusterIAM
}

// MockAKSData contains mock data for AKS cluster testing
type MockAKSData struct {
	ClusterInput        []interface{}
	MetadataInput       []interface{}
	SpecInput           []interface{}
	ConfigInput         []interface{}
	ManagedClusterInput []interface{}
	NodePoolInput       []interface{}
	MaintenanceInput    []interface{}

	ClusterStruct        *rafay.AKSCluster
	MetadataStruct       *rafay.AKSClusterMetadata
	SpecStruct           *rafay.AKSClusterSpec
	ConfigStruct         *rafay.AKSClusterConfig
	ManagedClusterStruct *rafay.AKSManagedCluster
	NodePoolStruct       []*rafay.AKSNodePool
	MaintenanceStruct    []*rafay.AKSMaintenanceConfig
}

// NewMockTestData creates a new instance with all mock data initialized
func NewMockTestData() *MockTestData {
	return &MockTestData{
		EKS: NewMockEKSData(),
		AKS: NewMockAKSData(),
	}
}

// NewMockEKSData creates mock data for EKS testing
func NewMockEKSData() MockEKSData {
	return MockEKSData{
		ClusterInput: []interface{}{
			map[string]interface{}{
				"kind": "Cluster",
				"metadata": []interface{}{
					map[string]interface{}{
						"name":    "test-eks-cluster",
						"project": "test-project",
						"labels": map[string]interface{}{
							"env":        "test",
							"version":    "1.0",
							"managed-by": "terraform",
						},
					},
				},
				"spec": []interface{}{
					map[string]interface{}{
						"type":                "eks",
						"blueprint":           "minimal",
						"blueprintversion":    "1.0",
						"cloudprovider":       "aws",
						"crossAccountRoleARN": "arn:aws:iam::123456789012:role/test-role",
						"cniprovider":         "aws-cni",
						"proxyconfig": []interface{}{
							map[string]interface{}{
								"http_proxy":               "http://proxy.example.com:8080",
								"https_proxy":              "https://proxy.example.com:8080",
								"no_proxy":                 "localhost,127.0.0.1,.example.com",
								"proxy_auth":               "user:password",
								"allow_insecure_bootstrap": true,
								"enabled":                  true,
							},
						},
					},
				},
			},
		},
		ClusterConfigInput: []interface{}{
			map[string]interface{}{
				"kind":       "ClusterConfig",
				"apiversion": "eksctl.io/v1alpha5",
				"metadata": []interface{}{
					map[string]interface{}{
						"name":    "test-eks-cluster",
						"region":  "us-west-2",
						"version": "1.21",
						"tags": map[string]interface{}{
							"Environment": "test",
							"Team":        "platform",
							"Project":     "terraform-provider-rafay",
						},
						"annotations": map[string]interface{}{
							"rafay.io/managed": "true",
						},
					},
				},
				"kubernetes_network_config": []interface{}{
					map[string]interface{}{
						"ip_family":         "IPv4",
						"service_ipv4_cidr": "10.100.0.0/16",
					},
				},
				"availability_zones": []interface{}{"us-west-2a", "us-west-2b", "us-west-2c"},
			},
		},
		MetadataInput: []interface{}{
			map[string]interface{}{
				"name":    "test-eks-cluster",
				"project": "test-project",
				"labels": map[string]interface{}{
					"env":        "test",
					"version":    "1.0",
					"managed-by": "terraform",
				},
			},
		},
		AccessConfigInput: []interface{}{
			map[string]interface{}{
				"bootstrap_cluster_creator_admin_permissions": true,
				"authentication_mode":                         "API_AND_CONFIG_MAP",
				"access_entries": []interface{}{
					map[string]interface{}{
						"principal_arn": "arn:aws:iam::123456789012:user/test-user",
						"type":          "STANDARD",
						"username":      "test-user",
						"access_policies": []interface{}{
							map[string]interface{}{
								"policy_arn": "arn:aws:eks::aws:cluster-access-policy/AmazonEKSViewPolicy",
								"access_scope": []interface{}{
									map[string]interface{}{
										"type":       "cluster",
										"namespaces": []interface{}{"default", "kube-system"},
									},
								},
							},
						},
					},
				},
			},
		},
		VPCInput: []interface{}{
			map[string]interface{}{
				"id":                   "vpc-12345678",
				"cidr":                 "10.0.0.0/16",
				"enable_dns_hostnames": true,
				"enable_dns_support":   true,
				"subnets": []interface{}{
					map[string]interface{}{
						"private": map[string]interface{}{
							"us-west-2a": map[string]interface{}{
								"id":   "subnet-private-1",
								"cidr": "10.0.1.0/24",
							},
							"us-west-2b": map[string]interface{}{
								"id":   "subnet-private-2",
								"cidr": "10.0.2.0/24",
							},
							"us-west-2c": map[string]interface{}{
								"id":   "subnet-private-3",
								"cidr": "10.0.3.0/24",
							},
						},
						"public": map[string]interface{}{
							"us-west-2a": map[string]interface{}{
								"id":   "subnet-public-1",
								"cidr": "10.0.101.0/24",
							},
							"us-west-2b": map[string]interface{}{
								"id":   "subnet-public-2",
								"cidr": "10.0.102.0/24",
							},
							"us-west-2c": map[string]interface{}{
								"id":   "subnet-public-3",
								"cidr": "10.0.103.0/24",
							},
						},
					},
				},
				"nat": []interface{}{
					map[string]interface{}{
						"gateway": "HighlyAvailable",
					},
				},
				"cluster_endpoints": []interface{}{
					map[string]interface{}{
						"private_access":       true,
						"public_access":        true,
						"public_access_cidrs":  []interface{}{"0.0.0.0/0"},
						"private_access_cidrs": []interface{}{"10.0.0.0/8"},
					},
				},
			},
		},
		NodeGroupsInput: []interface{}{
			map[string]interface{}{
				"name":               "worker-nodes",
				"instance_type":      "m5.large",
				"desired_capacity":   3,
				"min_size":           1,
				"max_size":           5,
				"volume_size":        50,
				"volume_type":        "gp3",
				"volume_encrypted":   true,
				"ami_family":         "AmazonLinux2",
				"availability_zones": []interface{}{"us-west-2a", "us-west-2b"},
				"private_networking": true,
				"ssh": []interface{}{
					map[string]interface{}{
						"allow":                     true,
						"public_key_name":           "my-key-pair",
						"source_security_group_ids": []interface{}{"sg-12345678"},
						"enable_ssm":                true,
					},
				},
				"iam": []interface{}{
					map[string]interface{}{
						"attach_policy_arns": []interface{}{
							"arn:aws:iam::aws:policy/AmazonEKSWorkerNodePolicy",
							"arn:aws:iam::aws:policy/AmazonEKS_CNI_Policy",
							"arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly",
						},
						"instance_profile_arn": "arn:aws:iam::123456789012:instance-profile/NodeInstanceProfile",
						"instance_role_arn":    "arn:aws:iam::123456789012:role/NodeInstanceRole",
					},
				},
				"security_groups": []interface{}{
					map[string]interface{}{
						"attach_ids": []interface{}{"sg-87654321"},
					},
				},
				"tags": map[string]interface{}{
					"Environment": "test",
					"NodeGroup":   "worker-nodes",
					"Team":        "platform",
				},
				"labels": map[string]interface{}{
					"nodegroup-type": "worker",
					"workload-type":  "general",
				},
			},
		},
		IAMInput: []interface{}{
			map[string]interface{}{
				"with_oidc": true,
				"service_accounts_metadata": []interface{}{
					map[string]interface{}{
						"name":      "cluster-service-accounts",
						"namespace": "kube-system",
					},
				},
				"service_accounts": []interface{}{
					map[string]interface{}{
						"metadata": []interface{}{
							map[string]interface{}{
								"name":      "aws-load-balancer-controller",
								"namespace": "kube-system",
								"labels": map[string]interface{}{
									"app.kubernetes.io/name":      "aws-load-balancer-controller",
									"app.kubernetes.io/component": "controller",
								},
								"annotations": map[string]interface{}{
									"eks.amazonaws.com/role-arn": "arn:aws:iam::123456789012:role/AWSLoadBalancerControllerRole",
								},
							},
						},
						"attach_policy_arns": []interface{}{
							"arn:aws:iam::123456789012:policy/AWSLoadBalancerControllerIAMPolicy",
						},
						"attach_policy": []interface{}{
							map[string]interface{}{
								"statement": []interface{}{
									map[string]interface{}{
										"effect": "Allow",
										"action": []interface{}{
											"ec2:DescribeVpcs",
											"ec2:DescribeSubnets",
											"ec2:DescribeSecurityGroups",
										},
										"resource": "*",
									},
								},
							},
						},
						"well_known_policies": []interface{}{
							map[string]interface{}{
								"auto_scaler":                  true,
								"aws_load_balancer_controller": true,
								"cert_manager":                 false,
								"cluster_autoscaler":           true,
							},
						},
						"permissions_boundary": "arn:aws:iam::123456789012:policy/PermissionsBoundary",
						"role_only":            false,
						"tags": map[string]interface{}{
							"Environment": "test",
							"Component":   "load-balancer-controller",
						},
					},
				},
				"pod_identity_associations": []interface{}{
					map[string]interface{}{
						"namespace":       "default",
						"service_account": "my-service-account",
						"role_arn":        "arn:aws:iam::123456789012:role/my-pod-role",
						"permission_policy": `{
							"Version": "2012-10-17",
							"Statement": [
								{
									"Effect": "Allow",
									"Action": [
										"s3:GetObject",
										"s3:PutObject"
									],
									"Resource": "arn:aws:s3:::my-bucket/*"
								}
							]
						}`,
						"tags": map[string]interface{}{
							"Environment": "test",
							"Component":   "pod-identity",
						},
					},
				},
			},
		},

		// Struct representations
		ClusterStruct: &rafay.EKSCluster{
			Kind: "Cluster",
			Metadata: &rafay.EKSClusterMetadata{
				Name:    "test-eks-cluster",
				Project: "test-project",
				Labels: map[string]string{
					"env":        "test",
					"version":    "1.0",
					"managed-by": "terraform",
				},
			},
			Spec: &rafay.EKSSpec{
				Type:                "eks",
				Blueprint:           "minimal",
				BlueprintVersion:    "1.0",
				CloudProvider:       "aws",
				CrossAccountRoleArn: "arn:aws:iam::123456789012:role/test-role",
				CniProvider:         "aws-cni",
				ProxyConfig: &rafay.ProxyConfig{
					HttpProxy:              "http://proxy.example.com:8080",
					HttpsProxy:             "https://proxy.example.com:8080",
					NoProxy:                "localhost,127.0.0.1,.example.com",
					ProxyAuth:              "user:password",
					AllowInsecureBootstrap: true,
					Enabled:                true,
				},
			},
		},
		ClusterConfigStruct: &rafay.EKSClusterConfig{
			Kind:       "ClusterConfig",
			APIVersion: "eksctl.io/v1alpha5",
			Metadata: &rafay.EKSClusterConfigMetadata{
				Name:    "test-eks-cluster",
				Region:  "us-west-2",
				Version: "1.21",
				Tags: map[string]string{
					"Environment": "test",
					"Team":        "platform",
					"Project":     "terraform-provider-rafay",
				},
				Annotations: map[string]string{
					"rafay.io/managed": "true",
				},
			},
			KubernetesNetworkConfig: &rafay.KubernetesNetworkConfig{
				IPFamily:        "IPv4",
				ServiceIPv4CIDR: "10.100.0.0/16",
			},
			AvailabilityZones: []string{"us-west-2a", "us-west-2b", "us-west-2c"},
		},
		MetadataStruct: &rafay.EKSClusterMetadata{
			Name:    "test-eks-cluster",
			Project: "test-project",
			Labels: map[string]string{
				"env":        "test",
				"version":    "1.0",
				"managed-by": "terraform",
			},
		},
		AccessConfigStruct: &rafay.EKSClusterAccess{
			BootstrapClusterCreatorAdminPermissions: true,
			AuthenticationMode:                      "API_AND_CONFIG_MAP",
			AccessEntries: []*rafay.EKSAccessEntry{
				{
					PrincipalARN: "arn:aws:iam::123456789012:user/test-user",
					Type:         "STANDARD",
					AccessPolicies: []*rafay.EKSAccessPolicy{
						{
							PolicyARN: "arn:aws:eks::aws:cluster-access-policy/AmazonEKSViewPolicy",
							AccessScope: &rafay.EKSAccessScope{
								Type:       "cluster",
								Namespaces: []string{"default", "kube-system"},
							},
						},
					},
				},
			},
		},
		VPCStruct: &rafay.EKSClusterVPC{
			ID:   "vpc-12345678",
			CIDR: "10.0.0.0/16",
			Subnets: &rafay.ClusterSubnets{
				Private: rafay.AZSubnetMapping{
					"us-west-2a": rafay.AZSubnetSpec{ID: "subnet-private-1"},
					"us-west-2b": rafay.AZSubnetSpec{ID: "subnet-private-2"},
					"us-west-2c": rafay.AZSubnetSpec{ID: "subnet-private-3"},
				},
				Public: rafay.AZSubnetMapping{
					"us-west-2a": rafay.AZSubnetSpec{ID: "subnet-public-1"},
					"us-west-2b": rafay.AZSubnetSpec{ID: "subnet-public-2"},
					"us-west-2c": rafay.AZSubnetSpec{ID: "subnet-public-3"},
				},
			},
			NAT: &rafay.ClusterNAT{
				Gateway: "HighlyAvailable",
			},
			ClusterEndpoints: &rafay.ClusterEndpoints{
				PrivateAccess: &[]bool{true}[0],
				PublicAccess:  &[]bool{true}[0],
			},
		},
		IAMStruct: &rafay.EKSClusterIAM{
			WithOIDC: &[]bool{true}[0],
			ServiceAccounts: []*rafay.EKSClusterIAMServiceAccount{
				{
					Metadata: &rafay.EKSClusterIAMMeta{
						Name:      "aws-load-balancer-controller",
						Namespace: "kube-system",
						Labels: map[string]string{
							"app.kubernetes.io/name":      "aws-load-balancer-controller",
							"app.kubernetes.io/component": "controller",
						},
						Annotations: map[string]string{
							"eks.amazonaws.com/role-arn": "arn:aws:iam::123456789012:role/AWSLoadBalancerControllerRole",
						},
					},
					AttachPolicyARNs: []string{
						"arn:aws:iam::123456789012:policy/AWSLoadBalancerControllerIAMPolicy",
					},
					WellKnownPolicies: &rafay.WellKnownPolicies{
						AutoScaler:                &[]bool{true}[0],
						AWSLoadBalancerController: &[]bool{true}[0],
						CertManager:               &[]bool{false}[0],
					},
					PermissionsBoundary: "arn:aws:iam::123456789012:policy/PermissionsBoundary",
					RoleOnly:            &[]bool{false}[0],
					Tags: map[string]string{
						"Environment": "test",
						"Component":   "load-balancer-controller",
					},
				},
			},
		},
	}
}

// NewMockAKSData creates mock data for AKS testing
func NewMockAKSData() MockAKSData {
	return MockAKSData{
		ClusterInput: []interface{}{
			map[string]interface{}{
				"apiversion": "rafay.io/v1alpha5",
				"kind":       "Cluster",
				"metadata": []interface{}{
					map[string]interface{}{
						"name":    "test-aks-cluster",
						"project": "test-project",
						"labels": map[string]interface{}{
							"env":        "test",
							"version":    "1.0",
							"managed-by": "terraform",
						},
					},
				},
				"spec": []interface{}{
					map[string]interface{}{
						"type":             "aks",
						"blueprint":        "minimal",
						"blueprintversion": "1.0",
						"cloudprovider":    "azure",
					},
				},
			},
		},
		MetadataInput: []interface{}{
			map[string]interface{}{
				"name":    "test-aks-cluster",
				"project": "test-project",
				"labels": map[string]interface{}{
					"env":        "test",
					"version":    "1.0",
					"managed-by": "terraform",
				},
			},
		},
		SpecInput: []interface{}{
			map[string]interface{}{
				"type":             "aks",
				"blueprint":        "minimal",
				"blueprintversion": "1.0",
				"cloudprovider":    "azure",
				"cluster_config": []interface{}{
					map[string]interface{}{
						"apiversion": "rafay.io/v1alpha5",
						"kind":       "Cluster",
						"metadata": []interface{}{
							map[string]interface{}{
								"name": "test-aks-cluster",
							},
						},
						"spec": []interface{}{
							map[string]interface{}{
								"subscription_id":     "12345678-1234-1234-1234-123456789012",
								"resource_group_name": "test-rg",
							},
						},
					},
				},
			},
		},
		ConfigInput: []interface{}{
			map[string]interface{}{
				"apiversion": "rafay.io/v1alpha5",
				"kind":       "Cluster",
				"metadata": []interface{}{
					map[string]interface{}{
						"name": "test-aks-cluster",
					},
				},
				"spec": []interface{}{
					map[string]interface{}{
						"subscription_id":     "12345678-1234-1234-1234-123456789012",
						"resource_group_name": "test-rg",
						"managed_cluster": []interface{}{
							map[string]interface{}{
								"location": "East US",
								"tags": map[string]interface{}{
									"Environment": "test",
									"Team":        "platform",
								},
								"identity": []interface{}{
									map[string]interface{}{
										"type": "SystemAssigned",
									},
								},
								"properties": []interface{}{
									map[string]interface{}{
										"dns_prefix":          "test-aks",
										"kubernetes_version":  "1.25.6",
										"node_resource_group": "MC_test-rg_test-aks_eastus",
									},
								},
							},
						},
					},
				},
			},
		},
		ManagedClusterInput: []interface{}{
			map[string]interface{}{
				"location": "East US",
				"tags": map[string]interface{}{
					"Environment": "test",
					"Team":        "platform",
					"Project":     "terraform-provider-rafay",
				},
				"identity": []interface{}{
					map[string]interface{}{
						"type": "SystemAssigned",
					},
				},
				"properties": []interface{}{
					map[string]interface{}{
						"dns_prefix":          "test-aks",
						"kubernetes_version":  "1.25.6",
						"node_resource_group": "MC_test-rg_test-aks_eastus",
						"api_server_access_profile": []interface{}{
							map[string]interface{}{
								"enable_private_cluster":             true,
								"private_dns_zone":                   "system",
								"enable_private_cluster_public_fqdn": false,
								"authorized_ip_ranges":               []interface{}{"10.0.0.0/8"},
							},
						},
						"network_profile": []interface{}{
							map[string]interface{}{
								"network_plugin":     "azure",
								"network_policy":     "azure",
								"dns_service_ip":     "10.0.0.10",
								"service_cidr":       "10.0.0.0/16",
								"docker_bridge_cidr": "172.17.0.1/16",
								"pod_cidr":           "10.244.0.0/16",
								"outbound_type":      "loadBalancer",
							},
						},
						"auto_scaler_profile": []interface{}{
							map[string]interface{}{
								"balance_similar_node_groups":      true,
								"expander":                         "random",
								"max_empty_bulk_delete":            "10",
								"max_graceful_termination_sec":     "600",
								"max_node_provision_time":          "15m",
								"max_total_unready_percentage":     "45",
								"new_pod_scale_up_delay":           "0s",
								"ok_total_unready_count":           "3",
								"scale_down_delay_after_add":       "10m",
								"scale_down_delay_after_delete":    "10s",
								"scale_down_delay_after_failure":   "3m",
								"scale_down_unneeded_time":         "10m",
								"scale_down_utilization_threshold": "0.5",
								"scan_interval":                    "10s",
								"skip_nodes_with_local_storage":    false,
								"skip_nodes_with_system_pods":      true,
							},
						},
					},
				},
			},
		},
		NodePoolInput: []interface{}{
			map[string]interface{}{
				"apiversion": "2022-03-01",
				"name":       "nodepool1",
				"properties": []interface{}{
					map[string]interface{}{
						"count":               3,
						"vm_size":             "Standard_DS2_v2",
						"os_type":             "Linux",
						"type":                "VirtualMachineScaleSets",
						"mode":                "System",
						"max_pods":            30,
						"availability_zones":  []interface{}{"1", "2", "3"},
						"enable_auto_scaling": true,
						"min_count":           1,
						"max_count":           5,
						"os_disk_size_gb":     100,
						"os_disk_type":        "Managed",
						"vnet_subnet_id":      "/subscriptions/12345678-1234-1234-1234-123456789012/resourceGroups/test-rg/providers/Microsoft.Network/virtualNetworks/test-vnet/subnets/test-subnet",
						"pod_subnet_id":       "/subscriptions/12345678-1234-1234-1234-123456789012/resourceGroups/test-rg/providers/Microsoft.Network/virtualNetworks/test-vnet/subnets/pod-subnet",
						"kubelet_config": []interface{}{
							map[string]interface{}{
								"cpu_manager_policy":      "static",
								"cpu_cfs_quota":           true,
								"cpu_cfs_quota_period":    "100ms",
								"image_gc_high_threshold": 85,
								"image_gc_low_threshold":  80,
								"topology_manager_policy": "single-numa-node",
							},
						},
						"linux_os_config": []interface{}{
							map[string]interface{}{
								"transparent_huge_page_enabled": "always",
								"transparent_huge_page_defrag":  "always",
								"swap_file_size_mb":             1024,
								"sysctls": []interface{}{
									map[string]interface{}{
										"vm_max_map_count":            262144,
										"fs_inotify_max_user_watches": 1048576,
										"kernel_threads_max":          515785,
									},
								},
							},
						},
						"upgrade_settings": []interface{}{
							map[string]interface{}{
								"max_surge": "33%",
							},
						},
						"node_labels": map[string]interface{}{
							"nodepool-type": "system",
							"workload-type": "general",
						},
						"node_taints": []interface{}{
							map[string]interface{}{
								"key":    "CriticalAddonsOnly",
								"value":  "true",
								"effect": "NoSchedule",
							},
						},
						"tags": map[string]interface{}{
							"Environment": "test",
							"NodePool":    "nodepool1",
							"Team":        "platform",
						},
					},
				},
			},
		},
		MaintenanceInput: []interface{}{
			map[string]interface{}{
				"name": "aksManagedAutoUpgradeSchedule",
				"properties": []interface{}{
					map[string]interface{}{
						"maintenance_window": []interface{}{
							map[string]interface{}{
								"schedule": []interface{}{
									map[string]interface{}{
										"weekly": []interface{}{
											map[string]interface{}{
												"interval_weeks": 1,
												"day_of_week":    "Sunday",
											},
										},
									},
								},
								"duration_hours": 4,
								"utc_offset":     "+00:00",
								"start_date":     "2023-01-01",
								"start_time":     "01:00",
							},
						},
						"not_allowed_time": []interface{}{
							map[string]interface{}{
								"start": "2023-12-23T00:00:00Z",
								"end":   "2023-12-25T23:59:59Z",
							},
						},
					},
				},
			},
		},

		// Struct representations
		ClusterStruct: &rafay.AKSCluster{
			APIVersion: "rafay.io/v1alpha5",
			Kind:       "Cluster",
			Metadata: &rafay.AKSClusterMetadata{
				Name:    "test-aks-cluster",
				Project: "test-project",
				Labels: map[string]string{
					"env":        "test",
					"version":    "1.0",
					"managed-by": "terraform",
				},
			},
			Spec: &rafay.AKSClusterSpec{
				Type:             "aks",
				Blueprint:        "minimal",
				BlueprintVersion: "1.0",
				CloudProvider:    "azure",
			},
		},
		MetadataStruct: &rafay.AKSClusterMetadata{
			Name:    "test-aks-cluster",
			Project: "test-project",
			Labels: map[string]string{
				"env":        "test",
				"version":    "1.0",
				"managed-by": "terraform",
			},
		},
		SpecStruct: &rafay.AKSClusterSpec{
			Type:             "aks",
			Blueprint:        "minimal",
			BlueprintVersion: "1.0",
			CloudProvider:    "azure",
		},
		ConfigStruct: &rafay.AKSClusterConfig{
			APIVersion: "rafay.io/v1alpha5",
			Kind:       "Cluster",
			Metadata: &rafay.AKSClusterConfigMetadata{
				Name: "test-aks-cluster",
			},
			Spec: &rafay.AKSClusterConfigSpec{
				SubscriptionID:    "12345678-1234-1234-1234-123456789012",
				ResourceGroupName: "test-rg",
			},
		},
		ManagedClusterStruct: &rafay.AKSManagedCluster{
			Location: "East US",
			Tags: map[string]interface{}{
				"Environment": "test",
				"Team":        "platform",
				"Project":     "terraform-provider-rafay",
			},
			Identity: &rafay.AKSManagedClusterIdentity{
				Type: "SystemAssigned",
			},
			Properties: &rafay.AKSManagedClusterProperties{
				DNSPrefix:         "test-aks",
				KubernetesVersion: "1.25.6",
				NodeResourceGroup: "MC_test-rg_test-aks_eastus",
			},
		},
	}
}

// GetCtyValue returns a cty.Value for testing purposes
func GetCtyValue(data map[string]interface{}) cty.Value {
	values := make(map[string]cty.Value)
	for k, v := range data {
		switch val := v.(type) {
		case string:
			values[k] = cty.StringVal(val)
		case bool:
			values[k] = cty.BoolVal(val)
		case int:
			values[k] = cty.NumberIntVal(int64(val))
		case int64:
			values[k] = cty.NumberIntVal(val)
		case float64:
			values[k] = cty.NumberFloatVal(val)
		case []interface{}:
			listVals := make([]cty.Value, len(val))
			for i, item := range val {
				if str, ok := item.(string); ok {
					listVals[i] = cty.StringVal(str)
				} else {
					listVals[i] = cty.NullVal(cty.String)
				}
			}
			values[k] = cty.ListVal(listVals)
		case map[string]interface{}:
			values[k] = GetCtyValue(val)
		default:
			values[k] = cty.NullVal(cty.String)
		}
	}
	return cty.ObjectVal(values)
}

// MockResourceData creates a mock ResourceData for testing
func MockResourceData(resourceSchema map[string]*schema.Schema, data map[string]interface{}) *schema.ResourceData {
	// Use schema.Resource to create a proper ResourceData
	resource := &schema.Resource{
		Schema: resourceSchema,
	}

	// Create ResourceData using the resource
	d := resource.TestResourceData()

	// Set the config values
	for key, value := range data {
		d.Set(key, value)
	}

	return d
}

// MockDiagnostics creates mock diagnostics for testing
func MockDiagnostics() diag.Diagnostics {
	return diag.Diagnostics{}
}

// MockContext creates a mock context for testing
func MockContext() context.Context {
	return context.Background()
}

// MockInterface creates a mock interface{} for testing
func MockInterface() interface{} {
	return struct {
		TestField string
	}{
		TestField: "test-value",
	}
}
