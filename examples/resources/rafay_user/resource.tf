resource "rafay_user" "user" {
  user_name = "sampleUser"
  first_name = "Bob"
  last_name = "Ross"
  phone = ""
  groups = ["group1-InfraAdmin"]
  generate_apikey = true
}

output "apikey" {
  description = "user api key"
  sensitive = true
  value       = rafay_user.user.apikey
}