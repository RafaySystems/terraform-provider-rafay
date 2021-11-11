resource "rafay_aks_cluster" "cluster" {
  name            = "demo-terraform"
  projectname     = "dev"
  yamlfilepath    = "<file-path/aks-cluster.yaml>"
  yamlfileversion = "0"
  waitflag        = "1"
}
