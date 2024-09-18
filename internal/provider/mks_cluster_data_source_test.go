package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccMksClusterDataSource tests the data source for the Rafay MKS cluster

func TestAccMksClusterDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testFwProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testProviderConfig + testMksClusterDataSource(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.rafay_mks_cluster.mks-example-cluster", "metadata.name", "mks-example-cluster"),
					resource.TestCheckResourceAttr("data.rafay_mks_cluster.mks-example-cluster", "metadata.project", "defaultproject"),
				),
			},
		},
	})
}

// Helper function to return the initial configuration
func testMksClusterDataSource() string {
	return `
data "rafay_mks_cluster" "mks-example-cluster" {
  metadata = {
    name    = "mks-example-cluster"
    project = "defaultproject"
  }
}
`
}
