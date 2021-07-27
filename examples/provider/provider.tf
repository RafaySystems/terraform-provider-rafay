terraform {
  required_providers {
    rafay = {
      version = ">= 0.1"
      source  = "RafaySystems/rafay"
    }
  }
}

provider "rafay" {
  provider_config_file = var.provider_config_file # Rafay provider config file (defaults to ~/.rafay/cli/config.json)
}
