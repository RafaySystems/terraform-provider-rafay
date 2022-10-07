resource "rafay_user" "user" {
  user_name = "sampleUser"
  first_name = "Bob"
  last_name = "Ross"
  phone = "14083074010"
  groups = ["group1", "group2"]
  generate_apikey = true
  console_access = true
}

output "apikey" {
  description = "user api key"
  sensitive = true
  value     = rafay_user.user.apikey
}
output "api_secret" {
  description = "user api secret"
  sensitive = true
  value       = rafay_user.user.api_secret
}
