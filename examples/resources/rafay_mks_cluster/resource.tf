resource "rafay_mks_cluster" "cluster" {
  name            = "demo-terraform-mks"
  projectname     = "dev"
  yamlfilepath    = "<file-path/mks-cluster.yaml>"
  yamlfileversion = ""
  waitflag        = "1"
}
