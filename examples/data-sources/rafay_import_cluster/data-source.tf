data "rafay_import_cluster" "import-sample-cluster" {
  metadata = {
    name    = "import-cluster"
    project = "sample-project"
  }
}

output "import_cluster" {
  description = "import_cluster"
  value       = data.rafay_import_cluster.import-sample-cluster
}
