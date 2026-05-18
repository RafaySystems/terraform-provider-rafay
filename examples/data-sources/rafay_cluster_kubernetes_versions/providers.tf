# This example uses the RafaySystems provider. The data source rafay_cluster_kubernetes_versions
# is only available in the local provider build. From the repo root run: make install
# Then run terraform init and terraform plan in this directory.
terraform {
  required_providers {
    rafay = {
      version = "=1.1.52"
      source  = "registry.terraform.io/rafaysystems/rafay"
    }
  }
}

variable "rafay_config_file" {
  description = "rafay provider config file for authentication"
  default     = "/home/ujwal/.rafay/cli/config.json"
  sensitive   = true
}

# provider "rafay" {
#   provider_config_file = var.rafay_config_file
# }
