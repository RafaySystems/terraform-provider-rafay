data "rafay_group" "group" {
    name    = "group_name"
}

output "group" {
  description = "group"
  value       = data.rafay_group.group
}
