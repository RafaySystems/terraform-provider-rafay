data "rafay_namespaces" "list" {
  projectname = "defaultproject"
}

output "namespace_list" {
  description = "clusters list"
  value       = data.rafay_namespaces.list.namespaces
}
