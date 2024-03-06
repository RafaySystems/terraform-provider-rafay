#variable "provider_config_file" {}
variable "rafay_config_file" {
  description = "rafay provider config file for authentication"
  sensitive   = true
}

variable "git_token" {
  description = "git token for authentication"
  sensitive   = true
}

variable "git_user" {
  description = "git user for authentication"
  sensitive   = true
}

variable "delete_action_value" {
  description = "git user for authentication"
  default = "none"
}