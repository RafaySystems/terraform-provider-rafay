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
  default     = "/Users/gopim/Downloads/gopi_org-gopi@rafay.co.json"
  sensitive   = true
}

# variable "rafay_config_file" {
#   description = "rafay provider config file for authentication"
#   sensitive   = true
#   default     = "/Users/user1/.rafay/cli/config.json"
# }

provider "rafay" {
  provider_config_file = var.rafay_config_file
}
