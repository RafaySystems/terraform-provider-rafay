resource "rafay_import_cluster" "import_cluster" {
  clustername       = "terraform-importcluster"
  projectname       = "terraform"
  blueprint         = "default"
  location          = "losangeles-us"
  kubeconfig_path   = "<file-path/kubeconfig.yaml>"
  labels            = {
    "key1" = "value1"
    "key2" = "value2"
  }
}
