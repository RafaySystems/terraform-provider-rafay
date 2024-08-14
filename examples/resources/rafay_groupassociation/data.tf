data "rafay_groupassociation" "association" {
    group    = "rctl-grp"
    project = "sample-shetty"
}

output "groupassociation" {
  description = "groupassociation"
  value       = data.rafay_groupassociation.association
}
