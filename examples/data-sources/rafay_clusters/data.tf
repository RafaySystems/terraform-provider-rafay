data "rafay_clusters" "list" {
  projectname = "defaultproject"
}

output "cluster_list" {
  description = "clusters list"
  value       = data.rafay_clusters.list.clusters
}
