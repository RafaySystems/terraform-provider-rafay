variable "rafay_config_file" {
  description = "rafay provider config file for authentication"
  sensitive   = true
  default     = "/Users/user1/.rafay/cli/config.json"
}

variable "name" {
  description = "rafay provider static resource name"
  sensitive   = false
  default     = "my-static-resource"
}

variable "project" {
  description = "project name where resource to be created"
  default     = "terraform"
}