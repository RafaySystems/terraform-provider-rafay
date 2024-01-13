data "rafay_user" "user" {
    user_name    = "name"
}

output "user_groups" {
  description = "user_groups"
  value       = data.rafay_user.user.groups
}
