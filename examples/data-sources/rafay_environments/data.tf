data "rafay_environments" "list" {
  projectname = "defaultproject"
}

output "environment_list" {
  description = "environments list"
  value       = data.rafay_environments.list.environments
}
