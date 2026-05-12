---
page_title: "rafay_fleetplan Data Source - terraform-provider-rafay"
subcategory: "Fleet Management"
description: |-
  Reads an existing fleet plan from the Rafay platform.
---

# rafay_fleetplan (Data Source)

Use this data source to read an existing fleet plan from the Rafay platform.

## Example Usage

```terraform
data "rafay_fleetplan" "environment_fleetplan" {
  metadata {
    project = "defaultproject"
    name    = "fleetplan-env"
  }
}
```
