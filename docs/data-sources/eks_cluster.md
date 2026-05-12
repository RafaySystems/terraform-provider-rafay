---
page_title: "rafay_eks_cluster Data Source - terraform-provider-rafay"
subcategory: "Cluster Management"
description: |-
  Reads an existing EKS cluster from the Rafay platform.
---

# rafay_eks_cluster (Data Source)

Use this data source to read an existing EKS cluster from the Rafay platform.

## Example Usage

```terraform
data "rafay_eks_cluster" "cluster" {
  cluster {
    metadata {
      name    = "cluster-name"
      project = "demo"
    }
  }
}

output "eks_cluster" {
  description = "eks_cluster"
  value       = data.rafay_eks_cluster.cluster
}
```
