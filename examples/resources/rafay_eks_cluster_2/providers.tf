terraform {
  required_providers {
    rafay = {
      version = "1.1.59"
      source  = "registry.terraform.io/rafaysystems/rafay"
    }
  }
}



variable "rafay_config_file" {
  description = "rafay provider config file for authentication"
  default     = "/Users/gopim/Downloads/gopi_org-gopikrishna@rafay.co-(6).json"
  sensitive   = true
}

provider "rafay" {
  provider_config_file = var.rafay_config_file
}

