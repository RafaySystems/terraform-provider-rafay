variable "rafay_config_file" {
  description = "rafay provider config file for authentication"
  sensitive   = true
  default     = "/Users/user1/.rafay/cli/config.json"
}

variable "name" {
  description = "rafay provider resource template name"
  sensitive   = false
  default     = "test-rt-one"
}

variable "project" {
  description = "project name where resource to be created"
  default     = "terraform"
}

variable "r_version" {
  description = "version of the resource"
  default     = "v1"
}

variable "repo_name" {
  description = "repository name of the resource config"
  default     = "envmgr-demo"
}

variable "branch" {
  description = "branch of the repository"
  default     = "main"
}

variable "path" {
  description = "path of the repository"
  default     = "terraform/aws"
}

variable "configcontext_name" {
  description = "config context dependency"
  default     = "my-config-context"
}

variable "agent_name" {
  description = "agent to process resource template"
  default     = "newagentd"
}