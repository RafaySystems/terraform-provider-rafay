---
page_title: "rafay_gke_cluster Data Source - terraform-provider-rafay"
subcategory: "Cluster Management"
description: |-
  Reads an existing GKE cluster from the Rafay platform.
---

# rafay_gke_cluster (Data Source)

Use this data source to read an existing GKE cluster from the Rafay platform.

## Example Usage

```terraform
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
```
