# To rotate the api key
# terraform apply -replace=rafay_access_apikey.sampleuser
resource "rafay_access_apikey" "sampleuser10" {
  user_name = "sampleuser@sample.com"
  lifecycle {
    create_before_destroy = true
  }
}

output "apikey" {
  description = "user api key"
  sensitive   = true
  value       = rafay_access_apikey.sampleuser.apikey
}

output "api_secret" {
  description = "user api secret"
  sensitive   = true
  value       = rafay_access_apikey.sampleuser.api_secret
}
