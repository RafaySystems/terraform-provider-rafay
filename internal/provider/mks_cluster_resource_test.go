package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

// Todo: Figure out way to automate bringing up oci instance for testing

func TestAccMksClusterResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testFwProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			// https://developer.hashicorp.com/terraform/plugin/framework/provider-servers#protocol-version
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		Steps: []resource.TestStep{
			{ // Create and Read testing
				Config: testMksClusterResourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rafay_mks_cluster.mks-example-cluster", "metadata.name", "mks-example-cluster"),
					resource.TestCheckResourceAttr("rafay_mks_cluster.mks-example-cluster", "metadata.project", "defaultproject"),
					resource.TestCheckResourceAttr("rafay_mks_cluster.mks-example-cluster", "spec.config.dedicated_control_plane", "true"),
					resource.TestCheckResourceAttr("rafay_mks_cluster.mks-example-cluster", "spec.config.high_availability", "false"),
				),
			},
			// Update and Read testing
			{
				Config: testMksClusterResourceConfigUpdated(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rafay_mks_cluster.mks-example-cluster", "metadata.name", "mks-example-cluster"),
					resource.TestCheckResourceAttr("rafay_mks_cluster.mks-example-cluster", "metadata.project", "defaultproject"),
					resource.TestCheckResourceAttr("rafay_mks_cluster.mks-example-cluster", "spec.config.dedicated_control_plane", "true"),
					resource.TestCheckResourceAttr("rafay_mks_cluster.mks-example-cluster", "spec.config.high_availability", "false"),
				),
			},

			// Delete testing automatically occurs in TestCase
		},
	})
}

// Helper function to return the initial configuration
// Cluster Configuration: No Ha Dedeicated Control Plane with one worker node
// Bring up OCI instances and provide the node details
func testMksClusterResourceConfig() string {
	return `
resource "rafay_mks_cluster" "mks-example-cluster" {
  api_version = "infra.k8smgmt.io/v3"
  kind        = "Cluster"

  metadata = {
    annotations = {
      "key2" = "value2"
    }
    description  = "This is a sample MKS cluster."
    display_name = "mks-example-cluster"
    labels = {
      "env" = "development"
    }
    name    = "mks-example-cluster"
    project = "defaultproject"
  }

  spec = {
    blueprint = {
      name    = "minimal"
    }
    cloud_credentials = "vasu-mks-ssh-010"
    config = {
      auto_approve_nodes      = true
      dedicated_control_plane = true
      high_availability       = false
      kubernetes_version      = "v1.28.9"   
      kubernetes_upgrade = {
        strategy = "sequential"
        params = {
          worker_concurrency = "50%"
        }
      }
      network = {
        cni = {
          name    = "Calico"
          version = "3.26.1"
        }
        pod_subnet     = "10.244.0.0/16"
        service_subnet = "10.96.0.0/12"
      }

      nodes = {
        "mks-sample-cp-node" = {
          arch            = "amd64"
          hostname        = "mks-sample-cp-node"
          operating_system = "Ubuntu22.04"
          private_ip      = "10.12.1.148"
          roles           = ["ControlPlane"]
        },
        "mks-sample-w-node" = {
          arch              = "amd64"
          hostname          = "mks-sample-w-node"
          operating_system  = "Ubuntu22.04"
          private_ip        = "10.12.50.133"
          roles             = ["Worker"]
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
              key    = "infra"
              value  = "true"
            }
          ]
        }
      }
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
// Update the kubernetes_version to v1.29.4
func testMksClusterResourceConfigUpdated() string {
	return `
resource "rafay_mks_cluster" "mks-example-cluster" {
  api_version = "infra.k8smgmt.io/v3"
  kind        = "Cluster"

  metadata = {
    annotations = {
      "key2" = "value2"
    }
    description  = "This is a sample MKS cluster."
    display_name = "mks-example-cluster"
    labels = {
      "env" = "development"
    }
    name    = "mks-example-cluster"
    project = "defaultproject"
  }

  spec = {
    blueprint = {
      name    = "minimal"
    }
    cloud_credentials = "vasu-mks-ssh-010"
    config = {
      auto_approve_nodes      = true
      dedicated_control_plane = true
      high_availability       = false
      kubernetes_version      = "v1.29.4"   
      kubernetes_upgrade = {
        strategy = "sequential"
        params = {
          worker_concurrency = "50%"
        }
      }
      network = {
        cni = {
          name    = "Calico"
          version = "3.26.1"
        }
        pod_subnet     = "10.244.0.0/16"
        service_subnet = "10.96.0.0/12"
      }
      
      nodes = {
        "mks-sample-cp-node" = {
          arch            = "amd64"
          hostname        = "mks-sample-cp-node"
          operating_system = "Ubuntu22.04"
          private_ip      = "10.12.1.148"
          roles           = ["ControlPlane"]
        },
        "mks-sample-w-node" = {
          arch              = "amd64"
          hostname          = "mks-sample-w-node"
          operating_system  = "Ubuntu22.04"
          private_ip        = "10.12.50.133"
          roles             = ["Worker"]
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
              key    = "infra"
              value  = "true"
            }
          ]
        }
      }
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
