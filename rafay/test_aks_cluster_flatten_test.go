package rafay

import (
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFlattenAKSClusterMetadata tests the flattenAKSClusterMetadata function
func TestFlattenAKSClusterMetadata(t *testing.T) {
	tests := []struct {
		name     string
		input    *AKSClusterMetadata
		p        []interface{}
		expected []interface{}
	}{
		{
			name:     "nil input",
			input:    nil,
			p:        []interface{}{},
			expected: nil,
		},
		{
			name: "complete metadata",
			input: &AKSClusterMetadata{
				Name:    "test-aks-cluster",
				Project: "test-project",
				Labels: map[string]string{
					"env":     "test",
					"version": "1.0",
				},
			},
			p: []interface{}{},
			expected: []interface{}{
				map[string]interface{}{
					"name":    "test-aks-cluster",
					"project": "test-project",
					"labels": map[string]interface{}{
						"env":     "test",
						"version": "1.0",
					},
				},
			},
		},
		{
			name: "minimal metadata",
			input: &AKSClusterMetadata{
				Name: "minimal-aks-cluster",
			},
			p: []interface{}{},
			expected: []interface{}{
				map[string]interface{}{
					"name": "minimal-aks-cluster",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := flattenAKSClusterMetadata(tt.input, tt.p)

			if tt.expected == nil {
				assert.Nil(t, result)
				return
			}

			require.NotNil(t, result)
			assert.Len(t, result, 1)

			resultMap := result[0].(map[string]interface{})
			expectedMap := tt.expected[0].(map[string]interface{})

			assert.Equal(t, expectedMap["name"], resultMap["name"])

			if expectedMap["project"] != nil {
				assert.Equal(t, expectedMap["project"], resultMap["project"])
			}

			if expectedMap["labels"] != nil {
				assert.Equal(t, expectedMap["labels"], resultMap["labels"])
			}
		})
	}
}

// TestFlattenAKSClusterSpec tests the flattenAKSClusterSpec function
func TestFlattenAKSClusterSpec(t *testing.T) {
	tests := []struct {
		name     string
		input    *AKSClusterSpec
		p        []interface{}
		rawState cty.Value
		expected []interface{}
	}{
		{
			name:     "nil input",
			input:    nil,
			p:        []interface{}{},
			rawState: cty.NullVal(cty.Object(map[string]cty.Type{})),
			expected: nil,
		},
		{
			name: "complete spec",
			input: &AKSClusterSpec{
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
			p: []interface{}{},
			rawState: cty.ListVal([]cty.Value{
				cty.ObjectVal(map[string]cty.Value{
					"cluster_config": cty.ListVal([]cty.Value{
						cty.ObjectVal(map[string]cty.Value{
							"apiversion": cty.StringVal("rafay.io/v1alpha5"),
							"kind":       cty.StringVal("Cluster"),
							"metadata": cty.ListVal([]cty.Value{
								cty.ObjectVal(map[string]cty.Value{
									"name": cty.StringVal("test-aks-cluster"),
								}),
							}),
							"spec": cty.ListVal([]cty.Value{
								cty.ObjectVal(map[string]cty.Value{
									"subscription_id":     cty.StringVal("12345678-1234-1234-1234-123456789012"),
									"resource_group_name": cty.StringVal("test-rg"),
								}),
							}),
						}),
					}),
				}),
			}),
			expected: []interface{}{
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
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := flattenAKSClusterSpec(tt.input, tt.p, tt.rawState)

			if tt.expected == nil {
				assert.Nil(t, result)
				return
			}

			require.NotNil(t, result)
			assert.Len(t, result, 1)

			resultMap := result[0].(map[string]interface{})
			expectedMap := tt.expected[0].(map[string]interface{})

			assert.Equal(t, expectedMap["type"], resultMap["type"])
			assert.Equal(t, expectedMap["blueprint"], resultMap["blueprint"])
			assert.Equal(t, expectedMap["blueprintversion"], resultMap["blueprintversion"])
			assert.Equal(t, expectedMap["cloudprovider"], resultMap["cloudprovider"])

			if expectedMap["cluster_config"] != nil {
				assert.NotNil(t, resultMap["cluster_config"])
				resultConfig := resultMap["cluster_config"].([]interface{})[0].(map[string]interface{})
				expectedConfig := expectedMap["cluster_config"].([]interface{})[0].(map[string]interface{})

				assert.Equal(t, expectedConfig["apiversion"], resultConfig["apiversion"])
				assert.Equal(t, expectedConfig["kind"], resultConfig["kind"])

				if expectedConfig["metadata"] != nil {
					assert.NotNil(t, resultConfig["metadata"])
					resultMetadata := resultConfig["metadata"].([]interface{})[0].(map[string]interface{})
					expectedMetadata := expectedConfig["metadata"].([]interface{})[0].(map[string]interface{})
					assert.Equal(t, expectedMetadata["name"], resultMetadata["name"])
				}

				if expectedConfig["spec"] != nil {
					assert.NotNil(t, resultConfig["spec"])
					resultSpec := resultConfig["spec"].([]interface{})[0].(map[string]interface{})
					expectedSpec := expectedConfig["spec"].([]interface{})[0].(map[string]interface{})
					assert.Equal(t, expectedSpec["subscription_id"], resultSpec["subscription_id"])
					assert.Equal(t, expectedSpec["resource_group_name"], resultSpec["resource_group_name"])
				}
			}
		})
	}
}

// TestFlattenAKSCluster tests the flattenAKSCluster function
func TestFlattenAKSCluster(t *testing.T) {
	tests := []struct {
		name      string
		input     *AKSCluster
		expectErr bool
	}{
		{
			name:      "nil input",
			input:     nil,
			expectErr: false,
		},
		{
			name: "complete cluster",
			input: &AKSCluster{
				APIVersion: "rafay.io/v1alpha5",
				Kind:       "Cluster",
				Metadata: &AKSClusterMetadata{
					Name:    "test-aks-cluster",
					Project: "test-project",
					Labels: map[string]string{
						"env": "test",
					},
				},
				Spec: &AKSClusterSpec{
					Type:             "aks",
					Blueprint:        "minimal",
					BlueprintVersion: "1.0",
					CloudProvider:    "azure",
				},
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock ResourceData
			d := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
				"apiversion": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"kind": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"metadata": {
					Type:     schema.TypeList,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"name": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"project": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"labels": {
								Type:     schema.TypeMap,
								Optional: true,
								Elem:     &schema.Schema{Type: schema.TypeString},
							},
						},
					},
				},
				"spec": {
					Type:     schema.TypeList,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"type": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"blueprint": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"blueprintversion": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"cloudprovider": {
								Type:     schema.TypeString,
								Optional: true,
							},
						},
					},
				},
			}, map[string]interface{}{})

			err := flattenAKSCluster(d, tt.input)

			if tt.expectErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			if tt.input != nil {
				// Verify the data was set correctly
				if tt.input.APIVersion != "" {
					assert.Equal(t, tt.input.APIVersion, d.Get("apiversion").(string))
				}
				if tt.input.Kind != "" {
					assert.Equal(t, tt.input.Kind, d.Get("kind").(string))
				}

				if tt.input.Metadata != nil {
					metadata := d.Get("metadata").([]interface{})
					require.Len(t, metadata, 1)
					metadataMap := metadata[0].(map[string]interface{})

					if tt.input.Metadata.Name != "" {
						assert.Equal(t, tt.input.Metadata.Name, metadataMap["name"].(string))
					}
					if tt.input.Metadata.Project != "" {
						assert.Equal(t, tt.input.Metadata.Project, metadataMap["project"].(string))
					}
				}

				if tt.input.Spec != nil {
					spec := d.Get("spec").([]interface{})
					require.Len(t, spec, 1)
					specMap := spec[0].(map[string]interface{})

					if tt.input.Spec.Type != "" {
						assert.Equal(t, tt.input.Spec.Type, specMap["type"].(string))
					}
					if tt.input.Spec.Blueprint != "" {
						assert.Equal(t, tt.input.Spec.Blueprint, specMap["blueprint"].(string))
					}
					if tt.input.Spec.CloudProvider != "" {
						assert.Equal(t, tt.input.Spec.CloudProvider, specMap["cloudprovider"].(string))
					}
				}
			}
		})
	}
}

// TestFlattenAKSManagedCluster tests the flattenAKSManagedCluster function
func TestFlattenAKSManagedCluster(t *testing.T) {
	tests := []struct {
		name     string
		input    *AKSManagedCluster
		p        []interface{}
		expected []interface{}
	}{
		{
			name:     "nil input",
			input:    nil,
			p:        []interface{}{},
			expected: nil,
		},
		{
			name: "complete managed cluster",
			input: &AKSManagedCluster{
				Location: "East US",
				Tags: map[string]interface{}{
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
				},
			},
			p: []interface{}{},
			expected: []interface{}{
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
		{
			name: "minimal managed cluster",
			input: &AKSManagedCluster{
				Location: "West US 2",
			},
			p: []interface{}{},
			expected: []interface{}{
				map[string]interface{}{
					"location": "West US 2",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := flattenAKSManagedCluster(tt.input, tt.p)

			if tt.expected == nil {
				assert.Nil(t, result)
				return
			}

			require.NotNil(t, result)
			assert.Len(t, result, 1)

			resultMap := result[0].(map[string]interface{})
			expectedMap := tt.expected[0].(map[string]interface{})

			assert.Equal(t, expectedMap["location"], resultMap["location"])

			if expectedMap["tags"] != nil {
				assert.Equal(t, expectedMap["tags"], resultMap["tags"])
			}

			if expectedMap["identity"] != nil {
				assert.NotNil(t, resultMap["identity"])
				resultIdentity := resultMap["identity"].([]interface{})[0].(map[string]interface{})
				expectedIdentity := expectedMap["identity"].([]interface{})[0].(map[string]interface{})
				assert.Equal(t, expectedIdentity["type"], resultIdentity["type"])
			}

			if expectedMap["properties"] != nil {
				assert.NotNil(t, resultMap["properties"])
				resultProperties := resultMap["properties"].([]interface{})[0].(map[string]interface{})
				expectedProperties := expectedMap["properties"].([]interface{})[0].(map[string]interface{})

				if expectedProperties["dns_prefix"] != nil {
					assert.Equal(t, expectedProperties["dns_prefix"], resultProperties["dns_prefix"])
				}
				if expectedProperties["kubernetes_version"] != nil {
					assert.Equal(t, expectedProperties["kubernetes_version"], resultProperties["kubernetes_version"])
				}
				if expectedProperties["node_resource_group"] != nil {
					assert.Equal(t, expectedProperties["node_resource_group"], resultProperties["node_resource_group"])
				}
			}
		})
	}
}

// TestFlattenAKSNodePool tests the flattenAKSNodePool function
func TestFlattenAKSNodePool(t *testing.T) {
	tests := []struct {
		name     string
		input    []*AKSNodePool
		p        []interface{}
		rawState cty.Value
		expected []interface{}
	}{
		{
			name:     "nil input",
			input:    nil,
			p:        []interface{}{},
			rawState: cty.NullVal(cty.Object(map[string]cty.Type{})),
			expected: []interface{}{},
		},
		{
			name:     "empty input",
			input:    []*AKSNodePool{},
			p:        []interface{}{},
			rawState: cty.NullVal(cty.Object(map[string]cty.Type{})),
			expected: []interface{}{},
		},
		{
			name: "single node pool",
			input: []*AKSNodePool{
				{
					APIVersion: "2022-03-01",
					Name:       "nodepool1",
					Properties: &AKSNodePoolProperties{
						Count:             &[]int{3}[0],
						VmSize:            "Standard_DS2_v2",
						OsType:            "Linux",
						Type:              "VirtualMachineScaleSets",
						Mode:              "System",
						MaxPods:           &[]int{30}[0],
						AvailabilityZones: []string{"1", "2", "3"},
						EnableAutoScaling: &[]bool{true}[0],
						MinCount:          &[]int{1}[0],
						MaxCount:          &[]int{5}[0],
						OsDiskSizeGB:      &[]int{100}[0],
						OsDiskType:        "Managed",
					},
				},
			},
			p:        []interface{}{},
			rawState: cty.ObjectVal(map[string]cty.Value{}),
			expected: []interface{}{
				map[string]interface{}{
					"apiversion": "2022-03-01",
					"name":       "nodepool1",
					"properties": []interface{}{
						map[string]interface{}{
							"count":               int64(3),
							"vm_size":             "Standard_DS2_v2",
							"os_type":             "Linux",
							"type":                "VirtualMachineScaleSets",
							"mode":                "System",
							"max_pods":            int64(30),
							"availability_zones":  []string{"1", "2", "3"},
							"enable_auto_scaling": true,
							"min_count":           int64(1),
							"max_count":           int64(5),
							"os_disk_size_gb":     int64(100),
							"os_disk_type":        "Managed",
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := flattenAKSNodePool(tt.input, tt.p, tt.rawState)
			assert.Len(t, result, len(tt.expected))

			for i, expected := range tt.expected {
				if i < len(result) {
					resultMap := result[i].(map[string]interface{})
					expectedMap := expected.(map[string]interface{})

					assert.Equal(t, expectedMap["apiversion"], resultMap["apiversion"])
					assert.Equal(t, expectedMap["name"], resultMap["name"])

					if expectedMap["properties"] != nil {
						assert.NotNil(t, resultMap["properties"])
						resultProperties := resultMap["properties"].([]interface{})[0].(map[string]interface{})
						expectedProperties := expectedMap["properties"].([]interface{})[0].(map[string]interface{})

						assert.Equal(t, expectedProperties["count"], resultProperties["count"])
						assert.Equal(t, expectedProperties["vm_size"], resultProperties["vm_size"])
						assert.Equal(t, expectedProperties["os_type"], resultProperties["os_type"])
						assert.Equal(t, expectedProperties["type"], resultProperties["type"])
						assert.Equal(t, expectedProperties["mode"], resultProperties["mode"])
						assert.Equal(t, expectedProperties["max_pods"], resultProperties["max_pods"])
						assert.Equal(t, expectedProperties["availability_zones"], resultProperties["availability_zones"])
						assert.Equal(t, expectedProperties["enable_auto_scaling"], resultProperties["enable_auto_scaling"])
						assert.Equal(t, expectedProperties["min_count"], resultProperties["min_count"])
						assert.Equal(t, expectedProperties["max_count"], resultProperties["max_count"])
						assert.Equal(t, expectedProperties["os_disk_size_gb"], resultProperties["os_disk_size_gb"])
						assert.Equal(t, expectedProperties["os_disk_type"], resultProperties["os_disk_type"])
					}
				}
			}
		})
	}
}

// TestFlattenAKSMaintenanceConfigs tests the flattenAKSMaintenanceConfigs function
func TestFlattenAKSMaintenanceConfigs(t *testing.T) {
	tests := []struct {
		name     string
		input    []*AKSMaintenanceConfig
		p        []interface{}
		expected []interface{}
	}{
		{
			name:     "nil input",
			input:    nil,
			p:        []interface{}{},
			expected: []interface{}{},
		},
		{
			name:     "empty input",
			input:    []*AKSMaintenanceConfig{},
			p:        []interface{}{},
			expected: []interface{}{},
		},
		{
			name: "single maintenance config",
			input: []*AKSMaintenanceConfig{
				{
					Name: "aksManagedAutoUpgradeSchedule",
					Properties: &AKSMaintenanceConfigProperties{
						MaintenanceWindow: &AKSMaintenanceWindow{
							Schedule: &AKSMaintenanceSchedule{
								WeeklySchedule: &AKSMaintenanceWeeklySchedule{
									IntervalWeeks: 1,
									DayOfWeek:     "Sunday",
								},
							},
							DurationHours: 4,
							UtcOffset:     "+00:00",
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
			p: []interface{}{},
			expected: []interface{}{
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
													"interval_weeks": int64(1),
													"day_of_week":    "Sunday",
												},
											},
										},
									},
									"duration_hours": int64(4),
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
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := flattenAKSMaintenanceConfigs(tt.input, tt.p)
			assert.Len(t, result, len(tt.expected))

			for i, expected := range tt.expected {
				if i < len(result) {
					resultMap := result[i].(map[string]interface{})
					expectedMap := expected.(map[string]interface{})

					assert.Equal(t, expectedMap["name"], resultMap["name"])

					if expectedMap["properties"] != nil {
						assert.NotNil(t, resultMap["properties"])
						resultProperties := resultMap["properties"].([]interface{})[0].(map[string]interface{})
						expectedProperties := expectedMap["properties"].([]interface{})[0].(map[string]interface{})

						if expectedProperties["maintenance_window"] != nil {
							assert.NotNil(t, resultProperties["maintenance_window"])
							resultMW := resultProperties["maintenance_window"].([]interface{})[0].(map[string]interface{})
							expectedMW := expectedProperties["maintenance_window"].([]interface{})[0].(map[string]interface{})

							assert.Equal(t, expectedMW["duration_hours"], resultMW["duration_hours"])
							assert.Equal(t, expectedMW["utc_offset"], resultMW["utc_offset"])
							assert.Equal(t, expectedMW["start_date"], resultMW["start_date"])
							assert.Equal(t, expectedMW["start_time"], resultMW["start_time"])

							if expectedMW["schedule"] != nil {
								assert.NotNil(t, resultMW["schedule"])
								resultSchedule := resultMW["schedule"].([]interface{})[0].(map[string]interface{})
								expectedSchedule := expectedMW["schedule"].([]interface{})[0].(map[string]interface{})

								if expectedSchedule["weekly"] != nil {
									assert.NotNil(t, resultSchedule["weekly"])
									resultWeekly := resultSchedule["weekly"].([]interface{})[0].(map[string]interface{})
									expectedWeekly := expectedSchedule["weekly"].([]interface{})[0].(map[string]interface{})

									assert.Equal(t, expectedWeekly["interval_weeks"], resultWeekly["interval_weeks"])
									assert.Equal(t, expectedWeekly["day_of_week"], resultWeekly["day_of_week"])
								}
							}
						}

						if expectedProperties["not_allowed_time"] != nil {
							assert.NotNil(t, resultProperties["not_allowed_time"])
							resultNAT := resultProperties["not_allowed_time"].([]interface{})
							expectedNAT := expectedProperties["not_allowed_time"].([]interface{})

							assert.Len(t, resultNAT, len(expectedNAT))

							for j, expectedTimeSpan := range expectedNAT {
								if j < len(resultNAT) {
									resultTimeSpan := resultNAT[j].(map[string]interface{})
									expectedTimeSpanMap := expectedTimeSpan.(map[string]interface{})

									assert.Equal(t, expectedTimeSpanMap["start"], resultTimeSpan["start"])
									assert.Equal(t, expectedTimeSpanMap["end"], resultTimeSpan["end"])
								}
							}
						}
					}
				}
			}
		})
	}
}

// Benchmark tests for AKS flatten functions
func BenchmarkFlattenAKSClusterMetadata(b *testing.B) {
	input := &AKSClusterMetadata{
		Name:    "benchmark-aks-cluster",
		Project: "benchmark-project",
		Labels: map[string]string{
			"env":     "benchmark",
			"version": "1.0",
			"team":    "platform",
		},
	}
	p := []interface{}{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		flattenAKSClusterMetadata(input, p)
	}
}

func BenchmarkFlattenAKSManagedCluster(b *testing.B) {
	input := &AKSManagedCluster{
		Location: "East US",
		Tags: map[string]interface{}{
			"Environment": "benchmark",
			"Team":        "platform",
		},
		Identity: &AKSManagedClusterIdentity{
			Type: "SystemAssigned",
		},
		Properties: &AKSManagedClusterProperties{
			DNSPrefix:         "benchmark-aks",
			KubernetesVersion: "1.25.6",
			NodeResourceGroup: "MC_benchmark-rg_benchmark-aks_eastus",
		},
	}
	p := []interface{}{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		flattenAKSManagedCluster(input, p)
	}
}
