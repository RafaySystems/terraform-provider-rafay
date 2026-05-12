data "rafay_addon" "tfdemoaddon1" {
  metadata {
    name    = "tfdemoaddon1"
    project = "terraform"
  }
}

output "addon_spec" {
  description = "spec"
  value       = data.rafay_addon.tfdemoaddon1.spec
}
