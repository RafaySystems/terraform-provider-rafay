---
page_title: "rafay_aks_cluster Data Source - terraform-provider-rafay"
subcategory: "Cluster Management"
description: |-
  Reads an existing AKS cluster from the Rafay platform.
---

# rafay_aks_cluster (Data Source)

Use this data source to read an existing AKS cluster from the Rafay platform.

## Example Usage

```terraform
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
```
