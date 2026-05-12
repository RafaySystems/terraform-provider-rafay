---
page_title: "rafay_cluster_blueprint_status Data Source - terraform-provider-rafay"
subcategory: "Cluster Management"
description: |-
  Reads the blueprint sync status for a cluster from the Rafay platform.
---

# rafay_cluster_blueprint_status (Data Source)

Use this data source to get the cluster blueprint sync status from the Rafay platform.

## Example Usage

```terraform
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
```
