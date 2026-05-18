data "rafay_cluster_kubernetes_versions" "versions" {
  cluster_type               = "mks"
  project                    = "defaultproject"
  cluster_name               = "test-cluster"
  include_deprecated_version = true
}

output "cluster_type" {
  description = "Cluster type"
  value       = data.rafay_cluster_kubernetes_versions.versions.cluster_type
}

output "default_version" {
  description = "Default Kubernetes version"
  value       = data.rafay_cluster_kubernetes_versions.versions.default_version
}

output "latest_version" {
  description = "Latest Kubernetes version"
  value       = data.rafay_cluster_kubernetes_versions.versions.latest_version
}

output "versions" {
  description = "List of available Kubernetes versions"
  value       = data.rafay_cluster_kubernetes_versions.versions.versions
}
