package blueprint_test

import (
	"embed"
	"fmt"
	"testing"

	"github.com/RafaySystems/terraform-provider-rafay/rafay"
	"github.com/RafaySystems/terraform-provider-rafay/tests/helpers"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

//go:embed testdata/*.tf
var blueprintFixtures embed.FS

func blueprintProviderFactory() map[string]func() (*schema.Provider, error) {
	return map[string]func() (*schema.Provider, error){
		"rafay": func() (*schema.Provider, error) {
			provider := &schema.Provider{
				Schema: rafay.Schema(),
				ResourcesMap: map[string]*schema.Resource{
					"rafay_blueprint": rafay.ResourceBluePrint(),
				},
				DataSourcesMap: map[string]*schema.Resource{
					"rafay_blueprint": rafay.DataBluePrint(),
				},
				ConfigureContextFunc: rafay.ProviderConfigure,
			}
			return provider, nil
		},
	}
}

func TestResourceBlueprint(t *testing.T) {
	configurations := []string{
		fmt.Sprintf(helpers.LoadFixture(t, blueprintFixtures, "custom_blueprint_with_most_config.tf"), "test-blueprint-1", "defaultproject", "4.0.0"),
	}

	for _, configuration := range configurations {
		resource.ParallelTest(t, resource.TestCase{
			ProviderFactories: blueprintProviderFactory(),
			Steps:             []resource.TestStep{{Config: configuration}},
		})
	}
}
