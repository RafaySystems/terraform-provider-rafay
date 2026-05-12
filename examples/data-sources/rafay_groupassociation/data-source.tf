data "rafay_groupassociation" "association" {
  group   = "group-name"
  project = "demo"
}

output "groupassociation" {
  description = "groupassociation"
  value       = data.rafay_groupassociation.association
}
