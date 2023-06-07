resource "rafay_environment_template" "aws-et" {
  metadata {
    name    = var.name
    project = var.project
  }
  spec {
    version = var.r_version
    resources {
      type = "dynamic"
      kind = "resourcetemplate"
      name = var.rt_name
      resource_options {
        version = "v1"
        dedicated = true
      }
      depends_on {
        name = var.sr_name
      }
    }
    resources {
      type = "static"
      kind = "resource"
      name = var.sr_name
    }
    variables {
      name       = "name"
      value_type = "text"
      value      = "rds-envmgr"
      options {
        description = "this is a dummy variable"
        sensitive   = false
        required    = true
        override {
          type = "allowed"
        }
      }
    }
    hooks {
      on_init {
        name = "infracost"
        description = "this is an infracost hook"
        type = "http"
        agents {
          name = var.agent_name
        }
        on_failure = "continue"
        options {
          http {
            body = "initializing environment template"
            endpoint = "https://some-endpoint.com/post"
            headers = {
              TOKEN = "my-token"
              KEY   = "my-key"
            }
            method = "POST"
            success_condition = "200 OK"
          }
        }
        retry {
          enabled = true
          max_count = 2
        }
        timeout_seconds = 1000
      }
    }
    agents {
      name = var.agent_name
    }
    contexts {
      name = var.configcontext_name
    }
  }
}