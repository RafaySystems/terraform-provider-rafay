---
page_title: "rafay_aks_cluster_v3 Data Source - terraform-provider-rafay"
subcategory: "Cluster Management"
description: |-
  Reads an existing AKS cluster (v3) from the Rafay platform.
---

# rafay_aks_cluster_v3 (Data Source)

Use this data source to read an existing AKS cluster (v3) from the Rafay platform.

## Example Usage

```terraform
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
```
