resource "rafay_user" "user" {
  user_name = "sampleUser"
  first_name = "Bob"
  last_name = "Ross"
  phone = ""
  groups = ["group1-InfraAdmin"]
  generate_apikey = true
}