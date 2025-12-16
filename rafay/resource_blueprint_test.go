package rafay

import (
	"context"
	"fmt"
	"testing"

	"github.com/RafaySystems/rafay-common/pkg/hub/client/options"
	v3 "github.com/RafaySystems/rafay-common/pkg/hub/client/typed/infra/v3"
	"github.com/RafaySystems/rafay-common/proto/types/hub/commonpb"
	"github.com/RafaySystems/rafay-common/proto/types/hub/infrapb"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockBlueprintClient is a mock of BlueprintClient interface
type MockBlueprintClient struct {
	mock.Mock
}

func (m *MockBlueprintClient) Apply(ctx context.Context, blueprint *infrapb.Blueprint, opts options.ApplyOptions) error {
	args := m.Called(ctx, blueprint, opts)
	return args.Error(0)
}

func (m *MockBlueprintClient) Get(ctx context.Context, opts options.GetOptions) (*infrapb.Blueprint, error) {
	args := m.Called(ctx, opts)
	if args.Get(0) != nil {
		return args.Get(0).(*infrapb.Blueprint), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockBlueprintClient) Delete(ctx context.Context, opts options.DeleteOptions) error {
	args := m.Called(ctx, opts)
	return args.Error(0)
}

func (m *MockBlueprintClient) List(ctx context.Context, opts options.ListOptions) (*infrapb.BlueprintList, error) {
	args := m.Called(ctx, opts)
	if args.Get(0) != nil {
		return args.Get(0).(*infrapb.BlueprintList), args.Error(1)
	}
	return nil, args.Error(1)
}

func TestResourceBlueprint(t *testing.T) {
	tests := []struct {
		name string
		run  func(*testing.T, *MockBlueprintClient)
	}{
		{"Create", testResourceBlueprintCreateHCL},
		{"Read", testResourceBlueprintReadHCL},
		{"Update", testResourceBlueprintUpdateHCL},
		{"Delete", testResourceBlueprintDeleteHCL},
		{"ReadComplex", testResourceBlueprintReadComplexHCL},
		{"ReadComplex2", testResourceBlueprintReadComplexHCL2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockBlueprintClient)

			oldGetBlueprintClient := getBlueprintClient
			defer func() { getBlueprintClient = oldGetBlueprintClient }()
			getBlueprintClient = func() (v3.BlueprintClient, error) {
				return mockClient, nil
			}

			tt.run(t, mockClient)

			// Reset mock expectations and calls after each test
			mockClient.ExpectedCalls = nil
			mockClient.Calls = nil
		})
	}
}

func testResourceBlueprintCreateHCL(t *testing.T, mockClient *MockBlueprintClient) {
	expectedBP := &infrapb.Blueprint{
		Metadata: &commonpb.Metadata{
			Name:    "test-blueprint-create",
			Project: "test-project",
		},
		Spec: &infrapb.BlueprintSpec{
			Version: "v1",
			Type:    "custom",
			DefaultAddons: &infrapb.DefaultAddons{
				EnableIngress: true,
			},
		},
	}

	mockClient.On("Apply", mock.Anything, mock.MatchedBy(func(blueprint *infrapb.Blueprint) bool {
		fmt.Println("I AM HERE MAN")
		return blueprint.Metadata.Name == "test-blueprint-create" && blueprint.Metadata.Project == "test-project"
	}), mock.Anything).Return(nil)
	mockClient.On("Get", mock.Anything, mock.MatchedBy(func(opts options.GetOptions) bool {
		return opts.Name == "test-blueprint-create" && opts.Project == "test-project"
	})).Return(expectedBP, nil)
	mockClient.On("Delete", mock.Anything, mock.Anything).Return(nil)

	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: map[string]func() (*schema.Provider, error){
			"rafay": func() (*schema.Provider, error) {
				return New("v1")(), nil
			},
		},
		Steps: []resource.TestStep{
			{
				Config: `
resource "rafay_blueprint" "tftest" {
  metadata {
    name    = "test-blueprint-create"
    project = "test-project"
  }
  spec {
    version = "v1"
    default_addons {
        enable_ingress = true
    }
  }
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rafay_blueprint.tftest", "metadata.0.name", "test-blueprint-create"),
					resource.TestCheckResourceAttr("rafay_blueprint.tftest", "spec.0.version", "v1"),
				),
			},
		},
	})
}

func testResourceBlueprintReadHCL(t *testing.T, mockClient *MockBlueprintClient) {
	expectedBP := &infrapb.Blueprint{
		Metadata: &commonpb.Metadata{
			Name:    "test-blueprint-read",
			Project: "test-project",
		},
		Spec: &infrapb.BlueprintSpec{
			Version: "v1",
		},
	}
	mockClient.On("Get", mock.Anything, mock.MatchedBy(func(opts options.GetOptions) bool {
		return opts.Name == "test-blueprint-read" && opts.Project == "test-project"
	})).Return(expectedBP, nil)
	mockClient.On("Delete", mock.Anything, mock.Anything).Return(nil)

	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: map[string]func() (*schema.Provider, error){
			"rafay": func() (*schema.Provider, error) {
				return New("v1")(), nil
			},
		},
		Steps: []resource.TestStep{
			{
				Config: `
resource "rafay_blueprint" "tftest" {
  metadata {
    name    = "test-blueprint-read"
    project = "test-project"
  }
  spec {
    version = "v1"
  }
}
`,
				ImportState:   true,
				ResourceName:  "rafay_blueprint.tftest",
				ImportStateId: "test-blueprint-read/test-project",
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rafay_blueprint.tftest", "spec.0.version", "v1"),
				),
			},
		},
	})
}

func testResourceBlueprintUpdateHCL(t *testing.T, mockClient *MockBlueprintClient) {
	expectedBPV1 := &infrapb.Blueprint{
		Metadata: &commonpb.Metadata{
			Name:    "test-blueprint-update",
			Project: "test-project",
		},
		Spec: &infrapb.BlueprintSpec{
			Version: "v1",
			Type:    "custom",
		},
	}
	expectedBPV2 := &infrapb.Blueprint{
		Metadata: &commonpb.Metadata{
			Name:    "test-blueprint-update",
			Project: "test-project",
		},
		Spec: &infrapb.BlueprintSpec{
			Version: "v2",
			Type:    "custom",
		},
	}

	mockClient.On("Apply", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockClient.On("Get", mock.Anything, mock.MatchedBy(func(opts options.GetOptions) bool {
		return opts.Name == "test-blueprint-update" && opts.Project == "test-project"
	})).Return(expectedBPV1, nil).Once()
	mockClient.On("Get", mock.Anything, mock.MatchedBy(func(opts options.GetOptions) bool {
		return opts.Name == "test-blueprint-update" && opts.Project == "test-project"
	})).Return(expectedBPV2, nil)
	mockClient.On("Delete", mock.Anything, mock.Anything).Return(nil)

	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: map[string]func() (*schema.Provider, error){
			"rafay": func() (*schema.Provider, error) {
				return New("v1")(), nil
			},
		},
		Steps: []resource.TestStep{
			{
				Config: `
resource "rafay_blueprint" "tftest" {
  metadata {
    name    = "test-blueprint-update"
    project = "test-project"
  }
  spec {
    version = "v1"
  }
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rafay_blueprint.tftest", "spec.0.version", "v1"),
				),
			},
			{
				Config: `
resource "rafay_blueprint" "tftest" {
  metadata {
    name    = "test-blueprint-update"
    project = "test-project"
  }
  spec {
    version = "v2"
  }
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rafay_blueprint.tftest", "spec.0.version", "v2"),
				),
			},
		},
	})
}

func testResourceBlueprintDeleteHCL(t *testing.T, mockClient *MockBlueprintClient) {
	expectedBP := &infrapb.Blueprint{
		Metadata: &commonpb.Metadata{
			Name:    "test-blueprint-delete",
			Project: "test-project",
		},
		Spec: &infrapb.BlueprintSpec{
			Version: "v1",
			Type:    "custom",
		},
	}

	mockClient.On("Apply", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockClient.On("Get", mock.Anything, mock.MatchedBy(func(opts options.GetOptions) bool {
		return opts.Name == "test-blueprint-delete" && opts.Project == "test-project"
	})).Return(expectedBP, nil)
	mockClient.On("Delete", mock.Anything, mock.MatchedBy(func(opts options.DeleteOptions) bool {
		return opts.Name == "test-blueprint-delete" && opts.Project == "test-project"
	})).Return(nil)

	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: map[string]func() (*schema.Provider, error){
			"rafay": func() (*schema.Provider, error) {
				return New("v1")(), nil
			},
		},
		Steps: []resource.TestStep{
			{
				Config: `
resource "rafay_blueprint" "tftest" {
  metadata {
    name    = "test-blueprint-delete"
    project = "test-project"
  }
  spec {
    version = "v1"
  }
}
`,
			},
		},
	})
}

func testResourceBlueprintReadComplexHCL(t *testing.T, mockClient *MockBlueprintClient) {
	expectedBP := &infrapb.Blueprint{
		Metadata: &commonpb.Metadata{
			Name:    "custom-blueprint",
			Project: "terraform",
		},
		Spec: &infrapb.BlueprintSpec{
			Version: "v0",
			Base: &infrapb.BlueprintBase{
				Name:    "default",
				Version: "1.16.0",
			},
			NamespaceConfig: &infrapb.NsConfig{
				SyncType:   "managed",
				EnableSync: true,
			},
			DefaultAddons: &infrapb.DefaultAddons{
				EnableIngress:                    true,
				EnableCsiSecretStore:             true,
				EnableMonitoring:                 true,
				EnableVM:                         false,
				DisableAwsNodeTerminationHandler: true,
				CsiSecretStoreConfig: &infrapb.CsiSecretStoreConfig{
					EnableSecretRotation: true,
					SyncSecrets:          true,
					RotationPollInterval: "2m",
					Providers: &infrapb.SecretStoreProvider{
						Aws: true,
					},
				},
				Monitoring: &infrapb.MonitoringConfig{
					MetricsServer: &infrapb.MonitoringComponent{
						Enabled:   true,
						Discovery: &infrapb.MonitoringDiscoveryConfig{},
					},
					HelmExporter: &infrapb.MonitoringComponent{
						Enabled: true,
					},
					KubeStateMetrics: &infrapb.MonitoringComponent{
						Enabled: true,
					},
					NodeExporter: &infrapb.MonitoringComponent{
						Enabled: true,
					},
					PrometheusAdapter: &infrapb.MonitoringComponent{
						Enabled: true,
					},
					Resources: &commonpb.ResourceRequirements{
						Limits: &commonpb.ResourceQuantity{
							Memory: "200Mi",
							Cpu:    "100m",
						},
					},
				},
			},
			Drift: &commonpb.DriftSpec{
				Action:  "Deny",
				Enabled: true,
			},
			DriftWebhook: &infrapb.DriftWebhook{
				Enabled: true,
			},
			Placement: &infrapb.BlueprintPlacement{
				AutoPublish: false,
			},
		},
	}

	mockClient.On("Get", mock.Anything, mock.MatchedBy(func(opts options.GetOptions) bool {
		return opts.Name == "custom-blueprint" && opts.Project == "terraform"
	})).Return(expectedBP, nil)

	resourceSchema := resourceBluePrint().Schema
	resourceDataMap := map[string]interface{}{
		"metadata": []interface{}{
			map[string]interface{}{
				"name":    "custom-blueprint",
				"project": "terraform",
			},
		},
		"spec": []interface{}{
			map[string]interface{}{
				"version": "v0",
				"base": []interface{}{
					map[string]interface{}{
						"name":    "default",
						"version": "1.16.0",
					},
				},
				"namespace_config": []interface{}{
					map[string]interface{}{
						"sync_type":   "managed",
						"enable_sync": true,
					},
				},
				"default_addons": []interface{}{
					map[string]interface{}{
						"enable_ingress":                       true,
						"enable_csi_secret_store":              true,
						"enable_monitoring":                    true,
						"enable_vm":                            false,
						"disable_aws_node_termination_handler": true,
						"csi_secret_store_config": []interface{}{
							map[string]interface{}{
								"enable_secret_rotation": true,
								"sync_secrets":           true,
								"rotation_poll_interval": "2m",
								"providers": []interface{}{
									map[string]interface{}{
										"aws": true,
									},
								},
							},
						},
						"monitoring": []interface{}{
							map[string]interface{}{
								"metrics_server": []interface{}{
									map[string]interface{}{
										"enabled":   true,
										"discovery": []interface{}{map[string]interface{}{}},
									},
								},
								"helm_exporter": []interface{}{
									map[string]interface{}{
										"enabled": true,
									},
								},
								"kube_state_metrics": []interface{}{
									map[string]interface{}{
										"enabled": true,
									},
								},
								"node_exporter": []interface{}{
									map[string]interface{}{
										"enabled": true,
									},
								},
								"prometheus_adapter": []interface{}{
									map[string]interface{}{
										"enabled": true,
									},
								},
								"resources": []interface{}{
									map[string]interface{}{
										"limits": []interface{}{
											map[string]interface{}{
												"memory": "200Mi",
												"cpu":    "100m",
											},
										},
									},
								},
							},
						},
					},
				},
				"drift": []interface{}{
					map[string]interface{}{
						"action":  "Deny",
						"enabled": true,
					},
				},
				"drift_webhook": []interface{}{
					map[string]interface{}{
						"enabled": true,
					},
				},
				"placement": []interface{}{
					map[string]interface{}{
						"auto_publish": false,
					},
				},
			},
		},
	}
	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId("custom-blueprint")

	ctx := context.Background()
	diags := resourceBluePrintRead(ctx, d, nil)

	assert.False(t, diags.HasError())
	mockClient.AssertExpectations(t)

	// Verify Metadata
	metadata := d.Get("metadata").([]interface{})
	assert.Len(t, metadata, 1)
	mMap := metadata[0].(map[string]interface{})
	assert.Equal(t, "custom-blueprint", mMap["name"])
	assert.Equal(t, "terraform", mMap["project"])

	// Verify Spec Top Level
	spec := d.Get("spec").([]interface{})
	assert.Len(t, spec, 1)
	sMap := spec[0].(map[string]interface{})
	assert.Equal(t, "v0", sMap["version"])

	// Verify Base
	base := sMap["base"].([]interface{})
	assert.Len(t, base, 1)
	bMap := base[0].(map[string]interface{})
	assert.Equal(t, "default", bMap["name"])
	assert.Equal(t, "1.16.0", bMap["version"])

	// Verify Namespace Config
	nc := sMap["namespace_config"].([]interface{})
	assert.Len(t, nc, 1)
	ncMap := nc[0].(map[string]interface{})
	assert.Equal(t, "managed", ncMap["sync_type"])
	assert.Equal(t, true, ncMap["enable_sync"])

	// Verify Default Addons
	da := sMap["default_addons"].([]interface{})
	assert.Len(t, da, 1)
	daMap := da[0].(map[string]interface{})
	assert.Equal(t, true, daMap["enable_ingress"])
	assert.Equal(t, true, daMap["enable_csi_secret_store"])
	assert.Equal(t, true, daMap["enable_monitoring"])
	assert.Equal(t, false, daMap["enable_vm"])
	assert.Equal(t, true, daMap["disable_aws_node_termination_handler"])

	// Verify CSI Secret Store Config
	css := daMap["csi_secret_store_config"].([]interface{})
	assert.Len(t, css, 1)
	cssMap := css[0].(map[string]interface{})
	assert.Equal(t, true, cssMap["enable_secret_rotation"])
	assert.Equal(t, true, cssMap["sync_secrets"])
	assert.Equal(t, "2m", cssMap["rotation_poll_interval"])

	// Verify Monitoring
	mon := daMap["monitoring"].([]interface{})
	assert.Len(t, mon, 1)
	monMap := mon[0].(map[string]interface{})

	ms := monMap["metrics_server"].([]interface{})
	assert.Len(t, ms, 1)
	assert.Equal(t, true, ms[0].(map[string]interface{})["enabled"])

	he := monMap["helm_exporter"].([]interface{})
	assert.Len(t, he, 1)
	assert.Equal(t, true, he[0].(map[string]interface{})["enabled"])

	res := monMap["resources"].([]interface{})
	assert.Len(t, res, 1)
	resMap := res[0].(map[string]interface{})
	limits := resMap["limits"].([]interface{})
	assert.Len(t, limits, 1)
	lMap := limits[0].(map[string]interface{})
	assert.Equal(t, "200Mi", lMap["memory"])
	assert.Equal(t, "100m", lMap["cpu"])

	// Verify Drift
	drift := sMap["drift"].([]interface{})
	assert.Len(t, drift, 1)
	drMap := drift[0].(map[string]interface{})
	assert.Equal(t, "Deny", drMap["action"])
	assert.Equal(t, true, drMap["enabled"])

	// Verify Drift Webhook
	dw := sMap["drift_webhook"].([]interface{})
	assert.Len(t, dw, 1)
	dwMap := dw[0].(map[string]interface{})
	assert.Equal(t, true, dwMap["enabled"])

	// Verify Placement
	pl := sMap["placement"].([]interface{})
	assert.Len(t, pl, 1)
	plMap := pl[0].(map[string]interface{})
	assert.Equal(t, false, plMap["auto_publish"])
}

func testResourceBlueprintReadComplexHCL2(t *testing.T, mockClient *MockBlueprintClient) {
	expectedBP := &infrapb.Blueprint{
		Metadata: &commonpb.Metadata{
			Name:    "custom-blueprint",
			Project: "terraform",
		},
		Spec: &infrapb.BlueprintSpec{
			Version: "v0",
			Type:    "custom",
			Base: &infrapb.BlueprintBase{
				Name:    "default",
				Version: "1.16.0",
			},
			NamespaceConfig: &infrapb.NsConfig{
				SyncType:   "managed",
				EnableSync: true,
			},
			DefaultAddons: &infrapb.DefaultAddons{
				EnableIngress:                    true,
				EnableCsiSecretStore:             true,
				EnableMonitoring:                 true,
				EnableVM:                         false,
				DisableAwsNodeTerminationHandler: true,
				CsiSecretStoreConfig: &infrapb.CsiSecretStoreConfig{
					EnableSecretRotation: true,
					SyncSecrets:          true,
					RotationPollInterval: "2m",
					Providers: &infrapb.SecretStoreProvider{
						Aws: true,
					},
				},
				Monitoring: &infrapb.MonitoringConfig{
					MetricsServer: &infrapb.MonitoringComponent{
						Enabled: true,
						Discovery: &infrapb.MonitoringDiscoveryConfig{
							Namespace: "rafay-system",
						},
					},
					HelmExporter: &infrapb.MonitoringComponent{
						Enabled: true,
					},
					KubeStateMetrics: &infrapb.MonitoringComponent{
						Enabled: true,
					},
					NodeExporter: &infrapb.MonitoringComponent{
						Enabled: true,
					},
					PrometheusAdapter: &infrapb.MonitoringComponent{
						Enabled: true,
					},
					Resources: &commonpb.ResourceRequirements{
						Limits: &commonpb.ResourceQuantity{
							Memory: "300Mi",
							Cpu:    "100m",
						},
					},
				},
			},
			Drift: &commonpb.DriftSpec{
				Action:  "Deny",
				Enabled: true,
			},
			DriftWebhook: &infrapb.DriftWebhook{
				Enabled: true,
			},
			Placement: &infrapb.BlueprintPlacement{
				AutoPublish: false,
			},
		},
	}

	mockClient.On("Get", mock.Anything, mock.MatchedBy(func(opts options.GetOptions) bool {
		return opts.Name == "custom-blueprint" && opts.Project == "terraform"
	})).Return(expectedBP, nil)

	mockClient.On("Delete", mock.Anything, mock.Anything).Return(nil)

	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: map[string]func() (*schema.Provider, error){
			"rafay": func() (*schema.Provider, error) {
				return New("v1")(), nil
			},
		},
		Steps: []resource.TestStep{
			{
				Config: `
provider "rafay" {
  ignore_insecure_tls_error = true
  api_key                  = "test-api-key"
  rest_endpoint            = "https://test-endpoint"
  project                  = "terraform"
}

resource "rafay_blueprint" "blueprint" {
  metadata {
    name    = "custom-blueprint"
    project = "terraform"
  }
  spec {
    version = "v0"
    base {
      name    = "default"
      version = "1.16.0"
    }
    namespace_config {
      sync_type   = "managed"
      enable_sync = true
    }
    default_addons {
      enable_ingress          = true
      enable_csi_secret_store = true
      enable_monitoring       = true
      enable_vm               = false
      disable_aws_node_termination_handler = true

      csi_secret_store_config {
        enable_secret_rotation = true
        sync_secrets           = true
        rotation_poll_interval = "2m"
        providers {
          aws = true
        }
      }
      monitoring {
        metrics_server {
          enabled = true
          discovery {
            namespace = "rafay-system"
          }
        }
        helm_exporter {
          enabled = true
        }
        kube_state_metrics {
          enabled = true
        }
        node_exporter {
          enabled = true
        }
        prometheus_adapter {
          enabled = true
        }
        resources {
          limits {
            memory = "200Mi"
            cpu    = "100m"
          }
        }
      }
    }
    drift {
      action  = "Deny"
      enabled = true
    }
    drift_webhook {
      enabled = true
    }
    placement {
      auto_publish = false
    }
  }
}
`,
				ImportState:   true,
				ResourceName:  "rafay_blueprint.blueprint",
				ImportStateId: "custom-blueprint/terraform",
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rafay_blueprint.blueprint", "metadata.0.name", "custom-blueprint"),
					resource.TestCheckResourceAttr("rafay_blueprint.blueprint", "spec.0.version", "v0"),
					resource.TestCheckResourceAttr("rafay_blueprint.blueprint", "spec.0.base.0.name", "default"),
					resource.TestCheckResourceAttr("rafay_blueprint.blueprint", "spec.0.base.0.version", "1.16.0"),
					resource.TestCheckResourceAttr("rafay_blueprint.blueprint", "spec.0.default_addons.0.enable_ingress", "true"),
					resource.TestCheckResourceAttr("rafay_blueprint.blueprint", "spec.0.default_addons.0.monitoring.0.resources.0.limits.0.memory", "300Mi"),
					resource.TestCheckResourceAttr("rafay_blueprint.blueprint", "spec.0.drift.0.action", "Deny"),
				),
			},
		},
	})
}
