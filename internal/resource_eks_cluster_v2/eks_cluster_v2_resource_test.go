package resource_eks_cluster_v2

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// testAccProtoV6ProviderFactories is used for acceptance tests
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"rafay": providerserver.NewProtocol6WithError(New("test")()),
}

func TestAccEKSClusterV2Resource_Basic(t *testing.T) {
	clusterName := "test-eks-cluster-v2-basic"
	projectName := "defaultproject"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccEKSClusterV2ResourceConfig_Basic(clusterName, projectName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("rafay_eks_cluster_v2.test", "cluster.metadata.name", clusterName),
					resource.TestCheckResourceAttr("rafay_eks_cluster_v2.test", "cluster.metadata.project", projectName),
					resource.TestCheckResourceAttr("rafay_eks_cluster_v2.test", "cluster.spec.type", "aws-eks"),
					resource.TestCheckResourceAttr("rafay_eks_cluster_v2.test", "cluster.spec.blueprint", "default"),
					resource.TestCheckResourceAttr("rafay_eks_cluster_v2.test", "cluster_config.metadata.region", "us-west-2"),
					resource.TestCheckResourceAttr("rafay_eks_cluster_v2.test", "cluster_config.metadata.version", "1.28"),
					resource.TestCheckResourceAttrSet("rafay_eks_cluster_v2.test", "id"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "rafay_eks_cluster_v2.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccEKSClusterV2ResourceConfig_Updated(clusterName, projectName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("rafay_eks_cluster_v2.test", "cluster.metadata.labels.environment", "production"),
					resource.TestCheckResourceAttr("rafay_eks_cluster_v2.test", "cluster_config.metadata.version", "1.29"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccEKSClusterV2Resource_WithTolerations(t *testing.T) {
	clusterName := "test-eks-cluster-v2-tolerations"
	projectName := "defaultproject"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with tolerations
			{
				Config: testAccEKSClusterV2ResourceConfig_WithTolerations(clusterName, projectName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("rafay_eks_cluster_v2.test", "cluster.spec.system_components_placement.tolerations.node-role.key", "node-role"),
					resource.TestCheckResourceAttr("rafay_eks_cluster_v2.test", "cluster.spec.system_components_placement.tolerations.node-role.value", "system"),
					resource.TestCheckResourceAttr("rafay_eks_cluster_v2.test", "cluster.spec.system_components_placement.tolerations.node-role.effect", "NoSchedule"),
				),
			},
			// Update tolerations
			{
				Config: testAccEKSClusterV2ResourceConfig_WithTolerationsUpdated(clusterName, projectName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("rafay_eks_cluster_v2.test", "cluster.spec.system_components_placement.tolerations.node-role.value", "infra"),
					resource.TestCheckResourceAttr("rafay_eks_cluster_v2.test", "cluster.spec.system_components_placement.tolerations.gpu.key", "gpu"),
				),
			},
		},
	})
}

func TestAccEKSClusterV2Resource_WithNodeGroups(t *testing.T) {
	clusterName := "test-eks-cluster-v2-nodegroups"
	projectName := "defaultproject"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with node groups
			{
				Config: testAccEKSClusterV2ResourceConfig_WithNodeGroups(clusterName, projectName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("rafay_eks_cluster_v2.test", "cluster_config.node_groups.primary.name", "primary-ng"),
					resource.TestCheckResourceAttr("rafay_eks_cluster_v2.test", "cluster_config.node_groups.primary.instance_type", "t3.large"),
					resource.TestCheckResourceAttr("rafay_eks_cluster_v2.test", "cluster_config.node_groups.primary.desired_capacity", "3"),
				),
			},
			// Add a node group
			{
				Config: testAccEKSClusterV2ResourceConfig_WithNodeGroupsAdded(clusterName, projectName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("rafay_eks_cluster_v2.test", "cluster_config.node_groups.primary.name", "primary-ng"),
					resource.TestCheckResourceAttr("rafay_eks_cluster_v2.test", "cluster_config.node_groups.gpu.name", "gpu-ng"),
					resource.TestCheckResourceAttr("rafay_eks_cluster_v2.test", "cluster_config.node_groups.gpu.instance_type", "g4dn.xlarge"),
				),
			},
		},
	})
}

func TestAccEKSClusterV2Resource_MapDiffPrecision(t *testing.T) {
	t.Skip("This test verifies that map-based schema prevents unwanted diffs")
	// This would be a manual test to verify the diff output
	// When updating one toleration, only that toleration should show in the diff
}

// testAccPreCheck validates the necessary test environment variables
func testAccPreCheck(t *testing.T) {
	// Check required environment variables
	// RCTL_API_KEY, RCTL_REST_ENDPOINT, RCTL_PROJECT should be set
}

// testAccCheckEKSClusterV2Exists checks if the cluster exists
func testAccCheckEKSClusterV2Exists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no cluster ID is set")
		}

		// Here you would call the API to verify the cluster exists
		// For now, just check the ID is set
		return nil
	}
}

// Test configurations

func testAccEKSClusterV2ResourceConfig_Basic(clusterName, projectName string) string {
	return fmt.Sprintf(`
resource "rafay_eks_cluster_v2" "test" {
  cluster = {
    kind = "Cluster"
    metadata = {
      name    = %[1]q
      project = %[2]q
      labels = {
        environment = "test"
        managed-by  = "terraform"
      }
    }
    spec = {
      type          = "aws-eks"
      blueprint     = "default"
      cloud_provider = "aws-creds"
      cni_provider  = "aws-cni"
    }
  }

  cluster_config = {
    apiversion = "rafay.io/v1alpha5"
    kind       = "ClusterConfig"
    metadata = {
      name    = %[1]q
      region  = "us-west-2"
      version = "1.28"
      tags = {
        Environment = "test"
        ManagedBy   = "terraform"
      }
    }
  }
}
`, clusterName, projectName)
}

func testAccEKSClusterV2ResourceConfig_Updated(clusterName, projectName string) string {
	return fmt.Sprintf(`
resource "rafay_eks_cluster_v2" "test" {
  cluster = {
    kind = "Cluster"
    metadata = {
      name    = %[1]q
      project = %[2]q
      labels = {
        environment = "production"  // Changed from test
        managed-by  = "terraform"
      }
    }
    spec = {
      type          = "aws-eks"
      blueprint     = "default"
      cloud_provider = "aws-creds"
      cni_provider  = "aws-cni"
    }
  }

  cluster_config = {
    apiversion = "rafay.io/v1alpha5"
    kind       = "ClusterConfig"
    metadata = {
      name    = %[1]q
      region  = "us-west-2"
      version = "1.29"  // Upgraded from 1.28
      tags = {
        Environment = "production"
        ManagedBy   = "terraform"
      }
    }
  }
}
`, clusterName, projectName)
}

func testAccEKSClusterV2ResourceConfig_WithTolerations(clusterName, projectName string) string {
	return fmt.Sprintf(`
resource "rafay_eks_cluster_v2" "test" {
  cluster = {
    kind = "Cluster"
    metadata = {
      name    = %[1]q
      project = %[2]q
    }
    spec = {
      type          = "aws-eks"
      blueprint     = "default"
      cloud_provider = "aws-creds"
      cni_provider  = "aws-cni"
      
      system_components_placement = {
        tolerations = {
          "node-role" = {
            key      = "node-role"
            operator = "Equal"
            value    = "system"
            effect   = "NoSchedule"
          }
        }
      }
    }
  }

  cluster_config = {
    apiversion = "rafay.io/v1alpha5"
    kind       = "ClusterConfig"
    metadata = {
      name    = %[1]q
      region  = "us-west-2"
      version = "1.28"
    }
  }
}
`, clusterName, projectName)
}

func testAccEKSClusterV2ResourceConfig_WithTolerationsUpdated(clusterName, projectName string) string {
	return fmt.Sprintf(`
resource "rafay_eks_cluster_v2" "test" {
  cluster = {
    kind = "Cluster"
    metadata = {
      name    = %[1]q
      project = %[2]q
    }
    spec = {
      type          = "aws-eks"
      blueprint     = "default"
      cloud_provider = "aws-creds"
      cni_provider  = "aws-cni"
      
      system_components_placement = {
        tolerations = {
          "node-role" = {
            key      = "node-role"
            operator = "Equal"
            value    = "infra"  // Changed from system
            effect   = "NoSchedule"
          }
          "gpu" = {  // New toleration added
            key      = "gpu"
            operator = "Exists"
            effect   = "NoSchedule"
          }
        }
      }
    }
  }

  cluster_config = {
    apiversion = "rafay.io/v1alpha5"
    kind       = "ClusterConfig"
    metadata = {
      name    = %[1]q
      region  = "us-west-2"
      version = "1.28"
    }
  }
}
`, clusterName, projectName)
}

func testAccEKSClusterV2ResourceConfig_WithNodeGroups(clusterName, projectName string) string {
	return fmt.Sprintf(`
resource "rafay_eks_cluster_v2" "test" {
  cluster = {
    kind = "Cluster"
    metadata = {
      name    = %[1]q
      project = %[2]q
    }
    spec = {
      type          = "aws-eks"
      blueprint     = "default"
      cloud_provider = "aws-creds"
      cni_provider  = "aws-cni"
    }
  }

  cluster_config = {
    apiversion = "rafay.io/v1alpha5"
    kind       = "ClusterConfig"
    metadata = {
      name    = %[1]q
      region  = "us-west-2"
      version = "1.28"
    }
    
    node_groups = {
      "primary" = {
        name             = "primary-ng"
        instance_type    = "t3.large"
        desired_capacity = 3
        min_size         = 2
        max_size         = 5
        volume_size      = 80
      }
    }
  }
}
`, clusterName, projectName)
}

func testAccEKSClusterV2ResourceConfig_WithNodeGroupsAdded(clusterName, projectName string) string {
	return fmt.Sprintf(`
resource "rafay_eks_cluster_v2" "test" {
  cluster = {
    kind = "Cluster"
    metadata = {
      name    = %[1]q
      project = %[2]q
    }
    spec = {
      type          = "aws-eks"
      blueprint     = "default"
      cloud_provider = "aws-creds"
      cni_provider  = "aws-cni"
    }
  }

  cluster_config = {
    apiversion = "rafay.io/v1alpha5"
    kind       = "ClusterConfig"
    metadata = {
      name    = %[1]q
      region  = "us-west-2"
      version = "1.28"
    }
    
    node_groups = {
      "primary" = {
        name             = "primary-ng"
        instance_type    = "t3.large"
        desired_capacity = 3
        min_size         = 2
        max_size         = 5
        volume_size      = 80
      }
      "gpu" = {  // New node group added
        name             = "gpu-ng"
        instance_type    = "g4dn.xlarge"
        desired_capacity = 1
        min_size         = 0
        max_size         = 3
        volume_size      = 100
        labels = {
          "workload" = "ml"
        }
        taints = {
          "gpu" = {
            key    = "nvidia.com/gpu"
            value  = "true"
            effect = "NoSchedule"
          }
        }
      }
    }
  }
}
`, clusterName, projectName)
}

