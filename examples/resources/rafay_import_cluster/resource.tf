resource "rafay_import_cluster" "import_cluster" {
  clustername      = "terraform-importcluster"
  projectname      = "dev1-proj"
  blueprint        = "default"
  location         = "losangeles-us"
  kube_config_path = ""
  description      = ""
}