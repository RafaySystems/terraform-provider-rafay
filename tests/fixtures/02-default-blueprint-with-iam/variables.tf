variable "name" {
  description = "cluster name"
  sensitive   = false
  default     = "test-cluster"
}

variable "project" {
  description = "rafay project"
  sensitive   = false
  default     = "akshay"
}

variable "cloud_provider" {
  description = "cloud provider"
  sensitive   = false
  default     = "qa-automation"
}

variable "instance_type" {
  description = "node instance type"
  sensitive   = false
  default     = "t3.large"
}

variable "volume_type" {
  description = "node volume type"
  sensitive   = false
  default     = "gp3"
}