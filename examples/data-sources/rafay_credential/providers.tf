terraform {
  required_providers {
    rafay = {
      version = "=1.1.28"
      source  = "registry.terraform.io/rafay/rafay"
    }
  }
}

variable "rafay_config_file" {
  description = "rafay provider config file for authentication"
  default     = "/home/ujwal/.rafay/cli/config.json"
  sensitive   = true
}

# provider "rafay" {
#   provider_config_file = var.rafay_config_file
# }