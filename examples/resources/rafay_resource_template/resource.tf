resource "rafay_resource_template" "aws-elasticache" {
  metadata {
    name    = "aws-elasticache"
    project = "terraform"
  }
  spec {
    version  = "v1"
    provider = "terraform"
    provider_options {
      terraform {
        version                = "v1.4.4"
        use_system_state_store = true
      }
    }
    repository_options {
      name           = "envmgr-repo"
      branch         = "tests"
      directory_path = "test"
    }
    contexts {
      name = "aws-rds-config"
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
                    image = "eaasunittest/infracost:demo"
                    envvars = {
                      DOWNLOAD_TOKEN = "$(ctx.activities[\"aws-elasticache.artifact\"].output.files[\"job.tar.zst\"].token)"
                      DOWNLOAD_URL   = "$(ctx.activities[\"aws-elasticache.artifact\"].output.files[\"job.tar.zst\"].url)"
                    }
                    working_dir_path = "/tmp/workdir"
                  }
                }
                on_failure = "continue"
              }
            }
          }
        }
      }
    }
  }
}