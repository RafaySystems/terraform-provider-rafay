variable "rafay_config_file" {
  description = "rafay provider config file for authentication"
  sensitive   = true
  default     = "/home/bharat/.rafay/cli/config.json"
}