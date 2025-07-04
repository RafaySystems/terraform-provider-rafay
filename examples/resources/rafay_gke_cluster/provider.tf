terraform {
  required_providers {
    rafay = {
      version = ">= 1.1.15"
      source  = "registry.terraform.io/RafaySystems/rafay"
    }
  }
}

provider "rafay" {
  provider_config_file = var.rafay_config_file
}
