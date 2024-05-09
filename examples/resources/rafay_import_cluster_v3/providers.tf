terraform {
  required_providers {
    rafay = {
      version = "= 1.1.28"
      source  = "rafay/rafay"
    }
  }
}

variable "rafay_config_file" {
  description = "rafay provider config file for authentication"
  sensitive   = true
  default     = "/Users/mastik5h/Desktop/rctl_config.json"
}

provider "rafay" {
  provider_config_file = var.rafay_config_file
}
