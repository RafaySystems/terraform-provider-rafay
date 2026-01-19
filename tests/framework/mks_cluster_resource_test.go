//go:build planonly
// +build planonly

package framework

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

// Minimal env so provider Configure() is happy in plan-only runs.
func setDummyEnv(t *testing.T) {
	_ = os.Setenv("RCTL_API_KEY", "dummy")
	_ = os.Setenv("RCTL_REST_ENDPOINT", "console.example.dev")
}

func TestAccMksClusterResource(t *testing.T) {
	setDummyEnv(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testFwProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			// https://developer.hashicorp.com/terraform/plugin/framework/provider-servers#protocol-version
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		Steps: []resource.TestStep{
			{ // Plan-only testing
				Config:             testMksClusterResourceConfig(),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// Helper function to return the initial configuration
// Cluster Configuration: No HA Dedicated Control Plane with one worker node
// Bring up OCI instances and provide the node details
func testMksClusterResourceConfig() string {
	return `
provider "rafay" {
  api_key       = "dummy"
  rest_endpoint = "console.example.dev"
  project       = "defaultproject"
  ignore_insecure_tls_error = true
}

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
    cloud_credentials = "mks-ssh-creds"
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
provider "rafay" {
  api_key       = "dummy"
  rest_endpoint = "console.example.dev"
  project       = "defaultproject"
  ignore_insecure_tls_error = true
}

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
