variable "num_environments" {
  description = "Number of environments to create"
  type        = number
  default     = 1
}

variable "et_name" {
  description = "Environment template name"
  type        = string
}

variable "et_version" {
  description = "Environment template version"
  type        = string
}

variable "agent" {
  description = "Agent name"
  type        = string
}

variable "project" {
  description = "Project name"
  type        = string
}

variable "name_prefix" {
  description = "Prefix for environment names"
  type        = string
  default     = "env"
}

variable "rafay_config_file" {
  description = "rafay provider config file for authentication"
  sensitive   = true
  default     = "/Users/user1/.rafay/cli/config.json"
}