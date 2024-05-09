resource "rafay_import_cluster_v3" "tf-v3-import1" {
  metadata {
    name    = "imp-v3-tf-1"
    project = "defaultproject"
  }
  spec {
    type = "imported"
    blueprint_config {
      name = "minimal"
      version = "latest"
    }
    # system_components_placement {
    #   node_selector = {
    #     app       = "infra"
    #     dedicated = "true"
    #   }
    #   tolerations {
    #     effect   = "PreferNoSchedule"
    #     key      = "app"
    #     operator = "Equal"
    #     value    = "infra"
    #   }
    #   daemon_set_override {
    #     node_selection_enabled = false
    #     tolerations {
    #       key      = "app1dedicated"
    #       value    = true
    #       effect   = "NoSchedule"
    #       operator = "Equal"
    #     }
    #   }
    # }
    config {
      provision_environment = "CLOUD"
      kubernetes_provider = "AKS"
      imported_cluster_location = "azure/centralindia"
    }
  }
  # bootstrap_path    = "imported-v3-tf-1-bootstrap.yaml"
  # values_path       = "<optional_path/values.yaml>"
}
meta {
    xyz = "xyz"
  }

# output "values_data" {
#   value = rafay_import_cluster.import_cluster.values_data
# }

# output "values_path" {
#   value = rafay_import_cluster.import_cluster.values_path
# }

# output "bootstrap_data" {
#   value = rafay_import_cluster.import_cluster.bootstrap_data
# }

# output "bootstrap_path" {
#   value = rafay_import_cluster.import_cluster.bootstrap_path
# }
