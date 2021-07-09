terraform {
  required_providers {
    rafay = {
      version = ">= 0.1"
      source  = "registry.terraform.io/rafay/rafay"
    }
  }
}

variable "rafay_config_file" {
  description = "rafay provider config file for authentication"
  default = "/Users/stephanbenny/Downloads/rafay_qa-benny_at_rafay.co.json"
  sensitive   = true
}

provider "rafay" {
  provider_config_file = var.rafay_config_file
}
