resource "rafay_eks_cluster" "cluster" {
  name            = "demo-terraform"
  projectname     = "dev"
  yamlfilepath    = "<file-path/aks-cluster.yaml>"
  yamlfileversion = ""
  waitflag        = "0"
}
