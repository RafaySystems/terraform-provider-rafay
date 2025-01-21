data "rafay_blueprints" "list" {
  projectname = "defaultproject"
}

output "blueprint_list" {
  description = "blueprints list"
  value       = data.rafay_blueprints.list.blueprints
}
