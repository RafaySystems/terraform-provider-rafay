terraform {
  required_providers {
    rafay = {
      source = "registry.terraform.io/Rafay/rafay"
    }
  }
}

provider "rafay" {
  provider_config_file = var.rafay_config_file
}
