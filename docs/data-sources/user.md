---
page_title: "rafay_user Data Source - terraform-provider-rafay"
subcategory: "Security & Access"
description: |-
  Reads an existing user from the Rafay platform.
---

# rafay_user (Data Source)

Use this data source to read an existing user from the Rafay platform.

## Example Usage

```terraform
data "rafay_user" "user" {
  user_name = "name"
}

output "user_groups" {
  description = "user_groups"
  value       = data.rafay_user.user.groups
}
```
