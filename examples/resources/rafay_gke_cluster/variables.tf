variable "rafay_config_file" {
  description = "rafay provider config file for authentication"
  sensitive   = true
  default     = "/Users/user1/.rafay/cli/config.json"
}

variable "name" {
  description = "rafay provider config context name"
  sensitive   = false
  default     = "test-cc-one"
}

variable "project" {
  description = "project name where resource to be created"
  sensitive   = false
  default     = "terraform"
}
