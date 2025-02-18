data "rafay_credentials_v3" "list" {
  projectname = "defaultproject"
}

output "credential_list" {
  description = "credentials list"
  value       = data.rafay_credentials_v3.list.credentials_list
}