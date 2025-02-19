data "rafay_credentials" "list" {
  projectname = "defaultproject"
}

output "credential_list" {
  description = "credentials list"
  value       = data.rafay_credentials.list.credentials
}