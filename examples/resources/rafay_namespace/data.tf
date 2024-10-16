#Example of data namespace 
data "rafay_namespace" "tfdemonamespace1" {
  metadata {
    name = "tfdemonamespace1"
    project = "terraform"
  }
}

output "tfdemonamespace_out" {
  description = "spec"
  value       = data.rafay_namespace.tfdemonamespace1.spec
}

#Example of data namespaces 
data "rafay_namespaces" "all" {
  metadata {
    project = "terraform"
  }
}

output "allnamespaces" {
  description = "spec"
  value       = data.rafay_namespaces.all
}
