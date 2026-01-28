package blueprint_test

import (
	"embed"
	"fmt"
	"os"
	"testing"

	"github.com/RafaySystems/terraform-provider-rafay/rafay"
	"github.com/RafaySystems/terraform-provider-rafay/tests/helpers"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

//go:embed testdata/*.tf
var fixtures embed.FS

func blueprintProviderFactory() map[string]func() (*schema.Provider, error) {
	return map[string]func() (*schema.Provider, error){
		"rafay": func() (*schema.Provider, error) {
			provider := &schema.Provider{
				Schema: rafay.Schema(),
				ResourcesMap: map[string]*schema.Resource{
					"rafay_blueprint": rafay.ResourceBluePrint(),
					"rafay_addon":     rafay.ResourceAddon(),
				},
				DataSourcesMap: map[string]*schema.Resource{
					"rafay_blueprint": rafay.DataBluePrint(),
					"rafay_addon":     rafay.DataAddon(),
				},
				ConfigureContextFunc: rafay.ProviderConfigure,
			}
			return provider, nil
		},
	}
}

func TestResourceBlueprint(t *testing.T) {
	// to test more scenarios, add more configurations to this slice
	configurations := []struct {
		name   string
		config string
	}{
		{
			name:   "complex-blueprint-1",
			config: fmt.Sprintf(helpers.LoadFixture(t, fixtures, "testdata/custom_blueprint_with_most_config.tf"), "complex-blueprint-1", os.Getenv("RCTL_PROJECT"), os.Getenv("BASE_BLUEPRINT_VERSION")),
		},
		{
			name:   "blueprint-with-addons",
			config: fmt.Sprintf(helpers.LoadFixture(t, fixtures, "testdata/blueprint_with_addons.tf"), os.Getenv("RCTL_PROJECT"), os.Getenv("RCTL_PROJECT"), os.Getenv("BASE_BLUEPRINT_VERSION")),
		},
	}

	for _, tc := range configurations {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			resource.ParallelTest(t, resource.TestCase{
				ProviderFactories: blueprintProviderFactory(),
				Steps:             []resource.TestStep{{Config: tc.config}},
			})
		})
	}
}
