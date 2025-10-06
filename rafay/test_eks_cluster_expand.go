package rafay

import (
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestExpandEKSCluster tests the expandEKSCluster function
func TestExpandEKSCluster(t *testing.T) {
	tests := []struct {
		name     string
		input    []interface{}
		expected *EKSCluster
	}{
		{
			name:     "empty input",
			input:    []interface{}{},
			expected: &EKSCluster{},
		},
		{
			name:     "nil input",
			input:    nil,
			expected: &EKSCluster{},
		},
		{
			name: "complete cluster",
			input: []interface{}{
				map[string]interface{}{
					"kind": "Cluster",
					"metadata": []interface{}{
						map[string]interface{}{
							"name":    "test-cluster",
							"project": "test-project",
							"labels": map[string]interface{}{
								"env": "test",
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
						},
					},
				},
			},
			expected: &EKSCluster{
				Kind: "Cluster",
				Metadata: &EKSClusterMetadata{
					Name:    "test-cluster",
					Project: "test-project",
					Labels: map[string]string{
						"env": "test",
					},
				},
				Spec: &EKSSpec{
					Type:                "eks",
					Blueprint:           "minimal",
					BlueprintVersion:    "1.0",
					CloudProvider:       "aws",
					CrossAccountRoleArn: "arn:aws:iam::123456789012:role/test-role",
					CniProvider:         "aws-cni",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expandEKSCluster(tt.input)
			assert.Equal(t, tt.expected.Kind, result.Kind)

			if tt.expected.Metadata != nil {
				require.NotNil(t, result.Metadata)
				assert.Equal(t, tt.expected.Metadata.Name, result.Metadata.Name)
				assert.Equal(t, tt.expected.Metadata.Project, result.Metadata.Project)
				assert.Equal(t, tt.expected.Metadata.Labels, result.Metadata.Labels)
			}

			if tt.expected.Spec != nil {
				require.NotNil(t, result.Spec)
				assert.Equal(t, tt.expected.Spec.Type, result.Spec.Type)
				assert.Equal(t, tt.expected.Spec.Blueprint, result.Spec.Blueprint)
				assert.Equal(t, tt.expected.Spec.CloudProvider, result.Spec.CloudProvider)
			}
		})
	}
}

// TestExpandEKSMetaMetadata tests the expandEKSMetaMetadata function
func TestExpandEKSMetaMetadata(t *testing.T) {
	tests := []struct {
		name     string
		input    []interface{}
		expected *EKSClusterMetadata
	}{
		{
			name:     "empty input",
			input:    []interface{}{},
			expected: &EKSClusterMetadata{},
		},
		{
			name: "complete metadata",
			input: []interface{}{
				map[string]interface{}{
					"name":    "test-cluster",
					"project": "test-project",
					"labels": map[string]interface{}{
						"env":     "test",
						"version": "1.0",
					},
				},
			},
			expected: &EKSClusterMetadata{
				Name:    "test-cluster",
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
					"name": "test-cluster",
				},
			},
			expected: &EKSClusterMetadata{
				Name: "test-cluster",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expandEKSMetaMetadata(tt.input)
			assert.Equal(t, tt.expected.Name, result.Name)
			assert.Equal(t, tt.expected.Project, result.Project)
			assert.Equal(t, tt.expected.Labels, result.Labels)
		})
	}
}

// TestExpandEKSClusterConfig tests the expandEKSClusterConfig function
func TestExpandEKSClusterConfig(t *testing.T) {
	tests := []struct {
		name      string
		input     []interface{}
		rawConfig cty.Value
		expected  *EKSClusterConfig
		expectErr bool
	}{
		{
			name:      "empty input",
			input:     []interface{}{},
			rawConfig: cty.NullVal(cty.Object(map[string]cty.Type{})),
			expected:  &EKSClusterConfig{},
			expectErr: false,
		},
		{
			name: "complete config",
			input: []interface{}{
				map[string]interface{}{
					"kind":       "ClusterConfig",
					"apiversion": "eksctl.io/v1alpha5",
					"metadata": []interface{}{
						map[string]interface{}{
							"name":    "test-cluster",
							"region":  "us-west-2",
							"version": "1.21",
							"tags": map[string]interface{}{
								"Environment": "test",
							},
						},
					},
					"kubernetes_network_config": []interface{}{
						map[string]interface{}{
							"ip_family":         "IPv4",
							"service_ipv4_cidr": "10.100.0.0/16",
						},
					},
					"availability_zones": []interface{}{"us-west-2a", "us-west-2b"},
				},
			},
			rawConfig: cty.ObjectVal(map[string]cty.Value{
				"kind":       cty.StringVal("ClusterConfig"),
				"apiversion": cty.StringVal("eksctl.io/v1alpha5"),
			}),
			expected: &EKSClusterConfig{
				Kind:       "ClusterConfig",
				APIVersion: "eksctl.io/v1alpha5",
				Metadata: &EKSClusterConfigMetadata{
					Name:    "test-cluster",
					Region:  "us-west-2",
					Version: "1.21",
					Tags: map[string]string{
						"Environment": "test",
					},
				},
				KubernetesNetworkConfig: &KubernetesNetworkConfig{
					IPFamily:        "IPv4",
					ServiceIPv4CIDR: "10.100.0.0/16",
				},
				AvailabilityZones: []string{"us-west-2a", "us-west-2b"},
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := expandEKSClusterConfig(tt.input, tt.rawConfig)

			if tt.expectErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected.Kind, result.Kind)
			assert.Equal(t, tt.expected.APIVersion, result.APIVersion)

			if tt.expected.Metadata != nil {
				require.NotNil(t, result.Metadata)
				assert.Equal(t, tt.expected.Metadata.Name, result.Metadata.Name)
				assert.Equal(t, tt.expected.Metadata.Region, result.Metadata.Region)
				assert.Equal(t, tt.expected.Metadata.Version, result.Metadata.Version)
				assert.Equal(t, tt.expected.Metadata.Tags, result.Metadata.Tags)
			}

			if tt.expected.KubernetesNetworkConfig != nil {
				require.NotNil(t, result.KubernetesNetworkConfig)
				assert.Equal(t, tt.expected.KubernetesNetworkConfig.IPFamily, result.KubernetesNetworkConfig.IPFamily)
				assert.Equal(t, tt.expected.KubernetesNetworkConfig.ServiceIPv4CIDR, result.KubernetesNetworkConfig.ServiceIPv4CIDR)
			}

			assert.Equal(t, tt.expected.AvailabilityZones, result.AvailabilityZones)
		})
	}
}

// TestExpandAccessConfig tests the expandAccessConfig function
func TestExpandAccessConfig(t *testing.T) {
	tests := []struct {
		name     string
		input    []interface{}
		expected *EKSClusterAccess
	}{
		{
			name:     "empty input",
			input:    []interface{}{},
			expected: &EKSClusterAccess{},
		},
		{
			name: "complete access config",
			input: []interface{}{
				map[string]interface{}{
					"bootstrap_cluster_creator_admin_permissions": true,
					"authentication_mode":                         "API_AND_CONFIG_MAP",
					"access_entries": []interface{}{
						map[string]interface{}{
							"principal_arn": "arn:aws:iam::123456789012:user/test-user",
							"type":          "STANDARD",
							"access_policies": []interface{}{
								map[string]interface{}{
									"policy_arn": "arn:aws:eks::aws:cluster-access-policy/AmazonEKSViewPolicy",
									"access_scope": []interface{}{
										map[string]interface{}{
											"type": "cluster",
										},
									},
								},
							},
						},
					},
				},
			},
			expected: &EKSClusterAccess{
				BootstrapClusterCreatorAdminPermissions: true,
				AuthenticationMode:                      "API_AND_CONFIG_MAP",
				AccessEntries: []*EKSAccessEntry{
					{
						PrincipalARN: "arn:aws:iam::123456789012:user/test-user",
						Type:         "STANDARD",
						AccessPolicies: []*EKSAccessPolicy{
							{
								PolicyARN: "arn:aws:eks::aws:cluster-access-policy/AmazonEKSViewPolicy",
								AccessScope: &EKSAccessScope{
									Type: "cluster",
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
			result := expandAccessConfig(tt.input)
			assert.Equal(t, tt.expected.BootstrapClusterCreatorAdminPermissions, result.BootstrapClusterCreatorAdminPermissions)
			assert.Equal(t, tt.expected.AuthenticationMode, result.AuthenticationMode)

			if tt.expected.AccessEntries != nil {
				require.NotNil(t, result.AccessEntries)
				assert.Len(t, result.AccessEntries, len(tt.expected.AccessEntries))

				for i, expectedEntry := range tt.expected.AccessEntries {
					assert.Equal(t, expectedEntry.PrincipalARN, result.AccessEntries[i].PrincipalARN)
					assert.Equal(t, expectedEntry.Type, result.AccessEntries[i].Type)

					if expectedEntry.AccessPolicies != nil {
						require.NotNil(t, result.AccessEntries[i].AccessPolicies)
						assert.Len(t, result.AccessEntries[i].AccessPolicies, len(expectedEntry.AccessPolicies))

						for j, expectedPolicy := range expectedEntry.AccessPolicies {
							assert.Equal(t, expectedPolicy.PolicyARN, result.AccessEntries[i].AccessPolicies[j].PolicyARN)

							if expectedPolicy.AccessScope != nil {
								require.NotNil(t, result.AccessEntries[i].AccessPolicies[j].AccessScope)
								assert.Equal(t, expectedPolicy.AccessScope.Type, result.AccessEntries[i].AccessPolicies[j].AccessScope.Type)
							}
						}
					}
				}
			}
		})
	}
}

// TestExpandVPC tests the expandVPC function
func TestExpandVPC(t *testing.T) {
	tests := []struct {
		name      string
		input     []interface{}
		rawConfig cty.Value
		expected  *EKSClusterVPC
		expectErr bool
	}{
		{
			name:      "empty input",
			input:     []interface{}{},
			rawConfig: cty.NullVal(cty.Object(map[string]cty.Type{})),
			expected:  &EKSClusterVPC{},
			expectErr: false,
		},
		{
			name: "complete vpc config",
			input: []interface{}{
				map[string]interface{}{
					"id":   "vpc-12345678",
					"cidr": "10.0.0.0/16",
					"subnets": []interface{}{
						map[string]interface{}{
							"private": map[string]interface{}{
								"us-west-2a": map[string]interface{}{
									"id": "subnet-private-1",
								},
								"us-west-2b": map[string]interface{}{
									"id": "subnet-private-2",
								},
							},
							"public": map[string]interface{}{
								"us-west-2a": map[string]interface{}{
									"id": "subnet-public-1",
								},
								"us-west-2b": map[string]interface{}{
									"id": "subnet-public-2",
								},
							},
						},
					},
					"nat": []interface{}{
						map[string]interface{}{
							"gateway": "Single",
						},
					},
					"cluster_endpoints": []interface{}{
						map[string]interface{}{
							"private_access": true,
							"public_access":  false,
						},
					},
				},
			},
			rawConfig: cty.ObjectVal(map[string]cty.Value{
				"id":   cty.StringVal("vpc-12345678"),
				"cidr": cty.StringVal("10.0.0.0/16"),
			}),
			expected: &EKSClusterVPC{
				ID:   "vpc-12345678",
				CIDR: "10.0.0.0/16",
				Subnets: &ClusterSubnets{
					Private: AZSubnetMapping{
						"us-west-2a": AZSubnetSpec{ID: "subnet-private-1"},
						"us-west-2b": AZSubnetSpec{ID: "subnet-private-2"},
					},
					Public: AZSubnetMapping{
						"us-west-2a": AZSubnetSpec{ID: "subnet-public-1"},
						"us-west-2b": AZSubnetSpec{ID: "subnet-public-2"},
					},
				},
				NAT: &ClusterNAT{
					Gateway: "Single",
				},
				ClusterEndpoints: &ClusterEndpoints{
					PrivateAccess: &[]bool{true}[0],
					PublicAccess:  &[]bool{false}[0],
				},
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := expandVPC(tt.input, tt.rawConfig)

			if tt.expectErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected.ID, result.ID)
			assert.Equal(t, tt.expected.CIDR, result.CIDR)

			if tt.expected.Subnets != nil {
				require.NotNil(t, result.Subnets)
				assert.Equal(t, tt.expected.Subnets.Private, result.Subnets.Private)
				assert.Equal(t, tt.expected.Subnets.Public, result.Subnets.Public)
			}

			if tt.expected.NAT != nil {
				require.NotNil(t, result.NAT)
				assert.Equal(t, tt.expected.NAT.Gateway, result.NAT.Gateway)
			}

			if tt.expected.ClusterEndpoints != nil {
				require.NotNil(t, result.ClusterEndpoints)
				assert.Equal(t, tt.expected.ClusterEndpoints.PrivateAccess, result.ClusterEndpoints.PrivateAccess)
				assert.Equal(t, tt.expected.ClusterEndpoints.PublicAccess, result.ClusterEndpoints.PublicAccess)
			}
		})
	}
}

// TestExpandNodeGroups tests the expandNodeGroups function
func TestExpandNodeGroups(t *testing.T) {
	tests := []struct {
		name     string
		input    []interface{}
		expected []*NodeGroup
	}{
		{
			name:     "empty input",
			input:    []interface{}{},
			expected: []*NodeGroup{},
		},
		{
			name: "single node group",
			input: []interface{}{
				map[string]interface{}{
					"name":             "worker-nodes",
					"instance_type":    "m5.large",
					"desired_capacity": 2,
					"min_size":         1,
					"max_size":         4,
					"volume_size":      20,
					"ami_family":       "AmazonLinux2",
					"ssh": []interface{}{
						map[string]interface{}{
							"allow":                     true,
							"public_key_name":           "my-key",
							"source_security_group_ids": []interface{}{"sg-12345678"},
						},
					},
					"iam": []interface{}{
						map[string]interface{}{
							"attach_policy_arns": []interface{}{
								"arn:aws:iam::aws:policy/AmazonEKSWorkerNodePolicy",
								"arn:aws:iam::aws:policy/AmazonEKS_CNI_Policy",
								"arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly",
							},
						},
					},
					"tags": map[string]interface{}{
						"Environment": "test",
						"NodeGroup":   "worker-nodes",
					},
				},
			},
			expected: []*NodeGroup{
				{
					VolumeSize: &[]int{20}[0],
					AMIFamily:  "AmazonLinux2",
					SSH: &NodeGroupSSH{
						Allow:                  &[]bool{true}[0],
						PublicKeyName:          "my-key",
						SourceSecurityGroupIDs: []string{"sg-12345678"},
					},
					IAM: &NodeGroupIAM{
						AttachPolicyARNs: []string{
							"arn:aws:iam::aws:policy/AmazonEKSWorkerNodePolicy",
							"arn:aws:iam::aws:policy/AmazonEKS_CNI_Policy",
							"arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly",
						},
					},
					Tags: map[string]string{
						"Environment": "test",
						"NodeGroup":   "worker-nodes",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expandNodeGroups(tt.input)
			assert.Len(t, result, len(tt.expected))

			for i, expected := range tt.expected {
				if i < len(result) {
					assert.Equal(t, expected.Name, result[i].Name)
					assert.Equal(t, expected.InstanceType, result[i].InstanceType)
					assert.Equal(t, expected.AMIFamily, result[i].AMIFamily)

					// Check scaling config fields directly on NodeGroup
					if expected.DesiredCapacity != nil {
						require.NotNil(t, result[i].DesiredCapacity)
						assert.Equal(t, *expected.DesiredCapacity, *result[i].DesiredCapacity)
					}
					if expected.MinSize != nil {
						require.NotNil(t, result[i].MinSize)
						assert.Equal(t, *expected.MinSize, *result[i].MinSize)
					}
					if expected.MaxSize != nil {
						require.NotNil(t, result[i].MaxSize)
						assert.Equal(t, *expected.MaxSize, *result[i].MaxSize)
					}

					assert.Equal(t, expected.VolumeSize, result[i].VolumeSize)
					assert.Equal(t, expected.Tags, result[i].Tags)
				}
			}
		})
	}
}

// TestExpandIAMFields tests the expandIAMFields function
func TestExpandIAMFields(t *testing.T) {
	tests := []struct {
		name     string
		input    []interface{}
		expected *EKSClusterIAM
	}{
		{
			name:     "empty input",
			input:    []interface{}{},
			expected: &EKSClusterIAM{},
		},
		{
			name: "complete iam config",
			input: []interface{}{
				map[string]interface{}{
					"with_oidc": true,
					"service_accounts": []interface{}{
						map[string]interface{}{
							"metadata": []interface{}{
								map[string]interface{}{
									"name":      "aws-load-balancer-controller",
									"namespace": "kube-system",
									"labels": map[string]interface{}{
										"app.kubernetes.io/name": "aws-load-balancer-controller",
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
								},
							},
						},
					},
					"pod_identity_associations": []interface{}{
						map[string]interface{}{
							"namespace":         "default",
							"service_account":   "my-service-account",
							"role_arn":          "arn:aws:iam::123456789012:role/my-pod-role",
							"permission_policy": `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Action":"s3:GetObject","Resource":"*"}]}`,
						},
					},
				},
			},
			expected: &EKSClusterIAM{
				WithOIDC: &[]bool{true}[0],
				ServiceAccounts: []*EKSClusterIAMServiceAccount{
					{
						Metadata: &EKSClusterIAMMeta{
							Name:      "aws-load-balancer-controller",
							Namespace: "kube-system",
							Labels: map[string]string{
								"app.kubernetes.io/name": "aws-load-balancer-controller",
							},
						},
						AttachPolicyARNs: []string{
							"arn:aws:iam::123456789012:policy/AWSLoadBalancerControllerIAMPolicy",
						},
						AttachPolicy: map[string]interface{}{
							"Statement": []map[string]interface{}{
								{
									"Effect": "Allow",
									"Action": []string{
										"ec2:DescribeVpcs",
										"ec2:DescribeSubnets",
									},
									"Resource": "*",
								},
							},
						},
						WellKnownPolicies: &WellKnownPolicies{
							AutoScaler:                &[]bool{true}[0],
							AWSLoadBalancerController: &[]bool{true}[0],
						},
					},
				},
				PodIdentityAssociations: []*IAMPodIdentityAssociation{
					{
						Namespace:          "default",
						ServiceAccountName: "my-service-account",
						RoleARN:            "arn:aws:iam::123456789012:role/my-pod-role",
						PermissionPolicy: map[string]interface{}{
							"Version": "2012-10-17",
							"Statement": []interface{}{
								map[string]interface{}{
									"Effect":   "Allow",
									"Action":   "s3:GetObject",
									"Resource": "*",
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
			result := expandIAMFields(tt.input)
			assert.Equal(t, tt.expected.WithOIDC, result.WithOIDC)

			if tt.expected.ServiceAccounts != nil {
				require.NotNil(t, result.ServiceAccounts)
				assert.Len(t, result.ServiceAccounts, len(tt.expected.ServiceAccounts))

				for i, expectedSA := range tt.expected.ServiceAccounts {
					if i < len(result.ServiceAccounts) {
						if expectedSA.Metadata != nil {
							require.NotNil(t, result.ServiceAccounts[i].Metadata)
							assert.Equal(t, expectedSA.Metadata.Name, result.ServiceAccounts[i].Metadata.Name)
							assert.Equal(t, expectedSA.Metadata.Namespace, result.ServiceAccounts[i].Metadata.Namespace)
							assert.Equal(t, expectedSA.Metadata.Labels, result.ServiceAccounts[i].Metadata.Labels)
						}

						assert.Equal(t, expectedSA.AttachPolicyARNs, result.ServiceAccounts[i].AttachPolicyARNs)

						if expectedSA.WellKnownPolicies != nil {
							require.NotNil(t, result.ServiceAccounts[i].WellKnownPolicies)
							assert.Equal(t, expectedSA.WellKnownPolicies.AutoScaler, result.ServiceAccounts[i].WellKnownPolicies.AutoScaler)
							assert.Equal(t, expectedSA.WellKnownPolicies.AWSLoadBalancerController, result.ServiceAccounts[i].WellKnownPolicies.AWSLoadBalancerController)
						}
					}
				}
			}

			if tt.expected.PodIdentityAssociations != nil {
				require.NotNil(t, result.PodIdentityAssociations)
				assert.Len(t, result.PodIdentityAssociations, len(tt.expected.PodIdentityAssociations))

				for i, expectedPIA := range tt.expected.PodIdentityAssociations {
					if i < len(result.PodIdentityAssociations) {
						assert.Equal(t, expectedPIA.Namespace, result.PodIdentityAssociations[i].Namespace)
						assert.Equal(t, expectedPIA.ServiceAccountName, result.PodIdentityAssociations[i].ServiceAccountName)
						assert.Equal(t, expectedPIA.RoleARN, result.PodIdentityAssociations[i].RoleARN)
						assert.Equal(t, expectedPIA.PermissionPolicy, result.PodIdentityAssociations[i].PermissionPolicy)
					}
				}
			}
		})
	}
}

// Benchmark tests
func BenchmarkExpandEKSCluster(b *testing.B) {
	input := []interface{}{
		map[string]interface{}{
			"kind": "Cluster",
			"metadata": []interface{}{
				map[string]interface{}{
					"name":    "benchmark-cluster",
					"project": "benchmark-project",
					"labels": map[string]interface{}{
						"env": "benchmark",
					},
				},
			},
			"spec": []interface{}{
				map[string]interface{}{
					"type":          "eks",
					"blueprint":     "minimal",
					"cloudprovider": "aws",
				},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		expandEKSCluster(input)
	}
}

func BenchmarkExpandEKSMetaMetadata(b *testing.B) {
	input := []interface{}{
		map[string]interface{}{
			"name":    "benchmark-cluster",
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
		expandEKSMetaMetadata(input)
	}
}
