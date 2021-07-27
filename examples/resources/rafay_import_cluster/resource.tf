resource "rafay_import_cluster" "import_cluster" {
  clustername       = "terraform-importcluster"
  projectname       = "dev-proj"
  blueprint         = "default"
  blueprint_version = ""
  location          = "losangeles-us"
  kubeconfig_path   = "<file-path/kubeconfig.yaml>"
  description       = ""
}