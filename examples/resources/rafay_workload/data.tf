# Description: This file contains the data source for the Rafay Workload
data "rafay_workload" "testworkload" {
  metadata {
    name    = "testworkload"
    project = "benny-cilium-test"
    annotations = {
      "key1" = "value1"
      "key2" = "value2"
    }
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

