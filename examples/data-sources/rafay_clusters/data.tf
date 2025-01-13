data "rafay_blueprints" "ns" {
  projectname = "defaultproject"
}

output "clusterlist" {
  description = "namespaces"
  value       = data.rafay_blueprints.ns.blueprints
}
