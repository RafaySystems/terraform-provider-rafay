variable "rafay_project" {
  description = "Rafay project name"
  type        = string
  default     = "demo-project"
}

variable "cluster_name" {
  description = "Cluster name"
  type        = string
  default     = "demo-cluster"
}

variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "us-west-2"
}

variable "kubernetes_version" {
  description = "Kubernetes control plane version"
  type        = string
  default     = "1.26"
}

variable "node_instance_type" {
  description = "Default instance type for node groups"
  type        = string
  default     = "t3.medium"
}

variable "node_min_size" {
  description = "Default node group min size"
  type        = number
  default     = 1
}

variable "node_desired" {
  description = "Default node group desired size"
  type        = number
  default     = 2
}

variable "node_max_size" {
  description = "Default node group max size"
  type        = number
  default     = 3
}