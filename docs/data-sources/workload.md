---
page_title: "rafay_workload Data Source - terraform-provider-rafay"
subcategory: "Configuration"
description: |-
  Reads an existing workload from the Rafay platform.
---

# rafay_workload (Data Source)

Use this data source to read an existing workload from the Rafay platform.

## Example Usage

```terraform
data "rafay_workload" "testworkload" {
  metadata {
    name    = "testworkload"
    project = "benny-cilium-test"
  }
}

output "rafay_workload_metadata" {
  description = "metadata"
  value       = data.rafay_workload.testworkload.metadata
}

output "rafay_workload_spec" {
  description = "spec"
  value       = data.rafay_workload.testworkload.spec
}

output "rafay_workload_condition" {
  description = "condition"
  value       = data.rafay_workload.testworkload.condition
}

output "rafay_workload_condition_status" {
  description = "condition_status"
  value       = data.rafay_workload.testworkload.condition_status
}

output "rafay_workload_condition_reason" {
  description = "condition_reason"
  value       = data.rafay_workload.testworkload.reason
}
```
