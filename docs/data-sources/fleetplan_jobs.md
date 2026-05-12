---
page_title: "rafay_fleetplan_jobs Data Source - terraform-provider-rafay"
subcategory: "Fleet Management"
description: |-
  Lists fleet plan jobs for a given fleet plan in the Rafay platform.
---

# rafay_fleetplan_jobs (Data Source)

Use this data source to list fleet plan jobs for a given fleet plan in the Rafay platform.

## Example Usage

```terraform
data "rafay_fleetplan_jobs" "fleetplan_jobs" {
  fleetplan_name = "fleetplan-env"
  project        = "defaultproject"
}
```

## Schema

### Required

- `fleetplan_name` (String) FleetPlan name
- `project` (String) Project name

### Read-Only

- `jobs` (List of Object) List of fleetplan jobs with name, state, creation time, and other details
