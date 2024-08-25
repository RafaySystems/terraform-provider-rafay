package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestMksClusterResourceCreate(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testFwProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testMksClusterResourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rafay_mks_cluster.example", "metadata.0.name", "test-cluster"),
					resource.TestCheckResourceAttr("rafay_mks_cluster.example", "metadata.0.project", "test-project"),
					resource.TestCheckResourceAttr("rafay_mks_cluster.example", "proxy.0.enabled", "true"),
					resource.TestCheckResourceAttr("rafay_mks_cluster.example", "system_component_placement.0.region", "us-west-1"),
				),
			},
		},
	})
}

func TestMksClusterResourceRead(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testFwProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testMksClusterResourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rafay_mks_cluster.example", "metadata.0.name", "test-cluster"),
					resource.TestCheckResourceAttr("rafay_mks_cluster.example", "proxy.0.enabled", "true"),
				),
			},
		},
	})
}

func TestMksClusterResourceUpdate(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testFwProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testMksClusterResourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rafay_mks_cluster.example", "metadata.0.name", "test-cluster"),
				),
			},
			{
				Config: testMksClusterResourceConfigUpdated(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rafay_mks_cluster.example", "metadata.0.name", "updated-cluster"),
				),
			},
		},
	})
}

func TestMksClusterResourceImport(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testFwProviderFactories,
		Steps: []resource.TestStep{
			{
				ResourceName:      "rafay_mks_cluster.example",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestMksClusterResourceInvalidConfig(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testFwProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testMksClusterResourceConfigInvalid(),
				ExpectError: regexp.MustCompile("expected error message or pattern"),
			},
		},
	})
}

// Helper function to return the initial configuration
func testMksClusterResourceConfig() string {
	return `
resource "rafay_mks_cluster" "example" {
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

// Helper function to return the updated configuration
func testMksClusterResourceConfigUpdated() string {
	return `
resource "rafay_mks_cluster" "example" {
  metadata {
    name    = "updated-cluster"
    project = "test-project"
  }

  proxy {
    enabled = true
  }

  system_component_placement {
    region = "us-east-1"
  }
}
`
}

// Helper function to return an invalid configuration
func testMksClusterResourceConfigInvalid() string {
	return `
resource "rafay_mks_cluster" "example" {
  metadata {
    name    = ""
    project = "test-project"
  }

  proxy {
    enabled = true
  }
}
`
}
