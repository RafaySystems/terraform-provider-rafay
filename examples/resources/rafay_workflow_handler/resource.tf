resource "rafay_workflow_handler" "workflow_handler" {
  metadata {
    name    = var.name
    project = var.project
  }
  spec {
    config {
      type            = var.type
      timeout_seconds = 100
      max_retry_count = 3
      container {
        image     = var.image
        arguments = ["--log-level=3"]
        commands  = ["run main.go"]
        image_pull_credentials {
          password = "gibberesh"
          registry = "hub.docker.io"
          username = "randomuser"
        }
        kube_config_options {
          kube_config    = "path/to/kubeconfig.json"
          out_of_cluster = true
        }
        kube_options {
          labels = {
            "name" : "terraform"
          }
          resources = ["pods", "deployments"]
          namespace = "rafay-core"
          affinity {
            node_affinity {
              required_during_scheduling_ignored_during_execution {
              node_selector_terms {
                match_expressions {
                key      = "kubernetes.io/e2e-az-name"
                operator = "In"
                values   = ["e2e-az1", "e2e-az2"]
                }
              }
              }
              preferred_during_scheduling_ignored_during_execution {
                weight = 1
                preference {
                  match_expressions {
                  key      = "another-node-label-key"
                  operator = "In"
                  values   = ["another-node-label-value"]
                  }
                }
              }
            }
            pod_affinity {
              required_during_scheduling_ignored_during_execution {
              label_selector {
                match_expressions {
                key      = "security"
                operator = "In"
                values   = ["S1"]
                }
              }
              topology_key = "kubernetes.io/hostname"
              }
              preferred_during_scheduling_ignored_during_execution {
              weight = 1
              pod_affinity_term {
                label_selector {
                match_expressions {
                  key      = "security"
                  operator = "In"
                  values   = ["S2"]
                }
                }
                topology_key = "kubernetes.io/hostname"
              }
              }
            }
            pod_anti_affinity {
              required_during_scheduling_ignored_during_execution {
              label_selector {
                match_expressions {
                key      = "security"
                operator = "In"
                values   = ["S1"]
                }
              }
              topology_key = "kubernetes.io/hostname"
              }
              preferred_during_scheduling_ignored_during_execution {
              weight = 1
              pod_affinity_term {
                label_selector {
                match_expressions {
                  key      = "security"
                  operator = "In"
                  values   = ["S1"]
                }
                }
                topology_key = "kubernetes.io/hostname"
              }
              }
            }
          }
        }
        volumes {
          use_pvc {
            value = true
          }
          mount_path = "/tmp/var"
          pvc_size_gb = "20"
          pvc_storage_class = "hdb"
        }
      }
    }
  }
}