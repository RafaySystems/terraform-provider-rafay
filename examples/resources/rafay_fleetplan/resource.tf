resource "rafay_fleetplan" "fp_clusters" {
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

resource "rafay_fleetplan" "fp_environments" {
  metadata {
    name    = "fleetplan2"
    project = "defaultproject"
  }
  spec {
    fleet {
      kind = "environments"

      projects {
        name = "defaultproject"
      }

      projects {
        name = "project1"
      }
      target_batch_size = 2
    }
    operation_workflow {
      operations {
        name = "op1"
        action {
          type        = "resourceDeploy"
          description = "deploy environment resources"
        }
      }
      operations {
        name = "op2"
        action {
          type        = "templateVersionUpdate"
          description = "update template version"
          environment_template_version_update_config {
            version = "v1.1"
          }
        }
      }
      operations {
        name = "op3"
        action {
          type        = "environmentVariableUpdate"
          description = "update cluster blueprint"
          environment_variable_update_config {
            key = "Blueprint Name"
            value = "minimal"
            value_type = "text"
          }
          continue_on_failure = true
        }
      }
      operations {
        name = "op4"
        action {
          type        = "resourceDestroy"
          description = "destroy environment resources"
        }
      }
    }
    schedules {
      name = "schedule1"
      description = "schedule1 description"
      type = "recurring"
      cadence {
        cron_expression = "0 18 * * *"
        cron_timezone = "Asia/Kolkata"
      }
      opt_out_options {
        allow_opt_out {
          value = true
        }
        max_allowed_duration = "20m"
        max_allowed_times = 5
      }
    }
  }
}

resource "rafay_fleetplan_trigger" "fp_trigger" {
  depends_on = [ rafay_fleetplan.fp_environments ]
  
  fleetplan_name = rafay_fleetplan.fp_environments.metadata[0].name
  project = rafay_fleetplan.fp_environments.metadata[0].project
  trigger_value = ""
}

data "rafay_fleetplan" "environment_fleetplan" {
  depends_on = [ rafay_fleetplan.fp_environments ]

  metadata {
    project = rafay_fleetplan.fp_environments.metadata[0].project
    name = rafay_fleetplan.fp_environments.metadata[0].name
  }
}

data "rafay_fleetplan_jobs" "fleetplan_jobs" {
  depends_on = [ rafay_fleetplan.fp_environments ]
  fleetplan_name = rafay_fleetplan.fp_environments.metadata[0].name
  project = rafay_fleetplan.fp_environments.metadata[0].project
}

data "rafay_fleetplan_job" "job1" {
  depends_on = [ rafay_fleetplan.fp_environments, data.rafay_fleetplan_jobs.fleetplan_jobs ]
  fleetplan_name = rafay_fleetplan.fp_environments.metadata[0].name
  project = rafay_fleetplan.fp_environments.metadata[0].project
  name = data.rafay_fleetplan_jobs.fleetplan_jobs.jobs[0].job_name
}

output "environment_fleetplan_spec" {
  value = data.rafay_fleetplan.environment_fleetplan.spec
}

output "environment_fleetplan_meta" {
  value = data.rafay_fleetplan.environment_fleetplan.metadata
}

output "fleetplan_jobs" {
  value = data.rafay_fleetplan_jobs.fleetplan_jobs.jobs
}

output "fleetplan_job_status" {
  value = data.rafay_fleetplan_job.job1.status
}


