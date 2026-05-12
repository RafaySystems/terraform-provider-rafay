---
page_title: "rafay_fleetplans Data Source - terraform-provider-rafay"
subcategory: "Fleet Management"
description: |-
  Lists all fleet plans in the Rafay platform.
---

# rafay_fleetplans (Data Source)

Use this data source to list all fleet plans in the Rafay platform.

## Example Usage

```terraform
data "rafay_fleetplans" "list" {
  project = "defaultproject"
}
```

## Schema

### Required

- `project` (String) Project name from where fleetplans to be listed

### Optional

- `type` (String) Resource type of the fleet plan. Defaults to `clusters`.

### Read-Only

- `fleetplans` (List of Object) List of fleetplans
