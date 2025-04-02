data "rafay_cluster_blueprint_status" "bp-sync-status" {
  metadata {
    name    = "cluster1"
    project = "defaultproject"
  }
}

output "cluster_meta" {
  description = "metadata"
  value       = data.rafay_cluster_blueprint_status.bp-sync-status.metadata
}

output "status" {
  description = "status"
  value       = data.rafay_cluster_blueprint_status.bp-sync-status.status
}


