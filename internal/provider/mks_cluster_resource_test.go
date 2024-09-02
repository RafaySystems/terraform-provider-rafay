package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccMksClusterResourceCreate(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testFwProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testMksClusterResourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rafay_mks_cluster.example", "metadata.0.name", "example-cluster"),
					resource.TestCheckResourceAttr("rafay_mks_cluster.example", "metadata.0.project", "terraform"),
					resource.TestCheckResourceAttr("rafay_mks_cluster.example", "proxy.0.enabled", "true"),
				),
			},
		},
	})
}

func TestAccMksClusterResourceRead(t *testing.T) {
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

func TestAccMksClusterResourceUpdate(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testFwProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testMksClusterResourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rafay_mks_cluster.example", "metadata.0.name", "example-cluster"),
				),
			},
			{
				Config: testMksClusterResourceConfigUpdated(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rafay_mks_cluster.example", "metadata.0.name", "example-cluster"),
				),
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
  api_version = "infra.k8smgmt.io/v3"
  kind        = "Cluster"

  metadata = {
    annotations = {
      "key2" = "value2"
    }
    description  = "This is a sample MKS cluster."
    display_name = "example-cluster"
    labels = {
      "env" = "development"
    }
    name    = "example-cluster"
    project = "terraform"
  }

  spec = {
    blueprint = {
      name    = "minimal"
    }
    config = {
      auto_approve_nodes     = true
      dedicated_control_plane = false
      high_availability       = false
      kubernetes_version      = "v1.28.9"   
      kubernetes_upgrade = {
        strategy = "sequential"
        params = {
          worker_concurrency = "50%"
        }
      }

      location = "mumbai-in"
      network = {
        cni = {
          name    = "Calico"
          version = "3.26.1"
        }
        pod_subnet     = "10.244.0.0/16"
        service_subnet = "10.96.0.0/12"
      }

      nodes = [ 
        {
        arch            = "amd64"
        hostname        = "sample-node"
        operating_system = "Ubuntu22.04"
        private_ip      = "10.12.71.72"
        roles           = ["ControlPlane", "Worker"]

        labels =  {
          "app"   = "infra"
          "infra" = "true"
        }

        taints = [
          {
            effect = "NoSchedule"
            key    = "app"
            value  = "infra"
          },
          {
            effect = "NoSchedule"
            key = "infra"
            value = "true"
          }
        ]
       }
      ]
    }
    system_components_placement = {
      daemon_set_override = {
        daemon_set_tolerations = [ 
          {
            effect             = "NoSchedule"
            key                = "app"
            operator           = "Equal"
            value              = "infra"
          },
          {
            effect             = "NoSchedule"
            key                = "infra"
            operator           = "Equal"
            value              = "true"
          },
      ]
        node_selection_enabled = true
      }
      node_selector = {
        "app" = "infra"
        "infra" = "true"
      }
      tolerations = [
        {
          effect   = "NoSchedule"
          key      = "app"
          operator = "Equal"
          value    = "infra"
        },
        {
          effect   = "NoSchedule"
          key      = "infra"
          operator = "Equal"
          value    = "true"
        },
      ]
    }
    type = "mks"
  }
}
`
}

// Helper function to return the updated configuration
func testMksClusterResourceConfigUpdated() string {
	return `
	resource "rafay_mks_cluster" "example-cluster" {
	  api_version = "infra.k8smgmt.io/v3"
	  kind        = "Cluster"
	
	  metadata = {
		annotations = {
		  "key2" = "value2"
		}
		description  = "This is a sample MKS cluster."
		display_name = "example-cluster"
		labels = {
		  "env" = "development"
		}
		name    = "example-cluster"
		project = "terraform"
	  }
	
	  spec = {
		blueprint = {
		  name    = "minimal"
		}
		config = {
		  auto_approve_nodes     = true
		  dedicated_control_plane = false
		  high_availability       = false
		  kubernetes_version      = "v1.29.1"   
		  kubernetes_upgrade = {
			strategy = "sequential"
			params = {
			  worker_concurrency = "50%"
			}
		  }
	
		  location = "mumbai-in"
		  network = {
			cni = {
			  name    = "Calico"
			  version = "3.26.1"
			}
			pod_subnet     = "10.244.0.0/16"
			service_subnet = "10.96.0.0/12"
		  }
	
		  nodes = [ 
			{
			arch            = "amd64"
			hostname        = "sample-node"
			operating_system = "Ubuntu22.04"
			private_ip      = "10.12.71.72"
			roles           = ["ControlPlane", "Worker"]
	
			labels =  {
			  "app"   = "infra"
			  "infra" = "true"
			}
	
			taints = [
			  {
				effect = "NoSchedule"
				key    = "app"
				value  = "infra"
			  },
			  {
				effect = "NoSchedule"
				key = "infra"
				value = "true"
			  }
			]
		   }
		  ]
		}
		system_components_placement = {
		  daemon_set_override = {
			daemon_set_tolerations = [ 
			  {
				effect             = "NoSchedule"
				key                = "app"
				operator           = "Equal"
				value              = "infra"
			  },
			  {
				effect             = "NoSchedule"
				key                = "infra"
				operator           = "Equal"
				value              = "true"
			  },
		  ]
			node_selection_enabled = true
		  }
		  node_selector = {
			"app" = "infra"
			"infra" = "true"
		  }
		  tolerations = [
			{
			  effect   = "NoSchedule"
			  key      = "app"
			  operator = "Equal"
			  value    = "infra"
			},
			{
			  effect   = "NoSchedule"
			  key      = "infra"
			  operator = "Equal"
			  value    = "true"
			},
		  ]
		}
		type = "mks"
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
}
`
}
