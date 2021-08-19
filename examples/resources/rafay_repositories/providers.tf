terraform {
  required_providers {
    rafay = {
      version = ">= 0.1"
      source  = "registry.terraform.io/RafaySystems/rafay"
    }
  }
}

variable "rafay_config_file" {
  description = "rafay provider config file for authentication"
  default     = "<rafay-config-json-file>"
  sensitive   = true
}

provider "rafay" {
  provider_config_file = var.rafay_config_file
}
