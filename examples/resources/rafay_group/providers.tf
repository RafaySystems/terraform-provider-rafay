terraform {
  required_providers {
    rafay = {
      version = ">= 0.1"
      source  = "registry.terraform.io/rafay/rafay"
    }
  }
}

provider "rafay" {
  provider_config_file = var.provider_config_file
}
