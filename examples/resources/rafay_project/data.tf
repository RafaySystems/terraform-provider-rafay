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

output "project_spec_drift_webhook" {
  description = "drift_webhook"
  value       = data.rafay_project.tfdemoproject.spec.0.drift_webhook
}

output "project_spec_drift_webhook_enabled" {
  description = "drift_webhook"
  value       = data.rafay_project.tfdemoproject.spec.0.drift_webhook.0.enabled
}



