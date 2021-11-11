resource "rafay_aks_cluster" "cluster" {
  name            = "demo-cluster-aks"
  projectname     = "dev3"
  yamlfilepath    = "/Users/krishna/code/src/github.com/RafaySystems/terraform-provider-rafay/examples/resources/rafay_aks_cluster/aks-cluster.yaml"
  yamlfileversion = "1"
  waitflag        = "0"
}
