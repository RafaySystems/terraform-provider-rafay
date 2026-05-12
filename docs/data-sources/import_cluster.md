---
page_title: "rafay_import_cluster Data Source - terraform-provider-rafay"
subcategory: "Cluster Management"
description: |-
  Reads an existing imported cluster from the Rafay platform.
---

# rafay_import_cluster (Data Source)

Use this data source to read an existing imported cluster from the Rafay platform.

## Example Usage

```terraform
data "rafay_import_cluster" "import-sample-cluster" {
  metadata = {
    name    = "import-cluster"
    project = "sample-project"
  }
}

output "import_cluster" {
  description = "import_cluster"
  value       = data.rafay_import_cluster.import-sample-cluster
}
```
