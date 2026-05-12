---
page_title: "rafay_groupassociation Data Source - terraform-provider-rafay"
subcategory: "Security & Access"
description: |-
  Reads an existing group association from the Rafay platform.
---

# rafay_groupassociation (Data Source)

Use this data source to read an existing group association from the Rafay platform.

## Example Usage

```terraform
data "rafay_groupassociation" "association" {
  group   = "group-name"
  project = "demo"
}

output "groupassociation" {
  description = "groupassociation"
  value       = data.rafay_groupassociation.association
}
```
