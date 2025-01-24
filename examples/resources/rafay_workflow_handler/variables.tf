variable "rafay_config_file" {
  description = "rafay provider config file for authentication"
  sensitive   = true
  default     = "/Users/chaitanyaangadala/.rafay/cli/config.json"
}

variable "name" {
  description = "rafay provider workflow_handler name"
  sensitive   = false
  default     = "test-terraform-workflow-handler"
}

variable "project" {
  description = "project name where resource to be created"
  sensitive   = false
  default     = "defaultproject"
}

variable "type" {
  description = "workflow_handler type to be created"
  sensitive   = false
  default     = "container"
}

variable "image" {
  description = "image if the workflow_handler type is container"
  sensitive   = false
  default     = "dockerhub.io/envmgr"
}