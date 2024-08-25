package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestMksClusterDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testFwProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testProviderConfig + testMksClusterDataSource(),
				Check:  resource.ComposeAggregateTestCheckFunc(
				// Verify number of coffees returned
				),
			},
		},
	})
}



// Helper function to return the initial configuration
func testMksClusterDataSource() string {
	return `
datasource "rafay_mks_cluster" "example" {
  metadata {
    name    = "test-cluster"
    project = "test-project"
  }

  proxy {
    enabled = true
  }

  system_component_placement {
    region = "us-west-1"
  }
}
`
}