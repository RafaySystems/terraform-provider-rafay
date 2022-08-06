resource "rafay_aks_cluster_spec" "demo-terraform-aks" {
  name            = "demo-terraform-aks"
  projectname     = "terraform"
  yamlfilepath    = "/Users/testuser/terraform-provider-rafay/examples/resources/rafay_aks_cluster_spec/aks-cluster.yaml"
  yamlfileversion = "0"
  checkdiff       = true
}
