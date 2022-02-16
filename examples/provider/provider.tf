terraform {
  required_providers {
    rafay = {
      version = ">= 0.1"
      source  = "RafaySystems/rafay"
    }
  }
}

provider "rafay" {
  provider_config_file = "/Users/cangadala/Downloads/rafay_qa-cangadala@rafay.co.json"
}
