data "rafay_aks_cluster" "cluster" {
  apiversion = "rafay.io/v1alpha1"
  kind       = "Cluster"
  metadata {
    name    = "cluster-name"
    project = "demo"
  }
}

output "aks_cluster" {
  description = "aks_cluster"
  value       = data.rafay_aks_cluster.cluster
}