resource "rafay_fleetplan" "fp1" {
  metadata {
    name    = "fleetplan1"
    project = "defaultproject"
  }
  spec {
    fleet {
      kind = "clusters"
      labels = {
        role = "qa"
      }

      projects {
        name = "defaultproject"
      }
    }

    operation_workflow {
      operations {
        name = "op1"
        action {
          type        = "controlPlaneUpgrade"
          description = "upgrading control plane"
          control_plane_upgrade_config {
            version = "1.25"
          }
          name = "action1"
        }
        prehooks {
          name        = "prehooks1"
          description = "list all pods 10"
          inject      = ["KUBECONFIG"]
          container_config {
            runner {
              type = "cluster"
            }
            image     = "bitnami/kubectl"
            arguments = ["get", "po", "-A"]
          }
          timeout_seconds = 100
        }
        prehooks {
          name        = "prehooks2"
          description = "list all pods 2"
          inject      = ["KUBECONFIG"]
          container_config {
            runner {
              type = "cluster"
            }
            image            = "bitnami/kubectl"
            arguments        = ["get", "po", "-A"]
            cpu_limit_milli  = "1000"
            memory_limit_mb  = "100"
            working_dir_path = "/var/"
          }
        }
      }
      operations {
        name = "op2"
        action {
          type        = "patch"
          description = "upgrading control plane and nodegroup"
          patch_config {
            op    = "replace"
            path  = ".spec.config.managedNodeGroups[0].maxSize"
            value = jsonencode(18)
          }
          patch_config {
            op    = "replace"
            path  = ".spec.blueprintConfig.name"
            value = jsonencode("minimal")
          }
          continue_on_failure = true
          name                = "action2"
        }
        posthooks {
          name        = "posthooks1"
          description = "list all pods 1"
          inject      = ["KUBECONFIG"]
          container_config {
            runner {
              type = "cluster"
            }
            image     = "bitnami/kubectl"
            arguments = ["get", "po", "-A"]
          }
          success_condition = "if #status.container.exitCode == 0 { success: true }"
        }
      }
      operations {
        name = "op3"
        action {
          name        = "action3"
          type        = "nodeGroupsUpgrade"
          description = "upgrading nodegroup"
          node_groups_upgrade_config {
            version = "1.25"
            names   = ["ng1", "ng2"]
          }
        }
      }
      operations {
        name = "op4"
        action {
          name        = "action4"
          type        = "blueprintUpdate"
          description = "updating blueprint with named action"
          blueprint_update_config {
            name    = "default"
            version = "latest"
          }
        }
      }
    }

    agents {
      name = "demoagent"
    }
  }
}

