package rafay

import (
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestExpandAKSClusterMetadata tests the expandAKSClusterMetadata function
func TestExpandAKSClusterMetadata(t *testing.T) {
	tests := []struct {
		name     string
		input    []interface{}
		expected *AKSClusterMetadata
	}{
		{
			name:     "empty input",
			input:    []interface{}{},
			expected: &AKSClusterMetadata{},
		},
		{
			name: "complete metadata",
			input: []interface{}{
				map[string]interface{}{
					"name":    "test-aks-cluster",
					"project": "test-project",
					"labels": map[string]interface{}{
						"env":     "test",
						"version": "1.0",
					},
				},
			},
			expected: &AKSClusterMetadata{
				Name:    "test-aks-cluster",
				Project: "test-project",
				Labels: map[string]string{
					"env":     "test",
					"version": "1.0",
				},
			},
		},
		{
			name: "partial metadata",
			input: []interface{}{
				map[string]interface{}{
					"name": "minimal-aks-cluster",
				},
			},
			expected: &AKSClusterMetadata{
				Name: "minimal-aks-cluster",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expandAKSClusterMetadata(tt.input)
			assert.Equal(t, tt.expected.Name, result.Name)
			assert.Equal(t, tt.expected.Project, result.Project)
			assert.Equal(t, tt.expected.Labels, result.Labels)
		})
	}
}

// TestExpandAKSClusterSpec tests the expandAKSClusterSpec function
func TestExpandAKSClusterSpec(t *testing.T) {
	tests := []struct {
		name      string
		input     []interface{}
		rawConfig cty.Value
		expected  *AKSClusterSpec
	}{
		{
			name:      "empty input",
			input:     []interface{}{},
			rawConfig: cty.NullVal(cty.Object(map[string]cty.Type{})),
			expected:  &AKSClusterSpec{},
		},
		{
			name: "complete spec",
			input: []interface{}{
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
			rawConfig: cty.ObjectVal(map[string]cty.Value{
				"type":          cty.StringVal("aks"),
				"blueprint":     cty.StringVal("minimal"),
				"cloudprovider": cty.StringVal("azure"),
			}),
			expected: &AKSClusterSpec{
				Type:             "aks",
				Blueprint:        "minimal",
				BlueprintVersion: "1.0",
				CloudProvider:    "azure",
				AKSClusterConfig: &AKSClusterConfig{
					APIVersion: "rafay.io/v1alpha5",
					Kind:       "Cluster",
					Metadata: &AKSClusterConfigMetadata{
						Name: "test-aks-cluster",
					},
					Spec: &AKSClusterConfigSpec{
						SubscriptionID:    "12345678-1234-1234-1234-123456789012",
						ResourceGroupName: "test-rg",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expandAKSClusterSpec(tt.input, tt.rawConfig)
			assert.Equal(t, tt.expected.Type, result.Type)
			assert.Equal(t, tt.expected.Blueprint, result.Blueprint)
			assert.Equal(t, tt.expected.BlueprintVersion, result.BlueprintVersion)
			assert.Equal(t, tt.expected.CloudProvider, result.CloudProvider)

			if tt.expected.AKSClusterConfig != nil {
				require.NotNil(t, result.AKSClusterConfig)
				assert.Equal(t, tt.expected.AKSClusterConfig.APIVersion, result.AKSClusterConfig.APIVersion)
				assert.Equal(t, tt.expected.AKSClusterConfig.Kind, result.AKSClusterConfig.Kind)

				if tt.expected.AKSClusterConfig.Metadata != nil {
					require.NotNil(t, result.AKSClusterConfig.Metadata)
					assert.Equal(t, tt.expected.AKSClusterConfig.Metadata.Name, result.AKSClusterConfig.Metadata.Name)
				}

				if tt.expected.AKSClusterConfig.Spec != nil {
					require.NotNil(t, result.AKSClusterConfig.Spec)
					assert.Equal(t, tt.expected.AKSClusterConfig.Spec.SubscriptionID, result.AKSClusterConfig.Spec.SubscriptionID)
					assert.Equal(t, tt.expected.AKSClusterConfig.Spec.ResourceGroupName, result.AKSClusterConfig.Spec.ResourceGroupName)
				}
			}
		})
	}
}

// TestExpandAKSClusterConfig tests the expandAKSClusterConfig function
func TestExpandAKSClusterConfig(t *testing.T) {
	tests := []struct {
		name      string
		input     []interface{}
		rawConfig cty.Value
		expected  *AKSClusterConfig
	}{
		{
			name:      "empty input",
			input:     []interface{}{},
			rawConfig: cty.NullVal(cty.Object(map[string]cty.Type{})),
			expected:  &AKSClusterConfig{},
		},
		{
			name: "complete config",
			input: []interface{}{
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
									"properties": []interface{}{
										map[string]interface{}{
											"dns_prefix":         "test-aks",
											"kubernetes_version": "1.25.6",
										},
									},
								},
							},
						},
					},
				},
			},
			rawConfig: cty.ObjectVal(map[string]cty.Value{
				"apiversion": cty.StringVal("rafay.io/v1alpha5"),
				"kind":       cty.StringVal("Cluster"),
			}),
			expected: &AKSClusterConfig{
				APIVersion: "rafay.io/v1alpha5",
				Kind:       "Cluster",
				Metadata: &AKSClusterConfigMetadata{
					Name: "test-aks-cluster",
				},
				Spec: &AKSClusterConfigSpec{
					SubscriptionID:    "12345678-1234-1234-1234-123456789012",
					ResourceGroupName: "test-rg",
					ManagedCluster: &AKSManagedCluster{
						Location: "East US",
						Properties: &AKSManagedClusterProperties{
							DNSPrefix:         "test-aks",
							KubernetesVersion: "1.25.6",
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expandAKSClusterConfig(tt.input, tt.rawConfig)
			assert.Equal(t, tt.expected.APIVersion, result.APIVersion)
			assert.Equal(t, tt.expected.Kind, result.Kind)

			if tt.expected.Metadata != nil {
				require.NotNil(t, result.Metadata)
				assert.Equal(t, tt.expected.Metadata.Name, result.Metadata.Name)
			}

			if tt.expected.Spec != nil {
				require.NotNil(t, result.Spec)
				assert.Equal(t, tt.expected.Spec.SubscriptionID, result.Spec.SubscriptionID)
				assert.Equal(t, tt.expected.Spec.ResourceGroupName, result.Spec.ResourceGroupName)

				if tt.expected.Spec.ManagedCluster != nil {
					require.NotNil(t, result.Spec.ManagedCluster)
					assert.Equal(t, tt.expected.Spec.ManagedCluster.Location, result.Spec.ManagedCluster.Location)

					if tt.expected.Spec.ManagedCluster.Properties != nil {
						require.NotNil(t, result.Spec.ManagedCluster.Properties)
						assert.Equal(t, tt.expected.Spec.ManagedCluster.Properties.DNSPrefix, result.Spec.ManagedCluster.Properties.DNSPrefix)
						assert.Equal(t, tt.expected.Spec.ManagedCluster.Properties.KubernetesVersion, result.Spec.ManagedCluster.Properties.KubernetesVersion)
					}
				}
			}
		})
	}
}

// TestExpandAKSManagedCluster tests the expandAKSConfigManagedCluster function
func TestExpandAKSManagedCluster(t *testing.T) {
	tests := []struct {
		name     string
		input    []interface{}
		expected *AKSManagedCluster
	}{
		{
			name:     "empty input",
			input:    []interface{}{},
			expected: &AKSManagedCluster{},
		},
		{
			name: "complete managed cluster",
			input: []interface{}{
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
							"api_server_access_profile": []interface{}{
								map[string]interface{}{
									"enable_private_cluster":             true,
									"private_dns_zone":                   "system",
									"enable_private_cluster_public_fqdn": false,
								},
							},
							"network_profile": []interface{}{
								map[string]interface{}{
									"network_plugin":     "azure",
									"network_policy":     "azure",
									"dns_service_ip":     "10.0.0.10",
									"service_cidr":       "10.0.0.0/16",
									"docker_bridge_cidr": "172.17.0.1/16",
								},
							},
						},
					},
				},
			},
			expected: &AKSManagedCluster{
				Location: "East US",
				Tags: map[string]string{
					"Environment": "test",
					"Team":        "platform",
				},
				Identity: &AKSManagedClusterIdentity{
					Type: "SystemAssigned",
				},
				Properties: &AKSManagedClusterProperties{
					DNSPrefix:         "test-aks",
					KubernetesVersion: "1.25.6",
					NodeResourceGroup: "MC_test-rg_test-aks_eastus",
					APIServerAccessProfile: &AKSManagedClusterAPIServerAccessProfile{
						EnablePrivateCluster:           &[]bool{true}[0],
						PrivateDNSZone:                 "system",
						EnablePrivateClusterPublicFQDN: &[]bool{false}[0],
					},
					NetworkProfile: &AKSManagedClusterNetworkProfile{
						NetworkPlugin:    "azure",
						NetworkPolicy:    "azure",
						DNSServiceIP:     "10.0.0.10",
						ServiceCIDR:      "10.0.0.0/16",
						DockerBridgeCIDR: "172.17.0.1/16",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expandAKSConfigManagedCluster(tt.input)
			assert.Equal(t, tt.expected.Location, result.Location)
			assert.Equal(t, tt.expected.Tags, result.Tags)

			if tt.expected.Identity != nil {
				require.NotNil(t, result.Identity)
				assert.Equal(t, tt.expected.Identity.Type, result.Identity.Type)
			}

			if tt.expected.Properties != nil {
				require.NotNil(t, result.Properties)
				assert.Equal(t, tt.expected.Properties.DNSPrefix, result.Properties.DNSPrefix)
				assert.Equal(t, tt.expected.Properties.KubernetesVersion, result.Properties.KubernetesVersion)
				assert.Equal(t, tt.expected.Properties.NodeResourceGroup, result.Properties.NodeResourceGroup)

				if tt.expected.Properties.APIServerAccessProfile != nil {
					require.NotNil(t, result.Properties.APIServerAccessProfile)
					assert.Equal(t, tt.expected.Properties.APIServerAccessProfile.EnablePrivateCluster, result.Properties.APIServerAccessProfile.EnablePrivateCluster)
					assert.Equal(t, tt.expected.Properties.APIServerAccessProfile.PrivateDNSZone, result.Properties.APIServerAccessProfile.PrivateDNSZone)
					assert.Equal(t, tt.expected.Properties.APIServerAccessProfile.EnablePrivateClusterPublicFQDN, result.Properties.APIServerAccessProfile.EnablePrivateClusterPublicFQDN)
				}

				if tt.expected.Properties.NetworkProfile != nil {
					require.NotNil(t, result.Properties.NetworkProfile)
					assert.Equal(t, tt.expected.Properties.NetworkProfile.NetworkPlugin, result.Properties.NetworkProfile.NetworkPlugin)
					assert.Equal(t, tt.expected.Properties.NetworkProfile.NetworkPolicy, result.Properties.NetworkProfile.NetworkPolicy)
					assert.Equal(t, tt.expected.Properties.NetworkProfile.DNSServiceIP, result.Properties.NetworkProfile.DNSServiceIP)
					assert.Equal(t, tt.expected.Properties.NetworkProfile.ServiceCIDR, result.Properties.NetworkProfile.ServiceCIDR)
					assert.Equal(t, tt.expected.Properties.NetworkProfile.DockerBridgeCIDR, result.Properties.NetworkProfile.DockerBridgeCIDR)
				}
			}
		})
	}
}

// TestExpandAKSNodePool tests the expandAKSNodePool function
func TestExpandAKSNodePool(t *testing.T) {
	tests := []struct {
		name      string
		input     []interface{}
		rawConfig cty.Value
		expected  []*AKSNodePool
	}{
		{
			name:      "empty input",
			input:     []interface{}{},
			rawConfig: cty.NullVal(cty.Object(map[string]cty.Type{})),
			expected:  []*AKSNodePool{},
		},
		{
			name: "single node pool",
			input: []interface{}{
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
							"kubelet_config": []interface{}{
								map[string]interface{}{
									"cpu_manager_policy":      "static",
									"cpu_cfs_quota":           &[]bool{true}[0],
									"cpu_cfs_quota_period":    "100ms",
									"image_gc_high_threshold": 85,
									"image_gc_low_threshold":  80,
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
										},
									},
								},
							},
						},
					},
				},
			},
			rawConfig: cty.ObjectVal(map[string]cty.Value{
				"apiversion": cty.StringVal("2022-03-01"),
				"name":       cty.StringVal("nodepool1"),
			}),
			expected: []*AKSNodePool{
				{
					APIVersion: "2022-03-01",
					Name:       "nodepool1",
					Properties: &AKSNodePoolProperties{
						Count:             &[]int64{3}[0],
						VMSize:            "Standard_DS2_v2",
						OSType:            "Linux",
						Type:              "VirtualMachineScaleSets",
						Mode:              "System",
						MaxPods:           &[]int64{30}[0],
						AvailabilityZones: []string{"1", "2", "3"},
						EnableAutoScaling: &[]bool{true}[0],
						MinCount:          &[]int64{1}[0],
						MaxCount:          &[]int64{5}[0],
						OSDiskSizeGB:      &[]int64{100}[0],
						OSDiskType:        "Managed",
						KubeletConfig: &AKSNodePoolKubeletConfig{
							CPUManagerPolicy:     "static",
							CPUCfsQuota:          &[]bool{true}[0],
							CPUCfsQuotaPeriod:    "100ms",
							ImageGcHighThreshold: &[]int64{85}[0],
							ImageGcLowThreshold:  &[]int64{80}[0],
						},
						LinuxOSConfig: &AKSNodePoolLinuxOsConfig{
							TransparentHugePageEnabled: "always",
							TransparentHugePageDefrag:  "always",
							SwapFileSizeMB:             &[]int64{1024}[0],
							Sysctls: &AKSNodePoolLinuxOsConfigSysctls{
								VMMaxMapCount:           &[]int64{262144}[0],
								FsInotifyMaxUserWatches: &[]int64{1048576}[0],
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expandAKSNodePool(tt.input, tt.rawConfig)
			assert.Len(t, result, len(tt.expected))

			for i, expected := range tt.expected {
				if i < len(result) {
					assert.Equal(t, expected.APIVersion, result[i].APIVersion)
					assert.Equal(t, expected.Name, result[i].Name)

					if expected.Properties != nil {
						require.NotNil(t, result[i].Properties)
						assert.Equal(t, expected.Properties.Count, result[i].Properties.Count)
						assert.Equal(t, expected.Properties.VMSize, result[i].Properties.VMSize)
						assert.Equal(t, expected.Properties.OSType, result[i].Properties.OSType)
						assert.Equal(t, expected.Properties.Type, result[i].Properties.Type)
						assert.Equal(t, expected.Properties.Mode, result[i].Properties.Mode)
						assert.Equal(t, expected.Properties.MaxPods, result[i].Properties.MaxPods)
						assert.Equal(t, expected.Properties.AvailabilityZones, result[i].Properties.AvailabilityZones)
						assert.Equal(t, expected.Properties.EnableAutoScaling, result[i].Properties.EnableAutoScaling)
						assert.Equal(t, expected.Properties.MinCount, result[i].Properties.MinCount)
						assert.Equal(t, expected.Properties.MaxCount, result[i].Properties.MaxCount)
						assert.Equal(t, expected.Properties.OSDiskSizeGB, result[i].Properties.OSDiskSizeGB)
						assert.Equal(t, expected.Properties.OSDiskType, result[i].Properties.OSDiskType)

						if expected.Properties.KubeletConfig != nil {
							require.NotNil(t, result[i].Properties.KubeletConfig)
							assert.Equal(t, expected.Properties.KubeletConfig.CPUManagerPolicy, result[i].Properties.KubeletConfig.CPUManagerPolicy)
							assert.Equal(t, expected.Properties.KubeletConfig.CPUCfsQuota, result[i].Properties.KubeletConfig.CPUCfsQuota)
							assert.Equal(t, expected.Properties.KubeletConfig.CPUCfsQuotaPeriod, result[i].Properties.KubeletConfig.CPUCfsQuotaPeriod)
							assert.Equal(t, expected.Properties.KubeletConfig.ImageGcHighThreshold, result[i].Properties.KubeletConfig.ImageGcHighThreshold)
							assert.Equal(t, expected.Properties.KubeletConfig.ImageGcLowThreshold, result[i].Properties.KubeletConfig.ImageGcLowThreshold)
						}

						if expected.Properties.LinuxOSConfig != nil {
							require.NotNil(t, result[i].Properties.LinuxOSConfig)
							assert.Equal(t, expected.Properties.LinuxOSConfig.TransparentHugePageEnabled, result[i].Properties.LinuxOSConfig.TransparentHugePageEnabled)
							assert.Equal(t, expected.Properties.LinuxOSConfig.TransparentHugePageDefrag, result[i].Properties.LinuxOSConfig.TransparentHugePageDefrag)
							assert.Equal(t, expected.Properties.LinuxOSConfig.SwapFileSizeMB, result[i].Properties.LinuxOSConfig.SwapFileSizeMB)

							if expected.Properties.LinuxOSConfig.Sysctls != nil {
								require.NotNil(t, result[i].Properties.LinuxOSConfig.Sysctls)
								assert.Equal(t, expected.Properties.LinuxOSConfig.Sysctls.VMMaxMapCount, result[i].Properties.LinuxOSConfig.Sysctls.VMMaxMapCount)
								assert.Equal(t, expected.Properties.LinuxOSConfig.Sysctls.FsInotifyMaxUserWatches, result[i].Properties.LinuxOSConfig.Sysctls.FsInotifyMaxUserWatches)
							}
						}
					}
				}
			}
		})
	}
}

// TestExpandAKSMaintenanceConfigs tests the expandAKSMaintenanceConfigs function
func TestExpandAKSMaintenanceConfigs(t *testing.T) {
	tests := []struct {
		name     string
		input    []interface{}
		expected []*AKSMaintenanceConfig
	}{
		{
			name:     "empty input",
			input:    []interface{}{},
			expected: []*AKSMaintenanceConfig{},
		},
		{
			name: "single maintenance config",
			input: []interface{}{
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
			expected: []*AKSMaintenanceConfig{
				{
					Name: "aksManagedAutoUpgradeSchedule",
					Properties: &AKSMaintenanceConfigProperties{
						MaintenanceWindow: &AKSMaintenanceWindow{
							Schedule: &AKSMaintenanceSchedule{
								Weekly: &AKSMaintenanceWeeklySchedule{
									IntervalWeeks: &[]int64{1}[0],
									DayOfWeek:     "Sunday",
								},
							},
							DurationHours: &[]int64{4}[0],
							UTCOffset:     "+00:00",
							StartDate:     "2023-01-01",
							StartTime:     "01:00",
						},
						NotAllowedTime: []*AKSMaintenanceTimeSpan{
							{
								Start: "2023-12-23T00:00:00Z",
								End:   "2023-12-25T23:59:59Z",
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expandAKSMaintenanceConfigs(tt.input)
			assert.Len(t, result, len(tt.expected))

			for i, expected := range tt.expected {
				if i < len(result) {
					assert.Equal(t, expected.Name, result[i].Name)

					if expected.Properties != nil {
						require.NotNil(t, result[i].Properties)

						if expected.Properties.MaintenanceWindow != nil {
							require.NotNil(t, result[i].Properties.MaintenanceWindow)
							assert.Equal(t, expected.Properties.MaintenanceWindow.DurationHours, result[i].Properties.MaintenanceWindow.DurationHours)
							assert.Equal(t, expected.Properties.MaintenanceWindow.UTCOffset, result[i].Properties.MaintenanceWindow.UTCOffset)
							assert.Equal(t, expected.Properties.MaintenanceWindow.StartDate, result[i].Properties.MaintenanceWindow.StartDate)
							assert.Equal(t, expected.Properties.MaintenanceWindow.StartTime, result[i].Properties.MaintenanceWindow.StartTime)

							if expected.Properties.MaintenanceWindow.Schedule != nil {
								require.NotNil(t, result[i].Properties.MaintenanceWindow.Schedule)

								if expected.Properties.MaintenanceWindow.Schedule.Weekly != nil {
									require.NotNil(t, result[i].Properties.MaintenanceWindow.Schedule.Weekly)
									assert.Equal(t, expected.Properties.MaintenanceWindow.Schedule.Weekly.IntervalWeeks, result[i].Properties.MaintenanceWindow.Schedule.Weekly.IntervalWeeks)
									assert.Equal(t, expected.Properties.MaintenanceWindow.Schedule.Weekly.DayOfWeek, result[i].Properties.MaintenanceWindow.Schedule.Weekly.DayOfWeek)
								}
							}
						}

						if expected.Properties.NotAllowedTime != nil {
							require.NotNil(t, result[i].Properties.NotAllowedTime)
							assert.Len(t, result[i].Properties.NotAllowedTime, len(expected.Properties.NotAllowedTime))

							for j, expectedTimeSpan := range expected.Properties.NotAllowedTime {
								if j < len(result[i].Properties.NotAllowedTime) {
									assert.Equal(t, expectedTimeSpan.Start, result[i].Properties.NotAllowedTime[j].Start)
									assert.Equal(t, expectedTimeSpan.End, result[i].Properties.NotAllowedTime[j].End)
								}
							}
						}
					}
				}
			}
		})
	}
}

// Benchmark tests for AKS expand functions
func BenchmarkExpandAKSClusterMetadata(b *testing.B) {
	input := []interface{}{
		map[string]interface{}{
			"name":    "benchmark-aks-cluster",
			"project": "benchmark-project",
			"labels": map[string]interface{}{
				"env":     "benchmark",
				"version": "1.0",
				"team":    "platform",
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		expandAKSClusterMetadata(input)
	}
}

func BenchmarkExpandAKSClusterSpec(b *testing.B) {
	input := []interface{}{
		map[string]interface{}{
			"type":             "aks",
			"blueprint":        "minimal",
			"blueprintversion": "1.0",
			"cloudprovider":    "azure",
		},
	}
	rawConfig := cty.ObjectVal(map[string]cty.Value{
		"type":          cty.StringVal("aks"),
		"blueprint":     cty.StringVal("minimal"),
		"cloudprovider": cty.StringVal("azure"),
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		expandAKSClusterSpec(input, rawConfig)
	}
}
