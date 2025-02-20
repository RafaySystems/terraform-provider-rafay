data "rafay_cloud_credential" "cloud-sample-credential" {
  name    = "cloud-sample-credential"
  project = "sample-project"
}


output "credential" {
  description = "credentials details"
  value       = data.rafay_cloud_credential.cloud-sample-credential
}