data "rafay_blueprint" "blueprint1" {
  metadata {
    name    = "blueprint1"
    project = "terrafrom"
  }
}

output "blueprint_meta" {
  description = "metadata"
  value       = data.rafay_blueprint.blueprint1.metadata
}

output "blueprint_spec" {
  description = "spec"
  value       = data.rafay_blueprint.blueprint1.spec
}
