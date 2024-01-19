data "rafay_gke_cluster" "cluster" {
  metadata {
    name    = "cluster-gke"
    project = "demo"
  }
}

output "gke_cluster" {
  description = "gke_cluster"
  value       = data.rafay_gke_cluster.cluster
}
