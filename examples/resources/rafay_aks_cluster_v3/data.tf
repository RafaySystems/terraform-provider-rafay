data "rafay_aks_cluster_v3" "cluster" {
  metadata {
    name    = "cluster-name"
    project = "demo"
  }
}

output "aks_cluster_v3" {
  description = "aks_cluster_v3"
  value       = data.rafay_aks_cluster_v3.cluster
}
