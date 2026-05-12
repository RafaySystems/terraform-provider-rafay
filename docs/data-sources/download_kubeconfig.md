---
page_title: "rafay_download_kubeconfig Data Source - terraform-provider-rafay"
subcategory: "Cluster Management"
description: |-
  Reads kubeconfig for clusters from the Rafay platform.
---

# rafay_download_kubeconfig (Data Source)

Use this data source to download kubeconfig for clusters from the Rafay platform.

## Example Usage

```terraform
# get kubeconfig for all cluster
data "rafay_download_kubeconfig" "allcluster" {
}

output "kubeconfig" {
  description = "kubeconfig"
  value       = data.rafay_download_kubeconfig.allcluster.kubeconfig
}

# get kubeconfig for a cluster
data "rafay_download_kubeconfig" "kubeconfig_cluster" {
  cluster = "cluster-name"
}

output "kubeconfig_cluster" {
  description = "kubeconfig_cluster"
  value       = data.rafay_download_kubeconfig.kubeconfig_cluster.kubeconfig
}

# get kubeconfig for a cluster and set namespace
data "rafay_download_kubeconfig" "kubeconfig_cluster_namespace" {
  cluster   = "cluster-name"
  namespace = "demo"
}

output "kubeconfig_cluster_namespace" {
  description = "kubeconfig_cluster_namespace"
  value       = data.rafay_download_kubeconfig.kubeconfig_cluster_namespace.kubeconfig
}
```
