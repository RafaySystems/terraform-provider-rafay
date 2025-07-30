terraform {
  required_providers {
    rafay = {
      version = "= 1.1.28"
      source  = "registry.terraform.io/rafaysystems/rafay"
    }
  }
}

provider "rafay" {
  provider_config_file = var.rafay_config_file
}


