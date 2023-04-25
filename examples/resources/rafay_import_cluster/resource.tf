resource "rafay_import_cluster" "import_cluster" {
  clustername       = "terraform-importcluster"
  projectname       = "terraform"
  blueprint         = "default"
  kubeconfig_path   = "<file-path/kubeconfig.yaml>"
  location          = "losangeles-us"
  values_path       = "<optional_path/values.yaml>"
  bootstrap_path    = "<optional_path/bootstrap.yaml>"
}

output "values_data" {
  value = rafay_import_cluster.import_cluster.values_data
}

output "values_path" {
  value = rafay_import_cluster.import_cluster.values_path
}

output "bootstrap_data" {
  value = rafay_import_cluster.import_cluster.bootstrap_data
}

output "bootstrap_path" {
  value = rafay_import_cluster.import_cluster.bootstrap_path
}