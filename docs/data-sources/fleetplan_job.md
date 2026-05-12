---
page_title: "rafay_fleetplan_job Data Source - terraform-provider-rafay"
subcategory: "Fleet Management"
description: |-
  Reads a specific fleet plan job from the Rafay platform.
---

# rafay_fleetplan_job (Data Source)

Use this data source to read the status of a specific fleet plan job from the Rafay platform.

## Example Usage

```terraform
data "rafay_fleetplan_job" "job1" {
  fleetplan_name = "fleetplan-env"
  project        = "defaultproject"
  name           = "1"
}
```

## Schema

### Required

- `fleetplan_name` (String) FleetPlan name
- `name` (String) FleetPlan job name
- `project` (String) Project name

### Read-Only

- `status` (List of Object) Fleet plan job status, including targets and their operations
