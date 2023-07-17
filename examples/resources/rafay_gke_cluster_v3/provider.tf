terraform {
  required_providers {
    rafay = {
      "source": "rafay/rafay",
      "version": "1.1.13"
    }
  }
}

provider "rafay" {
  provider_config_file = var.rafay_config_file
}
