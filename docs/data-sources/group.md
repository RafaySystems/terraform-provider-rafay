---
page_title: "rafay_group Data Source - terraform-provider-rafay"
subcategory: "Security & Access"
description: |-
  Reads an existing group from the Rafay platform.
---

# rafay_group (Data Source)

Use this data source to read an existing group from the Rafay platform.

## Example Usage

```terraform
data "rafay_group" "group" {
  name = "group_name"
}

output "group" {
  description = "group"
  value       = data.rafay_group.group
}
```
