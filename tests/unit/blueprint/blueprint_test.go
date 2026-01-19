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

//go:embed testdata/*
var blueprintFixtures embed.FS

func complexBlueprintConfig(t *testing.T, memory string) string {
	fixture := helpers.LoadFixture(t, blueprintFixtures, "testdata/complex_blueprint.tf")
	return fmt.Sprintf(fixture, memory)
}

func blueprintFixture(t *testing.T, fileName string) *infrapb.Blueprint {
	return mustBlueprintFromJSON(t, helpers.LoadFixture(t, blueprintFixtures, fileName))
}

func TestResourceBlueprint(t *testing.T) {
	tests := []struct {
		name string
		run  func(*testing.T, blueprintTestConfig)
	}{
		{"create", testResourceBlueprintCreate},
		{"update", testResourceBlueprintUpdate},
		{"delete", testResourceBlueprintDelete},
		{"read", testResourceBlueprintRead},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt := tt
			tt.run(t, newBlueprintTestConfig())
		})
	}
}

func testResourceBlueprintCreate(t *testing.T, cfg blueprintTestConfig) {
	expectedBP := blueprintFixture(t, "testdata/complex_blueprint.json")

	cfg.mockClient.On("Apply", mock.Anything, mock.MatchedBy(func(blueprint *infrapb.Blueprint) bool {
		return blueprint.Metadata.Name == "custom-blueprint" && blueprint.Metadata.Project == "terraform"
	}), mock.Anything).Return(nil)
	cfg.mockClient.On("Get", mock.Anything, mock.MatchedBy(func(opts options.GetOptions) bool {
		return opts.Name == "custom-blueprint" && opts.Project == "terraform"
	})).Return(expectedBP, nil)
	cfg.mockClient.On("Delete", mock.Anything, mock.Anything).Return(nil)

	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: cfg.providerFactories,
		Steps: []resource.TestStep{
			{
				Config: complexBlueprintConfig(t, "300Mi"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rafay_blueprint.blueprint", "metadata.0.name", "custom-blueprint"),
					resource.TestCheckResourceAttr("rafay_blueprint.blueprint", "spec.0.version", "v0"),
				),
			},
		},
	})
}

func testResourceBlueprintUpdate(t *testing.T, cfg blueprintTestConfig) {
	expectedBPV1 := blueprintFixture(t, "testdata/update_blueprint_v1.json")
	expectedBPV2 := blueprintFixture(t, "testdata/update_blueprint_v2.json")

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

func testResourceBlueprintDelete(t *testing.T, cfg blueprintTestConfig) {
	expectedBP := blueprintFixture(t, "testdata/delete_blueprint.json")

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

func testResourceBlueprintRead(t *testing.T, cfg blueprintTestConfig) {
	expectedBP := blueprintFixture(t, "testdata/complex_blueprint.json")

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
