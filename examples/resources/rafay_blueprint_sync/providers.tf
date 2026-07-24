terraform {
  required_providers {
    rafay = {
      version = ">= 0.1"
      source  = "registry.opentofu.org/rafaysystems/rafay"
    }
  }
}

provider "rafay" {
  provider_config_file = var.rafay_config_file
}
