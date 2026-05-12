---
page_title: "rafay_blueprint Data Source - terraform-provider-rafay"
subcategory: "Blueprints & Add-ons"
description: |-
  Reads an existing blueprint from the Rafay platform.
---

# rafay_blueprint (Data Source)

Use this data source to read an existing blueprint from the Rafay platform.

## Example Usage

```terraform
data "rafay_blueprint" "blueprint1" {
  metadata {
    name    = "blueprint1"
    project = "terrafrom"
  }
}

output "blueprint_meta" {
  description = "metadata"
  value       = data.rafay_blueprint.blueprint1.metadata
}

output "blueprint_spec" {
  description = "spec"
  value       = data.rafay_blueprint.blueprint1.spec
}
```
