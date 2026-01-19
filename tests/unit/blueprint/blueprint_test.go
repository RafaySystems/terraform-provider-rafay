package blueprint

import (
	"embed"
	"fmt"
	"testing"

	"github.com/RafaySystems/rafay-common/pkg/hub/client/options"
	"github.com/RafaySystems/rafay-common/proto/types/hub/infrapb"
	"github.com/RafaySystems/terraform-provider-rafay/tests/helpers"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/stretchr/testify/mock"
)

//go:embed testdata/*.tf
var blueprintFixtures embed.FS

func complexBlueprintConfig(t *testing.T, memory string) string {
	fixture := helpers.LoadFixture(t, blueprintFixtures, "testdata/complex_blueprint.tf")
	return fmt.Sprintf(fixture, memory)
}

func TestResourceBlueprint(t *testing.T) {
	tests := []struct {
		name string
		run  func(*testing.T, blueprintTestConfig)
	}{
		{"Create", testResourceBlueprintCreateHCL},
		{"Read", testResourceBlueprintReadHCL},
		{"Update", testResourceBlueprintUpdateHCL},
		{"Delete", testResourceBlueprintDeleteHCL},
		{"ReadComplex", testResourceBlueprintReadComplexHCL},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt := tt
			tt.run(t, newBlueprintTestConfig())
		})
	}
}

func testResourceBlueprintCreateHCL(t *testing.T, cfg blueprintTestConfig) {
	expectedBP := mustBlueprintFromJSON(t, `
{
  "metadata": {
    "name": "test-blueprint-create",
    "project": "test-project"
  },
  "spec": {
    "version": "v1",
    "type": "custom",
    "defaultAddons": {
      "enableIngress": true
    }
  }
}
`)

	cfg.mockClient.On("Apply", mock.Anything, mock.MatchedBy(func(blueprint *infrapb.Blueprint) bool {
		return blueprint.Metadata.Name == "test-blueprint-create" && blueprint.Metadata.Project == "test-project"
	}), mock.Anything).Return(nil)
	cfg.mockClient.On("Get", mock.Anything, mock.MatchedBy(func(opts options.GetOptions) bool {
		return opts.Name == "test-blueprint-create" && opts.Project == "test-project"
	})).Return(expectedBP, nil)
	cfg.mockClient.On("Delete", mock.Anything, mock.Anything).Return(nil)

	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: cfg.providerFactories,
		Steps: []resource.TestStep{
			{
				Config: helpers.LoadFixture(t, blueprintFixtures, "testdata/create.tf"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rafay_blueprint.tftest", "metadata.0.name", "test-blueprint-create"),
					resource.TestCheckResourceAttr("rafay_blueprint.tftest", "spec.0.version", "v1"),
				),
			},
		},
	})
}

func testResourceBlueprintReadHCL(t *testing.T, cfg blueprintTestConfig) {
	expectedBP := mustBlueprintFromJSON(t, `
{
  "metadata": {
    "name": "test-blueprint-read",
    "project": "test-project"
  },
  "spec": {
    "version": "v1"
  }
}
`)
	cfg.mockClient.On("Get", mock.Anything, mock.MatchedBy(func(opts options.GetOptions) bool {
		return opts.Name == "test-blueprint-read" && opts.Project == "test-project"
	})).Return(expectedBP, nil)
	cfg.mockClient.On("Delete", mock.Anything, mock.Anything).Return(nil)

	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: cfg.providerFactories,
		Steps: []resource.TestStep{
			{
				Config:        helpers.LoadFixture(t, blueprintFixtures, "testdata/read.tf"),
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

func testResourceBlueprintUpdateHCL(t *testing.T, cfg blueprintTestConfig) {
	expectedBPV1 := mustBlueprintFromJSON(t, `
{
  "metadata": {
    "name": "test-blueprint-update",
    "project": "test-project"
  },
  "spec": {
    "version": "v1",
    "type": "custom"
  }
}
`)
	expectedBPV2 := mustBlueprintFromJSON(t, `
{
  "metadata": {
    "name": "test-blueprint-update",
    "project": "test-project"
  },
  "spec": {
    "version": "v2",
    "type": "custom"
  }
}
`)

	cfg.mockClient.On("Apply", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	cfg.mockClient.On("Get", mock.Anything, mock.MatchedBy(func(opts options.GetOptions) bool {
		return opts.Name == "test-blueprint-update" && opts.Project == "test-project"
	})).Return(expectedBPV1, nil).Once()
	cfg.mockClient.On("Get", mock.Anything, mock.MatchedBy(func(opts options.GetOptions) bool {
		return opts.Name == "test-blueprint-update" && opts.Project == "test-project"
	})).Return(expectedBPV2, nil)
	cfg.mockClient.On("Delete", mock.Anything, mock.Anything).Return(nil)

	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: cfg.providerFactories,
		Steps: []resource.TestStep{
			{
				Config: helpers.LoadFixture(t, blueprintFixtures, "testdata/update_v1.tf"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rafay_blueprint.tftest", "spec.0.version", "v1"),
				),
			},
			{
				Config: helpers.LoadFixture(t, blueprintFixtures, "testdata/update_v2.tf"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rafay_blueprint.tftest", "spec.0.version", "v2"),
				),
			},
		},
	})
}

func testResourceBlueprintDeleteHCL(t *testing.T, cfg blueprintTestConfig) {
	expectedBP := mustBlueprintFromJSON(t, `
{
  "metadata": {
    "name": "test-blueprint-delete",
    "project": "test-project"
  },
  "spec": {
    "version": "v1",
    "type": "custom"
  }
}
`)

	cfg.mockClient.On("Apply", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	cfg.mockClient.On("Get", mock.Anything, mock.MatchedBy(func(opts options.GetOptions) bool {
		return opts.Name == "test-blueprint-delete" && opts.Project == "test-project"
	})).Return(expectedBP, nil)
	cfg.mockClient.On("Delete", mock.Anything, mock.MatchedBy(func(opts options.DeleteOptions) bool {
		return opts.Name == "test-blueprint-delete" && opts.Project == "test-project"
	})).Return(nil)

	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: cfg.providerFactories,
		Steps: []resource.TestStep{
			{
				Config: helpers.LoadFixture(t, blueprintFixtures, "testdata/delete.tf"),
			},
		},
	})
}

func testResourceBlueprintReadComplexHCL(t *testing.T, cfg blueprintTestConfig) {
	expectedBP := mustBlueprintFromJSON(t, `
{
  "metadata": {
    "name": "custom-blueprint",
    "project": "terraform"
  },
  "spec": {
    "version": "v0",
    "type": "custom",
    "base": {
      "name": "default",
      "version": "1.16.0"
    },
    "namespaceConfig": {
      "syncType": "managed",
      "enableSync": true
    },
    "defaultAddons": {
      "enableIngress": true,
      "enableCsiSecretStore": true,
      "enableMonitoring": true,
      "enableVM": false,
      "disableAwsNodeTerminationHandler": true,
      "csiSecretStoreConfig": {
        "enableSecretRotation": true,
        "syncSecrets": true,
        "rotationPollInterval": "2m",
        "providers": {
          "aws": true
        }
      },
      "monitoring": {
        "metricsServer": {
          "enabled": true,
          "discovery": {
            "namespace": "rafay-system"
          }
        },
        "helmExporter": {
          "enabled": true
        },
        "kubeStateMetrics": {
          "enabled": true
        },
        "nodeExporter": {
          "enabled": true
        },
        "prometheusAdapter": {
          "enabled": true
        },
        "resources": {
          "limits": {
            "memory": "300Mi",
            "cpu": "100m"
          }
        }
      }
    },
    "drift": {
      "action": "Deny",
      "enabled": true
    },
    "driftWebhook": {
      "enabled": true
    },
    "placement": {
      "autoPublish": false
    }
  }
}
`)

	cfg.mockClient.On("Get", mock.Anything, mock.MatchedBy(func(opts options.GetOptions) bool {
		return opts.Name == "custom-blueprint" && opts.Project == "terraform"
	})).Return(expectedBP, nil)
	cfg.mockClient.On("Apply", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	cfg.mockClient.On("Delete", mock.Anything, mock.Anything).Return(nil)

	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: cfg.providerFactories,
		Steps: []resource.TestStep{
			{
				Config:             complexBlueprintConfig(t, "200Mi"),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				Config:       complexBlueprintConfig(t, "300Mi"),
				ResourceName: "rafay_blueprint.blueprint",
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
			{
				Config:             complexBlueprintConfig(t, "300Mi"),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}
