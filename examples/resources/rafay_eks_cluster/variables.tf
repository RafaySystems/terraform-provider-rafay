#variable "provider_config_file" {}

variable "name" {
  description = "rafay provider static resource name"
  sensitive   = false
  default     = "my-static-resource"
}

variable "project" {
  description = "rafay provider static resource name"
  sensitive   = false
  default     = "akshay"
}

variable "cloudprovider" {
  description = "rafay provider static resource name"
  sensitive   = false
  default     = "qa-automation"
}