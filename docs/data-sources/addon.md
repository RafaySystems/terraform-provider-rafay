---
page_title: "rafay_addon Data Source - terraform-provider-rafay"
subcategory: "Blueprints & Add-ons"
description: |-
  Reads an existing add-on from the Rafay platform.
---

# rafay_addon (Data Source)

Use this data source to read an existing add-on from the Rafay platform.

## Example Usage

```terraform
data "rafay_addon" "tfdemoaddon1" {
  metadata {
    name    = "tfdemoaddon1"
    project = "terraform"
  }
}

output "addon_spec" {
  description = "spec"
  value       = data.rafay_addon.tfdemoaddon1.spec
}
```
