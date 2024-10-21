resource "rafay_resource_template" "aws-elasticache" {
  metadata {
    name    = var.name
    project = var.project
  }
  spec {
    version  = var.r_version
    provider = "terraform"
    provider_options {
      terraform {
        version = "v1.4.4"
        backend_type = "custom"
        backend_configs = ["path"]
        var_files       = ["path"]
        plugin_dirs     = ["path"]
        lock {
          value = true
        }
        refresh {
          value = true
        }
        lock_timeout_seconds = 1
      }
    }
    repository_options {
      name           = var.repo_name
      branch         = var.branch
      directory_path = var.path
    }
    contexts {
      name = var.configcontext_name
    }
    variables {
      name       = "name"
      value_type = "text"
      value      = "aws-elasticache"
      options {
        description = "this is the resource name to be applied"
        sensitive   = false
        required    = true
      }
    }
    hooks {
      provider {
        terraform {
          deploy {
            init {
              before {
                name = "infracost"
                type = "container"
                options {
                  container {
                    image     = "eaasunittest/infracost:demo"
                    arguments = ["--verbose"]
                    commands  = ["scan"]
                    envvars = {
                      DOWNLOAD_TOKEN = "$(ctx.activities[\"aws-elasticache.artifact\"].output.files[\"job.tar.zst\"].token)"
                      DOWNLOAD_URL   = "$(ctx.activities[\"aws-elasticache.artifact\"].output.files[\"job.tar.zst\"].url)"
                    }
                    working_dir_path = "/workdir"
                  }
                }
                on_failure = "continue"
                execute_once = true
              }
              after {
                name = "internal-approval"
                type = "approval"
                options {
                  approval {
                    type = "internal"
                  }
                }
              }
            }
            output {
              before {
                name = "webhook"
                type = "http"
                options {
                  http {
                    endpoint = "https://jsonplaceholder.typicode.com/todos/1"
                    method   = "POST"
                    headers = {
                      X-TOKEN = "token"
                    }
                    body              = "request-body"
                    success_condition = "200OK"
                  }
                }
              }
            }
          }
        }
      }
    }
    agents {
      name = var.agent_name
    }
  }
}