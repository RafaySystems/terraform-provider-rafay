resource "rafay_aks_cluster_spec" "demo-terraform-aks" {
  name            = "demo-terraform-aks"
  projectname     = "upgrade"
  yamlfilepath    = "/Users/stephanbenny/code/src/github.com/RafaySystems/terraform-provider-rafay/examples/resources/rafay_aks_cluster_spec/aks-cluster.yaml"
  yamlfileversion = "0"
}
