variable "rafay_config_file" {
  description = "rafay provider config file for authentication"
  sensitive   = true
  default     = "/Users/phanindra/Downloads/phani-eaas.json"
}

# variable "name" {
#   description = "rafay provider environment template name"
#   default     = "my-environment-template"
# }

# variable "project" {
#   description = "project name where resource to be created"
#   default     = "terraform"
# }

# variable "r_version" {
#   description = "version of the resource"
#   default     = "v1"
# }

# variable "rt_name" {
#   description = "resource template name to be bundled within environment template"
#   default     = "my-resource-template"
# }

# variable "sr_name" {
#   description = "static resource name to be bundled within environment template"
#   default     = "my-static-resource"
# }

# variable "configcontext_name" {
#   description = "config context dependency"
#   default     = "my-config-context"
# }

# variable "agent_name" {
#   description = "agent responsible for processing"
#   default     = "my-agent"
# }