resource "rafay_pipeline" "tfdemopipeline1" {
  metadata {
    name    = "email-test"
    project = "terraform"
  }
  spec {
    active = false
    sharing {
      enabled = false
    }
    stages {
      config {
        approvers {
          sso_user  = false
          user_name = "hardik@rafay.co"
        }
        timeout = "2m0s"
        type    = "Email"
      }
      name = "email"
      type = "Approval"
    }
  }
}

resource "rafay_pipeline" "tfdemopipeline" {
  metadata {
    name = "test"
    project = "terraform"
    annotations = {}
    labels      = {}
  }
  spec {
    stages {
        name =  "s1"
        type = "SystemSync"
        next {
            name = "s2"
        }
        config  {
            git_to_system_sync = true
            included_resources {
                name =  "Workload"
            }
            excluded_resources {
                name =  "OPAConstraint"
            }
             source_repo {
                 repository = "test1"
                 revision =  "main"
                 path {
                    name = "project"
                 }
             }
            #destination_repo {}
            source_as_destination = true
            action {
                destroy       = false
                refresh       = false
                secret_groups = []
            }
            
        }
        variables {
            name = "x"
            type = "String"
            value = "trigger.name"
        }
    }
    stages{
        name = "s2"
        type = "Approval"
        next {
            name = "s3"
        }
        config {
          type = "Email"
          approvers {
              user_name = "user@company.co"
          }  
          timeout = "10s"
          action {
                destroy       = false
                refresh       = false
                secret_groups = []
          }
        }
    }
    stages {
        name = "s3"
        type = "DeployWorkload"
        next {
            name = "s4"
        }
        config {
            git_to_system_sync                      = false
            persist_working_directory               = false
            source_as_destination                   = false
            system_to_git_sync                      = false
            use_revision_from_webhook_trigger_event = false
            workload                                = "w2"

            action {
                destroy       = false
                refresh       = false
                secret_groups = []
            }
        }
    }
    stages {
        name = "s4"
        type = "InfraProvisioner"
        next {
          name = "s4"
        }
        config {
          type =  "Terraform"
          provisioner =  "i1"
          revision =  "main"
          agents {
              name = "agent1"
          }
          action {
            action = "Apply"
            refresh = true
            secret_groups = []
          }
        }
    }
    stages {
        name = "s5"
        type = "DeployWorkloadTemplate"
        config {
          workload_template =  "fayas-qctemp"
          namespace =  "main"
          placement {
            selector = "rafay.dev/clusterName=shishir-gitops"
          }
          use_revision_from_webhook_trigger_event = false

          overrides {
            type = "HelmValues"
            template {
                repository = "test1"
                revision = "main"
                paths {
                    name = "project1"
                }
            }
            weight = 4
          }

          overrides {
            type = "HelmValues"
            template {
                inline = "debug: {{ .stages.stage2.status}}"
            }
            weight = 2
          }

          action {
            destroy = false
            refresh = false
            secret_groups = []
          }
        }
    }
    triggers {
        type =  "Webhook"
        name = "t1"
        config {
            repo {
                provider = "Github"
                repository = "test1"
                revision =  "main"
                paths {
                    name = "project"
                }
            }
        }
        variables {
            name = "x"
            type = "String"
            value = "trigger.name"
        }
    }
    triggers {
        type =  "Webhook"
        name = "t2"
        config {
            repo {
                provider = "AzureRepos"
                repository = "test1"
                revision =  "main"
                paths {
                    name = "project"
                }
            }
        }
        variables {
            name = "x"
            type = "String"
            value = "trigger.name"
        }
    }
    triggers {
        type =  "Cron"
        name = "t3"
        config {
            cron_expression = "0 0 * * *"
            repo {
                provider = "AzureRepos"
                repository = "test1"
                revision =  "main"
                paths {
                    name = "project"
                }
            }
        }
        variables {
            name = "x"
            type = "String"
            value = "trigger.name"
        }
    }
    sharing  {
      enabled = false
    }
    active = false
  }
}
