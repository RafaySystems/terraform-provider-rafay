variable "rafay_config_file" {
  description = "rafay provider config file for authentication"
  sensitive   = true
  default     = "/Users/user1/.rafay/cli/config.json"
}

variable "name" {
  description = "rafay provider environment name"
  sensitive   = false
  default     = "my-environment"
}

variable "project" {
  description = "project name where resource to be created"
  sensitive   = false
  default     = "terraform"
}

variable "et_name" {
  description = "environment template name based of which environment has to be created"
  sensitive   = false
  default     = "my-environment-template"
}

variable "et_version" {
  description = "environment template version based of which environment has to be created"
  sensitive   = false
  default     = "v1"
}