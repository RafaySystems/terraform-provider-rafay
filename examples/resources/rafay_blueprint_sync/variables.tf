variable "rafay_config_file" {
  description = "rafay provider config file for authentication"
  sensitive   = true
  default     = "/Users/krishna/.rafay/cli/config.json"
}

# Every apply re-publishes regardless of this value. It only controls how
# the backend handles a sync that's already in progress: false errors out,
# true restarts it. Pass -var 'force_sync=true' when you need to override
# an in-progress sync, e.g.:
#   terraform apply -var 'force_sync=true'
variable "force_sync" {
  description = "Whether to force-restart a blueprint sync that's already in progress."
  type        = bool
  default     = false
}
