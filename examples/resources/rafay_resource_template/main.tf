resource "rafay_resource_template" "testrt" {
  metadata {
    name    = "bptftestrt"
    project = "defaultproject"
  }
  spec {
    version  = "v1"
    provider = "system"
    provider_options {
      # terraform {
      #   version = "v1.5.7"
      #   backend_type = "system"
      #   lock {
      #     value = true
      #   }
      #   refresh {
      #     value = true
      #   }
      #   lock_timeout_seconds = 1
      # }
      system {
        kind = "blueprint"
        version = "v1"
      }
    }
    # repository_options {
    #   name           = "caas-eks"
    #   branch         = "main"
    #   directory_path = "new-gitops"
    # }
    # contexts {
    #   name = var.configcontext_name
    # }
    variables {
      name       = "blueprint_spec"
      value_type = "json"
      options {
        description = "this is the resource spec to be applied"
        sensitive   = false
        required    = true
      }
    }
 
    agents {
      name = "phani-tb"
    }
  }
}