resource "rafay_import_cluster" "import_cluster" {
  metadata {
    name       = "terraform-importcluster"
    project       = "dev"
  }
  spec {
    type = "eks"
    blueprint         = "default"
    blueprint_version = ""
    location          = "losangeles-us"
    kubeconfig_path   = "<file-path/kubeconfig.yaml>"
    description       = ""
  }
  /*
  clustername       = "terraform-importcluster"
  projectname       = "dev-proj"
  blueprint         = "default"
  blueprint_version = ""
  location          = "losangeles-us"
  kubeconfig_path   = "<file-path/kubeconfig.yaml>"
  description       = ""
  */
}