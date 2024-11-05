data "rafay_import_cluster" "cluster" {
    clustername = "import"
    projectname = "tf-project"
}

output "eks_cluster" {
  description = "import_cluster"
  value       = data.rafay_import_cluster.cluster
}