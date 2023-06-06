variable "rafay_config_file" {
  description = "rafay provider config file for authentication"
  sensitive   = true
  default     = "/Users/user1/.rafay/cli/config.json"
}

variable "name" {
  description = "rafay provider resource template name"
  sensitive = false
  default = "my-resource-template"
}

variable "project" {
  description = "project name where resource to be created"
  default = "terraform"
}

variable "r_version" {
  description = "version of the resource"
  default = "v1"
}

variable "repo_name" {
  description = "repository name of the resource config"
  default = "envmgr-repo"
}

variable "branch" {
  description = "branch of the repository"
  default = "tests"
}

variable "path" {
  description = "path of the repository"
  default = "test"
}

variable "configcontext_name" {
  description = "config context dependency"
  default = "my-config-context"
}