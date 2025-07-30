variable "rafay_config_file" {
  description = "rafay provider config file for authentication"
  default     = "/Users/gopim/Downloads/gopi_org-gopikrishna@rafay.co5.json"
  sensitive   = true
}
variable "name" {
  description = "rafay provider config context name"
  sensitive   = false
  default     = "test-cc-one"
}

variable "project" {
  description = "project name where resource to be created"
  sensitive   = false
  default     = "defaultproject"
}
