package rafay_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/RafaySystems/terraform-provider-rafay/rafay"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var (
	customBlueprintWithMostConfigSet = fmt.Sprintf(`
resource "rafay_blueprint" "blueprint" {
  metadata {
    name    = "custom-blueprint"
    project = "%s"
  }
  spec {
    version = "v0"
    base {
      name    = "default"
      version = "4.0.0"
    }
    namespace_config {
      sync_type   = "managed"
      enable_sync = true
    }
    default_addons {
      enable_ingress          = true
      enable_csi_secret_store = true
      enable_monitoring       = true
      enable_vm               = false
      disable_aws_node_termination_handler = true

      csi_secret_store_config {
        enable_secret_rotation = true
        sync_secrets           = true
        rotation_poll_interval = "2m"
        providers {
          aws = true
        }
      }
      monitoring {
        metrics_server {
          enabled = true
          discovery {
            namespace = "rafay-system"
          }
        }
        helm_exporter {
          enabled = true
        }
        kube_state_metrics {
          enabled = true
        }
        node_exporter {
          enabled = true
        }
        prometheus_adapter {
          enabled = true
        }
        resources {
          limits {
            memory = "200Mi"
            cpu    = "100m"
          }
        }
      }
    }
    drift {
      action  = "Deny"
      enabled = true
    }
    drift_webhook {
      enabled = true
    }
    placement {
      auto_publish = false
    }
  }
}
`, os.Getenv("RCTL_PROJECT"))
)

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

func TestResourceBlueprintAcceptance(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: blueprintProviderFactory(),
		Steps:             []resource.TestStep{{Config: customBlueprintWithMostConfigSet}},
	})
}
