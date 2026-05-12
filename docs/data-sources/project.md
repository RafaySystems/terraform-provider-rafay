---
page_title: "rafay_project Data Source - terraform-provider-rafay"
subcategory: "Projects"
description: |-
  Reads an existing project from the Rafay platform.
---

# rafay_project (Data Source)

Use this data source to read an existing project from the Rafay platform.

## Example Usage

```terraform
data "rafay_project" "tfdemoproject" {
  metadata {
    name = "tfdemoproject"
  }
}

output "project_meta" {
  description = "metadata"
  value       = data.rafay_project.tfdemoproject.metadata
}

output "project_spec" {
  description = "spec"
  value       = data.rafay_project.tfdemoproject.spec
}

output "project_spec_driftwebhook_enabled" {
  description = "driftwebhook_enabled"
  value       = data.rafay_project.tfdemoproject.spec.0.drift_webhook.0.enabled
}
```
