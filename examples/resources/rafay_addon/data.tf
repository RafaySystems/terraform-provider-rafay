data "rafay_addon" "tfdemoaddon1" {
  metadata {
    name    = "tfdemoaddon1"
    project = "terraform"
  }
}

output "addon_meta" {
  description = "metadata"
  value       = data.rafay_addon.tfdemoaddon1.metadata
}

output "addon_spec" {
  description = "spec"
  value       = data.rafay_addon.tfdemoaddon1.spec
}
